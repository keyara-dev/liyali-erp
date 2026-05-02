import * as React from "react";
import { cn } from "@/lib/utils";

export interface DetailShellProps {
  /** Top header slot — typically <PageHeader> or a custom DocumentHeader. */
  header: React.ReactNode;
  /** Main content (left column on lg+). */
  children: React.ReactNode;
  /** Sidebar content (right column on lg+, stacked below main on mobile). */
  sidebar?: React.ReactNode;
  /** Sidebar width on lg+ as a Tailwind grid template fragment. Default "320px". */
  sidebarWidth?: string;
  className?: string;
}

export function DetailShell({
  header,
  children,
  sidebar,
  sidebarWidth = "320px",
  className,
}: DetailShellProps) {
  return (
    <div className={cn("space-y-5", className)}>
      <div>{header}</div>
      {sidebar ? (
        <div
          className="grid gap-5 lg:gap-6"
          style={{
            gridTemplateColumns: `minmax(0, 1fr) minmax(0, ${sidebarWidth})`,
          }}
        >
          <div className="min-w-0 lg:col-start-1 col-span-full lg:col-span-1">
            {children}
          </div>
          <aside className="min-w-0 lg:col-start-2 col-span-full lg:col-span-1">
            {sidebar}
          </aside>
        </div>
      ) : (
        <div className="min-w-0">{children}</div>
      )}
    </div>
  );
}
