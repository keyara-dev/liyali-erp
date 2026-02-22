import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  getAllUsers,
  getUserById,
  getUserStatistics,
  updateUser,
  updateUserStatus,
  getUserActivity,
  getUserSessions,
  terminateUserSession,
  terminateAllUserSessions,
  resetUserPassword,
  getUserOrganizations,
  updateUserOrganizationRole,
  removeUserFromOrganization,
  type UserFilters,
  type UpdateUserRequest,
} from "@/app/_actions/users";

// --- Query Hooks ---

export function useUsers(filters?: UserFilters) {
  return useQuery({
    queryKey: ["users", filters],
    queryFn: async () => {
      const result = await getAllUsers(filters);
      if (!result.success) throw new Error(result.message);
      return result.data;
    },
  });
}

export function useUser(id: string) {
  return useQuery({
    queryKey: ["users", id],
    queryFn: async () => {
      const result = await getUserById(id);
      if (!result.success) throw new Error(result.message);
      return result.data;
    },
    enabled: !!id,
  });
}

export function useUserStats() {
  return useQuery({
    queryKey: ["users", "statistics"],
    queryFn: async () => {
      const result = await getUserStatistics();
      if (!result.success) throw new Error(result.message);
      return result.data;
    },
  });
}

export function useUserActivity(userId: string, page = 1, limit = 50) {
  return useQuery({
    queryKey: ["users", userId, "activity", { page, limit }],
    queryFn: async () => {
      const result = await getUserActivity(userId, page, limit);
      if (!result.success) throw new Error(result.message);
      return result.data;
    },
    enabled: !!userId,
  });
}

export function useUserSessions(userId: string) {
  return useQuery({
    queryKey: ["users", userId, "sessions"],
    queryFn: async () => {
      const result = await getUserSessions(userId);
      if (!result.success) throw new Error(result.message);
      return result.data;
    },
    enabled: !!userId,
  });
}

export function useUserOrganizations(userId: string) {
  return useQuery({
    queryKey: ["users", userId, "organizations"],
    queryFn: async () => {
      const result = await getUserOrganizations(userId);
      if (!result.success) throw new Error(result.message);
      return result.data;
    },
    enabled: !!userId,
  });
}

// --- Mutation Hooks ---

export function useUpdateUser() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateUserRequest }) =>
      updateUser(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["users"] });
    },
  });
}

export function useUpdateUserStatus() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({
      id,
      status,
      reason,
    }: {
      id: string;
      status: "active" | "suspended" | "inactive";
      reason?: string;
    }) => updateUserStatus(id, status, reason),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["users"] });
    },
  });
}

export function useResetUserPassword() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({
      id,
      sendEmail,
    }: {
      id: string;
      sendEmail?: boolean;
    }) => resetUserPassword(id, sendEmail),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["users"] });
    },
  });
}

export function useTerminateUserSession() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({
      userId,
      sessionId,
    }: {
      userId: string;
      sessionId: string;
    }) => terminateUserSession(userId, sessionId),
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({
        queryKey: ["users", variables.userId, "sessions"],
      });
    },
  });
}

export function useTerminateAllUserSessions() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (userId: string) => terminateAllUserSessions(userId),
    onSuccess: (_data, userId) => {
      queryClient.invalidateQueries({
        queryKey: ["users", userId, "sessions"],
      });
    },
  });
}

export function useUpdateUserOrgRole() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({
      userId,
      organizationId,
      role,
      status,
    }: {
      userId: string;
      organizationId: string;
      role: string;
      status?: "active" | "suspended";
    }) => updateUserOrganizationRole(userId, organizationId, role, status),
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({
        queryKey: ["users", variables.userId, "organizations"],
      });
      queryClient.invalidateQueries({ queryKey: ["users"] });
    },
  });
}

export function useRemoveUserFromOrg() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({
      userId,
      organizationId,
    }: {
      userId: string;
      organizationId: string;
    }) => removeUserFromOrganization(userId, organizationId),
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({
        queryKey: ["users", variables.userId, "organizations"],
      });
      queryClient.invalidateQueries({ queryKey: ["users"] });
    },
  });
}
