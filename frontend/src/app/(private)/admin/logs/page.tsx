import { ActivityLogsClient } from './_components/activity-logs-client'
import { requireAdminRole } from '@/lib/admin-guard'

export const metadata = {
  title: 'Activity Logs',
  description: 'View system activity logs and audit trail',
}

// Disable static generation for this page
export const dynamic = 'force-dynamic'

export default async function ActivityLogsPage() {
  // Verify admin role at server level
  const { userId, userRole } = await requireAdminRole()

  return <ActivityLogsClient userId={userId} userRole={userRole} />
}
