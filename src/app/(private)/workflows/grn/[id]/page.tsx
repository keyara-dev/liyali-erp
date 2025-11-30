import { auth } from '@/auth'
import { redirect } from 'next/navigation'
import { GrnDetail } from './_components/grn-detail'

export const metadata = {
  title: 'Goods Received Note Details',
  description: 'View and manage GRN details',
}

interface GrnDetailPageProps {
  params: {
    id: string
  }
}

export default async function GrnDetailPage({ params }: GrnDetailPageProps) {
  const session = await auth()

  if (!session?.user) {
    redirect('/login')
  }

  return (
    <GrnDetail
      grnId={params.id}
      userId={session.user.id}
      userRole={(session.user as any).role}
    />
  )
}
