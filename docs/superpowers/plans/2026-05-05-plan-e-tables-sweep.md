# UI Revamp — Plan E: Procurement List Tables Sweep

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Convert all six procurement list tables from `DataTable` (TanStack wrapper) or raw `useReactTable` to `<DataList>` + `<FilterBar>` with mobile cards, adopting shared status/priority/doc-type primitives.

**Architecture:** Each route's table currently builds its own column array, filters (sometimes), and pagination. Plan E swaps the rendering layer to `DataList` (responsive table+card primitive from Plan A) + `FilterBar` (search + filter row from Plan A). Existing data hooks, server actions, and per-row action menus stay intact — only the rendering shell + filter UI change. Mobile cards adopt a consistent two-row meta layout: row1 = doc# + status pill, row2 = priority/type/date meta + action row. `budgets-table.tsx` is the only route still using raw `useReactTable` directly — it gets the biggest cleanup.

**Tech Stack:** Next.js App Router, TypeScript, Tailwind v4, ShadCN UI, DataList (Plan A), FilterBar (Plan A), StatusBadge (existing canonical), PriorityBadge + DocumentTypeChip (Plan B), Vitest.

**Out-of-scope (handled in later plans):**
- Plan F — detail-page splits + DetailShell adoption
- Plan G — vendor detail page + non-admin reports hub decision; remaining modal migrations (approval-action, edit-PO, create-budget, create-po-from-req, confirmation dialogs, PDF/attachment viewers)
- Plan H — approval flow polish
- Plan I — cleanup / stub removal

---

## Existing infrastructure (reuse, don't duplicate)

- `DataList` + `DataListColumn` (`ui/data-list.tsx`) — Plan A primitive. `align?: "left"|"right"|"center"` and `priority?: "always"|"md"|"lg"` available.
- `FilterBar` (`ui/filter-bar.tsx`) — Plan A primitive. Slots: `search`, `filters`, `meta`, `hasActiveFilters`, `onReset`.
- `CustomPagination` (`ui/custom-pagination.tsx`) — keep using as-is for paginated routes.
- `StatusBadge` (`status-badge.tsx`) — canonical; supports `type="document"|"action"|"execution"|"approval"|...`
- `PriorityBadge` (`ui/priority-badge.tsx`) — Plan B Task 1.
- `DocumentTypeChip` (`ui/document-type-chip.tsx`) — Plan B Task 4.
- Existing per-route action menu components (e.g. `purchase-orders-table-actions.tsx`, etc) — keep as-is, render inside a `cell:` callback.

---

## File Structure

**Modify:**
- `frontend/src/components/ui/data-list.tsx` — add JSDoc note that `align` is desktop-only (Plan D carry-forward)
- `frontend/src/app/(private)/(main)/requisitions/_components/requisitions-table.tsx` (494 LOC → DataList + FilterBar)
- `frontend/src/app/(private)/(main)/purchase-orders/_components/purchase-orders-table.tsx` (232 LOC → DataList + FilterBar with status/vendor/date filters)
- `frontend/src/app/(private)/(main)/payment-vouchers/_components/payment-vouchers-table.tsx` (249 LOC → DataList + FilterBar with status/method/date filters)
- `frontend/src/app/(private)/(main)/grn/_components/grn-table.tsx` (316 LOC → DataList + FilterBar with status/po/date filters)
- `frontend/src/app/(private)/(main)/budgets/_components/budgets-table.tsx` (338 LOC raw `useReactTable` → DataList + FilterBar with year/department/status filters)
- `frontend/src/app/(private)/(main)/vendors/_components/vendors-table.tsx` (301 LOC → DataList + FilterBar with category/status filters; add server pagination if hook supports it)

---

## Migration pattern (used by Tasks 2–7)

Each table task follows this shape. The pattern is documented once here; per-task notes call out route-specific column priorities + filter selects.

### Step A: Read the existing file

Identify:
1. The `useFooQuery(...)` hook call shape (filters / pagination / response shape).
2. The `columns: ColumnDef<Foo>[]` array (column id + header + cell render).
3. Any inline `<Card>` filter strip with search + select(s).
4. The action menu component mounted in the last column.
5. Pagination component (likely `CustomPagination` already).

### Step B: Replace TanStack DataTable / useReactTable with DataList

Replace this:
```tsx
import { DataTable } from "@/components/ui/data-table";
// or
import { useReactTable, getCoreRowModel, ... } from "@tanstack/react-table";

<DataTable columns={columns} data={data} ... />
```

With:
```tsx
import { DataList, DataListColumn } from "@/components/ui/data-list";

const columns: DataListColumn<Foo>[] = [
  { id: "name", header: "Name", cell: (row) => row.name },
  { id: "status", header: "Status", cell: (row) => <StatusBadge status={row.status} type="document" /> },
  // priority: "md" → hidden on <md, visible md+
  // priority: "lg" → hidden on <lg, visible lg+
  // align: "right" → text-right (desktop only)
  { id: "amount", header: "Total", priority: "md", align: "right", cell: (row) => formatCurrency(row.total) },
  { id: "actions", header: <span className="sr-only">Actions</span>, cell: (row) => <FooRowActions row={row} /> },
];

<DataList<Foo>
  rows={rows}
  columns={columns}
  getRowId={(r) => r.id}
  isLoading={isLoading}
  emptyMessage="No foos found."
  mobileCard={(row) => /* two-row card layout */}
/>
```

### Step C: Wrap search + filters in `<FilterBar>`

Replace inline `<Card>` filter strip (or zero-filter-UI) with:

```tsx
<FilterBar
  search={
    <div className="relative">
      <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
      <Input
        placeholder="Search foos…"
        value={searchQuery}
        onChange={(e) => setSearchQuery(e.target.value)}
        className="pl-10"
      />
    </div>
  }
  filters={
    <>
      <Select value={statusFilter} onValueChange={setStatusFilter}>
        <SelectTrigger className="w-35">
          <SelectValue placeholder="Status" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="all">All statuses</SelectItem>
          {/* per-route options */}
        </SelectContent>
      </Select>
      {/* additional Selects per route */}
    </>
  }
  hasActiveFilters={hasActiveFilters}
  onReset={clearFilters}
  meta={`${filtered.length} foo${filtered.length === 1 ? "" : "s"}${hasActiveFilters ? " (filtered)" : ""}`}
/>
```

Use `useDebounce` (`@/hooks/use-debounce`) on `searchQuery` if filtering server-side; client-side filtering on a paginated fetch is acceptable as a fallback.

### Step D: Mobile card layout

Two-row layout. First row = doc number + status pill (heaviest visual). Second row = wrapped meta (priority, date, amount). Action row below if applicable.

```tsx
mobileCard={(row) => (
  <div className="flex flex-col gap-2">
    <div className="flex items-start justify-between gap-2">
      <div className="min-w-0">
        <div className="font-medium text-primary line-clamp-1">{row.documentNumber}</div>
        <div className="text-xs text-muted-foreground line-clamp-1">{row.title}</div>
      </div>
      <StatusBadge status={row.status} type="document" />
    </div>
    <div className="flex flex-wrap items-center gap-2 text-xs text-muted-foreground">
      {/* per-route meta — vendor, amount, date, etc */}
    </div>
    <div className="pt-1">
      <FooRowActions row={row} variant="compact" />
    </div>
  </div>
)}
```

If existing row-action component does not have a "compact" variant, render it as-is — it'll just appear at desktop size on mobile cards which is acceptable.

### Step E: Verify TS + commit

Run `cd frontend && pnpm tsc --noEmit`. Clean → commit with the `refactor(<route>): <table> uses DataList + FilterBar` message.

---

## Task 1: DataList JSDoc — `align` is desktop-only (Plan D carry-forward)

**Files:**
- Modify: `frontend/src/components/ui/data-list.tsx`

- [ ] **Step 1: Update JSDoc**

Open `frontend/src/components/ui/data-list.tsx`. Find the `DataListColumn<T>` interface and the existing comment for `align`:

```tsx
  /** Text alignment for header and cell. Default 'left'. */
  align?: "left" | "right" | "center";
```

Replace with:

```tsx
  /**
   * Text alignment for header and cell on the desktop table only.
   * The mobile card layout is owned by the `mobileCard` render prop and
   * ignores `align`. Default `"left"`.
   */
  align?: "left" | "right" | "center";
```

- [ ] **Step 2: Verify TS compiles**

Run: `cd frontend && pnpm tsc --noEmit`
Expected: clean.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/ui/data-list.tsx
git commit -m "docs(ui): clarify DataList align is desktop-only"
```

---

## Task 2: Migrate `requisitions-table.tsx`

**Files:**
- Modify: `frontend/src/app/(private)/(main)/requisitions/_components/requisitions-table.tsx`

Largest table at 494 LOC. Existing inline filter card + active-filter chip badges with raw `bg-blue-100/bg-orange-100/bg-purple-100` color classes.

- [ ] **Step 1: Read the existing file** (per migration pattern Step A above)

- [ ] **Step 2: Apply migration pattern Steps B–D**

Per-route specifics for requisitions:

**Filters needed in `<FilterBar>`:**
- Search (text)
- Status select: `all` / `draft` / `submitted` / `in_approval` / `approved` / `rejected`
- Department select (if `useDepartments()` hook is already imported in this file or the `requisitions-client`)
- Priority select: `all` / `urgent` / `high` / `medium` / `low`

If a department select is hard to wire because the data lives outside the table, skip it and document as a follow-up — Plan E priority is the table+filter shell, not new data joins.

**Columns (in order):**
1. `requisitionNumber` — header "Document", priority `always`, cell renders bold link to detail page
2. `title` — header "Title", priority `lg`, cell renders truncated title
3. `requesterName` — header "Requester", priority `md`, cell renders name + role
4. `priority` — header "Priority", priority `md`, cell renders `<PriorityBadge priority={row.priority} />`
5. `totalAmount` — header "Amount", priority `md`, align `right`, cell renders `formatCurrency(row.totalAmount)`
6. `status` — header "Status", priority `always`, cell renders `<StatusBadge status={row.status} type="document" />`
7. `createdAt` — header "Created", priority `lg`, cell renders `new Date(row.createdAt).toLocaleDateString()`
8. `actions` — header `<span className="sr-only">Actions</span>`, priority `always`, cell renders existing row-action menu

**Mobile card:**
- Row 1: requisition number + status badge
- Row 2: requester · amount · date
- Row 3: priority badge + action menu

**Replace raw active-filter chip badges** that currently use `bg-blue-100`, `bg-orange-100`, etc. — `<FilterBar.meta>` displays count + (filtered) text, no need for individual chip-per-filter visual.

**Pagination:** if currently fetched without pagination (per the audit, `useRequisitions(1, 100, ...)`), keep the current behavior — don't introduce new pagination semantics in this task.

- [ ] **Step 3: Verify TS compiles**

Run: `cd frontend && pnpm tsc --noEmit`
Expected: clean.

- [ ] **Step 4: Commit**

```bash
git add "frontend/src/app/(private)/(main)/requisitions/_components/requisitions-table.tsx"
git commit -m "refactor(requisitions): requisitions-table uses DataList + FilterBar"
```

---

## Task 3: Migrate `purchase-orders-table.tsx`

**Files:**
- Modify: `frontend/src/app/(private)/(main)/purchase-orders/_components/purchase-orders-table.tsx`

Currently NO filter UI. Add a basic FilterBar with search + status + (optional) vendor + (optional) date.

- [ ] **Step 1: Read the existing file** (per migration pattern Step A)

- [ ] **Step 2: Apply migration pattern Steps B–D**

Per-route specifics for purchase orders:

**Filters needed in `<FilterBar>`:**
- Search (text — debounced if filtering server-side)
- Status select: `all` / `draft` / `pending_approval` / `issued` / `closed` / `rejected`
- Vendor select (only if `useVendors()` is already imported elsewhere in this file's neighborhood; otherwise skip)

**Columns:**
1. `poNumber` — "PO #", priority `always`
2. `vendor` — "Vendor", priority `md`, cell renders `row.vendor.name`
3. `requisition` — "Linked Req", priority `lg`, cell renders linked req number or "—"
4. `total` — "Total", priority `md`, align `right`
5. `status` — "Status", priority `always`, cell renders `<StatusBadge type="document" />`
6. `issuedAt` — "Issued", priority `lg`, cell renders date or "—"
7. `actions` — existing row-action menu

**Mobile card:**
- Row 1: PO number + status
- Row 2: vendor · total · issued date

**Action menu:** existing menu component should be referenced from the cell renderer. If the file currently has a `console.log("Download PDF")` stub in the menu, leave it — Plan I cleans these up. Don't introduce new behavior.

- [ ] **Step 3: Verify TS compiles**

Run: `cd frontend && pnpm tsc --noEmit`
Expected: clean.

- [ ] **Step 4: Commit**

```bash
git add "frontend/src/app/(private)/(main)/purchase-orders/_components/purchase-orders-table.tsx"
git commit -m "refactor(po): purchase-orders-table uses DataList + FilterBar"
```

---

## Task 4: Migrate `payment-vouchers-table.tsx`

**Files:**
- Modify: `frontend/src/app/(private)/(main)/payment-vouchers/_components/payment-vouchers-table.tsx`

Currently NO filter UI. Add FilterBar with status + payment method + search.

- [ ] **Step 1: Read the existing file**

- [ ] **Step 2: Apply migration pattern Steps B–D**

Per-route specifics for payment vouchers:

**Filters in `<FilterBar>`:**
- Search (text)
- Status select: `all` / `draft` / `pending_approval` / `approved` / `paid` / `rejected`
- Payment method select: `all` / `bank_transfer` / `check` / `cash` / `mobile_money` (use whatever the type union actually is — check `frontend/src/types/payment-voucher.ts`)

**Columns:**
1. `pvNumber` — "PV #", priority `always`
2. `payee` — "Payee", priority `md`, cell renders `row.payeeName`
3. `linkedPo` — "Linked PO", priority `lg`, cell renders linked PO number or "—"
4. `amount` — "Amount", priority `md`, align `right`
5. `status` — "Status", priority `always`
6. `dueDate` — "Due", priority `lg`, cell renders date with overdue highlighting (red text + "Overdue" tag if `dueDate < now && status !== "paid"`)
7. `actions` — existing row-action menu

**Mobile card:**
- Row 1: PV number + status
- Row 2: payee · amount · due date (with overdue highlight)

- [ ] **Step 3: Verify TS compiles**

Run: `cd frontend && pnpm tsc --noEmit`
Expected: clean.

- [ ] **Step 4: Commit**

```bash
git add "frontend/src/app/(private)/(main)/payment-vouchers/_components/payment-vouchers-table.tsx"
git commit -m "refactor(pv): payment-vouchers-table uses DataList + FilterBar"
```

---

## Task 5: Migrate `grn-table.tsx`

**Files:**
- Modify: `frontend/src/app/(private)/(main)/grn/_components/grn-table.tsx`

Currently NO filter UI.

- [ ] **Step 1: Read the existing file**

- [ ] **Step 2: Apply migration pattern Steps B–D**

Per-route specifics for GRN:

**Filters in `<FilterBar>`:**
- Search (text)
- Status select: `all` / `received` / `confirmed` / `rejected`
- Linked PO filter (text input that filters by PO number client-side)

**Columns:**
1. `grnNumber` — "GRN #", priority `always`
2. `linkedPo` — "PO", priority `md`, cell renders linked PO number
3. `vendor` — "Vendor", priority `lg`, cell renders vendor name from PO
4. `receivedBy` — "Received by", priority `lg`
5. `receivedDate` — "Received", priority `md`, cell renders date
6. `status` — "Status", priority `always`
7. `actions` — existing row-action menu (note: "Download PDF" is currently a `console.log` stub — leave for Plan I)

**Mobile card:**
- Row 1: GRN number + status
- Row 2: PO · received-by · received-date

- [ ] **Step 3: Verify TS compiles**

Run: `cd frontend && pnpm tsc --noEmit`
Expected: clean.

- [ ] **Step 4: Commit**

```bash
git add "frontend/src/app/(private)/(main)/grn/_components/grn-table.tsx"
git commit -m "refactor(grn): grn-table uses DataList + FilterBar"
```

---

## Task 6: Migrate `budgets-table.tsx` (HIGH PRIORITY — only raw useReactTable)

**Files:**
- Modify: `frontend/src/app/(private)/(main)/budgets/_components/budgets-table.tsx`

This is the only table in the codebase still using raw `useReactTable` directly (instead of going through the project's `DataTable` wrapper). It also has a hardcoded `grid-cols-8` skeleton (`BudgetsTableSkeleton` defined inline) that won't adapt if columns change. Plan E moves this onto `DataList` like every other route.

- [ ] **Step 1: Read the existing file**

In addition to the standard read, identify:
- The inline `BudgetsTableSkeleton` component defined within this file (lines 46–82 per the audit). This will be deleted — `DataList` provides skeleton rendering when `isLoading={true}`.
- All TanStack imports: `ColumnDef`, `useReactTable`, `getCoreRowModel`, `getSortedRowModel`, `getPaginationRowModel`, `getFilteredRowModel`, `flexRender`, `SortingState`, `ColumnFiltersState`, `VisibilityState`. These all go away.

- [ ] **Step 2: Apply migration pattern Steps B–D**

Per-route specifics for budgets:

**Filters in `<FilterBar>`:**
- Search (text)
- Fiscal year select: dynamic from `data.map((b) => b.fiscalYear)` distinct, sorted desc, plus `all`
- Department select: dynamic from `data.map((b) => b.department?.name)` if available, plus `all`
- Status select: `all` / `draft` / `submitted` / `approved` / `closed` / `rejected`

**Columns:**
1. `name` — "Budget", priority `always`
2. `department` — "Department", priority `md`
3. `fiscalYear` — "FY", priority `md`, align `right`
4. `total` — "Total", priority `md`, align `right`, cell renders `formatCurrency(row.totalBudget)`
5. `utilization` — "Used", priority `lg`, align `right`, cell renders `${utilizationPct}%` or a small `<Progress>` bar (use `Progress` from `ui/progress.tsx`)
6. `status` — "Status", priority `always`
7. `actions` — existing row-action menu

**Mobile card:**
- Row 1: budget name + status
- Row 2: department · FY · total
- Row 3: utilization progress bar (full width)

**Delete the inline `BudgetsTableSkeleton`** — `<DataList isLoading>` handles loading state.

**Drop all TanStack imports** that are no longer used.

- [ ] **Step 3: Verify TS compiles**

Run: `cd frontend && pnpm tsc --noEmit`
Expected: clean. If `Progress` is missing from `ui/progress.tsx`, fall back to a simple `<div className="h-1.5 w-full rounded-full bg-muted overflow-hidden"><div className="h-full bg-primary" style={{ width: `${pct}%` }} /></div>`.

- [ ] **Step 4: Commit**

```bash
git add "frontend/src/app/(private)/(main)/budgets/_components/budgets-table.tsx"
git commit -m "refactor(budgets): budgets-table uses DataList (drop raw useReactTable)"
```

---

## Task 7: Migrate `vendors-table.tsx`

**Files:**
- Modify: `frontend/src/app/(private)/(main)/vendors/_components/vendors-table.tsx`

7-column table with no mobile card and no pagination (audit noted `useVendors()` fetches all).

- [ ] **Step 1: Read the existing file**

Identify the existing column definitions (Code, Name, Email, Phone, Category, Status, Actions per the audit).

Check if `useVendors()` accepts pagination params. Open `frontend/src/hooks/use-vendor-queries.ts` (or wherever the hook lives) and confirm. If it does, wire pagination. If it doesn't, skip pagination and note it for Plan I — Plan E priority is the rendering layer.

- [ ] **Step 2: Apply migration pattern Steps B–D**

Per-route specifics for vendors:

**Filters in `<FilterBar>`:**
- Search (text — searches name + code + email)
- Category select: dynamic from data, plus `all`
- Status select: `all` / `active` / `inactive` (or whatever the type union is — check `frontend/src/types/vendor.ts`)

**Columns:**
1. `code` — "Code", priority `always`
2. `name` — "Name", priority `always`
3. `email` — "Email", priority `lg`
4. `phone` — "Phone", priority `lg`
5. `category` — "Category", priority `md`, cell renders `<Badge variant="outline">{row.category}</Badge>`
6. `status` — "Status", priority `always`, cell renders `<StatusBadge status={row.status} type="document" />` or `type="health"` depending on what fits — check `lib/status-badges.ts` for which type accepts `active`/`inactive`. If unclear, default to `<Badge variant={row.status === "active" ? "default" : "secondary"}>{row.status}</Badge>`.
7. `actions` — existing edit/delete menu (currently opens `vendor-form-sheet` for edit)

**Mobile card:**
- Row 1: vendor name + status badge
- Row 2: code · category
- Row 3: email · phone (truncated)

- [ ] **Step 3: Verify TS compiles**

Run: `cd frontend && pnpm tsc --noEmit`
Expected: clean.

- [ ] **Step 4: Commit**

```bash
git add "frontend/src/app/(private)/(main)/vendors/_components/vendors-table.tsx"
git commit -m "refactor(vendors): vendors-table uses DataList + FilterBar"
```

---

## Task 8: Verification pass

- [ ] **Step 1: Full type-check**

Run: `cd frontend && pnpm tsc --noEmit`
Expected: clean.

- [ ] **Step 2: Plan A+B+C+D regression**

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
  src/__tests__/components/base/empty-state.test.tsx \
  src/__tests__/components/admin/reports-header.test.tsx \
  src/__tests__/integration/admin-reports.test.tsx \
  src/__tests__/hooks/use-date-range-url-state.test.ts
```
Expected: all pass.

- [ ] **Step 3: Manual mobile smoke (optional)**

If able, `cd frontend && pnpm dev` and on mobile viewport (390px wide):
- `/requisitions` → table → mobile card layout, filter wraps
- `/purchase-orders` → same
- `/payment-vouchers` → same
- `/grn` → same
- `/budgets` → same (verify no raw TanStack remnants)
- `/vendors` → same

Verify desktop ≥1024px: tables render full width with column priorities respected (no awkward overflow).

- [ ] **Step 4: Cleanup commit if needed**

```bash
git add -A
git commit -m "chore(ui): plan-E verification cleanup"
```

---

## Self-Review Notes

- **Spec coverage:** Plan D carry-forward (Task 1: DataList align JSDoc) lands. Six tables (Tasks 2–7) all migrate to DataList + FilterBar with mobile cards. Verification (Task 8). All from the consolidated audit.
- **Type consistency:** `DataListColumn<T>` shape used identically across Tasks 2–7. `align: "right"` reserved for numeric columns (amount, year, utilization). `priority: "md" | "lg"` chosen consistently — `always` for doc#/status/actions; `md` for medium-importance fields; `lg` for least-important.
- **No placeholders:** Per-task notes specify EXACT column ids, headers, priorities, alignment, mobile-card layout, and filter selects. The migration pattern is documented once, with each task pointing back to it for shared steps and adding only route-specific notes.
- **Out-of-scope discipline:** No detail-page changes, no modal changes (those are Plan F+G), no new pagination wiring beyond what existing hooks support, no data-shape changes, no action-stub removal (`console.log` stays — Plan I).

## Plan F carry-forward notes

- `requisitions-table` department filter: deferred until department data is plumbed properly.
- `purchase-orders-table` vendor filter: same — deferred to Plan F.
- `vendors-table` server pagination: deferred until `useVendors()` is paginated.
- "Download PDF" / "Delete" `console.log` stubs across PO/GRN/requisitions row-action menus: Plan I.
- `budgets-table`: Progress component fallback used if `ui/progress.tsx` is missing — confirm during Task 6 and standardize.
