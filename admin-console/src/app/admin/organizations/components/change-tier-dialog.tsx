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
import { Textarea } from "@/components/ui/textarea";
import { Checkbox } from "@/components/ui/checkbox";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { ArrowUpDown, AlertTriangle, Loader2 } from "lucide-react";
import { toast } from "sonner";
import {
  getAllSubscriptionTiers,
  changeOrganizationTier,
  type SubscriptionTier,
} from "@/app/_actions/subscriptions";
import { type Organization } from "@/app/_actions/organizations";

interface ChangeTierDialogProps {
  organization: Organization;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess: () => void;
}

export function ChangeTierDialog({
  organization,
  open,
  onOpenChange,
  onSuccess,
}: ChangeTierDialogProps) {
  const [tiers, setTiers] = useState<SubscriptionTier[]>([]);
  const [selectedTier, setSelectedTier] = useState("");
  const [reason, setReason] = useState("");
  const [overrideLimits, setOverrideLimits] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [isLoadingTiers, setIsLoadingTiers] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    if (open) {
      loadTiers();
      setSelectedTier("");
      setReason("");
      setOverrideLimits(false);
      setError("");
    }
  }, [open]);

  const loadTiers = async () => {
    setIsLoadingTiers(true);
    try {
      const result = await getAllSubscriptionTiers();
      if (result.success && result.data) {
        setTiers(result.data.filter((t) => t.isActive));
      }
    } catch {
      toast.error("Failed to load subscription tiers");
    } finally {
      setIsLoadingTiers(false);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");

    if (!selectedTier) {
      setError("Please select a tier");
      return;
    }

    if (selectedTier === organization.subscription_tier) {
      setError("Please select a different tier than the current one");
      return;
    }

    if (reason.trim().length < 10) {
      setError("Reason must be at least 10 characters");
      return;
    }

    setIsLoading(true);
    try {
      const result = await changeOrganizationTier(organization.id, {
        newTier: selectedTier,
        reason: reason.trim(),
        overrideLimits: overrideLimits,
      });

      if (result.success) {
        toast.success("Organization tier changed successfully");
        onOpenChange(false);
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

  const getTierBadge = (tier: string) => {
    switch (tier) {
      case "enterprise":
        return <Badge variant="default">Enterprise</Badge>;
      case "professional":
        return (
          <Badge className="bg-blue-100 text-blue-800">Professional</Badge>
        );
      case "basic":
        return <Badge variant="secondary">Basic</Badge>;
      default:
        return <Badge variant="outline">{tier}</Badge>;
    }
  };

  const selectedTierDetails = tiers.find((t) => t.name === selectedTier);

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <ArrowUpDown className="h-5 w-5" />
            Change Subscription Tier
          </DialogTitle>
          <DialogDescription>
            Change the subscription tier for{" "}
            <strong>{organization.name}</strong>
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="rounded-lg border p-3 bg-muted/50">
            <div className="flex items-center justify-between">
              <span className="text-sm font-medium">Current Tier:</span>
              {getTierBadge(organization.subscription_tier)}
            </div>
          </div>

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
                      disabled={tier.name === organization.subscription_tier}
                    >
                      <div className="flex items-center gap-2">
                        <span>{tier.displayName || tier.name}</span>
                        <span className="text-xs text-muted-foreground">
                          ${tier.priceMonthly}/mo &middot;{" "}
                          {tier.maxTeamMembers} users
                        </span>
                      </div>
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            )}
          </div>

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

          <div className="flex items-center space-x-2">
            <Checkbox
              id="overrideLimits"
              checked={overrideLimits}
              onCheckedChange={(checked) =>
                setOverrideLimits(checked as boolean)
              }
            />
            <label
              htmlFor="overrideLimits"
              className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
            >
              Override existing limits with new tier defaults
            </label>
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
            <Button type="submit" disabled={!selectedTier} isLoading={isLoading} loadingText="Changing...">
              Change Tier
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
