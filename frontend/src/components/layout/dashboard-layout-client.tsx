"use client";

import React from "react";
import { AppSidebar } from "./sidebar/app-sidebar";
import { useSidebar } from "../ui/sidebar";

interface DashboardLayoutClientProps {
  children: React.ReactNode;
}

export function DashboardLayoutClient({ children }: DashboardLayoutClientProps) {
  const { isMobile, open } = useSidebar();

  // On mobile, sidebar is a Sheet drawer so grid is just 1fr
  // On desktop, sidebar width depends on open state
  // When collapsed on desktop, show icon width (4rem) instead of full width
  const gridColumns = isMobile
    ? "1fr"
    : !open
      ? "4rem 1fr"
      : "var(--sidebar-width) 1fr";

  return (
    <div
      className="max-w-[1560px] mx-auto w-full relative"
      style={{
        display: "grid",
        gridTemplateColumns: gridColumns,
        minHeight: "100vh",
        transition: "grid-template-columns 200ms ease-linear",
      }}
    >
      {!isMobile && (
        <div
          style={{
            position: "sticky",
            top: 0,
            height: "100vh",
            zIndex: 40,
            overflow: "hidden",
          }}
        >
          <AppSidebar />
        </div>
      )}
      {isMobile && <AppSidebar />}
      <div className="flex flex-col">
        {children}
      </div>
    </div>
  );
}
