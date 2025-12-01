# Phase 7: Quick Start Guide

Get up and running with Phase 7 notification components in minutes.

---

## 🚀 Quick Start

### 1. Add Notification Bell to Header

In your header component (e.g., `src/components/layout/header/index.tsx`):

```tsx
import Notifications from "@/components/layout/header/notifications";

export default function Header() {
  return (
    <header>
      {/* Other header items */}
      <Notifications /> {/* Already configured to use NotificationBell */}
    </header>
  );
}
```

**That's it!** The bell will automatically:
- Show current user's unread notification count
- Display recent notifications in dropdown
- Auto-refresh every 30 seconds
- Mark notifications as read on click

---

### 2. Use Notification Modal for Actions

When user clicks to approve/reject:

```tsx
import { NotificationActionModal } from "@/components/notifications/notification-action-modal";
import { useState } from "react";

export default function ApprovalPanel() {
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [notification, setNotification] = useState(null);

  const handleApprove = () => {
    setNotification(task);
    setIsModalOpen(true);
  };

  return (
    <>
      <button onClick={handleApprove}>Approve</button>

      <NotificationActionModal
        notification={notification!}
        isOpen={isModalOpen}
        onOpenChange={setIsModalOpen}
        onApprove={async (signature, remarks) => {
          await approveWithSignature(task.id, signature, remarks);
          setIsModalOpen(false);
        }}
        onReject={async (reason) => {
          await rejectWithReason(task.id, reason);
          setIsModalOpen(false);
        }}
      />
    </>
  );
}
```

---

### 3. Display Notifications Using NotificationItem

For lists of notifications:

```tsx
import { NotificationItem } from "@/components/notifications/notification-item";
import { useDeleteNotification } from "@/hooks/use-notifications";

export default function NotificationList({ notifications }) {
  const deleteNotif = useDeleteNotification();

  return (
    <div className="space-y-3">
      {notifications.map((notif) => (
        <NotificationItem
          key={notif.id}
          notification={notif}
          variant="full" // or "compact"
          onDelete={(id) => deleteNotif.mutateAsync({ notificationId: id })}
        />
      ))}
    </div>
  );
}
```

---

### 4. Add Notification Preferences Page

User settings for notification types:

```tsx
import { NotificationPreferences } from "@/components/notifications/notification-preferences";
import { useParams } from "next/navigation";

export default function SettingsPage() {
  const { userId } = useParams();

  return (
    <div className="max-w-2xl">
      <h1>Notification Settings</h1>
      <NotificationPreferences userId={userId} />
    </div>
  );
}
```

---

### 5. Access Full Notifications Page

Already available at: `/workflows/notifications`

No additional setup needed! Users can access:
- Full notification history with pagination
- Filter by type, status, and search message
- Mark all as read
- Delete notifications
- Bulk select with checkboxes

---

## 📚 Common Tasks

### Get Unread Count

```tsx
import { useUnreadNotificationCount } from "@/hooks/use-notifications";

const { data } = useUnreadNotificationCount(userId);
const unreadCount = data?.count || 0;
```

### Mark All as Read

```tsx
import { useMarkAllNotificationsAsRead } from "@/hooks/use-notifications";

const mutation = useMarkAllNotificationsAsRead();
await mutation.mutateAsync({ userId });
```

### Fetch Paginated Notifications

```tsx
import { useUserNotifications } from "@/hooks/use-notifications";

const { data, isLoading } = useUserNotifications(
  userId,
  page, // 1-based
  20, // items per page
  {
    type: "TASK_ASSIGNED", // optional filter
    isRead: false, // optional filter
  }
);

const notifications = data?.notifications || [];
```

### Set Up Auto-Polling

```tsx
import { useNotificationPolling } from "@/hooks/use-notifications";

// Automatically polls every 30 seconds
useNotificationPolling(userId, 30 * 1000);
```

---

## 🎯 Use Cases

### Use Case 1: Display Unread Badge

```tsx
import { useUnreadNotificationCount } from "@/hooks/use-notifications";

function UnreadBadge({ userId }) {
  const { data } = useUnreadNotificationCount(userId);
  const count = data?.count || 0;

  if (count === 0) return null;

  return (
    <span className="bg-red-500 text-white rounded-full px-2 py-1">
      {count > 9 ? "9+" : count}
    </span>
  );
}
```

### Use Case 2: Show Recent Notifications

```tsx
import { useNotificationBell } from "@/hooks/use-notifications";
import { NotificationItem } from "@/components/notifications/notification-item";

function RecentNotifications({ userId }) {
  const { recentNotifications, isLoading } = useNotificationBell(userId);

  if (isLoading) return <div>Loading...</div>;

  return (
    <div className="space-y-2">
      {recentNotifications.map((notif) => (
        <NotificationItem
          key={notif.id}
          notification={notif}
          variant="compact"
        />
      ))}
    </div>
  );
}
```

### Use Case 3: Notification List with Filters

```tsx
import { useUserNotifications } from "@/hooks/use-notifications";
import { useState } from "react";

function NotificationFilters({ userId }) {
  const [typeFilter, setTypeFilter] = useState(null);
  const [statusFilter, setStatusFilter] = useState(null);

  const { data } = useUserNotifications(userId, 1, 20, {
    type: typeFilter,
    isRead: statusFilter === "read" ? true : statusFilter === "unread" ? false : undefined,
  });

  return (
    <div>
      <select onChange={(e) => setTypeFilter(e.target.value)}>
        <option value="">All Types</option>
        <option value="TASK_ASSIGNED">Task Assigned</option>
        <option value="TASK_APPROVED">Task Approved</option>
        {/* ... */}
      </select>

      {data?.notifications.map((notif) => (
        <NotificationItem key={notif.id} notification={notif} />
      ))}
    </div>
  );
}
```

### Use Case 4: Quick Approve with Modal

```tsx
import { NotificationActionModal } from "@/components/notifications/notification-action-modal";
import { useApproveTask } from "@/hooks/use-workflows";
import { useState } from "react";

function QuickApprove({ task }) {
  const [isOpen, setIsOpen] = useState(false);
  const approveMutation = useApproveTask();

  return (
    <>
      <button onClick={() => setIsOpen(true)}>Approve Now</button>

      <NotificationActionModal
        notification={task.notification}
        isOpen={isOpen}
        onOpenChange={setIsOpen}
        onApprove={async (signature, remarks) => {
          await approveMutation.mutateAsync({
            taskId: task.id,
            signature,
            remarks,
          });
          setIsOpen(false);
        }}
        actionType="approve"
      />
    </>
  );
}
```

---

## 🔧 Configuration

### Polling Interval

Change auto-refresh interval in your component:

```tsx
// Refresh every 60 seconds instead of 30
useNotificationPolling(userId, 60 * 1000);
```

### Compact vs Full Display

Choose variant based on context:

```tsx
// Dropdown or sidebar - use compact
<NotificationItem notification={notif} variant="compact" />

// Notification page - use full
<NotificationItem notification={notif} variant="full" />
```

### Modal Action Type

Pre-select modal action:

```tsx
// Opens with approve already selected
<NotificationActionModal
  notification={notif}
  actionType="approve"
  // ...
/>
```

---

## 📋 Type Definitions

### Notification Types (7 types)

```typescript
type NotificationType =
  | "TASK_ASSIGNED"      // New task assigned to you
  | "TASK_REASSIGNED"    // Task reassigned to you
  | "TASK_APPROVED"      // Your submission approved
  | "TASK_REJECTED"      // Your submission rejected
  | "WORKFLOW_COMPLETE"  // Workflow fully completed
  | "APPROVAL_OVERDUE"   // Approval past due date
  | "COMMENT_ADDED";     // Comment added to item
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
  createdAt: Date;
  entityId?: string;
  entityType?: string;
  entityNumber?: string;
  rejectionReason?: string;
  reassignmentReason?: string;
}
```

---

## 🧪 Testing Tips

### Test Bell Component

```tsx
// In your test file
import { render, screen } from "@testing-library/react";
import { NotificationBell } from "@/components/notifications/notification-bell";

test("displays unread count", () => {
  render(<NotificationBell userId="123" />);
  // Mock useNotificationBell to return unreadCount: 5
  expect(screen.getByText("5")).toBeInTheDocument();
});
```

### Test Modal Approve

```tsx
test("calls onApprove with signature and remarks", async () => {
  const onApprove = jest.fn();
  const notification = { /* ... */ };

  render(
    <NotificationActionModal
      notification={notification}
      isOpen={true}
      onApprove={onApprove}
      onOpenChange={() => {}}
    />
  );

  // Fill form and submit
  const approveButton = screen.getByText("Approve");
  fireEvent.click(approveButton);

  expect(onApprove).toHaveBeenCalled();
});
```

---

## ⚠️ Common Issues

### Issue: Notifications not updating

**Solution**: Ensure polling is active
```tsx
useNotificationPolling(userId, 30 * 1000);
```

### Issue: Modal not opening

**Solution**: Verify state management
```tsx
const [isOpen, setIsOpen] = useState(false);
<NotificationActionModal isOpen={isOpen} onOpenChange={setIsOpen} />
```

### Issue: Signature not captured

**Solution**: Check Canvas ref is properly initialized
- Canvas element should render
- Mouse events should work on canvas
- Check browser console for errors

### Issue: Preferences not saving

**Solution**: Verify user ID and network
```tsx
const mutation = useUpdateNotificationPreferences();
// Check React DevTools for mutation state
```

---

## 📖 Documentation Links

- [Detailed Component Reference](PHASE_7_COMPONENT_REFERENCE.md)
- [Full Completion Report](PHASE_7_COMPLETION.md)
- [Phase 8 Readiness Guide](PHASE_8_READINESS.md)

---

## 🎓 Next Steps

1. **Integrate into your app**
   - Add NotificationBell to header ✓
   - Use NotificationActionModal in approval flows
   - Add NotificationsPage to navigation

2. **Customize styling**
   - Update Tailwind color classes if needed
   - Adjust padding/spacing for your design
   - Modify notification type colors

3. **Hook up workflows**
   - Connect modal to your approval endpoint
   - Trigger notifications on workflow events
   - Set user preferences

4. **Test end-to-end**
   - Create a test notification
   - Approve/reject through modal
   - Verify in notifications page

---

## 🚀 Production Checklist

- [ ] Bell appears in header
- [ ] Unread count updates
- [ ] Modal opens for actions
- [ ] Signature captures correctly
- [ ] Notifications persist after approval
- [ ] Preferences save and load
- [ ] All error cases handled gracefully
- [ ] Mobile responsive works
- [ ] Performance acceptable (polling interval set)

---

**You're ready to use Phase 7 components!** 🎉
