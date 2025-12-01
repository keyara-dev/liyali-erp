/**
 * Centralized session configuration
 * All timeout values are defined here for consistency and easy adjustment
 *
 * Session Flow:
 * 1. User logs in → Session expires in 1 hour (SESSION_EXPIRY_TIME)
 * 2. User idle for 30 minutes → Screen lock appears (IDLE_TIMEOUT)
 * 3. User has 90 seconds to click "I'm still here" (SCREEN_LOCK_COUNTDOWN)
 * 4. If clicked → Session extends to 1 hour from that moment
 * 5. If not clicked → Session terminates after 90 seconds
 */
export const SESSION_CONFIG = {
  // Maximum session duration: 1 hour from login
  SESSION_EXPIRY_TIME: 1 * 60 * 60 * 1000,

  // Idle timeout: After 30 minutes of inactivity, show screen lock
  IDLE_TIMEOUT: 30 * 60 * 1000,

  // Screen lock countdown: User has 90 seconds to click "I'm still here"
  SCREEN_LOCK_COUNTDOWN: 90 * 1000,

  // Token refresh: Refresh session before expiry
  TOKEN_REFRESH_INTERVAL: 55 * 60 * 1000, // 55 minutes (before 60-minute expiry)

  // Session TTL (for backwards compatibility): Maximum session duration
  SESSION_TTL: 1 * 60 * 60 * 1000,
} as const;

/**
 * Calculated constants derived from SESSION_CONFIG
 * Used for progress calculations and expiry time computation
 */

// ✅ Screen lock countdown in seconds (for progress circle calculation)
export const SCREEN_LOCK_COUNTDOWN_SECONDS = SESSION_CONFIG.SCREEN_LOCK_COUNTDOWN / 1000;

// ✅ SVG circular progress total (stroke dash array total)
export const PROGRESS_CIRCLE_TOTAL = 100.5;

// ✅ Session expiry time in milliseconds (used for cookie expiry)
export const SESSION_EXPIRY_MS = SESSION_CONFIG.SESSION_TTL;
