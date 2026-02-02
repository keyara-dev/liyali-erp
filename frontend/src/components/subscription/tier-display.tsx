"use client";

import { useState } from "react";
import {
  Crown,
  Zap,
  Building2,
  ArrowRight,
  GemIcon,
  Gem,
  ChevronRightIcon,
} from "lucide-react";
import { motion } from "framer-motion";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";
import { useOrganizationContext } from "@/hooks/use-organization";
import { UpgradeModal } from "@/components/subscription/upgrade-modal";

const TIER_CONFIG = {
  STARTER: {
    label: "Starter",
    icon: Zap,
    color: "bg-blue-500/20 text-blue-400 border-blue-500/30",
    gradient: "from-blue-500 to-blue-600",
    description: "For growing teams",
    canUpgrade: true,
  },
  PRO: {
    label: "Pro",
    icon: Crown,
    color: "bg-purple-500/20 text-purple-400 border-purple-500/30",
    gradient: "from-purple-500 to-purple-600",
    description: "For established departments",
    canUpgrade: false,
  },
  ENTERPRISE: {
    label: "Enterprise",
    icon: Building2,
    color: "bg-emerald-500/20 text-emerald-400 border-emerald-500/30",
    gradient: "from-emerald-500 to-emerald-600",
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
        <motion.div whileHover={{ scale: 1.05 }} transition={{ duration: 0.2 }}>
          <Badge
            variant="outline"
            className={`${tierConfig.color} backdrop-blur-sm`}
          >
            <IconComponent className="h-3 w-3 mr-1" />
            {tierConfig.label}
          </Badge>
        </motion.div>
        {canUpgrade && (
          <motion.div
            whileHover={{ scale: 1.05 }}
            transition={{ duration: 0.2 }}
          >
            <Badge
              // size="sm"
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
              className="text-xs h-6 cursor-pointer flex gap-1  border-primary-300 text-primary "
            >
              <Gem className="h-3 w-3" />
              Upgrade
            </Badge>
          </motion.div>
        )}
        <UpgradeModal
          isOpen={showUpgradeModal}
          onClose={() => setShowUpgradeModal(false)}
          currentTier={tier}
        />
      </div>
    );
  }

  return (
    <>
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5 }}
        onClick={() => (canUpgrade ? setShowUpgradeModal(true) : undefined)}
        className={`cursor-pointer ${canUpgrade ? "hover:opacity-80" : ""}`}
      >
        <Card className="border border-primary/20 shadow-none px-0 backdrop-blur-md">
          <CardContent className="py-2 px-2">
            <div className="flex items-center  w-full justify-between">
              <div className="flex items-center">
                <div className="flex items-center gap-2">
                  <motion.div
                    className={`p-2 rounded-lg bg-linear-to-r ${tierConfig.gradient} shadow-lg`}
                    whileHover={{ scale: 1.1, rotate: 5 }}
                    transition={{ duration: 0.2 }}
                  >
                    <IconComponent className="h-5 w-5 text-white " />
                  </motion.div>
                  <div className="flex flex-col">
                    <span className="font-medium">{tierConfig.label} Plan</span>
                    <p className="text-xs truncate ">
                      {tierConfig.description}
                    </p>
                  </div>
                </div>
              </div>

              {canUpgrade && <ChevronRightIcon className="h-4 w-4 " />}
            </div>
          </CardContent>
        </Card>
      </motion.div>

      <UpgradeModal
        isOpen={showUpgradeModal}
        onClose={() => setShowUpgradeModal(false)}
        currentTier={tier}
      />
    </>
  );
}
