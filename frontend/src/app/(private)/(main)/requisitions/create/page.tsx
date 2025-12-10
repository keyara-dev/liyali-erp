import { redirect } from "next/navigation";
import { CreateRequisitionClient } from "./_components/create-requisition-client";
import { verifySession } from "@/lib/auth";

export const metadata = {
  title: "Create Requisition",
  description: "Create a new requisition form for purchase approval",
};

export default async function CreateRequisitionPage() {
  const { session, isAuthenticated } = await verifySession();

  if (!isAuthenticated) {
    redirect("/login");
  }

  return (
    <CreateRequisitionClient
      userId={String(session?.user?.id)}
      userRole={String(session?.user?.role)}
      userName={String(session?.user?.name) || "User"}
    />
  );
}
