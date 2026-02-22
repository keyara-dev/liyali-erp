import { useQuery } from "@tanstack/react-query";
import {
  getAdminDashboardMetrics,
  getSystemHealth,
} from "@/app/_actions/dashboard";

export function useDashboardMetrics() {
  return useQuery({
    queryKey: ["dashboard", "metrics"],
    queryFn: async () => {
      const result = await getAdminDashboardMetrics();
      if (!result.success) throw new Error(result.message);
      return result.data;
    },
  });
}

export function useDashboardSystemHealth() {
  return useQuery({
    queryKey: ["dashboard", "system-health"],
    queryFn: async () => {
      const result = await getSystemHealth();
      if (!result.success) throw new Error(result.message);
      return result.data;
    },
    refetchInterval: 60000, // Refresh every 60 seconds
  });
}
