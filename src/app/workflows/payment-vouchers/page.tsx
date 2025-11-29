import { auth } from '@/auth'
import { redirect } from 'next/navigation'
import { PaymentVouchersClient } from './_components/payment-vouchers-client'

export const metadata = {
  title: 'Payment Vouchers',
  description: 'Manage and approve payment vouchers',
}

export default async function PaymentVouchersPage() {
  const session = await auth()

  if (!session?.user) {
    redirect('/login')
  }

  return (
    <PaymentVouchersClient
      userId={session.user.id}
      userRole={(session.user as any).role}
    />
  )
}
