"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import { AdminGuard } from "@/components/auth/permission-guard";
import { RoleModal } from "./role-modal";
import { PermissionsModal } from "./permissions-modal";
import { getRolesAction, deleteRoleAction } from "@/app/_actions/roles";

export default function RolesPage() {
  const queryClient = useQueryClient();
  const [showRoleModal, setShowRoleModal] = useState(false);
  const [showPermissionsModal, setShowPermissionsModal] = useState(false);
  const [selectedRole, setSelectedRole] = useState<any>(null);

  // Fetch roles
  const { data: roles, isLoading, error, refetch } = useQuery({
    queryKey: ["organization-roles"],
    queryFn: () => getRolesAction(),
    staleTime: 5 * 60 * 1000, // 5 minutes
  });

  // Delete mutation
  const deleteMutation = useMutation({
    mutationFn: (roleId: string) => deleteRoleAction(roleId),
    onSuccess: () => {
      refetch();
      queryClient.invalidateQueries({ queryKey: ["organization-roles"] });
    },
    onError: (error: any) => {
      alert("Error deleting role: " + (error?.message || "Unknown error"));
    },
  });

  const handleEditRole = (role: any) => {
    setSelectedRole(role);
    setShowRoleModal(true);
  };

  const handleDeleteRole = (roleId: string) => {
    if (confirm("Are you sure you want to delete this role?")) {
      deleteMutation.mutate(roleId);
    }
  };

  const handleManagePermissions = (role: any) => {
    setSelectedRole(role);
    setShowPermissionsModal(true);
  };

  const handleCloseRoleModal = () => {
    setShowRoleModal(false);
    setSelectedRole(null);
    refetch();
  };

  const handleClosePermissionsModal = () => {
    setShowPermissionsModal(false);
    setSelectedRole(null);
  };

  return (
    <AdminGuard>
      <div className="space-y-6 p-6">
        {/* Header */}
        <div className="flex justify-between items-center">
          <div>
            <h1 className="text-3xl font-bold">Organization Roles</h1>
            <p className="text-gray-600 mt-2">
              Create and manage custom roles for your organization
            </p>
          </div>
          <button
            onClick={() => {
              setSelectedRole(null);
              setShowRoleModal(true);
            }}
            className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
          >
            Create Role
          </button>
        </div>

        {/* Error state */}
        {error && (
          <div className="p-4 bg-red-50 border border-red-200 rounded-lg text-red-700">
            Error loading roles. Please try again.
          </div>
        )}

        {/* Loading state */}
        {isLoading && (
          <div className="p-4 bg-gray-50 border border-gray-200 rounded-lg text-gray-700">
            Loading roles...
          </div>
        )}

        {/* Roles table */}
        {!isLoading && roles && roles.data && (
          <div className="overflow-x-auto border border-gray-200 rounded-lg">
            <table className="w-full">
              <thead className="bg-gray-50 border-b border-gray-200">
                <tr>
                  <th className="px-6 py-3 text-left text-sm font-medium text-gray-700">
                    Name
                  </th>
                  <th className="px-6 py-3 text-left text-sm font-medium text-gray-700">
                    Description
                  </th>
                  <th className="px-6 py-3 text-left text-sm font-medium text-gray-700">
                    Status
                  </th>
                  <th className="px-6 py-3 text-right text-sm font-medium text-gray-700">
                    Actions
                  </th>
                </tr>
              </thead>
              <tbody>
                {roles.data.map((role: any) => (
                  <tr
                    key={role.id}
                    className="border-b border-gray-200 hover:bg-gray-50"
                  >
                    <td className="px-6 py-4 text-sm font-medium text-gray-900">
                      {role.name}
                      {role.isDefault && (
                        <span className="ml-2 px-2 py-1 bg-blue-100 text-blue-700 text-xs rounded">
                          System
                        </span>
                      )}
                    </td>
                    <td className="px-6 py-4 text-sm text-gray-600">
                      {role.description}
                    </td>
                    <td className="px-6 py-4 text-sm">
                      <span
                        className={`px-2 py-1 rounded text-xs font-medium ${
                          role.isActive
                            ? "bg-green-100 text-green-700"
                            : "bg-gray-100 text-gray-700"
                        }`}
                      >
                        {role.isActive ? "Active" : "Inactive"}
                      </span>
                    </td>
                    <td className="px-6 py-4 text-sm text-right space-x-2">
                      <button
                        onClick={() => handleManagePermissions(role)}
                        className="text-blue-600 hover:text-blue-700 font-medium"
                      >
                        Permissions
                      </button>
                      {!role.isDefault && (
                        <>
                          <button
                            onClick={() => handleEditRole(role)}
                            className="text-blue-600 hover:text-blue-700 font-medium"
                          >
                            Edit
                          </button>
                          <button
                            onClick={() => handleDeleteRole(role.id)}
                            disabled={deleteMutation.isPending}
                            className="text-red-600 hover:text-red-700 font-medium disabled:opacity-50"
                          >
                            Delete
                          </button>
                        </>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>

            {roles.data.length === 0 && (
              <div className="p-6 text-center text-gray-600">
                No roles found. Create your first role to get started.
              </div>
            )}
          </div>
        )}

        {/* Modals */}
        <RoleModal
          role={selectedRole}
          open={showRoleModal}
          onClose={handleCloseRoleModal}
        />

        <PermissionsModal
          role={selectedRole}
          open={showPermissionsModal}
          onClose={handleClosePermissionsModal}
        />
      </div>
    </AdminGuard>
  );
}
