/**
 * Workflow Persistence Layer
 * Handles saving, loading, and querying custom workflows
 *
 * Current: File-based storage in-memory with JSON serialization
 * Future: Replace with database storage (PostgreSQL)
 */

import fs from "fs";
import path from "path";
import type {
  CustomWorkflow,
  WorkflowAssignment,
  WorkflowDefault,
  WorkflowEntityType,
} from "@/types";

// ============================================================================
// IN-MEMORY STORE (for MVP)
// ============================================================================

/**
 * In-memory storage maps
 * These will be replaced with database queries in production
 */
let workflowsStore = new Map<string, CustomWorkflow>();
let assignmentsStore = new Map<string, WorkflowAssignment>();
let defaultsStore = new Map<string, WorkflowDefault>();

/**
 * Initialize store from file
 * In production, this would be a database connection pool
 */
function initializeStore() {
  // For MVP, we start with empty maps
  // In production, this would load from database
  workflowsStore.clear();
  assignmentsStore.clear();
  defaultsStore.clear();
}

// Initialize on module load
initializeStore();

// ============================================================================
// WORKFLOW PERSISTENCE
// ============================================================================

/**
 * Save or update a workflow
 * If workflow.id doesn't exist, it's new
 * If it does exist, we create a new version (immutable versioning)
 */
export async function saveWorkflow(workflow: CustomWorkflow): Promise<void> {
  if (!workflow.id) {
    throw new Error("Workflow must have an id");
  }

  // For new workflows, start at version 1
  if (!workflowsStore.has(workflow.id)) {
    workflow.version = 1;
  }

  // Store the workflow
  workflowsStore.set(workflow.id, workflow);

  console.log(`[Workflows] Saved workflow: ${workflow.name} (v${workflow.version})`);
}

/**
 * Get a specific workflow by ID
 * Optionally specify version for historical lookups
 */
export async function getWorkflow(
  workflowId: string,
  version?: number
): Promise<CustomWorkflow | null> {
  const workflow = workflowsStore.get(workflowId);

  if (!workflow) {
    return null;
  }

  // If specific version requested and doesn't match, return null
  // In production DB, we'd query by id AND version
  if (version && workflow.version !== version) {
    return null;
  }

  return workflow;
}

/**
 * Get all workflows with optional filtering
 */
export async function listWorkflows(filters?: {
  entityType?: WorkflowEntityType;
  isActive?: boolean;
  isTemplate?: boolean;
}): Promise<CustomWorkflow[]> {
  let workflows = Array.from(workflowsStore.values());

  // Apply filters
  if (filters?.entityType) {
    workflows = workflows.filter((wf) =>
      wf.applicableEntityTypes.includes(filters.entityType!)
    );
  }

  if (filters?.isActive !== undefined) {
    workflows = workflows.filter((wf) => wf.isActive === filters.isActive);
  }

  if (filters?.isTemplate !== undefined) {
    workflows = workflows.filter((wf) => wf.isTemplate === filters.isTemplate);
  }

  // Sort by creation date (newest first)
  return workflows.sort(
    (a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
  );
}

/**
 * Create a new version of a workflow
 * Keeps old version, creates new one with incremented version number
 */
export async function createWorkflowVersion(
  baseWorkflow: CustomWorkflow
): Promise<CustomWorkflow> {
  const existing = await getWorkflow(baseWorkflow.id);

  if (!existing) {
    throw new Error(`Cannot create version: workflow ${baseWorkflow.id} not found`);
  }

  // Increment version number
  const newWorkflow: CustomWorkflow = {
    ...baseWorkflow,
    version: existing.version + 1,
    updatedAt: new Date(),
  };

  await saveWorkflow(newWorkflow);
  return newWorkflow;
}

/**
 * Mark a workflow as deprecated
 * Cannot be used for new entities, but existing assignments continue
 */
export async function deprecateWorkflow(
  workflowId: string,
  reason?: string
): Promise<CustomWorkflow | null> {
  const workflow = await getWorkflow(workflowId);

  if (!workflow) {
    return null;
  }

  workflow.isActive = false;
  workflow.deprecatedAt = new Date();
  workflow.deprecationReason = reason;

  await saveWorkflow(workflow);
  return workflow;
}

/**
 * Delete a workflow (cascades to check assignments)
 * Should not delete if in use; instead deprecate
 */
export async function deleteWorkflow(workflowId: string): Promise<boolean> {
  const workflow = await getWorkflow(workflowId);

  if (!workflow) {
    return false;
  }

  // Check if workflow is in use
  const usageCount = await countWorkflowUsage(workflowId);
  if (usageCount > 0) {
    throw new Error(
      `Cannot delete workflow ${workflowId}: ${usageCount} entities still using it. Deprecate instead.`
    );
  }

  workflowsStore.delete(workflowId);
  return true;
}

// ============================================================================
// WORKFLOW ASSIGNMENT PERSISTENCE
// ============================================================================

/**
 * Save a workflow assignment
 * Creates when an entity (requisition, budget) is assigned a workflow
 */
export async function saveAssignment(
  assignment: WorkflowAssignment
): Promise<void> {
  if (!assignment.id) {
    throw new Error("Assignment must have an id");
  }

  assignmentsStore.set(assignment.id, assignment);

  console.log(
    `[Workflows] Saved assignment: entity=${assignment.entityId} workflow=${assignment.workflowId}`
  );
}

/**
 * Get assignment for an entity
 */
export async function getAssignmentByEntityId(
  entityId: string,
  entityType: WorkflowEntityType
): Promise<WorkflowAssignment | null> {
  // Find assignment where entityId and entityType match
  for (const assignment of assignmentsStore.values()) {
    if (assignment.entityId === entityId && assignment.entityType === entityType) {
      return assignment;
    }
  }
  return null;
}

/**
 * Get assignment by ID
 */
export async function getAssignment(assignmentId: string): Promise<WorkflowAssignment | null> {
  return assignmentsStore.get(assignmentId) || null;
}

/**
 * List assignments with filtering
 */
export async function listAssignments(filters?: {
  workflowId?: string;
  entityType?: WorkflowEntityType;
  currentStageNumber?: number;
}): Promise<WorkflowAssignment[]> {
  let assignments = Array.from(assignmentsStore.values());

  if (filters?.workflowId) {
    assignments = assignments.filter((a) => a.workflowId === filters.workflowId);
  }

  if (filters?.entityType) {
    assignments = assignments.filter((a) => a.entityType === filters.entityType);
  }

  if (filters?.currentStageNumber !== undefined) {
    assignments = assignments.filter((a) => a.currentStageNumber === filters.currentStageNumber);
  }

  return assignments;
}

/**
 * Update assignment (e.g., when stage completes)
 */
export async function updateAssignment(
  assignmentId: string,
  updates: Partial<WorkflowAssignment>
): Promise<WorkflowAssignment | null> {
  const assignment = await getAssignment(assignmentId);

  if (!assignment) {
    return null;
  }

  const updated = { ...assignment, ...updates };
  await saveAssignment(updated);
  return updated;
}

// ============================================================================
// WORKFLOW DEFAULTS PERSISTENCE
// ============================================================================

/**
 * Set default workflow for an entity type
 */
export async function setWorkflowDefault(
  entityType: WorkflowEntityType,
  workflowId: string,
  version: number,
  userId: string
): Promise<WorkflowDefault> {
  const id = `default-${entityType}`;

  const defaultConfig: WorkflowDefault = {
    id,
    entityType,
    defaultWorkflowId: workflowId,
    workflowVersion: version,
    canEntityOverride: true,
    effectiveDate: new Date(),
    createdBy: userId,
    createdAt: new Date(),
  };

  defaultsStore.set(id, defaultConfig);
  return defaultConfig;
}

/**
 * Get default workflow for an entity type
 */
export async function getWorkflowDefault(
  entityType: WorkflowEntityType
): Promise<WorkflowDefault | null> {
  const id = `default-${entityType}`;
  return defaultsStore.get(id) || null;
}

/**
 * Get all workflow defaults
 */
export async function getAllWorkflowDefaults(): Promise<WorkflowDefault[]> {
  return Array.from(defaultsStore.values());
}

// ============================================================================
// STATISTICS & QUERIES
// ============================================================================

/**
 * Count how many entities are using a workflow
 */
export async function countWorkflowUsage(workflowId: string): Promise<number> {
  const assignments = await listAssignments({ workflowId });
  return assignments.filter((a) => !a.completedAt).length; // Count active assignments
}

/**
 * Get workflow usage statistics
 */
export async function getWorkflowStats(
  workflowId: string
): Promise<{
  totalUsages: number;
  activeAssignments: number;
  completedAssignments: number;
  lastUsed?: Date;
}> {
  const assignments = await listAssignments({ workflowId });

  const stats = {
    totalUsages: assignments.length,
    activeAssignments: assignments.filter((a) => !a.completedAt).length,
    completedAssignments: assignments.filter((a) => a.completedAt).length,
    lastUsed: assignments.length > 0 ? new Date(Math.max(...assignments.map((a) => new Date(a.assignedAt).getTime()))) : undefined,
  };

  return stats;
}

/**
 * Get all pending approvals for a user
 */
export async function getPendingApprovalsForUser(userId: string): Promise<WorkflowAssignment[]> {
  const assignments = Array.from(assignmentsStore.values());

  return assignments.filter((assignment) => {
    // Find if this user is assigned to current stage
    if (!assignment.stageHistory || assignment.stageHistory.length === 0) {
      return false;
    }

    const currentStage = assignment.stageHistory.find(
      (s) => s.stageNumber === assignment.currentStageNumber
    );

    if (!currentStage) {
      return false;
    }

    // Check if user is assigned and stage is pending
    return currentStage.assignedTo === userId && currentStage.status === "PENDING";
  });
}

// ============================================================================
// EXPORT UTILITIES
// ============================================================================

/**
 * Get current store state (for debugging/testing)
 */
export function getStoreState() {
  return {
    workflows: Array.from(workflowsStore.values()),
    assignments: Array.from(assignmentsStore.values()),
    defaults: Array.from(defaultsStore.values()),
  };
}

/**
 * Clear all stores (for testing)
 */
export function clearStores() {
  workflowsStore.clear();
  assignmentsStore.clear();
  defaultsStore.clear();
}

/**
 * Seed with sample workflows (for development)
 */
export async function seedSampleWorkflows(): Promise<void> {
  const sampleWorkflows: CustomWorkflow[] = [
    {
      id: "wf-two-stage-fast",
      name: "2-Stage Fast Track",
      description: "Quick approval workflow for small requisitions",
      version: 1,
      applicableEntityTypes: ["REQUISITION", "BUDGET"],
      isTemplate: true,
      isActive: true,
      stages: [
        {
          stageNumber: 1,
          stageName: "Department Manager Review",
          description: "Manager reviews and approves department requisition",
          approverAssignmentType: "ROLE",
          requiredRole: "DEPARTMENT_MANAGER",
          requiresSignature: true,
          commentsType: "OPTIONAL",
          canBeReassigned: true,
          canBeRejected: true,
          canBeReversed: true,
          displayOrder: 1,
          onApprove: {
            nextStage: 2,
            notifyUsers: true,
          },
          onReject: {
            nextStage: "DRAFT",
            notifyRequester: true,
            requiresRejectionReason: true,
          },
        },
        {
          stageNumber: 2,
          stageName: "Finance Review",
          description: "Finance officer reviews budget and completes approval",
          approverAssignmentType: "ROLE",
          requiredRole: "FINANCE_OFFICER",
          requiresSignature: true,
          commentsType: "OPTIONAL",
          canBeReassigned: true,
          canBeRejected: true,
          canBeReversed: true,
          displayOrder: 2,
          onApprove: {
            nextStage: "FINAL",
            setEntityStatus: "APPROVED",
            notifyUsers: true,
          },
          onReject: {
            nextStage: "DRAFT",
            setEntityStatus: "REJECTED",
            notifyRequester: true,
            requiresRejectionReason: true,
          },
        },
      ],
      totalStages: 2,
      usageCount: 0,
      createdBy: "system",
      createdAt: new Date(),
    },
    {
      id: "wf-four-stage-standard",
      name: "4-Stage Standard",
      description: "Standard approval workflow with multiple levels",
      version: 1,
      applicableEntityTypes: ["REQUISITION", "PURCHASE_ORDER", "BUDGET"],
      isTemplate: true,
      isActive: true,
      stages: [
        {
          stageNumber: 1,
          stageName: "Department Head Review",
          approverAssignmentType: "ROLE",
          requiredRole: "DEPARTMENT_MANAGER",
          requiresSignature: true,
          commentsType: "OPTIONAL",
          canBeReassigned: true,
          canBeRejected: true,
          canBeReversed: true,
          displayOrder: 1,
          onApprove: { nextStage: 2, notifyUsers: true },
          onReject: {
            nextStage: "DRAFT",
            notifyRequester: true,
            requiresRejectionReason: true,
          },
        },
        {
          stageNumber: 2,
          stageName: "Finance Officer Review",
          approverAssignmentType: "ROLE",
          requiredRole: "FINANCE_OFFICER",
          requiresSignature: true,
          commentsType: "REQUIRED",
          canBeReassigned: true,
          canBeRejected: true,
          canBeReversed: true,
          displayOrder: 2,
          onApprove: { nextStage: 3, notifyUsers: true },
          onReject: {
            nextStage: "DRAFT",
            notifyRequester: true,
            requiresRejectionReason: true,
          },
        },
        {
          stageNumber: 3,
          stageName: "Director Approval",
          approverAssignmentType: "ROLE",
          requiredRole: "DIRECTOR",
          requiresSignature: true,
          commentsType: "OPTIONAL",
          canBeReassigned: true,
          canBeRejected: true,
          canBeReversed: true,
          displayOrder: 3,
          onApprove: { nextStage: 4, notifyUsers: true },
          onReject: {
            nextStage: "DRAFT",
            notifyRequester: true,
            requiresRejectionReason: true,
          },
        },
        {
          stageNumber: 4,
          stageName: "Final CFO Sign-off",
          approverAssignmentType: "ROLE",
          requiredRole: "CFO",
          requiresSignature: true,
          commentsType: "OPTIONAL",
          canBeReassigned: true,
          canBeRejected: true,
          canBeReversed: false,
          displayOrder: 4,
          onApprove: {
            nextStage: "FINAL",
            setEntityStatus: "APPROVED",
            notifyUsers: true,
          },
          onReject: {
            nextStage: "DRAFT",
            setEntityStatus: "REJECTED",
            notifyRequester: true,
            requiresRejectionReason: true,
          },
        },
      ],
      totalStages: 4,
      usageCount: 0,
      createdBy: "system",
      createdAt: new Date(),
    },
  ];

  for (const workflow of sampleWorkflows) {
    await saveWorkflow(workflow);
  }

  console.log(`[Workflows] Seeded ${sampleWorkflows.length} sample workflows`);
}
