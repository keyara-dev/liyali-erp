/**
 * Storage Initialization
 * Initializes localStorage with seed data on app startup
 *
 * This serves as the single source of truth for all app data
 * until backend APIs are integrated.
 */

import { STORAGE_KEYS, saveDocuments, isStorageInitialized } from './storage';
import {
  createSeedPurchaseOrders,
  createSeedRequisitions,
  createSeedPaymentVouchers,
  createSeedGoodsReceivedNotes,
} from './seed-data';

/**
 * Initialize localStorage with seed data if empty
 * Safe to call multiple times - only initializes once
 */
export function initializeStorage(): void {
  if (typeof window === 'undefined') return;

  // Skip if already initialized
  if (isStorageInitialized()) {
    console.log('✓ Storage already initialized');
    return;
  }

  try {
    // Initialize Purchase Orders
    const purchaseOrders = createSeedPurchaseOrders();
    saveDocuments(STORAGE_KEYS.PURCHASE_ORDERS, purchaseOrders);
    console.log(`✓ Initialized ${purchaseOrders.length} purchase orders`);

    // Initialize Requisitions
    const requisitions = createSeedRequisitions();
    saveDocuments(STORAGE_KEYS.REQUISITIONS, requisitions);
    console.log(`✓ Initialized ${requisitions.length} requisitions`);

    // Initialize Payment Vouchers
    const paymentVouchers = createSeedPaymentVouchers();
    saveDocuments(STORAGE_KEYS.PAYMENT_VOUCHERS, paymentVouchers);
    console.log(`✓ Initialized ${paymentVouchers.length} payment vouchers`);

    // Initialize Goods Received Notes
    const goodsReceivedNotes = createSeedGoodsReceivedNotes();
    saveDocuments(STORAGE_KEYS.GOODS_RECEIVED_NOTES, goodsReceivedNotes);
    console.log(`✓ Initialized ${goodsReceivedNotes.length} goods received notes`);

    console.log('✓ All storage initialized successfully');
  } catch (error) {
    console.error('Failed to initialize storage:', error);
  }
}

/**
 * Reset storage with fresh seed data (for development/testing)
 */
export function resetStorage(): void {
  if (typeof window === 'undefined') return;

  try {
    // Clear all
    Object.values(STORAGE_KEYS).forEach((key) => {
      localStorage.removeItem(key);
    });

    // Reinitialize
    initializeStorage();
    console.log('✓ Storage reset successfully');
  } catch (error) {
    console.error('Failed to reset storage:', error);
  }
}
