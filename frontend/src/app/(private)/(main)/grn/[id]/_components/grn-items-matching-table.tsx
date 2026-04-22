"use client";

import type { GRNItem } from "@/types/goods-received-note";
import { cn } from "@/lib/utils";

interface GRNItemsMatchingTableProps {
  items: GRNItem[];
}

const CONDITION_BADGE: Record<string, string> = {
  good: "bg-green-100 text-green-800 dark:bg-green-900/40 dark:text-green-200",
  damaged: "bg-red-100 text-red-800 dark:bg-red-900/40 dark:text-red-200",
  missing:
    "bg-amber-100 text-amber-800 dark:bg-amber-900/40 dark:text-amber-200",
};

export function GRNItemsMatchingTable({ items }: GRNItemsMatchingTableProps) {
  const totalOrdered = items.reduce(
    (sum, i) => sum + (i.quantityOrdered || 0),
    0,
  );
  const totalReceived = items.reduce(
    (sum, i) => sum + (i.quantityReceived || 0),
    0,
  );
  const totalVariance = totalReceived - totalOrdered;

  return (
    <div className="rounded-lg border border-border overflow-hidden text-sm">
      {/* Column header — desktop */}
      <div className="hidden sm:grid grid-cols-[2rem_1fr_4rem_4rem_4.5rem_6rem] gap-3 px-4 py-2 bg-muted/60 border-b border-border">
        <span className="text-xs font-medium text-muted-foreground">#</span>
        <span className="text-xs font-medium text-muted-foreground uppercase tracking-wider">
          Description
        </span>
        <span className="text-xs font-medium text-muted-foreground uppercase tracking-wider text-center">
          Ordered
        </span>
        <span className="text-xs font-medium text-muted-foreground uppercase tracking-wider text-center">
          Received
        </span>
        <span className="text-xs font-medium text-muted-foreground uppercase tracking-wider text-center">
          Variance
        </span>
        <span className="text-xs font-medium text-muted-foreground uppercase tracking-wider text-center">
          Condition
        </span>
      </div>

      {/* Rows */}
      <div className="divide-y divide-border/60">
        {items.map((item, index) => {
          const variance = item.variance ?? 0;
          const isPositive = variance > 0;
          const isNegative = variance < 0;
          const conditionKey = (item.condition || "").toLowerCase();

          return (
            <div
              key={item.id || index}
              className="grid grid-cols-[2rem_1fr_auto] sm:grid-cols-[2rem_1fr_4rem_4rem_4.5rem_6rem] items-center gap-x-3 gap-y-0.5 px-4 py-2.5 hover:bg-muted/30 transition-colors"
            >
              {/* # */}
              <span className="text-xs text-muted-foreground/60 font-mono tabular-nums self-start pt-0.5">
                {String(index + 1).padStart(2, "0")}
              </span>

              {/* Description + optional notes */}
              <div className="min-w-0">
                <p className="font-medium leading-snug truncate">
                  {item.description || "—"}
                </p>
                {item.notes && (
                  <p className="text-xs text-muted-foreground truncate mt-0.5">
                    {item.notes}
                  </p>
                )}
                {/* Mobile inline stats */}
                <p className="text-xs text-muted-foreground mt-0.5 sm:hidden">
                  <span className="tabular-nums">
                    {item.quantityReceived}
                  </span>
                  <span> of </span>
                  <span className="tabular-nums">{item.quantityOrdered}</span>
                  {variance !== 0 && (
                    <span
                      className={cn(
                        "ml-1 font-medium",
                        isNegative ? "text-red-600" : "text-green-600",
                      )}
                    >
                      ({isPositive ? "+" : ""}
                      {variance})
                    </span>
                  )}
                </p>
              </div>

              {/* Condition badge — mobile (inline right) */}
              <span
                className={cn(
                  "sm:hidden px-2 py-0.5 rounded text-[10px] font-medium uppercase tracking-wider",
                  CONDITION_BADGE[conditionKey] ||
                    "bg-gray-100 text-gray-800",
                )}
              >
                {item.condition}
              </span>

              {/* Ordered — desktop */}
              <span className="hidden sm:block text-center text-muted-foreground tabular-nums">
                {item.quantityOrdered}
              </span>

              {/* Received — desktop */}
              <span className="hidden sm:block text-center font-medium tabular-nums">
                {item.quantityReceived}
              </span>

              {/* Variance — desktop */}
              <span
                className={cn(
                  "hidden sm:block text-center font-semibold tabular-nums",
                  isNegative && "text-red-600 dark:text-red-400",
                  isPositive && "text-green-600 dark:text-green-400",
                  !variance && "text-muted-foreground",
                )}
              >
                {isPositive ? "+" : ""}
                {variance}
              </span>

              {/* Condition — desktop */}
              <span className="hidden sm:flex justify-center">
                <span
                  className={cn(
                    "px-2 py-0.5 rounded text-xs font-medium capitalize",
                    CONDITION_BADGE[conditionKey] ||
                      "bg-gray-100 text-gray-800",
                  )}
                >
                  {item.condition || "—"}
                </span>
              </span>
            </div>
          );
        })}
      </div>

      {/* Summary footer */}
      <div className="grid grid-cols-[1fr_auto] sm:grid-cols-[2rem_1fr_4rem_4rem_4.5rem_6rem] items-center gap-3 px-4 py-2.5 bg-muted/40 border-t border-border">
        {/* Mobile layout */}
        <div className="sm:hidden flex items-center gap-2">
          <span className="text-xs text-muted-foreground">
            {items.length} item{items.length !== 1 ? "s" : ""}
          </span>
          {totalVariance !== 0 && (
            <span
              className={cn(
                "text-xs font-medium",
                totalVariance < 0 ? "text-red-600" : "text-green-600",
              )}
            >
              · {totalVariance > 0 ? "+" : ""}
              {totalVariance} variance
            </span>
          )}
        </div>
        <span className="sm:hidden font-bold tabular-nums text-right">
          {totalReceived} / {totalOrdered}
        </span>

        {/* Desktop layout */}
        <span className="hidden sm:block" />
        <span className="hidden sm:block text-xs font-medium text-muted-foreground uppercase tracking-wider">
          Totals
        </span>
        <span className="hidden sm:block text-center font-semibold tabular-nums">
          {totalOrdered}
        </span>
        <span className="hidden sm:block text-center font-semibold tabular-nums">
          {totalReceived}
        </span>
        <span
          className={cn(
            "hidden sm:block text-center font-bold tabular-nums",
            totalVariance < 0 && "text-red-600 dark:text-red-400",
            totalVariance > 0 && "text-green-600 dark:text-green-400",
            !totalVariance && "text-muted-foreground",
          )}
        >
          {totalVariance > 0 ? "+" : ""}
          {totalVariance}
        </span>
        <span className="hidden sm:block" />
      </div>
    </div>
  );
}
