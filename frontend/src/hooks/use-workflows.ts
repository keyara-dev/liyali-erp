/**
 * Workflow React Query Hooks
 *
 * Custom React hooks for managing workflow data fetching, caching, and mutations.
 * Uses React Query (TanStack Query) for efficient state management and real-time updates.
 */

'use client';

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Workflow, WorkflowFormData, WorkflowListFilter } from '@/app/_actions/workflows';

import {
  createWorkflow,
  getWorkflowById,
  getWorkflows,
  updateWorkflow,
  deleteWorkflow,
  activateWorkflow,
  deactivateWorkflow,
  setDefaultWorkflow,
  getDefaultWorkflow,
} from '@/app/_actions/workflows';

const WORKFLOWS_QUERY_KEY = 'workflows';
const WORKFLOW_QUERY_KEY = 'workflow';
const DEFAULT_WORKFLOWS_QUERY_KEY = 'default-workflows';

/**
 * Hook: Get all workflows with optional filtering
 * @param filter Optional filter criteria
 * @returns Query result with workflows array
 */
export function useWorkflows(filter?: WorkflowListFilter) {
  return useQuery({
    queryKey: [WORKFLOWS_QUERY_KEY, filter],
    queryFn: async () => getWorkflows(filter),
    staleTime: 5 * 60 * 1000, // 5 minutes
  });
}

/**
 * Hook: Get single workflow by ID
 * @param workflowId Workflow ID
 * @returns Query result with workflow
 */
export function useWorkflow(workflowId: string) {
  return useQuery({
    queryKey: [WORKFLOW_QUERY_KEY, workflowId],
    queryFn: async () => getWorkflowById(workflowId),
    enabled: !!workflowId,
    staleTime: 10 * 60 * 1000, // 10 minutes
  });
}

/**
 * Hook: Get default workflow for entity type
 * @param entityType Entity type
 * @returns Query result with default workflow
 */
export function useDefaultWorkflow(entityType: string) {
  return useQuery({
    queryKey: [DEFAULT_WORKFLOWS_QUERY_KEY, entityType],
    queryFn: async () => getDefaultWorkflow(entityType),
    enabled: !!entityType,
    staleTime: 30 * 60 * 1000, // 30 minutes
  });
}

/**
 * Hook: Create a new workflow
 * @returns Mutation for creating workflow
 */
export function useCreateWorkflow() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (formData: WorkflowFormData) => createWorkflow(formData),
    onSuccess: () => {
      // Invalidate workflows list
      queryClient.invalidateQueries({
        queryKey: [WORKFLOWS_QUERY_KEY],
      });
    },
    onError: (error) => {
      console.error('Failed to create workflow:', error);
    },
  });
}

/**
 * Hook: Update an existing workflow
 * @returns Mutation for updating workflow
 */
export function useUpdateWorkflow() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ workflowId, formData }: { workflowId: string; formData: Partial<WorkflowFormData> }) =>
      updateWorkflow(workflowId, formData),
    onSuccess: (data, variables) => {
      // Invalidate workflows list
      queryClient.invalidateQueries({
        queryKey: [WORKFLOWS_QUERY_KEY],
      });

      // Invalidate specific workflow
      queryClient.invalidateQueries({
        queryKey: [WORKFLOW_QUERY_KEY, variables.workflowId],
      });
    },
    onError: (error) => {
      console.error('Failed to update workflow:', error);
    },
  });
}

/**
 * Hook: Delete a workflow
 * @returns Mutation for deleting workflow
 */
export function useDeleteWorkflow() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (workflowId: string) => deleteWorkflow(workflowId),
    onSuccess: (data, workflowId) => {
      // Invalidate workflows list
      queryClient.invalidateQueries({
        queryKey: [WORKFLOWS_QUERY_KEY],
      });

      // Remove specific workflow from cache
      queryClient.removeQueries({
        queryKey: [WORKFLOW_QUERY_KEY, workflowId],
      });
    },
    onError: (error) => {
      console.error('Failed to delete workflow:', error);
    },
  });
}

/**
 * Hook: Activate a workflow
 * @returns Mutation for activating workflow
 */
export function useActivateWorkflow() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (workflowId: string) => activateWorkflow(workflowId),
    onSuccess: (data, workflowId) => {
      // Invalidate workflows list
      queryClient.invalidateQueries({
        queryKey: [WORKFLOWS_QUERY_KEY],
      });

      // Invalidate specific workflow
      queryClient.invalidateQueries({
        queryKey: [WORKFLOW_QUERY_KEY, workflowId],
      });
    },
    onError: (error) => {
      console.error('Failed to activate workflow:', error);
    },
  });
}

/**
 * Hook: Deactivate a workflow
 * @returns Mutation for deactivating workflow
 */
export function useDeactivateWorkflow() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (workflowId: string) => deactivateWorkflow(workflowId),
    onSuccess: (data, workflowId) => {
      // Invalidate workflows list
      queryClient.invalidateQueries({
        queryKey: [WORKFLOWS_QUERY_KEY],
      });

      // Invalidate specific workflow
      queryClient.invalidateQueries({
        queryKey: [WORKFLOW_QUERY_KEY, workflowId],
      });
    },
    onError: (error) => {
      console.error('Failed to deactivate workflow:', error);
    },
  });
}

/**
 * Hook: Set default workflow for entity type
 * @returns Mutation for setting default workflow
 */
export function useSetDefaultWorkflow() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ workflowId, entityType }: { workflowId: string; entityType: string }) =>
      setDefaultWorkflow(workflowId, entityType),
    onSuccess: (data, variables) => {
      // Invalidate default workflows
      queryClient.invalidateQueries({
        queryKey: [DEFAULT_WORKFLOWS_QUERY_KEY, variables.entityType],
      });

      // Invalidate all default workflows
      queryClient.invalidateQueries({
        queryKey: [DEFAULT_WORKFLOWS_QUERY_KEY],
      });
    },
    onError: (error) => {
      console.error('Failed to set default workflow:', error);
    },
  });
}

/**
 * Hook: Invalidate pending approvals cache
 * Utility hook for other hooks to invalidate pending approvals
 * @returns Function to invalidate pending approvals
 */
export function useInvalidatePendingApprovals() {
  const queryClient = useQueryClient();

  return (userId?: string) => {
    if (userId) {
      queryClient.invalidateQueries({
        queryKey: ['pending-approvals', userId],
      });
    } else {
      queryClient.invalidateQueries({
        queryKey: ['pending-approvals'],
      });
    }
  };
}