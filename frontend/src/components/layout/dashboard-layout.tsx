import React from "react";
import { SidebarInset } from "../ui/sidebar";
import { SiteHeader } from "./header";
import { DashboardLayoutClient } from "./dashboard-layout-client";

function DashboardLayoutGrid({ children }: { children: React.ReactNode }) {
  return (
    <DashboardLayoutClient>
      <SidebarInset className="flex flex-col">
        <SiteHeader />
        <div className="flex-1">
          <div className="@container/main p-4 xl:group-data-[theme-content-layout=centered]/layout:container xl:group-data-[theme-content-layout=centered]/layout:mx-auto">
            {children}
          </div>
        </div>
      </SidebarInset>
    </DashboardLayoutClient>
  );
}
export default DashboardLayoutGrid;
