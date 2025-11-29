# Workflow System - Quick Start Guide

## What's Been Created

A complete, mocked workflow management system ready for frontend development:

- ✅ Workflow document types (Purchase Order, Payment Voucher, Requisition)
- ✅ Multi-stage approval workflows with automatic progression
- ✅ Role-Based Access Control (RBAC) with custom roles
- ✅ User management and role assignment
- ✅ Immutable audit trails
- ✅ Attachment management with role-based visibility
- ✅ All data stored in-memory (ready for database integration)

---

## File Locations

### Core Types & Configuration
- **Types**: `src/types/workflow.ts` - All TypeScript types
- **RBAC**: `src/lib/rbac.ts` - Role/permission management
- **Mock Data**: `src/lib/mock-data.ts` - Data factories

### Server Actions (Mocked)
- **Workflows**: `src/app/_actions/workflow.ts` - Document operations
- **RBAC**: `src/app/_actions/rbac.ts` - Role management
- **Users**: `src/app/_actions/user-management.ts` - User management

### Testing
- **Test API**: `src/app/api/workflows/test/route.ts` - Comprehensive test endpoint

---

## Quick Examples

### 1. Create a Purchase Order Draft

```typescript
import { createWorkflowDocument } from '@/app/_actions/workflow'

const result = await createWorkflowDocument('PURCHASE_ORDER', {
  vendorName: 'Acme Supplies',
  vendorId: 'V-001',
  items: [
    {
      id: 'item-1',
      description: 'Office Chairs',
      quantity: 10,
      unitCost: 500,
      totalCost: 5000
    }
  ],
  totalAmount: 5000,
  currency: 'ZMW',
  deliveryDate: new Date('2024-12-31'),
  specialInstructions: 'Urgent'
})

if (result.success) {
  console.log('Created:', result.data.documentNumber) // PO-123456-789
}
```

### 2. Submit for Approval

```typescript
import { submitDocument } from '@/app/_actions/workflow'

const result = await submitDocument(documentId)

if (result.success) {
  console.log('Status:', result.data.status) // 'IN_APPROVAL'
  console.log('Current Stage:', result.data.currentStage) // 1
}
```

### 3. Approve a Document

```typescript
import { approveDocument } from '@/app/_actions/workflow'

const result = await approveDocument(
  documentId,
  'Approved - All items within budget'
)

if (result.success) {
  console.log('Approved by:', result.data.approver.name)
  console.log('Timestamp:', result.data.timestamp)
}
```

### 4. Create a Custom Role

```typescript
import { createRole } from '@/app/_actions/rbac'

const result = await createRole(
  'Senior Approver',
  'Users with high-level approval authority',
  [
    'approve_document',
    'reject_document',
    'reassign_approver',
    'view_audit_log'
  ]
)

if (result.success) {
  console.log('Role created:', result.data.id)
}
```

### 5. Assign Role to User

```typescript
import { assignCustomRoleToUser } from '@/app/_actions/user-management'

const result = await assignCustomRoleToUser(
  userId,        // User ID
  customRoleId   // Custom role ID from step 4
)

if (result.success) {
  console.log('Role assigned to user')
}
```

### 6. Get All Approvals Pending for a Role

```typescript
import { getPendingApprovals } from '@/app/_actions/workflow'

const result = await getPendingApprovals('FINANCE_OFFICER')

if (result.success) {
  console.log(`${result.data.length} documents awaiting finance approval`)
  result.data.forEach(doc => {
    console.log(`- ${doc.documentNumber}: ${doc.status}`)
  })
}
```

### 7. View Audit Log

```typescript
import { getAuditLog } from '@/app/_actions/workflow'

const result = await getAuditLog(documentId)

if (result.success) {
  result.data.forEach(entry => {
    console.log(`${entry.action} by ${entry.approver.name} on ${entry.timestamp}`)
    if (entry.comments) {
      console.log(`  Comments: ${entry.comments}`)
    }
  })
}
```

### 8. Get Dashboard Stats

```typescript
import { getDashboardStats } from '@/app/_actions/workflow'

const result = await getDashboardStats(userId)

if (result.success) {
  console.log('Documents created:', result.data.createdDocuments)
  console.log('Pending approvals:', result.data.pendingApprovals)
  console.log('Approved:', result.data.approvedDocuments)
  console.log('Rejected:', result.data.rejectedDocuments)
}
```

---

## Workflow Approval Chains

### Purchase Order
```
1. Department Manager (required)
2. Finance Officer (required)
3. Director (required)
4. CFO (optional)
```

### Payment Voucher
```
1. Department Manager (required)
2. Finance Officer (required)
3. CFO (required)
```

### Requisition
```
1. Department Manager (required)
2. Director (required)
3. Finance Officer (required)
```

---

## Mock Users Available

### Requesters
- John Mwale (user-req-1) - Operations
- Sarah Banda (user-req-2) - HR

### Department Managers
- James Chileshe (user-dm-1) - Operations
- Maria Chiyanda (user-dm-2) - HR

### Finance Officers
- Paul Nkosi (user-fo-1) - Finance
- Grace Mvula (user-fo-2) - Finance

### Directors
- David Moyo (user-dir-1) - Operations

### CFO
- Catherine Phiri (user-cfo-1) - Finance

### Compliance Officer
- Victor Zulu (user-co-1) - Legal

### Admin
- Admin User (user-admin-1) - IT

---

## Default Built-in Roles

### REQUESTER
Can create, edit, and submit documents
```
- view_draft
- edit_draft
- submit_document
- view_attachments
- add_attachments
- view_comments
- add_comments
```

### DEPARTMENT_MANAGER
Can approve/reject documents at department level
```
- view_draft
- approve_document
- reject_document
- view_attachments
- view_comments
- add_comments
- view_audit_log
```

### FINANCE_OFFICER
Can verify documents and reassign approvers
```
- approve_document
- reject_document
- reassign_approver
- view_attachments
- view_comments
- add_comments
- view_audit_log
```

### DIRECTOR
Can approve high-value documents
```
- approve_document
- reject_document
- view_attachments
- view_comments
- add_comments
- view_audit_log
```

### CFO
Executive approval authority
```
- approve_document
- reject_document
- view_attachments
- view_comments
- add_comments
- view_audit_log
```

### COMPLIANCE_OFFICER
View-only audit trail access
```
- view_audit_log
- view_attachments
- view_comments
```

### ADMIN
Full system access
```
All permissions available
```

---

## All Available Permissions

| Permission | Description |
|-----------|-------------|
| `view_draft` | View draft documents |
| `edit_draft` | Edit draft documents |
| `submit_document` | Submit documents for approval |
| `approve_document` | Approve documents |
| `reject_document` | Reject documents |
| `reassign_approver` | Reassign approvers |
| `view_attachments` | View attachments |
| `add_attachments` | Upload attachments |
| `view_comments` | View approval comments |
| `add_comments` | Add approval comments |
| `view_audit_log` | View audit logs |
| `manage_approvers` | Manage approvers |
| `manage_workflows` | Manage workflow configuration |

---

## Test Everything at Once

```bash
curl http://localhost:3001/api/workflows/test
```

**What it tests:**
1. Role management (retrieves all roles)
2. User management (retrieves users)
3. Document creation (creates 3 different document types)
4. Document submission (submits 2 documents)
5. Approval workflow (approves and checks progression)
6. Rejection workflow (rejects document and verifies reset)
7. Dashboard stats (retrieves statistics)
8. Audit logs (verifies immutable trail)

---

## Response Format

All server actions return consistent responses:

```typescript
{
  success: boolean          // true if successful
  message: string          // Human-readable message
  data: T | null          // Response data (generic)
  status: number          // HTTP status code
  statusText: string      // HTTP status text
}
```

### Success Example
```json
{
  "success": true,
  "message": "Document submitted for approval",
  "data": {
    "id": "doc-123",
    "documentNumber": "PO-123456-789",
    "status": "IN_APPROVAL",
    "currentStage": 1
  },
  "status": 200,
  "statusText": "OK"
}
```

### Error Example
```json
{
  "success": false,
  "message": "Document is not pending approval",
  "data": null,
  "status": 400,
  "statusText": "BAD REQUEST"
}
```

---

## Key Features

✅ **Automatic Workflow Progression**
- Submit document → Auto-assigns first approver
- Approve at stage N → Auto-advances to stage N+1
- Last stage → Document marked APPROVED

✅ **Immutable Audit Trail**
- Every approval action logged
- Timestamps, approver info, comments preserved
- Cannot be modified (append-only)

✅ **Role-Based Access**
- Built-in roles for common scenarios
- Custom roles can be created by admins
- Permissions assigned per role
- Users assigned multiple roles if needed

✅ **Flexible Approvers**
- Approvers can be reassigned before they act
- Role-based automatic assignment
- Manual override possible

✅ **Attachment Management**
- Role-based visibility
- Audit trail includes who uploaded what
- Metadata stored with document

---

## Next Steps

1. **Create UI Components** - Use these server actions in React components
2. **Build Forms** - Forms for Purchase Order, Payment Voucher, Requisition
3. **Create Dashboard** - Show pending approvals, created documents
4. **Build Approval Interface** - Approve/reject workflow UI
5. **Replace Mock Backend** - When ready, connect to real API

---

## Documentation

- **Full API Docs**: `WORKFLOW_MOCK_API_DOCUMENTATION.md`
- **Type Definitions**: `src/types/workflow.ts`
- **Mock Data**: `src/lib/mock-data.ts`

---

## Quick Troubleshooting

### "Document not found"
- Ensure document ID is correct
- Check that document was created in current session
- Data resets when server restarts

### "User not authorized"
- Check that session.user exists
- For role operations, ensure user has ADMIN role

### "Invalid permission"
- Use permissions from `ALL_PERMISSIONS` in `rbac.ts`
- Check spelling (underscores matter)

### Data disappeared
- In-memory store clears on server restart
- For persistence, connect to database

---

## Support Files

- `/src/types/workflow.ts` - All TypeScript types
- `/src/lib/rbac.ts` - RBAC system with custom role functions
- `/src/lib/mock-data.ts` - Mock data factories
- `/src/app/_actions/workflow.ts` - Workflow operations
- `/src/app/_actions/rbac.ts` - Role management operations
- `/src/app/_actions/user-management.ts` - User management operations
- `/src/app/api/workflows/test/route.ts` - Test endpoint
