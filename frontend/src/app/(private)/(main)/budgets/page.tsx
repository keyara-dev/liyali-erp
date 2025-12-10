import { verifySession } from '@/lib/auth'
import { redirect } from 'next/navigation'
import { BudgetsClient } from './_components/budgets-client'

export const metadata = {
  title: 'Budgets',
  description: 'Manage and approve budgets',
}

export default async function BudgetsPage() {
  const { session } = await verifySession()

  if (!session?.user) {
    redirect('/login')
  }

  return (
    <BudgetsClient userId={session.user.id} userRole={(session.user as any).role} />
  )
}
