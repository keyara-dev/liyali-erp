import { auth } from '@/auth'
import { redirect } from 'next/navigation'
import { CreateWorkflowClient } from './_components/create-workflow-client'

export const metadata = {
  title: 'Create Workflow',
  description: 'Create a new custom approval workflow',
}

export default async function CreateWorkflowPage() {
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
    <CreateWorkflowClient userId={session.user.id} userRole={userRole} />
  )
}
