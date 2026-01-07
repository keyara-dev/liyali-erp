/**
 * Notification Server Actions
 *
 * Server actions for notification operations (CRUD, marking read, preferences).
 * These actions are called from client components and handle notification business logic.
 */

'use server';

import {
  NotificationInterface as Notification,
  NotificationPrefs as NotificationPreferences,
  NotificationTypeEnum as NotificationType,
  GetNotificationsRes as GetNotificationsResponse,
  CreateNotificationReq as CreateNotificationRequest,
  CreateNotificationResponse,
  MarkNotificationReadReq as MarkNotificationReadRequest,
  MarkNotificationReadResponse,
  MarkAllNotificationsReadReq as MarkAllNotificationsReadRequest,
  MarkAllNotificationsReadResponse,
  DeleteNotificationReq as DeleteNotificationRequest,
  DeleteNotificationResponse,
  GetUnreadCountReq as GetUnreadCountRequest,
  GetUnreadCountResponse,
  GetNotificationPreferencesReq as GetNotificationPreferencesRequest,
  GetNotificationPreferencesResponse,
  UpdateNotificationPreferencesReq as UpdateNotificationPreferencesRequest,
  UpdateNotificationPreferencesResponse,
} from '@/types';

import {
  createNotification as persistCreateNotification,
  getNotification,
  getUserNotifications,
  getUserUnreadCount,
  markNotificationAsRead,
  markAllNotificationsAsRead,
  deleteNotification as persistDeleteNotification,
  getNotificationPreferences,
  saveNotificationPreferences,
  getUnreadNotifications as persistGetUnreadNotifications,
  markNotificationActionTaken,
} from '@/lib/notification-persistence';

/**
 * Get notifications for current user
 * @param userId User ID
 * @param page Page number
 * @param pageSize Items per page
 * @param filters Optional filters
 * @returns Paginated notifications
 */
export async function getNotifications(
  userId: string,
  page: number = 1,
  pageSize: number = 20,
  filters?: {
    type?: NotificationType;
    isRead?: boolean;
    startDate?: Date;
    endDate?: Date;
  }
): Promise<GetNotificationsResponse> {
  try {
    if (!userId) {
      throw new Error('User ID is required');
    }

    const result = await getUserNotifications(userId, page, pageSize, filters);

    return {
      notifications: result.notifications,
      total: result.total,
      page: result.page,
      pageSize: result.pageSize,
      hasMore: result.hasMore,
    };
  } catch (error) {
    console.error('[getNotifications] Error:', error);
    throw new Error('Failed to fetch notifications');
  }
}

/**
 * Get unread notifications for user
 * @param userId User ID
 * @param limit Number of notifications to return
 * @returns Array of unread notifications
 */
export async function getUnreadNotifications(
  userId: string,
  limit: number = 10
): Promise<Notification[]> {
  try {
    if (!userId) {
      throw new Error('User ID is required');
    }

    return await persistGetUnreadNotifications(userId);
  } catch (error) {
    console.error('[getUnreadNotifications] Error:', error);
    throw new Error('Failed to fetch unread notifications');
  }
}

/**
 * Get unread notification count for user
 * @param request Request with userId
 * @returns Count of unread notifications
 */
export async function getUnreadCount(
  request: GetUnreadCountRequest
): Promise<GetUnreadCountResponse> {
  try {
    if (!request.userId) {
      throw new Error('User ID is required');
    }

    const count = await getUserUnreadCount(request.userId);

    return {
      count,
      userId: request.userId,
    };
  } catch (error) {
    console.error('[getUnreadCount] Error:', error);
    throw new Error('Failed to get unread count');
  }
}

/**
 * Create a new notification
 * @param request Notification creation request
 * @returns Created notification
 */
export async function createNotificationAction(
  request: CreateNotificationRequest
): Promise<CreateNotificationResponse> {
  try {
    if (!request.userId) {
      throw new Error('User ID is required');
    }
    if (!request.type) {
      throw new Error('Notification type is required');
    }
    if (!request.title || !request.message) {
      throw new Error('Title and message are required');
    }

    const notification = await persistCreateNotification({
      userId: request.userId,
      type: request.type,
      title: request.title,
      message: request.message,
      entityId: request.entityId,
      entityType: request.entityType,
      entityNumber: request.entityNumber,
      relatedUserId: request.relatedUserId,
      relatedUserName: request.relatedUserName,
      isRead: false,
      actionTaken: false,
      quickAction: request.quickAction,
      quickActionData: request.quickActionData,
      importance: request.importance || 'MEDIUM',
      rejectionReason: request.rejectionReason,
      reassignmentReason: request.reassignmentReason,
      createdAt: new Date(),
      expiresAt: request.expiresAt,
    });

    return {
      notification,
      success: true,
    };
  } catch (error) {
    console.error('[createNotification] Error:', error);
    throw new Error('Failed to create notification');
  }
}

/**
 * Mark a notification as read
 * @param request Request with notification ID
 * @returns Updated notification
 */
export async function markAsRead(
  request: MarkNotificationReadRequest
): Promise<MarkNotificationReadResponse> {
  try {
    if (!request.notificationId) {
      throw new Error('Notification ID is required');
    }

    const notification = await markNotificationAsRead(request.notificationId);

    if (!notification) {
      throw new Error('Notification not found');
    }

    return {
      notification,
      success: true,
    };
  } catch (error) {
    console.error('[markAsRead] Error:', error);
    throw new Error('Failed to mark notification as read');
  }
}

/**
 * Mark all notifications as read for a user
 * @param request Request with user ID
 * @returns Count of marked notifications
 */
export async function markAllAsRead(
  request: MarkAllNotificationsReadRequest
): Promise<MarkAllNotificationsReadResponse> {
  try {
    if (!request.userId) {
      throw new Error('User ID is required');
    }

    const count = await markAllNotificationsAsRead(request.userId);

    return {
      count,
      success: true,
    };
  } catch (error) {
    console.error('[markAllAsRead] Error:', error);
    throw new Error('Failed to mark all notifications as read');
  }
}

/**
 * Mark a notification action as taken
 * @param notificationId Notification ID
 * @returns Updated notification
 */
export async function markActionTaken(notificationId: string): Promise<Notification | null> {
  try {
    if (!notificationId) {
      throw new Error('Notification ID is required');
    }

    const notification = await markNotificationActionTaken(notificationId);

    if (!notification) {
      throw new Error('Notification not found');
    }

    return notification;
  } catch (error) {
    console.error('[markActionTaken] Error:', error);
    throw new Error('Failed to mark notification action as taken');
  }
}

/**
 * Delete a notification
 * @param request Request with notification ID
 * @returns Success status
 */
export async function deleteNotificationAction(
  request: DeleteNotificationRequest
): Promise<DeleteNotificationResponse> {
  try {
    if (!request.notificationId) {
      throw new Error('Notification ID is required');
    }

    const success = await persistDeleteNotification(request.notificationId);

    if (!success) {
      throw new Error('Notification not found');
    }

    return {
      success: true,
    };
  } catch (error) {
    console.error('[deleteNotification] Error:', error);
    throw new Error('Failed to delete notification');
  }
}

/**
 * Get notification preferences for a user
 * @param request Request with user ID
 * @returns User preferences
 */
export async function getPreferences(
  request: GetNotificationPreferencesRequest
): Promise<GetNotificationPreferencesResponse> {
  try {
    if (!request.userId) {
      throw new Error('User ID is required');
    }

    const preferences = await getNotificationPreferences(request.userId);

    return {
      preferences,
    };
  } catch (error) {
    console.error('[getPreferences] Error:', error);
    throw new Error('Failed to fetch notification preferences');
  }
}

/**
 * Update notification preferences for a user
 * @param request Request with user ID and preferences
 * @returns Updated preferences
 */
export async function updatePreferences(
  request: UpdateNotificationPreferencesRequest
): Promise<UpdateNotificationPreferencesResponse> {
  try {
    if (!request.userId) {
      throw new Error('User ID is required');
    }
    if (!request.preferences) {
      throw new Error('Preferences are required');
    }

    const preferences = await saveNotificationPreferences(
      request.userId,
      request.preferences
    );

    return {
      preferences,
      success: true,
    };
  } catch (error) {
    console.error('[updatePreferences] Error:', error);
    throw new Error('Failed to update notification preferences');
  }
}

/**
 * Trigger helper: Create TASK_ASSIGNED notification
 * Called when a new approval task is assigned
 */
export async function notifyTaskAssigned(
  approverId: string,
  approverName: string,
  entityId: string,
  entityType: string,
  entityNumber: string,
  currentStageName: string
): Promise<Notification> {
  return await persistCreateNotification({
    userId: approverId,
    type: 'TASK_ASSIGNED',
    title: 'New approval task',
    message: `${entityType} #${entityNumber} needs your approval at ${currentStageName} stage`,
    entityId,
    entityType: entityType as any,
    entityNumber,
    isRead: false,
    actionTaken: false,
    quickAction: {
      type: 'REVIEW_AND_APPROVE',
      label: 'Review Now',
      params: { entityId },
    },
    importance: 'HIGH',
    createdAt: new Date(),
  });
}

/**
 * Trigger helper: Create TASK_REASSIGNED notification
 * Called when an approval task is reassigned to a new user
 */
export async function notifyTaskReassigned(
  newApproverId: string,
  newApproverName: string,
  entityId: string,
  entityType: string,
  entityNumber: string,
  reassignedBy: string,
  reassignedByName: string,
  reassignmentReason?: string
): Promise<Notification> {
  return await persistCreateNotification({
    userId: newApproverId,
    type: 'TASK_REASSIGNED',
    title: 'Task reassigned to you',
    message: `${entityType} #${entityNumber} was reassigned to you by ${reassignedByName}`,
    entityId,
    entityType: entityType as any,
    entityNumber,
    relatedUserId: reassignedBy,
    relatedUserName: reassignedByName,
    isRead: false,
    actionTaken: false,
    reassignmentReason,
    quickAction: {
      type: 'REVIEW_AND_APPROVE',
      label: 'Review Now',
      params: { entityId },
    },
    importance: 'HIGH',
    createdAt: new Date(),
  });
}

/**
 * Trigger helper: Create TASK_APPROVED notification
 * Called when an approval task is approved by an approver
 */
export async function notifyTaskApproved(
  createdById: string,
  entityId: string,
  entityType: string,
  entityNumber: string,
  approvedBy: string,
  approvedByName: string
): Promise<Notification> {
  return await persistCreateNotification({
    userId: createdById,
    type: 'TASK_APPROVED',
    title: 'Task approved',
    message: `Your ${entityType} #${entityNumber} was approved by ${approvedByName}`,
    entityId,
    entityType: entityType as any,
    entityNumber,
    relatedUserId: approvedBy,
    relatedUserName: approvedByName,
    isRead: false,
    actionTaken: false,
    quickAction: {
      type: 'VIEW_ONLY',
      label: 'View',
      params: { entityId },
    },
    importance: 'MEDIUM',
    createdAt: new Date(),
  });
}

/**
 * Trigger helper: Create TASK_REJECTED notification
 * Called when an approval task is rejected by an approver
 */
export async function notifyTaskRejected(
  createdById: string,
  entityId: string,
  entityType: string,
  entityNumber: string,
  rejectedBy: string,
  rejectedByName: string,
  rejectionReason: string
): Promise<Notification> {
  return await persistCreateNotification({
    userId: createdById,
    type: 'TASK_REJECTED',
    title: 'Task rejected',
    message: `Your ${entityType} #${entityNumber} was rejected: ${rejectionReason}`,
    entityId,
    entityType: entityType as any,
    entityNumber,
    relatedUserId: rejectedBy,
    relatedUserName: rejectedByName,
    isRead: false,
    actionTaken: false,
    rejectionReason,
    quickAction: {
      type: 'REVISE_AND_RESUBMIT',
      label: 'Revise & Resubmit',
      params: { entityId },
    },
    importance: 'HIGH',
    createdAt: new Date(),
  });
}

/**
 * Trigger helper: Create WORKFLOW_COMPLETE notification
 * Called when an approval workflow is fully completed
 */
export async function notifyWorkflowComplete(
  createdById: string,
  entityId: string,
  entityType: string,
  entityNumber: string,
  finalApprovedBy: string,
  finalApprovedByName: string
): Promise<Notification> {
  return await persistCreateNotification({
    userId: createdById,
    type: 'WORKFLOW_COMPLETE',
    title: 'Approval complete',
    message: `Your ${entityType} #${entityNumber} was fully approved!`,
    entityId,
    entityType: entityType as any,
    entityNumber,
    relatedUserId: finalApprovedBy,
    relatedUserName: finalApprovedByName,
    isRead: false,
    actionTaken: false,
    quickAction: {
      type: 'VIEW_ONLY',
      label: 'View',
      params: { entityId },
    },
    importance: 'MEDIUM',
    createdAt: new Date(),
  });
}
