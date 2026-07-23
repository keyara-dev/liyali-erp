"use server";

import type { APIResponse } from "@/types";
import authenticatedApiClient, {
  handleError,
  successResponse,
} from "./api-config";

export interface ImpersonationLog {
  id: string;
  impersonator_id: string;
  impersonator_email: string;
  target_id: string;
  target_email: string;
  impersonation_type: "platform_user" | "admin_user";
  token_jti: string;
  reason?: string;
  expires_at: string;
  revoked: boolean;
  revoked_at?: string;
  revoked_by?: string;
  created_at: string;
}

export interface ImpersonationStats {
  total: number;
  active: number;
  revoked: number;
  platform_user: number;
  admin_user: number;
  top_impersonators_30d: {
    impersonator_id: string;
    impersonator_email: string;
    count: number;
  }[];
}

export interface ImpersonationLogFilters {
  impersonatorId?: string;
  targetId?: string;
  impersonationType?: "platform_user" | "admin_user";
  revoked?: boolean;
  page?: number;
  limit?: number;
}

export async function getImpersonationLogs(
  filters?: ImpersonationLogFilters,
): Promise<APIResponse<ImpersonationLog[]>> {
  try {
    const params = new URLSearchParams();
    if (filters?.impersonatorId)
      params.set("impersonator_id", filters.impersonatorId);
    if (filters?.targetId) params.set("target_id", filters.targetId);
    if (filters?.impersonationType)
      params.set("impersonation_type", filters.impersonationType);
    if (filters?.revoked !== undefined)
      params.set("revoked", String(filters.revoked));
    if (filters?.page) params.set("page", String(filters.page));
    if (filters?.limit) params.set("limit", String(filters.limit));

    const query = params.toString();
    const result = await authenticatedApiClient({
      url: `/api/v1/admin/impersonation/logs${query ? `?${query}` : ""}`,
      method: "GET",
    });

    if (!result.success) return handleError(result.message);
    return result;
  } catch (error) {
    return handleError(error);
  }
}

export async function getImpersonationLog(
  id: string,
): Promise<APIResponse<ImpersonationLog>> {
  try {
    const result = await authenticatedApiClient({
      url: `/api/v1/admin/impersonation/logs/${id}`,
      method: "GET",
    });
    if (!result.success) return handleError(result.message);
    return successResponse(result.data);
  } catch (error) {
    return handleError(error);
  }
}

export async function revokeImpersonationLog(
  id: string,
): Promise<APIResponse<void>> {
  try {
    const result = await authenticatedApiClient({
      url: `/api/v1/admin/impersonation/logs/${id}/revoke`,
      method: "POST",
    });
    if (!result.success) return handleError(result.message);
    return successResponse(undefined);
  } catch (error) {
    return handleError(error);
  }
}

export async function getImpersonationStats(): Promise<
  APIResponse<ImpersonationStats>
> {
  try {
    const result = await authenticatedApiClient({
      url: "/api/v1/admin/impersonation/stats",
      method: "GET",
    });
    if (!result.success) return handleError(result.message);
    return successResponse(result.data);
  } catch (error) {
    return handleError(error);
  }
}
