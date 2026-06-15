"use client";

import { useRouter } from "next/navigation";
import Link from "next/link";
import { useCallback, useMemo, useEffect, useState } from "react";
import {
  Download,
  Eye,
  FileText,
  MoreVertical,
  Pencil,
  Search,
} from "lucide-react";

import { DataList, DataListColumn } from "@/components/ui/data-list";
import { FilterBar } from "@/components/ui/filter-bar";
import { StatusBadge } from "@/components/status-badge";
import { UserCell } from "@/components/ui/user-cell";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import type { GoodsReceivedNote } from "@/app/_actions/grn-actions";
import type { PurchaseOrder } from "@/types/purchase-order";
import { useGRNs } from "@/hooks/use-grn-queries";
import { usePurchaseOrders } from "@/hooks/use-purchase-order-queries";
import { formatCurrency } from "@/lib/utils";
import { useDebounce } from "@/hooks/use-debounce";

interface GrnTableProps {
  userId: string;
  userRole: string;
  refreshTrigger: number;
  onRefresh: () => void;
  /** When set, only GRNs linked to this PO doc number are shown. */
  poFilter?: string;
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
    if (item.description)
      priceByDesc.set(item.description, item.unitPrice ?? 0);
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
        <Button
          variant="outline"
          size="icon"
          className="h-8 w-8"
          aria-label="Row actions"
        >
          <MoreVertical className="h-4 w-4" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        <DropdownMenuItem onClick={() => router.push(`/grn/${grn.id}`)}>
          <Eye className="mr-2 h-4 w-4" />
          View Details
        </DropdownMenuItem>
        {/* Edit is only meaningful on DRAFT GRNs. Route to detail page for now;
            a dedicated edit dialog on the detail page can be added later. */}
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
  poFilter,
}: GrnTableProps) {
  const router = useRouter();
  const { data: grns = [], isLoading, refetch } = useGRNs(1, 50);
  const { data: purchaseOrders = [] } = usePurchaseOrders();

  useEffect(() => {
    refetch();
  }, [refreshTrigger, refetch]);

  // Filter state
  const [searchQuery, setSearchQuery] = useState("");
  const [statusFilter, setStatusFilter] = useState<string>("all");
  const debouncedSearch = useDebounce(searchQuery, 500);

  // Map POs by documentNumber so each GRN row can resolve vendor / amount /
  // currency / PO link without an N+1 request.
  const posByDocNumber = useMemo(() => {
    const map = new Map<string, PurchaseOrder>();
    for (const po of purchaseOrders as PurchaseOrder[]) {
      if (po.documentNumber) map.set(po.documentNumber, po);
    }
    return map;
  }, [purchaseOrders]);

  // Client-side filtering
  const filteredGrns = useMemo(() => {
    let filtered = grns;

    if (poFilter) {
      filtered = filtered.filter(
        (g) => g.poDocumentNumber === poFilter,
      );
    }

    if (statusFilter !== "all") {
      filtered = filtered.filter(
        (g) => g.status?.toLowerCase() === statusFilter,
      );
    }

    if (debouncedSearch) {
      const s = debouncedSearch.toLowerCase();
      filtered = filtered.filter(
        (g) =>
          g.documentNumber?.toLowerCase().includes(s) ||
          g.poDocumentNumber?.toLowerCase().includes(s) ||
          (g.receiver?.name || g.receivedByName)?.toLowerCase().includes(s) ||
          // vendorName is on the linked PO, not on the GRN itself
          posByDocNumber
            .get(g.poDocumentNumber)
            ?.vendorName?.toLowerCase()
            .includes(s),
      );
    }

    return filtered;
  }, [grns, statusFilter, debouncedSearch, posByDocNumber, poFilter]);

  const hasActiveFilters = Boolean(searchQuery) || statusFilter !== "all";

  const clearFilters = useCallback(() => {
    setSearchQuery("");
    setStatusFilter("all");
  }, []);

  const columns: DataListColumn<GoodsReceivedNote>[] = [
    {
      id: "documentNumber",
      header: "GRN #",
      priority: "always",
      cell: (row) => (
        <div className="font-semibold uppercase">
          {row.documentNumber || row.id}
        </div>
      ),
    },
    {
      id: "poDocumentNumber",
      header: "PO",
      priority: "md",
      cell: (row) => {
        const poDoc = row.poDocumentNumber;
        const po = poDoc ? posByDocNumber.get(poDoc) : undefined;
        if (!poDoc) return <span className="text-muted-foreground">—</span>;
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
      id: "poTitle",
      header: "Title",
      priority: "lg",
      cell: (row) => {
        const po = posByDocNumber.get(row.poDocumentNumber);
        const title = po?.title || row.notes;
        return (
          <span className="font-medium capitalize line-clamp-1">
            {title || "—"}
          </span>
        );
      },
    },
    {
      id: "vendor",
      header: "Vendor",
      priority: "lg",
      cell: (row) => {
        const vendor = posByDocNumber
          .get(row.poDocumentNumber)
          ?.vendorName?.trim();
        return <span className="font-medium capitalize">{vendor || "—"}</span>;
      },
    },
    {
      id: "amount",
      header: "Received Value",
      priority: "lg",
      align: "right",
      cell: (row) => {
        const po = posByDocNumber.get(row.poDocumentNumber);
        const amount = computeReceivedAmount(row, po);
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
      id: "receivedBy",
      header: "Received by",
      priority: "lg",
      cell: (row) => (
        <UserCell user={row.receiver} fallbackName={row.receivedByName} />
      ),
    },
    {
      id: "receivedDate",
      header: "Received",
      priority: "md",
      cell: (row) => {
        const date = row.receivedDate || row.createdAt;
        return (
          <span className="text-sm text-muted-foreground">
            {date ? new Date(date).toLocaleDateString() : "—"}
          </span>
        );
      },
    },
    {
      id: "status",
      header: "Status",
      cell: (row) => (
        <StatusBadge status={row.status || "DRAFT"} type="document" />
      ),
    },
    {
      id: "actions",
      header: <span className="sr-only">Actions</span>,
      cell: (row) => {
        const canModify =
          row.createdBy === userId ||
          row.receivedBy === userId ||
          GRN_EDIT_ROLES.includes(userRole);
        return (
          <GrnOptionsMenu grn={row} router={router} canModify={canModify} />
        );
      },
    },
  ];

  return (
    <div className="space-y-4">
      <FilterBar
        search={
          <div className="relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
            <Input
              placeholder="Search GRN number, PO reference, receiver…"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="pl-10"
            />
          </div>
        }
        filters={
          <Select value={statusFilter} onValueChange={setStatusFilter}>
            <SelectTrigger className="w-40">
              <SelectValue placeholder="Status" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All statuses</SelectItem>
              <SelectItem value="draft">Draft</SelectItem>
              <SelectItem value="pending">Pending</SelectItem>
              <SelectItem value="approved">Approved</SelectItem>
              <SelectItem value="rejected">Rejected</SelectItem>
              <SelectItem value="revision">Revision</SelectItem>
              <SelectItem value="completed">Completed</SelectItem>
              <SelectItem value="cancelled">Cancelled</SelectItem>
            </SelectContent>
          </Select>
        }
        hasActiveFilters={hasActiveFilters}
        onReset={clearFilters}
        meta={`${filteredGrns.length} GRN${filteredGrns.length === 1 ? "" : "s"}${hasActiveFilters ? " (filtered)" : ""}`}
      />

      <DataList<GoodsReceivedNote>
        rows={filteredGrns}
        columns={columns}
        getRowId={(row) => row.id}
        isLoading={isLoading}
        emptyMessage={
          <div className="flex flex-col items-center gap-3 py-4">
            <FileText className="h-10 w-10 text-muted-foreground" />
            <div>
              <p className="font-medium">No Goods Received Notes Found</p>
              <p className="text-xs text-muted-foreground mt-1">
                {hasActiveFilters
                  ? "No GRNs match your current filters. Try adjusting your search criteria."
                  : "No GRNs have been created yet."}
              </p>
            </div>
          </div>
        }
        mobileCard={(row) => {
          const po = posByDocNumber.get(row.poDocumentNumber);
          const vendor = po?.vendorName?.trim();
          const title = po?.title || row.notes;
          const canModify =
            row.createdBy === userId ||
            row.receivedBy === userId ||
            GRN_EDIT_ROLES.includes(userRole);
          return (
            <div className="flex flex-col gap-2">
              <div className="flex items-start justify-between gap-2">
                <div className="min-w-0">
                  <div className="font-medium text-primary line-clamp-1 uppercase">
                    {row.documentNumber || row.id}
                  </div>
                  {title && (
                    <div className="text-xs font-medium line-clamp-1 capitalize">
                      {title}
                    </div>
                  )}
                  <div className="text-xs text-muted-foreground line-clamp-1">
                    {row.poDocumentNumber ? `PO ${row.poDocumentNumber}` : "—"}
                  </div>
                </div>
                <StatusBadge status={row.status || "DRAFT"} type="document" />
              </div>
              <div className="flex flex-wrap items-center gap-2 text-xs text-muted-foreground">
                {vendor && <span className="capitalize">{vendor}</span>}
                {(row.receiver?.name || row.receivedByName) && (
                  <span>{row.receiver?.name || row.receivedByName}</span>
                )}
                {(row.receivedDate || row.createdAt) && (
                  <span>
                    {new Date(
                      row.receivedDate || row.createdAt,
                    ).toLocaleDateString()}
                  </span>
                )}
              </div>
              <div className="pt-1">
                <GrnOptionsMenu
                  grn={row}
                  router={router}
                  canModify={canModify}
                />
              </div>
            </div>
          );
        }}
      />
    </div>
  );
}
