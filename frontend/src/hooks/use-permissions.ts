"use client";

import { useMemo } from "react";
import { useSession } from "./use-session";
import type { User } from "@/types";

/**
 * Represents a permission as a resource-action pair
 * @example
 * ```typescript
 * const permission: PermissionCheck = { resource: "requisition", action: "approve" };
 * ```
 */
export interface PermissionCheck {
  resource: string;
  action: string;
}

/**
 * Data structure returned by usePermissions hook
 */
export interface PermissionsData {
  /** Check if user has a specific permission */
  hasPermission: (resource: string, action: string) => boolean;
  /** Check if user has ALL of the required permissions */
  hasAllPermissions: (permissions: PermissionCheck[]) => boolean;
  /** Check if user has ANY of the required permissions */
  hasAnyPermission: (permissions: PermissionCheck[]) => boolean;
  /** Get all permissions for the current user's role */
  getPermissions: () => PermissionCheck[];
  /** Check if user is an admin */
  isAdmin: () => boolean;
  /** Check if user is an approver */
  isApprover: () => boolean;
  /** Check if user is a requester */
  isRequester: () => boolean;
  /** Check if user is finance */
  isFinance: () => boolean;
  /** User's current role */
  userRole: string | null;
  /** Whether permissions are still loading */
  isLoading: boolean;
  /** Any errors encountered */
  error: Error | null;
}

/**
 * Hardcoded role-to-permission mapping (mirrors backend)
 * This is used as a fallback when the user's role information is available
 */
const ROLE_PERMISSIONS: Record<string, PermissionCheck[]> = {
  admin: [
    // Requisition permissions
    { resource: "requisition", action: "view" },
    { resource: "requisition", action: "create" },
    { resource: "requisition", action: "edit" },
    { resource: "requisition", action: "delete" },
    { resource: "requisition", action: "approve" },
    { resource: "requisition", action: "reject" },

    // Budget permissions
    { resource: "budget", action: "view" },
    { resource: "budget", action: "create" },
    { resource: "budget", action: "edit" },
    { resource: "budget", action: "delete" },
    { resource: "budget", action: "approve" },
    { resource: "budget", action: "reject" },

    // Purchase Order permissions
    { resource: "purchase_order", action: "view" },
    { resource: "purchase_order", action: "create" },
    { resource: "purchase_order", action: "edit" },
    { resource: "purchase_order", action: "delete" },
    { resource: "purchase_order", action: "approve" },
    { resource: "purchase_order", action: "reject" },

    // Payment Voucher permissions
    { resource: "payment_voucher", action: "view" },
    { resource: "payment_voucher", action: "create" },
    { resource: "payment_voucher", action: "edit" },
    { resource: "payment_voucher", action: "delete" },
    { resource: "payment_voucher", action: "approve" },
    { resource: "payment_voucher", action: "reject" },

    // GRN permissions
    { resource: "grn", action: "view" },
    { resource: "grn", action: "create" },
    { resource: "grn", action: "edit" },
    { resource: "grn", action: "delete" },

    // Vendor permissions
    { resource: "vendor", action: "view" },
    { resource: "vendor", action: "create" },
    { resource: "vendor", action: "edit" },
    { resource: "vendor", action: "delete" },

    // Category permissions
    { resource: "category", action: "view" },
    { resource: "category", action: "create" },
    { resource: "category", action: "edit" },
    { resource: "category", action: "delete" },

    // Organization permissions
    { resource: "organization", action: "view" },
    { resource: "organization", action: "edit" },
    { resource: "organization", action: "manage_users" },
    { resource: "organization", action: "manage_workflows" },

    // Analytics & Audit
    { resource: "analytics", action: "view" },
    { resource: "audit_log", action: "view" },
  ],
  approver: [
    { resource: "requisition", action: "view" },
    { resource: "requisition", action: "create" },
    { resource: "requisition", action: "edit" },
    { resource: "requisition", action: "approve" },
    { resource: "requisition", action: "reject" },

    { resource: "budget", action: "view" },
    { resource: "budget", action: "approve" },
    { resource: "budget", action: "reject" },

    { resource: "purchase_order", action: "view" },
    { resource: "purchase_order", action: "approve" },
    { resource: "purchase_order", action: "reject" },

    { resource: "payment_voucher", action: "view" },
    { resource: "payment_voucher", action: "approve" },
    { resource: "payment_voucher", action: "reject" },

    { resource: "grn", action: "view" },
    { resource: "vendor", action: "view" },
    { resource: "category", action: "view" },
    { resource: "analytics", action: "view" },
  ],
  requester: [
    { resource: "requisition", action: "view" },
    { resource: "requisition", action: "create" },
    { resource: "requisition", action: "edit" },

    { resource: "budget", action: "view" },
    { resource: "purchase_order", action: "view" },
    { resource: "payment_voucher", action: "view" },
    { resource: "grn", action: "view" },
    { resource: "vendor", action: "view" },
    { resource: "category", action: "view" },
  ],
  finance: [
    { resource: "requisition", action: "view" },
    { resource: "requisition", action: "approve" },
    { resource: "requisition", action: "reject" },

    { resource: "budget", action: "view" },
    { resource: "budget", action: "create" },
    { resource: "budget", action: "edit" },
    { resource: "budget", action: "approve" },
    { resource: "budget", action: "reject" },

    { resource: "purchase_order", action: "view" },
    { resource: "purchase_order", action: "approve" },
    { resource: "purchase_order", action: "reject" },

    { resource: "payment_voucher", action: "view" },
    { resource: "payment_voucher", action: "create" },
    { resource: "payment_voucher", action: "edit" },
    { resource: "payment_voucher", action: "approve" },
    { resource: "payment_voucher", action: "reject" },

    { resource: "grn", action: "view" },
    { resource: "vendor", action: "view" },
    { resource: "category", action: "view" },
    { resource: "analytics", action: "view" },
    { resource: "audit_log", action: "view" },
  ],
  viewer: [
    { resource: "requisition", action: "view" },
    { resource: "budget", action: "view" },
    { resource: "purchase_order", action: "view" },
    { resource: "payment_voucher", action: "view" },
    { resource: "grn", action: "view" },
    { resource: "vendor", action: "view" },
    { resource: "category", action: "view" },
    { resource: "analytics", action: "view" },
  ],
};

/**
 * Hook for permission-based access control
 *
 * Provides methods to check if the current user has specific permissions
 * based on their role. Integrates with the backend permission system.
 *
 * @returns {PermissionsData} Object with permission checking methods
 *
 * @example
 * ```typescript
 * const { hasPermission, isAdmin } = usePermissions();
 *
 * return (
 *   <>
 *     {hasPermission("requisition", "approve") && (
 *       <button onClick={approve}>Approve</button>
 *     )}
 *     {isAdmin() && (
 *       <button onClick={manageUsers}>Manage Users</button>
 *     )}
 *   </>
 * );
 * ```
 */
export function usePermissions(): PermissionsData {
  const { user, isLoading, error } = useSession();

  const permissions = useMemo(() => {
    if (!user || !user.role) {
      return [];
    }

    const roleName = user.role.toLowerCase();
    return ROLE_PERMISSIONS[roleName] || [];
  }, [user]);

  const hasPermission = (resource: string, action: string): boolean => {
    if (!user || !user.role) {
      return false;
    }

    return permissions.some(
      (perm) => perm.resource === resource && perm.action === action
    );
  };

  const hasAllPermissions = (requiredPerms: PermissionCheck[]): boolean => {
    return requiredPerms.every((perm) =>
      hasPermission(perm.resource, perm.action)
    );
  };

  const hasAnyPermission = (requiredPerms: PermissionCheck[]): boolean => {
    return requiredPerms.some((perm) =>
      hasPermission(perm.resource, perm.action)
    );
  };

  const getPermissions = (): PermissionCheck[] => {
    return [...permissions];
  };

  const isAdmin = (): boolean => {
    return user?.role?.toLowerCase() === "admin";
  };

  const isApprover = (): boolean => {
    return user?.role?.toLowerCase() === "approver";
  };

  const isRequester = (): boolean => {
    return user?.role?.toLowerCase() === "requester";
  };

  const isFinance = (): boolean => {
    return user?.role?.toLowerCase() === "finance";
  };

  return {
    hasPermission,
    hasAllPermissions,
    hasAnyPermission,
    getPermissions,
    isAdmin,
    isApprover,
    isRequester,
    isFinance,
    userRole: user?.role ?? null,
    isLoading,
    error,
  };
}
