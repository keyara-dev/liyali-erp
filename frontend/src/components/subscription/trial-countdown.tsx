"use client";

import { useState, useEffect } from "react";
import { Clock, AlertTriangle, Crown, X } from "lucide-react";
import { motion, AnimatePresence } from "framer-motion";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Progress } from "@/components/ui/progress";
import { useOrganizationContext } from "@/hooks/use-organization";
import { useTrialCountdown } from "@/hooks/use-subscription-queries";
import { UpgradeModal } from "@/components/subscription/upgrade-modal";
import { Spinner } from "@/components/ui/spinner";

interface TrialCountdownProps {
  className?: string;
  compact?: boolean;
  dismissible?: boolean;
}

export function TrialCountdown({
  className,
  compact = false,
  dismissible = false,
}: TrialCountdownProps) {
  const { currentOrganization } = useOrganizationContext();
  const { trialInfo, isLoading } = useTrialCountdown(
    currentOrganization?.id || "",
  );
  const [showUpgradeModal, setShowUpgradeModal] = useState(false);
  const [isDismissed, setIsDismissed] = useState(false);

  // Auto-dismiss after showing for a while (for non-critical states)
  useEffect(() => {
    if (trialInfo?.urgency === "low" && dismissible) {
      const timer = setTimeout(() => {
        setIsDismissed(true);
      }, 10000); // Auto-dismiss after 10 seconds for low urgency

      return () => clearTimeout(timer);
    }
  }, [trialInfo?.urgency, dismissible]);

  if (!currentOrganization || isLoading || !trialInfo || isDismissed) {
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

  // Don't show for inactive trials unless they're expired/in grace period
  if (status === "inactive") {
    return null;
  }

  const getBackgroundStyle = () => {
    if (urgency === "high") {
      return "bg-red-50 border-red-200 dark:bg-red-900/30 dark:border-red-400/50 backdrop-blur-md";
    }
    if (urgency === "medium") {
      return "bg-yellow-50 border-yellow-200 dark:bg-yellow-900/30 dark:border-yellow-400/50 backdrop-blur-md";
    }
    return "bg-blue-50 border-blue-200 dark:bg-blue-900/30 dark:border-blue-400/50 backdrop-blur-md";
  };

  const getTextColor = () => {
    if (urgency === "high") {
      return "text-red-800 dark:text-red-300";
    }
    if (urgency === "medium") {
      return "text-yellow-800 dark:text-yellow-300";
    }
    return "text-blue-800 dark:text-blue-300";
  };

  const getIconBgColor = () => {
    if (urgency === "high") {
      return "bg-red-100 dark:bg-red-500/30";
    }
    if (urgency === "medium") {
      return "bg-yellow-100 dark:bg-yellow-500/30";
    }
    return "bg-blue-100 dark:bg-blue-500/30";
  };

  const getIcon = () => {
    switch (status) {
      case "expiring":
      case "expired":
      case "grace":
        return <AlertTriangle className="h-4 w-4" />;
      default:
        return <Clock className="h-4 w-4" />;
    }
  };

  const trialProgress = isTrialActive ? ((14 - daysRemaining) / 14) * 100 : 100;

  if (compact) {
    return (
      <>
        <AnimatePresence>
          <motion.div
            initial={{ opacity: 0, y: -20 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -20 }}
            transition={{ duration: 0.3 }}
            className={`${getBackgroundStyle()} ${getTextColor()} p-3 rounded-lg border ${className} relative overflow-hidden`}
          >
            {/* Floating Background Element */}
            <motion.div
              className="absolute top-0 right-0 w-[60px] h-[60px] bg-blue-600/5 rounded-full blur-[30px]"
              animate={{ scale: [1, 1.3, 1], opacity: [0.3, 0.6, 0.3] }}
              transition={{ duration: 3, repeat: Infinity }}
            />

            <div className="flex items-center justify-between relative z-10">
              <div className="flex items-center gap-2">
                <motion.div
                  animate={{ rotate: [0, 10, -10, 0] }}
                  transition={{ duration: 2, repeat: Infinity }}
                >
                  {getIcon()}
                </motion.div>
                <span className="text-sm font-medium">{message}</span>
              </div>
              <div className="flex items-center gap-2">
                {(isExpiringSoon || isExpired || inGracePeriod) && (
                  <motion.div
                    whileHover={{ scale: 1.05 }}
                    whileTap={{ scale: 0.95 }}
                  >
                    <Button
                      size="sm"
                      onClick={() => setShowUpgradeModal(true)}
                      className="h-7 px-3 text-xs bg-gradient-to-r from-purple-600 to-blue-600 hover:from-purple-700 hover:to-blue-700 shadow-[0_0_10px_rgba(147,51,234,0.3)]"
                    >
                      Upgrade
                    </Button>
                  </motion.div>
                )}
                {dismissible && (
                  <motion.div
                    whileHover={{ scale: 1.1 }}
                    whileTap={{ scale: 0.9 }}
                  >
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => setIsDismissed(true)}
                      className="h-7 w-7 p-0 hover:bg-slate-200/50 text-slate-600 hover:text-slate-800 dark:hover:bg-slate-700/50 dark:text-slate-300 dark:hover:text-white"
                    >
                      <X className="h-3 w-3" />
                    </Button>
                  </motion.div>
                )}
              </div>
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

  return (
    <>
      <AnimatePresence>
        <motion.div
          initial={{ opacity: 0, scale: 0.95 }}
          animate={{ opacity: 1, scale: 1 }}
          exit={{ opacity: 0, scale: 0.95 }}
          transition={{ duration: 0.3 }}
        >
          <Card
            className={`${getBackgroundStyle()} border ${className} relative overflow-hidden`}
          >
            {/* Dark Background with Floating Elements */}
            <motion.div
              className="absolute top-[20%] right-[10%] w-[120px] h-[120px] bg-blue-600/5 rounded-full blur-[60px]"
              animate={{ scale: [1, 1.3, 1], opacity: [0.3, 0.6, 0.3] }}
              transition={{ duration: 6, repeat: Infinity }}
            />
            <motion.div
              className="absolute bottom-[10%] left-[10%] w-[80px] h-[80px] bg-purple-600/5 rounded-full blur-[40px]"
              animate={{ scale: [1.2, 1, 1.2], opacity: [0.2, 0.5, 0.2] }}
              transition={{ duration: 8, repeat: Infinity, delay: 2 }}
            />

            {/* Floating Math Operator */}
            <motion.div
              className="absolute top-[15%] left-[5%] text-2xl font-black text-blue-500/5 blur-[1px]"
              animate={{ y: [0, -5, 0], rotate: [0, 5, 0] }}
              transition={{ duration: 4, repeat: Infinity, ease: "easeInOut" }}
            >
              !
            </motion.div>

            <CardContent className="p-6 relative z-10">
              <div className="space-y-4">
                {/* Header */}
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <motion.div
                      className={`p-2 rounded-full ${getIconBgColor()} backdrop-blur-sm`}
                      whileHover={{ scale: 1.1, rotate: 10 }}
                      transition={{ duration: 0.2 }}
                    >
                      {getIcon()}
                    </motion.div>
                    <div>
                      <h3 className={`font-semibold ${getTextColor()}`}>
                        {status === "active" && "Trial Active"}
                        {status === "expiring" && "Trial Expiring Soon"}
                        {status === "expired" && "Trial Expired"}
                        {status === "grace" && "Grace Period"}
                      </h3>
                      <p className={`text-sm ${getTextColor()} opacity-80`}>
                        {message}
                      </p>
                    </div>
                  </div>

                  {dismissible && urgency === "low" && (
                    <motion.div
                      whileHover={{ scale: 1.1 }}
                      whileTap={{ scale: 0.9 }}
                    >
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => setIsDismissed(true)}
                        className="text-slate-600 hover:text-slate-800 hover:bg-slate-200/50 dark:text-slate-300 dark:hover:text-white dark:hover:bg-slate-700/50"
                      >
                        <X className="h-4 w-4" />
                      </Button>
                    </motion.div>
                  )}
                </div>

                {/* Progress Bar (for active trials) */}
                {isTrialActive && (
                  <div className="space-y-2">
                    <div className="flex justify-between text-xs opacity-80">
                      <span className="text-slate-600 dark:text-slate-300">
                        Trial Progress
                      </span>
                      <span className={getTextColor()}>
                        {daysRemaining} days left
                      </span>
                    </div>
                    <motion.div
                      initial={{ width: 0 }}
                      animate={{ width: "100%" }}
                      transition={{ duration: 1, delay: 0.5 }}
                    >
                      <Progress
                        value={trialProgress}
                        className="h-2 bg-slate-200 dark:bg-slate-700"
                      />
                    </motion.div>
                  </div>
                )}

                {/* Action Buttons */}
                <div className="flex gap-3">
                  <motion.div
                    className="flex-1"
                    whileHover={{ scale: 1.02 }}
                    whileTap={{ scale: 0.98 }}
                  >
                    <Button
                      onClick={() => setShowUpgradeModal(true)}
                      className={`w-full ${
                        urgency === "high"
                          ? "bg-gradient-to-r from-red-600 to-purple-600 hover:from-red-700 hover:to-purple-700 shadow-[0_0_20px_rgba(220,38,127,0.3)]"
                          : "bg-gradient-to-r from-purple-600 to-blue-600 hover:from-purple-700 hover:to-blue-700 shadow-[0_0_20px_rgba(147,51,234,0.3)]"
                      } transition-all`}
                    >
                      <Crown className="h-4 w-4 mr-2" />
                      Upgrade to Pro
                    </Button>
                  </motion.div>

                  {status === "active" && urgency === "low" && (
                    <motion.div
                      whileHover={{ scale: 1.02 }}
                      whileTap={{ scale: 0.98 }}
                    >
                      <Button
                        variant="ghost"
                        onClick={() => setIsDismissed(true)}
                        className="px-4 text-slate-600 hover:text-slate-800 hover:bg-slate-200/50 dark:text-slate-300 dark:hover:text-white dark:hover:bg-slate-700/50"
                      >
                        Remind Later
                      </Button>
                    </motion.div>
                  )}
                </div>

                {/* Additional Info */}
                {(isExpired || inGracePeriod) && (
                  <motion.div
                    initial={{ opacity: 0 }}
                    animate={{ opacity: 1 }}
                    transition={{ delay: 0.5 }}
                    className={`text-xs ${getTextColor()} opacity-80 text-center bg-slate-100/50 dark:bg-slate-800/30 p-3 rounded-lg backdrop-blur-sm`}
                  >
                    {inGracePeriod
                      ? "You have limited access during the grace period. Upgrade to restore full functionality."
                      : "Your trial has ended. Upgrade now to continue using all features."}
                  </motion.div>
                )}
              </div>
            </CardContent>
          </Card>
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
