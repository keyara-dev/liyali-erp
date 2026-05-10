# Persist Manual Vendor Name Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Persist manually-typed vendor names on Purchase Orders, Payment Vouchers, and Requisitions so PO submit-for-approval succeeds when the selected quotation has no system vendor_id.

**Architecture:** Add nullable `vendor_name` / `preferred_vendor_name` columns via SQL migration 017. Strip `gorm:"-"` tags from two model fields so GORM persists them. Wire request → model assignment in three Create handlers and three Update handlers. Update three response builders to fall back to the stored column when the `Vendor` GORM relation is nil. The canonical `Vendor.Name` always wins when the relation is loaded.

**Tech Stack:** Go 1.x · Fiber · GORM · PostgreSQL · SQLite (in-memory tests). Frontend untouched.

**Spec:** [docs/superpowers/specs/2026-05-10-persist-manual-vendor-name-design.md](../specs/2026-05-10-persist-manual-vendor-name-design.md)

---

## File map

**Backend — created**
- `backend/database/migrations/017_persist_vendor_name.up.sql`
- `backend/database/migrations/017_persist_vendor_name.down.sql`

**Backend — modified**
- `backend/models/models.go` — strip `gorm:"-"` from `Requisition.PreferredVendorName` (line 67) and `PurchaseOrder.VendorName` (line 146)
- `backend/types/documents.go` — add `VendorName` to `CreatePurchaseOrderRequest`, `CreatePaymentVoucherRequest`, `UpdatePaymentVoucherRequest`; add `PreferredVendorName` to `CreateRequisitionRequest`, `UpdateRequisitionRequest`
- `backend/handlers/purchase_order.go` — Create/Update handlers + `modelToPurchaseOrderResponse`
- `backend/handlers/payment_voucher.go` — Create/Update handlers + `modelToPaymentVoucherResponse`
- `backend/handlers/requisition.go` — Create/Update handlers + requisition response builder

**Backend — tests added**
- `backend/handlers/purchase_orders_http_test.go` — `TestPurchaseOrder_PersistManualVendorName`
- `backend/handlers/payment_voucher_http_test.go` (or sibling) — `TestPaymentVoucher_PersistManualVendorName`
- `backend/handlers/requisition_http_test.go` (or sibling) — `TestRequisition_PersistManualPreferredVendorName`

**Frontend** — no changes.

---

## Task 1: SQL migration 017

**Files:**
- Create: `backend/database/migrations/017_persist_vendor_name.up.sql`
- Create: `backend/database/migrations/017_persist_vendor_name.down.sql`

- [ ] **Step 1: Write the up migration**

Create `backend/database/migrations/017_persist_vendor_name.up.sql`:

```sql
-- ============================================================================
-- PERSIST MANUAL VENDOR NAME
-- Migration: 017_persist_vendor_name
-- Adds a nullable vendor_name / preferred_vendor_name column to documents
-- so quotations with manually-typed (non-system) vendors can be selected
-- as the supplier without losing the vendor name on save.
-- ============================================================================

ALTER TABLE purchase_orders   ADD COLUMN IF NOT EXISTS vendor_name           VARCHAR(255);
ALTER TABLE payment_vouchers  ADD COLUMN IF NOT EXISTS vendor_name           VARCHAR(255);
ALTER TABLE requisitions      ADD COLUMN IF NOT EXISTS preferred_vendor_name VARCHAR(255);
```

- [ ] **Step 2: Write the down migration**

Create `backend/database/migrations/017_persist_vendor_name.down.sql`:

```sql
ALTER TABLE purchase_orders   DROP COLUMN IF EXISTS vendor_name;
ALTER TABLE payment_vouchers  DROP COLUMN IF EXISTS vendor_name;
ALTER TABLE requisitions      DROP COLUMN IF EXISTS preferred_vendor_name;
```

- [ ] **Step 3: Apply the migration locally**

Run from `backend/`:

```bash
go run database/migrate_all.go
```

Expected output: log line confirming `017_persist_vendor_name.up.sql` applied. No errors.

- [ ] **Step 4: Verify columns exist**

Run:

```bash
psql "$DATABASE_URL" -c "\d purchase_orders"  | grep vendor_name
psql "$DATABASE_URL" -c "\d payment_vouchers" | grep vendor_name
psql "$DATABASE_URL" -c "\d requisitions"     | grep preferred_vendor_name
```

Expected: each command prints one row showing the new column as `character varying(255)`.

- [ ] **Step 5: Commit**

```bash
git add backend/database/migrations/017_persist_vendor_name.up.sql backend/database/migrations/017_persist_vendor_name.down.sql
git commit -m "feat(db): add vendor_name columns for manual vendor entry"
```

---

## Task 2: Model — strip `gorm:"-"` tags

**Files:**
- Modify: `backend/models/models.go:67` and `backend/models/models.go:146`

- [ ] **Step 1: Edit `Requisition.PreferredVendorName`**

Replace at line 67:

```go
PreferredVendorName string        `gorm:"-" json:"preferredVendorName"`
```

with:

```go
PreferredVendorName string        `json:"preferredVendorName"`                       // Persisted; falls back to PreferredVendor.Name on read
```

- [ ] **Step 2: Edit `PurchaseOrder.VendorName`**

Replace at line 146:

```go
VendorName    string     `gorm:"-" json:"vendorName,omitempty"`    // Computed from Vendor.Name
```

with:

```go
VendorName    string     `json:"vendorName,omitempty"`              // Persisted; falls back to Vendor.Name on read
```

- [ ] **Step 3: Build to verify the package still compiles**

Run from `backend/`:

```bash
go build ./...
```

Expected: no output (success).

- [ ] **Step 4: Commit**

```bash
git add backend/models/models.go
git commit -m "feat(models): persist VendorName / PreferredVendorName columns"
```

---

## Task 3: Add request-type fields

**Files:**
- Modify: `backend/types/documents.go`

- [ ] **Step 1: Add `VendorName` to `CreatePurchaseOrderRequest`**

In `backend/types/documents.go`, locate `CreatePurchaseOrderRequest` (line 207). Insert `VendorName` immediately after `VendorID`:

```go
type CreatePurchaseOrderRequest struct {
    VendorID          string                 `json:"vendorId"`
    VendorName        string                 `json:"vendorName"`
    Items             []POItem               `json:"items" validate:"required,min=1"`
    // ... rest unchanged
}
```

- [ ] **Step 2: Add `VendorName` to `CreatePaymentVoucherRequest`**

Locate `CreatePaymentVoucherRequest` (line 317). Insert after `VendorID`:

```go
type CreatePaymentVoucherRequest struct {
    VendorID      string  `json:"vendorId"`
    VendorName    string  `json:"vendorName"`
    InvoiceNumber string  `json:"invoiceNumber" validate:"required"`
    // ... rest unchanged
}
```

- [ ] **Step 3: Add `VendorName` to `UpdatePaymentVoucherRequest`**

Locate `UpdatePaymentVoucherRequest` (line 339). Insert after `VendorID`:

```go
type UpdatePaymentVoucherRequest struct {
    VendorID      string  `json:"vendorId"`
    VendorName    string  `json:"vendorName"`
    InvoiceNumber string  `json:"invoiceNumber"`
    // ... rest unchanged
}
```

- [ ] **Step 4: Add `PreferredVendorName` to `CreateRequisitionRequest`**

Locate `CreateRequisitionRequest` (line 13). Insert after `PreferredVendorID`:

```go
type CreateRequisitionRequest struct {
    // ... existing fields up to PreferredVendorID
    PreferredVendorID   *string           `json:"preferredVendorId" validate:"omitempty,uuid"`
    PreferredVendorName string            `json:"preferredVendorName"`
    IsEstimate          bool              `json:"isEstimate"`
    // ... rest unchanged
}
```

- [ ] **Step 5: Add `PreferredVendorName` to `UpdateRequisitionRequest`**

Locate `UpdateRequisitionRequest` (line 38). Insert after `PreferredVendorID`:

```go
type UpdateRequisitionRequest struct {
    // ... existing fields up to PreferredVendorID
    PreferredVendorID   *string                `json:"preferredVendorId" validate:"omitempty,uuid"`
    PreferredVendorName string                 `json:"preferredVendorName"`
    IsEstimate          *bool                  `json:"isEstimate"`
    // ... rest unchanged
}
```

- [ ] **Step 6: Build**

```bash
go build ./...
```

Expected: no output.

- [ ] **Step 7: Commit**

```bash
git add backend/types/documents.go
git commit -m "feat(types): add VendorName / PreferredVendorName to request DTOs"
```

---

## Task 4: PO — failing test for manual vendor persistence

**Files:**
- Modify: `backend/handlers/purchase_orders_http_test.go`

- [ ] **Step 1: Add the failing test**

Append to `backend/handlers/purchase_orders_http_test.go`:

```go
func TestPurchaseOrder_PersistManualVendorName(t *testing.T) {
    db := setupTestDB(t)
    defer teardownTestDB(t, db)

    app := newPurchaseOrderApp(t)

    // CREATE: manual vendor (no vendor_id, vendorName only)
    createBody := map[string]interface{}{
        "vendorId":     "",
        "vendorName":   "LIKS BUSINESS SOLUTIONS",
        "items":        []map[string]interface{}{{"description": "Laptops", "quantity": 25, "unitPrice": 150000, "amount": 3750000}},
        "totalAmount":  3750000.0,
        "currency":     "ZMW",
        "deliveryDate": time.Now().Add(7 * 24 * time.Hour).Format(time.RFC3339),
        "title":        "Purchase of 25 Laptops",
        "department":   "Information Technology",
        "priority":     "high",
    }
    resp := testRequest(app, http.MethodPost, "/purchase-orders", createBody)
    if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
        t.Fatalf("create PO: expected 200/201, got %d", resp.StatusCode)
    }

    var createResp struct {
        Data types.PurchaseOrderResponse `json:"data"`
    }
    decodeJSON(t, resp, &createResp)

    if createResp.Data.VendorName != "LIKS BUSINESS SOLUTIONS" {
        t.Errorf("create response: expected vendorName=LIKS BUSINESS SOLUTIONS, got %q", createResp.Data.VendorName)
    }
    if createResp.Data.VendorID != "" {
        t.Errorf("create response: expected empty vendorId, got %q", createResp.Data.VendorID)
    }

    // GET: fetch the same PO and confirm vendor_name persists
    poID := createResp.Data.ID
    getResp := testRequest(app, http.MethodGet, "/purchase-orders/"+poID, nil)
    if getResp.StatusCode != http.StatusOK {
        t.Fatalf("get PO: expected 200, got %d", getResp.StatusCode)
    }

    var getBody struct {
        Data types.PurchaseOrderResponse `json:"data"`
    }
    decodeJSON(t, getResp, &getBody)

    if getBody.Data.VendorName != "LIKS BUSINESS SOLUTIONS" {
        t.Errorf("get response: expected vendorName=LIKS BUSINESS SOLUTIONS, got %q", getBody.Data.VendorName)
    }

    // UPDATE: change the manual vendor name
    updateBody := map[string]interface{}{
        "vendorId":   "",
        "vendorName": "MICOP BUSINESS VENTURES",
    }
    updResp := testRequest(app, http.MethodPut, "/purchase-orders/"+poID, updateBody)
    if updResp.StatusCode != http.StatusOK {
        t.Fatalf("update PO: expected 200, got %d", updResp.StatusCode)
    }

    var updRespBody struct {
        Data types.PurchaseOrderResponse `json:"data"`
    }
    decodeJSON(t, updResp, &updRespBody)

    if updRespBody.Data.VendorName != "MICOP BUSINESS VENTURES" {
        t.Errorf("update response: expected vendorName=MICOP BUSINESS VENTURES, got %q", updRespBody.Data.VendorName)
    }
}
```

If `decodeJSON` does not exist in the test file, add this helper near the top of the file (above the test functions):

```go
func decodeJSON(t *testing.T, resp *http.Response, out interface{}) {
    t.Helper()
    defer resp.Body.Close()
    if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
        t.Fatalf("decodeJSON: %v", err)
    }
}
```

If `encoding/json` and `net/http` are not yet imported in this file, add them.

- [ ] **Step 2: Run test to confirm it fails**

Run from `backend/`:

```bash
go test ./handlers/ -run TestPurchaseOrder_PersistManualVendorName -v
```

Expected: FAIL. The create response will return `vendorName=""` because the create handler does not yet propagate `req.VendorName` to the model.

- [ ] **Step 3: Commit the failing test**

```bash
git add backend/handlers/purchase_orders_http_test.go
git commit -m "test(po): failing test for manual vendor name persistence"
```

---

## Task 5: PO — wire create + update handlers + response

**Files:**
- Modify: `backend/handlers/purchase_order.go`

- [ ] **Step 1: Update `CreatePurchaseOrder` to set `VendorName`**

In `backend/handlers/purchase_order.go`, locate the `order := models.PurchaseOrder{` literal around line 276. Add `VendorName: req.VendorName,` immediately after the `VendorID:` line:

```go
order := models.PurchaseOrder{
    ID:                orderID,
    OrganizationID:    tenant.OrganizationID,
    DocumentNumber:    documentNumber,
    VendorID:          vendorIDPtr,
    VendorName:        req.VendorName,                // NEW
    Status:            models.StatusDraft,
    // ... rest unchanged
}
```

- [ ] **Step 2: Update `UpdatePurchaseOrder` to assign `VendorName`**

In the same file, locate the existing `if req.VendorID != "" {` block around line 504. Immediately after the closing brace of that block, insert:

```go
// Persist vendor name when a vendor change is part of this update.
// Guard prevents metadata-only updates (quotation list saves) from
// clobbering the stored name to empty.
if req.VendorID != "" || req.VendorName != "" {
    if order.VendorName != req.VendorName {
        changes["vendorName"] = map[string]string{"old": order.VendorName, "new": req.VendorName}
    }
    order.VendorName = req.VendorName
}
```

- [ ] **Step 3: Update `modelToPurchaseOrderResponse` fallback**

In the same file, locate `modelToPurchaseOrderResponse` around line 762. Replace the existing block:

```go
vendorName := ""
var vendorResp *types.VendorResponse
if order.Vendor != nil {
    vendorName = order.Vendor.Name
    vr := modelToVendorResponse(*order.Vendor)
    vendorResp = &vr
}
```

with:

```go
vendorName := order.VendorName              // stored fallback
var vendorResp *types.VendorResponse
if order.Vendor != nil {
    vendorName = order.Vendor.Name           // canonical wins when relation present
    vr := modelToVendorResponse(*order.Vendor)
    vendorResp = &vr
}
```

- [ ] **Step 4: Run the test to verify it now passes**

```bash
go test ./handlers/ -run TestPurchaseOrder_PersistManualVendorName -v
```

Expected: PASS.

- [ ] **Step 5: Run the full PO test suite for regressions**

```bash
go test ./handlers/ -run TestPurchaseOrder -v
```

Expected: all tests pass.

- [ ] **Step 6: Commit**

```bash
git add backend/handlers/purchase_order.go
git commit -m "feat(po): persist manual vendor name on create + update"
```

---

## Task 6: PV — failing test for manual vendor persistence

**Files:**
- Modify: `backend/handlers/payment_voucher_http_test.go` (or whichever existing PV test file lives in `backend/handlers/`)

- [ ] **Step 1: Locate the existing PV HTTP test file**

```bash
ls backend/handlers/payment_voucher*test* backend/handlers/payment_vouchers*test*
```

Use the file that already exists. If none exists, create `backend/handlers/payment_voucher_http_test.go` with a `newPaymentVoucherApp(t)` helper modelled on `newPurchaseOrderApp` in `purchase_orders_http_test.go`.

- [ ] **Step 2: Add the failing test**

Append to that file:

```go
func TestPaymentVoucher_PersistManualVendorName(t *testing.T) {
    db := setupTestDB(t)
    defer teardownTestDB(t, db)

    app := newPaymentVoucherApp(t)

    createBody := map[string]interface{}{
        "vendorId":      "",
        "vendorName":    "BLUE FOX FARMS LIMITED",
        "invoiceNumber": "INV-2026-001",
        "amount":        3547200.0,
        "currency":      "ZMW",
        "paymentMethod": "bank_transfer",
        "glCode":        "5100",
        "description":   "Payment for laptop procurement quotation",
    }
    resp := testRequest(app, http.MethodPost, "/payment-vouchers", createBody)
    if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
        t.Fatalf("create PV: expected 200/201, got %d", resp.StatusCode)
    }

    var createResp struct {
        Data types.PaymentVoucherResponse `json:"data"`
    }
    decodeJSON(t, resp, &createResp)

    if createResp.Data.VendorName != "BLUE FOX FARMS LIMITED" {
        t.Errorf("create response: expected vendorName=BLUE FOX FARMS LIMITED, got %q", createResp.Data.VendorName)
    }

    pvID := createResp.Data.ID
    getResp := testRequest(app, http.MethodGet, "/payment-vouchers/"+pvID, nil)
    if getResp.StatusCode != http.StatusOK {
        t.Fatalf("get PV: expected 200, got %d", getResp.StatusCode)
    }

    var getBody struct {
        Data types.PaymentVoucherResponse `json:"data"`
    }
    decodeJSON(t, getResp, &getBody)

    if getBody.Data.VendorName != "BLUE FOX FARMS LIMITED" {
        t.Errorf("get response: expected vendorName=BLUE FOX FARMS LIMITED, got %q", getBody.Data.VendorName)
    }

    updateBody := map[string]interface{}{
        "vendorId":   "",
        "vendorName": "MICOP BUSINESS VENTURES",
    }
    updResp := testRequest(app, http.MethodPut, "/payment-vouchers/"+pvID, updateBody)
    if updResp.StatusCode != http.StatusOK {
        t.Fatalf("update PV: expected 200, got %d", updResp.StatusCode)
    }
    var updRespBody struct {
        Data types.PaymentVoucherResponse `json:"data"`
    }
    decodeJSON(t, updResp, &updRespBody)
    if updRespBody.Data.VendorName != "MICOP BUSINESS VENTURES" {
        t.Errorf("update response: expected vendorName=MICOP BUSINESS VENTURES, got %q", updRespBody.Data.VendorName)
    }
}
```

Make sure `decodeJSON`, `encoding/json`, and `net/http` are available — reuse the helper from Task 4 if it lives in a shared `_test.go` file, or duplicate it if test files do not share helpers.

If `newPaymentVoucherApp` does not exist, add it modelled on the PO version:

```go
func newPaymentVoucherApp(t *testing.T) *fiber.App {
    t.Helper()
    app := fiber.New()
    auth := withTenantCtx(testOrgID, testUserID, testUserRole)

    app.Get("/payment-vouchers", auth, GetPaymentVouchers)
    app.Post("/payment-vouchers", auth, CreatePaymentVoucher)
    app.Get("/payment-vouchers/:id", auth, GetPaymentVoucher)
    app.Put("/payment-vouchers/:id", auth, UpdatePaymentVoucher)
    app.Delete("/payment-vouchers/:id", auth, DeletePaymentVoucher)
    return app
}
```

- [ ] **Step 3: Run the test to confirm it fails**

```bash
go test ./handlers/ -run TestPaymentVoucher_PersistManualVendorName -v
```

Expected: FAIL — create response returns empty `vendorName`.

- [ ] **Step 4: Commit the failing test**

```bash
git add backend/handlers/payment_voucher_http_test.go
git commit -m "test(pv): failing test for manual vendor name persistence"
```

---

## Task 7: PV — wire create + update handlers + response

**Files:**
- Modify: `backend/handlers/payment_voucher.go`

- [ ] **Step 1: Update `CreatePaymentVoucher` to set `VendorName`**

In `backend/handlers/payment_voucher.go`, locate the `voucher := models.PaymentVoucher{` literal around line 313. Add `VendorName: req.VendorName,` immediately after the `VendorID:` line:

```go
voucher := models.PaymentVoucher{
    ID:             uuid.New().String(),
    OrganizationID: tenant.OrganizationID,
    DocumentNumber: documentNumber,
    VendorID:       vendorIDPtr,
    VendorName:     req.VendorName,                  // NEW
    InvoiceNumber:  req.InvoiceNumber,
    // ... rest unchanged
}
```

- [ ] **Step 2: Update `UpdatePaymentVoucher` to assign `VendorName`**

Locate the existing `if req.VendorID != "" {` block around line 481. Immediately after that block, insert:

```go
if req.VendorID != "" || req.VendorName != "" {
    voucher.VendorName = req.VendorName
}
```

- [ ] **Step 3: Update `modelToPaymentVoucherResponse` fallback**

Locate `modelToPaymentVoucherResponse` around line 622. Replace:

```go
vendorName := ""
var vendorResp *types.VendorResponse
if voucher.Vendor != nil {
    vendorName = voucher.Vendor.Name
    vr := modelToVendorResponse(*voucher.Vendor)
    vendorResp = &vr
}
```

with:

```go
vendorName := voucher.VendorName             // stored fallback
var vendorResp *types.VendorResponse
if voucher.Vendor != nil {
    vendorName = voucher.Vendor.Name          // canonical wins
    vr := modelToVendorResponse(*voucher.Vendor)
    vendorResp = &vr
}
```

- [ ] **Step 4: Run the test to verify it passes**

```bash
go test ./handlers/ -run TestPaymentVoucher_PersistManualVendorName -v
```

Expected: PASS.

- [ ] **Step 5: Run the full PV test suite**

```bash
go test ./handlers/ -run TestPaymentVoucher -v
```

Expected: all tests pass.

- [ ] **Step 6: Commit**

```bash
git add backend/handlers/payment_voucher.go
git commit -m "feat(pv): persist manual vendor name on create + update"
```

---

## Task 8: Requisition — failing test for manual preferred-vendor persistence

**Files:**
- Modify: existing requisition HTTP test file under `backend/handlers/` (e.g. `requisition_http_test.go` or `requisitions_http_test.go`)

- [ ] **Step 1: Locate existing requisition test file**

```bash
ls backend/handlers/requisition*test*
```

Use whichever exists. If a `newRequisitionApp(t)` helper does not exist, add one modelled on the PO version, registering at minimum: `POST /requisitions`, `GET /requisitions/:id`, `PUT /requisitions/:id`.

- [ ] **Step 2: Add the failing test**

Append to that file:

```go
func TestRequisition_PersistManualPreferredVendorName(t *testing.T) {
    db := setupTestDB(t)
    defer teardownTestDB(t, db)

    app := newRequisitionApp(t)

    createBody := map[string]interface{}{
        "title":               "Test requisition with manual preferred vendor",
        "description":         "Need supplies from a vendor not in the system",
        "department":          "Information Technology",
        "priority":            "medium",
        "items":               []map[string]interface{}{{"description": "Toner", "quantity": 5, "unitPrice": 200, "amount": 1000}},
        "totalAmount":         1000.0,
        "currency":            "ZMW",
        "preferredVendorName": "AD-HOC SUPPLIER LTD",
    }
    resp := testRequest(app, http.MethodPost, "/requisitions", createBody)
    if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
        t.Fatalf("create requisition: expected 200/201, got %d", resp.StatusCode)
    }

    var createResp struct {
        Data types.RequisitionResponse `json:"data"`
    }
    decodeJSON(t, resp, &createResp)

    if createResp.Data.PreferredVendorName != "AD-HOC SUPPLIER LTD" {
        t.Errorf("create response: expected preferredVendorName=AD-HOC SUPPLIER LTD, got %q", createResp.Data.PreferredVendorName)
    }

    reqID := createResp.Data.ID
    getResp := testRequest(app, http.MethodGet, "/requisitions/"+reqID, nil)
    if getResp.StatusCode != http.StatusOK {
        t.Fatalf("get requisition: expected 200, got %d", getResp.StatusCode)
    }
    var getBody struct {
        Data types.RequisitionResponse `json:"data"`
    }
    decodeJSON(t, getResp, &getBody)
    if getBody.Data.PreferredVendorName != "AD-HOC SUPPLIER LTD" {
        t.Errorf("get response: expected preferredVendorName=AD-HOC SUPPLIER LTD, got %q", getBody.Data.PreferredVendorName)
    }

    updateBody := map[string]interface{}{
        "preferredVendorName": "ANOTHER AD-HOC VENDOR",
    }
    updResp := testRequest(app, http.MethodPut, "/requisitions/"+reqID, updateBody)
    if updResp.StatusCode != http.StatusOK {
        t.Fatalf("update requisition: expected 200, got %d", updResp.StatusCode)
    }
    var updRespBody struct {
        Data types.RequisitionResponse `json:"data"`
    }
    decodeJSON(t, updResp, &updRespBody)
    if updRespBody.Data.PreferredVendorName != "ANOTHER AD-HOC VENDOR" {
        t.Errorf("update response: expected preferredVendorName=ANOTHER AD-HOC VENDOR, got %q", updRespBody.Data.PreferredVendorName)
    }
}
```

- [ ] **Step 3: Run test to confirm it fails**

```bash
go test ./handlers/ -run TestRequisition_PersistManualPreferredVendorName -v
```

Expected: FAIL.

- [ ] **Step 4: Commit the failing test**

```bash
git add backend/handlers/requisition_http_test.go
git commit -m "test(requisition): failing test for manual preferred vendor name"
```

---

## Task 9: Requisition — wire create + update handlers + response

**Files:**
- Modify: `backend/handlers/requisition.go`

- [ ] **Step 1: Update `CreateRequisition` to set `PreferredVendorName`**

In `backend/handlers/requisition.go`, locate the `models.Requisition{` literal around line 320. Add `PreferredVendorName: req.PreferredVendorName,` immediately after the `PreferredVendorID:` line:

```go
PreferredVendorID:   req.PreferredVendorID,
PreferredVendorName: req.PreferredVendorName,        // NEW
IsEstimate:          req.IsEstimate,
```

- [ ] **Step 2: Update `UpdateRequisition` to assign `PreferredVendorName`**

Locate the `if req.PreferredVendorID != nil {` block around line 536. Immediately after the closing brace of that block, insert:

```go
// Persist preferred vendor name when supplied (covers manual ad-hoc vendors).
if req.PreferredVendorName != "" || (req.PreferredVendorID != nil && *req.PreferredVendorID == "") {
    requisition.PreferredVendorName = req.PreferredVendorName
}
```

The second clause allows a client to clear the manual name by sending an explicit empty `preferredVendorId` along with an empty `preferredVendorName`.

- [ ] **Step 3: Update the requisition response builder fallback**

Locate the requisition response builder around line 788. Replace:

```go
preferredVendorName := ""
var preferredVendorResp *types.VendorResponse
if req.PreferredVendor != nil {
    preferredVendorName = req.PreferredVendor.Name
    vr := modelToVendorResponse(*req.PreferredVendor)
    preferredVendorResp = &vr
}
```

with:

```go
preferredVendorName := req.PreferredVendorName            // stored fallback
var preferredVendorResp *types.VendorResponse
if req.PreferredVendor != nil {
    preferredVendorName = req.PreferredVendor.Name         // canonical wins
    vr := modelToVendorResponse(*req.PreferredVendor)
    preferredVendorResp = &vr
}
```

- [ ] **Step 4: Run the test to verify it passes**

```bash
go test ./handlers/ -run TestRequisition_PersistManualPreferredVendorName -v
```

Expected: PASS.

- [ ] **Step 5: Run the full requisition test suite**

```bash
go test ./handlers/ -run TestRequisition -v
```

Expected: all tests pass.

- [ ] **Step 6: Commit**

```bash
git add backend/handlers/requisition.go
git commit -m "feat(requisition): persist manual preferred vendor name on create + update"
```

---

## Task 10: Full backend regression check

- [ ] **Step 1: Run the full backend test suite**

```bash
go test ./...
```

Expected: all tests pass. If failures appear in unrelated suites, investigate; do not silence.

- [ ] **Step 2: Build the backend binary**

```bash
go build ./...
```

Expected: clean build.

- [ ] **Step 3: Manual smoke test — manual vendor PO end-to-end**

Start the backend (`go run main.go` or your usual dev command) and the frontend (`pnpm dev` from `frontend/`). Then:

1. Navigate to a draft PO.
2. Open Supporting Docs tab → Quotations → Add Quotation.
3. Pick *None (Enter Manually)*. Type a vendor name (e.g. `LIKS BUSINESS SOLUTIONS`). Enter an amount. Save.
4. Click *Select* on that quotation.
5. Confirm the PO header *Vendor* field now shows `LIKS BUSINESS SOLUTIONS` (was previously `—`).
6. Click *Submit for Approval*. Confirm the *You must select a vendor* error is gone.
7. Confirm the workflow proceeds and the submit dialog summary shows the vendor name.

Expected: all six checks pass.

- [ ] **Step 4: Final commit (only if anything was tweaked during smoke testing)**

If steps 1–3 above caused additional changes, commit them with a descriptive message. Otherwise skip.

---

## Self-review notes

- All five spec sections (migration, model tags, request DTO fields, handler wiring, response fallback) have an explicit task.
- No placeholders. Every code step shows the exact code.
- Type consistency: `VendorName` is the JSON/Go field everywhere; `vendor_name` is the column. `PreferredVendorName` / `preferred_vendor_name` for requisitions. Same names used in tests.
- Out-of-scope items from spec (GRN, backfill, vendor master-data validation, PDF) are not in the plan — intentional.
