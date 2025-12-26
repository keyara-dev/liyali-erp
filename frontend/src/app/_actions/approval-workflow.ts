'use server';

import {
  ApprovalTask,
  ApprovalHistory,
  ApproveTaskRequest,
  RejectTaskRequest,
  ReassignTaskRequest,
} from '@/types/workflow';
import { APIResponse } from '@/types';
import {
  handleError,
  successResponse,
} from './api-config';
import authenticatedApiClient from './api-config';

/**
 * Get all approval tasks with pagination and filtering
 * Calls: GET /api/v1/approvals?page=...&limit=...&status=...&document_type=...&assigned_to_me=...
 */
export async function getApprovalTasks(
  filters?: {
    status?: 'PENDING' | 'APPROVED' | 'REJECTED' | 'CANCELLED';
    documentType?: string;
    assignedToMe?: boolean;
  },
  page: number = 1,
  limit: number = 10
): Promise<APIResponse<ApprovalTask[]>> {
  const params = new URLSearchParams();
  params.set('page', page.toString());
  params.set('limit', limit.toString());

  if (filters?.status) {
    params.set('status', filters.status);
  }
  if (filters?.documentType) {
    params.set('document_type', filters.documentType);
  }
  if (filters?.assignedToMe) {
    params.set('assigned_to_me', 'true');
  }

  const url = `/api/v1/approvals?${params.toString()}`;

  try {
    const response = await authenticatedApiClient({
      method: 'GET',
      url,
    });

    return successResponse(response.data?.data || [], 'Approval tasks retrieved successfully');
  } catch (error: any) {
    return handleError(error, 'GET', url);
  }
}

/**
 * Get single approval task with full details
 * Calls: GET /api/v1/approvals/{id}
 */
export async function getApprovalTaskDetail(
  taskId: string
): Promise<APIResponse<any>> {
  const url = `/api/v1/approvals/${taskId}`;

  try {
    const response = await authenticatedApiClient({
      method: 'GET',
      url,
    });

    return successResponse(response.data?.data, 'Approval task retrieved successfully');
  } catch (error: any) {
    return handleError(error, 'GET', url);
  }
}

/**
 * Approve an approval task
 * Calls: POST /api/v1/approvals/{id}/approve
 */
export async function approveApprovalTask(
  data: ApproveTaskRequest
): Promise<APIResponse<any>> {
  const url = `/api/v1/approvals/${data.taskId}/approve`;

  try {
    const response = await authenticatedApiClient({
      method: 'POST',
      url,
      data: {
        comments: data.comments,
        signature: data.signature,
        stageNumber: data.stageNumber,
      },
    });

    return successResponse(response.data?.data, 'Task approved successfully');
  } catch (error: any) {
    return handleError(error, 'POST', url);
  }
}

/**
 * Reject an approval task
 * Calls: POST /api/v1/approvals/{id}/reject
 */
export async function rejectApprovalTask(
  data: RejectTaskRequest
): Promise<APIResponse<any>> {
  const url = `/api/v1/approvals/${data.taskId}/reject`;

  try {
    const response = await authenticatedApiClient({
      method: 'POST',
      url,
      data: {
        remarks: data.remarks,
        comments: data.comments,
        signature: data.signature,
        returnTo: data.returnTo,
      },
    });

    return successResponse(response.data?.data, 'Task rejected successfully');
  } catch (error: any) {
    return handleError(error, 'POST', url);
  }
}

/**
 * Reassign an approval task
 * Calls: POST /api/v1/approvals/{id}/reassign
 */
export async function reassignApprovalTask(
  data: ReassignTaskRequest
): Promise<APIResponse<any>> {
  const url = `/api/v1/approvals/${data.taskId}/reassign`;

  try {
    const response = await authenticatedApiClient({
      method: 'POST',
      url,
      data: {
        newApproverId: data.newApproverId,
        reason: data.reason,
      },
    });

    return successResponse(response.data?.data, 'Task reassigned successfully');
  } catch (error: any) {
    return handleError(error, 'POST', url);
  }
}

/**
 * Get approval history for a document
 * Calls: GET /api/v1/documents/{documentId}/approval-history
 */
export async function getApprovalHistory(
  documentId: string
): Promise<APIResponse<ApprovalHistory[]>> {
  const url = `/api/v1/documents/${documentId}/approval-history`;

  try {
    const response = await authenticatedApiClient({
      method: 'GET',
      url,
    });

    return successResponse(response.data?.data || [], 'Approval history retrieved successfully');
  } catch (error: any) {
    return handleError(error, 'GET', url);
  }
}
