"use client";

import {
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarMenuSub,
  SidebarMenuSubButton,
  SidebarMenuSubItem,
  useSidebar,
} from "@/components/ui/sidebar";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
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
  Zap,
  FileCheck,
  CheckSquare,
  DollarSign,
  GitBranch,
  QrCode,
  type LucideIcon,
  ChevronRight,
  Blocks,
  ClipboardCopy,
} from "lucide-react";
import Link from "next/link";

import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible";
import { usePathname } from "next/navigation";
import { useIsMobile } from "@/hooks/use-mobile";

interface NavItem {
  title: string;
  href: string;
  icon?: LucideIcon;
  items?: NavItem[];
}

interface NavGroup {
  title: string;
  items: NavItem[];
}

export const routes: NavGroup[] = [
  {
    title: "MAIN",
    items: [
      {
        title: "Dashboard",
        href: "/home",
        icon: LayoutDashboard,
      },
      {
        title: "Search",
        href: "/search",
        icon: Search,
      },
      {
        title: "Budgeting",
        href: "/budgets",
        icon: DollarSign,
      },
      {
        title: "Procurement",
        href: "/(procurement)",
        icon: ClipboardCopy,
        items: [
          {
            title: "Requisitions",
            href: "/requisitions",
            icon: FileText,
          },
          {
            title: "Purchase Orders",
            href: "/purchase-orders",
            icon: FileCheck,
          },
          {
            title: "Payment Vouchers",
            href: "/payment-vouchers",
            icon: FileText,
          },
          {
            title: "Goods Received Notes",
            href: "/grn",
            icon: FileCheck,
          },
        ],
      },
    ],
  },
  {
    title: "MANAGEMENT",
    items: [
      {
        title: "Tasks",
        href: "/tasks",
        icon: CheckSquare,
      },
      {
        title: "Document Verification",
        href: "/admin/verification",
        icon: QrCode,
      },
      {
        title: "Reports & Analytics",
        href: "/admin/reports",
        icon: BarChart3Icon,
      },
    ],
  },
  {
    title: "ADMIN",
    items: [
      {
        title: "User Management",
        href: "/admin/users",
        icon: UsersIcon,
      },
      {
        title: "Processes & Workflows",
        href: "/admin/workflows",
        icon: GitBranch,
      },
      {
        title: "System Configurations",
        href: "/admin",
        icon: SettingsIcon,
        items: [
          {
            title: "Categories",
            href: "/admin/categories",
            icon: Blocks,
          },
          {
            title: "System Monitoring",
            href: "/admin/monitoring",
            icon: Zap,
          },
          {
            title: "Compliance Tracking",
            href: "/admin/compliance/tracking",
            icon: ShieldAlert,
          },
        ],
      },
    ],
  },
];

export function NavMain() {
  const pathname = usePathname();
  const isMobile = useIsMobile();

  return (
    <>
      {routes &&
        routes.map((nav: NavGroup) => (
          <SidebarGroup key={nav.title}>
            <SidebarGroupLabel>{nav.title}</SidebarGroupLabel>
            <SidebarGroupContent className="flex flex-col gap-2">
              <SidebarMenu>
                {nav.items.map((item) => (
                  <SidebarMenuItem key={item.title}>
                    {Array.isArray(item.items) && item.items.length > 0 ? (
                      <>
                        <div className="hidden group-data-[collapsible=icon]:block">
                          <DropdownMenu>
                            <DropdownMenuTrigger asChild>
                              <SidebarMenuButton tooltip={item.title}>
                                {item.icon && <item.icon />}
                                <span>{item.title}</span>
                                <ChevronRight className="ml-auto transition-transform duration-200 group-data-[state=open]/collapsible:rotate-90" />
                              </SidebarMenuButton>
                            </DropdownMenuTrigger>
                            <DropdownMenuContent
                              side={isMobile ? "bottom" : "right"}
                              align={isMobile ? "end" : "start"}
                              className="min-w-48 rounded-lg"
                            >
                              <DropdownMenuLabel>
                                {item.title}
                              </DropdownMenuLabel>
                              {item.items?.map((item) => (
                                <DropdownMenuItem
                                  className="hover:text-foreground active:text-foreground active:bg-primary/10! hover:bg-primary/10!"
                                  asChild
                                  key={item.title}
                                >
                                  <Link href={item.href}>{item.title}</Link>
                                </DropdownMenuItem>
                              ))}
                            </DropdownMenuContent>
                          </DropdownMenu>
                        </div>
                        <Collapsible className="group/collapsible block group-data-[collapsible=icon]:hidden">
                          <CollapsibleTrigger asChild>
                            <SidebarMenuButton
                              className="hover:text-foreground active:text-foreground hover:bg-primary/10 active:bg-primary/10"
                              tooltip={item.title}
                            >
                              {item.icon && <item.icon />}
                              <span>{item.title}</span>
                              <ChevronRight className="ml-auto transition-transform duration-200 group-data-[state=open]/collapsible:rotate-90" />
                            </SidebarMenuButton>
                          </CollapsibleTrigger>
                          <CollapsibleContent>
                            <SidebarMenuSub>
                              {item?.items?.map((subItem, key) => (
                                <SidebarMenuSubItem key={key}>
                                  <SidebarMenuSubButton
                                    className="hover:text-foreground active:text-foreground hover:bg-primary/10 active:bg-primary/10"
                                    isActive={pathname === subItem.href}
                                    asChild
                                  >
                                    <Link
                                      href={subItem.href}
                                      // target={subItem.newTab ? "_blank" : ""}
                                    >
                                      <span>{subItem.title}</span>
                                    </Link>
                                  </SidebarMenuSubButton>
                                </SidebarMenuSubItem>
                              ))}
                            </SidebarMenuSub>
                          </CollapsibleContent>
                        </Collapsible>
                      </>
                    ) : (
                      <SidebarMenuButton
                        className="hover:text-foreground active:text-foreground hover:bg-primary/10 active:bg-primary/10"
                        isActive={pathname === item.href}
                        tooltip={item.title}
                        asChild
                      >
                        <Link
                          href={item.href}
                          // target={item.newTab ? "_blank" : ""}
                        >
                          {item.icon && <item.icon />}
                          <span>{item.title}</span>
                        </Link>
                      </SidebarMenuButton>
                    )}
                    {/* {!!item.isComing && (
                      <SidebarMenuBadge className="peer-hover/menu-button:text-foreground opacity-50">
                        Coming
                      </SidebarMenuBadge>
                    )}
                    {!!item.isNew && (
                      <SidebarMenuBadge
                        className={cn(
                          "border border-green-400 text-green-600 peer-hover/menu-button:text-green-600",
                          {
                            "absolute top-1.5 right-8 opacity-80":
                              Array.isArray(item.items) && item.items.length > 0
                          }
                        )}>
                        New
                      </SidebarMenuBadge>
                    )}
                    {!!item.isDataBadge && (
                      <SidebarMenuBadge className="peer-hover/menu-button:text-foreground">
                        {item.isDataBadge}
                      </SidebarMenuBadge>
                    )} */}
                  </SidebarMenuItem>
                ))}
              </SidebarMenu>
            </SidebarGroupContent>
          </SidebarGroup>
        ))}
    </>
  );
}
