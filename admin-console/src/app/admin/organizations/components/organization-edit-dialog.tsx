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
  updateOrganization,
  type Organization,
  type UpdateOrganizationRequest,
} from "@/app/_actions/organizations";
import { Building2, Mail, Phone, Globe, Settings, Users } from "lucide-react";

interface OrganizationEditDialogProps {
  organization: Organization | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onOrganizationUpdated: () => void;
}

const SUBSCRIPTION_TIERS = [
  {
    value: "basic",
    label: "Basic",
    description: "Essential features for small teams",
  },
  {
    value: "professional",
    label: "Professional",
    description: "Advanced features for growing businesses",
  },
  {
    value: "enterprise",
    label: "Enterprise",
    description: "Full feature set for large organizations",
  },
];

const AVAILABLE_FEATURES = [
  "advanced_analytics",
  "custom_workflows",
  "api_access",
  "sso_integration",
  "advanced_reporting",
  "custom_branding",
  "priority_support",
  "audit_logs",
  "data_export",
  "webhook_integration",
];

export function OrganizationEditDialog({
  organization,
  open,
  onOpenChange,
  onOrganizationUpdated,
}: OrganizationEditDialogProps) {
  const [isLoading, setIsLoading] = useState(false);
  const [formData, setFormData] = useState<UpdateOrganizationRequest>({});

  useEffect(() => {
    if (organization && open) {
      setFormData({
        name: organization.name,
        domain: organization.domain,
        subscription_tier: organization.subscription_tier,
        settings: {
          max_users: organization.settings?.max_users || 50,
          features_enabled: organization.settings?.features_enabled || [],
          custom_branding: organization.settings?.custom_branding || false,
          api_access: organization.settings?.api_access || false,
        },
        contact_info: {
          admin_name: organization.contact_info?.admin_name || "",
          admin_email: organization.contact_info?.admin_email || "",
          phone: organization.contact_info?.phone || "",
          address: organization.contact_info?.address || "",
        },
      });
    }
  }, [organization, open]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!organization) return;

    setIsLoading(true);
    try {
      const result = await updateOrganization(organization.id, formData);
      if (result.success) {
        toast.success("Organization updated successfully");
        onOrganizationUpdated();
        onOpenChange(false);
      } else {
        toast.error(result.message || "Failed to update organization");
      }
    } catch (error) {
      toast.error("Failed to update organization");
    } finally {
      setIsLoading(false);
    }
  };

  const handleInputChange = (
    field: keyof UpdateOrganizationRequest,
    value: any,
  ) => {
    setFormData((prev) => ({
      ...prev,
      [field]: value,
    }));
  };

  const handleSettingsChange = (field: string, value: any) => {
    setFormData((prev) => ({
      ...prev,
      settings: {
        ...prev.settings,
        [field]: value,
      },
    }));
  };

  const handleContactInfoChange = (field: string, value: string) => {
    setFormData((prev) => ({
      ...prev,
      contact_info: {
        ...prev.contact_info,
        [field]: value,
      },
    }));
  };

  const handleFeatureToggle = (feature: string, enabled: boolean) => {
    const currentFeatures = formData.settings?.features_enabled || [];
    const updatedFeatures = enabled
      ? [...currentFeatures, feature]
      : currentFeatures.filter((f) => f !== feature);

    handleSettingsChange("features_enabled", updatedFeatures);
  };

  if (!organization) return null;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-3xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Building2 className="h-5 w-5" />
            Edit Organization: {organization.name}
          </DialogTitle>
          <DialogDescription>
            Update organization information, settings, and features
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-6">
          {/* Basic Information */}
          <Card>
            <CardHeader>
              <CardTitle className="text-lg flex items-center gap-2">
                <Building2 className="h-4 w-4" />
                Basic Information
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid gap-4 md:grid-cols-2">
                <div className="space-y-2">
                  <Label htmlFor="name">Organization Name</Label>
                  <Input
                    id="name"
                    value={formData.name || ""}
                    onChange={(e) => handleInputChange("name", e.target.value)}
                    placeholder="Enter organization name"
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="domain">Domain</Label>
                  <Input
                    id="domain"
                    value={formData.domain || ""}
                    onChange={(e) =>
                      handleInputChange("domain", e.target.value)
                    }
                    placeholder="Enter domain"
                  />
                </div>
              </div>

              <div className="space-y-2">
                <Label htmlFor="subscription_tier">Subscription Tier</Label>
                <Select
                  value={formData.subscription_tier}
                  onValueChange={(value) =>
                    handleInputChange("subscription_tier", value as any)
                  }
                >
                  <SelectTrigger>
                    <SelectValue placeholder="Select subscription tier" />
                  </SelectTrigger>
                  <SelectContent>
                    {SUBSCRIPTION_TIERS.map((tier) => (
                      <SelectItem key={tier.value} value={tier.value}>
                        <div className="flex flex-col">
                          <span className="font-medium">{tier.label}</span>
                          <span className="text-sm text-muted-foreground">
                            {tier.description}
                          </span>
                        </div>
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </CardContent>
          </Card>

          {/* Settings */}
          <Card>
            <CardHeader>
              <CardTitle className="text-lg flex items-center gap-2">
                <Settings className="h-4 w-4" />
                Organization Settings
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="max_users">Maximum Users</Label>
                <Input
                  id="max_users"
                  type="number"
                  value={formData.settings?.max_users || ""}
                  onChange={(e) =>
                    handleSettingsChange(
                      "max_users",
                      parseInt(e.target.value) || 0,
                    )
                  }
                  placeholder="Enter maximum number of users"
                />
              </div>

              <div className="space-y-4">
                <div className="flex items-center justify-between">
                  <div className="space-y-0.5">
                    <Label htmlFor="custom_branding">Custom Branding</Label>
                    <p className="text-sm text-muted-foreground">
                      Allow organization to customize branding and appearance
                    </p>
                  </div>
                  <Switch
                    id="custom_branding"
                    checked={formData.settings?.custom_branding || false}
                    onCheckedChange={(checked) =>
                      handleSettingsChange("custom_branding", checked)
                    }
                  />
                </div>

                <div className="flex items-center justify-between">
                  <div className="space-y-0.5">
                    <Label htmlFor="api_access">API Access</Label>
                    <p className="text-sm text-muted-foreground">
                      Enable API access for this organization
                    </p>
                  </div>
                  <Switch
                    id="api_access"
                    checked={formData.settings?.api_access || false}
                    onCheckedChange={(checked) =>
                      handleSettingsChange("api_access", checked)
                    }
                  />
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Features */}
          <Card>
            <CardHeader>
              <CardTitle className="text-lg">Enabled Features</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="grid gap-3 md:grid-cols-2">
                {AVAILABLE_FEATURES.map((feature) => {
                  const isEnabled =
                    formData.settings?.features_enabled?.includes(feature) ||
                    false;
                  return (
                    <div
                      key={feature}
                      className="flex items-center justify-between"
                    >
                      <div className="space-y-0.5">
                        <Label className="text-sm font-medium">
                          {feature
                            .replace(/_/g, " ")
                            .replace(/\b\w/g, (l) => l.toUpperCase())}
                        </Label>
                      </div>
                      <Switch
                        checked={isEnabled}
                        onCheckedChange={(checked) =>
                          handleFeatureToggle(feature, checked)
                        }
                      />
                    </div>
                  );
                })}
              </div>
            </CardContent>
          </Card>

          {/* Contact Information */}
          <Card>
            <CardHeader>
              <CardTitle className="text-lg flex items-center gap-2">
                <Mail className="h-4 w-4" />
                Contact Information
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid gap-4 md:grid-cols-2">
                <div className="space-y-2">
                  <Label htmlFor="admin_name">Admin Name</Label>
                  <Input
                    id="admin_name"
                    value={formData.contact_info?.admin_name || ""}
                    onChange={(e) =>
                      handleContactInfoChange("admin_name", e.target.value)
                    }
                    placeholder="Enter admin name"
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="admin_email">Admin Email</Label>
                  <Input
                    id="admin_email"
                    type="email"
                    value={formData.contact_info?.admin_email || ""}
                    onChange={(e) =>
                      handleContactInfoChange("admin_email", e.target.value)
                    }
                    placeholder="Enter admin email"
                  />
                </div>
              </div>

              <div className="space-y-2">
                <Label htmlFor="phone">Phone Number</Label>
                <Input
                  id="phone"
                  value={formData.contact_info?.phone || ""}
                  onChange={(e) =>
                    handleContactInfoChange("phone", e.target.value)
                  }
                  placeholder="Enter phone number"
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="address">Address</Label>
                <Textarea
                  id="address"
                  value={formData.contact_info?.address || ""}
                  onChange={(e) =>
                    handleContactInfoChange("address", e.target.value)
                  }
                  placeholder="Enter organization address"
                  rows={3}
                />
              </div>
            </CardContent>
          </Card>

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={isLoading}>
              {isLoading ? "Updating..." : "Update Organization"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
