export type DashboardVariant = "requester" | "approver" | "procurement" | "admin";

/**
 * Determine which dashboard variant to render for the current user.
 *
 * Priority order (highest first):
 *   admin → approver → procurement → requester
 */
export function getDashboardVariant(
  role: string,
  permissions: string[]
): DashboardVariant {
  const r = role.toLowerCase();

  if (
    ["admin", "super_admin"].includes(r) ||
    permissions.includes("admin.view")
  ) {
    return "admin";
  }

  if (
    ["manager", "finance", "approver", "department_manager"].includes(r) ||
    permissions.some((p) => p.endsWith(".approve"))
  ) {
    return "approver";
  }

  if (r === "procurement" || permissions.includes("purchase_order.create")) {
    return "procurement";
  }

  return "requester";
}
