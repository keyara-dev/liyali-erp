'use client';

import { useEffect } from 'react';
import { initializeStorage } from '@/lib/storage';
import { initializeWorkflowStores } from '@/lib/workflow-stores';

/**
 * Hook to initialize all data stores on app startup
 * Initializes both localStorage and in-memory workflow stores
 * Should be called once in a root layout or context provider
 *
 * When backend APIs are ready:
 * 1. Remove this hook
 * 2. Remove the initialization call from providers
 * 3. Delete the /lib/storage folder
 */
export function useInitializeStorage(): void {
  useEffect(() => {
    // Initialize in-memory workflow stores
    initializeWorkflowStores();

    // Initialize localStorage with seed data
    initializeStorage();
  }, []);
}
