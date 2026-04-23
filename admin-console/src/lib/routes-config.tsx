type AdminRouteItem = {
  title: string;
  href: string;
  icon?: string;
  badge?: string;
  items?: AdminRouteItem[];
};

type AdminRoutesType = {
  title: string;
  items: AdminRouteItem[];
};

export const admin_routes: AdminRoutesType[] = [
  {
    title: "Support Overview",
    items: [
      {
        title: "Dashboard",
        href: "/admin/dashboard",
        icon: "LayoutDashboard",
      },
      {
        title: "System Health",
        href: "/admin/system-health",
        icon: "Activity",
      },
      {
        title: "Analytics",
        href: "/admin/analytics",
        icon: "BarChart3",
      },
    ],
  },
  {
    title: "Customer Support",
    items: [
      {
        title: "Tickets",
        href: "/admin/tickets",
        icon: "Ticket",
      },
      {
        title: "Organizations",
        href: "/admin/organizations",
        icon: "Building2",
      },
      {
        title: "Users",
        href: "/admin/users",
        icon: "Users",
      },
      {
        title: "Admin Users",
        href: "/admin/admin-users",
        icon: "UserCog",
      },
    ],
  },
  {
    title: "Diagnostics",
    items: [
      {
        title: "Audit Logs",
        href: "/admin/audit-logs",
        icon: "FileText",
      },
      { title: "API Monitoring", href: "/admin/api-monitoring", icon: "Zap" },
      {
        title: "Impersonation Logs",
        href: "/admin/impersonation",
        icon: "Eye",
      },
    ],
  },
  {
    title: "Platform Control",
    items: [
      {
        title: "Roles & Permissions",
        href: "/admin/roles",
        icon: "Shield",
      },
      {
        title: "Subscriptions",
        href: "/admin/subscriptions",
        icon: "CreditCard",
        badge: "3",
      },
      {
        title: "System Settings",
        href: "/admin/settings",
        icon: "Settings",
      },
      {
        title: "Feature Flags",
        href: "/admin/feature-flags",
        icon: "Flag",
      },
      { title: "Notifications", href: "/admin/notifications", icon: "Bell" },
    ],
  },
];
