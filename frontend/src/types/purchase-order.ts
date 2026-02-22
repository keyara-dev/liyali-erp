/**
 * Purchase Order Types
 * Aligned with backend models and database schema
 */

// ============================================================================
// CORE PURCHASE ORDER TYPES
// ============================================================================

export interface POItem {
  id?: string;
  description: string;
  quantity: number;
  unitPrice: number;
  amount: number;
  itemNumber?: string;
  itemCode?: string;
  totalPrice?: number; // Alias for amount
  unit?: string;
  category?: string;
  notes?: string;
}

export interface PurchaseOrder {
  // Core fields
  id: string;
  organizationId: string;
  documentNumber: string;
  vendorId: string;
  vendor?: any;
  vendorName: string;
  status: PurchaseOrderStatus; // draft, pending, approved, rejected, fulfilled, completed, cancelled
  items: POItem[];
  totalAmount: number;
  currency: string;
  deliveryDate: Date;
  approvalStage: number;
  approvalHistory: any[];
  linkedRequisition: string;
  createdAt: Date;
  updatedAt: Date;

  // Business requirement fields
  description: string;
  department: string;
  departmentId: string;
  requiredByDate: Date; // Same as deliveryDate
  priority: string;
  budgetCode: string;
  costCenter: string;
  projectCode: string;
  sourceRequisitionId: string; // Same as linkedRequisition
  subtotal: number;
  tax: number;
  total: number; // Same as totalAmount
  createdBy: string;
  ownerId: string; // Same as createdBy
  glCode: string;
  title: string;

  // UI compatibility fields
  currentStage?: number;
  actionHistory?: any[];
  metadata?: Record<string, any>;
  type?: string;
  createdByUser?: any;
  approvalChain?: any[]; // For PDF generation
  requestedBy?: string; // For PV creation
  requestedByName?: string; // For PV creation
  requestedByRole?: string; // For PV creation
}

// ============================================================================
// REQUEST TYPES
// ============================================================================

export interface CreatePurchaseOrderRequest {
  vendorId: string;
  vendorName?: string;
  items: POItem[];
  totalAmount: number;
  currency: string;
  deliveryDate: Date;
  requiredByDate?: Date; // Alias for deliveryDate
  linkedRequisition: string;

  // Business requirement fields
  description: string;
  department: string;
  departmentId: string;
  priority: string;
  budgetCode: string;
  costCenter: string;
  projectCode: string;
  title: string;
  glCode: string;
  subtotal: number;
  tax: number;
  createdBy: string;
  createdByName?: string;
  createdByRole?: string;
  sourceRequisitionId?: string;
}

export interface UpdatePurchaseOrderRequest {
  purchaseOrderId: string;
  poId?: string; // Alias for purchaseOrderId
  vendorId?: string;
  vendorName?: string;
  items?: POItem[];
  totalAmount?: number;
  currency?: string;
  deliveryDate?: Date;
  requiredByDate?: Date; // Alias for deliveryDate
  description?: string;
  title?: string;
  priority?: string;
  budgetCode?: string;
  costCenter?: string;
  projectCode?: string;
}

export interface SubmitPurchaseOrderRequest {
  purchaseOrderId: string;
  poId?: string; // Alias for purchaseOrderId
  workflowId: string; // REQUIRED - Workflow to use for approval
  submittingUserId: string;
  submittedBy?: string; // Alias for submittingUserId
  submittedByName: string;
  submittedByRole: string;
  comments?: string;
}

export interface ApprovePurchaseOrderRequest {
  purchaseOrderId: string;
  poId?: string; // Alias for purchaseOrderId
  approvingUserId: string;
  approvingUserName: string;
  approvingUserRole: string;
  signature: string;
  comments?: string;
  stageNumber?: number;
}

export interface RejectPurchaseOrderRequest {
  purchaseOrderId: string;
  poId?: string; // Alias for purchaseOrderId
  rejectingUserId: string;
  rejectingUserName: string;
  rejectingUserRole: string;
  remarks: string;
  signature: string;
  comments?: string;
}

// ============================================================================
// STATISTICS TYPES
// ============================================================================

export interface PurchaseOrderStats {
  total: number;
  draft: number;
  pending: number;
  approved: number;
  fulfilled: number;
  thisMonth: number;
  totalAmount: number;
}

// ============================================================================
// TYPE ALIASES
// ============================================================================

export type PurchaseOrderStatus =
  | "draft"
  | "pending"
  | "approved"
  | "rejected"
  | "fulfilled"
  | "completed"
  | "cancelled";
