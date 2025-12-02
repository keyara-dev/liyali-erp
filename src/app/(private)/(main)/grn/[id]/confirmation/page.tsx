import { auth } from '@/auth'
import { redirect } from 'next/navigation'
import { GRNConfirmationClient } from './_components/grn-confirmation-client'

export const metadata = {
  title: 'GRN Confirmation',
  description: 'Confirm goods received',
}

interface GRNConfirmationPageProps {
  params: {
    id: string
  }
}

export default async function GRNConfirmationPage({
  params,
}: GRNConfirmationPageProps) {
  const session = await auth()

  if (!session?.user) {
    redirect('/login')
  }

  return (
    <GRNConfirmationClient
      grnId={params.id}
      userId={session.user.id}
      userRole={(session.user as any).role}
    />
  )
}
