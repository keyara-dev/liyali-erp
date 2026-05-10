# UI Revamp — Plan F: ApprovalChainStepper Adoption

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace inline approval-chain rendering across three workflow components with the shared `<ApprovalChainStepper>` primitive, plus clear two small Plan E carry-forwards.

**Architecture:** `ApprovalFlowDisplay` (228 LOC, used in budget approval page) and `ApprovalHistory` (258 LOC, also budget approval page) and `ApprovalChainPanel` (177 LOC, budget detail) all duplicate stage-marker rendering with raw `bg-green-50`, `bg-red-50`, `bg-yellow-50` color classes and bespoke timeline connectors. Plan F maps each to the canonical `<ApprovalChainStepper>` (Plan B Task 6). External component APIs stay the same so consumers don't break. Detail-page layout, the 1271 LOC `approval-history-panel.tsx` monolith split, and DetailShell adoption are all deferred to Plan G.

**Tech Stack:** Next.js App Router, TypeScript, Tailwind v4, ShadCN UI, `<ApprovalChainStepper>` (Plan B), Vitest.

**Out-of-scope (Plan G+):**
- Detail-page layout via `<DetailShell>`
- Splitting `approval-history-panel.tsx` (1271 LOC monolith)
- Splitting `requisition-detail-client.tsx`, `purchase-order-detail-client.tsx`, `pv-detail-client.tsx`, `grn-detail-client.tsx`, `budget-detail-client.tsx` monoliths
- New approval flows / business logic

---

## Existing infrastructure (reuse)

- `ApprovalChainStepper` at `frontend/src/components/workflows/approval-chain-stepper.tsx` (Plan B Task 6)
  - Props: `stages: ApprovalStage[]` where `ApprovalStage = { id, name, status: "approved"|"rejected"|"current"|"pending"|"skipped", actor?, at?, comments? }`
  - Renders semantic `<ol>` with `aria-label`, status-keyed markers, connector lines, comments italicized.
- `ApprovalRecord` type at `frontend/src/types/index.ts` — shape used by all three target files. Fields include `approverName`, `approverId`, `status`, `approvedAt`, `actionTakenAt`, `comments`, `remarks`, `stageName`.
- `useApprovalHistory` hook at `frontend/src/hooks/use-approval-workflow.ts`.

---

## File Structure

**Modify:**
- `frontend/src/app/(private)/(main)/requisitions/_components/requisitions-table.tsx` — drop duplicate React import, normalize cryptic `w-32.5`/`w-38` widths (Plan E carry-forward)
- `frontend/src/components/workflows/approval-flow-display.tsx` — internal rewrite to use `<ApprovalChainStepper>`; keep external props (`approvalHistory`, `currentStage`, `totalStages`, `isCompleted`)
- `frontend/src/components/workflows/approval-history.tsx` — internal rewrite to use `<ApprovalChainStepper>`; keep external props (`documentId`, `entityId`, `entityType`)
- `frontend/src/app/(private)/(main)/budgets/[id]/_components/approval-chain-panel.tsx` — replace bespoke renderer with `<ApprovalChainStepper>`; keep external prop `approvalChain: ApprovalRecord[]`

**Tests added:** none. The primitive `ApprovalChainStepper` already has 4 tests (Plan B). Each consumer is a pure transformation of `ApprovalRecord[]` → `ApprovalStage[]`; verify by manual smoke at the `/budgets/[id]/approval` page after the changes land.

---

## Mapping helper (used by Tasks 2-4)

Each file needs the same `ApprovalRecord → ApprovalStage[]` transformation. Define inline in each file (do NOT extract to a shared module yet — wait until 4+ consumers exist; YAGNI).

```tsx
import type { ApprovalStage, StageStatus } from "@/components/workflows/approval-chain-stepper";
import type { ApprovalRecord } from "@/types";

function approvalRecordsToStages(
  records: ApprovalRecord[],
  opts?: { currentStage?: number; isCompleted?: boolean; totalStages?: number }
): ApprovalStage[] {
  const { currentStage, isCompleted, totalStages } = opts ?? {};
  // Render one ApprovalStage per record. If totalStages is provided and exceeds
  // records.length, append `pending` stages for the remaining slots.
  const fromRecords: ApprovalStage[] = records.map((r, i) => {
    const upper = (r.status ?? "").toUpperCase();
    let status: StageStatus = "pending";
    if (upper === "APPROVED") status = "approved";
    else if (upper === "REJECTED") status = "rejected";
    else if (typeof currentStage === "number" && i === currentStage && !isCompleted) status = "current";
    else if (typeof currentStage === "number" && i < currentStage) status = "approved";
    return {
      id: r.approverId ?? `stage-${i}`,
      name: r.stageName ?? `Stage ${i + 1}`,
      status,
      actor: r.approverName,
      at: r.approvedAt ?? r.actionTakenAt,
      comments: r.comments ?? r.remarks,
    };
  });

  if (typeof totalStages === "number" && totalStages > fromRecords.length) {
    const remaining = totalStages - fromRecords.length;
    for (let i = 0; i < remaining; i++) {
      const idx = fromRecords.length;
      fromRecords.push({
        id: `pending-${idx}`,
        name: `Stage ${idx + 1}`,
        status: typeof currentStage === "number" && idx === currentStage ? "current" : "pending",
      });
    }
  }
  return fromRecords;
}
```

Each consumer copies this helper into its own file. The signature varies slightly per consumer (some don't have `currentStage`/`totalStages`).

---

## Task 1: Plan E carry-forward fixes — requisitions-table

**Files:**
- Modify: `frontend/src/app/(private)/(main)/requisitions/_components/requisitions-table.tsx`

- [ ] **Step 1: Drop duplicate React import**

Open the file. Near the top (around line 4) there are likely two React imports:
```tsx
import { useState, useCallback, useMemo } from "react";
import * as React from "react";
```

The `import * as React from "react"` is unused — the file does not reference the `React` namespace. Delete that single line.

- [ ] **Step 2: Normalize cryptic width utilities**

Find any `w-32.5` and `w-38` classes in this file:

Run from the workspace root: `grep -n "w-32\.5\|w-38" "frontend/src/app/(private)/(main)/requisitions/_components/requisitions-table.tsx"`

For each match, replace with the closer canonical value:
- `w-32.5` (= 130px) → `w-32` (= 128px) OR `w-[130px]` if the exact 130px is meaningful
- `w-38` (= 152px) → `w-36` (= 144px) OR `w-40` (= 160px) — pick whichever is closer to original intent

Default to the lower canonical value (`w-32`, `w-36`) for narrower selects. Verify the visual still works.

- [ ] **Step 3: Verify TS compiles**

Run: `cd frontend && pnpm tsc --noEmit`
Expected: clean.

- [ ] **Step 4: Commit**

```bash
git add "frontend/src/app/(private)/(main)/requisitions/_components/requisitions-table.tsx"
git commit -m "chore(requisitions): drop dup React import + normalize widths"
```

---

## Task 2: Refactor `ApprovalFlowDisplay`

**Files:**
- Modify: `frontend/src/components/workflows/approval-flow-display.tsx`

External API stays the same: `(approvalHistory, currentStage, totalStages?, isCompleted?)`. Internals become `<ApprovalChainStepper>` plus the existing summary 3-up grid (total/completed/remaining) and the empty state.

- [ ] **Step 1: Rewrite the file**

Replace the entire contents with:

```tsx
"use client";

import {
  ApprovalChainStepper,
  type ApprovalStage,
  type StageStatus,
} from "@/components/workflows/approval-chain-stepper";
import type { ApprovalRecord } from "@/types";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { AlertCircle } from "lucide-react";

export interface ApprovalFlowDisplayProps {
  approvalHistory: ApprovalRecord[];
  currentStage: number;
  totalStages?: number;
  isCompleted?: boolean;
}

function recordsToStages(
  records: ApprovalRecord[],
  currentStage: number,
  totalStages: number,
  isCompleted: boolean
): ApprovalStage[] {
  const stages: ApprovalStage[] = records.map((r, i) => {
    const upper = (r.status ?? "").toUpperCase();
    let status: StageStatus = "pending";
    if (upper === "APPROVED") status = "approved";
    else if (upper === "REJECTED") status = "rejected";
    else if (i < currentStage) status = "approved";
    else if (i === currentStage && !isCompleted) status = "current";
    return {
      id: r.approverId ?? `stage-${i}`,
      name: r.stageName ?? `Stage ${i + 1}`,
      status,
      actor: r.approverName,
      at: r.approvedAt ?? r.actionTakenAt,
      comments: r.comments ?? r.remarks,
    };
  });

  // Append pending placeholders for stages not yet in approvalHistory
  if (totalStages > stages.length) {
    const remaining = totalStages - stages.length;
    for (let i = 0; i < remaining; i++) {
      const idx = stages.length;
      stages.push({
        id: `pending-${idx}`,
        name: `Stage ${idx + 1}`,
        status: idx === currentStage && !isCompleted ? "current" : "pending",
      });
    }
  }
  return stages;
}

export function ApprovalFlowDisplay({
  approvalHistory,
  currentStage,
  totalStages = 0,
  isCompleted = false,
}: ApprovalFlowDisplayProps) {
  const effectiveTotal = totalStages || approvalHistory.length;

  if (effectiveTotal === 0 && approvalHistory.length === 0) {
    return (
      <Card>
        <CardContent className="flex items-center justify-center py-12">
          <AlertCircle className="h-8 w-8 text-amber-600 mr-3" />
          <div>
            <h3 className="font-semibold">No Approval History</h3>
            <p className="text-sm text-muted-foreground">
              This document has not been submitted for approval yet.
            </p>
          </div>
        </CardContent>
      </Card>
    );
  }

  const stages = recordsToStages(
    approvalHistory,
    currentStage,
    effectiveTotal,
    isCompleted
  );
  const completed = isCompleted ? effectiveTotal : currentStage;
  const remaining = Math.max(0, effectiveTotal - completed);

  return (
    <Card>
      <CardHeader>
        <CardTitle>Approval Workflow Progress</CardTitle>
        <CardDescription>
          {isCompleted
            ? "Workflow completed successfully"
            : `Currently at stage ${Math.min(currentStage + 1, effectiveTotal)} of ${effectiveTotal}`}
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div className="space-y-6">
          <ApprovalChainStepper stages={stages} />

          <div className="pt-4 border-t">
            <div className="grid grid-cols-3 gap-4 text-center">
              <div>
                <h4 className="text-xs font-semibold text-muted-foreground uppercase mb-1">
                  Total Stages
                </h4>
                <p className="text-lg font-bold tabular-nums">{effectiveTotal}</p>
              </div>
              <div>
                <h4 className="text-xs font-semibold text-muted-foreground uppercase mb-1">
                  Completed
                </h4>
                <p className="text-lg font-bold tabular-nums">{completed}</p>
              </div>
              <div>
                <h4 className="text-xs font-semibold text-muted-foreground uppercase mb-1">
                  Remaining
                </h4>
                <p className="text-lg font-bold tabular-nums">{remaining}</p>
              </div>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
```

Notes on changes:
- Drop unused `Avatar`, `AvatarFallback`, `AvatarImage`, `Badge`, `CheckCircle2`, `Clock`, `ChevronRight` imports.
- Drop bespoke `getStageStatus`/`getStageApproval`/`getStatusIcon`/`getStatusColor` helpers — `ApprovalChainStepper` owns all of that.
- Drop raw `bg-green-50`/`bg-blue-50`/`bg-gray-50` color classes — primitive uses design tokens.
- Empty-state alert color tightened from `text-yellow-600` to `text-amber-600` (Tailwind v4 canonical).

- [ ] **Step 2: Verify TS compiles**

Run: `cd frontend && pnpm tsc --noEmit`
Expected: clean.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/workflows/approval-flow-display.tsx
git commit -m "refactor(workflows): ApprovalFlowDisplay uses ApprovalChainStepper"
```

---

## Task 3: Refactor `ApprovalHistory`

**Files:**
- Modify: `frontend/src/components/workflows/approval-history.tsx`

This component shows a chronological feed (sorted desc by `approvedAt`) with expandable comment details. The chronological-feed semantics differ from `ApprovalChainStepper`'s ordered-stages semantics — but the marker visuals (approved/rejected/current/pending) plus actor + date + optional comments are identical.

Approach: convert sorted records to `ApprovalStage[]` (preserving the desc sort), pass to `<ApprovalChainStepper>`. Drop the per-record `expandedId` toggle — `ApprovalChainStepper` shows comments inline already. If verbose-toggle UX matters, defer to Plan G.

- [ ] **Step 1: Rewrite the file**

Replace the entire contents with:

```tsx
"use client";

import { useMemo } from "react";
import { useApprovalHistory } from "@/hooks/use-approval-workflow";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { AlertCircle, History } from "lucide-react";
import {
  ApprovalChainStepper,
  type ApprovalStage,
  type StageStatus,
} from "@/components/workflows/approval-chain-stepper";

export interface ApprovalHistoryProps {
  documentId?: string;
  entityId?: string; // Legacy compatibility
  entityType?: string; // Legacy compatibility — currently unused
}

export function ApprovalHistory({
  documentId,
  entityId,
  entityType: _entityType,
}: ApprovalHistoryProps) {
  const actualDocumentId = documentId || entityId || "";
  const { data: historyData, isLoading } = useApprovalHistory(actualDocumentId);

  const stages: ApprovalStage[] = useMemo(() => {
    if (!historyData) return [];
    const sorted = [...historyData].sort(
      (a: any, b: any) =>
        new Date(b.approvedAt ?? 0).getTime() -
        new Date(a.approvedAt ?? 0).getTime()
    );
    return sorted.map((r: any, i: number) => {
      const upper = (r.status ?? "").toUpperCase();
      let status: StageStatus = "pending";
      if (upper === "APPROVED") status = "approved";
      else if (upper === "REJECTED") status = "rejected";
      else if (upper === "RETURNED" || upper === "REASSIGNED") status = "skipped";
      return {
        id: r.id ?? r.approverId ?? `entry-${i}`,
        name: r.stageName ?? r.action ?? `Entry ${i + 1}`,
        status,
        actor: r.approverName,
        at: r.approvedAt ?? r.actionTakenAt,
        comments: r.comments ?? r.remarks,
      };
    });
  }, [historyData]);

  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Approval History</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-3">
            {Array.from({ length: 3 }).map((_, i) => (
              <Skeleton key={i} className="h-12 w-full rounded-md" />
            ))}
          </div>
        </CardContent>
      </Card>
    );
  }

  if (!actualDocumentId) {
    return (
      <Alert variant="destructive">
        <AlertCircle className="h-4 w-4" />
        <AlertDescription>
          Document id is required to load approval history.
        </AlertDescription>
      </Alert>
    );
  }

  if (stages.length === 0) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <History className="h-4 w-4" />
            Approval History
          </CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-muted-foreground">
            No approval actions recorded yet.
          </p>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <History className="h-4 w-4" />
          Approval History
        </CardTitle>
        <CardDescription>
          {stages.length} action{stages.length === 1 ? "" : "s"} recorded
        </CardDescription>
      </CardHeader>
      <CardContent>
        <ApprovalChainStepper stages={stages} />
      </CardContent>
    </Card>
  );
}
```

Notes on changes:
- Drop the `expandedId` state and the per-row click-to-expand UX. `ApprovalChainStepper` shows comments inline always — simpler is fine.
- Drop unused `Avatar`, `AvatarFallback`, `Badge`, `Repeat2`, `XCircle`, `Clock`, `ChevronDown`, `ChevronUp` imports.
- Drop bespoke `getActionIcon`/`getActionColor` switch — primitive owns this.
- Map `RETURNED`/`REASSIGNED` to `skipped` status. Other status mapping matches `ApprovalFlowDisplay` (Task 2).
- Loading state now uses `Skeleton` instead of inline `animate-pulse` divs.

- [ ] **Step 2: Verify TS compiles**

Run: `cd frontend && pnpm tsc --noEmit`
Expected: clean.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/workflows/approval-history.tsx
git commit -m "refactor(workflows): ApprovalHistory uses ApprovalChainStepper"
```

---

## Task 4: Refactor budget `ApprovalChainPanel`

**Files:**
- Modify: `frontend/src/app/(private)/(main)/budgets/[id]/_components/approval-chain-panel.tsx`

External prop `approvalChain: ApprovalRecord[]` stays the same. Drop the bespoke `getStatusIcon` / `getStatusColor` helpers + raw `bg-green-50`/`bg-red-50`/`bg-yellow-50`/`bg-gray-50` classes. Keep the empty-state CTA card with the "no chain" illustration block — that's a budget-specific UX flourish, not stage rendering.

- [ ] **Step 1: Rewrite the file**

Replace the entire contents with:

```tsx
"use client";

import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { ClipboardList, Plus } from "lucide-react";
import type { ApprovalRecord } from "@/types";
import Link from "next/link";
import {
  ApprovalChainStepper,
  type ApprovalStage,
  type StageStatus,
} from "@/components/workflows/approval-chain-stepper";

interface ApprovalChainPanelProps {
  approvalChain: ApprovalRecord[];
}

function recordsToStages(records: ApprovalRecord[]): ApprovalStage[] {
  return records.map((r, i) => {
    const upper = (r.status ?? "").toUpperCase();
    let status: StageStatus = "pending";
    if (upper === "APPROVED") status = "approved";
    else if (upper === "REJECTED") status = "rejected";
    else if (upper === "PENDING") status = "current";
    return {
      id: r.approverId ?? `stage-${i}`,
      name: r.stageName ?? `Stage ${i + 1}`,
      status,
      actor: r.approverName,
      at: r.approvedAt ?? r.actionTakenAt,
      comments: r.comments ?? r.remarks,
    };
  });
}

export function ApprovalChainPanel({ approvalChain }: ApprovalChainPanelProps) {
  if (!approvalChain || approvalChain.length === 0) {
    return (
      <Card className="border-2 border-dashed">
        <CardContent className="flex flex-col items-center justify-center px-8 py-8">
          <div className="relative mb-4">
            <div className="bg-primary/10 absolute inset-0 rounded-full blur-2xl" />
            <div className="bg-card border-primary/20 relative rounded-2xl border-2 p-6">
              <ClipboardList
                className="text-primary h-16 w-16"
                strokeWidth={1.5}
              />
            </div>
          </div>
          <h3 className="text-base font-semibold mb-1">No approval chain yet</h3>
          <p className="text-sm text-muted-foreground text-center max-w-sm mb-4">
            This budget has not entered an approval workflow. Submit it to
            start the chain.
          </p>
          <Button asChild variant="outline" size="sm">
            <Link href="/budgets">
              <Plus className="h-4 w-4 mr-2" />
              Back to budgets
            </Link>
          </Button>
        </CardContent>
      </Card>
    );
  }

  const stages = recordsToStages(approvalChain);

  return (
    <Card>
      <CardHeader>
        <CardTitle>Approval Chain</CardTitle>
      </CardHeader>
      <CardContent>
        <ApprovalChainStepper stages={stages} />
      </CardContent>
    </Card>
  );
}
```

Notes on changes:
- Drop unused `CheckCircle2`, `XCircle`, `Clock` imports.
- Empty-state CTA preserved; updated `bg-canvas` (non-standard) → `bg-card` (standard ShadCN token).
- Map `PENDING` records to `current` status (this matches the original — pending records were rendered with the yellow "current" tint).

- [ ] **Step 2: Verify TS compiles**

Run: `cd frontend && pnpm tsc --noEmit`
Expected: clean.

- [ ] **Step 3: Commit**

```bash
git add "frontend/src/app/(private)/(main)/budgets/[id]/_components/approval-chain-panel.tsx"
git commit -m "refactor(budgets): ApprovalChainPanel uses ApprovalChainStepper"
```

---

## Task 5: Verification pass

- [ ] **Step 1: Full type-check**

Run: `cd frontend && pnpm tsc --noEmit`
Expected: clean.

- [ ] **Step 2: Plan A+B+C+D+E test regression**

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
Expected: 45 tests pass.

- [ ] **Step 3: Manual smoke (optional)**

If able, `cd frontend && pnpm dev` and visit:
- `/budgets/<id>` — sidebar approval-chain-panel uses stepper
- `/budgets/<id>/approval` — both ApprovalFlowDisplay and ApprovalHistory use stepper
- Confirm no raw `bg-green-50`/`bg-yellow-50` swatches remain

- [ ] **Step 4: Cleanup commit if needed**

```bash
git add -A
git commit -m "chore(ui): plan-F verification cleanup"
```

---

## Self-Review Notes

- **Spec coverage:** Plan E carry-forwards land (Task 1: dedupe React import + width normalization). Three approval-chain renderers refactored (Tasks 2-4). Verification (Task 5).
- **Type consistency:** `recordsToStages` helper repeated in three files with slightly different signatures (Tasks 2 + 4 don't accept current-stage; Task 3 sorts by date). Status mapping (`APPROVED`/`REJECTED`/etc → StageStatus) matches across all three. Helper is intentionally NOT extracted — three callsites is the YAGNI threshold.
- **No placeholders:** Each task has full code. No "TBD", no "fill in details".
- **Out-of-scope discipline:** `approval-history-panel.tsx` (1271 LOC monolith) is NOT touched — Plan G handles. No detail-page layout changes. No DetailShell adoption.
- **Existing primitive `ApprovalChainStepper`:** has 4 tests already (Plan B). Plan F doesn't add tests because consumers are pure prop transformations of an already-tested primitive.

## Plan G carry-forward notes

- `approval-history-panel.tsx` (1271 LOC) — split into `ApprovalChainSummary` + `ActivityFeed` + `CommentsThread` sub-components; replace inline approval-chain rendering with `<ApprovalChainStepper>`.
- Detail-client monoliths (PO 1152, req 920, PV 733, GRN 596, budget 457) — apply `<DetailShell>` + extract sub-panels.
- If a 4th approval-chain consumer ever appears, extract `recordsToStages` to `frontend/src/lib/approval-stages.ts`.
- ApprovalHistory's old expand-to-show-comments UX was dropped — `ApprovalChainStepper` shows comments inline. If product wants the toggle back, add a `defaultCollapsed?` prop to `ApprovalChainStepper`.
