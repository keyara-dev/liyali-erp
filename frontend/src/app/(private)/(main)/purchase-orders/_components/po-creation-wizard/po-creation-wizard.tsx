"use client";

import { useState } from "react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { toast } from "sonner";
import { useQueryClient } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { ShoppingCart } from "lucide-react";
import { QUERY_KEYS } from "@/lib/constants";
import {
  createPurchaseOrderFromRequisition,
  updatePurchaseOrder,
} from "@/app/_actions/purchase-orders";
import type { Requisition } from "@/types/requisition";
import { WizardStepIndicator } from "./wizard-step-indicator";
import { Step1PODetails } from "./step1-po-details";
import { StepLineItems } from "./step-line-items";
import { Step2VendorQuotes } from "./step2-vendor-quotes";
import { Step3ShippingTax } from "./step3-shipping-tax";
import { Step4ReviewConfirm } from "./step4-review-confirm";
import { useWizardState } from "./use-wizard-state";

// ============================================================================
// TYPES
// ============================================================================

export interface POCreationWizardProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  requisition: Requisition;
}

// ============================================================================
// CONSTANTS
// ============================================================================

const WIZARD_STEPS = [
  { label: "PO Details" },
  { label: "Line Items" },
  { label: "Vendor & Quotes" },
  { label: "Shipping & Tax" },
  { label: "Review & Confirm" },
];

// ============================================================================
// COMPONENT
// ============================================================================

/**
 * POCreationWizard — root dialog component that wires all four wizard steps.
 *
 * Owns the WizardState via useWizardState. Handles step navigation with
 * validation gating on Step 1. On close, resets all state back to Step 1.
 * On submit, calls createPurchaseOrderFromRequisition, then patches quotations
 * and shipping metadata via updatePurchaseOrder (non-blocking), invalidates
 * the purchase orders cache, and shows a success toast.
 *
 * Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 1.6, 1.7, 3.1, 3.3
 */
export function POCreationWizard({
  open,
  onOpenChange,
  requisition,
}: POCreationWizardProps) {
  const queryClient = useQueryClient();
  const router = useRouter();
  const [currentStep, setCurrentStep] = useState<1 | 2 | 3 | 4 | 5>(1);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const {
    wizardState,
    setStep1,
    setItems,
    setStep2,
    setStep3,
    setStep4,
    resetWizard,
  } = useWizardState(requisition);

  // ── navigation ─────────────────────────────────────────────────────────────
  // Visual order: 1 PO Details · 2 Line Items · 3 Vendor & Quotes ·
  //               4 Shipping & Tax · 5 Review & Confirm

  // Step 1 (PO Details) → Step 2 (Line Items): called after Step 1 validation passes
  const handleStep1Next = () => {
    setCurrentStep(2);
  };

  // Step 2 (Line Items) → Step 3 (Vendor & Quotes)
  const handleLineItemsNext = () => {
    setCurrentStep(3);
  };

  // Step 2 (Line Items) → Step 1 (PO Details)
  const handleLineItemsBack = () => {
    setCurrentStep(1);
  };

  // Step 3 (Vendor & Quotes) → Step 4 (Shipping & Tax)
  const handleStep2Next = () => {
    setCurrentStep(4);
  };

  // Step 3 (Vendor & Quotes) → Step 2 (Line Items)
  const handleStep2Back = () => {
    setCurrentStep(2);
  };

  // Step 4 (Shipping & Tax) → Step 5 (Review & Confirm)
  const handleStep3Next = () => {
    setCurrentStep(5);
  };

  // Step 4 (Shipping & Tax) → Step 3 (Vendor & Quotes)
  const handleStep3Back = () => {
    setCurrentStep(3);
  };

  // Step 5 (Review & Confirm) → Step 4 (Shipping & Tax)
  const handleStep4Back = () => {
    setCurrentStep(4);
  };

  // ── close / reset ──────────────────────────────────────────────────────────

  const handleOpenChange = (nextOpen: boolean) => {
    if (!nextOpen && !isSubmitting) {
      // Req 1.7: discard all WizardState and reset to Step 1
      resetWizard();
      setCurrentStep(1);
    }
    onOpenChange(nextOpen);
  };

  // ── submit ─────────────────────────────────────────────────────────────────

  const handleSubmit = async () => {
    setIsSubmitting(true);

    // Derive the selected vendor from Step 2 state
    const selectedVendor = wizardState.step2.selectedVendorLocalId
      ? wizardState.step2.vendors.find(
          (v) => v.localId === wizardState.step2.selectedVendorLocalId,
        )
      : null;

    // Edited PO line items (a copy of the REQ items adjusted in the Line Items
    // step). Normalize amounts and derive the items-only subtotal. The source
    // requisition is never touched — these become the new PO's items only.
    const editedItems = wizardState.items.map((i) => ({
      ...i,
      amount: i.quantity * i.unitPrice,
      totalPrice: i.quantity * i.unitPrice,
    }));
    const itemsSubtotal = editedItems.reduce((s, i) => s + i.amount, 0);

    try {
      // Req 5.5: call createPurchaseOrderFromRequisition with wizard state
      const result = await createPurchaseOrderFromRequisition(
        requisition,
        wizardState.step4.workflowId,
        selectedVendor?.vendorId || undefined,
        selectedVendor?.vendorName || undefined,
        wizardState.step4.procurementFlow,
        editedItems,
        itemsSubtotal,
      );

      if (!result.success || !result.data) {
        throw new Error(result.message || "Failed to create purchase order");
      }

      const createdPO = result.data;

      // Quotations to persist: use live quotations from step2 (which includes
      // any newly added ones), falling back to the REQ's existing quotations
      const liveQuotations =
        wizardState.step2.quotations ??
        (requisition.metadata?.quotations as any[]) ??
        [];

      // Req 3.1, 3.3: Build shippingMeta from step3 — include only non-empty
      // string fields (trim check) and numeric fields > 0
      const step3 = wizardState.step3;
      const shippingMeta: Record<string, string | number> = {};

      const stringFields = [
        "receiverName",
        "receiverDept",
        "receiverAddress",
        "receiverContact",
        "receiverEmail",
        "purchaseType",
        "fundSource",
      ] as const;

      for (const field of stringFields) {
        const val = step3[field];
        if (typeof val === "string" && val.trim() !== "") {
          shippingMeta[field] = val;
        }
      }

      const taxRateNum = parseFloat(step3.taxRate);
      if (!isNaN(taxRateNum) && taxRateNum > 0) {
        shippingMeta.taxRate = taxRateNum;
      }

      const deliveryCostNum = parseFloat(step3.deliveryCost);
      if (!isNaN(deliveryCostNum) && deliveryCostNum > 0) {
        shippingMeta.deliveryCost = deliveryCostNum;
      }

      // Req 3.3: deep-merge metadata — quotations + shippingMeta + selected quotation file
      // Also update totalAmount to include tax + delivery so the stored value is always the true grand total.
      // itemsSubtotal here is the edited line-items subtotal computed above.
      const selectedQuotationFileId = wizardState.step2.selectedQuotationFileId;
      const wizardTaxAmount =
        !isNaN(taxRateNum) && taxRateNum > 0
          ? Math.round(((itemsSubtotal * taxRateNum) / 100) * 100) / 100
          : 0;
      const wizardDeliveryCost =
        !isNaN(deliveryCostNum) && deliveryCostNum > 0 ? deliveryCostNum : 0;
      const grandTotal = itemsSubtotal + wizardTaxAmount + wizardDeliveryCost;

      updatePurchaseOrder({
        poId: createdPO.id,
        purchaseOrderId: createdPO.id,
        metadata: {
          quotations: liveQuotations,
          ...shippingMeta,
          ...(selectedQuotationFileId
            ? { selectedQuotationFileUrl: selectedQuotationFileId }
            : {}),
        },
        // Only update totalAmount if tax or delivery was entered — otherwise leave as-is
        ...(grandTotal > itemsSubtotal ? { totalAmount: grandTotal } : {}),
      }).catch(() => {
        toast.warning(
          "Purchase order created, but quotations could not be saved. You can add them from the PO detail page.",
        );
      });

      // Req 5.7: invalidate purchase orders cache
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.PURCHASE_ORDERS.ALL],
      });

      // Req 5.7: show success toast, close dialog, navigate to PO detail
      toast.success("Purchase order created successfully");
      handleOpenChange(false);
      router.push(`/purchase-orders/${createdPO.id}`);
    } catch (err) {
      // Req 5.8: re-throw so Step4 can display the inline error
      throw err;
    } finally {
      setIsSubmitting(false);
    }
  };

  // ── render ─────────────────────────────────────────────────────────────────

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent
        className="w-full max-w-lg sm:max-w-2xl p-0 flex flex-col h-[90svh] max-h-[90vh] overflow-hidden"
        onInteractOutside={(e) => e.preventDefault()}
        data-testid="po-creation-wizard"
      >
        <DialogHeader className="p-4 pb-3 shrink-0 border-b">
          <DialogTitle className="flex items-center gap-2">
            <ShoppingCart className="h-5 w-5" />
            Create Purchase Order
          </DialogTitle>

          {/* Req 1.2: step indicator */}
          <div className="pt-2">
            <WizardStepIndicator
              currentStep={currentStep}
              steps={WIZARD_STEPS}
            />
          </div>
        </DialogHeader>

        {/* Step content */}
        {currentStep === 1 && (
          <Step1PODetails
            data={wizardState.step1}
            requisition={requisition}
            onChange={setStep1}
            onNext={handleStep1Next}
          />
        )}

        {/* Step 2 — Line Items (editable copy of the REQ items) */}
        {currentStep === 2 && (
          <StepLineItems
            items={wizardState.items}
            requisition={requisition}
            currency={wizardState.step1.currency || requisition.currency || ""}
            onChange={setItems}
            onNext={handleLineItemsNext}
            onBack={handleLineItemsBack}
          />
        )}

        {currentStep === 3 && (
          <Step2VendorQuotes
            data={wizardState.step2}
            requisition={requisition}
            onChange={setStep2}
            onNext={handleStep2Next}
            onBack={handleStep2Back}
          />
        )}

        {/* Step 4 — Shipping & Tax */}
        {currentStep === 4 && (
          <Step3ShippingTax
            data={wizardState.step3}
            requisition={requisition}
            onChange={setStep3}
            onNext={handleStep3Next}
            onBack={handleStep3Back}
          />
        )}

        {/* Step 5 — Review & Confirm */}
        {currentStep === 5 && (
          <Step4ReviewConfirm
            wizardState={wizardState}
            requisition={requisition}
            onChange={setStep4}
            onSubmit={handleSubmit}
            onBack={handleStep4Back}
            isSubmitting={isSubmitting}
          />
        )}
      </DialogContent>
    </Dialog>
  );
}
