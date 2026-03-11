"use server";

import type { APIResponse } from "@/types";
import authenticatedApiClient, {
  handleError,
  successResponse,
} from "./api-config";

// ============================================================================
// Support — Documents
// ============================================================================

export interface SupportDocumentParams {
  orgId?: string;
  userId?: string;
  type?: string;
  status?: string;
  search?: string;
  page?: number;
  limit?: number;
}

/**
 * Get platform-wide documents for support diagnosis
 */
export async function getSupportDocuments(
  params?: SupportDocumentParams,
): Promise<APIResponse<any[]>> {
  const searchParams = new URLSearchParams();
  if (params?.orgId) searchParams.set("org_id", params.orgId);
  if (params?.userId) searchParams.set("user_id", params.userId);
  if (params?.type) searchParams.set("type", params.type);
  if (params?.status) searchParams.set("status", params.status);
  if (params?.search) searchParams.set("search", params.search);
  if (params?.page) searchParams.set("page", String(params.page));
  if (params?.limit) searchParams.set("limit", String(params.limit));

  const query = searchParams.toString();
  const url = `/api/v1/admin/support/documents${query ? `?${query}` : ""}`;

  try {
    const response = await authenticatedApiClient({ url, method: "GET" });
    return successResponse(
      response?.data?.data || response?.data,
      "Documents retrieved successfully",
      response?.data?.pagination,
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}

/**
 * Get a single document by ID for support diagnosis
 */
export async function getSupportDocument(id: string): Promise<APIResponse<any>> {
  const url = `/api/v1/admin/support/documents/${id}`;

  try {
    const response = await authenticatedApiClient({ url, method: "GET" });
    return successResponse(
      response?.data?.data || response?.data,
      "Document retrieved successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}

// ============================================================================
// Support — Workflow Tasks
// ============================================================================

export interface SupportWorkflowTaskParams {
  orgId?: string;
  entityId?: string;
  status?: string;
  stalled?: boolean;
  page?: number;
  limit?: number;
}

/**
 * Get platform-wide workflow tasks for support
 */
export async function getSupportWorkflowTasks(
  params?: SupportWorkflowTaskParams,
): Promise<APIResponse<any[]>> {
  const searchParams = new URLSearchParams();
  if (params?.orgId) searchParams.set("org_id", params.orgId);
  if (params?.entityId) searchParams.set("entity_id", params.entityId);
  if (params?.status) searchParams.set("status", params.status);
  if (params?.stalled) searchParams.set("stalled", "true");
  if (params?.page) searchParams.set("page", String(params.page));
  if (params?.limit) searchParams.set("limit", String(params.limit));

  const query = searchParams.toString();
  const url = `/api/v1/admin/support/workflow-tasks${query ? `?${query}` : ""}`;

  try {
    const response = await authenticatedApiClient({ url, method: "GET" });
    return successResponse(
      response?.data?.data || response?.data,
      "Workflow tasks retrieved successfully",
      response?.data?.pagination,
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}

/**
 * Get a single workflow task with full details
 */
export async function getSupportWorkflowTask(
  id: string,
): Promise<APIResponse<any>> {
  const url = `/api/v1/admin/support/workflow-tasks/${id}`;

  try {
    const response = await authenticatedApiClient({ url, method: "GET" });
    return successResponse(
      response?.data?.data || response?.data,
      "Workflow task retrieved successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}

/**
 * Reassign a stuck/claimed workflow task to a different user
 */
export async function reassignWorkflowTask(
  taskId: string,
  data: { newAssigneeId: string; reason: string },
): Promise<APIResponse<any>> {
  const url = `/api/v1/admin/support/workflow-tasks/${taskId}/reassign`;

  try {
    const response = await authenticatedApiClient({
      url,
      method: "POST",
      data: { new_assignee_id: data.newAssigneeId, reason: data.reason },
    });
    return successResponse(
      response?.data?.data || response?.data,
      "Workflow task reassigned successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}

/**
 * Reset a stuck/expired claimed task back to pending
 */
export async function resetWorkflowTask(
  taskId: string,
  data: { reason: string },
): Promise<APIResponse<any>> {
  const url = `/api/v1/admin/support/workflow-tasks/${taskId}/reset`;

  try {
    const response = await authenticatedApiClient({
      url,
      method: "POST",
      data,
    });
    return successResponse(
      response?.data?.data || response?.data,
      "Workflow task reset successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}
