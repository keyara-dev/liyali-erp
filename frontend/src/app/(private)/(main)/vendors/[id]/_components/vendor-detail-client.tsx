"use client";

import { useMemo } from "react";
import { useVendorById } from "@/hooks/use-vendor-queries";
import { usePurchaseOrders } from "@/hooks/use-purchase-order-queries";
import { DetailShell } from "@/components/layout/detail-shell";
import { PageHeader } from "@/components/base/page-header";
import { Skeleton } from "@/components/ui/skeleton";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { StatGrid } from "@/components/ui/stat-grid";
import {
  AlertCircle,
  ShoppingCart,
  CheckCircle2,
  Clock,
  CircleDollarSign,
} from "lucide-react";
import { formatCurrency } from "@/lib/utils";
import { VendorProfileCard } from "./vendor-profile-card";
import { VendorBankingCard } from "./vendor-banking-card";
import { VendorRecentPos } from "./vendor-recent-pos";

interface VendorDetailClientProps {
  vendorId: string;
}

export function VendorDetailClient({ vendorId }: VendorDetailClientProps) {
  const { data: vendor, isLoading, error } = useVendorById(vendorId);
  const { data: allPos } = usePurchaseOrders();

  const stats = useMemo(() => {
    const list = (allPos ?? []).filter((p) => p.vendorId === vendorId);
    const total = list.length;
    const approved = list.filter(
      (p) => p.status?.toUpperCase() === "APPROVED"
    ).length;
    const pending = list.filter((p) =>
      ["DRAFT", "PENDING", "PENDING_APPROVAL"].includes(
        p.status?.toUpperCase() ?? ""
      )
    ).length;
    const spend = list.reduce(
      (sum, p) => sum + Number(p.total ?? p.totalAmount ?? 0),
      0
    );
    return { total, approved, pending, spend };
  }, [allPos, vendorId]);

  if (isLoading) {
    return (
      <div className="space-y-5">
        <Skeleton className="h-12 w-full" />
        <Skeleton className="h-32 w-full" />
        <Skeleton className="h-72 w-full" />
      </div>
    );
  }

  if (error || !vendor) {
    return (
      <Alert variant="destructive">
        <AlertCircle className="h-4 w-4" />
        <AlertDescription>
          Failed to load vendor. The vendor may have been deleted.
        </AlertDescription>
      </Alert>
    );
  }

  return (
    <DetailShell
      header={
        <PageHeader
          title={vendor.name}
          subtitle={`Vendor code: ${vendor.vendorCode}`}
          showBackButton
        />
      }
      sidebar={
        <div className="space-y-5">
          <StatGrid
            items={[
              {
                label: "Total POs",
                value: stats.total,
                icon: <ShoppingCart className="h-3 w-3 sm:h-4 sm:w-4" />,
                accent: "blue",
              },
              {
                label: "Approved",
                value: stats.approved,
                icon: <CheckCircle2 className="h-3 w-3 sm:h-4 sm:w-4" />,
                accent: "emerald",
              },
              {
                label: "Pending",
                value: stats.pending,
                icon: <Clock className="h-3 w-3 sm:h-4 sm:w-4" />,
                accent: "amber",
              },
              {
                label: "Total spend",
                value: formatCurrency(stats.spend),
                icon: <CircleDollarSign className="h-3 w-3 sm:h-4 sm:w-4" />,
                accent: "violet",
              },
            ]}
          />
        </div>
      }
    >
      <div className="space-y-5">
        <VendorProfileCard vendor={vendor} />
        <VendorBankingCard vendor={vendor} />
        <VendorRecentPos vendorId={vendorId} />
      </div>
    </DetailShell>
  );
}
