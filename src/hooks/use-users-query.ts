'use client'

import { useQuery } from '@tanstack/react-query'

export interface User {
  id: string
  name: string
  email: string
  role: string
  department?: string
  avatar?: string
}

/**
 * Hook to fetch all users
 */
export function useUsers() {
  return useQuery({
    queryKey: ['users'],
    queryFn: async () => {
      // Mock implementation - replace with actual API call
      return Promise.resolve([
        {
          id: 'user-001',
          name: 'John Requester',
          email: 'requester@liyali.com',
          role: 'REQUESTER',
          department: 'Operations'
        },
        {
          id: 'user-002',
          name: 'Sarah Manager',
          email: 'manager@liyali.com',
          role: 'DEPARTMENT_MANAGER',
          department: 'Finance'
        },
        {
          id: 'user-003',
          name: 'James Finance',
          email: 'finance@liyali.com',
          role: 'FINANCE_OFFICER',
          department: 'Finance'
        }
      ] as User[])
    }
  })
}

/**
 * Hook to fetch a single user by ID
 */
export function useUser(userId: string) {
  return useQuery({
    queryKey: ['user', userId],
    queryFn: async () => {
      // Mock implementation - replace with actual API call
      const users = [
        {
          id: 'user-001',
          name: 'John Requester',
          email: 'requester@liyali.com',
          role: 'REQUESTER',
          department: 'Operations'
        },
        {
          id: 'user-002',
          name: 'Sarah Manager',
          email: 'manager@liyali.com',
          role: 'DEPARTMENT_MANAGER',
          department: 'Finance'
        }
      ] as User[]

      return users.find(u => u.id === userId) || null
    },
    enabled: !!userId
  })
}

/**
 * Alias for useUsers - commonly used name
 */
export const useGetUsers = useUsers
