"use client";

import { useState } from "react";
import {
  Crown,
  Check,
  Mail,
  Zap,
  Building2,
  X,
  Plus,
  Lock,
} from "lucide-react";
import { motion } from "framer-motion";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";
import { Spinner } from "@/components/ui/spinner";
import { useOrganizationContext } from "@/hooks/use-organization";
import {
  useSubscriptionPlans,
} from "@/hooks/use-subscription-queries";

const TIER_ICONS = {
  STARTER_PLAN: Zap,
  PRO_PLAN: Crown,
  ENTERPRISE: Building2,
};

const TIER_COLORS = {
  STARTER_PLAN: "from-blue-500 to-blue-600",
  PRO_PLAN: "from-purple-500 to-purple-600",
  ENTERPRISE: "from-emerald-500 to-emerald-600",
};

interface UpgradeModalProps {
  isOpen: boolean;
  onClose: () => void;
  currentTier: string;
}

export function UpgradeModal({
  isOpen,
  onClose,
  currentTier,
}: UpgradeModalProps) {
  const { currentOrganization } = useOrganizationContext();
  const [selectedPlan, setSelectedPlan] = useState<string>("");
  const [billingCycle, setBillingCycle] = useState<"monthly" | "yearly">(
    "monthly",
  );
  const [step, setStep] = useState<"plans" | "payment">("plans");

  // Use TanStack Query hook for fetching plans
  const {
    data: plans = [],
    isLoading: plansLoading,
    error: plansQueryError,
  } = useSubscriptionPlans();

  const plansError = plansQueryError?.message || null;

  // Show all plans and highlight the current one
  const allPlans = plans || [];

  // Determine current plan slug from tier
  const getCurrentPlanSlug = (tier: string) => {
    const normalized = tier.toUpperCase();
    if (normalized.includes("STARTER")) return "STARTER_PLAN";
    if (normalized.includes("PRO")) return "PRO_PLAN";
    if (normalized.includes("ENTERPRISE")) return "ENTERPRISE";
    return "STARTER_PLAN"; // Default fallback
  };

  const currentPlanSlug = getCurrentPlanSlug(currentTier);

  const handlePlanSelect = (planSlug: string) => {
    // Don't allow selecting the current plan
    if (planSlug === currentPlanSlug) {
      return;
    }

    setSelectedPlan(planSlug);
    if (planSlug === "ENTERPRISE") {
      // Enterprise requires contact
      handleContactSales();
    } else {
      setStep("payment");
    }
  };

  const handleContactSales = () => {
    // TODO: Implement contact sales flow
    window.open(
      "mailto:sales@liyali.com?subject=Enterprise Plan Inquiry",
      "_blank",
    );
    onClose();
  };

  const handleUpgrade = async () => {
    if (!currentOrganization || !selectedPlan) return;

    // Payment processor not yet integrated — redirect to contact sales
    // instead of attempting a tier change without charging.
    window.open(
      `mailto:sales@liyali.com?subject=Upgrade Request - ${selectedPlan}&body=Organization: ${currentOrganization.name}%0APlan: ${selectedPlan}%0ABilling: ${billingCycle}`,
      "_blank",
    );
    onClose();
    setStep("plans");
    setSelectedPlan("");
  };

  const selectedPlanData = plans?.find((p: any) => p.slug === selectedPlan);

  const yearlyDiscount =
    selectedPlanData && selectedPlanData.priceMonthly > 0
      ? Math.round(
          (1 -
            selectedPlanData.priceYearly / 12 / selectedPlanData.priceMonthly) *
            100,
        )
      : 0;

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="max-w-6xl! max-h-[90vh] overflow-y-auto bg-slate-900/98 border-slate-600 backdrop-blur-md">
        {/* Dark Blue Theme Background with Floating Elements */}
        <div className="absolute inset-0 overflow-hidden pointer-events-none">
          <motion.div
            className="absolute top-[10%] right-[10%] w-[200px] h-[200px] bg-blue-600/10 rounded-full blur-[80px]"
            animate={{ scale: [1, 1.2, 1], opacity: [0.3, 0.6, 0.3] }}
            transition={{ duration: 8, repeat: Infinity }}
          />
          <motion.div
            className="absolute bottom-[10%] left-[10%] w-[150px] h-[150px] bg-purple-600/10 rounded-full blur-[60px]"
            animate={{ scale: [1.2, 1, 1.2], opacity: [0.2, 0.5, 0.2] }}
            transition={{ duration: 10, repeat: Infinity, delay: 2 }}
          />

          {/* Floating Math Operators */}
          <motion.div
            className="absolute top-[20%] left-[5%] text-4xl font-black text-blue-500/5 blur-[1px]"
            animate={{ y: [0, -10, 0], rotate: [0, 5, 0] }}
            transition={{ duration: 6, repeat: Infinity, ease: "easeInOut" }}
          >
            <Plus className="" />
          </motion.div>
          <motion.div
            className="absolute bottom-[30%] right-[5%] text-5xl font-black text-purple-400/5 blur-[1px]"
            animate={{ y: [0, 10, 0], rotate: [0, -10, 0] }}
            transition={{
              duration: 8,
              repeat: Infinity,
              ease: "easeInOut",
              delay: 1,
            }}
          >
            <X className="" />
          </motion.div>
        </div>

        <DialogHeader className="relative z-10">
          <DialogTitle className="text-2xl text-white font-semibold">
            {step === "plans" ? "Choose Your Plan" : "Complete Your Upgrade"}
          </DialogTitle>
          <DialogDescription className="text-slate-300">
            {step === "plans"
              ? "Compare all available plans and upgrade to unlock more features"
              : "Review your selection and complete the upgrade"}
          </DialogDescription>
        </DialogHeader>

        <div className="relative z-10">
          {step === "plans" && (
            <PlansStep
              plans={allPlans}
              currentTier={currentTier}
              currentPlanSlug={currentPlanSlug}
              onSelectPlan={handlePlanSelect}
              isLoading={plansLoading}
              error={plansError}
            />
          )}

          {step === "payment" && selectedPlanData && (
            <PaymentStep
              plan={selectedPlanData}
              billingCycle={billingCycle}
              onBillingCycleChange={setBillingCycle}
              onUpgrade={handleUpgrade}
              onBack={() => setStep("plans")}
              isLoading={upgradeMutation.isPending}
              yearlyDiscount={yearlyDiscount}
            />
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
}

interface PlansStepProps {
  plans: any[];
  currentTier: string;
  currentPlanSlug: string;
  onSelectPlan: (planSlug: string) => void;
  isLoading: boolean;
  error?: string | null;
}

function PlansStep({
  plans,
  currentTier,
  currentPlanSlug,
  onSelectPlan,
  isLoading,
  error,
}: PlansStepProps) {
  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-12">
        <Spinner className="h-8 w-8 text-blue-400" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex flex-col items-center justify-center py-12 text-center">
        <div className="text-red-400 mb-4">
          <Crown className="h-12 w-12 mx-auto mb-2" />
          <h3 className="text-lg font-semibold text-white mb-2">
            Error Loading Plans
          </h3>
          <p className="text-sm text-slate-300">{error}</p>
        </div>
      </div>
    );
  }

  if (!plans || plans.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-12 text-center">
        <div className="text-slate-400 mb-4">
          <Crown className="h-12 w-12 mx-auto mb-2" />
          <h3 className="text-lg font-semibold text-white mb-2">
            No Plans Available
          </h3>
          <p className="text-sm text-slate-300">
            Unable to load subscription plans. Please contact support.
          </p>
        </div>
        <div className="text-xs text-slate-400 mt-4">
          Current tier: {currentTier} | Plans loaded: {plans?.length || 0}
        </div>
      </div>
    );
  }

  // Sort plans by sortOrder and compute unique features per tier
  const sortedPlans = [...plans].sort(
    (a: any, b: any) => a.sortOrder - b.sortOrder,
  );

  // Build feature display name lookup
  const featureDisplayNames: Record<string, { displayName: string; description: string }> = {};
  for (const plan of sortedPlans) {
    for (const detail of plan.featureDetails || []) {
      featureDisplayNames[detail.name] = {
        displayName: detail.displayName,
        description: detail.description,
      };
    }
  }

  // Compute unique features and inherited tier for each plan
  let previousFeatures = new Set<string>();
  const plansWithUniqueFeatures = sortedPlans.map((plan: any, idx: number) => {
    const currentFeatures = new Set<string>(plan.features || []);
    const uniqueFeatures = (plan.features || []).filter(
      (f: string) => !previousFeatures.has(f),
    );
    const inheritedTierName =
      previousFeatures.size > 0 ? sortedPlans[idx - 1]?.displayName : undefined;

    const uniqueFeatureDetails = uniqueFeatures.slice(0, 10).map((f: string) => ({
      name: f,
      displayName:
        featureDisplayNames[f]?.displayName ||
        f.replace(/_/g, " ").replace(/\b\w/g, (c: string) => c.toUpperCase()),
      description: featureDisplayNames[f]?.description || "",
    }));

    previousFeatures = currentFeatures;

    return { ...plan, uniqueFeatureDetails, inheritedTierName };
  });

  return (
    <div className="space-y-4">
      <div className="text-center mb-6">
        <p className="text-slate-300 text-sm">
          You&apos;re currently on the{" "}
          <span className="font-semibold text-white">{currentTier}</span> plan
        </p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-6 py-6">
        {plansWithUniqueFeatures.map((plan) => (
          <PlanCard
            key={plan.id}
            plan={plan}
            isCurrentPlan={plan.slug === currentPlanSlug}
            onSelect={() => onSelectPlan(plan.slug)}
          />
        ))}
      </div>
    </div>
  );
}

interface PlanCardProps {
  plan: any;
  isCurrentPlan: boolean;
  onSelect: () => void;
}

function PlanCard({
  plan,
  isCurrentPlan,
  onSelect,
}: PlanCardProps) {
  const IconComponent =
    TIER_ICONS[plan.slug as keyof typeof TIER_ICONS] || Crown;
  const gradient =
    TIER_COLORS[plan.slug as keyof typeof TIER_COLORS] || TIER_COLORS.PRO_PLAN;
  const isRecommended = plan.slug === "PRO_PLAN";
  const isEnterprise = plan.slug === "ENTERPRISE";

  return (
    <motion.div
      whileHover={!isCurrentPlan ? { y: -4, scale: 1.02 } : {}}
      transition={{ duration: 0.2 }}
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
    >
      <Card
        className={`relative transition-all duration-300 backdrop-blur-md ${
          isCurrentPlan
            ? "bg-blue-900/50 border-blue-400 ring-2 ring-blue-400/60 cursor-default"
            : isRecommended
              ? "cursor-pointer bg-slate-800/70 border-slate-500 hover:shadow-2xl ring-2 ring-purple-400/60 scale-105 hover:bg-slate-800/90 hover:border-slate-400"
              : "cursor-pointer bg-slate-800/70 border-slate-500 hover:shadow-2xl hover:bg-slate-800/90 hover:border-slate-400"
        }`}
      >
        {isCurrentPlan && (
          <motion.div
            className="absolute -top-4 left-1/2 transform -translate-x-1/2"
            initial={{ scale: 0 }}
            animate={{ scale: 1 }}
            transition={{ delay: 0.3, type: "spring", stiffness: 500 }}
          >
            <Badge className="bg-blue-600/90 text-white shadow-lg shadow-blue-500/30 border border-blue-400/50">
              Current Plan
            </Badge>
          </motion.div>
        )}

        {!isCurrentPlan && isRecommended && (
          <motion.div
            className="absolute -top-4 left-1/2 transform -translate-x-1/2"
            initial={{ scale: 0 }}
            animate={{ scale: 1 }}
            transition={{ delay: 0.3, type: "spring", stiffness: 500 }}
          >
            <Badge className="bg-purple-600/90 text-white shadow-lg shadow-purple-500/30 border border-purple-400/50">
              Most Popular
            </Badge>
          </motion.div>
        )}

        <CardHeader className="text-center pb-4">
          <motion.div
            className={`mx-auto p-3 rounded-full bg-gradient-to-r ${gradient} w-fit shadow-lg`}
            whileHover={{ scale: 1.1, rotate: 5 }}
            transition={{ duration: 0.2 }}
          >
            <IconComponent className="h-6 w-6 text-white" />
          </motion.div>
          <CardTitle className="text-xl text-white font-semibold">
            {plan.displayName}
          </CardTitle>
          <CardDescription className="text-slate-300">
            {plan.description}
          </CardDescription>

          <div className="pt-4">
            {isEnterprise ? (
              <div className="text-2xl font-bold text-white">
                Custom Pricing
              </div>
            ) : (
              <div className="space-y-1">
                <div className="text-3xl font-bold text-white">
                  ${plan.priceMonthly}
                  <span className="text-lg font-normal text-slate-400">
                    /month
                  </span>
                </div>
                {plan.priceMonthly > 0 && plan.priceYearly > 0 && (
                  <div className="text-sm text-slate-400">
                    or ${plan.priceYearly}/year (save{" "}
                    {Math.round(
                      (1 - plan.priceYearly / 12 / plan.priceMonthly) * 100,
                    )}
                    %)
                  </div>
                )}
              </div>
            )}
          </div>
        </CardHeader>

        <CardContent className="space-y-4">
          {plan.inheritedTierName ? (
            <div className="flex items-center gap-2 px-3 py-2 rounded-lg bg-blue-500/10 border border-blue-500/20 mb-1">
              <Zap className="h-3.5 w-3.5 text-blue-400" />
              <span className="text-sm font-medium text-blue-300">
                Everything in {plan.inheritedTierName}, plus:
              </span>
            </div>
          ) : (
            <p className="text-sm font-semibold text-slate-300 mb-1">Includes:</p>
          )}
          <div className="space-y-3">
            {(plan.uniqueFeatureDetails || []).map((feature: any, index: number) => (
              <motion.div
                key={feature.name}
                className="flex items-start gap-3 text-sm"
                initial={{ opacity: 0, x: -20 }}
                animate={{ opacity: 1, x: 0 }}
                transition={{ delay: index * 0.05 }}
              >
                <div className="w-5 h-5 rounded-full flex items-center justify-center shrink-0 border border-green-400 bg-green-500/30 text-green-300 mt-0.5">
                  <Check className="h-3 w-3" />
                </div>
                <div className="flex-1">
                  <div className="font-medium text-slate-200">
                    {feature.displayName}
                  </div>
                  {feature.description && (
                    <div className="text-xs text-slate-400 mt-0.5">
                      {feature.description}
                    </div>
                  )}
                </div>
              </motion.div>
            ))}
          </div>

          <motion.div
            whileHover={!isCurrentPlan ? { scale: 1.02 } : {}}
            whileTap={!isCurrentPlan ? { scale: 0.98 } : {}}
          >
            <Button
              onClick={isCurrentPlan ? undefined : onSelect}
              disabled={isCurrentPlan}
              className={`w-full transition-all ${
                isCurrentPlan
                  ? "bg-blue-600/60 border border-blue-400 text-blue-100 cursor-not-allowed"
                  : isRecommended
                    ? "bg-purple-600 hover:bg-purple-500 text-white shadow-[0_0_20px_rgba(147,51,234,0.3)]"
                    : "bg-transparent border border-slate-500 text-white hover:bg-slate-700 hover:border-slate-400"
              }`}
            >
              {isCurrentPlan ? (
                <>
                  <Check className="h-4 w-4 mr-2" />
                  Current Plan
                </>
              ) : isEnterprise ? (
                <>
                  <Mail className="h-4 w-4 mr-2" />
                  Contact Sales
                </>
              ) : (
                <>
                  <Crown className="h-4 w-4 mr-2" />
                  Upgrade to {plan.displayName}
                </>
              )}
            </Button>
          </motion.div>
        </CardContent>
      </Card>
    </motion.div>
  );
}

interface PaymentStepProps {
  plan: any;
  billingCycle: "monthly" | "yearly";
  onBillingCycleChange: (cycle: "monthly" | "yearly") => void;
  onUpgrade: () => void;
  onBack: () => void;
  isLoading: boolean;
  yearlyDiscount: number;
}

function PaymentStep({
  plan,
  billingCycle,
  onBillingCycleChange,
  onUpgrade,
  onBack,
  isLoading,
  yearlyDiscount,
}: PaymentStepProps) {
  const price =
    billingCycle === "monthly" ? plan.priceMonthly : plan.priceYearly;
  const totalPrice = billingCycle === "monthly" ? price : price;

  return (
    <div className="space-y-6 py-6">
      {/* Plan Summary */}
      <Card className="bg-slate-800/70 border-slate-500">
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-white">
            <Crown className="h-5 w-5" />
            {plan.name}
          </CardTitle>
          <CardDescription className="text-slate-300">
            {plan.description}
          </CardDescription>
        </CardHeader>
      </Card>

      {/* Billing Cycle Selection */}
      <div className="space-y-4">
        <h3 className="font-medium text-white">Billing Cycle</h3>
        <RadioGroup
          value={billingCycle}
          onValueChange={(value) =>
            onBillingCycleChange(value as "monthly" | "yearly")
          }
        >
          <div className="flex items-center space-x-2 p-4 border border-slate-500 rounded-lg bg-slate-800/50">
            <RadioGroupItem value="monthly" id="monthly" />
            <Label
              htmlFor="monthly"
              className="flex-1 cursor-pointer text-slate-200"
            >
              <div className="flex justify-between w-full items-center">
                <div>
                  <div className="font-medium text-white">Monthly</div>
                  <div className="text-sm text-slate-400">Pay monthly</div>
                </div>
                <div className="font-bold text-white text-xl">
                  ${plan.priceMonthly}/month
                </div>
              </div>
            </Label>
          </div>

          <div className="flex items-center space-x-2 p-4 border border-slate-500 rounded-lg bg-slate-800/50">
            <RadioGroupItem value="yearly" id="yearly" />
            <Label
              htmlFor="yearly"
              className="flex-1 cursor-pointer text-slate-200"
            >
              <div className="flex justify-between w-full items-center">
                <div>
                  <div className="font-medium flex items-center gap-2 text-white">
                    Yearly
                    {yearlyDiscount > 0 && (
                      <Badge
                        variant="secondary"
                        className="bg-green-600/20 text-green-300 border-green-500/30"
                      >
                        Save {yearlyDiscount}%
                      </Badge>
                    )}
                  </div>
                  <div className="text-sm text-slate-400">Pay annually</div>
                </div>
                <div className="text-xl">
                  <div className="font-bold text-white">
                    ${plan.priceYearly}/year
                  </div>
                  <div className="text-base text-slate-400">
                    ${Math.round(plan.priceYearly / 12)}/month
                  </div>
                </div>
              </div>
            </Label>
          </div>
        </RadioGroup>
      </div>

      <Separator />

      {/* Order Summary */}
      <div className="space-y-4">
        <h3 className="font-medium text-white">Order Summary</h3>
        <div className="space-y-2">
          <div className="flex justify-between text-slate-200">
            <span>
              {plan.name} ({billingCycle})
            </span>
            <span className="text-white">${totalPrice}</span>
          </div>
          <Separator className="bg-slate-600" />
          <div className="flex justify-between font-bold text-2xl text-white">
            <span>Total</span>
            <span className="">${totalPrice}</span>
          </div>
        </div>
      </div>

      {/* Actions */}
      <div className="flex gap-3">
        <Button
          variant="secondary"
          onClick={onBack}
          className="flex-1 border-slate-500 text-slate-200 bg-primary"
        >
          Back
        </Button>
        <Button
          onClick={onUpgrade}
          disabled={isLoading}
          isLoading={isLoading}
          loadingText="Processing..."
          className="flex-1 bg-purple-600 hover:bg-purple-500 text-white"
        >
          <Mail className="h-4 w-4 mr-2" />
          Contact Sales to Upgrade
        </Button>
      </div>

      {/* Note */}
      <div className="text-xs text-slate-400 text-center flex items-center justify-center">
        <Lock className="h-4 w-4 mr-2" />
        Our sales team will reach out to complete your upgrade and set up billing.
      </div>
    </div>
  );
}
