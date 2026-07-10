"use client";

import { AlertTriangle } from "lucide-react";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { cn } from "@/lib/utils";
import type { Vendor } from "@/types/core";

interface VendorComplianceWarningProps {
  /** Vendor entity — warnings are derived client-side from missing ZRA TPIN / PACRA fields */
  vendor?: Vendor | null;
  /** Precomputed warnings (e.g. PurchaseOrder.complianceWarnings from the backend) */
  warnings?: string[];
  className?: string;
}

/**
 * Amber, warn-only banner surfaced when a vendor is missing its ZRA TPIN or
 * PACRA registration number. Compliance is never blocking — this component is
 * purely informational so procurement can follow up with the vendor.
 *
 * Pass `warnings` when the caller already has a backend-computed list (e.g.
 * `PurchaseOrder.complianceWarnings`); pass `vendor` to derive the warnings
 * client-side from the vendor's current fields. Renders nothing when there is
 * nothing to warn about.
 */
export function VendorComplianceWarning({
  vendor,
  warnings,
  className,
}: VendorComplianceWarningProps) {
  const messages = warnings ?? deriveVendorComplianceWarnings(vendor);

  if (messages.length === 0) return null;

  return (
    <Alert
      className={cn(
        "border-amber-200 bg-amber-50 dark:bg-amber-950/20 dark:border-amber-900 py-3",
        className,
      )}
    >
      <AlertTriangle className="h-4 w-4 text-amber-600 dark:text-amber-500" />
      <AlertDescription className="text-amber-800 dark:text-amber-300">
        <p className="text-sm font-medium text-amber-900 dark:text-amber-200 mb-1">
          Compliance incomplete
        </p>
        <ul className="text-xs space-y-0.5 list-disc list-inside">
          {messages.map((message) => (
            <li key={message}>{message}</li>
          ))}
        </ul>
      </AlertDescription>
    </Alert>
  );
}

/**
 * Derives compliance warnings from a vendor's ZRA TPIN / PACRA fields.
 * Mirrors the backend's vendorComplianceWarnings (legacy taxId counts as TPIN).
 */
export function deriveVendorComplianceWarnings(
  vendor?: Vendor | null,
): string[] {
  if (!vendor) return [];
  const messages: string[] = [];
  if (!vendor.zraTpin && !vendor.taxId) {
    messages.push("Vendor is missing a ZRA TPIN");
  }
  if (!vendor.pacraRegNumber) {
    messages.push("Vendor is missing a PACRA registration number");
  }
  return messages;
}
