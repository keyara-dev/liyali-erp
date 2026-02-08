import type { Metadata } from "next";
import React from "react";
import { ThemeProvider } from "next-themes";
import NextTopLoader from "nextjs-toploader";
import { cn } from "@/lib/utils";
import { Toaster } from "sonner";

import "./globals.css";

export const metadata: Metadata = {
  title: {
    default: "Liyali Admin Console",
    template: "%s | Liyali Admin",
  },
  description:
    "Administrative portal for Liyali Gateway system management. Manage users, organizations, subscriptions, and system settings.",
  robots: {
    index: false,
    follow: false,
  },
  metadataBase: new URL(
    process.env.NEXT_PUBLIC_APP_URL || "http://localhost:3001",
  ),
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body
        suppressHydrationWarning
        className={cn("bg-background font-sans antialiased")}
      >
        <ThemeProvider
          attribute="class"
          defaultTheme="light"
          enableSystem
          disableTransitionOnChange
        >
          {children}
          <Toaster position="top-center" richColors />
          <NextTopLoader
            color="var(--primary)"
            showSpinner={false}
            height={2}
          />
        </ThemeProvider>
      </body>
    </html>
  );
}
