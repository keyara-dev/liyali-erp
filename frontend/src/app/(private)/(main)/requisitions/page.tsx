import { redirect } from "next/navigation";
import { RequisitionsClient } from "./_components/requisitions-client";
import { verifySession } from "@/lib/auth";

export const metadata = {
  title: "Requisitions",
  description: "Manage and approve requisition forms",
};

export default async function RequisitionsPage() {
  const { session, isAuthenticated } = await verifySession();

  if (!session || !isAuthenticated) {
    redirect("/login");
  }

  return (
    <RequisitionsClient userId={session.user.id} userRole={session.user.role} />
  );
}
