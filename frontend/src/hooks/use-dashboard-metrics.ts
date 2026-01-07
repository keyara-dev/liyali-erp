"use client";

import { useQuery } from "@tanstack/react-query";
import { getDashboardMetrics } from "@/app/_actions/dashboard";
import { usePendingApprovalCount } from "./use-approval-workflow";
import type { DashboardMetrics } from "@/types";

/**
 * Hook for fetching dashboard metrics with enhanced pending approval count
 * 
 * Combines dashboard metrics from the backend with real-time pending approval count
 * from the approval workflow system. Uses React Query for caching and error handling.
 *
 * @returns {Object} Object with metrics data, loading state, and error
 *
 * @example
 * ```typescript
 * const { data: metrics, isLoading, error } = useDashboardMetrics();
 *
 * if (isLoading) return <div>Loading...</div>;
 * if (error) return <div>Error: {error}</div>;
 *
 * return <DashboardView metrics={metrics} />;
 * ```
 */
export function useDashboardMetrics() {
  // Fetch pending approval count from approval workflow
  const { data: pendingCount = 0 } = usePendingApprovalCount();

  return useQuery({
    queryKey: ['dashboard-metrics', pendingCount],
    queryFn: async (): Promise<DashboardMetrics> => {
      const result = await getDashboardMetrics();
      
      if (!result.success || !result.data) {
        throw new Error(result.message || "Failed to load dashboard metrics");
      }

      // Enhance metrics with real pending approval count from backend
      return {
        ...result.data,
        pendingApproval: pendingCount,
      };
    },
    staleTime: 2 * 60 * 1000, // 2 minutes
    gcTime: 5 * 60 * 1000, // 5 minutes (formerly cacheTime)
    retry: 2,
    retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 10000),
  });
}