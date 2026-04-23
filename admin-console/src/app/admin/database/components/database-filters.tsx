"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { SelectField } from "@/components/ui/select-field";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Badge } from "@/components/ui/badge";
import { Calendar } from "@/components/ui/calendar";
import {
  Search,
  Filter,
  X,
  Download,
  Calendar as CalendarIcon,
} from "lucide-react";
import { format } from "date-fns";
import { cn } from "@/lib/utils";
import type {
  DatabaseFilters,
  DatabaseConnection,
} from "@/app/_actions/database";

interface DatabaseFiltersProps {
  filters: DatabaseFilters;
  onFiltersChange: (filters: DatabaseFilters) => void;
  onReset: () => void;
  onExport: (connectionId: string, format: "sql" | "csv" | "json") => void;
  searchTerm: string;
  onSearchChange: (search: string) => void;
  connections?: DatabaseConnection[];
}

export function DatabaseFiltersComponent({
  filters,
  onFiltersChange,
  onReset,
  onExport,
  searchTerm,
  onSearchChange,
  connections = [],
}: DatabaseFiltersProps) {
  const [showAdvanced, setShowAdvanced] = useState(false);
  const [startDate, setStartDate] = useState<Date>();
  const [endDate, setEndDate] = useState<Date>();

  const activeFiltersCount = Object.keys(filters).filter(
    (key) => filters[key as keyof DatabaseFilters] !== undefined,
  ).length;

  const handleFilterChange = (key: keyof DatabaseFilters, value: any) => {
    const newFilters = { ...filters };
    if (value === undefined || value === "" || value === "all") {
      delete newFilters[key];
    } else {
      newFilters[key] = value;
    }
    onFiltersChange(newFilters);
  };

  const handleDateRangeChange = (
    type: "start_date" | "end_date",
    date: Date | undefined,
  ) => {
    if (type === "start_date") {
      setStartDate(date);
      handleFilterChange("start_date", date?.toISOString());
    } else {
      setEndDate(date);
      handleFilterChange("end_date", date?.toISOString());
    }
  };

  const handleTimeRangeChange = (range: string) => {
    handleFilterChange("time_range", range);
    // Clear custom date range when using preset
    if (range !== "custom") {
      setStartDate(undefined);
      setEndDate(undefined);
      handleFilterChange("start_date", undefined);
      handleFilterChange("end_date", undefined);
    }
  };

  const clearAllFilters = () => {
    setStartDate(undefined);
    setEndDate(undefined);
    onReset();
  };

  const timeRanges = [
    { value: "1h", label: "Last Hour" },
    { value: "6h", label: "Last 6 Hours" },
    { value: "24h", label: "Last 24 Hours" },
    { value: "7d", label: "Last 7 Days" },
    { value: "30d", label: "Last 30 Days" },
    { value: "custom", label: "Custom Range" },
  ];

  const databaseTypes = [
    "postgresql",
    "mysql",
    "mongodb",
    "redis",
    "elasticsearch",
  ];
  const statusOptions = ["connected", "disconnected", "error", "maintenance"];

  return (
    <div className="space-y-4">
      {/* Search and Quick Filters */}
      <div className="flex flex-col sm:flex-row gap-4">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground h-4 w-4" />
          <Input
            placeholder="Search connections, databases, tables..."
            value={searchTerm}
            onChange={(e) => onSearchChange(e.target.value)}
            className="pl-10"
          />
        </div>

        <div className="flex items-center gap-2">
          {/* Time Range Quick Select */}
          <SelectField
            options={timeRanges}
            value={filters.time_range || "24h"}
            onValueChange={handleTimeRangeChange}
            classNames={{ wrapper: "w-40" }}
          />

          <Button
            variant="outline"
            onClick={() => setShowAdvanced(!showAdvanced)}
            className="relative"
          >
            <Filter className="mr-2 h-4 w-4" />
            Filters
            {activeFiltersCount > 0 && (
              <Badge
                variant="destructive"
                className="ml-2 h-5 w-5 rounded-full p-0 text-xs"
              >
                {activeFiltersCount}
              </Badge>
            )}
          </Button>

          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="outline">
                <Download className="mr-2 h-4 w-4" />
                Export
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="w-48">
              <DropdownMenuLabel>Export Database</DropdownMenuLabel>
              <DropdownMenuSeparator />

              {connections.length > 0 ? (
                connections.map((connection, index) => (
                  <div
                    key={
                      connection.id ||
                      `${connection.name || "connection"}-${index}`
                    }
                  >
                    <DropdownMenuLabel className="text-xs text-muted-foreground">
                      {connection.name}
                    </DropdownMenuLabel>
                    <DropdownMenuItem
                      onClick={() => onExport(connection.id, "sql")}
                    >
                      SQL Dump
                    </DropdownMenuItem>
                    <DropdownMenuItem
                      onClick={() => onExport(connection.id, "csv")}
                    >
                      CSV Export
                    </DropdownMenuItem>
                    <DropdownMenuItem
                      onClick={() => onExport(connection.id, "json")}
                    >
                      JSON Export
                    </DropdownMenuItem>
                    <DropdownMenuSeparator />
                  </div>
                ))
              ) : (
                <DropdownMenuItem disabled>
                  No connections available
                </DropdownMenuItem>
              )}
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </div>

      {/* Advanced Filters */}
      {showAdvanced && (
        <div className="border rounded-lg p-4 space-y-4 bg-muted/50">
          <div className="flex items-center justify-between">
            <h3 className="text-sm font-medium">Advanced Filters</h3>
            <Button
              variant="ghost"
              size="sm"
              onClick={clearAllFilters}
              className="text-muted-foreground hover:text-foreground"
            >
              <X className="mr-1 h-3 w-3" />
              Clear All
            </Button>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            {/* Connection Filter */}
            <SelectField
              label="Connection"
              placeholder="All connections"
              options={[
                { value: "all", label: "All Connections" },
                ...connections.map((c) => ({ value: c.id, label: c.name })),
              ]}
              value={filters.connection_id || "all"}
              onValueChange={(value) => handleFilterChange("connection_id", value === "all" ? undefined : value)}
            />

            {/* Database Type Filter */}
            <SelectField
              label="Database Type"
              placeholder="All types"
              options={[
                { value: "all", label: "All Types" },
                ...databaseTypes.map((t) => ({ value: t, label: t.charAt(0).toUpperCase() + t.slice(1) })),
              ]}
              value={filters.type || "all"}
              onValueChange={(value) => handleFilterChange("type", value === "all" ? undefined : value)}
            />

            {/* Status Filter */}
            <SelectField
              label="Status"
              placeholder="All statuses"
              options={[
                { value: "all", label: "All Statuses" },
                ...statusOptions.map((s) => ({ value: s, label: s.charAt(0).toUpperCase() + s.slice(1) })),
              ]}
              value={filters.status || "all"}
              onValueChange={(value) => handleFilterChange("status", value === "all" ? undefined : value)}
            />
          </div>

          {/* Custom Date Range */}
          {filters.time_range === "custom" && (
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4 pt-4 border-t">
              <div className="space-y-2">
                <Label>Start Date</Label>
                <Popover>
                  <PopoverTrigger asChild>
                    <Button
                      variant="outline"
                      className={cn(
                        "w-full justify-start text-left font-normal",
                        !startDate && "text-muted-foreground",
                      )}
                    >
                      <CalendarIcon className="mr-2 h-4 w-4" />
                      {startDate ? (
                        format(startDate, "PPP")
                      ) : (
                        <span>Pick start date</span>
                      )}
                    </Button>
                  </PopoverTrigger>
                  <PopoverContent className="w-auto p-0" align="start">
                    <Calendar
                      mode="single"
                      selected={startDate}
                      onSelect={(date) =>
                        handleDateRangeChange("start_date", date)
                      }
                      initialFocus
                    />
                  </PopoverContent>
                </Popover>
              </div>

              <div className="space-y-2">
                <Label>End Date</Label>
                <Popover>
                  <PopoverTrigger asChild>
                    <Button
                      variant="outline"
                      className={cn(
                        "w-full justify-start text-left font-normal",
                        !endDate && "text-muted-foreground",
                      )}
                    >
                      <CalendarIcon className="mr-2 h-4 w-4" />
                      {endDate ? (
                        format(endDate, "PPP")
                      ) : (
                        <span>Pick end date</span>
                      )}
                    </Button>
                  </PopoverTrigger>
                  <PopoverContent className="w-auto p-0" align="start">
                    <Calendar
                      mode="single"
                      selected={endDate}
                      onSelect={(date) =>
                        handleDateRangeChange("end_date", date)
                      }
                      initialFocus
                    />
                  </PopoverContent>
                </Popover>
              </div>
            </div>
          )}

          {/* Active Filters Display */}
          {activeFiltersCount > 0 && (
            <div className="flex flex-wrap gap-2 pt-2 border-t">
              <span className="text-sm text-muted-foreground">
                Active filters:
              </span>
              {filters.connection_id && (
                <Badge variant="secondary" className="text-xs">
                  Connection:{" "}
                  {connections.find((c) => c.id === filters.connection_id)
                    ?.name || filters.connection_id}
                  <button
                    onClick={() =>
                      handleFilterChange("connection_id", undefined)
                    }
                    className="ml-1 hover:text-destructive"
                  >
                    <X className="h-3 w-3" />
                  </button>
                </Badge>
              )}
              {filters.type && (
                <Badge variant="secondary" className="text-xs">
                  Type: {filters.type}
                  <button
                    onClick={() => handleFilterChange("type", undefined)}
                    className="ml-1 hover:text-destructive"
                  >
                    <X className="h-3 w-3" />
                  </button>
                </Badge>
              )}
              {filters.status && (
                <Badge variant="secondary" className="text-xs">
                  Status: {filters.status}
                  <button
                    onClick={() => handleFilterChange("status", undefined)}
                    className="ml-1 hover:text-destructive"
                  >
                    <X className="h-3 w-3" />
                  </button>
                </Badge>
              )}
              {filters.time_range && filters.time_range !== "24h" && (
                <Badge variant="secondary" className="text-xs">
                  Time:{" "}
                  {
                    timeRanges.find((r) => r.value === filters.time_range)
                      ?.label
                  }
                  <button
                    onClick={() => handleFilterChange("time_range", undefined)}
                    className="ml-1 hover:text-destructive"
                  >
                    <X className="h-3 w-3" />
                  </button>
                </Badge>
              )}
            </div>
          )}
        </div>
      )}
    </div>
  );
}
