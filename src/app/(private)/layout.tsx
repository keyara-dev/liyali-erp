"use client";

import { SiteHeader } from "@/components/layout/header";
import { AppSidebar } from "@/components/layout/sidebar/app-sidebar";
import {
  SidebarInset,
  SidebarProvider,
  useSidebar,
} from "@/components/ui/sidebar";

function LayoutGrid({ children }: { children: React.ReactNode }) {
  const { state, isMobile, open } = useSidebar();

  // On mobile, sidebar is a Sheet drawer so grid is just 1fr
  // On desktop, sidebar width depends on open state
  const gridColumns = isMobile
    ? "1fr"
    : !open
      ? "0 1fr"
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
        <SidebarInset className="flex flex-col">
          <SiteHeader />
          <div className="flex-1">
            <div className="@container/main p-4 xl:group-data-[theme-content-layout=centered]/layout:container xl:group-data-[theme-content-layout=centered]/layout:mx-auto">
              {children}
            </div>
          </div>
        </SidebarInset>
      </div>
    </div>
  );
}

export default function MainNavProvider({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <>
      <SidebarProvider
        style={
          {
            "--sidebar-width": "calc(var(--spacing) * 64)",
            "--header-height": "calc(var(--spacing) * 14)",
          } as React.CSSProperties
        }
      >
        <LayoutGrid>{children}</LayoutGrid>
      </SidebarProvider>
    </>
  );
}
