'use client';

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { QUERY_KEYS } from '@/lib/constants';
import { toast } from 'sonner';

export interface WorkflowStage {
  id: string;
  order: number;
  name: string;
  description: string;
  approverRole: string;
  requiredApprovals: number;
  canReject: boolean;
  canReassign: boolean;
}

export interface WorkflowFormData {
  name: string;
  description: string;
  documentType: string;
  stages: WorkflowStage[];
  isDefault: boolean;
}

export interface Workflow {
  id: string;
  name: string;
  description: string;
  documentType: string;
  stages: number;
  status: 'ACTIVE' | 'DEPRECATED';
  createdAt: string;
  updatedAt: string;
  createdBy: string;
}

/**
 * Fetch all workflows
 * @param onSuccess - Optional callback on success
 * @returns Query result with workflows array
 */
export const useWorkflows = (onSuccess?: (data: Workflow[]) => void) =>
  useQuery({
    queryKey: [QUERY_KEYS.WORKFLOWS.ALL],
    queryFn: async () => {
      const response = await fetch('/api/workflows');
      if (!response.ok) throw new Error('Failed to fetch workflows');
      return response.json();
    },
    staleTime: 5 * 60 * 1000, // 5 minutes
    onSuccess,
  });

/**
 * Fetch a specific workflow by ID
 * @param workflowId - Workflow ID to fetch
 * @returns Query result with workflow details
 */
export const useWorkflowById = (workflowId: string) =>
  useQuery({
    queryKey: [QUERY_KEYS.WORKFLOWS.DETAIL, workflowId],
    queryFn: async () => {
      const response = await fetch(`/api/workflows/${workflowId}`);
      if (!response.ok) throw new Error('Failed to fetch workflow');
      return response.json();
    },
    enabled: !!workflowId,
    staleTime: 5 * 60 * 1000, // 5 minutes
  });

/**
 * Create a new workflow
 * @param onSuccess - Optional callback on success
 * @returns Mutation object with mutateAsync, isPending, error
 */
export const useCreateWorkflow = (onSuccess?: (data: Workflow) => void) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (formData: WorkflowFormData) => {
      const response = await fetch('/api/workflows', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(formData),
      });
      if (!response.ok) throw new Error('Failed to create workflow');
      return response.json();
    },
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.WORKFLOWS.ALL] });
      toast.success('Workflow created successfully');
      if (onSuccess) onSuccess(data);
    },
    onError: (error: any) => {
      toast.error(error.message || 'Failed to create workflow');
    },
  });
};

/**
 * Update an existing workflow
 * @param workflowId - Workflow ID to update
 * @param onSuccess - Optional callback on success
 * @returns Mutation object with mutateAsync, isPending, error
 */
export const useUpdateWorkflow = (
  workflowId: string,
  onSuccess?: (data: Workflow) => void
) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (formData: WorkflowFormData) => {
      const response = await fetch(`/api/workflows/${workflowId}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(formData),
      });
      if (!response.ok) throw new Error('Failed to update workflow');
      return response.json();
    },
    onSuccess: (data) => {
      queryClient.setQueryData(
        [QUERY_KEYS.WORKFLOWS.DETAIL, workflowId],
        data
      );
      queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.WORKFLOWS.ALL] });
      toast.success('Workflow updated successfully');
      if (onSuccess) onSuccess(data);
    },
    onError: (error: any) => {
      toast.error(error.message || 'Failed to update workflow');
    },
  });
};

/**
 * Delete a workflow
 * @param onSuccess - Optional callback on success
 * @returns Mutation object with mutateAsync, isPending, error
 */
export const useDeleteWorkflow = (onSuccess?: () => void) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (workflowId: string) => {
      const response = await fetch(`/api/workflows/${workflowId}`, {
        method: 'DELETE',
      });
      if (!response.ok) throw new Error('Failed to delete workflow');
      return response.json();
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.WORKFLOWS.ALL] });
      toast.success('Workflow deleted successfully');
      if (onSuccess) onSuccess();
    },
    onError: (error: any) => {
      toast.error(error.message || 'Failed to delete workflow');
    },
  });
};

/**
 * Duplicate a workflow
 * @param onSuccess - Optional callback on success
 * @returns Mutation object with mutateAsync, isPending, error
 */
export const useDuplicateWorkflow = (onSuccess?: (data: Workflow) => void) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (workflowId: string) => {
      const response = await fetch(`/api/workflows/${workflowId}/duplicate`, {
        method: 'POST',
      });
      if (!response.ok) throw new Error('Failed to duplicate workflow');
      return response.json();
    },
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.WORKFLOWS.ALL] });
      toast.success('Workflow duplicated successfully');
      if (onSuccess) onSuccess(data);
    },
    onError: (error: any) => {
      toast.error(error.message || 'Failed to duplicate workflow');
    },
  });
};
