"use client";

import { useCallback } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { usePathname, useRouter } from "next/navigation";
import { clearOrganizationCache, invalidateQueries } from "@/lib/cache-manager";
import {
  revalidateOrganizationCache,
  revalidateSpecificPaths,
  revalidateSpecificTags,
} from "@/app/_actions/cache-revalidation";

/**
 * Hook for manual cache revalidation
 * Provides functions to revalidate different types of cache
 */
export function useCacheRevalidation() {
  const queryClient = useQueryClient();
  const pathname = usePathname();
  const router = useRouter();

  /**
   * Revalidate all organization-scoped cache (client + server)
   */
  const revalidateOrganizationData = useCallback(
    async (organizationId?: string) => {
      try {
        console.log("[useCacheRevalidation] Revalidating organization data...");

        // Clear client-side cache
        await clearOrganizationCache({
          queryClient,
          clearLocalStorage: true,
          revalidateServerCache: true,
        });

        // Force refresh the current page to show new data
        router.refresh();

        console.log(
          "[useCacheRevalidation] Organization data revalidated successfully"
        );
        return { success: true };
      } catch (error) {
        console.error(
          "[useCacheRevalidation] Failed to revalidate organization data:",
          error
        );
        return { success: false, error };
      }
    },
    [queryClient, router]
  );

  /**
   * Revalidate specific query patterns (client-side only)
   */
  const revalidateQueries = useCallback(
    async (patterns: string[]) => {
      try {
        console.log(
          `[useCacheRevalidation] Revalidating queries: ${patterns.join(", ")}`
        );

        await invalidateQueries(patterns, queryClient);

        console.log("[useCacheRevalidation] Queries revalidated successfully");
        return { success: true };
      } catch (error) {
        console.error(
          "[useCacheRevalidation] Failed to revalidate queries:",
          error
        );
        return { success: false, error };
      }
    },
    [queryClient]
  );

  /**
   * Revalidate specific server-side paths
   */
  const revalidatePaths = useCallback(
    async (paths: string[]) => {
      try {
        console.log(
          `[useCacheRevalidation] Revalidating paths: ${paths.join(", ")}`
        );

        const result = await revalidateSpecificPaths(paths);

        if (result.success) {
          // Force refresh if current path is being revalidated
          if (paths.some((path) => pathname.startsWith(path))) {
            router.refresh();
          }
        }

        return result;
      } catch (error) {
        console.error(
          "[useCacheRevalidation] Failed to revalidate paths:",
          error
        );
        return { success: false, error };
      }
    },
    [pathname, router]
  );

  /**
   * Revalidate specific server-side cache tags
   */
  const revalidateTags = useCallback(
    async (tags: string[]) => {
      try {
        console.log(
          `[useCacheRevalidation] Revalidating tags: ${tags.join(", ")}`
        );

        const result = await revalidateSpecificTags(tags);

        if (result.success) {
          // Refresh the page to show updated data
          router.refresh();
        }

        return result;
      } catch (error) {
        console.error(
          "[useCacheRevalidation] Failed to revalidate tags:",
          error
        );
        return { success: false, error };
      }
    },
    [router]
  );

  /**
   * Revalidate current page only
   */
  const revalidateCurrentPage = useCallback(async () => {
    try {
      console.log(
        `[useCacheRevalidation] Revalidating current page: ${pathname}`
      );

      // Invalidate queries for current page
      await queryClient.invalidateQueries();

      // Revalidate server-side cache for current path
      await revalidatePaths([pathname]);

      console.log(
        "[useCacheRevalidation] Current page revalidated successfully"
      );
      return { success: true };
    } catch (error) {
      console.error(
        "[useCacheRevalidation] Failed to revalidate current page:",
        error
      );
      return { success: false, error };
    }
  }, [pathname, queryClient, revalidatePaths]);

  /**
   * Force hard refresh (clears everything and reloads page)
   */
  const forceHardRefresh = useCallback(async () => {
    try {
      console.log("[useCacheRevalidation] Performing hard refresh...");

      // Clear all client-side cache
      await queryClient.clear();

      // Clear localStorage (except preserved keys)
      await clearOrganizationCache({
        queryClient,
        clearLocalStorage: true,
        revalidateServerCache: false, // Skip server revalidation for hard refresh
      });

      // Force reload the page
      window.location.reload();

      return { success: true };
    } catch (error) {
      console.error(
        "[useCacheRevalidation] Failed to perform hard refresh:",
        error
      );
      return { success: false, error };
    }
  }, [queryClient]);

  return {
    revalidateOrganizationData,
    revalidateQueries,
    revalidatePaths,
    revalidateTags,
    revalidateCurrentPage,
    forceHardRefresh,
  };
}

/**
 * Hook for getting cache statistics
 */
export function useCacheStats() {
  const queryClient = useQueryClient();

  const getCacheStats = useCallback(() => {
    const cache = queryClient.getQueryCache();
    const queries = cache.getAll();

    return {
      totalQueries: queries.length,
      activeQueries: queries.filter((q) => q.getObserversCount() > 0).length,
      staleQueries: queries.filter((q) => q.isStale()).length,
      errorQueries: queries.filter((q) => q.state.status === "error").length,
      loadingQueries: queries.filter((q) => q.state.status === "pending")
        .length,
    };
  }, [queryClient]);

  return { getCacheStats };
}
