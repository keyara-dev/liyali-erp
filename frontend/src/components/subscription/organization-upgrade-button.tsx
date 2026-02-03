"use client";

import { useState } from "react";
import { Crown, Zap, Building2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { useOrganizationTrialStatus } from "@/hooks/use-subscription-queries";
import { UpgradeModal } from "@/components/subscription/upgrade-modal";

const TIER_CONFIG = {
  STARTER_PLAN: {
    label: "Starter",
    icon: Zap,
    color: "bg-blue-100 text-blue-700 border-blue-200",
  },
  PRO_PLAN: {
    label: "Pro",
    icon: Crown,
    color: "bg-purple-100 text-purple-700 border-purple-200",
  },
  ENTERPRISE: {
    label: "Enterprise",
    icon: Building2,
    color: "bg-emerald-100 text-emerald-700 border-emerald-200",
  },
} as const;

interface OrganizationUpgradeButtonProps {
  organizationId: string;
  organizationName: string;
  className?: string;
  variant?: "button" | "badge";
}

export function OrganizationUpgradeButton({
  organizationId,
  organizationName,
  className,
  variant = "button",
}: OrganizationUpgradeButtonProps) {
  const { data: trialStatus } = useOrganizationTrialStatus(organizationId);
  const [showUpgradeModal, setShowUpgradeModal] = useState(false);

  if (!trialStatus) {
    return null;
  }

  const planSlug = trialStatus.planSlug as keyof typeof TIER_CONFIG;
  const tierConfig = TIER_CONFIG[planSlug] || TIER_CONFIG.STARTER_PLAN;
  const IconComponent = tierConfig.icon;

  // Show upgrade button for starter plan or expired trials
  const shouldShowUpgrade =
    planSlug === "STARTER_PLAN" ||
    trialStatus.isExpired ||
    (trialStatus.daysRemaining <= 3 &&
      trialStatus.subscriptionStatus === "trial");

  if (!shouldShowUpgrade) {
    if (variant === "badge") {
      return (
        <Badge variant="outline" className={`${tierConfig.color} ${className}`}>
          <IconComponent className="h-3 w-3 mr-1" />
          {tierConfig.label}
        </Badge>
      );
    }
    return null;
  }

  if (variant === "badge") {
    return (
      <>
        <Button
          variant="outline"
          size="sm"
          onClick={() => setShowUpgradeModal(true)}
          className={`${className} text-xs px-2 py-1 h-6`}
        >
          Upgrade
        </Button>

        <UpgradeModal
          isOpen={showUpgradeModal}
          onClose={() => setShowUpgradeModal(false)}
          currentTier={planSlug}
        />
      </>
    );
  }

  return (
    <>
      <Button
        variant="outline"
        size="sm"
        onClick={() => setShowUpgradeModal(true)}
        className={className}
      >
        <Crown className="h-4 w-4 mr-2" />
        Upgrade
      </Button>

      <UpgradeModal
        isOpen={showUpgradeModal}
        onClose={() => setShowUpgradeModal(false)}
        currentTier={planSlug}
      />
    </>
  );
}
