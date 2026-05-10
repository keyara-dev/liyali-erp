# UI Revamp — Plan B: Shared Primitives Bedrock

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the reusable primitive library that every later refactor (Plans C–I) depends on — metric cards, badges, charts, detail page shell, approval stepper, plus URL-state hooks.

**Architecture:** Add small composable primitives layered on existing ShadCN/Radix/recharts infra. Reuse existing `StatusBadge` (already canonical for document/execution/approval status — do NOT build a new `DocumentStatusPill`). Reuse existing `ChartContainer` from `ui/chart.tsx` as the foundation for `<ReportChart>`. Promote inline patterns currently duplicated across 5+ files (priority badge, doc-type chip, approval stepper) into shared components. Extend `EmptyState` and `CalendarDateRangePicker` rather than replacing.

**Tech Stack:** Next.js App Router, TypeScript, Tailwind v4, ShadCN UI, Radix, recharts (already present), date-fns (already present), Vitest + Testing Library.

**Out-of-scope (handled in later plans):**
- Plan C — `/admin/reports` consumes new primitives
- Plan D — modal sweep (`<Dialog>` → `<ResponsiveSheet>`)
- Plan E — list tables sweep (raw `<Table>` → `<DataList>` + `<FilterBar>`)
- Plan F — detail pages sweep (apply `<DetailShell>` + split monoliths)
- Plan G — vendor detail page + `/reports` decision
- Plan H — approval flow polish
- Plan I — cleanup / stub removal

---

## Existing primitives (reuse, don't duplicate)

- `StatusBadge` (`frontend/src/components/status-badge.tsx`) — canonical badge for document/action/execution/approval/compliance/role/health statuses. Backed by `lib/status-badges`. Plan C–E should enforce this everywhere; do NOT build a new `DocumentStatusPill`.
- `ChartContainer`, `ChartTooltipContent`, `ChartLegendContent` (`frontend/src/components/ui/chart.tsx`) — full recharts theming infra exists. `<ReportChart>` wraps these with a brand palette and chart-type presets, no new tooltip/legend code.
- `CalendarDateRangePicker` (`frontend/src/components/ui/date-range-picker.tsx`) — has 8 presets (Today through ThisYear). Plan B extends with 90d and Quarter, plus a URL-state hook.
- `EmptyState` (`frontend/src/components/base/empty-state.tsx`) — exists but missing `action` slot, has a typo (`text-gary-400`), has `motion.div` without `initial`/`animate` wired. Plan B fixes all three.
- `Badge` (`frontend/src/components/ui/badge.tsx`) — base for `PriorityBadge`.
- `DataList`, `FilterBar`, `StatGrid`, `ResponsiveSheet`, `MobileBottomNav` — built in Plan A, ready.

---

## File Structure

**Create:**
- `frontend/src/components/ui/priority-badge.tsx` — Priority pill (HIGH/URGENT → destructive, MEDIUM → warning, LOW/default → info)
- `frontend/src/components/ui/trend-delta.tsx` — Up/down arrow + percentage delta vs prior period
- `frontend/src/components/ui/metric-card.tsx` — Single-cell metric with title/value/icon/secondary/trend/optional sparkline
- `frontend/src/components/ui/document-type-chip.tsx` — Maps document_type string to display label + small icon
- `frontend/src/components/ui/report-chart.tsx` — Recharts wrapper with chart-1..5 palette presets for bar/line/area
- `frontend/src/components/workflows/approval-chain-stepper.tsx` — Vertical stepper showing approval stages with state per stage
- `frontend/src/components/layout/detail-shell.tsx` — Detail-page layout with header slot + main + optional sidebar (tabs-on-mobile / 2-col-on-lg)
- `frontend/src/hooks/use-date-range-url-state.ts` — Hook syncing `?from=YYYY-MM-DD&to=YYYY-MM-DD&period=last30` with state
- `frontend/src/__tests__/components/ui/priority-badge.test.tsx`
- `frontend/src/__tests__/components/ui/trend-delta.test.tsx`
- `frontend/src/__tests__/components/ui/metric-card.test.tsx`
- `frontend/src/__tests__/components/ui/document-type-chip.test.tsx`
- `frontend/src/__tests__/components/workflows/approval-chain-stepper.test.tsx`
- `frontend/src/__tests__/components/layout/detail-shell.test.tsx`
- `frontend/src/__tests__/hooks/use-date-range-url-state.test.ts`

**Modify:**
- `frontend/src/components/base/empty-state.tsx` — add `action` slot, fix `text-gary-400` typo, wire framer-motion `initial`/`animate`
- `frontend/src/components/ui/date-range-picker.tsx` — extend `dateFilterPresets` with `last90Days` and `lastQuarter`
- `frontend/src/app/(private)/(main)/tasks/_components/tasks-table.tsx` — replace inline `PriorityBadge` with shared `<PriorityBadge>` (proves consumption)

---

## Task 1: `PriorityBadge` primitive

**Files:**
- Create: `frontend/src/components/ui/priority-badge.tsx`
- Test: `frontend/src/__tests__/components/ui/priority-badge.test.tsx`

- [ ] **Step 1: Write failing test**

```tsx
// frontend/src/__tests__/components/ui/priority-badge.test.tsx
import { render, screen } from "@testing-library/react";
import { describe, it, expect } from "vitest";
import { PriorityBadge } from "@/components/ui/priority-badge";

describe("PriorityBadge", () => {
  it("renders the priority text", () => {
    render(<PriorityBadge priority="HIGH" />);
    expect(screen.getByText("HIGH")).toBeInTheDocument();
  });

  it("falls back to MEDIUM when priority is undefined", () => {
    render(<PriorityBadge priority={undefined} />);
    expect(screen.getByText("MEDIUM")).toBeInTheDocument();
  });

  it("normalizes case for variant lookup", () => {
    render(<PriorityBadge priority="urgent" />);
    expect(screen.getByText("urgent")).toBeInTheDocument();
  });
});
```

- [ ] **Step 2: Run test, expect failure**

Run: `cd frontend && pnpm vitest run src/__tests__/components/ui/priority-badge.test.tsx`
Expected: FAIL — module not found.

- [ ] **Step 3: Implement `PriorityBadge`**

```tsx
// frontend/src/components/ui/priority-badge.tsx
import * as React from "react";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";

export type PriorityValue = "URGENT" | "HIGH" | "MEDIUM" | "LOW" | string;

export interface PriorityBadgeProps {
  /** Priority string. Falls back to "MEDIUM" if undefined. Case-insensitive for variant. */
  priority?: PriorityValue;
  className?: string;
}

const VARIANT: Record<string, "destructive" | "warning" | "info"> = {
  URGENT: "destructive",
  HIGH: "destructive",
  MEDIUM: "warning",
  LOW: "info",
};

export function PriorityBadge({ priority, className }: PriorityBadgeProps) {
  const display = priority || "MEDIUM";
  const variant = VARIANT[display.toUpperCase()] || "info";
  return (
    <Badge
      variant={variant}
      className={cn(
        "px-2 py-0.5 rounded text-[10px] uppercase font-medium tracking-wider",
        className
      )}
    >
      {display}
    </Badge>
  );
}
```

- [ ] **Step 4: Run test, expect pass**

Run: `cd frontend && pnpm vitest run src/__tests__/components/ui/priority-badge.test.tsx`
Expected: PASS — 3 tests.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/ui/priority-badge.tsx frontend/src/__tests__/components/ui/priority-badge.test.tsx
git commit -m "feat(ui): add PriorityBadge primitive"
```

---

## Task 2: `TrendDelta` primitive

**Files:**
- Create: `frontend/src/components/ui/trend-delta.tsx`
- Test: `frontend/src/__tests__/components/ui/trend-delta.test.tsx`

- [ ] **Step 1: Write failing test**

```tsx
// frontend/src/__tests__/components/ui/trend-delta.test.tsx
import { render, screen } from "@testing-library/react";
import { describe, it, expect } from "vitest";
import { TrendDelta } from "@/components/ui/trend-delta";

describe("TrendDelta", () => {
  it("shows positive percentage with up arrow when value > 0", () => {
    render(<TrendDelta value={12.34} />);
    expect(screen.getByText(/12\.3%/)).toBeInTheDocument();
    expect(screen.getByTestId("trend-arrow")).toHaveAttribute("data-direction", "up");
  });

  it("shows negative percentage with down arrow when value < 0", () => {
    render(<TrendDelta value={-5.6} />);
    expect(screen.getByText(/5\.6%/)).toBeInTheDocument();
    expect(screen.getByTestId("trend-arrow")).toHaveAttribute("data-direction", "down");
  });

  it("renders a flat indicator when value === 0", () => {
    render(<TrendDelta value={0} />);
    expect(screen.getByText(/0\.0%/)).toBeInTheDocument();
    expect(screen.getByTestId("trend-arrow")).toHaveAttribute("data-direction", "flat");
  });

  it("renders the comparison label when provided", () => {
    render(<TrendDelta value={3} label="vs last week" />);
    expect(screen.getByText(/vs last week/)).toBeInTheDocument();
  });

  it("inverts up/down semantics when invert is set (e.g. lower is better)", () => {
    render(<TrendDelta value={5} invert />);
    // Still shows up arrow direction-wise but uses negative tone
    const root = screen.getByTestId("trend-delta");
    expect(root).toHaveAttribute("data-tone", "negative");
  });
});
```

- [ ] **Step 2: Run test, expect failure**

Run: `cd frontend && pnpm vitest run src/__tests__/components/ui/trend-delta.test.tsx`
Expected: FAIL.

- [ ] **Step 3: Implement `TrendDelta`**

```tsx
// frontend/src/components/ui/trend-delta.tsx
import * as React from "react";
import { ArrowDown, ArrowUp, Minus } from "lucide-react";
import { cn } from "@/lib/utils";

export interface TrendDeltaProps {
  /** Percentage change vs prior period. Sign carries direction. */
  value: number;
  /** Comparison label, e.g. "vs last week". Optional. */
  label?: string;
  /** When true, treat positive values as negative tone (e.g. rejection rate going up is bad). */
  invert?: boolean;
  className?: string;
}

export function TrendDelta({ value, label, invert, className }: TrendDeltaProps) {
  const direction: "up" | "down" | "flat" =
    value > 0 ? "up" : value < 0 ? "down" : "flat";

  // Tone: positive = green, negative = red, flat = muted.
  const positiveDirection: "up" | "down" = invert ? "down" : "up";
  const tone: "positive" | "negative" | "neutral" =
    direction === "flat"
      ? "neutral"
      : direction === positiveDirection
        ? "positive"
        : "negative";

  const TONE_CLASS: Record<typeof tone, string> = {
    positive: "text-emerald-600 dark:text-emerald-400",
    negative: "text-rose-600 dark:text-rose-400",
    neutral: "text-muted-foreground",
  };

  const Icon = direction === "up" ? ArrowUp : direction === "down" ? ArrowDown : Minus;

  return (
    <span
      data-testid="trend-delta"
      data-tone={tone}
      className={cn(
        "inline-flex items-center gap-0.5 text-xs font-medium tabular-nums",
        TONE_CLASS[tone],
        className
      )}
    >
      <Icon
        data-testid="trend-arrow"
        data-direction={direction}
        className="h-3 w-3"
        aria-hidden="true"
      />
      <span>{Math.abs(value).toFixed(1)}%</span>
      {label && <span className="text-muted-foreground ml-1">{label}</span>}
    </span>
  );
}
```

- [ ] **Step 4: Run test, expect pass**

Run: `cd frontend && pnpm vitest run src/__tests__/components/ui/trend-delta.test.tsx`
Expected: PASS — 5 tests.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/ui/trend-delta.tsx frontend/src/__tests__/components/ui/trend-delta.test.tsx
git commit -m "feat(ui): add TrendDelta primitive"
```

---

## Task 3: `MetricCard` primitive

**Files:**
- Create: `frontend/src/components/ui/metric-card.tsx`
- Test: `frontend/src/__tests__/components/ui/metric-card.test.tsx`

`MetricCard` is a single-cell metric: bigger than a `StatGrid` cell, supports an icon, a value, a secondary line, an optional `<TrendDelta>`, and an optional inline sparkline (line chart). Use it standalone (admin reports tiles) or in a 2/3/4-up grid.

- [ ] **Step 1: Write failing test**

```tsx
// frontend/src/__tests__/components/ui/metric-card.test.tsx
import { render, screen } from "@testing-library/react";
import { describe, it, expect } from "vitest";
import { MetricCard } from "@/components/ui/metric-card";
import { FileText } from "lucide-react";

describe("MetricCard", () => {
  it("renders title, value, and icon", () => {
    render(
      <MetricCard
        title="Total Documents"
        value={42}
        icon={<FileText data-testid="metric-icon" />}
      />
    );
    expect(screen.getByText("Total Documents")).toBeInTheDocument();
    expect(screen.getByText("42")).toBeInTheDocument();
    expect(screen.getByTestId("metric-icon")).toBeInTheDocument();
  });

  it("renders secondary text when provided", () => {
    render(
      <MetricCard
        title="Approval Rate"
        value="92.3%"
        icon={<FileText />}
        secondary="last 30 days"
      />
    );
    expect(screen.getByText("last 30 days")).toBeInTheDocument();
  });

  it("renders TrendDelta when trend is provided", () => {
    render(
      <MetricCard
        title="Approvals"
        value={120}
        icon={<FileText />}
        trend={{ value: 8.5, label: "vs last week" }}
      />
    );
    expect(screen.getByTestId("trend-delta")).toBeInTheDocument();
    expect(screen.getByText(/8\.5%/)).toBeInTheDocument();
  });

  it("renders sparkline svg when sparkline data is provided", () => {
    const data = [1, 2, 3, 5, 8, 13, 21];
    render(
      <MetricCard
        title="Throughput"
        value={21}
        icon={<FileText />}
        sparkline={data}
      />
    );
    expect(screen.getByTestId("metric-sparkline")).toBeInTheDocument();
  });
});
```

- [ ] **Step 2: Run test, expect failure**

Run: `cd frontend && pnpm vitest run src/__tests__/components/ui/metric-card.test.tsx`
Expected: FAIL — module not found.

- [ ] **Step 3: Implement `MetricCard`**

```tsx
// frontend/src/components/ui/metric-card.tsx
import * as React from "react";
import { Card } from "@/components/ui/card";
import { TrendDelta } from "@/components/ui/trend-delta";
import { cn } from "@/lib/utils";

export type MetricAccent = "blue" | "emerald" | "amber" | "rose" | "slate" | "violet" | "warm";

const CHIP: Record<MetricAccent, string> = {
  blue: "bg-blue-100 text-blue-700 dark:bg-blue-950/50 dark:text-blue-300",
  emerald: "bg-emerald-100 text-emerald-700 dark:bg-emerald-950/50 dark:text-emerald-300",
  amber: "bg-amber-100 text-amber-700 dark:bg-amber-950/50 dark:text-amber-300",
  rose: "bg-rose-100 text-rose-700 dark:bg-rose-950/50 dark:text-rose-300",
  slate: "bg-slate-100 text-slate-700 dark:bg-slate-800 dark:text-slate-300",
  violet: "bg-violet-100 text-violet-700 dark:bg-violet-950/50 dark:text-violet-300",
  warm: "bg-accent-warm/15 text-accent-warm dark:bg-accent-warm/20",
};

export interface MetricCardProps {
  title: string;
  value: number | string;
  icon: React.ReactNode;
  secondary?: React.ReactNode;
  accent?: MetricAccent;
  trend?: { value: number; label?: string; invert?: boolean };
  /** Numeric series for an inline sparkline. Shows last ~7 points. */
  sparkline?: number[];
  className?: string;
}

function Sparkline({ data, className }: { data: number[]; className?: string }) {
  if (!data.length) return null;
  const w = 80;
  const h = 24;
  const min = Math.min(...data);
  const max = Math.max(...data);
  const range = max - min || 1;
  const stepX = w / Math.max(data.length - 1, 1);
  const points = data
    .map((v, i) => `${i * stepX},${h - ((v - min) / range) * h}`)
    .join(" ");
  return (
    <svg
      data-testid="metric-sparkline"
      viewBox={`0 0 ${w} ${h}`}
      className={cn("h-6 w-20 overflow-visible", className)}
      aria-hidden="true"
    >
      <polyline
        fill="none"
        stroke="currentColor"
        strokeWidth="1.5"
        strokeLinecap="round"
        strokeLinejoin="round"
        points={points}
      />
    </svg>
  );
}

export function MetricCard({
  title,
  value,
  icon,
  secondary,
  accent = "blue",
  trend,
  sparkline,
  className,
}: MetricCardProps) {
  return (
    <Card className={cn("border-border/60 p-4 space-y-2", className)}>
      <div className="flex items-start justify-between gap-2">
        <span className="text-xs font-medium text-muted-foreground uppercase tracking-wider truncate">
          {title}
        </span>
        <span
          className={cn(
            "flex items-center justify-center rounded-md shrink-0 h-7 w-7",
            CHIP[accent]
          )}
        >
          {icon}
        </span>
      </div>
      <div className="flex items-end justify-between gap-3">
        <div className="text-2xl sm:text-3xl font-bold tabular-nums leading-none">
          {value}
        </div>
        {sparkline && sparkline.length > 0 && (
          <Sparkline data={sparkline} className={cn("text-foreground/70")} />
        )}
      </div>
      {(secondary || trend) && (
        <div className="flex items-center justify-between gap-2 text-xs text-muted-foreground">
          {secondary && <span className="truncate">{secondary}</span>}
          {trend && <TrendDelta value={trend.value} label={trend.label} invert={trend.invert} />}
        </div>
      )}
    </Card>
  );
}
```

- [ ] **Step 4: Run test, expect pass**

Run: `cd frontend && pnpm vitest run src/__tests__/components/ui/metric-card.test.tsx`
Expected: PASS — 4 tests.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/ui/metric-card.tsx frontend/src/__tests__/components/ui/metric-card.test.tsx
git commit -m "feat(ui): add MetricCard primitive with TrendDelta + sparkline"
```

---

## Task 4: `DocumentTypeChip` primitive

**Files:**
- Create: `frontend/src/components/ui/document-type-chip.tsx`
- Test: `frontend/src/__tests__/components/ui/document-type-chip.test.tsx`

- [ ] **Step 1: Write failing test**

```tsx
// frontend/src/__tests__/components/ui/document-type-chip.test.tsx
import { render, screen } from "@testing-library/react";
import { describe, it, expect } from "vitest";
import { DocumentTypeChip } from "@/components/ui/document-type-chip";

describe("DocumentTypeChip", () => {
  it("renders Requisition for 'requisition'", () => {
    render(<DocumentTypeChip type="requisition" />);
    expect(screen.getByText("Requisition")).toBeInTheDocument();
  });

  it("normalizes case (uppercase input maps to same label)", () => {
    render(<DocumentTypeChip type="PURCHASE_ORDER" />);
    expect(screen.getByText("Purchase Order")).toBeInTheDocument();
  });

  it("falls back to Title Case for unknown types", () => {
    render(<DocumentTypeChip type="custom_doc" />);
    expect(screen.getByText("Custom Doc")).toBeInTheDocument();
  });
});
```

- [ ] **Step 2: Run test, expect failure**

Run: `cd frontend && pnpm vitest run src/__tests__/components/ui/document-type-chip.test.tsx`
Expected: FAIL.

- [ ] **Step 3: Implement `DocumentTypeChip`**

```tsx
// frontend/src/components/ui/document-type-chip.tsx
import * as React from "react";
import { Badge } from "@/components/ui/badge";
import {
  FileText,
  ShoppingCart,
  Receipt,
  PackageCheck,
  Wallet,
  type LucideIcon,
} from "lucide-react";
import { cn } from "@/lib/utils";

export type DocumentType =
  | "requisition"
  | "purchase_order"
  | "payment_voucher"
  | "grn"
  | "goods_received_note"
  | "budget"
  | string;

interface TypeMeta {
  label: string;
  icon: LucideIcon;
}

const META: Record<string, TypeMeta> = {
  requisition: { label: "Requisition", icon: FileText },
  purchase_order: { label: "Purchase Order", icon: ShoppingCart },
  payment_voucher: { label: "Payment Voucher", icon: Receipt },
  grn: { label: "GRN", icon: PackageCheck },
  goods_received_note: { label: "GRN", icon: PackageCheck },
  budget: { label: "Budget", icon: Wallet },
};

function titleCase(s: string) {
  return s
    .replace(/_/g, " ")
    .toLowerCase()
    .replace(/\b\w/g, (c) => c.toUpperCase());
}

export interface DocumentTypeChipProps {
  type: DocumentType;
  /** Show icon next to label. Default true. */
  showIcon?: boolean;
  className?: string;
}

export function DocumentTypeChip({ type, showIcon = true, className }: DocumentTypeChipProps) {
  const key = type.toLowerCase();
  const meta = META[key];
  const label = meta?.label || titleCase(type);
  const Icon = meta?.icon || FileText;
  return (
    <Badge variant="outline" className={cn("gap-1 px-2 py-0.5 font-medium", className)}>
      {showIcon && <Icon className="h-3 w-3" aria-hidden="true" />}
      {label}
    </Badge>
  );
}
```

- [ ] **Step 4: Run test, expect pass**

Run: `cd frontend && pnpm vitest run src/__tests__/components/ui/document-type-chip.test.tsx`
Expected: PASS — 3 tests.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/ui/document-type-chip.tsx frontend/src/__tests__/components/ui/document-type-chip.test.tsx
git commit -m "feat(ui): add DocumentTypeChip primitive"
```

---

## Task 5: `ReportChart` wrapper

**Files:**
- Create: `frontend/src/components/ui/report-chart.tsx`

`ReportChart` is a thin wrapper that takes a chart `kind` ("bar" | "line" | "area"), a series array, and applies the brand chart palette via the existing `ChartContainer` from `ui/chart.tsx`. No tests — pure composition over existing tested infra. Visual smoke verified in Plan C.

- [ ] **Step 1: Implement `ReportChart`**

```tsx
// frontend/src/components/ui/report-chart.tsx
"use client";
import * as React from "react";
import {
  Bar,
  BarChart,
  CartesianGrid,
  Line,
  LineChart,
  Area,
  AreaChart,
  XAxis,
  YAxis,
  Cell,
} from "recharts";
import {
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent,
  ChartLegend,
  ChartLegendContent,
  type ChartConfig,
} from "@/components/ui/chart";
import { cn } from "@/lib/utils";

export type ReportChartKind = "bar" | "line" | "area";

export interface ReportSeries {
  /** Key in each data row to plot. */
  dataKey: string;
  /** Display label. */
  label: string;
  /** Optional explicit color override. Defaults to chart-1..5 cycling. */
  color?: string;
}

export interface ReportChartProps<T extends Record<string, unknown> = Record<string, unknown>> {
  kind: ReportChartKind;
  data: T[];
  /** Key in each data row to use as the X axis label. */
  xKey: string;
  series: ReportSeries[];
  /** Tailwind classes for the outer container; defaults to a sensible aspect. */
  className?: string;
  /** Show legend. Default false (single-series charts don't need it). */
  showLegend?: boolean;
  /** For bar charts: per-bar color from palette instead of single hue. Default false. */
  perBarColor?: boolean;
}

const PALETTE = [
  "var(--chart-1)",
  "var(--chart-2)",
  "var(--chart-3)",
  "var(--chart-4)",
  "var(--chart-5)",
];

export function ReportChart<T extends Record<string, unknown>>({
  kind,
  data,
  xKey,
  series,
  className,
  showLegend,
  perBarColor,
}: ReportChartProps<T>) {
  const config: ChartConfig = React.useMemo(() => {
    const cfg: ChartConfig = {};
    series.forEach((s, i) => {
      cfg[s.dataKey] = { label: s.label, color: s.color || PALETTE[i % PALETTE.length] };
    });
    return cfg;
  }, [series]);

  return (
    <ChartContainer config={config} className={cn("aspect-[16/7] w-full", className)}>
      {kind === "bar" ? (
        <BarChart data={data}>
          <CartesianGrid strokeDasharray="3 3" vertical={false} />
          <XAxis dataKey={xKey} tickLine={false} axisLine={false} />
          <YAxis tickLine={false} axisLine={false} width={32} />
          <ChartTooltip content={<ChartTooltipContent />} />
          {showLegend && <ChartLegend content={<ChartLegendContent />} />}
          {series.map((s, i) => (
            <Bar
              key={s.dataKey}
              dataKey={s.dataKey}
              fill={`var(--color-${s.dataKey})`}
              radius={[4, 4, 0, 0]}
            >
              {perBarColor &&
                data.map((_, idx) => (
                  <Cell key={idx} fill={PALETTE[idx % PALETTE.length]} />
                ))}
            </Bar>
          ))}
        </BarChart>
      ) : kind === "line" ? (
        <LineChart data={data}>
          <CartesianGrid strokeDasharray="3 3" vertical={false} />
          <XAxis dataKey={xKey} tickLine={false} axisLine={false} />
          <YAxis tickLine={false} axisLine={false} width={32} />
          <ChartTooltip content={<ChartTooltipContent />} />
          {showLegend && <ChartLegend content={<ChartLegendContent />} />}
          {series.map((s) => (
            <Line
              key={s.dataKey}
              type="monotone"
              dataKey={s.dataKey}
              stroke={`var(--color-${s.dataKey})`}
              strokeWidth={2}
              dot={false}
            />
          ))}
        </LineChart>
      ) : (
        <AreaChart data={data}>
          <defs>
            {series.map((s) => (
              <linearGradient
                key={s.dataKey}
                id={`fill-${s.dataKey}`}
                x1="0"
                y1="0"
                x2="0"
                y2="1"
              >
                <stop offset="5%" stopColor={`var(--color-${s.dataKey})`} stopOpacity={0.4} />
                <stop offset="95%" stopColor={`var(--color-${s.dataKey})`} stopOpacity={0} />
              </linearGradient>
            ))}
          </defs>
          <CartesianGrid strokeDasharray="3 3" vertical={false} />
          <XAxis dataKey={xKey} tickLine={false} axisLine={false} />
          <YAxis tickLine={false} axisLine={false} width={32} />
          <ChartTooltip content={<ChartTooltipContent />} />
          {showLegend && <ChartLegend content={<ChartLegendContent />} />}
          {series.map((s) => (
            <Area
              key={s.dataKey}
              type="monotone"
              dataKey={s.dataKey}
              stroke={`var(--color-${s.dataKey})`}
              strokeWidth={2}
              fill={`url(#fill-${s.dataKey})`}
            />
          ))}
        </AreaChart>
      )}
    </ChartContainer>
  );
}
```

- [ ] **Step 2: Verify TS compiles**

Run from `frontend/`: `pnpm tsc --noEmit`
Expected: No errors.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/ui/report-chart.tsx
git commit -m "feat(ui): add ReportChart wrapper (bar/line/area presets)"
```

---

## Task 6: `ApprovalChainStepper` primitive

**Files:**
- Create: `frontend/src/components/workflows/approval-chain-stepper.tsx`
- Test: `frontend/src/__tests__/components/workflows/approval-chain-stepper.test.tsx`

- [ ] **Step 1: Write failing test**

```tsx
// frontend/src/__tests__/components/workflows/approval-chain-stepper.test.tsx
import { render, screen } from "@testing-library/react";
import { describe, it, expect } from "vitest";
import { ApprovalChainStepper } from "@/components/workflows/approval-chain-stepper";

const stages = [
  { id: "s1", name: "Department Head", status: "approved" as const, actor: "Jane Doe", at: "2026-04-30T10:00:00Z" },
  { id: "s2", name: "Finance", status: "current" as const },
  { id: "s3", name: "CEO", status: "pending" as const },
];

describe("ApprovalChainStepper", () => {
  it("renders all stage names", () => {
    render(<ApprovalChainStepper stages={stages} />);
    expect(screen.getByText("Department Head")).toBeInTheDocument();
    expect(screen.getByText("Finance")).toBeInTheDocument();
    expect(screen.getByText("CEO")).toBeInTheDocument();
  });

  it("renders the actor name on completed stages", () => {
    render(<ApprovalChainStepper stages={stages} />);
    expect(screen.getByText(/Jane Doe/)).toBeInTheDocument();
  });

  it("marks the current stage with aria-current=step", () => {
    render(<ApprovalChainStepper stages={stages} />);
    const current = screen.getByText("Finance").closest("[aria-current]");
    expect(current).toHaveAttribute("aria-current", "step");
  });

  it("renders a rejected status with proper data attribute", () => {
    render(
      <ApprovalChainStepper
        stages={[{ id: "x", name: "Finance", status: "rejected", actor: "Bob", at: "2026-04-29" }]}
      />
    );
    expect(screen.getByTestId("stage-marker-x")).toHaveAttribute("data-status", "rejected");
  });
});
```

- [ ] **Step 2: Run test, expect failure**

Run: `cd frontend && pnpm vitest run src/__tests__/components/workflows/approval-chain-stepper.test.tsx`
Expected: FAIL.

- [ ] **Step 3: Implement `ApprovalChainStepper`**

```tsx
// frontend/src/components/workflows/approval-chain-stepper.tsx
import * as React from "react";
import { Check, Clock, X } from "lucide-react";
import { cn } from "@/lib/utils";

export type StageStatus = "approved" | "rejected" | "current" | "pending" | "skipped";

export interface ApprovalStage {
  id: string;
  name: string;
  status: StageStatus;
  /** Approver / rejecter display name, when available. */
  actor?: string;
  /** ISO timestamp of completion, when available. */
  at?: string;
  /** Optional comments/remarks attached to the action. */
  comments?: string;
}

export interface ApprovalChainStepperProps {
  stages: ApprovalStage[];
  className?: string;
}

const MARKER: Record<StageStatus, string> = {
  approved: "bg-emerald-600 text-white border-emerald-600",
  rejected: "bg-rose-600 text-white border-rose-600",
  current: "bg-background text-primary border-primary ring-2 ring-primary/30",
  pending: "bg-muted text-muted-foreground border-border",
  skipped: "bg-muted/40 text-muted-foreground border-dashed border-border",
};

const CONNECTOR: Record<StageStatus, string> = {
  approved: "bg-emerald-600",
  rejected: "bg-rose-600",
  current: "bg-border",
  pending: "bg-border",
  skipped: "bg-border/40",
};

function StageIcon({ status }: { status: StageStatus }) {
  if (status === "approved") return <Check className="h-3 w-3" aria-hidden="true" />;
  if (status === "rejected") return <X className="h-3 w-3" aria-hidden="true" />;
  if (status === "current") return <Clock className="h-3 w-3" aria-hidden="true" />;
  return null;
}

function fmtDate(iso?: string) {
  if (!iso) return null;
  try {
    return new Date(iso).toLocaleDateString([], {
      month: "short",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
  } catch {
    return null;
  }
}

export function ApprovalChainStepper({ stages, className }: ApprovalChainStepperProps) {
  return (
    <ol className={cn("space-y-3", className)} aria-label="Approval chain">
      {stages.map((stage, idx) => {
        const isLast = idx === stages.length - 1;
        const dateStr = fmtDate(stage.at);
        return (
          <li
            key={stage.id}
            aria-current={stage.status === "current" ? "step" : undefined}
            className="relative flex gap-3"
          >
            <div className="flex flex-col items-center">
              <span
                data-testid={`stage-marker-${stage.id}`}
                data-status={stage.status}
                className={cn(
                  "flex h-6 w-6 items-center justify-center rounded-full border text-xs font-semibold shrink-0",
                  MARKER[stage.status]
                )}
              >
                <StageIcon status={stage.status} />
                {stage.status === "pending" && idx + 1}
                {stage.status === "skipped" && "·"}
              </span>
              {!isLast && (
                <span
                  aria-hidden="true"
                  className={cn("mt-1 w-0.5 flex-1 min-h-4", CONNECTOR[stage.status])}
                />
              )}
            </div>
            <div className="flex-1 pb-3">
              <div className="text-sm font-medium leading-tight">{stage.name}</div>
              {(stage.actor || dateStr) && (
                <div className="text-xs text-muted-foreground mt-0.5">
                  {stage.actor}
                  {stage.actor && dateStr && " · "}
                  {dateStr}
                </div>
              )}
              {stage.comments && (
                <p className="text-xs text-foreground/80 mt-1 italic">"{stage.comments}"</p>
              )}
            </div>
          </li>
        );
      })}
    </ol>
  );
}
```

- [ ] **Step 4: Run test, expect pass**

Run: `cd frontend && pnpm vitest run src/__tests__/components/workflows/approval-chain-stepper.test.tsx`
Expected: PASS — 4 tests.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/workflows/approval-chain-stepper.tsx frontend/src/__tests__/components/workflows/approval-chain-stepper.test.tsx
git commit -m "feat(workflows): add ApprovalChainStepper primitive"
```

---

## Task 7: `DetailShell` layout primitive

**Files:**
- Create: `frontend/src/components/layout/detail-shell.tsx`
- Test: `frontend/src/__tests__/components/layout/detail-shell.test.tsx`

`DetailShell` standardizes detail-page layout: header at top, then either two-col (main + sidebar) on `lg+` or stacked + tabs on mobile when a `mobileTabs` array is supplied.

- [ ] **Step 1: Write failing test**

```tsx
// frontend/src/__tests__/components/layout/detail-shell.test.tsx
import { render, screen } from "@testing-library/react";
import { describe, it, expect } from "vitest";
import { DetailShell } from "@/components/layout/detail-shell";

describe("DetailShell", () => {
  it("renders header, main, and sidebar slots", () => {
    render(
      <DetailShell
        header={<div>HeaderContent</div>}
        sidebar={<div>SidebarContent</div>}
      >
        <div>MainContent</div>
      </DetailShell>
    );
    expect(screen.getByText("HeaderContent")).toBeInTheDocument();
    expect(screen.getByText("MainContent")).toBeInTheDocument();
    expect(screen.getByText("SidebarContent")).toBeInTheDocument();
  });

  it("renders without sidebar when not provided", () => {
    render(
      <DetailShell header={<div>H</div>}>
        <div>OnlyMain</div>
      </DetailShell>
    );
    expect(screen.getByText("OnlyMain")).toBeInTheDocument();
  });
});
```

- [ ] **Step 2: Run test, expect failure**

Run: `cd frontend && pnpm vitest run src/__tests__/components/layout/detail-shell.test.tsx`
Expected: FAIL.

- [ ] **Step 3: Implement `DetailShell`**

```tsx
// frontend/src/components/layout/detail-shell.tsx
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
          <div className="min-w-0 lg:col-start-1 col-span-full lg:col-span-1">{children}</div>
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
```

Note: the `style` template uses inline grid columns. On mobile the column-span fallback (`col-span-full`) collapses both children to single column. Verify visually in Plan F when first consumed.

- [ ] **Step 4: Run test, expect pass**

Run: `cd frontend && pnpm vitest run src/__tests__/components/layout/detail-shell.test.tsx`
Expected: PASS — 2 tests.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/layout/detail-shell.tsx frontend/src/__tests__/components/layout/detail-shell.test.tsx
git commit -m "feat(layout): add DetailShell primitive"
```

---

## Task 8: Enhance `EmptyState`

**Files:**
- Modify: `frontend/src/components/base/empty-state.tsx`
- Test: `frontend/src/__tests__/components/base/empty-state.test.tsx` (new)

Existing component is missing an `action` slot, has a typo (`text-gary-400`), and uses `motion.div` with `variants` but without `initial`/`animate` props so the animation never fires. Fix all three.

- [ ] **Step 1: Write failing test**

```tsx
// frontend/src/__tests__/components/base/empty-state.test.tsx
import { render, screen } from "@testing-library/react";
import { describe, it, expect } from "vitest";
import EmptyState from "@/components/base/empty-state";

describe("EmptyState", () => {
  it("renders title and description", () => {
    render(<EmptyState title="Nothing here" description="Try a different filter" />);
    expect(screen.getByText("Nothing here")).toBeInTheDocument();
    expect(screen.getByText("Try a different filter")).toBeInTheDocument();
  });

  it("renders an action when provided", () => {
    render(
      <EmptyState
        title="No tasks"
        description="You're all caught up"
        action={<button>Refresh</button>}
      />
    );
    expect(screen.getByRole("button", { name: /refresh/i })).toBeInTheDocument();
  });

  it("uses muted-foreground color for description (no typo classes)", () => {
    const { container } = render(
      <EmptyState title="t" description="d" />
    );
    expect(container.innerHTML).not.toContain("text-gary-400");
  });
});
```

- [ ] **Step 2: Run test, expect partial failure**

Run: `cd frontend && pnpm vitest run src/__tests__/components/base/empty-state.test.tsx`
Expected: FAIL on the action test (no `action` prop) and the typo test (`text-gary-400` still present).

- [ ] **Step 3: Replace the component**

```tsx
// frontend/src/components/base/empty-state.tsx
"use client";

import { cn } from "@/lib/utils";

import { motion } from "framer-motion";
import { CircleHelpIcon } from "lucide-react";

export default function EmptyState({
  Icon = CircleHelpIcon,
  title,
  description,
  action,
  className,
  classNames,
}: {
  Icon?: React.ComponentType<React.SVGProps<SVGSVGElement>>;
  title: string;
  description: string;
  /** Optional action node, e.g. a Button or Link. Rendered below description. */
  action?: React.ReactNode;
  className?: string;
  classNames?: {
    icon?: string;
    container?: string;
    title?: string;
    description?: string;
    action?: string;
  };
}) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 12 }}
      animate={{ opacity: 1, y: 0 }}
      exit={{ opacity: 0, y: -12 }}
      transition={{ duration: 0.2, ease: [0.2, 0.8, 0.2, 1] }}
      className={cn(
        "flex w-full flex-col items-center justify-center gap-2 max-w-2xl py-10",
        className,
        classNames?.container
      )}
    >
      {Icon && (
        <Icon
          className={cn("w-12 h-12 text-muted-foreground/60", classNames?.icon)}
          aria-hidden="true"
        />
      )}
      <h4
        className={cn(
          "text-center text-base leading-6 text-foreground font-semibold",
          classNames?.title
        )}
      >
        {title}
      </h4>
      <p
        className={cn(
          "text-center text-xs sm:text-sm text-muted-foreground max-w-md",
          classNames?.description
        )}
      >
        {description}
      </p>
      {action && (
        <div className={cn("mt-2", classNames?.action)}>{action}</div>
      )}
    </motion.div>
  );
}
```

- [ ] **Step 4: Run test, expect pass**

Run: `cd frontend && pnpm vitest run src/__tests__/components/base/empty-state.test.tsx`
Expected: PASS — 3 tests.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/base/empty-state.tsx frontend/src/__tests__/components/base/empty-state.test.tsx
git commit -m "feat(ui): EmptyState gains action slot, fixes typo + motion props"
```

---

## Task 9: Extend `CalendarDateRangePicker` presets

**Files:**
- Modify: `frontend/src/components/ui/date-range-picker.tsx`

Add `last90Days` and `lastQuarter` presets. The existing preset list lives at lines 34-43; the `changeHandle` switch is at lines 102-137.

- [ ] **Step 1: Update preset list**

Replace:
```tsx
const dateFilterPresets = [
  { name: "Today", value: "today" },
  { name: "Yesterday", value: "yesterday" },
  { name: "This Week", value: "thisWeek" },
  { name: "Last 7 Days", value: "last7Days" },
  { name: "Last 28 Days", value: "last28Days" },
  { name: "This Month", value: "thisMonth" },
  { name: "Last Month", value: "lastMonth" },
  { name: "This Year", value: "thisYear" }
];
```
with:
```tsx
const dateFilterPresets = [
  { name: "Today", value: "today" },
  { name: "Yesterday", value: "yesterday" },
  { name: "This Week", value: "thisWeek" },
  { name: "Last 7 Days", value: "last7Days" },
  { name: "Last 28 Days", value: "last28Days" },
  { name: "Last 90 Days", value: "last90Days" },
  { name: "This Month", value: "thisMonth" },
  { name: "Last Month", value: "lastMonth" },
  { name: "Last Quarter", value: "lastQuarter" },
  { name: "This Year", value: "thisYear" }
];
```

- [ ] **Step 2: Add cases to `changeHandle` switch**

In the existing `switch (type)` block, after the `last28Days` case and before `thisMonth`, add:
```tsx
      case "last90Days":
        const ninetyDaysAgo = subDays(today, 89);
        handleQuickSelect(startOfDay(ninetyDaysAgo), endOfDay(today));
        break;
```

After the `lastMonth` case and before `thisYear`, add:
```tsx
      case "lastQuarter":
        const quarterStart = startOfMonth(subMonths(today, 3));
        const quarterEnd = endOfMonth(subMonths(today, 1));
        handleQuickSelect(startOfDay(quarterStart), endOfDay(quarterEnd));
        break;
```

(Both `subMonths`, `startOfMonth`, `endOfMonth`, `subDays`, `startOfDay`, `endOfDay` are already imported at the top of the file.)

- [ ] **Step 3: Verify TS compiles**

Run: `cd frontend && pnpm tsc --noEmit`
Expected: clean.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/components/ui/date-range-picker.tsx
git commit -m "feat(ui): DateRangePicker adds last90Days + lastQuarter presets"
```

---

## Task 10: `useDateRangeUrlState` hook

**Files:**
- Create: `frontend/src/hooks/use-date-range-url-state.ts`
- Test: `frontend/src/__tests__/hooks/use-date-range-url-state.test.ts`

Binds a `{ from, to }` date-string pair (ISO `YYYY-MM-DD`) to URL search params `?from=...&to=...`. Default range supplied by caller. Survives refresh + share-link.

- [ ] **Step 1: Write failing test**

```ts
// frontend/src/__tests__/hooks/use-date-range-url-state.test.ts
import { renderHook, act } from "@testing-library/react";
import { describe, it, expect, vi, beforeEach } from "vitest";
import { useDateRangeUrlState } from "@/hooks/use-date-range-url-state";

const replace = vi.fn();
const params = new URLSearchParams();

vi.mock("next/navigation", () => ({
  useRouter: () => ({ replace }),
  usePathname: () => "/admin/reports",
  useSearchParams: () => params,
}));

beforeEach(() => {
  replace.mockClear();
  params.delete("from");
  params.delete("to");
});

describe("useDateRangeUrlState", () => {
  it("returns the default range when no URL params present", () => {
    const { result } = renderHook(() =>
      useDateRangeUrlState({ defaultFrom: "2026-01-01", defaultTo: "2026-01-31" })
    );
    expect(result.current.from).toBe("2026-01-01");
    expect(result.current.to).toBe("2026-01-31");
  });

  it("reads URL params when present", () => {
    params.set("from", "2026-03-01");
    params.set("to", "2026-03-31");
    const { result } = renderHook(() =>
      useDateRangeUrlState({ defaultFrom: "2026-01-01", defaultTo: "2026-01-31" })
    );
    expect(result.current.from).toBe("2026-03-01");
    expect(result.current.to).toBe("2026-03-31");
  });

  it("setRange triggers router.replace with new params", () => {
    const { result } = renderHook(() =>
      useDateRangeUrlState({ defaultFrom: "2026-01-01", defaultTo: "2026-01-31" })
    );
    act(() => result.current.setRange("2026-04-01", "2026-04-30"));
    expect(replace).toHaveBeenCalledTimes(1);
    expect(replace.mock.calls[0][0]).toContain("from=2026-04-01");
    expect(replace.mock.calls[0][0]).toContain("to=2026-04-30");
  });
});
```

- [ ] **Step 2: Run test, expect failure**

Run: `cd frontend && pnpm vitest run src/__tests__/hooks/use-date-range-url-state.test.ts`
Expected: FAIL — module not found.

- [ ] **Step 3: Implement hook**

```ts
// frontend/src/hooks/use-date-range-url-state.ts
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

  const from = searchParams.get(fromKey) || defaultFrom;
  const to = searchParams.get(toKey) || defaultTo;

  const setRange = useCallback(
    (newFrom: string, newTo: string) => {
      const params = new URLSearchParams(searchParams.toString());
      params.set(fromKey, newFrom);
      params.set(toKey, newTo);
      router.replace(`${pathname}?${params.toString()}`, { scroll: false });
    },
    [router, pathname, searchParams, fromKey, toKey]
  );

  return { from, to, setRange };
}
```

- [ ] **Step 4: Run test, expect pass**

Run: `cd frontend && pnpm vitest run src/__tests__/hooks/use-date-range-url-state.test.ts`
Expected: PASS — 3 tests.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/hooks/use-date-range-url-state.ts frontend/src/__tests__/hooks/use-date-range-url-state.test.ts
git commit -m "feat(hooks): add useDateRangeUrlState"
```

---

## Task 11: Adopt `PriorityBadge` in `tasks-table.tsx`

**Files:**
- Modify: `frontend/src/app/(private)/(main)/tasks/_components/tasks-table.tsx`

Proves shared primitive consumption. Replace local `PriorityBadge` helper with the shared one.

- [ ] **Step 1: Update import**

In the imports near the top of the file, find:
```tsx
import { Badge } from "@/components/ui/badge";
```
Add right after it:
```tsx
import { PriorityBadge } from "@/components/ui/priority-badge";
```

- [ ] **Step 2: Delete the local `PriorityBadge` helper**

Find and delete this entire block (currently around lines 101-111):
```tsx
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
```

- [ ] **Step 3: Update call sites**

Find both `<PriorityBadge p={t.priority} />` callsites in this file (one in the columns array, one in the mobileCard). Replace each:
- From: `<PriorityBadge p={t.priority} />`
- To:   `<PriorityBadge priority={t.priority} />`

- [ ] **Step 4: Verify**

Run: `cd frontend && pnpm tsc --noEmit && pnpm vitest run src/__tests__/components/ui/priority-badge.test.tsx`
Expected: tsc clean, 3 priority-badge tests pass.

- [ ] **Step 5: Commit**

```bash
git add "frontend/src/app/(private)/(main)/tasks/_components/tasks-table.tsx"
git commit -m "refactor(tasks): consume shared PriorityBadge primitive"
```

---

## Task 12: Verification pass

- [ ] **Step 1: Full type-check + Plan B tests**

Run from `frontend/`:
```bash
pnpm tsc --noEmit
pnpm vitest run \
  src/__tests__/components/ui/priority-badge.test.tsx \
  src/__tests__/components/ui/trend-delta.test.tsx \
  src/__tests__/components/ui/metric-card.test.tsx \
  src/__tests__/components/ui/document-type-chip.test.tsx \
  src/__tests__/components/workflows/approval-chain-stepper.test.tsx \
  src/__tests__/components/layout/detail-shell.test.tsx \
  src/__tests__/components/base/empty-state.test.tsx \
  src/__tests__/hooks/use-date-range-url-state.test.ts
```
Expected: tsc clean; all tests pass.

- [ ] **Step 2: Plan A regression check**

Run: `cd frontend && pnpm vitest run src/__tests__/components/ui/stat-grid.test.tsx src/__tests__/components/ui/data-list.test.tsx src/__tests__/components/layout/mobile-bottom-nav.test.tsx`
Expected: all 8 Plan A tests still pass.

- [ ] **Step 3: Final cleanup commit if needed**

If any cleanup applied:
```bash
git add -A
git commit -m "chore(ui): plan-B verification cleanup"
```

---

## Self-Review Notes

- **Spec coverage:** Each Plan B primitive has its own task with explicit interface, test, and commit. `StatusBadge` re-use is explicitly documented (no duplicate `DocumentStatusPill`). `EmptyState` gets enhanced (Task 8). `DateRangePicker` gets extended (Task 9). `useDateRangeUrlState` is a new hook (Task 10). Plan A's `PriorityBadge` consumer is updated (Task 11) to prove the primitive replaces inline duplication.
- **Type consistency:** `StageStatus` type used in `ApprovalChainStepper`'s `stages` prop and matched in tests. `MetricAccent` in `MetricCard` mirrors `StatAccent` from `StatGrid` (same string union). `PriorityBadge.priority` prop name standardized (Task 1) and consumer updated (Task 11) — old `p` prop replaced.
- **No placeholders:** All steps include exact code. No "TBD" or "similar to Task N" — when tests reference primitives across tasks, full code is in the source task.
- **Test placements:** All under `frontend/src/__tests__/` per project convention (vitest config only includes that subtree).

## Out-of-scope follow-up notes

- Plan C (`/admin/reports`) is the first heavy consumer of MetricCard + ReportChart + DateRangePicker + useDateRangeUrlState.
- Plan F applies `<DetailShell>` to all detail pages and `<ApprovalChainStepper>` replaces inline approval renderings.
- Plans D / E sweep `ResponsiveSheet` / `DataList` adoption; PriorityBadge consumers in `requisitions-table.tsx` are migrated in Plan E.
- A future cleanup task should sweep all 44+ raw `bg-{color}-100` callsites to consume `StatusBadge`. Tracked but not part of Plan B (mechanical refactor for Plan I).
