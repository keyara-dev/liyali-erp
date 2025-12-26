'use server';

import {
  Requisition,
  CreateRequisitionRequest,
  UpdateRequisitionRequest,
  SubmitRequisitionRequest,
  ApproveRequisitionRequest,
  RejectRequisitionRequest,
  RequisitionStats,
} from '@/types/requisition';
import { APIResponse } from '@/types';
import {
  handleError,
  successResponse,
  badRequestResponse,
} from './api-config';
import authenticatedApiClient from './api-config';

/**
 * Create a new requisition
 * Calls: POST /api/v1/requisitions
 */
export async function createRequisition(
  data: CreateRequisitionRequest
): Promise<APIResponse<Requisition>> {
  const url = `/api/v1/requisitions`;

  try {
    const response = await authenticatedApiClient({
      method: 'POST',
      url,
      data: {
        title: data.title,
        description: data.description,
        department: data.department,
        departmentId: data.departmentId,
        requiredByDate: data.requiredByDate,
        priority: data.priority,
        items: data.items,
        budgetCode: data.budgetCode,
        costCenter: data.costCenter,
        projectCode: data.projectCode,
        createdBy: data.createdBy,
        createdByName: data.createdByName,
        createdByRole: data.createdByRole,
      },
    });

    return successResponse(response.data?.data, 'Requisition created successfully');
  } catch (error: any) {
    return handleError(error, 'POST', url);
  }
}

/**
 * Get all requisitions with pagination
 * Calls: GET /api/v1/requisitions?page=...&limit=...&status=...
 */
export async function getRequisitions(
  page: number = 1,
  limit: number = 10,
  filters?: {
    status?: string;
    department?: string;
  }
): Promise<APIResponse<Requisition[]>> {
  const params = new URLSearchParams();
  params.set('page', page.toString());
  params.set('limit', limit.toString());

  if (filters?.status) {
    params.set('status', filters.status);
  }
  if (filters?.department) {
    params.set('department', filters.department);
  }

  const url = `/api/v1/requisitions?${params.toString()}`;

  try {
    const response = await authenticatedApiClient({
      method: 'GET',
      url,
    });

    return successResponse(response.data?.data || [], 'Requisitions retrieved successfully');
  } catch (error: any) {
    return handleError(error, 'GET', url);
  }
}

/**
 * Get requisition by ID
 * Calls: GET /api/v1/requisitions/{id}
 */
export async function getRequisitionById(requisitionId: string): Promise<APIResponse<Requisition>> {
  const url = `/api/v1/requisitions/${requisitionId}`;

  try {
    const response = await authenticatedApiClient({
      method: 'GET',
      url,
    });

    return successResponse(response.data?.data, 'Requisition retrieved successfully');
  } catch (error: any) {
    return handleError(error, 'GET', url);
  }
}

/**
 * Update requisition
 * Calls: PUT /api/v1/requisitions/{id}
 */
export async function updateRequisition(
  data: UpdateRequisitionRequest
): Promise<APIResponse<Requisition>> {
  const url = `/api/v1/requisitions/${data.requisitionId}`;

  try {
    const response = await authenticatedApiClient({
      method: 'PUT',
      url,
      data: {
        title: data.title,
        description: data.description,
        requiredByDate: data.requiredByDate,
        priority: data.priority,
        items: data.items,
        budgetCode: data.budgetCode,
        costCenter: data.costCenter,
        projectCode: data.projectCode,
      },
    });

    return successResponse(response.data?.data, 'Requisition updated successfully');
  } catch (error: any) {
    return handleError(error, 'PUT', url);
  }
}

/**
 * Submit requisition for approval
 * Calls: POST /api/v1/requisitions/{id}/submit
 */
export async function submitRequisitionForApproval(
  data: SubmitRequisitionRequest
): Promise<APIResponse<Requisition>> {
  const url = `/api/v1/requisitions/${data.requisitionId}/submit`;

  try {
    const response = await authenticatedApiClient({
      method: 'POST',
      url,
      data: {
        comments: data.comments,
        submittedBy: data.submittedBy,
        submittedByName: data.submittedByName,
        submittedByRole: data.submittedByRole,
      },
    });

    return successResponse(response.data?.data, 'Requisition submitted for approval');
  } catch (error: any) {
    return handleError(error, 'POST', url);
  }
}

/**
 * Approve requisition
 * Calls: POST /api/v1/requisitions/{id}/approve
 */
export async function approveRequisition(
  data: ApproveRequisitionRequest
): Promise<APIResponse<Requisition>> {
  const url = `/api/v1/requisitions/${data.requisitionId}/approve`;

  try {
    const response = await authenticatedApiClient({
      method: 'POST',
      url,
      data: {
        signature: data.signature,
        comments: data.comments,
        stageNumber: data.stageNumber,
        approvingUserId: data.approvingUserId,
        approvingUserName: data.approvingUserName,
        approvingUserRole: data.approvingUserRole,
      },
    });

    return successResponse(
      response.data?.data,
      'Requisition approved and Purchase Order created'
    );
  } catch (error: any) {
    return handleError(error, 'POST', url);
  }
}

/**
 * Reject requisition
 * Calls: POST /api/v1/requisitions/{id}/reject
 */
export async function rejectRequisition(
  data: RejectRequisitionRequest
): Promise<APIResponse<Requisition>> {
  const url = `/api/v1/requisitions/${data.requisitionId}/reject`;

  try {
    const response = await authenticatedApiClient({
      method: 'POST',
      url,
      data: {
        signature: data.signature,
        remarks: data.remarks,
        comments: data.comments,
        rejectingUserId: data.rejectingUserId,
        rejectingUserName: data.rejectingUserName,
        rejectingUserRole: data.rejectingUserRole,
      },
    });

    return successResponse(response.data?.data, 'Requisition rejected and returned to draft');
  } catch (error: any) {
    return handleError(error, 'POST', url);
  }
}

/**
 * Get requisition statistics
 * Calls: GET /api/v1/requisitions/stats
 */
export async function getRequisitionStats(): Promise<APIResponse<RequisitionStats>> {
  const url = `/api/v1/requisitions/stats`;

  try {
    const response = await authenticatedApiClient({
      method: 'GET',
      url,
    });

    return successResponse(response.data?.data, 'Statistics retrieved successfully');
  } catch (error: any) {
    return handleError(error, 'GET', url);
  }
}

/**
 * Delete requisition (DRAFT only)
 * Calls: DELETE /api/v1/requisitions/{id}
 */
export async function deleteRequisition(requisitionId: string): Promise<APIResponse> {
  const url = `/api/v1/requisitions/${requisitionId}`;

  try {
    await authenticatedApiClient({
      method: 'DELETE',
      url,
    });

    return successResponse(null, 'Requisition deleted successfully');
  } catch (error: any) {
    return handleError(error, 'DELETE', url);
  }
}
