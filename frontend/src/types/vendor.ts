/**
 * Vendor Types
 * Note: Core Vendor type moved to core.ts
 */

// Re-export core Vendor type
export type { Vendor } from './core';

// ================== REQUEST TYPES ==================

export interface CreateVendorRequest {
  name: string;
  email: string;
  phone: string;
  country: string;
  city: string;
  bankAccount: string;
  taxId: string;
  bankName?: string;
  accountName?: string;
  accountNumber?: string;
  branchCode?: string;
  swiftCode?: string;
  contactPerson?: string;
  physicalAddress?: string;
}

export interface UpdateVendorRequest {
  name?: string;
  email?: string;
  phone?: string;
  country?: string;
  city?: string;
  bankAccount?: string;
  taxId?: string;
  active?: boolean;
  bankName?: string;
  accountName?: string;
  accountNumber?: string;
  branchCode?: string;
  swiftCode?: string;
  contactPerson?: string;
  physicalAddress?: string;
}

// ================== FILTER TYPES ==================

export interface VendorFilters {
  active?: boolean;
  country?: string;
  search?: string;
}