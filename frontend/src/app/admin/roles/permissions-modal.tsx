"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { useState, useMemo } from "react";
import {
  getRolePermissionsAction,
  getAvailablePermissionsAction,
  assignPermissionAction,
  removePermissionAction,
} from "@/app/_actions/roles-permissions";

interface PermissionsModalProps {
  role?: any;
  open: boolean;
  onClose: () => void;
}

export function PermissionsModal({ role, open, onClose }: PermissionsModalProps) {
  const queryClient = useQueryClient();
  const [searchTerm, setSearchTerm] = useState("");

  // Fetch assigned permissions
  const { data: assignedPermsData, isLoading: assignedLoading } = useQuery({
    queryKey: ["role-permissions", role?.id],
    queryFn: () => getRolePermissionsAction(role?.id),
    enabled: open && !!role?.id,
    staleTime: 5 * 60 * 1000,
  });

  // Fetch available permissions
  const { data: availablePermsData, isLoading: availableLoading } = useQuery({
    queryKey: ["available-permissions"],
    queryFn: () => getAvailablePermissionsAction(),
    enabled: open,
    staleTime: 5 * 60 * 1000,
  });

  // Assign permission mutation
  const assignMutation = useMutation({
    mutationFn: (permId: string) => assignPermissionAction(role?.id, permId),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ["role-permissions", role?.id],
      });
    },
  });

  // Remove permission mutation
  const removeMutation = useMutation({
    mutationFn: (permId: string) => removePermissionAction(role?.id, permId),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ["role-permissions", role?.id],
      });
    },
  });

  // Get assigned permission IDs
  const assignedPermIds = useMemo(() => {
    if (!assignedPermsData?.data) return new Set();
    return new Set(assignedPermsData.data.map((p: any) => p.id));
  }, [assignedPermsData]);

  // Get available permissions
  const availablePerms = useMemo(() => {
    if (!availablePermsData?.data) return [];
    return availablePermsData.data;
  }, [availablePermsData]);

  // Group permissions by resource
  const groupedPerms = useMemo(() => {
    const grouped: Record<string, any[]> = {};

    availablePerms.forEach((perm: any) => {
      if (!grouped[perm.resource]) {
        grouped[perm.resource] = [];
      }
      grouped[perm.resource].push(perm);
    });

    return grouped;
  }, [availablePerms]);

  // Filter permissions based on search
  const filteredGroupedPerms = useMemo(() => {
    if (!searchTerm.trim()) return groupedPerms;

    const filtered: Record<string, any[]> = {};
    const searchLower = searchTerm.toLowerCase();

    Object.entries(groupedPerms).forEach(([resource, perms]) => {
      const filteredPerms = perms.filter(
        (p: any) =>
          p.resource.toLowerCase().includes(searchLower) ||
          p.action.toLowerCase().includes(searchLower) ||
          p.description?.toLowerCase().includes(searchLower)
      );

      if (filteredPerms.length > 0) {
        filtered[resource] = filteredPerms;
      }
    });

    return filtered;
  }, [groupedPerms, searchTerm]);

  const handleTogglePermission = (permId: string) => {
    if (assignedPermIds.has(permId)) {
      removeMutation.mutate(permId);
    } else {
      assignMutation.mutate(permId);
    }
  };

  if (!open || !role) return null;

  const isLoading = assignedLoading || availableLoading;

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg shadow-lg max-w-2xl w-full mx-4 max-h-[90vh] flex flex-col">
        {/* Header */}
        <div className="px-6 py-4 border-b border-gray-200">
          <h2 className="text-xl font-bold text-gray-900">
            Permissions for {role.name}
          </h2>
          <p className="text-gray-600 text-sm mt-1">
            Select which permissions this role should have
          </p>
        </div>

        {/* Search */}
        <div className="px-6 py-4 border-b border-gray-200">
          <input
            type="text"
            placeholder="Search permissions..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
        </div>

        {/* Permissions list */}
        <div className="flex-1 overflow-y-auto p-6 space-y-6">
          {isLoading ? (
            <div className="text-center text-gray-600">Loading permissions...</div>
          ) : Object.keys(filteredGroupedPerms).length === 0 ? (
            <div className="text-center text-gray-600">No permissions found</div>
          ) : (
            Object.entries(filteredGroupedPerms).map(([resource, perms]) => (
              <div key={resource}>
                <h3 className="font-semibold text-gray-900 capitalize mb-3">
                  {resource}
                </h3>
                <div className="space-y-2 ml-4">
                  {(perms as any[]).map((perm) => (
                    <label
                      key={perm.id}
                      className="flex items-center gap-3 p-2 rounded hover:bg-gray-50 cursor-pointer"
                    >
                      <input
                        type="checkbox"
                        checked={assignedPermIds.has(perm.id)}
                        onChange={() => handleTogglePermission(perm.id)}
                        disabled={
                          assignMutation.isPending ||
                          removeMutation.isPending
                        }
                        className="w-4 h-4 rounded border-gray-300 text-blue-600 focus:ring-2 focus:ring-blue-500"
                      />
                      <div className="flex-1">
                        <div className="font-medium text-gray-900 text-sm">
                          {perm.action}
                        </div>
                        {perm.description && (
                          <div className="text-xs text-gray-600">
                            {perm.description}
                          </div>
                        )}
                      </div>
                    </label>
                  ))}
                </div>
              </div>
            ))
          )}
        </div>

        {/* Footer */}
        <div className="px-6 py-4 border-t border-gray-200 flex justify-end">
          <button
            onClick={onClose}
            className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
          >
            Done
          </button>
        </div>
      </div>
    </div>
  );
}
