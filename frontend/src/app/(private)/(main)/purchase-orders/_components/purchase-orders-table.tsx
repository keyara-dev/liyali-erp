"use client";

import { useCallback, useMemo, useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { Download, Eye, FileText, MoreVertical, Search } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Badge } from "@/components/ui/badge";
import { DataList, DataListColumn } from "@/components/ui/data-list";
import { FilterBar } from "@/components/ui/filter-bar";
import { StatusBadge } from "@/components/status-badge";
import { PurchaseOrder } from "@/types/purchase-order";
import { usePurchaseOrders } from "@/hooks/use-purchase-order-queries";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { formatCurrency } from "@/lib/utils";
import { useDebounce } from "@/hooks/use-debounce";

interface PurchaseOrdersTableProps {
  userId: string;
  userRole: string;
  refreshTrigger: number;
  onRefresh: () => void;
}

// Options dropdown component
function PoOptionsMenu({
  po,
  router,
}: {
  po: PurchaseOrder;
  router: ReturnType<typeof useRouter>;
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
      <DropdownMenuContent align="end" className="w-44">
        <DropdownMenuItem
          onClick={() => router.push(`/purchase-orders/${po.id}`)}
        >
          <Eye className="mr-2 h-4 w-4" />
          View Details
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

export function PurchaseOrdersTable({
  userId: _userId,
  userRole: _userRole,
  refreshTrigger: _refreshTrigger,
  onRefresh: _onRefresh,
}: PurchaseOrdersTableProps) {
  const router = useRouter();
  const { data: purchaseOrders = [], isLoading, refetch } = usePurchaseOrders();

  // Refetch when refreshTrigger changes
  useEffect(() => {
    refetch();
  }, [_refreshTrigger, refetch]);

  // Filter state
  const [searchQuery, setSearchQuery] = useState("");
  const [statusFilter, setStatusFilter] = useState<string>("all");
  const [routingFilter, setRoutingFilter] = useState<string>("all");
  const debouncedSearch = useDebounce(searchQuery, 500);

  // Client-side filtering
  const filteredPos = useMemo(() => {
    let filtered = purchaseOrders;

    if (statusFilter !== "all") {
      filtered = filtered.filter(
        (p) => p.status?.toLowerCase() === statusFilter,
      );
    }

    if (routingFilter !== "all") {
      filtered = filtered.filter((p) => p.routingType === routingFilter);
    }

    if (debouncedSearch) {
      const s = debouncedSearch.toLowerCase();
      filtered = filtered.filter(
        (p) =>
          p.documentNumber?.toLowerCase().includes(s) ||
          p.vendorName?.toLowerCase().includes(s) ||
          p.vendor?.name?.toLowerCase().includes(s) ||
          p.linkedRequisition?.toLowerCase().includes(s),
      );
    }

    return filtered;
  }, [purchaseOrders, statusFilter, routingFilter, debouncedSearch]);

  const hasActiveFilters =
    Boolean(searchQuery) || statusFilter !== "all" || routingFilter !== "all";

  const clearFilters = useCallback(() => {
    setSearchQuery("");
    setStatusFilter("all");
    setRoutingFilter("all");
  }, []);

  const columns: DataListColumn<PurchaseOrder>[] = [
    {
      id: "documentNumber",
      header: "PO #",
      priority: "always",
      cell: (row) => (
        <div className="font-semibold uppercase">
          {row.documentNumber || row.id}
        </div>
      ),
    },
    {
      id: "title",
      header: "Title",
      priority: "md",
      cell: (row) => (
        <span className="font-medium capitalize line-clamp-1">
          {row.title || "—"}
        </span>
      ),
    },
    {
      id: "vendor",
      header: "Vendor",
      priority: "md",
      cell: (row) => (
        <span className="font-medium">
          {row.vendor?.name ?? row.vendorName ?? "—"}
        </span>
      ),
    },
    {
      id: "linkedRequisition",
      header: "Linked Req",
      priority: "lg",
      cell: (row) => (
        <span className="text-muted-foreground">
          {row.linkedRequisition || "—"}
        </span>
      ),
    },
    {
      id: "totalAmount",
      header: "Total",
      priority: "md",
      align: "right",
      cell: (row) => (
        <div className="font-medium">
          {formatCurrency(row.total ?? row.totalAmount ?? 0, row.currency)}
        </div>
      ),
    },
    {
      id: "status",
      header: "Status",
      cell: (row) => (
        <div className="flex flex-col gap-1">
          <StatusBadge status={row.status} type="document" />
          {row.routingType === "direct_payment" && (
            <Badge
              variant="outline"
              className="border-purple-500 text-purple-700 text-[10px] px-1.5 py-0 w-fit"
            >
              Direct Payment
            </Badge>
          )}
          {row.routingType === "accounting" && (
            <Badge
              variant="outline"
              className="border-amber-500 text-amber-700 text-[10px] px-1.5 py-0 w-fit"
            >
              Accounting
            </Badge>
          )}
        </div>
      ),
    },
    {
      id: "deliveryDate",
      header: "Delivery",
      priority: "lg",
      cell: (row) =>
        row.deliveryDate
          ? new Date(row.deliveryDate).toLocaleDateString()
          : "—",
    },
    {
      id: "actions",
      header: <span className="sr-only">Actions</span>,
      cell: (row) => <PoOptionsMenu po={row} router={router} />,
    },
  ];

  return (
    <div className="space-y-4">
      <FilterBar
        search={
          <div className="relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
            <Input
              placeholder="Search PO number, vendor, requisition…"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="pl-10"
            />
          </div>
        }
        filters={
          <>
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
                <SelectItem value="fulfilled">Fulfilled</SelectItem>
                <SelectItem value="cancelled">Cancelled</SelectItem>
              </SelectContent>
            </Select>

            <Select value={routingFilter} onValueChange={setRoutingFilter}>
              <SelectTrigger className="w-44">
                <SelectValue placeholder="Routing type" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All types</SelectItem>
                <SelectItem value="procurement">Procurement</SelectItem>
                <SelectItem value="accounting">Accounting</SelectItem>
                <SelectItem value="direct_payment">Direct Payment</SelectItem>
              </SelectContent>
            </Select>
          </>
        }
        hasActiveFilters={hasActiveFilters}
        onReset={clearFilters}
        meta={`${filteredPos.length} purchase order${filteredPos.length === 1 ? "" : "s"}${hasActiveFilters ? " (filtered)" : ""}`}
      />

      <DataList<PurchaseOrder>
        rows={filteredPos}
        columns={columns}
        getRowId={(row) => row.id}
        isLoading={isLoading}
        emptyMessage={
          <div className="flex flex-col items-center gap-3 py-4">
            <FileText className="h-10 w-10 text-muted-foreground" />
            <div>
              <p className="font-medium">No Purchase Orders Found</p>
              <p className="text-xs text-muted-foreground mt-1">
                {hasActiveFilters
                  ? "No purchase orders match your current filters. Try adjusting your search criteria."
                  : "No purchase orders have been created yet."}
              </p>
            </div>
          </div>
        }
        mobileCard={(row) => (
          <div className="flex flex-col gap-2">
            <div className="flex items-start justify-between gap-2">
              <div className="min-w-0">
                <div className="font-medium text-primary line-clamp-1 uppercase">
                  {row.documentNumber || row.id}
                </div>
                {row.title && (
                  <div className="text-xs font-medium line-clamp-1 capitalize">
                    {row.title}
                  </div>
                )}
                <div className="text-xs text-muted-foreground line-clamp-1">
                  {row.vendor?.name ?? row.vendorName}
                </div>
              </div>
              <div className="flex flex-col items-end gap-1">
                <StatusBadge status={row.status} type="document" />
                {row.routingType === "direct_payment" && (
                  <Badge
                    variant="outline"
                    className="border-purple-500 text-purple-700 text-[10px] px-1.5 py-0"
                  >
                    Direct Payment
                  </Badge>
                )}
                {row.routingType === "accounting" && (
                  <Badge
                    variant="outline"
                    className="border-amber-500 text-amber-700 text-[10px] px-1.5 py-0"
                  >
                    Accounting
                  </Badge>
                )}
              </div>
            </div>
            <div className="flex flex-wrap items-center gap-2 text-xs text-muted-foreground">
              <span>
                {formatCurrency(
                  row.total ?? row.totalAmount ?? 0,
                  row.currency,
                )}
              </span>
              {row.deliveryDate && (
                <span>{new Date(row.deliveryDate).toLocaleDateString()}</span>
              )}
            </div>
            <div className="pt-1">
              <PoOptionsMenu po={row} router={router} />
            </div>
          </div>
        )}
      />
    </div>
  );
}
