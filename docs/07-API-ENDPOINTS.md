# API Endpoints Documentation

**Status**: Complete Architecture with Proposed Backend Endpoints
**Last Updated**: 2025-12-12
**Phase**: Phase 11 (localStorage) → Phase 12 (PostgreSQL with REST APIs)

---

## Overview

This document defines all API endpoints required for the Liyali Gateway workflow approval system. Currently, the application uses **localStorage as the data source** (Phase 11). When migrating to Phase 12 (PostgreSQL backend), these endpoints should be implemented.

### Current Architecture
- **Data Source**: Browser localStorage (JSON-based)
- **Data Access**: Direct function calls (storage hooks)
- **State Management**: React Query with client-side mutations
- **Limitations**: Single-user, single-device, no real persistence

### Future Architecture (Phase 12)
- **Data Source**: PostgreSQL database
- **Server**: Node.js/Express or similar backend
- **Authentication**: OAuth 2.0 / JWT tokens
- **Real Persistence**: Multi-user, multi-device, cloud-ready

---

## 1. Document Management Endpoints

### 1.1 Purchase Orders

#### GET /api/purchase-orders
Retrieve all purchase orders with optional filtering and pagination.

**Query Parameters:**
```
limit=10              // Number of records per page (default: 10, max: 100)
page=1                // Page number (default: 1)
status=ALL            // Filter by status: ALL, DRAFT, SUBMITTED, IN_REVIEW, APPROVED, REJECTED, REVERSED
creatorId=user-123    // Filter by creator user ID
vendorId=VENDOR-001   // Filter by vendor
startDate=2024-01-01  // Filter by creation date (ISO format)
endDate=2024-12-31    // Filter by creation date (ISO format)
sortBy=createdAt      // Sort field: createdAt, documentNumber, status, totalAmount
sortOrder=DESC        // Sort order: ASC or DESC
search=PO-2024        // Free-text search on documentNumber or vendorName
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "documents": [
      {
        "id": "po-550e8400-e29b-41d4-a716-446655440000",
        "type": "PURCHASE_ORDER",
        "documentNumber": "PO-2024-001",
        "status": "APPROVED",
        "currentStage": 4,
        "createdBy": "user-1",
        "createdByUser": {
          "id": "user-1",
          "name": "John Mwale",
          "email": "john@example.com",
          "role": "requester"
        },
        "createdAt": "2024-12-01T10:30:00Z",
        "updatedAt": "2024-12-10T15:45:00Z",
        "metadata": {
          "vendorName": "Mitete Supplies Ltd",
          "vendorId": "VENDOR-001",
          "totalAmount": 18750,
          "currency": "ZMW",
          "deliveryDate": "2024-12-31",
          "items": [
            {
              "id": "item-001",
              "description": "Office Chairs - Ergonomic",
              "quantity": 15,
              "unitCost": 450,
              "totalCost": 6750
            }
          ]
        }
      }
    ],
    "pagination": {
      "total": 42,
      "page": 1,
      "limit": 10,
      "totalPages": 5
    }
  },
  "message": "Purchase orders retrieved successfully"
}
```

#### POST /api/purchase-orders
Create a new purchase order.

**Request Body:**
```json
{
  "documentNumber": "PO-2024-NEW",
  "vendorName": "New Vendor Ltd",
  "vendorId": "VENDOR-NEW",
  "createdBy": "user-1",
  "totalAmount": 50000,
  "currency": "ZMW",
  "deliveryDate": "2025-01-15",
  "items": [
    {
      "description": "Bulk Item 1",
      "quantity": 100,
      "unitCost": 500,
      "totalCost": 50000
    }
  ],
  "specialInstructions": "Rush delivery required"
}
```

**Response (201 Created):**
```json
{
  "success": true,
  "data": {
    "id": "po-550e8400-e29b-41d4-a716-446655440001",
    "type": "PURCHASE_ORDER",
    "documentNumber": "PO-2024-NEW",
    "status": "DRAFT",
    "currentStage": 0,
    "createdBy": "user-1",
    "createdAt": "2025-12-12T10:00:00Z",
    "updatedAt": "2025-12-12T10:00:00Z",
    "metadata": {
      "vendorName": "New Vendor Ltd",
      "vendorId": "VENDOR-NEW",
      "totalAmount": 50000,
      "currency": "ZMW",
      "deliveryDate": "2025-01-15",
      "items": [
        {
          "id": "item-new-001",
          "description": "Bulk Item 1",
          "quantity": 100,
          "unitCost": 500,
          "totalCost": 50000
        }
      ]
    }
  },
  "message": "Purchase order created successfully"
}
```

#### GET /api/purchase-orders/:id
Retrieve a specific purchase order by ID.

**Response (200 OK):** (Same format as POST response)

#### PUT /api/purchase-orders/:id
Update an existing purchase order (only DRAFT status allowed).

**Request Body:**
```json
{
  "vendorName": "Updated Vendor",
  "totalAmount": 55000,
  "items": [],
  "specialInstructions": "Updated instructions"
}
```

**Response (200 OK):** (Updated document object)

#### DELETE /api/purchase-orders/:id
Delete a purchase order (only DRAFT status allowed).

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Purchase order deleted successfully"
}
```

---

### 1.2 Requisitions

#### GET /api/requisitions
Retrieve all requisitions with filtering and pagination.

**Query Parameters:** (Same as purchase orders)

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "documents": [
      {
        "id": "req-550e8400-e29b-41d4-a716-446655440000",
        "type": "REQUISITION",
        "documentNumber": "REQ-2024-001",
        "status": "IN_REVIEW",
        "currentStage": 2,
        "createdBy": "user-2",
        "createdByUser": {
          "id": "user-2",
          "name": "Sarah Banda",
          "email": "sarah@example.com",
          "role": "requester"
        },
        "createdAt": "2024-11-28T08:15:00Z",
        "updatedAt": "2024-12-05T14:20:00Z",
        "metadata": {
          "department": "IT",
          "amount": 75000,
          "currency": "ZMW",
          "justification": "Need for upgraded hardware",
          "items": [
            {
              "id": "item-req-001",
              "description": "Laptop Computers",
              "quantity": 5,
              "estimatedCost": 75000
            }
          ]
        }
      }
    ],
    "pagination": {
      "total": 28,
      "page": 1,
      "limit": 10,
      "totalPages": 3
    }
  }
}
```

#### POST /api/requisitions
Create a new requisition.

**Request Body:**
```json
{
  "documentNumber": "REQ-2024-NEW",
  "department": "Finance",
  "amount": 100000,
  "currency": "ZMW",
  "justification": "Equipment upgrade required",
  "createdBy": "user-2",
  "items": [
    {
      "description": "Office Equipment",
      "quantity": 10,
      "estimatedCost": 100000
    }
  ]
}
```

**Response (201 Created):** (Same format as purchase orders)

#### GET /api/requisitions/:id
Retrieve a specific requisition.

#### PUT /api/requisitions/:id
Update a requisition (only DRAFT status).

#### DELETE /api/requisitions/:id
Delete a requisition (only DRAFT status).

---

### 1.3 Payment Vouchers

#### GET /api/payment-vouchers
Retrieve all payment vouchers with filtering and pagination.

**Query Parameters:** (Same as purchase orders)

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "documents": [
      {
        "id": "pv-550e8400-e29b-41d4-a716-446655440000",
        "type": "PAYMENT_VOUCHER",
        "documentNumber": "PV-2024-001",
        "status": "APPROVED",
        "currentStage": 4,
        "createdBy": "user-3",
        "createdByUser": {
          "id": "user-3",
          "name": "Finance Officer",
          "email": "finance@example.com",
          "role": "finance"
        },
        "createdAt": "2024-12-05T09:00:00Z",
        "updatedAt": "2024-12-10T16:30:00Z",
        "metadata": {
          "payeeName": "Service Provider Ltd",
          "payeeAccount": "123456789",
          "amount": 50000,
          "currency": "ZMW",
          "invoiceNumber": "INV-2024-500",
          "paymentDate": "2024-12-15",
          "description": "Services rendered in November"
        }
      }
    ],
    "pagination": {
      "total": 156,
      "page": 1,
      "limit": 10,
      "totalPages": 16
    }
  }
}
```

#### POST /api/payment-vouchers
Create a new payment voucher.

**Request Body:**
```json
{
  "documentNumber": "PV-2024-NEW",
  "payeeName": "Vendor Company",
  "payeeAccount": "987654321",
  "amount": 75000,
  "currency": "ZMW",
  "invoiceNumber": "INV-2024-600",
  "paymentDate": "2025-01-05",
  "description": "December services invoice",
  "createdBy": "user-3"
}
```

**Response (201 Created):** (Same format as purchase orders)

#### GET /api/payment-vouchers/:id
Retrieve a specific payment voucher.

#### PUT /api/payment-vouchers/:id
Update a payment voucher (only DRAFT status).

#### DELETE /api/payment-vouchers/:id
Delete a payment voucher (only DRAFT status).

---

### 1.4 Goods Received Notes (GRN)

#### GET /api/goods-received-notes
Retrieve all GRNs with filtering and pagination.

**Query Parameters:**
```
limit=10
page=1
status=ALL              // DRAFT, IN_REVIEW, APPROVED
creatorId=user-123
poId=po-550e8400       // Filter by related Purchase Order
startDate=2024-01-01
endDate=2024-12-31
search=GRN-2024
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "documents": [
      {
        "id": "grn-550e8400-e29b-41d4-a716-446655440000",
        "type": "GOODS_RECEIVED_NOTE",
        "documentNumber": "GRN-2024-001",
        "status": "APPROVED",
        "currentStage": 3,
        "createdBy": "user-4",
        "createdByUser": {
          "id": "user-4",
          "name": "Warehouse Manager",
          "email": "warehouse@example.com",
          "role": "warehouse"
        },
        "createdAt": "2024-12-10T11:00:00Z",
        "updatedAt": "2024-12-11T09:15:00Z",
        "metadata": {
          "poId": "po-550e8400-e29b-41d4-a716-446655440000",
          "poNumber": "PO-2024-001",
          "vendorName": "Mitete Supplies Ltd",
          "receivedQuantity": 15,
          "totalQuantity": 15,
          "amount": 6750,
          "receivedDate": "2024-12-10",
          "warehouseLocation": "Shelf A-12"
        }
      }
    ],
    "pagination": {
      "total": 89,
      "page": 1,
      "limit": 10,
      "totalPages": 9
    }
  }
}
```

#### POST /api/goods-received-notes
Create a new GRN.

**Request Body:**
```json
{
  "documentNumber": "GRN-2024-NEW",
  "poId": "po-550e8400-e29b-41d4-a716-446655440001",
  "poNumber": "PO-2024-NEW",
  "vendorName": "New Vendor Ltd",
  "receivedQuantity": 100,
  "totalQuantity": 100,
  "amount": 50000,
  "receivedDate": "2025-12-12",
  "warehouseLocation": "Shelf B-05",
  "createdBy": "user-4"
}
```

**Response (201 Created):** (Same format as purchase orders)

#### GET /api/goods-received-notes/:id
Retrieve a specific GRN.

#### PUT /api/goods-received-notes/:id
Update a GRN (only DRAFT status).

#### DELETE /api/goods-received-notes/:id
Delete a GRN (only DRAFT status).

---

## 2. Search & Filter Endpoints

### GET /api/search
Unified search endpoint for all document types.

**Query Parameters:**
```
documentNumber=PO-2024       // Partial match (case-insensitive)
documentType=ALL             // ALL, PURCHASE_ORDER, REQUISITION, PAYMENT_VOUCHER, GOODS_RECEIVED_NOTE
status=APPROVED              // ALL, DRAFT, SUBMITTED, IN_REVIEW, APPROVED, REJECTED, REVERSED
startDate=2024-01-01         // ISO format (includes entire start date)
endDate=2024-12-31           // ISO format (includes entire end date)
creatorId=user-123           // Filter by creator
limit=10                      // Results per page
page=1                        // Page number
sortBy=createdAt              // Sort field: createdAt, documentNumber, status, totalAmount
sortOrder=DESC                // ASC or DESC
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "results": [
      {
        "id": "po-550e8400-e29b-41d4-a716-446655440000",
        "type": "PURCHASE_ORDER",
        "documentNumber": "PO-2024-001",
        "status": "APPROVED",
        "createdAt": "2024-12-01T10:30:00Z",
        "amount": 18750,
        "creator": "John Mwale"
      },
      {
        "id": "req-550e8400-e29b-41d4-a716-446655440000",
        "type": "REQUISITION",
        "documentNumber": "REQ-2024-001",
        "status": "IN_REVIEW",
        "createdAt": "2024-11-28T08:15:00Z",
        "amount": 75000,
        "creator": "Sarah Banda"
      }
    ],
    "pagination": {
      "total": 156,
      "page": 1,
      "limit": 10,
      "totalPages": 16
    }
  },
  "message": "Search completed successfully"
}
```

---

## 3. Workflow & Approval Endpoints

### GET /api/approvals/tasks
Get approval tasks assigned to the current user.

**Query Parameters:**
```
status=pending           // pending, approved, rejected
limit=10
page=1
sortBy=createdAt
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "tasks": [
      {
        "id": "task-550e8400",
        "documentId": "po-550e8400-e29b-41d4-a716-446655440001",
        "documentNumber": "PO-2024-002",
        "documentType": "PURCHASE_ORDER",
        "status": "pending",
        "currentStage": 1,
        "assignedTo": "user-5",
        "assignedToName": "Approver Name",
        "createdAt": "2024-12-08T13:00:00Z",
        "dueDate": "2024-12-15",
        "metadata": {
          "vendorName": "TechXpress Solutions",
          "amount": 43250,
          "requester": "Sarah Banda"
        }
      }
    ],
    "pagination": {
      "total": 24,
      "page": 1,
      "limit": 10,
      "totalPages": 3
    }
  }
}
```

### POST /api/approvals/tasks/:taskId/approve
Approve a task.

**Request Body:**
```json
{
  "approverId": "user-5",
  "signature": "base64-encoded-signature-image",
  "comments": "Approved as requested",
  "stageNumber": 1
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "taskId": "task-550e8400",
    "documentId": "po-550e8400-e29b-41d4-a716-446655440001",
    "previousStatus": "pending",
    "newStatus": "approved",
    "nextStage": 2,
    "message": "Task approved successfully"
  }
}
```

### POST /api/approvals/tasks/:taskId/reject
Reject a task.

**Request Body:**
```json
{
  "rejectingUserId": "user-5",
  "signature": "base64-encoded-signature-image",
  "remarks": "Requires vendor clarification"
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "taskId": "task-550e8400",
    "documentId": "po-550e8400-e29b-41d4-a716-446655440001",
    "status": "rejected",
    "message": "Task rejected successfully"
  }
}
```

### POST /api/approvals/tasks/:taskId/reassign
Reassign a task to another approver.

**Request Body:**
```json
{
  "reassignedBy": "user-5",
  "newApproverId": "user-6",
  "newApproverName": "Secondary Approver",
  "reason": "Delegating to subject matter expert"
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "taskId": "task-550e8400",
    "previousAssignee": "user-5",
    "newAssignee": "user-6",
    "message": "Task reassigned successfully"
  }
}
```

---

## 4. Bulk Operations Endpoints

### POST /api/approvals/bulk/approve
Approve multiple tasks at once.

**Request Body:**
```json
{
  "taskIds": [
    "task-550e8400",
    "task-550e8401",
    "task-550e8402"
  ],
  "approverId": "user-5",
  "signature": "base64-encoded-signature",
  "comments": "Bulk approval - routine orders"
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "approved": 3,
    "failed": 0,
    "results": [
      {
        "taskId": "task-550e8400",
        "status": "success",
        "message": "Approved"
      }
    ]
  },
  "message": "3 tasks approved successfully"
}
```

### POST /api/approvals/bulk/reject
Reject multiple tasks at once.

**Request Body:**
```json
{
  "taskIds": [
    "task-550e8400",
    "task-550e8401"
  ],
  "rejectingUserId": "user-5",
  "signature": "base64-encoded-signature",
  "remarks": "Budget constraints"
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "rejected": 2,
    "failed": 0
  },
  "message": "2 tasks rejected successfully"
}
```

### POST /api/approvals/bulk/reassign
Reassign multiple tasks to another approver.

**Request Body:**
```json
{
  "taskIds": ["task-550e8400", "task-550e8401"],
  "reassignedBy": "user-5",
  "newApproverId": "user-6",
  "newApproverName": "Secondary Approver",
  "reason": "Delegation due to workload"
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "reassigned": 2,
    "failed": 0
  },
  "message": "2 tasks reassigned successfully"
}
```

---

## 5. Analytics & Reporting Endpoints

### GET /api/analytics/dashboard
Get dashboard metrics and KPIs.

**Query Parameters:**
```
startDate=2024-01-01
endDate=2024-12-31
groupBy=week           // day, week, month
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "summary": {
      "totalDocuments": 425,
      "totalAmount": 8750000,
      "averageApprovalTime": "3.2 days",
      "approvalsInProgress": 24,
      "approvalsPending": 156
    },
    "topMetrics": {
      "highestAmountDocument": {
        "documentNumber": "PO-2024-007",
        "amount": 125000,
        "type": "PURCHASE_ORDER"
      },
      "slowestApprovals": [
        {
          "documentNumber": "REQ-2024-005",
          "daysInApproval": 8
        }
      ]
    },
    "trends": {
      "dailyApprovals": [
        {
          "date": "2024-12-01",
          "approved": 5,
          "rejected": 1,
          "pending": 8
        }
      ],
      "documentTypeBreakdown": {
        "PURCHASE_ORDER": 150,
        "REQUISITION": 125,
        "PAYMENT_VOUCHER": 100,
        "GOODS_RECEIVED_NOTE": 50
      }
    }
  }
}
```

### GET /api/analytics/bottlenecks
Identify approval workflow bottlenecks.

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "bottlenecks": [
      {
        "stage": 2,
        "documentType": "PAYMENT_VOUCHER",
        "pendingCount": 18,
        "averageDaysWaiting": 4.5,
        "recommendedAction": "Add approvers to this stage"
      }
    ]
  }
}
```

---

## 6. Download & Export Endpoints

### GET /api/documents/:documentId/download
Download a document as PDF.

**Response**: PDF file download

### POST /api/exports/documents
Export documents in various formats.

**Request Body:**
```json
{
  "documentIds": ["po-550e8400", "req-550e8400"],
  "format": "excel",     // excel, csv, pdf
  "includeMetadata": true,
  "dateRange": {
    "startDate": "2024-01-01",
    "endDate": "2024-12-31"
  }
}
```

**Response**: File download or stream

---

## 7. User & Role Management Endpoints

### GET /api/users
List all users in the system.

**Query Parameters:**
```
role=approver          // requester, approver, admin, finance, warehouse
department=IT
active=true
limit=50
page=1
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "users": [
      {
        "id": "user-1",
        "name": "John Mwale",
        "email": "john@example.com",
        "role": "requester",
        "department": "Operations",
        "active": true,
        "createdAt": "2024-01-15T08:00:00Z"
      }
    ],
    "pagination": {
      "total": 52,
      "page": 1,
      "limit": 50,
      "totalPages": 2
    }
  }
}
```

### GET /api/users/:userId
Get a specific user's details.

### PUT /api/users/:userId
Update user information and roles.

---

## 8. System & Health Endpoints

### GET /api/health
Health check endpoint.

**Response (200 OK):**
```json
{
  "success": true,
  "status": "healthy",
  "timestamp": "2025-12-12T10:30:00Z",
  "version": "1.0.0"
}
```

### GET /api/config
Get system configuration (public data only).

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "appName": "Liyali Gateway",
    "version": "1.0.0",
    "documentTypes": [
      "PURCHASE_ORDER",
      "REQUISITION",
      "PAYMENT_VOUCHER",
      "GOODS_RECEIVED_NOTE"
    ],
    "statuses": [
      "DRAFT",
      "SUBMITTED",
      "IN_REVIEW",
      "APPROVED",
      "REJECTED",
      "REVERSED"
    ],
    "currencies": ["ZMW", "USD", "EUR"],
    "maxUploadSize": 10485760
  }
}
```

---

## Error Handling

All endpoints follow this error response format:

**400 Bad Request:**
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid request parameters",
    "details": [
      {
        "field": "status",
        "message": "Invalid status value"
      }
    ]
  }
}
```

**401 Unauthorized:**
```json
{
  "success": false,
  "error": {
    "code": "UNAUTHORIZED",
    "message": "Authentication required"
  }
}
```

**403 Forbidden:**
```json
{
  "success": false,
  "error": {
    "code": "FORBIDDEN",
    "message": "User does not have permission to perform this action"
  }
}
```

**404 Not Found:**
```json
{
  "success": false,
  "error": {
    "code": "NOT_FOUND",
    "message": "Resource not found"
  }
}
```

**500 Internal Server Error:**
```json
{
  "success": false,
  "error": {
    "code": "INTERNAL_SERVER_ERROR",
    "message": "An unexpected error occurred",
    "requestId": "req-12345"
  }
}
```

---

## Rate Limiting

All endpoints are subject to rate limiting:

- **Anonymous**: 100 requests/hour
- **Authenticated**: 1000 requests/hour
- **Admin**: Unlimited

Rate limit headers included in responses:
```
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1702400400
```

---

## Authentication

All endpoints (except `/api/health` and `/api/config`) require authentication via JWT token:

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

---

## Implementation Priority

### Phase 12a - Initial Backend Setup
1. Document CRUD endpoints (1.1-1.4)
2. Search/Filter endpoints (2)
3. User management endpoints (7)
4. Health check (8)

### Phase 12b - Workflow Integration
1. Approval endpoints (3)
2. Bulk operations (4)

### Phase 12c - Analytics & Reporting
1. Analytics endpoints (5)
2. Download/Export endpoints (6)

---
