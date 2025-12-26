# Workflow Backend Integration Plan

**Date**: 2025-12-26
**Status**: 📋 PLANNING PHASE
**Objective**: Move workflow system from frontend mock data/localStorage to backend-powered with RBAC alignment

---

## Executive Summary

The frontend has a comprehensive, well-implemented workflow system with multi-stage approvals, role-based assignments, audit trails, and digital signatures. The backend has RBAC infrastructure and approval routing logic. This plan bridges them by:

1. **Creating backend approval task APIs** - Replace mock approval store
2. **Implementing workflow state transitions** - Backend-driven state machine
3. **Building server actions** - Connect frontend to backend APIs
4. **Creating React Query hooks** - State management for approval workflows
5. **Refactoring components** - Use real data instead of mock/localStorage
6. **RBAC alignment** - Enforce permissions on both frontend and backend

---

## Current State Assessment

### Frontend ✅ (Well Implemented)
- **Workflow Types**: 5 document types (Requisition, PO, PV, GRN, Budget)
- **Approval Stages**: Multi-stage approval chains (1-4 stages each)
- **State Management**: localStorage persistence + in-memory store
- **RBAC**: Role enum defined (8 roles)
- **Features**: Signatures, audit trails, comments, reassignment, reversal
- **Components**: Full UI for approval flows, panels, history
- **Validation**: Comprehensive workflow validation
- **Status**: Production-ready code quality, but **mock data only**

### Backend ✅ (Partially Implemented)
- **RBAC**: Permission service with role-based access control
- **Approval Rules**: Dynamic routing based on amount/department/priority
- **State Machine**: Workflow state transitions defined
- **Approval Handlers**: Approve/reject/reassign endpoints exist
- **Models**: ApprovalTask, ApprovalRecord, AuditLog, Notification
- **Middleware**: Permission checks and tenant isolation
- **Routes**: Authorization-protected endpoints
- **Status**: **Stub endpoints** - Missing approval task management APIs

### Gap Analysis 🔴

| Feature | Frontend | Backend | Status |
|---------|----------|---------|--------|
| Multi-stage approvals | ✅ Designed | ✅ Rules engine | ⏳ Not connected |
| State transitions | ✅ Types defined | ✅ State machine | ⏳ Not connected |
| RBAC enforcement | ✅ UI respects roles | ✅ Middleware checks | ⏳ Not coordinated |
| Approval tasks | ❌ Mock data | ⏳ Stub endpoint | ❌ Missing |
| Approval history | ✅ In-memory store | ✅ Database model | ⏳ Not connected |
| Notifications | ✅ UI ready | ✅ Model exists | ❌ Not implemented |
| Digital signatures | ✅ Captures base64 | ✅ Database field | ⏳ Not connected |
| Audit logging | ✅ Frontend logs | ✅ Backend logs | ⏳ Not coordinated |

---

## Integration Architecture

```
Frontend (React)
    ↓
Server Actions (Node.js)
    ↓
authenticatedApiClient (axios)
    ↓
Backend Handlers (Go Fiber)
    ↓
Approval Service
    ├─ ApprovalRoutingService (route to approvers)
    ├─ WorkflowStateMachine (validate transitions)
    ├─ PermissionService (check RBAC)
    └─ NotificationService (send notifications)
    ↓
Database (PostgreSQL)
    ├─ ApprovalTask
    ├─ ApprovalHistory
    ├─ AuditLog
    └─ Notification
```

---

## Implementation Phases

### Phase 1: Backend Approval Task APIs (Week 1)

**Objective**: Create REST endpoints for approval task management

#### 1.1 Get Approval Tasks Endpoint
**Path**: `GET /api/v1/approvals`
**Query Parameters**:
- `status` (optional): pending, approved, rejected
- `document_type` (optional): requisition, budget, po, pv, grn
- `page` (default: 1)
- `limit` (default: 10)
- `assigned_to_me` (default: false)

**Response**:
```json
{
  "success": true,
  "data": [
    {
      "id": "task-id",
      "organizationId": "org-id",
      "documentId": "doc-id",
      "documentType": "requisition",
      "documentNumber": "REQ-1735243125-a1b2c3d4",
      "approverId": "user-id",
      "approverName": "John Doe",
      "approverRole": "DEPARTMENT_MANAGER",
      "status": "PENDING",
      "stage": 1,
      "priority": "HIGH",
      "createdAt": "2025-12-26T10:00:00Z",
      "dueAt": "2025-12-28T10:00:00Z",
      "overdue": false,
      "document": {
        "id": "doc-id",
        "title": "Office Supplies Q1",
        "amount": 5000,
        "requester": "Jane Smith",
        "status": "IN_REVIEW"
      }
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 10,
    "total": 45,
    "totalPages": 5
  }
}
```

**Permissions**: `RequirePermission("approval", "view")`

**Handler Implementation**:
```go
func GetApprovalTasks(c fiber.Ctx) error {
  // Extract query params
  // Get organization from context
  // Filter by status, document_type
  // If assigned_to_me=true, filter by current user
  // Fetch from database with document details
  // Return paginated results
}
```

#### 1.2 Get Single Approval Task
**Path**: `GET /api/v1/approvals/{id}`

**Response**:
```json
{
  "success": true,
  "data": {
    "id": "task-id",
    "documentId": "doc-id",
    "documentType": "requisition",
    "documentNumber": "REQ-1735243125-a1b2c3d4",
    "approverId": "user-id",
    "approverRole": "DEPARTMENT_MANAGER",
    "status": "PENDING",
    "stage": 1,
    "totalStages": 4,
    "priority": "HIGH",
    "createdAt": "2025-12-26T10:00:00Z",
    "dueAt": "2025-12-28T10:00:00Z",
    "document": {
      "id": "doc-id",
      "type": "requisition",
      "number": "REQ-1735243125-a1b2c3d4",
      "title": "Office Supplies Q1",
      "description": "...",
      "amount": 5000,
      "currency": "ZMW",
      "requester": {
        "id": "user-id",
        "name": "Jane Smith",
        "role": "REQUESTER"
      },
      "department": "Admin",
      "priority": "HIGH",
      "status": "IN_REVIEW",
      "items": [...],
      "approvalHistory": [
        {
          "stage": 1,
          "approver": "John Doe",
          "role": "DEPARTMENT_MANAGER",
          "status": "APPROVED",
          "comments": "Approved",
          "signature": "data:image/png;base64,...",
          "approvedAt": "2025-12-26T11:00:00Z"
        }
      ]
    },
    "validActions": ["APPROVE", "REJECT", "REASSIGN"],
    "requiredFields": {
      "comments": false,
      "signature": true,
      "remarksForRejection": false
    }
  }
}
```

#### 1.3 Approve Approval Task
**Path**: `POST /api/v1/approvals/{id}/approve`
**Method**: POST

**Request**:
```json
{
  "comments": "Approved, proceed to next stage",
  "signature": "data:image/png;base64,...",
  "stageNumber": 1
}
```

**Response**:
```json
{
  "success": true,
  "data": {
    "taskId": "task-id",
    "documentId": "doc-id",
    "status": "APPROVED",
    "nextStage": 2,
    "totalStages": 4,
    "documentStatus": "IN_REVIEW",
    "nextApprover": {
      "id": "user-id-2",
      "name": "Finance Officer",
      "role": "FINANCE_OFFICER"
    },
    "nextTask": {
      "id": "task-id-2",
      "createdAt": "2025-12-26T12:00:00Z",
      "dueAt": "2025-12-28T12:00:00Z"
    }
  }
}
```

**Logic**:
1. Verify user is assigned approver
2. Check permission: `require("approval", "approve")`
3. Validate signature is provided
4. Get approval configuration for document type
5. Add approval record to ApprovalHistory
6. Check if final stage → update document status to APPROVED
7. If not final stage → create next ApprovalTask
8. Create AuditLog entry
9. Send notification to next approver (if applicable)
10. Return response

#### 1.4 Reject Approval Task
**Path**: `POST /api/v1/approvals/{id}/reject`

**Request**:
```json
{
  "remarks": "Missing itemization details",
  "comments": "Please provide detailed breakdown of items",
  "signature": "data:image/png;base64,...",
  "returnTo": "REQUESTER"
}
```

**Response**:
```json
{
  "success": true,
  "data": {
    "taskId": "task-id",
    "documentId": "doc-id",
    "status": "REJECTED",
    "documentStatus": "DRAFT",
    "returnedTo": {
      "id": "original-requester-id",
      "name": "Jane Smith",
      "role": "REQUESTER"
    }
  }
}
```

#### 1.5 Reassign Approval Task
**Path**: `POST /api/v1/approvals/{id}/reassign`

**Request**:
```json
{
  "newApproverId": "user-id-3",
  "reason": "Original approver on leave",
  "reassignedBy": "user-id-manager"
}
```

**Response**:
```json
{
  "success": true,
  "data": {
    "taskId": "task-id",
    "previousApproverId": "user-id",
    "newApproverId": "user-id-3",
    "reason": "Original approver on leave",
    "reassignedAt": "2025-12-26T12:00:00Z"
  }
}
```

#### 1.6 Get Approval History
**Path**: `GET /api/v1/documents/{documentId}/approval-history`

**Response**:
```json
{
  "success": true,
  "data": [
    {
      "id": "history-id",
      "stage": 1,
      "stageName": "Department Manager Review",
      "approver": {
        "id": "user-id",
        "name": "John Doe",
        "role": "DEPARTMENT_MANAGER"
      },
      "status": "APPROVED",
      "action": "APPROVED",
      "comments": "Approved",
      "remarks": null,
      "signature": "data:image/png;base64,...",
      "approvedAt": "2025-12-26T11:00:00Z",
      "duration": 3600  // seconds to approve
    }
  ]
}
```

---

### Phase 2: Frontend Server Actions (Week 1)

**Objective**: Create server actions to call approval task APIs

**File**: `frontend/src/app/_actions/approval-workflow.ts`

```typescript
'use server';

import { authenticatedApiClient, handleError, successResponse } from './api-config';
import { APIResponse } from '@/types';
import {
  ApprovalTask,
  ApprovalTaskDetail,
  ApproveTaskRequest,
  RejectTaskRequest,
  ReassignTaskRequest,
  ApprovalHistory,
} from '@/types/workflow';

/**
 * Get all approval tasks for current user
 */
export async function getApprovalTasks(
  filters?: {
    status?: 'PENDING' | 'APPROVED' | 'REJECTED';
    documentType?: string;
    assignedToMe?: boolean;
  },
  page: number = 1,
  limit: number = 10
): Promise<APIResponse<ApprovalTask[]>> {
  const url = `/api/v1/approvals`;
  const params = new URLSearchParams();
  params.set('page', page.toString());
  params.set('limit', limit.toString());

  if (filters?.status) params.set('status', filters.status);
  if (filters?.documentType) params.set('document_type', filters.documentType);
  if (filters?.assignedToMe) params.set('assigned_to_me', 'true');

  try {
    const response = await authenticatedApiClient({
      method: 'GET',
      url: `${url}?${params.toString()}`,
    });
    return successResponse(response.data?.data, 'Approval tasks retrieved');
  } catch (error: any) {
    return handleError(error, 'GET', url);
  }
}

/**
 * Get single approval task detail
 */
export async function getApprovalTaskDetail(
  taskId: string
): Promise<APIResponse<ApprovalTaskDetail>> {
  const url = `/api/v1/approvals/${taskId}`;
  try {
    const response = await authenticatedApiClient({
      method: 'GET',
      url,
    });
    return successResponse(response.data?.data, 'Approval task retrieved');
  } catch (error: any) {
    return handleError(error, 'GET', url);
  }
}

/**
 * Approve an approval task
 */
export async function approveApprovalTask(
  data: ApproveTaskRequest
): Promise<APIResponse<any>> {
  const url = `/api/v1/approvals/${data.taskId}/approve`;
  try {
    const response = await authenticatedApiClient({
      method: 'POST',
      url,
      data: {
        comments: data.comments,
        signature: data.signature,
        stageNumber: data.stageNumber,
      },
    });
    return successResponse(response.data?.data, 'Task approved successfully');
  } catch (error: any) {
    return handleError(error, 'POST', url);
  }
}

/**
 * Reject an approval task
 */
export async function rejectApprovalTask(
  data: RejectTaskRequest
): Promise<APIResponse<any>> {
  const url = `/api/v1/approvals/${data.taskId}/reject`;
  try {
    const response = await authenticatedApiClient({
      method: 'POST',
      url,
      data: {
        remarks: data.remarks,
        comments: data.comments,
        signature: data.signature,
        returnTo: data.returnTo,
      },
    });
    return successResponse(response.data?.data, 'Task rejected successfully');
  } catch (error: any) {
    return handleError(error, 'POST', url);
  }
}

/**
 * Reassign an approval task
 */
export async function reassignApprovalTask(
  data: ReassignTaskRequest
): Promise<APIResponse<any>> {
  const url = `/api/v1/approvals/${data.taskId}/reassign`;
  try {
    const response = await authenticatedApiClient({
      method: 'POST',
      url,
      data: {
        newApproverId: data.newApproverId,
        reason: data.reason,
      },
    });
    return successResponse(response.data?.data, 'Task reassigned successfully');
  } catch (error: any) {
    return handleError(error, 'POST', url);
  }
}

/**
 * Get approval history for a document
 */
export async function getApprovalHistory(
  documentId: string
): Promise<APIResponse<ApprovalHistory[]>> {
  const url = `/api/v1/documents/${documentId}/approval-history`;
  try {
    const response = await authenticatedApiClient({
      method: 'GET',
      url,
    });
    return successResponse(response.data?.data, 'Approval history retrieved');
  } catch (error: any) {
    return handleError(error, 'GET', url);
  }
}
```

---

### Phase 3: React Query Hooks (Week 1)

**Objective**: Create custom hooks for approval workflow state management

**File**: `frontend/src/hooks/use-approval-workflow.ts`

```typescript
'use client';

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { QUERY_KEYS } from '@/lib/constants';
import { toast } from 'sonner';
import {
  getApprovalTasks,
  getApprovalTaskDetail,
  approveApprovalTask,
  rejectApprovalTask,
  reassignApprovalTask,
  getApprovalHistory,
} from '@/app/_actions/approval-workflow';
import { ApprovalTask, ApprovalTaskDetail } from '@/types/workflow';

/**
 * Fetch approval tasks with filters and pagination
 */
export const useApprovalTasks = (
  filters?: { status?: string; documentType?: string; assignedToMe?: boolean },
  page: number = 1,
  limit: number = 10
) =>
  useQuery({
    queryKey: [QUERY_KEYS.APPROVALS.ALL, filters, page, limit],
    queryFn: async () => {
      const response = await getApprovalTasks(filters, page, limit);
      if (!response.success) throw new Error(response.message);
      return response.data;
    },
    staleTime: 5 * 60 * 1000, // 5 minutes
  });

/**
 * Fetch single approval task detail
 */
export const useApprovalTaskDetail = (taskId: string) =>
  useQuery({
    queryKey: [QUERY_KEYS.APPROVALS.BY_ID, taskId],
    queryFn: async () => {
      const response = await getApprovalTaskDetail(taskId);
      if (!response.success) throw new Error(response.message);
      return response.data;
    },
    enabled: !!taskId,
    staleTime: 5 * 60 * 1000,
  });

/**
 * Approve approval task mutation
 */
export const useApproveTask = (taskId: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data) => {
      const response = await approveApprovalTask({
        taskId,
        ...data,
      });
      if (!response.success) throw new Error(response.message);
      return response;
    },
    onSuccess: (response) => {
      toast.success('Task approved successfully');

      // Invalidate relevant queries
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.APPROVALS.ALL],
      });
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.APPROVALS.BY_ID, taskId],
      });
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.APPROVALS.PENDING_COUNT],
      });

      // Invalidate related document
      if (response.data?.documentId) {
        queryClient.invalidateQueries({
          queryKey: [QUERY_KEYS.DOCUMENTS.BY_ID, response.data.documentId],
        });
      }
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to approve task');
    },
  });
};

/**
 * Reject approval task mutation
 */
export const useRejectTask = (taskId: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data) => {
      const response = await rejectApprovalTask({
        taskId,
        ...data,
      });
      if (!response.success) throw new Error(response.message);
      return response;
    },
    onSuccess: (response) => {
      toast.success('Task rejected successfully');

      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.APPROVALS.ALL],
      });
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.APPROVALS.BY_ID, taskId],
      });

      if (response.data?.documentId) {
        queryClient.invalidateQueries({
          queryKey: [QUERY_KEYS.DOCUMENTS.BY_ID, response.data.documentId],
        });
      }
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to reject task');
    },
  });
};

/**
 * Reassign approval task mutation
 */
export const useReassignTask = (taskId: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data) => {
      const response = await reassignApprovalTask({
        taskId,
        ...data,
      });
      if (!response.success) throw new Error(response.message);
      return response;
    },
    onSuccess: () => {
      toast.success('Task reassigned successfully');

      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.APPROVALS.ALL],
      });
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.APPROVALS.BY_ID, taskId],
      });
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to reassign task');
    },
  });
};

/**
 * Fetch approval history for document
 */
export const useApprovalHistory = (documentId: string) =>
  useQuery({
    queryKey: [QUERY_KEYS.APPROVALS.HISTORY, documentId],
    queryFn: async () => {
      const response = await getApprovalHistory(documentId);
      if (!response.success) throw new Error(response.message);
      return response.data;
    },
    enabled: !!documentId,
    staleTime: 5 * 60 * 1000,
  });

/**
 * Get pending approval count
 */
export const usePendingApprovalCount = () =>
  useQuery({
    queryKey: [QUERY_KEYS.APPROVALS.PENDING_COUNT],
    queryFn: async () => {
      const response = await getApprovalTasks(
        { status: 'PENDING', assignedToMe: true },
        1,
        1
      );
      if (!response.success) throw new Error(response.message);
      return response.data.length; // Would need pagination info from API
    },
    staleTime: 2 * 60 * 1000, // 2 minutes
  });
```

---

### Phase 4: Update Type Definitions

**Objective**: Ensure types align between frontend and backend

**File**: `frontend/src/types/workflow.ts` (Update)

Add new types:
```typescript
export interface ApprovalTask {
  id: string;
  organizationId: string;
  documentId: string;
  documentType: string;
  documentNumber: string;
  approverId: string;
  approverName: string;
  approverRole: string;
  status: 'PENDING' | 'APPROVED' | 'REJECTED';
  stage: number;
  priority: string;
  createdAt: Date;
  dueAt: Date;
  overdue: boolean;
  document?: {
    id: string;
    title: string;
    amount: number;
    requester: string;
    status: string;
  };
}

export interface ApprovalTaskDetail extends ApprovalTask {
  totalStages: number;
  document: {
    // full document details
  };
  validActions: ('APPROVE' | 'REJECT' | 'REASSIGN')[];
  requiredFields: {
    comments: boolean;
    signature: boolean;
    remarksForRejection: boolean;
  };
}

export interface ApproveTaskRequest {
  taskId: string;
  comments: string;
  signature: string;
  stageNumber: number;
}

export interface RejectTaskRequest {
  taskId: string;
  remarks: string;
  comments: string;
  signature: string;
  returnTo?: string;
}

export interface ReassignTaskRequest {
  taskId: string;
  newApproverId: string;
  reason: string;
}

export interface ApprovalHistory {
  id: string;
  stage: number;
  stageName: string;
  approver: {
    id: string;
    name: string;
    role: string;
  };
  status: string;
  action: string;
  comments?: string;
  remarks?: string;
  signature?: string;
  approvedAt: Date;
  duration?: number;
}
```

---

### Phase 5: Refactor Components

**Objective**: Update UI components to use real backend data

Components to update:
- `approval-flow-display.tsx` - Show real workflow stages
- `approval-action-panel.tsx` - Call backend endpoints
- `approval-history.tsx` - Display real approval records
- `approval-modal.tsx` - Submit to backend
- Requisition, PO, PV, Budget approval panels

**Example Component Update**:
```typescript
'use client';

import { useApprovalTaskDetail, useApproveTask } from '@/hooks/use-approval-workflow';
import { ApprovalSignatureModal } from './approval-signature-modal';

export function ApprovalActionPanel({ taskId }: { taskId: string }) {
  const { data: task, isLoading } = useApprovalTaskDetail(taskId);
  const approveMutation = useApproveTask(taskId);

  const handleApprove = async (data: ApproveTaskRequest) => {
    await approveMutation.mutateAsync(data);
    // Modal closes automatically on success via toast
  };

  if (isLoading) return <div>Loading...</div>;
  if (!task) return <div>Task not found</div>;

  return (
    <div className="approval-panel">
      <h2>{task.document.title}</h2>
      <ApprovalSignatureModal
        onApprove={handleApprove}
        isLoading={approveMutation.isPending}
      />
    </div>
  );
}
```

---

## Implementation Steps Summary

### Step 1: Backend (Parallel)
- [ ] Create `GetApprovalTasks` handler
- [ ] Create `GetApprovalTask` handler
- [ ] Create `ApproveTask` handler
- [ ] Create `RejectTask` handler
- [ ] Create `ReassignTask` handler
- [ ] Create `GetApprovalHistory` handler
- [ ] Add routes and permissions
- [ ] Test with Postman
- [ ] Commit and document

### Step 2: Frontend Server Actions (Parallel)
- [ ] Create `approval-workflow.ts` server actions
- [ ] Import types from backend responses
- [ ] Test locally with mock data
- [ ] Commit

### Step 3: Frontend Hooks & Types
- [ ] Update `workflow.ts` types
- [ ] Create `use-approval-workflow.ts` hooks
- [ ] Add QUERY_KEYS for approvals
- [ ] Test hook logic

### Step 4: Component Updates
- [ ] Update `approval-action-panel.tsx`
- [ ] Update `approval-flow-display.tsx`
- [ ] Update `approval-history.tsx`
- [ ] Update document-specific approval panels
- [ ] Remove references to mock stores
- [ ] Update imports to use hooks

### Step 5: Integration Testing
- [ ] E2E test approval flow
- [ ] Test RBAC enforcement
- [ ] Test state transitions
- [ ] Test notifications
- [ ] Test audit logging

### Step 6: Documentation
- [ ] Update API documentation
- [ ] Create workflow integration guide
- [ ] Document RBAC matrix
- [ ] Create approval flow diagrams

---

## Database Queries Needed

### Query 1: Get approval tasks with document details
```sql
SELECT
  at.id, at.organization_id, at.document_id, at.document_type,
  at.approver_id, at.status, at.stage, at.created_at,
  u.name as approver_name, u.role as approver_role,
  r.title, r.total_amount, r.priority
FROM approval_tasks at
JOIN users u ON at.approver_id = u.id
LEFT JOIN requisitions r ON at.document_id = r.id AND at.document_type = 'requisition'
WHERE at.organization_id = ?
  AND at.approver_id = ?
  AND at.status = ?
ORDER BY at.created_at DESC
LIMIT ? OFFSET ?;
```

### Query 2: Get approval history with approver details
```sql
SELECT
  ah.*, u.name as approver_name, u.role as approver_role
FROM approval_history ah
JOIN users u ON ah.approver_id = u.id
WHERE ah.document_id = ?
ORDER BY ah.stage ASC;
```

---

## RBAC Integration Points

| Component | Permission Check | Current | Needed |
|-----------|------------------|---------|--------|
| View tasks | `approval.view` | ❌ | ✅ |
| Approve | `approval.approve` | ⚠️ | ✅ |
| Reject | `approval.reject` | ⚠️ | ✅ |
| Reassign | `approval.reassign` | ❌ | ✅ |
| View history | `approval.view_history` | ❌ | ✅ |
| View other's tasks | `approval.view_all` | ❌ | ✅ |

---

## Security Considerations

1. **User Verification**: Ensure approver_id matches current user
2. **Permission Checks**: Verify role has approval permission
3. **Organization Scoping**: All queries filtered by organization_id
4. **Signature Validation**: Store signature with timestamp
5. **Audit Logging**: Log all approval actions
6. **Rate Limiting**: Consider rate limiting approval endpoints
7. **Signature Verification**: Validate signature format (base64 image)

---

## Testing Strategy

### Unit Tests
- Permission checks
- State transition validation
- Approval routing logic

### Integration Tests
- Full approval workflow (draft → approved)
- Rejection flow (returns to requester)
- Reassignment flow
- Multi-stage approvals

### E2E Tests
- User A creates requisition
- User B (approver) approves
- User C (finance) approves
- Verify notifications sent
- Check audit log created

### RBAC Tests
- Non-approvers cannot approve
- Finance officer sees their tasks
- Admin sees all tasks
- Custom roles work correctly

---

## Dependencies

**Frontend**:
- React Query (already have)
- axios (already have)
- TypeScript (already have)
- sonner (already have)

**Backend**:
- GORM (already have)
- Go Fiber (already have)
- PostgreSQL (already have)

**No new external dependencies needed!**

---

## Timeline Estimate

| Phase | Task | Hours | Days |
|-------|------|-------|------|
| 1 | Backend APIs | 8 | 1 |
| 2 | Server Actions | 4 | 0.5 |
| 3 | Hooks & Types | 4 | 0.5 |
| 4 | Components | 8 | 1 |
| 5 | Integration Testing | 4 | 0.5 |
| 6 | Documentation | 4 | 0.5 |
| **Total** | | **32 hours** | **4 days** |

---

## Success Criteria

- ✅ Approval tasks retrieved from backend
- ✅ Approve/reject actions call backend APIs
- ✅ Approval history displayed from database
- ✅ RBAC enforced on both frontend and backend
- ✅ State transitions validated by backend
- ✅ No mock data or localStorage used
- ✅ Digital signatures captured and stored
- ✅ Notifications sent on state changes
- ✅ Audit logs created for all actions
- ✅ 100% of approval workflows tested

---

## Next Steps

1. Review this plan with team
2. Create GitHub issues for each phase
3. Assign team members
4. Begin Phase 1 (Backend APIs) immediately
5. Phases 2-4 can run in parallel once APIs are created

---

**Created**: 2025-12-26
**Status**: 📋 PLANNING COMPLETE - READY FOR IMPLEMENTATION
**Owner**: Development Team
**Reviewed By**: (Pending)

