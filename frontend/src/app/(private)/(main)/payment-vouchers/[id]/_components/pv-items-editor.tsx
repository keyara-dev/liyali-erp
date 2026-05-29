"use client";

import { useEffect, useMemo, useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { Plus, Trash2, Save, RotateCcw } from "lucide-react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { formatCurrency } from "@/lib/utils";
import { QUERY_KEYS } from "@/lib/constants";
import { updatePaymentVoucher } from "@/app/_actions/payment-vouchers";
import type { PaymentItem } from "@/types/payment-voucher";

/**
 * Per-row state used by the editor. The `_key` field is a stable client-side
 * key for React; it is stripped before sending to the backend.
 */
interface ItemRow extends PaymentItem {
  _key: string;
}

const newRow = (item?: Partial<PaymentItem>): ItemRow => ({
  _key:
    typeof crypto !== "undefined" && "randomUUID" in crypto
      ? crypto.randomUUID()
      : `pv-item-${Date.now()}-${Math.random().toString(36).slice(2)}`,
  description: item?.description ?? "",
  amount: item?.amount ?? 0,
  glCode: item?.glCode ?? "",
  taxAmount: item?.taxAmount ?? 0,
});

const seedRows = (items: PaymentItem[] | undefined): ItemRow[] =>
  (items ?? []).map((it) => newRow(it));

interface PVItemsEditorProps {
  /** Payment voucher ID. */
  pvId: string;
  /** Current items on the PV. */
  items: PaymentItem[] | undefined;
  /** Currency code (used for display only — the totals are saved as numbers). */
  currency: string;
  /** ID of the user performing the edit. */
  userId: string;
}

/**
 * Editable line-item table for DRAFT payment vouchers.
 *
 * Mirrors the styling of the GRN create-dialog item editor: a transparent
 * grid of inline inputs with a trailing trash button per row and an
 * "Add line" footer button. A sticky action bar appears at the bottom while
 * unsaved changes exist.
 *
 * Edits are persisted via {@link updatePaymentVoucher}, which the backend
 * accepts only when the PV is still DRAFT (it returns 403 otherwise).
 */
export function PVItemsEditor({
  pvId,
  items,
  currency,
  userId,
}: PVItemsEditorProps) {
  const queryClient = useQueryClient();
  const [rows, setRows] = useState<ItemRow[]>(() => seedRows(items));
  const [isSaving, setIsSaving] = useState(false);

  // Re-seed when the upstream items change (e.g. after a successful save the
  // detail query refetches and feeds new items back through props).
  useEffect(() => {
    setRows(seedRows(items));
  }, [items]);

  const originalSignature = useMemo(
    () => JSON.stringify(items ?? []),
    [items],
  );
  const currentSignature = useMemo(
    () =>
      JSON.stringify(
        rows.map(({ _key: _, ...rest }) => ({
          ...rest,
          amount: Number(rest.amount) || 0,
          taxAmount: Number(rest.taxAmount) || 0,
        })),
      ),
    [rows],
  );
  const isDirty = originalSignature !== currentSignature;

  const totals = useMemo(() => {
    let amount = 0;
    let tax = 0;
    for (const r of rows) {
      amount += Number(r.amount) || 0;
      tax += Number(r.taxAmount) || 0;
    }
    return { amount, tax };
  }, [rows]);

  const updateRow = <K extends keyof PaymentItem>(
    key: string,
    field: K,
    value: PaymentItem[K],
  ) => {
    setRows((prev) =>
      prev.map((r) => (r._key === key ? { ...r, [field]: value } : r)),
    );
  };

  const addRow = () => setRows((prev) => [...prev, newRow()]);
  const removeRow = (key: string) =>
    setRows((prev) => prev.filter((r) => r._key !== key));

  const handleReset = () => setRows(seedRows(items));

  const handleSave = async () => {
    // Strip empty trailing description rows entirely — they're noise.
    const cleaned = rows
      .filter((r) => r.description.trim() !== "" || (Number(r.amount) || 0) > 0)
      .map(({ _key: _, ...rest }) => ({
        description: rest.description.trim(),
        amount: Number(rest.amount) || 0,
        glCode: (rest.glCode ?? "").trim(),
        taxAmount: Number(rest.taxAmount) || 0,
      }));

    setIsSaving(true);
    try {
      const res = await updatePaymentVoucher({
        paymentVoucherId: pvId,
        pvId,
        items: cleaned,
        updatedBy: userId,
      });

      if (!res.success) {
        toast.error(res.message || "Failed to update payment voucher items");
        return;
      }

      toast.success("Payment voucher items saved");
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.PAYMENT_VOUCHERS.BY_ID, pvId],
      });
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.PAYMENT_VOUCHERS.ALL],
      });
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.PAYMENT_VOUCHERS.STATS],
      });
    } catch (err) {
      const message =
        err instanceof Error ? err.message : "Failed to update items";
      toast.error(message);
    } finally {
      setIsSaving(false);
    }
  };

  return (
    <div className="space-y-3">
      <Label className="text-sm font-semibold">Items</Label>

      <div className="rounded-lg border border-border overflow-hidden">
        {/* Column headers */}
        <div className="grid grid-cols-[1.75rem_1fr_7rem_6rem_6rem_1.75rem] gap-x-3 px-3 py-2 bg-muted/60 border-b border-border">
          <span className="text-xs font-medium text-muted-foreground">#</span>
          <span className="text-xs font-medium text-muted-foreground uppercase tracking-wider">
            Description
          </span>
          <span className="text-xs font-medium text-muted-foreground uppercase tracking-wider">
            GL Code
          </span>
          <span className="text-xs font-medium text-muted-foreground uppercase tracking-wider text-right">
            Amount
          </span>
          <span className="text-xs font-medium text-muted-foreground uppercase tracking-wider text-right">
            Tax
          </span>
          <span />
        </div>

        {/* Rows */}
        <div className="divide-y divide-border/60">
          {rows.length === 0 && (
            <div className="px-3 py-4 text-sm text-muted-foreground">
              No items yet — add a line below.
            </div>
          )}
          {rows.map((row, index) => (
            <div
              key={row._key}
              className="grid grid-cols-[1.75rem_1fr_7rem_6rem_6rem_1.75rem] gap-x-3 px-3 py-2 items-center hover:bg-muted/20 transition-colors"
            >
              <span className="text-xs text-muted-foreground/50 font-mono tabular-nums">
                {String(index + 1).padStart(2, "0")}
              </span>

              <input
                className="min-w-0 w-full bg-transparent text-sm placeholder:text-muted-foreground/40 focus:outline-none focus:bg-muted/30 rounded px-1.5 py-1 -mx-1.5 border border-transparent focus:border-border transition-colors"
                placeholder="Item description…"
                value={row.description}
                disabled={isSaving}
                onChange={(e) =>
                  updateRow(row._key, "description", e.target.value)
                }
              />

              <input
                className="min-w-0 w-full bg-transparent text-sm font-mono placeholder:text-muted-foreground/40 focus:outline-none focus:bg-muted/30 rounded px-1.5 py-1 -mx-1.5 border border-transparent focus:border-border transition-colors"
                placeholder="GL code"
                value={row.glCode ?? ""}
                disabled={isSaving}
                onChange={(e) => updateRow(row._key, "glCode", e.target.value)}
              />

              <input
                type="number"
                min={0}
                step="0.01"
                className="w-full bg-transparent text-sm text-right tabular-nums placeholder:text-muted-foreground/40 focus:outline-none focus:bg-muted/30 rounded px-1.5 py-1 -mx-1.5 border border-transparent focus:border-border transition-colors"
                value={Number.isFinite(row.amount) ? row.amount : 0}
                disabled={isSaving}
                onChange={(e) =>
                  updateRow(row._key, "amount", Number(e.target.value))
                }
              />

              <input
                type="number"
                min={0}
                step="0.01"
                className="w-full bg-transparent text-sm text-right tabular-nums placeholder:text-muted-foreground/40 focus:outline-none focus:bg-muted/30 rounded px-1.5 py-1 -mx-1.5 border border-transparent focus:border-border transition-colors"
                placeholder="0"
                value={Number.isFinite(row.taxAmount ?? 0) ? row.taxAmount : 0}
                disabled={isSaving}
                onChange={(e) =>
                  updateRow(row._key, "taxAmount", Number(e.target.value))
                }
              />

              <button
                type="button"
                onClick={() => removeRow(row._key)}
                disabled={isSaving}
                aria-label="Remove line"
                className="text-muted-foreground/40 hover:text-red-500 transition-colors flex items-center justify-center disabled:opacity-50"
              >
                <Trash2 className="h-3.5 w-3.5" />
              </button>
            </div>
          ))}

          {/* Add row */}
          <button
            type="button"
            onClick={addRow}
            disabled={isSaving}
            className="w-full flex items-center gap-2 px-3 py-2.5 text-sm text-muted-foreground hover:text-foreground hover:bg-muted/30 transition-colors disabled:opacity-50"
          >
            <Plus className="h-3.5 w-3.5" />
            Add line
          </button>
        </div>

        {/* Totals footer */}
        <div className="grid grid-cols-[1fr_auto_auto] gap-x-6 px-3 py-2 bg-muted/40 border-t border-border text-sm">
          <span className="text-xs text-muted-foreground">
            {rows.length} item{rows.length !== 1 ? "s" : ""}
          </span>
          {totals.tax > 0 && (
            <span className="text-xs text-muted-foreground tabular-nums">
              Tax: {formatCurrency(totals.tax, currency)}
            </span>
          )}
          <span className="font-semibold tabular-nums text-right">
            Total: {formatCurrency(totals.amount, currency)}
          </span>
        </div>
      </div>

      {/* Dirty-state action bar */}
      {isDirty && (
        <div className="flex items-center justify-between rounded-md border border-amber-200 bg-amber-50 dark:border-amber-800 dark:bg-amber-950/30 px-3 py-2">
          <p className="text-xs text-amber-900 dark:text-amber-200">
            Unsaved item changes — the headline amount will be recalculated
            from the line items on save.
          </p>
          <div className="flex items-center gap-2">
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={handleReset}
              disabled={isSaving}
            >
              <RotateCcw className="h-3.5 w-3.5 mr-1.5" />
              Reset
            </Button>
            <Button
              type="button"
              size="sm"
              onClick={handleSave}
              disabled={isSaving}
              isLoading={isSaving}
              loadingText="Saving…"
            >
              <Save className="h-3.5 w-3.5 mr-1.5" />
              Save items
            </Button>
          </div>
        </div>
      )}
    </div>
  );
}
