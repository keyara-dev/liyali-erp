"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useOrganizationContext } from "@/hooks/use-organization";
import { USER_ORGS_QUERY_KEY } from "@/hooks/use-user-organizations";
import { toast } from "sonner";

/**
 * Hook for managing organization tier operations
 * Handles tier checks, upgrades, and related functionality
 */
export function useOrganizationTier() {
  const { currentOrganization } = useOrganizationContext();
  const queryClient = useQueryClient();

  const currentTier = (currentOrganization?.tier?.toUpperCase() ||
    "STARTER") as "STARTER" | "PRO" | "ENTERPRISE";

  // Tier checking utilities
  const isStarter = currentTier === "STARTER";
  const isPro = currentTier === "PRO";
  const isEnterprise = currentTier === "ENTERPRISE";
  const canUpgrade = isStarter; // Only STARTER can self-upgrade to PRO

  // Upgrade to Pro mutation
  const upgradeToPro = useMutation({
    mutationFn: async () => {
      // TODO: Implement actual API call to upgrade organization
      const response = await fetch(
        `/api/v1/organizations/${currentOrganization?.id}/upgrade`,
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            targetTier: "PRO",
            paymentMethod: "stripe", // or other payment provider
          }),
        }
      );

      if (!response.ok) {
        throw new Error("Failed to upgrade organization");
      }

      return response.json();
    },
    onSuccess: () => {
      // Invalidate and refetch organization data
      queryClient.invalidateQueries({ queryKey: USER_ORGS_QUERY_KEY });

      toast.success("Successfully upgraded to Pro plan!", {
        description: "Your organization now has access to all Pro features.",
      });
    },
    onError: (error: any) => {
      toast.error("Upgrade failed", {
        description: error.message || "Please try again or contact support.",
      });
    },
  });

  // Feature access checks based on tier
  const hasFeature = (feature: string): boolean => {
    const TIER_FEATURES = {
      STARTER: ["basic_workflows", "standard_reporting", "email_support"],
      PRO: [
        "basic_workflows",
        "standard_reporting",
        "email_support",
        "advanced_workflows",
        "custom_approval_chains",
        "advanced_analytics",
        "api_access",
        "priority_support",
        "custom_integrations",
        "advanced_permissions",
      ],
      ENTERPRISE: [
        "basic_workflows",
        "standard_reporting",
        "email_support",
        "advanced_workflows",
        "custom_approval_chains",
        "advanced_analytics",
        "api_access",
        "priority_support",
        "custom_integrations",
        "advanced_permissions",
        "enterprise_security",
        "custom_branding",
        "dedicated_support",
        "sla_guarantees",
        "compliance_features",
        "custom_development",
        "on_premise_deployment",
      ],
    };

    return TIER_FEATURES[currentTier]?.includes(feature) || false;
  };

  // Get tier limits
  const getTierLimits = () => {
    const TIER_LIMITS = {
      STARTER: {
        maxUsers: 50,
        maxWorkflows: 50,
        maxIntegrations: 5,
        storageGB: 10,
      },
      PRO: {
        maxUsers: 200,
        maxWorkflows: 200,
        maxIntegrations: 50,
        storageGB: 100,
      },
      ENTERPRISE: {
        maxUsers: -1, // Unlimited
        maxWorkflows: -1, // Unlimited
        maxIntegrations: -1, // Unlimited
        storageGB: -1, // Unlimited
      },
    };

    return TIER_LIMITS[currentTier];
  };

  return {
    // Current tier info
    currentTier,
    currentOrganization,

    // Tier checks
    isStarter,
    isPro,
    isEnterprise,
    canUpgrade,

    // Feature access
    hasFeature,
    getTierLimits,

    // Actions
    upgradeToPro: upgradeToPro.mutateAsync,
    isUpgrading: upgradeToPro.isPending,
  };
}
