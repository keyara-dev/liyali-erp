"use client";

import { useState } from "react";
import { FileText, Eye, Download, Loader2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { StatusBadge } from "@/components/status-badge";
import { toast } from "sonner";
import { useOrganizationContext } from "@/hooks/use-organization";
import dynamic from "next/dynamic";

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

// Types
import type { PurchaseOrderChain } from "@/types/purchase-order";

const PDFPreviewDialog = dynamic(
  () =>
    import("@/components/modals/pdf-preview-dialog").then(
      (mod) => mod.PDFPreviewDialog,
    ),
  { ssr: false },
);

// ── Types ─────────────────────────────────────────────────────────────────────

interface LinkedDoc {
  type: "requisition" | "purchase-order" | "grn" | "payment-voucher";
  label: string;
  id: string;
  documentNumber: string;
  status?: string;
}

interface LinkedDocumentsPDFSectionProps {
  /** The current PO's source requisition ID */
  sourceRequisitionId?: string;
  /** Chain data from usePurchaseOrderChain */
  chain?: PurchaseOrderChain;
  /** The current PO's own ID — excluded from the linked list */
  currentPoId: string;
}

// ── Component ─────────────────────────────────────────────────────────────────

/**
 * Shows linked procurement chain documents (REQ, GRN, PV) with on-demand
 * PDF preview and download. No files are stored — PDFs are generated
 * client-side from live document data each time the user clicks.
 */
export function LinkedDocumentsPDFSection({
  sourceRequisitionId,
  chain,
  currentPoId,
}: LinkedDocumentsPDFSectionProps) {
  const { currentOrganization } = useOrganizationContext();
  const [loadingId, setLoadingId] = useState<string | null>(null);
  const [previewBlob, setPreviewBlob] = useState<Blob | null>(null);
  const [previewOpen, setPreviewOpen] = useState(false);
  const [previewTitle, setPreviewTitle] = useState("");

  // Build the list of linked documents from available chain data
  const linkedDocs: LinkedDoc[] = [];

  if (sourceRequisitionId) {
    linkedDocs.push({
      type: "requisition",
      label: "Source Requisition",
      id: sourceRequisitionId,
      documentNumber: chain?.requisitionDocumentNumber || sourceRequisitionId,
      status: chain?.requisitionStatus,
    });
  }

  if (chain?.grnId) {
    linkedDocs.push({
      type: "grn",
      label: "Goods Received Note",
      id: chain.grnId,
      documentNumber: chain.grnDocumentNumber || chain.grnId,
      status: chain.grnStatus,
    });
  }

  if (chain?.pvId) {
    linkedDocs.push({
      type: "payment-voucher",
      label: "Payment Voucher",
      id: chain.pvId,
      documentNumber: chain.pvDocumentNumber || chain.pvId,
      status: chain.pvStatus,
    });
  }

  if (linkedDocs.length === 0) return null;

  const docHeader = {
    logoUrl: currentOrganization?.logoUrl,
    orgName: currentOrganization?.name,
    tagline: currentOrganization?.tagline,
  };

  // ── Fetch + generate PDF ───────────────────────────────────────────────────

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
      toast.error(
        err instanceof Error ? err.message : "Failed to generate PDF",
      );
    } finally {
      setLoadingId(null);
    }
  };

  // ── Render ─────────────────────────────────────────────────────────────────

  return (
    <>
      <div className="mt-6 space-y-3">
        <h3 className="text-sm font-semibold flex items-center gap-2">
          <FileText className="h-4 w-4 text-muted-foreground" />
          Linked Procurement Documents
        </h3>
        <p className="text-xs text-muted-foreground">
          Preview or download PDFs for all documents in this procurement chain.
        </p>

        <div className="space-y-2">
          {linkedDocs.map((doc) => {
            const isPreviewing = loadingId === `preview-${doc.id}`;
            const isDownloading = loadingId === `download-${doc.id}`;
            const isBusy = isPreviewing || isDownloading;

            return (
              <div
                key={doc.id}
                className="flex items-center justify-between gap-3 p-3 rounded-lg border bg-muted/20"
              >
                <div className="flex items-center gap-3 min-w-0">
                  <FileText className="h-4 w-4 text-muted-foreground shrink-0" />
                  <div className="min-w-0">
                    <p className="text-xs text-muted-foreground">{doc.label}</p>
                    <p className="text-sm font-medium truncate">
                      {doc.documentNumber}
                    </p>
                  </div>
                  {doc.status && (
                    <StatusBadge status={doc.status} type="document" />
                  )}
                </div>

                <div className="flex items-center gap-2 shrink-0">
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
