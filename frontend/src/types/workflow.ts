/**
 * Workflow Types
 * Document types, statuses, and approval workflows
 */

// Workflow Document Types
export type WorkflowDocumentType =
  | "PURCHASE_ORDER"
  | "PAYMENT_VOUCHER"
  | "REQUISITION"
  | "GOODS_RECEIVED_NOTE";

export type DocumentStatus =
  | "DRAFT"
  | "SUBMITTED"
  | "IN_REVIEW"
  | "APPROVED"
  | "REJECTED"
  | "REVERSED";

export type ApprovalAction =
  | "APPROVED"
  | "REJECTED"
  | "COMMENTED"
  | "REASSIGNED"
  | "REVERSED";

// User & Role Types (from auth.ts)
export type UserRole =
  | "requester"
  | "department_manager"
  | "finance_officer"
  | "director"
  | "cfo"
  | "compliance_officer"
  | "admin";

export type WorkflowPermission =
  | "view_draft"
  | "edit_draft"
  | "submit_document"
  | "approve_document"
  | "reject_document"
  | "reassign_approver"
  | "view_attachments"
  | "add_attachments"
  | "view_comments"
  | "add_comments"
  | "view_audit_log"
  | "manage_approvers"
  | "manage_workflows";

// Alias for backward compatibility
export type Permission = WorkflowPermission;

// Base Workflow Document
export type WorkflowDocument = {
  id: string;
  type: WorkflowDocumentType;
  documentNumber: string;
  status: DocumentStatus;
  currentStage: number;
  createdBy: string;
  createdByUser?: User;
  createdAt: Date;
  updatedAt: Date;
  metadata: Record<string, any>;
};

// Workflow Step Definition
export type WorkflowStep = {
  id?: string;
  workflowType: WorkflowDocumentType;
  stepOrder: number;
  roleName: UserRole;
  description: string;
  isRequired: boolean;
  permissions?: WorkflowPermission[];
};

// Approver Assignment
export type Approver = {
  id: string;
  documentId: string;
  stepOrder: number;
  userId: string;
  user?: User;
  role: UserRole;
  assignedAt: Date;
  canReassign: boolean;
  status: "PENDING" | "APPROVED" | "REJECTED" | "SKIPPED";
};

// Approval Log Entry
export type ApprovalLogEntry = {
  id: string;
  documentId: string;
  approver: User;
  approverId: string;
  action: ApprovalAction;
  timestamp: Date;
  comments?: string;
  remarks?: string; // Required for rejections, optional for approvals
  signature?: string; // Digital signature (base64 encoded)
  ipAddress?: string;
};

// Attachment
export type Attachment = {
  id: string;
  documentId: string;
  fileName: string;
  fileSize: number;
  fileType: string;
  uploadedBy: User;
  uploadedById: string;
  uploadedAt: Date;
  storagePath: string;
  visibleToRoles: UserRole[];
};

// Document-Specific Types
export type PurchaseOrderItem = {
  id: string;
  description: string;
  quantity: number;
  unitCost: number;
  totalCost: number;
};

export type PurchaseOrder = WorkflowDocument & {
  metadata: {
    vendorName: string;
    vendorId: string;
    items: PurchaseOrderItem[];
    totalAmount: number;
    currency: string;
    deliveryDate: Date;
    specialInstructions?: string;
  };
};

export type PaymentVoucher = WorkflowDocument & {
  metadata: {
    payeeName: string;
    payeeId: string;
    amount: number;
    currency: string;
    reason: string;
    accountCode: string;
    department: string;
  };
};

export type RequisitionItem = {
  id: string;
  itemDescription: string;
  quantity: number;
  estimatedCost: number;
};

export type RequisitionForm = WorkflowDocument & {
  metadata: {
    department: string;
    requestedFor: string;
    items: RequisitionItem[];
    justification: string;
    budgetCode: string;
  };
};

// User Type (basic)
export type User = {
  id: string;
  name: string;
  email: string;
  role: UserRole;
  department?: string;
  avatar?: string;
};

// Response wrapper
export type PaginatedResponse<T> = {
  data: T[];
  pagination: {
    page: number;
    limit: number;
    total: number;
    totalPages: number;
  };
};

// Dynamic Approval Configuration Types
export type ReversalBehavior =
  | "BACK_TO_CREATOR"
  | "BACK_TO_HANDLER"
  | "PREVIOUS_STAGE"
  | "TO_SPECIFIC_USER";

export type ApprovalStageConfig = {
  stageNumber: number;
  stageName: string;
  description?: string;
  requiredRole: string;
  alternativeRoles?: string[];
  canReverse: boolean;
  reversalBehavior: ReversalBehavior;
  reversalTargetRole?: string;
  reversalResetsPreviousStages?: boolean;
  requiresComments?: boolean;
  requiredValidations?: string[];
  onApprovalActions?: {
    generateQRCode?: boolean;
    generatePaymentReference?: boolean;
    createAuditLog?: boolean;
    notifyVendor?: boolean;
    createPaymentVoucher?: boolean;
  };
  slaHours?: number;
  escalationRoleAfterSLA?: string;
};

export type ApprovalRecord = {
  stageNumber: number;
  stageName: string;
  assignedTo: string;
  assignedRole: string;
  status: "PENDING" | "APPROVED" | "REVERSED" | "REJECTED";
  actionTakenAt?: Date;
  actionTakenBy?: string;
  comments?: string;
  reversedAt?: Date;
  reversalReason?: string;
  validationsPassed?: string[];
  validationsFailed?: string[];
};

export type ApprovalState = {
  documentId: string;
  documentType: string;
  configVersion: string;
  currentStageNumber: number;
  totalStages: number;
  stageHistory: ApprovalRecord[];
  status: DocumentStatus;
  submittedAt?: Date;
  approvedAt?: Date;
  rejectedAt?: Date;
  lastModifiedAt: Date;
  lastModifiedBy: string;
};

export type DocumentApprovalConfig = {
  documentType: WorkflowDocumentType;
  configVersion: string;
  effectiveDate: Date;
  description: string;
  totalStages: number;
  approvalStages: ApprovalStageConfig[];
  fallbackStages?: ApprovalStageConfig[];
  allowConcurrentApprovals?: boolean;
  allowMultipleReversals?: boolean;
  requireFinalSignoff?: boolean;
  createdAt: Date;
  createdBy: string;
};

export type ApproveDocumentRequest = {
  documentId: string;
  documentType: string;
  approvingUserId: string;
  comments?: string;
  validations?: Record<string, boolean>;
};

export type ApproveDocumentResponse = {
  success: boolean;
  message: string;
  newStageNumber?: number;
  isFinalApproval?: boolean;
  generatedQRCode?: string;
  generatedPaymentReference?: string;
  error?: string;
};

export type ReverseDocumentRequest = {
  documentId: string;
  documentType: string;
  reversingUserId: string;
  reversalReason: string;
};

export type ReverseDocumentResponse = {
  success: boolean;
  message: string;
  reversedToStage?: number;
  reversedToRole?: string;
  error?: string;
};

export type SearchFilters = {
  documentNumber: string;
  documentType: "ALL" | WorkflowDocumentType;
  status: "ALL" | DocumentStatus;
  startDate: string;
  endDate: string;
};

// ==========================================
// NEW: Backend-Powered Approval Workflow Types
// ==========================================

/**
 * Approval Task - represents a pending approval action
 */
export interface ApprovalTask {
  id: string;
  organizationId: string;
  documentId: string;
  documentType: string;
  documentNumber: string;
  approverId: string;
  approverName?: string;
  approverRole?: string;
  status: 'PENDING' | 'APPROVED' | 'REJECTED' | 'CANCELLED';
  stage: number;
  totalStages?: number;
  priority?: string;
  comments?: string;
  createdAt: Date;
  updatedAt: Date;
  dueAt?: Date;
  overdue?: boolean;
  document?: {
    id: string;
    title: string;
    amount?: number;
    currency?: string;
    requester?: string;
    status?: string;
  };
}

/**
 * Approval history record - single approval entry in document history
 */
export interface ApprovalHistory {
  id?: string;
  stage: number;
  stageName?: string;
  approverId: string;
  approverName: string;
  approverRole?: string;
  status: 'APPROVED' | 'REJECTED';
  action?: string;
  comments?: string;
  remarks?: string;
  signature?: string; // Base64 encoded signature image
  approvedAt: Date;
  duration?: number; // Seconds to approve
}

/**
 * Request to approve a task
 */
export interface ApproveTaskRequest {
  taskId: string;
  comments: string;
  signature: string; // Base64 encoded signature image
  stageNumber: number;
}

/**
 * Request to reject a task
 */
export interface RejectTaskRequest {
  taskId: string;
  remarks: string; // Required reason for rejection
  comments: string;
  signature: string; // Base64 encoded signature image
  returnTo?: string;
}

/**
 * Request to reassign a task
 */
export interface ReassignTaskRequest {
  taskId: string;
  newApproverId: string;
  reason: string;
}
