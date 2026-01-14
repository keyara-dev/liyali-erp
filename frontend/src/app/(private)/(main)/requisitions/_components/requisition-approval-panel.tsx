"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import { SignatureCanvas } from "@/components/ui/signature-canvas";
import { Upload, Send, XCircle, Loader2 } from "lucide-react";
import {
  useApprovalTasks,
  useApproveTask,
  useRejectTask,
} from "@/hooks/use-approval-workflow";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";

interface ApprovalActionPanelProps {
  requisitionId: string;
  onApprovalComplete: () => void;
}

export function ApprovalActionPanel({
  requisitionId,
  onApprovalComplete,
}: ApprovalActionPanelProps) {
  const [action, setAction] = useState<"approve" | "reject" | null>(null);
  const [comments, setComments] = useState("");
  const [remarks, setRemarks] = useState("");
  const [signature, setSignature] = useState("");
  const [showAttachmentDialog, setShowAttachmentDialog] = useState(false);

  // Fetch approval tasks for this requisition
  const { data: approvalData } = useApprovalTasks(
    { documentType: "REQUISITION", assignedToMe: true },
    1,
    100
  );

  const approvalTasks = approvalData?.data || [];

  // Find the approval task for this requisition
  const task = approvalTasks?.find(
    (t) =>
      t.documentId === requisitionId ||
      t.entityId === requisitionId ||
      (t as any).requisitionId === requisitionId
  );
  const taskId = task?.id || "";

  const approveMutation = useApproveTask(taskId);
  const rejectMutation = useRejectTask(taskId);

  const handleApprove = async () => {
    if (!signature || !taskId) {
      return;
    }

    try {
      await approveMutation.mutateAsync({
        comments,
        signature,
        stageNumber: task?.stage || (task as any)?.stageIndex || 1,
      });

      // Reset form and call completion callback
      setComments("");
      setRemarks("");
      setSignature("");
      setAction(null);
      onApprovalComplete();
    } catch (error) {
      // Error handled by hook's onError callback
    }
  };

  const handleReject = async () => {
    if (!remarks.trim() || !taskId) {
      return;
    }

    try {
      await rejectMutation.mutateAsync({
        remarks,
        comments: remarks,
        signature,
      });

      // Reset form and call completion callback
      setComments("");
      setRemarks("");
      setSignature("");
      setAction(null);
      onApprovalComplete();
    } catch (error) {
      // Error handled by hook's onError callback
    }
  };

  const isLoading = approveMutation.isPending || rejectMutation.isPending;

  if (action === null) {
    return (
      <div className="space-y-3">
        <h3 className="font-semibold text-sm">Action Required</h3>
        <div className="grid grid-cols-2 gap-2">
          <Button
            onClick={() => setAction("approve")}
            disabled={isLoading || !task}
            className="bg-green-600 hover:bg-green-700 gap-2"
          >
            <Send className="h-4 w-4" />
            Approve
          </Button>
          <Button
            onClick={() => setAction("reject")}
            disabled={isLoading || !task}
            variant="destructive"
            className="gap-2"
          >
            <XCircle className="h-4 w-4" />
            Reject
          </Button>
        </div>
        {!task && (
          <p className="text-xs text-gray-500 text-center">
            No pending approval task found for this requisition
          </p>
        )}
      </div>
    );
  }

  return (
    <div className="space-y-4 p-4 bg-blue-50 rounded-lg border border-blue-200">
      <div>
        <h3 className="font-semibold mb-2">
          {action === "approve" ? "Approve Requisition" : "Reject Requisition"}
        </h3>
        <p className="text-sm text-gray-600 mb-4">
          {action === "approve"
            ? "Add a signature and optional comments to approve"
            : "Provide remarks explaining the rejection reason"}
        </p>
      </div>

      {action === "approve" ? (
        <>
          <div className="space-y-2">
            <Label htmlFor="comments">Comments (Optional)</Label>
            <Textarea
              id="comments"
              placeholder="Add any approval comments or recommendations..."
              value={comments}
              onChange={(e) => setComments(e.target.value)}
              rows={3}
              className="resize-none"
              disabled={isLoading}
            />
          </div>

          <SignatureCanvas
            onSignatureChange={setSignature}
            disabled={isLoading}
          />
        </>
      ) : (
        <div className="space-y-2">
          <Label htmlFor="remarks">Remarks *</Label>
          <Textarea
            id="remarks"
            placeholder="Required: Please explain why this requisition is being rejected..."
            value={remarks}
            onChange={(e) => setRemarks(e.target.value)}
            rows={4}
            className="resize-none"
            disabled={isLoading}
          />
          <p className="text-xs text-muted-foreground">
            Detailed remarks are required for rejection to help the requester
            understand the issues
          </p>
        </div>
      )}

      <Button
        variant="outline"
        size="sm"
        onClick={() => setShowAttachmentDialog(true)}
        className="gap-2 w-full text-gray-700"
        disabled={isLoading}
      >
        <Upload className="h-4 w-4" />
        Add Supporting Documents
      </Button>

      <div className="flex gap-2">
        <Button
          onClick={action === "approve" ? handleApprove : handleReject}
          disabled={
            isLoading ||
            (action === "reject" && !remarks.trim()) ||
            (action === "approve" && !signature)
          }
          className={
            action === "approve"
              ? "bg-green-600 hover:bg-green-700 flex-1"
              : "bg-red-600 hover:bg-red-700 flex-1"
          }
        >
          {isLoading ? (
            <>
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              Processing...
            </>
          ) : action === "approve" ? (
            "Confirm Approval"
          ) : (
            "Confirm Rejection"
          )}
        </Button>
        <Button
          variant="outline"
          onClick={() => {
            setAction(null);
            setComments("");
            setRemarks("");
            setSignature("");
          }}
          disabled={isLoading}
        >
          Cancel
        </Button>
      </div>

      {/* Attachment Dialog */}
      <Dialog
        open={showAttachmentDialog}
        onOpenChange={setShowAttachmentDialog}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Add Supporting Documents</DialogTitle>
            <DialogDescription>
              Upload documents to support your approval or rejection decision
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div className="border-2 border-dashed rounded-lg p-6 text-center">
              <Upload className="h-8 w-8 text-gray-400 mx-auto mb-2" />
              <p className="text-sm text-gray-600">
                Click or drag files here to upload
              </p>
              <p className="text-xs text-gray-500 mt-1">
                PDF, DOC, XLS up to 10MB
              </p>
            </div>
            <Button
              onClick={() => setShowAttachmentDialog(false)}
              className="w-full"
            >
              Continue
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  );
}
