# UI Revamp — Plan D: Modal Sweep + Plan C Carry-Forwards

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Convert 6 high-impact procurement form modals from `<Dialog>` to `<ResponsiveSheet>` (Drawer on mobile, Dialog on desktop), and clear 4 Plan C carry-forwards (I1/I2/I3/M5).

**Architecture:** `<ResponsiveSheet>` (Plan B Task 5) wraps Radix `Dialog` on `md+` and `vaul` `Drawer` on `<md`. Migration is mechanical: replace `<Dialog>`/`<DialogContent>` pair with `<ResponsiveSheet>`, lift `DialogTitle`/`DialogDescription` into props, lift `DialogFooter` content into `footer` prop. Form bodies stay unchanged. Each modal is one task — implementer reads the existing file, applies the pattern, verifies TS, commits. Confirmation-style dialogs (submit/delete/mark-paid) and the global PDF/attachment preview dialogs are NOT in scope — those are small enough to work on mobile already.

**Tech Stack:** Next.js App Router, TypeScript, Tailwind v4, ShadCN UI, Radix Dialog, `vaul` 1.x (already used by `<ResponsiveSheet>`).

**Out-of-scope (handled in later plans):**
- Plan E — list tables sweep across procurement
- Plan F — detail-page split + DetailShell adoption
- Plan G — vendor detail page + non-admin reports hub decision
- Plan H — approval flow polish (signature pad migration deferred here since signature pad is embedded in `approval-action-modal`, not a standalone modal)
- Plan I — cleanup / stub removal

---

## Modals NOT in scope (intentional)

These are confirmation-only or read-only dialogs; they are small enough to work on mobile already and do NOT benefit from drawer treatment:
- `requisition-submit-dialog.tsx` (confirm-only)
- `payment-voucher-submit-dialog.tsx` (confirm-only)
- `purchase-order-submit-dialog.tsx` (confirm-only)
- `grn-submit-dialog.tsx` (confirm-only)
- `budget-submit-dialog.tsx` (confirm-only)
- `budget-edit-dialog.tsx` (small)
- `budget-delete-dialog.tsx` (alert-dialog confirmation)
- `mark-paid-dialog.tsx` (small)
- `approval-confirmation-dialog.tsx` (confirm-only)
- `attachment-preview-dialog.tsx` (read-only viewer; needs full screen on mobile but is a separate Plan I item)
- `pdf-preview-dialog.tsx` (read-only viewer; separate Plan I item)
- `create-user-dialog.tsx` (admin only, separate sweep)
- `create-organization-modal.tsx` (welcome flow, separate sweep)
- `approval-action-modal.tsx` — kept on Dialog because the embedded signature pad needs careful Plan H treatment
- `edit-purchase-order-dialog.tsx` — out of scope; PO detail page splits in Plan F
- `create-budget-dialog.tsx` + `create-po-from-requisition-dialog.tsx` — defer to a follow-up Plan D2 if needed; current Plan D fits 6 modals + 4 carry-forwards in one shippable PR

---

## File Structure

**Modify (carry-forwards):**
- `frontend/src/components/ui/data-list.tsx` — add `align?: "left" | "right" | "center"` to `DataListColumn` (Plan C I2)
- `frontend/src/__tests__/components/ui/data-list.test.tsx` — add test for `align` prop
- `frontend/src/app/(private)/admin/_components/admin-reports-client.tsx` — lift `useDateRangeUrlState`, pass `from`/`to`/`setRange` down to `ReportsHeader` as props (Plan C I1)
- `frontend/src/app/(private)/admin/_components/reports-header.tsx` — accept `from`/`to`/`setRange` props instead of calling the hook directly
- `frontend/src/__tests__/components/admin/reports-header.test.tsx` — update test for new prop signature; switch to `findByRole("menuitem", ...)` (Plan C I3)
- `frontend/src/app/(private)/admin/_components/user-activity-reports.tsx` — promote `rejections` column priority `md → lg` (Plan C M5)

**Modify (modal sweep — 6 files):**
- `frontend/src/app/(private)/(main)/vendors/_components/vendor-form-sheet.tsx` — Dialog → ResponsiveSheet (file misnamed since inception)
- `frontend/src/app/(private)/(main)/requisitions/_components/create-requisition-dialog.tsx` — Dialog → ResponsiveSheet
- `frontend/src/app/(private)/(main)/grn/_components/create-grn-dialog.tsx` — Dialog → ResponsiveSheet
- `frontend/src/app/(private)/(main)/grn/[id]/_components/quality-issue-dialog.tsx` — Dialog → ResponsiveSheet
- `frontend/src/app/(private)/(main)/payment-vouchers/_components/create-pv-from-po-dialog.tsx` — Dialog → ResponsiveSheet
- `frontend/src/app/(private)/(main)/budgets/[id]/_components/add-budget-item-dialog.tsx` — Dialog → ResponsiveSheet
- `frontend/src/app/(private)/(main)/budgets/[id]/_components/edit-budget-item-dialog.tsx` — Dialog → ResponsiveSheet

(Total: 7 dialog files but budgets add+edit are batched into one task since they share structure.)

---

## Migration pattern (used by Tasks 5–10)

This pattern applies when migrating any `<Dialog>`-based form modal to `<ResponsiveSheet>`. Each modal task points back here.

**Before (typical pattern):**
```tsx
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";

<Dialog open={open} onOpenChange={onOpenChange}>
  <DialogContent className="sm:max-w-lg max-h-[90svh] overflow-y-auto">
    <DialogHeader>
      <DialogTitle>Create requisition</DialogTitle>
      <DialogDescription>Submit a new request for approval.</DialogDescription>
    </DialogHeader>
    {/* form body */}
    <FormFields />
    <DialogFooter>
      <Button variant="outline" onClick={onClose}>Cancel</Button>
      <Button type="submit">Save</Button>
    </DialogFooter>
  </DialogContent>
</Dialog>
```

**After:**
```tsx
import { ResponsiveSheet } from "@/components/ui/responsive-sheet";

<ResponsiveSheet
  open={open}
  onOpenChange={onOpenChange}
  title="Create requisition"
  description="Submit a new request for approval."
  desktopMaxWidth="sm:max-w-lg"
  footer={
    <div className="flex flex-wrap items-center justify-end gap-2">
      <Button variant="outline" onClick={onClose}>Cancel</Button>
      <Button type="submit" form="requisition-form">Save</Button>
    </div>
  }
>
  {/* form body */}
  <FormFields />
</ResponsiveSheet>
```

**Notes:**
- Drop `DialogHeader`, `DialogTitle`, `DialogDescription`, `DialogFooter`, `DialogContent`, `Dialog` imports — only keep `Dialog` if it's used elsewhere in the file (rare).
- Move title/description STRINGS into the `title` and `description` props. If the original used JSX nodes (e.g., `<DialogDescription asChild><div>…</div></DialogDescription>`), pass JSX directly as `description={<div>…</div>}`.
- Footer buttons go inside the `footer` prop wrapped in a `<div className="flex flex-wrap items-center justify-end gap-2">`.
- Form `<form>` elements need an `id` attribute and the submit button needs `form="that-id"` if the button is in `footer` (footer renders OUTSIDE the form's DOM tree on mobile because vaul puts content in a separate container). Add the id if absent.
- `desktopMaxWidth` defaults to `"sm:max-w-lg"`. Keep whatever max-width the original `DialogContent` had (e.g., `sm:max-w-2xl`).
- Drop the `max-h-[90svh] overflow-y-auto p-0` classes from the old DialogContent — `<ResponsiveSheet>` handles all that internally.
- If the dialog has internal padding control beyond the standard, pass it via `className`.

---

## Task 1: DataList `align` column prop (Plan C I2)

**Files:**
- Modify: `frontend/src/components/ui/data-list.tsx`
- Modify: `frontend/src/__tests__/components/ui/data-list.test.tsx`

- [ ] **Step 1: Write failing test**

Add this test inside the existing `describe("DataList", ...)` block in `frontend/src/__tests__/components/ui/data-list.test.tsx`:

```tsx
  it("applies align class to header and cell when align is set", () => {
    const { container } = render(
      <DataList<Row>
        rows={[{ id: "1", name: "alpha" }]}
        getRowId={(r) => r.id}
        columns={[{ id: "name", header: "Name", cell: (r) => r.name, align: "right" }]}
        mobileCard={(r) => <div>{r.name}</div>}
      />
    );
    // Header and cell on desktop branch should both have text-right
    const ths = container.querySelectorAll("th");
    const tds = container.querySelectorAll("td");
    expect(Array.from(ths).some((el) => el.className.includes("text-right"))).toBe(true);
    expect(Array.from(tds).some((el) => el.className.includes("text-right"))).toBe(true);
  });
```

- [ ] **Step 2: Run test, expect failure**

Run: `cd frontend && pnpm vitest run src/__tests__/components/ui/data-list.test.tsx`
Expected: 5 pass + 1 fail (new align test).

- [ ] **Step 3: Update `DataList`**

Open `frontend/src/components/ui/data-list.tsx`. Find the `DataListColumn` interface and add an `align?` field:

```tsx
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
  /** Text alignment for header and cell. Default 'left'. */
  align?: "left" | "right" | "center";
  className?: string;
}
```

Add an `ALIGN` map next to the existing `HIDE` map (somewhere near the top of the module, before `export function DataList`):

```tsx
const ALIGN: Record<NonNullable<DataListColumn<unknown>["align"]>, string> = {
  left: "",
  right: "text-right",
  center: "text-center",
};
```

Find the four places where header/cell `className` is composed via `cn(HIDE[c.priority || "always"], c.className)` (two in the loading branch, two in the live-data branch). For each, add `ALIGN[c.align || "left"]` to the `cn` call:

```tsx
className={cn(HIDE[c.priority || "always"], ALIGN[c.align || "left"], c.className)}
```

The mobile card branch ignores columns and doesn't need updating.

- [ ] **Step 4: Run test, expect pass**

Run: `cd frontend && pnpm vitest run src/__tests__/components/ui/data-list.test.tsx`
Expected: 6 pass.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/ui/data-list.tsx frontend/src/__tests__/components/ui/data-list.test.tsx
git commit -m "feat(ui): DataList column gains align prop"
```

---

## Task 2: Lift `useDateRangeUrlState` to single callsite (Plan C I1)

**Files:**
- Modify: `frontend/src/app/(private)/admin/_components/reports-header.tsx`
- Modify: `frontend/src/app/(private)/admin/_components/admin-reports-client.tsx`

`ReportsHeader` currently calls `useDateRangeUrlState` itself. `AdminReportsClient` ALSO calls it. URL is the source of truth so they read the same values, but two callsites = two `defaultRange()` definitions and twice the hooks. Lift the call up to `AdminReportsClient` and pass `from`/`to`/`setRange` (and `defaultFrom`/`defaultTo` for the picker initial value) down to `ReportsHeader` as props.

- [ ] **Step 1: Update `ReportsHeader` to accept date-range props**

Open `frontend/src/app/(private)/admin/_components/reports-header.tsx`. Replace the entire file with:

```tsx
"use client";
import * as React from "react";
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
import { cn } from "@/lib/utils";

export type ReportsExportFormat = "csv";

export interface ReportsHeaderProps {
  title: string;
  subtitle: string;
  /** ISO YYYY-MM-DD. Required — owner is the parent. */
  from: string;
  /** ISO YYYY-MM-DD. */
  to: string;
  onRangeChange: (from: string, to: string) => void;
  onRefresh: () => void;
  onExport: (format: ReportsExportFormat) => void;
  isRefreshing: boolean;
  className?: string;
}

export function ReportsHeader({
  title,
  subtitle,
  from,
  to,
  onRangeChange,
  onRefresh,
  onExport,
  isRefreshing,
  className,
}: ReportsHeaderProps) {
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
            onChange={onRangeChange}
          />
        </div>
      </div>
    </div>
  );
}
```

(Removed `useDateRangeUrlState` import and the internal `defaultRange()` helper — both are now the parent's responsibility.)

- [ ] **Step 2: Update `AdminReportsClient` to pass down the props**

Open `frontend/src/app/(private)/admin/_components/admin-reports-client.tsx`. Find the `<ReportsHeader>` JSX and change it from:
```tsx
      <ReportsHeader
        title="Admin Reports"
        subtitle="Workflow approvals, user activity, system metrics"
        onRefresh={handleRefresh}
        onExport={handleExport}
        isRefreshing={isRefreshing}
      />
```
to:
```tsx
      <ReportsHeader
        title="Admin Reports"
        subtitle="Workflow approvals, user activity, system metrics"
        from={from}
        to={to}
        onRangeChange={setRange}
        onRefresh={handleRefresh}
        onExport={handleExport}
        isRefreshing={isRefreshing}
      />
```

(`from`, `to`, `setRange` are already returned by the `useDateRangeUrlState` call in this file. Verify that destructuring exists; if it currently destructures only `from` and `to`, add `setRange`.)

- [ ] **Step 3: Verify TS compiles**

Run: `cd frontend && pnpm tsc --noEmit`
Expected: clean.

- [ ] **Step 4: Commit**

```bash
git add "frontend/src/app/(private)/admin/_components/reports-header.tsx" "frontend/src/app/(private)/admin/_components/admin-reports-client.tsx"
git commit -m "refactor(admin): lift useDateRangeUrlState to single callsite"
```

---

## Task 3: Update `reports-header` test for new signature + menuitem query (Plan C I3)

**Files:**
- Modify: `frontend/src/__tests__/components/admin/reports-header.test.tsx`

- [ ] **Step 1: Replace the test file**

Replace the entire contents of `frontend/src/__tests__/components/admin/reports-header.test.tsx` with:

```tsx
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, it, expect, vi } from "vitest";
import { ReportsHeader } from "@/app/(private)/admin/_components/reports-header";

const baseProps = {
  title: "Admin Reports",
  subtitle: "Workflow approvals, user activity, system metrics",
  from: "2026-01-01",
  to: "2026-01-31",
  onRangeChange: vi.fn(),
  onRefresh: vi.fn(),
  onExport: vi.fn(),
  isRefreshing: false,
};

describe("ReportsHeader", () => {
  it("renders title and subtitle", () => {
    render(<ReportsHeader {...baseProps} />);
    expect(screen.getByText("Admin Reports")).toBeInTheDocument();
    expect(
      screen.getByText("Workflow approvals, user activity, system metrics")
    ).toBeInTheDocument();
  });

  it("calls onRefresh when refresh button is clicked", async () => {
    const onRefresh = vi.fn();
    const user = userEvent.setup();
    render(<ReportsHeader {...baseProps} onRefresh={onRefresh} />);
    await user.click(screen.getByRole("button", { name: /refresh/i }));
    expect(onRefresh).toHaveBeenCalledTimes(1);
  });

  it("disables refresh when isRefreshing is true", () => {
    render(<ReportsHeader {...baseProps} isRefreshing />);
    expect(screen.getByRole("button", { name: /refresh/i })).toBeDisabled();
  });

  it("invokes onExport with chosen format from menu", async () => {
    const onExport = vi.fn();
    const user = userEvent.setup();
    render(<ReportsHeader {...baseProps} onExport={onExport} />);
    await user.click(screen.getByRole("button", { name: /export/i }));
    await user.click(await screen.findByRole("menuitem", { name: /csv/i }));
    expect(onExport).toHaveBeenCalledWith("csv");
  });
});
```

(Removed `next/navigation` mock — `ReportsHeader` no longer calls `useDateRangeUrlState`. Switched export-menu query to `findByRole("menuitem", ...)` for stability when button text or other items change.)

- [ ] **Step 2: Run test**

Run: `cd frontend && pnpm vitest run src/__tests__/components/admin/reports-header.test.tsx`
Expected: 4 pass.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/__tests__/components/admin/reports-header.test.tsx
git commit -m "test(admin): reports-header test updated for prop-driven signature"
```

---

## Task 4: Promote `rejections` column priority in user activity (Plan C M5)

**Files:**
- Modify: `frontend/src/app/(private)/admin/_components/user-activity-reports.tsx`

- [ ] **Step 1: Apply edit**

In `frontend/src/app/(private)/admin/_components/user-activity-reports.tsx`, find the `rejections` column definition. It currently looks like:

```tsx
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
```

Change `priority: "md"` to `priority: "lg"`.

While editing, take advantage of the new `align` prop from Task 1 to clean up the `<span className="text-right block">` boilerplate. Replace the column block above (and the matching `approvals` and `active` columns that use the same pattern) with:

```tsx
    {
      id: "approvals",
      header: "Approvals",
      align: "right",
      priority: "md",
      cell: (u) => <span className="tabular-nums">{u.approvalCount}</span>,
    },
    {
      id: "rejections",
      header: "Rejections",
      align: "right",
      priority: "lg",
      cell: (u) => <span className="tabular-nums">{u.rejectionCount}</span>,
    },
    {
      id: "active",
      header: "Active",
      align: "right",
      priority: "lg",
      cell: (u) => <span className="tabular-nums">{u.activeDocuments}</span>,
    },
```

- [ ] **Step 2: Verify TS compiles**

Run: `cd frontend && pnpm tsc --noEmit`
Expected: clean.

- [ ] **Step 3: Commit**

```bash
git add "frontend/src/app/(private)/admin/_components/user-activity-reports.tsx"
git commit -m "refactor(admin): promote rejections column priority lg + use align prop"
```

---

## Task 5: Migrate `vendor-form-sheet.tsx` to ResponsiveSheet

**Files:**
- Modify: `frontend/src/app/(private)/(main)/vendors/_components/vendor-form-sheet.tsx`

This file is misnamed since inception ("sheet" but uses `<Dialog>`). Convert to actual `<ResponsiveSheet>`.

- [ ] **Step 1: Read the existing file**

Read `frontend/src/app/(private)/(main)/vendors/_components/vendor-form-sheet.tsx` end-to-end. Identify:
- The exported component name and props
- How `Dialog`, `DialogContent`, `DialogHeader`, `DialogTitle`, `DialogDescription`, `DialogFooter` are used
- The `<form>` element and the submit button location
- Current `desktopMaxWidth` value (probably `sm:max-w-xl` or similar)

- [ ] **Step 2: Apply migration pattern**

Apply the migration pattern documented at the top of this plan ("Migration pattern (used by Tasks 5–10)"):
1. Add `import { ResponsiveSheet } from "@/components/ui/responsive-sheet";` near the existing imports.
2. Remove the `Dialog`, `DialogContent`, `DialogHeader`, `DialogTitle`, `DialogDescription`, `DialogFooter` imports.
3. Replace the outer `<Dialog open={...} onOpenChange={...}><DialogContent>...</DialogContent></Dialog>` with `<ResponsiveSheet open={...} onOpenChange={...} title="..." description="..." desktopMaxWidth="..." footer={...}>...</ResponsiveSheet>`.
4. If the form element doesn't have an `id`, add one (e.g. `id="vendor-form"`) and add `form="vendor-form"` to the submit button in the footer prop.
5. Drop any `max-h-[90svh] overflow-y-auto p-0` classes that were on the old `DialogContent`.

- [ ] **Step 3: Verify TS compiles**

Run: `cd frontend && pnpm tsc --noEmit`
Expected: clean.

- [ ] **Step 4: Smoke check (optional)**

If able, run `cd frontend && pnpm dev` and visit `/vendors`. Click "Create vendor" or whatever opens this dialog. Verify:
- Desktop ≥768px: dialog renders unchanged
- Mobile <768px: drawer slides up from bottom
- Form submit and cancel still work

- [ ] **Step 5: Commit**

```bash
git add "frontend/src/app/(private)/(main)/vendors/_components/vendor-form-sheet.tsx"
git commit -m "refactor(vendors): vendor-form-sheet uses ResponsiveSheet"
```

---

## Task 6: Migrate `create-requisition-dialog.tsx` to ResponsiveSheet

**Files:**
- Modify: `frontend/src/app/(private)/(main)/requisitions/_components/create-requisition-dialog.tsx`

- [ ] **Step 1: Read the existing file**

Read the file. Note that this file is large (~959 LOC). The migration only touches the OUTER `<Dialog>` shell — the multi-step form body stays unchanged. Locate:
- The top-level `<Dialog open={...} onOpenChange={...}>...<DialogContent>...</DialogContent></Dialog>` block
- `DialogTitle`, `DialogDescription`, `DialogFooter` (and any `DialogHeader` wrapper)
- The submit button(s) and the `<form>` element they belong to

- [ ] **Step 2: Apply migration pattern**

Per the migration pattern at top of plan. Specifically:
1. Add `import { ResponsiveSheet } from "@/components/ui/responsive-sheet";`.
2. Remove dialog primitive imports.
3. Wrap the outer body in `<ResponsiveSheet>` with `title`, `description`, `desktopMaxWidth` (preserve original max-width — likely `sm:max-w-2xl` or `sm:max-w-3xl`), `open`, `onOpenChange`.
4. Lift the `<DialogFooter>...buttons...</DialogFooter>` content into the `footer` prop.
5. Add `id="requisition-form"` to the form element and `form="requisition-form"` to the submit button if needed.

If the file has multiple internal sections that use accordions/tabs, do NOT restructure them — only the outer dialog shell changes.

- [ ] **Step 3: Verify TS compiles**

Run: `cd frontend && pnpm tsc --noEmit`
Expected: clean.

- [ ] **Step 4: Commit**

```bash
git add "frontend/src/app/(private)/(main)/requisitions/_components/create-requisition-dialog.tsx"
git commit -m "refactor(requisitions): create-requisition-dialog uses ResponsiveSheet"
```

---

## Task 7: Migrate `create-grn-dialog.tsx` to ResponsiveSheet

**Files:**
- Modify: `frontend/src/app/(private)/(main)/grn/_components/create-grn-dialog.tsx`

- [ ] **Step 1: Read the existing file**

Read the file (~539 LOC). Locate the outer Dialog shell and the form/submit structure.

- [ ] **Step 2: Apply migration pattern**

Per the pattern at top of plan. Form id suggestion: `id="grn-form"`.

- [ ] **Step 3: Verify TS compiles**

Run: `cd frontend && pnpm tsc --noEmit`
Expected: clean.

- [ ] **Step 4: Commit**

```bash
git add "frontend/src/app/(private)/(main)/grn/_components/create-grn-dialog.tsx"
git commit -m "refactor(grn): create-grn-dialog uses ResponsiveSheet"
```

---

## Task 8: Migrate `quality-issue-dialog.tsx` to ResponsiveSheet

**Files:**
- Modify: `frontend/src/app/(private)/(main)/grn/[id]/_components/quality-issue-dialog.tsx`

- [ ] **Step 1: Read the existing file**

Read the file. Locate Dialog shell.

- [ ] **Step 2: Apply migration pattern**

Per the pattern at top of plan. Form id suggestion: `id="grn-quality-issue-form"`.

- [ ] **Step 3: Verify TS compiles**

Run: `cd frontend && pnpm tsc --noEmit`
Expected: clean.

- [ ] **Step 4: Commit**

```bash
git add "frontend/src/app/(private)/(main)/grn/[id]/_components/quality-issue-dialog.tsx"
git commit -m "refactor(grn): quality-issue-dialog uses ResponsiveSheet"
```

---

## Task 9: Migrate `create-pv-from-po-dialog.tsx` to ResponsiveSheet

**Files:**
- Modify: `frontend/src/app/(private)/(main)/payment-vouchers/_components/create-pv-from-po-dialog.tsx`

This dialog (394 LOC) embeds an "Approved POs" selection table. After migration, the embedded table will display inside a drawer on mobile — its own internal scrolling already exists, so no extra work needed.

- [ ] **Step 1: Read the existing file**

Read the file. Locate Dialog shell. Note the embedded table — it stays intact.

- [ ] **Step 2: Apply migration pattern**

Per the pattern at top of plan. Likely `desktopMaxWidth="sm:max-w-3xl"` (for the embedded table). Form id suggestion: `id="pv-from-po-form"`.

- [ ] **Step 3: Verify TS compiles**

Run: `cd frontend && pnpm tsc --noEmit`
Expected: clean.

- [ ] **Step 4: Commit**

```bash
git add "frontend/src/app/(private)/(main)/payment-vouchers/_components/create-pv-from-po-dialog.tsx"
git commit -m "refactor(pv): create-pv-from-po-dialog uses ResponsiveSheet"
```

---

## Task 10: Migrate budget item dialogs (add + edit)

**Files:**
- Modify: `frontend/src/app/(private)/(main)/budgets/[id]/_components/add-budget-item-dialog.tsx`
- Modify: `frontend/src/app/(private)/(main)/budgets/[id]/_components/edit-budget-item-dialog.tsx`

Both dialogs share the same form structure (line-item fields). Migrate both in one task.

- [ ] **Step 1: Read both files**

Read both files. They likely look almost identical, just differing in `title`/`description` and submit handler.

- [ ] **Step 2: Apply migration pattern to `add-budget-item-dialog.tsx`**

Per the pattern. Form id: `id="add-budget-item-form"`.

- [ ] **Step 3: Apply migration pattern to `edit-budget-item-dialog.tsx`**

Per the pattern. Form id: `id="edit-budget-item-form"`.

- [ ] **Step 4: Verify TS compiles**

Run: `cd frontend && pnpm tsc --noEmit`
Expected: clean.

- [ ] **Step 5: Commit**

```bash
git add "frontend/src/app/(private)/(main)/budgets/[id]/_components/add-budget-item-dialog.tsx" "frontend/src/app/(private)/(main)/budgets/[id]/_components/edit-budget-item-dialog.tsx"
git commit -m "refactor(budgets): add+edit budget-item dialogs use ResponsiveSheet"
```

---

## Task 11: Verification pass

- [ ] **Step 1: Full type-check**

Run: `cd frontend && pnpm tsc --noEmit`
Expected: clean.

- [ ] **Step 2: Plan D + carry-forward tests**

Run:
```bash
cd frontend && pnpm vitest run \
  src/__tests__/components/ui/data-list.test.tsx \
  src/__tests__/components/admin/reports-header.test.tsx \
  src/__tests__/integration/admin-reports.test.tsx \
  src/__tests__/hooks/use-date-range-url-state.test.ts
```
Expected: 6 (data-list) + 4 (reports-header) + 3 (integration) + 5 (date-range-url-state) = 18 pass.

- [ ] **Step 3: Plan A + B + C regression**

Run:
```bash
cd frontend && pnpm vitest run \
  src/__tests__/components/ui/stat-grid.test.tsx \
  src/__tests__/components/ui/priority-badge.test.tsx \
  src/__tests__/components/ui/trend-delta.test.tsx \
  src/__tests__/components/ui/metric-card.test.tsx \
  src/__tests__/components/ui/document-type-chip.test.tsx \
  src/__tests__/components/workflows/approval-chain-stepper.test.tsx \
  src/__tests__/components/layout/detail-shell.test.tsx \
  src/__tests__/components/layout/mobile-bottom-nav.test.tsx \
  src/__tests__/components/base/empty-state.test.tsx
```
Expected: all pass (Plan A 8 + Plan B 22 across these files).

- [ ] **Step 4: Manual mobile smoke (optional)**

If able, `cd frontend && pnpm dev` and on mobile viewport (390px wide):
- `/vendors` → click "New vendor" → drawer slides up, form usable
- `/requisitions` → click "Create requisition" → drawer
- `/grn` → click "Create GRN" → drawer
- `/payment-vouchers` → click "Create from PO" → drawer
- `/budgets/<id>` → click "Add item" → drawer
- `/admin/reports` → date range works, no console errors

Verify on desktop ≥768px:
- All same flows render as Dialog as before (no visual regression)

- [ ] **Step 5: Cleanup commit if needed**

```bash
git add -A
git commit -m "chore(ui): plan-D verification cleanup"
```

---

## Self-Review Notes

- **Spec coverage:** All four Plan C carry-forwards land (Tasks 1-4: I2, I1, I3, M5). Six modal targets land (Tasks 5-10): vendor + req + grn + quality + pv-from-po + budget items. Verification (Task 11) checks TS + tests + regression.
- **Type consistency:** `DataListColumn.align` type added in Task 1 is consumed in Task 4. `ReportsHeader` prop signature change in Task 2 matches the test update in Task 3 and the parent change in Task 2 step 2. Form id naming convention is documented (`<route>-form`, `add-budget-item-form`, etc.).
- **No placeholders:** Each task has explicit code or an explicit "read file + apply migration pattern (linked above)" pointer with the pattern fully documented at the top of the plan. The pattern itself is concrete code, not "fill in details".
- **Modal task structure:** Tasks 5-10 follow the same shape (read → migrate per pattern → tsc → commit). The pattern is centralized to avoid repeating ~30 lines of code six times. Plan-readers can hold Tasks 5-10 in mind together.
- **Out-of-scope discipline:** Confirmation dialogs, viewer dialogs, admin/onboarding modals, signature-pad-embedded modal, and detail-page-embedded edit dialogs are explicitly listed in "Modals NOT in scope" with reasoning.

## Plan E carry-forward notes

- Many internal forms still use `grid grid-cols-2 gap-3` without responsive collapse — Plan E should sweep these. Plan D only changes the OUTER shell.
- Form bodies remain large (e.g. `create-requisition-dialog.tsx` 959 LOC). Plan F (detail-page splits) and a follow-up plan for form-body decomposition should chip away at these.
- `approval-action-modal.tsx` migration deferred to Plan H (signature pad needs full mobile treatment).
- `attachment-preview-dialog.tsx` and `pdf-preview-dialog.tsx` are read-only viewers needing fullscreen on mobile — separate Plan I item with different UX (full-screen sheet).
