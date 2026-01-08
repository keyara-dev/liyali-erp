"use client";

import { useEffect, useState } from "react";
import { useTokenRefresh } from "@/hooks/use-auth-queries";
import { toast } from "sonner";

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
  enabled = true 
}: TokenRefreshProviderProps) {
  // Disable token refresh for first 2 minutes after page load to avoid conflicts with fresh sessions
  const [isInitialLoad, setIsInitialLoad] = useState(true);
  
  useEffect(() => {
    const timer = setTimeout(() => {
      setIsInitialLoad(false);
    }, 2 * 60 * 1000); // 2 minutes
    
    return () => clearTimeout(timer);
  }, []);

  const {
    isRefreshing,
    refreshError,
    needsRefresh,
    isAuthenticated,
    session,
  } = useTokenRefresh(enabled && !isInitialLoad);

  // Show toast notifications for refresh events (optional)
  useEffect(() => {
    if (refreshError) {
      console.error("Token refresh failed:", refreshError);
      
      // Only show user-facing error for critical issues
      if (refreshError.message?.includes("No refresh token") ||
          refreshError.message?.includes("Invalid or expired")) {
        toast.error("Session expired. Please log in again.");
      }
    }
  }, [refreshError]);

  // Optional: Show refresh indicator (for debugging)
  useEffect(() => {
    if (process.env.NODE_ENV === "development" && isRefreshing) {
      console.log("🔄 Refreshing token...");
    }
  }, [isRefreshing]);

  return (
    <>
      {children}
      
      {/* Optional: Debug info in development */}
      {process.env.NODE_ENV === "development" && (
        <div className="fixed bottom-4 right-4 bg-black/80 text-white p-2 rounded text-xs font-mono">
          <div>Auth: {isAuthenticated ? "✅" : "❌"}</div>
          <div>Refresh: {needsRefresh ? "🔄" : "✅"}</div>
          <div>Token expires: {session?.expiresAt ? new Date(session.expiresAt).toLocaleTimeString() : "N/A"}</div>
          {isRefreshing && <div className="text-yellow-400">Refreshing...</div>}
          {refreshError && <div className="text-red-400">Error: {refreshError.message}</div>}
        </div>
      )}
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