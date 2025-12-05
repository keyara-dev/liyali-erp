import { redirect } from 'next/navigation'
import { logoutAction } from '@/app/_actions/auth-actions'

export async function GET() {
  try {
    // Clear the session using the logout action
    await logoutAction()

    // Redirect to login page after successful logout
    redirect('/login')
  } catch (error) {
    console.error('Logout error:', error)
    // Still redirect to login even if logout fails
    redirect('/login')
  }
}
