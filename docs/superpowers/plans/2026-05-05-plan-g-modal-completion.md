# UI Revamp — Plan G: Modal Completion + Plan E Test Follow-up

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Migrate the 4 remaining major form modals to `<ResponsiveSheet>`, update Plan E navigation tests for the new DropdownMenu pattern, standardize the row-action trigger across all 6 procurement tables, and resolve the dead `/reports` nav link.

**Architecture:** Plan D migrated 7 modals; Plan G covers the remaining four (`approval-action-modal`, `edit-purchase-order-dialog`, `create-budget-dialog`, `create-po-from-requisition-dialog`) using the same Dialog→ResponsiveSheet pattern. The approval-action modal embeds a signature canvas — verify the canvas still works inside a Vaul drawer on mobile (touch events, contentRect for `getBoundingClientRect` scaling). Test follow-up: 4 navigation integration tests broke when Plan E consolidated row actions into DropdownMenu — update them to navigate trigger → menuitem. Trigger standardization: requisitions row uses default `<Button variant="outline">` (h-9), others use `size="icon" h-8 w-8` — pick one. Dead-link decision: `MobileBottomNav.MORE_LINKS` references `/reports` (a non-admin route that does not exist); drop the link; admins still reach `/admin/reports` via the existing nav-main entry.

**Tech Stack:** Next.js App Router, TypeScript, Tailwind v4, ShadCN UI, `<ResponsiveSheet>` (Plan B + D), `vaul`, Vitest.

**Out-of-scope (Plan H+):**
- Detail-page monolith splits (PO 1152, req 920, PV 733, GRN 596, budget 457 LOC) — Plan H
- `approval-history-panel.tsx` (1271 LOC) split — Plan H
- Confirmation-only dialogs (submit/delete/mark-paid) — small enough as-is
- PDF/attachment viewer dialogs — need full-screen sheet treatment, separate Plan
- Vendor detail page (`/vendors/[id]`) — Plan H

---

## File Structure

**Modify (modal migrations — 4 files):**
- `frontend/src/components/workflows/approval-action-modal.tsx` (372 LOC) — Dialog → ResponsiveSheet; verify signature canvas behavior in drawer
- `frontend/src/app/(private)/(main)/purchase-orders/_components/edit-purchase-order-dialog.tsx` (368 LOC)
- `frontend/src/app/(private)/(main)/budgets/_components/create-budget-dialog.tsx` (313 LOC)
- `frontend/src/app/(private)/(main)/purchase-orders/_components/create-po-from-requisition-dialog.tsx` (398 LOC)

**Modify (test updates — 4 files):**
- `frontend/src/__tests__/integration/purchase-orders/navigation.test.tsx` — navigate via DropdownMenu trigger → menuitem
- `frontend/src/__tests__/integration/payment-vouchers/navigation.test.tsx`
- `frontend/src/__tests__/integration/grn/navigation.test.tsx`
- `frontend/src/__tests__/integration/requisitions/navigation.test.tsx`

**Modify (cleanup):**
- `frontend/src/app/(private)/(main)/requisitions/_components/requisitions-table.tsx` — wrap row dropdown trigger in `<Button variant="outline" size="icon" className="h-8 w-8">` to match the other 5 tables
- `frontend/src/components/layout/mobile-bottom-nav.tsx` — drop `/reports` from `MORE_LINKS`

---

## Migration pattern (used by Tasks 1–4)

Same pattern as Plan D Tasks 5–10. Centralized once here.

### Steps per modal

1. Read the existing file end-to-end. Identify outer `<Dialog>...<DialogContent>...</DialogContent></Dialog>` shell, `DialogTitle`, `DialogDescription`, `DialogFooter`, current `sm:max-w-*` width, and form/submit structure.
2. Add `import { ResponsiveSheet } from "@/components/ui/responsive-sheet";`. Drop the dialog primitive imports that are no longer used.
3. Replace shell with `<ResponsiveSheet open={...} onOpenChange={...} title="..." description="..." desktopMaxWidth="sm:max-w-*" footer={...}>`.
4. Move `DialogFooter` content into `footer` prop wrapped in `<div className="flex flex-wrap items-center justify-end gap-2">`.
5. If form has no `id`, add `id="<route>-form"` and `form="<route>-form"` to the submit button in the footer.
6. Drop `max-h-[90svh] overflow-y-auto p-0` from old `DialogContent` (ResponsiveSheet handles).
7. Preserve original `sm:max-w-*` as `desktopMaxWidth`. If unprefixed `max-w-Nxl!`, normalize to `sm:max-w-Nxl`.
8. Form modals with multi-step state OR data that would be lost on dismiss: pass `dismissibleOnOutsideClick={false}`.

---

## Task 1: Migrate `approval-action-modal.tsx` to ResponsiveSheet

**Files:**
- Modify: `frontend/src/components/workflows/approval-action-modal.tsx`

This modal has a signature canvas (`<DigitalSignaturePad />`). Per project memory, the canvas already uses `getBoundingClientRect()` scale factors so drawing works at any CSS width. Vaul drawer should be fine for touch input; verify if smoke-testing.

- [ ] **Step 1: Read the file**

Read end-to-end. Note the shell, the title/description, the embedded `<DigitalSignaturePad />` component, the radio group for "Approve" / "Reject" / "Return for revision" (per project memory), the comments textarea, and the action buttons.

- [ ] **Step 2: Apply migration pattern**

Per the pattern at top. Specifics:
1. Import `ResponsiveSheet`. Drop dialog primitive imports.
2. Replace outer `Dialog` + `DialogContent` with `ResponsiveSheet`.
3. `title` / `description`: pass strings (or JSX if mixed icon + text).
4. `footer`: action buttons wrapped in flex container.
5. `desktopMaxWidth`: keep current width (likely `sm:max-w-lg`).
6. Pass `dismissibleOnOutsideClick={false}` — losing entered comments + drawn signature is high-cost.
7. If body has internal padding wrappers (e.g. `<div className="p-6">`), drop that since ResponsiveSheet wraps in `px-6 pb-4` desktop / `px-4 pb-4` mobile.

- [ ] **Step 3: Verify TS compiles**

Run from `frontend/`: `pnpm tsc --noEmit`
Expected: clean.

- [ ] **Step 4: Smoke check (optional)**

If able, `cd frontend && pnpm dev`. Navigate to `/tasks?tab=approvals`, claim a task, click Approve/Reject. On mobile viewport (390px wide), verify:
- Drawer slides up from bottom
- Signature canvas accepts touch drawing (try with mouse drag in DevTools touch emulation)
- Comments textarea accepts input
- Approve/Reject buttons reachable

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/workflows/approval-action-modal.tsx
git commit -m "refactor(workflows): approval-action-modal uses ResponsiveSheet"
```

---

## Task 2: Migrate `edit-purchase-order-dialog.tsx` to ResponsiveSheet

**Files:**
- Modify: `frontend/src/app/(private)/(main)/purchase-orders/_components/edit-purchase-order-dialog.tsx`

- [ ] **Step 1: Read the file**

Read end-to-end. Identify the form fields (likely vendor, items, total, notes — read the file).

- [ ] **Step 2: Apply migration pattern**

Per the pattern. Form id: `id="edit-po-form"` if `<form>` exists.

- [ ] **Step 3: Verify TS compiles**

Run: `cd frontend && pnpm tsc --noEmit`
Expected: clean.

- [ ] **Step 4: Commit**

```bash
git add "frontend/src/app/(private)/(main)/purchase-orders/_components/edit-purchase-order-dialog.tsx"
git commit -m "refactor(po): edit-purchase-order-dialog uses ResponsiveSheet"
```

---

## Task 3: Migrate `create-budget-dialog.tsx` to ResponsiveSheet

**Files:**
- Modify: `frontend/src/app/(private)/(main)/budgets/_components/create-budget-dialog.tsx`

- [ ] **Step 1: Read the file**

Read end-to-end.

- [ ] **Step 2: Apply migration pattern**

Per the pattern. Form id: `id="create-budget-form"`. Pass `dismissibleOnOutsideClick={false}` (multi-field form, easy to lose state).

- [ ] **Step 3: Verify TS compiles**

Run: `cd frontend && pnpm tsc --noEmit`
Expected: clean.

- [ ] **Step 4: Commit**

```bash
git add "frontend/src/app/(private)/(main)/budgets/_components/create-budget-dialog.tsx"
git commit -m "refactor(budgets): create-budget-dialog uses ResponsiveSheet"
```

---

## Task 4: Migrate `create-po-from-requisition-dialog.tsx` to ResponsiveSheet

**Files:**
- Modify: `frontend/src/app/(private)/(main)/purchase-orders/_components/create-po-from-requisition-dialog.tsx`

- [ ] **Step 1: Read the file**

Read end-to-end. This dialog likely embeds a requisition selection table or summary.

- [ ] **Step 2: Apply migration pattern**

Per the pattern. Form id: `id="create-po-from-req-form"`. Pass `dismissibleOnOutsideClick={false}`.

If the dialog embeds a selection table with internal scroll, preserve the inner scroll wrapper (e.g. `max-h-[400px] overflow-auto` on the table container) — that's table-internal scrolling, not sheet-internal.

- [ ] **Step 3: Verify TS compiles**

Run: `cd frontend && pnpm tsc --noEmit`
Expected: clean.

- [ ] **Step 4: Commit**

```bash
git add "frontend/src/app/(private)/(main)/purchase-orders/_components/create-po-from-requisition-dialog.tsx"
git commit -m "refactor(po): create-po-from-requisition-dialog uses ResponsiveSheet"
```

---

## Task 5: Standardize row-action trigger size in requisitions-table

**Files:**
- Modify: `frontend/src/app/(private)/(main)/requisitions/_components/requisitions-table.tsx`

The other 5 procurement tables (PO/PV/GRN/budget/vendor) use `<Button variant="outline" size="icon" className="h-8 w-8">` for the row action menu trigger. Requisitions uses default size, causing visual misalignment in the actions column.

- [ ] **Step 1: Find the trigger**

Open the file. Locate the existing `<DropdownMenuTrigger asChild>` block in the row-action menu (likely inside `ReqOptionsMenu` or directly in the columns array). The trigger looks like:

```tsx
<DropdownMenuTrigger asChild>
  <Button variant="outline">
    <MoreHorizontal className="h-4 w-4" />
  </Button>
</DropdownMenuTrigger>
```

(or with `<MoreVertical />` instead of `<MoreHorizontal />`)

- [ ] **Step 2: Apply size standardization**

Replace the `<Button>` opening tag with:

```tsx
<Button variant="outline" size="icon" className="h-8 w-8">
```

Add an `aria-label` for the trigger if absent — this also helps integration tests find the button:

```tsx
<Button
  variant="outline"
  size="icon"
  className="h-8 w-8"
  aria-label="Row actions"
>
  <MoreHorizontal className="h-4 w-4" />
</Button>
```

If the trigger appears in multiple places (column cell + mobile card), update both.

- [ ] **Step 3: Verify TS compiles**

Run: `cd frontend && pnpm tsc --noEmit`
Expected: clean.

- [ ] **Step 4: Commit**

```bash
git add "frontend/src/app/(private)/(main)/requisitions/_components/requisitions-table.tsx"
git commit -m "refactor(requisitions): row-action trigger matches other tables"
```

---

## Task 6: Add aria-label to row-action triggers across remaining tables

**Files:**
- Modify: `frontend/src/app/(private)/(main)/purchase-orders/_components/purchase-orders-table.tsx`
- Modify: `frontend/src/app/(private)/(main)/payment-vouchers/_components/payment-vouchers-table.tsx`
- Modify: `frontend/src/app/(private)/(main)/grn/_components/grn-table.tsx`
- Modify: `frontend/src/app/(private)/(main)/budgets/_components/budgets-table.tsx`

Each table's existing icon-only `<Button size="icon" className="h-8 w-8">` trigger lacks `aria-label`. Add it. This makes the triggers reachable by `screen.getByRole("button", { name: /row actions/i })` in the navigation integration tests (Task 7+).

- [ ] **Step 1: Update each file**

In each of the 4 files, find the row-action `<DropdownMenuTrigger asChild>` block. The `<Button>` should look like:

```tsx
<Button variant="outline" size="icon" className="h-8 w-8">
  <MoreHorizontal className="h-4 w-4" />
</Button>
```

Add `aria-label="Row actions"`:

```tsx
<Button
  variant="outline"
  size="icon"
  className="h-8 w-8"
  aria-label="Row actions"
>
  <MoreHorizontal className="h-4 w-4" />
</Button>
```

Apply to all 4 files.

If a file's row-action component is extracted (e.g. `<PoOptionsMenu>`), edit the trigger inside that component and ensure the `aria-label` lands on the actual trigger button.

- [ ] **Step 2: Verify TS compiles**

Run: `cd frontend && pnpm tsc --noEmit`
Expected: clean.

- [ ] **Step 3: Commit**

```bash
git add \
  "frontend/src/app/(private)/(main)/purchase-orders/_components/purchase-orders-table.tsx" \
  "frontend/src/app/(private)/(main)/payment-vouchers/_components/payment-vouchers-table.tsx" \
  "frontend/src/app/(private)/(main)/grn/_components/grn-table.tsx" \
  "frontend/src/app/(private)/(main)/budgets/_components/budgets-table.tsx"
git commit -m "a11y(tables): row-action triggers gain aria-label"
```

(If row-action triggers live inside extracted menu components, adjust the file list to include those component files instead.)

---

## Task 7: Update navigation integration tests for DropdownMenu pattern

**Files:**
- Modify: `frontend/src/__tests__/integration/purchase-orders/navigation.test.tsx`
- Modify: `frontend/src/__tests__/integration/payment-vouchers/navigation.test.tsx`
- Modify: `frontend/src/__tests__/integration/grn/navigation.test.tsx`
- Modify: `frontend/src/__tests__/integration/requisitions/navigation.test.tsx`

After Plan E, row actions live inside a DropdownMenu. Tests that did `getByRole("button", { name: /view/i })` or `findByRole("button", { name: /view details/i })` directly on the row no longer work — the View action is a `menuitem` reachable only after clicking the trigger.

- [ ] **Step 1: Pick one test as a template — `purchase-orders/navigation.test.tsx`**

Read the failing tests. Find every `findByRole("button", { name: /view/i })` (or similar action lookups: `/edit/i`, `/delete/i`, `/download/i`).

For each such lookup, replace the click step with a 2-step flow:

```tsx
// Before:
const viewButton = await screen.findByRole("button", { name: /view/i });
await user.click(viewButton);

// After:
const trigger = (await screen.findAllByRole("button", { name: /row actions/i }))[0];
await user.click(trigger);
const viewItem = await screen.findByRole("menuitem", { name: /view/i });
await user.click(viewItem);
```

(The first `(await findAllByRole(...))[0]` picks the first row's trigger when multiple rows render. Adjust to pick the row by content if the test asserts a specific document.)

- [ ] **Step 2: Apply the same pattern to the other 3 test files**

For `payment-vouchers/navigation.test.tsx`, `grn/navigation.test.tsx`, `requisitions/navigation.test.tsx`: locate every row-action `findByRole("button", ...)` or `getByRole("button", ...)` for action labels (View / Edit / Delete / Download). Replace with the trigger + menuitem 2-step flow.

If a test asserts mocked `useRouter().push` was called with a specific URL, that assertion stays — only the click-target lookup changes.

- [ ] **Step 3: Run all 4 tests**

Run: `cd frontend && pnpm vitest run src/__tests__/integration/purchase-orders/navigation.test.tsx src/__tests__/integration/payment-vouchers/navigation.test.tsx src/__tests__/integration/grn/navigation.test.tsx src/__tests__/integration/requisitions/navigation.test.tsx`
Expected: all pass. Report counts.

If a test fails for a reason unrelated to the click pattern (e.g. CardContent mock missing), apply the same fix from Plan E commit `71dec69`: add `CardContent`, `CardHeader`, `CardTitle`, `CardDescription`, `CardFooter` to any partial `vi.mock("@/components/ui/card", ...)`.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/__tests__/integration/
git commit -m "test(integration): navigation tests use DropdownMenu pattern"
```

---

## Task 8: Drop dead `/reports` link from MobileBottomNav

**Files:**
- Modify: `frontend/src/components/layout/mobile-bottom-nav.tsx`
- Modify: `frontend/src/__tests__/components/layout/mobile-bottom-nav.test.tsx` (if any test asserts /reports is in MORE_LINKS)

The non-admin `/reports` route does not exist in this codebase. Plan A's `MobileBottomNav.MORE_LINKS` references it as a placeholder. Admins reach `/admin/reports` via the existing nav-main entry; non-admins have nothing to see.

- [ ] **Step 1: Drop the link**

Open `frontend/src/components/layout/mobile-bottom-nav.tsx`. Find the `MORE_LINKS` array — it should contain a `{ href: "/reports", label: "Reports" }` entry. Delete that entry.

- [ ] **Step 2: Audit the test**

Open `frontend/src/__tests__/components/layout/mobile-bottom-nav.test.tsx`. If any test asserts `Reports` is in the More drawer, drop that assertion. Most likely the test only asserts the 4 primary tabs, so no change needed.

- [ ] **Step 3: Run the test**

Run: `cd frontend && pnpm vitest run src/__tests__/components/layout/mobile-bottom-nav.test.tsx`
Expected: pass.

- [ ] **Step 4: Verify TS compiles**

Run: `cd frontend && pnpm tsc --noEmit`
Expected: clean.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/layout/mobile-bottom-nav.tsx frontend/src/__tests__/components/layout/mobile-bottom-nav.test.tsx
git commit -m "chore(layout): drop dead /reports link from mobile More drawer"
```

---

## Task 9: Verification pass

- [ ] **Step 1: Full type-check**

Run: `cd frontend && pnpm tsc --noEmit`
Expected: clean.

- [ ] **Step 2: All known-good tests + integration tests**

Run:
```bash
cd frontend && pnpm vitest run \
  src/__tests__/components/ui/ \
  src/__tests__/components/workflows/ \
  src/__tests__/components/layout/ \
  src/__tests__/components/admin/ \
  src/__tests__/components/base/ \
  src/__tests__/hooks/ \
  src/__tests__/integration/admin-reports.test.tsx \
  src/__tests__/integration/purchase-orders/navigation.test.tsx \
  src/__tests__/integration/payment-vouchers/navigation.test.tsx \
  src/__tests__/integration/grn/navigation.test.tsx \
  src/__tests__/integration/requisitions/navigation.test.tsx
```
Expected: all pass.

- [ ] **Step 3: Manual mobile smoke (optional)**

If able, `cd frontend && pnpm dev`. On mobile viewport (390px wide):
- `/tasks?tab=approvals` — claim a task → Approve. Drawer slides up. Signature canvas works. Comments accept input.
- `/purchase-orders/<id>` — Edit PO → drawer
- `/budgets` — Create Budget → drawer
- `/requisitions/<id>` → Create PO from Requisition → drawer
- `/home` mobile bottom nav → tap More → no `/reports` entry

Verify desktop ≥1024px: dialogs render unchanged.

- [ ] **Step 4: Cleanup commit if needed**

```bash
git add -A
git commit -m "chore(ui): plan-G verification cleanup"
```

---

## Self-Review Notes

- **Spec coverage:** 4 modal migrations (Tasks 1-4). 4 nav-test updates (Task 7). Trigger size standardization (Tasks 5-6). Dead `/reports` link removal (Task 8). Verification (Task 9).
- **Type consistency:** All modal migrations use `desktopMaxWidth="sm:max-w-*"` convention from Plan D. Form id naming: `<route>-form` (e.g. `edit-po-form`, `create-budget-form`). Nav-test 2-step pattern (`row actions trigger → menuitem`) is consistent across all 4 test files.
- **No placeholders:** Each task has explicit code or a clear "read file + apply migration pattern (linked above)" pointer with the pattern fully documented.
- **Out-of-scope discipline:** Detail-page splits, approval-history-panel.tsx (1271 LOC), vendor detail page, PDF/attachment viewers, and confirmation dialogs are all explicitly NOT in Plan G. Each is named with rationale at top of plan.

## Plan H carry-forward notes

- **Detail-page monolith splits:** PO 1152, req 920, PV 733, GRN 596, budget 457 LOC. Apply `<DetailShell>`. Split into header/main/sidebar/items/approval/activity sub-components.
- **`approval-history-panel.tsx` (1271 LOC):** split into `ApprovalChainSummary` + `ActivityFeed` + `CommentsThread`; replace inline approval-chain rendering with `<ApprovalChainStepper>` (already adopted in Plan F's other 3 consumers).
- **Vendor detail page:** create `/vendors/[id]` with profile + KPIs (`StatGrid`) + recent POs (`DataList`).
- **Confirmation/viewer dialogs:** evaluate whether `ResponsiveConfirm` (over AlertDialog) or full-screen-sheet variants are needed. Plan I.
- **Action stub cleanup:** PO/GRN Download PDF, requisition Delete (`console.log`). Plan I.
- **Aria-label propagation:** if Tasks 5+6 expose row-action triggers via shared row-action menu components rather than inline, the patches may need to land in those component files instead. Implementer discovers during Task 5+6 execution.
