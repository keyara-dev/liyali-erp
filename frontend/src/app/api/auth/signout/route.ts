import { redirect } from 'next/navigation'
import { logoutAction } from '@/app/_actions/auth-actions'

export async function GET() {
  try {
    // Clear the session using the logout action
    await logoutAction()
  } catch (error) {
    console.error('Logout error:', error)
  }

  // Redirect to login page after logout (handles both success and failure)
  // Note: redirect() throws a special Next.js error that must not be caught
  redirect('/login')
}
