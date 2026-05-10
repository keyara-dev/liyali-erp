# UI Revamp — Plan A: Foundation, Mobile Shell & /tasks Page

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Establish design-token foundation, mobile bottom-nav shell, and a complete revamp of the `/tasks` page (both Tasks and Approvals tabs) under a Swiss-utility aesthetic that is mobile-first and visually consistent.

**Architecture:** Add a small set of new primitives (`DataList`, `FilterBar`, `StatGrid`, `ResponsiveSheet`, `MobileBottomNav`) layered on existing ShadCN/Radix components and Tailwind v4 tokens. Introduce a complementary warm-coral accent paired with the existing cerulean primary. Convert the `/tasks` page to a single coherent layout that swaps tables for stacked cards under `md`. Mount a bottom tab bar on mobile via the dashboard layout, with a "More" drawer (vaul) for overflow nav. Defer dashboard, modal-sweep, other tables, and approval-detail polish to follow-up plans (B–E, listed at end).

**Tech Stack:** Next.js 15 App Router, TypeScript, Tailwind v4, ShadCN UI (Radix), `vaul` (drawer, already installed), `sonner` (toasts, already installed), Vitest + Testing Library for component smoke tests.

**Out-of-scope (deferred):**
- Plan B — Modal → ResponsiveSheet sweep across requisition/PO/PV/GRN/approval modals
- Plan C — Other tables (requisition/PO/PV/GRN/users/orgs) → DataList
- Plan D — Dashboard tighten (greeting card, metrics, quick actions)
- Plan E — Approval flow polish (task detail, signature pad, timeline stepper, document review pane)

---

## File Structure

**Create:**
- `frontend/src/components/ui/data-list.tsx` — Responsive table+mobile-card wrapper
- `frontend/src/components/ui/filter-bar.tsx` — Standardized search + filter strip
- `frontend/src/components/ui/stat-grid.tsx` — Reusable stat strip (replaces duplicated `StatCell` in tasks/approvals)
- `frontend/src/components/ui/responsive-sheet.tsx` — Dialog on `md+`, vaul Drawer on `<md`
- `frontend/src/components/layout/mobile-bottom-nav.tsx` — Bottom tab bar with More drawer
- `frontend/src/components/layout/mobile-bottom-nav.test.tsx` — Smoke test for nav rendering
- `frontend/src/components/ui/data-list.test.tsx` — Smoke test for table↔card swap
- `frontend/src/components/ui/stat-grid.test.tsx` — Smoke test for stat values

**Modify:**
- `frontend/src/app/globals.css` — Add `--accent-warm` token + dark variant; align `--accent` to a usable color; add `--motion-fast/--motion-base` vars
- `frontend/src/components/layout/dashboard-layout.tsx` — Mount `<MobileBottomNav />`; add bottom safe-area padding on mobile
- `frontend/src/app/(private)/(main)/tasks/_components/tasks-client.tsx` — Replace `Tabs` shell with new layout, lift `TaskStatsCards` out of "Tasks" tab into a unified header strip, route `?tab=` properly
- `frontend/src/app/(private)/(main)/tasks/_components/tasks-table.tsx` — Convert to `DataList`, drop inline filter bar (use shared `FilterBar`), drop inline pagination bar style
- `frontend/src/app/(private)/(main)/tasks/_components/approvals-list.tsx` — Replace inline `StatCell` + filter card with `StatGrid` + `FilterBar`; tokens replace raw `bg-blue-100` etc
- `frontend/src/app/(private)/(main)/tasks/_components/task-stats-cards.tsx` — Reduce to a thin wrapper around `StatGrid`

**Delete:** none (keep existing primitives; new ones are additive)

---

## Design Tokens

| Token | Light | Dark | Use |
|---|---|---|---|
| `--primary` | `oklch(50.989% 0.23128 262.506)` (existing #0c54e7) | unchanged | brand actions |
| `--accent-warm` | `oklch(72% 0.16 38)` (warm coral, complements blue 262°) | `oklch(76% 0.15 38)` | highlights, hover, FAB, active mobile tab |
| `--accent-warm-foreground` | `oklch(100% 0 0)` | `oklch(15% 0 0)` | text on warm bg |
| `--motion-fast` | `120ms` | — | hover/press |
| `--motion-base` | `200ms` | — | modal/drawer/route |
| `--motion-ease` | `cubic-bezier(0.2, 0.8, 0.2, 1)` | — | default easing |

Note: existing `--accent` is a near-grey lavender that does not function as an accent. Plan repurposes `--accent-warm` as the new functional accent without breaking existing `bg-accent` callsites (we leave `--accent` alone).

---

## Task 1: Add design tokens

**Files:**
- Modify: `frontend/src/app/globals.css` (add at end of `:root` and `.dark` blocks; add `@theme` color reference)

- [ ] **Step 1: Add light-mode tokens**

In `:root` block, after the existing `--accent-foreground` line, add:

```css
  /* Warm accent (complement to cerulean primary) */
  --accent-warm: oklch(72% 0.16 38);
  --accent-warm-foreground: oklch(100% 0 0);

  /* Motion */
  --motion-fast: 120ms;
  --motion-base: 200ms;
  --motion-ease: cubic-bezier(0.2, 0.8, 0.2, 1);
```

- [ ] **Step 2: Add dark-mode tokens**

In `.dark` block, after `--accent-foreground` line, add:

```css
  --accent-warm: oklch(76% 0.15 38);
  --accent-warm-foreground: oklch(15% 0 0);
```

- [ ] **Step 3: Wire `@theme` color reference**

In the `@theme` block, after the `--color-accent-foreground` line, add:

```css
  --color-accent-warm: var(--accent-warm);
  --color-accent-warm-foreground: var(--accent-warm-foreground);
```

- [ ] **Step 4: Build to verify tokens compile**

Run: `cd frontend && pnpm build` (or `npm run build` if pnpm not installed — check `package.json` scripts)
Expected: Build succeeds; no Tailwind errors about unknown tokens.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/app/globals.css
git commit -m "feat(ui): add warm accent + motion design tokens"
```

---

## Task 2: Create `StatGrid` primitive

**Files:**
- Create: `frontend/src/components/ui/stat-grid.tsx`
- Test: `frontend/src/components/ui/stat-grid.test.tsx`

- [ ] **Step 1: Write failing test**

```tsx
// frontend/src/components/ui/stat-grid.test.tsx
import { render, screen } from "@testing-library/react";
import { describe, it, expect } from "vitest";
import { StatGrid } from "./stat-grid";
import { Clock } from "lucide-react";

describe("StatGrid", () => {
  it("renders all stat cells with labels and values", () => {
    render(
      <StatGrid
        items={[
          { label: "Pending", value: 7, icon: <Clock data-testid="icon" />, accent: "amber" },
          { label: "Done", value: 3, icon: <Clock />, accent: "emerald" },
        ]}
      />
    );
    expect(screen.getByText("Pending")).toBeInTheDocument();
    expect(screen.getByText("7")).toBeInTheDocument();
    expect(screen.getByText("Done")).toBeInTheDocument();
    expect(screen.getByText("3")).toBeInTheDocument();
  });

  it("renders secondary text when provided", () => {
    render(
      <StatGrid
        items={[
          { label: "X", value: 1, icon: <Clock />, accent: "blue", secondary: "extra" },
        ]}
      />
    );
    expect(screen.getByText("extra")).toBeInTheDocument();
  });
});
```

- [ ] **Step 2: Run test, expect failure**

Run: `cd frontend && pnpm vitest run src/components/ui/stat-grid.test.tsx`
Expected: FAIL — `Cannot find module './stat-grid'`.

- [ ] **Step 3: Implement `StatGrid`**

```tsx
// frontend/src/components/ui/stat-grid.tsx
import * as React from "react";
import { Card, CardContent } from "@/components/ui/card";
import { cn } from "@/lib/utils";

export type StatAccent = "amber" | "blue" | "rose" | "emerald" | "slate" | "violet" | "warm";

const CHIP: Record<StatAccent, string> = {
  amber: "bg-amber-100 text-amber-700 dark:bg-amber-950/50 dark:text-amber-300",
  blue: "bg-blue-100 text-blue-700 dark:bg-blue-950/50 dark:text-blue-300",
  rose: "bg-rose-100 text-rose-700 dark:bg-rose-950/50 dark:text-rose-300",
  emerald: "bg-emerald-100 text-emerald-700 dark:bg-emerald-950/50 dark:text-emerald-300",
  slate: "bg-slate-100 text-slate-700 dark:bg-slate-800 dark:text-slate-300",
  violet: "bg-violet-100 text-violet-700 dark:bg-violet-950/50 dark:text-violet-300",
  warm: "bg-accent-warm/15 text-accent-warm dark:bg-accent-warm/20",
};

const VALUE: Partial<Record<StatAccent, string>> = {
  amber: "text-amber-600 dark:text-amber-400",
  blue: "text-blue-600 dark:text-blue-400",
  rose: "text-rose-600 dark:text-rose-400",
  emerald: "text-emerald-600 dark:text-emerald-400",
  warm: "text-accent-warm",
};

export interface StatItem {
  label: string;
  value: number | string;
  icon: React.ReactNode;
  accent: StatAccent;
  secondary?: string;
  emphasizeValue?: boolean;
}

export interface StatGridProps {
  items: StatItem[];
  /** Tailwind grid-cols class for base; defaults to a sensible 2/3/N pattern. */
  className?: string;
}

export function StatGrid({ items, className }: StatGridProps) {
  const cols = items.length;
  const mdCols =
    cols >= 5 ? "md:grid-cols-5" :
    cols === 4 ? "md:grid-cols-4" :
    cols === 3 ? "md:grid-cols-3" :
    "md:grid-cols-2";
  return (
    <Card className="border-border/60 p-0">
      <CardContent
        className={cn(
          "grid grid-cols-2 sm:grid-cols-3 divide-y sm:divide-y-0 sm:divide-x divide-border/60 p-0",
          mdCols,
          className
        )}
      >
        {items.map((it) => (
          <div key={it.label} className="p-2.5 sm:p-3 space-y-0.5 sm:space-y-1">
            <div className="flex items-center justify-between gap-1.5">
              <span className="text-[10px] sm:text-xs font-medium text-muted-foreground uppercase tracking-wider truncate">
                {it.label}
              </span>
              <span
                className={cn(
                  "flex items-center justify-center rounded-md shrink-0 h-5 w-5 sm:h-6 sm:w-6",
                  CHIP[it.accent]
                )}
              >
                {it.icon}
              </span>
            </div>
            <div
              className={cn(
                "text-base sm:text-xl font-bold tabular-nums leading-tight",
                it.emphasizeValue && VALUE[it.accent]
              )}
            >
              {it.value}
            </div>
            {it.secondary && (
              <p className="text-[10px] sm:text-[11px] text-muted-foreground leading-tight truncate">
                {it.secondary}
              </p>
            )}
          </div>
        ))}
      </CardContent>
    </Card>
  );
}
```

- [ ] **Step 4: Run test, expect pass**

Run: `cd frontend && pnpm vitest run src/components/ui/stat-grid.test.tsx`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/ui/stat-grid.tsx frontend/src/components/ui/stat-grid.test.tsx
git commit -m "feat(ui): add StatGrid primitive"
```

---

## Task 3: Create `FilterBar` primitive

**Files:**
- Create: `frontend/src/components/ui/filter-bar.tsx`

- [ ] **Step 1: Implement `FilterBar`**

```tsx
// frontend/src/components/ui/filter-bar.tsx
"use client";
import * as React from "react";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { X } from "lucide-react";
import { cn } from "@/lib/utils";

export interface FilterBarProps {
  /** Search field — usually the existing `<Search />` from `ui/search-field`. */
  search?: React.ReactNode;
  /** Filter controls (Selects, etc) — render in a wrap row. */
  filters?: React.ReactNode;
  /** Whether any filter is active. */
  hasActiveFilters?: boolean;
  onReset?: () => void;
  /** Right-side meta text (e.g. "Showing 12 results"). */
  meta?: React.ReactNode;
  className?: string;
}

export function FilterBar({
  search,
  filters,
  hasActiveFilters,
  onReset,
  meta,
  className,
}: FilterBarProps) {
  return (
    <Card className={cn("border-border/60", className)}>
      <CardContent className="p-3 sm:p-4 space-y-2">
        <div className="grid gap-2 md:grid-cols-[1fr_auto] md:items-start">
          <div className="min-w-0">{search}</div>
          {filters && (
            <div className="flex flex-wrap gap-2 md:justify-end">{filters}</div>
          )}
        </div>
        {(meta || hasActiveFilters) && (
          <div className="flex items-center justify-between gap-2">
            <span className="text-xs text-muted-foreground">{meta}</span>
            {hasActiveFilters && onReset && (
              <Button variant="ghost" size="sm" onClick={onReset} className="h-7 text-xs">
                <X className="h-3 w-3 mr-1" />
                Reset
              </Button>
            )}
          </div>
        )}
      </CardContent>
    </Card>
  );
}
```

- [ ] **Step 2: Verify TypeScript compiles**

Run: `cd frontend && pnpm tsc --noEmit`
Expected: No errors in `filter-bar.tsx`.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/ui/filter-bar.tsx
git commit -m "feat(ui): add FilterBar primitive"
```

---

## Task 4: Create `DataList` responsive primitive

**Files:**
- Create: `frontend/src/components/ui/data-list.tsx`
- Test: `frontend/src/components/ui/data-list.test.tsx`

`DataList` renders a `<Table>` on `md+` and a stacked `<div>`-card list on `<md`. Caller supplies columns + a `mobileCard` render function.

- [ ] **Step 1: Write failing test**

```tsx
// frontend/src/components/ui/data-list.test.tsx
import { render, screen } from "@testing-library/react";
import { describe, it, expect } from "vitest";
import { DataList } from "./data-list";

interface Row { id: string; name: string; }

describe("DataList", () => {
  const rows: Row[] = [
    { id: "1", name: "alpha" },
    { id: "2", name: "beta" },
  ];

  it("renders rows in table mode", () => {
    render(
      <DataList<Row>
        rows={rows}
        getRowId={(r) => r.id}
        columns={[{ id: "name", header: "Name", cell: (r) => r.name }]}
        mobileCard={(r) => <div>{r.name}</div>}
      />
    );
    expect(screen.getByText("alpha")).toBeInTheDocument();
    expect(screen.getByText("beta")).toBeInTheDocument();
  });

  it("renders empty state", () => {
    render(
      <DataList<Row>
        rows={[]}
        getRowId={(r) => r.id}
        columns={[{ id: "name", header: "Name", cell: (r) => r.name }]}
        mobileCard={(r) => <div>{r.name}</div>}
        emptyMessage="No rows."
      />
    );
    expect(screen.getByText("No rows.")).toBeInTheDocument();
  });
});
```

- [ ] **Step 2: Run test, expect failure**

Run: `cd frontend && pnpm vitest run src/components/ui/data-list.test.tsx`
Expected: FAIL — module not found.

- [ ] **Step 3: Implement `DataList`**

```tsx
// frontend/src/components/ui/data-list.tsx
"use client";
import * as React from "react";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Skeleton } from "@/components/ui/skeleton";
import { cn } from "@/lib/utils";

export interface DataListColumn<T> {
  id: string;
  header: React.ReactNode;
  cell: (row: T) => React.ReactNode;
  /**
   * Visibility breakpoint for the column.
   * 'always' = always visible on desktop table
   * 'md' = visible md+
   * 'lg' = visible lg+
   */
  priority?: "always" | "md" | "lg";
  className?: string;
}

export interface DataListProps<T> {
  rows: T[];
  columns: DataListColumn<T>[];
  getRowId: (row: T) => string;
  mobileCard: (row: T) => React.ReactNode;
  isLoading?: boolean;
  /** Number of skeleton rows to render while loading. Default 5. */
  skeletonRows?: number;
  emptyMessage?: React.ReactNode;
  onRowClick?: (row: T) => void;
  className?: string;
}

const HIDE: Record<NonNullable<DataListColumn<unknown>["priority"]>, string> = {
  always: "",
  md: "hidden md:table-cell",
  lg: "hidden lg:table-cell",
};

export function DataList<T>({
  rows,
  columns,
  getRowId,
  mobileCard,
  isLoading,
  skeletonRows = 5,
  emptyMessage = "No results.",
  onRowClick,
  className,
}: DataListProps<T>) {
  if (isLoading) {
    return (
      <>
        {/* Desktop skeleton */}
        <div className="hidden md:block rounded-md border">
          <Table>
            <TableHeader>
              <TableRow>
                {columns.map((c) => (
                  <TableHead key={c.id} className={cn(HIDE[c.priority || "always"])}>
                    {c.header}
                  </TableHead>
                ))}
              </TableRow>
            </TableHeader>
            <TableBody>
              {Array.from({ length: skeletonRows }).map((_, i) => (
                <TableRow key={i}>
                  {columns.map((c) => (
                    <TableCell key={c.id} className={cn(HIDE[c.priority || "always"])}>
                      <Skeleton className="h-4 w-24" />
                    </TableCell>
                  ))}
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
        {/* Mobile skeleton */}
        <div className="md:hidden space-y-2">
          {Array.from({ length: skeletonRows }).map((_, i) => (
            <div key={i} className="rounded-md border p-3 space-y-2">
              <Skeleton className="h-4 w-2/3" />
              <Skeleton className="h-3 w-1/2" />
              <Skeleton className="h-3 w-1/3" />
            </div>
          ))}
        </div>
      </>
    );
  }

  if (rows.length === 0) {
    return (
      <div className={cn("rounded-md border py-10 text-center text-sm text-muted-foreground", className)}>
        {emptyMessage}
      </div>
    );
  }

  return (
    <div className={className}>
      {/* Desktop / tablet: table */}
      <div className="hidden md:block rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              {columns.map((c) => (
                <TableHead key={c.id} className={cn(HIDE[c.priority || "always"], c.className)}>
                  {c.header}
                </TableHead>
              ))}
            </TableRow>
          </TableHeader>
          <TableBody>
            {rows.map((row) => (
              <TableRow
                key={getRowId(row)}
                onClick={onRowClick ? () => onRowClick(row) : undefined}
                className={onRowClick ? "cursor-pointer" : undefined}
              >
                {columns.map((c) => (
                  <TableCell key={c.id} className={cn(HIDE[c.priority || "always"], c.className)}>
                    {c.cell(row)}
                  </TableCell>
                ))}
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>
      {/* Mobile: card stack */}
      <div className="md:hidden space-y-2">
        {rows.map((row) => (
          <div
            key={getRowId(row)}
            onClick={onRowClick ? () => onRowClick(row) : undefined}
            className={cn(
              "rounded-md border bg-card p-3 transition-colors",
              onRowClick && "cursor-pointer active:bg-muted/40"
            )}
          >
            {mobileCard(row)}
          </div>
        ))}
      </div>
    </div>
  );
}
```

- [ ] **Step 4: Run test, expect pass**

Run: `cd frontend && pnpm vitest run src/components/ui/data-list.test.tsx`
Expected: PASS — both rendering tests succeed (note: mobile cards render even on jsdom; both branches will be in the DOM, so we look for text, which appears in both).

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/ui/data-list.tsx frontend/src/components/ui/data-list.test.tsx
git commit -m "feat(ui): add DataList responsive table+card primitive"
```

---

## Task 5: Create `ResponsiveSheet` primitive

**Files:**
- Create: `frontend/src/components/ui/responsive-sheet.tsx`

Wraps Radix `Dialog` on `md+` and `vaul` `Drawer` on `<md`. Provides a single API so we can incrementally migrate modals in Plan B without touching consumers each time.

- [ ] **Step 1: Implement `ResponsiveSheet`**

```tsx
// frontend/src/components/ui/responsive-sheet.tsx
"use client";
import * as React from "react";
import { Drawer as Vaul } from "vaul";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { cn } from "@/lib/utils";

function useIsMobile() {
  const [is, setIs] = React.useState(false);
  React.useEffect(() => {
    const mq = window.matchMedia("(max-width: 767px)");
    const update = () => setIs(mq.matches);
    update();
    mq.addEventListener("change", update);
    return () => mq.removeEventListener("change", update);
  }, []);
  return is;
}

export interface ResponsiveSheetProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  title?: React.ReactNode;
  description?: React.ReactNode;
  children: React.ReactNode;
  footer?: React.ReactNode;
  /** Tailwind max-w on desktop. */
  desktopMaxWidth?: string;
  className?: string;
}

export function ResponsiveSheet({
  open,
  onOpenChange,
  title,
  description,
  children,
  footer,
  desktopMaxWidth = "sm:max-w-lg",
  className,
}: ResponsiveSheetProps) {
  const isMobile = useIsMobile();

  if (isMobile) {
    return (
      <Vaul.Root open={open} onOpenChange={onOpenChange}>
        <Vaul.Portal>
          <Vaul.Overlay className="fixed inset-0 bg-black/40 z-50" />
          <Vaul.Content
            className={cn(
              "fixed bottom-0 left-0 right-0 z-50 mt-24 flex max-h-[90svh] flex-col rounded-t-xl bg-background border-t",
              className
            )}
          >
            <div className="mx-auto mt-2 h-1.5 w-12 shrink-0 rounded-full bg-muted" />
            {(title || description) && (
              <div className="px-4 pt-3 pb-2 space-y-1">
                {title && <Vaul.Title className="text-base font-semibold">{title}</Vaul.Title>}
                {description && (
                  <Vaul.Description asChild>
                    <div className="text-sm text-muted-foreground">{description}</div>
                  </Vaul.Description>
                )}
              </div>
            )}
            <div className="flex-1 overflow-y-auto px-4 pb-4">{children}</div>
            {footer && (
              <div className="border-t p-3 pb-[max(0.75rem,env(safe-area-inset-bottom))]">
                {footer}
              </div>
            )}
          </Vaul.Content>
        </Vaul.Portal>
      </Vaul.Root>
    );
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent
        className={cn(desktopMaxWidth, "overflow-y-auto max-h-[90svh] p-0", className)}
      >
        {(title || description) && (
          <DialogHeader className="px-6 pt-5 pb-2">
            {title && <DialogTitle>{title}</DialogTitle>}
            {description && (
              <DialogDescription asChild>
                <div className="text-sm text-muted-foreground">{description}</div>
              </DialogDescription>
            )}
          </DialogHeader>
        )}
        <div className="px-6 pb-4">{children}</div>
        {footer && <DialogFooter className="px-6 py-4 border-t">{footer}</DialogFooter>}
      </DialogContent>
    </Dialog>
  );
}
```

- [ ] **Step 2: Verify TS compiles**

Run: `cd frontend && pnpm tsc --noEmit`
Expected: No errors. (If `Vaul.Title`/`Vaul.Description` types are missing, fall back to a `<h2>`/`<p>` and adjust — vaul exposes these in v1.x.)

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/ui/responsive-sheet.tsx
git commit -m "feat(ui): add ResponsiveSheet (Dialog desktop, Drawer mobile)"
```

---

## Task 6: Create `MobileBottomNav` with More drawer

**Files:**
- Create: `frontend/src/components/layout/mobile-bottom-nav.tsx`
- Test: `frontend/src/components/layout/mobile-bottom-nav.test.tsx`

Top 4 routes pinned: Home, Tasks, Documents (a hub view — for now point at requisitions), More. "More" opens a vaul drawer listing all sidebar nav items. Hide on `md+`.

- [ ] **Step 1: Write failing test**

```tsx
// frontend/src/components/layout/mobile-bottom-nav.test.tsx
import { render, screen } from "@testing-library/react";
import { describe, it, expect, vi } from "vitest";
import { MobileBottomNav } from "./mobile-bottom-nav";

vi.mock("next/navigation", () => ({
  usePathname: () => "/home",
  useRouter: () => ({ push: vi.fn() }),
}));

describe("MobileBottomNav", () => {
  it("renders 4 primary tabs", () => {
    render(<MobileBottomNav />);
    expect(screen.getByRole("link", { name: /home/i })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /tasks/i })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /documents/i })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /more/i })).toBeInTheDocument();
  });
});
```

- [ ] **Step 2: Run test, expect failure**

Run: `cd frontend && pnpm vitest run src/components/layout/mobile-bottom-nav.test.tsx`
Expected: FAIL — module not found.

- [ ] **Step 3: Implement `MobileBottomNav`**

```tsx
// frontend/src/components/layout/mobile-bottom-nav.tsx
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
  { href: "/reports", label: "Reports" },
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
```

- [ ] **Step 4: Run test, expect pass**

Run: `cd frontend && pnpm vitest run src/components/layout/mobile-bottom-nav.test.tsx`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/layout/mobile-bottom-nav.tsx frontend/src/components/layout/mobile-bottom-nav.test.tsx
git commit -m "feat(layout): add MobileBottomNav with More drawer"
```

---

## Task 7: Mount `MobileBottomNav` in dashboard layout

**Files:**
- Modify: `frontend/src/components/layout/dashboard-layout.tsx`

- [ ] **Step 1: Import and mount nav, raise content bottom padding**

In `frontend/src/components/layout/dashboard-layout.tsx`:

Replace:
```tsx
import { TrialBottomBanner } from "@/components/subscription/trial-bottom-banner";
```
with:
```tsx
import { TrialBottomBanner } from "@/components/subscription/trial-bottom-banner";
import { MobileBottomNav } from "./mobile-bottom-nav";
```

Replace the closing block that contains `<TrialBottomBanner />` and the closing `</div>` of the root grid with:
```tsx
      {/* Bottom Trial Banner */}
      <TrialBottomBanner />

      {/* Mobile bottom navigation */}
      <MobileBottomNav />
    </div>
```

Also raise the content padding-bottom on mobile to clear the bar (~64px). Replace:
```tsx
            <div className="@container/main p-4 pb-24 xl:group-data-[theme-content-layout=centered]/layout:container xl:group-data-[theme-content-layout=centered]/layout:mx-auto">
```
with:
```tsx
            <div className="@container/main p-4 pb-[calc(5rem+env(safe-area-inset-bottom))] md:pb-24 xl:group-data-[theme-content-layout=centered]/layout:container xl:group-data-[theme-content-layout=centered]/layout:mx-auto">
```

- [ ] **Step 2: Manually verify**

Run: `cd frontend && pnpm dev`
Open: http://localhost:3000/home in mobile viewport (DevTools, 390px wide).
Expected: Bottom nav visible with 4 tabs; "More" opens drawer; tabs navigate; nav hidden at `md+` (≥768px).

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/layout/dashboard-layout.tsx
git commit -m "feat(layout): mount MobileBottomNav, add safe-area bottom padding"
```

---

## Task 8: Refactor `task-stats-cards.tsx` to use `StatGrid`

**Files:**
- Modify: `frontend/src/app/(private)/(main)/tasks/_components/task-stats-cards.tsx`

- [ ] **Step 1: Replace inline `StatCell` with `StatGrid`**

Full replacement of the file:

```tsx
"use client";

import { useTaskStats } from "@/hooks/use-task-queries";
import { AlertCircle, CheckCircle2, Clock, Zap } from "lucide-react";
import { StatGrid } from "@/components/ui/stat-grid";
import { Skeleton } from "@/components/ui/skeleton";
import { Card, CardContent } from "@/components/ui/card";

interface TaskStatsCardsProps {
  userId: string;
  refreshTrigger: number;
}

export function TaskStatsCards({
  userId,
  refreshTrigger: _refreshTrigger,
}: TaskStatsCardsProps) {
  const { data: stats, isLoading } = useTaskStats(userId);

  if (isLoading || !stats) {
    return (
      <Card className="border-border/60 p-0">
        <CardContent className="grid grid-cols-2 md:grid-cols-4 divide-y md:divide-y-0 md:divide-x divide-border/60 p-0">
          {[...Array(4)].map((_, i) => (
            <div key={i} className="p-2.5 sm:p-3 space-y-1">
              <Skeleton className="h-3 w-16" />
              <Skeleton className="h-5 sm:h-6 w-8" />
              <Skeleton className="h-2.5 w-20" />
            </div>
          ))}
        </CardContent>
      </Card>
    );
  }

  return (
    <StatGrid
      items={[
        {
          label: "Pending",
          value: stats.pendingTasks,
          icon: <Clock className="h-3 w-3 sm:h-4 sm:w-4" />,
          accent: "amber",
          secondary: "Tasks awaiting action",
        },
        {
          label: "High Priority",
          value: stats.highPriorityTasks,
          icon: <Zap className="h-3 w-3 sm:h-4 sm:w-4" />,
          accent: "blue",
          secondary: "Urgent tasks",
        },
        {
          label: "Overdue",
          value: stats.overdueTasks,
          icon: <AlertCircle className="h-3 w-3 sm:h-4 sm:w-4" />,
          accent: "rose",
          secondary: "Past due date",
          emphasizeValue: true,
        },
        {
          label: "Completed",
          value: stats.completedTasks,
          icon: <CheckCircle2 className="h-3 w-3 sm:h-4 sm:w-4" />,
          accent: "emerald",
          secondary: "Finished tasks",
          emphasizeValue: true,
        },
      ]}
    />
  );
}
```

- [ ] **Step 2: Verify TS compiles**

Run: `cd frontend && pnpm tsc --noEmit`
Expected: No errors.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/app/\(private\)/\(main\)/tasks/_components/task-stats-cards.tsx
git commit -m "refactor(tasks): TaskStatsCards uses StatGrid primitive"
```

---

## Task 9: Refactor `tasks-table.tsx` to `DataList` + shared `FilterBar`

**Files:**
- Modify: `frontend/src/app/(private)/(main)/tasks/_components/tasks-table.tsx`

Goal: drop the inline `flex flex-col gap-4 p-4 bg-muted/40` filter card; use `<FilterBar>`. Replace `useReactTable` + `<Table>` with `<DataList>`. Sorting becomes simpler (server-side or client memo) — for now keep client `sortBy` semantics minimal: title (default) and dueDate ascending toggle. Define a mobile card.

- [ ] **Step 1: Full replacement of the file**

```tsx
"use client";

import { useState } from "react";
import * as React from "react";
import { useQueryClient } from "@tanstack/react-query";
import { useRouter } from "next/navigation";

import { StatusBadge } from "@/components/status-badge";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components";
import { Input } from "@/components/ui/input";
import { Search } from "lucide-react";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { CustomPagination } from "@/components/ui/custom-pagination";
import { DataList, DataListColumn } from "@/components/ui/data-list";
import { FilterBar } from "@/components/ui/filter-bar";
import { useApprovalTasks } from "@/hooks/use-approval-workflow";
import { useDebounce } from "@/hooks/use-debounce";
import { capitalize } from "@/lib/utils";
import { QUERY_KEYS } from "@/lib/constants";
import { WorkflowActionButtons } from "@/components/workflows/workflow-action-buttons";
import {
  claimWorkflowTask,
  approveApprovalTask,
  rejectApprovalTask,
  reassignApprovalTask,
} from "@/app/_actions/workflow-approval-actions";

interface WorkflowTask {
  id: string;
  status: string;
  claimedBy?: string;
  assignedRole?: string;
  assignedUserId?: string;
  assignedTo?: string;
  stageNumber?: number;
  stageName?: string;
  claimExpiry?: string;
  entityType?: string;
  entityId?: string;
  documentType?: string;
  documentId?: string;
  documentNumber?: string;
  title?: string;
  taskType?: string;
  priority?: string;
  dueAt?: string;
  dueDate?: string;
}

const TASK_TYPE_LABELS: Record<string, string> = {
  BUDGET_APPROVAL: "Budget Approval",
  REQUISITION_APPROVAL: "Requisition Approval",
  PURCHASE_ORDER_APPROVAL: "PO Approval",
  PAYMENT_VOUCHER_APPROVAL: "Payment Approval",
  GOODS_RECEIVED_NOTE_CONFIRMATION: "GRN Confirmation",
  GOODS_RECEIVED_NOTE_APPROVAL: "GRN Confirmation",
};

function getTaskTypeLabel(type: string) {
  return (
    TASK_TYPE_LABELS[type] ||
    type?.replace(/_/g, " ").replace(/\b\w/g, (l) => l.toUpperCase()) ||
    "Approval"
  );
}

function getDocumentNumber(t: WorkflowTask) {
  return (
    t.documentNumber ||
    `${t.entityType || t.documentType}-${(t.entityId || t.documentId || "").slice(-3)}`
  );
}

function getTitle(t: WorkflowTask) {
  return (
    t.title ||
    `${capitalize(t.entityType || t.documentType || "").replaceAll("_", " ")} Requires Approval`
  );
}

function PriorityBadge({ p }: { p?: string }) {
  const v = p?.toUpperCase();
  return (
    <Badge
      variant={v === "HIGH" || v === "URGENT" ? "destructive" : v === "MEDIUM" ? "warning" : "info"}
      className="px-2 py-0.5 rounded text-[10px] uppercase font-medium tracking-wider"
    >
      {p || "MEDIUM"}
    </Badge>
  );
}

function DueCell({ t }: { t: WorkflowTask }) {
  const dueDate = t.dueAt || t.dueDate;
  if (!dueDate) return <span className="text-muted-foreground">—</span>;
  const d = new Date(dueDate);
  const overdue = d < new Date() && t.status?.toUpperCase() !== "APPROVED";
  return (
    <span className={overdue ? "text-rose-600 font-medium" : ""}>
      {d.toLocaleDateString()}
      {overdue && <span className="ml-1.5 text-[10px] uppercase">Overdue</span>}
    </span>
  );
}

export function TasksTable() {
  const router = useRouter();
  const queryClient = useQueryClient();
  const [paginationState, setPaginationState] = React.useState({ page: 1, page_size: 10 });
  const [searchQuery, setSearchQuery] = useState("");
  const [statusFilter, setStatusFilter] = useState<string>("all");
  const [documentTypeFilter, setDocumentTypeFilter] = useState<string>("all");
  const [priorityFilter, setPriorityFilter] = useState<string>("all");
  const debouncedSearch = useDebounce(searchQuery, 500);

  const apiFilters = React.useMemo(
    () => ({
      status: statusFilter !== "all" ? (statusFilter.toUpperCase() as any) : undefined,
      documentType: documentTypeFilter !== "all" ? documentTypeFilter : undefined,
      priority: priorityFilter !== "all" ? priorityFilter : undefined,
      assignedToMe: false,
    }),
    [statusFilter, documentTypeFilter, priorityFilter]
  );

  const { data: approvalData, isLoading } = useApprovalTasks(
    apiFilters,
    paginationState.page,
    paginationState.page_size
  );

  const tasks = (approvalData?.data || []) as WorkflowTask[];
  const paginationMeta = approvalData?.pagination;

  const filteredTasks = React.useMemo(() => {
    if (!debouncedSearch) return tasks;
    const s = debouncedSearch.toLowerCase();
    return tasks.filter(
      (t) =>
        t.title?.toLowerCase().includes(s) ||
        t.documentNumber?.toLowerCase().includes(s) ||
        t.stageName?.toLowerCase().includes(s) ||
        t.entityType?.toLowerCase().includes(s)
    );
  }, [tasks, debouncedSearch]);

  const handleClaimTask = React.useCallback(
    async (taskId: string) => {
      const r = await claimWorkflowTask(taskId);
      if (!r.success) throw new Error(r.message);
      await queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.APPROVALS.ALL] });
    },
    [queryClient]
  );
  const handleApproveTask = React.useCallback(
    async (taskId: string, data?: { signature: string; comments: string }) => {
      const r = await approveApprovalTask(taskId, {
        signature: data?.signature || "",
        comments: data?.comments || "Approved",
        stageNumber: 1,
      });
      if (!r.success) throw new Error(r.message);
      await queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.APPROVALS.ALL] });
    },
    [queryClient]
  );
  const handleRejectTask = React.useCallback(
    async (
      taskId: string,
      data?: {
        signature: string;
        comments: string;
        rejectionType?: "reject" | "return_to_draft" | "return_to_previous_stage";
      }
    ) => {
      const r = await rejectApprovalTask(taskId, {
        signature: data?.signature || "",
        remarks: data?.comments || "Rejected",
        rejectionType: data?.rejectionType || "reject",
      });
      if (!r.success) throw new Error(r.message);
      await queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.APPROVALS.ALL] });
    },
    [queryClient]
  );
  const handleReassignTask = React.useCallback(
    async (taskId: string, newUserId: string, reason: string) => {
      const r = await reassignApprovalTask(taskId, { newApproverId: newUserId, reason });
      if (!r.success) throw new Error(r.message);
      await queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.APPROVALS.ALL] });
    },
    [queryClient]
  );
  const handleRefresh = React.useCallback(
    () => queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.APPROVALS.ALL] }),
    [queryClient]
  );

  const columns: DataListColumn<WorkflowTask>[] = [
    {
      id: "title",
      header: "Task",
      cell: (t) => (
        <div className="flex flex-col">
          <span className="font-medium capitalize line-clamp-1">{getTitle(t)}</span>
          <span className="text-xs text-muted-foreground">{getDocumentNumber(t)}</span>
        </div>
      ),
    },
    {
      id: "stageName",
      header: "Stage",
      priority: "md",
      cell: (t) => <span className="text-sm">{t.stageName || "—"}</span>,
    },
    {
      id: "taskType",
      header: "Type",
      priority: "lg",
      cell: (t) => (
        <span className="text-sm">
          {getTaskTypeLabel(
            t.taskType || ((t.entityType || t.documentType || "") as string).toUpperCase() + "_APPROVAL"
          )}
        </span>
      ),
    },
    { id: "priority", header: "Priority", priority: "md", cell: (t) => <PriorityBadge p={t.priority} /> },
    {
      id: "status",
      header: "Status",
      cell: (t) => <StatusBadge status={t.status} type="execution" />,
    },
    { id: "dueAt", header: "Due", priority: "md", cell: (t) => <DueCell t={t} /> },
    {
      id: "actions",
      header: <span className="sr-only">Actions</span>,
      cell: (t) => (
        <WorkflowActionButtons
          task={t as any}
          onClaim={handleClaimTask}
          onApprove={handleApproveTask}
          onReject={handleRejectTask}
          onReassign={handleReassignTask}
          onRefresh={handleRefresh}
          variant="table"
          showViewButton={false}
          onView={(task) => {
            const docType = (task.entityType || task.documentType || "").toLowerCase();
            const docId = task.entityId || task.documentId;
            const routes: Record<string, string> = {
              requisition: `/requisitions/${docId}`,
              purchase_order: `/purchase-orders/${docId}`,
              payment_voucher: `/payment-vouchers/${docId}`,
              grn: `/grn/${docId}`,
              goods_received_note: `/grn/${docId}`,
              budget: `/budgets/${docId}`,
            };
            router.push(routes[docType] || `/tasks/${task.id}`);
          }}
        />
      ),
    },
  ];

  const clearFilters = () => {
    setSearchQuery("");
    setStatusFilter("all");
    setDocumentTypeFilter("all");
    setPriorityFilter("all");
  };
  const hasActiveFilters =
    Boolean(searchQuery) ||
    statusFilter !== "all" ||
    documentTypeFilter !== "all" ||
    priorityFilter !== "all";

  return (
    <div className="space-y-4">
      <FilterBar
        search={
          <div className="relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
            <Input
              placeholder="Search tasks…"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="pl-10"
            />
          </div>
        }
        filters={
          <>
            <Select value={statusFilter} onValueChange={setStatusFilter}>
              <SelectTrigger className="w-[140px]">
                <SelectValue placeholder="Status" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All statuses</SelectItem>
                <SelectItem value="pending">Pending</SelectItem>
                <SelectItem value="claimed">Claimed</SelectItem>
                <SelectItem value="completed">Completed</SelectItem>
              </SelectContent>
            </Select>
            <Select value={documentTypeFilter} onValueChange={setDocumentTypeFilter}>
              <SelectTrigger className="w-[160px]">
                <SelectValue placeholder="Document" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All types</SelectItem>
                <SelectItem value="requisition">Requisition</SelectItem>
                <SelectItem value="purchase_order">Purchase Order</SelectItem>
                <SelectItem value="payment_voucher">Payment Voucher</SelectItem>
                <SelectItem value="grn">GRN</SelectItem>
                <SelectItem value="budget">Budget</SelectItem>
              </SelectContent>
            </Select>
            <Select value={priorityFilter} onValueChange={setPriorityFilter}>
              <SelectTrigger className="w-[130px]">
                <SelectValue placeholder="Priority" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All priorities</SelectItem>
                <SelectItem value="urgent">Urgent</SelectItem>
                <SelectItem value="high">High</SelectItem>
                <SelectItem value="medium">Medium</SelectItem>
                <SelectItem value="low">Low</SelectItem>
              </SelectContent>
            </Select>
          </>
        }
        hasActiveFilters={hasActiveFilters}
        onReset={clearFilters}
        meta={`${filteredTasks.length} task${filteredTasks.length === 1 ? "" : "s"}${hasActiveFilters ? " (filtered)" : ""}`}
      />

      <DataList<WorkflowTask>
        rows={filteredTasks}
        columns={columns}
        getRowId={(t) => t.id}
        isLoading={isLoading}
        emptyMessage="No tasks found."
        mobileCard={(t) => (
          <div className="flex flex-col gap-2">
            <div className="flex items-start justify-between gap-2">
              <div className="min-w-0">
                <div className="font-medium capitalize line-clamp-1">{getTitle(t)}</div>
                <div className="text-xs text-muted-foreground">{getDocumentNumber(t)}</div>
              </div>
              <StatusBadge status={t.status} type="execution" />
            </div>
            <div className="flex flex-wrap items-center gap-2 text-xs text-muted-foreground">
              <PriorityBadge p={t.priority} />
              {t.stageName && <span>{t.stageName}</span>}
              <DueCell t={t} />
            </div>
            <div className="pt-1">
              <WorkflowActionButtons
                task={t as any}
                onClaim={handleClaimTask}
                onApprove={handleApproveTask}
                onReject={handleRejectTask}
                onReassign={handleReassignTask}
                onRefresh={handleRefresh}
                variant="compact"
                showViewButton
                onView={(task) => {
                  const docType = (task.entityType || task.documentType || "").toLowerCase();
                  const docId = task.entityId || task.documentId;
                  const routes: Record<string, string> = {
                    requisition: `/requisitions/${docId}`,
                    purchase_order: `/purchase-orders/${docId}`,
                    payment_voucher: `/payment-vouchers/${docId}`,
                    grn: `/grn/${docId}`,
                    goods_received_note: `/grn/${docId}`,
                    budget: `/budgets/${docId}`,
                  };
                  router.push(routes[docType] || `/tasks/${task.id}`);
                }}
              />
            </div>
          </div>
        )}
      />

      {paginationMeta && (
        <CustomPagination
          pagination={{
            ...paginationMeta,
            page: paginationState.page,
            page_size: paginationState.page_size,
            limit: paginationMeta.limit || paginationState.page_size,
            totalCount: paginationMeta.totalCount || paginationMeta.total || 0,
            total_pages: paginationMeta.total_pages || paginationMeta.totalPages || 0,
            has_next: paginationMeta.has_next ?? paginationMeta.hasNext ?? false,
            has_prev: paginationMeta.has_prev ?? paginationMeta.hasPrev ?? false,
          }}
          updatePagination={(np: { page: number; page_size?: number }) =>
            setPaginationState((prev) => ({
              ...prev,
              page: np.page,
              page_size: np.page_size || prev.page_size,
            }))
          }
          allowSetPageSize
          showDetails
        />
      )}
    </div>
  );
}
```

- [ ] **Step 2: Verify TS compiles**

Run: `cd frontend && pnpm tsc --noEmit`
Expected: No errors. (If `WorkflowActionButtons` `compact` variant lacks a row-friendly mode in mobile, fall back to `inline` — confirm by reading `frontend/src/components/workflows/workflow-action-buttons.tsx`.)

- [ ] **Step 3: Visual smoke**

Run: `cd frontend && pnpm dev`
Open `/tasks?tab=tasks` at desktop and mobile widths.
Expected:
- Desktop ≥1024px: full table, columns Task / Stage / Type / Priority / Status / Due / Actions
- 768–1023px: table without Type column
- <768px: stacked cards, each showing title, doc number, status, priority/stage/due meta row, action buttons

- [ ] **Step 4: Commit**

```bash
git add frontend/src/app/\(private\)/\(main\)/tasks/_components/tasks-table.tsx
git commit -m "refactor(tasks): convert TasksTable to DataList + FilterBar"
```

---

## Task 10: Refactor `approvals-list.tsx` to use shared primitives

**Files:**
- Modify: `frontend/src/app/(private)/(main)/tasks/_components/approvals-list.tsx`

Drop the local `StatCell` (now provided by `StatGrid`). Drop the inline filter card (use `FilterBar`). Keep `TaskGroup` local (it is approval-specific). Replace raw `bg-blue-100`/`bg-emerald-100` group title chips with the `BADGE_CLASSES` map but keep them — those are already reasonable. Remove the duplicate top-level page header (`<h2>Approval Tasks</h2>` block) — the parent `tasks-client.tsx` `<PageHeader>` already serves the page; the inner heading is redundant.

- [ ] **Step 1: Apply edits**

Replace lines 5–9 imports:
```tsx
import { useApprovalTasks } from "@/hooks/use-approval-workflow";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { SelectField } from "@/components/ui/select-field";
import Search from "@/components/ui/search-field";
```
with:
```tsx
import { useApprovalTasks } from "@/hooks/use-approval-workflow";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { SelectField } from "@/components/ui/select-field";
import Search from "@/components/ui/search-field";
import { StatGrid } from "@/components/ui/stat-grid";
import { FilterBar } from "@/components/ui/filter-bar";
```

Replace the entire `return (...)` JSX (the body from `return (` after the `if (error)` block to the closing of the function) with:

```tsx
  return (
    <div className="space-y-4">
      {/* Header row: refresh action only — page title is provided by parent PageHeader */}
      <div className="flex items-center justify-end">
        <Button
          onClick={handleRefresh}
          variant="outline"
          size="sm"
          disabled={isTasksLoading}
          className="gap-2"
        >
          <RefreshCw className={cn("h-3.5 w-3.5", isTasksLoading && "animate-spin")} />
          Refresh
        </Button>
      </div>

      <StatGrid
        items={[
          { label: "Claimed by Me", value: stats.claimedByMe, icon: <Users className="h-3 w-3 sm:h-4 sm:w-4" />, accent: "blue" },
          { label: "Available", value: stats.available, icon: <Clock className="h-3 w-3 sm:h-4 sm:w-4" />, accent: "emerald" },
          { label: "Claimed by Others", value: stats.claimedByOthers, icon: <UserCheck className="h-3 w-3 sm:h-4 sm:w-4" />, accent: "amber" },
          { label: "Completed", value: stats.completed, icon: <CheckCircle2 className="h-3 w-3 sm:h-4 sm:w-4" />, accent: "slate" },
          { label: "Total (view)", value: stats.total, icon: <ListFilter className="h-3 w-3 sm:h-4 sm:w-4" />, accent: "violet" },
        ]}
      />

      <FilterBar
        search={
          <Search
            placeholder="Search by document number, type, or stage…"
            value={searchQuery}
            onChange={(v) => setSearchQuery(v)}
            isClearable
          />
        }
        filters={
          <>
            <SelectField
              placeholder="Status"
              classNames={{ wrapper: "md:w-44" }}
              value={statusFilter}
              onValueChange={(v) => setStatusFilter(v as StatusFilter)}
              options={[
                { value: "all", label: "All statuses" },
                { value: "pending", label: "Available" },
                { value: "claimed", label: "Claimed" },
                { value: "completed", label: "Completed" },
              ]}
            />
            <SelectField
              placeholder="Priority"
              classNames={{ wrapper: "md:w-40" }}
              value={priorityFilter}
              onValueChange={(v) => setPriorityFilter(v as PriorityFilter)}
              options={[
                { value: "all", label: "All priorities" },
                { value: "HIGH", label: "High" },
                { value: "MEDIUM", label: "Medium" },
                { value: "LOW", label: "Low" },
              ]}
            />
            <SelectField
              placeholder="Sort"
              classNames={{ wrapper: "md:w-40" }}
              value={sortBy}
              onValueChange={(v) => setSortBy(v as SortBy)}
              options={[
                { value: "date", label: "Newest" },
                { value: "priority", label: "Priority" },
                { value: "name", label: "Document" },
              ]}
            />
          </>
        }
        hasActiveFilters={hasActiveFilters}
        onReset={clearFilters}
        meta={`Showing ${stats.total} task${stats.total !== 1 ? "s" : ""}`}
      />

      {/* Task groups */}
      <div className="space-y-6">
        {groupedTasks.claimedByMe.length > 0 && (
          <TaskGroup title="Claimed by You" count={groupedTasks.claimedByMe.length} accent="blue">
            {groupedTasks.claimedByMe.map((task) => (
              <ApprovalTaskCard
                key={task.id}
                taskId={task.id}
                currentUserId={currentUser.id}
                currentUserRole={currentUser.role}
              />
            ))}
          </TaskGroup>
        )}

        {groupedTasks.available.length > 0 && (
          <TaskGroup title="Available Tasks" count={groupedTasks.available.length} accent="emerald">
            {groupedTasks.available.map((task) => (
              <ApprovalTaskCard
                key={task.id}
                taskId={task.id}
                currentUserId={currentUser.id}
                currentUserRole={currentUser.role}
              />
            ))}
          </TaskGroup>
        )}

        {groupedTasks.claimedByOthers.length > 0 && (
          <TaskGroup title="Claimed by Others" count={groupedTasks.claimedByOthers.length} accent="amber">
            {groupedTasks.claimedByOthers.map((task) => (
              <ApprovalTaskCard
                key={task.id}
                taskId={task.id}
                currentUserId={currentUser.id}
                currentUserRole={currentUser.role}
              />
            ))}
          </TaskGroup>
        )}

        {groupedTasks.completed.length > 0 && (
          <TaskGroup title="Completed" count={groupedTasks.completed.length} accent="slate">
            {groupedTasks.completed.map((task) => (
              <ApprovalTaskCard
                key={task.id}
                taskId={task.id}
                currentUserId={currentUser.id}
                currentUserRole={currentUser.role}
              />
            ))}
          </TaskGroup>
        )}

        {filteredTasks.length === 0 && !isTasksLoading && (
          <Card className="border-dashed border-border/60">
            <CardContent className="py-10 text-center">
              <Inbox className="h-8 w-8 text-muted-foreground/60 mx-auto mb-3" />
              <p className="font-medium text-sm mb-1">No approval tasks</p>
              <p className="text-xs text-muted-foreground mb-3">
                {hasActiveFilters
                  ? "No tasks match your current filters."
                  : "There are no approval tasks assigned to your role right now."}
              </p>
              {hasActiveFilters && (
                <Button variant="outline" size="sm" onClick={clearFilters}>
                  Reset filters
                </Button>
              )}
            </CardContent>
          </Card>
        )}

        {isTasksLoading && (
          <div className="grid gap-3">
            {[1, 2, 3].map((i) => (
              <Card key={i} className="p-4 border-border/60">
                <div className="animate-pulse space-y-3">
                  <div className="flex items-center gap-3">
                    <div className="h-9 w-9 rounded-md bg-muted" />
                    <div className="flex-1 space-y-1.5">
                      <div className="h-4 bg-muted rounded w-48" />
                      <div className="h-3 bg-muted rounded w-32" />
                    </div>
                  </div>
                  <div className="h-3 bg-muted rounded w-3/4" />
                </div>
              </Card>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
```

Then delete the local `StatCell` + `CHIP_CLASSES` constants (the `Accent` type, `BADGE_CLASSES`, and `TaskGroup` stay). Specifically delete the block from `// ── Sub-components ──` down through the end of `function StatCell(...)` (keep `TaskGroup` and the `BADGE_CLASSES` map it depends on).

Final tail of the file should look like:
```tsx
// ── Sub-components ──────────────────────────────────────────────────────────

type Accent = "blue" | "emerald" | "amber" | "slate" | "violet";

const BADGE_CLASSES: Record<Accent, string> = {
  blue: "bg-blue-600 text-white",
  emerald: "bg-emerald-600 text-white",
  amber: "bg-amber-600 text-white",
  slate: "bg-slate-600 text-white",
  violet: "bg-violet-600 text-white",
};

function TaskGroup({
  title,
  count,
  accent,
  children,
}: {
  title: string;
  count: number;
  accent: Accent;
  children: React.ReactNode;
}) {
  return (
    <section className="space-y-2">
      <div className="flex items-center gap-2">
        <h3 className="text-sm font-semibold">{title}</h3>
        <Badge className={cn("text-xs", BADGE_CLASSES[accent])}>{count}</Badge>
      </div>
      <div className="grid gap-2.5">{children}</div>
    </section>
  );
}
```

- [ ] **Step 2: Verify TS compiles**

Run: `cd frontend && pnpm tsc --noEmit`
Expected: No errors. The `AlertCircle` import becomes unused — remove it from the lucide-react import line.

- [ ] **Step 3: Visual smoke**

Run: `cd frontend && pnpm dev` and open `/tasks?tab=approvals` mobile + desktop.
Expected: One stat strip (no duplicates), one filter bar, refresh button right-aligned, task groups render unchanged.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/app/\(private\)/\(main\)/tasks/_components/approvals-list.tsx
git commit -m "refactor(tasks): ApprovalsList uses StatGrid + FilterBar"
```

---

## Task 11: Polish `tasks-client.tsx` — unified header and tab styling

**Files:**
- Modify: `frontend/src/app/(private)/(main)/tasks/_components/tasks-client.tsx`

- [ ] **Step 1: Update tabs to scroll-friendly mobile pills with token-driven active state**

Full replacement:

```tsx
"use client";

import { useState, useEffect } from "react";
import { useSearchParams, useRouter, usePathname } from "next/navigation";
import { PageHeader } from "@/components/base/page-header";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { TasksTable } from "./tasks-table";
import { TaskStatsCards } from "./task-stats-cards";
import { ApprovalsList } from "./approvals-list";

interface TasksClientProps {
  userId: string;
  userRole: string;
}

type TabValue = "tasks" | "approvals";

export function TasksClient({ userId, userRole }: TasksClientProps) {
  const searchParams = useSearchParams();
  const router = useRouter();
  const pathname = usePathname();
  const [refreshTrigger] = useState(0);
  const [activeTab, setActiveTab] = useState<TabValue>("tasks");

  useEffect(() => {
    const tabParam = searchParams.get("tab");
    if (tabParam === "approvals" || tabParam === "tasks") {
      setActiveTab(tabParam);
    }
  }, [searchParams]);

  const handleTabChange = (value: string) => {
    const v = value as TabValue;
    setActiveTab(v);
    // Keep URL in sync so refreshes/back-button preserve tab state.
    const params = new URLSearchParams(searchParams.toString());
    params.set("tab", v);
    router.replace(`${pathname}?${params.toString()}`, { scroll: false });
  };

  return (
    <div className="space-y-5">
      <PageHeader
        title="Workflows"
        subtitle="Tasks and approvals assigned to you"
        showBackButton={false}
      />

      <Tabs value={activeTab} onValueChange={handleTabChange} className="space-y-5">
        <TabsList className="inline-flex h-9 w-full sm:w-auto bg-muted/60 p-1 rounded-lg">
          <TabsTrigger
            value="tasks"
            className="flex-1 sm:flex-initial sm:px-6 data-[state=active]:bg-background data-[state=active]:text-foreground data-[state=active]:shadow-sm rounded-md"
          >
            Tasks
          </TabsTrigger>
          <TabsTrigger
            value="approvals"
            className="flex-1 sm:flex-initial sm:px-6 data-[state=active]:bg-background data-[state=active]:text-foreground data-[state=active]:shadow-sm rounded-md"
          >
            Approvals
          </TabsTrigger>
        </TabsList>

        <TabsContent value="tasks" className="space-y-4 mt-0">
          <TaskStatsCards userId={userId} refreshTrigger={refreshTrigger} />
          <TasksTable />
        </TabsContent>

        <TabsContent value="approvals" className="space-y-4 mt-0">
          <ApprovalsList userId={userId} userRole={userRole} />
        </TabsContent>
      </Tabs>
    </div>
  );
}
```

- [ ] **Step 2: Verify TS compiles**

Run: `cd frontend && pnpm tsc --noEmit`
Expected: No errors.

- [ ] **Step 3: Visual smoke**

Run: `cd frontend && pnpm dev` and open `/tasks` mobile + desktop.
Expected:
- Tabs are full-width on mobile, auto-width on `sm+`
- Switching tabs updates `?tab=` in the URL
- Active tab uses background-elevated pill
- Stats strip + filter bar + content render consistently across both tabs

- [ ] **Step 4: Commit**

```bash
git add frontend/src/app/\(private\)/\(main\)/tasks/_components/tasks-client.tsx
git commit -m "refactor(tasks): unified tab shell, mobile-friendly pills, ?tab= sync"
```

---

## Task 12: Verification pass

- [ ] **Step 1: Full type-check + tests**

Run:
```bash
cd frontend && pnpm tsc --noEmit && pnpm vitest run
```
Expected: All pass.

- [ ] **Step 2: Manual mobile QA checklist**

Open DevTools mobile emulation (390×844, iPhone 14). Walk through:
- [ ] `/home` — bottom nav visible, 4 tabs + More
- [ ] `/tasks` — tabs full-width, swipe through Tasks then Approvals
- [ ] `/tasks` Tasks tab — stats grid 2-col, filter bar wraps, task list shows mobile cards (no horizontal scroll), action buttons tappable (≥36px)
- [ ] `/tasks` Approvals tab — single stats strip, single filter bar, task groups visible, refresh button reachable
- [ ] `/tasks` More drawer — opens from bottom nav, closes on back-tap, links navigate

- [ ] **Step 3: Final commit if any cleanup**

```bash
git add -A
git commit -m "chore(ui): plan-A verification cleanup"
```

---

## Follow-up Plans (do NOT include in this plan)

- **Plan B — Modal sweep:** Migrate `approval-action-modal.tsx`, `digital-signature-pad` consumers, `create-pv-from-po-dialog.tsx`, GRN receive modal, and form modals across requisition/PO/PV/GRN to `<ResponsiveSheet>`. One sub-task per consumer.
- **Plan C — Tables:** Convert requisitions, purchase-orders, payment-vouchers, GRN, users, organizations tables to `<DataList>` with mobile cards. Audit column priorities per table.
- **Plan D — Dashboard:** Tighten greeting card, swap metrics to `<StatGrid>`, convert quick actions to horizontal chip rail on mobile, role-variant polish, extract `MORE_LINKS` to a shared nav config consumed by both `nav-main.tsx` and `mobile-bottom-nav.tsx`.
- **Plan E — Approval flow polish:** Task detail page redesign (side-by-side document review on `lg`, tabbed on mobile), signature pad full-width canvas + haptic press, vertical stage stepper, claim countdown chip.

Each follow-up plan is independently shippable and uses primitives this plan introduces.
