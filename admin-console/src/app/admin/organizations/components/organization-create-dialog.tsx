"use client";

import { useState, useMemo } from "react";
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
import { Badge } from "@/components/ui/badge";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "@/components/ui/command";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { toast } from "sonner";
import {
  createOrganization,
  type CreateOrganizationRequest,
} from "@/app/_actions/organizations";
import { useSubscriptionTiers } from "@/hooks/use-subscriptions";
import { useAdminUsers } from "@/hooks/use-admin-users";
import {
  Building2,
  User,
  Settings,
  ChevronsUpDown,
  Check,
  X,
  Loader2,
} from "lucide-react";
import { cn } from "@/lib/utils";
import type { AdminUser } from "@/app/_actions/admin-users";

interface OrganizationCreateDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onOrganizationCreated: () => void;
}

const TRIAL_DURATIONS = [
  { value: 7, label: "7 days" },
  { value: 14, label: "14 days" },
  { value: 30, label: "30 days" },
  { value: 60, label: "60 days" },
  { value: 90, label: "90 days" },
];

const emptyForm = (): CreateOrganizationRequest => ({
  name: "",
  domain: "",
  admin_user_id: "",
  subscription_tier: "",
  trial_days: 30,
  max_users: 50,
});

export function OrganizationCreateDialog({
  open,
  onOpenChange,
  onOrganizationCreated,
}: OrganizationCreateDialogProps) {
  const [isLoading, setIsLoading] = useState(false);
  const [formData, setFormData] =
    useState<CreateOrganizationRequest>(emptyForm());
  const [errors, setErrors] = useState<Record<string, string>>({});

  // User search-select state
  const [userPickerOpen, setUserPickerOpen] = useState(false);
  const [userSearch, setUserSearch] = useState("");
  const [selectedUser, setSelectedUser] = useState<AdminUser | null>(null);

  const { data: tiers, isLoading: tiersLoading } = useSubscriptionTiers();
  const { data: adminUsers, isLoading: usersLoading } = useAdminUsers(
    userSearch ? { search: userSearch } : {},
  );

  const selectedTier = tiers?.find(
    (t) => t.name === formData.subscription_tier,
  );

  const validateForm = (): boolean => {
    const newErrors: Record<string, string> = {};
    if (!formData.name.trim()) newErrors.name = "Organization name is required";
    if (!formData.domain.trim()) {
      newErrors.domain = "Domain is required";
    } else if (
      !/^[a-zA-Z0-9][a-zA-Z0-9-]*[a-zA-Z0-9]*\.[a-zA-Z]{2,}$/.test(
        formData.domain,
      )
    ) {
      newErrors.domain = "Please enter a valid domain (e.g., company.com)";
    }
    if (!formData.admin_user_id)
      newErrors.admin_user_id = "Please select an admin user";
    if (!formData.subscription_tier)
      newErrors.subscription_tier = "Subscription tier is required";
    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!validateForm()) return;

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
    } catch {
      toast.error("Failed to create organization");
    } finally {
      setIsLoading(false);
    }
  };

  const resetForm = () => {
    setFormData(emptyForm());
    setErrors({});
    setSelectedUser(null);
    setUserSearch("");
  };

  const handleField = (field: keyof CreateOrganizationRequest, value: any) => {
    setFormData((prev) => ({ ...prev, [field]: value }));
    if (errors[field]) setErrors((prev) => ({ ...prev, [field]: "" }));
  };

  const handleSelectUser = (user: AdminUser) => {
    setSelectedUser(user);
    handleField("admin_user_id", user.id);
    setUserPickerOpen(false);
    setUserSearch("");
  };

  const handleClearUser = () => {
    setSelectedUser(null);
    handleField("admin_user_id", "");
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl! max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Building2 className="h-5 w-5" />
            Create New Organization
          </DialogTitle>
          <DialogDescription asChild>
            <div>
              Set up a new organization and attach an existing admin user.
            </div>
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-6">
          {/* Organization Information */}
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
                    onChange={(e) => handleField("name", e.target.value)}
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
                    onChange={(e) => handleField("domain", e.target.value)}
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
                Admin User <span className="text-red-500">*</span>
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              <p className="text-sm text-muted-foreground">
                Select an existing admin user to manage this organization.
              </p>

              {selectedUser ? (
                /* Selected user pill */
                <div className="flex items-center gap-3 p-3 rounded-lg border bg-muted/30">
                  <div className="flex h-9 w-9 items-center justify-center rounded-full bg-primary/10 text-primary font-semibold text-sm shrink-0">
                    {(selectedUser.full_name || selectedUser.email)
                      .charAt(0)
                      .toUpperCase()}
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="font-medium text-sm truncate">
                      {selectedUser.full_name ||
                        `${selectedUser.first_name} ${selectedUser.last_name}`.trim() ||
                        selectedUser.email}
                    </div>
                    <div className="text-xs text-muted-foreground truncate">
                      {selectedUser.email}
                    </div>
                  </div>
                  {selectedUser.is_super_admin && (
                    <Badge variant="secondary" className="text-xs shrink-0">
                      Super Admin
                    </Badge>
                  )}
                  <Button
                    type="button"
                    variant="ghost"
                    size="sm"
                    className="h-7 w-7 p-0 shrink-0"
                    onClick={handleClearUser}
                  >
                    <X className="h-4 w-4" />
                  </Button>
                </div>
              ) : (
                /* Search popover */
                <Popover
                  open={userPickerOpen}
                  onOpenChange={setUserPickerOpen}
                  modal={true}
                >
                  <PopoverTrigger asChild>
                    <Button
                      type="button"
                      variant="outline"
                      role="combobox"
                      className={cn(
                        "w-full justify-between font-normal",
                        errors.admin_user_id && "border-red-500",
                      )}
                    >
                      <span className="text-muted-foreground">
                        Search and select an admin user...
                      </span>
                      <ChevronsUpDown className="h-4 w-4 shrink-0 opacity-50" />
                    </Button>
                  </PopoverTrigger>
                  <PopoverContent
                    className="w-[--radix-popover-trigger-width] p-0"
                    align="start"
                  >
                    <Command shouldFilter={false}>
                      <CommandInput
                        placeholder="Search by name or email..."
                        value={userSearch}
                        onValueChange={setUserSearch}
                      />
                      <CommandList>
                        {usersLoading ? (
                          <div className="flex items-center justify-center py-6 text-sm text-muted-foreground gap-2">
                            <Loader2 className="h-4 w-4 animate-spin" />
                            Searching...
                          </div>
                        ) : !adminUsers?.length ? (
                          <CommandEmpty>No admin users found.</CommandEmpty>
                        ) : (
                          <CommandGroup>
                            {adminUsers.map((user) => (
                              <CommandItem
                                key={user.id}
                                value={user.id}
                                onSelect={() => handleSelectUser(user)}
                                className="flex items-center gap-3 py-2"
                              >
                                <div className="flex h-8 w-8 items-center justify-center rounded-full bg-primary/10 text-primary font-semibold text-xs shrink-0">
                                  {(user.full_name || user.email)
                                    .charAt(0)
                                    .toUpperCase()}
                                </div>
                                <div className="flex-1 min-w-0">
                                  <div className="text-sm font-medium truncate">
                                    {user.full_name ||
                                      `${user.first_name} ${user.last_name}`.trim() ||
                                      user.email}
                                  </div>
                                  <div className="text-xs text-muted-foreground truncate">
                                    {user.email}
                                  </div>
                                </div>
                                {user.is_super_admin && (
                                  <Badge
                                    variant="secondary"
                                    className="text-xs shrink-0"
                                  >
                                    Super Admin
                                  </Badge>
                                )}
                                <Check
                                  className={cn(
                                    "h-4 w-4 shrink-0",
                                    "opacity-0",
                                  )}
                                />
                              </CommandItem>
                            ))}
                          </CommandGroup>
                        )}
                      </CommandList>
                    </Command>
                  </PopoverContent>
                </Popover>
              )}

              {errors.admin_user_id && (
                <p className="text-sm text-red-500">{errors.admin_user_id}</p>
              )}
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
              <div className="space-y-2">
                <Label htmlFor="subscription_tier">
                  Subscription Tier <span className="text-red-500">*</span>
                </Label>
                <Select
                  value={formData.subscription_tier}
                  onValueChange={(v) => handleField("subscription_tier", v)}
                  disabled={tiersLoading}
                >
                  <SelectTrigger
                    className={errors.subscription_tier ? "border-red-500" : ""}
                  >
                    <SelectValue
                      placeholder={
                        tiersLoading
                          ? "Loading tiers..."
                          : "Select subscription tier"
                      }
                    />
                  </SelectTrigger>
                  <SelectContent>
                    {tiers?.map((tier) => (
                      <SelectItem key={tier.id} value={tier.name}>
                        <div className="flex items-center justify-between gap-4 w-full">
                          <span className="font-medium">
                            {tier.displayName}
                          </span>
                          <span className="text-sm text-muted-foreground">
                            ${tier.priceMonthly}/month
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
                  onValueChange={(v) => handleField("trial_days", parseInt(v))}
                >
                  <SelectTrigger>
                    <SelectValue placeholder="Select duration" />
                  </SelectTrigger>
                  <SelectContent>
                    {TRIAL_DURATIONS.map((d) => (
                      <SelectItem key={d.value} value={d.value.toString()}>
                        {d.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              <div className="space-y-2">
                <Label htmlFor="max_users">Maximum Users</Label>
                <Input
                  id="max_users"
                  type="number"
                  value={formData.max_users ?? ""}
                  onChange={(e) => {
                    const v = parseInt(e.target.value);
                    if (!isNaN(v)) handleField("max_users", v);
                  }}
                  placeholder="50"
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
              {isLoading ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Creating...
                </>
              ) : (
                "Create Organization"
              )}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
