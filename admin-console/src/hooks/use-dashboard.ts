import { useQuery } from "@tanstack/react-query";
import {
  getAdminDashboardMetrics,
  getSystemHealth,
  type AdminDashboardMetrics,
} from "@/app/_actions/dashboard";
import { queryKeys } from "@/lib/query-keys";

const normalizeDashboardMetrics = (
  data: AdminDashboardMetrics | null | undefined,
): AdminDashboardMetrics => ({
  total_organizations: data?.total_organizations ?? 0,
  active_organizations: data?.active_organizations ?? 0,
  trial_organizations: data?.trial_organizations ?? 0,
  expired_trials: data?.expired_trials ?? 0,
  total_users: data?.total_users ?? 0,
  active_users: data?.active_users ?? 0,
  system_health: {
    uptime: data?.system_health?.uptime ?? "0%",
    cpu_usage: data?.system_health?.cpu_usage ?? 0,
    memory_usage: data?.system_health?.memory_usage ?? 0,
    disk_usage: data?.system_health?.disk_usage ?? 0,
  },
  recent_activities: Array.isArray(data?.recent_activities)
    ? data.recent_activities
    : [],
});

export function useDashboardMetrics() {
  return useQuery({
    queryKey: queryKeys.dashboard.metrics(),
    queryFn: async () => {
      const result = await getAdminDashboardMetrics();
      if (!result.success) throw new Error(result.message);
      return normalizeDashboardMetrics(result.data);
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
