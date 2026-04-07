/**
 * Property-Based Tests for Step2VendorQuotes — QuotationCollectionSection Count
 *
 * **Property 6: QuotationCollectionSection count matches vendor count**
 * For any N vendors present in data.vendors, exactly N VendorEntryRow
 * components are rendered — one per vendor.
 *
 * **Validates: Requirements 3.8**
 */

import { describe, it, vi, afterEach } from "vitest";
import * as fc from "fast-check";
import { render, cleanup } from "@testing-library/react";
import { Step2VendorQuotes } from "@/app/(private)/(main)/purchase-orders/_components/po-creation-wizard/step2-vendor-quotes";
import type {
  WizardStep2State,
  WizardVendorEntry,
} from "@/app/(private)/(main)/purchase-orders/_components/po-creation-wizard/types";
import type { Requisition } from "@/types/requisition";

// ── Mocks ──────────────────────────────────────────────────────────────────

// jsdom doesn't implement scrollIntoView — mock it to prevent Radix UI errors
window.HTMLElement.prototype.scrollIntoView = vi.fn();

// Mock useVendors so no real API call is made
vi.mock("@/hooks/use-vendor-queries", () => ({
  useVendors: () => ({
    data: [
      { id: "v1", name: "Vendor Alpha" },
      { id: "v2", name: "Vendor Beta" },
      { id: "v3", name: "Vendor Gamma" },
    ],
    isLoading: false,
  }),
  useCreateVendor: () => ({
    mutate: vi.fn(),
    isPending: false,
  }),
}));

// Mock QuotationCollectionSection to avoid deep rendering complexity
vi.mock(
  "@/app/(private)/(main)/requisitions/_components/quotation-collection-section",
  () => ({
    QuotationCollectionSection: ({
      requisitionId,
    }: {
      requisitionId: string;
    }) => <div data-testid={`quotation-section-${requisitionId}`} />,
  }),
);

// ── Stubs ──────────────────────────────────────────────────────────────────

const stubRequisition: Requisition = {
  id: "req-1",
  organizationId: "org-1",
  documentNumber: "REQ-001",
  requesterId: "user-1",
  requesterName: "Test User",
  title: "Test Requisition",
  description: "Test description",
  department: "Finance",
  departmentId: "dept-1",
  status: "APPROVED",
  priority: "MEDIUM",
  items: [],
  totalAmount: 5000,
  currency: "ZMW",
  approvalStage: 1,
  approvalHistory: [],
  categoryName: "",
  preferredVendorName: "",
  isEstimate: false,
  createdAt: new Date(),
  updatedAt: new Date(),
  budgetCode: "",
  requestedByName: "Test User",
  requestedByRole: "user",
  requestedBy: "user-1",
  totalApprovalStages: 1,
  requestedDate: new Date(),
  requiredByDate: new Date("2025-12-31"),
  costCenter: "",
  projectCode: "",
  createdBy: "user-1",
  createdByName: "Test User",
  createdByRole: "user",
};

/** Build N unique WizardVendorEntry objects */
function buildVendors(n: number): WizardVendorEntry[] {
  return Array.from({ length: n }, (_, i) => ({
    localId: `local-${i}`,
    vendorId: `vendor-${i}`,
    vendorName: `Vendor ${i}`,
    quotations: [],
  }));
}

// ── Tests ──────────────────────────────────────────────────────────────────

describe("Property 6: QuotationCollectionSection count matches vendor count", () => {
  afterEach(() => {
    cleanup();
  });

  /**
   * For any N vendors in data.vendors, exactly N [data-testid^="vendor-entry-row-"]
   * elements are rendered.
   *
   * **Validates: Requirements 3.8**
   */
  it("should render exactly N vendor-entry-row elements for N vendors", () => {
    fc.assert(
      fc.property(fc.integer({ min: 0, max: 10 }), (n) => {
        const vendors = buildVendors(n);

        const step2Data: WizardStep2State = {
          vendors,
          selectedVendorLocalId: null,
        };

        const { container } = render(
          <Step2VendorQuotes
            data={step2Data}
            requisition={stubRequisition}
            onChange={vi.fn()}
            onNext={vi.fn()}
            onBack={vi.fn()}
          />,
        );

        // Count elements with data-testid starting with "vendor-entry-row-"
        const rows = container.querySelectorAll(
          "[data-testid^='vendor-entry-row-']",
        );

        const result = rows.length === n;

        cleanup();

        return result;
      }),
      { numRuns: 100 },
    );
  });

  /**
   * When N > 0, the CostComparisonPanel should also be rendered.
   * When N === 0, it should not be rendered.
   *
   * **Validates: Requirements 3.9**
   */
  it("should show CostComparisonPanel only when at least one vendor is present", () => {
    fc.assert(
      fc.property(fc.integer({ min: 0, max: 10 }), (n) => {
        const vendors = buildVendors(n);

        const step2Data: WizardStep2State = {
          vendors,
          selectedVendorLocalId: null,
        };

        const { container } = render(
          <Step2VendorQuotes
            data={step2Data}
            requisition={stubRequisition}
            onChange={vi.fn()}
            onNext={vi.fn()}
            onBack={vi.fn()}
          />,
        );

        // CostComparisonPanel renders a table with "Cost Comparison" heading
        const hasCostComparison =
          container.textContent?.includes("Cost Comparison") ?? false;

        const result = n > 0 ? hasCostComparison : !hasCostComparison;

        cleanup();

        return result;
      }),
      { numRuns: 100 },
    );
  });
});
