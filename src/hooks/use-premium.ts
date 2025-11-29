"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  fetchPremiumStats,
  fetchPremiumUsers,
  processPremiumExpirations,
  extendPremiumGrace,
} from "@/app/_actions/premium";

// Query keys
const PREMIUM_STATS_KEY = ["premium", "stats"] as const;
const PREMIUM_USERS_KEY = ["premium", "users"] as const;

/**
 * Hook to fetch premium statistics
 */
export function usePremiumStats() {
  return useQuery({
    queryKey: PREMIUM_STATS_KEY,
    queryFn: async () => {
      const result = await fetchPremiumStats();
      if (!result.success) throw new Error(result.message);
      return result.data;
    },
    staleTime: 1000 * 60 * 5, // 5 minutes
  });
}

/**
 * Hook to fetch premium users
 */
export function usePremiumUsers() {
  return useQuery({
    queryKey: PREMIUM_USERS_KEY,
    queryFn: async () => {
      const result = await fetchPremiumUsers();
      if (!result.success) throw new Error(result.message);
      return result.data;
    },
    staleTime: 1000 * 60 * 5, // 5 minutes
  });
}

/**
 * Hook to process premium expirations
 */
export function useProcessPremiumExpirations() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: processPremiumExpirations,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: PREMIUM_STATS_KEY });
      queryClient.invalidateQueries({ queryKey: PREMIUM_USERS_KEY });
    },
  });
}

/**
 * Hook to extend premium grace period
 */
export function useExtendPremiumGrace() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: extendPremiumGrace,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: PREMIUM_USERS_KEY });
      queryClient.invalidateQueries({ queryKey: PREMIUM_STATS_KEY });
    },
  });
}
