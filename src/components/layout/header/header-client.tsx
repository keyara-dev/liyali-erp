"use client";

import { PanelLeftIcon } from "lucide-react";
import { Separator } from "@/components/ui/separator";
import Search from "@/components/layout/header/search";
import ThemeSwitch from "@/components/layout/header/theme-switch";
import UserMenu from "@/components/layout/header/user-menu";
import { Button } from "@/components/ui/button";
import { useSidebar } from "@/components/ui/sidebar";

export function HeaderClient() {
  const { toggleSidebar } = useSidebar();

  return (
    <>
      <Button onClick={toggleSidebar} size="icon" variant="ghost">
        <PanelLeftIcon />
      </Button>
      <Search />
      <div className="ml-auto flex items-center gap-2">
        <ThemeSwitch />
        <UserMenu />
      </div>
    </>
  );
}
