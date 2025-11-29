import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  fetchPremiumConfig,
  updatePremiumConfig,
} from "@/app/_actions/premium-config";

const PREMIUM_CONFIG_KEY = ["premium-config"] as const;

interface UpdatePremiumConfigParams {
  premiumCost: number;
  premiumDuration: number;
  isActive: boolean;
}

/**
 * Hook for fetching premium configuration
 */
export function usePremiumConfig() {
  return useQuery({
    queryKey: PREMIUM_CONFIG_KEY,
    queryFn: async () => {
      const result = await fetchPremiumConfig();
      if (!result.success) {
        throw new Error(result.message);
      }
      return result.data;
    },
  });
}

/**
 * Hook for updating premium configuration
 */
export function useUpdatePremiumConfig() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: updatePremiumConfig,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: PREMIUM_CONFIG_KEY });
    },
  });
}
