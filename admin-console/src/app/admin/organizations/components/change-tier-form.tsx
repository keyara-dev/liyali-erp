"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { Checkbox } from "@/components/ui/checkbox";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { AlertTriangle, ArrowUpDown, Loader2 } from "lucide-react";
import { toast } from "sonner";
import {
  changeOrganizationTier,
  type SubscriptionTier,
  type SubscriptionFeature,
} from "@/app/_actions/subscriptions";
import { getTierBadge } from "@/lib/tier-utils";

interface ChangeTierFormProps {
  organizationId: string;
  currentTier: string;
  tiers: SubscriptionTier[];
  isLoadingTiers?: boolean;
  /** Show the override-limits checkbox (used by org change-tier dialog and manage-sub tab) */
  showOverrideLimits?: boolean;
  /** Show a side-by-side tier comparison preview (used by top-level change-tier dialog) */
  showTierComparison?: boolean;
  /** Required when showTierComparison is true */
  features?: SubscriptionFeature[];
  onSuccess: () => void;
  /** If true, submit button spans full width */
  fullWidthSubmit?: boolean;
}

export function ChangeTierForm({
  organizationId,
  currentTier,
  tiers,
  isLoadingTiers = false,
  showOverrideLimits = false,
  showTierComparison = false,
  features = [],
  onSuccess,
  fullWidthSubmit = false,
}: ChangeTierFormProps) {
  const [selectedTier, setSelectedTier] = useState("");
  const [reason, setReason] = useState("");
  const [overrideLimits, setOverrideLimits] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState("");

  const selectedTierDetails = tiers.find((t) => t.name === selectedTier);

  // Unique features the selected tier adds vs current
  const tierComparisonFeatures = showTierComparison
    ? features
        .filter(
          (f) =>
            selectedTierDetails?.features?.includes(f.name) &&
            !f.name.startsWith("_"),
        )
        .slice(0, 10)
    : [];

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");

    if (!selectedTier) {
      setError("Please select a tier");
      return;
    }
    if (selectedTier === currentTier) {
      setError("Please select a different tier than the current one");
      return;
    }
    if (reason.trim().length < 10) {
      setError("Reason must be at least 10 characters");
      return;
    }

    setIsLoading(true);
    try {
      const result = await changeOrganizationTier(organizationId, {
        newTier: selectedTier,
        reason: reason.trim(),
        overrideLimits,
      });

      if (result.success) {
        toast.success("Organization tier changed successfully");
        setSelectedTier("");
        setReason("");
        setOverrideLimits(false);
        onSuccess();
      } else {
        setError(result.message || "Failed to change tier");
      }
    } catch {
      setError("Failed to change tier");
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      {/* Current tier */}
      <div className="rounded-lg border p-3 bg-muted/50">
        <div className="flex items-center justify-between">
          <span className="text-sm font-medium">Current Tier:</span>
          {getTierBadge(currentTier)}
        </div>
      </div>

      {/* Tier select */}
      <div className="space-y-2">
        <label className="text-sm font-medium">New Tier</label>
        {isLoadingTiers ? (
          <div className="flex items-center gap-2 text-sm text-muted-foreground">
            <Loader2 className="h-4 w-4 animate-spin" />
            Loading tiers...
          </div>
        ) : (
          <Select value={selectedTier} onValueChange={setSelectedTier}>
            <SelectTrigger>
              <SelectValue placeholder="Select a tier" />
            </SelectTrigger>
            <SelectContent>
              {tiers.map((tier) => (
                <SelectItem
                  key={tier.id}
                  value={tier.name}
                  disabled={tier.name === currentTier}
                >
                  <div className="flex items-center gap-2">
                    <span>{tier.displayName || tier.name}</span>
                    <span className="text-xs text-muted-foreground">
                      ${tier.priceMonthly}/mo &middot; {tier.maxTeamMembers}{" "}
                      users
                    </span>
                  </div>
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        )}
      </div>

      {/* Selected tier detail card */}
      {selectedTierDetails && (
        <div className="rounded-lg border p-3 bg-blue-50 dark:bg-blue-950/20 space-y-1">
          <p className="text-sm font-medium">
            {selectedTierDetails.displayName || selectedTierDetails.name}
          </p>
          <p className="text-xs text-muted-foreground">
            {selectedTierDetails.description}
          </p>
          <div className="flex gap-4 text-xs text-muted-foreground pt-1">
            <span>${selectedTierDetails.priceMonthly}/month</span>
            <span>{selectedTierDetails.maxTeamMembers} max users</span>
            <span>{selectedTierDetails.maxDocuments} documents</span>
          </div>
        </div>
      )}

      {/* Tier comparison preview (top-level dialog only) */}
      {showTierComparison && selectedTierDetails && tierComparisonFeatures.length > 0 && (
        <div className="rounded-lg border p-3 space-y-2">
          <p className="text-xs font-medium text-muted-foreground uppercase tracking-wider">
            Features included
          </p>
          <ul className="grid grid-cols-2 gap-1 text-xs text-muted-foreground">
            {tierComparisonFeatures.map((f) => (
              <li key={f.id} className="flex items-center gap-1">
                <span className="text-green-500">✓</span>
                {f.displayName || f.name}
              </li>
            ))}
          </ul>
        </div>
      )}

      {/* Reason */}
      <div className="space-y-2">
        <label className="text-sm font-medium">
          Reason for Change <span className="text-red-500">*</span>
        </label>
        <Textarea
          value={reason}
          onChange={(e) => setReason(e.target.value)}
          placeholder="Explain why this tier change is needed (min 10 characters)..."
          rows={3}
        />
        <p className="text-xs text-muted-foreground">
          {reason.length}/10 minimum characters
        </p>
      </div>

      {/* Override limits checkbox */}
      {showOverrideLimits && (
        <div className="flex items-center space-x-2">
          <Checkbox
            id="changeTierOverrideLimits"
            checked={overrideLimits}
            onCheckedChange={(checked) => setOverrideLimits(checked as boolean)}
          />
          <label
            htmlFor="changeTierOverrideLimits"
            className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
          >
            Override existing limits with new tier defaults
          </label>
        </div>
      )}

      {error && (
        <div className="flex items-center gap-2 text-sm text-red-600 bg-red-50 dark:bg-red-950/20 p-2 rounded">
          <AlertTriangle className="h-4 w-4 shrink-0" />
          {error}
        </div>
      )}

      <Button
        type="submit"
        disabled={!selectedTier || isLoading}
        isLoading={isLoading}
        loadingText="Changing..."
        className={fullWidthSubmit ? "w-full" : undefined}
      >
        <ArrowUpDown className="mr-2 h-4 w-4" />
        Change Tier
      </Button>
    </form>
  );
}
