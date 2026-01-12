import { verifySession } from "@/lib/auth";
import { redirect } from "next/navigation";
import { PropsWithChildren } from "react";
import DashboardLayoutProvider from "../(main)/layout";

export const dynamic = "force-dynamic";

export default async function AdminLayout({ children }: PropsWithChildren) {
  const { session, isAuthenticated } = await verifySession();

  if (!session || !isAuthenticated) {
    redirect("/login");
  }

  // Check if user is admin
  if (session.user.role !== "admin" || !session.user) {
    redirect("/access-denied");
  }

  return <DashboardLayoutProvider>{children}</DashboardLayoutProvider>;
}
