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
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Checkbox } from "@/components/ui/checkbox";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { Settings, AlertTriangle, Loader2, Shield } from "lucide-react";
import { toast } from "sonner";
import {
  getAllSubscriptionFeatures,
  overrideOrganizationLimits,
  type SubscriptionFeature,
} from "@/app/_actions/subscriptions";
import { type Organization } from "@/app/_actions/organizations";

interface OverrideLimitsDialogProps {
  organization: Organization;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess: () => void;
}

export function OverrideLimitsDialog({
  organization,
  open,
  onOpenChange,
  onSuccess,
}: OverrideLimitsDialogProps) {
  const [availableFeatures, setAvailableFeatures] = useState<
    SubscriptionFeature[]
  >([]);
  const [maxUsers, setMaxUsers] = useState<number | undefined>(undefined);
  const [storageLimitGb, setStorageLimitGb] = useState<number | undefined>(
    undefined,
  );
  const [selectedFeatures, setSelectedFeatures] = useState<string[]>([]);
  const [reason, setReason] = useState("");
  const [expiresAt, setExpiresAt] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [isLoadingFeatures, setIsLoadingFeatures] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    if (open) {
      loadFeatures();
      setMaxUsers(undefined);
      setStorageLimitGb(undefined);
      setSelectedFeatures([]);
      setReason("");
      setExpiresAt("");
      setError("");
    }
  }, [open]);

  const loadFeatures = async () => {
    setIsLoadingFeatures(true);
    try {
      const result = await getAllSubscriptionFeatures();
      if (result.success && result.data) {
        setAvailableFeatures(result.data.filter((f) => f.isActive));
      }
    } catch {
      toast.error("Failed to load features");
    } finally {
      setIsLoadingFeatures(false);
    }
  };

  const toggleFeature = (featureName: string) => {
    setSelectedFeatures((prev) =>
      prev.includes(featureName)
        ? prev.filter((f) => f !== featureName)
        : [...prev, featureName],
    );
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");

    if (reason.trim().length < 10) {
      setError("Reason must be at least 10 characters");
      return;
    }

    if (
      maxUsers === undefined &&
      storageLimitGb === undefined &&
      selectedFeatures.length === 0
    ) {
      setError(
        "Please specify at least one override (users, storage, or features)",
      );
      return;
    }

    setIsLoading(true);
    try {
      const request: {
        max_users?: number;
        storage_limit_gb?: number;
        features?: string[];
        reason: string;
        expires_at?: string;
      } = {
        reason: reason.trim(),
      };

      if (maxUsers !== undefined) request.max_users = maxUsers;
      if (storageLimitGb !== undefined)
        request.storage_limit_gb = storageLimitGb;
      if (selectedFeatures.length > 0) request.features = selectedFeatures;
      if (expiresAt) request.expires_at = new Date(expiresAt).toISOString();

      const result = await overrideOrganizationLimits(organization.id, request);

      if (result.success) {
        toast.success("Organization limits overridden successfully");
        onOpenChange(false);
        onSuccess();
      } else {
        setError(result.message || "Failed to override limits");
      }
    } catch {
      setError("Failed to override limits");
    } finally {
      setIsLoading(false);
    }
  };

  const currentFeatures = organization.settings?.features_enabled || [];

  // Group features by category
  const featuresByCategory = availableFeatures.reduce(
    (acc, feature) => {
      const category = feature.category || "Other";
      if (!acc[category]) acc[category] = [];
      acc[category].push(feature);
      return acc;
    },
    {} as Record<string, SubscriptionFeature[]>,
  );

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Shield className="h-5 w-5" />
            Override Organization Limits
          </DialogTitle>
          <DialogDescription>
            Temporarily override limits for <strong>{organization.name}</strong>
            . These overrides take precedence over tier defaults.
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4">
          {/* Current Settings */}
          <div className="rounded-lg border p-3 bg-muted/50">
            <p className="text-sm font-medium mb-2">Current Settings</p>
            <div className="grid grid-cols-2 gap-2 text-sm">
              <div>
                <span className="text-muted-foreground">Tier:</span>{" "}
                <Badge variant="outline">
                  {organization.subscription_tier}
                </Badge>
              </div>
              <div>
                <span className="text-muted-foreground">Max Users:</span>{" "}
                {organization.settings?.max_users || "Default"}
              </div>
              <div className="col-span-2">
                <span className="text-muted-foreground">Features:</span>{" "}
                {currentFeatures.length > 0 ? (
                  <span className="text-xs">{currentFeatures.join(", ")}</span>
                ) : (
                  "None"
                )}
              </div>
            </div>
          </div>

          <Separator />

          {/* Limit Overrides */}
          <Card>
            <CardHeader className="pb-3">
              <CardTitle className="text-sm flex items-center gap-2">
                <Settings className="h-4 w-4" />
                Limit Overrides
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <label className="text-sm font-medium">
                    Max Users Override
                  </label>
                  <Input
                    type="number"
                    min="1"
                    placeholder="Leave empty to keep default"
                    value={maxUsers ?? ""}
                    onChange={(e) =>
                      setMaxUsers(
                        e.target.value ? parseInt(e.target.value) : undefined,
                      )
                    }
                  />
                </div>
                <div className="space-y-2">
                  <label className="text-sm font-medium">
                    Storage Limit (GB)
                  </label>
                  <Input
                    type="number"
                    min="1"
                    placeholder="Leave empty to keep default"
                    value={storageLimitGb ?? ""}
                    onChange={(e) =>
                      setStorageLimitGb(
                        e.target.value ? parseInt(e.target.value) : undefined,
                      )
                    }
                  />
                </div>
              </div>

              <div className="space-y-2">
                <label className="text-sm font-medium">
                  Override Expiration
                </label>
                <Input
                  type="date"
                  value={expiresAt}
                  onChange={(e) => setExpiresAt(e.target.value)}
                  min={new Date().toISOString().split("T")[0]}
                />
                <p className="text-xs text-muted-foreground">
                  Leave empty for a permanent override
                </p>
              </div>
            </CardContent>
          </Card>

          {/* Feature Overrides */}
          <Card>
            <CardHeader className="pb-3">
              <CardTitle className="text-sm">
                Additional Features Grant
              </CardTitle>
            </CardHeader>
            <CardContent>
              {isLoadingFeatures ? (
                <div className="flex items-center gap-2 text-sm text-muted-foreground">
                  <Loader2 className="h-4 w-4 animate-spin" />
                  Loading features...
                </div>
              ) : (
                <div className="space-y-4">
                  {Object.entries(featuresByCategory).map(
                    ([category, features]) => (
                      <div key={category}>
                        <p className="text-xs font-medium text-muted-foreground uppercase tracking-wider mb-2">
                          {category}
                        </p>
                        <div className="grid grid-cols-2 gap-2">
                          {features.map((feature) => {
                            const alreadyEnabled = currentFeatures.includes(
                              feature.name,
                            );
                            return (
                              <div
                                key={feature.id}
                                className="flex items-center space-x-2"
                              >
                                <Checkbox
                                  id={feature.id}
                                  checked={
                                    alreadyEnabled ||
                                    selectedFeatures.includes(feature.name)
                                  }
                                  disabled={alreadyEnabled}
                                  onCheckedChange={() =>
                                    toggleFeature(feature.name)
                                  }
                                />
                                <label
                                  htmlFor={feature.id}
                                  className={`text-sm leading-none ${alreadyEnabled ? "text-muted-foreground" : ""}`}
                                >
                                  {feature.displayName || feature.name}
                                  {alreadyEnabled && (
                                    <span className="text-xs ml-1">
                                      (active)
                                    </span>
                                  )}
                                </label>
                              </div>
                            );
                          })}
                        </div>
                      </div>
                    ),
                  )}
                </div>
              )}
            </CardContent>
          </Card>

          {/* Reason */}
          <div className="space-y-2">
            <label className="text-sm font-medium">
              Reason for Override <span className="text-red-500">*</span>
            </label>
            <Textarea
              value={reason}
              onChange={(e) => setReason(e.target.value)}
              placeholder="Explain why these limits are being overridden (min 10 characters)..."
              rows={3}
            />
            <p className="text-xs text-muted-foreground">
              {reason.length}/10 minimum characters
            </p>
          </div>

          {error && (
            <div className="flex items-center gap-2 text-sm text-red-600 bg-red-50 dark:bg-red-950/20 p-2 rounded">
              <AlertTriangle className="h-4 w-4 flex-shrink-0" />
              {error}
            </div>
          )}

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
              disabled={isLoading}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={isLoading}>
              {isLoading ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Applying...
                </>
              ) : (
                "Apply Overrides"
              )}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
