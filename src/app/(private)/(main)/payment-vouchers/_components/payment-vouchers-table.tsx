"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { ColumnDef } from "@tanstack/react-table";
import {
  ArrowUpDown,
  MoreHorizontal,
  Download,
  Eye,
  CheckCircle2,
  XCircle,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { DataTable } from "@/components/ui/data-table";
import { StatusBadge as CentralizedStatusBadge } from "@/components/status-badge";
import { usePaymentVouchersAsWorkflowDocuments } from "@/hooks/use-payment-voucher-storage";

interface PaymentVouchersTableProps {
  userId: string;
  userRole: string;
  refreshTrigger: number;
  onRefresh: () => void;
}

// Stage indicator
function StageIndicator({
  currentStage,
  totalStages,
}: {
  currentStage: number;
  totalStages: number;
}) {
  return (
    <div className="flex items-center gap-1">
      <span className="text-sm font-medium">{currentStage}</span>
      <span className="text-xs text-muted-foreground">of {totalStages}</span>
    </div>
  );
}

// PV Document type for table display
interface PVDocumentRow {
  id: string;
  documentNumber: string;
  title: string;
  status: string;
  priority: string;
  vendorName: string;
  department: string;
  totalAmount: number;
  currency: string;
  createdAt: Date;
  createdBy: string;
  submittedAt?: Date;
  approvedAt?: Date;
  currentApprovalStage?: number;
  totalApprovalStages?: number;
}

// Columns definition
function getColumns(
  onViewClick: (id: string) => void
): ColumnDef<PVDocumentRow>[] {
  return [
    {
      id: "voucherNumber",
      accessorKey: "documentNumber",
      header: ({ column }) => (
        <Button
          variant="ghost"
          onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
          className="h-8 p-0"
        >
          Voucher No.
          <ArrowUpDown className="ml-2 h-4 w-4" />
        </Button>
      ),
      cell: ({ row }) => (
        <div className="font-medium">{row.getValue("voucherNumber")}</div>
      ),
    },
    {
      id: "vendor",
      accessorKey: "vendorName",
      header: "Vendor",
      cell: ({ row }) => <div>{row.original.vendorName || "-"}</div>,
    },
    {
      id: "amount",
      accessorKey: "totalAmount",
      header: ({ column }) => (
        <Button
          variant="ghost"
          onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
          className="h-8 p-0"
        >
          Amount
          <ArrowUpDown className="ml-2 h-4 w-4" />
        </Button>
      ),
      cell: ({ row }) => (
        <div className="font-medium">
          {row.original.currency} {(row.original.totalAmount || 0).toLocaleString()}
        </div>
      ),
    },
    {
      id: "status",
      accessorKey: "status",
      header: "Status",
      cell: ({ row }) => (
        <CentralizedStatusBadge
          status={row.getValue("status")}
          type="document"
        />
      ),
    },
    {
      id: "stage",
      accessorKey: "currentApprovalStage",
      header: "Stage",
      cell: ({ row }) => (
        <StageIndicator
          currentStage={row.original.currentApprovalStage || 1}
          totalStages={row.original.totalApprovalStages || 3}
        />
      ),
    },
    {
      id: "createdDate",
      accessorKey: "createdAt",
      header: ({ column }) => (
        <Button
          variant="ghost"
          onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
          className="h-8 p-0"
        >
          Created
          <ArrowUpDown className="ml-2 h-4 w-4" />
        </Button>
      ),
      cell: ({ row }) => (
        <div className="text-sm">
          {new Date(row.original.createdAt).toLocaleDateString()}
        </div>
      ),
    },
    {
      id: "actions",
      header: "Actions",
      cell: ({ row }) => (
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" className="h-8 w-8 p-0">
              <span className="sr-only">Open menu</span>
              <MoreHorizontal className="h-4 w-4" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuLabel>Actions</DropdownMenuLabel>
            <DropdownMenuItem
              onClick={() => onViewClick(row.original.id)}
              className="flex items-center gap-2"
            >
              <Eye className="h-4 w-4" />
              View Details
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem className="flex items-center gap-2">
              <Download className="h-4 w-4" />
              Download PDF
            </DropdownMenuItem>
            {row.original.status === "IN_REVIEW" && (
              <>
                <DropdownMenuSeparator />
                <DropdownMenuItem className="flex items-center gap-2 text-green-600">
                  <CheckCircle2 className="h-4 w-4" />
                  Approve
                </DropdownMenuItem>
                <DropdownMenuItem className="flex items-center gap-2 text-red-600">
                  <XCircle className="h-4 w-4" />
                  Reject
                </DropdownMenuItem>
              </>
            )}
          </DropdownMenuContent>
        </DropdownMenu>
      ),
    },
  ];
}

export function PaymentVouchersTable({
  userId,
  userRole,
  refreshTrigger,
  onRefresh,
}: PaymentVouchersTableProps) {
  const router = useRouter();
  const { data: paymentVouchers, isLoading } =
    usePaymentVouchersAsWorkflowDocuments(true);
  const [data, setData] = useState<PVDocumentRow[]>([]);

  useEffect(() => {
    if (paymentVouchers && paymentVouchers.length > 0) {
      // Filter by current user's payment vouchers
      const userPVs = paymentVouchers.filter((pv) => pv.createdBy === userId);
      setData(userPVs);
    } else {
      setData([]);
    }
  }, [paymentVouchers, userId, refreshTrigger]);

  const handleViewClick = (id: string) => {
    router.push(`/payment-vouchers/${id}`);
  };

  const columns = getColumns(handleViewClick);

  return (
    <div className="space-y-4">
      <DataTable columns={columns} data={data} />
    </div>
  );
}
