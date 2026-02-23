"use client";

import { useState, useEffect } from "react";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  Database,
  RefreshCw,
  Activity,
  HardDrive,
  Search,
} from "lucide-react";
import { toast } from "sonner";
import {
  exportDatabaseData,
  type DatabaseFilters,
} from "@/app/_actions/database";
import {
  useDatabaseConnections,
  useDatabaseMetrics,
  useDatabaseQueries,
  useDatabaseBackups,
  useDatabaseStats,
} from "@/hooks/use-database";
import { DatabaseFiltersComponent } from "./components/database-filters";
import { DatabaseStatsGrid } from "./components/database-stats-grid";
import { DatabaseConnectionsTable } from "./components/database-connections-table";
import { DatabaseQueryPanel } from "./components/database-query-panel";
import { DatabaseBackupsPanel } from "./components/database-backups-panel";

export default function DatabasePage() {
  const [activeTab, setActiveTab] = useState("overview");
  const [filters, setFilters] = useState<DatabaseFilters>({
    time_range: "24h",
  });
  const [searchTerm, setSearchTerm] = useState("");

  // TanStack Query hooks
  const {
    data: connections = [],
    isLoading: isLoadingConnections,
    refetch: refetchConnections,
    isRefetching,
  } = useDatabaseConnections(filters);
  const { data: metrics = [], refetch: refetchMetrics } =
    useDatabaseMetrics(filters);
  const { data: queries = [], refetch: refetchQueries } =
    useDatabaseQueries(filters);
  const { data: backups = [], refetch: refetchBackups } =
    useDatabaseBackups(filters);
  const { data: stats, refetch: refetchStats } = useDatabaseStats();

  const isLoading = isLoadingConnections;
  const isRefreshing = isRefetching;

  useEffect(() => {
    const delayedSearch = setTimeout(() => {
      if (searchTerm !== (filters.search || "")) {
        setFilters((prev) => ({ ...prev, search: searchTerm || undefined }));
      }
    }, 500);

    return () => clearTimeout(delayedSearch);
  }, [searchTerm]);

  const handleRefresh = () => {
    refetchConnections();
    refetchMetrics();
    refetchQueries();
    refetchBackups();
    refetchStats();
  };

  const handleFiltersChange = (newFilters: DatabaseFilters) => {
    setFilters(newFilters);
  };

  const handleResetFilters = () => {
    setFilters({ time_range: "24h" });
    setSearchTerm("");
  };

  const handleExport = async (
    connectionId: string,
    format: "sql" | "csv" | "json",
  ) => {
    try {
      const result = await exportDatabaseData(connectionId, {
        format,
        include_schema: true,
        include_data: true,
      });

      if (result.success) {
        toast.success(
          `Database export initiated. Download will be available shortly.`,
        );
        if (result.data?.download_url) {
          window.open(result.data.download_url, "_blank");
        }
      } else {
        toast.error("Failed to export database");
      }
    } catch (error) {
      console.error("Error exporting database:", error);
      toast.error("Failed to export database");
    }
  };

  const handleDataUpdated = () => {
    handleRefresh();
  };

  return (
    <div className="flex-1 space-y-4 p-4 md:p-8 pt-6">
      <div className="flex items-center justify-between space-y-2">
        <div>
          <h2 className="text-3xl font-bold tracking-tight">
            Database Management
          </h2>
          <p className="text-muted-foreground">
            Monitor and manage database connections, queries, and backups
          </p>
        </div>
        <div className="flex items-center space-x-2">
          <Button
            variant="outline"
            size="sm"
            onClick={handleRefresh}
            disabled={isRefreshing}
          >
            <RefreshCw
              className={`mr-2 h-4 w-4 ${isRefreshing ? "animate-spin" : ""}`}
            />
            Refresh
          </Button>
        </div>
      </div>

      {/* Filters */}
      <DatabaseFiltersComponent
        filters={filters}
        onFiltersChange={handleFiltersChange}
        onReset={handleResetFilters}
        onExport={handleExport}
        searchTerm={searchTerm}
        onSearchChange={setSearchTerm}
        connections={connections}
      />

      {/* Stats Grid */}
      <DatabaseStatsGrid stats={stats ?? null} isLoading={isLoading} />

      {/* Main Content Tabs */}
      <Tabs
        value={activeTab}
        onValueChange={setActiveTab}
        className="space-y-4"
      >
        <TabsList className="grid w-full grid-cols-4">
          <TabsTrigger value="overview" className="flex items-center gap-2">
            <Database className="h-4 w-4" />
            Overview
          </TabsTrigger>
          <TabsTrigger value="connections" className="flex items-center gap-2">
            <Activity className="h-4 w-4" />
            Connections
          </TabsTrigger>
          <TabsTrigger value="queries" className="flex items-center gap-2">
            <Search className="h-4 w-4" />
            Queries
          </TabsTrigger>
          <TabsTrigger value="backups" className="flex items-center gap-2">
            <HardDrive className="h-4 w-4" />
            Backups
          </TabsTrigger>
        </TabsList>

        <TabsContent value="overview" className="space-y-4">
          <div className="grid gap-6">
            {/* Connection Status Overview */}
            <div className="grid gap-6 md:grid-cols-2">
              <Card>
                <CardHeader>
                  <CardTitle className="text-lg">Connection Health</CardTitle>
                  <CardDescription>
                    Current status of database connections
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  {connections.length === 0 ? (
                    <div className="text-center py-8">
                      <Database className="h-8 w-8 text-muted-foreground mx-auto mb-2" />
                      <p className="text-sm text-muted-foreground">
                        No connections found
                      </p>
                    </div>
                  ) : (
                    <div className="space-y-3">
                      {connections.slice(0, 5).map((connection) => (
                        <div
                          key={connection.id}
                          className="flex items-center justify-between p-3 border rounded-lg"
                        >
                          <div className="flex items-center gap-3">
                            <Database className="h-4 w-4 text-muted-foreground" />
                            <div>
                              <p className="font-medium">{connection.name}</p>
                              <p className="text-xs text-muted-foreground">
                                {connection.type} • {connection.host}:
                                {connection.port}
                              </p>
                            </div>
                          </div>
                          <div className="flex items-center gap-2">
                            <span
                              className={`h-2 w-2 rounded-full ${
                                connection.status === "connected"
                                  ? "bg-green-600"
                                  : connection.status === "error"
                                    ? "bg-red-600"
                                    : "bg-gray-400"
                              }`}
                            />
                            <span className="text-xs capitalize">
                              {connection.status}
                            </span>
                          </div>
                        </div>
                      ))}
                      {connections.length > 5 && (
                        <p className="text-xs text-muted-foreground text-center">
                          ... and {connections.length - 5} more connections
                        </p>
                      )}
                    </div>
                  )}
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle className="text-lg">Recent Activity</CardTitle>
                  <CardDescription>
                    Latest database queries and operations
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  {queries.length === 0 ? (
                    <div className="text-center py-8">
                      <Search className="h-8 w-8 text-muted-foreground mx-auto mb-2" />
                      <p className="text-sm text-muted-foreground">
                        No recent queries
                      </p>
                    </div>
                  ) : (
                    <div className="space-y-3">
                      {queries.slice(0, 5).map((query) => {
                        const connection = connections.find(
                          (c) => c.id === query.connection_id,
                        );
                        return (
                          <div
                            key={query.id}
                            className="flex items-center justify-between p-3 border rounded-lg"
                          >
                            <div className="flex-1">
                              <div className="flex items-center gap-2">
                                <span className="text-sm font-medium">
                                  {connection?.name || "Unknown"}
                                </span>
                                <span
                                  className={`text-xs px-2 py-1 rounded ${
                                    query.status === "completed"
                                      ? "bg-green-100 text-green-800"
                                      : query.status === "failed"
                                        ? "bg-red-100 text-red-800"
                                        : query.status === "running"
                                          ? "bg-blue-100 text-blue-800"
                                          : "bg-gray-100 text-gray-800"
                                  }`}
                                >
                                  {query.status}
                                </span>
                              </div>
                              <p className="text-xs text-muted-foreground font-mono mt-1">
                                {query.query_text.length > 60
                                  ? `${query.query_text.substring(0, 60)}...`
                                  : query.query_text}
                              </p>
                            </div>
                            <div className="text-xs text-muted-foreground">
                              {query.execution_time
                                ? `${query.execution_time.toFixed(0)}ms`
                                : new Date(
                                    query.started_at,
                                  ).toLocaleTimeString()}
                            </div>
                          </div>
                        );
                      })}
                    </div>
                  )}
                </CardContent>
              </Card>
            </div>

            {/* Recent Backups */}
            <Card>
              <CardHeader>
                <CardTitle className="text-lg">Recent Backups</CardTitle>
                <CardDescription>
                  Latest database backup operations
                </CardDescription>
              </CardHeader>
              <CardContent>
                {backups.length === 0 ? (
                  <div className="text-center py-8">
                    <HardDrive className="h-8 w-8 text-muted-foreground mx-auto mb-2" />
                    <p className="text-sm text-muted-foreground">
                      No recent backups
                    </p>
                  </div>
                ) : (
                  <div className="space-y-3">
                    {backups.slice(0, 5).map((backup) => {
                      const connection = connections.find(
                        (c) => c.id === backup.connection_id,
                      );
                      return (
                        <div
                          key={backup.id}
                          className="flex items-center justify-between p-3 border rounded-lg"
                        >
                          <div className="flex items-center gap-3">
                            <HardDrive className="h-4 w-4 text-muted-foreground" />
                            <div>
                              <div className="flex items-center gap-2">
                                <p className="font-medium">
                                  {connection?.name || "Unknown"}
                                </p>
                                <span
                                  className={`text-xs px-2 py-1 rounded ${
                                    backup.backup_type === "full"
                                      ? "bg-blue-100 text-blue-800"
                                      : backup.backup_type === "incremental"
                                        ? "bg-green-100 text-green-800"
                                        : "bg-orange-100 text-orange-800"
                                  }`}
                                >
                                  {backup.backup_type}
                                </span>
                              </div>
                              <p className="text-xs text-muted-foreground">
                                {backup.backup_method} •{" "}
                                {(backup.file_size / 1024 / 1024).toFixed(1)} MB
                              </p>
                            </div>
                          </div>
                          <div className="flex items-center gap-2">
                            <span
                              className={`h-2 w-2 rounded-full ${
                                backup.status === "completed"
                                  ? "bg-green-600"
                                  : backup.status === "failed"
                                    ? "bg-red-600"
                                    : backup.status === "running"
                                      ? "bg-blue-600"
                                      : "bg-gray-400"
                              }`}
                            />
                            <span className="text-xs capitalize">
                              {backup.status}
                            </span>
                          </div>
                        </div>
                      );
                    })}
                  </div>
                )}
              </CardContent>
            </Card>
          </div>
        </TabsContent>

        <TabsContent value="connections" className="space-y-4">
          <DatabaseConnectionsTable
            connections={connections}
            isLoading={isLoading}
            onConnectionUpdated={handleDataUpdated}
          />
        </TabsContent>

        <TabsContent value="queries" className="space-y-4">
          <DatabaseQueryPanel
            connections={connections}
            queries={queries}
            isLoading={isLoading}
            onQueryUpdated={handleDataUpdated}
          />
        </TabsContent>

        <TabsContent value="backups" className="space-y-4">
          <DatabaseBackupsPanel
            connections={connections}
            backups={backups}
            isLoading={isLoading}
            onBackupUpdated={handleDataUpdated}
          />
        </TabsContent>
      </Tabs>
    </div>
  );
}
