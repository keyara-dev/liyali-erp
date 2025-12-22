# CRUD Operations Testing Guide

**Phase 12C: CRUD Handlers Implementation Complete**

This guide provides detailed instructions for testing all CRUD endpoints for the five main workflow documents (Requisition, Budget, Purchase Order, Payment Voucher, GRN) and Vendor management.

---

## Prerequisites

1. Backend API running on `http://localhost:8080`
2. Valid JWT token from authentication endpoints
3. Pre-seeded test data (users and vendors automatically created)

### Obtaining a Test Token

```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@liyali.com",
    "password": "any_password"
  }'
```

Save the returned `token` value for use in subsequent requests.

---

## Requisition CRUD Operations

### 1. Create Requisition
**Endpoint**: `POST /api/requisitions`
**Authentication**: Required
**Status**: Draft

```bash
curl -X POST http://localhost:8080/api/requisitions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "title": "Office Supplies Requisition",
    "description": "Monthly office supplies requisition for Q4 2025 including stationery and printing materials",
    "department": "Administration",
    "priority": "medium",
    "items": [
      {
        "description": "A4 Paper (500 sheets)",
        "quantity": 10,
        "unitPrice": 5.50,
        "amount": 55.00
      },
      {
        "description": "Pens (Box of 50)",
        "quantity": 5,
        "unitPrice": 12.00,
        "amount": 60.00
      }
    ],
    "totalAmount": 115.00,
    "currency": "USD"
  }'
```

**Success Response (201)**:
```json
{
  "success": true,
  "data": {
    "id": "req-uuid-here",
    "requesterId": "user-admin-001",
    "requesterName": "Admin User",
    "title": "Office Supplies Requisition",
    "description": "Monthly office supplies...",
    "department": "Administration",
    "status": "draft",
    "priority": "medium",
    "items": [...],
    "totalAmount": 115.00,
    "currency": "USD",
    "approvalStage": 0,
    "approvalHistory": [],
    "createdAt": "2025-12-22T21:30:00Z",
    "updatedAt": "2025-12-22T21:30:00Z"
  }
}
```

### 2. Get Requisitions (List)
**Endpoint**: `GET /api/requisitions`
**Authentication**: Required
**Query Parameters**:
- `page` (default: 1)
- `limit` (default: 10)
- `status` (optional: draft, pending, approved, rejected)
- `department` (optional)
- `priority` (optional: low, medium, high)

```bash
curl -X GET "http://localhost:8080/api/requisitions?page=1&limit=10&status=draft" \
  -H "Authorization: Bearer <token>"
```

### 3. Get Requisition (Detail)
**Endpoint**: `GET /api/requisitions/:id`

```bash
curl -X GET http://localhost:8080/api/requisitions/req-uuid-here \
  -H "Authorization: Bearer <token>"
```

### 4. Update Requisition
**Endpoint**: `PUT /api/requisitions/:id`
**Note**: Only draft and pending requisitions can be updated

```bash
curl -X PUT http://localhost:8080/api/requisitions/req-uuid-here \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "title": "Updated Office Supplies Requisition",
    "totalAmount": 125.00
  }'
```

### 5. Delete Requisition
**Endpoint**: `DELETE /api/requisitions/:id`
**Note**: Only draft requisitions can be deleted

```bash
curl -X DELETE http://localhost:8080/api/requisitions/req-uuid-here \
  -H "Authorization: Bearer <token>"
```

### 6. Approve Requisition
**Endpoint**: `POST /api/requisitions/:id/approve`

```bash
curl -X POST http://localhost:8080/api/requisitions/req-uuid-here/approve \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "comments": "Approved for procurement",
    "signature": "John_Approver_20251222"
  }'
```

### 7. Reject Requisition
**Endpoint**: `POST /api/requisitions/:id/reject`

```bash
curl -X POST http://localhost:8080/api/requisitions/req-uuid-here/reject \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "remarks": "Please clarify the budget allocation for this requisition",
    "signature": "John_Approver_20251222"
  }'
```

### 8. Reassign Requisition
**Endpoint**: `POST /api/requisitions/:id/reassign`

```bash
curl -X POST http://localhost:8080/api/requisitions/req-uuid-here/reassign \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "newApproverId": "user-approver-001",
    "reason": "Escalating to senior approver"
  }'
```

---

## Budget CRUD Operations

### 1. Create Budget
**Endpoint**: `POST /api/budgets`

```bash
curl -X POST http://localhost:8080/api/budgets \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "budgetCode": "BDG-2025-Q4-001",
    "department": "IT",
    "fiscalYear": "2025",
    "totalBudget": 50000.00,
    "allocatedAmount": 20000.00
  }'
```

### 2. Get Budgets (List)
```bash
curl -X GET "http://localhost:8080/api/budgets?fiscalYear=2025&department=IT" \
  -H "Authorization: Bearer <token>"
```

### 3. Get Budget (Detail)
```bash
curl -X GET http://localhost:8080/api/budgets/budget-uuid-here \
  -H "Authorization: Bearer <token>"
```

### 4. Update Budget
```bash
curl -X PUT http://localhost:8080/api/budgets/budget-uuid-here \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "allocatedAmount": 25000.00
  }'
```

### 5. Delete Budget
```bash
curl -X DELETE http://localhost:8080/api/budgets/budget-uuid-here \
  -H "Authorization: Bearer <token>"
```

### 6. Approve Budget
```bash
curl -X POST http://localhost:8080/api/budgets/budget-uuid-here/approve \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "comments": "Budget approved",
    "signature": "Finance_Officer_20251222"
  }'
```

### 7. Reject Budget
```bash
curl -X POST http://localhost:8080/api/budgets/budget-uuid-here/reject \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "remarks": "Budget allocation exceeds departmental limits",
    "signature": "Finance_Officer_20251222"
  }'
```

---

## Purchase Order CRUD Operations

### 1. Create Purchase Order
**Endpoint**: `POST /api/purchase-orders`
**Note**: Requires valid vendor ID

```bash
curl -X POST http://localhost:8080/api/purchase-orders \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "vendorId": "vendor-001",
    "items": [
      {
        "description": "Laptop Dell XPS 15",
        "quantity": 2,
        "unitPrice": 1200.00,
        "amount": 2400.00
      }
    ],
    "totalAmount": 2400.00,
    "currency": "USD",
    "deliveryDate": "2025-12-31T00:00:00Z",
    "linkedRequisition": "req-uuid-here"
  }'
```

### 2. Get Purchase Orders (List)
```bash
curl -X GET "http://localhost:8080/api/purchase-orders?status=draft&vendorId=vendor-001" \
  -H "Authorization: Bearer <token>"
```

### 3. Get Purchase Order (Detail)
```bash
curl -X GET http://localhost:8080/api/purchase-orders/po-uuid-here \
  -H "Authorization: Bearer <token>"
```

### 4. Update Purchase Order
```bash
curl -X PUT http://localhost:8080/api/purchase-orders/po-uuid-here \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "deliveryDate": "2025-12-29T00:00:00Z"
  }'
```

### 5. Delete Purchase Order
```bash
curl -X DELETE http://localhost:8080/api/purchase-orders/po-uuid-here \
  -H "Authorization: Bearer <token>"
```

### 6. Approve Purchase Order
```bash
curl -X POST http://localhost:8080/api/purchase-orders/po-uuid-here/approve \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "comments": "PO approved for vendor ABC Supplies",
    "signature": "Approver_Signature_20251222"
  }'
```

### 7. Reject Purchase Order
```bash
curl -X POST http://localhost:8080/api/purchase-orders/po-uuid-here/reject \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "remarks": "Vendor pricing not competitive. Please obtain alternative quotes.",
    "signature": "Approver_Signature_20251222"
  }'
```

---

## Payment Voucher CRUD Operations

### 1. Create Payment Voucher
**Endpoint**: `POST /api/payment-vouchers`

```bash
curl -X POST http://localhost:8080/api/payment-vouchers \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "vendorId": "vendor-001",
    "invoiceNumber": "INV-2025-12-001",
    "amount": 2400.00,
    "currency": "USD",
    "paymentMethod": "bank_transfer",
    "glCode": "6100",
    "description": "Payment for laptop purchase order PO-1234567890",
    "linkedPO": "po-uuid-here"
  }'
```

### 2. Get Payment Vouchers (List)
```bash
curl -X GET "http://localhost:8080/api/payment-vouchers?status=draft" \
  -H "Authorization: Bearer <token>"
```

### 3. Get Payment Voucher (Detail)
```bash
curl -X GET http://localhost:8080/api/payment-vouchers/pv-uuid-here \
  -H "Authorization: Bearer <token>"
```

### 4. Update Payment Voucher
```bash
curl -X PUT http://localhost:8080/api/payment-vouchers/pv-uuid-here \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "paymentMethod": "check"
  }'
```

### 5. Delete Payment Voucher
```bash
curl -X DELETE http://localhost:8080/api/payment-vouchers/pv-uuid-here \
  -H "Authorization: Bearer <token>"
```

### 6. Approve Payment Voucher
```bash
curl -X POST http://localhost:8080/api/payment-vouchers/pv-uuid-here/approve \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "comments": "Payment approved. Process bank transfer.",
    "signature": "Finance_Officer_20251222"
  }'
```

### 7. Reject Payment Voucher
```bash
curl -X POST http://localhost:8080/api/payment-vouchers/pv-uuid-here/reject \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "remarks": "Invoice does not match PO details. Please resubmit with corrections.",
    "signature": "Finance_Officer_20251222"
  }'
```

---

## GRN (Goods Received Note) CRUD Operations

### 1. Create GRN
**Endpoint**: `POST /api/grns`
**Note**: Requires valid PO number

```bash
curl -X POST http://localhost:8080/api/grns \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "poNumber": "PO-1703350800-abcd1234",
    "items": [
      {
        "description": "Laptop Dell XPS 15",
        "quantityOrdered": 2,
        "quantityReceived": 2,
        "variance": 0,
        "condition": "good"
      }
    ],
    "receivedBy": "Warehouse Manager"
  }'
```

### 2. Get GRNs (List)
```bash
curl -X GET "http://localhost:8080/api/grns?status=draft" \
  -H "Authorization: Bearer <token>"
```

### 3. Get GRN (Detail)
```bash
curl -X GET http://localhost:8080/api/grns/grn-uuid-here \
  -H "Authorization: Bearer <token>"
```

### 4. Update GRN
```bash
curl -X PUT http://localhost:8080/api/grns/grn-uuid-here \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "qualityIssues": [
      {
        "itemDescription": "Laptop Dell XPS 15",
        "issueType": "cosmetic_damage",
        "description": "Minor scratch on keyboard",
        "severity": "low"
      }
    ]
  }'
```

### 5. Delete GRN
```bash
curl -X DELETE http://localhost:8080/api/grns/grn-uuid-here \
  -H "Authorization: Bearer <token>"
```

### 6. Approve GRN
```bash
curl -X POST http://localhost:8080/api/grns/grn-uuid-here/approve \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "comments": "Goods received and verified",
    "signature": "Warehouse_Manager_20251222"
  }'
```

### 7. Reject GRN
```bash
curl -X POST http://localhost:8080/api/grns/grn-uuid-here/reject \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "remarks": "Significant quantity variance detected. Units received: 1, Expected: 2",
    "signature": "Warehouse_Manager_20251222"
  }'
```

---

## Vendor CRUD Operations

### 1. Create Vendor
**Endpoint**: `POST /api/vendors`

```bash
curl -X POST http://localhost:8080/api/vendors \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "name": "Tech Solutions Ltd",
    "email": "sales@techsolutions.com",
    "phone": "+1-555-0123",
    "country": "United States",
    "city": "Boston",
    "bankAccount": "1122334455",
    "taxId": "12-3456789"
  }'
```

### 2. Get Vendors (List)
```bash
curl -X GET "http://localhost:8080/api/vendors?active=true&country=United%20States" \
  -H "Authorization: Bearer <token>"
```

### 3. Get Vendor (Detail)
```bash
curl -X GET http://localhost:8080/api/vendors/vendor-uuid-here \
  -H "Authorization: Bearer <token>"
```

### 4. Update Vendor
```bash
curl -X PUT http://localhost:8080/api/vendors/vendor-uuid-here \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "phone": "+1-555-0124",
    "bankAccount": "5544332211"
  }'
```

### 5. Delete Vendor
**Note**: Deactivates vendor (soft delete)

```bash
curl -X DELETE http://localhost:8080/api/vendors/vendor-uuid-here \
  -H "Authorization: Bearer <token>"
```

---

## Common Error Scenarios

### 400 Bad Request - Missing Required Field
```json
{
  "success": false,
  "message": "Title is required and must be at least 3 characters"
}
```

### 401 Unauthorized - Invalid Token
```json
{
  "success": false,
  "message": "Invalid authorization header"
}
```

### 404 Not Found - Document Not Found
```json
{
  "success": false,
  "message": "Requisition not found"
}
```

### 409 Conflict - Duplicate Entry
```json
{
  "success": false,
  "message": "Vendor with this email already exists"
}
```

### 422 Unprocessable - Business Logic Error
```json
{
  "success": false,
  "message": "Cannot update requisition in approved status"
}
```

---

## Testing Workflow Complete Example

### Step 1: Create Requisition
```bash
REQ_ID=$(curl -s -X POST http://localhost:8080/api/requisitions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '...' | jq -r '.data.id')
```

### Step 2: Get and Verify
```bash
curl -X GET http://localhost:8080/api/requisitions/$REQ_ID \
  -H "Authorization: Bearer $TOKEN" | jq
```

### Step 3: Approve
```bash
curl -X POST http://localhost:8080/api/requisitions/$REQ_ID/approve \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '...' | jq
```

### Step 4: Create Purchase Order
```bash
curl -X POST http://localhost:8080/api/purchase-orders \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{\"linkedRequisition\": \"$REQ_ID\", ...}" | jq
```

---

## Notes

- All timestamps are in ISO 8601 format (UTC)
- Approval stage increments with each approval
- Approval history maintains complete audit trail
- Document status follows workflow: draft → pending → approved/rejected
- Some operations require specific roles (defined in middleware)
- Rate limiting and additional security can be configured per environment

---

**Last Updated**: December 22, 2025
**Status**: Phase 12C Complete - All CRUD handlers implemented and tested
