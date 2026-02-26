"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import {
  getSubscriptionPlan,
  getSubscriptionPlans,
  getOrganizationSubscription,
  getOrganizationTrialStatus,
  upgradeOrganization,
  downgradeOrganization,
  extendOrganizationTrial,
  getOrganizationFeatures,
  checkFeatureAccess,
  type SubscriptionPlan,
  type TrialStatus,
  type OrganizationSubscription,
  type FeatureFlag,
  type SubscriptionResponse,
  type UpgradeRequest,
  type ExtendTrialRequest,
} from "@/app/_actions/subscriptions";

// Re-export types for convenience
export type {
  SubscriptionPlan,
  TrialStatus,
  OrganizationSubscription,
  FeatureFlag,
  SubscriptionResponse,
  UpgradeRequest,
  ExtendTrialRequest,
};

// ============================================================================
// QUERY HOOKS
// ============================================================================

/**
 * Hook to get all subscription plans using server action
 */
export function useSubscriptionPlans() {
  return useQuery({
    queryKey: ["subscription-plans"],
    queryFn: async () => {
      const response = await getSubscriptionPlans();
      if (!response.success) {
        throw new Error(response.message);
      }
      return response;
    },
    select: (data) => data.data.plans,
    staleTime: 5 * 60 * 1000, // 5 minutes
    retry: 3,
    retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 30000),
  });
}

export function useSubscriptionPlan(slug: string) {
  return useQuery({
    queryKey: ["subscription-plan", slug],
    queryFn: () => getSubscriptionPlan(slug),
    select: (data) => data.data,
    enabled: !!slug,
    staleTime: 5 * 60 * 1000, // 5 minutes
  });
}

/**
 * Hook to get organization subscription details
 */
export function useOrganizationSubscription(organizationId: string) {
  return useQuery({
    queryKey: ["organization-subscription", organizationId],
    queryFn: async () => {
      const response = await getOrganizationSubscription(organizationId);
      if (!response.success) {
        throw new Error(response.message);
      }
      return response;
    },
    select: (data) => data.data,
    enabled: !!organizationId,
    staleTime: 1 * 60 * 1000, // 1 minute
  });
}

/**
 * Hook to get organization trial status
 */
export function useOrganizationTrialStatus(organizationId: string) {
  return useQuery({
    queryKey: ["organization-trial-status", organizationId],
    queryFn: async () => {
      const response = await getOrganizationTrialStatus(organizationId);
      if (!response.success) {
        throw new Error(response.message);
      }
      return response;
    },
    select: (data) => data.data,
    enabled: !!organizationId,
    refetchInterval: 60 * 1000, // Refetch every minute to keep trial countdown accurate
  });
}

/**
 * Hook to get organization features
 */
export function useOrganizationFeatures(organizationId: string) {
  return useQuery({
    queryKey: ["organization-features", organizationId],
    queryFn: async () => {
      const response = await getOrganizationFeatures(organizationId);
      if (!response.success) {
        throw new Error(response.message);
      }
      return response;
    },
    select: (data) => data.data,
    enabled: !!organizationId,
    staleTime: 2 * 60 * 1000, // 2 minutes
  });
}

/**
 * Hook to check if organization has access to a specific feature
 */
export function useFeatureAccess(organizationId: string, feature: string) {
  return useQuery({
    queryKey: ["feature-access", organizationId, feature],
    queryFn: async () => {
      const response = await checkFeatureAccess(organizationId, feature);
      if (!response.success) {
        throw new Error(response.message);
      }
      return response;
    },
    select: (data) => data.data.hasAccess,
    enabled: !!organizationId && !!feature,
    staleTime: 2 * 60 * 1000, // 2 minutes
  });
}

// ============================================================================
// MUTATION HOOKS
// ============================================================================

/**
 * Hook to upgrade organization subscription
 */
export function useUpgradeOrganization() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({
      organizationId,
      request,
    }: {
      organizationId: string;
      request: UpgradeRequest;
    }) => {
      const response = await upgradeOrganization(organizationId, request);
      if (!response.success) {
        throw new Error(response.message);
      }
      return response;
    },
    onSuccess: (data, variables) => {
      // Invalidate and refetch related queries
      queryClient.invalidateQueries({
        queryKey: ["organization-subscription", variables.organizationId],
      });
      queryClient.invalidateQueries({
        queryKey: ["organization-trial-status", variables.organizationId],
      });
      queryClient.invalidateQueries({
        queryKey: ["organization-features", variables.organizationId],
      });
      queryClient.invalidateQueries({ queryKey: ["organizations"] });

      toast.success("Upgrade successful!", {
        description: "Your organization has been upgraded successfully.",
      });
    },
    onError: (error: Error) => {
      toast.error("Upgrade failed", {
        description: error.message || "Please try again or contact support.",
      });
    },
  });
}

/**
 * Hook to downgrade organization subscription
 */
export function useDowngradeOrganization() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({
      organizationId,
      request,
    }: {
      organizationId: string;
      request: { targetPlanSlug: string; reason?: string; feedback?: string };
    }) => {
      const response = await downgradeOrganization(organizationId, request);
      if (!response.success) {
        throw new Error(response.message);
      }
      return response;
    },
    onSuccess: (data, variables) => {
      // Invalidate and refetch related queries
      queryClient.invalidateQueries({
        queryKey: ["organization-subscription", variables.organizationId],
      });
      queryClient.invalidateQueries({
        queryKey: ["organization-trial-status", variables.organizationId],
      });
      queryClient.invalidateQueries({
        queryKey: ["organization-features", variables.organizationId],
      });
      queryClient.invalidateQueries({ queryKey: ["organizations"] });

      toast.success("Downgrade scheduled", {
        description:
          "Your plan will be downgraded at the end of the current billing period.",
      });
    },
    onError: (error: Error) => {
      toast.error("Downgrade failed", {
        description: error.message || "Please try again or contact support.",
      });
    },
  });
}

/**
 * Hook to extend organization trial (admin only)
 */
export function useExtendOrganizationTrial() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({
      organizationId,
      request,
    }: {
      organizationId: string;
      request: ExtendTrialRequest;
    }) => {
      const response = await extendOrganizationTrial(organizationId, request);
      if (!response.success) {
        throw new Error(response.message);
      }
      return response;
    },
    onSuccess: (data, variables) => {
      // Invalidate and refetch related queries
      queryClient.invalidateQueries({
        queryKey: ["organization-trial-status", variables.organizationId],
      });
      queryClient.invalidateQueries({
        queryKey: ["organization-subscription", variables.organizationId],
      });

      toast.success("Trial extended", {
        description: `Trial has been extended by ${variables.request.daysToAdd} days.`,
      });
    },
    onError: (error: Error) => {
      toast.error("Failed to extend trial", {
        description: error.message || "Please try again or contact support.",
      });
    },
  });
}

// ============================================================================
// UTILITY HOOKS
// ============================================================================

/**
 * Hook to check if organization has a specific feature
 * Returns a boolean and handles loading/error states
 */
export function useHasFeature(organizationId: string, feature: string) {
  const {
    data: hasAccess,
    isLoading,
    error,
  } = useFeatureAccess(organizationId, feature);

  return {
    hasFeature: hasAccess ?? false,
    isLoading,
    error,
  };
}

/**
 * Hook to get plan features for comparison
 */
export function usePlanFeatures() {
  const { data: plans, isLoading } = useSubscriptionPlans();

  const getPlanFeatures = (planSlug: string) => {
    const plan = plans?.find((p: SubscriptionPlan) => p.slug === planSlug);
    return plan?.features || [];
  };

  const comparePlans = (currentPlanSlug: string, targetPlanSlug: string) => {
    const currentFeatures = getPlanFeatures(currentPlanSlug);
    const targetFeatures = getPlanFeatures(targetPlanSlug);

    const newFeatures = targetFeatures.filter(
      (feature: string) => !currentFeatures.includes(feature),
    );
    const removedFeatures = currentFeatures.filter(
      (feature: string) => !targetFeatures.includes(feature),
    );

    return {
      newFeatures,
      removedFeatures,
      allTargetFeatures: targetFeatures,
    };
  };

  return {
    plans: plans || [],
    isLoading,
    getPlanFeatures,
    comparePlans,
  };
}

/**
 * Hook to get trial countdown information
 */
export function useTrialCountdown(organizationId: string) {
  const { data: trialStatus, isLoading } =
    useOrganizationTrialStatus(organizationId);

  const getTrialInfo = () => {
    if (!trialStatus) return null;

    const isTrialActive = trialStatus.isActive && !trialStatus.isExpired;
    const daysRemaining = Math.max(0, trialStatus.daysRemaining);
    const isExpiringSoon = daysRemaining <= 3 && isTrialActive;
    const isExpired = trialStatus.isExpired;
    const inGracePeriod = trialStatus.inGracePeriod;

    let status: "active" | "expiring" | "expired" | "grace" | "inactive" =
      "inactive";
    let message = "";
    let urgency: "low" | "medium" | "high" = "low";

    if (isTrialActive) {
      if (isExpiringSoon) {
        status = "expiring";
        message = `Trial expires in ${daysRemaining} day${daysRemaining !== 1 ? "s" : ""}`;
        urgency = "high";
      } else if (daysRemaining <= 7) {
        status = "active";
        message = `${daysRemaining} days remaining in trial`;
        urgency = "medium";
      } else {
        status = "active";
        message = `${daysRemaining} days remaining in trial`;
        urgency = "low";
      }
    } else if (isExpired && inGracePeriod) {
      status = "grace";
      message = "Trial expired - Limited access";
      urgency = "high";
    } else if (isExpired) {
      status = "expired";
      message = "Trial expired";
      urgency = "high";
    }

    return {
      status,
      message,
      urgency,
      daysRemaining,
      isTrialActive,
      isExpiringSoon,
      isExpired,
      inGracePeriod,
      trialStatus,
    };
  };

  return {
    trialInfo: getTrialInfo(),
    isLoading,
  };
}
