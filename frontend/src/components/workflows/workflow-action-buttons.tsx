"use client";

import { useState } from "react";
import {
  CheckCircle2,
  X,
  UserCheck,
  Clock,
  MoreHorizontal,
  Users,
  XCircle,
  User,
  AlertTriangle,
  Eye,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { useSession } from "@/hooks/use-session";
import { toast } from "sonner";
import { formatDistanceToNow } from "date-fns";

// Define WorkflowTask interface locally since it's not exported from types
interface WorkflowTask {
  id: string;
  status: string;
  claimedBy?: string;
  assignedRole?: string;
  assignedUserId?: string;
  stageName?: string;
  claimExpiry?: string;
  entityType?: string;
  entityId?: string;
  documentType?: string;
  documentId?: string;
}

interface WorkflowActionButtonsProps {
  task: WorkflowTask;
  onClaim?: (taskId: string) => Promise<void>;
  onApprove?: (taskId: string) => Promise<void>;
  onReject?: (taskId: string) => Promise<void>;
  onRefresh?: () => void;
  variant?: "table" | "detail" | "dropdown" | "inline" | "compact"; // Enhanced variants
  showViewButton?: boolean;
  showStatus?: boolean;
  onView?: (task: WorkflowTask) => void;
  onActionComplete?: () => void;
}

export function WorkflowActionButtons({
  task,
  onClaim,
  onApprove,
  onReject,
  onRefresh,
  variant = "table",
  showViewButton = true,
  showStatus = false,
  onView,
  onActionComplete,
}: WorkflowActionButtonsProps) {
  const { user } = useSession();
  const [isLoading, setIsLoading] = useState<string | null>(null);

  // Status badge component
  const StatusBadge = () => {
    const isPending = task.status === "pending";
    const isClaimedByUser = task.claimedBy === user?.id;
    const isClaimedByOther = task.claimedBy && task.claimedBy !== user?.id;
    const isCompleted =
      task.status === "completed" || task.status === "approved";

    if (isPending && !task.claimedBy) {
      return (
        <Badge variant="secondary" className="text-xs">
          Pending Approval
        </Badge>
      );
    }
    if (isClaimedByUser) {
      return (
        <Badge variant="default" className="text-xs bg-blue-600">
          Claimed by You
        </Badge>
      );
    }
    if (isClaimedByOther) {
      return (
        <Badge variant="destructive" className="text-xs">
          Claimed by Other
        </Badge>
      );
    }
    if (isCompleted) {
      return (
        <Badge variant="default" className="text-xs bg-green-600">
          Completed
        </Badge>
      );
    }
    return (
      <Badge variant="secondary" className="text-xs">
        Unknown
      </Badge>
    );
  };

  const canUserClaimTask = () => {
    if (!user) return false;

    // Admin can claim any task
    if (user.role === "admin") return true;

    // User can claim tasks assigned to their role or specifically to them
    return task.assignedRole === user.role || task.assignedUserId === user.id;
  };

  const isTaskClaimedByUser = () => {
    if (!user) return false;
    return task.claimedBy === user.id;
  };

  const isTaskClaimed = () => {
    return task.status === "claimed" || task.claimedBy !== null;
  };

  const canUserApproveReject = () => {
    if (!user) return false;

    // User can only approve/reject if they have claimed the task
    return isTaskClaimedByUser();
  };

  const isPending = task.status === "pending";
  const canClaim = canUserClaimTask() && !isTaskClaimed();
  const canApproveReject = canUserApproveReject();

  const handleAction = async (
    action: string,
    handler?: (taskId: string) => Promise<void>
  ) => {
    if (!handler) return;

    setIsLoading(action);
    try {
      await handler(task.id);
      toast.success(`Task ${action}d successfully`);
      onRefresh?.();
      onActionComplete?.();
    } catch (error) {
      toast.error(`Failed to ${action} task`);
    } finally {
      setIsLoading(null);
    }
  };

  const handleView = () => {
    if (onView) {
      onView(task);
      return;
    }

    // Default view navigation
    const docType = (task.entityType || task.documentType || "").toLowerCase();
    const docId = task.entityId || task.documentId;
    const routes: Record<string, string> = {
      requisition: `/requisitions/${docId}`,
      purchase_order: `/purchase-orders/${docId}`,
      payment_voucher: `/payment-vouchers/${docId}`,
      goods_received_note: `/grn/${docId}`,
      budget: `/budgets/${docId}`,
    };
    const url = routes[docType || ""] || `/tasks/${task.id}`;
    window.location.href = url;
  };

  // Compact variant (for table cells)
  if (variant === "compact") {
    return (
      <div className="flex items-center gap-2">
        {showStatus && <StatusBadge />}

        {isPending && canClaim && (
          <Button
            size="sm"
            variant="outline"
            onClick={() => handleAction("claim", onClaim)}
            disabled={isLoading === "claim"}
            className="h-7 px-2 text-xs"
          >
            <Users className="h-3 w-3 mr-1" />
            Claim
          </Button>
        )}

        {isTaskClaimed() && canApproveReject && (
          <div className="flex gap-1">
            <Button
              size="sm"
              onClick={() => handleAction("approve", onApprove)}
              disabled={isLoading === "approve"}
              className="h-7 px-2 text-xs bg-green-600 hover:bg-green-700"
            >
              <CheckCircle2 className="h-3 w-3 mr-1" />
              Approve
            </Button>
            <Button
              size="sm"
              onClick={() => handleAction("reject", onReject)}
              disabled={isLoading === "reject"}
              className="h-7 px-2 text-xs bg-red-600 hover:bg-red-700"
            >
              <XCircle className="h-3 w-3 mr-1" />
              Reject
            </Button>
          </div>
        )}
      </div>
    );
  }

  // Inline variant (for detail pages with full status display)
  if (variant === "inline") {
    return (
      <div className="space-y-4">
        {/* Status Display */}
        <div className="flex items-center justify-between p-4 bg-gray-50 rounded-lg border">
          <div className="flex items-center gap-3">
            <div className="flex items-center gap-2">
              <Clock className="h-4 w-4 text-gray-500" />
              <span className="text-sm font-medium">Workflow Status:</span>
            </div>
            <StatusBadge />
          </div>

          {task.claimExpiry && isTaskClaimedByUser() && (
            <div className="text-sm text-gray-600">
              Expires: {formatDistanceToNow(new Date(task.claimExpiry))}{" "}
              remaining
            </div>
          )}
        </div>

        {/* Action Buttons */}
        <div className="flex flex-wrap gap-3">
          {isPending && !isTaskClaimed() && canClaim && (
            <Button
              onClick={() => handleAction("claim", onClaim)}
              disabled={isLoading === "claim"}
              className="bg-blue-600 hover:bg-blue-700"
            >
              {isLoading === "claim" ? (
                <>
                  <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2" />
                  Claiming...
                </>
              ) : (
                <>
                  <Users className="h-4 w-4 mr-2" />
                  Claim for Review
                </>
              )}
            </Button>
          )}

          {isTaskClaimed() && canApproveReject && (
            <>
              <Button
                onClick={() => handleAction("approve", onApprove)}
                disabled={isLoading === "approve"}
                className="bg-green-600 hover:bg-green-700"
              >
                {isLoading === "approve" ? (
                  <>
                    <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2" />
                    Approving...
                  </>
                ) : (
                  <>
                    <CheckCircle2 className="h-4 w-4 mr-2" />
                    Approve
                  </>
                )}
              </Button>

              <Button
                onClick={() => handleAction("reject", onReject)}
                disabled={isLoading === "reject"}
                className="bg-red-600 hover:bg-red-700"
              >
                {isLoading === "reject" ? (
                  <>
                    <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2" />
                    Rejecting...
                  </>
                ) : (
                  <>
                    <XCircle className="h-4 w-4 mr-2" />
                    Reject
                  </>
                )}
              </Button>
            </>
          )}

          {isTaskClaimed() && !isTaskClaimedByUser() && (
            <div className="flex items-center gap-2 text-gray-600 bg-gray-100 px-3 py-2 rounded">
              <User className="h-4 w-4" />
              <span className="text-sm">
                Currently being reviewed by another user
              </span>
            </div>
          )}

          {isPending && !canClaim && (
            <div className="flex items-center gap-2 text-amber-600 bg-amber-50 px-3 py-2 rounded border border-amber-200">
              <AlertTriangle className="h-4 w-4" />
              <span className="text-sm">
                Requires "{task.assignedRole}" role to approve
              </span>
            </div>
          )}
        </div>
      </div>
    );
  }

  // Dropdown variant (for table action menus)
  if (variant === "dropdown") {
    return (
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button variant="ghost" className="h-8 w-8 p-0">
            <span className="sr-only">Open menu</span>
            <MoreHorizontal className="h-4 w-4" />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end" className="w-56">
          {/* Status Display */}
          <div className="px-2 py-1.5 text-sm font-medium text-gray-700 border-b">
            Workflow Status
          </div>
          <div className="px-2 py-1.5 flex items-center justify-between">
            <span className="text-sm text-gray-600">Current:</span>
            <StatusBadge />
          </div>

          {task.stageName && (
            <div className="px-2 py-1.5 flex items-center justify-between">
              <span className="text-sm text-gray-600">Stage:</span>
              <span className="text-sm font-medium">{task.stageName}</span>
            </div>
          )}

          <DropdownMenuSeparator />

          {/* Actions */}
          {isPending && canClaim && !isTaskClaimed() && (
            <DropdownMenuItem
              onClick={() => handleAction("claim", onClaim)}
              disabled={isLoading === "claim"}
            >
              <Users className="mr-2 h-4 w-4" />
              <span>Claim Task</span>
            </DropdownMenuItem>
          )}

          {isTaskClaimed() && canApproveReject && (
            <>
              <DropdownMenuItem
                onClick={() => handleAction("approve", onApprove)}
                disabled={isLoading === "approve"}
              >
                <CheckCircle2 className="mr-2 h-4 w-4 text-green-600" />
                <span>Approve</span>
              </DropdownMenuItem>

              <DropdownMenuItem
                onClick={() => handleAction("reject", onReject)}
                disabled={isLoading === "reject"}
              >
                <XCircle className="mr-2 h-4 w-4 text-red-600" />
                <span>Reject</span>
              </DropdownMenuItem>
            </>
          )}

          {isTaskClaimed() && !isTaskClaimedByUser() && (
            <DropdownMenuItem disabled>
              <User className="mr-2 h-4 w-4" />
              <span>Claimed by Other</span>
            </DropdownMenuItem>
          )}

          {isPending && !canClaim && (
            <DropdownMenuItem disabled>
              <AlertTriangle className="mr-2 h-4 w-4 text-amber-500" />
              <span>Requires {task.assignedRole} role</span>
            </DropdownMenuItem>
          )}

          <DropdownMenuSeparator />

          <DropdownMenuItem onClick={handleView}>
            <Eye className="mr-2 h-4 w-4" />
            <span>View Details</span>
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    );
  }

  if (variant === "detail") {
    // Detail page layout - larger buttons, more prominent
    return (
      <div className="flex flex-wrap gap-3">
        {isPending && (
          <>
            {/* Step 1: Show Claim button if task is not claimed and user can claim */}
            {!isTaskClaimed() && canClaim && onClaim && (
              <Button
                size="default"
                variant="outline"
                onClick={() => handleAction("claim", onClaim)}
                disabled={isLoading === "claim"}
              >
                <UserCheck className="h-4 w-4 mr-2" />
                {isLoading === "claim" ? "Claiming..." : "Claim Task"}
              </Button>
            )}

            {/* Step 2: Show Approve/Reject buttons only if user has claimed the task */}
            {isTaskClaimed() && canApproveReject && (
              <>
                {onApprove && (
                  <Button
                    size="default"
                    onClick={() => handleAction("approve", onApprove)}
                    disabled={isLoading === "approve"}
                    className="bg-green-600 hover:bg-green-700"
                  >
                    <CheckCircle2 className="h-4 w-4 mr-2" />
                    {isLoading === "approve" ? "Approving..." : "Approve"}
                  </Button>
                )}

                {onReject && (
                  <Button
                    size="default"
                    variant="destructive"
                    onClick={() => handleAction("reject", onReject)}
                    disabled={isLoading === "reject"}
                  >
                    <X className="h-4 w-4 mr-2" />
                    {isLoading === "reject" ? "Rejecting..." : "Reject"}
                  </Button>
                )}
              </>
            )}

            {/* Show status if task is claimed by someone else */}
            {isTaskClaimed() && !isTaskClaimedByUser() && (
              <div className="flex items-center text-sm text-gray-600">
                <UserCheck className="h-4 w-4 mr-2" />
                Task claimed by another user
              </div>
            )}
          </>
        )}

        {!isPending && task.status === "approved" && showStatus && (
          <Button
            size="default"
            variant="outline"
            disabled
            className="text-green-600"
          >
            <CheckCircle2 className="h-4 w-4 mr-2" />
            Approved
          </Button>
        )}

        {!isPending && task.status === "rejected" && showStatus && (
          <Button
            size="default"
            variant="outline"
            disabled
            className="text-red-600"
          >
            <X className="h-4 w-4 mr-2" />
            Rejected
          </Button>
        )}
      </div>
    );
  }

  // Table layout - compact buttons
  return (
    <div className="flex items-center justify-end gap-2 ml-auto">
      {showViewButton && (
        <Button
          size="sm"
          variant="outline"
          onClick={handleView}
          className="text-xs"
        >
          <Clock className="h-4 w-4 mr-1" />
          View
        </Button>
      )}
      {/* Action Buttons for Pending Tasks */}
      {isPending && (
        <>
          {/* Step 1: Show Claim button if task is not claimed and user can claim */}
          {!isTaskClaimed() && canClaim && onClaim && (
            <Button
              size="sm"
              variant="outline"
              onClick={() => handleAction("claim", onClaim)}
              disabled={isLoading === "claim"}
            >
              <UserCheck className="h-4 w-4 mr-1" />
              {isLoading === "claim" ? "..." : "Claim"}
            </Button>
          )}

          {/* Step 2: Show Approve/Reject buttons only if user has claimed the task */}
          {isTaskClaimed() && canApproveReject && (
            <>
              {onApprove && (
                <Button
                  size="sm"
                  variant="default"
                  onClick={() => handleAction("approve", onApprove)}
                  disabled={isLoading === "approve"}
                  className="bg-green-600 hover:bg-green-700"
                >
                  <CheckCircle2 className="h-4 w-4 mr-1" />
                  {isLoading === "approve" ? "..." : "Approve"}
                </Button>
              )}

              {onReject && (
                <Button
                  size="sm"
                  variant="destructive"
                  onClick={() => handleAction("reject", onReject)}
                  disabled={isLoading === "reject"}
                >
                  <X className="h-4 w-4 mr-1" />
                  {isLoading === "reject" ? "..." : "Reject"}
                </Button>
              )}
            </>
          )}

          {/* Show claimed status if task is claimed by someone else */}
          {isTaskClaimed() && !isTaskClaimedByUser() && canClaim && (
            <Badge variant={"outline"} className="text-xs text-gray-500">
              Claimed
            </Badge>
          )}
        </>
      )}

      {/* More Actions Dropdown */}
      <>
        {showStatus && <StatusBadge />}

        {/* More Actions Dropdown */}
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" size="sm">
              <MoreHorizontal className="h-4 w-4" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuItem onClick={handleView}>
              <Clock className="h-4 w-4 mr-2" />
              View Details
            </DropdownMenuItem>

            {canClaim && onClaim && !isTaskClaimed() && (
              <>
                <DropdownMenuSeparator />
                <DropdownMenuItem
                  onClick={() => handleAction("claim", onClaim)}
                  disabled={isLoading === "claim"}
                >
                  <UserCheck className="h-4 w-4 mr-2" />
                  Claim Task
                </DropdownMenuItem>
              </>
            )}
          </DropdownMenuContent>
        </DropdownMenu>
      </>
    </div>
  );
}
