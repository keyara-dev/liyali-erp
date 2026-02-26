"use client";

import { useState, useEffect } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { UserAvatarUpload } from "@/components/ui/user-avatar-upload";
import { updateUserProfile } from "@/app/_actions/settings";
import { User } from "@/types/auth";
import { AlertCircle, CheckCircle } from "lucide-react";

interface AccountSettingsProps {
  user: User | null;
  onProfileUpdate?: (updatedUser: User) => void;
}

export function AccountSettings({
  user,
  onProfileUpdate,
}: AccountSettingsProps) {
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);
  const [formData, setFormData] = useState({
    name: user?.name || "",
    email: user?.email || "",
    department: user?.department || "",
    avatar: user?.avatar || "",
  });

  // Sync form data when user changes
  useEffect(() => {
    if (user) {
      setFormData({
        name: user.name || "",
        email: user.email || "",
        department: user.department || "",
        avatar: user.avatar || "",
      });
    }
  }, [user]);

  const handleInputChange = (field: string, value: string) => {
    setFormData((prev) => ({
      ...prev,
      [field]: value,
    }));
  };

  const handleAvatarChange = (url: string) => {
    setFormData((prev) => ({
      ...prev,
      avatar: url,
    }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setSuccess(null);
    setIsLoading(true);

    try {
      const result = await updateUserProfile({
        name: formData.name,
        email: formData.email,
        department: formData.department,
        avatar: formData.avatar,
      });

      if (result.success) {
        setSuccess("Profile updated successfully");
        if (result.data && onProfileUpdate) {
          onProfileUpdate(result.data);
        }
      } else {
        setError(result.message || "Failed to update profile");
      }
    } catch (err) {
      setError("An error occurred while updating profile");
      console.error(err);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>Account Information</CardTitle>
        <CardDescription>
          Manage your personal account details and profile information
        </CardDescription>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit} className="space-y-6">
          {error && (
            <div className="flex items-center gap-2 p-3 rounded-lg bg-red-50 text-red-700 border border-red-200">
              <AlertCircle className="h-4 w-4 flex-shrink-0" />
              <p className="text-sm">{error}</p>
            </div>
          )}

          {success && (
            <div className="flex items-center gap-2 p-3 rounded-lg bg-green-50 text-green-700 border border-green-200">
              <CheckCircle className="h-4 w-4 flex-shrink-0" />
              <p className="text-sm">{success}</p>
            </div>
          )}

          {/* Avatar Upload */}
          <div className="space-y-2">
            <Label>Profile Picture</Label>
            <UserAvatarUpload
              currentAvatarUrl={formData.avatar}
              userName={formData.name || "User"}
              onAvatarChange={handleAvatarChange}
              disabled={isLoading}
              size="lg"
            />
          </div>

          <Input
            label="Full Name"
            id="name"
            placeholder="Enter your full name"
            value={formData.name}
            onChange={(e) => handleInputChange("name", e.target.value)}
            disabled={isLoading}
          />

          <Input
            label="Email Address"
            id="email"
            type="email"
            placeholder="Enter your email address"
            value={formData.email}
            onChange={(e) => handleInputChange("email", e.target.value)}
            disabled={isLoading}
          />
          <p className="text-xs text-muted-foreground">
            We'll use this email for account notifications and password
            recovery
          </p>

          <Input
            label="Department"
            id="department"
            placeholder="Your department"
            value={formData.department}
            onChange={(e) => handleInputChange("department", e.target.value)}
            disabled={isLoading}
          />

          <Input
            label="Role"
            id="role"
            value={user?.role || "N/A"}
            disabled
            className="cursor-not-allowed"
          />
          <p className="text-xs text-muted-foreground">
            Your role is managed by administrators
          </p>

          <div className="flex gap-3 pt-4">
            <Button type="submit" disabled={isLoading} isLoading={isLoading} loadingText="Saving...">
              Save Changes
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>
  );
}
