"use client";

import React, { useState } from "react";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Badge } from "@/components/ui/badge";
import {
  MoreHorizontal,
  Users,
  CheckCircle,
  XCircle,
  Clock,
  User,
  AlertTriangle,
  Eye,
} from "lucide-react";
import { useApprovalWorkflow } from "@/hooks/use-approval-workflow";
import { ClaimTaskModal } from "./claim-task-modal";
import { ApprovalActionModal } from "./approval-action-modal";
import { toast } from "sonner";
import { formatDistanceToNow } from "date-fns";

interface WorkflowActionButtonProps {
  documentId: string;
  documentType: string;
  currentUserId: string;
  currentUserRole: string;
  variant?: "dropdown" | "inline" | "compact";
  showStatus?: boolean;
  onActionComplete?: () => void;
}

export function WorkflowActionButton({
  documentId,
  documentType,
  currentUserId,
  currentUserRole,
  variant = "dropdown",
  showStatus = true,
  onActionComplete,
}: WorkflowActionButtonProps) {
  const [showClaimModal, setShowClaimModal] = useState(false);
  const [showApprovalModal, setShowApprovalModal] = useState(false);
  const [approvalAction, setApprovalAction] = useState<"approve" | "reject">(
    "approve"
  );

  // Get workflow task for this document
  const {
    task,
    isLoading,
    claim,
    unclaim,
    approve,
    reject,
    isClaiming,
    isUnclaiming,
    isApproving,
    isRejecting,
    isProcessing,
  } = useApprovalWorkflow(documentId);

  if (isLoading) {
    return (
      <div className="flex items-center gap-2">
        <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-gray-400" />
        <span className="text-sm text-gray-500">Loading...</span>
      </div>
    );
  }

  if (!task) {
    return showStatus ? (
      <Badge variant="secondary" className="text-xs">
        No Workflow
      </Badge>
    ) : null;
  }

  // Task state calculations
  const isPending = task.status === "pending";
  const isClaimedByMe =
    task.status === "claimed" && task.claimedBy === currentUserId;
  const isClaimedByOther =
    task.status === "claimed" && task.claimedBy !== currentUserId;
  const isCompleted = task.status === "completed";
  const canUserClaim =
    task.assignedRole === currentUserRole ||
    task.assignedUserId === currentUserId;

  // Handle actions
  const handleClaim = async () => {
    try {
      await claim();
      setShowClaimModal(false);
      toast.success("Task claimed successfully!");
      onActionComplete?.();
    } catch (error: any) {
      toast.error(error.message || "Failed to claim task");
    }
  };

  const handleUnclaim = async () => {
    try {
      await unclaim();
      toast.success("Task unclaimed successfully!");
      onActionComplete?.();
    } catch (error: any) {
      toast.error(error.message || "Failed to unclaim task");
    }
  };

  const handleApprovalAction = async (data: {
    comments: string;
    signature: string;
  }) => {
    try {
      if (approvalAction === "approve") {
        await approve({
          ...data,
          expectedVersion: task.version,
        });
        toast.success("Task approved successfully!");
      } else {
        await reject({
          remarks: data.comments,
          signature: data.signature,
          expectedVersion: task.version,
        });
        toast.success("Task rejected successfully!");
      }
      setShowApprovalModal(false);
      onActionComplete?.();
    } catch (error: any) {
      if (
        error.message.includes("version") ||
        error.message.includes("modified by another user")
      ) {
        toast.error(
          "Task was modified by another user. Please refresh and try again."
        );
      } else {
        toast.error(error.message || `Failed to ${approvalAction} task`);
      }
    }
  };

  // Status badge component
  const StatusBadge = () => {
    if (isPending) {
      return (
        <Badge variant="secondary" className="text-xs">
          Pending Approval
        </Badge>
      );
    }
    if (isClaimedByMe) {
      return (
        <Badge variant="default" className="text-xs bg-blue-600">
          Claimed by You
        </Badge>
      );
    }
    if (isClaimedByOther) {
      return (
        <Badge variant="destructive" className="text-xs">
          Claimed by {task.claimerName || "Other"}
        </Badge>
      );
    }
    if (isCompleted) {
      return (
        <Badge variant="success" className="text-xs">
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

  // Compact variant (for table cells)
  if (variant === "compact") {
    return (
      <div className="flex items-center gap-2">
        {showStatus && <StatusBadge />}

        {isPending && canUserClaim && (
          <Button
            size="sm"
            variant="outline"
            onClick={() => setShowClaimModal(true)}
            disabled={isProcessing}
            className="h-7 px-2 text-xs"
          >
            <Users className="h-3 w-3 mr-1" />
            Claim
          </Button>
        )}

        {isClaimedByMe && (
          <div className="flex gap-1">
            <Button
              size="sm"
              onClick={() => {
                setApprovalAction("approve");
                setShowApprovalModal(true);
              }}
              disabled={isProcessing}
              className="h-7 px-2 text-xs bg-green-600 hover:bg-green-700"
            >
              <CheckCircle className="h-3 w-3 mr-1" />
              Approve
            </Button>
            <Button
              size="sm"
              onClick={() => {
                setApprovalAction("reject");
                setShowApprovalModal(true);
              }}
              disabled={isProcessing}
              className="h-7 px-2 text-xs bg-red-600 hover:bg-red-700"
            >
              <XCircle className="h-3 w-3 mr-1" />
              Reject
            </Button>
          </div>
        )}

        {/* Modals */}
        <ClaimTaskModal
          isOpen={showClaimModal}
          onClose={() => setShowClaimModal(false)}
          onConfirm={handleClaim}
          isLoading={isClaiming}
          taskDetails={{
            entityType: documentType,
            entityId: documentId,
            stageName: task.stageName,
            assignedRole: task.assignedRole,
          }}
        />

        <ApprovalActionModal
          isOpen={showApprovalModal}
          onClose={() => setShowApprovalModal(false)}
          onConfirm={handleApprovalAction}
          isLoading={isApproving || isRejecting}
          action={approvalAction}
          taskDetails={{
            entityType: documentType,
            entityId: documentId,
            stageName: task.stageName,
            claimedBy: "You",
            claimExpiry: task.claimExpiry || "",
          }}
        />
      </div>
    );
  }

  // Inline variant (for detail pages)
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

          {task.claimExpiry && isClaimedByMe && (
            <div className="text-sm text-gray-600">
              Expires: {formatDistanceToNow(new Date(task.claimExpiry))}{" "}
              remaining
            </div>
          )}
        </div>

        {/* Action Buttons */}
        <div className="flex flex-wrap gap-3">
          {isPending && canUserClaim && (
            <Button
              onClick={() => setShowClaimModal(true)}
              disabled={isProcessing}
              className="bg-blue-600 hover:bg-blue-700"
            >
              {isClaiming ? (
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

          {isClaimedByMe && (
            <>
              <Button
                onClick={() => {
                  setApprovalAction("approve");
                  setShowApprovalModal(true);
                }}
                disabled={isProcessing}
                className="bg-green-600 hover:bg-green-700"
              >
                {isApproving ? (
                  <>
                    <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2" />
                    Approving...
                  </>
                ) : (
                  <>
                    <CheckCircle className="h-4 w-4 mr-2" />
                    Approve
                  </>
                )}
              </Button>

              <Button
                onClick={() => {
                  setApprovalAction("reject");
                  setShowApprovalModal(true);
                }}
                disabled={isProcessing}
                className="bg-red-600 hover:bg-red-700"
              >
                {isRejecting ? (
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

              <Button
                variant="outline"
                onClick={handleUnclaim}
                disabled={isProcessing}
              >
                {isUnclaiming ? (
                  <>
                    <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-gray-600 mr-2" />
                    Unclaiming...
                  </>
                ) : (
                  "Release Claim"
                )}
              </Button>
            </>
          )}

          {isClaimedByOther && (
            <div className="flex items-center gap-2 text-gray-600 bg-gray-100 px-3 py-2 rounded">
              <User className="h-4 w-4" />
              <span className="text-sm">
                Currently being reviewed by {task.claimerName || "another user"}
              </span>
            </div>
          )}

          {isPending && !canUserClaim && (
            <div className="flex items-center gap-2 text-amber-600 bg-amber-50 px-3 py-2 rounded border border-amber-200">
              <AlertTriangle className="h-4 w-4" />
              <span className="text-sm">
                Requires "{task.assignedRole}" role to approve
              </span>
            </div>
          )}
        </div>

        {/* Modals */}
        <ClaimTaskModal
          isOpen={showClaimModal}
          onClose={() => setShowClaimModal(false)}
          onConfirm={handleClaim}
          isLoading={isClaiming}
          taskDetails={{
            entityType: documentType,
            entityId: documentId,
            stageName: task.stageName,
            assignedRole: task.assignedRole,
          }}
        />

        <ApprovalActionModal
          isOpen={showApprovalModal}
          onClose={() => setShowApprovalModal(false)}
          onConfirm={handleApprovalAction}
          isLoading={isApproving || isRejecting}
          action={approvalAction}
          taskDetails={{
            entityType: documentType,
            entityId: documentId,
            stageName: task.stageName,
            claimedBy: "You",
            claimExpiry: task.claimExpiry || "",
          }}
        />
      </div>
    );
  }

  // Dropdown variant (default - for table action menus)
  return (
    <>
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
          {isPending && canUserClaim && (
            <DropdownMenuItem
              onClick={() => setShowClaimModal(true)}
              disabled={isProcessing}
            >
              <Users className="mr-2 h-4 w-4" />
              <span>Claim Task</span>
            </DropdownMenuItem>
          )}

          {isClaimedByMe && (
            <>
              <DropdownMenuItem
                onClick={() => {
                  setApprovalAction("approve");
                  setShowApprovalModal(true);
                }}
                disabled={isProcessing}
              >
                <CheckCircle className="mr-2 h-4 w-4 text-green-600" />
                <span>Approve</span>
              </DropdownMenuItem>

              <DropdownMenuItem
                onClick={() => {
                  setApprovalAction("reject");
                  setShowApprovalModal(true);
                }}
                disabled={isProcessing}
              >
                <XCircle className="mr-2 h-4 w-4 text-red-600" />
                <span>Reject</span>
              </DropdownMenuItem>

              <DropdownMenuSeparator />

              <DropdownMenuItem onClick={handleUnclaim} disabled={isProcessing}>
                <User className="mr-2 h-4 w-4" />
                <span>Release Claim</span>
              </DropdownMenuItem>
            </>
          )}

          {isClaimedByOther && (
            <DropdownMenuItem disabled>
              <User className="mr-2 h-4 w-4" />
              <span>Claimed by {task.claimerName || "Other"}</span>
            </DropdownMenuItem>
          )}

          {isPending && !canUserClaim && (
            <DropdownMenuItem disabled>
              <AlertTriangle className="mr-2 h-4 w-4 text-amber-500" />
              <span>Requires {task.assignedRole} role</span>
            </DropdownMenuItem>
          )}

          <DropdownMenuSeparator />

          <DropdownMenuItem>
            <Eye className="mr-2 h-4 w-4" />
            <span>View Workflow History</span>
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>

      {/* Modals */}
      <ClaimTaskModal
        isOpen={showClaimModal}
        onClose={() => setShowClaimModal(false)}
        onConfirm={handleClaim}
        isLoading={isClaiming}
        taskDetails={{
          entityType: documentType,
          entityId: documentId,
          stageName: task.stageName,
          assignedRole: task.assignedRole,
        }}
      />

      <ApprovalActionModal
        isOpen={showApprovalModal}
        onClose={() => setShowApprovalModal(false)}
        onConfirm={handleApprovalAction}
        isLoading={isApproving || isRejecting}
        action={approvalAction}
        taskDetails={{
          entityType: documentType,
          entityId: documentId,
          stageName: task.stageName,
          claimedBy: "You",
          claimExpiry: task.claimExpiry || "",
        }}
      />
    </>
  );
}
