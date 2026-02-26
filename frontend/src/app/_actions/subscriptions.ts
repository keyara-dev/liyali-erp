"use server";

import { APIResponse } from "@/types";
import { axios, handleError, successResponse } from "./api-config";
import authenticatedApiClient from "./api-config";

// Try multiple environment variable names and fallback to localhost
const BACKEND_URL =
  process.env.BASE_URL ||
  process.env.NEXT_PUBLIC_API_URL ||
  process.env.API_URL ||
  "http://localhost:8080";

export interface FeatureDetail {
  id: string;
  name: string;
  displayName: string;
  description: string;
  category: string;
}

export interface SubscriptionPlan {
  id: string;
  name: string;
  slug: string;
  displayName: string;
  description: string;
  priceMonthly: number;
  priceYearly: number;
  maxWorkspaces: number;
  maxTeamMembers: number;
  maxDocuments: number;
  maxWorkflows: number;
  maxCustomRoles: number;
  features: string[];
  featureDetails: FeatureDetail[];
  isActive: boolean;
  sortOrder: number;
  metadata?: {
    offline_capabilities?: boolean;
    api_access?: boolean;
    custom_roles?: boolean;
    priority_support?: boolean;
    dedicated_instance?: boolean;
    sla_guarantees?: boolean;
    custom_pricing?: boolean;
  };
  createdAt: string;
  updatedAt: string;
}

export interface TrialStatus {
  organizationId: string;
  subscriptionStatus: "trial" | "active" | "past_due" | "canceled" | "expired";
  trialStartDate?: string;
  trialEndDate?: string;
  gracePeriodEndsAt?: string;
  planSlug: string;
  planName: string;
  daysRemaining: number;
  isExpired: boolean;
  isActive: boolean;
  inGracePeriod: boolean;
}

export interface OrganizationSubscription {
  id: string;
  organizationId: string;
  planId: string;
  plan?: SubscriptionPlan;
  stripeSubscriptionId?: string;
  status: "trial" | "active" | "past_due" | "canceled" | "expired";
  currentPeriodStart?: string;
  currentPeriodEnd?: string;
  cancelAtPeriodEnd: boolean;
  paymentFailedCount: number;
  lastPaymentFailedAt?: string;
  createdAt: string;
  updatedAt: string;
}

export interface FeatureFlag {
  id: string;
  name: string;
  description: string;
  planRequirements: string[];
  isTrialAllowed: boolean;
  isEnterpriseOnly: boolean;
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface SubscriptionResponse {
  organization?: any;
  subscription?: OrganizationSubscription;
  plan?: SubscriptionPlan;
  trialStatus?: TrialStatus;
  availableFeatures?: FeatureFlag[];
  planLimits?: {
    organizationId: string;
    maxUsersAllowed: number;
    planMaxUsers: number;
    planMetadata: Record<string, any>;
    currentUserCount: number;
    canAddUsers: boolean;
  };
}

export interface UpgradeRequest {
  targetPlanSlug: "PRO_PLAN" | "ENTERPRISE";
  paymentMethodId?: string;
  billingCycle: "monthly" | "yearly";
  promoCode?: string;
}

export interface ExtendTrialRequest {
  daysToAdd: number;
  reason: string;
}

export interface SubscriptionPlansResponse {
  success: boolean;
  message: string;
  data: {
    plans: SubscriptionPlan[];
  };
}

export async function getSubscriptionPlans(): Promise<APIResponse> {
  const url = `/api/v1/subscriptions/plans`;
  try {
    const response = await axios.get(url, {
      headers: {
        "Content-Type": "application/json",
      },
    });

    return successResponse(response.data?.data);
  } catch (error) {
    return handleError(error, "GET", url);
  }
}

export async function getSubscriptionPlan(slug: string): Promise<APIResponse> {
  const url = `/api/v1/subscriptions/plans/${slug}`;
  try {
    const response = await axios.get(url, {
      headers: {
        "Content-Type": "application/json",
      },
    });

    return successResponse(response.data?.data);
  } catch (error) {
    return handleError(error, "GET", url);
  }
}

// ============================================================================
// AUTHENTICATED SUBSCRIPTION OPERATIONS
// ============================================================================

/**
 * Get organization subscription details
 */
export async function getOrganizationSubscription(
  organizationId: string,
): Promise<APIResponse> {
  const url = `/api/v1/organizations/${organizationId}/subscription`;

  try {
    const response = await authenticatedApiClient({
      url,
      method: "GET",
    });

    return successResponse(response.data?.data);
  } catch (error) {
    return handleError(error, "GET", url);
  }
}

/**
 * Get organization trial status
 */
export async function getOrganizationTrialStatus(
  organizationId: string,
): Promise<APIResponse> {
  const url = `/api/v1/organizations/${organizationId}/trial-status`;

  try {
    const response = await authenticatedApiClient({
      url,
      method: "GET",
    });

    return successResponse(response.data?.data);
  } catch (error) {
    return handleError(error, "GET", url);
  }
}

/**
 * Upgrade organization subscription
 */
export async function upgradeOrganization(
  organizationId: string,
  request: UpgradeRequest,
): Promise<APIResponse> {
  const url = `/api/v1/organizations/${organizationId}/upgrade`;

  try {
    const response = await authenticatedApiClient({
      url,
      method: "POST",
      data: request,
    });

    return successResponse(response.data?.data);
  } catch (error) {
    return handleError(error, "POST", url);
  }
}

/**
 * Downgrade organization subscription
 */
export async function downgradeOrganization(
  organizationId: string,
  request: { targetPlanSlug: string; reason?: string; feedback?: string },
): Promise<APIResponse> {
  const url = `/api/v1/organizations/${organizationId}/downgrade`;

  try {
    const response = await authenticatedApiClient({
      url,
      method: "POST",
      data: request,
    });

    return successResponse(response.data?.data);
  } catch (error) {
    return handleError(error, "POST", url);
  }
}

/**
 * Extend organization trial (admin only)
 */
export async function extendOrganizationTrial(
  organizationId: string,
  request: ExtendTrialRequest,
): Promise<APIResponse> {
  const url = `/api/v1/organizations/${organizationId}/trial/extend`;

  try {
    const response = await authenticatedApiClient({
      url,
      method: "POST",
      data: request,
    });

    return successResponse(response.data?.data);
  } catch (error) {
    return handleError(error, "POST", url);
  }
}

/**
 * Get organization features
 */
export async function getOrganizationFeatures(
  organizationId: string,
): Promise<APIResponse> {
  const url = `/api/v1/organizations/${organizationId}/features`;

  try {
    const response = await authenticatedApiClient({
      url,
      method: "GET",
    });

    return successResponse(response.data?.data);
  } catch (error) {
    return handleError(error, "GET", url);
  }
}

/**
 * Get organization resource usage and limits (tenant-scoped)
 */
export async function getOrganizationUsage(): Promise<APIResponse> {
  const url = `/api/v1/organization/usage`;

  try {
    const response = await authenticatedApiClient({
      url,
      method: "GET",
    });

    return successResponse(response.data?.data);
  } catch (error) {
    return handleError(error, "GET", url);
  }
}

/**
 * Check feature access for organization
 */
export async function checkFeatureAccess(
  organizationId: string,
  feature: string,
): Promise<APIResponse> {
  const url = `/api/v1/organizations/${organizationId}/features/check?feature=${encodeURIComponent(feature)}`;

  try {
    const response = await authenticatedApiClient({
      url,
      method: "GET",
    });

    return successResponse(response.data?.data);
  } catch (error) {
    return handleError(error, "GET", url);
  }
}
