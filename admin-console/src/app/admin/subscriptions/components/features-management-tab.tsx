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
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import {
  Plus,
  Edit,
  Trash2,
  Settings,
  Tag,
  ToggleLeft,
  ToggleRight,
} from "lucide-react";
import { toast } from "sonner";
import {
  getAllSubscriptionFeatures,
  createSubscriptionFeature,
  updateSubscriptionFeature,
  deleteSubscriptionFeature,
  type SubscriptionFeature,
} from "@/app/_actions/subscriptions";

const FEATURE_CATEGORIES = [
  "Core Features",
  "Advanced Features",
  "Integrations",
  "Analytics",
  "Security",
  "Support",
  "Storage",
  "API Access",
];

export function FeaturesManagementTab() {
  const [features, setFeatures] = useState<SubscriptionFeature[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [editingFeature, setEditingFeature] =
    useState<SubscriptionFeature | null>(null);
  const [isCreating, setIsCreating] = useState(false);
  const [selectedCategory, setSelectedCategory] = useState<string>("all");

  // Form state
  const [formData, setFormData] = useState({
    name: "",
    display_name: "",
    description: "",
    category: "Core Features",
    is_active: true,
  });

  useEffect(() => {
    loadFeatures();
  }, []);

  const loadFeatures = async () => {
    setIsLoading(true);
    try {
      const result = await getAllSubscriptionFeatures();
      if (result.success) {
        setFeatures(result.data || []);
      } else {
        toast.error("Failed to load features");
      }
    } catch (error) {
      toast.error("Failed to load features");
    } finally {
      setIsLoading(false);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    try {
      let result;
      if (editingFeature) {
        result = await updateSubscriptionFeature(editingFeature.id, formData);
      } else {
        result = await createSubscriptionFeature(formData);
      }

      if (result.success) {
        toast.success(
          editingFeature
            ? "Feature updated successfully"
            : "Feature created successfully",
        );
        resetForm();
        loadFeatures();
      } else {
        toast.error(result.message || "Operation failed");
      }
    } catch (error) {
      toast.error("Operation failed");
    }
  };

  const handleDelete = async (featureId: string) => {
    if (!confirm("Are you sure you want to delete this feature?")) return;

    try {
      const result = await deleteSubscriptionFeature(featureId);
      if (result.success) {
        toast.success("Feature deleted successfully");
        loadFeatures();
      } else {
        toast.error(result.message || "Failed to delete feature");
      }
    } catch (error) {
      toast.error("Failed to delete feature");
    }
  };

  const toggleFeatureStatus = async (feature: SubscriptionFeature) => {
    try {
      const result = await updateSubscriptionFeature(feature.id, {
        is_active: !feature.is_active,
      });

      if (result.success) {
        toast.success(
          `Feature ${feature.is_active ? "deactivated" : "activated"}`,
        );
        loadFeatures();
      } else {
        toast.error("Failed to update feature status");
      }
    } catch (error) {
      toast.error("Failed to update feature status");
    }
  };

  const resetForm = () => {
    setFormData({
      name: "",
      display_name: "",
      description: "",
      category: "Core Features",
      is_active: true,
    });
    setEditingFeature(null);
    setIsCreating(false);
  };

  const startEdit = (feature: SubscriptionFeature) => {
    setFormData({
      name: feature.name,
      display_name: feature.display_name,
      description: feature.description,
      category: feature.category,
      is_active: feature.is_active,
    });
    setEditingFeature(feature);
    setIsCreating(true);
  };

  const filteredFeatures =
    selectedCategory === "all"
      ? features
      : features.filter((f) => f.category === selectedCategory);

  const featuresByCategory = FEATURE_CATEGORIES.reduce(
    (acc, category) => {
      acc[category] = features.filter((f) => f.category === category);
      return acc;
    },
    {} as Record<string, SubscriptionFeature[]>,
  );

  if (isLoading) {
    return <div>Loading features...</div>;
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-lg font-medium">Features Management</h3>
          <p className="text-sm text-muted-foreground">
            Manage available features for subscription tiers
          </p>
        </div>
        <Button onClick={() => setIsCreating(true)}>
          <Plus className="mr-2 h-4 w-4" />
          Add Feature
        </Button>
      </div>

      {/* Category Filter */}
      <div className="flex flex-wrap gap-2">
        <Button
          variant={selectedCategory === "all" ? "default" : "outline"}
          size="sm"
          onClick={() => setSelectedCategory("all")}
        >
          All Features ({features.length})
        </Button>
        {FEATURE_CATEGORIES.map((category) => (
          <Button
            key={category}
            variant={selectedCategory === category ? "default" : "outline"}
            size="sm"
            onClick={() => setSelectedCategory(category)}
          >
            {category} ({featuresByCategory[category]?.length || 0})
          </Button>
        ))}
      </div>

      {/* Create/Edit Form */}
      {isCreating && (
        <Card>
          <CardHeader>
            <CardTitle>
              {editingFeature ? "Edit Feature" : "Create New Feature"}
            </CardTitle>
            <CardDescription>
              {editingFeature
                ? "Update feature details"
                : "Define a new feature for subscription tiers"}
            </CardDescription>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleSubmit} className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="text-sm font-medium">Name (Internal)</label>
                  <Input
                    value={formData.name}
                    onChange={(e) =>
                      setFormData((prev) => ({ ...prev, name: e.target.value }))
                    }
                    placeholder="advanced_analytics"
                    required
                  />
                </div>
                <div>
                  <label className="text-sm font-medium">Display Name</label>
                  <Input
                    value={formData.display_name}
                    onChange={(e) =>
                      setFormData((prev) => ({
                        ...prev,
                        display_name: e.target.value,
                      }))
                    }
                    placeholder="Advanced Analytics"
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
                  placeholder="Detailed analytics and reporting capabilities..."
                  rows={2}
                />
              </div>

              <div>
                <label className="text-sm font-medium">Category</label>
                <select
                  value={formData.category}
                  onChange={(e) =>
                    setFormData((prev) => ({
                      ...prev,
                      category: e.target.value,
                    }))
                  }
                  className="w-full p-2 border rounded-md"
                  required
                >
                  {FEATURE_CATEGORIES.map((category) => (
                    <option key={category} value={category}>
                      {category}
                    </option>
                  ))}
                </select>
              </div>

              <div className="flex items-center space-x-2">
                <input
                  type="checkbox"
                  id="is_active"
                  checked={formData.is_active}
                  onChange={(e) =>
                    setFormData((prev) => ({
                      ...prev,
                      is_active: e.target.checked,
                    }))
                  }
                />
                <label htmlFor="is_active" className="text-sm font-medium">
                  Active
                </label>
              </div>

              <div className="flex gap-2">
                <Button type="submit">
                  {editingFeature ? "Update Feature" : "Create Feature"}
                </Button>
                <Button type="button" variant="outline" onClick={resetForm}>
                  Cancel
                </Button>
              </div>
            </form>
          </CardContent>
        </Card>
      )}

      {/* Features List */}
      <div className="space-y-4">
        {selectedCategory === "all" ? (
          // Group by category when showing all
          FEATURE_CATEGORIES.map((category) => {
            const categoryFeatures = featuresByCategory[category];
            if (!categoryFeatures?.length) return null;

            return (
              <Card key={category}>
                <CardHeader>
                  <CardTitle className="flex items-center gap-2">
                    <Tag className="h-5 w-5" />
                    {category}
                    <Badge variant="secondary">{categoryFeatures.length}</Badge>
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="grid gap-3 md:grid-cols-2">
                    {categoryFeatures.map((feature) => (
                      <FeatureCard
                        key={feature.id}
                        feature={feature}
                        onEdit={startEdit}
                        onDelete={handleDelete}
                        onToggleStatus={toggleFeatureStatus}
                      />
                    ))}
                  </div>
                </CardContent>
              </Card>
            );
          })
        ) : (
          // Show filtered features
          <div className="grid gap-3 md:grid-cols-2 lg:grid-cols-3">
            {filteredFeatures.map((feature) => (
              <FeatureCard
                key={feature.id}
                feature={feature}
                onEdit={startEdit}
                onDelete={handleDelete}
                onToggleStatus={toggleFeatureStatus}
              />
            ))}
          </div>
        )}
      </div>
    </div>
  );
}

function FeatureCard({
  feature,
  onEdit,
  onDelete,
  onToggleStatus,
}: {
  feature: SubscriptionFeature;
  onEdit: (feature: SubscriptionFeature) => void;
  onDelete: (id: string) => void;
  onToggleStatus: (feature: SubscriptionFeature) => void;
}) {
  return (
    <div className="border rounded-lg p-4 space-y-3">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <h4 className="font-medium">{feature.display_name}</h4>
          <Badge variant={feature.is_active ? "success" : "secondary"}>
            {feature.is_active ? "Active" : "Inactive"}
          </Badge>
        </div>
        <div className="flex gap-1">
          <Button
            size="icon"
            variant="ghost"
            onClick={() => onToggleStatus(feature)}
            title={feature.is_active ? "Deactivate" : "Activate"}
          >
            {feature.is_active ? (
              <ToggleRight className="h-4 w-4 text-green-600" />
            ) : (
              <ToggleLeft className="h-4 w-4 text-gray-400" />
            )}
          </Button>
          <Button size="icon" variant="ghost" onClick={() => onEdit(feature)}>
            <Edit className="h-4 w-4" />
          </Button>
          <Button
            size="icon"
            variant="ghost"
            onClick={() => onDelete(feature.id)}
          >
            <Trash2 className="h-4 w-4" />
          </Button>
        </div>
      </div>

      <p className="text-sm text-muted-foreground">{feature.description}</p>

      <div className="flex items-center gap-2 text-xs">
        <Badge variant="outline">{feature.category}</Badge>
        <span className="text-muted-foreground">ID: {feature.name}</span>
      </div>
    </div>
  );
}
