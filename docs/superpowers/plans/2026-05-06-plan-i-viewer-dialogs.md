# UI Revamp — Plan I: Viewer Dialogs Full-Screen

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a `fullScreen` prop to `<ResponsiveSheet>`, then migrate `pdf-preview-dialog` and `attachment-preview-dialog` to consume it. Viewers need maximum screen real estate, especially on mobile where the standard 90svh constraint cuts off PDF pages and image previews.

**Architecture:** `<ResponsiveSheet>` (Plan B Task 5, hardened in Plan D) currently caps at `max-h-[90svh]` on both branches. Plan I adds an opt-in `fullScreen` prop: desktop branch uses `max-w-[95vw] max-h-[95vh]` instead of the configured `desktopMaxWidth`; mobile branch uses `inset-0` (top + bottom + sides) so the drawer fills the viewport. Body padding stays consistent. Both viewer dialogs (read-only PDF + attachment image) are migrated as proof. Two existing toolbars (zoom, page nav, download) survive intact — they live inside `children`. Footer stays sticky via the existing Plan D fix.

**Tech Stack:** Next.js App Router, TypeScript, Tailwind v4, ShadCN UI, `vaul` 1.x, react-pdf 9.x, Vitest.

**Out-of-scope (future plans):**
- Confirmation dialogs migration (`*-submit-dialog`, `budget-delete-dialog`) — substantive forms with workflow selection, not pure confirms; defer
- Detail-page monolith splits (PO 1152, req 920, PV 733, GRN 596, budget 457 LOC)
- `approval-history-panel.tsx` 1271 LOC split
- Pre-existing test failures (submit-dialog, vendor-form-validation property test)
- PO `total` vs `totalAmount` field unification
- Server-side vendor pagination

---

## Existing infrastructure

- `<ResponsiveSheet>` at `frontend/src/components/ui/responsive-sheet.tsx` (Plan B + D)
  - Mobile: `<Vaul.Content>` `fixed bottom-0 left-0 right-0 mt-24 max-h-[90svh] flex flex-col`
  - Desktop: `<DialogContent>` with `cn(desktopMaxWidth, "flex flex-col max-h-[90svh] p-0 overflow-hidden gap-0")`
- `pdf-preview-dialog.tsx` (275 LOC) — uses Dialog with `h-[95vh] max-h-[95vh] min-w-5xl max-w-5xl gap-0 p-0`
- `attachment-preview-dialog.tsx` (332 LOC) — uses Dialog with `h-[90vh] max-h-[90vh] min-w-5xl max-w-5xl gap-0 p-0`

---

## File Structure

**Modify:**
- `frontend/src/components/ui/responsive-sheet.tsx` — add `fullScreen?: boolean` prop, gate width/height/inset/rounded classes accordingly
- `frontend/src/components/modals/pdf-preview-dialog.tsx` — replace Dialog with `<ResponsiveSheet fullScreen>`
- `frontend/src/components/modals/attachment-preview-dialog.tsx` — replace Dialog with `<ResponsiveSheet fullScreen>`

**No tests added.** ResponsiveSheet has implicit coverage via Plan D modal smoke tests; viewer dialogs are pure composition.

---

## Task 1: Add `fullScreen` prop to `<ResponsiveSheet>`

**Files:**
- Modify: `frontend/src/components/ui/responsive-sheet.tsx`

- [ ] **Step 1: Replace the file**

Open `frontend/src/components/ui/responsive-sheet.tsx`. Replace the entire body with:

```tsx
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
  /** Tailwind max-w on desktop. Ignored when `fullScreen` is true. */
  desktopMaxWidth?: string;
  /**
   * Fill the viewport. Desktop: 95vw × 95vh. Mobile: inset-0 (full-screen drawer
   * with top sheet edge instead of a partial pull-up). Use for PDF/image
   * viewers and other read-only content that needs maximum screen area.
   * Default: false.
   */
  fullScreen?: boolean;
  /**
   * When false, clicking the backdrop / pressing Escape / dragging the sheet
   * down does NOT dismiss. Use for forms with unsaved state.
   * Default: true.
   */
  dismissibleOnOutsideClick?: boolean;
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
  fullScreen = false,
  dismissibleOnOutsideClick = true,
  className,
}: ResponsiveSheetProps) {
  const isMobile = useIsMobile();

  if (isMobile) {
    return (
      <Vaul.Root
        open={open}
        onOpenChange={onOpenChange}
        dismissible={dismissibleOnOutsideClick}
      >
        <Vaul.Portal>
          <Vaul.Overlay className="fixed inset-0 bg-black/40 z-50" />
          <Vaul.Content
            className={cn(
              "fixed left-0 right-0 z-50 flex flex-col bg-background border-t",
              fullScreen
                ? "inset-0 rounded-none"
                : "bottom-0 mt-24 max-h-[90svh] rounded-t-xl",
              className
            )}
          >
            {!fullScreen && (
              <div className="mx-auto mt-2 h-1.5 w-12 shrink-0 rounded-full bg-muted" />
            )}
            {(title || description) && (
              <div
                className={cn(
                  "px-4 pb-2 space-y-1 shrink-0",
                  fullScreen ? "pt-4" : "pt-3"
                )}
              >
                {title && (
                  <Vaul.Title className="text-base font-semibold">
                    {title}
                  </Vaul.Title>
                )}
                {description && (
                  <Vaul.Description asChild>
                    <div className="text-sm text-muted-foreground">
                      {description}
                    </div>
                  </Vaul.Description>
                )}
              </div>
            )}
            <div className="flex-1 overflow-y-auto px-4 pb-4 min-h-0">
              {children}
            </div>
            {footer && (
              <div className="shrink-0 border-t bg-background p-3 pb-[max(0.75rem,env(safe-area-inset-bottom))]">
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
        className={cn(
          fullScreen
            ? "max-w-[95vw] sm:max-w-[95vw] max-h-[95vh] w-[95vw]"
            : cn(desktopMaxWidth, "max-h-[90svh]"),
          "flex flex-col p-0 overflow-hidden gap-0",
          className
        )}
        onInteractOutside={
          dismissibleOnOutsideClick ? undefined : (e) => e.preventDefault()
        }
        onEscapeKeyDown={
          dismissibleOnOutsideClick ? undefined : (e) => e.preventDefault()
        }
      >
        {(title || description) && (
          <DialogHeader className="px-6 pt-5 pb-2 shrink-0">
            {title && <DialogTitle>{title}</DialogTitle>}
            {description && (
              <DialogDescription asChild>
                <div className="text-sm text-muted-foreground">
                  {description}
                </div>
              </DialogDescription>
            )}
          </DialogHeader>
        )}
        <div className="flex-1 overflow-y-auto px-6 pb-4 min-h-0">
          {children}
        </div>
        {footer && (
          <DialogFooter className="shrink-0 px-6 py-4 border-t bg-background">
            {footer}
          </DialogFooter>
        )}
      </DialogContent>
    </Dialog>
  );
}
```

Key additions:
- New `fullScreen?: boolean` prop, default `false`. Documented in JSDoc.
- Mobile: when `fullScreen`, replaces `bottom-0 mt-24 max-h-[90svh] rounded-t-xl` with `inset-0 rounded-none`. Drag handle hidden. Header `pt-3 → pt-4`.
- Desktop: when `fullScreen`, ignores `desktopMaxWidth` and applies `max-w-[95vw] sm:max-w-[95vw] max-h-[95vh] w-[95vw]`.

- [ ] **Step 2: Verify TS compiles**

Run from `frontend/`: `pnpm tsc --noEmit`
Expected: clean.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/ui/responsive-sheet.tsx
git commit -m "feat(ui): ResponsiveSheet adds fullScreen prop"
```

---

## Task 2: Migrate `pdf-preview-dialog.tsx` to ResponsiveSheet fullScreen

**Files:**
- Modify: `frontend/src/components/modals/pdf-preview-dialog.tsx`

The current dialog uses `<DialogContent className="h-[95vh] max-h-[95vh] min-w-5xl max-w-5xl gap-0 p-0">`. Migrate to `<ResponsiveSheet fullScreen>`. Existing toolbar (zoom, page nav, download) lives inside the body or footer.

- [ ] **Step 1: Read the file**

Read `frontend/src/components/modals/pdf-preview-dialog.tsx` end-to-end (~275 LOC). Identify:
- The `<Dialog>...<DialogContent>` shell
- `<DialogHeader>` / `<DialogTitle>` (likely shows the file name)
- `<DialogFooter>` (likely has zoom controls + page nav + download button) OR the toolbar is inline in the body
- The PDF viewport (`<Document>` from react-pdf with `<Page>`)

- [ ] **Step 2: Apply migration pattern**

Replace the imports:
```tsx
// Drop:
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";

// Add:
import { ResponsiveSheet } from "@/components/ui/responsive-sheet";
```

Replace the outer `<Dialog open={open} onOpenChange={onOpenChange}><DialogContent>...</DialogContent></Dialog>` with:

```tsx
<ResponsiveSheet
  open={open}
  onOpenChange={onOpenChange}
  fullScreen
  title={fileName}
  footer={
    <div className="flex flex-wrap items-center justify-between gap-2 w-full">
      {/* Existing toolbar: zoom controls + page nav + download button */}
    </div>
  }
>
  {/* PDF viewport content */}
</ResponsiveSheet>
```

Move the `<DialogTitle>` content into the `title` prop. Move the `<DialogFooter>` content (zoom + page nav + download) into the `footer` prop.

If the file uses a custom header layout (e.g., file name + close X button on the right), pass JSX to `title`:
```tsx
title={
  <span className="flex items-center justify-between gap-2 pr-8">
    <span className="truncate">{fileName}</span>
  </span>
}
```
(The Dialog primitive automatically renders a close X — Vaul does not. The standard ResponsiveSheet drag handle is hidden in fullScreen mode. Add a close button in the footer or as part of the title row if accessibility requires it.)

Drop the inline `h-[95vh] max-h-[95vh] min-w-5xl max-w-5xl gap-0 p-0` classes — `fullScreen` handles the sizing. Drop the duplicate close X if `Dialog` was rendering one.

If the body needs to maintain an explicit height (the PDF `<Page>` may need a parent with `h-full`), ResponsiveSheet's body wrapper already has `flex-1 overflow-y-auto`. Pass `className="h-full"` on the inner PDF container if needed.

- [ ] **Step 3: Verify TS compiles**

Run from `frontend/`: `pnpm tsc --noEmit`
Expected: clean.

- [ ] **Step 4: Smoke check (optional)**

If able, run `cd frontend && pnpm dev`. Navigate to any document detail page (e.g. `/purchase-orders/<id>`) and trigger PDF preview. On mobile (390px wide), drawer should fill the viewport. Page nav, zoom, and download should all still work.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/modals/pdf-preview-dialog.tsx
git commit -m "refactor(modals): pdf-preview-dialog uses ResponsiveSheet fullScreen"
```

---

## Task 3: Migrate `attachment-preview-dialog.tsx` to ResponsiveSheet fullScreen

**Files:**
- Modify: `frontend/src/components/modals/attachment-preview-dialog.tsx`

The current dialog uses `<DialogContent className="h-[90vh] max-h-[90vh] min-w-5xl max-w-5xl gap-0 p-0">`. Same pattern as Task 2 — image viewer instead of PDF.

- [ ] **Step 1: Read the file**

Read `frontend/src/components/modals/attachment-preview-dialog.tsx` end-to-end (~332 LOC). Identify:
- The `<Dialog>...<DialogContent>` shell
- `<DialogHeader>` / `<DialogTitle>` (likely shows the attachment file name)
- `<DialogFooter>` or inline toolbar (zoom, rotate, download)
- The image / file rendering area

- [ ] **Step 2: Apply migration pattern**

Same shape as Task 2:
1. Drop dialog primitive imports.
2. Add `import { ResponsiveSheet } from "@/components/ui/responsive-sheet";`.
3. Replace outer Dialog shell with `<ResponsiveSheet fullScreen open={...} onOpenChange={...} title={...} footer={...}>`.
4. Move title text into `title` prop.
5. Move footer/toolbar content into `footer` prop wrapped in `<div className="flex flex-wrap items-center justify-between gap-2 w-full">`.
6. Drop the explicit `h-[90vh] max-h-[90vh] min-w-5xl max-w-5xl gap-0 p-0` classes.

If the body has an inline error state (e.g., `<div className="max-w-md rounded-lg bg-destructive/10 p-6 text-center">`), keep it as-is.

- [ ] **Step 3: Verify TS compiles**

Run from `frontend/`: `pnpm tsc --noEmit`
Expected: clean.

- [ ] **Step 4: Smoke check (optional)**

Trigger an attachment preview from any detail page. On mobile (390px wide), drawer should fill the viewport; image should fit within the body using `object-contain`.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/modals/attachment-preview-dialog.tsx
git commit -m "refactor(modals): attachment-preview-dialog uses ResponsiveSheet fullScreen"
```

---

## Task 4: Verification pass

- [ ] **Step 1: Full type-check**

Run: `cd frontend && pnpm tsc --noEmit`
Expected: clean.

- [ ] **Step 2: Plan A–H regression**

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
Expected: all known-good tests pass. Pre-existing submit-dialog + vendor-form-validation failures stay as-is (not Plan I scope).

- [ ] **Step 3: Manual mobile smoke (optional)**

If able, `cd frontend && pnpm dev`. On mobile viewport (390px wide):
- Open a PDF preview → drawer fills viewport (no partial top gap)
- Open an attachment preview → image renders within viewport bounds, doesn't overflow
- Tap outside → dismisses (default `dismissibleOnOutsideClick` is true; viewers don't have unsaved state)

Verify desktop ≥1024px:
- PDF preview at 95vw × 95vh — large but not edge-to-edge
- Attachment preview same

- [ ] **Step 4: Cleanup commit if needed**

```bash
git add -A
git commit -m "chore(ui): plan-I verification cleanup"
```

---

## Self-Review Notes

- **Spec coverage:**
  - `fullScreen` prop on ResponsiveSheet → Task 1 ✓
  - PDF viewer migration → Task 2 ✓
  - Attachment viewer migration → Task 3 ✓
  - Verification → Task 4 ✓
- **Type consistency:** `fullScreen?: boolean` default `false` consistent across the prop, the JSDoc, and the consumer invocations in Tasks 2+3 (both pass `fullScreen` as a flag prop, not a string value).
- **No placeholders:** Each task has explicit code or "read file + apply pattern" with the pattern fully documented. The two viewer migrations are similar enough that Task 3's body refers back to Task 2 for the shared steps — but each task contains its own complete spec.
- **Out-of-scope discipline:** Submit dialogs (substantive forms, defer), detail-page splits (massive, future plans), pre-existing test failures (not regressions caused here), field-name unification all explicitly NOT in Plan I.

## Future plan carry-forwards

- **Submit dialogs migration**: 6 files (`*-submit-dialog.tsx`) currently use Dialog. They contain workflow selection + warnings + comments. Migrate to `<ResponsiveSheet>` (NOT fullScreen — they're forms, not viewers). Probably one plan per file or 2 batched.
- **Budget delete dialog**: small AlertDialog confirmation. Either leave on AlertDialog with mobile-friendly sizing or build a `<ResponsiveConfirm>` primitive over AlertDialog.
- **Detail-page monolith splits**: PO 1152, req 920, PV 733, GRN 596, budget 457 LOC. Each route → its own plan applying `<DetailShell>` from Plan B + extracting sub-panels.
- **`approval-history-panel.tsx` 1271 LOC split**: separate plan. Split into `ApprovalChainSummary` + `ActivityFeed` + `CommentsThread`. Use `<ApprovalChainStepper>` from Plan F.
- **PO `total` vs `totalAmount`**: backend + frontend type unification. Touches every PO consumer.
- **Submit-dialog test failures**: 5 pre-existing. Tests assume Workflow Selection structure that may have changed. Update tests, not the dialog.
