"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { toast } from "sonner";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import {
  ArrowLeft,
  TrendingUp,
  FileText,
  DollarSign,
  Download,
  Eye,
} from "lucide-react";
import { Skeleton } from "@/components/ui/skeleton";
import { PageHeader } from "@/components/base/page-header";
import { PDFPreviewDialog } from "@/components/modals/pdf-preview-dialog";
import {
  exportPaymentVoucherPDF,
  getPaymentVoucherPDFBlob,
} from "@/lib/pdf/pdf-export";

interface PVDetailClientProps {
  pvId: string;
  userId: string;
  userRole: string;
}

interface PaymentVoucher {
  id: string;
  voucherNumber: string;
  status: "DRAFT" | "SUBMITTED" | "IN_REVIEW" | "APPROVED" | "REJECTED";
  invoiceNumber: string;
  invoiceDate: string;
  vendorName: string;
  vendorId: string;
  amount: number;
  description: string;
  paymentMethod: "CHEQUE" | "BANK_TRANSFER" | "CASH";
  bankDetails?: {
    bankName: string;
    accountNumber: string;
    accountHolder: string;
  };
  glCode: string;
  costCenter: string;
  requestedBy: string;
  requestDate: string;
  dueDate: string;
  currentStage: number;
  stageName: string;
  expenses: Array<{
    id: string;
    description: string;
    amount: number;
    category: string;
    glCode: string;
  }>;
  createdAt: string;
  updatedAt: string;
}

const STAGE_NAMES: Record<number, string> = {
  1: "Department Manager Review",
  2: "Finance Officer Review",
  3: "CFO Approval",
};

const PAYMENT_METHODS: Record<string, string> = {
  CHEQUE: "Cheque",
  BANK_TRANSFER: "Bank Transfer",
  CASH: "Cash",
};

// Mock data generator
function generateMockPV(pvId: string): PaymentVoucher {
  const paymentMethod = ["CHEQUE", "BANK_TRANSFER", "CASH"][
    Math.floor(Math.random() * 3)
  ] as "CHEQUE" | "BANK_TRANSFER" | "CASH";
  const currentStage = Math.floor(Math.random() * 3) + 1;

  return {
    id: pvId,
    voucherNumber: `PV-2024-${String(Math.floor(Math.random() * 9000) + 1000).padStart(4, "0")}`,
    status: "IN_REVIEW",
    invoiceNumber: `INV-${Math.random().toString(36).substring(7).toUpperCase()}`,
    invoiceDate: new Date(Date.now() - 10 * 24 * 60 * 60 * 1000).toISOString(),
    vendorName: "Office Supplies Ltd.",
    vendorId: "VENDOR-001",
    amount: 15500,
    description: "Office supplies and equipment procurement",
    paymentMethod,
    bankDetails:
      paymentMethod === "BANK_TRANSFER"
        ? {
            bankName: "First National Bank",
            accountNumber: "1234567890",
            accountHolder: "Office Supplies Ltd.",
          }
        : undefined,
    glCode: "5100",
    costCenter: "CC-002",
    requestedBy: "REQ-USER-002",
    requestDate: new Date(Date.now() - 5 * 24 * 60 * 60 * 1000).toISOString(),
    dueDate: new Date(Date.now() + 10 * 24 * 60 * 60 * 1000).toISOString(),
    currentStage,
    stageName: STAGE_NAMES[currentStage],
    expenses: [
      {
        id: "exp-1",
        description: "Printer paper and cartridges",
        amount: 5500,
        category: "Supplies",
        glCode: "5100",
      },
      {
        id: "exp-2",
        description: "Desk organizers and filing systems",
        amount: 4200,
        category: "Office Equipment",
        glCode: "5100",
      },
      {
        id: "exp-3",
        description: "Cleaning and maintenance supplies",
        amount: 3500,
        category: "Facilities",
        glCode: "5200",
      },
      {
        id: "exp-4",
        description: "Miscellaneous office items",
        amount: 2300,
        category: "Supplies",
        glCode: "5100",
      },
    ],
    createdAt: new Date(Date.now() - 5 * 24 * 60 * 60 * 1000).toISOString(),
    updatedAt: new Date(Date.now() - 2 * 24 * 60 * 60 * 1000).toISOString(),
  };
}

export function PVDetailClient({
  pvId,
  userId,
  userRole,
}: PVDetailClientProps) {
  const router = useRouter();
  const [pv, setPV] = useState<PaymentVoucher | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [isExporting, setIsExporting] = useState(false);
  const [previewOpen, setPreviewOpen] = useState(false);
  const [previewBlob, setPreviewBlob] = useState<Blob | null>(null);

  useEffect(() => {
    // Simulate data loading
    const timer = setTimeout(() => {
      setPV(generateMockPV(pvId));
      setIsLoading(false);
    }, 500);

    return () => clearTimeout(timer);
  }, [pvId]);

  const handleExportPDF = async () => {
    if (!pv) return;
    try {
      setIsExporting(true);
      // Convert mock PV to PaymentVoucher type from types
      await exportPaymentVoucherPDF(pv as any);
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
      const blob = await getPaymentVoucherPDFBlob(pv as any);
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
          title={pv.voucherNumber}
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
        >
          <Eye className="h-4 w-4" />
          {isExporting ? "Loading..." : "Preview"}
        </Button>
        <Button
          onClick={handleExportPDF}
          disabled={isExporting}
          variant="outline"
          className="gap-2 h-11 mt-2"
        >
          <Download className="h-4 w-4" />
          {isExporting ? "Exporting..." : "Export PDF"}
        </Button>
      </div>

      {/* Status and Stage Info */}
      <div className="grid gap-4 md:grid-cols-2">
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium">Current Stage</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-lg font-semibold">{pv.stageName}</div>
            <p className="text-xs text-muted-foreground mt-1">
              Stage {pv.currentStage} of 3
            </p>
            <div className="flex gap-1 mt-3">
              {[1, 2, 3].map((stage) => (
                <div
                  key={stage}
                  className={`h-2 flex-1 rounded-full ${
                    stage <= pv.currentStage ? "bg-blue-600" : "bg-gray-200"
                  }`}
                />
              ))}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium">Total Amount</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-green-600">
              K{pv.amount.toLocaleString("en-ZM")}
            </div>
            <p className="text-xs text-muted-foreground mt-1">
              {pv.expenses.length} expense items
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Invoice and Payment Details */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <FileText className="h-5 w-5" />
            Invoice Information
          </CardTitle>
        </CardHeader>
        <CardContent className="grid gap-4 md:grid-cols-2">
          <div>
            <p className="text-sm text-muted-foreground">Invoice Number</p>
            <p className="font-semibold">{pv.invoiceNumber}</p>
          </div>
          <div>
            <p className="text-sm text-muted-foreground">Invoice Date</p>
            <p className="font-semibold">
              {new Date(pv.invoiceDate).toLocaleDateString()}
            </p>
          </div>
          <div>
            <p className="text-sm text-muted-foreground">Vendor Name</p>
            <p className="font-semibold">{pv.vendorName}</p>
          </div>
          <div>
            <p className="text-sm text-muted-foreground">Description</p>
            <p className="font-semibold">{pv.description}</p>
          </div>
        </CardContent>
      </Card>

      {/* Payment Method Details */}
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
            <p className="font-semibold">{PAYMENT_METHODS[pv.paymentMethod]}</p>
          </div>
          {pv.bankDetails && (
            <>
              <div>
                <p className="text-sm text-muted-foreground">Bank Name</p>
                <p className="font-semibold">{pv.bankDetails.bankName}</p>
              </div>
              <div>
                <p className="text-sm text-muted-foreground">Account Holder</p>
                <p className="font-semibold">{pv.bankDetails.accountHolder}</p>
              </div>
              <div>
                <p className="text-sm text-muted-foreground">Account Number</p>
                <p className="font-semibold font-mono text-sm">
                  {pv.bankDetails.accountNumber}
                </p>
              </div>
            </>
          )}
        </CardContent>
      </Card>

      {/* GL Code and Cost Center */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <TrendingUp className="h-5 w-5" />
            Accounting Details
          </CardTitle>
        </CardHeader>
        <CardContent className="grid gap-4 md:grid-cols-2">
          <div>
            <p className="text-sm text-muted-foreground">GL Code</p>
            <p className="font-semibold font-mono">{pv.glCode}</p>
          </div>
          <div>
            <p className="text-sm text-muted-foreground">Cost Center</p>
            <p className="font-semibold font-mono">{pv.costCenter}</p>
          </div>
          <div>
            <p className="text-sm text-muted-foreground">Request Date</p>
            <p className="font-semibold">
              {new Date(pv.requestDate).toLocaleDateString()}
            </p>
          </div>
          <div>
            <p className="text-sm text-muted-foreground">Due Date</p>
            <p className="font-semibold">
              {new Date(pv.dueDate).toLocaleDateString()}
            </p>
          </div>
        </CardContent>
      </Card>

      {/* Expense Details */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Expense Items</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead className="border-b bg-muted/50">
                <tr>
                  <th className="text-left font-semibold py-3 px-4">
                    Description
                  </th>
                  <th className="text-left font-semibold py-3 px-4">
                    Category
                  </th>
                  <th className="text-left font-semibold py-3 px-4">GL Code</th>
                  <th className="text-right font-semibold py-3 px-4">Amount</th>
                </tr>
              </thead>
              <tbody>
                {pv.expenses.map((expense) => (
                  <tr key={expense.id} className="border-b hover:bg-muted/30">
                    <td className="py-3 px-4 font-medium">
                      {expense.description}
                    </td>
                    <td className="py-3 px-4 text-muted-foreground">
                      {expense.category}
                    </td>
                    <td className="py-3 px-4 font-mono text-sm">
                      {expense.glCode}
                    </td>
                    <td className="py-3 px-4 text-right font-semibold">
                      K{expense.amount.toLocaleString("en-ZM")}
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
                    K{pv.amount.toLocaleString("en-ZM")}
                  </td>
                </tr>
              </tfoot>
            </table>
          </div>
        </CardContent>
      </Card>

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
          fileName={`PV-${pv.voucherNumber}.pdf`}
          onDownload={handleExportPDF}
        />
      )}
    </div>
  );
}
