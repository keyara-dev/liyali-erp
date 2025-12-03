/**
 * Workflow Resolution & Orchestration
 * Determines workflow for entities, assigns approvers, manages reassignments
 *
 * KEY FEATURES:
 * - Resolve which workflow applies to an entity
 * - Assign approvers based on role or specific user
 * - Support reassignment by assigned user or admin
 * - Track reassignment audit trail
 * - Progress workflow through stages based on actions
 */

import type {
  CustomWorkflow,
  WorkflowAssignment,
  WorkflowEntityType,
  WorkflowStage,
  StageExecution,
  StageAssignment,
  ReassignStageRequest,
  ApproveStageRequest,
  RejectStageRequest,
} from "@/types";
import {
  getWorkflow,
  getWorkflowDefault,
  getAssignment,
  updateAssignment,
  getPendingApprovalsForUser,
} from "./workflow-persistence";
import { getCurrentUser } from "@/lib/auth";
import type { UserRole, User } from "@/types";
import { DEMO_USERS } from "@/lib/auth";

// ============================================================================
// WORKFLOW RESOLUTION
// ============================================================================

/**
 * Resolve which workflow applies to an entity
 *
 * Priority:
 * 1. Explicit workflow assigned to entity
 * 2. Default workflow for entity type
 * 3. Fallback (null - use legacy system)
 */
export async function resolveWorkflowForEntity(
  entityId: string,
  entityType: WorkflowEntityType,
  explicitWorkflowId?: string
): Promise<CustomWorkflow | null> {
  // Priority 1: Explicit workflow
  if (explicitWorkflowId) {
    const workflow = await getWorkflow(explicitWorkflowId);
    if (workflow) {
      return workflow;
    }
  }

  // Priority 2: Default workflow for entity type
  const defaultConfig = await getWorkflowDefault(entityType);
  if (defaultConfig) {
    const workflow = await getWorkflow(defaultConfig.defaultWorkflowId, defaultConfig.workflowVersion);
    if (workflow) {
      return workflow;
    }
  }

  // Priority 3: Fallback to null (caller handles legacy system)
  return null;
}

/**
 * Get first stage of a workflow
 * Determines who should approve at stage 1
 */
export function getFirstStage(workflow: CustomWorkflow): WorkflowStage {
  const firstStage = workflow.stages.find((s) => s.stageNumber === 1);
  if (!firstStage) {
    throw new Error("Workflow must have a stage 1");
  }
  return firstStage;
}

/**
 * Get stage by stage number
 */
export function getStage(workflow: CustomWorkflow, stageNumber: number): WorkflowStage | null {
  return workflow.stages.find((s) => s.stageNumber === stageNumber) || null;
}

// ============================================================================
// APPROVER ASSIGNMENT
// ============================================================================

/**
 * Determine who should approve at a given stage
 *
 * Returns either:
 * - Specific user ID (if USER or ROLE_OR_USER with specificUserId)
 * - Find user with required role (if ROLE)
 * - Null if cannot determine approver
 */
export async function getApproverForStage(
  stage: WorkflowStage
): Promise<{ userId: string; userName: string; role?: UserRole } | null> {
  // Case 1: Specific user assigned
  if (stage.approverAssignmentType === "USER" || stage.approverAssignmentType === "ROLE_OR_USER") {
    if (stage.specificUserId) {
      // Look up user in demo users
      for (const [email, config] of Object.entries(DEMO_USERS)) {
        if (config.user.id === stage.specificUserId) {
          return {
            userId: config.user.id,
            userName: config.user.name,
            role: config.user.role as UserRole,
          };
        }
      }
    }

    // If ROLE_OR_USER and specific user not found, fallback to role
    if (stage.approverAssignmentType === "ROLE_OR_USER" && !stage.specificUserId) {
      // Continue to next case
    } else if (stage.approverAssignmentType === "USER") {
      // Must have specific user
      return null;
    }
  }

  // Case 2: Role-based assignment
  if (stage.approverAssignmentType === "ROLE" || stage.approverAssignmentType === "ROLE_OR_USER") {
    if (!stage.requiredRole) {
      return null;
    }

    // Find user with required role
    const userWithRole = findUserByRole(stage.requiredRole);
    if (userWithRole) {
      return {
        userId: userWithRole.id,
        userName: userWithRole.name,
        role: stage.requiredRole,
      };
    }
  }

  return null;
}

/**
 * Find a user with a specific role
 * In real system, would query database
 * For MVP, searches demo users
 */
function findUserByRole(role: UserRole): { id: string; name: string } | null {
  for (const [, config] of Object.entries(DEMO_USERS)) {
    if (config.user.role === role) {
      return {
        id: config.user.id,
        name: config.user.name,
      };
    }
  }
  return null;
}

// ============================================================================
// STAGE EXECUTION & PROGRESSION
// ============================================================================

/**
 * Move assignment to next stage after approval
 *
 * Returns:
 * - { nextStageNumber: number, isComplete: false } - moved to next stage
 * - { nextStageNumber: null, isComplete: true } - workflow completed
 */
export async function progressToNextStage(
  assignment: WorkflowAssignment,
  workflow: CustomWorkflow,
  approverUserId: string,
  approverUserName: string,
  approverUserRole: UserRole,
  comments?: string,
  signature?: string
): Promise<{
  nextStageNumber: number | null;
  isComplete: boolean;
  nextApprover?: { userId: string; userName: string; role?: UserRole };
}> {
  // Get current stage that was just approved
  const currentStage = getStage(workflow, assignment.currentStageNumber);
  if (!currentStage) {
    throw new Error(`Invalid stage: ${assignment.currentStageNumber}`);
  }

  // Record approval in stage history
  const stageExecution = assignment.stageHistory.find((s) => s.stageNumber === assignment.currentStageNumber);
  if (stageExecution) {
    stageExecution.status = "APPROVED";
    stageExecution.completedAt = new Date();
    stageExecution.completedBy = approverUserId;
    stageExecution.comments = comments;
    stageExecution.signature = signature;

    // Track in assignment history
    if (stageExecution.assignmentHistory) {
      const assignment = stageExecution.assignmentHistory.find((a) => a.status === "ASSIGNED");
      if (assignment) {
        assignment.status = "COMPLETED";
        assignment.completedAt = new Date();
      }
    }
  }

  // Determine next stage from workflow configuration
  const nextStageNum = currentStage.onApprove.nextStage;

  if (nextStageNum === "FINAL") {
    // Workflow completed successfully
    assignment.completedAt = new Date();
    assignment.completedBy = approverUserId;
    return {
      nextStageNumber: null,
      isComplete: true,
    };
  }

  // Move to next stage
  const nextStage = getStage(workflow, nextStageNum as number);
  if (!nextStage) {
    throw new Error(`Invalid next stage: ${nextStageNum}`);
  }

  // Determine who approves at next stage
  const nextApprover = await getApproverForStage(nextStage);
  if (!nextApprover) {
    throw new Error(`Cannot find approver for stage ${nextStage.stageNumber}`);
  }

  // Create stage execution for next stage
  const nextStageExecution: StageExecution = {
    stageNumber: nextStage.stageNumber,
    stageName: nextStage.stageName,
    assignedTo: nextApprover.userId,
    assignedRole: nextApprover.role,
    status: "PENDING",
    startedAt: new Date(),
    assignmentHistory: [
      {
        assignedTo: nextApprover.userId,
        assignedAt: new Date(),
        assignedBy: approverUserId,
        status: "ASSIGNED",
      },
    ],
  };

  assignment.stageHistory.push(nextStageExecution);
  assignment.currentStageNumber = nextStage.stageNumber;
  assignment.stageStartedAt = new Date();

  return {
    nextStageNumber: nextStage.stageNumber,
    isComplete: false,
    nextApprover,
  };
}

/**
 * Reject at a stage
 *
 * Returns where workflow should go
 */
export async function rejectAtStage(
  assignment: WorkflowAssignment,
  workflow: CustomWorkflow,
  rejectorUserId: string,
  rejectorUserName: string,
  rejectorUserRole: UserRole,
  remarks: string,
  signature: string
): Promise<{
  rejectedToStage: number | "DRAFT" | "REJECTED";
  reason: string;
}> {
  // Get current stage that was rejected
  const currentStage = getStage(workflow, assignment.currentStageNumber);
  if (!currentStage) {
    throw new Error(`Invalid stage: ${assignment.currentStageNumber}`);
  }

  // Record rejection in stage history
  const stageExecution = assignment.stageHistory.find((s) => s.stageNumber === assignment.currentStageNumber);
  if (stageExecution) {
    stageExecution.status = "REJECTED";
    stageExecution.completedAt = new Date();
    stageExecution.completedBy = rejectorUserId;
    stageExecution.remarks = remarks;
    stageExecution.signature = signature;

    if (stageExecution.assignmentHistory) {
      const assign = stageExecution.assignmentHistory.find((a) => a.status === "ASSIGNED");
      if (assign) {
        assign.status = "COMPLETED";
        assign.completedAt = new Date();
      }
    }
  }

  // Determine where to send on rejection
  const rejectTarget = currentStage.onReject.nextStage;

  if (rejectTarget === "REJECTED" || rejectTarget === "DRAFT") {
    // Workflow ends (for now - in future, may allow resubmit from DRAFT)
    assignment.currentStageNumber = 0; // Reset to "not started"
    return {
      rejectedToStage: rejectTarget,
      reason: `Rejected at stage ${currentStage.stageNumber}: ${remarks}`,
    };
  }

  // Go back to specified stage
  const targetStage = getStage(workflow, rejectTarget as number);
  if (!targetStage) {
    throw new Error(`Invalid rejection target stage: ${rejectTarget}`);
  }

  // Clear history for stages after target
  assignment.stageHistory = assignment.stageHistory.filter((s) => s.stageNumber <= rejectTarget);

  assignment.currentStageNumber = rejectTarget as number;
  assignment.stageStartedAt = new Date();

  return {
    rejectedToStage: rejectTarget as number,
    reason: `Rejected at stage ${currentStage.stageNumber} and sent back to stage ${rejectTarget}`,
  };
}

// ============================================================================
// REASSIGNMENT LOGIC
// ============================================================================

/**
 * Check if a user can reassign a stage
 *
 * Rules:
 * 1. Currently assigned user can always reassign (if stage allows)
 * 2. ADMIN can reassign any stage (if stage allows)
 * 3. Only if stage.canBeReassigned = true
 */
export async function canReassign(
  assignment: WorkflowAssignment,
  workflow: CustomWorkflow,
  requestingUserId: string
): Promise<{ can: boolean; reason?: string }> {
  const currentStageExecution = assignment.stageHistory.find(
    (s) => s.stageNumber === assignment.currentStageNumber
  );

  if (!currentStageExecution) {
    return { can: false, reason: "Stage not found" };
  }

  const currentStage = getStage(workflow, assignment.currentStageNumber);
  if (!currentStage) {
    return { can: false, reason: "Invalid stage" };
  }

  // Check if stage allows reassignment
  if (!currentStage.canBeReassigned) {
    return { can: false, reason: `Stage "${currentStage.stageName}" does not allow reassignment` };
  }

  // Check if user has permission
  const requestingUser = await getCurrentUser();
  if (!requestingUser) {
    return { can: false, reason: "User not authenticated" };
  }

  // Rule 1: Currently assigned user can reassign
  if (currentStageExecution.assignedTo === requestingUserId) {
    return { can: true };
  }

  // Rule 2: ADMIN can reassign
  if (requestingUser.role === "ADMIN") {
    return { can: true };
  }

  return { can: false, reason: "Only assigned user or admin can reassign" };
}

/**
 * Reassign a stage to a different user
 *
 * Records:
 * - Who was assigned before
 * - Who reassigned
 * - When
 * - Why (optional reason)
 */
export async function reassignStage(
  request: ReassignStageRequest,
  assignment: WorkflowAssignment,
  workflow: CustomWorkflow
): Promise<{
  success: boolean;
  message: string;
  previousApprover?: string;
  newApprover?: string;
}> {
  // Get current stage execution
  const stageExecution = assignment.stageHistory.find(
    (s) => s.stageNumber === request.stageNumber
  );

  if (!stageExecution) {
    return {
      success: false,
      message: `Stage ${request.stageNumber} not found in assignment history`,
    };
  }

  // Cannot reassign completed stages
  if (stageExecution.status !== "PENDING") {
    return {
      success: false,
      message: `Cannot reassign stage ${request.stageNumber}: status is ${stageExecution.status}`,
    };
  }

  const previousApproverId = stageExecution.assignedTo;

  // Update stage execution
  stageExecution.assignedTo = request.newApproverId;

  // Record in assignment history
  if (!stageExecution.assignmentHistory) {
    stageExecution.assignmentHistory = [];
  }

  // Mark previous assignment as reassigned
  const previousAssignment = stageExecution.assignmentHistory.find(
    (a) => a.assignedTo === previousApproverId && a.status === "ASSIGNED"
  );
  if (previousAssignment) {
    previousAssignment.status = "REASSIGNED_TO_OTHER";
  }

  // Add new assignment record
  stageExecution.assignmentHistory.push({
    assignedTo: request.newApproverId,
    assignedAt: new Date(),
    assignedBy: request.reassignedBy,
    reassignmentReason: request.reassignmentReason,
    status: "ASSIGNED",
  });

  return {
    success: true,
    message: `Stage reassigned successfully`,
    previousApprover: previousApproverId,
    newApprover: request.newApproverId,
  };
}

// ============================================================================
// APPROVER LOOKUP
// ============================================================================

/**
 * Get all pending approvals for a user
 * Used by task system to show pending work
 */
export async function getPendingApprovalsForUserId(userId: string): Promise<WorkflowAssignment[]> {
  return getPendingApprovalsForUser(userId);
}

/**
 * Get all approvals that admin can reassign
 */
export async function getAllPendingApprovals(): Promise<WorkflowAssignment[]> {
  // This would come from persistence layer
  // For now, return empty - would be implemented in Phase 2
  const allAssignments: WorkflowAssignment[] = [];
  return allAssignments.filter((a) => {
    const currentStage = a.stageHistory.find((s) => s.stageNumber === a.currentStageNumber);
    return currentStage && currentStage.status === "PENDING";
  });
}

// ============================================================================
// STAGE INFORMATION
// ============================================================================

/**
 * Get detailed information about a stage
 * Used for UI to display requirements
 */
export function getStageInfo(stage: WorkflowStage) {
  return {
    stageName: stage.stageName,
    description: stage.description,
    approverType: stage.approverAssignmentType,
    approverRole: stage.requiredRole,
    requiresSignature: stage.requiresSignature,
    requiresComments: stage.commentsType === "REQUIRED",
    commentsType: stage.commentsType,
    canBeReassigned: stage.canBeReassigned,
    canBeRejected: stage.canBeRejected,
    canBeReversed: stage.canBeReversed,
  };
}

/**
 * Get next stage information (for preview)
 */
export function getNextStageInfo(
  stage: WorkflowStage,
  workflow: CustomWorkflow,
  action: "APPROVE" | "REJECT"
): { stageName?: string; status?: string } | null {
  const nextStageNum = action === "APPROVE" ? stage.onApprove.nextStage : stage.onReject.nextStage;

  if (nextStageNum === "FINAL" || nextStageNum === "REJECTED" || nextStageNum === "DRAFT") {
    return { status: nextStageNum };
  }

  const nextStage = getStage(workflow, nextStageNum as number);
  if (nextStage) {
    return { stageName: nextStage.stageName };
  }

  return null;
}
