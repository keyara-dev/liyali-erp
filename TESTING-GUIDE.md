# Testing Guide - Phase 2 Features

## Quick Start

### 1. Import Postman Collection
1. Open Postman
2. Click "Import" → "Upload Files"
3. Select `postman-collection.json`
4. Collection will be imported with all endpoints

### 2. Configure Environment
In Postman, set these variables:
- `base_url`: `http://localhost:8080` (or your backend URL)
- `auth_token`: Leave empty (will be set after login)
- `category_id`: Leave empty (will be set after category creation)
- `vendor_id`: Your existing vendor ID
- `requisition_id`: Leave empty (will be set after requisition creation)

### 3. Run Tests in Order
Follow this sequence in Postman:

```
1. Authentication
   ├─ Login (saves auth_token)

2. Category Management
   ├─ Create Category (saves category_id)
   ├─ List Categories
   ├─ Get Category by ID
   ├─ Get Category Budget Codes
   ├─ Add Budget Code to Category
   ├─ Update Category
   ├─ Delete Category
   └─ Remove Budget Code from Category

3. Requisitions with Enhancements
   ├─ Create Requisition with Category & Supplier (saves requisition_id)
   ├─ List Requisitions
   ├─ Get Requisition by ID (verify new fields)
   └─ Update Requisition

4. User Profile & Last Login
   ├─ Get User Profile
   └─ Login (Check LastLogin in Response)

5. Analytics
   ├─ Get Requisition Metrics (All Time)
   ├─ Get Requisition Metrics (Date Range)
   ├─ Get Requisition Metrics (By Department)
   ├─ Get Requisition Metrics (Weekly)
   ├─ Get Requisition Metrics (Monthly)
   ├─ Get Approval Metrics
   └─ Get Dashboard
```

---

## Unit Tests

### Running Tests

**Run all tests:**
```bash
cd backend
go test ./... -v
```

**Run specific test file:**
```bash
go test -v ./handlers -run TestCreateCategory
```

**Run with coverage:**
```bash
go test ./... -v -cover
```

**Run category tests only:**
```bash
go test -v ./handlers -run Category
```

**Run analytics tests only:**
```bash
go test -v ./services -run Analytics
```

### Test Files

#### 1. `backend/handlers/category_handler_test.go`
Tests all category CRUD operations:
- ✅ TestCreateCategory - Valid creation, missing name, name too short
- ✅ TestGetCategories - List with pagination
- ✅ TestUpdateCategory - Update fields and verify database
- ✅ TestDeleteCategory - Soft deletion verification
- ✅ TestAddBudgetCodeToCategory - Add budget code mappings
- ✅ TestGetCategoryBudgetCodes - Retrieve budget codes
- ✅ TestRemoveBudgetCodeFromCategory - Remove mappings

**Run these tests:**
```bash
go test -v ./handlers -run Category
```

#### 2. `backend/services/analytics_service_test.go`
Tests all analytics calculations:
- ✅ TestGetStatusCounts - Verify status breakdown
- ✅ TestCalculateRejectionRate - Verify rejection percentage
- ✅ TestGetRejectionsOverTime - Verify time-series data
- ✅ TestGetRejectionReasons - Extract and count rejection reasons
- ✅ TestGetTopRejectingApprovers - Calculate approver statistics
- ✅ TestAnalyticsWithDateRange - Filter by date range

**Run these tests:**
```bash
go test -v ./services -run Analytics
```

---

## Manual Testing Scenarios

### Scenario 1: Category Management Workflow

**Create Category:**
```bash
curl -X POST http://localhost:8080/api/v1/categories \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Office Supplies",
    "description": "General office supplies",
    "budgetCodes": ["BDG-001", "BDG-002"]
  }'
```

**Expected Response:**
```json
{
  "success": true,
  "data": {
    "id": "uuid...",
    "name": "Office Supplies",
    "description": "General office supplies",
    "budgetCodes": ["BDG-001", "BDG-002"],
    "active": true,
    "createdAt": "2025-12-24T10:30:00Z",
    "updatedAt": "2025-12-24T10:30:00Z"
  }
}
```

**List Categories:**
```bash
curl -X GET "http://localhost:8080/api/v1/categories?page=1&limit=10&active=true" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**Add Budget Code:**
```bash
curl -X POST http://localhost:8080/api/v1/categories/CATEGORY_ID/budget-codes \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"budgetCode": "BDG-003"}'
```

---

### Scenario 2: Enhanced Requisition Workflow

**Create Requisition with Category:**
```bash
curl -X POST http://localhost:8080/api/v1/requisitions \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Office Supplies Purchase",
    "description": "Monthly office supplies purchase",
    "department": "Finance",
    "priority": "high",
    "items": [
      {"description": "Paper", "quantity": 100, "unitPrice": 5, "amount": 500}
    ],
    "totalAmount": 500,
    "currency": "USD",
    "categoryId": "CATEGORY_ID",
    "preferredVendorId": "VENDOR_ID",
    "isEstimate": true
  }'
```

**Expected Response includes:**
```json
{
  "id": "...",
  "categoryId": "CATEGORY_ID",
  "categoryName": "Office Supplies",
  "preferredVendorId": "VENDOR_ID",
  "preferredVendorName": "Vendor Name",
  "isEstimate": true
}
```

**Get Requisition:**
```bash
curl -X GET http://localhost:8080/api/v1/requisitions/REQUISITION_ID \
  -H "Authorization: Bearer YOUR_TOKEN"
```

---

### Scenario 3: Last Login Tracking

**Login and Check Response:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@liyali.com",
    "password": "admin123"
  }'
```

**Expected Response includes:**
```json
{
  "user": {
    "id": "...",
    "email": "admin@liyali.com",
    "lastLogin": "2025-12-24T10:30:00Z"
  }
}
```

**Verify Database:**
```sql
SELECT id, email, last_login FROM users WHERE email = 'admin@liyali.com';
```

---

### Scenario 4: Analytics Testing

**Get All Metrics:**
```bash
curl -X GET http://localhost:8080/api/v1/analytics/requisitions/metrics \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**Expected Response:**
```json
{
  "success": true,
  "data": {
    "statusCounts": {
      "draft": 5,
      "pending": 3,
      "approved": 10,
      "rejected": 2
    },
    "rejectionRate": 11.76,
    "rejectionsOverTime": [
      {
        "date": "2025-12-24",
        "rejections": 1,
        "total": 5,
        "rate": 20.0
      }
    ],
    "rejectionReasons": [
      {"reason": "Budget exceeded", "count": 1, "percentage": 50.0}
    ],
    "topRejectingApprovers": [
      {
        "approverId": "...",
        "approverName": "Jane Approver",
        "rejections": 2,
        "approvals": 8,
        "rejectionRate": 20.0
      }
    ],
    "totalRequisitions": 20
  }
}
```

**Get Metrics by Date Range:**
```bash
curl -X GET "http://localhost:8080/api/v1/analytics/requisitions/metrics?start_date=2025-12-01&end_date=2025-12-31&period=weekly" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**Get Department-specific Metrics:**
```bash
curl -X GET "http://localhost:8080/api/v1/analytics/requisitions/metrics?department=Finance" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**Get Dashboard:**
```bash
curl -X GET http://localhost:8080/api/v1/analytics/dashboard \
  -H "Authorization: Bearer YOUR_TOKEN"
```

---

## Database Verification

### Check New Tables

```sql
-- Check categories table
SELECT * FROM categories LIMIT 5;

-- Check category_budget_codes table
SELECT * FROM category_budget_codes LIMIT 5;

-- Count categories
SELECT COUNT(*) as total_categories FROM categories;

-- Count category mappings
SELECT COUNT(*) as total_mappings FROM category_budget_codes;
```

### Check Updated Columns

```sql
-- Verify requisitions have new columns
\d requisitions;
-- Should show: category_id, preferred_vendor_id, is_estimate

-- Verify users have last_login
\d users;
-- Should show: last_login

-- Check for NULL last_login values
SELECT id, email, last_login FROM users WHERE last_login IS NOT NULL LIMIT 5;
```

### Verify Data Relationships

```sql
-- Check requisitions with categories
SELECT
  r.id,
  r.title,
  c.name as category_name,
  r.is_estimate
FROM requisitions r
LEFT JOIN categories c ON r.category_id = c.id
WHERE r.category_id IS NOT NULL
LIMIT 5;

-- Check requisitions with preferred vendors
SELECT
  r.id,
  r.title,
  v.name as vendor_name
FROM requisitions r
LEFT JOIN vendors v ON r.preferred_vendor_id = v.id
WHERE r.preferred_vendor_id IS NOT NULL
LIMIT 5;
```

---

## Troubleshooting

### Issue: "Category not found" error when creating requisition

**Solution:**
1. Create a category first using the create category endpoint
2. Copy the returned category ID
3. Use that ID in the requisition creation request

### Issue: Last login is always NULL

**Solution:**
1. Ensure the User model has the `last_login` field
2. Restart the server after database migration
3. Perform a new login - should update the timestamp

### Issue: Analytics returns empty results

**Solution:**
1. Create test requisitions with various statuses
2. Some should have status="rejected"
3. Wait for at least one requisition to be created
4. Analytics aggregates existing data

### Issue: Budget code not found in category

**Solution:**
1. Verify the category ID is correct
2. Check if the budget code was actually created
3. Query `category_budget_codes` table directly

### Issue: Postman shows authentication error

**Solution:**
1. Run Login request first
2. Copy the token from response
3. Set `{{auth_token}}` variable to the token value
4. Try other requests again

---

## Performance Considerations

### For Large Datasets

If you have many requisitions, analytics may take time:

```bash
# Get metrics for a specific date range (faster)
GET /api/v1/analytics/requisitions/metrics?start_date=2025-12-01&end_date=2025-12-31

# Filter by department (faster)
GET /api/v1/analytics/requisitions/metrics?department=Finance
```

### Caching Recommendation

For production, consider caching analytics results:
```go
// Cache for 1 hour
cache.Set("analytics:metrics", metrics, 1*time.Hour)
```

---

## Success Checklist

- [ ] Backend builds without errors
- [ ] Database migrations complete successfully
- [ ] Categories can be created/read/updated/deleted
- [ ] Budget codes can be linked to categories
- [ ] Requisitions accept category & supplier fields
- [ ] Last login timestamps are recorded
- [ ] Analytics endpoints return data
- [ ] All unit tests pass
- [ ] Postman collection tests work end-to-end
- [ ] Database contains expected data

---

## Next Steps

1. **All tests pass** → Move to frontend integration
2. **Frontend ready** → Deploy to staging
3. **Staging approved** → Deploy to production
4. **Live** → Monitor analytics for insights
