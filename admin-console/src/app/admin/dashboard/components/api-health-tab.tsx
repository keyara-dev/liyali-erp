"use client";

import { Activity, Zap, AlertOctagon } from "lucide-react";
import { Skeleton } from "@/components/ui/skeleton";
import {
  useAPIStats,
  useAPIErrors,
  useAPIAlerts,
  useRealTimeMetrics,
} from "@/hooks/use-api-monitoring";
import type { APIError, APIAlert } from "@/app/_actions/api-monitoring";
import { APIStatsGrid } from "@/app/admin/api-monitoring/components/api-stats-grid";
import { APIPerformanceChart } from "@/app/admin/api-monitoring/components/api-performance-chart";
import { APIErrorsPanel } from "@/app/admin/api-monitoring/components/api-errors-panel";
import { APIAlertsPanel } from "@/app/admin/api-monitoring/components/api-alerts-panel";

function RealTimeStrip() {
  const { data, isLoading } = useRealTimeMetrics();

  if (isLoading) {
    return (
      <div className="flex gap-4 p-3 rounded-lg border bg-muted/30">
        {[...Array(3)].map((_, i) => (
          <Skeleton key={i} className="h-8 w-32" />
        ))}
      </div>
    );
  }

  if (!data) return null;

  return (
    <div className="flex flex-wrap gap-4 p-3 rounded-lg border bg-muted/30 items-center">
      <div className="flex items-center gap-1.5">
        <Zap className="h-3.5 w-3.5 text-blue-500" />
        <span className="text-xs text-muted-foreground">Req/sec:</span>
        <span className="text-sm font-semibold">{data.current_rps.toFixed(1)}</span>
      </div>
      <div className="flex items-center gap-1.5">
        <Activity className="h-3.5 w-3.5 text-green-500" />
        <span className="text-xs text-muted-foreground">Avg response:</span>
        <span className="text-sm font-semibold">{data.avg_response_time.toFixed(0)}ms</span>
      </div>
      <div className="flex items-center gap-1.5">
        <AlertOctagon className="h-3.5 w-3.5 text-red-500" />
        <span className="text-xs text-muted-foreground">Error rate:</span>
        <span className="text-sm font-semibold">{data.error_rate.toFixed(2)}%</span>
      </div>
      <div className="ml-auto flex items-center gap-1.5">
        <span className="inline-block h-2 w-2 rounded-full bg-green-500 animate-pulse" />
        <span className="text-xs text-muted-foreground">Live · refreshes every 10s</span>
      </div>
    </div>
  );
}

export function APIHealthTab() {
  const { data: apiStats, isLoading: statsLoading } = useAPIStats();
  const { data: errorsData, isLoading: errorsLoading, refetch: refetchErrors } =
    useAPIErrors();
  const { data: alertsData, isLoading: alertsLoading, refetch: refetchAlerts } =
    useAPIAlerts();

  const errors: APIError[] = Array.isArray(errorsData) ? errorsData : [];
  const alerts: APIAlert[] = Array.isArray(alertsData) ? alertsData : [];

  return (
    <div className="space-y-6">
      {/* Real-time strip */}
      <RealTimeStrip />

      {/* API Stats Grid */}
      <APIStatsGrid stats={apiStats ?? null} isLoading={statsLoading} />

      {/* Performance Chart */}
      <APIPerformanceChart />

      {/* Errors + Alerts side by side */}
      <div className="space-y-6">
        <APIErrorsPanel
          errors={errors}
          isLoading={errorsLoading}
          onErrorUpdated={() => refetchErrors()}
        />
        <APIAlertsPanel
          alerts={alerts}
          isLoading={alertsLoading}
          onAlertUpdated={() => refetchAlerts()}
        />
      </div>
    </div>
  );
}
