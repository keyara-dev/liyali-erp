"use client";

import { useState, useEffect } from "react";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import {
  Plus,
  Edit,
  Trash2,
  DollarSign,
  Users,
  HardDrive,
  Check,
  FileText,
  Workflow,
  Shield,
} from "lucide-react";
import { toast } from "sonner";
import {
  getAllSubscriptionTiers,
  createSubscriptionTier,
  updateSubscriptionTier,
  deleteSubscriptionTier,
  getAllSubscriptionFeatures,
  type SubscriptionTier,
  type SubscriptionFeature,
  type CreateTierRequest,
} from "@/app/_actions/subscriptions";

export function SubscriptionTiersTab() {
  const [tiers, setTiers] = useState<SubscriptionTier[]>([]);
  const [features, setFeatures] = useState<SubscriptionFeature[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [editingTier, setEditingTier] = useState<SubscriptionTier | null>(null);
  const [isCreating, setIsCreating] = useState(false);

  // Form state
  const [formData, setFormData] = useState<CreateTierRequest>({
    name: "",
    displayName: "",
    description: "",
    priceMonthly: 0,
    priceYearly: 0,
    maxWorkspaces: 1,
    maxTeamMembers: 10,
    maxDocuments: 100,
    maxWorkflows: 5,
    maxCustomRoles: 0,
    features: [],
    isActive: true,
    sortOrder: 0,
  });

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    setIsLoading(true);
    try {
      const [tiersResult, featuresResult] = await Promise.all([
        getAllSubscriptionTiers(),
        getAllSubscriptionFeatures(),
      ]);

      if (tiersResult.success) {
        console.log("Raw tiers data:", tiersResult.data);
        // Ensure features is always an array
        const normalizedTiers = (tiersResult.data || []).map((tier) => ({
          ...tier,
          features: Array.isArray(tier.features)
            ? tier.features
            : typeof tier.features === "string"
              ? JSON.parse(tier.features)
              : [],
        }));
        console.log("Normalized tiers:", normalizedTiers);
        setTiers(normalizedTiers);
      } else {
        console.error("Failed to load tiers:", tiersResult.message);
        toast.error(tiersResult.message || "Failed to load tiers");
      }
      if (featuresResult.success) {
        console.log("Features data:", featuresResult.data);
        setFeatures(featuresResult.data || []);
      } else {
        console.error("Failed to load features:", featuresResult.message);
      }
    } catch (error) {
      console.error("Error loading subscription data:", error);
      toast.error("Failed to load subscription data");
    } finally {
      setIsLoading(false);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    try {
      let result;
      if (editingTier) {
        result = await updateSubscriptionTier({
          ...formData,
          id: editingTier.id,
        });
      } else {
        result = await createSubscriptionTier(formData);
      }

      if (result.success) {
        toast.success(
          editingTier
            ? "Tier updated successfully"
            : "Tier created successfully",
        );
        resetForm();
        loadData();
      } else {
        toast.error(result.message || "Operation failed");
      }
    } catch (error) {
      toast.error("Operation failed");
    }
  };

  const handleDelete = async (tierId: string) => {
    if (!confirm("Are you sure you want to delete this tier?")) return;

    try {
      const result = await deleteSubscriptionTier(tierId);
      if (result.success) {
        toast.success("Tier deleted successfully");
        loadData();
      } else {
        toast.error(result.message || "Failed to delete tier");
      }
    } catch (error) {
      toast.error("Failed to delete tier");
    }
  };

  const resetForm = () => {
    setFormData({
      name: "",
      displayName: "",
      description: "",
      priceMonthly: 0,
      priceYearly: 0,
      maxWorkspaces: 1,
      maxTeamMembers: 10,
      maxDocuments: 100,
      maxWorkflows: 5,
      maxCustomRoles: 0,
      features: [],
      isActive: true,
      sortOrder: 0,
    });
    setEditingTier(null);
    setIsCreating(false);
  };

  const startEdit = (tier: SubscriptionTier) => {
    setFormData({
      name: tier.name,
      displayName: tier.displayName,
      description: tier.description,
      priceMonthly: tier.priceMonthly,
      priceYearly: tier.priceYearly,
      maxWorkspaces: tier.maxWorkspaces,
      maxTeamMembers: tier.maxTeamMembers,
      maxDocuments: tier.maxDocuments,
      maxWorkflows: tier.maxWorkflows,
      maxCustomRoles: tier.maxCustomRoles,
      features: tier.features,
      isActive: tier.isActive,
      sortOrder: tier.sortOrder,
    });
    setEditingTier(tier);
    setIsCreating(true);
  };

  const toggleFeature = (featureId: string) => {
    setFormData((prev) => ({
      ...prev,
      features: prev.features.includes(featureId)
        ? prev.features.filter((id) => id !== featureId)
        : [...prev.features, featureId],
    }));
  };

  if (isLoading) {
    return (
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <div>
            <Skeleton className="h-6 w-48 mb-2" />
            <Skeleton className="h-4 w-64" />
          </div>
          <Skeleton className="h-10 w-32" />
        </div>
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {[...Array(3)].map((_, i) => (
            <Card key={i}>
              <CardHeader>
                <Skeleton className="h-6 w-32 mb-2" />
                <Skeleton className="h-4 w-full" />
              </CardHeader>
              <CardContent className="space-y-4">
                <Skeleton className="h-8 w-24" />
                <div className="space-y-2">
                  <Skeleton className="h-4 w-full" />
                  <Skeleton className="h-4 w-full" />
                  <Skeleton className="h-4 w-3/4" />
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-lg font-medium">Subscription Tiers</h3>
          <p className="text-sm text-muted-foreground">
            Manage pricing tiers and their features
          </p>
        </div>
        <Button onClick={() => setIsCreating(true)}>
          <Plus className="mr-2 h-4 w-4" />
          Add Tier
        </Button>
      </div>

      {/* Create/Edit Form */}
      {isCreating && (
        <Card>
          <CardHeader>
            <CardTitle>
              {editingTier ? "Edit Tier" : "Create New Tier"}
            </CardTitle>
            <CardDescription>
              {editingTier
                ? "Update tier details and features"
                : "Define a new subscription tier"}
            </CardDescription>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleSubmit} className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="text-sm font-medium">Name</label>
                  <Input
                    value={formData.name}
                    onChange={(e) =>
                      setFormData((prev) => ({ ...prev, name: e.target.value }))
                    }
                    placeholder="starter, pro, custom"
                    required
                    disabled={!!editingTier}
                  />
                  <p className="text-xs text-muted-foreground mt-1">
                    Lowercase, no spaces. Cannot be changed after creation.
                  </p>
                </div>
                <div>
                  <label className="text-sm font-medium">Display Name</label>
                  <Input
                    value={formData.displayName}
                    onChange={(e) =>
                      setFormData((prev) => ({
                        ...prev,
                        displayName: e.target.value,
                      }))
                    }
                    placeholder="Professional Plan"
                    required
                  />
                </div>
              </div>

              <div>
                <label className="text-sm font-medium">Description</label>
                <Textarea
                  value={formData.description}
                  onChange={(e) =>
                    setFormData((prev) => ({
                      ...prev,
                      description: e.target.value,
                    }))
                  }
                  placeholder="Perfect for growing teams..."
                  rows={2}
                />
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="text-sm font-medium">
                    Monthly Price ($)
                  </label>
                  <Input
                    type="number"
                    value={formData.priceMonthly}
                    onChange={(e) =>
                      setFormData((prev) => ({
                        ...prev,
                        priceMonthly: parseFloat(e.target.value) || 0,
                      }))
                    }
                    min="0"
                    step="0.01"
                  />
                </div>
                <div>
                  <label className="text-sm font-medium">
                    Yearly Price ($)
                  </label>
                  <Input
                    type="number"
                    value={formData.priceYearly}
                    onChange={(e) =>
                      setFormData((prev) => ({
                        ...prev,
                        priceYearly: parseFloat(e.target.value) || 0,
                      }))
                    }
                    min="0"
                    step="0.01"
                  />
                </div>
              </div>

              <div className="grid grid-cols-3 gap-4">
                <div>
                  <label className="text-sm font-medium">
                    Max Team Members
                  </label>
                  <Input
                    type="number"
                    value={formData.maxTeamMembers}
                    onChange={(e) =>
                      setFormData((prev) => ({
                        ...prev,
                        maxTeamMembers: parseInt(e.target.value) || 0,
                      }))
                    }
                    min="-1"
                    placeholder="-1 for unlimited"
                  />
                  <p className="text-xs text-muted-foreground mt-1">
                    Use -1 for unlimited
                  </p>
                </div>
                <div>
                  <label className="text-sm font-medium">Max Workspaces</label>
                  <Input
                    type="number"
                    value={formData.maxWorkspaces}
                    onChange={(e) =>
                      setFormData((prev) => ({
                        ...prev,
                        maxWorkspaces: parseInt(e.target.value) || 0,
                      }))
                    }
                    min="-1"
                    placeholder="-1 for unlimited"
                  />
                </div>
                <div>
                  <label className="text-sm font-medium">Max Documents</label>
                  <Input
                    type="number"
                    value={formData.maxDocuments}
                    onChange={(e) =>
                      setFormData((prev) => ({
                        ...prev,
                        maxDocuments: parseInt(e.target.value) || 0,
                      }))
                    }
                    min="-1"
                    placeholder="-1 for unlimited"
                  />
                </div>
              </div>

              <div className="grid grid-cols-3 gap-4">
                <div>
                  <label className="text-sm font-medium">Max Workflows</label>
                  <Input
                    type="number"
                    value={formData.maxWorkflows}
                    onChange={(e) =>
                      setFormData((prev) => ({
                        ...prev,
                        maxWorkflows: parseInt(e.target.value) || 0,
                      }))
                    }
                    min="-1"
                    placeholder="-1 for unlimited"
                  />
                </div>
                <div>
                  <label className="text-sm font-medium">
                    Max Custom Roles
                  </label>
                  <Input
                    type="number"
                    value={formData.maxCustomRoles}
                    onChange={(e) =>
                      setFormData((prev) => ({
                        ...prev,
                        maxCustomRoles: parseInt(e.target.value) || 0,
                      }))
                    }
                    min="-1"
                    placeholder="-1 for unlimited"
                  />
                </div>
                <div>
                  <label className="text-sm font-medium">Sort Order</label>
                  <Input
                    type="number"
                    value={formData.sortOrder}
                    onChange={(e) =>
                      setFormData((prev) => ({
                        ...prev,
                        sortOrder: parseInt(e.target.value) || 0,
                      }))
                    }
                    min="0"
                  />
                </div>
              </div>

              {/* Features Selection */}
              <div>
                <label className="text-sm font-medium mb-2 block">
                  Features
                </label>
                <div className="grid grid-cols-2 gap-2 max-h-40 overflow-y-auto border rounded p-2">
                  {features.map((feature) => (
                    <div
                      key={feature.id}
                      className="flex items-center space-x-2"
                    >
                      <input
                        type="checkbox"
                        id={feature.id}
                        checked={formData.features.includes(feature.id)}
                        onChange={() => toggleFeature(feature.id)}
                        className="rounded"
                      />
                      <label htmlFor={feature.id} className="text-sm">
                        {feature.displayName}
                      </label>
                    </div>
                  ))}
                </div>
              </div>

              <div className="flex items-center space-x-2">
                <input
                  type="checkbox"
                  id="is_active"
                  checked={formData.isActive}
                  onChange={(e) =>
                    setFormData((prev) => ({
                      ...prev,
                      isActive: e.target.checked,
                    }))
                  }
                />
                <label htmlFor="is_active" className="text-sm font-medium">
                  Active
                </label>
              </div>

              <div className="flex gap-2">
                <Button type="submit">
                  {editingTier ? "Update Tier" : "Create Tier"}
                </Button>
                <Button type="button" variant="outline" onClick={resetForm}>
                  Cancel
                </Button>
              </div>
            </form>
          </CardContent>
        </Card>
      )}

      {/* Tiers List */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {tiers.map((tier) => (
          <Card key={tier.id} className="relative">
            <CardHeader>
              <div className="flex items-center justify-between">
                <CardTitle className="flex items-center gap-2">
                  {tier.displayName || tier.name || "Unnamed Tier"}
                  {!tier.isActive && (
                    <Badge variant="secondary">Inactive</Badge>
                  )}
                </CardTitle>
                <div className="flex gap-1">
                  <Button
                    size="icon"
                    variant="ghost"
                    onClick={() => startEdit(tier)}
                  >
                    <Edit className="h-4 w-4" />
                  </Button>
                  <Button
                    size="icon"
                    variant="ghost"
                    onClick={() => handleDelete(tier.id)}
                  >
                    <Trash2 className="h-4 w-4" />
                  </Button>
                </div>
              </div>
              <CardDescription>
                {tier.description || "No description"}
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-1">
                  <DollarSign className="h-4 w-4" />
                  <span className="font-semibold">${tier.priceMonthly}/mo</span>
                </div>
                <div className="text-sm text-muted-foreground">
                  ${tier.priceYearly}/yr
                </div>
              </div>

              <div className="space-y-2">
                <div className="flex items-center gap-2 text-sm">
                  <Users className="h-4 w-4" />
                  {tier.maxTeamMembers === -1
                    ? "Unlimited"
                    : `Up to ${tier.maxTeamMembers}`}{" "}
                  team members
                </div>
                <div className="flex items-center gap-2 text-sm">
                  <HardDrive className="h-4 w-4" />
                  {tier.maxWorkspaces === -1
                    ? "Unlimited"
                    : tier.maxWorkspaces}{" "}
                  workspace{tier.maxWorkspaces !== 1 ? "s" : ""}
                </div>
                <div className="flex items-center gap-2 text-sm text-muted-foreground">
                  <FileText className="h-4 w-4" />
                  {tier.maxDocuments === -1
                    ? "Unlimited"
                    : tier.maxDocuments}{" "}
                  documents
                </div>
                <div className="flex items-center gap-2 text-sm text-muted-foreground">
                  <Workflow className="h-4 w-4" />
                  {tier.maxWorkflows === -1
                    ? "Unlimited"
                    : tier.maxWorkflows}{" "}
                  workflows
                </div>
                {tier.maxCustomRoles > 0 && (
                  <div className="flex items-center gap-2 text-sm text-muted-foreground">
                    <Shield className="h-4 w-4" />
                    {tier.maxCustomRoles === -1
                      ? "Unlimited"
                      : tier.maxCustomRoles}{" "}
                    custom roles
                  </div>
                )}
              </div>

              <div>
                <p className="text-sm font-medium mb-2">Features:</p>
                <div className="space-y-1">
                  {Array.isArray(tier.features) &&
                    tier.features.slice(0, 3).map((featureId) => {
                      const feature = features.find((f) => f.id === featureId);
                      return feature ? (
                        <div
                          key={featureId}
                          className="flex items-center gap-2 text-sm"
                        >
                          <Check className="h-3 w-3 text-green-600" />
                          {feature.displayName}
                        </div>
                      ) : null;
                    })}
                  {Array.isArray(tier.features) && tier.features.length > 3 && (
                    <div className="text-sm text-muted-foreground">
                      +{tier.features.length - 3} more features
                    </div>
                  )}
                </div>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  );
}
