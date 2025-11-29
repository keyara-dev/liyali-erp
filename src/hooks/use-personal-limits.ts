import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  fetchAllPersonalLimits,
  fetchPersonalLimitsBySeller,
  upsertPersonalLimits,
} from "@/app/_actions/personal-limits";

const PERSONAL_LIMITS_KEY = ["personal-limits"] as const;

interface UpsertPersonalLimitsParams {
  sellerId: string;
  personalMaxProducts?: number | null;
  personalMaxImages?: number | null;
  adminAllocatedProducts?: number;
  adminAllocatedImages?: number;
  allocatedBy?: string;
  notes?: string | null;
}

/**
 * Hook for fetching all personal limits
 */
export function useAllPersonalLimits() {
  return useQuery({
    queryKey: PERSONAL_LIMITS_KEY,
    queryFn: async () => {
      const result = await fetchAllPersonalLimits();
      if (!result.success) {
        throw new Error(result.message);
      }
      return result.data;
    },
  });
}

/**
 * Hook for fetching personal limits for a specific seller
 */
export function usePersonalLimitsBySeller(sellerId: string) {
  return useQuery({
    queryKey: [...PERSONAL_LIMITS_KEY, sellerId],
    queryFn: async () => {
      const result = await fetchPersonalLimitsBySeller(sellerId);
      if (!result.success) {
        throw new Error(result.message);
      }
      return result.data;
    },
  });
}

/**
 * Hook for upserting personal limits
 */
export function useUpsertPersonalLimits() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: upsertPersonalLimits,
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: PERSONAL_LIMITS_KEY });
      queryClient.invalidateQueries({ queryKey: [...PERSONAL_LIMITS_KEY, variables.sellerId] });
    },
  });
}
