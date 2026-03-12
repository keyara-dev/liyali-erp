"use client";

import { useQuery } from "@tanstack/react-query";
import { getBranches } from "@/app/_actions/config-actions";
import { queryKeys } from "@/lib/query-keys";

export interface Branch {
  id: string;
  name: string;
  code: string;
  province_id: string;
  town_id: string;
  address?: string;
  is_active: boolean;
}

export const useActiveBranches = () =>
  useQuery({
    queryKey: queryKeys.config.branches(),
    queryFn: async () => {
      const response = await getBranches({ isActive: true, page_size: 100 });
      return response.success && Array.isArray(response.data)
        ? (response.data as Branch[]).filter((b) => b.is_active)
        : [];
    },
    staleTime: 5 * 60 * 1000,
  });
