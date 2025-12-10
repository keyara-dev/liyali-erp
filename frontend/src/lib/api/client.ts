/**
 * API Client
 * Central point for all API communication
 *
 * Currently uses mock/server actions
 * Will be replaced with actual HTTP calls when backend API is ready
 *
 * Migration Path:
 * 1. Replace server action imports with HTTP client calls
 * 2. Keep the same method signatures for compatibility
 * 3. Update error handling for network errors
 * 4. Update cache invalidation logic if needed
 */

import { APIResponse } from '@/types';
import { Requisition, RequisitionStats, CreateRequisitionRequest } from '@/types/requisition';
import { PurchaseOrder, PurchaseOrderStats } from '@/types/purchase-order';
import { PaymentVoucher, PaymentVoucherStats } from '@/types/payment-voucher';

// For now, import server actions
import {
  getRequisitions as getRequisitionsAction,
  getRequisitionById as getRequisitionByIdAction,
  createRequisition as createRequisitionAction,
  getRequisitionStats as getRequisitionStatsAction,
} from '@/app/_actions/requisitions';

import {
  getPurchaseOrders as getPurchaseOrdersAction,
  getPurchaseOrderById as getPurchaseOrderByIdAction,
  getPurchaseOrderStats as getPurchaseOrderStatsAction,
} from '@/app/_actions/purchase-orders';

import {
  getPaymentVouchers as getPaymentVouchersAction,
  getPaymentVoucherById as getPaymentVoucherByIdAction,
  getPaymentVoucherStats as getPaymentVoucherStatsAction,
} from '@/app/_actions/payment-vouchers';

/**
 * API Client
 * Single source for all API calls
 *
 * When backend is ready, replace imports above with:
 * const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:3001/api';
 *
 * Then implement like:
 * const requisitions = {
 *   getAll: () => fetch(`${API_BASE_URL}/requisitions`).then(r => r.json()),
 *   getById: (id) => fetch(`${API_BASE_URL}/requisitions/${id}`).then(r => r.json()),
 *   ...
 * }
 */

export const apiClient = {
  /**
   * Requisitions API
   */
  requisitions: {
    getAll: async (): Promise<APIResponse<Requisition[]>> => {
      return getRequisitionsAction();
    },

    getById: async (id: string): Promise<APIResponse<Requisition>> => {
      return getRequisitionByIdAction(id);
    },

    create: async (data: CreateRequisitionRequest): Promise<APIResponse<Requisition>> => {
      return createRequisitionAction(data);
    },

    getStats: async (): Promise<APIResponse<RequisitionStats>> => {
      return getRequisitionStatsAction();
    },
  },

  /**
   * Purchase Orders API
   */
  purchaseOrders: {
    getAll: async (): Promise<APIResponse<PurchaseOrder[]>> => {
      return getPurchaseOrdersAction();
    },

    getById: async (id: string): Promise<APIResponse<PurchaseOrder>> => {
      return getPurchaseOrderByIdAction(id);
    },

    getStats: async (): Promise<APIResponse<PurchaseOrderStats>> => {
      return getPurchaseOrderStatsAction();
    },
  },

  /**
   * Payment Vouchers API
   */
  paymentVouchers: {
    getAll: async (): Promise<APIResponse<PaymentVoucher[]>> => {
      return getPaymentVouchersAction();
    },

    getById: async (id: string): Promise<APIResponse<PaymentVoucher>> => {
      return getPaymentVoucherByIdAction(id);
    },

    getStats: async (): Promise<APIResponse<PaymentVoucherStats>> => {
      return getPaymentVoucherStatsAction();
    },
  },
};

/**
 * Migration Checklist
 *
 * When real backend API is ready:
 * ✅ Create new HTTP client (axios/fetch wrapper)
 * ✅ Update apiClient methods to use HTTP
 * ✅ Add proper error handling for network failures
 * ✅ Add request/response interceptors if needed
 * ✅ Add timeout handling
 * ✅ Update offline queue processor to execute API calls
 * ✅ Test with real API
 * ✅ Remove server action imports
 * ✅ Remove documentStore usage from server actions
 * ✅ Remove mock data from requisitions.ts, purchase-orders.ts, payment-vouchers.ts
 */
