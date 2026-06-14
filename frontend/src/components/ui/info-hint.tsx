"use client";

import * as React from "react";
import { Info } from "lucide-react";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { cn } from "@/lib/utils";

interface InfoHintProps {
  /** Popover body — keep it short; this is supplementary detail, not the primary UI. */
  children: React.ReactNode;
  /** Accessible label for the trigger button. */
  label?: string;
  /** Optional visible text beside the icon (e.g. on a mobile-only trigger). */
  triggerLabel?: React.ReactNode;
  /** Extra classes for the trigger button. */
  className?: string;
  /** Extra classes for the popover content. */
  contentClassName?: string;
  /** Extra classes for the icon. */
  iconClassName?: string;
  side?: "top" | "right" | "bottom" | "left";
  align?: "start" | "center" | "end";
}

/**
 * Small info icon that reveals supplementary text in a popover. Use as a
 * compact fallback (typically mobile-only) for detail that's shown inline on
 * larger screens. Pass `triggerLabel` to show text beside the icon.
 */
export function InfoHint({
  children,
  label = "More information",
  triggerLabel,
  className,
  contentClassName,
  iconClassName,
  side = "top",
  align = "center",
}: InfoHintProps) {
  return (
    <Popover>
      <PopoverTrigger asChild>
        <button
          type="button"
          aria-label={label}
          onClick={(e) => e.stopPropagation()}
          className={cn(
            triggerLabel
              ? "inline-flex shrink-0 items-center gap-1 rounded-md text-xs text-muted-foreground transition-colors hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
              : "inline-flex size-5 shrink-0 items-center justify-center rounded-full text-muted-foreground/70 transition-colors hover:bg-muted hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring",
            className,
          )}
        >
          <Info className={cn("size-3.5", iconClassName)} />
          {triggerLabel ? <span>{triggerLabel}</span> : null}
        </button>
      </PopoverTrigger>
      <PopoverContent
        side={side}
        align={align}
        onClick={(e) => e.stopPropagation()}
        className={cn(
          "w-72 max-w-[min(18rem,calc(100vw-2rem))] text-sm leading-relaxed",
          contentClassName,
        )}
      >
        {children}
      </PopoverContent>
    </Popover>
  );
}
