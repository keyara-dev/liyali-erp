"use client";

import { useState, useEffect } from "react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Switch } from "@/components/ui/switch";
import { Separator } from "@/components/ui/separator";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { toast } from "sonner";
import {
  updateUser,
  type PlatformUser,
  type UpdateUserRequest,
} from "@/app/_actions/users";
import { User, Mail, Phone, Building2, Shield } from "lucide-react";

interface UserEditDialogProps {
  user: PlatformUser | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onUserUpdated: () => void;
}

const USER_ROLES = [
  { value: "admin", label: "Admin" },
  { value: "manager", label: "Manager" },
  { value: "user", label: "User" },
  { value: "viewer", label: "Viewer" },
];

const USER_STATUSES = [
  { value: "active", label: "Active" },
  { value: "suspended", label: "Suspended" },
  { value: "inactive", label: "Inactive" },
  { value: "pending", label: "Pending" },
];

export function UserEditDialog({
  user,
  open,
  onOpenChange,
  onUserUpdated,
}: UserEditDialogProps) {
  const [isLoading, setIsLoading] = useState(false);
  const [formData, setFormData] = useState<UpdateUserRequest>({});

  useEffect(() => {
    if (user && open) {
      setFormData({
        name: user.name,
        email: user.email,
        role: user.role,
        status: user.status,
        phone: user.phone || "",
        profile: {
          department: user.profile?.department || "",
          job_title: user.profile?.job_title || "",
          phone: user.profile?.phone || user.phone || "",
        },
      });
    }
  }, [user, open]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!user) return;

    setIsLoading(true);
    try {
      const result = await updateUser(user.id, formData);
      if (result.success) {
        toast.success("User updated successfully");
        onUserUpdated();
        onOpenChange(false);
      } else {
        toast.error(result.message || "Failed to update user");
      }
    } catch (error) {
      toast.error("Failed to update user");
    } finally {
      setIsLoading(false);
    }
  };

  const handleInputChange = (field: keyof UpdateUserRequest, value: any) => {
    setFormData((prev) => ({
      ...prev,
      [field]: value,
    }));
  };

  const handleProfileChange = (field: string, value: string) => {
    setFormData((prev) => ({
      ...prev,
      profile: {
        ...prev.profile,
        [field]: value,
      },
    }));
  };

  if (!user) return null;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <User className="h-5 w-5" />
            Edit User: {user.name}
          </DialogTitle>
          <DialogDescription>
            Update user information and permissions
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-6">
          {/* Basic Information */}
          <Card>
            <CardHeader>
              <CardTitle className="text-lg flex items-center gap-2">
                <User className="h-4 w-4" />
                Basic Information
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid gap-4 md:grid-cols-2">
                <div className="space-y-2">
                  <Label htmlFor="name">Full Name</Label>
                  <Input
                    id="name"
                    value={formData.name || ""}
                    onChange={(e) => handleInputChange("name", e.target.value)}
                    placeholder="Enter full name"
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="email">Email Address</Label>
                  <Input
                    id="email"
                    type="email"
                    value={formData.email || ""}
                    onChange={(e) => handleInputChange("email", e.target.value)}
                    placeholder="Enter email address"
                  />
                </div>
              </div>

              <div className="grid gap-4 md:grid-cols-2">
                <div className="space-y-2">
                  <Label htmlFor="phone">Phone Number</Label>
                  <Input
                    id="phone"
                    value={formData.phone || ""}
                    onChange={(e) => handleInputChange("phone", e.target.value)}
                    placeholder="Enter phone number"
                  />
                </div>
                <div className="space-y-2">
                  <Label>Email Verified</Label>
                  <div className="flex items-center gap-2">
                    <Badge
                      variant={user.email_verified ? "default" : "secondary"}
                    >
                      {user.email_verified ? "Verified" : "Not Verified"}
                    </Badge>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Role and Status */}
          <Card>
            <CardHeader>
              <CardTitle className="text-lg flex items-center gap-2">
                <Shield className="h-4 w-4" />
                Role and Status
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid gap-4 md:grid-cols-2">
                <div className="space-y-2">
                  <Label htmlFor="role">Role</Label>
                  <Select
                    value={formData.role}
                    onValueChange={(value) => handleInputChange("role", value)}
                  >
                    <SelectTrigger>
                      <SelectValue placeholder="Select role" />
                    </SelectTrigger>
                    <SelectContent>
                      {USER_ROLES.map((role) => (
                        <SelectItem key={role.value} value={role.value}>
                          {role.label}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
                <div className="space-y-2">
                  <Label htmlFor="status">Status</Label>
                  <Select
                    value={formData.status}
                    onValueChange={(value) =>
                      handleInputChange("status", value as any)
                    }
                  >
                    <SelectTrigger>
                      <SelectValue placeholder="Select status" />
                    </SelectTrigger>
                    <SelectContent>
                      {USER_STATUSES.map((status) => (
                        <SelectItem key={status.value} value={status.value}>
                          {status.label}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Profile Information */}
          <Card>
            <CardHeader>
              <CardTitle className="text-lg flex items-center gap-2">
                <Building2 className="h-4 w-4" />
                Profile Information
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid gap-4 md:grid-cols-2">
                <div className="space-y-2">
                  <Label htmlFor="department">Department</Label>
                  <Input
                    id="department"
                    value={formData.profile?.department || ""}
                    onChange={(e) =>
                      handleProfileChange("department", e.target.value)
                    }
                    placeholder="Enter department"
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="job_title">Job Title</Label>
                  <Input
                    id="job_title"
                    value={formData.profile?.job_title || ""}
                    onChange={(e) =>
                      handleProfileChange("job_title", e.target.value)
                    }
                    placeholder="Enter job title"
                  />
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Organization Memberships */}
          {user.organizations.length > 0 && (
            <Card>
              <CardHeader>
                <CardTitle className="text-lg flex items-center gap-2">
                  <Building2 className="h-4 w-4" />
                  Organization Memberships ({user.organizations.length})
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-3">
                  {user.organizations.map((org) => (
                    <div
                      key={org.organization_id}
                      className="flex items-center justify-between rounded-lg border p-3"
                    >
                      <div>
                        <div className="flex items-center gap-2">
                          <span className="font-medium">
                            {org.organization_name}
                          </span>
                          {org.is_primary && (
                            <Badge variant="default" className="text-xs">
                              Primary
                            </Badge>
                          )}
                        </div>
                        <div className="text-sm text-muted-foreground">
                          Role: {org.role} • Status: {org.status}
                        </div>
                      </div>
                      <Badge
                        variant={
                          org.status === "active" ? "default" : "secondary"
                        }
                      >
                        {org.status}
                      </Badge>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          )}

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={isLoading}>
              {isLoading ? "Updating..." : "Update User"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
