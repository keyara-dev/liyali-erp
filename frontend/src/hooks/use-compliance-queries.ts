'use client';

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { QUERY_KEYS } from '@/lib/constants';
import { toast } from 'sonner';

export interface ComplianceItem {
  id: string;
  name: string;
  requirement: string;
  status: 'compliant' | 'non-compliant' | 'pending';
  dueDate: string;
  completionDate?: string;
  evidence: string[];
  responsible: string;
}

export interface ComplianceTrackingData {
  requirements: ComplianceItem[];
  totalRequirements: number;
  compliantCount: number;
  nonCompliantCount: number;
  pendingCount: number;
  complianceScore: number;
}

/**
 * Fetch all compliance requirements and tracking data
 * @param onSuccess - Optional callback on success
 * @returns Query result with compliance tracking data
 */
export const useComplianceRequirements = (
  onSuccess?: (data: ComplianceTrackingData) => void
) =>
  useQuery({
    queryKey: [QUERY_KEYS.COMPLIANCE.ALL],
    queryFn: async () => {
      const response = await fetch('/api/compliance/requirements');
      if (!response.ok) throw new Error('Failed to fetch compliance requirements');
      const data = response.json();
      if (onSuccess) onSuccess(data);
      return data;
    },
    staleTime: 10 * 60 * 1000, // 10 minutes
  });

/**
 * Update compliance item status
 * @param onSuccess - Optional callback on success
 * @returns Mutation object with mutateAsync, isPending, error
 */
export const useUpdateComplianceStatus = (
  onSuccess?: (data: ComplianceItem) => void
) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (payload: { itemId: string; status: 'compliant' | 'non-compliant' | 'pending'; completionDate?: string; evidence?: string[] }) => {
      const response = await fetch(`/api/compliance/requirements/${payload.itemId}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          status: payload.status,
          completionDate: payload.completionDate,
          evidence: payload.evidence,
        }),
      });
      if (!response.ok) throw new Error('Failed to update compliance status');
      return response.json();
    },
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.COMPLIANCE?.ALL || 'compliance-all'] });
      toast.success('Compliance status updated successfully');
      if (onSuccess) onSuccess(data);
    },
    onError: (error: any) => {
      toast.error(error.message || 'Failed to update compliance status');
    },
  });
};

/**
 * Add evidence to a compliance item
 * @param onSuccess - Optional callback on success
 * @returns Mutation object with mutateAsync, isPending, error
 */
export const useAddComplianceEvidence = (
  onSuccess?: (data: ComplianceItem) => void
) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (payload: { itemId: string; evidence: string[] }) => {
      const response = await fetch(`/api/compliance/requirements/${payload.itemId}/evidence`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ evidence: payload.evidence }),
      });
      if (!response.ok) throw new Error('Failed to add evidence');
      return response.json();
    },
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.COMPLIANCE?.ALL || 'compliance-all'] });
      toast.success('Evidence added successfully');
      if (onSuccess) onSuccess(data);
    },
    onError: (error: any) => {
      toast.error(error.message || 'Failed to add evidence');
    },
  });
};
