"use client";
import * as React from "react";
import { format, subDays, startOfDay, endOfDay } from "date-fns";
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
import { useDateRangeUrlState } from "@/hooks/use-date-range-url-state";
import { cn } from "@/lib/utils";

export type ReportsExportFormat = "csv";

export interface ReportsHeaderProps {
  title: string;
  subtitle: string;
  onRefresh: () => void;
  onExport: (format: ReportsExportFormat) => void;
  isRefreshing: boolean;
  className?: string;
}

function defaultRange() {
  const today = new Date();
  const from = format(startOfDay(subDays(today, 27)), "yyyy-MM-dd");
  const to = format(endOfDay(today), "yyyy-MM-dd");
  return { from, to };
}

export function ReportsHeader({
  title,
  subtitle,
  onRefresh,
  onExport,
  isRefreshing,
  className,
}: ReportsHeaderProps) {
  const initial = React.useMemo(defaultRange, []);
  const { from, to, setRange } = useDateRangeUrlState({
    defaultFrom: initial.from,
    defaultTo: initial.to,
  });

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
            onChange={(newFrom, newTo) => setRange(newFrom, newTo)}
          />
        </div>
      </div>
    </div>
  );
}
