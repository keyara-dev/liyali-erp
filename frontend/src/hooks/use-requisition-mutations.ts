"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import {
  createRequisition,
  updateRequisition,
} from "@/app/_actions/requisitions";
import {
  CreateRequisitionRequest,
  UpdateRequisitionRequest,
} from "@/types/requisition";
import { QUERY_KEYS } from "@/lib/constants";

/**
 * Hook for creating a new requisition
 */
export const useCreateRequisition = (onSuccess?: () => void) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: CreateRequisitionRequest) => {
      const result = await createRequisition(data);
      if (!result.success) {
        throw new Error(result.message || "Failed to create requisition");
      }
      return result;
    },
    onSuccess: (result) => {
      toast.success(
        `Requisition ${result.data?.documentNumber} created successfully`
      );

      // Invalidate relevant queries
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.REQUISITIONS.ALL],
      });
      queryClient.invalidateQueries({
        queryKey: ["requisitions"],
      });

      onSuccess?.();
    },
    onError: (error: Error) => {
      toast.error(error.message || "Failed to create requisition");
    },
  });
};

/**
 * Hook for updating an existing requisition
 */
export const useUpdateRequisition = (onSuccess?: () => void) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: UpdateRequisitionRequest) => {
      const result = await updateRequisition(data);
      if (!result.success) {
        throw new Error(result.message || "Failed to update requisition");
      }
      return result;
    },
    onSuccess: (result) => {
      toast.success("Requisition updated successfully");

      // Invalidate relevant queries
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.REQUISITIONS.ALL],
      });
      queryClient.invalidateQueries({
        queryKey: ["requisitions"],
      });

      // Invalidate specific requisition if we have the ID
      if (result.data?.id) {
        queryClient.invalidateQueries({
          queryKey: [QUERY_KEYS.REQUISITIONS.BY_ID, result.data.id],
        });
      }

      onSuccess?.();
    },
    onError: (error: Error) => {
      toast.error(error.message || "Failed to update requisition");
    },
  });
};

/**
 * Hook for submitting a requisition for approval
 */
export const useSubmitRequisition = (onSuccess?: () => void) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: {
      requisitionId: string;
      submittedBy: string;
      submittedByName: string;
      submittedByRole: string;
      comments?: string;
    }) => {
      // This would call a submitRequisition action when available
      // For now, we'll update the status to 'submitted'
      const result = await updateRequisition({
        requisitionId: data.requisitionId,
        // Add status update when the backend supports it
      });
      if (!result.success) {
        throw new Error(result.message || "Failed to submit requisition");
      }
      return result;
    },
    onSuccess: (result) => {
      toast.success("Requisition submitted for approval");

      // Invalidate relevant queries
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.REQUISITIONS.ALL],
      });
      queryClient.invalidateQueries({
        queryKey: ["requisitions"],
      });

      onSuccess?.();
    },
    onError: (error: Error) => {
      toast.error(error.message || "Failed to submit requisition");
    },
  });
};
