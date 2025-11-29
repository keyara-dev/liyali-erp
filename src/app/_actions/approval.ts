'use server';

import {
  ApproveDocumentRequest,
  ApproveDocumentResponse,
  ReverseDocumentRequest,
  ReverseDocumentResponse,
  ApprovalState,
  ApprovalRecord,
} from '@/types/workflow';
import {
  getApprovalConfig,
  getCurrentApprovalStage,
  getNextApprovalStage,
  isFinalApprovalStage,
  userHasApprovalRole,
  getReversalTargetStage,
  validateStageRequirements,
  commentsRequiredForApproval,
} from '@/lib/approval-config';
import { store } from '@/lib/mock-data';

/**
 * Generic Approval Handler
 *
 * This server action handles approval for any document type.
 * It uses the configuration system to determine:
 * - Who can approve
 * - What validations are required
 * - What happens next
 * - What special actions to take (QR code, audit log, etc.)
 */
export async function approveDocument(
  request: ApproveDocumentRequest
): Promise<ApproveDocumentResponse> {
  try {
    // 1. Load document from store
    const document = store.documents.get(request.documentId);
    if (!document) {
      return {
        success: false,
        message: 'Document not found',
        error: 'NOT_FOUND',
      };
    }

    // 2. Load approval state
    let state = store.approvalStates.get(request.documentId);
    if (!state) {
      // Create initial state if doesn't exist
      const config = getApprovalConfig(request.documentType);
      state = {
        documentId: request.documentId,
        documentType: request.documentType,
        configVersion: config.configVersion,
        currentStageNumber: 1,
        totalStages: config.totalStages,
        stageHistory: [],
        status: 'DRAFT',
        lastModifiedAt: new Date(),
        lastModifiedBy: request.approvingUserId,
      };
    }

    // 3. Get configuration
    const config = getApprovalConfig(request.documentType);
    const currentStage = getCurrentApprovalStage(state);

    if (!currentStage) {
      return {
        success: false,
        message: 'Invalid approval stage',
        error: 'BAD_STAGE',
      };
    }

    // 4. Verify user has required role
    const user = store.users.get(request.approvingUserId);
    if (!user) {
      return {
        success: false,
        message: 'User not found',
        error: 'USER_NOT_FOUND',
      };
    }

    const userRoles = user.roleIds || [];
    if (!userHasApprovalRole(state, userRoles)) {
      return {
        success: false,
        message: `User does not have required role: ${currentStage.requiredRole}`,
        error: 'UNAUTHORIZED',
      };
    }

    // 5. Verify comments if required
    if (commentsRequiredForApproval(state) && !request.comments) {
      return {
        success: false,
        message: 'Comments are required for approval at this stage',
        error: 'COMMENTS_REQUIRED',
      };
    }

    // 6. Run validations if required
    if (currentStage.requiredValidations && request.validations) {
      const validationResult = validateStageRequirements(state, request.validations);
      if (!validationResult.valid) {
        return {
          success: false,
          message: `Validations failed: ${validationResult.failedValidations.join(', ')}`,
          error: 'VALIDATION_FAILED',
        };
      }
    }

    // 7. Record this approval in stage history
    const approvalRecord: ApprovalRecord = {
      stageNumber: currentStage.stageNumber,
      stageName: currentStage.stageName,
      assignedRole: currentStage.requiredRole,
      assignedTo: request.approvingUserId,
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

    // 8. Determine next state
    const isFinal = isFinalApprovalStage(state);

    if (isFinal) {
      state.status = 'APPROVED';
      state.approvedAt = new Date();
      state.currentStageNumber = config.totalStages;
    } else {
      const nextStage = getNextApprovalStage(state);
      if (nextStage) {
        state.currentStageNumber = nextStage.stageNumber;
      }
    }

    state.lastModifiedAt = new Date();
    state.lastModifiedBy = request.approvingUserId;
    state.status = isFinal ? 'APPROVED' : 'IN_APPROVAL';

    // 9. Execute stage-specific actions
    let qrCode: string | undefined;
    let paymentReference: string | undefined;

    if (currentStage.onApprovalActions?.generateQRCode) {
      qrCode = await generateQRCode(document);
      (document as any).qrCode = qrCode;
    }

    if (currentStage.onApprovalActions?.generatePaymentReference) {
      paymentReference = generatePaymentReference();
      (document as any).paymentReference = paymentReference;
    }

    if (currentStage.onApprovalActions?.createPaymentVoucher) {
      // For GRN: auto-create Payment Voucher
      await autoCreatePaymentVoucher(document.id);
    }

    if (currentStage.onApprovalActions?.createAuditLog) {
      const auditLogId = `audit-${Date.now()}`;
      store.auditLogs.set(auditLogId, {
        id: auditLogId,
        documentId: document.id,
        action: isFinal ? 'FINAL_APPROVAL' : 'STAGE_APPROVAL',
        userId: request.approvingUserId,
        timestamp: new Date(),
        details: `Document approved at stage ${currentStage.stageNumber}: ${currentStage.stageName}`,
        metadata: {
          stageName: currentStage.stageName,
          stageNumber: currentStage.stageNumber,
          approverRole: currentStage.requiredRole,
        },
      });
    }

    // 10. Update stores
    store.documents.set(request.documentId, document);
    store.approvalStates.set(request.documentId, state);

    // 11. Send notifications (future: email, in-app notifications)
    if (isFinal) {
      // Document is fully approved - could notify creator, stakeholders
      console.log(`Document ${request.documentId} fully approved by ${user.name}`);
    } else {
      // Document moving to next stage - could notify next approver
      const nextStage = getNextApprovalStage(state);
      if (nextStage) {
        console.log(
          `Document ${request.documentId} moved to stage ${nextStage.stageNumber}: ${nextStage.stageName}`
        );
      }
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
    console.error('Approval failed:', error);
    return {
      success: false,
      message: 'Approval failed',
      error: error instanceof Error ? error.message : 'UNKNOWN_ERROR',
    };
  }
}

/**
 * Generic Reversal Handler
 *
 * Handles reversal of approvals for any document type.
 * Determines where the document goes based on configuration.
 */
export async function reverseDocument(
  request: ReverseDocumentRequest
): Promise<ReverseDocumentResponse> {
  try {
    // 1. Load document and state
    const document = store.documents.get(request.documentId);
    const state = store.approvalStates.get(request.documentId);

    if (!document || !state) {
      return {
        success: false,
        message: 'Document not found',
        error: 'NOT_FOUND',
      };
    }

    // 2. Get configuration
    const config = getApprovalConfig(request.documentType);
    const currentStage = getCurrentApprovalStage(state);

    if (!currentStage) {
      return {
        success: false,
        message: 'Invalid approval stage',
        error: 'BAD_STAGE',
      };
    }

    // 3. Check if reversal is allowed at this stage
    if (!currentStage.canReverse) {
      return {
        success: false,
        message: `Reversals not allowed at stage: ${currentStage.stageName}`,
        error: 'REVERSAL_NOT_ALLOWED',
      };
    }

    // 4. Verify user has permission to reverse
    const user = store.users.get(request.reversingUserId);
    if (!user) {
      return {
        success: false,
        message: 'User not found',
        error: 'USER_NOT_FOUND',
      };
    }

    // Check if user is the one who approved at this stage
    const lastApproval = state.stageHistory
      .filter((r) => r.stageNumber === currentStage.stageNumber && r.status === 'APPROVED')
      .pop();

    if (!lastApproval || lastApproval.actionTakenBy !== request.reversingUserId) {
      // In production, you might want to allow managers/admins to reverse too
      // For now, only the approver can reverse their own approval
      console.warn(
        `User ${request.reversingUserId} attempted to reverse but is not the approver`
      );
      // Continue anyway for now - can add stricter checks later
    }

    // 5. Record the reversal
    const reversalRecord: ApprovalRecord = {
      stageNumber: currentStage.stageNumber,
      stageName: currentStage.stageName,
      assignedRole: currentStage.requiredRole,
      assignedTo: lastApproval?.assignedTo || request.reversingUserId,
      status: 'REVERSED',
      reversedAt: new Date(),
      reversalReason: request.reversalReason,
    };

    state.stageHistory.push(reversalRecord);

    // 6. Determine where reversal goes based on configuration
    let targetStageNumber = 1; // Default: back to first stage
    let targetRole = currentStage.requiredRole;

    if (currentStage.reversalBehavior === 'BACK_TO_CREATOR') {
      targetStageNumber = 1;
      const firstStage = config.approvalStages.find((s) => s.stageNumber === 1);
      if (firstStage) {
        targetRole = firstStage.requiredRole;
      }
    } else if (
      currentStage.reversalBehavior === 'BACK_TO_HANDLER' ||
      currentStage.reversalBehavior === 'TO_SPECIFIC_USER'
    ) {
      if (currentStage.reversalTargetRole) {
        targetRole = currentStage.reversalTargetRole;
      }
      targetStageNumber = state.currentStageNumber;
    } else if (currentStage.reversalBehavior === 'PREVIOUS_STAGE') {
      targetStageNumber = Math.max(1, currentStage.stageNumber - 1);
      const previousStage = config.approvalStages.find(
        (s) => s.stageNumber === targetStageNumber
      );
      if (previousStage) {
        targetRole = previousStage.requiredRole;
      }
    }

    // 7. Reset approval state if configured
    if (currentStage.reversalResetsPreviousStages) {
      state.stageHistory = state.stageHistory.filter(
        (r) => r.stageNumber < currentStage.stageNumber
      );
    }

    state.status = 'REVERSED';
    state.currentStageNumber = targetStageNumber;
    state.lastModifiedAt = new Date();
    state.lastModifiedBy = request.reversingUserId;

    // 8. Create reversal audit log
    const auditLogId = `audit-reversal-${Date.now()}`;
    store.auditLogs.set(auditLogId, {
      id: auditLogId,
      documentId: request.documentId,
      action: 'REVERSAL',
      userId: request.reversingUserId,
      timestamp: new Date(),
      details: `Document reversed from stage ${currentStage.stageNumber} by ${user.name}. Reason: ${request.reversalReason}`,
      metadata: {
        previousStage: currentStage.stageNumber,
        targetStage: targetStageNumber,
        reversalReason: request.reversalReason,
        targetRole,
      },
    });

    // 9. Update store
    store.approvalStates.set(request.documentId, state);

    // 10. Send notification to handler
    console.log(
      `Document ${request.documentId} reversed to ${targetRole} for correction. Reason: ${request.reversalReason}`
    );

    return {
      success: true,
      message: `Document reversed. Returned to ${targetRole} for correction.`,
      reversedToStage: targetStageNumber,
      reversedToRole: targetRole,
    };
  } catch (error) {
    console.error('Reversal failed:', error);
    return {
      success: false,
      message: 'Reversal failed',
      error: error instanceof Error ? error.message : 'UNKNOWN_ERROR',
    };
  }
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

/**
 * Generate QR code for payment voucher
 * In production, this would use a QR code library
 */
async function generateQRCode(document: any): Promise<string> {
  const qrData = {
    documentId: document.id,
    type: document.type,
    amount: document.metadata?.amount || document.metadata?.netAmount,
    date: new Date().toISOString(),
    reference: document.paymentReference || 'REF-' + Date.now(),
  };

  // In production, use qrcode library
  // For now, return a mock QR code string
  return `QR_${btoa(JSON.stringify(qrData)).substring(0, 50)}`;
}

/**
 * Generate unique payment reference number
 */
function generatePaymentReference(): string {
  const year = new Date().getFullYear();
  const month = String(new Date().getMonth() + 1).padStart(2, '0');
  const random = Math.random().toString(36).substring(2, 8).toUpperCase();
  return `PV-${year}${month}-${random}`;
}

/**
 * Auto-create Payment Voucher from GRN
 * Called when GRN is confirmed
 */
async function autoCreatePaymentVoucher(grnId: string): Promise<string> {
  const grn = store.documents.get(grnId);
  if (!grn) {
    throw new Error(`GRN ${grnId} not found`);
  }

  // Create new payment voucher document
  const pvId = `pv-${Date.now()}`;
  const paymentVoucher = {
    id: pvId,
    type: 'PAYMENT_VOUCHER',
    documentNumber: `PV-${Date.now()}`,
    status: 'DRAFT',
    currentStage: 1,
    createdBy: grn.createdBy,
    createdAt: new Date(),
    updatedAt: new Date(),
    metadata: {
      grnId: grnId,
      poId: grn.metadata?.poId,
      vendorName: grn.metadata?.vendorName,
      amount: grn.metadata?.totalAmount,
      ...grn.metadata,
    },
  };

  // Create initial approval state for Payment Voucher
  const pvState: ApprovalState = {
    documentId: pvId,
    documentType: 'PAYMENT_VOUCHER',
    configVersion: '1.0',
    currentStageNumber: 1,
    totalStages: 4,
    stageHistory: [],
    status: 'SUBMITTED',
    submittedAt: new Date(),
    lastModifiedAt: new Date(),
    lastModifiedBy: grn.createdBy,
  };

  store.documents.set(pvId, paymentVoucher);
  store.approvalStates.set(pvId, pvState);

  return pvId;
}

/**
 * Submit document for approval
 * Moves document from DRAFT to SUBMITTED/IN_APPROVAL
 */
export async function submitDocumentForApproval(
  documentId: string,
  documentType: string,
  submittingUserId: string
): Promise<{ success: boolean; message: string; error?: string }> {
  try {
    const document = store.documents.get(documentId);
    if (!document) {
      return { success: false, message: 'Document not found', error: 'NOT_FOUND' };
    }

    let state = store.approvalStates.get(documentId);
    if (!state) {
      const config = getApprovalConfig(documentType);
      state = {
        documentId,
        documentType,
        configVersion: config.configVersion,
        currentStageNumber: 1,
        totalStages: config.totalStages,
        stageHistory: [],
        status: 'SUBMITTED',
        submittedAt: new Date(),
        lastModifiedAt: new Date(),
        lastModifiedBy: submittingUserId,
      };
    } else {
      state.status = 'SUBMITTED';
      state.submittedAt = new Date();
      state.currentStageNumber = 1;
      state.lastModifiedAt = new Date();
      state.lastModifiedBy = submittingUserId;
    }

    document.status = 'IN_APPROVAL';
    document.updatedAt = new Date();

    store.documents.set(documentId, document);
    store.approvalStates.set(documentId, state);

    return {
      success: true,
      message: 'Document submitted for approval',
    };
  } catch (error) {
    return {
      success: false,
      message: 'Submission failed',
      error: error instanceof Error ? error.message : 'UNKNOWN_ERROR',
    };
  }
}

/**
 * Get approval state for a document
 */
export async function getApprovalState(
  documentId: string
): Promise<ApprovalState | null> {
  return store.approvalStates.get(documentId) || null;
}
