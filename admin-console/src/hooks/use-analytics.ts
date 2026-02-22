import { useQuery, useMutation } from "@tanstack/react-query";
import {
  getAnalyticsOverview,
  getUserAnalytics,
  getOrganizationAnalytics,
  getRevenueAnalytics,
  getUsageAnalytics,
  exportAnalyticsReport,
  type AnalyticsFilters,
} from "@/app/_actions/analytics";

export function useAnalyticsOverview(filters?: AnalyticsFilters) {
  return useQuery({
    queryKey: ["analytics", "overview", filters],
    queryFn: async () => {
      const result = await getAnalyticsOverview(filters);
      if (!result.success) throw new Error(result.message);
      return result.data;
    },
  });
}

export function useUserAnalytics(filters?: AnalyticsFilters) {
  return useQuery({
    queryKey: ["analytics", "users", filters],
    queryFn: async () => {
      const result = await getUserAnalytics(filters);
      if (!result.success) throw new Error(result.message);
      return result.data;
    },
  });
}

export function useOrganizationAnalytics(filters?: AnalyticsFilters) {
  return useQuery({
    queryKey: ["analytics", "organizations", filters],
    queryFn: async () => {
      const result = await getOrganizationAnalytics(filters);
      if (!result.success) throw new Error(result.message);
      return result.data;
    },
  });
}

export function useRevenueAnalytics(filters?: AnalyticsFilters) {
  return useQuery({
    queryKey: ["analytics", "revenue", filters],
    queryFn: async () => {
      const result = await getRevenueAnalytics(filters);
      if (!result.success) throw new Error(result.message);
      return result.data;
    },
  });
}

export function useUsageAnalytics(filters?: AnalyticsFilters) {
  return useQuery({
    queryKey: ["analytics", "usage", filters],
    queryFn: async () => {
      const result = await getUsageAnalytics(filters);
      if (!result.success) throw new Error(result.message);
      return result.data;
    },
  });
}

export function useExportAnalyticsReport() {
  return useMutation({
    mutationFn: ({
      type,
      format,
      filters,
    }: {
      type: "overview" | "users" | "organizations" | "revenue" | "usage";
      format: "csv" | "pdf" | "excel";
      filters?: AnalyticsFilters;
    }) => exportAnalyticsReport(type, format, filters),
  });
}
