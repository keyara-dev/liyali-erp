"use client";

import { useCallback, useMemo, useState } from "react";
import type { Requisition } from "@/types/requisition";
import type {
  WizardState,
  WizardStep1State,
  WizardStep2State,
  WizardStep3State,
  WizardStep4State,
} from "./types";

// ============================================================================
// HELPERS
// ============================================================================

/**
 * Derives the initial Step 1 state from a source Requisition.
 * Requirements: 2.6
 */
function buildInitialStep1(requisition: Requisition): WizardStep1State {
  return {
    title: requisition.title ?? "",
    description: requisition.description ?? "",
    departmentId: requisition.departmentId ?? "",
    department: requisition.department ?? "",
    priority:
      (requisition.priority?.toUpperCase() as WizardStep1State["priority"]) ??
      "MEDIUM",
    budgetCode: requisition.budgetCode ?? "",
    costCenter: requisition.costCenter ?? "",
    projectCode: requisition.projectCode ?? "",
    deliveryDate: (() => {
      if (!requisition.requiredByDate) return null;
      const d = new Date(requisition.requiredByDate);
      return isNaN(d.getTime()) ? null : d;
    })(),
    currency: requisition.currency ?? "",
  };
}

const INITIAL_STEP2: WizardStep2State = {
  vendors: [],
  selectedVendorLocalId: null,
};

const INITIAL_STEP3: WizardStep3State = {
  receiverName: "",
  receiverDept: "",
  receiverAddress: "",
  receiverContact: "",
  receiverEmail: "",
  purchaseType: "",
  fundSource: "",
  taxRate: "",
  deliveryCost: "",
  userEnteredTaxOrDelivery: false,
};

const INITIAL_STEP4: WizardStep4State = {
  workflowId: "",
  procurementFlow: "",
};

/**
 * Derives the initial Step 3 state from a source Requisition.
 * Pre-populates receiver name, department, and fund source from the requisition
 * so the user doesn't have to re-enter information already captured on the REQ.
 *
 * Requirements: 2.2, 2.3
 */
export function buildInitialStep3(requisition: Requisition): WizardStep3State {
  return {
    receiverName: requisition.requesterName ?? "",
    receiverDept: requisition.department ?? "",
    receiverAddress: "",
    receiverContact: "",
    receiverEmail: "",
    purchaseType: "",
    fundSource: requisition.costCenter ?? requisition.budgetCode ?? "",
    taxRate: "",
    deliveryCost: "",
    userEnteredTaxOrDelivery: false,
  };
}

// ============================================================================
// HOOK
// ============================================================================

export interface UseWizardStateReturn {
  wizardState: WizardState;
  setStep1: (step1: WizardStep1State) => void;
  setStep2: (step2: WizardStep2State) => void;
  setStep3: (step3: WizardStep3State) => void;
  setStep4: (step4: WizardStep4State) => void;
  resetWizard: () => void;
}

/**
 * Manages the accumulated state for the PO creation wizard.
 *
 * - Initialises Step 1 from the provided `requisition` on first render.
 * - Exposes granular setters for each step so child components only update
 *   their own slice of state.
 * - `resetWizard` restores all state back to the initial values derived from
 *   the requisition.
 *
 * Requirements: 2.6, 7.1, 7.2, 7.3
 */
export function useWizardState(requisition: Requisition): UseWizardStateReturn {
  const initialStep1 = useMemo(
    () => buildInitialStep1(requisition),
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [requisition.id], // re-derive only when the requisition identity changes
  );

  const [wizardState, setWizardState] = useState<WizardState>(() => ({
    step1: buildInitialStep1(requisition),
    step2: INITIAL_STEP2,
    step3: buildInitialStep3(requisition),
    step4: INITIAL_STEP4,
  }));

  const setStep1 = useCallback((step1: WizardStep1State) => {
    setWizardState((prev) => ({ ...prev, step1 }));
  }, []);

  const setStep2 = useCallback((step2: WizardStep2State) => {
    setWizardState((prev) => ({ ...prev, step2 }));
  }, []);

  const setStep3 = useCallback((step3: WizardStep3State) => {
    setWizardState((prev) => ({ ...prev, step3 }));
  }, []);

  const setStep4 = useCallback((step4: WizardStep4State) => {
    setWizardState((prev) => ({ ...prev, step4 }));
  }, []);

  const resetWizard = useCallback(() => {
    setWizardState({
      step1: initialStep1,
      step2: INITIAL_STEP2,
      step3: buildInitialStep3(requisition),
      step4: INITIAL_STEP4,
    });
  }, [initialStep1, requisition]);

  return { wizardState, setStep1, setStep2, setStep3, setStep4, resetWizard };
}
