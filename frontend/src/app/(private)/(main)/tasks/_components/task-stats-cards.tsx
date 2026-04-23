"use client";

import { Card, CardContent } from "@/components/ui/card";
import { useTaskStats } from "@/hooks/use-task-queries";
import { AlertCircle, CheckCircle2, Clock, Zap } from "lucide-react";
import { cn } from "@/lib/utils";

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
              <div className="h-3 bg-muted rounded w-16 animate-pulse" />
              <div className="h-5 sm:h-6 bg-muted rounded w-8 animate-pulse" />
              <div className="h-2.5 bg-muted rounded w-20 animate-pulse" />
            </div>
          ))}
        </CardContent>
      </Card>
    );
  }

  return (
    <Card className="border-border/60 p-0">
      <CardContent className="grid grid-cols-2 md:grid-cols-4 divide-y md:divide-y-0 md:divide-x divide-border/60 p-0">
        <StatCell
          icon={<Clock className="h-3 w-3 sm:h-4 sm:w-4" />}
          label="Pending"
          value={stats.pendingTasks}
          secondary="Tasks awaiting action"
          accent="amber"
        />
        <StatCell
          icon={<Zap className="h-3 w-3 sm:h-4 sm:w-4" />}
          label="High Priority"
          value={stats.highPriorityTasks}
          secondary="Urgent tasks"
          accent="blue"
        />
        <StatCell
          icon={<AlertCircle className="h-3 w-3 sm:h-4 sm:w-4" />}
          label="Overdue"
          value={stats.overdueTasks}
          secondary="Past due date"
          accent="rose"
          emphasizeValue
        />
        <StatCell
          icon={<CheckCircle2 className="h-3 w-3 sm:h-4 sm:w-4" />}
          label="Completed"
          value={stats.completedTasks}
          secondary="Finished tasks"
          accent="emerald"
          emphasizeValue
        />
      </CardContent>
    </Card>
  );
}

// ── Sub-component ───────────────────────────────────────────────────────────

type Accent = "amber" | "blue" | "rose" | "emerald";

interface StatCellProps {
  icon: React.ReactNode;
  label: string;
  value: number;
  secondary: string;
  accent: Accent;
  /** Tint the primary number with the accent (used for overdue/completed). */
  emphasizeValue?: boolean;
}

const CHIP_CLASSES: Record<Accent, string> = {
  amber: "bg-amber-100 text-amber-700 dark:bg-amber-950/50 dark:text-amber-300",
  blue: "bg-blue-100 text-blue-700 dark:bg-blue-950/50 dark:text-blue-300",
  rose: "bg-rose-100 text-rose-700 dark:bg-rose-950/50 dark:text-rose-300",
  emerald:
    "bg-emerald-100 text-emerald-700 dark:bg-emerald-950/50 dark:text-emerald-300",
};

const VALUE_CLASSES: Record<Accent, string> = {
  amber: "text-amber-600 dark:text-amber-400",
  blue: "text-blue-600 dark:text-blue-400",
  rose: "text-rose-600 dark:text-rose-400",
  emerald: "text-emerald-600 dark:text-emerald-400",
};

function StatCell({
  icon,
  label,
  value,
  secondary,
  accent,
  emphasizeValue,
}: StatCellProps) {
  return (
    <div className="p-2.5 sm:p-3 space-y-0.5 sm:space-y-1">
      <div className="flex items-center justify-between gap-1.5">
        <span className="text-[10px] sm:text-xs font-medium text-muted-foreground uppercase tracking-wider truncate">
          {label}
        </span>
        <span
          className={cn(
            "flex items-center justify-center rounded-md shrink-0 h-5 w-5 sm:h-6 sm:w-6",
            CHIP_CLASSES[accent],
          )}
        >
          {icon}
        </span>
      </div>
      <div
        className={cn(
          "text-base sm:text-xl font-bold tabular-nums leading-tight",
          emphasizeValue && VALUE_CLASSES[accent],
        )}
      >
        {value}
      </div>
      <p className="text-[10px] sm:text-[11px] text-muted-foreground leading-tight truncate">
        {secondary}
      </p>
    </div>
  );
}
