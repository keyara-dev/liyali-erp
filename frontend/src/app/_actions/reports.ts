"use server";

import {
  SystemStatistics,
  ApprovalMetrics,
  UserActivityMetrics,
  AnalyticsDashboard,
  DateRange,
} from "@/types/reports";
import authenticatedApiClient from "./api-config";
import { handleError } from "./api-config";

/**
 * Get system statistics for admin reports
 */
export async function getSystemStatistics(
  dateRange?: DateRange,
): Promise<SystemStatistics> {
  try {
    const params = new URLSearchParams();
    if (dateRange?.startDate) params.append("start_date", dateRange.startDate);
    if (dateRange?.endDate) params.append("end_date", dateRange.endDate);

    const url = `/api/v1/admin/reports/system-stats${params.toString() ? `?${params}` : ""}`;

    const response = await authenticatedApiClient({
      method: "GET",
      url,
    });

    return response.data;
  } catch (error: any) {
    throw handleError(error, "GET", "/api/v1/admin/reports/system-stats");
  }
}

/**
 * Get approval metrics for admin reports
 */
export async function getApprovalMetrics(
  dateRange?: DateRange,
): Promise<ApprovalMetrics> {
  try {
    const params = new URLSearchParams();
    if (dateRange?.startDate) params.append("start_date", dateRange.startDate);
    if (dateRange?.endDate) params.append("end_date", dateRange.endDate);

    const url = `/api/v1/admin/reports/approval-metrics${params.toString() ? `?${params}` : ""}`;

    const response = await authenticatedApiClient({
      method: "GET",
      url,
    });

    return response.data;
  } catch (error: any) {
    throw handleError(error, "GET", "/api/v1/admin/reports/approval-metrics");
  }
}

/**
 * Get user activity metrics for admin reports
 */
export async function getUserActivityMetrics(
  dateRange?: DateRange,
): Promise<UserActivityMetrics> {
  try {
    const params = new URLSearchParams();
    if (dateRange?.startDate) params.append("start_date", dateRange.startDate);
    if (dateRange?.endDate) params.append("end_date", dateRange.endDate);

    const url = `/api/v1/admin/reports/user-activity${params.toString() ? `?${params}` : ""}`;

    const response = await authenticatedApiClient({
      method: "GET",
      url,
    });

    return response.data;
  } catch (error: any) {
    throw handleError(error, "GET", "/api/v1/admin/reports/user-activity");
  }
}

/**
 * Get analytics dashboard data for admin reports
 */
export async function getAnalyticsDashboard(
  dateRange?: DateRange,
): Promise<AnalyticsDashboard> {
  try {
    const params = new URLSearchParams();
    if (dateRange?.startDate) params.append("start_date", dateRange.startDate);
    if (dateRange?.endDate) params.append("end_date", dateRange.endDate);

    const url = `/api/v1/admin/reports/analytics${params.toString() ? `?${params}` : ""}`;

    const response = await authenticatedApiClient({
      method: "GET",
      url,
    });

    return response.data;
  } catch (error: any) {
    throw handleError(error, "GET", "/api/v1/admin/reports/analytics");
  }
}
