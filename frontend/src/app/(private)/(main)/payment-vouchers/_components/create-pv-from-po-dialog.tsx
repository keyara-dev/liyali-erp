"use client";

import { useState, useMemo } from "react";
import Link from "next/link";
import { ResponsiveSheet } from "@/components/ui/responsive-sheet";
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

  // Fetch all GRNs for this PO so we can distinguish between "no GRN yet"
  // and "GRN exists but not yet approved" — the empty-state message changes.
  const { data: allGRNs = [] } = useGRNs(1, 50, {
    poDocumentNumber: purchaseOrder.documentNumber,
  });
  const grns = useMemo(
    () =>
      allGRNs.filter((g: any) => g.status?.toUpperCase() === "APPROVED"),
    [allGRNs],
  );
  const pendingGRNs = useMemo(
    () =>
      allGRNs.filter(
        (g: any) =>
          g.status?.toUpperCase() !== "APPROVED" &&
          g.status?.toUpperCase() !== "REJECTED" &&
          g.status?.toUpperCase() !== "CANCELLED",
      ),
    [allGRNs],
  );

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

  const footerContent = (
    <div className="flex flex-col-reverse sm:flex-row justify-end gap-2 w-full">
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
  );

  return (
    <ResponsiveSheet
      open={open}
      onOpenChange={handleClose}
      title={
        <span className="flex items-center gap-2">
          <FileText className="h-5 w-5" />
          Create Payment Voucher
        </span>
      }
      description="Create a payment voucher from this approved purchase order. The PV starts as a draft — you'll pick the approval workflow when submitting it."
      desktopMaxWidth="sm:max-w-2xl"
      dismissibleOnOutsideClick={false}
      footer={footerContent}
    >
        <div className="space-y-4 min-w-0">
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
                {pendingGRNs.length > 0 ? (
                  <Alert className="border-amber-300 bg-amber-50 dark:border-amber-700 dark:bg-amber-950/40">
                    <AlertCircle className="h-4 w-4 text-amber-600 dark:text-amber-400" />
                    <AlertDescription className="space-y-2 text-amber-800 dark:text-amber-200">
                      <p>
                        A GRN exists for this PO but hasn&apos;t been approved
                        yet. Submit and approve it first before creating a
                        payment voucher.
                      </p>
                      <div className="flex flex-col gap-1">
                        {pendingGRNs.map((g: any) => (
                          <Link
                            key={g.id}
                            href={`/grn/${g.id}`}
                            className="text-xs font-mono text-amber-900 dark:text-amber-100 underline underline-offset-2 hover:no-underline"
                          >
                            {g.documentNumber} ({g.status})
                          </Link>
                        ))}
                      </div>
                    </AlertDescription>
                  </Alert>
                ) : (
                  <Alert variant="destructive">
                    <AlertCircle className="h-4 w-4" />
                    <AlertDescription>
                      No GRNs found for PO {purchaseOrder.documentNumber}.
                      Goods must be received and the GRN approved before
                      creating a payment voucher.
                    </AlertDescription>
                  </Alert>
                )}
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
    </ResponsiveSheet>
  );
}
