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

// --- Query Hooks ---

export function useFeatureFlags(filters?: FeatureFlagFilters) {
  return useQuery({
    queryKey: ["feature-flags", filters],
    queryFn: () => getFeatureFlags(filters),
  });
}

export function useFeatureFlag(id: string) {
  return useQuery({
    queryKey: ["feature-flags", id],
    queryFn: () => getFeatureFlag(id),
    enabled: !!id,
  });
}

export function useFeatureFlagStats() {
  return useQuery({
    queryKey: ["feature-flags", "stats"],
    queryFn: () => getFeatureFlagStats(),
  });
}

export function useFeatureFlagAnalytics(flagKey: string) {
  return useQuery({
    queryKey: ["feature-flags", flagKey, "analytics"],
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
      queryClient.invalidateQueries({ queryKey: ["feature-flags"] });
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
      queryClient.invalidateQueries({ queryKey: ["feature-flags"] });
    },
  });
}

export function useDeleteFeatureFlag() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => deleteFeatureFlag(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["feature-flags"] });
    },
  });
}

export function useToggleFeatureFlag() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => toggleFeatureFlag(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["feature-flags"] });
    },
  });
}

export function useArchiveFeatureFlag() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => archiveFeatureFlag(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["feature-flags"] });
    },
  });
}
