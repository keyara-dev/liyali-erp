"use client";

import { useCallback } from "react";
import { Plus, Trash2 } from "lucide-react";
import type { POItem } from "@/types/purchase-order";
import type { RequisitionItem } from "@/types/requisition";
import {
  computeLineItemVariance,
  lineItemVarianceColorClass,
} from "@/app/(private)/(main)/purchase-orders/_components/po-creation-wizard/types";

export interface POItemsEditableTableProps {
  items: POItem[];
  /** Controlled — every add / edit / remove reports the full next list. */
  onItemsChange: (items: POItem[]) => void;
  currency: string;
  /** Optional REQ items for the REQ Est. Price + Variance columns (matched by index). */
  reqItems?: RequisitionItem[];
  /** Factors taxRate / deliveryCost into the gradient total when present. */
  metadata?: Record<string, unknown>;
  /** Hide the gradient total footer (e.g. when the parent renders its own). */
  showSummary?: boolean;
}

export function newPOItem(): POItem {
  return {
    id: `tmp-${Math.round(performance.now() * 1000)}`,
    description: "",
    quantity: 1,
    unitPrice: 0,
    amount: 0,
  };
}

/**
 * Controlled, presentation-only PO line-item table: editable rows, add/remove,
 * optional REQ price-variance columns, and a gradient total footer. Holds no
 * state and makes no network calls — the parent owns `items` and persistence.
 *
 * Shared by the PO detail-page editor (POItemsEditor) and the PO creation
 * wizard's Line Items step.
 */
export function POItemsEditableTable({
  items,
  onItemsChange,
  currency,
  reqItems,
  metadata,
  showSummary = true,
}: POItemsEditableTableProps) {
  const hasReqItems = reqItems !== undefined && reqItems.length > 0;

  // Items-only subtotal (live).
  const itemsSubtotal = items.reduce(
    (sum, item) => sum + item.quantity * item.unitPrice,
    0,
  );

  // Factor in tax / delivery from metadata for the grand total display.
  const metaTaxRate = metadata?.taxRate
    ? parseFloat(String(metadata.taxRate))
    : 0;
  const metaDeliveryCost = metadata?.deliveryCost
    ? parseFloat(String(metadata.deliveryCost))
    : 0;
  const metaTaxAmount =
    !isNaN(metaTaxRate) && metaTaxRate > 0
      ? Math.round(((itemsSubtotal * metaTaxRate) / 100) * 100) / 100
      : 0;
  const metaDeliveryCostValue = !isNaN(metaDeliveryCost) ? metaDeliveryCost : 0;
  const grandTotal = itemsSubtotal + metaTaxAmount + metaDeliveryCostValue;
  const hasTaxOrDelivery = metaTaxAmount > 0 || metaDeliveryCostValue > 0;

  // ── mutations (controlled) ───────────────────────────────────────────────────

  const handleAddItem = () => {
    onItemsChange([...items, newPOItem()]);
  };

  const handleRemoveItem = (id: string) => {
    onItemsChange(items.filter((i) => i.id !== id));
  };

  const handleUpdateItem = useCallback(
    (id: string, field: keyof POItem, value: string | number) => {
      onItemsChange(
        items.map((item) => {
          if (item.id !== id) return item;
          const updated = { ...item, [field]: value };
          if (field === "quantity" || field === "unitPrice") {
            updated.amount = updated.quantity * updated.unitPrice;
            updated.totalPrice = updated.amount;
          }
          return updated;
        }),
      );
    },
    [items, onItemsChange],
  );

  // One track per column. With REQ items: # · Description · Qty · Unit · Total ·
  // REQ Est · Variance · Delete (8). Without: drop REQ Est + Variance (6).
  const gridCols = hasReqItems
    ? "grid-cols-[2rem_minmax(11rem,1fr)_3.25rem_7rem_6.5rem_6.5rem_5.5rem_2.25rem]"
    : "grid-cols-[2rem_minmax(11rem,1fr)_3.25rem_7rem_6.5rem_2.25rem]";

  return (
    <div className="space-y-3">
      <div className="rounded-lg border border-border overflow-hidden">
        <div className="overflow-x-auto">
          {/* Column headers */}
          <div
            className={`grid ${gridCols} gap-x-3 px-3 py-2 bg-muted/60 border-b border-border`}
          >
            <span className="text-xs font-medium text-muted-foreground">#</span>
            <span className="text-xs font-medium text-muted-foreground uppercase tracking-wider">
              Description
            </span>
            <span className="text-xs font-medium text-muted-foreground uppercase tracking-wider text-center">
              Qty
            </span>
            <span className="text-xs font-medium text-muted-foreground uppercase tracking-wider text-right">
              Unit Price ({currency})
            </span>
            <span className="text-xs font-medium text-muted-foreground uppercase tracking-wider text-right">
              Total
            </span>
            {hasReqItems && (
              <>
                <span className="text-xs font-medium text-muted-foreground uppercase tracking-wider text-right">
                  REQ Est. Price
                </span>
                <span className="text-xs font-medium text-muted-foreground uppercase tracking-wider text-right">
                  Variance
                </span>
              </>
            )}
            <span />
          </div>

          {/* Item rows */}
          <div className="divide-y divide-border/60">
            {items.map((item, index) => {
              const reqItem = reqItems?.[index];
              const reqEstPrice = reqItem?.unitPrice ?? 0;

              let varianceLabel: string | null = null;
              let varianceColorCls = "text-muted-foreground";
              if (hasReqItems) {
                if (!reqItem || reqEstPrice === 0) {
                  varianceLabel = null;
                } else {
                  const variance = computeLineItemVariance(
                    item.unitPrice,
                    reqEstPrice,
                  );
                  varianceColorCls = lineItemVarianceColorClass(
                    variance,
                    reqEstPrice,
                  );
                  const pct = Math.abs((variance / reqEstPrice) * 100).toFixed(1);
                  if (variance > 0) varianceLabel = `▲ +${pct}%`;
                  else if (variance < 0) varianceLabel = `▼ -${pct}%`;
                  else varianceLabel = `= 0%`;
                }
              }

              return (
                <div
                  key={item.id}
                  className={`grid ${gridCols} gap-x-3 px-3 py-2 items-center hover:bg-muted/20 transition-colors`}
                >
                  {/* # */}
                  <span className="text-xs text-muted-foreground/50 font-mono tabular-nums">
                    {String(index + 1).padStart(2, "0")}
                  </span>

                  {/* Description */}
                  <input
                    className="min-w-0 w-full bg-transparent text-sm placeholder:text-muted-foreground/40 focus:outline-none focus:bg-muted/30 rounded px-1.5 py-1 border border-transparent focus:border-border transition-colors"
                    placeholder="Item description…"
                    value={item.description}
                    onChange={(e) =>
                      handleUpdateItem(item.id!, "description", e.target.value)
                    }
                  />

                  {/* Qty */}
                  <input
                    type="number"
                    min="1"
                    className="w-full bg-transparent text-sm text-center tabular-nums placeholder:text-muted-foreground/40 focus:outline-none focus:bg-muted/30 rounded px-1 py-1 border border-transparent focus:border-border transition-colors"
                    value={item.quantity}
                    onChange={(e) =>
                      handleUpdateItem(
                        item.id!,
                        "quantity",
                        parseInt(e.target.value) || 1,
                      )
                    }
                  />

                  {/* Unit price */}
                  <input
                    type="number"
                    min="0"
                    step="0.01"
                    placeholder="0.00"
                    className="w-full bg-transparent text-sm text-right tabular-nums placeholder:text-muted-foreground/40 focus:outline-none focus:bg-muted/30 rounded px-1.5 py-1 border border-transparent focus:border-border transition-colors"
                    value={item.unitPrice || ""}
                    onChange={(e) =>
                      handleUpdateItem(
                        item.id!,
                        "unitPrice",
                        parseFloat(e.target.value) || 0,
                      )
                    }
                  />

                  {/* Line total */}
                  <span className="text-sm font-semibold text-right tabular-nums">
                    {(item.quantity * item.unitPrice).toLocaleString("en-ZM", {
                      minimumFractionDigits: 2,
                      maximumFractionDigits: 2,
                    })}
                  </span>

                  {/* REQ Est. Price */}
                  {hasReqItems && (
                    <span className="text-sm text-right tabular-nums text-muted-foreground">
                      {reqItem
                        ? reqEstPrice.toLocaleString("en-ZM", {
                            minimumFractionDigits: 2,
                            maximumFractionDigits: 2,
                          })
                        : "—"}
                    </span>
                  )}

                  {/* Variance */}
                  {hasReqItems && (
                    <span
                      className={`text-xs font-medium text-right tabular-nums ${varianceLabel !== null ? varianceColorCls : "text-muted-foreground"}`}
                    >
                      {varianceLabel !== null ? varianceLabel : "—"}
                    </span>
                  )}

                  {/* Delete */}
                  <button
                    type="button"
                    onClick={() => handleRemoveItem(item.id!)}
                    aria-label="Remove item"
                    className="size-7 rounded-md bg-destructive/5 text-muted-foreground/50 hover:bg-destructive/10 hover:text-destructive transition-colors flex items-center justify-center"
                  >
                    <Trash2 className="h-3.5 w-3.5" />
                  </button>
                </div>
              );
            })}
          </div>
        </div>

        {/* Add item row — full width, outside the horizontal scroll */}
        <button
          type="button"
          onClick={handleAddItem}
          className="w-full flex items-center gap-2 px-3 py-2.5 text-sm text-muted-foreground hover:text-foreground hover:bg-muted/30 border-t border-border/60 transition-colors"
        >
          <Plus className="h-3.5 w-3.5" />
          Add item
        </button>
      </div>

      {/* Summary — gradient footer */}
      {showSummary && items.length > 0 && (
        <div className="gradient-primary rounded-lg p-4">
          {hasTaxOrDelivery ? (
            <div className="space-y-1.5 text-sm text-white">
              <div className="flex justify-between">
                <span className="opacity-80">Items Subtotal</span>
                <span className="font-mono">
                  {currency}{" "}
                  {itemsSubtotal.toLocaleString("en-ZM", {
                    minimumFractionDigits: 2,
                    maximumFractionDigits: 2,
                  })}
                </span>
              </div>
              {metaTaxAmount > 0 && (
                <div className="flex justify-between">
                  <span className="opacity-80">Tax ({metaTaxRate}%)</span>
                  <span className="font-mono">
                    {currency}{" "}
                    {metaTaxAmount.toLocaleString("en-ZM", {
                      minimumFractionDigits: 2,
                      maximumFractionDigits: 2,
                    })}
                  </span>
                </div>
              )}
              {metaDeliveryCostValue > 0 && (
                <div className="flex justify-between">
                  <span className="opacity-80">Delivery Cost</span>
                  <span className="font-mono">
                    {currency}{" "}
                    {metaDeliveryCostValue.toLocaleString("en-ZM", {
                      minimumFractionDigits: 2,
                      maximumFractionDigits: 2,
                    })}
                  </span>
                </div>
              )}
              <div className="flex items-center justify-between border-t border-white/30 pt-1.5">
                <span className="font-semibold">Total Amount</span>
                <span className="text-2xl font-bold tracking-tight">
                  {currency}{" "}
                  {grandTotal.toLocaleString("en-ZM", {
                    minimumFractionDigits: 2,
                    maximumFractionDigits: 2,
                  })}
                </span>
              </div>
            </div>
          ) : (
            <div className="flex items-center justify-between">
              <span className="font-semibold text-white">Total Amount</span>
              <span className="text-2xl font-bold text-white tracking-tight">
                {currency}{" "}
                {itemsSubtotal.toLocaleString("en-ZM", {
                  minimumFractionDigits: 2,
                  maximumFractionDigits: 2,
                })}
              </span>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
