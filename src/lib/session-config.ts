/**
 * Centralized session configuration
 * All timeout values are defined here for consistency and easy adjustment
 */
export const SESSION_CONFIG = {
  // Idle detection: User must interact within 30 minutes or screen locks
  IDLE_TIMEOUT: 30 * 60 * 1000,

  // Screen lock countdown: User has 90 seconds to click "I'm still here"
  SCREEN_LOCK_COUNTDOWN: 90 * 1000,

  // Session TTL: Maximum session duration is 24 hours
  SESSION_TTL: 24 * 60 * 60 * 1000,

  // Token refresh: Refresh every 23 hours (before 24-hour expiry)
  TOKEN_REFRESH_INTERVAL: 23 * 60 * 60 * 1000
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
