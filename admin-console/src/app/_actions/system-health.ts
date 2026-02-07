"use server";

import type { APIResponse } from "@/types";
import authenticatedApiClient, {
  handleError,
  successResponse,
} from "./api-config";

export interface SystemHealth {
  overall_status: "healthy" | "warning" | "critical";
  uptime_percentage: number;
  uptime_duration: string;
  last_updated: string;
  database: {
    status: "healthy" | "warning" | "critical";
    connection_count: number;
    query_performance: number;
    storage_usage: number;
    last_backup: string;
  };
  api: {
    status: "healthy" | "warning" | "critical";
    response_time: number;
    error_rate: number;
    requests_per_minute: number;
    active_sessions: number;
  };
  cache: {
    status: "healthy" | "warning" | "critical";
    hit_rate: number;
    memory_usage: number;
    eviction_rate: number;
  };
  queue: {
    status: "healthy" | "warning" | "critical";
    pending_jobs: number;
    failed_jobs: number;
    processing_rate: number;
  };
}

export interface SystemMetrics {
  timestamp: string;
  average_response_time: number;
  response_time_trend: "up" | "down" | "stable";
  server: {
    cpu_usage: number;
    memory_usage: number;
    disk_usage: number;
    load_average: string;
    active_connections: number;
  };
  database: {
    active_connections: number;
    slow_queries: number;
    cache_hit_ratio: number;
    storage_size: string;
    backup_status: "success" | "failed" | "in_progress";
  };
  api: {
    total_requests: number;
    successful_requests: number;
    failed_requests: number;
    average_response_time: number;
    peak_response_time: number;
  };
  performance_history: Array<{
    timestamp: string;
    cpu_usage: number;
    memory_usage: number;
    response_time: number;
    requests_per_second: number;
  }>;
}

export interface SystemAlert {
  id: string;
  title: string;
  description: string;
  severity: "low" | "medium" | "high" | "critical";
  status: "active" | "acknowledged" | "resolved";
  component: "database" | "api" | "server" | "cache" | "queue" | "general";
  created_at: string;
  updated_at: string;
  acknowledged_by?: string;
  acknowledged_at?: string;
  resolved_at?: string;
  metadata?: Record<string, any>;
}

export interface SystemLog {
  id: string;
  timestamp: string;
  level: "debug" | "info" | "warn" | "error" | "fatal";
  component: string;
  message: string;
  metadata?: Record<string, any>;
  user_id?: string;
  request_id?: string;
}

export interface PerformanceMetric {
  metric_name: string;
  current_value: number;
  previous_value: number;
  change_percentage: number;
  trend: "up" | "down" | "stable";
  threshold_warning: number;
  threshold_critical: number;
  unit: string;
}

/**
 * Get overall system health status
 */
export async function getSystemHealth(): Promise<
  APIResponse<SystemHealth | null>
> {
  const url = "/api/v1/admin/system/health";

  try {
    const response = await authenticatedApiClient({
      url: url,
      method: "GET",
    });

    return successResponse(
      response?.data?.data || response?.data,
      "System health retrieved successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}

/**
 * Get detailed system metrics
 */
export async function getSystemMetrics(): Promise<
  APIResponse<SystemMetrics | null>
> {
  const url = "/api/v1/admin/system/metrics";

  try {
    const response = await authenticatedApiClient({
      url: url,
      method: "GET",
    });

    return successResponse(
      response?.data?.data || response?.data,
      "System metrics retrieved successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}

/**
 * Get system alerts
 */
export async function getSystemAlerts(
  status?: "active" | "acknowledged" | "resolved",
  severity?: "low" | "medium" | "high" | "critical",
): Promise<APIResponse<SystemAlert[]>> {
  const params = new URLSearchParams();
  if (status) params.append("status", status);
  if (severity) params.append("severity", severity);

  const url = `/api/v1/admin/system/alerts${params.toString() ? `?${params.toString()}` : ""}`;

  try {
    const response = await authenticatedApiClient({
      url: url,
      method: "GET",
    });

    return successResponse(
      response?.data?.data || response?.data || [],
      "System alerts retrieved successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}

/**
 * Acknowledge a system alert
 */
export async function acknowledgeAlert(
  alertId: string,
): Promise<APIResponse<null>> {
  const url = `/api/v1/admin/system/alerts/${alertId}/acknowledge`;

  try {
    await authenticatedApiClient({
      url: url,
      method: "POST",
    });

    return successResponse(null, "Alert acknowledged successfully");
  } catch (error: Error | any) {
    return handleError(error);
  }
}

/**
 * Resolve a system alert
 */
export async function resolveAlert(
  alertId: string,
): Promise<APIResponse<null>> {
  const url = `/api/v1/admin/system/alerts/${alertId}/resolve`;

  try {
    await authenticatedApiClient({
      url: url,
      method: "POST",
    });

    return successResponse(null, "Alert resolved successfully");
  } catch (error: Error | any) {
    return handleError(error);
  }
}

/**
 * Get system logs
 */
export async function getSystemLogs(
  level?: "debug" | "info" | "warn" | "error" | "fatal",
  component?: string,
  page: number = 1,
  limit: number = 100,
): Promise<
  APIResponse<{
    logs: SystemLog[];
    total: number;
    page: number;
    limit: number;
  }>
> {
  const params = new URLSearchParams();
  if (level) params.append("level", level);
  if (component) params.append("component", component);
  params.append("page", page.toString());
  params.append("limit", limit.toString());

  const url = `/api/v1/admin/system/logs?${params.toString()}`;

  try {
    const response = await authenticatedApiClient({
      url: url,
      method: "GET",
    });

    return successResponse(
      response?.data?.data || response?.data,
      "System logs retrieved successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}

/**
 * Get performance metrics
 */
export async function getPerformanceMetrics(): Promise<
  APIResponse<PerformanceMetric[]>
> {
  const url = "/api/v1/admin/system/performance";

  try {
    const response = await authenticatedApiClient({
      url: url,
      method: "GET",
    });

    return successResponse(
      response?.data?.data || response?.data || [],
      "Performance metrics retrieved successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}

/**
 * Run system health check
 */
export async function runHealthCheck(): Promise<
  APIResponse<SystemHealth | null>
> {
  const url = "/api/v1/admin/system/health/check";

  try {
    const response = await authenticatedApiClient({
      url: url,
      method: "POST",
    });

    return successResponse(
      response?.data?.data || response?.data,
      "Health check completed successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}

/**
 * Get system configuration
 */
export async function getSystemConfiguration(): Promise<APIResponse<any>> {
  const url = "/api/v1/admin/system/config";

  try {
    const response = await authenticatedApiClient({
      url: url,
      method: "GET",
    });

    return successResponse(
      response?.data?.data || response?.data,
      "System configuration retrieved successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}

/**
 * Update system configuration
 */
export async function updateSystemConfiguration(
  config: Record<string, any>,
): Promise<APIResponse<any>> {
  const url = "/api/v1/admin/system/config";

  try {
    const response = await authenticatedApiClient({
      url: url,
      method: "PUT",
      data: config,
    });

    return successResponse(
      response?.data?.data || response?.data,
      "System configuration updated successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}

/**
 * Restart system service
 */
export async function restartService(
  serviceName: string,
): Promise<APIResponse<null>> {
  const url = `/api/v1/admin/system/services/${serviceName}/restart`;

  try {
    await authenticatedApiClient({
      url: url,
      method: "POST",
    });

    return successResponse(
      null,
      `Service ${serviceName} restarted successfully`,
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}

/**
 * Clear system cache
 */
export async function clearSystemCache(): Promise<APIResponse<null>> {
  const url = "/api/v1/admin/system/cache/clear";

  try {
    await authenticatedApiClient({
      url: url,
      method: "POST",
    });

    return successResponse(null, "System cache cleared successfully");
  } catch (error: Error | any) {
    return handleError(error);
  }
}
