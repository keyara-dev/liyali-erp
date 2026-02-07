/**
 * Simple notification utility to replace toast
 */

export interface NotifyOptions {
  title?: string;
  description?: string;
  variant?: "default" | "destructive" | "success";
  duration?: number;
}

export function notify(message: string, options?: NotifyOptions) {
  // For now, use console.log and alert as fallback
  // This can be replaced with a proper notification system later
  const { title, variant = "default", duration = 3000 } = options || {};

  const prefix =
    variant === "destructive" ? "❌" : variant === "success" ? "✅" : "ℹ️";
  const fullMessage = title
    ? `${prefix} ${title}: ${message}`
    : `${prefix} ${message}`;

  console.log(fullMessage);

  // Show browser notification if available
  if ("Notification" in window && Notification.permission === "granted") {
    new Notification(title || "Admin Console", {
      body: message,
      icon: "/favicon.ico",
    });
  } else {
    // Fallback to alert for important messages
    if (variant === "destructive") {
      alert(fullMessage);
    }
  }
}

// Convenience methods
export const notifySuccess = (message: string, title?: string) =>
  notify(message, { variant: "success", title });

export const notifyError = (message: string, title?: string) =>
  notify(message, { variant: "destructive", title });

export const notifyInfo = (message: string, title?: string) =>
  notify(message, { variant: "default", title });
