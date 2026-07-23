"use client";
import * as React from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { Drawer as Vaul } from "vaul";
import { Home, ClipboardList, FileText, Menu, X } from "lucide-react";
import { cn } from "@/lib/utils";

interface PrimaryTab {
  href: string;
  label: string;
  icon: React.ReactNode;
}

const PRIMARY: PrimaryTab[] = [
  { href: "/home", label: "Home", icon: <Home className="h-5 w-5" /> },
  { href: "/tasks", label: "Tasks", icon: <ClipboardList className="h-5 w-5" /> },
  { href: "/requisitions", label: "Documents", icon: <FileText className="h-5 w-5" /> },
];

interface MoreLink {
  href: string;
  label: string;
}

// Keep this list aligned with `nav-main.tsx`. Authoritative source can be
// extracted in Plan D — for now, hardcode the common routes.
const MORE_LINKS: MoreLink[] = [
  { href: "/purchase-orders", label: "Purchase Orders" },
  { href: "/payment-vouchers", label: "Payment Vouchers" },
  { href: "/grn", label: "Goods Received Notes" },
  { href: "/budgets", label: "Budgets" },
  { href: "/settings", label: "Settings" },
];

export function MobileBottomNav() {
  const pathname = usePathname();
  const [moreOpen, setMoreOpen] = React.useState(false);

  return (
    <>
      <nav
        className={cn(
          "md:hidden fixed bottom-0 inset-x-0 z-40",
          "border-t bg-background/95 backdrop-blur",
          "pb-[env(safe-area-inset-bottom)]"
        )}
        aria-label="Primary"
      >
        <ul className="grid grid-cols-4">
          {PRIMARY.map((t) => {
            const active = pathname === t.href || pathname.startsWith(t.href + "/");
            return (
              <li key={t.href}>
                <Link
                  href={t.href}
                  className={cn(
                    "flex flex-col items-center justify-center gap-0.5 py-2 text-[11px] font-medium",
                    "transition-colors",
                    active
                      ? "text-accent-warm"
                      : "text-muted-foreground hover:text-foreground"
                  )}
                  aria-current={active ? "page" : undefined}
                >
                  {t.icon}
                  <span>{t.label}</span>
                </Link>
              </li>
            );
          })}
          <li>
            <button
              type="button"
              onClick={() => setMoreOpen(true)}
              className={cn(
                "w-full flex flex-col items-center justify-center gap-0.5 py-2 text-[11px] font-medium",
                "text-muted-foreground hover:text-foreground transition-colors"
              )}
            >
              <Menu className="h-5 w-5" />
              <span>More</span>
            </button>
          </li>
        </ul>
      </nav>

      <Vaul.Root open={moreOpen} onOpenChange={setMoreOpen}>
        <Vaul.Portal>
          <Vaul.Overlay className="fixed inset-0 bg-black/40 z-50" />
          <Vaul.Content
            className="fixed bottom-0 inset-x-0 z-50 mt-24 flex max-h-[80svh] flex-col rounded-t-xl bg-background border-t"
          >
            <div className="mx-auto mt-2 h-1.5 w-12 shrink-0 rounded-full bg-muted" />
            <div className="flex items-center justify-between px-4 pt-3 pb-2">
              <Vaul.Title className="text-base font-semibold">More</Vaul.Title>
              <button
                onClick={() => setMoreOpen(false)}
                className="text-muted-foreground"
                aria-label="Close"
              >
                <X className="h-5 w-5" />
              </button>
            </div>
            <Vaul.Description className="sr-only">Navigation menu</Vaul.Description>
            <ul className="flex-1 overflow-y-auto px-2 pb-[max(1rem,env(safe-area-inset-bottom))]">
              {MORE_LINKS.map((l) => (
                <li key={l.href}>
                  <Link
                    href={l.href}
                    onClick={() => setMoreOpen(false)}
                    className="flex items-center px-3 py-3 rounded-md text-sm hover:bg-muted transition-colors"
                  >
                    {l.label}
                  </Link>
                </li>
              ))}
            </ul>
          </Vaul.Content>
        </Vaul.Portal>
      </Vaul.Root>
    </>
  );
}
