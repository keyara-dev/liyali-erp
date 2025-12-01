import { redirect } from 'next/navigation'

export const metadata = {
  title: 'Approvals',
  description: 'View and manage your approval tasks',
}

export default function ApprovalsRedirect() {
  redirect('/workflows/tasks?tab=approvals')
}
