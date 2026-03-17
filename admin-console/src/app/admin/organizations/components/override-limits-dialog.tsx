"use client";

import { useState, useEffect } from "react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Shield } from "lucide-react";
import { toast } from "sonner";
import {
  getAllSubscriptionFeatures,
  type SubscriptionFeature,
} from "@/app/_actions/subscriptions";
import { type Organization } from "@/app/_actions/organizations";
import { OverrideLimitsForm } from "./override-limits-form";

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
  const [isLoadingFeatures, setIsLoadingFeatures] = useState(false);

  useEffect(() => {
    if (open) {
      loadFeatures();
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

        <OverrideLimitsForm
          organizationId={organization.id}
          currentSettings={{
            max_users: organization.settings?.max_users,
            features_enabled: organization.settings?.features_enabled,
            subscription_tier: organization.subscription_tier,
          }}
          availableFeatures={availableFeatures}
          isLoadingFeatures={isLoadingFeatures}
          onSuccess={() => {
            onOpenChange(false);
            onSuccess();
          }}
        />
      </DialogContent>
    </Dialog>
  );
}
