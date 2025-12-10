/**
 * Storage Hooks
 * High-level hooks for managing documents from storage
 *
 * These hooks provide a clean API for components to read/write data
 * When backend APIs are ready, these hooks can be updated to use them
 * without changing component code
 */

import { PurchaseOrder, PaymentVoucher, RequisitionForm } from '@/types/workflow';
import { STORAGE_KEYS, getDocuments, getDocumentById, saveDocument, deleteDocument } from './storage';

// ============================================================================
// Purchase Order Hooks
// ============================================================================

export function getPurchaseOrders(): PurchaseOrder[] {
  return getDocuments<PurchaseOrder>(STORAGE_KEYS.PURCHASE_ORDERS);
}

export function getPurchaseOrderById(id: string): PurchaseOrder | null {
  return getDocumentById<PurchaseOrder>(STORAGE_KEYS.PURCHASE_ORDERS, id);
}

export function savePurchaseOrder(po: PurchaseOrder): PurchaseOrder {
  return saveDocument<PurchaseOrder>(STORAGE_KEYS.PURCHASE_ORDERS, po);
}

export function deletePurchaseOrder(id: string): void {
  deleteDocument(STORAGE_KEYS.PURCHASE_ORDERS, id);
}

export function filterPurchaseOrders(predicate: (po: PurchaseOrder) => boolean): PurchaseOrder[] {
  return getPurchaseOrders().filter(predicate);
}

export function getPurchaseOrdersByStatus(status: string): PurchaseOrder[] {
  return filterPurchaseOrders((po) => (po as any).status === status);
}

export function getPurchaseOrdersByCreator(creatorId: string): PurchaseOrder[] {
  return filterPurchaseOrders((po) => po.createdBy === creatorId);
}

// ============================================================================
// Requisition Hooks
// ============================================================================

export function getRequisitions(): RequisitionForm[] {
  return getDocuments<RequisitionForm>(STORAGE_KEYS.REQUISITIONS);
}

export function getRequisitionById(id: string): RequisitionForm | null {
  return getDocumentById<RequisitionForm>(STORAGE_KEYS.REQUISITIONS, id);
}

export function saveRequisition(req: RequisitionForm): RequisitionForm {
  return saveDocument<RequisitionForm>(STORAGE_KEYS.REQUISITIONS, req);
}

export function deleteRequisition(id: string): void {
  deleteDocument(STORAGE_KEYS.REQUISITIONS, id);
}

export function filterRequisitions(predicate: (req: RequisitionForm) => boolean): RequisitionForm[] {
  return getRequisitions().filter(predicate);
}

export function getRequisitionsByStatus(status: string): RequisitionForm[] {
  return filterRequisitions((req) => (req as any).status === status);
}

export function getRequisitionsByCreator(creatorId: string): RequisitionForm[] {
  return filterRequisitions((req) => req.createdBy === creatorId);
}

export function getRequisitionsByDepartment(department: string): RequisitionForm[] {
  return filterRequisitions((req) => req.metadata?.department === department);
}

// ============================================================================
// Payment Voucher Hooks
// ============================================================================

export function getPaymentVouchers(): PaymentVoucher[] {
  return getDocuments<PaymentVoucher>(STORAGE_KEYS.PAYMENT_VOUCHERS);
}

export function getPaymentVoucherById(id: string): PaymentVoucher | null {
  return getDocumentById<PaymentVoucher>(STORAGE_KEYS.PAYMENT_VOUCHERS, id);
}

export function savePaymentVoucher(pv: PaymentVoucher): PaymentVoucher {
  return saveDocument<PaymentVoucher>(STORAGE_KEYS.PAYMENT_VOUCHERS, pv);
}

export function deletePaymentVoucher(id: string): void {
  deleteDocument(STORAGE_KEYS.PAYMENT_VOUCHERS, id);
}

export function filterPaymentVouchers(predicate: (pv: PaymentVoucher) => boolean): PaymentVoucher[] {
  return getPaymentVouchers().filter(predicate);
}

export function getPaymentVouchersByStatus(status: string): PaymentVoucher[] {
  return filterPaymentVouchers((pv) => (pv as any).status === status);
}

export function getPaymentVouchersByCreator(creatorId: string): PaymentVoucher[] {
  return filterPaymentVouchers((pv) => pv.createdBy === creatorId);
}

export function getPaymentVouchersByAmount(minAmount: number, maxAmount: number): PaymentVoucher[] {
  return filterPaymentVouchers((pv) => {
    const amount = pv.metadata?.amount || 0;
    return amount >= minAmount && amount <= maxAmount;
  });
}

// ============================================================================
// Bulk Operations
// ============================================================================

export function getAllDocuments() {
  return {
    purchaseOrders: getPurchaseOrders(),
    requisitions: getRequisitions(),
    paymentVouchers: getPaymentVouchers(),
  };
}

export function getDocumentsByStatus(status: string) {
  return {
    purchaseOrders: getPurchaseOrdersByStatus(status),
    requisitions: getRequisitionsByStatus(status),
    paymentVouchers: getPaymentVouchersByStatus(status),
  };
}

export function getDocumentsByCreator(creatorId: string) {
  return {
    purchaseOrders: getPurchaseOrdersByCreator(creatorId),
    requisitions: getRequisitionsByCreator(creatorId),
    paymentVouchers: getPaymentVouchersByCreator(creatorId),
  };
}
