import { auth } from "@/auth";
import { redirect } from "next/navigation";
import { EditWorkflowClient } from "./_components/edit-workflow-client";
import { verifySession } from "@/lib/auth";

export const metadata = {
  title: "Edit Workflow",
  description: "Edit an existing approval workflow",
};

interface EditWorkflowPageProps {
  params: Promise<{
    id: string;
  }>;
}

export default async function EditWorkflowPage({
  params,
}: EditWorkflowPageProps) {
  const { session } = await verifySession();

  const { id } = await params;

  // Only allow admin users
  const userRole = (session?.user as any).role;
  if (userRole !== "ADMIN") {
    redirect("/home");
  }

  return (
    <EditWorkflowClient
      workflowId={id}
      userId={String(session?.user?.id)}
      userRole={userRole}
    />
  );
}
