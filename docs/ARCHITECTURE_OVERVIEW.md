# Workflow System - Architecture Overview

## Architecture Pattern: Server Actions First

This workflow system follows a **Server Actions First** approach:

- ✅ **Server Actions** for all data operations (create, read, update, delete)
- ✅ **In-memory stores** for mocked data during development
- ✅ **API routes** only when needed for external integrations
- ✅ **Type-safe** end-to-end with TypeScript
- ✅ **Authentication** built-in via NextAuth session
- ✅ **Response consistency** across all operations

---

## Directory Structure

```
src/
├── types/
│   └── workflow.ts                          # All TypeScript types
│       ├── WorkflowDocument
│       ├── PurchaseOrder, PaymentVoucher, Requisition
│       ├── Approver, ApprovalLogEntry
│       ├── Attachment
│       ├── User, UserRole, Permission
│       └── Pagination types
│
├── lib/
│   ├── rbac.ts                              # RBAC System
│   │   ├── Role/Permission definitions
│   │   ├── Custom role management functions
│   │   ├── Permission checking utilities
│   │   └── In-memory customRolesStore
│   │
│   └── mock-data.ts                         # Mock Data Factories
│       ├── Mock users (all roles)
│       ├── Document factories (PO, PV, Requisition)
│       ├── Approver/log/attachment factories
│       └── Mock data generation utilities
│
├── app/_actions/
│   ├── workflow.ts                          # ✅ WORKFLOW OPERATIONS
│   │   ├── Document: create, submit, get, update
│   │   ├── Approval: approve, reject, getLog, getPending
│   │   ├── Approver: assign, reassign, get
│   │   ├── Attachment: upload, get, delete
│   │   ├── Configuration: getWorkflowSteps
│   │   ├── Reporting: getDashboardStats, getAuditLog
│   │   └── In-memory: documentStore, approversStore, approvalLogsStore, etc.
│   │
│   ├── rbac.ts                              # ✅ ROLE & PERMISSION MANAGEMENT
│   │   ├── getAllRoles, getBuiltInRoles, getCustomRoles
│   │   ├── createRole, updateRole, deleteRole
│   │   ├── getAllPermissions, updateRolePermissions
│   │   ├── addRolePermission, removeRolePermission
│   │   └── Uses functions from lib/rbac.ts
│   │
│   └── user-management.ts                   # ✅ USER MANAGEMENT
│       ├── getAllUsers, getUsersByRole, searchUsers
│       ├── assignCustomRoleToUser, removeCustomRoleFromUser
│       ├── getUserCustomRoles, getUsersWithRole
│       ├── bulkAssignRolesToUser, getAvailableRolesForUser
│       └── In-memory: userRoleAssignmentsStore
│
├── app/api/
│   └── workflows/
│       └── test/
│           └── route.ts                     # 🧪 Comprehensive test endpoint
│               └── GET /api/workflows/test - Tests all server actions
│
└── components/
    └── [TODO: Build UI components using server actions]
```

---

## Data Flow

### Document Creation & Approval

```
┌─────────────────────────────────────────────────────────┐
│                    CLIENT (Component)                     │
│                   (React Component)                       │
└────────────────────┬────────────────────────────────────┘
                     │ (User Action)
                     ▼
┌─────────────────────────────────────────────────────────┐
│                  SERVER ACTION                           │
│         ('use server' function in app/_actions/)        │
│                                                          │
│  1. Verify authentication (NextAuth session)             │
│  2. Check authorization (RBAC)                           │
│  3. Validate input                                       │
│  4. Operate on in-memory store                           │
│  5. Return APIResponse<T>                                │
└────────────────────┬────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────┐
│              IN-MEMORY DATA STORES                       │
│                                                          │
│  • documentStore (Map)                                   │
│  • approversStore (Map)                                  │
│  • approvalLogsStore (Map)                               │
│  • attachmentsStore (Map)                                │
│  • customRolesStore (Map)                                │
│  • userRoleAssignmentsStore (Map)                        │
│                                                          │
│  [On server restart: data is cleared]                    │
│  [On production: Replace with database calls]            │
└─────────────────────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────┐
│                 RETURN RESPONSE                          │
│                                                          │
│  {                                                       │
│    success: boolean                                      │
│    message: string                                       │
│    data: T | null                                        │
│    status: number                                        │
│    statusText: string                                    │
│  }                                                       │
└────────────────────┬────────────────────────────────────┘
                     │
                     ▼
                   CLIENT
         (Component processes response)
```

---

## Server Actions Categorized

### Workflow Operations (`app/_actions/workflow.ts`)

**Document Lifecycle:**
- `createWorkflowDocument()` - New draft
- `submitDocument()` - Submit to approval
- `getDocument()` - Retrieve document
- `updateDocumentDraft()` - Edit draft
- `getDocumentsByCreator()` - Get user's documents

**Approval Workflow:**
- `approveDocument()` - Approve at current stage
- `rejectDocument()` - Return to draft
- `getApprovalLog()` - View approval history
- `getPendingApprovals()` - Get approvals for user's role

**Approver Management:**
- `assignApprover()` - Assign to step
- `reassignApprover()` - Change assignee
- `getDocumentApprovers()` - Get all approvers

**Attachments:**
- `uploadAttachment()` - Upload with role-based visibility
- `getAttachments()` - Get document attachments
- `deleteAttachment()` - Remove attachment

**Workflow Config & Reporting:**
- `getWorkflowSteps()` - Get approval chain
- `getDashboardStats()` - User statistics
- `getAuditLog()` - Full approval history

---

### RBAC Operations (`app/_actions/rbac.ts`)

**Role Management:**
- `getAllRoles()` - Retrieve all roles
- `getBuiltInRoles()` - Built-in roles only
- `getCustomRoles()` - Custom roles only
- `getRoleById()` - Single role
- `createRole()` - New custom role (admin only)
- `updateRole()` - Modify role (admin only)
- `deleteRole()` - Remove custom role (admin only)

**Permission Management:**
- `getAllPermissions()` - List all available permissions
- `updateRolePermissions()` - Replace all permissions
- `addRolePermission()` - Add single permission
- `removeRolePermission()` - Remove single permission

---

### User Management Operations (`app/_actions/user-management.ts`)

**User Retrieval:**
- `getAllUsers()` - All users
- `getUsersByRole()` - By built-in role
- `getUserById()` - Single user
- `searchUsers()` - Search by name/email/department

**Role Assignment:**
- `assignCustomRoleToUser()` - Assign custom role
- `removeCustomRoleFromUser()` - Remove custom role
- `getUserCustomRoles()` - User's custom roles
- `getUsersWithRole()` - Users with specific role
- `bulkAssignRolesToUser()` - Assign multiple roles
- `getAvailableRolesForUser()` - Not yet assigned

---

## Authentication & Authorization

### Authentication (NextAuth)
```typescript
// Every server action verifies session
const session = await auth()
if (!session?.user) return unauthorizedResponse()
```

### Authorization (RBAC)
```typescript
// Admin-only operations
const userRole = (session.user as any).role
if (!isAdmin(userRole)) {
  return forbiddenResponse()
}

// Permission checking
if (!hasPermission(userRole, 'approve_document')) {
  return forbiddenResponse()
}

// Role-specific operations
if (!canApproveAtStep(userRole, documentType, stepOrder)) {
  return forbiddenResponse()
}
```

---

## Workflow States

```
┌──────────────────────────────────────────────┐
│                    DRAFT                      │ ◄─── Created
├──────────────────────────────────────────────┤
│ (Can be edited)                              │
├──────────────────────────────────────────────┤
│             User clicks Submit                │
│                   │                           │
│                   ▼                           │
│              SUBMITTED                        │
│                   │                           │
│                   ▼                           │
│             IN_APPROVAL                       │ ◄─── Stage 1 active
│          (Approver can act)                   │
│                   │                           │
│         ┌─────────┴─────────┐                │
│         │                   │                │
│        ▼                   ▼                 │
│    APPROVED           REJECTED                │
│     (Final)      (Reset to DRAFT)            │
│                                              │
│    If more stages: Auto-advance to next     │
│    stage and assign next approver            │
└──────────────────────────────────────────────┘
```

---

## Default Approval Chains

### Purchase Order
```
Stage 1: DEPARTMENT_MANAGER (Required)
Stage 2: FINANCE_OFFICER (Required)
Stage 3: DIRECTOR (Required)
Stage 4: CFO (Optional)
```

### Payment Voucher
```
Stage 1: DEPARTMENT_MANAGER (Required)
Stage 2: FINANCE_OFFICER (Required)
Stage 3: CFO (Required)
```

### Requisition
```
Stage 1: DEPARTMENT_MANAGER (Required)
Stage 2: DIRECTOR (Required)
Stage 3: FINANCE_OFFICER (Required)
```

---

## Immutable Audit Trail

Every action is logged and never modified:

```typescript
ApprovalLogEntry {
  id: string              // Unique
  documentId: string      // Which document
  approver: User          // Who acted
  action: 'APPROVED' | 'REJECTED' | 'REASSIGNED' | 'COMMENTED'
  timestamp: Date         // When
  comments?: string       // What they said
  signature?: string      // Digital signature (future)
  ipAddress?: string      // Security audit (future)
}
```

**Characteristics:**
- Append-only (never deleted or modified)
- Immutable (stored as-is)
- Complete (all actions logged)
- Timestamps (preserves history)
- User attribution (who did what)

---

## Response Format (Standard)

All server actions return this format:

```typescript
interface APIResponse<T> {
  success: boolean      // Operation succeeded?
  message: string       // Human-readable message
  data: T | null        // Generic response data
  status: number        // HTTP-like status code
  statusText: string    // HTTP status text
}
```

### Success Response
```typescript
{
  success: true,
  message: "Document submitted for approval",
  data: {
    id: "doc-123",
    documentNumber: "PO-123456",
    status: "IN_APPROVAL",
    currentStage: 1
  },
  status: 200,
  statusText: "OK"
}
```

### Error Response
```typescript
{
  success: false,
  message: "You do not have permission to approve this document",
  data: null,
  status: 403,
  statusText: "FORBIDDEN"
}
```

---

## In-Memory Store Details

### documentStore
```typescript
Map<string, WorkflowDocument>
- Key: Document ID
- Value: Complete document with metadata
- Cleared on server restart
```

### approversStore
```typescript
Map<string, Approver[]>
- Key: Document ID
- Value: Array of approver assignments for that document
- Tracks approval stage progression
```

### approvalLogsStore
```typescript
Map<string, ApprovalLogEntry[]>
- Key: Document ID
- Value: Complete immutable approval history
- Never modified, only appended
```

### attachmentsStore
```typescript
Map<string, Attachment[]>
- Key: Document ID
- Value: Array of attachments with metadata
- Includes role-based visibility settings
```

### customRolesStore (in rbac.ts)
```typescript
Map<string, CustomRole>
- Key: Role ID
- Value: Role definition with permissions
- Initialized with built-in roles
```

### userRoleAssignmentsStore (in user-management.ts)
```typescript
Map<string, UserRoleAssignment>
- Key: `${userId}-${customRoleId}`
- Value: Assignment details with timestamps
- Links users to custom roles
```

---

## Database Integration (Future)

To migrate to real database:

1. **Replace in-memory stores** with database queries
2. **Keep server action signatures** the same
3. **Keep response format** the same
4. **Update error handling** to handle DB errors
5. **Add transactions** for complex operations

```typescript
// Before (In-memory)
const document = documentStore.get(documentId)

// After (Database)
const document = await db.WorkflowDocument.findUnique({
  where: { id: documentId }
})
```

---

## Testing

### Comprehensive Test Endpoint
```
GET /api/workflows/test
```

Tests:
- ✅ Role management (getAllRoles)
- ✅ User management (getAllUsers, getUsersByRole)
- ✅ Document creation (all 3 types)
- ✅ Document submission (2 documents)
- ✅ Approval workflow (approval + progression)
- ✅ Dashboard stats (statistics retrieval)
- ✅ Rejection workflow (rejection + reset)
- ✅ Audit logs (immutable trail verification)

---

## Key Features

✅ **Server Actions First** - All logic in server actions
✅ **Type Safety** - End-to-end TypeScript
✅ **Authentication** - NextAuth session required
✅ **Authorization** - RBAC with custom roles
✅ **Audit Trail** - Immutable approval history
✅ **Automatic Progression** - Workflow advances automatically
✅ **Flexible Approvers** - Can be reassigned
✅ **Attachment Management** - Role-based visibility
✅ **Mock Data** - Ready for frontend development
✅ **Database Agnostic** - Easy to migrate to real DB

---

## Next Steps

1. **Build React Components** - Use server actions in components
2. **Create Form Pages** - For document creation
3. **Build Approval Interface** - Approve/reject workflow
4. **Create Dashboard** - Show pending items
5. **Build Admin Pages** - Role and user management
6. **Connect to Backend** - When API is ready

All server actions are ready to use! 🎉
