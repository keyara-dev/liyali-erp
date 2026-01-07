'use client';

import { useState } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useRouter } from 'next/navigation';
import { toast } from 'sonner';
import { useOrganizationContext } from '@/contexts/organization-context';
import { logoutAction } from '@/app/_actions/auth';
import { 
  switchOrganization, 
  createOrganization, 
  updateOrganization, 
  addOrganizationMember, 
  removeOrganizationMember,
  updateOrganizationSettings,
  CreateOrganizationRequest,
  UpdateOrganizationRequest,
  AddMemberRequest,
  OrganizationSettings
} from '@/app/_actions/organizations';
import { handleOfflineMutation, isOfflineResult } from '@/lib/offline-mutation-helper';

/**
 * Hook for handling organization selection/switching
 * Manages the flow of switching organizations and navigating to home
 */
export function useSelectOrganization() {
  const router = useRouter();
  const { switchWorkspace } = useOrganizationContext();
  const [isRedirecting, setIsRedirecting] = useState(false);

  const mutation = useMutation({
    mutationFn: async (orgId: string) => {
      await switchWorkspace(orgId);
    },
    onSuccess: () => {
      setIsRedirecting(true);
      router.push('/home');
    },
    onError: (error) => {
      console.error('Failed to switch organization:', error);
      setIsRedirecting(false);
    },
  });

  return {
    selectOrganization: mutation.mutateAsync,
    isPending: mutation.isPending || isRedirecting,
    error: mutation.error,
  };
}

/**
 * Hook for switching organizations using backend API
 */
export function useSwitchOrganizationMutation() {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: async (orgId: string) => {
      return await switchOrganization(orgId);
    },
    onSuccess: () => {
      // Invalidate all queries to refetch with new organization context
      queryClient.invalidateQueries();
    },
    onError: (error) => {
      console.error('Failed to switch organization:', error);
    },
  });

  return {
    switchOrganization: mutation.mutateAsync,
    isPending: mutation.isPending,
    error: mutation.error,
  };
}

/**
 * Hook for creating an organization
 */
export function useCreateOrganizationMutation() {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: async (data: CreateOrganizationRequest) => {
      return await handleOfflineMutation(
        async () => {
          const response = await createOrganization(data);
          if (!response.success) {
            throw new Error(response.message);
          }
          return response.data;
        },
        {
          operation: 'CREATE',
          entity: 'organization',
          data,
          successMessage: 'Organization created successfully',
          offlineMessage: 'Organization saved offline. Will sync when connected.',
        }
      );
    },
    onSuccess: (result) => {
      if (isOfflineResult(result)) {
        // Already handled by offline helper
      } else {
        toast.success('Organization created successfully');
      }
      queryClient.invalidateQueries({ queryKey: ['organizations'] });
    },
    onError: (error) => {
      console.error('Failed to create organization:', error);
      toast.error(error?.message || 'Failed to create organization');
    },
  });

  return {
    createOrganization: mutation.mutateAsync,
    isPending: mutation.isPending,
    error: mutation.error,
  };
}

/**
 * Hook for updating an organization
 */
export function useUpdateOrganizationMutation() {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: async (data: UpdateOrganizationRequest) => {
      return await handleOfflineMutation(
        async () => {
          const response = await updateOrganization(data);
          if (!response.success) {
            throw new Error(response.message);
          }
          return response.data;
        },
        {
          operation: 'UPDATE',
          entity: 'organization',
          data,
          entityId: data.id,
          successMessage: 'Organization updated successfully',
          offlineMessage: 'Organization changes saved offline. Will sync when connected.',
        }
      );
    },
    onSuccess: (result, variables) => {
      if (isOfflineResult(result)) {
        // Already handled by offline helper
      } else {
        toast.success('Organization updated successfully');
      }
      queryClient.invalidateQueries({ queryKey: ['organizations'] });
      queryClient.invalidateQueries({ queryKey: ['organization', variables.id] });
    },
    onError: (error) => {
      console.error('Failed to update organization:', error);
    },
  });

  return {
    updateOrganization: mutation.mutateAsync,
    isPending: mutation.isPending,
    error: mutation.error,
  };
}

/**
 * Hook for adding organization member
 */
export function useAddMemberMutation() {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: async (data: AddMemberRequest) => {
      const response = await addOrganizationMember(data);
      if (!response.success) {
        throw new Error(response.message);
      }
      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['organization-members'] });
    },
    onError: (error) => {
      console.error('Failed to add member:', error);
    },
  });

  return {
    addMember: mutation.mutateAsync,
    isPending: mutation.isPending,
    error: mutation.error,
  };
}

/**
 * Hook for removing organization member
 */
export function useRemoveMemberMutation() {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: async (userId: string) => {
      const response = await removeOrganizationMember(userId);
      if (!response.success) {
        throw new Error(response.message);
      }
      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['organization-members'] });
    },
    onError: (error) => {
      console.error('Failed to remove member:', error);
    },
  });

  return {
    removeMember: mutation.mutateAsync,
    isPending: mutation.isPending,
    error: mutation.error,
  };
}

/**
 * Hook for updating organization settings
 */
export function useUpdateSettingsMutation() {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: async (data: OrganizationSettings) => {
      const response = await updateOrganizationSettings(data);
      if (!response.success) {
        throw new Error(response.message);
      }
      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['organization-settings'] });
    },
    onError: (error) => {
      console.error('Failed to update settings:', error);
    },
  });

  return {
    updateSettings: mutation.mutateAsync,
    isPending: mutation.isPending,
    error: mutation.error,
  };
}

/**
 * Hook for handling user logout
 */
export function useLogout() {
  const router = useRouter();

  const mutation = useMutation({
    mutationFn: async () => {
      await logoutAction();
    },
    onSuccess: () => {
      router.push('/login');
    },
    onError: (error) => {
      console.error('Logout failed:', error);
    },
  });

  return {
    logout: mutation.mutateAsync,
    isPending: mutation.isPending,
    error: mutation.error,
  };
}
