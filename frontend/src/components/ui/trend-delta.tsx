import * as React from "react";
import { ArrowDown, ArrowUp, Minus } from "lucide-react";
import { cn } from "@/lib/utils";

export interface TrendDeltaProps {
  /** Percentage change vs prior period. Sign carries direction. */
  value: number;
  /** Comparison label, e.g. "vs last week". Optional. */
  label?: string;
  /** When true, treat positive values as negative tone (e.g. rejection rate going up is bad). */
  invert?: boolean;
  className?: string;
}

export function TrendDelta({ value, label, invert, className }: TrendDeltaProps) {
  const direction: "up" | "down" | "flat" =
    value > 0 ? "up" : value < 0 ? "down" : "flat";

  // Tone: positive = green, negative = red, flat = muted.
  const positiveDirection: "up" | "down" = invert ? "down" : "up";
  const tone: "positive" | "negative" | "neutral" =
    direction === "flat"
      ? "neutral"
      : direction === positiveDirection
        ? "positive"
        : "negative";

  const TONE_CLASS: Record<typeof tone, string> = {
    positive: "text-emerald-600 dark:text-emerald-400",
    negative: "text-rose-600 dark:text-rose-400",
    neutral: "text-muted-foreground",
  };

  const Icon = direction === "up" ? ArrowUp : direction === "down" ? ArrowDown : Minus;

  return (
    <span
      data-testid="trend-delta"
      data-tone={tone}
      className={cn(
        "inline-flex items-center gap-0.5 text-xs font-medium tabular-nums",
        TONE_CLASS[tone],
        className
      )}
    >
      <Icon
        data-testid="trend-arrow"
        data-direction={direction}
        className="h-3 w-3"
        aria-hidden="true"
      />
      <span>{Math.abs(value).toFixed(1)}%</span>
      {label && <span className="text-muted-foreground ml-1">{label}</span>}
    </span>
  );
}
