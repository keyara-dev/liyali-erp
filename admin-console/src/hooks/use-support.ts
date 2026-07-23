import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  createSupportTicket,
  getSupportTicket,
  getSupportTicketStats,
  getSupportTickets,
  updateSupportTicket,
  type CreateSupportTicketRequest,
  type SupportTicketFilters,
  type UpdateSupportTicketRequest,
} from "@/app/_actions/support";
import { queryKeys } from "@/lib/query-keys";

export function useSupportTickets(filters?: SupportTicketFilters) {
  return useQuery({
    queryKey: queryKeys.support.tickets(filters),
    queryFn: async () => {
      const result = await getSupportTickets(filters);
      if (!result.success) throw new Error(result.message);
      return result.data;
    },
  });
}

export function useSupportTicket(id: string) {
  return useQuery({
    queryKey: queryKeys.support.ticket(id),
    queryFn: async () => {
      const result = await getSupportTicket(id);
      if (!result.success) throw new Error(result.message);
      return result.data;
    },
    enabled: !!id,
  });
}

export function useSupportTicketStats() {
  return useQuery({
    queryKey: queryKeys.support.stats(),
    queryFn: async () => {
      const result = await getSupportTicketStats();
      if (!result.success) throw new Error(result.message);
      return result.data;
    },
  });
}

export function useCreateSupportTicket() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (request: CreateSupportTicketRequest) =>
      createSupportTicket(request),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.support.all });
    },
  });
}

export function useUpdateSupportTicket() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({
      id,
      request,
    }: {
      id: string;
      request: UpdateSupportTicketRequest;
    }) => updateSupportTicket(id, request),
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.support.all });
      queryClient.invalidateQueries({
        queryKey: queryKeys.support.ticket(variables.id),
      });
    },
  });
}
