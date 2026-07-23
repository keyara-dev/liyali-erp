"use client";

import { useState } from "react";
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
  User,
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
} from "lucide-react";
import { PageHeader } from "@/components/base/page-header";
import { RequisitionItemsList } from "./requisition-items-list";
import { Requisition, RequisitionAttachment } from "@/types/requisition";
import {
  ActivityLogContent,
  ApprovalChainContent,
  ApprovalActionContent,
  WorkflowStatusSummary,
} from "./approval-history-panel";
import { CreateRequisitionDialog } from "./create-requisition-dialog";
import { POCreationWizard } from "@/app/(private)/(main)/purchase-orders/_components/po-creation-wizard";
import { toast } from "sonner";
import { buildChainLinks } from "@/components/linked-documents";
import {
  SupportingDocuments,
  type UploadedAttachment,
} from "@/components/supporting-documents";
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

import { RequisitionSubmitDialog } from "./requisition-submit-dialog";
import { ConfirmationModal } from "@/components/modals/confirmation-modal";
import { Badge } from "@/components";
import { DocumentLoadingPage } from "@/components/base/document-loading-page";
import ErrorDisplay from "@/components/base/error-display";
import { useRequisitionDetail } from "@/hooks/use-requisition-detail";
import { usePermissions } from "@/hooks/use-permissions";
import { useRouter } from "next/navigation";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { getAuditEvents, type AuditEvent } from "@/app/_actions/audit";
import { QuotationCollectionSection } from "./quotation-collection-section";
import { updateRequisition } from "@/app/_actions/requisitions";
import { useVendors } from "@/hooks/use-vendor-queries";
import { type Quotation } from "@/types/core";
import { QUERY_KEYS } from "@/lib/constants";

interface RequisitionDetailClientProps {
  requisitionId: string;
  userId: string;
  userRole: string;
  initialRequisition?: Requisition;
}

export function RequisitionDetailClient({
  requisitionId,
  userId,
  userRole,
  initialRequisition,
}: RequisitionDetailClientProps) {
  const router = useRouter();
  const queryClient = useQueryClient();
  const [editInitialStep, setEditInitialStep] = useState<"details" | "items">(
    "details",
  );
  const [isCreatePOOpen, setIsCreatePOOpen] = useState(false);

  const { data: vendors = [] } = useVendors({ active: true });

  const { data: auditEventsData } = useQuery({
    queryKey: ["audit-events", "requisition", requisitionId],
    queryFn: async () => {
      const res = await getAuditEvents("requisition", requisitionId);
      return res.success ? ((res.data as AuditEvent[]) ?? []) : [];
    },
    enabled: !!requisitionId,
  });

  // Use the new hook to manage all document detail logic
  const {
    document: requisition,
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
  } = useRequisitionDetail({
    requisitionId,
    userId,
    userRole,
    initialRequisition,
  });

  const { hasPermission, isAdmin, isFinance } = usePermissions();

  const canGeneratePO =
    requisition?.status?.toUpperCase() === "APPROVED" &&
    (hasPermission("purchase_order", "create") || isAdmin() || isFinance());

  if (isLoading) return <DocumentLoadingPage />;

  if (!requisition)
    return (
      <ErrorDisplay
        title="Requisition Not Found"
        message="The requisition you're looking for doesn't exist."
        showBackButton
      />
    );

  const attachments: RequisitionAttachment[] =
    requisition.attachments ||
    (requisition.metadata?.attachments as RequisitionAttachment[]) ||
    [];

  const quotations: Quotation[] =
    (requisition.metadata?.quotations as Quotation[]) || [];

  const canEditQuotations =
    ["admin", "finance", "approver"].includes(userRole.toLowerCase()) ||
    requisition.requesterId === userId;

  // Chain-doc rows only exist once the requisition has a PO/GRN/PV to link —
  // mirrors the old gating on the standalone LinkedDocuments mount.
  const reqChainDocs =
    requisition.status?.toUpperCase() === "APPROVED"
      ? buildChainLinks(chain, "requisition")
      : [];

  const handleSaveQuotations = async (newQuotations: Quotation[]) => {
    await updateRequisition({
      requisitionId: requisition.id,
      quotations: newQuotations,
    });
    queryClient.invalidateQueries({
      queryKey: [QUERY_KEYS.REQUISITIONS.BY_ID, requisition.id],
    });
  };

  // Persists a SupportingDocuments upload (already on ImageKit) into the
  // requisition's own attachments — read-modify-write via updateRequisition.
  const handleSupportingDocUpload = async (att: UploadedAttachment) => {
    await updateRequisition({
      requisitionId: requisition.id,
      attachments: [...attachments, att],
    });
    handleDocumentUpdated();
  };

  // Custom submit handler to pass additional requisition data
  const handleSubmitForApproval = async (
    workflowId: string,
    comments?: string,
  ) => {
    await handleSubmit(workflowId, comments, {
      submittedBy: userId,
      submittedByName: requisition.requestedByName || "User",
      submittedByRole: requisition.requestedByRole || userRole,
    });
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-start justify-between gap-4">
        <PageHeader
          title={requisition.documentNumber}
          subtitle={`${requisition.title || "Untitled Requisition"} • Created ${new Date(requisition.createdAt).toLocaleDateString("en-ZM", { year: "numeric", month: "long", day: "numeric" })}${requisition.updatedAt && new Date(requisition.updatedAt).getTime() !== new Date(requisition.createdAt).getTime() ? ` • Updated ${new Date(requisition.updatedAt).toLocaleDateString("en-ZM", { year: "numeric", month: "long", day: "numeric" })}` : ""}`}
          badges={[
            {
              status: requisition.status,
              type: "document",
            },
          ]}
          onBackClick={() => router.back()}
          showBackButton={true}
        />
        <div className="flex gap-2 mt-2">
          <Button
            onClick={handlePreviewPDF}
            disabled={isExporting}
            variant="outline"
            className="gap-2 h-11"
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
            className="gap-2 h-11"
          >
            <Download className="h-4 w-4" />
            Export PDF
          </Button>
          {canGeneratePO && (
            <Button
              onClick={() => setIsCreatePOOpen(true)}
              variant="default"
              className="gap-2 h-11"
            >
              <ShoppingCart className="h-4 w-4" />
              Generate PO
            </Button>
          )}
          {permissions.canEdit && (
            <Button
              onClick={() => {
                setEditInitialStep("details");
                handleEdit();
              }}
              variant="outline"
              className="gap-2 h-11"
            >
              <Pencil className="h-4 w-4" />
              Edit Requisition
            </Button>
          )}
          {permissions.canSubmit && (
            <Button
              onClick={() => setShowSubmitDialog(true)}
              className="gap-2 h-11"
            >
              <Send className="h-4 w-4" />
              Submit for Approval
            </Button>
          )}
          {permissions.canWithdraw && (
            <Button
              onClick={() => setShowWithdrawModal(true)}
              variant="outline"
              className="gap-2 h-11 text-amber-600 border-amber-300 hover:bg-amber-50"
            >
              <Undo2 className="h-4 w-4" />
              Withdraw
            </Button>
          )}
        </div>
      </div>

      {/* Requisition Details Card */}
      <div className="gradient-primary border-0 overflow-hidden rounded-lg p-6">
        <h2 className="text-lg font-semibold mb-6 text-primary-foreground">
          Requisition Details
        </h2>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          <div className="space-y-1">
            <label className="text-xs font-semibold text-primary-foreground/80 uppercase tracking-wider flex items-center gap-1">
              <FileText className="h-3 w-3" />
              Title
            </label>
            <p className="text-base font-medium text-primary-foreground">
              {requisition.title || "—"}
            </p>
          </div>

          <div className="space-y-1">
            <label className="text-xs font-semibold text-primary-foreground/80 uppercase tracking-wider flex items-center gap-1">
              <Building className="h-3 w-3" />
              Department
            </label>
            <p className="text-base font-medium text-primary-foreground">
              {requisition.department || "—"}
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
                  requisition.priority?.toUpperCase() === "URGENT"
                    ? "bg-red-100 text-red-800 border-red-200"
                    : requisition.priority?.toUpperCase() === "HIGH"
                      ? "bg-orange-100 text-orange-800 border-orange-200"
                      : requisition.priority?.toUpperCase() === "MEDIUM"
                        ? "bg-blue-100 text-blue-800 border-blue-200"
                        : "bg-gray-100 text-gray-800 border-gray-200"
                }`}
              >
                {requisition.priority || "Medium"}
              </Badge>
            </div>
          </div>

          <div className="space-y-1">
            <label className="text-xs font-semibold text-primary-foreground/80 uppercase tracking-wider flex items-center gap-1">
              <User className="h-3 w-3" />
              Requested By
            </label>
            <p className="text-base font-medium text-primary-foreground">
              {requisition.requesterName || requisition.requestedByName || "—"}
            </p>
            {requisition.requestedByRole && (
              <p className="text-xs text-primary-foreground/60">
                {requisition.requestedByRole}
              </p>
            )}
          </div>

          {requisition.requestedFor && (
            <div className="space-y-1">
              <label className="text-xs font-semibold text-primary-foreground/80 uppercase tracking-wider flex items-center gap-1">
                <User className="h-3 w-3" />
                Requested For
              </label>
              <p className="text-base font-medium text-primary-foreground">
                {requisition.requestedFor}
              </p>
            </div>
          )}

          <div className="space-y-1">
            <label className="text-xs font-semibold text-primary-foreground/80 uppercase tracking-wider flex items-center gap-1">
              <DollarSign className="h-3 w-3" />
              Estimated Cost
            </label>
            <p className="text-base font-bold text-primary-foreground">
              {requisition.currency}{" "}
              {requisition.totalAmount?.toLocaleString("en-ZM", {
                minimumFractionDigits: 2,
                maximumFractionDigits: 2,
              }) || "0.00"}
            </p>
          </div>

          <div className="space-y-1">
            <label className="text-xs font-semibold text-primary-foreground/80 uppercase tracking-wider flex items-center gap-1">
              <Tag className="h-3 w-3" />
              Budget Code
            </label>
            <p className="text-sm font-medium font-mono bg-white/10 px-2 py-1 rounded text-primary-foreground">
              {requisition.budgetCode || "—"}
            </p>
          </div>

          {requisition.costCenter && (
            <div className="space-y-1">
              <label className="text-xs font-semibold text-primary-foreground/80 uppercase tracking-wider flex items-center gap-1">
                <Building className="h-3 w-3" />
                Cost Center
              </label>
              <p className="text-sm font-medium font-mono bg-white/10 px-2 py-1 rounded text-primary-foreground">
                {requisition.costCenter}
              </p>
            </div>
          )}

          {requisition.projectCode && (
            <div className="space-y-1">
              <label className="text-xs font-semibold text-primary-foreground/80 uppercase tracking-wider flex items-center gap-1">
                <FileText className="h-3 w-3" />
                Project Code
              </label>
              <p className="text-sm font-medium font-mono bg-white/10 px-2 py-1 rounded text-primary-foreground">
                {requisition.projectCode}
              </p>
            </div>
          )}

          <div className="space-y-1">
            <label className="text-xs font-semibold text-primary-foreground/80 uppercase tracking-wider flex items-center gap-1">
              <Calendar className="h-3 w-3" />
              Created Date
            </label>
            <p className="text-sm font-medium text-primary-foreground">
              {new Date(requisition.createdAt).toLocaleDateString("en-ZM", {
                year: "numeric",
                month: "long",
                day: "numeric",
              })}
            </p>
          </div>

          {requisition.updatedAt &&
            new Date(requisition.updatedAt).getTime() !==
              new Date(requisition.createdAt).getTime() && (
              <div className="space-y-1">
                <label className="text-xs font-semibold text-primary-foreground/80 uppercase tracking-wider flex items-center gap-1">
                  <Clock className="h-3 w-3" />
                  Last Updated
                </label>
                <p className="text-sm font-medium text-primary-foreground">
                  {new Date(requisition.updatedAt).toLocaleDateString("en-ZM", {
                    year: "numeric",
                    month: "long",
                    day: "numeric",
                  })}
                </p>
              </div>
            )}

          {requisition.requiredByDate && (
            <div className="space-y-1">
              <label className="text-xs font-semibold text-primary-foreground/80 uppercase tracking-wider flex items-center gap-1">
                <Calendar className="h-3 w-3" />
                Due Date
              </label>
              <p
                className={`text-sm font-medium ${
                  new Date(requisition.requiredByDate) < new Date() &&
                  requisition.status?.toUpperCase() !== "COMPLETED"
                    ? "text-red-200 font-bold"
                    : "text-primary-foreground"
                }`}
              >
                {new Date(requisition.requiredByDate).toLocaleDateString(
                  "en-ZM",
                  {
                    year: "numeric",
                    month: "long",
                    day: "numeric",
                  },
                )}
                {new Date(requisition.requiredByDate) < new Date() &&
                  requisition.status?.toUpperCase() !== "COMPLETED" && (
                    <span className="ml-2 text-xs">(Overdue)</span>
                  )}
              </p>
            </div>
          )}

          {(requisition.categoryName || requisition.otherCategoryText) && (
            <div className="space-y-1">
              <label className="text-xs font-semibold text-primary-foreground/80 uppercase tracking-wider">
                Category
              </label>
              <p className="text-sm font-medium text-primary-foreground">
                {requisition.categoryName ||
                  requisition.otherCategoryText ||
                  "—"}
                {requisition.otherCategoryText && (
                  <span className="text-xs text-primary-foreground/60 ml-1">
                    (Custom)
                  </span>
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
                : requisition.currentApprovalStage &&
                    requisition.totalApprovalStages
                  ? `${requisition.currentApprovalStage}/${requisition.totalApprovalStages}`
                  : `${requisition.approvalStage || 0}/1`}
            </p>
          </div>

          {requisition.isEstimate && (
            <div className="space-y-1">
              <p className="text-xs font-semibold text-primary-foreground/80 uppercase tracking-wider">
                Estimate
              </p>
              <div className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-yellow-100 text-yellow-800 border border-yellow-200">
                Estimated Costs
              </div>
            </div>
          )}
        </div>

        {/* Description */}
        {requisition.description && (
          <div className="mt-6 pt-6 border-t border-white/20">
            <label className="text-xs font-semibold text-primary-foreground/80 uppercase tracking-wider block mb-2">
              Description / Justification
            </label>
            <p className="text-sm text-primary-foreground leading-relaxed">
              {requisition.description}
            </p>
          </div>
        )}

        {/* Additional Metadata — exclude keys already shown as dedicated fields */}
        {requisition.metadata &&
          Object.entries(requisition.metadata).filter(
            ([key]) =>
              ![
                "attachments",
                "categoryName",
                "otherCategoryText",
                "requestedFor",
                "requestedByName",
                "requestedByRole",
                "title",
                "department",
                "departmentId",
                "priority",
                "budgetCode",
                "costCenter",
                "projectCode",
                "description",
                "isEstimate",
                "requiredByDate",
              ].includes(key),
          ).length > 0 && (
            <div className="mt-6 pt-6 border-t border-white/20">
              <label className="text-xs font-semibold text-primary-foreground/80 uppercase tracking-wider block mb-3">
                Additional Information
              </label>
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                {Object.entries(requisition.metadata)
                  .filter(
                    ([key]) =>
                      ![
                        "attachments",
                        "categoryName",
                        "otherCategoryText",
                        "requestedFor",
                        "requestedByName",
                        "requestedByRole",
                        "title",
                        "department",
                        "departmentId",
                        "priority",
                        "budgetCode",
                        "costCenter",
                        "projectCode",
                        "description",
                        "isEstimate",
                        "requiredByDate",
                      ].includes(key),
                  )
                  .map(([key, value]) => (
                    <div key={key} className="space-y-1">
                      <label className="text-xs font-medium text-primary-foreground/70 capitalize">
                        {key
                          .replace(/([A-Z])/g, " $1")
                          .replace(/^./, (str) => str.toUpperCase())}
                      </label>
                      <p className="text-sm text-primary-foreground">
                        {typeof value === "object"
                          ? JSON.stringify(value, null, 2)
                          : String(value)}
                      </p>
                    </div>
                  ))}
              </div>
            </div>
          )}

        {/* Auto-Created Purchase Order */}
        {requisition?.automationUsed && requisition?.autoCreatedPO && (
          <div className="mt-6 pt-6 border-t border-white/20">
            <label className="text-xs font-semibold text-primary-foreground/80 uppercase tracking-wider  mb-3 flex items-center gap-2">
              <span className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-green-100 text-green-800 border border-green-200">
                Automated
              </span>
              Auto-Generated Purchase Order
            </label>
            <div className="bg-white/10 rounded-lg p-4 border border-white/20">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div className="space-y-2">
                  <p className="text-sm font-medium text-primary-foreground">
                    PO Number:
                    <span className="ml-2 font-mono bg-white/20 px-2 py-1 rounded text-xs">
                      {(requisition.autoCreatedPO as any)?.documentNumber ||
                        "Generated"}
                    </span>
                  </p>
                  <p className="text-xs text-primary-foreground/80">
                    This purchase order was automatically created when the
                    requisition was approved.
                  </p>
                </div>
                <div className="flex items-center justify-end">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => {
                      const poId = (requisition.autoCreatedPO as any)?.id;
                      if (poId) {
                        router.push(`/purchase-orders/${poId}`);
                      }
                    }}
                    className="bg-white/10 border-white/30 text-primary-foreground hover:bg-white/20"
                  >
                    <Eye className="h-4 w-4 mr-2" />
                    View Purchase Order
                  </Button>
                </div>
              </div>
            </div>
          </div>
        )}
      </div>

      {/* ── Tabbed Content ──────────────────────────────────────────── */}
      <Card className="p-6 border-0 shadow-sm">
        <Tabs defaultValue="items" className="w-full">
          <TabsList className="grid w-full grid-cols-5 h-auto">
            <TabsTrigger
              value="items"
              className="gap-1.5 text-xs sm:text-sm px-2 py-2"
            >
              <ShoppingCart className="h-4 w-4 shrink-0" />
              <span className="hidden sm:inline">Requisition</span> Items
              {requisition.items?.length > 0 && (
                <Badge
                  variant="secondary"
                  className="ml-1 text-xs h-5 min-w-5 px-1.5"
                >
                  {requisition.items.length}
                </Badge>
              )}
            </TabsTrigger>
            <TabsTrigger
              value="documents"
              className="gap-1.5 text-xs sm:text-sm px-2 py-2"
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
              className="gap-1.5 text-xs sm:text-sm px-2 py-2"
            >
              <CheckSquare className="h-4 w-4 shrink-0" />
              <span className="hidden sm:inline">Approval</span> Action
            </TabsTrigger>
            <TabsTrigger
              value="chain"
              className="gap-1.5 text-xs sm:text-sm px-2 py-2"
            >
              <GitBranch className="h-4 w-4 shrink-0" />
              <span className="hidden sm:inline">Approval</span> Chain
            </TabsTrigger>
            <TabsTrigger
              value="activity"
              className="gap-1.5 text-xs sm:text-sm px-2 py-2"
            >
              <Activity className="h-4 w-4 shrink-0" />
              <span className="hidden sm:inline">Activity</span> Log
              {requisition.actionHistory &&
                requisition.actionHistory.length > 0 && (
                  <Badge
                    variant="secondary"
                    className="ml-1 text-xs h-5 min-w-5 px-1.5"
                  >
                    {requisition.actionHistory.length}
                  </Badge>
                )}
            </TabsTrigger>
          </TabsList>

          {/* ── Tab 1: Requisition Items ── */}
          <TabsContent value="items" className="mt-6">
            <div className="flex items-center justify-between mb-6">
              <h2 className="text-lg font-semibold">
                Items ({requisition.items?.length || 0})
              </h2>
              <div className="flex items-center gap-2">
                {requisition.isEstimate && (
                  <span className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-yellow-100 text-yellow-800 border border-yellow-200">
                    Estimated Costs
                  </span>
                )}
                {permissions.canEdit && (
                  <Button
                    onClick={() => {
                      setEditInitialStep("items");
                      handleEdit();
                    }}
                    variant="outline"
                    size="sm"
                    className="gap-2"
                  >
                    <Pencil className="h-3.5 w-3.5" />
                    {requisition.items?.length ? "Edit Items" : "Add Items"}
                  </Button>
                )}
              </div>
            </div>

            {requisition.items && requisition.items.length > 0 ? (
              <RequisitionItemsList
                items={requisition.items}
                currency={requisition.currency}
                isEstimate={requisition.isEstimate}
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
              documentId={requisition.id}
              documentType="requisition"
              chainDocs={reqChainDocs}
              canUpload={permissions.canEdit}
              onUpload={handleSupportingDocUpload}
              showViewLinks={userRole.toLowerCase() !== "requester"}
            />

            {/* ── Quotations section — shown only on APPROVED reqs ── */}
            {requisition.status?.toUpperCase() === "APPROVED" && (
              <QuotationCollectionSection
                quotations={quotations}
                requisitionId={requisition.id}
                currency={requisition.currency || "ZMW"}
                vendors={vendors}
                canEdit={canEditQuotations}
                onSave={handleSaveQuotations}
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
              <ApprovalActionContent
                requisitionId={requisitionId}
                requisition={requisition as any}
                workflowStatus={approvalData?.workflowStatus}
                isLoading={approvalData?.isLoading || false}
                onApprovalComplete={handleApprovalComplete}
              />
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
                  requisition={requisition as any}
                  approvalChain={requisition?.approvalChain}
                  approvalHistory={approvalData?.approvalHistory || []}
                  workflowStatus={approvalData?.workflowStatus}
                  availableApprovers={approvalData?.availableApprovers || []}
                  isLoading={approvalData?.isLoading || false}
                />
                <WorkflowStatusSummary
                  requisition={requisition as any}
                  workflowStatus={approvalData?.workflowStatus}
                />
              </>
            )}
          </TabsContent>

          {/* ── Tab 5: Activity Log (Timeline) ── */}
          <TabsContent value="activity" className="space-y-4 mt-6">
            <ActivityLogContent
              actionHistory={requisition?.actionHistory}
              auditEvents={auditEventsData}
            />
          </TabsContent>
        </Tabs>
      </Card>

      {/* Create PO from Requisition — 3-step wizard */}
      <POCreationWizard
        open={isCreatePOOpen}
        onOpenChange={setIsCreatePOOpen}
        requisition={requisition}
      />

      {/* PDF Preview Dialog */}
      {previewBlob && (
        <PDFPreviewDialog
          open={previewOpen}
          onOpenChange={setPreviewOpen}
          pdfBlob={previewBlob}
          fileName={`Requisition: ${requisition.documentNumber}`}
          onDownload={handleExportPDF}
        />
      )}

      {/* Edit Requisition Dialog */}
      <CreateRequisitionDialog
        open={isEditDialogOpen}
        onOpenChange={setIsEditDialogOpen}
        onRequisitionCreated={handleDocumentUpdated}
        userId={userId}
        editingRequisition={requisition}
        isEditing={true}
        initialStep={editInitialStep}
      />

      {/* Submit Dialog */}
      <RequisitionSubmitDialog
        open={showSubmitDialog}
        onOpenChange={setShowSubmitDialog}
        requisition={requisition}
        onSubmit={handleSubmitForApproval}
        isSubmitting={submitMutation.isPending}
      />

      {/* Withdraw Confirmation Modal */}
      <ConfirmationModal
        open={showWithdrawModal}
        onOpenChange={setShowWithdrawModal}
        onConfirm={handleWithdraw}
        type="withdraw"
        title="Withdraw Requisition"
        description={`Are you sure you want to withdraw requisition ${requisition.documentNumber || requisition.id}? It will be reverted to draft status and you can edit and re-submit it later.`}
        isLoading={withdrawMutation?.isPending || false}
      />
    </div>
  );
}
