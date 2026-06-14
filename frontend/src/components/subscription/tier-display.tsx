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
  Sparkles,
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
  const canUpgrade = tier !== "ENTERPRISE" && showUpgradeButton;

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

  const showUpgradeCta = canUpgrade;

  const ctaCopy =
    tier === "PRO"
      ? {
          eyebrow: "Go Enterprise",
          title: "Scale without limits",
          subtitle: "SSO, dedicated support & unlimited everything.",
        }
      : {
          eyebrow: "Go Pro",
          title: "Unlock the full workspace",
          subtitle: "Unlimited workflows, more seats & advanced approvals.",
        };

  return (
    <>
      {showUpgradeCta ? (
        <motion.div
          initial={{ opacity: 0, y: 12 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5, ease: "easeOut" }}
          className="group relative overflow-hidden rounded-xl border border-primary-500/25 bg-linear-to-br from-primary-500/12 via-primary-500/5 to-transparent p-3 shadow-sm"
        >
          {/* Ambient glow — brand blue + a gold spark, both very low opacity */}
          <motion.div
            aria-hidden
            className="pointer-events-none absolute -right-6 -top-8 h-24 w-24 rounded-full bg-primary-500/30 blur-2xl"
            animate={{ scale: [1, 1.25, 1], opacity: [0.35, 0.65, 0.35] }}
            transition={{ duration: 6, repeat: Infinity, ease: "easeInOut" }}
          />
          <motion.div
            aria-hidden
            className="pointer-events-none absolute -bottom-10 -left-8 h-24 w-24 rounded-full bg-amber-400/20 blur-2xl"
            animate={{ scale: [1.2, 1, 1.2], opacity: [0.2, 0.4, 0.2] }}
            transition={{
              duration: 8,
              repeat: Infinity,
              ease: "easeInOut",
              delay: 1.5,
            }}
          />

          <div className="relative z-1 flex items-start gap-2.5">
            <motion.div
              className="relative flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-linear-to-br from-primary-500 to-primary-700 shadow-lg shadow-primary-600/30"
              whileHover={{ scale: 1.08, rotate: 4 }}
              transition={{ duration: 0.2 }}
            >
              <Gem className="h-4 w-4 text-white" />
              <motion.span
                aria-hidden
                className="absolute -right-1 -top-1 text-amber-300"
                animate={{
                  scale: [1, 1.3, 1],
                  opacity: [0.6, 1, 0.6],
                  rotate: [0, 15, 0],
                }}
                transition={{ duration: 2.5, repeat: Infinity, ease: "easeInOut" }}
              >
                <Sparkles className="h-3 w-3" />
              </motion.span>
            </motion.div>

            <div className="min-w-0">
              <p className="text-[0.65rem] font-semibold uppercase tracking-[0.14em] text-amber-500 dark:text-amber-300">
                {ctaCopy.eyebrow}
              </p>
              <p className="truncate text-sm font-semibold leading-tight text-sidebar-foreground">
                {ctaCopy.title}
              </p>
            </div>
          </div>

          <p className="relative z-1 mt-2 text-xs leading-snug text-muted-foreground">
            {ctaCopy.subtitle}
          </p>

          <Button
            type="button"
            onClick={() => setShowUpgradeModal(true)}
            className="relative z-1 mt-3 w-full"
          >
            <Gem className="size-4" />
            Upgrade plan
          </Button>
        </motion.div>
      ) : (
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5 }}
          onClick={() => (canUpgrade ? setShowUpgradeModal(true) : undefined)}
          className={`${canUpgrade ? "cursor-pointer hover:opacity-80" : "cursor-default"}`}
        >
          <Card className="border border-border hover:bg-primary-50/10 shadow-none px-0 backdrop-blur-md">
            <CardContent className="p-1 px-1">
              <div className="flex items-center  w-full justify-between">
                <div className="flex items-center">
                  <div className="flex items-center gap-2">
                    <motion.div
                      className={`p-1.5 rounded-lg bg-linear-to-r ${tierConfig.gradient} shadow-lg`}
                      whileHover={{ scale: 1.1, rotate: 5 }}
                      transition={{ duration: 0.2 }}
                    >
                      <IconComponent className="h-4 w-4 text-white " />
                    </motion.div>
                    <div className="flex flex-col">
                      <span className="font-medium">
                        {tierConfig.label} Plan
                      </span>
                    </div>
                  </div>
                </div>

                {canUpgrade && <ChevronRightIcon className="h-4 w-4 " />}
              </div>
            </CardContent>
          </Card>
        </motion.div>
      )}

      <UpgradeModal
        isOpen={showUpgradeModal}
        onClose={() => setShowUpgradeModal(false)}
        currentTier={tier}
      />
    </>
  );
}
