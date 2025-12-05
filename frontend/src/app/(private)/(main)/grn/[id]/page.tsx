import { auth } from '@/auth'
import { redirect } from 'next/navigation'
import { GRNDetailClient } from './_components/grn-detail-client'

export const metadata = {
  title: 'Goods Received Note Details',
  description: 'View and confirm goods received',
}

interface GRNDetailPageProps {
  params: {
    id: string
  }
}

export default async function GRNDetailPage({
  params,
}: GRNDetailPageProps) {
  const session = await auth()

  if (!session?.user) {
    redirect('/login')
  }

  return (
    <GRNDetailClient
      grnId={params.id}
      userId={session.user.id}
      userRole={(session.user as any).role}
    />
  )
}
