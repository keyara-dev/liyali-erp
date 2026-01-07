'use server';

/**
 * GRN Server Actions
 * Handles all GRN operations by calling the backend API using authenticatedApiClient
 * Follows the established pattern from auth.ts and other server actions
 */

import { APIResponse } from '@/types';
import {
  handleError,
  successResponse,
  badRequestResponse,
} from './api-config';
import authenticatedApiClient from './api-config';

interface QualityIssue {
  id: string;
  itemId: string;
  description: string;
  severity: 'LOW' | 'MEDIUM' | 'HIGH';
}

interface GRNItem {
  id: string;
  itemNumber: number;
  description: string;
  poQuantity: number;
  receivedQuantity: number;
  unit: string;
  variance: number;
  damage: number;
  damageNotes?: string;
  condition: 'GOOD' | 'DAMAGED' | 'PARTIAL';
}

export interface GoodsReceivedNote {
  id: string;
  grnNumber: string;
  poNumber: string;
  status: 'DRAFT' | 'SUBMITTED' | 'CONFIRMED' | 'REJECTED' | 'APPROVED';
  warehouseLocation: string;
  receivedDate: string;
  receivedBy: string;
  approvedBy?: string;
  items: GRNItem[];
  qualityIssues: QualityIssue[];
  notes?: string;
  currentStage: number;
  stageName: string;
  createdAt: string;
  updatedAt: string;
  
  // Additional fields for compatibility
  approvalStage: number;        // Maps to currentStage
  approvalHistory: any[];       // Approval history array
  createdBy?: string;           // Creator user ID
  metadata?: {                  // Metadata for UI compatibility
    poId?: string;
    poNumber?: string;
    vendorName?: string;
    amount?: number;
  };
}

/**
 * Get a single GRN by ID
 * Calls: GET /api/v1/grns/{id}
 */
export async function getGRNAction(grnId: string): Promise<APIResponse<GoodsReceivedNote>> {
  const url = `/api/v1/grns/${grnId}`;

  try {
    const response = await authenticatedApiClient({
      method: 'GET',
      url,
    });

    return successResponse(response.data?.data, 'GRN retrieved successfully');
  } catch (error: any) {
    console.error('Error fetching GRN:', error);
    return handleError(error, 'GET', url);
  }
}

/**
 * Get all GRNs with pagination
 * Calls: GET /api/v1/grns?page=1&limit=10&status=DRAFT&poNumber=PO-123
 */
export async function getGRNsAction(
  page: number = 1,
  limit: number = 10,
  filters?: {
    status?: string;
    poNumber?: string;
  }
): Promise<APIResponse<GoodsReceivedNote[]>> {
  const params = new URLSearchParams();
  params.set('page', page.toString());
  params.set('limit', limit.toString());

  if (filters?.status) {
    params.set('status', filters.status);
  }
  if (filters?.poNumber) {
    params.set('poNumber', filters.poNumber);
  }

  const url = `/api/v1/grns?${params.toString()}`;

  try {
    const response = await authenticatedApiClient({
      method: 'GET',
      url,
    });

    return successResponse(response.data?.data || [], 'GRNs fetched successfully');
  } catch (error: any) {
    return handleError(error, 'GET', url);
  }
}

/**
 * Create a new GRN from a Purchase Order
 * Calls: POST /api/v1/grns
 */
export async function createGRNAction(
  poNumber: string,
  items: GRNItem[],
  receivedBy: string,
  warehouseLocation?: string,
  notes?: string
): Promise<APIResponse<GoodsReceivedNote>> {
  const url = `/api/v1/grns`;

  try {
    const payload = {
      poNumber,
      items,
      receivedBy,
      warehouseLocation: warehouseLocation || '',
      notes: notes || '',
    };

    const response = await authenticatedApiClient({
      method: 'POST',
      url,
      data: payload,
    });

    return successResponse(response.data?.data, 'GRN created successfully');
  } catch (error: any) {
    return handleError(error, 'POST', url);
  }
}

/**
 * Update an existing GRN
 * Calls: PUT /api/v1/grns/{id}
 * Can update items and quality issues
 */
export async function updateGRNAction(
  grnId: string,
  updates: {
    items?: GRNItem[];
    receivedBy?: string;
    qualityIssues?: QualityIssue[];
    warehouseLocation?: string;
    notes?: string;
  }
): Promise<APIResponse<GoodsReceivedNote>> {
  const url = `/api/v1/grns/${grnId}`;

  try {
    const response = await authenticatedApiClient({
      method: 'PUT',
      url,
      data: updates,
    });

    return successResponse(response.data?.data, 'GRN updated successfully');
  } catch (error: any) {
    return handleError(error, 'PUT', url);
  }
}

/**
 * Add a quality issue to a GRN
 * Updates the GRN with the new quality issue in the qualityIssues array
 * Calls: PUT /api/v1/grns/{id} with qualityIssues
 */
export async function addQualityIssueToGRN(
  grnId: string,
  issue: Omit<QualityIssue, 'id'>
): Promise<APIResponse<GoodsReceivedNote>> {
  try {
    // First fetch the current GRN to get existing quality issues
    const currentResult = await getGRNAction(grnId);

    if (!currentResult.success || !currentResult.data) {
      return badRequestResponse(`GRN with ID ${grnId} not found`);
    }

    const grn = currentResult.data;

    // Create new issue with unique ID
    const newIssue: QualityIssue = {
      id: `issue-${Date.now()}`,
      ...issue,
    };

    // Add issue to the existing quality issues
    const updatedQualityIssues = [...(grn.qualityIssues || []), newIssue];

    // Update the GRN with the new quality issues
    return await updateGRNAction(grnId, {
      qualityIssues: updatedQualityIssues,
    });
  } catch (error: any) {
    return handleError(error, 'PUT', `/api/v1/grns/${grnId}`);
  }
}

/**
 * Remove a quality issue from a GRN
 * Calls: PUT /api/v1/grns/{id} with updated qualityIssues array
 */
export async function removeQualityIssueFromGRN(
  grnId: string,
  issueId: string
): Promise<APIResponse<GoodsReceivedNote>> {
  try {
    // First fetch the current GRN
    const currentResult = await getGRNAction(grnId);

    if (!currentResult.success || !currentResult.data) {
      return badRequestResponse(`GRN with ID ${grnId} not found`);
    }

    const grn = currentResult.data;

    // Filter out the quality issue
    const updatedQualityIssues = (grn.qualityIssues || []).filter(
      (issue) => issue.id !== issueId
    );

    // Update the GRN with the filtered quality issues
    return await updateGRNAction(grnId, {
      qualityIssues: updatedQualityIssues,
    });
  } catch (error: any) {
    return handleError(error, 'PUT', `/api/v1/grns/${grnId}`);
  }
}

/**
 * Update a quality issue in a GRN
 * Calls: PUT /api/v1/grns/{id} with updated qualityIssues array
 */
export async function updateQualityIssueInGRN(
  grnId: string,
  issueId: string,
  updates: Partial<Omit<QualityIssue, 'id'>>
): Promise<APIResponse<GoodsReceivedNote>> {
  try {
    // First fetch the current GRN
    const currentResult = await getGRNAction(grnId);

    if (!currentResult.success || !currentResult.data) {
      return badRequestResponse(`GRN with ID ${grnId} not found`);
    }

    const grn = currentResult.data;

    // Update the specific quality issue
    const updatedQualityIssues = (grn.qualityIssues || []).map((issue) =>
      issue.id === issueId
        ? {
            ...issue,
            ...updates,
          }
        : issue
    );

    // Update the GRN with the updated quality issues
    return await updateGRNAction(grnId, {
      qualityIssues: updatedQualityIssues,
    });
  } catch (error: any) {
    return handleError(error, 'PUT', `/api/v1/grns/${grnId}`);
  }
}

/**
 * Approve a GRN
 * Calls: POST /api/v1/grns/{id}/approve
 */
export async function approveGRNAction(
  grnId: string,
  signature: string,
  comments?: string
): Promise<APIResponse<GoodsReceivedNote>> {
  const url = `/api/v1/grns/${grnId}/approve`;

  try {
    const payload = {
      signature,
      comments: comments || '',
    };

    const response = await authenticatedApiClient({
      method: 'POST',
      url,
      data: payload,
    });

    return successResponse(response.data?.data, 'GRN approved successfully');
  } catch (error: any) {
    return handleError(error, 'POST', url);
  }
}

/**
 * Reject a GRN
 * Calls: POST /api/v1/grns/{id}/reject
 */
export async function rejectGRNAction(
  grnId: string,
  signature: string,
  remarks: string
): Promise<APIResponse<GoodsReceivedNote>> {
  const url = `/api/v1/grns/${grnId}/reject`;

  try {
    if (remarks.length < 10) {
      return badRequestResponse('Remarks must be at least 10 characters');
    }

    const payload = {
      signature,
      remarks,
    };

    const response = await authenticatedApiClient({
      method: 'POST',
      url,
      data: payload,
    });

    return successResponse(response.data?.data, 'GRN rejected successfully');
  } catch (error: any) {
    return handleError(error, 'POST', url);
  }
}

/**
 * Delete a GRN (only DRAFT GRNs can be deleted)
 * Calls: DELETE /api/v1/grns/{id}
 */
export async function deleteGRNAction(grnId: string): Promise<APIResponse<null>> {
  const url = `/api/v1/grns/${grnId}`;

  try {
    await authenticatedApiClient({
      method: 'DELETE',
      url,
    });

    return successResponse(null, 'GRN deleted successfully');
  } catch (error: any) {
    return handleError(error, 'DELETE', url);
  }
}

/**
 * Confirm a GRN (Mark as confirmed/received)
 * This would be called after all quality checks are done
 * Backend needs to implement: POST /api/v1/grns/{id}/confirm
 */
export async function confirmGRNAction(grnId: string): Promise<APIResponse<GoodsReceivedNote>> {
  const url = `/api/v1/grns/${grnId}/confirm`;

  try {
    const response = await authenticatedApiClient({
      method: 'POST',
      url,
      data: {},
    });

    return successResponse(response.data?.data, 'GRN confirmed successfully');
  } catch (error: any) {
    return handleError(error, 'POST', url);
  }
}
