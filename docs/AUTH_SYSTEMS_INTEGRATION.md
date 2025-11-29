# Authentication Systems Integration Guide

## Overview

The Liyali Gateway project now has **TWO complementary authentication systems**:

1. **Simulated Auth System** (`src/lib/auth.ts`)
   - Simple demo user validation
   - Base64-encoded cookie sessions
   - Lightweight, perfect for development/testing
   - No external dependencies

2. **JWT-Based Session System** (`src/lib/session.ts`)
   - Encrypted JWT tokens
   - Advanced session features (screen lock, permissions)
   - Production-ready approach
   - More sophisticated token handling

Both systems work **together seamlessly** through the integration layer (`src/lib/auth-integration.ts`).

## System Architecture

```
┌────────────────────────────────────────────────────────────────┐
│                    Application Code                             │
│  (Pages, Server Components, API Routes)                        │
└────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌────────────────────────────────────────────────────────────────┐
│                 Unified Auth Integration                        │
│          (src/lib/auth-integration.ts)                         │
│                                                                │
│  • unifiedLogin()                                             │
│  • unifiedLogout()                                            │
│  • unifiedGetCurrentUser()                                    │
│  • unifiedIsAuthenticated()                                   │
│  • validateAuthSession()                                      │
│  • syncUserSession()                                          │
└────────────────────────────────────────────────────────────────┘
              │                                    │
              ▼                                    ▼
   ┌──────────────────────┐          ┌──────────────────────┐
   │  Simulated Auth      │          │  JWT Session System  │
   │  (src/lib/auth.ts)   │          │ (src/lib/session.ts) │
   │                      │          │                      │
   │ • Demo users (7)     │          │ • Encrypted tokens   │
   │ • Email/password     │          │ • Multiple sessions  │
   │ • Role checking      │          │ • Screen lock        │
   │ • Base64 cookies     │          │ • Permissions        │
   └──────────────────────┘          └──────────────────────┘
              │                                    │
              ▼                                    ▼
   ┌──────────────────────┐          ┌──────────────────────┐
   │  Demo User Store     │          │  Encrypted Cookies   │
   │  (In Memory)         │          │  (HTTP-Only)         │
   │                      │          │                      │
   │ DEMO_USERS map       │          │ AUTH_SESSION         │
   │ 7 user accounts      │          │ USER_SESSION         │
   │ with passwords       │          │ PERMISSIONS_SESSION  │
   │                      │          │ SCREEN_LOCK_SESSION  │
   └──────────────────────┘          └──────────────────────┘
```

## When to Use Each System

### Use Simulated Auth (src/lib/auth.ts) for:
- ✅ Testing during development
- ✅ Demo accounts with quick access
- ✅ Role-based testing without setup
- ✅ Understanding auth flow
- ✅ Quick prototyping

**Functions:**
```typescript
import { getCurrentUser, login, logout, hasRole, isAdmin } from '@/auth'

const user = await getCurrentUser()
const result = await login('admin@liyali.com', 'password123')
await logout()
```

### Use JWT Session System (src/lib/session.ts) for:
- ✅ Production-grade authentication
- ✅ Advanced session management
- ✅ Screen lock/idle timeout
- ✅ Permissions system
- ✅ Multiple session tracking

**Functions:**
```typescript
import {
  createAuthSession,
  verifySession,
  deleteSession,
  createUserSession,
  getScreenLockState,
  setScreenLockCookie
} from '@/lib/session'

const { isAuthenticated, session } = await verifySession()
await setScreenLockCookie(true)
```

### Use Unified Integration (src/lib/auth-integration.ts) for:
- ✅ Seamless login/logout across both systems
- ✅ Checking authentication status
- ✅ Getting current user (tries both systems)
- ✅ Creating sessions that work everywhere

**Functions:**
```typescript
import {
  unifiedLogin,
  unifiedLogout,
  unifiedGetCurrentUser,
  unifiedIsAuthenticated,
  validateAuthSession
} from '@/lib/auth-integration'

const result = await unifiedLogin(email, password)
// Creates both simulated and JWT sessions
```

## Data Flow Diagram

### Login Flow (Unified)

```
User Submits Form
    │
    ▼
unifiedLogin(email, password)
    │
    ├─→ Simulated Auth: Validate credentials
    │   (Check DEMO_USERS)
    │
    ├─→ If valid, create JWT session
    │   • generateAccessToken()
    │   • createAuthSession()
    │   • createUserSession()
    │
    └─→ Return { success: true, user }
        │
        ▼
    Both systems now have active sessions:
    • auth_session (JWT encrypted)
    • user_session (JWT encrypted)
    • user data in memory
```

### Logout Flow (Unified)

```
User Clicks Logout
    │
    ▼
unifiedLogout()
    │
    ├─→ Delete all JWT cookies
    │   • AUTH_SESSION
    │   • USER_SESSION
    │   • PERMISSIONS_SESSION
    │   • SCREEN_LOCK_SESSION
    │
    ├─→ Clear simulated auth state
    │
    └─→ Redirect to /login
```

### Authentication Check (Unified)

```
Protected Page Load
    │
    ▼
unifiedGetCurrentUser()
    │
    ├─→ Try JWT verification first
    │   const { isAuthenticated, session } = await verifySession()
    │
    ├─→ If valid JWT session found
    │   return user
    │
    └─→ If not, fall back to simulated auth
        const user = await getCurrentUser()
        return user
```

## Type System

### Core Types (src/lib/types/index.ts)

```typescript
export type UserType =
  | 'REQUESTER'
  | 'DEPARTMENT_MANAGER'
  | 'FINANCE_OFFICER'
  | 'DIRECTOR'
  | 'CFO'
  | 'COMPLIANCE_OFFICER'
  | 'ADMIN'

export interface User {
  id: string
  name: string
  email: string
  role: UserType
  department?: string
  avatar?: string
  user_type?: UserType
  expiresAt?: Date | string
}

export interface AuthSession {
  accessToken: string
  user_type?: UserType
  user_id?: string
  user?: Partial<User>
  change_password?: boolean
  mfa_required?: boolean
  organization_id?: string
  expiresAt?: Date | string
  permissions?: Permission[]
}
```

## Session Management

### Multiple Session Types

The system manages **4 types of cookies** simultaneously:

```
┌──────────────────────────────────────────────────┐
│ Session Cookies (HTTP-Only)                      │
├──────────────────────────────────────────────────┤
│                                                  │
│ 1. AUTH_SESSION                                  │
│    Contains: accessToken, user_type, user_id     │
│    Expires: 30 minutes                           │
│    Purpose: Authentication verification         │
│                                                  │
│ 2. USER_SESSION                                  │
│    Contains: User object (name, email, role)     │
│    Expires: 1 hour                               │
│    Purpose: Quick access to user info           │
│                                                  │
│ 3. PERMISSIONS_SESSION                          │
│    Contains: User permissions array              │
│    Expires: 1 hour                               │
│    Purpose: Authorization checks                │
│                                                  │
│ 4. SCREEN_LOCK_SESSION                          │
│    Contains: Lock state, timestamp               │
│    Expires: 90 seconds                           │
│    Purpose: Idle timeout screen lock           │
│                                                  │
└──────────────────────────────────────────────────┘
```

## Session Lifecycle

### Creation

```typescript
// User logs in
const result = await unifiedLogin(email, password)

// Creates:
// 1. Simulated auth state (in memory)
// 2. AUTH_SESSION cookie (JWT encrypted, 30 min)
// 3. USER_SESSION cookie (JWT encrypted, 1 hour)
// 4. PERMISSIONS_SESSION cookie (JWT encrypted, 1 hour)
```

### Validation

```typescript
// On each request/page load
const user = await unifiedGetCurrentUser()

// Checks:
// 1. AUTH_SESSION - Is auth valid?
// 2. USER_SESSION - Is user data available?
// 3. Simulated auth - Fallback if JWT fails
```

### Refresh

```typescript
// Sessions auto-refresh on activity
await updateAuthSession({
  // New fields
})

// Creates new JWT token with 30-minute expiration
```

### Termination

```typescript
// User logs out
await unifiedLogout()

// Deletes:
// 1. AUTH_SESSION cookie
// 2. USER_SESSION cookie
// 3. PERMISSIONS_SESSION cookie
// 4. SCREEN_LOCK_SESSION cookie
// 5. Simulated auth state
```

## Demo Users

All demo users work with both systems:

```
Email: requester@liyali.com
Role: REQUESTER
Password: password123
Session Type: Both (simulated + JWT)

Email: admin@liyali.com
Role: ADMIN
Password: password123
Session Type: Both (simulated + JWT)

... (5 more users available)
```

## Configuration

### Session Timeout Configuration

File: `src/lib/session-config.ts`

```typescript
export const SESSION_CONFIG = {
  // User must interact within 5 minutes
  IDLE_TIMEOUT: 5 * 60 * 1000,

  // Show screen lock for 90 seconds
  SCREEN_LOCK_COUNTDOWN: 90 * 1000,

  // Maximum session duration 30 minutes
  SESSION_TTL: 30 * 60 * 1000,

  // Refresh token at 25 minutes
  TOKEN_REFRESH_INTERVAL: 25 * 60 * 1000
}
```

### Constants

File: `src/lib/constants.ts`

```typescript
export const AUTH_SESSION = "__com.liyali-portal.com__"
export const USER_SESSION = "__com.liyali-user__"
export const PERMISSIONS_SESSION = "__com.liyali-pem__"
export const SCREEN_LOCK_SESSION = "__com.liyali-screen-lock__"
```

## Usage Examples

### Example 1: Server Component Authentication Check

```typescript
// src/app/protected/page.tsx
import { unifiedGetCurrentUser } from '@/lib/auth-integration'
import { redirect } from 'next/navigation'

export default async function ProtectedPage() {
  const user = await unifiedGetCurrentUser()

  if (!user) {
    redirect('/login')
  }

  return (
    <div>
      <h1>Welcome, {user.name}</h1>
      <p>Role: {user.role}</p>
    </div>
  )
}
```

### Example 2: Login Page Using Unified System

```typescript
// src/app/login/_components/login-form.tsx
'use client'

import { unifiedLogin } from '@/lib/auth-integration'
import { useRouter } from 'next/navigation'

export function LoginForm() {
  const router = useRouter()

  const handleSubmit = async (email: string, password: string) => {
    const result = await unifiedLogin(email, password)

    if (result.success) {
      // Both systems created sessions automatically
      router.push('/workflows/dashboard')
    } else {
      alert(result.error)
    }
  }

  // ... form code
}
```

### Example 3: Session Validation in API Route

```typescript
// src/app/api/protected/route.ts
import { validateAuthSession } from '@/lib/auth-integration'

export async function GET() {
  const { isValid, user, error } = await validateAuthSession()

  if (!isValid) {
    return Response.json({ error }, { status: 401 })
  }

  return Response.json({
    message: `Hello, ${user?.name}`,
    user
  })
}
```

### Example 4: Role-Based Access

```typescript
// src/app/admin/page.tsx
import { unifiedGetCurrentUser } from '@/lib/auth-integration'
import { redirect } from 'next/navigation'

export default async function AdminPage() {
  const user = await unifiedGetCurrentUser()

  if (!user) {
    redirect('/login')
  }

  if (user.role !== 'ADMIN') {
    redirect('/workflows')
  }

  return <div>Admin Dashboard</div>
}
```

## Environment Variables Required

```bash
# For JWT token encryption (must be at least 32 characters)
AUTH_SECRET=your_long_secure_secret_key_here_at_least_32_chars

# For session configuration (optional, uses defaults)
NODE_ENV=development
```

## Security Considerations

### What's Secure

✅ HTTP-Only cookies (prevents XSS)
✅ Encrypted JWT tokens (tampering protection)
✅ SameSite attribute (CSRF protection)
✅ Automatic session expiration
✅ Screen lock on idle
✅ Secure flag in production

### What's NOT Secure (Development Only)

⚠️ Demo passwords hardcoded
⚠️ No password hashing (demo only)
⚠️ Demo users visible in source code
⚠️ Base64 encoding (not encryption in simulated auth)

### For Production

When deploying to production:

1. **Replace Demo Users**
   ```typescript
   // Instead of DEMO_USERS map:
   const user = await db.user.findUnique({ where: { email } })
   const isValid = await bcrypt.compare(password, user.passwordHash)
   ```

2. **Use Real Database**
   - Store user credentials securely
   - Use bcrypt/argon2 for password hashing
   - Store sessions in database (optional)

3. **Enable Secure Flags**
   - Set `secure: true` for HTTPS
   - Set `sameSite: 'strict'`
   - Use strong AUTH_SECRET

4. **Add Rate Limiting**
   - Limit login attempts
   - Implement CAPTCHA
   - Monitor for brute force attacks

## Troubleshooting

### Session Not Persisting

**Problem:** User logs in but session is lost on page reload

**Solution:**
1. Check cookies are enabled in browser
2. Verify AUTH_SECRET is set in .env
3. Check cookie domain matches
4. Clear browser cookies and try again

### JWT Token Expired

**Problem:** Getting "Token expired" error

**Solution:**
1. Session TTL is 30 minutes (configured in session-config.ts)
2. User needs to log in again after expiration
3. Increase SESSION_CONFIG.SESSION_TTL if needed
4. Check server/client time sync

### Simulated Auth Not Working

**Problem:** Login fails even with correct credentials

**Solution:**
1. Use exact email: `admin@liyali.com` (lowercase)
2. Password must be exactly: `password123`
3. Check console for error messages
4. Verify DEMO_USERS map in src/lib/auth.ts

### Screen Lock Not Showing

**Problem:** Idle screen lock not appearing

**Solution:**
1. Ensure user doesn't interact for 5 minutes (IDLE_TIMEOUT)
2. Check SCREEN_LOCK_SESSION cookie exists
3. Verify screen lock component is mounted
4. Check browser console for errors

## Migration Path to Production

### Phase 1: Add Database Users
1. Create users table
2. Implement password hashing
3. Update login logic to use database
4. Keep JWT session system (proven secure)

### Phase 2: Enterprise Features
1. Add permissions system
2. Implement role-based authorization
3. Add audit logging
4. Add multi-factor authentication

### Phase 3: Advanced Security
1. Implement token refresh mechanism
2. Add account lockout
3. Add suspicious activity detection
4. Implement password policies

## API Reference

### src/lib/auth-integration.ts

```typescript
// Login - works with both systems
unifiedLogin(email: string, password: string): Promise<{
  success: boolean
  user?: User
  error?: string
}>

// Logout - clears all sessions
unifiedLogout(): Promise<void>

// Get current user - checks both systems
unifiedGetCurrentUser(): Promise<User | null>

// Check authentication status
unifiedIsAuthenticated(): Promise<boolean>

// Validate session with error handling
validateAuthSession(): Promise<{
  isValid: boolean
  user: User | null
  error?: string
}>

// Keep sessions in sync
syncUserSession(user: User): Promise<void>
```

## Key Files Structure

```
src/
├── lib/
│   ├── auth.ts                    (Simulated auth)
│   ├── session.ts                 (JWT sessions)
│   ├── auth-integration.ts        (Unified API)
│   ├── session-config.ts          (Configuration)
│   ├── constants.ts               (Cookie names, etc.)
│   └── types/
│       └── index.ts               (Type definitions)
│
├── auth.ts                        (Public exports)
│
├── app/
│   ├── login/
│   │   ├── page.tsx
│   │   └── _components/
│   │       └── login-form.tsx
│   │
│   ├── _actions/
│   │   ├── auth.ts                (Old, compatibility)
│   │   └── auth-actions.ts        (New)
│   │
│   └── (protected pages)
│       └── page.tsx               (All use unified system)
```

---

## Summary

The dual authentication system provides:

✅ **Flexibility**: Choose between simple (simulated) or advanced (JWT)
✅ **Compatibility**: Both systems work together seamlessly
✅ **Security**: Multiple layers of protection
✅ **Scalability**: Easy to migrate to production
✅ **Development Speed**: Quick testing with demo users
✅ **Enterprise Ready**: JWT system scales to production

Start with the unified API (`auth-integration.ts`) and let it handle the complexity behind the scenes!
