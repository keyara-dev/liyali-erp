import * as React from "react";
import { Card, CardContent } from "@/components/ui/card";
import { cn } from "@/lib/utils";

export type StatAccent = "amber" | "blue" | "rose" | "emerald" | "slate" | "violet" | "warm";

const CHIP: Record<StatAccent, string> = {
  amber: "bg-amber-100 text-amber-700 dark:bg-amber-950/50 dark:text-amber-300",
  blue: "bg-blue-100 text-blue-700 dark:bg-blue-950/50 dark:text-blue-300",
  rose: "bg-rose-100 text-rose-700 dark:bg-rose-950/50 dark:text-rose-300",
  emerald: "bg-emerald-100 text-emerald-700 dark:bg-emerald-950/50 dark:text-emerald-300",
  slate: "bg-slate-100 text-slate-700 dark:bg-slate-800 dark:text-slate-300",
  violet: "bg-violet-100 text-violet-700 dark:bg-violet-950/50 dark:text-violet-300",
  warm: "bg-accent-warm/15 text-accent-warm dark:bg-accent-warm/20",
};

const VALUE: Partial<Record<StatAccent, string>> = {
  amber: "text-amber-600 dark:text-amber-400",
  blue: "text-blue-600 dark:text-blue-400",
  rose: "text-rose-600 dark:text-rose-400",
  emerald: "text-emerald-600 dark:text-emerald-400",
  warm: "text-accent-warm",
};

export interface StatItem {
  label: string;
  value: number | string;
  icon: React.ReactNode;
  accent: StatAccent;
  secondary?: string;
  emphasizeValue?: boolean;
}

export interface StatGridProps {
  items: StatItem[];
  /** Tailwind grid-cols class for base; defaults to a sensible 2/3/N pattern. */
  className?: string;
}

export function StatGrid({ items, className }: StatGridProps) {
  const cols = items.length;
  const mdCols =
    cols >= 5 ? "md:grid-cols-5" :
    cols === 4 ? "md:grid-cols-4" :
    cols === 3 ? "md:grid-cols-3" :
    "md:grid-cols-2";
  return (
    <Card className="border-border/60 p-0">
      <CardContent
        className={cn(
          "grid grid-cols-2 sm:grid-cols-3 divide-y sm:divide-y-0 sm:divide-x divide-border/60 p-0",
          mdCols,
          className
        )}
      >
        {items.map((it, idx) => (
          <div key={`${it.label}-${idx}`} className="p-2.5 sm:p-3 space-y-0.5 sm:space-y-1">
            <div className="flex items-center justify-between gap-1.5">
              <span className="text-[10px] sm:text-xs font-medium text-muted-foreground uppercase tracking-wider truncate">
                {it.label}
              </span>
              <span
                className={cn(
                  "flex items-center justify-center rounded-md shrink-0 h-5 w-5 sm:h-6 sm:w-6",
                  CHIP[it.accent]
                )}
              >
                {it.icon}
              </span>
            </div>
            <div
              className={cn(
                "text-base sm:text-xl font-bold tabular-nums leading-tight",
                it.emphasizeValue && VALUE[it.accent]
              )}
            >
              {it.value}
            </div>
            {it.secondary && (
              <p className="text-[10px] sm:text-[11px] text-muted-foreground leading-tight truncate">
                {it.secondary}
              </p>
            )}
          </div>
        ))}
      </CardContent>
    </Card>
  );
}
