import { getCurrentUser } from '@/auth'
import { redirect } from 'next/navigation'
import { MonitoringClient } from './_components/monitoring-client'

export const metadata = {
  title: 'System Monitoring',
  description: 'Real-time system performance and workflow monitoring',
}

export default async function MonitoringPage() {
  const user = await getCurrentUser()

  if (!user) {
    redirect('/login')
  }

  // Check if user is admin or compliance officer
  if (!['ADMIN', 'COMPLIANCE_OFFICER'].includes(user.role)) {
    redirect('/home')
  }

  return (
    <MonitoringClient userId={user.id} userRole={user.role} />
  )
}
