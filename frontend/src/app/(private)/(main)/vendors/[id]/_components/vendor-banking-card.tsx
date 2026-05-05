"use client";

import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Landmark } from "lucide-react";
import type { Vendor } from "@/types/core";

interface VendorBankingCardProps {
  vendor: Vendor;
}

function Row({ label, value }: { label: string; value: React.ReactNode }) {
  return (
    <div className="flex justify-between gap-4 py-2 border-b border-border/40 last:border-b-0">
      <span className="text-xs font-medium text-muted-foreground uppercase tracking-wider">
        {label}
      </span>
      <span className="text-sm font-mono tabular-nums break-all text-right">
        {value || "—"}
      </span>
    </div>
  );
}

export function VendorBankingCard({ vendor }: VendorBankingCardProps) {
  return (
    <Card className="border-border/60">
      <CardHeader>
        <CardTitle className="text-base flex items-center gap-2">
          <Landmark className="h-4 w-4" aria-hidden="true" />
          Banking
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-1">
          <Row label="Bank" value={vendor.bankName} />
          <Row label="Branch" value={vendor.branchCode} />
          <Row label="Account name" value={vendor.accountName} />
          <Row label="Account number" value={vendor.accountNumber} />
          <Row label="SWIFT" value={vendor.swiftCode} />
          <Row label="Tax ID" value={vendor.taxId} />
        </div>
      </CardContent>
    </Card>
  );
}
