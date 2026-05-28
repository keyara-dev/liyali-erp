"use client";

import { useState } from "react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";

interface ApprovalConfirmationDialogProps {
  open: boolean;
  documentId: string;
  documentType: "GRN" | "PAYMENT_VOUCHER" | "PURCHASE_ORDER" | "REQUISITION";
  documentNumber: string;
  vendor: string;
  amount: string;
  stageNumber: number;
  totalStages: number;
  stageName: string;
  onApprove: (data: any) => Promise<void>;
  onCancel: () => void;
}

export function ApprovalConfirmationDialog({
  open,
  documentId,
  documentType,
  documentNumber,
  vendor,
  amount,
  stageNumber,
  totalStages,
  stageName,
  onApprove,
  onCancel,
}: ApprovalConfirmationDialogProps) {
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [approvalComments, setApprovalComments] = useState("");

  const handleSubmit = async () => {
    setIsSubmitting(true);
    try {
      await onApprove({
        documentId,
        documentType,
        comments: approvalComments,
      });
      setApprovalComments("");
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onCancel}>
      <DialogContent className="sm:max-w-[500px] max-h-[90svh] flex flex-col p-0 overflow-hidden">
        <DialogHeader className="shrink-0 px-6 pt-6 pb-4 border-b">
          <DialogTitle>Approve {documentType}</DialogTitle>
          <DialogDescription>
            Stage {stageNumber} of {totalStages}: {stageName}
          </DialogDescription>
        </DialogHeader>

        <div className="flex-1 overflow-y-auto px-6 py-4 space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label className="text-xs text-muted-foreground">Document</Label>
              <p className="text-sm font-medium">{documentNumber}</p>
            </div>
            <div className="space-y-2">
              <Label className="text-xs text-muted-foreground">Amount</Label>
              <p className="text-sm font-medium">{amount}</p>
            </div>
            <div className="space-y-2 col-span-2">
              <Label className="text-xs text-muted-foreground">Vendor</Label>
              <p className="text-sm font-medium">{vendor}</p>
            </div>
          </div>

          <Textarea
            label="Comments (Optional)"
            id="comments"
            placeholder="Add any approval comments..."
            value={approvalComments}
            onChange={(e) => setApprovalComments(e.target.value)}
            className="h-24"
          />
        </div>

        <div className="shrink-0 border-t bg-background px-6 py-4 flex flex-col-reverse gap-3 sm:flex-row sm:justify-end">
          <Button
            variant="outline"
            onClick={onCancel}
            disabled={isSubmitting}
            className="w-full sm:w-auto"
          >
            Cancel
          </Button>
          <Button
            onClick={handleSubmit}
            disabled={isSubmitting}
            isLoading={isSubmitting}
            loadingText="Approving..."
            className="w-full sm:w-auto"
          >
            Approve
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}
