"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Badge } from "@/components/ui/badge";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  CreditCard,
  ArrowRight,
  AlertTriangle,
  Check,
} from "lucide-react";
import { changeOrganizationTier } from "@/app/_actions/subscriptions";
import { toast } from "sonner";
import {
  useSubscriptionTiers,
  useSubscriptionFeatures,
} from "@/hooks/use-subscriptions";

interface ChangeTierDialogProps {
  organization: {
    id: string;
    name: string;
    subscription_tier: string;
    subscription_status: string;
  };
  onSuccess: () => void;
}

export function ChangeTierDialog({
  organization,
  onSuccess,
}: ChangeTierDialogProps) {
  const [isOpen, setIsOpen] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [selectedTier, setSelectedTier] = useState(
    organization.subscription_tier,
  );
  const [reason, setReason] = useState("");
  const [error, setError] = useState("");

  const { data: tiers, isLoading: tiersLoading } = useSubscriptionTiers();
  const { data: features } = useSubscriptionFeatures();

  const currentTier = tiers?.find(
    (t) => t.name === organization.subscription_tier,
  );
  const newTier = tiers?.find((t) => t.name === selectedTier);

  const getTierFeatures = (tier: any) => {
    if (!tier || !features) return [];
    const tierFeatureNames = Array.isArray(tier.features) ? tier.features : [];
    return tierFeatureNames
      .map((featureName: string) =>
        features.find((f: any) => f.name === featureName),
      )
      .filter(Boolean);
  };

  const currentTierFeatures = getTierFeatures(currentTier);
  const newTierFeatures = getTierFeatures(newTier);

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setError("");

    if (selectedTier === organization.subscription_tier) {
      setError("Please select a different tier");
      return;
    }
    if (!reason.trim() || reason.length < 10) {
      setError("Reason must be at least 10 characters long");
      return;
    }

    setIsLoading(true);
    try {
      const result = await changeOrganizationTier(organization.id, {
        newTier: selectedTier,
        reason: reason.trim(),
      });

      if (result.success) {
        toast.success(
          `Subscription tier changed to ${newTier?.displayName || selectedTier}`,
        );
        handleClose();
        onSuccess();
      } else {
        setError(result.message || "Failed to change subscription tier");
      }
    } catch (err) {
      setError(
        err instanceof Error ? err.message : "Failed to change subscription tier",
      );
    } finally {
      setIsLoading(false);
    }
  };

  const handleClose = () => {
    setIsOpen(false);
    setError("");
    setReason("");
    setSelectedTier(organization.subscription_tier);
  };

  return (
    <>
      <Button
        variant="outline"
        size="sm"
        onClick={() => setIsOpen(true)}
        disabled={tiersLoading}
        isLoading={tiersLoading}
        loadingText="Change Tier"
      >
        <CreditCard className="mr-2 h-4 w-4" />
        Change Tier
      </Button>

      <Dialog open={isOpen} onOpenChange={(open) => !open && handleClose()}>
        <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <CreditCard className="h-5 w-5" />
              Change Subscription Tier
            </DialogTitle>
            <DialogDescription>
              Change the subscription tier for {organization.name}
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-6">
            {/* Current Tier Info */}
            <div className="rounded-lg border p-4 bg-muted/50">
              <div className="flex items-center justify-between mb-2">
                <span className="font-medium">{organization.name}</span>
                <Badge variant="outline">
                  {organization.subscription_status}
                </Badge>
              </div>
              <div className="flex items-center gap-2 text-sm">
                <span className="text-muted-foreground">Current Tier:</span>
                <Badge variant="default">
                  {currentTier?.displayName || organization.subscription_tier}
                </Badge>
                <span className="text-muted-foreground">
                  ${currentTier?.priceMonthly || 0}/month
                </span>
              </div>
            </div>

            <form onSubmit={handleSubmit} className="space-y-6">
              {/* Tier Selection */}
              <div>
                <label className="text-sm font-medium mb-2 block">
                  Select New Tier
                </label>
                <Select value={selectedTier} onValueChange={setSelectedTier}>
                  <SelectTrigger>
                    <SelectValue placeholder="Select a tier" />
                  </SelectTrigger>
                  <SelectContent>
                    {(tiers || []).map((tier) => (
                      <SelectItem
                        key={tier.id}
                        value={tier.name}
                        disabled={tier.name === organization.subscription_tier}
                      >
                        <div className="flex items-center justify-between w-full">
                          <span className="font-medium">{tier.displayName}</span>
                          <span className="text-muted-foreground text-sm ml-4">
                            ${tier.priceMonthly}/month
                          </span>
                        </div>
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              {/* Tier Comparison */}
              {selectedTier !== organization.subscription_tier && newTier && (
                <div className="grid grid-cols-9 gap-4">
                  <div className="rounded-lg border col-span-4 p-4">
                    <div className="text-sm font-medium mb-2 text-muted-foreground">
                      Current
                    </div>
                    <div className="space-y-2">
                      <div className="font-semibold">
                        {currentTier?.displayName}
                      </div>
                      <div className="text-sm text-muted-foreground">
                        ${currentTier?.priceMonthly}/month
                      </div>
                      <div className="text-xs space-y-1">
                        <div>
                          👥{" "}
                          {currentTier?.maxTeamMembers === -1
                            ? "Unlimited"
                            : currentTier?.maxTeamMembers}{" "}
                          users
                        </div>
                        <div>
                          📄{" "}
                          {currentTier?.maxDocuments === -1
                            ? "Unlimited"
                            : currentTier?.maxDocuments}{" "}
                          documents
                        </div>
                      </div>
                    </div>
                  </div>

                  <div className="flex items-center justify-center">
                    <ArrowRight className="h-6 w-6 text-muted-foreground" />
                  </div>

                  <div className="rounded-lg border col-span-4 p-4 bg-primary/5 border-primary/20">
                    <div className="text-sm font-medium mb-2 text-primary">
                      New Tier
                    </div>
                    <div className="space-y-2">
                      <div className="font-semibold">{newTier.displayName}</div>
                      <div className="text-sm text-muted-foreground">
                        ${newTier.priceMonthly}/month
                      </div>
                      <div className="text-xs space-y-1">
                        <div>
                          👥{" "}
                          {newTier.maxTeamMembers === -1
                            ? "Unlimited"
                            : newTier.maxTeamMembers}{" "}
                          users
                        </div>
                        <div>
                          📄{" "}
                          {newTier.maxDocuments === -1
                            ? "Unlimited"
                            : newTier.maxDocuments}{" "}
                          documents
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              )}

              {/* Features Preview */}
              {selectedTier !== organization.subscription_tier && newTier && (
                <div className="rounded-lg border p-4 bg-muted/30">
                  {(() => {
                    const sortedTiers = [...(tiers || [])].sort(
                      (a, b) => a.sortOrder - b.sortOrder,
                    );
                    const newTierIndex = sortedTiers.findIndex(
                      (t) => t.name === newTier.name,
                    );
                    const previousTier =
                      newTierIndex > 0 ? sortedTiers[newTierIndex - 1] : null;
                    const previousFeatureNames = new Set(
                      previousTier?.features || [],
                    );
                    const uniqueFeatures = newTierFeatures
                      .filter((f: any) => !previousFeatureNames.has(f.name))
                      .slice(0, 10);

                    return (
                      <>
                        {previousTier ? (
                          <div className="flex items-center gap-2 px-3 py-2 rounded-lg bg-primary/10 border border-primary/20 mb-3">
                            <ArrowRight className="h-3.5 w-3.5 text-primary" />
                            <span className="text-sm font-medium text-primary">
                              Everything in {previousTier.displayName}, plus:
                            </span>
                          </div>
                        ) : (
                          <div className="text-sm font-medium mb-3">
                            Features Included:
                          </div>
                        )}
                        <div className="space-y-2 max-h-60 overflow-y-auto">
                          {uniqueFeatures.map((feature: any) => (
                            <div
                              key={feature.id}
                              className="flex items-start gap-2 text-sm"
                            >
                              <Check className="h-4 w-4 text-green-600 shrink-0 mt-0.5" />
                              <div className="flex-1">
                                <div className="font-medium">
                                  {feature.displayName}
                                </div>
                                {feature.description && (
                                  <div className="text-xs text-muted-foreground mt-0.5">
                                    {feature.description}
                                  </div>
                                )}
                              </div>
                            </div>
                          ))}
                        </div>
                      </>
                    );
                  })()}
                </div>
              )}

              {/* Reason */}
              <div>
                <label htmlFor="topLevelReason" className="text-sm font-medium">
                  Reason for Change *
                </label>
                <Textarea
                  id="topLevelReason"
                  required
                  value={reason}
                  onChange={(e) => setReason(e.target.value)}
                  placeholder="Explain why you're changing the subscription tier (e.g., customer upgrade request, trial conversion, special arrangement)..."
                  className="mt-1"
                  rows={3}
                />
                <p className="text-xs text-muted-foreground mt-1">
                  Minimum 10 characters required. This will be logged for audit
                  purposes.
                </p>
              </div>

              {error && (
                <div className="flex items-center gap-2 text-sm text-red-600 bg-red-50 p-3 rounded">
                  <AlertTriangle className="h-4 w-4 shrink-0" />
                  {error}
                </div>
              )}

              <div className="flex gap-2 pt-2">
                <Button
                  type="button"
                  variant="outline"
                  onClick={handleClose}
                  disabled={isLoading}
                  className="flex-1"
                >
                  Cancel
                </Button>
                <Button
                  type="submit"
                  disabled={
                    isLoading ||
                    !reason.trim() ||
                    selectedTier === organization.subscription_tier
                  }
                  isLoading={isLoading}
                  loadingText="Changing Tier..."
                  className="flex-1"
                >
                  <CreditCard className="mr-2 h-4 w-4" />
                  Change Tier
                </Button>
              </div>
            </form>
          </div>
        </DialogContent>
      </Dialog>
    </>
  );
}
