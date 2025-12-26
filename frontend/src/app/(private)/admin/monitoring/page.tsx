import { MonitoringClient } from './_components/monitoring-client'
import { verifySession } from '@/lib/auth'
import { redirect } from 'next/navigation'

export const metadata = {
  title: 'System Monitoring',
  description: 'Real-time system performance and workflow monitoring',
}

// Disable static generation for this page
export const dynamic = 'force-dynamic'

export default async function MonitoringPage() {
  // Get authenticated user context
  const { session, isAuthenticated } = await verifySession()

  if (!isAuthenticated || !session?.user) {
    redirect('/login')
  }

  // Verify admin role
  if (session.user.role !== 'ADMIN' && session.user.role !== 'SUPERADMIN') {
    redirect('/unauthorized')
  }

  // Pass actual user context from session
  return (
    <MonitoringClient
      userId={session.user.id}
      userRole={session.user.role}
    />
  )
}
