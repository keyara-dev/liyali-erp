import { redirect } from "next/navigation";
import { verifySession } from "@/lib/auth";

export default async function HomePage() {
  // Check if user is authenticated
  const { isAuthenticated, session } = await verifySession();

  if (!isAuthenticated || !session?.user_id) {
    // Not logged in - redirect to login
    redirect("/login");
  }

  // User is authenticated - redirect based on their role
  const userRole = session.role || "REQUESTER";

  // Map roles to their respective pages
  const roleRoutes: Record<string, string> = {
    ADMIN: "/home",
    FINANCE_OFFICER: "/home",
    DIRECTOR: "/home",
    CFO: "/home",
    DEPARTMENT_MANAGER: "/home/requisitions",
    REQUESTER: "/requisitions",
    COMPLIANCE_OFFICER: "/admin/compliance/tracking",
  };

  const redirectUrl = roleRoutes[userRole] ?? "/home";
  redirect(redirectUrl);
}
