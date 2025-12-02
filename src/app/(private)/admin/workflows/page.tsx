import { auth } from '@/auth'
import { redirect } from 'next/navigation'
import { WorkflowsClient } from './_components/workflows-client'

export const metadata = {
  title: 'Workflow Management',
  description: 'Create and manage custom approval workflows',
}

export default async function WorkflowsPage() {
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
    <WorkflowsClient userId={session.user.id} userRole={userRole} />
  )
}
