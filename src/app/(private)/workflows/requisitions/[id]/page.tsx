import { auth } from '@/auth'
import { redirect } from 'next/navigation'
import { RequisitionDetailClient } from '../_components/requisition-detail-client'

export const metadata = {
  title: 'Requisition Details',
  description: 'View and manage requisition details',
}

interface RequisitionDetailPageProps {
  params: {
    id: string
  }
}

export default async function RequisitionDetailPage({
  params,
}: RequisitionDetailPageProps) {
  const session = await auth()

  if (!session?.user) {
    redirect('/login')
  }

  return (
    <RequisitionDetailClient
      requisitionId={params.id}
      userId={session.user.id}
      userRole={(session.user as any).role}
    />
  )
}
