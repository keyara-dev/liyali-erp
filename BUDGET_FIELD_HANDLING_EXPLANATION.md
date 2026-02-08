# How Backend Handles totalBudget and allocatedAmount

## Backend Processing Flow

### 1. Request Type Definition

**Location**: `backend/types/documents.go` lines 103-113

```go
type CreateBudgetRequest struct {
    BudgetCode      string  `json:"budgetCode,omitempty"`
    Name            string  `json:"name,omitempty"`
    Description     string  `json:"description,omitempty"`
    Department      string  `json:"department" validate:"required"`
    DepartmentID    string  `json:"departmentId,omitempty"`
    FiscalYear      string  `json:"fiscalYear" validate:"required"`
    TotalBudget     float64 `json:"totalBudget" validate:"required,gt=0"`      // Must be > 0
    AllocatedAmount float64 `json:"allocatedAmount" validate:"required,gte=0"` // Must be >= 0
    Currency        string  `json:"currency,omitempty"`
}
```

**Validation Rules**:

- `TotalBudget`: Required, must be greater than 0 (`gt=0`)
- `AllocatedAmount`: Required, must be greater than or equal to 0 (`gte=0`)

---

### 2. Backend Handler Processing

**Location**: `backend/handlers/budget.go` lines 109-195

#### Step 1: Parse Request

```go
var req types.CreateBudgetRequest
if err := c.BodyParser(&req); err != nil {
    return utils.SendBadRequestError(c, "Invalid request body")
}
```

#### Step 2: Validate TotalBudget

```go
if req.TotalBudget <= 0 {
    return utils.SendBadRequestError(c, "Total budget must be greater than 0")
}
```

- ❌ Rejects if `totalBudget` is 0 or negative

#### Step 3: Validate AllocatedAmount

```go
if req.AllocatedAmount < 0 {
    return utils.SendBadRequestError(c, "Allocated amount cannot be negative")
}
```

- ✅ Allows `allocatedAmount` to be 0
- ❌ Rejects if `allocatedAmount` is negative

#### Step 4: Calculate RemainingAmount

```go
remainingAmount := req.TotalBudget - req.AllocatedAmount
```

**Formula**: `remainingAmount = totalBudget - allocatedAmount`

#### Step 5: Create Budget Record

```go
budget := models.Budget{
    ID:              budgetID,
    OrganizationID:  tenant.OrganizationID,
    OwnerID:         userID,
    BudgetCode:      req.BudgetCode,
    Name:            req.Name,
    Description:     req.Description,
    Department:      req.Department,
    DepartmentID:    req.DepartmentID,
    Status:          "draft",
    FiscalYear:      req.FiscalYear,
    TotalBudget:     req.TotalBudget,        // Stored as-is
    AllocatedAmount: req.AllocatedAmount,    // Stored as-is
    RemainingAmount: remainingAmount,        // Calculated
    Currency:        req.Currency,
    ApprovalStage:   0,
    CreatedBy:       userID,
    CreatedAt:       time.Now(),
    UpdatedAt:       time.Now(),
}
```

---

## Scenarios and Results

### Scenario 1: Original Frontend Code (INCORRECT)

**Frontend sends**:

```json
{
  "totalBudget": 50000,
  "allocatedAmount": 50000
}
```

**Backend calculates**:

```
remainingAmount = 50000 - 50000 = 0
```

**Result in Database**:

- `total_budget`: 50000
- `allocated_amount`: 50000
- `remaining_amount`: 0

**Problem**: ❌ Budget appears fully allocated immediately, with no remaining amount!

---

### Scenario 2: Fixed Frontend Code (CORRECT)

**Frontend sends**:

```json
{
  "totalBudget": 50000,
  "allocatedAmount": 0
}
```

**Backend calculates**:

```
remainingAmount = 50000 - 0 = 50000
```

**Result in Database**:

- `total_budget`: 50000
- `allocated_amount`: 0
- `remaining_amount`: 50000

**Result**: ✅ Budget starts with full amount available!

---

### Scenario 3: Partially Allocated Budget

**Frontend sends**:

```json
{
  "totalBudget": 50000,
  "allocatedAmount": 15000
}
```

**Backend calculates**:

```
remainingAmount = 50000 - 15000 = 35000
```

**Result in Database**:

- `total_budget`: 50000
- `allocated_amount`: 15000
- `remaining_amount`: 35000

**Result**: ✅ Budget starts with 15000 already allocated, 35000 remaining.

---

## Business Logic Interpretation

### What is `allocatedAmount`?

The `allocatedAmount` represents **money that has already been committed or assigned** from the budget, even before the budget is approved. This could be:

1. **Pre-allocated funds**: Money already earmarked for specific projects
2. **Carried-over commitments**: Existing obligations from previous periods
3. **Reserved amounts**: Funds set aside for known upcoming expenses

### Typical Use Cases

#### Use Case 1: New Budget (Most Common)

```json
{
  "totalBudget": 100000,
  "allocatedAmount": 0
}
```

- Start with full budget available
- Allocate funds as requisitions/POs are created

#### Use Case 2: Budget with Pre-commitments

```json
{
  "totalBudget": 100000,
  "allocatedAmount": 25000
}
```

- Total budget is 100K
- 25K already committed to existing contracts
- Only 75K available for new allocations

#### Use Case 3: Budget Adjustment

```json
{
  "totalBudget": 150000,
  "allocatedAmount": 80000
}
```

- Budget increased from 100K to 150K
- 80K already allocated to approved requisitions
- 70K available for new requests

---

## Frontend Fix Rationale

### Why `allocatedAmount: 0` is Correct

When creating a **new budget**, the standard practice is:

1. ✅ Set `totalBudget` to the full budget amount
2. ✅ Set `allocatedAmount` to `0` (nothing allocated yet)
3. ✅ Backend calculates `remainingAmount = totalBudget - 0 = totalBudget`

As requisitions and purchase orders are created and approved:

- `allocatedAmount` increases
- `remainingAmount` decreases
- `totalBudget` stays constant (unless budget is revised)

### When Would You Set `allocatedAmount > 0`?

Only in special cases:

- Importing existing budgets with commitments
- Mid-year budget creation with existing obligations
- Budget transfers with pre-allocated amounts

For a standard "Create New Budget" form, **`allocatedAmount` should always start at 0**.

---

## Database Schema

**Table**: `budgets`

```sql
CREATE TABLE budgets (
    id VARCHAR(255) PRIMARY KEY,
    organization_id VARCHAR(255) NOT NULL,
    owner_id VARCHAR(255) NOT NULL,
    budget_code VARCHAR(100) NOT NULL,
    name VARCHAR(255),
    description TEXT,
    department VARCHAR(100),
    department_id VARCHAR(255),
    status VARCHAR(50) DEFAULT 'draft',
    fiscal_year VARCHAR(10),
    total_budget DECIMAL(15,2),      -- Total budget amount
    allocated_amount DECIMAL(15,2),  -- Amount already allocated
    remaining_amount DECIMAL(15,2),  -- Calculated: total - allocated
    currency VARCHAR(3) DEFAULT 'USD',
    approval_stage INTEGER DEFAULT 0,
    created_by VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

---

## Summary

### Backend Behavior

1. ✅ Accepts both `totalBudget` and `allocatedAmount` from frontend
2. ✅ Validates `totalBudget > 0`
3. ✅ Validates `allocatedAmount >= 0`
4. ✅ Calculates `remainingAmount = totalBudget - allocatedAmount`
5. ✅ Stores all three values in database

### Frontend Fix

- **Before**: `allocatedAmount: parseFloat(formData.totalAmount)` ❌
- **After**: `allocatedAmount: 0` ✅

### Impact

- New budgets now correctly start with full amount available
- `remainingAmount` equals `totalBudget` for new budgets
- Budget allocation tracking works as intended
