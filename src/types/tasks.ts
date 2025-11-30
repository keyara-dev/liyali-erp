/**
 * Tasks Types
 * Pending workflow actions and tasks assigned to users
 */

export type TaskType =
  | "REQUISITION_APPROVAL"
  | "PURCHASE_ORDER_APPROVAL"
  | "PAYMENT_VOUCHER_APPROVAL"
  | "GOODS_RECEIVED_NOTE_CONFIRMATION"
  | "BUDGET_APPROVAL";

export type TaskPriority = "LOW" | "MEDIUM" | "HIGH" | "URGENT";

export type TaskStatus = "PENDING" | "IN_PROGRESS" | "COMPLETED" | "OVERDUE";

export interface Task {
  id: string;
  taskType: TaskType;
  documentId: string;
  documentNumber: string;
  documentType: string;
  title: string;
  description: string;
  assignedTo: string; // User ID
  assignedRole: string;
  priority: TaskPriority;
  status: TaskStatus;
  createdAt: Date;
  dueDate: Date;
  completedAt?: Date;
  completedBy?: string;
  metadata: {
    documentData?: Record<string, any>; // Amount, vendor name, etc.
    currentApprovalStage?: number;
    totalApprovalStages?: number;
    approvalStageName?: string;
    relatedDocumentNumber?: string; // e.g., PO number for GRN
  };
}

export interface TaskFilters {
  status?: TaskStatus;
  priority?: TaskPriority;
  taskType?: TaskType;
  assignedRole?: string;
  dueDate?: {
    start?: string;
    end?: string;
  };
}

export interface TaskStats {
  totalTasks: number;
  pendingTasks: number;
  overdueTasks: number;
  completedTasks: number;
  highPriorityTasks: number;
  byType: Record<TaskType, number>;
  byPriority: Record<TaskPriority, number>;
}

export interface TaskWithDocument extends Task {
  documentDetails: {
    amount?: number;
    vendor?: string;
    department?: string;
    description?: string;
  };
}
