import { TrendingDown, TrendingUp, Minus } from "lucide-react";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { cn } from "@/lib/utils";
import { formatCurrency } from "@/lib/utils";
import {
  computeVariance,
  varianceColorClass,
} from "@/app/(private)/(main)/purchase-orders/_components/po-creation-wizard/types";

interface CostComparisonPanelProps {
  estimatedCost: number;
  currency: string;
  vendors: Array<{
    vendorId: string;
    vendorName: string;
    quotedAmount?: number; // undefined = no quotation yet
    isSelected?: boolean;
  }>;
}

export function CostComparisonPanel({
  estimatedCost,
  currency,
  vendors,
}: CostComparisonPanelProps) {
  const rows = vendors.map((vendor) => {
    const hasQuote = vendor.quotedAmount !== undefined;
    const variance = hasQuote
      ? computeVariance(estimatedCost, vendor.quotedAmount!)
      : null;
    const colorClass = variance
      ? varianceColorClass(variance.absolute, variance.percentage)
      : "";
    const isUnder = variance ? variance.absolute < 0 : false;
    const isOver = variance ? variance.absolute > 0 : false;
    const VarianceIcon = isUnder
      ? TrendingDown
      : isOver
        ? TrendingUp
        : Minus;

    return {
      vendor,
      hasQuote,
      variance,
      colorClass,
      isUnder,
      isOver,
      VarianceIcon,
    };
  });

  return (
    <div className="min-w-0 rounded-lg border border-blue-200 dark:border-blue-800 bg-blue-50 dark:bg-blue-950/30 p-3 space-y-3">
      <h4 className="text-xs font-semibold text-blue-900 dark:text-blue-100 uppercase tracking-wider">
        Cost Comparison
      </h4>

      {/* Mobile: stacked card layout */}
      <div className="space-y-2 sm:hidden">
        {rows.map(
          ({
            vendor,
            hasQuote,
            variance,
            colorClass,
            isUnder,
            isOver,
            VarianceIcon,
          }) => (
            <div
              key={vendor.vendorId}
              className={cn(
                "rounded-md border border-blue-200 dark:border-blue-800 p-2.5 space-y-1.5",
                vendor.isSelected && "bg-blue-100/60 dark:bg-blue-900/30",
              )}
            >
              <div className="flex items-center justify-between gap-2">
                <span className="text-sm font-medium text-blue-900 dark:text-blue-100 truncate">
                  {vendor.vendorName}
                </span>
                {vendor.isSelected && (
                  <span className="text-[10px] font-semibold uppercase tracking-wider text-blue-600 dark:text-blue-400 shrink-0">
                    Selected
                  </span>
                )}
              </div>
              <div className="grid grid-cols-2 gap-x-3 gap-y-1 text-xs">
                <span className="text-blue-700/80 dark:text-blue-300/80">
                  Est. Cost
                </span>
                <span className="text-blue-900 dark:text-blue-100 font-medium text-right tabular-nums">
                  {formatCurrency(estimatedCost, currency)}
                </span>
                <span className="text-blue-700/80 dark:text-blue-300/80">
                  Quoted
                </span>
                <span className="text-blue-900 dark:text-blue-100 text-right tabular-nums">
                  {hasQuote
                    ? formatCurrency(vendor.quotedAmount!, currency)
                    : "—"}
                </span>
                <span className="text-blue-700/80 dark:text-blue-300/80">
                  Variance
                </span>
                <span className="text-right">
                  {variance ? (
                    <span
                      className={cn(
                        "inline-flex items-center gap-1 font-medium tabular-nums",
                        colorClass,
                      )}
                    >
                      <VarianceIcon className="h-3 w-3 shrink-0" />
                      {isUnder ? "−" : isOver ? "+" : ""}
                      {formatCurrency(Math.abs(variance.absolute), currency)}
                      <span className="font-normal">
                        ({isUnder ? "−" : isOver ? "+" : ""}
                        {Math.abs(variance.percentage).toFixed(1)}%)
                      </span>
                    </span>
                  ) : (
                    <span className="text-muted-foreground">—</span>
                  )}
                </span>
              </div>
            </div>
          ),
        )}
      </div>

      {/* Desktop: table layout, horizontal scroll contained within panel */}
      <div className="hidden sm:block overflow-x-auto">
        <Table>
          <TableHeader>
            <TableRow className="border-blue-200 dark:border-blue-800">
              <TableHead className="text-blue-700 dark:text-blue-300 text-xs font-semibold">
                Vendor
              </TableHead>
              <TableHead className="text-blue-700 dark:text-blue-300 text-xs font-semibold whitespace-nowrap">
                Est. Cost (REQ)
              </TableHead>
              <TableHead className="text-blue-700 dark:text-blue-300 text-xs font-semibold whitespace-nowrap">
                Quoted Price
              </TableHead>
              <TableHead className="text-blue-700 dark:text-blue-300 text-xs font-semibold">
                Variance
              </TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {rows.map(
              ({
                vendor,
                hasQuote,
                variance,
                colorClass,
                isUnder,
                isOver,
                VarianceIcon,
              }) => (
                <TableRow
                  key={vendor.vendorId}
                  className={cn(
                    "border-blue-200 dark:border-blue-800",
                    vendor.isSelected &&
                      "bg-blue-100/60 dark:bg-blue-900/30 font-medium",
                  )}
                >
                  <TableCell className="text-blue-900 dark:text-blue-100 text-sm">
                    {vendor.vendorName}
                    {vendor.isSelected && (
                      <span className="ml-2 text-xs text-blue-600 dark:text-blue-400 font-normal">
                        (selected)
                      </span>
                    )}
                  </TableCell>
                  <TableCell className="text-blue-900 dark:text-blue-100 text-sm font-medium whitespace-nowrap">
                    {formatCurrency(estimatedCost, currency)}
                  </TableCell>
                  <TableCell className="text-blue-900 dark:text-blue-100 text-sm whitespace-nowrap">
                    {hasQuote
                      ? formatCurrency(vendor.quotedAmount!, currency)
                      : "—"}
                  </TableCell>
                  <TableCell className="text-sm">
                    {variance ? (
                      <span
                        className={cn(
                          "flex items-center gap-1 font-medium whitespace-nowrap",
                          colorClass,
                        )}
                      >
                        <VarianceIcon className="h-3.5 w-3.5 shrink-0" />
                        {isUnder ? "−" : isOver ? "+" : ""}
                        {formatCurrency(Math.abs(variance.absolute), currency)}
                        <span className="font-normal text-xs">
                          ({isUnder ? "−" : isOver ? "+" : ""}
                          {Math.abs(variance.percentage).toFixed(1)}%)
                        </span>
                      </span>
                    ) : (
                      <span className="text-muted-foreground">—</span>
                    )}
                  </TableCell>
                </TableRow>
              ),
            )}
          </TableBody>
        </Table>
      </div>
    </div>
  );
}
