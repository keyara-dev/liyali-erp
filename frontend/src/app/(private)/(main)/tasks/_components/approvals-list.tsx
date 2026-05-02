"use client";

import { useState, useMemo } from "react";
import * as React from "react";
import { useApprovalTasks } from "@/hooks/use-approval-workflow";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { SelectField } from "@/components/ui/select-field";
import Search from "@/components/ui/search-field";
import { StatGrid } from "@/components/ui/stat-grid";
import { FilterBar } from "@/components/ui/filter-bar";
import {
  CheckCircle2,
  Clock,
  RefreshCw,
  Users,
  UserCheck,
  AlertTriangle,
  ListFilter,
  Inbox,
} from "lucide-react";
import { ApprovalTaskCard } from "@/components/workflows/approval-task-card";
import { ApprovalTask } from "@/types";
import { canUserActOnWorkflowTask } from "@/lib/workflow-utils";
import { cn } from "@/lib/utils";

interface ApprovalsListProps {
  userId: string;
  userRole: string;
}

type StatusFilter = "all" | "pending" | "claimed" | "completed";
type PriorityFilter = "all" | "HIGH" | "MEDIUM" | "LOW";
type SortBy = "date" | "priority" | "name";

const APPROVER_ROLES = ["admin", "approver", "finance"];

export function ApprovalsList({ userId, userRole }: ApprovalsListProps) {
  const [statusFilter, setStatusFilter] = useState<StatusFilter>("pending");
  const [priorityFilter, setPriorityFilter] = useState<PriorityFilter>("all");
  const [searchQuery, setSearchQuery] = useState("");
  const [sortBy, setSortBy] = useState<SortBy>("date");
  const [page] = useState(1);
  const limit = 50;

  const isBuiltInApprover = APPROVER_ROLES.some(
    (role) => role.toLowerCase() === userRole.toLowerCase(),
  );
  const currentUser = {
    id: userId,
    role: userRole,
    name: "Current User",
    isBuiltInApprover,
  };

  const filters = React.useMemo(
    () =>
      statusFilter === "all"
        ? {}
        : { status: statusFilter.toUpperCase() as any },
    [statusFilter],
  );

  const {
    data: approvalData,
    isLoading: isTasksLoading,
    error,
    refetch,
  } = useApprovalTasks(filters, page, limit);

  const tasks = approvalData?.data || [];

  const handleRefresh = () => refetch();

  const canUserAccessTask = (task: ApprovalTask) =>
    canUserActOnWorkflowTask(currentUser, task);

  const filteredTasks = useMemo(() => {
    return tasks
      .filter((task) => {
        if (priorityFilter !== "all" && task.priority !== priorityFilter) {
          return false;
        }
        if (
          searchQuery &&
          !`${task.entityType} ${task.entityId} ${task.stageName} ${task.documentNumber ?? ""} ${task.title ?? ""}`
            .toLowerCase()
            .includes(searchQuery.toLowerCase())
        ) {
          return false;
        }
        return true;
      })
      .sort((a, b) => {
        switch (sortBy) {
          case "priority": {
            const order = { URGENT: 0, HIGH: 1, MEDIUM: 2, LOW: 3 };
            return (
              (order[(a.priority as keyof typeof order) || "MEDIUM"] ?? 2) -
              (order[(b.priority as keyof typeof order) || "MEDIUM"] ?? 2)
            );
          }
          case "name":
            return `${a.documentNumber ?? a.entityType}`.localeCompare(
              `${b.documentNumber ?? b.entityType}`,
            );
          case "date":
          default:
            return (
              new Date(b.createdAt || 0).getTime() -
              new Date(a.createdAt || 0).getTime()
            );
        }
      });
  }, [tasks, priorityFilter, searchQuery, sortBy]);

  const groupedTasks = useMemo(
    () => ({
      claimedByMe: filteredTasks.filter((t) => {
        const s = t.status?.toUpperCase();
        return s === "CLAIMED" && t.claimedBy === currentUser.id;
      }),
      available: filteredTasks.filter((t) => {
        const s = t.status?.toUpperCase();
        return s === "PENDING" && canUserAccessTask(t);
      }),
      claimedByOthers: filteredTasks.filter((t) => {
        const s = t.status?.toUpperCase();
        return s === "CLAIMED" && t.claimedBy !== currentUser.id;
      }),
      completed: filteredTasks.filter((t) => {
        const s = t.status?.toUpperCase();
        return s === "APPROVED" || s === "REJECTED" || s === "COMPLETED";
      }),
    }),
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [filteredTasks, currentUser.id, currentUser.role, isBuiltInApprover],
  );

  const stats = {
    total: filteredTasks.length,
    claimedByMe: groupedTasks.claimedByMe.length,
    available: groupedTasks.available.length,
    claimedByOthers: groupedTasks.claimedByOthers.length,
    completed: groupedTasks.completed.length,
  };

  const hasActiveFilters =
    searchQuery !== "" ||
    statusFilter !== "pending" ||
    priorityFilter !== "all" ||
    sortBy !== "date";

  const clearFilters = () => {
    setSearchQuery("");
    setStatusFilter("pending");
    setPriorityFilter("all");
    setSortBy("date");
  };

  if (error) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <div className="text-center space-y-3">
          <AlertTriangle className="h-10 w-10 text-red-500 mx-auto" />
          <h2 className="text-lg font-semibold">
            Failed to load approval tasks
          </h2>
          <p className="text-sm text-muted-foreground">
            Please try refreshing or contact support if the issue persists.
          </p>
          <Button onClick={handleRefresh} variant="outline" size="sm">
            <RefreshCw className="h-4 w-4 mr-2" />
            Try Again
          </Button>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {/* Header row: refresh action only — page title is provided by parent PageHeader */}
      <div className="flex items-center justify-end">
        <Button
          onClick={handleRefresh}
          variant="outline"
          size="sm"
          disabled={isTasksLoading}
          className="gap-2"
        >
          <RefreshCw className={cn("h-3.5 w-3.5", isTasksLoading && "animate-spin")} />
          Refresh
        </Button>
      </div>

      <StatGrid
        items={[
          { label: "Claimed by Me", value: stats.claimedByMe, icon: <Users className="h-3 w-3 sm:h-4 sm:w-4" />, accent: "blue" },
          { label: "Available", value: stats.available, icon: <Clock className="h-3 w-3 sm:h-4 sm:w-4" />, accent: "emerald" },
          { label: "Claimed by Others", value: stats.claimedByOthers, icon: <UserCheck className="h-3 w-3 sm:h-4 sm:w-4" />, accent: "amber" },
          { label: "Completed", value: stats.completed, icon: <CheckCircle2 className="h-3 w-3 sm:h-4 sm:w-4" />, accent: "slate" },
          { label: "Total (view)", value: stats.total, icon: <ListFilter className="h-3 w-3 sm:h-4 sm:w-4" />, accent: "violet" },
        ]}
      />

      <FilterBar
        search={
          <Search
            placeholder="Search by document number, type, or stage…"
            value={searchQuery}
            onChange={(v) => setSearchQuery(v)}
            isClearable
          />
        }
        filters={
          <>
            <SelectField
              placeholder="Status"
              classNames={{ wrapper: "md:w-44" }}
              value={statusFilter}
              onValueChange={(v) => setStatusFilter(v as StatusFilter)}
              options={[
                { value: "all", label: "All statuses" },
                { value: "pending", label: "Available" },
                { value: "claimed", label: "Claimed" },
                { value: "completed", label: "Completed" },
              ]}
            />
            <SelectField
              placeholder="Priority"
              classNames={{ wrapper: "md:w-40" }}
              value={priorityFilter}
              onValueChange={(v) => setPriorityFilter(v as PriorityFilter)}
              options={[
                { value: "all", label: "All priorities" },
                { value: "HIGH", label: "High" },
                { value: "MEDIUM", label: "Medium" },
                { value: "LOW", label: "Low" },
              ]}
            />
            <SelectField
              placeholder="Sort"
              classNames={{ wrapper: "md:w-40" }}
              value={sortBy}
              onValueChange={(v) => setSortBy(v as SortBy)}
              options={[
                { value: "date", label: "Newest" },
                { value: "priority", label: "Priority" },
                { value: "name", label: "Document" },
              ]}
            />
          </>
        }
        hasActiveFilters={hasActiveFilters}
        onReset={clearFilters}
        meta={`Showing ${stats.total} task${stats.total !== 1 ? "s" : ""}`}
      />

      {/* Task groups */}
      <div className="space-y-6">
        {groupedTasks.claimedByMe.length > 0 && (
          <TaskGroup title="Claimed by You" count={groupedTasks.claimedByMe.length} accent="blue">
            {groupedTasks.claimedByMe.map((task) => (
              <ApprovalTaskCard
                key={task.id}
                taskId={task.id}
                currentUserId={currentUser.id}
                currentUserRole={currentUser.role}
              />
            ))}
          </TaskGroup>
        )}

        {groupedTasks.available.length > 0 && (
          <TaskGroup title="Available Tasks" count={groupedTasks.available.length} accent="emerald">
            {groupedTasks.available.map((task) => (
              <ApprovalTaskCard
                key={task.id}
                taskId={task.id}
                currentUserId={currentUser.id}
                currentUserRole={currentUser.role}
              />
            ))}
          </TaskGroup>
        )}

        {groupedTasks.claimedByOthers.length > 0 && (
          <TaskGroup title="Claimed by Others" count={groupedTasks.claimedByOthers.length} accent="amber">
            {groupedTasks.claimedByOthers.map((task) => (
              <ApprovalTaskCard
                key={task.id}
                taskId={task.id}
                currentUserId={currentUser.id}
                currentUserRole={currentUser.role}
              />
            ))}
          </TaskGroup>
        )}

        {groupedTasks.completed.length > 0 && (
          <TaskGroup title="Completed" count={groupedTasks.completed.length} accent="slate">
            {groupedTasks.completed.map((task) => (
              <ApprovalTaskCard
                key={task.id}
                taskId={task.id}
                currentUserId={currentUser.id}
                currentUserRole={currentUser.role}
              />
            ))}
          </TaskGroup>
        )}

        {filteredTasks.length === 0 && !isTasksLoading && (
          <Card className="border-dashed border-border/60">
            <CardContent className="py-10 text-center">
              <Inbox className="h-8 w-8 text-muted-foreground/60 mx-auto mb-3" />
              <p className="font-medium text-sm mb-1">No approval tasks</p>
              <p className="text-xs text-muted-foreground mb-3">
                {hasActiveFilters
                  ? "No tasks match your current filters."
                  : "There are no approval tasks assigned to your role right now."}
              </p>
              {hasActiveFilters && (
                <Button variant="outline" size="sm" onClick={clearFilters}>
                  Reset filters
                </Button>
              )}
            </CardContent>
          </Card>
        )}

        {isTasksLoading && (
          <div className="grid gap-3">
            {[1, 2, 3].map((i) => (
              <Card key={i} className="p-4 border-border/60">
                <div className="animate-pulse space-y-3">
                  <div className="flex items-center gap-3">
                    <div className="h-9 w-9 rounded-md bg-muted" />
                    <div className="flex-1 space-y-1.5">
                      <div className="h-4 bg-muted rounded w-48" />
                      <div className="h-3 bg-muted rounded w-32" />
                    </div>
                  </div>
                  <div className="h-3 bg-muted rounded w-3/4" />
                </div>
              </Card>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}

// ── Sub-components ──────────────────────────────────────────────────────────

type Accent = "blue" | "emerald" | "amber" | "slate" | "violet";

const BADGE_CLASSES: Record<Accent, string> = {
  blue: "bg-blue-600 text-white",
  emerald: "bg-emerald-600 text-white",
  amber: "bg-amber-600 text-white",
  slate: "bg-slate-600 text-white",
  violet: "bg-violet-600 text-white",
};

function TaskGroup({
  title,
  count,
  accent,
  children,
}: {
  title: string;
  count: number;
  accent: Accent;
  children: React.ReactNode;
}) {
  return (
    <section className="space-y-2">
      <div className="flex items-center gap-2">
        <h3 className="text-sm font-semibold">{title}</h3>
        <Badge className={cn("text-xs", BADGE_CLASSES[accent])}>{count}</Badge>
      </div>
      <div className="grid gap-2.5">{children}</div>
    </section>
  );
}
