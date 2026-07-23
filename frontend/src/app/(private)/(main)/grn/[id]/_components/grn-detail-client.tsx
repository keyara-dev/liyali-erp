"use client";

import { useMemo, useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { toast } from "sonner";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import {
  Package,
  FileText,
  AlertTriangle,
  AlertCircle,
  Plus,
  Send,
  CheckCircle2,
  ClipboardList,
  Layers,
  Warehouse,
  Calendar,
  User,
  Link as LinkIcon,
  Eye,
  Download,
} from "lucide-react";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { PageHeader } from "@/components/base/page-header";
import { DocumentLoadingPage } from "@/components/base/document-loading-page";
import ErrorDisplay from "@/components/base/error-display";
import { GRNItemsMatchingTable } from "./grn-items-matching-table";
import { QualityIssueReportDialog } from "./quality-issue-dialog";
import { useAddQualityIssueMutation } from "@/hooks/use-quality-issue-mutations";
import { useGRNDetail } from "@/hooks/use-grn-detail";
import { useOrganizationMembersQuery } from "@/hooks/use-organization-queries";
import { usePurchaseOrders } from "@/hooks/use-purchase-order-queries";
import type { PurchaseOrder } from "@/types/purchase-order";
import { usePaymentVouchers } from "@/hooks/use-payment-voucher-queries";
import type { PaymentVoucher } from "@/types/payment-voucher";
import { Badge } from "@/components";
import type { QualityIssue } from "@/types/goods-received-note";
import { GRNSubmitDialog } from "./grn-submit-dialog";
import { GRNSignoffPanel } from "./grn-signoff-panel";
import { PDFPreviewDialog } from "@/components/modals/pdf-preview-dialog";
import { type LinkedDoc } from "@/components/linked-documents";
import {
  SupportingDocuments,
  type UploadedAttachment,
} from "@/components/supporting-documents";
import { cn } from "@/lib/utils";
import { completeGRNAction, updateGRNAction } from "@/app/_actions/grn-actions";
import { usePermissions } from "@/hooks/use-permissions";
import { useQueryClient } from "@tanstack/react-query";
import { QUERY_KEYS } from "@/lib/constants";

interface GRNDetailClientProps {
  grnId: string;
  userId: string;
  userRole: string;
}

export function GRNDetailClient({
  grnId,
  userId,
  userRole,
}: GRNDetailClientProps) {
  const router = useRouter();
  const queryClient = useQueryClient();
  const [isQualityDialogOpen, setIsQualityDialogOpen] = useState(false);
  const [isCompleting, setIsCompleting] = useState(false);

  const {
    document: grn,
    isLoading,
    permissions,
    showSubmitDialog,
    setShowSubmitDialog,
    handleSubmitForApproval,
    submitMutation,
    isExporting,
    previewOpen,
    setPreviewOpen,
    previewBlob,
    handlePreviewPDF,
    handleExportPDF,
    refetch,
  } = useGRNDetail({
    grnId,
    userId,
    userRole,
  });

  const { hasPermission } = usePermissions();

  const addQualityIssueMutation = useAddQualityIssueMutation(grnId);

  // Resolve the linked PO so we can deep-link to /purchase-orders/{id}.
  // GRN only carries poDocumentNumber, so we cross-reference against the PO
  // list. Covered by the same query as the GRN list page, so the cache is
  // usually warm when arriving from there.
  const { data: purchaseOrders = [] } = usePurchaseOrders();
  const linkedPO = useMemo(() => {
    if (!grn?.poDocumentNumber) return undefined;
    return (purchaseOrders as PurchaseOrder[]).find(
      (po) => po.documentNumber === grn.poDocumentNumber,
    );
  }, [purchaseOrders, grn?.poDocumentNumber]);

  // Resolve linkedPV doc number → PV id for navigation.
  const { data: paymentVouchers = [] } = usePaymentVouchers();
  const linkedPVRecord = useMemo(() => {
    if (!grn?.linkedPV) return undefined;
    return (paymentVouchers as PaymentVoucher[]).find(
      (pv) => pv.documentNumber === grn.linkedPV,
    );
  }, [paymentVouchers, grn?.linkedPV]);

  // Linked procurement documents — REQ (via the PO) → PO → PV. GRN carries no
  // chain object, so resolve from the cross-referenced PO and PV records.
  const grnLinkedDocs = useMemo<LinkedDoc[]>(() => {
    const links: LinkedDoc[] = [];
    if (linkedPO?.sourceRequisitionId) {
      links.push({
        type: "requisition",
        label: "Requisition",
        id: linkedPO.sourceRequisitionId,
        documentNumber:
          linkedPO.linkedRequisition || linkedPO.sourceRequisitionId,
      });
    }
    if (linkedPO?.id) {
      links.push({
        type: "purchase-order",
        label: "Purchase Order",
        id: linkedPO.id,
        documentNumber: grn?.poDocumentNumber || linkedPO.documentNumber,
        status: linkedPO.status,
      });
    }
    if (linkedPVRecord?.id) {
      links.push({
        type: "payment-voucher",
        label: "Payment Voucher",
        id: linkedPVRecord.id,
        documentNumber: grn?.linkedPV || linkedPVRecord.documentNumber,
        status: linkedPVRecord.status,
      });
    }
    return links;
  }, [linkedPO, linkedPVRecord, grn?.poDocumentNumber, grn?.linkedPV]);

  // Org members for resolving user IDs → names (Received By, Approved By).
  // Backend returns a paginated shape { members, total, ... } for page >= 1 —
  // normalize defensively so a shape change doesn't crash the page.
  const { data: membersData } = useOrganizationMembersQuery(1, 100);
  const userLookup = useMemo(() => {
    const map = new Map<string, { name?: string; email?: string }>();
    const list: Array<Record<string, any>> = Array.isArray(membersData)
      ? (membersData as Array<Record<string, any>>)
      : ((membersData as any)?.members ?? []);
    for (const m of list) {
      // `user_id` is the canonical user identifier on an org_members row;
      // `id` is the membership row id. Backends may return either, so check both.
      const id = m?.user_id || m?.userId || m?.id;
      if (id) map.set(id, { name: m.name, email: m.email });
    }
    return map;
  }, [membersData]);

  const resolveUser = (id?: string) => {
    if (!id) return "—";
    const hit = userLookup.get(id);
    return hit?.name || hit?.email || id;
  };

  const handleBack = () => {
    router.back();
  };

  const handleMarkComplete = async () => {
    if (!grn) return;
    if (grn.signoffStatus !== "READY") {
      toast.error("Both signatures are required before completing the GRN");
      return;
    }
    setIsCompleting(true);
    try {
      const res = await completeGRNAction({ grnId });
      if (res.success) {
        toast.success("GRN marked complete");
        // Completion cascades to PO.delivery_status + may auto-create a PV
        // via AutoCreatePVFromPO. Invalidate everything downstream.
        queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.GRN.ALL] });
        queryClient.invalidateQueries({
          queryKey: [QUERY_KEYS.GRN.BY_ID, grnId],
        });
        queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.PURCHASE_ORDERS.ALL] });
        queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.PURCHASE_ORDERS.BY_ID] });
        queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.PAYMENT_VOUCHERS.ALL] });
        queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.PAYMENT_VOUCHERS.BY_ID] });
        queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.DASHBOARD.METRICS] });
      } else {
        toast.error(res.message || "Failed to complete GRN");
      }
    } finally {
      setIsCompleting(false);
    }
  };

  const handleAddQualityIssue = async (issue: Omit<QualityIssue, "id">) => {
    try {
      await addQualityIssueMutation.mutateAsync(issue);
      toast.success("Quality issue reported and saved");
    } catch (error) {
      console.error("Error saving quality issue:", error);
      toast.error("Failed to save quality issue");
    }
  };

  if (isLoading) return <DocumentLoadingPage />;

  if (!grn)
    return (
      <ErrorDisplay
        title="GRN Not Found"
        message="The goods received note you're looking for doesn't exist."
      />
    );

  // GRN never had its own attachments UI before — metadata.attachments is a
  // new read path introduced alongside SupportingDocuments (Task C3).
  const grnAttachments: UploadedAttachment[] =
    (grn.metadata?.attachments as UploadedAttachment[]) ?? [];

  // Task C2 made GRN metadata-only patches (e.g. attachments) allowed on any
  // status, specifically so uploads aren't blocked once the GRN leaves DRAFT.
  const canUploadGRN = permissions.isCreator || hasPermission("grn", "edit");

  // Persists a SupportingDocuments upload (already on ImageKit) into the
  // GRN's own metadata.attachments — read-modify-write via updateGRNAction.
  const handleSupportingDocUpload = async (att: UploadedAttachment) => {
    await updateGRNAction(grnId, {
      metadata: { attachments: [...grnAttachments, att] },
    });
    refetch();
  };

  const qualityIssues = grn.qualityIssues ?? [];
  const hasQualityIssues = qualityIssues.length > 0;
  const items = grn.items ?? [];
  const hasVariances = items.some(
    (item: { variance: number }) => item.variance !== 0,
  );
  const goodCount = items.filter(
    (i: { condition: string }) => i.condition?.toLowerCase() === "good",
  ).length;

  // Stage display is status-aware. A GRN has no workflow stage until submitted,
  // so show "Awaiting submission" in DRAFT rather than leaking `undefined / 2`.
  const statusKey = grn.status?.toUpperCase();
  const totalStages = 2;
  const currentStageNum = Number(grn.currentStage) || 0;
  const isDraft = statusKey === "DRAFT";
  const isTerminal = statusKey === "APPROVED" || statusKey === "COMPLETED";
  const stagePrimary = isDraft
    ? "Awaiting submission"
    : isTerminal
      ? "Completed"
      : currentStageNum > 0
        ? `${currentStageNum} / ${totalStages}`
        : "—";
  const stageSecondary = isDraft
    ? "Submit to start the approval workflow"
    : grn.stageName || (isTerminal ? "Approval complete" : "In progress");
  const stagePercent = isDraft
    ? 0
    : isTerminal
      ? 100
      : Math.min(100, Math.round((currentStageNum / totalStages) * 100));

  const headerActions = (
    <div className="flex flex-wrap items-center gap-2">
      <Button
        size="sm"
        variant="outline"
        onClick={handlePreviewPDF}
        disabled={isExporting}
        isLoading={isExporting}
        loadingText="Loading..."
        className="gap-1.5"
      >
        <Eye className="h-3.5 w-3.5" />
        Preview
      </Button>
      <Button
        size="sm"
        variant="outline"
        onClick={handleExportPDF}
        disabled={isExporting}
        isLoading={isExporting}
        loadingText="Exporting..."
        className="gap-1.5"
      >
        <Download className="h-3.5 w-3.5" />
        Export PDF
      </Button>
      {permissions.canSubmit && grn.signoffStatus === "READY" && (
        <>
          <Button
            size="sm"
            onClick={() => setShowSubmitDialog(true)}
            className="gap-1.5"
          >
            <Send className="h-3.5 w-3.5" />
            Submit to Workflow
          </Button>
          <Button
            size="sm"
            variant="outline"
            onClick={handleMarkComplete}
            disabled={isCompleting}
            isLoading={isCompleting}
            loadingText="Completing..."
            className="gap-1.5"
          >
            <CheckCircle2 className="h-3.5 w-3.5" />
            Mark Complete
          </Button>
        </>
      )}
    </div>
  );

  return (
    <div className="space-y-5">
      {/* Header */}
      <PageHeader
        title={grn.documentNumber}
        subtitle="Goods Received Note"
        badges={[
          {
            status: grn.status,
            type: "document",
          },
        ]}
        onBackClick={handleBack}
        showBackButton={true}
        actions={headerActions}
      />

      {/* Compact stat strip */}
      <Card className="border-border/60 p-0">
        <CardContent className="grid grid-cols-2 md:grid-cols-4 divide-y md:divide-y-0 md:divide-x divide-border/60 p-0">
          <StatCell
            icon={<Layers className="h-4 w-4" />}
            label="Stage"
            primary={stagePrimary}
            secondary={stageSecondary}
            accent={isDraft ? "slate" : isTerminal ? "emerald" : "blue"}
            progress={stagePercent}
          />
          <StatCell
            icon={<Package className="h-4 w-4" />}
            label="Items"
            primary={String(items.length)}
            secondary={`${goodCount} in good condition`}
            accent="emerald"
          />
          <StatCell
            icon={<ClipboardList className="h-4 w-4" />}
            label="Variances"
            primary={hasVariances ? "Yes" : "None"}
            secondary={
              hasVariances ? "Some qty mismatches" : "All qty match PO"
            }
            accent={hasVariances ? "amber" : "slate"}
          />
          <StatCell
            icon={<AlertTriangle className="h-4 w-4" />}
            label="Quality Issues"
            primary={String(qualityIssues.length)}
            secondary={
              hasQualityIssues ? "Reported during intake" : "None reported"
            }
            accent={hasQualityIssues ? "rose" : "slate"}
          />
        </CardContent>
      </Card>

      {/* Sign-off (Received By + Certified By). Only relevant while DRAFT;
          completed / approved GRNs render the signature blocks inside the PDF. */}
      {isDraft && (
        <GRNSignoffPanel
          grn={grn}
          userId={userId}
          userRole={userRole}
          defaultReceiverName={resolveUser(grn.receivedBy)}
        />
      )}

      {/* GRN Information — compact inline */}
      <Card className="border-border/60">
        <CardContent className="p-4">
          <div className="flex items-center gap-2 mb-3">
            <FileText className="h-4 w-4 text-muted-foreground" />
            <span className="text-sm font-semibold">GRN Information</span>
          </div>
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-x-6 gap-y-3 text-sm">
            <InfoField
              icon={<LinkIcon className="h-3.5 w-3.5" />}
              label="PO Reference"
              value={
                grn.poDocumentNumber ? (
                  linkedPO ? (
                    <Link
                      href={`/purchase-orders/${linkedPO.id}`}
                      className="font-mono text-blue-600 hover:underline"
                    >
                      {grn.poDocumentNumber}
                    </Link>
                  ) : (
                    <span className="font-mono text-muted-foreground">
                      {grn.poDocumentNumber}
                    </span>
                  )
                ) : (
                  "—"
                )
              }
            />
            <InfoField
              icon={<Warehouse className="h-3.5 w-3.5" />}
              label="Warehouse"
              value={grn.warehouseLocation || "—"}
            />
            <InfoField
              icon={<Calendar className="h-3.5 w-3.5" />}
              label="Received Date"
              value={
                grn.receivedDate
                  ? new Date(grn.receivedDate).toLocaleDateString("en-ZM", {
                      year: "numeric",
                      month: "short",
                      day: "numeric",
                    })
                  : "—"
              }
            />
            <InfoField
              icon={<User className="h-3.5 w-3.5" />}
              label="Received By"
              value={resolveUser(grn.receivedBy)}
            />
            {grn.approvedBy && (
              <InfoField
                icon={<CheckCircle2 className="h-3.5 w-3.5" />}
                label="Approved By"
                value={resolveUser(grn.approvedBy)}
              />
            )}
            {grn.linkedPV && (
              <InfoField
                icon={<LinkIcon className="h-3.5 w-3.5" />}
                label="Source Payment Voucher"
                value={
                  linkedPVRecord?.id ? (
                    <Link
                      href={`/payment-vouchers/${linkedPVRecord.id}`}
                      className="font-mono text-blue-600 hover:underline"
                    >
                      {grn.linkedPV}
                    </Link>
                  ) : (
                    <span className="font-mono">{grn.linkedPV}</span>
                  )
                }
              />
            )}
          </div>
        </CardContent>
      </Card>

      {/* Inline banners for urgent issues */}
      {(hasQualityIssues || hasVariances) && (
        <div className="space-y-2">
          {hasQualityIssues && (
            <div className="flex items-start gap-2.5 rounded-md border border-yellow-200 bg-yellow-50 dark:bg-yellow-950/30 dark:border-yellow-800 p-3 text-sm">
              <AlertTriangle className="h-4 w-4 text-yellow-600 dark:text-yellow-400 shrink-0 mt-0.5" />
              <div>
                <p className="font-medium text-yellow-900 dark:text-yellow-200">
                  {qualityIssues.length} quality issue
                  {qualityIssues.length !== 1 ? "s" : ""} reported
                </p>
                <p className="text-xs text-yellow-800 dark:text-yellow-300">
                  See the Reports &amp; Issues tab for details.
                </p>
              </div>
            </div>
          )}
          {hasVariances && (
            <div className="flex items-start gap-2.5 rounded-md border border-orange-200 bg-orange-50 dark:bg-orange-950/30 dark:border-orange-800 p-3 text-sm">
              <AlertCircle className="h-4 w-4 text-orange-600 dark:text-orange-400 shrink-0 mt-0.5" />
              <div>
                <p className="font-medium text-orange-900 dark:text-orange-200">
                  Quantity variances detected
                </p>
                <p className="text-xs text-orange-800 dark:text-orange-300">
                  Some items received differ from PO quantities.
                </p>
              </div>
            </div>
          )}
        </div>
      )}

      {/* Linked procurement chain documents + chain-wide supporting files */}
      <SupportingDocuments
        documentId={grnId}
        documentType="grn"
        chainDocs={grnLinkedDocs}
        canUpload={canUploadGRN}
        onUpload={handleSupportingDocUpload}
        showViewLinks={userRole.toLowerCase() !== "requester"}
      />

      {/* Tabs: Items / Reports & Issues */}
      <Tabs defaultValue="items" className="w-full">
        <TabsList>
          <TabsTrigger value="items" className="gap-2">
            <Package className="h-4 w-4" />
            Items
            <span className="text-xs text-muted-foreground">
              ({items.length})
            </span>
          </TabsTrigger>
          <TabsTrigger value="reports" className="gap-2">
            <ClipboardList className="h-4 w-4" />
            Reports &amp; Issues
            {(hasQualityIssues || (grn.notes && grn.notes.length > 0)) && (
              <span className="text-xs text-muted-foreground">
                ({qualityIssues.length})
              </span>
            )}
          </TabsTrigger>
        </TabsList>

        <TabsContent value="items" className="mt-4">
          <GRNItemsMatchingTable items={items} />
        </TabsContent>

        <TabsContent value="reports" className="mt-4 space-y-4">
          {/* Quality Issues */}
          <div className="rounded-lg border border-border">
            <div className="flex items-center justify-between p-4 border-b">
              <div className="flex items-center gap-2">
                <AlertTriangle className="h-4 w-4 text-muted-foreground" />
                <span className="text-sm font-semibold">Quality Issues</span>
                <Badge variant="secondary" className="text-xs">
                  {qualityIssues.length}
                </Badge>
              </div>
              <Button
                size="sm"
                variant="outline"
                onClick={() => setIsQualityDialogOpen(true)}
                className="gap-1.5"
              >
                <Plus className="h-3.5 w-3.5" />
                Report Issue
              </Button>
            </div>
            <div className="p-4">
              {hasQualityIssues ? (
                <div className="space-y-2">
                  {qualityIssues.map(
                    (
                      issue: {
                        id?: string;
                        itemDescription: string;
                        issueType: string;
                        description: string;
                        severity: string;
                      },
                      index: number,
                    ) => {
                      const severityKey = issue.severity?.toUpperCase();
                      const severityStyles: Record<string, string> = {
                        LOW: "border-yellow-200 bg-yellow-50 dark:border-yellow-800 dark:bg-yellow-950/30",
                        MEDIUM:
                          "border-orange-200 bg-orange-50 dark:border-orange-800 dark:bg-orange-950/30",
                        HIGH: "border-red-200 bg-red-50 dark:border-red-800 dark:bg-red-950/30",
                      };
                      return (
                        <div
                          key={issue.id || index}
                          className={cn(
                            "p-3 border rounded-md",
                            severityStyles[severityKey] ||
                              "border-border bg-muted/30",
                          )}
                        >
                          <div className="flex items-start justify-between gap-3">
                            <div className="flex-1 min-w-0">
                              <p className="font-medium text-sm truncate">
                                {issue.itemDescription}
                              </p>
                              <p className="text-[11px] text-muted-foreground uppercase tracking-wider">
                                {issue.issueType}
                              </p>
                              <p className="text-sm mt-1.5">
                                {issue.description}
                              </p>
                            </div>
                            <Badge
                              variant="outline"
                              className="text-xs shrink-0"
                            >
                              {issue.severity}
                            </Badge>
                          </div>
                        </div>
                      );
                    },
                  )}
                </div>
              ) : (
                <div className="text-center py-6">
                  <p className="text-sm text-muted-foreground">
                    No quality issues reported yet
                  </p>
                  <p className="text-xs text-muted-foreground mt-1">
                    Click &ldquo;Report Issue&rdquo; to add quality concerns
                    during inspection
                  </p>
                </div>
              )}
            </div>
          </div>

          {/* Notes */}
          {grn.notes && (
            <div className="rounded-lg border border-border">
              <div className="flex items-center gap-2 p-4 border-b">
                <FileText className="h-4 w-4 text-muted-foreground" />
                <span className="text-sm font-semibold">Notes</span>
              </div>
              <div className="p-4">
                <p className="text-sm whitespace-pre-wrap">{grn.notes}</p>
              </div>
            </div>
          )}
        </TabsContent>
      </Tabs>

      {/* Submit / Confirm dialogs */}
      <GRNSubmitDialog
        open={showSubmitDialog}
        onOpenChange={setShowSubmitDialog}
        grn={grn}
        onSubmit={handleSubmitForApproval}
        isSubmitting={submitMutation.isPending}
      />

      <QualityIssueReportDialog
        open={isQualityDialogOpen}
        onOpenChange={setIsQualityDialogOpen}
        items={items}
        onAddIssue={handleAddQualityIssue}
      />

      {/* PDF Preview Dialog */}
      {previewBlob && (
        <PDFPreviewDialog
          open={previewOpen}
          onOpenChange={setPreviewOpen}
          pdfBlob={previewBlob}
          fileName={`Goods Received Note: ${grn.documentNumber}`}
          onDownload={handleExportPDF}
        />
      )}
    </div>
  );
}

// ── Sub-components ──────────────────────────────────────────────────────────

interface StatCellProps {
  icon: React.ReactNode;
  label: string;
  primary: string;
  secondary?: string;
  accent?: "blue" | "emerald" | "amber" | "rose" | "slate";
  progress?: number; // 0–100 for the bottom progress bar
}

const ACCENT_CLASSES = {
  blue: "bg-blue-100 text-blue-700 dark:bg-blue-950/50 dark:text-blue-300",
  emerald:
    "bg-emerald-100 text-emerald-700 dark:bg-emerald-950/50 dark:text-emerald-300",
  amber: "bg-amber-100 text-amber-700 dark:bg-amber-950/50 dark:text-amber-300",
  rose: "bg-rose-100 text-rose-700 dark:bg-rose-950/50 dark:text-rose-300",
  slate: "bg-slate-100 text-slate-600 dark:bg-slate-800/60 dark:text-slate-300",
} as const;

const PROGRESS_CLASSES = {
  blue: "bg-blue-500",
  emerald: "bg-emerald-500",
  amber: "bg-amber-500",
  rose: "bg-rose-500",
  slate: "bg-slate-400",
} as const;

function StatCell({
  icon,
  label,
  primary,
  secondary,
  accent = "slate",
  progress,
}: StatCellProps) {
  return (
    <div className="p-4 space-y-1.5">
      <div className="flex items-center gap-2">
        <span
          className={cn(
            "flex items-center justify-center rounded-md h-6 w-6",
            ACCENT_CLASSES[accent],
          )}
        >
          {icon}
        </span>
        <span className="text-xs font-medium text-muted-foreground uppercase tracking-wider">
          {label}
        </span>
      </div>
      <div className="text-xl font-bold tabular-nums leading-tight">
        {primary}
      </div>
      {secondary && (
        <p className="text-xs text-muted-foreground">{secondary}</p>
      )}
      {typeof progress === "number" && (
        <div className="h-1 bg-muted rounded-full overflow-hidden mt-1">
          <div
            className={cn("h-full transition-all", PROGRESS_CLASSES[accent])}
            style={{ width: `${progress}%` }}
          />
        </div>
      )}
    </div>
  );
}

interface InfoFieldProps {
  icon: React.ReactNode;
  label: string;
  value: React.ReactNode;
}

function InfoField({ icon, label, value }: InfoFieldProps) {
  return (
    <div className="min-w-0">
      <div className="flex items-center gap-1.5 text-xs text-muted-foreground mb-0.5">
        <span className="text-muted-foreground/70">{icon}</span>
        <span>{label}</span>
      </div>
      <div className="font-medium truncate">{value}</div>
    </div>
  );
}
