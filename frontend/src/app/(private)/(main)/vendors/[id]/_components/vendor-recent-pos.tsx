"use client";

import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { DataList, type DataListColumn } from "@/components/ui/data-list";
import { StatusBadge } from "@/components/status-badge";
import { usePurchaseOrders } from "@/hooks/use-purchase-order-queries";
import { formatCurrency } from "@/lib/utils";
import { useRouter } from "next/navigation";
import type { PurchaseOrder } from "@/types/purchase-order";

interface VendorRecentPosProps {
  vendorId: string;
  limit?: number;
}

export function VendorRecentPos({ vendorId, limit = 10 }: VendorRecentPosProps) {
  const router = useRouter();
  const { data: pos, isLoading } = usePurchaseOrders();
  const rows = (pos ?? [])
    .filter((po) => po.vendorId === vendorId)
    .slice(0, limit);

  const columns: DataListColumn<PurchaseOrder>[] = [
    {
      id: "documentNumber",
      header: "PO #",
      cell: (po) => (
        <span className="font-medium text-primary">{po.documentNumber}</span>
      ),
    },
    {
      id: "total",
      header: "Total",
      align: "right",
      priority: "md",
      cell: (po) => (
        <span className="tabular-nums">
          {formatCurrency(po.total ?? po.totalAmount ?? 0)}
        </span>
      ),
    },
    {
      id: "status",
      header: "Status",
      cell: (po) => <StatusBadge status={po.status} type="document" />,
    },
    {
      id: "delivery",
      header: "Delivery",
      priority: "lg",
      cell: (po) =>
        po.deliveryDate ? (
          <span className="text-sm text-muted-foreground">
            {new Date(po.deliveryDate).toLocaleDateString()}
          </span>
        ) : (
          <span className="text-muted-foreground">—</span>
        ),
    },
  ];

  return (
    <Card className="border-border/60">
      <CardHeader>
        <CardTitle className="text-base">Recent Purchase Orders</CardTitle>
      </CardHeader>
      <CardContent>
        <DataList<PurchaseOrder>
          rows={rows}
          columns={columns}
          getRowId={(po) => po.id}
          isLoading={isLoading}
          emptyMessage="No purchase orders for this vendor yet."
          onRowClick={(po) => router.push(`/purchase-orders/${po.id}`)}
          mobileCard={(po) => (
            <div className="flex flex-col gap-2">
              <div className="flex items-start justify-between gap-2">
                <div className="min-w-0">
                  <div className="font-medium text-primary line-clamp-1">
                    {po.documentNumber}
                  </div>
                </div>
                <StatusBadge status={po.status} type="document" />
              </div>
              <div className="flex flex-wrap items-center gap-2 text-xs text-muted-foreground">
                <span>
                  {formatCurrency(po.total ?? po.totalAmount ?? 0)}
                </span>
                {po.deliveryDate && (
                  <span>
                    {new Date(po.deliveryDate).toLocaleDateString()}
                  </span>
                )}
              </div>
            </div>
          )}
        />
      </CardContent>
    </Card>
  );
}
