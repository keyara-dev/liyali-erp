"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { QUERY_KEYS } from "@/lib/constants";
import { queryKeys } from "@/lib/query-keys";
import {
  getRequisitions,
  getRequisitionById,
  createRequisition,
  updateRequisition,
  submitRequisitionForApproval,
  approveRequisition,
  rejectRequisition,
  deleteRequisition,
  getRequisitionStats,
} from "@/app/_actions/requisitions";
import {
  Requisition,
  RequisitionStats,
  CreateRequisitionRequest,
  UpdateRequisitionRequest,
  SubmitRequisitionRequest,
  ApproveRequisitionRequest,
  RejectRequisitionRequest,
} from "@/types/requisition";
import { toast } from "sonner";
import { handleOfflineMutation, isOfflineResult } from "@/lib/offline-mutation-helper";

/**
 * Fetch all requisitions with pagination
 * Standard data - 5 minute refresh interval
 *
 * @param page - Page number (default: 1)
 * @param limit - Items per page (default: 10)
 * @param filters - Optional filters (status, department)
 * @returns Query result with requisitions array
 *
 * @example
 * const { data: requisitions } = useRequisitions(1, 10, { status: 'DRAFT' })
 */
export const useRequisitions = (
  page: number = 1,
  limit: number = 10,
  filters?: {
    status?: string;
    department?: string;
  }
) =>
  useQuery({
    queryKey: [QUERY_KEYS.REQUISITIONS.ALL, page, limit, filters],
    queryFn: async () => {
      const response = await getRequisitions(page, limit, filters);
      return response.success ? response.data : [];
    },
    staleTime: 5 * 60 * 1000, // 5 minutes
  });

/**
 * Fetch a specific requisition by ID
 *
 * @param requisitionId - Requisition ID to fetch
 * @param initialData - Optional initial data
 * @returns Query result with single requisition
 *
 * @example
 * const { data: requisition } = useRequisitionById(requisitionId)
 */
export const useRequisitionById = (
  requisitionId: string,
  initialData?: Requisition
) =>
  useQuery({
    queryKey: [QUERY_KEYS.REQUISITIONS.BY_ID, requisitionId],
    queryFn: async () => {
      const response = await getRequisitionById(requisitionId);
      if (!response.success) throw new Error(response.message);
      return response.data;
    },
    initialData,
    staleTime: 5 * 60 * 1000, // 5 minutes
    enabled: !!requisitionId,
  });

/**
 * Fetch requisition statistics
 *
 * @param initialStats - Optional initial data from server component
 * @returns Query result with requisition statistics
 *
 * @example
 * const { data: stats } = useRequisitionStats()
 */
export const useRequisitionStats = (initialStats?: RequisitionStats) =>
  useQuery({
    queryKey: [QUERY_KEYS.REQUISITIONS.STATS],
    queryFn: async () => {
      const response = await getRequisitionStats();
      return response.success ? response.data : null;
    },
    initialData: initialStats,
    staleTime: 10 * 60 * 1000, // 10 minutes
  });

/**
 * Create or update requisition mutation
 * Handles both create (no ID) and update (with ID) operations
 * Only DRAFT requisitions can be updated
 *
 * @param onSuccess - Callback after successful mutation
 * @returns Mutation object with mutate and mutateAsync
 *
 * @example
 * const saveMutation = useSaveRequisition(() => {
 *   queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.REQUISITIONS.ALL] })
 * })
 *
 * // Create
 * await saveMutation.mutateAsync({
 *   title: 'Office Supplies',
 *   department: 'Admin',
 *   items: [...],
 *   createdBy: userId
 * })
 *
 * // Update
 * await saveMutation.mutateAsync({
 *   requisitionId: 'req-1',
 *   title: 'Office Supplies Updated',
 *   items: [...]
 * })
 */
export const useSaveRequisition = (onSuccess?: () => void) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (
      data:
        | CreateRequisitionRequest
        | (UpdateRequisitionRequest & { requisitionId?: string })
    ) => {
      const isUpdate = "requisitionId" in data && data.requisitionId;
      
      return await handleOfflineMutation(
        async () => {
          const response = isUpdate
            ? await updateRequisition(data as UpdateRequisitionRequest)
            : await createRequisition(data as CreateRequisitionRequest);

          if (!response.success) {
            throw new Error(response.message);
          }
          return response;
        },
        {
          operation: isUpdate ? 'UPDATE' : 'CREATE',
          entity: 'requisition',
          data,
          entityId: isUpdate ? (data as UpdateRequisitionRequest).requisitionId : undefined,
          successMessage: isUpdate ? 'Requisition updated successfully' : 'Requisition created successfully',
          offlineMessage: isUpdate 
            ? 'Requisition changes saved offline. Will sync when connected.'
            : 'Requisition saved offline. Will sync when connected.',
        }
      );
    },
    onSuccess: (result) => {
      if (isOfflineResult(result)) {
        // Already handled by offline helper
      } else {
        const isUpdate = (
          result.data as Requisition & { requisitionId?: string }
        )?.requisitionId;
        toast.success(
          isUpdate
            ? "Requisition updated successfully"
            : "Requisition created successfully"
        );
      }

      // Invalidate requisition queries
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.REQUISITIONS.ALL],
      });
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.REQUISITIONS.STATS],
      });

      // Invalidate dashboard metrics
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.DASHBOARD.METRICS],
      });
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.DASHBOARD.ACTIVITIES],
      });

      onSuccess?.();
    },
    onError: (error: Error) => {
      toast.error(error.message || "Failed to save requisition");
    },
  });
};

/**
 * Submit requisition for approval mutation
 *
 * @param requisitionId - Requisition ID to submit
 * @param onSuccess - Callback after successful submission
 * @returns Mutation object
 *
 * @example
 * const submitMutation = useSubmitRequisitionForApproval(requisitionId)
 * await submitMutation.mutateAsync({
 *   submittedBy: userId,
 *   submittedByName: 'John Doe',
 *   submittedByRole: 'REQUESTER',
 *   comments: 'Please review'
 * })
 */
export const useSubmitRequisitionForApproval = (
  requisitionId: string,
  onSuccess?: () => void
) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (
      data: Omit<SubmitRequisitionRequest, "requisitionId">
    ) => {
      return await handleOfflineMutation(
        async () => {
          const response = await submitRequisitionForApproval({
            requisitionId,
            ...data,
          });

          if (!response.success) {
            throw new Error(response.message);
          }
          return response;
        },
        {
          operation: 'SUBMIT',
          entity: 'requisition',
          data: { requisitionId, ...data },
          entityId: requisitionId,
          successMessage: 'Requisition submitted for approval',
          offlineMessage: 'Requisition submission saved offline. Will sync when connected.',
        }
      );
    },
    onSuccess: (result) => {
      if (isOfflineResult(result)) {
        // Already handled by offline helper
      } else {
        toast.success("Requisition submitted for approval");
      }

      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.REQUISITIONS.BY_ID, requisitionId],
      });
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.REQUISITIONS.ALL],
      });
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.REQUISITIONS.STATS],
      });

      // Invalidate dashboard metrics
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.DASHBOARD.METRICS],
      });
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.DASHBOARD.ACTIVITIES],
      });

      onSuccess?.();
    },
    onError: (error: Error) => {
      toast.error(error.message || "Failed to submit requisition");
    },
  });
};

/**
 * Approve requisition mutation
 *
 * @param requisitionId - Requisition ID to approve
 * @param onSuccess - Callback after successful approval
 * @returns Mutation object
 *
 * @example
 * const approveMutation = useApproveRequisition(requisitionId)
 * await approveMutation.mutateAsync({
 *   approvingUserId: userId,
 *   approvingUserName: 'John Doe',
 *   approvingUserRole: 'DEPARTMENT_MANAGER',
 *   signature: signatureDataUrl,
 *   comments: 'Approved'
 * })
 */
export const useApproveRequisition = (
  requisitionId: string,
  onSuccess?: (data: any) => void
) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (
      data: Omit<ApproveRequisitionRequest, "requisitionId">
    ) => {
      const response = await approveRequisition({
        requisitionId,
        ...data,
      });

      if (!response.success) {
        throw new Error(response.message);
      }
      return response;
    },
    onSuccess: (response) => {
      const data = response.data;
      
      // Check if automation was used
      if (data?.automationUsed && data?.autoCreatedPO) {
        toast.success("Requisition approved and Purchase Order created automatically");
        
        // Invalidate PO queries since a new PO was created
        queryClient.invalidateQueries({
          queryKey: [QUERY_KEYS.PURCHASE_ORDERS.ALL],
        });
        queryClient.invalidateQueries({
          queryKey: [QUERY_KEYS.PURCHASE_ORDERS.STATS],
        });
        
        // Optionally set the new PO in cache
        if (data.autoCreatedPO.id) {
          queryClient.setQueryData(
            [QUERY_KEYS.PURCHASE_ORDERS.BY_ID, data.autoCreatedPO.id],
            data.autoCreatedPO
          );
        }
      } else {
        toast.success("Requisition approved successfully");
      }

      // Standard requisition cache invalidation
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.REQUISITIONS.BY_ID, requisitionId],
      });
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.REQUISITIONS.ALL],
      });
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.REQUISITIONS.STATS],
      });
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.APPROVALS_PENDING],
      });

      // Invalidate dashboard metrics
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.DASHBOARD.METRICS],
      });
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.DASHBOARD.ACTIVITIES],
      });

      onSuccess?.(response.data);
    },
    onError: (error: Error) => {
      toast.error(error.message || "Failed to approve requisition");
    },
  });
};

/**
 * Reject requisition mutation
 *
 * @param requisitionId - Requisition ID to reject
 * @param onSuccess - Callback after successful rejection
 * @returns Mutation object
 *
 * @example
 * const rejectMutation = useRejectRequisition(requisitionId)
 * await rejectMutation.mutateAsync({
 *   rejectingUserId: userId,
 *   rejectingUserName: 'John Doe',
 *   rejectingUserRole: 'DEPARTMENT_MANAGER',
 *   remarks: 'Missing required information',
 *   signature: signatureDataUrl
 * })
 */
export const useRejectRequisition = (
  requisitionId: string,
  onSuccess?: () => void
) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (
      data: Omit<RejectRequisitionRequest, "requisitionId">
    ) => {
      const response = await rejectRequisition({
        requisitionId,
        ...data,
      });

      if (!response.success) {
        throw new Error(response.message);
      }
      return response;
    },
    onSuccess: () => {
      toast.success("Requisition rejected and returned to draft");

      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.REQUISITIONS.BY_ID, requisitionId],
      });
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.REQUISITIONS.ALL],
      });
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.REQUISITIONS.STATS],
      });
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.APPROVALS_PENDING],
      });

      // Invalidate dashboard metrics
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.DASHBOARD.METRICS],
      });
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.DASHBOARD.ACTIVITIES],
      });

      onSuccess?.();
    },
    onError: (error: Error) => {
      toast.error(error.message || "Failed to reject requisition");
    },
  });
};

/**
 * Delete requisition mutation
 * Only DRAFT requisitions can be deleted
 *
 * @param requisitionId - Requisition ID to delete
 * @param onSuccess - Callback after successful deletion
 * @returns Mutation object
 *
 * @example
 * const deleteMutation = useDeleteRequisition(requisitionId)
 * await deleteMutation.mutateAsync()
 */
export const useDeleteRequisition = (
  requisitionId: string,
  onSuccess?: () => void
) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async () => {
      const response = await deleteRequisition(requisitionId);

      if (!response.success) {
        throw new Error(response.message);
      }
      return response;
    },
    onSuccess: () => {
      toast.success("Requisition deleted successfully");

      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.REQUISITIONS.ALL],
      });
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.REQUISITIONS.STATS],
      });

      // Invalidate dashboard metrics
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.DASHBOARD.METRICS],
      });
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.DASHBOARD.ACTIVITIES],
      });

      onSuccess?.();
    },
    onError: (error: Error) => {
      toast.error(error.message || "Failed to delete requisition");
    },
  });
};
