"use client";

import { useEffect, useState } from "react";
import { getRefreshToken } from "@/app/_actions/auth";
import type { APIResponse } from "@/types";

interface UseRefreshTokenResponse {
  data: any;
  error: Error | null;
  isLoading: boolean;
}

/**
 * Hook to refresh user token at intervals
 * Automatically refreshes JWT token when user is active
 * @param shouldRefresh - Whether token refresh should be active
 * @param interval - Interval in milliseconds for token refresh (default: 20 minutes)
 * @returns Object with data, error, and loading state
 */
export function useRefreshToken(
  shouldRefresh: boolean = true,
  interval: number = 20 * 60 * 1000
): UseRefreshTokenResponse {
  const [data, setData] = useState<any>(null);
  const [error, setError] = useState<Error | null>(null);
  const [isLoading, setIsLoading] = useState(false);

  useEffect(() => {
    if (!shouldRefresh) {
      return;
    }

    setIsLoading(true);

    // Perform initial refresh
    const performRefresh = async () => {
      try {
        const response = await getRefreshToken();

        if (response.success) {
          setData(response.data);
          setError(null);
        } else {
          setError(new Error(response.message || "Failed to refresh token"));
          setData(null);
        }
      } catch (err: any) {
        setError(
          err instanceof Error ? err : new Error("Token refresh failed")
        );
        setData(null);
      } finally {
        setIsLoading(false);
      }
    };

    performRefresh();

    // Setup interval for periodic refresh
    const refreshInterval = setInterval(() => {
      performRefresh();
    }, interval);

    return () => clearInterval(refreshInterval);
  }, [shouldRefresh, interval]);

  return { data, error, isLoading };
}
