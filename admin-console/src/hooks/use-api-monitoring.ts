import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  getAPIEndpoints,
  getAPIEndpoint,
  getAPIMetrics,
  getEndpointMetrics,
  getAPIErrors,
  getAPIError,
  resolveAPIError,
  getAPIAlerts,
  acknowledgeAPIAlert,
  resolveAPIAlert,
  getAPIStats,
  getAPIPerformanceData,
  testAPIEndpoint,
  updateEndpointConfig,
  exportAPIData,
  getAPICategories,
  getRealTimeMetrics,
  type APIFilters,
} from "@/app/_actions/api-monitoring";
import { queryKeys } from "@/lib/query-keys";

// --- Query Hooks ---

export function useAPIEndpoints(filters?: APIFilters) {
  return useQuery({
    queryKey: queryKeys.apiMonitoring.endpoints(filters),
    queryFn: async () => {
      const result = await getAPIEndpoints(filters);
      if (!result.success) throw new Error(result.message);
      return result.data;
    },
  });
}

export function useAPIEndpoint(endpointId: string) {
  return useQuery({
    queryKey: queryKeys.apiMonitoring.endpoint(endpointId),
    queryFn: async () => {
      const result = await getAPIEndpoint(endpointId);
      if (!result.success) throw new Error(result.message);
      return result.data;
    },
    enabled: !!endpointId,
  });
}

export function useAPIMetrics(filters?: APIFilters) {
  return useQuery({
    queryKey: queryKeys.apiMonitoring.metrics(filters),
    queryFn: async () => {
      const result = await getAPIMetrics(filters);
      if (!result.success) throw new Error(result.message);
      return result.data;
    },
  });
}

export function useEndpointMetrics(
  endpointId: string,
  timeRange: string = "24h",
) {
  return useQuery({
    queryKey: queryKeys.apiMonitoring.endpointMetrics(endpointId, timeRange),
    queryFn: async () => {
      const result = await getEndpointMetrics(endpointId, timeRange);
      if (!result.success) throw new Error(result.message);
      return result.data;
    },
    enabled: !!endpointId,
  });
}

export function useAPIErrors(filters?: APIFilters) {
  return useQuery({
    queryKey: queryKeys.apiMonitoring.errors(filters),
    queryFn: async () => {
      const result = await getAPIErrors(filters);
      if (!result.success) throw new Error(result.message);
      return result.data;
    },
  });
}

export function useAPIError(errorId: string) {
  return useQuery({
    queryKey: queryKeys.apiMonitoring.error(errorId),
    queryFn: async () => {
      const result = await getAPIError(errorId);
      if (!result.success) throw new Error(result.message);
      return result.data;
    },
    enabled: !!errorId,
  });
}

export function useAPIAlerts(filters?: APIFilters) {
  return useQuery({
    queryKey: queryKeys.apiMonitoring.alerts(filters),
    queryFn: async () => {
      const result = await getAPIAlerts(filters);
      if (!result.success) throw new Error(result.message);
      return result.data;
    },
  });
}

export function useAPIStats() {
  return useQuery({
    queryKey: queryKeys.apiMonitoring.stats(),
    queryFn: async () => {
      const result = await getAPIStats();
      if (!result.success) throw new Error(result.message);
      return result.data;
    },
  });
}

export function useAPIPerformanceData(
  timeRange: string = "24h",
  interval: string = "5m",
) {
  return useQuery({
    queryKey: queryKeys.apiMonitoring.performance(timeRange, interval),
    queryFn: async () => {
      const result = await getAPIPerformanceData(timeRange, interval);
      if (!result.success) throw new Error(result.message);
      return result.data;
    },
  });
}

export function useAPICategories() {
  return useQuery({
    queryKey: queryKeys.apiMonitoring.categories(),
    queryFn: async () => {
      const result = await getAPICategories();
      if (!result.success) throw new Error(result.message);
      return result.data;
    },
  });
}

export function useRealTimeMetrics() {
  return useQuery({
    queryKey: queryKeys.apiMonitoring.realtime(),
    queryFn: async () => {
      const result = await getRealTimeMetrics();
      if (!result.success) throw new Error(result.message);
      return result.data;
    },
    refetchInterval: 10000, // Refresh every 10 seconds
  });
}

// --- Mutation Hooks ---

export function useResolveAPIError() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({
      errorId,
      notes,
    }: {
      errorId: string;
      notes?: string;
    }) => resolveAPIError(errorId, notes),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: queryKeys.apiMonitoring.all,
      });
    },
  });
}

export function useAcknowledgeAPIAlert() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ alertId, notes }: { alertId: string; notes?: string }) =>
      acknowledgeAPIAlert(alertId, notes),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: queryKeys.apiMonitoring.alerts(),
      });
    },
  });
}

export function useResolveAPIAlert() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({
      alertId,
      notes,
    }: {
      alertId: string;
      notes?: string;
    }) => resolveAPIAlert(alertId, notes),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: queryKeys.apiMonitoring.alerts(),
      });
    },
  });
}

export function useTestAPIEndpoint() {
  return useMutation({
    mutationFn: ({
      endpointId,
      testData,
    }: {
      endpointId: string;
      testData?: Record<string, any>;
    }) => testAPIEndpoint(endpointId, testData),
  });
}

export function useUpdateEndpointConfig() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({
      endpointId,
      config,
    }: {
      endpointId: string;
      config: {
        rate_limit?: number;
        timeout?: number;
        is_deprecated?: boolean;
        description?: string;
      };
    }) => updateEndpointConfig(endpointId, config),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: queryKeys.apiMonitoring.endpoints(),
      });
    },
  });
}

export function useExportAPIData() {
  return useMutation({
    mutationFn: ({
      type,
      format,
      filters,
    }: {
      type: "endpoints" | "metrics" | "errors" | "alerts";
      format: "csv" | "json" | "excel";
      filters?: APIFilters;
    }) => exportAPIData(type, format, filters),
  });
}
