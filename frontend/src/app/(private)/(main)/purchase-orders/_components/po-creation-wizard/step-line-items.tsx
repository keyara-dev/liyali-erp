"use client";

import { useState } from "react";
import { Package, FileText } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import type { POItem } from "@/types/purchase-order";
import type { Requisition } from "@/types/requisition";
import { POItemsEditableTable } from "../po-items-editable-table";

export interface StepLineItemsProps {
  items: POItem[];
  requisition: Requisition;
  currency: string;
  onChange: (items: POItem[]) => void;
  onNext: () => void;
  onBack: () => void;
}

/**
 * PO Creation Wizard — Line Items step.
 *
 * Seeded from a deep copy of the source requisition's items (see
 * buildInitialItems in use-wizard-state). Users adjust descriptions,
 * quantities, and unit prices for the PO here; the original REQ document is
 * never modified — its items are shown read-only as the REQ Est. Price /
 * Variance reference columns.
 */
export function StepLineItems({
  items,
  requisition,
  currency,
  onChange,
  onNext,
  onBack,
}: StepLineItemsProps) {
  const [error, setError] = useState<string | null>(null);

  const handleNext = () => {
    if (items.length === 0) {
      setError("Add at least one line item before continuing.");
      return;
    }
    const invalid = items.some(
      (i) => !i.description.trim() || i.quantity <= 0 || i.unitPrice < 0,
    );
    if (invalid) {
      setError(
        "Every item needs a description, a quantity greater than 0, and a unit price of 0 or more.",
      );
      return;
    }
    setError(null);
    onNext();
  };

  return (
    <div className="flex flex-col flex-1 min-h-0 px-2">
      <div className="flex-1 overflow-y-auto p-4 space-y-4">
        {/* Context — these start as a copy of the REQ items */}
        <div className="rounded-lg border bg-muted/40 p-3 flex items-start gap-2">
          <FileText className="h-4 w-4 text-muted-foreground mt-0.5 shrink-0" />
          <p className="text-xs text-muted-foreground">
            Line items start from requisition{" "}
            <span className="font-mono font-medium text-foreground">
              {requisition.documentNumber}
            </span>
            . Adjust descriptions, quantities, or unit prices for this purchase
            order — the original requisition is not affected. The{" "}
            <span className="font-medium">REQ Est. Price</span> and{" "}
            <span className="font-medium">Variance</span> columns compare your
            edits against the requisition&apos;s estimate.
          </p>
        </div>

        <div className="flex items-center gap-2">
          <Package className="h-4 w-4 text-muted-foreground" />
          <span className="text-sm font-semibold">Line Items</span>
          <span className="text-xs text-muted-foreground">
            ({items.length})
          </span>
        </div>

        <Separator />

        <POItemsEditableTable
          items={items}
          onItemsChange={(next) => {
            onChange(next);
            if (error) setError(null);
          }}
          currency={currency}
          reqItems={requisition.items}
        />

        {error && (
          <p
            className="text-xs text-red-600 dark:text-red-400"
            data-testid="line-items-error"
          >
            {error}
          </p>
        )}
      </div>

      {/* Sticky Footer */}
      <div className="shrink-0 border-t bg-card/5 backdrop-blur-xs flex justify-between gap-2 p-4">
        <Button
          type="button"
          variant="ghost"
          onClick={onBack}
          className="w-full sm:w-auto"
        >
          Back
        </Button>
        <Button
          type="button"
          onClick={handleNext}
          className="w-full sm:w-auto"
          data-testid="line-items-next-button"
        >
          Next
        </Button>
      </div>
    </div>
  );
}
