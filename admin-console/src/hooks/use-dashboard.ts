import { useQuery } from "@tanstack/react-query";
import {
  getAdminDashboardMetrics,
  getSystemHealth,
} from "@/app/_actions/dashboard";
import { queryKeys } from "@/lib/query-keys";

export function useDashboardMetrics() {
  return useQuery({
    queryKey: queryKeys.dashboard.metrics(),
    queryFn: async () => {
      const result = await getAdminDashboardMetrics();
      if (!result.success) throw new Error(result.message);
      return result.data;
    },
  });
}

export function useDashboardSystemHealth() {
  return useQuery({
    queryKey: queryKeys.dashboard.systemHealth(),
    queryFn: async () => {
      const result = await getSystemHealth();
      if (!result.success) throw new Error(result.message);
      return result.data;
    },
    refetchInterval: 60000, // Refresh every 60 seconds
  });
}
