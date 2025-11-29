# Screen Lock & Idle Timeout Integration Guide

## Overview

The screen-lock system provides **idle detection and auto-logout functionality** for enhanced security. It monitors user inactivity and automatically locks the screen after 5 minutes of idle time, with a 90-second countdown before automatic logout.

## Architecture

```
┌─────────────────────────────────────────────────────┐
│  Application (wrapped with IdleTimerContainer)      │
└──────────────────┬──────────────────────────────────┘
                   │
                   ▼
    ┌──────────────────────────────┐
    │  IdleTimerContainer          │
    │  (src/components/base/       │
    │   screen-lock.tsx)           │
    │                              │
    │ • useIdleTimer hook          │
    │ • Monitors user activity     │
    │ • Manages dialog state       │
    │ • BroadcastChannel sync      │
    └──────────────────────────────┘
           │                │
           │                └──────────────────────────┐
           │                                            │
           ▼                                            ▼
    ┌──────────────────┐                    ┌──────────────────┐
    │  ScreenLock      │                    │ Session System   │
    │  Component       │                    │ (session.ts)     │
    │                  │                    │                  │
    │ • Countdown      │                    │ • JWT tokens     │
    │   timer (90s)    │──────────────────→ │ • Cookies        │
    │ • Dialog UI      │ Server Actions     │ • Expiration     │
    │ • User buttons   │                    │ • Refresh logic  │
    └──────────────────┘                    └──────────────────┘
           │
           ▼
    ┌──────────────────────────────┐
    │  Server Actions              │
    │  (auth-actions.ts)           │
    │                              │
    │ • lockScreenOnUserIdle()     │
    │ • checkScreenLockState()     │
    │ • logUserOut()               │
    │ • getRefreshToken()          │
    └──────────────────────────────┘
```

## Configuration

### Session Timeouts (src/lib/session-config.ts)

```typescript
export const SESSION_CONFIG = {
  // User must interact within 5 minutes or screen locks
  IDLE_TIMEOUT: 5 * 60 * 1000,

  // Show screen lock for 90 seconds
  SCREEN_LOCK_COUNTDOWN: 90 * 1000,

  // Maximum session duration is 30 minutes
  SESSION_TTL: 30 * 60 * 1000,

  // Refresh token at 25 minutes
  TOKEN_REFRESH_INTERVAL: 25 * 60 * 1000
}
```

### Constants (src/lib/constants.ts)

```typescript
export const AUTH_SESSION = "__com.liyali-portal.com__"
export const USER_SESSION = "__com.liyali-user__"
export const PERMISSIONS_SESSION = "__com.liyali-pem__"
export const SCREEN_LOCK_SESSION = "__com.liyali-screen-lock__"
```

## How It Works

### 1. Idle Detection

```
User Activity Detected → Timer Resets
            ↓
    (No activity for 5 minutes)
            ↓
    Timer Expires → onIdle() Triggered
            ↓
    Screen Lock Dialog Opens
            ↓
    90-Second Countdown Starts
```

### 2. Screen Lock Sequence

```
Timeline:
0:00 - 5:00  User is active, timer tracks inactivity
5:00         Idle detected, dialog opens
5:00 - 5:90  Countdown timer (90 seconds)
6:30         Automatic logout if no user interaction
```

### 3. User Actions During Lock

**Option 1: "I'm still here"**
- Extends session by 30 minutes
- Clears screen lock cookie
- Resets idle timer
- Dialog closes

**Option 2: "Log Out"**
- Immediately logs out user
- Deletes all cookies
- Redirects to /login

**Option 3: No action**
- 90-second countdown completes
- Automatic logout
- Redirects to /login

## Implementation

### Step 1: Add to Root Layout

```typescript
// src/app/layout.tsx
import { IdleTimerContainer } from '@/components/base/screen-lock'
import { verifySession } from '@/lib/session'

export default async function RootLayout({ children }) {
  const { isAuthenticated, session } = await verifySession()

  return (
    <html>
      <body>
        {isAuthenticated && (
          <IdleTimerContainer session={session} />
        )}
        {children}
      </body>
    </html>
  )
}
```

### Step 2: Ensure Session Exists

The component requires:
- `AuthSession` type from `@/lib/types`
- JWT session cookies from `session.ts`
- Server actions from `auth-actions.ts`

### Step 3: Install Dependencies

The component uses:
- `react-idle-timer` - For idle detection
- `sonner` - For toast notifications
- `@/components/ui/dialog` - Dialog UI
- `BroadcastChannel API` - Multi-tab sync

## Components & Files

### Main Component: IdleTimerContainer

**File:** `src/components/base/screen-lock.tsx`

**Props:**
```typescript
interface IdleTimerContainerProps {
  session: AuthSession | null
}
```

**Features:**
- Automatic idle detection (5 minutes)
- Screen lock countdown (90 seconds)
- Multi-tab synchronization
- Persistent lock state across reloads
- Token refresh on activity
- Automatic logout on timeout

### Sub-component: ScreenLock

**Props:**
```typescript
interface ScreenLockProps {
  open: boolean
  onStillHere?: () => Promise<void>
  isLoading: boolean
  setIsLoading: React.Dispatch<React.SetStateAction<boolean>>
  handleUserLogOut: () => void
  hasLoggedOutRef: React.MutableRefObject<boolean>
}
```

## Server Actions

### lockScreenOnUserIdle(isLocked: boolean)

```typescript
import { lockScreenOnUserIdle } from '@/app/_actions/auth-actions'

// Lock screen
const success = await lockScreenOnUserIdle(true)

// Unlock screen
const success = await lockScreenOnUserIdle(false)
```

**What it does:**
- Sets/clears screen lock cookie
- Verifies the update was successful
- Returns true/false for success

### checkScreenLockState()

```typescript
import { checkScreenLockState } from '@/app/_actions/auth-actions'

const isLocked = await checkScreenLockState()
```

**What it does:**
- Reads screen lock cookie
- Checks if still valid (within 95 seconds)
- Returns boolean lock state

### logUserOut(reason: string)

```typescript
import { logUserOut } from '@/app/_actions/auth-actions'

const result = await logUserOut('Session expired')
```

**What it does:**
- Deletes all session cookies
- Clears JWT tokens
- Clears auth state
- Returns success response

### getRefreshToken()

```typescript
import { getRefreshToken } from '@/app/_actions/auth-actions'

const result = await getRefreshToken()
```

**What it does:**
- Extends session expiration (30 minutes)
- Updates JWT token
- Returns new expiration time

## Hooks

### useRefreshToken

**File:** `src/hooks/use-users-query-data.ts`

```typescript
import { useRefreshToken } from '@/hooks/use-users-query-data'

const { data, error, isLoading } = useRefreshToken(
  shouldRefresh = true,
  interval = 20 * 60 * 1000
)
```

**Parameters:**
- `shouldRefresh` - Whether to actively refresh (boolean)
- `interval` - Refresh interval in milliseconds (default: 20 minutes)

**Returns:**
- `data` - Token refresh response
- `error` - Error object if refresh fails
- `isLoading` - Loading state

### useLogger

**File:** `src/lib/logger.ts`

```typescript
import { useLogger } from '@/lib/logger'

const logger = useLogger('MyComponent')

logger.debug('Debug message', { extra: 'data' })
logger.info('Info message')
logger.warn('Warning message')
logger.error('Error message', error)
```

## Multi-Tab Synchronization

The screen-lock state syncs across browser tabs using:

1. **Primary:** BroadcastChannel API
   - Modern, efficient cross-tab communication
   - Works in most browsers

2. **Fallback:** localStorage events
   - Works in older browsers and Firefox private mode
   - Triggers `storage` event when updated

**How it works:**
```
Tab 1: User goes idle
  ↓
setIsDialogOpen(true)
  ↓
Broadcasts: { type: "SCREEN_LOCK_CHANGED", isLocked: true }
  ↓
Tab 2 & 3: Receive message
  ↓
setState("Idle")
setIsDialogOpen(true)
  ↓
All tabs show lock screen
```

## State Persistence

Screen lock state persists across page reloads:

```typescript
// On component mount
const isLocked = await checkScreenLockState()

if (isLocked) {
  // Lock screen is still valid
  setState('Idle')
  setIsDialogOpen(true)
}
```

The cookie expires in 95 seconds, accounting for:
- 90-second countdown (SCREEN_LOCK_COUNTDOWN)
- 5-second buffer for network/clock differences

## Session Refresh Flow

```
User Active & Idle Timer Running
        ↓
useRefreshToken() checks every 20 minutes
        ↓
getRefreshToken() called
        ↓
Session extended by 30 minutes
        ↓
New JWT token created
        ↓
Token refresh error?
  YES → Toast warning: "Session may be expiring"
  NO  → Silent refresh continues
```

## Error Handling

### Screen Lock Fails

```typescript
// Issue: lockScreenOnUserIdle(true) returns false
// Solution: Component shows modal REGARDLESS
//
// From screen-lock.tsx lines 444-451:
if (!lockSuccess) {
  logger.warn("Screen lock cookie not set, but showing modal anyway")
  // CONTINUE TO SHOW MODAL
  setIsDialogOpen(true)
}
```

### Token Refresh Fails

```typescript
// Issue: getRefreshToken() throws error
// Solution: Show warning toast to user
//
// From screen-lock.tsx lines 394-410:
if (refreshError) {
  toast.warning(
    "⚠️ Your session may be expiring. Please save your work...",
    { duration: 10000 }
  )
}
```

### Logout Fails

```typescript
// Issue: logUserOut() fails
// Solution: Force redirect to /login anyway
//
// From screen-lock.tsx lines 544-545:
window.location.replace("/login")
```

## Logging

All screen-lock events are logged for debugging:

```typescript
import { logger } from '@/lib/logger'

// Debug level (dev only)
logger.debug('Screen lock state changed', {
  component: 'IdleTimerContainer',
  isLocked: true
})

// Info level (always)
logger.info('✅ Screen lock activated successfully')

// Warning level
logger.warn('⚠️ Screen lock cookie missing during logout')

// Error level
logger.error('❌ Logout error', error, { component: 'IdleTimerContainer' })
```

### Log Output

```
2024-01-15T10:30:45.123Z DEBUG [IdleTimerContainer] Screen lock state changed
2024-01-15T10:30:45.124Z INFO ✅ Screen lock activated successfully
2024-01-15T10:31:14.456Z WARN ⚠️ Screen lock cookie missing during logout
2024-01-15T10:32:00.789Z ERROR ❌ Logout error [Error: Session deleted]
```

## Browser Compatibility

| Feature | Support | Fallback |
|---------|---------|----------|
| BroadcastChannel | Chrome, Firefox, Safari, Edge | localStorage |
| HTTP-Only Cookies | All modern browsers | Essential |
| useIdleTimer | All browsers (npm package) | None |
| localStorage | All browsers | BroadcastChannel |

## Performance Considerations

### Memory Usage
- Minimal: Event listeners cleanup on unmount
- Interval cleared when dialog closes
- No state updates if dialog not visible

### CPU Usage
- Idle detection: Throttled at 500ms
- Countdown: 1-second updates only during lock
- Token refresh: Every 20 minutes (background)

### Network
- Token refresh: 1 request per 20 minutes (background)
- Screen lock state: Cookie operations (fast)
- No polling if lock not needed

## Testing

### Test Idle Detection

```typescript
// Simulate idle timeout
jest.useFakeTimers()
jest.advanceTimersByTime(5 * 60 * 1000) // 5 minutes
expect(mockOnIdle).toHaveBeenCalled()
```

### Test Screen Lock UI

```typescript
render(<IdleTimerContainer session={mockSession} />)
jest.advanceTimersByTime(5 * 60 * 1000) // Trigger idle
expect(screen.getByText('Are you still there?')).toBeInTheDocument()
```

### Test Multi-Tab Sync

```typescript
// Simulate BroadcastChannel message
const channel = new BroadcastChannel('screen-lock-state')
channel.postMessage({ type: 'SCREEN_LOCK_CHANGED', isLocked: true })
// Verify other tabs receive message
```

## Common Issues & Solutions

### Issue: Dialog doesn't appear after idle

**Cause:** Component not wrapped around content
```typescript
// ❌ Wrong - IdleTimerContainer not used
<div>{children}</div>

// ✅ Correct - IdleTimerContainer wraps content
<IdleTimerContainer session={session}>
  {children}
</IdleTimerContainer>
```

**Solution:** Wrap entire app with `IdleTimerContainer` in root layout

### Issue: Screen lock not persisting across page reload

**Cause:** Cookie not being set
```typescript
// Check if setScreenLockCookie is called
const success = await lockScreenOnUserIdle(true)
console.log('Lock success:', success)
```

**Solution:** Verify JWT session is active before locking

### Issue: Multiple logout calls

**Cause:** Race condition in logout handler
```typescript
// ✅ Fixed with hasLoggedOutRef flag
if (hasLoggedOutRef.current) return
hasLoggedOutRef.current = true
```

### Issue: Countdown timer not accurate

**Cause:** Browser tab not in focus
```typescript
// Browser may throttle JavaScript execution
// Solution: Use server-side session expiration as backup
```

## Advanced Customization

### Change Idle Timeout

```typescript
// src/lib/session-config.ts
export const SESSION_CONFIG = {
  IDLE_TIMEOUT: 10 * 60 * 1000, // 10 minutes instead of 5
  SCREEN_LOCK_COUNTDOWN: 90 * 1000,
  SESSION_TTL: 30 * 60 * 1000,
  TOKEN_REFRESH_INTERVAL: 25 * 60 * 1000
}
```

### Change Lock Countdown

```typescript
// src/lib/session-config.ts
export const SESSION_CONFIG = {
  IDLE_TIMEOUT: 5 * 60 * 1000,
  SCREEN_LOCK_COUNTDOWN: 120 * 1000, // 2 minutes instead of 90s
  SESSION_TTL: 30 * 60 * 1000,
  TOKEN_REFRESH_INTERVAL: 25 * 60 * 1000
}
```

### Add Custom Toast Styling

```typescript
// screen-lock.tsx
toast.success("Session extended. Welcome back!", {
  position: 'top-center',
  duration: 5000,
  style: {
    background: '#22c55e',
    color: 'white'
  }
})
```

### Skip Idle Detection on Certain Routes

```typescript
// screen-lock.tsx lines 611-614
if (pathname.startsWith("/checkout")) return null
if (pathname.startsWith("/invoice")) return null
if (pathname.startsWith("/subscriptions")) return null
```

## Summary

The screen-lock system provides:
✅ Automatic idle detection (5 minutes)
✅ Secure lockscreen dialog
✅ Automatic logout (90 seconds)
✅ Session extension option
✅ Multi-tab synchronization
✅ Persistent state across reloads
✅ Comprehensive logging
✅ Production-ready error handling

The system integrates seamlessly with the JWT session system for enterprise-grade session management and security!
