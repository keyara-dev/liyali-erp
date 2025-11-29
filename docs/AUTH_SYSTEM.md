# Simulated Authentication System

## Overview

This project uses a **simulated authentication system** that replaces NextAuth.js for development and testing purposes. This lightweight system provides session management, role-based access control, and demo user accounts without external dependencies.

## Architecture

### Core Files

- **[src/lib/auth.ts](src/lib/auth.ts)** - Core authentication logic and session management
- **[src/app/_actions/auth-actions.ts](src/app/_actions/auth-actions.ts)** - Server action wrappers for auth functions
- **[src/auth.ts](src/auth.ts)** - Public API exports (replaces NextAuth)
- **[src/app/login/page.tsx](src/app/login/page.tsx)** - Login page
- **[src/app/login/_components/login-form.tsx](src/app/login/_components/login-form.tsx)** - Login form component

## How It Works

### Session Management

Sessions are stored as Base64-encoded JSON in HTTP-only cookies:

```typescript
// Session Structure
interface Session {
  user: AuthUser
  expiresAt: number
}

// Cookie: `auth_session`
// Expires after 24 hours
// HttpOnly, Secure (in production), SameSite=Lax
```

### Authentication Flow

1. **Login**: User provides email and password
2. **Validation**: Credentials checked against `DEMO_USERS` map
3. **Session Creation**: Session cookie created with 24-hour expiration
4. **Redirect**: User redirected to `/workflows/dashboard`

### Protected Pages

All page.tsx files check for authenticated user:

```typescript
// All pages follow this pattern
import { getCurrentUser } from '@/auth'
import { redirect } from 'next/navigation'

export default async function PageName() {
  const user = await getCurrentUser()

  if (!user) {
    redirect('/login')
  }

  // Role-based checks for admin pages
  if (!['ADMIN', 'COMPLIANCE_OFFICER'].includes(user.role)) {
    redirect('/workflows')
  }

  return <ClientComponent userId={user.id} userRole={user.role} />
}
```

## Demo Users

All demo users use the same password: `password123`

| Email | Role | Department | Avatar |
|-------|------|------------|--------|
| requester@liyali.com | REQUESTER | Operations | 👤 |
| manager@liyali.com | DEPARTMENT_MANAGER | Finance | 👥 |
| finance@liyali.com | FINANCE_OFFICER | Finance | 💼 |
| director@liyali.com | DIRECTOR | Executive | 👔 |
| cfo@liyali.com | CFO | Finance | 💎 |
| compliance@liyali.com | COMPLIANCE_OFFICER | Compliance | ✅ |
| admin@liyali.com | ADMIN | Administration | ⚙️ |

## User Roles

```typescript
type UserRole =
  | 'REQUESTER'
  | 'DEPARTMENT_MANAGER'
  | 'FINANCE_OFFICER'
  | 'DIRECTOR'
  | 'CFO'
  | 'COMPLIANCE_OFFICER'
  | 'ADMIN'
```

### Role Permissions

| Role | Dashboard | Search | Create | Reports | Users | Logs | Compliance | Monitoring | QR Verify |
|------|-----------|--------|--------|---------|-------|------|-----------|-----------|-----------|
| REQUESTER | ✓ | ✓ | ✓ | ✗ | ✗ | ✗ | ✗ | ✗ | ✓ |
| DEPARTMENT_MANAGER | ✓ | ✓ | ✓ | ✗ | ✗ | ✗ | ✗ | ✗ | ✓ |
| FINANCE_OFFICER | ✓ | ✓ | ✓ | ✗ | ✗ | ✗ | ✗ | ✗ | ✓ |
| DIRECTOR | ✓ | ✓ | ✓ | ✗ | ✗ | ✗ | ✗ | ✗ | ✓ |
| CFO | ✓ | ✓ | ✓ | ✗ | ✗ | ✗ | ✗ | ✗ | ✓ |
| COMPLIANCE_OFFICER | ✓ | ✓ | ✓ | ✓ | ✗ | ✓ | ✓ | ✓ | ✓ |
| ADMIN | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |

## API Reference

### Core Functions (src/lib/auth.ts)

#### `getSession(): Promise<Session | null>`
Retrieve the current session from cookies.

```typescript
const session = await getSession()
if (session) {
  console.log(session.user.name) // Get user info
}
```

#### `getCurrentUser(): Promise<AuthUser | null>`
Get the current authenticated user.

```typescript
const user = await getCurrentUser()
if (user) {
  console.log(`Welcome ${user.name}`)
}
```

#### `login(email: string, password: string): Promise<{success: boolean, user?: AuthUser, error?: string}>`
Authenticate user and create session.

```typescript
const result = await login('admin@liyali.com', 'password123')
if (result.success) {
  // User is now authenticated
}
```

#### `logout(): Promise<void>`
Clear session and log user out.

```typescript
await logout()
// User is no longer authenticated
```

#### `hasRole(role: UserRole | UserRole[]): Promise<boolean>`
Check if user has specific role(s).

```typescript
if (await hasRole('ADMIN')) {
  // User is admin
}

if (await hasRole(['ADMIN', 'COMPLIANCE_OFFICER'])) {
  // User is either admin or compliance officer
}
```

#### `isAdmin(): Promise<boolean>`
Convenience function to check admin status.

```typescript
if (await isAdmin()) {
  // Show admin-only features
}
```

### Server Action Functions (src/app/_actions/auth-actions.ts)

#### `loginAction(email: string, password: string): Promise<APIResponse<AuthUser>>`
Server action wrapper for login.

```typescript
const result = await loginAction('admin@liyali.com', 'password123')
if (result.success) {
  router.push('/workflows/dashboard')
}
```

#### `logoutAction(): Promise<APIResponse<null>>`
Server action wrapper for logout.

```typescript
await logoutAction()
router.push('/login')
```

#### `getCurrentUserAction(): Promise<APIResponse<AuthUser>>`
Server action to get current user.

```typescript
const result = await getCurrentUserAction()
if (result.success) {
  const user = result.data
}
```

#### `requireAuth(): Promise<AuthUser>`
Redirect to login if not authenticated.

```typescript
// In a server component
const user = await requireAuth()
// If not authenticated, automatically redirects to /login
```

#### `requireRole(allowedRoles: string[]): Promise<AuthUser>`
Redirect if user doesn't have required role.

```typescript
const user = await requireRole(['ADMIN', 'COMPLIANCE_OFFICER'])
// Redirects to /login if not authenticated
// Redirects to /workflows if role not in allowedRoles
```

## Protected Routes

### Public Routes
- `/login` - Login page (redirects to dashboard if already authenticated)
- `/api/*` - API routes (if any)

### Authenticated Routes
- `/workflows/*` - All workflow pages (requires authentication)
- `/admin/*` - Admin pages (requires ADMIN or COMPLIANCE_OFFICER role)
- `/compliance/*` - Compliance pages (requires ADMIN or COMPLIANCE_OFFICER role)
- `/monitoring/*` - Monitoring pages (requires ADMIN or COMPLIANCE_OFFICER role)

### Route Access Rules

| Route | Required Role | Redirect If Unauthorized |
|-------|---------------|--------------------------|
| `/login` | None | `/workflows/dashboard` |
| `/workflows/dashboard` | Any authenticated user | `/login` |
| `/workflows/search` | Any authenticated user | `/login` |
| `/workflows/requisitions/*` | Any authenticated user | `/login` |
| `/admin/reports` | ADMIN, COMPLIANCE_OFFICER | `/login` or `/workflows` |
| `/admin/users` | ADMIN | `/login` or `/workflows` |
| `/admin/logs` | ADMIN, COMPLIANCE_OFFICER | `/login` or `/workflows` |
| `/compliance/tracking` | ADMIN, COMPLIANCE_OFFICER | `/login` or `/workflows` |
| `/monitoring` | ADMIN, COMPLIANCE_OFFICER | `/login` or `/workflows` |
| `/verification/qr` | Any authenticated user | `/login` |

## Environment Variables

No environment variables are required for this authentication system. It uses:
- Built-in Next.js `cookies()` API for session storage
- No external auth providers
- No database connections

## Development vs Production

### Development
- Sessions stored in HTTP cookies (not secure)
- Demo users available
- All functionality enabled

### Production
- Sessions use Secure flag: `secure: process.env.NODE_ENV === 'production'`
- Would need to migrate to production user database
- Consider implementing:
  - Password hashing (bcrypt)
  - Database storage for users
  - Token refresh logic
  - Rate limiting on login endpoint
  - CSRF protection

## Migration to Production

To migrate to a production auth system:

1. **Database Setup**
   - Create users table
   - Add password hashing
   - Store user records

2. **Update Demo Users**
   - Replace `DEMO_USERS` with database query
   - Implement proper password hashing

3. **Add Security**
   - Implement rate limiting
   - Add CSRF protection
   - Use signed cookies
   - Add refresh tokens

4. **Example Production Code**
   ```typescript
   // Instead of DEMO_USERS map:
   const user = await db.user.findUnique({
     where: { email }
   })

   const isValid = await bcrypt.compare(password, user.passwordHash)
   ```

## Troubleshooting

### User Is Not Authenticated After Login
- Check that cookies are being set properly
- Verify cookie domain matches your application domain
- Check browser cookie settings (not blocking cookies)

### Session Expires Too Quickly
- Session timeout is 24 hours - check if clock is skewed
- Sessions expire when `expiresAt < Date.now()`
- Increase `SESSION_TIMEOUT` constant to extend duration

### Role-Based Access Not Working
- Verify user role is set in `DEMO_USERS`
- Check page permission logic: `if (!['ADMIN'].includes(user.role))`
- Ensure correct role name from `UserRole` type

### Cannot Log Out
- Check that `logout()` is being called before redirect
- Verify cookie is being deleted (use dev tools)
- Check redirect URL after logout

## Examples

### Check if User Is Admin (Client Component)
```typescript
'use client'
import { isAdminAction } from '@/app/_actions/auth-actions'
import { useEffect, useState } from 'react'

export function AdminFeature() {
  const [isAdmin, setIsAdmin] = useState(false)

  useEffect(() => {
    isAdminAction().then(setIsAdmin)
  }, [])

  if (!isAdmin) return null
  return <div>Admin content only</div>
}
```

### Protect a Page
```typescript
// src/app/admin/special/page.tsx
import { requireRole } from '@/auth'

export default async function SpecialPage() {
  const user = await requireRole(['ADMIN'])

  return (
    <div>
      <h1>Welcome, {user.name}</h1>
      <p>This page is only for admins</p>
    </div>
  )
}
```

### Create a Logout Button
```typescript
'use client'
import { logoutAction } from '@/app/_actions/auth-actions'
import { useRouter } from 'next/navigation'

export function LogoutButton() {
  const router = useRouter()

  const handleLogout = async () => {
    await logoutAction()
    router.push('/login')
  }

  return <button onClick={handleLogout}>Logout</button>
}
```

## Security Notes

⚠️ **This authentication system is for development and testing only.**

### Not Suitable for Production Because:
- Passwords are hardcoded in source code
- No password hashing or salting
- No rate limiting on login attempts
- Sessions stored in easily decodable Base64
- No CSRF protection
- No account lockout mechanisms

### For Production, Add:
1. **Database** - Store user credentials securely
2. **Password Hashing** - Use bcrypt or similar
3. **Rate Limiting** - Limit login attempts
4. **CSRF Protection** - Use tokens for state-changing operations
5. **Audit Logging** - Log authentication events
6. **Multi-factor Authentication** - Additional security layer
7. **Token Refresh** - Implement proper token lifecycle

## Support

For issues or questions about the authentication system:
1. Check the examples above
2. Review the source code in `src/lib/auth.ts`
3. Check page implementations in `/workflows`, `/admin`, etc.
4. Review error messages in the browser console
