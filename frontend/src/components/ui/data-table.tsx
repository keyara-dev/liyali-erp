"use client";

import * as React from "react";
import {
  ColumnDef,
  flexRender,
  getCoreRowModel,
  useReactTable,
  getPaginationRowModel,
  SortingState,
  getSortedRowModel,
  ColumnFiltersState,
  getFilteredRowModel,
} from "@tanstack/react-table";

import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Empty,
  EmptyContent,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "./empty";
import { ArrowLeft, ClipboardXIcon, Lightbulb } from "lucide-react";
import { useRouter } from "next/navigation";
import { CustomPagination } from "./custom-pagination";
import { ActionButtons, type ActionButton } from "./action-buttons";
import type { Pagination } from "@/types";

interface DataTableProps<TData, TValue> {
  columns: ColumnDef<TData, TValue>[];
  data: TData[];
  searchKey?: string;
  searchPlaceholder?: string;
  actions?: (row: TData) => ActionButton[];
  renderRowActions?: (row: TData) => React.ReactNode;
  hideSearchBar?: boolean;
}

export function DataTable<TData, TValue>({
  columns,
  data,
  searchKey,
  searchPlaceholder = "Search...",
  actions,
  renderRowActions,
  hideSearchBar = false,
}: DataTableProps<TData, TValue>) {
  const [sorting, setSorting] = React.useState<SortingState>([]);
  const [columnFilters, setColumnFilters] = React.useState<ColumnFiltersState>(
    []
  );

  const router = useRouter();

  // Add actions column if actions or renderRowActions are provided
  const finalColumns = React.useMemo(() => {
    const cols = [...columns];
    if (actions || renderRowActions) {
      cols.push({
        id: "actions",
        cell: ({ row }) => (
          <div className="flex items-center justify-end max-w-max ml-auto gap-2">
            {actions && (
              <ActionButtons actions={actions(row.original)} align="end" />
            )}
            {renderRowActions && renderRowActions(row.original)}
          </div>
        ),
      } as ColumnDef<TData, TValue>);
    }
    return cols;
  }, [columns, actions, renderRowActions]);

  const table = useReactTable({
    data,
    columns: finalColumns,
    getCoreRowModel: getCoreRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    onSortingChange: setSorting,
    getSortedRowModel: getSortedRowModel(),
    onColumnFiltersChange: setColumnFilters,
    getFilteredRowModel: getFilteredRowModel(),
    state: {
      sorting,
      columnFilters,
    },
  });

  return (
    <div className="space-y-4">
      {searchKey && !hideSearchBar && (
        <Input
          placeholder={searchPlaceholder}
          value={(table.getColumn(searchKey)?.getFilterValue() as string) ?? ""}
          onChange={(event) =>
            table.getColumn(searchKey)?.setFilterValue(event.target.value)
          }
          className="max-w-sm"
        />
      )}
      <div className="rounded-lg border">
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
                  {" "}
                  <Empty>
                    <EmptyHeader>
                      <EmptyMedia variant="icon">
                        <ClipboardXIcon />
                      </EmptyMedia>
                      <EmptyTitle>No Results</EmptyTitle>
                      <EmptyDescription>
                        There is nothing to show here yet
                      </EmptyDescription>
                    </EmptyHeader>
                    <EmptyContent>
                      <div className="flex items-center gap-2">
                        <Button
                          onClick={() => {
                            router.back();
                          }}
                        >
                          <ArrowLeft className="h-4 w-4" />
                          Go Back
                        </Button>
                      </div>
                    </EmptyContent>
                  </Empty>
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>
      <CustomPagination
        pagination={{
          page: table.getState().pagination.pageIndex + 1,
          page_size: table.getState().pagination.pageSize,
          total_pages: table.getPageCount(),
          totalCount: data.length,
          has_next: table.getCanNextPage(),
          has_prev: table.getCanPreviousPage(),
        }}
        updatePagination={({ page, page_size }) => {
          if (page_size && page_size !== table.getState().pagination.pageSize) {
            table.setPageSize(page_size);
          }
          table.setPageIndex(page - 1);
        }}
        allowSetPageSize={true}
        showDetails={true}
      />
    </div>
  );
}
