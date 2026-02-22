import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  getRoles,
  getRole,
  getRoleStats,
  createRole,
  updateRole,
  deleteRole,
  getPermissions,
  getPermissionsByCategory,
  getRoleUsers,
  assignRoleToUsers,
  removeRoleFromUsers,
  cloneRole,
  getRoleAuditHistory,
  type RoleFilters,
  type CreateRoleRequest,
  type UpdateRoleRequest,
} from "@/app/_actions/roles";

// --- Query Hooks ---

export function useRoles(filters?: RoleFilters) {
  return useQuery({
    queryKey: ["roles", filters],
    queryFn: async () => {
      const result = await getRoles(filters);
      if (!result.success) throw new Error(result.message);
      return result.data;
    },
  });
}

export function useRole(id: string) {
  return useQuery({
    queryKey: ["roles", id],
    queryFn: async () => {
      const result = await getRole(id);
      if (!result.success) throw new Error(result.message);
      return result.data;
    },
    enabled: !!id,
  });
}

export function useRoleStats() {
  return useQuery({
    queryKey: ["roles", "stats"],
    queryFn: async () => {
      const result = await getRoleStats();
      if (!result.success) throw new Error(result.message);
      return result.data;
    },
  });
}

export function usePermissions() {
  return useQuery({
    queryKey: ["permissions"],
    queryFn: async () => {
      const result = await getPermissions();
      if (!result.success) throw new Error(result.message);
      return result.data;
    },
    staleTime: 5 * 60 * 1000, // permissions rarely change
  });
}

export function usePermissionsByCategory() {
  return useQuery({
    queryKey: ["permissions", "by-category"],
    queryFn: async () => {
      const result = await getPermissionsByCategory();
      if (!result.success) throw new Error(result.message);
      return result.data;
    },
    staleTime: 5 * 60 * 1000,
  });
}

export function useRoleUsers(roleId: string) {
  return useQuery({
    queryKey: ["roles", roleId, "users"],
    queryFn: async () => {
      const result = await getRoleUsers(roleId);
      if (!result.success) throw new Error(result.message);
      return result.data;
    },
    enabled: !!roleId,
  });
}

export function useRoleAuditHistory(roleId: string) {
  return useQuery({
    queryKey: ["roles", roleId, "audit"],
    queryFn: async () => {
      const result = await getRoleAuditHistory(roleId);
      if (!result.success) throw new Error(result.message);
      return result.data;
    },
    enabled: !!roleId,
  });
}

// --- Mutation Hooks ---

export function useCreateRole() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateRoleRequest) => createRole(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["roles"] });
    },
  });
}

export function useUpdateRole() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: UpdateRoleRequest) => updateRole(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["roles"] });
    },
  });
}

export function useDeleteRole() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => deleteRole(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["roles"] });
    },
  });
}

export function useAssignRoleToUsers() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({
      roleId,
      userIds,
    }: {
      roleId: string;
      userIds: string[];
    }) => assignRoleToUsers(roleId, userIds),
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({
        queryKey: ["roles", variables.roleId, "users"],
      });
      queryClient.invalidateQueries({ queryKey: ["roles"] });
    },
  });
}

export function useRemoveRoleFromUsers() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({
      roleId,
      userIds,
    }: {
      roleId: string;
      userIds: string[];
    }) => removeRoleFromUsers(roleId, userIds),
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({
        queryKey: ["roles", variables.roleId, "users"],
      });
      queryClient.invalidateQueries({ queryKey: ["roles"] });
    },
  });
}

export function useCloneRole() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({
      roleId,
      newName,
      newDisplayName,
    }: {
      roleId: string;
      newName: string;
      newDisplayName: string;
    }) => cloneRole(roleId, newName, newDisplayName),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["roles"] });
    },
  });
}
