"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { QUERY_KEYS } from "@/lib/constants";
import {
  getPaymentVouchers,
  getPaymentVoucherById,
  createPaymentVoucher,
  updatePaymentVoucher,
  submitPaymentVoucherForApproval,
  approvePaymentVoucher,
  rejectPaymentVoucher,
  markPaymentVoucherAsPaid,
  deletePaymentVoucher,
  getPaymentVoucherStats,
} from "@/app/_actions/payment-vouchers";
import {
  PaymentVoucher,
  PaymentVoucherStats,
  CreatePaymentVoucherRequest,
  UpdatePaymentVoucherRequest,
  SubmitPaymentVoucherRequest,
  ApprovePaymentVoucherRequest,
  RejectPaymentVoucherRequest,
  MarkPaymentVoucherPaidRequest,
} from "@/types/payment-voucher";
import { toast } from "sonner";

/**
 * Fetch all payment vouchers
 * Standard data - 5 minute refresh interval
 *
 * @param initialPVs - Optional initial data from server component
 * @returns Query result with payment vouchers array
 *
 * @example
 * const { data: paymentVouchers } = usePaymentVouchers(initialPVs)
 */
export const usePaymentVouchers = (initialPVs?: PaymentVoucher[]) =>
  useQuery({
    queryKey: [QUERY_KEYS.PAYMENT_VOUCHERS.ALL],
    queryFn: async () => {
      const response = await getPaymentVouchers();
      return response.success ? response.data : [];
    },
    initialData: initialPVs,
    staleTime: 5 * 60 * 1000, // 5 minutes
  });

/**
 * Fetch a specific payment voucher by ID
 *
 * @param pvId - Payment Voucher ID to fetch
 * @returns Query result with single payment voucher
 *
 * @example
 * const { data: paymentVoucher } = usePaymentVoucherById(pvId)
 */
export const usePaymentVoucherById = (
  pvId: string,
  initialData?: PaymentVoucher
) =>
  useQuery({
    queryKey: [QUERY_KEYS.PAYMENT_VOUCHERS.BY_ID, pvId],
    queryFn: async () => {
      const response = await getPaymentVoucherById(pvId);
      if (!response.success) throw new Error(response.message);
      return response.data;
    },
    initialData,
    staleTime: 5 * 60 * 1000, // 5 minutes
  });

/**
 * Fetch payment voucher statistics
 *
 * @param initialStats - Optional initial data from server component
 * @returns Query result with payment voucher statistics
 *
 * @example
 * const { data: stats } = usePaymentVoucherStats()
 */
export const usePaymentVoucherStats = (initialStats?: PaymentVoucherStats) =>
  useQuery({
    queryKey: [QUERY_KEYS.PAYMENT_VOUCHERS.STATS],
    queryFn: async () => {
      const response = await getPaymentVoucherStats();
      return response.success ? response.data : null;
    },
    initialData: initialStats,
    staleTime: 10 * 60 * 1000, // 10 minutes
  });

/**
 * Create or update payment voucher mutation
 * Handles both create (no ID) and update (with ID) operations
 * Only DRAFT payment vouchers can be updated
 *
 * @param onSuccess - Callback after successful mutation
 * @returns Mutation object with mutate and mutateAsync
 *
 * @example
 * const saveMutation = useSavePaymentVoucher(() => {
 *   queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.PAYMENT_VOUCHERS.ALL] })
 * })
 *
 * // Create
 * await saveMutation.mutateAsync({
 *   title: 'Payment for PO-001',
 *   vendorName: 'Supplier Inc',
 *   items: [...],
 *   createdBy: userId
 * })
 *
 * // Update
 * await saveMutation.mutateAsync({
 *   pvId: 'pv-1',
 *   title: 'Payment for PO-001 Updated',
 *   items: [...]
 * })
 */
export const useSavePaymentVoucher = (onSuccess?: () => void) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (
      data:
        | CreatePaymentVoucherRequest
        | (UpdatePaymentVoucherRequest & { pvId?: string })
    ) => {
      const response =
        "pvId" in data && data.pvId
          ? await updatePaymentVoucher(data as UpdatePaymentVoucherRequest)
          : await createPaymentVoucher(data as CreatePaymentVoucherRequest);

      if (!response.success) {
        throw new Error(response.message);
      }
      return response;
    },
    onSuccess: (response) => {
      const isUpdate = (
        response.data as PaymentVoucher & { pvId?: string }
      )?.pvId;
      toast.success(
        isUpdate
          ? "Payment voucher updated successfully"
          : "Payment voucher created successfully"
      );

      // Invalidate payment voucher queries
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.PAYMENT_VOUCHERS.ALL],
      });
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.PAYMENT_VOUCHERS.STATS],
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
      toast.error(error.message || "Failed to save payment voucher");
    },
  });
};

/**
 * Submit payment voucher for approval mutation
 *
 * @param pvId - Payment Voucher ID to submit
 * @param onSuccess - Callback after successful submission
 * @returns Mutation object
 *
 * @example
 * const submitMutation = useSubmitPaymentVoucherForApproval(pvId)
 * await submitMutation.mutateAsync({
 *   submittedBy: userId,
 *   submittedByName: 'John Doe',
 *   submittedByRole: 'REQUESTER',
 *   comments: 'Please review'
 * })
 */
export const useSubmitPaymentVoucherForApproval = (
  pvId: string,
  onSuccess?: () => void
) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (
      data: Omit<SubmitPaymentVoucherRequest, "pvId">
    ) => {
      const response = await submitPaymentVoucherForApproval({
        pvId,
        ...data,
      });

      if (!response.success) {
        throw new Error(response.message);
      }
      return response;
    },
    onSuccess: () => {
      toast.success("Payment voucher submitted for approval");

      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.PAYMENT_VOUCHERS.BY_ID, pvId],
      });
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.PAYMENT_VOUCHERS.ALL],
      });
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.PAYMENT_VOUCHERS.STATS],
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
      toast.error(error.message || "Failed to submit payment voucher");
    },
  });
};

/**
 * Approve payment voucher mutation
 *
 * @param pvId - Payment Voucher ID to approve
 * @param onSuccess - Callback after successful approval
 * @returns Mutation object
 *
 * @example
 * const approveMutation = useApprovePaymentVoucher(pvId)
 * await approveMutation.mutateAsync({
 *   approvingUserId: userId,
 *   approvingUserName: 'John Doe',
 *   approvingUserRole: 'FINANCE_MANAGER',
 *   signature: signatureDataUrl,
 *   comments: 'Approved'
 * })
 */
export const useApprovePaymentVoucher = (
  pvId: string,
  onSuccess?: () => void
) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (
      data: Omit<ApprovePaymentVoucherRequest, "pvId">
    ) => {
      const response = await approvePaymentVoucher({
        pvId,
        ...data,
      });

      if (!response.success) {
        throw new Error(response.message);
      }
      return response;
    },
    onSuccess: () => {
      toast.success("Payment voucher approved");

      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.PAYMENT_VOUCHERS.BY_ID, pvId],
      });
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.PAYMENT_VOUCHERS.ALL],
      });
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.PAYMENT_VOUCHERS.STATS],
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
      toast.error(error.message || "Failed to approve payment voucher");
    },
  });
};

/**
 * Reject payment voucher mutation
 *
 * @param pvId - Payment Voucher ID to reject
 * @param onSuccess - Callback after successful rejection
 * @returns Mutation object
 *
 * @example
 * const rejectMutation = useRejectPaymentVoucher(pvId)
 * await rejectMutation.mutateAsync({
 *   rejectingUserId: userId,
 *   rejectingUserName: 'John Doe',
 *   rejectingUserRole: 'FINANCE_MANAGER',
 *   remarks: 'Budget exceeded',
 *   signature: signatureDataUrl,
 *   comments: 'Requires adjustment'
 * })
 */
export const useRejectPaymentVoucher = (
  pvId: string,
  onSuccess?: () => void
) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (
      data: Omit<RejectPaymentVoucherRequest, "pvId">
    ) => {
      const response = await rejectPaymentVoucher({
        pvId,
        ...data,
      });

      if (!response.success) {
        throw new Error(response.message);
      }
      return response;
    },
    onSuccess: () => {
      toast.success("Payment voucher rejected and returned to draft");

      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.PAYMENT_VOUCHERS.BY_ID, pvId],
      });
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.PAYMENT_VOUCHERS.ALL],
      });
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.PAYMENT_VOUCHERS.STATS],
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
      toast.error(error.message || "Failed to reject payment voucher");
    },
  });
};

/**
 * Mark payment voucher as paid mutation
 *
 * @param pvId - Payment Voucher ID to mark as paid
 * @param onSuccess - Callback after successful marking
 * @returns Mutation object
 *
 * @example
 * const markPaidMutation = useMarkPaymentVoucherAsPaid(pvId)
 * await markPaidMutation.mutateAsync({
 *   paidAmount: 5000,
 *   paidDate: new Date(),
 *   referenceNumber: 'TRANSFER-123',
 *   markedBy: userId,
 *   markedByName: 'Finance Officer',
 *   markedByRole: 'FINANCE_OFFICER'
 * })
 */
export const useMarkPaymentVoucherAsPaid = (
  pvId: string,
  onSuccess?: () => void
) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (
      data: Omit<MarkPaymentVoucherPaidRequest, "pvId">
    ) => {
      const response = await markPaymentVoucherAsPaid({
        pvId,
        ...data,
      });

      if (!response.success) {
        throw new Error(response.message);
      }
      return response;
    },
    onSuccess: () => {
      toast.success("Payment voucher marked as paid");

      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.PAYMENT_VOUCHERS.BY_ID, pvId],
      });
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.PAYMENT_VOUCHERS.ALL],
      });
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.PAYMENT_VOUCHERS.STATS],
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
      toast.error(error.message || "Failed to mark payment voucher as paid");
    },
  });
};

/**
 * Delete payment voucher mutation
 * Only DRAFT payment vouchers can be deleted
 *
 * @param onSuccess - Callback after successful deletion
 * @returns Mutation object
 *
 * @example
 * const deleteMutation = useDeletePaymentVoucher()
 * await deleteMutation.mutateAsync(pvId)
 */
export const useDeletePaymentVoucher = (onSuccess?: () => void) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (pvId: string) => {
      const response = await deletePaymentVoucher(pvId);

      if (!response.success) {
        throw new Error(response.message);
      }
      return response;
    },
    onSuccess: () => {
      toast.success("Payment voucher deleted successfully");

      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.PAYMENT_VOUCHERS.ALL],
      });
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.PAYMENT_VOUCHERS.STATS],
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
      toast.error(error.message || "Failed to delete payment voucher");
    },
  });
};
