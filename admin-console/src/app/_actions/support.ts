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

// ============================================================================
// Support — Tickets
// ============================================================================

export interface SupportTicketOrganization {
  id: string;
  name: string;
  slug?: string;
}

export interface SupportTicketUser {
  id: string;
  name: string;
  email: string;
}

export interface SupportTicketAdmin {
  id: string;
  name: string;
  email: string;
}

export interface SupportTicket {
  id: string;
  ticket_number: string;
  organization_id?: string | null;
  organization?: SupportTicketOrganization | null;
  user_id?: string | null;
  user?: SupportTicketUser | null;
  created_by_admin_id?: string | null;
  created_by_admin?: SupportTicketAdmin | null;
  assigned_to_admin_id?: string | null;
  assigned_to_admin?: SupportTicketAdmin | null;
  source: "manual" | "user_app" | "email";
  category: string;
  priority: "low" | "medium" | "high" | "urgent";
  status: "open" | "in_progress" | "waiting_on_customer" | "resolved" | "closed";
  subject: string;
  description: string;
  internal_notes?: string;
  external_reference?: string;
  resolution_summary?: string;
  resolved_at?: string | null;
  closed_at?: string | null;
  created_at: string;
  updated_at: string;
}

export interface SupportTicketFilters {
  organization_id?: string;
  user_id?: string;
  status?: string;
  priority?: string;
  source?: string;
  search?: string;
  page?: number;
  limit?: number;
}

export interface SupportTicketStats {
  total_tickets: number;
  open_tickets: number;
  in_progress_tickets: number;
  waiting_tickets: number;
  resolved_tickets: number;
  closed_tickets: number;
  manual_tickets: number;
  user_app_tickets: number;
  email_tickets: number;
  overdue_tickets: number;
}

export interface CreateSupportTicketRequest {
  organization_id?: string;
  user_id?: string;
  assigned_to_admin_id?: string;
  subject: string;
  description: string;
  category?: string;
  priority?: "low" | "medium" | "high" | "urgent";
  external_reference?: string;
  internal_notes?: string;
}

export interface UpdateSupportTicketRequest {
  organization_id?: string | null;
  user_id?: string | null;
  assigned_to_admin_id?: string | null;
  subject?: string;
  description?: string;
  category?: string;
  priority?: "low" | "medium" | "high" | "urgent";
  status?: "open" | "in_progress" | "waiting_on_customer" | "resolved" | "closed";
  external_reference?: string;
  internal_notes?: string;
  resolution_summary?: string;
}

export async function getSupportTickets(
  filters?: SupportTicketFilters,
): Promise<
  APIResponse<{
    tickets: SupportTicket[];
    total: number;
    page: number;
    limit: number;
    totalPages: number;
  }>
> {
  const params = new URLSearchParams();
  if (filters?.organization_id) params.set("organization_id", filters.organization_id);
  if (filters?.user_id) params.set("user_id", filters.user_id);
  if (filters?.status) params.set("status", filters.status);
  if (filters?.priority) params.set("priority", filters.priority);
  if (filters?.source) params.set("source", filters.source);
  if (filters?.search) params.set("search", filters.search);
  if (filters?.page) params.set("page", String(filters.page));
  if (filters?.limit) params.set("limit", String(filters.limit));

  const query = params.toString();
  const url = `/api/v1/admin/support/tickets${query ? `?${query}` : ""}`;

  try {
    const response = await authenticatedApiClient({ url, method: "GET" });
    return successResponse(
      response?.data?.data || response?.data,
      "Support tickets retrieved successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}

export async function getSupportTicket(
  id: string,
): Promise<APIResponse<SupportTicket>> {
  const url = `/api/v1/admin/support/tickets/${id}`;

  try {
    const response = await authenticatedApiClient({ url, method: "GET" });
    return successResponse(
      response?.data?.data || response?.data,
      "Support ticket retrieved successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}

export async function getSupportTicketStats(): Promise<
  APIResponse<SupportTicketStats>
> {
  const url = "/api/v1/admin/support/tickets/stats";

  try {
    const response = await authenticatedApiClient({ url, method: "GET" });
    return successResponse(
      response?.data?.data || response?.data,
      "Support ticket statistics retrieved successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}

export async function createSupportTicket(
  request: CreateSupportTicketRequest,
): Promise<APIResponse<SupportTicket>> {
  const url = "/api/v1/admin/support/tickets";

  try {
    const response = await authenticatedApiClient({
      url,
      method: "POST",
      data: request,
    });
    return successResponse(
      response?.data?.data || response?.data,
      "Support ticket created successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}

export async function updateSupportTicket(
  id: string,
  request: UpdateSupportTicketRequest,
): Promise<APIResponse<SupportTicket>> {
  const url = `/api/v1/admin/support/tickets/${id}`;

  try {
    const response = await authenticatedApiClient({
      url,
      method: "PUT",
      data: request,
    });
    return successResponse(
      response?.data?.data || response?.data,
      "Support ticket updated successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}
