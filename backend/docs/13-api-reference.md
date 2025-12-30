# API Reference

Complete API reference for the Liyali Gateway Backend.

## API Overview

The Liyali Gateway Backend provides a comprehensive RESTful API with:

- **Base URL**: `http://localhost:8080/api/v1`
- **Authentication**: JWT Bearer tokens
- **Content Type**: `application/json`
- **Response Format**: Consistent JSON responses
- **Pagination**: Cursor and offset-based pagination
- **Error Handling**: Standardized error responses

## Authentication

### Register User

```http
POST /api/v1/auth/register
Content-Type: application/json

{
  "email": "user@company.com",
  "password": "SecurePassword123!",
  "name": "John Doe",
  "organizationId": "org-uuid"
}
```

**Response:**
```json
{
  "success": true,
  "message": "User registered successfully",
  "data": {
    "user": {
      "id": "user-uuid",
      "email": "user@company.com",
      "name": "John Doe",
      "organizationId": "org-uuid",
      "roles": ["user"],
      "createdAt": "2024-01-01T10:00:00Z"
    },
    "tokens": {
      "accessToken": "jwt-access-token",
      "refreshToken": "jwt-refresh-token",
      "expiresIn": 86400
    }
  }
}
```

### Login

```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "user@company.com",
  "password": "SecurePassword123!"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "user": {
      "id": "user-uuid",
      "email": "user@company.com",
      "name": "John Doe",
      "organizationId": "org-uuid",
      "roles": ["user", "manager"],
      "permissions": ["requisitions.create", "requisitions.read"]
    },
    "tokens": {
      "accessToken": "jwt-access-token",
      "refreshToken": "jwt-refresh-token",
      "expiresIn": 86400
    }
  }
}
```

### Refresh Token

```http
POST /api/v1/auth/refresh
Content-Type: application/json

{
  "refreshToken": "jwt-refresh-token"
}
```

### Logout

```http
POST /api/v1/auth/logout
Authorization: Bearer jwt-access-token
```

## Organizations

### Create Organization

```http
POST /api/v1/organizations
Authorization: Bearer jwt-access-token
Content-Type: application/json

{
  "name": "Acme Corporation",
  "description": "Leading technology company",
  "address": "123 Business St, City, State 12345",
  "phone": "+1-555-0123",
  "email": "contact@acme.com",
  "website": "https://acme.com"
}
```

### Get Organizations

```http
GET /api/v1/organizations?page=1&limit=20
Authorization: Bearer jwt-access-token
```

### Get Organization by ID

```http
GET /api/v1/organizations/{id}
Authorization: Bearer jwt-access-token
```

### Update Organization

```http
PUT /api/v1/organizations/{id}
Authorization: Bearer jwt-access-token
Content-Type: application/json

{
  "name": "Updated Organization Name",
  "description": "Updated description"
}
```

## Users

### Create User

```http
POST /api/v1/users
Authorization: Bearer jwt-access-token
Content-Type: application/json

{
  "email": "newuser@company.com",
  "name": "New User",
  "roles": ["user"],
  "department": "IT",
  "position": "Developer"
}
```

### Get Users

```http
GET /api/v1/users?page=1&limit=20&department=IT&role=user
Authorization: Bearer jwt-access-token
```

### Get User by ID

```http
GET /api/v1/users/{id}
Authorization: Bearer jwt-access-token
```

### Update User

```http
PUT /api/v1/users/{id}
Authorization: Bearer jwt-access-token
Content-Type: application/json

{
  "name": "Updated Name",
  "department": "Finance",
  "roles": ["user", "manager"]
}
```

## Requisitions

### Create Requisition

```http
POST /api/v1/requisitions
Authorization: Bearer jwt-access-token
Content-Type: application/json

{
  "title": "Office Supplies Request",
  "description": "Monthly office supplies order",
  "items": [
    {
      "name": "Laptop",
      "description": "MacBook Pro 16-inch",
      "quantity": 2,
      "unitPrice": 1200.00,
      "totalPrice": 2400.00,
      "category": "Electronics",
      "specifications": {
        "brand": "Apple",
        "model": "MacBook Pro",
        "memory": "16GB",
        "storage": "512GB SSD"
      }
    }
  ],
  "totalAmount": 2400.00,
  "priority": "medium",
  "department": "IT",
  "justification": "Replacement for old laptops",
  "expectedDeliveryDate": "2024-02-01T00:00:00Z",
  "budgetCode": "IT-2024-001"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Requisition created successfully",
  "data": {
    "id": "req-uuid",
    "documentNumber": "REQ-2024-001",
    "organizationId": "org-uuid",
    "title": "Office Supplies Request",
    "description": "Monthly office supplies order",
    "status": "draft",
    "priority": "medium",
    "department": "IT",
    "totalAmount": 2400.00,
    "currency": "USD",
    "items": [
      {
        "id": "item-uuid",
        "name": "Laptop",
        "description": "MacBook Pro 16-inch",
        "quantity": 2,
        "unitPrice": 1200.00,
        "totalPrice": 2400.00,
        "category": "Electronics"
      }
    ],
    "createdBy": "user-uuid",
    "createdAt": "2024-01-01T10:00:00Z",
    "updatedAt": "2024-01-01T10:00:00Z"
  }
}
```

### Get Requisitions

```http
GET /api/v1/requisitions?page=1&limit=20&status=pending&department=IT&minAmount=1000&maxAmount=5000&startDate=2024-01-01&endDate=2024-01-31&sort=createdAt&order=desc
Authorization: Bearer jwt-access-token
```

**Query Parameters:**
- `page` - Page number (default: 1)
- `limit` - Items per page (default: 20, max: 100)
- `status` - Filter by status (draft, submitted, approved, rejected, completed)
- `priority` - Filter by priority (low, medium, high, urgent)
- `department` - Filter by department
- `createdBy` - Filter by creator
- `minAmount` - Minimum amount filter
- `maxAmount` - Maximum amount filter
- `startDate` - Created after date (ISO 8601)
- `endDate` - Created before date (ISO 8601)
- `sort` - Sort field (createdAt, updatedAt, totalAmount, title)
- `order` - Sort order (asc, desc)

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": "req-uuid",
      "documentNumber": "REQ-2024-001",
      "title": "Office Supplies Request",
      "status": "approved",
      "priority": "medium",
      "totalAmount": 2400.00,
      "createdBy": "user-uuid",
      "createdAt": "2024-01-01T10:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "total": 25,
    "totalPages": 3,
    "pageSize": 20,
    "hasNext": true,
    "hasPrev": false
  }
}
```

### Get Requisition by ID

```http
GET /api/v1/requisitions/{id}
Authorization: Bearer jwt-access-token
```

### Update Requisition

```http
PUT /api/v1/requisitions/{id}
Authorization: Bearer jwt-access-token
Content-Type: application/json

{
  "title": "Updated Office Supplies Request",
  "description": "Updated monthly office supplies order",
  "priority": "high",
  "expectedDeliveryDate": "2024-01-15T00:00:00Z"
}
```

### Submit Requisition for Approval

```http
POST /api/v1/requisitions/{id}/submit
Authorization: Bearer jwt-access-token
Content-Type: application/json

{
  "comment": "Ready for approval"
}
```

### Delete Requisition

```http
DELETE /api/v1/requisitions/{id}
Authorization: Bearer jwt-access-token
```

## Budgets

### Create Budget

```http
POST /api/v1/budgets
Authorization: Bearer jwt-access-token
Content-Type: application/json

{
  "name": "IT Department Budget 2024",
  "description": "Annual budget for IT department",
  "totalAmount": 100000.00,
  "allocatedAmount": 0.00,
  "spentAmount": 0.00,
  "fiscalYear": "2024",
  "department": "IT",
  "startDate": "2024-01-01T00:00:00Z",
  "endDate": "2024-12-31T23:59:59Z",
  "categories": [
    {
      "name": "Hardware",
      "allocatedAmount": 60000.00,
      "description": "Computer hardware and equipment"
    },
    {
      "name": "Software",
      "allocatedAmount": 40000.00,
      "description": "Software licenses and subscriptions"
    }
  ]
}
```

### Get Budgets

```http
GET /api/v1/budgets?page=1&limit=20&fiscalYear=2024&department=IT
Authorization: Bearer jwt-access-token
```

### Get Budget by ID

```http
GET /api/v1/budgets/{id}
Authorization: Bearer jwt-access-token
```

### Update Budget

```http
PUT /api/v1/budgets/{id}
Authorization: Bearer jwt-access-token
Content-Type: application/json

{
  "name": "Updated IT Budget 2024",
  "totalAmount": 120000.00
}
```

## Payment Vouchers

### Create Payment Voucher

```http
POST /api/v1/payment-vouchers
Authorization: Bearer jwt-access-token
Content-Type: application/json

{
  "purchaseOrderId": "po-uuid",
  "vendorId": "vendor-uuid",
  "title": "Payment for MacBook Purchase",
  "description": "Payment voucher for approved purchase order",
  "amount": 4998.00,
  "currency": "USD",
  "paymentMethod": "bank_transfer",
  "dueDate": "2024-02-15T00:00:00Z",
  "bankDetails": {
    "accountName": "Vendor Company Ltd",
    "accountNumber": "1234567890",
    "routingNumber": "021000021",
    "bankName": "First National Bank"
  },
  "invoiceNumber": "INV-2024-001",
  "taxAmount": 399.84,
  "notes": "Payment for approved laptop purchase"
}
```

### Get Payment Vouchers

```http
GET /api/v1/payment-vouchers?page=1&limit=20&status=pending&vendorId=vendor-uuid&minAmount=1000&maxAmount=10000
Authorization: Bearer jwt-access-token
```

### Get Payment Voucher by ID

```http
GET /api/v1/payment-vouchers/{id}
Authorization: Bearer jwt-access-token
```

### Update Payment Voucher

```http
PUT /api/v1/payment-vouchers/{id}
Authorization: Bearer jwt-access-token
Content-Type: application/json

{
  "dueDate": "2024-02-10T00:00:00Z",
  "paymentMethod": "check",
  "notes": "Updated payment method to check"
}
```

### Approve Payment Voucher

```http
POST /api/v1/payment-vouchers/{id}/approve
Authorization: Bearer jwt-access-token
Content-Type: application/json

{
  "comment": "Payment approved for processing"
}
```

### Reject Payment Voucher

```http
POST /api/v1/payment-vouchers/{id}/reject
Authorization: Bearer jwt-access-token
Content-Type: application/json

{
  "comment": "Invoice details need verification",
  "reason": "documentation_incomplete"
}
```

## Goods Received Notes (GRN)

### Create GRN

```http
POST /api/v1/grns
Authorization: Bearer jwt-access-token
Content-Type: application/json

{
  "purchaseOrderId": "po-uuid",
  "title": "MacBook Delivery Receipt",
  "description": "Goods received note for laptop delivery",
  "receivedDate": "2024-02-10T14:30:00Z",
  "receivedBy": "John Doe",
  "items": [
    {
      "name": "MacBook Pro 16-inch",
      "orderedQuantity": 2,
      "receivedQuantity": 2,
      "condition": "good",
      "serialNumbers": ["ABC123456", "ABC123457"],
      "notes": "All items received in good condition"
    }
  ],
  "deliveryNote": "DN-2024-001",
  "carrier": "FedEx",
  "trackingNumber": "1234567890",
  "notes": "Delivered to IT department reception"
}
```

### Get GRNs

```http
GET /api/v1/grns?page=1&limit=20&status=pending&purchaseOrderId=po-uuid
Authorization: Bearer jwt-access-token
```

### Get GRN by ID

```http
GET /api/v1/grns/{id}
Authorization: Bearer jwt-access-token
```

### Update GRN

```http
PUT /api/v1/grns/{id}
Authorization: Bearer jwt-access-token
Content-Type: application/json

{
  "notes": "Updated delivery notes",
  "condition": "damaged",
  "damageReport": "Minor scratches on packaging"
}
```

### Approve GRN

```http
POST /api/v1/grns/{id}/approve
Authorization: Bearer jwt-access-token
Content-Type: application/json

{
  "comment": "Goods received and verified"
}
```

### Reject GRN

```http
POST /api/v1/grns/{id}/reject
Authorization: Bearer jwt-access-token
Content-Type: application/json

{
  "comment": "Items damaged during delivery",
  "reason": "damaged_goods"
}
```

## Categories

### Create Category

```http
POST /api/v1/categories
Authorization: Bearer jwt-access-token
Content-Type: application/json

{
  "name": "Office Supplies",
  "description": "General office supplies and materials",
  "budgetCodes": ["BDG-001", "BDG-002"],
  "parentCategoryId": null,
  "active": true
}
```

### Get Categories

```http
GET /api/v1/categories?page=1&limit=20&active=true&parentId=parent-uuid
Authorization: Bearer jwt-access-token
```

### Get Category by ID

```http
GET /api/v1/categories/{id}
Authorization: Bearer jwt-access-token
```

### Update Category

```http
PUT /api/v1/categories/{id}
Authorization: Bearer jwt-access-token
Content-Type: application/json

{
  "name": "Updated Office Supplies",
  "description": "Updated description",
  "active": true
}
```

### Delete Category

```http
DELETE /api/v1/categories/{id}
Authorization: Bearer jwt-access-token
```

### Get Category Budget Codes

```http
GET /api/v1/categories/{id}/budget-codes
Authorization: Bearer jwt-access-token
```

### Add Budget Code to Category

```http
POST /api/v1/categories/{id}/budget-codes
Authorization: Bearer jwt-access-token
Content-Type: application/json

{
  "budgetCode": "BDG-003"
}
```

### Remove Budget Code from Category

```http
DELETE /api/v1/categories/{id}/budget-codes/{budgetCode}
Authorization: Bearer jwt-access-token
```

## Vendors

### Create Vendor

```http
POST /api/v1/vendors
Authorization: Bearer jwt-access-token
Content-Type: application/json

{
  "name": "Apple Inc.",
  "description": "Technology hardware and software vendor",
  "contactPerson": "John Smith",
  "email": "sales@apple.com",
  "phone": "+1-800-APL-CARE",
  "address": {
    "street": "One Apple Park Way",
    "city": "Cupertino",
    "state": "CA",
    "zipCode": "95014",
    "country": "USA"
  },
  "taxId": "94-2404110",
  "paymentTerms": "Net 30",
  "categories": ["Electronics", "Software"],
  "active": true
}
```

### Get Vendors

```http
GET /api/v1/vendors?page=1&limit=20&active=true&category=Electronics
Authorization: Bearer jwt-access-token
```

### Get Vendor by ID

```http
GET /api/v1/vendors/{id}
Authorization: Bearer jwt-access-token
```

### Update Vendor

```http
PUT /api/v1/vendors/{id}
Authorization: Bearer jwt-access-token
Content-Type: application/json

{
  "contactPerson": "Jane Doe",
  "email": "sales-updated@apple.com",
  "paymentTerms": "Net 15"
}
```

## Approvals

### Get Approval Tasks

```http
GET /api/v1/approvals?page=1&limit=20&status=pending&assignedTo=current&documentType=requisition
Authorization: Bearer jwt-access-token
```

**Query Parameters:**
- `page` - Page number
- `limit` - Items per page
- `status` - Filter by status (pending, approved, rejected)
- `assignedTo` - Filter by assignee (current, user-id)
- `documentType` - Filter by document type
- `priority` - Filter by priority
- `overdue` - Filter overdue tasks (true/false)

### Get Approval Task by ID

```http
GET /api/v1/approvals/{id}
Authorization: Bearer jwt-access-token
```

### Approve Task

```http
POST /api/v1/approvals/{id}/approve
Authorization: Bearer jwt-access-token
Content-Type: application/json

{
  "comment": "Approved with conditions",
  "conditions": "Must be delivered by end of month",
  "delegateTo": null
}
```

### Reject Task

```http
POST /api/v1/approvals/{id}/reject
Authorization: Bearer jwt-access-token
Content-Type: application/json

{
  "comment": "Budget constraints - please reduce amount",
  "reason": "budget_exceeded"
}
```

### Reassign Task

```http
POST /api/v1/approvals/{id}/reassign
Authorization: Bearer jwt-access-token
Content-Type: application/json

{
  "assignedTo": "user-uuid",
  "comment": "Reassigning to department manager"
}
```

### Get Overdue Tasks

```http
GET /api/v1/approvals/tasks/overdue?page=1&limit=20
Authorization: Bearer jwt-access-token
```

### Bulk Approve

```http
POST /api/v1/approvals/bulk/approve
Authorization: Bearer jwt-access-token
Content-Type: application/json

{
  "approvalIds": [
    "approval-uuid-1",
    "approval-uuid-2",
    "approval-uuid-3"
  ],
  "comment": "Bulk approval for Q1 requisitions"
}
```

### Bulk Reject

```http
POST /api/v1/approvals/bulk/reject
Authorization: Bearer jwt-access-token
Content-Type: application/json

{
  "approvalIds": [
    "approval-uuid-1",
    "approval-uuid-2"
  ],
  "comment": "Budget constraints for these items",
  "reason": "budget_exceeded"
}
```

### Bulk Reassign

```http
POST /api/v1/approvals/bulk/reassign
Authorization: Bearer jwt-access-token
Content-Type: application/json

{
  "approvalIds": [
    "approval-uuid-1",
    "approval-uuid-2"
  ],
  "assignedTo": "user-uuid",
  "comment": "Reassigning to senior manager"
}
```

## Workflows

### Get Workflows

```http
GET /api/v1/workflows?documentType=requisition&active=true&isDefault=false
Authorization: Bearer jwt-access-token
```

### Get Workflow by ID

```http
GET /api/v1/workflows/{id}
Authorization: Bearer jwt-access-token
```

### Get Default Workflow

```http
GET /api/v1/workflows/default/{documentType}
Authorization: Bearer jwt-access-token
```

### Create Workflow

```http
POST /api/v1/workflows
Authorization: Bearer jwt-access-token
Content-Type: application/json

{
  "name": "IT Requisition Approval Workflow",
  "description": "Standard approval process for IT requisitions",
  "documentType": "requisition",
  "isDefault": false,
  "stages": [
    {
      "name": "Department Manager Review",
      "description": "Initial review by department manager",
      "order": 1,
      "stageType": "approval",
      "requiredRole": "manager",
      "isParallel": false,
      "isOptional": false,
      "timeoutHours": 48,
      "escalationRole": "org_admin"
    },
    {
      "name": "Finance Review",
      "description": "Financial review for amounts over $1000",
      "order": 2,
      "stageType": "approval",
      "requiredRole": "finance_manager",
      "isParallel": false,
      "isOptional": false,
      "timeoutHours": 24,
      "conditions": [
        {
          "field": "totalAmount",
          "operator": "greater_than",
          "value": "1000"
        }
      ]
    }
  ]
}
```

### Update Workflow

```http
PUT /api/v1/workflows/{id}
Authorization: Bearer jwt-access-token
Content-Type: application/json

{
  "name": "Updated IT Workflow",
  "description": "Updated workflow with new stages"
}
```

### Activate Workflow

```http
POST /api/v1/workflows/{id}/activate
Authorization: Bearer jwt-access-token
```

### Deactivate Workflow

```http
POST /api/v1/workflows/{id}/deactivate
Authorization: Bearer jwt-access-token
```

### Delete Workflow

```http
DELETE /api/v1/workflows/{id}
Authorization: Bearer jwt-access-token
```

## Notifications

### Get Notifications

```http
GET /api/v1/notifications?page=1&limit=20&unread=true&type=approval
Authorization: Bearer jwt-access-token
```

### Get Notification by ID

```http
GET /api/v1/notifications/{id}
Authorization: Bearer jwt-access-token
```

### Mark Notification as Read

```http
PUT /api/v1/notifications/{id}/read
Authorization: Bearer jwt-access-token
```

### Mark All Notifications as Read

```http
PUT /api/v1/notifications/read-all
Authorization: Bearer jwt-access-token
```

### Get Notification Statistics

```http
GET /api/v1/notifications/stats
Authorization: Bearer jwt-access-token
```

**Response:**
```json
{
  "success": true,
  "data": {
    "total": 25,
    "unread": 8,
    "byType": {
      "approval": 12,
      "workflow": 8,
      "system": 5
    }
  }
}
```

### Delete Notification

```http
DELETE /api/v1/notifications/{id}
Authorization: Bearer jwt-access-token
```

## Audit Logs

### Get Audit Logs

```http
GET /api/v1/audit-logs?page=1&limit=20&action=create&entityType=requisition&userId=user-uuid&startDate=2024-01-01&endDate=2024-01-31
Authorization: Bearer jwt-access-token
```

**Query Parameters:**
- `page` - Page number
- `limit` - Items per page
- `action` - Filter by action (create, update, delete, approve, reject)
- `entityType` - Filter by entity type (requisition, budget, purchase_order, etc.)
- `userId` - Filter by user who performed the action
- `startDate` - Filter by date range start
- `endDate` - Filter by date range end

### Get Document Audit Logs

```http
GET /api/v1/audit-logs/document/{documentId}?page=1&limit=20
Authorization: Bearer jwt-access-token
```

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": "audit-uuid",
      "action": "create",
      "entityType": "requisition",
      "entityId": "req-uuid",
      "userId": "user-uuid",
      "userName": "John Doe",
      "timestamp": "2024-01-01T10:00:00Z",
      "ipAddress": "192.168.1.100",
      "userAgent": "Mozilla/5.0...",
      "changes": {
        "title": {
          "old": null,
          "new": "Office Supplies Request"
        },
        "totalAmount": {
          "old": null,
          "new": 1500.00
        }
      },
      "metadata": {
        "correlationId": "req-123e4567-e89b-12d3-a456-426614174000"
      }
    }
  ],
  "pagination": {
    "page": 1,
    "total": 15,
    "totalPages": 1,
    "pageSize": 20,
    "hasNext": false,
    "hasPrev": false
  }
}
```

## Documents (Generic)

### Search Documents

```http
GET /api/v1/documents/search?q=laptop&type=requisition&status=approved&minAmount=1000&maxAmount=5000&page=1&limit=20
Authorization: Bearer jwt-access-token
```

**Query Parameters:**
- `q` - Search query (full-text search)
- `type` - Document type filter
- `status` - Status filter
- `priority` - Priority filter
- `createdBy` - Creator filter
- `assignedTo` - Assignee filter
- `minAmount` - Minimum amount filter
- `maxAmount` - Maximum amount filter
- `startDate` - Created after date
- `endDate` - Created before date
- `tags` - Tag filter (comma-separated)
- `sort` - Sort field (relevance, date, amount, title)
- `order` - Sort order (asc, desc)
- `page` - Page number
- `limit` - Results per page

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": "doc-uuid",
      "documentType": "requisition",
      "documentNumber": "REQ-2024-001",
      "title": "Laptop Purchase Request",
      "description": "MacBook Pro for development team",
      "status": "approved",
      "priority": "medium",
      "totalAmount": 2400.00,
      "currency": "USD",
      "createdBy": "user-uuid",
      "createdAt": "2024-01-01T10:00:00Z",
      "relevanceScore": 0.95,
      "highlights": {
        "title": ["<mark>Laptop</mark> Purchase Request"],
        "description": ["MacBook Pro for development team"]
      }
    }
  ],
  "facets": {
    "documentType": {
      "requisition": 15,
      "budget": 5,
      "purchase_order": 8
    },
    "status": {
      "approved": 12,
      "pending": 8,
      "rejected": 3
    }
  },
  "pagination": {
    "page": 1,
    "total": 28,
    "totalPages": 2,
    "pageSize": 20,
    "hasNext": true,
    "hasPrev": false
  }
}
```

### Get Document by ID

```http
GET /api/v1/documents/{id}
Authorization: Bearer jwt-access-token
```

### Get Documents by Type

```http
GET /api/v1/documents/type/{documentType}?page=1&limit=20
Authorization: Bearer jwt-access-token
```

### Document Statistics

```http
GET /api/v1/documents/stats
Authorization: Bearer jwt-access-token
```

**Response:**
```json
{
  "success": true,
  "data": {
    "totalDocuments": 150,
    "documentsByType": {
      "requisition": 45,
      "budget": 20,
      "purchase_order": 35,
      "payment_voucher": 25,
      "grn": 15,
      "category": 8,
      "vendor": 2
    },
    "documentsByStatus": {
      "draft": 20,
      "submitted": 15,
      "under_review": 25,
      "approved": 60,
      "rejected": 10,
      "completed": 15,
      "cancelled": 5
    },
    "totalValue": 125000.00,
    "averageValue": 833.33
  }
}
```

## Workflows

### Create Workflow Template

```http
POST /api/v1/workflows/templates
Authorization: Bearer jwt-access-token
Content-Type: application/json

{
  "name": "Requisition Approval Workflow",
  "description": "Standard approval process for purchase requisitions",
  "documentType": "requisition",
  "stages": [
    {
      "name": "Department Manager Review",
      "description": "Initial review by department manager",
      "order": 1,
      "stageType": "approval",
      "requiredRole": "manager",
      "isParallel": false,
      "isOptional": false,
      "timeoutHours": 48,
      "escalationRole": "org_admin"
    },
    {
      "name": "Finance Review",
      "description": "Financial review for amounts over $1000",
      "order": 2,
      "stageType": "approval",
      "requiredRole": "finance_manager",
      "isParallel": false,
      "isOptional": false,
      "timeoutHours": 24,
      "conditions": [
        {
          "field": "totalAmount",
          "operator": "greater_than",
          "value": "1000"
        }
      ]
    }
  ]
}
```

### Get Workflow Templates

```http
GET /api/v1/workflows/templates?documentType=requisition&active=true
Authorization: Bearer jwt-access-token
```

### Start Workflow

```http
POST /api/v1/workflows/instances
Authorization: Bearer jwt-access-token
Content-Type: application/json

{
  "workflowTemplateId": "workflow-template-uuid",
  "documentId": "document-uuid",
  "comment": "Starting approval process for laptop purchase"
}
```

### Execute Workflow Stage

```http
POST /api/v1/workflows/instances/{id}/execute
Authorization: Bearer jwt-access-token
Content-Type: application/json

{
  "stageId": "stage-uuid",
  "decision": "approved",
  "comment": "Approved with minor modifications",
  "delegateTo": null
}
```

**Possible Decisions:**
- `approved` - Approve and move to next stage
- `rejected` - Reject and end workflow
- `delegated` - Delegate to another user
- `escalated` - Escalate to higher authority
- `returned` - Return to previous stage for modifications

### Get Pending Approvals

```http
GET /api/v1/workflows/pending?assignedTo=current&page=1&limit=20
Authorization: Bearer jwt-access-token
```

### Bulk Approval

```http
POST /api/v1/workflows/bulk/approve
Authorization: Bearer jwt-access-token
Content-Type: application/json

{
  "workflowInstanceIds": [
    "workflow-instance-uuid-1",
    "workflow-instance-uuid-2",
    "workflow-instance-uuid-3"
  ],
  "comment": "Bulk approval for Q1 requisitions",
  "decision": "approved"
}
```

## Analytics

### Dashboard Analytics

```http
GET /api/v1/analytics/dashboard
Authorization: Bearer jwt-access-token
```

### Document Analytics

```http
GET /api/v1/analytics/documents?period=month&groupBy=type&startDate=2024-01-01&endDate=2024-01-31
Authorization: Bearer jwt-access-token
```

### Workflow Analytics

```http
GET /api/v1/analytics/workflows?templateId=workflow-uuid&period=month
Authorization: Bearer jwt-access-token
```

### User Analytics

```http
GET /api/v1/analytics/users/{userId}?period=month
Authorization: Bearer jwt-access-token
```

## Error Responses

### Standard Error Format

```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "details": [
      {
        "field": "email",
        "message": "Email is required"
      },
      {
        "field": "password",
        "message": "Password must be at least 8 characters"
      }
    ]
  },
  "timestamp": "2024-01-01T10:00:00Z",
  "path": "/api/v1/auth/register"
}
```

### HTTP Status Codes

- `200` - OK
- `201` - Created
- `204` - No Content
- `400` - Bad Request
- `401` - Unauthorized
- `403` - Forbidden
- `404` - Not Found
- `409` - Conflict
- `422` - Unprocessable Entity
- `423` - Locked (Account locked)
- `429` - Too Many Requests
- `500` - Internal Server Error

### Common Error Codes

- `VALIDATION_ERROR` - Request validation failed
- `AUTHENTICATION_ERROR` - Authentication failed
- `AUTHORIZATION_ERROR` - Insufficient permissions
- `NOT_FOUND` - Resource not found
- `CONFLICT` - Resource conflict (duplicate)
- `ACCOUNT_LOCKED` - Account is locked
- `RATE_LIMIT_EXCEEDED` - Too many requests
- `INTERNAL_ERROR` - Internal server error

## Rate Limiting

The API implements rate limiting to prevent abuse:

- **Authentication endpoints**: 5 requests per minute per IP
- **General API endpoints**: 100 requests per minute per user
- **Search endpoints**: 50 requests per minute per user
- **Bulk operations**: 10 requests per minute per user

Rate limit headers are included in responses:
```http
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1640995200
```

## Pagination

### Offset-based Pagination

```http
GET /api/v1/requisitions?page=1&limit=20
```

**Response:**
```json
{
  "data": [...],
  "pagination": {
    "page": 1,
    "total": 100,
    "totalPages": 5,
    "pageSize": 20,
    "hasNext": true,
    "hasPrev": false
  }
}
```

### Cursor-based Pagination

```http
GET /api/v1/documents/search?cursor=eyJjcmVhdGVkX2F0IjoiMjAyNC0wMS0wMVQxMDowMDowMFoifQ&limit=20
```

**Response:**
```json
{
  "data": [...],
  "pagination": {
    "nextCursor": "eyJjcmVhdGVkX2F0IjoiMjAyNC0wMS0wMVQxMTowMDowMFoifQ",
    "hasMore": true,
    "limit": 20
  }
}
```

## Webhooks

### Register Webhook

```http
POST /api/v1/webhooks
Authorization: Bearer jwt-access-token
Content-Type: application/json

{
  "url": "https://your-app.com/webhooks/liyali",
  "events": [
    "requisition.created",
    "requisition.approved",
    "workflow.completed"
  ],
  "secret": "your-webhook-secret"
}
```

### Webhook Events

- `requisition.created` - New requisition created
- `requisition.updated` - Requisition updated
- `requisition.approved` - Requisition approved
- `requisition.rejected` - Requisition rejected
- `workflow.started` - Workflow started
- `workflow.completed` - Workflow completed
- `user.created` - New user created
- `organization.updated` - Organization updated

### Webhook Payload

```json
{
  "event": "requisition.approved",
  "timestamp": "2024-01-01T10:00:00Z",
  "data": {
    "id": "req-uuid",
    "title": "Office Supplies Request",
    "status": "approved",
    "totalAmount": 2400.00,
    "approvedBy": "manager-uuid",
    "approvedAt": "2024-01-01T10:00:00Z"
  },
  "organization": {
    "id": "org-uuid",
    "name": "Acme Corporation"
  }
}
```

## SDK Examples

### JavaScript/Node.js

```javascript
const LiyaliGateway = require('@liyali/gateway-sdk');

const client = new LiyaliGateway({
  baseURL: 'http://localhost:8080/api/v1',
  apiKey: 'your-api-key'
});

// Create requisition
const requisition = await client.requisitions.create({
  title: 'Office Supplies',
  description: 'Monthly supplies order',
  totalAmount: 1500.00,
  items: [
    {
      name: 'Laptop',
      quantity: 1,
      unitPrice: 1500.00,
      totalPrice: 1500.00
    }
  ]
});

// Search documents
const results = await client.documents.search({
  query: 'laptop',
  type: 'requisition',
  status: 'approved'
});
```

### Python

```python
from liyali_gateway import LiyaliGateway

client = LiyaliGateway(
    base_url='http://localhost:8080/api/v1',
    api_key='your-api-key'
)

# Create requisition
requisition = client.requisitions.create({
    'title': 'Office Supplies',
    'description': 'Monthly supplies order',
    'totalAmount': 1500.00,
    'items': [
        {
            'name': 'Laptop',
            'quantity': 1,
            'unitPrice': 1500.00,
            'totalPrice': 1500.00
        }
    ]
})

# Search documents
results = client.documents.search(
    query='laptop',
    type='requisition',
    status='approved'
)
```

## OpenAPI Specification

The complete OpenAPI 3.0 specification is available at:
- **JSON**: `http://localhost:8080/api/v1/openapi.json`
- **YAML**: `http://localhost:8080/api/v1/openapi.yaml`
- **Swagger UI**: `http://localhost:8080/docs`

## Postman Collection

Import the Postman collection for easy API testing:
- **Collection**: `docs/postman-collection.json`
- **Environment**: `docs/postman-environment.json`

## Next Steps

- **Development**: Set up [Development Environment](./11-development.md)
- **Testing**: Implement [API Testing](./12-testing.md)
- **Deployment**: Deploy to [Production](./14-deployment.md)
- **Monitoring**: Set up [API Monitoring](./15-monitoring.md)