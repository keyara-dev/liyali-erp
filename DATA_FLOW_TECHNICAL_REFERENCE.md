# Data Flow Technical Reference

## Quick Reference: Type Mappings

### Core Types Mapping

| Frontend (TypeScript) | Backend (Go) | Database (PostgreSQL) | Notes |
|----------------------|--------------|----------------------|-------|
| `string` | `string` | `VARCHAR(255)` | Standard text |
| `number` | `float64` | `DECIMAL(15,2)` | Monetary values |
| `number` | `int` | `INTEGER` | Counts, stages |
| `Date` | `time.Time` | `TIMESTAMP` | Dates/times |
| `boolean` | `bool` | `BOOLEAN` | Flags |
| `any[]` | `datatypes.JSONType[T]` | `JSONB` | Complex arrays |
| `Record<string, any>` | `datatypes.JSON` | `JSONB` | Flexible objects |
| `string \| null` | `*string` | `VARCHAR(255)` | Optional text |
| `Date \| null` | `*time.Time` | `TIMESTAMP` | Optional dates |

---

## API Request/Response Patterns

### Standard API Response Format

```typescript
// Success Response
{
  success: true,
  message: "Operation completed successfully",
  data: { /* entity data */ },
  status: 200
}

// Error Response
{
  success: false,
  message: "Error description",
  error: "error_code",
  status: 400,
  errors: {
    "field_name": ["error message"]
  }
}

// Paginated Response
{
  success: true,
  data: [ /* array of entities */ ],
  pagination: {
    page: 1,
    limit: 10,
    total: 100,
    totalPages: 10,
    hasNext: true,
    hasPrev: false
  }
}
```

---

## Entity-Specific Data Flows

### Requisition Data Flow

**Creation Flow**:
```
Frontend Form
  ↓ (CreateRequisitionRequest)
  ├─ title: string
  ├─ description: string
  ├─ department: string
  ├─ departmentId: string
  ├─ items: RequisitionItem[]
  ├─ budgetCode: string
  ├─ costCenter: string
  ├─ projectCode: string
  ├─ requiredByDate: Date
  ├─ priority: Priority
  ├─ createdBy: string
  ├─ createdByName: string
  └─ createdByRole: UserRole
  ↓
Backend Handler
  ├─ Validates all fields
  ├─ Generates REQNumber (unique)
  ├─ Creates Requisition model
  ├─ Stores in DB
  └─ Returns Requisition
  ↓
Database (requisitions table)
  ├─ id: UUID
  ├─ req_number: VARCHAR (unique)
  ├─ title: VARCHAR
  ├─ description: TEXT
  ├─ items: JSONB
  ├─ budget_code: VARCHAR
  ├─ cost_center: VARCHAR
  ├─ project_code: VARCHAR
  ├─ required_by_date: TIMESTAMP
  ├─ priority: VARCHAR
  ├─ created_by: VARCHAR (FK → users.id)
  ├─ created_by_name: VARCHAR
  ├─ created_by_role: VARCHAR
  ├─ action_history: JSONB
  ├─ metadata: JSONB
  ├─ created_at: TIMESTAMP
  └─ updated_at: TIMESTAMP
  ↓
Frontend Receives
  └─ Complete Requisition object with all fields
```

**Approval Flow**:
```
Frontend (ApproveTaskRequest)
  ├─ comments?: string
  ├─ signature: string
  └─ stageNumber?: number
  ↓
Backend Handler
  ├─ Validates signature
  ├─ Updates ApprovalTask status
  ├─ Adds to ApprovalHistory
  ├─ Updates ActionHistory
  └─ Returns updated Requisition
  ↓
Database Updates
  ├─ approval_tasks: status = 'approved'
  ├─ requisitions: approval_history (JSONB append)
  ├─ requisitions: action_history (JSONB append)
  └─ requisitions: updated_at = NOW()
```

---

### Purchase Order Data Flow

**Creation from Requisition**:
```
Frontend (CreatePurchaseOrderRequest)
  ├─ vendorId: string
  ├─ items: POItem[]
  ├─ totalAmount: number
  ├─ currency: string
  ├─ deliveryDate: Date
  ├─ linkedRequisition: string (REQ ID)
  ├─ title: string
  ├─ description: string
  ├─ department: string
  ├─ departmentId: string
  ├─ priority: Priority
  ├─ budgetCode: string
  ├─ costCenter: string
  ├─ projectCode: string
  ├─ glCode: string
  ├─ subtotal: number
  ├─ tax: number
  ├─ createdBy: string
  └─ sourceRequisitionId: string
  ↓
Backend Handler
  ├─ Validates vendor exists
  ├─ Validates budget availability
  ├─ Generates PONumber (unique)
  ├─ Creates PurchaseOrder model
  ├─ Links to Requisition
  ├─ Stores in DB
  └─ Returns PurchaseOrder
  ↓
Database (purchase_orders table)
  ├─ id: UUID
  ├─ po_number: VARCHAR (unique)
  ├─ vendor_id: VARCHAR (FK → vendors.id)
  ├─ items: JSONB
  ├─ total_amount: DECIMAL
  ├─ currency: VARCHAR
  ├─ delivery_date: TIMESTAMP
  ├─ linked_requisition: VARCHAR (FK → requisitions.id)
  ├─ title: VARCHAR
  ├─ description: TEXT
  ├─ department: VARCHAR
  ├─ department_id: VARCHAR
  ├─ priority: VARCHAR
  ├─ budget_code: VARCHAR
  ├─ cost_center: VARCHAR
  ├─ project_code: VARCHAR
  ├─ gl_code: VARCHAR
  ├─ subtotal: DECIMAL
  ├─ tax: DECIMAL
  ├─ total: DECIMAL
  ├─ created_by: VARCHAR (FK → users.id)
  ├─ owner_id: VARCHAR
  ├─ action_history: JSONB
  ├─ metadata: JSONB
  ├─ created_at: TIMESTAMP
  └─ updated_at: TIMESTAMP
```

---

### Payment Voucher Data Flow

**Creation from Purchase Order**:
```
Frontend (CreatePaymentVoucherRequest)
  ├─ vendorId: string
  ├─ invoiceNumber: string
  ├─ amount: number
  ├─ currency: string
  ├─ paymentMethod: PaymentMethod (bank_transfer | cash)
  ├─ glCode: string
  ├─ description: string
  ├─ linkedPO: string (PO ID)
  ├─ title: string
  ├─ department: string
  ├─ departmentId: string
  ├─ priority: Priority
  ├─ items: PaymentItem[]
  ├─ budgetCode: string
  ├─ costCenter: string
  ├─ projectCode: string
  ├─ taxAmount: number
  ├─ withholdingTaxAmount: number
  ├─ paymentDueDate: Date
  ├─ bankDetails: object
  ├─ createdBy: string
  └─ sourcePurchaseOrderNumber: string
  ↓
Backend Handler
  ├─ Validates vendor exists
  ├─ Validates PO exists
  ├─ Generates VoucherNumber (unique)
  ├─ Creates PaymentVoucher model
  ├─ Links to PO
  ├─ Stores in DB
  └─ Returns PaymentVoucher
  ↓
Database (payment_vouchers table)
  ├─ id: UUID
  ├─ voucher_number: VARCHAR (unique)
  ├─ vendor_id: VARCHAR (FK → vendors.id)
  ├─ invoice_number: VARCHAR
  ├─ amount: DECIMAL
  ├─ currency: VARCHAR
  ├─ payment_method: VARCHAR
  ├─ gl_code: VARCHAR
  ├─ description: TEXT
  ├─ linked_po: VARCHAR (FK → purchase_orders.id)
  ├─ title: VARCHAR
  ├─ department: VARCHAR
  ├─ department_id: VARCHAR
  ├─ priority: VARCHAR
  ├─ items: JSONB
  ├─ budget_code: VARCHAR
  ├─ cost_center: VARCHAR
  ├─ project_code: VARCHAR
  ├─ tax_amount: DECIMAL
  ├─ withholding_tax_amount: DECIMAL
  ├─ paid_amount: DECIMAL
  ├─ payment_due_date: TIMESTAMP
  ├─ bank_details: JSONB
  ├─ created_by: VARCHAR (FK → users.id)
  ├─ owner_id: VARCHAR
  ├─ action_history: JSONB
  ├─ metadata: JSONB
  ├─ created_at: TIMESTAMP
  └─ updated_at: TIMESTAMP
```

---

### Goods Received Note (GRN) Data Flow

**Creation from Purchase Order**:
```
Frontend (CreateGRNRequest)
  ├─ poNumber: string
  ├─ items: GRNItem[]
  │  ├─ description: string
  │  ├─ quantityOrdered: number
  │  ├─ quantityReceived: number
  │  ├─ variance: number
  │  ├─ condition: ItemCondition
  │  └─ notes?: string
  ├─ receivedBy: string
  ├─ warehouseLocation: string
  ├─ notes: string
  └─ createdBy: string
  ↓
Backend Handler
  ├─ Validates PO exists
  ├─ Validates quantities
  ├─ Generates GRNNumber (unique)
  ├─ Creates GoodsReceivedNote model
  ├─ Links to PO
  ├─ Stores in DB
  └─ Returns GoodsReceivedNote
  ↓
Database (goods_received_notes table)
  ├─ id: UUID
  ├─ grn_number: VARCHAR (unique)
  ├─ po_number: VARCHAR (FK → purchase_orders.po_number)
  ├─ items: JSONB
  ├─ quality_issues: JSONB
  ├─ received_by: VARCHAR
  ├─ received_date: TIMESTAMP
  ├─ warehouse_location: VARCHAR
  ├─ notes: TEXT
  ├─ created_by: VARCHAR (FK → users.id)
  ├─ owner_id: VARCHAR
  ├─ approval_stage: INTEGER
  ├─ approval_history: JSONB
  ├─ automation_used: BOOLEAN
  ├─ auto_created_pv: JSONB
  ├─ action_history: JSONB
  ├─ metadata: JSONB
  ├─ created_at: TIMESTAMP
  └─ updated_at: TIMESTAMP
```

---

## Pagination Implementation Details

### Request Parameters

```typescript
// Standard pagination request
GET /api/v1/requisitions?page=1&limit=10&status=pending&department=sales

// Query Parameters
page: number (default: 1)
limit: number (default: 10, max: 100)
status?: string (filter by status)
department?: string (filter by department)
search?: string (search in title/description)
sortBy?: string (field to sort by)
sortOrder?: 'asc' | 'desc' (sort direction)
```

### Response Structure

```typescript
{
  success: true,
  data: [
    { /* requisition 1 */ },
    { /* requisition 2 */ },
    // ... up to limit items
  ],
  pagination: {
    page: 1,
    limit: 10,
    total: 150,
    totalPages: 15,
    hasNext: true,
    hasPrev: false,
    // Aliases for backward compatibility
    page_size: 10,
    totalCount: 150,
    total_pages: 15,
    has_next: true,
    has_prev: false
  }
}
```

### Pagination Calculation

```go
// Backend calculation
totalPages := (total + limit - 1) / limit  // Ceiling division
hasNext := page < totalPages
hasPrev := page > 1
offset := (page - 1) * limit
```

---

## JSONB Field Structures

### RequisitionItem (stored in requisitions.items)

```json
{
  "id": "item-123",
  "description": "Office Supplies",
  "quantity": 100,
  "unitPrice": 25.50,
  "amount": 2550.00,
  "unit": "box",
  "category": "supplies",
  "notes": "Bulk order discount applied"
}
```

### ApprovalHistory (stored in requisitions.approval_history)

```json
[
  {
    "approverId": "user-123",
    "approverName": "John Doe",
    "status": "approved",
    "comments": "Looks good",
    "signature": "base64-encoded-signature",
    "approvedAt": "2024-01-15T10:30:00Z",
    "stageNumber": 1,
    "stageName": "Department Manager",
    "actionTakenBy": "user-123",
    "actionTakenAt": "2024-01-15T10:30:00Z"
  }
]
```

### ActionHistory (stored in requisitions.action_history)

```json
[
  {
    "id": "action-123",
    "action": "created",
    "performedBy": "user-456",
    "performedByName": "Jane Smith",
    "performedByRole": "requester",
    "timestamp": "2024-01-15T09:00:00Z",
    "changes": {
      "status": ["draft", "pending"]
    },
    "comments": "Initial submission",
    "stageNumber": 1,
    "stageName": "Draft"
  }
]
```

### Metadata (stored in requisitions.metadata)

```json
{
  "customField1": "value1",
  "customField2": "value2",
  "internalNotes": "Some internal tracking",
  "externalReference": "EXT-12345",
  "tags": ["urgent", "bulk-order"]
}
```

---

## Database Indexes for Performance

### Requisitions Table Indexes

```sql
-- Primary key
PRIMARY KEY (id)

-- Unique constraints
UNIQUE (organization_id, req_number)

-- Foreign keys
FOREIGN KEY (organization_id) REFERENCES organizations(id)
FOREIGN KEY (created_by) REFERENCES users(id)

-- Performance indexes
INDEX idx_requisitions_organization_id (organization_id)
INDEX idx_requisitions_status (status)
INDEX idx_requisitions_department_id (department_id)
INDEX idx_requisitions_created_by (created_by)
INDEX idx_requisitions_budget_code (budget_code)
INDEX idx_requisitions_cost_center (cost_center)
INDEX idx_requisitions_created_at (created_at DESC)
```

### Purchase Orders Table Indexes

```sql
-- Primary key
PRIMARY KEY (id)

-- Unique constraints
UNIQUE (organization_id, po_number)

-- Foreign keys
FOREIGN KEY (organization_id) REFERENCES organizations(id)
FOREIGN KEY (vendor_id) REFERENCES vendors(id)
FOREIGN KEY (source_requisition_id) REFERENCES requisitions(id)
FOREIGN KEY (created_by) REFERENCES users(id)

-- Performance indexes
INDEX idx_purchase_orders_organization_id (organization_id)
INDEX idx_purchase_orders_status (status)
INDEX idx_purchase_orders_vendor_id (vendor_id)
INDEX idx_purchase_orders_department_id (department_id)
INDEX idx_purchase_orders_created_by (created_by)
INDEX idx_purchase_orders_budget_code (budget_code)
INDEX idx_purchase_orders_cost_center (cost_center)
INDEX idx_purchase_orders_source_requisition_id (source_requisition_id)
INDEX idx_purchase_orders_created_at (created_at DESC)
```

### Payment Vouchers Table Indexes

```sql
-- Primary key
PRIMARY KEY (id)

-- Unique constraints
UNIQUE (organization_id, voucher_number)

-- Foreign keys
FOREIGN KEY (organization_id) REFERENCES organizations(id)
FOREIGN KEY (vendor_id) REFERENCES vendors(id)
FOREIGN KEY (created_by) REFERENCES users(id)

-- Performance indexes
INDEX idx_payment_vouchers_organization_id (organization_id)
INDEX idx_payment_vouchers_status (status)
INDEX idx_payment_vouchers_vendor_id (vendor_id)
INDEX idx_payment_vouchers_department_id (department_id)
INDEX idx_payment_vouchers_created_by (created_by)
INDEX idx_payment_vouchers_budget_code (budget_code)
INDEX idx_payment_vouchers_cost_center (cost_center)
INDEX idx_payment_vouchers_payment_due_date (payment_due_date)
INDEX idx_payment_vouchers_created_at (created_at DESC)
```

---

## Error Handling Patterns

### Validation Errors

```typescript
{
  success: false,
  message: "Validation failed",
  status: 400,
  errors: {
    "title": ["Title is required", "Title must be at least 3 characters"],
    "budgetCode": ["Budget code not found"],
    "items": ["At least one item is required"]
  }
}
```

### Authorization Errors

```typescript
{
  success: false,
  message: "Unauthorized",
  error: "insufficient_permissions",
  status: 403
}
```

### Not Found Errors

```typescript
{
  success: false,
  message: "Requisition not found",
  error: "not_found",
  status: 404
}
```

### Server Errors

```typescript
{
  success: false,
  message: "Internal server error",
  error: "internal_error",
  status: 500
}
```

---

## Transaction Patterns

### Multi-Step Operations

**Requisition to PO to GRN to PV Flow**:

```
Transaction 1: Create Requisition
  ├─ Insert requisition record
  ├─ Create approval task
  └─ Send notification

Transaction 2: Approve Requisition
  ├─ Update requisition status
  ├─ Update approval task
  ├─ Create next approval task (if multi-stage)
  └─ Send notification

Transaction 3: Create Purchase Order
  ├─ Insert purchase order record
  ├─ Link to requisition
  ├─ Create approval task
  └─ Send notification

Transaction 4: Receive Goods (GRN)
  ├─ Insert GRN record
  ├─ Link to PO
  ├─ Update PO status
  ├─ Create approval task
  └─ Send notification

Transaction 5: Create Payment Voucher
  ├─ Insert payment voucher record
  ├─ Link to PO and GRN
  ├─ Create approval task
  └─ Send notification

Transaction 6: Mark as Paid
  ├─ Update PV status
  ├─ Update PO status
  ├─ Record payment details
  └─ Send notification
```

---

## Backward Compatibility Aliases

### Field Name Aliases

```typescript
// Requisition
reqNumber ← REQNumber
requisitionNumber ← REQNumber

// Purchase Order
poNumber ← PONumber
purchaseOrderNumber ← PONumber

// Payment Voucher
voucherNumber ← VoucherNumber
pvNumber ← VoucherNumber
paymentVoucherNumber ← VoucherNumber

// GRN
grnNumber ← GRNNumber
goodsReceivedNoteNumber ← GRNNumber
```

### Status Aliases

```typescript
// Uppercase variants
'DRAFT' ← 'draft'
'PENDING' ← 'pending'
'APPROVED' ← 'approved'
'REJECTED' ← 'rejected'
'COMPLETED' ← 'completed'
'CANCELLED' ← 'cancelled'
```

### Pagination Aliases

```typescript
// Snake case variants
page_size ← pageSize
total_pages ← totalPages
has_next ← hasNext
has_prev ← hasPrev
totalCount ← total
```

---

## Performance Optimization Tips

### Query Optimization

1. **Use Indexes**: Always filter by indexed columns
   ```sql
   -- Good: Uses index
   SELECT * FROM requisitions WHERE organization_id = ? AND status = ?
   
   -- Bad: Full table scan
   SELECT * FROM requisitions WHERE title LIKE ?
   ```

2. **Pagination**: Always use pagination for large result sets
   ```sql
   -- Good: Limited result set
   SELECT * FROM requisitions LIMIT 10 OFFSET 0
   
   -- Bad: Unbounded result set
   SELECT * FROM requisitions
   ```

3. **Selective Columns**: Only fetch needed columns
   ```sql
   -- Good: Specific columns
   SELECT id, req_number, title, status FROM requisitions
   
   -- Bad: All columns including JSONB
   SELECT * FROM requisitions
   ```

### Caching Strategy

1. **Cache Frequently Accessed Data**:
   - User permissions
   - Organization settings
   - Vendor list
   - Budget codes

2. **Cache Duration**:
   - User data: 5 minutes
   - Organization data: 15 minutes
   - Master data: 1 hour
   - Approval tasks: 1 minute

3. **Invalidation**:
   - Invalidate on create/update/delete
   - Invalidate related caches
   - Use cache tags for grouped invalidation

---

## Testing Checklist

### Unit Tests

- [ ] Type definitions compile without errors
- [ ] Request/response types match API contracts
- [ ] Enum values are properly defined
- [ ] Optional fields are properly marked

### Integration Tests

- [ ] Create requisition → Verify in DB
- [ ] Update requisition → Verify changes in DB
- [ ] Approve requisition → Verify approval history
- [ ] Create PO from requisition → Verify linking
- [ ] Create GRN from PO → Verify linking
- [ ] Create PV from GRN → Verify linking
- [ ] Pagination works correctly
- [ ] Filters work correctly
- [ ] Sorting works correctly

### End-to-End Tests

- [ ] Complete requisition workflow
- [ ] Complete PO workflow
- [ ] Complete GRN workflow
- [ ] Complete PV workflow
- [ ] Multi-stage approvals
- [ ] Rejection and return flows
- [ ] Reassignment flows

---

**Last Updated**: 2024
**Version**: 1.0
**Status**: Production Ready
