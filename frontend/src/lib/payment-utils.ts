/**
 * Payment-related utility functions
 * Separated from server actions to avoid "use server" constraints
 */

import type { LinkedPVSummary } from "@/types/purchase-order";

/**
 * Whether a purchase order already has a *live* payment voucher and therefore
 * cannot get a new one. Mirrors the backend one-live-PV-per-PO gate, which
 * blocks creation when a linked PV exists whose status is NOT a terminal-failure
 * state (CANCELLED / REJECTED). Used to hide/disable "Create PV" for POs the
 * backend would otherwise reject with a 409.
 */
export function hasBlockingPaymentVoucher(po: {
  linkedPV?: LinkedPVSummary | null;
}): boolean {
  const pv = po.linkedPV;
  if (!pv) return false;
  const status = (pv.status ?? "").toUpperCase();
  return status !== "CANCELLED" && status !== "REJECTED";
}

/**
 * Remaining balance available for new payment vouchers on a purchase order.
 * Prefers the backend-computed `balance` field; falls back to
 * `totalAmount - amountCommitted` when `balance` isn't present (e.g. stale
 * cached PO objects predating multi-PV support).
 */
export function poRemainingBalance(po: {
  totalAmount?: number;
  amountCommitted?: number;
  balance?: number;
}): number {
  return typeof po.balance === "number"
    ? po.balance
    : (po.totalAmount ?? 0) - (po.amountCommitted ?? 0);
}

/**
 * Whether a purchase order still has room for another payment voucher.
 * Replaces `hasBlockingPaymentVoucher` as the gate for "Create PV" now that a
 * PO can carry multiple PVs capped at its remaining balance instead of just one.
 */
export function canCreateAnotherPV(po: {
  totalAmount?: number;
  amountCommitted?: number;
  balance?: number;
}): boolean {
  return poRemainingBalance(po) > 0.01;
}

export function generatePaymentReference(): string {
  const year = new Date().getFullYear();
  const month = String(new Date().getMonth() + 1).padStart(2, '0');
  const random = Math.random().toString(36).substring(2, 8).toUpperCase();
  return `PV-${year}${month}-${random}`;
}
