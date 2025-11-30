import { auth } from '@/auth'
import { redirect } from 'next/navigation'
import { TasksClient } from './_components/tasks-client'

export const metadata = {
  title: 'Tasks',
  description: 'View and manage your pending workflow tasks',
}

export default async function TasksPage() {
  const session = await auth()

  if (!session?.user) {
    redirect('/login')
  }

  return (
    <TasksClient userId={session.user.id} userRole={(session.user as any).role} />
  )
}
