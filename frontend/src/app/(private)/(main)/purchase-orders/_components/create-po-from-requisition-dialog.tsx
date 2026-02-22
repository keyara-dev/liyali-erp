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
import { Label } from "@/components/ui/label";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Requisition } from "@/types/requisition";
import { Loader2, FileText, CheckCircle2, AlertCircle } from "lucide-react";
import { Separator } from "@/components/ui/separator";
import { WorkflowSelector } from "@/components/workflows/workflow-selector";
import { useConfigurationStatus } from "@/hooks/use-configuration-status";
import { ConfigurationChecklistBanner } from "@/components/ui/configuration-checklist-banner";

interface CreatePOFromRequisitionDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  requisition: Requisition;
  onConfirm: (workflowId: string) => Promise<void>;
  isCreating: boolean;
}

export function CreatePOFromRequisitionDialog({
  open,
  onOpenChange,
  requisition,
  onConfirm,
  isCreating,
}: CreatePOFromRequisitionDialogProps) {
  const [workflowId, setWorkflowId] = useState("");
  const [workflowError, setWorkflowError] = useState<string | null>(null);

  // Check configuration status
  const configStatus = useConfigurationStatus({
    includeWorkflow: true,
    workflowEntityType: "purchase_order",
  });

  const canCreate =
    workflowId &&
    requisition.status === "approved" &&
    configStatus.allConfigured;

  const handleConfirm = async () => {
    // Validate workflow selection
    if (!workflowId) {
      setWorkflowError("Please select a workflow");
      return;
    }

    if (!canCreate) return;

    setWorkflowError(null);
    await onConfirm(workflowId);
    setWorkflowId("");
  };

  const handleClose = () => {
    if (!isCreating) {
      setWorkflowId("");
      setWorkflowError(null);
      onOpenChange(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={handleClose}>
      <DialogContent className="max-w-lg max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <FileText className="h-5 w-5" />
            Create Purchase Order
          </DialogTitle>
          <DialogDescription>
            Select an approval workflow for the new purchase order. The PO will
            be created from the approved requisition below.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4">
          {/* Configuration Checklist Banner */}
          {!configStatus.allConfigured && (
            <ConfigurationChecklistBanner
              requirements={configStatus.requirements}
              isLoading={configStatus.isLoading}
              title="Configuration Required"
              description="Complete the following configurations before creating a purchase order:"
            />
          )}

          {/* Workflow Selector */}
          <WorkflowSelector
            entityType="purchase_order"
            value={workflowId}
            onChange={setWorkflowId}
            disabled={isCreating}
            required
            error={workflowError || undefined}
            showDetails={true}
          />

          <Separator />

          {/* Requisition Summary */}
          <div className="space-y-3 rounded-lg border p-4 bg-muted/50">
            <div className="flex items-center justify-between mb-2">
              <h4 className="text-sm font-semibold">Source Requisition</h4>
              <span className="text-xs px-2 py-1 rounded-full bg-green-100 text-green-800 border border-green-200">
                Approved
              </span>
            </div>

            <div className="flex justify-between items-center">
              <span className="text-sm font-medium">Document Number:</span>
              <span className="text-sm font-mono">
                {requisition.documentNumber}
              </span>
            </div>
            <div className="flex justify-between items-center">
              <span className="text-sm font-medium">Title:</span>
              <span className="text-sm">{requisition.title}</span>
            </div>
            <div className="flex justify-between items-center">
              <span className="text-sm font-medium">Department:</span>
              <span className="text-sm">{requisition.department}</span>
            </div>

            <Separator />

            <div className="flex justify-between items-center">
              <span className="text-sm font-medium">Total Amount:</span>
              <span className="text-sm font-mono text-blue-600">
                {requisition.currency}{" "}
                {requisition.totalAmount?.toLocaleString("en-ZM", {
                  minimumFractionDigits: 2,
                  maximumFractionDigits: 2,
                }) || "0.00"}
              </span>
            </div>

            <div className="flex justify-between items-center">
              <span className="text-sm font-medium">Items:</span>
              <span className="text-sm">
                {requisition.items?.length || 0} item
                {requisition.items?.length !== 1 ? "s" : ""}
              </span>
            </div>

            {requisition.vendorName && (
              <div className="flex justify-between items-center">
                <span className="text-sm font-medium">Vendor:</span>
                <span className="text-sm">{requisition.vendorName}</span>
              </div>
            )}
          </div>

          {/* Info Alert */}
          {canCreate && (
            <Alert>
              <CheckCircle2 className="h-4 w-4" />
              <AlertDescription>
                A new purchase order will be created with the selected workflow.
                The PO will be in draft status and can be edited before
                submission.
              </AlertDescription>
            </Alert>
          )}

          {requisition.status !== "approved" && (
            <Alert variant="destructive">
              <AlertCircle className="h-4 w-4" />
              <AlertDescription>
                Only approved requisitions can be converted to purchase orders.
              </AlertDescription>
            </Alert>
          )}
        </div>

        {/* Actions */}
        <DialogFooter className="gap-2 sm:gap-0">
          <Button
            type="button"
            variant="outline"
            onClick={handleClose}
            disabled={isCreating}
          >
            Cancel
          </Button>
          <Button onClick={handleConfirm} disabled={isCreating || !canCreate}>
            {isCreating ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Creating...
              </>
            ) : (
              <>
                <FileText className="mr-2 h-4 w-4" />
                Create Purchase Order
              </>
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
