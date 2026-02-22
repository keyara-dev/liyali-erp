import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  getAuditLogs,
  getAuditLogStats,
  getAuditLogAnalytics,
  getAuditLogDetails,
  exportAuditLogs,
  getSecurityEvents,
  createAuditLog,
  getAuditLogRetentionSettings,
  updateAuditLogRetentionSettings,
  type AuditLogFilters,
} from "@/app/_actions/audit-logs";

export function useAuditLogs(
  filters?: AuditLogFilters,
  page: number = 1,
  limit: number = 50,
) {
  return useQuery({
    queryKey: ["audit-logs", filters, page, limit],
    queryFn: async () => {
      const result = await getAuditLogs(filters, page, limit);
      if (!result.success) throw new Error(result.message);
      return result.data;
    },
  });
}

export function useAuditLogStats(filters?: AuditLogFilters) {
  return useQuery({
    queryKey: ["audit-logs", "stats", filters],
    queryFn: async () => {
      const result = await getAuditLogStats(filters);
      if (!result.success) throw new Error(result.message);
      return result.data;
    },
  });
}

export function useAuditLogAnalytics(filters?: AuditLogFilters) {
  return useQuery({
    queryKey: ["audit-logs", "analytics", filters],
    queryFn: async () => {
      const result = await getAuditLogAnalytics(filters);
      if (!result.success) throw new Error(result.message);
      return result.data;
    },
  });
}

export function useAuditLogDetails(logId: string) {
  return useQuery({
    queryKey: ["audit-logs", logId],
    queryFn: async () => {
      const result = await getAuditLogDetails(logId);
      if (!result.success) throw new Error(result.message);
      return result.data;
    },
    enabled: !!logId,
  });
}

export function useSecurityEvents(filters?: AuditLogFilters) {
  return useQuery({
    queryKey: ["audit-logs", "security-events", filters],
    queryFn: async () => {
      const result = await getSecurityEvents(filters);
      if (!result.success) throw new Error(result.message);
      return result.data;
    },
  });
}

export function useAuditLogRetentionSettings() {
  return useQuery({
    queryKey: ["audit-logs", "retention-settings"],
    queryFn: async () => {
      const result = await getAuditLogRetentionSettings();
      if (!result.success) throw new Error(result.message);
      return result.data;
    },
  });
}

export function useExportAuditLogs() {
  return useMutation({
    mutationFn: ({
      format,
      filters,
    }: {
      format: "csv" | "json" | "pdf";
      filters?: AuditLogFilters;
    }) => exportAuditLogs(format, filters),
  });
}

export function useCreateAuditLog() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: {
      action: string;
      action_type: string;
      resource_type: string;
      resource_id?: string;
      details: Record<string, any>;
      severity?: string;
    }) => createAuditLog(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["audit-logs"] });
    },
  });
}

export function useUpdateAuditLogRetentionSettings() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (settings: {
      retention_days: number;
      auto_archive: boolean;
      archive_format: string;
      compliance_mode: boolean;
    }) => updateAuditLogRetentionSettings(settings),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ["audit-logs", "retention-settings"],
      });
    },
  });
}
