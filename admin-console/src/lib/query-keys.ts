/**
 * Centralized React Query key factories.
 *
 * Usage:
 *   queryKey: queryKeys.adminUsers.list(filters)
 *   invalidateQueries({ queryKey: queryKeys.adminUsers.all })
 *
 * The `all` property is a plain array (not a function) so it can be spread
 * or passed directly as a prefix for broad invalidation.
 */

export const queryKeys = {
  // ── Admin Users ────────────────────────────────────────────────────────────
  adminUsers: {
    all: ["admin-users"] as const,
    list: (filters?: unknown) => ["admin-users", filters] as const,
    detail: (id: string) => ["admin-users", id] as const,
    stats: () => ["admin-users", "stats"] as const,
    roles: () => ["admin-users", "roles"] as const,
    activity: (userId: string, limit?: number) =>
      ["admin-users", userId, "activity", { limit }] as const,
    sessions: (userId: string) => ["admin-users", userId, "sessions"] as const,
  },

  // ── Analytics ──────────────────────────────────────────────────────────────
  analytics: {
    all: ["analytics"] as const,
    overview: (filters?: unknown) => ["analytics", "overview", filters] as const,
    users: (filters?: unknown) => ["analytics", "users", filters] as const,
    organizations: (filters?: unknown) =>
      ["analytics", "organizations", filters] as const,
    revenue: (filters?: unknown) => ["analytics", "revenue", filters] as const,
    usage: (filters?: unknown) => ["analytics", "usage", filters] as const,
  },

  // ── API Monitoring ─────────────────────────────────────────────────────────
  apiMonitoring: {
    all: ["api-monitoring"] as const,
    endpoints: (filters?: unknown) =>
      ["api-monitoring", "endpoints", filters] as const,
    endpoint: (id: string) => ["api-monitoring", "endpoints", id] as const,
    metrics: (filters?: unknown) =>
      ["api-monitoring", "metrics", filters] as const,
    endpointMetrics: (id: string, timeRange?: unknown) =>
      ["api-monitoring", "endpoints", id, "metrics", timeRange] as const,
    errors: (filters?: unknown) =>
      ["api-monitoring", "errors", filters] as const,
    error: (id: string) => ["api-monitoring", "errors", id] as const,
    alerts: (filters?: unknown) =>
      ["api-monitoring", "alerts", filters] as const,
    stats: () => ["api-monitoring", "stats"] as const,
    performance: (timeRange?: unknown, interval?: unknown) =>
      ["api-monitoring", "performance", timeRange, interval] as const,
    categories: () => ["api-monitoring", "categories"] as const,
    realtime: () => ["api-monitoring", "realtime"] as const,
  },

  // ── Audit Logs ─────────────────────────────────────────────────────────────
  auditLogs: {
    all: ["audit-logs"] as const,
    list: (filters?: unknown, page?: number, limit?: number) =>
      ["audit-logs", filters, page, limit] as const,
    stats: (filters?: unknown) => ["audit-logs", "stats", filters] as const,
    analytics: (filters?: unknown) =>
      ["audit-logs", "analytics", filters] as const,
    detail: (id: string) => ["audit-logs", id] as const,
    securityEvents: (filters?: unknown) =>
      ["audit-logs", "security-events", filters] as const,
    retentionSettings: () => ["audit-logs", "retention-settings"] as const,
  },

  // ── Dashboard ──────────────────────────────────────────────────────────────
  dashboard: {
    all: ["dashboard"] as const,
    metrics: () => ["dashboard", "metrics"] as const,
    systemHealth: () => ["dashboard", "system-health"] as const,
  },

  // ── Database ───────────────────────────────────────────────────────────────
  database: {
    all: ["database"] as const,
    connections: (filters?: unknown) =>
      ["database", "connections", filters] as const,
    connection: (id: string) => ["database", "connections", id] as const,
    metrics: (filters?: unknown) => ["database", "metrics", filters] as const,
    tables: (connectionId: string, filters?: unknown) =>
      ["database", "connections", connectionId, "tables", filters] as const,
    queries: (filters?: unknown) => ["database", "queries", filters] as const,
    backups: (filters?: unknown) => ["database", "backups", filters] as const,
    migrations: (connectionId: string) =>
      ["database", "connections", connectionId, "migrations"] as const,
    stats: () => ["database", "stats"] as const,
    schemas: (connectionId: string) =>
      ["database", "connections", connectionId, "schemas"] as const,
    performance: (connectionId: string, timeRange?: unknown) =>
      [
        "database",
        "connections",
        connectionId,
        "performance",
        timeRange,
      ] as const,
  },

  // ── Feature Flags ──────────────────────────────────────────────────────────
  featureFlags: {
    all: ["feature-flags"] as const,
    list: (filters?: unknown) => ["feature-flags", filters] as const,
    detail: (id: string) => ["feature-flags", id] as const,
    stats: () => ["feature-flags", "stats"] as const,
    analytics: (flagKey: string) =>
      ["feature-flags", flagKey, "analytics"] as const,
  },

  // ── Notifications ──────────────────────────────────────────────────────────
  notifications: {
    all: ["admin-notifications"] as const,
    list: (filters?: unknown) => ["admin-notifications", filters] as const,
    stats: () => ["admin-notification-stats"] as const,
  },

  // ── Organizations ──────────────────────────────────────────────────────────
  organizations: {
    all: ["organizations"] as const,
    list: (filters?: unknown) => ["organizations", filters] as const,
    detail: (id: string) => ["organizations", id] as const,
    stats: () => ["organizations", "statistics"] as const,
    users: (orgId: string, page?: number, limit?: number) =>
      ["organizations", orgId, "users", { page, limit }] as const,
    activity: (orgId: string, page?: number, limit?: number) =>
      ["organizations", orgId, "activity", { page, limit }] as const,
    trial: (orgId: string) => ["organizations", orgId, "trial"] as const,
    subscription: (orgId: string) =>
      ["organizations", orgId, "subscription"] as const,
  },

  // ── Roles & Permissions ────────────────────────────────────────────────────
  roles: {
    all: ["roles"] as const,
    list: (filters?: unknown) => ["roles", filters] as const,
    detail: (id: string) => ["roles", id] as const,
    stats: () => ["roles", "stats"] as const,
    permissions: () => ["permissions"] as const,
    permissionsByCategory: () => ["permissions", "by-category"] as const,
    roleUsers: (roleId: string) => ["roles", roleId, "users"] as const,
    roleAudit: (roleId: string) => ["roles", roleId, "audit"] as const,
  },

  // ── Settings ───────────────────────────────────────────────────────────────
  settings: {
    all: ["settings"] as const,
    list: (filters?: unknown) => ["settings", filters] as const,
    detail: (id: string) => ["settings", id] as const,
    stats: () => ["settings", "stats"] as const,
    health: () => ["settings", "health"] as const,
    envVariables: (environment?: unknown) =>
      ["settings", "env-variables", environment] as const,
  },

  // ── Subscriptions ──────────────────────────────────────────────────────────
  subscriptions: {
    all: ["subscriptions"] as const,
    tiers: () => ["subscriptions", "tiers"] as const,
    tier: (id: string) => ["subscriptions", "tiers", id] as const,
    features: () => ["subscriptions", "features"] as const,
    trials: () => ["subscriptions", "trials"] as const,
    analytics: () => ["subscriptions", "analytics"] as const,
    statistics: () => ["subscriptions", "statistics"] as const,
  },

  // ── System Health ──────────────────────────────────────────────────────────
  systemHealth: {
    all: ["system-health"] as const,
    overview: () => ["system-health"] as const,
    metrics: () => ["system-health", "metrics"] as const,
    alerts: (status?: unknown, severity?: unknown) =>
      ["system-health", "alerts", status, severity] as const,
    logs: (
      level?: unknown,
      component?: unknown,
      page?: number,
      limit?: number,
    ) => ["system-health", "logs", level, component, page, limit] as const,
    performance: () => ["system-health", "performance"] as const,
    config: () => ["system-health", "config"] as const,
  },

  // ── Impersonation Logs ─────────────────────────────────────────────────────
  impersonation: {
    all: ["impersonation"] as const,
    logs: (filters?: unknown) => ["impersonation", "logs", filters] as const,
    log: (id: string) => ["impersonation", "logs", id] as const,
    stats: () => ["impersonation", "stats"] as const,
  },

  // ── Users (platform users, not admin users) ────────────────────────────────
  users: {
    all: ["users"] as const,
    list: (filters?: unknown) => ["users", filters] as const,
    detail: (id: string) => ["users", id] as const,
    stats: () => ["users", "statistics"] as const,
    activity: (userId: string, page?: number, limit?: number) =>
      ["users", userId, "activity", { page, limit }] as const,
    sessions: (userId: string) => ["users", userId, "sessions"] as const,
    organizations: (userId: string) =>
      ["users", userId, "organizations"] as const,
  },
};
