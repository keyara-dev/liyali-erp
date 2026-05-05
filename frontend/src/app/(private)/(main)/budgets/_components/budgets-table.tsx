"use client";

import { useMemo, useEffect, useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { Eye, FolderOpen } from "lucide-react";

import { DataList, DataListColumn } from "@/components/ui/data-list";
import { FilterBar } from "@/components/ui/filter-bar";
import {
  Empty,
  EmptyContent,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@/components/ui/empty";
import { StatusBadge } from "@/components/status-badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { CustomPagination } from "@/components/ui/custom-pagination";
import { useBudgets } from "@/hooks/use-budget-queries";
import { Budget } from "@/types/budget";
import { QUERY_KEYS } from "@/lib/constants";
import { useDebounce } from "@/hooks/use-debounce";

interface BudgetsTableProps {
  userRole: string;
  refreshTrigger: number;
  onBudgetAction: () => void;
  initialData?: Budget[];
}

const PAGE_SIZE = 10;

export function BudgetsTable({
  userRole,
  refreshTrigger,
  onBudgetAction,
  initialData,
}: BudgetsTableProps) {
  const router = useRouter();
  const queryClient = useQueryClient();
  const {
    data: budgetsData,
    isLoading,
    refetch,
  } = useBudgets(initialData);

  // Defensive: queryFn should always return [], but guard against null
  const budgets = budgetsData ?? [];

  // Refetch when refreshTrigger changes (after budget creation)
  useEffect(() => {
    if (refreshTrigger > 0) {
      refetch();
      queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.BUDGETS.ALL] });
    }
  }, [refreshTrigger, refetch, queryClient]);

  // Filter state
  const [searchQuery, setSearchQuery] = useState("");
  const [statusFilter, setStatusFilter] = useState<string>("all");
  const [yearFilter, setYearFilter] = useState<string>("all");
  const [departmentFilter, setDepartmentFilter] = useState<string>("all");
  const [page, setPage] = useState(1);
  const debouncedSearch = useDebounce(searchQuery, 500);

  // Reset to page 1 whenever filters change
  useEffect(() => {
    setPage(1);
  }, [debouncedSearch, statusFilter, yearFilter, departmentFilter]);

  // Dynamic filter options derived from data
  const fiscalYears = useMemo(
    () =>
      Array.from(new Set(budgets.map((b) => b.fiscalYear).filter(Boolean)))
        .sort()
        .reverse(),
    [budgets],
  );

  const departments = useMemo(
    () =>
      Array.from(new Set(budgets.map((b) => b.department).filter(Boolean))).sort(),
    [budgets],
  );

  // Client-side filtering
  const filteredBudgets = useMemo(() => {
    let filtered = budgets;
    if (statusFilter !== "all") {
      filtered = filtered.filter(
        (b) => b.status?.toLowerCase() === statusFilter.toLowerCase(),
      );
    }
    if (yearFilter !== "all") {
      filtered = filtered.filter((b) => String(b.fiscalYear) === yearFilter);
    }
    if (departmentFilter !== "all") {
      filtered = filtered.filter((b) => b.department === departmentFilter);
    }
    if (debouncedSearch) {
      const s = debouncedSearch.toLowerCase();
      filtered = filtered.filter(
        (b) =>
          b.name?.toLowerCase().includes(s) ||
          b.budgetCode?.toLowerCase().includes(s) ||
          b.description?.toLowerCase?.().includes(s),
      );
    }
    return filtered;
  }, [budgets, statusFilter, yearFilter, departmentFilter, debouncedSearch]);

  // Pagination
  const totalRows = filteredBudgets.length;
  const totalPages = Math.max(1, Math.ceil(totalRows / PAGE_SIZE));
  const safePage = Math.min(page, totalPages);
  const pagedBudgets = filteredBudgets.slice(
    (safePage - 1) * PAGE_SIZE,
    safePage * PAGE_SIZE,
  );

  const paginationData = useMemo(
    () => ({
      page: safePage,
      limit: PAGE_SIZE,
      total: totalRows,
      totalPages,
      hasNext: safePage < totalPages,
      hasPrev: safePage > 1,
      page_size: PAGE_SIZE,
      total_pages: totalPages,
      totalCount: totalRows,
      has_next: safePage < totalPages,
      has_prev: safePage > 1,
    }),
    [safePage, totalRows, totalPages],
  );

  const hasActiveFilters =
    Boolean(searchQuery) ||
    statusFilter !== "all" ||
    yearFilter !== "all" ||
    departmentFilter !== "all";

  const clearFilters = () => {
    setSearchQuery("");
    setStatusFilter("all");
    setYearFilter("all");
    setDepartmentFilter("all");
  };

  // DataList columns
  const columns: DataListColumn<Budget>[] = [
    {
      id: "budgetCode",
      header: "Budget Code",
      priority: "always",
      cell: (row) => (
        <div className="font-semibold">{row.budgetCode}</div>
      ),
    },
    {
      id: "name",
      header: "Name",
      priority: "md",
      cell: (row) => (
        <div>
          <div className="font-medium line-clamp-1">{row.name}</div>
          {row.description && (
            <div className="text-xs text-muted-foreground line-clamp-1">
              {row.description}
            </div>
          )}
        </div>
      ),
    },
    {
      id: "department",
      header: "Department",
      priority: "md",
      cell: (row) => <div>{row.department ?? "—"}</div>,
    },
    {
      id: "fiscalYear",
      header: "FY",
      priority: "md",
      align: "right",
      cell: (row) => <div className="tabular-nums">{row.fiscalYear}</div>,
    },
    {
      id: "totalBudget",
      header: "Total Budget",
      priority: "md",
      align: "right",
      cell: (row) => (
        <div className="font-medium tabular-nums">
          K{(row.totalBudget || 0).toLocaleString()}
        </div>
      ),
    },
    {
      id: "utilization",
      header: "Used",
      priority: "lg",
      align: "right",
      cell: (row) => {
        const used = row.allocatedAmount ?? 0;
        const total = row.totalBudget ?? 0;
        const pct = total > 0 ? Math.round((used / total) * 100) : 0;
        return <span className="tabular-nums">{pct}%</span>;
      },
    },
    {
      id: "approvalStage",
      header: "Stage",
      priority: "lg",
      cell: (row) => {
        const status = row.status;
        const stage = row.approvalStage;
        return (
          <div className="text-sm">
            {status?.toUpperCase() === "DRAFT"
              ? "Not submitted"
              : status?.toUpperCase() === "APPROVED"
                ? "Completed"
                : status?.toUpperCase() === "REJECTED"
                  ? "Rejected"
                  : stage > 0
                    ? `Stage ${stage}`
                    : "Pending"}
          </div>
        );
      },
    },
    {
      id: "status",
      header: "Status",
      priority: "always",
      cell: (row) => <StatusBadge status={row.status} type="document" />,
    },
    {
      id: "actions",
      header: <span className="sr-only">Actions</span>,
      cell: (row) => (
        <Button
          size="sm"
          variant="outline"
          onClick={() => router.push(`/budgets/${row.id}`)}
          aria-label="Row actions"
        >
          <Eye className="h-4 w-4 mr-1" />
          View Details
        </Button>
      ),
    },
  ];

  return (
    <div className="space-y-4">
      {/* Filter Bar */}
      <FilterBar
        hasActiveFilters={hasActiveFilters}
        onReset={clearFilters}
        search={
          <Input
            placeholder="Search budgets..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="h-8 w-full sm:w-56"
          />
        }
        filters={
          <>
            <Select value={yearFilter} onValueChange={setYearFilter}>
              <SelectTrigger className="h-8 w-full sm:w-32">
                <SelectValue placeholder="Fiscal Year" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All Years</SelectItem>
                {fiscalYears.map((fy) => (
                  <SelectItem key={fy} value={String(fy)}>
                    FY {fy}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <Select value={departmentFilter} onValueChange={setDepartmentFilter}>
              <SelectTrigger className="h-8 w-full sm:w-40">
                <SelectValue placeholder="Department" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All Departments</SelectItem>
                {departments.map((dept) => (
                  <SelectItem key={dept} value={dept}>
                    {dept}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <Select value={statusFilter} onValueChange={setStatusFilter}>
              <SelectTrigger className="h-8 w-full sm:w-36">
                <SelectValue placeholder="Status" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All Statuses</SelectItem>
                <SelectItem value="draft">Draft</SelectItem>
                <SelectItem value="pending">Pending</SelectItem>
                <SelectItem value="approved">Approved</SelectItem>
                <SelectItem value="rejected">Rejected</SelectItem>
                <SelectItem value="completed">Completed</SelectItem>
                <SelectItem value="cancelled">Cancelled</SelectItem>
              </SelectContent>
            </Select>
          </>
        }
      />

      {/* Data table */}
      <DataList
        rows={pagedBudgets}
        columns={columns}
        getRowId={(row) => row.id}
        isLoading={isLoading}
        onRowClick={(row) => router.push(`/budgets/${row.id}`)}
        emptyMessage={
          <Empty className="border-0">
            <EmptyContent>
              <EmptyHeader>
                <EmptyMedia variant="icon">
                  <FolderOpen />
                </EmptyMedia>
                <EmptyTitle>No budgets found</EmptyTitle>
                <EmptyDescription>
                  {hasActiveFilters
                    ? "No budgets match your current filters. Try clearing them."
                    : "You haven't created any budgets yet. Create your first budget to get started."}
                </EmptyDescription>
              </EmptyHeader>
            </EmptyContent>
          </Empty>
        }
        mobileCard={(row) => {
          const used = row.allocatedAmount ?? 0;
          const total = row.totalBudget ?? 0;
          const pct = total > 0 ? Math.round((used / total) * 100) : 0;
          const status = row.status;
          const stage = row.approvalStage;
          const stageLabel =
            status?.toUpperCase() === "DRAFT"
              ? "Not submitted"
              : status?.toUpperCase() === "APPROVED"
                ? "Completed"
                : status?.toUpperCase() === "REJECTED"
                  ? "Rejected"
                  : stage > 0
                    ? `Stage ${stage}`
                    : "Pending";

          return (
            <div className="flex flex-col gap-2">
              <div className="flex items-start justify-between gap-2">
                <div className="min-w-0">
                  <div className="font-semibold text-primary line-clamp-1">
                    {row.budgetCode}
                  </div>
                  <div className="text-xs text-muted-foreground line-clamp-1">
                    {row.name}
                  </div>
                  <div className="text-xs text-muted-foreground line-clamp-1">
                    {row.department ?? "—"} · FY {row.fiscalYear}
                  </div>
                </div>
                <StatusBadge status={row.status} type="document" />
              </div>
              <div className="flex flex-wrap items-center gap-2 text-xs text-muted-foreground">
                <span>K{(total).toLocaleString()}</span>
                <span>·</span>
                <span className="tabular-nums">{pct}% used</span>
                <span>·</span>
                <span>{stageLabel}</span>
              </div>
              {/* Progress bar */}
              <div className="h-1.5 w-full rounded-full bg-muted overflow-hidden">
                <div
                  className="h-full bg-primary transition-all"
                  style={{ width: `${Math.min(pct, 100)}%` }}
                />
              </div>
              <div className="pt-1">
                <Button
                  size="sm"
                  variant="outline"
                  onClick={() => router.push(`/budgets/${row.id}`)}
                >
                  <Eye className="h-4 w-4 mr-1" />
                  View Details
                </Button>
              </div>
            </div>
          );
        }}
      />

      {/* Pagination */}
      {totalRows > PAGE_SIZE && (
        <CustomPagination
          pagination={paginationData}
          updatePagination={(newPagination) => {
            if (newPagination.page) setPage(newPagination.page);
          }}
          showDetails
        />
      )}
    </div>
  );
}
