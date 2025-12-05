import { ComplianceTrackingClient } from './_components/compliance-tracking-client'

export const metadata = {
  title: 'Compliance Tracking',
  description: 'Track regulatory compliance and audit requirements',
}

// Disable static generation for this page
export const dynamic = 'force-dynamic'

export default function ComplianceTrackingPage() {
  // Use default values for client rendering
  return <ComplianceTrackingClient userId="system" userRole="ADMIN" />
}
