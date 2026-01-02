"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { QUERY_KEYS } from "@/lib/constants";
import {
  getPurchaseOrders,
  getPurchaseOrderById,
  createPurchaseOrder,
  updatePurchaseOrder,
  submitPurchaseOrderForApproval,
  approvePurchaseOrder,
  rejectPurchaseOrder,
  deletePurchaseOrder,
  getPurchaseOrderStats,
} from "@/app/_actions/purchase-orders";
import {
  PurchaseOrder,
  PurchaseOrderStats,
  CreatePurchaseOrderRequest,
  UpdatePurchaseOrderRequest,
  SubmitPurchaseOrderRequest,
  ApprovePurchaseOrderRequest,
  RejectPurchaseOrderRequest,
} from "@/types/purchase-order";
import { toast } from "sonner";

/**
 * Fetch all purchase orders
 * Standard data - 5 minute refresh interval
 *
 * @param initialPOs - Optional initial data from server component
 * @returns Query result with purchase orders array
 *
 * @example
 * const { data: purchaseOrders } = usePurchaseOrders(initialPOs)
 */
export const usePurchaseOrders = (initialPOs?: PurchaseOrder[]) =>
  useQuery({
    queryKey: [QUERY_KEYS.PURCHASE_ORDERS.ALL],
    queryFn: async () => {
      const response = await getPurchaseOrders();
      return response.success ? response.data : [];
    },
    initialData: initialPOs,
    staleTime: 5 * 60 * 1000, // 5 minutes
  });

/**
 * Fetch a specific purchase order by ID
 *
 * @param poId - Purchase Order ID to fetch
 * @returns Query result with single purchase order
 *
 * @example
 * const { data: purchaseOrder } = usePurchaseOrderById(poId)
 */
export const usePurchaseOrderById = (
  poId: string,
  initialData?: PurchaseOrder
) =>
  useQuery({
    queryKey: [QUERY_KEYS.PURCHASE_ORDERS.BY_ID, poId],
    queryFn: async () => {
      const response = await getPurchaseOrderById(poId);
      if (!response.success) throw new Error(response.message);
      return response.data;
    },
    initialData,
    staleTime: 5 * 60 * 1000, // 5 minutes
  });

/**
 * Fetch purchase order statistics
 *
 * @param initialStats - Optional initial data from server component
 * @returns Query result with purchase order statistics
 *
 * @example
 * const { data: stats } = usePurchaseOrderStats()
 */
export const usePurchaseOrderStats = (initialStats?: PurchaseOrderStats) =>
  useQuery({
    queryKey: [QUERY_KEYS.PURCHASE_ORDERS.STATS],
    queryFn: async () => {
      const response = await getPurchaseOrderStats();
      return response.success ? response.data : null;
    },
    initialData: initialStats,
    staleTime: 10 * 60 * 1000, // 10 minutes
  });

/**
 * Create or update purchase order mutation
 * Handles both create (no ID) and update (with ID) operations
 * Only DRAFT purchase orders can be updated
 *
 * @param onSuccess - Callback after successful mutation
 * @returns Mutation object with mutate and mutateAsync
 *
 * @example
 * const saveMutation = useSavePurchaseOrder(() => {
 *   queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.PURCHASE_ORDERS.ALL] })
 * })
 *
 * // Create
 * await saveMutation.mutateAsync({
 *   title: 'IT Equipment',
 *   vendorName: 'Tech Supplier',
 *   items: [...],
 *   createdBy: userId
 * })
 *
 * // Update
 * await saveMutation.mutateAsync({
 *   poId: 'po-1',
 *   title: 'IT Equipment Updated',
 *   items: [...]
 * })
 */
export const useSavePurchaseOrder = (onSuccess?: () => void) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (
      data:
        | CreatePurchaseOrderRequest
        | (UpdatePurchaseOrderRequest & { poId?: string })
    ) => {
      const response =
        "poId" in data && data.poId
          ? await updatePurchaseOrder(data as UpdatePurchaseOrderRequest)
          : await createPurchaseOrder(data as CreatePurchaseOrderRequest);

      if (!response.success) {
        throw new Error(response.message);
      }
      return response;
    },
    onSuccess: (response) => {
      const isUpdate = (
        response.data as PurchaseOrder & { poId?: string }
      )?.poId;
      toast.success(
        isUpdate
          ? "Purchase order updated successfully"
          : "Purchase order created successfully"
      );

      // Invalidate purchase order queries
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.PURCHASE_ORDERS.ALL],
      });
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.PURCHASE_ORDERS.STATS],
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
      toast.error(error.message || "Failed to save purchase order");
    },
  });
};

/**
 * Submit purchase order for approval mutation
 *
 * @param poId - Purchase Order ID to submit
 * @param onSuccess - Callback after successful submission
 * @returns Mutation object
 *
 * @example
 * const submitMutation = useSubmitPurchaseOrderForApproval(poId)
 * await submitMutation.mutateAsync({
 *   submittedBy: userId,
 *   submittedByName: 'John Doe',
 *   submittedByRole: 'REQUESTER',
 *   comments: 'Please review'
 * })
 */
export const useSubmitPurchaseOrderForApproval = (
  poId: string,
  onSuccess?: () => void
) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (
      data: Omit<SubmitPurchaseOrderRequest, "poId">
    ) => {
      const response = await submitPurchaseOrderForApproval({
        poId,
        ...data,
      });

      if (!response.success) {
        throw new Error(response.message);
      }
      return response;
    },
    onSuccess: () => {
      toast.success("Purchase order submitted for approval");

      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.PURCHASE_ORDERS.BY_ID, poId],
      });
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.PURCHASE_ORDERS.ALL],
      });
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.PURCHASE_ORDERS.STATS],
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
      toast.error(error.message || "Failed to submit purchase order");
    },
  });
};

/**
 * Approve purchase order mutation
 * Handles automatic GRN creation when PO is fully approved
 *
 * @param poId - Purchase Order ID to approve
 * @param onSuccess - Callback after successful approval
 * @returns Mutation object
 *
 * @example
 * const approveMutation = useApprovePurchaseOrder(poId)
 * await approveMutation.mutateAsync({
 *   approvingUserId: userId,
 *   approvingUserName: 'John Doe',
 *   approvingUserRole: 'PROCUREMENT_MANAGER',
 *   signature: signatureDataUrl,
 *   comments: 'Approved'
 * })
 */
export const useApprovePurchaseOrder = (
  poId: string,
  onSuccess?: () => void
) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (
      data: Omit<ApprovePurchaseOrderRequest, "poId">
    ) => {
      const response = await approvePurchaseOrder({
        poId,
        ...data,
      });

      if (!response.success) {
        throw new Error(response.message);
      }
      return response;
    },
    onSuccess: (response) => {
      // Check if automation was used
      const automationUsed = response.data?.automationUsed;
      const autoCreatedGRN = response.data?.autoCreatedGRN;

      if (automationUsed && autoCreatedGRN) {
        toast.success("Purchase order approved and GRN created automatically");
      } else {
        toast.success("Purchase order approved");
      }

      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.PURCHASE_ORDERS.BY_ID, poId],
      });
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.PURCHASE_ORDERS.ALL],
      });
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.PURCHASE_ORDERS.STATS],
      });

      // If GRN was auto-created, invalidate GRN cache
      if (automationUsed && autoCreatedGRN) {
        queryClient.invalidateQueries({
          queryKey: [QUERY_KEYS.GRN.ALL],
        });
        queryClient.invalidateQueries({
          queryKey: [QUERY_KEYS.GRN.STATS],
        });
      }

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
      toast.error(error.message || "Failed to approve purchase order");
    },
  });
};

/**
 * Reject purchase order mutation
 *
 * @param poId - Purchase Order ID to reject
 * @param onSuccess - Callback after successful rejection
 * @returns Mutation object
 *
 * @example
 * const rejectMutation = useRejectPurchaseOrder(poId)
 * await rejectMutation.mutateAsync({
 *   rejectingUserId: userId,
 *   rejectingUserName: 'John Doe',
 *   rejectingUserRole: 'FINANCE_MANAGER',
 *   remarks: 'Budget exceeded',
 *   signature: signatureDataUrl,
 *   comments: 'Requires adjustment'
 * })
 */
export const useRejectPurchaseOrder = (
  poId: string,
  onSuccess?: () => void
) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (
      data: Omit<RejectPurchaseOrderRequest, "poId">
    ) => {
      const response = await rejectPurchaseOrder({
        poId,
        ...data,
      });

      if (!response.success) {
        throw new Error(response.message);
      }
      return response;
    },
    onSuccess: () => {
      toast.success("Purchase order rejected and returned to draft");

      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.PURCHASE_ORDERS.BY_ID, poId],
      });
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.PURCHASE_ORDERS.ALL],
      });
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.PURCHASE_ORDERS.STATS],
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
      toast.error(error.message || "Failed to reject purchase order");
    },
  });
};

/**
 * Delete purchase order mutation
 * Only DRAFT purchase orders can be deleted
 *
 * @param onSuccess - Callback after successful deletion
 * @returns Mutation object
 *
 * @example
 * const deleteMutation = useDeletePurchaseOrder()
 * await deleteMutation.mutateAsync(poId)
 */
export const useDeletePurchaseOrder = (onSuccess?: () => void) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (poId: string) => {
      const response = await deletePurchaseOrder(poId);

      if (!response.success) {
        throw new Error(response.message);
      }
      return response;
    },
    onSuccess: () => {
      toast.success("Purchase order deleted successfully");

      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.PURCHASE_ORDERS.ALL],
      });
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.PURCHASE_ORDERS.STATS],
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
      toast.error(error.message || "Failed to delete purchase order");
    },
  });
};
