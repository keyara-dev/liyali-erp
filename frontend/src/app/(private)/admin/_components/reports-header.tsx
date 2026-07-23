"use client";
import * as React from "react";
import { Download, RefreshCw } from "lucide-react";
import { Button } from "@/components/ui/button";
import { PageHeader } from "@/components/base/page-header";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import CalendarDateRangePicker from "@/components/ui/date-range-picker";
import { cn } from "@/lib/utils";

export type ReportsExportFormat = "csv";

export interface ReportsHeaderProps {
  title: string;
  subtitle: string;
  /** ISO YYYY-MM-DD. Required — owner is the parent. */
  from: string;
  /** ISO YYYY-MM-DD. */
  to: string;
  onRangeChange: (from: string, to: string) => void;
  onRefresh: () => void;
  onExport: (format: ReportsExportFormat) => void;
  isRefreshing: boolean;
  className?: string;
}

export function ReportsHeader({
  title,
  subtitle,
  from,
  to,
  onRangeChange,
  onRefresh,
  onExport,
  isRefreshing,
  className,
}: ReportsHeaderProps) {
  const initialFromDate = React.useMemo(() => new Date(from), [from]);
  const initialToDate = React.useMemo(() => new Date(to), [to]);

  return (
    <div className={cn("space-y-3", className)}>
      <PageHeader
        title={title}
        subtitle={subtitle}
        showBackButton={false}
        actions={
          <div className="flex flex-wrap items-center gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={onRefresh}
              disabled={isRefreshing}
            >
              <RefreshCw
                className={cn("h-4 w-4 mr-2", isRefreshing && "animate-spin")}
              />
              Refresh
            </Button>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="outline" size="sm">
                  <Download className="h-4 w-4 mr-2" />
                  Export
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <DropdownMenuItem onClick={() => onExport("csv")}>
                  Export current view (CSV)
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        }
      />
      <div className="flex items-center gap-2">
        <span className="text-xs font-medium text-muted-foreground uppercase tracking-wider">
          Date range
        </span>
        <div className="flex-1 max-w-sm">
          <CalendarDateRangePicker
            initialFrom={initialFromDate}
            initialTo={initialToDate}
            onChange={onRangeChange}
          />
        </div>
      </div>
    </div>
  );
}
