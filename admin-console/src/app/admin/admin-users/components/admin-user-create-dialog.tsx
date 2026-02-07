"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Switch } from "@/components/ui/switch";
import { Badge } from "@/components/ui/badge";
import { Checkbox } from "@/components/ui/checkbox";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Separator } from "@/components/ui/separator";
import { Eye, EyeOff, Shield, Mail, Lock } from "lucide-react";
import { notify } from "@/lib/notify";
import {
  createAdminUser,
  type CreateAdminUserRequest,
  type AdminRole,
} from "@/app/_actions/admin-users";

interface AdminUserCreateDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  roles: AdminRole[];
  onUserCreated: () => void;
}

export function AdminUserCreateDialog({
  open,
  onOpenChange,
  roles,
  onUserCreated,
}: AdminUserCreateDialogProps) {
  const [isLoading, setIsLoading] = useState(false);
  const [showPassword, setShowPassword] = useState(false);
  const [formData, setFormData] = useState<CreateAdminUserRequest>({
    email: "",
    first_name: "",
    last_name: "",
    password: "",
    is_active: true,
    is_super_admin: false,
    role_ids: [],
    send_welcome_email: true,
    require_password_change: true,
  });

  const [errors, setErrors] = useState<Record<string, string>>({});

  const validateForm = (): boolean => {
    const newErrors: Record<string, string> = {};

    if (!formData.email) {
      newErrors.email = "Email is required";
    } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(formData.email)) {
      newErrors.email = "Please enter a valid email address";
    }

    if (!formData.first_name) {
      newErrors.first_name = "First name is required";
    }

    if (!formData.last_name) {
      newErrors.last_name = "Last name is required";
    }

    if (!formData.password) {
      newErrors.password = "Password is required";
    } else if (formData.password.length < 8) {
      newErrors.password = "Password must be at least 8 characters";
    }

    if (formData.role_ids.length === 0) {
      newErrors.roles = "At least one role must be selected";
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!validateForm()) {
      return;
    }

    setIsLoading(true);

    try {
      const result = await createAdminUser(formData);

      if (result.success) {
        notify("Admin user created successfully", {
          variant: "success",
        });
        onUserCreated();
        handleClose();
      } else {
        notify(result.message || "Failed to create admin user", {
          variant: "destructive",
        });
      }
    } catch (error) {
      console.error("Error creating admin user:", error);
      notify("Failed to create admin user", {
        variant: "destructive",
      });
    } finally {
      setIsLoading(false);
    }
  };

  const handleClose = () => {
    setFormData({
      email: "",
      first_name: "",
      last_name: "",
      password: "",
      is_active: true,
      is_super_admin: false,
      role_ids: [],
      send_welcome_email: true,
      require_password_change: true,
    });
    setErrors({});
    setShowPassword(false);
    onOpenChange(false);
  };

  const handleRoleToggle = (roleId: string, checked: boolean) => {
    if (checked) {
      setFormData((prev) => ({
        ...prev,
        role_ids: [...prev.role_ids, roleId],
      }));
    } else {
      setFormData((prev) => ({
        ...prev,
        role_ids: prev.role_ids.filter((id) => id !== roleId),
      }));
    }
  };

  const generatePassword = () => {
    const chars =
      "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789!@#$%^&*";
    let password = "";
    for (let i = 0; i < 12; i++) {
      password += chars.charAt(Math.floor(Math.random() * chars.length));
    }
    setFormData((prev) => ({ ...prev, password }));
  };

  const selectedRoles = roles.filter((role) =>
    formData.role_ids.includes(role.id),
  );

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Shield className="h-5 w-5" />
            Create Admin User
          </DialogTitle>
          <DialogDescription>
            Create a new admin user with specific roles and permissions.
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-6">
          {/* Basic Information */}
          <div className="space-y-4">
            <h3 className="text-sm font-medium">Basic Information</h3>

            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="first_name">
                  First Name <span className="text-red-500">*</span>
                </Label>
                <Input
                  id="first_name"
                  value={formData.first_name}
                  onChange={(e) =>
                    setFormData((prev) => ({
                      ...prev,
                      first_name: e.target.value,
                    }))
                  }
                  className={errors.first_name ? "border-red-500" : ""}
                />
                {errors.first_name && (
                  <p className="text-sm text-red-500">{errors.first_name}</p>
                )}
              </div>

              <div className="space-y-2">
                <Label htmlFor="last_name">
                  Last Name <span className="text-red-500">*</span>
                </Label>
                <Input
                  id="last_name"
                  value={formData.last_name}
                  onChange={(e) =>
                    setFormData((prev) => ({
                      ...prev,
                      last_name: e.target.value,
                    }))
                  }
                  className={errors.last_name ? "border-red-500" : ""}
                />
                {errors.last_name && (
                  <p className="text-sm text-red-500">{errors.last_name}</p>
                )}
              </div>
            </div>

            <div className="space-y-2">
              <Label htmlFor="email">
                Email Address <span className="text-red-500">*</span>
              </Label>
              <div className="relative">
                <Mail className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground h-4 w-4" />
                <Input
                  id="email"
                  type="email"
                  value={formData.email}
                  onChange={(e) =>
                    setFormData((prev) => ({ ...prev, email: e.target.value }))
                  }
                  className={`pl-10 ${errors.email ? "border-red-500" : ""}`}
                />
              </div>
              {errors.email && (
                <p className="text-sm text-red-500">{errors.email}</p>
              )}
            </div>
          </div>

          <Separator />

          {/* Password */}
          <div className="space-y-4">
            <h3 className="text-sm font-medium">Password</h3>

            <div className="space-y-2">
              <Label htmlFor="password">
                Password <span className="text-red-500">*</span>
              </Label>
              <div className="flex gap-2">
                <div className="relative flex-1">
                  <Lock className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground h-4 w-4" />
                  <Input
                    id="password"
                    type={showPassword ? "text" : "password"}
                    value={formData.password}
                    onChange={(e) =>
                      setFormData((prev) => ({
                        ...prev,
                        password: e.target.value,
                      }))
                    }
                    className={`pl-10 pr-10 ${errors.password ? "border-red-500" : ""}`}
                  />
                  <button
                    type="button"
                    onClick={() => setShowPassword(!showPassword)}
                    className="absolute right-3 top-1/2 transform -translate-y-1/2 text-muted-foreground hover:text-foreground"
                  >
                    {showPassword ? (
                      <EyeOff className="h-4 w-4" />
                    ) : (
                      <Eye className="h-4 w-4" />
                    )}
                  </button>
                </div>
                <Button
                  type="button"
                  variant="outline"
                  onClick={generatePassword}
                >
                  Generate
                </Button>
              </div>
              {errors.password && (
                <p className="text-sm text-red-500">{errors.password}</p>
              )}
            </div>

            <div className="flex items-center space-x-2">
              <Checkbox
                id="require_password_change"
                checked={formData.require_password_change}
                onCheckedChange={(checked) =>
                  setFormData((prev) => ({
                    ...prev,
                    require_password_change: checked as boolean,
                  }))
                }
              />
              <Label htmlFor="require_password_change" className="text-sm">
                Require password change on first login
              </Label>
            </div>
          </div>

          <Separator />

          {/* Admin Settings */}
          <div className="space-y-4">
            <h3 className="text-sm font-medium">Admin Settings</h3>

            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <div className="space-y-0.5">
                  <Label>Active Status</Label>
                  <p className="text-sm text-muted-foreground">
                    User can log in and access the system
                  </p>
                </div>
                <Switch
                  checked={formData.is_active}
                  onCheckedChange={(checked) =>
                    setFormData((prev) => ({ ...prev, is_active: checked }))
                  }
                />
              </div>

              <div className="flex items-center justify-between">
                <div className="space-y-0.5">
                  <Label>Super Admin</Label>
                  <p className="text-sm text-muted-foreground">
                    Full system access with all permissions
                  </p>
                </div>
                <Switch
                  checked={formData.is_super_admin}
                  onCheckedChange={(checked) =>
                    setFormData((prev) => ({
                      ...prev,
                      is_super_admin: checked,
                    }))
                  }
                />
              </div>

              <div className="flex items-center justify-between">
                <div className="space-y-0.5">
                  <Label>Send Welcome Email</Label>
                  <p className="text-sm text-muted-foreground">
                    Send login credentials via email
                  </p>
                </div>
                <Switch
                  checked={formData.send_welcome_email}
                  onCheckedChange={(checked) =>
                    setFormData((prev) => ({
                      ...prev,
                      send_welcome_email: checked,
                    }))
                  }
                />
              </div>
            </div>
          </div>

          <Separator />

          {/* Role Assignment */}
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <h3 className="text-sm font-medium">
                Role Assignment <span className="text-red-500">*</span>
              </h3>
              {selectedRoles.length > 0 && (
                <Badge variant="outline">{selectedRoles.length} selected</Badge>
              )}
            </div>

            {errors.roles && (
              <p className="text-sm text-red-500">{errors.roles}</p>
            )}

            <div className="grid gap-3 max-h-48 overflow-y-auto border rounded-lg p-3">
              {roles.map((role) => (
                <div
                  key={role.id}
                  className="flex items-center space-x-3 p-2 hover:bg-muted/50 rounded"
                >
                  <Checkbox
                    id={`role-${role.id}`}
                    checked={formData.role_ids.includes(role.id)}
                    onCheckedChange={(checked) =>
                      handleRoleToggle(role.id, checked as boolean)
                    }
                  />
                  <div className="flex-1">
                    <Label
                      htmlFor={`role-${role.id}`}
                      className="font-medium cursor-pointer"
                    >
                      {role.display_name}
                    </Label>
                    <p className="text-sm text-muted-foreground">
                      {role.description}
                    </p>
                  </div>
                  {role.is_system_role && (
                    <Badge variant="destructive" className="text-xs">
                      System
                    </Badge>
                  )}
                </div>
              ))}
            </div>

            {selectedRoles.length > 0 && (
              <div className="space-y-2">
                <Label className="text-sm">Selected Roles:</Label>
                <div className="flex flex-wrap gap-2">
                  {selectedRoles.map((role) => (
                    <Badge key={role.id} variant="secondary">
                      {role.display_name}
                      <button
                        type="button"
                        onClick={() => handleRoleToggle(role.id, false)}
                        className="ml-1 hover:text-destructive"
                      >
                        ×
                      </button>
                    </Badge>
                  ))}
                </div>
              </div>
            )}
          </div>

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={handleClose}
              disabled={isLoading}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={isLoading}>
              {isLoading ? "Creating..." : "Create Admin User"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
