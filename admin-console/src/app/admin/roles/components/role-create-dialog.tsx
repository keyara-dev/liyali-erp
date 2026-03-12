"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import { ScrollArea } from "@/components/ui/scroll-area";
import { toast } from "sonner";
import {
  createRole,
  type Permission,
  type CreateRoleRequest,
} from "@/app/_actions/roles";

interface RoleCreateDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  permissions: Permission[];
  onRoleCreated: () => void;
}

export function RoleCreateDialog({
  open,
  onOpenChange,
  permissions,
  onRoleCreated,
}: RoleCreateDialogProps) {
  const [isLoading, setIsLoading] = useState(false);
  const [formData, setFormData] = useState<CreateRoleRequest>({
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

  const handleInputChange = (field: keyof CreateRoleRequest, value: any) => {
    setFormData((prev) => ({
      ...prev,
      [field]: value,
    }));
  };

  const handlePermissionToggle = (permissionId: string, checked: boolean) => {
    if (checked) {
      setFormData((prev) => ({
        ...prev,
        permission_ids: [...prev.permission_ids, permissionId],
      }));
    } else {
      setFormData((prev) => ({
        ...prev,
        permission_ids: prev.permission_ids.filter((id) => id !== permissionId),
      }));
    }
  };

  const handleCategoryToggle = (category: string, checked: boolean) => {
    const categoryPermissions = permissionsByCategory[category] || [];
    const categoryPermissionIds = categoryPermissions.map((p) => p.id);

    if (checked) {
      setFormData((prev) => ({
        ...prev,
        permission_ids: [
          ...new Set([...prev.permission_ids, ...categoryPermissionIds]),
        ],
      }));
    } else {
      setFormData((prev) => ({
        ...prev,
        permission_ids: prev.permission_ids.filter(
          (id) => !categoryPermissionIds.includes(id),
        ),
      }));
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!formData.name.trim() || !formData.display_name.trim()) {
      toast.error("Please fill in all required fields");
      return;
    }

    if (formData.permission_ids.length === 0) {
      toast.error("Please select at least one permission");
      return;
    }

    setIsLoading(true);

    try {
      const result = await createRole(formData);

      if (result.success) {
        toast.success("Role created successfully");
        onRoleCreated();
        handleClose();
      } else {
        toast.error(result.message || "Failed to create role");
      }
    } catch (error) {
      console.error("Error creating role:", error);
      toast.error("Failed to create role");
    } finally {
      setIsLoading(false);
    }
  };

  const handleClose = () => {
    setFormData({
      name: "",
      display_name: "",
      description: "",
      permission_ids: [],
      is_active: true,
    });
    onOpenChange(false);
  };

  const getCategoryPermissionCount = (category: string) => {
    const categoryPermissions = permissionsByCategory[category] || [];
    const selectedCount = categoryPermissions.filter((p) =>
      formData.permission_ids.includes(p.id),
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

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-4xl max-h-[90vh] overflow-hidden">
        <DialogHeader>
          <DialogTitle>Create New Role</DialogTitle>
          <DialogDescription>
            Create a new role with specific permissions for users
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-6">
          <div className="grid gap-6 md:grid-cols-2">
            {/* Basic Information */}
            <div className="space-y-4">
              <Input
                name="name"
                label="Role Name"
                required
                value={formData.name}
                onChange={(e) => handleInputChange("name", e.target.value)}
                placeholder="e.g., content_manager"
                descriptionText="Internal name (lowercase, underscores only)"
              />

              <Input
                name="display_name"
                label="Display Name"
                required
                value={formData.display_name}
                onChange={(e) =>
                  handleInputChange("display_name", e.target.value)
                }
                placeholder="e.g., Content Manager"
              />

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
                />
                <Label htmlFor="is_active">Active Role</Label>
              </div>
            </div>

            {/* Permissions Selection */}
            <div className="space-y-4">
              <div>
                <Label>Permissions *</Label>
                <p className="text-xs text-muted-foreground">
                  Select permissions for this role (
                  {formData.permission_ids.length} selected)
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
                                  checked={formData.permission_ids.includes(
                                    permission.id,
                                  )}
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
            <Button type="button" variant="outline" onClick={handleClose}>
              Cancel
            </Button>
            <Button
              type="submit"
              disabled={isLoading}
              isLoading={isLoading}
              loadingText="Creating..."
            >
              Create Role
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
