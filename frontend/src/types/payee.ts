/**
 * Payee Types
 * Aligned with backend models and database schema
 */

// ============================================================================
// CORE PAYEE TYPES
// ============================================================================

export type PayeeType = "vendor" | "employee" | "other";

export interface Payee {
  id: string;
  organizationId: string;
  payeeType: PayeeType;
  name: string;
  email?: string;
  phone?: string;
  bankName?: string;
  bankAccount?: string;
  taxId?: string;
  sourceVendorId?: string;
  sourceUserId?: string;
  createdAt: string;
  updatedAt: string;
}

/**
 * Snapshot of payee data captured at the time of requisition/PV creation.
 * Stored in JSONB so the payment record reflects who was paid even if
 * the payee record is later modified.
 */
export interface PayeeSnapshot {
  name: string;
  payeeType: PayeeType;
  email?: string;
  phone?: string;
  bankName?: string;
  bankAccount?: string;
  taxId?: string;
}

// ============================================================================
// REQUEST TYPES
// ============================================================================

export interface CreatePayeeInput {
  payeeType: PayeeType;
  name: string;
  email?: string;
  phone?: string;
  bankName?: string;
  bankAccount?: string;
  taxId?: string;
  sourceVendorId?: string;
  sourceUserId?: string;
}

export interface UpdatePayeeInput {
  payeeType?: PayeeType;
  name?: string;
  email?: string;
  phone?: string;
  bankName?: string;
  bankAccount?: string;
  taxId?: string;
}
