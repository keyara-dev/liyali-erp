import { EditWorkflowClient } from "./_components/edit-workflow-client";

export const metadata = {
  title: "Edit Workflow",
  description: "Edit an existing approval workflow",
};

interface EditWorkflowPageProps {
  params: Promise<{
    id: string;
  }>;
}

// Disable static generation for this page
export const dynamic = 'force-dynamic'

export default async function EditWorkflowPage({
  params,
}: EditWorkflowPageProps) {
  const { id } = await params;

  return (
    <EditWorkflowClient
      workflowId={id}
      userId="system"
      userRole="ADMIN"
    />
  );
}
