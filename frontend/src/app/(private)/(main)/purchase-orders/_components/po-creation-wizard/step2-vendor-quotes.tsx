"use client";

import { useState, useCallback } from "react";
import { UserPlus, AlertTriangle } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { SearchSelectField } from "@/components/ui/search-select-field";
import { useVendors } from "@/hooks/use-vendor-queries";
import { CostComparisonPanel } from "@/components/purchase-orders/cost-comparison-panel";
import { VendorEntryRow } from "./vendor-entry-row";
import { InlineVendorForm } from "./inline-vendor-form";
import type { Vendor } from "@/types/vendor";
import type { Quotation } from "@/types/core";
import type { Requisition } from "@/types/requisition";
import type { WizardStep2State, WizardVendorEntry } from "./types";

// ============================================================================
// TYPES
// ============================================================================

export interface Step2Props {
  data: WizardStep2State;
  requisition: Requisition;
  onChange: (data: WizardStep2State) => void;
  onNext: () => void;
  onBack: () => void;
}

// ============================================================================
// HELPERS
// ============================================================================

/** Generate a stable local ID without external dependencies */
function generateLocalId(): string {
  return `local-${Date.now()}-${Math.random().toString(36).slice(2, 9)}`;
}

/** Build a WizardVendorEntry from a Vendor */
function vendorToEntry(vendor: Vendor): WizardVendorEntry {
  return {
    localId: generateLocalId(),
    vendorId: vendor.id,
    vendorName: vendor.name,
    quotations: [],
  };
}

// ============================================================================
// COMPONENT
// ============================================================================

/**
 * Step 2 of the PO Creation Wizard — Vendor Selection & Quotations.
 *
 * - Vendor selection combobox using `useVendors` (GET /api/v1/vendors)
 * - One VendorEntryRow per vendor in data.vendors
 * - InlineVendorForm shown when "Add New Vendor" is clicked
 * - CostComparisonPanel shown when at least one vendor is present
 * - "Next" always enabled but shows a warning when no vendor is added
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

  // Fetch existing vendors for the selection combobox (Req 3.1)
  const { data: allVendors = [], isLoading: vendorsLoading } = useVendors();

  // ── vendor selection combobox ──────────────────────────────────────────────

  const handleVendorSelect = useCallback(
    (vendorId: string) => {
      if (!vendorId) return;

      // Avoid duplicates (Req 3.7)
      const alreadyAdded = data.vendors.some((v) => v.vendorId === vendorId);
      if (alreadyAdded) return;

      const vendor = Array.isArray(allVendors)
        ? allVendors.find((v) => v.id === vendorId)
        : undefined;
      if (!vendor) return;

      const entry = vendorToEntry(vendor);
      onChange({
        ...data,
        vendors: [...data.vendors, entry],
      });
    },
    [data, allVendors, onChange],
  );

  // ── inline vendor form ─────────────────────────────────────────────────────

  const handleNewVendorSaved = useCallback(
    (vendor: Vendor) => {
      const entry = vendorToEntry(vendor);
      // Add vendor and pre-select it as supplier (Req 3.5)
      onChange({
        ...data,
        vendors: [...data.vendors, entry],
        selectedVendorLocalId: entry.localId,
      });
      setShowInlineForm(false);
    },
    [data, onChange],
  );

  // ── per-vendor callbacks ───────────────────────────────────────────────────

  const handleSelectAsSupplier = useCallback(
    (localId: string) => {
      onChange({ ...data, selectedVendorLocalId: localId });
    },
    [data, onChange],
  );

  const handleQuotationsChange = useCallback(
    (localId: string, quotations: Quotation[]) => {
      // Derive the quoted amount from the first quotation (if any)
      const quotedAmount =
        quotations.length > 0 ? quotations[0].amount : undefined;

      onChange({
        ...data,
        vendors: data.vendors.map((v) =>
          v.localId === localId ? { ...v, quotations, quotedAmount } : v,
        ),
      });
    },
    [data, onChange],
  );

  const handleRemoveVendor = useCallback(
    (localId: string) => {
      const updated = data.vendors.filter((v) => v.localId !== localId);
      onChange({
        ...data,
        vendors: updated,
        // Clear selected supplier if the removed vendor was selected
        selectedVendorLocalId:
          data.selectedVendorLocalId === localId
            ? null
            : data.selectedVendorLocalId,
      });
    },
    [data, onChange],
  );

  // ── next handler ───────────────────────────────────────────────────────────

  const handleNext = () => {
    if (data.vendors.length === 0) {
      // Allow advancing but show a warning (Req 3.11)
      setShowNoVendorWarning(true);
      return;
    }
    setShowNoVendorWarning(false);
    onNext();
  };

  const handleNextAnyway = () => {
    setShowNoVendorWarning(false);
    onNext();
  };

  // ── cost comparison data ───────────────────────────────────────────────────

  const costComparisonVendors = data.vendors.map((v) => ({
    vendorId: v.vendorId || v.localId,
    vendorName: v.vendorName,
    quotedAmount: v.quotedAmount,
    isSelected: v.localId === data.selectedVendorLocalId,
  }));

  // Vendor options for the combobox — exclude already-added vendors
  const addedVendorIds = new Set(data.vendors.map((v) => v.vendorId));
  const vendorOptions = Array.isArray(allVendors)
    ? allVendors
        .filter((v) => !addedVendorIds.has(v.id))
        .map((v) => ({ id: v.id, name: v.name }))
    : [];

  // ── render ─────────────────────────────────────────────────────────────────

  return (
    <div className="space-y-6" data-testid="step2-vendor-quotes">
      {/* ── Vendor selection panel (Req 3.1) ── */}
      <div className="space-y-3">
        <div className="flex items-center justify-between gap-3">
          <div className="flex-1">
            <SearchSelectField
              label="Add Existing Vendor"
              placeholder="Search and select a vendor..."
              options={vendorOptions}
              isLoading={vendorsLoading}
              onValueChange={handleVendorSelect}
              value=""
              listItemName="name"
              data-testid="vendor-select-combobox"
            />
          </div>

          {/* Add New Vendor button (Req 3.2) */}
          {!showInlineForm && (
            <Button
              type="button"
              variant="outline"
              size="sm"
              className="mt-5 shrink-0"
              onClick={() => setShowInlineForm(true)}
              data-testid="add-new-vendor-btn"
            >
              <UserPlus className="mr-1.5 h-4 w-4" />
              Add New Vendor
            </Button>
          )}
        </div>

        {/* Inline vendor form (Req 3.2, 3.3, 3.4, 3.5, 3.6) */}
        {showInlineForm && (
          <InlineVendorForm
            onSaved={handleNewVendorSaved}
            onCancel={() => setShowInlineForm(false)}
          />
        )}
      </div>

      {/* ── Vendor entry rows (Req 3.8) ── */}
      {data.vendors.length > 0 && (
        <div className="space-y-4" data-testid="vendor-entry-list">
          {data.vendors.map((vendor) => (
            <VendorEntryRow
              key={vendor.localId}
              vendor={vendor}
              currency={requisition.currency ?? "ZMW"}
              estimatedCost={requisition.totalAmount ?? 0}
              isSelected={vendor.localId === data.selectedVendorLocalId}
              onSelectAsSupplier={handleSelectAsSupplier}
              onQuotationsChange={handleQuotationsChange}
              onRemove={handleRemoveVendor}
              vendors={Array.isArray(allVendors) ? allVendors : []}
              requisitionId={requisition.id}
            />
          ))}
        </div>
      )}

      {/* ── Cost comparison panel (Req 3.9) ── */}
      {data.vendors.length > 0 && (
        <CostComparisonPanel
          estimatedCost={requisition.totalAmount ?? 0}
          currency={requisition.currency ?? "ZMW"}
          vendors={costComparisonVendors}
        />
      )}

      {/* ── No vendor warning (Req 3.11) ── */}
      {showNoVendorWarning && (
        <Alert
          variant="default"
          className="border-amber-300 bg-amber-50 dark:bg-amber-950/30 dark:border-amber-700"
          data-testid="no-vendor-warning"
        >
          <AlertTriangle className="h-4 w-4 text-amber-600 dark:text-amber-400" />
          <AlertDescription className="text-amber-800 dark:text-amber-200">
            No vendor has been added. You can still proceed to the next step,
            but no supplier will be assigned to this purchase order.
            <div className="mt-2 flex gap-2">
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={() => setShowNoVendorWarning(false)}
                data-testid="no-vendor-warning-cancel"
              >
                Go Back
              </Button>
              <Button
                type="button"
                size="sm"
                variant="default"
                className="bg-amber-600 hover:bg-amber-700 text-white"
                onClick={handleNextAnyway}
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
