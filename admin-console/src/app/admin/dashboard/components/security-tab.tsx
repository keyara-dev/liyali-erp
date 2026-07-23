"use client";

import {
  useAuditLogStats,
  useAuditLogAnalytics,
  useSecurityEvents,
} from "@/hooks/use-audit-logs";
import { AuditLogStatsGrid } from "@/app/admin/audit-logs/components/audit-log-stats-grid";
import { AuditLogAnalyticsCharts } from "@/app/admin/audit-logs/components/audit-log-analytics-charts";
import { SecurityEventsPanel } from "@/app/admin/audit-logs/components/security-events-panel";

export function SecurityTab() {
  const { data: statsData, isLoading: statsLoading } = useAuditLogStats();
  const { data: analyticsData, isLoading: analyticsLoading } =
    useAuditLogAnalytics();
  const { data: securityData, isLoading: securityLoading } =
    useSecurityEvents();

  return (
    <div className="space-y-6">
      <AuditLogStatsGrid
        stats={statsData ?? null}
        isLoading={statsLoading}
      />
      <AuditLogAnalyticsCharts
        analytics={analyticsData ?? null}
        stats={statsData ?? null}
        isLoading={analyticsLoading || statsLoading}
      />
      <SecurityEventsPanel
        stats={statsData ?? null}
        isLoading={securityLoading || statsLoading}
      />
    </div>
  );
}
