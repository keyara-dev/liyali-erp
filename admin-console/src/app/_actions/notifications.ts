"use server";

import authenticatedApiClient, {
  handleError,
  successResponse,
} from "./api-config";

// Types
export interface AdminNotification {
  id: string;
  organization_id: string;
  recipient_id: string;
  type: string;
  subject: string;
  body: string;
  document_id: string;
  document_type: string;
  is_read: boolean;
  read_at: string | null;
  importance: string;
  created_at: string;
  updated_at: string;
}

export interface NotificationStats {
  total: number;
  unread: number;
  read: number;
  today: number;
  by_type: Record<string, number>;
  collected_at: string;
}

export interface NotificationFilters {
  page?: number;
  limit?: number;
  type?: string;
  status?: string;
  search?: string;
}

export interface CreateNotificationRequest {
  subject: string;
  body: string;
  type?: string;
  importance?: string;
  recipient_ids?: string[];
  organization_id?: string;
  broadcast?: boolean;
}

// Get admin notifications with filters
export async function getAdminNotifications(filters?: NotificationFilters) {
  try {
    const params = new URLSearchParams();
    if (filters?.page) params.append("page", String(filters.page));
    if (filters?.limit) params.append("limit", String(filters.limit));
    if (filters?.type) params.append("type", filters.type);
    if (filters?.status) params.append("status", filters.status);
    if (filters?.search) params.append("search", filters.search);

    const response = await authenticatedApiClient({
      url: `/api/v1/admin/notifications?${params.toString()}`,
      method: "GET",
    });

    const data = response?.data?.data || response?.data;
    const meta = response?.data?.meta;
    return successResponse(data, "Notifications retrieved", meta);
  } catch (error) {
    return handleError(error);
  }
}

// Get notification stats
export async function getAdminNotificationStats() {
  try {
    const response = await authenticatedApiClient({
      url: "/api/v1/admin/notifications/stats",
      method: "GET",
    });

    const data = response?.data?.data || response?.data;
    return successResponse(data, "Notification stats retrieved");
  } catch (error) {
    return handleError(error);
  }
}

// Create a notification (broadcast or targeted)
export async function createAdminNotification(
  req: CreateNotificationRequest,
) {
  try {
    const response = await authenticatedApiClient({
      url: "/api/v1/admin/notifications",
      method: "POST",
      data: req,
    });

    const data = response?.data?.data || response?.data;
    return successResponse(data, "Notification created");
  } catch (error) {
    return handleError(error);
  }
}

// Delete a notification
export async function deleteAdminNotification(id: string) {
  try {
    const response = await authenticatedApiClient({
      url: `/api/v1/admin/notifications/${id}`,
      method: "DELETE",
    });

    const data = response?.data?.data || response?.data;
    return successResponse(data, "Notification deleted");
  } catch (error) {
    return handleError(error);
  }
}

// Bulk delete notifications
export async function bulkDeleteAdminNotifications(ids: string[]) {
  try {
    const response = await authenticatedApiClient({
      url: "/api/v1/admin/notifications/bulk-delete",
      method: "POST",
      data: { ids },
    });

    const data = response?.data?.data || response?.data;
    return successResponse(data, "Notifications deleted");
  } catch (error) {
    return handleError(error);
  }
}

// Mark notification as read
export async function markAdminNotificationRead(id: string) {
  try {
    const response = await authenticatedApiClient({
      url: `/api/v1/admin/notifications/${id}/read`,
      method: "POST",
    });

    const data = response?.data?.data || response?.data;
    return successResponse(data, "Notification marked as read");
  } catch (error) {
    return handleError(error);
  }
}
