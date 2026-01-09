/**
 * Cache Manager for Organization-Scoped Data
 * Handles cache invalidation when switching between organizations
 * Includes both client-side (React Query) and server-side (Next.js) cache invalidation
 */

import { QueryClient } from "@tanstack/react-query";

export interface CacheManagerOptions {
  queryClient?: QueryClient;
  clearLocalStorage?: boolean;
  preserveKeys?: string[];
  revalidateServerCache?: boolean;
}

/**
 * Organization-scoped query keys that should be invalidated on org switch
 */
const ORGANIZATION_SCOPED_QUERIES = [
  "requisitions",
  "purchase-orders",
  "grn",
  "goods-received-notes",
  "budgets",
  "analytics",
  "dashboard",
  "workflows",
  "approvals",
  "categories",
  "vendors",
  "users",
  "members",
  "settings",
  "reports",
  "notifications",
  "audit-logs",
  "documents",
  "templates",
  "payment-vouchers",
  "invoices",
  "contracts",
  "projects",
  "departments",
  "cost-centers",
  "inventory",
  "assets",
] as const;

/**
 * LocalStorage keys that are organization-specific and should be cleared
 */
const ORGANIZATION_SCOPED_STORAGE_PATTERNS = [
  "requisitions",
  "purchase-orders",
  "grn",
  "budget",
  "analytics",
  "workflow",
  "approval",
  "dashboard",
  "filters",
  "preferences",
  "table-state",
  "form-data",
  "draft",
  "cache",
] as const;

/**
 * Keys that should never be cleared (global app state)
 */
const PRESERVE_STORAGE_KEYS = [
  "current-organization-id",
  "theme",
  "sidebar-state",
  "user-preferences",
  "auth-session",
  "language",
  "timezone",
] as const;

class CacheManager {
  private queryClient: QueryClient | null = null;

  constructor() {
    // Try to get the global query client
    if (typeof window !== "undefined") {
      this.queryClient = (window as any).queryClient || null;
    }
  }

  /**
   * Set the query client instance
   */
  setQueryClient(queryClient: QueryClient) {
    this.queryClient = queryClient;
  }

  /**
   * Clear all organization-scoped cache data
   */
  async clearOrganizationCache(options: CacheManagerOptions = {}) {
    const {
      queryClient = this.queryClient,
      clearLocalStorage = true,
      preserveKeys = [],
      revalidateServerCache = true,
    } = options;

    console.log("[CacheManager] Clearing organization-scoped cache...");

    // Clear React Query cache
    if (queryClient) {
      await this.clearQueryCache(queryClient);
    }

    // Clear localStorage
    if (clearLocalStorage) {
      this.clearOrganizationStorage(preserveKeys);
    }

    // Revalidate server-side cache
    if (revalidateServerCache) {
      await this.revalidateServerCache();
    }

    console.log("[CacheManager] Organization cache cleared successfully");
  }

  /**
   * Clear React Query cache for organization-scoped queries
   */
  private async clearQueryCache(queryClient: QueryClient) {
    try {
      // Option 1: Clear all cache (most thorough)
      await queryClient.clear();

      // Option 2: Invalidate specific organization-scoped queries
      // for (const queryKey of ORGANIZATION_SCOPED_QUERIES) {
      //   await queryClient.invalidateQueries({
      //     queryKey: [queryKey],
      //     exact: false,
      //   });
      // }

      // Force refetch of active queries
      await queryClient.refetchQueries({
        type: "active",
        stale: true,
      });

      console.log("[CacheManager] React Query cache cleared");
    } catch (error) {
      console.error("[CacheManager] Failed to clear React Query cache:", error);
    }
  }

  /**
   * Clear organization-specific localStorage data
   */
  private clearOrganizationStorage(preserveKeys: string[] = []) {
    if (typeof window === "undefined") return;

    try {
      const allPreserveKeys = [...PRESERVE_STORAGE_KEYS, ...preserveKeys];
      const keysToRemove: string[] = [];

      // Identify keys to remove
      for (let i = 0; i < localStorage.length; i++) {
        const key = localStorage.key(i);
        if (!key) continue;

        // Skip preserved keys
        if (allPreserveKeys.some((preserveKey) => key.includes(preserveKey))) {
          continue;
        }

        // Remove organization-scoped keys
        if (
          ORGANIZATION_SCOPED_STORAGE_PATTERNS.some((pattern) =>
            key.includes(pattern)
          )
        ) {
          keysToRemove.push(key);
        }
      }

      // Remove identified keys
      keysToRemove.forEach((key) => {
        localStorage.removeItem(key);
        console.log(`[CacheManager] Removed localStorage key: ${key}`);
      });

      console.log(
        `[CacheManager] Cleared ${keysToRemove.length} localStorage keys`
      );
    } catch (error) {
      console.error("[CacheManager] Failed to clear localStorage:", error);
    }
  }

  /**
   * Revalidate server-side cache using Next.js revalidation
   */
  private async revalidateServerCache(organizationId?: string) {
    try {
      console.log("[CacheManager] Revalidating server-side cache...");

      // Import the server action dynamically to avoid SSR issues
      const { revalidateOrganizationCache } =
        await import("@/app/_actions/cache-revalidation");

      const result = await revalidateOrganizationCache(organizationId);

      if (result.success) {
        console.log(
          "[CacheManager] Server-side cache revalidated successfully"
        );
      } else {
        console.warn(
          "[CacheManager] Server-side cache revalidation failed:",
          result.message
        );
      }

      return result;
    } catch (error) {
      console.error(
        "[CacheManager] Failed to revalidate server-side cache:",
        error
      );
      return {
        success: false,
        message: "Failed to revalidate server cache",
        error,
      };
    }
  }

  /**
   * Invalidate specific query patterns
   */
  async invalidateQueries(patterns: string[], queryClient?: QueryClient) {
    const client = queryClient || this.queryClient;
    if (!client) return;

    try {
      for (const pattern of patterns) {
        await client.invalidateQueries({
          queryKey: [pattern],
          exact: false,
        });
      }
      console.log(`[CacheManager] Invalidated queries: ${patterns.join(", ")}`);
    } catch (error) {
      console.error("[CacheManager] Failed to invalidate queries:", error);
    }
  }

  /**
   * Prefetch critical data for new organization
   */
  async prefetchOrganizationData(
    organizationId: string,
    queryClient?: QueryClient
  ) {
    const client = queryClient || this.queryClient;
    if (!client) return;

    console.log(
      `[CacheManager] Prefetching data for organization: ${organizationId}`
    );

    try {
      // Prefetch critical queries that are likely to be needed immediately
      const criticalQueries = [
        "dashboard",
        "notifications",
        "user-permissions",
        "organization-settings",
      ];

      // Note: You would implement actual prefetch logic here based on your API structure
      // This is a placeholder for the concept
      for (const queryKey of criticalQueries) {
        // await client.prefetchQuery({
        //   queryKey: [queryKey, organizationId],
        //   queryFn: () => fetchDataForQuery(queryKey, organizationId),
        //   staleTime: 5 * 60 * 1000, // 5 minutes
        // });
      }

      console.log("[CacheManager] Critical data prefetched");
    } catch (error) {
      console.error(
        "[CacheManager] Failed to prefetch organization data:",
        error
      );
    }
  }

  /**
   * Get cache statistics
   */
  getCacheStats() {
    if (!this.queryClient) return null;

    const cache = this.queryClient.getQueryCache();
    const queries = cache.getAll();

    return {
      totalQueries: queries.length,
      activeQueries: queries.filter((q) => q.getObserversCount() > 0).length,
      staleQueries: queries.filter((q) => q.isStale()).length,
      organizationScopedQueries: queries.filter((q) =>
        ORGANIZATION_SCOPED_QUERIES.some((pattern) =>
          q.queryKey.some(
            (key) => typeof key === "string" && key.includes(pattern)
          )
        )
      ).length,
    };
  }
}

// Global cache manager instance
export const cacheManager = new CacheManager();

// Convenience functions
export const clearOrganizationCache = (options?: CacheManagerOptions) =>
  cacheManager.clearOrganizationCache(options);

export const invalidateQueries = (
  patterns: string[],
  queryClient?: QueryClient
) => cacheManager.invalidateQueries(patterns, queryClient);

export const prefetchOrganizationData = (
  organizationId: string,
  queryClient?: QueryClient
) => cacheManager.prefetchOrganizationData(organizationId, queryClient);

export const getCacheStats = () => cacheManager.getCacheStats();

// Initialize cache manager with query client when available
if (typeof window !== "undefined") {
  // Wait for query client to be available
  const checkForQueryClient = () => {
    if ((window as any).queryClient) {
      cacheManager.setQueryClient((window as any).queryClient);
    } else {
      setTimeout(checkForQueryClient, 100);
    }
  };
  checkForQueryClient();
}
