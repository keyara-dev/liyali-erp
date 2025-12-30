/**
 * Centralized session configuration
 * All timeout values are defined here for consistency and easy adjustment
 *
 * Session Flow:
 * 1. User logs in → Session expires based on backend's expiresIn value (default: 1 hour)
 * 2. User idle for 30 minutes → Screen lock appears (IDLE_TIMEOUT)
 * 3. User has 90 seconds to click "I'm still here" (SCREEN_LOCK_COUNTDOWN)
 * 4. If clicked → Session extends based on refresh token response
 * 5. If not clicked → Session terminates after 90 seconds
 */
export const SESSION_CONFIG = {
  // Maximum session duration: 1 hour from login (fallback if backend doesn't provide expiresIn)
  SESSION_EXPIRY_TIME: 1 * 60 * 60 * 1000,

  // Idle timeout: After 30 minutes of inactivity, show screen lock
  IDLE_TIMEOUT: 30 * 60 * 1000,

  // Screen lock countdown: User has 90 seconds to click "I'm still here"
  SCREEN_LOCK_COUNTDOWN: 90 * 1000,

  // Token refresh: Refresh session before expiry (5 minutes before expiration)
  TOKEN_REFRESH_BUFFER: 5 * 60 * 1000, // 5 minutes buffer

  // Session TTL (for backwards compatibility): Maximum session duration
  SESSION_TTL: 1 * 60 * 60 * 1000,
} as const;

/**
 * Calculate token refresh interval based on backend's expiresIn value
 * @param expiresInSeconds - Expiration time in seconds from backend
 * @returns Refresh interval in milliseconds (5 minutes before expiration)
 */
export function calculateRefreshInterval(expiresInSeconds: number): number {
  const expirationMs = expiresInSeconds * 1000;
  return Math.max(expirationMs - SESSION_CONFIG.TOKEN_REFRESH_BUFFER, 60 * 1000); // At least 1 minute
}

/**
 * Check if a token should be refreshed based on its expiration time
 * @param expiresAt - Token expiration date
 * @param bufferMs - Buffer time in milliseconds (default: 5 minutes)
 * @returns true if token should be refreshed
 */
export function shouldRefreshToken(expiresAt: Date | string, bufferMs: number = SESSION_CONFIG.TOKEN_REFRESH_BUFFER): boolean {
  const expiration = typeof expiresAt === 'string' ? new Date(expiresAt) : expiresAt;
  const now = new Date();
  const timeUntilExpiry = expiration.getTime() - now.getTime();
  return timeUntilExpiry <= bufferMs;
}

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
