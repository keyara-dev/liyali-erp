"use client";

import { useCallback, useMemo, useEffect, useState } from "react";
import * as React from "react";
import { useRouter } from "next/navigation";
import {
  Eye,
  Pencil,
  CheckCircle2,
  XCircle,
  MoreVertical,
  Send,
  PlusCircle,
  FileText,
  Undo2,
  Search,
} from "lucide-react";

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
import { DataList, DataListColumn } from "@/components/ui/data-list";
import { FilterBar } from "@/components/ui/filter-bar";
import { PriorityBadge } from "@/components/ui/priority-badge";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { Requisition } from "@/types/requisition";
import { useRequisitions } from "@/hooks/use-requisition-queries";
import { useWithdrawRequisition } from "@/hooks/use-requisition-mutations";
import { useApprovalWorkflowStatus } from "@/hooks/use-approval-history";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { ConfirmationModal } from "@/components/modals/confirmation-modal";
import { formatCurrency } from "@/lib/utils";
import { useDebounce } from "@/hooks/use-debounce";

import { RequisitionFilters } from "./requisitions-filters";

interface RequisitionsTableProps {
  userId: string;
  userRole: string;
  refreshTrigger: number;
  onEditRequisition: (requisition: Requisition) => void;
  onCreateRequisition: () => void;
  filters?: RequisitionFilters;
  initialData?: Requisition[];
}

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
  const [showWithdrawModal, setShowWithdrawModal] = useState(false);

  const withdrawMutation = useWithdrawRequisition(onRefresh);
  const { data: workflowStatus } = useApprovalWorkflowStatus(req.id);

  const handleWithdraw = async () => {
    try {
      await withdrawMutation.mutateAsync(req.id);
      setShowWithdrawModal(false);
    } catch (error) {
      console.error("Withdraw error:", error);
    }
  };

  const reqStatus = req.status?.toUpperCase();
  const canSubmit = reqStatus === "DRAFT" && req.requesterId === userId;
  const canWithdraw = reqStatus === "PENDING" && req.requesterId === userId;
  const canEdit = reqStatus === "DRAFT" && req.requesterId === userId;
  const canApprove = workflowStatus?.canApprove && reqStatus === "PENDING";
  const canReject = workflowStatus?.canReject && reqStatus === "PENDING";

  return (
    <>
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
              onClick={() => router.push(`/requisitions/${req.id}`)}
            >
              <Send className="mr-2 h-4 w-4 text-blue-600" />
              Submit for Approval
            </DropdownMenuItem>
          )}

          {canWithdraw && (
            <DropdownMenuItem
              onClick={() => setShowWithdrawModal(true)}
              className="text-amber-600 focus:text-amber-600"
            >
              <Undo2 className="mr-2 h-4 w-4" />
              Withdraw
            </DropdownMenuItem>
          )}

          {canApprove && (
            <DropdownMenuItem
              onClick={() =>
                router.push(`/requisitions/${req.id}?tab=approvals`)
              }
            >
              <CheckCircle2 className="mr-2 h-4 w-4 text-green-600" />
              Approve
            </DropdownMenuItem>
          )}

          {canReject && (
            <DropdownMenuItem
              onClick={() =>
                router.push(`/requisitions/${req.id}?tab=approvals`)
              }
            >
              <XCircle className="mr-2 h-4 w-4 text-red-600" />
              Reject
            </DropdownMenuItem>
          )}

          {reqStatus === "DRAFT" && req.requesterId === userId && (
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

      {/* Withdraw Confirmation Modal */}
      <ConfirmationModal
        open={showWithdrawModal}
        onOpenChange={setShowWithdrawModal}
        onConfirm={handleWithdraw}
        type="withdraw"
        title="Withdraw Requisition"
        description={`Are you sure you want to withdraw requisition ${req.documentNumber || req.id}? It will be reverted to draft status and you can edit and re-submit it later.`}
        isLoading={withdrawMutation.isPending}
      />
    </>
  );
}

export function RequisitionsTable({
  userId,
  userRole,
  refreshTrigger,
  onEditRequisition,
  onCreateRequisition,
  filters = {},
  initialData,
}: RequisitionsTableProps) {
  const router = useRouter();

  // Local filter state — managed inside this component so FilterBar is self-contained
  const [searchQuery, setSearchQuery] = useState(filters.searchTerm || "");
  const [statusFilter, setStatusFilter] = useState(filters.status || "all");
  const [priorityFilter, setPriorityFilter] = useState(filters.priority || "all");
  // Department filter deferred — API accepts it via the server-side filters prop
  const debouncedSearch = useDebounce(searchQuery, 400);

  const {
    data: requisitions = [],
    isLoading,
    refetch,
  } = useRequisitions(
    1,
    100,
    filters.status
      ? { status: filters.status, department: filters.department }
      : undefined,
    initialData,
  );

  // Refetch when refreshTrigger changes
  useEffect(() => {
    refetch();
  }, [refreshTrigger, refetch]);

  // Client-side filtering (search + priority; status/department come from parent filters prop)
  const filteredData = useMemo(() => {
    if (!requisitions || requisitions.length === 0) return [];

    let filtered = [...requisitions];

    // Search filter (debounced)
    if (debouncedSearch) {
      const s = debouncedSearch.toLowerCase();
      filtered = filtered.filter(
        (req) =>
          req.documentNumber?.toLowerCase().includes(s) ||
          req.title?.toLowerCase().includes(s) ||
          req.requesterName?.toLowerCase().includes(s),
      );
    }

    // Status filter (local override — works on top of any server-side status)
    if (statusFilter !== "all") {
      filtered = filtered.filter(
        (req) => req.status?.toLowerCase() === statusFilter.toLowerCase(),
      );
    }

    // Priority filter
    if (priorityFilter !== "all") {
      filtered = filtered.filter(
        (req) => req.priority?.toLowerCase() === priorityFilter.toLowerCase(),
      );
    }

    // Date range from parent filters prop
    if (filters.startDate) {
      filtered = filtered.filter(
        (req) => new Date(req.createdAt) >= filters.startDate!,
      );
    }
    if (filters.endDate) {
      filtered = filtered.filter(
        (req) => new Date(req.createdAt) <= filters.endDate!,
      );
    }

    return filtered;
  }, [requisitions, debouncedSearch, statusFilter, priorityFilter, filters.startDate, filters.endDate]);

  const hasActiveFilters =
    Boolean(searchQuery) || statusFilter !== "all" || priorityFilter !== "all";

  const clearFilters = useCallback(() => {
    setSearchQuery("");
    setStatusFilter("all");
    setPriorityFilter("all");
  }, []);

  const columns: DataListColumn<Requisition>[] = [
    {
      id: "documentNumber",
      header: "Document",
      priority: "always",
      cell: (row) => (
        <div className="font-semibold uppercase">
          {row.documentNumber || row.id}
        </div>
      ),
    },
    {
      id: "title",
      header: "Title",
      priority: "lg",
      cell: (row) => (
        <Tooltip>
          <TooltipTrigger asChild>
            <span className="max-w-50 truncate capitalize font-medium cursor-help line-clamp-1">
              {row.title || "-"}
            </span>
          </TooltipTrigger>
          <TooltipContent>
            <p className="max-w-xs">{row.title || "No title"}</p>
            {row.description && (
              <p className="text-xs text-muted-foreground mt-1 max-w-xs">
                {row.description.substring(0, 100)}
                {row.description.length > 100 ? "..." : ""}
              </p>
            )}
          </TooltipContent>
        </Tooltip>
      ),
    },
    {
      id: "requesterName",
      header: "Requester",
      priority: "md",
      cell: (row) => (
        <span className="font-medium capitalize">{row.requesterName || "-"}</span>
      ),
    },
    {
      id: "priority",
      header: "Priority",
      priority: "md",
      cell: (row) => <PriorityBadge priority={row.priority} />,
    },
    {
      id: "totalAmount",
      header: "Amount",
      priority: "md",
      align: "right",
      cell: (row) => (
        <div className="font-medium">
          {row.totalAmount
            ? formatCurrency(row.totalAmount, row.currency)
            : "-"}
        </div>
      ),
    },
    {
      id: "status",
      header: "Status",
      cell: (row) => <StatusBadge status={row.status} type="document" />,
    },
    {
      id: "createdAt",
      header: "Created",
      priority: "lg",
      cell: (row) => {
        const createdDate = new Date(row.createdAt);
        const requiredByDate = row.requiredByDate
          ? new Date(row.requiredByDate)
          : null;
        const now = new Date();
        const isOverdue =
          requiredByDate &&
          requiredByDate < now &&
          row.status?.toUpperCase() !== "COMPLETED";
        const isUrgent =
          requiredByDate &&
          requiredByDate.getTime() - now.getTime() < 7 * 24 * 60 * 60 * 1000;

        return (
          <div className="space-y-0.5">
            <div className="text-sm text-muted-foreground">
              {createdDate.toLocaleDateString()}
            </div>
            {requiredByDate && (
              <div
                className={`text-xs ${
                  isOverdue
                    ? "text-red-600 font-medium"
                    : isUrgent
                      ? "text-orange-600"
                      : "text-muted-foreground/70"
                }`}
              >
                Due: {requiredByDate.toLocaleDateString()}
                {isOverdue && <span className="ml-1">(Overdue)</span>}
              </div>
            )}
          </div>
        );
      },
    },
    {
      id: "actions",
      header: <span className="sr-only">Actions</span>,
      cell: (row) => (
        <ReqOptionsMenu
          req={row}
          router={router}
          onEditRequisition={onEditRequisition}
          userId={userId}
          userRole={userRole}
          onRefresh={refetch}
        />
      ),
    },
  ];

  return (
    <div className="space-y-4">
      <FilterBar
        search={
          <div className="relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
            <Input
              placeholder="Search document number, title, requester…"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="pl-10"
            />
          </div>
        }
        filters={
          <>
            <Select value={statusFilter} onValueChange={setStatusFilter}>
              <SelectTrigger className="w-38">
                <SelectValue placeholder="Status" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All statuses</SelectItem>
                <SelectItem value="draft">Draft</SelectItem>
                <SelectItem value="submitted">Submitted</SelectItem>
                <SelectItem value="in_approval">In Approval</SelectItem>
                <SelectItem value="approved">Approved</SelectItem>
                <SelectItem value="rejected">Rejected</SelectItem>
              </SelectContent>
            </Select>

            <Select value={priorityFilter} onValueChange={setPriorityFilter}>
              <SelectTrigger className="w-32.5">
                <SelectValue placeholder="Priority" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All priorities</SelectItem>
                <SelectItem value="urgent">Urgent</SelectItem>
                <SelectItem value="high">High</SelectItem>
                <SelectItem value="medium">Medium</SelectItem>
                <SelectItem value="low">Low</SelectItem>
              </SelectContent>
            </Select>
            {/* Department filter deferred — use parent RequisitionFilters for dept filtering */}
          </>
        }
        hasActiveFilters={hasActiveFilters}
        onReset={clearFilters}
        meta={`${filteredData.length} requisition${filteredData.length === 1 ? "" : "s"}${hasActiveFilters ? " (filtered)" : ""}`}
      />

      <DataList<Requisition>
        rows={filteredData}
        columns={columns}
        getRowId={(row) => row.id}
        isLoading={isLoading}
        emptyMessage={
          <div className="flex flex-col items-center gap-3 py-4">
            <FileText className="h-10 w-10 text-muted-foreground" />
            <div>
              <p className="font-medium">No Requisitions Found</p>
              <p className="text-xs text-muted-foreground mt-1">
                {hasActiveFilters || filters.status || filters.department || filters.searchTerm
                  ? "No requisitions match your current filters. Try adjusting your search criteria."
                  : "Get started by creating your first requisition"}
              </p>
            </div>
            {!hasActiveFilters && !filters.status && !filters.department && !filters.searchTerm && (
              <Button onClick={onCreateRequisition} size="sm">
                <PlusCircle className="h-4 w-4 mr-2" />
                Create Requisition
              </Button>
            )}
          </div>
        }
        mobileCard={(row) => (
          <div className="flex flex-col gap-2">
            <div className="flex items-start justify-between gap-2">
              <div className="min-w-0">
                <div className="font-medium text-primary line-clamp-1 uppercase">
                  {row.documentNumber || row.id}
                </div>
                <div className="text-xs text-muted-foreground line-clamp-1">
                  {row.title}
                </div>
              </div>
              <StatusBadge status={row.status} type="document" />
            </div>
            <div className="flex flex-wrap items-center gap-2 text-xs text-muted-foreground">
              <PriorityBadge priority={row.priority} />
              <span>{row.requesterName}</span>
              <span>
                {row.totalAmount
                  ? formatCurrency(row.totalAmount, row.currency)
                  : "-"}
              </span>
              <span>{new Date(row.createdAt).toLocaleDateString()}</span>
            </div>
            <div className="pt-1">
              <ReqOptionsMenu
                req={row}
                router={router}
                onEditRequisition={onEditRequisition}
                userId={userId}
                userRole={userRole}
                onRefresh={refetch}
              />
            </div>
          </div>
        )}
      />
    </div>
  );
}
