# Session Management & Idle Timeout

Complete guide to Liyali Gateway's session management system with idle detection, screen lock, and automatic logout.

---

## Overview

The session management system ensures secure user sessions with:

- **1-hour maximum session duration** - Users must re-authenticate after 1 hour
- **30-minute idle timeout** - Screen locks after 30 minutes of inactivity
- **90-second recovery window** - Users can click "I'm still here" to extend session
- **Cross-tab synchronization** - Session state synced across browser tabs
- **Automatic token refresh** - Background refresh before expiry (55 minutes)

---

## Session Flow

```
User Login
    ↓
Session expires in 1 hour (SESSION_EXPIRY_TIME)
    ↓
[After 30 minutes of inactivity]
    ↓
Screen Lock Modal appears
User has 90 seconds to click "I'm still here"
    ↓
    ├─→ User clicks "I'm still here"
    │       ↓
    │   Session extends to 1 hour from that moment
    │   Idle timer resets
    │   Modal closes
    │
    └─→ 90 seconds pass without action
            ↓
        Automatic logout
        Session terminated
        Redirect to login page
```

---

## Configuration

**File**: `src/lib/session-config.ts`

```typescript
export const SESSION_CONFIG = {
  // Maximum session duration: 1 hour from login
  SESSION_EXPIRY_TIME: 1 * 60 * 60 * 1000,

  // Idle timeout: After 30 minutes of inactivity, show screen lock
  IDLE_TIMEOUT: 30 * 60 * 1000,

  // Screen lock countdown: User has 90 seconds to click "I'm still here"
  SCREEN_LOCK_COUNTDOWN: 90 * 1000,

  // Token refresh: Refresh session before expiry (55 minutes)
  TOKEN_REFRESH_INTERVAL: 55 * 60 * 1000,

  // Session TTL (backwards compatibility)
  SESSION_TTL: 1 * 60 * 60 * 1000,
};
```

### Adjustment Guide

To change timeout values, modify `SESSION_CONFIG` in `src/lib/session-config.ts`:

```typescript
// Example: Change to 2-hour sessions with 45-minute idle timeout
export const SESSION_CONFIG = {
  SESSION_EXPIRY_TIME: 2 * 60 * 60 * 1000,    // 2 hours
  IDLE_TIMEOUT: 45 * 60 * 1000,               // 45 minutes
  SCREEN_LOCK_COUNTDOWN: 120 * 1000,          // 2 minutes
  TOKEN_REFRESH_INTERVAL: 110 * 60 * 1000,    // 110 minutes
  SESSION_TTL: 2 * 60 * 60 * 1000,            // 2 hours
};
```

---

## Components & Hooks

### ScreenLock Component

**File**: `src/components/base/screen-lock.tsx`

Two-part component system:

#### 1. ScreenLock (Modal)

Displays the idle timeout warning with countdown timer.

**Props**:
```typescript
interface ScreenLockProps {
  open: boolean;                              // Modal visibility
  onStillHere?: () => Promise<void>;         // "I'm still here" callback
  isLoading: boolean;                         // Loading state during refresh
  setIsLoading: React.Dispatch<...>;          // Loading state setter
  handleUserLogOut: () => void;               // Logout handler
  hasLoggedOutRef: React.MutableRefObject;    // Prevent multiple logouts
}
```

**Features**:
- Circular countdown timer (SVG progress circle)
- Displays remaining seconds
- "I'm still here" button to extend session
- "Log Out" button for immediate logout
- Prevents close on backdrop click

#### 2. IdleTimerContainer (Wrapper)

Manages idle detection, screen lock state, and multi-tab synchronization.

**Key Features**:
- Uses `react-idle-timer` library for activity detection
- Tracks user inactivity across mouse, keyboard, and touch events
- Syncs lock state across browser tabs (BroadcastChannel + localStorage fallback)
- Persists lock state in cookies (survives page reload)
- Background token refresh when user is active

**Usage in Layout**:

```typescript
// In (private)/layout.tsx
async function MainNavProvider({
  children,
}: {
  children: React.ReactNode;
}) {
  // Server component - fetch session
  const { getSession } = await import("@/lib/auth");
  const session = await getSession();

  // Pass session to client component
  return <SessionProvider session={session}>{children}</SessionProvider>;
}

function SessionProvider({
  children,
  session,
}: {
  children: React.ReactNode;
  session: any;
}) {
  return (
    <>
      <IdleTimerContainer session={session} />
      {/* Rest of layout */}
    </>
  );
}
```

---

## Server Actions

**File**: `src/app/_actions/auth-actions.ts`

### lockScreenOnUserIdle(isLocked: boolean)

Set or clear the screen lock cookie.

```typescript
// Lock the screen
const success = await lockScreenOnUserIdle(true);

// Unlock the screen
const success = await lockScreenOnUserIdle(false);
```

**Returns**: `boolean` - Success status

**What it does**:
- Creates/deletes SCREEN_LOCK_SESSION cookie
- Cookie expires in 90 seconds
- Includes timestamp for verification

### checkScreenLockState()

Verify if screen lock cookie exists and is still valid.

```typescript
const isLocked = await checkScreenLockState();
if (isLocked) {
  // Show lock modal
}
```

**Returns**: `boolean` - True if locked, false if unlocked or expired

**Validation**:
- Checks cookie exists
- Verifies timestamp is not older than 95 seconds
- Returns false if cookie is missing or expired

### logUserOut(reason: string)

Terminate user session and clear all cookies.

```typescript
const response = await logUserOut("Session expired due to inactivity");

if (response.success) {
  window.location.replace("/login");
}
```

**Returns**: `APIResponse<null>`

**Clears**:
- AUTH_SESSION cookie
- USER_SESSION cookie
- PERMISSIONS_SESSION cookie
- SCREEN_LOCK_SESSION cookie

### getRefreshToken()

Refresh JWT token to extend session.

```typescript
const response = await getRefreshToken();

if (response.success) {
  console.log("Token refreshed");
  // Session extended
} else {
  console.error("Token refresh failed");
  // Session will expire
}
```

**Returns**: `APIResponse<any>`

**Side Effect**: Updates AUTH_SESSION cookie with new expiry time

---

## Hooks

### useRefreshToken()

**File**: `src/hooks/use-auth-queries.ts`

Background token refresh hook.

```typescript
const { data, error, isLoading } = useRefreshToken(
  shouldRefresh: boolean = true,
  interval?: number           // milliseconds (default: 20 minutes)
);
```

**Usage Example**:

```typescript
// In IdleTimerContainer
const { data: refreshData, error: refreshError, isLoading: isRefreshing } =
  useRefreshToken(
    Boolean(loggedIn && !isIdle)  // Only refresh when logged in and active
  );

// Handle refresh errors
useEffect(() => {
  if (refreshError) {
    toast.warning("Session may be expiring. Please save your work.");
  }
}, [refreshError]);
```

**Behavior**:
- Automatically calls `getRefreshToken()` at specified interval
- Pauses when `shouldRefresh` is false
- Updates background without user interaction
- Shows error state if refresh fails

---

## Activity Detection

The system tracks these user activities:

- Mouse movements
- Mouse clicks
- Keyboard input (all keys)
- Touch events
- Scroll events

Any of these activities resets the 30-minute idle timer.

**Non-activity Detection** (these don't count as activity):

- Page visibility changes
- Window focus/blur (can be enabled)
- Idle state is checked every 500ms (throttle rate)

---

## Multi-Tab Synchronization

When screen lock state changes, it's broadcast to other open tabs:

**Primary Method**: BroadcastChannel API
- Real-time, instant synchronization
- Works in all modern browsers
- Fails silently in private browsing mode

**Fallback Method**: localStorage
- Works when BroadcastChannel unavailable
- Triggered by storage events
- Slight delay (~100ms) compared to BroadcastChannel

**Behavior**:
```
Tab A: User idle 30 minutes
    ↓
Lock screen in Tab A
    ↓
Broadcast message: { type: "SCREEN_LOCK_CHANGED", isLocked: true }
    ↓
Tab B receives message
    ↓
Tab B also shows lock modal
    ↓
User clicks "I'm still here" in Tab A
    ↓
Unlock broadcast: { type: "SCREEN_LOCK_CHANGED", isLocked: false }
    ↓
Tab B automatically closes lock modal
```

---

## Cookie Structure

### AUTH_SESSION

Contains user authentication data and expiry time.

```
{
  accessToken: "token_user-007_1234567890",
  user_type: "ADMIN",
  user_id: "user-007",
  user: {
    id: "user-007",
    name: "Admin User",
    email: "admin@liyali.com",
    role: "ADMIN",
    department: "Administration"
  },
  expiresAt: "2025-11-30T15:30:00Z"
}
```

**Expiry**: 1 hour from creation (30 minutes for old sessions)
**Secure**: httpOnly, secure (production), sameSite=strict

### SCREEN_LOCK_SESSION

Contains lock state and timestamp.

```
{
  locked: true,
  timestamp: "2025-11-30T15:00:00.000Z"
}
```

**Expiry**: 90 seconds (auto-logout timer)
**Purpose**: Survives page reload, persists across tabs

---

## Error Handling

### Silent Failure Recovery

If token refresh fails, the system:
1. Shows warning toast to user
2. Continues session until expiry
3. Attempts refresh again at next interval
4. Logs errors for monitoring

```typescript
// In IdleTimerContainer
useEffect(() => {
  if (refreshError) {
    logger.error("Background token refresh failed - session may be expiring");
    toast.warning(
      "⚠️ Your session may be expiring. Please save your work and log back in if needed.",
      { duration: 10000 }
    );
  }
}, [refreshError]);
```

### Lock Screen Failures

Even if screen lock cookie creation fails:
1. Modal still displays to user
2. Countdown timer still runs
3. User can still click "I'm still here"
4. Logout proceeds after 90 seconds

This ensures users are never left without protection.

```typescript
// In IdleTimerContainer.onIdle()
let lockSuccess = false;
try {
  lockSuccess = await lockScreenOnUserIdle(true);
} catch (lockError) {
  logger.error("Exception while setting screen lock cookie");
  // Continue - we'll show modal even if cookie fails
}

// CRITICAL FIX: Show modal REGARDLESS of whether cookie was set
if (!lockSuccess) {
  logger.warn("Screen lock cookie not set, but showing modal anyway");
  // CONTINUE TO SHOW MODAL
}
setIsDialogOpen(true);
setState("Idle");
```

---

## Logging & Debugging

Comprehensive logging using `src/lib/logger.ts`:

```typescript
// Check logs in browser console:
logger.debug("Activity detected", { userId: "user-007" });
logger.info("Session extended", { newExpiry: "2025-11-30T15:30:00Z" });
logger.warn("Token refresh failed", { error: "Network error" });
logger.error("Logout failed", { reason: "Cookie deletion error" });
```

**Debug Features**:
- Session state logging on changes
- Token refresh progress tracking
- Multi-tab synchronization logging
- Lock/unlock operation logging
- Error conditions with full context

Enable debug mode in browser console:
```javascript
// Enable detailed logging
localStorage.setItem("DEBUG", "screen-lock:*");

// Disable
localStorage.removeItem("DEBUG");
```

---

## User Experience

### Idle Warning Modal

When user becomes idle:

1. **Modal appears** with countdown timer
2. **Visual feedback**:
   - Large countdown numbers (90, 89, 88...)
   - Circular progress indicator
   - Clear warning message
   - Two action buttons

3. **User can**:
   - Click "I'm still here" → session extends 1 hour
   - Click "Log Out" → immediate logout
   - Click anywhere on modal → closes (no action)

4. **After 90 seconds**:
   - If "I'm still here" not clicked → automatic logout
   - User redirected to login page
   - Session cookies cleared

### Toast Notifications

**Session Extended**:
```
✅ Session extended. Welcome back!
```

**Session Expired**:
```
❌ Session expired. Please log in again.
```

**Token Refresh Failed**:
```
⚠️ Your session may be expiring. Please save your work and log back in if needed.
```

---

## Testing Session Management

### Manual Testing Checklist

- [ ] Login → Session created with 1-hour expiry
- [ ] Be idle 30 minutes → Screen lock modal appears
- [ ] Click "I'm still here" → Session extends, modal closes
- [ ] Open two tabs → Idle in one → Both show lock modal
- [ ] Unlock in one tab → Other tab also unlocks
- [ ] Let countdown expire → Automatic logout
- [ ] Refresh page while locked → Lock state persists
- [ ] Token refresh → Background operation without interruption
- [ ] Network timeout during "I'm still here" → Error shown, can retry

### Testing Helpers

Temporarily change timeouts for testing:

```typescript
// In src/lib/session-config.ts for testing
export const SESSION_CONFIG = {
  SESSION_EXPIRY_TIME: 5 * 60 * 1000,         // 5 minutes (was 1 hour)
  IDLE_TIMEOUT: 2 * 60 * 1000,                // 2 minutes (was 30 minutes)
  SCREEN_LOCK_COUNTDOWN: 10 * 1000,           // 10 seconds (was 90 seconds)
  TOKEN_REFRESH_INTERVAL: 4 * 60 * 1000,      // 4 minutes (was 55 minutes)
  SESSION_TTL: 5 * 60 * 1000,                 // 5 minutes (was 1 hour)
};
```

---

## Security Considerations

### Session Hijacking Prevention

- **httpOnly cookies**: Prevent JavaScript access (XSS protection)
- **secure flag**: HTTPS only (production)
- **sameSite=strict**: Prevent CSRF attacks
- **JWT expiry**: Session expires after 1 hour max
- **Digital signature**: Cookie contents verified with HMAC-SHA256

### Idle Timeout Security

- **Activity monitoring**: Logout after 30 minutes inactivity
- **Recovery window**: 90 seconds to prove user presence
- **Screen lock**: Visible warning before logout
- **Cross-tab sync**: One compromised tab doesn't affect security

### Token Refresh Security

- **Automatic refresh**: Keeps session fresh without user action
- **Background operation**: Doesn't interrupt user workflow
- **Error handling**: Warns user if refresh fails
- **Expiry fallback**: Hard expiry at 1 hour (no infinite sessions)

---

## Troubleshooting

### Screen Lock Not Appearing

**Symptoms**: User becomes idle but modal doesn't show

**Debug Steps**:
1. Check browser console for errors
2. Verify `SESSION_CONFIG.IDLE_TIMEOUT` is not 0
3. Verify `IdleTimerContainer` is rendered in layout
4. Check if user is actually authenticated (`session` prop passed)
5. Look for console errors: `"Screen lock state changed"`

**Solution**:
- Ensure `useIdleTimer` dependency is installed: `npm install react-idle-timer`
- Check that `(private)` layout is properly updated
- Verify session object has `accessToken` field

### Session Expires Without Warning

**Symptoms**: User logged out without seeing lock modal

**Debug Steps**:
1. Check if SESSION_EXPIRY_TIME is too short
2. Verify idle timer is detecting activity
3. Check if lock modal appeared but not visible (z-index issue)
4. Look for token refresh errors in console

**Solution**:
- Increase `IDLE_TIMEOUT` or `SESSION_EXPIRY_TIME` in config
- Check if user activity is being detected (mouse, keyboard, etc.)
- Verify Dialog component is rendering on top of layout

### Lock State Not Syncing Across Tabs

**Symptoms**: Lock in one tab but not in another

**Debug Steps**:
1. Check if browser supports BroadcastChannel (most do)
2. Verify private browsing mode (BroadcastChannel disabled)
3. Check localStorage for `__SCREEN_LOCK_STATE__` key
4. Look for storage event listeners in console

**Solution**:
- System automatically falls back to localStorage if needed
- No action required - fallback is transparent
- Check browser privacy settings if unsure

### Session Not Extending After "I'm Still Here"

**Symptoms**: Clicked button but still logged out

**Debug Steps**:
1. Check `lockScreenOnUserIdle(false)` server action response
2. Verify `AUTH_SESSION` cookie was updated
3. Check for network errors in browser Network tab
4. Look for error logs in console: "Failed to unlock screen"

**Solution**:
- Check if getRefreshToken() is working
- Verify server actions are callable (not permission denied)
- Try fallback token refresh by checking getRefreshToken() directly

---

## Related Documentation

- [AUTH.md](./AUTH.md) - Authentication system overview
- [API.md](./API.md) - Server actions reference
- [SECURITY.md](./SECURITY.md) - Security best practices

---

**Last Updated**: 2025-11-30
**Version**: 1.0.0

---

## Quick Reference

| Feature | Configuration | Time |
|---------|---------------|------|
| Max session duration | SESSION_EXPIRY_TIME | 1 hour |
| Idle detection | IDLE_TIMEOUT | 30 minutes |
| Lock countdown | SCREEN_LOCK_COUNTDOWN | 90 seconds |
| Token refresh | TOKEN_REFRESH_INTERVAL | 55 minutes |
| Activity throttle | idleTimer.throttle | 500ms |
| Lock cookie expiry | setScreenLockCookie | 90 seconds |

