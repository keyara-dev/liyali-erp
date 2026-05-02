"use client";

import { useState } from "react";
import * as React from "react";
import { useQueryClient } from "@tanstack/react-query";
import { useRouter } from "next/navigation";

import { StatusBadge } from "@/components/status-badge";
import { PriorityBadge } from "@/components/ui/priority-badge";
import { Input } from "@/components/ui/input";
import { Search } from "lucide-react";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { CustomPagination } from "@/components/ui/custom-pagination";
import { DataList, DataListColumn } from "@/components/ui/data-list";
import { FilterBar } from "@/components/ui/filter-bar";
import { useApprovalTasks } from "@/hooks/use-approval-workflow";
import { useDebounce } from "@/hooks/use-debounce";
import { capitalize } from "@/lib/utils";
import { QUERY_KEYS } from "@/lib/constants";
import { WorkflowActionButtons } from "@/components/workflows/workflow-action-buttons";
import {
  claimWorkflowTask,
  approveApprovalTask,
  rejectApprovalTask,
  reassignApprovalTask,
} from "@/app/_actions/workflow-approval-actions";

interface WorkflowTask {
  id: string;
  status: string;
  claimedBy?: string;
  assignedRole?: string;
  assignedUserId?: string;
  assignedTo?: string;
  stageNumber?: number;
  stageName?: string;
  claimExpiry?: string;
  entityType?: string;
  entityId?: string;
  documentType?: string;
  documentId?: string;
  documentNumber?: string;
  title?: string;
  taskType?: string;
  priority?: string;
  dueAt?: string;
  dueDate?: string;
}

const TASK_TYPE_LABELS: Record<string, string> = {
  BUDGET_APPROVAL: "Budget Approval",
  REQUISITION_APPROVAL: "Requisition Approval",
  PURCHASE_ORDER_APPROVAL: "PO Approval",
  PAYMENT_VOUCHER_APPROVAL: "Payment Approval",
  GOODS_RECEIVED_NOTE_CONFIRMATION: "GRN Confirmation",
  GOODS_RECEIVED_NOTE_APPROVAL: "GRN Confirmation",
};

function getTaskTypeLabel(type: string) {
  return (
    TASK_TYPE_LABELS[type] ||
    type?.replace(/_/g, " ").replace(/\b\w/g, (l) => l.toUpperCase()) ||
    "Approval"
  );
}

function getDocumentNumber(t: WorkflowTask) {
  return (
    t.documentNumber ||
    `${t.entityType || t.documentType}-${(t.entityId || t.documentId || "").slice(-3)}`
  );
}

function getTitle(t: WorkflowTask) {
  return (
    t.title ||
    `${capitalize(t.entityType || t.documentType || "").replaceAll("_", " ")} Requires Approval`
  );
}

function getDocRoute(task: WorkflowTask): string {
  const docType = (task.entityType || task.documentType || "").toLowerCase();
  const docId = task.entityId || task.documentId;
  const routes: Record<string, string> = {
    requisition: `/requisitions/${docId}`,
    purchase_order: `/purchase-orders/${docId}`,
    payment_voucher: `/payment-vouchers/${docId}`,
    grn: `/grn/${docId}`,
    goods_received_note: `/grn/${docId}`,
    budget: `/budgets/${docId}`,
  };
  return routes[docType] || `/tasks/${task.id}`;
}

function DueCell({ t }: { t: WorkflowTask }) {
  const dueDate = t.dueAt || t.dueDate;
  if (!dueDate) return <span className="text-muted-foreground">—</span>;
  const d = new Date(dueDate);
  const overdue = d < new Date() && t.status?.toUpperCase() !== "APPROVED";
  return (
    <span className={overdue ? "text-rose-600 font-medium" : ""}>
      {d.toLocaleDateString()}
      {overdue && <span className="ml-1.5 text-[10px] uppercase">Overdue</span>}
    </span>
  );
}

export function TasksTable() {
  const router = useRouter();
  const queryClient = useQueryClient();
  const [paginationState, setPaginationState] = React.useState({ page: 1, page_size: 10 });
  const [searchQuery, setSearchQuery] = useState("");
  const [statusFilter, setStatusFilter] = useState<string>("all");
  const [documentTypeFilter, setDocumentTypeFilter] = useState<string>("all");
  const [priorityFilter, setPriorityFilter] = useState<string>("all");
  const debouncedSearch = useDebounce(searchQuery, 500);

  const apiFilters = React.useMemo(
    () => ({
      status: statusFilter !== "all" ? (statusFilter.toUpperCase() as any) : undefined,
      documentType: documentTypeFilter !== "all" ? documentTypeFilter : undefined,
      priority: priorityFilter !== "all" ? priorityFilter : undefined,
      assignedToMe: false,
    }),
    [statusFilter, documentTypeFilter, priorityFilter]
  );

  const { data: approvalData, isLoading } = useApprovalTasks(
    apiFilters,
    paginationState.page,
    paginationState.page_size
  );

  const tasks = (approvalData?.data || []) as WorkflowTask[];
  const paginationMeta = approvalData?.pagination;

  const filteredTasks = React.useMemo(() => {
    if (!debouncedSearch) return tasks;
    const s = debouncedSearch.toLowerCase();
    return tasks.filter(
      (t) =>
        t.title?.toLowerCase().includes(s) ||
        t.documentNumber?.toLowerCase().includes(s) ||
        t.stageName?.toLowerCase().includes(s) ||
        t.entityType?.toLowerCase().includes(s)
    );
  }, [tasks, debouncedSearch]);

  const handleClaimTask = React.useCallback(
    async (taskId: string) => {
      const r = await claimWorkflowTask(taskId);
      if (!r.success) throw new Error(r.message);
      await queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.APPROVALS.ALL] });
    },
    [queryClient]
  );
  const handleApproveTask = React.useCallback(
    async (taskId: string, data?: { signature: string; comments: string }) => {
      const r = await approveApprovalTask(taskId, {
        signature: data?.signature || "",
        comments: data?.comments || "Approved",
        stageNumber: 1,
      });
      if (!r.success) throw new Error(r.message);
      await queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.APPROVALS.ALL] });
    },
    [queryClient]
  );
  const handleRejectTask = React.useCallback(
    async (
      taskId: string,
      data?: {
        signature: string;
        comments: string;
        rejectionType?: "reject" | "return_to_draft" | "return_to_previous_stage";
      }
    ) => {
      const r = await rejectApprovalTask(taskId, {
        signature: data?.signature || "",
        remarks: data?.comments || "Rejected",
        rejectionType: data?.rejectionType || "reject",
      });
      if (!r.success) throw new Error(r.message);
      await queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.APPROVALS.ALL] });
    },
    [queryClient]
  );
  const handleReassignTask = React.useCallback(
    async (taskId: string, newUserId: string, reason: string) => {
      const r = await reassignApprovalTask(taskId, { newApproverId: newUserId, reason });
      if (!r.success) throw new Error(r.message);
      await queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.APPROVALS.ALL] });
    },
    [queryClient]
  );
  const handleRefresh = React.useCallback(
    () => queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.APPROVALS.ALL] }),
    [queryClient]
  );

  const columns: DataListColumn<WorkflowTask>[] = [
    {
      id: "title",
      header: "Task",
      cell: (t) => (
        <div className="flex flex-col">
          <span className="font-medium capitalize line-clamp-1">{getTitle(t)}</span>
          <span className="text-xs text-muted-foreground">{getDocumentNumber(t)}</span>
        </div>
      ),
    },
    {
      id: "stageName",
      header: "Stage",
      priority: "md",
      cell: (t) => <span className="text-sm">{t.stageName || "—"}</span>,
    },
    {
      id: "taskType",
      header: "Type",
      priority: "lg",
      cell: (t) => (
        <span className="text-sm">
          {getTaskTypeLabel(
            t.taskType || ((t.entityType || t.documentType || "") as string).toUpperCase() + "_APPROVAL"
          )}
        </span>
      ),
    },
    { id: "priority", header: "Priority", priority: "md", cell: (t) => <PriorityBadge priority={t.priority} /> },
    {
      id: "status",
      header: "Status",
      cell: (t) => <StatusBadge status={t.status} type="execution" />,
    },
    { id: "dueAt", header: "Due", priority: "md", cell: (t) => <DueCell t={t} /> },
    {
      id: "actions",
      header: <span className="sr-only">Actions</span>,
      cell: (t) => (
        <WorkflowActionButtons
          task={t as any}
          onClaim={handleClaimTask}
          onApprove={handleApproveTask}
          onReject={handleRejectTask}
          onReassign={handleReassignTask}
          onRefresh={handleRefresh}
          variant="table"
          showViewButton={false}
          onView={(task) => router.push(getDocRoute(task))}
        />
      ),
    },
  ];

  const clearFilters = () => {
    setSearchQuery("");
    setStatusFilter("all");
    setDocumentTypeFilter("all");
    setPriorityFilter("all");
  };
  const hasActiveFilters =
    Boolean(searchQuery) ||
    statusFilter !== "all" ||
    documentTypeFilter !== "all" ||
    priorityFilter !== "all";

  return (
    <div className="space-y-4">
      <FilterBar
        search={
          <div className="relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
            <Input
              placeholder="Search tasks…"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="pl-10"
            />
          </div>
        }
        filters={
          <>
            <Select value={statusFilter} onValueChange={setStatusFilter}>
              <SelectTrigger className="w-35">
                <SelectValue placeholder="Status" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All statuses</SelectItem>
                <SelectItem value="pending">Pending</SelectItem>
                <SelectItem value="claimed">Claimed</SelectItem>
                <SelectItem value="completed">Completed</SelectItem>
              </SelectContent>
            </Select>
            <Select value={documentTypeFilter} onValueChange={setDocumentTypeFilter}>
              <SelectTrigger className="w-40">
                <SelectValue placeholder="Document" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All types</SelectItem>
                <SelectItem value="requisition">Requisition</SelectItem>
                <SelectItem value="purchase_order">Purchase Order</SelectItem>
                <SelectItem value="payment_voucher">Payment Voucher</SelectItem>
                <SelectItem value="grn">GRN</SelectItem>
                <SelectItem value="budget">Budget</SelectItem>
              </SelectContent>
            </Select>
            <Select value={priorityFilter} onValueChange={setPriorityFilter}>
              <SelectTrigger className="w-[130px]">
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
          </>
        }
        hasActiveFilters={hasActiveFilters}
        onReset={clearFilters}
        meta={`${filteredTasks.length} task${filteredTasks.length === 1 ? "" : "s"}${hasActiveFilters ? " (filtered)" : ""}`}
      />

      <DataList<WorkflowTask>
        rows={filteredTasks}
        columns={columns}
        getRowId={(t) => t.id}
        isLoading={isLoading}
        emptyMessage="No tasks found."
        mobileCard={(t) => (
          <div className="flex flex-col gap-2">
            <div className="flex items-start justify-between gap-2">
              <div className="min-w-0">
                <div className="font-medium capitalize line-clamp-1">{getTitle(t)}</div>
                <div className="text-xs text-muted-foreground">{getDocumentNumber(t)}</div>
              </div>
              <StatusBadge status={t.status} type="execution" />
            </div>
            <div className="flex flex-wrap items-center gap-2 text-xs text-muted-foreground">
              <PriorityBadge priority={t.priority} />
              {t.stageName && <span>{t.stageName}</span>}
              <DueCell t={t} />
            </div>
            <div className="pt-1">
              <WorkflowActionButtons
                task={t as any}
                onClaim={handleClaimTask}
                onApprove={handleApproveTask}
                onReject={handleRejectTask}
                onReassign={handleReassignTask}
                onRefresh={handleRefresh}
                variant="compact"
                showViewButton
                onView={(task) => router.push(getDocRoute(task))}
              />
            </div>
          </div>
        )}
      />

      {paginationMeta && (
        <CustomPagination
          pagination={{
            ...paginationMeta,
            page: paginationState.page,
            page_size: paginationState.page_size,
            limit: paginationMeta.limit || paginationState.page_size,
            totalCount: paginationMeta.totalCount || paginationMeta.total || 0,
            total_pages: paginationMeta.total_pages || paginationMeta.totalPages || 0,
            has_next: paginationMeta.has_next ?? paginationMeta.hasNext ?? false,
            has_prev: paginationMeta.has_prev ?? paginationMeta.hasPrev ?? false,
          }}
          updatePagination={(np: { page: number; page_size?: number }) =>
            setPaginationState((prev) => ({
              ...prev,
              page: np.page,
              page_size: np.page_size || prev.page_size,
            }))
          }
          allowSetPageSize
          showDetails
        />
      )}
    </div>
  );
}
