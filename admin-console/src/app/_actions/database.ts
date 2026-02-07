"use server";

import type { APIResponse } from "@/types";
import authenticatedApiClient, {
  handleError,
  successResponse,
} from "./api-config";

export interface DatabaseConnection {
  id: string;
  name: string;
  type: "postgresql" | "mysql" | "mongodb" | "redis" | "elasticsearch";
  host: string;
  port: number;
  database: string;
  username: string;
  is_primary: boolean;
  is_replica: boolean;
  status: "connected" | "disconnected" | "error" | "maintenance";
  connection_pool_size: number;
  active_connections: number;
  max_connections: number;
  last_health_check: string;
  created_at: string;
  updated_at: string;
}

export interface DatabaseMetrics {
  connection_id: string;
  connection_name: string;
  cpu_usage: number;
  memory_usage: number;
  disk_usage: number;
  disk_space_total: number;
  disk_space_used: number;
  disk_space_available: number;
  active_connections: number;
  total_connections: number;
  queries_per_second: number;
  slow_queries: number;
  cache_hit_ratio: number;
  replication_lag: number;
  uptime: number;
  timestamp: string;
}

export interface DatabaseTable {
  id: string;
  connection_id: string;
  schema_name: string;
  table_name: string;
  table_type: "table" | "view" | "materialized_view";
  row_count: number;
  size_bytes: number;
  index_count: number;
  last_analyzed: string;
  created_at: string;
  updated_at: string;
}

export interface DatabaseQuery {
  id: string;
  connection_id: string;
  query_text: string;
  query_hash: string;
  execution_time: number;
  rows_affected: number;
  status: "running" | "completed" | "failed" | "cancelled";
  user_id?: string;
  application: string;
  started_at: string;
  completed_at?: string;
  error_message?: string;
}

export interface DatabaseBackup {
  id: string;
  connection_id: string;
  backup_type: "full" | "incremental" | "differential";
  backup_method: "pg_dump" | "mysqldump" | "mongodump" | "custom";
  file_path: string;
  file_size: number;
  status: "running" | "completed" | "failed" | "cancelled";
  started_at: string;
  completed_at?: string;
  error_message?: string;
  retention_days: number;
  is_automated: boolean;
  created_by?: string;
}

export interface DatabaseMigration {
  id: string;
  connection_id: string;
  migration_name: string;
  version: string;
  status: "pending" | "running" | "completed" | "failed" | "rolled_back";
  migration_type: "schema" | "data" | "index" | "constraint";
  sql_up: string;
  sql_down: string;
  execution_time?: number;
  applied_at?: string;
  rolled_back_at?: string;
  error_message?: string;
  applied_by?: string;
}

export interface DatabaseFilters {
  search?: string;
  connection_id?: string;
  status?: string;
  type?: string;
  time_range?: string;
  start_date?: string;
  end_date?: string;
}

export interface DatabaseStats {
  total_connections: number;
  active_connections: number;
  primary_connections: number;
  replica_connections: number;
  total_databases: number;
  total_tables: number;
  total_size_bytes: number;
  avg_cpu_usage: number;
  avg_memory_usage: number;
  avg_disk_usage: number;
  total_queries_today: number;
  slow_queries_today: number;
  active_backups: number;
  pending_migrations: number;
  connections_by_type: Array<{
    type: string;
    count: number;
    percentage: number;
  }>;
  top_databases_by_size: Array<{
    connection_id: string;
    connection_name: string;
    database_name: string;
    size_bytes: number;
    table_count: number;
  }>;
  recent_slow_queries: Array<{
    query_id: string;
    connection_name: string;
    execution_time: number;
    query_text: string;
    started_at: string;
  }>;
}

/**
 * Get all database connections with filtering
 */
export async function getDatabaseConnections(
  filters?: DatabaseFilters,
): Promise<APIResponse<DatabaseConnection[]>> {
  const params = new URLSearchParams();

  if (filters?.search) params.append("search", filters.search);
  if (filters?.status) params.append("status", filters.status);
  if (filters?.type) params.append("type", filters.type);

  const url = `/api/v1/admin/database/connections${params.toString() ? `?${params.toString()}` : ""}`;

  try {
    const response = await authenticatedApiClient({
      url: url,
      method: "GET",
    });

    return successResponse(
      response?.data?.data || response?.data,
      "Database connections retrieved successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}

/**
 * Get database connection by ID
 */
export async function getDatabaseConnection(
  connectionId: string,
): Promise<APIResponse<DatabaseConnection | null>> {
  const url = `/api/v1/admin/database/connections/${connectionId}`;

  try {
    const response = await authenticatedApiClient({
      url: url,
      method: "GET",
    });

    return successResponse(
      response?.data?.data || response?.data,
      "Database connection retrieved successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}

/**
 * Test database connection
 */
export async function testDatabaseConnection(connectionId: string): Promise<
  APIResponse<{
    success: boolean;
    response_time: number;
    error_message?: string;
    connection_info?: Record<string, any>;
  }>
> {
  const url = `/api/v1/admin/database/connections/${connectionId}/test`;

  try {
    const response = await authenticatedApiClient({
      url: url,
      method: "POST",
    });

    return successResponse(
      response?.data?.data || response?.data,
      "Database connection test completed",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}

/**
 * Get database metrics
 */
export async function getDatabaseMetrics(
  filters?: DatabaseFilters,
): Promise<APIResponse<DatabaseMetrics[]>> {
  const params = new URLSearchParams();

  if (filters?.connection_id)
    params.append("connection_id", filters.connection_id);
  if (filters?.time_range) params.append("time_range", filters.time_range);
  if (filters?.start_date) params.append("start_date", filters.start_date);
  if (filters?.end_date) params.append("end_date", filters.end_date);

  const url = `/api/v1/admin/database/metrics${params.toString() ? `?${params.toString()}` : ""}`;

  try {
    const response = await authenticatedApiClient({
      url: url,
      method: "GET",
    });

    return successResponse(
      response?.data?.data || response?.data,
      "Database metrics retrieved successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}

/**
 * Get database tables
 */
export async function getDatabaseTables(
  connectionId: string,
  filters?: { search?: string; schema?: string },
): Promise<APIResponse<DatabaseTable[]>> {
  const params = new URLSearchParams();

  if (filters?.search) params.append("search", filters.search);
  if (filters?.schema) params.append("schema", filters.schema);

  const url = `/api/v1/admin/database/connections/${connectionId}/tables${params.toString() ? `?${params.toString()}` : ""}`;

  try {
    const response = await authenticatedApiClient({
      url: url,
      method: "GET",
    });

    return successResponse(
      response?.data?.data || response?.data,
      "Database tables retrieved successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}

/**
 * Get database queries
 */
export async function getDatabaseQueries(
  filters?: DatabaseFilters,
): Promise<APIResponse<DatabaseQuery[]>> {
  const params = new URLSearchParams();

  if (filters?.connection_id)
    params.append("connection_id", filters.connection_id);
  if (filters?.status) params.append("status", filters.status);
  if (filters?.time_range) params.append("time_range", filters.time_range);

  const url = `/api/v1/admin/database/queries${params.toString() ? `?${params.toString()}` : ""}`;

  try {
    const response = await authenticatedApiClient({
      url: url,
      method: "GET",
    });

    return successResponse(
      response?.data?.data || response?.data,
      "Database queries retrieved successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}

/**
 * Execute database query
 */
export async function executeDatabaseQuery(
  connectionId: string,
  query: string,
  options?: { limit?: number; timeout?: number },
): Promise<
  APIResponse<{
    columns: string[];
    rows: any[][];
    row_count: number;
    execution_time: number;
    query_id: string;
  }>
> {
  const url = `/api/v1/admin/database/connections/${connectionId}/execute`;

  try {
    const response = await authenticatedApiClient({
      url: url,
      method: "POST",
      data: {
        query,
        limit: options?.limit || 1000,
        timeout: options?.timeout || 30000,
      },
    });

    return successResponse(
      response?.data?.data || response?.data,
      "Query executed successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}

/**
 * Cancel database query
 */
export async function cancelDatabaseQuery(
  queryId: string,
): Promise<APIResponse<void>> {
  const url = `/api/v1/admin/database/queries/${queryId}/cancel`;

  try {
    const response = await authenticatedApiClient({
      url: url,
      method: "POST",
    });

    return successResponse(
      response?.data?.data || response?.data,
      "Query cancelled successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}

/**
 * Get database backups
 */
export async function getDatabaseBackups(
  filters?: DatabaseFilters,
): Promise<APIResponse<DatabaseBackup[]>> {
  const params = new URLSearchParams();

  if (filters?.connection_id)
    params.append("connection_id", filters.connection_id);
  if (filters?.status) params.append("status", filters.status);
  if (filters?.time_range) params.append("time_range", filters.time_range);

  const url = `/api/v1/admin/database/backups${params.toString() ? `?${params.toString()}` : ""}`;

  try {
    const response = await authenticatedApiClient({
      url: url,
      method: "GET",
    });

    return successResponse(
      response?.data?.data || response?.data,
      "Database backups retrieved successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}

/**
 * Create database backup
 */
export async function createDatabaseBackup(
  connectionId: string,
  options: {
    backup_type: "full" | "incremental" | "differential";
    retention_days?: number;
    description?: string;
  },
): Promise<APIResponse<DatabaseBackup>> {
  const url = `/api/v1/admin/database/connections/${connectionId}/backup`;

  try {
    const response = await authenticatedApiClient({
      url: url,
      method: "POST",
      data: options,
    });

    return successResponse(
      response?.data?.data || response?.data,
      "Database backup initiated successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}

/**
 * Restore database backup
 */
export async function restoreDatabaseBackup(
  backupId: string,
  options?: {
    target_connection_id?: string;
    restore_data?: boolean;
    restore_schema?: boolean;
  },
): Promise<APIResponse<{ restore_id: string }>> {
  const url = `/api/v1/admin/database/backups/${backupId}/restore`;

  try {
    const response = await authenticatedApiClient({
      url: url,
      method: "POST",
      data: options || {},
    });

    return successResponse(
      response?.data?.data || response?.data,
      "Database restore initiated successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}

/**
 * Get database migrations
 */
export async function getDatabaseMigrations(
  connectionId: string,
): Promise<APIResponse<DatabaseMigration[]>> {
  const url = `/api/v1/admin/database/connections/${connectionId}/migrations`;

  try {
    const response = await authenticatedApiClient({
      url: url,
      method: "GET",
    });

    return successResponse(
      response?.data?.data || response?.data,
      "Database migrations retrieved successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}

/**
 * Run database migration
 */
export async function runDatabaseMigration(
  connectionId: string,
  migrationId: string,
): Promise<APIResponse<{ execution_id: string }>> {
  const url = `/api/v1/admin/database/connections/${connectionId}/migrations/${migrationId}/run`;

  try {
    const response = await authenticatedApiClient({
      url: url,
      method: "POST",
    });

    return successResponse(
      response?.data?.data || response?.data,
      "Database migration initiated successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}

/**
 * Rollback database migration
 */
export async function rollbackDatabaseMigration(
  connectionId: string,
  migrationId: string,
): Promise<APIResponse<{ rollback_id: string }>> {
  const url = `/api/v1/admin/database/connections/${connectionId}/migrations/${migrationId}/rollback`;

  try {
    const response = await authenticatedApiClient({
      url: url,
      method: "POST",
    });

    return successResponse(
      response?.data?.data || response?.data,
      "Database migration rollback initiated successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}

/**
 * Get database statistics
 */
export async function getDatabaseStats(): Promise<
  APIResponse<DatabaseStats | null>
> {
  const url = "/api/v1/admin/database/stats";

  try {
    const response = await authenticatedApiClient({
      url: url,
      method: "GET",
    });

    return successResponse(
      response?.data?.data || response?.data,
      "Database statistics retrieved successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}

/**
 * Optimize database table
 */
export async function optimizeDatabaseTable(
  connectionId: string,
  tableName: string,
  options?: {
    analyze?: boolean;
    vacuum?: boolean;
    reindex?: boolean;
  },
): Promise<APIResponse<{ optimization_id: string }>> {
  const url = `/api/v1/admin/database/connections/${connectionId}/tables/${tableName}/optimize`;

  try {
    const response = await authenticatedApiClient({
      url: url,
      method: "POST",
      data: options || {},
    });

    return successResponse(
      response?.data?.data || response?.data,
      "Table optimization initiated successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}

/**
 * Export database data
 */
export async function exportDatabaseData(
  connectionId: string,
  options: {
    format: "sql" | "csv" | "json";
    tables?: string[];
    include_schema?: boolean;
    include_data?: boolean;
  },
): Promise<APIResponse<{ download_url: string; expires_at: string }>> {
  const url = `/api/v1/admin/database/connections/${connectionId}/export`;

  try {
    const response = await authenticatedApiClient({
      url: url,
      method: "POST",
      data: options,
    });

    return successResponse(
      response?.data?.data || response?.data,
      "Database export initiated successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}

/**
 * Get database schemas
 */
export async function getDatabaseSchemas(
  connectionId: string,
): Promise<APIResponse<string[]>> {
  const url = `/api/v1/admin/database/connections/${connectionId}/schemas`;

  try {
    const response = await authenticatedApiClient({
      url: url,
      method: "GET",
    });

    return successResponse(
      response?.data?.data || response?.data,
      "Database schemas retrieved successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}

/**
 * Monitor database performance
 */
export async function getDatabasePerformanceMetrics(
  connectionId: string,
  timeRange: string = "24h",
): Promise<APIResponse<DatabaseMetrics[]>> {
  const url = `/api/v1/admin/database/connections/${connectionId}/performance?time_range=${timeRange}`;

  try {
    const response = await authenticatedApiClient({
      url: url,
      method: "GET",
    });

    return successResponse(
      response?.data?.data || response?.data,
      "Database performance metrics retrieved successfully",
    );
  } catch (error: Error | any) {
    return handleError(error);
  }
}
