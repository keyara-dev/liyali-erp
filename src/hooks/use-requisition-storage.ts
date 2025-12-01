'use client';

import { useEffect, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Requisition } from '@/types/requisition';
import { WorkflowDocument } from '@/types/workflow';
import { getRequisitions } from '@/app/_actions/requisitions';
import { QUERY_KEYS } from '@/lib/constants';

const REQUISITIONS_STORAGE_KEY = 'liyali_requisitions';

// ============================================================================
// STORAGE UTILITIES
// ============================================================================

/**
 * Load all requisitions from localStorage
 */
function loadRequisitionsFromStorage(): Requisition[] {
  try {
    if (typeof window === 'undefined') return [];
    const stored = localStorage.getItem(REQUISITIONS_STORAGE_KEY);
    if (!stored) return [];
    const parsed = JSON.parse(stored);
    return Array.isArray(parsed) ? parsed : [];
  } catch (error) {
    console.error('Failed to load requisitions from storage:', error);
    return [];
  }
}

/**
 * Save requisition to localStorage
 */
function saveRequisitionToStorage(requisition: Requisition): void {
  try {
    if (typeof window === 'undefined') return;
    const requisitions = loadRequisitionsFromStorage();
    const index = requisitions.findIndex(r => r.id === requisition.id);
    if (index >= 0) {
      requisitions[index] = requisition;
    } else {
      requisitions.push(requisition);
    }
    localStorage.setItem(REQUISITIONS_STORAGE_KEY, JSON.stringify(requisitions));
  } catch (error) {
    console.error('Failed to save requisition to storage:', error);
  }
}

/**
 * Save multiple requisitions to localStorage
 */
function saveRequisitionsToStorage(requisitions: Requisition[]): void {
  try {
    if (typeof window === 'undefined') return;
    localStorage.setItem(REQUISITIONS_STORAGE_KEY, JSON.stringify(requisitions));
  } catch (error) {
    console.error('Failed to save requisitions to storage:', error);
  }
}

/**
 * Get a specific requisition by ID from localStorage
 */
function getRequisitionFromStorage(requisitionId: string): Requisition | null {
  try {
    if (typeof window === 'undefined') return null;
    const requisitions = loadRequisitionsFromStorage();
    return requisitions.find(r => r.id === requisitionId) || null;
  } catch (error) {
    console.error('Failed to get requisition from storage:', error);
    return null;
  }
}

/**
 * Delete a requisition from localStorage
 */
function deleteRequisitionFromStorage(requisitionId: string): void {
  try {
    if (typeof window === 'undefined') return;
    const requisitions = loadRequisitionsFromStorage();
    const filtered = requisitions.filter(r => r.id !== requisitionId);
    localStorage.setItem(REQUISITIONS_STORAGE_KEY, JSON.stringify(filtered));
  } catch (error) {
    console.error('Failed to delete requisition from storage:', error);
  }
}

/**
 * Clear all requisitions from localStorage
 */
function clearRequisitionsStorage(): void {
  try {
    if (typeof window === 'undefined') return;
    localStorage.removeItem(REQUISITIONS_STORAGE_KEY);
  } catch (error) {
    console.error('Failed to clear requisitions storage:', error);
  }
}

// ============================================================================
// DATA CONVERSION
// ============================================================================

/**
 * Convert a Requisition to a WorkflowDocument for display in tables
 */
function requisitionToWorkflowDocument(requisition: Requisition): WorkflowDocument {
  return {
    id: requisition.id,
    type: 'REQUISITION',
    documentNumber: requisition.requisitionNumber,
    status: requisition.status as any,
    currentStage: requisition.currentApprovalStage || 1,
    createdBy: requisition.requestedBy,
    createdAt: requisition.requestedDate instanceof Date
      ? requisition.requestedDate
      : new Date(requisition.requestedDate),
    updatedAt: new Date(),
    metadata: {
      title: requisition.title,
      description: requisition.description,
      department: requisition.department,
      requestedFor: requisition.title,
      totalAmount: requisition.totalAmount,
      amount: requisition.totalAmount,
      priority: requisition.priority,
      itemCount: requisition.items?.length || 0,
    },
  };
}

/**
 * Public export of conversion function for use in components
 */
export function convertRequisitionToWorkflowDocument(requisition: Requisition): WorkflowDocument {
  return requisitionToWorkflowDocument(requisition);
}

// ============================================================================
// REACT HOOKS
// ============================================================================

/**
 * Hook to manage requisition data with localStorage persistence
 * Syncs data between server state and browser localStorage
 *
 * @returns Object with hydration state and storage functions
 *
 * @example
 * const { isHydrated, loadFromStorage, saveToStorage } = useRequisitionStorage()
 */
export function useRequisitionStorage() {
  const [isHydrated, setIsHydrated] = useState(false);

  useEffect(() => {
    setIsHydrated(true);
  }, []);

  return {
    isHydrated,
    loadFromStorage: loadRequisitionsFromStorage,
    loadOneFromStorage: getRequisitionFromStorage,
    saveToStorage: saveRequisitionToStorage,
    saveMultiple: saveRequisitionsToStorage,
    deleteFromStorage: deleteRequisitionFromStorage,
    clearStorage: clearRequisitionsStorage,
  };
}

/**
 * React Query hook for fetching all requisitions with localStorage fallback
 * Merges API data with localStorage data for complete view
 *
 * @param includeStorageData - Whether to include localStorage data (default: true)
 * @returns Query result with merged requisitions array
 *
 * @example
 * const { data: requisitions, isLoading } = useRequisitionsWithStorage()
 */
export const useRequisitionsWithStorage = (includeStorageData = true) =>
  useQuery({
    queryKey: [QUERY_KEYS.REQUISITIONS.ALL, 'with-storage'],
    queryFn: async () => {
      let allRequisitions: Requisition[] = [];

      // Load from API
      try {
        const response = await getRequisitions();
        if (response.success && response.data) {
          allRequisitions = response.data;
        }
      } catch (error) {
        console.error('Failed to fetch requisitions from API:', error);
      }

      // Also load from localStorage
      if (includeStorageData && typeof window !== 'undefined') {
        try {
          const storedRequisitions = loadRequisitionsFromStorage();
          if (storedRequisitions && storedRequisitions.length > 0) {
            // Merge: prioritize API data, add missing from localStorage
            const apiIds = new Set(allRequisitions.map(r => r.id));
            const localOnlyRequisitions = storedRequisitions.filter(
              r => !apiIds.has(r.id)
            );

            allRequisitions = [...allRequisitions, ...localOnlyRequisitions];
          }
        } catch (storageError) {
          console.error('Failed to load requisitions from storage:', storageError);
        }
      }

      return allRequisitions;
    },
    staleTime: 5 * 60 * 1000, // 5 minutes
    gcTime: 10 * 60 * 1000, // 10 minutes (formerly cacheTime)
  });

/**
 * React Query hook for fetching a specific requisition with localStorage fallback
 *
 * @param requisitionId - ID of the requisition to fetch
 * @param initialData - Optional initial data from server component
 * @returns Query result with single requisition
 *
 * @example
 * const { data: requisition } = useRequisitionWithStorage(requisitionId)
 */
export const useRequisitionWithStorage = (
  requisitionId: string,
  initialData?: Requisition
) =>
  useQuery({
    queryKey: [QUERY_KEYS.REQUISITIONS.BY_ID, requisitionId, 'with-storage'],
    queryFn: async () => {
      // First try localStorage
      if (typeof window !== 'undefined') {
        const stored = getRequisitionFromStorage(requisitionId);
        if (stored) {
          return stored;
        }
      }

      // Then try API (will be handled by existing hook)
      return null;
    },
    initialData,
    enabled: !!requisitionId,
    staleTime: 5 * 60 * 1000,
    gcTime: 10 * 60 * 1000,
  });

/**
 * Hook to convert requisitions to workflow documents for table display
 * Combines localStorage and API requisitions with format conversion
 *
 * @param includeStorageData - Whether to include localStorage data (default: true)
 * @returns Query result with WorkflowDocument array
 *
 * @example
 * const { data: documents } = useRequisitionsAsWorkflowDocuments()
 */
export const useRequisitionsAsWorkflowDocuments = (includeStorageData = true) =>
  useQuery({
    queryKey: [QUERY_KEYS.REQUISITIONS.ALL, 'as-documents'],
    queryFn: async () => {
      let allRequisitions: Requisition[] = [];

      // Load from API
      try {
        const response = await getRequisitions();
        if (response.success && response.data) {
          allRequisitions = response.data;
        }
      } catch (error) {
        console.error('Failed to fetch requisitions from API:', error);
      }

      // Also load from localStorage
      if (includeStorageData && typeof window !== 'undefined') {
        try {
          const storedRequisitions = loadRequisitionsFromStorage();
          if (storedRequisitions && storedRequisitions.length > 0) {
            // Merge: prioritize API data, add missing from localStorage
            const apiIds = new Set(allRequisitions.map(r => r.id));
            const localOnlyRequisitions = storedRequisitions.filter(
              r => !apiIds.has(r.id)
            );

            allRequisitions = [...allRequisitions, ...localOnlyRequisitions];
          }
        } catch (storageError) {
          console.error('Failed to load requisitions from storage:', storageError);
        }
      }

      return allRequisitions.map(requisitionToWorkflowDocument);
    },
    staleTime: 5 * 60 * 1000,
    gcTime: 10 * 60 * 1000,
  });
