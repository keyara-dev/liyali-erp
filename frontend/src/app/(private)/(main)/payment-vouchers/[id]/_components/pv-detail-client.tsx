"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { toast } from "sonner";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { ArrowLeft, FileText, DollarSign, Download, Eye } from "lucide-react";
import { Skeleton } from "@/components/ui/skeleton";
import { PageHeader } from "@/components/base/page-header";
import { PDFPreviewDialog } from "@/components/modals/pdf-preview-dialog";
import {
  exportPaymentVoucherPDF,
  getPaymentVoucherPDFBlob,
} from "@/lib/pdf/pdf-export";
import { useOrganizationContext } from "@/hooks/use-organization";
import { usePaymentVoucherById } from "@/hooks/use-payment-voucher-queries";

interface PVDetailClientProps {
  pvId: string;
  userId: string;
  userRole: string;
}

const PAYMENT_METHODS: Record<string, string> = {
  CHEQUE: "Cheque",
  BANK_TRANSFER: "Bank Transfer",
  CASH: "Cash",
};

export function PVDetailClient({ pvId }: PVDetailClientProps) {
  const router = useRouter();
  const { currentOrganization } = useOrganizationContext();

  // Fetch real data from backend
  const { data: pv, isLoading, refetch } = usePaymentVoucherById(pvId);

  const [isExporting, setIsExporting] = useState(false);
  const [previewOpen, setPreviewOpen] = useState(false);
  const [previewBlob, setPreviewBlob] = useState<Blob | null>(null);

  const handleExportPDF = async () => {
    if (!pv) return;
    try {
      setIsExporting(true);
      // Refetch latest data before export
      const { data: freshPV } = await refetch();

      if (!freshPV) {
        toast.error("Failed to fetch latest data");
        return;
      }

      await exportPaymentVoucherPDF(freshPV, {
        logoUrl: currentOrganization?.logoUrl,
        orgName: currentOrganization?.name,
        tagline: currentOrganization?.tagline,
      });
      toast.success("Payment Voucher exported as PDF");
    } catch (error) {
      console.error("PDF export error:", error);
      toast.error("Failed to export PDF");
    } finally {
      setIsExporting(false);
    }
  };

  const handlePreviewPDF = async () => {
    if (!pv) return;
    try {
      setIsExporting(true);
      // Refetch latest data before preview
      const { data: freshPV } = await refetch();

      if (!freshPV) {
        toast.error("Failed to fetch latest data");
        return;
      }

      const blob = await getPaymentVoucherPDFBlob(freshPV, {
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
    router.push(`/payment-vouchers/${pvId}/approval`);
  };

  const handleBack = () => {
    router.back();
  };

  if (isLoading || !pv) {
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
          title={pv.documentNumber}
          subtitle="Payment Voucher Details"
          badges={[
            {
              status: pv.status,
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

      {/* Status and Total Amount */}
      <div className="grid gap-4 md:grid-cols-2">
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium">
              Document Status
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-lg font-semibold">{pv.status}</div>
            <p className="text-xs text-muted-foreground mt-1">
              Current workflow status
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium">Total Amount</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-green-600">
              K{(pv.amount || pv.totalAmount || 0).toLocaleString("en-ZM")}
            </div>
            <p className="text-xs text-muted-foreground mt-1">
              Payment voucher total
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Payment Voucher Details */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <FileText className="h-5 w-5" />
            Payment Voucher Information
          </CardTitle>
        </CardHeader>
        <CardContent className="grid gap-4 md:grid-cols-2">
          {pv.vendorName && (
            <div>
              <p className="text-sm text-muted-foreground">Vendor Name</p>
              <p className="font-semibold">{pv.vendorName}</p>
            </div>
          )}
          {pv.invoiceNumber && (
            <div>
              <p className="text-sm text-muted-foreground">Invoice Number</p>
              <p className="font-semibold">{pv.invoiceNumber}</p>
            </div>
          )}
          {pv.description && (
            <div className="md:col-span-2">
              <p className="text-sm text-muted-foreground">Description</p>
              <p className="font-semibold">{pv.description}</p>
            </div>
          )}
          {pv.requestedDate && (
            <div>
              <p className="text-sm text-muted-foreground">Requested Date</p>
              <p className="font-semibold">
                {new Date(pv.requestedDate).toLocaleDateString()}
              </p>
            </div>
          )}
          {pv.paymentDueDate && (
            <div>
              <p className="text-sm text-muted-foreground">Payment Due Date</p>
              <p className="font-semibold">
                {new Date(pv.paymentDueDate).toLocaleDateString()}
              </p>
            </div>
          )}
          <div>
            <p className="text-sm text-muted-foreground">Created Date</p>
            <p className="font-semibold">
              {new Date(pv.createdAt).toLocaleDateString()}
            </p>
          </div>
        </CardContent>
      </Card>

      {/* Payment Method Details */}
      {pv.paymentMethod && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <DollarSign className="h-5 w-5" />
              Payment Method
            </CardTitle>
          </CardHeader>
          <CardContent className="grid gap-4 md:grid-cols-2">
            <div>
              <p className="text-sm text-muted-foreground">Payment Method</p>
              <p className="font-semibold">
                {PAYMENT_METHODS[pv.paymentMethod] || pv.paymentMethod}
              </p>
            </div>
            {pv.bankDetails && (
              <>
                <div>
                  <p className="text-sm text-muted-foreground">Bank Name</p>
                  <p className="font-semibold">{pv.bankDetails.bankName}</p>
                </div>
                <div>
                  <p className="text-sm text-muted-foreground">
                    Account Holder
                  </p>
                  <p className="font-semibold">
                    {pv.bankDetails.accountHolder}
                  </p>
                </div>
                <div>
                  <p className="text-sm text-muted-foreground">
                    Account Number
                  </p>
                  <p className="font-semibold font-mono text-sm">
                    {pv.bankDetails.accountNumber}
                  </p>
                </div>
              </>
            )}
          </CardContent>
        </Card>
      )}

      {/* Items/Line Items if available */}
      {pv.items && pv.items.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle className="text-lg">Line Items</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead className="border-b bg-muted/50">
                  <tr>
                    <th className="text-left font-semibold py-3 px-4">
                      Description
                    </th>
                    <th className="text-right font-semibold py-3 px-4">
                      Quantity
                    </th>
                    <th className="text-right font-semibold py-3 px-4">
                      Unit Price
                    </th>
                    <th className="text-right font-semibold py-3 px-4">
                      Total
                    </th>
                  </tr>
                </thead>
                <tbody>
                  {pv.items.map((item: any, index: number) => (
                    <tr
                      key={item.id || index}
                      className="border-b hover:bg-muted/30"
                    >
                      <td className="py-3 px-4 font-medium">
                        {item.description || item.itemDescription}
                      </td>
                      <td className="py-3 px-4 text-right">{item.quantity}</td>
                      <td className="py-3 px-4 text-right">
                        K{(item.unitPrice || 0).toLocaleString("en-ZM")}
                      </td>
                      <td className="py-3 px-4 text-right font-semibold">
                        K
                        {(
                          item.totalPrice ||
                          item.quantity * item.unitPrice ||
                          0
                        ).toLocaleString("en-ZM")}
                      </td>
                    </tr>
                  ))}
                </tbody>
                <tfoot className="border-t bg-muted/30">
                  <tr>
                    <td
                      colSpan={3}
                      className="py-3 px-4 font-semibold text-right"
                    >
                      Total:
                    </td>
                    <td className="py-3 px-4 text-right font-bold text-green-600">
                      K
                      {(pv.amount || pv.totalAmount || 0).toLocaleString(
                        "en-ZM",
                      )}
                    </td>
                  </tr>
                </tfoot>
              </table>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Action Buttons */}
      <div className="flex gap-4 pt-4">
        <Button variant="outline" onClick={handleBack}>
          Cancel
        </Button>
        {pv.status === "IN_REVIEW" && (
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
          fileName={`Payment Voucher: ${pv.documentNumber}`}
          onDownload={handleExportPDF}
        />
      )}
    </div>
  );
}
