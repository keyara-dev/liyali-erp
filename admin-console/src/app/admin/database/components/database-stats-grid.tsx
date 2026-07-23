"use client";

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import {
  Database,
  Server,
  HardDrive,
  Activity,
  Clock,
  AlertTriangle,
  CheckCircle,
  BarChart3,
} from "lucide-react";
import {
  PieChart,
  Pie,
  Cell,
  ResponsiveContainer,
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
} from "recharts";
import type { DatabaseStats } from "@/app/_actions/database";

interface DatabaseStatsGridProps {
  stats: DatabaseStats | null;
  isLoading: boolean;
}

const COLORS = ["#0088FE", "#00C49F", "#FFBB28", "#FF8042", "#8884D8"];

export function DatabaseStatsGrid({
  stats,
  isLoading,
}: DatabaseStatsGridProps) {
  if (isLoading) {
    return (
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        {[...Array(8)].map((_, i) => (
          <Card key={i}>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <div className="h-4 w-20 bg-muted animate-pulse rounded" />
              <div className="h-4 w-4 bg-muted animate-pulse rounded" />
            </CardHeader>
            <CardContent>
              <div className="h-8 w-16 bg-muted animate-pulse rounded mb-2" />
              <div className="h-3 w-24 bg-muted animate-pulse rounded" />
            </CardContent>
          </Card>
        ))}
      </div>
    );
  }

  if (!stats) {
    return (
      <Card>
        <CardContent className="flex items-center justify-center h-32">
          <p className="text-muted-foreground">No statistics available</p>
        </CardContent>
      </Card>
    );
  }

  const connectionTypeData = (stats.connections_by_type ?? []).map((item) => ({
    name: item.type,
    value: item.count,
    percentage: item.percentage,
  }));

  const topDatabasesData = (stats.top_databases_by_size ?? []).slice(0, 5);

  const getSlowQueryKey = (
    query: DatabaseStats["recent_slow_queries"][number],
    index: number,
  ) =>
    query.query_id?.trim() ||
    `${query.connection_name || "connection"}-${query.started_at || "started"}-${index}`;

  const formatBytes = (bytes: number) => {
    if (bytes === 0) return "0 Bytes";
    const k = 1024;
    const sizes = ["Bytes", "KB", "MB", "GB", "TB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
  };

  return (
    <div className="space-y-6">
      {/* Main Stats Grid */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              Total Connections
            </CardTitle>
            <Database className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.total_connections}</div>
            <p className="text-xs text-muted-foreground">
              {stats.active_connections} active
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              Primary/Replica
            </CardTitle>
            <Server className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {stats.primary_connections}/{stats.replica_connections}
            </div>
            <p className="text-xs text-muted-foreground">
              Primary/Replica split
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              Total Databases
            </CardTitle>
            <Database className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.total_databases}</div>
            <p className="text-xs text-muted-foreground">
              {stats.total_tables} tables
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Size</CardTitle>
            <HardDrive className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {formatBytes(stats.total_size_bytes)}
            </div>
            <p className="text-xs text-muted-foreground">
              Across all databases
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Avg CPU Usage</CardTitle>
            <Activity className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {stats.avg_cpu_usage.toFixed(1)}%
            </div>
            <p className="text-xs text-muted-foreground">
              Across all connections
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              Avg Memory Usage
            </CardTitle>
            <BarChart3 className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {stats.avg_memory_usage.toFixed(1)}%
            </div>
            <p className="text-xs text-muted-foreground">Memory utilization</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Queries Today</CardTitle>
            <Activity className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {stats.total_queries_today.toLocaleString()}
            </div>
            <p className="text-xs text-muted-foreground">
              {stats.slow_queries_today} slow queries
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Operations</CardTitle>
            <Clock className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="flex items-center gap-4">
              <div className="flex items-center gap-1">
                <CheckCircle className="h-4 w-4 text-green-600" />
                <span className="text-sm">{stats.active_backups} backups</span>
              </div>
              <div className="flex items-center gap-1">
                <AlertTriangle className="h-4 w-4 text-orange-600" />
                <span className="text-sm">
                  {stats.pending_migrations} migrations
                </span>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Charts Grid */}
      <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
        {/* Connection Types Distribution */}
        <Card className="col-span-1">
          <CardHeader>
            <CardTitle className="text-lg">Connection Types</CardTitle>
            <CardDescription>Distribution by database type</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="h-64">
              <ResponsiveContainer width="100%" height="100%">
                <PieChart>
                  <Pie
                    data={connectionTypeData}
                    cx="50%"
                    cy="50%"
                    labelLine={false}
                    label={({ name, percentage }) =>
                      `${name} (${percentage.toFixed(1)}%)`
                    }
                    outerRadius={80}
                    fill="#8884d8"
                    dataKey="value"
                  >
                    {connectionTypeData.map((entry, index) => (
                      <Cell
                        key={`cell-${index}`}
                        fill={COLORS[index % COLORS.length]}
                      />
                    ))}
                  </Pie>
                  <Tooltip />
                </PieChart>
              </ResponsiveContainer>
            </div>
          </CardContent>
        </Card>

        {/* Top Databases by Size */}
        <Card className="col-span-2">
          <CardHeader>
            <CardTitle className="text-lg">Top Databases by Size</CardTitle>
            <CardDescription>
              Largest databases by storage usage
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="h-64">
              <ResponsiveContainer width="100%" height="100%">
                <BarChart data={topDatabasesData}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis
                    dataKey="database_name"
                    angle={-45}
                    textAnchor="end"
                    height={80}
                  />
                  <YAxis tickFormatter={(value) => formatBytes(value)} />
                  <Tooltip
                    formatter={(value: number) => [formatBytes(value), "Size"]}
                  />
                  <Bar dataKey="size_bytes" fill="#0088FE" />
                </BarChart>
              </ResponsiveContainer>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Recent Slow Queries */}
      {(stats.recent_slow_queries ?? []).length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle className="text-lg">Recent Slow Queries</CardTitle>
            <CardDescription>
              Queries with high execution times requiring attention
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {(stats.recent_slow_queries ?? []).map((query, index) => (
                <div
                  key={getSlowQueryKey(query, index)}
                  className="flex items-center justify-between p-3 border rounded-lg"
                >
                  <div className="flex-1">
                    <div className="flex items-center gap-2">
                      <Badge variant="outline" className="text-xs">
                        {query.connection_name}
                      </Badge>
                      <span className="text-sm font-medium">
                        {query.execution_time.toFixed(0)}ms
                      </span>
                    </div>
                    <p className="text-sm text-muted-foreground mt-1 font-mono">
                      {query.query_text.length > 100
                        ? `${query.query_text.substring(0, 100)}...`
                        : query.query_text}
                    </p>
                    <p className="text-xs text-muted-foreground">
                      {new Date(query.started_at).toLocaleString()}
                    </p>
                  </div>
                  <div className="text-right">
                    <div className="text-lg font-bold text-orange-600">
                      #{index + 1}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}

      {/* Database Details */}
      <div className="grid gap-6 md:grid-cols-2">
        {/* Connection Status Summary */}
        <Card>
          <CardHeader>
            <CardTitle className="text-lg">Connection Status</CardTitle>
            <CardDescription>
              Current status of all database connections
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <CheckCircle className="h-4 w-4 text-green-600" />
                  <span className="text-sm">Connected</span>
                </div>
                <Badge variant="default">{stats.active_connections}</Badge>
              </div>
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <AlertTriangle className="h-4 w-4 text-red-600" />
                  <span className="text-sm">Issues</span>
                </div>
                <Badge variant="destructive">
                  {stats.total_connections - stats.active_connections}
                </Badge>
              </div>
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <Database className="h-4 w-4 text-blue-600" />
                  <span className="text-sm">Primary</span>
                </div>
                <Badge variant="outline">{stats.primary_connections}</Badge>
              </div>
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <Server className="h-4 w-4 text-gray-600" />
                  <span className="text-sm">Replica</span>
                </div>
                <Badge variant="secondary">{stats.replica_connections}</Badge>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Resource Utilization */}
        <Card>
          <CardHeader>
            <CardTitle className="text-lg">Resource Utilization</CardTitle>
            <CardDescription>
              Average resource usage across all connections
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <div>
                <div className="flex items-center justify-between mb-2">
                  <span className="text-sm">CPU Usage</span>
                  <span className="text-sm font-medium">
                    {stats.avg_cpu_usage.toFixed(1)}%
                  </span>
                </div>
                <div className="w-full bg-gray-200 rounded-full h-2">
                  <div
                    className={`h-2 rounded-full ${
                      stats.avg_cpu_usage > 80
                        ? "bg-red-600"
                        : stats.avg_cpu_usage > 60
                          ? "bg-orange-600"
                          : "bg-green-600"
                    }`}
                    style={{ width: `${Math.min(stats.avg_cpu_usage, 100)}%` }}
                  ></div>
                </div>
              </div>

              <div>
                <div className="flex items-center justify-between mb-2">
                  <span className="text-sm">Memory Usage</span>
                  <span className="text-sm font-medium">
                    {stats.avg_memory_usage.toFixed(1)}%
                  </span>
                </div>
                <div className="w-full bg-gray-200 rounded-full h-2">
                  <div
                    className={`h-2 rounded-full ${
                      stats.avg_memory_usage > 80
                        ? "bg-red-600"
                        : stats.avg_memory_usage > 60
                          ? "bg-orange-600"
                          : "bg-green-600"
                    }`}
                    style={{
                      width: `${Math.min(stats.avg_memory_usage, 100)}%`,
                    }}
                  ></div>
                </div>
              </div>

              <div>
                <div className="flex items-center justify-between mb-2">
                  <span className="text-sm">Disk Usage</span>
                  <span className="text-sm font-medium">
                    {stats.avg_disk_usage.toFixed(1)}%
                  </span>
                </div>
                <div className="w-full bg-gray-200 rounded-full h-2">
                  <div
                    className={`h-2 rounded-full ${
                      stats.avg_disk_usage > 80
                        ? "bg-red-600"
                        : stats.avg_disk_usage > 60
                          ? "bg-orange-600"
                          : "bg-green-600"
                    }`}
                    style={{ width: `${Math.min(stats.avg_disk_usage, 100)}%` }}
                  ></div>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
