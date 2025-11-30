import { getCurrentUser } from '@/auth'
import { redirect } from 'next/navigation'
import { ComplianceTrackingClient } from './_components/compliance-tracking-client'

export const metadata = {
  title: 'Compliance Tracking',
  description: 'Track regulatory compliance and audit requirements',
}

export default async function ComplianceTrackingPage() {
  const user = await getCurrentUser()

  if (!user) {
    redirect('/login')
  }

  // Check if user is compliance officer or admin
  if (!['ADMIN', 'COMPLIANCE_OFFICER'].includes(user.role)) {
    redirect('/workflows')
  }

  return (
    <ComplianceTrackingClient userId={user.id} userRole={user.role} />
  )
}
