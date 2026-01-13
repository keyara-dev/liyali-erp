"use client";

import { useState, useEffect } from "react";
import { Download, ChevronLeft, CheckCircle2, XCircle } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { WorkflowDocument } from "@/types/workflow";
import { ApprovalConfirmationDialog } from "@/components/modals/approval-confirmation-dialog";
import { ApprovalHistory } from "@/components/approval-history";
import { DocumentLinks } from "@/components/document-links";
import { generateGrnPDF } from "@/lib/pdf-generators/grn-pdf";
import Link from "next/link";

interface GrnDetailProps {
  grnId: string;
  userId: string;
  userRole: string;
}

export function GrnDetail({ grnId, userId, userRole }: GrnDetailProps) {
  const [grn, setGrn] = useState<WorkflowDocument | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [showApprovalDialog, setShowApprovalDialog] = useState(false);
  const [isDownloadingPDF, setIsDownloadingPDF] = useState(false);

  useEffect(() => {
    loadGrn();
  }, [grnId]);

  const loadGrn = async () => {
    setIsLoading(true);
    try {
      // Mock data - will be replaced with API call
      const mockGRN: WorkflowDocument = {
        id: grnId,
        type: "GOODS_RECEIVED_NOTE",
        documentNumber: "GRN-2024-001",
        status: "IN_REVIEW",
        currentStage: 1,
        createdBy: "user-1",
        createdAt: new Date("2024-11-27"),
        updatedAt: new Date("2024-11-28"),
        metadata: {
          poId: "po-1",
          poNumber: "PO-2024-001",
          requisitionId: "req-1",
          vendorName: "Broadway Ventures",
          receivedQuantity: 5,
          totalQuantity: 5,
          amount: 7500.0,
          receivedDate: "2024-11-27",
          items: [
            {
              id: "item-1",
              description: "Office Furniture",
              poQuantity: 5,
              receivedQuantity: 5,
              unitCost: 1000.0,
              totalCost: 5000.0,
            },
            {
              id: "item-2",
              description: "Installation Service",
              poQuantity: 1,
              receivedQuantity: 1,
              unitCost: 2500.0,
              totalCost: 2500.0,
            },
          ],
        },
      };
      setGrn(mockGRN);
    } catch (error) {
      console.error("Error loading GRN:", error);
    } finally {
      setIsLoading(false);
    }
  };

  const handleDownloadPDF = async () => {
    setIsDownloadingPDF(true);
    try {
      if (grn) {
        // Call PDF generator with @react-pdf/renderer
        await generateGrnPDF(grn);
      }
    } catch (error) {
      console.error("Error downloading PDF:", error);
    } finally {
      setIsDownloadingPDF(false);
    }
  };

  const handleApproveSubmit = async (data: any) => {
    try {
      console.log("Approving GRN with data:", data);
      loadGrn();
      setShowApprovalDialog(false);
    } catch (error) {
      console.error("Error approving GRN:", error);
    }
  };

  if (isLoading) {
    return <div className="text-center py-8">Loading...</div>;
  }

  if (!grn) {
    return <div className="text-center py-8 text-red-600">GRN not found</div>;
  }

  const canApprove = grn.status === "IN_REVIEW";
  const statusVariant =
    grn.status === "APPROVED"
      ? "default"
      : grn.status === "REJECTED"
        ? "destructive"
        : grn.status === "IN_REVIEW"
          ? "secondary"
          : "outline";

  return (
    <div className="space-y-6">
      {/* Header with back button */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <Link href="//grn">
            <Button variant="ghost" size="sm" className="gap-2">
              <ChevronLeft className="h-4 w-4" />
              Back to GRNs
            </Button>
          </Link>
          <div>
            <h1 className="text-2xl font-bold">{grn.documentNumber}</h1>
            <p className="text-sm text-muted-foreground">
              Received on {grn.createdAt ? new Date(grn.createdAt).toLocaleDateString() : 'Unknown'}
            </p>
          </div>
        </div>
        <div className="flex items-center gap-3">
          <Badge variant={statusVariant}>{grn.status}</Badge>
          <Button
            variant="outline"
            size="sm"
            onClick={handleDownloadPDF}
            disabled={isDownloadingPDF}
            className="gap-2"
          >
            <Download className="h-4 w-4" />
            {isDownloadingPDF ? "Generating..." : "Download PDF"}
          </Button>
        </div>
      </div>

      {/* Main content grid */}
      <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
        {/* Left column - GRN Details (2/3 width) */}
        <div className="lg:col-span-2 space-y-6">
          {/* Purchase Order Reference */}
          <Card>
            <CardHeader>
              <CardTitle>Purchase Order Reference</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="flex items-center justify-between bg-blue-50 dark:bg-blue-950/20 p-4 rounded border border-blue-200">
                <div>
                  <p className="text-sm text-muted-foreground">PO Number</p>
                  <p className="font-medium">{grn.metadata?.poNumber}</p>
                </div>
                <Link href={`//purchase-orders/${grn.metadata?.poId}`}>
                  <Button variant="outline" size="sm">
                    View Purchase Order
                  </Button>
                </Link>
              </div>
            </CardContent>
          </Card>

          {/* Goods Received Details */}
          <Card>
            <CardHeader>
              <CardTitle>Goods Received Details</CardTitle>
            </CardHeader>
            <CardContent className="grid grid-cols-2 gap-6">
              <div>
                <p className="text-sm text-muted-foreground">Vendor</p>
                <p className="font-medium">{grn.metadata?.vendorName}</p>
              </div>
              <div>
                <p className="text-sm text-muted-foreground">Received Date</p>
                <p className="font-medium">
                  {grn.metadata?.receivedDate 
                    ? new Date(grn.metadata.receivedDate).toLocaleDateString()
                    : grn.createdAt 
                      ? new Date(grn.createdAt).toLocaleDateString()
                      : 'Unknown'}
                </p>
              </div>
              <div>
                <p className="text-sm text-muted-foreground">
                  Total Quantity Received
                </p>
                <p className="font-bold text-lg">
                  {grn.metadata?.receivedQuantity}
                </p>
              </div>
              <div>
                <p className="text-sm text-muted-foreground">Total Amount</p>
                <p className="font-bold text-lg">
                  K {(grn.metadata?.amount || 0).toLocaleString()}
                </p>
              </div>
            </CardContent>
          </Card>

          {/* Document Links */}
          <DocumentLinks
            currentDocument={grn}
            linkedDocuments={{
              requisition: grn.metadata?.requisitionId
                ? { id: grn.metadata.requisitionId, documentNumber: "REQ-2024-001" }
                : undefined,
              purchaseOrder: grn.metadata?.poId
                ? { id: grn.metadata.poId, documentNumber: grn.metadata.poNumber }
                : undefined,
            }}
          />

          {/* Items Table */}
          <Card>
            <CardHeader>
              <CardTitle>Items Received</CardTitle>
              <CardDescription>
                {(grn.metadata?.items || []).length} items received
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="overflow-x-auto">
                <table className="w-full text-sm">
                  <thead>
                    <tr className="border-b">
                      <th className="text-left py-2 px-2">Description</th>
                      <th className="text-right py-2 px-2">PO Qty</th>
                      <th className="text-right py-2 px-2">Received Qty</th>
                      <th className="text-right py-2 px-2">Unit Cost</th>
                      <th className="text-right py-2 px-2">Total</th>
                    </tr>
                  </thead>
                  <tbody>
                    {(grn.metadata?.items || []).map((item: any) => (
                      <tr key={item.id} className="border-b hover:bg-muted/50">
                        <td className="py-3 px-2">{item.description}</td>
                        <td className="text-right py-3 px-2">
                          {item.poQuantity}
                        </td>
                        <td className="text-right py-3 px-2">
                          <span
                            className={
                              item.receivedQuantity === item.poQuantity
                                ? "text-green-600 font-medium"
                                : "text-orange-600 font-medium"
                            }
                          >
                            {item.receivedQuantity}
                          </span>
                        </td>
                        <td className="text-right py-3 px-2">
                          K {item.unitCost.toLocaleString()}
                        </td>
                        <td className="text-right py-3 px-2 font-medium">
                          K {item.totalCost.toLocaleString()}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                  <tfoot>
                    <tr className="font-bold bg-muted/50">
                      <td colSpan={4} className="py-3 px-2 text-right">
                        Total:
                      </td>
                      <td className="text-right py-3 px-2">
                        K {(grn.metadata?.amount || 0).toLocaleString()}
                      </td>
                    </tr>
                  </tfoot>
                </table>
              </div>
            </CardContent>
          </Card>

          {/* Approval History */}
          <Card>
            <CardHeader>
              <CardTitle>Approval History</CardTitle>
              <CardDescription>
                Stage {grn.currentStage} of {1} - Warehouse Manager
              </CardDescription>
            </CardHeader>
            <CardContent>
              <ApprovalHistory state={{} as any} />
            </CardContent>
          </Card>
        </div>

        {/* Right column - Status & Actions (1/3 width) */}
        <div className="space-y-6">
          {/* Status Card */}
          <Card>
            <CardHeader>
              <CardTitle>Status</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div>
                <p className="text-sm text-muted-foreground mb-2">
                  Current Stage
                </p>
                <div className="flex items-center justify-between bg-muted p-3 rounded">
                  <span className="font-medium">
                    Stage {grn.currentStage} of {1}
                  </span>
                  <span className="text-sm">Warehouse Manager</span>
                </div>
              </div>

              <div className="pt-4 border-t space-y-2">
                <p className="text-sm text-muted-foreground">Document Status</p>
                <Badge
                  variant={statusVariant}
                  className="w-full justify-center py-2"
                >
                  {grn.status}
                </Badge>
              </div>
            </CardContent>
          </Card>

          {/* Approval Actions */}
          {canApprove && (
            <Card className="border-green-200 bg-green-50 dark:bg-green-950/20">
              <CardHeader>
                <CardTitle className="text-green-900 dark:text-green-100">
                  Your Action Required
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-3">
                <p className="text-sm text-green-800 dark:text-green-200">
                  This GRN is waiting for your approval.
                </p>
                <div className="grid grid-cols-1 gap-2">
                  <Button
                    onClick={() => setShowApprovalDialog(true)}
                    className="gap-2 bg-green-600 hover:bg-green-700"
                  >
                    <CheckCircle2 className="h-4 w-4" />
                    Approve with Signature
                  </Button>
                  <Button variant="outline" className="gap-2">
                    <XCircle className="h-4 w-4" />
                    Reject
                  </Button>
                </div>
              </CardContent>
            </Card>
          )}

          {/* Timeline */}
          <Card>
            <CardHeader>
              <CardTitle>Timeline</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="text-sm space-y-3">
                <div className="flex gap-3">
                  <div className="w-2 h-2 rounded-full bg-green-500 mt-1 flex-shrink-0" />
                  <div>
                    <p className="font-medium">Created</p>
                    <p className="text-xs text-muted-foreground">
                      {grn.createdAt ? new Date(grn.createdAt).toLocaleString() : 'Unknown'}
                    </p>
                  </div>
                </div>
                <div className="flex gap-3">
                  <div className="w-2 h-2 rounded-full bg-blue-500 mt-1 flex-shrink-0" />
                  <div>
                    <p className="font-medium">Last Updated</p>
                    <p className="text-xs text-muted-foreground">
                      {grn.updatedAt ? new Date(grn.updatedAt).toLocaleString() : 'Unknown'}
                    </p>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>

      {/* Approval Dialog */}
      <ApprovalConfirmationDialog
        open={showApprovalDialog}
        documentId={grn.id}
        documentType="GRN"
        documentNumber={grn.documentNumber || ''}
        vendor={grn.metadata?.vendorName || ""}
        amount={`K ${(grn.metadata?.amount || 0).toLocaleString()}`}
        stageNumber={grn.currentStage || 1}
        totalStages={1 || 1}
        stageName="Warehouse Manager"
        onApprove={handleApproveSubmit}
        onCancel={() => setShowApprovalDialog(false)}
      />
    </div>
  );
}
