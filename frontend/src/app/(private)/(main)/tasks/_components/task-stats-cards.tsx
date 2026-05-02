"use client";

import { useTaskStats } from "@/hooks/use-task-queries";
import { AlertCircle, CheckCircle2, Clock, Zap } from "lucide-react";
import { StatGrid } from "@/components/ui/stat-grid";
import { Skeleton } from "@/components/ui/skeleton";
import { Card, CardContent } from "@/components/ui/card";

interface TaskStatsCardsProps {
  userId: string;
  refreshTrigger: number;
}

export function TaskStatsCards({
  userId,
  refreshTrigger: _refreshTrigger,
}: TaskStatsCardsProps) {
  const { data: stats, isLoading } = useTaskStats(userId);

  if (isLoading || !stats) {
    return (
      <Card className="border-border/60 p-0">
        <CardContent className="grid grid-cols-2 md:grid-cols-4 divide-y md:divide-y-0 md:divide-x divide-border/60 p-0">
          {[...Array(4)].map((_, i) => (
            <div key={i} className="p-2.5 sm:p-3 space-y-1">
              <Skeleton className="h-3 w-16" />
              <Skeleton className="h-5 sm:h-6 w-8" />
              <Skeleton className="h-2.5 w-20" />
            </div>
          ))}
        </CardContent>
      </Card>
    );
  }

  return (
    <StatGrid
      items={[
        {
          label: "Pending",
          value: stats.pendingTasks,
          icon: <Clock className="h-3 w-3 sm:h-4 sm:w-4" />,
          accent: "amber",
          secondary: "Tasks awaiting action",
        },
        {
          label: "High Priority",
          value: stats.highPriorityTasks,
          icon: <Zap className="h-3 w-3 sm:h-4 sm:w-4" />,
          accent: "blue",
          secondary: "Urgent tasks",
        },
        {
          label: "Overdue",
          value: stats.overdueTasks,
          icon: <AlertCircle className="h-3 w-3 sm:h-4 sm:w-4" />,
          accent: "rose",
          secondary: "Past due date",
          emphasizeValue: true,
        },
        {
          label: "Completed",
          value: stats.completedTasks,
          icon: <CheckCircle2 className="h-3 w-3 sm:h-4 sm:w-4" />,
          accent: "emerald",
          secondary: "Finished tasks",
          emphasizeValue: true,
        },
      ]}
    />
  );
}
