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
  Filter,
  X,
  Calendar as CalendarIcon,
  Search,
  RotateCcw,
} from "lucide-react";
import { format } from "date-fns";
import { type OrganizationFilters } from "@/app/_actions/organizations";

interface OrganizationAdvancedFiltersProps {
  filters: OrganizationFilters;
  onFiltersChange: (filters: OrganizationFilters) => void;
  onReset: () => void;
}

const SUBSCRIPTION_TIERS = [
  { value: "basic", label: "Basic" },
  { value: "professional", label: "Professional" },
  { value: "enterprise", label: "Enterprise" },
];

const TRIAL_STATUSES = [
  { value: "trial", label: "Trial" },
  { value: "subscribed", label: "Subscribed" },
  { value: "expired", label: "Expired" },
];

const SORT_OPTIONS = [
  { value: "name", label: "Name" },
  { value: "created_at", label: "Created Date" },
  { value: "user_count", label: "User Count" },
  { value: "trial_end_date", label: "Trial End Date" },
];

export function OrganizationAdvancedFilters({
  filters,
  onFiltersChange,
  onReset,
}: OrganizationAdvancedFiltersProps) {
  const [showAdvanced, setShowAdvanced] = useState(false);
  const [dateFrom, setDateFrom] = useState<Date>();
  const [dateTo, setDateTo] = useState<Date>();

  const updateFilter = (key: keyof OrganizationFilters, value: any) => {
    onFiltersChange({
      ...filters,
      [key]: value,
      page: 1, // Reset to first page when filters change
    });
  };

  const getActiveFiltersCount = () => {
    let count = 0;
    if (filters.search) count++;
    if (filters.status && filters.status !== "all") count++;
    if (filters.subscription_tier) count++;
    if (filters.trial_status) count++;
    return count;
  };

  const clearAllFilters = () => {
    setDateFrom(undefined);
    setDateTo(undefined);
    onReset();
  };

  const activeFiltersCount = getActiveFiltersCount();

  return (
    <div className="space-y-4">
      {/* Basic Search and Quick Filters */}
      <div className="flex flex-col sm:flex-row gap-4">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Search organizations by name or domain..."
            value={filters.search || ""}
            onChange={(e) => updateFilter("search", e.target.value)}
            className="pl-10"
          />
        </div>

        <div className="flex gap-2">
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
        </div>
      </div>

      {/* Status Quick Filters */}
      <div className="flex flex-wrap gap-2">
        <Button
          variant={
            filters.status === "all" || !filters.status ? "default" : "outline"
          }
          size="sm"
          onClick={() => updateFilter("status", "all")}
        >
          All Organizations
        </Button>
        <Button
          variant={filters.status === "active" ? "default" : "outline"}
          size="sm"
          onClick={() => updateFilter("status", "active")}
        >
          Active
        </Button>
        <Button
          variant={filters.status === "suspended" ? "default" : "outline"}
          size="sm"
          onClick={() => updateFilter("status", "suspended")}
        >
          Suspended
        </Button>
        <Button
          variant={filters.status === "pending" ? "default" : "outline"}
          size="sm"
          onClick={() => updateFilter("status", "pending")}
        >
          Pending
        </Button>
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
            {/* Subscription Tier Filter */}
            <div className="space-y-2">
              <Label htmlFor="tier-filter">Subscription Tier</Label>
              <Select
                value={filters.subscription_tier || ""}
                onValueChange={(value) =>
                  updateFilter("subscription_tier", value || undefined)
                }
              >
                <SelectTrigger>
                  <SelectValue placeholder="All tiers" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="">All tiers</SelectItem>
                  {SUBSCRIPTION_TIERS.map((tier) => (
                    <SelectItem key={tier.value} value={tier.value}>
                      {tier.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            {/* Trial Status Filter */}
            <div className="space-y-2">
              <Label htmlFor="trial-filter">Trial Status</Label>
              <Select
                value={filters.trial_status || ""}
                onValueChange={(value) =>
                  updateFilter("trial_status", value || undefined)
                }
              >
                <SelectTrigger>
                  <SelectValue placeholder="All statuses" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="">All statuses</SelectItem>
                  {TRIAL_STATUSES.map((status) => (
                    <SelectItem key={status.value} value={status.value}>
                      {status.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            {/* Sort By */}
            <div className="space-y-2">
              <Label htmlFor="sort-filter">Sort by</Label>
              <div className="flex gap-2">
                <Select
                  value={filters.sort_by || "created_at"}
                  onValueChange={(value) => updateFilter("sort_by", value)}
                >
                  <SelectTrigger className="flex-1">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {SORT_OPTIONS.map((option) => (
                      <SelectItem key={option.value} value={option.value}>
                        {option.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() =>
                    updateFilter(
                      "sort_order",
                      filters.sort_order === "asc" ? "desc" : "asc",
                    )
                  }
                >
                  {filters.sort_order === "asc" ? "↑" : "↓"}
                </Button>
              </div>
            </div>

            {/* Results per page */}
            <div className="space-y-2">
              <Label htmlFor="limit-filter">Results per page</Label>
              <Select
                value={filters.limit?.toString() || "20"}
                onValueChange={(value) =>
                  updateFilter("limit", parseInt(value))
                }
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="10">10</SelectItem>
                  <SelectItem value="20">20</SelectItem>
                  <SelectItem value="50">50</SelectItem>
                  <SelectItem value="100">100</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>

          {/* Date Range Filters */}
          <div className="space-y-2">
            <Label>Creation Date Range</Label>
            <div className="flex gap-2">
              <Popover>
                <PopoverTrigger asChild>
                  <Button
                    variant="outline"
                    className="justify-start text-left font-normal"
                  >
                    <CalendarIcon className="mr-2 h-4 w-4" />
                    {dateFrom ? format(dateFrom, "PPP") : "From date"}
                  </Button>
                </PopoverTrigger>
                <PopoverContent className="w-auto p-0" align="start">
                  <Calendar
                    mode="single"
                    selected={dateFrom}
                    onSelect={setDateFrom}
                    initialFocus
                  />
                </PopoverContent>
              </Popover>

              <Popover>
                <PopoverTrigger asChild>
                  <Button
                    variant="outline"
                    className="justify-start text-left font-normal"
                  >
                    <CalendarIcon className="mr-2 h-4 w-4" />
                    {dateTo ? format(dateTo, "PPP") : "To date"}
                  </Button>
                </PopoverTrigger>
                <PopoverContent className="w-auto p-0" align="start">
                  <Calendar
                    mode="single"
                    selected={dateTo}
                    onSelect={setDateTo}
                    initialFocus
                  />
                </PopoverContent>
              </Popover>
            </div>
          </div>
        </div>
      )}

      {/* Active Filters Display */}
      {activeFiltersCount > 0 && (
        <div className="flex flex-wrap gap-2">
          <span className="text-sm text-muted-foreground">Active filters:</span>
          {filters.search && (
            <Badge variant="secondary" className="gap-1">
              Search: {filters.search}
              <X
                className="h-3 w-3 cursor-pointer"
                onClick={() => updateFilter("search", "")}
              />
            </Badge>
          )}
          {filters.status && filters.status !== "all" && (
            <Badge variant="secondary" className="gap-1">
              Status: {filters.status}
              <X
                className="h-3 w-3 cursor-pointer"
                onClick={() => updateFilter("status", "all")}
              />
            </Badge>
          )}
          {filters.subscription_tier && (
            <Badge variant="secondary" className="gap-1">
              Tier: {filters.subscription_tier}
              <X
                className="h-3 w-3 cursor-pointer"
                onClick={() => updateFilter("subscription_tier", undefined)}
              />
            </Badge>
          )}
          {filters.trial_status && (
            <Badge variant="secondary" className="gap-1">
              Trial: {filters.trial_status}
              <X
                className="h-3 w-3 cursor-pointer"
                onClick={() => updateFilter("trial_status", undefined)}
              />
            </Badge>
          )}
        </div>
      )}
    </div>
  );
}
