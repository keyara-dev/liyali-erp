"use client";

import { useEffect, useState } from "react";
import * as React from "react";
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
import { ArrowUpDown, Eye } from "lucide-react";

import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { StatusBadge } from "@/components/status-badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { CustomPagination } from "@/components/ui/custom-pagination";
import { useBudgetsWithStorage } from "@/hooks/use-budget-storage";
import { Budget } from "@/types/budget";
import { Pagination } from "@/types";

interface BudgetsTableProps {
  userId: string;
  userRole: string;
  refreshTrigger: number;
  onBudgetAction: () => void;
}

export function BudgetsTable({
  userId,
  userRole,
  refreshTrigger,
  onBudgetAction,
}: BudgetsTableProps) {
  const router = useRouter();
  const { data: budgetsFromHook = [], isLoading: hookLoading } =
    useBudgetsWithStorage(true);
  const [budgets, setBudgets] = useState<Budget[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    if (budgetsFromHook && budgetsFromHook.length > 0) {
      // Filter by current user's budgets
      const userBudgets = budgetsFromHook.filter(
        (budget) => budget.createdBy === userId
      );
      setBudgets(userBudgets);
      setIsLoading(false);
    } else {
      setBudgets([]);
      setIsLoading(false);
    }
  }, [budgetsFromHook, userId, refreshTrigger]);
  const [sorting, setSorting] = React.useState<SortingState>([]);
  const [columnFilters, setColumnFilters] = React.useState<ColumnFiltersState>(
    []
  );
  const [columnVisibility, setColumnVisibility] =
    React.useState<VisibilityState>({});
  const [pagination, setPagination] = React.useState<Pagination>({
    page: 1,
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
      accessorKey: "budgetNumber",
      header: ({ column }) => (
        <Button
          variant="ghost"
          onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
          className="-ml-3"
        >
          Budget Number
          <ArrowUpDown className="ml-2 h-4 w-4" />
        </Button>
      ),
      cell: ({ row }) => (
        <div className="font-semibold">{row.getValue("budgetNumber")}</div>
      ),
    },
    {
      accessorKey: "name",
      header: "Budget Name",
      cell: ({ row }) => <div className="max-w-xs">{row.getValue("name")}</div>,
    },
    {
      accessorKey: "department",
      header: "Department",
      cell: ({ row }) => <div>{row.getValue("department")}</div>,
    },
    {
      accessorKey: "totalAmount",
      header: "Total Amount",
      cell: ({ row }) => (
        <div className="font-medium">
          {formatCurrency(row.original.totalAmount, row.original.currency)}
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
      accessorKey: "currentApprovalStage",
      header: "Approval Stage",
      cell: ({ row }) => (
        <div className="text-sm">
          {row.original.currentApprovalStage} /{" "}
          {row.original.totalApprovalStages}
        </div>
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
                            header.getContext()
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
                        cell.getContext()
                      )}
                    </TableCell>
                  ))}
                </TableRow>
              ))
            ) : (
              <TableRow>
                <TableCell
                  colSpan={columns.length}
                  className="h-24 text-center"
                >
                  {isLoading ? "Loading budgets..." : "No budgets found."}
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>

      {/* Pagination */}
      <CustomPagination
        pagination={{
          ...pagination,
          total_pages: Math.ceil(budgets.length / pagination.page_size!),
          totalCount: budgets.length,
          has_next:
            pagination.page < Math.ceil(budgets.length / pagination.page_size!),
          has_prev: pagination.page > 1,
        }}
        updatePagination={(newPagination) => {
          setPagination((prev) => ({
            ...prev,
            page: newPagination.page,
            page_size: newPagination.page_size || prev.page_size,
          }));
          table.setPageIndex(newPagination.page - 1);
          if (newPagination.page_size) {
            table.setPageSize(newPagination.page_size);
          }
        }}
        allowSetPageSize
        showDetails
      />
    </div>
  );
}
