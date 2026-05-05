"use client";
import * as React from "react";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Skeleton } from "@/components/ui/skeleton";
import { cn } from "@/lib/utils";

export interface DataListColumn<T> {
  id: string;
  header: React.ReactNode;
  cell: (row: T) => React.ReactNode;
  /**
   * Visibility breakpoint for the column.
   * 'always' = always visible on desktop table
   * 'md' = visible md+
   * 'lg' = visible lg+
   */
  priority?: "always" | "md" | "lg";
  /**
   * Text alignment for header and cell on the desktop table only.
   * The mobile card layout is owned by the `mobileCard` render prop and
   * ignores `align`. Default `"left"`.
   */
  align?: "left" | "right" | "center";
  className?: string;
}

export interface DataListProps<T> {
  rows: T[];
  columns: DataListColumn<T>[];
  getRowId: (row: T) => string;
  mobileCard: (row: T) => React.ReactNode;
  isLoading?: boolean;
  /** Number of skeleton rows to render while loading. Default 5. */
  skeletonRows?: number;
  emptyMessage?: React.ReactNode;
  onRowClick?: (row: T) => void;
  className?: string;
}

const HIDE: Record<NonNullable<DataListColumn<unknown>["priority"]>, string> = {
  always: "",
  md: "hidden md:table-cell",
  lg: "hidden lg:table-cell",
};

const ALIGN: Record<NonNullable<DataListColumn<unknown>["align"]>, string> = {
  left: "",
  right: "text-right",
  center: "text-center",
};

/**
 * Returns true when a click/keydown event originates inside an interactive
 * descendant (button, link, form control) so that row-level onRowClick is
 * suppressed.  Note: [role='button'] is intentionally excluded here because
 * the row wrapper itself carries that role; checking it would always match.
 * Explicit <button> / <a> tags are sufficient for the action-cell guard.
 */
function shouldIgnoreRowClick(target: EventTarget | null): boolean {
  if (!(target instanceof HTMLElement)) return false;
  return Boolean(
    target.closest("button, a, input, select, textarea, label")
  );
}

export function DataList<T>({
  rows,
  columns,
  getRowId,
  mobileCard,
  isLoading,
  skeletonRows = 5,
  emptyMessage = "No results.",
  onRowClick,
  className,
}: DataListProps<T>) {
  if (isLoading) {
    return (
      <div className={className}>
        {/* Desktop skeleton */}
        <div className="hidden md:block rounded-md border overflow-x-auto">
          <Table>
            <TableHeader>
              <TableRow>
                {columns.map((c) => (
                  <TableHead key={c.id} className={cn(HIDE[c.priority || "always"], ALIGN[c.align || "left"])}>
                    {c.header}
                  </TableHead>
                ))}
              </TableRow>
            </TableHeader>
            <TableBody>
              {Array.from({ length: skeletonRows }).map((_, i) => (
                <TableRow key={i}>
                  {columns.map((c) => (
                    <TableCell key={c.id} className={cn(HIDE[c.priority || "always"], ALIGN[c.align || "left"])}>
                      <Skeleton className="h-4 w-24" />
                    </TableCell>
                  ))}
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
        {/* Mobile skeleton */}
        <div className="md:hidden space-y-2">
          {Array.from({ length: skeletonRows }).map((_, i) => (
            <div key={i} className="rounded-md border p-3 space-y-2">
              <Skeleton className="h-4 w-2/3" />
              <Skeleton className="h-3 w-1/2" />
              <Skeleton className="h-3 w-1/3" />
            </div>
          ))}
        </div>
      </div>
    );
  }

  if (rows.length === 0) {
    return (
      <div className={cn("rounded-md border py-10 text-center text-sm text-muted-foreground", className)}>
        {emptyMessage}
      </div>
    );
  }

  return (
    <div className={className}>
      {/* Desktop / tablet: table */}
      <div className="hidden md:block rounded-md border overflow-x-auto">
        <Table>
          <TableHeader>
            <TableRow>
              {columns.map((c) => (
                <TableHead key={c.id} className={cn(HIDE[c.priority || "always"], ALIGN[c.align || "left"], c.className)}>
                  {c.header}
                </TableHead>
              ))}
            </TableRow>
          </TableHeader>
          <TableBody>
            {rows.map((row) => (
              <TableRow
                key={getRowId(row)}
                role={onRowClick ? "button" : undefined}
                tabIndex={onRowClick ? 0 : undefined}
                onClick={
                  onRowClick
                    ? (e) => {
                        if (!shouldIgnoreRowClick(e.target)) onRowClick(row);
                      }
                    : undefined
                }
                onKeyDown={
                  onRowClick
                    ? (e) => {
                        if (e.key === "Enter" || e.key === " ") {
                          if (e.key === " ") e.preventDefault();
                          if (!shouldIgnoreRowClick(e.target)) onRowClick(row);
                        }
                      }
                    : undefined
                }
                className={onRowClick ? "cursor-pointer" : undefined}
              >
                {columns.map((c) => (
                  <TableCell key={c.id} className={cn(HIDE[c.priority || "always"], ALIGN[c.align || "left"], c.className)}>
                    {c.cell(row)}
                  </TableCell>
                ))}
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>
      {/* Mobile: card stack */}
      <div className="md:hidden space-y-2">
        {rows.map((row) => (
          <div
            key={getRowId(row)}
            role={onRowClick ? "button" : undefined}
            tabIndex={onRowClick ? 0 : undefined}
            onClick={
              onRowClick
                ? (e) => {
                    if (!shouldIgnoreRowClick(e.target)) onRowClick(row);
                  }
                : undefined
            }
            onKeyDown={
              onRowClick
                ? (e) => {
                    if (e.key === "Enter" || e.key === " ") {
                      if (e.key === " ") e.preventDefault();
                      if (!shouldIgnoreRowClick(e.target)) onRowClick(row);
                    }
                  }
                : undefined
            }
            className={cn(
              "rounded-md border bg-card p-3 transition-colors",
              onRowClick && "cursor-pointer active:bg-muted/40"
            )}
          >
            {mobileCard(row)}
          </div>
        ))}
      </div>
    </div>
  );
}
