'use client';

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { QUERY_KEYS } from '@/lib/constants';
import {
  getGRNAction,
  getGRNsAction,
  createGRNAction,
  updateGRNAction,
  approveGRNAction,
  rejectGRNAction,
  deleteGRNAction,
  confirmGRNAction,
  GoodsReceivedNote,
} from '@/app/_actions/grn-actions';
import { toast } from 'sonner';
import { APIResponse } from '@/types';

export type { GoodsReceivedNote };

/**
 * Fetch all GRNs with pagination
 * Standard data - 5 minute refresh interval
 *
 * @param page - Page number (default: 1)
 * @param limit - Items per page (default: 10)
 * @param filters - Optional filters (status, poNumber)
 * @returns Query result with GRNs array
 *
 * @example
 * const { data: grns } = useGRNs(1, 10, { status: 'DRAFT' })
 */
export const useGRNs = (
  page: number = 1,
  limit: number = 10,
  filters?: {
    status?: string;
    poNumber?: string;
  }
) =>
  useQuery({
    queryKey: [QUERY_KEYS.GRN.ALL, page, limit, filters],
    queryFn: async () => {
      const response = await getGRNsAction(page, limit, filters);
      return response.success ? response.data : [];
    },
    staleTime: 5 * 60 * 1000, // 5 minutes
  });

/**
 * Fetch a specific GRN by ID
 *
 * @param grnId - GRN ID to fetch
 * @param initialData - Optional initial data
 * @returns Query result with single GRN
 *
 * @example
 * const { data: grn } = useGRNById(grnId)
 */
export const useGRNById = (grnId: string, initialData?: GoodsReceivedNote) =>
  useQuery({
    queryKey: [QUERY_KEYS.GRN.BY_ID, grnId],
    queryFn: async () => {
      const response = await getGRNAction(grnId);
      if (!response.success) throw new Error(response.message);
      return response.data;
    },
    initialData,
    staleTime: 5 * 60 * 1000, // 5 minutes
    enabled: !!grnId,
  });

/**
 * Create a new GRN from a Purchase Order
 * Automatically invalidates GRN list queries
 *
 * @param onSuccess - Optional callback after successful creation
 * @returns Mutation object with mutateAsync, isPending, error
 *
 * @example
 * const { mutateAsync: createGRN, isPending } = useCreateGRN();
 * await createGRN({
 *   poNumber: 'PO-123',
 *   items: [...],
 *   receivedBy: 'user-id'
 * });
 */
export const useCreateGRN = (onSuccess?: (data: GoodsReceivedNote) => void) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({
      poNumber,
      items,
      receivedBy,
      warehouseLocation,
      notes,
    }: {
      poNumber: string;
      items: any[];
      receivedBy: string;
      warehouseLocation?: string;
      notes?: string;
    }) => {
      const response = await createGRNAction(
        poNumber,
        items,
        receivedBy,
        warehouseLocation,
        notes
      );
      if (!response.success) {
        throw new Error(response.message);
      }
      return response;
    },
    onSuccess: (response) => {
      toast.success('GRN created successfully');
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.GRN.ALL],
      });

      if (onSuccess && response.data) {
        onSuccess(response.data);
      }
    },
    onError: (error: any) => {
      toast.error(error.message || 'Failed to create GRN');
    },
  });
};

/**
 * Update an existing GRN
 * Automatically invalidates GRN queries
 *
 * @param grnId - GRN ID to update
 * @param onSuccess - Optional callback after successful update
 * @returns Mutation object with mutateAsync, isPending, error
 *
 * @example
 * const { mutateAsync: updateGRN } = useUpdateGRN(grnId);
 * await updateGRN({
 *   items: [...],
 *   qualityIssues: [...]
 * });
 */
export const useUpdateGRN = (grnId: string, onSuccess?: (data: GoodsReceivedNote) => void) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (updates: {
      items?: any[];
      receivedBy?: string;
      qualityIssues?: any[];
      warehouseLocation?: string;
      notes?: string;
    }) => {
      const response = await updateGRNAction(grnId, updates);
      if (!response.success) {
        throw new Error(response.message);
      }
      return response;
    },
    onSuccess: (response) => {
      toast.success('GRN updated successfully');

      // Update the specific GRN in cache
      queryClient.setQueryData([QUERY_KEYS.GRN.BY_ID, grnId], response.data);

      // Invalidate list queries
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.GRN.ALL],
      });

      if (onSuccess && response.data) {
        onSuccess(response.data);
      }
    },
    onError: (error: any) => {
      toast.error(error.message || 'Failed to update GRN');
    },
  });
};

/**
 * Approve a GRN
 * Handles automatic Payment Voucher creation when GRN is approved
 * Automatically invalidates GRN queries
 *
 * @param grnId - GRN ID to approve
 * @param onSuccess - Optional callback after successful approval
 * @returns Mutation object with mutateAsync, isPending, error
 *
 * @example
 * const { mutateAsync: approveGRN } = useApproveGRN(grnId);
 * await approveGRN({
 *   signature: 'user-signature',
 *   comments: 'Looks good'
 * });
 */
export const useApproveGRN = (grnId: string, onSuccess?: (data: GoodsReceivedNote) => void) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({
      signature,
      comments,
    }: {
      signature: string;
      comments?: string;
    }) => {
      const response = await approveGRNAction(grnId, signature, comments);
      if (!response.success) {
        throw new Error(response.message);
      }
      return response;
    },
    onSuccess: (response) => {
      // Check if automation was used
      const automationUsed = (response as any).automationUsed;
      const autoCreatedPV = (response as any).autoCreatedPV;

      if (automationUsed && autoCreatedPV) {
        toast.success('GRN approved and Payment Voucher created automatically');
      } else {
        toast.success('GRN approved successfully');
      }

      queryClient.setQueryData([QUERY_KEYS.GRN.BY_ID, grnId], response.data);
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.GRN.ALL],
      });

      // If Payment Voucher was auto-created, invalidate PV cache
      if (automationUsed && autoCreatedPV) {
        queryClient.invalidateQueries({
          queryKey: [QUERY_KEYS.PAYMENT_VOUCHERS.ALL],
        });
        queryClient.invalidateQueries({
          queryKey: [QUERY_KEYS.PAYMENT_VOUCHERS.STATS],
        });
      }

      // Invalidate dashboard metrics
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.DASHBOARD.METRICS],
      });
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.DASHBOARD.ACTIVITIES],
      });

      if (onSuccess && response.data) {
        onSuccess(response.data);
      }
    },
    onError: (error: any) => {
      toast.error(error.message || 'Failed to approve GRN');
    },
  });
};

/**
 * Reject a GRN
 * Automatically invalidates GRN queries
 *
 * @param grnId - GRN ID to reject
 * @param onSuccess - Optional callback after successful rejection
 * @returns Mutation object with mutateAsync, isPending, error
 *
 * @example
 * const { mutateAsync: rejectGRN } = useRejectGRN(grnId);
 * await rejectGRN({
 *   signature: 'user-signature',
 *   remarks: 'Items damaged'
 * });
 */
export const useRejectGRN = (grnId: string, onSuccess?: (data: GoodsReceivedNote) => void) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({
      signature,
      remarks,
    }: {
      signature: string;
      remarks: string;
    }) => {
      const response = await rejectGRNAction(grnId, signature, remarks);
      if (!response.success) {
        throw new Error(response.message);
      }
      return response;
    },
    onSuccess: (response) => {
      toast.success('GRN rejected successfully');

      queryClient.setQueryData([QUERY_KEYS.GRN.BY_ID, grnId], response.data);
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.GRN.ALL],
      });

      if (onSuccess && response.data) {
        onSuccess(response.data);
      }
    },
    onError: (error: any) => {
      toast.error(error.message || 'Failed to reject GRN');
    },
  });
};

/**
 * Confirm a GRN (Mark as confirmed/received)
 * Automatically invalidates GRN queries
 *
 * @param grnId - GRN ID to confirm
 * @param onSuccess - Optional callback after successful confirmation
 * @returns Mutation object with mutateAsync, isPending, error
 *
 * @example
 * const { mutateAsync: confirmGRN } = useConfirmGRN(grnId);
 * await confirmGRN();
 */
export const useConfirmGRN = (grnId: string, onSuccess?: (data: GoodsReceivedNote) => void) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async () => {
      const response = await confirmGRNAction(grnId);
      if (!response.success) {
        throw new Error(response.message);
      }
      return response;
    },
    onSuccess: (response) => {
      toast.success('GRN confirmed successfully');

      queryClient.setQueryData([QUERY_KEYS.GRN.BY_ID, grnId], response.data);
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.GRN.ALL],
      });

      if (onSuccess && response.data) {
        onSuccess(response.data);
      }
    },
    onError: (error: any) => {
      toast.error(error.message || 'Failed to confirm GRN');
    },
  });
};

/**
 * Delete a GRN (only DRAFT GRNs can be deleted)
 * Automatically invalidates GRN queries
 *
 * @param grnId - GRN ID to delete
 * @param onSuccess - Optional callback after successful deletion
 * @returns Mutation object with mutateAsync, isPending, error
 *
 * @example
 * const { mutateAsync: deleteGRN } = useDeleteGRN(grnId);
 * await deleteGRN();
 */
export const useDeleteGRN = (grnId: string, onSuccess?: () => void) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async () => {
      const response = await deleteGRNAction(grnId);
      if (!response.success) {
        throw new Error(response.message);
      }
      return response;
    },
    onSuccess: () => {
      toast.success('GRN deleted successfully');

      queryClient.removeQueries({
        queryKey: [QUERY_KEYS.GRN.BY_ID, grnId],
      });

      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.GRN.ALL],
      });

      if (onSuccess) {
        onSuccess();
      }
    },
    onError: (error: any) => {
      toast.error(error.message || 'Failed to delete GRN');
    },
  });
};
