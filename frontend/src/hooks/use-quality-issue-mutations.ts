"use client";

/**
 * Quality Issue Mutations
 * React Query mutations for GRN quality issue operations
 */

import { useMutation, useQueryClient } from "@tanstack/react-query";
import {
  addQualityIssueToGRN,
  removeQualityIssueFromGRN,
  updateQualityIssueInGRN,
} from "@/app/_actions/grn-actions";
import type { QualityIssue } from "@/types/goods-received-note";
import { QUERY_KEYS } from "@/lib/constants";

/**
 * Mutation hook for adding a quality issue to a GRN
 * Automatically invalidates related queries and updates local state
 */
export function useAddQualityIssueMutation(grnId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (issue: Omit<QualityIssue, "id">) => {
      const result = await addQualityIssueToGRN(grnId, issue);
      return result;
    },
    onSuccess: () => {
      // Invalidate the detail + list caches so fresh data is fetched.
      // We don't setQueryData directly because the mutation result is wrapped
      // in an APIResponse, which would clobber the detail cache with the
      // wrong shape.
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.GRN.BY_ID, grnId],
      });
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.GRN.ALL],
      });
    },
    onError: (error: Error) => {
      console.error("Failed to add quality issue:", error);
    },
  });
}

/**
 * Mutation hook for removing a quality issue from a GRN
 */
export function useRemoveQualityIssueMutation(grnId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (issueId: string) => {
      const result = await removeQualityIssueFromGRN(grnId, issueId);
      return result;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.GRN.BY_ID, grnId],
      });
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.GRN.ALL],
      });
    },
    onError: (error: Error) => {
      console.error("Failed to remove quality issue:", error);
    },
  });
}

/**
 * Mutation hook for updating a quality issue in a GRN
 */
export function useUpdateQualityIssueMutation(grnId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({
      issueId,
      updates,
    }: {
      issueId: string;
      updates: Partial<Omit<QualityIssue, "id">>;
    }) => {
      const result = await updateQualityIssueInGRN(grnId, issueId, updates);
      return result;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.GRN.BY_ID, grnId],
      });
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.GRN.ALL],
      });
    },
    onError: (error: Error) => {
      console.error("Failed to update quality issue:", error);
    },
  });
}
