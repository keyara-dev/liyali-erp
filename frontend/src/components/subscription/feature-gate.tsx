"use client";

import { ReactNode, useState } from "react";
import { Crown, Lock, Zap, Building2 } from "lucide-react";
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
import { Alert, AlertDescription } from "@/components/ui/alert";
import { useOrganizationContext } from "@/hooks/use-organization";
import { useHasFeature } from "@/hooks/use-subscription-queries";
import { UpgradeModal } from "@/components/subscription/upgrade-modal";

interface FeatureGateProps {
  feature: string;
  children: ReactNode;
  fallback?: ReactNode;
  requiredPlan?: "PRO_PLAN" | "ENTERPRISE";
  showUpgrade?: boolean;
  className?: string;
}

export function FeatureGate({
  feature,
  children,
  fallback,
  requiredPlan = "PRO_PLAN",
  showUpgrade = true,
  className,
}: FeatureGateProps) {
  const { currentOrganization } = useOrganizationContext();
  const { hasFeature, isLoading } = useHasFeature(
    currentOrganization?.id || "",
    feature,
  );
  const [showUpgradeModal, setShowUpgradeModal] = useState(false);

  if (isLoading) {
    return (
      <div className={`animate-pulse ${className}`}>
        <div className="h-32 bg-muted rounded-lg"></div>
      </div>
    );
  }

  if (hasFeature) {
    return <>{children}</>;
  }

  if (fallback) {
    return <>{fallback}</>;
  }

  return (
    <>
      <FeatureLockedCard
        feature={feature}
        requiredPlan={requiredPlan}
        onUpgrade={showUpgrade ? () => setShowUpgradeModal(true) : undefined}
        className={className}
      />

      {showUpgrade && (
        <UpgradeModal
          isOpen={showUpgradeModal}
          onClose={() => setShowUpgradeModal(false)}
          currentTier="STARTER_PLAN"
        />
      )}
    </>
  );
}

interface FeatureLockedCardProps {
  feature: string;
  requiredPlan: "PRO_PLAN" | "ENTERPRISE";
  onUpgrade?: () => void;
  className?: string;
}

function FeatureLockedCard({
  feature,
  requiredPlan,
  onUpgrade,
  className,
}: FeatureLockedCardProps) {
  const planConfig = {
    PRO_PLAN: {
      name: "Pro",
      icon: Crown,
      color: "text-purple-600 dark:text-purple-400",
      bgColor: "bg-purple-50 dark:bg-purple-950/30",
      borderColor: "border-purple-200 dark:border-purple-800",
    },
    ENTERPRISE: {
      name: "Enterprise",
      icon: Building2,
      color: "text-emerald-600 dark:text-emerald-400",
      bgColor: "bg-emerald-50 dark:bg-emerald-950/30",
      borderColor: "border-emerald-200 dark:border-emerald-800",
    },
  };

  const config = planConfig[requiredPlan];
  const IconComponent = config.icon;

  const featureDisplayNames: Record<string, string> = {
    custom_roles: "Custom Role Management",
    offline_capabilities: "Offline Access",
    api_access: "API Access",
    priority_support: "Priority Support",
    advanced_analytics: "Advanced Analytics",
    dedicated_instance: "Dedicated Instance",
    sla_guarantees: "SLA Guarantees",
    custom_integrations: "Custom Integrations",
    models_modification: "Model Customization",
    unlimited_users: "Unlimited Users",
  };

  const featureName =
    featureDisplayNames[feature] ||
    feature.replace(/_/g, " ").replace(/\b\w/g, (l) => l.toUpperCase());

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.3 }}
      className={className}
    >
      <Card
        className={`${config.bgColor} ${config.borderColor} border-2 border-dashed`}
      >
        <CardHeader className="text-center pb-4">
          <div className="mx-auto p-3 rounded-full bg-background shadow-sm w-fit">
            <Lock className={`h-6 w-6 ${config.color}`} />
          </div>
          <CardTitle className="flex items-center justify-center gap-2">
            <span>{featureName}</span>
            <Badge
              variant="outline"
              className={`${config.color} border-current`}
            >
              <IconComponent className="h-3 w-3 mr-1" />
              {config.name}
            </Badge>
          </CardTitle>
          <CardDescription>
            This feature is available with the {config.name} plan
          </CardDescription>
        </CardHeader>

        <CardContent className="text-center space-y-4">
          <Alert className="text-left">
            <IconComponent className="h-4 w-4" />
            <AlertDescription>
              Upgrade to {config.name} to unlock {featureName.toLowerCase()} and
              other premium features.
            </AlertDescription>
          </Alert>

          {onUpgrade && (
            <Button onClick={onUpgrade} className="w-full">
              <Crown className="h-4 w-4 mr-2" />
              Upgrade to {config.name}
            </Button>
          )}
        </CardContent>
      </Card>
    </motion.div>
  );
}

// Utility component for inline feature checks
interface InlineFeatureGateProps {
  feature: string;
  children: ReactNode;
  fallback?: ReactNode;
  requiredPlan?: "PRO_PLAN" | "ENTERPRISE";
}

export function InlineFeatureGate({
  feature,
  children,
  fallback,
  requiredPlan = "PRO_PLAN",
}: InlineFeatureGateProps) {
  const { currentOrganization } = useOrganizationContext();
  const { hasFeature, isLoading } = useHasFeature(
    currentOrganization?.id || "",
    feature,
  );

  if (isLoading) {
    return null;
  }

  if (hasFeature) {
    return <>{children}</>;
  }

  if (fallback) {
    return <>{fallback}</>;
  }

  return null;
}

// Hook for conditional rendering based on features
export function useFeatureGate(feature: string) {
  const { currentOrganization } = useOrganizationContext();
  const { hasFeature, isLoading } = useHasFeature(
    currentOrganization?.id || "",
    feature,
  );

  return {
    hasFeature,
    isLoading,
    FeatureGate: ({
      children,
      fallback,
      ...props
    }: Omit<FeatureGateProps, "feature">) => (
      <FeatureGate feature={feature} fallback={fallback} {...props}>
        {children}
      </FeatureGate>
    ),
  };
}

// Component for showing feature availability in menus/lists
interface FeatureBadgeProps {
  feature: string;
  requiredPlan?: "PRO_PLAN" | "ENTERPRISE";
  className?: string;
}

export function FeatureBadge({
  feature,
  requiredPlan = "PRO_PLAN",
  className,
}: FeatureBadgeProps) {
  const { currentOrganization } = useOrganizationContext();
  const { hasFeature, isLoading } = useHasFeature(
    currentOrganization?.id || "",
    feature,
  );

  if (isLoading || hasFeature) {
    return null;
  }

  const planConfig = {
    PRO_PLAN: {
      name: "Pro",
      icon: Crown,
      color: "bg-purple-100 text-purple-700 dark:bg-purple-900 dark:text-purple-300",
    },
    ENTERPRISE: {
      name: "Enterprise",
      icon: Building2,
      color: "bg-emerald-100 text-emerald-700 dark:bg-emerald-900 dark:text-emerald-300",
    },
  };

  const config = planConfig[requiredPlan];
  const IconComponent = config.icon;

  return (
    <Badge variant="outline" className={`${config.color} ${className}`}>
      <IconComponent className="h-3 w-3 mr-1" />
      {config.name}
    </Badge>
  );
}
