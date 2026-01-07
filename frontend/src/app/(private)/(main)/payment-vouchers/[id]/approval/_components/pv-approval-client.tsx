"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { toast } from "sonner";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { DollarSign, FileText, TrendingUp, ArrowLeft } from "lucide-react";
import { Skeleton } from "@/components/ui/skeleton";
import { PageHeader } from "@/components/base/page-header";
import { ApprovalActionPanel } from "@/components/workflows/approval-action-panel";
import type { ApprovalTask } from "@/types";

interface PVApprovalClientProps {
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
  approvalStage?: number;        // Add approvalStage field
  paymentDueDate?: Date;         // Add paymentDueDate field
  stageName: string;
  expenses: Array<{
    id: string;
    description: string;
    amount: number;
    category: string;
    glCode: string;
  }>;
  createdAt: string | Date;
  updatedAt: string | Date;
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

// Convert PV to ApprovalTask format
function convertPVToApprovalTask(
  pv: PaymentVoucher,
  userId: string
): ApprovalTask {
  return {
    id: pv.id,
    organizationId: "default-org", // Should come from context
    documentId: pv.id,
    documentType: "payment_voucher",
    approverId: userId,
    status: "pending",
    stage: pv.approvalStage || 1,
    createdAt: typeof pv.createdAt === 'string' ? new Date(pv.createdAt) : pv.createdAt,
    updatedAt: typeof pv.updatedAt === 'string' ? new Date(pv.updatedAt) : pv.updatedAt,
    
    // Legacy compatibility fields
    entityId: pv.id,
    entityType: "PAYMENT_VOUCHER",
    entityNumber: pv.voucherNumber,
    stageName: "Payment Approval",
    stageIndex: pv.approvalStage || 1,
    importance: pv.amount > 10000 ? "high" : "medium",
    approverName: "Current Approver",
    approverUserId: userId,
    actionDate: new Date(),
    dueDate: pv.paymentDueDate || new Date(),
    workflowId: "pv-workflow-v1",
    workflowName: "3-Stage Payment Voucher Approval",
  };
}

export function PVApprovalClient({
  pvId,
  userId,
  userRole,
}: PVApprovalClientProps) {
  const router = useRouter();
  const [pv, setPV] = useState<PaymentVoucher | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    // Simulate data loading
    const timer = setTimeout(() => {
      setPV(generateMockPV(pvId));
      setIsLoading(false);
    }, 500);

    return () => clearTimeout(timer);
  }, [pvId]);

  const handleBack = () => {
    router.back();
  };

  if (isLoading || !pv) {
    return (
      <div className="space-y-6">
        <div className="space-y-4">
          <Skeleton className="h-12 w-48" />
          <Skeleton className="h-96 w-full" />
          <Skeleton className="h-96 w-full" />
        </div>
      </div>
    );
  }

  const approvalTask = convertPVToApprovalTask(pv, userId);

  return (
    <div className="space-y-6">
      <PageHeader
        title={pv.voucherNumber}
        subtitle="Payment Voucher Approval"
        onBackClick={handleBack}
        showBackButton={true}
      />

      <div className="grid gap-6 md:grid-cols-3">
        {/* Main Content */}
        <div className="md:col-span-2 space-y-6">
          {/* Invoice and Vendor Information */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2 text-lg">
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
                <p className="font-semibold text-sm">{pv.description}</p>
              </div>
            </CardContent>
          </Card>

          {/* Payment Method Details */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2 text-lg">
                <DollarSign className="h-5 w-5" />
                Payment Method
              </CardTitle>
            </CardHeader>
            <CardContent className="grid gap-4 md:grid-cols-2">
              <div>
                <p className="text-sm text-muted-foreground">Payment Method</p>
                <p className="font-semibold">
                  {PAYMENT_METHODS[pv.paymentMethod]}
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

          {/* GL Code and Cost Center */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2 text-lg">
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
                      <th className="text-left font-semibold py-3 px-4">
                        GL Code
                      </th>
                      <th className="text-right font-semibold py-3 px-4">
                        Amount
                      </th>
                    </tr>
                  </thead>
                  <tbody>
                    {pv.expenses.map((expense) => (
                      <tr
                        key={expense.id}
                        className="border-b hover:bg-muted/30"
                      >
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
        </div>

        {/* Approval Panel */}
        <div>
          <ApprovalActionPanel
            task={approvalTask}
            onApprovalComplete={() => {
              toast.success("Payment voucher approved successfully");
              router.push("/payment-vouchers");
            }}
          />
        </div>
      </div>
    </div>
  );
}
