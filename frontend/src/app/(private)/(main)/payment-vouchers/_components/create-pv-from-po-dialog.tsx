"use client";

import { useState, useMemo } from "react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Label } from "@/components/ui/label";
import { SelectField } from "@/components/ui/select-field";
import { PurchaseOrder } from "@/types/purchase-order";
import {
  FileText,
  CheckCircle2,
  AlertCircle,
  Package,
  Truck,
  Wallet,
} from "lucide-react";
import { Separator } from "@/components/ui/separator";
import { useConfigurationStatus } from "@/hooks/use-configuration-status";
import { ConfigurationChecklistBanner } from "@/components/ui/configuration-checklist-banner";
import { useVendors } from "@/hooks/use-vendor-queries";
import { useGRNs } from "@/hooks/use-grn-queries";
import { useOrganizationSettingsQuery } from "@/hooks/use-organization-queries";
import { formatCurrency } from "@/lib/utils";

interface CreatePVFromPODialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  purchaseOrder: PurchaseOrder;
  onConfirm: (
    workflowId: string,
    vendorId?: string,
    vendorName?: string,
    linkedGRNDocumentNumber?: string,
  ) => Promise<void>;
  isCreating: boolean;
}

export function CreatePVFromPODialog({
  open,
  onOpenChange,
  purchaseOrder,
  onConfirm,
  isCreating,
}: CreatePVFromPODialogProps) {
  const [selectedVendorId, setSelectedVendorId] = useState(
    purchaseOrder.vendorId ?? "",
  );
  const [selectedVendorName, setSelectedVendorName] = useState(
    purchaseOrder.vendorName ?? "",
  );
  const [selectedGRNDocNumber, setSelectedGRNDocNumber] = useState("");

  const { data: vendors = [] } = useVendors();
  const { data: orgSettings } = useOrganizationSettingsQuery();

  // Resolve effective procurement flow: PO override → org default → "goods_first"
  const effectiveFlow = useMemo(() => {
    if (purchaseOrder.procurementFlow) {
      return purchaseOrder.procurementFlow;
    }
    return orgSettings?.procurementFlow ?? "goods_first";
  }, [purchaseOrder.procurementFlow, orgSettings?.procurementFlow]);

  const isGoodsFirst = effectiveFlow === "goods_first";

  // Fetch approved GRNs for this PO (only needed for goods_first)
  const { data: grns = [] } = useGRNs(1, 50, {
    status: "APPROVED",
    poDocumentNumber: purchaseOrder.documentNumber,
  });

  // Configuration check — workflow is picked at submit time, so skip it here
  const configStatus = useConfigurationStatus({
    includeWorkflow: false,
  });

  const canCreate =
    purchaseOrder.status?.toUpperCase() === "APPROVED" &&
    configStatus.allConfigured &&
    (!isGoodsFirst || selectedGRNDocNumber !== "");

  const handleConfirm = async () => {
    if (isGoodsFirst && !selectedGRNDocNumber) {
      return; // GRN selection is enforced by canCreate
    }

    if (!canCreate) return;

    // Pass empty workflowId — it's stored but unused at creation (see
    // payment-vouchers.ts); the submit dialog captures the real workflow.
    await onConfirm(
      "",
      selectedVendorId || undefined,
      selectedVendorName || undefined,
      isGoodsFirst ? selectedGRNDocNumber : undefined,
    );
    setSelectedVendorId(purchaseOrder.vendorId ?? "");
    setSelectedVendorName(purchaseOrder.vendorName ?? "");
    setSelectedGRNDocNumber("");
  };

  const handleClose = () => {
    if (!isCreating) {
      setSelectedVendorId(purchaseOrder.vendorId ?? "");
      setSelectedVendorName(purchaseOrder.vendorName ?? "");
      setSelectedGRNDocNumber("");
      onOpenChange(false);
    }
  };

  const handleVendorChange = (value: string) => {
    const nextId = value === "__none__" ? "" : value;
    setSelectedVendorId(nextId);
    if (nextId === "") {
      setSelectedVendorName("");
    } else {
      const vendor = vendors.find((v) => v.id === nextId);
      setSelectedVendorName(vendor?.name ?? "");
    }
  };

  const selectedGRN = grns.find(
    (g) => g.documentNumber === selectedGRNDocNumber,
  );

  return (
    <Dialog open={open} onOpenChange={handleClose}>
      <DialogContent
        className="max-w-2xl! p-0 flex flex-col h-[90svh] max-h-[90vh] overflow-hidden"
        onInteractOutside={(e) => e.preventDefault()}
      >
        <DialogHeader className="p-4 pb-3 shrink-0 border-b">
          <DialogTitle className="flex items-center gap-2">
            <FileText className="h-5 w-5" />
            Create Payment Voucher
          </DialogTitle>
          <DialogDescription>
            Create a payment voucher from this approved purchase order. The PV
            starts as a draft — you'll pick the approval workflow when
            submitting it.
          </DialogDescription>
        </DialogHeader>

        <div className="flex-1 overflow-y-auto p-4 space-y-4 min-w-0">
          {/* Configuration Checklist Banner */}
          {!configStatus.allConfigured && (
            <ConfigurationChecklistBanner
              requirements={configStatus.requirements}
              title="Configuration Required"
              description="Complete the following configurations before creating a payment voucher:"
            />
          )}

          {/* Procurement flow banner */}
          {(() => {
            const FlowIcon = isGoodsFirst ? Truck : Wallet;
            const isOverride = !!purchaseOrder.procurementFlow;
            return (
              <div
                className={`flex items-start gap-3 rounded-lg border p-3 ${
                  isGoodsFirst
                    ? "border-blue-200 bg-blue-50 dark:border-blue-800 dark:bg-blue-950/30"
                    : "border-amber-200 bg-amber-50 dark:border-amber-800 dark:bg-amber-950/30"
                }`}
              >
                <div
                  className={`rounded-md p-1.5 shrink-0 ${
                    isGoodsFirst
                      ? "bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-200"
                      : "bg-amber-100 text-amber-700 dark:bg-amber-900 dark:text-amber-200"
                  }`}
                >
                  <FlowIcon className="h-4 w-4" />
                </div>
                <div className="flex-1 min-w-0">
                  <div className="flex flex-wrap items-center gap-2">
                    <span className="text-sm font-semibold">
                      {isGoodsFirst ? "Goods-First" : "Payment-First"}
                    </span>
                    <span className="text-xs text-muted-foreground">
                      {isOverride ? "PO override" : "Organization default"}
                    </span>
                  </div>
                  <p className="text-xs text-muted-foreground mt-0.5">
                    {isGoodsFirst
                      ? "An approved GRN is required before this payment voucher can proceed."
                      : "Payment is processed upfront — the GRN confirms delivery later."}
                  </p>
                </div>
              </div>
            );
          })()}

          {/* Vendor Selector (optional) */}
          <SelectField
            label="Vendor"
            placeholder="No vendor (optional)"
            value={selectedVendorId || "__none__"}
            onValueChange={handleVendorChange}
            isDisabled={isCreating}
            options={[
              { value: "__none__", label: "No vendor" },
              ...vendors
                .filter((v) => v.active)
                .map((v) => ({ value: v.id, label: v.name })),
            ]}
          />

          {/* GRN Selector — required for goods_first flow */}
          {isGoodsFirst &&
            (grns.length === 0 ? (
              <div className="space-y-1.5">
                <Label className="flex items-center gap-1.5">
                  <Package className="h-4 w-4" />
                  Linked GRN <span className="text-destructive">*</span>
                </Label>
                <Alert variant="destructive">
                  <AlertCircle className="h-4 w-4" />
                  <AlertDescription>
                    No approved GRNs found for PO {purchaseOrder.documentNumber}
                    . Goods must be received and the GRN approved before
                    creating a payment voucher.
                  </AlertDescription>
                </Alert>
              </div>
            ) : (
              <div className="space-y-2">
                <SelectField
                  label="Linked GRN"
                  required
                  placeholder="Select approved GRN"
                  descriptionText="Goods-first flow requires an approved GRN for this PO before payment can be processed."
                  value={selectedGRNDocNumber}
                  onValueChange={setSelectedGRNDocNumber}
                  isDisabled={isCreating}
                  options={grns.map((grn) => ({
                    value: grn.documentNumber,
                    label: `${grn.documentNumber} — received ${new Date(
                      grn.receivedDate,
                    ).toLocaleDateString("en-ZM")}`,
                  }))}
                />
                {selectedGRN && (
                  <div className="rounded-md border bg-muted/50 p-3 text-sm space-y-1">
                    <div className="flex justify-between">
                      <span className="text-muted-foreground">
                        Received by:
                      </span>
                      <span>{selectedGRN.receivedBy}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-muted-foreground">Items:</span>
                      <span>{selectedGRN.items?.length ?? 0}</span>
                    </div>
                    {selectedGRN.warehouseLocation && (
                      <div className="flex justify-between">
                        <span className="text-muted-foreground">Location:</span>
                        <span>{selectedGRN.warehouseLocation}</span>
                      </div>
                    )}
                  </div>
                )}
              </div>
            ))}

          <Separator />

          {/* Purchase Order Summary */}
          <div className="space-y-3 rounded-lg border p-4 bg-muted/50">
            <div className="flex items-center justify-between mb-2">
              <h4 className="text-sm font-semibold">Source Purchase Order</h4>
              <span className="text-xs px-2 py-1 rounded-full bg-green-100 text-green-800 border border-green-200">
                Approved
              </span>
            </div>

            <div className="flex justify-between items-center">
              <span className="text-sm font-medium">PO Number:</span>
              <span className="text-sm font-mono">
                {purchaseOrder.documentNumber}
              </span>
            </div>
            {purchaseOrder.vendorName && (
              <div className="flex justify-between items-center">
                <span className="text-sm font-medium">PO Vendor:</span>
                <span className="text-sm">{purchaseOrder.vendorName}</span>
              </div>
            )}
            <div className="flex justify-between items-center">
              <span className="text-sm font-medium">Department:</span>
              <span className="text-sm">{purchaseOrder.department}</span>
            </div>

            <Separator />

            <div className="flex justify-between items-center">
              <span className="text-sm font-medium">Total Amount:</span>
              <span className="text-sm font-mono text-blue-600">
                {formatCurrency(
                  purchaseOrder.totalAmount,
                  purchaseOrder.currency,
                )}
              </span>
            </div>

            <div className="flex justify-between items-center">
              <span className="text-sm font-medium">Items:</span>
              <span className="text-sm">
                {purchaseOrder.items?.length || 0} item
                {purchaseOrder.items?.length !== 1 ? "s" : ""}
              </span>
            </div>

            {purchaseOrder.deliveryDate && (
              <div className="flex justify-between items-center">
                <span className="text-sm font-medium">Delivery Date:</span>
                <span className="text-sm">
                  {new Date(purchaseOrder.deliveryDate).toLocaleDateString(
                    "en-ZM",
                    {
                      year: "numeric",
                      month: "short",
                      day: "numeric",
                    },
                  )}
                </span>
              </div>
            )}
          </div>

          {/* Info Alert */}
          {canCreate && (
            <Alert>
              <CheckCircle2 className="h-4 w-4" />
              <AlertDescription>
                A new payment voucher will be created with the selected
                workflow. The PV will be in draft status and can be edited
                before submission.
              </AlertDescription>
            </Alert>
          )}

          {purchaseOrder.status?.toUpperCase() !== "APPROVED" && (
            <Alert variant="destructive">
              <AlertCircle className="h-4 w-4" />
              <AlertDescription>
                Only approved purchase orders can be converted to payment
                vouchers.
              </AlertDescription>
            </Alert>
          )}
        </div>

        {/* Sticky Footer */}
        <div className="shrink-0 border-t bg-card/5 backdrop-blur-xs flex flex-col-reverse sm:flex-row justify-end gap-2 p-4">
          <Button
            type="button"
            variant="outline"
            onClick={handleClose}
            disabled={isCreating}
          >
            Cancel
          </Button>
          <Button
            onClick={handleConfirm}
            disabled={isCreating || !canCreate}
            isLoading={isCreating}
            loadingText="Creating..."
          >
            <FileText className="mr-2 h-4 w-4" />
            Create Payment Voucher
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}
