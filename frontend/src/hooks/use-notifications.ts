"use client";

import React from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";

// Import server actions
import {
  getNotifications,
  getRecentNotifications,
  getNotificationStats,
  markNotificationsAsRead,
  markAllNotificationsAsRead,
  deleteNotification,
  getNotificationPreferences,
  updateNotificationPreferences,
  type Notification,
  type NotificationStats,
  type NotificationPreferences,
} from "@/app/_actions/notifications";

// Re-export types for convenience
export type { Notification, NotificationStats, NotificationPreferences };

// ============================================================================
// HOOKS
// ============================================================================

export function useNotifications(
  params: {
    page?: number;
    limit?: number;
    type?: string;
    unreadOnly?: boolean;
  } = {}
) {
  return useQuery({
    queryKey: ["notifications", params],
    queryFn: () => getNotifications(params),
    staleTime: 30000, // 30 seconds
    refetchInterval: 60000, // Refetch every minute
  });
}

export function useRecentNotifications() {
  return useQuery({
    queryKey: ["notifications", "recent"],
    queryFn: getRecentNotifications,
    staleTime: 15000, // 15 seconds
    refetchInterval: 30000, // Refetch every 30 seconds
  });
}

export function useNotificationStats() {
  return useQuery({
    queryKey: ["notifications", "stats"],
    queryFn: getNotificationStats,
    staleTime: 15000, // 15 seconds
    refetchInterval: 30000, // Refetch every 30 seconds
  });
}

export function useMarkAsRead() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: markNotificationsAsRead,
    onSuccess: () => {
      // Invalidate and refetch notification queries
      queryClient.invalidateQueries({ queryKey: ["notifications"] });
      toast.success("Notifications marked as read");
    },
    onError: (error) => {
      console.error("Failed to mark notifications as read:", error);
      toast.error("Failed to mark notifications as read");
    },
  });
}

export function useMarkAllAsRead() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: markAllNotificationsAsRead,
    onSuccess: () => {
      // Invalidate and refetch notification queries
      queryClient.invalidateQueries({ queryKey: ["notifications"] });
      toast.success("All notifications marked as read");
    },
    onError: (error) => {
      console.error("Failed to mark all notifications as read:", error);
      toast.error("Failed to mark all notifications as read");
    },
  });
}

export function useDeleteNotification() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: deleteNotification,
    onSuccess: () => {
      // Invalidate and refetch notification queries
      queryClient.invalidateQueries({ queryKey: ["notifications"] });
      toast.success("Notification deleted");
    },
    onError: (error) => {
      console.error("Failed to delete notification:", error);
      toast.error("Failed to delete notification");
    },
  });
}

export function useGetNotificationPreferences(params: { userId: string }) {
  return useQuery({
    queryKey: ["notification-preferences", params.userId],
    queryFn: () => getNotificationPreferences(params.userId),
    staleTime: 300000, // 5 minutes
  });
}

export function useUpdateNotificationPreferences() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: updateNotificationPreferences,
    onSuccess: (_, variables) => {
      // Invalidate and refetch notification preferences
      queryClient.invalidateQueries({
        queryKey: ["notification-preferences", variables.userId],
      });
      toast.success("Notification preferences updated");
    },
    onError: (error) => {
      console.error("Failed to update notification preferences:", error);
      toast.error("Failed to update notification preferences");
    },
  });
}

// ============================================================================
// UTILITY FUNCTIONS
// ============================================================================

export function getNotificationIcon(type: string): string {
  switch (type) {
    case "approval_required":
      return "⏳";
    case "document_approved":
      return "✅";
    case "document_rejected":
      return "❌";
    case "assignment":
      return "📋";
    case "status_change":
      return "🔄";
    default:
      return "📢";
  }
}

export function getNotificationColor(importance: string): string {
  switch (importance) {
    case "HIGH":
      return "text-red-600 bg-red-50 border-red-200";
    case "MEDIUM":
      return "text-blue-600 bg-blue-50 border-blue-200";
    case "LOW":
      return "text-gray-600 bg-gray-50 border-gray-200";
    default:
      return "text-gray-600 bg-gray-50 border-gray-200";
  }
}

export function formatNotificationTime(createdAt: string): string {
  const now = new Date();
  const notificationTime = new Date(createdAt);
  const diffInMinutes = Math.floor(
    (now.getTime() - notificationTime.getTime()) / (1000 * 60)
  );

  if (diffInMinutes < 1) {
    return "Just now";
  } else if (diffInMinutes < 60) {
    return `${diffInMinutes}m ago`;
  } else if (diffInMinutes < 1440) {
    const hours = Math.floor(diffInMinutes / 60);
    return `${hours}h ago`;
  } else {
    const days = Math.floor(diffInMinutes / 1440);
    return `${days}d ago`;
  }
}

export function getDocumentUrl(
  documentType: string,
  documentId: string
): string {
  switch (documentType.toLowerCase()) {
    case "requisition":
      return `/requisitions/${documentId}`;
    case "purchase_order":
      return `/purchase-orders/${documentId}`;
    case "payment_voucher":
      return `/payment-vouchers/${documentId}`;
    case "grn":
      return `/grns/${documentId}`;
    case "budget":
      return `/budgets/${documentId}`;
    default:
      return `/documents/${documentId}`;
  }
}

// ============================================================================
// SPECIALIZED HOOKS
// ============================================================================

export function useNotificationBell(userId: string) {
  const recentQuery = useRecentNotifications();
  const statsQuery = useNotificationStats();

  return {
    unreadCount: statsQuery.data?.data?.pending || 0,
    recentNotifications: recentQuery.data?.data || [],
    isLoading: recentQuery.isLoading || statsQuery.isLoading,
    error: recentQuery.error || statsQuery.error,
  };
}

export function useMarkNotificationAsRead() {
  return useMarkAsRead();
}

export function useNotificationPolling(userId: string, intervalMs: number) {
  const queryClient = useQueryClient();

  React.useEffect(() => {
    const interval = setInterval(() => {
      // Refetch recent notifications and stats
      queryClient.invalidateQueries({ queryKey: ["notifications", "recent"] });
      queryClient.invalidateQueries({ queryKey: ["notifications", "stats"] });
    }, intervalMs);

    return () => clearInterval(interval);
  }, [queryClient, intervalMs]);
}
