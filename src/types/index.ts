/**
 * Consolidated Type Exports
 * All types are now organized in separate files:
 * - auth.ts: Authentication and account types
 * - api.ts: API response types
 * - workflow.ts: Workflow document types
 * - settings.ts: Settings and configuration types
 * - system-status.ts: System status types
 * - forms.ts: Form-related types
 */

// Re-export all auth-related types
export type {
  UserType,
  User,
  AuthSession,
  Permission,
  SessionResponse,
} from "./auth";

// Re-export all workflow types
export type {
  WorkflowDocumentType,
  DocumentStatus,
  ApprovalAction,
  UserRole,
  WorkflowPermission,
  WorkflowDocument,
  WorkflowStep,
  Approver,
  ApprovalLogEntry,
  Attachment,
  PurchaseOrderItem,
  PurchaseOrder,
  PaymentVoucher,
  RequisitionItem,
  RequisitionForm,
  PaginatedResponse,
  ReversalBehavior,
  ApprovalStageConfig,
  ApprovalRecord,
  ApprovalState,
  DocumentApprovalConfig,
  ApproveDocumentRequest,
  ApproveDocumentResponse,
  ReverseDocumentRequest,
  ReverseDocumentResponse,
  SearchFilters,
} from "./workflow";

// Re-export all settings types
export type {
  Currency,
  AccountTier,
  SignupSettings,
  SettingsData,
} from "./settings";

// Re-export form types
export type {
  ChangePassword,
  ChangePasswordRequest,
  ChangePasswordResponse,
} from "./forms";

// Additional common types
export type CurrencyKey = "USD" | "ZMW";

export type Pagination = {
  page: number;
  limit: number;
  total: number;
  totalPages: number;
  hasNext: boolean;
  hasPrev: boolean;
};

export type APIResponse = {
  success: boolean;
  message: string;
  data: any;
  status: number;
  [x: string]: unknown;
};

export type ErrorState = {
  status: boolean;
  message: string;
  type?: "error" | "success" | "info" | "warning";
  [x: string]: unknown;
};
