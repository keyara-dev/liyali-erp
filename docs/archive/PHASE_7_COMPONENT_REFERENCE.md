# Phase 7: Component Reference Guide

Quick reference for all Phase 7 notification components.

## Components Overview

| Component               | Path                                                         | Lines | Purpose                                     |
| ----------------------- | ------------------------------------------------------------ | ----- | ------------------------------------------- |
| NotificationBell        | `src/components/notifications/notification-bell.tsx`         | 175   | Header bell with unread badge and dropdown  |
| NotificationActionModal | `src/components/notifications/notification-action-modal.tsx` | 380   | Approve/reject modal with signature capture |
| NotificationItem        | `src/components/notifications/notification-item.tsx`         | 220   | Reusable notification display component     |
| NotificationPreferences | `src/components/notifications/notification-preferences.tsx`  | 150   | User notification settings                  |
| NotificationsPage       | `src/app/(private)/(main)/notifications/page.tsx`            | 210+  | Full notifications history page             |

## Component Usage Examples

### NotificationBell

```tsx
import { NotificationBell } from "@/components/notifications/notification-bell";

// In your header component (server component wrapper)
<NotificationBell userId={user.id} />;
```

**Props**:

- `userId: string` - User ID to fetch notifications for

**Features**:

- Auto-refresh every 30 seconds
- Shows unread count badge
- Recent 5 notifications in dropdown
- Click to mark as read

---

### NotificationActionModal

```tsx
import { NotificationActionModal } from "@/components/notifications/notification-action-modal";

const [isModalOpen, setIsModalOpen] = useState(false);
const [notification, setNotification] = useState<Notification | null>(null);

<NotificationActionModal
  notification={notification!}
  isOpen={isModalOpen}
  onOpenChange={setIsModalOpen}
  onApprove={async (signature, remarks) => {
    // Handle approval with signature and optional remarks
    await approveWorkflow(signature, remarks);
  }}
  onReject={async (reason) => {
    // Handle rejection with required reason
    await rejectWorkflow(reason);
  }}
  actionType="approve" // or "reject"
/>;
```

**Props**:

- `notification: Notification` - Notification being actioned
- `isOpen: boolean` - Modal visibility
- `onOpenChange: (open: boolean) => void` - Close handler
- `onApprove: (signature: string, remarks?: string) => Promise<void>` - Approve callback
- `onReject: (reason: string) => Promise<void>` - Reject callback
- `actionType?: "approve" | "reject"` - Pre-select action type

**Features**:

- Two-mode UI (preview → action)
- Digital signature capture with HTML5 Canvas
- Form validation
- Loading states during submission

---

### NotificationItem

```tsx
import { NotificationItem } from '@/components/notifications/notification-item';

// Compact variant (for dropdowns)
<NotificationItem
  notification={notification}
  variant="compact"
  onMarkAsRead={(id) => markAsRead(id)}
/>

// Full variant (for notification history)
<NotificationItem
  notification={notification}
  variant="full"
  onDelete={(id) => deleteNotification(id)}
  showCheckbox={true}
  isSelected={selectedIds.includes(notification.id)}
  onSelectionChange={(checked) => {
    // Handle selection
  }}
/>
```

**Props**:

- `notification: Notification` - The notification to display
- `variant?: "compact" | "full"` - Display variant (default: "full")
- `onDelete?: (id: string) => void` - Delete handler
- `onMarkAsRead?: (id: string) => void` - Mark as read handler
- `isDeleting?: boolean` - Show loading state for delete (default: false)
- `showCheckbox?: boolean` - Show/hide checkbox (default: false)
- `isSelected?: boolean` - Checkbox selection state (default: false)
- `onSelectionChange?: (selected: boolean) => void` - Checkbox change handler

**Features**:

- Compact and full display variants
- Notification type icons and colored badges
- Unread indicator dot
- Relative timestamps
- Rejection/reassignment reason display
- Checkbox support for bulk operations

---

### NotificationPreferences

```tsx
import { NotificationPreferences } from "@/components/notifications/notification-preferences";

<NotificationPreferences
  userId={user.id}
  onSaved={() => {
    // Preferences saved successfully
    toast.success("Preferences saved");
  }}
/>;
```

**Props**:

- `userId: string` - User ID
- `onSaved?: () => void` - Callback when preferences saved

**Features**:

- 7 notification type toggles
- Descriptive text for each type
- Change detection (save button disabled when no changes)
- Success confirmation message
- Error handling with logging

---

### NotificationsPage (Server Component)

```tsx
// Automatic - serves at /workflows/notifications
// Users can access directly via navigation
```

**Features**:

- Paginated notifications list
- Type, status, and message search filters
- Mark all as read button
- Delete individual notifications
- Bulk selection with checkboxes
- "Load more" pagination
- Empty state handling
- Loading skeleton UI

---

## Related Hooks

### useNotificationBell()

```tsx
const { unreadCount, recentNotifications, isLoading } =
  useNotificationBell(userId);
```

Returns unread count and recent 5 notifications.

### useMarkNotificationAsRead()

```tsx
const mutation = useMarkNotificationAsRead();
await mutation.mutateAsync({ notificationId });
```

Marks a single notification as read.

### useMarkAllNotificationsAsRead()

```tsx
const mutation = useMarkAllNotificationsAsRead();
await mutation.mutateAsync({ userId });
```

Marks all notifications as read for a user.

### useDeleteNotification()

```tsx
const mutation = useDeleteNotification();
await mutation.mutateAsync({ notificationId });
```

Deletes a notification.

### useUserNotifications()

```tsx
const { data, isLoading } = useUserNotifications(userId, page, pageSize, {
  type,
  isRead,
});
```

Fetches paginated notifications with optional filters.

### useGetNotificationPreferences()

```tsx
const { data: preferences } = useGetNotificationPreferences({ userId });
```

Fetches notification preferences for a user.

### useUpdateNotificationPreferences()

```tsx
const mutation = useUpdateNotificationPreferences();
await mutation.mutateAsync({ userId, preferences });
```

Updates notification preferences.

### useNotificationPolling()

```tsx
useNotificationPolling(userId, 30 * 1000); // Poll every 30 seconds
```

Sets up automatic polling for notifications.

---

## Data Types

### Notification Type (7 types)

```typescript
type NotificationType =
  | "TASK_ASSIGNED" // New approval task
  | "TASK_REASSIGNED" // Task reassigned to you
  | "TASK_APPROVED" // Your submission approved
  | "TASK_REJECTED" // Your submission rejected
  | "WORKFLOW_COMPLETE" // Workflow fully completed
  | "APPROVAL_OVERDUE" // Overdue approval
  | "COMMENT_ADDED"; // Comment added to item
```

### Notification Interface

```typescript
interface Notification {
  id: string;
  userId: string;
  type: NotificationType;
  title: string;
  message: string;
  isRead: boolean;
  actionTaken: boolean;
  createdAt: Date;
  expiresAt?: Date;
  entityId?: string;
  entityType?: string;
  entityNumber?: string;
  relatedUserId?: string;
  relatedUserName?: string;
  quickAction?: {
    type: string;
    label: string;
    params: Record<string, any>;
  };
  rejectionReason?: string;
  reassignmentReason?: string;
  importance?: "LOW" | "MEDIUM" | "HIGH";
}
```

### NotificationPreferences Interface

```typescript
interface NotificationPreferences {
  userId: string;
  TASK_ASSIGNED: boolean;
  TASK_REASSIGNED: boolean;
  TASK_APPROVED: boolean;
  TASK_REJECTED: boolean;
  WORKFLOW_COMPLETE: boolean;
  APPROVAL_OVERDUE: boolean;
  COMMENT_ADDED: boolean;
}
```

---

## Styling & Colors

### Notification Type Colors

```typescript
const colors = {
  TASK_ASSIGNED: "bg-blue-100 text-blue-800",
  TASK_REASSIGNED: "bg-purple-100 text-purple-800",
  TASK_APPROVED: "bg-green-100 text-green-800",
  TASK_REJECTED: "bg-red-100 text-red-800",
  WORKFLOW_COMPLETE: "bg-yellow-100 text-yellow-800",
  APPROVAL_OVERDUE: "bg-orange-100 text-orange-800",
  COMMENT_ADDED: "bg-cyan-100 text-cyan-800",
};
```

### Icon Mapping

- TASK_ASSIGNED: ⚡ (Zap)
- TASK_REASSIGNED: 🔄 (Repeat2)
- TASK_APPROVED: ✅ (CheckCircle2)
- TASK_REJECTED: ⚠️ (AlertCircle)
- WORKFLOW_COMPLETE: ⏱️ (Clock)
- APPROVAL_OVERDUE: ⚠️ (AlertCircle)
- COMMENT_ADDED: 💬 (MessageSquare)

---

## Query Keys (Constants)

Located in `src/lib/constants.ts`:

```typescript
QUERY_KEYS.NOTIFICATIONS = {
  ALL: "notifications-all",
  UNREAD: "notifications-unread",
  UNREAD_COUNT: "notifications-unread-count",
  PREFERENCES: "notification-preferences",
};
```

---

## Integration Points

### Phase 5 (Server Actions)

Notification components use these server actions:

- `getNotifications()`
- `markAsRead()`
- `markAllAsRead()`
- `deleteNotificationAction()`
- `getPreferences()`
- `updatePreferences()`

### Phase 6 (React Query)

All data fetching through React Query hooks:

- Automatic caching
- Automatic refetch on focus
- Mutation handling with optimistic updates
- Error boundary integration

### Authentication

- Uses custom JWT auth via `getCurrentUser()`
- Async server component pattern
- Suspense boundaries for streaming

---

## Best Practices

1. **Always provide userId to components** that require it
2. **Use mutation callbacks** for error handling and optimistic updates
3. **Implement Suspense boundaries** when using server components
4. **Clean up polling** when components unmount (handled automatically)
5. **Validate notification data** before passing to components
6. **Handle loading states** in parent components
7. **Use NotificationItem** for consistent display across the app
8. **Access preferences settings** before allowing notifications

---

## Troubleshooting

### Notifications not appearing?

1. Check `useNotificationBell()` is fetching data
2. Verify user has unread notifications
3. Check React Query cache status in React DevTools
4. Verify polling is active (30-second intervals)

### Modal not opening?

1. Verify `isOpen` state is true
2. Check `notification` prop is not null
3. Verify `onOpenChange` callback is working
4. Check form validation (signature/reason required)

### Preferences not saving?

1. Check user ID is correct
2. Verify network request in DevTools
3. Check for React Query mutations errors
4. Review server action responses

---

## Files Location Summary

```
src/
├── components/
│   └── notifications/
│       ├── notification-bell.tsx
│       ├── notification-action-modal.tsx
│       ├── notification-item.tsx
│       └── notification-preferences.tsx
├── app/
│   ├── (private)/
│   │   └── workflows/
│   │       └── notifications/
│   │           └── page.tsx
│   └── _actions/
│       └── notifications.ts (Phase 5)
├── hooks/
│   └── use-notifications.ts (Phase 6)
└── lib/
    └── constants.ts (Query keys)
```
