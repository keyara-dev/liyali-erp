"use client";

import { useDashboardMetrics } from "@/hooks/use-dashboard-metrics";
import { PageHeader } from "@/components/base/page-header";
import { WorkflowStatusChart } from "./workflow-status-chart";
import { RecentActivity } from "./recent-activity";
import { GreetingCard } from "./greeting-card";
import { LoadingDashboard } from "../loading";
import CustomAlert from "@/components/ui/custom-alert";

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
        <LoadingDashboard />
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
        <CustomAlert
          type="error"
          message="Something went wrong while fetching dashboard metrics."
        />
        <div className="rounded-lg border bg-red-50 text-destructive p-6">
          <div className="flex items-center justify-center h-32">
            <p className="text-muted-foreground">
              Failed to load dashboard data. <br /> Please try again later.
            </p>
            {error && process.env.NODE_ENV != "production" && (
              <pre>{JSON.stringify(error, null, 2)}</pre>
            )}
          </div>
        </div>
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
        <div className="md:col-span-2 grid">
          <RecentActivity metrics={metrics} />
        </div>
      </div>
    </div>
  );
}
