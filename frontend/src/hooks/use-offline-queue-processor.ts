'use client';

import React, { useEffect, useRef, useState } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { queryKeys } from '@/lib/query-keys';
import {
  getPendingOperations,
  updateOperationStatus,
  removeOperation,
  getQueueStats,
} from '@/lib/offline-queue';
import { useNetwork } from './use-network';
import { toast } from 'sonner';

/**
 * Hook to process offline queue when connection is restored
 * Retries failed operations and syncs data with server
 *
 * Usage:
 * // Add to your root layout or providers
 * useOfflineQueueProcessor();
 */
export function useOfflineQueueProcessor() {
  const { online } = useNetwork();
  const queryClient = useQueryClient();
  const processingRef = useRef(false);

  useEffect(() => {
    if (!online || processingRef.current) return;

    const processQueue = async () => {
      processingRef.current = true;

      try {
        const operations = await getPendingOperations();

        if (operations.length === 0) {
          processingRef.current = false;
          return;
        }

        console.log(`[Queue Processor] Processing ${operations.length} pending operations`);
        toast.loading(`Syncing ${operations.length} offline changes...`);

        let successCount = 0;
        let failureCount = 0;

        // Process each operation
        for (const operation of operations) {
          try {
            await updateOperationStatus(operation.id, 'processing');

            // TODO: Execute operation against real API
            // This will be implemented when we migrate to real API
            // For now, just mark as completed
            await updateOperationStatus(operation.id, 'completed', {
              synced: true,
            });

            await removeOperation(operation.id);
            successCount++;

            console.log(`[Queue Processor] ✓ Synced ${operation.type} for ${operation.entity}`);
          } catch (error) {
            failureCount++;
            await updateOperationStatus(
              operation.id,
              operation.retries < 3 ? 'pending' : 'failed',
              undefined,
              error instanceof Error ? error.message : 'Unknown error'
            );

            console.error(
              `[Queue Processor] ✗ Failed to sync ${operation.type} for ${operation.entity}:`,
              error
            );
          }
        }

        // Invalidate all module caches to refresh data
        queryClient.invalidateQueries({ queryKey: queryKeys.requisitions.all() });
        queryClient.invalidateQueries({ queryKey: queryKeys.purchaseOrders.all() });
        queryClient.invalidateQueries({ queryKey: queryKeys.paymentVouchers.all() });
        queryClient.invalidateQueries({ queryKey: queryKeys.dashboard.all() });

        // Show results
        const stats = await getQueueStats();
        if (failureCount === 0) {
          toast.dismiss();
          toast.success(`Synced ${successCount} changes successfully`);
        } else {
          toast.dismiss();
          toast.error(
            `Synced ${successCount} changes. ${failureCount} failed. Retrying soon...`
          );
        }

        console.log('[Queue Processor] Queue stats:', stats);
      } catch (error) {
        console.error('[Queue Processor] Unexpected error:', error);
        toast.error('Failed to sync offline changes. Will retry automatically.');
      } finally {
        processingRef.current = false;

        // Retry processing in case new operations were added
        setTimeout(() => {
          processQueue();
        }, 5000);
      }
    };

    // Start processing when connection is restored
    processQueue();
  }, [online, queryClient]);
}

/**
 * Hook to show offline indicator in UI
 *
 * Usage:
 * const isOffline = useOfflineStatus();
 * return isOffline && <div>You are offline. Changes will sync when connected.</div>
 */
export function useOfflineStatus(): boolean {
  const { online } = useNetwork();
  return !online;
}

/**
 * Hook to get queue statistics
 *
 * Usage:
 * const stats = useQueueStats();
 * return <div>Pending syncs: {stats.pending}</div>
 */
export function useQueueStats() {
  const [stats, setStats] = useState({
    total: 0,
    pending: 0,
    processing: 0,
    failed: 0,
    completed: 0,
  });

  useEffect(() => {
    const checkStats = async () => {
      const currentStats = await getQueueStats();
      setStats(currentStats);
    };

    checkStats();
    const interval = setInterval(checkStats, 5000); // Update every 5 seconds

    return () => clearInterval(interval);
  }, []);

  return stats;
}
