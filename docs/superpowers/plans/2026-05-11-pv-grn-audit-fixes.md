# PV + GRN Audit Fixes Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

## Audit verification — 2026-05-11

Audited each of the 12 tasks against `main` after the `feature/direct-payment-flow` merge (commit `04c816c`). Confirmed all 12 tasks STILL BROKEN; the direct-payment work did not address any audit issue. Decided scope: execute Phases 1–4 (Tasks 1–11). Phase 5 (Task 12, optional) skipped.

Direct-payment regression strategy: after each phase, re-run the 27-test direct-payment sweep (handlers Payee/MarkPaid/RecoverFromPO/Procurement* + services SubmitRequisitionWithRouting* + utils scope tests). Roll back any change that breaks them.



**Goal:** Close the 10 remaining issues from the PV/GRN end-to-end audit on branch `feature/direct-payment-flow`. Restore document-chain navigation, add GRN→PO delivery cascade, harden submit-time scope and item validation, fix enum drift, and add reciprocal UI links.

**Architecture:** Five independent phases, each shippable on its own. Phase 1 = trivial wins (frontend href fixes + backend switch cases + test cleanup). Phase 2 = security/integrity guards (scope on submit, item-level GRN validation). Phase 3 = the only schema-changing work (PO delivery status + per-item received quantity, plus GRN approval cascade). Phase 4 = UI polish (reciprocal links + GRN edit page). Phase 5 = optional audit-symmetry task for GRN completion.

**Tech Stack:** Go (Fiber + GORM), PostgreSQL, golang-migrate SQL migrations, Next.js App Router + TypeScript, React Query.

---

## File Structure

**Phase 1 — Quick Wins**
- Modify: `frontend/src/components/document-links.tsx` — fix 4 hrefs
- Modify: `backend/services/workflow_execution_service.go:2595-2658` — add GRN cases to priority/due-date helpers
- Modify: `backend/models/models.go:283` — drop "paid" from GRN status comment
- Modify: `backend/tests/unit/grn_handler_test.go:120-150` — replace "RECEIVED" with "COMPLETED" + add cancelled case

**Phase 2 — Security / Integrity**
- Modify: `backend/handlers/payment_voucher.go:736-810` — add scope gate to `SubmitPaymentVoucher`
- Modify: `backend/handlers/grn.go:160-238` — extract + apply item-level validation against PO line items
- Test: `backend/tests/unit/grn_handler_test.go` — new tests for over-receipt + item-mismatch
- Test: `backend/tests/unit/payment_voucher_handler_test.go` — new test for submit-scope gate

**Phase 3 — PO Delivery Cascade**
- Create: `backend/database/migrations/019_po_delivery_tracking.up.sql`
- Create: `backend/database/migrations/019_po_delivery_tracking.down.sql`
- Modify: `backend/models/models.go:157-217` — add `DeliveryStatus` to PO, add `ReceivedQuantity` to POItem
- Modify: `backend/types/documents.go:255-269` — add `ReceivedQuantity` to `POItem`
- Modify: `backend/services/workflow_execution_service.go` — new `cascadeGRNApprovalToPO` helper invoked from terminal-approve path
- Test: `backend/tests/unit/workflow_execution_service_test.go` — three scenarios (single GRN full, partial, multi-GRN total)

**Phase 4 — UI Polish**
- Modify: `frontend/src/app/(private)/(main)/payment-vouchers/[id]/_components/pv-detail-client.tsx` — render `linkedGRN` as `<Link>`
- Modify: `frontend/src/app/(private)/(main)/grn/[id]/_components/grn-detail-client.tsx` — replace text `linkedPV` with `<Link>`
- Modify: `frontend/src/app/(private)/(main)/grn/_components/grn-table.tsx:90-95` — gate Edit button until edit route exists OR build edit page (decision in Task 11)
- (Optional create) `frontend/src/app/(private)/(main)/grn/[id]/edit/page.tsx`

**Phase 5 — Audit Symmetry (Optional)**
- Modify: `backend/models/status.go` — add `TaskKindGRNCompletion = "grn_completion"`
- Modify: `backend/services/workflow_execution_service.go` — mirror PaymentExecution side-effect creation for GRN terminal-approve

---

## Phase 1 — Quick Wins

### Task 1: Fix document-chain navigation hrefs

**Files:**
- Modify: `frontend/src/components/document-links.tsx:97,113,123,139,213,235`

- [ ] **Step 1: Open the file and confirm current broken hrefs**

Run: `grep -n "/purchase-orders\b\|/payment-vouchers\b" frontend/src/components/document-links.tsx`
Expected: lines 97, 113, 123, 139, 213, 235 all link to list pages instead of `/{type}/${id}`.

- [ ] **Step 2: Patch procurement-path PO href (line 97)**

Replace:
```tsx
href: showViewLinks && chain.poId ? `/purchase-orders` : undefined,
```
With:
```tsx
href: showViewLinks && chain.poId ? `/purchase-orders/${chain.poId}` : undefined,
```

- [ ] **Step 3: Patch procurement-path PV href (line 113)**

Replace:
```tsx
href: showViewLinks && chain.pvId ? `/payment-vouchers` : undefined,
```
With:
```tsx
href: showViewLinks && chain.pvId ? `/payment-vouchers/${chain.pvId}` : undefined,
```

- [ ] **Step 4: Patch direct-payment-path PO href (line 123)**

Replace:
```tsx
href: showViewLinks && chain.poId ? `/purchase-orders` : undefined,
```
With:
```tsx
href: showViewLinks && chain.poId ? `/purchase-orders/${chain.poId}` : undefined,
```

- [ ] **Step 5: Patch direct-payment-path PV href (line 139)**

Replace:
```tsx
href: showViewLinks && chain.pvId ? `/payment-vouchers` : undefined,
```
With:
```tsx
href: showViewLinks && chain.pvId ? `/payment-vouchers/${chain.pvId}` : undefined,
```

- [ ] **Step 6: Patch legacy fallback PO href (line 213)**

Replace:
```tsx
<Link href={`/purchase-orders`}>
```
With:
```tsx
<Link href={`/purchase-orders/${purchaseOrder.id}`}>
```

- [ ] **Step 7: Patch legacy fallback PV href (line 235)**

Replace:
```tsx
<Link href={`/payment-vouchers`}>
```
With:
```tsx
<Link href={`/payment-vouchers/${paymentVoucher.id}`}>
```

- [ ] **Step 8: Manually verify in browser**

Run dev server, open a requisition that has `chain.poId` set, click View next to Purchase Order in the chain card. URL should be `/purchase-orders/<uuid>`, page should render PO detail.

- [ ] **Step 9: Commit**

```bash
git add frontend/src/components/document-links.tsx
git commit -m "fix(document-links): route View buttons to detail pages not list pages"
```

---

### Task 2: Add GRN cases to workflow priority/due-date helpers

**Files:**
- Modify: `backend/services/workflow_execution_service.go:2595-2658`

GRN has no `Priority` or `RequiredByDate` field in the model. The current code falls through to the default. We make the design choice explicit (so a future reader does not assume it was forgotten) and key off the linked PO when the PO carries those fields.

- [ ] **Step 1: Add explicit GRN case to `getDocumentPriority`**

In `backend/services/workflow_execution_service.go`, locate the `getDocumentPriority` switch around line 2599. Add a case before the closing brace, BEFORE the final `return defaultPriority`:

```go
case "grn", "goods_received_note":
    var grn models.GoodsReceivedNote
    if err := tx.Where("id = ?", entityID).First(&grn).Error; err == nil {
        if grn.PODocumentNumber != "" {
            var po models.PurchaseOrder
            if err := tx.Where("document_number = ?", grn.PODocumentNumber).First(&po).Error; err == nil {
                if po.Priority != "" {
                    return strings.ToLower(po.Priority)
                }
            }
        }
    }
```

- [ ] **Step 2: Add explicit GRN case to `getDocumentDueDate`**

In the same file locate `getDocumentDueDate` switch around line 2629. Add the case before the closing brace, before `return nil`:

```go
case "grn", "goods_received_note":
    var grn models.GoodsReceivedNote
    if err := tx.Where("id = ?", entityID).First(&grn).Error; err == nil {
        if grn.PODocumentNumber != "" {
            var po models.PurchaseOrder
            if err := tx.Where("document_number = ?", grn.PODocumentNumber).First(&po).Error; err == nil {
                if po.RequiredByDate != nil {
                    return po.RequiredByDate
                }
                if !po.DeliveryDate.IsZero() {
                    return &po.DeliveryDate
                }
            }
        }
    }
```

- [ ] **Step 3: Remove the stale "budget and goods_received_note don't have priority field" comment**

Find the comment on line 2621 and delete it (the new case makes it incorrect).

- [ ] **Step 4: Build verifies**

Run: `cd backend && go build ./...`
Expected: no compile errors.

- [ ] **Step 5: Run existing workflow service tests**

Run: `cd backend && go test ./tests/unit/ -run WorkflowExecutionService -count=1`
Expected: all existing tests pass.

- [ ] **Step 6: Commit**

```bash
git add backend/services/workflow_execution_service.go
git commit -m "feat(workflow): inherit GRN task priority + due-date from linked PO"
```

---

### Task 3: Fix GRN status enum drift in model comment + tests

**Files:**
- Modify: `backend/models/models.go:283`
- Modify: `backend/tests/unit/grn_handler_test.go:124-141`

`models/status.go` is the source of truth: GRN uses `DRAFT, PENDING, APPROVED, REJECTED, REVISION, COMPLETED, CANCELLED`. There is no `PAID` or `RECEIVED` for GRN. The model comment and the test both lie about this.

- [ ] **Step 1: Correct the GRN struct comment**

Open `backend/models/models.go:283`. Replace:
```go
Status            string          `json:"status"` // draft, pending, approved, rejected, paid, completed, cancelled
```
With:
```go
Status            string          `json:"status"` // DRAFT, PENDING, APPROVED, REJECTED, REVISION, COMPLETED, CANCELLED — see models/status.go
```

- [ ] **Step 2: Run failing test for the corrected validation set**

Update `backend/tests/unit/grn_handler_test.go:124-141`. The current `validStatuses` map and table both reference `RECEIVED`. Replace the function body with:

```go
func TestGRNStatusValidation(t *testing.T) {
	validStatuses := map[string]bool{
		"DRAFT":     true,
		"PENDING":   true,
		"APPROVED":  true,
		"REJECTED":  true,
		"REVISION":  true,
		"COMPLETED": true,
		"CANCELLED": true,
	}

	tests := []struct {
		name          string
		status        string
		shouldBeValid bool
	}{
		{"Draft", "DRAFT", true},
		{"Pending", "PENDING", true},
		{"Approved", "APPROVED", true},
		{"Completed", "COMPLETED", true},
		{"Cancelled", "CANCELLED", true},
		{"InvalidPaid", "PAID", false},
		{"InvalidReceived", "RECEIVED", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := validStatuses[tt.status]
			if isValid != tt.shouldBeValid {
				t.Errorf("Expected %v, got %v for status %q", tt.shouldBeValid, isValid, tt.status)
			}
		})
	}
}
```

- [ ] **Step 3: Run the test**

Run: `cd backend && go test ./tests/unit/ -run TestGRNStatusValidation -v`
Expected: PASS.

- [ ] **Step 4: Search for other "RECEIVED" references in GRN context**

Run: `grep -rn "RECEIVED" backend/tests backend/handlers backend/services`
Expected: zero remaining hits in GRN-related code. (If any handler/service mentions "RECEIVED" for GRN, replace with `models.StatusCompleted` and re-run go build.)

- [ ] **Step 5: Commit**

```bash
git add backend/models/models.go backend/tests/unit/grn_handler_test.go
git commit -m "fix(grn): align status enum comment + tests with status.go constants"
```

---

## Phase 2 — Security & Data Integrity

### Task 4: Apply document scope gate to `SubmitPaymentVoucher`

**Files:**
- Modify: `backend/handlers/payment_voucher.go:736-794`
- Test: `backend/tests/unit/payment_voucher_handler_test.go`

`GetPaymentVoucher` (single fetch) already applies `GetDocumentScope`. `SubmitPaymentVoucher` does not — a non-procurement user can submit a PV they would not be able to *list*. We add the same gate.

- [ ] **Step 1: Write the failing test**

Append to `backend/tests/unit/payment_voucher_handler_test.go`:

```go
func TestSubmitPaymentVoucher_NonProcurementWithoutInvolvement_Forbidden(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	org := seedOrganization(t, db, "Org1")
	owner := seedUser(t, db, org.ID, "owner@ex.com", "requester")
	other := seedUser(t, db, org.ID, "other@ex.com", "requester")
	approvedPO := seedApprovedPO(t, db, org.ID, owner.ID)
	pv := seedDraftPV(t, db, org.ID, owner.ID, approvedPO.DocumentNumber)
	wf := seedWorkflow(t, db, org.ID, "payment_voucher")

	app := newTestApp(t, db, other.ID, org.ID, "requester")
	body := []byte(fmt.Sprintf(`{"workflowId":"%s"}`, wf.ID))
	req := httptest.NewRequest("POST", "/payment-vouchers/"+pv.ID+"/submit", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	if resp.StatusCode != fiber.StatusForbidden {
		t.Fatalf("expected 403, got %d", resp.StatusCode)
	}
}
```

(If helpers like `seedDraftPV` / `seedApprovedPO` / `newTestApp` already exist in the file, reuse them with matching signatures. If they do not, mirror the patterns from `payment_vouchers_http_test.go` which was added on this branch.)

- [ ] **Step 2: Run the test and confirm it fails**

Run: `cd backend && go test ./tests/unit/ -run TestSubmitPaymentVoucher_NonProcurementWithoutInvolvement_Forbidden -v`
Expected: FAIL — current handler returns 200 OK because scope is not enforced.

- [ ] **Step 3: Add scope check to the handler**

In `backend/handlers/payment_voucher.go`, inside `SubmitPaymentVoucher`, immediately after the existing voucher load at line 766 (`if err := config.DB.Where("id = ? AND organization_id = ?"...)`), insert:

```go
// Scope gate: requester must either own the PV or be involved via workflow tasks,
// unless they hold a privileged role (admin/finance/etc).
userRole := strings.ToLower(c.Locals("userRole").(string))
scope := utils.GetDocumentScope(config.DB, userID, userRole, organizationID)
if !scope.CanViewAll && !scope.IsProcurement {
    var count int64
    config.DB.Model(&models.PaymentVoucher{}).
        Where("id = ? AND organization_id = ? AND created_by = ?", id, organizationID, userID).
        Count(&count)
    if count == 0 {
        // Allow if user has a pending workflow task on this PV.
        var taskCount int64
        config.DB.Table("workflow_tasks").
            Where("entity_id = ? AND entity_type = ? AND organization_id = ?", id, "payment_voucher", organizationID).
            Where("assigned_user_id = ? OR claimed_by = ?", userID, userID).
            Count(&taskCount)
        if taskCount == 0 {
            return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
                "success": false,
                "message": "You do not have permission to submit this payment voucher",
            })
        }
    }
}
```

Make sure `utils` and `strings` are already imported in this file (they are).

- [ ] **Step 4: Run the test again — should now pass**

Run: `cd backend && go test ./tests/unit/ -run TestSubmitPaymentVoucher_NonProcurementWithoutInvolvement_Forbidden -v`
Expected: PASS.

- [ ] **Step 5: Run the full PV test suite to ensure no regression**

Run: `cd backend && go test ./tests/unit/ -run PaymentVoucher -count=1`
Expected: all pass.

- [ ] **Step 6: Commit**

```bash
git add backend/handlers/payment_voucher.go backend/tests/unit/payment_voucher_handler_test.go
git commit -m "fix(pv): enforce document scope on SubmitPaymentVoucher"
```

---

### Task 5: GRN item-level validation against PO line items

**Files:**
- Modify: `backend/handlers/grn.go:160-360`
- Test: `backend/tests/unit/grn_handler_test.go`

Currently `QuantityReceived` is accepted blindly. Add two guards: (a) `QuantityReceived <= QuantityOrdered` for the *PO line*, not the request value, and (b) sum across all non-cancelled GRNs for the same PO must not exceed PO line quantity. Until Phase 3 lands, we compute the running total from existing GRNs on the fly.

- [ ] **Step 1: Write failing test — over-receipt on single GRN**

Append to `backend/tests/unit/grn_handler_test.go`:

```go
func TestCreateGRN_OverReceiptVsPO_Rejected(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	org := seedOrganization(t, db, "Org1")
	user := seedUser(t, db, org.ID, "user@ex.com", "requester")
	po := seedApprovedPOWithItems(t, db, org.ID, user.ID, []types.POItem{
		{Description: "Widget", Quantity: 10, UnitPrice: 5.0, Amount: 50.0},
	})

	body := map[string]interface{}{
		"poDocumentNumber": po.DocumentNumber,
		"receivedBy":       user.ID,
		"items": []map[string]interface{}{
			{
				"description":      "Widget",
				"quantityOrdered":  10,
				"quantityReceived": 15, // over-receipt
				"condition":        "good",
			},
		},
	}
	resp := postJSON(t, db, user.ID, org.ID, "requester", "/grns", body)

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}
```

- [ ] **Step 2: Run the test and confirm it fails**

Run: `cd backend && go test ./tests/unit/ -run TestCreateGRN_OverReceiptVsPO_Rejected -v`
Expected: FAIL — handler currently returns 201.

- [ ] **Step 3: Write failing test — split GRN total exceeds PO line**

Add a second test in the same file:

```go
func TestCreateGRN_SecondGRNExceedsPOTotal_Rejected(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	org := seedOrganization(t, db, "Org1")
	user := seedUser(t, db, org.ID, "user@ex.com", "requester")
	po := seedApprovedPOWithItems(t, db, org.ID, user.ID, []types.POItem{
		{Description: "Widget", Quantity: 10, UnitPrice: 5.0, Amount: 50.0},
	})

	// First GRN receives 6 (approved, non-cancelled).
	first := seedApprovedGRN(t, db, org.ID, user.ID, po.DocumentNumber, []types.GRNItem{
		{Description: "Widget", QuantityOrdered: 10, QuantityReceived: 6, Condition: "good"},
	})
	_ = first

	// Second GRN would push total to 12 > ordered 10.
	body := map[string]interface{}{
		"poDocumentNumber": po.DocumentNumber,
		"receivedBy":       user.ID,
		"items": []map[string]interface{}{
			{
				"description":      "Widget",
				"quantityOrdered":  10,
				"quantityReceived": 6,
				"condition":        "good",
			},
		},
	}
	resp := postJSON(t, db, user.ID, org.ID, "requester", "/grns", body)
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}
```

- [ ] **Step 4: Run the second test and confirm it fails**

Run: `cd backend && go test ./tests/unit/ -run TestCreateGRN_SecondGRNExceedsPOTotal_Rejected -v`
Expected: FAIL (also the existing one-to-one guard rejects on a different code path — adjust the test to cancel the first GRN before creating the second if the one-to-one rule blocks it. Make sure the assertion specifically checks for the over-quantity error text from the message we add in Step 5).

- [ ] **Step 5: Implement validation in `CreateGRN`**

In `backend/handlers/grn.go`, immediately after the PO existence check completes (after line 207, before the "Resolve effective procurement flow" comment), insert:

```go
// Build a description→ordered-quantity map from PO items.
poItems := po.Items.Data()
poByDesc := make(map[string]int, len(poItems))
for _, it := range poItems {
    poByDesc[strings.TrimSpace(strings.ToLower(it.Description))] += it.Quantity
}

// Validate every GRN line: description must exist on PO, and per-line received
// must not exceed ordered (single-GRN guard).
for _, ln := range req.Items {
    key := strings.TrimSpace(strings.ToLower(ln.Description))
    ordered, ok := poByDesc[key]
    if !ok {
        return utils.SendBadRequestError(c, fmt.Sprintf(
            "GRN item %q does not match any line on PO %s", ln.Description, po.DocumentNumber))
    }
    if ln.QuantityReceived <= 0 {
        return utils.SendBadRequestError(c, fmt.Sprintf(
            "GRN item %q must have quantityReceived > 0", ln.Description))
    }
    if ln.QuantityReceived > ordered {
        return utils.SendBadRequestError(c, fmt.Sprintf(
            "GRN item %q: quantityReceived %d exceeds ordered %d on PO %s",
            ln.Description, ln.QuantityReceived, ordered, po.DocumentNumber))
    }
}

// Cross-GRN guard: sum received across non-cancelled GRNs for this PO + this request
// must not exceed PO ordered per item.
var existingGRNs []models.GoodsReceivedNote
config.DB.Where("po_document_number = ? AND organization_id = ? AND UPPER(status) != ?",
    req.PODocumentNumber, tenant.OrganizationID, "CANCELLED").
    Find(&existingGRNs)
receivedByDesc := make(map[string]int)
for _, g := range existingGRNs {
    for _, it := range g.Items.Data() {
        receivedByDesc[strings.TrimSpace(strings.ToLower(it.Description))] += it.QuantityReceived
    }
}
for _, ln := range req.Items {
    key := strings.TrimSpace(strings.ToLower(ln.Description))
    total := receivedByDesc[key] + ln.QuantityReceived
    if total > poByDesc[key] {
        return utils.SendBadRequestError(c, fmt.Sprintf(
            "GRN item %q: total received across GRNs would be %d, exceeds PO %s ordered %d",
            ln.Description, total, po.DocumentNumber, poByDesc[key]))
    }
}
```

- [ ] **Step 6: Run both new tests — should pass**

Run: `cd backend && go test ./tests/unit/ -run "TestCreateGRN_OverReceiptVsPO_Rejected|TestCreateGRN_SecondGRNExceedsPOTotal_Rejected" -v`
Expected: PASS.

- [ ] **Step 7: Run the full GRN handler test suite**

Run: `cd backend && go test ./tests/unit/ -run GRN -count=1`
Expected: all pass. If existing tests broke because they used loose item descriptions, fix the fixtures to match PO line descriptions exactly.

- [ ] **Step 8: Commit**

```bash
git add backend/handlers/grn.go backend/tests/unit/grn_handler_test.go
git commit -m "feat(grn): validate items match PO lines and block over-receipt"
```

---

## Phase 3 — PO Delivery Cascade

### Task 6: Migration — PO delivery tracking columns

**Files:**
- Create: `backend/database/migrations/019_po_delivery_tracking.up.sql`
- Create: `backend/database/migrations/019_po_delivery_tracking.down.sql`

- [ ] **Step 1: Write the up migration**

Create `backend/database/migrations/019_po_delivery_tracking.up.sql`:

```sql
-- Add delivery tracking to purchase orders.
-- delivery_status: NOT_DELIVERED | PARTIALLY_DELIVERED | FULLY_DELIVERED
-- It is independent of the workflow status (draft/pending/approved/...).

ALTER TABLE purchase_orders
    ADD COLUMN IF NOT EXISTS delivery_status TEXT NOT NULL DEFAULT 'NOT_DELIVERED';

CREATE INDEX IF NOT EXISTS idx_po_delivery_status
    ON purchase_orders (organization_id, delivery_status);

-- Per-item received quantity lives in the items JSONB blob — no schema change
-- needed there. Backfill: anything already FULFILLED keeps NOT_DELIVERED
-- (the legacy fulfilled flag is unrelated to physical receipt).
```

- [ ] **Step 2: Write the down migration**

Create `backend/database/migrations/019_po_delivery_tracking.down.sql`:

```sql
DROP INDEX IF EXISTS idx_po_delivery_status;
ALTER TABLE purchase_orders DROP COLUMN IF EXISTS delivery_status;
```

- [ ] **Step 3: Run the migration locally**

Run: `cd backend && make migrate-up` (or whatever the project's migration command is — check `Makefile`).
Expected: migration `019` applied, no errors. Verify with `psql -c "\d purchase_orders"` that `delivery_status` exists.

- [ ] **Step 4: Roll back and re-apply to test reversibility**

Run: `cd backend && make migrate-down && make migrate-up`
Expected: clean down + up.

- [ ] **Step 5: Commit**

```bash
git add backend/database/migrations/019_po_delivery_tracking.up.sql backend/database/migrations/019_po_delivery_tracking.down.sql
git commit -m "feat(db): add delivery_status to purchase_orders"
```

---

### Task 7: Model updates — `PurchaseOrder.DeliveryStatus` + `POItem.ReceivedQuantity`

**Files:**
- Modify: `backend/models/models.go:157-217` — add field to PurchaseOrder
- Modify: `backend/types/documents.go:255-269` — add field to POItem
- Modify: `backend/models/status.go` — add delivery-status constants

- [ ] **Step 1: Add delivery-status constants**

In `backend/models/status.go`, append after `StatusCancelled`:

```go
// PO delivery_status: tracks physical receipt independent of workflow status.
const (
    DeliveryStatusNotDelivered       = "NOT_DELIVERED"
    DeliveryStatusPartiallyDelivered = "PARTIALLY_DELIVERED"
    DeliveryStatusFullyDelivered     = "FULLY_DELIVERED"
)
```

- [ ] **Step 2: Add `DeliveryStatus` to the PO model**

In `backend/models/models.go`, inside the `PurchaseOrder` struct (after the existing `RoutingType` field around line 190), add:

```go
// Physical delivery tracking — independent of workflow Status.
DeliveryStatus string `gorm:"column:delivery_status;type:text;not null;default:'NOT_DELIVERED';index" json:"deliveryStatus"`
```

- [ ] **Step 3: Add `ReceivedQuantity` to `POItem`**

In `backend/types/documents.go`, inside the `POItem` struct (after the existing `Quantity` field on line 257), add:

```go
ReceivedQuantity int `json:"receivedQuantity,omitempty"` // Running total received across all non-cancelled GRNs
```

- [ ] **Step 4: Build verifies**

Run: `cd backend && go build ./...`
Expected: no compile errors.

- [ ] **Step 5: Commit**

```bash
git add backend/models/models.go backend/models/status.go backend/types/documents.go
git commit -m "feat(po): add DeliveryStatus + per-item ReceivedQuantity tracking"
```

---

### Task 8: Cascade GRN approval to PO delivery status + received quantities

**Files:**
- Modify: `backend/services/workflow_execution_service.go` — add `cascadeGRNApprovalToPO` and call from terminal-approve
- Test: `backend/tests/unit/workflow_execution_service_test.go`

When a GRN reaches terminal approve, recompute the linked PO's per-item received quantities and overall delivery_status from the set of non-cancelled, approved GRNs.

- [ ] **Step 1: Write failing test — full single GRN sets PO to FULLY_DELIVERED**

Append to `backend/tests/unit/workflow_execution_service_test.go`:

```go
func TestGRNApprovalCascadesToPO_SingleFullGRN_FullyDelivered(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	org := seedOrganization(t, db, "Org1")
	user := seedUser(t, db, org.ID, "u@ex.com", "approver")
	po := seedApprovedPOWithItems(t, db, org.ID, user.ID, []types.POItem{
		{Description: "Widget", Quantity: 10, UnitPrice: 5.0},
	})
	grn := seedPendingGRN(t, db, org.ID, user.ID, po.DocumentNumber, []types.GRNItem{
		{Description: "Widget", QuantityOrdered: 10, QuantityReceived: 10, Condition: "good"},
	})
	wf := seedSingleStageWorkflow(t, db, org.ID, "grn")
	assignment := assignWorkflow(t, db, org.ID, grn.ID, "grn", wf.ID, user.ID)

	svc := services.NewWorkflowExecutionService(db, nil)
	_, err := svc.ApproveWorkflowTaskWithVersion(context.Background(), assignment.ID, user.ID, "approve", "ok", "", 1)
	if err != nil {
		t.Fatalf("approve failed: %v", err)
	}

	var updated models.PurchaseOrder
	db.Where("id = ?", po.ID).First(&updated)
	if updated.DeliveryStatus != models.DeliveryStatusFullyDelivered {
		t.Fatalf("expected FULLY_DELIVERED, got %s", updated.DeliveryStatus)
	}
	items := updated.Items.Data()
	if items[0].ReceivedQuantity != 10 {
		t.Fatalf("expected items[0].ReceivedQuantity=10, got %d", items[0].ReceivedQuantity)
	}
}
```

- [ ] **Step 2: Write failing test — partial GRN sets PO to PARTIALLY_DELIVERED**

```go
func TestGRNApprovalCascadesToPO_PartialGRN_PartiallyDelivered(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	org := seedOrganization(t, db, "Org1")
	user := seedUser(t, db, org.ID, "u@ex.com", "approver")
	po := seedApprovedPOWithItems(t, db, org.ID, user.ID, []types.POItem{
		{Description: "Widget", Quantity: 10, UnitPrice: 5.0},
	})
	grn := seedPendingGRN(t, db, org.ID, user.ID, po.DocumentNumber, []types.GRNItem{
		{Description: "Widget", QuantityOrdered: 10, QuantityReceived: 4, Condition: "good"},
	})
	wf := seedSingleStageWorkflow(t, db, org.ID, "grn")
	assignment := assignWorkflow(t, db, org.ID, grn.ID, "grn", wf.ID, user.ID)

	svc := services.NewWorkflowExecutionService(db, nil)
	_, err := svc.ApproveWorkflowTaskWithVersion(context.Background(), assignment.ID, user.ID, "approve", "ok", "", 1)
	if err != nil {
		t.Fatalf("approve failed: %v", err)
	}

	var updated models.PurchaseOrder
	db.Where("id = ?", po.ID).First(&updated)
	if updated.DeliveryStatus != models.DeliveryStatusPartiallyDelivered {
		t.Fatalf("expected PARTIALLY_DELIVERED, got %s", updated.DeliveryStatus)
	}
	items := updated.Items.Data()
	if items[0].ReceivedQuantity != 4 {
		t.Fatalf("expected items[0].ReceivedQuantity=4, got %d", items[0].ReceivedQuantity)
	}
}
```

- [ ] **Step 3: Write failing test — split GRNs sum to fully delivered**

```go
func TestGRNApprovalCascadesToPO_TwoGRNsTotalToFull_FullyDelivered(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	org := seedOrganization(t, db, "Org1")
	user := seedUser(t, db, org.ID, "u@ex.com", "approver")
	po := seedApprovedPOWithItems(t, db, org.ID, user.ID, []types.POItem{
		{Description: "Widget", Quantity: 10, UnitPrice: 5.0},
	})

	// First GRN already approved (qty 6) — but the one-to-one index blocks a second
	// non-cancelled GRN per PO. So cancel the first before creating the second.
	first := seedApprovedGRN(t, db, org.ID, user.ID, po.DocumentNumber, []types.GRNItem{
		{Description: "Widget", QuantityOrdered: 10, QuantityReceived: 6, Condition: "good"},
	})
	_ = first
	// Skip this test if migration 013 still blocks multi-GRN; document it instead.
	t.Skip("Multi-GRN-per-PO requires relaxing migration 013 unique index — out of scope for this plan.")
}
```

> Note: the existing partial unique index `idx_po_only_one_active_grn` enforces one non-cancelled GRN per PO. Multi-GRN-per-PO is a real-world need (partial deliveries) but a separate design decision. This third test is intentionally skipped with a clear note; address in a follow-up.

- [ ] **Step 4: Run the three tests and confirm the first two fail**

Run: `cd backend && go test ./tests/unit/ -run TestGRNApprovalCascadesToPO -v`
Expected: first two FAIL (no cascade yet), third is SKIP.

- [ ] **Step 5: Implement `cascadeGRNApprovalToPO`**

In `backend/services/workflow_execution_service.go`, add this method (place near `updateDocumentStatusScopedAs`):

```go
// cascadeGRNApprovalToPO recomputes the parent PO's delivery_status and per-item
// received quantities from the set of non-cancelled GRNs for that PO. Called from
// the terminal-approve path when entity_type == "grn".
func (s *WorkflowExecutionService) cascadeGRNApprovalToPO(tx *gorm.DB, grnID string) error {
    var grn models.GoodsReceivedNote
    if err := tx.Where("id = ?", grnID).First(&grn).Error; err != nil {
        return fmt.Errorf("cascade: load GRN: %w", err)
    }
    if grn.PODocumentNumber == "" {
        return nil // payment-first GRN with no PO link? still recompute below.
    }

    var po models.PurchaseOrder
    if err := tx.Where("document_number = ? AND organization_id = ?",
        grn.PODocumentNumber, grn.OrganizationID).First(&po).Error; err != nil {
        return fmt.Errorf("cascade: load PO: %w", err)
    }

    // Sum received per description across all non-cancelled GRNs (the one we just
    // approved is already APPROVED at this point, so include it).
    var grns []models.GoodsReceivedNote
    if err := tx.Where("po_document_number = ? AND organization_id = ? AND UPPER(status) != ?",
        po.DocumentNumber, po.OrganizationID, "CANCELLED").Find(&grns).Error; err != nil {
        return fmt.Errorf("cascade: list GRNs: %w", err)
    }
    receivedByDesc := make(map[string]int)
    for _, g := range grns {
        for _, it := range g.Items.Data() {
            key := strings.TrimSpace(strings.ToLower(it.Description))
            receivedByDesc[key] += it.QuantityReceived
        }
    }

    // Apply to PO items.
    items := po.Items.Data()
    allFull, anyReceived := true, false
    for i := range items {
        key := strings.TrimSpace(strings.ToLower(items[i].Description))
        items[i].ReceivedQuantity = receivedByDesc[key]
        if items[i].ReceivedQuantity > 0 {
            anyReceived = true
        }
        if items[i].ReceivedQuantity < items[i].Quantity {
            allFull = false
        }
    }

    newDeliveryStatus := models.DeliveryStatusNotDelivered
    switch {
    case allFull && anyReceived:
        newDeliveryStatus = models.DeliveryStatusFullyDelivered
    case anyReceived:
        newDeliveryStatus = models.DeliveryStatusPartiallyDelivered
    }

    return tx.Model(&models.PurchaseOrder{}).
        Where("id = ?", po.ID).
        Updates(map[string]interface{}{
            "items":           datatypes.NewJSONType(items),
            "delivery_status": newDeliveryStatus,
            "updated_at":      time.Now(),
        }).Error
}
```

- [ ] **Step 6: Wire the cascade into the terminal-approve path**

In the same file, find the block around line 1033 where `updateDocumentStatusScopedAs` is invoked for the terminal approve (entity status → APPROVED). Immediately after that call succeeds, add:

```go
if strings.EqualFold(assignment.EntityType, "grn") {
    if err := s.cascadeGRNApprovalToPO(tx, assignment.EntityID); err != nil {
        return fmt.Errorf("post-approval GRN cascade: %w", err)
    }
}
```

(Place it next to the existing `if strings.EqualFold(assignment.EntityType, "payment_voucher")` block that creates the payment-execution task — same pattern.)

- [ ] **Step 7: Run all three new tests**

Run: `cd backend && go test ./tests/unit/ -run TestGRNApprovalCascadesToPO -v`
Expected: tests 1 and 2 PASS, test 3 SKIP.

- [ ] **Step 8: Run the full workflow-service test suite**

Run: `cd backend && go test ./tests/unit/ -run WorkflowExecutionService -count=1`
Expected: all pass.

- [ ] **Step 9: Commit**

```bash
git add backend/services/workflow_execution_service.go backend/tests/unit/workflow_execution_service_test.go
git commit -m "feat(grn): cascade approval to PO delivery_status + received quantities"
```

---

## Phase 4 — UI Polish

### Task 9: Render `linkedGRN` as a navigable link on PV detail

**Files:**
- Modify: `frontend/src/app/(private)/(main)/payment-vouchers/[id]/_components/pv-detail-client.tsx`

- [ ] **Step 1: Locate the "Linked Documents" area in the PV detail file**

Run: `grep -n "linkedPO\|LinkedPO\|linkedGRN\|DocumentLinks" frontend/src/app/\(private\)/\(main\)/payment-vouchers/\[id\]/_components/pv-detail-client.tsx`

Identify the JSX section that renders the linked-document badges next to the `DocumentLinks` component. If `linkedGRN` is not rendered at all, add it next to where `linkedPO` is shown.

- [ ] **Step 2: Add the linked-GRN row**

In the same JSX block where `linkedPO` is rendered, add (use the same styling primitives the existing row uses; this is the canonical pattern):

```tsx
{voucher.linkedGRN && (
  <div className="flex items-center justify-between bg-background p-3 rounded border">
    <div>
      <p className="text-sm text-muted-foreground">Linked Goods Received Note</p>
      <p className="font-medium font-mono">{voucher.linkedGRN}</p>
    </div>
    {voucher.linkedGRNId && (
      <Link href={`/grn/${voucher.linkedGRNId}`}>
        <Button variant="outline" size="sm">View GRN</Button>
      </Link>
    )}
  </div>
)}
```

> If the API does not return `linkedGRNId` (only the document number), add a server action `getGRNByDocumentNumber(documentNumber: string)` and use `useQuery` to resolve the id. Alternatively, accept the document-number-only display without a link until the API exposes the id.

- [ ] **Step 3: Manually verify**

Run the dev server, open a PV that has `linkedGRN` populated (a goods-first flow PV), confirm the row renders and the View button navigates to `/grn/<id>`.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/app/\(private\)/\(main\)/payment-vouchers/\[id\]/_components/pv-detail-client.tsx
git commit -m "feat(pv-detail): render linked GRN with View navigation"
```

---

### Task 10: Make `linkedPV` clickable on GRN detail

**Files:**
- Modify: `frontend/src/app/(private)/(main)/grn/[id]/_components/grn-detail-client.tsx:342-348`

- [ ] **Step 1: Find the current text-only render**

Run: `grep -n "linkedPV" frontend/src/app/\(private\)/\(main\)/grn/\[id\]/_components/grn-detail-client.tsx`
Expected: a span around line 342-348 that renders `grn.linkedPV` inside `<span className="font-mono">…</span>` with no navigation.

- [ ] **Step 2: Replace the text span with a Link**

Replace the existing render block with (use the same key/value styling as the linkedPO row in this same file — confirm pattern first):

```tsx
{grn.linkedPV && (
  <DetailField
    label="Linked Payment Voucher"
    value={
      grn.linkedPVId ? (
        <Link
          href={`/payment-vouchers/${grn.linkedPVId}`}
          className="font-mono text-primary hover:underline"
        >
          {grn.linkedPV}
        </Link>
      ) : (
        <span className="font-mono">{grn.linkedPV}</span>
      )
    }
  />
)}
```

> Same caveat as Task 9 about `linkedPVId` — if only the document number is available client-side, fetch the id via a server action or accept text-only display.

- [ ] **Step 3: Manually verify**

Run the dev server, open a GRN that has `linkedPV` populated (a payment-first flow GRN), confirm the row renders and click navigates to `/payment-vouchers/<id>`.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/app/\(private\)/\(main\)/grn/\[id\]/_components/grn-detail-client.tsx
git commit -m "feat(grn-detail): make linked PV a navigable link"
```

---

### Task 11: GRN edit page — decide and act

**Files:**
- Either create: `frontend/src/app/(private)/(main)/grn/[id]/edit/page.tsx`
- Or modify: `frontend/src/app/(private)/(main)/grn/_components/grn-table.tsx:90-95`

The current GRN table renders an Edit button for non-approved GRNs that links to `/grn/${grn.id}/edit`. That route does not exist, so it 404s. Two options.

- [ ] **Step 1: Confirm the route is missing**

Run: `ls frontend/src/app/\(private\)/\(main\)/grn/\[id\]/`
Expected: only `page.tsx` and `_components/` — no `edit/` directory.

- [ ] **Step 2: Pick a path**

**Option A (preferred — quick)**: Remove the Edit menu item; users can edit a draft GRN via the detail page (which already has an Edit toggle if present, or we add one). Replace the dropdown menu item with a router push to the detail page where edit is supported.

In `grn-table.tsx:90-95`, replace:
```tsx
{grn.status?.toUpperCase() !== "APPROVED" && canModify && (
  <DropdownMenuItem onClick={() => router.push(`/grn/${grn.id}/edit`)}>
    <Pencil className="mr-2 h-4 w-4" />
    Edit
  </DropdownMenuItem>
)}
```
With:
```tsx
{grn.status?.toUpperCase() === "DRAFT" && canModify && (
  <DropdownMenuItem onClick={() => router.push(`/grn/${grn.id}?mode=edit`)}>
    <Pencil className="mr-2 h-4 w-4" />
    Edit
  </DropdownMenuItem>
)}
```
Then in `grn-detail-client.tsx`, read `mode=edit` from the URL with `useSearchParams()` and open the existing edit dialog/form on mount.

**Option B**: Build a standalone `edit/page.tsx` that mirrors `create-grn-dialog.tsx`. This is real work; defer to a separate plan if chosen.

- [ ] **Step 3: Implement Option A**

Apply the table change above. Then in `grn-detail-client.tsx` add at the top of the component:

```tsx
const searchParams = useSearchParams();
useEffect(() => {
  if (searchParams.get("mode") === "edit" && grn?.status?.toUpperCase() === "DRAFT") {
    setEditOpen(true);
  }
}, [searchParams, grn?.status]);
```

(Reuse whatever `setEditOpen` state the detail page already exposes for its existing edit flow. If none exists, surface the create dialog in edit mode by passing the current GRN as initial data — out of scope for Option A; default to a flash banner explaining edit-on-detail is not yet wired and route to detail.)

- [ ] **Step 4: Manually verify**

Dev server, GRN list, click Edit on a draft GRN. Should land on `/grn/<id>?mode=edit` and open the edit experience.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/app/\(private\)/\(main\)/grn/_components/grn-table.tsx \
        frontend/src/app/\(private\)/\(main\)/grn/\[id\]/_components/grn-detail-client.tsx
git commit -m "fix(grn): route Edit button to detail page with mode=edit (no 404)"
```

---

## Phase 5 — Audit Symmetry (Optional)

### Task 12: GRN completion task for parity with PV's payment_execution

**Files:**
- Modify: `backend/models/status.go`
- Modify: `backend/services/workflow_execution_service.go`

PV terminal-approve creates a `payment_execution` task so the PAID transition has a claim + signature + actor audit trail. GRN terminal-approve currently sets status to APPROVED with no equivalent. If the business wants a "physical confirmation of receipt" step (warehouse confirms goods checked-in), add a `grn_completion` task.

**Decision required**: This is optional. Skip if business does not need a post-approval completion step. Document why in a comment if skipping.

- [ ] **Step 1: Add the task-kind constant**

In `backend/models/status.go`, append to the WorkflowTask.Kind block:

```go
const (
    TaskKindApproval         = "approval"
    TaskKindPaymentExecution = "payment_execution"
    TaskKindGRNCompletion    = "grn_completion"
)
```

- [ ] **Step 2: Mirror `createPaymentExecutionTask` for GRN**

In `workflow_execution_service.go`, copy the existing `createPaymentExecutionTask` method, rename to `createGRNCompletionTask`, swap kind constant, swap entity-type label, and adjust the assignment-resolver SQL if needed. Wire it from the terminal-approve branch right next to where `cascadeGRNApprovalToPO` was added.

- [ ] **Step 3: Add a test that GRN approval creates a grn_completion task**

(Pattern: copy a `TestApproveWorkflowTask_TerminalApprove_CreatesPaymentExecutionTask`-style test.)

- [ ] **Step 4: Run + commit**

```bash
go test ./tests/unit/ -run GRNCompletion -v
git add backend/models/status.go backend/services/workflow_execution_service.go backend/tests/unit/workflow_execution_service_test.go
git commit -m "feat(grn): post-approval completion task mirrors payment_execution"
```

---

## Self-Review Notes

- **Spec coverage**: Each of the 10 remaining audit issues maps to a task — Issue 5→Task 1, Issue 9→Task 2, Issue 8→Task 3, Issue 3→Task 4, Issue 2→Task 5, Issue 1→Tasks 6+7+8, Issue 6→Task 11, Issue 7→Tasks 9+10, Issue 10→Task 12.
- **Placeholder scan**: every step has either code or an explicit command. The two known-ambiguous spots (`linkedGRNId` / `linkedPVId` API fields) are flagged inline with fallback behavior, not left blank.
- **Type consistency**: `DeliveryStatus*` constants defined in Task 7 are referenced verbatim in Task 8. `cascadeGRNApprovalToPO` matches its declaration. `ReceivedQuantity` field name consistent across model + service + test.
- **Known scope-out**: Multi-GRN-per-PO physical receipt requires relaxing the unique partial index from migration 013. That is a deliberate separate decision and is called out in Task 8 Step 3 with a `t.Skip` rather than silently handled.

---

## Execution Handoff

Plan complete and saved to `docs/superpowers/plans/2026-05-11-pv-grn-audit-fixes.md`. Two execution options:

**1. Subagent-Driven (recommended)** — I dispatch a fresh subagent per task, review between tasks, fast iteration.

**2. Inline Execution** — Execute tasks in this session using executing-plans, batch execution with checkpoints.

Which approach?
