"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { toast } from "sonner";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { ArrowLeft, Building2, TrendingUp, Download, Eye } from "lucide-react";
import { Skeleton } from "@/components/ui/skeleton";
import { PageHeader } from "@/components/base/page-header";
import { POItemsTable } from "./po-items-table";
import { PDFPreviewDialog } from "@/components/modals/pdf-preview-dialog";
import {
  exportPurchaseOrderPDF,
  getPurchaseOrderPDFBlob,
} from "@/lib/pdf/pdf-export";
import { useOrganizationContext } from "@/hooks/use-organization";
import { usePurchaseOrderById } from "@/hooks/use-purchase-order-queries";

interface PODetailClientProps {
  poId: string;
  userId: string;
  userRole: string;
}

export function PODetailClient({ poId }: PODetailClientProps) {
  const router = useRouter();
  const { currentOrganization } = useOrganizationContext();

  // Fetch real data from backend
  const { data: po, isLoading, refetch } = usePurchaseOrderById(poId);

  const [isExporting, setIsExporting] = useState(false);
  const [previewOpen, setPreviewOpen] = useState(false);
  const [previewBlob, setPreviewBlob] = useState<Blob | null>(null);

  const handleExportPDF = async () => {
    if (!po) return;
    try {
      setIsExporting(true);
      // Refetch latest data before export
      const { data: freshPO } = await refetch();

      if (!freshPO) {
        toast.error("Failed to fetch latest data");
        return;
      }

      await exportPurchaseOrderPDF(freshPO, {
        logoUrl: currentOrganization?.logoUrl,
        orgName: currentOrganization?.name,
        tagline: currentOrganization?.tagline,
      });
      toast.success("Purchase Order exported as PDF");
    } catch (error) {
      console.error("PDF export error:", error);
      toast.error("Failed to export PDF");
    } finally {
      setIsExporting(false);
    }
  };

  const handlePreviewPDF = async () => {
    if (!po) return;
    try {
      setIsExporting(true);
      // Refetch latest data before preview
      const { data: freshPO } = await refetch();

      if (!freshPO) {
        toast.error("Failed to fetch latest data");
        return;
      }

      const blob = await getPurchaseOrderPDFBlob(freshPO, {
        logoUrl: currentOrganization?.logoUrl,
        orgName: currentOrganization?.name,
        tagline: currentOrganization?.tagline,
      });
      setPreviewBlob(blob);
      setPreviewOpen(true);
    } catch (error) {
      console.error("PDF preview error:", error);
      toast.error("Failed to generate PDF preview");
    } finally {
      setIsExporting(false);
    }
  };

  const handleApprove = () => {
    toast.success("Navigating to approval...");
    router.push(`/purchase-orders/${poId}/approval`);
  };

  const handleBack = () => {
    router.back();
  };

  if (isLoading || !po) {
    return (
      <div className="space-y-6">
        <Button variant="ghost" size="sm" onClick={handleBack}>
          <ArrowLeft className="h-4 w-4 mr-2" />
          Back
        </Button>
        <div className="space-y-4">
          <Skeleton className="h-12 w-48" />
          <Skeleton className="h-96 w-full" />
          <Skeleton className="h-96 w-full" />
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-start justify-between gap-4">
        <PageHeader
          title={po.documentNumber}
          subtitle="Purchase Order Details"
          badges={[
            {
              status: po.status,
              type: "document",
            },
          ]}
          onBackClick={handleBack}
          showBackButton={true}
        />
        <Button
          onClick={handlePreviewPDF}
          disabled={isExporting}
          variant="outline"
          className="gap-2 h-11 mt-2 mr-2"
          isLoading={isExporting}
          loadingText="Loading..."
        >
          <Eye className="h-4 w-4" />
          Preview
        </Button>
        <Button
          onClick={handleExportPDF}
          disabled={isExporting}
          variant="outline"
          className="gap-2 h-11 mt-2"
          isLoading={isExporting}
          loadingText="Exporting..."
        >
          <Download className="h-4 w-4" />
          Export PDF
        </Button>
      </div>

      {/* Vendor Information */}
      {(po.vendor || po.vendorName) && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Building2 className="h-5 w-5" />
              Vendor Information
            </CardTitle>
          </CardHeader>
          <CardContent className="grid gap-4 md:grid-cols-2">
            <div>
              <p className="text-sm text-muted-foreground">Vendor Name</p>
              <p className="font-semibold">
                {po.vendor?.name || po.vendorName}
              </p>
            </div>
            {po.vendor?.email && (
              <div>
                <p className="text-sm text-muted-foreground">Email</p>
                <p className="font-semibold text-blue-600">{po.vendor.email}</p>
              </div>
            )}
            {po.vendor?.phone && (
              <div>
                <p className="text-sm text-muted-foreground">Phone</p>
                <p className="font-semibold">{po.vendor.phone}</p>
              </div>
            )}
            {po.vendor?.city && po.vendor?.country && (
              <div className="md:col-span-2">
                <p className="text-sm text-muted-foreground">Location</p>
                <p className="font-semibold">
                  {po.vendor.city}, {po.vendor.country}
                </p>
              </div>
            )}
          </CardContent>
        </Card>
      )}

      {/* PO Details and Status */}
      <div className="grid gap-4 md:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <TrendingUp className="h-5 w-5" />
              Order Details
            </CardTitle>
          </CardHeader>
          <CardContent className="grid gap-4 md:grid-cols-2">
            <div>
              <p className="text-sm text-muted-foreground">Created Date</p>
              <p className="font-semibold">
                {new Date(po.createdAt).toLocaleDateString()}
              </p>
            </div>
            <div>
              <p className="text-sm text-muted-foreground">Delivery Date</p>
              <p className="font-semibold">
                {new Date(po.deliveryDate).toLocaleDateString()}
              </p>
            </div>
            {po.department && (
              <div>
                <p className="text-sm text-muted-foreground">Department</p>
                <p className="font-semibold">{po.department}</p>
              </div>
            )}
            {po.priority && (
              <div>
                <p className="text-sm text-muted-foreground">Priority</p>
                <p className="font-semibold capitalize">{po.priority}</p>
              </div>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium">Total Amount</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-green-600">
              K{po.totalAmount?.toLocaleString("en-ZM") || "0"}
            </div>
            <p className="text-xs text-muted-foreground mt-1">
              {po.items?.length || 0} items
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Line Items */}
      {po.items && po.items.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle>Line Items</CardTitle>
          </CardHeader>
          <CardContent>
            <POItemsTable
              items={po.items.map((item, index) => ({
                ...item,
                id: item.id || `item-${index}`,
                itemNumber: index + 1,
                totalPrice: item.amount || item.totalPrice || 0,
                unit: item.unit || "unit",
              }))}
            />
          </CardContent>
        </Card>
      )}

      {/* Cost Summary */}
      <Card>
        <CardHeader>
          <CardTitle>Cost Summary</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-2 max-w-xs ml-auto">
            <div className="flex justify-between text-sm">
              <span className="text-muted-foreground">Subtotal:</span>
              <span className="font-semibold">
                K{(po.subtotal || po.totalAmount || 0).toLocaleString("en-ZM")}
              </span>
            </div>
            {po.tax !== undefined && (
              <div className="flex justify-between text-sm">
                <span className="text-muted-foreground">Tax:</span>
                <span className="font-semibold">
                  K{po.tax.toLocaleString("en-ZM")}
                </span>
              </div>
            )}
            <div className="border-t pt-2 flex justify-between">
              <span className="font-semibold">Total:</span>
              <span className="text-lg font-bold text-green-600">
                K{(po.totalAmount || 0).toLocaleString("en-ZM")}
              </span>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Action Buttons */}
      <div className="flex gap-4 pt-4">
        <Button variant="outline" onClick={handleBack}>
          Cancel
        </Button>
        {po.status === "pending" && (
          <Button
            onClick={handleApprove}
            className="bg-blue-600 hover:bg-blue-700"
          >
            Review & Approve
          </Button>
        )}
      </div>

      {/* PDF Preview Dialog */}
      {previewBlob && (
        <PDFPreviewDialog
          open={previewOpen}
          onOpenChange={setPreviewOpen}
          pdfBlob={previewBlob}
          fileName={`Purchase Order: ${po.documentNumber}`}
          onDownload={handleExportPDF}
        />
      )}
    </div>
  );
}
