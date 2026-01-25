"use client";

import { useEffect, useState } from "react";
import { useTokenRefresh } from "@/hooks/use-auth-queries";
import { toast } from "sonner";
import { SessionDebug } from "@/components/debug/session-debug";

interface TokenRefreshProviderProps {
  children: React.ReactNode;
  enabled?: boolean;
}

/**
 * Provider component that handles automatic token refresh
 * Place this high in your component tree to enable automatic token refresh
 *
 * @param children - Child components
 * @param enabled - Whether token refresh should be active (default: true)
 */
export function TokenRefreshProvider({
  children,
  enabled = true,
}: TokenRefreshProviderProps) {
  // Disable token refresh for first 2 minutes after page load to avoid conflicts with fresh sessions
  const [isInitialLoad, setIsInitialLoad] = useState(true);

  useEffect(() => {
    const timer = setTimeout(
      () => {
        setIsInitialLoad(false);
      },
      2 * 60 * 1000,
    ); // 2 minutes

    return () => clearTimeout(timer);
  }, []);

  const { isRefreshing, refreshError, needsRefresh, isAuthenticated, session } =
    useTokenRefresh(enabled && !isInitialLoad);

  // Show toast notifications for refresh events (optional)
  useEffect(() => {
    if (refreshError) {
      console.error("Token refresh failed:", refreshError);

      // Only show user-facing error for critical issues
      if (
        refreshError.message?.includes("No refresh token") ||
        refreshError.message?.includes("Invalid or expired")
      ) {
        toast.error("Session expired. Please log in again.");
      }
    }
  }, [refreshError]);

  // Debug logging for token refresh events
  useEffect(() => {
    if (process.env.NODE_ENV === "development") {
      console.log("🔍 Token Refresh Debug:", {
        isAuthenticated,
        needsRefresh,
        isRefreshing,
        hasRefreshToken: !!session?.refresh_token,
        expiresAt: session?.expiresAt,
        timeUntilExpiry: session?.expiresAt
          ? Math.round(
              (new Date(session.expiresAt).getTime() - Date.now()) / 1000,
            ) + "s"
          : "N/A",
        refreshError: refreshError?.message,
      });
    }
  }, [isAuthenticated, needsRefresh, isRefreshing, session, refreshError]);

  // Optional: Show refresh indicator (for debugging)
  useEffect(() => {
    if (process.env.NODE_ENV === "development" && isRefreshing) {
      console.log("🔄 Refreshing token...");
    }
  }, [isRefreshing]);

  return (
    <>
      {children}

      {/* Debug component in development */}
      {process.env.NODE_ENV === "development" && <SessionDebug />}
    </>
  );
}

/**
 * Hook to manually trigger token refresh
 * Useful for "Extend Session" buttons or user interactions
 */
export function useExtendSession() {
  const { refreshNow, isRefreshing } = useTokenRefresh();

  const extendSession = async () => {
    try {
      await refreshNow();
      toast.success("Session extended successfully");
    } catch (error) {
      toast.error("Failed to extend session");
      console.error("Session extension failed:", error);
    }
  };

  return {
    extendSession,
    isExtending: isRefreshing,
  };
}
