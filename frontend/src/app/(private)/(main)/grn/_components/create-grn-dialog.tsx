"use client";

import { useState, useMemo, useEffect } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { QUERY_KEYS } from "@/lib/constants";
import { ResponsiveSheet } from "@/components/ui/responsive-sheet";
import { Button } from "@/components/ui/button";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { SelectField } from "@/components/ui/select-field";
import { Separator } from "@/components/ui/separator";
import {
  Package,
  AlertCircle,
  Plus,
  Trash2,
  Truck,
  Wallet,
} from "lucide-react";
import { toast } from "sonner";
import { useOrganizationSettingsQuery } from "@/hooks/use-organization-queries";
import { useSession } from "@/hooks/use-session";
import { usePurchaseOrders } from "@/hooks/use-purchase-order-queries";
import { usePaymentVouchers } from "@/hooks/use-payment-voucher-queries";
import { createGRNAction } from "@/app/_actions/grn-actions";
import type { GRNItem } from "@/types/goods-received-note";
import type { PurchaseOrder } from "@/types/purchase-order";
import type { PaymentVoucher } from "@/types/payment-voucher";
import { formatCurrency } from "@/lib/utils";
import { Textarea } from "@/components";

interface CreateGRNDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess?: () => void;
  /**
   * Preselect a source document (from the "Ready for GRN" list). Passing one of
   * these forces the matching flow regardless of the org default and prefills
   * the items, so the user lands on a ready-to-confirm form.
   */
  initialPurchaseOrder?: PurchaseOrder;
  initialPaymentVoucher?: PaymentVoucher;
}

interface ItemRow extends GRNItem {
  _key: string;
}

function buildItemsFromPOItems(poItems: PurchaseOrder["items"]): ItemRow[] {
  return (poItems ?? []).map((item, i) => ({
    _key: `item-${i}`,
    description: item.description,
    quantityOrdered: item.quantity,
    quantityReceived: item.quantity,
    variance: 0,
    condition: "good",
    notes: "",
    // Snapshot SKU/catalog code from the matching PO line so the printed
    // GRN can render the "Item Code" column from the standard form.
    itemCode: (item as { itemCode?: string }).itemCode ?? "",
    remarks: "",
  }));
}

export function CreateGRNDialog({
  open,
  onOpenChange,
  onSuccess,
  initialPurchaseOrder,
  initialPaymentVoucher,
}: CreateGRNDialogProps) {
  const { user } = useSession();
  const queryClient = useQueryClient();
  const { data: orgSettings } = useOrganizationSettingsQuery();
  const { data: purchaseOrders = [] } = usePurchaseOrders();
  const { data: paymentVouchers = [] } = usePaymentVouchers();

  const [selectedPOId, setSelectedPOId] = useState("");
  const [selectedPVDocNumber, setSelectedPVDocNumber] = useState("");
  const [items, setItems] = useState<ItemRow[]>([]);
  const [warehouseLocation, setWarehouseLocation] = useState("");
  const [notes, setNotes] = useState("");
  const [consignmentNote, setConsignmentNote] = useState("");
  const [isCreating, setIsCreating] = useState(false);

  const orgFlow = orgSettings?.procurementFlow ?? "goods_first";
  // A row-launched dialog forces the flow that matches its preselected source,
  // overriding the org default (e.g. open a PO in a payment-first org).
  const isSourcePreselected = Boolean(
    initialPurchaseOrder || initialPaymentVoucher,
  );
  const effectiveFlow = initialPaymentVoucher
    ? "payment_first"
    : initialPurchaseOrder
      ? "goods_first"
      : orgFlow;

  // Approved POs for goods_first mode
  const approvedPOs = useMemo(
    () =>
      (purchaseOrders as PurchaseOrder[]).filter(
        (po) => po.status?.toUpperCase() === "APPROVED",
      ),
    [purchaseOrders],
  );

  // Approved / paid PVs for payment_first mode (only PVs linked to a PO)
  const approvedPVs = useMemo(
    () =>
      (paymentVouchers as PaymentVoucher[]).filter(
        (pv) =>
          (pv.status?.toUpperCase() === "APPROVED" ||
            pv.status?.toUpperCase() === "PAID") &&
          pv.linkedPO,
      ),
    [paymentVouchers],
  );

  const selectedPO = useMemo(
    () => approvedPOs.find((po) => po.id === selectedPOId),
    [approvedPOs, selectedPOId],
  );

  const selectedPV = useMemo(
    () => approvedPVs.find((pv) => pv.documentNumber === selectedPVDocNumber),
    [approvedPVs, selectedPVDocNumber],
  );

  // For payment_first, find the PO linked to the selected PV
  const pvLinkedPO = useMemo(() => {
    if (!selectedPV) return undefined;
    return (purchaseOrders as PurchaseOrder[]).find(
      (po) => po.documentNumber === selectedPV.linkedPO,
    );
  }, [selectedPV, purchaseOrders]);

  // The effective PO for item population
  const effectivePO =
    effectiveFlow === "payment_first" ? pvLinkedPO : selectedPO;

  // Preselect + prefill when launched from a "Ready for GRN" row.
  useEffect(() => {
    if (!open) return;
    if (initialPurchaseOrder) {
      setSelectedPOId(initialPurchaseOrder.id);
      setItems(buildItemsFromPOItems(initialPurchaseOrder.items));
    } else if (initialPaymentVoucher) {
      setSelectedPVDocNumber(initialPaymentVoucher.documentNumber);
      const linkedPO = (purchaseOrders as PurchaseOrder[]).find(
        (po) => po.documentNumber === initialPaymentVoucher.linkedPO,
      );
      setItems(buildItemsFromPOItems(linkedPO?.items ?? []));
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open, initialPurchaseOrder?.id, initialPaymentVoucher?.documentNumber]);

  const handlePOSelect = (poId: string) => {
    setSelectedPOId(poId);
    const po = approvedPOs.find((p) => p.id === poId);
    if (po) {
      setItems(buildItemsFromPOItems(po.items));
    } else {
      setItems([]);
    }
  };

  const handlePVSelect = (pvDocNumber: string) => {
    setSelectedPVDocNumber(pvDocNumber);
    const pv = approvedPVs.find((p) => p.documentNumber === pvDocNumber);
    if (pv) {
      const linkedPO = (purchaseOrders as PurchaseOrder[]).find(
        (po) => po.documentNumber === pv.linkedPO,
      );
      setItems(buildItemsFromPOItems(linkedPO?.items ?? []));
    } else {
      setItems([]);
    }
  };

  const updateItem = (
    key: string,
    field: keyof GRNItem,
    value: string | number,
  ) => {
    setItems((prev) =>
      prev.map((item) => {
        if (item._key !== key) return item;
        const updated = { ...item, [field]: value };
        if (field === "quantityReceived" || field === "quantityOrdered") {
          updated.variance =
            Number(updated.quantityReceived) - Number(updated.quantityOrdered);
        }
        return updated;
      }),
    );
  };

  const addItem = () => {
    setItems((prev) => [
      ...prev,
      {
        _key: `item-${Date.now()}`,
        description: "",
        quantityOrdered: 0,
        quantityReceived: 0,
        variance: 0,
        condition: "good",
        notes: "",
        itemCode: "",
        remarks: "",
      },
    ]);
  };

  const removeItem = (key: string) => {
    setItems((prev) => prev.filter((i) => i._key !== key));
  };

  const poDocumentNumber =
    effectiveFlow === "payment_first"
      ? (pvLinkedPO?.documentNumber ?? "")
      : (selectedPO?.documentNumber ?? "");

  const canCreate =
    poDocumentNumber !== "" &&
    warehouseLocation.trim() !== "" &&
    items.length > 0 &&
    items.every(
      (i) =>
        i.description.trim() !== "" &&
        Number(i.quantityOrdered) > 0 &&
        Number(i.quantityReceived) >= 0 &&
        Number(i.quantityReceived) <= Number(i.quantityOrdered),
    ) &&
    (effectiveFlow === "goods_first" || selectedPVDocNumber !== "");

  const handleCreate = async () => {
    if (!canCreate || !user) return;

    setIsCreating(true);
    try {
      const grnItems: GRNItem[] = items.map(({ _key: _, ...item }) => item);
      const response = await createGRNAction(
        poDocumentNumber,
        grnItems,
        user.id,
        warehouseLocation,
        notes,
        effectiveFlow === "payment_first" ? selectedPVDocNumber : undefined,
        consignmentNote,
      );

      if (response.success) {
        toast.success("GRN created successfully");
        // A new GRN affects: GRN lists, the linked PO detail (create-GRN button),
        // the linked PV (goods-first eligibility), and dashboard counters.
        queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.GRN.ALL] });
        queryClient.invalidateQueries({
          queryKey: [QUERY_KEYS.PURCHASE_ORDERS.ALL],
        });
        queryClient.invalidateQueries({
          queryKey: [QUERY_KEYS.PURCHASE_ORDERS.BY_ID],
        });
        queryClient.invalidateQueries({
          queryKey: [QUERY_KEYS.PAYMENT_VOUCHERS.ALL],
        });
        // payment_first: linked PV "GRN received" chip needs to flip — was missing.
        queryClient.invalidateQueries({
          queryKey: [QUERY_KEYS.PAYMENT_VOUCHERS.BY_ID],
        });
        queryClient.invalidateQueries({
          queryKey: [QUERY_KEYS.DASHBOARD.METRICS],
        });
        handleClose();
        onSuccess?.();
      } else {
        toast.error(response.message || "Failed to create GRN");
      }
    } catch (err: any) {
      toast.error(err.message || "Failed to create GRN");
    } finally {
      setIsCreating(false);
    }
  };

  const handleClose = () => {
    if (!isCreating) {
      setSelectedPOId("");
      setSelectedPVDocNumber("");
      setItems([]);
      setWarehouseLocation("");
      setNotes("");
      setConsignmentNote("");
      onOpenChange(false);
    }
  };

  const isGoodsFirst = effectiveFlow === "goods_first";
  const FlowIcon = isGoodsFirst ? Truck : Wallet;

  return (
    <ResponsiveSheet
      open={open}
      onOpenChange={handleClose}
      title={
        <span className="flex items-center gap-2">
          <Package className="h-5 w-5" />
          Create Goods Received Note
        </span>
      }
      description="Record goods received against a purchase order."
      desktopMaxWidth="sm:max-w-2xl"
      dismissibleOnOutsideClick={false}
      footer={
        <div className="flex flex-wrap items-center justify-end gap-2">
          <Button
            type="button"
            variant="outline"
            onClick={handleClose}
            disabled={isCreating}
          >
            Cancel
          </Button>
          <Button
            onClick={handleCreate}
            disabled={isCreating || !canCreate}
            isLoading={isCreating}
            loadingText="Creating..."
          >
            <Package className="mr-2 h-4 w-4" />
            Create GRN
          </Button>
        </div>
      }
    >
        <div className="space-y-4 min-w-0">
          {/* Procurement flow banner */}
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
              <div className="flex items-center gap-2">
                <span className="text-sm font-semibold">
                  {isGoodsFirst ? "Goods-First" : "Payment-First"}
                </span>
                <span className="text-xs text-muted-foreground">
                  {isSourcePreselected ? "Selected source" : "Organization default"}
                </span>
              </div>
              <p className="text-xs text-muted-foreground mt-0.5">
                {isGoodsFirst
                  ? "Record receipt against an approved PO. Payment follows delivery confirmation."
                  : "Delivery follows payment — select the approved PV that funded this receipt."}
              </p>
            </div>
          </div>

          {/* Source document selector */}
          {isGoodsFirst ? (
            approvedPOs.length === 0 ? (
              <div className="space-y-1.5">
                <Label>
                  Purchase Order <span className="text-destructive">*</span>
                </Label>
                <Alert>
                  <AlertCircle className="h-4 w-4" />
                  <AlertDescription>
                    No approved purchase orders available.
                  </AlertDescription>
                </Alert>
              </div>
            ) : (
              <div className="space-y-2">
                <SelectField
                  label="Purchase Order"
                  required
                  placeholder="Select approved PO"
                  value={selectedPOId}
                  onValueChange={handlePOSelect}
                  isDisabled={isCreating}
                  options={approvedPOs.map((po) => ({
                    value: po.id,
                    label: `${po.documentNumber} — ${po.title || po.description || "Untitled"}`,
                  }))}
                />
                {selectedPO && (
                  <div className="rounded-md border bg-muted/50 p-3 text-sm space-y-1">
                    <div className="flex justify-between">
                      <span className="text-muted-foreground">Vendor:</span>
                      <span>{selectedPO.vendorName || "—"}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-muted-foreground">Department:</span>
                      <span>{selectedPO.department}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-muted-foreground">Total:</span>
                      <span className="font-mono text-blue-600">
                        {formatCurrency(
                          selectedPO.totalAmount,
                          selectedPO.currency,
                        )}
                      </span>
                    </div>
                  </div>
                )}
              </div>
            )
          ) : approvedPVs.length === 0 ? (
            <div className="space-y-1.5">
              <Label>
                Payment Voucher <span className="text-destructive">*</span>
              </Label>
              <Alert>
                <AlertCircle className="h-4 w-4" />
                <AlertDescription>
                  No approved or paid payment vouchers available.
                </AlertDescription>
              </Alert>
            </div>
          ) : (
            <div className="space-y-2">
              <SelectField
                label="Payment Voucher (approved / paid)"
                required
                placeholder="Select approved / paid PV"
                descriptionText="Payment-first flow: select the PV that funded this delivery."
                value={selectedPVDocNumber}
                onValueChange={handlePVSelect}
                isDisabled={isCreating}
                options={approvedPVs.map((pv) => ({
                  value: pv.documentNumber,
                  label: `${pv.documentNumber} — PO: ${pv.linkedPO} (${pv.status})`,
                }))}
              />
              {selectedPV && (
                <div className="rounded-md border bg-muted/50 p-3 text-sm space-y-1">
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">Linked PO:</span>
                    <span className="font-mono">{selectedPV.linkedPO}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">Vendor:</span>
                    <span>{selectedPV.vendorName || "—"}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">Amount:</span>
                    <span className="font-mono text-blue-600">
                      {formatCurrency(selectedPV.amount, selectedPV.currency)}
                    </span>
                  </div>
                </div>
              )}
            </div>
          )}

          {/* Items */}
          {(effectivePO || items.length > 0) && (
            <div className="space-y-3">
              <Label className="text-sm font-semibold">
                Items <span className="text-destructive">*</span>
              </Label>

              <div className="rounded-lg border border-border overflow-hidden">
                {/* Column headers */}
                <div className="grid grid-cols-[1.75rem_5rem_1fr_4rem_4rem_6rem_1fr_1.75rem] gap-x-3 px-3 py-2 bg-muted/60 border-b border-border">
                  <span className="text-xs font-medium text-muted-foreground">
                    #
                  </span>
                  <span className="text-xs font-medium text-muted-foreground uppercase tracking-wider">
                    Code
                  </span>
                  <span className="text-xs font-medium text-muted-foreground uppercase tracking-wider">
                    Description
                  </span>
                  <span className="text-xs font-medium text-muted-foreground uppercase tracking-wider text-center">
                    Ordered
                  </span>
                  <span className="text-xs font-medium text-muted-foreground uppercase tracking-wider text-center">
                    Received
                  </span>
                  <span className="text-xs font-medium text-muted-foreground uppercase tracking-wider">
                    Condition
                  </span>
                  <span className="text-xs font-medium text-muted-foreground uppercase tracking-wider">
                    Remarks
                  </span>
                  <span />
                </div>

                {/* Item rows */}
                <div className="divide-y divide-border/60">
                  {items.map((item, index) => (
                    <div
                      key={item._key}
                      className="grid grid-cols-[1.75rem_5rem_1fr_4rem_4rem_6rem_1fr_1.75rem] gap-x-3 px-3 py-2 items-center hover:bg-muted/20 transition-colors"
                    >
                      {/* # */}
                      <span className="text-xs text-muted-foreground/50 font-mono tabular-nums">
                        {String(index + 1).padStart(2, "0")}
                      </span>

                      {/* Item Code */}
                      <input
                        className="min-w-0 w-full bg-transparent text-sm font-mono placeholder:text-muted-foreground/40 focus:outline-none focus:bg-muted/30 rounded px-1.5 py-1 -mx-1.5 border border-transparent focus:border-border transition-colors"
                        placeholder="SKU"
                        value={item.itemCode ?? ""}
                        disabled={isCreating}
                        onChange={(e) =>
                          updateItem(item._key, "itemCode", e.target.value)
                        }
                      />

                      {/* Description */}
                      <input
                        className="min-w-0 w-full bg-transparent text-sm placeholder:text-muted-foreground/40 focus:outline-none focus:bg-muted/30 rounded px-1.5 py-1 -mx-1.5 border border-transparent focus:border-border transition-colors"
                        placeholder="Item description…"
                        value={item.description}
                        disabled={isCreating}
                        onChange={(e) =>
                          updateItem(item._key, "description", e.target.value)
                        }
                      />

                      {/* Ordered */}
                      <input
                        type="number"
                        min={0}
                        className="w-full bg-transparent text-sm text-center tabular-nums placeholder:text-muted-foreground/40 focus:outline-none focus:bg-muted/30 rounded px-1 py-1 border border-transparent focus:border-border transition-colors"
                        value={item.quantityOrdered}
                        disabled={isCreating}
                        onChange={(e) =>
                          updateItem(
                            item._key,
                            "quantityOrdered",
                            Number(e.target.value),
                          )
                        }
                      />

                      {/* Received */}
                      <input
                        type="number"
                        min={0}
                        className="w-full bg-transparent text-sm text-center tabular-nums placeholder:text-muted-foreground/40 focus:outline-none focus:bg-muted/30 rounded px-1 py-1 border border-transparent focus:border-border transition-colors"
                        value={item.quantityReceived}
                        disabled={isCreating}
                        onChange={(e) =>
                          updateItem(
                            item._key,
                            "quantityReceived",
                            Number(e.target.value),
                          )
                        }
                      />

                      {/* Condition */}
                      <select
                        className="w-full bg-transparent text-sm placeholder:text-muted-foreground/40 focus:outline-none focus:bg-muted/30 rounded px-1.5 py-1 -mx-1.5 border border-transparent focus:border-border transition-colors appearance-none cursor-pointer"
                        value={item.condition}
                        disabled={isCreating}
                        onChange={(e) =>
                          updateItem(item._key, "condition", e.target.value)
                        }
                      >
                        <option value="good">Good</option>
                        <option value="damaged">Damaged</option>
                        <option value="missing">Missing</option>
                      </select>

                      {/* Remarks */}
                      <input
                        className="min-w-0 w-full bg-transparent text-sm placeholder:text-muted-foreground/40 focus:outline-none focus:bg-muted/30 rounded px-1.5 py-1 -mx-1.5 border border-transparent focus:border-border transition-colors"
                        placeholder="Per-line remarks"
                        value={item.remarks ?? ""}
                        disabled={isCreating}
                        onChange={(e) =>
                          updateItem(item._key, "remarks", e.target.value)
                        }
                      />

                      {/* Delete */}
                      <button
                        type="button"
                        onClick={() => removeItem(item._key)}
                        disabled={isCreating}
                        className="text-muted-foreground/30 hover:text-red-500 transition-colors flex items-center justify-center disabled:opacity-50"
                      >
                        <Trash2 className="h-3.5 w-3.5" />
                      </button>
                    </div>
                  ))}

                  {/* Add item row */}
                  <button
                    type="button"
                    onClick={addItem}
                    disabled={isCreating}
                    className="w-full flex items-center gap-2 px-3 py-2.5 text-sm text-muted-foreground hover:text-foreground hover:bg-muted/30 transition-colors disabled:opacity-50"
                  >
                    <Plus className="h-3.5 w-3.5" />
                    Add item
                  </button>
                </div>
              </div>
            </div>
          )}

          <Separator />

          {/* Receipt details */}

          <Input
            label="Delivery Consignment Note"
            id="consignment-note"
            value={consignmentNote}
            onChange={(e) => setConsignmentNote(e.target.value)}
            placeholder="e.g. CN-2026-04812"
            descriptionText="Carrier / supplier delivery note reference printed on the GRN."
            disabled={isCreating}
          />

          <Input
            label="Warehouse Location"
            id="warehouse-location"
            required
            value={warehouseLocation}
            onChange={(e) => setWarehouseLocation(e.target.value)}
            placeholder="e.g. Warehouse A, Bay 3"
            disabled={isCreating}
          />
          <Textarea
            label="Notes"
            id="notes"
            rows={3}
            value={notes}
            onChange={(e) => setNotes(e.target.value)}
            placeholder="Optional notes"
            disabled={isCreating}
          />
        </div>
    </ResponsiveSheet>
  );
}
