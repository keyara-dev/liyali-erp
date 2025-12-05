import { CreateWorkflowClient } from "./_components/create-workflow-client";

export const metadata = {
  title: "Create Workflow",
  description: "Create a new custom approval workflow",
};

// Disable static generation for this page
export const dynamic = 'force-dynamic'

export default function CreateWorkflowPage() {
  // Use default values for client rendering
  return <CreateWorkflowClient userId="system" userRole="ADMIN" />;
}
