# Procurement Flow Hardening — PO → PV → GRN

**Date:** 2026-06-05
**Status:** Approved for implementation (Phase A first)
**Author:** audit + design session

## Background

A four-domain audit of the Purchase Order → Payment Voucher → Goods Received
Note flow found correctness bugs, an IDOR gap, fragmented automation, and a
terminal dead-end in the payment_first mode. This spec captures the agreed
fixes and the target automation model.

## Canonical model (decisions)

- **Document chains by mode:**
  - `goods_first`  = PO → GRN → PV  (receive goods, then pay)
  - `payment_first` = PO → PV → GRN  (pay, then receive goods)
- **PO=APPROVED is a hard entry gate** for creating any PV or GRN, in **both**
  modes. Auto-approval still flows: the PO auto-approves, then the chain proceeds.
- **Automation is per-org, per-document, three levels:** `manual` (create as
  DRAFT, human submits) | `auto_submit` (create + submit to workflow) |
  `auto_approve` (create + submit + approve). `auto_approve` is bounded by an
  org-level `AutoApproveMaxAmount`; documents above the cap fall back to
  `auto_submit` so a human always approves large amounts.
- **Status transitions:** keep targeted per-call-site guards now; centralizing
  all transitions through `WorkflowStateMachine` is a separate later effort.

---

## Phase A — Correctness & security (no new behavior)

Each item is an independent, testable change. Ship as discrete commits.

### A1 — Close IDOR on item/budget updates
`UpdatePurchaseOrderItems` (purchase_order.go:~1130) and `UpdateBudget`
(budget.go:~351) load with only `id + organization_id`. Add
`utils.GetDocumentScope(...).ApplyToQuery(query, ownerField, entityType, "")`
exactly like their sibling mutations (`UpdatePurchaseOrder`, `DeletePurchaseOrder`,
`DeleteBudget`). Owner field: `created_by` for PO, the budget's owner column for
budget.

### A2 — Port missing guards into `CreatePaymentVoucherFromPO` + `AutoCreatePVFromCompletedGRN`
The `/from-po` endpoint (document_extras_handler.go:316-412) is the endpoint the
UI actually calls and currently enforces none of the guards the manual
`CreatePaymentVoucher` (payment_voucher.go:267-356) has. Add, matching the
manual path:
- **PO=APPROVED gate** — reject if `strings.ToUpper(po.Status) != "APPROVED"`.
- **Duplicate-PV guard** — reject if a non-CANCELLED/non-REJECTED PV already
  links this PO (allows retry after terminal failure).
- **Amount caps** — `req.TotalAmount <= po.TotalAmount` and (goods_first)
  `req.TotalAmount <= receivedValue`.
- **GRN status** — accept `APPROVED` **or** `COMPLETED` (the GRN auto-advances to
  COMPLETED; the current APPROVED-only check blocks the normal goods_first case).

Apply the same caps/guards inside `AutoCreatePVFromCompletedGRN`
(workflow_execution_service.go:2336) as defense-in-depth.

### A3 — PO=APPROVED entry gate for GRN create in payment_first
`CreateGRN` (grn.go:~274) skips the PO-APPROVED requirement for payment_first.
Per the decision, enforce PO=APPROVED in **both** modes. This removes the
dead-end where a payment_first GRN created against a non-APPROVED PO could never
be submitted (SubmitGRN requires APPROVED) and the cascades (which gate on
APPROVED) never close the PO.

### A4 — `MarkGRNComplete` re-validates PO/PV
`MarkGRNComplete` (grn.go:1316-1321) checks only signoff=READY + status=DRAFT
before cascading. Mirror `SubmitGRN`'s re-checks: linked PO still APPROVED, linked
PV (if any) still valid. Prevents cascading delivery onto a cancelled/rejected PO
and auto-creating a PV against it.

### A5 — Converge the `MarkPaidWithPOP` PAID path
`MarkPaidWithPOP` (document_extras_handler.go:942) flips status to PAID directly:
no amount-match, no routing_type guard, leaves the `payment_execution` workflow
task open. Add: amount-match within 0.01 of the approved amount; complete/close
the open `payment_execution` task (reuse the logic in `MarkPaymentVoucherPaid`);
keep it usable for direct_payment but guard procurement PVs from bypassing the
workflow task.

### A6 — Single procurement-flow resolver
Add `resolveProcurementFlow(db, po, orgID) string` returning a normalized
(`strings.ToLower(strings.TrimSpace(...))`) flow with precedence PO override →
org default → `goods_first`. Replace the 4 ad-hoc resolutions (grn.go:261,
document_extras_handler.go:338, payment_voucher.go:295/300,
workflow_execution_service.go:2351). Two of them currently skip normalization.

### A7 — Server-side total recompute
On create of PO / REQ / PV, compute the stored total from line items
(`Σ(quantity × unitPrice)`, plus tax + delivery for PO) rather than trusting the
client `totalAmount`. PV *update* already recomputes; make create consistent.
**Overwrite** the stored total with the computed value (single source of truth);
do not merely warn on mismatch.

### A8 — Robust FULLY_DELIVERED matching
`cascadeGRNApprovalToPO` (workflow_execution_service.go:2262-2286) matches GRN
lines to PO lines by lowercased-trimmed description. Switch to matching by item
`id` (fall back to index) so PO line edits, duplicate descriptions, or whitespace
don't block completion. Also guard the empty-items edge case.

### A9 — Low-severity cleanups
- Manual `CreatePaymentVoucher`: inherit PO currency when `LinkedPO` is set.
- `WithdrawPaymentVoucher`: use `voucher.CreatedBy` instead of re-deriving creator
  from ActionHistory (fixes inability to withdraw auto-created PVs).
- `MarkPaidWithPOP`: `io.ReadFull` instead of `f.Read` for PoP bytes.
- Drop the non-canonical `super_admin` string from gateway role gates (or document
  why it stays). Keep `admin`/`finance`.
- `cascadeGRNApprovalToPO`: only write delivery_status when PO status warrants it.

Out of scope for A (deferred): receiver-sign-off actor gate (product question),
`LinkedRequisition` UUID-vs-docnumber normalization (cosmetic).

---

## Phase B — Configurable automation model (new behavior)

Consolidate the two fragmented automation systems
(`OrganizationSettings` flags + `document_automation_service.AutomationConfig`)
into **OrganizationSettings**, wired to the Settings UI.

### Settings schema (replaces the 3 orphaned bools)
```
AutoCreateGRNFromPO   bool      // goods_first: PO approved → spawn GRN
AutoCreatePVFromPO    bool      // payment_first: PO approved → spawn PV
AutoCreatePVFromGRN   bool      // goods_first: GRN completed → spawn PV
GRNAutomationLevel    string    // manual | auto_submit | auto_approve
PVAutomationLevel     string    // manual | auto_submit | auto_approve
AutoApproveMaxAmount  float64   // auto_approve applies at/below; above → auto_submit
```

### Orchestration (mode-aware)
- **goods_first:** PO approved → (AutoCreateGRNFromPO) GRN DRAFT → apply
  `GRNAutomationLevel` → on GRN COMPLETED → (AutoCreatePVFromGRN) PV → apply
  `PVAutomationLevel`.
- **payment_first:** PO approved → (AutoCreatePVFromPO) PV → apply
  `PVAutomationLevel`; GRN created later, apply `GRNAutomationLevel`.

`applyAutomationLevel(doc, level, amount, cap)`:
- `manual` → leave DRAFT.
- `auto_submit` → submit to workflow.
- `auto_approve` → if `amount <= cap` submit + auto-approve; else auto_submit.

### Wiring
- Add the fields to `UpdateOrganizationSettings` request struct, model-build, and
  service update column map.
- Add a frontend Settings UI section (workspace-settings) for the flags +
  per-document level selects + the cap.
- Retire `document_automation_service.AutomationConfig` duplication or have it read
  from `OrganizationSettings` (single source of truth).

### Phase B testing
- Unit: `applyAutomationLevel` cap fallback; resolver precedence.
- Integration: full goods_first and payment_first auto-chains at each level.

---

## Testing (Phase A)
- Backend unit/integration per fix (Go). Reuse existing handler test files.
- Each guard gets a negative test (rejects bad input) + positive (happy path
  still works).
- Run `go build ./...` + `go test ./...` and frontend `tsc --noEmit` before each
  commit.

## Rollout
Phase A first (independent commits), then Phase B as a second design/plan pass.
