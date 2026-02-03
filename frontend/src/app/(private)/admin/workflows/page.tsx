import { WorkflowsClient } from "../_components/workflows-client";
import { requireAdminRole } from "@/lib/admin-guard";

export const metadata = {
  title: "Workflow Management",
  description: "Create and manage custom approval workflows",
};

// Disable static generation for this page
export const dynamic = 'force-dynamic'

export default async function WorkflowsPage() {
  // Verify admin role at server level
  const { userId, userRole } = await requireAdminRole()

  return <WorkflowsClient userId={userId} userRole={userRole} />;
}
