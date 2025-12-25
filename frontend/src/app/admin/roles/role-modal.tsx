"use client";

import { useState, useEffect } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { createRoleAction, updateRoleAction } from "@/app/_actions/roles";

interface RoleModalProps {
  role?: any;
  open: boolean;
  onClose: () => void;
}

export function RoleModal({ role, open, onClose }: RoleModalProps) {
  const queryClient = useQueryClient();
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [errors, setErrors] = useState<Record<string, string>>({});

  // Initialize form when role changes
  useEffect(() => {
    if (role) {
      setName(role.name || "");
      setDescription(role.description || "");
    } else {
      setName("");
      setDescription("");
    }
    setErrors({});
  }, [role, open]);

  // Create mutation
  const createMutation = useMutation({
    mutationFn: () => createRoleAction(name, description),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["organization-roles"] });
      onClose();
    },
    onError: (error: any) => {
      setErrors({
        submit: error?.message || "Failed to create role",
      });
    },
  });

  // Update mutation
  const updateMutation = useMutation({
    mutationFn: () => updateRoleAction(role.id, name, description),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["organization-roles"] });
      onClose();
    },
    onError: (error: any) => {
      setErrors({
        submit: error?.message || "Failed to update role",
      });
    },
  });

  const validate = (): boolean => {
    const newErrors: Record<string, string> = {};

    if (!name.trim()) {
      newErrors.name = "Role name is required";
    } else if (name.trim().length < 3) {
      newErrors.name = "Role name must be at least 3 characters";
    }

    if (!description.trim()) {
      newErrors.description = "Description is required";
    } else if (description.trim().length < 10) {
      newErrors.description = "Description must be at least 10 characters";
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();

    if (!validate()) {
      return;
    }

    if (role?.id) {
      updateMutation.mutate();
    } else {
      createMutation.mutate();
    }
  };

  if (!open) return null;

  const isLoading = createMutation.isPending || updateMutation.isPending;
  const title = role ? "Edit Role" : "Create Role";

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg shadow-lg max-w-md w-full mx-4">
        {/* Header */}
        <div className="px-6 py-4 border-b border-gray-200">
          <h2 className="text-xl font-bold text-gray-900">{title}</h2>
        </div>

        {/* Content */}
        <form onSubmit={handleSubmit} className="p-6 space-y-4">
          {/* Error message */}
          {errors.submit && (
            <div className="p-3 bg-red-50 border border-red-200 rounded text-red-700 text-sm">
              {errors.submit}
            </div>
          )}

          {/* Name field */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Role Name <span className="text-red-500">*</span>
            </label>
            <input
              type="text"
              value={name}
              onChange={(e) => {
                setName(e.target.value);
                if (errors.name) setErrors({ ...errors, name: "" });
              }}
              placeholder="e.g., Senior Manager"
              className={`w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 ${
                errors.name
                  ? "border-red-500 focus:ring-red-500"
                  : "border-gray-300 focus:ring-blue-500"
              }`}
              disabled={isLoading}
            />
            {errors.name && (
              <p className="text-red-500 text-sm mt-1">{errors.name}</p>
            )}
          </div>

          {/* Description field */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Description <span className="text-red-500">*</span>
            </label>
            <textarea
              value={description}
              onChange={(e) => {
                setDescription(e.target.value);
                if (errors.description)
                  setErrors({ ...errors, description: "" });
              }}
              placeholder="What is this role responsible for?"
              rows={4}
              className={`w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 resize-none ${
                errors.description
                  ? "border-red-500 focus:ring-red-500"
                  : "border-gray-300 focus:ring-blue-500"
              }`}
              disabled={isLoading}
            />
            {errors.description && (
              <p className="text-red-500 text-sm mt-1">{errors.description}</p>
            )}
          </div>
        </form>

        {/* Footer */}
        <div className="px-6 py-4 border-t border-gray-200 flex gap-2 justify-end">
          <button
            onClick={onClose}
            disabled={isLoading}
            className="px-4 py-2 text-gray-700 border border-gray-300 rounded-lg hover:bg-gray-50 disabled:opacity-50"
          >
            Cancel
          </button>
          <button
            onClick={handleSubmit}
            disabled={isLoading}
            className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50"
          >
            {isLoading ? "Loading..." : role ? "Update" : "Create"}
          </button>
        </div>
      </div>
    </div>
  );
}
