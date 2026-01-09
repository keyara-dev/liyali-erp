"use client";

import { useOrganizationStore } from "@/stores/organization-store";
import type { Organization } from "@/app/_actions/organizations";

export interface OrganizationContextType {
  currentOrganization: Organization | null;
  userOrganizations: Organization[];
  switchWorkspace: (orgId: string) => Promise<void>;
  isLoading: boolean;
  error: string | null;
  refreshOrganizations: () => void;
  retryFetch: () => void;
}

/**
 * Hook that provides the same interface as the original OrganizationContext
 * This allows for easy migration from React Context to Zustand
 */
export function useOrganizationContext(): OrganizationContextType {
  const {
    currentOrganization,
    userOrganizations,
    switchWorkspace,
    isLoading,
    error,
    refreshOrganizations,
    retryFetch,
    isInitialized,
  } = useOrganizationStore();

  return {
    currentOrganization,
    userOrganizations,
    switchWorkspace,
    isLoading: isLoading || !isInitialized,
    error,
    refreshOrganizations: async () => {
      await refreshOrganizations();
    },
    retryFetch: async () => {
      await retryFetch();
    },
  };
}

// Export the store hook for direct access when needed
export { useOrganizationStore };
