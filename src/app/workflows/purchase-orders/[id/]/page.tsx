import { auth } from '@/auth'
import { redirect } from 'next/navigation'
import { PurchaseOrderDetail } from './_components/purchase-order-detail'

export const metadata = {
  title: 'Purchase Order Details',
  description: 'View and manage purchase order details',
}

export default async function PurchaseOrderDetailPage({
  params,
}: {
  params: { id: string }
}) {
  const session = await auth()

  if (!session?.user) {
    redirect('/login')
  }

  return (
    <PurchaseOrderDetail
      poId={params.id}
      userId={session.user.id}
      userRole={(session.user as any).role}
    />
  )
}
