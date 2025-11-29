import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  fetchAccountTiers,
  createAccountTier,
  updateAccountTier,
  deleteAccountTier,
} from "@/app/_actions/account-tiers";

const ACCOUNT_TIERS_KEY = ["account-tiers"] as const;

interface CreateAccountTierParams {
  name: string;
  description?: string;
  maxProducts?: number;
  maxImagesPerProduct?: number;
  maxVariantsPerProduct?: number;
  price?: number;
  currencyId?: string;
  billingPeriod?: string;
  isDefault?: boolean;
  active?: boolean;
  sortOrder?: number;
  whatsappBusinessEnabled?: boolean;
}

interface UpdateAccountTierParams {
  tierId: string;
  updates: Record<string, any>;
}

/**
 * Hook for fetching all account tiers
 */
export function useAccountTiers() {
  return useQuery({
    queryKey: ACCOUNT_TIERS_KEY,
    queryFn: async () => {
      const result = await fetchAccountTiers();
      if (!result.success) {
        throw new Error(result.message);
      }
      return result.data;
    },
  });
}

/**
 * Hook for creating a new account tier
 */
export function useCreateAccountTier() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: createAccountTier,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ACCOUNT_TIERS_KEY });
    },
  });
}

/**
 * Hook for updating an account tier
 */
export function useUpdateAccountTier() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: updateAccountTier,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ACCOUNT_TIERS_KEY });
    },
  });
}

/**
 * Hook for deleting an account tier
 */
export function useDeleteAccountTier() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: deleteAccountTier,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ACCOUNT_TIERS_KEY });
    },
  });
}
