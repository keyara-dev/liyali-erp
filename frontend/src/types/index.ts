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

// Re-export all custom workflow types
export type {
  WorkflowEntityType,
  ApproverAssignmentType,
  StageTransition,
  WorkflowStage,
  CustomWorkflow,
  WorkflowAssignment,
  StageExecution,
  StageAssignment,
  WorkflowDefault,
  WorkflowStats,
  WorkflowValidationError,
  CreateWorkflowRequest,
  UpdateWorkflowRequest,
  AssignWorkflowRequest,
  ApproveStageRequest,
  RejectStageRequest,
  ReassignStageRequest,
  ReverseStageRequest,
} from "./custom-workflow";

// Re-export all settings types
export type {
  Currency,
  AccountTier,
  SettingsData,
} from "./settings";

// Re-export all dashboard types
export type {
  DashboardMetrics,
  SignupSettings,
  SignupAnalytics,
} from "./dashboard";

// Re-export form types
export type {
  ChangePassword,
  ChangePasswordRequest,
  ChangePasswordResponse,
} from "./forms";

// Re-export currency types
export type {
  CurrencyData,
} from "./currencies";

// Re-export event host types
export type {
  EventHost,
} from "./event-hosts";

// Re-export premium types
export type {
  PremiumPlan,
} from "./premium";

// Re-export premium config types
export type {
  PremiumConfig,
} from "./premium-config";

// Re-export WhatsApp types
export type {
  WhatsAppConfig,
} from "./whatsapp";

// Re-export user management types
export type {
  UserRoleAssignment,
} from "./user-management";

// Additional common types
export type CurrencyKey = "USD" | "ZMW";

export type Pagination = {
  page: number;
  limit?: number;
  page_size?: number;
  total?: number;
  totalCount?: number;
  total_pages?: number;
  totalPages?: number;
  has_next?: boolean;
  hasNext?: boolean;
  has_prev?: boolean;
  hasPrev?: boolean;
};

export type APIResponse<T = any> = {
  success: boolean;
  message: string;
  data?: T | null;
  status: number;
  [x: string]: unknown;
};

export type ErrorState = {
  status: boolean;
  message: string;
  type?: "error" | "success" | "info" | "warning";
  [x: string]: unknown;
};

// Re-export task types
export type {
  Task,
  TaskType,
  TaskPriority,
  TaskStatus,
  TaskFilters,
  TaskStats,
  TaskWithDocument,
  ApprovalTask,
  ApprovalTaskDetail,
} from "./tasks";

// Re-export notification types
export type {
  NotificationType,
  QuickActionType,
  NotificationImportance,
  QuickAction,
  Notification,
  NotificationPreferences,
  GetNotificationsRequest,
  GetNotificationsResponse,
  CreateNotificationRequest,
  CreateNotificationResponse,
  MarkNotificationReadRequest,
  MarkNotificationReadResponse,
  MarkAllNotificationsReadRequest,
  MarkAllNotificationsReadResponse,
  DeleteNotificationRequest,
  DeleteNotificationResponse,
  GetUnreadCountRequest,
  GetUnreadCountResponse,
  GetNotificationPreferencesRequest,
  GetNotificationPreferencesResponse,
  UpdateNotificationPreferencesRequest,
  UpdateNotificationPreferencesResponse,
  TaskAssignedEvent,
  TaskReassignedEvent,
  TaskApprovedEvent,
  TaskRejectedEvent,
  WorkflowCompleteEvent,
} from "./notifications";
