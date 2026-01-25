"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { QUERY_KEYS } from "@/lib/constants";
import {
  getBudgets,
  getBudgetById,
  createBudget,
  updateBudget,
  submitBudgetForApproval,
} from "@/app/_actions/budgets";
import { Budget, CreateBudgetRequest } from "@/types/budget";
import { useBudgetStorage } from "@/hooks/use-budget-storage";
import { toast } from "sonner";

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
      const response = await getBudgets({ userId });
      return response.success ? response.data : [];
    },
    initialData: initialBudgets,
    staleTime: Infinity, // Static data
  });

/**
 * Fetch all budgets (for dropdowns and selection)
 * Static data - rarely changes
 *
 * @param initialBudgets - Optional initial data from server component
 * @returns Query result with budgets array
 *
 * @example
 * const { data: budgets, isLoading } = useAllBudgets()
 */
export const useAllBudgets = (initialBudgets?: Budget[]) =>
  useQuery({
    queryKey: [QUERY_KEYS.BUDGETS.ALL],
    queryFn: async () => {
      try {
        const response = await getBudgets({}, 1, 100); // Get first 100 budgets
        console.log("Budget response:", response); // Debug log
        return response.success && Array.isArray(response.data)
          ? response.data
          : [];
      } catch (error) {
        console.error("Error fetching budgets:", error);
        return [];
      }
    },
    initialData: initialBudgets || [],
    staleTime: 5 * 60 * 1000, // 5 minutes
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
      toast.success(
        isUpdate
          ? "Budget updated successfully"
          : "Budget created successfully",
      );

      // Invalidate budget queries
      queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.BUDGETS.BY_USER] });
      queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.BUDGETS.ALL] });
      queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.BUDGETS.STATS] });

      onSuccess?.();
    },
    onError: (error: Error) => {
      toast.error(error.message || "Failed to save budget");
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
 * await submitMutation.mutateAsync({ submittingUserId: userId })
 */
export const useSubmitBudgetForApproval = (
  budgetId: string,
  onSuccess?: () => void,
) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: { submittingUserId: string }) => {
      const response = await submitBudgetForApproval({
        budgetId,
        submittedBy: data.submittingUserId,
        submittedByRole: "requester",
        submittingUserId: data.submittingUserId,
      });

      if (!response.success) {
        throw new Error(response.message);
      }
      return response;
    },
    onSuccess: () => {
      toast.success("Budget submitted for approval");

      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.BUDGETS.BY_ID, budgetId],
      });
      queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.BUDGETS.BY_USER] });

      onSuccess?.();
    },
    onError: (error: Error) => {
      toast.error(error.message || "Failed to submit budget");
    },
  });
};

/**
 * Update budget mutation (for items, metadata, etc.)
 *
 * @param budgetId - Budget ID to update
 * @param onSuccess - Callback after successful update
 * @returns Mutation object
 *
 * @example
 * const updateMutation = useUpdateBudget(budgetId)
 * await updateMutation.mutateAsync({
 *   items: [...updatedItems],
 *   name: 'Updated budget name'
 * })
 */
export const useUpdateBudget = (budgetId: string, onSuccess?: () => void) => {
  const queryClient = useQueryClient();
  const { saveToStorage } = useBudgetStorage();

  return useMutation({
    mutationFn: async (updates: Partial<Budget>) => {
      const response = await updateBudget(budgetId, updates);

      if (!response.success) {
        throw new Error(response.message);
      }
      return response;
    },
    onSuccess: (response) => {
      // Save to localStorage
      if (response.data) {
        saveToStorage(response.data);
      }

      toast.success("Budget updated successfully");

      // Invalidate budget queries
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.BUDGETS.BY_ID, budgetId],
      });
      queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.BUDGETS.BY_USER] });
      queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.BUDGETS.ALL] });

      onSuccess?.();
    },
    onError: (error: Error) => {
      toast.error(error.message || "Failed to update budget");
    },
  });
};
