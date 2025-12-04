import { auth } from '@/auth'
import { redirect } from 'next/navigation'
import { PVCreateClient } from './_components/pv-create-client'

export const metadata = {
  title: 'Create Payment Voucher',
  description: 'Create a new payment voucher from an approved purchase order',
}

export default async function CreatePaymentVoucherPage() {
  const session = await auth()

  if (!session?.user) {
    redirect('/login')
  }

  return (
    <PVCreateClient
      userId={session.user.id}
      userName={session.user.name || 'User'}
      userRole={(session.user as any).role || 'USER'}
    />
  )
}
