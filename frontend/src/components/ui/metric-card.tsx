import * as React from "react";
import { Card } from "@/components/ui/card";
import { TrendDelta } from "@/components/ui/trend-delta";
import { cn } from "@/lib/utils";

export type MetricAccent = "blue" | "emerald" | "amber" | "rose" | "slate" | "violet" | "warm";

const CHIP: Record<MetricAccent, string> = {
  blue: "bg-blue-100 text-blue-700 dark:bg-blue-950/50 dark:text-blue-300",
  emerald: "bg-emerald-100 text-emerald-700 dark:bg-emerald-950/50 dark:text-emerald-300",
  amber: "bg-amber-100 text-amber-700 dark:bg-amber-950/50 dark:text-amber-300",
  rose: "bg-rose-100 text-rose-700 dark:bg-rose-950/50 dark:text-rose-300",
  slate: "bg-slate-100 text-slate-700 dark:bg-slate-800 dark:text-slate-300",
  violet: "bg-violet-100 text-violet-700 dark:bg-violet-950/50 dark:text-violet-300",
  warm: "bg-accent-warm/15 text-accent-warm dark:bg-accent-warm/20",
};

export interface MetricCardProps {
  title: string;
  value: number | string;
  icon: React.ReactNode;
  secondary?: React.ReactNode;
  accent?: MetricAccent;
  trend?: { value: number; label?: string; invert?: boolean };
  /** Numeric series for an inline sparkline. Shows last ~7 points. */
  sparkline?: number[];
  className?: string;
}

function Sparkline({ data, className }: { data: number[]; className?: string }) {
  if (!data.length) return null;
  const w = 80;
  const h = 24;
  const min = Math.min(...data);
  const max = Math.max(...data);
  const range = max - min || 1;
  const stepX = w / Math.max(data.length - 1, 1);
  const points = data
    .map((v, i) => `${i * stepX},${h - ((v - min) / range) * h}`)
    .join(" ");
  return (
    <svg
      data-testid="metric-sparkline"
      viewBox={`0 0 ${w} ${h}`}
      className={cn("h-6 w-20 overflow-visible", className)}
      aria-hidden="true"
    >
      <polyline
        fill="none"
        stroke="currentColor"
        strokeWidth="1.5"
        strokeLinecap="round"
        strokeLinejoin="round"
        points={points}
      />
    </svg>
  );
}

export function MetricCard({
  title,
  value,
  icon,
  secondary,
  accent = "blue",
  trend,
  sparkline,
  className,
}: MetricCardProps) {
  return (
    <Card className={cn("border-border/60 p-4 space-y-2", className)}>
      <div className="flex items-start justify-between gap-2">
        <span className="text-xs font-medium text-muted-foreground uppercase tracking-wider truncate">
          {title}
        </span>
        <span
          className={cn(
            "flex items-center justify-center rounded-md shrink-0 h-7 w-7",
            CHIP[accent]
          )}
        >
          {icon}
        </span>
      </div>
      <div className="flex items-end justify-between gap-3">
        <div className="text-2xl sm:text-3xl font-bold tabular-nums leading-none">
          {value}
        </div>
        {sparkline && sparkline.length > 0 && (
          <Sparkline data={sparkline} className={cn("text-foreground/70")} />
        )}
      </div>
      {(secondary || trend) && (
        <div className="flex items-center justify-between gap-2 text-xs text-muted-foreground">
          {secondary && <span className="truncate">{secondary}</span>}
          {trend && <TrendDelta value={trend.value} label={trend.label} invert={trend.invert} />}
        </div>
      )}
    </Card>
  );
}
