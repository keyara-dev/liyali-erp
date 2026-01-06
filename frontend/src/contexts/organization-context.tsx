'use client';

import { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { fetchUserOrganizations, switchOrganization, type Organization } from '@/app/_actions/organizations';

export interface OrganizationContextType {
  currentOrganization: Organization | null;
  userOrganizations: Organization[];
  switchWorkspace: (orgId: string) => Promise<void>;
  isLoading: boolean;
  error: string | null;
  refreshOrganizations: () => void;
  retryFetch: () => void;
}

const OrganizationContext = createContext<OrganizationContextType | undefined>(undefined);

export function OrganizationProvider({ children }: { children: ReactNode }) {
  const queryClient = useQueryClient();
  const [currentOrgId, setCurrentOrgId] = useState<string | null>(null);
  const [isClient, setIsClient] = useState(false);

  // Ensure we're on the client side
  useEffect(() => {
    setIsClient(true);
  }, []);

  // Fetch user's organizations
  const { data: organizations = [], isLoading, error, refetch } = useQuery({
    queryKey: ['organizations'],
    queryFn: () => fetchUserOrganizations(),
    staleTime: 5 * 60 * 1000, // 5 minutes
    gcTime: 10 * 60 * 1000, // 10 minutes (formerly cacheTime)
    retry: 3,
    retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 30000),
    enabled: isClient, // Only run query on client side
  });

  // Get current organization
  const currentOrganization = organizations.find(org => org.id === currentOrgId) || null;

  // Set initial current org - prioritize localStorage, then first available
  // Organization from signup is available in the organizations list
  useEffect(() => {
    if (isClient && organizations.length > 0 && !currentOrgId) {
      // Try these in order:
      // 1. Organization from localStorage (from previous session or signup)
      // 2. Default to first available organization (new user's personal org)
      const saved = localStorage.getItem('current-organization-id');
      
      // Validate that the saved org ID exists in the current organizations list
      const validOrgId = saved && organizations.some(org => org.id === saved) 
        ? saved 
        : organizations[0].id;

      setCurrentOrgId(validOrgId);
      localStorage.setItem('current-organization-id', validOrgId);
    }
  }, [isClient, organizations, currentOrgId]);

  // Switch workspace mutation
  const switchMutation = useMutation({
    mutationFn: (orgId: string) => switchOrganization(orgId),
    onSuccess: (orgId) => {
      setCurrentOrgId(orgId);
      localStorage.setItem('current-organization-id', orgId);

      // Invalidate all queries to refetch with new org context
      queryClient.invalidateQueries();
    },
  });

  const switchWorkspace = async (orgId: string) => {
    await switchMutation.mutateAsync(orgId);
  };

  return (
    <OrganizationContext.Provider
      value={{
        currentOrganization,
        userOrganizations: organizations,
        switchWorkspace,
        isLoading: isLoading || !isClient,
        error: error?.message || null,
        refreshOrganizations: () => refetch(),
        retryFetch: () => refetch(),
      }}
    >
      {children}
    </OrganizationContext.Provider>
  );
}

export function useOrganizationContext() {
  const context = useContext(OrganizationContext);
  if (!context) {
    throw new Error('useOrganizationContext must be used within OrganizationProvider');
  }
  return context;
}
