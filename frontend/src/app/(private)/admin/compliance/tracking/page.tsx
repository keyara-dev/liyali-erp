import { ComplianceTrackingClient } from '../../_components/compliance-tracking-client'
import { requireAdminRole } from '@/lib/admin-guard'

export const metadata = {
  title: 'Compliance Tracking',
  description: 'Track regulatory compliance and audit requirements',
}

export const dynamic = 'force-dynamic'

export default async function ComplianceTrackingPage() {
  const { userId, userRole } = await requireAdminRole()

  return <ComplianceTrackingClient userId={userId} userRole={userRole} />
}
