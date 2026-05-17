/**
 * Property-Based Tests for PO Creation Wizard — Line Item Variance
 *
 * **Property 12: Variance colour coding is correct for all values**
 * For any `variance` value and `reqEstPrice`:
 * - If `variance <= 0`: green colour class (PO price at or under REQ estimate)
 * - If `reqEstPrice === 0`: neutral colour class (no reference price)
 * - If `variance > 0` and `variance / reqEstPrice <= 0.10`: amber colour class
 * - If `variance > 0` and `variance / reqEstPrice > 0.10`: red colour class
 *
 * **Validates: Requirements 6.4, 6.5, 6.6, 6.7**
 */

import { describe, it, expect } from "vitest";
import * as fc from "fast-check";
import {
  computeLineItemVariance,
  lineItemVarianceColorClass,
} from "@/app/(private)/(main)/purchase-orders/_components/po-creation-wizard/types";

// ============================================================================
// computeLineItemVariance
// ============================================================================

describe("computeLineItemVariance", () => {
  /**
   * For any poUnitPrice and reqEstPrice, the variance SHALL equal
   * poUnitPrice - reqEstPrice exactly.
   *
   * **Validates: Requirements 6.4**
   */
  it("should return poUnitPrice - reqEstPrice for all inputs", () => {
    fc.assert(
      fc.property(
        fc.float({ noNaN: true, noDefaultInfinity: true }),
        fc.float({ noNaN: true, noDefaultInfinity: true }),
        (poUnitPrice, reqEstPrice) => {
          const variance = computeLineItemVariance(poUnitPrice, reqEstPrice);
          expect(variance).toBe(poUnitPrice - reqEstPrice);
        },
      ),
      { numRuns: 100 },
    );
  });

  it("should return a negative variance when PO price is below REQ price", () => {
    fc.assert(
      fc.property(
        fc.float({ min: Math.fround(0.01), max: 1_000_000, noNaN: true }),
        fc.float({ min: Math.fround(0.01), max: 1_000_000, noNaN: true }),
        (delta, reqEstPrice) => {
          const poUnitPrice = reqEstPrice - delta;
          const variance = computeLineItemVariance(poUnitPrice, reqEstPrice);
          expect(variance).toBeLessThan(0);
        },
      ),
      { numRuns: 100 },
    );
  });

  it("should return zero variance when PO price equals REQ price", () => {
    fc.assert(
      fc.property(
        fc.float({ min: 0, max: 1_000_000, noNaN: true }),
        (price) => {
          const variance = computeLineItemVariance(price, price);
          expect(variance).toBe(0);
        },
      ),
      { numRuns: 100 },
    );
  });

  it("should return a positive variance when PO price is above REQ price", () => {
    fc.assert(
      fc.property(
        fc.float({ min: Math.fround(0.01), max: 1_000_000, noNaN: true }),
        fc.float({ min: Math.fround(0.01), max: 1_000_000, noNaN: true }),
        (delta, reqEstPrice) => {
          const poUnitPrice = reqEstPrice + delta;
          const variance = computeLineItemVariance(poUnitPrice, reqEstPrice);
          expect(variance).toBeGreaterThan(0);
        },
      ),
      { numRuns: 100 },
    );
  });
});

// ============================================================================
// lineItemVarianceColorClass — Property 12
// ============================================================================

describe("Property 12: Variance colour coding is correct for all values", () => {
  /**
   * When variance <= 0, the colour class SHALL be green regardless of reqEstPrice.
   *
   * **Validates: Requirements 6.5**
   */
  it("should return green class when variance is negative (PO price below REQ estimate)", () => {
    fc.assert(
      fc.property(
        fc.float({
          max: Math.fround(-0.001),
          noNaN: true,
          noDefaultInfinity: true,
        }),
        fc.float({ min: 0, noNaN: true, noDefaultInfinity: true }),
        (variance, reqEstPrice) => {
          const colorClass = lineItemVarianceColorClass(variance, reqEstPrice);
          expect(colorClass).toContain("green");
        },
      ),
      { numRuns: 100 },
    );
  });

  /**
   * When variance === 0, the colour class SHALL be green (at budget).
   *
   * **Validates: Requirements 6.5**
   */
  it("should return green class when variance is exactly zero", () => {
    fc.assert(
      fc.property(
        fc.float({ min: 0, noNaN: true, noDefaultInfinity: true }),
        (reqEstPrice) => {
          const colorClass = lineItemVarianceColorClass(0, reqEstPrice);
          expect(colorClass).toContain("green");
        },
      ),
      { numRuns: 100 },
    );
  });

  /**
   * When variance > 0 and reqEstPrice === 0, the colour class SHALL be neutral/muted.
   *
   * **Validates: Requirements 6.4**
   */
  it("should return neutral class when reqEstPrice is 0 (no reference price)", () => {
    fc.assert(
      fc.property(
        fc.float({
          min: Math.fround(0.001),
          noNaN: true,
          noDefaultInfinity: true,
        }),
        (variance) => {
          const colorClass = lineItemVarianceColorClass(variance, 0);
          expect(colorClass).toContain("muted");
        },
      ),
      { numRuns: 100 },
    );
  });

  /**
   * When variance > 0 and variance / reqEstPrice <= 0.10, the colour class SHALL be amber.
   *
   * **Validates: Requirements 6.7**
   */
  it("should return amber class when variance is positive and within 10% of reqEstPrice", () => {
    fc.assert(
      fc.property(
        // reqEstPrice > 0
        fc.float({ min: 1, max: 1_000_000, noNaN: true }),
        // ratio safely in (0, 0.09] to avoid floating-point boundary issues at exactly 0.10
        fc.float({
          min: Math.fround(0.001),
          max: Math.fround(0.09),
          noNaN: true,
        }),
        (reqEstPrice, ratio) => {
          const variance = ratio * reqEstPrice;
          const colorClass = lineItemVarianceColorClass(variance, reqEstPrice);
          expect(colorClass).toContain("amber");
        },
      ),
      { numRuns: 100 },
    );
  });

  /**
   * When variance > 0 and variance / reqEstPrice > 0.10, the colour class SHALL be red.
   *
   * **Validates: Requirements 6.7**
   */
  it("should return red class when variance is positive and exceeds 10% of reqEstPrice", () => {
    fc.assert(
      fc.property(
        // reqEstPrice > 0
        fc.float({ min: 1, max: 1_000_000, noNaN: true }),
        // ratio > 0.10 — use values clearly above 10%
        fc.float({ min: Math.fround(0.11), max: 10, noNaN: true }),
        (reqEstPrice, ratio) => {
          const variance = ratio * reqEstPrice;
          const colorClass = lineItemVarianceColorClass(variance, reqEstPrice);
          expect(colorClass).toContain("red");
        },
      ),
      { numRuns: 100 },
    );
  });

  /**
   * The colour class is never empty — always one of the four tiers.
   *
   * **Validates: Requirements 6.4, 6.5, 6.6, 6.7**
   */
  it("should always return a non-empty colour class", () => {
    fc.assert(
      fc.property(
        fc.float({ noNaN: true, noDefaultInfinity: true }),
        fc.float({ min: 0, noNaN: true, noDefaultInfinity: true }),
        (variance, reqEstPrice) => {
          const colorClass = lineItemVarianceColorClass(variance, reqEstPrice);
          expect(colorClass.length).toBeGreaterThan(0);
        },
      ),
      { numRuns: 100 },
    );
  });
});
