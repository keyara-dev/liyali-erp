"use client";

import { useState } from "react";
import Link from "next/link";
import dynamic from "next/dynamic";
import { FileText, Eye, Download, Loader2, ArrowUpRight } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { StatusBadge } from "@/components/status-badge";
import { useOrganizationContext } from "@/hooks/use-organization";

// Actions
import { getRequisitionById } from "@/app/_actions/requisitions";
import { getPurchaseOrderById } from "@/app/_actions/purchase-orders";
import { getGRNAction } from "@/app/_actions/grn-actions";
import { getPaymentVoucherById } from "@/app/_actions/payment-vouchers";

// PDF generators
import {
  getRequisitionPDFBlob,
  getPurchaseOrderPDFBlob,
  getGrnPDFBlob,
  getPaymentVoucherPDFBlob,
  downloadBlob,
} from "@/lib/pdf/pdf-export";

const PDFPreviewDialog = dynamic(
  () =>
    import("@/components/modals/pdf-preview-dialog").then(
      (mod) => mod.PDFPreviewDialog,
    ),
  { ssr: false },
);

// ── Types ───────────────────────────────────────────────────────────────────

export type LinkedDocType =
  | "requisition"
  | "purchase-order"
  | "grn"
  | "payment-voucher";

export interface LinkedDoc {
  type: LinkedDocType;
  /** Row label, e.g. "Purchase Order", "Goods Received Note" */
  label: string;
  id: string;
  documentNumber: string;
  status?: string;
}

interface LinkedDocumentsProps {
  /** Existing links only — an empty list renders nothing. */
  docs: LinkedDoc[];
  /** Section heading. */
  title?: string;
  /** Section sub-text. */
  description?: string;
  /**
   * Requester role: hide the View (navigate) button so they see status +
   * preview/download but cannot jump to documents they don't own.
   */
  showViewLinks?: boolean;
}

// ── Helpers ───────────────────────────────────────────────────────────────────

/**
 * Shape shared by RequisitionChain and PurchaseOrderChain — every chain slot
 * carries an id, a document number and a status. Both interfaces are
 * structurally compatible with this subset.
 */
export interface DocChain {
  requisitionId?: string;
  requisitionDocumentNumber?: string;
  requisitionStatus?: string;
  poId?: string;
  poDocumentNumber?: string;
  poStatus?: string;
  grnId?: string;
  grnDocumentNumber?: string;
  grnStatus?: string;
  pvId?: string;
  pvDocumentNumber?: string;
  pvStatus?: string;
}

/**
 * One entry in the nested GetDocumentChain response's parentDocuments /
 * childDocuments arrays (backend/handlers/document_chain.go
 * buildDocumentChain, ~166-297): `{id, type, documentNumber, status, ...}`.
 * `type` is the backend's underscored form ("purchase_order",
 * "payment_voucher") — never the hyphenated LinkedDocType form.
 */
interface RawChainDoc {
  id?: string;
  type?: string;
  documentNumber?: string;
  status?: string;
  [key: string]: unknown;
}

/**
 * Raw shape of GET /api/v1/document-chain/:id?documentType=... — the
 * unwrapped `data` payload (or, tolerantly, the still-wrapped
 * `{success, data}` response). Every field is nested (parentDocuments /
 * childDocuments arrays); there is no flat poId/grnId/pvId anywhere on it.
 */
interface RawDocumentChain {
  documentId?: string;
  documentType?: string;
  documentNumber?: string;
  status?: string;
  parentDocuments?: RawChainDoc[];
  childDocuments?: RawChainDoc[];
  procurementFlow?: string;
  routingType?: string;
  [key: string]: unknown;
}

/** Maps a backend chain-doc `type` string (either underscored or hyphenated)
 * to the DocChain slot it fills. Returns undefined for anything unrecognized
 * so malformed entries are silently skipped rather than crashing. */
function chainSlotFor(
  type: string | undefined,
): "requisition" | "purchase-order" | "grn" | "payment-voucher" | undefined {
  switch ((type || "").toLowerCase()) {
    case "requisition":
      return "requisition";
    case "purchase_order":
    case "purchase-order":
      return "purchase-order";
    case "grn":
    case "goods_received_note":
      return "grn";
    case "payment_voucher":
    case "payment-voucher":
      return "payment-voucher";
    default:
      return undefined;
  }
}

/**
 * Adapts the NESTED GetDocumentChain response
 * (`{documentId, documentType, parentDocuments: [...], childDocuments: [...]}`)
 * into the FLAT DocChain slots buildChainLinks expects. This is the fix for
 * the empty Linked Documents ribbon: the chain hooks (usePurchaseOrderChain,
 * usePaymentVoucherChain, useDocumentChain) feed buildChainLinks the raw
 * backend response, whose shape never matches DocChain's flat fields, so
 * every slot silently reads as undefined and the ribbon renders nothing.
 *
 * Tolerant of:
 * - the `{success, data}` response wrapper as well as the already-unwrapped
 *   chain object (server actions in this codebase typically unwrap before
 *   returning, but this defends against either);
 * - missing/empty parentDocuments / childDocuments arrays;
 * - multiple documents of the same type in childDocuments (multi-partial-
 *   payments means a PO can have several live PVs) — childDocuments is
 *   ordered created_at ASC, and slots are filled in parent → self → child
 *   order, so the LAST matching entry (the newest) wins a slot. The full
 *   multi-PV list is still available elsewhere (PO detail's linked-PVs list,
 *   the chain-attachments zone); this single-slot ribbon intentionally shows
 *   only the latest.
 * - the anchor document itself: included from documentId/documentType/
 *   documentNumber/status when present, so a slot is still filled if a
 *   caller ever passes a chain to buildChainLinks for a *different*
 *   document's type than the one it was fetched for. buildChainLinks always
 *   filters out `currentType`, so for the normal call pattern (chain fetched
 *   for the same doc whose page renders it) this is inert.
 */
export function toDocChain(raw: unknown): DocChain {
  if (!raw || typeof raw !== "object") return {};

  const maybeWrapped = raw as { success?: boolean; data?: unknown };
  const source: RawDocumentChain =
    maybeWrapped.data && typeof maybeWrapped.data === "object"
      ? (maybeWrapped.data as RawDocumentChain)
      : (raw as RawDocumentChain);

  const chain: DocChain = {};

  const applySlot = (doc: RawChainDoc | undefined) => {
    if (!doc || !doc.id) return;
    const slot = chainSlotFor(doc.type);
    if (!slot) return;

    switch (slot) {
      case "requisition":
        chain.requisitionId = doc.id;
        chain.requisitionDocumentNumber = doc.documentNumber;
        chain.requisitionStatus = doc.status;
        break;
      case "purchase-order":
        chain.poId = doc.id;
        chain.poDocumentNumber = doc.documentNumber;
        chain.poStatus = doc.status;
        break;
      case "grn":
        chain.grnId = doc.id;
        chain.grnDocumentNumber = doc.documentNumber;
        chain.grnStatus = doc.status;
        break;
      case "payment-voucher":
        chain.pvId = doc.id;
        chain.pvDocumentNumber = doc.documentNumber;
        chain.pvStatus = doc.status;
        break;
    }
  };

  const parents = Array.isArray(source.parentDocuments)
    ? source.parentDocuments
    : [];
  const children = Array.isArray(source.childDocuments)
    ? source.childDocuments
    : [];

  parents.forEach(applySlot);
  applySlot({
    id: source.documentId,
    type: source.documentType,
    documentNumber: source.documentNumber,
    status: source.status,
  });
  children.forEach(applySlot);

  return chain;
}

const LABELS: Record<LinkedDocType, string> = {
  requisition: "Requisition",
  "purchase-order": "Purchase Order",
  grn: "Goods Received Note",
  "payment-voucher": "Payment Voucher",
};

const ROUTES: Record<LinkedDocType, string> = {
  requisition: "/requisitions",
  "purchase-order": "/purchase-orders",
  grn: "/grn",
  "payment-voucher": "/payment-vouchers",
};

function hrefFor(doc: LinkedDoc): string {
  return `${ROUTES[doc.type]}/${doc.id}`;
}

/**
 * Map a procurement chain to the linked-document list, skipping the current
 * document and any slot that hasn't been created yet (existing links only).
 */
export function buildChainLinks(
  chain: DocChain | undefined,
  currentType: LinkedDocType,
): LinkedDoc[] {
  if (!chain) return [];

  const slots: Array<{
    type: LinkedDocType;
    id?: string;
    documentNumber?: string;
    status?: string;
  }> = [
    {
      type: "requisition",
      id: chain.requisitionId,
      documentNumber: chain.requisitionDocumentNumber,
      status: chain.requisitionStatus,
    },
    {
      type: "purchase-order",
      id: chain.poId,
      documentNumber: chain.poDocumentNumber,
      status: chain.poStatus,
    },
    {
      type: "grn",
      id: chain.grnId,
      documentNumber: chain.grnDocumentNumber,
      status: chain.grnStatus,
    },
    {
      type: "payment-voucher",
      id: chain.pvId,
      documentNumber: chain.pvDocumentNumber,
      status: chain.pvStatus,
    },
  ];

  return slots
    .filter((s) => s.type !== currentType && !!s.id)
    .map((s) => ({
      type: s.type,
      label: LABELS[s.type],
      id: s.id as string,
      documentNumber: s.documentNumber || (s.id as string),
      status: s.status,
    }));
}

// ── Component ───────────────────────────────────────────────────────────────

/**
 * Unified linked-documents section. For each related document in the
 * procurement chain it offers View (navigate), Preview (inline PDF) and
 * Download (PDF). PDFs are generated client-side from live data on demand —
 * nothing is stored.
 */
export function LinkedDocuments({
  docs,
  title = "Linked Documents",
  description = "View, preview or download related documents in this procurement chain.",
  showViewLinks = true,
}: LinkedDocumentsProps) {
  const { currentOrganization } = useOrganizationContext();
  const [loadingId, setLoadingId] = useState<string | null>(null);
  const [previewBlob, setPreviewBlob] = useState<Blob | null>(null);
  const [previewOpen, setPreviewOpen] = useState(false);
  const [previewTitle, setPreviewTitle] = useState("");

  if (docs.length === 0) return null;

  const docHeader = {
    logoUrl: currentOrganization?.logoUrl,
    orgName: currentOrganization?.name,
    tagline: currentOrganization?.tagline,
  };

  const generatePDFBlob = async (doc: LinkedDoc): Promise<Blob> => {
    switch (doc.type) {
      case "requisition": {
        const res = await getRequisitionById(doc.id);
        if (!res.success || !res.data)
          throw new Error("Failed to load requisition");
        return getRequisitionPDFBlob(res.data as any, docHeader);
      }
      case "purchase-order": {
        const res = await getPurchaseOrderById(doc.id);
        if (!res.success || !res.data)
          throw new Error("Failed to load purchase order");
        return getPurchaseOrderPDFBlob(res.data as any, docHeader);
      }
      case "grn": {
        const res = await getGRNAction(doc.id);
        if (!res.success || !res.data) throw new Error("Failed to load GRN");
        return getGrnPDFBlob(res.data as any, docHeader);
      }
      case "payment-voucher": {
        const res = await getPaymentVoucherById(doc.id);
        if (!res.success || !res.data)
          throw new Error("Failed to load payment voucher");
        return getPaymentVoucherPDFBlob(res.data as any, docHeader);
      }
    }
  };

  const handlePreview = async (doc: LinkedDoc) => {
    setLoadingId(`preview-${doc.id}`);
    try {
      const blob = await generatePDFBlob(doc);
      setPreviewBlob(blob);
      setPreviewTitle(`${doc.label}: ${doc.documentNumber}`);
      setPreviewOpen(true);
    } catch (err) {
      toast.error(
        err instanceof Error ? err.message : "Failed to generate PDF preview",
      );
    } finally {
      setLoadingId(null);
    }
  };

  const handleDownload = async (doc: LinkedDoc) => {
    setLoadingId(`download-${doc.id}`);
    try {
      const blob = await generatePDFBlob(doc);
      downloadBlob(blob, `${doc.documentNumber}.pdf`);
      toast.success(`${doc.documentNumber}.pdf downloaded`);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to generate PDF");
    } finally {
      setLoadingId(null);
    }
  };

  return (
    <>
      <div className="space-y-3">
        <div className="space-y-1">
          <h3 className="text-sm font-semibold flex items-center gap-2">
            <FileText className="h-4 w-4 text-muted-foreground" />
            {title}
          </h3>
          <p className="text-xs text-muted-foreground">{description}</p>
        </div>

        <div className="space-y-2">
          {docs.map((doc) => {
            const isPreviewing = loadingId === `preview-${doc.id}`;
            const isDownloading = loadingId === `download-${doc.id}`;
            const isBusy = isPreviewing || isDownloading;

            return (
              <div
                key={`${doc.type}-${doc.id}`}
                className="flex flex-col gap-3 rounded-lg border bg-muted/20 p-3 sm:flex-row sm:items-center sm:justify-between"
              >
                <div className="flex min-w-0 items-center gap-3">
                  <FileText className="h-4 w-4 shrink-0 text-muted-foreground" />
                  <div className="min-w-0">
                    <p className="text-xs text-muted-foreground">{doc.label}</p>
                    <p className="truncate text-sm font-medium font-mono">
                      {doc.documentNumber}
                    </p>
                  </div>
                  {doc.status && (
                    <StatusBadge status={doc.status} type="document" />
                  )}
                </div>

                <div className="grid grid-cols-3 gap-2 sm:flex sm:shrink-0 sm:items-center">
                  {showViewLinks && (
                    <Button
                      asChild
                      variant="outline"
                      size="sm"
                      className="gap-1.5"
                    >
                      <Link href={hrefFor(doc)}>
                        <ArrowUpRight className="h-3.5 w-3.5" />
                        View
                      </Link>
                    </Button>
                  )}
                  <Button
                    variant="outline"
                    size="sm"
                    disabled={isBusy}
                    onClick={() => handlePreview(doc)}
                    className="gap-1.5"
                  >
                    {isPreviewing ? (
                      <Loader2 className="h-3.5 w-3.5 animate-spin" />
                    ) : (
                      <Eye className="h-3.5 w-3.5" />
                    )}
                    Preview
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    disabled={isBusy}
                    onClick={() => handleDownload(doc)}
                    className="gap-1.5"
                  >
                    {isDownloading ? (
                      <Loader2 className="h-3.5 w-3.5 animate-spin" />
                    ) : (
                      <Download className="h-3.5 w-3.5" />
                    )}
                    Download
                  </Button>
                </div>
              </div>
            );
          })}
        </div>
      </div>

      {previewBlob && (
        <PDFPreviewDialog
          open={previewOpen}
          onOpenChange={setPreviewOpen}
          pdfBlob={previewBlob}
          fileName={previewTitle}
          onDownload={async () => {
            if (previewBlob) downloadBlob(previewBlob, `${previewTitle}.pdf`);
          }}
        />
      )}
    </>
  );
}
