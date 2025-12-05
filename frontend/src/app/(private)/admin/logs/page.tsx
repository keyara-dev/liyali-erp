import { ActivityLogsClient } from './_components/activity-logs-client'

export const metadata = {
  title: 'Activity Logs',
  description: 'View system activity logs and audit trail',
}

// Disable static generation for this page
export const dynamic = 'force-dynamic'

export default function ActivityLogsPage() {
  // Use default values for client rendering
  return <ActivityLogsClient userId="system" userRole="ADMIN" />
}
