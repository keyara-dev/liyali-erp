export const BASE_URL =
  process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

export const ADMIN_SESSION = "__com.liyali-admin.com__";
export const ADMIN_USER_SESSION = "__com.liyali-admin-user__";
export const ADMIN_PERMISSIONS_SESSION = "__com.liyali-admin-pem__";

export const MAX_FILE_SIZE = 5 * 1024 * 1024; // 5MB
export const DEFAULT_DATE_RANGE_DAYS = 30; // 30 DAYS
export const DEFAULT_PAGINATION = { page: 1, limit: 20 };

export const DEFAULT_DATE_RANGE = {
  start_date: new Date(
    new Date().getTime() - DEFAULT_DATE_RANGE_DAYS * 24 * 60 * 60 * 1000,
  )
    .toISOString()
    .split("T")[0],
  end_date: new Date().toISOString().split("T")[0],
  range: "",
};

export const MONTHS = [
  "January",
  "February",
  "March",
  "April",
  "May",
  "June",
  "July",
  "August",
  "September",
  "October",
  "November",
  "December",
];

// ADMIN QUERY KEYS - Single source of truth for all React Query keys
export const ADMIN_QUERY_KEYS = {
  // Organizations
  ORGANIZATIONS: {
    ALL: "admin-organizations-all",
    BY_ID: "admin-organization-by-id",
    STATS: "admin-organizations-stats",
    TRIAL_STATUS: "admin-organization-trial-status",
  },

  // Users
  USERS: {
    ALL: "admin-users-all",
    BY_ID: "admin-user-by-id",
    ADMIN_USERS: "admin-admin-users",
    STATS: "admin-users-stats",
  },

  // Subscriptions
  SUBSCRIPTIONS: {
    ALL: "admin-subscriptions-all",
    PLANS: "admin-subscription-plans",
    BY_ORG: "admin-subscription-by-org",
  },

  // Analytics
  ANALYTICS: {
    DASHBOARD: "admin-analytics-dashboard",
    SYSTEM_HEALTH: "admin-system-health",
    API_MONITORING: "admin-api-monitoring",
  },

  // Audit Logs
  AUDIT_LOGS: {
    ALL: "admin-audit-logs-all",
    BY_USER: "admin-audit-logs-by-user",
    BY_ACTION: "admin-audit-logs-by-action",
  },

  // System
  SYSTEM: {
    SETTINGS: "admin-system-settings",
    FEATURE_FLAGS: "admin-feature-flags",
    NOTIFICATIONS: "admin-notifications",
  },
};

// Admin role — only super_admin is permitted to access the admin console
export const ADMIN_ROLES = {
  SUPER_ADMIN: "super_admin",
} as const;

export type AdminRole = (typeof ADMIN_ROLES)[keyof typeof ADMIN_ROLES];
