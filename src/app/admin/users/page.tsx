import { getCurrentUser } from '@/auth'
import { redirect } from 'next/navigation'
import { UserManagementClient } from './_components/user-management-client'

export const metadata = {
  title: 'User Management',
  description: 'Manage user roles and access permissions',
}

export default async function UserManagementPage() {
  const user = await getCurrentUser()

  if (!user) {
    redirect('/login')
  }

  // Check if user is admin
  if (user.role !== 'ADMIN') {
    redirect('/workflows')
  }

  return (
    <UserManagementClient userId={user.id} userRole={user.role} />
  )
}
