"use client";

import { useMemo, useRef, useState } from "react";
import Link from "next/link";
import { useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import {
  Paperclip,
  Upload,
  Loader2,
  AlertCircle,
  FileText,
  ImageIcon,
  Eye,
  Download,
  Trash2,
  ArrowUpRight,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components";
import {
  Empty,
  EmptyContent,
  EmptyDescription,
  EmptyMedia,
} from "@/components/ui/empty";
import { cn } from "@/lib/utils";
import { uploadToImageKit } from "@/lib/imagekit";
import { LinkedDocuments, type LinkedDoc } from "@/components/linked-documents";
import { useChainAttachments } from "@/hooks/use-document-chain-queries";
import type {
  ChainAttachment,
  ChainDocumentType,
} from "@/app/_actions/document-chain";

// ── Types ───────────────────────────────────────────────────────────────────

/**
 * Result of an in-browser ImageKit upload, shaped the same as
 * RequisitionAttachment / PurchaseOrderAttachment so pages can merge it
 * straight into their own metadata.attachments array.
 */
export interface UploadedAttachment {
  fileId: string;
  fileName: string;
  fileUrl: string;
  fileSize: number;
  mimeType: string;
  uploadedAt: string;
}

export interface SupportingDocumentsProps {
  /** ID of the document this component is mounted on (REQ/PO/GRN/PV). */
  documentId: string;
  /** Type of the document this component is mounted on. */
  documentType: ChainDocumentType;
  /** Chain-doc rows (Req/PO/GRN/PV) — existing client-generated-PDF
   * View/Preview/Download mechanism, unchanged. Empty list renders nothing
   * for this zone. */
  chainDocs: LinkedDoc[];
  /** Whether the current user may upload a new supporting document here. */
  canUpload: boolean;
  /** Called with the freshly-uploaded file after ImageKit upload succeeds.
   * The PAGE is responsible for persisting it into ITS OWN document's
   * metadata.attachments (read-modify-write) — this component only uploads
   * the file and hands back the result. */
  onUpload?: (att: UploadedAttachment) => Promise<void>;
  /** Whether a given aggregated file may be deleted by the current user.
   * Only offer deletion for files that live on THIS document's own metadata
   * (not files copied in from another document in the chain). */
  canDeleteFile?: (att: ChainAttachment) => boolean;
  /** Called with the fileId to delete when the user confirms removal. */
  onDeleteFile?: (fileId: string) => Promise<void>;
  /** Requester role etc: hide "View" navigation to documents they don't own. */
  showViewLinks?: boolean;
}

interface AttachmentGroup {
  sourceDocType: string;
  sourceDocId: string;
  sourceDocNumber: string;
  items: ChainAttachment[];
}

// ── Helpers ───────────────────────────────────────────────────────────────────

const SOURCE_LABELS: Record<string, string> = {
  requisition: "Requisition",
  purchase_order: "Purchase Order",
  grn: "Goods Received Note",
  payment_voucher: "Payment Voucher",
};

const SOURCE_ROUTES: Record<string, string> = {
  requisition: "/requisitions",
  purchase_order: "/purchase-orders",
  grn: "/grn",
  payment_voucher: "/payment-vouchers",
};

const KIND_META: Record<
  ChainAttachment["kind"],
  { label: string; className: string }
> = {
  attachment: {
    label: "Attachment",
    className: "bg-muted text-muted-foreground border-transparent",
  },
  quotation: {
    label: "Quotation",
    className:
      "bg-blue-100 text-blue-700 border-blue-200 dark:bg-blue-950/40 dark:text-blue-300 dark:border-blue-800",
  },
  proof_of_payment: {
    label: "Proof of Payment",
    className:
      "bg-emerald-100 text-emerald-700 border-emerald-200 dark:bg-emerald-950/40 dark:text-emerald-300 dark:border-emerald-800",
  },
};

function formatBytes(bytes?: number): string {
  if (!bytes) return "";
  const k = 1024;
  const sizes = ["B", "KB", "MB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(1))} ${sizes[i]}`;
}

function groupBySourceDoc(attachments: ChainAttachment[]): AttachmentGroup[] {
  const map = new Map<string, AttachmentGroup>();
  for (const att of attachments) {
    const key = `${att.sourceDocType}:${att.sourceDocId}`;
    let group = map.get(key);
    if (!group) {
      group = {
        sourceDocType: att.sourceDocType,
        sourceDocId: att.sourceDocId,
        sourceDocNumber: att.sourceDocNumber,
        items: [],
      };
      map.set(key, group);
    }
    group.items.push(att);
  }
  return Array.from(map.values());
}

// ── Component ───────────────────────────────────────────────────────────────

/**
 * Unified supporting-documents section for the 4 procurement detail pages.
 * Three zones:
 *  (a) chain-doc rows (Req/PO/GRN/PV) via the existing LinkedDocuments
 *      View/Preview/Download mechanism, unchanged;
 *  (b) every supporting-file attachment aggregated across the whole
 *      procurement chain (requisition, PO, GRN(s), PV(s)), grouped by
 *      source document, including each PV's proof of payment; and
 *  (c) an upload button (when canUpload) that uploads straight to ImageKit
 *      and hands the result back to the page to persist on its own document.
 */
export function SupportingDocuments({
  documentId,
  documentType,
  chainDocs,
  canUpload,
  onUpload,
  canDeleteFile,
  onDeleteFile,
  showViewLinks = true,
}: SupportingDocumentsProps) {
  const queryClient = useQueryClient();
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [isUploading, setIsUploading] = useState(false);
  const [deletingId, setDeletingId] = useState<string | null>(null);

  const {
    data,
    isLoading,
    isError,
    refetch,
  } = useChainAttachments(documentId, documentType);

  const attachments = data?.attachments ?? [];
  const groups = useMemo(() => groupBySourceDoc(attachments), [attachments]);

  const queryKey = ["document-chain", documentId, "attachments"];

  const handleFileSelected = async (
    e: React.ChangeEvent<HTMLInputElement>,
  ) => {
    const file = e.target.files?.[0];
    if (!file) return;
    setIsUploading(true);
    try {
      const result = await uploadToImageKit(file, `${documentType}/attachments`);
      const uploaded: UploadedAttachment = {
        fileId: result.fileId,
        fileName: result.name,
        fileUrl: result.url,
        fileSize: result.size,
        mimeType: file.type,
        uploadedAt: new Date().toISOString(),
      };
      await onUpload?.(uploaded);
      queryClient.invalidateQueries({ queryKey });
      toast.success("Document uploaded successfully");
    } catch (err) {
      toast.error(
        err instanceof Error ? err.message : "Failed to upload document",
      );
    } finally {
      setIsUploading(false);
      if (fileInputRef.current) fileInputRef.current.value = "";
    }
  };

  const handleDownload = (att: ChainAttachment) => {
    if (!att.fileUrl) return;
    const link = document.createElement("a");
    link.href = att.fileUrl;
    link.download = att.fileName;
    link.target = "_blank";
    link.rel = "noopener noreferrer";
    link.click();
  };

  const handleDelete = async (fileId: string) => {
    if (!onDeleteFile) return;
    setDeletingId(fileId);
    try {
      await onDeleteFile(fileId);
      queryClient.invalidateQueries({ queryKey });
      toast.success("Document removed");
    } catch (err) {
      toast.error(
        err instanceof Error ? err.message : "Failed to remove document",
      );
    } finally {
      setDeletingId(null);
    }
  };

  return (
    <div className="space-y-6">
      {/* Zone (a): chain-doc rows — View / Preview / Download, unchanged */}
      <LinkedDocuments docs={chainDocs} showViewLinks={showViewLinks} />

      {/* Zone (b) + (c): chain-wide aggregated files + upload */}
      <div className="space-y-3">
        <div className="flex items-center justify-between gap-2 flex-wrap">
          <div className="space-y-1">
            <h3 className="text-sm font-semibold flex items-center gap-2">
              <Paperclip className="h-4 w-4 text-muted-foreground" />
              Supporting Documents
              {attachments.length > 0 && (
                <span className="text-xs font-normal text-muted-foreground">
                  ({attachments.length})
                </span>
              )}
            </h3>
            <p className="text-xs text-muted-foreground">
              Files uploaded anywhere in this procurement chain.
            </p>
          </div>
          {canUpload && (
            <>
              <input
                ref={fileInputRef}
                type="file"
                className="hidden"
                accept="application/pdf,image/*"
                onChange={handleFileSelected}
              />
              <Button
                variant="outline"
                size="sm"
                className="gap-2"
                disabled={isUploading}
                isLoading={isUploading}
                onClick={() => fileInputRef.current?.click()}
              >
                <Upload className="h-4 w-4" />
                Upload Document
              </Button>
            </>
          )}
        </div>

        {isLoading ? (
          <div className="flex items-center justify-center gap-2 py-8 text-sm text-muted-foreground">
            <Loader2 className="h-4 w-4 animate-spin" />
            Loading supporting documents...
          </div>
        ) : isError ? (
          <div className="text-center py-8 text-red-500">
            <AlertCircle className="h-6 w-6 mx-auto mb-2" />
            <p className="text-sm">Failed to load supporting documents</p>
            <button
              onClick={() => refetch()}
              className="mt-2 text-xs text-blue-600 hover:underline"
            >
              Try again
            </button>
          </div>
        ) : groups.length === 0 ? (
          <Empty>
            <EmptyMedia variant="icon">
              <Paperclip className="h-6 w-6" />
            </EmptyMedia>
            <EmptyContent>
              <EmptyDescription>
                {canUpload
                  ? "No supporting documents yet — upload one above"
                  : "No supporting documents attached"}
              </EmptyDescription>
            </EmptyContent>
          </Empty>
        ) : (
          <div className="space-y-3">
            {groups.map((group) => (
              <div
                key={`${group.sourceDocType}-${group.sourceDocId}`}
                className="rounded-lg border overflow-hidden"
              >
                <div className="flex items-center gap-2 px-3 py-2 bg-muted/40 border-b">
                  <FileText className="h-3.5 w-3.5 text-muted-foreground shrink-0" />
                  <span className="text-xs font-medium text-muted-foreground">
                    {SOURCE_LABELS[group.sourceDocType] ??
                      group.sourceDocType}
                  </span>
                  {showViewLinks && SOURCE_ROUTES[group.sourceDocType] ? (
                    <Link
                      href={`${SOURCE_ROUTES[group.sourceDocType]}/${group.sourceDocId}`}
                      className="text-xs font-mono font-medium hover:underline"
                    >
                      {group.sourceDocNumber}
                    </Link>
                  ) : (
                    <span className="text-xs font-mono font-medium">
                      {group.sourceDocNumber}
                    </span>
                  )}
                </div>
                <div className="divide-y">
                  {group.items.map((att) => {
                    const kindMeta = KIND_META[att.kind];
                    const isPDF = att.mimeType === "application/pdf";
                    const isImage = att.mimeType?.startsWith("image/");
                    const canDelete =
                      !!att.fileId && !!canDeleteFile?.(att) && !!onDeleteFile;
                    const isDeleting = deletingId === att.fileId;

                    return (
                      <div
                        key={`${att.sourceDocId}-${att.fileId || att.fileName}`}
                        className="flex items-center gap-3 p-3 hover:bg-muted/30 transition"
                      >
                        <div className="shrink-0">
                          {isPDF ? (
                            <FileText className="h-5 w-5 text-red-500" />
                          ) : isImage ? (
                            <ImageIcon className="h-5 w-5 text-blue-500" />
                          ) : (
                            <FileText className="h-5 w-5 text-muted-foreground" />
                          )}
                        </div>
                        <div className="min-w-0 flex-1">
                          <p className="text-sm font-medium truncate">
                            {att.fileName}
                          </p>
                          <div className="flex items-center gap-1.5 flex-wrap mt-1">
                            <Badge
                              variant="outline"
                              className={cn(
                                "text-[10px] px-1.5 py-0 h-4 font-medium",
                                kindMeta.className,
                              )}
                            >
                              {kindMeta.label}
                            </Badge>
                            {att.fromRequisition && (
                              <span className="text-[10px] px-1.5 py-0.5 rounded bg-amber-100 text-amber-700 font-medium">
                                From Requisition
                              </span>
                            )}
                            {att.category && att.category !== "quotation" && (
                              <span className="text-[10px] px-1.5 py-0.5 rounded bg-muted text-muted-foreground capitalize">
                                {att.category.replace(/_/g, " ")}
                              </span>
                            )}
                            {att.fileSize ? (
                              <span className="text-[10px] text-muted-foreground">
                                {formatBytes(att.fileSize)}
                              </span>
                            ) : null}
                          </div>
                        </div>
                        <div className="flex items-center gap-1 shrink-0">
                          {att.fileUrl ? (
                            <>
                              <Button
                                variant="ghost"
                                size="icon"
                                className="h-8 w-8"
                                title="Preview"
                                onClick={() =>
                                  window.open(
                                    att.fileUrl,
                                    "_blank",
                                    "noopener,noreferrer",
                                  )
                                }
                              >
                                <Eye className="h-4 w-4" />
                              </Button>
                              <Button
                                variant="ghost"
                                size="icon"
                                className="h-8 w-8"
                                title="Download"
                                onClick={() => handleDownload(att)}
                              >
                                <Download className="h-4 w-4" />
                              </Button>
                            </>
                          ) : att.downloadRef ? (
                            <Button
                              asChild
                              variant="outline"
                              size="sm"
                              className="gap-1.5 text-xs h-7"
                            >
                              <Link href={att.downloadRef}>
                                <ArrowUpRight className="h-3.5 w-3.5" />
                                View on PV
                              </Link>
                            </Button>
                          ) : null}
                          {canDelete && (
                            <Button
                              variant="ghost"
                              size="icon"
                              className="h-8 w-8 text-muted-foreground/50 hover:text-red-500"
                              title="Remove document"
                              disabled={isDeleting}
                              onClick={() => handleDelete(att.fileId!)}
                            >
                              {isDeleting ? (
                                <Loader2 className="h-4 w-4 animate-spin" />
                              ) : (
                                <Trash2 className="h-4 w-4" />
                              )}
                            </Button>
                          )}
                        </div>
                      </div>
                    );
                  })}
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
