import React from "react";
import { ThemeProvider } from "next-themes";
import NextTopLoader from "nextjs-toploader";
import { cn } from "@/lib/utils";
import { Toaster } from "sonner";

import "./globals.css";

export const metadata = {
  title: process.env.NEXT_PUBLIC_APP_NAME || "Liyali Admin Console",
  description:
    process.env.NEXT_PUBLIC_APP_DESCRIPTION ||
    "Administrative portal for Liyali Gateway system management",
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
