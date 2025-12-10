import { verifySession } from '@/lib/auth'
import { redirect } from 'next/navigation'
import { SettingsClient } from './_components/settings-client'

export const metadata = {
  title: 'Settings',
  description: 'Manage your account settings, security, and preferences',
}

export default async function SettingsPage() {
  const { session } = await verifySession()

  if (!session?.user) {
    redirect('/login')
  }

  return <SettingsClient user={session.user as any} />
}
