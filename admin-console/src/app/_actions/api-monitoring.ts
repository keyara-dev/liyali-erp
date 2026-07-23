"use server";

import type { APIResponse } from "@/types";
import authenticatedApiClient, {
  handleError,
  successResponse,
} from "./api-config";

export interface APIEndpoint {
  id: string;
  path: string;
  method: string;
  name: string;
  description: string;
  category: string;
  is_public: boolean;
  is_deprecated: boolean;
  version: string;
  rate_limit: number;
  timeout: number;
  created_at: string;
  updated_at: string;
}

export interface APIMetrics {
  endpoint_id: string;
  endpoint_path: string;
  method: string;
  total_requests: number;
  successful_requests: number;
  failed_requests: number;
  avg_response_time: number;
  min_response_time: number;
  max_response_time: number;
  p95_response_time: number;
  p99_response_time: number;
  error_rate: number;
  success_rate: number;
  requests_per_minute: number;
  last_request_at: string;
  period_start: string;
  period_end: string;
}

export interface APIError {
  id: string;
  endpoint_id: string;
  endpoint_path: string;
  method: string;
  status_code: number;
  error_message: string;
  error_type: string;
  request_id: string;
  user_id?: string;
  ip_address: string;
  user_agent: string;
  request_body?: string;
  response_body?: string;
  response_time: number;
  occurred_at: string;
  resolved_at?: string;
  is_resolved: boolean;
}

export interface APIAlert {
  id: string;
  endpoint_id?: string;
  alert_type: string;
  severity: "low" | "medium" | "high" | "critical";
  title: string;
  description: string;
  threshold_value: number;
  current_value: number;
  is_active: boolean;
  triggered_at: string;
  resolved_at?: string;
  acknowledged_at?: string;
  acknowledged_by?: string;
  notification_sent: boolean;
}

export interface APIFilters {
  search?: string;
  method?: string;
  category?: string;
  is_public?: boolean;
  is_deprecated?: boolean;
  status_code?: number;
  error_type?: string;
  severity?: string;
  time_range?: string;
  start_date?: string;
  end_date?: string;
}

export interface APIStats {
  total_endpoints: number;
  active_endpoints: number;
  deprecated_endpoints: number;
  public_endpoints: number;
  private_endpoints: number;
  total_requests_today: number;
  total_errors_today: number;
  avg_response_time_today: number;
  error_rate_today: number;
  uptime_percentage: number;
  active_alerts: number;
  critical_alerts: number;
  endpoints_by_category: Array<{
    category: string;
    count: number;
    percentage: number;
  }>;
  requests_by_method: Array<{
    method: string;
    count: number;
    percentage: number;
  }>;
  top_endpoints: Array<{
    endpoint_id: string;
    path: string;
    method: string;
    request_count: number;
    avg_response_time: number;
    error_rate: number;
  }>;
  slowest_endpoints: Array<{
    endpoint_id: string;
    path: string;
    method: string;
    avg_response_time: number;
    p95_response_time: number;
  }>;
  error_distribution: Array<{
    status_code: number;
    count: number;
    percentage: number;
  }>;
}

export interface APIPerformanceData {
  timestamp: string;
  requests_per_minute: number;
  avg_response_time: number;
  error_rate: number;
  active_connections: number;
  cpu_usage: number;
  memory_usage: number;
}

/**
 * Get all API endpoints with filtering
 */
export async function getAPIEndpoints(
  filters?: APIFilters,
): Promise<APIResponse<APIEndpoint[]>> {
  const params = new URLSearchParams();

  if (filters?.search) params.append("search", filters.search);
  if (filters?.method) params.append("method", filters.method);
  if (filters?.category) params.append("category", filters.category);
  if (filters?.is_public !== undefined)
    params.append("is_public", filters.is_public.toString());
  if (filters?.is_deprecated !== undefined)
    params.append("is_deprecated", filters.is_deprecated.toString());

  const url = `/api/v1/admin/api-monitoring/endpoints${params.toString() ? `?${params.toString()}` : ""}`;

  try {
    const response = await authenticatedApiClient({
      url: url,
      method: "GET",
    });

    return successResponse(
      response?.data?.data || response?.data,
      "API endpoints retrieved successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}

/**
 * Get API endpoint by ID
 */
export async function getAPIEndpoint(
  endpointId: string,
): Promise<APIResponse<APIEndpoint | null>> {
  const url = `/api/v1/admin/api-monitoring/endpoints/${endpointId}`;

  try {
    const response = await authenticatedApiClient({
      url: url,
      method: "GET",
    });

    return successResponse(
      response?.data?.data || response?.data,
      "API endpoint retrieved successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}

/**
 * Get API metrics for endpoints
 */
export async function getAPIMetrics(
  filters?: APIFilters,
): Promise<APIResponse<APIMetrics[]>> {
  const params = new URLSearchParams();

  if (filters?.time_range) params.append("time_range", filters.time_range);
  if (filters?.start_date) params.append("start_date", filters.start_date);
  if (filters?.end_date) params.append("end_date", filters.end_date);
  if (filters?.method) params.append("method", filters.method);
  if (filters?.category) params.append("category", filters.category);

  const url = `/api/v1/admin/api-monitoring/metrics${params.toString() ? `?${params.toString()}` : ""}`;

  try {
    const response = await authenticatedApiClient({
      url: url,
      method: "GET",
    });

    return successResponse(
      response?.data?.data || response?.data,
      "API metrics retrieved successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}

/**
 * Get API metrics for specific endpoint
 */
export async function getEndpointMetrics(
  endpointId: string,
  timeRange: string = "24h",
): Promise<APIResponse<APIMetrics | null>> {
  const url = `/api/v1/admin/api-monitoring/endpoints/${endpointId}/metrics?time_range=${timeRange}`;

  try {
    const response = await authenticatedApiClient({
      url: url,
      method: "GET",
    });

    return successResponse(
      response?.data?.data || response?.data,
      "Endpoint metrics retrieved successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}

/**
 * Get API errors with filtering
 */
export async function getAPIErrors(
  filters?: APIFilters,
): Promise<APIResponse<APIError[]>> {
  const params = new URLSearchParams();

  if (filters?.search) params.append("search", filters.search);
  if (filters?.method) params.append("method", filters.method);
  if (filters?.status_code)
    params.append("status_code", filters.status_code.toString());
  if (filters?.error_type) params.append("error_type", filters.error_type);
  if (filters?.time_range) params.append("time_range", filters.time_range);
  if (filters?.start_date) params.append("start_date", filters.start_date);
  if (filters?.end_date) params.append("end_date", filters.end_date);

  const url = `/api/v1/admin/api-monitoring/errors${params.toString() ? `?${params.toString()}` : ""}`;

  try {
    const response = await authenticatedApiClient({
      url: url,
      method: "GET",
    });

    return successResponse(
      response?.data?.data || response?.data,
      "API errors retrieved successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}

/**
 * Get API error by ID
 */
export async function getAPIError(
  errorId: string,
): Promise<APIResponse<APIError | null>> {
  const url = `/api/v1/admin/api-monitoring/errors/${errorId}`;

  try {
    const response = await authenticatedApiClient({
      url: url,
      method: "GET",
    });

    return successResponse(
      response?.data?.data || response?.data,
      "API error retrieved successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}

/**
 * Resolve API error
 */
export async function resolveAPIError(
  errorId: string,
  resolution_notes?: string,
): Promise<APIResponse<void>> {
  const url = `/api/v1/admin/api-monitoring/errors/${errorId}/resolve`;

  try {
    const response = await authenticatedApiClient({
      url: url,
      method: "POST",
      data: { resolution_notes },
    });

    return successResponse(
      response?.data?.data || response?.data,
      "API error resolved successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}

/**
 * Get API alerts with filtering
 */
export async function getAPIAlerts(
  filters?: APIFilters,
): Promise<APIResponse<APIAlert[]>> {
  const params = new URLSearchParams();

  if (filters?.severity) params.append("severity", filters.severity);
  if (filters?.time_range) params.append("time_range", filters.time_range);

  const url = `/api/v1/admin/api-monitoring/alerts${params.toString() ? `?${params.toString()}` : ""}`;

  try {
    const response = await authenticatedApiClient({
      url: url,
      method: "GET",
    });

    return successResponse(
      response?.data?.data || response?.data,
      "API alerts retrieved successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}

/**
 * Acknowledge API alert
 */
export async function acknowledgeAPIAlert(
  alertId: string,
  notes?: string,
): Promise<APIResponse<void>> {
  const url = `/api/v1/admin/api-monitoring/alerts/${alertId}/acknowledge`;

  try {
    const response = await authenticatedApiClient({
      url: url,
      method: "POST",
      data: { notes },
    });

    return successResponse(
      response?.data?.data || response?.data,
      "API alert acknowledged successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}

/**
 * Resolve API alert
 */
export async function resolveAPIAlert(
  alertId: string,
  resolution_notes?: string,
): Promise<APIResponse<void>> {
  const url = `/api/v1/admin/api-monitoring/alerts/${alertId}/resolve`;

  try {
    const response = await authenticatedApiClient({
      url: url,
      method: "POST",
      data: { resolution_notes },
    });

    return successResponse(
      response?.data?.data || response?.data,
      "API alert resolved successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}

/**
 * Get API monitoring statistics
 */
export async function getAPIStats(): Promise<APIResponse<APIStats | null>> {
  const url = "/api/v1/admin/api-monitoring/stats";

  try {
    const response = await authenticatedApiClient({
      url: url,
      method: "GET",
    });

    return successResponse(
      response?.data?.data || response?.data,
      "API statistics retrieved successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}

/**
 * Get API performance data over time
 */
export async function getAPIPerformanceData(
  timeRange: string = "24h",
  interval: string = "5m",
): Promise<APIResponse<APIPerformanceData[]>> {
  const params = new URLSearchParams();
  params.append("time_range", timeRange);
  params.append("interval", interval);

  const url = `/api/v1/admin/api-monitoring/performance?${params.toString()}`;

  try {
    const response = await authenticatedApiClient({
      url: url,
      method: "GET",
    });

    return successResponse(
      response?.data?.data || response?.data,
      "API performance data retrieved successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}

/**
 * Test API endpoint
 */
export async function testAPIEndpoint(
  endpointId: string,
  testData?: Record<string, any>,
): Promise<
  APIResponse<{
    success: boolean;
    status_code: number;
    response_time: number;
    response_body: any;
    error_message?: string;
  }>
> {
  const url = `/api/v1/admin/api-monitoring/endpoints/${endpointId}/test`;

  try {
    const response = await authenticatedApiClient({
      url: url,
      method: "POST",
      data: testData || {},
    });

    return successResponse(
      response?.data?.data || response?.data,
      "API endpoint test completed",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}

/**
 * Update endpoint configuration
 */
export async function updateEndpointConfig(
  endpointId: string,
  config: {
    rate_limit?: number;
    timeout?: number;
    is_deprecated?: boolean;
    description?: string;
  },
): Promise<APIResponse<APIEndpoint>> {
  const url = `/api/v1/admin/api-monitoring/endpoints/${endpointId}/config`;

  try {
    const response = await authenticatedApiClient({
      url: url,
      method: "PUT",
      data: config,
    });

    return successResponse(
      response?.data?.data || response?.data,
      "Endpoint configuration updated successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}

/**
 * Export API monitoring data
 */
export async function exportAPIData(
  type: "endpoints" | "metrics" | "errors" | "alerts",
  format: "csv" | "json" | "excel",
  filters?: APIFilters,
): Promise<APIResponse<{ download_url: string; expires_at: string }>> {
  const params = new URLSearchParams();

  params.append("type", type);
  params.append("format", format);
  if (filters?.time_range) params.append("time_range", filters.time_range);
  if (filters?.method) params.append("method", filters.method);
  if (filters?.category) params.append("category", filters.category);

  const url = `/api/v1/admin/api-monitoring/export?${params.toString()}`;

  try {
    const response = await authenticatedApiClient({
      url: url,
      method: "POST",
    });

    return successResponse(
      response?.data?.data || response?.data,
      "API monitoring data export initiated successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}

/**
 * Get API categories
 */
export async function getAPICategories(): Promise<APIResponse<string[]>> {
  const url = "/api/v1/admin/api-monitoring/categories";

  try {
    const response = await authenticatedApiClient({
      url: url,
      method: "GET",
    });

    return successResponse(
      response?.data?.data || response?.data,
      "API categories retrieved successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}

/**
 * Create API alert rule
 */
export async function createAlertRule(rule: {
  endpoint_id?: string;
  alert_type: string;
  threshold_value: number;
  severity: "low" | "medium" | "high" | "critical";
  description: string;
}): Promise<APIResponse<APIAlert>> {
  const url = "/api/v1/admin/api-monitoring/alert-rules";

  try {
    const response = await authenticatedApiClient({
      url: url,
      method: "POST",
      data: rule,
    });

    return successResponse(
      response?.data?.data || response?.data,
      "Alert rule created successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}

/**
 * Get real-time API metrics
 */
export async function getRealTimeMetrics(): Promise<
  APIResponse<{
    current_rps: number;
    avg_response_time: number;
    error_rate: number;
    active_connections: number;
    queue_size: number;
    cpu_usage: number;
    memory_usage: number;
    timestamp: string;
  }>
> {
  const url = "/api/v1/admin/api-monitoring/realtime";

  try {
    const response = await authenticatedApiClient({
      url: url,
      method: "GET",
    });

    return successResponse(
      response?.data?.data || response?.data,
      "Real-time metrics retrieved successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}
