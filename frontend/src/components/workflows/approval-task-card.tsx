"use client";

import React, { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { Card } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  Clock,
  User,
  Users,
  CheckCircle,
  XCircle,
  AlertTriangle,
  FileText,
  ShoppingCart,
  Wallet,
  Package,
  DollarSign,
  Calendar,
  Timer,
  Eye,
  UserCheck,
  Zap,
} from "lucide-react";
import { formatDistanceToNow, isAfter, format } from "date-fns";
import { useApprovalWorkflow } from "@/hooks/use-approval-workflow";
import {
  canUserActOnWorkflowTask,
  formatRoleForDisplay,
} from "@/lib/workflow-utils";
import { ClaimTaskModal } from "./claim-task-modal";
import { ApprovalActionModal } from "./approval-action-modal";
import { toast } from "sonner";
import { cn } from "@/lib/utils";

interface ApprovalTaskCardProps {
  taskId: string;
  currentUserId: string;
  currentUserRole: string;
}

// ── Doc-type visual config ──────────────────────────────────────────────────

type DocAccent = {
  icon: React.ComponentType<{ className?: string }>;
  chip: string;
  label: string;
  route: (id: string) => string;
};

const DOC_CONFIG: Record<string, DocAccent> = {
  requisition: {
    icon: FileText,
    chip: "bg-violet-100 text-violet-700 dark:bg-violet-950/50 dark:text-violet-300",
    label: "Requisition",
    route: (id) => `/requisitions/${id}`,
  },
  purchase_order: {
    icon: ShoppingCart,
    chip: "bg-blue-100 text-blue-700 dark:bg-blue-950/50 dark:text-blue-300",
    label: "Purchase Order",
    route: (id) => `/purchase-orders/${id}`,
  },
  payment_voucher: {
    icon: Wallet,
    chip: "bg-emerald-100 text-emerald-700 dark:bg-emerald-950/50 dark:text-emerald-300",
    label: "Payment Voucher",
    route: (id) => `/payment-vouchers/${id}`,
  },
  grn: {
    icon: Package,
    chip: "bg-amber-100 text-amber-700 dark:bg-amber-950/50 dark:text-amber-300",
    label: "Goods Received Note",
    route: (id) => `/grn/${id}`,
  },
  goods_received_note: {
    icon: Package,
    chip: "bg-amber-100 text-amber-700 dark:bg-amber-950/50 dark:text-amber-300",
    label: "Goods Received Note",
    route: (id) => `/grn/${id}`,
  },
  budget: {
    icon: DollarSign,
    chip: "bg-teal-100 text-teal-700 dark:bg-teal-950/50 dark:text-teal-300",
    label: "Budget",
    route: (id) => `/budgets/${id}`,
  },
};

const FALLBACK_CONFIG: DocAccent = {
  icon: FileText,
  chip: "bg-slate-100 text-slate-700 dark:bg-slate-800 dark:text-slate-300",
  label: "Document",
  route: (id) => `/tasks/${id}`,
};

function getDocConfig(entityType?: string): DocAccent {
  return DOC_CONFIG[(entityType || "").toLowerCase()] || FALLBACK_CONFIG;
}

// ── Priority pill ───────────────────────────────────────────────────────────

const PRIORITY_STYLES: Record<string, string> = {
  URGENT:
    "bg-red-100 text-red-700 border-red-200 dark:bg-red-950/50 dark:text-red-300 dark:border-red-900",
  HIGH: "bg-orange-100 text-orange-700 border-orange-200 dark:bg-orange-950/50 dark:text-orange-300 dark:border-orange-900",
  MEDIUM:
    "bg-blue-100 text-blue-700 border-blue-200 dark:bg-blue-950/50 dark:text-blue-300 dark:border-blue-900",
  LOW: "bg-slate-100 text-slate-700 border-slate-200 dark:bg-slate-800 dark:text-slate-300 dark:border-slate-700",
};

// ── Component ───────────────────────────────────────────────────────────────

export function ApprovalTaskCard({
  taskId,
  currentUserId,
  currentUserRole,
}: ApprovalTaskCardProps) {
  const router = useRouter();
  const {
    task,
    isLoading,
    error,
    claim,
    unclaim,
    approve,
    reject,
    isClaiming,
    isUnclaiming,
    isApproving,
    isRejecting,
    isProcessing,
  } = useApprovalWorkflow(taskId);

  const [showClaimModal, setShowClaimModal] = useState(false);
  const [showApprovalModal, setShowApprovalModal] = useState(false);
  const [approvalAction, setApprovalAction] = useState<"approve" | "reject">(
    "approve",
  );
  const [timeRemaining, setTimeRemaining] = useState<number>(0);

  useEffect(() => {
    if (task?.claimExpiry) {
      const updateTimer = () => {
        const remaining =
          new Date(task.claimExpiry).getTime() - new Date().getTime();
        setTimeRemaining(Math.max(0, remaining));
      };
      updateTimer();
      const interval = setInterval(updateTimer, 60000);
      return () => clearInterval(interval);
    }
  }, [task?.claimExpiry]);

  if (isLoading) {
    return (
      <Card className="p-4 border-border/60">
        <div className="animate-pulse space-y-3">
          <div className="flex items-center gap-3">
            <div className="h-9 w-9 rounded-md bg-muted" />
            <div className="flex-1 space-y-1.5">
              <div className="h-4 bg-muted rounded w-48" />
              <div className="h-3 bg-muted rounded w-32" />
            </div>
            <div className="h-5 w-16 bg-muted rounded" />
          </div>
          <div className="h-3 bg-muted rounded w-3/4" />
        </div>
      </Card>
    );
  }

  if (error || !task) {
    return (
      <Card className="border-red-200 dark:border-red-900 p-4">
        <div className="flex items-center gap-2 text-red-600 dark:text-red-400 text-sm">
          <AlertTriangle className="h-4 w-4" />
          <span>Failed to load task details</span>
        </div>
      </Card>
    );
  }

  const taskStatus = task.status?.toUpperCase();
  const isPending = taskStatus === "PENDING";
  const isClaimedByMe =
    taskStatus === "CLAIMED" && task.claimedBy === currentUserId;
  const isClaimedByOther =
    taskStatus === "CLAIMED" && task.claimedBy !== currentUserId;
  const isApproved = taskStatus === "APPROVED" || taskStatus === "COMPLETED";
  const isRejected = taskStatus === "REJECTED";
  const isTerminal = isApproved || isRejected;

  const canUserClaim = canUserActOnWorkflowTask(
    { id: currentUserId, role: currentUserRole as any },
    task,
  );

  const isClaimExpired =
    task.claimExpiry && isAfter(new Date(), new Date(task.claimExpiry));
  const minutesRemaining = Math.floor(timeRemaining / (1000 * 60));

  const docConfig = getDocConfig(task.entityType);
  const DocIcon = docConfig.icon;
  const docLabel = task.taskType
    ? task.taskType.replace(/_/g, " ").toLowerCase()
    : docConfig.label;
  const displayTitle = task.documentNumber || docConfig.label;
  const displaySubtitle = task.title || docConfig.label + " Approval";
  const priority = (task.priority || "MEDIUM").toString().toUpperCase();

  const handleClaim = async () => {
    try {
      await claim();
      setShowClaimModal(false);
      toast.success("Task claimed — you can now approve or reject it.");
    } catch (err: any) {
      toast.error(err.message || "Failed to claim task");
    }
  };

  const handleUnclaim = async () => {
    try {
      await unclaim();
      toast.success("Task unclaimed — others can now claim it.");
    } catch (err: any) {
      toast.error(err.message || "Failed to unclaim task");
    }
  };

  const handleApprovalAction = async (data: {
    comments: string;
    signature: string;
  }) => {
    try {
      if (approvalAction === "approve") {
        await approve({ ...data, expectedVersion: task.version });
        toast.success("Task approved");
      } else {
        await reject({
          remarks: data.comments,
          signature: data.signature,
          expectedVersion: task.version,
        });
        toast.success("Task rejected");
      }
      setShowApprovalModal(false);
    } catch (err: any) {
      if (
        err.message?.includes("version") ||
        err.message?.includes("modified by another user")
      ) {
        toast.error(
          "Task was modified by another user. Please refresh and try again.",
        );
      } else {
        toast.error(err.message || `Failed to ${approvalAction} task`);
      }
    }
  };

  const handleView = () => {
    router.push(docConfig.route(task.entityId || task.documentId));
  };

  // Card border/tint by state
  const cardTone = isClaimedByMe
    ? "border-blue-300 dark:border-blue-800 bg-blue-50/40 dark:bg-blue-950/20"
    : isRejected
      ? "border-red-200 dark:border-red-900 bg-red-50/30 dark:bg-red-950/20"
      : isApproved
        ? "border-emerald-200 dark:border-emerald-900 bg-emerald-50/30 dark:bg-emerald-950/20"
        : "border-border/60 hover:border-border";

  return (
    <>
      <Card className={cn("p-4 transition-colors", cardTone)}>
        {/* Header row: icon chip | title + docNumber | status */}
        <div className="flex items-start gap-3">
          <span
            className={cn(
              "flex items-center justify-center rounded-md h-9 w-9 shrink-0",
              docConfig.chip,
            )}
          >
            <DocIcon className="h-4 w-4" />
          </span>

          <div className="flex-1 min-w-0 space-y-0.5">
            <div className="flex flex-wrap items-center gap-x-2 gap-y-1">
              <span className="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">
                {docLabel}
              </span>
              <PriorityPill priority={priority} />
            </div>
            <div className="flex flex-wrap items-baseline gap-x-2 gap-y-0.5">
              <p className="font-semibold text-sm truncate">{displaySubtitle}</p>
              <span className="font-mono text-xs text-blue-600 dark:text-blue-400">
                {displayTitle}
              </span>
            </div>
            {task.stageName && (
              <p className="text-xs text-muted-foreground">
                Stage: <span className="font-medium">{task.stageName}</span>
              </p>
            )}
          </div>

          <StatusBadge
            isPending={isPending}
            isClaimedByMe={isClaimedByMe}
            isClaimedByOther={isClaimedByOther}
            isApproved={isApproved}
            isRejected={isRejected}
            claimerName={task.claimerName}
          />
        </div>

        {/* Meta rows */}
        <div className="mt-3 grid grid-cols-1 sm:grid-cols-2 gap-x-4 gap-y-1.5 text-xs">
          <MetaRow
            icon={<Users className="h-3.5 w-3.5" />}
            label="Required role"
            value={formatRoleForDisplay(
              task.assignedRole,
              task.assignedRoleName,
            )}
          />
          <MetaRow
            icon={<Calendar className="h-3.5 w-3.5" />}
            label="Created"
            value={`${formatDistanceToNow(new Date(task.createdAt))} ago`}
          />
          {task.dueDate && !isTerminal && (
            <MetaRow
              icon={<Clock className="h-3.5 w-3.5" />}
              label="Due"
              value={
                <span
                  className={
                    isAfter(new Date(), new Date(task.dueDate))
                      ? "text-red-600 dark:text-red-400 font-medium"
                      : ""
                  }
                >
                  {formatDistanceToNow(new Date(task.dueDate))}{" "}
                  {isAfter(new Date(), new Date(task.dueDate))
                    ? "overdue"
                    : "remaining"}
                </span>
              }
            />
          )}
          {isClaimedByMe && task.claimExpiry && (
            <MetaRow
              icon={<Timer className="h-3.5 w-3.5 text-amber-500" />}
              label="Claim expires"
              value={
                <span
                  className={
                    minutesRemaining < 10
                      ? "text-red-600 dark:text-red-400 font-medium"
                      : "text-amber-600 dark:text-amber-400"
                  }
                >
                  {minutesRemaining} min remaining
                </span>
              }
            />
          )}
          {isApproved && task.approverName && (
            <MetaRow
              icon={<UserCheck className="h-3.5 w-3.5 text-emerald-600" />}
              label="Approved by"
              value={task.approverName}
            />
          )}
          {isApproved && task.approvedAt && (
            <MetaRow
              icon={<CheckCircle className="h-3.5 w-3.5 text-emerald-600" />}
              label="Approved on"
              value={format(new Date(task.approvedAt), "PP")}
            />
          )}
          {isRejected && task.rejectedBy && (
            <MetaRow
              icon={<XCircle className="h-3.5 w-3.5 text-red-600" />}
              label="Rejected by"
              value={task.rejectedBy}
            />
          )}
          {isClaimedByOther && (
            <MetaRow
              icon={<User className="h-3.5 w-3.5 text-amber-600" />}
              label="Being reviewed by"
              value={task.claimerName || "Another user"}
            />
          )}
        </div>

        {/* Action bar */}
        <div className="mt-3 pt-3 border-t border-border/40 flex flex-wrap items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={handleView}
            className="gap-1.5"
          >
            <Eye className="h-3.5 w-3.5" />
            View Details
          </Button>

          {isPending && canUserClaim && (
            <Button
              size="sm"
              onClick={() => setShowClaimModal(true)}
              disabled={isClaiming || isProcessing}
              isLoading={isClaiming}
              loadingText="Claiming..."
              className="gap-1.5 bg-blue-600 hover:bg-blue-700"
            >
              <Users className="h-3.5 w-3.5" />
              Claim Task
            </Button>
          )}

          {isClaimedByMe && !isClaimExpired && (
            <>
              <Button
                size="sm"
                onClick={() => {
                  setApprovalAction("approve");
                  setShowApprovalModal(true);
                }}
                disabled={isProcessing}
                isLoading={isApproving}
                loadingText="Approving..."
                className="gap-1.5 bg-emerald-600 hover:bg-emerald-700"
              >
                <CheckCircle className="h-3.5 w-3.5" />
                Approve
              </Button>
              <Button
                size="sm"
                onClick={() => {
                  setApprovalAction("reject");
                  setShowApprovalModal(true);
                }}
                disabled={isProcessing}
                isLoading={isRejecting}
                loadingText="Rejecting..."
                className="gap-1.5 bg-red-600 hover:bg-red-700"
              >
                <XCircle className="h-3.5 w-3.5" />
                Reject
              </Button>
              <Button
                size="sm"
                variant="outline"
                onClick={handleUnclaim}
                disabled={isProcessing}
                isLoading={isUnclaiming}
                loadingText="Unclaiming..."
              >
                Unclaim
              </Button>
            </>
          )}

          {isPending && !canUserClaim && (
            <span className="text-xs text-muted-foreground">
              {task.assignedUserId
                ? "Assigned to a specific user."
                : `Requires the "${formatRoleForDisplay(task.assignedRole, task.assignedRoleName)}" role.`}
            </span>
          )}

          {isClaimedByOther && (
            <span className="text-xs text-muted-foreground">
              Claim expires in{" "}
              {task.claimExpiry
                ? Math.floor(
                    (new Date(task.claimExpiry).getTime() -
                      new Date().getTime()) /
                      (1000 * 60),
                  )
                : 0}{" "}
              min
            </span>
          )}
        </div>
      </Card>

      <ClaimTaskModal
        isOpen={showClaimModal}
        onClose={() => setShowClaimModal(false)}
        onConfirm={handleClaim}
        isLoading={isClaiming}
        taskDetails={{
          entityType: task.entityType,
          entityId: task.entityId,
          stageName: task.stageName,
          assignedRole: formatRoleForDisplay(
            task.assignedRole,
            task.assignedRoleName,
          ),
        }}
      />

      <ApprovalActionModal
        isOpen={showApprovalModal}
        onClose={() => setShowApprovalModal(false)}
        onConfirm={handleApprovalAction}
        isLoading={isApproving || isRejecting}
        action={approvalAction}
        taskDetails={{
          entityType: task.entityType,
          entityId: task.entityId,
          stageName: task.stageName,
          claimedBy: "You",
          claimExpiry: task.claimExpiry || "",
        }}
      />
    </>
  );
}

// ── Sub-components ──────────────────────────────────────────────────────────

function PriorityPill({ priority }: { priority: string }) {
  const style = PRIORITY_STYLES[priority] || PRIORITY_STYLES.MEDIUM;
  return (
    <span
      className={cn(
        "inline-flex items-center gap-1 px-1.5 py-0.5 rounded-sm text-[10px] font-semibold uppercase tracking-wider border",
        style,
      )}
    >
      {priority === "URGENT" && <Zap className="h-2.5 w-2.5" />}
      {priority}
    </span>
  );
}

interface StatusBadgeProps {
  isPending: boolean;
  isClaimedByMe: boolean;
  isClaimedByOther: boolean;
  isApproved: boolean;
  isRejected: boolean;
  claimerName?: string;
}

function StatusBadge({
  isPending,
  isClaimedByMe,
  isClaimedByOther,
  isApproved,
  isRejected,
  claimerName,
}: StatusBadgeProps) {
  if (isApproved) {
    return (
      <Badge className="bg-emerald-100 text-emerald-800 dark:bg-emerald-950/50 dark:text-emerald-300 border-emerald-200 dark:border-emerald-900 shrink-0">
        Approved
      </Badge>
    );
  }
  if (isRejected) {
    return (
      <Badge className="bg-red-100 text-red-800 dark:bg-red-950/50 dark:text-red-300 border-red-200 dark:border-red-900 shrink-0">
        Rejected
      </Badge>
    );
  }
  if (isClaimedByMe) {
    return (
      <Badge className="bg-blue-600 text-white shrink-0">Claimed by You</Badge>
    );
  }
  if (isClaimedByOther) {
    return (
      <Badge
        variant="outline"
        className="border-amber-300 text-amber-700 dark:border-amber-800 dark:text-amber-300 shrink-0"
      >
        Claimed by {claimerName || "Other"}
      </Badge>
    );
  }
  if (isPending) {
    return (
      <Badge variant="secondary" className="shrink-0">
        Available
      </Badge>
    );
  }
  return <Badge variant="outline" className="shrink-0">Unknown</Badge>;
}

function MetaRow({
  icon,
  label,
  value,
}: {
  icon: React.ReactNode;
  label: string;
  value: React.ReactNode;
}) {
  return (
    <div className="flex items-center gap-1.5 min-w-0">
      <span className="text-muted-foreground shrink-0">{icon}</span>
      <span className="text-muted-foreground shrink-0">{label}:</span>
      <span className="font-medium truncate">{value}</span>
    </div>
  );
}
