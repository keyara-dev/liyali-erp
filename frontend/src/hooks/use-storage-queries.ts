'use client';

/**
 * React Query Hooks for Storage
 * Provides React Query integration for storage data
 *
 * These hooks wrap the storage layer with React Query for:
 * - Automatic refetching
 * - Caching
 * - Background updates
 *
 * When backend APIs are ready, simply update the queryFn
 * to call API endpoints instead of storage functions
 */

import { useQuery } from '@tanstack/react-query';
import { PurchaseOrder, PaymentVoucher, RequisitionForm, WorkflowDocument } from '@/types/workflow';
import {
  getPurchaseOrders,
  getRequisitions,
  getPaymentVouchers,
  getPurchaseOrdersByCreator,
  getRequisitionsByCreator,
  getPaymentVouchersByCreator,
} from '@/lib/storage';

// ============================================================================
// Purchase Order Queries
// ============================================================================

export const usePurchaseOrdersQuery = () => {
  return useQuery({
    queryKey: ['purchaseOrders'],
    queryFn: () => getPurchaseOrders(),
    staleTime: 5 * 60 * 1000, // 5 minutes
    gcTime: 10 * 60 * 1000, // 10 minutes
  });
};

export const usePurchaseOrdersByCreatorQuery = (userId: string) => {
  return useQuery({
    queryKey: ['purchaseOrders', 'byCreator', userId],
    queryFn: () => getPurchaseOrdersByCreator(userId),
    staleTime: 5 * 60 * 1000,
    gcTime: 10 * 60 * 1000,
    enabled: !!userId,
  });
};

export const usePurchaseOrdersAsWorkflowDocumentsQuery = (userId?: string) => {
  return useQuery({
    queryKey: ['purchaseOrders', 'asDocuments', userId],
    queryFn: () => {
      let orders = getPurchaseOrders();
      if (userId) {
        orders = orders.filter((po) => po.createdBy === userId);
      }
      return convertToWorkflowDocuments(orders);
    },
    staleTime: 5 * 60 * 1000,
    gcTime: 10 * 60 * 1000,
  });
};

// ============================================================================
// Requisition Queries
// ============================================================================

export const useRequisitionsQuery = () => {
  return useQuery({
    queryKey: ['requisitions'],
    queryFn: () => getRequisitions(),
    staleTime: 5 * 60 * 1000,
    gcTime: 10 * 60 * 1000,
  });
};

export const useRequisitionsByCreatorQuery = (userId: string) => {
  return useQuery({
    queryKey: ['requisitions', 'byCreator', userId],
    queryFn: () => getRequisitionsByCreator(userId),
    staleTime: 5 * 60 * 1000,
    gcTime: 10 * 60 * 1000,
    enabled: !!userId,
  });
};

export const useRequisitionsAsWorkflowDocumentsQuery = (userId?: string) => {
  return useQuery({
    queryKey: ['requisitions', 'asDocuments', userId],
    queryFn: () => {
      let reqs = getRequisitions();
      if (userId) {
        reqs = reqs.filter((req) => req.createdBy === userId);
      }
      return convertToWorkflowDocuments(reqs);
    },
    staleTime: 5 * 60 * 1000,
    gcTime: 10 * 60 * 1000,
  });
};

// ============================================================================
// Payment Voucher Queries
// ============================================================================

export const usePaymentVouchersQuery = () => {
  return useQuery({
    queryKey: ['paymentVouchers'],
    queryFn: () => getPaymentVouchers(),
    staleTime: 5 * 60 * 1000,
    gcTime: 10 * 60 * 1000,
  });
};

export const usePaymentVouchersByCreatorQuery = (userId: string) => {
  return useQuery({
    queryKey: ['paymentVouchers', 'byCreator', userId],
    queryFn: () => getPaymentVouchersByCreator(userId),
    staleTime: 5 * 60 * 1000,
    gcTime: 10 * 60 * 1000,
    enabled: !!userId,
  });
};

export const usePaymentVouchersAsWorkflowDocumentsQuery = (userId?: string) => {
  return useQuery({
    queryKey: ['paymentVouchers', 'asDocuments', userId],
    queryFn: () => {
      let vouchers = getPaymentVouchers();
      if (userId) {
        vouchers = vouchers.filter((pv) => pv.createdBy === userId);
      }
      return convertToWorkflowDocuments(vouchers);
    },
    staleTime: 5 * 60 * 1000,
    gcTime: 10 * 60 * 1000,
  });
};

// ============================================================================
// Helper Functions
// ============================================================================

function convertToWorkflowDocuments(
  documents: (PurchaseOrder | RequisitionForm | PaymentVoucher)[]
): WorkflowDocument[] {
  return documents.map((doc) => ({
    id: doc.id,
    type: doc.type,
    documentNumber: (doc as any).documentNumber || doc.id,
    status: (doc as any).status || 'DRAFT',
    currentStage: doc.currentStage || 0,
    createdBy: doc.createdBy,
    createdByUser: doc.createdByUser,
    createdAt: doc.createdAt instanceof Date ? doc.createdAt : new Date(doc.createdAt),
    updatedAt: doc.updatedAt instanceof Date ? doc.updatedAt : new Date(doc.updatedAt),
    metadata: doc.metadata,
  }));
}
