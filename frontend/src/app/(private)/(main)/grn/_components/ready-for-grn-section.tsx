"use client";

import { useMemo, useState } from "react";
import { useRouter } from "next/navigation";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";
import { StatusBadge } from "@/components/status-badge";
import {
  Package,
  Eye,
  ChevronLeft,
  ChevronRight,
  Loader2,
  ArrowRight,
  CheckCircle2,
  Truck,
  Wallet,
} from "lucide-react";
import { usePurchaseOrders } from "@/hooks/use-purchase-order-queries";
import { usePaymentVouchers } from "@/hooks/use-payment-voucher-queries";
import { useGRNs } from "@/hooks/use-grn-queries";
import { usePermissions } from "@/hooks/use-permissions";
import { formatCurrency } from "@/lib/utils";
import { CreateGRNDialog } from "./create-grn-dialog";
import type { PurchaseOrder } from "@/types/purchase-order";
import type { PaymentVoucher } from "@/types/payment-voucher";
import type { GoodsReceivedNote } from "@/hooks/use-grn-queries";

const PAGE_SIZE = 5;

interface ReadyForGrnSectionProps {
  userId: string;
  userRole: string;
  /** Bump the parent GRN table after a GRN is created. */
  onChanged?: () => void;
}

interface PORow {
  po: PurchaseOrder;
  grn?: GoodsReceivedNote;
}

interface PVRow {
  pv: PaymentVoucher;
  grn?: GoodsReceivedNote;
}

export function ReadyForGrnSection({ onChanged }: ReadyForGrnSectionProps) {
  const router = useRouter();
  const { hasPermission, isAdmin, isFinance } = usePermissions();
  const canCreate =
    hasPermission("grn", "create") || isAdmin() || isFinance();

  const { data: purchaseOrders = [], isLoading: posLoading } =
    usePurchaseOrders();
  const { data: paymentVouchers = [], isLoading: pvsLoading } =
    usePaymentVouchers();
  const { data: grns = [], isLoading: grnsLoading } = useGRNs(1, 200);

  const [dialogOpen, setDialogOpen] = useState(false);
  const [sourcePO, setSourcePO] = useState<PurchaseOrder | undefined>();
  const [sourcePV, setSourcePV] = useState<PaymentVoucher | undefined>();

  const [poPage, setPoPage] = useState(1);
  const [pvPage, setPvPage] = useState(1);

  // Approved POs (goods-first source), paired with their GRN if one exists.
  const poRows = useMemo<PORow[]>(() => {
    const grnList = grns as GoodsReceivedNote[];
    return (purchaseOrders as PurchaseOrder[])
      .filter((po) => po.status?.toUpperCase() === "APPROVED")
      .map((po) => ({
        po,
        grn: grnList.find((g) => g.poDocumentNumber === po.documentNumber),
      }));
  }, [purchaseOrders, grns]);

  // Approved/paid PVs linked to a PO (payment-first source).
  const pvRows = useMemo<PVRow[]>(() => {
    const grnList = grns as GoodsReceivedNote[];
    return (paymentVouchers as PaymentVoucher[])
      .filter((pv) => {
        const s = pv.status?.toUpperCase();
        return (s === "APPROVED" || s === "PAID") && !!pv.linkedPO;
      })
      .map((pv) => ({
        pv,
        grn:
          grnList.find((g) => g.linkedPV === pv.documentNumber) ??
          (pv.linkedGRN
            ? grnList.find((g) => g.documentNumber === pv.linkedGRN)
            : undefined),
      }));
  }, [paymentVouchers, grns]);

  const isLoading = posLoading || pvsLoading || grnsLoading;

  const openForPO = (po: PurchaseOrder) => {
    setSourcePV(undefined);
    setSourcePO(po);
    setDialogOpen(true);
  };

  const openForPV = (pv: PaymentVoucher) => {
    setSourcePO(undefined);
    setSourcePV(pv);
    setDialogOpen(true);
  };

  if (isLoading && poRows.length === 0 && pvRows.length === 0) {
    return (
      <Card className="p-8">
        <div className="flex items-center justify-center">
          <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
          <span className="ml-3 text-muted-foreground">
            Loading documents ready for goods receipt...
          </span>
        </div>
      </Card>
    );
  }

  if (poRows.length === 0 && pvRows.length === 0) {
    return (
      <Card className="p-8">
        <div className="text-center">
          <Package className="mx-auto mb-4 h-12 w-12 text-muted-foreground" />
          <h3 className="mb-2 text-lg font-semibold">Nothing ready for GRN</h3>
          <p className="text-muted-foreground">
            Approved purchase orders and approved/paid payment vouchers will
            appear here once available for goods receipt.
          </p>
        </div>
      </Card>
    );
  }

  const poStart = (poPage - 1) * PAGE_SIZE;
  const pagedPORows = poRows.slice(poStart, poStart + PAGE_SIZE);
  const pvStart = (pvPage - 1) * PAGE_SIZE;
  const pagedPVRows = pvRows.slice(pvStart, pvStart + PAGE_SIZE);

  return (
    <>
      <div className="space-y-6">
        {/* ── Approved Purchase Orders (goods-first) ── */}
        {poRows.length > 0 && (
          <Card>
            <div className="p-6">
              <div className="mb-4 flex items-center justify-between gap-3">
                <div className="flex items-center gap-2">
                  <Truck className="h-4 w-4 text-muted-foreground" />
                  <div>
                    <h3 className="text-base font-semibold">
                      Approved Purchase Orders
                    </h3>
                    <p className="text-sm text-muted-foreground">
                      Record goods received against an approved PO
                    </p>
                  </div>
                </div>
                <Badge variant="secondary">
                  {poRows.length} PO{poRows.length !== 1 ? "s" : ""}
                </Badge>
              </div>

              <div className="rounded-md border">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>PO Number</TableHead>
                      <TableHead>Title</TableHead>
                      <TableHead>Vendor</TableHead>
                      <TableHead>Amount</TableHead>
                      <TableHead>Items</TableHead>
                      <TableHead>Approved</TableHead>
                      <TableHead className="text-right">Actions</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {pagedPORows.map(({ po, grn }) => (
                      <TableRow key={po.id}>
                        <TableCell className="font-mono text-sm">
                          {po.documentNumber}
                        </TableCell>
                        <TableCell className="max-w-[180px] truncate font-medium capitalize">
                          {po.title || po.description || "—"}
                        </TableCell>
                        <TableCell>{po.vendorName || "—"}</TableCell>
                        <TableCell className="font-mono">
                          {formatCurrency(po.totalAmount, po.currency)}
                        </TableCell>
                        <TableCell>{po.items?.length || 0}</TableCell>
                        <TableCell className="text-sm">
                          {po.updatedAt
                            ? new Date(po.updatedAt).toLocaleDateString(
                                "en-ZM",
                                { year: "numeric", month: "short", day: "numeric" },
                              )
                            : "—"}
                        </TableCell>
                        <TableCell className="text-right">
                          <div className="flex items-center justify-end gap-2">
                            <Button
                              variant="ghost"
                              size="sm"
                              onClick={() =>
                                router.push(`/purchase-orders/${po.id}`)
                              }
                            >
                              <Eye className="mr-1 h-4 w-4" />
                              View
                            </Button>
                            {grn ? (
                              <div className="flex items-center gap-2">
                                <Badge
                                  variant="outline"
                                  className="gap-1 border-green-300 text-green-700 dark:border-green-800 dark:text-green-400"
                                >
                                  <CheckCircle2 className="h-3 w-3" />
                                  Received
                                </Badge>
                                <Button
                                  variant="outline"
                                  size="sm"
                                  onClick={() => router.push(`/grn/${grn.id}`)}
                                >
                                  View GRN
                                  <ArrowRight className="ml-1 h-3 w-3" />
                                </Button>
                              </div>
                            ) : (
                              canCreate && (
                                <Button
                                  variant="default"
                                  size="sm"
                                  onClick={() => openForPO(po)}
                                >
                                  <Package className="mr-1 h-4 w-4" />
                                  Create GRN
                                </Button>
                              )
                            )}
                          </div>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>

              {poRows.length > PAGE_SIZE && (
                <Pagination
                  page={poPage}
                  count={pagedPORows.length}
                  total={poRows.length}
                  onPrev={() => setPoPage((p) => Math.max(1, p - 1))}
                  onNext={() =>
                    setPoPage((p) =>
                      poStart + PAGE_SIZE < poRows.length ? p + 1 : p,
                    )
                  }
                  label="POs"
                />
              )}
            </div>
          </Card>
        )}

        {/* ── Approved / Paid Payment Vouchers (payment-first) ── */}
        {pvRows.length > 0 && (
          <Card>
            <div className="p-6">
              <div className="mb-4 flex items-center justify-between gap-3">
                <div className="flex items-center gap-2">
                  <Wallet className="h-4 w-4 text-muted-foreground" />
                  <div>
                    <h3 className="text-base font-semibold">
                      Approved / Paid Payment Vouchers
                    </h3>
                    <p className="text-sm text-muted-foreground">
                      Record delivery funded by an approved/paid PV
                    </p>
                  </div>
                </div>
                <Badge variant="secondary">
                  {pvRows.length} PV{pvRows.length !== 1 ? "s" : ""}
                </Badge>
              </div>

              <div className="rounded-md border">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>PV Number</TableHead>
                      <TableHead>Linked PO</TableHead>
                      <TableHead>Vendor</TableHead>
                      <TableHead>Amount</TableHead>
                      <TableHead>Status</TableHead>
                      <TableHead className="text-right">Actions</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {pagedPVRows.map(({ pv, grn }) => (
                      <TableRow key={pv.id}>
                        <TableCell className="font-mono text-sm">
                          {pv.documentNumber}
                        </TableCell>
                        <TableCell className="font-mono text-sm">
                          {pv.linkedPO || "—"}
                        </TableCell>
                        <TableCell>{pv.vendorName || "—"}</TableCell>
                        <TableCell className="font-mono">
                          {formatCurrency(pv.amount, pv.currency)}
                        </TableCell>
                        <TableCell>
                          <StatusBadge status={pv.status} type="document" />
                        </TableCell>
                        <TableCell className="text-right">
                          <div className="flex items-center justify-end gap-2">
                            <Button
                              variant="ghost"
                              size="sm"
                              onClick={() =>
                                router.push(`/payment-vouchers/${pv.id}`)
                              }
                            >
                              <Eye className="mr-1 h-4 w-4" />
                              View
                            </Button>
                            {grn ? (
                              <div className="flex items-center gap-2">
                                <Badge
                                  variant="outline"
                                  className="gap-1 border-green-300 text-green-700 dark:border-green-800 dark:text-green-400"
                                >
                                  <CheckCircle2 className="h-3 w-3" />
                                  Received
                                </Badge>
                                <Button
                                  variant="outline"
                                  size="sm"
                                  onClick={() => router.push(`/grn/${grn.id}`)}
                                >
                                  View GRN
                                  <ArrowRight className="ml-1 h-3 w-3" />
                                </Button>
                              </div>
                            ) : (
                              canCreate && (
                                <Button
                                  variant="default"
                                  size="sm"
                                  onClick={() => openForPV(pv)}
                                >
                                  <Package className="mr-1 h-4 w-4" />
                                  Create GRN
                                </Button>
                              )
                            )}
                          </div>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>

              {pvRows.length > PAGE_SIZE && (
                <Pagination
                  page={pvPage}
                  count={pagedPVRows.length}
                  total={pvRows.length}
                  onPrev={() => setPvPage((p) => Math.max(1, p - 1))}
                  onNext={() =>
                    setPvPage((p) =>
                      pvStart + PAGE_SIZE < pvRows.length ? p + 1 : p,
                    )
                  }
                  label="PVs"
                />
              )}
            </div>
          </Card>
        )}
      </div>

      <CreateGRNDialog
        open={dialogOpen}
        onOpenChange={setDialogOpen}
        initialPurchaseOrder={sourcePO}
        initialPaymentVoucher={sourcePV}
        onSuccess={onChanged}
      />
    </>
  );
}

function Pagination({
  page,
  count,
  total,
  onPrev,
  onNext,
  label,
}: {
  page: number;
  count: number;
  total: number;
  onPrev: () => void;
  onNext: () => void;
  label: string;
}) {
  return (
    <div className="mt-4 flex items-center justify-between">
      <div className="text-sm text-muted-foreground">
        Page {page} • Showing {count} of {total} {label}
      </div>
      <div className="flex items-center gap-2">
        <Button
          variant="outline"
          size="sm"
          onClick={onPrev}
          disabled={page === 1}
        >
          <ChevronLeft className="mr-1 h-4 w-4" />
          Previous
        </Button>
        <Button
          variant="outline"
          size="sm"
          onClick={onNext}
          disabled={page * PAGE_SIZE >= total}
        >
          Next
          <ChevronRight className="ml-1 h-4 w-4" />
        </Button>
      </div>
    </div>
  );
}
