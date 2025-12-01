/**
 * Notification React Query Hooks
 *
 * Custom React hooks for managing notification data fetching, caching, and mutations.
 * Uses React Query (TanStack Query) for efficient state management and real-time updates.
 */

'use client';

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  Notification,
  NotificationPreferences,
  NotificationType,
  GetNotificationsResponse,
  CreateNotificationRequest,
  MarkNotificationReadRequest,
  MarkAllNotificationsReadRequest,
  DeleteNotificationRequest,
  GetUnreadCountRequest,
  GetNotificationPreferencesRequest,
  UpdateNotificationPreferencesRequest,
} from '@/types';
import { QUERY_KEYS } from '@/lib/constants';

import {
  getNotifications,
  getUnreadNotifications,
  getUnreadCount,
  createNotificationAction,
  markAsRead,
  markAllAsRead,
  deleteNotificationAction,
  getPreferences,
  updatePreferences,
  markActionTaken,
} from '@/app/_actions/notifications';

const NOTIFICATIONS_QUERY_KEY = QUERY_KEYS.NOTIFICATIONS.ALL;
const UNREAD_COUNT_QUERY_KEY = QUERY_KEYS.NOTIFICATIONS.UNREAD_COUNT;
const PREFERENCES_QUERY_KEY = QUERY_KEYS.NOTIFICATIONS.PREFERENCES;

/**
 * Hook: Get paginated notifications for a user
 * @param userId User ID
 * @param page Page number
 * @param pageSize Items per page
 * @param filters Optional filters
 * @returns Query result with notifications
 */
export function useUserNotifications(
  userId: string,
  page: number = 1,
  pageSize: number = 20,
  filters?: {
    type?: NotificationType;
    isRead?: boolean;
    startDate?: Date;
    endDate?: Date;
  }
) {
  return useQuery({
    queryKey: [NOTIFICATIONS_QUERY_KEY, userId, page, pageSize, filters],
    queryFn: async () => getNotifications(userId, page, pageSize, filters),
    enabled: !!userId,
    staleTime: 30 * 1000, // 30 seconds
    refetchInterval: 60 * 1000, // Refetch every 60 seconds
  });
}

/**
 * Hook: Get unread notifications for a user
 * @param userId User ID
 * @returns Query result with unread notifications
 */
export function useUnreadNotifications(userId: string) {
  return useQuery({
    queryKey: [NOTIFICATIONS_QUERY_KEY, userId, 'unread'],
    queryFn: async () => getUnreadNotifications(userId),
    enabled: !!userId,
    staleTime: 10 * 1000, // 10 seconds
    refetchInterval: 30 * 1000, // Refetch every 30 seconds
  });
}

/**
 * Hook: Get unread notification count for a user
 * @param userId User ID
 * @returns Query result with count
 */
export function useUnreadNotificationCount(userId: string) {
  return useQuery({
    queryKey: [UNREAD_COUNT_QUERY_KEY, userId],
    queryFn: async () => getUnreadCount({ userId }),
    enabled: !!userId,
    staleTime: 10 * 1000, // 10 seconds
    refetchInterval: 30 * 1000, // Refetch every 30 seconds
  });
}

/**
 * Hook: Create a notification
 * @returns Mutation for creating notification
 */
export function useCreateNotification() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (request: CreateNotificationRequest) =>
      createNotificationAction(request),
    onSuccess: (data) => {
      // Invalidate relevant queries
      queryClient.invalidateQueries({
        queryKey: [NOTIFICATIONS_QUERY_KEY, data.notification.userId],
      });
      queryClient.invalidateQueries({
        queryKey: [UNREAD_COUNT_QUERY_KEY, data.notification.userId],
      });
    },
    onError: (error) => {
      console.error('Failed to create notification:', error);
    },
  });
}

/**
 * Hook: Mark a notification as read
 * @returns Mutation for marking notification as read
 */
export function useMarkNotificationAsRead() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (request: MarkNotificationReadRequest) =>
      markAsRead(request),
    onSuccess: (data) => {
      // Update query cache immediately
      queryClient.setQueryData(
        [NOTIFICATIONS_QUERY_KEY, data.notification.userId],
        (oldData: GetNotificationsResponse | undefined) => {
          if (!oldData) return oldData;
          return {
            ...oldData,
            notifications: oldData.notifications.map((n) =>
              n.id === data.notification.id ? data.notification : n
            ),
          };
        }
      );

      // Invalidate unread count
      queryClient.invalidateQueries({
        queryKey: [UNREAD_COUNT_QUERY_KEY, data.notification.userId],
      });
    },
    onError: (error) => {
      console.error('Failed to mark notification as read:', error);
    },
  });
}

/**
 * Hook: Mark all notifications as read
 * @returns Mutation for marking all as read
 */
export function useMarkAllNotificationsAsRead() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (request: MarkAllNotificationsReadRequest) =>
      markAllAsRead(request),
    onSuccess: (data) => {
      // Invalidate all notification queries for this user
      queryClient.invalidateQueries({
        queryKey: [NOTIFICATIONS_QUERY_KEY, data.userId],
      });

      // Invalidate unread count
      queryClient.invalidateQueries({
        queryKey: [UNREAD_COUNT_QUERY_KEY, data.userId],
      });
    },
    onError: (error) => {
      console.error('Failed to mark all notifications as read:', error);
    },
  });
}

/**
 * Hook: Delete a notification
 * @returns Mutation for deleting notification
 */
export function useDeleteNotification() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (request: DeleteNotificationRequest) =>
      deleteNotificationAction(request),
    onSuccess: (data, variables) => {
      // We don't know the userId here, so invalidate broadly
      queryClient.invalidateQueries({
        queryKey: [NOTIFICATIONS_QUERY_KEY],
      });
      queryClient.invalidateQueries({
        queryKey: [UNREAD_COUNT_QUERY_KEY],
      });
    },
    onError: (error) => {
      console.error('Failed to delete notification:', error);
    },
  });
}

/**
 * Hook: Mark notification action as taken
 * @returns Mutation for marking action taken
 */
export function useMarkNotificationActionTaken() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (notificationId: string) =>
      markActionTaken(notificationId),
    onSuccess: (notification) => {
      // Update notification queries
      queryClient.invalidateQueries({
        queryKey: [NOTIFICATIONS_QUERY_KEY, notification.userId],
      });
    },
    onError: (error) => {
      console.error('Failed to mark action taken:', error);
    },
  });
}

/**
 * Hook: Get notification preferences
 * @param userId User ID
 * @returns Query result with preferences
 */
export function useNotificationPreferences(userId: string) {
  return useQuery({
    queryKey: [PREFERENCES_QUERY_KEY, userId],
    queryFn: async () => getPreferences({ userId }),
    enabled: !!userId,
    staleTime: 5 * 60 * 1000, // 5 minutes
  });
}

/**
 * Hook: Update notification preferences
 * @returns Mutation for updating preferences
 */
export function useUpdateNotificationPreferences() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (request: UpdateNotificationPreferencesRequest) =>
      updatePreferences(request),
    onSuccess: (data) => {
      // Update preferences cache
      queryClient.setQueryData(
        [PREFERENCES_QUERY_KEY, data.preferences.userId],
        data
      );
    },
    onError: (error) => {
      console.error('Failed to update notification preferences:', error);
    },
  });
}

/**
 * Hook: Polling effect for real-time notifications
 * Automatically refetches notifications at intervals
 * @param userId User ID
 * @param pollingInterval Interval in ms (0 to disable)
 */
export function useNotificationPolling(
  userId: string,
  pollingInterval: number = 30 * 1000 // 30 seconds default
) {
  const queryClient = useQueryClient();

  // Set up polling
  useQuery({
    queryKey: [NOTIFICATIONS_QUERY_KEY, userId, 'polling'],
    queryFn: async () => {
      if (!userId) return null;
      const result = await getNotifications(userId, 1, 5);
      return result;
    },
    enabled: !!userId && pollingInterval > 0,
    refetchInterval: pollingInterval,
    staleTime: Infinity, // Don't mark as stale, let interval control refetch
  });

  // Also poll unread count
  useQuery({
    queryKey: [UNREAD_COUNT_QUERY_KEY, userId, 'polling'],
    queryFn: async () => {
      if (!userId) return null;
      return getUnreadCount({ userId });
    },
    enabled: !!userId && pollingInterval > 0,
    refetchInterval: pollingInterval,
    staleTime: Infinity,
  });
}

/**
 * Hook: Combined hook for notification bell (unread count + recent notifications)
 * @param userId User ID
 * @returns Object with unread count and recent notifications
 */
export function useNotificationBell(userId: string) {
  const unreadCountQuery = useUnreadNotificationCount(userId);
  const recentNotificationsQuery = useUserNotifications(userId, 1, 5);

  return {
    unreadCount: unreadCountQuery.data?.count || 0,
    recentNotifications: recentNotificationsQuery.data?.notifications || [],
    isLoading: unreadCountQuery.isLoading || recentNotificationsQuery.isLoading,
    isError: unreadCountQuery.isError || recentNotificationsQuery.isError,
    error: unreadCountQuery.error || recentNotificationsQuery.error,
  };
}

/**
 * Hook: Quick action handler
 * Combines mark as read and mark action taken
 * @returns Function to execute quick action
 */
export function useQuickActionHandler() {
  const markAsReadMutation = useMarkNotificationAsRead();
  const markActionTakenMutation = useMarkNotificationActionTaken();

  return async (notificationId: string) => {
    try {
      // Mark as read
      await markAsReadMutation.mutateAsync({
        notificationId,
      });

      // Mark action taken
      await markActionTakenMutation.mutateAsync(notificationId);

      return true;
    } catch (error) {
      console.error('Quick action failed:', error);
      return false;
    }
  };
}

/**
 * Hook: Invalidate notification queries (useful after approval/rejection)
 * @returns Function to invalidate queries
 */
export function useInvalidateNotifications() {
  const queryClient = useQueryClient();

  return (userId?: string) => {
    if (userId) {
      queryClient.invalidateQueries({
        queryKey: [NOTIFICATIONS_QUERY_KEY, userId],
      });
      queryClient.invalidateQueries({
        queryKey: [UNREAD_COUNT_QUERY_KEY, userId],
      });
    } else {
      queryClient.invalidateQueries({
        queryKey: [NOTIFICATIONS_QUERY_KEY],
      });
      queryClient.invalidateQueries({
        queryKey: [UNREAD_COUNT_QUERY_KEY],
      });
    }
  };
}

/**
 * Hook: Get notification preferences for a user
 * @param request Request with user ID
 * @returns Query result with preferences
 */
export function useGetNotificationPreferences(
  request: GetNotificationPreferencesRequest
) {
  return useQuery({
    queryKey: [PREFERENCES_QUERY_KEY, request.userId],
    queryFn: async () => getPreferences(request),
    enabled: !!request.userId,
    staleTime: 5 * 60 * 1000, // 5 minutes
  });
}
