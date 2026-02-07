"use client";

import { useState } from "react";
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
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { toast } from "sonner";
import {
  createOrganization,
  type CreateOrganizationRequest,
} from "@/app/_actions/organizations";
import { Building2, Mail, User, Settings, Clock } from "lucide-react";

interface OrganizationCreateDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onOrganizationCreated: () => void;
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

const TRIAL_DURATIONS = [
  { value: 7, label: "7 days" },
  { value: 14, label: "14 days" },
  { value: 30, label: "30 days" },
  { value: 60, label: "60 days" },
  { value: 90, label: "90 days" },
];

export function OrganizationCreateDialog({
  open,
  onOpenChange,
  onOrganizationCreated,
}: OrganizationCreateDialogProps) {
  const [isLoading, setIsLoading] = useState(false);
  const [formData, setFormData] = useState<CreateOrganizationRequest>({
    name: "",
    domain: "",
    admin_name: "",
    admin_email: "",
    subscription_tier: "basic",
    trial_days: 30,
    settings: {
      max_users: 50,
      features_enabled: [],
    },
    contact_info: {
      phone: "",
      address: "",
    },
  });

  const [errors, setErrors] = useState<Record<string, string>>({});

  const validateForm = (): boolean => {
    const newErrors: Record<string, string> = {};

    if (!formData.name.trim()) {
      newErrors.name = "Organization name is required";
    }

    if (!formData.domain.trim()) {
      newErrors.domain = "Domain is required";
    } else if (
      !/^[a-zA-Z0-9][a-zA-Z0-9-]*[a-zA-Z0-9]*\.[a-zA-Z]{2,}$/.test(
        formData.domain,
      )
    ) {
      newErrors.domain = "Please enter a valid domain (e.g., company.com)";
    }

    if (!formData.admin_name.trim()) {
      newErrors.admin_name = "Admin name is required";
    }

    if (!formData.admin_email.trim()) {
      newErrors.admin_email = "Admin email is required";
    } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(formData.admin_email)) {
      newErrors.admin_email = "Please enter a valid email address";
    }

    if (!formData.subscription_tier) {
      newErrors.subscription_tier = "Subscription tier is required";
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
      const result = await createOrganization(formData);
      if (result.success) {
        toast.success("Organization created successfully");
        onOrganizationCreated();
        onOpenChange(false);
        resetForm();
      } else {
        toast.error(result.message || "Failed to create organization");
      }
    } catch (error) {
      toast.error("Failed to create organization");
    } finally {
      setIsLoading(false);
    }
  };

  const resetForm = () => {
    setFormData({
      name: "",
      domain: "",
      admin_name: "",
      admin_email: "",
      subscription_tier: "basic",
      trial_days: 30,
      settings: {
        max_users: 50,
        features_enabled: [],
      },
      contact_info: {
        phone: "",
        address: "",
      },
    });
    setErrors({});
  };

  const handleInputChange = (
    field: keyof CreateOrganizationRequest,
    value: any,
  ) => {
    setFormData((prev) => ({
      ...prev,
      [field]: value,
    }));

    // Clear error when user starts typing
    if (errors[field]) {
      setErrors((prev) => ({
        ...prev,
        [field]: "",
      }));
    }
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

  const selectedTier = SUBSCRIPTION_TIERS.find(
    (tier) => tier.value === formData.subscription_tier,
  );

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Building2 className="h-5 w-5" />
            Create New Organization
          </DialogTitle>
          <DialogDescription>
            Set up a new organization with admin user and initial configuration
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-6">
          {/* Basic Information */}
          <Card>
            <CardHeader>
              <CardTitle className="text-lg flex items-center gap-2">
                <Building2 className="h-4 w-4" />
                Organization Information
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid gap-4 md:grid-cols-2">
                <div className="space-y-2">
                  <Label htmlFor="name">
                    Organization Name <span className="text-red-500">*</span>
                  </Label>
                  <Input
                    id="name"
                    value={formData.name}
                    onChange={(e) => handleInputChange("name", e.target.value)}
                    placeholder="Enter organization name"
                    className={errors.name ? "border-red-500" : ""}
                  />
                  {errors.name && (
                    <p className="text-sm text-red-500">{errors.name}</p>
                  )}
                </div>
                <div className="space-y-2">
                  <Label htmlFor="domain">
                    Domain <span className="text-red-500">*</span>
                  </Label>
                  <Input
                    id="domain"
                    value={formData.domain}
                    onChange={(e) =>
                      handleInputChange("domain", e.target.value)
                    }
                    placeholder="company.com"
                    className={errors.domain ? "border-red-500" : ""}
                  />
                  {errors.domain && (
                    <p className="text-sm text-red-500">{errors.domain}</p>
                  )}
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Admin User */}
          <Card>
            <CardHeader>
              <CardTitle className="text-lg flex items-center gap-2">
                <User className="h-4 w-4" />
                Admin User
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid gap-4 md:grid-cols-2">
                <div className="space-y-2">
                  <Label htmlFor="admin_name">
                    Admin Name <span className="text-red-500">*</span>
                  </Label>
                  <Input
                    id="admin_name"
                    value={formData.admin_name}
                    onChange={(e) =>
                      handleInputChange("admin_name", e.target.value)
                    }
                    placeholder="Enter admin full name"
                    className={errors.admin_name ? "border-red-500" : ""}
                  />
                  {errors.admin_name && (
                    <p className="text-sm text-red-500">{errors.admin_name}</p>
                  )}
                </div>
                <div className="space-y-2">
                  <Label htmlFor="admin_email">
                    Admin Email <span className="text-red-500">*</span>
                  </Label>
                  <Input
                    id="admin_email"
                    type="email"
                    value={formData.admin_email}
                    onChange={(e) =>
                      handleInputChange("admin_email", e.target.value)
                    }
                    placeholder="admin@company.com"
                    className={errors.admin_email ? "border-red-500" : ""}
                  />
                  {errors.admin_email && (
                    <p className="text-sm text-red-500">{errors.admin_email}</p>
                  )}
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Subscription & Trial */}
          <Card>
            <CardHeader>
              <CardTitle className="text-lg flex items-center gap-2">
                <Settings className="h-4 w-4" />
                Subscription & Trial
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid gap-4 md:grid-cols-2">
                <div className="space-y-2">
                  <Label htmlFor="subscription_tier">
                    Subscription Tier <span className="text-red-500">*</span>
                  </Label>
                  <Select
                    value={formData.subscription_tier}
                    onValueChange={(value) =>
                      handleInputChange("subscription_tier", value as any)
                    }
                  >
                    <SelectTrigger
                      className={
                        errors.subscription_tier ? "border-red-500" : ""
                      }
                    >
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
                  {errors.subscription_tier && (
                    <p className="text-sm text-red-500">
                      {errors.subscription_tier}
                    </p>
                  )}
                  {selectedTier && (
                    <p className="text-sm text-muted-foreground">
                      {selectedTier.description}
                    </p>
                  )}
                </div>

                <div className="space-y-2">
                  <Label htmlFor="trial_days">Trial Duration</Label>
                  <Select
                    value={formData.trial_days?.toString()}
                    onValueChange={(value) =>
                      handleInputChange("trial_days", parseInt(value))
                    }
                  >
                    <SelectTrigger>
                      <SelectValue placeholder="Select trial duration" />
                    </SelectTrigger>
                    <SelectContent>
                      {TRIAL_DURATIONS.map((duration) => (
                        <SelectItem
                          key={duration.value}
                          value={duration.value.toString()}
                        >
                          {duration.label}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
              </div>

              <div className="space-y-2">
                <Label htmlFor="max_users">Maximum Users</Label>
                <Input
                  id="max_users"
                  type="number"
                  value={formData.settings?.max_users || ""}
                  onChange={(e) =>
                    handleSettingsChange(
                      "max_users",
                      parseInt(e.target.value) || 50,
                    )
                  }
                  placeholder="50"
                />
              </div>
            </CardContent>
          </Card>

          {/* Contact Information */}
          <Card>
            <CardHeader>
              <CardTitle className="text-lg flex items-center gap-2">
                <Mail className="h-4 w-4" />
                Contact Information (Optional)
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
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
              onClick={() => {
                onOpenChange(false);
                resetForm();
              }}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={isLoading}>
              {isLoading ? "Creating..." : "Create Organization"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
