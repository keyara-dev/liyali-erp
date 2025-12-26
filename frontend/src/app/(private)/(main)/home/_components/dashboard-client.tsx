"use client";

import { useState, useEffect } from "react";
import { getDashboardMetrics } from "@/app/_actions/dashboard";
import { usePendingApprovalCount } from "@/hooks/use-approval-workflow";
import { DashboardMetrics } from "@/types";
import { PageHeader } from "@/components/base/page-header";
import { WorkflowStatusChart } from "./workflow-status-chart";
import { ApprovalTimeChart } from "./approval-time-chart";
import { QuickActions } from "./quick-actions";
import { RecentActivity } from "./recent-activity";
import { GreetingCard } from "./greeting-card";

interface DashboardClientProps {
  userId: string;
  userRole: string;
}

export function DashboardClient({ userId, userRole }: DashboardClientProps) {
  const [metrics, setMetrics] = useState<DashboardMetrics | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Fetch pending approval count from new approval workflow
  const pendingCount = usePendingApprovalCount();

  useEffect(() => {
    async function fetchMetrics() {
      setIsLoading(true);
      setError(null);
      try {
        const result = await getDashboardMetrics();
        if (result.success && result.data) {
          // Enhance metrics with real pending approval count from backend
          const enhancedMetrics = {
            ...result.data,
            pendingApproval: pendingCount,
          };
          setMetrics(enhancedMetrics);
        } else {
          setError(result.message || "Failed to load dashboard metrics");
        }
      } catch (err) {
        console.error("Failed to fetch dashboard metrics:", err);
        setError("Failed to load dashboard metrics");
      } finally {
        setIsLoading(false);
      }
    }

    fetchMetrics();
  }, [pendingCount]);

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
          subtitle={error || "Failed to load dashboard"}
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
      <GreetingCard userName="User" userRole={userRole} metrics={metrics} />

      {/* Recent Activity */}
      <RecentActivity metrics={metrics} />

      {/* Charts and Actions Grid */}
      <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
        {/* Approval Time Chart */}
        <div className="md:col-span-1">
          <ApprovalTimeChart metrics={metrics} />
        </div>

        {/* Quick Actions */}
        <div className="md:col-span-1 lg:col-span-1">
          <QuickActions userRole={userRole} />
        </div>

        {/* Workflow Status Chart */}
        <div className="md:col-span-1">
          <WorkflowStatusChart metrics={metrics} />
        </div>
      </div>
    </div>
  );
}
