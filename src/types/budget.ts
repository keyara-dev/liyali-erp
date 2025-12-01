/**
 * Budget Types
 * Budget management, budget items, and budget workflow
 */

export type BudgetStatus =
  | "DRAFT"
  | "SUBMITTED"
  | "IN_APPROVAL"
  | "APPROVED"
  | "REJECTED"
  | "ACTIVE"
  | "CLOSED";

export interface BudgetItem {
  id: string;
  category: string;
  description: string;
  allocatedAmount: number;
  spentAmount: number;
  remainingAmount: number;
  notes?: string;
  createdAt: Date;
  updatedAt: Date;
}

export interface Budget {
  id: string;
  budgetNumber: string; // e.g., "BDG-2024-001"
  name: string;
  description?: string;
  department: string;
  departmentId: string;
  fiscalYear: string; // e.g., "2024"
  totalAmount: number;
  currency: string; // e.g., "USD", "ZMW"
  items: BudgetItem[];
  status: BudgetStatus;
  createdBy: string; // User ID
  createdAt: Date;
  updatedAt: Date;
  submittedAt?: Date;
  approvedAt?: Date;
  rejectedAt?: Date;
  approvalChain?: ApprovalRecord[]; // Approval audit trail
  currentApprovalStage?: number;
  totalApprovalStages?: number;
  comments?: string;
  rejectionReason?: string;
}

export interface ApprovalRecord {
  stageNumber: number;
  stageName: string;
  assignedTo: string; // User ID
  assignedRole: string;
  status: "PENDING" | "APPROVED" | "REJECTED" | "REVERSED";
  actionTakenAt?: Date;
  actionTakenBy?: string;
  comments?: string;
  remarks?: string; // Required for rejections, optional for approvals
  signature?: string; // Digital signature (base64 encoded)
  reversedAt?: Date;
  reversalReason?: string;
}

export interface BudgetFilters {
  department?: string;
  status?: BudgetStatus;
  fiscalYear?: string;
  createdBy?: string;
  startDate?: string;
  endDate?: string;
}

export interface BudgetFormData {
  name: string;
  description?: string;
  department: string;
  fiscalYear: string;
  items: BudgetItem[];
  comments?: string;
}

export interface CreateBudgetRequest {
  name: string;
  description?: string;
  department: string;
  departmentId: string;
  fiscalYear: string;
  totalAmount: number;
  currency: string;
  items: Omit<BudgetItem, "id" | "createdAt" | "updatedAt">[];
  createdBy: string;
}

export interface ApproveBudgetRequest {
  budgetId: string;
  approvingUserId: string;
  approvingUserRole: string;
  comments?: string;
  stageNumber?: number;
}

export interface RejectBudgetRequest {
  budgetId: string;
  rejectingUserId: string;
  rejectionReason: string;
  comments?: string;
}

export interface SubmitBudgetRequest {
  budgetId: string;
  submittingUserId: string;
}
