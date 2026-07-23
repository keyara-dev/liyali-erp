"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { SelectField } from "@/components/ui/select-field";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { Calendar } from "@/components/ui/calendar";
import { Badge } from "@/components/ui/badge";
import {
  Filter,
  Calendar as CalendarIcon,
  X,
  RotateCcw,
  Download,
  Search,
} from "lucide-react";
import { format } from "date-fns";
import { type AuditLogFilters } from "@/app/_actions/audit-logs";

interface AuditLogFiltersComponentProps {
  filters: AuditLogFilters;
  onFiltersChange: (filters: AuditLogFilters) => void;
  onReset: () => void;
  onExport: (format: "csv" | "json" | "pdf") => void;
  searchTerm: string;
  onSearchChange: (search: string) => void;
}

const DATE_RANGES = [
  { value: "1h", label: "Last hour" },
  { value: "24h", label: "Last 24 hours" },
  { value: "7d", label: "Last 7 days" },
  { value: "30d", label: "Last 30 days" },
  { value: "90d", label: "Last 90 days" },
  { value: "custom", label: "Custom range" },
];

const ACTION_TYPES = [
  { value: "create", label: "Create" },
  { value: "update", label: "Update" },
  { value: "delete", label: "Delete" },
  { value: "view", label: "View" },
  { value: "login", label: "Login" },
  { value: "logout", label: "Logout" },
  { value: "export", label: "Export" },
  { value: "import", label: "Import" },
  { value: "system", label: "System" },
];

const RESOURCE_TYPES = [
  { value: "user", label: "User" },
  { value: "organization", label: "Organization" },
  { value: "subscription", label: "Subscription" },
  { value: "document", label: "Document" },
  { value: "workflow", label: "Workflow" },
  { value: "system", label: "System" },
];

const SEVERITIES = [
  { value: "low", label: "Low" },
  { value: "medium", label: "Medium" },
  { value: "high", label: "High" },
  { value: "critical", label: "Critical" },
];

const STATUSES = [
  { value: "success", label: "Success" },
  { value: "failure", label: "Failure" },
  { value: "warning", label: "Warning" },
];

export function AuditLogFiltersComponent({
  filters,
  onFiltersChange,
  onReset,
  onExport,
  searchTerm,
  onSearchChange,
}: AuditLogFiltersComponentProps) {
  const [showAdvanced, setShowAdvanced] = useState(false);
  const [startDate, setStartDate] = useState<Date>();
  const [endDate, setEndDate] = useState<Date>();

  const updateFilter = (key: keyof AuditLogFilters, value: any) => {
    onFiltersChange({
      ...filters,
      [key]: value,
    });
  };

  const handleDateRangeChange = (range: string) => {
    updateFilter("date_range", range as any);
    if (range !== "custom") {
      updateFilter("start_date", undefined);
      updateFilter("end_date", undefined);
      setStartDate(undefined);
      setEndDate(undefined);
    }
  };

  const handleCustomDateChange = () => {
    if (startDate && endDate) {
      updateFilter("start_date", format(startDate, "yyyy-MM-dd"));
      updateFilter("end_date", format(endDate, "yyyy-MM-dd"));
    }
  };

  const getActiveFiltersCount = () => {
    let count = 0;
    if (filters.user_id) count++;
    if (filters.organization_id) count++;
    if (filters.action_type) count++;
    if (filters.resource_type) count++;
    if (filters.severity) count++;
    if (filters.status) count++;
    if (filters.ip_address) count++;
    if (
      filters.date_range === "custom" &&
      filters.start_date &&
      filters.end_date
    )
      count++;
    return count;
  };

  const clearAllFilters = () => {
    setStartDate(undefined);
    setEndDate(undefined);
    onSearchChange("");
    onReset();
  };

  const activeFiltersCount = getActiveFiltersCount();

  return (
    <div className="space-y-4">
      {/* Search and Basic Filters */}
      <div className="flex flex-col sm:flex-row gap-4">
        <div className="flex-1">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground h-4 w-4" />
            <Input
              placeholder="Search audit logs..."
              value={searchTerm}
              onChange={(e) => onSearchChange(e.target.value)}
              className="pl-10"
            />
          </div>
        </div>

        <div className="flex items-center gap-4">
          <SelectField
            label="Date Range"
            options={DATE_RANGES}
            value={filters.date_range}
            onValueChange={handleDateRangeChange}
            classNames={{ wrapper: "w-40" }}
          />

          {filters.date_range === "custom" && (
            <div className="flex items-center gap-2">
              <Popover>
                <PopoverTrigger asChild>
                  <Button variant="outline" size="sm">
                    <CalendarIcon className="mr-2 h-4 w-4" />
                    {startDate ? format(startDate, "MMM dd") : "Start date"}
                  </Button>
                </PopoverTrigger>
                <PopoverContent className="w-auto p-0" align="start">
                  <Calendar
                    mode="single"
                    selected={startDate}
                    onSelect={(date) => {
                      setStartDate(date);
                      if (date && endDate) handleCustomDateChange();
                    }}
                  />
                </PopoverContent>
              </Popover>

              <Popover>
                <PopoverTrigger asChild>
                  <Button variant="outline" size="sm">
                    <CalendarIcon className="mr-2 h-4 w-4" />
                    {endDate ? format(endDate, "MMM dd") : "End date"}
                  </Button>
                </PopoverTrigger>
                <PopoverContent className="w-auto p-0" align="start">
                  <Calendar
                    mode="single"
                    selected={endDate}
                    onSelect={(date) => {
                      setEndDate(date);
                      if (startDate && date) handleCustomDateChange();
                    }}
                  />
                </PopoverContent>
              </Popover>
            </div>
          )}
        </div>

        <div className="flex items-center gap-2">
          <Button
            variant={showAdvanced ? "default" : "outline"}
            size="sm"
            onClick={() => setShowAdvanced(!showAdvanced)}
            className="relative"
          >
            <Filter className="mr-2 h-4 w-4" />
            Advanced Filters
            {activeFiltersCount > 0 && (
              <Badge
                variant="secondary"
                className="ml-2 h-5 w-5 rounded-full p-0 text-xs"
              >
                {activeFiltersCount}
              </Badge>
            )}
          </Button>

          {activeFiltersCount > 0 && (
            <Button variant="outline" size="sm" onClick={clearAllFilters}>
              <RotateCcw className="mr-2 h-4 w-4" />
              Reset
            </Button>
          )}

          <Popover>
            <PopoverTrigger asChild>
              <Button variant="outline" size="sm">
                <Download className="mr-2 h-4 w-4" />
                Export
              </Button>
            </PopoverTrigger>
            <PopoverContent className="w-48" align="end">
              <div className="space-y-2">
                <Button
                  variant="ghost"
                  size="sm"
                  className="w-full justify-start"
                  onClick={() => onExport("csv")}
                >
                  Export as CSV
                </Button>
                <Button
                  variant="ghost"
                  size="sm"
                  className="w-full justify-start"
                  onClick={() => onExport("json")}
                >
                  Export as JSON
                </Button>
                <Button
                  variant="ghost"
                  size="sm"
                  className="w-full justify-start"
                  onClick={() => onExport("pdf")}
                >
                  Export as PDF
                </Button>
              </div>
            </PopoverContent>
          </Popover>
        </div>
      </div>

      {/* Advanced Filters Panel */}
      {showAdvanced && (
        <div className="rounded-lg border p-4 space-y-4 bg-muted/20">
          <div className="flex items-center justify-between">
            <h3 className="text-sm font-medium">Advanced Filters</h3>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setShowAdvanced(false)}
            >
              <X className="h-4 w-4" />
            </Button>
          </div>

          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
            <SelectField
              label="Action Type"
              placeholder="All actions"
              options={ACTION_TYPES}
              value={filters.action_type || ""}
              onValueChange={(value) =>
                updateFilter("action_type", value || undefined)
              }
            />

            <SelectField
              label="Resource Type"
              placeholder="All resources"
              options={RESOURCE_TYPES}
              value={filters.resource_type || ""}
              onValueChange={(value) =>
                updateFilter("resource_type", value || undefined)
              }
            />

            <SelectField
              label="Severity"
              placeholder="All severities"
              options={SEVERITIES}
              value={filters.severity || ""}
              onValueChange={(value) =>
                updateFilter("severity", value || undefined)
              }
            />

            <SelectField
              label="Status"
              placeholder="All statuses"
              options={STATUSES}
              value={filters.status || ""}
              onValueChange={(value) =>
                updateFilter("status", value || undefined)
              }
            />

            <Input
              label="IP Address"
              placeholder="Filter by IP address"
              value={filters.ip_address || ""}
              onChange={(e) =>
                updateFilter("ip_address", e.target.value || undefined)
              }
            />
          </div>
        </div>
      )}

      {/* Active Filters Display */}
      {activeFiltersCount > 0 && (
        <div className="flex flex-wrap gap-2">
          <span className="text-sm text-muted-foreground">Active filters:</span>
          {filters.action_type && (
            <Badge variant="secondary" className="gap-1">
              Action: {filters.action_type}
              <X
                className="h-3 w-3 cursor-pointer"
                onClick={() => updateFilter("action_type", undefined)}
              />
            </Badge>
          )}
          {filters.resource_type && (
            <Badge variant="secondary" className="gap-1">
              Resource: {filters.resource_type}
              <X
                className="h-3 w-3 cursor-pointer"
                onClick={() => updateFilter("resource_type", undefined)}
              />
            </Badge>
          )}
          {filters.severity && (
            <Badge variant="secondary" className="gap-1">
              Severity: {filters.severity}
              <X
                className="h-3 w-3 cursor-pointer"
                onClick={() => updateFilter("severity", undefined)}
              />
            </Badge>
          )}
          {filters.status && (
            <Badge variant="secondary" className="gap-1">
              Status: {filters.status}
              <X
                className="h-3 w-3 cursor-pointer"
                onClick={() => updateFilter("status", undefined)}
              />
            </Badge>
          )}
          {filters.ip_address && (
            <Badge variant="secondary" className="gap-1">
              IP: {filters.ip_address}
              <X
                className="h-3 w-3 cursor-pointer"
                onClick={() => updateFilter("ip_address", undefined)}
              />
            </Badge>
          )}
          {filters.date_range === "custom" &&
            filters.start_date &&
            filters.end_date && (
              <Badge variant="secondary" className="gap-1">
                Custom: {filters.start_date} to {filters.end_date}
                <X
                  className="h-3 w-3 cursor-pointer"
                  onClick={() => {
                    updateFilter("date_range", "24h");
                    updateFilter("start_date", undefined);
                    updateFilter("end_date", undefined);
                  }}
                />
              </Badge>
            )}
        </div>
      )}
    </div>
  );
}
