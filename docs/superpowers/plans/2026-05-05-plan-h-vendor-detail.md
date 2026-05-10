# UI Revamp — Plan H: Vendor Detail Page + Cleanups

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a `/vendors/[id]` detail page using `<DetailShell>` + `<StatGrid>` + `<DataList>` primitives, make `/vendors` table rows clickable to it, plus drop two stale action stubs and a Plan G width-normalization carry-forward.

**Architecture:** No detail page exists for vendors today — clicking a row opens an inline edit dialog. Plan H creates a proper `/vendors/[id]` route as the primitive showcase: server component delegates to `VendorDetailClient`, which wraps `<DetailShell>` (Plan B Task 7) with vendor profile + banking panels in the main slot and a `<StatGrid>` + recent-POs `<DataList>` in the sidebar. Click on a vendors-table row navigates to detail; inline edit panel goes away. Existing `useVendorById(id)` hook already returns the data; existing `usePurchaseOrders({ vendorId })` filter already works. No new hooks or backend work.

**Tech Stack:** Next.js App Router (server + client), TypeScript, Tailwind v4, ShadCN UI, `<DetailShell>` (Plan B), `<StatGrid>` (Plan A), `<DataList>` (Plan A), `<StatusBadge>` + `<DocumentTypeChip>`, Vitest.

**Out-of-scope (Plan I):**
- Detail-page monolith splits (PO 1152, req 920, PV 733, GRN 596, budget 457 LOC)
- `approval-history-panel.tsx` (1271 LOC) split
- Confirmation/viewer dialog migrations
- Server-side vendor pagination
- Vendor edit-from-detail-page UX (current edit dialog stays accessible from list table)

---

## Existing infrastructure (reuse)

- `<DetailShell>` at `frontend/src/components/layout/detail-shell.tsx` (Plan B Task 7)
- `<StatGrid>` at `frontend/src/components/ui/stat-grid.tsx` (Plan A)
- `<DataList>` at `frontend/src/components/ui/data-list.tsx` (Plan A + D)
- `<StatusBadge>` at `frontend/src/components/status-badge.tsx`
- `<DocumentTypeChip>` at `frontend/src/components/ui/document-type-chip.tsx` (Plan B Task 4)
- `<PageHeader>` at `frontend/src/components/base/page-header.tsx`
- `useVendorById(id)` at `frontend/src/hooks/use-vendor-queries.ts`
- `usePurchaseOrders(filters)` at `frontend/src/hooks/use-purchase-order-queries.ts` — accepts `{ vendorId }`
- `formatCurrency` at `frontend/src/lib/utils.ts`
- `Vendor` type at `frontend/src/types/core.ts` (re-exported from `frontend/src/types/vendor.ts`) — fields: `id`, `vendorCode`, `name`, `email`, `phone`, `contactPerson?`, `physicalAddress`, `city`, `country`, `taxId`, `bankName`, `branchCode?`, `accountName`, `accountNumber`, `swiftCode?`, `active`
- `PurchaseOrder.vendorId` field already wired

---

## File Structure

**Create:**
- `frontend/src/app/(private)/(main)/vendors/[id]/page.tsx` — server component, auth + delegate
- `frontend/src/app/(private)/(main)/vendors/[id]/_components/vendor-detail-client.tsx` — main client component using `<DetailShell>`
- `frontend/src/app/(private)/(main)/vendors/[id]/_components/vendor-profile-card.tsx` — contact + address card
- `frontend/src/app/(private)/(main)/vendors/[id]/_components/vendor-banking-card.tsx` — banking info card
- `frontend/src/app/(private)/(main)/vendors/[id]/_components/vendor-recent-pos.tsx` — DataList of POs filtered by vendorId

**Modify:**
- `frontend/src/app/(private)/(main)/vendors/_components/vendors-table.tsx` — add row click → navigate to detail; remove inline `VendorDetailPanel` if it became dead after navigation flips the trigger
- `frontend/src/app/(private)/(main)/purchase-orders/_components/create-po-from-requisition-dialog.tsx` — `desktopMaxWidth` from `max-w-5xl` to `sm:max-w-5xl` (Plan G carry-forward)
- `frontend/src/app/(private)/(main)/purchase-orders/_components/purchase-orders-table.tsx` — drop `console.log("Download PDF")` stub (or extracted `PoOptionsMenu` component if trigger lives there)
- `frontend/src/app/(private)/(main)/grn/_components/grn-table.tsx` — drop `console.log("Download PDF")` stub (or extracted `GrnOptionsMenu`)
- `frontend/src/app/(private)/(main)/requisitions/_components/requisitions-table.tsx` — drop `console.log` Delete stub (or extracted `ReqOptionsMenu`)

**No tests added.** New vendor detail components are pure composition over already-tested primitives.

---

## Task 1: Plan G carry-forward — normalize `max-w-5xl`

**Files:**
- Modify: `frontend/src/app/(private)/(main)/purchase-orders/_components/create-po-from-requisition-dialog.tsx`

- [ ] **Step 1: Update the prop**

Open the file. Find the `<ResponsiveSheet>` invocation (search for `desktopMaxWidth=`). Replace:

```tsx
desktopMaxWidth="max-w-5xl"
```

with:

```tsx
desktopMaxWidth="sm:max-w-5xl"
```

- [ ] **Step 2: Verify TS compiles**

Run from `frontend/`: `pnpm tsc --noEmit`
Expected: clean.

- [ ] **Step 3: Commit**

```bash
git add "frontend/src/app/(private)/(main)/purchase-orders/_components/create-po-from-requisition-dialog.tsx"
git commit -m "fix(po): create-po-from-requisition-dialog desktopMaxWidth uses sm: prefix"
```

---

## Task 2: Drop action-stub `console.log` calls

**Files:**
- Modify: `frontend/src/app/(private)/(main)/purchase-orders/_components/purchase-orders-table.tsx` (or extracted `PoOptionsMenu`)
- Modify: `frontend/src/app/(private)/(main)/grn/_components/grn-table.tsx` (or extracted `GrnOptionsMenu`)
- Modify: `frontend/src/app/(private)/(main)/requisitions/_components/requisitions-table.tsx` (or extracted `ReqOptionsMenu`)

For each file, find the `DropdownMenuItem` that calls `console.log("Download PDF")` or `console.log("Delete ...")`. Two options per stub:
1. Remove the menu item entirely.
2. Keep the menu item but mark it disabled with a tooltip saying "Coming soon".

Choose option 1 (remove) for cleanest UX — re-add later when the action lands.

- [ ] **Step 1: PO Download PDF stub**

Open `purchase-orders-table.tsx` (or `PoOptionsMenu` if extracted). Find:

```tsx
<DropdownMenuItem onClick={() => console.log("Download PDF", po.id)}>
  <Download className="h-4 w-4 mr-2" />
  Download PDF
</DropdownMenuItem>
```

(Exact match may vary — search for `console.log` in the file.)

Delete the entire `<DropdownMenuItem>...</DropdownMenuItem>` block. If `Download` is now an unused lucide-react import, also drop it.

- [ ] **Step 2: GRN Download PDF stub**

Open `grn-table.tsx` (or `GrnOptionsMenu`). Find the same pattern. Delete the menu item. Drop unused `Download` import.

- [ ] **Step 3: Requisition Delete stub**

Open `requisitions-table.tsx` (or `ReqOptionsMenu`). Find:

```tsx
<DropdownMenuItem onClick={() => console.log("Delete requisition", req.id)}>
  <Trash2 className="h-4 w-4 mr-2" />
  Delete
</DropdownMenuItem>
```

Delete the menu item. Drop unused `Trash2` import.

- [ ] **Step 4: Verify TS compiles**

Run from `frontend/`: `pnpm tsc --noEmit`
Expected: clean.

- [ ] **Step 5: Commit**

```bash
git add \
  "frontend/src/app/(private)/(main)/purchase-orders/_components/purchase-orders-table.tsx" \
  "frontend/src/app/(private)/(main)/grn/_components/grn-table.tsx" \
  "frontend/src/app/(private)/(main)/requisitions/_components/requisitions-table.tsx"
git commit -m "chore(tables): drop console.log action stubs (Download PDF, Delete)"
```

(Adjust the file list if the stubs lived in extracted menu components.)

---

## Task 3: Server-side vendor detail route entry

**Files:**
- Create: `frontend/src/app/(private)/(main)/vendors/[id]/page.tsx`

- [ ] **Step 1: Create page.tsx**

```tsx
// frontend/src/app/(private)/(main)/vendors/[id]/page.tsx
import { redirect } from "next/navigation";
import { verifySession } from "@/lib/auth";
import { VendorDetailClient } from "./_components/vendor-detail-client";

export const metadata = {
  title: "Vendor Detail",
  description: "View vendor profile, banking details, and recent purchase orders.",
};

export default async function VendorDetailPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { session } = await verifySession();
  if (!session?.user) {
    redirect("/login");
  }

  const { id } = await params;
  return <VendorDetailClient vendorId={id} />;
}
```

NOTE: confirm the `params` shape matches Next.js 15 conventions. Other detail pages in this repo (e.g. `frontend/src/app/(private)/(main)/budgets/[id]/page.tsx`) are the source of truth — match their `params: Promise<...>` style.

- [ ] **Step 2: Verify TS compiles**

Run from `frontend/`: `pnpm tsc --noEmit`
Expected: clean (will fail until VendorDetailClient is created in Task 4 — that's expected, fix in Task 4 commit boundary).

If you want to commit Task 3 + Task 4 together, skip the verify here and commit at the end of Task 4 with both files.

- [ ] **Step 3: (Optional) commit**

If proceeding to Task 4 immediately, defer commit to Task 4 step 5. Otherwise:

```bash
git add "frontend/src/app/(private)/(main)/vendors/[id]/page.tsx"
git commit -m "feat(vendors): add /vendors/[id] route entry"
```

---

## Task 4: VendorDetailClient + sub-components

**Files:**
- Create: `frontend/src/app/(private)/(main)/vendors/[id]/_components/vendor-detail-client.tsx`
- Create: `frontend/src/app/(private)/(main)/vendors/[id]/_components/vendor-profile-card.tsx`
- Create: `frontend/src/app/(private)/(main)/vendors/[id]/_components/vendor-banking-card.tsx`
- Create: `frontend/src/app/(private)/(main)/vendors/[id]/_components/vendor-recent-pos.tsx`

- [ ] **Step 1: Create `vendor-profile-card.tsx`**

```tsx
// frontend/src/app/(private)/(main)/vendors/[id]/_components/vendor-profile-card.tsx
"use client";

import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Mail, Phone, MapPin, User } from "lucide-react";
import type { Vendor } from "@/types/core";

interface VendorProfileCardProps {
  vendor: Vendor;
}

function Field({
  icon: Icon,
  label,
  value,
}: {
  icon: React.ElementType;
  label: string;
  value: React.ReactNode;
}) {
  return (
    <div className="flex items-start gap-3">
      <Icon
        className="h-4 w-4 text-muted-foreground shrink-0 mt-0.5"
        aria-hidden="true"
      />
      <div className="min-w-0 flex-1">
        <p className="text-xs font-medium text-muted-foreground uppercase tracking-wider">
          {label}
        </p>
        <p className="text-sm break-words">{value || "—"}</p>
      </div>
    </div>
  );
}

export function VendorProfileCard({ vendor }: VendorProfileCardProps) {
  const fullAddress = [vendor.physicalAddress, vendor.city, vendor.country]
    .filter(Boolean)
    .join(", ");

  return (
    <Card className="border-border/60">
      <CardHeader>
        <CardTitle className="text-base">Profile</CardTitle>
      </CardHeader>
      <CardContent className="grid gap-4 sm:grid-cols-2">
        <Field icon={User} label="Contact person" value={vendor.contactPerson} />
        <Field icon={Mail} label="Email" value={vendor.email} />
        <Field icon={Phone} label="Phone" value={vendor.phone} />
        <Field
          icon={MapPin}
          label="Address"
          value={fullAddress || vendor.physicalAddress}
        />
      </CardContent>
    </Card>
  );
}
```

- [ ] **Step 2: Create `vendor-banking-card.tsx`**

```tsx
// frontend/src/app/(private)/(main)/vendors/[id]/_components/vendor-banking-card.tsx
"use client";

import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Landmark, Hash, BadgeCheck } from "lucide-react";
import type { Vendor } from "@/types/core";

interface VendorBankingCardProps {
  vendor: Vendor;
}

function Row({ label, value }: { label: string; value: React.ReactNode }) {
  return (
    <div className="flex justify-between gap-4 py-2 border-b border-border/40 last:border-b-0">
      <span className="text-xs font-medium text-muted-foreground uppercase tracking-wider">
        {label}
      </span>
      <span className="text-sm font-mono tabular-nums break-all text-right">
        {value || "—"}
      </span>
    </div>
  );
}

export function VendorBankingCard({ vendor }: VendorBankingCardProps) {
  return (
    <Card className="border-border/60">
      <CardHeader>
        <CardTitle className="text-base flex items-center gap-2">
          <Landmark className="h-4 w-4" aria-hidden="true" />
          Banking
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-1">
          <Row label="Bank" value={vendor.bankName} />
          <Row label="Branch" value={vendor.branchCode} />
          <Row label="Account name" value={vendor.accountName} />
          <Row label="Account number" value={vendor.accountNumber} />
          <Row label="SWIFT" value={vendor.swiftCode} />
          <Row label="Tax ID" value={vendor.taxId} />
        </div>
      </CardContent>
    </Card>
  );
}
```

- [ ] **Step 3: Create `vendor-recent-pos.tsx`**

```tsx
// frontend/src/app/(private)/(main)/vendors/[id]/_components/vendor-recent-pos.tsx
"use client";

import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { DataList, type DataListColumn } from "@/components/ui/data-list";
import { StatusBadge } from "@/components/status-badge";
import { usePurchaseOrders } from "@/hooks/use-purchase-order-queries";
import { formatCurrency } from "@/lib/utils";
import { useRouter } from "next/navigation";
import type { PurchaseOrder } from "@/types/purchase-order";

interface VendorRecentPosProps {
  vendorId: string;
  /** Optional: limit visible rows. Default 10. */
  limit?: number;
}

export function VendorRecentPos({ vendorId, limit = 10 }: VendorRecentPosProps) {
  const router = useRouter();
  const { data: pos, isLoading } = usePurchaseOrders({ vendorId });
  const rows = (pos ?? []).slice(0, limit);

  const columns: DataListColumn<PurchaseOrder>[] = [
    {
      id: "documentNumber",
      header: "PO #",
      cell: (po) => (
        <span className="font-medium text-primary">{po.documentNumber}</span>
      ),
    },
    {
      id: "total",
      header: "Total",
      align: "right",
      priority: "md",
      cell: (po) => (
        <span className="tabular-nums">
          {formatCurrency((po as any).total ?? (po as any).totalAmount ?? 0)}
        </span>
      ),
    },
    {
      id: "status",
      header: "Status",
      cell: (po) => <StatusBadge status={po.status} type="document" />,
    },
    {
      id: "delivery",
      header: "Delivery",
      priority: "lg",
      cell: (po) =>
        (po as any).deliveryDate ? (
          <span className="text-sm text-muted-foreground">
            {new Date((po as any).deliveryDate).toLocaleDateString()}
          </span>
        ) : (
          <span className="text-muted-foreground">—</span>
        ),
    },
  ];

  return (
    <Card className="border-border/60">
      <CardHeader>
        <CardTitle className="text-base">Recent Purchase Orders</CardTitle>
      </CardHeader>
      <CardContent>
        <DataList<PurchaseOrder>
          rows={rows}
          columns={columns}
          getRowId={(po) => po.id}
          isLoading={isLoading}
          emptyMessage="No purchase orders for this vendor yet."
          onRowClick={(po) => router.push(`/purchase-orders/${po.id}`)}
          mobileCard={(po) => (
            <div className="flex flex-col gap-2">
              <div className="flex items-start justify-between gap-2">
                <div className="min-w-0">
                  <div className="font-medium text-primary line-clamp-1">
                    {po.documentNumber}
                  </div>
                </div>
                <StatusBadge status={po.status} type="document" />
              </div>
              <div className="flex flex-wrap items-center gap-2 text-xs text-muted-foreground">
                <span>
                  {formatCurrency(
                    (po as any).total ?? (po as any).totalAmount ?? 0
                  )}
                </span>
                {(po as any).deliveryDate && (
                  <span>
                    {new Date(
                      (po as any).deliveryDate
                    ).toLocaleDateString()}
                  </span>
                )}
              </div>
            </div>
          )}
        />
      </CardContent>
    </Card>
  );
}
```

NOTE: the `(po as any)` casts are a workaround for the `total ?? totalAmount` aliasing in the existing schema (Plan E used the same pattern for the same reason). Acceptable — proper unification is a Plan I cleanup.

- [ ] **Step 4: Create `vendor-detail-client.tsx`**

```tsx
// frontend/src/app/(private)/(main)/vendors/[id]/_components/vendor-detail-client.tsx
"use client";

import { useMemo } from "react";
import { useVendorById } from "@/hooks/use-vendor-queries";
import { usePurchaseOrders } from "@/hooks/use-purchase-order-queries";
import { DetailShell } from "@/components/layout/detail-shell";
import { PageHeader } from "@/components/base/page-header";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { StatGrid } from "@/components/ui/stat-grid";
import {
  AlertCircle,
  ShoppingCart,
  CheckCircle2,
  Clock,
  CircleDollarSign,
} from "lucide-react";
import { formatCurrency } from "@/lib/utils";
import { VendorProfileCard } from "./vendor-profile-card";
import { VendorBankingCard } from "./vendor-banking-card";
import { VendorRecentPos } from "./vendor-recent-pos";

interface VendorDetailClientProps {
  vendorId: string;
}

export function VendorDetailClient({ vendorId }: VendorDetailClientProps) {
  const { data: vendor, isLoading, error } = useVendorById(vendorId);
  const { data: pos } = usePurchaseOrders({ vendorId });

  const stats = useMemo(() => {
    const list = pos ?? [];
    const total = list.length;
    const approved = list.filter(
      (p) => p.status?.toUpperCase() === "APPROVED"
    ).length;
    const pending = list.filter((p) =>
      ["DRAFT", "PENDING", "PENDING_APPROVAL"].includes(
        p.status?.toUpperCase() ?? ""
      )
    ).length;
    const spend = list.reduce(
      (sum, p) =>
        sum + Number((p as any).total ?? (p as any).totalAmount ?? 0),
      0
    );
    return { total, approved, pending, spend };
  }, [pos]);

  if (isLoading) {
    return (
      <div className="space-y-5">
        <Skeleton className="h-12 w-full" />
        <Skeleton className="h-32 w-full" />
        <Skeleton className="h-72 w-full" />
      </div>
    );
  }

  if (error || !vendor) {
    return (
      <Alert variant="destructive">
        <AlertCircle className="h-4 w-4" />
        <AlertDescription>
          Failed to load vendor. The vendor may have been deleted.
        </AlertDescription>
      </Alert>
    );
  }

  return (
    <DetailShell
      header={
        <PageHeader
          title={vendor.name}
          subtitle={`Vendor code: ${vendor.vendorCode}`}
          showBackButton
          badges={[
            {
              status: vendor.active ? "active" : "inactive",
              type: "health",
            },
          ]}
        />
      }
      sidebar={
        <div className="space-y-5">
          <StatGrid
            items={[
              {
                label: "Total POs",
                value: stats.total,
                icon: <ShoppingCart className="h-3 w-3 sm:h-4 sm:w-4" />,
                accent: "blue",
              },
              {
                label: "Approved",
                value: stats.approved,
                icon: <CheckCircle2 className="h-3 w-3 sm:h-4 sm:w-4" />,
                accent: "emerald",
              },
              {
                label: "Pending",
                value: stats.pending,
                icon: <Clock className="h-3 w-3 sm:h-4 sm:w-4" />,
                accent: "amber",
              },
              {
                label: "Total spend",
                value: formatCurrency(stats.spend),
                icon: <CircleDollarSign className="h-3 w-3 sm:h-4 sm:w-4" />,
                accent: "violet",
              },
            ]}
          />
        </div>
      }
    >
      <div className="space-y-5">
        <VendorProfileCard vendor={vendor} />
        <VendorBankingCard vendor={vendor} />
        <VendorRecentPos vendorId={vendorId} />
      </div>
    </DetailShell>
  );
}
```

NOTES:
- `<PageHeader badges>` accepts an array of `{ status, type }` per existing API. If `type="health"` doesn't accept `"active"|"inactive"`, fall back to `<Badge>` inline alongside the title in PageHeader's `actions` slot. Verify by reading `frontend/src/lib/status-badges.ts` `HEALTH_STATUS_CONFIG`.
- `StatGrid` accepts 4 items → renders 4-up on `lg+`, 2-up on mobile. The sidebar slot in `<DetailShell>` is `lg+` only — on mobile the sidebar stacks below children, so the 4-up StatGrid will render across full width which is fine.

- [ ] **Step 5: Verify TS compiles**

Run from `frontend/`: `pnpm tsc --noEmit`
Expected: clean.

If `useVendorById` returns a different shape than expected, adjust the destructuring. If `PageHeader.badges` rejects the `type="health"` value, drop the badge and replace with an inline `<Badge>` next to the title (or in the `actions` slot).

- [ ] **Step 6: Commit (combined with Task 3 entry if not yet committed)**

```bash
git add \
  "frontend/src/app/(private)/(main)/vendors/[id]/page.tsx" \
  "frontend/src/app/(private)/(main)/vendors/[id]/_components/"
git commit -m "feat(vendors): add /vendors/[id] detail page (DetailShell + StatGrid + DataList)"
```

---

## Task 5: Make `/vendors` table rows clickable to detail

**Files:**
- Modify: `frontend/src/app/(private)/(main)/vendors/_components/vendors-table.tsx`

- [ ] **Step 1: Wire row click → push detail route**

Open the file. Find the `<DataList<Vendor>>` invocation. It already accepts an `onRowClick` prop (from Plan A `DataList`).

If `onRowClick` is currently set to `setSelectedVendor` (opening the inline panel), change to:

```tsx
import { useRouter } from "next/navigation";
// ... inside component
const router = useRouter();

<DataList<Vendor>
  // ... existing props
  onRowClick={(v) => router.push(`/vendors/${v.id}`)}
/>
```

Add the `useRouter` import if missing.

- [ ] **Step 2: Decide fate of inline `VendorDetailPanel`**

The vendors-table file likely has an inline `<VendorDetailPanel selected={selectedVendor} ...>` panel that opens beside the table. Two options:

(a) **Drop it.** Detail page replaces the inline panel. Cleanest.
(b) **Keep it.** As a compact preview-on-hover. More complex.

Choose option (a) unless the panel renders unique data (action buttons, edit form) not in the new detail page. To drop:

- Remove the `<VendorDetailPanel>` JSX.
- Remove `selectedVendor` state and any toggle handlers that exist solely to feed the panel.
- Remove the `VendorDetailPanel` import.

If the file's existing `selectedVendor` state is also used by an Edit button (opens `vendor-form-sheet` for in-place edit), keep that logic — the Edit action still lives in the row dropdown. Only drop the panel + state that ONLY existed for the panel.

- [ ] **Step 3: Verify TS compiles**

Run from `frontend/`: `pnpm tsc --noEmit`
Expected: clean.

- [ ] **Step 4: Commit**

```bash
git add "frontend/src/app/(private)/(main)/vendors/_components/vendors-table.tsx"
git commit -m "feat(vendors): vendors-table row click navigates to /vendors/[id]"
```

---

## Task 6: Verification pass

- [ ] **Step 1: Full type-check**

Run: `cd frontend && pnpm tsc --noEmit`
Expected: clean.

- [ ] **Step 2: Plan A–G regression**

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
Expected: all known-good tests still pass. Pre-existing submit-dialog failures (Plan E) and `/vendor-form-validation` property test (also pre-existing) are NOT Plan H regressions.

- [ ] **Step 3: Manual smoke (optional)**

If able, `cd frontend && pnpm dev`. Navigate to `/vendors`, click a row → verify it navigates to `/vendors/<id>`. Verify:
- Page header shows vendor name + code + active badge
- StatGrid shows 4 stats (renders OK even with 0 POs)
- Profile card shows contact + address
- Banking card shows account/SWIFT/etc
- Recent POs DataList renders (or empty state if vendor has no POs)
- Mobile viewport: sidebar stacks below main; StatGrid wraps to 2x2
- Back button returns to `/vendors`

Verify on mobile (390px wide): all panels render single-column, recent-POs DataList shows mobile cards.

- [ ] **Step 4: Cleanup commit if needed**

```bash
git add -A
git commit -m "chore(ui): plan-H verification cleanup"
```

---

## Self-Review Notes

- **Spec coverage:**
  - Plan G width carry-forward → Task 1 ✓
  - Action stub cleanup (3 stubs) → Task 2 ✓
  - Vendor detail route + client → Task 3 + 4 ✓
  - Vendor table row navigation → Task 5 ✓
  - Verification → Task 6 ✓

- **Type consistency:**
  - `vendorId: string` consistent across `VendorDetailClientProps`, `VendorRecentPosProps`, page params, and the `useVendorById` / `usePurchaseOrders({ vendorId })` hook calls.
  - `Vendor` type imported from `@/types/core` consistently across all 3 sub-cards.
  - `(po as any).total ?? (po as any).totalAmount` aliasing pattern matches existing usage in Plan E `purchase-orders-table.tsx`.

- **No placeholders:** Each step has explicit code, exact file paths, exact commands. No "TBD" or "implement later". The two notes about prop-shape verification (`PageHeader.badges` and Next 15 `params` shape) are legitimate verify-and-adapt instructions, not deferrals.

- **Out-of-scope discipline:** Detail-page monolith splits, `approval-history-panel.tsx` split, viewer dialogs, and edit-from-detail-page UX are all explicitly NOT in Plan H. Each named with rationale.

## Plan I carry-forward notes

- **Detail-page monolith splits.** PO 1152, req 920, PV 733, GRN 596, budget 457 LOC. Apply `<DetailShell>` like Plan H does for vendor.
- **`approval-history-panel.tsx` (1271 LOC).** Split into `ApprovalChainSummary` + `ActivityFeed` + `CommentsThread`. Replace inline approval-chain rendering with `<ApprovalChainStepper>`.
- **PDF / attachment viewer dialogs.** Need full-screen sheet variant — `<ResponsiveSheet>` body currently caps at `max-h-[90svh]`. Either add a `fullScreen?: boolean` prop or build a new `<FullScreenSheet>` primitive.
- **Confirmation-only dialog migration.** Evaluate whether to convert `*-submit-dialog`, `*-delete-dialog`, `mark-paid-dialog`, etc to `ResponsiveSheet` or keep them as `AlertDialog` with mobile-friendly sizing.
- **PO `total` vs `totalAmount` field aliasing.** Unify schema or pick one canonical field. Plan H + E used `(po as any).total ?? (po as any).totalAmount` workaround.
- **Server-side vendor pagination.** `useVendors()` doesn't accept page/limit yet. Add backend filter param + frontend wiring.
- **Pre-existing test failures from Plan E (submit-dialog 5 failures, vendor-form-validation 1 failure).** Update tests for new component structures.
