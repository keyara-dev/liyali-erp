"use client";

import { useState } from "react";
import {
  Crown,
  Zap,
  Building2,
  Clock,
  AlertTriangle,
  CheckCircle,
  XCircle,
} from "lucide-react";
import { motion } from "framer-motion";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Progress } from "@/components/ui/progress";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Separator } from "@/components/ui/separator";
import { Spinner } from "@/components/ui/spinner";
import { useOrganizationContext } from "@/hooks/use-organization";
import {
  useOrganizationSubscription,
  useTrialCountdown,
} from "@/hooks/use-subscription-queries";
import { UpgradeModal } from "@/components/subscription/upgrade-modal";

const TIER_CONFIG = {
  STARTER_PLAN: {
    label: "Starter",
    icon: Zap,
    color: "bg-blue-500/20 text-blue-400 border-blue-500/30",
    gradient: "from-blue-500 to-blue-600",
  },
  PRO_PLAN: {
    label: "Pro",
    icon: Crown,
    color: "bg-purple-500/20 text-purple-400 border-purple-500/30",
    gradient: "from-purple-500 to-purple-600",
  },
  ENTERPRISE: {
    label: "Enterprise",
    icon: Building2,
    color: "bg-emerald-500/20 text-emerald-400 border-emerald-500/30",
    gradient: "from-emerald-500 to-emerald-600",
  },
} as const;

interface SubscriptionManagerProps {
  className?: string;
}

export function SubscriptionManager({ className }: SubscriptionManagerProps) {
  const { currentOrganization } = useOrganizationContext();
  const [showUpgradeModal, setShowUpgradeModal] = useState(false);

  const { data: subscription, isLoading: subscriptionLoading } =
    useOrganizationSubscription(currentOrganization?.id || "");

  const { trialInfo, isLoading: trialLoading } = useTrialCountdown(
    currentOrganization?.id || "",
  );

  if (!currentOrganization) {
    return null;
  }

  if (subscriptionLoading || trialLoading) {
    return (
      <Card
        className={`${className} bg-slate-800/60 border-slate-700 backdrop-blur-md`}
      >
        <CardContent className="p-6">
          <div className="flex items-center justify-center">
            <Spinner className="h-8 w-8 text-blue-400" />
          </div>
        </CardContent>
      </Card>
    );
  }

  const currentPlan = subscription?.plan;
  const planSlug = currentPlan?.slug as keyof typeof TIER_CONFIG;
  const tierConfig = planSlug
    ? TIER_CONFIG[planSlug]
    : TIER_CONFIG.STARTER_PLAN;
  const IconComponent = tierConfig.icon;

  return (
    <>
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5 }}
      >
        <Card
          className={`${className} bg-slate-800/60 border-slate-700 backdrop-blur-md relative overflow-hidden`}
        >
          {/* Dark Background with Floating Elements */}
          <motion.div
            className="absolute top-[10%] right-[10%] w-[150px] h-[150px] bg-blue-600/5 rounded-full blur-[80px]"
            animate={{ scale: [1, 1.2, 1], opacity: [0.3, 0.6, 0.3] }}
            transition={{ duration: 8, repeat: Infinity }}
          />
          <motion.div
            className="absolute bottom-[10%] left-[10%] w-[100px] h-[100px] bg-purple-600/5 rounded-full blur-[60px]"
            animate={{ scale: [1.2, 1, 1.2], opacity: [0.2, 0.5, 0.2] }}
            transition={{ duration: 10, repeat: Infinity, delay: 2 }}
          />

          {/* Floating Math Operators */}
          <motion.div
            className="absolute top-[20%] left-[5%] text-3xl font-black text-blue-500/5 blur-[1px]"
            animate={{ y: [0, -10, 0], rotate: [0, 5, 0] }}
            transition={{ duration: 6, repeat: Infinity, ease: "easeInOut" }}
          >
            $
          </motion.div>

          <CardHeader className="relative z-10">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                <motion.div
                  className={`p-2 rounded-lg bg-gradient-to-r ${tierConfig.gradient} shadow-lg`}
                  whileHover={{ scale: 1.1, rotate: 5 }}
                  transition={{ duration: 0.2 }}
                >
                  <IconComponent className="h-5 w-5 text-white" />
                </motion.div>
                <div>
                  <CardTitle className="text-lg text-white">
                    {currentPlan?.name || "Starter Plan"}
                  </CardTitle>
                  <CardDescription className="text-slate-400">
                    {currentPlan?.description ||
                      "Getting started with procurement workflows"}
                  </CardDescription>
                </div>
              </div>
              <Badge
                variant="outline"
                className={`${tierConfig.color} backdrop-blur-sm`}
              >
                {tierConfig.label}
              </Badge>
            </div>
          </CardHeader>

          <CardContent className="space-y-6 relative z-10">
            {/* Trial Status */}
            {trialInfo && (
              <TrialStatusSection
                trialInfo={trialInfo}
                onUpgrade={() => setShowUpgradeModal(true)}
              />
            )}

            {/* Plan Features */}
            <div>
              <h4 className="font-medium mb-3 text-white">Plan Features</h4>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                {currentPlan?.features.map((feature: any, index: number) => (
                  <motion.div
                    key={index}
                    className="flex items-center gap-3 text-sm text-slate-300"
                    initial={{ opacity: 0, x: -20 }}
                    animate={{ opacity: 1, x: 0 }}
                    transition={{ delay: index * 0.1 }}
                  >
                    <div className="w-5 h-5 rounded-full flex items-center justify-center shrink-0 border border-green-500 bg-green-500/20 text-green-400">
                      <CheckCircle className="h-3 w-3" />
                    </div>
                    <span>{feature}</span>
                  </motion.div>
                ))}
              </div>
            </div>

            {/* Usage Limits */}
            {subscription?.planLimits && (
              <UsageLimitsSection planLimits={subscription.planLimits} />
            )}

            {/* Actions */}
            <div className="flex gap-3">
              {planSlug === "STARTER_PLAN" && (
                <motion.div
                  className="flex-1"
                  whileHover={{ scale: 1.02 }}
                  whileTap={{ scale: 0.98 }}
                >
                  <Button
                    onClick={() => setShowUpgradeModal(true)}
                    className="w-full bg-linear-to-r from-purple-600 to-blue-600 hover:from-purple-700 hover:to-blue-700 shadow-[0_0_20px_rgba(147,51,234,0.3)]"
                  >
                    <Crown className="h-4 w-4 mr-2" />
                    Upgrade to Pro
                  </Button>
                </motion.div>
              )}

              {planSlug === "PRO_PLAN" && (
                <motion.div
                  className="flex-1"
                  whileHover={{ scale: 1.02 }}
                  whileTap={{ scale: 0.98 }}
                >
                  <Button
                    variant="outline"
                    onClick={() => setShowUpgradeModal(true)}
                    className="w-full bg-transparent border-slate-600 text-slate-300 hover:bg-slate-700 hover:border-slate-500"
                  >
                    <Building2 className="h-4 w-4 mr-2" />
                    Contact for Enterprise
                  </Button>
                </motion.div>
              )}
            </div>
          </CardContent>
        </Card>
      </motion.div>

      <UpgradeModal
        isOpen={showUpgradeModal}
        onClose={() => setShowUpgradeModal(false)}
        currentTier={planSlug || "STARTER_PLAN"}
      />
    </>
  );
}

interface TrialStatusSectionProps {
  trialInfo: NonNullable<ReturnType<typeof useTrialCountdown>["trialInfo"]>;
  onUpgrade: () => void;
}

function TrialStatusSection({ trialInfo, onUpgrade }: TrialStatusSectionProps) {
  const { status, message, urgency, daysRemaining, isTrialActive } = trialInfo;

  const getAlertStyle = (): string => {
    switch (urgency) {
      case "high":
        return "bg-red-900/20 border-red-500/30 backdrop-blur-md";
      default:
        return "bg-yellow-900/20 border-yellow-500/30 backdrop-blur-md";
    }
  };

  const getTextColor = (): string => {
    switch (urgency) {
      case "high":
        return "text-red-400";
      default:
        return "text-yellow-400";
    }
  };

  const getIcon = () => {
    switch (status) {
      case "active":
        return <Clock className="h-4 w-4" />;
      case "expiring":
        return <AlertTriangle className="h-4 w-4" />;
      case "expired":
        return <XCircle className="h-4 w-4" />;
      case "grace":
        return <AlertTriangle className="h-4 w-4" />;
      default:
        return <Clock className="h-4 w-4" />;
    }
  };

  const trialProgress = isTrialActive ? ((14 - daysRemaining) / 14) * 100 : 100;

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h4 className="font-medium text-white">Trial Status</h4>
        {isTrialActive && (
          <span className="text-sm text-slate-400">
            {daysRemaining} of 14 days remaining
          </span>
        )}
      </div>

      {isTrialActive && (
        <div className="space-y-2">
          <motion.div
            initial={{ width: 0 }}
            animate={{ width: "100%" }}
            transition={{ duration: 1, delay: 0.3 }}
          >
            <Progress value={trialProgress} className="h-2 bg-slate-700" />
          </motion.div>
          <div className="flex justify-between text-xs text-slate-400">
            <span>Trial started</span>
            <span>Trial ends</span>
          </div>
        </div>
      )}

      <Alert className={getAlertStyle()}>
        <motion.div
          animate={{ rotate: [0, 10, -10, 0] }}
          transition={{ duration: 2, repeat: Infinity }}
        >
          {getIcon()}
        </motion.div>
        <AlertDescription
          className={`flex items-center justify-between ${getTextColor()}`}
        >
          <span className="font-medium">{message}</span>
          {(status === "expiring" ||
            status === "expired" ||
            status === "grace") && (
            <motion.div whileHover={{ scale: 1.05 }} whileTap={{ scale: 0.95 }}>
              <Button
                size="sm"
                onClick={onUpgrade}
                className={`${
                  urgency === "high"
                    ? "bg-gradient-to-r from-red-600 to-purple-600 hover:from-red-700 hover:to-purple-700 shadow-[0_0_15px_rgba(220,38,127,0.3)]"
                    : "bg-gradient-to-r from-purple-600 to-blue-600 hover:from-purple-700 hover:to-blue-700 shadow-[0_0_15px_rgba(147,51,234,0.3)]"
                } transition-all`}
              >
                Upgrade Now
              </Button>
            </motion.div>
          )}
        </AlertDescription>
      </Alert>
    </div>
  );
}

interface UsageLimitsSectionProps {
  planLimits: {
    organizationId: string;
    maxUsersAllowed: number;
    planMaxUsers: number;
    planMetadata: Record<string, any>;
    currentUserCount: number;
    canAddUsers: boolean;
  };
}

function UsageLimitsSection({ planLimits }: UsageLimitsSectionProps) {
  const { currentUserCount, maxUsersAllowed, canAddUsers } = planLimits;
  const isUnlimited = maxUsersAllowed === -1;
  const usagePercentage = isUnlimited
    ? 0
    : (currentUserCount / maxUsersAllowed) * 100;

  return (
    <div className="space-y-4">
      <Separator className="bg-slate-700" />
      <div>
        <h4 className="font-medium mb-3 text-white">Usage & Limits</h4>

        <div className="space-y-3">
          {/* User Limit */}
          <div className="flex items-center justify-between text-sm">
            <span className="text-slate-400">Users</span>
            <span className="font-medium text-white">
              {currentUserCount} / {isUnlimited ? "Unlimited" : maxUsersAllowed}
            </span>
          </div>

          {!isUnlimited && (
            <div className="space-y-2">
              <motion.div
                initial={{ width: 0 }}
                animate={{ width: "100%" }}
                transition={{ duration: 1, delay: 0.5 }}
              >
                <Progress
                  value={usagePercentage}
                  className="h-2 bg-slate-700"
                />
              </motion.div>
              {!canAddUsers && (
                <Alert className="bg-yellow-900/20 border-yellow-500/30 backdrop-blur-md">
                  <AlertTriangle className="h-4 w-4 text-yellow-400" />
                  <AlertDescription className="text-yellow-400">
                    You've reached your user limit. Upgrade to add more users.
                  </AlertDescription>
                </Alert>
              )}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
