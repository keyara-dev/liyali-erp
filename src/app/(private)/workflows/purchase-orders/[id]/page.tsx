import { auth } from '@/auth'
import { redirect } from 'next/navigation'
import { PODetailClient } from './_components/po-detail-client'

export const metadata = {
  title: 'Purchase Order Details',
  description: 'View and manage purchase order details',
}

interface PODetailPageProps {
  params: {
    id: string
  }
}

export default async function PODetailPage({
  params,
}: PODetailPageProps) {
  const session = await auth()

  if (!session?.user) {
    redirect('/login')
  }

  return (
    <PODetailClient
      poId={params.id}
      userId={session.user.id}
      userRole={(session.user as any).role}
    />
  )
}
