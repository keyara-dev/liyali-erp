/**
 * Centralized status badge configuration
 * Single source of truth for all status-related badge styling across the application
 */

type BadgeVariant = 'default' | 'secondary' | 'destructive' | 'outline'

// Document Workflow Status
export type DocumentStatus = 'DRAFT' | 'SUBMITTED' | 'IN_APPROVAL' | 'APPROVED' | 'REJECTED' | 'REVERSED'

// Activity/Action Status
export type ActivityAction = 'created' | 'approved' | 'rejected' | 'submitted' | 'edited' | 'viewed' | 'deleted'

// Activity Execution Status
export type ExecutionStatus = 'success' | 'failed' | 'pending'

// Approval Status
export type ApprovalStatus = 'APPROVED' | 'REJECTED' | 'PENDING'

// Compliance Status
export type ComplianceStatus = 'compliant' | 'non-compliant' | 'pending'

// User Role
export type UserRole = 'REQUESTER' | 'DEPARTMENT_MANAGER' | 'FINANCE_OFFICER' | 'DIRECTOR' | 'CFO' | 'COMPLIANCE_OFFICER' | 'ADMIN'

// Service/System Health Status
export type HealthStatus = 'healthy' | 'issues' | 'down'

/**
 * Document Status Configuration
 * Used for requisitions, purchase orders, GRNs, and payment vouchers
 */
export const DOCUMENT_STATUS_CONFIG: Record<DocumentStatus, {
  variant: BadgeVariant
  label: string
  description?: string
}> = {
  DRAFT: {
    variant: 'outline',
    label: 'Draft',
    description: 'Document is in draft status and can be edited'
  },
  SUBMITTED: {
    variant: 'secondary',
    label: 'Submitted',
    description: 'Document has been submitted for review'
  },
  IN_APPROVAL: {
    variant: 'secondary',
    label: 'In Review',
    description: 'Document is pending approval'
  },
  APPROVED: {
    variant: 'default',
    label: 'Approved',
    description: 'Document has been approved'
  },
  REJECTED: {
    variant: 'destructive',
    label: 'Rejected',
    description: 'Document has been rejected'
  },
  REVERSED: {
    variant: 'secondary',
    label: 'Reversed',
    description: 'Document has been reversed'
  }
}

/**
 * Activity Action Configuration
 * Used for logging actions performed in the system
 */
export const ACTIVITY_ACTION_CONFIG: Record<ActivityAction, {
  variant: BadgeVariant
  label: string
}> = {
  created: {
    variant: 'default',
    label: 'Created'
  },
  approved: {
    variant: 'default',
    label: 'Approved'
  },
  rejected: {
    variant: 'destructive',
    label: 'Rejected'
  },
  submitted: {
    variant: 'default',
    label: 'Submitted'
  },
  edited: {
    variant: 'outline',
    label: 'Edited'
  },
  viewed: {
    variant: 'outline',
    label: 'Viewed'
  },
  deleted: {
    variant: 'destructive',
    label: 'Deleted'
  }
}

/**
 * Execution Status Configuration
 * Used for activity/task execution status
 */
export const EXECUTION_STATUS_CONFIG: Record<ExecutionStatus, {
  variant: BadgeVariant
  label: string
}> = {
  success: {
    variant: 'default',
    label: 'Success'
  },
  failed: {
    variant: 'destructive',
    label: 'Failed'
  },
  pending: {
    variant: 'outline',
    label: 'Pending'
  }
}

/**
 * Approval Status Configuration
 * Used for approval workflow steps
 */
export const APPROVAL_STATUS_CONFIG: Record<ApprovalStatus, {
  variant: BadgeVariant
  label: string
}> = {
  APPROVED: {
    variant: 'default',
    label: 'Approved'
  },
  REJECTED: {
    variant: 'destructive',
    label: 'Rejected'
  },
  PENDING: {
    variant: 'secondary',
    label: 'Pending'
  }
}

/**
 * Compliance Status Configuration
 * Used for compliance tracking
 */
export const COMPLIANCE_STATUS_CONFIG: Record<ComplianceStatus, {
  variant: BadgeVariant
  label: string
}> = {
  compliant: {
    variant: 'default',
    label: 'Compliant'
  },
  'non-compliant': {
    variant: 'destructive',
    label: 'Non-Compliant'
  },
  pending: {
    variant: 'outline',
    label: 'Pending'
  }
}

/**
 * User Role Configuration
 * Used for user role badges
 */
export const USER_ROLE_CONFIG: Record<UserRole, {
  variant: BadgeVariant
  label: string
}> = {
  REQUESTER: {
    variant: 'outline',
    label: 'Requester'
  },
  DEPARTMENT_MANAGER: {
    variant: 'secondary',
    label: 'Department Manager'
  },
  FINANCE_OFFICER: {
    variant: 'default',
    label: 'Finance Officer'
  },
  DIRECTOR: {
    variant: 'default',
    label: 'Director'
  },
  CFO: {
    variant: 'default',
    label: 'CFO'
  },
  COMPLIANCE_OFFICER: {
    variant: 'secondary',
    label: 'Compliance Officer'
  },
  ADMIN: {
    variant: 'destructive',
    label: 'Admin'
  }
}

/**
 * Health Status Configuration
 * Used for system and service health monitoring
 */
export const HEALTH_STATUS_CONFIG: Record<HealthStatus, {
  variant: BadgeVariant
  label: string
}> = {
  healthy: {
    variant: 'default',
    label: '✓ Healthy'
  },
  issues: {
    variant: 'outline',
    label: '⚠ Issues'
  },
  down: {
    variant: 'destructive',
    label: '✗ Down'
  }
}

/**
 * Helper function to get document status variant
 */
export function getDocumentStatusVariant(status: string): BadgeVariant {
  const config = DOCUMENT_STATUS_CONFIG[status as DocumentStatus]
  return config?.variant || 'outline'
}

/**
 * Helper function to get document status label
 */
export function getDocumentStatusLabel(status: string): string {
  const config = DOCUMENT_STATUS_CONFIG[status as DocumentStatus]
  return config?.label || status
}

/**
 * Helper function to get activity action variant
 */
export function getActivityActionVariant(action: string): BadgeVariant {
  const config = ACTIVITY_ACTION_CONFIG[action as ActivityAction]
  return config?.variant || 'outline'
}

/**
 * Helper function to get activity action label
 */
export function getActivityActionLabel(action: string): string {
  const config = ACTIVITY_ACTION_CONFIG[action as ActivityAction]
  return config?.label || action
}

/**
 * Helper function to get execution status variant
 */
export function getExecutionStatusVariant(status: string): BadgeVariant {
  const config = EXECUTION_STATUS_CONFIG[status as ExecutionStatus]
  return config?.variant || 'outline'
}

/**
 * Helper function to get approval status variant
 */
export function getApprovalStatusVariant(status: string): BadgeVariant {
  const config = APPROVAL_STATUS_CONFIG[status as ApprovalStatus]
  return config?.variant || 'outline'
}

/**
 * Helper function to get compliance status variant
 */
export function getComplianceStatusVariant(status: string): BadgeVariant {
  const config = COMPLIANCE_STATUS_CONFIG[status as ComplianceStatus]
  return config?.variant || 'outline'
}

/**
 * Helper function to get user role variant
 */
export function getUserRoleVariant(role: string): BadgeVariant {
  const config = USER_ROLE_CONFIG[role as UserRole]
  return config?.variant || 'outline'
}

/**
 * Helper function to get health status variant
 */
export function getHealthStatusVariant(status: string): BadgeVariant {
  const config = HEALTH_STATUS_CONFIG[status as HealthStatus]
  return config?.variant || 'outline'
}
