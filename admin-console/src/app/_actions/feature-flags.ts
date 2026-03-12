"use server";

import type { APIResponse } from "@/types";
import authenticatedApiClient, { handleError, successResponse } from "./api-config";

// Feature Flags API Actions
// Comprehensive feature toggle management system

export interface FeatureFlag {
  id: string;
  key: string;
  name: string;
  description: string;
  type: "boolean" | "string" | "number" | "json";
  default_value: string;
  enabled: boolean;
  environment: "all" | "production" | "staging" | "development";
  category:
    | "feature"
    | "experiment"
    | "operational"
    | "killswitch"
    | "permission";
  tags: string[];
  targeting: {
    enabled: boolean;
    rules: TargetingRule[];
    rolloutPercentage: number;
    userSegments: string[];
  };
  variations: Variation[];
  created_at: string;
  updated_at: string;
  created_by: string;
  updated_by: string;
  last_evaluated?: string;
  evaluation_count: number;
  is_archived: boolean;
  expires_at?: string;
}

export interface Variation {
  id: string;
  name: string;
  value: string;
  description?: string;
  weight: number;
  isControl: boolean;
}

export interface TargetingRule {
  id: string;
  name: string;
  conditions: Condition[];
  variation: string;
  enabled: boolean;
  priority: number;
}

export interface Condition {
  attribute: string;
  operator:
    | "equals"
    | "not_equals"
    | "contains"
    | "not_contains"
    | "greater_than"
    | "less_than"
    | "in"
    | "not_in";
  value: string | string[] | number;
}

export interface FeatureFlagEvaluation {
  flagKey: string;
  userId?: string;
  userAttributes?: Record<string, any>;
  variation: string;
  value: string;
  reason: "targeting" | "rollout" | "default" | "disabled";
  timestamp: string;
}

export interface FeatureFlagStats {
  total: number;
  enabled: number;
  disabled: number;
  archived: number;
  byCategory: Record<string, number>;
  byEnvironment: Record<string, number>;
  byType: Record<string, number>;
  recentlyCreated: number;
  recentlyUpdated: number;
  expiringSoon: number;
  evaluationsToday: number;
}

export interface FeatureFlagAnalytics {
  flagKey: string;
  evaluations: {
    total: number;
    byVariation: Record<string, number>;
    byDay: { date: string; count: number }[];
    byUser: { userId: string; count: number }[];
  };
  performance: {
    avgEvaluationTime: number;
    errorRate: number;
    cacheHitRate: number;
  };
  targeting: {
    rulesMatched: Record<string, number>;
    segmentsMatched: Record<string, number>;
    rolloutDistribution: Record<string, number>;
  };
}

export interface FeatureFlagFilters {
  search?: string;
  category?: string;
  environment?: string;
  type?: string;
  enabled?: boolean;
  archived?: boolean;
  tags?: string[];
  createdAfter?: string;
  createdBefore?: string;
  expiringBefore?: string;
}

export interface BulkFlagOperation {
  action: "enable" | "disable" | "archive" | "delete" | "export";
  flagIds: string[];
  environment?: string;
}

export interface FlagTemplate {
  id: string;
  name: string;
  description: string;
  category: string;
  template: Partial<FeatureFlag>;
  tags: string[];
  isPublic: boolean;
  createdAt: string;
  createdBy: string;
  usageCount: number;
}

export interface FlagAuditLog {
  id: string;
  flagKey: string;
  action:
    | "created"
    | "updated"
    | "enabled"
    | "disabled"
    | "archived"
    | "deleted";
  changes: Record<string, { old: any; new: any }>;
  userId: string;
  userName: string;
  timestamp: string;
  environment: string;
  ipAddress: string;
  userAgent: string;
  reason?: string;
}

// Feature Flag Management
export async function getFeatureFlags(
  filters?: FeatureFlagFilters,
): Promise<APIResponse<FeatureFlag[]>> {
  const params = new URLSearchParams();
  if (filters?.search) params.append("search", filters.search);
  if (filters?.category) params.append("category", filters.category);
  if (filters?.environment) params.append("environment", filters.environment);
  if (filters?.type) params.append("type", filters.type);
  if (filters?.enabled !== undefined)
    params.append("enabled", filters.enabled.toString());
  if (filters?.archived !== undefined)
    params.append("archived", filters.archived.toString());

  try {
    const response = await authenticatedApiClient({
      url: `/api/v1/admin/feature-flags?${params.toString()}`,
      method: "GET",
    });
    return successResponse(
      response?.data?.data || response?.data || [],
      "Feature flags retrieved successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}

export async function getFeatureFlag(
  id: string,
): Promise<APIResponse<FeatureFlag>> {
  try {
    const response = await authenticatedApiClient({
      url: `/api/v1/admin/feature-flags/${id}`,
      method: "GET",
    });
    return successResponse(
      response?.data?.data || response?.data,
      "Feature flag retrieved successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}

export async function createFeatureFlag(
  flag: Omit<
    FeatureFlag,
    | "id"
    | "created_at"
    | "updated_at"
    | "created_by"
    | "updated_by"
    | "evaluation_count"
  >,
): Promise<APIResponse<FeatureFlag>> {
  try {
    const response = await authenticatedApiClient({
      url: `/api/v1/admin/feature-flags`,
      method: "POST",
      data: flag,
    });
    return successResponse(
      response?.data?.data || response?.data,
      "Feature flag created successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}

export async function updateFeatureFlag(
  id: string,
  updates: Partial<FeatureFlag>,
): Promise<APIResponse<FeatureFlag>> {
  try {
    const response = await authenticatedApiClient({
      url: `/api/v1/admin/feature-flags/${id}`,
      method: "PUT",
      data: updates,
    });
    return successResponse(
      response?.data?.data || response?.data,
      "Feature flag updated successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}

export async function deleteFeatureFlag(
  id: string,
): Promise<APIResponse<null>> {
  try {
    await authenticatedApiClient({
      url: `/api/v1/admin/feature-flags/${id}`,
      method: "DELETE",
    });
    return successResponse(null, "Feature flag deleted successfully");
  } catch (error: Error | any) {
    return handleError(error);
  }
}

export async function toggleFeatureFlag(
  id: string,
): Promise<APIResponse<FeatureFlag>> {
  try {
    const response = await authenticatedApiClient({
      url: `/api/v1/admin/feature-flags/${id}/toggle`,
      method: "POST",
    });
    return successResponse(
      response?.data?.data || response?.data,
      "Feature flag toggled successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}

export async function archiveFeatureFlag(
  id: string,
): Promise<APIResponse<FeatureFlag>> {
  try {
    const response = await authenticatedApiClient({
      url: `/api/v1/admin/feature-flags/${id}/archive`,
      method: "POST",
    });
    return successResponse(
      response?.data?.data || response?.data,
      "Feature flag archived successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}

export async function bulkUpdateFlags(
  _operation: BulkFlagOperation,
): Promise<APIResponse<null>> {
  return {
    success: false,
    data: null,
    message: "Bulk flag operations not yet implemented",
  };
}

// Feature Flag Evaluation
export async function evaluateFeatureFlag(
  flagKey: string,
  userId?: string,
  userAttributes?: Record<string, any>,
): Promise<APIResponse<FeatureFlagEvaluation>> {
  try {
    const response = await authenticatedApiClient({
      url: `/api/v1/admin/feature-flags/${flagKey}/evaluate`,
      method: "POST",
      data: { user_id: userId, user_attributes: userAttributes },
    });
    return successResponse(
      response?.data?.data || response?.data,
      "Feature flag evaluated successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}

export async function getFeatureFlagEvaluations(
  _flagKey: string,
  _filters?: {
    startDate?: string;
    endDate?: string;
    userId?: string;
    variation?: string;
  },
): Promise<APIResponse<FeatureFlagEvaluation[]>> {
  return { success: false, data: [], message: "Evaluation history not yet implemented" };
}

// Analytics and Statistics
export async function getFeatureFlagStats(): Promise<
  APIResponse<FeatureFlagStats>
> {
  try {
    const response = await authenticatedApiClient({
      url: `/api/v1/admin/feature-flags/stats`,
      method: "GET",
    });
    return successResponse(
      response?.data?.data || response?.data,
      "Feature flag stats retrieved successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}

export async function getFeatureFlagAnalytics(
  flagKey: string,
): Promise<APIResponse<FeatureFlagAnalytics>> {
  try {
    const response = await authenticatedApiClient({
      url: `/api/v1/admin/feature-flags/${flagKey}/analytics`,
      method: "GET",
    });
    return successResponse(
      response?.data?.data || response?.data,
      "Feature flag analytics retrieved successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}

// Templates
export async function getFlagTemplates(): Promise<APIResponse<FlagTemplate[]>> {
  return { success: false, data: [], message: "Flag templates not yet implemented" };
}

// Audit Trail
export async function getFeatureFlagAudit(filters?: {
  flagKey?: string;
  userId?: string;
  action?: string;
  startDate?: string;
  endDate?: string;
}): Promise<APIResponse<FlagAuditLog[]>> {
  return { success: false, data: [], message: "Flag audit trail not yet implemented" };
}

// Import/Export
export async function exportFeatureFlags(
  flagIds?: string[],
): Promise<APIResponse<{ flags: FeatureFlag[]; metadata: { exportedAt: string; version: string } }>> {
  try {
    const listResult = await getFeatureFlags();
    if (!listResult.success) return handleError(new Error(listResult.message));
    const flags = listResult.data ?? [];
    const filteredFlags = flagIds ? flags.filter((f) => flagIds.includes(f.id)) : flags;
    return successResponse(
      { flags: filteredFlags, metadata: { exportedAt: new Date().toISOString(), version: "1.0.0" } },
      "Feature flags exported successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}

export async function importFeatureFlags(_data: {
  flags: Partial<FeatureFlag>[];
  overwriteExisting: boolean;
}): Promise<APIResponse<{ imported: number; skipped: number }>> {
  return { success: false, data: null as any, message: "Import functionality not yet implemented" };
}
