/**
 * Workflow Validation Layer
 * Validates workflow definitions for consistency and security
 */

import type {
  CustomWorkflow,
  WorkflowStage,
  WorkflowValidationError,
} from "@/types";
import { DEMO_USERS } from "@/lib/auth";

// ============================================================================
// VALIDATION UTILITIES
// ============================================================================

/**
 * Valid user roles in the system
 */
const VALID_ROLES = [
  "REQUESTER",
  "DEPARTMENT_MANAGER",
  "FINANCE_OFFICER",
  "DIRECTOR",
  "CFO",
  "COMPLIANCE_OFFICER",
  "ADMIN",
];

/**
 * Validate a complete workflow definition
 * Returns array of errors (empty array = valid)
 */
export function validateWorkflow(workflow: CustomWorkflow): WorkflowValidationError[] {
  const errors: WorkflowValidationError[] = [];

  // ============================================================================
  // BASIC VALIDATION
  // ============================================================================

  if (!workflow.name || workflow.name.trim().length === 0) {
    errors.push({
      field: "name",
      error: "Workflow name is required",
      severity: "ERROR",
    });
  }

  if (!workflow.applicableEntityTypes || workflow.applicableEntityTypes.length === 0) {
    errors.push({
      field: "applicableEntityTypes",
      error: "At least one entity type must be selected",
      severity: "ERROR",
    });
  }

  if (!workflow.stages || workflow.stages.length === 0) {
    errors.push({
      field: "stages",
      error: "Workflow must have at least one stage",
      severity: "ERROR",
    });
  }

  if (workflow.stages.length > 20) {
    errors.push({
      field: "stages",
      error: "Workflow cannot have more than 20 stages",
      severity: "ERROR",
    });
  }

  if (workflow.totalStages !== workflow.stages.length) {
    errors.push({
      field: "totalStages",
      error: `totalStages (${workflow.totalStages}) must match number of stages (${workflow.stages.length})`,
      severity: "ERROR",
    });
  }

  // ============================================================================
  // STAGE VALIDATION
  // ============================================================================

  const stageNumbers = new Set<number>();

  for (const stage of workflow.stages) {
    const stageErrors = validateStage(stage, workflow.stages.length);
    errors.push(...stageErrors);

    // Check for duplicate stage numbers
    if (stageNumbers.has(stage.stageNumber)) {
      errors.push({
        field: "stageNumber",
        stageNumber: stage.stageNumber,
        error: `Duplicate stage number: ${stage.stageNumber}`,
        severity: "ERROR",
      });
    }
    stageNumbers.add(stage.stageNumber);
  }

  // ============================================================================
  // STAGE ORDER VALIDATION
  // ============================================================================

  // Stages should be in order 1, 2, 3, etc.
  const sortedStages = [...workflow.stages].sort((a, b) => a.stageNumber - b.stageNumber);
  for (let i = 0; i < sortedStages.length; i++) {
    if (sortedStages[i].stageNumber !== i + 1) {
      errors.push({
        field: "stages",
        error: `Stages must be numbered sequentially (1, 2, 3...). Found gap at position ${i + 1}`,
        severity: "ERROR",
      });
      break;
    }
  }

  // ============================================================================
  // STAGE TRANSITION VALIDATION
  // ============================================================================

  for (const stage of workflow.stages) {
    // Validate onApprove.nextStage
    if (stage.onApprove.nextStage !== "FINAL") {
      const nextStageNum = stage.onApprove.nextStage as number;

      if (nextStageNum <= stage.stageNumber) {
        errors.push({
          field: "onApprove",
          stageNumber: stage.stageNumber,
          error: `onApprove.nextStage (${nextStageNum}) must be greater than current stage (${stage.stageNumber}) or 'FINAL'`,
          severity: "ERROR",
        });
      }

      if (nextStageNum > workflow.totalStages) {
        errors.push({
          field: "onApprove",
          stageNumber: stage.stageNumber,
          error: `onApprove.nextStage (${nextStageNum}) exceeds total stages (${workflow.totalStages})`,
          severity: "ERROR",
        });
      }
    }

    // Validate onReject.nextStage
    if (stage.onReject.nextStage !== "REJECTED" && stage.onReject.nextStage !== "DRAFT") {
      const rejectStageNum = stage.onReject.nextStage as number;

      if (rejectStageNum >= stage.stageNumber) {
        errors.push({
          field: "onReject",
          stageNumber: stage.stageNumber,
          error: `onReject.nextStage (${rejectStageNum}) must be less than current stage (${stage.stageNumber}), or 'REJECTED' or 'DRAFT'`,
          severity: "ERROR",
        });
      }
    }

    // Validate onReverse if present
    if (stage.onReverse && stage.onReverse.previousStage !== undefined) {
      if (stage.onReverse.previousStage >= stage.stageNumber) {
        errors.push({
          field: "onReverse",
          stageNumber: stage.stageNumber,
          error: `onReverse.previousStage (${stage.onReverse.previousStage}) must be less than current stage (${stage.stageNumber})`,
          severity: "ERROR",
        });
      }
    }
  }

  // ============================================================================
  // FINAL STAGE VALIDATION
  // ============================================================================

  const lastStage = sortedStages[sortedStages.length - 1];
  if (lastStage && lastStage.onApprove.nextStage !== "FINAL") {
    errors.push({
      field: "onApprove",
      stageNumber: lastStage.stageNumber,
      error: "Last stage must have onApprove.nextStage = 'FINAL'",
      severity: "ERROR",
    });
  }

  // ============================================================================
  // ROLE VALIDATION
  // ============================================================================

  for (const stage of workflow.stages) {
    if (stage.approverAssignmentType === "ROLE" || stage.approverAssignmentType === "ROLE_OR_USER") {
      if (!stage.requiredRole) {
        errors.push({
          field: "requiredRole",
          stageNumber: stage.stageNumber,
          error: `Stage ${stage.stageNumber}: requiredRole is required when approverAssignmentType is ROLE or ROLE_OR_USER`,
          severity: "ERROR",
        });
      } else if (!VALID_ROLES.includes(stage.requiredRole)) {
        errors.push({
          field: "requiredRole",
          stageNumber: stage.stageNumber,
          error: `Stage ${stage.stageNumber}: invalid role "${stage.requiredRole}". Valid roles: ${VALID_ROLES.join(", ")}`,
          severity: "ERROR",
        });
      }
    }

    if (stage.approverAssignmentType === "USER" || stage.approverAssignmentType === "ROLE_OR_USER") {
      if (!stage.specificUserId && !stage.specificUserEmail) {
        errors.push({
          field: "specificUserId",
          stageNumber: stage.stageNumber,
          error: `Stage ${stage.stageNumber}: specificUserId or specificUserEmail required when approverAssignmentType is USER or ROLE_OR_USER`,
          severity: "ERROR",
        });
      }

      // If user email provided, verify user exists in demo users
      if (stage.specificUserEmail) {
        const userExists = Object.keys(DEMO_USERS).includes(stage.specificUserEmail);
        if (!userExists) {
          errors.push({
            field: "specificUserEmail",
            stageNumber: stage.stageNumber,
            error: `Stage ${stage.stageNumber}: user with email "${stage.specificUserEmail}" not found`,
            severity: "WARNING",
          });
        }
      }
    }
  }

  // ============================================================================
  // ESCALATION VALIDATION
  // ============================================================================

  for (const stage of workflow.stages) {
    if (stage.escalationRole && !VALID_ROLES.includes(stage.escalationRole)) {
      errors.push({
        field: "escalationRole",
        stageNumber: stage.stageNumber,
        error: `Stage ${stage.stageNumber}: invalid escalation role "${stage.escalationRole}"`,
        severity: "ERROR",
      });
    }
  }

  // ============================================================================
  // NO INFINITE LOOPS
  // ============================================================================

  // Check for stage transitions that could create loops
  for (const stage of workflow.stages) {
    // If approval sends to stage X, and stage X sends back to current stage, it's a loop
    if (typeof stage.onApprove.nextStage === "number") {
      const nextStage = workflow.stages.find((s) => s.stageNumber === stage.onApprove.nextStage);
      if (nextStage && typeof nextStage.onReject.nextStage === "number") {
        if (nextStage.onReject.nextStage === stage.stageNumber) {
          errors.push({
            field: "transitions",
            stageNumber: stage.stageNumber,
            error: `Potential infinite loop: Stage ${stage.stageNumber} → ${nextStage.stageNumber} → ${stage.stageNumber}`,
            severity: "WARNING",
          });
        }
      }
    }
  }

  return errors;
}

/**
 * Validate a single workflow stage
 */
export function validateStage(
  stage: WorkflowStage,
  totalStages: number
): WorkflowValidationError[] {
  const errors: WorkflowValidationError[] = [];

  // Basic validation
  if (!stage.stageName || stage.stageName.trim().length === 0) {
    errors.push({
      field: "stageName",
      stageNumber: stage.stageNumber,
      error: "Stage name is required",
      severity: "ERROR",
    });
  }

  if (stage.stageNumber < 1 || stage.stageNumber > totalStages) {
    errors.push({
      field: "stageNumber",
      stageNumber: stage.stageNumber,
      error: `Stage number must be between 1 and ${totalStages}`,
      severity: "ERROR",
    });
  }

  if (!stage.approverAssignmentType) {
    errors.push({
      field: "approverAssignmentType",
      stageNumber: stage.stageNumber,
      error: "Approver assignment type is required",
      severity: "ERROR",
    });
  }

  if (!stage.onApprove) {
    errors.push({
      field: "onApprove",
      stageNumber: stage.stageNumber,
      error: "onApprove transition is required",
      severity: "ERROR",
    });
  }

  if (!stage.onReject) {
    errors.push({
      field: "onReject",
      stageNumber: stage.stageNumber,
      error: "onReject transition is required",
      severity: "ERROR",
    });
  }

  // Comments type validation
  if (!["OPTIONAL", "REQUIRED", "DISABLED"].includes(stage.commentsType)) {
    errors.push({
      field: "commentsType",
      stageNumber: stage.stageNumber,
      error: "Comments type must be OPTIONAL, REQUIRED, or DISABLED",
      severity: "ERROR",
    });
  }

  return errors;
}

/**
 * Check if a workflow is valid (has no errors)
 */
export function isWorkflowValid(workflow: CustomWorkflow): boolean {
  const errors = validateWorkflow(workflow);
  return errors.every((e) => e.severity !== "ERROR");
}

/**
 * Get only error-level validations (not warnings)
 */
export function getWorkflowErrors(workflow: CustomWorkflow): WorkflowValidationError[] {
  return validateWorkflow(workflow).filter((e) => e.severity === "ERROR");
}

/**
 * Get only warning-level validations
 */
export function getWorkflowWarnings(workflow: CustomWorkflow): WorkflowValidationError[] {
  return validateWorkflow(workflow).filter((e) => e.severity === "WARNING");
}

/**
 * Format validation errors for display
 */
export function formatValidationErrors(errors: WorkflowValidationError[]): string {
  if (errors.length === 0) return "✅ Workflow is valid";

  const grouped = errors.reduce(
    (acc, error) => {
      const key = error.stageNumber ? `Stage ${error.stageNumber}` : error.field;
      if (!acc[key]) acc[key] = [];
      acc[key].push(`${error.error} (${error.severity})`);
      return acc;
    },
    {} as Record<string, string[]>
  );

  return Object.entries(grouped)
    .map(([key, msgs]) => `${key}:\n  - ${msgs.join("\n  - ")}`)
    .join("\n");
}
