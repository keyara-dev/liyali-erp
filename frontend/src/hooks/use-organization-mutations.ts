'use client';

import { useMutation } from '@tanstack/react-query';
import { useRouter } from 'next/navigation';
import { useOrganizationContext } from '@/contexts/organization-context';
import { logoutAction } from '@/app/_actions/auth';

/**
 * Hook for handling organization selection/switching
 * Manages the flow of switching organizations and navigating to home
 *
 * @returns {Object} Object with selectOrganization mutation and isLoading state
 *
 * @example
 * ```typescript
 * const { selectOrganization, isPending } = useSelectOrganization();
 *
 * const handleClick = async () => {
 *   await selectOrganization(orgId);
 * };
 * ```
 */
export function useSelectOrganization() {
  const router = useRouter();
  const { switchWorkspace } = useOrganizationContext();

  const mutation = useMutation({
    mutationFn: async (orgId: string) => {
      await switchWorkspace(orgId);
    },
    onSuccess: () => {
      router.push('/home');
    },
    onError: (error) => {
      console.error('Failed to switch organization:', error);
    },
  });

  return {
    selectOrganization: mutation.mutateAsync,
    isPending: mutation.isPending,
    error: mutation.error,
  };
}

/**
 * Hook for handling user logout
 * Manages the flow of clearing session and redirecting to login
 *
 * @returns {Object} Object with logout mutation and isLoading state
 *
 * @example
 * ```typescript
 * const { logout, isPending } = useLogout();
 *
 * const handleLogout = async () => {
 *   await logout();
 * };
 * ```
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
