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
  Clock,
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
import { WorkflowDocument } from "@/types/workflow";
import { usePurchaseOrdersAsWorkflowDocuments } from "@/hooks/use-purchase-order-storage";

interface PurchaseOrdersTableProps {
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

// Columns definition
function getColumns(
  onViewClick: (id: string) => void
): ColumnDef<WorkflowDocument>[] {
  return [
    {
      id: "documentNumber",
      accessorKey: "documentNumber",
      header: ({ column }) => (
        <Button
          variant="ghost"
          onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
          className="h-8 p-0"
        >
          PO Number
          <ArrowUpDown className="ml-2 h-4 w-4" />
        </Button>
      ),
      cell: ({ row }) => (
        <div className="font-medium">{row.getValue("documentNumber")}</div>
      ),
    },
    {
      id: "vendor",
      accessorKey: "metadata.vendorName",
      header: "Vendor",
      cell: ({ row }) => <div>{row.original.metadata?.vendorName || "-"}</div>,
    },
    {
      id: "amount",
      accessorKey: "metadata.totalAmount",
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
          K {(row.original.metadata?.totalAmount || 0).toLocaleString()}
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
      accessorKey: "currentStage",
      header: "Stage",
      cell: ({ row }) => (
        <StageIndicator
          currentStage={row.original.currentStage || 1}
          totalStages={4}
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

export function PurchaseOrdersTable({
  userId,
  userRole,
  refreshTrigger,
  onRefresh,
}: PurchaseOrdersTableProps) {
  const router = useRouter();
  const { data: purchaseOrders, isLoading } =
    usePurchaseOrdersAsWorkflowDocuments(true);
  const [data, setData] = useState<WorkflowDocument[]>([]);

  useEffect(() => {
    if (purchaseOrders && purchaseOrders.length > 0) {
      // Filter by current user's purchase orders
      const userPOs = purchaseOrders.filter((po) => po.createdBy === userId);
      setData(userPOs);
    } else {
      setData([]);
    }
  }, [purchaseOrders, userId, refreshTrigger]);

  const handleViewClick = (id: string) => {
    router.push(`/purchase-orders/${id}`);
  };

  const columns = getColumns(handleViewClick);

  return (
    <div className="space-y-4">
      <DataTable columns={columns} data={data} />
    </div>
  );
}
