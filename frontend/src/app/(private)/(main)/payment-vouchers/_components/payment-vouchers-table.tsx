"use client";

import { useCallback, useMemo, useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import {
  Download,
  Eye,
  FileText,
  MoreVertical,
  Pencil,
  Search,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { DataList, DataListColumn } from "@/components/ui/data-list";
import { FilterBar } from "@/components/ui/filter-bar";
import { StatusBadge } from "@/components/status-badge";
import { usePaymentVouchers } from "@/hooks/use-payment-voucher-queries";
import { PaymentVoucher } from "@/types/payment-voucher";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { formatCurrency } from "@/lib/utils";
import { useDebounce } from "@/hooks/use-debounce";

interface PaymentVouchersTableProps {
  userId: string;
  userRole: string;
  refreshTrigger: number;
  onRefresh: () => void;
}

const FINANCE_EDIT_ROLES = ["admin", "finance"];

// Options dropdown component
function PvOptionsMenu({
  pv,
  router,
  canEdit,
}: {
  pv: PaymentVoucher;
  router: ReturnType<typeof useRouter>;
  canEdit: boolean;
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
          onClick={() => router.push(`/payment-vouchers/${pv.id}`)}
        >
          <Eye className="mr-2 h-4 w-4" />
          View Details
        </DropdownMenuItem>
        {canEdit && (
          <DropdownMenuItem
            onClick={() => router.push(`/payment-vouchers/${pv.id}/edit`)}
          >
            <Pencil className="mr-2 h-4 w-4" />
            Edit
          </DropdownMenuItem>
        )}
        <DropdownMenuItem
          onClick={() => console.log("Download PDF for PV:", pv.id)}
        >
          <Download className="mr-2 h-4 w-4" />
          Download
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

export function PaymentVouchersTable({
  userId,
  userRole,
  refreshTrigger,
  onRefresh: _onRefresh,
}: PaymentVouchersTableProps) {
  const router = useRouter();
  const { data: pvs = [], isLoading, refetch } = usePaymentVouchers();

  // Refetch when refreshTrigger changes
  useEffect(() => {
    refetch();
  }, [refreshTrigger, refetch]);

  // Filter state
  const [searchQuery, setSearchQuery] = useState("");
  const [statusFilter, setStatusFilter] = useState<string>("all");
  const [methodFilter, setMethodFilter] = useState<string>("all");
  const [routingFilter, setRoutingFilter] = useState<string>("all");
  const debouncedSearch = useDebounce(searchQuery, 500);

  // Client-side filtering
  const filteredPvs = useMemo(() => {
    let filtered = pvs;

    if (statusFilter !== "all") {
      filtered = filtered.filter(
        (p) => p.status?.toLowerCase() === statusFilter,
      );
    }

    if (methodFilter !== "all") {
      filtered = filtered.filter(
        (p) => p.paymentMethod?.toLowerCase() === methodFilter,
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
          p.linkedPO?.toLowerCase().includes(s),
      );
    }

    return filtered;
  }, [pvs, statusFilter, methodFilter, routingFilter, debouncedSearch]);

  const hasActiveFilters =
    Boolean(searchQuery) ||
    statusFilter !== "all" ||
    methodFilter !== "all" ||
    routingFilter !== "all";

  const clearFilters = useCallback(() => {
    setSearchQuery("");
    setStatusFilter("all");
    setMethodFilter("all");
    setRoutingFilter("all");
  }, []);

  const columns: DataListColumn<PaymentVoucher>[] = [
    {
      id: "documentNumber",
      header: "PV #",
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
      id: "payee",
      header: "Payee",
      priority: "md",
      cell: (row) => (
        <span className="font-medium">{row.vendorName ?? "—"}</span>
      ),
    },
    {
      id: "linkedPO",
      header: "Linked PO",
      priority: "lg",
      cell: (row) => (
        <span className="text-muted-foreground">{row.linkedPO || "—"}</span>
      ),
    },
    {
      id: "amount",
      header: "Amount",
      priority: "md",
      align: "right",
      cell: (row) => (
        <div className="font-medium">
          {formatCurrency(row.totalAmount ?? row.amount ?? 0, row.currency)}
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
            <div className="flex flex-col gap-0.5">
              <Badge
                variant="outline"
                className="border-purple-500 text-purple-700 text-[10px] px-1.5 py-0 w-fit"
              >
                Direct Payment
              </Badge>
              {(row.metadata as any)?.autoCreated &&
                (row.metadata as any)?.sourceReqID && (
                  <span className="text-[10px] text-muted-foreground">
                    Auto from REQ-
                    {String((row.metadata as any).sourceReqID).slice(0, 8)}
                  </span>
                )}
            </div>
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
      id: "dueDate",
      header: "Due",
      priority: "lg",
      cell: (row) => {
        if (!row.paymentDueDate)
          return <span className="text-muted-foreground">—</span>;
        const d = new Date(row.paymentDueDate);
        const overdue = d < new Date() && row.status?.toLowerCase() !== "paid";
        return (
          <span className={overdue ? "text-rose-600 font-medium" : ""}>
            {d.toLocaleDateString()}
            {overdue && (
              <span className="ml-1.5 text-[10px] uppercase">Overdue</span>
            )}
          </span>
        );
      },
    },
    {
      id: "actions",
      header: <span className="sr-only">Actions</span>,
      cell: (row) => (
        <PvOptionsMenu
          pv={row}
          router={router}
          canEdit={
            row.createdBy === userId || FINANCE_EDIT_ROLES.includes(userRole)
          }
        />
      ),
    },
  ];

  return (
    <div className="space-y-4">
      <FilterBar
        search={
          <div className="relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
            <Input
              placeholder="Search PV number, vendor, linked PO…"
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
                <SelectItem value="paid">Paid</SelectItem>
                <SelectItem value="cancelled">Cancelled</SelectItem>
              </SelectContent>
            </Select>

            <Select value={methodFilter} onValueChange={setMethodFilter}>
              <SelectTrigger className="w-44">
                <SelectValue placeholder="Payment Method" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All methods</SelectItem>
                <SelectItem value="bank_transfer">Bank Transfer</SelectItem>
                <SelectItem value="cash">Cash</SelectItem>
                <SelectItem value="check">Check</SelectItem>
                <SelectItem value="mobile_money">Mobile Money</SelectItem>
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
        meta={`${filteredPvs.length} voucher${filteredPvs.length === 1 ? "" : "s"}${hasActiveFilters ? " (filtered)" : ""}`}
      />

      <DataList<PaymentVoucher>
        rows={filteredPvs}
        columns={columns}
        getRowId={(row) => row.id}
        isLoading={isLoading}
        emptyMessage={
          <div className="flex flex-col items-center gap-3 py-4">
            <FileText className="h-10 w-10 text-muted-foreground" />
            <div>
              <p className="font-medium">No Payment Vouchers Found</p>
              <p className="text-xs text-muted-foreground mt-1">
                {hasActiveFilters
                  ? "No vouchers match your current filters. Try adjusting your search criteria."
                  : "No payment vouchers have been created yet."}
              </p>
            </div>
          </div>
        }
        mobileCard={(row) => {
          const canEdit =
            row.createdBy === userId || FINANCE_EDIT_ROLES.includes(userRole);
          return (
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
                    {row.vendorName}
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
                    row.totalAmount ?? row.amount ?? 0,
                    row.currency,
                  )}
                </span>
                {row.paymentDueDate &&
                  (() => {
                    const d = new Date(row.paymentDueDate);
                    const overdue =
                      d < new Date() && row.status?.toLowerCase() !== "paid";
                    return (
                      <span
                        className={overdue ? "text-rose-600 font-medium" : ""}
                      >
                        Due {d.toLocaleDateString()}
                        {overdue && " (overdue)"}
                      </span>
                    );
                  })()}
              </div>
              <div className="pt-1">
                <PvOptionsMenu pv={row} router={router} canEdit={canEdit} />
              </div>
            </div>
          );
        }}
      />
    </div>
  );
}
