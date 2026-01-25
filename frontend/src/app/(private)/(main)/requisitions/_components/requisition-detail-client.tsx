"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { Card } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
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
} from "lucide-react";
import { PageHeader } from "@/components/base/page-header";
import {
  useRequisitionById,
  useSubmitRequisitionForApproval,
} from "@/hooks/use-requisition-queries";
import { useRequisitionStorage } from "@/hooks/use-requisition-storage";
import { Requisition } from "@/types/requisition";
import { ApprovalHistoryPanel } from "./approval-history-panel";
import { CreateRequisitionDialog } from "./create-requisition-dialog";
import { DocumentLinks } from "@/components/document-links";
import { WorkflowDocument } from "@/types";
import {
  Empty,
  EmptyContent,
  EmptyDescription,
  EmptyMedia,
} from "@/components/ui/empty";
import { Package } from "lucide-react";
import {
  exportRequisitionPDF,
  getRequisitionPDFBlob,
} from "@/lib/pdf/pdf-export";
import { toast } from "sonner";
import { PDFPreviewDialog } from "@/components/modals/pdf-preview-dialog";
import { is } from "date-fns/locale";
import { Badge } from "@/components";

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
  const { saveToStorage } = useRequisitionStorage();
  const [isExporting, setIsExporting] = useState(false);
  const [previewOpen, setPreviewOpen] = useState(false);
  const [previewBlob, setPreviewBlob] = useState<Blob | null>(null);
  const [isEditDialogOpen, setIsEditDialogOpen] = useState(false);

  // Use the new hook with initialData from server component
  const {
    data: requisition,
    isLoading,
    refetch,
  } = useRequisitionById(requisitionId, initialRequisition);

  // Submit mutation
  const submitMutation = useSubmitRequisitionForApproval(requisitionId, () => {
    // After successful submission, refetch to get updated data
    refetch();
  });

  const handlePreviewPDF = async () => {
    if (!requisition) return;
    try {
      setIsExporting(true);
      const blob = await getRequisitionPDFBlob(requisition);
      setPreviewBlob(blob);
      setPreviewOpen(true);
    } catch (error) {
      console.error("PDF preview error:", error);
      toast.error("Failed to generate PDF preview");
    } finally {
      setIsExporting(false);
    }
  };

  const handleExportPDF = async () => {
    if (!requisition) return;
    try {
      setIsExporting(true);
      await exportRequisitionPDF(requisition);
      toast.success("Requisition exported as PDF");
    } catch (error) {
      console.error("PDF export error:", error);
      toast.error("Failed to export PDF");
    } finally {
      setIsExporting(false);
    }
  };

  const handleSubmitForApproval = async () => {
    if (!requisition) return;

    try {
      await submitMutation.mutateAsync({
        submittedBy: userId,
        submittedByName: requisition.requestedByName || "User",
        submittedByRole: requisition.requestedByRole || userRole,
        comments: `Submitted for approval on ${new Date().toLocaleDateString()}`,
      });

      // Also save to localStorage
      if (submitMutation.data?.data) {
        saveToStorage(submitMutation.data.data);
      }
    } catch (error) {
      console.error("Submit error:", error);
    }
  };

  const handleEditRequisition = () => {
    setIsEditDialogOpen(true);
  };

  const handleRequisitionUpdated = () => {
    setIsEditDialogOpen(false);
    refetch(); // Refresh the data
  };

  const isCreator =
    requisition?.requestedBy === userId || requisition?.requesterId === userId;
  const canEdit =
    isCreator &&
    (requisition?.status === "draft" || requisition?.status === "rejected");

  const canSubmit = requisition?.status === "draft" && isCreator;

  if (isLoading && !requisition) {
    return (
      <div className="flex items-center justify-center py-12">
        <div className="text-center">
          <div className="h-8 w-8 rounded-full border-4 border-blue-200 border-t-blue-600 animate-spin mx-auto mb-4"></div>
          <p className="text-gray-600">Loading requisition...</p>
        </div>
      </div>
    );
  }

  if (!requisition) {
    return (
      <div className="flex items-center justify-center py-12">
        <Card className="p-8 max-w-md text-center">
          <AlertCircle className="h-12 w-12 text-gray-400 mx-auto mb-4" />
          <h3 className="font-semibold text-lg mb-2">Requisition Not Found</h3>
          <p className="text-gray-600 mb-6">
            The requisition you're looking for doesn't exist.
          </p>
          <Button variant="outline" onClick={() => router.back()}>
            Go Back
          </Button>
        </Card>
      </div>
    );
  }

  const totalEstimatedCost = requisition?.totalAmount || 0;

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
          {canEdit && (
            <Button
              onClick={handleEditRequisition}
              variant="outline"
              className="gap-2 h-11"
            >
              <Pencil className="h-4 w-4" />
              Edit Requisition
            </Button>
          )}
          {canSubmit && (
            <Button
              onClick={handleSubmitForApproval}
              disabled={submitMutation.isPending}
              isLoading={submitMutation.isPending}
              loadingText="Submitting..."
              className="gap-2 h-11"
            >
              <Send className="h-4 w-4" />
              Submit for Approval
            </Button>
          )}
        </div>
      </div>

      <div className="grid grid-cols-1  gap-6">
        {/* Main Content */}
        <div className="  space-y-6">
          {/* Requisition Details */}
          <div className="gradient-primary border-0 overflow-hidden rounded-lg p-6">
            <h2 className="text-lg font-semibold mb-6 text-primary-foreground">
              Requisition Details
            </h2>

            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              {/* Basic Information */}
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
                      requisition.priority?.toLowerCase() === "urgent"
                        ? "bg-red-100 text-red-800 border-red-200"
                        : requisition.priority?.toLowerCase() === "high"
                          ? "bg-orange-100 text-orange-800 border-orange-200"
                          : requisition.priority?.toLowerCase() === "medium"
                            ? "bg-blue-100 text-blue-800 border-blue-200"
                            : "bg-gray-100 text-gray-800 border-gray-200"
                    }`}
                  >
                    {requisition.priority || "Medium"}
                  </Badge>
                </div>
              </div>

              {/* Requester Information */}
              <div className="space-y-1">
                <label className="text-xs font-semibold text-primary-foreground/80 uppercase tracking-wider flex items-center gap-1">
                  <User className="h-3 w-3" />
                  Requested By
                </label>
                <p className="text-base font-medium text-primary-foreground">
                  {requisition.requesterName ||
                    requisition.requestedByName ||
                    "—"}
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

              {/* Financial Information */}
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

              {/* Dates */}
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
                      {new Date(requisition.updatedAt).toLocaleDateString(
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

              {requisition.requiredByDate && (
                <div className="space-y-1">
                  <label className="text-xs font-semibold text-primary-foreground/80 uppercase tracking-wider flex items-center gap-1">
                    <Calendar className="h-3 w-3" />
                    Due Date
                  </label>
                  <p
                    className={`text-sm font-medium ${
                      new Date(requisition.requiredByDate) < new Date() &&
                      requisition.status !== "completed"
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
                      requisition.status !== "completed" && (
                        <span className="ml-2 text-xs">(Overdue)</span>
                      )}
                  </p>
                </div>
              )}

              {/* Category Information */}
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

              {/* Approval Information */}
              <div className="space-y-1">
                <label className="text-xs font-semibold text-primary-foreground/80 uppercase tracking-wider">
                  Approval Stage
                </label>
                <p className="text-sm font-medium font-mono bg-white/10 px-2 py-1 rounded text-primary-foreground">
                  {requisition.currentApprovalStage &&
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

            {/* Description - Full Width */}
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

            {/* Additional Metadata - Full Width */}
            {requisition.metadata &&
              Object.keys(requisition.metadata).length > 0 && (
                <div className="mt-6 pt-6 border-t border-white/20">
                  <label className="text-xs font-semibold text-primary-foreground/80 uppercase tracking-wider block mb-3">
                    Additional Information
                  </label>
                  <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                    {Object.entries(requisition.metadata).map(
                      ([key, value]) => (
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
                      ),
                    )}
                  </div>
                </div>
              )}

            {/* Auto-Created Purchase Order - Full Width */}
            {requisition?.automationUsed && requisition?.autoCreatedPO && (
              <div className="mt-6 pt-6 border-t border-white/20">
                <label className="text-xs font-semibold text-primary-foreground/80 uppercase tracking-wider  mb-3 flex items-center gap-2">
                  <span className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-green-100 text-green-800 border border-green-200">
                    ✓ Automated
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

          {/* Document Links */}
          {requisition.status === "approved" && (
            <DocumentLinks
              currentDocument={requisition as unknown as WorkflowDocument}
              linkedDocuments={{}}
            />
          )}

          {/* Items Section */}
          <Card className="p-6 border-0 shadow-sm">
            <div className="flex items-center justify-between mb-6">
              <h2 className="text-lg font-semibold">
                Items ({requisition.items?.length || 0})
              </h2>
              {requisition.isEstimate && (
                <span className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-yellow-100 text-yellow-800 border border-yellow-200">
                  Estimated Costs
                </span>
              )}
            </div>

            {requisition.items && requisition.items.length > 0 ? (
              <>
                <div className="space-y-3">
                  {requisition.items.map((item: any, index: number) => (
                    <div
                      key={item.id || index}
                      className="flex items-start justify-between p-4 rounded-lg border border-slate-200/10 hover:border-slate-300/20 hover:bg-slate-50/20 transition"
                    >
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center gap-2 mb-2">
                          <span className="inline-flex items-center justify-center w-6 h-6 rounded-full bg-slate-100 dark:bg-slate-600/10 text-xs font-semibold">
                            {index + 1}
                          </span>
                          <p className="font-medium text-foreground">
                            {item.description ||
                              item.itemDescription ||
                              "No description"}
                          </p>
                        </div>

                        <div className="ml-8 space-y-1">
                          <div className="flex items-center gap-4 text-sm text-muted-foreground">
                            <span>
                              <strong>Quantity:</strong> {item.quantity || 0}
                              {item.unit && (
                                <span className="ml-1">({item.unit})</span>
                              )}
                            </span>
                            <span>
                              <strong>Unit Price:</strong>{" "}
                              {requisition.currency}{" "}
                              {(
                                item.unitPrice ||
                                item.estimatedCost ||
                                0
                              ).toLocaleString("en-ZM", {
                                minimumFractionDigits: 2,
                                maximumFractionDigits: 2,
                              })}
                            </span>
                          </div>

                          {item.category && (
                            <div className="text-xs text-muted-foreground">
                              <strong>Category:</strong> {item.category}
                            </div>
                          )}

                          {item.notes && (
                            <div className="text-xs text-muted-foreground">
                              <strong>Notes:</strong> {item.notes}
                            </div>
                          )}

                          {item.id && (
                            <div className="text-xs text-muted-foreground">
                              <strong>Item ID:</strong> {item.id}
                            </div>
                          )}
                        </div>
                      </div>

                      <div className="text-right ml-4">
                        <p className="font-semibold text-lg">
                          {requisition.currency}{" "}
                          {(
                            item.amount ||
                            item.totalPrice ||
                            item.quantity *
                              (item.unitPrice || item.estimatedCost || 0)
                          ).toLocaleString("en-ZM", {
                            minimumFractionDigits: 2,
                            maximumFractionDigits: 2,
                          })}
                        </p>
                        <p className="text-xs text-muted-foreground">Total</p>
                      </div>
                    </div>
                  ))}
                </div>

                {/* Summary */}
                <div className="mt-6 pt-6 border-t bg-slate-50 dark:bg-slate-950 -mx-6 -mb-6 px-6 py-4 rounded-b-lg">
                  <div className="flex items-center justify-between mb-2">
                    <span className="text-sm text-muted-foreground">
                      Total Items: {requisition.items.length}
                    </span>
                    <span className="text-sm text-muted-foreground">
                      Currency: {requisition.currency}
                    </span>
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="font-semibold text-foreground">
                      Estimated Cost
                    </span>
                    <span className="text-2xl font-bold text-emerald-700 dark:text-emerald-400">
                      {requisition.currency}{" "}
                      {totalEstimatedCost.toLocaleString("en-ZM", {
                        minimumFractionDigits: 2,
                        maximumFractionDigits: 2,
                      })}
                    </span>
                  </div>
                </div>
              </>
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
          </Card>

          {/* Action History Panel */}
          <ApprovalHistoryPanel
            requisitionId={requisition?.id || requisitionId}
            requisition={requisition as any}
            userRole={userRole}
            actionHistory={requisition?.actionHistory}
            approvalChain={requisition?.approvalChain}
          />
        </div>

        {/* Sidebar - Empty for now, could be used for other widgets */}
        <div className="lg:col-span-1">
          {/* Future: Quick actions, related documents, etc. */}
        </div>
      </div>

      {/* PDF Preview Dialog */}
      {previewBlob && (
        <PDFPreviewDialog
          open={previewOpen}
          onOpenChange={setPreviewOpen}
          pdfBlob={previewBlob}
          fileName={`REQ-${requisition.documentNumber}.pdf`}
          onDownload={handleExportPDF}
        />
      )}

      {/* Edit Requisition Dialog */}
      <CreateRequisitionDialog
        open={isEditDialogOpen}
        onOpenChange={setIsEditDialogOpen}
        onRequisitionCreated={handleRequisitionUpdated}
        userId={userId}
        editingRequisition={requisition}
        isEditing={true}
      />
    </div>
  );
}
