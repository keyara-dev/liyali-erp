import { getCurrentUser } from '@/auth'
import { redirect } from 'next/navigation'
import { DashboardClient } from './_components/dashboard-client'

export const metadata = {
  title: 'Dashboard',
  description: 'View workflow metrics, approvals, and key statistics',
}

export default async function DashboardPage() {
  const user = await getCurrentUser()

  if (!user) {
    redirect('/login')
  }

  return (
    <DashboardClient userId={user.id} userRole={user.role} />
  )
}
