import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  getAdminNotifications,
  getAdminNotificationStats,
  createAdminNotification,
  deleteAdminNotification,
  bulkDeleteAdminNotifications,
  markAdminNotificationRead,
  type NotificationFilters,
  type CreateNotificationRequest,
} from "@/app/_actions/notifications";
import { queryKeys } from "@/lib/query-keys";

export function useAdminNotifications(filters?: NotificationFilters) {
  return useQuery({
    queryKey: queryKeys.notifications.list(filters),
    queryFn: async () => {
      const result = await getAdminNotifications(filters);
      if (!result.success) throw new Error(result.message);
      return (result as any).data || [];
    },
    retry: 2,
    retryDelay: 1000,
  });
}

export function useAdminNotificationStats() {
  return useQuery({
    queryKey: queryKeys.notifications.stats(),
    queryFn: async () => {
      const result = await getAdminNotificationStats();
      if (!result.success) throw new Error(result.message);
      return (
        (result as any).data || {
          total: 0,
          unread: 0,
          read: 0,
          today: 0,
          by_type: {},
        }
      );
    },
    retry: 2,
    retryDelay: 1000,
  });
}

export function useCreateAdminNotification() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (req: CreateNotificationRequest) => {
      const result = await createAdminNotification(req);
      if (!result.success) throw new Error(result.message);
      return (result as any).data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.notifications.all });
      queryClient.invalidateQueries({
        queryKey: queryKeys.notifications.stats(),
      });
    },
  });
}

export function useDeleteAdminNotification() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (id: string) => {
      const result = await deleteAdminNotification(id);
      if (!result.success) throw new Error(result.message);
      return (result as any).data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.notifications.all });
      queryClient.invalidateQueries({
        queryKey: queryKeys.notifications.stats(),
      });
    },
  });
}

export function useBulkDeleteAdminNotifications() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (ids: string[]) => {
      const result = await bulkDeleteAdminNotifications(ids);
      if (!result.success) throw new Error(result.message);
      return (result as any).data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.notifications.all });
      queryClient.invalidateQueries({
        queryKey: queryKeys.notifications.stats(),
      });
    },
  });
}

export function useMarkAdminNotificationRead() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (id: string) => {
      const result = await markAdminNotificationRead(id);
      if (!result.success) throw new Error(result.message);
      return (result as any).data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.notifications.all });
      queryClient.invalidateQueries({
        queryKey: queryKeys.notifications.stats(),
      });
    },
  });
}
