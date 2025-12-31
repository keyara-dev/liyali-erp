'use client';

import { useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { createUser, updateUser, deactivateUser, CreateUserRequest, UpdateUserRequest } from '@/app/_actions/users';
import { handleOfflineMutation, isOfflineResult } from '@/lib/offline-mutation-helper';

/**
 * Hook for creating a new user
 */
export function useCreateUserMutation() {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: async (data: CreateUserRequest) => {
      return await handleOfflineMutation(
        async () => {
          const response = await createUser(data);
          if (!response.success) {
            throw new Error(response.message);
          }
          return response.data;
        },
        {
          operation: 'CREATE',
          entity: 'user',
          data,
          successMessage: 'User created successfully',
          offlineMessage: 'User saved offline. Will sync when connected.',
        }
      );
    },
    onSuccess: (result) => {
      if (isOfflineResult(result)) {
        // Already handled by offline helper
      } else {
        toast.success('User created successfully');
      }
      // Invalidate users queries to refetch data
      queryClient.invalidateQueries({ queryKey: ['users'] });
    },
    onError: (error) => {
      console.error('Failed to create user:', error);
      toast.error(error?.message || 'Failed to create user');
    },
  });

  return {
    createUser: mutation.mutateAsync,
    isPending: mutation.isPending,
    error: mutation.error,
  };
}

/**
 * Hook for updating a user
 */
export function useUpdateUserMutation() {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: async (data: UpdateUserRequest) => {
      return await handleOfflineMutation(
        async () => {
          const response = await updateUser(data);
          if (!response.success) {
            throw new Error(response.message);
          }
          return response.data;
        },
        {
          operation: 'UPDATE',
          entity: 'user',
          data,
          entityId: data.id,
          successMessage: 'User updated successfully',
          offlineMessage: 'User changes saved offline. Will sync when connected.',
        }
      );
    },
    onSuccess: (result, variables) => {
      if (isOfflineResult(result)) {
        // Already handled by offline helper
      } else {
        toast.success('User updated successfully');
      }
      // Invalidate users queries and specific user query
      queryClient.invalidateQueries({ queryKey: ['users'] });
      queryClient.invalidateQueries({ queryKey: ['user', variables.id] });
    },
    onError: (error) => {
      console.error('Failed to update user:', error);
      toast.error(error?.message || 'Failed to update user');
    },
  });

  return {
    updateUser: mutation.mutateAsync,
    isPending: mutation.isPending,
    error: mutation.error,
  };
}

/**
 * Hook for deactivating a user
 */
export function useDeactivateUserMutation() {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: async (userId: string) => {
      return await handleOfflineMutation(
        async () => {
          const response = await deactivateUser(userId);
          if (!response.success) {
            throw new Error(response.message);
          }
          return response.data;
        },
        {
          operation: 'DELETE',
          entity: 'user',
          data: { userId },
          entityId: userId,
          successMessage: 'User deactivated successfully',
          offlineMessage: 'User deactivation saved offline. Will sync when connected.',
        }
      );
    },
    onSuccess: (result) => {
      if (isOfflineResult(result)) {
        // Already handled by offline helper
      } else {
        toast.success('User deactivated successfully');
      }
      // Invalidate users queries to refetch data
      queryClient.invalidateQueries({ queryKey: ['users'] });
    },
    onError: (error) => {
      console.error('Failed to deactivate user:', error);
      toast.error(error?.message || 'Failed to deactivate user');
    },
  });

  return {
    deactivateUser: mutation.mutateAsync,
    isPending: mutation.isPending,
    error: mutation.error,
  };
}