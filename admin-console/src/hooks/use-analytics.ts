import { useQuery, useMutation } from "@tanstack/react-query";
import {
  getAnalyticsOverview,
  getUserAnalytics,
  getOrganizationAnalytics,
  getRevenueAnalytics,
  getUsageAnalytics,
  exportAnalyticsReport,
  type AnalyticsFilters,
  type UsageAnalytics,
} from "@/app/_actions/analytics";
import { queryKeys } from "@/lib/query-keys";

const normalizeUsageAnalytics = (
  data: UsageAnalytics | null | undefined,
): UsageAnalytics => ({
  total_api_requests: data?.total_api_requests ?? 0,
  active_sessions: data?.active_sessions ?? 0,
  feature_usage: Array.isArray(data?.feature_usage)
    ? data.feature_usage
    : [],
  usage_trends: Array.isArray(data?.usage_trends) ? data.usage_trends : [],
  performance_metrics: {
    average_response_time:
      data?.performance_metrics?.average_response_time ?? 0,
    error_rate: data?.performance_metrics?.error_rate ?? 0,
    uptime_percentage: data?.performance_metrics?.uptime_percentage ?? 100,
    peak_concurrent_users:
      data?.performance_metrics?.peak_concurrent_users ?? 0,
  },
});

export function useAnalyticsOverview(filters?: AnalyticsFilters) {
  return useQuery({
    queryKey: queryKeys.analytics.overview(filters),
    queryFn: async () => {
      const result = await getAnalyticsOverview(filters);
      if (!result.success) throw new Error(result.message);

      // Provide default values to prevent UI breaks
      return (
        result.data || {
          total_users: 0,
          total_organizations: 0,
          total_revenue: 0,
          active_subscriptions: 0,
          growth_metrics: {
            user_growth_rate: 0,
            organization_growth_rate: 0,
            revenue_growth_rate: 0,
            churn_rate: 0,
          },
          key_metrics: {
            monthly_active_users: 0,
            average_session_duration: 0,
            feature_adoption_rate: 0,
            customer_satisfaction_score: 0,
          },
        }
      );
    },
    retry: 2,
    retryDelay: 1000,
  });
}

export function useUserAnalytics(filters?: AnalyticsFilters) {
  return useQuery({
    queryKey: queryKeys.analytics.users(filters),
    queryFn: async () => {
      const result = await getUserAnalytics(filters);
      if (!result.success) throw new Error(result.message);

      // Provide default values to prevent UI breaks
      return (
        result.data || {
          total_users: 0,
          active_users: 0,
          new_users_this_period: 0,
          user_growth_trend: [],
          user_demographics: {
            by_role: [],
            by_status: [],
            by_organization_size: [],
          },
          engagement_metrics: {
            daily_active_users: 0,
            weekly_active_users: 0,
            monthly_active_users: 0,
            average_session_duration: 0,
            sessions_per_user: 0,
          },
        }
      );
    },
    retry: 2,
    retryDelay: 1000,
  });
}

export function useOrganizationAnalytics(filters?: AnalyticsFilters) {
  return useQuery({
    queryKey: queryKeys.analytics.organizations(filters),
    queryFn: async () => {
      const result = await getOrganizationAnalytics(filters);
      if (!result.success) throw new Error(result.message);

      // Provide default values to prevent UI breaks
      return (
        result.data || {
          total_organizations: 0,
          active_organizations: 0,
          new_organizations_this_period: 0,
          organization_growth_trend: [],
          organization_distribution: {
            by_subscription_tier: [],
            by_status: [],
            by_user_count: [],
          },
          trial_metrics: {
            trial_organizations: 0,
            trial_conversion_rate: 0,
            average_trial_duration: 0,
            trials_expiring_soon: 0,
          },
        }
      );
    },
    retry: 2,
    retryDelay: 1000,
  });
}

export function useRevenueAnalytics(filters?: AnalyticsFilters) {
  return useQuery({
    queryKey: queryKeys.analytics.revenue(filters),
    queryFn: async () => {
      const result = await getRevenueAnalytics(filters);
      if (!result.success) throw new Error(result.message);

      // Provide default values to prevent UI breaks
      return (
        result.data || {
          total_revenue: 0,
          monthly_recurring_revenue: 0,
          annual_recurring_revenue: 0,
          revenue_growth_rate: 0,
          revenue_trend: [],
          revenue_by_tier: [],
          financial_metrics: {
            average_revenue_per_user: 0,
            customer_lifetime_value: 0,
            churn_rate: 0,
            net_revenue_retention: 0,
          },
        }
      );
    },
    retry: 2,
    retryDelay: 1000,
  });
}

export function useUsageAnalytics(filters?: AnalyticsFilters) {
  return useQuery({
    queryKey: queryKeys.analytics.usage(filters),
    queryFn: async () => {
      const result = await getUsageAnalytics(filters);
      if (!result.success) throw new Error(result.message);

      // Provide default values to prevent UI breaks
      return normalizeUsageAnalytics(result.data);
    },
    retry: 2,
    retryDelay: 1000,
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
