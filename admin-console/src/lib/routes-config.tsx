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
    title: "Overview",
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
    ],
  },
  {
    title: "Organization Management",
    items: [
      {
        title: "Organizations",
        href: "/admin/organizations",
        icon: "Building2",
      },
      {
        title: "Subscriptions",
        href: "/admin/subscriptions",
        icon: "CreditCard",
        badge: "3",
      },
    ],
  },
  {
    title: "User Management",
    items: [
      {
        title: "Users",
        href: "/admin/users",
        icon: "Users",
      },
      {
        title: "Roles & Permissions",
        href: "/admin/roles",
        icon: "Shield",
      },
      {
        title: "Admin Users",
        href: "/admin/admin-users",
        icon: "UserCog",
      },
    ],
  },
  {
    title: "System Management",
    items: [
      {
        title: "Analytics",
        href: "/admin/analytics",
        icon: "BarChart3",
      },
      {
        title: "Audit Logs",
        href: "/admin/audit-logs",
        icon: "FileText",
      },
      {
        title: "API Monitoring",
        href: "/admin/api-monitoring",
        icon: "Zap",
      },
      {
        title: "Database",
        href: "/admin/database",
        icon: "Database",
      },
    ],
  },
  {
    title: "Configuration",
    items: [
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
      {
        title: "Notifications",
        href: "/admin/notifications",
        icon: "Bell",
      },
    ],
  },
];
