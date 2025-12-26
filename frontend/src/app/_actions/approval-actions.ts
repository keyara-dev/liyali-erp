'use server';

/**
 * Approval Server Actions - Backend Powered
 * Delegates all operations to approval-workflow.ts which calls real backend APIs
 * These functions are kept for backward compatibility with existing code
 */

import {
  getApprovalTasks as getApprovalTasksFromBackend,
  getApprovalTaskDetail as getApprovalTaskDetailFromBackend,
  approveApprovalTask,
  rejectApprovalTask,
  reassignApprovalTask,
  getApprovalHistory as getApprovalHistoryFromBackend,
} from './approval-workflow';
import { ApprovalTask } from '@/types';

/**
 * Get all pending approval tasks for the current user
 * @param status - Optional status filter (pending, approved, rejected)
 * @returns List of approval tasks from backend API
 */
export async function getApprovalTasks(status?: string): Promise<{
  success: boolean;
  tasks: ApprovalTask[];
  total: number;
  message?: string;
}> {
  try {
    const filters = status ? { status: status as any } : undefined;
    const result = await getApprovalTasksFromBackend(filters, 1, 100);

    if (!result.success) {
      return {
        success: false,
        tasks: [],
        total: 0,
        message: result.message,
      };
    }

    const tasks = result.data || [];
    return {
      success: true,
      tasks,
      total: tasks.length,
    };
  } catch (error) {
    return {
      success: false,
      tasks: [],
      total: 0,
      message: error instanceof Error ? error.message : 'Failed to fetch approval tasks',
    };
  }
}

/**
 * Get detailed approval task information
 * @param taskId - ID of the approval task
 * @returns Detailed approval task with full document context from backend
 */
export async function getApprovalTaskDetail(taskId: string): Promise<{
  success: boolean;
  data?: any;
  message?: string;
}> {
  try {
    if (!taskId) {
      return {
        success: false,
        message: 'Task ID is required',
      };
    }

    const result = await getApprovalTaskDetailFromBackend(taskId);

    if (!result.success) {
      return {
        success: false,
        message: result.message,
      };
    }

    return {
      success: true,
      data: result.data,
    };
  } catch (error) {
    return {
      success: false,
      message: error instanceof Error ? error.message : 'Failed to fetch task detail',
    };
  }
}

/**
 * Get approval statistics
 * Shows counts and summary data from backend
 * @returns Approval statistics
 */
export async function getApprovalStats(): Promise<{
  success: boolean;
  stats?: {
    totalPending: number;
    highPriority: number;
    thisMonth: number;
    overdue: number;
  };
  message?: string;
}> {
  try {
    const result = await getApprovalTasksFromBackend({}, 1, 1000);

    if (!result.success) {
      return {
        success: false,
        message: result.message,
      };
    }

    const tasks = result.data || [];
    const now = new Date();
    const monthAgo = new Date(now.getFullYear(), now.getMonth(), 1);

    const stats = {
      totalPending: tasks.filter((t) => t.status === 'PENDING').length,
      highPriority: tasks.filter((t) => t.status === 'PENDING' && t.priority === 'HIGH').length,
      thisMonth: tasks.filter((t) => new Date(t.createdAt) >= monthAgo).length,
      overdue: tasks.filter((t) => t.status === 'PENDING' && new Date(t.dueDate) < now).length,
    };

    return {
      success: true,
      stats,
    };
  } catch (error) {
    return {
      success: false,
      message: error instanceof Error ? error.message : 'Failed to fetch approval statistics',
    };
  }
}

/**
 * Get approval history for a document
 * @param documentId - ID of the document
 * @param entityType - Type of entity (kept for backward compatibility)
 * @returns Approval history records from backend
 */
export async function getApprovalHistory(documentId: string, entityType?: string): Promise<{
  success: boolean;
  history?: Array<any>;
  message?: string;
}> {
  try {
    if (!documentId) {
      return {
        success: false,
        message: 'Document ID is required',
      };
    }

    const result = await getApprovalHistoryFromBackend(documentId);

    if (!result.success) {
      return {
        success: false,
        message: result.message,
      };
    }

    const history = result.data || [];
    return {
      success: true,
      history: history.map((record) => ({
        id: `${record.id || record.approverId}-${record.approvedAt}`,
        action: record.status,
        actionBy: record.approverId,
        actionAt: record.approvedAt,
        signature: record.signature ? '***SIGNATURE***' : undefined,
        remarks: record.comments,
      })),
    };
  } catch (error) {
    return {
      success: false,
      message: error instanceof Error ? error.message : 'Failed to fetch approval history',
    };
  }
}

/**
 * Approve an approval task
 * @param taskId - ID of the task to approve
 * @param approverId - ID of the approver (backend handles verification)
 * @param signature - Digital signature (base64)
 * @param remarks - Optional approval remarks
 * @returns Approval result from backend
 */
export async function approveTask(
  taskId: string,
  approverId: string,
  signature: string,
  remarks?: string
): Promise<{
  success: boolean;
  message: string;
  nextStage?: string;
  error?: string;
}> {
  try {
    if (!taskId) {
      return { success: false, message: 'Task ID is required', error: 'VALIDATION_ERROR' };
    }
    if (!signature) {
      return { success: false, message: 'Digital signature is required', error: 'SIGNATURE_REQUIRED' };
    }

    const result = await approveApprovalTask({
      taskId,
      comments: remarks || '',
      signature,
      stageNumber: 0,
    });

    if (!result.success) {
      return {
        success: false,
        message: result.message || 'Failed to approve task',
        error: 'APPROVAL_FAILED',
      };
    }

    return {
      success: true,
      message: result.message || 'Task approved successfully',
    };
  } catch (error) {
    console.error('Approval failed:', error);
    return {
      success: false,
      message: 'Approval operation failed',
      error: error instanceof Error ? error.message : 'UNKNOWN_ERROR',
    };
  }
}

/**
 * Reject an approval task
 * @param taskId - ID of the task to reject
 * @param rejectorId - ID of the rejector (backend handles verification)
 * @param signature - Digital signature (base64)
 * @param remarks - Rejection reason (required)
 * @returns Rejection result from backend
 */
export async function rejectTask(
  taskId: string,
  rejectorId: string,
  signature: string,
  remarks: string
): Promise<{
  success: boolean;
  message: string;
  error?: string;
}> {
  try {
    if (!taskId) {
      return { success: false, message: 'Task ID is required', error: 'VALIDATION_ERROR' };
    }
    if (!signature) {
      return { success: false, message: 'Digital signature is required', error: 'SIGNATURE_REQUIRED' };
    }
    if (!remarks || remarks.trim().length === 0) {
      return { success: false, message: 'Rejection reason is required', error: 'REMARKS_REQUIRED' };
    }

    const result = await rejectApprovalTask({
      taskId,
      remarks,
      comments: remarks,
      signature,
      returnTo: 'ORIGINAL_SUBMITTER',
    });

    if (!result.success) {
      return {
        success: false,
        message: result.message || 'Failed to reject task',
        error: 'REJECTION_FAILED',
      };
    }

    return {
      success: true,
      message: result.message || 'Task rejected successfully',
    };
  } catch (error) {
    console.error('Rejection failed:', error);
    return {
      success: false,
      message: 'Rejection operation failed',
      error: error instanceof Error ? error.message : 'UNKNOWN_ERROR',
    };
  }
}

/**
 * Reassign an approval task to a different approver
 * @param taskId - ID of the task to reassign
 * @param reassignedBy - ID of the user doing the reassignment (backend handles verification)
 * @param newApproverId - ID of the new approver
 * @param newApproverName - Name of the new approver (kept for backward compatibility)
 * @param reason - Reason for reassignment
 * @returns Reassignment result from backend
 */
export async function reassignTask(
  taskId: string,
  reassignedBy: string,
  newApproverId: string,
  newApproverName: string,
  reason: string
): Promise<{
  success: boolean;
  message: string;
  error?: string;
}> {
  try {
    if (!taskId) {
      return { success: false, message: 'Task ID is required', error: 'VALIDATION_ERROR' };
    }
    if (!newApproverId) {
      return { success: false, message: 'New approver ID is required', error: 'VALIDATION_ERROR' };
    }
    if (!reason || reason.trim().length === 0) {
      return { success: false, message: 'Reassignment reason is required', error: 'REASON_REQUIRED' };
    }

    const result = await reassignApprovalTask({
      taskId,
      newApproverId,
      reason,
    });

    if (!result.success) {
      return {
        success: false,
        message: result.message || 'Failed to reassign task',
        error: 'REASSIGNMENT_FAILED',
      };
    }

    return {
      success: true,
      message: result.message || 'Task reassigned successfully',
    };
  } catch (error) {
    console.error('Reassignment failed:', error);
    return {
      success: false,
      message: 'Reassignment operation failed',
      error: error instanceof Error ? error.message : 'UNKNOWN_ERROR',
    };
  }
}
