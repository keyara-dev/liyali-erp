"use client";

import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Landmark, ShieldCheck } from "lucide-react";
import { cn } from "@/lib/utils";
import type { Vendor } from "@/types/core";

interface VendorBankingCardProps {
  vendor: Vendor;
  className?: string;
}

function Row({ label, value }: { label: string; value: React.ReactNode }) {
  return (
    <div className="flex items-baseline justify-between gap-4 py-2 border-b border-border/40 last:border-b-0">
      <span className="text-[11px] font-medium text-muted-foreground uppercase tracking-wider shrink-0">
        {label}
      </span>
      <span className="text-sm font-mono tabular-nums break-all text-right">
        {value || <span className="text-muted-foreground font-sans">—</span>}
      </span>
    </div>
  );
}

export function VendorBankingCard({ vendor, className }: VendorBankingCardProps) {
  return (
    <Card className={cn("border-border/60", className)}>
      <CardHeader className="pb-2">
        <CardTitle className="text-base flex items-center gap-2">
          <span className="flex h-7 w-7 items-center justify-center rounded-lg bg-muted text-muted-foreground">
            <Landmark className="h-4 w-4" aria-hidden="true" />
          </span>
          Banking
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <div>
          <Row label="Bank" value={vendor.bankName} />
          <Row label="Branch" value={vendor.branchCode} />
          <Row label="Account name" value={vendor.accountName} />
          <Row label="Account number" value={vendor.accountNumber} />
          <Row label="SWIFT" value={vendor.swiftCode} />
        </div>

        <div className="rounded-lg border border-border/60 bg-muted/30 px-3 py-2">
          <p className="mb-1 flex items-center gap-1.5 text-[11px] font-medium text-muted-foreground uppercase tracking-wider">
            <ShieldCheck className="h-3.5 w-3.5" aria-hidden="true" />
            Tax &amp; Compliance
          </p>
          <Row label="ZRA TPIN" value={vendor.zraTpin || vendor.taxId} />
          <Row label="PACRA Reg. No." value={vendor.pacraRegNumber} />
        </div>
      </CardContent>
    </Card>
  );
}
