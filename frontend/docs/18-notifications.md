# Notification System

Frontend notification system with real-time updates and proper state management.

## Architecture

- **Server Actions**: `app/_actions/notifications.ts` - API communication
- **Hooks**: `hooks/use-notifications.ts` - React Query integration
- **Components**: `components/notifications/` - UI components

## Key Files

```
app/_actions/notifications.ts     # Server actions with authenticatedApiClient
hooks/use-notifications.ts        # React Query hooks
components/notifications/
├── notification-bell.tsx         # Header notification bell
├── notification-header.tsx       # Notification page header
└── notification-action-modal.tsx # Action modal
```

## Hooks

```typescript
// Get notifications with pagination
const { data, isLoading } = useNotifications({ page: 1, limit: 20 });

// Get notification statistics
const { data: stats } = useNotificationStats();

// Mark notifications as read
const markAsRead = useMarkAsRead();
markAsRead.mutate(["notif-id-1", "notif-id-2"]);

// Real-time polling
useNotificationPolling(userId, 30000); // Poll every 30 seconds
```

## Server Actions Pattern

All API calls use the established server action pattern with `authenticatedApiClient`:

```typescript
export async function getNotifications(params) {
  const response = await authenticatedApiClient({
    method: "GET",
    url: `/api/v1/notifications?${searchParams}`,
  });
  return successResponse(response.data?.data || []);
}
```

## State Management

Uses TanStack Query for:

- Caching and background updates
- Optimistic updates
- Error handling and retries
- Real-time synchronization
