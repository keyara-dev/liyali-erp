// Example usage of the user mutations hooks
// This file shows how to use the reusable user mutation hooks

import React from "react";
import {
  useCreateUser,
  useUpdateUser,
  useDeleteUser,
  useToggleUserStatus,
  useResetUserPassword,
  useBulkUserOperations,
} from "@/hooks/use-users-mutations";

export function UserManagementExample() {
  // Example 1: Create a new user
  const createUser = useCreateUser((data) => {
    console.log("User created:", data);
    // Navigate to user list or show success message
  });

  // Example 2: Update an existing user
  const updateUser = useUpdateUser((data) => {
    console.log("User updated:", data);
    // Refresh user data or show success message
  });

  // Example 3: Delete a user
  const deleteUser = useDeleteUser((data) => {
    console.log("User deleted:", data);
    // Remove user from list or refresh data
  });

  // Example 4: Toggle user status (activate/deactivate)
  const toggleStatus = useToggleUserStatus((data) => {
    console.log("User status toggled:", data);
  });

  // Example 5: Reset user password
  const resetPassword = useResetUserPassword((data) => {
    console.log("Password reset:", data);
  });

  // Example 6: Bulk operations
  const bulkOperations = useBulkUserOperations((results) => {
    console.log("Bulk operation results:", results);
  });

  const handleCreateUser = async () => {
    try {
      await createUser.mutateAsync({
        email: "newuser@example.com",
        password: "TempPassword123!",
        first_name: "John",
        last_name: "Doe",
        role: "requester",
        username: "johndoe",
        branch_id: "",
        department_id: "dept-123",
        role_id: "role-456",
      });
    } catch (error) {
      // Error is handled by the hook
      console.error("Failed to create user:", error);
    }
  };

  const handleUpdateUser = async (userId: string) => {
    try {
      await updateUser.mutateAsync({
        userId,
        data: {
          first_name: "Jane",
          last_name: "Smith",
          is_active: true,
        },
      });
    } catch (error) {
      console.error("Failed to update user:", error);
    }
  };

  const handleDeleteUser = async (userId: string) => {
    try {
      await deleteUser.mutateAsync(userId);
    } catch (error) {
      console.error("Failed to delete user:", error);
    }
  };

  const handleToggleStatus = async (userId: string, isActive: boolean) => {
    try {
      await toggleStatus.mutateAsync({ userId, isActive });
    } catch (error) {
      console.error("Failed to toggle user status:", error);
    }
  };

  const handleResetPassword = async (userId: string) => {
    try {
      await resetPassword.mutateAsync({
        userId,
        password: "NewTempPassword123!",
      });
    } catch (error) {
      console.error("Failed to reset password:", error);
    }
  };

  const handleBulkActivate = async (userIds: string[]) => {
    try {
      await bulkOperations.mutateAsync({
        operation: "activate",
        userIds,
      });
    } catch (error) {
      console.error("Failed to bulk activate users:", error);
    }
  };

  return (
    <div className="space-y-4">
      <h2>User Management Examples</h2>

      <div className="space-x-2">
        <button
          onClick={handleCreateUser}
          disabled={createUser.isPending}
          className="px-4 py-2 bg-blue-500 text-white rounded disabled:opacity-50"
        >
          {createUser.isPending ? "Creating..." : "Create User"}
        </button>

        <button
          onClick={() => handleUpdateUser("user-123")}
          disabled={updateUser.isPending}
          className="px-4 py-2 bg-green-500 text-white rounded disabled:opacity-50"
        >
          {updateUser.isPending ? "Updating..." : "Update User"}
        </button>

        <button
          onClick={() => handleDeleteUser("user-123")}
          disabled={deleteUser.isPending}
          className="px-4 py-2 bg-red-500 text-white rounded disabled:opacity-50"
        >
          {deleteUser.isPending ? "Deleting..." : "Delete User"}
        </button>

        <button
          onClick={() => handleToggleStatus("user-123", false)}
          disabled={toggleStatus.isPending}
          className="px-4 py-2 bg-yellow-500 text-white rounded disabled:opacity-50"
        >
          {toggleStatus.isPending ? "Toggling..." : "Deactivate User"}
        </button>

        <button
          onClick={() => handleResetPassword("user-123")}
          disabled={resetPassword.isPending}
          className="px-4 py-2 bg-purple-500 text-white rounded disabled:opacity-50"
        >
          {resetPassword.isPending ? "Resetting..." : "Reset Password"}
        </button>

        <button
          onClick={() => handleBulkActivate(["user-1", "user-2", "user-3"])}
          disabled={bulkOperations.isPending}
          className="px-4 py-2 bg-indigo-500 text-white rounded disabled:opacity-50"
        >
          {bulkOperations.isPending ? "Processing..." : "Bulk Activate"}
        </button>
      </div>

      {/* Loading states */}
      {createUser.isPending && <p>Creating user...</p>}
      {updateUser.isPending && <p>Updating user...</p>}
      {deleteUser.isPending && <p>Deleting user...</p>}
      {toggleStatus.isPending && <p>Toggling user status...</p>}
      {resetPassword.isPending && <p>Resetting password...</p>}
      {bulkOperations.isPending && <p>Processing bulk operation...</p>}
    </div>
  );
}

// Example of using hooks in a user list component
export function UserListExample({ users }: { users: any[] }) {
  const deleteUser = useDeleteUser();
  const toggleStatus = useToggleUserStatus();

  return (
    <div className="space-y-2">
      {users.map((user) => (
        <div
          key={user.id}
          className="flex items-center justify-between p-4 border rounded"
        >
          <div>
            <h3>{user.name}</h3>
            <p>{user.email}</p>
            <p>Status: {user.is_active ? "Active" : "Inactive"}</p>
          </div>
          <div className="space-x-2">
            <button
              onClick={() =>
                toggleStatus.mutateAsync({
                  userId: user.id,
                  isActive: !user.is_active,
                })
              }
              disabled={toggleStatus.isPending}
              className="px-3 py-1 bg-blue-500 text-white rounded text-sm"
            >
              {user.is_active ? "Deactivate" : "Activate"}
            </button>
            <button
              onClick={() => deleteUser.mutateAsync(user.id)}
              disabled={deleteUser.isPending}
              className="px-3 py-1 bg-red-500 text-white rounded text-sm"
            >
              Delete
            </button>
          </div>
        </div>
      ))}
    </div>
  );
}
