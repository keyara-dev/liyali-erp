"use client";

import { useState } from "react";
import { Loader2, Check, X } from "lucide-react";
import { Button } from "@/components/ui/button";
import { toast } from "sonner";
import { updatePurchaseOrderItems } from "@/app/_actions/purchase-orders";
import { useQueryClient } from "@tanstack/react-query";
import { QUERY_KEYS } from "@/lib/constants";
import type { POItem } from "@/types/purchase-order";
import type { RequisitionItem } from "@/types/requisition";
import {
  POItemsEditableTable,
  newPOItem,
} from "./po-items-editable-table";

interface POItemsEditorProps {
  poId: string;
  items: POItem[];
  currency: string;
  /** PO metadata — used to factor in existing taxRate and deliveryCost when computing totalAmount */
  metadata?: Record<string, unknown>;
  /** Optional REQ items for price comparison — matched by array index */
  reqItems?: RequisitionItem[];
  onSaved: (updatedItems: POItem[]) => void;
  onCancel: () => void;
}

/**
 * Inline item editor for the PO detail page. Wraps the shared
 * POItemsEditableTable with Save/Cancel actions and persistence: on save it
 * calls updatePurchaseOrderItems — the backend records a full audit snapshot
 * with old/new item values.
 */
export function POItemsEditor({
  poId,
  items: initialItems,
  currency,
  metadata,
  onSaved,
  onCancel,
  reqItems,
}: POItemsEditorProps) {
  const [items, setItems] = useState<POItem[]>(
    initialItems.length > 0
      ? initialItems.map((i) => ({ ...i, id: i.id || newPOItem().id }))
      : [newPOItem()],
  );
  const [saving, setSaving] = useState(false);
  const queryClient = useQueryClient();

  // ── save ───────────────────────────────────────────────────────────────────

  const handleSave = async () => {
    if (items.length === 0) {
      toast.error("Add at least one item");
      return;
    }
    const invalid = items.some(
      (i) => !i.description.trim() || i.quantity <= 0 || i.unitPrice < 0,
    );
    if (invalid) {
      toast.error(
        "All items need a description, quantity > 0, and unit price >= 0",
      );
      return;
    }

    const finalItems = items.map((i) => ({
      ...i,
      amount: i.quantity * i.unitPrice,
      totalPrice: i.quantity * i.unitPrice,
    }));
    const itemsSubtotal = finalItems.reduce((s, i) => s + i.amount, 0);

    // Factor in existing tax and delivery from PO metadata so totalAmount
    // stays consistent with what the Shipping & Tax tab has set.
    const taxRate = metadata?.taxRate
      ? parseFloat(String(metadata.taxRate))
      : 0;
    const deliveryCost = metadata?.deliveryCost
      ? parseFloat(String(metadata.deliveryCost))
      : 0;
    const taxAmount =
      !isNaN(taxRate) && taxRate > 0
        ? Math.round(((itemsSubtotal * taxRate) / 100) * 100) / 100
        : 0;
    const deliveryCostValue = !isNaN(deliveryCost) ? deliveryCost : 0;
    const total = itemsSubtotal + taxAmount + deliveryCostValue;

    setSaving(true);
    try {
      const result = await updatePurchaseOrderItems(poId, finalItems, total);
      if (!result.success) throw new Error(result.message || "Failed to save");

      // Refetch immediately so the header total updates without waiting
      await queryClient.refetchQueries({
        queryKey: [QUERY_KEYS.PURCHASE_ORDERS.BY_ID, poId],
      });
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.PURCHASE_ORDERS.ALL],
      });
      queryClient.invalidateQueries({
        queryKey: ["audit-events", "purchase_order", poId],
      });
      toast.success("Items updated");
      onSaved(finalItems);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to save items");
    } finally {
      setSaving(false);
    }
  };

  // ── render ─────────────────────────────────────────────────────────────────

  return (
    <div className="space-y-3">
      <POItemsEditableTable
        items={items}
        onItemsChange={setItems}
        currency={currency}
        reqItems={reqItems}
        metadata={metadata}
      />

      {/* Actions */}
      <div className="flex justify-end gap-2 pt-1">
        <Button
          type="button"
          variant="ghost"
          size="sm"
          onClick={onCancel}
          disabled={saving}
        >
          <X className="h-3.5 w-3.5 mr-1" />
          Cancel
        </Button>
        <Button type="button" size="sm" onClick={handleSave} disabled={saving}>
          {saving ? (
            <Loader2 className="h-3.5 w-3.5 mr-1 animate-spin" />
          ) : (
            <Check className="h-3.5 w-3.5 mr-1" />
          )}
          Save Items
        </Button>
      </div>
    </div>
  );
}
