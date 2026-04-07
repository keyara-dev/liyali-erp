"use client";

import { useState, useCallback } from "react";
import { UserPlus, AlertTriangle, CheckCircle2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Label } from "@/components/ui/label";
import { SelectField } from "@/components/ui/select-field";
import { useVendors } from "@/hooks/use-vendor-queries";
import { CostComparisonPanel } from "@/components/purchase-orders/cost-comparison-panel";
import { InlineVendorForm } from "./inline-vendor-form";
import type { Vendor } from "@/types/vendor";
import type { Quotation } from "@/types/core";
import type { Requisition } from "@/types/requisition";
import type { WizardStep2State, WizardVendorEntry } from "./types";

export interface Step2Props {
  data: WizardStep2State;
  requisition: Requisition;
  onChange: (data: WizardStep2State) => void;
  onNext: () => void;
  onBack: () => void;
}

function generateLocalId(): string {
  return `local-${Date.now()}-${Math.random().toString(36).slice(2, 9)}`;
}

function vendorToEntry(vendor: Vendor): WizardVendorEntry {
  return {
    localId: generateLocalId(),
    vendorId: vendor.id,
    vendorName: vendor.name,
    quotations: [],
  };
}

/**
 * Step 2 — Vendor & Quotes.
 *
 * If the requisition has quotations in metadata, shows them as a clickable
 * supplier selector (mirrors the old CreatePOFromRequisitionDialog).
 * If no quotations exist, falls back to a vendor dropdown.
 * "Add New Vendor" inline form is always available.
 *
 * Requirements: 3.1, 3.2, 3.5, 3.6, 3.7, 3.8, 3.9, 3.10, 3.11
 */
export function Step2VendorQuotes({
  data,
  requisition,
  onChange,
  onNext,
  onBack,
}: Step2Props) {
  const [showInlineForm, setShowInlineForm] = useState(false);
  const [showNoVendorWarning, setShowNoVendorWarning] = useState(false);

  const { data: allVendors = [], isLoading: vendorsLoading } = useVendors();

  // Quotations already attached to the requisition
  const reqQuotations: Quotation[] =
    (requisition.metadata?.quotations as Quotation[]) ?? [];

  // Derive the currently selected vendor entry
  const selectedEntry = data.selectedVendorLocalId
    ? (data.vendors.find((v) => v.localId === data.selectedVendorLocalId) ??
      null)
    : null;

  // ── select from existing quotations ───────────────────────────────────────

  const handleSelectFromQuotation = (q: Quotation) => {
    // Check if this vendor is already in the list
    const existing = data.vendors.find(
      (v) => v.vendorId === q.vendorId || v.vendorName === q.vendorName,
    );

    if (existing) {
      // Just select it
      onChange({ ...data, selectedVendorLocalId: existing.localId });
    } else {
      // Add and select
      const entry: WizardVendorEntry = {
        localId: generateLocalId(),
        vendorId: q.vendorId ?? "",
        vendorName: q.vendorName,
        quotations: [],
        quotedAmount: q.amount,
      };
      onChange({
        ...data,
        vendors: [...data.vendors, entry],
        selectedVendorLocalId: entry.localId,
      });
    }
  };

  // ── select from vendor dropdown (no-quotation fallback) ───────────────────

  const handleVendorDropdownChange = useCallback(
    (vendorId: string) => {
      if (!vendorId) {
        // Clear selection
        onChange({ ...data, selectedVendorLocalId: null });
        return;
      }
      const vendor = (allVendors as Vendor[]).find((v) => v.id === vendorId);
      if (!vendor) return;

      const existing = data.vendors.find((v) => v.vendorId === vendorId);
      if (existing) {
        onChange({ ...data, selectedVendorLocalId: existing.localId });
      } else {
        const entry = vendorToEntry(vendor);
        onChange({
          ...data,
          vendors: [...data.vendors, entry],
          selectedVendorLocalId: entry.localId,
        });
      }
    },
    [data, allVendors, onChange],
  );

  // ── inline new vendor ──────────────────────────────────────────────────────

  const handleNewVendorSaved = useCallback(
    (vendor: Vendor) => {
      const entry = vendorToEntry(vendor);
      onChange({
        ...data,
        vendors: [...data.vendors, entry],
        selectedVendorLocalId: entry.localId,
      });
      setShowInlineForm(false);
    },
    [data, onChange],
  );

  // ── next ───────────────────────────────────────────────────────────────────

  const handleNext = () => {
    if (!data.selectedVendorLocalId) {
      setShowNoVendorWarning(true);
      return;
    }
    setShowNoVendorWarning(false);
    onNext();
  };

  // ── cost comparison ────────────────────────────────────────────────────────

  const costComparisonVendors = data.vendors.map((v) => ({
    vendorId: v.vendorId || v.localId,
    vendorName: v.vendorName,
    quotedAmount: v.quotedAmount,
    isSelected: v.localId === data.selectedVendorLocalId,
  }));

  // The vendor ID currently selected (for the dropdown fallback)
  const selectedVendorId = selectedEntry?.vendorId ?? "";

  return (
    <div className="space-y-6" data-testid="step2-vendor-quotes">
      {/* ── Quotation-based vendor selector ── */}
      {reqQuotations.length > 0 ? (
        <div className="space-y-2">
          <Label className="text-sm font-medium">
            Select Supplier from Quotations
          </Label>
          <p className="text-xs text-muted-foreground">
            {reqQuotations.length} quotation
            {reqQuotations.length !== 1 ? "s" : ""} collected on this
            requisition.
          </p>
          <div className="space-y-2">
            {reqQuotations.map((q, i) => {
              const isSelected =
                selectedEntry?.vendorId === q.vendorId ||
                selectedEntry?.vendorName === q.vendorName;
              return (
                <button
                  key={`${q.vendorId}-${i}`}
                  type="button"
                  onClick={() => handleSelectFromQuotation(q)}
                  className={`w-full flex items-center justify-between rounded-md border px-3 py-2.5 text-left transition-colors ${
                    isSelected
                      ? "border-primary bg-primary/5"
                      : "hover:bg-muted/50"
                  }`}
                >
                  <span className="flex items-center gap-2 text-sm font-medium">
                    {isSelected && (
                      <CheckCircle2 className="h-4 w-4 text-primary shrink-0" />
                    )}
                    {q.vendorName}
                  </span>
                  <span className="text-sm font-mono text-muted-foreground">
                    {q.currency || requisition.currency}{" "}
                    {q.amount?.toLocaleString("en-ZM", {
                      minimumFractionDigits: 2,
                      maximumFractionDigits: 2,
                    })}
                  </span>
                </button>
              );
            })}
          </div>
          {!selectedEntry && (
            <p className="text-xs text-amber-600">
              Select a supplier above or leave blank to assign later.
            </p>
          )}
        </div>
      ) : (
        /* ── Fallback: vendor dropdown when no quotations ── */
        <div className="space-y-3">
          <Alert className="border-amber-200 bg-amber-50 dark:bg-amber-950/30 py-2">
            <AlertTriangle className="h-4 w-4 text-amber-600" />
            <AlertDescription className="text-amber-700 dark:text-amber-300 text-xs">
              No quotations on this requisition yet — select a vendor manually
              or add a new one below.
            </AlertDescription>
          </Alert>
          <SelectField
            label="Vendor (optional)"
            placeholder="No vendor — assign later"
            value={selectedVendorId}
            onValueChange={handleVendorDropdownChange}
            isLoading={vendorsLoading}
            options={(allVendors as Vendor[])
              .filter((v) => v.active)
              .map((v) => ({ value: v.id, name: v.name }))}
          />
        </div>
      )}

      {/* ── Add New Vendor ── */}
      {!showInlineForm ? (
        <Button
          type="button"
          variant="outline"
          size="sm"
          onClick={() => setShowInlineForm(true)}
          data-testid="add-new-vendor-btn"
        >
          <UserPlus className="mr-1.5 h-4 w-4" />
          Add New Vendor
        </Button>
      ) : (
        <InlineVendorForm
          onSaved={handleNewVendorSaved}
          onCancel={() => setShowInlineForm(false)}
        />
      )}

      {/* ── Cost comparison (when at least one vendor added) ── */}
      {data.vendors.length > 0 && (
        <CostComparisonPanel
          estimatedCost={requisition.totalAmount ?? 0}
          currency={requisition.currency ?? "ZMW"}
          vendors={costComparisonVendors}
        />
      )}

      {/* ── No vendor warning ── */}
      {showNoVendorWarning && (
        <Alert
          variant="default"
          className="border-amber-300 bg-amber-50 dark:bg-amber-950/30 dark:border-amber-700"
          data-testid="no-vendor-warning"
        >
          <AlertTriangle className="h-4 w-4 text-amber-600 dark:text-amber-400" />
          <AlertDescription className="text-amber-800 dark:text-amber-200">
            No supplier selected. You can still proceed — the supplier can be
            assigned later.
            <div className="mt-2 flex gap-2">
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={() => setShowNoVendorWarning(false)}
              >
                Go Back
              </Button>
              <Button
                type="button"
                size="sm"
                className="bg-amber-600 hover:bg-amber-700 text-white"
                onClick={() => {
                  setShowNoVendorWarning(false);
                  onNext();
                }}
                data-testid="no-vendor-warning-proceed"
              >
                Proceed Anyway
              </Button>
            </div>
          </AlertDescription>
        </Alert>
      )}

      {/* ── Footer ── */}
      <div className="flex justify-between pt-2">
        <Button
          type="button"
          variant="outline"
          onClick={onBack}
          data-testid="step2-back-button"
        >
          Back
        </Button>
        <Button
          type="button"
          onClick={handleNext}
          data-testid="step2-next-button"
        >
          Next
        </Button>
      </div>
    </div>
  );
}
