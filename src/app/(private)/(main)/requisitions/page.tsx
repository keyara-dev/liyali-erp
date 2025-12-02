import { auth } from '@/auth'
import { redirect } from 'next/navigation'
import { RequisitionsClient } from './_components/requisitions-client'

export const metadata = {
  title: 'Requisitions',
  description: 'Manage and approve requisition forms',
}

export default async function RequisitionsPage() {
  const session = await auth()

  if (!session?.user) {
    redirect('/login')
  }

  return (
    <RequisitionsClient userId={session.user.id} userRole={(session.user as any).role} />
  )
}
