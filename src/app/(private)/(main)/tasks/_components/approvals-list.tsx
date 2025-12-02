"use client";

import { useState } from "react";
import {
  useGetApprovalTasks,
  useGetApprovalStats,
} from "@/hooks/use-workflows";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Alert, AlertDescription } from "@/components/ui/alert";
import {
  CheckCircle2,
  AlertCircle,
  Clock,
  Filter,
  Search,
  ArrowRight,
} from "lucide-react";

const PRIORITY_COLORS = {
  HIGH: "text-red-600 dark:text-red-400",
  MEDIUM: "text-yellow-600 dark:text-yellow-400",
  LOW: "text-green-600 dark:text-green-400",
};

const PRIORITY_BG = {
  HIGH: "bg-red-50 dark:bg-red-900/20 border-red-200 dark:border-red-800",
  MEDIUM:
    "bg-yellow-50 dark:bg-yellow-900/20 border-yellow-200 dark:border-yellow-800",
  LOW: "bg-green-50 dark:bg-green-900/20 border-green-200 dark:border-green-800",
};

interface ApprovalsListProps {
  userId?: string;
}

export function ApprovalsList({ userId }: ApprovalsListProps) {
  const [statusFilter, setStatusFilter] = useState<
    "all" | "pending" | "approved" | "rejected"
  >("pending");
  const [priorityFilter, setPriorityFilter] = useState<
    "all" | "HIGH" | "MEDIUM" | "LOW"
  >("all");
  const [searchQuery, setSearchQuery] = useState("");
  const [sortBy, setSortBy] = useState<"date" | "priority" | "name">("date");

  const { data: tasksData, isLoading: isTasksLoading } = useGetApprovalTasks({
    status: statusFilter === "all" ? undefined : statusFilter,
  });

  const { data: statsData, isLoading: isStatsLoading } = useGetApprovalStats();

  const tasks = tasksData || [];

  // Filter tasks
  const filteredTasks = tasks
    .filter((task) => {
      if (priorityFilter !== "all" && task.importance !== priorityFilter) {
        return false;
      }
      if (
        searchQuery &&
        !`${task.entityType} ${task.entityNumber}`
          .toLowerCase()
          .includes(searchQuery.toLowerCase())
      ) {
        return false;
      }
      return true;
    })
    // Sort tasks
    .sort((a, b) => {
      switch (sortBy) {
        case "priority":
          const priorityOrder = { HIGH: 0, MEDIUM: 1, LOW: 2 };
          return (
            (priorityOrder[a.importance as keyof typeof priorityOrder] || 2) -
            (priorityOrder[b.importance as keyof typeof priorityOrder] || 2)
          );
        case "name":
          return `${a.entityType}${a.entityNumber}`.localeCompare(
            `${b.entityType}${b.entityNumber}`
          );
        case "date":
        default:
          return (
            new Date(b.createdAt || 0).getTime() -
            new Date(a.createdAt || 0).getTime()
          );
      }
    });

  if (isStatsLoading) {
    return (
      <div className="space-y-6">
        <Skeleton className="h-12 w-48" />
        <div className="grid gap-4 md:grid-cols-4">
          {[1, 2, 3, 4].map((i) => (
            <Skeleton key={i} className="h-24 w-full" />
          ))}
        </div>
      </div>
    );
  }

  const stats = statsData;

  return (
    <div className="space-y-6">
      {/* Statistics Cards */}
      {stats && (
        <div className="grid gap-4 md:grid-cols-4">
          <Card>
            <CardHeader className="pb-3">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                Total Pending
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stats.totalPending}</div>
              <p className="text-xs text-muted-foreground mt-1">
                Awaiting your action
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="pb-3">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                High Priority
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className={`text-2xl font-bold ${PRIORITY_COLORS.HIGH}`}>
                {stats.highPriority}
              </div>
              <p className="text-xs text-muted-foreground mt-1">
                Require immediate attention
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="pb-3">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                This Month
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stats.thisMonth}</div>
              <p className="text-xs text-muted-foreground mt-1">
                Approved this month
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="pb-3">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                Overdue
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold text-red-600 dark:text-red-400">
                {stats.overdue}
              </div>
              <p className="text-xs text-muted-foreground mt-1">
                Past due date
              </p>
            </CardContent>
          </Card>
        </div>
      )}

      {/* Filters */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg flex items-center gap-2">
            <Filter className="h-4 w-4" />
            Filters
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid gap-4 md:grid-cols-4">
            {/* Status Filter */}
            <div className="space-y-2">
              <label className="text-sm font-medium">Status</label>
              <Select
                value={statusFilter}
                onValueChange={(value: any) => setStatusFilter(value)}
              >
                <SelectTrigger>
                  <SelectValue placeholder="All statuses" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All Statuses</SelectItem>
                  <SelectItem value="pending">Pending</SelectItem>
                  <SelectItem value="approved">Approved</SelectItem>
                  <SelectItem value="rejected">Rejected</SelectItem>
                </SelectContent>
              </Select>
            </div>

            {/* Priority Filter */}
            <div className="space-y-2">
              <label className="text-sm font-medium">Priority</label>
              <Select
                value={priorityFilter}
                onValueChange={(value: any) => setPriorityFilter(value)}
              >
                <SelectTrigger>
                  <SelectValue placeholder="All priorities" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All Priorities</SelectItem>
                  <SelectItem value="HIGH">High</SelectItem>
                  <SelectItem value="MEDIUM">Medium</SelectItem>
                  <SelectItem value="LOW">Low</SelectItem>
                </SelectContent>
              </Select>
            </div>

            {/* Sort By */}
            <div className="space-y-2">
              <label className="text-sm font-medium">Sort By</label>
              <Select
                value={sortBy}
                onValueChange={(value: any) => setSortBy(value)}
              >
                <SelectTrigger>
                  <SelectValue placeholder="Sort" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="date">Date (Newest)</SelectItem>
                  <SelectItem value="priority">Priority</SelectItem>
                  <SelectItem value="name">Entity Name</SelectItem>
                </SelectContent>
              </Select>
            </div>

            {/* Search */}
            <div className="space-y-2">
              <label className="text-sm font-medium">Search</label>
              <div className="relative">
                <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
                <Input
                  placeholder="Search by entity..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="pl-8"
                />
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Approvals List */}
      <div className="space-y-3">
        {isTasksLoading ? (
          <>
            {[1, 2, 3].map((i) => (
              <Skeleton key={i} className="h-24 w-full" />
            ))}
          </>
        ) : filteredTasks.length === 0 ? (
          <Card>
            <CardContent className="flex flex-col items-center justify-center py-12">
              <CheckCircle2 className="h-12 w-12 text-green-600 mb-4 opacity-50" />
              <h3 className="font-semibold mb-1">All Caught Up!</h3>
              <p className="text-sm text-muted-foreground text-center">
                {statusFilter === "pending"
                  ? "You have no pending approvals."
                  : "No approvals match your filters."}
              </p>
            </CardContent>
          </Card>
        ) : (
          filteredTasks.map((task) => (
            <div
              key={task.id}
              className={`border rounded-lg p-4 hover:shadow-sm transition-shadow cursor-pointer ${
                task.importance
                  ? PRIORITY_BG[task.importance as keyof typeof PRIORITY_BG]
                  : ""
              }`}
            >
              <div className="flex items-start justify-between gap-4">
                <div className="flex-1 min-w-0">
                  {/* Header */}
                  <div className="flex items-center gap-2 mb-2">
                    <h3 className="font-semibold">
                      {task.entityType} #{task.entityNumber}
                    </h3>
                    {task.importance && (
                      <Badge
                        variant="outline"
                        className={`${PRIORITY_COLORS[task.importance as keyof typeof PRIORITY_COLORS]} border-current`}
                      >
                        {task.importance}
                      </Badge>
                    )}
                    <Badge variant="secondary">{task.stageName}</Badge>
                  </div>

                  {/* Details */}
                  <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
                    <div>
                      <span className="text-muted-foreground">Created</span>
                      <p className="font-medium">
                        {new Date(
                          task.createdAt || new Date()
                        ).toLocaleDateString()}
                      </p>
                    </div>
                    <div>
                      <span className="text-muted-foreground">Assigned</span>
                      <p className="font-medium">
                        {task.approverName || "Unassigned"}
                      </p>
                    </div>
                    {task.dueDate && (
                      <div>
                        <span className="text-muted-foreground">Due Date</span>
                        <p className="font-medium">
                          {new Date(task.dueDate).toLocaleDateString()}
                        </p>
                      </div>
                    )}
                    <div>
                      <span className="text-muted-foreground">Status</span>
                      <p className="font-medium">
                        {task.status === "pending"
                          ? "Pending"
                          : task.status === "approved"
                            ? "Approved"
                            : "Rejected"}
                      </p>
                    </div>
                  </div>
                </div>

                {/* Action Button */}
                {task.status === "pending" && (
                  <Button
                    variant="default"
                    className="flex-shrink-0"
                    onClick={() => {
                      // Navigate to approval page
                      window.location.href = `/${task.entityType.toLowerCase()}s/${task.entityId}/approval`;
                    }}
                  >
                    Review
                    <ArrowRight className="ml-2 h-4 w-4" />
                  </Button>
                )}
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  );
}
