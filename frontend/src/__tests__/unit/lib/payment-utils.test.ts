import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import {
  generatePaymentReference,
  hasBlockingPaymentVoucher,
  poRemainingBalance,
  canCreateAnotherPV,
} from "@/lib/payment-utils";

/**
 * generatePaymentReference() → string
 *
 * Format: PV-{YYYY}{MM}-{6-char-alphanumeric-uppercase}
 * Example: PV-202603-ABC123
 */

describe("generatePaymentReference", () => {
  it("returns a string", () => {
    expect(typeof generatePaymentReference()).toBe("string");
  });

  it("starts with 'PV-'", () => {
    expect(generatePaymentReference()).toMatch(/^PV-/);
  });

  it("matches the expected format PV-YYYYMM-XXXXXX", () => {
    // Format: PV-<4-digit year><2-digit month>-<6 uppercase alphanumeric chars>
    expect(generatePaymentReference()).toMatch(
      /^PV-\d{4}(0[1-9]|1[0-2])-[0-9A-Z]{6}$/
    );
  });

  it("uses the current year in the reference", () => {
    const year = new Date().getFullYear().toString();
    expect(generatePaymentReference()).toContain(year);
  });

  it("uses the current month (zero-padded) in the reference", () => {
    const month = String(new Date().getMonth() + 1).padStart(2, "0");
    const ref = generatePaymentReference();
    // The month portion occupies characters 7-8 (after "PV-YYYY")
    const monthPart = ref.slice(7, 9);
    expect(monthPart).toBe(month);
  });

  it("produces 6 uppercase alphanumeric characters in the random segment", () => {
    const ref = generatePaymentReference();
    const randomPart = ref.split("-")[2];
    expect(randomPart).toHaveLength(6);
    expect(randomPart).toMatch(/^[0-9A-Z]{6}$/);
  });

  it("generates unique references on successive calls", () => {
    const refs = new Set(Array.from({ length: 20 }, () => generatePaymentReference()));
    // With 36^6 ≈ 2.1 billion combinations, collisions are astronomically rare
    expect(refs.size).toBeGreaterThan(1);
  });

  describe("with a fixed date (mocked)", () => {
    beforeEach(() => {
      // January 5, 2026
      vi.useFakeTimers();
      vi.setSystemTime(new Date("2026-01-05T10:00:00Z"));
    });

    afterEach(() => {
      vi.useRealTimers();
    });

    it("embeds year 2026 and month 01 when date is January 2026", () => {
      const ref = generatePaymentReference();
      expect(ref).toMatch(/^PV-202601-/);
    });
  });

  describe("with a fixed date — December (month 12)", () => {
    beforeEach(() => {
      vi.useFakeTimers();
      vi.setSystemTime(new Date("2025-12-15T00:00:00Z"));
    });

    afterEach(() => {
      vi.useRealTimers();
    });

    it("embeds year 2025 and month 12 when date is December 2025", () => {
      const ref = generatePaymentReference();
      expect(ref).toMatch(/^PV-202512-/);
    });
  });
});

/**
 * hasBlockingPaymentVoucher(po) → boolean
 *
 * Mirrors the backend one-live-PV-per-PO gate: a PO already has a *live* PV
 * (and so cannot get a new one) when a linked PV exists whose status is NOT a
 * terminal-failure state (CANCELLED / REJECTED).
 */
describe("hasBlockingPaymentVoucher", () => {
  const po = (status?: string) =>
    status === undefined
      ? {}
      : { linkedPV: { id: "pv-1", documentNumber: "PV-1", status } };

  it("returns false when there is no linked PV", () => {
    expect(hasBlockingPaymentVoucher(po())).toBe(false);
  });

  it.each(["PAID", "APPROVED", "PENDING", "DRAFT", "paid", "Approved"])(
    "blocks when a live PV exists (status %s)",
    (status) => {
      expect(hasBlockingPaymentVoucher(po(status))).toBe(true);
    },
  );

  it.each(["CANCELLED", "REJECTED", "cancelled", "rejected"])(
    "does not block when the only PV is terminal-failure (status %s)",
    (status) => {
      expect(hasBlockingPaymentVoucher(po(status))).toBe(false);
    },
  );
});

/**
 * poRemainingBalance(po) → number
 *
 * Prefers the backend-computed `balance` field; falls back to
 * `totalAmount - amountCommitted` when `balance` is absent.
 */
describe("poRemainingBalance", () => {
  it("returns `balance` when present, even when it disagrees with totalAmount - amountCommitted", () => {
    expect(
      poRemainingBalance({
        totalAmount: 1000,
        amountCommitted: 200,
        balance: 750,
      }),
    ).toBe(750);
  });

  it("returns 0 balance as-is (not falling back), since 0 is a valid number", () => {
    expect(
      poRemainingBalance({ totalAmount: 1000, amountCommitted: 1000, balance: 0 }),
    ).toBe(0);
  });

  it("falls back to totalAmount - amountCommitted when balance is undefined", () => {
    expect(
      poRemainingBalance({ totalAmount: 1000, amountCommitted: 300 }),
    ).toBe(700);
  });

  it("treats missing totalAmount/amountCommitted as 0", () => {
    expect(poRemainingBalance({})).toBe(0);
  });

  it("can go negative when amountCommitted exceeds totalAmount (fallback path)", () => {
    expect(
      poRemainingBalance({ totalAmount: 500, amountCommitted: 600 }),
    ).toBe(-100);
  });
});

/**
 * canCreateAnotherPV(po) → boolean
 *
 * True when poRemainingBalance(po) > 0.01 — replaces hasBlockingPaymentVoucher
 * as the "Create PV" gate now that a PO can carry multiple PVs.
 */
describe("canCreateAnotherPV", () => {
  it("returns true when there is meaningful remaining balance", () => {
    expect(canCreateAnotherPV({ totalAmount: 1000, amountCommitted: 400 })).toBe(
      true,
    );
  });

  it("returns false when fully committed (balance = 0)", () => {
    expect(canCreateAnotherPV({ totalAmount: 1000, amountCommitted: 1000 })).toBe(
      false,
    );
  });

  it("returns false when balance is within the 0.01 epsilon (floating point noise)", () => {
    expect(canCreateAnotherPV({ balance: 0.005 })).toBe(false);
  });

  it("returns true when balance is just above the 0.01 epsilon", () => {
    expect(canCreateAnotherPV({ balance: 0.02 })).toBe(true);
  });

  it("returns false when balance is negative (overpaid/over-committed edge case)", () => {
    expect(canCreateAnotherPV({ balance: -50 })).toBe(false);
  });

  it("uses the explicit `balance` field over totalAmount/amountCommitted when present", () => {
    expect(
      canCreateAnotherPV({ totalAmount: 1000, amountCommitted: 1000, balance: 300 }),
    ).toBe(true);
  });
});
