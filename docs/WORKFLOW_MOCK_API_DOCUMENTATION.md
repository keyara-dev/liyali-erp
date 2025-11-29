# Workflow Management System - Mocked Server Actions Documentation

## Overview

This document outlines all the mocked server actions created for the workflow management system. These actions simulate backend API calls and work with an in-memory data store, allowing for complete frontend development and testing without requiring a backend server.

## Quick Start

### Test All Functionality
```bash
curl http://localhost:3001/api/workflows/test
```

This endpoint will execute a comprehensive test of all mocked server actions and return detailed results.

---

## Architecture

### File Structure

```
src/
├── types/
│   └── workflow.ts                    # All TypeScript type definitions
├── lib/
│   ├── rbac.ts                        # Role-Based Access Control system
│   └── mock-data.ts                   # Mock data factories
├── app/_actions/
│   ├── workflow.ts                    # Workflow document operations (mocked)
│   ├── rbac.ts                        # Role and permission management (mocked)
│   └── user-management.ts             # User and role assignment (mocked)
└── app/api/
    └── workflows/
        └── test/
            └── route.ts               # Test endpoint for all operations
```

### Data Storage

All mocked data is stored in-memory using JavaScript `Map` objects:

- `documentStore` - Stores workflow documents
- `approversStore` - Stores approver assignments
- `approvalLogsStore` - Stores approval history (immutable audit trail)
- `attachmentsStore` - Stores attachments
- `customRolesStore` - Stores custom role definitions (in RBAC)
- `userRoleAssignmentsStore` - Stores user-role assignments (in User Management)

**Note**: Data is cleared when the server restarts. For production, replace with a database.

---

## Core Components

### 1. Type Definitions (`src/types/workflow.ts`)

#### WorkflowDocumentType
```typescript
type WorkflowDocumentType = 'PURCHASE_ORDER' | 'PAYMENT_VOUCHER' | 'REQUISITION'
```

#### DocumentStatus
```typescript
type DocumentStatus = 'DRAFT' | 'SUBMITTED' | 'IN_APPROVAL' | 'APPROVED' | 'REJECTED'
```

#### Workflow Document
```typescript
type WorkflowDocument = {
  id: string
  type: WorkflowDocumentType
  documentNumber: string          // Auto-generated (e.g., PO-123456-789)
  status: DocumentStatus
  currentStage: number            // Which approval step we're at
  createdBy: string
  createdByUser?: User
  createdAt: Date
  updatedAt: Date
  metadata: Record<string, any>   // Form-specific data
}
```

#### User Roles
```typescript
type UserRole =
  | 'REQUESTER'              // Creates/submits documents
  | 'DEPARTMENT_MANAGER'     // First level approval
  | 'FINANCE_OFFICER'        // Finance verification
  | 'DIRECTOR'               // Senior approval
  | 'CFO'                    // Executive approval
  | 'COMPLIANCE_OFFICER'     // Audit trail access
  | 'ADMIN'                  // Full system access
```

#### Permissions (Consolidated)
```typescript
type Permission =
  | 'view_draft'
  | 'edit_draft'
  | 'submit_document'
  | 'approve_document'
  | 'reject_document'
  | 'reassign_approver'
  | 'view_attachments'
  | 'add_attachments'
  | 'view_comments'
  | 'add_comments'
  | 'view_audit_log'
  | 'manage_approvers'
  | 'manage_workflows'
```

---

## Workflow Server Actions (`src/app/_actions/workflow.ts`)

### Document Operations

#### `createWorkflowDocument(documentType, formData)`
Creates a new workflow document draft.

**Parameters:**
- `documentType` - 'PURCHASE_ORDER' | 'PAYMENT_VOUCHER' | 'REQUISITION'
- `formData` - Document-specific data object

**Returns:**
```typescript
APIResponse<WorkflowDocument>
```

**Example:**
```typescript
const result = await createWorkflowDocument('PURCHASE_ORDER', {
  vendorName: 'Acme Corp',
  vendorId: 'V-001',
  items: [
    { id: '1', description: 'Item 1', quantity: 5, unitCost: 100, totalCost: 500 }
  ],
  totalAmount: 500,
  currency: 'ZMW',
  deliveryDate: new Date(),
  specialInstructions: 'Urgent delivery needed'
})

if (result.success) {
  console.log('Document created:', result.data.documentNumber)
}
```

#### `submitDocument(documentId)`
Submits a draft document for approval. Auto-assigns first approver.

**Parameters:**
- `documentId` - ID of draft document

**Returns:**
```typescript
APIResponse<WorkflowDocument>
```

#### `getDocument(documentId)`
Retrieves a workflow document by ID.

**Parameters:**
- `documentId` - Document ID

**Returns:**
```typescript
APIResponse<WorkflowDocument>
```

#### `updateDocumentDraft(documentId, formData)`
Updates draft document data. Only works when status is 'DRAFT' or 'REJECTED'.

**Parameters:**
- `documentId` - Document ID
- `formData` - Updated form data

**Returns:**
```typescript
APIResponse<WorkflowDocument>
```

#### `getDocumentsByCreator(userId, page, limit)`
Retrieves all documents created by a specific user (paginated).

**Parameters:**
- `userId` - Creator's user ID
- `page` - Page number (default: 1)
- `limit` - Items per page (default: 10)

**Returns:**
```typescript
APIResponse<PaginatedResponse<WorkflowDocument>>
```

---

### Approval Operations

#### `approveDocument(documentId, comments?)`
Approves a document at current stage. Moves to next stage or marks as APPROVED.

**Parameters:**
- `documentId` - Document ID
- `comments` - Optional approval comments

**Returns:**
```typescript
APIResponse<ApprovalLogEntry>
```

**Workflow:**
1. Checks document status is 'IN_APPROVAL'
2. Creates immutable approval log entry
3. Updates approver status to 'APPROVED'
4. If last step: marks document as 'APPROVED'
5. Otherwise: advances to next stage and auto-assigns next approver

#### `rejectDocument(documentId, reason)`
Rejects a document and returns it to DRAFT status for modifications.

**Parameters:**
- `documentId` - Document ID
- `reason` - Rejection reason (required)

**Returns:**
```typescript
APIResponse<ApprovalLogEntry>
```

**Workflow:**
1. Checks document status is 'IN_APPROVAL'
2. Creates rejection log entry
3. Sets document status to 'REJECTED'
4. Resets currentStage to 0

#### `getApprovalLog(documentId)`
Retrieves all approval history for a document.

**Parameters:**
- `documentId` - Document ID

**Returns:**
```typescript
APIResponse<ApprovalLogEntry[]>
```

#### `getPendingApprovals(userRole)`
Retrieves all documents awaiting approval by a specific role.

**Parameters:**
- `userRole` - User's role (e.g., 'DEPARTMENT_MANAGER')

**Returns:**
```typescript
APIResponse<WorkflowDocument[]>
```

---

### Approver Management

#### `assignApprover(documentId, stepOrder, userId, role)`
Manually assigns an approver to a specific workflow step.

**Parameters:**
- `documentId` - Document ID
- `stepOrder` - Step number (1, 2, 3...)
- `userId` - User ID of approver
- `role` - User's role

**Returns:**
```typescript
APIResponse<Approver>
```

#### `reassignApprover(documentId, approverId, newUserId)`
Reassigns an approver to a different user (if canReassign is true).

**Parameters:**
- `documentId` - Document ID
- `approverId` - Approver assignment ID
- `newUserId` - New user's ID

**Returns:**
```typescript
APIResponse<Approver>
```

#### `getDocumentApprovers(documentId)`
Retrieves all approver assignments for a document.

**Parameters:**
- `documentId` - Document ID

**Returns:**
```typescript
APIResponse<Approver[]>
```

---

### Attachment Operations

#### `uploadAttachment(documentId, fileName, fileSize, fileType, visibleToRoles)`
Uploads an attachment to a document with role-based visibility.

**Parameters:**
- `documentId` - Document ID
- `fileName` - File name
- `fileSize` - File size in bytes
- `fileType` - MIME type (e.g., 'application/pdf')
- `visibleToRoles` - Array of roles that can view

**Returns:**
```typescript
APIResponse<Attachment>
```

#### `getAttachments(documentId)`
Retrieves all attachments for a document.

**Parameters:**
- `documentId` - Document ID

**Returns:**
```typescript
APIResponse<Attachment[]>
```

#### `deleteAttachment(documentId, attachmentId)`
Deletes an attachment.

**Parameters:**
- `documentId` - Document ID
- `attachmentId` - Attachment ID

**Returns:**
```typescript
APIResponse
```

---

### Workflow Configuration

#### `getWorkflowSteps(documentType)`
Retrieves the predefined approval steps for a document type.

**Parameters:**
- `documentType` - 'PURCHASE_ORDER' | 'PAYMENT_VOUCHER' | 'REQUISITION'

**Returns:**
```typescript
APIResponse<WorkflowStep[]>
```

**Example Response:**
```typescript
[
  {
    stepOrder: 1,
    roleName: 'DEPARTMENT_MANAGER',
    description: 'Department Level Approval',
    isRequired: true
  },
  {
    stepOrder: 2,
    roleName: 'FINANCE_OFFICER',
    description: 'Finance Verification',
    isRequired: true
  }
]
```

---

### Reporting

#### `getDashboardStats(userId)`
Retrieves dashboard statistics for a user.

**Parameters:**
- `userId` - User ID

**Returns:**
```typescript
APIResponse<{
  createdDocuments: number
  pendingApprovals: number
  approvedDocuments: number
  rejectedDocuments: number
}>
```

#### `getAuditLog(documentId)`
Retrieves full audit log (sorted by timestamp).

**Parameters:**
- `documentId` - Document ID

**Returns:**
```typescript
APIResponse<ApprovalLogEntry[]>
```

---

## RBAC Server Actions (`src/app/_actions/rbac.ts`)

### Role Management

#### `getAllRoles()`
Retrieves all roles (built-in and custom).

**Returns:**
```typescript
APIResponse<CustomRole[]>
```

#### `getBuiltInRoles()`
Retrieves only built-in roles (read-only).

**Returns:**
```typescript
APIResponse<CustomRole[]>
```

#### `getCustomRoles()`
Retrieves only custom roles created by admins.

**Returns:**
```typescript
APIResponse<CustomRole[]>
```

#### `getRoleById(roleId)`
Retrieves a specific role by ID.

**Parameters:**
- `roleId` - Role ID

**Returns:**
```typescript
APIResponse<CustomRole>
```

#### `createRole(name, description, permissions)`
Creates a new custom role (admin only).

**Parameters:**
- `name` - Role name (e.g., 'Senior Approver')
- `description` - Role description
- `permissions` - Array of permission codes

**Returns:**
```typescript
APIResponse<CustomRole>
```

**Example:**
```typescript
const result = await createRole(
  'Senior Approver',
  'High-level approvers with full audit access',
  [
    'approve_document',
    'reject_document',
    'reassign_approver',
    'view_audit_log',
    'view_attachments',
    'add_comments'
  ]
)
```

#### `updateRole(roleId, name?, description?)`
Updates a custom role (cannot modify built-in roles).

**Parameters:**
- `roleId` - Role ID
- `name` - New name (optional)
- `description` - New description (optional)

**Returns:**
```typescript
APIResponse<CustomRole>
```

#### `deleteRole(roleId)`
Deletes a custom role (cannot delete built-in roles).

**Parameters:**
- `roleId` - Role ID

**Returns:**
```typescript
APIResponse
```

---

### Permission Management

#### `getAllPermissions()`
Retrieves all available permissions with descriptions.

**Returns:**
```typescript
APIResponse<Array<{
  name: Permission
  description: string
}>>
```

#### `updateRolePermissions(roleId, permissions)`
Replaces all permissions for a role.

**Parameters:**
- `roleId` - Role ID
- `permissions` - New permission array

**Returns:**
```typescript
APIResponse<CustomRole>
```

#### `addRolePermission(roleId, permission)`
Adds a single permission to a role.

**Parameters:**
- `roleId` - Role ID
- `permission` - Permission code

**Returns:**
```typescript
APIResponse<CustomRole>
```

#### `removeRolePermission(roleId, permission)`
Removes a permission from a role.

**Parameters:**
- `roleId` - Role ID
- `permission` - Permission code

**Returns:**
```typescript
APIResponse<CustomRole>
```

---

## User Management Server Actions (`src/app/_actions/user-management.ts`)

### User Retrieval

#### `getAllUsers()`
Retrieves all users in the system.

**Returns:**
```typescript
APIResponse<User[]>
```

#### `getUsersByRole(role)`
Retrieves all users with a specific built-in role.

**Parameters:**
- `role` - Built-in role (e.g., 'DEPARTMENT_MANAGER')

**Returns:**
```typescript
APIResponse<User[]>
```

#### `getUserById(userId)`
Retrieves a specific user.

**Parameters:**
- `userId` - User ID

**Returns:**
```typescript
APIResponse<User>
```

#### `searchUsers(query)`
Searches users by name, email, or department.

**Parameters:**
- `query` - Search term

**Returns:**
```typescript
APIResponse<User[]>
```

---

### User-Role Assignment

#### `assignCustomRoleToUser(userId, customRoleId)`
Assigns a custom role to a user (admin only).

**Parameters:**
- `userId` - User ID
- `customRoleId` - Custom role ID

**Returns:**
```typescript
APIResponse<UserRoleAssignment>
```

#### `removeCustomRoleFromUser(userId, customRoleId)`
Removes a custom role from a user (admin only).

**Parameters:**
- `userId` - User ID
- `customRoleId` - Custom role ID

**Returns:**
```typescript
APIResponse
```

#### `getUserCustomRoles(userId)`
Retrieves all custom roles assigned to a user.

**Parameters:**
- `userId` - User ID

**Returns:**
```typescript
APIResponse
```

#### `getUsersWithRole(customRoleId)`
Retrieves all users assigned to a specific custom role.

**Parameters:**
- `customRoleId` - Custom role ID

**Returns:**
```typescript
APIResponse<User[]>
```

#### `bulkAssignRolesToUser(userId, customRoleIds)`
Assigns multiple roles to a user at once.

**Parameters:**
- `userId` - User ID
- `customRoleIds` - Array of role IDs

**Returns:**
```typescript
APIResponse
```

#### `getAvailableRolesForUser(userId)`
Retrieves roles available for assignment (not yet assigned to user).

**Parameters:**
- `userId` - User ID

**Returns:**
```typescript
APIResponse
```

---

## Default Workflow Paths

### Purchase Order Approval Chain
```
1. DEPARTMENT_MANAGER (Required) - Department approval
2. FINANCE_OFFICER (Required)    - Finance verification
3. DIRECTOR (Required)            - Director approval
4. CFO (Optional)                 - CFO final approval
```

### Payment Voucher Approval Chain
```
1. DEPARTMENT_MANAGER (Required) - Department approval
2. FINANCE_OFFICER (Required)    - Finance processing
3. CFO (Required)                - CFO approval
```

### Requisition Approval Chain
```
1. DEPARTMENT_MANAGER (Required) - Department manager approval
2. DIRECTOR (Required)            - Director approval
3. FINANCE_OFFICER (Required)    - Budget verification
```

---

## Default Role Permissions

### REQUESTER
- view_draft
- edit_draft
- submit_document
- view_attachments
- add_attachments
- view_comments
- add_comments

### DEPARTMENT_MANAGER
- view_draft
- approve_document
- reject_document
- view_attachments
- view_comments
- add_comments
- view_audit_log

### FINANCE_OFFICER
- approve_document
- reject_document
- reassign_approver
- view_attachments
- view_comments
- add_comments
- view_audit_log

### DIRECTOR
- approve_document
- reject_document
- view_attachments
- view_comments
- add_comments
- view_audit_log

### CFO
- approve_document
- reject_document
- view_attachments
- view_comments
- add_comments
- view_audit_log

### COMPLIANCE_OFFICER
- view_audit_log
- view_attachments
- view_comments

### ADMIN
All permissions available

---

## Mock Data

### Default Users

```typescript
// REQUESTER
{
  id: 'user-req-1',
  name: 'John Mwale',
  email: 'john.mwale@company.com',
  role: 'REQUESTER',
  department: 'Operations'
}

// DEPARTMENT_MANAGER
{
  id: 'user-dm-1',
  name: 'James Chileshe',
  email: 'james.chileshe@company.com',
  role: 'DEPARTMENT_MANAGER',
  department: 'Operations'
}

// FINANCE_OFFICER
{
  id: 'user-fo-1',
  name: 'Paul Nkosi',
  email: 'paul.nkosi@company.com',
  role: 'FINANCE_OFFICER',
  department: 'Finance'
}

// ... and more (see MOCK_USERS in src/lib/mock-data.ts)
```

---

## Testing

### Run All Tests
```bash
curl http://localhost:3001/api/workflows/test
```

This comprehensive test:
1. ✅ Tests role management
2. ✅ Tests user management
3. ✅ Creates 3 workflow documents (PO, PV, Requisition)
4. ✅ Submits documents for approval
5. ✅ Tests approval workflow
6. ✅ Retrieves dashboard stats
7. ✅ Tests rejection workflow
8. ✅ Verifies audit logs

---

## Error Handling

All server actions return a consistent response format:

```typescript
interface APIResponse<T> {
  success: boolean
  message: string
  data: T | null
  status: number
  statusText: string
}
```

### Common Status Codes
- **200 OK** - Successful GET/UPDATE/DELETE
- **201 CREATED** - Successful POST
- **400 BAD REQUEST** - Invalid input
- **401 UNAUTHORIZED** - Not authenticated
- **403 FORBIDDEN** - Permission denied
- **404 NOT FOUND** - Resource not found
- **500 INTERNAL SERVER ERROR** - Server error

### Example Error Response
```typescript
{
  success: false,
  message: "Document is not pending approval",
  data: null,
  status: 400,
  statusText: "BAD REQUEST"
}
```

---

## Migration to Real Backend

To migrate from mocked to real backend:

1. **Replace in-memory stores** with database calls
2. **Update server actions** to call actual API endpoints instead of local functions
3. **Keep TypeScript types** - they're agnostic to backend
4. **Preserve error handling** - maintain same response format
5. **Update authentication** - remove mock session injection

The types and API structure will remain the same!

---

## Summary

✅ **Complete mocked workflow system**
✅ **Role-based access control with custom roles**
✅ **User management with role assignment**
✅ **Multi-stage approval workflows**
✅ **Immutable audit trails**
✅ **Attachment management**
✅ **Comprehensive error handling**
✅ **Ready for UI development**
