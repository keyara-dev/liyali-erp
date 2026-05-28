"use client";

import { useCallback, useState } from "react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { PaymentVoucher } from "@/types/payment-voucher";
import type { Workflow } from "@/types/workflow-config";
import { Send, CheckCircle2, AlertCircle } from "lucide-react";
import { Separator } from "@/components/ui/separator";
import { WorkflowSelector } from "@/components/workflows/workflow-selector";
import { WorkflowRequirementBanner } from "@/components/ui/workflow-requirement-banner";
import { formatCurrency } from "@/lib/utils";

interface PaymentVoucherSubmitDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  paymentVoucher: PaymentVoucher;
  onSubmit: (workflowId: string, comments?: string) => Promise<void>;
  isSubmitting: boolean;
}

export function PaymentVoucherSubmitDialog({
  open,
  onOpenChange,
  paymentVoucher,
  onSubmit,
  isSubmitting,
}: PaymentVoucherSubmitDialogProps) {
  const [comments, setComments] = useState("");
  const [workflowId, setWorkflowId] = useState("");
  const [workflowError, setWorkflowError] = useState<string | null>(null);

  const hasItems = paymentVoucher.items && paymentVoucher.items.length > 0;
  const hasVendor = !!paymentVoucher.vendorId || !!paymentVoucher.vendorName;
  const hasInvoiceNumber = !!paymentVoucher.invoiceNumber;
  const hasValidAmount =
    paymentVoucher.totalAmount > 0 || paymentVoucher.amount > 0;
  const canSubmit =
    hasItems && hasVendor && hasInvoiceNumber && hasValidAmount && workflowId;

  const handleWorkflowSelect = useCallback(
    (_workflow: Workflow | null) => {},
    [],
  );

  const handleSubmit = async () => {
    if (!workflowId) {
      setWorkflowError("Please select a workflow");
      return;
    }
    if (!canSubmit) return;
    setWorkflowError(null);
    await onSubmit(workflowId, comments);
  };

  const handleClose = () => {
    if (!isSubmitting) {
      setComments("");
      setWorkflowId("");
      setWorkflowError(null);
      onOpenChange(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={handleClose}>
      <DialogContent
        className="max-w-lg max-h-[90svh] flex flex-col p-0 overflow-hidden"
        onInteractOutside={(e) => e.preventDefault()}
      >
        {/* Header */}
        <DialogHeader className="shrink-0 px-6 pt-6 pb-4 border-b">
          <DialogTitle className="flex items-center gap-2">
            <Send className="h-5 w-5" />
            Submit Payment Voucher for Approval
          </DialogTitle>
          <DialogDescription>
            Select an approval workflow before submitting.
          </DialogDescription>
        </DialogHeader>

        {/* Scrollable body */}
        <div className="flex-1 overflow-y-auto px-6 py-4 space-y-4">
          <WorkflowRequirementBanner entityType="payment_voucher" />

          <WorkflowSelector
            entityType="payment_voucher"
            value={workflowId}
            onChange={setWorkflowId}
            onWorkflowSelect={handleWorkflowSelect}
            disabled={isSubmitting}
            required
            error={workflowError || undefined}
            showDetails={true}
          />

          <Separator />

          <div className="space-y-3 rounded-lg border p-4 bg-muted/50">
            <div className="flex justify-between items-center">
              <span className="text-sm font-medium">Document Number:</span>
              <span className="text-sm font-mono">
                {paymentVoucher.documentNumber}
              </span>
            </div>
            <div className="flex justify-between items-center">
              <span className="text-sm font-medium">Title:</span>
              <span className="text-sm">{paymentVoucher.title}</span>
            </div>
            <div className="flex justify-between items-center">
              <span className="text-sm font-medium">Vendor:</span>
              <span className="text-sm">{paymentVoucher.vendorName}</span>
            </div>
            <div className="flex justify-between items-center">
              <span className="text-sm font-medium">Invoice Number:</span>
              <span className="text-sm font-mono">
                {paymentVoucher.invoiceNumber || "N/A"}
              </span>
            </div>
            <div className="flex justify-between items-center">
              <span className="text-sm font-medium">Department:</span>
              <span className="text-sm">{paymentVoucher.department}</span>
            </div>
            <Separator />
            <div className="flex justify-between items-center">
              <span className="text-sm font-medium">Total Amount:</span>
              <span className="text-sm font-mono text-primary">
                {formatCurrency(
                  paymentVoucher.totalAmount || paymentVoucher.amount,
                  paymentVoucher.currency,
                )}
              </span>
            </div>
            <Separator />
            <div className="flex justify-between items-center">
              <span className="text-sm font-medium">Items:</span>
              <span className="text-sm">
                {paymentVoucher.items?.length || 0} item
                {paymentVoucher.items?.length !== 1 ? "s" : ""}
              </span>
            </div>
          </div>

          {!hasItems && (
            <Alert variant="destructive">
              <AlertCircle className="h-4 w-4" />
              <AlertDescription>
                You must add at least one item before submitting.
              </AlertDescription>
            </Alert>
          )}
          {!hasVendor && (
            <Alert variant="destructive">
              <AlertCircle className="h-4 w-4" />
              <AlertDescription>
                You must select a vendor before submitting.
              </AlertDescription>
            </Alert>
          )}
          {!hasInvoiceNumber && (
            <Alert variant="destructive">
              <AlertCircle className="h-4 w-4" />
              <AlertDescription>
                You must provide an invoice number before submitting.
              </AlertDescription>
            </Alert>
          )}
          {!hasValidAmount && (
            <Alert variant="destructive">
              <AlertCircle className="h-4 w-4" />
              <AlertDescription>
                Payment amount must be greater than zero.
              </AlertDescription>
            </Alert>
          )}
          {canSubmit && (
            <Alert>
              <CheckCircle2 className="h-4 w-4" />
              <AlertDescription>
                Payment voucher is ready for submission.
              </AlertDescription>
            </Alert>
          )}

          <Textarea
            id="comments"
            label="Comments (Optional)"
            placeholder="Add any comments or notes for the approvers..."
            value={comments}
            onChange={(e) => setComments(e.target.value)}
            disabled={isSubmitting}
            rows={4}
          />
        </div>

        {/* Sticky footer */}
        <div className="shrink-0 border-t bg-background px-6 py-4 flex flex-col-reverse gap-3 sm:flex-row sm:justify-end">
          <Button
            type="button"
            variant="outline"
            onClick={handleClose}
            disabled={isSubmitting}
            className="w-full sm:w-auto"
          >
            Cancel
          </Button>
          <Button
            onClick={handleSubmit}
            disabled={isSubmitting || !canSubmit}
            isLoading={isSubmitting}
            loadingText="Submitting..."
            className="w-full sm:w-auto"
          >
            <Send className="mr-2 h-4 w-4" />
            Submit for Approval
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}
