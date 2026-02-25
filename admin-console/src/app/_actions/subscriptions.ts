"use server";

import type { APIResponse } from "@/types";
import authenticatedApiClient, {
  handleError,
  successResponse,
} from "./api-config";

export interface SubscriptionTier {
  id: string;
  name: string;
  display_name: string;
  description: string;
  price_monthly: number;
  price_yearly: number;
  max_workspaces: number;
  max_team_members: number;
  max_documents: number;
  max_workflows: number;
  max_custom_roles: number;
  features: string[];
  is_active: boolean;
  sort_order: number;
  created_at: string;
  updated_at: string;
}

export interface SubscriptionFeature {
  id: string;
  name: string;
  display_name: string;
  description: string;
  category: string;
  is_active: boolean;
  created_at: string;
}

export interface CreateTierRequest {
  name: string;
  display_name: string;
  description: string;
  price_monthly: number;
  price_yearly: number;
  max_workspaces: number;
  max_team_members: number;
  max_documents: number;
  max_workflows: number;
  max_custom_roles: number;
  features: string[];
  is_active: boolean;
  sort_order: number;
}

export interface UpdateTierRequest extends Partial<CreateTierRequest> {
  id: string;
}

export interface TrialResetRequest {
  trialDays: number;
  reason: string;
}

/**
 * Get all subscription tiers
 */
export async function getAllSubscriptionTiers(): Promise<
  APIResponse<SubscriptionTier[]>
> {
  const url = "/api/v1/admin/subscriptions/tiers";

  try {
    const response = await authenticatedApiClient({
      url: url,
      method: "GET",
    });

    return successResponse(
      response?.data?.data || response?.data,
      "Subscription tiers retrieved successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}

/**
 * Get subscription tier by ID
 */
export async function getSubscriptionTierById(
  tierId: string,
): Promise<APIResponse<SubscriptionTier | null>> {
  const url = `/api/v1/admin/subscriptions/tiers/${tierId}`;

  try {
    const response = await authenticatedApiClient({
      url: url,
      method: "GET",
    });

    return successResponse(
      response?.data?.data || response?.data,
      "Subscription tier retrieved successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}

/**
 * Create new subscription tier
 */
export async function createSubscriptionTier(
  request: CreateTierRequest,
): Promise<APIResponse<SubscriptionTier | null>> {
  const url = "/api/v1/admin/subscriptions/tiers";

  try {
    const response = await authenticatedApiClient({
      url: url,
      method: "POST",
      data: request,
    });

    return successResponse(
      response?.data?.data || response?.data,
      "Subscription tier created successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}

/**
 * Update subscription tier
 */
export async function updateSubscriptionTier(
  request: UpdateTierRequest,
): Promise<APIResponse<SubscriptionTier | null>> {
  const url = `/api/v1/admin/subscriptions/tiers/${request.id}`;

  try {
    const response = await authenticatedApiClient({
      url: url,
      method: "PUT",
      data: request,
    });

    return successResponse(
      response?.data?.data || response?.data,
      "Subscription tier updated successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}

/**
 * Delete subscription tier
 */
export async function deleteSubscriptionTier(
  tierId: string,
): Promise<APIResponse<null>> {
  const url = `/api/v1/admin/subscriptions/tiers/${tierId}`;

  try {
    await authenticatedApiClient({
      url: url,
      method: "DELETE",
    });

    return successResponse(null, "Subscription tier deleted successfully");
  } catch (error: Error | any) {
    return handleError(error);
  }
}

/**
 * Get all available features
 */
export async function getAllSubscriptionFeatures(): Promise<
  APIResponse<SubscriptionFeature[]>
> {
  const url = "/api/v1/admin/subscriptions/features";

  try {
    const response = await authenticatedApiClient({
      url: url,
      method: "GET",
    });

    return successResponse(
      response?.data?.data || response?.data,
      "Subscription features retrieved successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}

/**
 * Create new subscription feature
 */
export async function createSubscriptionFeature(
  request: Omit<SubscriptionFeature, "id" | "created_at">,
): Promise<APIResponse<SubscriptionFeature | null>> {
  const url = "/api/v1/admin/subscriptions/features";

  try {
    const response = await authenticatedApiClient({
      url: url,
      method: "POST",
      data: request,
    });

    return successResponse(
      response?.data?.data || response?.data,
      "Subscription feature created successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}

/**
 * Update subscription feature
 */
export async function updateSubscriptionFeature(
  featureId: string,
  request: Partial<Omit<SubscriptionFeature, "id" | "created_at">>,
): Promise<APIResponse<SubscriptionFeature | null>> {
  const url = `/api/v1/admin/subscriptions/features/${featureId}`;

  try {
    const response = await authenticatedApiClient({
      url: url,
      method: "PUT",
      data: request,
    });

    return successResponse(
      response?.data?.data || response?.data,
      "Subscription feature updated successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}

/**
 * Delete subscription feature
 */
export async function deleteSubscriptionFeature(
  featureId: string,
): Promise<APIResponse<null>> {
  const url = `/api/v1/admin/subscriptions/features/${featureId}`;

  try {
    await authenticatedApiClient({
      url: url,
      method: "DELETE",
    });

    return successResponse(null, "Subscription feature deleted successfully");
  } catch (error: Error | any) {
    return handleError(error);
  }
}

/**
 * Get organizations with trial status for trial management
 */
export async function getTrialOrganizations(): Promise<APIResponse<any[]>> {
  const url = "/api/v1/admin/subscriptions/trials";

  try {
    const response = await authenticatedApiClient({
      url: url,
      method: "GET",
    });

    return successResponse(
      response?.data?.data || response?.data,
      "Trial organizations retrieved successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}

/**
 * Reset organization trial period (Admin only)
 */
export async function resetOrganizationTrial(
  organizationId: string,
  request: TrialResetRequest,
): Promise<APIResponse<any>> {
  const url = `/api/v1/organizations/${organizationId}/trial/reset`;

  try {
    const response = await authenticatedApiClient({
      url: url,
      method: "POST",
      data: request,
    });

    return successResponse(
      response?.data?.data || response?.data,
      "Trial reset successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}

/**
 * Get subscription analytics
 */
export async function getSubscriptionAnalytics(): Promise<APIResponse<any>> {
  const url = "/api/v1/admin/subscriptions/analytics";

  try {
    const response = await authenticatedApiClient({
      url: url,
      method: "GET",
    });

    return successResponse(
      response?.data?.data || response?.data,
      "Subscription analytics retrieved successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}
/**
 * Get subscription statistics for admin dashboard
 */
export async function getSubscriptionStatistics(): Promise<
  APIResponse<{
    total_tiers: number;
    active_subscriptions: number;
    trial_organizations: number;
    monthly_revenue: number;
    revenue_growth: number;
  }>
> {
  const url = "/api/v1/admin/subscriptions/statistics";

  try {
    const response = await authenticatedApiClient({
      url: url,
      method: "GET",
    });

    return successResponse(
      response?.data?.data || response?.data,
      "Subscription statistics retrieved successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}

/**
 * Change organization subscription tier (Admin only)
 */
export async function changeOrganizationTier(
  organizationId: string,
  request: {
    new_tier: string;
    reason: string;
    override_limits?: boolean;
  },
): Promise<APIResponse<any>> {
  const url = `/api/v1/admin/organizations/${organizationId}/change-tier`;

  try {
    const response = await authenticatedApiClient({
      url: url,
      method: "POST",
      data: request,
    });

    return successResponse(
      response?.data?.data || response?.data,
      "Organization tier changed successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}

/**
 * Override organization limits (Admin only)
 */
export async function overrideOrganizationLimits(
  organizationId: string,
  request: {
    max_users?: number;
    storage_limit_gb?: number;
    features?: string[];
    reason: string;
    expires_at?: string;
  },
): Promise<APIResponse<any>> {
  const url = `/api/v1/admin/organizations/${organizationId}/override-limits`;

  try {
    const response = await authenticatedApiClient({
      url: url,
      method: "POST",
      data: request,
    });

    return successResponse(
      response?.data?.data || response?.data,
      "Organization limits overridden successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}

/**
 * Extend organization trial period (Admin only)
 * Adds days to the existing trial/grace period rather than resetting it
 */
export async function extendOrganizationTrial(
  organizationId: string,
  request: {
    daysToAdd: number;
    reason: string;
  },
): Promise<APIResponse<any>> {
  const url = `/api/v1/organizations/${organizationId}/trial/extend`;

  try {
    const response = await authenticatedApiClient({
      url: url,
      method: "POST",
      data: request,
    });

    return successResponse(
      response?.data?.data || response?.data,
      "Trial extended successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}

/**
 * Get audit logs for an organization (tier changes, limit overrides, etc.)
 */
export async function getOrganizationAuditLogs(
  organizationId: string,
  page: number = 1,
  limit: number = 20,
): Promise<APIResponse<any>> {
  const url = `/api/v1/admin/audit-logs?organizationId=${organizationId}&page=${page}&limit=${limit}`;

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
