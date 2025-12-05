/**
 * Payment Voucher Types
 * Handles payment vouchers created from approved purchase orders
 */

export type PaymentVoucherStatus = 'DRAFT' | 'SUBMITTED' | 'IN_REVIEW' | 'APPROVED' | 'REJECTED' | 'PAID';
export type PaymentVoucherPriority = 'URGENT' | 'HIGH' | 'MEDIUM' | 'LOW';
export type PaymentMethod = 'BANK_TRANSFER' | 'CHEQUE' | 'CASH' | 'MOBILE_MONEY';

/**
 * Payment Voucher Item
 * Line item linking to PO items
 */
export interface PaymentVoucherItem {
  id: string;
  pvId: string;
  poItemId: string;
  itemNumber: number;
  description: string;
  category: string;
  quantity: number;
  unitPrice: number;
  unit: string;
  totalPrice: number;
  paidQuantity?: number;
  notes?: string;
  createdAt: Date;
  updatedAt: Date;
}

/**
 * PV Approval Record
 * Track each approval stage in the PV workflow
 */
export interface PVApprovalRecord {
  stageNumber: number;
  stageName: string;
  assignedTo: string; // User ID
  assignedRole: string;
  status: 'PENDING' | 'APPROVED' | 'REJECTED' | 'REVERSED';
  actionTakenAt?: Date;
  actionTakenBy?: string; // User ID
  actionTakenByRole?: string;
  comments?: string;
  remarks?: string; // Required for rejections, optional for approvals
  signature?: string; // Digital signature (base64 encoded PNG)
  reversedAt?: Date;
  reversalReason?: string;
}

/**
 * PV Action History Entry
 * Track every action performed on a PV for audit trail
 */
export interface PVActionHistoryEntry {
  id: string; // UUID
  actionType: 'CREATE' | 'UPDATE' | 'SUBMIT' | 'APPROVE' | 'REJECT' | 'REVERSE' | 'DELETE' | 'MARK_PAID' | 'REVERT_TO_DRAFT';
  performedBy: string; // User ID
  performedByName: string;
  performedByRole: string;
  performedAt: Date;
  stageNumber?: number; // For approval-related actions
  stageName?: string;
  comments?: string;
  remarks?: string; // For rejections
  signature?: string; // Digital signature (base64 encoded PNG)
  previousStatus?: PaymentVoucherStatus;
  newStatus?: PaymentVoucherStatus;
  changedFields?: Record<string, { oldValue: any; newValue: any }>; // Track specific field changes
  metadata?: Record<string, any>; // Any additional metadata
}

/**
 * Payment Voucher
 * Main PV document created from approved purchase orders
 */
export interface PaymentVoucher {
  id: string;
  pvNumber: string; // e.g., "PV-2024-001"
  title: string;
  description?: string;
  vendorId?: string;
  vendorName: string;
  department: string;
  departmentId: string;
  requestedBy: string; // User ID
  requestedByName: string;
  requestedByRole: string;
  requestedDate: Date;
  paymentDueDate: Date;
  priority: PaymentVoucherPriority;
  paymentMethod: PaymentMethod;
  bankDetails?: {
    bankName?: string;
    accountName?: string;
    accountNumber?: string;
    routingNumber?: string;
  };
  status: PaymentVoucherStatus;

  // Line items
  items: PaymentVoucherItem[];
  totalAmount: number; // Sum of all item totals
  currency: string; // "ZMW" or "USD"
  exchangeRate?: number; // If applicable

  // Approval tracking
  approvalChain?: PVApprovalRecord[];
  currentApprovalStage?: number;
  totalApprovalStages?: number;

  // Action history for audit trail
  actionHistory?: PVActionHistoryEntry[];

  // Purchase Order linking
  sourcePurchaseOrderId: string;
  sourcePurchaseOrderNumber: string;
  createdFromPurchaseOrder: boolean;

  // Requisition linking (for traceability)
  sourceRequisitionId?: string;
  sourceRequisitionNumber?: string;

  // Payment tracking
  paidAmount?: number;
  paidDate?: Date;
  referenceNumber?: string; // Bank reference, cheque number, etc.

  // Metadata
  budgetCode?: string;
  costCenter?: string;
  projectCode?: string;
  taxAmount?: number;
  withholdingTaxAmount?: number;

  // Timestamps
  createdAt: Date;
  updatedAt: Date;
  submittedAt?: Date;
  approvedAt?: Date;
  rejectedAt?: Date;
  paidAt?: Date;
}

/**
 * Create Payment Voucher Request DTO
 */
export interface CreatePaymentVoucherRequest {
  title: string;
  description?: string;
  vendorId?: string;
  vendorName: string;
  department: string;
  departmentId: string;
  paymentDueDate: string | Date;
  priority: PaymentVoucherPriority;
  paymentMethod: PaymentMethod;
  bankDetails?: {
    bankName?: string;
    accountName?: string;
    accountNumber?: string;
    routingNumber?: string;
  };
  items: Omit<PaymentVoucherItem, 'id' | 'pvId' | 'createdAt' | 'updatedAt'>[];
  budgetCode?: string;
  costCenter?: string;
  projectCode?: string;
  taxAmount?: number;
  withholdingTaxAmount?: number;
  createdBy: string;
  createdByName: string;
  createdByRole: string;
  // Purchase Order linking
  sourcePurchaseOrderId: string;
  sourcePurchaseOrderNumber: string;
  sourceRequisitionId?: string;
  sourceRequisitionNumber?: string;
}

/**
 * Update Payment Voucher Request DTO
 */
export interface UpdatePaymentVoucherRequest {
  pvId: string;
  title?: string;
  description?: string;
  vendorId?: string;
  vendorName?: string;
  paymentDueDate?: string | Date;
  priority?: PaymentVoucherPriority;
  paymentMethod?: PaymentMethod;
  bankDetails?: {
    bankName?: string;
    accountName?: string;
    accountNumber?: string;
    routingNumber?: string;
  };
  items?: Omit<PaymentVoucherItem, 'pvId' | 'createdAt' | 'updatedAt'>[];
  budgetCode?: string;
  costCenter?: string;
  projectCode?: string;
  taxAmount?: number;
  withholdingTaxAmount?: number;
  updatedBy: string;
}

/**
 * Submit Payment Voucher for Approval Request DTO
 */
export interface SubmitPaymentVoucherRequest {
  pvId: string;
  submittedBy: string; // User ID
  submittedByName: string;
  submittedByRole: string;
  comments?: string;
}

/**
 * Approve Payment Voucher Request DTO
 */
export interface ApprovePaymentVoucherRequest {
  pvId: string;
  approvingUserId: string;
  approvingUserName: string;
  approvingUserRole: string;
  comments?: string;
  signature: string; // Digital signature (required)
  stageNumber?: number;
}

/**
 * Reject Payment Voucher Request DTO
 */
export interface RejectPaymentVoucherRequest {
  pvId: string;
  rejectingUserId: string;
  rejectingUserName: string;
  rejectingUserRole: string;
  remarks: string; // Required rejection reason
  comments?: string;
  signature: string; // Digital signature (required)
}

/**
 * Mark Payment Voucher as Paid Request DTO
 */
export interface MarkPaymentVoucherPaidRequest {
  pvId: string;
  paidAmount: number;
  paidDate: string | Date;
  referenceNumber?: string;
  markedBy: string;
  markedByName: string;
  markedByRole: string;
  comments?: string;
}

/**
 * Payment Voucher Stats
 * Analytics for payment vouchers
 */
export interface PaymentVoucherStats {
  total: number;
  draft: number;
  submitted: number;
  inApproval: number;
  approved: number;
  rejected: number;
  paid: number;
  totalValue: number;
  totalPaid: number;
  pendingPayment: number;
  averageApprovalTime: number; // in days
}
