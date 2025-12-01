# Phase 5 API Reference

Quick reference guide for using Phase 5 server actions and hooks.

---

## Notification Operations

### Get Notifications (Server Action)

```typescript
import { getNotifications } from '@/app/_actions/notifications';

// Basic usage
const result = await getNotifications('user-123');
// result: { notifications, total, page, pageSize, hasMore }

// With pagination and filters
const result = await getNotifications(
  'user-123',
  2, // page
  10, // pageSize
  {
    type: 'TASK_ASSIGNED',
    isRead: false,
    startDate: new Date('2024-01-01'),
    endDate: new Date('2024-12-31'),
  }
);
```

### Get Unread Count (Server Action)

```typescript
import { getUnreadCount } from '@/app/_actions/notifications';

const result = await getUnreadCount({ userId: 'user-123' });
// result: { count: 5, userId: 'user-123' }
```

### Create Notification (Server Action)

```typescript
import { createNotificationAction } from '@/app/_actions/notifications';

const notification = await createNotificationAction({
  userId: 'user-123',
  type: 'TASK_ASSIGNED',
  title: 'New approval task',
  message: 'Requisition #REQ-001 needs your approval',
  entityId: 'req-001',
  entityType: 'REQUISITION',
  entityNumber: 'REQ-001',
  quickAction: {
    type: 'REVIEW_AND_APPROVE',
    label: 'Review Now',
    params: { entityId: 'req-001' },
  },
  importance: 'HIGH',
});
```

### Mark as Read (Server Action)

```typescript
import { markAsRead } from '@/app/_actions/notifications';

const result = await markAsRead({ notificationId: 'notif-123' });
```

### Mark All as Read (Server Action)

```typescript
import { markAllAsRead } from '@/app/_actions/notifications';

const result = await markAllAsRead({ userId: 'user-123' });
// result: { count: 5, success: true }
```

### Delete Notification (Server Action)

```typescript
import { deleteNotificationAction } from '@/app/_actions/notifications';

const result = await deleteNotificationAction({ notificationId: 'notif-123' });
```

### Get Preferences (Server Action)

```typescript
import { getPreferences } from '@/app/_actions/notifications';

const result = await getPreferences({ userId: 'user-123' });
// result: { preferences: NotificationPreferences }
```

### Update Preferences (Server Action)

```typescript
import { updatePreferences } from '@/app/_actions/notifications';

const result = await updatePreferences({
  userId: 'user-123',
  preferences: {
    emailNotifications: true,
    notifyOn: {
      taskAssigned: true,
      taskReassigned: true,
      taskApproved: false,
      // ... other fields
    },
  },
});
```

---

## Notification Hooks (Client Components)

### Use User Notifications

```typescript
'use client';

import { useUserNotifications } from '@/hooks/use-notifications';

export function NotificationPage() {
  const { data, isLoading, error } = useUserNotifications(
    'user-123',
    1, // page
    20, // pageSize
    { type: 'TASK_ASSIGNED' } // filters
  );

  if (isLoading) return <div>Loading...</div>;
  if (error) return <div>Error loading notifications</div>;

  return (
    <div>
      {data?.notifications.map((notif) => (
        <div key={notif.id}>{notif.message}</div>
      ))}
    </div>
  );
}
```

### Use Unread Count

```typescript
'use client';

import { useUnreadNotificationCount } from '@/hooks/use-notifications';

export function NotificationBell() {
  const { data } = useUnreadNotificationCount('user-123');

  return (
    <div>
      🔔 {data?.count || 0}
    </div>
  );
}
```

### Use Mark as Read

```typescript
'use client';

import { useMarkNotificationAsRead } from '@/hooks/use-notifications';

export function NotificationItem({ notification }) {
  const markAsReadMutation = useMarkNotificationAsRead();

  const handleClick = async () => {
    await markAsReadMutation.mutateAsync({
      notificationId: notification.id,
    });
  };

  return (
    <div onClick={handleClick}>
      {notification.message}
    </div>
  );
}
```

### Use Notification Bell

```typescript
'use client';

import { useNotificationBell } from '@/hooks/use-notifications';

export function NotificationBellComponent() {
  const { unreadCount, recentNotifications, isLoading } =
    useNotificationBell('user-123');

  return (
    <div>
      <button>🔔 {unreadCount}</button>
      {recentNotifications.map((notif) => (
        <div key={notif.id}>{notif.message}</div>
      ))}
    </div>
  );
}
```

### Use Notification Polling

```typescript
'use client';

import { useNotificationPolling } from '@/hooks/use-notifications';

export function NotificationListener() {
  // Polls every 30 seconds
  useNotificationPolling('user-123', 30 * 1000);

  // Component renders while polling happens in background
  return <div>Notifications polling active</div>;
}
```

### Use Quick Action Handler

```typescript
'use client';

import { useQuickActionHandler } from '@/hooks/use-notifications';

export function QuickActionButton({ notificationId }) {
  const handleAction = useQuickActionHandler();

  const onClick = async () => {
    const success = await handleAction(notificationId);
    if (success) {
      // Notification marked as read and action taken
      // Navigate or show success toast
    }
  };

  return <button onClick={onClick}>Review Now</button>;
}
```

---

## Workflow Operations

### Create Workflow (Server Action)

```typescript
import { createWorkflow } from '@/app/_actions/workflows';

const workflow = await createWorkflow({
  name: '2-Stage Approval',
  description: 'Manager then Finance approval',
  applicableEntityTypes: ['REQUISITION', 'BUDGET'],
  isTemplate: true,
  stages: [
    {
      stageNumber: 1,
      stageName: 'Manager Review',
      approverAssignmentType: 'ROLE',
      requiredRole: 'DEPARTMENT_MANAGER',
      requiresSignature: true,
      canBeReassigned: true,
      displayOrder: 1,
      onApprove: { nextStage: 2 },
      onReject: { nextStage: 'DRAFT' },
    },
    {
      stageNumber: 2,
      stageName: 'Finance Review',
      approverAssignmentType: 'ROLE',
      requiredRole: 'FINANCE_OFFICER',
      requiresSignature: true,
      canBeReassigned: true,
      displayOrder: 2,
      onApprove: { nextStage: 'FINAL', setEntityStatus: 'APPROVED' },
      onReject: { nextStage: 'DRAFT' },
    },
  ],
  createdBy: 'admin-123',
});
```

### Assign Workflow (Server Action)

```typescript
import { assignWorkflowAction } from '@/app/_actions/workflows';

const assignment = await assignWorkflowAction({
  entityId: 'req-001',
  entityType: 'REQUISITION',
  workflowId: 'workflow-123',
  assignedBy: 'user-123',
});
```

### Approve Stage (Server Action)

```typescript
import { approveStageAction } from '@/app/_actions/workflows';

const result = await approveStageAction({
  assignmentId: 'assignment-123',
  approverId: 'user-456',
  approverName: 'John Manager',
  approverRole: 'DEPARTMENT_MANAGER',
  comments: 'Looks good',
  signature: 'base64-encoded-signature',
  entityNumber: 'REQ-001',
});

if (result.isComplete) {
  console.log('Workflow complete!');
} else {
  console.log('Next approver:', result.nextApprover);
}
```

### Reject Stage (Server Action)

```typescript
import { rejectStageAction } from '@/app/_actions/workflows';

const result = await rejectStageAction({
  assignmentId: 'assignment-123',
  rejectorId: 'user-456',
  rejectorName: 'John Manager',
  rejectionRemarks: 'Budget exceeded',
  entityNumber: 'REQ-001',
});

console.log('Workflow routed to:', result.targetStage);
```

### Reassign Stage (Server Action)

```typescript
import { reassignStageAction } from '@/app/_actions/workflows';

const result = await reassignStageAction({
  assignmentId: 'assignment-123',
  stageNumber: 1,
  newApproverId: 'user-789',
  reassignedBy: 'admin-123',
  reassignedByName: 'System Admin',
  reassignmentReason: 'Original approver out sick',
  entityNumber: 'REQ-001',
});

console.log('Reassigned from:', result.oldApprover?.name);
console.log('Reassigned to:', result.newApprover.name);
```

### Get Pending Approvals (Server Action)

```typescript
import { getPendingApprovalsAction } from '@/app/_actions/workflows';

const pending = await getPendingApprovalsAction('user-456');
// pending: WorkflowAssignment[] with currentStageNumber > 0
```

### Set Default Workflow (Server Action)

```typescript
import { setDefaultWorkflowAction } from '@/app/_actions/workflows';

await setDefaultWorkflowAction('REQUISITION', 'workflow-123');
```

### Get Default Workflow (Server Action)

```typescript
import { getDefaultWorkflowAction } from '@/app/_actions/workflows';

const defaultWorkflow = await getDefaultWorkflowAction('REQUISITION');
```

---

## Notification Trigger Helpers

These are called internally by workflow actions but can also be called manually:

### Notify Task Assigned

```typescript
import { notifyTaskAssigned } from '@/app/_actions/notifications';

await notifyTaskAssigned(
  'user-456', // approverId
  'John Manager', // approverName
  'req-001', // entityId
  'REQUISITION', // entityType
  'REQ-2024-001', // entityNumber
  'Manager Review' // currentStageName
);
```

### Notify Task Reassigned

```typescript
import { notifyTaskReassigned } from '@/app/_actions/notifications';

await notifyTaskReassigned(
  'user-789', // newApproverId
  'Jane Finance', // newApproverName
  'req-001', // entityId
  'REQUISITION', // entityType
  'REQ-2024-001', // entityNumber
  'admin-123', // reassignedBy
  'System Admin', // reassignedByName
  'Original approver out sick' // reassignmentReason
);
```

### Notify Task Approved

```typescript
import { notifyTaskApproved } from '@/app/_actions/notifications';

await notifyTaskApproved(
  'user-001', // createdById
  'req-001', // entityId
  'REQUISITION', // entityType
  'REQ-2024-001', // entityNumber
  'user-456', // approvedBy
  'John Manager' // approvedByName
);
```

### Notify Task Rejected

```typescript
import { notifyTaskRejected } from '@/app/_actions/notifications';

await notifyTaskRejected(
  'user-001', // createdById
  'req-001', // entityId
  'REQUISITION', // entityType
  'REQ-2024-001', // entityNumber
  'user-456', // rejectedBy
  'John Manager', // rejectedByName
  'Budget exceeded' // rejectionReason
);
```

### Notify Workflow Complete

```typescript
import { notifyWorkflowComplete } from '@/app/_actions/notifications';

await notifyWorkflowComplete(
  'user-001', // createdById
  'req-001', // entityId
  'REQUISITION', // entityType
  'REQ-2024-001', // entityNumber
  'user-456', // finalApprovedBy
  'Finance Officer' // finalApprovedByName
);
```

---

## Common Patterns

### Complete Approval Flow

```typescript
// 1. User clicks approve in modal
async function handleApprove() {
  // 2. Call approval action
  const result = await approveStageAction({
    assignmentId,
    approverId: currentUser.id,
    approverName: currentUser.name,
    approverRole: currentUser.role,
    comments: remarks,
    signature: signatureBase64,
    entityNumber: requisition.number,
  });

  // 3. Action automatically:
  // - Records approval
  // - Progresses workflow
  // - Creates notifications for next approver
  // - Marks complete if final

  if (result.isComplete) {
    toast.success('Approval complete!');
    // Redirect to dashboard
  } else {
    toast.success('Moved to next approval stage');
    // Stay on page or navigate
  }
}
```

### Reassignment Flow

```typescript
async function handleReassign() {
  const result = await reassignStageAction({
    assignmentId,
    stageNumber: currentStageNumber,
    newApproverId: selectedUserId,
    reassignedBy: currentUser.id,
    reassignedByName: currentUser.name,
    reassignmentReason: reason,
    entityNumber: requisition.number,
  });

  // Action automatically:
  // - Creates audit trail
  // - Notifies new approver
  // - Updates assignment

  toast.success(
    `Reassigned from ${result.oldApprover?.name} to ${result.newApprover.name}`
  );
}
```

### Notification Bell Setup

```typescript
export function AppBar() {
  const { unreadCount, recentNotifications } = useNotificationBell(
    currentUser.id
  );

  // Start polling
  useNotificationPolling(currentUser.id, 30 * 1000);

  return (
    <header>
      <div>
        <button className="relative">
          🔔
          {unreadCount > 0 && (
            <span className="badge">{unreadCount}</span>
          )}
        </button>
        <NotificationDropdown notifications={recentNotifications} />
      </div>
    </header>
  );
}
```

---

## Type Imports

```typescript
import type {
  Notification,
  NotificationType,
  NotificationPreferences,
  QuickAction,
  CustomWorkflow,
  WorkflowAssignment,
  StageExecution,
  StageAssignment,
} from '@/types';
```

---

## Error Handling

All server actions and hooks throw errors on failure:

```typescript
try {
  const result = await getNotifications(userId);
} catch (error) {
  console.error('Failed to fetch notifications:', error);
  // Show toast or error message
}
```

For hooks, check `error` property:

```typescript
const { data, isError, error } = useUserNotifications(userId);

if (isError) {
  return <ErrorMessage error={error} />;
}
```

---

## Next Phase

Once UI components are created in Phase 6-7, these APIs will be integrated into:

- Notification bell dropdown
- Approval modals
- Workflow selection forms
- Reassignment interfaces
- Notifications history page
- Admin dashboards

---

**Status**: ✅ Ready for Phase 6-7 UI implementation
