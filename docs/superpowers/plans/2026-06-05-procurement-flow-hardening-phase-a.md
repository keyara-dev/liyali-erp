# Procurement Flow Hardening — Phase A Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Close the correctness + security gaps in the PO → PV → GRN flow found in the 2026-06-05 audit, without adding new behavior.

**Architecture:** Targeted per-call-site guards. Extract two shared validators (PV-creation gate, GRN-link revalidation) so the divergent paths converge on one implementation. No central state-machine adoption in this phase.

**Tech Stack:** Go (Fiber + GORM), backend only. Tests via `go test ./...`. Reference spec: `docs/superpowers/specs/2026-06-05-procurement-flow-hardening-design.md`.

**Per-task loop:** write failing test → run (fails) → implement → run (passes) → `go build ./...` → commit. Run from `d:\dev\next-apps\liyali-gateway\backend`.

---

## File Structure

- `backend/handlers/payment_voucher.go` — extract `validateProcurementPVGate` from the manual PV path; reuse it.
- `backend/handlers/document_extras_handler.go` — call the validator in `CreatePaymentVoucherFromPO`; converge `MarkPaidWithPOP`; `io.ReadFull`.
- `backend/services/workflow_execution_service.go` — call the validator in `AutoCreatePVFromCompletedGRN`; robust FULLY_DELIVERED matching.
- `backend/handlers/grn.go` — PO=APPROVED gate for all flows; extract `revalidateGRNLinks` and call from `MarkGRNComplete`.
- `backend/handlers/purchase_order.go` — scope on `UpdatePurchaseOrderItems`; server-side total recompute on create.
- `backend/handlers/budget.go` — scope on `UpdateBudget`.
- `backend/utils/procurement_flow.go` (new) — `ResolveProcurementFlow` helper.

---

## Task 1: Close IDOR on `UpdatePurchaseOrderItems` (A1a)

**Files:** Modify `backend/handlers/purchase_order.go` (~1129-1130). Test: `backend/handlers/purchase_orders_http_test.go`.

- [ ] **Step 1: Write failing test** — a user who is neither creator nor involved gets 404 when PUT `/purchase-orders/:id/items` on another user's DRAFT PO.

```go
func TestUpdatePurchaseOrderItems_NonOwnerScopedOut(t *testing.T) {
    // Arrange: DRAFT PO created_by userA in org; caller = userB (role "requester", not involved).
    // Act: PUT /purchase-orders/{id}/items with valid items+total as userB.
    // Assert: 404 (scoped out), and the PO items are unchanged in DB.
}
```

- [ ] **Step 2: Run, verify FAIL** — `go test ./handlers/ -run TestUpdatePurchaseOrderItems_NonOwnerScopedOut -v` → currently 200 (bug).

- [ ] **Step 3: Implement** — replace the load in `UpdatePurchaseOrderItems`:

```go
scope := utils.GetDocumentScope(config.DB, tenant.UserID, tenant.UserRole, tenant.OrganizationID)
loadQuery := config.DB.Where("id = ? AND organization_id = ?", id, tenant.OrganizationID)
loadQuery = scope.ApplyToQuery(loadQuery, "created_by", "purchase_order", "")
var order models.PurchaseOrder
if err := loadQuery.First(&order).Error; err != nil {
    // existing not-found handling
}
```

(Mirrors `UpdatePurchaseOrder` at purchase_order.go:494-499.)

- [ ] **Step 4: Run, verify PASS** + add a positive test: creator still edits successfully (200).
- [ ] **Step 5:** `go build ./...`
- [ ] **Step 6: Commit** — `fix(po): scope UpdatePurchaseOrderItems to owner/involvement (IDOR)`

---

## Task 2: Close IDOR on `UpdateBudget` (A1b)

**Files:** Modify `backend/handlers/budget.go:349-354`. Test: budget handler test file.

- [ ] **Step 1: Failing test** — non-owner non-privileged user gets 404 on PUT `/budgets/:id`.
- [ ] **Step 2: Run, verify FAIL.**
- [ ] **Step 3: Implement** — replace load at 349-354 (mirror `DeleteBudget` at 530-532):

```go
scope := utils.GetDocumentScope(config.DB, tenant.UserID, tenant.UserRole, tenant.OrganizationID)
loadQuery := config.DB.Where("id = ? AND organization_id = ?", id, tenant.OrganizationID)
loadQuery = scope.ApplyToQuery(loadQuery, "created_by", "budget", "")
var budget models.Budget
if err := loadQuery.First(&budget).Error; err != nil {
    logging.LogError(c, err, "budget_not_found_for_update")
    return utils.SendNotFoundError(c, "Budget")
}
```

- [ ] **Step 4: Run, verify PASS** + positive test (owner edits OK).
- [ ] **Step 5:** `go build ./...`
- [ ] **Step 6: Commit** — `fix(budget): scope UpdateBudget to owner/involvement (IDOR)`

---

## Task 3: Shared procurement-flow resolver (A6)

**Files:** Create `backend/utils/procurement_flow.go`. Test: `backend/utils/procurement_flow_test.go`.

- [ ] **Step 1: Failing test:**

```go
func TestResolveProcurementFlow(t *testing.T) {
    // PO override wins, normalized:
    require.Equal(t, "payment_first", ResolveProcurementFlow("  Payment_First ", ""))
    // falls back to org default, normalized:
    require.Equal(t, "goods_first", ResolveProcurementFlow("", " Goods_First "))
    // final default:
    require.Equal(t, "goods_first", ResolveProcurementFlow("", ""))
}
```

- [ ] **Step 2: Run, verify FAIL** (undefined).
- [ ] **Step 3: Implement** — pure function (no DB; callers pass the two raw strings):

```go
package utils

import "strings"

// ResolveProcurementFlow returns the normalized effective flow with precedence
// PO override → org default → "goods_first".
func ResolveProcurementFlow(poOverride, orgDefault string) string {
    if v := strings.ToLower(strings.TrimSpace(poOverride)); v != "" {
        return v
    }
    if v := strings.ToLower(strings.TrimSpace(orgDefault)); v != "" {
        return v
    }
    return "goods_first"
}
```

- [ ] **Step 4: Run, verify PASS.**
- [ ] **Step 5: Replace the 4 ad-hoc resolutions** to call `utils.ResolveProcurementFlow(po.ProcurementFlow, orgSettings.ProcurementFlow)` (load orgSettings only when needed): grn.go:260-270, document_extras_handler.go:337-347, payment_voucher.go:295-304, workflow_execution_service.go:~2351. Keep behavior identical for already-normalized inputs.
- [ ] **Step 6:** `go build ./...` + `go test ./utils/ ./handlers/ -run Procurement -v`
- [ ] **Step 7: Commit** — `refactor(procurement): single normalized flow resolver`

---

## Task 4: Extract `validateProcurementPVGate` and reuse in manual PV (A2a)

**Files:** Modify `backend/handlers/payment_voucher.go` (extract from 267-356). Test: payment voucher handler test.

The manual path (payment_voucher.go:267-356) already implements the full gate: PO=APPROVED, live-PV duplicate guard, goods_first GRN APPROVED-or-COMPLETED + received-value cap, PO-total cap. Extract it verbatim into a reusable validator returning a message + HTTP status (0/"" = OK), then have the manual path call it.

- [ ] **Step 1: Failing test** for the extracted function:

```go
func TestValidateProcurementPVGate(t *testing.T) {
    // PO not APPROVED → status 400, message mentions "must be APPROVED".
    // Existing live PV → status 409.
    // goods_first + GRN COMPLETED + amount within received value + PO total → OK ("",0).
    // amount > PO total → 400 "exceeds linked PO".
}
```

- [ ] **Step 2: Run, verify FAIL.**
- [ ] **Step 3: Implement** — new function in payment_voucher.go (signature):

```go
// validateProcurementPVGate enforces the PV-creation rules against a linked PO
// (and GRN for goods_first). Returns ("",0) when valid, else (message, httpStatus).
func validateProcurementPVGate(db *gorm.DB, orgID, linkedPO, linkedGRN string, amount float64) (string, int) {
    // ... body = the logic currently at payment_voucher.go:268-356,
    // using ResolveProcurementFlow(linkedPO.ProcurementFlow, orgSettings.ProcurementFlow),
    // returning fiber.StatusConflict for the duplicate case and
    // fiber.StatusBadRequest for the rest.
}
```

Then replace the inline block in `CreatePaymentVoucher` with:

```go
if req.LinkedPO != "" {
    if msg, code := validateProcurementPVGate(config.DB, tenant.OrganizationID, req.LinkedPO, req.LinkedGRN, req.Amount); code != 0 {
        return c.Status(code).JSON(fiber.Map{"success": false, "message": msg})
    }
}
```

- [ ] **Step 4: Run, verify PASS** — existing manual-PV tests must still pass (regression guard).
- [ ] **Step 5:** `go build ./...`
- [ ] **Step 6: Commit** — `refactor(pv): extract validateProcurementPVGate from manual path`

---

## Task 5: Apply the PV gate to `CreatePaymentVoucherFromPO` (A2b)

**Files:** Modify `backend/handlers/document_extras_handler.go:316-412`. Test: document_extras handler test.

- [ ] **Step 1: Failing tests** — via POST `/payment-vouchers/from-po`:
  - PO not APPROVED → 400.
  - Second PV for same PO → 409.
  - amount > PO total → 400.
  - goods_first GRN in COMPLETED → **succeeds** (regression for the current APPROVED-only reject).
- [ ] **Step 2: Run, verify FAIL** (today: all 200 / the COMPLETED case 400).
- [ ] **Step 3: Implement** — after loading the PO (currency inheritance stays), call the validator and delete the local goods_first-only block (lines 349-366) since the validator subsumes it:

```go
grnDocNum := req.LinkedGRNDocumentNumber
if msg, code := validateProcurementPVGate(config.DB, tenant.OrganizationID, req.PurchaseOrderDocumentNumber, grnDocNum, req.TotalAmount); code != 0 {
    return c.Status(code).JSON(fiber.Map{"success": false, "message": msg})
}
```

Note: validator keys off document numbers; ensure `req.PurchaseOrderDocumentNumber` is populated (the PV stores `LinkedPO: req.PurchaseOrderDocumentNumber`). If only `PurchaseOrderID` is sent, look up `po.DocumentNumber` (already loaded) and pass that.

- [ ] **Step 4: Run, verify PASS.**
- [ ] **Step 5:** `go build ./...`
- [ ] **Step 6: Commit** — `fix(pv): enforce PO/duplicate/amount/GRN gates on from-po create`

---

## Task 6: Apply the PV gate to `AutoCreatePVFromCompletedGRN` (A2c)

**Files:** Modify `backend/services/workflow_execution_service.go:2336-2440`. Test: workflow execution service test.

- [ ] **Step 1: Failing test** — auto-create against a PO that already has a live PV does NOT create a second PV (count stays 1); auto-create amount is capped at PO total.
- [ ] **Step 2: Run, verify FAIL.**
- [ ] **Step 3: Implement** — before building the draft PV, run `validateProcurementPVGate(tx, orgID, po.DocumentNumber, grn.DocumentNumber, amount)`; if it returns non-zero, log + skip (return nil — auto path must not error the whole approval). Set the PV amount to `min(grn received value, po.TotalAmount)` rather than an unbounded value.
- [ ] **Step 4: Run, verify PASS.**
- [ ] **Step 5:** `go build ./...`
- [ ] **Step 6: Commit** — `fix(pv): cap + dedup auto-created PV from completed GRN`

---

## Task 7: PO=APPROVED entry gate for GRN in all flows (A3)

**Files:** Modify `backend/handlers/grn.go:272-278`. Test: grn handler test.

Per decision, payment_first must also require PO=APPROVED before a GRN is created.

- [ ] **Step 1: Failing test** — payment_first GRN create against a PENDING PO → 400 (today: allowed).
- [ ] **Step 2: Run, verify FAIL.**
- [ ] **Step 3: Implement** — change the gate to drop the `effectiveFlow != "payment_first"` exemption:

```go
// The PO must be APPROVED before goods can be received against it, in BOTH flows.
if strings.ToUpper(po.Status) != "APPROVED" {
    return utils.SendBadRequestError(c, fmt.Sprintf(
        "Cannot create GRN: linked PO %s is in %s status and must be APPROVED first.",
        po.DocumentNumber, po.Status))
}
```

- [ ] **Step 4: Run, verify PASS** — goods_first happy path (APPROVED PO) still creates GRN; payment_first with APPROVED PO still creates GRN.
- [ ] **Step 5:** `go build ./...`
- [ ] **Step 6: Commit** — `fix(grn): require PO=APPROVED before GRN create in both flows`

---

## Task 8: `MarkGRNComplete` revalidates PO/PV links (A4)

**Files:** Modify `backend/handlers/grn.go` — extract `revalidateGRNLinks` from `SubmitGRN` (909-938), call from `MarkGRNComplete` (~1316-1321). Test: grn handler test.

- [ ] **Step 1: Failing test** — GRN signed READY, then its linked PO is set CANCELLED; `MarkGRNComplete` returns 400 and does NOT cascade / create a PV.
- [ ] **Step 2: Run, verify FAIL** (today: cascades).
- [ ] **Step 3: Implement** — extract helper:

```go
// revalidateGRNLinks ensures the GRN's linked PO is still APPROVED and the
// linked PV (if any) is still APPROVED/PAID. Returns an error message (or "").
func revalidateGRNLinks(db *gorm.DB, grn *models.GoodsReceivedNote, orgID string) string {
    // body = logic from SubmitGRN:909-938, returning the message string instead of c.Status(...)
}
```

Refactor `SubmitGRN` to use it, and add to `MarkGRNComplete` right after the READY/DRAFT checks:

```go
if msg := revalidateGRNLinks(config.DB, &grn, tenant.OrganizationID); msg != "" {
    return utils.SendBadRequestError(c, msg)
}
```

- [ ] **Step 4: Run, verify PASS** — `SubmitGRN` existing tests still pass.
- [ ] **Step 5:** `go build ./...`
- [ ] **Step 6: Commit** — `fix(grn): revalidate PO/PV links in MarkGRNComplete`

---

## Task 9: Converge `MarkPaidWithPOP` onto the payment-execution task (A5)

**Files:** Modify `backend/handlers/document_extras_handler.go:942-1095`. Test: document_extras handler test.

`MarkPaidWithPOP` stores the PoP file then flips status to PAID directly — no amount-match, leaves the `payment_execution` task open. Add amount-match and route the PAID transition through the same task-completion path as `MarkPaymentVoucherPaid` (604-642).

- [ ] **Step 1: Failing tests:**
  - paidAmount ≠ voucher.Amount (>0.01) → 422 amount_mismatch.
  - On success, the `payment_execution` task for the PV is no longer PENDING/CLAIMED (it was completed), and PV is PAID.
- [ ] **Step 2: Run, verify FAIL.**
- [ ] **Step 3: Implement** — after PoP is persisted: (a) require a `paidAmount` form field and validate `|paidAmount - voucher.Amount| <= 0.01` (reuse the 422 amount_mismatch shape from MarkPaymentVoucherPaid:593-602); (b) find the PENDING/CLAIMED `payment_execution` task, auto-claim if needed, call `ApproveWorkflowTaskWithVersion` (mirror 604-642) instead of the raw `Update("status","PAID")`. Keep the direct_payment branch (no LinkedPO → CascadePVPaidToPO no-ops) working. Remove the non-canonical `super_admin` from the role gate at 953 (A9).
- [ ] **Step 4: Run, verify PASS.**
- [ ] **Step 5:** `go build ./...`
- [ ] **Step 6: Commit** — `fix(pv): converge MarkPaidWithPOP onto payment-execution task + amount match`

---

## Task 10: Robust FULLY_DELIVERED matching (A8)

**Files:** Modify `backend/services/workflow_execution_service.go:2255-2326`. Test: workflow execution service test.

- [ ] **Step 1: Failing test** — PO with two lines whose descriptions differ only by trailing whitespace from the GRN lines: after full receipt, PO delivery_status = FULLY_DELIVERED (today: blocked by description mismatch).
- [ ] **Step 2: Run, verify FAIL.**
- [ ] **Step 3: Implement** — change the aggregation to match GRN lines to PO lines by item `ID` when both present, else by positional index, instead of lowercased description. Keep the per-line `ReceivedQuantity >= Quantity` comparison. Guard the empty-items case (no lines → not FULLY_DELIVERED, no panic).
- [ ] **Step 4: Run, verify PASS** — existing single-line delivery tests still pass.
- [ ] **Step 5:** `go build ./...`
- [ ] **Step 6: Commit** — `fix(grn): match delivery lines by id/index, not description`

---

## Task 11: Server-side total recompute on create (A7)

**Files:** Modify create handlers in `purchase_order.go` (~305-312), `requisition.go` (~377), `payment_voucher.go` (~373). Test: respective handler tests.

- [ ] **Step 1: Failing test** — create a PO whose `totalAmount` (e.g. 999) disagrees with Σ(items) (e.g. 100); stored `total_amount` equals the computed value (100 + tax + delivery), not 999.
- [ ] **Step 2: Run, verify FAIL.**
- [ ] **Step 3: Implement** — before persisting, overwrite the stored total:
  - PO: `total = Σ(item.Quantity*item.UnitPrice) + taxAmount + deliveryCost` (read tax/delivery from metadata as the frontend does).
  - REQ: `total = Σ(item.Quantity*item.UnitPrice)`.
  - PV: `amount = Σ(item.Quantity*item.UnitPrice)` when items present (mirror the existing update-path recompute at payment_voucher.go:651-657).
- [ ] **Step 4: Run, verify PASS** — happy-path creates (consistent totals) unchanged.
- [ ] **Step 5:** `go build ./...`
- [ ] **Step 6: Commit** — `fix(docs): recompute stored totals from line items on create`

---

## Task 12: Low-severity cleanups (A9)

**Files:** `payment_voucher.go` (currency inherit + WithdrawPaymentVoucher), `document_extras_handler.go` (io.ReadFull). Tests where practical.

- [ ] **Step 1:** Manual `CreatePaymentVoucher`: when `req.LinkedPO != ""`, inherit `linkedPO.Currency` into `req.Currency` if empty (mirror from-po:333-335). Add a test.
- [ ] **Step 2:** `WithdrawPaymentVoucher` (payment_voucher.go:1047-1072): use `voucher.CreatedBy` for the creator check instead of scanning ActionHistory; add a test that an auto-created PV (blank PerformedBy) can be withdrawn by its CreatedBy.
- [ ] **Step 3:** `MarkPaidWithPOP` file read: replace `f.Read(buf)` (1006) with `io.ReadFull(f, buf)`.
- [ ] **Step 4:** `go build ./...` + `go test ./handlers/ -v` (targeted).
- [ ] **Step 5: Commit** — `fix(pv): currency inherit, withdraw-by-CreatedBy, io.ReadFull`

---

## Final verification

- [ ] `go build ./...` (from `backend/`) — clean.
- [ ] `go test ./...` — all pass.
- [ ] Frontend untouched in Phase A; no `tsc` needed unless a shared type changed.
- [ ] Re-read the audit HIGH/MED list against committed tasks; confirm each addressed or explicitly deferred (receiver actor gate, LinkedRequisition cosmetic, central state machine — all Phase-B/deferred).
