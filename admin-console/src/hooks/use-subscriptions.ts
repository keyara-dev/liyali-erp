import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  getAllSubscriptionTiers,
  getSubscriptionTierById,
  createSubscriptionTier,
  updateSubscriptionTier,
  deleteSubscriptionTier,
  getAllSubscriptionFeatures,
  createSubscriptionFeature,
  updateSubscriptionFeature,
  deleteSubscriptionFeature,
  getTrialOrganizations,
  getSubscriptionAnalytics,
  getSubscriptionStatistics,
  changeOrganizationTier,
  overrideOrganizationLimits,
  resetOrganizationTrial,
  extendOrganizationTrial,
  type CreateTierRequest,
  type UpdateTierRequest,
  type SubscriptionFeature,
  type TrialResetRequest,
} from "@/app/_actions/subscriptions";

// --- Query Hooks ---

export function useSubscriptionTiers() {
  return useQuery({
    queryKey: ["subscriptions", "tiers"],
    queryFn: async () => {
      const result = await getAllSubscriptionTiers();
      if (!result.success) throw new Error(result.message);
      return result.data;
    },
  });
}

export function useSubscriptionTier(tierId: string) {
  return useQuery({
    queryKey: ["subscriptions", "tiers", tierId],
    queryFn: async () => {
      const result = await getSubscriptionTierById(tierId);
      if (!result.success) throw new Error(result.message);
      return result.data;
    },
    enabled: !!tierId,
  });
}

export function useSubscriptionFeatures() {
  return useQuery({
    queryKey: ["subscriptions", "features"],
    queryFn: async () => {
      const result = await getAllSubscriptionFeatures();
      if (!result.success) throw new Error(result.message);
      return result.data;
    },
  });
}

export function useTrialOrganizations() {
  return useQuery({
    queryKey: ["subscriptions", "trials"],
    queryFn: async () => {
      const result = await getTrialOrganizations();
      if (!result.success) throw new Error(result.message);
      return result.data;
    },
  });
}

export function useSubscriptionAnalytics() {
  return useQuery({
    queryKey: ["subscriptions", "analytics"],
    queryFn: async () => {
      const result = await getSubscriptionAnalytics();
      if (!result.success) throw new Error(result.message);
      return result.data;
    },
  });
}

export function useSubscriptionStatistics() {
  return useQuery({
    queryKey: ["subscriptions", "statistics"],
    queryFn: async () => {
      const result = await getSubscriptionStatistics();
      if (!result.success) throw new Error(result.message);
      return result.data;
    },
  });
}

// --- Mutation Hooks ---

export function useCreateSubscriptionTier() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateTierRequest) => createSubscriptionTier(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["subscriptions", "tiers"] });
    },
  });
}

export function useUpdateSubscriptionTier() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: UpdateTierRequest) => updateSubscriptionTier(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["subscriptions", "tiers"] });
    },
  });
}

export function useDeleteSubscriptionTier() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (tierId: string) => deleteSubscriptionTier(tierId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["subscriptions", "tiers"] });
    },
  });
}

export function useCreateSubscriptionFeature() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (
      data: Omit<SubscriptionFeature, "id" | "createdAt">,
    ) => createSubscriptionFeature(data),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ["subscriptions", "features"],
      });
    },
  });
}

export function useUpdateSubscriptionFeature() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({
      featureId,
      data,
    }: {
      featureId: string;
      data: Partial<Omit<SubscriptionFeature, "id" | "createdAt">>;
    }) => updateSubscriptionFeature(featureId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ["subscriptions", "features"],
      });
    },
  });
}

export function useDeleteSubscriptionFeature() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (featureId: string) => deleteSubscriptionFeature(featureId),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ["subscriptions", "features"],
      });
    },
  });
}

export function useChangeOrganizationTier() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({
      organizationId,
      data,
    }: {
      organizationId: string;
      data: { newTier: string; reason: string; overrideLimits?: boolean };
    }) => changeOrganizationTier(organizationId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["subscriptions"] });
      queryClient.invalidateQueries({ queryKey: ["organizations"] });
    },
  });
}

export function useOverrideOrganizationLimits() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({
      organizationId,
      data,
    }: {
      organizationId: string;
      data: {
        max_users?: number;
        storage_limit_gb?: number;
        features?: string[];
        reason: string;
        expires_at?: string;
      };
    }) => overrideOrganizationLimits(organizationId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["organizations"] });
    },
  });
}

export function useResetOrganizationTrial() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({
      organizationId,
      data,
    }: {
      organizationId: string;
      data: TrialResetRequest;
    }) => resetOrganizationTrial(organizationId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["subscriptions", "trials"] });
      queryClient.invalidateQueries({ queryKey: ["organizations"] });
    },
  });
}

export function useExtendOrganizationTrial() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({
      organizationId,
      data,
    }: {
      organizationId: string;
      data: { daysToAdd: number; reason: string };
    }) => extendOrganizationTrial(organizationId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["subscriptions", "trials"] });
      queryClient.invalidateQueries({ queryKey: ["organizations"] });
    },
  });
}
