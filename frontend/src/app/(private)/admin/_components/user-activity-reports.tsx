"use client";

import { Alert, AlertDescription } from "@/components/ui/alert";
import { useUserActivity } from "@/hooks/use-reports-queries";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { User, Users, CheckCircle2, AlertCircle } from "lucide-react";
import { MetricCard } from "@/components/ui/metric-card";
import { DataList, DataListColumn } from "@/components/ui/data-list";
import EmptyState from "@/components/base/empty-state";
import type { DateRange, UserActivity } from "@/types/reports";

interface UserActivityReportsProps {
  dateRange?: DateRange;
}

function formatDate(iso?: string) {
  if (!iso) return "N/A";
  try {
    return new Date(iso).toLocaleDateString("en", {
      month: "short",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
  } catch {
    return "N/A";
  }
}

export function UserActivityReports({ dateRange }: UserActivityReportsProps) {
  const { data: activity, isLoading, error } = useUserActivity(dateRange);

  if (isLoading) {
    return (
      <div className="space-y-6">
        <div className="grid gap-3 grid-cols-2 lg:grid-cols-3">
          {Array.from({ length: 3 }).map((_, i) => (
            <Skeleton key={i} className="h-28 rounded-md" />
          ))}
        </div>
        <Skeleton className="h-44 rounded-md" />
        <Skeleton className="h-72 rounded-md" />
      </div>
    );
  }

  if (error) {
    return (
      <Alert variant="destructive">
        <AlertCircle className="h-4 w-4" />
        <AlertDescription>
          Failed to load user activity. Please try again.
        </AlertDescription>
      </Alert>
    );
  }

  if (!activity) {
    return (
      <EmptyState
        title="No user activity"
        description="No user activity found for the selected date range."
      />
    );
  }

  const topContributors = (activity.users ?? []).slice(0, 5);
  const allUsers = activity.users ?? [];

  const columns: DataListColumn<UserActivity>[] = [
    {
      id: "name",
      header: "User",
      cell: (u) => <span className="font-medium">{u.name}</span>,
    },
    {
      id: "role",
      header: "Role",
      priority: "md",
      cell: (u) => (
        <Badge variant="outline">{u.role.replace(/_/g, " ")}</Badge>
      ),
    },
    {
      id: "approvals",
      header: "Approvals",
      align: "right",
      priority: "md",
      cell: (u) => <span className="tabular-nums">{u.approvalCount}</span>,
    },
    {
      id: "rejections",
      header: "Rejections",
      align: "right",
      priority: "lg",
      cell: (u) => <span className="tabular-nums">{u.rejectionCount}</span>,
    },
    {
      id: "active",
      header: "Active",
      align: "right",
      priority: "lg",
      cell: (u) => <span className="tabular-nums">{u.activeDocuments}</span>,
    },
    {
      id: "last",
      header: "Last activity",
      priority: "lg",
      cell: (u) => (
        <span className="text-sm text-muted-foreground">
          {formatDate(u.lastActivity)}
        </span>
      ),
    },
  ];

  return (
    <div className="space-y-6">
      <div className="grid gap-3 grid-cols-2 lg:grid-cols-3">
        <MetricCard
          title="Active Users"
          value={activity.activeUsers ?? 0}
          icon={<Users className="h-4 w-4" />}
          accent="blue"
          secondary={`${allUsers.length} total users`}
        />
        <MetricCard
          title="Docs in Progress"
          value={activity.documentsInProgress ?? 0}
          icon={<User className="h-4 w-4" />}
          accent="violet"
          secondary="Across all users"
        />
        <MetricCard
          title="Total Actions"
          value={activity.totalActions ?? 0}
          icon={<CheckCircle2 className="h-4 w-4" />}
          accent="emerald"
          secondary="Approvals and rejections"
        />
      </div>

      <Card className="border-border/60">
        <CardHeader>
          <CardTitle className="text-base">Top Contributors</CardTitle>
        </CardHeader>
        <CardContent>
          {topContributors.length === 0 ? (
            <EmptyState
              title="No contributors yet"
              description="Approvals will appear here once users start acting on tasks."
            />
          ) : (
            <ul className="space-y-2">
              {topContributors.map((u, idx) => (
                <li
                  key={u.id}
                  className="flex items-center justify-between p-3 rounded-md border border-border/60"
                >
                  <div className="flex items-center gap-3 min-w-0">
                    <div className="h-9 w-9 rounded-full bg-primary/10 flex items-center justify-center font-semibold text-primary text-sm shrink-0">
                      {(u.name || "?").charAt(0).toUpperCase()}
                    </div>
                    <div className="min-w-0">
                      <p className="font-medium leading-tight truncate">
                        {u.name}{" "}
                        {idx === 0 && (
                          <span className="ml-1 text-[10px] uppercase text-amber-600 font-bold">
                            Top
                          </span>
                        )}
                      </p>
                      <p className="text-xs text-muted-foreground capitalize">
                        {u.role.replace(/_/g, " ")}
                      </p>
                    </div>
                  </div>
                  <div className="text-right shrink-0">
                    <Badge variant="secondary">
                      {u.approvalCount} approvals
                    </Badge>
                    <p className="text-xs text-muted-foreground mt-1">
                      {u.activeDocuments} active
                    </p>
                  </div>
                </li>
              ))}
            </ul>
          )}
        </CardContent>
      </Card>

      <Card className="border-border/60">
        <CardHeader>
          <CardTitle className="text-base">User Activity Log</CardTitle>
        </CardHeader>
        <CardContent>
          <DataList<UserActivity>
            rows={allUsers}
            columns={columns}
            getRowId={(u) => u.id}
            emptyMessage="No user activity found."
            mobileCard={(u) => (
              <div className="flex flex-col gap-2">
                <div className="flex items-start justify-between gap-2">
                  <div className="min-w-0">
                    <div className="font-medium">{u.name}</div>
                    <div className="text-xs text-muted-foreground capitalize">
                      {u.role.replace(/_/g, " ")}
                    </div>
                  </div>
                  <Badge variant="secondary">{u.approvalCount} ✓</Badge>
                </div>
                <div className="flex flex-wrap items-center gap-2 text-xs text-muted-foreground">
                  <span>{u.rejectionCount} rejected</span>
                  <span>·</span>
                  <span>{u.activeDocuments} active</span>
                  <span>·</span>
                  <span>{formatDate(u.lastActivity)}</span>
                </div>
              </div>
            )}
          />
        </CardContent>
      </Card>
    </div>
  );
}

export type { UserActivityReportsProps };
