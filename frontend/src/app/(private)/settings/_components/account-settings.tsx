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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Switch } from "@/components/ui/switch";
import { UserAvatarUpload } from "@/components/ui/user-avatar-upload";
import { ChangePasswordModal } from "./change-password";
import { updateAccountSettings } from "@/app/_actions/settings";
import { User } from "@/types/auth";
import { AlertCircle, CheckCircle, Lock } from "lucide-react";
import { toast } from "sonner";

interface AccountSettingsProps {
  user: User | null;
  onProfileUpdate?: (updatedUser: User) => void;
}

export function AccountSettings({
  user,
  onProfileUpdate,
}: AccountSettingsProps) {
  const prefs = user?.preferences;

  const [isLoading, setIsLoading] = useState(false);
  const [formData, setFormData] = useState({
    // Profile
    name: user?.name || "",
    email: user?.email || "",
    // Preferences
    avatar: prefs?.avatar || "",
    department: prefs?.department || "",
    language: prefs?.language || "en",
    theme: prefs?.theme || "system",
    timezone: prefs?.timezone || "Africa/Lusaka",
    emailNotifications: prefs?.emailNotifications ?? true,
    pushNotifications: prefs?.pushNotifications ?? false,
    activityNotifications: prefs?.activityNotifications ?? true,
  });

  useEffect(() => {
    if (user) {
      const p = user.preferences;
      setFormData({
        name: user.name || "",
        email: user.email || "",
        avatar: p?.avatar || "",
        department: p?.department || "",
        language: p?.language || "en",
        theme: p?.theme || "system",
        timezone: p?.timezone || "Africa/Lusaka",
        emailNotifications: p?.emailNotifications ?? true,
        pushNotifications: p?.pushNotifications ?? false,
        activityNotifications: p?.activityNotifications ?? true,
      });
    }
  }, [user]);

  const set = (field: string, value: string | boolean) =>
    setFormData((prev) => ({ ...prev, [field]: value }));

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);
    try {
      const result = await updateAccountSettings({
        name: formData.name,
        email: formData.email,
        preferences: {
          avatar: formData.avatar,
          department: formData.department,
          language: formData.language,
          theme: formData.theme,
          timezone: formData.timezone,
          emailNotifications: formData.emailNotifications,
          pushNotifications: formData.pushNotifications,
          activityNotifications: formData.activityNotifications,
        },
      });

      if (result.success) {
        toast.success("Settings saved successfully");
        if (result.data && onProfileUpdate) {
          onProfileUpdate(result.data as User);
        }
      } else {
        toast.error(result.message || "Failed to save settings");
      }
    } catch {
      toast.error("An error occurred while saving settings");
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>Account Information</CardTitle>
        <CardDescription>
          Manage your profile, preferences, and notification settings
        </CardDescription>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit} className="space-y-6">
          {/* Avatar */}
          <div className="space-y-2">
            <Label>Profile Picture</Label>
            <UserAvatarUpload
              currentAvatarUrl={formData.avatar}
              userName={formData.name || "User"}
              onAvatarChange={(url) => set("avatar", url)}
              disabled={isLoading}
              size="lg"
            />
          </div>

          {/* Profile fields */}
          <div className="grid gap-4 md:grid-cols-2">
            <Input
              label="Full Name"
              id="name"
              placeholder="Enter your full name"
              value={formData.name}
              onChange={(e) => set("name", e.target.value)}
              disabled={isLoading}
            />
            <Input
              label="Email Address"
              id="email"
              type="email"
              placeholder="Enter your email address"
              value={formData.email}
              onChange={(e) => set("email", e.target.value)}
              disabled={isLoading}
              descriptionText="Used for account notifications and password recovery"
            />
            <Input
              label="Department"
              id="department"
              placeholder="Your department"
              value={formData.department}
              onChange={(e) => set("department", e.target.value)}
              disabled={isLoading}
            />
            <Input
              label="Role"
              id="role"
              value={user?.role || "N/A"}
              disabled
              className="cursor-not-allowed"
              descriptionText="Your role is managed by administrators"
            />
          </div>

          {/* Password row */}
          <div className="flex items-center justify-between py-3 border-t border-b">
            <div className="flex items-center gap-2">
              <Lock className="h-4 w-4 text-muted-foreground" />
              <div>
                <p className="text-sm font-medium">Password</p>
                <p className="text-xs text-muted-foreground">
                  Update your account password
                </p>
              </div>
            </div>
            <ChangePasswordModal />
          </div>

          {/* Preferences */}
          <div className="space-y-4">
            <p className="text-sm font-medium text-muted-foreground uppercase tracking-wide">
              Preferences
            </p>

            {/* Display selects */}
            <div className="grid gap-4 md:grid-cols-3">
              <div className="space-y-2">
                <Label htmlFor="language">Language</Label>
                <Select
                  value={formData.language}
                  onValueChange={(v) => set("language", v)}
                  disabled={isLoading}
                >
                  <SelectTrigger id="language">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="en">English</SelectItem>
                    <SelectItem value="es">Español</SelectItem>
                    <SelectItem value="fr">Français</SelectItem>
                    <SelectItem value="pt">Português</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              <div className="space-y-2">
                <Label htmlFor="theme">Theme</Label>
                <Select
                  value={formData.theme}
                  onValueChange={(v) => set("theme", v)}
                  disabled={isLoading}
                >
                  <SelectTrigger id="theme">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="light">Light</SelectItem>
                    <SelectItem value="dark">Dark</SelectItem>
                    <SelectItem value="system">System Default</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              <div className="space-y-2">
                <Label htmlFor="timezone">Timezone</Label>
                <Select
                  value={formData.timezone}
                  onValueChange={(v) => set("timezone", v)}
                  disabled={isLoading}
                >
                  <SelectTrigger id="timezone">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="Africa/Lusaka">Africa/Lusaka (CAT)</SelectItem>
                    <SelectItem value="UTC">UTC</SelectItem>
                    <SelectItem value="Europe/London">Europe/London (GMT)</SelectItem>
                    <SelectItem value="America/New_York">America/New_York (EST)</SelectItem>
                    <SelectItem value="America/Los_Angeles">America/Los_Angeles (PST)</SelectItem>
                    <SelectItem value="Asia/Tokyo">Asia/Tokyo (JST)</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>

            {/* Notifications */}
            <div className="space-y-2">
              <div className="flex items-center justify-between p-3 rounded-lg border">
                <div>
                  <p className="font-medium text-sm">Email Notifications</p>
                  <p className="text-xs text-muted-foreground">
                    Receive important updates via email
                  </p>
                </div>
                <Switch
                  checked={formData.emailNotifications}
                  onCheckedChange={(v) => set("emailNotifications", v)}
                  disabled={isLoading}
                />
              </div>
              <div className="flex items-center justify-between p-3 rounded-lg border">
                <div>
                  <p className="font-medium text-sm">Push Notifications</p>
                  <p className="text-xs text-muted-foreground">
                    Receive instant notifications on your device
                  </p>
                </div>
                <Switch
                  checked={formData.pushNotifications}
                  onCheckedChange={(v) => set("pushNotifications", v)}
                  disabled={isLoading}
                />
              </div>
              <div className="flex items-center justify-between p-3 rounded-lg border">
                <div>
                  <p className="font-medium text-sm">Activity Notifications</p>
                  <p className="text-xs text-muted-foreground">
                    Get notified about workflow activities and approvals
                  </p>
                </div>
                <Switch
                  checked={formData.activityNotifications}
                  onCheckedChange={(v) => set("activityNotifications", v)}
                  disabled={isLoading}
                />
              </div>
            </div>
          </div>

          {/* Save */}
          <div className="flex justify-end">
            <Button
              type="submit"
              disabled={isLoading}
              isLoading={isLoading}
              loadingText="Saving..."
            >
              Save Changes
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>
  );
}
