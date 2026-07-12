import { describe, it, expect } from "vitest";
import { toDocChain, buildChainLinks } from "@/components/linked-documents";

/**
 * toDocChain adapts the NESTED GetDocumentChain backend response
 * (backend/handlers/document_chain.go buildDocumentChain, ~166-297):
 *   { documentId, documentType, parentDocuments: [...], childDocuments: [...] }
 * into the FLAT DocChain slots buildChainLinks/LinkedDocuments consume.
 *
 * This is the regression test for the empty "Linked Documents" ribbon bug —
 * before the fix, the chain hooks fed buildChainLinks the raw nested
 * response, whose fields never matched DocChain's flat requisitionId/poId/
 * grnId/pvId slots, so the ribbon silently rendered nothing.
 */

describe("toDocChain", () => {
  it("returns an empty object for null/undefined/non-object input", () => {
    expect(toDocChain(undefined)).toEqual({});
    expect(toDocChain(null)).toEqual({});
    expect(toDocChain("not an object")).toEqual({});
    expect(toDocChain(42)).toEqual({});
  });

  it("fills only the self slot when parentDocuments/childDocuments are missing", () => {
    const raw = { documentId: "po-1", documentType: "purchase_order" };
    // Self doc IS filled (documentId/documentType), but since documentNumber
    // and status are absent from the top-level response, the slot is still
    // populated with id but undefined documentNumber/status.
    const chain = toDocChain(raw);
    expect(chain.poId).toBe("po-1");
    expect(chain.poDocumentNumber).toBeUndefined();
    expect(chain.poStatus).toBeUndefined();
    expect(chain.requisitionId).toBeUndefined();
    expect(chain.grnId).toBeUndefined();
    expect(chain.pvId).toBeUndefined();
  });

  it("treats empty parentDocuments/childDocuments arrays as safe no-ops", () => {
    const raw = {
      documentId: "req-1",
      documentType: "requisition",
      parentDocuments: [],
      childDocuments: [],
    };
    const chain = toDocChain(raw);
    expect(chain.requisitionId).toBe("req-1");
    expect(chain.poId).toBeUndefined();
    expect(chain.grnId).toBeUndefined();
    expect(chain.pvId).toBeUndefined();
  });

  it("fills requisition/PO/GRN/PV slots from parent + self + child docs (PO anchor)", () => {
    // Mirrors a real GetDocumentChain(documentType=purchase_order) response:
    // parent REQ, self PO, child GRN + TWO child PVs (multi-partial-payment).
    const raw = {
      documentId: "po-1",
      documentType: "purchase_order",
      parentDocuments: [
        {
          id: "req-1",
          type: "requisition",
          documentNumber: "REQ-001",
          status: "APPROVED",
        },
      ],
      childDocuments: [
        {
          id: "grn-1",
          type: "grn",
          documentNumber: "GRN-001",
          status: "CONFIRMED",
        },
        {
          id: "pv-1",
          type: "payment_voucher",
          documentNumber: "PV-001",
          status: "PAID",
        },
        {
          id: "pv-2",
          type: "payment_voucher",
          documentNumber: "PV-002",
          status: "PENDING",
        },
      ],
      procurementFlow: "goods_first",
    };

    const chain = toDocChain(raw);

    expect(chain).toEqual({
      requisitionId: "req-1",
      requisitionDocumentNumber: "REQ-001",
      requisitionStatus: "APPROVED",
      poId: "po-1",
      poDocumentNumber: undefined,
      poStatus: undefined,
      grnId: "grn-1",
      grnDocumentNumber: "GRN-001",
      grnStatus: "CONFIRMED",
      // Multi-PV: childDocuments is ordered created_at ASC, so the LAST
      // entry (pv-2) is the latest and wins the single pv slot.
      pvId: "pv-2",
      pvDocumentNumber: "PV-002",
      pvStatus: "PENDING",
    });

    // And feeding it through buildChainLinks (as the real pages do) yields
    // the non-empty ribbon this whole fix is about — REQ, GRN, PV, excluding
    // the current document type (purchase-order).
    const links = buildChainLinks(chain, "purchase-order");
    expect(links.map((l) => l.type)).toEqual(["requisition", "grn", "payment-voucher"]);
    expect(links.find((l) => l.type === "payment-voucher")?.id).toBe("pv-2");
  });

  it("tolerates the {success, data} response wrapper as well as the unwrapped payload", () => {
    const unwrapped = {
      documentId: "pv-1",
      documentType: "payment_voucher",
      parentDocuments: [
        { id: "po-1", type: "purchase_order", documentNumber: "PO-001", status: "APPROVED" },
      ],
      childDocuments: [],
    };
    const wrapped = { success: true, data: unwrapped };

    expect(toDocChain(wrapped)).toEqual(toDocChain(unwrapped));
    expect(toDocChain(wrapped).poId).toBe("po-1");
  });

  it("skips chain-doc entries with an unrecognized type or missing id", () => {
    const raw = {
      documentId: "req-1",
      documentType: "requisition",
      parentDocuments: [],
      childDocuments: [
        { id: "", type: "purchase_order", documentNumber: "PO-001" }, // no id
        { id: "po-1", type: "unknown_type", documentNumber: "PO-001" }, // bad type
      ],
    };
    const chain = toDocChain(raw);
    expect(chain.poId).toBeUndefined();
  });

  it("accepts both underscored and hyphenated type strings", () => {
    const raw = {
      documentId: "x",
      documentType: "x",
      parentDocuments: [
        { id: "po-1", type: "purchase-order", documentNumber: "PO-001", status: "DRAFT" },
      ],
      childDocuments: [
        { id: "pv-1", type: "payment-voucher", documentNumber: "PV-001", status: "PAID" },
      ],
    };
    const chain = toDocChain(raw);
    expect(chain.poId).toBe("po-1");
    expect(chain.pvId).toBe("pv-1");
  });
});
