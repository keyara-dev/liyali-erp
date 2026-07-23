// frontend/src/app/(private)/admin/_components/system-statistics.tsx
"use client";

import { Alert, AlertDescription } from "@/components/ui/alert";
import { useSystemStats } from "@/hooks/use-reports-queries";
import {
  FileText,
  Clock,
  CheckCircle2,
  AlertCircle,
  TrendingUp,
} from "lucide-react";
import { MetricCard } from "@/components/ui/metric-card";
import { ReportChart } from "@/components/ui/report-chart";
import { StatusBadge } from "@/components/status-badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import EmptyState from "@/components/base/empty-state";
import type { DateRange } from "@/types/reports";

interface SystemStatisticsProps {
  dateRange?: DateRange;
}

interface DocTypeRow extends Record<string, unknown> {
  name: string;
  count: number;
}

export function SystemStatistics({ dateRange }: SystemStatisticsProps) {
  const { data: stats, isLoading, error } = useSystemStats(dateRange);

  if (isLoading) {
    return (
      <div className="space-y-6">
        <div className="grid gap-3 grid-cols-2 lg:grid-cols-4">
          {Array.from({ length: 4 }).map((_, i) => (
            <Skeleton key={i} className="h-28 rounded-md" />
          ))}
        </div>
        <Skeleton className="h-72 rounded-md" />
        <Skeleton className="h-56 rounded-md" />
      </div>
    );
  }

  if (error) {
    return (
      <Alert variant="destructive">
        <AlertCircle className="h-4 w-4" />
        <AlertDescription>
          Failed to load system statistics. Please try again.
        </AlertDescription>
      </Alert>
    );
  }

  if (!stats) {
    return (
      <EmptyState
        title="No statistics available"
        description="No data found for the selected date range."
      />
    );
  }

  const docTypeRows: DocTypeRow[] = [
    { name: "Requisitions", count: stats.documentTypeBreakdown?.requisitions ?? 0 },
    { name: "Purchase Orders", count: stats.documentTypeBreakdown?.purchaseOrders ?? 0 },
    { name: "Payment Vouchers", count: stats.documentTypeBreakdown?.paymentVouchers ?? 0 },
    { name: "GRN", count: stats.documentTypeBreakdown?.grn ?? 0 },
    { name: "Budgets", count: stats.documentTypeBreakdown?.budgets ?? 0 },
  ];

  const statusRows: { label: string; value: number; status: string }[] = [
    { label: "Draft", value: stats.statusBreakdown?.draft ?? 0, status: "draft" },
    { label: "Submitted", value: stats.statusBreakdown?.submitted ?? 0, status: "submitted" },
    { label: "In Review", value: stats.statusBreakdown?.inReview ?? 0, status: "in_approval" },
    { label: "Approved", value: stats.statusBreakdown?.approved ?? 0, status: "approved" },
    { label: "Rejected", value: stats.statusBreakdown?.rejected ?? 0, status: "rejected" },
  ];

  return (
    <div className="space-y-6">
      <div className="grid gap-3 grid-cols-2 lg:grid-cols-4">
        <MetricCard
          title="Total Documents"
          value={stats.totalDocuments ?? 0}
          icon={<FileText className="h-4 w-4" />}
          accent="blue"
          secondary="All time"
        />
        <MetricCard
          title="Approval Rate"
          value={`${(stats.approvalRate ?? 0).toFixed(1)}%`}
          icon={<TrendingUp className="h-4 w-4" />}
          accent="emerald"
          secondary={`${stats.approvedDocuments ?? 0} approved`}
        />
        <MetricCard
          title="Avg Approval Time"
          value={(stats.averageApprovalTime ?? 0).toFixed(1)}
          icon={<Clock className="h-4 w-4" />}
          accent="amber"
          secondary="days"
        />
        <MetricCard
          title="Rejection Rate"
          value={`${(stats.rejectionRate ?? 0).toFixed(1)}%`}
          icon={<AlertCircle className="h-4 w-4" />}
          accent="rose"
          secondary={`${stats.rejectedDocuments ?? 0} rejected`}
        />
      </div>

      <Card className="border-border/60">
        <CardHeader>
          <CardTitle className="text-base">Document Type Distribution</CardTitle>
        </CardHeader>
        <CardContent>
          <ReportChart<DocTypeRow>
            kind="bar"
            data={docTypeRows}
            xKey="name"
            series={[{ dataKey: "count", label: "Count" }]}
            perBarColor
          />
        </CardContent>
      </Card>

      <Card className="border-border/60">
        <CardHeader>
          <CardTitle className="text-base">Status Summary</CardTitle>
        </CardHeader>
        <CardContent>
          <ul className="divide-y divide-border/60">
            {statusRows.map((row) => (
              <li
                key={row.label}
                className="flex items-center justify-between py-3"
              >
                <div className="flex items-center gap-3">
                  <StatusBadge status={row.status} type="document" />
                </div>
                <span className="text-sm font-semibold tabular-nums">
                  {row.value}
                </span>
              </li>
            ))}
          </ul>
        </CardContent>
      </Card>
    </div>
  );
}

export type { SystemStatisticsProps };
