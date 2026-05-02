import * as React from "react";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";

export type PriorityValue = "URGENT" | "HIGH" | "MEDIUM" | "LOW" | string;

export interface PriorityBadgeProps {
  /** Priority string. Falls back to "MEDIUM" if undefined. Case-insensitive for variant. */
  priority?: PriorityValue;
  className?: string;
}

const VARIANT: Record<string, "destructive" | "warning" | "info"> = {
  URGENT: "destructive",
  HIGH: "destructive",
  MEDIUM: "warning",
  LOW: "info",
};

export function PriorityBadge({ priority, className }: PriorityBadgeProps) {
  const display = priority || "MEDIUM";
  const variant = VARIANT[display.toUpperCase()] || "info";
  return (
    <Badge
      variant={variant}
      className={cn(
        "px-2 py-0.5 rounded text-[10px] uppercase font-medium tracking-wider",
        className
      )}
    >
      {display}
    </Badge>
  );
}
