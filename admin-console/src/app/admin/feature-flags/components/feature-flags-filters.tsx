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
import { Calendar } from "@/components/ui/calendar";
import { Badge } from "@/components/ui/badge";
import {
  Search,
  Filter,
  Download,
  Upload,
  RefreshCw,
  X,
  Calendar as CalendarIcon,
  Flag,
  Tags,
} from "lucide-react";
import { format } from "date-fns";
import { cn } from "@/lib/utils";
import type { DateRange } from "react-day-picker";
import type { FeatureFlagFilters } from "@/app/_actions/feature-flags";

interface FeatureFlagsFiltersProps {
  filters: FeatureFlagFilters;
  onFiltersChange: (filters: FeatureFlagFilters) => void;
  onExport: () => void;
  onImport: () => void;
  onRefresh: () => void;
  isLoading?: boolean;
}

export function FeatureFlagsFilters({
  filters,
  onFiltersChange,
  onExport,
  onImport,
  onRefresh,
  isLoading = false,
}: FeatureFlagsFiltersProps) {
  const [showAdvanced, setShowAdvanced] = useState(false);
  const [dateRange, setDateRange] = useState<DateRange | undefined>();
  const [expiryDate, setExpiryDate] = useState<Date>();
  const [tagInput, setTagInput] = useState("");

  const categories = [
    { value: "feature", label: "Feature Flags" },
    { value: "experiment", label: "Experiments" },
    { value: "operational", label: "Operational" },
    { value: "killswitch", label: "Kill Switches" },
    { value: "permission", label: "Permissions" },
  ];

  const environments = [
    { value: "all", label: "All Environments" },
    { value: "production", label: "Production" },
    { value: "staging", label: "Staging" },
    { value: "development", label: "Development" },
  ];

  const types = [
    { value: "boolean", label: "Boolean" },
    { value: "string", label: "String" },
    { value: "number", label: "Number" },
    { value: "json", label: "JSON" },
  ];

  const handleFilterChange = (key: keyof FeatureFlagFilters, value: any) => {
    onFiltersChange({
      ...filters,
      [key]: value || undefined,
    });
  };

  const handleDateRangeChange = (range: DateRange | undefined) => {
    setDateRange(range);
    onFiltersChange({
      ...filters,
      createdAfter: range?.from?.toISOString(),
      createdBefore: range?.to?.toISOString(),
    });
  };

  const handleExpiryDateChange = (date?: Date) => {
    setExpiryDate(date);
    onFiltersChange({
      ...filters,
      expiringBefore: date?.toISOString(),
    });
  };

  const handleTagsChange = (value: string) => {
    setTagInput(value);
    const tags = value
      .split(",")
      .map((tag) => tag.trim())
      .filter((tag) => tag.length > 0);
    handleFilterChange("tags", tags.length > 0 ? tags : undefined);
  };

  const clearFilters = () => {
    onFiltersChange({});
    setDateRange(undefined);
    setExpiryDate(undefined);
    setTagInput("");
  };

  const getActiveFiltersCount = () => {
    return Object.values(filters).filter(
      (value) =>
        value !== undefined &&
        value !== "" &&
        value !== null &&
        !(Array.isArray(value) && value.length === 0),
    ).length;
  };

  return (
    <div className="space-y-4">
      {/* Main Filters Row */}
      <div className="flex flex-col sm:flex-row gap-4">
        {/* Search */}
        <div className="flex-1">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground h-4 w-4" />
            <Input
              placeholder="Search flags by key, name, or description..."
              value={filters.search || ""}
              onChange={(e) => handleFilterChange("search", e.target.value)}
              className="pl-10"
            />
          </div>
        </div>

        {/* Category Filter */}
        <Select
          value={filters.category || ""}
          onValueChange={(value) => handleFilterChange("category", value)}
        >
          <SelectTrigger className="w-[180px]">
            <SelectValue placeholder="All Categories" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="">All Categories</SelectItem>
            {categories.map((category) => (
              <SelectItem key={category.value} value={category.value}>
                {category.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>

        {/* Status Filter */}
        <Select
          value={filters.enabled?.toString() || ""}
          onValueChange={(value) =>
            handleFilterChange(
              "enabled",
              value === "" ? undefined : value === "true",
            )
          }
        >
          <SelectTrigger className="w-[140px]">
            <SelectValue placeholder="All Flags" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="">All Flags</SelectItem>
            <SelectItem value="true">Enabled</SelectItem>
            <SelectItem value="false">Disabled</SelectItem>
          </SelectContent>
        </Select>

        {/* Actions */}
        <div className="flex gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => setShowAdvanced(!showAdvanced)}
            className={cn(
              "relative",
              getActiveFiltersCount() > 0 && "border-primary",
            )}
          >
            <Filter className="h-4 w-4 mr-2" />
            Filters
            {getActiveFiltersCount() > 0 && (
              <Badge
                variant="secondary"
                className="ml-2 h-5 w-5 p-0 flex items-center justify-center text-xs"
              >
                {getActiveFiltersCount()}
              </Badge>
            )}
          </Button>

          <Button
            variant="outline"
            size="sm"
            onClick={onRefresh}
            disabled={isLoading}
          >
            <RefreshCw
              className={cn("h-4 w-4 mr-2", isLoading && "animate-spin")}
            />
            Refresh
          </Button>

          <Button variant="outline" size="sm" onClick={onExport}>
            <Download className="h-4 w-4 mr-2" />
            Export
          </Button>

          <Button variant="outline" size="sm" onClick={onImport}>
            <Upload className="h-4 w-4 mr-2" />
            Import
          </Button>
        </div>
      </div>

      {/* Advanced Filters */}
      {showAdvanced && (
        <div className="border rounded-lg p-4 space-y-4 bg-muted/50">
          <div className="flex items-center justify-between">
            <h3 className="text-sm font-medium flex items-center gap-2">
              <Flag className="h-4 w-4" />
              Advanced Filters
            </h3>
            <Button
              variant="ghost"
              size="sm"
              onClick={clearFilters}
              className="text-muted-foreground hover:text-foreground"
            >
              <X className="h-4 w-4 mr-1" />
              Clear All
            </Button>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            {/* Environment Filter */}
            <div className="space-y-2">
              <Label htmlFor="environment-filter">Environment</Label>
              <Select
                value={filters.environment || ""}
                onValueChange={(value) =>
                  handleFilterChange("environment", value)
                }
              >
                <SelectTrigger>
                  <SelectValue placeholder="All Environments" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="">All Environments</SelectItem>
                  {environments.map((env) => (
                    <SelectItem key={env.value} value={env.value}>
                      {env.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            {/* Type Filter */}
            <div className="space-y-2">
              <Label htmlFor="type-filter">Flag Type</Label>
              <Select
                value={filters.type || ""}
                onValueChange={(value) => handleFilterChange("type", value)}
              >
                <SelectTrigger>
                  <SelectValue placeholder="All Types" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="">All Types</SelectItem>
                  {types.map((type) => (
                    <SelectItem key={type.value} value={type.value}>
                      {type.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            {/* Archived Filter */}
            <div className="space-y-2">
              <Label htmlFor="archived-filter">Archive Status</Label>
              <Select
                value={filters.archived?.toString() || ""}
                onValueChange={(value) =>
                  handleFilterChange(
                    "archived",
                    value === "" ? undefined : value === "true",
                  )
                }
              >
                <SelectTrigger>
                  <SelectValue placeholder="All Flags" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="">All Flags</SelectItem>
                  <SelectItem value="false">Active</SelectItem>
                  <SelectItem value="true">Archived</SelectItem>
                </SelectContent>
              </Select>
            </div>

            {/* Tags Filter */}
            <div className="space-y-2">
              <Label htmlFor="tags-filter" className="flex items-center gap-1">
                <Tags className="h-3 w-3" />
                Tags
              </Label>
              <Input
                id="tags-filter"
                value={tagInput}
                onChange={(e) => handleTagsChange(e.target.value)}
                placeholder="tag1, tag2, tag3"
              />
              {filters.tags && filters.tags.length > 0 && (
                <div className="flex flex-wrap gap-1 mt-2">
                  {filters.tags.map((tag, index) => (
                    <Badge key={index} variant="secondary" className="text-xs">
                      {tag}
                    </Badge>
                  ))}
                </div>
              )}
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {/* Created Date Range Filter */}
            <div className="space-y-2">
              <Label>Created Date Range</Label>
              <Popover>
                <PopoverTrigger asChild>
                  <Button
                    variant="outline"
                    className={cn(
                      "w-full justify-start text-left font-normal",
                      !dateRange?.from && "text-muted-foreground",
                    )}
                  >
                    <CalendarIcon className="mr-2 h-4 w-4" />
                    {dateRange?.from ? (
                      dateRange.to ? (
                        <>
                          {format(dateRange.from, "LLL dd, y")} -{" "}
                          {format(dateRange.to, "LLL dd, y")}
                        </>
                      ) : (
                        format(dateRange.from, "LLL dd, y")
                      )
                    ) : (
                      "Pick a date range"
                    )}
                  </Button>
                </PopoverTrigger>
                <PopoverContent className="w-auto p-0" align="start">
                  <Calendar
                    initialFocus
                    mode="range"
                    defaultMonth={dateRange?.from}
                    selected={dateRange}
                    onSelect={handleDateRangeChange}
                    numberOfMonths={2}
                    required={false}
                  />
                </PopoverContent>
              </Popover>
            </div>

            {/* Expiry Date Filter */}
            <div className="space-y-2">
              <Label>Expiring Before</Label>
              <Popover>
                <PopoverTrigger asChild>
                  <Button
                    variant="outline"
                    className={cn(
                      "w-full justify-start text-left font-normal",
                      !expiryDate && "text-muted-foreground",
                    )}
                  >
                    <CalendarIcon className="mr-2 h-4 w-4" />
                    {expiryDate
                      ? format(expiryDate, "LLL dd, y")
                      : "Pick expiry date"}
                  </Button>
                </PopoverTrigger>
                <PopoverContent className="w-auto p-0" align="start">
                  <Calendar
                    mode="single"
                    selected={expiryDate}
                    onSelect={handleExpiryDateChange}
                    initialFocus
                  />
                </PopoverContent>
              </Popover>
            </div>
          </div>
        </div>
      )}

      {/* Active Filters Display */}
      {getActiveFiltersCount() > 0 && (
        <div className="flex flex-wrap gap-2">
          {filters.search && (
            <Badge variant="secondary" className="gap-1">
              Search: {filters.search}
              <X
                className="h-3 w-3 cursor-pointer"
                onClick={() => handleFilterChange("search", "")}
              />
            </Badge>
          )}
          {filters.category && (
            <Badge variant="secondary" className="gap-1">
              Category:{" "}
              {categories.find((c) => c.value === filters.category)?.label}
              <X
                className="h-3 w-3 cursor-pointer"
                onClick={() => handleFilterChange("category", "")}
              />
            </Badge>
          )}
          {filters.environment && (
            <Badge variant="secondary" className="gap-1">
              Environment:{" "}
              {environments.find((e) => e.value === filters.environment)?.label}
              <X
                className="h-3 w-3 cursor-pointer"
                onClick={() => handleFilterChange("environment", "")}
              />
            </Badge>
          )}
          {filters.type && (
            <Badge variant="secondary" className="gap-1">
              Type: {types.find((t) => t.value === filters.type)?.label}
              <X
                className="h-3 w-3 cursor-pointer"
                onClick={() => handleFilterChange("type", "")}
              />
            </Badge>
          )}
          {filters.enabled !== undefined && (
            <Badge variant="secondary" className="gap-1">
              {filters.enabled ? "Enabled Only" : "Disabled Only"}
              <X
                className="h-3 w-3 cursor-pointer"
                onClick={() => handleFilterChange("enabled", undefined)}
              />
            </Badge>
          )}
          {filters.archived !== undefined && (
            <Badge variant="secondary" className="gap-1">
              {filters.archived ? "Archived Only" : "Active Only"}
              <X
                className="h-3 w-3 cursor-pointer"
                onClick={() => handleFilterChange("archived", undefined)}
              />
            </Badge>
          )}
          {filters.tags && filters.tags.length > 0 && (
            <Badge variant="secondary" className="gap-1">
              Tags: {filters.tags.join(", ")}
              <X
                className="h-3 w-3 cursor-pointer"
                onClick={() => {
                  handleFilterChange("tags", undefined);
                  setTagInput("");
                }}
              />
            </Badge>
          )}
          {(filters.createdAfter || filters.createdBefore) && (
            <Badge variant="secondary" className="gap-1">
              Created Date Range
              <X
                className="h-3 w-3 cursor-pointer"
                onClick={() => {
                  handleFilterChange("createdAfter", undefined);
                  handleFilterChange("createdBefore", undefined);
                  setDateRange(undefined);
                }}
              />
            </Badge>
          )}
          {filters.expiringBefore && (
            <Badge variant="secondary" className="gap-1">
              Expiring Before:{" "}
              {format(new Date(filters.expiringBefore), "MMM dd, y")}
              <X
                className="h-3 w-3 cursor-pointer"
                onClick={() => {
                  handleFilterChange("expiringBefore", undefined);
                  setExpiryDate(undefined);
                }}
              />
            </Badge>
          )}
        </div>
      )}
    </div>
  );
}
