"use client";

import { ReactNode, useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { AlertTriangle, Crown, Lock } from "lucide-react";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { useOrganizationContext } from "@/hooks/use-organization";
import {
  useOrganizationTrialStatus,
  useFeatureAccess,
} from "@/hooks/use-subscription-queries";
import { UpgradeModal } from "@/components/subscription/upgrade-modal";

interface SubscriptionGuardProps {
  children: ReactNode;
  feature?: string;
  requiredPlan?: "PRO_PLAN" | "ENTERPRISE";
  fallbackPath?: string;
  showUpgradeModal?: boolean;
  blockAccess?: boolean;
}

export function SubscriptionGuard({
  children,
  feature,
  requiredPlan = "PRO_PLAN",
  fallbackPath,
  showUpgradeModal = true,
  blockAccess = false,
}: SubscriptionGuardProps) {
  const router = useRouter();
  const { currentOrganization } = useOrganizationContext();
  const { data: trialStatus } = useOrganizationTrialStatus(
    currentOrganization?.id || "",
  );
  const { hasFeature, isLoading } = useFeatureAccess(
    currentOrganization?.id || "",
    feature || "",
  );
  const [showModal, setShowModal] = useState(false);

  useEffect(() => {
    // Auto-show upgrade modal if trial is expired and we're blocking access
    if (trialStatus?.isExpired && blockAccess && showUpgradeModal) {
      setShowModal(true);
    }
  }, [trialStatus?.isExpired, blockAccess, showUpgradeModal]);

  if (isLoading) {
    return (
      <div className="flex items-center justify-center p-8">
        <div className="animate-pulse space-y-4 w-full max-w-md">
          <div className="h-4 bg-gray-200 rounded w-3/4"></div>
          <div className="h-4 bg-gray-200 rounded w-1/2"></div>
        </div>
      </div>
    );
  }

  // Check if organization is in trial and expired
  const isTrialExpired =
    trialStatus?.isExpired && trialStatus?.subscriptionStatus === "trial";
  const inGracePeriod = trialStatus?.inGracePeriod;

  // Check feature access if specified
  const hasRequiredFeature = feature ? hasFeature : true;

  // Allow access if:
  // 1. No feature requirement OR has the required feature
  // 2. Not in expired trial OR in grace period (limited access)
  const allowAccess = hasRequiredFeature && (!isTrialExpired || inGracePeriod);

  if (allowAccess) {
    return <>{children}</>;
  }

  // If we should redirect instead of showing upgrade UI
  if (fallbackPath && !showUpgradeModal) {
    router.push(fallbackPath);
    return null;
  }

  // Show upgrade UI
  return (
    <>
      <div className="container mx-auto px-4 py-8">
        <Card className="max-w-2xl mx-auto">
          <CardHeader className="text-center">
            <div className="mx-auto p-3 rounded-full bg-orange-100 w-fit mb-4">
              <Lock className="h-8 w-8 text-orange-600" />
            </div>
            <CardTitle className="text-2xl">
              {isTrialExpired ? "Trial Expired" : "Premium Feature Required"}
            </CardTitle>
            <CardDescription>
              {isTrialExpired
                ? "Your trial has ended. Upgrade to continue using all features."
                : `This feature requires the ${requiredPlan === "PRO_PLAN" ? "Pro" : "Enterprise"} plan.`}
            </CardDescription>
          </CardHeader>

          <CardContent className="space-y-6">
            {isTrialExpired && inGracePeriod && (
              <Alert>
                <AlertTriangle className="h-4 w-4" />
                <AlertDescription>
                  You're in a 7-day grace period with limited access. Upgrade
                  now to restore full functionality.
                </AlertDescription>
              </Alert>
            )}

            <div className="text-center space-y-4">
              <p className="text-muted-foreground">
                {isTrialExpired
                  ? "Choose a plan to continue your procurement workflow management."
                  : `Upgrade to ${requiredPlan === "PRO_PLAN" ? "Pro" : "Enterprise"} to unlock this feature and more.`}
              </p>

              <div className="flex flex-col sm:flex-row gap-3 justify-center">
                <Button
                  onClick={() => setShowModal(true)}
                  size="lg"
                  className="flex-1 sm:flex-none"
                >
                  <Crown className="h-4 w-4 mr-2" />
                  View Plans & Upgrade
                </Button>

                {fallbackPath && (
                  <Button
                    variant="outline"
                    onClick={() => router.push(fallbackPath)}
                    size="lg"
                    className="flex-1 sm:flex-none"
                  >
                    Go Back
                  </Button>
                )}
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      <UpgradeModal
        isOpen={showModal}
        onClose={() => setShowModal(false)}
        currentTier={(trialStatus?.planSlug as any) || "STARTER_PLAN"}
      />
    </>
  );
}
