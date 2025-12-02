'use client';

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { QUERY_KEYS } from '@/lib/constants';
import {
  getBudgets,
  getBudgetById,
  createBudget,
  submitBudgetForApproval,
  approveBudget,
  rejectBudget,
} from '@/app/_actions/budgets';
import { Budget, CreateBudgetRequest, ApproveBudgetRequest, RejectBudgetRequest } from '@/types/budget';
import { toast } from 'sonner';

/**
 * Fetch all budgets for a user
 * Static data - rarely changes
 *
 * @param userId - User ID to fetch budgets for
 * @param initialBudgets - Optional initial data from server component
 * @returns Query result with budgets array
 *
 * @example
 * const { data: budgets, isLoading } = useBudgets(userId, initialBudgets)
 */
export const useBudgets = (userId: string, initialBudgets?: Budget[]) =>
  useQuery({
    queryKey: [QUERY_KEYS.BUDGETS.BY_USER, userId],
    queryFn: async () => {
      const response = await getBudgets(userId);
      return response.success ? response.data : [];
    },
    initialData: initialBudgets,
    staleTime: Infinity, // Static data
  });

/**
 * Fetch a specific budget by ID
 *
 * @param budgetId - Budget ID to fetch
 * @returns Query result with single budget
 *
 * @example
 * const { data: budget } = useBudgetById(budgetId)
 */
export const useBudgetById = (budgetId: string) =>
  useQuery({
    queryKey: [QUERY_KEYS.BUDGETS.BY_ID, budgetId],
    queryFn: async () => {
      const response = await getBudgetById(budgetId);
      if (!response.success) throw new Error(response.message);
      return response.data;
    },
    staleTime: 5 * 60 * 1000, // 5 minutes
  });

/**
 * Create budget mutation
 *
 * @param onSuccess - Callback after successful mutation
 * @returns Mutation object with mutate and mutateAsync
 *
 * @example
 * const saveMutation = useSaveBudget(() => {
 *   queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.BUDGETS.BY_USER] })
 * })
 *
 * // Create
 * await saveMutation.mutateAsync({ name: 'Q1 Budget' })
 */
export const useSaveBudget = (onSuccess?: () => void) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: CreateBudgetRequest) => {
      const response = await createBudget(data);

      if (!response.success) {
        throw new Error(response.message);
      }
      return response;
    },
    onSuccess: (response) => {
      const isUpdate = (response.data as Budget & { id?: string })?.id;
      toast.success(isUpdate ? 'Budget updated successfully' : 'Budget created successfully');

      // Invalidate budget queries
      queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.BUDGETS.BY_USER] });
      queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.BUDGETS.ALL] });
      queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.BUDGETS.STATS] });

      onSuccess?.();
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to save budget');
    },
  });
};

/**
 * Submit budget for approval mutation
 *
 * @param budgetId - Budget ID to submit
 * @param onSuccess - Callback after successful submission
 * @returns Mutation object
 *
 * @example
 * const submitMutation = useSubmitBudgetForApproval(budgetId)
 * await submitMutation.mutateAsync({ submittedBy: userId, comments: 'Please review' })
 */
export const useSubmitBudgetForApproval = (budgetId: string, onSuccess?: () => void) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: { submittedBy: string; comments?: string }) => {
      const response = await submitBudgetForApproval({
        budgetId,
        ...data,
      });

      if (!response.success) {
        throw new Error(response.message);
      }
      return response;
    },
    onSuccess: () => {
      toast.success('Budget submitted for approval');

      queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.BUDGETS.BY_ID, budgetId] });
      queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.BUDGETS.BY_USER] });

      onSuccess?.();
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to submit budget');
    },
  });
};

/**
 * Approve budget mutation
 *
 * @param budgetId - Budget ID to approve
 * @param onSuccess - Callback after successful approval
 * @returns Mutation object
 *
 * @example
 * const approveMutation = useApproveBudget(budgetId)
 * await approveMutation.mutateAsync({
 *   approvingUserId: userId,
 *   approvingUserRole: 'FINANCE_OFFICER',
 *   signature: signatureDataUrl,
 *   comments: 'Approved'
 * })
 */
export const useApproveBudget = (budgetId: string, onSuccess?: () => void) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: Omit<ApproveBudgetRequest, 'budgetId'>) => {
      const response = await approveBudget({
        budgetId,
        ...data,
      });

      if (!response.success) {
        throw new Error(response.message);
      }
      return response;
    },
    onSuccess: () => {
      toast.success('Budget approved successfully');

      queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.BUDGETS.BY_ID, budgetId] });
      queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.BUDGETS.BY_USER] });
      queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.APPROVALS_PENDING] });

      onSuccess?.();
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to approve budget');
    },
  });
};

/**
 * Reject budget mutation
 *
 * @param budgetId - Budget ID to reject
 * @param onSuccess - Callback after successful rejection
 * @returns Mutation object
 *
 * @example
 * const rejectMutation = useRejectBudget(budgetId)
 * await rejectMutation.mutateAsync({
 *   rejectingUserId: userId,
 *   rejectingUserRole: 'FINANCE_OFFICER',
 *   remarks: 'Insufficient details',
 *   signature: signatureDataUrl
 * })
 */
export const useRejectBudget = (budgetId: string, onSuccess?: () => void) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: Omit<RejectBudgetRequest, 'budgetId'>) => {
      const response = await rejectBudget({
        budgetId,
        ...data,
      });

      if (!response.success) {
        throw new Error(response.message);
      }
      return response;
    },
    onSuccess: () => {
      toast.success('Budget rejected and returned to draft');

      queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.BUDGETS.BY_ID, budgetId] });
      queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.BUDGETS.BY_USER] });
      queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.APPROVALS_PENDING] });

      onSuccess?.();
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to reject budget');
    },
  });
};
