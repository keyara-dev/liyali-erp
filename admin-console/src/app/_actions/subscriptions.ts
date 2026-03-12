"use server";

import type { APIResponse } from "@/types";
import authenticatedApiClient, {
  handleError,
  successResponse,
} from "./api-config";

export interface SubscriptionTier {
  id: string;
  name: string;
  displayName: string;
  description: string;
  priceMonthly: number;
  priceYearly: number;
  maxWorkspaces: number;
  maxTeamMembers: number;
  maxDocuments: number;
  maxWorkflows: number;
  maxCustomRoles: number;
  maxRequisitions: number;
  maxBudgets: number;
  maxPurchaseOrders: number;
  maxPaymentVouchers: number;
  maxGRNs: number;
  maxDepartments: number;
  maxVendors: number;
  features: string[];
  isActive: boolean;
  sortOrder: number;
  createdAt: string;
  updatedAt: string;
  featureCount?: number;
  organizationCount?: number;
}

export interface SubscriptionFeature {
  id: string;
  name: string;
  displayName: string;
  description: string;
  category: string;
  isActive: boolean;
  createdAt: string;
}

export interface CreateTierRequest {
  name: string;
  displayName: string;
  description: string;
  priceMonthly: number;
  priceYearly: number;
  maxWorkspaces: number;
  maxTeamMembers: number;
  maxDocuments: number;
  maxWorkflows: number;
  maxCustomRoles: number;
  maxRequisitions?: number;
  maxBudgets?: number;
  maxPurchaseOrders?: number;
  maxPaymentVouchers?: number;
  maxGRNs?: number;
  maxDepartments?: number;
  maxVendors?: number;
  features: string[];
  isActive: boolean;
  sortOrder: number;
}

export interface UpdateTierRequest extends Partial<Omit<CreateTierRequest, "name">> {
  id: string;
}

export interface TrialResetRequest {
  trial_days: number;
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
  request: Omit<SubscriptionFeature, "id" | "createdAt">,
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
  request: Partial<Omit<SubscriptionFeature, "id" | "createdAt">>,
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
  const url = `/api/v1/admin/organizations/${organizationId}/trial/reset`;

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
    newTier: string;
    reason: string;
    overrideLimits?: boolean;
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
    days_to_add: number;
    reason: string;
  },
): Promise<APIResponse<any>> {
  const url = `/api/v1/admin/organizations/${organizationId}/trial/extend`;

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

/**
 * Get subscription change history for an organization
 */
export async function getOrganizationSubscriptionHistory(
  organizationId: string,
  params?: { page?: number; limit?: number },
): Promise<APIResponse<any[]>> {
  const searchParams = new URLSearchParams();
  if (params?.page) searchParams.set("page", String(params.page));
  if (params?.limit) searchParams.set("limit", String(params.limit));
  const query = searchParams.toString();
  const url = `/api/v1/admin/organizations/${organizationId}/subscription/history${query ? `?${query}` : ""}`;

  try {
    const response = await authenticatedApiClient({ url, method: "GET" });
    return successResponse(
      response?.data?.data || response?.data,
      "Subscription history retrieved successfully",
      response?.data?.pagination,
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}
