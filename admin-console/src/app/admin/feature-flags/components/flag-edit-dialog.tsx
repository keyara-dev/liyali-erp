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
import { SelectField } from "@/components/ui/select-field";
import { Switch } from "@/components/ui/switch";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  Flag,
  TestTube,
  Zap,
  Shield,
  Bell,
  Plus,
  X,
  Target,
  Users,
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

interface FlagEditDialogProps {
  flag?: FeatureFlag | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSave: (
    flag: Omit<
      FeatureFlag,
      | "id"
      | "created_at"
      | "updated_at"
      | "created_by"
      | "updated_by"
      | "evaluation_count"
      | "last_evaluated"
    >,
  ) => void;
  isLoading?: boolean;
}

export function FlagEditDialog({
  flag,
  open,
  onOpenChange,
  onSave,
  isLoading = false,
}: FlagEditDialogProps) {
  const [formData, setFormData] = useState({
    key: "",
    name: "",
    description: "",
    type: "boolean" as FeatureFlag["type"],
    default_value: "false",
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
    is_archived: false,
    expires_at: undefined as string | undefined,
  });

  const [tagInput, setTagInput] = useState("");
  const [segmentInput, setSegmentInput] = useState("");
  const [expiryDate, setExpiryDate] = useState<Date>();
  const [validationErrors, setValidationErrors] = useState<
    Record<string, string>
  >({});

  const isEditing = !!flag;

  const categories = [
    { value: "feature", label: "Feature Flag" },
    { value: "experiment", label: "Experiment" },
    { value: "operational", label: "Operational" },
    { value: "killswitch", label: "Kill Switch" },
    { value: "permission", label: "Permission" },
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

  const operators = [
    { value: "equals", label: "Equals" },
    { value: "not_equals", label: "Not Equals" },
    { value: "contains", label: "Contains" },
    { value: "not_contains", label: "Not Contains" },
    { value: "greater_than", label: "Greater Than" },
    { value: "less_than", label: "Less Than" },
    { value: "in", label: "In List" },
    { value: "not_in", label: "Not In List" },
  ];

  useEffect(() => {
    if (flag) {
      setFormData({
        key: flag.key,
        name: flag.name,
        description: flag.description,
        type: flag.type,
        default_value: flag.default_value,
        enabled: flag.enabled,
        environment: flag.environment,
        category: flag.category,
        tags: flag.tags,
        targeting: flag.targeting,
        variations: flag.variations,
        is_archived: flag.is_archived,
        expires_at: flag.expires_at,
      });
      setTagInput(flag.tags.join(", "));
      setSegmentInput(flag.targeting.userSegments.join(", "));
      if (flag.expires_at) {
        setExpiryDate(new Date(flag.expires_at));
      }
    } else {
      const defaultVariations = getDefaultVariations("boolean");
      setFormData({
        key: "",
        name: "",
        description: "",
        type: "boolean",
        default_value: "false",
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
        is_archived: false,
        expires_at: undefined,
      });
      setTagInput("");
      setSegmentInput("");
      setExpiryDate(undefined);
    }
    setValidationErrors({});
  }, [flag, open]);

  const getDefaultVariations = (type: FeatureFlag["type"]): Variation[] => {
    switch (type) {
      case "boolean":
        return [
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
      case "string":
        return [
          {
            id: "control",
            name: "Control",
            value: "control",
            description: "Control variation",
            weight: 50,
            isControl: true,
          },
          {
            id: "variant",
            name: "Variant",
            value: "variant",
            description: "Test variation",
            weight: 50,
            isControl: false,
          },
        ];
      default:
        return [];
    }
  };

  const handleInputChange = (field: string, value: any) => {
    setFormData((prev) => ({
      ...prev,
      [field]: value,
    }));

    if (validationErrors[field]) {
      setValidationErrors((prev) => ({
        ...prev,
        [field]: "",
      }));
    }
  };

  const handleTypeChange = (type: FeatureFlag["type"]) => {
    const defaultVariations = getDefaultVariations(type);
    const default_value =
      type === "boolean" ? "false" : type === "number" ? "0" : "";

    setFormData((prev) => ({
      ...prev,
      type,
      default_value,
      variations: defaultVariations,
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
    handleInputChange("targeting", {
      ...formData.targeting,
      userSegments: segments,
    });
  };

  const handleExpiryDateChange = (date?: Date) => {
    setExpiryDate(date);
    handleInputChange("expires_at", date?.toISOString());
  };

  const addVariation = () => {
    const newVariation: Variation = {
      id: `var-${Date.now()}`,
      name: `Variation ${formData.variations.length + 1}`,
      value: "",
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
      variation: formData.variations[0]?.id || "",
      enabled: true,
      priority: formData.targeting.rules.length + 1,
    };

    handleInputChange("targeting", {
      ...formData.targeting,
      rules: [...formData.targeting.rules, newRule],
    });
  };

  const updateTargetingRule = (
    index: number,
    updates: Partial<TargetingRule>,
  ) => {
    const updatedRules = formData.targeting.rules.map((rule, i) =>
      i === index ? { ...rule, ...updates } : rule,
    );
    handleInputChange("targeting", {
      ...formData.targeting,
      rules: updatedRules,
    });
  };

  const removeTargetingRule = (index: number) => {
    const updatedRules = formData.targeting.rules.filter((_, i) => i !== index);
    handleInputChange("targeting", {
      ...formData.targeting,
      rules: updatedRules,
    });
  };

  const addCondition = (ruleIndex: number) => {
    const newCondition: Condition = {
      attribute: "",
      operator: "equals",
      value: "",
    };

    const updatedRules = formData.targeting.rules.map((rule, i) =>
      i === ruleIndex
        ? { ...rule, conditions: [...rule.conditions, newCondition] }
        : rule,
    );

    handleInputChange("targeting", {
      ...formData.targeting,
      rules: updatedRules,
    });
  };

  const updateCondition = (
    ruleIndex: number,
    conditionIndex: number,
    updates: Partial<Condition>,
  ) => {
    const updatedRules = formData.targeting.rules.map((rule, i) =>
      i === ruleIndex
        ? {
            ...rule,
            conditions: rule.conditions.map((condition, j) =>
              j === conditionIndex ? { ...condition, ...updates } : condition,
            ),
          }
        : rule,
    );

    handleInputChange("targeting", {
      ...formData.targeting,
      rules: updatedRules,
    });
  };

  const removeCondition = (ruleIndex: number, conditionIndex: number) => {
    const updatedRules = formData.targeting.rules.map((rule, i) =>
      i === ruleIndex
        ? {
            ...rule,
            conditions: rule.conditions.filter((_, j) => j !== conditionIndex),
          }
        : rule,
    );

    handleInputChange("targeting", {
      ...formData.targeting,
      rules: updatedRules,
    });
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

    if (formData.variations.length === 0) {
      errors.variations = "At least one variation is required";
    }

    if (formData.targeting.enabled && formData.variations.length > 0) {
      const totalWeight = formData.variations.reduce(
        (sum, variation) => sum + variation.weight,
        0,
      );
      if (totalWeight !== 100) {
        errors.variations = "Variation weights must sum to 100%";
      }
    }

    setValidationErrors(errors);
    return Object.keys(errors).length === 0;
  };

  const handleSave = () => {
    if (!validateForm()) return;

    onSave(formData);
  };

  const categoryIconMap: Record<string, React.ElementType> = {
    feature: Flag,
    experiment: TestTube,
    operational: Zap,
    killswitch: Shield,
    permission: Bell,
  };
  const CategoryIcon = categoryIconMap[formData.category] || Flag;

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

          {/* Basic Information Tab */}
          <TabsContent value="basic" className="space-y-4">
            <div className="grid grid-cols-2 gap-4">
              <Input
                name="key"
                label="Flag Key"
                required
                value={formData.key}
                onChange={(e) => handleInputChange("key", e.target.value)}
                placeholder="e.g., new_checkout_flow"
                disabled={isEditing}
                isInvalid={!!validationErrors.key}
                errorText={validationErrors.key}
              />
              <Input
                name="name"
                label="Display Name"
                required
                value={formData.name}
                onChange={(e) => handleInputChange("name", e.target.value)}
                placeholder="e.g., New Checkout Flow"
                isInvalid={!!validationErrors.name}
                errorText={validationErrors.name}
              />
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
              <SelectField
                label="Category"
                value={formData.category}
                onValueChange={(value) =>
                  handleInputChange("category", value)
                }
                options={categories}
                classNames={{ wrapper: "max-w-full" }}
              />
              <SelectField
                label="Type"
                value={formData.type}
                onValueChange={(v) => handleTypeChange(v as FeatureFlag["type"])}
                options={types}
                classNames={{ wrapper: "max-w-full" }}
              />
              <SelectField
                label="Environment"
                value={formData.environment}
                onValueChange={(value) =>
                  handleInputChange("environment", value)
                }
                options={environments}
                classNames={{ wrapper: "max-w-full" }}
              />
            </div>

            <div className="space-y-2">
              <Input
                name="tags"
                label="Tags (comma-separated)"
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
          </TabsContent>

          {/* Variations Tab */}
          <TabsContent value="variations" className="space-y-4">
            <div className="flex items-center justify-between">
              <h3 className="text-lg font-medium">Flag Variations</h3>
              <Button onClick={addVariation} size="sm">
                <Plus className="h-4 w-4 mr-2" />
                Add Variation
              </Button>
            </div>

            {validationErrors.variations && (
              <Alert>
                <AlertDescription>
                  {validationErrors.variations}
                </AlertDescription>
              </Alert>
            )}

            <div className="space-y-4">
              {formData.variations.map((variation, index) => (
                <div
                  key={variation.id}
                  className="border rounded-lg p-4 space-y-4"
                >
                  <div className="flex items-center justify-between">
                    <h4 className="font-medium">Variation {index + 1}</h4>
                    <div className="flex items-center space-x-2">
                      {variation.isControl && (
                        <Badge variant="outline">Control</Badge>
                      )}
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => removeVariation(index)}
                        disabled={formData.variations.length <= 1}
                      >
                        <X className="h-4 w-4" />
                      </Button>
                    </div>
                  </div>

                  <div className="grid grid-cols-2 gap-4">
                    <Input
                      label="Name"
                      value={variation.name}
                      onChange={(e) =>
                        updateVariation(index, { name: e.target.value })
                      }
                      placeholder="Variation name"
                    />
                    <Input
                      label="Value"
                      value={variation.value}
                      onChange={(e) =>
                        updateVariation(index, { value: e.target.value })
                      }
                      placeholder="Variation value"
                    />
                  </div>

                  <Input
                    label="Description"
                    value={variation.description || ""}
                    onChange={(e) =>
                      updateVariation(index, { description: e.target.value })
                    }
                    placeholder="Describe this variation"
                  />

                  <div className="grid grid-cols-2 gap-4">
                    <Input
                      label="Weight (%)"
                      type="number"
                      min="0"
                      max="100"
                      value={variation.weight}
                      onChange={(e) =>
                        updateVariation(index, {
                          weight: Number(e.target.value),
                        })
                      }
                    />
                    <div className="flex items-center space-x-2 pt-6">
                      <Switch
                        checked={variation.isControl}
                        onCheckedChange={(checked) =>
                          updateVariation(index, { isControl: checked })
                        }
                      />
                      <Label>Control variation</Label>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </TabsContent>

          {/* Targeting Tab */}
          <TabsContent value="targeting" className="space-y-4">
            <div className="flex items-center space-x-2">
              <Switch
                checked={formData.targeting.enabled}
                onCheckedChange={(checked) =>
                  handleInputChange("targeting", {
                    ...formData.targeting,
                    enabled: checked,
                  })
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
                        handleInputChange("targeting", {
                          ...formData.targeting,
                          rolloutPercentage: Number(e.target.value),
                        })
                      }
                      classNames={{ wrapper: "w-24" }}
                    />
                    <span className="text-sm text-muted-foreground">
                      % of users will see this flag
                    </span>
                  </div>
                </div>

                {/* User Segments */}
                <div className="space-y-2">
                  <Label className="flex items-center gap-2">
                    <Users className="h-4 w-4" />
                    User Segments
                  </Label>
                  <Input
                    value={segmentInput}
                    onChange={(e) => handleSegmentsChange(e.target.value)}
                    placeholder="e.g., beta_users, premium_users"
                  />
                  {formData.targeting.userSegments.length > 0 && (
                    <div className="flex flex-wrap gap-1 mt-2">
                      {formData.targeting.userSegments.map((segment, index) => (
                        <Badge key={index} variant="secondary">
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

                  {formData.targeting.rules.map((rule, ruleIndex) => (
                    <div
                      key={rule.id}
                      className="border rounded-lg p-4 space-y-4"
                    >
                      <div className="flex items-center justify-between">
                        <div className="flex items-center space-x-2">
                          <Input
                            value={rule.name}
                            onChange={(e) =>
                              updateTargetingRule(ruleIndex, {
                                name: e.target.value,
                              })
                            }
                            placeholder="Rule name"
                            classNames={{ wrapper: "w-48" }}
                          />
                          <Switch
                            checked={rule.enabled}
                            onCheckedChange={(checked) =>
                              updateTargetingRule(ruleIndex, {
                                enabled: checked,
                              })
                            }
                          />
                        </div>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => removeTargetingRule(ruleIndex)}
                        >
                          <X className="h-4 w-4" />
                        </Button>
                      </div>

                      <SelectField
                        label="Target Variation"
                        value={rule.variation}
                        onValueChange={(value) =>
                          updateTargetingRule(ruleIndex, { variation: value })
                        }
                        options={formData.variations}
                        classNames={{ wrapper: "max-w-full" }}
                      />

                      <div className="space-y-2">
                        <div className="flex items-center justify-between">
                          <Label>Conditions</Label>
                          <Button
                            onClick={() => addCondition(ruleIndex)}
                            size="sm"
                            variant="outline"
                          >
                            <Plus className="h-4 w-4 mr-2" />
                            Add Condition
                          </Button>
                        </div>

                        {rule.conditions.map((condition, conditionIndex) => (
                          <div
                            key={conditionIndex}
                            className="grid grid-cols-4 gap-2 items-end"
                          >
                            <Input
                              label="Attribute"
                              classNames={{ label: "text-xs" }}
                              value={condition.attribute}
                              onChange={(e) =>
                                updateCondition(ruleIndex, conditionIndex, {
                                  attribute: e.target.value,
                                })
                              }
                              placeholder="e.g., userType"
                            />
                            <SelectField
                              label="Operator"
                              classNames={{
                                label: "text-xs",
                                wrapper: "max-w-full",
                              }}
                              value={condition.operator}
                              onValueChange={(value) =>
                                updateCondition(ruleIndex, conditionIndex, {
                                  operator: value as any,
                                })
                              }
                              options={operators}
                            />
                            <Input
                              label="Value"
                              classNames={{ label: "text-xs" }}
                              value={condition.value as string}
                              onChange={(e) =>
                                updateCondition(ruleIndex, conditionIndex, {
                                  value: e.target.value,
                                })
                              }
                              placeholder="e.g., premium"
                            />
                            <Button
                              variant="ghost"
                              size="sm"
                              onClick={() =>
                                removeCondition(ruleIndex, conditionIndex)
                              }
                            >
                              <X className="h-4 w-4" />
                            </Button>
                          </div>
                        ))}
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </TabsContent>

          {/* Settings Tab */}
          <TabsContent value="settings" className="space-y-4">
            <div className="space-y-4">
              <div className="flex items-center space-x-2">
                <Switch
                  checked={formData.enabled}
                  onCheckedChange={(checked) =>
                    handleInputChange("enabled", checked)
                  }
                />
                <Label>Enable flag immediately</Label>
              </div>

              <div className="space-y-2">
                <Label className="flex items-center gap-2">
                  <Calendar className="h-4 w-4" />
                  Expiry Date (Optional)
                </Label>
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
                      {expiryDate
                        ? format(expiryDate, "PPP")
                        : "Set expiry date"}
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

              <Input
                label="Default Value"
                value={formData.default_value}
                onChange={(e) =>
                  handleInputChange("default_value", e.target.value)
                }
                placeholder="Default value when flag is disabled"
              />

              <Alert>
                <Info className="h-4 w-4" />
                <AlertDescription>
                  The default value is returned when the flag is disabled or
                  when targeting rules don't match.
                </AlertDescription>
              </Alert>
            </div>
          </TabsContent>
        </Tabs>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button
            onClick={handleSave}
            disabled={isLoading}
            isLoading={isLoading}
            loadingText="Saving..."
          >
            {isEditing ? "Update Flag" : "Create Flag"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
