/**
 * Workflow React Query Hooks
 *
 * Custom React hooks for managing workflow data fetching, caching, and mutations.
 * Uses React Query (TanStack Query) for efficient state management and real-time updates.
 */

'use client';

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  CustomWorkflow,
  WorkflowAssignment,
  WorkflowEntityType,
  CreateWorkflowRequest,
  UpdateWorkflowRequest,
  AssignWorkflowRequest,
} from '@/types';

import {
  createWorkflow,
  getWorkflowAction,
  listWorkflowsAction,
  updateWorkflowAction,
  deprecateWorkflowAction,
  assignWorkflowAction,
  getAssignmentAction,
  getPendingApprovalsAction,
  setDefaultWorkflowAction,
  getDefaultWorkflowAction,
} from '@/app/_actions/workflows';

const WORKFLOWS_QUERY_KEY = 'workflows';
const WORKFLOW_QUERY_KEY = 'workflow';
const ASSIGNMENTS_QUERY_KEY = 'assignments';
const PENDING_APPROVALS_QUERY_KEY = 'pending-approvals';
const DEFAULT_WORKFLOWS_QUERY_KEY = 'default-workflows';

/**
 * Hook: Get all workflows
 * @param entityType Optional filter by entity type
 * @param onlyActive Filter to active workflows only
 * @returns Query result with workflows
 */
export function useWorkflows(
  entityType?: WorkflowEntityType,
  onlyActive: boolean = true
) {
  return useQuery({
    queryKey: [WORKFLOWS_QUERY_KEY, entityType, onlyActive],
    queryFn: async () => listWorkflowsAction(entityType, onlyActive),
    staleTime: 5 * 60 * 1000, // 5 minutes
  });
}

/**
 * Hook: Get a single workflow by ID
 * @param workflowId Workflow ID
 * @param version Optional specific version
 * @returns Query result with workflow
 */
export function useWorkflow(workflowId: string, version?: number) {
  return useQuery({
    queryKey: [WORKFLOW_QUERY_KEY, workflowId, version],
    queryFn: async () => getWorkflowAction(workflowId, version),
    enabled: !!workflowId,
    staleTime: 10 * 60 * 1000, // 10 minutes
  });
}

/**
 * Hook: Get default workflow for entity type
 * @param entityType Entity type
 * @returns Query result with default workflow
 */
export function useDefaultWorkflow(entityType: WorkflowEntityType) {
  return useQuery({
    queryKey: [DEFAULT_WORKFLOWS_QUERY_KEY, entityType],
    queryFn: async () => getDefaultWorkflowAction(entityType),
    enabled: !!entityType,
    staleTime: 30 * 60 * 1000, // 30 minutes
  });
}

/**
 * Hook: Get assignment for an entity
 * @param entityId Entity ID
 * @param entityType Entity type
 * @returns Query result with assignment
 */
export function useAssignment(entityId: string, entityType: WorkflowEntityType) {
  return useQuery({
    queryKey: [ASSIGNMENTS_QUERY_KEY, entityId, entityType],
    queryFn: async () => getAssignmentAction(entityId, entityType),
    enabled: !!entityId && !!entityType,
    staleTime: 1 * 60 * 1000, // 1 minute (frequently updated)
    refetchInterval: 30 * 1000, // Auto-refresh every 30 seconds
  });
}

/**
 * Hook: Get pending approvals for a user
 * @param userId User ID
 * @returns Query result with pending assignments
 */
export function usePendingApprovals(userId: string) {
  return useQuery({
    queryKey: [PENDING_APPROVALS_QUERY_KEY, userId],
    queryFn: async () => getPendingApprovalsAction(userId),
    enabled: !!userId,
    staleTime: 30 * 1000, // 30 seconds
    refetchInterval: 60 * 1000, // Auto-refresh every 60 seconds
  });
}

/**
 * Hook: Get pending approvals count for a user
 * @param userId User ID
 * @returns Count of pending approvals
 */
export function usePendingApprovalsCount(userId: string) {
  const { data } = usePendingApprovals(userId);
  return data?.length || 0;
}

/**
 * Hook: Create a workflow
 * @returns Mutation for creating workflow
 */
export function useCreateWorkflow() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (request: CreateWorkflowRequest) =>
      createWorkflow(request),
    onSuccess: (data) => {
      // Invalidate workflows list
      queryClient.invalidateQueries({
        queryKey: [WORKFLOWS_QUERY_KEY],
      });

      // Cache the new workflow
      queryClient.setQueryData(
        [WORKFLOW_QUERY_KEY, data.workflow.id, data.workflow.version],
        data.workflow
      );
    },
    onError: (error) => {
      console.error('Failed to create workflow:', error);
    },
  });
}

/**
 * Hook: Update a workflow (creates new version)
 * @returns Mutation for updating workflow
 */
export function useUpdateWorkflow() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (request: UpdateWorkflowRequest) =>
      updateWorkflowAction(request),
    onSuccess: (data) => {
      // Invalidate workflows list
      queryClient.invalidateQueries({
        queryKey: [WORKFLOWS_QUERY_KEY],
      });

      // Cache new version
      queryClient.setQueryData(
        [WORKFLOW_QUERY_KEY, data.workflow.id, data.workflow.version],
        data.workflow
      );

      // Invalidate old versions
      queryClient.invalidateQueries({
        queryKey: [WORKFLOW_QUERY_KEY, data.workflow.id],
      });
    },
    onError: (error) => {
      console.error('Failed to update workflow:', error);
    },
  });
}

/**
 * Hook: Deprecate a workflow
 * @returns Mutation for deprecating workflow
 */
export function useDeprecateWorkflow() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (workflowId: string) =>
      deprecateWorkflowAction(workflowId),
    onSuccess: (data) => {
      // Invalidate all workflow queries
      queryClient.invalidateQueries({
        queryKey: [WORKFLOWS_QUERY_KEY],
      });
      queryClient.invalidateQueries({
        queryKey: [WORKFLOW_QUERY_KEY, data.workflow.id],
      });
      queryClient.invalidateQueries({
        queryKey: [DEFAULT_WORKFLOWS_QUERY_KEY],
      });
    },
    onError: (error) => {
      console.error('Failed to deprecate workflow:', error);
    },
  });
}

/**
 * Hook: Assign workflow to entity
 * @returns Mutation for assigning workflow
 */
export function useAssignWorkflow() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (request: AssignWorkflowRequest) =>
      assignWorkflowAction(request),
    onSuccess: (data) => {
      // Cache the assignment
      queryClient.setQueryData(
        [ASSIGNMENTS_QUERY_KEY, data.assignment.entityId],
        data.assignment
      );
    },
    onError: (error) => {
      console.error('Failed to assign workflow:', error);
    },
  });
}

/**
 * Hook: Set default workflow for entity type
 * @returns Mutation for setting default
 */
export function useSetDefaultWorkflow() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (params: {
      entityType: WorkflowEntityType;
      workflowId: string;
      userId: string;
    }) => setDefaultWorkflowAction(params.entityType, params.workflowId, params.userId),
    onSuccess: (data, variables) => {
      // Invalidate defaults
      queryClient.invalidateQueries({
        queryKey: [DEFAULT_WORKFLOWS_QUERY_KEY, variables.entityType],
      });
    },
    onError: (error) => {
      console.error('Failed to set default workflow:', error);
    },
  });
}

/**
 * Hook: Invalidate all workflow queries
 * @returns Function to invalidate
 */
export function useInvalidateWorkflows() {
  const queryClient = useQueryClient();

  return (entityId?: string) => {
    if (entityId) {
      queryClient.invalidateQueries({
        queryKey: [ASSIGNMENTS_QUERY_KEY, entityId],
      });
    } else {
      queryClient.invalidateQueries({
        queryKey: [WORKFLOWS_QUERY_KEY],
      });
      queryClient.invalidateQueries({
        queryKey: [WORKFLOW_QUERY_KEY],
      });
      queryClient.invalidateQueries({
        queryKey: [ASSIGNMENTS_QUERY_KEY],
      });
      queryClient.invalidateQueries({
        queryKey: [DEFAULT_WORKFLOWS_QUERY_KEY],
      });
    }
  };
}

/**
 * Hook: Invalidate pending approvals
 * @returns Function to invalidate
 */
export function useInvalidatePendingApprovals() {
  const queryClient = useQueryClient();

  return (userId?: string) => {
    if (userId) {
      queryClient.invalidateQueries({
        queryKey: [PENDING_APPROVALS_QUERY_KEY, userId],
      });
    } else {
      queryClient.invalidateQueries({
        queryKey: [PENDING_APPROVALS_QUERY_KEY],
      });
    }
  };
}

/**
 * Hook: Get workflows for selection dropdown
 * Shows only active templates applicable to entity type
 * @param entityType Entity type to filter
 * @returns Formatted list for dropdown
 */
export function useWorkflowsForSelection(entityType: WorkflowEntityType) {
  const { data, isLoading } = useWorkflows(entityType, true);

  return {
    workflows: data?.filter((w) => w.isTemplate) || [],
    isLoading,
    hasOptions: (data?.length || 0) > 0,
  };
}

/**
 * Hook: Get workflow with fallback to default
 * Attempts to get specified workflow, falls back to default if not found
 * @param entityType Entity type
 * @param specifiedWorkflowId Optional specific workflow ID
 * @returns Query result with workflow or default
 */
export function useWorkflowWithFallback(
  entityType: WorkflowEntityType,
  specifiedWorkflowId?: string
) {
  const specificQuery = useWorkflow(specifiedWorkflowId || '');
  const defaultQuery = useDefaultWorkflow(entityType);

  const workflow = specifiedWorkflowId
    ? specificQuery.data
    : defaultQuery.data;

  return {
    workflow,
    isLoading:
      (specifiedWorkflowId ? specificQuery.isLoading : defaultQuery.isLoading) ||
      false,
    error:
      (specifiedWorkflowId ? specificQuery.error : defaultQuery.error) || null,
    isDefault: !specifiedWorkflowId,
  };
}

/**
 * Hook: Check if user has pending approvals
 * @param userId User ID
 * @returns Boolean indicating if has pending approvals
 */
export function useHasPendingApprovals(userId: string) {
  const { data } = usePendingApprovals(userId);
  return (data?.length || 0) > 0;
}

/**
 * Hook: Get workflow stages for display
 * Returns workflow stages with stage info
 * @param workflowId Workflow ID
 * @returns Workflow with formatted stages
 */
export function useWorkflowStages(workflowId: string) {
  const { data: workflow, ...rest } = useWorkflow(workflowId);

  return {
    stages: workflow?.stages || [],
    totalStages: workflow?.totalStages || 0,
    workflow,
    ...rest,
  };
}

/**
 * Hook: Get assignment with workflow details
 * Combines assignment and workflow queries
 * @param entityId Entity ID
 * @param entityType Entity type
 * @returns Assignment with resolved workflow
 */
export function useAssignmentWithWorkflow(entityId: string, entityType: WorkflowEntityType) {
  const assignmentQuery = useAssignment(entityId, entityType);
  const assignment = assignmentQuery.data;

  const workflowQuery = useWorkflow(
    assignment?.workflowId || '',
    assignment?.workflowVersion
  );

  return {
    assignment,
    workflow: workflowQuery.data,
    isLoading: assignmentQuery.isLoading || workflowQuery.isLoading,
    error: assignmentQuery.error || workflowQuery.error,
    currentStage: assignment
      ? workflowQuery.data?.stages.find(
          (s) => s.stageNumber === assignment.currentStageNumber
        )
      : undefined,
  };
}

/**
 * Hook: Workflow statistics
 * Get usage and status info
 * @param workflowId Workflow ID
 * @returns Workflow with stats
 */
export function useWorkflowStats(workflowId: string) {
  const { data: workflow, ...rest } = useWorkflow(workflowId);

  return {
    name: workflow?.name || '',
    version: workflow?.version || 1,
    usageCount: workflow?.usageCount || 0,
    isActive: workflow?.isActive || false,
    applicableEntityTypes: workflow?.applicableEntityTypes || [],
    stageCount: workflow?.totalStages || 0,
    ...rest,
  };
}

// Re-export approval task query hooks from use-approval-task-queries
export {
  useGetApprovalTasks,
  useGetApprovalTaskDetail,
  useGetApprovalStats,
  useGetTaskHistory,
} from './use-approval-task-queries';
