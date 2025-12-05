import { auth } from '@/auth'
import { redirect } from 'next/navigation'
import { POApprovalClient } from './_components/po-approval-client'

export const metadata = {
  title: 'Purchase Order Approval',
  description: 'Review and approve purchase order',
}

interface POApprovalPageProps {
  params: {
    id: string
  }
}

export default async function POApprovalPage({
  params,
}: POApprovalPageProps) {
  const session = await auth()

  if (!session?.user) {
    redirect('/login')
  }

  return (
    <POApprovalClient
      poId={params.id}
      userId={session.user.id}
      userRole={(session.user as any).role}
    />
  )
}
