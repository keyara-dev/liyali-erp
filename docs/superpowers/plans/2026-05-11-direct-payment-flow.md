# Direct Payment Flow Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a third workflow routing path `direct_payment` for strict finance payouts (wages/allowances), auto-create approved-PO + draft-PV, hide from procurement role, capture payee details, require proof-of-payment upload to close. Rename sidebar group from "Procurement" to "Source to Pay".

**Architecture:** Extends existing accounting auto-PO path. New denormalized `routing_type` column on requisitions/POs/PVs propagated from `Workflow.Conditions.RoutingType` at submission. New `payees` table with vendor/employee/other types. Procurement role excluded by extending `DocumentScope`. PV gains `paid` terminal status with mandatory POP attachment.

**Tech Stack:** Backend: Go 1.x + Fiber + GORM + PostgreSQL + Goose. Frontend: Next.js App Router + TypeScript + Tailwind v4 + ShadCN UI + React Query.

**Spec:** `docs/superpowers/specs/2026-05-11-direct-payment-flow-design.md`

---

## Phase 1 — Schema migration + payee CRUD

### Task 1.1: Database migration

**Files:**
- Create: `backend/database/migrations/018_direct_payment.up.sql`
- Create: `backend/database/migrations/018_direct_payment.down.sql`

- [ ] **Step 1: Write up migration**

```sql
-- backend/database/migrations/018_direct_payment.up.sql
-- ============================================================================
-- DIRECT PAYMENT FLOW
-- Migration: 018_direct_payment
-- Adds payees table, routing_type denormalized column on requisitions/POs/PVs,
-- and proof-of-payment fields on payment_vouchers.
-- ============================================================================

CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE IF NOT EXISTS payees (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id  UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    payee_type       TEXT NOT NULL CHECK (payee_type IN ('vendor','employee','other')),
    name             TEXT NOT NULL,
    email            TEXT,
    phone            TEXT,
    bank_name        TEXT,
    bank_account     TEXT,
    tax_id           TEXT,
    source_vendor_id UUID NULL REFERENCES vendors(id) ON DELETE SET NULL,
    source_user_id   UUID NULL REFERENCES users(id) ON DELETE SET NULL,
    deleted_at       TIMESTAMPTZ NULL,
    created_by       UUID REFERENCES users(id),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_payees_org_type_name ON payees (organization_id, payee_type, name);
CREATE INDEX IF NOT EXISTS idx_payees_name_trgm    ON payees USING gin (name gin_trgm_ops);

ALTER TABLE requisitions
    ADD COLUMN IF NOT EXISTS routing_type   TEXT NOT NULL DEFAULT 'procurement',
    ADD COLUMN IF NOT EXISTS payee_id       UUID REFERENCES payees(id),
    ADD COLUMN IF NOT EXISTS payee_snapshot JSONB;

ALTER TABLE purchase_orders
    ADD COLUMN IF NOT EXISTS routing_type TEXT NOT NULL DEFAULT 'procurement';

ALTER TABLE payment_vouchers
    ADD COLUMN IF NOT EXISTS routing_type     TEXT NOT NULL DEFAULT 'procurement',
    ADD COLUMN IF NOT EXISTS proof_of_payment JSONB,
    ADD COLUMN IF NOT EXISTS paid_at          TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS paid_by          UUID REFERENCES users(id);

CREATE INDEX IF NOT EXISTS idx_requisitions_routing_type_org
    ON requisitions (organization_id, routing_type);
CREATE INDEX IF NOT EXISTS idx_purchase_orders_routing_type_org
    ON purchase_orders (organization_id, routing_type);
CREATE INDEX IF NOT EXISTS idx_payment_vouchers_routing_type_org
    ON payment_vouchers (organization_id, routing_type);

-- Backfill from workflow conditions for existing rows.
UPDATE requisitions r
SET routing_type = COALESCE(NULLIF(w.conditions->>'routingType', ''), 'procurement')
FROM workflows w
WHERE r.workflow_id = w.id;

UPDATE purchase_orders po
SET routing_type = r.routing_type
FROM requisitions r
WHERE po.requisition_id = r.id;

UPDATE payment_vouchers pv
SET routing_type = po.routing_type
FROM purchase_orders po
WHERE pv.linked_po = po.document_number;
```

- [ ] **Step 2: Write down migration**

```sql
-- backend/database/migrations/018_direct_payment.down.sql
DROP INDEX IF EXISTS idx_payment_vouchers_routing_type_org;
DROP INDEX IF EXISTS idx_purchase_orders_routing_type_org;
DROP INDEX IF EXISTS idx_requisitions_routing_type_org;

ALTER TABLE payment_vouchers
    DROP COLUMN IF EXISTS paid_by,
    DROP COLUMN IF EXISTS paid_at,
    DROP COLUMN IF EXISTS proof_of_payment,
    DROP COLUMN IF EXISTS routing_type;

ALTER TABLE purchase_orders
    DROP COLUMN IF EXISTS routing_type;

ALTER TABLE requisitions
    DROP COLUMN IF EXISTS payee_snapshot,
    DROP COLUMN IF EXISTS payee_id,
    DROP COLUMN IF EXISTS routing_type;

DROP INDEX IF EXISTS idx_payees_name_trgm;
DROP INDEX IF EXISTS idx_payees_org_type_name;
DROP TABLE IF EXISTS payees;
```

- [ ] **Step 3: Run migration up**

Run: `cd backend && goose -dir database/migrations postgres "$DATABASE_URL" up`
Expected: `OK 018_direct_payment.up.sql`

- [ ] **Step 4: Verify schema**

Run: `psql "$DATABASE_URL" -c "\d payees" -c "\d+ requisitions" | grep routing_type`
Expected: `payees` table exists, `routing_type` column shown on requisitions.

- [ ] **Step 5: Test rollback then re-apply**

```bash
goose -dir database/migrations postgres "$DATABASE_URL" down
goose -dir database/migrations postgres "$DATABASE_URL" up
```
Expected: both succeed, no errors.

- [ ] **Step 6: Commit**

```bash
git add backend/database/migrations/018_direct_payment.up.sql backend/database/migrations/018_direct_payment.down.sql
git commit -m "feat(db): add payees table and routing_type columns for direct payment"
```

---

### Task 1.2: Payee model + GORM struct

**Files:**
- Modify: `backend/models/models.go` (add `Payee` struct near other entity models)

- [ ] **Step 1: Add Payee struct**

```go
// Payee — unified payee record for requisition payments. Vendor/Employee picks
// snapshot to this table for unified lookup; "other" entries live exclusively here.
type Payee struct {
    ID              string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
    OrganizationID  string         `gorm:"type:uuid;not null;index" json:"organizationId"`
    PayeeType       string         `gorm:"type:text;not null" json:"payeeType"` // vendor|employee|other
    Name            string         `gorm:"type:text;not null" json:"name"`
    Email           string         `gorm:"type:text" json:"email,omitempty"`
    Phone           string         `gorm:"type:text" json:"phone,omitempty"`
    BankName        string         `gorm:"type:text" json:"bankName,omitempty"`
    BankAccount     string         `gorm:"type:text" json:"bankAccount,omitempty"`
    TaxID           string         `gorm:"type:text;column:tax_id" json:"taxId,omitempty"`
    SourceVendorID  *string        `gorm:"type:uuid" json:"sourceVendorId,omitempty"`
    SourceUserID    *string        `gorm:"type:uuid" json:"sourceUserId,omitempty"`
    DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
    CreatedBy       *string        `gorm:"type:uuid" json:"createdBy,omitempty"`
    CreatedAt       time.Time      `json:"createdAt"`
    UpdatedAt       time.Time      `json:"updatedAt"`
}

func (Payee) TableName() string { return "payees" }
```

- [ ] **Step 2: Extend Requisition struct** — add three fields next to existing PreferredVendorName

```go
RoutingType    string  `gorm:"type:text;not null;default:'procurement';index" json:"routingType"`
PayeeID        *string `gorm:"type:uuid" json:"payeeId,omitempty"`
PayeeSnapshot  JSON    `gorm:"type:jsonb" json:"payeeSnapshot,omitempty"`
```

- [ ] **Step 3: Extend PurchaseOrder struct**

```go
RoutingType string `gorm:"type:text;not null;default:'procurement';index" json:"routingType"`
```

- [ ] **Step 4: Extend PaymentVoucher struct**

```go
RoutingType    string     `gorm:"type:text;not null;default:'procurement';index" json:"routingType"`
ProofOfPayment JSON       `gorm:"type:jsonb" json:"proofOfPayment,omitempty"`
PaidAt         *time.Time `json:"paidAt,omitempty"`
PaidBy         *string    `gorm:"type:uuid" json:"paidBy,omitempty"`
```

- [ ] **Step 5: Add status constant**

In the file that defines `PaymentVoucher` status constants (search for `StatusApproved` near PaymentVoucher), add:
```go
const PaymentVoucherStatusPaid = "paid"
```

- [ ] **Step 6: Build to verify**

Run: `cd backend && go build ./...`
Expected: no errors.

- [ ] **Step 7: Commit**

```bash
git add backend/models/models.go
git commit -m "feat(models): add Payee model and routing_type fields"
```

---

### Task 1.3: Payee handler tests (failing)

**Files:**
- Create: `backend/handlers/payees_http_test.go`

- [ ] **Step 1: Write failing tests**

Use the existing pattern from `vendors_http_test.go` (HTTP-level test via Fiber test app).

```go
package handlers

import (
    "encoding/json"
    "net/http"
    "strings"
    "testing"
)

func TestPayees_CreateAndList(t *testing.T) {
    app, ctx := setupTestApp(t)
    defer ctx.Cleanup()

    body := `{"payeeType":"other","name":"John Doe","email":"jd@example.com","bankName":"FNB","bankAccount":"1234567890"}`
    req := authedReq(t, ctx, http.MethodPost, "/api/payees", strings.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    resp, err := app.Test(req)
    if err != nil { t.Fatal(err) }
    if resp.StatusCode != http.StatusCreated {
        t.Fatalf("create: status=%d want 201", resp.StatusCode)
    }

    var created map[string]any
    json.NewDecoder(resp.Body).Decode(&created)
    if created["name"] != "John Doe" { t.Fatalf("name mismatch: %v", created["name"]) }
    if _, ok := created["id"].(string); !ok { t.Fatal("missing id") }

    // List
    req = authedReq(t, ctx, http.MethodGet, "/api/payees?type=other", nil)
    resp, err = app.Test(req)
    if err != nil { t.Fatal(err) }
    if resp.StatusCode != http.StatusOK { t.Fatalf("list: status=%d", resp.StatusCode) }
    var list struct { Items []map[string]any `json:"items"` }
    json.NewDecoder(resp.Body).Decode(&list)
    if len(list.Items) == 0 { t.Fatal("list empty after create") }
}

func TestPayees_SearchByName(t *testing.T) {
    app, ctx := setupTestApp(t)
    defer ctx.Cleanup()

    seedPayee(t, ctx, "vendor", "Alpha Suppliers")
    seedPayee(t, ctx, "vendor", "Beta Distributors")

    req := authedReq(t, ctx, http.MethodGet, "/api/payees?type=vendor&q=Alpha", nil)
    resp, err := app.Test(req)
    if err != nil { t.Fatal(err) }
    var list struct { Items []map[string]any `json:"items"` }
    json.NewDecoder(resp.Body).Decode(&list)
    if len(list.Items) != 1 || list.Items[0]["name"] != "Alpha Suppliers" {
        t.Fatalf("search returned %d items: %+v", len(list.Items), list.Items)
    }
}

func TestPayees_SoftDelete(t *testing.T) {
    app, ctx := setupTestApp(t)
    defer ctx.Cleanup()
    id := seedPayee(t, ctx, "other", "Soft Delete Me")

    req := authedReq(t, ctx, http.MethodDelete, "/api/payees/"+id, nil)
    resp, _ := app.Test(req)
    if resp.StatusCode != http.StatusNoContent { t.Fatalf("delete status=%d", resp.StatusCode) }

    // List excludes soft-deleted
    req = authedReq(t, ctx, http.MethodGet, "/api/payees?type=other", nil)
    resp, _ = app.Test(req)
    var list struct { Items []map[string]any `json:"items"` }
    json.NewDecoder(resp.Body).Decode(&list)
    for _, it := range list.Items {
        if it["id"] == id { t.Fatal("soft-deleted payee still listed") }
    }
}

func seedPayee(t *testing.T, ctx *TestCtx, payeeType, name string) string {
    t.Helper()
    body := `{"payeeType":"` + payeeType + `","name":"` + name + `"}`
    req := authedReq(t, ctx, http.MethodPost, "/api/payees", strings.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    resp, err := ctx.App.Test(req)
    if err != nil { t.Fatal(err) }
    if resp.StatusCode != http.StatusCreated { t.Fatalf("seed: %d", resp.StatusCode) }
    var p map[string]any
    json.NewDecoder(resp.Body).Decode(&p)
    return p["id"].(string)
}
```

> If `setupTestApp`/`authedReq`/`TestCtx` helper signatures differ in this repo, mirror the exact names used in `vendors_http_test.go`. Do NOT change those helpers.

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd backend && go test ./handlers -run TestPayees -v`
Expected: FAIL — handler/routes not registered.

- [ ] **Step 3: Commit failing tests**

```bash
git add backend/handlers/payees_http_test.go
git commit -m "test(payees): failing tests for payees CRUD endpoints"
```

---

### Task 1.4: Payee handler implementation

**Files:**
- Create: `backend/handlers/payee.go`

- [ ] **Step 1: Implement handler**

```go
package handlers

import (
    "github.com/gofiber/fiber/v2"
    "gorm.io/gorm"

    "github.com/liyali/gateway/backend/models"
)

type PayeeHandler struct {
    DB *gorm.DB
}

func NewPayeeHandler(db *gorm.DB) *PayeeHandler { return &PayeeHandler{DB: db} }

type listPayeesQuery struct {
    Type string `query:"type"`
    Q    string `query:"q"`
}

func (h *PayeeHandler) List(c *fiber.Ctx) error {
    user := mustUser(c)
    var q listPayeesQuery
    if err := c.QueryParser(&q); err != nil {
        return fiber.NewError(fiber.StatusBadRequest, "invalid query")
    }
    db := h.DB.Where("organization_id = ?", user.OrganizationID)
    if q.Type != "" {
        db = db.Where("payee_type = ?", q.Type)
    }
    if q.Q != "" {
        db = db.Where("name ILIKE ?", "%"+q.Q+"%")
    }
    var items []models.Payee
    if err := db.Order("name ASC").Limit(100).Find(&items).Error; err != nil {
        return fiber.NewError(fiber.StatusInternalServerError, err.Error())
    }
    return c.JSON(fiber.Map{"items": items})
}

type createPayeeReq struct {
    PayeeType      string  `json:"payeeType"`
    Name           string  `json:"name"`
    Email          string  `json:"email"`
    Phone          string  `json:"phone"`
    BankName       string  `json:"bankName"`
    BankAccount    string  `json:"bankAccount"`
    TaxID          string  `json:"taxId"`
    SourceVendorID *string `json:"sourceVendorId"`
    SourceUserID   *string `json:"sourceUserId"`
}

func (h *PayeeHandler) Create(c *fiber.Ctx) error {
    user := mustUser(c)
    var req createPayeeReq
    if err := c.BodyParser(&req); err != nil {
        return fiber.NewError(fiber.StatusBadRequest, "invalid body")
    }
    if req.Name == "" || req.PayeeType == "" {
        return fiber.NewError(fiber.StatusBadRequest, "name and payeeType required")
    }
    if req.PayeeType != "vendor" && req.PayeeType != "employee" && req.PayeeType != "other" {
        return fiber.NewError(fiber.StatusBadRequest, "payeeType must be vendor|employee|other")
    }
    p := models.Payee{
        OrganizationID: user.OrganizationID,
        PayeeType:      req.PayeeType,
        Name:           req.Name,
        Email:          req.Email,
        Phone:          req.Phone,
        BankName:       req.BankName,
        BankAccount:    req.BankAccount,
        TaxID:          req.TaxID,
        SourceVendorID: req.SourceVendorID,
        SourceUserID:   req.SourceUserID,
        CreatedBy:      &user.ID,
    }
    if err := h.DB.Create(&p).Error; err != nil {
        return fiber.NewError(fiber.StatusInternalServerError, err.Error())
    }
    return c.Status(fiber.StatusCreated).JSON(p)
}

func (h *PayeeHandler) Get(c *fiber.Ctx) error {
    user := mustUser(c)
    id := c.Params("id")
    var p models.Payee
    if err := h.DB.Where("id = ? AND organization_id = ?", id, user.OrganizationID).First(&p).Error; err != nil {
        return fiber.NewError(fiber.StatusNotFound, "payee not found")
    }
    return c.JSON(p)
}

func (h *PayeeHandler) Update(c *fiber.Ctx) error {
    user := mustUser(c)
    id := c.Params("id")
    var p models.Payee
    if err := h.DB.Where("id = ? AND organization_id = ?", id, user.OrganizationID).First(&p).Error; err != nil {
        return fiber.NewError(fiber.StatusNotFound, "payee not found")
    }
    var req createPayeeReq
    if err := c.BodyParser(&req); err != nil {
        return fiber.NewError(fiber.StatusBadRequest, "invalid body")
    }
    if req.Name != "" { p.Name = req.Name }
    if req.Email != "" { p.Email = req.Email }
    if req.Phone != "" { p.Phone = req.Phone }
    if req.BankName != "" { p.BankName = req.BankName }
    if req.BankAccount != "" { p.BankAccount = req.BankAccount }
    if req.TaxID != "" { p.TaxID = req.TaxID }
    if err := h.DB.Save(&p).Error; err != nil {
        return fiber.NewError(fiber.StatusInternalServerError, err.Error())
    }
    return c.JSON(p)
}

func (h *PayeeHandler) Delete(c *fiber.Ctx) error {
    user := mustUser(c)
    id := c.Params("id")
    res := h.DB.Where("id = ? AND organization_id = ?", id, user.OrganizationID).Delete(&models.Payee{})
    if res.Error != nil {
        return fiber.NewError(fiber.StatusInternalServerError, res.Error.Error())
    }
    if res.RowsAffected == 0 {
        return fiber.NewError(fiber.StatusNotFound, "payee not found")
    }
    return c.SendStatus(fiber.StatusNoContent)
}
```

> If the project does not use a `mustUser(c)` helper, mirror the exact auth-extraction pattern used by `vendor.go` for the same purpose (read `OrganizationID` + user ID from the locals/context).

- [ ] **Step 2: Register routes** — add to `handler_registry.go` (or wherever routes are wired)

```go
payeeHandler := handlers.NewPayeeHandler(db)
api := app.Group("/api", authMiddleware)
api.Get("/payees", payeeHandler.List)
api.Post("/payees", payeeHandler.Create)
api.Get("/payees/:id", payeeHandler.Get)
api.Put("/payees/:id", payeeHandler.Update)
api.Delete("/payees/:id", payeeHandler.Delete)
```

> Match the route group/auth middleware pattern already used for `/api/vendors`. Place adjacent to vendor route registration.

- [ ] **Step 3: Run tests**

Run: `cd backend && go test ./handlers -run TestPayees -v`
Expected: all PASS.

- [ ] **Step 4: Commit**

```bash
git add backend/handlers/payee.go backend/handlers/handler_registry.go
git commit -m "feat(payees): CRUD endpoints for unified payee table"
```

---

## Phase 2 — Workflow conditions + routing enum

### Task 2.1: Add `direct_payment` to RoutingType + validation

**Files:**
- Modify: `backend/models/enhanced_auth.go` (the `WorkflowConditions` struct and any routing-type constants near lines 188–199)
- Modify: `backend/services/workflow_service.go` (or wherever workflow saves are validated)

- [ ] **Step 1: Locate the routing type constants**

Run: `grep -n "RoutingType" backend/models/enhanced_auth.go backend/services/*.go`
Note the exact symbol(s) used (e.g. `RoutingTypeProcurement`).

- [ ] **Step 2: Add the new constant**

Add next to existing routing type constants:
```go
const (
    RoutingTypeProcurement   = "procurement"
    RoutingTypeAccounting    = "accounting"
    RoutingTypeDirectPayment = "direct_payment"
)
```

(If constants are named differently in the repo, add `RoutingTypeDirectPayment = "direct_payment"` to the same block and stay consistent with naming.)

- [ ] **Step 3: Write failing validation test**

Create or extend `backend/services/workflow_service_test.go`:

```go
func TestSaveWorkflow_DirectPaymentRejectsStages(t *testing.T) {
    svc, db := newWorkflowServiceForTest(t)
    defer db.Close()

    wf := &models.Workflow{
        Name:           "Direct Pay Wages",
        OrganizationID: "org-1",
        Conditions:     models.JSON{"routingType": "direct_payment", "autoApprove": true},
        Stages: []models.WorkflowStage{
            {Name: "Manager Approval", Order: 1},
        },
    }
    err := svc.Save(context.Background(), wf)
    if err == nil {
        t.Fatal("expected validation error for direct_payment with stages > 0")
    }
    if !strings.Contains(err.Error(), "must have 0 approval stages") {
        t.Fatalf("unexpected error: %v", err)
    }
}

func TestSaveWorkflow_DirectPaymentZeroStagesOK(t *testing.T) {
    svc, db := newWorkflowServiceForTest(t)
    defer db.Close()

    wf := &models.Workflow{
        Name:           "Direct Pay Wages",
        OrganizationID: "org-1",
        Conditions:     models.JSON{"routingType": "direct_payment", "autoApprove": true},
    }
    if err := svc.Save(context.Background(), wf); err != nil {
        t.Fatalf("expected no error, got: %v", err)
    }
}
```

> `newWorkflowServiceForTest` should follow the pattern of other service tests; if no such helper exists in `backend/services`, use whatever fixture pattern is in `workflow_execution_service_test.go`.

- [ ] **Step 4: Run tests to verify failure**

Run: `cd backend && go test ./services -run TestSaveWorkflow_DirectPayment -v`
Expected: FAIL — no validation in place.

- [ ] **Step 5: Add validation in workflow save**

In the workflow save service method (where `Conditions` is parsed), add:
```go
cond := wf.GetConditions()
if cond.RoutingType == RoutingTypeDirectPayment && len(wf.Stages) > 0 {
    return errors.New("direct_payment workflows must have 0 approval stages")
}
```

- [ ] **Step 6: Run tests**

Run: `cd backend && go test ./services -run TestSaveWorkflow_DirectPayment -v`
Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add backend/models/enhanced_auth.go backend/services/workflow_service.go backend/services/workflow_service_test.go
git commit -m "feat(workflow): add direct_payment routing type with stages=0 validation"
```

---

## Phase 3 — Auto-flow extensions (auto-PV draft)

### Task 3.1: Failing test for autoCreateDraftPV

**Files:**
- Modify: `backend/services/workflow_execution_service_test.go`

- [ ] **Step 1: Write failing tests**

```go
func TestSubmitRequisitionWithRouting_DirectPayment_CreatesAutoPOAndDraftPV(t *testing.T) {
    svc, db, fix := setupExecServiceTest(t)
    defer db.Close()

    wf := fix.SeedWorkflow(t, fix.Org.ID, models.JSON{
        "routingType":    "direct_payment",
        "autoApprove":    true,
        "autoGeneratePO": true,
        "autoApprovePO":  true,
    }, nil)

    req := fix.SeedRequisition(t, fix.Org.ID, fix.User.ID, models.Requisition{
        Status:         "pending",
        TotalAmount:    1500.00,
        WorkflowID:     &wf.ID,
        RoutingType:    "direct_payment",
        PayeeSnapshot:  models.JSON{"name": "John Doe", "payeeType": "employee"},
    })

    res, err := svc.SubmitRequisitionWithRouting(context.Background(), req.ID, fix.User.ID)
    if err != nil { t.Fatalf("submit: %v", err) }

    if res.AutoCreatedPOID == "" { t.Fatal("expected auto-PO created") }
    if res.AutoCreatedPVID == "" { t.Fatal("expected auto-draft-PV created") }
    if res.RoutingType != "direct_payment" { t.Fatalf("routing=%s", res.RoutingType) }

    var pv models.PaymentVoucher
    if err := db.First(&pv, "id = ?", res.AutoCreatedPVID).Error; err != nil { t.Fatal(err) }
    if pv.Status != "draft" { t.Fatalf("pv.Status=%s want draft", pv.Status) }
    if pv.RoutingType != "direct_payment" { t.Fatalf("pv.RoutingType=%s", pv.RoutingType) }
    if pv.ProcurementFlow != "payment_first" { t.Fatalf("pv.ProcurementFlow=%s", pv.ProcurementFlow) }
    if pv.VendorName != "John Doe" { t.Fatalf("pv.VendorName=%s", pv.VendorName) }
}

func TestSubmitRequisitionWithRouting_Accounting_DoesNotCreatePV(t *testing.T) {
    svc, db, fix := setupExecServiceTest(t)
    defer db.Close()

    wf := fix.SeedWorkflow(t, fix.Org.ID, models.JSON{
        "routingType":    "accounting",
        "autoApprove":    true,
        "autoGeneratePO": true,
        "autoApprovePO":  true,
    }, nil)
    req := fix.SeedRequisition(t, fix.Org.ID, fix.User.ID, models.Requisition{
        Status: "pending", TotalAmount: 100, WorkflowID: &wf.ID, RoutingType: "accounting",
    })

    res, err := svc.SubmitRequisitionWithRouting(context.Background(), req.ID, fix.User.ID)
    if err != nil { t.Fatal(err) }
    if res.AutoCreatedPVID != "" { t.Fatalf("accounting should not auto-create PV, got %s", res.AutoCreatedPVID) }
}

func TestSubmitRequisitionWithRouting_DirectPayment_MissingPayee(t *testing.T) {
    svc, db, fix := setupExecServiceTest(t)
    defer db.Close()
    wf := fix.SeedWorkflow(t, fix.Org.ID, models.JSON{"routingType": "direct_payment", "autoApprove": true}, nil)
    req := fix.SeedRequisition(t, fix.Org.ID, fix.User.ID, models.Requisition{
        Status: "pending", TotalAmount: 100, WorkflowID: &wf.ID, RoutingType: "direct_payment",
        // no payee_snapshot
    })
    _, err := svc.SubmitRequisitionWithRouting(context.Background(), req.ID, fix.User.ID)
    if err == nil || !strings.Contains(err.Error(), "payee") {
        t.Fatalf("expected payee error, got %v", err)
    }
}
```

> `setupExecServiceTest` is the existing test fixture for this service. If naming differs, match the existing test file's setup function exactly.

- [ ] **Step 2: Run tests — expect failure**

Run: `cd backend && go test ./services -run TestSubmitRequisitionWithRouting_DirectPayment -v`
Expected: FAIL — no PV creation, no payee validation.

- [ ] **Step 3: Commit failing tests**

```bash
git add backend/services/workflow_execution_service_test.go
git commit -m "test(routing): failing tests for direct_payment auto-PV creation"
```

---

### Task 3.2: Implement `autoCreateDraftPV` and routing changes

**Files:**
- Modify: `backend/services/workflow_execution_service.go`

- [ ] **Step 1: Extend SubmitRoutingResult**

Locate the struct (audit identified around lines 204–210):
```go
type SubmitRoutingResult struct {
    // existing fields...
    AutoCreatedPOID string `json:"autoCreatedPOID,omitempty"`
    AutoCreatedPVID string `json:"autoCreatedPVID,omitempty"`
    RoutingType     string `json:"routingType"`
}
```

- [ ] **Step 2: Modify SubmitRequisitionWithRouting**

Inside `SubmitRequisitionWithRouting`, after the workflow is resolved and before/around the existing `autoApproveAndGeneratePO` call (around lines 212–284):

```go
routingType := conditions.RoutingType
if routingType == "" { routingType = RoutingTypeProcurement }

// Denormalize onto requisition.
if err := s.db.Model(&req).Update("routing_type", routingType).Error; err != nil {
    return nil, fmt.Errorf("set routing_type: %w", err)
}

if routingType == RoutingTypeDirectPayment {
    if req.PayeeID == nil && (req.PayeeSnapshot == nil || len(req.PayeeSnapshot) == 0) {
        return nil, errors.New("direct_payment requires payee_id or payee_snapshot")
    }
    if len(wf.Stages) > 0 {
        return nil, errors.New("direct_payment workflows must have 0 approval stages")
    }
}

// Existing auto-PO path runs for both accounting AND direct_payment.
// After auto-PO succeeds:
if routingType == RoutingTypeDirectPayment && result.AutoCreatedPOID != "" {
    pvID, err := s.autoCreateDraftPV(ctx, &req, result.AutoCreatedPOID)
    if err != nil {
        log.Printf("autoCreateDraftPV failed for req=%s po=%s: %v", req.ID, result.AutoCreatedPOID, err)
        // Do NOT roll back PO — audit trail intact. Recovery is manual.
    } else {
        result.AutoCreatedPVID = pvID
    }
}

result.RoutingType = routingType
```

- [ ] **Step 3: Implement autoCreateDraftPV helper**

Add near `autoApproveAndGeneratePO`:

```go
func (s *WorkflowExecutionService) autoCreateDraftPV(ctx context.Context, req *models.Requisition, poID string) (string, error) {
    var po models.PurchaseOrder
    if err := s.db.WithContext(ctx).First(&po, "id = ?", poID).Error; err != nil {
        return "", fmt.Errorf("load po: %w", err)
    }

    // Propagate routing_type onto PO.
    if err := s.db.WithContext(ctx).Model(&po).Update("routing_type", "direct_payment").Error; err != nil {
        return "", fmt.Errorf("set po.routing_type: %w", err)
    }

    name := "Direct Payment"
    if req.PayeeSnapshot != nil {
        if v, ok := req.PayeeSnapshot["name"].(string); ok && v != "" {
            name = v
        }
    }

    pv := models.PaymentVoucher{
        OrganizationID:  req.OrganizationID,
        Status:          "draft",
        CreatedBy:       req.CreatedBy,
        LinkedPO:        po.DocumentNumber,
        RoutingType:     "direct_payment",
        VendorName:      name,
        Amount:          req.TotalAmount,
        ProcurementFlow: "payment_first",
        Metadata: models.JSON{
            "autoCreated":   true,
            "sourceReqID":   req.ID,
            "payeeSnapshot": req.PayeeSnapshot,
        },
    }
    if err := s.db.WithContext(ctx).Create(&pv).Error; err != nil {
        return "", fmt.Errorf("create draft pv: %w", err)
    }
    return pv.ID, nil
}
```

> If `PaymentVoucher.CreatedBy` is `*string`, adapt the assignment. If a "system user" UUID is preferred over `req.CreatedBy`, look for `SystemUserID` constant in the codebase; otherwise use `req.CreatedBy` (the requester) — finance role visibility is handled by the document-scope filter, not by `CreatedBy`.

- [ ] **Step 4: Run tests**

Run: `cd backend && go test ./services -run TestSubmitRequisitionWithRouting -v`
Expected: PASS (all three direct_payment + accounting tests).

- [ ] **Step 5: Run full test suite for regressions**

Run: `cd backend && go test ./... -count=1`
Expected: all green.

- [ ] **Step 6: Commit**

```bash
git add backend/services/workflow_execution_service.go
git commit -m "feat(routing): auto-create draft PV for direct_payment requisitions"
```

---

## Phase 4 — Document scope filter

### Task 4.1: Failing tests for HideDirectPayment

**Files:**
- Modify: `backend/utils/document_scope_test.go`

- [ ] **Step 1: Write tests**

```go
func TestGetDocumentScope_ProcurementUserHidesDirectPayment(t *testing.T) {
    db, fix := setupScopeTest(t)
    defer db.Close()

    procUser := fix.SeedUser(t, "procurement", nil)
    scope, err := utils.GetDocumentScope(db, procUser.ID, "procurement", fix.Org.ID)
    if err != nil { t.Fatal(err) }
    if !scope.HideDirectPayment {
        t.Fatal("procurement user should hide direct_payment")
    }
}

func TestGetDocumentScope_FinanceUserDoesNotHide(t *testing.T) {
    db, fix := setupScopeTest(t)
    defer db.Close()
    fin := fix.SeedUser(t, "finance", nil)
    scope, _ := utils.GetDocumentScope(db, fin.ID, "finance", fix.Org.ID)
    if scope.HideDirectPayment { t.Fatal("finance user should not hide") }
}

func TestApplyToQuery_HidesDirectPaymentRows(t *testing.T) {
    db, fix := setupScopeTest(t)
    defer db.Close()
    fix.SeedRequisition(t, fix.Org.ID, fix.User.ID, models.Requisition{RoutingType: "procurement"})
    fix.SeedRequisition(t, fix.Org.ID, fix.User.ID, models.Requisition{RoutingType: "direct_payment"})

    scope := utils.DocumentScope{
        CanViewAll: false, IsProcurement: true, HideDirectPayment: true,
        UserID: fix.User.ID, OrgID: fix.Org.ID,
    }
    var reqs []models.Requisition
    q := scope.ApplyToQuery(db.Model(&models.Requisition{}), "requisitions.created_by", "requisition", "")
    if err := q.Find(&reqs).Error; err != nil { t.Fatal(err) }
    for _, r := range reqs {
        if r.RoutingType == "direct_payment" {
            t.Fatal("direct_payment requisition leaked to procurement scope")
        }
    }
}
```

- [ ] **Step 2: Run tests — expect failure**

Run: `cd backend && go test ./utils -run TestGetDocumentScope -run TestApplyToQuery -v`
Expected: FAIL — `HideDirectPayment` field missing.

- [ ] **Step 3: Commit failing tests**

```bash
git add backend/utils/document_scope_test.go
git commit -m "test(scope): failing tests for HideDirectPayment filter"
```

---

### Task 4.2: Extend DocumentScope

**Files:**
- Modify: `backend/utils/document_scope.go`

- [ ] **Step 1: Add field + populate**

In the `DocumentScope` struct (around lines 35–60):
```go
type DocumentScope struct {
    CanViewAll        bool
    IsProcurement     bool
    HideDirectPayment bool   // procurement role without finance/admin override
    UserID            string
    OrgID             string
    UserRole          string
}
```

In `GetDocumentScope` (lines ~60–93), after determining the role + permissions:

```go
// Procurement-only users (no finance/admin override) cannot see direct_payment chain.
hideDirect := false
if userRole == "procurement" && !scope.CanViewAll {
    hideDirect = true
}
// Also true when role isn't admin/finance and lacks payment_voucher.view permission
if !scope.CanViewAll && userRole != "finance" && userRole != "admin" {
    // check explicit permissions...
    if !hasPermission(db, userID, "payment_voucher.view") {
        hideDirect = true
    }
}
scope.HideDirectPayment = hideDirect
```

> Use whatever permission-check helper already exists; `hasPermission` is illustrative. If only roles drive visibility today, drop the permission check and key off role alone.

- [ ] **Step 2: Extend ApplyToQuery**

At the end of the existing method, after the existing owner/involvement filter logic:

```go
if scope.HideDirectPayment {
    // Note: ownerField may contain a table-qualified column name; we add a
    // separate condition keyed to whichever entity table is being queried.
    // The entityType arg lets us pick the right column.
    switch entityType {
    case "requisition":
        query = query.Where("requisitions.routing_type != ?", "direct_payment")
    case "purchase_order":
        query = query.Where("purchase_orders.routing_type != ?", "direct_payment")
    case "payment_voucher":
        query = query.Where("payment_vouchers.routing_type != ?", "direct_payment")
    case "grn":
        // GRNs join PO; filter via the linked PO.
        query = query.Where("EXISTS (SELECT 1 FROM purchase_orders po WHERE po.document_number = grns.linked_po AND po.routing_type != ?)", "direct_payment")
    case "budget":
        // Budgets are unaffected — direct payments may consume budgets too.
    }
}
return query
```

- [ ] **Step 3: Run scope tests**

Run: `cd backend && go test ./utils -v`
Expected: PASS.

- [ ] **Step 4: Commit**

```bash
git add backend/utils/document_scope.go
git commit -m "feat(scope): hide direct_payment chain from procurement users"
```

---

### Task 4.3: Apply scope filter to entity handlers

**Files:**
- Modify: `backend/handlers/requisition.go` (list + single-get)
- Modify: `backend/handlers/purchase_order.go` (list + single-get)
- Modify: `backend/handlers/payment_voucher.go` (list + single-get)
- Modify: `backend/handlers/grn.go` (list + single-get)

- [ ] **Step 1: Audit existing scope usage** — these handlers already call `GetDocumentScope` + `ApplyToQuery`. Confirm `HideDirectPayment` is honored by walking each call site.

Run: `grep -n "ApplyToQuery\|GetDocumentScope" backend/handlers/requisition.go backend/handlers/purchase_order.go backend/handlers/payment_voucher.go backend/handlers/grn.go`

- [ ] **Step 2: Ensure single-get also enforces filter**

Audit said `GetPurchaseOrder` already filters by `organization_id`. Add the routing_type check by calling scope.ApplyToQuery on the single-row fetch:

```go
scope, _ := utils.GetDocumentScope(db, user.ID, user.Role, user.OrganizationID)
q := db.Where("id = ? AND organization_id = ?", id, user.OrganizationID)
q = scope.ApplyToQuery(q.Model(&models.PurchaseOrder{}), "purchase_orders.created_by", "purchase_order", "")
var po models.PurchaseOrder
if err := q.First(&po).Error; err != nil {
    return fiber.NewError(fiber.StatusNotFound, "purchase order not found")
}
```

Repeat the same change to single-get on `requisition.go`, `payment_voucher.go`, `grn.go` (the audit confirmed `GetPurchaseOrder` lacked it; verify each by code reading).

- [ ] **Step 3: Add failing then passing HTTP test**

In `backend/handlers/purchase_orders_http_test.go`:

```go
func TestPO_ProcurementUserCannotSeeDirectPayment(t *testing.T) {
    app, ctx := setupTestApp(t)
    defer ctx.Cleanup()

    // Seed two POs: one direct_payment, one procurement.
    dpPO := seedPO(t, ctx, "direct_payment")
    procPO := seedPO(t, ctx, "procurement")

    ctx.LoginAs(t, "procurement")
    // List
    req := authedReq(t, ctx, http.MethodGet, "/api/purchase-orders", nil)
    resp, _ := app.Test(req)
    var page struct{ Items []map[string]any `json:"items"` }
    json.NewDecoder(resp.Body).Decode(&page)
    for _, p := range page.Items {
        if p["id"] == dpPO { t.Fatal("direct_payment PO leaked to procurement list") }
    }
    found := false
    for _, p := range page.Items { if p["id"] == procPO { found = true } }
    if !found { t.Fatal("procurement PO not visible to procurement user") }

    // Single
    req = authedReq(t, ctx, http.MethodGet, "/api/purchase-orders/"+dpPO, nil)
    resp, _ = app.Test(req)
    if resp.StatusCode != http.StatusNotFound {
        t.Fatalf("direct_payment PO single-get returned %d, want 404", resp.StatusCode)
    }
}
```

- [ ] **Step 4: Run tests**

Run: `cd backend && go test ./handlers -run TestPO_Procurement -v`
Expected: PASS.

- [ ] **Step 5: Repeat HTTP-test coverage for PV + requisition + GRN**

Mirror the same test pattern in `payment_vouchers_http_test.go`, `requisitions_http_test.go`, `grns_http_test.go`. Same shape: seed direct + non-direct rows, login as procurement, assert direct rows absent from list and 404 on single-get.

- [ ] **Step 6: Commit**

```bash
git add backend/handlers/requisition.go backend/handlers/purchase_order.go backend/handlers/payment_voucher.go backend/handlers/grn.go backend/handlers/*_http_test.go
git commit -m "feat(handlers): enforce HideDirectPayment on single-get + list for all entities"
```

---

## Phase 5 — Mark-as-Paid + recovery endpoints

### Task 5.1: MarkPaymentVoucherPaid handler (failing test)

**Files:**
- Modify: `backend/handlers/payment_vouchers_http_test.go`

- [ ] **Step 1: Write tests**

```go
func TestMarkPaid_RequiresApprovedStatus(t *testing.T) {
    app, ctx := setupTestApp(t)
    defer ctx.Cleanup()
    pvID := seedDraftPV(t, ctx, "direct_payment")
    body, contentType := multipartFile(t, "popFile", "slip.pdf", "application/pdf", []byte("PDF"))
    body = appendField(body, "paidDate", "2026-05-11")
    req := authedReq(t, ctx, http.MethodPost, "/api/payment-vouchers/"+pvID+"/mark-paid", body)
    req.Header.Set("Content-Type", contentType)
    resp, _ := app.Test(req)
    if resp.StatusCode != http.StatusConflict {
        t.Fatalf("status=%d want 409 (status must be approved)", resp.StatusCode)
    }
}

func TestMarkPaid_RequiresPOPFile(t *testing.T) {
    app, ctx := setupTestApp(t)
    defer ctx.Cleanup()
    pvID := seedApprovedPV(t, ctx, "direct_payment")
    body, contentType := multipartNoFile(t)
    req := authedReq(t, ctx, http.MethodPost, "/api/payment-vouchers/"+pvID+"/mark-paid", body)
    req.Header.Set("Content-Type", contentType)
    resp, _ := app.Test(req)
    if resp.StatusCode != http.StatusBadRequest {
        t.Fatalf("status=%d want 400 (no POP)", resp.StatusCode)
    }
}

func TestMarkPaid_HappyPath(t *testing.T) {
    app, ctx := setupTestApp(t)
    defer ctx.Cleanup()
    pvID := seedApprovedPV(t, ctx, "direct_payment")
    body, contentType := multipartFile(t, "popFile", "slip.pdf", "application/pdf", []byte("PDF"))
    body = appendField(body, "paidDate", "2026-05-11")
    req := authedReq(t, ctx, http.MethodPost, "/api/payment-vouchers/"+pvID+"/mark-paid", body)
    req.Header.Set("Content-Type", contentType)
    resp, _ := app.Test(req)
    if resp.StatusCode != http.StatusOK { t.Fatalf("status=%d", resp.StatusCode) }
    var pv map[string]any
    json.NewDecoder(resp.Body).Decode(&pv)
    if pv["status"] != "paid" { t.Fatalf("status=%v", pv["status"]) }
    if pv["proofOfPayment"] == nil { t.Fatal("proofOfPayment missing") }
}
```

> `multipartFile`/`appendField`/`multipartNoFile` are illustrative helpers — implement small `bytes.Buffer`+`multipart.Writer` helpers in the test file if not already present, mirroring patterns used elsewhere in this test directory.

- [ ] **Step 2: Run tests — expect failure**

Run: `cd backend && go test ./handlers -run TestMarkPaid -v`
Expected: FAIL — endpoint not registered.

- [ ] **Step 3: Commit failing tests**

```bash
git add backend/handlers/payment_vouchers_http_test.go
git commit -m "test(pv): failing tests for mark-paid endpoint"
```

---

### Task 5.2: MarkPaymentVoucherPaid implementation

**Files:**
- Modify: `backend/handlers/payment_voucher.go`
- Modify: `backend/handlers/handler_registry.go` (route registration)

- [ ] **Step 1: Add handler method**

```go
type markPaidForm struct {
    PaidDate string `form:"paidDate"`
    Notes    string `form:"notes"`
}

func (h *PaymentVoucherHandler) MarkPaid(c *fiber.Ctx) error {
    user := mustUser(c)
    id := c.Params("id")

    var pv models.PaymentVoucher
    if err := h.DB.Where("id = ? AND organization_id = ?", id, user.OrganizationID).First(&pv).Error; err != nil {
        return fiber.NewError(fiber.StatusNotFound, "payment voucher not found")
    }
    if pv.Status != "approved" {
        return fiber.NewError(fiber.StatusConflict, "payment voucher must be approved before mark-paid")
    }

    var form markPaidForm
    if err := c.BodyParser(&form); err != nil {
        return fiber.NewError(fiber.StatusBadRequest, "invalid form")
    }
    if form.PaidDate == "" {
        return fiber.NewError(fiber.StatusBadRequest, "paidDate required")
    }
    paidAt, err := time.Parse("2006-01-02", form.PaidDate)
    if err != nil { return fiber.NewError(fiber.StatusBadRequest, "paidDate must be YYYY-MM-DD") }

    file, err := c.FormFile("popFile")
    if err != nil || file == nil {
        return fiber.NewError(fiber.StatusBadRequest, "popFile required")
    }
    if file.Size > 10*1024*1024 {
        return fiber.NewError(fiber.StatusBadRequest, "popFile too large (max 10MB)")
    }
    if !isAllowedPOPMime(file.Header.Get("Content-Type")) {
        return fiber.NewError(fiber.StatusBadRequest, "popFile must be PDF/JPG/PNG")
    }

    // Upload via existing attachment service.
    uploaded, err := h.AttachmentService.Upload(c.Context(), user.OrganizationID, "payment_voucher", pv.ID, file)
    if err != nil {
        return fiber.NewError(fiber.StatusInternalServerError, "upload failed: "+err.Error())
    }

    pop := models.JSON{
        "fileId":     uploaded.ID,
        "fileName":   uploaded.Name,
        "fileUrl":    uploaded.URL,
        "mimeType":   uploaded.MimeType,
        "uploadedBy": user.ID,
        "uploadedAt": time.Now().UTC().Format(time.RFC3339),
    }

    err = h.DB.Transaction(func(tx *gorm.DB) error {
        if err := tx.Model(&pv).Updates(map[string]any{
            "status":           "paid",
            "proof_of_payment": pop,
            "paid_at":          paidAt,
            "paid_by":          user.ID,
        }).Error; err != nil { return err }

        // Cascade requisition completion for direct_payment chains.
        if pv.RoutingType == "direct_payment" {
            if md, ok := pv.Metadata["sourceReqID"].(string); ok && md != "" {
                if err := tx.Model(&models.Requisition{}).Where("id = ?", md).Update("status", "completed").Error; err != nil {
                    return err
                }
            }
        }
        return nil
    })
    if err != nil {
        return fiber.NewError(fiber.StatusInternalServerError, err.Error())
    }

    // Re-read to return fresh row.
    h.DB.First(&pv, "id = ?", id)
    return c.JSON(pv)
}

func isAllowedPOPMime(mime string) bool {
    switch mime {
    case "application/pdf", "image/jpeg", "image/png":
        return true
    }
    return false
}
```

> If `h.AttachmentService` is not the exact API name, use the same upload entry point used by requisition attachments. Verify by `grep -n "Attachment" backend/services/*.go`.

- [ ] **Step 2: Register route**

In the existing PV route group:
```go
api.Post("/payment-vouchers/:id/mark-paid", pvHandler.MarkPaid)
```

- [ ] **Step 3: Run tests**

Run: `cd backend && go test ./handlers -run TestMarkPaid -v`
Expected: PASS.

- [ ] **Step 4: Commit**

```bash
git add backend/handlers/payment_voucher.go backend/handlers/handler_registry.go
git commit -m "feat(pv): mark-paid endpoint with proof-of-payment upload"
```

---

### Task 5.3: Recovery endpoint

**Files:**
- Modify: `backend/handlers/payment_voucher.go`
- Modify: `backend/handlers/payment_vouchers_http_test.go`

- [ ] **Step 1: Write failing test**

```go
func TestRecoverFromPO_CreatesDraftPV(t *testing.T) {
    app, ctx := setupTestApp(t)
    defer ctx.Cleanup()
    poID := seedDirectPaymentPOWithoutPV(t, ctx)
    ctx.LoginAs(t, "admin")
    req := authedReq(t, ctx, http.MethodPost, "/api/payment-vouchers/recover-from-po/"+poID, nil)
    resp, _ := app.Test(req)
    if resp.StatusCode != http.StatusCreated { t.Fatalf("status=%d", resp.StatusCode) }

    // Idempotent
    resp2, _ := app.Test(req)
    if resp2.StatusCode != http.StatusOK { t.Fatalf("idempotent call status=%d, want 200", resp2.StatusCode) }
}
```

- [ ] **Step 2: Implement**

```go
func (h *PaymentVoucherHandler) RecoverFromPO(c *fiber.Ctx) error {
    user := mustUser(c)
    if !isAdminOrFinance(user) {
        return fiber.NewError(fiber.StatusForbidden, "admin/finance only")
    }
    poID := c.Params("poId")

    var po models.PurchaseOrder
    if err := h.DB.Where("id = ? AND organization_id = ?", poID, user.OrganizationID).First(&po).Error; err != nil {
        return fiber.NewError(fiber.StatusNotFound, "PO not found")
    }
    if po.RoutingType != "direct_payment" {
        return fiber.NewError(fiber.StatusBadRequest, "recovery only valid for direct_payment POs")
    }

    var existing models.PaymentVoucher
    if err := h.DB.Where("linked_po = ?", po.DocumentNumber).First(&existing).Error; err == nil {
        return c.Status(fiber.StatusOK).JSON(existing) // idempotent
    }

    // Reuse the autoCreateDraftPV pattern — minimal duplication.
    pv := models.PaymentVoucher{
        OrganizationID:  po.OrganizationID,
        Status:          "draft",
        CreatedBy:       po.CreatedBy,
        LinkedPO:        po.DocumentNumber,
        RoutingType:     "direct_payment",
        VendorName:      po.VendorName,
        Amount:          po.TotalAmount,
        ProcurementFlow: "payment_first",
        Metadata:        models.JSON{"recovered": true, "sourcePOID": po.ID},
    }
    if err := h.DB.Create(&pv).Error; err != nil {
        return fiber.NewError(fiber.StatusInternalServerError, err.Error())
    }
    return c.Status(fiber.StatusCreated).JSON(pv)
}
```

- [ ] **Step 3: Register route + run tests**

```go
api.Post("/payment-vouchers/recover-from-po/:poId", pvHandler.RecoverFromPO)
```
Run: `cd backend && go test ./handlers -run TestRecoverFromPO -v` → PASS.

- [ ] **Step 4: Commit**

```bash
git add backend/handlers/payment_voucher.go backend/handlers/payment_vouchers_http_test.go backend/handlers/handler_registry.go
git commit -m "feat(pv): recovery endpoint when autoCreateDraftPV fails"
```

---

## Phase 6 — Frontend types + API client

### Task 6.1: TypeScript types

**Files:**
- Modify: `frontend/src/types/requisition.ts`
- Modify: `frontend/src/types/payment-voucher.ts`
- Modify: `frontend/src/types/purchase-order.ts`
- Create: `frontend/src/types/payee.ts`

- [ ] **Step 1: Create payee.ts**

```ts
// frontend/src/types/payee.ts
export type PayeeType = 'vendor' | 'employee' | 'other';

export interface Payee {
  id: string;
  organizationId: string;
  payeeType: PayeeType;
  name: string;
  email?: string;
  phone?: string;
  bankName?: string;
  bankAccount?: string;
  taxId?: string;
  sourceVendorId?: string;
  sourceUserId?: string;
  createdAt: string;
  updatedAt: string;
}

export interface PayeeSnapshot {
  name: string;
  payeeType: PayeeType;
  email?: string;
  phone?: string;
  bankName?: string;
  bankAccount?: string;
  taxId?: string;
}

export interface CreatePayeeInput {
  payeeType: PayeeType;
  name: string;
  email?: string;
  phone?: string;
  bankName?: string;
  bankAccount?: string;
  taxId?: string;
  sourceVendorId?: string;
  sourceUserId?: string;
}
```

- [ ] **Step 2: Add RoutingType + extend Requisition**

```ts
// frontend/src/types/requisition.ts (append)
import type { PayeeSnapshot } from './payee';

export type RoutingType = 'procurement' | 'accounting' | 'direct_payment';

// Extend the existing Requisition interface — add to it, do not duplicate:
//   routingType: RoutingType;
//   payeeId?: string;
//   payeeSnapshot?: PayeeSnapshot;
```

(Open the file and add the three fields to the `Requisition` interface near the other denormalized fields.)

- [ ] **Step 3: Extend PaymentVoucher**

```ts
// frontend/src/types/payment-voucher.ts
import type { RoutingType } from './requisition';

export interface ProofOfPayment {
  fileId: string;
  fileName: string;
  fileUrl: string;
  mimeType: string;
  uploadedBy: string;
  uploadedAt: string;
}

// Update PaymentVoucherStatus union to add 'paid':
export type PaymentVoucherStatus =
  | 'draft'
  | 'submitted'
  | 'approved'
  | 'rejected'
  | 'paid';

// Extend PaymentVoucher interface — add:
//   routingType: RoutingType;
//   proofOfPayment?: ProofOfPayment;
//   paidAt?: string;
//   paidBy?: string;
```

- [ ] **Step 4: Extend PurchaseOrder**

```ts
// frontend/src/types/purchase-order.ts
import type { RoutingType } from './requisition';

// Add to PurchaseOrder interface:
//   routingType: RoutingType;
```

- [ ] **Step 5: Typecheck**

Run: `cd frontend && pnpm typecheck`
Expected: no errors.

- [ ] **Step 6: Commit**

```bash
git add frontend/src/types/payee.ts frontend/src/types/requisition.ts frontend/src/types/payment-voucher.ts frontend/src/types/purchase-order.ts
git commit -m "feat(types): add Payee, RoutingType, ProofOfPayment types"
```

---

### Task 6.2: Payee API client + React Query hooks

**Files:**
- Create: `frontend/src/lib/api/payees.ts`
- Create: `frontend/src/hooks/use-payees.ts`

- [ ] **Step 1: API client**

```ts
// frontend/src/lib/api/payees.ts
import { apiClient } from '@/lib/api/client';
import type { Payee, PayeeType, CreatePayeeInput } from '@/types/payee';

export async function listPayees(params: { type?: PayeeType; q?: string }) {
  const search = new URLSearchParams();
  if (params.type) search.set('type', params.type);
  if (params.q) search.set('q', params.q);
  const res = await apiClient.get<{ items: Payee[] }>(`/api/payees?${search}`);
  return res.data.items;
}

export async function createPayee(input: CreatePayeeInput) {
  const res = await apiClient.post<Payee>('/api/payees', input);
  return res.data;
}

export async function getPayee(id: string) {
  const res = await apiClient.get<Payee>(`/api/payees/${id}`);
  return res.data;
}
```

> Match the exact `apiClient` import path used by sibling files in `lib/api/` (e.g. `vendors.ts`).

- [ ] **Step 2: React Query hooks**

```ts
// frontend/src/hooks/use-payees.ts
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { createPayee, listPayees } from '@/lib/api/payees';
import type { PayeeType, CreatePayeeInput } from '@/types/payee';

export function usePayees(type: PayeeType | undefined, q?: string) {
  return useQuery({
    queryKey: ['payees', { type, q }],
    queryFn: () => listPayees({ type, q }),
    enabled: type !== undefined,
  });
}

export function useCreatePayee() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: CreatePayeeInput) => createPayee(input),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['payees'] }),
  });
}
```

- [ ] **Step 3: Typecheck**

Run: `cd frontend && pnpm typecheck`
Expected: no errors.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/lib/api/payees.ts frontend/src/hooks/use-payees.ts
git commit -m "feat(api): payees API client and React Query hooks"
```

---

## Phase 7 — Requisition form (request type + payee block)

### Task 7.1: Request type radio + form state

**Files:**
- Modify: the requisition create form (locate via `grep -rn "preferredVendorName" frontend/src/app/(private)`) — likely `frontend/src/app/(private)/(main)/(procurement)/requisitions/new/page.tsx` or a sub-component

- [ ] **Step 1: Locate the form file**

Run: `grep -rn "preferredVendorName" frontend/src/app frontend/src/components --include='*.tsx' -l`
Pick the requisition create form file (will be the one with field array + submit handler).

- [ ] **Step 2: Add request type to schema**

In the zod schema (or whatever validation lib is used):

```ts
const requisitionSchema = z.object({
  // ...existing fields
  routingType: z.enum(['procurement', 'accounting', 'direct_payment']).default('procurement'),
  payeeId: z.string().uuid().optional(),
  payeeSnapshot: z.object({
    name: z.string().min(1),
    payeeType: z.enum(['vendor', 'employee', 'other']),
    email: z.string().email().optional(),
    phone: z.string().optional(),
    bankName: z.string().optional(),
    bankAccount: z.string().optional(),
    taxId: z.string().optional(),
  }).optional(),
}).refine(
  (val) => val.routingType !== 'direct_payment' || val.payeeId || val.payeeSnapshot,
  { message: 'Direct payment requires a payee', path: ['payeeId'] }
);
```

- [ ] **Step 3: Render request type radio at top of form**

Above the existing vendor/payee section:

```tsx
<FormField
  control={form.control}
  name="routingType"
  render={({ field }) => (
    <FormItem className="space-y-3">
      <FormLabel>Request type</FormLabel>
      <RadioGroup
        value={field.value === 'direct_payment' ? 'direct_payment' : 'standard'}
        onValueChange={(v) => field.onChange(v === 'direct_payment' ? 'direct_payment' : 'procurement')}
        className="grid grid-cols-1 sm:grid-cols-2 gap-3"
      >
        <RadioCard value="standard" title="Goods / Services" description="Standard procurement or accounting route" />
        <RadioCard value="direct_payment" title="Direct Payment" description="Wages, allowances, individual payouts" />
      </RadioGroup>
      <FormMessage />
    </FormItem>
  )}
/>
```

> `RadioCard` is a shadcn radio styled as a card — use the same component already in use elsewhere in the codebase for similar choice cards. Run `grep -rn "RadioCard\|RadioGroupCard" frontend/src` to find the convention.

- [ ] **Step 4: Hide vendor section when direct_payment selected**

Find the existing vendor block and wrap:
```tsx
{form.watch('routingType') !== 'direct_payment' && (
  <>{/* existing vendor section */}</>
)}
```

- [ ] **Step 5: Typecheck + visual sanity**

Run: `cd frontend && pnpm typecheck && pnpm dev`
Manual: open `/requisitions/new`, click "Direct Payment", confirm vendor block hides.

- [ ] **Step 6: Commit**

```bash
git add frontend/src/app/(private)/(main)/(procurement)/requisitions/new
git commit -m "feat(req-form): add request type radio with direct payment option"
```

---

### Task 7.2: Payee block component

**Files:**
- Create: `frontend/src/components/requisitions/payee-block.tsx`

- [ ] **Step 1: Build component**

```tsx
'use client';

import { useState } from 'react';
import { Controller, useFormContext } from 'react-hook-form';
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group';
import { Label } from '@/components/ui/label';
import { Input } from '@/components/ui/input';
import { Combobox } from '@/components/ui/combobox';
import { usePayees, useCreatePayee } from '@/hooks/use-payees';
import { useVendors } from '@/hooks/use-vendors';
import { useUsers } from '@/hooks/use-users';

type SourceMode = 'new' | 'vendor' | 'employee' | 'other';

export function PayeeBlock() {
  const { control, setValue, watch } = useFormContext();
  const [mode, setMode] = useState<SourceMode>('new');
  const [newType, setNewType] = useState<'vendor' | 'employee' | 'other'>('other');
  const createPayee = useCreatePayee();

  // Data sources for dropdowns
  const vendors = useVendors();
  const employees = useUsers();
  const otherPayees = usePayees('other');

  return (
    <fieldset className="space-y-4 rounded-md border p-4">
      <legend className="text-sm font-medium">Payee</legend>

      <RadioGroup value={mode} onValueChange={(v) => setMode(v as SourceMode)} className="grid grid-cols-2 sm:grid-cols-4 gap-2">
        {(['new', 'vendor', 'employee', 'other'] as const).map((m) => (
          <div key={m} className="flex items-center space-x-2">
            <RadioGroupItem value={m} id={`payee-${m}`} />
            <Label htmlFor={`payee-${m}`} className="capitalize">{m}</Label>
          </div>
        ))}
      </RadioGroup>

      {mode === 'new' && (
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
          <div>
            <Label>Type</Label>
            <RadioGroup value={newType} onValueChange={(v) => setNewType(v as typeof newType)} className="flex gap-3 mt-1">
              {(['employee', 'vendor', 'other'] as const).map((t) => (
                <Label key={t} className="flex items-center gap-1 cursor-pointer">
                  <RadioGroupItem value={t} /> {t}
                </Label>
              ))}
            </RadioGroup>
          </div>
          <Controller
            name="payeeSnapshot.name"
            control={control}
            render={({ field }) => (
              <div>
                <Label>Name *</Label>
                <Input {...field} placeholder="Full name" required />
              </div>
            )}
          />
          {(['email', 'phone', 'bankName', 'bankAccount', 'taxId'] as const).map((f) => (
            <Controller
              key={f}
              name={`payeeSnapshot.${f}` as const}
              control={control}
              render={({ field }) => (
                <div>
                  <Label className="capitalize">{f.replace(/([A-Z])/g, ' $1')}</Label>
                  <Input {...field} value={field.value ?? ''} />
                </div>
              )}
            />
          ))}
          <div className="sm:col-span-2 text-xs text-muted-foreground">
            New payees are saved for future reuse.
          </div>
        </div>
      )}

      {mode === 'vendor' && (
        <Combobox
          items={(vendors.data ?? []).map((v) => ({ value: v.id, label: v.name }))}
          onSelect={(id) => {
            const v = vendors.data?.find((x) => x.id === id);
            if (!v) return;
            setValue('payeeSnapshot', { name: v.name, payeeType: 'vendor' });
          }}
          placeholder="Select vendor"
        />
      )}

      {mode === 'employee' && (
        <Combobox
          items={(employees.data ?? []).map((u) => ({ value: u.id, label: u.fullName ?? u.email }))}
          onSelect={(id) => {
            const u = employees.data?.find((x) => x.id === id);
            if (!u) return;
            setValue('payeeSnapshot', { name: u.fullName ?? u.email, payeeType: 'employee', email: u.email });
          }}
          placeholder="Select employee"
        />
      )}

      {mode === 'other' && (
        <Combobox
          items={(otherPayees.data ?? []).map((p) => ({ value: p.id, label: p.name }))}
          onSelect={(id) => {
            const p = otherPayees.data?.find((x) => x.id === id);
            if (!p) return;
            setValue('payeeId', p.id);
            setValue('payeeSnapshot', { name: p.name, payeeType: 'other' });
          }}
          placeholder="Select existing payee"
        />
      )}
    </fieldset>
  );
}

// Helper to persist a "new" payee. Called from the form's submit handler.
export async function persistNewPayeeIfNeeded(
  values: { payeeSnapshot?: { name: string; payeeType: 'vendor'|'employee'|'other'; [k: string]: any }; payeeId?: string },
  newTypeFromState: 'vendor' | 'employee' | 'other',
  createPayee: ReturnType<typeof useCreatePayee>,
): Promise<{ payeeId: string; payeeSnapshot: any }> {
  if (values.payeeId && values.payeeSnapshot) return { payeeId: values.payeeId, payeeSnapshot: values.payeeSnapshot };
  if (!values.payeeSnapshot) throw new Error('No payee data');
  const created = await createPayee.mutateAsync({
    payeeType: newTypeFromState,
    name: values.payeeSnapshot.name,
    email: values.payeeSnapshot.email,
    phone: values.payeeSnapshot.phone,
    bankName: values.payeeSnapshot.bankName,
    bankAccount: values.payeeSnapshot.bankAccount,
    taxId: values.payeeSnapshot.taxId,
  });
  return { payeeId: created.id, payeeSnapshot: { ...values.payeeSnapshot, payeeType: newTypeFromState } };
}
```

> If `Combobox`, `useVendors`, or `useUsers` use different names in this repo, use the actual ones (run `grep -rn "Combobox\|useVendors" frontend/src --include='*.tsx' --include='*.ts' -l`).

- [ ] **Step 2: Wire into form** — show `PayeeBlock` only when `routingType === 'direct_payment'`

In the form file (from Task 7.1):
```tsx
{form.watch('routingType') === 'direct_payment' && <PayeeBlock />}
```

- [ ] **Step 3: Submit handler — persist "new" payee then submit**

Where the form's `onSubmit` lives:
```ts
const onSubmit = async (values: RequisitionFormValues) => {
  let payload = values;
  if (values.routingType === 'direct_payment' && !values.payeeId) {
    const persisted = await persistNewPayeeIfNeeded(values, newTypeFromState, createPayee);
    payload = { ...values, payeeId: persisted.payeeId, payeeSnapshot: persisted.payeeSnapshot };
  }
  await submitRequisition(payload);
};
```

(Lift `newTypeFromState` from `PayeeBlock` via context or duplicate the state at form level — pick the cleaner option for the existing form structure.)

- [ ] **Step 4: Typecheck + manual check**

Run: `cd frontend && pnpm typecheck && pnpm lint`
Manual: at `/requisitions/new`, pick Direct Payment → New → fill in name → submit → confirm requisition created with payeeId.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/requisitions/payee-block.tsx frontend/src/app/(private)/(main)/(procurement)/requisitions/new
git commit -m "feat(req-form): payee block with new/vendor/employee/other modes"
```

---

### Task 7.3: Submit wiring + line-item table reuse

**Files:**
- Modify: the requisition form file from Task 7.1

- [ ] **Step 1: Confirm line-item table is shared component**

Run: `grep -rn "LineItemsTable\|RequisitionLines" frontend/src --include='*.tsx' -l`
Identify the existing line-item table. **Reuse as-is** — no styling/component change.

- [ ] **Step 2: Ensure submit POSTs routing_type + payee fields**

The backend `POST /api/requisitions` should accept `routingType`, `payeeId`, `payeeSnapshot` already after Phase 3. If the request DTO doesn't accept them yet, extend the request struct in `backend/handlers/requisition.go` create handler. Add a small failing then passing test in `requisitions_http_test.go`:

```go
func TestCreateRequisition_AcceptsRoutingTypeAndPayee(t *testing.T) {
    app, ctx := setupTestApp(t)
    defer ctx.Cleanup()
    body := `{"routingType":"direct_payment","payeeSnapshot":{"name":"John Doe","payeeType":"employee"},"items":[{"description":"May allowance","quantity":1,"unitPrice":500}]}`
    req := authedReq(t, ctx, http.MethodPost, "/api/requisitions", strings.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    resp, _ := app.Test(req)
    if resp.StatusCode != http.StatusCreated { t.Fatalf("status=%d", resp.StatusCode) }
    var r map[string]any
    json.NewDecoder(resp.Body).Decode(&r)
    if r["routingType"] != "direct_payment" { t.Fatalf("routingType=%v", r["routingType"]) }
}
```

If this fails, extend the DTO and re-run.

- [ ] **Step 3: Typecheck full frontend + backend**

```bash
cd frontend && pnpm typecheck && pnpm lint
cd ../backend && go test ./handlers -run TestCreateRequisition -v
```
Expected: green.

- [ ] **Step 4: Commit**

```bash
git add backend/handlers/requisition.go backend/handlers/requisitions_http_test.go
git commit -m "feat(req): accept routingType + payee on create requisition"
```

---

## Phase 8 — PO + PV list/detail UI

### Task 8.1: Routing-type badges + filter chip on PO list

**Files:**
- Modify: `frontend/src/components/.../purchase-orders-table.tsx` (memory mentions this path)

- [ ] **Step 1: Add routing-type badge**

Find the row renderer. Add next to existing status badge:

```tsx
{po.routingType === 'direct_payment' && (
  <Badge variant="outline" className="border-purple-500 text-purple-700">Direct Payment</Badge>
)}
{po.routingType === 'accounting' && (
  <Badge variant="outline" className="border-amber-500 text-amber-700">Accounting</Badge>
)}
```

- [ ] **Step 2: Add filter chip group above table**

```tsx
const [filter, setFilter] = useState<'all' | 'procurement' | 'accounting' | 'direct_payment'>('all');
// pass filter as query param to the list API

<div className="flex gap-2 mb-3">
  {(['all', 'procurement', 'accounting', 'direct_payment'] as const).map((f) => (
    <Button key={f} variant={filter === f ? 'default' : 'outline'} size="sm" onClick={() => setFilter(f)}>
      {f.replace('_', ' ')}
    </Button>
  ))}
</div>
```

Update the data hook to forward `?routingType=` when the filter isn't `all`. Backend list handler reads the query param and adds `WHERE routing_type = ?` to the existing query.

- [ ] **Step 3: Backend supports the filter**

In `backend/handlers/purchase_order.go` list method, after scope filtering:
```go
if rt := c.Query("routingType"); rt != "" {
    query = query.Where("routing_type = ?", rt)
}
```

- [ ] **Step 4: Typecheck + manual**

Run: `cd frontend && pnpm typecheck && pnpm dev`
Manual: visit `/purchase-orders`, click "Direct Payment" chip, only direct-payment rows shown.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/purchase-orders backend/handlers/purchase_order.go
git commit -m "feat(po-ui): routing-type badge and filter chip on PO list"
```

---

### Task 8.2: PV list badge + auto-created pill

**Files:**
- Modify: `frontend/src/components/.../payment-vouchers-table.tsx` (memory path)

- [ ] **Step 1: Add badge + pill**

```tsx
{pv.routingType === 'direct_payment' && (
  <div className="flex flex-col gap-1">
    <Badge variant="outline" className="border-purple-500 text-purple-700">Direct Payment</Badge>
    {pv.metadata?.autoCreated && (
      <span className="text-xs text-muted-foreground">Auto from REQ-{String(pv.metadata?.sourceReqID ?? '').slice(0, 8)}</span>
    )}
  </div>
)}
```

- [ ] **Step 2: Add 'paid' to status filter options**

Find the status filter list in the PV table — append `'paid'`. Verify pill colour: emerald/green for `paid`.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/payment-vouchers
git commit -m "feat(pv-ui): direct-payment badge, auto-created pill, paid status"
```

---

### Task 8.3: Mark-as-Paid modal

**Files:**
- Create: `frontend/src/components/payment-vouchers/mark-paid-modal.tsx`

- [ ] **Step 1: Build the modal**

```tsx
'use client';

import { useState } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '@/lib/api/client';
import {
  Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter,
} from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { Button } from '@/components/ui/button';
import { toast } from 'sonner';

interface Props {
  pvId: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function MarkPaidModal({ pvId, open, onOpenChange }: Props) {
  const [file, setFile] = useState<File | null>(null);
  const [paidDate, setPaidDate] = useState<string>(new Date().toISOString().slice(0, 10));
  const [notes, setNotes] = useState('');
  const qc = useQueryClient();

  const mutation = useMutation({
    mutationFn: async () => {
      if (!file) throw new Error('Proof of payment required');
      if (file.size > 10 * 1024 * 1024) throw new Error('File exceeds 10MB');
      const allowed = ['application/pdf', 'image/jpeg', 'image/png'];
      if (!allowed.includes(file.type)) throw new Error('Must be PDF, JPG, or PNG');

      const fd = new FormData();
      fd.append('popFile', file);
      fd.append('paidDate', paidDate);
      if (notes) fd.append('notes', notes);
      const res = await apiClient.post(`/api/payment-vouchers/${pvId}/mark-paid`, fd, {
        headers: { 'Content-Type': 'multipart/form-data' },
      });
      return res.data;
    },
    onSuccess: () => {
      toast.success('Marked as paid');
      qc.invalidateQueries({ queryKey: ['payment-vouchers'] });
      qc.invalidateQueries({ queryKey: ['payment-voucher', pvId] });
      onOpenChange(false);
    },
    onError: (e: Error) => toast.error(e.message),
  });

  const canSubmit = !!file && !!paidDate && !mutation.isPending;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Mark as paid</DialogTitle>
          <DialogDescription asChild>
            <div>Upload proof of payment to complete this voucher.</div>
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-4">
          <div>
            <Label>Proof of payment file *</Label>
            <Input
              type="file"
              accept=".pdf,.jpg,.jpeg,.png"
              onChange={(e) => setFile(e.target.files?.[0] ?? null)}
            />
            <p className="text-xs text-muted-foreground mt-1">PDF, JPG or PNG, max 10MB.</p>
          </div>
          <div>
            <Label>Paid date *</Label>
            <Input type="date" value={paidDate} onChange={(e) => setPaidDate(e.target.value)} />
          </div>
          <div>
            <Label>Notes</Label>
            <Textarea value={notes} onChange={(e) => setNotes(e.target.value)} placeholder="Optional" />
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)} disabled={mutation.isPending}>Cancel</Button>
          <Button onClick={() => mutation.mutate()} disabled={!canSubmit}>
            {mutation.isPending ? 'Submitting…' : 'Confirm paid'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
```

- [ ] **Step 2: Wire trigger in PV detail page**

Find the PV detail page (`frontend/src/app/(private)/(main)/.../payment-vouchers/[id]/page.tsx` — exact path TBD via grep). Add:

```tsx
const [markPaidOpen, setMarkPaidOpen] = useState(false);
// ...
{pv.status === 'approved' && ['finance', 'admin'].includes(userRole) && (
  <>
    <Button onClick={() => setMarkPaidOpen(true)}>Mark as Paid</Button>
    <MarkPaidModal pvId={pv.id} open={markPaidOpen} onOpenChange={setMarkPaidOpen} />
  </>
)}
```

> `userRole` is read from the same session/permissions hook used elsewhere in the PV detail page — `useSession()` or `usePermissions()` per memory. If the existing PV detail uses a different gating pattern (e.g. `FINANCE_EDIT_ROLES.includes(userRole)` mentioned in memory for `payment-vouchers-table.tsx`), reuse that exact constant.

- [ ] **Step 3: Typecheck + manual**

Run: `cd frontend && pnpm typecheck && pnpm dev`
Manual: navigate to an approved PV → Mark as Paid → upload file → submit → status pill flips to `paid`.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/components/payment-vouchers/mark-paid-modal.tsx frontend/src/app/(private)/(main)
git commit -m "feat(pv-ui): mark-paid modal with proof-of-payment upload"
```

---

## Phase 9 — Nav rename + dashboard widgets

### Task 9.1: Sidebar nav rename

**Files:**
- Modify: `frontend/src/components/layout/sidebar/nav-main.tsx`

- [ ] **Step 1: Change title**

In `frontend/src/components/layout/sidebar/nav-main.tsx:95`, change:
```tsx
title: "Procurement",
```
to:
```tsx
title: "Source to Pay",
```

- [ ] **Step 2: Manual check**

Run: `cd frontend && pnpm dev`
Expected: sidebar group reads "Source to Pay"; collapsed sidebar tooltip shows "Source to Pay".

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/layout/sidebar/nav-main.tsx
git commit -m "feat(nav): rename Procurement to Source to Pay"
```

---

### Task 9.2: Finance "Awaiting Payment" widget

**Files:**
- Create: `frontend/src/app/(private)/(main)/home/_components/awaiting-payment.tsx`
- Modify: `frontend/src/app/(private)/(main)/home/_components/dashboard-client.tsx`

- [ ] **Step 1: Backend endpoint or reuse PV list**

Use the existing PV list endpoint with query `?status=approved&hasProofOfPayment=false`. Backend list handler in `payment_voucher.go` needs to support `hasProofOfPayment` query param:

```go
if hpp := c.Query("hasProofOfPayment"); hpp == "false" {
    query = query.Where("proof_of_payment IS NULL")
}
```

- [ ] **Step 2: Build widget**

```tsx
'use client';
import { useQuery } from '@tanstack/react-query';
import { apiClient } from '@/lib/api/client';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import Link from 'next/link';

export function AwaitingPaymentWidget() {
  const { data } = useQuery({
    queryKey: ['pv', 'awaiting-payment'],
    queryFn: async () => {
      const res = await apiClient.get('/api/payment-vouchers?status=approved&hasProofOfPayment=false&limit=10');
      return res.data.items as any[];
    },
  });

  return (
    <Card>
      <CardHeader>
        <CardTitle>Awaiting payment</CardTitle>
      </CardHeader>
      <CardContent className="space-y-2">
        {(data ?? []).length === 0 && <p className="text-sm text-muted-foreground">All caught up.</p>}
        {(data ?? []).map((pv) => (
          <Link key={pv.id} href={`/payment-vouchers/${pv.id}`} className="flex justify-between border-b py-2 last:border-0 hover:bg-muted/40 px-2 -mx-2 rounded">
            <span className="font-mono text-sm">{pv.documentNumber}</span>
            <span className="text-sm">{pv.vendorName}</span>
            <span className="text-sm">{pv.amount}</span>
          </Link>
        ))}
      </CardContent>
    </Card>
  );
}
```

- [ ] **Step 3: Wire into dashboard for finance variant**

In `dashboard-client.tsx`, when `variant === 'admin' || variant === 'finance'`, render `<AwaitingPaymentWidget />`. (Memory notes the dashboard uses `getDashboardVariant`; respect that pattern.)

- [ ] **Step 4: Manual + commit**

Run: `cd frontend && pnpm dev` → as finance user, dashboard shows widget.

```bash
git add frontend/src/app/(private)/(main)/home backend/handlers/payment_voucher.go
git commit -m "feat(dashboard): finance awaiting-payment widget"
```

---

### Task 9.3: Requester pipeline visual

**Files:**
- Modify: `frontend/src/app/(private)/(main)/home/_components/recent-tasks.tsx` (or create a new pipeline component next to it)

- [ ] **Step 1: Pipeline component**

```tsx
function currentStepIndex(req: any): number {
  // 0 = req submitted, 1 = PO created, 2 = PV approved, 3 = paid
  const pvStatus = req.linkedPV?.status;
  if (pvStatus === 'paid') return 3;
  if (pvStatus === 'approved') return 2;
  if (req.linkedPO?.id) return 1;
  return 0;
}

function PipelineRow({ req }: { req: any }) {
  const steps = ['Req submitted', 'PO created', 'PV approved', 'Paid'];
  const idx = currentStepIndex(req);
  return (
    <div className="flex items-center gap-2 py-2">
      <span className="font-mono text-sm">{req.documentNumber}</span>
      <div className="flex-1 grid grid-cols-4 gap-1">
        {steps.map((s, i) => (
          <div key={s} className={`h-2 rounded ${i <= idx ? 'bg-emerald-500' : 'bg-muted'}`} title={s} />
        ))}
      </div>
    </div>
  );
}
```

`req.linkedPO` and `req.linkedPV` are included in the requisition list response payload — if not, extend the list serializer in `backend/handlers/requisition.go` to include them.

- [ ] **Step 2: Render only requesters' direct-payment requisitions**

Filter `useRequisitions({ paymentType: 'direct_payment', requester: userId })`. Append below existing recent tasks.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/app/(private)/(main)/home/_components
git commit -m "feat(dashboard): requester direct-payment pipeline visual"
```

---

## Phase 10 — Workflow admin UI

### Task 10.1: Workflow editor routing-type picker

**Files:**
- Modify: workflow editor form (locate via `grep -rn "routingType" frontend/src --include='*.tsx' -l` and pick the admin editor)

- [ ] **Step 1: Replace text input with select**

```tsx
<FormField
  name="conditions.routingType"
  render={({ field }) => (
    <FormItem>
      <FormLabel>Routing type</FormLabel>
      <Select value={field.value} onValueChange={field.onChange}>
        <SelectTrigger><SelectValue placeholder="Pick routing type" /></SelectTrigger>
        <SelectContent>
          <SelectItem value="procurement">Procurement</SelectItem>
          <SelectItem value="accounting">Accounting</SelectItem>
          <SelectItem value="direct_payment">Direct Payment</SelectItem>
        </SelectContent>
      </Select>
    </FormItem>
  )}
/>
```

- [ ] **Step 2: When direct_payment selected, force flags + hide stages**

```tsx
const routing = form.watch('conditions.routingType');
useEffect(() => {
  if (routing === 'direct_payment') {
    form.setValue('conditions.autoApprove', true);
    form.setValue('conditions.autoGeneratePO', true);
    form.setValue('conditions.autoApprovePO', true);
    form.setValue('stages', []);
  }
}, [routing]);

{routing !== 'direct_payment' && <StagesEditor />}
{routing === 'direct_payment' && (
  <Alert>Direct payment workflows skip the approval workflow and auto-create an approved PO and a draft PV. No stages required.</Alert>
)}
```

- [ ] **Step 3: Manual check + commit**

Manual: at workflow admin, create workflow → pick Direct Payment → save → confirm round-trip persists the routing type.

```bash
git add frontend/src/app/(private)/.../workflows
git commit -m "feat(workflow-admin): routing type picker with direct payment option"
```

---

## Phase 11 — E2E + final verification

### Task 11.1: Playwright happy-path E2E

**Files:**
- Create: `frontend/e2e/direct-payment-flow.spec.ts`

- [ ] **Step 1: Write E2E**

```ts
import { test, expect } from '@playwright/test';

test('direct payment happy path', async ({ page }) => {
  // Login as requester (test fixture or helper)
  await loginAs(page, 'requester');

  // Submit direct-payment requisition
  await page.goto('/requisitions/new');
  await page.getByRole('radio', { name: /direct payment/i }).check();
  await page.getByRole('radio', { name: /^new$/i }).check();
  await page.getByLabel('Name *').fill('John Doe');
  await page.getByLabel('Bank Account').fill('1234567890');
  // line item
  await page.getByLabel('Description').fill('May allowance');
  await page.getByLabel('Quantity').fill('1');
  await page.getByLabel('Unit Price').fill('500');
  await page.getByRole('button', { name: /submit/i }).click();
  await expect(page.getByText(/created successfully/i)).toBeVisible();

  // Switch to finance — find auto-PV
  await loginAs(page, 'finance');
  await page.goto('/payment-vouchers');
  const row = page.getByRole('row', { name: /John Doe/ }).first();
  await expect(row).toBeVisible();
  await row.click();

  // Submit PV for approval (placeholder — use existing submit action)
  await page.getByRole('button', { name: /submit for approval/i }).click();

  // Approve PV (would need approver login)
  await loginAs(page, 'approver');
  await page.goto('/approvals');
  await page.getByRole('button', { name: /approve/i }).first().click();

  // Back to finance, mark paid
  await loginAs(page, 'finance');
  await page.goto('/payment-vouchers');
  await page.getByRole('row', { name: /John Doe/ }).first().click();
  await page.getByRole('button', { name: /mark as paid/i }).click();
  await page.setInputFiles('input[type="file"]', 'frontend/e2e/fixtures/pop.pdf');
  await page.getByRole('button', { name: /confirm paid/i }).click();
  await expect(page.getByText(/paid/i)).toBeVisible();
});

test('procurement role cannot see direct payment chain', async ({ page }) => {
  await loginAs(page, 'procurement');
  await page.goto('/purchase-orders');
  await expect(page.getByText(/direct payment/i)).toHaveCount(0);
  await page.goto('/payment-vouchers');
  await expect(page.getByText(/direct payment/i)).toHaveCount(0);
});
```

> `loginAs` is the existing test helper. Match its signature exactly.
> Create `frontend/e2e/fixtures/pop.pdf` (any small PDF) for the upload step.

- [ ] **Step 2: Run E2E**

Run: `cd frontend && pnpm playwright test direct-payment-flow`
Expected: PASS.

- [ ] **Step 3: Commit**

```bash
git add frontend/e2e/direct-payment-flow.spec.ts frontend/e2e/fixtures/pop.pdf
git commit -m "test(e2e): direct payment happy path and procurement visibility"
```

---

### Task 11.2: Final full-suite run + verification

- [ ] **Step 1: Backend full test suite**

Run: `cd backend && go test ./... -count=1`
Expected: all green.

- [ ] **Step 2: Frontend typecheck + lint + unit**

Run: `cd frontend && pnpm typecheck && pnpm lint && pnpm test`
Expected: all green.

- [ ] **Step 3: Migration up→down→up roundtrip on a fresh DB**

```bash
goose -dir backend/database/migrations postgres "$DATABASE_URL_TEST" reset
goose -dir backend/database/migrations postgres "$DATABASE_URL_TEST" up
goose -dir backend/database/migrations postgres "$DATABASE_URL_TEST" down
goose -dir backend/database/migrations postgres "$DATABASE_URL_TEST" up
```
Expected: clean each pass.

- [ ] **Step 4: Manual smoke**

Visit dev environment:
1. Workflow admin → create Direct Payment workflow.
2. Requester → submit direct-payment requisition with new payee.
3. Finance dashboard → "Awaiting payment" widget shows the auto-PV after PV approval.
4. Mark paid via modal → POP upload succeeds, status → `paid`.
5. As procurement role → confirm none of the above documents appear in any list.

- [ ] **Step 5: Final commit**

Nothing to commit beyond the prior tasks. Tag the release:
```bash
git tag direct-payment-v1
```

---

## Summary of files touched

**Backend:**
- `backend/database/migrations/018_direct_payment.up.sql` + `.down.sql`
- `backend/models/models.go` (Payee model, routing_type + PoP fields)
- `backend/models/enhanced_auth.go` (RoutingType constant)
- `backend/services/workflow_service.go` (+ test)
- `backend/services/workflow_execution_service.go` (+ test)
- `backend/utils/document_scope.go` (+ test)
- `backend/handlers/payee.go` (+ http test)
- `backend/handlers/payment_voucher.go` (+ http test, mark-paid + recovery)
- `backend/handlers/purchase_order.go` (+ http test, routing_type query filter)
- `backend/handlers/requisition.go` (+ http test, accept routing/payee fields)
- `backend/handlers/grn.go` (scope filter)
- `backend/handlers/handler_registry.go` (route registration)

**Frontend:**
- `frontend/src/types/payee.ts` (new)
- `frontend/src/types/requisition.ts`, `payment-voucher.ts`, `purchase-order.ts`
- `frontend/src/lib/api/payees.ts` (new)
- `frontend/src/hooks/use-payees.ts` (new)
- `frontend/src/components/requisitions/payee-block.tsx` (new)
- `frontend/src/components/payment-vouchers/mark-paid-modal.tsx` (new)
- Requisition create form, PO + PV table components
- Workflow admin editor
- `frontend/src/components/layout/sidebar/nav-main.tsx` (rename)
- `frontend/src/app/(private)/(main)/home/_components/awaiting-payment.tsx` (new)
- Dashboard client integration
- `frontend/e2e/direct-payment-flow.spec.ts` (new)
