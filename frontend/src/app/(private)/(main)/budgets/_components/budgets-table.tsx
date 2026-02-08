"use client";

import { useCallback, useMemo, useEffect } from "react";
import * as React from "react";
import { useQueryClient } from "@tanstack/react-query";
import {
  ColumnDef,
  ColumnFiltersState,
  SortingState,
  VisibilityState,
  flexRender,
  getCoreRowModel,
  getFilteredRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  useReactTable,
} from "@tanstack/react-table";
import { useRouter } from "next/navigation";
import { ArrowUpDown, Eye, FolderOpen } from "lucide-react";

import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
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
import { CustomPagination } from "@/components/ui/custom-pagination";
import { useBudgets } from "@/hooks/use-budget-queries";
import { Budget } from "@/types/budget";
import { Pagination } from "@/types";
import { QUERY_KEYS } from "@/lib/constants";

interface BudgetsTableProps {
  userRole: string;
  refreshTrigger: number;
  onBudgetAction: () => void;
}

export function BudgetsTable({
  userRole,
  refreshTrigger,
  onBudgetAction,
}: BudgetsTableProps) {
  const router = useRouter();
  const queryClient = useQueryClient();
  const {
    data: budgetsFromHook = [],
    isLoading: hookLoading,
    refetch,
  } = useBudgets(); // Get all budgets for the organization

  // Refetch when refreshTrigger changes (after budget creation)
  useEffect(() => {
    if (refreshTrigger > 0) {
      refetch();
      queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.BUDGETS.ALL] });
    }
  }, [refreshTrigger, refetch, queryClient]);

  const budgets = useMemo(() => {
    if (budgetsFromHook && budgetsFromHook.length > 0) {
      return budgetsFromHook;
    }
    return [];
  }, [budgetsFromHook]);

  const [sorting, setSorting] = React.useState<SortingState>([]);
  const [columnFilters, setColumnFilters] = React.useState<ColumnFiltersState>(
    [],
  );
  const [columnVisibility, setColumnVisibility] =
    React.useState<VisibilityState>({});
  const [pagination, setPagination] = React.useState<Pagination>({
    page: 1,
    limit: 10,
    total: 0,
    totalPages: 1,
    hasNext: false,
    hasPrev: false,
    page_size: 10,
    total_pages: 1,
    totalCount: 0,
    has_next: false,
    has_prev: false,
  });

  const formatCurrency = (amount: number, currency: string = "USD") => {
    return new Intl.NumberFormat("en-US", {
      style: "currency",
      currency: currency,
    }).format(amount);
  };

  const columns: ColumnDef<Budget>[] = [
    {
      accessorKey: "budgetCode",
      header: ({ column }) => (
        <Button
          variant="ghost"
          onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
          className="-ml-3"
        >
          Budget Code
          <ArrowUpDown className="ml-2 h-4 w-4" />
        </Button>
      ),
      cell: ({ row }) => (
        <div className="font-semibold">{row.getValue("budgetCode")}</div>
      ),
    },
    {
      accessorKey: "department",
      header: "Department",
      cell: ({ row }) => <div>{row.getValue("department")}</div>,
    },
    {
      accessorKey: "totalBudget",
      header: "Total Budget",
      cell: ({ row }) => (
        <div className="font-medium">
          K{(row.original.totalBudget || 0).toLocaleString()}
        </div>
      ),
    },
    {
      accessorKey: "allocatedAmount",
      header: "Allocated Amount",
      cell: ({ row }) => (
        <div className="font-medium">
          K{(row.original.allocatedAmount || 0).toLocaleString()}
        </div>
      ),
    },
    {
      accessorKey: "fiscalYear",
      header: "Fiscal Year",
      cell: ({ row }) => <div>{row.getValue("fiscalYear")}</div>,
    },
    {
      accessorKey: "status",
      header: "Status",
      cell: ({ row }) => (
        <StatusBadge status={row.getValue("status")} type="document" />
      ),
    },
    {
      accessorKey: "approvalStage",
      header: "Approval Stage",
      cell: ({ row }) => (
        <div className="text-sm">Stage {row.original.approvalStage}</div>
      ),
    },
    {
      id: "actions",
      cell: ({ row }) => (
        <Button
          size="sm"
          variant="outline"
          onClick={() => router.push(`/budgets/${row.original.id}`)}
        >
          <Eye className="h-4 w-4 mr-1" />
          View Details
        </Button>
      ),
    },
  ];

  const table = useReactTable({
    data: budgets,
    columns,
    onSortingChange: setSorting,
    onColumnFiltersChange: setColumnFilters,
    getCoreRowModel: getCoreRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    onColumnVisibilityChange: setColumnVisibility,
    state: {
      sorting,
      columnFilters,
      columnVisibility,
    },
  });

  const handleUpdatePagination = useCallback(
    (newPagination: { page: number; page_size?: number }) => {
      setPagination((prev) => ({
        ...prev,
        page: newPagination.page,
        page_size: newPagination.page_size || prev.page_size,
      }));
      table.setPageIndex(newPagination.page - 1);
      if (newPagination.page_size) {
        table.setPageSize(newPagination.page_size);
      }
    },
    [table],
  );

  return (
    <div className="space-y-4">
      <div className="rounded-md border">
        <Table>
          <TableHeader>
            {table.getHeaderGroups().map((headerGroup) => (
              <TableRow key={headerGroup.id}>
                {headerGroup.headers.map((header) => {
                  return (
                    <TableHead key={header.id}>
                      {header.isPlaceholder
                        ? null
                        : flexRender(
                            header.column.columnDef.header,
                            header.getContext(),
                          )}
                    </TableHead>
                  );
                })}
              </TableRow>
            ))}
          </TableHeader>
          <TableBody>
            {table.getRowModel().rows?.length ? (
              table.getRowModel().rows.map((row) => (
                <TableRow
                  key={row.id}
                  data-state={row.getIsSelected() && "selected"}
                >
                  {row.getVisibleCells().map((cell) => (
                    <TableCell key={cell.id}>
                      {flexRender(
                        cell.column.columnDef.cell,
                        cell.getContext(),
                      )}
                    </TableCell>
                  ))}
                </TableRow>
              ))
            ) : (
              <TableRow>
                <TableCell colSpan={columns.length} className="h-24 p-0">
                  <Empty className="border-0">
                    <EmptyContent>
                      <EmptyHeader>
                        <EmptyMedia variant="icon">
                          <FolderOpen />
                        </EmptyMedia>
                        <EmptyTitle>
                          {hookLoading
                            ? "Loading budgets..."
                            : "No budgets found"}
                        </EmptyTitle>
                        <EmptyDescription>
                          {hookLoading
                            ? "Please wait while we fetch your budgets."
                            : "You haven't created any budgets yet. Create your first budget to get started."}
                        </EmptyDescription>
                      </EmptyHeader>
                    </EmptyContent>
                  </Empty>
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>

      {/* Pagination */}
      <CustomPagination
        pagination={useMemo(
          () => ({
            ...pagination,
            total_pages: Math.ceil(
              budgets.length / (pagination.page_size || 10),
            ),
            totalCount: budgets.length,
            has_next:
              pagination.page <
              Math.ceil(budgets.length / (pagination.page_size || 10)),
            has_prev: pagination.page > 1,
          }),
          [pagination, budgets.length],
        )}
        updatePagination={handleUpdatePagination}
        allowSetPageSize
        showDetails
      />
    </div>
  );
}
