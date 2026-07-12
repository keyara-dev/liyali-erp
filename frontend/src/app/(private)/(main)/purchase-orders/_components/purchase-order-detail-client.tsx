"use client";

import { useState } from "react";
import Link from "next/link";
import { Card } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  Send,
  AlertCircle,
  Download,
  Eye,
  Pencil,
  Calendar,
  Building,
  DollarSign,
  Clock,
  Tag,
  FileText,
  Undo2,
  Paperclip,
  ShoppingCart,
  CheckSquare,
  GitBranch,
  Activity,
  TrendingUp,
  TrendingDown,
  Minus,
  AlertTriangle,
  Truck,
  Info,
  Wallet,
} from "lucide-react";
import { PageHeader } from "@/components/base/page-header";
import { PurchaseOrderItemsList } from "./purchase-order-items-list";
import { POItemsEditor } from "./po-items-editor";
import { PurchaseOrder, PurchaseOrderAttachment } from "@/types/purchase-order";
import {
  ActivityLogContent,
  ApprovalChainContent,
  ApprovalActionContent,
  WorkflowStatusSummary,
} from "@/app/(private)/(main)/requisitions/_components/approval-history-panel";
import {
  Empty,
  EmptyContent,
  EmptyDescription,
  EmptyMedia,
} from "@/components/ui/empty";
import { Package } from "lucide-react";
import dynamic from "next/dynamic";

const PDFPreviewDialog = dynamic(
  () =>
    import("@/components/modals/pdf-preview-dialog").then(
      (mod) => mod.PDFPreviewDialog,
    ),
  { ssr: false },
);

import { PurchaseOrderSubmitDialog } from "./purchase-order-submit-dialog";
import { ConfirmationModal } from "@/components/modals/confirmation-modal";
import { Alert, AlertDescription } from "@/components/ui/alert";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { QuotationCollectionSection } from "@/app/(private)/(main)/requisitions/_components/quotation-collection-section";
import { POShippingEditor } from "./po-shipping-editor";
import { buildChainLinks } from "@/components/linked-documents";
import {
  SupportingDocuments,
  type UploadedAttachment,
} from "@/components/supporting-documents";
import type { ChainAttachment } from "@/app/_actions/document-chain";
import { useVendors } from "@/hooks/use-vendor-queries";
import {
  VendorComplianceWarning,
  deriveVendorComplianceWarnings,
} from "@/components/vendor-compliance-warning";
import type { Quotation } from "@/types/core";
import { Badge } from "@/components";
import { DocumentLoadingPage } from "@/components/base/document-loading-page";
import ErrorDisplay from "@/components/base/error-display";
import { usePurchaseOrderDetail } from "@/hooks/use-purchase-order-detail";
import { usePermissions } from "@/hooks/use-permissions";
import { useRouter } from "next/navigation";
import { useQuery } from "@tanstack/react-query";
import { getAuditEvents, type AuditEvent } from "@/app/_actions/audit";
import { getRequisitionById } from "@/app/_actions/requisitions";
import { updatePurchaseOrder } from "@/app/_actions/purchase-orders";
import { createPaymentVoucherFromPurchaseOrder } from "@/app/_actions/payment-vouchers";
import { toast } from "sonner";
import { formatCurrency } from "@/lib/utils";
import { EditPurchaseOrderDialog } from "./edit-purchase-order-dialog";
import {
  CreatePVFromPODialog,
  type CreatePVFromPOOptions,
} from "@/app/(private)/(main)/payment-vouchers/_components/create-pv-from-po-dialog";
import { poRemainingBalance, canCreateAnotherPV } from "@/lib/payment-utils";

/**
 * Props for the PurchaseOrderDetailClient component
 */
interface PurchaseOrderDetailClientProps {
  /** Purchase Order ID */
  purchaseOrderId: string;
  /** Current user ID */
  userId: string;
  /** Current user role */
  userRole: string;
  /** Optional initial PO data from server-side rendering */
  initialPurchaseOrder?: PurchaseOrder;
}

/**
 * Main client component for Purchase Order detail page
 *
 * Manages all UI state and interactions for the PO detail page including:
 * - Displaying PO metadata and items
 * - Handling submission for approval
 * - Managing approval workflow interactions
 * - Displaying approval chain and activity log
 * - PDF preview and export
 * - Attachment preview
 * - Permission-based action buttons
 *
 * This component follows the same pattern as the Requisition detail page
 * for consistency across document types.
 *
 * @param props - Component props
 * @param props.purchaseOrderId - Purchase Order ID to display
 * @param props.userId - Current user ID for permission checks
 * @param props.userRole - Current user role for permission checks
 * @param props.initialPurchaseOrder - Optional initial PO data from server
 *
 * @example
 * ```tsx
 * <PurchaseOrderDetailClient
 *   purchaseOrderId="po-123"
 *   userId="user-456"
 *   userRole="PROCUREMENT_OFFICER"
 *   initialPurchaseOrder={serverPO}
 * />
 * ```
 *
 * **Validates: Requirements 6.1, 11.6, 12.1, 12.5, 12.6**
 */
export function PurchaseOrderDetailClient({
  purchaseOrderId,
  userId,
  userRole,
  initialPurchaseOrder,
}: PurchaseOrderDetailClientProps) {
  const router = useRouter();
  const [editingItems, setEditingItems] = useState(false);
  const [isCreatePVDialogOpen, setIsCreatePVDialogOpen] = useState(false);
  const [isCreatingPV, setIsCreatingPV] = useState(false);
  const { data: vendors = [] } = useVendors({ active: true });
  const { hasPermission } = usePermissions();

  // Use the custom hook to manage all document detail logic
  // This hook handles data fetching, mutations, UI state, and permissions
  const { data: auditEventsData } = useQuery({
    queryKey: ["audit-events", "purchase_order", purchaseOrderId],
    queryFn: async () => {
      const res = await getAuditEvents("purchase_order", purchaseOrderId);
      return res.success ? ((res.data as AuditEvent[]) ?? []) : [];
    },
    enabled: !!purchaseOrderId,
  });

  const {
    document: purchaseOrder,
    isLoading,
    chain,
    approvalData,
    isExporting,
    previewOpen,
    setPreviewOpen,
    previewBlob,
    isEditDialogOpen,
    setIsEditDialogOpen,
    showSubmitDialog,
    setShowSubmitDialog,
    showWithdrawModal,
    setShowWithdrawModal,
    handlePreviewPDF,
    handleExportPDF,
    handleSubmitForApproval: handleSubmit,
    handleEdit,
    handleDocumentUpdated,
    handleWithdraw,
    handleApprovalComplete,
    permissions,
    submitMutation,
    withdrawMutation,
  } = usePurchaseOrderDetail({
    poId: purchaseOrderId,
    userId,
    userRole,
    initialPurchaseOrder,
  });

  // Fetch linked requisition items when PO was created from a requisition
  const { data: linkedRequisitionData } = useQuery({
    queryKey: ["requisition", purchaseOrder?.sourceRequisitionId],
    queryFn: async () => {
      if (!purchaseOrder?.sourceRequisitionId) return null;
      const res = await getRequisitionById(purchaseOrder.sourceRequisitionId);
      return res.success ? res.data : null;
    },
    enabled: !!purchaseOrder?.sourceRequisitionId,
  });
  const linkedRequisitionItems = linkedRequisitionData?.items;

  // Show loading state while fetching initial data
  if (isLoading) return <DocumentLoadingPage />;

  // Show error state if PO not found
  if (!purchaseOrder)
    return (
      <ErrorDisplay
        title="Purchase Order Not Found"
        message="The purchase order you're looking for doesn't exist."
        showBackButton
      />
    );

  // Extract attachments: merge PO's own + REQ's (tagged fromRequisition)
  const attachments: PurchaseOrderAttachment[] =
    (purchaseOrder.metadata?.attachments as PurchaseOrderAttachment[]) || [];

  // Extract quotations from PO metadata
  const quotations: Quotation[] =
    (purchaseOrder.metadata?.quotations as Quotation[]) ?? [];

  // Chain-doc rows (Req/PO/GRN/PV) fed into SupportingDocuments' zone (a) —
  // mechanism unchanged from the old standalone LinkedDocuments mount.
  const poChainDocs = (() => {
    const links = buildChainLinks(chain, "purchase-order");
    // Chain may omit the source requisition — backfill from the PO.
    if (
      !links.some((l) => l.type === "requisition") &&
      purchaseOrder.sourceRequisitionId
    ) {
      links.unshift({
        type: "requisition",
        label: "Requisition",
        id: purchaseOrder.sourceRequisitionId,
        documentNumber:
          linkedRequisitionData?.documentNumber ||
          purchaseOrder.sourceRequisitionId,
        status: linkedRequisitionData?.status,
      });
    }
    return links;
  })();

  const isDraft = purchaseOrder.status?.toUpperCase() === "DRAFT";

  // Look up full vendor details from the vendors list
  const vendorDetails = vendors.find((v) => v.id === purchaseOrder.vendorId);

  // Vendor compliance (warn-only): prefer the live backend-computed warnings
  // on the PO response (present only when non-empty); fall back to deriving
  // from the vendor record's ZRA TPIN / PACRA fields.
  const complianceWarnings =
    purchaseOrder.complianceWarnings ??
    deriveVendorComplianceWarnings(vendorDetails);

  const canEditQuotations = isDraft;

  // Line-item editing authority mirrors the backend `/items` endpoint, which
  // gates on `purchase_order:edit` + DRAFT status (no creator/owner scope).
  // So any authorized user — not just the literal creator — may reconcile
  // unit price / quantity on a DRAFT PO to reach zero variance against the REQ.
  const canEditItems =
    isDraft && (permissions.canEdit || hasPermission("purchase_order", "edit"));

  // Toolbar "Edit" (metadata dialog: title/dept/priority/budget/vendor). Backend
  // PUT /:id allows privileged/procurement (CanViewAll/IsProcurement) to edit any
  // org PO — only super_admin/admin/finance hold purchase_order:edit and all are
  // CanViewAll, so the permission proxy never 404s. Editable statuses match the
  // dialog's own guard (DRAFT/REJECTED).
  const isEditableStatus =
    isDraft || purchaseOrder.status?.toUpperCase() === "REJECTED";
  const canEditPO =
    permissions.canEdit ||
    (isEditableStatus && hasPermission("purchase_order", "edit"));

  // Persists a SupportingDocuments upload (already on ImageKit) into the PO's
  // own metadata.attachments — read-modify-write, keeping REQ-copied entries
  // last so the PO's own uploads sort first. SupportingDocuments does the
  // ImageKit upload + success/error toast itself; this only persists.
  const handleSupportingDocUpload = async (att: UploadedAttachment) => {
    const existingOwn = attachments.filter((a) => !a.fromRequisition);
    const fromReq = attachments.filter((a) => a.fromRequisition);
    const merged = [...existingOwn, att, ...fromReq];
    await updatePurchaseOrder({
      purchaseOrderId: purchaseOrderId,
      poId: purchaseOrderId,
      metadata: { ...purchaseOrder.metadata, attachments: merged },
    });
    handleDocumentUpdated();
  };

  const handleDeleteAttachment = async (fileId: string) => {
    const updated = attachments.filter((a) => a.fileId !== fileId);
    await updatePurchaseOrder({
      purchaseOrderId: purchaseOrderId,
      poId: purchaseOrderId,
      metadata: { ...purchaseOrder.metadata, attachments: updated },
    });
    handleDocumentUpdated();
  };

  // Own-doc, non-copied files only — mirrors the old tab's isDraft gate.
  const canDeleteAttachment = (att: ChainAttachment) =>
    isDraft &&
    att.sourceDocType === "purchase_order" &&
    att.sourceDocId === purchaseOrderId &&
    !att.fromRequisition;

  const handleSaveQuotations = async (updated: Quotation[]) => {
    await updatePurchaseOrder({
      purchaseOrderId,
      poId: purchaseOrderId,
      metadata: { ...purchaseOrder.metadata, quotations: updated },
    });
    handleDocumentUpdated();
  };

  const handleSelectVendor = async (
    vendorId: string,
    vendorName: string,
    amount: number,
    fileUrl: string,
  ) => {
    await updatePurchaseOrder({
      purchaseOrderId,
      poId: purchaseOrderId,
      vendorId,
      vendorName,
      totalAmount: amount,
      metadata: {
        ...purchaseOrder.metadata,
        selectedQuotationFileUrl: fileUrl,
      },
    });
    handleDocumentUpdated();
  };

  const handleConfirmCreatePV = async ({
    workflowId,
    vendorId,
    vendorName,
    linkedGRNDocumentNumber,
    amount,
    paymentType,
    narration,
  }: CreatePVFromPOOptions) => {
    setIsCreatingPV(true);
    try {
      const response = await createPaymentVoucherFromPurchaseOrder(
        purchaseOrder,
        workflowId,
        vendorId,
        vendorName,
        linkedGRNDocumentNumber,
        amount,
        paymentType,
        narration,
      );
      if (response.success && response.data) {
        toast.success("Payment Voucher created successfully");
        setIsCreatePVDialogOpen(false);
        handleDocumentUpdated();
        router.push(`/payment-vouchers/${response.data.id}`);
      } else {
        toast.error(response.message || "Failed to create Payment Voucher");
      }
    } catch (error) {
      console.error("Error creating PV:", error);
      toast.error("An error occurred while creating the Payment Voucher");
    } finally {
      setIsCreatingPV(false);
    }
  };

  /**
   * Custom submit handler that passes additional PO metadata
   * This ensures the submission includes submitter information
   */
  const handleSubmitForApproval = async (
    workflowId: string,
    comments?: string,
  ) => {
    await handleSubmit(workflowId, comments, {
      submittedBy: userId,
      submittedByName: purchaseOrder.requestedByName || "User",
      submittedByRole: purchaseOrder.requestedByRole || userRole,
    });
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <PageHeader
        title={purchaseOrder.documentNumber}
        subtitle={`${purchaseOrder.title || "Untitled Purchase Order"} • Created ${new Date(purchaseOrder.createdAt).toLocaleDateString("en-ZM", { year: "numeric", month: "long", day: "numeric" })}${purchaseOrder.updatedAt && new Date(purchaseOrder.updatedAt).getTime() !== new Date(purchaseOrder.createdAt).getTime() ? ` • Updated ${new Date(purchaseOrder.updatedAt).toLocaleDateString("en-ZM", { year: "numeric", month: "long", day: "numeric" })}` : ""}`}
        badges={[{ status: purchaseOrder.status, type: "document" }]}
        onBackClick={() => router.back()}
        showBackButton={true}
        actions={
          <>
            <Button
              onClick={handlePreviewPDF}
              disabled={isExporting}
              variant="outline"
              className="gap-2 h-9"
            >
              <Eye className="h-4 w-4" />
              Preview
            </Button>
            <Button
              onClick={handleExportPDF}
              disabled={isExporting}
              isLoading={isExporting}
              loadingText="Exporting..."
              variant="outline"
              className="gap-2 h-9"
            >
              <Download className="h-4 w-4" />
              Export PDF
            </Button>
            {canEditPO && (
              <Button
                onClick={handleEdit}
                variant="outline"
                className="gap-2 h-9"
              >
                <Pencil className="h-4 w-4" />
                Edit
              </Button>
            )}
            {permissions.canSubmit && (
              <Button
                onClick={() => setShowSubmitDialog(true)}
                className="gap-2 h-9"
              >
                <Send className="h-4 w-4" />
                Submit for Approval
              </Button>
            )}
            {permissions.canWithdraw && (
              <Button
                onClick={() => setShowWithdrawModal(true)}
                variant="outline"
                className="gap-2 h-9 text-amber-600 border-amber-300 hover:bg-amber-50"
              >
                <Undo2 className="h-4 w-4" />
                Withdraw
              </Button>
            )}
          </>
        }
      />

      {/* Purchase Order Details Card */}
      <div className="gradient-primary border-0 overflow-hidden rounded-lg p-6">
        <h2 className="text-lg font-semibold mb-6 text-primary-foreground">
          Purchase Order Details
        </h2>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          <div className="space-y-1">
            <label className="text-xs font-semibold text-primary-foreground/80 uppercase tracking-wider flex items-center gap-1">
              <FileText className="h-3 w-3" />
              Title
            </label>
            <p className="text-base font-medium text-primary-foreground">
              {purchaseOrder.title || "—"}
            </p>
          </div>

          <div className="space-y-1">
            <label className="text-xs font-semibold text-primary-foreground/80 uppercase tracking-wider flex items-center gap-1">
              <Building className="h-3 w-3" />
              Vendor
            </label>
            <p className="text-base font-medium text-primary-foreground">
              {purchaseOrder.vendorName || "—"}
            </p>
          </div>

          <div className="space-y-1">
            <label className="text-xs font-semibold text-primary-foreground/80 uppercase tracking-wider flex items-center gap-1">
              <Building className="h-3 w-3" />
              Department
            </label>
            <p className="text-base font-medium text-primary-foreground">
              {purchaseOrder.department || "—"}
            </p>
          </div>

          <div className="space-y-1">
            <label className="text-xs font-semibold text-primary-foreground/80 uppercase tracking-wider flex items-center gap-1">
              <AlertCircle className="h-3 w-3" />
              Priority
            </label>
            <div className="flex items-center">
              <Badge
                className={`inline-flex capitalize items-center px-2 py-1 rounded-full text-xs font-medium border ${
                  purchaseOrder.priority?.toUpperCase() === "URGENT"
                    ? "bg-red-100 text-red-800 border-red-200"
                    : purchaseOrder.priority?.toUpperCase() === "HIGH"
                      ? "bg-orange-100 text-orange-800 border-orange-200"
                      : purchaseOrder.priority?.toUpperCase() === "MEDIUM"
                        ? "bg-blue-100 text-blue-800 border-blue-200"
                        : "bg-gray-100 text-gray-800 border-gray-200"
                }`}
              >
                {purchaseOrder.priority || "Medium"}
              </Badge>
            </div>
          </div>

          <div className="space-y-1">
            <label className="text-xs font-semibold text-primary-foreground/80 uppercase tracking-wider flex items-center gap-1">
              <FileText className="h-3 w-3" />
              REQ Estimated Cost
            </label>
            <p className="text-base font-medium text-primary-foreground">
              {(() => {
                const reqEstimated =
                  purchaseOrder.estimatedCost ||
                  linkedRequisitionData?.totalAmount ||
                  0;
                return reqEstimated > 0
                  ? formatCurrency(reqEstimated, purchaseOrder.currency)
                  : "—";
              })()}
            </p>
          </div>

          <div className="space-y-1">
            <label className="text-xs font-semibold text-primary-foreground/80 uppercase tracking-wider flex items-center gap-1">
              <DollarSign className="h-3 w-3" />
              Total Amount
            </label>
            <p className="text-base font-bold text-primary-foreground">
              {formatCurrency(
                purchaseOrder.totalAmount,
                purchaseOrder.currency,
              )}
            </p>
          </div>

          <div className="space-y-1">
            <label className="text-xs font-semibold text-primary-foreground/80 uppercase tracking-wider flex items-center gap-1">
              <Tag className="h-3 w-3" />
              Budget Code
            </label>
            <p className="text-sm font-medium font-mono bg-white/10 px-2 py-1 rounded text-primary-foreground">
              {purchaseOrder.budgetCode || "—"}
            </p>
          </div>

          {purchaseOrder.costCenter && (
            <div className="space-y-1">
              <label className="text-xs font-semibold text-primary-foreground/80 uppercase tracking-wider flex items-center gap-1">
                <Building className="h-3 w-3" />
                Cost Center
              </label>
              <p className="text-sm font-medium font-mono bg-white/10 px-2 py-1 rounded text-primary-foreground">
                {purchaseOrder.costCenter}
              </p>
            </div>
          )}

          {purchaseOrder.projectCode && (
            <div className="space-y-1">
              <label className="text-xs font-semibold text-primary-foreground/80 uppercase tracking-wider flex items-center gap-1">
                <FileText className="h-3 w-3" />
                Project Code
              </label>
              <p className="text-sm font-medium font-mono bg-white/10 px-2 py-1 rounded text-primary-foreground">
                {purchaseOrder.projectCode}
              </p>
            </div>
          )}

          <div className="space-y-1">
            <label className="text-xs font-semibold text-primary-foreground/80 uppercase tracking-wider flex items-center gap-1">
              <Calendar className="h-3 w-3" />
              Created Date
            </label>
            <p className="text-sm font-medium text-primary-foreground">
              {new Date(purchaseOrder.createdAt).toLocaleDateString("en-ZM", {
                year: "numeric",
                month: "long",
                day: "numeric",
              })}
            </p>
          </div>

          {purchaseOrder.updatedAt &&
            new Date(purchaseOrder.updatedAt).getTime() !==
              new Date(purchaseOrder.createdAt).getTime() && (
              <div className="space-y-1">
                <label className="text-xs font-semibold text-primary-foreground/80 uppercase tracking-wider flex items-center gap-1">
                  <Clock className="h-3 w-3" />
                  Last Updated
                </label>
                <p className="text-sm font-medium text-primary-foreground">
                  {new Date(purchaseOrder.updatedAt).toLocaleDateString(
                    "en-ZM",
                    {
                      year: "numeric",
                      month: "long",
                      day: "numeric",
                    },
                  )}
                </p>
              </div>
            )}

          {purchaseOrder.deliveryDate && (
            <div className="space-y-1">
              <label className="text-xs font-semibold text-primary-foreground/80 uppercase tracking-wider flex items-center gap-1">
                <Calendar className="h-3 w-3" />
                Delivery Date
              </label>
              <p
                className={`text-sm font-medium ${
                  new Date(purchaseOrder.deliveryDate) < new Date() &&
                  purchaseOrder.status?.toUpperCase() !== "COMPLETED"
                    ? "text-red-200 font-bold"
                    : "text-primary-foreground"
                }`}
              >
                {new Date(purchaseOrder.deliveryDate).toLocaleDateString(
                  "en-ZM",
                  {
                    year: "numeric",
                    month: "long",
                    day: "numeric",
                  },
                )}
                {new Date(purchaseOrder.deliveryDate) < new Date() &&
                  purchaseOrder.status?.toUpperCase() !== "COMPLETED" && (
                    <span className="ml-2 text-xs">(Overdue)</span>
                  )}
              </p>
            </div>
          )}

          <div className="space-y-1">
            <label className="text-xs font-semibold text-primary-foreground/80 uppercase tracking-wider">
              Approval Stage
            </label>
            <p className="text-sm font-medium font-mono bg-white/10 px-2 py-1 rounded text-primary-foreground">
              {approvalData?.workflowStatus?.currentStage &&
              approvalData?.workflowStatus?.totalStages
                ? `${approvalData.workflowStatus.currentStage}/${approvalData.workflowStatus.totalStages}`
                : purchaseOrder.currentStage &&
                    approvalData?.workflowStatus?.totalStages
                  ? `${purchaseOrder.currentStage}/${approvalData.workflowStatus.totalStages}`
                  : `${purchaseOrder.approvalStage || 0}/1`}
            </p>
          </div>

          {purchaseOrder.linkedRequisition && (
            <div className="space-y-1">
              <label className="text-xs font-semibold text-primary-foreground/80 uppercase tracking-wider flex items-center gap-1">
                <FileText className="h-3 w-3" /> Source Requisition
              </label>
              <a
                href={`/requisitions/${purchaseOrder.sourceRequisitionId || purchaseOrder.linkedRequisition}`}
                className="text-base font-medium text-primary-foreground underline underline-offset-2 hover:opacity-80"
              >
                {purchaseOrder.linkedRequisition}
              </a>
            </div>
          )}
        </div>

        {/* Description */}
        {purchaseOrder.description && (
          <div className="mt-6 pt-6 border-t border-white/20">
            <label className="text-xs font-semibold text-primary-foreground/80 uppercase tracking-wider block mb-2">
              Description
            </label>
            <p className="text-sm text-primary-foreground leading-relaxed">
              {purchaseOrder.description}
            </p>
          </div>
        )}
      </div>

      {/* Vendor Details Card — shown when vendor is selected */}
      {purchaseOrder.vendorId && (
        <div className="rounded-lg border p-4 space-y-4 bg-muted/30">
          <h3 className="text-sm font-semibold flex items-center gap-2 flex-wrap">
            <Building className="h-4 w-4" />
            Supplier Details —{" "}
            {purchaseOrder.vendorName ||
              vendorDetails?.name ||
              "Unknown Vendor"}
            {complianceWarnings.length > 0 && (
              <Badge
                variant="outline"
                className="border-amber-300 bg-amber-50 text-amber-800 dark:border-amber-800 dark:bg-amber-950/30 dark:text-amber-300"
              >
                <AlertTriangle className="h-3 w-3 mr-1" />
                Compliance incomplete
              </Badge>
            )}
          </h3>

          {/* Vendor compliance warning — warn-only, never blocks the PO */}
          {complianceWarnings.length > 0 && (
            <VendorComplianceWarning warnings={complianceWarnings} />
          )}

          {/* Cost Comparison Section — always shown when vendor is selected */}
          {(() => {
            // REQ Estimated Cost: snapshot captured on the PO at creation time;
            // fall back to the live linked REQ total when no snapshot was stored.
            const reqEstimated =
              purchaseOrder.estimatedCost ||
              linkedRequisitionData?.totalAmount ||
              0;
            const hasReqEstimate = reqEstimated > 0;
            const poTotal = purchaseOrder.totalAmount;
            const selectedFileUrl = purchaseOrder.metadata
              ?.selectedQuotationFileUrl as string | undefined;
            const selectedQuotation = selectedFileUrl
              ? quotations.find((q) => q.fileUrl === selectedFileUrl)
              : undefined;
            const actual =
              selectedQuotation?.amount ?? purchaseOrder.totalAmount;

            // Primary variance (shown on the component): Supplier vs REQ Estimate.
            const baseForVariance = hasReqEstimate ? reqEstimated : poTotal;
            const diff = actual - baseForVariance;
            const pct =
              baseForVariance > 0 ? (diff / baseForVariance) * 100 : 0;
            const isOver = diff > 0;
            const isUnder = diff < 0;
            const color = isUnder
              ? "text-green-600 dark:text-green-400"
              : Math.abs(pct) <= 10
                ? "text-amber-600 dark:text-amber-400"
                : "text-red-600 dark:text-red-400";
            const Icon = isUnder ? TrendingDown : isOver ? TrendingUp : Minus;

            // Secondary variance (in popover): Supplier vs PO Total Amount.
            const diffPo = actual - poTotal;
            const pctPo = poTotal > 0 ? (diffPo / poTotal) * 100 : 0;
            const signed = (v: number) =>
              `${v > 0 ? "+" : v < 0 ? "−" : ""}${formatCurrency(
                Math.abs(v),
                purchaseOrder.currency,
              )}`;

            return (
              <div className="rounded-lg border border-blue-200 dark:border-blue-800 bg-blue-50 dark:bg-blue-950/30 p-3 space-y-3">
                <h4 className="text-xs font-semibold text-blue-900 dark:text-blue-100 uppercase tracking-wider">
                  Cost Comparison
                </h4>
                <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
                  <div>
                    <span className="text-xs text-blue-700 dark:text-blue-300 font-medium">
                      REQ Estimated Cost
                    </span>
                    <p className="text-base font-bold text-blue-900 dark:text-blue-100">
                      {hasReqEstimate
                        ? formatCurrency(reqEstimated, purchaseOrder.currency)
                        : "—"}
                    </p>
                  </div>
                  <div>
                    <span className="text-xs text-blue-700 dark:text-blue-300 font-medium">
                      PO Total Amount
                    </span>
                    <p className="text-base font-bold text-blue-900 dark:text-blue-100">
                      {formatCurrency(poTotal, purchaseOrder.currency)}
                    </p>
                  </div>
                  <div>
                    <span className="text-xs text-blue-700 dark:text-blue-300 font-medium">
                      Selected Supplier Price
                    </span>
                    <p className="text-base font-bold text-blue-900 dark:text-blue-100">
                      {formatCurrency(actual, purchaseOrder.currency)}
                    </p>
                  </div>
                  <div>
                    <span className="text-xs text-blue-700 dark:text-blue-300 font-medium flex items-center gap-1">
                      <Icon className="h-3 w-3" />
                      Price Variance
                      <Popover>
                        <PopoverTrigger asChild>
                          <button
                            type="button"
                            aria-label="Variance breakdown"
                            className="inline-flex text-blue-700/80 dark:text-blue-300/80 hover:text-blue-900 dark:hover:text-blue-100 transition-colors focus-visible:outline-none"
                          >
                            <Info className="h-3 w-3" />
                          </button>
                        </PopoverTrigger>
                        <PopoverContent align="end" className="w-72 space-y-3">
                          <p className="text-sm font-semibold">
                            Price Variance breakdown
                          </p>
                          <div className="space-y-2 text-xs">
                            <div className="flex items-center justify-between gap-2">
                              <span className="text-muted-foreground">
                                vs REQ Estimated Cost
                              </span>
                              <span className="font-semibold">
                                {hasReqEstimate ? (
                                  <>
                                    {signed(diff)} (
                                    {pct > 0 ? "+" : pct < 0 ? "−" : ""}
                                    {Math.abs(pct).toFixed(1)}%)
                                  </>
                                ) : (
                                  "—"
                                )}
                              </span>
                            </div>
                            <div className="flex items-center justify-between gap-2">
                              <span className="text-muted-foreground">
                                vs PO Total Amount
                              </span>
                              <span className="font-semibold">
                                {signed(diffPo)} (
                                {pctPo > 0 ? "+" : pctPo < 0 ? "−" : ""}
                                {Math.abs(pctPo).toFixed(1)}%)
                              </span>
                            </div>
                          </div>
                          <p className="text-[11px] text-muted-foreground">
                            The figure on the card compares the selected
                            supplier price against the{" "}
                            {hasReqEstimate
                              ? "requisition estimate"
                              : "PO total"}
                            .
                          </p>
                        </PopoverContent>
                      </Popover>
                    </span>
                    <p className={`text-base font-bold ${color}`}>
                      {isUnder ? "−" : isOver ? "+" : ""}
                      {formatCurrency(Math.abs(diff), purchaseOrder.currency)}
                      <span className="text-sm font-normal ml-1">
                        ({isUnder ? "−" : isOver ? "+" : ""}
                        {Math.abs(pct).toFixed(1)}%)
                      </span>
                    </p>
                  </div>
                </div>
              </div>
            );
          })()}

          {/* Vendor Contact Details - only show if vendor details are available */}
          {vendorDetails &&
            (vendorDetails.contactPerson ||
              vendorDetails.email ||
              vendorDetails.phone ||
              vendorDetails.physicalAddress ||
              vendorDetails.bankName ||
              vendorDetails.accountNumber ||
              vendorDetails.zraTpin ||
              vendorDetails.taxId ||
              vendorDetails.pacraRegNumber) && (
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-x-6 gap-y-2 text-sm">
                {vendorDetails.contactPerson && (
                  <div>
                    <span className="text-xs text-muted-foreground">
                      Contact Person
                    </span>
                    <p className="font-medium">{vendorDetails.contactPerson}</p>
                  </div>
                )}
                {vendorDetails.email && (
                  <div>
                    <span className="text-xs text-muted-foreground">Email</span>
                    <p className="font-medium">{vendorDetails.email}</p>
                  </div>
                )}
                {vendorDetails.phone && (
                  <div>
                    <span className="text-xs text-muted-foreground">Phone</span>
                    <p className="font-medium">{vendorDetails.phone}</p>
                  </div>
                )}
                {vendorDetails.physicalAddress && (
                  <div>
                    <span className="text-xs text-muted-foreground">
                      Address
                    </span>
                    <p className="font-medium">
                      {vendorDetails.physicalAddress}
                    </p>
                  </div>
                )}
                {vendorDetails.bankName && (
                  <div>
                    <span className="text-xs text-muted-foreground">Bank</span>
                    <p className="font-medium">{vendorDetails.bankName}</p>
                  </div>
                )}
                {vendorDetails.accountNumber && (
                  <div>
                    <span className="text-xs text-muted-foreground">
                      Account Number
                    </span>
                    <p className="font-medium font-mono">
                      {vendorDetails.accountNumber}
                    </p>
                  </div>
                )}
                {(vendorDetails.zraTpin || vendorDetails.taxId) && (
                  <div>
                    <span className="text-xs text-muted-foreground">
                      ZRA TPIN
                    </span>
                    <p className="font-medium font-mono">
                      {vendorDetails.zraTpin || vendorDetails.taxId}
                    </p>
                  </div>
                )}
                {vendorDetails.pacraRegNumber && (
                  <div>
                    <span className="text-xs text-muted-foreground">
                      PACRA Reg. No.
                    </span>
                    <p className="font-medium font-mono">
                      {vendorDetails.pacraRegNumber}
                    </p>
                  </div>
                )}
              </div>
            )}
        </div>
      )}

      {/* Payment Voucher summary — status chip, paid/committed/balance amounts,
          and the list of linked PVs when this PO has multiple (partial payments).
          The "Create PV" action is gated on remaining balance, not "no PV exists
          yet", since a PO can now carry several PVs up to its total. */}
      {(purchaseOrder.status?.toUpperCase() === "APPROVED" ||
        purchaseOrder.status?.toUpperCase() === "FULFILLED" ||
        !!purchaseOrder.linkedPV ||
        (purchaseOrder.linkedPVs?.length ?? 0) > 0) && (
        <Card className="p-4 border-0 shadow-sm space-y-3">
          <div className="flex items-center justify-between flex-wrap gap-2">
            <div className="flex items-center gap-2">
              <Wallet className="h-4 w-4 text-muted-foreground" />
              <h3 className="text-sm font-semibold">Payment Voucher</h3>
              {(() => {
                const paymentStatus = purchaseOrder.paymentStatus ?? "unpaid";
                const styles: Record<string, string> = {
                  fully_paid:
                    "bg-green-100 text-green-800 border-green-200 dark:bg-green-950/40 dark:text-green-300 dark:border-green-800",
                  partially_paid:
                    "bg-amber-100 text-amber-800 border-amber-200 dark:bg-amber-950/40 dark:text-amber-300 dark:border-amber-800",
                  unpaid:
                    "bg-gray-100 text-gray-800 border-gray-200 dark:bg-gray-800/40 dark:text-gray-300 dark:border-gray-700",
                };
                const labels: Record<string, string> = {
                  fully_paid: "Fully Paid",
                  partially_paid: "Partially Paid",
                  unpaid: "Unpaid",
                };
                return (
                  <span
                    className={`text-xs px-2 py-1 rounded-full border font-medium ${styles[paymentStatus]}`}
                  >
                    {labels[paymentStatus]}
                  </span>
                );
              })()}
            </div>
            {purchaseOrder.status?.toUpperCase() === "APPROVED" &&
              canCreateAnotherPV(purchaseOrder) && (
                <Button
                  variant="default"
                  size="sm"
                  onClick={() => setIsCreatePVDialogOpen(true)}
                >
                  <FileText className="h-4 w-4 mr-1" />
                  Create PV
                </Button>
              )}
          </div>

          {(purchaseOrder.amountPaid !== undefined ||
            purchaseOrder.amountCommitted !== undefined ||
            purchaseOrder.balance !== undefined) && (
            <div className="grid grid-cols-3 gap-3 text-center rounded-lg border bg-muted/30 p-3">
              <div>
                <p className="text-xs text-muted-foreground">Paid</p>
                <p className="text-sm font-semibold font-mono">
                  {formatCurrency(
                    purchaseOrder.amountPaid ?? 0,
                    purchaseOrder.currency,
                  )}
                </p>
              </div>
              <div>
                <p className="text-xs text-muted-foreground">Committed</p>
                <p className="text-sm font-semibold font-mono">
                  {formatCurrency(
                    purchaseOrder.amountCommitted ?? 0,
                    purchaseOrder.currency,
                  )}
                </p>
              </div>
              <div>
                <p className="text-xs text-muted-foreground">Balance</p>
                <p className="text-sm font-semibold font-mono">
                  {formatCurrency(
                    poRemainingBalance(purchaseOrder),
                    purchaseOrder.currency,
                  )}
                </p>
              </div>
            </div>
          )}

          {purchaseOrder.linkedPVs && purchaseOrder.linkedPVs.length > 0 ? (
            <div className="space-y-1.5">
              {purchaseOrder.linkedPVs.map((pv) => (
                <Link
                  key={pv.id}
                  href={`/payment-vouchers/${pv.id}`}
                  className="flex items-center justify-between gap-2 rounded-md border p-2 text-sm hover:bg-muted/50 transition-colors"
                >
                  <span className="font-mono truncate">
                    {pv.documentNumber}
                  </span>
                  <span className="text-xs text-muted-foreground shrink-0">
                    {pv.status}
                  </span>
                  <span className="font-mono text-xs shrink-0">
                    {pv.amount !== undefined
                      ? formatCurrency(pv.amount, purchaseOrder.currency)
                      : "—"}
                  </span>
                </Link>
              ))}
            </div>
          ) : (
            !purchaseOrder.linkedPV && (
              <p className="text-xs text-muted-foreground">
                No payment voucher yet
              </p>
            )
          )}
        </Card>
      )}

      {/* ── Tabbed Content ──────────────────────────────────────────── */}
      <Card className="p-6 border-0 shadow-sm">
        <Tabs defaultValue="items" className="w-full">
          <div className="overflow-x-auto no-scrollbar -mx-6 px-6">
            <TabsList className="flex min-w-full h-auto">
              <TabsTrigger
                value="items"
                className="gap-1.5 text-xs sm:text-sm px-2 py-2 flex-1 shrink-0 whitespace-nowrap"
              >
                <ShoppingCart className="h-4 w-4 shrink-0" />
                <span className="hidden sm:inline">PO</span> Items
                {purchaseOrder.items?.length > 0 && (
                  <Badge
                    variant="secondary"
                    className="ml-1 text-xs h-5 min-w-5 px-1.5"
                  >
                    {purchaseOrder.items.length}
                  </Badge>
                )}
              </TabsTrigger>
              <TabsTrigger
                value="documents"
                className="gap-1.5 text-xs sm:text-sm px-2 py-2 flex-1 shrink-0 whitespace-nowrap"
              >
                <Paperclip className="h-4 w-4 shrink-0" />
                <span className="hidden sm:inline">Supporting</span> Docs
                {attachments.length > 0 && (
                  <Badge
                    variant="secondary"
                    className="ml-1 text-xs h-5 min-w-5 px-1.5"
                  >
                    {attachments.length}
                  </Badge>
                )}
              </TabsTrigger>
              <TabsTrigger
                value="action"
                className="gap-1.5 text-xs sm:text-sm px-2 py-2 flex-1 shrink-0 whitespace-nowrap"
              >
                <CheckSquare className="h-4 w-4 shrink-0" />
                <span className="hidden sm:inline">Approval</span> Action
              </TabsTrigger>
              <TabsTrigger
                value="chain"
                className="gap-1.5 text-xs sm:text-sm px-2 py-2 flex-1 shrink-0 whitespace-nowrap"
              >
                <GitBranch className="h-4 w-4 shrink-0" />
                <span className="hidden sm:inline">Approval</span> Chain
              </TabsTrigger>
              <TabsTrigger
                value="shipping"
                className="gap-1.5 text-xs sm:text-sm px-2 py-2 flex-1 shrink-0 whitespace-nowrap"
              >
                <Truck className="h-4 w-4 shrink-0" />
                <span className="hidden sm:inline">Shipping</span> &amp; Tax
              </TabsTrigger>
              <TabsTrigger
                value="activity"
                className="gap-1.5 text-xs sm:text-sm px-2 py-2 flex-1 shrink-0 whitespace-nowrap"
              >
                <Activity className="h-4 w-4 shrink-0" />
                <span className="hidden sm:inline">Activity</span> Log
                {purchaseOrder.actionHistory &&
                  purchaseOrder.actionHistory.length > 0 && (
                    <Badge
                      variant="secondary"
                      className="ml-1 text-xs h-5 min-w-5 px-1.5"
                    >
                      {purchaseOrder.actionHistory.length}
                    </Badge>
                  )}
              </TabsTrigger>
            </TabsList>
          </div>

          {/* ── Tab 1: PO Items ── */}
          <TabsContent value="items" className="mt-6">
            <div className="flex items-center justify-between mb-6">
              <h2 className="text-lg font-semibold">
                Items ({purchaseOrder.items?.length || 0})
              </h2>
              {canEditItems && !editingItems && (
                <Button
                  variant="outline"
                  size="sm"
                  aria-label="Edit line items"
                  onClick={() => setEditingItems(true)}
                >
                  <Pencil className="h-3.5 w-3.5 mr-1" />
                  {purchaseOrder.items?.length ? "Edit Items" : "Add Items"}
                </Button>
              )}
            </div>

            {editingItems && isDraft ? (
              <POItemsEditor
                poId={purchaseOrderId}
                items={(purchaseOrder.items ?? []).map((item, index) => ({
                  id: item.id || `item-${index}`,
                  description: item.description || "",
                  quantity: item.quantity || 0,
                  unitPrice: item.unitPrice || 0,
                  amount: item.totalPrice || item.amount || 0,
                  totalPrice: item.totalPrice || item.amount || 0,
                  unit: item.unit,
                  category: item.category,
                  notes: item.notes,
                }))}
                currency={purchaseOrder.currency || "ZMW"}
                metadata={
                  purchaseOrder.metadata as Record<string, unknown> | undefined
                }
                reqItems={linkedRequisitionItems}
                onSaved={() => setEditingItems(false)}
                onCancel={() => setEditingItems(false)}
              />
            ) : purchaseOrder.items && purchaseOrder.items.length > 0 ? (
              <PurchaseOrderItemsList
                items={purchaseOrder.items}
                currency={purchaseOrder.currency}
              />
            ) : (
              <Empty>
                <EmptyMedia variant="icon">
                  <Package className="h-6 w-6" />
                </EmptyMedia>
                <EmptyContent>
                  <EmptyDescription>No items added yet</EmptyDescription>
                </EmptyContent>
              </Empty>
            )}
          </TabsContent>

          {/* ── Tab 2: Supporting Documents ── */}
          <TabsContent value="documents" className="mt-6 space-y-6">
            <SupportingDocuments
              documentId={purchaseOrderId}
              documentType="purchase-order"
              chainDocs={poChainDocs}
              canUpload={isDraft}
              onUpload={handleSupportingDocUpload}
              canDeleteFile={canDeleteAttachment}
              onDeleteFile={handleDeleteAttachment}
              showViewLinks={userRole.toLowerCase() !== "requester"}
            />

            {/* Quotations section */}
            {!purchaseOrder.automationUsed && (
              <QuotationCollectionSection
                quotations={quotations}
                requisitionId={purchaseOrderId}
                currency={purchaseOrder.currency || "ZMW"}
                vendors={vendors}
                canEdit={canEditQuotations}
                onSave={handleSaveQuotations}
                selectedVendorId={purchaseOrder.vendorId}
                selectedQuotationFileId={
                  purchaseOrder.metadata?.selectedQuotationFileUrl as
                    | string
                    | undefined
                }
                onSelectVendor={handleSelectVendor}
                showVendorSelection={isDraft}
              />
            )}

          </TabsContent>

          {/* ── Tab 3: Approval Action ── */}
          <TabsContent value="action" className="space-y-4 mt-6">
            {approvalData?.hasError ? (
              <div className="text-center py-8 text-red-500">
                <AlertCircle className="h-8 w-8 mx-auto mb-2" />
                <p className="text-sm">Failed to load approval data</p>
                <button
                  onClick={approvalData.refetchAll}
                  className="mt-2 text-xs text-blue-600 hover:underline"
                >
                  Try again
                </button>
              </div>
            ) : (
              <>
                {purchaseOrder.quotationGateOverridden && (
                  <Alert className="border-amber-200 bg-amber-50 mb-4">
                    <AlertTriangle className="h-4 w-4 text-amber-600" />
                    <AlertDescription>
                      <p className="text-amber-800 font-medium text-sm">
                        Quotation Override Applied
                      </p>
                      {purchaseOrder.bypassJustification && (
                        <p className="text-amber-700 text-xs mt-1">
                          {purchaseOrder.bypassJustification}
                        </p>
                      )}
                    </AlertDescription>
                  </Alert>
                )}
                <ApprovalActionContent
                  requisitionId={purchaseOrderId}
                  requisition={purchaseOrder as any}
                  workflowStatus={approvalData?.workflowStatus}
                  isLoading={approvalData?.isLoading || false}
                  onApprovalComplete={handleApprovalComplete}
                />
              </>
            )}
          </TabsContent>

          {/* ── Tab 4: Approval Chain ── */}
          <TabsContent value="chain" className="space-y-4 mt-6">
            {approvalData?.hasError ? (
              <div className="text-center py-8 text-red-500">
                <AlertCircle className="h-8 w-8 mx-auto mb-2" />
                <p className="text-sm">Failed to load approval data</p>
                <button
                  onClick={approvalData.refetchAll}
                  className="mt-2 text-xs text-blue-600 hover:underline"
                >
                  Try again
                </button>
              </div>
            ) : (
              <>
                <ApprovalChainContent
                  requisition={purchaseOrder as any}
                  approvalChain={purchaseOrder?.approvalChain}
                  approvalHistory={approvalData?.approvalHistory || []}
                  workflowStatus={approvalData?.workflowStatus}
                  availableApprovers={approvalData?.availableApprovers || []}
                  isLoading={approvalData?.isLoading || false}
                />
                <WorkflowStatusSummary
                  requisition={purchaseOrder as any}
                  workflowStatus={approvalData?.workflowStatus}
                />
              </>
            )}
          </TabsContent>

          {/* ── Tab 5: Activity Log (Timeline) ── */}
          <TabsContent value="activity" className="space-y-4 mt-6">
            <ActivityLogContent
              actionHistory={purchaseOrder?.actionHistory}
              auditEvents={auditEventsData}
            />
          </TabsContent>

          {/* ── Tab 6: Shipping & Tax ── */}
          <TabsContent value="shipping" className="mt-6">
            <POShippingEditor
              poId={purchaseOrderId}
              purchaseOrder={purchaseOrder}
              canEdit={
                permissions.canEdit ||
                ["admin", "finance"].includes(userRole?.toLowerCase())
              }
              onSaved={handleDocumentUpdated}
            />
          </TabsContent>
        </Tabs>
      </Card>

      {/* PDF Preview Dialog */}
      {previewBlob && (
        <PDFPreviewDialog
          open={previewOpen}
          onOpenChange={setPreviewOpen}
          pdfBlob={previewBlob}
          fileName={`Purchase Order: ${purchaseOrder.documentNumber}`}
          onDownload={handleExportPDF}
        />
      )}

      {/* Submit Dialog */}
      <PurchaseOrderSubmitDialog
        open={showSubmitDialog}
        onOpenChange={setShowSubmitDialog}
        purchaseOrder={purchaseOrder}
        onSubmit={handleSubmitForApproval}
        isSubmitting={submitMutation.isPending}
      />

      {/* Withdraw Confirmation Modal */}
      <ConfirmationModal
        open={showWithdrawModal}
        onOpenChange={setShowWithdrawModal}
        onConfirm={handleWithdraw}
        type="withdraw"
        title="Withdraw Purchase Order"
        description={`Are you sure you want to withdraw purchase order ${purchaseOrder.documentNumber || purchaseOrder.id}? It will be reverted to draft status and you can edit and re-submit it later.`}
        isLoading={withdrawMutation?.isPending || false}
      />


      {/* Edit Purchase Order Dialog */}
      <EditPurchaseOrderDialog
        open={isEditDialogOpen}
        onOpenChange={setIsEditDialogOpen}
        purchaseOrder={purchaseOrder}
        onSuccess={handleDocumentUpdated}
      />

      {/* Create Payment Voucher Dialog */}
      <CreatePVFromPODialog
        open={isCreatePVDialogOpen}
        onOpenChange={setIsCreatePVDialogOpen}
        purchaseOrder={purchaseOrder}
        onConfirm={handleConfirmCreatePV}
        isCreating={isCreatingPV}
      />
    </div>
  );
}
