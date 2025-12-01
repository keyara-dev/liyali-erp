/**
 * Requisition Types
 * Handles purchase request documents with approval workflows
 */

export type RequisitionStatus = 'DRAFT' | 'SUBMITTED' | 'IN_REVIEW' | 'APPROVED' | 'REJECTED';
export type RequisitionPriority = 'URGENT' | 'HIGH' | 'MEDIUM' | 'LOW';

/**
 * Requisition Item
 * Line item within a requisition for requested goods/services
 */
export interface RequisitionItem {
  id: string;
  requisitionId: string;
  itemNumber: number;
  description: string;
  category: string; // e.g., "Office Supplies", "Equipment", "Services"
  quantity: number;
  unitPrice: number;
  unit: string; // e.g., "pcs", "kg", "hours"
  totalPrice: number; // quantity * unitPrice
  notes?: string;
  createdAt: Date;
  updatedAt: Date;
}

/**
 * Approval Record
 * Track each approval stage in the requisition workflow
 */
export interface ApprovalRecord {
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
 * Requisition
 * Main purchase request document
 */
export interface Requisition {
  id: string;
  requisitionNumber: string; // e.g., "REQ-2024-001"
  title: string;
  description?: string;
  department: string;
  departmentId: string;
  requestedBy: string; // User ID
  requestedByName: string;
  requestedByRole: string;
  requestedDate: Date;
  requiredByDate: Date;
  priority: RequisitionPriority;
  status: RequisitionStatus;

  // Line items
  items: RequisitionItem[];
  totalAmount: number; // Sum of all item totals
  currency: string; // "ZMW" or "USD"

  // Approval tracking
  approvalChain?: ApprovalRecord[];
  currentApprovalStage?: number;
  totalApprovalStages?: number;

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
 * Create Requisition Request DTO
 */
export interface CreateRequisitionRequest {
  title: string;
  description?: string;
  department: string;
  departmentId: string;
  requiredByDate: string | Date;
  priority: RequisitionPriority;
  items: Omit<RequisitionItem, 'id' | 'requisitionId' | 'createdAt' | 'updatedAt'>[];
  budgetCode?: string;
  costCenter?: string;
  projectCode?: string;
  createdBy: string;
  createdByName: string;
  createdByRole: string;
}

/**
 * Update Requisition Request DTO
 */
export interface UpdateRequisitionRequest {
  requisitionId: string;
  title?: string;
  description?: string;
  requiredByDate?: string | Date;
  priority?: RequisitionPriority;
  items?: Omit<RequisitionItem, 'requisitionId' | 'createdAt' | 'updatedAt'>[];
  budgetCode?: string;
  costCenter?: string;
  projectCode?: string;
  updatedBy: string;
}

/**
 * Submit Requisition for Approval Request DTO
 */
export interface SubmitRequisitionRequest {
  requisitionId: string;
  submittedBy: string; // User ID
  submittedByName: string;
  submittedByRole: string;
  comments?: string;
}

/**
 * Approve Requisition Request DTO
 */
export interface ApproveRequisitionRequest {
  requisitionId: string;
  approvingUserId: string;
  approvingUserName: string;
  approvingUserRole: string;
  comments?: string;
  signature: string; // Digital signature (required)
  stageNumber?: number;
}

/**
 * Reject Requisition Request DTO
 */
export interface RejectRequisitionRequest {
  requisitionId: string;
  rejectingUserId: string;
  rejectingUserName: string;
  rejectingUserRole: string;
  remarks: string; // Required rejection reason
  comments?: string;
  signature: string; // Digital signature (required)
}

/**
 * Requisition Stats
 * Analytics for requisitions
 */
export interface RequisitionStats {
  total: number;
  draft: number;
  submitted: number;
  inApproval: number;
  approved: number;
  rejected: number;
  totalValue: number;
  averageApprovalTime: number; // in days
}
