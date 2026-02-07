"use server";

import type { APIResponse } from "@/types";
import authenticatedApiClient, {
  handleError,
  successResponse,
} from "./api-config";

export interface AuditLog {
  id: string;
  action: string;
  action_type:
    | "create"
    | "update"
    | "delete"
    | "view"
    | "login"
    | "logout"
    | "export"
    | "import"
    | "system";
  user_id: string;
  user_name: string;
  user_email: string;
  organization_id?: string;
  organization_name?: string;
  resource_type: string;
  resource_id?: string;
  details: Record<string, any>;
  metadata: {
    ip_address?: string;
    user_agent?: string;
    location?: string;
    device_type?: string;
    session_id?: string;
  };
  timestamp: string;
  severity: "low" | "medium" | "high" | "critical";
  status: "success" | "failure" | "warning";
  duration_ms?: number;
}

export interface AuditLogFilters {
  date_range?: "1h" | "24h" | "7d" | "30d" | "90d" | "custom";
  start_date?: string;
  end_date?: string;
  user_id?: string;
  organization_id?: string;
  action_type?: string;
  resource_type?: string;
  severity?: string;
  status?: string;
  search?: string;
  ip_address?: string;
}

export interface AuditLogStats {
  total_logs: number;
  logs_today: number;
  failed_actions: number;
  critical_events: number;
  unique_users: number;
  top_actions: Array<{
    action: string;
    count: number;
    percentage: number;
  }>;
  activity_by_hour: Array<{
    hour: string;
    count: number;
    failed_count: number;
  }>;
  security_events: {
    failed_logins: number;
    suspicious_activities: number;
    policy_violations: number;
    unauthorized_access_attempts: number;
  };
}

export interface AuditLogAnalytics {
  user_activity: Array<{
    user_id: string;
    user_name: string;
    action_count: number;
    last_activity: string;
    risk_score: number;
  }>;
  resource_access: Array<{
    resource_type: string;
    access_count: number;
    unique_users: number;
    last_accessed: string;
  }>;
  geographic_distribution: Array<{
    country: string;
    region: string;
    count: number;
    percentage: number;
  }>;
  device_analytics: Array<{
    device_type: string;
    count: number;
    percentage: number;
  }>;
}

/**
 * Get audit logs with filtering and pagination
 */
export async function getAuditLogs(
  filters?: AuditLogFilters,
  page = 1,
  limit = 50,
): Promise<
  APIResponse<{
    logs: AuditLog[];
    total: number;
    page: number;
    totalPages: number;
  }>
> {
  const params = new URLSearchParams();

  params.append("page", page.toString());
  params.append("limit", limit.toString());

  if (filters?.date_range) params.append("date_range", filters.date_range);
  if (filters?.start_date) params.append("start_date", filters.start_date);
  if (filters?.end_date) params.append("end_date", filters.end_date);
  if (filters?.user_id) params.append("user_id", filters.user_id);
  if (filters?.organization_id)
    params.append("organization_id", filters.organization_id);
  if (filters?.action_type) params.append("action_type", filters.action_type);
  if (filters?.resource_type)
    params.append("resource_type", filters.resource_type);
  if (filters?.severity) params.append("severity", filters.severity);
  if (filters?.status) params.append("status", filters.status);
  if (filters?.search) params.append("search", filters.search);
  if (filters?.ip_address) params.append("ip_address", filters.ip_address);

  const url = `/api/v1/admin/audit-logs?${params.toString()}`;

  try {
    const response = await authenticatedApiClient({
      url: url,
      method: "GET",
    });

    return successResponse(
      response?.data?.data || response?.data,
      "Audit logs retrieved successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}

/**
 * Get audit log statistics
 */
export async function getAuditLogStats(
  filters?: AuditLogFilters,
): Promise<APIResponse<AuditLogStats | null>> {
  const params = new URLSearchParams();

  if (filters?.date_range) params.append("date_range", filters.date_range);
  if (filters?.start_date) params.append("start_date", filters.start_date);
  if (filters?.end_date) params.append("end_date", filters.end_date);
  if (filters?.organization_id)
    params.append("organization_id", filters.organization_id);

  const url = `/api/v1/admin/audit-logs/stats${params.toString() ? `?${params.toString()}` : ""}`;

  try {
    const response = await authenticatedApiClient({
      url: url,
      method: "GET",
    });

    return successResponse(
      response?.data?.data || response?.data,
      "Audit log statistics retrieved successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}

/**
 * Get audit log analytics
 */
export async function getAuditLogAnalytics(
  filters?: AuditLogFilters,
): Promise<APIResponse<AuditLogAnalytics | null>> {
  const params = new URLSearchParams();

  if (filters?.date_range) params.append("date_range", filters.date_range);
  if (filters?.start_date) params.append("start_date", filters.start_date);
  if (filters?.end_date) params.append("end_date", filters.end_date);
  if (filters?.organization_id)
    params.append("organization_id", filters.organization_id);

  const url = `/api/v1/admin/audit-logs/analytics${params.toString() ? `?${params.toString()}` : ""}`;

  try {
    const response = await authenticatedApiClient({
      url: url,
      method: "GET",
    });

    return successResponse(
      response?.data?.data || response?.data,
      "Audit log analytics retrieved successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}

/**
 * Get audit log details by ID
 */
export async function getAuditLogDetails(
  logId: string,
): Promise<APIResponse<AuditLog | null>> {
  const url = `/api/v1/admin/audit-logs/${logId}`;

  try {
    const response = await authenticatedApiClient({
      url: url,
      method: "GET",
    });

    return successResponse(
      response?.data?.data || response?.data,
      "Audit log details retrieved successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}

/**
 * Export audit logs
 */
export async function exportAuditLogs(
  format: "csv" | "json" | "pdf",
  filters?: AuditLogFilters,
): Promise<APIResponse<{ download_url: string; expires_at: string }>> {
  const params = new URLSearchParams();

  params.append("format", format);
  if (filters?.date_range) params.append("date_range", filters.date_range);
  if (filters?.start_date) params.append("start_date", filters.start_date);
  if (filters?.end_date) params.append("end_date", filters.end_date);
  if (filters?.user_id) params.append("user_id", filters.user_id);
  if (filters?.organization_id)
    params.append("organization_id", filters.organization_id);
  if (filters?.action_type) params.append("action_type", filters.action_type);
  if (filters?.resource_type)
    params.append("resource_type", filters.resource_type);
  if (filters?.severity) params.append("severity", filters.severity);
  if (filters?.status) params.append("status", filters.status);

  const url = `/api/v1/admin/audit-logs/export?${params.toString()}`;

  try {
    const response = await authenticatedApiClient({
      url: url,
      method: "POST",
    });

    return successResponse(
      response?.data?.data || response?.data,
      "Audit logs export initiated successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}

/**
 * Get security events summary
 */
export async function getSecurityEvents(
  filters?: AuditLogFilters,
): Promise<APIResponse<any>> {
  const params = new URLSearchParams();

  if (filters?.date_range) params.append("date_range", filters.date_range);
  if (filters?.start_date) params.append("start_date", filters.start_date);
  if (filters?.end_date) params.append("end_date", filters.end_date);

  const url = `/api/v1/admin/audit-logs/security-events${params.toString() ? `?${params.toString()}` : ""}`;

  try {
    const response = await authenticatedApiClient({
      url: url,
      method: "GET",
    });

    return successResponse(
      response?.data?.data || response?.data,
      "Security events retrieved successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}

/**
 * Create manual audit log entry
 */
export async function createAuditLog(data: {
  action: string;
  action_type: string;
  resource_type: string;
  resource_id?: string;
  details: Record<string, any>;
  severity?: string;
}): Promise<APIResponse<AuditLog>> {
  const url = "/api/v1/admin/audit-logs";

  try {
    const response = await authenticatedApiClient({
      url: url,
      method: "POST",
      data: data,
    });

    return successResponse(
      response?.data?.data || response?.data,
      "Audit log created successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}

/**
 * Get audit log retention settings
 */
export async function getAuditLogRetentionSettings(): Promise<
  APIResponse<any>
> {
  const url = "/api/v1/admin/audit-logs/retention-settings";

  try {
    const response = await authenticatedApiClient({
      url: url,
      method: "GET",
    });

    return successResponse(
      response?.data?.data || response?.data,
      "Audit log retention settings retrieved successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}

/**
 * Update audit log retention settings
 */
export async function updateAuditLogRetentionSettings(settings: {
  retention_days: number;
  auto_archive: boolean;
  archive_format: string;
  compliance_mode: boolean;
}): Promise<APIResponse<any>> {
  const url = "/api/v1/admin/audit-logs/retention-settings";

  try {
    const response = await authenticatedApiClient({
      url: url,
      method: "PUT",
      data: settings,
    });

    return successResponse(
      response?.data?.data || response?.data,
      "Audit log retention settings updated successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}
