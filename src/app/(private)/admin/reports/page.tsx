import { AdminReportsClient } from './_components/admin-reports-client'

export const metadata = {
  title: 'Admin Reports',
  description: 'View approval statistics, user activity, and system reports',
}

// Disable static generation for this page
export const dynamic = 'force-dynamic'

export default function AdminReportsPage() {
  // Use default values for client rendering
  return <AdminReportsClient userId="system" userRole="ADMIN" />
}
