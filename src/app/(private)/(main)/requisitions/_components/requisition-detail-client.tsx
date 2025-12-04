"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { Card } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Send, AlertCircle, Download, Eye } from "lucide-react";
import { PageHeader } from "@/components/base/page-header";
import { useRequisitionById, useSubmitRequisitionForApproval } from "@/hooks/use-requisition-queries";
import { useRequisitionStorage } from "@/hooks/use-requisition-storage";
import { Requisition } from "@/types/requisition";
import { ApprovalHistoryPanel } from "./approval-history-panel";
import { ActionHistoryPanel } from "./action-history-panel";
import { EditRequisitionPanel } from "./edit-requisition-panel";
import { DocumentLinks } from "@/components/document-links";
import { WorkflowDocument } from "@/types";
import {
  Empty,
  EmptyContent,
  EmptyDescription,
  EmptyMedia,
} from "@/components/ui/empty";
import { Package } from "lucide-react";
import { exportRequisitionPDF, getRequisitionPDFBlob } from "@/lib/pdf/pdf-export";
import { toast } from "sonner";
import { PDFPreviewDialog } from "@/components/pdf-preview-dialog";

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
      console.error('PDF preview error:', error);
      toast.error('Failed to generate PDF preview');
    } finally {
      setIsExporting(false);
    }
  };

  const handleExportPDF = async () => {
    if (!requisition) return;
    try {
      setIsExporting(true);
      await exportRequisitionPDF(requisition);
      toast.success('Requisition exported as PDF');
    } catch (error) {
      console.error('PDF export error:', error);
      toast.error('Failed to export PDF');
    } finally {
      setIsExporting(false);
    }
  };

  const handleSubmitForApproval = async () => {
    if (!requisition) return;

    try {
      await submitMutation.mutateAsync({
        submittedBy: userId,
        submittedByName: (requisition.requestedByName || 'User'),
        submittedByRole: (requisition.requestedByRole || userRole),
        comments: `Submitted for approval on ${new Date().toLocaleDateString()}`,
      });

      // Also save to localStorage
      if (submitMutation.data?.data) {
        saveToStorage(submitMutation.data.data);
      }
    } catch (error) {
      console.error('Submit error:', error);
    }
  };

  const isCreator = requisition?.requestedBy === userId;
  const canEdit =
    isCreator &&
    (requisition?.status === "DRAFT" || requisition?.status === "REJECTED");
  const canSubmit = canEdit;

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

  const totalItems = requisition?.items?.length || 0;
  const totalEstimatedCost = requisition?.totalAmount || 0;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-start justify-between gap-4">
        <PageHeader
          title={requisition.requisitionNumber}
          subtitle={`Created ${new Date(requisition.createdAt).toLocaleDateString("en-ZM", { year: "numeric", month: "long", day: "numeric" })}`}
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
            {isExporting ? "Loading..." : "Preview"}
          </Button>
          <Button
            onClick={handleExportPDF}
            disabled={isExporting}
            variant="outline"
            className="gap-2 h-11"
          >
            <Download className="h-4 w-4" />
            {isExporting ? "Exporting..." : "Export PDF"}
          </Button>
          {canSubmit && (
            <Button
              onClick={handleSubmitForApproval}
              disabled={submitMutation.isPending}
              className="gap-2 h-11"
            >
              <Send className="h-4 w-4" />
              {submitMutation.isPending ? "Submitting..." : "Submit for Approval"}
            </Button>
          )}
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Main Content */}
        <div className="lg:col-span-2 space-y-6">
          {/* Requisition Details */}
          <div className="gradient-primary border-0 overflow-hidden rounded-lg p-6">
            <h2 className="text-lg font-semibold mb-6 text-primary-foreground">
              Requisition Details
            </h2>

            <div className="grid grid-cols-2 gap-8">
              <div className="space-y-1">
                <label className="text-xs font-semibold text-primary-foreground/80 uppercase tracking-wider">
                  Department
                </label>
                <p className="text-base font-medium text-primary-foreground">
                  {requisition.department || "—"}
                </p>
              </div>

              <div className="space-y-1">
                <label className="text-xs font-semibold text-primary-foreground/80 uppercase tracking-wider">
                  Priority
                </label>
                <p className="text-base font-medium text-primary-foreground">
                  {requisition.priority || "—"}
                </p>
              </div>

              <div className="col-span-2 space-y-1">
                <label className="text-xs font-semibold text-primary-foreground/80 uppercase tracking-wider">
                  Description
                </label>
                <p className="text-sm text-primary-foreground leading-relaxed">
                  {requisition.description || "No description provided"}
                </p>
              </div>

              <div className="space-y-1">
                <label className="text-xs font-semibold text-primary-foreground/80 uppercase tracking-wider">
                  Budget Code
                </label>
                <p className="text-sm font-medium font-mono bg-white/10 px-2 py-1 rounded text-primary-foreground">
                  {requisition.budgetCode || "—"}
                </p>
              </div>

              <div className="space-y-1">
                <label className="text-xs font-semibold text-primary-foreground/80 uppercase tracking-wider">
                  Approval Stage
                </label>
                <p className="text-sm font-medium font-mono bg-white/10 px-2 py-1 rounded text-primary-foreground">
                  {requisition.currentApprovalStage &&
                  requisition.totalApprovalStages
                    ? `${requisition.currentApprovalStage}/${requisition.totalApprovalStages}`
                    : "—"}
                </p>
              </div>
            </div>
          </div>

          {/* Document Links */}
          {requisition.status === "APPROVED" && (
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
            </div>

            {requisition.items && requisition.items.length > 0 ? (
              <>
                <div className="space-y-3">
                  {requisition.items.map((item: any, index: number) => (
                    <div
                      key={item.id}
                      className="flex items-start justify-between p-4 rounded-lg border border-slate-200/10 hover:border-slate-300/20 hover:bg-slate-50/20 transition"
                    >
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center gap-2 mb-1">
                          <span className="inline-flex items-center justify-center w-6 h-6 rounded-full bg-slate-100 dark:bg-slate-600/10 text-xs font-semibold">
                            {index + 1}
                          </span>
                          <p className="font-medium text-foreground truncate">
                            {item.description}
                          </p>
                        </div>
                        <p className="text-xs text-muted-foreground">
                          {item.quantity} × ZMW{" "}
                          {item.unitPrice.toLocaleString("en-ZM", {
                            minimumFractionDigits: 2,
                            maximumFractionDigits: 2,
                          })}
                        </p>
                      </div>
                      <div className="text-right ml-4">
                        <p className="font-semibold text-sm">
                          ZMW{" "}
                          {item.totalPrice.toLocaleString("en-ZM", {
                            minimumFractionDigits: 2,
                            maximumFractionDigits: 2,
                          })}
                        </p>
                      </div>
                    </div>
                  ))}
                </div>

                {/* Total */}
                <div className="mt-6 pt-6 border-t flex items-center justify-between bg-slate-50 dark:bg-slate-950  -mx-6 -mb-6 px-6 py-4 rounded-b-lg">
                  <span className="font-semibold text-foreground">
                    Total Estimated Cost
                  </span>
                  <span className="text-2xl font-bold text-emerald-700 dark:text-emerald-400">
                    ZMW{" "}
                    {totalEstimatedCost.toLocaleString("en-ZM", {
                      minimumFractionDigits: 2,
                      maximumFractionDigits: 2,
                    })}
                  </span>
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

          {/* Edit Panel - Only for Creator in DRAFT/REJECTED status */}
          {canEdit && (
            <EditRequisitionPanel
              requisition={requisition as any}
              onRequisitionUpdated={refetch}
            />
          )}

          {/* Action History Panel */}
          <ActionHistoryPanel
            actionHistory={requisition.actionHistory}
            approvalChain={requisition.approvalChain}
          />
        </div>

        {/* Sidebar - Approval History */}
        <div className="lg:col-span-1">
          <ApprovalHistoryPanel
            requisitionId={requisitionId}
            requisition={requisition as any}
            userRole={userRole}
          />
        </div>
      </div>

      {/* PDF Preview Dialog */}
      {previewBlob && (
        <PDFPreviewDialog
          open={previewOpen}
          onOpenChange={setPreviewOpen}
          pdfBlob={previewBlob}
          fileName={`REQ-${requisition.requisitionNumber}.pdf`}
          onDownload={handleExportPDF}
        />
      )}
    </div>
  );
}
