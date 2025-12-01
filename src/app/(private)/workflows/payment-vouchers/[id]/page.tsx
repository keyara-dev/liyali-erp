import { auth } from '@/auth'
import { redirect } from 'next/navigation'
import { PVDetailClient } from './_components/pv-detail-client'

export const metadata = {
  title: 'Payment Voucher Details',
  description: 'View and manage payment voucher details',
}

interface PVDetailPageProps {
  params: {
    id: string
  }
}

export default async function PVDetailPage({
  params,
}: PVDetailPageProps) {
  const session = await auth()

  if (!session?.user) {
    redirect('/login')
  }

  return (
    <PVDetailClient
      pvId={params.id}
      userId={session.user.id}
      userRole={(session.user as any).role}
    />
  )
}
