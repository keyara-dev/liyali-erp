"use client";
/**
 * Screen Lock Component
 *
 * IMPROVED: Multi-tab synchronization now only locks OTHER tabs when one tab idles
 * - Each tab has a unique ID (tabIdRef)
 * - When a tab locks, it broadcasts its ID
 * - Other tabs only apply the lock if it came from a DIFFERENT tab
 * - This prevents active tabs from being locked due to another tab's inactivity
 *
 * This solves the issue where having an unused tab open would lock your active work
 */
import { useEffect, useState, useRef, useCallback, useMemo } from "react";
import { useIdleTimer } from "react-idle-timer";

import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";

import { toast } from "sonner";
import {
  SESSION_CONFIG,
  SCREEN_LOCK_COUNTDOWN_SECONDS,
  PROGRESS_CIRCLE_TOTAL,
} from "@/lib/session-config";
import { logger } from "@/lib/logger";
import {
  checkScreenLockState,
  getRefreshToken,
  lockScreenOnUserIdle,
  logUserOut,
} from "@/app/_actions/auth";
import { AuthSession } from "@/types";
import { useTokenRefresh } from "@/hooks/use-auth-queries";

const DEFAULT_TIMEOUT = SESSION_CONFIG.SCREEN_LOCK_COUNTDOWN;
const SCREEN_LOCK_CHANNEL = "screen-lock-state";

interface ScreenLockProps {
  open: boolean;
  onStillHere?: () => Promise<void>;
  isLoading: boolean;
  handleUserLogOut: () => void;
  hasLoggedOutRef: React.RefObject<boolean>;
}

/**
 * Custom hook for countdown timer logic
 * Handles timer state, interval cleanup, and timeout callbacks
 *
 * IMPROVED: Only executes timeout if dialog is still open (prevents logout from hidden dialogs)
 * This prevents inactive tabs from logging out the user if their dialog wasn't actually shown
 */
const useCountdownTimer = (
  open: boolean,
  onTimeout: () => void,
  hasLoggedOutRef: React.RefObject<boolean>,
  timeoutSeconds: number = DEFAULT_TIMEOUT / 1000
) => {
  const [seconds, setSeconds] = useState(timeoutSeconds);
  const intervalRef = useRef<NodeJS.Timeout | null>(null);

  useEffect(() => {
    // Reset state and flag when dialog opens/closes
    if (open) {
      setSeconds(timeoutSeconds);
      if (hasLoggedOutRef.current) {
        hasLoggedOutRef.current = false;
      }
    } else {
      // Clear interval when dialog closes
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
        intervalRef.current = null;
      }
      return;
    }

    // Clear any existing interval before starting new one
    if (intervalRef.current) {
      clearInterval(intervalRef.current);
    }

    intervalRef.current = setInterval(() => {
      setSeconds((prevSeconds) => {
        const newSeconds = prevSeconds - 1;

        if (newSeconds <= 0) {
          if (intervalRef.current) {
            clearInterval(intervalRef.current);
            intervalRef.current = null;
          }
          return 0;
        }

        return newSeconds;
      });
    }, 1000);

    // ✅ Ensure cleanup on unmount
    return () => {
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
        intervalRef.current = null;
      }
    };
  }, [open, timeoutSeconds]);

  // ✅ IMPROVED: Separate effect to handle timeout when seconds reach 0
  // This ensures we check the current 'open' state before executing logout
  useEffect(() => {
    if (
      seconds <= 0 &&
      open &&
      hasLoggedOutRef.current !== undefined &&
      !hasLoggedOutRef.current
    ) {
      onTimeout();
    }
  }, [seconds, open, onTimeout, hasLoggedOutRef]);

  return seconds;
};

function ScreenLock({
  open,
  onStillHere,
  isLoading,
  handleUserLogOut,
  hasLoggedOutRef,
}: ScreenLockProps) {
  const seconds = useCountdownTimer(open, handleUserLogOut, hasLoggedOutRef);

  const handleRefreshAuthToken = useCallback(async () => {
    try {
      // Call the parent's onStillHere callback if provided
      if (onStillHere) {
        await onStillHere();
      } else {
        // Fallback to original behavior
        await lockScreenOnUserIdle(false);
      }
    } catch (error) {
      logger.error("Error in handleRefreshAuthToken", error, {
        component: "ScreenLock",
      });
      // Error is handled by parent's handleStillHere callback
      // This just ensures we don't throw unhandled errors
      throw error; // Re-throw for parent to handle
    }
  }, [onStillHere]);

  const progress = useMemo(
    () => (seconds / SCREEN_LOCK_COUNTDOWN_SECONDS) * PROGRESS_CIRCLE_TOTAL,
    [seconds]
  );

  return (
    <Dialog open={open}>
      <DialogContent className="sm:max-w-md" showCloseButton={false}>
        <DialogHeader>
          <DialogTitle>Are you still there?</DialogTitle>
          <DialogDescription>
            You have been idle for some time now, you will be logged out
            automatically in
          </DialogDescription>
        </DialogHeader>
        <div className="flex flex-col items-center justify-center gap-4 py-4">
          <div className="relative h-36 w-36">
            <svg
              className="h-full w-full"
              width="32"
              height="32"
              viewBox="0 0 36 36"
            >
              <circle
                className="stroke-slate-200 dark:stroke-slate-700"
                strokeWidth="4"
                fill="transparent"
                r="16"
                cx="18"
                cy="18"
              />
              <circle
                className="stroke-primary"
                strokeWidth="4"
                strokeDasharray={`${progress}, ${PROGRESS_CIRCLE_TOTAL}`}
                strokeLinecap="round"
                fill="transparent"
                r="16"
                cx="18"
                cy="18"
                style={{
                  transform: "rotate(-90deg)",
                  transformOrigin: "50% 50%",
                }}
              />
            </svg>
            <div className="absolute inset-0 flex flex-col items-center justify-center">
              <span className="text-5xl font-black">{seconds}</span>
              <span className="text-muted-foreground border-input/50 mt-1 rounded-full border p-2 py-1 text-xs font-medium">
                seconds
              </span>
            </div>
          </div>
        </div>
        <DialogFooter className="sm:justify-end">
          <Button
            variant="destructive"
            disabled={isLoading}
            onClick={handleUserLogOut}
          >
            Log Out
          </Button>
          <Button
            disabled={isLoading}
            isLoading={isLoading}
            onClick={handleRefreshAuthToken}
          >
            I'm still here
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

/**
 * Custom hook for multi-tab synchronization
 * Handles BroadcastChannel with localStorage fallback
 *
 * IMPROVED: Separates "this tab is idle" from "should show lock dialog"
 * - isIdle = THIS tab detected idle (only set by THIS tab's idle timer)
 * - isDialogOpen = Whether to show the dialog (can be from any tab)
 * - This prevents idle timers from being triggered by other tabs' lock events
 */
const useScreenLockSync = (loggedIn: boolean) => {
  const [isIdle, setIsIdle] = useState(false);
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const broadcastChannelRef = useRef<BroadcastChannel | null>(null);
  const tabIdRef = useRef(Math.random().toString(36).substring(7)); // Unique tab ID
  const thisTabInitiatedLock = useRef(false); // Track if THIS tab initiated the lock

  useEffect(() => {
    if (!loggedIn || typeof window === "undefined") return;

    let storageListener: ((e: StorageEvent) => void) | null = null;

    const syncState = (isLocked: boolean, sourceTabId?: string) => {
      // ✅ IMPROVED: Only lock dialog if the lock came from a DIFFERENT tab
      // This prevents active tabs from showing dialogs due to other tabs' inactivity
      const isFromOtherTab = sourceTabId && sourceTabId !== tabIdRef.current;

      logger.info("🔄 Screen lock state sync received", {
        component: "useScreenLockSync",
        isLocked,
        sourceTabId: sourceTabId || "unknown",
        currentTabId: tabIdRef.current,
        isFromOtherTab,
        willApplyDialogLock: isFromOtherTab || !sourceTabId,
        thisTabInitiatedLock: thisTabInitiatedLock.current,
      });

      // ✅ CRITICAL: Only show dialog if lock came from ANOTHER tab AND this tab didn't initiate it
      // Never apply isIdle=true from other tabs - only THIS tab's idle timer can set that
      if ((isFromOtherTab || !sourceTabId) && !isLocked) {
        // Clear dialog if other tab unlocked
        setIsDialogOpen(false);
      } else if (isFromOtherTab && isLocked && !thisTabInitiatedLock.current) {
        // Show dialog only if another tab locked AND this tab didn't initiate
        setIsDialogOpen(true);
      } else if (!sourceTabId && isLocked) {
        // Fallback: if no sourceTabId, show dialog (backward compatibility)
        setIsDialogOpen(true);
      } else {
        logger.debug("⏭️ Ignoring lock event (same tab or already locked)", {
          component: "useScreenLockSync",
          isFromOtherTab,
          thisTabInitiatedLock: thisTabInitiatedLock.current,
        });
      }
    };

    const handleBroadcastMessage = (event: MessageEvent) => {
      if (event.data.type === "SCREEN_LOCK_CHANGED") {
        syncState(event.data.isLocked, event.data.sourceTabId);
      }
    };

    const handleStorageChange = (event: StorageEvent) => {
      if (event.key === SCREEN_LOCK_CHANNEL) {
        try {
          const data = event.newValue ? JSON.parse(event.newValue) : null;
          if (data?.type === "SCREEN_LOCK_CHANGED") {
            syncState(data.isLocked, data.sourceTabId);
          }
        } catch (error) {
          logger.debug("Failed to parse storage event data", {
            component: "useScreenLockSync",
          });
        }
      }
    };

    try {
      // Try BroadcastChannel first
      broadcastChannelRef.current = new BroadcastChannel(SCREEN_LOCK_CHANNEL);
      broadcastChannelRef.current.addEventListener(
        "message",
        handleBroadcastMessage
      );
      logger.debug("✅ BroadcastChannel initialized for multi-tab sync", {
        component: "useScreenLockSync",
      });
    } catch (error) {
      logger.warn(
        "⚠️ BroadcastChannel not supported, using localStorage fallback for multi-tab sync",
        {
          component: "useScreenLockSync",
          error: (error as Error)?.message,
        }
      );
    }

    // Add localStorage listener as fallback (works even if BroadcastChannel fails)
    window.addEventListener("storage", handleStorageChange);
    storageListener = handleStorageChange;

    return () => {
      if (broadcastChannelRef.current) {
        broadcastChannelRef.current.removeEventListener(
          "message",
          handleBroadcastMessage
        );
        broadcastChannelRef.current.close();
        broadcastChannelRef.current = null;
      }
      if (storageListener) {
        window.removeEventListener("storage", storageListener);
        storageListener = null;
      }
    };
  }, [loggedIn]);

  const broadcastState = useCallback((isLocked: boolean) => {
    // ✅ CRITICAL: Track if THIS tab initiated the lock
    // This prevents our own lock message from triggering dialog on this tab again
    if (isLocked) {
      thisTabInitiatedLock.current = true;
    } else {
      thisTabInitiatedLock.current = false;
    }

    const message = {
      type: "SCREEN_LOCK_CHANGED",
      isLocked,
      sourceTabId: tabIdRef.current, // ✅ IMPROVED: Include source tab ID
      timestamp: Date.now(),
    };

    // Try BroadcastChannel first
    if (broadcastChannelRef.current) {
      try {
        broadcastChannelRef.current.postMessage(message);
        logger.debug(
          "📢 Broadcasted screen lock state change via BroadcastChannel",
          {
            component: "useScreenLockSync",
            isLocked,
            sourceTabId: tabIdRef.current,
            thisTabInitiatedLock: thisTabInitiatedLock.current,
            method: "BroadcastChannel",
          }
        );
        return;
      } catch (error) {
        logger.debug(
          "Failed to broadcast via BroadcastChannel, falling back to localStorage",
          {
            component: "useScreenLockSync",
            error: (error as Error)?.message,
          }
        );
      }
    }

    // Fallback: Use localStorage for browsers without BroadcastChannel
    try {
      localStorage.setItem(SCREEN_LOCK_CHANNEL, JSON.stringify(message));
      logger.debug("📢 Broadcasted screen lock state change via localStorage", {
        component: "useScreenLockSync",
        isLocked,
        sourceTabId: tabIdRef.current,
        thisTabInitiatedLock: thisTabInitiatedLock.current,
        method: "localStorage",
      });
    } catch (error) {
      logger.debug("Failed to broadcast via localStorage", {
        component: "useScreenLockSync",
        error: (error as Error)?.message,
      });
    }
  }, []);

  return { isIdle, isDialogOpen, setIsIdle, setIsDialogOpen, broadcastState };
};

export function IdleTimerContainer({
  session,
}: {
  session: AuthSession | null;
}) {
  const [isLoading, setIsLoading] = useState(false);
  const hasLoggedOutRef = useRef(false);

  const loggedIn = !!session?.access_token;

  // Use custom hook for multi-tab synchronization and state management
  const { isIdle, isDialogOpen, setIsIdle, setIsDialogOpen, broadcastState } =
    useScreenLockSync(loggedIn);

  /**
   * State Synchronization Pattern:
   * - isIdle: Indicates if user is in locked state (always synced with isDialogOpen locally)
   * - isDialogOpen: Controls dialog visibility and enables multi-tab broadcast
   * - broadcastState(): Broadcasts state changes to other tabs via BroadcastChannel/localStorage
   *
   * When idle: setIsIdle(true), setIsDialogOpen(true), broadcastState(true)
   * When unlocking: setIsIdle(false), setIsDialogOpen(false), broadcastState(false)
   */

  // ✅ Debug: Log when loggedIn status changes
  useEffect(() => {
    logger.debug("📋 IdleTimerContainer logged-in status", {
      component: "IdleTimerContainer",
      loggedIn,
      hasAccessToken: !!session?.access_token,
      session: session
        ? {
            user_id: (session as any)?.user_id,
            user_type: (session as any)?.user_type,
          }
        : null,
    });
  }, [loggedIn, session]);

  // ✅ Check for persisted lock state on mount (survives page reload)
  useEffect(() => {
    const checkPersistedLockState = async () => {
      try {
        const isLocked = await checkScreenLockState();
        logger.debug("🔍 Checking persisted lock state on mount", {
          component: "IdleTimerContainer",
          isLocked,
          loggedIn,
        });

        if (isLocked && loggedIn) {
          logger.info(
            "🔒 Screen lock state detected from cookie, restoring lock",
            {
              component: "IdleTimerContainer",
              isLocked,
            }
          );
          setIsIdle(true);
          setIsDialogOpen(true);
          broadcastState(true);
        } else if (!isLocked) {
          logger.debug("✅ No persisted lock state, starting fresh", {
            component: "IdleTimerContainer",
          });
        }
      } catch (error) {
        logger.error("❌ Error checking persisted lock state", error, {
          component: "IdleTimerContainer",
        });
      }
    };

    if (loggedIn) {
      checkPersistedLockState();
    }
  }, [loggedIn, setIsIdle, setIsDialogOpen, broadcastState]);

  // ✅ Handle background token refresh errors
  const { refreshError } = useTokenRefresh(Boolean(loggedIn && !isIdle));

  useEffect(() => {
    if (refreshError) {
      logger.error(
        "🔄 Background token refresh failed - session may be expiring",
        refreshError,
        {
          component: "IdleTimerContainer",
        }
      );
      toast.warning(
        "⚠️ Your session may be expiring. Please save your work and log back in if needed.",
        { duration: 10000 }
      );
    }
  }, [refreshError]);

  // ✅ Idle timeout callback - show modal regardless of cookie success
  const onIdle = useCallback(async () => {
    logger.debug("🔒 Idle timeout detected, attempting to lock screen", {
      component: "IdleTimerContainer.onIdle",
    });

    // ✅ CRITICAL: Update local state immediately to prevent multiple triggers
    setIsIdle(true);
    setIsDialogOpen(true);
    broadcastState(true);

    try {
      const lockSuccess = await lockScreenOnUserIdle(true);
      if (!lockSuccess) {
        logger.warn(
          "Screen lock cookie not set, but showing modal anyway (user requirement)",
          {
            component: "IdleTimerContainer.onIdle",
          }
        );
      } else {
        logger.info("✅ Screen lock activated successfully", {
          component: "IdleTimerContainer.onIdle",
        });
      }
    } catch (lockError) {
      logger.error(
        "Exception while setting screen lock cookie - will show modal anyway",
        lockError,
        {
          component: "IdleTimerContainer.onIdle",
        }
      );
    }
  }, [broadcastState, setIsIdle]);

  const onActive = useCallback(() => {
    // Don't reset idle state while dialog is open
    if (isIdle) return;
    idleTimer.reset();
  }, [isIdle]);

  const idleTimer = useIdleTimer({
    onIdle,
    onActive,
    timeout: SESSION_CONFIG.IDLE_TIMEOUT,
    throttle: 500,
    disabled: !loggedIn || isIdle,
  });

  const handleUserLogOut = useCallback(async () => {
    if (hasLoggedOutRef.current) return;
    hasLoggedOutRef.current = true;

    setIsLoading(true);
    // ✅ CRITICAL: Reset all idle-related state consistently
    setIsIdle(false);
    setIsDialogOpen(false);
    broadcastState(false);

    try {
      logger.info("🚪 Logging user out - session timed out", {
        component: "IdleTimerContainer.handleUserLogOut",
      });

      const response = await logUserOut();

      if (response.success) {
        logger.info("✅ Logout successful", {
          component: "IdleTimerContainer.handleUserLogOut",
        });
      } else {
        logger.warn(
          "⚠️ Logout response indicated failure, but proceeding with redirect",
          {
            component: "IdleTimerContainer.handleUserLogOut",
          }
        );
      }

      window.location.replace("/login");
    } catch (error) {
      logger.error("❌ Logout error", error, {
        component: "IdleTimerContainer.handleUserLogOut",
      });
      window.location.replace("/login");
    } finally {
      setIsLoading(false);
    }
  }, [broadcastState, setIsIdle]);

  const handleStillHere = useCallback(async () => {
    setIsLoading(true);
    try {
      logger.debug(
        "🔓 User clicked 'I'm still here' - attempting to unlock screen",
        {
          component: "IdleTimerContainer.handleStillHere",
        }
      );

      const success = await lockScreenOnUserIdle(false);

      if (success) {
        logger.info("✅ Screen unlocked and session refreshed", {
          component: "IdleTimerContainer.handleStillHere",
        });
        // ✅ CRITICAL: Reset all idle state when unlocking
        setIsIdle(false);
        setIsDialogOpen(false);
        broadcastState(false);
        idleTimer.reset();
        toast.success("Session extended. Welcome back!");
        return;
      }

      logger.warn(
        "Screen unlock returned false, attempting fallback token refresh",
        {
          component: "IdleTimerContainer.handleStillHere",
        }
      );

      const refreshResponse = await getRefreshToken();

      if (refreshResponse.success) {
        logger.info("✅ Fallback: Token refreshed successfully", {
          component: "IdleTimerContainer.handleStillHere",
        });
        // ✅ CRITICAL: Reset all idle state when unlocking
        setIsIdle(false);
        setIsDialogOpen(false);
        broadcastState(false);
        idleTimer.reset();
        toast.success("Session restored. You're all set!");
        return;
      }

      logger.error("Both unlock and refresh failed", {
        component: "IdleTimerContainer.handleStillHere",
      });

      toast.error("Session expired. Please log in again.");
      await handleUserLogOut();
    } catch (error) {
      logger.error("❌ Critical error in handleStillHere", error, {
        component: "IdleTimerContainer.handleStillHere",
      });
      toast.error("An unexpected error occurred. Logging out...");
      await handleUserLogOut();
    } finally {
      setIsLoading(false);
    }
  }, [idleTimer, handleUserLogOut, broadcastState, setIsIdle]);

  // Debug logging for state changes
  useEffect(() => {
    logger.debug("🔍 Screen lock state changed", {
      component: "IdleTimerContainer.render",
      isIdle,
      isDialogOpen,
      loggedIn,
    });
  }, [isIdle, isDialogOpen, loggedIn]);

  // Render the ScreenLock component when dialog should be open
  if (!isDialogOpen) return null;

  return (
    <ScreenLock
      open={isDialogOpen}
      onStillHere={handleStillHere}
      isLoading={isLoading}
      handleUserLogOut={handleUserLogOut}
      hasLoggedOutRef={hasLoggedOutRef}
    />
  );
}

export default ScreenLock;
