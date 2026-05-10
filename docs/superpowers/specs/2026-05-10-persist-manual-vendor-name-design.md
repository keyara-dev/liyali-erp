# Persist Manual Vendor Name on Documents — Design

**Date:** 2026-05-10
**Status:** Approved
**Scope:** PurchaseOrder, PaymentVoucher, Requisition

## Problem

When a user adds a quotation with a manually typed vendor name (vendor not in system, `vendorId` empty), and selects that quotation as the supplier for a Purchase Order, the vendor name is lost on save. The submit-for-approval dialog then blocks with *"You must select a vendor before submitting"* even though a quotation is selected.

Reproduction:
1. Create PO from requisition.
2. Add quotation: pick *"None (Enter Manually)"*, type vendor name, enter amount.
3. Click *Select* on that quotation.
4. PO header shows vendor as `—`. Submit dialog blocks.

## Root Cause

Three independent defects in the backend vendor-name pipeline:

1. **`models.PurchaseOrder.VendorName`** is tagged `gorm:"-"` — never persisted.
2. **`models.Requisition.PreferredVendorName`** is tagged `gorm:"-"` — never persisted.
3. **`models.PaymentVoucher.VendorName`** is *not* `gorm:"-"`, but the underlying `vendor_name` DB column does not exist in any migration. Insert/update of PV rows with a non-empty `VendorName` likely errors silently (or is masked by all real-world PVs having a `vendor_id`).
4. **Response builders** (`modelToPurchaseOrderResponse`, `modelToPaymentVoucherResponse`, requisition equivalent) read vendor name only from the `Vendor` GORM relation. When `VendorID` is null, no name is surfaced.

Frontend validation already accepts `vendorId || vendorName` (`purchase-order-submit-dialog.tsx:48-50`). The bug is purely backend.

GRN inherits vendor from its linked PO — it has no own `vendor_id` / `vendor_name`. Fixing PO automatically fixes GRN downstream.

## Design

### 1. Migration `017_persist_vendor_name`

Add nullable `VARCHAR(255)` columns to three tables:

```sql
-- 017_persist_vendor_name.up.sql
ALTER TABLE purchase_orders   ADD COLUMN IF NOT EXISTS vendor_name           VARCHAR(255);
ALTER TABLE payment_vouchers  ADD COLUMN IF NOT EXISTS vendor_name           VARCHAR(255);
ALTER TABLE requisitions      ADD COLUMN IF NOT EXISTS preferred_vendor_name VARCHAR(255);
```

```sql
-- 017_persist_vendor_name.down.sql
ALTER TABLE purchase_orders   DROP COLUMN IF EXISTS vendor_name;
ALTER TABLE payment_vouchers  DROP COLUMN IF EXISTS vendor_name;
ALTER TABLE requisitions      DROP COLUMN IF EXISTS preferred_vendor_name;
```

No backfill — existing rows with linked vendors keep working via `Vendor.Name`. Pre-fix manual-vendor PO drafts with lost names are not recoverable.

### 2. Model changes (`backend/models/models.go`)

Remove the `gorm:"-"` directive (and update the misleading "Computed from Vendor.Name" comment to "Persisted; falls back to Vendor.Name on read"):

```go
// Requisition (line 67)
PreferredVendorName string `json:"preferredVendorName"`

// PurchaseOrder (line 146)
VendorName string `json:"vendorName,omitempty"`

// PaymentVoucher (line 209) — already untagged, leave as is
VendorName string `json:"vendorName,omitempty"`
```

### 3. Create handlers — set `VendorName` on initial model

**`backend/handlers/purchase_order.go` `CreatePurchaseOrder`** (around line 276):

```go
order := models.PurchaseOrder{
    ...
    VendorID:   vendorIDPtr,
    VendorName: req.VendorName,   // NEW — persists manual name
    ...
}
```

**`backend/handlers/payment_voucher.go` `CreatePaymentVoucher`** — audit and add `VendorName: req.VendorName` if missing.

**`backend/handlers/requisition.go` `CreateRequisition`** — add `PreferredVendorName: req.PreferredVendorName` if missing.

### 4. Update handlers — accept `VendorName` updates

**PO update handler** (`backend/handlers/purchase_order.go` around line 504):

```go
if req.VendorID != "" {
    fromVendorID := ""
    if order.VendorID != nil { fromVendorID = *order.VendorID }
    if fromVendorID != req.VendorID {
        changes["vendorId"] = map[string]string{"old": fromVendorID, "new": req.VendorID}
    }
    order.VendorID = &req.VendorID
}
// NEW: persist vendor name when sent alongside a vendor change
if req.VendorID != "" || req.VendorName != "" {
    if order.VendorName != req.VendorName {
        changes["vendorName"] = map[string]string{"old": order.VendorName, "new": req.VendorName}
    }
    order.VendorName = req.VendorName
}
```

The guard `req.VendorID != "" || req.VendorName != ""` ensures pure metadata-only updates (quotation list edits) do not clobber the stored vendor name to empty.

**`isMetadataOnly` guard** (line 467) keeps checking only `req.VendorID == ""` — `VendorName` is not part of the gate. Quotation-list saves still bypass the status check.

Apply the same pattern to PV update handler and Requisition update handler (for `PreferredVendorName`).

### 5. Response builders — fallback chain

**`modelToPurchaseOrderResponse`** (line 762):

```go
vendorName := order.VendorName            // stored value
var vendorResp *types.VendorResponse
if order.Vendor != nil {
    vendorName = order.Vendor.Name        // canonical wins
    vr := modelToVendorResponse(*order.Vendor)
    vendorResp = &vr
}
```

Apply equivalent change to `modelToPaymentVoucherResponse` and the requisition response builder.

### 6. Frontend

No changes. `frontend/src/app/(private)/(main)/purchase-orders/_components/purchase-order-submit-dialog.tsx` already gates on `vendorId?.trim() || vendorName?.trim()`. Once backend returns `vendorName`, the gate passes.

`handleSelectVendor` in `purchase-order-detail-client.tsx:273` already sends `vendorName` in the `updatePurchaseOrder` call. No change needed.

## Affected files

**Backend**
- `backend/database/migrations/017_persist_vendor_name.up.sql` (new)
- `backend/database/migrations/017_persist_vendor_name.down.sql` (new)
- `backend/models/models.go` — remove two `gorm:"-"` tags, update comments
- `backend/handlers/purchase_order.go` — create + update handlers + response builder
- `backend/handlers/payment_voucher.go` — create + update handlers + response builder
- `backend/handlers/requisition.go` — create + update handlers + response builder

**Frontend** — none.

**Tests** — add coverage:
- PO create with `vendorName` only (no `vendorId`) → name persists, response returns it.
- PO update from `vendorId` → manual `vendorName` → response surfaces manual name, vendor relation cleared.
- PO submit-for-approval succeeds when only `vendorName` set (integration via existing submit handler — no change there but verify no regression).
- PV equivalent.
- Requisition equivalent for `preferredVendorName`.

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Existing PV rows accumulated GORM write errors due to missing column | Migration adds the column; subsequent writes succeed |
| Vendor name diverges from `Vendor.Name` if vendor renamed in master data | Response builder always prefers `Vendor.Name` when relation present — stored column only used as fallback when `VendorID` null |
| Update handler clobbers stored name on metadata-only PUT | Guarded by `req.VendorID != "" || req.VendorName != ""` — pure metadata saves leave name alone |
| Audit/snapshot code references `vendorName` | Already uses the response field (`AUDIT_SNAPSHOT_IMPLEMENTATION.md` references `qMap["vendorName"]` from quotations metadata, unrelated) — no change needed |

## Out of Scope

- GRN model changes (no own vendor field).
- Backfilling lost vendor names on pre-existing drafts.
- Vendor master-data validation when `vendorName` provided without `vendorId` (deliberate — supports ad-hoc vendors).
- Changes to PDF generation (it already reads `vendorName` from response).
