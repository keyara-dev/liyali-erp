"use client";

import { useState } from "react";
import { AlertTriangle, Crown, X } from "lucide-react";
import { motion, AnimatePresence } from "framer-motion";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { useOrganizationContext } from "@/hooks/use-organization";
import { useTrialCountdown } from "@/hooks/use-subscription-queries";
import { UpgradeModal } from "@/components/subscription/upgrade-modal";

interface TrialExpiryBannerProps {
  className?: string;
}

export function TrialExpiryBanner({ className }: TrialExpiryBannerProps) {
  const { currentOrganization } = useOrganizationContext();
  const { trialInfo } = useTrialCountdown(currentOrganization?.id || "");
  const [showUpgradeModal, setShowUpgradeModal] = useState(false);
  const [isDismissed, setIsDismissed] = useState(false);

  const isPaidTier = currentOrganization?.tier === "pro" || currentOrganization?.tier === "custom";

  if (!trialInfo || isDismissed || isPaidTier) {
    return null;
  }

  const {
    status,
    message,
    urgency,
    daysRemaining,
    isExpiringSoon,
    isExpired,
    inGracePeriod,
  } = trialInfo;

  // Only show banner for expiring, expired, or grace period trials
  if (!isExpiringSoon && !isExpired && !inGracePeriod) {
    return null;
  }

  const getBannerVariant = () => {
    if (urgency === "high") return "destructive";
    return "default";
  };

  const getBannerMessage = () => {
    if (isExpired && inGracePeriod) {
      return "Your trial has expired. You're in a 7-day grace period with limited access.";
    }
    if (isExpired) {
      return "Your trial has expired. Upgrade now to continue using all features.";
    }
    if (isExpiringSoon) {
      return `Your trial expires in ${daysRemaining} day${daysRemaining !== 1 ? "s" : ""}. Upgrade to avoid interruption.`;
    }
    return message;
  };

  const getBannerStyle = () => {
    if (urgency === "high") {
      return "bg-red-900/20 border-red-500/30 backdrop-blur-md";
    }
    return "bg-yellow-900/20 border-yellow-500/30 backdrop-blur-md";
  };

  const getTextColor = () => {
    if (urgency === "high") return "text-red-400";
    return "text-yellow-400";
  };

  return (
    <>
      <AnimatePresence>
        <motion.div
          initial={{ opacity: 0, y: -20, scale: 0.95 }}
          animate={{ opacity: 1, y: 0, scale: 1 }}
          exit={{ opacity: 0, y: -20, scale: 0.95 }}
          transition={{ duration: 0.3 }}
          className={className}
        >
          {/* Dark Background with Floating Elements */}
          <div className="relative overflow-hidden">
            <motion.div
              className="absolute top-0 right-0 w-[100px] h-[100px] bg-red-600/5 rounded-full blur-[40px]"
              animate={{ scale: [1, 1.2, 1], opacity: [0.3, 0.6, 0.3] }}
              transition={{ duration: 4, repeat: Infinity }}
            />

            <Alert className={`relative ${getBannerStyle()} border`}>
              <AlertTriangle className={`h-4 w-4 ${getTextColor()}`} />
              <AlertDescription
                className={`flex items-center justify-between pr-8 ${getTextColor()}`}
              >
                <span className="font-medium">{getBannerMessage()}</span>
                <div className="flex items-center gap-2">
                  <motion.div
                    whileHover={{ scale: 1.05 }}
                    whileTap={{ scale: 0.95 }}
                  >
                    <Button
                      size="sm"
                      onClick={() => setShowUpgradeModal(true)}
                      className="shrink-0 bg-gradient-to-r from-purple-600 to-blue-600 hover:from-purple-700 hover:to-blue-700 shadow-[0_0_15px_rgba(147,51,234,0.3)]"
                    >
                      <Crown className="h-4 w-4 mr-2" />
                      Upgrade Now
                    </Button>
                  </motion.div>
                </div>
              </AlertDescription>
              <motion.div whileHover={{ scale: 1.1 }} whileTap={{ scale: 0.9 }}>
                <Button
                  variant="ghost"
                  size="sm"
                  className={`absolute top-2 right-2 h-6 w-6 p-0 ${getTextColor()} hover:bg-slate-700/50`}
                  onClick={() => setIsDismissed(true)}
                >
                  <X className="h-4 w-4" />
                </Button>
              </motion.div>
            </Alert>
          </div>
        </motion.div>
      </AnimatePresence>

      <UpgradeModal
        isOpen={showUpgradeModal}
        onClose={() => setShowUpgradeModal(false)}
        currentTier="STARTER_PLAN"
      />
    </>
  );
}
