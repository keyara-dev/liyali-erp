"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
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
  Clock,
} from "lucide-react";
import { format } from "date-fns";
import { cn } from "@/lib/utils";
import type { APIFilters } from "@/app/_actions/api-monitoring";

interface APIMonitoringFiltersProps {
  filters: APIFilters;
  onFiltersChange: (filters: APIFilters) => void;
  onReset: () => void;
  onExport: (
    type: "endpoints" | "metrics" | "errors" | "alerts",
    format: "csv" | "json" | "excel",
  ) => void;
  searchTerm: string;
  onSearchChange: (search: string) => void;
  categories?: string[];
}

export function APIMonitoringFiltersComponent({
  filters,
  onFiltersChange,
  onReset,
  onExport,
  searchTerm,
  onSearchChange,
  categories = [],
}: APIMonitoringFiltersProps) {
  const [showAdvanced, setShowAdvanced] = useState(false);
  const [startDate, setStartDate] = useState<Date>();
  const [endDate, setEndDate] = useState<Date>();

  const activeFiltersCount = Object.keys(filters).filter(
    (key) => filters[key as keyof APIFilters] !== undefined,
  ).length;

  const handleFilterChange = (key: keyof APIFilters, value: any) => {
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

  const httpMethods = [
    "GET",
    "POST",
    "PUT",
    "DELETE",
    "PATCH",
    "HEAD",
    "OPTIONS",
  ];
  const statusCodes = [200, 201, 400, 401, 403, 404, 422, 500, 502, 503, 504];
  const errorTypes = [
    "timeout",
    "validation",
    "authentication",
    "authorization",
    "server_error",
    "network",
  ];
  const severityLevels = ["low", "medium", "high", "critical"];

  return (
    <div className="space-y-4">
      {/* Search and Quick Filters */}
      <div className="flex flex-col sm:flex-row gap-4">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground h-4 w-4" />
          <Input
            placeholder="Search endpoints, paths, errors..."
            value={searchTerm}
            onChange={(e) => onSearchChange(e.target.value)}
            className="pl-10"
          />
        </div>

        <div className="flex items-center gap-2">
          {/* Time Range Quick Select */}
          <Select
            value={filters.time_range || "24h"}
            onValueChange={handleTimeRangeChange}
          >
            <SelectTrigger className="w-40">
              <Clock className="mr-2 h-4 w-4" />
              <SelectValue placeholder="Time range" />
            </SelectTrigger>
            <SelectContent>
              {timeRanges.map((range) => (
                <SelectItem key={range.value} value={range.value}>
                  {range.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>

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
              <DropdownMenuLabel>Export Data</DropdownMenuLabel>
              <DropdownMenuSeparator />

              <DropdownMenuLabel className="text-xs text-muted-foreground">
                Endpoints
              </DropdownMenuLabel>
              <DropdownMenuItem onClick={() => onExport("endpoints", "csv")}>
                Endpoints (CSV)
              </DropdownMenuItem>
              <DropdownMenuItem onClick={() => onExport("endpoints", "excel")}>
                Endpoints (Excel)
              </DropdownMenuItem>

              <DropdownMenuSeparator />
              <DropdownMenuLabel className="text-xs text-muted-foreground">
                Metrics
              </DropdownMenuLabel>
              <DropdownMenuItem onClick={() => onExport("metrics", "csv")}>
                Metrics (CSV)
              </DropdownMenuItem>
              <DropdownMenuItem onClick={() => onExport("metrics", "json")}>
                Metrics (JSON)
              </DropdownMenuItem>

              <DropdownMenuSeparator />
              <DropdownMenuLabel className="text-xs text-muted-foreground">
                Errors & Alerts
              </DropdownMenuLabel>
              <DropdownMenuItem onClick={() => onExport("errors", "csv")}>
                Errors (CSV)
              </DropdownMenuItem>
              <DropdownMenuItem onClick={() => onExport("alerts", "csv")}>
                Alerts (CSV)
              </DropdownMenuItem>
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
            {/* HTTP Method Filter */}
            <div className="space-y-2">
              <Label htmlFor="method-filter">HTTP Method</Label>
              <Select
                value={filters.method || "all"}
                onValueChange={(value) =>
                  handleFilterChange(
                    "method",
                    value === "all" ? undefined : value,
                  )
                }
              >
                <SelectTrigger id="method-filter">
                  <SelectValue placeholder="All methods" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All Methods</SelectItem>
                  {httpMethods.map((method) => (
                    <SelectItem key={method} value={method}>
                      {method}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            {/* Category Filter */}
            <div className="space-y-2">
              <Label htmlFor="category-filter">Category</Label>
              <Select
                value={filters.category || "all"}
                onValueChange={(value) =>
                  handleFilterChange(
                    "category",
                    value === "all" ? undefined : value,
                  )
                }
              >
                <SelectTrigger id="category-filter">
                  <SelectValue placeholder="All categories" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All Categories</SelectItem>
                  {categories.map((category) => (
                    <SelectItem key={category} value={category}>
                      {category}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            {/* Public/Private Filter */}
            <div className="space-y-2">
              <Label htmlFor="visibility-filter">Visibility</Label>
              <Select
                value={
                  filters.is_public === undefined
                    ? "all"
                    : filters.is_public
                      ? "public"
                      : "private"
                }
                onValueChange={(value) =>
                  handleFilterChange(
                    "is_public",
                    value === "all"
                      ? undefined
                      : value === "public"
                        ? true
                        : false,
                  )
                }
              >
                <SelectTrigger id="visibility-filter">
                  <SelectValue placeholder="All endpoints" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All Endpoints</SelectItem>
                  <SelectItem value="public">Public</SelectItem>
                  <SelectItem value="private">Private</SelectItem>
                </SelectContent>
              </Select>
            </div>

            {/* Deprecated Filter */}
            <div className="space-y-2">
              <Label htmlFor="deprecated-filter">Status</Label>
              <Select
                value={
                  filters.is_deprecated === undefined
                    ? "all"
                    : filters.is_deprecated
                      ? "deprecated"
                      : "active"
                }
                onValueChange={(value) =>
                  handleFilterChange(
                    "is_deprecated",
                    value === "all"
                      ? undefined
                      : value === "deprecated"
                        ? true
                        : false,
                  )
                }
              >
                <SelectTrigger id="deprecated-filter">
                  <SelectValue placeholder="All statuses" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All Statuses</SelectItem>
                  <SelectItem value="active">Active</SelectItem>
                  <SelectItem value="deprecated">Deprecated</SelectItem>
                </SelectContent>
              </Select>
            </div>

            {/* Status Code Filter */}
            <div className="space-y-2">
              <Label htmlFor="status-filter">Status Code</Label>
              <Select
                value={filters.status_code?.toString() || "all"}
                onValueChange={(value) =>
                  handleFilterChange(
                    "status_code",
                    value === "all" ? undefined : parseInt(value),
                  )
                }
              >
                <SelectTrigger id="status-filter">
                  <SelectValue placeholder="All codes" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All Status Codes</SelectItem>
                  {statusCodes.map((code) => (
                    <SelectItem key={code} value={code.toString()}>
                      {code}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            {/* Error Type Filter */}
            <div className="space-y-2">
              <Label htmlFor="error-type-filter">Error Type</Label>
              <Select
                value={filters.error_type || "all"}
                onValueChange={(value) =>
                  handleFilterChange(
                    "error_type",
                    value === "all" ? undefined : value,
                  )
                }
              >
                <SelectTrigger id="error-type-filter">
                  <SelectValue placeholder="All error types" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All Error Types</SelectItem>
                  {errorTypes.map((type) => (
                    <SelectItem key={type} value={type}>
                      {type.replace("_", " ").toUpperCase()}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            {/* Severity Filter */}
            <div className="space-y-2">
              <Label htmlFor="severity-filter">Alert Severity</Label>
              <Select
                value={filters.severity || "all"}
                onValueChange={(value) =>
                  handleFilterChange(
                    "severity",
                    value === "all" ? undefined : value,
                  )
                }
              >
                <SelectTrigger id="severity-filter">
                  <SelectValue placeholder="All severities" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All Severities</SelectItem>
                  {severityLevels.map((severity) => (
                    <SelectItem key={severity} value={severity}>
                      {severity.charAt(0).toUpperCase() + severity.slice(1)}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
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
              {filters.method && (
                <Badge variant="secondary" className="text-xs">
                  Method: {filters.method}
                  <button
                    onClick={() => handleFilterChange("method", undefined)}
                    className="ml-1 hover:text-destructive"
                  >
                    <X className="h-3 w-3" />
                  </button>
                </Badge>
              )}
              {filters.category && (
                <Badge variant="secondary" className="text-xs">
                  Category: {filters.category}
                  <button
                    onClick={() => handleFilterChange("category", undefined)}
                    className="ml-1 hover:text-destructive"
                  >
                    <X className="h-3 w-3" />
                  </button>
                </Badge>
              )}
              {filters.is_public !== undefined && (
                <Badge variant="secondary" className="text-xs">
                  Visibility: {filters.is_public ? "Public" : "Private"}
                  <button
                    onClick={() => handleFilterChange("is_public", undefined)}
                    className="ml-1 hover:text-destructive"
                  >
                    <X className="h-3 w-3" />
                  </button>
                </Badge>
              )}
              {filters.status_code && (
                <Badge variant="secondary" className="text-xs">
                  Status: {filters.status_code}
                  <button
                    onClick={() => handleFilterChange("status_code", undefined)}
                    className="ml-1 hover:text-destructive"
                  >
                    <X className="h-3 w-3" />
                  </button>
                </Badge>
              )}
              {filters.severity && (
                <Badge variant="secondary" className="text-xs">
                  Severity: {filters.severity}
                  <button
                    onClick={() => handleFilterChange("severity", undefined)}
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
