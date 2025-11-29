"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  fetchSignupSettings,
  fetchSignupAnalytics,
  toggleSignupSettings,
} from "@/app/_actions/dashboard";

// Query keys
const SIGNUP_SETTINGS_KEY = ["signup", "settings"] as const;
const SIGNUP_ANALYTICS_KEY = (start?: string, end?: string) =>
  ["signup", "analytics", start, end] as const;

/**
 * Hook to fetch signup settings
 */
export function useSignupSettings() {
  return useQuery({
    queryKey: SIGNUP_SETTINGS_KEY,
    queryFn: async () => {
      const result = await fetchSignupSettings();
      if (!result.success) throw new Error(result.message);
      return result.data;
    },
    staleTime: 1000 * 60 * 5, // 5 minutes
  });
}

/**
 * Hook to fetch signup analytics
 */
export function useSignupAnalytics(params?: { start?: string; end?: string }) {
  return useQuery({
    queryKey: SIGNUP_ANALYTICS_KEY(params?.start, params?.end),
    queryFn: async () => {
      const result = await fetchSignupAnalytics(params);
      if (!result.success) throw new Error(result.message);
      return result.data;
    },
    staleTime: 1000 * 60 * 5, // 5 minutes
  });
}

/**
 * Hook to toggle signup settings
 */
export function useToggleSignupSettings() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (enabled: boolean) => toggleSignupSettings(enabled),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: SIGNUP_SETTINGS_KEY });
    },
  });
}
