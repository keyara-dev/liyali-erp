"use client";

import { Truck } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Separator } from "@/components/ui/separator";
import { formatCurrency } from "@/lib/utils";
import type { Requisition } from "@/types/requisition";
import type { WizardStep3State } from "./types";

// ============================================================================
// TYPES
// ============================================================================

export interface Step3ShippingTaxProps {
  data: WizardStep3State;
  requisition: Requisition;
  onChange: (step3: WizardStep3State) => void;
  onNext: () => void;
  onBack: () => void;
}

// ============================================================================
// COMPONENT
// ============================================================================

/**
 * Step 3 of the PO Creation Wizard — Shipping & Tax.
 *
 * Renders labelled input fields for all 9 shipping/tax fields. All fields are
 * optional (Req 2.5). Pre-populates Receiver Name, Department, and Fund Source
 * from the linked Requisition (Req 2.2, 2.3). Displays a live totals preview
 * only after the user explicitly types into Tax Rate or Delivery Cost (Req 2.4).
 * Computes tax amount in real time as (subtotal × taxRate / 100) (Req 2.6).
 *
 * Requirements: 2.1, 2.2, 2.3, 2.4, 2.5, 2.6
 */
export function Step3ShippingTax({
  data,
  requisition,
  onChange,
  onNext,
  onBack,
}: Step3ShippingTaxProps) {
  // ── field helpers ──────────────────────────────────────────────────────────

  const set = <K extends keyof WizardStep3State>(
    key: K,
    value: WizardStep3State[K],
  ) => {
    onChange({ ...data, [key]: value });
  };

  const handleTaxRateChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    onChange({
      ...data,
      taxRate: e.target.value,
      userEnteredTaxOrDelivery: true,
    });
  };

  const handleDeliveryCostChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    onChange({
      ...data,
      deliveryCost: e.target.value,
      userEnteredTaxOrDelivery: true,
    });
  };

  // ── live totals computation ────────────────────────────────────────────────

  const subtotal = requisition.totalAmount ?? 0;
  const taxRateNum = parseFloat(data.taxRate);
  const deliveryCostNum = parseFloat(data.deliveryCost);

  const taxAmount =
    !isNaN(taxRateNum) && taxRateNum > 0
      ? Math.round(((subtotal * taxRateNum) / 100) * 100) / 100
      : 0;

  const deliveryCostValue = !isNaN(deliveryCostNum) ? deliveryCostNum : 0;
  const totalOrderValue = subtotal + taxAmount + deliveryCostValue;

  const currency = requisition.currency ?? "ZMW";

  // ── render ─────────────────────────────────────────────────────────────────

  return (
    <div
      className="flex flex-col flex-1 min-h-0 px-2"
      data-testid="step3-shipping-tax"
    >
      <div className="flex-1 overflow-y-auto p-4 space-y-6 min-w-0">
        {/* ── Section header ── */}
        <div className="flex items-center gap-2">
          <Truck className="h-4 w-4 text-muted-foreground" />
          <span className="text-sm font-semibold">Shipping & Tax Details</span>
          <span className="text-xs text-muted-foreground">
            (all fields optional)
          </span>
        </div>

        <Separator />

        {/* ── Receiver details ── */}
        <div className="space-y-4">
          <p className="text-xs font-medium text-muted-foreground uppercase tracking-wide">
            Receiver Information
          </p>

          <Input
            label="Receiver Name"
            name="step3-receiver-name"
            placeholder="e.g. John Banda"
            value={data.receiverName}
            onChange={(e) => set("receiverName", e.target.value)}
          />

          <Input
            label="Department"
            name="step3-receiver-dept"
            placeholder="e.g. Finance"
            value={data.receiverDept}
            onChange={(e) => set("receiverDept", e.target.value)}
          />

          <Input
            label="Address"
            name="step3-receiver-address"
            placeholder="e.g. Plot 123, Cairo Road, Lusaka"
            value={data.receiverAddress}
            onChange={(e) => set("receiverAddress", e.target.value)}
          />

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <Input
              label="Contact / Phone"
              name="step3-receiver-contact"
              placeholder="e.g. +260 97 1234567"
              value={data.receiverContact}
              onChange={(e) => set("receiverContact", e.target.value)}
            />

            <Input
              label="Email Address"
              name="step3-receiver-email"
              type="email"
              placeholder="e.g. receiver@example.com"
              value={data.receiverEmail}
              onChange={(e) => set("receiverEmail", e.target.value)}
            />
          </div>
        </div>

        <Separator />

        {/* ── Procurement & funding ── */}
        <div className="space-y-4">
          <p className="text-xs font-medium text-muted-foreground uppercase tracking-wide">
            Procurement & Funding
          </p>

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <Input
              label="Purchase Type"
              name="step3-purchase-type"
              placeholder="e.g. Direct Purchase"
              value={data.purchaseType}
              onChange={(e) => set("purchaseType", e.target.value)}
            />

            <Input
              label="Fund Source"
              name="step3-fund-source"
              placeholder="e.g. Operating Budget"
              value={data.fundSource}
              onChange={(e) => set("fundSource", e.target.value)}
            />
          </div>
        </div>

        <Separator />

        {/* ── Tax & delivery ── */}
        <div className="space-y-4">
          <p className="text-xs font-medium text-muted-foreground uppercase tracking-wide">
            Tax & Delivery
          </p>

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <Input
              label="Tax Rate (%)"
              name="step3-tax-rate"
              type="number"
              min="0"
              max="100"
              step="0.01"
              placeholder="e.g. 16"
              value={data.taxRate}
              onChange={handleTaxRateChange}
            />

            <Input
              label="Delivery Cost"
              name="step3-delivery-cost"
              type="number"
              min="0"
              step="0.01"
              placeholder="e.g. 500.00"
              value={data.deliveryCost}
              onChange={handleDeliveryCostChange}
            />
          </div>
        </div>

        {/* ── Live totals preview (Req 2.4) — only shown after user types tax/delivery ── */}
        {data.userEnteredTaxOrDelivery && (
          <div
            className="rounded-lg border bg-muted/40 p-4 space-y-3"
            data-testid="live-totals-preview"
          >
            <p className="text-sm font-semibold">Order Totals Preview</p>
            <Separator />
            <div className="space-y-2 text-sm">
              <div className="flex justify-between">
                <span className="text-muted-foreground">Sub Total</span>
                <span
                  className="font-mono font-medium"
                  data-testid="totals-subtotal"
                >
                  {formatCurrency(subtotal, currency)}
                </span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">
                  Tax Amount
                  {!isNaN(taxRateNum) && taxRateNum > 0
                    ? ` (${taxRateNum}%)`
                    : ""}
                </span>
                <span
                  className="font-mono font-medium"
                  data-testid="totals-tax-amount"
                >
                  {formatCurrency(taxAmount, currency)}
                </span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Delivery Cost</span>
                <span
                  className="font-mono font-medium"
                  data-testid="totals-delivery-cost"
                >
                  {formatCurrency(deliveryCostValue, currency)}
                </span>
              </div>
              <Separator />
              <div className="flex justify-between font-semibold">
                <span>Total Order Value</span>
                <span
                  className="font-mono text-blue-600 dark:text-blue-400"
                  data-testid="totals-total-order-value"
                >
                  {formatCurrency(totalOrderValue, currency)}
                </span>
              </div>
            </div>
          </div>
        )}
      </div>

      {/* ── Sticky Footer ── */}
      <div className="shrink-0 border-t bg-card/5 backdrop-blur-xs flex flex-col-reverse sm:flex-row justify-between gap-2 p-4">
        <Button
          type="button"
          variant="outline"
          onClick={onBack}
          className="w-full sm:w-auto"
          data-testid="step3-back-button"
        >
          Back
        </Button>
        {/* Next is always enabled — all fields optional per Req 2.5 */}
        <Button
          type="button"
          onClick={onNext}
          className="w-full sm:w-auto"
          data-testid="step3-next-button"
        >
          Next
        </Button>
      </div>
    </div>
  );
}
