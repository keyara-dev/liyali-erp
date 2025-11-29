# Dynamic Approval Configuration System

**Date**: 2024-11-29
**Status**: Design Specification
**Purpose**: Enable flexible, configurable approval workflows across all document types
**Key Requirement**: "Dynamic enough to accommodate the flows shown... with fallbacks"

---

## 1. Overview

The system must support variable approval workflows because different documents require different approval chains:

- **Requisition**: 4-stage approval (Head → Principal → Director → Procurement)
- **Purchase Order**: 4-stage approval (Dept Head → Auditor → Director Finance → Principal Officer)
- **Goods Received Note**: Simple 1-stage (Stores Officer confirmation)
- **Payment Voucher**: 5 steps (Accountant generation → 4-stage approval with reversals)

This design makes approval workflows **configuration-driven** rather than hardcoded, with sensible **fallback defaults**.

---

## 2. Core Concepts

### 2.1 Approval Stage

A single step in an approval workflow where a specific role reviews and approves (or reverses) a document.

```typescript
interface ApprovalStageConfig {
  // Identification
  stageNumber: number
  stageName: string
  description?: string

  // Who approves
  requiredRole: string
  alternativeRoles?: string[] // Fallback roles if primary unavailable

  // Reversal configuration
  canReverse: boolean
  reversalBehavior: 'BACK_TO_CREATOR' | 'BACK_TO_HANDLER' | 'PREVIOUS_STAGE' | 'TO_SPECIFIC_USER'
  reversalTargetRole?: string // If TO_SPECIFIC_USER, which role gets it
  reversalResetsPreviousStages?: boolean // If true, resets all approved stages

  // Validation
  requiresComments?: boolean
  requiredValidations?: string[] // e.g., ["bankInfoValidation", "budgetCheck"]

  // Special actions on approval
  onApprovalActions?: {
    generateQRCode?: boolean
    generatePaymentReference?: boolean
    createAuditLog?: boolean
    notifyVendor?: boolean
    createPaymentVoucher?: boolean // For GRN
  }

  // Timeline
  slaHours?: number
  escalationRoleAfterSLA?: string
}
```

### 2.2 Document Type Approval Workflow

Complete approval configuration for a document type.

```typescript
interface DocumentApprovalConfig {
  // Document Type
  documentType: 'REQUISITION' | 'PURCHASE_ORDER' | 'GOODS_RECEIVED_NOTE' | 'PAYMENT_VOUCHER'

  // Version
  configVersion: string
  effectiveDate: Date

  // Stages
  approvalStages: ApprovalStageConfig[]
  totalStages: number

  // Fallback configuration
  fallbackStages?: ApprovalStageConfig[] // Used if stages not configured

  // General rules
  allowConcurrentApprovals?: boolean
  allowMultipleReversals?: boolean
  requireFinalSignoff?: boolean

  // Metadata
  description: string
  createdAt: Date
  createdBy: string
}
```

### 2.3 Approval State

Current state of a document within its approval workflow.

```typescript
interface ApprovalState {
  documentId: string
  documentType: string
  configVersion: string

  // Current position in workflow
  currentStageNumber: number
  totalStages: number

  // Stage history
  stageHistory: ApprovalRecord[]

  // Overall status
  status: 'DRAFT' | 'SUBMITTED' | 'IN_APPROVAL' | 'APPROVED' | 'REJECTED' | 'REVERSED'
  submittedAt?: Date
  approvedAt?: Date
  rejectedAt?: Date

  // Metadata
  lastModifiedAt: Date
  lastModifiedBy: string
}

interface ApprovalRecord {
  stageNumber: number
  stageName: string
  assignedTo: string // user ID
  assignedRole: string
  status: 'PENDING' | 'APPROVED' | 'REVERSED' | 'REJECTED'

  actionTakenAt?: Date
  actionTakenBy?: string // user ID
  comments?: string

  reversedAt?: Date
  reversalReason?: string

  validationsPassed?: string[]
  validationsFailed?: string[]
}
```

---

## 3. Configuration Examples

### 3.1 Requisition (4-Stage Approval)

```typescript
const requisitionConfig: DocumentApprovalConfig = {
  documentType: 'REQUISITION',
  configVersion: '1.0',
  effectiveDate: new Date('2024-01-01'),
  description: 'Standard 4-stage requisition approval workflow',
  totalStages: 4,

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
      alternativeRoles: ['CHIEF_EXECUTIVE'],
      canReverse: true,
      reversalBehavior: 'BACK_TO_CREATOR',
      requiresComments: true,
    },
    {
      stageNumber: 3,
      stageName: 'Finance Director Review',
      description: 'Finance director checks budget and accounting',
      requiredRole: 'DIRECTOR_FINANCE',
      alternativeRoles: ['FINANCE_MANAGER'],
      canReverse: true,
      reversalBehavior: 'BACK_TO_CREATOR',
      requiresComments: false,
      requiredValidations: ['budgetAvailable', 'accountCodeValid'],
    },
    {
      stageNumber: 4,
      stageName: 'Procurement Officer Processing',
      description: 'Procurement officer adds supplier and creates PO',
      requiredRole: 'PROCUREMENT_OFFICER',
      canReverse: false, // Final stage cannot reverse
      reversalBehavior: 'BACK_TO_CREATOR',
      requiresComments: false,
      onApprovalActions: {
        createPaymentVoucher: false,
        createAuditLog: true,
      },
    },
  ],

  // Fallback if not configured elsewhere
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
      requiredRole: 'DIRECTOR_FINANCE',
      canReverse: true,
      reversalBehavior: 'BACK_TO_CREATOR',
    },
  ],

  allowConcurrentApprovals: false,
  allowMultipleReversals: true,
  requireFinalSignoff: true,
  createdAt: new Date(),
  createdBy: 'system',
};
```

### 3.2 Purchase Order (4-Stage Approval with Reversals)

```typescript
const purchaseOrderConfig: DocumentApprovalConfig = {
  documentType: 'PURCHASE_ORDER',
  configVersion: '1.0',
  effectiveDate: new Date('2024-01-01'),
  description: '4-stage PO approval with reversals to Procurement Officer',
  totalStages: 4,

  approvalStages: [
    {
      stageNumber: 1,
      stageName: 'Department Head Approval',
      description: 'Department head approves procurement need',
      requiredRole: 'DEPARTMENT_MANAGER',
      canReverse: true,
      reversalBehavior: 'TO_SPECIFIC_USER',
      reversalTargetRole: 'PROCUREMENT_OFFICER',
      reversalResetsPreviousStages: false,
      requiresComments: false,
    },
    {
      stageNumber: 2,
      stageName: 'Auditor Review',
      description: 'Internal auditor reviews for compliance',
      requiredRole: 'AUDITOR',
      canReverse: true,
      reversalBehavior: 'TO_SPECIFIC_USER',
      reversalTargetRole: 'PROCUREMENT_OFFICER',
      reversalResetsPreviousStages: false,
      requiresComments: true,
      requiredValidations: ['complianceCheck', 'vendorApproval'],
    },
    {
      stageNumber: 3,
      stageName: 'Director Finance Approval',
      description: 'Finance director approves budget allocation',
      requiredRole: 'DIRECTOR_FINANCE',
      canReverse: true,
      reversalBehavior: 'TO_SPECIFIC_USER',
      reversalTargetRole: 'PROCUREMENT_OFFICER',
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
      reversalTargetRole: 'PROCUREMENT_OFFICER',
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
  createdAt: new Date(),
  createdBy: 'system',
};
```

### 3.3 Goods Received Note (1-Stage Simple)

```typescript
const grnConfig: DocumentApprovalConfig = {
  documentType: 'GOODS_RECEIVED_NOTE',
  configVersion: '1.0',
  effectiveDate: new Date('2024-01-01'),
  description: 'Simple GRN confirmation by Stores Officer',
  totalStages: 1,

  approvalStages: [
    {
      stageNumber: 1,
      stageName: 'Goods Receipt Confirmation',
      description: 'Stores officer confirms goods received and creates Payment Voucher',
      requiredRole: 'STORES_OFFICER',
      alternativeRoles: ['PROCUREMENT_OFFICER'],
      canReverse: false, // GRN is just receipt confirmation
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
  createdAt: new Date(),
  createdBy: 'system',
};
```

### 3.4 Payment Voucher (4-Stage with Accountant Step)

```typescript
const paymentVoucherConfig: DocumentApprovalConfig = {
  documentType: 'PAYMENT_VOUCHER',
  configVersion: '1.0',
  effectiveDate: new Date('2024-01-01'),
  description: '4-stage PV approval with reversals to Accountant',
  totalStages: 4,

  // Note: Step 0 is "Accountant Generates PV from GRN" - not an approval stage
  // This is a precursor action, not part of approval workflow
  // Approval starts at Stage 1

  approvalStages: [
    {
      stageNumber: 1,
      stageName: 'Department Head Approval',
      description: 'Department head reviews payment voucher',
      requiredRole: 'DEPARTMENT_MANAGER',
      canReverse: true,
      reversalBehavior: 'TO_SPECIFIC_USER',
      reversalTargetRole: 'ACCOUNTANT',
      reversalResetsPreviousStages: false,
      requiresComments: false,
    },
    {
      stageNumber: 2,
      stageName: 'Auditor Review',
      description: 'Auditor reviews for compliance and accuracy',
      requiredRole: 'AUDITOR',
      canReverse: true,
      reversalBehavior: 'TO_SPECIFIC_USER',
      reversalTargetRole: 'ACCOUNTANT',
      reversalResetsPreviousStages: false,
      requiresComments: true,
      requiredValidations: ['voucherCalculationCorrect', 'documentsComplete'],
    },
    {
      stageNumber: 3,
      stageName: 'Director Finance Approval',
      description: 'Finance director approves payment',
      requiredRole: 'DIRECTOR_FINANCE',
      canReverse: true,
      reversalBehavior: 'TO_SPECIFIC_USER',
      reversalTargetRole: 'ACCOUNTANT',
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
      reversalTargetRole: 'ACCOUNTANT',
      reversalResetsPreviousStages: false,
      requiresComments: false,
      onApprovalActions: {
        generateQRCode: true,
        generatePaymentReference: true,
        createAuditLog: true,
      },
    },
  ],

  // Fallback for if not configured
  fallbackStages: [
    {
      stageNumber: 1,
      stageName: 'Default Finance Review',
      requiredRole: 'DIRECTOR_FINANCE',
      canReverse: true,
      reversalBehavior: 'BACK_TO_CREATOR',
    },
  ],

  allowConcurrentApprovals: false,
  allowMultipleReversals: true,
  requireFinalSignoff: true,
  createdAt: new Date(),
  createdBy: 'system',
};
```

---

## 4. Configuration Management

### 4.1 Configuration Store

Configuration is stored in-memory with loading from defaults:

```typescript
// Configuration registry
const approvalConfigs = new Map<string, DocumentApprovalConfig>([
  ['REQUISITION', requisitionConfig],
  ['PURCHASE_ORDER', purchaseOrderConfig],
  ['GOODS_RECEIVED_NOTE', grnConfig],
  ['PAYMENT_VOUCHER', paymentVoucherConfig],
]);

// Get configuration with fallback
function getApprovalConfig(
  documentType: string,
  version?: string
): DocumentApprovalConfig {
  const config = approvalConfigs.get(documentType);

  // Return configured version or fallback
  if (config) {
    return config;
  }

  // If no config found, create minimal fallback
  return createFallbackConfig(documentType);
}

// Create fallback configuration
function createFallbackConfig(documentType: string): DocumentApprovalConfig {
  return {
    documentType,
    configVersion: 'fallback-1.0',
    effectiveDate: new Date(),
    description: `Fallback configuration for ${documentType}`,
    totalStages: 2,
    approvalStages: [
      {
        stageNumber: 1,
        stageName: 'Manager Review',
        requiredRole: 'MANAGER',
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
```

### 4.2 Get Current Stage

```typescript
function getCurrentApprovalStage(
  state: ApprovalState
): ApprovalStageConfig | null {
  const config = getApprovalConfig(state.documentType);
  return config.approvalStages.find(
    (s) => s.stageNumber === state.currentStageNumber
  ) || null;
}
```

### 4.3 Get Next Stage

```typescript
function getNextApprovalStage(
  state: ApprovalState
): ApprovalStageConfig | null {
  const config = getApprovalConfig(state.documentType);
  const nextStageNumber = state.currentStageNumber + 1;

  return config.approvalStages.find((s) => s.stageNumber === nextStageNumber) || null;
}
```

### 4.4 Is Final Stage

```typescript
function isFinalApprovalStage(state: ApprovalState): boolean {
  const config = getApprovalConfig(state.documentType);
  return state.currentStageNumber === config.totalStages;
}
```

---

## 5. Generic Approval Handler

Single server action that handles approval for any document type using configuration:

```typescript
interface ApproveDocumentRequest {
  documentId: string
  documentType: string
  approvingUserId: string
  comments?: string
  validations?: Record<string, boolean>
}

interface ApproveDocumentResponse {
  success: boolean
  message: string
  newStageNumber?: number
  isFinalApproval?: boolean
  generatedQRCode?: string
  generatedPaymentReference?: string
  error?: string
}

async function approveDocument(
  request: ApproveDocumentRequest
): Promise<ApproveDocumentResponse> {
  try {
    // 1. Load document and current state
    const document = store.documents.get(request.documentId);
    if (!document) {
      return { success: false, message: 'Document not found', error: 'NOT_FOUND' };
    }

    const state = store.approvalStates.get(request.documentId);
    if (!state) {
      return { success: false, message: 'Approval state not found', error: 'NO_STATE' };
    }

    // 2. Get configuration
    const config = getApprovalConfig(request.documentType);
    const currentStage = getCurrentApprovalStage(state);

    if (!currentStage) {
      return { success: false, message: 'Invalid approval stage', error: 'BAD_STAGE' };
    }

    // 3. Verify user has required role
    const user = store.users.get(request.approvingUserId);
    if (!user || !user.roleIds.includes(currentStage.requiredRole)) {
      // Check alternative roles
      if (currentStage.alternativeRoles) {
        const hasAlternativeRole = currentStage.alternativeRoles.some((role) =>
          user?.roleIds.includes(role)
        );
        if (!hasAlternativeRole) {
          return {
            success: false,
            message: `User does not have required role: ${currentStage.requiredRole}`,
            error: 'UNAUTHORIZED',
          };
        }
      } else {
        return {
          success: false,
          message: `User does not have required role: ${currentStage.requiredRole}`,
          error: 'UNAUTHORIZED',
        };
      }
    }

    // 4. Run validations if required
    if (currentStage.requiredValidations) {
      for (const validation of currentStage.requiredValidations) {
        if (request.validations && !request.validations[validation]) {
          return {
            success: false,
            message: `Validation failed: ${validation}`,
            error: 'VALIDATION_FAILED',
          };
        }
      }
    }

    // 5. Record this approval
    const approvalRecord: ApprovalRecord = {
      stageNumber: currentStage.stageNumber,
      stageName: currentStage.stageName,
      assignedRole: currentStage.requiredRole,
      assignedTo: user.id,
      status: 'APPROVED',
      actionTakenAt: new Date(),
      actionTakenBy: request.approvingUserId,
      comments: request.comments,
      validationsPassed: request.validations
        ? Object.entries(request.validations)
            .filter(([, passed]) => passed)
            .map(([key]) => key)
        : [],
    };

    state.stageHistory.push(approvalRecord);

    // 6. Move to next stage or mark as approved
    const isFinal = isFinalApprovalStage(state);

    if (isFinal) {
      state.status = 'APPROVED';
      state.approvedAt = new Date();
      state.currentStageNumber = config.totalStages; // Stay at final stage
    } else {
      const nextStage = getNextApprovalStage(state);
      if (nextStage) {
        state.currentStageNumber = nextStage.stageNumber;
      }
    }

    state.lastModifiedAt = new Date();
    state.lastModifiedBy = request.approvingUserId;

    // 7. Execute stage-specific actions
    let qrCode: string | undefined;
    let paymentReference: string | undefined;

    if (currentStage.onApprovalActions?.generateQRCode) {
      qrCode = await generateQRCode(document);
      // Store on document
      (document as any).qrCode = qrCode;
    }

    if (currentStage.onApprovalActions?.generatePaymentReference) {
      paymentReference = generatePaymentReference();
      // Store on document
      (document as any).paymentReference = paymentReference;
    }

    if (currentStage.onApprovalActions?.createPaymentVoucher) {
      // For GRN: auto-create Payment Voucher
      await autoCreatePaymentVoucher(document.id);
    }

    if (currentStage.onApprovalActions?.createAuditLog) {
      store.auditLogs.set(`${document.id}-approval-${Date.now()}`, {
        documentId: document.id,
        action: 'FINAL_APPROVAL',
        userId: request.approvingUserId,
        timestamp: new Date(),
        details: `Document approved at stage ${currentStage.stageNumber}`,
      });
    }

    // 8. Update state in store
    store.approvalStates.set(request.documentId, state);

    // 9. Send notifications
    if (isFinal) {
      // Document is fully approved - notify creator
      // (implementation details...)
    } else {
      // Document moving to next stage - notify next approver
      // (implementation details...)
    }

    return {
      success: true,
      message: isFinal
        ? 'Document fully approved'
        : `Moved to stage ${state.currentStageNumber}`,
      newStageNumber: state.currentStageNumber,
      isFinalApproval: isFinal,
      generatedQRCode: qrCode,
      generatedPaymentReference: paymentReference,
    };
  } catch (error) {
    return {
      success: false,
      message: 'Approval failed',
      error: error instanceof Error ? error.message : 'UNKNOWN_ERROR',
    };
  }
}
```

---

## 6. Generic Reversal Handler

Single server action that handles reversal for any document type:

```typescript
interface ReverseDocumentRequest {
  documentId: string
  documentType: string
  reversingUserId: string
  reversalReason: string
}

interface ReverseDocumentResponse {
  success: boolean
  message: string
  reversedToStage?: number
  reversedToRole?: string
  error?: string
}

async function reverseDocument(
  request: ReverseDocumentRequest
): Promise<ReverseDocumentResponse> {
  try {
    // 1. Load document and state
    const document = store.documents.get(request.documentId);
    const state = store.approvalStates.get(request.documentId);

    if (!document || !state) {
      return { success: false, message: 'Document not found', error: 'NOT_FOUND' };
    }

    // 2. Get configuration
    const config = getApprovalConfig(request.documentType);
    const currentStage = getCurrentApprovalStage(state);

    if (!currentStage) {
      return { success: false, message: 'Invalid approval stage', error: 'BAD_STAGE' };
    }

    // 3. Check if reversal is allowed at this stage
    if (!currentStage.canReverse) {
      return {
        success: false,
        message: `Reversals not allowed at stage: ${currentStage.stageName}`,
        error: 'REVERSAL_NOT_ALLOWED',
      };
    }

    // 4. Get the last approval at this stage
    const lastApproval = state.stageHistory
      .filter((r) => r.stageNumber === currentStage.stageNumber && r.status === 'APPROVED')
      .pop();

    if (!lastApproval) {
      return {
        success: false,
        message: 'No approval found to reverse',
        error: 'NO_APPROVAL_TO_REVERSE',
      };
    }

    // 5. Record the reversal
    const reversalRecord: ApprovalRecord = {
      stageNumber: currentStage.stageNumber,
      stageName: currentStage.stageName,
      assignedRole: currentStage.requiredRole,
      assignedTo: lastApproval.assignedTo,
      status: 'REVERSED',
      reversedAt: new Date(),
      reversalReason: request.reversalReason,
    };

    state.stageHistory.push(reversalRecord);

    // 6. Determine where reversal goes
    let targetStageNumber = 1; // Default: back to first stage
    let targetRole = currentStage.requiredRole;

    if (currentStage.reversalBehavior === 'BACK_TO_CREATOR') {
      targetStageNumber = 1; // Goes back to first stage
    } else if (
      currentStage.reversalBehavior === 'BACK_TO_HANDLER' ||
      currentStage.reversalBehavior === 'TO_SPECIFIC_USER'
    ) {
      if (currentStage.reversalTargetRole) {
        targetRole = currentStage.reversalTargetRole;
      }
      // Handler gets it, not back in approval chain
      targetStageNumber = state.currentStageNumber;
    } else if (currentStage.reversalBehavior === 'PREVIOUS_STAGE') {
      targetStageNumber = Math.max(1, currentStage.stageNumber - 1);
    }

    // 7. Reset approval state
    if (currentStage.reversalResetsPreviousStages) {
      // Clear all stages after this one
      state.stageHistory = state.stageHistory.filter(
        (r) => r.stageNumber < currentStage.stageNumber
      );
    }

    state.status = 'REVERSED';
    state.currentStageNumber = targetStageNumber;
    state.lastModifiedAt = new Date();
    state.lastModifiedBy = request.reversingUserId;

    // 8. Update store
    store.approvalStates.set(request.documentId, state);

    // 9. Send notification to handler
    // Handler needs to correct and resubmit
    // (implementation details...)

    return {
      success: true,
      message: `Document reversed. Returned to ${targetRole} for correction.`,
      reversedToStage: targetStageNumber,
      reversedToRole: targetRole,
    };
  } catch (error) {
    return {
      success: false,
      message: 'Reversal failed',
      error: error instanceof Error ? error.message : 'UNKNOWN_ERROR',
    };
  }
}
```

---

## 7. Integration with Document Types

### 7.1 Update Data Models

All document types need approval state tracking:

```typescript
interface WorkflowDocument {
  id: string
  type: 'REQUISITION' | 'PURCHASE_ORDER' | 'GOODS_RECEIVED_NOTE' | 'PAYMENT_VOUCHER'

  // Document-specific fields...

  // Approval tracking (from ApprovalState)
  status: 'DRAFT' | 'SUBMITTED' | 'IN_APPROVAL' | 'APPROVED' | 'REJECTED' | 'REVERSED'
  currentStageNumber: number
  totalStages: number

  // Timestamps
  createdAt: Date
  submittedAt?: Date
  approvedAt?: Date
  rejectedAt?: Date
}
```

### 7.2 Server Actions Updated

Replace stage-specific approval handlers with generic ones:

```typescript
// OLD (per-stage specific):
// async function approvePaymentVoucherStage1(...) { ... }
// async function approvePaymentVoucherStage2(...) { ... }
// async function approvePaymentVoucherStage3(...) { ... }

// NEW (generic, configuration-driven):
async function approveDocument(request: ApproveDocumentRequest) {
  return approveDocument(request);
}

async function reverseDocument(request: ReverseDocumentRequest) {
  return reverseDocument(request);
}

// Specific helpers still exist for document creation:
async function autoCreatePurchaseOrder(requisitionId: string) { ... }
async function autoCreatePaymentVoucher(grnId: string) { ... }
async function generateQRCode(document: PaymentVoucher) { ... }
async function generatePaymentReference() { ... }
```

---

## 8. Benefits of This Design

### 8.1 Flexibility
- Add new document types without changing approval logic
- Change approval stages without code changes
- Support different approval chains in different departments
- Easy to configure for different organizations

### 8.2 Maintainability
- Single approval handler (not N handlers per stage)
- Configuration stored in one place
- Easy to audit approval flow
- Clear role and responsibility definitions

### 8.3 Fallbacks
- Sensible defaults if configuration missing
- System continues to work even with bad config
- Graceful degradation

### 8.4 Extensibility
- Easy to add new validation types
- Easy to add new action types (notifications, integrations)
- Easy to add SLA tracking
- Easy to add escalation rules

---

## 9. Implementation Roadmap

### Phase 1: Core System (2 days)
- Create configuration interfaces (done above)
- Create configuration store with defaults
- Create generic approval handler
- Create generic reversal handler
- Update server actions

### Phase 2: Integration (3 days)
- Update Requisition workflow to use dynamic system
- Update Purchase Order workflow to use dynamic system
- Update Goods Received Note to use dynamic system
- Update Payment Voucher workflow to use dynamic system
- Test all document types

### Phase 3: Enhancement (2 days)
- Add SLA tracking
- Add escalation rules
- Add more validation types
- Create configuration UI (admin interface)
- Add audit reporting

### Phase 4: Advanced (Optional)
- Multi-department configurations
- Conditional stages (stages that only run if conditions met)
- Parallel approvals
- Weighted approvals (majority vote)

---

## 10. Migration Path from Simple to Dynamic

If you have existing approval implementations, migration is straightforward:

1. Define configuration for existing document type
2. Replace old approval handlers with calls to generic handler
3. Update UI to use configuration (stage names, role requirements)
4. Gradual rollout by document type

---

## 11. Configuration Schema Summary

```typescript
// The complete schema:
type ApprovalConfig = {
  documentType: DocumentType
  configVersion: string
  effectiveDate: Date
  description: string
  totalStages: number
  approvalStages: Array<{
    stageNumber: number
    stageName: string
    description?: string
    requiredRole: string
    alternativeRoles?: string[]
    canReverse: boolean
    reversalBehavior: 'BACK_TO_CREATOR' | 'BACK_TO_HANDLER' | 'PREVIOUS_STAGE' | 'TO_SPECIFIC_USER'
    reversalTargetRole?: string
    reversalResetsPreviousStages?: boolean
    requiresComments?: boolean
    requiredValidations?: string[]
    onApprovalActions?: {
      generateQRCode?: boolean
      generatePaymentReference?: boolean
      createAuditLog?: boolean
      notifyVendor?: boolean
      createPaymentVoucher?: boolean
    }
    slaHours?: number
    escalationRoleAfterSLA?: string
  }>
  fallbackStages?: ApprovalStageConfig[]
  allowConcurrentApprovals?: boolean
  allowMultipleReversals?: boolean
  requireFinalSignoff?: boolean
  createdAt: Date
  createdBy: string
}
```

---

## 12. Next Steps

1. **Create `src/lib/approval-config.ts`** - Configuration management and helpers
2. **Create `src/app/_actions/approval.ts`** - Generic approval and reversal handlers
3. **Update `src/types/workflow.ts`** - Add ApprovalState and ApprovalRecord types
4. **Update all document type handlers** - Use generic handlers
5. **Test with actual workflows** - Ensure all document types work correctly

---

**Status**: Design complete and ready for implementation
**Created**: 2024-11-29
**Next Phase**: Implement configuration system in code (src/lib/approval-config.ts and src/app/_actions/approval.ts)
