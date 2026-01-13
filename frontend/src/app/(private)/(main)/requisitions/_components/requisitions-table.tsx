"use client";

import { useCallback, useMemo, useEffect } from "react";
import { useRouter } from "next/navigation";
import { ColumnDef } from "@tanstack/react-table";
import {
  ArrowUpDown,
  Eye,
  Pencil,
  CheckCircle2,
  XCircle,
  MoreVertical,
  Send,
} from "lucide-react";

import { StatusBadge } from "@/components/status-badge";
import { Button } from "@/components/ui/button";
import { DataTable } from "@/components/ui/data-table";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { Requisition } from "@/types/requisition";
import {
  useRequisitions,
  useSubmitRequisitionForApproval,
} from "@/hooks/use-requisition-queries";
import { useApprovalWorkflowStatus } from "@/hooks/use-approval-history";
import type { ActionButton } from "@/components/ui/action-buttons";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { toast } from "sonner";

interface RequisitionsTableProps {
  userId: string;
  userRole: string;
  refreshTrigger: number;
  onEditRequisition: (requisition: Requisition) => void; // Add edit callback
}

const columns: ColumnDef<Requisition>[] = [
  {
    accessorKey: "documentNumber",
    header: ({ column }) => (
      <Button
        variant="ghost"
        onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
        className="-ml-3"
      >
        Document Number
        <ArrowUpDown className="ml-2 h-4 w-4" />
      </Button>
    ),
    cell: ({ row }) => (
      <div className="font-semibold">
        {row.original.documentNumber || row.original.id}
      </div>
    ),
  },
  {
    accessorKey: "title",
    header: "Title",
    cell: ({ row }) => (
      <Tooltip>
        <TooltipTrigger asChild>
          <div className="max-w-[200px] truncate font-medium cursor-help">
            {row.original.title || "-"}
          </div>
        </TooltipTrigger>
        <TooltipContent>
          <p className="max-w-xs">{row.original.title || "No title"}</p>
          {row.original.description && (
            <p className="text-xs text-muted-foreground mt-1 max-w-xs">
              {row.original.description.substring(0, 100)}
              {row.original.description.length > 100 ? "..." : ""}
            </p>
          )}
        </TooltipContent>
      </Tooltip>
    ),
  },
  // {
  //   accessorKey: 'requesterName',
  //   header: 'Requested By',
  //   cell: ({ row }) => (
  //     <div>{row.original.requesterName || '-'}</div>
  //   ),
  // },
  // {
  //   accessorKey: 'requestedFor',
  //   header: 'Requested For',
  //   cell: ({ row }) => (
  //     <div className="text-sm text-muted-foreground">
  //       {row.original.requestedFor || '-'}
  //     </div>
  //   ),
  // },
  {
    accessorKey: "department",
    header: "Department",
    cell: ({ row }) => <div>{row.original.department || "-"}</div>,
  },
  {
    accessorKey: "priority",
    header: "Priority",
    cell: ({ row }) => {
      const priority = row.original.priority?.toLowerCase();
      const priorityColors = {
        urgent: "bg-red-100 text-red-800 border-red-200",
        high: "bg-orange-100 text-orange-800 border-orange-200",
        medium: "bg-blue-100 text-blue-800 border-blue-200",
        low: "bg-gray-100 text-gray-800 border-gray-200",
      };

      return (
        <span
          className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium border ${
            priorityColors[priority as keyof typeof priorityColors] ||
            priorityColors.medium
          }`}
        >
          {row.original.priority || "Medium"}
        </span>
      );
    },
  },
  {
    id: "itemsCount",
    header: "Items",
    cell: ({ row }) => {
      const itemsCount = row.original.items?.length || 0;
      return (
        <div className="text-center">
          <span className="inline-flex items-center justify-center w-6 h-6 text-xs font-medium bg-gray-100 rounded-full">
            {itemsCount}
          </span>
        </div>
      );
    },
  },
  {
    id: "totalAmount",
    header: ({ column }) => (
      <Button
        variant="ghost"
        onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
        className="-ml-3"
      >
        Total Amount
        <ArrowUpDown className="ml-2 h-4 w-4" />
      </Button>
    ),
    cell: ({ row }) => {
      const amount = row.original.totalAmount;
      return (
        <div className="font-medium">
          {amount
            ? `${row.original.currency} ${amount.toLocaleString("en-ZM", {
                minimumFractionDigits: 2,
                maximumFractionDigits: 2,
              })}`
            : "-"}
        </div>
      );
    },
  },
  // {
  //   accessorKey: 'budgetCode',
  //   header: 'Budget Code',
  //   cell: ({ row }) => (
  //     <div className="text-sm font-mono">
  //       {row.original.budgetCode || '-'}
  //     </div>
  //   ),
  // },
  {
    accessorKey: "requiredByDate",
    header: ({ column }) => (
      <Button
        variant="ghost"
        onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
        className="-ml-3"
      >
        Required By
        <ArrowUpDown className="ml-2 h-4 w-4" />
      </Button>
    ),
    cell: ({ row }) => {
      if (!row.original.requiredByDate)
        return <div className="text-muted-foreground">-</div>;

      const date = new Date(row.original.requiredByDate);
      const now = new Date();
      const isOverdue = date < now && row.original.status !== "completed";
      const isUrgent = date.getTime() - now.getTime() < 7 * 24 * 60 * 60 * 1000; // Within 7 days

      return (
        <div
          className={`text-sm ${
            isOverdue
              ? "text-red-600 font-medium"
              : isUrgent
                ? "text-orange-600"
                : "text-muted-foreground"
          }`}
        >
          {date.toLocaleDateString()}
          {isOverdue && <span className="ml-1 text-xs">(Overdue)</span>}
        </div>
      );
    },
  },
  {
    accessorKey: "status",
    header: "Status",
    cell: ({ row }) => (
      <StatusBadge status={row.original.status} type="document" />
    ),
  },
  {
    accessorKey: "createdAt",
    header: ({ column }) => (
      <Button
        variant="ghost"
        onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
        className="-ml-3"
      >
        Date Created
        <ArrowUpDown className="ml-2 h-4 w-4" />
      </Button>
    ),
    cell: ({ row }) => {
      const date = new Date(row.original.createdAt);
      return (
        <div className="text-sm text-muted-foreground">
          {date.toLocaleDateString()}
        </div>
      );
    },
  },
];

// Options dropdown component
function ReqOptionsMenu({
  req,
  router,
  onEditRequisition,
  userId,
  userRole,
  onRefresh,
}: {
  req: Requisition;
  router: ReturnType<typeof useRouter>;
  onEditRequisition: (requisition: Requisition) => void;
  userId: string;
  userRole: string;
  onRefresh: () => void;
}) {
  const submitMutation = useSubmitRequisitionForApproval(req.id, onRefresh);
  const { data: workflowStatus } = useApprovalWorkflowStatus(req.id);

  const handleSubmitForApproval = async () => {
    try {
      await submitMutation.mutateAsync({
        submittedBy: userId,
        submittedByName: req.requesterName || "User",
        submittedByRole: userRole,
        comments: `Submitted for approval on ${new Date().toLocaleDateString()}`,
      });
    } catch (error) {
      console.error("Submit error:", error);
    }
  };

  const canSubmit = req.status === "draft" && req.requesterId === userId;
  const canEdit = req.status === "draft" && req.requesterId === userId;
  const canApprove = workflowStatus?.canApprove && req.status === "pending";
  const canReject = workflowStatus?.canReject && req.status === "pending";

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant={"outline"}>
          <MoreVertical className="h-4 w-4" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="w-48">
        <DropdownMenuItem
          onClick={() => router.push(`/requisitions/${req.id}`)}
        >
          <Eye className="mr-2 h-4 w-4" />
          View Details
        </DropdownMenuItem>

        {canEdit && (
          <DropdownMenuItem onClick={() => onEditRequisition(req)}>
            <Pencil className="mr-2 h-4 w-4" />
            Edit Requisition
          </DropdownMenuItem>
        )}

        {canSubmit && (
          <DropdownMenuItem
            onClick={handleSubmitForApproval}
            disabled={submitMutation.isPending}
          >
            <Send className="mr-2 h-4 w-4 text-blue-600" />
            {submitMutation.isPending ? "Submitting..." : "Submit for Approval"}
          </DropdownMenuItem>
        )}

        {canApprove && (
          <DropdownMenuItem
            onClick={() => router.push(`/requisitions/${req.id}?tab=approvals`)}
          >
            <CheckCircle2 className="mr-2 h-4 w-4 text-green-600" />
            Approve
          </DropdownMenuItem>
        )}

        {canReject && (
          <DropdownMenuItem
            onClick={() => router.push(`/requisitions/${req.id}?tab=approvals`)}
          >
            <XCircle className="mr-2 h-4 w-4 text-red-600" />
            Reject
          </DropdownMenuItem>
        )}

        {req.status === "draft" && req.requesterId === userId && (
          <DropdownMenuItem
            onClick={() => console.log("Delete requisition:", req.id)}
            className="text-red-600 focus:text-red-600"
          >
            <XCircle className="mr-2 h-4 w-4" />
            Delete
          </DropdownMenuItem>
        )}

        {/* Show additional info */}
        {req.categoryName && (
          <div className="px-2 py-1.5 text-xs text-muted-foreground border-t">
            Category: {req.categoryName}
          </div>
        )}
        {req.otherCategoryText && (
          <div className="px-2 py-1.5 text-xs text-muted-foreground">
            Custom: {req.otherCategoryText}
          </div>
        )}

        {/* Show workflow status */}
        {workflowStatus && (
          <div className="px-2 py-1.5 text-xs text-muted-foreground border-t">
            Workflow: Stage {workflowStatus.currentStage}/
            {workflowStatus.totalStages}
          </div>
        )}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

export function RequisitionsTable({
  userId,
  userRole,
  refreshTrigger,
  onEditRequisition,
}: RequisitionsTableProps) {
  const router = useRouter();
  const { data: requisitions = [], refetch } = useRequisitions(1, 50); // Get first 50 requisitions

  // Refetch when refreshTrigger changes
  useEffect(() => {
    refetch();
  }, [refreshTrigger, refetch]);

  // Memoize the data to prevent unnecessary re-renders
  // React Query returns a new array reference on each render,
  // so we memoize based on the actual content changes
  const data = useMemo(() => {
    if (requisitions && requisitions.length > 0) {
      return requisitions;
    }
    return [];
  }, [requisitions]);

  // const getActions = useCallback(
  //   (req: Requisition): ActionButton[] => {
  //     const actions: ActionButton[] = [
  //       {
  //         icon: <Eye className="h-3.5 w-3.5" />,
  //         label: 'View',
  //         tooltip: 'View Details',
  //         onClick: () => router.push(`/requisitions/${req.id}`),
  //       },
  //     ];

  //     // Only allow edit and delete for draft status
  //     if (req.status === 'draft') {
  //       actions.push(
  //         {
  //           icon: <Pencil className="h-3.5 w-3.5" />,
  //           label: 'Edit',
  //           tooltip: 'Edit Requisition',
  //           onClick: () => onEditRequisition(req), // Use callback instead of navigation
  //         }
  //       );
  //     }

  //     return actions;
  //   },
  //   [router, onEditRequisition]
  // );

  return (
    <TooltipProvider>
      <DataTable
        columns={columns}
        data={data}
        searchKey="title"
        searchPlaceholder="Search by title, document number, or requester..."
        // actions={getActions}
        renderRowActions={(req: Requisition) => (
          <ReqOptionsMenu
            req={req}
            router={router}
            onEditRequisition={onEditRequisition}
            userId={userId}
            userRole={userRole}
            onRefresh={refetch}
          />
        )}
      />
    </TooltipProvider>
  );
}
