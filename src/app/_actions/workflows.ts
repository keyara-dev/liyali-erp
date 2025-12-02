/**
 * Custom Workflow Server Actions
 *
 * Server actions for workflow CRUD operations, validation, and orchestration.
 * These actions handle the complete workflow lifecycle:
 * - Creating and managing reusable workflows
 * - Assigning workflows to entities (requisitions, budgets)
 * - Progressing through approval stages
 * - Handling rejections and reversals
 * - Managing task reassignments
 */

'use server';

import {
  CustomWorkflow,
  WorkflowAssignment,
  WorkflowEntityType,
  CreateWorkflowRequest,
  UpdateWorkflowRequest,
  AssignWorkflowRequest,
  ApproveStageRequest,
  RejectStageRequest,
  ReassignStageRequest,
  ReverseStageRequest,
  StageExecution,
} from '@/types';

import {
  saveWorkflow,
  getWorkflow,
  listWorkflows,
  deprecateWorkflow,
  createWorkflowVersion,
  saveAssignment,
  getAssignment,
  getAssignmentByEntityId,
  updateAssignment,
  setWorkflowDefault,
  getWorkflowDefault,
  countWorkflowUsage,
  getPendingApprovalsForUser,
} from '@/lib/workflow-persistence';

import {
  validateWorkflow,
  getWorkflowErrors,
} from '@/lib/workflow-validation';

import {
  resolveWorkflowForEntity,
  getApproverForStage,
  progressToNextStage,
  rejectAtStage,
  reassignStage,
  canReassign,
  getStageInfo,
  getNextStageInfo,
  getPendingApprovalsForUserId,
} from '@/lib/workflow-resolution';

import {
  notifyTaskAssigned,
  notifyTaskReassigned,
  notifyTaskApproved,
  notifyTaskRejected,
  notifyWorkflowComplete,
} from './notifications';

import { type UserRole } from '@/lib/auth';
import { v4 as uuid } from 'uuid';

/**
 * Create a new workflow
 * @param request Workflow creation request
 * @returns Created workflow
 */
export async function createWorkflow(
  request: CreateWorkflowRequest
): Promise<{ workflow: CustomWorkflow; success: boolean }> {
  try {
    // Create workflow object
    const workflow: CustomWorkflow = {
      id: uuid(),
      name: request.name,
      description: request.description,
      version: 1,
      applicableEntityTypes: request.applicableEntityTypes,
      isTemplate: request.isTemplate ?? true,
      isActive: true,
      stages: request.stages,
      totalStages: request.stages.length,
      usageCount: 0,
      createdBy: request.createdBy,
      createdAt: new Date(),
      updatedAt: new Date(),
    };

    // Validate workflow
    const errors = validateWorkflow(workflow);
    if (errors.length > 0) {
      throw new Error(`Workflow validation failed: ${errors.map((e) => e.error).join(', ')}`);
    }

    // Save workflow
    await saveWorkflow(workflow);

    return {
      workflow,
      success: true,
    };
  } catch (error) {
    console.error('[createWorkflow] Error:', error);
    throw new Error('Failed to create workflow');
  }
}

/**
 * Get a workflow by ID
 * @param workflowId Workflow ID
 * @param version Optional specific version
 * @returns Workflow
 */
export async function getWorkflowAction(
  workflowId: string,
  version?: number
): Promise<CustomWorkflow | null> {
  try {
    if (!workflowId) {
      throw new Error('Workflow ID is required');
    }

    const workflow = await getWorkflow(workflowId, version);
    return workflow || null;
  } catch (error) {
    console.error('[getWorkflow] Error:', error);
    throw new Error('Failed to fetch workflow');
  }
}

/**
 * List workflows with optional filters
 * @param entityType Optional entity type filter
 * @param onlyActive Only active workflows
 * @returns Array of workflows
 */
export async function listWorkflowsAction(
  entityType?: WorkflowEntityType,
  onlyActive: boolean = true
): Promise<CustomWorkflow[]> {
  try {
    const workflows = await listWorkflows({
      entityType,
      isActive: onlyActive,
    });

    return workflows;
  } catch (error) {
    console.error('[listWorkflows] Error:', error);
    throw new Error('Failed to list workflows');
  }
}

/**
 * Update a workflow (creates new version)
 * @param request Update request
 * @returns New workflow version
 */
export async function updateWorkflowAction(
  request: UpdateWorkflowRequest
): Promise<{ workflow: CustomWorkflow; success: boolean }> {
  try {
    if (!request.workflowId) {
      throw new Error('Workflow ID is required');
    }

    const existing = await getWorkflow(request.workflowId);
    if (!existing) {
      throw new Error('Workflow not found');
    }

    // Create new version
    const updated: CustomWorkflow = {
      ...existing,
      ...(request.name && { name: request.name }),
      ...(request.description && { description: request.description }),
      ...(request.applicableEntityTypes && { applicableEntityTypes: request.applicableEntityTypes }),
      ...(request.stages && { stages: request.stages }),
      version: existing.version + 1,
      updatedAt: new Date(),
      updatedBy: request.updatedBy,
    };

    // Validate
    const errors = validateWorkflow(updated);
    if (errors.length > 0) {
      throw new Error(`Workflow validation failed: ${errors.map((e) => e.error).join(', ')}`);
    }

    // Save new version
    await createWorkflowVersion(updated);

    return {
      workflow: updated,
      success: true,
    };
  } catch (error) {
    console.error('[updateWorkflow] Error:', error);
    throw new Error('Failed to update workflow');
  }
}

/**
 * Deprecate a workflow (mark as inactive)
 * @param workflowId Workflow ID
 * @returns Updated workflow
 */
export async function deprecateWorkflowAction(
  workflowId: string
): Promise<{ workflow: CustomWorkflow; success: boolean }> {
  try {
    if (!workflowId) {
      throw new Error('Workflow ID is required');
    }

    const workflow = await deprecateWorkflow(workflowId);
    if (!workflow) {
      throw new Error('Workflow not found');
    }

    return {
      workflow,
      success: true,
    };
  } catch (error) {
    console.error('[deprecateWorkflow] Error:', error);
    throw new Error('Failed to deprecate workflow');
  }
}

/**
 * Assign a workflow to an entity
 * @param request Assignment request
 * @returns Created assignment
 */
export async function assignWorkflowAction(
  request: AssignWorkflowRequest
): Promise<{ assignment: WorkflowAssignment; success: boolean }> {
  try {
    if (!request.entityId || !request.entityType || !request.workflowId) {
      throw new Error('Entity ID, type, and workflow ID are required');
    }

    const workflow = await getWorkflow(request.workflowId);
    if (!workflow) {
      throw new Error('Workflow not found');
    }

    if (!workflow.isActive) {
      throw new Error('Workflow is not active');
    }

    if (!workflow.applicableEntityTypes.includes(request.entityType)) {
      throw new Error('Workflow is not applicable to this entity type');
    }

    const assignment: WorkflowAssignment = {
      id: uuid(),
      entityId: request.entityId,
      entityType: request.entityType,
      workflowId: workflow.id,
      workflowVersion: workflow.version,
      currentStageNumber: 0,
      stageHistory: [],
      assignedAt: new Date(),
      assignedBy: request.assignedBy,
    };

    await saveAssignment(assignment);

    return {
      assignment,
      success: true,
    };
  } catch (error) {
    console.error('[assignWorkflow] Error:', error);
    throw new Error('Failed to assign workflow');
  }
}

/**
 * Get assignment for an entity
 * @param entityId Entity ID
 * @param entityType Entity type
 * @returns Assignment or null
 */
export async function getAssignmentAction(entityId: string, entityType: WorkflowEntityType): Promise<WorkflowAssignment | null> {
  try {
    if (!entityId) {
      throw new Error('Entity ID is required');
    }

    const assignment = await getAssignmentByEntityId(entityId, entityType);
    return assignment || null;
  } catch (error) {
    console.error('[getAssignment] Error:', error);
    throw new Error('Failed to fetch assignment');
  }
}

/**
 * Progress workflow to next stage (approve)
 * @param request Approval request
 * @returns Updated assignment
 */
export async function approveStageAction(
  request: ApproveStageRequest
): Promise<{ assignment: WorkflowAssignment; nextApprover?: { userId: string; userName: string; role?: UserRole } | null; isComplete: boolean }> {
  try {
    if (!request.assignmentId) {
      throw new Error('Assignment ID is required');
    }

    // Get assignment and workflow
    const assignment = await getAssignment(request.assignmentId);
    if (!assignment) {
      throw new Error('Assignment not found');
    }

    const workflow = await getWorkflow(assignment.workflowId, assignment.workflowVersion);
    if (!workflow) {
      throw new Error('Workflow not found');
    }

    // Progress to next stage
    const result = await progressToNextStage(
      assignment,
      workflow,
      request.approvingUserId,
      '', // approverName not available in request
      'USER' as any, // approverRole not available in request, using default
      request.comments,
      request.signature
    );

    // Notify next approver if not complete
    if (!result.isComplete && result.nextApprover) {
      const currentStage = workflow.stages.find((s) => s.stageNumber === assignment.currentStageNumber);
      if (currentStage) {
        await notifyTaskAssigned(
          result.nextApprover.userId,
          result.nextApprover.userName,
          assignment.entityId,
          assignment.entityType,
          assignment.entityId,
          currentStage.stageName
        );
      }
    }

    // Notify creator if complete
    if (result.isComplete) {
      await notifyWorkflowComplete(
        assignment.assignedBy,
        assignment.entityId,
        assignment.entityType,
        assignment.entityId,
        request.approvingUserId,
        ''
      );
    }

    // Notify creator of approval (if not final stage)
    if (!result.isComplete) {
      await notifyTaskApproved(
        assignment.assignedBy,
        assignment.entityId,
        assignment.entityType,
        assignment.entityId,
        request.approvingUserId,
        ''
      );
    }

    // Update assignment in storage
    const updated = await updateAssignment(assignment.id, assignment);
    if (!updated) {
      throw new Error('Failed to update assignment');
    }

    return {
      assignment: updated,
      nextApprover: result.nextApprover,
      isComplete: result.isComplete,
    };
  } catch (error) {
    console.error('[approveStage] Error:', error);
    throw new Error('Failed to approve stage');
  }
}

/**
 * Reject at current stage
 * @param request Rejection request
 * @returns Updated assignment
 */
export async function rejectStageAction(
  request: RejectStageRequest
): Promise<{ assignment: WorkflowAssignment; targetStage: string }> {
  try {
    if (!request.assignmentId) {
      throw new Error('Assignment ID is required');
    }

    // Get assignment and workflow
    const assignment = await getAssignment(request.assignmentId);
    if (!assignment) {
      throw new Error('Assignment not found');
    }

    const workflow = await getWorkflow(assignment.workflowId, assignment.workflowVersion);
    if (!workflow) {
      throw new Error('Workflow not found');
    }

    // Reject at stage
    const result = await rejectAtStage(
      assignment,
      workflow,
      request.rejectingUserId,
      '', // rejectorUserName not available in request
      'USER' as any, // rejectorUserRole not available in request, using default
      request.remarks,
      request.signature
    );

    // Notify creator of rejection
    await notifyTaskRejected(
      assignment.assignedBy,
      assignment.entityId,
      assignment.entityType,
      assignment.entityId,
      request.rejectingUserId,
      '',
      request.remarks
    );

    // Update assignment
    const updated = await updateAssignment(assignment.id, assignment);
    if (!updated) {
      throw new Error('Failed to update assignment');
    }

    return {
      assignment: updated,
      targetStage: result.rejectedToStage.toString(),
    };
  } catch (error) {
    console.error('[rejectStage] Error:', error);
    throw new Error('Failed to reject stage');
  }
}

/**
 * Reassign a stage to a different user
 * @param request Reassignment request
 * @returns Updated assignment
 */
export async function reassignStageAction(
  request: ReassignStageRequest
): Promise<{
  assignment: WorkflowAssignment;
  oldApprover: { userId: string; name: string } | null;
  newApprover: { userId: string; name: string };
}> {
  try {
    if (!request.assignmentId) {
      throw new Error('Assignment ID is required');
    }

    // Get assignment and workflow
    const assignment = await getAssignment(request.assignmentId);
    if (!assignment) {
      throw new Error('Assignment not found');
    }

    const workflow = await getWorkflow(assignment.workflowId, assignment.workflowVersion);
    if (!workflow) {
      throw new Error('Workflow not found');
    }

    // Check permission
    const permission = await canReassign(assignment, workflow, request.reassignedBy);
    if (!permission.can) {
      throw new Error(permission.reason || 'You do not have permission to reassign this task');
    }

    // Perform reassignment
    const result = await reassignStage(
      {
        assignmentId: request.assignmentId,
        stageNumber: request.stageNumber,
        newApproverId: request.newApproverId,
        reassignedBy: request.reassignedBy,
        reassignmentReason: request.reassignmentReason,
      },
      assignment,
      workflow
    );

    // Notify new approver
    await notifyTaskReassigned(
      request.newApproverId,
      '', // newApprover name not available in result
      assignment.entityId,
      assignment.entityType,
      assignment.entityId,
      request.reassignedBy,
      '',
      request.reassignmentReason
    );

    // Update assignment
    const updated = await updateAssignment(assignment.id, assignment);
    if (!updated) {
      throw new Error('Failed to update assignment');
    }

    return {
      assignment: updated,
      oldApprover: result.previousApprover ? { userId: result.previousApprover, name: result.previousApprover } : null,
      newApprover: { userId: request.newApproverId, name: '' },
    };
  } catch (error) {
    console.error('[reassignStage] Error:', error);
    throw new Error('Failed to reassign stage');
  }
}

/**
 * Get pending approvals for a user
 * @param userId User ID
 * @returns Array of pending assignments
 */
export async function getPendingApprovalsAction(
  userId: string
): Promise<WorkflowAssignment[]> {
  try {
    if (!userId) {
      throw new Error('User ID is required');
    }

    return await getPendingApprovalsForUserId(userId);
  } catch (error) {
    console.error('[getPendingApprovals] Error:', error);
    throw new Error('Failed to fetch pending approvals');
  }
}

/**
 * Set default workflow for entity type
 * @param entityType Entity type
 * @param workflowId Workflow ID
 * @param userId User ID setting the default
 * @returns Success status
 */
export async function setDefaultWorkflowAction(
  entityType: WorkflowEntityType,
  workflowId: string,
  userId: string
): Promise<{ success: boolean }> {
  try {
    if (!entityType || !workflowId || !userId) {
      throw new Error('Entity type, workflow ID, and user ID are required');
    }

    const workflow = await getWorkflow(workflowId);
    if (!workflow) {
      throw new Error('Workflow not found');
    }

    await setWorkflowDefault(entityType, workflowId, workflow.version, userId);

    return { success: true };
  } catch (error) {
    console.error('[setDefaultWorkflow] Error:', error);
    throw new Error('Failed to set default workflow');
  }
}

/**
 * Get default workflow for entity type
 * @param entityType Entity type
 * @returns Default workflow or null
 */
export async function getDefaultWorkflowAction(
  entityType: WorkflowEntityType
): Promise<CustomWorkflow | null> {
  try {
    if (!entityType) {
      throw new Error('Entity type is required');
    }

    const workflowDefault = await getWorkflowDefault(entityType);
    if (!workflowDefault) {
      return null;
    }

    return await getWorkflow(workflowDefault.defaultWorkflowId);
  } catch (error) {
    console.error('[getDefaultWorkflow] Error:', error);
    throw new Error('Failed to fetch default workflow');
  }
}
