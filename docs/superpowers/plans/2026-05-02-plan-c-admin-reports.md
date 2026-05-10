# UI Revamp — Plan C: `/admin/reports` End-to-End Refactor

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Refactor every component under `/admin/reports` to consume Plan B primitives, add a global URL-bound date range, replace the per-tab Export switch with a `DropdownMenu`, and prove the new vocabulary works end-to-end on a real route.

**Architecture:** The admin reports route currently has a tab shell (`AdminReportsClient`) plus three sibling client components (`SystemStatistics`, `ApprovalReports`, `UserActivityReports`) and an analytics tab (`AnalyticsDashboard`). Each builds its own bespoke `<Card>` metric tiles and raw `<Table>`s. Plan C swaps every metric tile to `<MetricCard>`, every table to `<DataList>`, every chart to `<ReportChart>`, the inline `<h1>` to `<PageHeader>`, and adds a global date range picker bound via `useDateRangeUrlState` that drips down to the four React Query hooks (which already accept a `dateRange` param). Two Plan B carry-forwards land first: `useDateRangeUrlState` no-op guard + searchParams stability, and `ReportChart` generic-typing tightening — both used heavily in this plan.

**Tech Stack:** Next.js 15 App Router, TypeScript, Tailwind v4, ShadCN UI, recharts (via Plan B `ReportChart`), `useQuery` + existing `useSystemStats`/`useApprovalMetrics`/`useUserActivity`/`useAnalyticsDashboard`, Vitest.

**Out-of-scope (handled in later plans):**
- Plan D — modal sweep across procurement
- Plan E — list tables sweep across procurement
- Plan F — detail-page split + DetailShell adoption
- Plan G — vendor detail page + non-admin reports hub decision
- Plan H — approval flow polish
- Plan I — cleanup / stub removal

---

## Existing infrastructure (reuse)

- `MetricCard` (`ui/metric-card.tsx`) — Plan B Task 3
- `DataList`, `FilterBar`, `StatGrid` — Plan A
- `ReportChart` (`ui/report-chart.tsx`) — Plan B Task 5
- `EmptyState` (`base/empty-state.tsx`) — Plan B Task 8 (now has `action` slot)
- `PageHeader` (`base/page-header.tsx`) — already widely used
- `CalendarDateRangePicker` (`ui/date-range-picker.tsx`) — Plan B Task 9 (now has `last90Days` + `lastQuarter` presets)
- `useDateRangeUrlState` (`hooks/use-date-range-url-state.ts`) — Plan B Task 10
- `StatusBadge` (`components/status-badge.tsx`) — already supports document/approval/execution types
- `DocumentTypeChip` (`ui/document-type-chip.tsx`) — Plan B Task 4
- `useSystemStats`, `useApprovalMetrics`, `useUserActivity`, `useAnalyticsDashboard` — already accept `DateRange` param (`{ from: string; to: string }`)
- `exportSystemStatsToCSV`, `exportApprovalMetricsToCSV`, `exportUserActivityToCSV`, `exportAnalyticsDashboardToCSV` — exist in `lib/export-utils`
- `DropdownMenu*` — exists at `ui/dropdown-menu.tsx`

---

## File Structure

**Create:**
- `frontend/src/app/(private)/admin/_components/reports-header.tsx` — `<ReportsHeader>` containing `<PageHeader>` + date-range picker + refresh + export-menu actions; reads/writes URL state via `useDateRangeUrlState`

**Modify:**
- `frontend/src/hooks/use-date-range-url-state.ts` — add no-op guard in `setRange`; depend on `searchParams.toString()` to stabilize `setRange` identity (Plan B carry-forward)
- `frontend/src/components/ui/report-chart.tsx` — tighten generics: `xKey: keyof T & string`, `series[].dataKey: keyof T & string` (Plan B carry-forward)
- `frontend/src/app/(private)/admin/_components/admin-reports-client.tsx` — drop inline `<h1>`/Refresh/Export; mount `<ReportsHeader>`; thread `dateRange` down to children; tabs adopt Plan A pill styling
- `frontend/src/app/(private)/admin/_components/system-statistics.tsx` — 4 bespoke cards → `MetricCard` grid; doc-type bar chart → `<ReportChart kind="bar" perBarColor>`; status summary list → status pills using `<StatusBadge>`; consume `dateRange` prop
- `frontend/src/app/(private)/admin/_components/approval-reports.tsx` — 3 bespoke cards → `MetricCard`; inline search → `FilterBar`; raw `<Table>` → `<DataList>` with mobile card; replace inline `formatDocumentType` with `<DocumentTypeChip>`; consume `dateRange` prop
- `frontend/src/app/(private)/admin/_components/user-activity-reports.tsx` — 3 bespoke cards → `MetricCard`; bespoke initials avatar → `<UserAvatar>`; raw `<Table>` → `<DataList>` with mobile card; `FilterBar` for role + search; consume `dateRange` prop
- `frontend/src/components/workflows/analytics-dashboard.tsx` — swap raw recharts to `<ReportChart>` for visual consistency; consume `dateRange` prop

**Delete:**
- `frontend/src/app/(private)/admin/reports/loading.tsx` — dead (sub-components render their own loading states; route loading never shows)

**Test files:**
- `frontend/src/__tests__/components/admin/reports-header.test.tsx` (new)
- The four refactored components don't gain new tests in Plan C (they're orchestration / data-display; consumed primitives are already tested). Add a single integration smoke test that mounts `AdminReportsClient` with mocked hooks to confirm tab switching + date-range URL sync still work.
- `frontend/src/__tests__/integration/admin-reports.test.tsx` (new)

---

## Task 1: Plan B carry-forward — `setRange` no-op guard + searchParams stability

**Files:**
- Modify: `frontend/src/hooks/use-date-range-url-state.ts`
- Modify: `frontend/src/__tests__/hooks/use-date-range-url-state.test.ts`

- [ ] **Step 1: Extend the existing test**

Open `frontend/src/__tests__/hooks/use-date-range-url-state.test.ts`. After the existing 3 `it()` blocks (inside the existing `describe`), add:

```ts
  it("setRange does NOT call router.replace when from/to are unchanged", () => {
    const { result } = renderHook(() =>
      useDateRangeUrlState({ defaultFrom: "2026-01-01", defaultTo: "2026-01-31" })
    );
    act(() => result.current.setRange("2026-01-01", "2026-01-31"));
    expect(replace).not.toHaveBeenCalled();
  });

  it("setRange identity is stable across renders when params are unchanged", () => {
    const { result, rerender } = renderHook(() =>
      useDateRangeUrlState({ defaultFrom: "2026-01-01", defaultTo: "2026-01-31" })
    );
    const first = result.current.setRange;
    rerender();
    expect(result.current.setRange).toBe(first);
  });
```

- [ ] **Step 2: Run test, expect first to fail (no-op guard not yet in place)**

Run: `cd frontend && pnpm vitest run src/__tests__/hooks/use-date-range-url-state.test.ts`
Expected: 4 pass + 1 fail (the no-op guard test). The identity-stability test may pass already or fail depending on `searchParams` reference; both are addressed by Step 3.

- [ ] **Step 3: Update hook**

Replace the entire body of `frontend/src/hooks/use-date-range-url-state.ts` with:

```ts
"use client";
import { useCallback } from "react";
import { usePathname, useRouter, useSearchParams } from "next/navigation";

export interface UseDateRangeUrlStateOptions {
  /** Fallback ISO `YYYY-MM-DD` when `?from` is absent. */
  defaultFrom: string;
  /** Fallback ISO `YYYY-MM-DD` when `?to` is absent. */
  defaultTo: string;
  /** Param key for from (default "from"). */
  fromKey?: string;
  /** Param key for to (default "to"). */
  toKey?: string;
}

export interface UseDateRangeUrlStateResult {
  from: string;
  to: string;
  setRange: (from: string, to: string) => void;
}

export function useDateRangeUrlState({
  defaultFrom,
  defaultTo,
  fromKey = "from",
  toKey = "to",
}: UseDateRangeUrlStateOptions): UseDateRangeUrlStateResult {
  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();
  const searchString = searchParams.toString();

  const from = searchParams.get(fromKey) || defaultFrom;
  const to = searchParams.get(toKey) || defaultTo;

  const setRange = useCallback(
    (newFrom: string, newTo: string) => {
      const params = new URLSearchParams(searchString);
      const currentFrom = params.get(fromKey) || defaultFrom;
      const currentTo = params.get(toKey) || defaultTo;
      if (newFrom === currentFrom && newTo === currentTo) {
        return;
      }
      params.set(fromKey, newFrom);
      params.set(toKey, newTo);
      router.replace(`${pathname}?${params.toString()}`, { scroll: false });
    },
    [router, pathname, searchString, fromKey, toKey, defaultFrom, defaultTo]
  );

  return { from, to, setRange };
}
```

Key changes:
- `searchString = searchParams.toString()` — stable string used as dep, replacing `searchParams` object dep that changed identity each render in some Next.js versions.
- No-op guard inside `setRange` short-circuits when both values match the current URL state.

- [ ] **Step 4: Run test, expect all pass**

Run: `cd frontend && pnpm vitest run src/__tests__/hooks/use-date-range-url-state.test.ts`
Expected: 5 pass.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/hooks/use-date-range-url-state.ts frontend/src/__tests__/hooks/use-date-range-url-state.test.ts
git commit -m "feat(hooks): useDateRangeUrlState gains no-op guard + stable identity"
```

---

## Task 2: Plan B carry-forward — `ReportChart` generic tightening

**Files:**
- Modify: `frontend/src/components/ui/report-chart.tsx`

The current types accept `xKey: string` and `series[].dataKey: string` regardless of `T`. Tighten to `keyof T & string` so consumers get compile-time errors on typos.

- [ ] **Step 1: Update the type signatures**

Open `frontend/src/components/ui/report-chart.tsx`. Replace the type and prop definitions (the section from `export interface ReportSeries` through `export interface ReportChartProps<...>`) with:

```tsx
export interface ReportSeries<T = Record<string, unknown>> {
  /** Key in each data row to plot. Constrained to `keyof T` when `T` is supplied. */
  dataKey: keyof T & string;
  /** Display label. */
  label: string;
  /** Optional explicit color override. Defaults to chart-1..5 cycling. */
  color?: string;
}

export interface ReportChartProps<T extends Record<string, unknown> = Record<string, unknown>> {
  kind: ReportChartKind;
  data: T[];
  /** Key in each data row to use as the X axis label. */
  xKey: keyof T & string;
  series: ReportSeries<T>[];
  /** Tailwind classes for the outer container; defaults to a sensible aspect. */
  className?: string;
  /** Show legend. Default false (single-series charts don't need it). */
  showLegend?: boolean;
  /** For bar charts: per-bar color from palette instead of single hue. Default false. */
  perBarColor?: boolean;
}
```

The function signature stays the same (`export function ReportChart<T extends Record<string, unknown>>(...)`); the body is unchanged. The `series.forEach` call inside `useMemo` already references `s.dataKey` as a string-indexable key, which still works with the tighter type.

- [ ] **Step 2: Verify TS compiles**

Run: `cd frontend && pnpm tsc --noEmit`
Expected: clean.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/ui/report-chart.tsx
git commit -m "feat(ui): ReportChart tightens generics to keyof T"
```

---

## Task 3: Build `<ReportsHeader>`

**Files:**
- Create: `frontend/src/app/(private)/admin/_components/reports-header.tsx`
- Test: `frontend/src/__tests__/components/admin/reports-header.test.tsx`

`<ReportsHeader>` is a presentation+behavior component owning the page title, the global date range picker, the refresh button, and the export menu. It calls `useDateRangeUrlState` for storage and accepts `onRefresh` + `onExport(format)` callbacks from the parent.

- [ ] **Step 1: Write failing test**

```tsx
// frontend/src/__tests__/components/admin/reports-header.test.tsx
import { render, screen, fireEvent } from "@testing-library/react";
import { describe, it, expect, vi } from "vitest";
import { ReportsHeader } from "@/app/(private)/admin/_components/reports-header";

vi.mock("next/navigation", () => ({
  useRouter: () => ({ replace: vi.fn() }),
  usePathname: () => "/admin/reports",
  useSearchParams: () => new URLSearchParams(),
}));

describe("ReportsHeader", () => {
  it("renders title and subtitle", () => {
    render(
      <ReportsHeader
        title="Admin Reports"
        subtitle="Workflow approvals, user activity, system metrics"
        onRefresh={vi.fn()}
        onExport={vi.fn()}
        isRefreshing={false}
      />
    );
    expect(screen.getByText("Admin Reports")).toBeInTheDocument();
    expect(
      screen.getByText("Workflow approvals, user activity, system metrics")
    ).toBeInTheDocument();
  });

  it("calls onRefresh when refresh button is clicked", () => {
    const onRefresh = vi.fn();
    render(
      <ReportsHeader
        title="X"
        subtitle="Y"
        onRefresh={onRefresh}
        onExport={vi.fn()}
        isRefreshing={false}
      />
    );
    fireEvent.click(screen.getByRole("button", { name: /refresh/i }));
    expect(onRefresh).toHaveBeenCalledTimes(1);
  });

  it("disables refresh when isRefreshing is true", () => {
    render(
      <ReportsHeader
        title="X"
        subtitle="Y"
        onRefresh={vi.fn()}
        onExport={vi.fn()}
        isRefreshing
      />
    );
    expect(screen.getByRole("button", { name: /refresh/i })).toBeDisabled();
  });

  it("invokes onExport with chosen format from menu", () => {
    const onExport = vi.fn();
    render(
      <ReportsHeader
        title="X"
        subtitle="Y"
        onRefresh={vi.fn()}
        onExport={onExport}
        isRefreshing={false}
      />
    );
    fireEvent.click(screen.getByRole("button", { name: /export/i }));
    fireEvent.click(screen.getByText(/csv/i));
    expect(onExport).toHaveBeenCalledWith("csv");
  });
});
```

- [ ] **Step 2: Run test, expect failure**

Run: `cd frontend && pnpm vitest run src/__tests__/components/admin/reports-header.test.tsx`
Expected: FAIL — module not found.

- [ ] **Step 3: Implement `<ReportsHeader>`**

```tsx
// frontend/src/app/(private)/admin/_components/reports-header.tsx
"use client";
import * as React from "react";
import { format, subDays, startOfDay, endOfDay } from "date-fns";
import { Download, RefreshCw } from "lucide-react";
import { Button } from "@/components/ui/button";
import { PageHeader } from "@/components/base/page-header";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import CalendarDateRangePicker from "@/components/ui/date-range-picker";
import { useDateRangeUrlState } from "@/hooks/use-date-range-url-state";
import { cn } from "@/lib/utils";

export type ReportsExportFormat = "csv";

export interface ReportsHeaderProps {
  title: string;
  subtitle: string;
  onRefresh: () => void;
  onExport: (format: ReportsExportFormat) => void;
  isRefreshing: boolean;
  className?: string;
}

function defaultRange() {
  const today = new Date();
  const from = format(startOfDay(subDays(today, 27)), "yyyy-MM-dd");
  const to = format(endOfDay(today), "yyyy-MM-dd");
  return { from, to };
}

export function ReportsHeader({
  title,
  subtitle,
  onRefresh,
  onExport,
  isRefreshing,
  className,
}: ReportsHeaderProps) {
  const initial = React.useMemo(defaultRange, []);
  const { from, to, setRange } = useDateRangeUrlState({
    defaultFrom: initial.from,
    defaultTo: initial.to,
  });

  const initialFromDate = React.useMemo(() => new Date(from), [from]);
  const initialToDate = React.useMemo(() => new Date(to), [to]);

  return (
    <div className={cn("space-y-3", className)}>
      <PageHeader
        title={title}
        subtitle={subtitle}
        showBackButton={false}
        actions={
          <div className="flex flex-wrap items-center gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={onRefresh}
              disabled={isRefreshing}
            >
              <RefreshCw
                className={cn("h-4 w-4 mr-2", isRefreshing && "animate-spin")}
              />
              Refresh
            </Button>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="outline" size="sm">
                  <Download className="h-4 w-4 mr-2" />
                  Export
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <DropdownMenuItem onClick={() => onExport("csv")}>
                  Export current view (CSV)
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        }
      />
      <div className="flex items-center gap-2">
        <span className="text-xs font-medium text-muted-foreground uppercase tracking-wider">
          Date range
        </span>
        <div className="flex-1 max-w-sm">
          <CalendarDateRangePicker
            initialFrom={initialFromDate}
            initialTo={initialToDate}
            onChange={(newFrom, newTo) => setRange(newFrom, newTo)}
          />
        </div>
      </div>
    </div>
  );
}
```

- [ ] **Step 4: Run test, expect pass**

Run: `cd frontend && pnpm vitest run src/__tests__/components/admin/reports-header.test.tsx`
Expected: 4 pass.

- [ ] **Step 5: Commit**

```bash
git add "frontend/src/app/(private)/admin/_components/reports-header.tsx" "frontend/src/__tests__/components/admin/reports-header.test.tsx"
git commit -m "feat(admin): add ReportsHeader with date range + export menu"
```

---

## Task 4: Refactor `system-statistics.tsx`

**Files:**
- Modify: `frontend/src/app/(private)/admin/_components/system-statistics.tsx`

Drop bespoke 4-up `<Card>` blocks → `MetricCard` row. Doc-type bar chart → `<ReportChart kind="bar" perBarColor>`. Status summary list → simpler list using `<StatusBadge type="document">` for each row. Consume `dateRange` prop, pass to `useSystemStats`.

- [ ] **Step 1: Replace the file**

Full replacement:

```tsx
// frontend/src/app/(private)/admin/_components/system-statistics.tsx
"use client";

import { Alert, AlertDescription } from "@/components/ui/alert";
import { useSystemStats } from "@/hooks/use-reports-queries";
import {
  FileText,
  Clock,
  CheckCircle2,
  AlertCircle,
  TrendingUp,
} from "lucide-react";
import { MetricCard } from "@/components/ui/metric-card";
import { ReportChart } from "@/components/ui/report-chart";
import { StatusBadge } from "@/components/status-badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import EmptyState from "@/components/base/empty-state";
import type { DateRange } from "@/types/reports";

interface SystemStatisticsProps {
  dateRange?: DateRange;
}

interface DocTypeRow {
  name: string;
  count: number;
}

export function SystemStatistics({ dateRange }: SystemStatisticsProps) {
  const { data: stats, isLoading, error } = useSystemStats(dateRange);

  if (isLoading) {
    return (
      <div className="space-y-6">
        <div className="grid gap-3 grid-cols-2 lg:grid-cols-4">
          {Array.from({ length: 4 }).map((_, i) => (
            <Skeleton key={i} className="h-28 rounded-md" />
          ))}
        </div>
        <Skeleton className="h-72 rounded-md" />
        <Skeleton className="h-56 rounded-md" />
      </div>
    );
  }

  if (error) {
    return (
      <Alert variant="destructive">
        <AlertCircle className="h-4 w-4" />
        <AlertDescription>
          Failed to load system statistics. Please try again.
        </AlertDescription>
      </Alert>
    );
  }

  if (!stats) {
    return (
      <EmptyState
        title="No statistics available"
        description="No data found for the selected date range."
      />
    );
  }

  const docTypeRows: DocTypeRow[] = [
    { name: "Requisitions", count: stats.documentTypeBreakdown?.requisitions ?? 0 },
    { name: "Purchase Orders", count: stats.documentTypeBreakdown?.purchaseOrders ?? 0 },
    { name: "Payment Vouchers", count: stats.documentTypeBreakdown?.paymentVouchers ?? 0 },
    { name: "GRN", count: stats.documentTypeBreakdown?.grn ?? 0 },
    { name: "Budgets", count: stats.documentTypeBreakdown?.budgets ?? 0 },
  ];

  const statusRows: { label: string; value: number; status: string }[] = [
    { label: "Draft", value: stats.statusBreakdown?.draft ?? 0, status: "draft" },
    { label: "Submitted", value: stats.statusBreakdown?.submitted ?? 0, status: "submitted" },
    { label: "In Review", value: stats.statusBreakdown?.inReview ?? 0, status: "in_approval" },
    { label: "Approved", value: stats.statusBreakdown?.approved ?? 0, status: "approved" },
    { label: "Rejected", value: stats.statusBreakdown?.rejected ?? 0, status: "rejected" },
  ];

  return (
    <div className="space-y-6">
      <div className="grid gap-3 grid-cols-2 lg:grid-cols-4">
        <MetricCard
          title="Total Documents"
          value={stats.totalDocuments ?? 0}
          icon={<FileText className="h-4 w-4" />}
          accent="blue"
          secondary="All time"
        />
        <MetricCard
          title="Approval Rate"
          value={`${(stats.approvalRate ?? 0).toFixed(1)}%`}
          icon={<TrendingUp className="h-4 w-4" />}
          accent="emerald"
          secondary={`${stats.approvedDocuments ?? 0} approved`}
        />
        <MetricCard
          title="Avg Approval Time"
          value={(stats.averageApprovalTime ?? 0).toFixed(1)}
          icon={<Clock className="h-4 w-4" />}
          accent="amber"
          secondary="days"
        />
        <MetricCard
          title="Rejection Rate"
          value={`${(stats.rejectionRate ?? 0).toFixed(1)}%`}
          icon={<AlertCircle className="h-4 w-4" />}
          accent="rose"
          secondary={`${stats.rejectedDocuments ?? 0} rejected`}
        />
      </div>

      <Card className="border-border/60">
        <CardHeader>
          <CardTitle className="text-base">Document Type Distribution</CardTitle>
        </CardHeader>
        <CardContent>
          <ReportChart<DocTypeRow>
            kind="bar"
            data={docTypeRows}
            xKey="name"
            series={[{ dataKey: "count", label: "Count" }]}
            perBarColor
          />
        </CardContent>
      </Card>

      <Card className="border-border/60">
        <CardHeader>
          <CardTitle className="text-base">Status Summary</CardTitle>
        </CardHeader>
        <CardContent>
          <ul className="divide-y divide-border/60">
            {statusRows.map((row) => (
              <li
                key={row.label}
                className="flex items-center justify-between py-3"
              >
                <div className="flex items-center gap-3">
                  <StatusBadge status={row.status} type="document" />
                </div>
                <span className="text-sm font-semibold tabular-nums">
                  {row.value}
                </span>
              </li>
            ))}
          </ul>
        </CardContent>
      </Card>
    </div>
  );
}

export type { SystemStatisticsProps };
```

- [ ] **Step 2: Verify TS compiles**

Run: `cd frontend && pnpm tsc --noEmit`
Expected: clean.

- [ ] **Step 3: Commit**

```bash
git add "frontend/src/app/(private)/admin/_components/system-statistics.tsx"
git commit -m "refactor(admin): SystemStatistics uses MetricCard + ReportChart + StatusBadge"
```

---

## Task 5: Refactor `approval-reports.tsx`

**Files:**
- Modify: `frontend/src/app/(private)/admin/_components/approval-reports.tsx`

Drop 3 bespoke cards → `MetricCard`. Inline search → `FilterBar`. Raw `<Table>` → `<DataList>` with mobile card. `formatDocumentType` switch removed in favor of `<DocumentTypeChip>`.

- [ ] **Step 1: Replace the file**

```tsx
// frontend/src/app/(private)/admin/_components/approval-reports.tsx
"use client";

import { useState } from "react";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { useApprovalMetrics } from "@/hooks/use-reports-queries";
import { Input } from "@/components/ui/input";
import { Search, AlertCircle, CheckCircle2, XCircle, Clock } from "lucide-react";
import { Skeleton } from "@/components/ui/skeleton";
import { MetricCard } from "@/components/ui/metric-card";
import { FilterBar } from "@/components/ui/filter-bar";
import { DataList, DataListColumn } from "@/components/ui/data-list";
import { DocumentTypeChip } from "@/components/ui/document-type-chip";
import { StatusBadge } from "@/components/status-badge";
import EmptyState from "@/components/base/empty-state";
import type { DateRange, ApprovalActivity } from "@/types/reports";

interface ApprovalReportsProps {
  dateRange?: DateRange;
}

export function ApprovalReports({ dateRange }: ApprovalReportsProps) {
  const [searchTerm, setSearchTerm] = useState("");
  const { data: metrics, isLoading, error } = useApprovalMetrics(dateRange);

  if (isLoading) {
    return (
      <div className="space-y-6">
        <div className="grid gap-3 grid-cols-2 lg:grid-cols-3">
          {Array.from({ length: 3 }).map((_, i) => (
            <Skeleton key={i} className="h-28 rounded-md" />
          ))}
        </div>
        <Skeleton className="h-72 rounded-md" />
      </div>
    );
  }

  if (error) {
    return (
      <Alert variant="destructive">
        <AlertCircle className="h-4 w-4" />
        <AlertDescription>
          Failed to load approval reports. Please try again.
        </AlertDescription>
      </Alert>
    );
  }

  if (!metrics) {
    return (
      <EmptyState
        title="No approval data"
        description="No approvals found for the selected date range."
      />
    );
  }

  const filtered: ApprovalActivity[] = (metrics.recentApprovals ?? []).filter(
    (item) =>
      (item.documentNumber || "")
        .toLowerCase()
        .includes(searchTerm.toLowerCase()) ||
      (item.approverName || "").toLowerCase().includes(searchTerm.toLowerCase())
  );

  const columns: DataListColumn<ApprovalActivity>[] = [
    {
      id: "doc",
      header: "Document",
      cell: (a) => (
        <span className="font-medium text-primary">{a.documentNumber}</span>
      ),
    },
    {
      id: "type",
      header: "Type",
      priority: "md",
      cell: (a) => <DocumentTypeChip type={a.documentType} />,
    },
    {
      id: "status",
      header: "Status",
      cell: (a) => <StatusBadge status={a.action} type="action" />,
    },
    {
      id: "approver",
      header: "Approver",
      priority: "md",
      cell: (a) => (
        <span className="text-sm text-muted-foreground">{a.approverName}</span>
      ),
    },
    {
      id: "time",
      header: "Time",
      priority: "lg",
      cell: (a) => (
        <span className="text-sm text-muted-foreground">
          {new Date(a.createdAt).toLocaleDateString("en", {
            month: "short",
            day: "numeric",
            hour: "2-digit",
            minute: "2-digit",
          })}
        </span>
      ),
    },
  ];

  const hasActiveFilters = searchTerm.length > 0;

  return (
    <div className="space-y-6">
      <div className="grid gap-3 grid-cols-2 lg:grid-cols-3">
        <MetricCard
          title="Approved"
          value={metrics.totalApproved ?? 0}
          icon={<CheckCircle2 className="h-4 w-4" />}
          accent="emerald"
          secondary={`${(metrics.approvalRate ?? 0).toFixed(1)}% approval rate`}
        />
        <MetricCard
          title="Rejections"
          value={metrics.totalRejected ?? 0}
          icon={<XCircle className="h-4 w-4" />}
          accent="rose"
          secondary={`${(100 - (metrics.approvalRate ?? 0)).toFixed(1)}% rejection rate`}
        />
        <MetricCard
          title="Pending Review"
          value={metrics.totalPending ?? 0}
          icon={<Clock className="h-4 w-4" />}
          accent="amber"
          secondary="Awaiting next approver"
        />
      </div>

      <FilterBar
        search={
          <div className="relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
            <Input
              placeholder="Search by document or approver…"
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="pl-10"
            />
          </div>
        }
        hasActiveFilters={hasActiveFilters}
        onReset={() => setSearchTerm("")}
        meta={`${filtered.length} approval${filtered.length === 1 ? "" : "s"}${hasActiveFilters ? " (filtered)" : ""}`}
      />

      <DataList<ApprovalActivity>
        rows={filtered}
        columns={columns}
        getRowId={(a) => a.id}
        emptyMessage="No approvals found."
        mobileCard={(a) => (
          <div className="flex flex-col gap-2">
            <div className="flex items-start justify-between gap-2">
              <div className="min-w-0">
                <div className="font-medium text-primary">{a.documentNumber}</div>
                <div className="text-xs text-muted-foreground">
                  {a.approverName}
                </div>
              </div>
              <StatusBadge status={a.action} type="action" />
            </div>
            <div className="flex flex-wrap items-center gap-2 text-xs text-muted-foreground">
              <DocumentTypeChip type={a.documentType} />
              <span>
                {new Date(a.createdAt).toLocaleDateString("en", {
                  month: "short",
                  day: "numeric",
                })}
              </span>
            </div>
          </div>
        )}
      />
    </div>
  );
}

export type { ApprovalReportsProps };
```

- [ ] **Step 2: Verify TS compiles**

Run: `cd frontend && pnpm tsc --noEmit`
Expected: clean.

- [ ] **Step 3: Commit**

```bash
git add "frontend/src/app/(private)/admin/_components/approval-reports.tsx"
git commit -m "refactor(admin): ApprovalReports uses MetricCard + DataList + DocumentTypeChip"
```

---

## Task 6: Refactor `user-activity-reports.tsx`

**Files:**
- Modify: `frontend/src/app/(private)/admin/_components/user-activity-reports.tsx`

3 bespoke cards → `MetricCard`. Top contributors uses existing `<UserAvatar>` (or fallback initial circle if `<UserAvatar>` requires more user fields than available). Activity log raw `<Table>` → `<DataList>` with mobile card.

- [ ] **Step 1: Replace the file**

```tsx
// frontend/src/app/(private)/admin/_components/user-activity-reports.tsx
"use client";

import { Alert, AlertDescription } from "@/components/ui/alert";
import { useUserActivity } from "@/hooks/use-reports-queries";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { User, Users, CheckCircle2, AlertCircle } from "lucide-react";
import { MetricCard } from "@/components/ui/metric-card";
import { DataList, DataListColumn } from "@/components/ui/data-list";
import EmptyState from "@/components/base/empty-state";
import type { DateRange, UserActivity } from "@/types/reports";

interface UserActivityReportsProps {
  dateRange?: DateRange;
}

function formatDate(iso?: string) {
  if (!iso) return "N/A";
  try {
    return new Date(iso).toLocaleDateString("en", {
      month: "short",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
  } catch {
    return "N/A";
  }
}

export function UserActivityReports({ dateRange }: UserActivityReportsProps) {
  const { data: activity, isLoading, error } = useUserActivity(dateRange);

  if (isLoading) {
    return (
      <div className="space-y-6">
        <div className="grid gap-3 grid-cols-2 lg:grid-cols-3">
          {Array.from({ length: 3 }).map((_, i) => (
            <Skeleton key={i} className="h-28 rounded-md" />
          ))}
        </div>
        <Skeleton className="h-44 rounded-md" />
        <Skeleton className="h-72 rounded-md" />
      </div>
    );
  }

  if (error) {
    return (
      <Alert variant="destructive">
        <AlertCircle className="h-4 w-4" />
        <AlertDescription>
          Failed to load user activity. Please try again.
        </AlertDescription>
      </Alert>
    );
  }

  if (!activity) {
    return (
      <EmptyState
        title="No user activity"
        description="No user activity found for the selected date range."
      />
    );
  }

  const topContributors = (activity.users ?? []).slice(0, 5);
  const allUsers = activity.users ?? [];

  const columns: DataListColumn<UserActivity>[] = [
    {
      id: "name",
      header: "User",
      cell: (u) => <span className="font-medium">{u.name}</span>,
    },
    {
      id: "role",
      header: "Role",
      priority: "md",
      cell: (u) => (
        <Badge variant="outline">{u.role.replace(/_/g, " ")}</Badge>
      ),
    },
    {
      id: "approvals",
      header: <span className="text-right block">Approvals</span>,
      priority: "md",
      cell: (u) => (
        <span className="text-right block tabular-nums">{u.approvalCount}</span>
      ),
    },
    {
      id: "rejections",
      header: <span className="text-right block">Rejections</span>,
      priority: "md",
      cell: (u) => (
        <span className="text-right block tabular-nums">
          {u.rejectionCount}
        </span>
      ),
    },
    {
      id: "active",
      header: <span className="text-right block">Active</span>,
      priority: "lg",
      cell: (u) => (
        <span className="text-right block tabular-nums">
          {u.activeDocuments}
        </span>
      ),
    },
    {
      id: "last",
      header: "Last activity",
      priority: "lg",
      cell: (u) => (
        <span className="text-sm text-muted-foreground">
          {formatDate(u.lastActivity)}
        </span>
      ),
    },
  ];

  return (
    <div className="space-y-6">
      <div className="grid gap-3 grid-cols-2 lg:grid-cols-3">
        <MetricCard
          title="Active Users"
          value={activity.activeUsers ?? 0}
          icon={<Users className="h-4 w-4" />}
          accent="blue"
          secondary={`${allUsers.length} total users`}
        />
        <MetricCard
          title="Docs in Progress"
          value={activity.documentsInProgress ?? 0}
          icon={<User className="h-4 w-4" />}
          accent="violet"
          secondary="Across all users"
        />
        <MetricCard
          title="Total Actions"
          value={activity.totalActions ?? 0}
          icon={<CheckCircle2 className="h-4 w-4" />}
          accent="emerald"
          secondary="Approvals and rejections"
        />
      </div>

      <Card className="border-border/60">
        <CardHeader>
          <CardTitle className="text-base">Top Contributors</CardTitle>
        </CardHeader>
        <CardContent>
          {topContributors.length === 0 ? (
            <EmptyState
              title="No contributors yet"
              description="Approvals will appear here once users start acting on tasks."
            />
          ) : (
            <ul className="space-y-2">
              {topContributors.map((u, idx) => (
                <li
                  key={u.id}
                  className="flex items-center justify-between p-3 rounded-md border border-border/60"
                >
                  <div className="flex items-center gap-3 min-w-0">
                    <div className="h-9 w-9 rounded-full bg-primary/10 flex items-center justify-center font-semibold text-primary text-sm shrink-0">
                      {(u.name || "?").charAt(0).toUpperCase()}
                    </div>
                    <div className="min-w-0">
                      <p className="font-medium leading-tight truncate">
                        {u.name}{" "}
                        {idx === 0 && (
                          <span className="ml-1 text-[10px] uppercase text-amber-600 font-bold">
                            Top
                          </span>
                        )}
                      </p>
                      <p className="text-xs text-muted-foreground capitalize">
                        {u.role.replace(/_/g, " ")}
                      </p>
                    </div>
                  </div>
                  <div className="text-right shrink-0">
                    <Badge variant="secondary">
                      {u.approvalCount} approvals
                    </Badge>
                    <p className="text-xs text-muted-foreground mt-1">
                      {u.activeDocuments} active
                    </p>
                  </div>
                </li>
              ))}
            </ul>
          )}
        </CardContent>
      </Card>

      <Card className="border-border/60">
        <CardHeader>
          <CardTitle className="text-base">User Activity Log</CardTitle>
        </CardHeader>
        <CardContent>
          <DataList<UserActivity>
            rows={allUsers}
            columns={columns}
            getRowId={(u) => u.id}
            emptyMessage="No user activity found."
            mobileCard={(u) => (
              <div className="flex flex-col gap-2">
                <div className="flex items-start justify-between gap-2">
                  <div className="min-w-0">
                    <div className="font-medium">{u.name}</div>
                    <div className="text-xs text-muted-foreground capitalize">
                      {u.role.replace(/_/g, " ")}
                    </div>
                  </div>
                  <Badge variant="secondary">{u.approvalCount} ✓</Badge>
                </div>
                <div className="flex flex-wrap items-center gap-2 text-xs text-muted-foreground">
                  <span>{u.rejectionCount} rejected</span>
                  <span>·</span>
                  <span>{u.activeDocuments} active</span>
                  <span>·</span>
                  <span>{formatDate(u.lastActivity)}</span>
                </div>
              </div>
            )}
          />
        </CardContent>
      </Card>
    </div>
  );
}

export type { UserActivityReportsProps };
```

- [ ] **Step 2: Verify TS compiles**

Run: `cd frontend && pnpm tsc --noEmit`
Expected: clean.

- [ ] **Step 3: Commit**

```bash
git add "frontend/src/app/(private)/admin/_components/user-activity-reports.tsx"
git commit -m "refactor(admin): UserActivityReports uses MetricCard + DataList"
```

---

## Task 7: Refactor `admin-reports-client.tsx`

**Files:**
- Modify: `frontend/src/app/(private)/admin/_components/admin-reports-client.tsx`

Mount `<ReportsHeader>`, parse current `dateRange` from URL once via `useDateRangeUrlState`, thread it down to children, switch tabs to Plan A pill style.

- [ ] **Step 1: Replace the file**

```tsx
// frontend/src/app/(private)/admin/_components/admin-reports-client.tsx
"use client";

import { useState, useMemo } from "react";
import { format, subDays, startOfDay, endOfDay } from "date-fns";
import { useQueryClient } from "@tanstack/react-query";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { ApprovalReports } from "./approval-reports";
import { UserActivityReports } from "./user-activity-reports";
import { SystemStatistics } from "./system-statistics";
import { AnalyticsDashboard } from "@/components/workflows/analytics-dashboard";
import { ReportsHeader } from "./reports-header";
import { QUERY_KEYS } from "@/lib/constants";
import { notify } from "@/lib/utils";
import { useDateRangeUrlState } from "@/hooks/use-date-range-url-state";
import {
  useSystemStats,
  useApprovalMetrics,
  useUserActivity,
  useAnalyticsDashboard,
} from "@/hooks/use-reports-queries";
import {
  exportSystemStatsToCSV,
  exportApprovalMetricsToCSV,
  exportUserActivityToCSV,
  exportAnalyticsDashboardToCSV,
} from "@/lib/export-utils";

interface AdminReportsClientProps {
  userId: string;
  userRole: string;
}

function defaultRange() {
  const today = new Date();
  return {
    from: format(startOfDay(subDays(today, 27)), "yyyy-MM-dd"),
    to: format(endOfDay(today), "yyyy-MM-dd"),
  };
}

export function AdminReportsClient({
  userId: _userId,
  userRole: _userRole,
}: AdminReportsClientProps) {
  const [activeTab, setActiveTab] = useState("overview");
  const [isRefreshing, setIsRefreshing] = useState(false);
  const queryClient = useQueryClient();

  const initial = useMemo(defaultRange, []);
  const { from, to } = useDateRangeUrlState({
    defaultFrom: initial.from,
    defaultTo: initial.to,
  });
  const dateRange = useMemo(() => ({ from, to }), [from, to]);

  const { data: systemStats } = useSystemStats(dateRange);
  const { data: approvalMetrics } = useApprovalMetrics(dateRange);
  const { data: userActivity } = useUserActivity(dateRange);
  const { data: analytics } = useAnalyticsDashboard(dateRange);

  const handleRefresh = async () => {
    setIsRefreshing(true);
    try {
      await Promise.all([
        queryClient.invalidateQueries({
          queryKey: [QUERY_KEYS.REPORTS.SYSTEM_STATS],
        }),
        queryClient.invalidateQueries({
          queryKey: [QUERY_KEYS.REPORTS.APPROVAL_METRICS],
        }),
        queryClient.invalidateQueries({
          queryKey: [QUERY_KEYS.REPORTS.USER_ACTIVITY],
        }),
        queryClient.invalidateQueries({
          queryKey: [QUERY_KEYS.REPORTS.ANALYTICS],
        }),
      ]);
      notify({
        title: "Success",
        description: "Reports refreshed successfully",
        type: "success",
      });
    } catch {
      notify({
        title: "Error",
        description: "Failed to refresh reports. Please try again.",
        type: "error",
      });
    } finally {
      setIsRefreshing(false);
    }
  };

  const handleExport = (formatChoice: "csv") => {
    if (formatChoice !== "csv") return;
    try {
      switch (activeTab) {
        case "overview":
          if (!systemStats)
            return notify({ title: "Error", description: "No data to export", type: "error" });
          exportSystemStatsToCSV(systemStats);
          break;
        case "analytics":
          if (!analytics)
            return notify({ title: "Error", description: "No data to export", type: "error" });
          exportAnalyticsDashboardToCSV(analytics);
          break;
        case "approvals":
          if (!approvalMetrics)
            return notify({ title: "Error", description: "No data to export", type: "error" });
          exportApprovalMetricsToCSV(approvalMetrics);
          break;
        case "activity":
          if (!userActivity)
            return notify({ title: "Error", description: "No data to export", type: "error" });
          exportUserActivityToCSV(userActivity);
          break;
        default:
          notify({ title: "Error", description: "Unknown tab selected", type: "error" });
          return;
      }
      notify({
        title: "Exported",
        description: "Current view downloaded as CSV",
        type: "success",
      });
    } catch {
      notify({
        title: "Error",
        description: "An error occurred during export",
        type: "error",
      });
    }
  };

  return (
    <div className="space-y-5">
      <ReportsHeader
        title="Admin Reports"
        subtitle="Workflow approvals, user activity, system metrics"
        onRefresh={handleRefresh}
        onExport={handleExport}
        isRefreshing={isRefreshing}
      />

      <Tabs value={activeTab} onValueChange={setActiveTab} className="space-y-5">
        <TabsList className="inline-flex h-9 w-full sm:w-auto bg-muted/60 p-1 rounded-lg">
          {[
            { v: "overview", label: "Overview" },
            { v: "analytics", label: "Analytics" },
            { v: "approvals", label: "Approvals" },
            { v: "activity", label: "Activity" },
          ].map((t) => (
            <TabsTrigger
              key={t.v}
              value={t.v}
              className="flex-1 sm:flex-initial sm:px-6 data-[state=active]:bg-background data-[state=active]:text-foreground data-[state=active]:shadow-sm rounded-md"
            >
              {t.label}
            </TabsTrigger>
          ))}
        </TabsList>

        <TabsContent value="overview" className="mt-0">
          <SystemStatistics dateRange={dateRange} />
        </TabsContent>

        <TabsContent value="analytics" className="mt-0">
          <AnalyticsDashboard dateRange={dateRange} />
        </TabsContent>

        <TabsContent value="approvals" className="mt-0">
          <ApprovalReports dateRange={dateRange} />
        </TabsContent>

        <TabsContent value="activity" className="mt-0">
          <UserActivityReports dateRange={dateRange} />
        </TabsContent>
      </Tabs>
    </div>
  );
}
```

- [ ] **Step 2: Verify TS compiles**

Run: `cd frontend && pnpm tsc --noEmit`
Expected: clean. If `AnalyticsDashboard` does not yet accept `dateRange` prop, this will surface here — fix in Task 8.

- [ ] **Step 3: Commit**

```bash
git add "frontend/src/app/(private)/admin/_components/admin-reports-client.tsx"
git commit -m "refactor(admin): AdminReportsClient adopts ReportsHeader + URL date range"
```

---

## Task 8: Update `AnalyticsDashboard` to accept `dateRange` and use `<ReportChart>`

**Files:**
- Modify: `frontend/src/components/workflows/analytics-dashboard.tsx`

The existing component is 414 LOC and renders multiple raw recharts charts. Plan C does NOT do a full visual redesign of analytics — it only:
1. Adds a `dateRange?: DateRange` prop and passes it to `useAnalyticsDashboard`.
2. Replaces any direct `BarChart`/`LineChart`/`AreaChart` recharts imports + ResponsiveContainer wrappers with `<ReportChart>` calls so the brand palette + tooltip apply consistently.

Because this file is large and not previously read line-by-line, the implementer should follow this surgical procedure.

- [ ] **Step 1: Read the existing file**

Read `frontend/src/components/workflows/analytics-dashboard.tsx` end-to-end to understand what's there. Confirm: file imports recharts directly, calls `useAnalyticsDashboard()` with no args, renders one or more `<ResponsiveContainer><BarChart|LineChart|AreaChart>...` blocks.

- [ ] **Step 2: Add `dateRange` prop**

Find the existing component signature, e.g.:
```tsx
export function AnalyticsDashboard() {
```
Replace with:
```tsx
import type { DateRange } from "@/types/reports";

interface AnalyticsDashboardProps {
  dateRange?: DateRange;
}

export function AnalyticsDashboard({ dateRange }: AnalyticsDashboardProps = {}) {
```

(Add the type import at the top with the other imports if not already present.)

Find the `useAnalyticsDashboard()` call. Change to `useAnalyticsDashboard(dateRange)`.

- [ ] **Step 3: Swap each chart instance to `<ReportChart>`**

For EACH recharts chart in the file (typically there will be 1–4), replace the entire `<ResponsiveContainer>...<XxxChart>...</XxxChart></ResponsiveContainer>` block with the equivalent `<ReportChart>` invocation.

Example mapping for a bar chart:
```tsx
// Before:
<ResponsiveContainer width="100%" height={300}>
  <BarChart data={someData}>
    <CartesianGrid ... />
    <XAxis dataKey="name" />
    <YAxis />
    <Tooltip />
    <Bar dataKey="count" fill="..." />
  </BarChart>
</ResponsiveContainer>

// After:
<ReportChart
  kind="bar"
  data={someData}
  xKey="name"
  series={[{ dataKey: "count", label: "Count" }]}
/>
```

For a line chart with multiple series:
```tsx
<ReportChart
  kind="line"
  data={trendData}
  xKey="date"
  series={[
    { dataKey: "approved", label: "Approved" },
    { dataKey: "rejected", label: "Rejected" },
  ]}
  showLegend
/>
```

For an area chart:
```tsx
<ReportChart
  kind="area"
  data={cycleData}
  xKey="day"
  series={[{ dataKey: "avgHours", label: "Avg cycle (h)" }]}
/>
```

If any chart is a pie/scatter/composed type that `ReportChart` doesn't support (`bar` | `line` | `area` only), leave it as-is and note it as a follow-up. Do NOT extend `ReportChart` in this task.

After swap, remove now-unused imports from `recharts` (and from `@/components/ui/chart` if any direct usage remained).

- [ ] **Step 4: Verify TS compiles + tests**

Run: `cd frontend && pnpm tsc --noEmit`
Expected: clean.

Run: `cd frontend && pnpm vitest run src/__tests__/components/ui/`
Expected: all primitive tests still pass.

- [ ] **Step 5: Commit**

```bash
git add "frontend/src/components/workflows/analytics-dashboard.tsx"
git commit -m "refactor(workflows): AnalyticsDashboard accepts dateRange + uses ReportChart"
```

If charts that `ReportChart` doesn't support remain, append a one-line note to the commit message body listing them, e.g.:
```
- pie-chart of doc statuses kept on raw recharts (ReportChart kinds = bar|line|area)
```

---

## Task 9: Drop dead `loading.tsx`

**Files:**
- Delete: `frontend/src/app/(private)/admin/reports/loading.tsx`

The 105-line `loading.tsx` is never reached: each sub-component renders its own loading state (now `Skeleton`s after Tasks 4-6), so the route-level `loading.tsx` is unreachable in practice. Remove it.

- [ ] **Step 1: Confirm dead status**

Run: `cd frontend && grep -r "admin/reports/loading" src/ --include='*.tsx' --include='*.ts'`
Expected: no matches (the file is referenced only by Next.js routing convention, which only invokes it during the server render before the client tree mounts — and now `AdminReportsClient` mounts immediately with hooks that have their own loading states).

To be safe, briefly start the dev server and navigate to `/admin/reports`. Confirm no flash of the bespoke loading skeleton (and that the new `<Skeleton>` blocks inside sub-components show during data fetch).

If you cannot run the dev server in CI: manually inspect `loading.tsx` and confirm its contents do not reference any logic that is required (it should be a pure visual skeleton).

- [ ] **Step 2: Delete the file**

```bash
rm "frontend/src/app/(private)/admin/reports/loading.tsx"
```

- [ ] **Step 3: Verify build**

Run: `cd frontend && pnpm tsc --noEmit`
Expected: clean.

- [ ] **Step 4: Commit**

```bash
git add -A "frontend/src/app/(private)/admin/reports/"
git commit -m "chore(admin): drop dead reports/loading.tsx (sub-components own loading state)"
```

If you found that `loading.tsx` actually IS reached during real navigation (e.g., it shows briefly on first paint), do NOT delete — instead replace its contents with a skeleton matching the new `MetricCard` + chart layout, and commit as `refactor(admin): align loading.tsx with new reports layout`.

---

## Task 10: Integration smoke test

**Files:**
- Create: `frontend/src/__tests__/integration/admin-reports.test.tsx`

Confirms tab switching + date-range URL sync still work end-to-end with mocked hooks. Does not assert visual layout — pure orchestration check.

- [ ] **Step 1: Write test**

```tsx
// frontend/src/__tests__/integration/admin-reports.test.tsx
import { render, screen, fireEvent } from "@testing-library/react";
import { describe, it, expect, vi, beforeEach } from "vitest";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { AdminReportsClient } from "@/app/(private)/admin/_components/admin-reports-client";

const replace = vi.fn();
const params = new URLSearchParams();

vi.mock("next/navigation", () => ({
  useRouter: () => ({ replace }),
  usePathname: () => "/admin/reports",
  useSearchParams: () => params,
}));

// Stub the four data hooks so the tree renders without a network call.
vi.mock("@/hooks/use-reports-queries", () => ({
  useSystemStats: () => ({ data: undefined, isLoading: true, error: null }),
  useApprovalMetrics: () => ({ data: undefined, isLoading: true, error: null }),
  useUserActivity: () => ({ data: undefined, isLoading: true, error: null }),
  useAnalyticsDashboard: () => ({ data: undefined, isLoading: true, error: null }),
}));

// Stub the analytics dashboard so it doesn't pull deeper deps in the smoke test.
vi.mock("@/components/workflows/analytics-dashboard", () => ({
  AnalyticsDashboard: ({ dateRange }: { dateRange?: { from: string; to: string } }) => (
    <div data-testid="analytics-dashboard">
      analytics:{dateRange?.from}–{dateRange?.to}
    </div>
  ),
}));

// Stub notify (toast) since jsdom won't render Sonner cleanly here.
vi.mock("@/lib/utils", async () => {
  const actual = await vi.importActual<typeof import("@/lib/utils")>("@/lib/utils");
  return { ...actual, notify: vi.fn() };
});

beforeEach(() => {
  replace.mockClear();
  params.delete("from");
  params.delete("to");
});

function renderWithClient(ui: React.ReactNode) {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return render(<QueryClientProvider client={client}>{ui}</QueryClientProvider>);
}

describe("AdminReportsClient (integration)", () => {
  it("renders all 4 tabs with the correct titles", () => {
    renderWithClient(<AdminReportsClient userId="u1" userRole="admin" />);
    expect(screen.getByRole("tab", { name: /overview/i })).toBeInTheDocument();
    expect(screen.getByRole("tab", { name: /analytics/i })).toBeInTheDocument();
    expect(screen.getByRole("tab", { name: /approvals/i })).toBeInTheDocument();
    expect(screen.getByRole("tab", { name: /activity/i })).toBeInTheDocument();
  });

  it("switches active tab to Analytics on click", () => {
    renderWithClient(<AdminReportsClient userId="u1" userRole="admin" />);
    fireEvent.click(screen.getByRole("tab", { name: /analytics/i }));
    expect(screen.getByTestId("analytics-dashboard")).toBeInTheDocument();
  });

  it("renders the page title and subtitle", () => {
    renderWithClient(<AdminReportsClient userId="u1" userRole="admin" />);
    expect(screen.getByText("Admin Reports")).toBeInTheDocument();
    expect(
      screen.getByText("Workflow approvals, user activity, system metrics")
    ).toBeInTheDocument();
  });
});
```

- [ ] **Step 2: Run test**

Run: `cd frontend && pnpm vitest run src/__tests__/integration/admin-reports.test.tsx`
Expected: 3 pass.

If `AdminReportsClient` import path resolution fails because of the parenthesized route segment in the path, the test config likely already supports it (other tests reference `(private)/(main)/...` paths via the `@/` alias). If not, escape the path or use a relative import.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/__tests__/integration/admin-reports.test.tsx
git commit -m "test(admin): integration smoke for AdminReportsClient tabs + header"
```

---

## Task 11: Verification pass

- [ ] **Step 1: Full type-check**

Run: `cd frontend && pnpm tsc --noEmit`
Expected: clean.

- [ ] **Step 2: Plan C tests**

Run:
```bash
cd frontend && pnpm vitest run \
  src/__tests__/hooks/use-date-range-url-state.test.ts \
  src/__tests__/components/admin/reports-header.test.tsx \
  src/__tests__/integration/admin-reports.test.tsx
```
Expected: 5 + 4 + 3 = 12 pass.

- [ ] **Step 3: Plan A + B regression**

Run:
```bash
cd frontend && pnpm vitest run \
  src/__tests__/components/ui/stat-grid.test.tsx \
  src/__tests__/components/ui/data-list.test.tsx \
  src/__tests__/components/ui/priority-badge.test.tsx \
  src/__tests__/components/ui/trend-delta.test.tsx \
  src/__tests__/components/ui/metric-card.test.tsx \
  src/__tests__/components/ui/document-type-chip.test.tsx \
  src/__tests__/components/workflows/approval-chain-stepper.test.tsx \
  src/__tests__/components/layout/detail-shell.test.tsx \
  src/__tests__/components/layout/mobile-bottom-nav.test.tsx \
  src/__tests__/components/base/empty-state.test.tsx
```
Expected: all pass (Plan A 8 + Plan B 27 - hook tests already counted = 33 unique ⇒ tally with the actual run).

- [ ] **Step 4: Manual smoke (optional)**

If able, run `cd frontend && pnpm dev` and visit:
- `/admin/reports` → header with date range, 4 tabs
- `/admin/reports?tab=approvals` → approvals tab still loads (NOTE: tab state is local to component, not URL — see Plan I follow-up)
- `/admin/reports?from=2026-01-01&to=2026-01-31` → URL state propagates, hooks refetch

Verify:
- Mobile (390px): MetricCard grid 2-col, FilterBar wraps, DataList shows mobile cards
- Desktop ≥1024px: MetricCard 3- or 4-col, DataList shows full table

- [ ] **Step 5: Cleanup commit if any**

```bash
git add -A
git commit -m "chore(admin): plan-C verification cleanup"
```

---

## Self-Review Notes

- **Spec coverage:** Each section of the original spec maps to a task — `MetricCard` (4-6), `DataList` (5,6), `PageHeader` (3), Export `DropdownMenu` (3), global date range (3, used by 4-7), drop dead `loading.tsx` (9), `setRange` no-op guard (1), `ReportChart` generics tightening (2), consistency rules (4-7 use shared primitives end-to-end).
- **Type consistency:** `dateRange?: DateRange` prop name standardized across `SystemStatistics`, `ApprovalReports`, `UserActivityReports`, `AnalyticsDashboard`, all 4 hooks. `ReportsHeader.onExport(format: "csv")` matches `AdminReportsClient.handleExport(formatChoice: "csv")` — note the literal type union is `"csv"` only; future PDF support extends both signatures together.
- **No placeholders:** Each task includes complete code; no "TBD" or "similar to Task N". Task 8 (AnalyticsDashboard) requires reading the existing file because it's 414 LOC and varies — surgical procedure documented step-by-step instead of pasted code.
- **Test placements:** All under `frontend/src/__tests__/`. Integration test path resolves via `@/` alias.
- **Out-of-scope discipline:** No procurement files touched. No detail pages. No modal sweeps.

## Plan D carry-forward notes

- Tab state is local to `AdminReportsClient` — not synced to URL. Could reuse `?tab=` pattern from Plan A `tasks-client.tsx`. Defer to Plan I cleanup.
- `ReportChart` doesn't support pie/donut. If `AnalyticsDashboard` has any, they remain raw recharts. Add a `kind: "pie"` extension to `ReportChart` in a follow-up if multiple consumers need it.
- Top Contributors avatar still uses bespoke initials circle. `<UserAvatar>` requires more user fields than `UserActivity` provides; promotion to a shared `<InitialsAvatar>` primitive is reasonable in Plan I.
- Currency / count / percentage formatting helpers for `ReportChart` tooltips (Plan B carry-forward #3) deferred — not blocking Plan C since the default tooltip suffices.
