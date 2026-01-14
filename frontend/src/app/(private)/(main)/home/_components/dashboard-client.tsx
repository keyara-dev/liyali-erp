"use client";

import { useDashboardMetrics } from "@/hooks/use-dashboard-metrics";
import { PageHeader } from "@/components/base/page-header";
import { WorkflowStatusChart } from "./workflow-status-chart";
import { RecentActivity } from "./recent-activity";
import { GreetingCard } from "./greeting-card";

interface DashboardClientProps {
  userId: string;
  userName: string;
  userRole: string;
}

export function DashboardClient({
  userId,
  userName,
  userRole,
}: DashboardClientProps) {
  // Use React Query hook for dashboard metrics
  const { data: metrics, isLoading, error } = useDashboardMetrics();

  if (isLoading) {
    return (
      <div className="space-y-6">
        <PageHeader
          title="Dashboard"
          subtitle="Loading workflow metrics..."
          showBackButton={false}
        />
      </div>
    );
  }

  if (error || !metrics) {
    return (
      <div className="space-y-6">
        <PageHeader
          title="Dashboard"
          subtitle={error?.message || "Failed to load dashboard"}
          showBackButton={false}
        />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title="Dashboard"
        subtitle="View your workflow metrics and recent activity"
        showBackButton={false}
      />

      {/* Greeting Card with Quick Actions and Analytics */}
      <GreetingCard userName={userName} userRole={userRole} metrics={metrics} />

      {/* Charts and Actions Grid */}
      <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
        <div className="md:col-span-1">
          <WorkflowStatusChart metrics={metrics} />
        </div>
      </div>

      {/* Recent Activity */}
      <RecentActivity metrics={metrics} />
    </div>
  );
}
