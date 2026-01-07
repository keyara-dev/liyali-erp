# API Reference

## Base URL
```
http://localhost:8080/api/v1
```

## Authentication
All endpoints require JWT authentication:
```
Authorization: Bearer <jwt-token>
X-Organization-ID: <organization-id>
```

## Response Format
```json
{
  "success": true,
  "data": {...},
  "message": "Success message",
  "pagination": {
    "page": 1,
    "limit": 10,
    "total": 100,
    "totalPages": 10,
    "hasNext": true,
    "hasPrev": false
  }
}
```

## Authentication Endpoints

### Register User
```http
POST /auth/register
Content-Type: application/json

{
  "name": "John Doe",
  "email": "john@example.com",
  "password": "securepassword"
}
```

### Login
```http
POST /auth/login
Content-Type: application/json

{
  "email": "john@example.com",
  "password": "securepassword"
}
```

### Refresh Token
```http
POST /auth/refresh
Content-Type: application/json

{
  "refreshToken": "refresh-token"
}
```

### Logout
```http
POST /auth/logout
Authorization: Bearer <token>
```

## Organization Endpoints

### List Organizations
```http
GET /organizations
Authorization: Bearer <token>
```

### Create Organization
```http
POST /organizations
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "My Company",
  "description": "Company description"
}
```

### Get Organization
```http
GET /organizations/{id}
Authorization: Bearer <token>
```

### Update Organization
```http
PUT /organizations/{id}
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "Updated Name",
  "description": "Updated description"
}
```

### Organization Members
```http
GET /organizations/{id}/members
POST /organizations/{id}/members
DELETE /organizations/{id}/members/{memberId}
```

## Role Management Endpoints

### List Roles
```http
GET /organizations/{orgId}/roles
Authorization: Bearer <token>
```

### Create Role
```http
POST /organizations/{orgId}/roles
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "Custom Role",
  "description": "Role description",
  "permissions": ["requisitions.create", "requisitions.read"]
}
```

### Update Role
```http
PUT /organizations/{orgId}/roles/{roleId}
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "Updated Role",
  "description": "Updated description"
}
```

### Assign Permission to Role
```http
POST /organizations/{orgId}/roles/{roleId}/permissions
Authorization: Bearer <token>
Content-Type: application/json

{
  "permissionId": "permission-uuid"
}
```

## Requisition Endpoints

### List Requisitions
```http
GET /requisitions?page=1&limit=10&status=pending
Authorization: Bearer <token>
X-Organization-ID: <org-id>
```

### Create Requisition
```http
POST /requisitions
Authorization: Bearer <token>
X-Organization-ID: <org-id>
Content-Type: application/json

{
  "title": "Office Supplies",
  "description": "Monthly office supplies order",
  "department": "Operations",
  "priority": "medium",
  "items": [
    {
      "description": "Paper",
      "quantity": 10,
      "unitPrice": 5.00,
      "amount": 50.00
    }
  ],
  "totalAmount": 50.00,
  "currency": "USD",
  "budgetCode": "OP-2024-001",
  "requiredByDate": "2024-02-15T00:00:00Z"
}
```

### Get Requisition
```http
GET /requisitions/{id}
Authorization: Bearer <token>
X-Organization-ID: <org-id>
```

### Update Requisition
```http
PUT /requisitions/{id}
Authorization: Bearer <token>
X-Organization-ID: <org-id>
Content-Type: application/json

{
  "title": "Updated Title",
  "priority": "high"
}
```

### Approve Requisition
```http
POST /requisitions/{id}/approve
Authorization: Bearer <token>
X-Organization-ID: <org-id>
Content-Type: application/json

{
  "comments": "Approved for processing",
  "signature": "digital-signature"
}
```

### Reject Requisition
```http
POST /requisitions/{id}/reject
Authorization: Bearer <token>
X-Organization-ID: <org-id>
Content-Type: application/json

{
  "remarks": "Insufficient budget",
  "comments": "Please revise budget allocation"
}
```

## Budget Endpoints

### List Budgets
```http
GET /budgets?fiscalYear=2024
Authorization: Bearer <token>
X-Organization-ID: <org-id>
```

### Create Budget
```http
POST /budgets
Authorization: Bearer <token>
X-Organization-ID: <org-id>
Content-Type: application/json

{
  "budgetCode": "OP-2024-001",
  "name": "Operations Budget 2024",
  "department": "Operations",
  "fiscalYear": 2024,
  "totalBudget": 100000.00,
  "currency": "USD"
}
```

## Purchase Order Endpoints

### List Purchase Orders
```http
GET /purchase-orders?status=pending
Authorization: Bearer <token>
X-Organization-ID: <org-id>
```

### Create Purchase Order
```http
POST /purchase-orders
Authorization: Bearer <token>
X-Organization-ID: <org-id>
Content-Type: application/json

{
  "title": "Office Equipment",
  "vendorId": "vendor-uuid",
  "items": [
    {
      "description": "Laptop",
      "quantity": 2,
      "unitPrice": 1000.00,
      "amount": 2000.00
    }
  ],
  "totalAmount": 2000.00,
  "currency": "USD",
  "deliveryDate": "2024-02-20T00:00:00Z"
}
```

## Payment Voucher Endpoints

### List Payment Vouchers
```http
GET /payment-vouchers?status=pending
Authorization: Bearer <token>
X-Organization-ID: <org-id>
```

### Create Payment Voucher
```http
POST /payment-vouchers
Authorization: Bearer <token>
X-Organization-ID: <org-id>
Content-Type: application/json

{
  "voucherNumber": "PV-2024-001",
  "vendorId": "vendor-uuid",
  "invoiceNumber": "INV-001",
  "amount": 1500.00,
  "currency": "USD",
  "paymentMethod": "bank_transfer",
  "description": "Payment for services"
}
```

## Vendor Endpoints

### List Vendors
```http
GET /vendors?active=true
Authorization: Bearer <token>
X-Organization-ID: <org-id>
```

### Create Vendor
```http
POST /vendors
Authorization: Bearer <token>
X-Organization-ID: <org-id>
Content-Type: application/json

{
  "name": "ABC Supplies",
  "email": "contact@abcsupplies.com",
  "phone": "+1234567890",
  "address": "123 Business St, City, State 12345",
  "category": "Office Supplies"
}
```

## Analytics Endpoints

### Dashboard Metrics
```http
GET /analytics/dashboard?startDate=2024-01-01&endDate=2024-12-31
Authorization: Bearer <token>
X-Organization-ID: <org-id>
```

### Requisition Metrics
```http
GET /analytics/requisitions/metrics?period=monthly
Authorization: Bearer <token>
X-Organization-ID: <org-id>
```

### Approval Metrics
```http
GET /analytics/approvals/metrics?department=Operations
Authorization: Bearer <token>
X-Organization-ID: <org-id>
```

## Query Parameters

### Pagination
- `page`: Page number (default: 1)
- `limit`: Items per page (default: 10, max: 100)

### Filtering
- `status`: Filter by status (draft, pending, approved, rejected)
- `department`: Filter by department
- `priority`: Filter by priority (low, medium, high, urgent)
- `startDate`: Start date for date range
- `endDate`: End date for date range

### Sorting
- `sortBy`: Field to sort by
- `sortOrder`: Sort order (asc, desc)

## Error Responses

### 400 Bad Request
```json
{
  "success": false,
  "message": "Validation error",
  "errors": [
    {
      "field": "email",
      "message": "Invalid email format"
    }
  ]
}
```

### 401 Unauthorized
```json
{
  "success": false,
  "message": "Invalid or expired token"
}
```

### 403 Forbidden
```json
{
  "success": false,
  "message": "Insufficient permissions"
}
```

### 404 Not Found
```json
{
  "success": false,
  "message": "Resource not found"
}
```

### 500 Internal Server Error
```json
{
  "success": false,
  "message": "Internal server error"
}
```

## Rate Limiting

- **Authentication endpoints**: 5 requests per minute per IP
- **General endpoints**: 100 requests per minute per user
- **Bulk operations**: 10 requests per minute per user

## Postman Collection

Import the provided `postman-collection.json` for pre-configured API tests with:
- Authentication flow
- All CRUD operations
- Error scenarios
- Sample data

---

**Next**: See [Development Guide](DEVELOPMENT.md) for testing and development workflow