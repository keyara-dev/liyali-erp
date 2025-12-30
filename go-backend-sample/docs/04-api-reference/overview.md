# Liyali Gateway - API Specification

**Version**: 1.0
**Base URL**: `https://api.liyali.com` (production) | `http://localhost:3001` (development)
**Protocol**: REST over HTTPS
**Data Format**: JSON
**Authentication**: JWT Bearer Token
**Last Updated**: December 23, 2025

---

## 📋 Table of Contents

1. [Authentication](#authentication)
2. [Common Headers](#common-headers)
3. [Response Format](#response-format)
4. [Error Handling](#error-handling)
5. [Pagination](#pagination)
6. [API Endpoints](#api-endpoints)
   - [Auth Endpoints](#auth-endpoints)
   - [Approval Tasks](#approval-tasks)
   - [Bulk Operations](#bulk-operations)
   - [Workflows](#workflows)
   - [Documents](#documents)
   - [Analytics](#analytics)
   - [Notifications](#notifications)
   - [Users](#users)
7. [Webhook Events](#webhook-events)
8. [Rate Limiting](#rate-limiting)
9. [API Versioning](#api-versioning)

---

## 🔐 Authentication

All API requests (except authentication endpoints) require a valid JWT bearer token.

### Authentication Flow

```
1. User initiates OAuth login → /api/auth/login
2. OAuth provider redirects back with authorization code
3. Backend exchanges code for access token
4. Backend generates JWT token
5. Client stores JWT token
6. Client includes token in all subsequent requests
```

### JWT Token Format

```json
{
  "user_id": "uuid",
  "email": "user@example.com",
  "name": "John Doe",
  "role": "DEPARTMENT_MANAGER",
  "department": "Sales",
  "exp": 1735657200,
  "iat": 1735567200
}
```

### Token Expiration
- **Access Token**: 1 hour
- **Refresh Token**: 8 hours
- **Idle Timeout**: 1 hour of inactivity

---

## 📨 Common Headers

### Request Headers

```http
Authorization: Bearer <jwt_token>
Content-Type: application/json
Accept: application/json
X-Request-ID: uuid (optional, for tracing)
```

### Response Headers

```http
Content-Type: application/json
X-Request-ID: uuid
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1735567200
```

---

## 📦 Response Format

### Success Response

```json
{
  "success": true,
  "data": {
    // Response payload
  },
  "meta": {
    "timestamp": "2025-12-23T10:30:00Z",
    "request_id": "uuid"
  }
}
```

### Paginated Response

```json
{
  "success": true,
  "data": {
    "items": [...],
    "total": 150,
    "page": 1,
    "limit": 20,
    "page_count": 8
  },
  "meta": {
    "timestamp": "2025-12-23T10:30:00Z",
    "request_id": "uuid"
  }
}
```

---

## ⚠️ Error Handling

### Error Response Format

```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Missing required fields",
    "details": [
      {
        "field": "signature",
        "message": "Signature is required"
      }
    ]
  },
  "meta": {
    "timestamp": "2025-12-23T10:30:00Z",
    "request_id": "uuid"
  }
}
```

### HTTP Status Codes

| Code | Meaning | Description |
|------|---------|-------------|
| 200 | OK | Request successful |
| 201 | Created | Resource created successfully |
| 204 | No Content | Successful deletion |
| 400 | Bad Request | Invalid request parameters |
| 401 | Unauthorized | Missing or invalid authentication |
| 403 | Forbidden | Insufficient permissions |
| 404 | Not Found | Resource not found |
| 409 | Conflict | Resource conflict (e.g., duplicate) |
| 422 | Unprocessable Entity | Validation failed |
| 429 | Too Many Requests | Rate limit exceeded |
| 500 | Internal Server Error | Server error |
| 503 | Service Unavailable | Service temporarily unavailable |

### Error Codes

| Code | Description |
|------|-------------|
| `VALIDATION_ERROR` | Request validation failed |
| `AUTHENTICATION_ERROR` | Authentication failed |
| `AUTHORIZATION_ERROR` | Insufficient permissions |
| `NOT_FOUND` | Resource not found |
| `CONFLICT` | Resource conflict |
| `RATE_LIMIT_EXCEEDED` | Too many requests |
| `INTERNAL_ERROR` | Internal server error |

---

## 📄 Pagination

### Query Parameters

```
?page=1          # Page number (default: 1)
?limit=20        # Items per page (default: 20, max: 100)
?sort=created_at # Sort field
?order=desc      # Sort order (asc/desc)
```

### Example Request

```http
GET /api/approvals/tasks?page=2&limit=50&sort=created_at&order=desc
```

### Example Response

```json
{
  "success": true,
  "data": {
    "items": [...],
    "total": 150,
    "page": 2,
    "limit": 50,
    "page_count": 3
  }
}
```

---

## 🔌 API Endpoints

---

## Auth Endpoints

### POST /api/auth/login

Initiate OAuth login flow.

**Request**:
```http
POST /api/auth/login
Content-Type: application/json

{
  "provider": "entra_id",  // "entra_id" | "google" | "github"
  "redirect_uri": "http://localhost:3000/auth/callback"
}
```

**Response** (200 OK):
```json
{
  "success": true,
  "data": {
    "auth_url": "https://login.microsoftonline.com/oauth2/authorize?..."
  }
}
```

---

### POST /api/auth/callback

Handle OAuth callback and generate JWT token.

**Request**:
```http
POST /api/auth/callback
Content-Type: application/json

{
  "provider": "entra_id",
  "code": "authorization_code",
  "state": "random_state"
}
```

**Response** (200 OK):
```json
{
  "success": true,
  "data": {
    "access_token": "jwt_token",
    "refresh_token": "refresh_token",
    "expires_in": 3600,
    "user": {
      "id": "uuid",
      "email": "user@example.com",
      "name": "John Doe",
      "role": "DEPARTMENT_MANAGER",
      "department": "Sales"
    }
  }
}
```

---

### POST /api/auth/refresh

Refresh access token using refresh token.

**Request**:
```http
POST /api/auth/refresh
Content-Type: application/json

{
  "refresh_token": "refresh_token"
}
```

**Response** (200 OK):
```json
{
  "success": true,
  "data": {
    "access_token": "new_jwt_token",
    "expires_in": 3600
  }
}
```

---

### POST /api/auth/logout

Invalidate current session.

**Request**:
```http
POST /api/auth/logout
Authorization: Bearer <jwt_token>
```

**Response** (200 OK):
```json
{
  "success": true,
  "data": {
    "message": "Logged out successfully"
  }
}
```

---

## Approval Tasks

### GET /api/approvals/tasks

Fetch approval tasks for authenticated user.

**Query Parameters**:
- `status` (optional): `pending` | `approved` | `rejected`
- `page` (optional): Page number (default: 1)
- `limit` (optional): Items per page (default: 20, max: 100)
- `sort` (optional): Sort field (default: `created_at`)
- `order` (optional): `asc` | `desc` (default: `desc`)

**Request**:
```http
GET /api/approvals/tasks?status=pending&page=1&limit=20
Authorization: Bearer <jwt_token>
```

**Response** (200 OK):
```json
{
  "success": true,
  "data": {
    "tasks": [
      {
        "id": "uuid",
        "entity_id": "uuid",
        "entity_type": "REQUISITION",
        "entity_number": "REQ-2024-001",
        "status": "pending",
        "stage_name": "Manager Review",
        "stage_index": 1,
        "importance": "HIGH",
        "approver_user": {
          "id": "uuid",
          "name": "John Manager",
          "email": "manager@example.com",
          "role": "DEPARTMENT_MANAGER"
        },
        "created_at": "2025-12-20T10:00:00Z",
        "due_date": "2025-12-25T17:00:00Z",
        "workflow_id": "uuid",
        "workflow_name": "3-Stage Requisition",
        "document": {
          "description": "Office supplies",
          "amount": 2500.00,
          "department_id": "dept-001"
        }
      }
    ],
    "total": 15,
    "page": 1,
    "limit": 20,
    "page_count": 1
  }
}
```

---

### GET /api/approvals/tasks/:id

Get single approval task details with full history.

**Request**:
```http
GET /api/approvals/tasks/uuid
Authorization: Bearer <jwt_token>
```

**Response** (200 OK):
```json
{
  "success": true,
  "data": {
    "task": {
      "id": "uuid",
      "entity_id": "uuid",
      "entity_type": "REQUISITION",
      "entity_number": "REQ-2024-001",
      "status": "pending",
      "stage_name": "Manager Review",
      "stage_index": 1,
      "importance": "HIGH",
      "approver_user": {
        "id": "uuid",
        "name": "John Manager",
        "email": "manager@example.com"
      },
      "created_at": "2025-12-20T10:00:00Z",
      "due_date": "2025-12-25T17:00:00Z",
      "workflow": {
        "id": "uuid",
        "name": "3-Stage Requisition",
        "stages": [
          {
            "order": 1,
            "name": "Manager Review",
            "approver_roles": ["DEPARTMENT_MANAGER"]
          },
          {
            "order": 2,
            "name": "Finance Review",
            "approver_roles": ["FINANCE_OFFICER"]
          }
        ]
      },
      "history": [
        {
          "id": "uuid",
          "action": "submitted",
          "approver_user": {
            "name": "Jane Requester"
          },
          "timestamp": "2025-12-20T10:00:00Z",
          "remarks": "Urgent request"
        }
      ],
      "document": {
        "description": "Office supplies",
        "amount": 2500.00
      }
    }
  }
}
```

---

### POST /api/approvals/tasks/:id/approve

Approve an approval task.

**Request**:
```http
POST /api/approvals/tasks/uuid/approve
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
  "assignment_id": "uuid",
  "stage_number": 1,
  "signature": "base64_encoded_signature_image",
  "comments": "Approved for procurement"
}
```

**Response** (200 OK):
```json
{
  "success": true,
  "data": {
    "task_id": "uuid",
    "action": "approved",
    "new_status": "approved",
    "next_stage": "Finance Review",
    "timestamp": "2025-12-23T10:30:00Z"
  }
}
```

**Errors**:
- `400`: Missing required fields
- `403`: User is not the assigned approver
- `404`: Task not found
- `409`: Task already processed

---

### POST /api/approvals/tasks/:id/reject

Reject an approval task.

**Request**:
```http
POST /api/approvals/tasks/uuid/reject
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
  "signature": "base64_encoded_signature_image",
  "remarks": "Cost exceeds budget allocation"
}
```

**Response** (200 OK):
```json
{
  "success": true,
  "data": {
    "task_id": "uuid",
    "action": "rejected",
    "new_status": "rejected",
    "reason": "Cost exceeds budget allocation",
    "timestamp": "2025-12-23T10:30:00Z"
  }
}
```

**Errors**:
- `400`: Missing signature or remarks
- `403`: User is not the assigned approver
- `404`: Task not found

---

### POST /api/approvals/tasks/:id/reassign

Reassign approval task to another user.

**Request**:
```http
POST /api/approvals/tasks/uuid/reassign
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
  "new_approver_id": "uuid",
  "new_approver_name": "Jane Manager",
  "reason": "Original approver on leave"
}
```

**Response** (200 OK):
```json
{
  "success": true,
  "data": {
    "task_id": "uuid",
    "action": "reassigned",
    "new_approver": "Jane Manager",
    "timestamp": "2025-12-23T10:30:00Z"
  }
}
```

---

## Bulk Operations

### POST /api/approvals/bulk/approve

Approve multiple tasks at once.

**Request**:
```http
POST /api/approvals/bulk/approve
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
  "task_ids": ["uuid1", "uuid2", "uuid3"],
  "remarks": "Batch approval for routine requests"
}
```

**Response** (200 OK):
```json
{
  "success": true,
  "data": {
    "approved": 3,
    "failed": 0,
    "message": "Successfully approved 3 tasks",
    "timestamp": "2025-12-23T10:30:00Z",
    "results": [
      {
        "task_id": "uuid1",
        "status": "success"
      },
      {
        "task_id": "uuid2",
        "status": "success"
      },
      {
        "task_id": "uuid3",
        "status": "success"
      }
    ]
  }
}
```

---

### POST /api/approvals/bulk/reject

Reject multiple tasks at once.

**Request**:
```http
POST /api/approvals/bulk/reject
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
  "task_ids": ["uuid1", "uuid2"],
  "remarks": "Insufficient budget justification"
}
```

**Response** (200 OK):
```json
{
  "success": true,
  "data": {
    "rejected": 2,
    "failed": 0,
    "message": "Successfully rejected 2 tasks",
    "timestamp": "2025-12-23T10:30:00Z"
  }
}
```

**Errors**:
- `400`: Missing remarks (required for rejection)

---

### POST /api/approvals/bulk/reassign

Reassign multiple tasks at once.

**Request**:
```http
POST /api/approvals/bulk/reassign
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
  "task_ids": ["uuid1", "uuid2"],
  "new_approver_id": "uuid",
  "new_approver_name": "Jane Manager",
  "reason": "Workload balancing"
}
```

**Response** (200 OK):
```json
{
  "success": true,
  "data": {
    "reassigned": 2,
    "failed": 0,
    "new_approver": "Jane Manager",
    "timestamp": "2025-12-23T10:30:00Z"
  }
}
```

---

## Workflows

### GET /api/workflows

List all workflows.

**Query Parameters**:
- `entity_type` (optional): Filter by document type
- `status` (optional): `published` | `draft`

**Request**:
```http
GET /api/workflows?entity_type=REQUISITION&status=published
Authorization: Bearer <jwt_token>
```

**Response** (200 OK):
```json
{
  "success": true,
  "data": {
    "workflows": [
      {
        "id": "uuid",
        "name": "3-Stage Requisition",
        "description": "Standard requisition approval workflow",
        "entity_type": "REQUISITION",
        "status": "published",
        "stages": [
          {
            "order": 1,
            "name": "Manager Review",
            "approver_roles": ["DEPARTMENT_MANAGER"],
            "allow_reassign": true
          },
          {
            "order": 2,
            "name": "Finance Review",
            "approver_roles": ["FINANCE_OFFICER"],
            "allow_reassign": true
          },
          {
            "order": 3,
            "name": "Director Approval",
            "approver_roles": ["DIRECTOR", "CFO"],
            "allow_reassign": false
          }
        ],
        "created_at": "2025-01-01T00:00:00Z",
        "updated_at": "2025-01-10T00:00:00Z",
        "created_by": "admin_uuid"
      }
    ],
    "total": 5
  }
}
```

---

### GET /api/workflows/:id

Get single workflow details.

**Request**:
```http
GET /api/workflows/uuid
Authorization: Bearer <jwt_token>
```

**Response** (200 OK):
```json
{
  "success": true,
  "data": {
    "workflow": {
      "id": "uuid",
      "name": "3-Stage Requisition",
      "description": "Standard requisition approval workflow",
      "entity_type": "REQUISITION",
      "status": "published",
      "stages": [...],
      "created_at": "2025-01-01T00:00:00Z",
      "updated_at": "2025-01-10T00:00:00Z"
    }
  }
}
```

---

### POST /api/workflows

Create a new workflow.

**Request**:
```http
POST /api/workflows
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
  "name": "Custom PO Workflow",
  "description": "Custom purchase order approval",
  "entity_type": "PURCHASE_ORDER",
  "status": "draft",
  "stages": [
    {
      "order": 1,
      "name": "Procurement Review",
      "approver_roles": ["PROCUREMENT_OFFICER"],
      "allow_reassign": true
    },
    {
      "order": 2,
      "name": "Budget Verification",
      "approver_roles": ["FINANCE_OFFICER"],
      "allow_reassign": true
    }
  ]
}
```

**Response** (201 Created):
```json
{
  "success": true,
  "data": {
    "workflow": {
      "id": "new_uuid",
      "name": "Custom PO Workflow",
      // ... full workflow object
    }
  }
}
```

**Errors**:
- `400`: Validation error
- `403`: Insufficient permissions (ADMIN only)

---

### PUT /api/workflows/:id

Update an existing workflow.

**Request**:
```http
PUT /api/workflows/uuid
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
  "name": "Updated Workflow Name",
  "description": "Updated description",
  "status": "published"
}
```

**Response** (200 OK):
```json
{
  "success": true,
  "data": {
    "workflow": {
      // ... updated workflow object
    }
  }
}
```

---

### DELETE /api/workflows/:id

Delete a workflow.

**Request**:
```http
DELETE /api/workflows/uuid
Authorization: Bearer <jwt_token>
```

**Response** (204 No Content)

**Errors**:
- `403`: Insufficient permissions (ADMIN only)
- `404`: Workflow not found
- `409`: Cannot delete workflow in use

---

## Documents

Document endpoints follow the same pattern for each document type:
- `/api/requisitions`
- `/api/budgets`
- `/api/purchase-orders`
- `/api/payment-vouchers`
- `/api/grn`

### GET /api/requisitions

List requisitions.

**Query Parameters**:
- `status` (optional): Filter by status
- `page`, `limit`, `sort`, `order`: Pagination

**Request**:
```http
GET /api/requisitions?status=APPROVED&page=1&limit=20
Authorization: Bearer <jwt_token>
```

**Response** (200 OK):
```json
{
  "success": true,
  "data": {
    "requisitions": [
      {
        "id": "uuid",
        "requisition_number": "REQ-2024-001",
        "status": "APPROVED",
        "description": "Office supplies",
        "total_amount": 2500.00,
        "currency": "ZMW",
        "department_id": "dept-001",
        "requester": {
          "id": "uuid",
          "name": "John Requester"
        },
        "created_at": "2025-12-20T10:00:00Z",
        "approved_at": "2025-12-22T15:30:00Z",
        "items": [
          {
            "description": "Printer Paper A4",
            "quantity": 10,
            "unit_price": 50.00,
            "total": 500.00
          }
        ]
      }
    ],
    "total": 45,
    "page": 1,
    "limit": 20,
    "page_count": 3
  }
}
```

---

### GET /api/requisitions/:id

Get single requisition details.

**Request**:
```http
GET /api/requisitions/uuid
Authorization: Bearer <jwt_token>
```

**Response** (200 OK):
```json
{
  "success": true,
  "data": {
    "requisition": {
      "id": "uuid",
      "requisition_number": "REQ-2024-001",
      "status": "APPROVED",
      "description": "Office supplies",
      "total_amount": 2500.00,
      "currency": "ZMW",
      "items": [...],
      "approval_chain": [...],
      "action_history": [...],
      "created_at": "2025-12-20T10:00:00Z",
      "approved_at": "2025-12-22T15:30:00Z"
    }
  }
}
```

---

### POST /api/requisitions

Create a new requisition.

**Request**:
```http
POST /api/requisitions
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
  "description": "Office supplies for Q1 2025",
  "department_id": "dept-001",
  "justification": "Regular quarterly supplies",
  "items": [
    {
      "description": "Printer Paper A4",
      "quantity": 10,
      "unit_price": 50.00,
      "total": 500.00
    },
    {
      "description": "Pens (box of 50)",
      "quantity": 5,
      "unit_price": 100.00,
      "total": 500.00
    }
  ],
  "total_amount": 1000.00,
  "currency": "ZMW"
}
```

**Response** (201 Created):
```json
{
  "success": true,
  "data": {
    "requisition": {
      "id": "new_uuid",
      "requisition_number": "REQ-2024-042",
      "status": "DRAFT",
      // ... full requisition object
    }
  }
}
```

---

### PUT /api/requisitions/:id

Update a requisition (only if status is DRAFT).

**Request**:
```http
PUT /api/requisitions/uuid
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
  "description": "Updated description",
  "items": [...]
}
```

**Response** (200 OK):
```json
{
  "success": true,
  "data": {
    "requisition": {
      // ... updated requisition
    }
  }
}
```

**Errors**:
- `403`: Cannot update requisition not in DRAFT status

---

### POST /api/requisitions/:id/submit

Submit requisition for approval.

**Request**:
```http
POST /api/requisitions/uuid/submit
Authorization: Bearer <jwt_token>
```

**Response** (200 OK):
```json
{
  "success": true,
  "data": {
    "requisition": {
      "id": "uuid",
      "status": "SUBMITTED",
      "current_approval_stage": 1,
      "next_approver": {
        "id": "uuid",
        "name": "John Manager"
      }
    }
  }
}
```

---

## Analytics

### GET /api/analytics/metrics

Get dashboard metrics.

**Request**:
```http
GET /api/analytics/metrics
Authorization: Bearer <jwt_token>
```

**Response** (200 OK):
```json
{
  "success": true,
  "data": {
    "total_pending": 15,
    "total_approved": 142,
    "total_rejected": 8,
    "avg_approval_time": "2.5 days",
    "sla_compliance": 85,
    "by_type": {
      "REQUISITION": {
        "pending": 5,
        "approved": 50,
        "rejected": 2
      },
      "PURCHASE_ORDER": {
        "pending": 3,
        "approved": 42,
        "rejected": 3
      }
    }
  }
}
```

---

### GET /api/analytics/trends

Get 7-day approval trends.

**Request**:
```http
GET /api/analytics/trends?days=7
Authorization: Bearer <jwt_token>
```

**Response** (200 OK):
```json
{
  "success": true,
  "data": {
    "trends": [
      {
        "date": "2025-12-17",
        "approved": 8,
        "rejected": 1,
        "pending": 3
      },
      {
        "date": "2025-12-18",
        "approved": 12,
        "rejected": 0,
        "pending": 5
      }
      // ... 7 days
    ]
  }
}
```

---

### GET /api/analytics/bottleneck

Get bottleneck analysis.

**Request**:
```http
GET /api/analytics/bottleneck
Authorization: Bearer <jwt_token>
```

**Response** (200 OK):
```json
{
  "success": true,
  "data": {
    "bottlenecks": [
      {
        "stage_name": "Finance Review",
        "avg_time_days": 3.2,
        "pending_count": 12,
        "severity": "HIGH",
        "recommendation": "Consider adding additional finance officers"
      },
      {
        "stage_name": "Manager Review",
        "avg_time_days": 1.8,
        "pending_count": 5,
        "severity": "MEDIUM",
        "recommendation": "Within acceptable limits"
      }
    ]
  }
}
```

---

## Notifications

### GET /api/notifications

Get user notifications.

**Query Parameters**:
- `is_read` (optional): `true` | `false`
- `type` (optional): Filter by notification type

**Request**:
```http
GET /api/notifications?is_read=false
Authorization: Bearer <jwt_token>
```

**Response** (200 OK):
```json
{
  "success": true,
  "data": {
    "notifications": [
      {
        "id": "uuid",
        "type": "task_assigned",
        "title": "New Task Assigned",
        "message": "You have been assigned approval task REQ-2024-001",
        "task_id": "uuid",
        "is_read": false,
        "created_at": "2025-12-23T10:00:00Z"
      }
    ],
    "total": 5,
    "unread_count": 3
  }
}
```

---

### POST /api/notifications/:id/read

Mark notification as read.

**Request**:
```http
POST /api/notifications/uuid/read
Authorization: Bearer <jwt_token>
```

**Response** (200 OK):
```json
{
  "success": true,
  "data": {
    "notification": {
      "id": "uuid",
      "is_read": true,
      "read_at": "2025-12-23T10:30:00Z"
    }
  }
}
```

---

### DELETE /api/notifications/:id

Delete a notification.

**Request**:
```http
DELETE /api/notifications/uuid
Authorization: Bearer <jwt_token>
```

**Response** (204 No Content)

---

## Users

### GET /api/users

List users (ADMIN only).

**Query Parameters**:
- `role` (optional): Filter by role
- `department` (optional): Filter by department
- `is_active` (optional): Filter by active status

**Request**:
```http
GET /api/users?role=DEPARTMENT_MANAGER
Authorization: Bearer <jwt_token>
```

**Response** (200 OK):
```json
{
  "success": true,
  "data": {
    "users": [
      {
        "id": "uuid",
        "email": "manager@example.com",
        "name": "John Manager",
        "role": "DEPARTMENT_MANAGER",
        "department": "Sales",
        "is_active": true,
        "last_login": "2025-12-23T09:00:00Z",
        "created_at": "2025-01-01T00:00:00Z"
      }
    ],
    "total": 25
  }
}
```

---

### GET /api/users/me

Get current user profile.

**Request**:
```http
GET /api/users/me
Authorization: Bearer <jwt_token>
```

**Response** (200 OK):
```json
{
  "success": true,
  "data": {
    "user": {
      "id": "uuid",
      "email": "user@example.com",
      "name": "John Doe",
      "role": "DEPARTMENT_MANAGER",
      "department": "Sales",
      "permissions": [
        "approve_stage_1",
        "reject_stage_1",
        "reassign_stage_1"
      ]
    }
  }
}
```

---

## 🪝 Webhook Events

### Available Events

| Event | Description |
|-------|-------------|
| `task.assigned` | Task assigned to user |
| `task.approved` | Task approved |
| `task.rejected` | Task rejected |
| `task.reassigned` | Task reassigned |
| `document.created` | New document created |
| `document.submitted` | Document submitted for approval |
| `workflow.completed` | Workflow completed |

### Webhook Payload Example

```json
{
  "event": "task.approved",
  "timestamp": "2025-12-23T10:30:00Z",
  "data": {
    "task_id": "uuid",
    "entity_type": "REQUISITION",
    "entity_number": "REQ-2024-001",
    "approver": {
      "id": "uuid",
      "name": "John Manager"
    },
    "stage_number": 1,
    "stage_name": "Manager Review"
  }
}
```

---

## 🚦 Rate Limiting

### Limits

| Tier | Requests/Hour | Requests/Day |
|------|---------------|--------------|
| Authenticated | 1,000 | 10,000 |
| Admin | 5,000 | 50,000 |

### Rate Limit Headers

```http
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1735567200
```

### Rate Limit Exceeded Response

```json
{
  "success": false,
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "Rate limit exceeded. Try again in 30 minutes.",
    "retry_after": 1800
  }
}
```

---

## 🔢 API Versioning

### Current Version: v1

All endpoints are prefixed with `/api/` (current version).

Future versions will use:
- `/api/v2/...`
- `/api/v3/...`

Version 1 will be supported for at least 12 months after v2 release.

---

## 📚 Additional Resources

- [Backend Development Planner](./development-planner.md)
- [Go Backend Implementation Guide](../../docs/BACKEND-GUIDE-GO.md)
- [Frontend Documentation](../../docs/README.md)

---

## 🔐 Security Best Practices

1. **Always use HTTPS** in production
2. **Never log sensitive data** (passwords, tokens, signatures)
3. **Validate all inputs** on server side
4. **Use parameterized queries** to prevent SQL injection
5. **Implement CSRF protection** for state-changing operations
6. **Rate limit all endpoints** to prevent abuse
7. **Audit all sensitive actions** (approvals, deletions, changes)

---

**Version**: 1.0
**Status**: Ready for Implementation
**Last Updated**: December 23, 2025
