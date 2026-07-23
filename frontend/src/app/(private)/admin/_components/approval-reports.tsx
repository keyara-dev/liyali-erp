"use client";

import { useState } from "react";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { useApprovalMetrics } from "@/hooks/use-reports-queries";
import { Input } from "@/components/ui/input";
import { Search, AlertCircle, CheckCircle2, XCircle, Clock } from "lucide-react";
import { Skeleton } from "@/components/ui/skeleton";
import { MetricCard } from "@/components/ui/metric-card";
import { FilterBar } from "@/components/ui/filter-bar";
import { DataList, DataListColumn } from "@/components/ui/data-list";
import { DocumentTypeChip } from "@/components/ui/document-type-chip";
import { StatusBadge } from "@/components/status-badge";
import EmptyState from "@/components/base/empty-state";
import type { DateRange, ApprovalActivity } from "@/types/reports";

interface ApprovalReportsProps {
  dateRange?: DateRange;
}

export function ApprovalReports({ dateRange }: ApprovalReportsProps) {
  const [searchTerm, setSearchTerm] = useState("");
  const { data: metrics, isLoading, error } = useApprovalMetrics(dateRange);

  if (isLoading) {
    return (
      <div className="space-y-6">
        <div className="grid gap-3 grid-cols-2 lg:grid-cols-3">
          {Array.from({ length: 3 }).map((_, i) => (
            <Skeleton key={i} className="h-28 rounded-md" />
          ))}
        </div>
        <Skeleton className="h-72 rounded-md" />
      </div>
    );
  }

  if (error) {
    return (
      <Alert variant="destructive">
        <AlertCircle className="h-4 w-4" />
        <AlertDescription>
          Failed to load approval reports. Please try again.
        </AlertDescription>
      </Alert>
    );
  }

  if (!metrics) {
    return (
      <EmptyState
        title="No approval data"
        description="No approvals found for the selected date range."
      />
    );
  }

  const filtered: ApprovalActivity[] = (metrics.recentApprovals ?? []).filter(
    (item) =>
      (item.documentNumber || "")
        .toLowerCase()
        .includes(searchTerm.toLowerCase()) ||
      (item.approverName || "").toLowerCase().includes(searchTerm.toLowerCase())
  );

  const columns: DataListColumn<ApprovalActivity>[] = [
    {
      id: "doc",
      header: "Document",
      cell: (a) => (
        <span className="font-medium text-primary">{a.documentNumber}</span>
      ),
    },
    {
      id: "type",
      header: "Type",
      priority: "md",
      cell: (a) => <DocumentTypeChip type={a.documentType} />,
    },
    {
      id: "status",
      header: "Status",
      cell: (a) => <StatusBadge status={a.action} type="action" />,
    },
    {
      id: "approver",
      header: "Approver",
      priority: "md",
      cell: (a) => (
        <span className="text-sm text-muted-foreground">{a.approverName}</span>
      ),
    },
    {
      id: "time",
      header: "Time",
      priority: "lg",
      cell: (a) => (
        <span className="text-sm text-muted-foreground">
          {new Date(a.createdAt).toLocaleDateString("en", {
            month: "short",
            day: "numeric",
            hour: "2-digit",
            minute: "2-digit",
          })}
        </span>
      ),
    },
  ];

  const hasActiveFilters = searchTerm.length > 0;

  return (
    <div className="space-y-6">
      <div className="grid gap-3 grid-cols-2 lg:grid-cols-3">
        <MetricCard
          title="Approved"
          value={metrics.totalApproved ?? 0}
          icon={<CheckCircle2 className="h-4 w-4" />}
          accent="emerald"
          secondary={`${(metrics.approvalRate ?? 0).toFixed(1)}% approval rate`}
        />
        <MetricCard
          title="Rejections"
          value={metrics.totalRejected ?? 0}
          icon={<XCircle className="h-4 w-4" />}
          accent="rose"
          secondary={`${(100 - (metrics.approvalRate ?? 0)).toFixed(1)}% rejection rate`}
        />
        <MetricCard
          title="Pending Review"
          value={metrics.totalPending ?? 0}
          icon={<Clock className="h-4 w-4" />}
          accent="amber"
          secondary="Awaiting next approver"
        />
      </div>

      <FilterBar
        search={
          <div className="relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
            <Input
              placeholder="Search by document or approver…"
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="pl-10"
            />
          </div>
        }
        hasActiveFilters={hasActiveFilters}
        onReset={() => setSearchTerm("")}
        meta={`${filtered.length} approval${filtered.length === 1 ? "" : "s"}${hasActiveFilters ? " (filtered)" : ""}`}
      />

      <DataList<ApprovalActivity>
        rows={filtered}
        columns={columns}
        getRowId={(a) => a.id}
        emptyMessage="No approvals found."
        mobileCard={(a) => (
          <div className="flex flex-col gap-2">
            <div className="flex items-start justify-between gap-2">
              <div className="min-w-0">
                <div className="font-medium text-primary">{a.documentNumber}</div>
                <div className="text-xs text-muted-foreground">
                  {a.approverName}
                </div>
              </div>
              <StatusBadge status={a.action} type="action" />
            </div>
            <div className="flex flex-wrap items-center gap-2 text-xs text-muted-foreground">
              <DocumentTypeChip type={a.documentType} />
              <span>
                {new Date(a.createdAt).toLocaleDateString("en", {
                  month: "short",
                  day: "numeric",
                })}
              </span>
            </div>
          </div>
        )}
      />
    </div>
  );
}

export type { ApprovalReportsProps };
