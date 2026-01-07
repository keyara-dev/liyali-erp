"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { SelectField } from "@/components/ui/select-field";
import { Plus, UserPlus } from "lucide-react";
import { toast } from "sonner";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { createUserWithDepartment } from "@/app/_actions/user-departments";
import { getActiveDepartments } from "@/app/_actions/departments";

interface CreateUserWithDepartmentProps {
  showTrigger?: boolean;
  triggerVariant?: "default" | "outline" | "ghost";
  triggerSize?: "sm" | "default" | "lg";
}

const USER_ROLES = [
  { id: "viewer", name: "Viewer" },
  { id: "user", name: "User" },
  { id: "manager", name: "Manager" },
  { id: "admin", name: "Admin" },
];

export default function CreateUserWithDepartment({
  showTrigger = true,
  triggerVariant = "default",
  triggerSize = "default",
}: CreateUserWithDepartmentProps) {
  const queryClient = useQueryClient();
  const [open, setOpen] = useState(false);
  const [formData, setFormData] = useState({
    email: "",
    name: "",
    password: "",
    confirmPassword: "",
    role: "viewer",
    departmentId: "",
  });
  const [errors, setErrors] = useState<Record<string, string>>({});

  // Fetch departments
  const { data: departmentsResponse, isLoading: departmentsLoading } = useQuery({
    queryKey: ['active-departments'],
    queryFn: () => getActiveDepartments(),
    staleTime: 5 * 60 * 1000,
  });

  const departments = departmentsResponse?.success ? departmentsResponse.data || [] : [];

  // Create user mutation
  const createUserMutation = useMutation({
    mutationFn: createUserWithDepartment,
    onSuccess: (response) => {
      if (response.success) {
        toast.success("User created successfully");
        queryClient.invalidateQueries({ queryKey: ['organization-users'] });
        queryClient.invalidateQueries({ queryKey: ['users'] });
        handleClose();
      } else {
        toast.error(response.message);
        setErrors({ general: response.message || 'An error occurred' });
      }
    },
    onError: (error: Error) => {
      toast.error(error.message || "Failed to create user");
      setErrors({ general: error.message });
    },
  });

  const validateForm = (): boolean => {
    const newErrors: Record<string, string> = {};

    if (!formData.email.trim()) {
      newErrors.email = "Email is required";
    } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(formData.email)) {
      newErrors.email = "Please enter a valid email address";
    }

    if (!formData.name.trim()) {
      newErrors.name = "Name is required";
    }

    if (!formData.password) {
      newErrors.password = "Password is required";
    } else if (formData.password.length < 6) {
      newErrors.password = "Password must be at least 6 characters";
    }

    if (formData.password !== formData.confirmPassword) {
      newErrors.confirmPassword = "Passwords do not match";
    }

    if (!formData.role) {
      newErrors.role = "Role is required";
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!validateForm()) {
      return;
    }

    createUserMutation.mutate({
      email: formData.email.trim(),
      name: formData.name.trim(),
      password: formData.password,
      role: formData.role,
      departmentId: formData.departmentId || undefined,
    });
  };

  const handleClose = () => {
    setOpen(false);
    setFormData({
      email: "",
      name: "",
      password: "",
      confirmPassword: "",
      role: "viewer",
      departmentId: "",
    });
    setErrors({});
  };

  const handleInputChange = (field: string, value: string) => {
    setFormData(prev => ({ ...prev, [field]: value }));
    // Clear error when user starts typing
    if (errors[field]) {
      setErrors(prev => ({ ...prev, [field]: "" }));
    }
  };

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      {showTrigger && (
        <DialogTrigger asChild>
          <Button variant={triggerVariant} size={triggerSize} className="gap-2">
            <UserPlus className="h-4 w-4" />
            Add User
          </Button>
        </DialogTrigger>
      )}
      
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Create New User</DialogTitle>
          <DialogDescription>
            Add a new user to your organization and assign them to a department.
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4">
          {errors.general && (
            <div className="rounded-md bg-destructive/10 p-3 text-sm text-destructive">
              {errors.general}
            </div>
          )}

          <div className="space-y-2">
            <Label htmlFor="email">Email Address</Label>
            <Input
              id="email"
              type="email"
              placeholder="user@company.com"
              value={formData.email}
              onChange={(e) => handleInputChange("email", e.target.value)}
              className={errors.email ? "border-destructive" : ""}
            />
            {errors.email && (
              <p className="text-sm text-destructive">{errors.email}</p>
            )}
          </div>

          <div className="space-y-2">
            <Label htmlFor="name">Full Name</Label>
            <Input
              id="name"
              type="text"
              placeholder="John Doe"
              value={formData.name}
              onChange={(e) => handleInputChange("name", e.target.value)}
              className={errors.name ? "border-destructive" : ""}
            />
            {errors.name && (
              <p className="text-sm text-destructive">{errors.name}</p>
            )}
          </div>

          <div className="space-y-2">
            <Label htmlFor="password">Password</Label>
            <Input
              id="password"
              type="password"
              placeholder="Enter password"
              value={formData.password}
              onChange={(e) => handleInputChange("password", e.target.value)}
              className={errors.password ? "border-destructive" : ""}
            />
            {errors.password && (
              <p className="text-sm text-destructive">{errors.password}</p>
            )}
          </div>

          <div className="space-y-2">
            <Label htmlFor="confirmPassword">Confirm Password</Label>
            <Input
              id="confirmPassword"
              type="password"
              placeholder="Confirm password"
              value={formData.confirmPassword}
              onChange={(e) => handleInputChange("confirmPassword", e.target.value)}
              className={errors.confirmPassword ? "border-destructive" : ""}
            />
            {errors.confirmPassword && (
              <p className="text-sm text-destructive">{errors.confirmPassword}</p>
            )}
          </div>

          <div className="space-y-2">
            <Label htmlFor="role">Role</Label>
            <SelectField
              value={formData.role}
              onValueChange={(value) => handleInputChange("role", value)}
              placeholder="Select a role..."
              options={USER_ROLES}
              className={errors.role ? "border-destructive" : ""}
            />
            {errors.role && (
              <p className="text-sm text-destructive">{errors.role}</p>
            )}
          </div>

          <div className="space-y-2">
            <Label htmlFor="department">Department (Optional)</Label>
            <SelectField
              value={formData.departmentId}
              onValueChange={(value) => handleInputChange("departmentId", value)}
              placeholder="Select a department..."
              options={[
                { id: "", name: "No department" },
                ...departments.map((dept: any) => ({
                  id: dept.id,
                  name: dept.name
                }))
              ]}
              disabled={departmentsLoading}
            />
            {departmentsLoading && (
              <p className="text-sm text-muted-foreground">Loading departments...</p>
            )}
          </div>

          <div className="flex justify-end gap-3 pt-4">
            <Button
              type="button"
              variant="outline"
              onClick={handleClose}
              disabled={createUserMutation.isPending}
            >
              Cancel
            </Button>
            <Button
              type="submit"
              disabled={createUserMutation.isPending}
            >
              {createUserMutation.isPending ? "Creating..." : "Create User"}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}