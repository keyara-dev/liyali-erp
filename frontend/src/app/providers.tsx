"use client";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { ReactQueryDevtools } from "@tanstack/react-query-devtools";

import { ThemeProvider as NextThemesProvider } from "next-themes";
import { Toaster } from "sonner";

import { SiteHeader } from "@/components/layout/header";
import { AppSidebar } from "@/components/layout/sidebar/app-sidebar";
import { IdleTimerContainer } from "@/components/base/screen-lock";
import {
  SidebarInset,
  SidebarProvider,
  useSidebar,
} from "@/components/ui/sidebar";
import { useInitializeStorage } from "@/hooks/use-initialize-storage";
import { useOfflineQueueProcessor } from "@/hooks/use-offline-queue-processor";
import { OrganizationProvider } from "@/contexts/organization-context";
import { TokenRefreshProvider } from "@/components/auth/token-refresh-provider";

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 5 * 60 * 1000,        // 5 minutes - data considered fresh
      gcTime: 10 * 60 * 1000,          // 10 minutes - kept in memory
      retry: 3,                         // Retry failed queries 3 times
      retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 30000), // Exponential backoff
      refetchOnWindowFocus: false,      // Don't auto-refetch on window focus
      refetchOnReconnect: true,         // Refetch when network reconnects
      refetchOnMount: true,             // Refetch on component mount if stale
    },
    mutations: {
      retry: (failureCount, error: any) => {
        // Don't retry if offline - let offline queue handle it
        if (error?.type === "Network Error" || !navigator.onLine) {
          return false;
        }
        return failureCount < 1; // Retry once for other errors
      },
      onError: (error) => {
        console.error('Mutation error:', error);
      },
    },
  },
});

function StorageInitializer({ children }: { children: React.ReactNode }) {
  useInitializeStorage();
  useOfflineQueueProcessor(); // Add offline sync processor
  return <>{children}</>;
}

export function Providers({ children }: { children: React.ReactNode }) {
  return (
    <>
      <NextThemesProvider
        attribute="class"
        defaultTheme="light"
        disableTransitionOnChange
      >
        <QueryClientProvider client={queryClient}>
          <OrganizationProvider>
            <StorageInitializer>  <TokenRefreshProvider>
        {children}
      </TokenRefreshProvider></StorageInitializer>
            <Toaster
              position="top-right"
              expand
              richColors
              theme="system"
              closeButton
            />
            <ReactQueryDevtools initialIsOpen={false} />
          </OrganizationProvider>
        </QueryClientProvider>
      </NextThemesProvider>
    </>
  );
}
