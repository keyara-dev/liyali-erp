"use client";

import { useState, useMemo } from "react";
import { format, subDays, startOfDay, endOfDay } from "date-fns";
import { useQueryClient } from "@tanstack/react-query";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { ApprovalReports } from "./approval-reports";
import { UserActivityReports } from "./user-activity-reports";
import { SystemStatistics } from "./system-statistics";
import { AnalyticsDashboard } from "@/components/workflows/analytics-dashboard";
import { ReportsHeader } from "./reports-header";
import { QUERY_KEYS } from "@/lib/constants";
import { notify } from "@/lib/utils";
import { useDateRangeUrlState } from "@/hooks/use-date-range-url-state";
import type { DateRange } from "@/types/reports";
import {
  useSystemStats,
  useApprovalMetrics,
  useUserActivity,
  useAnalyticsDashboard,
} from "@/hooks/use-reports-queries";
import {
  exportSystemStatsToCSV,
  exportApprovalMetricsToCSV,
  exportUserActivityToCSV,
  exportAnalyticsDashboardToCSV,
} from "@/lib/export-utils";

interface AdminReportsClientProps {
  userId: string;
  userRole: string;
}

function defaultRange() {
  const today = new Date();
  return {
    from: format(startOfDay(subDays(today, 27)), "yyyy-MM-dd"),
    to: format(endOfDay(today), "yyyy-MM-dd"),
  };
}

export function AdminReportsClient({
  userId: _userId,
  userRole: _userRole,
}: AdminReportsClientProps) {
  const [activeTab, setActiveTab] = useState("overview");
  const [isRefreshing, setIsRefreshing] = useState(false);
  const queryClient = useQueryClient();

  const initial = useMemo(defaultRange, []);
  const { from, to, setRange } = useDateRangeUrlState({
    defaultFrom: initial.from,
    defaultTo: initial.to,
  });
  // Map URL string params to the DateRange shape expected by report hooks/components
  const dateRange = useMemo<DateRange>(
    () => ({ startDate: from, endDate: to }),
    [from, to]
  );

  const { data: systemStats } = useSystemStats(dateRange);
  const { data: approvalMetrics } = useApprovalMetrics(dateRange);
  const { data: userActivity } = useUserActivity(dateRange);
  const { data: analytics } = useAnalyticsDashboard(dateRange);

  const handleRefresh = async () => {
    setIsRefreshing(true);
    try {
      await Promise.all([
        queryClient.invalidateQueries({
          queryKey: [QUERY_KEYS.REPORTS.SYSTEM_STATS],
        }),
        queryClient.invalidateQueries({
          queryKey: [QUERY_KEYS.REPORTS.APPROVAL_METRICS],
        }),
        queryClient.invalidateQueries({
          queryKey: [QUERY_KEYS.REPORTS.USER_ACTIVITY],
        }),
        queryClient.invalidateQueries({
          queryKey: [QUERY_KEYS.REPORTS.ANALYTICS],
        }),
      ]);
      notify({
        title: "Success",
        description: "Reports refreshed successfully",
        type: "success",
      });
    } catch {
      notify({
        title: "Error",
        description: "Failed to refresh reports. Please try again.",
        type: "error",
      });
    } finally {
      setIsRefreshing(false);
    }
  };

  const handleExport = (formatChoice: "csv") => {
    if (formatChoice !== "csv") return;
    try {
      switch (activeTab) {
        case "overview":
          if (!systemStats)
            return notify({ title: "Error", description: "No data to export", type: "error" });
          exportSystemStatsToCSV(systemStats);
          break;
        case "analytics":
          if (!analytics)
            return notify({ title: "Error", description: "No data to export", type: "error" });
          exportAnalyticsDashboardToCSV(analytics);
          break;
        case "approvals":
          if (!approvalMetrics)
            return notify({ title: "Error", description: "No data to export", type: "error" });
          exportApprovalMetricsToCSV(approvalMetrics);
          break;
        case "activity":
          if (!userActivity)
            return notify({ title: "Error", description: "No data to export", type: "error" });
          exportUserActivityToCSV(userActivity);
          break;
        default:
          notify({ title: "Error", description: "Unknown tab selected", type: "error" });
          return;
      }
      notify({
        title: "Exported",
        description: "Current view downloaded as CSV",
        type: "success",
      });
    } catch {
      notify({
        title: "Error",
        description: "An error occurred during export",
        type: "error",
      });
    }
  };

  return (
    <div className="space-y-5">
      <ReportsHeader
        title="Admin Reports"
        subtitle="Workflow approvals, user activity, system metrics"
        from={from}
        to={to}
        onRangeChange={setRange}
        onRefresh={handleRefresh}
        onExport={handleExport}
        isRefreshing={isRefreshing}
      />

      <Tabs value={activeTab} onValueChange={setActiveTab} className="space-y-5">
        <TabsList className="inline-flex h-9 w-full sm:w-auto bg-muted/60 p-1 rounded-lg">
          {[
            { v: "overview", label: "Overview" },
            { v: "analytics", label: "Analytics" },
            { v: "approvals", label: "Approvals" },
            { v: "activity", label: "Activity" },
          ].map((t) => (
            <TabsTrigger
              key={t.v}
              value={t.v}
              className="flex-1 sm:flex-initial sm:px-6 data-[state=active]:bg-background data-[state=active]:text-foreground data-[state=active]:shadow-sm rounded-md"
            >
              {t.label}
            </TabsTrigger>
          ))}
        </TabsList>

        <TabsContent value="overview" className="mt-0">
          <SystemStatistics dateRange={dateRange} />
        </TabsContent>

        <TabsContent value="analytics" className="mt-0">
          <AnalyticsDashboard dateRange={dateRange} />
        </TabsContent>

        <TabsContent value="approvals" className="mt-0">
          <ApprovalReports dateRange={dateRange} />
        </TabsContent>

        <TabsContent value="activity" className="mt-0">
          <UserActivityReports dateRange={dateRange} />
        </TabsContent>
      </Tabs>
    </div>
  );
}
