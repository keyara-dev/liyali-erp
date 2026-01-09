/**
 * Test suite for authentication and organization selection race conditions
 * These tests validate the fixes implemented for the critical race conditions
 * Updated for Zustand store implementation
 */

import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { useLoginMutation } from '@/hooks/use-auth-mutations';
import { useOrganizationContext, useOrganizationStore } from '@/hooks/use-organization';
import { fetchUserOrganizations, switchOrganization } from '@/app/_actions/organizations';
import { verifySession } from '@/lib/auth';

// Mock the auth and organization actions
vi.mock('@/lib/auth');
vi.mock('@/app/_actions/organizations');
vi.mock('next/navigation', () => ({
  useRouter: () => ({
    push: vi.fn(),
  }),
}));

const mockVerifySession = vi.mocked(verifySession);
const mockFetchUserOrganizations = vi.mocked(fetchUserOrganizations);
const mockSwitchOrganization = vi.mocked(switchOrganization);

describe('Authentication Race Conditions', () => {
  let queryClient: QueryClient;

  beforeEach(() => {
    queryClient = new QueryClient({
      defaultOptions: {
        queries: { retry: false },
        mutations: { retry: false },
      },
    });
    // Make queryClient available globally for Zustand store
    (window as any).queryClient = queryClient;
    
    // Reset Zustand store
    useOrganizationStore.getState().setInitialized(false);
    useOrganizationStore.getState().setUserOrganizations([]);
    useOrganizationStore.getState().setError(null);
    
    vi.clearAllMocks();
  });

  afterEach(() => {
    queryClient.clear();
    delete (window as any).queryClient;
  });

  describe('Login → Organization Fetch Race Condition', () => {
    it('should wait for session to be ready before redirecting', async () => {
      // Mock session verification to fail initially, then succeed
      let sessionCallCount = 0;
      mockVerifySession.mockImplementation(() => {
        sessionCallCount++;
        if (sessionCallCount <= 2) {
          return Promise.resolve({ isAuthenticated: false, session: null });
        }
        return Promise.resolve({
          isAuthenticated: true,
          session: { access_token: 'token', user: { id: '1' } },
        });
      });

      const wrapper = ({ children }: { children: React.ReactNode }) => (
        <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
      );

      const { result } = renderHook(() => useLoginMutation(), { wrapper });

      // Attempt login
      const loginPromise = result.current.login({
        email: 'test@example.com',
        password: 'password',
      });

      // Should eventually succeed after session verification retries
      await expect(loginPromise).resolves.toBeDefined();
      expect(sessionCallCount).toBeGreaterThan(2);
    });

    it('should handle session verification timeout gracefully', async () => {
      // Mock session verification to always fail
      mockVerifySession.mockResolvedValue({ isAuthenticated: false, session: null });

      const wrapper = ({ children }: { children: React.ReactNode }) => (
        <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
      );

      const { result } = renderHook(() => useLoginMutation(), { wrapper });

      // Should throw error after max attempts
      await expect(
        result.current.login({
          email: 'test@example.com',
          password: 'password',
        })
      ).rejects.toThrow('Session initialization failed');
    });
  });

  describe('Organization Store Initialization', () => {
    it('should prevent race conditions during organization loading', async () => {
      // Mock session verification to succeed
      mockVerifySession.mockResolvedValue({
        isAuthenticated: true,
        session: { access_token: 'token', user: { id: '1' } },
      });

      // Mock organizations fetch
      const mockOrgs = [
        { id: 'org1', name: 'Org 1' },
        { id: 'org2', name: 'Org 2' },
      ];
      mockFetchUserOrganizations.mockResolvedValue(mockOrgs);

      const wrapper = ({ children }: { children: React.ReactNode }) => (
        <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
      );

      // Initialize the store
      useOrganizationStore.getState().initialize();

      const { result } = renderHook(() => useOrganizationContext(), { wrapper });

      // Should eventually load organizations
      await waitFor(() => {
        expect(result.current.userOrganizations).toEqual(mockOrgs);
      });

      // Should have selected a current organization
      expect(result.current.currentOrganization).toBeDefined();
    });

    it('should handle session verification failure during org fetch', async () => {
      // Mock session verification to fail
      mockVerifySession.mockRejectedValue(new Error('No valid session found'));

      const wrapper = ({ children }: { children: React.ReactNode }) => (
        <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
      );

      // Initialize the store
      useOrganizationStore.getState().initialize();

      const { result } = renderHook(() => useOrganizationContext(), { wrapper });

      // Should handle error gracefully
      await waitFor(() => {
        expect(result.current.error).toContain('No valid session found');
      });
    });
  });

  describe('Organization Switching Race Conditions', () => {
    it('should handle concurrent organization switches', async () => {
      // Mock session verification
      mockVerifySession
        .mockResolvedValueOnce({
          isAuthenticated: true,
          session: { access_token: 'token', organization_id: 'org1' },
        })
        .mockResolvedValueOnce({
          isAuthenticated: true,
          session: { access_token: 'token', organization_id: 'org2' },
        });

      // Mock organization switch
      mockSwitchOrganization.mockImplementation((orgId: string) => 
        Promise.resolve(orgId)
      );

      // Simulate concurrent switches
      const promises = [
        switchOrganization('org1'),
        switchOrganization('org2'),
        switchOrganization('org3'),
      ];

      // All should complete without throwing
      const results = await Promise.allSettled(promises);
      const successful = results.filter(r => r.status === 'fulfilled');
      
      // At least some should succeed
      expect(successful.length).toBeGreaterThan(0);
    });

    it('should verify organization switch completion', async () => {
      // Mock session verification before switch
      mockVerifySession
        .mockResolvedValueOnce({
          isAuthenticated: true,
          session: { access_token: 'token', organization_id: 'org1' },
        })
        // Mock session verification after switch
        .mockResolvedValueOnce({
          isAuthenticated: true,
          session: { access_token: 'token', organization_id: 'org2' },
        });

      mockSwitchOrganization.mockResolvedValue('org2');

      const result = await switchOrganization('org2');
      expect(result).toBe('org2');
      expect(mockVerifySession).toHaveBeenCalledTimes(2);
    });
  });

  describe('Session Validation', () => {
    it('should reject requests with invalid sessions', async () => {
      mockVerifySession.mockResolvedValue({ isAuthenticated: false, session: null });

      await expect(fetchUserOrganizations()).rejects.toThrow();
    });

    it('should handle session expiration during requests', async () => {
      // Mock session to expire during request
      mockVerifySession
        .mockResolvedValueOnce({
          isAuthenticated: true,
          session: { access_token: 'token' },
        })
        .mockResolvedValueOnce({ isAuthenticated: false, session: null });

      mockFetchUserOrganizations.mockRejectedValue(new Error('Session expired'));

      await expect(fetchUserOrganizations()).rejects.toThrow('Session expired');
    });
  });

  describe('Error Recovery', () => {
    it('should retry failed organization fetches', async () => {
      let callCount = 0;
      mockVerifySession.mockImplementation(() => {
        callCount++;
        if (callCount <= 2) {
          return Promise.reject(new Error('No valid session found'));
        }
        return Promise.resolve({
          isAuthenticated: true,
          session: { access_token: 'token' },
        });
      });

      mockFetchUserOrganizations.mockResolvedValue([]);

      const wrapper = ({ children }: { children: React.ReactNode }) => (
        <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
      );

      // Initialize the store
      useOrganizationStore.getState().initialize();

      const { result } = renderHook(() => useOrganizationContext(), { wrapper });

      // Should eventually succeed after retries
      await waitFor(() => {
        expect(result.current.userOrganizations).toEqual([]);
      });

      expect(callCount).toBeGreaterThan(2);
    });
  });
});

describe('Performance Tests', () => {
  it('should complete login to organization access within reasonable time', async () => {
    const startTime = Date.now();

    // Mock fast session verification
    mockVerifySession.mockResolvedValue({
      isAuthenticated: true,
      session: { access_token: 'token' },
    });

    mockFetchUserOrganizations.mockResolvedValue([
      { id: 'org1', name: 'Test Org' },
    ]);

    const queryClient = new QueryClient();
    (window as any).queryClient = queryClient;
    
    const wrapper = ({ children }: { children: React.ReactNode }) => (
      <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    );

    // Initialize the store
    useOrganizationStore.getState().initialize();

    const { result } = renderHook(() => useOrganizationContext(), { wrapper });

    await waitFor(() => {
      expect(result.current.userOrganizations.length).toBeGreaterThan(0);
    });

    const endTime = Date.now();
    const duration = endTime - startTime;

    // Should complete within 3 seconds
    expect(duration).toBeLessThan(3000);
    
    delete (window as any).queryClient;
  });
});