"use client";

import { useState } from "react";
import { Clock, Crown, X, AlertTriangle } from "lucide-react";
import { motion, AnimatePresence } from "framer-motion";
import { Button } from "@/components/ui/button";
import { useOrganizationContext } from "@/hooks/use-organization";
import { useTrialCountdown } from "@/hooks/use-subscription-queries";
import { UpgradeModal } from "@/components/subscription/upgrade-modal";

interface TrialBottomBannerProps {
  className?: string;
}

export function TrialBottomBanner({ className }: TrialBottomBannerProps) {
  const { currentOrganization } = useOrganizationContext();
  const { trialInfo } = useTrialCountdown(currentOrganization?.id || "");
  const [showUpgradeModal, setShowUpgradeModal] = useState(false);
  const [isDismissed, setIsDismissed] = useState(false);

  if (!trialInfo || isDismissed || !currentOrganization) {
    return null;
  }

  const {
    status,
    message,
    urgency,
    daysRemaining,
    isTrialActive,
    isExpiringSoon,
    isExpired,
    inGracePeriod,
  } = trialInfo;

  // Only show for active trials or urgent situations
  if (status === "inactive" && !isExpired && !inGracePeriod) {
    return null;
  }

  // Don't show if trial has many days left (only show when <= 7 days or urgent)
  if (isTrialActive && daysRemaining > 7 && urgency === "low") {
    return null;
  }

  const getBackgroundStyle = () => {
    if (urgency === "high") {
      return "bg-gradient-to-r from-red-900/50 to-purple-900/50 border-red-500/30";
    }
    if (urgency === "medium") {
      return "bg-gradient-to-r from-amber-900/50 to-purple-900/50 border-amber-500/30";
    }
    return "bg-gradient-to-r from-blue-900/50 to-purple-900/50 border-blue-500/30";
  };

  const getIcon = () => {
    if (isExpired || inGracePeriod) {
      return <AlertTriangle className="h-5 w-5" />;
    }
    return <Clock className="h-5 w-5" />;
  };

  const getButtonText = () => {
    if (isExpired || inGracePeriod) {
      return "Upgrade Now";
    }
    if (isExpiringSoon) {
      return "Upgrade Before Expiry";
    }
    return "Upgrade to Pro";
  };

  return (
    <>
      <AnimatePresence>
        <motion.div
          initial={{ y: 100, opacity: 0 }}
          animate={{ y: 0, opacity: 1 }}
          exit={{ y: 100, opacity: 0 }}
          transition={{ duration: 0.3, type: "spring", stiffness: 300 }}
          className={`fixed bottom-0 left-0 right-0 z-50 opacity-70 ${className}`}
        >
          <div
            className={`${getBackgroundStyle()} backdrop-blur-md border-t shadow-2xl`}
          >
            <div className="max-w-7xl mx-auto px-4 py-3">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <motion.div
                    animate={{ rotate: [0, 10, -10, 0] }}
                    transition={{ duration: 2, repeat: Infinity }}
                    className="text-white"
                  >
                    {getIcon()}
                  </motion.div>
                  <div className="text-white">
                    <span className="font-semibold">
                      {isTrialActive && `${daysRemaining} days left in trial`}
                      {isExpired && "Trial expired"}
                      {inGracePeriod && "Grace period active"}
                    </span>
                    <span className="ml-2 opacity-95">
                      {isTrialActive && "- Upgrade to unlock all features"}
                      {(isExpired || inGracePeriod) &&
                        "- Upgrade to restore full access"}
                    </span>
                  </div>
                </div>

                <div className="flex items-center gap-3">
                  <motion.div
                    whileHover={{ scale: 1.05 }}
                    whileTap={{ scale: 0.95 }}
                  >
                    <Button
                      onClick={() => setShowUpgradeModal(true)}
                      className={`${
                        urgency === "high"
                          ? "bg-red-600 hover:bg-red-500 shadow-[0_0_20px_rgba(220,38,127,0.4)]"
                          : "bg-purple-600 hover:bg-purple-500 shadow-[0_0_20px_rgba(147,51,234,0.4)]"
                      } text-white font-semibold transition-all`}
                    >
                      <Crown className="h-4 w-4 mr-2" />
                      {getButtonText()}
                    </Button>
                  </motion.div>

                  <motion.div
                    whileHover={{ scale: 1.1 }}
                    whileTap={{ scale: 0.9 }}
                  >
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => setIsDismissed(true)}
                      className="text-white/70 hover:text-white hover:bg-white/10 h-8 w-8 p-0"
                    >
                      <X className="h-4 w-4" />
                    </Button>
                  </motion.div>
                </div>
              </div>
            </div>
          </div>
        </motion.div>
      </AnimatePresence>

      <UpgradeModal
        isOpen={showUpgradeModal}
        onClose={() => setShowUpgradeModal(false)}
        currentTier={currentOrganization.tier?.toUpperCase() || "STARTER"}
      />
    </>
  );
}
