# Liyali Gateway API Documentation

Base URL: `http://localhost:3001/api`

## Table of Contents
- [Authentication](#authentication)
- [Approval Tasks](#approval-tasks)
- [Workflows](#workflows)
- [Documents](#documents)
- [Analytics](#analytics)
- [Notifications](#notifications)
- [Audit Logs](#audit-logs)

---

## Authentication

All protected endpoints require a Bearer token in the Authorization header:
```
Authorization: Bearer <access_token>
```

### Register
**POST** `/auth/register`

Create a new user account.

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "SecurePass123!",
  "name": "John Doe",
  "role": "REQUESTER",
  "department": "Engineering"
}
```

**Response:** `201 Created`
```json
{
  "message": "User registered successfully",
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "name": "John Doe",
    "role": "REQUESTER"
  }
}
```

### Login
**POST** `/auth/login`

Authenticate and receive access tokens.

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "SecurePass123!"
}
```

**Response:** `200 OK`
```json
{
  "access_token": "eyJhbGc...",
  "refresh_token": "uuid",
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "name": "John Doe",
    "role": "REQUESTER"
  }
}
```

### Refresh Token
**POST** `/auth/refresh`

Get a new access token using a refresh token.

**Request Body:**
```json
{
  "refresh_token": "uuid"
}
```

**Response:** `200 OK`
```json
{
  "access_token": "eyJhbGc..."
}
```

### Logout
**POST** `/auth/logout`

Invalidate a refresh token.

**Request Body:**
```json
{
  "refresh_token": "uuid"
}
```

**Response:** `200 OK`
```json
{
  "message": "Logged out successfully"
}
```

### Get Current User
**GET** `/auth/me` 🔒

Get the authenticated user's information.

**Response:** `200 OK`
```json
{
  "id": "uuid",
  "email": "user@example.com",
  "name": "John Doe",
  "role": "REQUESTER",
  "department": "Engineering"
}
```

### Change Password
**POST** `/auth/change-password` 🔒

Change the authenticated user's password.

**Request Body:**
```json
{
  "current_password": "OldPass123!",
  "new_password": "NewPass123!"
}
```

**Response:** `200 OK`
```json
{
  "message": "Password changed successfully"
}
```

---

## Approval Tasks

### List Tasks
**GET** `/approvals/tasks?status=PENDING&limit=20&offset=0` 🔒

Get approval tasks assigned to the current user.

**Query Parameters:**
- `status` (optional): Filter by status (PENDING, APPROVED, REJECTED)
- `limit` (optional, default: 20, max: 100): Number of results
- `offset` (optional, default: 0): Pagination offset

**Response:** `200 OK`
```json
{
  "tasks": [
    {
      "id": "uuid",
      "document_id": "uuid",
      "status": "PENDING",
      "current_stage": 1,
      "total_stages": 3,
      "priority": "HIGH",
      "due_date": "2025-01-15T10:00:00Z",
      "notes": "Please review urgently"
    }
  ],
  "total": 15,
  "limit": 20,
  "offset": 0
}
```

### List Overdue Tasks
**GET** `/approvals/tasks/overdue?limit=20&offset=0` 🔒

Get overdue approval tasks.

**Response:** `200 OK` (same format as List Tasks)

### Get Task by ID
**GET** `/approvals/tasks/:id` 🔒

Get a specific approval task with its history.

**Response:** `200 OK`
```json
{
  "task": {
    "id": "uuid",
    "document_id": "uuid",
    "status": "PENDING",
    "current_stage": 1,
    "total_stages": 3
  },
  "history": [
    {
      "id": "uuid",
      "task_id": "uuid",
      "user_id": "uuid",
      "action": "COMMENTED",
      "comment": "Looks good",
      "created_at": "2025-01-10T14:30:00Z"
    }
  ]
}
```

### Approve Task
**POST** `/approvals/tasks/:id/approve` 🔒

Approve an approval task.

**Request Body:**
```json
{
  "signature": "base64_encoded_signature",
  "comment": "Approved with conditions"
}
```

**Response:** `200 OK`
```json
{
  "task": {
    "id": "uuid",
    "status": "APPROVED",
    "current_stage": 2
  },
  "message": "Task approved successfully"
}
```

### Reject Task
**POST** `/approvals/tasks/:id/reject` 🔒

Reject an approval task.

**Request Body:**
```json
{
  "reason": "Budget exceeded",
  "comment": "Please revise and resubmit"
}
```

**Response:** `200 OK`

### Reassign Task
**POST** `/approvals/tasks/:id/reassign` 🔒

Reassign a task to another user.

**Request Body:**
```json
{
  "new_assignee_id": "uuid",
  "reason": "Transferring to specialist"
}
```

**Response:** `200 OK`

### Add Comment
**POST** `/approvals/tasks/:id/comment` 🔒

Add a comment to an approval task.

**Request Body:**
```json
{
  "comment": "Please provide additional documentation"
}
```

**Response:** `200 OK`

### Bulk Approve
**POST** `/approvals/bulk/approve` 🔒

Approve multiple tasks at once.

**Request Body:**
```json
{
  "task_ids": ["uuid1", "uuid2", "uuid3"],
  "signature": "base64_encoded_signature",
  "comment": "Batch approval"
}
```

**Response:** `200 OK`
```json
{
  "approved": 3,
  "failed": 0,
  "results": [...]
}
```

### Bulk Reject
**POST** `/approvals/bulk/reject` 🔒

Reject multiple tasks at once.

**Request Body:**
```json
{
  "task_ids": ["uuid1", "uuid2"],
  "reason": "Insufficient information",
  "comment": "Please provide more details"
}
```

**Response:** `200 OK`

### Bulk Reassign
**POST** `/approvals/bulk/reassign` 🔒

Reassign multiple tasks at once.

**Request Body:**
```json
{
  "task_ids": ["uuid1", "uuid2"],
  "new_assignee_id": "uuid",
  "reason": "Workload distribution"
}
```

**Response:** `200 OK`

---

## Workflows

### List Workflows
**GET** `/workflows?document_type=REQUISITION&limit=20&offset=0` 🔒

Get all workflows.

**Query Parameters:**
- `document_type` (optional): Filter by document type
- `is_active` (optional): Filter by active status (true/false)
- `limit` (optional, default: 20, max: 100)
- `offset` (optional, default: 0)

**Response:** `200 OK`
```json
{
  "workflows": [
    {
      "id": "uuid",
      "name": "Standard Requisition Workflow",
      "document_type": "REQUISITION",
      "stages": [...],
      "is_active": true,
      "created_at": "2025-01-01T00:00:00Z"
    }
  ],
  "total": 5,
  "limit": 20,
  "offset": 0
}
```

### Get Workflow
**GET** `/workflows/:id` 🔒

Get a specific workflow.

**Response:** `200 OK`

### Get Default Workflow
**GET** `/workflows/default/:documentType` 🔒

Get the default workflow for a document type.

**Response:** `200 OK`

### Create Workflow
**POST** `/workflows` 🔒

Create a new workflow.

**Request Body:**
```json
{
  "name": "Custom Approval Workflow",
  "description": "Three-stage approval process",
  "document_type": "BUDGET",
  "stages": [
    {
      "stage": 1,
      "name": "Manager Review",
      "approvers": ["MANAGER"],
      "required": 1,
      "sla_hours": 24
    }
  ],
  "is_active": true
}
```

**Response:** `201 Created`

### Update Workflow
**PUT** `/workflows/:id` 🔒

Update an existing workflow.

**Response:** `200 OK`

### Activate Workflow
**POST** `/workflows/:id/activate` 🔒

Activate a workflow.

**Response:** `200 OK`

### Deactivate Workflow
**POST** `/workflows/:id/deactivate` 🔒

Deactivate a workflow.

**Response:** `200 OK`

### Delete Workflow
**DELETE** `/workflows/:id` 🔒

Delete a workflow.

**Response:** `200 OK`

---

## Documents

### List Documents
**GET** `/documents?status=DRAFT&document_type=REQUISITION&limit=20&offset=0` 🔒

Get all documents (admins/managers see all, users see their own).

**Query Parameters:**
- `status` (optional): Filter by status
- `document_type` (optional): Filter by type
- `limit` (optional, default: 20, max: 100)
- `offset` (optional, default: 0)

**Response:** `200 OK`
```json
{
  "documents": [
    {
      "id": "uuid",
      "document_type": "REQUISITION",
      "document_number": "REQ-2025-001",
      "title": "Office Supplies",
      "amount": 500.00,
      "currency": "USD",
      "status": "DRAFT",
      "created_by": "uuid",
      "created_at": "2025-01-10T10:00:00Z"
    }
  ],
  "total": 25,
  "limit": 20,
  "offset": 0
}
```

### Get My Documents
**GET** `/documents/my?limit=20&offset=0` 🔒

Get documents created by the current user.

**Response:** `200 OK` (same format as List Documents)

### Get Document
**GET** `/documents/:id` 🔒

Get a specific document.

**Response:** `200 OK`

### Get Document by Number
**GET** `/documents/number/:number` 🔒

Get a document by its document number.

**Response:** `200 OK`

### Create Document
**POST** `/documents` 🔒

Create a new document.

**Request Body:**
```json
{
  "document_type": "REQUISITION",
  "title": "Office Supplies - Q1",
  "description": "Quarterly office supplies purchase",
  "amount": 1500.00,
  "currency": "USD",
  "department": "Operations",
  "workflow_id": "uuid",
  "data": {
    "items": [
      {"name": "Paper", "quantity": 100, "unit_price": 10.00}
    ]
  },
  "metadata": {}
}
```

**Response:** `201 Created`

### Update Document
**PUT** `/documents/:id` 🔒

Update a document (only in DRAFT status).

**Response:** `200 OK`

### Submit Document
**POST** `/documents/:id/submit` 🔒

Submit a document for approval.

**Response:** `200 OK`

### Delete Document
**DELETE** `/documents/:id` 🔒

Delete a document (only in DRAFT status).

**Response:** `200 OK`

---

## Analytics

### Dashboard Metrics
**GET** `/analytics/metrics` 🔒

Get comprehensive dashboard metrics.

**Response:** `200 OK`
```json
{
  "total_documents": 150,
  "documents_by_status": {
    "DRAFT": 25,
    "SUBMITTED": 40,
    "APPROVED": 75,
    "REJECTED": 10
  },
  "pending_approvals": 15,
  "overdue_approvals": 3,
  "approvals_by_status": {
    "PENDING": 15,
    "APPROVED": 80,
    "REJECTED": 12
  },
  "active_workflows": 5,
  "average_approval_time_hours": 18.5,
  "documents_by_type": {
    "REQUISITION": 80,
    "BUDGET": 30,
    "PURCHASE_ORDER": 40
  }
}
```

### Trend Data
**GET** `/analytics/trends?days=7` 🔒

Get trend data for the specified number of days.

**Query Parameters:**
- `days` (optional, default: 7, max: 90): Number of days

**Response:** `200 OK`
```json
{
  "days": 7,
  "trends": [
    {
      "date": "2025-01-20",
      "documents_created": 5,
      "documents_approved": 8,
      "documents_rejected": 1,
      "approvals_completed": 12
    }
  ]
}
```

### Bottleneck Analysis
**GET** `/analytics/bottlenecks` 🔒

Identify workflow bottlenecks.

**Response:** `200 OK`
```json
{
  "bottlenecks": [
    {
      "document_id": "uuid",
      "document_type": "BUDGET",
      "stage": 2,
      "pending_tasks": 8,
      "average_time_in_stage_hours": 72.5
    }
  ],
  "count": 1
}
```

---

## Notifications

### List Notifications
**GET** `/notifications?limit=20&offset=0` 🔒

Get all notifications for the current user.

**Query Parameters:**
- `limit` (optional, default: 20, max: 100)
- `offset` (optional, default: 0)

**Response:** `200 OK`
```json
{
  "notifications": [
    {
      "id": "uuid",
      "type": "TASK_ASSIGNED",
      "title": "New Approval Task",
      "message": "You have been assigned a new approval task",
      "related_id": "uuid",
      "is_read": false,
      "created_at": "2025-01-20T10:00:00Z"
    }
  ],
  "total": 5,
  "unread_count": 3,
  "limit": 20,
  "offset": 0
}
```

### List Unread Notifications
**GET** `/notifications/unread?limit=20&offset=0` 🔒

Get unread notifications.

**Response:** `200 OK` (same format as List Notifications)

### Get Unread Count
**GET** `/notifications/unread/count` 🔒

Get the count of unread notifications.

**Response:** `200 OK`
```json
{
  "unread_count": 5
}
```

### Get Notification
**GET** `/notifications/:id` 🔒

Get a specific notification.

**Response:** `200 OK`

### Mark as Read
**POST** `/notifications/:id/read` 🔒

Mark a notification as read.

**Response:** `200 OK`

### Mark All as Read
**POST** `/notifications/read-all` 🔒

Mark all notifications as read.

**Response:** `200 OK`
```json
{
  "message": "all notifications marked as read"
}
```

### Delete Notification
**DELETE** `/notifications/:id` 🔒

Delete a notification.

**Response:** `200 OK`
```json
{
  "message": "notification deleted successfully"
}
```

---

## Audit Logs

**Note:** Most audit log endpoints require ADMIN or MANAGER role.

### List All Audit Logs
**GET** `/audit-logs?limit=50&offset=0&resource_type=DOCUMENT&action=CREATE&user_id=uuid` 🔒 (Admin/Manager only)

Get all audit logs with optional filtering.

**Query Parameters:**
- `limit` (optional, default: 50, max: 200)
- `offset` (optional, default: 0)
- `resource_type` (optional): Filter by resource type
- `action` (optional): Filter by action
- `user_id` (optional): Filter by user

**Response:** `200 OK`
```json
{
  "audit_logs": [
    {
      "id": "uuid",
      "user_id": "uuid",
      "action": "CREATE",
      "resource_type": "DOCUMENT",
      "resource_id": "uuid",
      "changes": {...},
      "ip_address": "127.0.0.1",
      "user_agent": "Mozilla/5.0...",
      "created_at": "2025-01-20T10:00:00Z"
    }
  ],
  "total": 150,
  "limit": 50,
  "offset": 0
}
```

### Get My Audit Logs
**GET** `/audit-logs/my?limit=50&offset=0` 🔒

Get audit logs for the current user's actions.

**Response:** `200 OK` (same format as List All Audit Logs)

### Get Audit Log
**GET** `/audit-logs/:id` 🔒 (Admin/Manager only)

Get a specific audit log entry.

**Response:** `200 OK`

### Get Audit Logs by Resource
**GET** `/audit-logs/resource/:resource_type/:resource_id?limit=50&offset=0` 🔒 (Admin/Manager only)

Get all audit logs for a specific resource.

**Response:** `200 OK`
```json
{
  "audit_logs": [...],
  "resource_type": "DOCUMENT",
  "resource_id": "uuid",
  "total": 25,
  "limit": 50,
  "offset": 0
}
```

---

## Status Codes

- `200 OK`: Request succeeded
- `201 Created`: Resource created successfully
- `400 Bad Request`: Invalid request data
- `401 Unauthorized`: Missing or invalid authentication
- `403 Forbidden`: Insufficient permissions
- `404 Not Found`: Resource not found
- `409 Conflict`: Resource conflict (e.g., duplicate email)
- `500 Internal Server Error`: Server error

## Notification Types

- `TASK_ASSIGNED`: Approval task assigned to user
- `TASK_APPROVED`: Task was approved
- `TASK_REJECTED`: Task was rejected
- `TASK_REASSIGNED`: Task was reassigned
- `DOCUMENT_SUBMITTED`: Document submitted for approval
- `DOCUMENT_APPROVED`: Document fully approved

## Document Types

- `REQUISITION`: Purchase requisition
- `BUDGET`: Budget proposal
- `PURCHASE_ORDER`: Purchase order
- `PAYMENT_VOUCHER`: Payment voucher
- `GRN`: Goods Received Note

## User Roles

- `ADMIN`: Full system access
- `MANAGER`: Can view all documents and audit logs
- `DEPARTMENT_MANAGER`: Can manage department documents
- `FINANCE_MANAGER`: Can approve financial documents
- `APPROVER`: Can approve assigned tasks
- `REQUESTER`: Can create and submit documents

---

🔒 = Requires authentication
