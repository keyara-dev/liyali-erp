import { WorkflowsClient } from "./_components/workflows-client";

export const metadata = {
  title: "Workflow Management",
  description: "Create and manage custom approval workflows",
};

// Disable static generation for this page
export const dynamic = 'force-dynamic'

export default function WorkflowsPage() {
  // Use default values for client rendering
  return <WorkflowsClient userId="system" userRole="ADMIN" />;
}
