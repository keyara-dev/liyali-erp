import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  getImpersonationLogs,
  getImpersonationLog,
  getImpersonationStats,
  revokeImpersonationLog,
  type ImpersonationLogFilters,
} from "@/app/_actions/impersonation";
import { queryKeys } from "@/lib/query-keys";

export function useImpersonationLogs(filters?: ImpersonationLogFilters) {
  return useQuery({
    queryKey: queryKeys.impersonation.logs(filters),
    queryFn: async () => {
      const result = await getImpersonationLogs(filters);
      if (!result.success) throw new Error(result.message);
      return result.data ?? [];
    },
  });
}

export function useImpersonationLog(id: string) {
  return useQuery({
    queryKey: queryKeys.impersonation.log(id),
    queryFn: async () => {
      const result = await getImpersonationLog(id);
      if (!result.success) throw new Error(result.message);
      return result.data;
    },
    enabled: !!id,
  });
}

export function useImpersonationStats() {
  return useQuery({
    queryKey: queryKeys.impersonation.stats(),
    queryFn: async () => {
      const result = await getImpersonationStats();
      if (!result.success) throw new Error(result.message);
      return result.data;
    },
  });
}

export function useRevokeImpersonationLog() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => revokeImpersonationLog(id),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: queryKeys.impersonation.all,
      });
    },
  });
}
