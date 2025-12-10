import { redirect } from "next/navigation";
import { DashboardClient } from "./_components/dashboard-client";
import { verifySession } from "@/lib/auth";

export const metadata = {
  title: "Dashboard",
  description: "View workflow metrics, approvals, and key statistics",
};

export default async function DashboardPage() {
  const { session, isAuthenticated } = await verifySession();

  if (!session || !isAuthenticated) {
    redirect("/login");
  }

  return (
    <DashboardClient
      userId={String(session?.user?.id)}
      userRole={String(session?.user?.role)}
    />
  );
}
