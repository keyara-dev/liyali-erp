import {
  DocumentApprovalConfig,
  ApprovalStageConfig,
  ApprovalState,
  WorkflowDocumentType,
} from '@/types/workflow';

/**
 * Dynamic Approval Configuration System
 *
 * This system manages configurable approval workflows for different document types.
 * Each document type can have a different number of approval stages with different rules.
 *
 * Supports fallback defaults when configuration is missing.
 */

// ============================================================================
// CONFIGURATION DEFINITIONS
// ============================================================================

/**
 * Requisition: 4-stage approval workflow
 * Department Head → Principal Officer → Director Finance → Procurement Officer
 */
const requisitionConfig: DocumentApprovalConfig = {
  documentType: 'REQUISITION',
  configVersion: '1.0',
  effectiveDate: new Date('2024-01-01'),
  description: 'Standard 4-stage requisition approval workflow',
  totalStages: 4,
  createdAt: new Date('2024-01-01'),
  createdBy: 'system',

  approvalStages: [
    {
      stageNumber: 1,
      stageName: 'Department Head Review',
      description: 'Department manager reviews requisition',
      requiredRole: 'DEPARTMENT_MANAGER',
      alternativeRoles: ['DEPARTMENT_HEAD'],
      canReverse: true,
      reversalBehavior: 'BACK_TO_CREATOR',
      requiresComments: false,
    },
    {
      stageNumber: 2,
      stageName: 'Principal Officer Review',
      description: 'Executive reviews for strategic alignment',
      requiredRole: 'PRINCIPAL_OFFICER',
      alternativeRoles: ['CFO'],
      canReverse: true,
      reversalBehavior: 'BACK_TO_CREATOR',
      requiresComments: true,
    },
    {
      stageNumber: 3,
      stageName: 'Finance Director Review',
      description: 'Finance director checks budget and accounting',
      requiredRole: 'DIRECTOR',
      alternativeRoles: ['FINANCE_OFFICER'],
      canReverse: true,
      reversalBehavior: 'BACK_TO_CREATOR',
      requiresComments: false,
      requiredValidations: ['budgetAvailable', 'accountCodeValid'],
    },
    {
      stageNumber: 4,
      stageName: 'Procurement Officer Processing',
      description: 'Procurement officer adds supplier and creates PO',
      requiredRole: 'ADMIN',
      canReverse: false,
      reversalBehavior: 'BACK_TO_CREATOR',
      requiresComments: false,
      onApprovalActions: {
        createAuditLog: true,
      },
    },
  ],

  fallbackStages: [
    {
      stageNumber: 1,
      stageName: 'Default Head Review',
      requiredRole: 'DEPARTMENT_MANAGER',
      canReverse: true,
      reversalBehavior: 'BACK_TO_CREATOR',
    },
    {
      stageNumber: 2,
      stageName: 'Default Director Review',
      requiredRole: 'DIRECTOR',
      canReverse: true,
      reversalBehavior: 'BACK_TO_CREATOR',
    },
  ],

  allowConcurrentApprovals: false,
  allowMultipleReversals: true,
  requireFinalSignoff: true,
};

/**
 * Purchase Order: 4-stage approval with reversals to Procurement Officer
 * Department Head → Auditor → Director Finance → Principal Officer
 */
const purchaseOrderConfig: DocumentApprovalConfig = {
  documentType: 'PURCHASE_ORDER',
  configVersion: '1.0',
  effectiveDate: new Date('2024-01-01'),
  description: '4-stage PO approval with reversals to Procurement Officer',
  totalStages: 4,
  createdAt: new Date('2024-01-01'),
  createdBy: 'system',

  approvalStages: [
    {
      stageNumber: 1,
      stageName: 'Department Head Approval',
      description: 'Department head approves procurement need',
      requiredRole: 'DEPARTMENT_MANAGER',
      canReverse: true,
      reversalBehavior: 'TO_SPECIFIC_USER',
      reversalTargetRole: 'ADMIN',
      reversalResetsPreviousStages: false,
      requiresComments: false,
    },
    {
      stageNumber: 2,
      stageName: 'Auditor Review',
      description: 'Internal auditor reviews for compliance',
      requiredRole: 'COMPLIANCE_OFFICER',
      canReverse: true,
      reversalBehavior: 'TO_SPECIFIC_USER',
      reversalTargetRole: 'ADMIN',
      reversalResetsPreviousStages: false,
      requiresComments: true,
      requiredValidations: ['complianceCheck', 'vendorApproval'],
    },
    {
      stageNumber: 3,
      stageName: 'Director Finance Approval',
      description: 'Finance director approves budget allocation',
      requiredRole: 'DIRECTOR',
      canReverse: true,
      reversalBehavior: 'TO_SPECIFIC_USER',
      reversalTargetRole: 'ADMIN',
      reversalResetsPreviousStages: false,
      requiresComments: false,
      requiredValidations: ['budgetAvailable'],
    },
    {
      stageNumber: 4,
      stageName: 'Principal Officer Final Approval',
      description: 'Executive final approval and order authorization',
      requiredRole: 'PRINCIPAL_OFFICER',
      canReverse: true,
      reversalBehavior: 'TO_SPECIFIC_USER',
      reversalTargetRole: 'ADMIN',
      reversalResetsPreviousStages: false,
      requiresComments: false,
      onApprovalActions: {
        createAuditLog: true,
      },
    },
  ],

  allowConcurrentApprovals: false,
  allowMultipleReversals: true,
  requireFinalSignoff: true,
};

/**
 * Goods Received Note: 1-stage simple confirmation
 * Stores Officer confirms goods received and auto-creates Payment Voucher
 */
const grnConfig: DocumentApprovalConfig = {
  documentType: 'GOODS_RECEIVED_NOTE',
  configVersion: '1.0',
  effectiveDate: new Date('2024-01-01'),
  description: 'Simple GRN confirmation by Stores Officer',
  totalStages: 1,
  createdAt: new Date('2024-01-01'),
  createdBy: 'system',

  approvalStages: [
    {
      stageNumber: 1,
      stageName: 'Goods Receipt Confirmation',
      description: 'Stores officer confirms goods received and creates Payment Voucher',
      requiredRole: 'ADMIN',
      alternativeRoles: ['COMPLIANCE_OFFICER'],
      canReverse: false,
      reversalBehavior: 'BACK_TO_CREATOR',
      requiresComments: false,
      onApprovalActions: {
        createPaymentVoucher: true,
        createAuditLog: true,
      },
    },
  ],

  allowConcurrentApprovals: false,
  allowMultipleReversals: false,
  requireFinalSignoff: false,
};

/**
 * Payment Voucher: 4-stage approval with reversals to Accountant
 * Department Head → Auditor → Director Finance → Principal Officer
 * (Note: Accountant generates PV from GRN first, which is a precursor action)
 */
const paymentVoucherConfig: DocumentApprovalConfig = {
  documentType: 'PAYMENT_VOUCHER',
  configVersion: '1.0',
  effectiveDate: new Date('2024-01-01'),
  description: '4-stage PV approval with reversals to Accountant',
  totalStages: 4,
  createdAt: new Date('2024-01-01'),
  createdBy: 'system',

  approvalStages: [
    {
      stageNumber: 1,
      stageName: 'Department Head Approval',
      description: 'Department head reviews payment voucher',
      requiredRole: 'DEPARTMENT_MANAGER',
      canReverse: true,
      reversalBehavior: 'TO_SPECIFIC_USER',
      reversalTargetRole: 'FINANCE_OFFICER',
      reversalResetsPreviousStages: false,
      requiresComments: false,
    },
    {
      stageNumber: 2,
      stageName: 'Auditor Review',
      description: 'Auditor reviews for compliance and accuracy',
      requiredRole: 'COMPLIANCE_OFFICER',
      canReverse: true,
      reversalBehavior: 'TO_SPECIFIC_USER',
      reversalTargetRole: 'FINANCE_OFFICER',
      reversalResetsPreviousStages: false,
      requiresComments: true,
      requiredValidations: ['voucherCalculationCorrect', 'documentsComplete'],
    },
    {
      stageNumber: 3,
      stageName: 'Director Finance Approval',
      description: 'Finance director approves payment',
      requiredRole: 'DIRECTOR',
      canReverse: true,
      reversalBehavior: 'TO_SPECIFIC_USER',
      reversalTargetRole: 'FINANCE_OFFICER',
      reversalResetsPreviousStages: false,
      requiresComments: false,
      requiredValidations: ['bankInfoValidated', 'fundAvailable'],
    },
    {
      stageNumber: 4,
      stageName: 'Principal Officer Final Approval',
      description: 'Executive final approval, generates QR code and payment reference',
      requiredRole: 'PRINCIPAL_OFFICER',
      canReverse: true,
      reversalBehavior: 'TO_SPECIFIC_USER',
      reversalTargetRole: 'FINANCE_OFFICER',
      reversalResetsPreviousStages: false,
      requiresComments: false,
      onApprovalActions: {
        generateQRCode: true,
        generatePaymentReference: true,
        createAuditLog: true,
      },
    },
  ],

  fallbackStages: [
    {
      stageNumber: 1,
      stageName: 'Default Finance Review',
      requiredRole: 'DIRECTOR',
      canReverse: true,
      reversalBehavior: 'BACK_TO_CREATOR',
    },
  ],

  allowConcurrentApprovals: false,
  allowMultipleReversals: true,
  requireFinalSignoff: true,
};

// ============================================================================
// CONFIGURATION REGISTRY
// ============================================================================

/**
 * Registry of all approval configurations
 * Maps document type to its approval workflow configuration
 */
const approvalConfigs = new Map<string, DocumentApprovalConfig>([
  ['REQUISITION', requisitionConfig],
  ['PURCHASE_ORDER', purchaseOrderConfig],
  ['GOODS_RECEIVED_NOTE', grnConfig],
  ['PAYMENT_VOUCHER', paymentVoucherConfig],
]);

// ============================================================================
// CONFIGURATION RETRIEVAL & MANAGEMENT
// ============================================================================

/**
 * Get approval configuration for a document type
 * Returns configured version or fallback if not found
 */
export function getApprovalConfig(documentType: string): DocumentApprovalConfig {
  const config = approvalConfigs.get(documentType);

  if (config) {
    return config;
  }

  // If no config found, create minimal fallback
  return createFallbackConfig(documentType);
}

/**
 * Create fallback configuration when no config exists
 * Provides sensible defaults for any document type
 */
function createFallbackConfig(documentType: string): DocumentApprovalConfig {
  return {
    documentType: documentType as WorkflowDocumentType,
    configVersion: 'fallback-1.0',
    effectiveDate: new Date(),
    description: `Fallback configuration for ${documentType}`,
    totalStages: 2,
    approvalStages: [
      {
        stageNumber: 1,
        stageName: 'Manager Review',
        requiredRole: 'DEPARTMENT_MANAGER',
        canReverse: true,
        reversalBehavior: 'BACK_TO_CREATOR',
      },
      {
        stageNumber: 2,
        stageName: 'Director Approval',
        requiredRole: 'DIRECTOR',
        canReverse: true,
        reversalBehavior: 'BACK_TO_CREATOR',
      },
    ],
    createdAt: new Date(),
    createdBy: 'system',
  };
}

/**
 * Update a configuration
 * Useful for admin interfaces to modify approval workflows
 */
export function updateApprovalConfig(config: DocumentApprovalConfig): void {
  approvalConfigs.set(config.documentType, config);
}

/**
 * Get all registered configurations
 */
export function getAllApprovalConfigs(): DocumentApprovalConfig[] {
  return Array.from(approvalConfigs.values());
}

// ============================================================================
// APPROVAL STATE UTILITIES
// ============================================================================

/**
 * Get the current approval stage for a document
 */
export function getCurrentApprovalStage(
  state: ApprovalState
): ApprovalStageConfig | null {
  const config = getApprovalConfig(state.documentType);
  return (
    config.approvalStages.find((s) => s.stageNumber === state.currentStageNumber) ||
    null
  );
}

/**
 * Get the next approval stage after current one
 */
export function getNextApprovalStage(
  state: ApprovalState
): ApprovalStageConfig | null {
  const config = getApprovalConfig(state.documentType);
  const nextStageNumber = state.currentStageNumber + 1;

  return (
    config.approvalStages.find((s) => s.stageNumber === nextStageNumber) || null
  );
}

/**
 * Get a specific approval stage by number
 */
export function getApprovalStageByNumber(
  documentType: string,
  stageNumber: number
): ApprovalStageConfig | null {
  const config = getApprovalConfig(documentType);
  return config.approvalStages.find((s) => s.stageNumber === stageNumber) || null;
}

/**
 * Check if current stage is the final approval stage
 */
export function isFinalApprovalStage(state: ApprovalState): boolean {
  const config = getApprovalConfig(state.documentType);
  return state.currentStageNumber === config.totalStages;
}

/**
 * Get total number of approval stages for a document type
 */
export function getTotalApprovalStages(documentType: string): number {
  const config = getApprovalConfig(documentType);
  return config.totalStages;
}

/**
 * Check if reversal is allowed at current stage
 */
export function canReverseAtStage(state: ApprovalState): boolean {
  const stage = getCurrentApprovalStage(state);
  return stage?.canReverse ?? false;
}

/**
 * Get reversal target stage for a document
 */
export function getReversalTargetStage(
  state: ApprovalState
): ApprovalStageConfig | null {
  const stage = getCurrentApprovalStage(state);
  if (!stage) return null;

  const config = getApprovalConfig(state.documentType);

  if (stage.reversalBehavior === 'BACK_TO_CREATOR') {
    return config.approvalStages.find((s) => s.stageNumber === 1) || null;
  } else if (
    stage.reversalBehavior === 'BACK_TO_HANDLER' ||
    stage.reversalBehavior === 'TO_SPECIFIC_USER'
  ) {
    // Handler gets it, return current stage
    return stage;
  } else if (stage.reversalBehavior === 'PREVIOUS_STAGE') {
    return config.approvalStages.find((s) => s.stageNumber === stage.stageNumber - 1) ||
      null;
  }

  return null;
}

/**
 * Check if user has required role for approving at current stage
 */
export function userHasApprovalRole(
  state: ApprovalState,
  userRoles: string[]
): boolean {
  const stage = getCurrentApprovalStage(state);
  if (!stage) return false;

  // Check primary role
  if (userRoles.includes(stage.requiredRole)) {
    return true;
  }

  // Check alternative roles
  if (stage.alternativeRoles) {
    return stage.alternativeRoles.some((role) => userRoles.includes(role));
  }

  return false;
}

/**
 * Get all users who can approve at current stage
 * (Used for notification/assignment purposes)
 */
export function getRequiredApprovalRoles(state: ApprovalState): string[] {
  const stage = getCurrentApprovalStage(state);
  if (!stage) return [];

  const roles = [stage.requiredRole];
  if (stage.alternativeRoles) {
    roles.push(...stage.alternativeRoles);
  }

  return roles;
}

/**
 * Calculate progress percentage for a document in approval
 */
export function getApprovalProgress(state: ApprovalState): number {
  if (state.totalStages === 0) return 0;
  return Math.round((state.currentStageNumber / state.totalStages) * 100);
}

/**
 * Get approval stage summary for UI display
 */
export function getApprovalStageSummary(
  state: ApprovalState
): {
  currentStage: number;
  totalStages: number;
  stageName: string;
  requiredRole: string;
  progress: number;
} {
  const stage = getCurrentApprovalStage(state);
  return {
    currentStage: state.currentStageNumber,
    totalStages: state.totalStages,
    stageName: stage?.stageName ?? 'Unknown Stage',
    requiredRole: stage?.requiredRole ?? 'Unknown Role',
    progress: getApprovalProgress(state),
  };
}

// ============================================================================
// VALIDATION UTILITIES
// ============================================================================

/**
 * Check if all required validations pass for a stage
 */
export function validateStageRequirements(
  state: ApprovalState,
  validations: Record<string, boolean>
): { valid: boolean; failedValidations: string[] } {
  const stage = getCurrentApprovalStage(state);
  if (!stage || !stage.requiredValidations) {
    return { valid: true, failedValidations: [] };
  }

  const failed = stage.requiredValidations.filter((v) => !validations[v]);
  return {
    valid: failed.length === 0,
    failedValidations: failed,
  };
}

/**
 * Check if comments are required for approval at current stage
 */
export function commentsRequiredForApproval(state: ApprovalState): boolean {
  const stage = getCurrentApprovalStage(state);
  return stage?.requiresComments ?? false;
}

// ============================================================================
// CONFIGURATION EXPORT
// ============================================================================

export {
  requisitionConfig,
  purchaseOrderConfig,
  grnConfig,
  paymentVoucherConfig,
};
