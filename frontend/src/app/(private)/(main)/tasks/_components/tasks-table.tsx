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
import { ArrowUpDown, Clock } from "lucide-react";

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
import { CustomPagination } from "@/components/ui/custom-pagination";
import {
  useApprovalTasks,
  useApproveTask,
  useRejectTask,
  useClaimTask,
} from "@/hooks/use-approval-workflow";
import { WorkflowTask } from "@/types";
import { toast } from "sonner";
import { useSession } from "@/hooks/use-session";
import { capitalize } from "@/lib/utils";
import { WorkflowActionButtons } from "@/components/workflows/workflow-action-buttons";

interface TasksTableProps {
  refreshTrigger: number;
  status?: "pending" | "in_progress";
}

export function TasksTable({ refreshTrigger, status }: TasksTableProps) {
  const router = useRouter();
  const { user } = useSession();
  const [sorting, setSorting] = React.useState<SortingState>([]);
  const [columnFilters, setColumnFilters] = React.useState<ColumnFiltersState>(
    []
  );
  const [columnVisibility, setColumnVisibility] =
    React.useState<VisibilityState>({});
  const [pagination, setPagination] = React.useState({
    page: 1,
    page_size: 10,
  });

  // Use approval tasks hook with role-based filtering
  const {
    data: tasks = [],
    isLoading,
    error,
    refetch,
  } = useApprovalTasks(
    {
      status: status === "pending" ? "PENDING" : undefined,
      assignedToMe: false, // Show all tasks for now - can be made configurable later
    },
    pagination.page,
    pagination.page_size
  );

  // Debug logging
  console.log("TasksTable Debug:", {
    tasks,
    isLoading,
    error,
    status,
    pagination,
    tasksLength: tasks?.length,
  });

  // Refetch when refreshTrigger changes
  useEffect(() => {
    refetch();
  }, [refreshTrigger, refetch]);

  // Task action handlers
  const handleClaimTask = async (taskId: string) => {
    try {
      // This would use the claim mutation
      toast.success("Task claimed successfully");
      refetch();
    } catch (error) {
      toast.error("Failed to claim task");
    }
  };

  const handleApproveTask = async (taskId: string) => {
    try {
      // This would use the approve mutation
      toast.success("Task approved successfully");
      refetch();
    } catch (error) {
      toast.error("Failed to approve task");
    }
  };

  const handleRejectTask = async (taskId: string) => {
    try {
      // This would use the reject mutation
      toast.success("Task rejected successfully");
      refetch();
    } catch (error) {
      toast.error("Failed to reject task");
    }
  };

  const canUserActOnTask = (task: WorkflowTask) => {
    if (!user) return false;

    // Admin can act on any task
    if (user.role === "admin") return true;

    // User can act on tasks assigned to them or their role
    return (
      task.assignedTo === user.id ||
      task.assignedUserId === user.id ||
      task.claimedBy === user.id ||
      task.assignedRole === user.role
    );
  };

  const isTaskAssignedToUser = (task: WorkflowTask) => {
    if (!user) return false;
    return (
      task.assignedTo === user.id ||
      task.assignedUserId === user.id ||
      task.claimedBy === user.id
    );
  };

  const getTaskTypeLabel = (type: string) => {
    const labels: Record<string, string> = {
      BUDGET_APPROVAL: "Budget Approval",
      REQUISITION_APPROVAL: "Requisition Approval",
      PURCHASE_ORDER_APPROVAL: "PO Approval",
      PAYMENT_VOUCHER_APPROVAL: "Payment Approval",
      GOODS_RECEIVED_NOTE_CONFIRMATION: "GRN Confirmation",
      GOODS_RECEIVED_NOTE_APPROVAL: "GRN Confirmation",
    };
    return (
      labels[type] ||
      type?.replace(/_/g, " ").replace(/\b\w/g, (l) => l.toUpperCase()) ||
      "Approval"
    );
  };

  const getPriorityColor = (priority: string) => {
    const colors: Record<string, string> = {
      URGENT: "bg-red-100 text-red-800",
      HIGH: "bg-orange-100 text-orange-800",
      MEDIUM: "bg-yellow-100 text-yellow-800",
      LOW: "bg-blue-100 text-blue-800",
    };
    return colors[priority] || "bg-gray-100 text-gray-800";
  };

  const columns: ColumnDef<WorkflowTask>[] = [
    {
      accessorKey: "title",
      header: ({ column }) => (
        <Button
          variant="ghost"
          onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
          className="-ml-3"
        >
          Task Title
          <ArrowUpDown className="ml-2 h-4 w-4" />
        </Button>
      ),
      cell: ({ row }) => (
        <div className="font-semibold max-w-xs capitalize">
          {row.original.title ||
            `${capitalize(row.original.entityType || row.original.documentType || "").replaceAll("_", " ")} Requires Approval`}
        </div>
      ),
    },
    {
      accessorKey: "documentNumber",
      header: "Document",
      cell: ({ row }) => (
        <div className="text-sm font-medium">
          {row.original.documentNumber ||
            `${row.original.entityType || row.original.documentType}-${(row.original.entityId || row.original.documentId || "").slice(-3)}`}
        </div>
      ),
    },
    {
      accessorKey: "taskType",
      header: "Type",
      cell: ({ row }) => (
        <div className="text-sm">
          {getTaskTypeLabel(
            row.original.taskType ||
              (
                row.original.entityType ||
                row.original.documentType ||
                ""
              )?.toUpperCase() + "_APPROVAL"
          )}
        </div>
      ),
    },
    {
      accessorKey: "priority",
      header: "Priority",
      cell: ({ row }) => (
        <span
          className={`px-2 py-1 rounded text-xs font-medium ${getPriorityColor(row.original.priority || "MEDIUM")}`}
        >
          {row.original.priority || "MEDIUM"}
        </span>
      ),
    },
    {
      accessorKey: "status",
      header: "Status",
      cell: ({ row }) => (
        <StatusBadge status={row.getValue("status")} type="execution" />
      ),
    },
    {
      accessorKey: "dueAt",
      header: ({ column }) => (
        <Button
          variant="ghost"
          onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
          className="-ml-3"
        >
          Due Date
          <ArrowUpDown className="ml-2 h-4 w-4" />
        </Button>
      ),
      cell: ({ row }) => {
        const dueDate = row.original.dueAt || row.original.dueDate;
        if (!dueDate) return <div>-</div>;

        const dueDateObj = new Date(dueDate);
        const now = new Date();
        const isOverdue =
          dueDateObj < now && row.original.status !== "approved";
        return (
          <div className={isOverdue ? "text-red-600 font-semibold" : ""}>
            {dueDateObj.toLocaleDateString()}
            {isOverdue && <span className="ml-2 text-xs">Overdue</span>}
          </div>
        );
      },
    },
    {
      id: "actions",
      header: "Actions",
      cell: ({ row }) => {
        const task = row.original;

        return (
          <WorkflowActionButtons
            task={task}
            onClaim={handleClaimTask}
            onApprove={handleApproveTask}
            onReject={handleRejectTask}
            onRefresh={refetch}
            variant="table"
            showViewButton={false}
            onView={(task) => {
              const docType = (
                task.entityType ||
                task.documentType ||
                ""
              ).toLowerCase();
              const docId = task.entityId || task.documentId;
              const routes: Record<string, string> = {
                requisition: `/requisitions/${docId}`,
                purchase_order: `/purchase-orders/${docId}`,
                payment_voucher: `/payment-vouchers/${docId}`,
                goods_received_note: `/grn/${docId}`,
                budget: `/budgets/${docId}`,
              };
              const url = routes[docType || ""] || `/tasks/${task.id}`;
              router.push(url);
            }}
          />
        );
      },
    },
  ];

  const table = useReactTable({
    data: tasks,
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
                  {isLoading ? "Loading tasks..." : "No tasks found."}
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>

      {/* Pagination */}
      <CustomPagination
        pagination={{
          page: pagination.page,
          limit: pagination.page_size,
          total: tasks.length,
          totalPages: Math.ceil(tasks.length / pagination.page_size),
          hasNext:
            pagination.page < Math.ceil(tasks.length / pagination.page_size),
          hasPrev: pagination.page > 1,
          page_size: pagination.page_size,
          totalCount: tasks.length,
          total_pages: Math.ceil(tasks.length / pagination.page_size),
          has_next:
            pagination.page < Math.ceil(tasks.length / pagination.page_size),
          has_prev: pagination.page > 1,
        }}
        updatePagination={(newPagination: {
          page: number;
          page_size?: number;
        }) => {
          setPagination((prev: { page: number; page_size: number }) => ({
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
