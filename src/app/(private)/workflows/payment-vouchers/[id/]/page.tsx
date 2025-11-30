import { auth } from '@/auth'
import { redirect } from 'next/navigation'
import { PaymentVoucherDetail } from './_components/payment-voucher-detail'

export const metadata = {
  title: 'Payment Voucher Details',
  description: 'View and manage payment voucher details',
}

export default async function PaymentVoucherDetailPage({
  params,
}: {
  params: { id: string }
}) {
  const session = await auth()

  if (!session?.user) {
    redirect('/login')
  }

  return (
    <PaymentVoucherDetail
      pvId={params.id}
      userId={session.user.id}
      userRole={(session.user as any).role}
    />
  )
}
