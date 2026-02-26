"use client";

import { ReactNode, useState } from "react";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { useResourceLimit } from "@/hooks/use-usage-queries";
import { UpgradeModal } from "@/components/subscription/upgrade-modal";
import { useOrganizationContext } from "@/hooks/use-organization";

type ResourceType =
  | "workspace"
  | "team_member"
  | "document"
  | "workflow"
  | "custom_role"
  | "requisition"
  | "budget"
  | "purchase_order"
  | "payment_voucher"
  | "grn"
  | "department"
  | "vendor";

interface LimitGateProps {
  resource: ResourceType;
  children: ReactNode;
}

/**
 * Wraps a create button/action. When the resource limit is reached,
 * disables the child and shows a tooltip with upgrade prompt.
 */
export function LimitGate({ resource, children }: LimitGateProps) {
  const { usage, limit, isAtLimit, isUnlimited, displayName, isLoading } =
    useResourceLimit(resource);
  const { currentOrganization } = useOrganizationContext();
  const [showUpgrade, setShowUpgrade] = useState(false);

  // While loading or if unlimited, render children as-is
  if (isLoading || isUnlimited) {
    return <>{children}</>;
  }

  if (!isAtLimit) {
    return <>{children}</>;
  }

  return (
    <>
      <TooltipProvider>
        <Tooltip>
          <TooltipTrigger asChild>
            <span
              className="inline-flex cursor-not-allowed"
              onClick={(e) => {
                e.preventDefault();
                e.stopPropagation();
                setShowUpgrade(true);
              }}
            >
              <span className="pointer-events-none opacity-50">{children}</span>
            </span>
          </TooltipTrigger>
          <TooltipContent side="top" className="max-w-xs text-center">
            <p className="font-medium">
              {displayName} limit reached ({usage}/{limit})
            </p>
            <p className="text-xs text-muted-foreground mt-1">
              <button
                onClick={() => setShowUpgrade(true)}
                className="text-primary underline underline-offset-2"
              >
                Upgrade your plan
              </button>{" "}
              to create more.
            </p>
          </TooltipContent>
        </Tooltip>
      </TooltipProvider>

      <UpgradeModal
        isOpen={showUpgrade}
        onClose={() => setShowUpgrade(false)}
        currentTier={currentOrganization?.tier || "starter"}
      />
    </>
  );
}
