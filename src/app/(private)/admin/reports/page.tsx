import { getCurrentUser } from '@/auth'
import { redirect } from 'next/navigation'
import { AdminReportsClient } from './_components/admin-reports-client'

export const metadata = {
  title: 'Admin Reports',
  description: 'View approval statistics, user activity, and system reports',
}

export default async function AdminReportsPage() {
  const user = await getCurrentUser()

  if (!user) {
    redirect('/login')
  }

  // Check if user is admin
  if (!['ADMIN', 'COMPLIANCE_OFFICER'].includes(user.role)) {
    redirect('/home')
  }

  return (
    <AdminReportsClient userId={user.id} userRole={user.role} />
  )
}
