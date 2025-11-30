"use client";

import {
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem
} from "@/components/ui/sidebar";
import {
  ActivityIcon,
  BarChart3Icon,
  HomeIcon,
  StoreIcon,
  UsersIcon,
  SettingsIcon,
  FileText,
  Search,
  LayoutDashboard,
  ShieldAlert,
  QrCode,
  Zap,
  FileCheck,
  CheckSquare,
  DollarSign,
  type LucideIcon
} from "lucide-react";
import Link from "next/link";
import { usePathname } from "next/navigation";

type NavItem = {
  title: string;
  href: string;
  icon: LucideIcon;
}[];

type NavGroup = {
  title: string;
  items: NavItem;
};

export const navItems: NavGroup[] = [
  {
    title: "Workflows",
    items: [
      {
        title: "Dashboard",
        href: "/workflows/dashboard",
        icon: LayoutDashboard
      },
      {
        title: "Tasks",
        href: "/workflows/tasks",
        icon: CheckSquare
      },
      {
        title: "Search Transactions",
        href: "/workflows/search",
        icon: Search
      },
      {
        title: "Requisitions",
        href: "/workflows/requisitions",
        icon: FileText
      },
      {
        title: "Purchase Orders",
        href: "/workflows/purchase-orders",
        icon: FileCheck
      },
      {
        title: "Payment Vouchers",
        href: "/workflows/payment-vouchers",
        icon: FileText
      },
      {
        title: "Goods Received Notes",
        href: "/workflows/grn",
        icon: FileCheck
      }
    ]
  },
  {
    title: "Budget Management",
    items: [
      {
        title: "Budgets",
        href: "/workflows/budgets",
        icon: DollarSign
      }
    ]
  },
  {
    title: "Admin",
    items: [
      {
        title: "Reports",
        href: "/admin/reports",
        icon: BarChart3Icon
      },
      {
        title: "User Management",
        href: "/admin/users",
        icon: UsersIcon
      },
      {
        title: "Activity Logs",
        href: "/admin/logs",
        icon: ActivityIcon
      }
    ]
  },
  {
    title: "Compliance & Monitoring",
    items: [
      {
        title: "Compliance Tracking",
        href: "/compliance/tracking",
        icon: ShieldAlert
      },
      {
        title: "System Monitoring",
        href: "/monitoring",
        icon: Zap
      },
      {
        title: "QR Verification",
        href: "/verification/qr",
        icon: QrCode
      }
    ]
  },
  {
    title: "Settings",
    items: [
      {
        title: "Settings",
        href: "/settings",
        icon: SettingsIcon
      }
    ]
  }
];

export function NavMain() {
  const pathname = usePathname();

  return (
    <>
      {navItems.map((group) => (
        <SidebarGroup key={group.title}>
          <SidebarGroupLabel>{group.title}</SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu>
              {group.items.map((item) => {
                const isActive =
                  pathname === item.href ||
                  (item.href !== "/" && pathname.startsWith(item.href));

                return (
                  <SidebarMenuItem key={item.title}>
                    <SidebarMenuButton asChild isActive={isActive}>
                      <Link href={item.href}>
                        <item.icon />
                        <span>{item.title}</span>
                      </Link>
                    </SidebarMenuButton>
                  </SidebarMenuItem>
                );
              })}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
      ))}
    </>
  );
}
