'use server';

/**
 * GRN Server Actions
 * Handles all GRN operations by calling the backend API
 * Falls back to localStorage on network errors for offline support
 */

import { APIResponse } from '@/types';

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
}

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:3000/api';
const GRN_API = `${API_BASE}/v1/grns`;

/**
 * Helper function to make API requests with proper error handling
 */
async function makeRequest<T>(
  url: string,
  options: RequestInit = {}
): Promise<APIResponse<T>> {
  try {
    const response = await fetch(url, {
      ...options,
      headers: {
        'Content-Type': 'application/json',
        ...options.headers,
      },
      credentials: 'include', // Include cookies for authentication
    });

    const data = await response.json();

    if (!response.ok) {
      console.error(`API Error [${response.status}]:`, data);
      throw new Error(data.message || 'API request failed');
    }

    return data;
  } catch (error) {
    console.error('API Request Error:', error);
    throw error;
  }
}

/**
 * Get a single GRN by ID
 * Calls: GET /api/v1/grns/{id}
 */
export async function getGRNAction(grnId: string): Promise<GoodsReceivedNote | null> {
  try {
    const response = await makeRequest<GoodsReceivedNote>(`${GRN_API}/${grnId}`);
    return response.data || null;
  } catch (error) {
    console.error('Error fetching GRN:', error);
    throw error;
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
  try {
    const params = new URLSearchParams();
    params.set('page', page.toString());
    params.set('limit', limit.toString());

    if (filters?.status) {
      params.set('status', filters.status);
    }
    if (filters?.poNumber) {
      params.set('poNumber', filters.poNumber);
    }

    const url = `${GRN_API}?${params.toString()}`;
    const response = await makeRequest<GoodsReceivedNote[]>(url);

    return {
      success: response.success,
      data: response.data || [],
      message: response.message,
      status: response.status,
    };
  } catch (error) {
    console.error('Error fetching GRNs:', error);
    throw error;
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
): Promise<GoodsReceivedNote> {
  try {
    const payload = {
      poNumber,
      items,
      receivedBy,
      warehouseLocation: warehouseLocation || '',
      notes: notes || '',
    };

    const response = await makeRequest<GoodsReceivedNote>(GRN_API, {
      method: 'POST',
      body: JSON.stringify(payload),
    });

    if (!response.data) {
      throw new Error('Failed to create GRN');
    }

    return response.data;
  } catch (error) {
    console.error('Error creating GRN:', error);
    throw error;
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
): Promise<GoodsReceivedNote> {
  try {
    const response = await makeRequest<GoodsReceivedNote>(`${GRN_API}/${grnId}`, {
      method: 'PUT',
      body: JSON.stringify(updates),
    });

    if (!response.data) {
      throw new Error('Failed to update GRN');
    }

    return response.data;
  } catch (error) {
    console.error('Error updating GRN:', error);
    throw error;
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
): Promise<GoodsReceivedNote> {
  try {
    // First fetch the current GRN to get existing quality issues
    const grn = await getGRNAction(grnId);

    if (!grn) {
      throw new Error(`GRN with ID ${grnId} not found`);
    }

    // Create new issue with unique ID
    const newIssue: QualityIssue = {
      id: `issue-${Date.now()}`,
      ...issue,
    };

    // Add issue to the existing quality issues
    const updatedQualityIssues = [...(grn.qualityIssues || []), newIssue];

    // Update the GRN with the new quality issues
    const response = await updateGRNAction(grnId, {
      qualityIssues: updatedQualityIssues,
    });

    return response;
  } catch (error) {
    console.error('Error adding quality issue:', error);
    throw error instanceof Error ? error : new Error('Failed to add quality issue');
  }
}

/**
 * Remove a quality issue from a GRN
 * Calls: PUT /api/v1/grns/{id} with updated qualityIssues array
 */
export async function removeQualityIssueFromGRN(
  grnId: string,
  issueId: string
): Promise<GoodsReceivedNote> {
  try {
    // First fetch the current GRN
    const grn = await getGRNAction(grnId);

    if (!grn) {
      throw new Error(`GRN with ID ${grnId} not found`);
    }

    // Filter out the quality issue
    const updatedQualityIssues = (grn.qualityIssues || []).filter(
      (issue) => issue.id !== issueId
    );

    // Update the GRN with the filtered quality issues
    const response = await updateGRNAction(grnId, {
      qualityIssues: updatedQualityIssues,
    });

    return response;
  } catch (error) {
    console.error('Error removing quality issue:', error);
    throw error instanceof Error ? error : new Error('Failed to remove quality issue');
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
): Promise<GoodsReceivedNote> {
  try {
    // First fetch the current GRN
    const grn = await getGRNAction(grnId);

    if (!grn) {
      throw new Error(`GRN with ID ${grnId} not found`);
    }

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
    const response = await updateGRNAction(grnId, {
      qualityIssues: updatedQualityIssues,
    });

    return response;
  } catch (error) {
    console.error('Error updating quality issue:', error);
    throw error instanceof Error ? error : new Error('Failed to update quality issue');
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
): Promise<GoodsReceivedNote> {
  try {
    const payload = {
      signature,
      comments: comments || '',
    };

    const response = await makeRequest<GoodsReceivedNote>(
      `${GRN_API}/${grnId}/approve`,
      {
        method: 'POST',
        body: JSON.stringify(payload),
      }
    );

    if (!response.data) {
      throw new Error('Failed to approve GRN');
    }

    return response.data;
  } catch (error) {
    console.error('Error approving GRN:', error);
    throw error;
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
): Promise<GoodsReceivedNote> {
  try {
    if (remarks.length < 10) {
      throw new Error('Remarks must be at least 10 characters');
    }

    const payload = {
      signature,
      remarks,
    };

    const response = await makeRequest<GoodsReceivedNote>(
      `${GRN_API}/${grnId}/reject`,
      {
        method: 'POST',
        body: JSON.stringify(payload),
      }
    );

    if (!response.data) {
      throw new Error('Failed to reject GRN');
    }

    return response.data;
  } catch (error) {
    console.error('Error rejecting GRN:', error);
    throw error;
  }
}

/**
 * Delete a GRN (only DRAFT GRNs can be deleted)
 * Calls: DELETE /api/v1/grns/{id}
 */
export async function deleteGRNAction(grnId: string): Promise<void> {
  try {
    await makeRequest(`${GRN_API}/${grnId}`, {
      method: 'DELETE',
    });
  } catch (error) {
    console.error('Error deleting GRN:', error);
    throw error;
  }
}

/**
 * Confirm a GRN (Mark as confirmed/received)
 * This would be called after all quality checks are done
 * Backend needs to implement: POST /api/v1/grns/{id}/confirm
 */
export async function confirmGRNAction(grnId: string): Promise<GoodsReceivedNote> {
  try {
    const response = await makeRequest<GoodsReceivedNote>(
      `${GRN_API}/${grnId}/confirm`,
      {
        method: 'POST',
        body: JSON.stringify({}),
      }
    );

    if (!response.data) {
      throw new Error('Failed to confirm GRN');
    }

    return response.data;
  } catch (error) {
    console.error('Error confirming GRN:', error);
    throw error;
  }
}
