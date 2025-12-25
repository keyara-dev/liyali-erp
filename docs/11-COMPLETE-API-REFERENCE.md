# Complete API Reference - All Endpoints

**Status**: Comprehensive endpoint documentation covering entire application
**Last Updated**: 2025-12-12
**Scope**: Phase 11 & 12 implementations
**Total Endpoints**: 80+

---

## Table of Contents

1. [Authentication & Auth](#1-authentication--auth)
2. [Document Management](#2-document-management)
3. [Search & Filter](#3-search--filter)
4. [Approval Workflows](#4-approval-workflows)
5. [Bulk Operations](#5-bulk-operations)
6. [User Management](#6-user-management)
7. [Role-Based Access Control (RBAC)](#7-role-based-access-control-rbac)
8. [Notifications](#8-notifications)
9. [Analytics & Reporting](#9-analytics--reporting)
10. [Configuration Management](#10-configuration-management)
11. [System & Health](#11-system--health)

---

## 1. Authentication & Auth

### POST /auth/login

Authenticate user with email and password.

**Request Body:**

```json
{
  "email": "user@example.com",
  "password": "securePassword123"
}
```

**Response (200 OK):**

```json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "id": "user-550e8400",
    "name": "John Doe",
    "email": "user@example.com",
    "role": "approver",
    "department": "Finance",
    "permissions": ["APPROVE_DOCUMENTS", "VIEW_REPORTS", "MANAGE_USERS"]
  },
  "status": 200
}
```

**Error (401 Unauthorized):**

```json
{
  "success": false,
  "message": "Invalid email or password",
  "status": 401
}
```

---

### POST /auth/logout

Logout current user and invalidate session.

**Response (200 OK):**

```json
{
  "success": true,
  "message": "Logged out successfully",
  "status": 200
}
```

---

### GET /auth/me

Get current authenticated user profile.

**Response (200 OK):**

```json
{
  "success": true,
  "message": "User retrieved successfully",
  "data": {
    "id": "user-550e8400",
    "name": "John Doe",
    "email": "user@example.com",
    "role": "approver",
    "department": "Finance",
    "permissions": ["APPROVE_DOCUMENTS", "VIEW_REPORTS"],
    "lastLogin": "2025-12-12T10:30:00Z",
    "active": true
  },
  "status": 200
}
```

**Error (401 Unauthorized):**

```json
{
  "success": false,
  "message": "No authenticated user found",
  "status": 401
}
```

---

### POST /auth/refresh-token

Refresh JWT access token.

**Response (200 OK):**

```json
{
  "success": true,
  "message": "Token refreshed",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expiresIn": 3600
  },
  "status": 200
}
```

---

### POST /auth/change-password

Change password for authenticated user.

**Request Body:**

```json
{
  "currentPassword": "oldPassword123",
  "newPassword": "newPassword456"
}
```

**Response (200 OK):**

```json
{
  "success": true,
  "message": "Password changed successfully",
  "status": 200
}
```

---

### POST /auth/send-reset-email

Send password reset email.

**Request Body:**

```json
{
  "email": "user@example.com"
}
```

**Response (200 OK):**

```json
{
  "success": true,
  "message": "Reset email sent",
  "status": 200
}
```

---

### POST /auth/reset-password

Reset password using reset token.

**Request Body:**

```json
{
  "token": "reset-token-abc123",
  "newPassword": "newPassword456"
}
```

**Response (200 OK):**

```json
{
  "success": true,
  "message": "Password reset successfully",
  "status": 200
}
```

---

### POST /auth/register

Create new user account.

**Request Body:**

```json
{
  "name": "Jane Smith",
  "email": "jane@example.com",
  "password": "securePassword123",
  "department": "Operations"
}
```

**Response (201 Created):**

```json
{
  "success": true,
  "message": "Account created successfully",
  "data": {
    "id": "user-new-123",
    "name": "Jane Smith",
    "email": "jane@example.com",
    "role": "requester"
  },
  "status": 201
}
```

---

### GET /auth/demo-users

Get list of available demo users for testing.

**Response (200 OK):**

```json
{
  "success": true,
  "message": "Demo users retrieved",
  "data": [
    {
      "id": "demo-requester",
      "name": "Demo Requester",
      "email": "requester@demo.com",
      "role": "requester",
      "password": "demo123"
    },
    {
      "id": "demo-approver",
      "name": "Demo Approver",
      "email": "approver@demo.com",
      "role": "approver",
      "password": "demo123"
    }
  ],
  "status": 200
}
```

---

### GET /auth/check-signup-availability

Check if new user registration is available.

**Response (200 OK):**

```json
{
  "success": true,
  "message": "Signup availability checked",
  "data": {
    "available": true,
    "reason": "Open registration"
  },
  "status": 200
}
```

---

### POST /auth/screen-lock

Lock screen on user idle timeout.

**Request Body:**

```json
{
  "idleTimeMinutes": 15
}
```

**Response (200 OK):**

```json
{
  "success": true,
  "message": "Screen locked",
  "status": 200
}
```

---

### GET /auth/screen-lock-state

Check if screen is locked.

**Response (200 OK):**

```json
{
  "success": true,
  "data": {
    "isLocked": false
  },
  "status": 200
}
```

---

## 2. Document Management

### 2.1 Purchase Orders

#### GET /api/purchase-orders

List all purchase orders with filtering and pagination.

**Query Parameters:**

```
limit=10, page=1, status=ALL, creatorId=user-123
startDate=2024-01-01, endDate=2024-12-31
sortBy=createdAt, sortOrder=DESC, search=PO-2024
```

**Response (200 OK):**

```json
{
  "success": true,
  "data": {
    "documents": [
      {
        "id": "po-550e8400",
        "type": "PURCHASE_ORDER",
        "documentNumber": "PO-2024-001",
        "status": "APPROVED",
        "currentStage": 4,
        "createdBy": "user-1",
        "createdAt": "2024-12-01T10:30:00Z",
        "metadata": {
          "vendorName": "Vendor Ltd",
          "totalAmount": 18750,
          "currency": "ZMW"
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
  "status": 200
}
```

#### POST /api/purchase-orders

Create new purchase order.

**Request Body:**

```json
{
  "documentNumber": "PO-2024-NEW",
  "vendorName": "New Vendor",
  "vendorId": "VENDOR-NEW",
  "totalAmount": 50000,
  "currency": "ZMW",
  "items": [
    {
      "description": "Item 1",
      "quantity": 10,
      "unitCost": 5000
    }
  ]
}
```

**Response (201 Created):** Same format as GET

#### GET /api/purchase-orders/:id

Get specific purchase order.

#### PUT /api/purchase-orders/:id

Update purchase order (DRAFT only).

#### DELETE /api/purchase-orders/:id

Delete purchase order (DRAFT only).

---

### 2.2 Requisitions

#### GET /api/requisitions

List all requisitions.

#### POST /api/requisitions

Create new requisition.

**Request Body:**

```json
{
  "documentNumber": "REQ-2024-NEW",
  "department": "IT",
  "amount": 100000,
  "currency": "ZMW",
  "justification": "Equipment upgrade",
  "items": [
    {
      "description": "Laptops",
      "quantity": 5,
      "estimatedCost": 100000
    }
  ]
}
```

#### GET /api/requisitions/:id

Get specific requisition.

#### PUT /api/requisitions/:id

Update requisition.

#### DELETE /api/requisitions/:id

Delete requisition.

---

### 2.3 Payment Vouchers

#### GET /api/payment-vouchers

List all payment vouchers.

#### POST /api/payment-vouchers

Create new payment voucher.

**Request Body:**

```json
{
  "documentNumber": "PV-2024-NEW",
  "payeeName": "Service Provider",
  "payeeAccount": "123456789",
  "amount": 50000,
  "currency": "ZMW",
  "invoiceNumber": "INV-2024-500",
  "paymentDate": "2024-12-15"
}
```

#### GET /api/payment-vouchers/:id

Get specific payment voucher.

#### PUT /api/payment-vouchers/:id

Update payment voucher.

#### DELETE /api/payment-vouchers/:id

Delete payment voucher.

---

### 2.4 Goods Received Notes

#### GET /api/goods-received-notes

List all GRNs.

#### POST /api/goods-received-notes

Create new GRN.

**Request Body:**

```json
{
  "documentNumber": "GRN-2024-NEW",
  "poId": "po-550e8400",
  "poNumber": "PO-2024-001",
  "vendorName": "Vendor Ltd",
  "receivedQuantity": 15,
  "totalQuantity": 15,
  "amount": 18750,
  "receivedDate": "2024-12-10"
}
```

#### GET /api/goods-received-notes/:id

Get specific GRN.

#### PUT /api/goods-received-notes/:id

Update GRN.

#### DELETE /api/goods-received-notes/:id

Delete GRN.

---

### 2.5 Budgets

#### GET /api/budgets

List all budgets.

#### POST /api/budgets

Create new budget.

**Request Body:**

```json
{
  "name": "Q1 2025 IT Budget",
  "description": "Quarterly IT department budget",
  "amount": 500000,
  "currency": "ZMW",
  "department": "IT",
  "startDate": "2025-01-01",
  "endDate": "2025-03-31"
}
```

#### GET /api/budgets/:id

Get specific budget.

#### PUT /api/budgets/:id

Update budget.

#### DELETE /api/budgets/:id

Delete budget.

---

## 3. Search & Filter

### GET /api/search

Unified search across all document types.

**Query Parameters:**

```
documentNumber=PO-2024
documentType=PURCHASE_ORDER  (or ALL)
status=APPROVED  (or ALL)
startDate=2024-01-01
endDate=2024-12-31
limit=10, page=1
sortBy=createdAt, sortOrder=DESC
```

**Response (200 OK):**

```json
{
  "success": true,
  "data": {
    "results": [
      {
        "id": "po-550e8400",
        "type": "PURCHASE_ORDER",
        "documentNumber": "PO-2024-001",
        "status": "APPROVED",
        "createdAt": "2024-12-01T10:30:00Z",
        "amount": 18750,
        "creator": "John Mwale"
      }
    ],
    "pagination": {
      "total": 42,
      "page": 1,
      "limit": 10,
      "totalPages": 5
    }
  },
  "status": 200
}
```

---

## 4. Approval Workflows

### GET /api/approvals/tasks

Get approval tasks assigned to current user.

**Query Parameters:**

```
status=pending  (pending, approved, rejected)
limit=10, page=1
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
        "documentId": "po-550e8400",
        "documentNumber": "PO-2024-002",
        "documentType": "PURCHASE_ORDER",
        "status": "pending",
        "currentStage": 1,
        "assignedTo": "user-5",
        "createdAt": "2024-12-08T13:00:00Z",
        "dueDate": "2024-12-15"
      }
    ],
    "pagination": {
      "total": 24,
      "page": 1,
      "limit": 10,
      "totalPages": 3
    }
  },
  "status": 200
}
```

---

### GET /api/approvals/tasks/:taskId

Get specific approval task details.

**Response (200 OK):**

```json
{
  "success": true,
  "data": {
    "id": "task-550e8400",
    "documentId": "po-550e8400",
    "document": {
      "id": "po-550e8400",
      "type": "PURCHASE_ORDER",
      "documentNumber": "PO-2024-002",
      "status": "IN_REVIEW"
    },
    "stage": 1,
    "assignedTo": "user-5",
    "createdAt": "2024-12-08T13:00:00Z"
  },
  "status": 200
}
```

---

### POST /api/approvals/tasks/:taskId/approve

Approve a task.

**Request Body:**

```json
{
  "approverId": "user-5",
  "signature": "base64-encoded-signature",
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
    "documentId": "po-550e8400",
    "previousStatus": "pending",
    "newStatus": "approved",
    "nextStage": 2
  },
  "message": "Task approved successfully",
  "status": 200
}
```

---

### POST /api/approvals/tasks/:taskId/reject

Reject a task.

**Request Body:**

```json
{
  "rejectingUserId": "user-5",
  "signature": "base64-encoded-signature",
  "remarks": "Requires vendor clarification"
}
```

**Response (200 OK):**

```json
{
  "success": true,
  "data": {
    "taskId": "task-550e8400",
    "documentId": "po-550e8400",
    "status": "rejected"
  },
  "message": "Task rejected successfully",
  "status": 200
}
```

---

### POST /api/approvals/tasks/:taskId/reassign

Reassign task to another approver.

**Request Body:**

```json
{
  "reassignedBy": "user-5",
  "newApproverId": "user-6",
  "newApproverName": "Secondary Approver",
  "reason": "Subject matter expert delegation"
}
```

**Response (200 OK):**

```json
{
  "success": true,
  "data": {
    "taskId": "task-550e8400",
    "previousAssignee": "user-5",
    "newAssignee": "user-6"
  },
  "message": "Task reassigned successfully",
  "status": 200
}
```

---

### POST /api/approvals/submit

Submit document for approval.

**Request Body:**

```json
{
  "documentId": "po-550e8400",
  "documentType": "PURCHASE_ORDER",
  "submittedBy": "user-1"
}
```

**Response (200 OK):**

```json
{
  "success": true,
  "data": {
    "documentId": "po-550e8400",
    "status": "SUBMITTED",
    "currentStage": 1
  },
  "message": "Document submitted for approval",
  "status": 200
}
```

---

### GET /api/approvals/history/:documentId

Get approval history for a document.

**Response (200 OK):**

```json
{
  "success": true,
  "data": [
    {
      "id": "history-1",
      "documentId": "po-550e8400",
      "stage": 1,
      "action": "APPROVED",
      "approver": "user-5",
      "approverName": "John Approver",
      "timestamp": "2024-12-08T14:30:00Z",
      "comments": "Approved",
      "signature": "base64-signature"
    }
  ],
  "status": 200
}
```

---

### POST /api/approvals/reverse

Reverse an approved document (return to draft).

**Request Body:**

```json
{
  "documentId": "po-550e8400",
  "reversedBy": "user-5",
  "reason": "Need to modify vendor details"
}
```

**Response (200 OK):**

```json
{
  "success": true,
  "data": {
    "documentId": "po-550e8400",
    "previousStatus": "APPROVED",
    "newStatus": "REVERSED"
  },
  "message": "Document reversed successfully",
  "status": 200
}
```

---

### POST /api/approvals/validate-signature

Validate signature for document approval.

**Request Body:**

```json
{
  "signature": "base64-encoded-signature",
  "userId": "user-5"
}
```

**Response (200 OK):**

```json
{
  "success": true,
  "data": {
    "isValid": true,
    "userId": "user-5"
  },
  "status": 200
}
```

---

### GET /api/approvals/available-approvers/:taskId

Get list of available approvers for reassignment.

**Response (200 OK):**

```json
{
  "success": true,
  "data": [
    {
      "id": "user-6",
      "name": "Alternative Approver",
      "email": "alt@example.com",
      "role": "approver"
    }
  ],
  "status": 200
}
```

---

## 5. Bulk Operations

### POST /api/approvals/bulk/approve

Approve multiple tasks at once.

**Request Body:**

```json
{
  "taskIds": ["task-550e8400", "task-550e8401", "task-550e8402"],
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
        "status": "success"
      }
    ]
  },
  "message": "3 tasks approved successfully",
  "status": 200
}
```

---

### POST /api/approvals/bulk/reject

Reject multiple tasks at once.

**Request Body:**

```json
{
  "taskIds": ["task-550e8400", "task-550e8401"],
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
  "message": "2 tasks rejected successfully",
  "status": 200
}
```

---

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
  "message": "2 tasks reassigned successfully",
  "status": 200
}
```

---

## 6. User Management

### GET /api/users

List all users in the system.

**Query Parameters:**

```
role=approver  (requester, approver, admin, finance, warehouse)
department=IT
active=true
limit=50, page=1
search=john
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
  },
  "status": 200
}
```

---

### GET /api/users/:userId

Get specific user profile.

**Response (200 OK):**

```json
{
  "success": true,
  "data": {
    "id": "user-1",
    "name": "John Mwale",
    "email": "john@example.com",
    "role": "requester",
    "department": "Operations",
    "permissions": ["CREATE_REQUISITION", "SUBMIT_APPROVAL"],
    "active": true,
    "lastLogin": "2025-12-12T10:00:00Z"
  },
  "status": 200
}
```

---

### POST /api/users

Create new user (admin only).

**Request Body:**

```json
{
  "name": "Jane Smith",
  "email": "jane@example.com",
  "role": "approver",
  "department": "Finance",
  "initialPassword": "tempPassword123"
}
```

**Response (201 Created):**

```json
{
  "success": true,
  "data": {
    "id": "user-new-123",
    "name": "Jane Smith",
    "email": "jane@example.com",
    "role": "approver"
  },
  "status": 201
}
```

---

### PUT /api/users/:userId

Update user information (admin or self).

**Request Body:**

```json
{
  "name": "Jane Doe",
  "department": "Operations",
  "role": "manager"
}
```

**Response (200 OK):**

```json
{
  "success": true,
  "data": {
    "id": "user-1",
    "name": "Jane Doe",
    "department": "Operations",
    "role": "manager"
  },
  "message": "User updated successfully",
  "status": 200
}
```

---

### DELETE /api/users/:userId

Deactivate or delete user (admin only).

**Response (200 OK):**

```json
{
  "success": true,
  "message": "User deleted successfully",
  "status": 200
}
```

---

### POST /api/users/:userId/activate

Activate a deactivated user.

**Response (200 OK):**

```json
{
  "success": true,
  "message": "User activated successfully",
  "status": 200
}
```

---

### POST /api/users/:userId/reset-password

Admin reset user password.

**Request Body:**

```json
{
  "newPassword": "tempPassword456"
}
```

**Response (200 OK):**

```json
{
  "success": true,
  "message": "Password reset successfully",
  "status": 200
}
```

---

## 7. Role-Based Access Control (RBAC)

### GET /api/rbac/roles

Get all roles in system.

**Response (200 OK):**

```json
{
  "success": true,
  "data": [
    {
      "id": "role-admin",
      "name": "Administrator",
      "description": "Full system access",
      "isBuiltIn": true,
      "permissions": ["*"]
    },
    {
      "id": "role-approver",
      "name": "Approver",
      "description": "Approve documents",
      "isBuiltIn": true,
      "permissions": ["APPROVE_DOCUMENTS", "VIEW_REPORTS"]
    }
  ],
  "status": 200
}
```

---

### GET /api/rbac/roles/builtin

Get built-in roles only.

---

### GET /api/rbac/roles/custom

Get custom roles only.

---

### GET /api/rbac/roles/:roleId

Get specific role details.

**Response (200 OK):**

```json
{
  "success": true,
  "data": {
    "id": "role-approver",
    "name": "Approver",
    "description": "Document approval role",
    "isBuiltIn": true,
    "permissions": [
      "APPROVE_DOCUMENTS",
      "REJECT_DOCUMENTS",
      "VIEW_DOCUMENTS",
      "VIEW_REPORTS"
    ]
  },
  "status": 200
}
```

---

### POST /api/rbac/roles

Create custom role (admin only).

**Request Body:**

```json
{
  "name": "Department Manager",
  "description": "Manage department requisitions",
  "permissions": ["CREATE_REQUISITION", "APPROVE_REQUISITION", "VIEW_REPORTS"]
}
```

**Response (201 Created):**

```json
{
  "success": true,
  "data": {
    "id": "role-custom-123",
    "name": "Department Manager",
    "permissions": ["CREATE_REQUISITION", "APPROVE_REQUISITION", "VIEW_REPORTS"]
  },
  "status": 201
}
```

---

### PUT /api/rbac/roles/:roleId

Update custom role.

**Request Body:**

```json
{
  "name": "Updated Role Name",
  "description": "Updated description",
  "permissions": ["PERMISSION_1", "PERMISSION_2"]
}
```

**Response (200 OK):**

```json
{
  "success": true,
  "data": {
    "id": "role-custom-123",
    "name": "Updated Role Name"
  },
  "status": 200
}
```

---

### DELETE /api/rbac/roles/:roleId

Delete custom role (built-in roles cannot be deleted).

---

### POST /api/rbac/roles/:roleId/permissions

Add permission to role.

**Request Body:**

```json
{
  "permission": "EXPORT_DOCUMENTS"
}
```

---

### DELETE /api/rbac/roles/:roleId/permissions/:permission

Remove permission from role.

---

### GET /api/rbac/permissions

Get all available permissions.

**Response (200 OK):**

```json
{
  "success": true,
  "data": [
    {
      "id": "APPROVE_DOCUMENTS",
      "name": "Approve Documents",
      "description": "Approve workflow documents",
      "category": "APPROVAL"
    },
    {
      "id": "CREATE_REQUISITION",
      "name": "Create Requisition",
      "description": "Create new requisitions",
      "category": "DOCUMENT_CREATION"
    }
  ],
  "status": 200
}
```

---

### POST /api/rbac/check

Check if user has specific permission.

**Request Body:**

```json
{
  "userId": "user-1",
  "permission": "APPROVE_DOCUMENTS"
}
```

**Response (200 OK):**

```json
{
  "success": true,
  "data": {
    "hasPermission": true,
    "userId": "user-1",
    "permission": "APPROVE_DOCUMENTS"
  },
  "status": 200
}
```

---

## 8. Notifications

### GET /api/notifications

Get notifications for current user.

**Query Parameters:**

```
page=1, pageSize=20
type=ASSIGNMENT  (type filter)
isRead=false  (unread only)
startDate=2024-12-01, endDate=2024-12-31
```

**Response (200 OK):**

```json
{
  "success": true,
  "data": {
    "notifications": [
      {
        "id": "notif-550e8400",
        "userId": "user-5",
        "documentId": "po-550e8400",
        "type": "ASSIGNMENT",
        "message": "You have been assigned to approve PO-2024-001",
        "isRead": false,
        "createdAt": "2024-12-12T10:00:00Z"
      }
    ],
    "total": 15,
    "page": 1,
    "pageSize": 20,
    "hasMore": false
  },
  "status": 200
}
```

---

### GET /api/notifications/unread-count

Get count of unread notifications.

**Response (200 OK):**

```json
{
  "success": true,
  "data": {
    "unreadCount": 5
  },
  "status": 200
}
```

---

### POST /api/notifications/:notificationId/read

Mark notification as read.

**Response (200 OK):**

```json
{
  "success": true,
  "message": "Notification marked as read",
  "status": 200
}
```

---

### POST /api/notifications/read-all

Mark all notifications as read.

**Response (200 OK):**

```json
{
  "success": true,
  "message": "All notifications marked as read",
  "status": 200
}
```

---

### DELETE /api/notifications/:notificationId

Delete notification.

---

### GET /api/notifications/preferences

Get user notification preferences.

**Response (200 OK):**

```json
{
  "success": true,
  "data": {
    "emailNotifications": true,
    "smsNotifications": false,
    "pushNotifications": true,
    "notificationTypes": {
      "ASSIGNMENT": true,
      "APPROVAL": true,
      "REJECTION": true,
      "SYSTEM": false
    }
  },
  "status": 200
}
```

---

### PUT /api/notifications/preferences

Update notification preferences.

**Request Body:**

```json
{
  "emailNotifications": true,
  "smsNotifications": false,
  "notificationTypes": {
    "ASSIGNMENT": true,
    "APPROVAL": false
  }
}
```

---

## 9. Analytics & Reporting

### GET /api/analytics/dashboard

Get dashboard metrics and KPIs.

**Query Parameters:**

```
startDate=2024-01-01
endDate=2024-12-31
groupBy=week  (day, week, month)
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
    "metrics": {
      "highestAmountDocument": {
        "documentNumber": "PO-2024-007",
        "amount": 125000
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
      ]
    }
  },
  "status": 200
}
```

---

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
        "recommendation": "Add approvers to this stage"
      }
    ]
  },
  "status": 200
}
```

---

### GET /api/analytics/metrics

Get approval metrics by user/department.

**Query Parameters:**

```
userId=user-5
department=Finance
startDate=2024-01-01, endDate=2024-12-31
```

**Response (200 OK):**

```json
{
  "success": true,
  "data": {
    "approvals": 125,
    "rejections": 8,
    "pendingCount": 5,
    "averageTimeToApprove": "2.1 days",
    "approvalRate": 0.94
  },
  "status": 200
}
```

---

### GET /api/analytics/trends

Get workflow trends over time.

**Query Parameters:**

```
startDate=2024-01-01
endDate=2024-12-31
groupBy=week
```

---

## 10. Configuration Management

### GET /api/config/branches

List all branches/locations.

**Response (200 OK):**

```json
{
  "success": true,
  "data": [
    {
      "id": "branch-001",
      "name": "Head Office",
      "location": "Lusaka",
      "code": "HO"
    },
    {
      "id": "branch-002",
      "name": "Northern Region",
      "location": "Ndola",
      "code": "NR"
    }
  ],
  "status": 200
}
```

---

### GET /api/config/branches/:id

Get specific branch.

---

### POST /api/config/branches

Create new branch (admin only).

**Request Body:**

```json
{
  "name": "Southern Region",
  "location": "Livingstone",
  "code": "SR"
}
```

---

### PUT /api/config/branches/:id

Update branch.

---

### DELETE /api/config/branches/:id

Delete branch.

---

### GET /api/config/currencies

Get list of supported currencies.

**Response (200 OK):**

```json
{
  "success": true,
  "data": [
    {
      "code": "ZMW",
      "name": "Zambian Kwacha",
      "symbol": "K"
    },
    {
      "code": "USD",
      "name": "US Dollar",
      "symbol": "$"
    }
  ],
  "status": 200
}
```

---

### GET /api/config/event-hosts

Get event hosts for calendar.

---

### GET /api/config/premium

Get premium configuration settings.

---

## 11. System & Health

### GET /api/health

Health check endpoint.

**Response (200 OK):**

```json
{
  "success": true,
  "status": "healthy",
  "timestamp": "2025-12-12T10:30:00Z",
  "version": "1.0.0",
  "uptime": 3600,
  "checks": {
    "database": "ok",
    "cache": "ok",
    "storage": "ok"
  }
}
```

---

### GET /api/config

Get system configuration.

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
      "GOODS_RECEIVED_NOTE",
      "BUDGET"
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
  },
  "status": 200
}
```

---

## Standard Response Format

All endpoints follow this standardized response format:

### Success Response

```json
{
  "success": true,
  "message": "Operation successful",
  "data": {},
  "status": 200,
  "statusText": "OK"
}
```

### Error Response

```json
{
  "success": false,
  "message": "Error description",
  "error": {
    "code": "ERROR_CODE",
    "details": "Additional details"
  },
  "status": 400,
  "statusText": "BAD_REQUEST"
}
```

---

## HTTP Status Codes

| Code | Meaning      | Usage                          |
| ---- | ------------ | ------------------------------ |
| 200  | OK           | Successful GET/PUT/DELETE      |
| 201  | Created      | Successful POST                |
| 400  | Bad Request  | Invalid parameters             |
| 401  | Unauthorized | Missing/invalid authentication |
| 403  | Forbidden    | Insufficient permissions       |
| 404  | Not Found    | Resource doesn't exist         |
| 409  | Conflict     | Resource conflict              |
| 500  | Server Error | Unexpected server error        |
| 503  | Unavailable  | Service unavailable            |

---

## Authentication

All endpoints (except `/auth/*` and `/api/health`) require JWT token:

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

---

## Rate Limiting

All endpoints are rate-limited:

```
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1702400400
```

---

## Endpoint Summary

**Total Endpoints: 80+**

| Category        | Count | Examples                                |
| --------------- | ----- | --------------------------------------- |
| Authentication  | 11    | Login, Register, Password Reset         |
| Documents       | 25    | CRUD for 5 document types               |
| Approvals       | 10    | Approve, Reject, Reassign, History      |
| Bulk Operations | 3     | Bulk Approve, Reject, Reassign          |
| Users           | 8     | List, Create, Update, Delete, Reset     |
| RBAC            | 9     | Roles, Permissions, Check Access        |
| Notifications   | 6     | Get, Mark Read, Preferences             |
| Analytics       | 4     | Dashboard, Bottlenecks, Metrics, Trends |
| Configuration   | 6     | Branches, Currencies, Hosts             |
| System          | 2     | Health, Config                          |

---

## Implementation Roadmap

### Phase 12a (Weeks 1-2)

- [ ] Implement auth endpoints
- [ ] Implement document CRUD endpoints
- [ ] Implement user management

### Phase 12b (Weeks 3-4)

- [ ] Implement approval workflow endpoints
- [ ] Implement RBAC endpoints
- [ ] Add authentication/authorization middleware

### Phase 12c (Weeks 5-6)

- [ ] Implement bulk operations
- [ ] Implement notifications
- [ ] Implement analytics endpoints

### Phase 12d (Weeks 7-8)

- [ ] Implement configuration endpoints
- [ ] Add comprehensive testing
- [ ] Performance optimization

---

## Notes

- All dates should be in ISO 8601 format (YYYY-MM-DDTHH:mm:ssZ)
- All IDs are UUIDs (UUID v4)
- Pagination uses zero-based indexing
- All monetary values are in the specified currency
- Timestamps are in UTC

---

**Total Endpoints Documented**: 80+
**Last Updated**: 2025-12-12
**Ready for Implementation**: Yes

---
