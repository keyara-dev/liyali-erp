import { redirect } from "next/navigation";
import { verifySession } from "@/lib/auth";

/**
 * Home Page - Authentication & Role-Based Redirect
 *
 * Checks if user is logged in and redirects them to their respective dashboard
 * based on their role:
 * - ADMIN: /workflows/dashboard
 * - FINANCE_OFFICER: /workflows/dashboard
 * - DIRECTOR: /workflows/dashboard
 * - CFO: /workflows/dashboard
 * - DEPARTMENT_MANAGER: /workflows/requisitions
 * - REQUESTER: /workflows/requisitions
 * - COMPLIANCE_OFFICER: /compliance/tracking
 */
export default async function HomePage() {
  // Check if user is authenticated
  const { isAuthenticated, session } = await verifySession();

  if (!isAuthenticated || !session?.user_id) {
    // Not logged in - redirect to login
    redirect("/login");
  }

  // User is authenticated - redirect based on their role
  const userRole = session.user_type || "REQUESTER";

  // Map roles to their respective pages
  const roleRoutes: Record<string, string> = {
    ADMIN: "/workflows/dashboard",
    FINANCE_OFFICER: "/workflows/dashboard",
    DIRECTOR: "/workflows/dashboard",
    CFO: "/workflows/dashboard",
    DEPARTMENT_MANAGER: "/workflows/requisitions",
    REQUESTER: "/workflows/requisitions",
    COMPLIANCE_OFFICER: "/compliance/tracking",
  };

  const redirectUrl = roleRoutes[userRole] ?? "/workflows/dashboard";
  redirect(redirectUrl);
}
