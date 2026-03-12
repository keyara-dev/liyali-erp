/**
 * Centralized status badge configuration
 * Single source of truth for all status-related badge styling across the application
 */

type BadgeVariant =
  | "default"
  | "secondary"
  | "destructive"
  | "outline"
  | "warning"
  | "info"
  | "success";

// Document Workflow Status
export type DocumentStatus =
  | "draft"
  | "submitted"
  | "pending"
  | "in_review"
  | "revision"
  | "approved"
  | "success"
  | "rejected"
  | "reversed";

// Activity/Action Status
export type ActivityAction =
  | "created"
  | "approved"
  | "rejected"
  | "submitted"
  | "edited"
  | "viewed"
  | "deleted";

// Activity Execution Status
export type ExecutionStatus =
  | "success"
  | "failed"
  | "pending"
  | "claimed"
  | "completed"
  | "in_progress";

// Approval Status
export type ApprovalStatus =
  | "approved"
  | "rejected"
  | "pending"
  | "in_progress";

// Compliance Status
export type ComplianceStatus = "compliant" | "non-compliant" | "pending";

// User Role
export type UserRole = "admin" | "approver" | "finance" | "requester";

// Service/System Health Status
export type HealthStatus = "healthy" | "issues" | "down";

/**
 * Document Status Configuration
 * Used for requisitions, purchase orders, GRNs, and payment vouchers
 */
export const DOCUMENT_STATUS_CONFIG: Record<
  DocumentStatus,
  {
    variant: BadgeVariant;
    label: string;
    description?: string;
  }
> = {
  draft: {
    variant: "outline",
    label: "Draft",
    description: "Document is in draft status and can be edited",
  },
  submitted: {
    variant: "info",
    label: "Submitted",
    description: "Document has been submitted for review",
  },
  in_review: {
    variant: "warning",
    label: "In Review",
    description: "Document is pending approval",
  },
  pending: {
    variant: "warning",
    label: "In Review",
    description: "Document is pending approval",
  },
  approved: {
    variant: "success",
    label: "Approved",
    description: "Document has been approved",
  },
  success: {
    variant: "success",
    label: "Success",
    description: "Document has been processed successfully",
  },
  revision: {
    variant: "warning",
    label: "Revision",
    description: "Document returned for revision at a previous approval stage",
  },
  rejected: {
    variant: "destructive",
    label: "Rejected",
    description: "Document has been rejected",
  },
  reversed: {
    variant: "secondary",
    label: "Reversed",
    description: "Document has been reversed",
  },
};

/**
 * Activity Action Configuration
 * Used for logging actions performed in the system
 */
export const ACTIVITY_ACTION_CONFIG: Record<
  ActivityAction,
  {
    variant: BadgeVariant;
    label: string;
  }
> = {
  created: {
    variant: "info",
    label: "Created",
  },
  approved: {
    variant: "success",
    label: "Approved",
  },
  rejected: {
    variant: "destructive",
    label: "Rejected",
  },
  submitted: {
    variant: "info",
    label: "Submitted",
  },
  edited: {
    variant: "warning",
    label: "Edited",
  },
  viewed: {
    variant: "outline",
    label: "Viewed",
  },
  deleted: {
    variant: "destructive",
    label: "Deleted",
  },
};

/**
 * Execution Status Configuration
 * Used for activity/task execution status
 */
export const EXECUTION_STATUS_CONFIG: Record<
  ExecutionStatus,
  {
    variant: BadgeVariant;
    label: string;
  }
> = {
  success: {
    variant: "success",
    label: "Success",
  },
  failed: {
    variant: "destructive",
    label: "Failed",
  },
  pending: {
    variant: "info",
    label: "Pending",
  },
  claimed: {
    variant: "warning",
    label: "Claimed",
  },
  completed: {
    variant: "success",
    label: "Completed",
  },
  in_progress: {
    variant: "warning",
    label: "In Progress",
  },
};

/**
 * Approval Status Configuration
 * Used for approval workflow steps
 */
export const APPROVAL_STATUS_CONFIG: Record<
  ApprovalStatus,
  {
    variant: BadgeVariant;
    label: string;
  }
> = {
  approved: {
    variant: "success",
    label: "Approved",
  },
  rejected: {
    variant: "destructive",
    label: "Rejected",
  },
  pending: {
    variant: "warning",
    label: "Pending",
  },
  in_progress: {
    variant: "warning",
    label: "In Progress",
  },
};

/**
 * Compliance Status Configuration
 * Used for compliance tracking
 */
export const COMPLIANCE_STATUS_CONFIG: Record<
  ComplianceStatus,
  {
    variant: BadgeVariant;
    label: string;
  }
> = {
  compliant: {
    variant: "success",
    label: "Compliant",
  },
  "non-compliant": {
    variant: "destructive",
    label: "Non-Compliant",
  },
  pending: {
    variant: "warning",
    label: "Pending",
  },
};

/**
 * User Role Configuration
 * Used for user role badges
 */
export const USER_ROLE_CONFIG: Record<
  UserRole,
  {
    variant: BadgeVariant;
    label: string;
  }
> = {
  admin: {
    variant: "destructive",
    label: "Admin",
  },
  approver: {
    variant: "default",
    label: "Approver",
  },
  finance: {
    variant: "default",
    label: "Finance",
  },
  requester: {
    variant: "outline",
    label: "Requester",
  },
};

/**
 * Health Status Configuration
 * Used for system and service health monitoring
 */
export const HEALTH_STATUS_CONFIG: Record<
  HealthStatus,
  {
    variant: BadgeVariant;
    label: string;
  }
> = {
  healthy: {
    variant: "default",
    label: "✓ Healthy",
  },
  issues: {
    variant: "outline",
    label: "⚠ Issues",
  },
  down: {
    variant: "destructive",
    label: "✗ Down",
  },
};

/**
 * Helper function to get document status variant
 */
export function getDocumentStatusVariant(status: string): BadgeVariant {
  const config = DOCUMENT_STATUS_CONFIG[status as DocumentStatus];
  return config?.variant || "outline";
}

/**
 * Helper function to get document status label
 */
export function getDocumentStatusLabel(status: string): string {
  const config = DOCUMENT_STATUS_CONFIG[status as DocumentStatus];
  return config?.label || status;
}

/**
 * Helper function to get activity action variant
 */
export function getActivityActionVariant(action: string): BadgeVariant {
  const config = ACTIVITY_ACTION_CONFIG[action as ActivityAction];
  return config?.variant || "outline";
}

/**
 * Helper function to get activity action label
 */
export function getActivityActionLabel(action: string): string {
  const config = ACTIVITY_ACTION_CONFIG[action as ActivityAction];
  return config?.label || action;
}

/**
 * Helper function to get execution status variant
 */
export function getExecutionStatusVariant(status: string): BadgeVariant {
  const config = EXECUTION_STATUS_CONFIG[status as ExecutionStatus];
  return config?.variant || "outline";
}

/**
 * Helper function to get approval status variant
 */
export function getApprovalStatusVariant(status: string): BadgeVariant {
  const config = APPROVAL_STATUS_CONFIG[status as ApprovalStatus];
  return config?.variant || "outline";
}

/**
 * Helper function to get compliance status variant
 */
export function getComplianceStatusVariant(status: string): BadgeVariant {
  const config = COMPLIANCE_STATUS_CONFIG[status as ComplianceStatus];
  return config?.variant || "outline";
}

/**
 * Helper function to get user role variant
 */
export function getUserRoleVariant(role: string): BadgeVariant {
  const config = USER_ROLE_CONFIG[role as UserRole];
  return config?.variant || "outline";
}

/**
 * Helper function to get health status variant
 */
export function getHealthStatusVariant(status: string): BadgeVariant {
  const config = HEALTH_STATUS_CONFIG[status as HealthStatus];
  return config?.variant || "outline";
}
