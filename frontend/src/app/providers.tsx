"use client";
import {
  isServer,
  QueryClient,
  QueryClientProvider,
  QueryCache,
  MutationCache,
} from "@tanstack/react-query";
import { ReactQueryDevtools } from "@tanstack/react-query-devtools";

import { ThemeProvider as NextThemesProvider } from "next-themes";
import { Toaster } from "sonner";

import { useOfflineQueueProcessor } from "@/hooks/use-offline-queue-processor";
import { TokenRefreshProvider } from "@/components/auth/token-refresh-provider";
import { TooltipProvider } from "@/components";
import { SessionExpiredModal } from "@/components/auth/session-expired-modal";
import { dispatchSessionExpired } from "@/lib/session-events";

function handleGlobalError(error: any) {
  if (error?.status === 401) {
    dispatchSessionExpired();
  }
}

function makeQueryClient() {
  return new QueryClient({
    queryCache: new QueryCache({ onError: handleGlobalError }),
    mutationCache: new MutationCache({ onError: handleGlobalError }),
    defaultOptions: {
      queries: {
        staleTime: 5 * 60 * 1000, // 5 minutes - data considered fresh
        gcTime: 10 * 60 * 1000, // 10 minutes - kept in memory
        retry: (failureCount, error: any) => {
          if (error?.status === 401) return false; // never retry on auth failure
          return failureCount < 3;
        },
        retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 30000),
        refetchOnWindowFocus: false,
        refetchOnReconnect: true,
        refetchOnMount: true,
      },
      mutations: {
        retry: (failureCount, error: any) => {
          if (error?.status === 401) return false;
          if (error?.type === "Network Error" || !navigator.onLine) return false;
          return failureCount < 1;
        },
      },
    },
  });
}

let browserQueryClient: QueryClient | undefined = undefined;

function getQueryClient() {
  if (isServer) {
    return makeQueryClient();
  } else {
    if (!browserQueryClient) browserQueryClient = makeQueryClient();
    return browserQueryClient;
  }
}

function StorageInitializer({ children }: { children: React.ReactNode }) {
  useOfflineQueueProcessor(); // Add offline sync processor
  return <>{children}</>;
}

export function Providers({ children }: { children: React.ReactNode }) {
  const queryClient = getQueryClient();

  return (
    <>
      <NextThemesProvider
        attribute="class"
        defaultTheme="light"
        disableTransitionOnChange
      >
        <QueryClientProvider client={queryClient}>
          <TooltipProvider>
            <StorageInitializer>
              <TokenRefreshProvider>{children}</TokenRefreshProvider>
            </StorageInitializer>
          </TooltipProvider>
          <SessionExpiredModal />
          <Toaster
            position="top-right"
            expand
            richColors
            theme="system"
            closeButton
          />
          <ReactQueryDevtools initialIsOpen={false} />
        </QueryClientProvider>
      </NextThemesProvider>
    </>
  );
}
