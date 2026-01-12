"use client";

import { useState } from "react";
import { Crown, Zap, Building2, ArrowRight, GemIcon } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";
import { useOrganizationContext } from "@/hooks/use-organization";
import { UpgradeModal } from "@/components/modals/upgrade-modal";

const TIER_CONFIG = {
  STARTER: {
    label: "Starter",
    icon: Zap,
    color: "bg-blue-100 text-blue-700 border-blue-200",
    description: "For growing teams",
    canUpgrade: true,
  },
  PRO: {
    label: "Pro",
    icon: Crown,
    color: "bg-purple-100 text-purple-700 border-purple-200",
    description: "For established departments",
    canUpgrade: false,
  },
  ENTERPRISE: {
    label: "Enterprise",
    icon: Building2,
    color: "bg-emerald-100 text-emerald-700 border-emerald-200",
    description: "For large organizations",
    canUpgrade: false,
  },
} as const;

interface TierDisplayProps {
  showUpgradeButton?: boolean;
  compact?: boolean;
}

export function TierDisplay({
  showUpgradeButton = true,
  compact = false,
}: TierDisplayProps) {
  const { currentOrganization } = useOrganizationContext();
  const [showUpgradeModal, setShowUpgradeModal] = useState(false);

  if (!currentOrganization) {
    return null;
  }

  const tier = (currentOrganization.tier?.toUpperCase() ||
    "STARTER") as keyof typeof TIER_CONFIG;
  const tierConfig = TIER_CONFIG[tier] || TIER_CONFIG.STARTER;
  const IconComponent = tierConfig.icon || GemIcon;
  const canUpgrade = tier === "STARTER" && showUpgradeButton;

  if (compact) {
    return (
      <div
        className="flex items-center gap-2"
        onClick={(e) => e.stopPropagation()}
        onMouseDown={(e) => e.stopPropagation()}
      >
        <Badge variant="outline" className={tierConfig.color}>
          <IconComponent className="h-3 w-3 mr-1" />
          {tierConfig.label}
        </Badge>
        {canUpgrade && (
          <Button
            size="sm"
            variant="outline"
            data-upgrade-button="true"
            onClick={(e) => {
              e.preventDefault();
              e.stopPropagation();
              setShowUpgradeModal(true);
            }}
            onMouseDown={(e) => {
              e.preventDefault();
              e.stopPropagation();
            }}
            className="text-xs h-6 p-2"
          >
            Upgrade
          </Button>
        )}
        <UpgradeModal
          open={showUpgradeModal}
          onOpenChange={(open) => {
            setShowUpgradeModal(open);
            // Prevent any potential event bubbling when modal closes
            if (!open) {
              setTimeout(() => {
                // Small delay to ensure modal is fully closed before any other events
              }, 10);
            }
          }}
          currentTier={tier}
          organizationName={currentOrganization.name}
        />
      </div>
    );
  }

  return (
    <>
      <Card className="border-l-4 border-l-primary/20">
        <CardContent className="p-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="flex items-center gap-2">
                <IconComponent className="h-5 w-5 text-muted-foreground" />
                <div>
                  <div className="flex items-center gap-2">
                    <span className="font-medium">{tierConfig.label} Plan</span>
                    <Badge variant="outline" className={tierConfig.color}>
                      Active
                    </Badge>
                  </div>
                  <p className="text-sm text-muted-foreground">
                    {tierConfig.description}
                  </p>
                </div>
              </div>
            </div>

            {canUpgrade && (
              <Button
                onClick={() => setShowUpgradeModal(true)}
                className="bg-linear-to-r from-blue-600 to-purple-600 hover:from-blue-700 hover:to-purple-700"
              >
                Upgrade now
                <ArrowRight className="h-4 w-4 ml-2" />
              </Button>
            )}
          </div>
        </CardContent>
      </Card>

      <UpgradeModal
        open={showUpgradeModal}
        onOpenChange={setShowUpgradeModal}
        currentTier={tier}
        organizationName={currentOrganization.name}
      />
    </>
  );
}
