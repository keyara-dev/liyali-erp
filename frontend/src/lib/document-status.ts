/**
 * Canonical document status strings written by the backend.
 *
 * These mirror backend/models/status.go exactly. The backend stores them as
 * uppercase strings and normalizes input via strings.ToUpper, so any frontend
 * comparison should either use these constants directly or apply
 * .toUpperCase() first.
 *
 * Do NOT invent new values (e.g. "SUBMITTED", "IN_REVIEW", "CONFIRMED") — those
 * are what caused the approval-UI outage audit. Add to the backend constant
 * file first, migrate, then mirror here.
 */
export const DOCUMENT_STATUS = {
  DRAFT: "DRAFT",
  PENDING: "PENDING",
  APPROVED: "APPROVED",
  REJECTED: "REJECTED",
  REVISION: "REVISION",
  PAID: "PAID", // PV only
  COMPLETED: "COMPLETED", // GRN only (post-confirm)
  FULFILLED: "FULFILLED", // PO only (reserved)
  CANCELLED: "CANCELLED",
} as const;

export type DocumentStatus =
  (typeof DOCUMENT_STATUS)[keyof typeof DOCUMENT_STATUS];
