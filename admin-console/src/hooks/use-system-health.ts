import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  getSystemHealth,
  getSystemMetrics,
  getSystemAlerts,
  getSystemLogs,
  getPerformanceMetrics,
  acknowledgeAlert,
  resolveAlert,
  runHealthCheck,
  getSystemConfiguration,
  updateSystemConfiguration,
  restartService,
  clearSystemCache,
} from "@/app/_actions/system-health";

// --- Query Hooks ---

export function useSystemHealth() {
  return useQuery({
    queryKey: ["system-health"],
    queryFn: async () => {
      const result = await getSystemHealth();
      if (!result.success) throw new Error(result.message);
      return result.data;
    },
    refetchInterval: 30000, // Refresh every 30 seconds
  });
}

export function useSystemMetrics() {
  return useQuery({
    queryKey: ["system-health", "metrics"],
    queryFn: async () => {
      const result = await getSystemMetrics();
      if (!result.success) throw new Error(result.message);
      return result.data;
    },
    refetchInterval: 30000,
  });
}

export function useSystemAlerts(
  status?: "active" | "acknowledged" | "resolved",
  severity?: "low" | "medium" | "high" | "critical",
) {
  return useQuery({
    queryKey: ["system-health", "alerts", status, severity],
    queryFn: async () => {
      const result = await getSystemAlerts(status, severity);
      if (!result.success) throw new Error(result.message);
      return result.data;
    },
  });
}

export function useSystemLogs(
  level?: "debug" | "info" | "warn" | "error" | "fatal",
  component?: string,
  page: number = 1,
  limit: number = 100,
) {
  return useQuery({
    queryKey: ["system-health", "logs", level, component, page, limit],
    queryFn: async () => {
      const result = await getSystemLogs(level, component, page, limit);
      if (!result.success) throw new Error(result.message);
      return result.data;
    },
  });
}

export function usePerformanceMetrics() {
  return useQuery({
    queryKey: ["system-health", "performance"],
    queryFn: async () => {
      const result = await getPerformanceMetrics();
      if (!result.success) throw new Error(result.message);
      return result.data;
    },
    refetchInterval: 30000,
  });
}

export function useSystemConfiguration() {
  return useQuery({
    queryKey: ["system-health", "config"],
    queryFn: async () => {
      const result = await getSystemConfiguration();
      if (!result.success) throw new Error(result.message);
      return result.data;
    },
  });
}

// --- Mutation Hooks ---

export function useAcknowledgeAlert() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (alertId: string) => acknowledgeAlert(alertId),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ["system-health", "alerts"],
      });
    },
  });
}

export function useResolveAlert() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (alertId: string) => resolveAlert(alertId),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ["system-health", "alerts"],
      });
    },
  });
}

export function useRunHealthCheck() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: () => runHealthCheck(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["system-health"] });
    },
  });
}

export function useUpdateSystemConfiguration() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (config: Record<string, any>) =>
      updateSystemConfiguration(config),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ["system-health", "config"],
      });
    },
  });
}

export function useRestartService() {
  return useMutation({
    mutationFn: (serviceName: string) => restartService(serviceName),
  });
}

export function useClearSystemCache() {
  return useMutation({
    mutationFn: () => clearSystemCache(),
  });
}
