"use client";

import { useState, useEffect } from "react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { ArrowUpDown } from "lucide-react";
import { toast } from "sonner";
import {
  getAllSubscriptionTiers,
  type SubscriptionTier,
} from "@/app/_actions/subscriptions";
import { type Organization } from "@/app/_actions/organizations";
import { ChangeTierForm } from "./change-tier-form";

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
  const [isLoadingTiers, setIsLoadingTiers] = useState(false);

  useEffect(() => {
    if (open) {
      loadTiers();
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

        <ChangeTierForm
          organizationId={organization.id}
          currentTier={organization.subscription_tier}
          tiers={tiers}
          isLoadingTiers={isLoadingTiers}
          showOverrideLimits
          onSuccess={() => {
            onOpenChange(false);
            onSuccess();
          }}
        />
      </DialogContent>
    </Dialog>
  );
}
