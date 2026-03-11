import { AdminReportsClient } from "../_components/admin-reports-client";
import { verifySession } from "@/lib/auth";
import { redirect } from "next/navigation";

export const metadata = {
  title: "Admin Reports",
  description: "View approval statistics, user activity, and system reports",
};

// Disable static generation for this page
export const dynamic = "force-dynamic";

export default async function AdminReportsPage() {
  // Get authenticated user context
  const { session, isAuthenticated } = await verifySession();

  if (!isAuthenticated || !session?.user) {
    redirect("/login");
  }

  // Verify admin role
  if (session.user.role !== "admin") {
    redirect("/unauthorized");
  }

  // Pass actual user context from session
  return (
    <AdminReportsClient userId={session.user.id} userRole={session.user.role} />
  );
}
