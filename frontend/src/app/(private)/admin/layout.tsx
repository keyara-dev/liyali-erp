import { verifySession } from "@/lib/auth";
import { redirect } from "next/navigation";
import { PropsWithChildren } from "react";

export default async function AdminLayout({ children }: PropsWithChildren) {
  const { session, isAuthenticated } = await verifySession();

  if (!session || !isAuthenticated) {
    redirect("/login");
  }

  // Check if user is admin
  if (
    session.user_type != "ADMIN" ||
    session.user.role !== "ADMIN" ||
    !session.user
  ) {
    redirect("/access-denied");
  }

  return <>{children}</>;
}
