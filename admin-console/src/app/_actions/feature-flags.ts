"use server";

import authenticatedApiClient from "./api-config";

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
): Promise<FeatureFlag[]> {
  try {
    const params = new URLSearchParams();

    if (filters?.search) params.append("search", filters.search);
    if (filters?.category) params.append("category", filters.category);
    if (filters?.environment) params.append("environment", filters.environment);
    if (filters?.type) params.append("type", filters.type);
    if (filters?.enabled !== undefined)
      params.append("enabled", filters.enabled.toString());
    if (filters?.archived !== undefined)
      params.append("archived", filters.archived.toString());

    const response = await authenticatedApiClient({
      url: `/api/v1/admin/feature-flags?${params.toString()}`,
      method: "GET",
    });

    return response.data.data || [];
  } catch (error) {
    console.error("Error fetching feature flags:", error);
    throw error;
  }
}

export async function getFeatureFlag(id: string): Promise<FeatureFlag | null> {
  try {
    const response = await authenticatedApiClient({
      url: `/api/v1/admin/feature-flags/${id}`,
      method: "GET",
    });

    return response.data.data || null;
  } catch (error) {
    console.error("Error fetching feature flag:", error);
    return null;
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
): Promise<FeatureFlag> {
  try {
    const response = await authenticatedApiClient({
      url: `/api/v1/admin/feature-flags`,
      method: "POST",
      data: flag,
    });

    return response.data.data;
  } catch (error) {
    console.error("Error creating feature flag:", error);
    throw error;
  }
}

export async function updateFeatureFlag(
  id: string,
  updates: Partial<FeatureFlag>,
): Promise<FeatureFlag> {
  try {
    const response = await authenticatedApiClient({
      url: `/api/v1/admin/feature-flags/${id}`,
      method: "PUT",
      data: updates,
    });

    return response.data.data;
  } catch (error) {
    console.error("Error updating feature flag:", error);
    throw error;
  }
}

export async function deleteFeatureFlag(id: string): Promise<void> {
  try {
    await authenticatedApiClient({
      url: `/api/v1/admin/feature-flags/${id}`,
      method: "DELETE",
    });
  } catch (error) {
    console.error("Error deleting feature flag:", error);
    throw error;
  }
}

export async function toggleFeatureFlag(id: string): Promise<FeatureFlag> {
  try {
    const response = await authenticatedApiClient({
      url: `/api/v1/admin/feature-flags/${id}/toggle`,
      method: "POST",
    });

    return response.data.data;
  } catch (error) {
    console.error("Error toggling feature flag:", error);
    throw error;
  }
}

export async function archiveFeatureFlag(id: string): Promise<FeatureFlag> {
  try {
    const response = await authenticatedApiClient({
      url: `/api/v1/admin/feature-flags/${id}/archive`,
      method: "POST",
    });

    return response.data.data;
  } catch (error) {
    console.error("Error archiving feature flag:", error);
    throw error;
  }
}

export async function bulkUpdateFlags(
  _operation: BulkFlagOperation,
): Promise<void> {
  // Bulk flag operations require a dedicated backend endpoint
  // For now, handle individual flags through toggleFeatureFlag/archiveFeatureFlag
  console.warn("Bulk flag operations not yet implemented");
}

// Feature Flag Evaluation
export async function evaluateFeatureFlag(
  flagKey: string,
  userId?: string,
  userAttributes?: Record<string, any>,
): Promise<FeatureFlagEvaluation> {
  try {
    const response = await authenticatedApiClient({
      url: `/api/v1/admin/feature-flags/${flagKey}/evaluate`,
      method: "POST",
      data: {
        user_id: userId,
        user_attributes: userAttributes,
      },
    });

    return response.data.data;
  } catch (error) {
    console.error("Error evaluating feature flag:", error);
    throw error;
  }
}

export async function getFeatureFlagEvaluations(
  flagKey: string,
  filters?: {
    startDate?: string;
    endDate?: string;
    userId?: string;
    variation?: string;
  },
): Promise<FeatureFlagEvaluation[]> {
  // This would be implemented with proper evaluation logging
  return [];
}

// Analytics and Statistics
export async function getFeatureFlagStats(): Promise<FeatureFlagStats> {
  try {
    const response = await authenticatedApiClient({
      url: `/api/v1/admin/feature-flags/stats`,
      method: "GET",
    });

    return response.data.data;
  } catch (error) {
    console.error("Error fetching feature flag stats:", error);
    throw error;
  }
}

export async function getFeatureFlagAnalytics(
  flagKey: string,
): Promise<FeatureFlagAnalytics> {
  try {
    const response = await authenticatedApiClient({
      url: `/api/v1/admin/feature-flags/${flagKey}/analytics`,
      method: "GET",
    });

    return response.data.data;
  } catch (error) {
    console.error("Error fetching feature flag analytics:", error);
    throw error;
  }
}

// Templates
export async function getFlagTemplates(): Promise<FlagTemplate[]> {
  // This would be a more complex feature - for now return empty array
  return [];
}

// Audit Trail
export async function getFeatureFlagAudit(filters?: {
  flagKey?: string;
  userId?: string;
  action?: string;
  startDate?: string;
  endDate?: string;
}): Promise<FlagAuditLog[]> {
  // This would be implemented with a proper audit log system
  return [];
}

// Import/Export
export async function exportFeatureFlags(flagIds?: string[]): Promise<{
  flags: FeatureFlag[];
  metadata: {
    exportedAt: string;
    exportedBy: string;
    version: string;
  };
}> {
  const flags = await getFeatureFlags();
  const filteredFlags = flagIds
    ? flags.filter((f) => flagIds.includes(f.id))
    : flags;

  return {
    flags: filteredFlags,
    metadata: {
      exportedAt: new Date().toISOString(),
      exportedBy: "current-user@example.com",
      version: "1.0.0",
    },
  };
}

export async function importFeatureFlags(data: {
  flags: Partial<FeatureFlag>[];
  overwriteExisting: boolean;
}): Promise<{
  success: boolean;
  imported: number;
  skipped: number;
  errors: string[];
}> {
  // This would be implemented in the backend
  return {
    success: false,
    imported: 0,
    skipped: 0,
    errors: ["Import functionality not yet implemented"],
  };
}
