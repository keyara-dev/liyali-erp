/**
 * Notification Persistence Layer
 *
 * Handles storage, retrieval, and querying of notifications.
 * Uses in-memory Maps for MVP (ready for database migration to PostgreSQL/MongoDB).
 */

import { ActivityNotification as Notification } from '@/types/activity';

// Define missing types for backward compatibility
interface NotificationPreferences {
  email: boolean;
  push: boolean;
  sms: boolean;
}

type NotificationType = 'approval_required' | 'approved' | 'rejected' | 'assigned';
import { v4 as uuid } from 'uuid';

/**
 * In-memory storage for notifications
 * Key: notificationId, Value: Notification
 */
const notificationStore = new Map<string, Notification>();

/**
 * In-memory storage for user notification preferences
 * Key: userId, Value: NotificationPreferences
 */
const preferencesStore = new Map<string, NotificationPreferences>();

/**
 * Default notification preferences for new users
 */
function getDefaultPreferences(userId: string): any {
  return {
    userId,
    emailNotifications: false,
    pushNotifications: true,
    inAppNotifications: true,
    notifyOn: {
      taskAssigned: true,
      taskReassigned: true,
      taskApproved: true,
      taskRejected: true,
      workflowComplete: true,
      approvalOverdue: true,
      commentsAdded: false,
    },
    groupNotifications: false,
    quietHours: {
      enabled: false,
      startHour: 22,
      endHour: 8,
    },
    createdAt: new Date(),
    updatedAt: new Date(),
  };
}

/**
 * Create a new notification
 * @param notification Partial notification object (id will be generated)
 * @returns Created notification
 */
export async function createNotification(
  notification: Omit<Notification, 'id'>
): Promise<Notification> {
  const id = uuid();
  const fullNotification: Notification = {
    id,
    ...notification,
  };

  notificationStore.set(id, fullNotification);
  return fullNotification;
}

/**
 * Get a notification by ID
 * @param notificationId Notification ID
 * @returns Notification or undefined
 */
export async function getNotification(notificationId: string): Promise<Notification | undefined> {
  return notificationStore.get(notificationId);
}

/**
 * Get all notifications for a user
 * @param userId User ID
 * @param page Page number (1-based)
 * @param pageSize Items per page
 * @param filters Optional filters
 * @returns Array of notifications and metadata
 */
export async function getUserNotifications(
  userId: string,
  page: number = 1,
  pageSize: number = 20,
  filters?: {
    type?: NotificationType;
    isRead?: boolean;
    startDate?: Date;
    endDate?: Date;
  }
): Promise<{
  notifications: Notification[];
  total: number;
  page: number;
  pageSize: number;
  hasMore: boolean;
}> {
  // Filter notifications for this user
  let userNotifications = Array.from(notificationStore.values()).filter(
    (n) => n.recipientId === userId || (n as any).userId === userId
  );

  // Apply optional filters
  if (filters) {
    if (filters.type) {
      userNotifications = userNotifications.filter((n) => n.type === filters.type);
    }
    if (filters.isRead !== undefined) {
      userNotifications = userNotifications.filter((n) => (n as any).isRead === filters.isRead);
    }
    if (filters.startDate) {
      userNotifications = userNotifications.filter((n) => n.createdAt >= filters.startDate!);
    }
    if (filters.endDate) {
      userNotifications = userNotifications.filter((n) => n.createdAt <= filters.endDate!);
    }
  }

  // Sort by creation date (newest first)
  userNotifications.sort((a, b) => b.createdAt.getTime() - a.createdAt.getTime());

  // Paginate
  const total = userNotifications.length;
  const startIndex = (page - 1) * pageSize;
  const endIndex = startIndex + pageSize;
  const paginatedNotifications = userNotifications.slice(startIndex, endIndex);

  return {
    notifications: paginatedNotifications,
    total,
    page,
    pageSize,
    hasMore: endIndex < total,
  };
}

/**
 * Get unread notification count for a user
 * @param userId User ID
 * @returns Count of unread notifications
 */
export async function getUserUnreadCount(userId: string): Promise<number> {
  return Array.from(notificationStore.values()).filter(
    (n) => (n.recipientId === userId || (n as any).userId === userId) && !(n as any).isRead
  ).length;
}

/**
 * Mark a notification as read
 * @param notificationId Notification ID
 * @returns Updated notification
 */
export async function markNotificationAsRead(
  notificationId: string
): Promise<Notification | undefined> {
  const notification = notificationStore.get(notificationId);
  if (!notification) return undefined;

  const updated: Notification = {
    ...notification,
    ...(notification as any).isRead !== undefined ? {} : { isRead: true },
    ...(notification as any).readAt !== undefined ? {} : { readAt: new Date() },
  } as Notification;

  notificationStore.set(notificationId, updated);
  return updated;
}

/**
 * Mark all notifications as read for a user
 * @param userId User ID
 * @returns Count of marked notifications
 */
export async function markAllNotificationsAsRead(userId: string): Promise<number> {
  let count = 0;
  const now = new Date();

  for (const [id, notification] of notificationStore.entries()) {
    const isUserNotification = notification.recipientId === userId || (notification as any).userId === userId;
    const isUnread = !(notification as any).isRead;
    
    if (isUserNotification && isUnread) {
      const updated = {
        ...notification,
        ...(notification as any).isRead !== undefined ? { isRead: true } : {},
        ...(notification as any).readAt !== undefined ? { readAt: now } : {},
      };
      notificationStore.set(id, updated as Notification);
      count++;
    }
  }

  return count;
}

/**
 * Delete a notification
 * @param notificationId Notification ID
 * @returns Success status
 */
export async function deleteNotification(notificationId: string): Promise<boolean> {
  return notificationStore.delete(notificationId);
}

/**
 * Delete notifications older than specified days
 * @param olderThanDays Delete notifications older than this many days
 * @returns Count of deleted notifications
 */
export async function deleteOldNotifications(olderThanDays: number): Promise<number> {
  const cutoffDate = new Date();
  cutoffDate.setDate(cutoffDate.getDate() - olderThanDays);

  let count = 0;
  for (const [id, notification] of notificationStore.entries()) {
    if (notification.createdAt < cutoffDate) {
      notificationStore.delete(id);
      count++;
    }
  }

  return count;
}

/**
 * Delete multiple notifications
 * @param notificationIds Array of notification IDs
 * @returns Count of deleted notifications
 */
export async function deleteNotifications(notificationIds: string[]): Promise<number> {
  let count = 0;
  for (const id of notificationIds) {
    if (notificationStore.delete(id)) {
      count++;
    }
  }
  return count;
}

/**
 * Get notification preferences for a user
 * @param userId User ID
 * @returns Preferences (defaults if not set)
 */
export async function getNotificationPreferences(
  userId: string
): Promise<NotificationPreferences> {
  return preferencesStore.get(userId) || getDefaultPreferences(userId);
}

/**
 * Save notification preferences for a user
 * @param userId User ID
 * @param preferences Preferences to save
 * @returns Saved preferences
 */
export async function saveNotificationPreferences(
  userId: string,
  preferences: Partial<NotificationPreferences>
): Promise<NotificationPreferences> {
  const existing = preferencesStore.get(userId) || getDefaultPreferences(userId);

  const updated: NotificationPreferences = {
    ...existing,
    ...preferences,
    userId, // Ensure userId is correct
    updatedAt: new Date(),
  };

  preferencesStore.set(userId, updated);
  return updated;
}

/**
 * Get notifications grouped by type for a user
 * @param userId User ID
 * @returns Notifications grouped by type
 */
export async function getNotificationsByType(
  userId: string
): Promise<Record<string, Notification[]>> {
  const userNotifications = Array.from(notificationStore.values()).filter(
    (n) => n.recipientId === userId || (n as any).userId === userId
  );

  const grouped: Record<string, Notification[]> = {
    approval_required: [],
    approved: [],
    rejected: [],
    assigned: [],
  };

  for (const notification of userNotifications) {
    if (!grouped[notification.type]) {
      grouped[notification.type] = [];
    }
    grouped[notification.type].push(notification);
  }

  return grouped;
}

/**
 * Get recent notifications for a user (last N)
 * @param userId User ID
 * @param limit Number of notifications to return
 * @returns Array of recent notifications
 */
export async function getRecentNotifications(
  userId: string,
  limit: number = 10
): Promise<Notification[]> {
  const userNotifications = Array.from(notificationStore.values())
    .filter((n) => n.recipientId === userId || (n as any).userId === userId)
    .sort((a, b) => b.createdAt.getTime() - a.createdAt.getTime())
    .slice(0, limit);

  return userNotifications;
}

/**
 * Get unread notifications for a user
 * @param userId User ID
 * @returns Array of unread notifications
 */
export async function getUnreadNotifications(userId: string): Promise<Notification[]> {
  return Array.from(notificationStore.values())
    .filter((n) => (n.recipientId === userId || (n as any).userId === userId) && !(n as any).isRead)
    .sort((a, b) => b.createdAt.getTime() - a.createdAt.getTime());
}

/**
 * Mark notifications as action taken
 * @param notificationId Notification ID
 * @returns Updated notification
 */
export async function markNotificationActionTaken(
  notificationId: string
): Promise<Notification | undefined> {
  const notification = notificationStore.get(notificationId);
  if (!notification) return undefined;

  const updated: Notification = {
    ...notification,
    ...((notification as any).actionTaken !== undefined ? {} : { actionTaken: true }),
    ...((notification as any).actionTakenAt !== undefined ? {} : { actionTakenAt: new Date() }),
  } as Notification;

  notificationStore.set(notificationId, updated);
  return updated;
}

/**
 * Seed sample notifications (for testing/demo)
 */
export async function seedSampleNotifications(): Promise<void> {
  const now = new Date();
  const sampleNotifications: Omit<Notification, 'id'>[] = [
    {
      organizationId: 'org-001',
      recipientId: 'user-001',
      type: 'approval_required',
      documentId: 'req-001',
      documentType: 'requisition',
      subject: 'New approval task',
      body: 'Requisition #REQ-2024-001 needs your approval',
      sent: false,
      createdAt: new Date(now.getTime() - 5 * 60000),
      updatedAt: new Date(now.getTime() - 5 * 60000),
      // Extended fields for UI compatibility
      userId: 'user-001',
      entityId: 'req-001',
      entityType: 'REQUISITION',
      entityNumber: 'REQ-2024-001',
      relatedUserId: 'user-002',
      relatedUserName: 'John Manager',
      isRead: false,
      actionTaken: false,
      importance: 'HIGH',
      quickAction: {
        type: 'REVIEW_AND_APPROVE',
        label: 'Review Now',
        params: { entityId: 'req-001' },
      },
    },
    {
      organizationId: 'org-001',
      recipientId: 'user-001',
      type: 'approved',
      documentId: 'req-001',
      documentType: 'requisition',
      subject: 'Task approved',
      body: 'Your Requisition #REQ-2024-001 was approved by John Manager',
      sent: true,
      sentAt: new Date(now.getTime() - 4 * 60000),
      createdAt: new Date(now.getTime() - 10 * 60000),
      updatedAt: new Date(now.getTime() - 4 * 60000),
      // Extended fields for UI compatibility
      userId: 'user-001',
      entityId: 'req-001',
      entityType: 'REQUISITION',
      entityNumber: 'REQ-2024-001',
      relatedUserId: 'user-002',
      relatedUserName: 'John Manager',
      isRead: true,
      readAt: new Date(now.getTime() - 4 * 60000),
      actionTaken: false,
      importance: 'MEDIUM',
      quickAction: {
        type: 'VIEW_ONLY',
        label: 'View',
        params: { entityId: 'req-001' },
      },
    },
    {
      organizationId: 'org-001',
      recipientId: 'user-003',
      type: 'assigned',
      documentId: 'req-002',
      documentType: 'requisition',
      subject: 'Task reassigned to you',
      body: 'You were assigned: Requisition #REQ-2024-002 (reassigned by Admin)',
      sent: false,
      createdAt: new Date(now.getTime() - 15 * 60000),
      updatedAt: new Date(now.getTime() - 15 * 60000),
      // Extended fields for UI compatibility
      userId: 'user-003',
      entityId: 'req-002',
      entityType: 'REQUISITION',
      entityNumber: 'REQ-2024-002',
      relatedUserId: 'admin-001',
      relatedUserName: 'System Admin',
      isRead: false,
      actionTaken: false,
      reassignmentReason: 'Original approver out sick',
      importance: 'HIGH',
      quickAction: {
        type: 'REVIEW_AND_APPROVE',
        label: 'Review Now',
        params: { entityId: 'req-002' },
      },
    },
  ];

  for (const notification of sampleNotifications) {
    await createNotification(notification);
  }
}

/**
 * Clear all notifications (for testing)
 */
export async function clearNotifications(): Promise<void> {
  notificationStore.clear();
}

/**
 * Clear all preferences (for testing)
 */
export async function clearPreferences(): Promise<void> {
  preferencesStore.clear();
}

/**
 * Get store state (for debugging)
 */
export function getStoreState() {
  return {
    notificationCount: notificationStore.size,
    preferencesCount: preferencesStore.size,
    notifications: Array.from(notificationStore.values()),
    preferences: Array.from(preferencesStore.values()),
  };
}

/**
 * Get all notifications (for admin/debug)
 */
export async function getAllNotifications(): Promise<Notification[]> {
  return Array.from(notificationStore.values()).sort(
    (a, b) => b.createdAt.getTime() - a.createdAt.getTime()
  );
}
