import { auth } from '@/auth'
import { redirect } from 'next/navigation'
import { PurchaseOrdersClient } from './_components/purchase-orders-client'

export const metadata = {
  title: 'Purchase Orders',
  description: 'Manage and approve purchase orders',
}

export default async function PurchaseOrdersPage() {
  const session = await auth()

  if (!session?.user) {
    redirect('/login')
  }

  return (
    <PurchaseOrdersClient
      userId={session.user.id}
      userRole={(session.user as any).role}
    />
  )
}
