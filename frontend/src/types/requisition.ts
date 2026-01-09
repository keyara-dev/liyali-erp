/**
 * Requisition Types
 * Aligned with backend models and database schema
 */

// ============================================================================
// CORE REQUISITION TYPES
// ============================================================================

export interface RequisitionItem {
  id?: string;
  description: string;
  itemDescription?: string; // Alias for description
  quantity: number;
  unitPrice: number;
  amount: number;
  estimatedCost?: number; // Alias for amount
  unit?: string;
  category?: string;
  notes?: string;
  itemNumber?: number; // For UI compatibility
  totalPrice?: number; // Alias for amount
}

export interface Requisition {
  // Core fields
  id: string;
  organizationId: string;
  reqNumber: string;
  requesterId: string;
  requester?: any;
  requesterName: string;
  title: string;
  description: string;
  department: string;
  departmentId: string;
  status: RequisitionStatus; // draft, pending, approved, rejected, completed, cancelled
  priority: RequisitionPriority; // low, medium, high, urgent
  items: RequisitionItem[];
  totalAmount: number;
  currency: string;
  approvalStage: number;
  approvalHistory: any[];
  categoryId?: string;
  category?: any;
  categoryName: string;
  preferredVendorId?: string;
  preferredVendor?: any;
  preferredVendorName: string;

  automationUsed?: boolean;
  autoCreatedPO?: boolean;
  isEstimate: boolean;
  createdAt: Date;
  updatedAt: Date;

  // Business requirement fields
  requisitionNumber: string; // Same as reqNumber
  budgetCode: string;
  requestedByName: string; // Same as requesterName
  requestedByRole: string;
  requestedBy: string; // Same as requesterId
  totalApprovalStages: number;
  requestedDate: Date;
  requiredByDate: Date;
  costCenter: string;
  projectCode: string;
  createdBy: string;
  createdByName: string;
  createdByRole: string;
  requestedFor?: string; // Who the requisition is for
  otherCategoryText?: string; // Custom category name when "OTHER" is selected

  // UI compatibility fields
  documentNumber?: string;
  currentStage?: number;
  currentApprovalStage?: number;
  actionHistory?: any[];
  metadata?: Record<string, any>;
  type?: string;
  createdByUser?: any;
  approvalChain?: any[]; // For PDF generation
  vendorId?: string; // For PO creation
  vendorName?: string; // For PO creation
}

// ============================================================================
// REQUEST TYPES
// ============================================================================

export interface CreateRequisitionRequest {
  title: string;
  description: string;
  department: string;
  departmentId: string;
  priority: string;
  items: RequisitionItem[];
  totalAmount: number;
  currency: string;
  categoryId?: string;
  preferredVendorId?: string;
  isEstimate: boolean;

  // Business requirement fields
  requiredByDate: Date;
  budgetCode: string;
  costCenter: string;
  projectCode: string;
  requestedFor?: string; // Who the requisition is for
  otherCategoryText?: string; // Custom category name when "OTHER" is selected
}

export interface UpdateRequisitionRequest {
  requisitionId: string;
  id?: string; // Alias for requisitionId
  title?: string;
  description?: string;
  department?: string;
  departmentId?: string;
  priority?: string;
  items?: RequisitionItem[];
  totalAmount?: number;
  currency?: string;
  categoryId?: string;
  preferredVendorId?: string;
  isEstimate?: boolean;
  requiredByDate?: Date;
  budgetCode?: string;
  costCenter?: string;
  projectCode?: string;
  requestedFor?: string; // Who the requisition is for
  otherCategoryText?: string; // Custom category name when "OTHER" is selected
}

export interface SubmitRequisitionRequest {
  requisitionId: string;
  submittedBy: string;
  submittedByName: string;
  submittedByRole: string;
  comments?: string;
}

export interface ApproveRequisitionRequest {
  requisitionId: string;
  approvingUserId: string;
  approvingUserName: string;
  approvingUserRole: string;
  signature: string;
  comments?: string;
  stageNumber?: number;
}

export interface RejectRequisitionRequest {
  requisitionId: string;
  rejectingUserId: string;
  rejectingUserName: string;
  rejectingUserRole: string;
  remarks: string;
  signature: string;
  comments?: string;
  returnTo?: "original_submitter" | "previous_stage";
}

// ============================================================================
// STATISTICS TYPES
// ============================================================================

export interface RequisitionStats {
  total: number;
  draft: number;
  pending: number;
  approved: number;
  rejected: number;
  thisMonth: number;
  totalAmount: number;
}

// ============================================================================
// TYPE ALIASES
// ============================================================================

export type RequisitionStatus =
  | "draft"
  | "pending"
  | "submitted"
  | "approved"
  | "rejected"
  | "completed"
  | "cancelled";
export type RequisitionPriority = "low" | "medium" | "high" | "urgent";
