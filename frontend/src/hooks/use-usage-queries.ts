"use client";

import { useQuery } from "@tanstack/react-query";
import { getOrganizationUsage } from "@/app/_actions/subscriptions";

export interface EffectiveLimits {
  organizationId: string;
  tierName: string;
  maxWorkspaces: number;
  maxTeamMembers: number;
  maxDocuments: number;
  maxWorkflows: number;
  maxCustomRoles: number;
  maxRequisitions: number;
  maxBudgets: number;
  maxPurchaseOrders: number;
  maxPaymentVouchers: number;
  maxGRNs: number;
  maxDepartments: number;
  maxVendors: number;
  hasOverrides: boolean;
}

export interface OrganizationUsage {
  organizationId: string;
  currentWorkspaces: number;
  currentTeamMembers: number;
  currentDocuments: number;
  currentWorkflows: number;
  currentCustomRoles: number;
  currentRequisitions: number;
  currentBudgets: number;
  currentPurchaseOrders: number;
  currentPaymentVouchers: number;
  currentGRNs: number;
  currentDepartments: number;
  currentVendors: number;
  workspacesPercent: number;
  teamMembersPercent: number;
  documentsPercent: number;
  workflowsPercent: number;
  customRolesPercent: number;
  requisitionsPercent: number;
  budgetsPercent: number;
  purchaseOrdersPercent: number;
  paymentVouchersPercent: number;
  grnsPercent: number;
  departmentsPercent: number;
  vendorsPercent: number;
}

export interface LimitsWithUsage {
  limits: EffectiveLimits;
  usage: OrganizationUsage;
}

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

const RESOURCE_LIMIT_MAP: Record<ResourceType, { limitKey: keyof EffectiveLimits; usageKey: keyof OrganizationUsage }> = {
  workspace: { limitKey: "maxWorkspaces", usageKey: "currentWorkspaces" },
  team_member: { limitKey: "maxTeamMembers", usageKey: "currentTeamMembers" },
  document: { limitKey: "maxDocuments", usageKey: "currentDocuments" },
  workflow: { limitKey: "maxWorkflows", usageKey: "currentWorkflows" },
  custom_role: { limitKey: "maxCustomRoles", usageKey: "currentCustomRoles" },
  requisition: { limitKey: "maxRequisitions", usageKey: "currentRequisitions" },
  budget: { limitKey: "maxBudgets", usageKey: "currentBudgets" },
  purchase_order: { limitKey: "maxPurchaseOrders", usageKey: "currentPurchaseOrders" },
  payment_voucher: { limitKey: "maxPaymentVouchers", usageKey: "currentPaymentVouchers" },
  grn: { limitKey: "maxGRNs", usageKey: "currentGRNs" },
  department: { limitKey: "maxDepartments", usageKey: "currentDepartments" },
  vendor: { limitKey: "maxVendors", usageKey: "currentVendors" },
};

const RESOURCE_DISPLAY_NAMES: Record<ResourceType, string> = {
  workspace: "workspace",
  team_member: "team member",
  document: "document",
  workflow: "workflow",
  custom_role: "custom role",
  requisition: "requisition",
  budget: "budget",
  purchase_order: "purchase order",
  payment_voucher: "payment voucher",
  grn: "goods received note",
  department: "department",
  vendor: "vendor",
};

/**
 * Hook to fetch organization usage and limits
 */
export function useOrganizationUsage() {
  return useQuery<LimitsWithUsage>({
    queryKey: ["organization-usage"],
    queryFn: async () => {
      const response = await getOrganizationUsage();
      if (!response.success) {
        throw new Error(response.message);
      }
      return response.data;
    },
    staleTime: 2 * 60 * 1000, // 2 minutes
    refetchOnWindowFocus: true,
  });
}

/**
 * Hook to check a specific resource's limit status
 */
export function useResourceLimit(resource: ResourceType) {
  const { data, isLoading } = useOrganizationUsage();

  if (!data || isLoading) {
    return {
      usage: 0,
      limit: 0,
      isAtLimit: false,
      isUnlimited: false,
      percentUsed: 0,
      displayName: RESOURCE_DISPLAY_NAMES[resource],
      isLoading,
    };
  }

  const { limitKey, usageKey } = RESOURCE_LIMIT_MAP[resource];
  const limit = data.limits[limitKey] as number;
  const usage = data.usage[usageKey] as number;
  const isUnlimited = limit === -1;
  const isAtLimit = !isUnlimited && usage >= limit;
  const percentUsed = isUnlimited ? 0 : limit > 0 ? (usage / limit) * 100 : 0;

  return {
    usage,
    limit,
    isAtLimit,
    isUnlimited,
    percentUsed,
    displayName: RESOURCE_DISPLAY_NAMES[resource],
    isLoading,
  };
}
