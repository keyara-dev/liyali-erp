"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { QUERY_KEYS } from "@/lib/constants";
import {
  getPayees,
  getPayeeById,
  createPayee,
  updatePayee,
  deletePayee,
} from "@/app/_actions/payees";
import type { CreatePayeeInput, UpdatePayeeInput, PayeeType } from "@/types/payee";
import { toast } from "sonner";

/**
 * Fetch all payees with optional type filter and search query
 */
export const usePayees = (type?: PayeeType, q?: string) =>
  useQuery({
    queryKey: [QUERY_KEYS.PAYEES.ALL, { type, q }],
    queryFn: async () => {
      const response = await getPayees({ type, q });
      return response.success ? response.data ?? [] : [];
    },
    staleTime: 5 * 60 * 1000, // 5 minutes
  });

/**
 * Fetch a single payee by ID
 */
export const usePayeeById = (id: string) =>
  useQuery({
    queryKey: [QUERY_KEYS.PAYEES.BY_ID, id],
    queryFn: async () => {
      const response = await getPayeeById(id);
      if (!response.success) throw new Error(response.message);
      return response.data;
    },
    staleTime: 5 * 60 * 1000,
    enabled: !!id,
  });

/**
 * Create a new payee
 */
export const useCreatePayee = (onSuccess?: () => void) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: CreatePayeeInput) => {
      const response = await createPayee(data);
      if (!response.success) throw new Error(response.message);
      return response;
    },
    onSuccess: () => {
      toast.success("Payee created successfully");
      queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.PAYEES.ALL] });
      onSuccess?.();
    },
    onError: (error: Error) => {
      toast.error(error.message || "Failed to create payee");
    },
  });
};

/**
 * Update an existing payee
 */
export const useUpdatePayee = (onSuccess?: () => void) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ id, data }: { id: string; data: UpdatePayeeInput }) => {
      const response = await updatePayee(id, data);
      if (!response.success) throw new Error(response.message);
      return response;
    },
    onSuccess: (_response, { id }) => {
      toast.success("Payee updated successfully");
      queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.PAYEES.ALL] });
      queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.PAYEES.BY_ID, id] });
      onSuccess?.();
    },
    onError: (error: Error) => {
      toast.error(error.message || "Failed to update payee");
    },
  });
};

/**
 * Delete a payee
 */
export const useDeletePayee = (onSuccess?: () => void) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (id: string) => {
      const response = await deletePayee(id);
      if (!response.success) throw new Error(response.message);
      return response;
    },
    onSuccess: () => {
      toast.success("Payee deleted successfully");
      queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.PAYEES.ALL] });
      onSuccess?.();
    },
    onError: (error: Error) => {
      toast.error(error.message || "Failed to delete payee");
    },
  });
};
