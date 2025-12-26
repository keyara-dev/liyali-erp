'use client';

import { useQuery } from '@tanstack/react-query';
import { QUERY_KEYS } from '@/lib/constants';
import { getGRNAction } from '@/app/_actions/grn-actions';

/**
 * GRN Data Types
 */
interface QualityIssue {
  id: string;
  itemId: string;
  description: string;
  severity: 'LOW' | 'MEDIUM' | 'HIGH';
}

interface GRNItem {
  id: string;
  itemNumber: number;
  description: string;
  poQuantity: number;
  receivedQuantity: number;
  unit: string;
  variance: number;
  damage: number;
  damageNotes?: string;
  condition: 'GOOD' | 'DAMAGED' | 'PARTIAL';
}

interface GoodsReceivedNote {
  id: string;
  grnNumber: string;
  poNumber: string;
  status: 'DRAFT' | 'SUBMITTED' | 'CONFIRMED' | 'REJECTED';
  warehouseLocation: string;
  receivedDate: string;
  receivedBy: string;
  approvedBy?: string;
  items: GRNItem[];
  qualityIssues: QualityIssue[];
  notes?: string;
  currentStage: number;
  stageName: string;
  createdAt: string;
  updatedAt: string;
}

export type { GoodsReceivedNote, QualityIssue, GRNItem };

/**
 * Fetch a specific GRN by ID
 * Uses React Query with 5-minute cache
 *
 * @param grnId - The GRN ID to fetch
 * @param enabled - Whether the query should run (default: true)
 * @returns Query result with single GRN
 *
 * @example
 * ```typescript
 * const { data: grn, isLoading, error } = useGRNById(grnId)
 * ```
 */
export const useGRNById = (grnId: string, enabled = true) =>
  useQuery({
    queryKey: [QUERY_KEYS.GRN.BY_ID, grnId],
    queryFn: async () => {
      try {
        const response = await getGRNAction(grnId);
        return response;
      } catch (error) {
        console.error('Error fetching GRN:', error);
        throw error;
      }
    },
    staleTime: 5 * 60 * 1000, // 5 minutes
    gcTime: 10 * 60 * 1000, // 10 minutes (formerly cacheTime)
    enabled: enabled && !!grnId,
  });

/**
 * Note: Full GRN list query (useGRNs) should be implemented
 * via server actions that call the backend API.
 * For now, individual GRN fetching is the primary hook.
 *
 * TODO: Implement useGRNs() for listing all GRNs with pagination
 * when the backend GRN listing API is fully integrated.
 */
