import { PropsWithChildren } from "react";
import { IdleTimerContainer } from "@/components/base/screen-lock";
import { SidebarProvider } from "@/components/ui/sidebar";
import DashboardLayoutGrid from "@/components/layout/dashboard-layout";
import { verifySession } from "@/lib/auth";

export default async function MainNavProvider({ children }: PropsWithChildren) {
  const { session, isAuthenticated } = await verifySession(); // Replace with actual session retrieval logic

  console.log({ session });

  return (
    <>
      {" "}
      <IdleTimerContainer session={session} />
      <SidebarProvider
        style={
          {
            "--sidebar-width": "calc(var(--spacing) * 64)",
            "--header-height": "calc(var(--spacing) * 14)",
          } as React.CSSProperties
        }
      >
        <DashboardLayoutGrid>{children}</DashboardLayoutGrid>
      </SidebarProvider>
    </>
  );
}
