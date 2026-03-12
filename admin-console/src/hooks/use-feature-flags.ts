import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  getFeatureFlags,
  getFeatureFlag,
  createFeatureFlag,
  updateFeatureFlag,
  deleteFeatureFlag,
  toggleFeatureFlag,
  archiveFeatureFlag,
  getFeatureFlagStats,
  getFeatureFlagAnalytics,
  type FeatureFlag,
  type FeatureFlagFilters,
} from "@/app/_actions/feature-flags";
import { queryKeys } from "@/lib/query-keys";

// --- Query Hooks ---

export function useFeatureFlags(filters?: FeatureFlagFilters) {
  return useQuery({
    queryKey: queryKeys.featureFlags.list(filters),
    queryFn: () => getFeatureFlags(filters),
  });
}

export function useFeatureFlag(id: string) {
  return useQuery({
    queryKey: queryKeys.featureFlags.detail(id),
    queryFn: () => getFeatureFlag(id),
    enabled: !!id,
  });
}

export function useFeatureFlagStats() {
  return useQuery({
    queryKey: queryKeys.featureFlags.stats(),
    queryFn: () => getFeatureFlagStats(),
  });
}

export function useFeatureFlagAnalytics(flagKey: string) {
  return useQuery({
    queryKey: queryKeys.featureFlags.analytics(flagKey),
    queryFn: () => getFeatureFlagAnalytics(flagKey),
    enabled: !!flagKey,
  });
}

// --- Mutation Hooks ---

export function useCreateFeatureFlag() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (
      flag: Omit<
        FeatureFlag,
        | "id"
        | "created_at"
        | "updated_at"
        | "created_by"
        | "updated_by"
        | "evaluation_count"
      >,
    ) => createFeatureFlag(flag),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.featureFlags.all });
    },
  });
}

export function useUpdateFeatureFlag() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({
      id,
      updates,
    }: {
      id: string;
      updates: Partial<FeatureFlag>;
    }) => updateFeatureFlag(id, updates),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.featureFlags.all });
    },
  });
}

export function useDeleteFeatureFlag() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => deleteFeatureFlag(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.featureFlags.all });
    },
  });
}

export function useToggleFeatureFlag() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => toggleFeatureFlag(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.featureFlags.all });
    },
  });
}

export function useArchiveFeatureFlag() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => archiveFeatureFlag(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.featureFlags.all });
    },
  });
}
