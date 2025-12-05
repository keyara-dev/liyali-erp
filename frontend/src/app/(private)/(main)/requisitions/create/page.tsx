import { getCurrentUser } from '@/auth'
import { redirect } from 'next/navigation'
import { CreateRequisitionClient } from './_components/create-requisition-client'

export const metadata = {
  title: 'Create Requisition',
  description: 'Create a new requisition form for purchase approval',
}

export default async function CreateRequisitionPage() {
  const user = await getCurrentUser()

  if (!user) {
    redirect('/login')
  }

  return (
    <CreateRequisitionClient
      userId={user.id}
      userRole={user.role}
      userName={user.name || 'User'}
    />
  )
}
