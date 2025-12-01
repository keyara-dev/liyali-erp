'use client'

import { useEffect, useRef, useState } from 'react'
import { fetchCurrencies } from "@/app/_actions/currencies";
import { fetchEventHosts } from "@/app/_actions/event-hosts";
import { fetchWhatsAppConfig } from "@/app/_actions/whatsapp";
import { QUERY_KEYS } from "@/lib/constants";
import { useQuery, UseQueryOptions } from "@tanstack/react-query";

/**
 * Generic useQuery hook with SSR-first approach
 *
 * Features:
 * - Server-side data fetching via server actions
 * - Initial data support (SSR hydration)
 * - Automatic refetch on stale data
 * - Proper error handling
 * - Loading states
 * - Deduplication of requests
 *
 * @template TData - Type of data being fetched
 * @param queryKey - Unique cache key
 * @param queryFn - Async function to fetch data (server action)
 * @param initialData - Optional initial data from server
 * @param options - Additional React Query options
 *
 * @example
 * // In server component:
 * const initialData = await getDashboardMetrics()
 *
 * // In client component:
 * function Dashboard({ initialData }) {
 *   const { data, isLoading, error } = useQueryData(
 *     ['dashboard', 'metrics'],
 *     async () => getDashboardMetrics(),
 *     initialData,
 *     { staleTime: 5 * 60 * 1000 }
 *   )
 * }
 */
export function useQueryData<TData = any>(
  queryKey: (string | number | object)[],
  queryFn: () => Promise<TData>,
  initialData?: TData,
  options?: Omit<UseQueryOptions<TData, Error, TData>, 'queryKey' | 'queryFn'>
) {
  const [isInitialized, setIsInitialized] = useState(!!initialData)
  const initialDataRef = useRef(initialData)

  // Set initial data immediately on mount
  useEffect(() => {
    if (initialData && !isInitialized) {
      setIsInitialized(true)
      initialDataRef.current = initialData
    }
  }, [initialData, isInitialized])

  return useQuery({
    queryKey,
    queryFn,
    initialData: initialDataRef.current,
    staleTime: options?.staleTime ?? 5 * 60 * 1000, // Default 5 minutes
    gcTime: options?.gcTime ?? 10 * 60 * 1000, // Default 10 minutes
    ...options,
  })
}

/**
 * Hook for fetching data with SSR support and refetch capability
 *
 * Usage with server component providing initial data:
 * @example
 * // pages/dashboard.tsx (Server Component)
 * export default async function DashboardPage() {
 *   const initialMetrics = await getDashboardMetrics()
 *   return <DashboardClient initialMetrics={initialMetrics} />
 * }
 *
 * // components/dashboard-client.tsx (Client Component)
 * 'use client'
 * function DashboardClient({ initialMetrics }) {
 *   const { data: metrics, isLoading, refetch } = useServerData(
 *     ['dashboard', 'metrics'],
 *     getDashboardMetrics,
 *     initialMetrics
 *   )
 *
 *   return (
 *     <div>
 *       {isLoading && <Spinner />}
 *       {metrics && <MetricsDisplay data={metrics} />}
 *       <button onClick={() => refetch()}>Refresh</button>
 *     </div>
 *   )
 * }
 */
export function useServerData<TData = any>(
  queryKey: (string | number | object)[],
  queryFn: () => Promise<TData>,
  initialData?: TData,
  staleTime: number = 5 * 60 * 1000
) {
  return useQueryData(queryKey, queryFn, initialData, { staleTime })
}

/**
 * Hook for data that needs frequent updates
 * @example
 * const { data: tasks } = useLiveData(['tasks'], getTasksForUser)
 */
export function useLiveData<TData = any>(
  queryKey: (string | number | object)[],
  queryFn: () => Promise<TData>,
  initialData?: TData,
  refetchInterval: number = 30 * 1000 // 30 seconds
) {
  return useQueryData(queryKey, queryFn, initialData, {
    staleTime: 0,
    refetchInterval,
  })
}

/**
 * Hook for static data that rarely changes
 * @example
 * const { data: departments } = useStaticData(['departments'], getDepartments)
 */
export function useStaticData<TData = any>(
  queryKey: (string | number | object)[],
  queryFn: () => Promise<TData>,
  initialData?: TData
) {
  return useQueryData(queryKey, queryFn, initialData, {
    staleTime: Infinity,
  })
}

// Legacy hooks using the new useQueryData implementation
export const useSellerProfiles = () =>
  useQueryData(
    [QUERY_KEYS.SELLERS],
    async () => await fetchEventHosts(),
    undefined,
    { staleTime: Infinity }
  );

export const useCurrencies = () =>
  useQueryData(
    [QUERY_KEYS.CURRENCIES],
    async () => await fetchCurrencies(),
    undefined,
    { staleTime: Infinity }
  );

export const useWhatsAppConfig = () =>
  useQueryData(
    [QUERY_KEYS.WHATSAPP_CONFIG],
    async () => await fetchWhatsAppConfig(),
    undefined,
    { staleTime: 5 * 60 * 1000 }
  );
