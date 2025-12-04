/**
 * Purchase Order Types
 * Handles purchase orders created from approved requisitions
 */

export type PurchaseOrderStatus = 'DRAFT' | 'SUBMITTED' | 'IN_REVIEW' | 'APPROVED' | 'REJECTED';
export type PurchaseOrderPriority = 'URGENT' | 'HIGH' | 'MEDIUM' | 'LOW';

/**
 * Purchase Order Item
 * Line item within a purchase order (sourced from requisition items)
 */
export interface PurchaseOrderItem {
  id: string;
  poId: string;
  itemNumber: number;
  description: string;
  category: string;
  quantity: number;
  unitPrice: number;
  unit: string;
  totalPrice: number;
  notes?: string;
  createdAt: Date;
  updatedAt: Date;
}

/**
 * PO Approval Record
 * Track each approval stage in the PO workflow
 */
export interface POApprovalRecord {
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
 * PO Action History Entry
 * Track every action performed on a PO for audit trail
 */
export interface POActionHistoryEntry {
  id: string; // UUID
  actionType: 'CREATE' | 'UPDATE' | 'SUBMIT' | 'APPROVE' | 'REJECT' | 'REVERSE' | 'DELETE' | 'REVERT_TO_DRAFT';
  performedBy: string; // User ID
  performedByName: string;
  performedByRole: string;
  performedAt: Date;
  stageNumber?: number; // For approval-related actions
  stageName?: string;
  comments?: string;
  remarks?: string; // For rejections
  signature?: string; // Digital signature (base64 encoded PNG)
  previousStatus?: PurchaseOrderStatus;
  newStatus?: PurchaseOrderStatus;
  changedFields?: Record<string, { oldValue: any; newValue: any }>; // Track specific field changes
  metadata?: Record<string, any>; // Any additional metadata
}

/**
 * Purchase Order
 * Main PO document created from approved requisitions
 */
export interface PurchaseOrder {
  id: string;
  poNumber: string; // e.g., "PO-2024-001"
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
  requiredByDate: Date;
  priority: PurchaseOrderPriority;
  status: PurchaseOrderStatus;

  // Line items
  items: PurchaseOrderItem[];
  totalAmount: number; // Sum of all item totals
  currency: string; // "ZMW" or "USD"

  // Approval tracking
  approvalChain?: POApprovalRecord[];
  currentApprovalStage?: number;
  totalApprovalStages?: number;

  // Action history for audit trail
  actionHistory?: POActionHistoryEntry[];

  // Requisition linking
  sourceRequisitionId: string;
  sourceRequisitionNumber: string;
  createdFromRequisition: boolean;

  // Metadata
  budgetCode?: string;
  costCenter?: string;
  projectCode?: string;

  // Timestamps
  createdAt: Date;
  updatedAt: Date;
  submittedAt?: Date;
  approvedAt?: Date;
  rejectedAt?: Date;
}

/**
 * Create Purchase Order Request DTO
 */
export interface CreatePurchaseOrderRequest {
  title: string;
  description?: string;
  vendorId?: string;
  vendorName: string;
  department: string;
  departmentId: string;
  requiredByDate: string | Date;
  priority: PurchaseOrderPriority;
  items: Omit<PurchaseOrderItem, 'id' | 'poId' | 'createdAt' | 'updatedAt'>[];
  budgetCode?: string;
  costCenter?: string;
  projectCode?: string;
  createdBy: string;
  createdByName: string;
  createdByRole: string;
  // Requisition linking
  sourceRequisitionId: string;
  sourceRequisitionNumber: string;
}

/**
 * Update Purchase Order Request DTO
 */
export interface UpdatePurchaseOrderRequest {
  poId: string;
  title?: string;
  description?: string;
  vendorId?: string;
  vendorName?: string;
  requiredByDate?: string | Date;
  priority?: PurchaseOrderPriority;
  items?: Omit<PurchaseOrderItem, 'poId' | 'createdAt' | 'updatedAt'>[];
  budgetCode?: string;
  costCenter?: string;
  projectCode?: string;
  updatedBy: string;
}

/**
 * Submit Purchase Order for Approval Request DTO
 */
export interface SubmitPurchaseOrderRequest {
  poId: string;
  submittedBy: string; // User ID
  submittedByName: string;
  submittedByRole: string;
  comments?: string;
}

/**
 * Approve Purchase Order Request DTO
 */
export interface ApprovePurchaseOrderRequest {
  poId: string;
  approvingUserId: string;
  approvingUserName: string;
  approvingUserRole: string;
  comments?: string;
  signature: string; // Digital signature (required)
  stageNumber?: number;
}

/**
 * Reject Purchase Order Request DTO
 */
export interface RejectPurchaseOrderRequest {
  poId: string;
  rejectingUserId: string;
  rejectingUserName: string;
  rejectingUserRole: string;
  remarks: string; // Required rejection reason
  comments?: string;
  signature: string; // Digital signature (required)
}

/**
 * Purchase Order Stats
 * Analytics for purchase orders
 */
export interface PurchaseOrderStats {
  total: number;
  draft: number;
  submitted: number;
  inApproval: number;
  approved: number;
  rejected: number;
  totalValue: number;
  averageApprovalTime: number; // in days
}
