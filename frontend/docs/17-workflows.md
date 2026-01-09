# Workflow System

Document approval workflows with multi-stage approvals and real-time tracking.

## Architecture

- **Server Actions**: `app/_actions/workflow-approval-actions.ts` - Approval API
- **Components**: `components/workflows/` - Workflow UI components
- **Services**: Backend workflow execution service integration

## Key Files

```
app/_actions/workflow-approval-actions.ts  # Approval workflow actions
components/workflows/
├── approval-action-panel.tsx              # Approval/rejection UI
└── workflow-details-form.tsx              # Workflow configuration
app/(private)/(main)/tasks/                # Task management pages
```

## Approval Actions

```typescript
// Get approval tasks
const tasks = await getApprovalTasks({ status: "PENDING" });

// Approve a task
await approveApprovalTask(taskId, {
  comments: "Approved for processing",
  signature: "digital_signature_hash",
});

// Reject a task
await rejectApprovalTask(taskId, {
  remarks: "Budget exceeded",
  returnTo: "requester",
});

// Bulk operations
await bulkApproveApprovalTasks(["task1", "task2"], "Batch approval");
```

## Workflow Status

Track document progress through approval stages:

```typescript
const status = await getApprovalWorkflowStatus(documentId);
// Returns: currentStage, totalStages, nextApprover, canApprove, stageProgress
```

## Integration

Workflows integrate with:

- **Notifications**: Real-time approval notifications
- **Documents**: Requisitions, POs, Payment Vouchers, GRNs
- **Audit**: Complete approval history tracking
- **RBAC**: Permission-based approval rights
