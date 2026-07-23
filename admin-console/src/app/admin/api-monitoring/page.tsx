"use client";

import { useState, useEffect } from "react";
import { Button } from "@/components/ui/button";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Zap, RefreshCw, Activity, AlertTriangle, Settings } from "lucide-react";
import { toast } from "sonner";
import {
  exportAPIData,
  type APIFilters,
} from "@/app/_actions/api-monitoring";
import {
  useAPIEndpoints,
  useAPIMetrics,
  useAPIErrors,
  useAPIAlerts,
  useAPIStats,
  useAPICategories,
} from "@/hooks/use-api-monitoring";
import { APIMonitoringFiltersComponent } from "./components/api-monitoring-filters";
import { APIStatsGrid } from "./components/api-stats-grid";
import { APIEndpointsTable } from "./components/api-endpoints-table";
import { APIErrorsPanel } from "./components/api-errors-panel";
import { APIAlertsPanel } from "./components/api-alerts-panel";
import { APIPerformanceChart } from "./components/api-performance-chart";

export default function APIMonitoringPage() {
  const [activeTab, setActiveTab] = useState("performance");
  const [filters, setFilters] = useState<APIFilters>({ time_range: "24h" });
  const [searchTerm, setSearchTerm] = useState("");

  // TanStack Query hooks
  const {
    data: endpoints = [],
    isLoading: isLoadingEndpoints,
    refetch: refetchEndpoints,
    isRefetching,
  } = useAPIEndpoints(filters);
  const { data: metrics = [], refetch: refetchMetrics } =
    useAPIMetrics(filters);
  const { data: errors = [], refetch: refetchErrors } =
    useAPIErrors(filters);
  const { data: alerts = [], refetch: refetchAlerts } =
    useAPIAlerts(filters);
  const { data: stats, refetch: refetchStats } = useAPIStats();
  const { data: categories = [] } = useAPICategories();

  const isLoading = isLoadingEndpoints;
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
    refetchEndpoints();
    refetchMetrics();
    refetchErrors();
    refetchAlerts();
    refetchStats();
  };

  const handleFiltersChange = (newFilters: APIFilters) => {
    setFilters(newFilters);
  };

  const handleResetFilters = () => {
    setFilters({ time_range: "24h" });
    setSearchTerm("");
  };

  const handleExport = async (
    type: "endpoints" | "metrics" | "errors" | "alerts",
    format: "csv" | "json" | "excel",
  ) => {
    try {
      const result = await exportAPIData(type, format, filters);
      if (result.success) {
        toast.success(
          `${type.charAt(0).toUpperCase() + type.slice(1)} export initiated. Download will be available shortly.`,
        );
        if (result.data?.download_url) {
          window.open(result.data.download_url, "_blank");
        }
      } else {
        toast.error(`Failed to export ${type}`);
      }
    } catch (error) {
      console.error(`Error exporting ${type}:`, error);
      toast.error(`Failed to export ${type}`);
    }
  };

  const handleDataUpdated = () => {
    handleRefresh();
  };

  return (
    <div className="flex-1 space-y-4 p-4 md:p-8 pt-6">
      <div className="flex items-center justify-between space-y-2">
        <div>
          <h2 className="text-3xl font-bold tracking-tight">API Monitoring</h2>
          <p className="text-muted-foreground">
            Monitor API performance, errors, and system health
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
      <APIMonitoringFiltersComponent
        filters={filters}
        onFiltersChange={handleFiltersChange}
        onReset={handleResetFilters}
        onExport={handleExport}
        searchTerm={searchTerm}
        onSearchChange={setSearchTerm}
        categories={categories}
      />

      {/* Stats Grid */}
      <APIStatsGrid stats={stats ?? null} isLoading={isLoading} />

      {/* Main Content Tabs */}
      <Tabs
        value={activeTab}
        onValueChange={setActiveTab}
        className="space-y-4"
      >
        <TabsList className="grid w-full grid-cols-4">
          <TabsTrigger value="endpoints" className="flex items-center gap-2">
            <Zap className="h-4 w-4" />
            Endpoints
          </TabsTrigger>
          <TabsTrigger value="performance" className="flex items-center gap-2">
            <Activity className="h-4 w-4" />
            Performance
          </TabsTrigger>
          <TabsTrigger value="errors" className="flex items-center gap-2">
            <AlertTriangle className="h-4 w-4" />
            Errors
          </TabsTrigger>
          <TabsTrigger value="alerts" className="flex items-center gap-2">
            <Settings className="h-4 w-4" />
            Alerts
          </TabsTrigger>
        </TabsList>

        <TabsContent value="endpoints" className="space-y-4">
          <APIEndpointsTable
            endpoints={endpoints}
            metrics={metrics}
            isLoading={isLoading}
            onEndpointUpdated={handleDataUpdated}
          />
        </TabsContent>

        <TabsContent value="performance" className="space-y-4">
          <APIPerformanceChart
            timeRange={filters.time_range}
            onTimeRangeChange={(range) =>
              setFilters((prev) => ({ ...prev, time_range: range }))
            }
          />
        </TabsContent>

        <TabsContent value="errors" className="space-y-4">
          <APIErrorsPanel
            errors={errors}
            isLoading={isLoading}
            onErrorUpdated={handleDataUpdated}
          />
        </TabsContent>

        <TabsContent value="alerts" className="space-y-4">
          <APIAlertsPanel
            alerts={alerts}
            isLoading={isLoading}
            onAlertUpdated={handleDataUpdated}
          />
        </TabsContent>
      </Tabs>
    </div>
  );
}
