/**
 * Payment Voucher Types
 * Aligned with backend models and database schema
 */

// Import shared types from core
import type { PaymentMethod } from "./core";

// ============================================================================
// CORE PAYMENT VOUCHER TYPES
// ============================================================================

export interface PaymentItem {
  description: string;
  amount: number;
  glCode: string;
  taxAmount?: number;
}

export interface PaymentVoucher {
  // Core fields
  id: string;
  organizationId: string;
  documentNumber: string;
  vendorId: string;
  vendor?: any;
  vendorName: string;
  invoiceNumber: string;
  status: string; // draft, pending, approved, rejected, paid, completed, cancelled
  amount: number;
  currency: string;
  paymentMethod: string; // bank_transfer, check, cash, wire_transfer
  glCode: string;
  description: string;
  approvalStage: number;
  approvalHistory: any[];
  linkedPO: string;
  createdAt: Date;
  updatedAt: Date;

  // Business requirement fields
  bankDetails: any;
  requestedDate: Date;
  totalAmount: number; // Same as amount
  items: PaymentItem[];
  budgetCode: string;
  costCenter: string;
  projectCode: string;
  taxAmount: number;
  withholdingTaxAmount: number;
  paidAmount: number;
  paidDate: Date;
  paymentDueDate: Date;
  requestedByName: string;
  title: string;
  department: string;
  departmentId: string;
  priority: string;
  submittedAt: Date;
  approvedAt: Date;
  createdBy: string;
  ownerId: string; // Same as createdBy

  // UI compatibility fields
  currentStage?: number;
  currentApprovalStage?: number;
  actionHistory?: any[];
  metadata?: Record<string, any>;
  type?: string;
  createdByUser?: any;
  approvalChain?: any[]; // For PDF generation
}

// ============================================================================
// REQUEST TYPES
// ============================================================================

export interface CreatePaymentVoucherRequest {
  vendorId: string;
  vendorName?: string;
  invoiceNumber: string;
  amount: number;
  currency: string;
  paymentMethod: string;
  glCode: string;
  description: string;
  linkedPO: string;

  // Business requirement fields
  title: string;
  department: string;
  departmentId: string;
  priority: string;
  items: PaymentItem[];
  budgetCode: string;
  costCenter: string;
  projectCode: string;
  taxAmount: number;
  withholdingTaxAmount: number;
  paymentDueDate: Date;
  bankDetails: any;
  createdBy: string;
  createdByName?: string;
  createdByRole?: string;
  sourcePurchaseOrderId?: string;
  sourceRequisitionId?: string;
}

export interface UpdatePaymentVoucherRequest {
  paymentVoucherId: string;
  pvId?: string; // Alias for paymentVoucherId
  vendorId?: string;
  vendorName?: string;
  invoiceNumber?: string;
  amount?: number;
  currency?: string;
  paymentMethod?: string;
  glCode?: string;
  description?: string;
  title?: string;
  priority?: string;
  items?: PaymentItem[];
  budgetCode?: string;
  costCenter?: string;
  projectCode?: string;
  taxAmount?: number;
  withholdingTaxAmount?: number;
  paymentDueDate?: Date;
  bankDetails?: any;
  updatedBy?: string;
}

export interface SubmitPaymentVoucherRequest {
  paymentVoucherId: string;
  pvId?: string; // Alias for paymentVoucherId
  workflowId: string; // REQUIRED - Workflow to use for approval
  submittingUserId: string;
  submittedBy?: string; // Alias for submittingUserId
  submittedByName: string;
  submittedByRole: string;
  comments?: string;
}

export interface ApprovePaymentVoucherRequest {
  paymentVoucherId: string;
  pvId?: string; // Alias for paymentVoucherId
  approvingUserId: string;
  approvingUserName: string;
  approvingUserRole: string;
  signature: string;
  comments?: string;
}

export interface RejectPaymentVoucherRequest {
  paymentVoucherId: string;
  pvId?: string; // Alias for paymentVoucherId
  rejectingUserId: string;
  rejectingUserName: string;
  rejectingUserRole: string;
  remarks: string;
  signature: string;
  comments?: string;
}

export interface MarkPaymentVoucherPaidRequest {
  paymentVoucherId: string;
  pvId?: string; // Alias for paymentVoucherId
  paidBy: string;
  markedBy?: string; // Alias for paidBy
  markedByName?: string;
  markedByRole?: string;
  paidAt: Date;
  paidDate?: Date; // Alias for paidAt
  paidAmount: number;
  paymentReference?: string;
  referenceNumber?: string; // Alias for paymentReference
  comments?: string;
}

// ============================================================================
// STATISTICS TYPES
// ============================================================================

export interface PaymentVoucherStats {
  total: number;
  draft: number;
  pending: number;
  approved: number;
  paid: number;
  thisMonth: number;
  totalAmount: number;
}

// ============================================================================
// TYPE ALIASES
// ============================================================================

export type PaymentVoucherStatus =
  | "draft"
  | "pending"
  | "approved"
  | "rejected"
  | "paid"
  | "completed"
  | "cancelled";
// Re-export PaymentMethod from core
export type { PaymentMethod } from "./core";
