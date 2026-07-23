"use client";

import { Info, Package, CreditCard, ArrowRight } from "lucide-react";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { cn } from "@/lib/utils";

/**
 * ProcurementFlowInfo — an info-icon trigger that opens a popover explaining the
 * two procurement flows (goods-first vs payment-first), the document chain each
 * produces, and what the system auto-creates on PO approval.
 *
 * Used next to the procurement-flow selector in both the PO creation wizard
 * (per-PO override) and workspace settings (org default) so users understand
 * the downstream effect of their choice.
 */
export function ProcurementFlowInfo({ className }: { className?: string }) {
  return (
    <Popover>
      <PopoverTrigger asChild>
        <button
          type="button"
          aria-label="How procurement flows work"
          className={cn(
            "inline-flex items-center justify-center rounded-full text-muted-foreground hover:text-foreground transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring",
            className,
          )}
        >
          <Info className="h-4 w-4" />
        </button>
      </PopoverTrigger>
      <PopoverContent align="start" className="w-80 space-y-4">
        <div>
          <p className="text-sm font-semibold">How procurement flows work</p>
          <p className="text-xs text-muted-foreground">
            The flow decides the order of documents after a PO is approved — and
            what the system auto-creates for you.
          </p>
        </div>

        {/* Goods-First */}
        <FlowExplainer
          icon={<Package className="h-4 w-4" />}
          accent="text-emerald-600 dark:text-emerald-400"
          title="Goods-First"
          tagline="Receive before you pay."
          chain={["PO", "GRN", "PV", "Payment"]}
          highlight="GRN"
          points={[
            "On PO approval, a draft GRN is created for the receiver to record goods and sign.",
            "Payment is blocked until that GRN is approved — protects against paying for goods you never got.",
            "Best for government, high-value orders, or new vendors.",
          ]}
        />

        {/* Payment-First */}
        <FlowExplainer
          icon={<CreditCard className="h-4 w-4" />}
          accent="text-amber-600 dark:text-amber-400"
          title="Payment-First"
          tagline="Pay upfront, confirm delivery later."
          chain={["PO", "PV", "Payment", "GRN"]}
          highlight="PV"
          points={[
            "On PO approval, a draft Payment Voucher is created so payment can be processed first.",
            "A GRN is recorded later, once goods arrive, to confirm receipt.",
            "Best for deposits, prepayments, or trusted recurring vendors.",
          ]}
        />
      </PopoverContent>
    </Popover>
  );
}

function FlowExplainer({
  icon,
  accent,
  title,
  tagline,
  chain,
  highlight,
  points,
}: {
  icon: React.ReactNode;
  accent: string;
  title: string;
  tagline: string;
  chain: string[];
  highlight: string;
  points: string[];
}) {
  return (
    <div className="space-y-1.5">
      <div className="flex items-center gap-1.5">
        <span className={accent}>{icon}</span>
        <span className="text-sm font-medium">{title}</span>
        <span className="text-xs text-muted-foreground">— {tagline}</span>
      </div>

      {/* Document chain */}
      <div className="flex flex-wrap items-center gap-1 text-[11px]">
        {chain.map((node, i) => (
          <span key={node} className="flex items-center gap-1">
            <span
              className={cn(
                "rounded px-1.5 py-0.5 font-medium",
                node === highlight
                  ? cn("bg-muted", accent)
                  : "bg-muted text-muted-foreground",
              )}
            >
              {node}
            </span>
            {i < chain.length - 1 && (
              <ArrowRight className="h-3 w-3 text-muted-foreground" />
            )}
          </span>
        ))}
      </div>

      <ul className="list-disc space-y-0.5 pl-4 text-xs text-muted-foreground">
        {points.map((p) => (
          <li key={p}>{p}</li>
        ))}
      </ul>
    </div>
  );
}
