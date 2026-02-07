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
import {
  AlertTriangle,
  Eye,
  EyeOff,
  Info,
  Settings,
  Shield,
  Zap,
  Link,
  Bell,
  Palette,
  RotateCcw,
} from "lucide-react";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { cn } from "@/lib/utils";
import type { SystemSetting } from "@/app/_actions/settings";

interface SettingEditDialogProps {
  setting?: SystemSetting | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSave: (
    setting: Omit<SystemSetting, "id" | "lastModified" | "modifiedBy">,
  ) => void;
  isLoading?: boolean;
}

export function SettingEditDialog({
  setting,
  open,
  onOpenChange,
  onSave,
  isLoading = false,
}: SettingEditDialogProps) {
  const [formData, setFormData] = useState({
    key: "",
    value: "",
    type: "string" as SystemSetting["type"],
    category: "general" as SystemSetting["category"],
    description: "",
    defaultValue: "",
    isRequired: false,
    isSecret: false,
    environment: "all" as SystemSetting["environment"],
    validation: {
      min: undefined as number | undefined,
      max: undefined as number | undefined,
      pattern: "",
      options: [] as string[],
    },
  });

  const [showValue, setShowValue] = useState(false);
  const [validationErrors, setValidationErrors] = useState<
    Record<string, string>
  >({});
  const [optionsInput, setOptionsInput] = useState("");

  const isEditing = !!setting;

  const categories = [
    { value: "general", label: "General", icon: Settings },
    { value: "security", label: "Security", icon: Shield },
    { value: "performance", label: "Performance", icon: Zap },
    { value: "integration", label: "Integration", icon: Link },
    { value: "notification", label: "Notification", icon: Bell },
    { value: "ui", label: "User Interface", icon: Palette },
  ];

  const types = [
    { value: "string", label: "String" },
    { value: "number", label: "Number" },
    { value: "boolean", label: "Boolean" },
    { value: "json", label: "JSON" },
    { value: "array", label: "Array" },
  ];

  const environments = [
    { value: "all", label: "All Environments" },
    { value: "production", label: "Production" },
    { value: "staging", label: "Staging" },
    { value: "development", label: "Development" },
  ];

  useEffect(() => {
    if (setting) {
      setFormData({
        key: setting.key,
        value: setting.value,
        type: setting.type,
        category: setting.category,
        description: setting.description,
        defaultValue: setting.defaultValue,
        isRequired: setting.isRequired,
        isSecret: setting.isSecret,
        environment: setting.environment,
        validation: {
          min: setting.validation?.min,
          max: setting.validation?.max,
          pattern: setting.validation?.pattern || "",
          options: setting.validation?.options || [],
        },
      });
      setOptionsInput(setting.validation?.options?.join(", ") || "");
      setShowValue(!setting.isSecret);
    } else {
      setFormData({
        key: "",
        value: "",
        type: "string",
        category: "general",
        description: "",
        defaultValue: "",
        isRequired: false,
        isSecret: false,
        environment: "all",
        validation: {
          min: undefined,
          max: undefined,
          pattern: "",
          options: [],
        },
      });
      setOptionsInput("");
      setShowValue(true);
    }
    setValidationErrors({});
  }, [setting, open]);

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

  const handleValidationChange = (field: string, value: any) => {
    setFormData((prev) => ({
      ...prev,
      validation: {
        ...prev.validation,
        [field]: value,
      },
    }));
  };

  const handleOptionsChange = (value: string) => {
    setOptionsInput(value);
    const options = value
      .split(",")
      .map((opt) => opt.trim())
      .filter((opt) => opt.length > 0);
    handleValidationChange("options", options);
  };

  const validateForm = () => {
    const errors: Record<string, string> = {};

    if (!formData.key.trim()) {
      errors.key = "Setting key is required";
    } else if (!/^[a-zA-Z][a-zA-Z0-9._]*$/.test(formData.key)) {
      errors.key =
        "Key must start with a letter and contain only letters, numbers, dots, and underscores";
    }

    if (!formData.description.trim()) {
      errors.description = "Description is required";
    }

    if (formData.type === "number") {
      if (formData.value && isNaN(Number(formData.value))) {
        errors.value = "Value must be a valid number";
      }
      if (formData.defaultValue && isNaN(Number(formData.defaultValue))) {
        errors.defaultValue = "Default value must be a valid number";
      }
    }

    if (formData.type === "boolean") {
      if (
        formData.value &&
        !["true", "false"].includes(formData.value.toLowerCase())
      ) {
        errors.value = "Value must be true or false";
      }
      if (
        formData.defaultValue &&
        !["true", "false"].includes(formData.defaultValue.toLowerCase())
      ) {
        errors.defaultValue = "Default value must be true or false";
      }
    }

    if (formData.type === "json") {
      if (formData.value) {
        try {
          JSON.parse(formData.value);
        } catch {
          errors.value = "Value must be valid JSON";
        }
      }
      if (formData.defaultValue) {
        try {
          JSON.parse(formData.defaultValue);
        } catch {
          errors.defaultValue = "Default value must be valid JSON";
        }
      }
    }

    if (formData.type === "array") {
      if (formData.value) {
        try {
          const parsed = JSON.parse(formData.value);
          if (!Array.isArray(parsed)) {
            errors.value = "Value must be a valid JSON array";
          }
        } catch {
          errors.value = "Value must be a valid JSON array";
        }
      }
      if (formData.defaultValue) {
        try {
          const parsed = JSON.parse(formData.defaultValue);
          if (!Array.isArray(parsed)) {
            errors.defaultValue = "Default value must be a valid JSON array";
          }
        } catch {
          errors.defaultValue = "Default value must be a valid JSON array";
        }
      }
    }

    if (
      formData.validation.min !== undefined &&
      formData.validation.max !== undefined
    ) {
      if (formData.validation.min > formData.validation.max) {
        errors.validation =
          "Minimum value cannot be greater than maximum value";
      }
    }

    setValidationErrors(errors);
    return Object.keys(errors).length === 0;
  };

  const handleSave = () => {
    if (!validateForm()) return;

    const settingData = {
      ...formData,
      validation: {
        ...formData.validation,
        min: formData.validation.min || undefined,
        max: formData.validation.max || undefined,
        pattern: formData.validation.pattern || undefined,
        options:
          formData.validation.options.length > 0
            ? formData.validation.options
            : undefined,
      },
    };

    // Prepare validation object
    const hasValidation =
      settingData.validation.min ||
      settingData.validation.max ||
      settingData.validation.pattern ||
      (settingData.validation.options &&
        settingData.validation.options.length > 0);

    const finalSettingData = {
      ...settingData,
      ...(hasValidation ? { validation: settingData.validation } : {}),
    };

    onSave(finalSettingData);
  };

  const resetToDefault = () => {
    setFormData((prev) => ({
      ...prev,
      value: prev.defaultValue,
    }));
  };

  const selectedCategory = categories.find(
    (cat) => cat.value === formData.category,
  );
  const CategoryIcon = selectedCategory?.icon || Settings;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <CategoryIcon className="h-5 w-5" />
            {isEditing ? "Edit Setting" : "Create Setting"}
          </DialogTitle>
          <DialogDescription>
            {isEditing
              ? "Modify the system setting configuration."
              : "Create a new system setting configuration."}
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-6">
          {/* Basic Information */}
          <div className="space-y-4">
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="key">Setting Key *</Label>
                <Input
                  id="key"
                  value={formData.key}
                  onChange={(e) => handleInputChange("key", e.target.value)}
                  placeholder="e.g., app.session_timeout"
                  disabled={isEditing}
                  className={cn(validationErrors.key && "border-red-500")}
                />
                {validationErrors.key && (
                  <p className="text-sm text-red-600">{validationErrors.key}</p>
                )}
              </div>

              <div className="space-y-2">
                <Label htmlFor="type">Type *</Label>
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
            </div>

            <div className="space-y-2">
              <Label htmlFor="description">Description *</Label>
              <Textarea
                id="description"
                value={formData.description}
                onChange={(e) =>
                  handleInputChange("description", e.target.value)
                }
                placeholder="Describe what this setting controls..."
                className={cn(validationErrors.description && "border-red-500")}
              />
              {validationErrors.description && (
                <p className="text-sm text-red-600">
                  {validationErrors.description}
                </p>
              )}
            </div>
          </div>

          <Separator />

          {/* Value Configuration */}
          <div className="space-y-4">
            <h3 className="text-lg font-medium">Value Configuration</h3>

            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="value" className="flex items-center gap-2">
                  Current Value
                  {formData.defaultValue && (
                    <Button
                      type="button"
                      variant="ghost"
                      size="sm"
                      onClick={resetToDefault}
                      className="h-6 px-2 text-xs"
                    >
                      <RotateCcw className="h-3 w-3 mr-1" />
                      Reset
                    </Button>
                  )}
                </Label>
                <div className="relative">
                  {formData.type === "json" || formData.type === "array" ? (
                    <Textarea
                      id="value"
                      value={formData.value}
                      onChange={(e) =>
                        handleInputChange("value", e.target.value)
                      }
                      placeholder={
                        formData.type === "json"
                          ? '{"key": "value"}'
                          : '["item1", "item2"]'
                      }
                      className={cn(
                        "font-mono text-sm",
                        validationErrors.value && "border-red-500",
                        formData.isSecret && !showValue && "text-security-disc",
                      )}
                    />
                  ) : (
                    <Input
                      id="value"
                      type={
                        formData.isSecret && !showValue ? "password" : "text"
                      }
                      value={formData.value}
                      onChange={(e) =>
                        handleInputChange("value", e.target.value)
                      }
                      placeholder={
                        formData.type === "boolean"
                          ? "true or false"
                          : formData.type === "number"
                            ? "123"
                            : "Enter value..."
                      }
                      className={cn(validationErrors.value && "border-red-500")}
                    />
                  )}
                  {formData.isSecret && (
                    <Button
                      type="button"
                      variant="ghost"
                      size="sm"
                      onClick={() => setShowValue(!showValue)}
                      className="absolute right-2 top-1/2 transform -translate-y-1/2 h-6 w-6 p-0"
                    >
                      {showValue ? (
                        <EyeOff className="h-3 w-3" />
                      ) : (
                        <Eye className="h-3 w-3" />
                      )}
                    </Button>
                  )}
                </div>
                {validationErrors.value && (
                  <p className="text-sm text-red-600">
                    {validationErrors.value}
                  </p>
                )}
              </div>

              <div className="space-y-2">
                <Label htmlFor="defaultValue">Default Value</Label>
                {formData.type === "json" || formData.type === "array" ? (
                  <Textarea
                    id="defaultValue"
                    value={formData.defaultValue}
                    onChange={(e) =>
                      handleInputChange("defaultValue", e.target.value)
                    }
                    placeholder={
                      formData.type === "json"
                        ? '{"key": "value"}'
                        : '["item1", "item2"]'
                    }
                    className={cn(
                      "font-mono text-sm",
                      validationErrors.defaultValue && "border-red-500",
                    )}
                  />
                ) : (
                  <Input
                    id="defaultValue"
                    value={formData.defaultValue}
                    onChange={(e) =>
                      handleInputChange("defaultValue", e.target.value)
                    }
                    placeholder={
                      formData.type === "boolean"
                        ? "true or false"
                        : formData.type === "number"
                          ? "123"
                          : "Enter default value..."
                    }
                    className={cn(
                      validationErrors.defaultValue && "border-red-500",
                    )}
                  />
                )}
                {validationErrors.defaultValue && (
                  <p className="text-sm text-red-600">
                    {validationErrors.defaultValue}
                  </p>
                )}
              </div>
            </div>
          </div>

          <Separator />

          {/* Configuration */}
          <div className="space-y-4">
            <h3 className="text-lg font-medium">Configuration</h3>

            <div className="grid grid-cols-2 gap-4">
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

            <div className="flex items-center space-x-6">
              <div className="flex items-center space-x-2">
                <Switch
                  id="isRequired"
                  checked={formData.isRequired}
                  onCheckedChange={(checked) =>
                    handleInputChange("isRequired", checked)
                  }
                />
                <Label htmlFor="isRequired" className="flex items-center gap-1">
                  Required Setting
                  <AlertTriangle className="h-3 w-3 text-red-500" />
                </Label>
              </div>

              <div className="flex items-center space-x-2">
                <Switch
                  id="isSecret"
                  checked={formData.isSecret}
                  onCheckedChange={(checked) =>
                    handleInputChange("isSecret", checked)
                  }
                />
                <Label htmlFor="isSecret" className="flex items-center gap-1">
                  Secret Setting
                  <Eye className="h-3 w-3 text-amber-500" />
                </Label>
              </div>
            </div>
          </div>

          {/* Validation Rules */}
          {(formData.type === "string" || formData.type === "number") && (
            <>
              <Separator />
              <div className="space-y-4">
                <h3 className="text-lg font-medium">Validation Rules</h3>

                {formData.type === "number" && (
                  <div className="grid grid-cols-2 gap-4">
                    <div className="space-y-2">
                      <Label htmlFor="min">Minimum Value</Label>
                      <Input
                        id="min"
                        type="number"
                        value={formData.validation.min || ""}
                        onChange={(e) =>
                          handleValidationChange(
                            "min",
                            e.target.value ? Number(e.target.value) : undefined,
                          )
                        }
                        placeholder="No minimum"
                      />
                    </div>
                    <div className="space-y-2">
                      <Label htmlFor="max">Maximum Value</Label>
                      <Input
                        id="max"
                        type="number"
                        value={formData.validation.max || ""}
                        onChange={(e) =>
                          handleValidationChange(
                            "max",
                            e.target.value ? Number(e.target.value) : undefined,
                          )
                        }
                        placeholder="No maximum"
                      />
                    </div>
                  </div>
                )}

                {formData.type === "string" && (
                  <>
                    <div className="space-y-2">
                      <Label htmlFor="pattern">
                        Validation Pattern (Regex)
                      </Label>
                      <Input
                        id="pattern"
                        value={formData.validation.pattern || ""}
                        onChange={(e) =>
                          handleValidationChange("pattern", e.target.value)
                        }
                        placeholder="e.g., ^[a-zA-Z0-9]+$"
                        className="font-mono text-sm"
                      />
                    </div>

                    <div className="space-y-2">
                      <Label htmlFor="options">
                        Allowed Values (comma-separated)
                      </Label>
                      <Input
                        id="options"
                        value={optionsInput}
                        onChange={(e) => handleOptionsChange(e.target.value)}
                        placeholder="e.g., option1, option2, option3"
                      />
                      {formData.validation.options.length > 0 && (
                        <div className="flex flex-wrap gap-1 mt-2">
                          {formData.validation.options.map((option, index) => (
                            <Badge key={index} variant="secondary">
                              {option}
                            </Badge>
                          ))}
                        </div>
                      )}
                    </div>
                  </>
                )}

                {validationErrors.validation && (
                  <Alert>
                    <AlertTriangle className="h-4 w-4" />
                    <AlertDescription>
                      {validationErrors.validation}
                    </AlertDescription>
                  </Alert>
                )}
              </div>
            </>
          )}

          {/* Warnings */}
          {formData.isRequired && (
            <Alert>
              <AlertTriangle className="h-4 w-4" />
              <AlertDescription>
                This setting is marked as required. The system may not function
                properly if this setting is not configured.
              </AlertDescription>
            </Alert>
          )}

          {formData.isSecret && (
            <Alert>
              <Info className="h-4 w-4" />
              <AlertDescription>
                This setting is marked as secret. Its value will be hidden in
                the UI and encrypted in storage.
              </AlertDescription>
            </Alert>
          )}
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button onClick={handleSave} disabled={isLoading}>
            {isLoading
              ? "Saving..."
              : isEditing
                ? "Update Setting"
                : "Create Setting"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
