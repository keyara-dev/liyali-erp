import { MonitoringClient } from './_components/monitoring-client'

export const metadata = {
  title: 'System Monitoring',
  description: 'Real-time system performance and workflow monitoring',
}

// Disable static generation for this page
export const dynamic = 'force-dynamic'

export default function MonitoringPage() {
  // Use default values for client rendering
  return <MonitoringClient userId="system" userRole="ADMIN" />
}
