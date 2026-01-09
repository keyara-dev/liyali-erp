"use client";

import { create } from "zustand";
import { subscribeWithSelector } from "zustand/middleware";
import {
  fetchUserOrganizations,
  switchOrganization,
  type Organization,
} from "@/app/_actions/organizations";

interface OrganizationState {
  // State
  currentOrganization: Organization | null;
  userOrganizations: Organization[];
  isLoading: boolean;
  error: string | null;
  isInitialized: boolean;

  // Actions
  setCurrentOrganization: (orgId: string) => void;
  setUserOrganizations: (organizations: Organization[]) => void;
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;
  setInitialized: (initialized: boolean) => void;
  switchWorkspace: (orgId: string) => Promise<void>;
  fetchOrganizations: () => Promise<void>;
  refreshOrganizations: () => Promise<void>;
  retryFetch: () => Promise<void>;
  initialize: () => void;
}

export const useOrganizationStore = create<OrganizationState>()(
  subscribeWithSelector((set, get) => ({
    // Initial state
    currentOrganization: null,
    userOrganizations: [],
    isLoading: false,
    error: null,
    isInitialized: false,

    // Actions
    setCurrentOrganization: (orgId: string) => {
      const { userOrganizations } = get();
      const organization = userOrganizations.find((org) => org.id === orgId);

      if (organization) {
        set({ currentOrganization: organization });
        localStorage.setItem("current-organization-id", orgId);
      }
    },

    setUserOrganizations: (organizations: Organization[]) => {
      set({ userOrganizations: organizations });
    },

    setLoading: (loading: boolean) => {
      set({ isLoading: loading });
    },

    setError: (error: string | null) => {
      set({ error });
    },

    setInitialized: (initialized: boolean) => {
      set({ isInitialized: initialized });
    },

    switchWorkspace: async (orgId: string) => {
      const { setLoading, setError, setCurrentOrganization } = get();

      try {
        setLoading(true);
        setError(null);

        await switchOrganization(orgId);
        setCurrentOrganization(orgId);

        // Invalidate React Query cache if available
        if (typeof window !== "undefined" && (window as any).queryClient) {
          (window as any).queryClient.invalidateQueries();
        }
      } catch (error: any) {
        console.error("Failed to switch workspace:", error);
        setError(error.message || "Failed to switch workspace");
        throw error;
      } finally {
        setLoading(false);
      }
    },

    fetchOrganizations: async () => {
      const {
        setLoading,
        setError,
        setUserOrganizations,
        setCurrentOrganization,
        isInitialized,
      } = get();

      if (!isInitialized) return;

      try {
        setLoading(true);
        setError(null);

        // Verify session before fetching to prevent race conditions
        try {
          const { verifySession } = await import("@/lib/auth");
          const { isAuthenticated } = await verifySession();
          if (!isAuthenticated) {
            throw new Error("No valid session found");
          }
        } catch (error) {
          console.error(
            "Session verification failed before fetching organizations:",
            error
          );
          throw new Error("No valid session found");
        }

        const organizations = await fetchUserOrganizations();
        setUserOrganizations(organizations);

        // Set current organization if not already set
        const { currentOrganization } = get();
        if (!currentOrganization && organizations.length > 0) {
          const saved = localStorage.getItem("current-organization-id");
          const validOrgId =
            saved && organizations.some((org) => org.id === saved)
              ? saved
              : organizations[0].id;

          setCurrentOrganization(validOrgId);
        }
      } catch (error: any) {
        console.error("Failed to fetch organizations:", error);
        setError(error.message || "Failed to fetch organizations");
      } finally {
        setLoading(false);
      }
    },

    refreshOrganizations: async () => {
      const { fetchOrganizations } = get();
      await fetchOrganizations();
    },

    retryFetch: async () => {
      const { fetchOrganizations } = get();
      await fetchOrganizations();
    },

    initialize: () => {
      if (typeof window !== "undefined") {
        const { setInitialized, fetchOrganizations } = get();
        setInitialized(true);
        fetchOrganizations();
      }
    },
  }))
);

// Auto-initialize when the store is created on the client side
if (typeof window !== "undefined") {
  useOrganizationStore.getState().initialize();
}

// Subscribe to organization changes to sync with localStorage
useOrganizationStore.subscribe(
  (state) => state.currentOrganization,
  (currentOrganization) => {
    if (currentOrganization && typeof window !== "undefined") {
      localStorage.setItem("current-organization-id", currentOrganization.id);
    }
  }
);
