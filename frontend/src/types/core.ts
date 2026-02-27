/**
 * Core Types
 * Fundamental types used across the entire application
 * These are the single source of truth for shared interfaces
 */

// ================== CORE API TYPES ==================

export interface APIResponse<T = any> {
  success: boolean;
  message?: string;
  data?: T;
  error?: string;
  pagination?: PaginationMeta;
  status?: number; // HTTP status code
  statusCode?: number; // Alias for status
  statusText?: string; // HTTP status text
  errors?: Record<string, string[]>;
}

export interface PaginationMeta {
  page: number;
  pageSize?: number; // From api.ts
  limit?: number; // From common.ts
  total: number;
  totalPages: number;
  hasNext: boolean;
  hasPrev: boolean;
  // Aliases for backward compatibility
  page_size?: number; // Alias for pageSize/limit
  totalCount?: number; // Alias for total
  total_pages?: number; // Alias for totalPages
  has_next?: boolean; // Alias for hasNext
  has_prev?: boolean; // Alias for hasPrev
}

export interface ListResponse<T> {
  success: boolean;
  data: T[];
  pagination?: PaginationMeta;
  meta?: PaginationMeta; // Alias for pagination
}

export interface DetailResponse<T> {
  success: boolean;
  data: T;
}

export interface MessageResponse {
  success: boolean;
  message: string;
}

export interface PaginatedResponse<T> {
  data: T[];
  pagination: PaginationMeta;
}

// ================== CORE ENUMS ==================

export type DocumentStatus =
  | "draft"
  | "pending"
  | "approved"
  | "rejected"
  | "completed"
  | "cancelled"
  | "submitted"
  | "paid" // For Payment Vouchers
  | "fulfilled" // For Purchase Orders
  // Legacy compatibility
  | "DRAFT"
  | "SUBMITTED"
  | "IN_REVIEW"
  | "APPROVED"
  | "REJECTED"
  | "PENDING"
  | "COMPLETED"
  | "CANCELLED";

export type Priority =
  | "low"
  | "medium"
  | "high"
  | "urgent"
  // Legacy compatibility
  | "URGENT"
  | "HIGH"
  | "LOW"
  | "MEDIUM";

export type ApprovalStatus =
  | "pending"
  | "approved"
  | "rejected"
  | "cancelled"
  | "reversed"
  // Legacy compatibility
  | "PENDING"
  | "APPROVED"
  | "REJECTED"
  | "REVERSED";

export type PaymentMethod =
  | "bank_transfer"
  | "cash"
  // Legacy compatibility
  | "BANK_TRANSFER"
  | "CASH";

// Unified User Role type (combining UserRole and UserType)
export type UserRole =
  | "admin"
  | "approver"
  | "requester"
  | "finance"
  | "viewer"
  | "department_manager"
  | "finance_manager"
  | "finance_officer"
  | "director"
  | "cfo"
  | "compliance_officer"
  | "ceo"
  | "superadmin"; // Add superadmin role

// Alias for backward compatibility
export type UserType = UserRole;

export type ItemCondition = "good" | "damaged" | "missing";

export type QualityIssueType =
  | "damaged"
  | "missing"
  | "wrong_item"
  | "quality_issue";

export type QualityIssueSeverity = "low" | "medium" | "high";

// ================== CORE USER & ORGANIZATION TYPES ==================

export interface User {
  id: string;
  email: string;
  name: string;
  role: UserRole;
  role_id?: string;
  role_name?: string;
  active?: boolean;
  lastLogin?: string | Date;
  currentOrganizationId?: string;
  currentOrganization?: Organization;
  isSuperAdmin?: boolean;
  createdAt?: string | Date;
  updatedAt?: string | Date;

  // Extended fields for UI compatibility
  permissions?: string[];
  preferences?: {
    avatar?: string;
    department?: string;
    language?: string;
    theme?: string;
    timezone?: string;
    emailNotifications?: boolean;
    pushNotifications?: boolean;
    activityNotifications?: boolean;
  };
  first_name?: string;
  last_name?: string;
  mfa_enabled?: boolean;
  department?: string;
  department_id?: string;
  phone?: string;
  username?: string;
  is_ldap_user?: boolean;
  is_active?: boolean;
  avatar?: string;
  created_at?: Date | string;
  updated_at?: Date | string;
  last_login?: Date | string;
  expiresAt?: Date | string;
  token?: string; // For session compatibility
}

export interface Organization {
  id: string;
  name: string;
  slug: string;
  description?: string;
  logoUrl?: string;
  tagline?: string;
  primaryColor?: string;
  active: boolean;
  tier: string;
  createdBy?: string;
  createdAt: string | Date;
  updatedAt?: string | Date;
}

export interface Permission {
  id: string;
  name: string;
  description?: string;
  resource?: string;
  action?: string;
}

// ================== CORE APPROVAL TYPES ==================

export interface ApprovalRecord {
  approverId: string;
  approverName: string;
  status: ApprovalStatus;
  comments?: string;
  signature?: string;
  approvedAt?: Date;

  // Extended fields for UI compatibility (now supported by backend)
  stageNumber?: number;
  stageName?: string;
  assignedTo?: string;
  assignedRole?: string;
  actionTakenBy?: string;
  actionTakenByRole?: string;
  actionTakenAt?: Date;
  remarks?: string;
}

export interface ApprovalTask {
  id: string;
  organizationId?: string;
  documentId: string;
  documentType: string;
  documentNumber?: string;
  approverId: string;
  approverName?: string;
  assignedTo?: string;
  status: ApprovalStatus | "pending" | "claimed" | "completed" | "expired";
  stage: number;
  comments?: string;
  signature?: string;
  approvedBy?: string;
  approvedAt?: Date;
  rejectedBy?: string;
  rejectedAt?: Date;
  rejectionReason?: string;
  priority?: Priority | string;
  dueAt?: Date;
  taskType?: string;
  title?: string;
  workflowId?: string;
  workflowName?: string;
  stageName?: string;
  importance?: "low" | "medium" | "high";
  createdAt: Date;
  updatedAt: Date;

  // Assignment fields (from WorkflowTask)
  assignedRole?: string; // Required role to claim/approve
  assignedUserId?: string; // Specific user assigned (after reassignment)

  // Claiming fields (from WorkflowTask)
  claimedBy?: string; // User ID who claimed the task
  claimerName?: string; // Name of the user who claimed the task
  claimedAt?: Date; // When the task was claimed
  claimExpiry?: Date; // When the claim expires
  version?: number; // For optimistic locking

  // Legacy compatibility fields
  entityId?: string; // Maps to documentId
  entityType?: string; // Maps to documentType
  entityNumber?: string; // Maps to documentNumber
  stageIndex?: number; // Maps to stage
  approverUserId?: string; // Maps to approverId
  dueDate?: Date; // Maps to dueAt
  actionDate?: Date; // Maps to updatedAt
}

export interface ActionHistoryEntry {
  id: string;
  action: string;
  performedBy: string;
  performedByName: string;
  performedByRole?: string; // Role of the user who performed the action
  timestamp: Date;
  performedAt?: Date; // Alias for timestamp
  changes?: Record<string, any>;
  comments?: string;
  actionType?: string;
  newStatus?: string;
  previousStatus?: string;
  remarks?: string;
  stageNumber?: number;
  stageName?: string;
  changedFields?: string[];
}

// ================== CORE REQUEST TYPES ==================

export interface ApproveTaskRequest {
  comments?: string;
  signature: string;
  stageNumber?: number;
}

export interface RejectTaskRequest {
  remarks: string;
  comments?: string;
  signature?: string;
  returnTo?: "original_submitter" | "previous_stage" | string;
}

export interface ReassignTaskRequest {
  newApproverId: string;
  reason: string;
}

export interface ApproveDocumentRequest {
  comments: string;
  signature: string;
}

export interface RejectDocumentRequest {
  remarks: string;
  signature: string;
}

export interface ReassignDocumentRequest {
  newApproverId: string;
  reason: string;
}

// ================== CORE UTILITY TYPES ==================

export interface SearchFilters {
  documentNumber: string;
  documentType: "all" | string;
  status: "all" | string;
  startDate: string;
  endDate: string;
}

export interface ValidationResult {
  valid: boolean;
  error?: string;
}

export type BadgeVariant =
  | "default"
  | "secondary"
  | "destructive"
  | "outline"
  | "ghost";

export type HealthStatus = "healthy" | "issues" | "down";

export type AllocationStatus = "under" | "full" | "over";

// ================== VENDOR & CATEGORY TYPES ==================

export interface Category {
  id: string;
  organizationId: string;
  name: string;
  description?: string;
  active: boolean;
  createdAt: Date;
  updatedAt: Date;
}

export interface Vendor {
  id: string;
  vendorCode: string;
  name: string;
  email?: string;
  phone?: string;
  country?: string;
  city?: string;
  bankAccount?: string;
  taxId?: string;
  active: boolean;
  createdBy?: string;
  createdAt: Date;
  updatedAt: Date;
}
