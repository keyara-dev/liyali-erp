import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  getDatabaseConnections,
  getDatabaseConnection,
  testDatabaseConnection,
  getDatabaseMetrics,
  getDatabaseTables,
  getDatabaseQueries,
  getDatabaseBackups,
  createDatabaseBackup,
  restoreDatabaseBackup,
  getDatabaseMigrations,
  getDatabaseStats,
  getDatabaseSchemas,
  getDatabasePerformanceMetrics,
  type DatabaseFilters,
} from "@/app/_actions/database";

// --- Query Hooks ---

export function useDatabaseConnections(filters?: DatabaseFilters) {
  return useQuery({
    queryKey: ["database", "connections", filters],
    queryFn: async () => {
      const result = await getDatabaseConnections(filters);
      if (!result.success) throw new Error(result.message);
      return result.data;
    },
  });
}

export function useDatabaseConnection(connectionId: string) {
  return useQuery({
    queryKey: ["database", "connections", connectionId],
    queryFn: async () => {
      const result = await getDatabaseConnection(connectionId);
      if (!result.success) throw new Error(result.message);
      return result.data;
    },
    enabled: !!connectionId,
  });
}

export function useDatabaseMetrics(filters?: DatabaseFilters) {
  return useQuery({
    queryKey: ["database", "metrics", filters],
    queryFn: async () => {
      const result = await getDatabaseMetrics(filters);
      if (!result.success) throw new Error(result.message);
      return result.data;
    },
  });
}

export function useDatabaseTables(
  connectionId: string,
  filters?: { search?: string; schema?: string },
) {
  return useQuery({
    queryKey: ["database", "connections", connectionId, "tables", filters],
    queryFn: async () => {
      const result = await getDatabaseTables(connectionId, filters);
      if (!result.success) throw new Error(result.message);
      return result.data;
    },
    enabled: !!connectionId,
  });
}

export function useDatabaseQueries(filters?: DatabaseFilters) {
  return useQuery({
    queryKey: ["database", "queries", filters],
    queryFn: async () => {
      const result = await getDatabaseQueries(filters);
      if (!result.success) throw new Error(result.message);
      return result.data;
    },
  });
}

export function useDatabaseBackups(filters?: DatabaseFilters) {
  return useQuery({
    queryKey: ["database", "backups", filters],
    queryFn: async () => {
      const result = await getDatabaseBackups(filters);
      if (!result.success) throw new Error(result.message);
      return result.data;
    },
  });
}

export function useDatabaseMigrations(connectionId: string) {
  return useQuery({
    queryKey: ["database", "connections", connectionId, "migrations"],
    queryFn: async () => {
      const result = await getDatabaseMigrations(connectionId);
      if (!result.success) throw new Error(result.message);
      return result.data;
    },
    enabled: !!connectionId,
  });
}

export function useDatabaseStats() {
  return useQuery({
    queryKey: ["database", "stats"],
    queryFn: async () => {
      const result = await getDatabaseStats();
      if (!result.success) throw new Error(result.message);
      return result.data;
    },
  });
}

export function useDatabaseSchemas(connectionId: string) {
  return useQuery({
    queryKey: ["database", "connections", connectionId, "schemas"],
    queryFn: async () => {
      const result = await getDatabaseSchemas(connectionId);
      if (!result.success) throw new Error(result.message);
      return result.data;
    },
    enabled: !!connectionId,
  });
}

export function useDatabasePerformanceMetrics(
  connectionId: string,
  timeRange: string = "24h",
) {
  return useQuery({
    queryKey: [
      "database",
      "connections",
      connectionId,
      "performance",
      timeRange,
    ],
    queryFn: async () => {
      const result = await getDatabasePerformanceMetrics(
        connectionId,
        timeRange,
      );
      if (!result.success) throw new Error(result.message);
      return result.data;
    },
    enabled: !!connectionId,
  });
}

// --- Mutation Hooks ---

export function useTestDatabaseConnection() {
  return useMutation({
    mutationFn: (connectionId: string) =>
      testDatabaseConnection(connectionId),
  });
}

export function useCreateDatabaseBackup() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({
      connectionId,
      options,
    }: {
      connectionId: string;
      options: {
        backup_type: "full" | "incremental" | "differential";
        retention_days?: number;
        description?: string;
      };
    }) => createDatabaseBackup(connectionId, options),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["database", "backups"] });
    },
  });
}

export function useRestoreDatabaseBackup() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({
      backupId,
      options,
    }: {
      backupId: string;
      options?: {
        target_connection_id?: string;
        restore_data?: boolean;
        restore_schema?: boolean;
      };
    }) => restoreDatabaseBackup(backupId, options),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["database"] });
    },
  });
}
