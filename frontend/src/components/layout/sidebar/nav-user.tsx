"use client";

import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger
} from "@/components/ui/dropdown-menu";
import {
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  useSidebar
} from "@/components/ui/sidebar";
import { 
  ChevronRightIcon, 
  DiamondIcon, 
  FileTextIcon, 
  LogOutIcon, 
  PaletteIcon, 
  UserIcon,
  ExternalLinkIcon,
  CircleIcon,
  MoreVertical
} from "lucide-react";

import Link from "next/link";
import { useSession } from "@/hooks/use-session";
import { useLogout } from "@/hooks/use-organization-mutations";

const getInitials = (name: string) => {
  return name
    .split(" ")
    .map((n) => n[0])
    .join("")
    .toUpperCase();
};

const formatRole = (role: string) => {
  return role
    .split("_")
    .map(word => word.charAt(0).toUpperCase() + word.slice(1).toLowerCase())
    .join(" ");
};

export function NavUser() {
  const { isMobile } = useSidebar();
  const { user } = useSession();
  const { logout, isPending } = useLogout();

  if (!user) {
    return (
      <div className="space-y-2 p-2">
        <div className="flex items-center gap-3 rounded-lg border p-3">
          <Avatar className="h-8 w-8">
            <AvatarFallback>...</AvatarFallback>
          </Avatar>
          <div className="flex-1">
            <div className="h-4 w-24 bg-muted animate-pulse rounded mb-1"></div>
            <div className="h-3 w-16 bg-muted animate-pulse rounded"></div>
          </div>
        </div>
      </div>
    );
  }

  const initials = getInitials(user.name);
  const formattedRole = formatRole(user.role);

  return (
    <div className="space-y-2 p-2">
      {/* User Profile Section with Dropdown */}
      <SidebarMenu>
        <SidebarMenuItem>
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <SidebarMenuButton
                size="lg"
                className="data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground h-auto p-3"
              >
                <Avatar className="h-8 w-8 rounded-full">
                  <AvatarImage src={user.avatar || `https://bundui-images.netlify.app/avatars/01.png`} alt={user.name} />
                  <AvatarFallback>{initials}</AvatarFallback>
                </Avatar>
                <div className="grid flex-1 text-left text-sm leading-tight">
                  <span className="truncate font-medium">{user.name}</span>
                  <span className="text-muted-foreground truncate text-xs">{formattedRole}</span>
                </div>
                <MoreVertical className="ml-auto h-4 w-4" />
              </SidebarMenuButton>
            </DropdownMenuTrigger>
            <DropdownMenuContent
              className="w-56 rounded-lg"
              side={isMobile ? "bottom" : "right"}
              align="end"
              sideOffset={4}
            >
              <DropdownMenuLabel className="p-0 font-normal">
                <div className="flex items-center gap-2 px-1 py-1.5 text-left text-sm">
                  <Avatar className="h-8 w-8 rounded-full">
                    <AvatarImage src={user.avatar || `https://bundui-images.netlify.app/avatars/01.png`} alt={user.name} />
                    <AvatarFallback>{initials}</AvatarFallback>
                  </Avatar>
                  <div className="grid flex-1 text-left text-sm leading-tight">
                    <span className="truncate font-medium">{user.name}</span>
                    <span className="text-muted-foreground truncate text-xs">{user.email}</span>
                  </div>
                </div>
              </DropdownMenuLabel>
              <DropdownMenuSeparator />
              <DropdownMenuItem asChild>
                <Link href="/settings" className="flex items-center gap-2">
                  <UserIcon className="h-4 w-4" />
                  <span>Account Settings</span>
                </Link>
              </DropdownMenuItem>
              <DropdownMenuItem asChild>
                <Link href="/settings" className="flex items-center gap-2">
                  <PaletteIcon className="h-4 w-4" />
                  <span>Theme</span>
                  <ChevronRightIcon className="ml-auto h-4 w-4" />
                </Link>
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem onClick={() => logout()} disabled={isPending}>
                <LogOutIcon className="h-4 w-4" />
                <span>{isPending ? "Logging out..." : "Log out"}</span>
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </SidebarMenuItem>
      </SidebarMenu>

      {/* Plan Section */}
      <div className="flex items-center gap-3 rounded-lg border p-3 hover:bg-accent/50 transition-colors cursor-pointer">
        <DiamondIcon className="h-5 w-5 text-muted-foreground" />
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <span className="text-sm font-medium">Pro Plan</span>
            <Badge variant="secondary" className="text-xs bg-green-100 text-green-700 hover:bg-green-100">
              Trial
            </Badge>
          </div>
        </div>
        <ChevronRightIcon className="h-4 w-4 text-muted-foreground" />
      </div>

      {/* Upgrade Button */}
      <Button 
        variant="outline" 
        className="w-full border-green-200 text-green-700 hover:bg-green-50 hover:text-green-800"
      >
        Upgrade now
      </Button>

      {/* Docs and Resources */}
      <div className="flex items-center gap-3 rounded-lg border p-3 hover:bg-accent/50 transition-colors cursor-pointer">
        <FileTextIcon className="h-5 w-5 text-muted-foreground" />
        <div className="flex-1">
          <span className="text-sm text-muted-foreground">Docs and resources</span>
        </div>
        <ChevronRightIcon className="h-4 w-4 text-muted-foreground" />
      </div>

      {/* System Status */}
      <div className="flex items-center gap-3 rounded-lg border p-3 hover:bg-accent/50 transition-colors cursor-pointer">
        <CircleIcon className="h-3 w-3 fill-green-500 text-green-500" />
        <div className="flex-1">
          <span className="text-sm text-muted-foreground">All systems operational</span>
        </div>
        <ExternalLinkIcon className="h-4 w-4 text-muted-foreground" />
      </div>
    </div>
  );
}
