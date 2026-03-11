import { MonitoringClient } from '../_components/monitoring-client'
import { requireAdminRole } from '@/lib/admin-guard'

export const metadata = {
  title: 'System Monitoring',
  description: 'Real-time system performance and workflow monitoring',
}

export const dynamic = 'force-dynamic'

export default async function MonitoringPage() {
  const { userId, userRole } = await requireAdminRole()

  return <MonitoringClient userId={userId} userRole={userRole} />
}
