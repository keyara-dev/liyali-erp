import { getCurrentUser } from '@/auth'
import { redirect } from 'next/navigation'
import { ActivityLogsClient } from './_components/activity-logs-client'

export const metadata = {
  title: 'Activity Logs',
  description: 'View system activity logs and audit trail',
}

export default async function ActivityLogsPage() {
  const user = await getCurrentUser()

  if (!user) {
    redirect('/login')
  }

  // Check if user is admin or compliance officer
  if (!['ADMIN', 'COMPLIANCE_OFFICER'].includes(user.role)) {
    redirect('/home')
  }

  return (
    <ActivityLogsClient userId={user.id} userRole={user.role} />
  )
}
