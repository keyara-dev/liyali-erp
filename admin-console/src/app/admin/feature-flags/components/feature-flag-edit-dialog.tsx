"use client";

import { useState, useEffect } from "react";
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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Switch } from "@/components/ui/switch";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  Flag,
  Beaker,
  Shield,
  AlertTriangle,
  Users,
  Plus,
  X,
  Target,
  Percent,
  Calendar,
  Info,
} from "lucide-react";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Calendar as CalendarComponent } from "@/components/ui/calendar";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { format } from "date-fns";
import { cn } from "@/lib/utils";
import type {
  FeatureFlag,
  Variation,
  TargetingRule,
  Condition,
} from "@/app/_actions/feature-flags";

interface FeatureFlagEditDialogProps {
  flag?: FeatureFlag | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSave: (
    flag: Omit<
      FeatureFlag,
      | "id"
      | "createdAt"
      | "updatedAt"
      | "createdBy"
      | "updatedBy"
      | "evaluationCount"
    >,
  ) => void;
  isLoading?: boolean;
}

export function FeatureFlagEditDialog({
  flag,
  open,
  onOpenChange,
  onSave,
  isLoading = false,
}: FeatureFlagEditDialogProps) {
  const [formData, setFormData] = useState({
    key: "",
    name: "",
    description: "",
    type: "boolean" as FeatureFlag["type"],
    defaultValue: "false",
    enabled: false,
    environment: "all" as FeatureFlag["environment"],
    category: "feature" as FeatureFlag["category"],
    tags: [] as string[],
    targeting: {
      enabled: false,
      rules: [] as TargetingRule[],
      rolloutPercentage: 0,
      userSegments: [] as string[],
    },
    variations: [] as Variation[],
    isArchived: false,
    expiresAt: undefined as string | undefined,
  });

  const [validationErrors, setValidationErrors] = useState<
    Record<string, string>
  >({});
  const [tagInput, setTagInput] = useState("");
  const [segmentInput, setSegmentInput] = useState("");
  const [expiryDate, setExpiryDate] = useState<Date>();

  const isEditing = !!flag;

  const categories = [
    { value: "feature", label: "Feature Flag", icon: Flag },
    { value: "experiment", label: "Experiment", icon: Beaker },
    { value: "operational", label: "Operational", icon: Shield },
    { value: "killswitch", label: "Kill Switch", icon: AlertTriangle },
    { value: "permission", label: "Permission", icon: Users },
  ];

  const types = [
    { value: "boolean", label: "Boolean" },
    { value: "string", label: "String" },
    { value: "number", label: "Number" },
    { value: "json", label: "JSON" },
  ];

  const environments = [
    { value: "all", label: "All Environments" },
    { value: "production", label: "Production" },
    { value: "staging", label: "Staging" },
    { value: "development", label: "Development" },
  ];

  useEffect(() => {
    if (flag) {
      setFormData({
        key: flag.key,
        name: flag.name,
        description: flag.description,
        type: flag.type,
        defaultValue: flag.defaultValue,
        enabled: flag.enabled,
        environment: flag.environment,
        category: flag.category,
        tags: flag.tags,
        targeting: flag.targeting,
        variations: flag.variations,
        isArchived: flag.isArchived,
        expiresAt: flag.expiresAt,
      });
      setTagInput(flag.tags.join(", "));
      setSegmentInput(flag.targeting.userSegments.join(", "));
      setExpiryDate(flag.expiresAt ? new Date(flag.expiresAt) : undefined);
    } else {
      // Set default variations based on type
      const defaultVariations: Variation[] = [
        {
          id: "enabled",
          name: "Enabled",
          value: "true",
          description: "Feature enabled",
          weight: 50,
          isControl: false,
        },
        {
          id: "disabled",
          name: "Disabled",
          value: "false",
          description: "Feature disabled (control)",
          weight: 50,
          isControl: true,
        },
      ];

      setFormData({
        key: "",
        name: "",
        description: "",
        type: "boolean",
        defaultValue: "false",
        enabled: false,
        environment: "all",
        category: "feature",
        tags: [],
        targeting: {
          enabled: false,
          rules: [],
          rolloutPercentage: 0,
          userSegments: [],
        },
        variations: defaultVariations,
        isArchived: false,
        expiresAt: undefined,
      });
      setTagInput("");
      setSegmentInput("");
      setExpiryDate(undefined);
    }
    setValidationErrors({});
  }, [flag, open]);

  const handleInputChange = (field: string, value: any) => {
    setFormData((prev) => ({
      ...prev,
      [field]: value,
    }));

    // Clear validation error when field is updated
    if (validationErrors[field]) {
      setValidationErrors((prev) => ({
        ...prev,
        [field]: "",
      }));
    }
  };

  const handleTargetingChange = (field: string, value: any) => {
    setFormData((prev) => ({
      ...prev,
      targeting: {
        ...prev.targeting,
        [field]: value,
      },
    }));
  };

  const handleTagsChange = (value: string) => {
    setTagInput(value);
    const tags = value
      .split(",")
      .map((tag) => tag.trim())
      .filter((tag) => tag.length > 0);
    handleInputChange("tags", tags);
  };

  const handleSegmentsChange = (value: string) => {
    setSegmentInput(value);
    const segments = value
      .split(",")
      .map((segment) => segment.trim())
      .filter((segment) => segment.length > 0);
    handleTargetingChange("userSegments", segments);
  };

  const handleExpiryDateChange = (date?: Date) => {
    setExpiryDate(date);
    handleInputChange("expiresAt", date?.toISOString());
  };

  const addVariation = () => {
    const newVariation: Variation = {
      id: `var-${Date.now()}`,
      name: `Variation ${formData.variations.length + 1}`,
      value: formData.type === "boolean" ? "true" : "variant",
      description: "",
      weight: 0,
      isControl: false,
    };

    handleInputChange("variations", [...formData.variations, newVariation]);
  };

  const updateVariation = (index: number, updates: Partial<Variation>) => {
    const updatedVariations = formData.variations.map((variation, i) =>
      i === index ? { ...variation, ...updates } : variation,
    );
    handleInputChange("variations", updatedVariations);
  };

  const removeVariation = (index: number) => {
    const updatedVariations = formData.variations.filter((_, i) => i !== index);
    handleInputChange("variations", updatedVariations);
  };

  const addTargetingRule = () => {
    const newRule: TargetingRule = {
      id: `rule-${Date.now()}`,
      name: `Rule ${formData.targeting.rules.length + 1}`,
      conditions: [],
      variation: formData.variations[0]?.id || "enabled",
      enabled: true,
      priority: formData.targeting.rules.length + 1,
    };

    handleTargetingChange("rules", [...formData.targeting.rules, newRule]);
  };

  const updateTargetingRule = (
    index: number,
    updates: Partial<TargetingRule>,
  ) => {
    const updatedRules = formData.targeting.rules.map((rule, i) =>
      i === index ? { ...rule, ...updates } : rule,
    );
    handleTargetingChange("rules", updatedRules);
  };

  const removeTargetingRule = (index: number) => {
    const updatedRules = formData.targeting.rules.filter((_, i) => i !== index);
    handleTargetingChange("rules", updatedRules);
  };

  const validateForm = () => {
    const errors: Record<string, string> = {};

    if (!formData.key.trim()) {
      errors.key = "Flag key is required";
    } else if (!/^[a-zA-Z][a-zA-Z0-9_]*$/.test(formData.key)) {
      errors.key =
        "Key must start with a letter and contain only letters, numbers, and underscores";
    }

    if (!formData.name.trim()) {
      errors.name = "Flag name is required";
    }

    if (!formData.description.trim()) {
      errors.description = "Description is required";
    }

    if (formData.variations.length < 2) {
      errors.variations = "At least 2 variations are required";
    }

    const totalWeight = formData.variations.reduce(
      (sum, v) => sum + v.weight,
      0,
    );
    if (totalWeight !== 100 && formData.variations.length > 0) {
      errors.variations = "Variation weights must sum to 100%";
    }

    setValidationErrors(errors);
    return Object.keys(errors).length === 0;
  };

  const handleSave = () => {
    if (!validateForm()) return;

    onSave(formData);
  };

  const selectedCategory = categories.find(
    (cat) => cat.value === formData.category,
  );
  const CategoryIcon = selectedCategory?.icon || Flag;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-4xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <CategoryIcon className="h-5 w-5" />
            {isEditing ? "Edit Feature Flag" : "Create Feature Flag"}
          </DialogTitle>
          <DialogDescription>
            {isEditing
              ? "Modify the feature flag configuration and targeting rules."
              : "Create a new feature flag with targeting and rollout controls."}
          </DialogDescription>
        </DialogHeader>

        <Tabs defaultValue="basic" className="space-y-6">
          <TabsList className="grid w-full grid-cols-4">
            <TabsTrigger value="basic">Basic Info</TabsTrigger>
            <TabsTrigger value="variations">Variations</TabsTrigger>
            <TabsTrigger value="targeting">Targeting</TabsTrigger>
            <TabsTrigger value="settings">Settings</TabsTrigger>
          </TabsList>

          {/* Basic Information */}
          <TabsContent value="basic" className="space-y-4">
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="key">Flag Key *</Label>
                <Input
                  id="key"
                  value={formData.key}
                  onChange={(e) => handleInputChange("key", e.target.value)}
                  placeholder="e.g., new_checkout_flow"
                  disabled={isEditing}
                  className={cn(validationErrors.key && "border-red-500")}
                />
                {validationErrors.key && (
                  <p className="text-sm text-red-600">{validationErrors.key}</p>
                )}
              </div>

              <div className="space-y-2">
                <Label htmlFor="name">Display Name *</Label>
                <Input
                  id="name"
                  value={formData.name}
                  onChange={(e) => handleInputChange("name", e.target.value)}
                  placeholder="e.g., New Checkout Flow"
                  className={cn(validationErrors.name && "border-red-500")}
                />
                {validationErrors.name && (
                  <p className="text-sm text-red-600">
                    {validationErrors.name}
                  </p>
                )}
              </div>
            </div>

            <div className="space-y-2">
              <Label htmlFor="description">Description *</Label>
              <Textarea
                id="description"
                value={formData.description}
                onChange={(e) =>
                  handleInputChange("description", e.target.value)
                }
                placeholder="Describe what this flag controls..."
                className={cn(validationErrors.description && "border-red-500")}
              />
              {validationErrors.description && (
                <p className="text-sm text-red-600">
                  {validationErrors.description}
                </p>
              )}
            </div>

            <div className="grid grid-cols-3 gap-4">
              <div className="space-y-2">
                <Label htmlFor="type">Flag Type</Label>
                <Select
                  value={formData.type}
                  onValueChange={(value) => handleInputChange("type", value)}
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {types.map((type) => (
                      <SelectItem key={type.value} value={type.value}>
                        {type.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              <div className="space-y-2">
                <Label htmlFor="category">Category</Label>
                <Select
                  value={formData.category}
                  onValueChange={(value) =>
                    handleInputChange("category", value)
                  }
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {categories.map((category) => {
                      const Icon = category.icon;
                      return (
                        <SelectItem key={category.value} value={category.value}>
                          <div className="flex items-center gap-2">
                            <Icon className="h-4 w-4" />
                            {category.label}
                          </div>
                        </SelectItem>
                      );
                    })}
                  </SelectContent>
                </Select>
              </div>

              <div className="space-y-2">
                <Label htmlFor="environment">Environment</Label>
                <Select
                  value={formData.environment}
                  onValueChange={(value) =>
                    handleInputChange("environment", value)
                  }
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {environments.map((env) => (
                      <SelectItem key={env.value} value={env.value}>
                        {env.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </div>

            <div className="space-y-2">
              <Label htmlFor="tags">Tags (comma-separated)</Label>
              <Input
                id="tags"
                value={tagInput}
                onChange={(e) => handleTagsChange(e.target.value)}
                placeholder="e.g., ui, checkout, experiment"
              />
              {formData.tags.length > 0 && (
                <div className="flex flex-wrap gap-1 mt-2">
                  {formData.tags.map((tag, index) => (
                    <Badge key={index} variant="secondary">
                      {tag}
                    </Badge>
                  ))}
                </div>
              )}
            </div>

            <div className="space-y-2">
              <Label htmlFor="expiry">Expiry Date (optional)</Label>
              <Popover>
                <PopoverTrigger asChild>
                  <Button
                    variant="outline"
                    className={cn(
                      "w-full justify-start text-left font-normal",
                      !expiryDate && "text-muted-foreground",
                    )}
                  >
                    <Calendar className="mr-2 h-4 w-4" />
                    {expiryDate ? format(expiryDate, "PPP") : "Set expiry date"}
                  </Button>
                </PopoverTrigger>
                <PopoverContent className="w-auto p-0" align="start">
                  <CalendarComponent
                    mode="single"
                    selected={expiryDate}
                    onSelect={handleExpiryDateChange}
                    initialFocus
                  />
                </PopoverContent>
              </Popover>
            </div>
          </TabsContent>

          {/* Variations */}
          <TabsContent value="variations" className="space-y-4">
            <div className="flex items-center justify-between">
              <h3 className="text-lg font-medium">Flag Variations</h3>
              <Button onClick={addVariation} size="sm">
                <Plus className="h-4 w-4 mr-2" />
                Add Variation
              </Button>
            </div>

            <div className="space-y-4">
              {formData.variations.map((variation, index) => (
                <div key={variation.id} className="border rounded-lg p-4">
                  <div className="flex items-center justify-between mb-4">
                    <h4 className="font-medium">Variation {index + 1}</h4>
                    {formData.variations.length > 2 && (
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => removeVariation(index)}
                      >
                        <X className="h-4 w-4" />
                      </Button>
                    )}
                  </div>

                  <div className="grid grid-cols-2 gap-4">
                    <div className="space-y-2">
                      <Label>Name</Label>
                      <Input
                        value={variation.name}
                        onChange={(e) =>
                          updateVariation(index, { name: e.target.value })
                        }
                        placeholder="Variation name"
                      />
                    </div>

                    <div className="space-y-2">
                      <Label>Value</Label>
                      <Input
                        value={variation.value}
                        onChange={(e) =>
                          updateVariation(index, { value: e.target.value })
                        }
                        placeholder="Variation value"
                      />
                    </div>
                  </div>

                  <div className="space-y-2 mt-4">
                    <Label>Description</Label>
                    <Input
                      value={variation.description || ""}
                      onChange={(e) =>
                        updateVariation(index, { description: e.target.value })
                      }
                      placeholder="Optional description"
                    />
                  </div>

                  <div className="flex items-center justify-between mt-4">
                    <div className="space-y-2">
                      <Label>Weight (%)</Label>
                      <Input
                        type="number"
                        min="0"
                        max="100"
                        value={variation.weight}
                        onChange={(e) =>
                          updateVariation(index, {
                            weight: Number(e.target.value),
                          })
                        }
                        className="w-20"
                      />
                    </div>

                    <div className="flex items-center space-x-2">
                      <Switch
                        checked={variation.isControl}
                        onCheckedChange={(checked) =>
                          updateVariation(index, { isControl: checked })
                        }
                      />
                      <Label>Control Group</Label>
                    </div>
                  </div>
                </div>
              ))}
            </div>

            {validationErrors.variations && (
              <Alert>
                <AlertTriangle className="h-4 w-4" />
                <AlertDescription>
                  {validationErrors.variations}
                </AlertDescription>
              </Alert>
            )}
          </TabsContent>

          {/* Targeting */}
          <TabsContent value="targeting" className="space-y-4">
            <div className="flex items-center space-x-2">
              <Switch
                checked={formData.targeting.enabled}
                onCheckedChange={(checked) =>
                  handleTargetingChange("enabled", checked)
                }
              />
              <Label className="flex items-center gap-2">
                <Target className="h-4 w-4" />
                Enable Targeting
              </Label>
            </div>

            {formData.targeting.enabled && (
              <div className="space-y-6">
                {/* Rollout Percentage */}
                <div className="space-y-2">
                  <Label className="flex items-center gap-2">
                    <Percent className="h-4 w-4" />
                    Rollout Percentage
                  </Label>
                  <div className="flex items-center space-x-4">
                    <Input
                      type="number"
                      min="0"
                      max="100"
                      value={formData.targeting.rolloutPercentage}
                      onChange={(e) =>
                        handleTargetingChange(
                          "rolloutPercentage",
                          Number(e.target.value),
                        )
                      }
                      className="w-24"
                    />
                    <span className="text-sm text-muted-foreground">
                      % of users will see the enabled variation
                    </span>
                  </div>
                </div>

                {/* User Segments */}
                <div className="space-y-2">
                  <Label>User Segments (comma-separated)</Label>
                  <Input
                    value={segmentInput}
                    onChange={(e) => handleSegmentsChange(e.target.value)}
                    placeholder="e.g., beta_users, premium_users"
                  />
                  {formData.targeting.userSegments.length > 0 && (
                    <div className="flex flex-wrap gap-1 mt-2">
                      {formData.targeting.userSegments.map((segment, index) => (
                        <Badge key={index} variant="outline">
                          {segment}
                        </Badge>
                      ))}
                    </div>
                  )}
                </div>

                {/* Targeting Rules */}
                <div className="space-y-4">
                  <div className="flex items-center justify-between">
                    <h4 className="font-medium">Targeting Rules</h4>
                    <Button onClick={addTargetingRule} size="sm">
                      <Plus className="h-4 w-4 mr-2" />
                      Add Rule
                    </Button>
                  </div>

                  {formData.targeting.rules.map((rule, index) => (
                    <div key={rule.id} className="border rounded-lg p-4">
                      <div className="flex items-center justify-between mb-4">
                        <Input
                          value={rule.name}
                          onChange={(e) =>
                            updateTargetingRule(index, { name: e.target.value })
                          }
                          placeholder="Rule name"
                          className="max-w-xs"
                        />
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => removeTargetingRule(index)}
                        >
                          <X className="h-4 w-4" />
                        </Button>
                      </div>

                      <div className="grid grid-cols-2 gap-4">
                        <div className="space-y-2">
                          <Label>Target Variation</Label>
                          <Select
                            value={rule.variation}
                            onValueChange={(value) =>
                              updateTargetingRule(index, { variation: value })
                            }
                          >
                            <SelectTrigger>
                              <SelectValue />
                            </SelectTrigger>
                            <SelectContent>
                              {formData.variations.map((variation) => (
                                <SelectItem
                                  key={variation.id}
                                  value={variation.id}
                                >
                                  {variation.name}
                                </SelectItem>
                              ))}
                            </SelectContent>
                          </Select>
                        </div>

                        <div className="flex items-center space-x-2">
                          <Switch
                            checked={rule.enabled}
                            onCheckedChange={(checked) =>
                              updateTargetingRule(index, { enabled: checked })
                            }
                          />
                          <Label>Rule Enabled</Label>
                        </div>
                      </div>

                      <Alert className="mt-4">
                        <Info className="h-4 w-4" />
                        <AlertDescription>
                          Targeting rule conditions will be implemented in a
                          future update.
                        </AlertDescription>
                      </Alert>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </TabsContent>

          {/* Settings */}
          <TabsContent value="settings" className="space-y-4">
            <div className="space-y-4">
              <div className="flex items-center space-x-2">
                <Switch
                  checked={formData.enabled}
                  onCheckedChange={(checked) =>
                    handleInputChange("enabled", checked)
                  }
                />
                <Label>Flag Enabled</Label>
              </div>

              <div className="flex items-center space-x-2">
                <Switch
                  checked={formData.isArchived}
                  onCheckedChange={(checked) =>
                    handleInputChange("isArchived", checked)
                  }
                />
                <Label>Archived</Label>
              </div>

              <div className="space-y-2">
                <Label htmlFor="defaultValue">Default Value</Label>
                <Input
                  id="defaultValue"
                  value={formData.defaultValue}
                  onChange={(e) =>
                    handleInputChange("defaultValue", e.target.value)
                  }
                  placeholder="Default value when flag is disabled"
                />
              </div>
            </div>
          </TabsContent>
        </Tabs>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button onClick={handleSave} disabled={isLoading}>
            {isLoading
              ? "Saving..."
              : isEditing
                ? "Update Flag"
                : "Create Flag"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
