"use client";

import { useRouter } from "next/navigation";
import { ColumnDef } from "@tanstack/react-table";
import Link from "next/link";
import { useCallback, useMemo, useEffect } from "react";
import {
  ArrowUpDown,
  Download,
  Eye,
  Pencil,
  MoreVertical,
} from "lucide-react";

import { DataTable } from "@/components/ui/data-table";
import { StatusBadge } from "@/components/status-badge";
import { Button } from "@/components/ui/button";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import type { ActionButton } from "@/components/ui/action-buttons";
import type { GoodsReceivedNote } from "@/app/_actions/grn-actions";
import type { PurchaseOrder } from "@/types/purchase-order";
import { useGRNs } from "@/hooks/use-grn-queries";
import { usePurchaseOrders } from "@/hooks/use-purchase-order-queries";
import { formatCurrency } from "@/lib/utils";

interface GrnTableProps {
  userId: string;
  userRole: string;
  refreshTrigger: number;
  onRefresh: () => void;
}

const GRN_EDIT_ROLES = ["admin", "finance"];

/**
 * Sum(quantityReceived × unitPrice) for each GRN item, using the linked PO's
 * items (matched by description) as the price source. Returns undefined when
 * the PO can't be resolved so the cell can render a dash instead of zero.
 */
function computeReceivedAmount(
  grn: GoodsReceivedNote,
  po: PurchaseOrder | undefined,
): number | undefined {
  if (!po) return undefined;
  const priceByDesc = new Map<string, number>();
  for (const item of po.items ?? []) {
    if (item.description) priceByDesc.set(item.description, item.unitPrice ?? 0);
  }
  return (grn.items ?? []).reduce((total, item) => {
    const price = priceByDesc.get(item.description) ?? 0;
    return total + price * (item.quantityReceived ?? 0);
  }, 0);
}

function GrnOptionsMenu({
  grn,
  router,
  canModify,
}: {
  grn: GoodsReceivedNote;
  router: ReturnType<typeof useRouter>;
  canModify: boolean;
}) {
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <button className="h-8 w-8 rounded-md border border-input bg-background px-2 py-1.5 hover:bg-accent hover:text-accent-foreground">
          <MoreVertical className="h-4 w-4" />
        </button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        <DropdownMenuItem
          onClick={() => console.log("Download PDF for GRN:", grn.id)}
        >
          <Download className="mr-2 h-4 w-4" />
          Download
        </DropdownMenuItem>
        {grn.status?.toUpperCase() === "DRAFT" && canModify && (
          <DropdownMenuItem onClick={() => router.push(`/grn/${grn.id}`)}>
            <Pencil className="mr-2 h-4 w-4" />
            Edit
          </DropdownMenuItem>
        )}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

export function GrnTable({
  userId,
  userRole,
  refreshTrigger,
  onRefresh: _onRefresh,
}: GrnTableProps) {
  const router = useRouter();
  const { data: grns = [], refetch } = useGRNs(1, 50);
  const { data: purchaseOrders = [] } = usePurchaseOrders();

  useEffect(() => {
    refetch();
  }, [refreshTrigger, refetch]);

  // Map POs by documentNumber so each GRN row can resolve vendor / amount /
  // currency / PO link without an N+1 request.
  const posByDocNumber = useMemo(() => {
    const map = new Map<string, PurchaseOrder>();
    for (const po of purchaseOrders as PurchaseOrder[]) {
      if (po.documentNumber) map.set(po.documentNumber, po);
    }
    return map;
  }, [purchaseOrders]);

  const columns = useMemo<ColumnDef<GoodsReceivedNote>[]>(
    () => [
      {
        accessorKey: "documentNumber",
        header: ({ column }) => (
          <Button
            variant="ghost"
            onClick={() =>
              column.toggleSorting(column.getIsSorted() === "asc")
            }
            className="-ml-3"
          >
            GRN Number
            <ArrowUpDown className="ml-2 h-4 w-4" />
          </Button>
        ),
        cell: ({ row }) => (
          <div className="font-semibold uppercase">
            {row.original.documentNumber || row.original.id}
          </div>
        ),
      },
      {
        accessorKey: "poDocumentNumber",
        header: "PO Reference",
        cell: ({ row }) => {
          const poDoc = row.original.poDocumentNumber;
          const po = poDoc ? posByDocNumber.get(poDoc) : undefined;
          if (!poDoc) {
            return <span className="text-muted-foreground">—</span>;
          }
          return po ? (
            <Link
              href={`/purchase-orders/${po.id}`}
              className="text-blue-600 hover:underline font-mono text-sm"
            >
              {poDoc}
            </Link>
          ) : (
            <span className="font-mono text-sm text-muted-foreground">
              {poDoc}
            </span>
          );
        },
      },
      {
        id: "vendor",
        header: "Vendor",
        cell: ({ row }) => {
          const po = posByDocNumber.get(row.original.poDocumentNumber);
          const vendor = po?.vendorName?.trim();
          return (
            <Tooltip>
              <TooltipTrigger asChild>
                <div className="max-w-[200px] truncate font-medium capitalize cursor-help">
                  {vendor || "—"}
                </div>
              </TooltipTrigger>
              {vendor && (
                <TooltipContent>
                  <p className="max-w-xs">{vendor}</p>
                </TooltipContent>
              )}
            </Tooltip>
          );
        },
      },
      {
        id: "itemsCount",
        header: "Items",
        cell: ({ row }) => {
          const itemsCount = row.original.items?.length || 0;
          return (
            <div className="text-center">
              <span className="inline-flex items-center justify-center w-6 h-6 text-xs font-medium bg-foreground/5 rounded-full">
                {itemsCount}
              </span>
            </div>
          );
        },
      },
      {
        id: "amount",
        header: ({ column }) => (
          <Button
            variant="ghost"
            onClick={() =>
              column.toggleSorting(column.getIsSorted() === "asc")
            }
            className="-ml-3"
          >
            Received Value
            <ArrowUpDown className="ml-2 h-4 w-4" />
          </Button>
        ),
        accessorFn: (row) =>
          computeReceivedAmount(row, posByDocNumber.get(row.poDocumentNumber)) ??
          0,
        cell: ({ row }) => {
          const po = posByDocNumber.get(row.original.poDocumentNumber);
          const amount = computeReceivedAmount(row.original, po);
          if (amount === undefined) {
            return <span className="text-muted-foreground">—</span>;
          }
          return (
            <div className="font-medium">
              {formatCurrency(amount, po?.currency || "ZMW")}
            </div>
          );
        },
      },
      {
        accessorKey: "status",
        header: "Status",
        cell: ({ row }) => (
          <StatusBadge
            status={row.original.status || "DRAFT"}
            type="document"
          />
        ),
      },
      {
        accessorKey: "receivedDate",
        header: ({ column }) => (
          <Button
            variant="ghost"
            onClick={() =>
              column.toggleSorting(column.getIsSorted() === "asc")
            }
            className="-ml-3"
          >
            Received Date
            <ArrowUpDown className="ml-2 h-4 w-4" />
          </Button>
        ),
        cell: ({ row }) => {
          const date = row.original.receivedDate || row.original.createdAt;
          return (
            <div className="text-sm text-muted-foreground">
              {date ? new Date(date).toLocaleDateString() : "—"}
            </div>
          );
        },
      },
    ],
    [posByDocNumber],
  );

  const getActions = useCallback(
    (grn: GoodsReceivedNote): ActionButton[] => {
      const canModify =
        grn.createdBy === userId ||
        grn.receivedBy === userId ||
        GRN_EDIT_ROLES.includes(userRole);
      return [
        {
          icon: <Eye className="h-3.5 w-3.5" />,
          label: "View",
          tooltip: "View Details",
          onClick: () => router.push(`/grn/${grn.id}`),
        },
        ...(grn.status?.toUpperCase() !== "APPROVED" && canModify
          ? [
              {
                icon: <Pencil className="h-3.5 w-3.5" />,
                label: "Edit",
                tooltip: "Edit GRN",
                onClick: () => router.push(`/grn/${grn.id}/edit`),
              },
            ]
          : []),
      ];
    },
    [router, userId, userRole],
  );

  return (
    <DataTable
      columns={columns}
      data={grns}
      actions={getActions}
      hideSearchBar={false}
      renderRowActions={(grn: GoodsReceivedNote) => {
        const canModify =
          grn.createdBy === userId ||
          grn.receivedBy === userId ||
          GRN_EDIT_ROLES.includes(userRole);
        return (
          <GrnOptionsMenu grn={grn} router={router} canModify={canModify} />
        );
      }}
    />
  );
}
