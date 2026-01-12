import { PropsWithChildren } from "react";
import { IdleTimerContainer } from "@/components/base/screen-lock";
import { SidebarProvider } from "@/components/ui/sidebar";
import { verifySession } from "@/lib/auth";
import { DashboardLayout } from "@/components/layout/dashboard-layout";

export const dynamic = "force-dynamic";

export default async function DashboardLayoutProvider({
  children,
}: PropsWithChildren) {
  const { session, isAuthenticated } = await verifySession(); // Replace with actual session retrieval logic

  return (
    <>
      <IdleTimerContainer session={session} />
      <SidebarProvider
        style={
          {
            "--sidebar-width": "calc(var(--spacing) * 64)",
            "--header-height": "calc(var(--spacing) * 14)",
          } as React.CSSProperties
        }
      >
        <DashboardLayout>{children}</DashboardLayout>
      </SidebarProvider>
    </>
  );
}
