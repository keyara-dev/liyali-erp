import { auth } from '@/auth'
import { redirect } from 'next/navigation'
import { EditWorkflowClient } from './_components/edit-workflow-client'

export const metadata = {
  title: 'Edit Workflow',
  description: 'Edit an existing approval workflow',
}

interface EditWorkflowPageProps {
  params: {
    id: string
  }
}

export default async function EditWorkflowPage({
  params,
}: EditWorkflowPageProps) {
  const session = await auth()

  if (!session?.user) {
    redirect('/login')
  }

  // Only allow admin users
  const userRole = (session.user as any).role
  if (userRole !== 'ADMIN') {
    redirect('/workflows/dashboard')
  }

  return (
    <EditWorkflowClient
      workflowId={params.id}
      userId={session.user.id}
      userRole={userRole}
    />
  )
}
