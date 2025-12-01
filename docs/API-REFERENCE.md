# API Reference

## Server Actions

### Approval Actions

#### getApprovalTasks()
```typescript
getApprovalTasks(filters?: { status?: 'pending' | 'approved' | 'rejected' })
→ { success: boolean, data: { tasks: ApprovalTask[] } }
```
Fetch list of approval tasks with optional filtering.

#### getApprovalTaskDetail()
```typescript
getApprovalTaskDetail(taskId: string)
→ { success: boolean, data: ApprovalTaskDetail }
```
Fetch single task with full details and history.

#### approveTask()
```typescript
approveTask({
  assignmentId: string
  stageNumber: number
  approvingUserId: string
  signature: string
  comments?: string
})
→ { success: boolean, data: { approved: number } }
```
Approve a task with signature and comments.

#### rejectTask()
```typescript
rejectTask({
  assignmentId: string
  rejectingUserId: string
  signature: string
  remarks: string
})
→ { success: boolean, data: { rejected: number } }
```
Reject a task with required reason.

#### reassignTask()
```typescript
reassignTask({
  assignmentId: string
  reassignedBy: string
  newApproverId: string
  newApproverName: string
  reason?: string
})
→ { success: boolean, data: { reassigned: number } }
```
Reassign task to different approver.

### Bulk Operations

#### bulkApproveTasks()
```typescript
bulkApproveTasks({
  taskIds: string[]
  remarks?: string
  userId: string
})
→ { success: boolean, data: { approved: number } }
```
Approve multiple tasks at once.

#### bulkRejectTasks()
```typescript
bulkRejectTasks({
  taskIds: string[]
  remarks: string
  userId: string
})
→ { success: boolean, data: { rejected: number } }
```
Reject multiple tasks with required reason.

#### bulkReassignTasks()
```typescript
bulkReassignTasks({
  taskIds: string[]
  newApproverId: string
  newApproverName: string
  reason?: string
  userId: string
})
→ { success: boolean, data: { reassigned: number } }
```
Reassign multiple tasks to same approver.

#### getAnalyticsMetrics()
```typescript
getAnalyticsMetrics(userId: string)
→ { success: boolean, data: MetricsData }
```
Get dashboard metrics (pending, approved, rejected, avg time, SLA).

#### getWorkflowTrends()
```typescript
getWorkflowTrends(userId?: string)
→ { success: boolean, data: TimelineData[] }
```
Get 7-day approval trends.

#### getBottleneckAnalysis()
```typescript
getBottleneckAnalysis()
→ { success: boolean, data: StageMetricsData[] }
```
Get stage performance and bottleneck info.

## React Query Hooks

### Queries

#### useGetApprovalTasks()
```typescript
useGetApprovalTasks(options?: { status?: string })
→ { data: ApprovalTask[], isLoading, error, refetch }
```
Hook to fetch approval tasks. Auto-refreshes every 30s.

#### useGetApprovalTaskDetail()
```typescript
useGetApprovalTaskDetail(taskId: string)
→ { data: ApprovalTaskDetail, isLoading, error }
```
Hook to fetch single task details.

#### useGetApprovalStats()
```typescript
useGetApprovalStats()
→ { data: ApprovalStats, isLoading, error }
```
Hook to fetch statistics (counts, percentages).

### Mutations

#### useApproveTaskMutation()
```typescript
useApproveTaskMutation()
→ { mutate, mutateAsync, isPending, error }
```
Mutation to approve a task.

#### useRejectTaskMutation()
```typescript
useRejectTaskMutation()
→ { mutate, mutateAsync, isPending, error }
```
Mutation to reject a task.

#### useReassignTaskMutation()
```typescript
useReassignTaskMutation()
→ { mutate, mutateAsync, isPending, error }
```
Mutation to reassign a task.

#### useApprovalActions()
```typescript
useApprovalActions()
→ { approveMutation, rejectMutation, reassignMutation }
```
Combined hook with all three mutations.

## Type Definitions

### ApprovalTask
```typescript
interface ApprovalTask {
  id: string
  entityId: string
  entityType: 'REQUISITION' | 'BUDGET' | 'PO' | 'PV' | 'GRN'
  entityNumber: string
  status: 'pending' | 'approved' | 'rejected'
  stageName: string
  stageIndex: number
  importance: 'LOW' | 'MEDIUM' | 'HIGH'
  approverName: string
  approverUserId: string
  createdAt: Date
  dueDate: Date
  workflowId: string
  workflowName: string
}
```

### ApprovalTaskDetail
```typescript
interface ApprovalTaskDetail {
  task: ApprovalTask
  workflow: Workflow
  entity: any  // Document details
  relatedApprovals: ApprovalHistory[]
  comments: string[]
}
```

### Workflow
```typescript
interface Workflow {
  id: string
  name: string
  description: string
  stages: WorkflowStage[]
  entityType: string
  status: 'published' | 'draft'
}
```

### WorkflowStage
```typescript
interface WorkflowStage {
  name: string
  order: number
  approvers: string[]
  allowReassign: boolean
}
```

### ApprovalHistory
```typescript
interface ApprovalHistory {
  id: string
  taskId: string
  action: 'approved' | 'rejected' | 'reassigned'
  approverName: string
  timestamp: Date
  signature?: string
  remarks?: string
}
```

## Usage Examples

### Approve a Task
```typescript
const { mutate } = useApproveTaskMutation()

mutate({
  assignmentId: 'task-123',
  stageNumber: 1,
  approvingUserId: 'user-456',
  signature: 'base64-signature',
  comments: 'Looks good'
})
```

### Get Approval Tasks
```typescript
const { data, isLoading } = useGetApprovalTasks({
  status: 'pending'
})

if (isLoading) return <Spinner />
return tasks.map(t => <TaskCard key={t.id} task={t} />)
```

### Bulk Approve
```typescript
const result = await bulkApproveTasks({
  taskIds: ['task-1', 'task-2', 'task-3'],
  remarks: 'All approved',
  userId: 'user-123'
})
```

## Error Handling

All operations return:
```typescript
{
  success: boolean
  data?: any
  error?: string
}
```

Handle errors:
```typescript
const { mutate, isError, error } = useApproveTaskMutation()

mutate(data, {
  onError: (error) => {
    console.error(error)
    toast.error(error.message)
  }
})
```

## Rate Limiting

Currently: No rate limiting (Phase 12: Add rate limiting)

## Performance

- Bulk operations optimized for 10-100 items
- Analytics queries cached for 30 seconds
- Tasks list auto-refreshes every 30 seconds
- No N+1 query problems

## REST API Routes (Phase 12)

When Phase 12 adds real backend, these routes will be created:

### Approval Routes

#### GET /api/approvals/tasks
List all approval tasks with filters.

**Query Parameters**:
- `status` (optional): 'pending' | 'approved' | 'rejected'
- `page` (optional): Page number (default: 1)
- `limit` (optional): Items per page (default: 20)

**Response**:
```json
{
  "success": true,
  "data": {
    "tasks": [
      {
        "id": "task-123",
        "entityId": "req-001",
        "entityType": "REQUISITION",
        "entityNumber": "REQ-2024-001",
        "status": "pending",
        "stageName": "Finance Officer Review",
        "stageIndex": 1,
        "importance": "HIGH",
        "approverName": "John Smith",
        "approverUserId": "user-456",
        "createdAt": "2024-12-01T10:00:00Z",
        "dueDate": "2024-12-05T10:00:00Z",
        "workflowId": "workflow-req-001",
        "workflowName": "3-Stage Requisition"
      }
    ],
    "total": 45,
    "page": 1,
    "pageCount": 3
  }
}
```

#### GET /api/approvals/tasks/:id
Get single task with full details.

**Response**:
```json
{
  "success": true,
  "data": {
    "task": { /* task object */ },
    "workflow": {
      "id": "wf-123",
      "name": "3-Stage Requisition",
      "stages": [
        { "name": "Manager Review", "order": 0 },
        { "name": "Finance Officer Review", "order": 1 },
        { "name": "CFO Approval", "order": 2 }
      ]
    },
    "entity": {
      "description": "Office supplies",
      "amount": 2500,
      "requester": "Jane Doe"
    },
    "approvalHistory": [
      {
        "action": "submitted",
        "actor": "user-789",
        "timestamp": "2024-12-01T10:00:00Z"
      }
    ]
  }
}
```

#### POST /api/approvals/tasks/:id/approve
Approve a task.

**Request Body**:
```json
{
  "assignmentId": "task-123",
  "stageNumber": 1,
  "approvingUserId": "user-456",
  "signature": "data:image/png;base64,iVBORw0KGg...",
  "comments": "Looks good, approved"
}
```

**Response**:
```json
{
  "success": true,
  "data": {
    "taskId": "task-123",
    "action": "approved",
    "newStatus": "pending",
    "nextStage": "CFO Approval",
    "timestamp": "2024-12-01T11:00:00Z"
  }
}
```

#### POST /api/approvals/tasks/:id/reject
Reject a task.

**Request Body**:
```json
{
  "assignmentId": "task-123",
  "rejectingUserId": "user-456",
  "signature": "data:image/png;base64,iVBORw0KGg...",
  "remarks": "Need clarification on department cost center"
}
```

**Response**:
```json
{
  "success": true,
  "data": {
    "taskId": "task-123",
    "action": "rejected",
    "newStatus": "rejected",
    "reason": "Need clarification on department cost center",
    "timestamp": "2024-12-01T11:00:00Z"
  }
}
```

#### POST /api/approvals/tasks/:id/reassign
Reassign a task to different approver.

**Request Body**:
```json
{
  "assignmentId": "task-123",
  "newApproverId": "user-789",
  "newApproverName": "Sarah Johnson",
  "reason": "Manager on leave, delegating to colleague"
}
```

**Response**:
```json
{
  "success": true,
  "data": {
    "taskId": "task-123",
    "action": "reassigned",
    "newApprover": "Sarah Johnson",
    "timestamp": "2024-12-01T11:00:00Z"
  }
}
```

### Bulk Operations Routes

#### POST /api/approvals/bulk/approve
Approve multiple tasks.

**Request Body**:
```json
{
  "taskIds": ["task-1", "task-2", "task-3"],
  "remarks": "All reviewed and approved",
  "userId": "user-456"
}
```

**Response**:
```json
{
  "success": true,
  "data": {
    "approved": 3,
    "failed": 0,
    "message": "Successfully approved 3 tasks",
    "timestamp": "2024-12-01T11:00:00Z"
  }
}
```

#### POST /api/approvals/bulk/reject
Reject multiple tasks.

**Request Body**:
```json
{
  "taskIds": ["task-1", "task-2", "task-3"],
  "remarks": "All missing required GL codes",
  "userId": "user-456"
}
```

**Response**:
```json
{
  "success": true,
  "data": {
    "rejected": 3,
    "failed": 0,
    "message": "Successfully rejected 3 tasks with reason"
  }
}
```

#### POST /api/approvals/bulk/reassign
Reassign multiple tasks.

**Request Body**:
```json
{
  "taskIds": ["task-1", "task-2"],
  "newApproverId": "user-789",
  "newApproverName": "Finance Officer",
  "reason": "Escalating to finance team"
}
```

**Response**:
```json
{
  "success": true,
  "data": {
    "reassigned": 2,
    "newApprover": "Finance Officer"
  }
}
```

### Analytics Routes

#### GET /api/analytics/metrics
Get dashboard metrics.

**Query Parameters**:
- `userId` (optional): User-specific or global metrics

**Response**:
```json
{
  "success": true,
  "data": {
    "totalPending": 24,
    "totalApproved": 187,
    "totalRejected": 12,
    "avgApprovalTime": "3.2 days",
    "slaCompliance": 94,
    "bottleneckStage": "Finance Officer Review",
    "bottleneckDays": 4.5
  }
}
```

#### GET /api/analytics/trends
Get 7-day approval trends.

**Query Parameters**:
- `days` (optional): Number of days (default: 7)
- `userId` (optional): User-specific trends

**Response**:
```json
{
  "success": true,
  "data": [
    {
      "date": "2024-11-25",
      "approved": 12,
      "rejected": 2,
      "pending": 8
    },
    {
      "date": "2024-11-26",
      "approved": 18,
      "rejected": 1,
      "pending": 12
    }
  ]
}
```

#### GET /api/analytics/bottleneck
Get stage performance and bottleneck analysis.

**Response**:
```json
{
  "success": true,
  "data": {
    "stages": [
      {
        "stage": "Department Manager",
        "avgTime": "1.2 days",
        "itemCount": 45,
        "slaCompliance": 98
      },
      {
        "stage": "Finance Officer",
        "avgTime": "4.5 days",
        "itemCount": 38,
        "slaCompliance": 85
      }
    ],
    "bottleneck": {
      "stage": "Finance Officer",
      "avgTime": "4.5 days",
      "recommendations": [
        "Add finance officer capacity",
        "Review approval criteria",
        "Implement parallel approvals"
      ]
    }
  }
}
```

### Workflow Routes

#### GET /api/workflows
List all workflows.

**Response**:
```json
{
  "success": true,
  "data": {
    "workflows": [
      {
        "id": "wf-req-001",
        "name": "3-Stage Requisition",
        "entityType": "REQUISITION",
        "stages": 3,
        "status": "published"
      }
    ]
  }
}
```

#### GET /api/workflows/:id
Get workflow details.

**Response**:
```json
{
  "success": true,
  "data": {
    "id": "wf-req-001",
    "name": "3-Stage Requisition",
    "description": "Approval workflow for purchase requisitions",
    "entityType": "REQUISITION",
    "stages": [
      {
        "order": 0,
        "name": "Manager Review",
        "approvers": ["role:manager"]
      },
      {
        "order": 1,
        "name": "Finance Officer Review",
        "approvers": ["role:finance"]
      },
      {
        "order": 2,
        "name": "CFO Approval",
        "approvers": ["role:cfo"]
      }
    ],
    "status": "published"
  }
}
```

### Error Responses

All endpoints return consistent error format:

```json
{
  "success": false,
  "error": "Validation failed",
  "details": {
    "field": "message"
  }
}
```

**Common Error Codes**:
- 400: Bad Request (validation failed)
- 401: Unauthorized (not authenticated)
- 403: Forbidden (insufficient permissions)
- 404: Not Found (resource not found)
- 500: Internal Server Error

## Phase 12 Changes

All function signatures stay the same. Only backend changes:
- Replace localStorage with PostgreSQL
- Add real authentication (OAuth 2.0)
- Add email notifications
- Add audit logging
- Replace server actions with REST API endpoints (or keep server actions)
