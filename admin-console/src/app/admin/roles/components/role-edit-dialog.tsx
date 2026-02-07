"use client";

import { useState, useEffect } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Switch } from "@/components/ui/switch";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import { ScrollArea } from "@/components/ui/scroll-area";
import { toast } from "sonner";
import {
  updateRole,
  type Role,
  type Permission,
  type UpdateRoleRequest,
} from "@/app/_actions/roles";

interface RoleEditDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  role: Role | null;
  permissions: Permission[];
  onRoleUpdated: () => void;
}

export function RoleEditDialog({
  open,
  onOpenChange,
  role,
  permissions,
  onRoleUpdated,
}: RoleEditDialogProps) {
  const [isLoading, setIsLoading] = useState(false);
  const [formData, setFormData] = useState<UpdateRoleRequest>({
    id: "",
    name: "",
    display_name: "",
    description: "",
    permission_ids: [],
    is_active: true,
  });

  // Group permissions by category
  const permissionsByCategory = permissions.reduce(
    (acc, permission) => {
      if (!acc[permission.category]) {
        acc[permission.category] = [];
      }
      acc[permission.category].push(permission);
      return acc;
    },
    {} as Record<string, Permission[]>,
  );

  useEffect(() => {
    if (role && open) {
      setFormData({
        id: role.id,
        name: role.name,
        display_name: role.display_name,
        description: role.description,
        permission_ids: role.permissions.map((p) => p.id),
        is_active: role.is_active,
      });
    }
  }, [role, open]);

  const handleInputChange = (field: keyof UpdateRoleRequest, value: any) => {
    setFormData((prev) => ({
      ...prev,
      [field]: value,
    }));
  };

  const handlePermissionToggle = (permissionId: string, checked: boolean) => {
    if (checked) {
      setFormData((prev) => ({
        ...prev,
        permission_ids: [...(prev.permission_ids || []), permissionId],
      }));
    } else {
      setFormData((prev) => ({
        ...prev,
        permission_ids: (prev.permission_ids || []).filter(
          (id) => id !== permissionId,
        ),
      }));
    }
  };

  const handleCategoryToggle = (category: string, checked: boolean) => {
    const categoryPermissions = permissionsByCategory[category] || [];
    const categoryPermissionIds = categoryPermissions.map((p) => p.id);

    if (checked) {
      // Add all permissions from this category
      setFormData((prev) => ({
        ...prev,
        permission_ids: [
          ...new Set([
            ...(prev.permission_ids || []),
            ...categoryPermissionIds,
          ]),
        ],
      }));
    } else {
      // Remove all permissions from this category
      setFormData((prev) => ({
        ...prev,
        permission_ids: (prev.permission_ids || []).filter(
          (id) => !categoryPermissionIds.includes(id),
        ),
      }));
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!formData.display_name?.trim()) {
      toast.error("Please fill in all required fields");
      return;
    }

    if (!formData.permission_ids || formData.permission_ids.length === 0) {
      toast.error("Please select at least one permission");
      return;
    }

    setIsLoading(true);

    try {
      const result = await updateRole(formData);

      if (result.success) {
        toast.success("Role updated successfully");
        onRoleUpdated();
      } else {
        toast.error(result.message || "Failed to update role");
      }
    } catch (error) {
      console.error("Error updating role:", error);
      toast.error("Failed to update role");
    } finally {
      setIsLoading(false);
    }
  };

  const getCategoryPermissionCount = (category: string) => {
    const categoryPermissions = permissionsByCategory[category] || [];
    const selectedCount = categoryPermissions.filter((p) =>
      (formData.permission_ids || []).includes(p.id),
    ).length;
    return { selected: selectedCount, total: categoryPermissions.length };
  };

  const isCategoryFullySelected = (category: string) => {
    const { selected, total } = getCategoryPermissionCount(category);
    return selected === total && total > 0;
  };

  const isCategoryPartiallySelected = (category: string) => {
    const { selected, total } = getCategoryPermissionCount(category);
    return selected > 0 && selected < total;
  };

  if (!role) return null;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-4xl max-h-[90vh] overflow-hidden">
        <DialogHeader>
          <DialogTitle>Edit Role</DialogTitle>
          <DialogDescription>
            Update role information and permissions
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-6">
          <div className="grid gap-6 md:grid-cols-2">
            {/* Basic Information */}
            <div className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="name">Role Name</Label>
                <Input
                  id="name"
                  value={formData.name}
                  onChange={(e) => handleInputChange("name", e.target.value)}
                  disabled={role.is_system_role}
                  placeholder="e.g., content_manager"
                />
                {role.is_system_role && (
                  <p className="text-xs text-muted-foreground">
                    System role names cannot be changed
                  </p>
                )}
              </div>

              <div className="space-y-2">
                <Label htmlFor="display_name">Display Name *</Label>
                <Input
                  id="display_name"
                  value={formData.display_name}
                  onChange={(e) =>
                    handleInputChange("display_name", e.target.value)
                  }
                  placeholder="e.g., Content Manager"
                  required
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="description">Description</Label>
                <Textarea
                  id="description"
                  value={formData.description}
                  onChange={(e) =>
                    handleInputChange("description", e.target.value)
                  }
                  placeholder="Describe what this role can do..."
                  rows={3}
                />
              </div>

              <div className="flex items-center space-x-2">
                <Switch
                  id="is_active"
                  checked={formData.is_active}
                  onCheckedChange={(checked) =>
                    handleInputChange("is_active", checked)
                  }
                  disabled={role.is_system_role}
                />
                <Label htmlFor="is_active">Active Role</Label>
                {role.is_system_role && (
                  <Badge variant="outline" className="text-xs">
                    System Role
                  </Badge>
                )}
              </div>

              {role.user_count > 0 && (
                <div className="p-3 bg-yellow-50 border border-yellow-200 rounded-lg">
                  <p className="text-sm text-yellow-800">
                    <strong>Warning:</strong> This role is assigned to{" "}
                    {role.user_count} user(s). Changes will affect their
                    permissions immediately.
                  </p>
                </div>
              )}
            </div>

            {/* Permissions Selection */}
            <div className="space-y-4">
              <div>
                <Label>Permissions *</Label>
                <p className="text-xs text-muted-foreground">
                  Select permissions for this role (
                  {(formData.permission_ids || []).length} selected)
                </p>
              </div>

              <ScrollArea className="h-96 border rounded-lg">
                <div className="p-4 space-y-4">
                  {Object.entries(permissionsByCategory).map(
                    ([category, categoryPermissions]) => (
                      <Card key={category}>
                        <CardHeader className="pb-3">
                          <div className="flex items-center justify-between">
                            <CardTitle className="text-sm capitalize">
                              {category.replace(/_/g, " ")}
                            </CardTitle>
                            <div className="flex items-center gap-2">
                              <Badge variant="outline" className="text-xs">
                                {getCategoryPermissionCount(category).selected}/
                                {getCategoryPermissionCount(category).total}
                              </Badge>
                              <input
                                type="checkbox"
                                checked={isCategoryFullySelected(category)}
                                ref={(el) => {
                                  if (el) {
                                    el.indeterminate =
                                      isCategoryPartiallySelected(category);
                                  }
                                }}
                                onChange={(e) =>
                                  handleCategoryToggle(
                                    category,
                                    e.target.checked,
                                  )
                                }
                                className="rounded border-gray-300"
                              />
                            </div>
                          </div>
                        </CardHeader>
                        <CardContent className="pt-0">
                          <div className="space-y-2">
                            {categoryPermissions.map((permission) => (
                              <div
                                key={permission.id}
                                className="flex items-center justify-between p-2 rounded-lg hover:bg-muted/50"
                              >
                                <div className="flex-1">
                                  <div className="font-medium text-sm">
                                    {permission.display_name}
                                  </div>
                                  <div className="text-xs text-muted-foreground">
                                    {permission.description}
                                  </div>
                                </div>
                                <input
                                  type="checkbox"
                                  checked={(
                                    formData.permission_ids || []
                                  ).includes(permission.id)}
                                  onChange={(e) =>
                                    handlePermissionToggle(
                                      permission.id,
                                      e.target.checked,
                                    )
                                  }
                                  className="rounded border-gray-300"
                                />
                              </div>
                            ))}
                          </div>
                        </CardContent>
                      </Card>
                    ),
                  )}
                </div>
              </ScrollArea>
            </div>
          </div>

          <Separator />

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={isLoading}>
              {isLoading ? "Updating..." : "Update Role"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
