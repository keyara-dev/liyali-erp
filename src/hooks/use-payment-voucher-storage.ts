'use client';

import { useEffect, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { PaymentVoucher } from '@/types/workflow';
import { WorkflowDocument } from '@/types/workflow';
import { QUERY_KEYS } from '@/lib/constants';

const PV_STORAGE_KEY = 'liyali_payment_vouchers';

// ============================================================================
// STORAGE UTILITIES
// ============================================================================

/**
 * Load all payment vouchers from localStorage
 */
function loadPaymentVouchersFromStorage(): PaymentVoucher[] {
  try {
    if (typeof window === 'undefined') return [];
    const stored = localStorage.getItem(PV_STORAGE_KEY);
    if (!stored) return [];
    const parsed = JSON.parse(stored);
    return Array.isArray(parsed) ? parsed : [];
  } catch (error) {
    console.error('Failed to load payment vouchers from storage:', error);
    return [];
  }
}

/**
 * Save payment voucher to localStorage
 */
function savePaymentVoucherToStorage(pv: PaymentVoucher): void {
  try {
    if (typeof window === 'undefined') return;
    const vouchers = loadPaymentVouchersFromStorage();
    const index = vouchers.findIndex(r => r.id === pv.id);
    if (index >= 0) {
      vouchers[index] = pv;
    } else {
      vouchers.push(pv);
    }
    localStorage.setItem(PV_STORAGE_KEY, JSON.stringify(vouchers));
  } catch (error) {
    console.error('Failed to save payment voucher to storage:', error);
  }
}

/**
 * Get a specific payment voucher by ID from localStorage
 */
function getPaymentVoucherFromStorage(pvId: string): PaymentVoucher | null {
  try {
    if (typeof window === 'undefined') return null;
    const vouchers = loadPaymentVouchersFromStorage();
    return vouchers.find(r => r.id === pvId) || null;
  } catch (error) {
    console.error('Failed to get payment voucher from storage:', error);
    return null;
  }
}

/**
 * Delete a payment voucher from localStorage
 */
function deletePaymentVoucherFromStorage(pvId: string): void {
  try {
    if (typeof window === 'undefined') return;
    const vouchers = loadPaymentVouchersFromStorage();
    const filtered = vouchers.filter(r => r.id !== pvId);
    localStorage.setItem(PV_STORAGE_KEY, JSON.stringify(filtered));
  } catch (error) {
    console.error('Failed to delete payment voucher from storage:', error);
  }
}

/**
 * Clear all payment vouchers from localStorage
 */
function clearPaymentVouchersStorage(): void {
  try {
    if (typeof window === 'undefined') return;
    localStorage.removeItem(PV_STORAGE_KEY);
  } catch (error) {
    console.error('Failed to clear payment vouchers storage:', error);
  }
}

// ============================================================================
// DATA CONVERSION
// ============================================================================

/**
 * Convert a PaymentVoucher to a WorkflowDocument for display in tables
 */
function paymentVoucherToWorkflowDocument(pv: PaymentVoucher): WorkflowDocument {
  return {
    id: pv.id,
    type: 'PAYMENT_VOUCHER',
    documentNumber: `PV-${pv.id.substring(0, 8).toUpperCase()}`,
    status: pv.status as any,
    currentStage: pv.currentStage || 1,
    createdBy: pv.createdBy,
    createdAt: pv.createdAt instanceof Date ? pv.createdAt : new Date(pv.createdAt),
    updatedAt: pv.updatedAt instanceof Date ? pv.updatedAt : new Date(pv.updatedAt),
    metadata: pv.metadata,
  };
}

/**
 * Public export of conversion function for use in components
 */
export function convertPaymentVoucherToWorkflowDocument(pv: PaymentVoucher): WorkflowDocument {
  return paymentVoucherToWorkflowDocument(pv);
}

// ============================================================================
// REACT HOOKS
// ============================================================================

/**
 * Hook to manage payment voucher data with localStorage persistence
 */
export function usePaymentVoucherStorage() {
  const [isHydrated, setIsHydrated] = useState(false);

  useEffect(() => {
    setIsHydrated(true);
  }, []);

  return {
    isHydrated,
    loadFromStorage: loadPaymentVouchersFromStorage,
    loadOneFromStorage: getPaymentVoucherFromStorage,
    saveToStorage: savePaymentVoucherToStorage,
    deleteFromStorage: deletePaymentVoucherFromStorage,
    clearStorage: clearPaymentVouchersStorage,
  };
}

/**
 * React Query hook for fetching all payment vouchers with localStorage fallback
 */
export const usePaymentVouchersWithStorage = (includeStorageData = true) =>
  useQuery({
    queryKey: [QUERY_KEYS.PAYMENT_VOUCHERS?.ALL || 'PAYMENT_VOUCHERS', 'with-storage'],
    queryFn: async () => {
      let allVouchers: PaymentVoucher[] = [];

      // Load from localStorage only (mock data)
      if (typeof window !== 'undefined') {
        try {
          const storedVouchers = loadPaymentVouchersFromStorage();
          if (storedVouchers && storedVouchers.length > 0) {
            allVouchers = storedVouchers;
          }
        } catch (storageError) {
          console.error('Failed to load payment vouchers from storage:', storageError);
        }
      }

      return allVouchers;
    },
    staleTime: 5 * 60 * 1000,
    gcTime: 10 * 60 * 1000,
  });

/**
 * React Query hook for fetching payment vouchers as workflow documents
 */
export const usePaymentVouchersAsWorkflowDocuments = (includeStorageData = true) =>
  useQuery({
    queryKey: [QUERY_KEYS.PAYMENT_VOUCHERS?.ALL || 'PAYMENT_VOUCHERS', 'as-documents'],
    queryFn: async () => {
      const vouchers = loadPaymentVouchersFromStorage();
      return vouchers.map(paymentVoucherToWorkflowDocument);
    },
    staleTime: 5 * 60 * 1000,
    gcTime: 10 * 60 * 1000,
  });
