"use client";
import { usePathname } from "next/navigation";
import { useEffect, useState, useRef, useCallback } from "react";
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
import { useRefreshToken } from "@/hooks/use-users-query-data";
import {
  lockScreenOnUserIdle,
  logUserOut,
  getRefreshToken,
  checkScreenLockState,
} from "@/app/_actions/auth-actions";
import { AuthSession } from "@/types";
import { toast } from "sonner";
import {
  SESSION_CONFIG,
  SCREEN_LOCK_COUNTDOWN_SECONDS,
  PROGRESS_CIRCLE_TOTAL,
} from "@/lib/session-config";
import { logger } from "@/lib/logger";

const DEFAULT_TIMEOUT = SESSION_CONFIG.SCREEN_LOCK_COUNTDOWN;

interface ScreenLockProps {
  open: boolean;
  onStillHere?: () => Promise<void>;
  isLoading: boolean;
  setIsLoading: React.Dispatch<React.SetStateAction<boolean>>;
  handleUserLogOut: () => void;
  hasLoggedOutRef: React.MutableRefObject<boolean>;
}

function ScreenLock({
  open,
  onStillHere,
  isLoading,
  setIsLoading,
  handleUserLogOut,
  hasLoggedOutRef,
}: ScreenLockProps) {
  const [seconds, setSeconds] = useState(DEFAULT_TIMEOUT / 1000); // REMAINING SECONDS

  const intervalRef = useRef<NodeJS.Timeout | null>(null);

  // Reset countdown when dialog opens or closes
  useEffect(() => {
    if (open) {
      setSeconds(DEFAULT_TIMEOUT / 1000);
      hasLoggedOutRef.current = false;
    } else {
      // Clear interval when dialog closes
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
        intervalRef.current = null;
      }
    }
  }, [open]);

  const handleRefreshAuthToken = useCallback(async () => {
    setIsLoading(true);

    // Call the parent's onStillHere callback if provided
    if (onStillHere) {
      await onStillHere();
    } else {
      // Fallback to original behavior
      await lockScreenOnUserIdle(false);
    }

    setIsLoading(false);
  }, [onStillHere]);

  useEffect(() => {
    if (!open) {
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
        intervalRef.current = null;
      }
      return;
    }

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
          // ✅ Use a flag to prevent calling after unmount
          if (!hasLoggedOutRef.current) {
            handleUserLogOut();
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
  }, [open, handleUserLogOut, hasLoggedOutRef]);

  return (
    <Dialog open={open}>
      <DialogContent className="sm:max-w-md" hideCloseButton>
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
                strokeDasharray={`${(seconds / SCREEN_LOCK_COUNTDOWN_SECONDS) * PROGRESS_CIRCLE_TOTAL}, ${PROGRESS_CIRCLE_TOTAL}`}
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
                {"seconds"}
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

export function IdleTimerContainer({
  session,
}: {
  session: AuthSession | null;
}) {
  const pathname = usePathname();
  const [isLoading, setIsLoading] = useState(false);
  const hasLoggedOutRef = useRef(false);
  const [state, setState] = useState("Active");
  const [count, setCount] = useState(0);
  const broadcastChannelRef = useRef<BroadcastChannel | null>(null);

  // Separate state for dialog rendering - ensures dialog opens immediately
  const [isDialogOpen, setIsDialogOpen] = useState(false);

  const loggedIn = session?.accessToken || false;
  const isIdle = state === "Idle";

  // ✅ Debug: Log when loggedIn status changes
  useEffect(() => {
    logger.debug("📋 IdleTimerContainer logged-in status", {
      component: "IdleTimerContainer",
      loggedIn,
      hasAccessToken: !!session?.accessToken,
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
          setState("Idle");
          // Immediately open dialog instead of waiting for state sync
          setIsDialogOpen(true);
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
  }, [loggedIn]);

  // ✅ Setup BroadcastChannel for multi-tab synchronization with localStorage fallback
  useEffect(() => {
    if (!loggedIn || typeof window === "undefined") return;

    let storageListenerCleanup: (() => void) | null = null;

    try {
      // Create BroadcastChannel for cross-tab communication
      broadcastChannelRef.current = new BroadcastChannel("screen-lock-state");

      // Listen for messages from other tabs
      const handleMessage = (event: MessageEvent) => {
        if (event.data.type === "SCREEN_LOCK_CHANGED") {
          const { isLocked } = event.data;
          logger.info("🔄 Screen lock state changed in another tab, syncing", {
            component: "IdleTimerContainer",
            isLocked,
            method: "BroadcastChannel",
          });
          setState(isLocked ? "Idle" : "Active");
          setIsDialogOpen(isLocked);
        }
      };

      broadcastChannelRef.current.addEventListener("message", handleMessage);

      // ✅ CRITICAL FIX: Add localStorage fallback for browsers that don't support BroadcastChannel
      // Browsers like Firefox private mode don't support BroadcastChannel
      // localStorage.setItem triggers 'storage' event in other tabs
      const handleStorageChange = (event: StorageEvent) => {
        if (event.key === "__SCREEN_LOCK_STATE__") {
          try {
            const data = event.newValue ? JSON.parse(event.newValue) : null;
            if (data?.type === "SCREEN_LOCK_CHANGED") {
              const { isLocked } = data;
              logger.info(
                "🔄 Screen lock state changed in another tab (localStorage), syncing",
                {
                  component: "IdleTimerContainer",
                  isLocked,
                  method: "localStorage",
                }
              );
              setState(isLocked ? "Idle" : "Active");
              setIsDialogOpen(isLocked);
            }
          } catch (error) {
            logger.debug("Failed to parse storage event data", {
              component: "IdleTimerContainer",
            });
          }
        }
      };

      window.addEventListener("storage", handleStorageChange);
      storageListenerCleanup = () =>
        window.removeEventListener("storage", handleStorageChange);

      return () => {
        if (broadcastChannelRef.current) {
          broadcastChannelRef.current.removeEventListener(
            "message",
            handleMessage
          );
          broadcastChannelRef.current.close();
          broadcastChannelRef.current = null;
        }
        if (storageListenerCleanup) {
          storageListenerCleanup();
        }
      };
    } catch (error) {
      logger.warn(
        "BroadcastChannel not supported, using localStorage fallback for multi-tab sync",
        {
          component: "IdleTimerContainer",
          error: (error as Error)?.message,
        }
      );

      // Fallback: Use localStorage for cross-tab communication
      const handleStorageChange = (event: StorageEvent) => {
        if (event.key === "__SCREEN_LOCK_STATE__") {
          try {
            const data = event.newValue ? JSON.parse(event.newValue) : null;
            if (data?.type === "SCREEN_LOCK_CHANGED") {
              const { isLocked } = data;
              logger.info(
                "🔄 Screen lock state changed in another tab (localStorage), syncing",
                {
                  component: "IdleTimerContainer",
                  isLocked,
                  method: "localStorage",
                }
              );
              setState(isLocked ? "Idle" : "Active");
              setIsDialogOpen(isLocked);
            }
          } catch (error) {
            logger.debug("Failed to parse storage event data", {
              component: "IdleTimerContainer",
            });
          }
        }
      };

      window.addEventListener("storage", handleStorageChange);
      storageListenerCleanup = () =>
        window.removeEventListener("storage", handleStorageChange);

      return () => {
        if (storageListenerCleanup) {
          storageListenerCleanup();
        }
      };
    }
  }, [loggedIn]);

  // ✅ Broadcast state changes to other tabs (BroadcastChannel + localStorage fallback)
  useEffect(() => {
    if (!loggedIn) return;

    const message = {
      type: "SCREEN_LOCK_CHANGED",
      isLocked: isIdle,
    };

    // Try BroadcastChannel first
    if (broadcastChannelRef.current) {
      try {
        broadcastChannelRef.current.postMessage(message);
        logger.debug("📢 Broadcasted screen lock state change to other tabs", {
          component: "IdleTimerContainer",
          isLocked: isIdle,
          method: "BroadcastChannel",
        });
        return;
      } catch (error) {
        logger.debug(
          "Failed to broadcast via BroadcastChannel, falling back to localStorage",
          {
            component: "IdleTimerContainer",
            error: (error as Error)?.message,
          }
        );
      }
    }

    // Fallback: Use localStorage for browsers without BroadcastChannel
    try {
      window.localStorage.setItem(
        "__SCREEN_LOCK_STATE__",
        JSON.stringify({ ...message, timestamp: Date.now() })
      );
      logger.debug("📢 Broadcasted screen lock state change via localStorage", {
        component: "IdleTimerContainer",
        isLocked: isIdle,
        method: "localStorage",
      });
    } catch (error) {
      logger.debug("Failed to broadcast via localStorage", {
        component: "IdleTimerContainer",
        error: (error as Error)?.message,
      });
    }
  }, [isIdle, loggedIn]);

  // ✅ Extract all values from refresh token hook for error handling
  const {
    data: refreshData,
    error: refreshError,
    isLoading: isRefreshing,
  } = useRefreshToken(Boolean(loggedIn && !isIdle));

  // ✅ Handle refresh errors - CRITICAL FIX: Prevent silent auth failure
  useEffect(() => {
    if (refreshError) {
      logger.error(
        "🔄 Background token refresh failed - session may be expiring",
        refreshError,
        {
          component: "IdleTimerContainer",
          refreshError: (refreshError as any)?.message,
        }
      );

      // Show user warning if token refresh consistently fails
      // This prevents the silent expiry issue where user works but token dies
      toast.warning(
        "⚠️ Your session may be expiring. Please save your work and log back in if needed.",
        {
          duration: 10000, // Show for 10 seconds
        }
      );
    }
  }, [refreshError]);

  // ✅ Optional: Log refresh state for debugging
  useEffect(() => {
    if (isRefreshing) {
      logger.debug("Background token refresh in progress", {
        component: "IdleTimerContainer",
      });
    }
  }, [isRefreshing]);

  const onIdle = async () => {
    try {
      logger.debug("🔒 Idle timeout detected, attempting to lock screen", {
        component: "IdleTimerContainer.onIdle",
      });

      // ✅ CRITICAL FIX: Show modal REGARDLESS of cookie success
      // User's explicit requirement: "open the modal regardless of what is on the screen"
      // We must show the modal when user is idle, even if cookie operation fails
      // This ensures the countdown timer always appears to protect the session

      // Try to set screen lock cookie, but don't block modal from showing
      let lockSuccess = false;
      try {
        lockSuccess = await lockScreenOnUserIdle(true);
      } catch (lockError) {
        logger.error(
          "Exception while setting screen lock cookie - will show modal anyway",
          lockError,
          {
            component: "IdleTimerContainer.onIdle",
          }
        );
        // Continue - we'll show modal even if cookie fails
      }

      if (!lockSuccess) {
        logger.warn(
          "Screen lock cookie not set, but showing modal anyway (user requirement)",
          {
            component: "IdleTimerContainer.onIdle",
            lockSuccess,
          }
        );
        // Don't toast error - user already knows they're idle, modal is sufficient
        // toast.error("Failed to lock screen. Please try again.");
        // CONTINUE TO SHOW MODAL
      } else {
        logger.info("✅ Screen lock activated successfully", {
          component: "IdleTimerContainer.onIdle",
          lockSuccess,
        });
      }

      // ✅ CRITICAL FIX: Open dialog BEFORE state change to prevent race condition
      // Setting state and dialog in separate calls can cause the component to render
      // with only one state update applied, causing the dialog to be missed
      // By opening dialog first, we ensure it's visible even if state updates are batched
      // MORE IMPORTANT: Show modal REGARDLESS of whether cookie was set
      setIsDialogOpen(true);
      setState("Idle");
    } catch (error) {
      logger.error("❌ Exception while activating screen lock", error, {
        component: "IdleTimerContainer.onIdle",
      });
      // ✅ CRITICAL FIX: Still show modal even on exception (user requirement)
      // "open the modal regardless of what is on the screen"
      logger.info("Showing modal despite error (user idle requirement)", {
        component: "IdleTimerContainer.onIdle",
      });
      setIsDialogOpen(true);
      setState("Idle");
    }
  };

  const onActive = () => {
    if (state === "Idle") return;
    // Reset local idle state if not idle
    setState("Active");
    // ✅ CRITICAL FIX: Reset the actual idle timer countdown
    // Without this, the timer continues counting even when user is active
    idleTimer.reset();
  };

  const onAction = async () => setCount(count + 1);

  const idleTimer = useIdleTimer({
    onIdle,
    onActive,
    onAction,
    timeout: SESSION_CONFIG.IDLE_TIMEOUT, // 5 minutes
    throttle: 500,
    disabled: !loggedIn,
  });

  const handleUserLogOut = useCallback(async () => {
    if (hasLoggedOutRef.current) return; // Prevent multiple logout calls
    hasLoggedOutRef.current = true;

    setIsLoading(true);
    setState("Active");
    setIsDialogOpen(false);
    setCount(0);

    try {
      // Verify screen lock cookie exists before proceeding with logout
      const screenLockExists = await checkScreenLockState();
      if (!screenLockExists) {
        logger.warn("⚠️ Screen lock cookie missing during logout attempt", {
          component: "IdleTimerContainer.handleUserLogOut",
          screenLockExists,
        });
        // Still proceed with logout to ensure proper cleanup
      }

      logger.info("🚪 Logging user out - session timed out", {
        component: "IdleTimerContainer.handleUserLogOut",
      });

      // Use server action for proper session cleanup (deletes cookies & JWT)
      const response = await logUserOut("User session timed out.");

      if (response.success) {
        logger.info("✅ Logout successful", {
          component: "IdleTimerContainer.handleUserLogOut",
        });
      } else {
        logger.warn(
          "⚠️ Logout response indicated failure, but proceeding with redirect",
          {
            component: "IdleTimerContainer.handleUserLogOut",
            response,
          }
        );
      }

      // Redirect to login regardless of response (session is already deleted)
      window.location.replace("/login");
    } catch (error) {
      logger.error("❌ Logout error", error, {
        component: "IdleTimerContainer.handleUserLogOut",
      });
      // Force redirect even on error
      window.location.replace("/login");
    } finally {
      setIsLoading(false);
    }
  }, []);

  // Callback to handle "I'm still here" button click
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
          success,
        });
        setState("Active");
        setIsDialogOpen(false);
        idleTimer.reset();
        toast.success("Session extended. Welcome back!");
        return;
      }

      logger.warn(
        "Screen unlock returned false, attempting fallback token refresh",
        {
          component: "IdleTimerContainer.handleStillHere",
          success,
        }
      );

      const refreshResponse = await getRefreshToken();

      if (refreshResponse.success) {
        logger.info("✅ Fallback: Token refreshed successfully", {
          component: "IdleTimerContainer.handleStillHere",
          refreshSuccess: refreshResponse.success,
        });
        setState("Active");
        setIsDialogOpen(false);
        idleTimer.reset();
        toast.success("Session restored. You're all set!");
        return;
      }

      logger.error("Both unlock and refresh failed", {
        component: "IdleTimerContainer.handleStillHere",
        unlockSuccess: success,
        refreshSuccess: refreshResponse.success,
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
  }, [idleTimer, handleUserLogOut]);

  /* NO TIMER ON EXTERNAL ROUTES */
  // if (pathname.startsWith("/checkout")) return null;
  // if (pathname.startsWith("/invoice")) return null;
  // if (pathname.startsWith("/subscriptions")) return null;

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
  if (isDialogOpen) {
    return (
      <ScreenLock
        open={isDialogOpen}
        onStillHere={handleStillHere}
        isLoading={isLoading}
        setIsLoading={setIsLoading}
        handleUserLogOut={handleUserLogOut}
        hasLoggedOutRef={hasLoggedOutRef}
      />
    );
  }

  return null;
}

export default ScreenLock;
