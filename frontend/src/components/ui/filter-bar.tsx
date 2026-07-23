"use client";
import * as React from "react";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { X } from "lucide-react";
import { cn } from "@/lib/utils";

export interface FilterBarProps {
  /** Search field — usually the existing `<Search />` from `ui/search-field`. */
  search?: React.ReactNode;
  /** Filter controls (Selects, etc) — render in a wrap row. */
  filters?: React.ReactNode;
  /** Whether any filter is active. */
  hasActiveFilters?: boolean;
  onReset?: () => void;
  /** Right-side meta text (e.g. "Showing 12 results"). */
  meta?: React.ReactNode;
  className?: string;
}

export function FilterBar({
  search,
  filters,
  hasActiveFilters,
  onReset,
  meta,
  className,
}: FilterBarProps) {
  return (
    <Card className={cn("border-border/60", className)}>
      <CardContent className="p-3 sm:p-4 space-y-2">
        <div className="grid gap-2 md:grid-cols-[1fr_auto] md:items-start">
          <div className="min-w-0">{search}</div>
          {filters && (
            <div className="flex flex-wrap gap-2 md:justify-end">{filters}</div>
          )}
        </div>
        {(meta || hasActiveFilters) && (
          <div className="flex items-center justify-between gap-2">
            <span className="text-xs text-muted-foreground">{meta}</span>
            {hasActiveFilters && onReset && (
              <Button variant="ghost" size="sm" onClick={onReset} className="h-7 text-xs">
                <X className="h-3 w-3 mr-1" />
                Reset
              </Button>
            )}
          </div>
        )}
      </CardContent>
    </Card>
  );
}
