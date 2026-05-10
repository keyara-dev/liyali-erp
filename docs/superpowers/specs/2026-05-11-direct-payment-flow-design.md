# Direct Payment Flow — Design Spec

**Date:** 2026-05-11
**Status:** Approved (brainstorming phase)
**Author:** Thompson Manda + Claude

## Purpose

Add a third workflow routing path — **Direct Payment** — for strict finance payouts to individuals (wages, allowances, ad-hoc payments). These bypass procurement entirely while preserving an auto-generated Purchase Order for audit, and auto-spawn a draft Payment Voucher for the finance team to approve and pay.

The procurement-team experience is unchanged. The accounting-route experience is unchanged. Direct payment is a new, parallel path.

Also rename the sidebar navigation group from **Procurement** to **Source to Pay** to reflect the expanded scope.

## Background

Today the system supports two workflow routing types declared on `Workflow.Conditions.RoutingType`:

- `procurement` — requester submits, procurement team manually creates a PO, GRN is recorded on receipt, PV is created against the PO, finance approves PV.
- `accounting` — requester submits, on approval the backend auto-generates an approved PO (`autoApproveAndGeneratePO`), and the rest of the chain (PV, optional GRN) is manual.

Procurement-role users today see **every** PO/PV regardless of routing type — there is no routing-aware visibility filter on the document scope. This is the gap that direct payment closes.

The existing `payment_first` procurement flow on Payment Vouchers already removes the GRN gate on PV approval; direct payment reuses that flag.

## Goals

1. Allow requesters to create a requisition that resolves to a workflow with `routingType=direct_payment`.
2. On approval, auto-create an approved PO (for audit) **and** an auto-draft PV (for finance to action).
3. Hide every artifact (requisition, PO, PV) of a direct-payment chain from procurement-role users.
4. Capture the payee — vendor, employee, or arbitrary individual/business — at requisition time, with persistence so future requests can reuse them.
5. Require a proof-of-payment file upload before a PV can be marked `paid` and the chain closed.
6. Rename the sidebar group to **Source to Pay**.

## Non-Goals

- Tax computation / withholding logic on payee records.
- Bulk payment runs (multiple payees in one PV).
- Bank integration / automated payouts. POP upload is the manual evidence step.
- Migration of existing `accounting`-route documents to `direct_payment`. Existing data stays as-is.

## Architecture Overview

```text
[Requester] ──submits──> Requisition (routing_type=direct_payment)
                            │
                            ▼  on approval (0 stages, auto-approve)
                ┌────────────────────────────┐
                │ autoApproveAndGeneratePO   │  → PO (status=approved, routing_type=direct_payment)
                │ autoCreateDraftPV          │  → PV (status=draft, routing_type=direct_payment, owner=finance)
                └────────────────────────────┘
                            │
                            ▼
[Finance] ──edits + submits──> PV approval workflow ──> PV.status=approved
                            │
                            ▼  upload POP file + paid_date
[Finance] ──MarkAsPaid──> PV.status=paid, requisition.status=completed
```

Procurement-role users never see any row in this chain — `DocumentScope` filters them out at query time.

## Data Model

### New table: `payees`

```sql
CREATE TABLE payees (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id),
    payee_type      TEXT NOT NULL CHECK (payee_type IN ('vendor','employee','other')),
    name            TEXT NOT NULL,
    email           TEXT,
    phone           TEXT,
    bank_name       TEXT,
    bank_account    TEXT,
    tax_id          TEXT,
    source_vendor_id UUID NULL REFERENCES vendors(id),
    source_user_id   UUID NULL REFERENCES users(id),
    created_by      UUID REFERENCES users(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_payees_org_type_name ON payees (organization_id, payee_type, name);
CREATE INDEX idx_payees_name_trgm ON payees USING gin (name gin_trgm_ops);
```

**Semantics:** the `payees` table is the canonical lookup for the requisition payee picker. Vendor and Employee picks resolve via the source FK columns but a row is also snapshotted into `payees` on first reference so the payee dropdowns return a unified list. "Other" entries (one-off individuals, sole traders) live exclusively in this table.

### Requisition additions

```sql
ALTER TABLE requisitions
    ADD COLUMN routing_type   TEXT NOT NULL DEFAULT 'procurement',
    ADD COLUMN payee_id       UUID REFERENCES payees(id),
    ADD COLUMN payee_snapshot JSONB;

CREATE INDEX idx_requisitions_routing_type_org ON requisitions (organization_id, routing_type);
```

- `routing_type` values: `procurement | accounting | direct_payment`. Denormalized at submission from `Workflow.Conditions.RoutingType` to avoid joins on hot list queries.
- Default `procurement` chosen for the column default because the bulk of historical data is procurement-routed; the migration also runs a one-time backfill (see Migrations below) that joins each row to its workflow to set the accurate value.
- `payee_snapshot` freezes the payee details at submission time so later edits to the `payees` row don't retroactively change historical documents.

### PO + PV additions

```sql
ALTER TABLE purchase_orders   ADD COLUMN routing_type TEXT NOT NULL DEFAULT 'procurement';
ALTER TABLE payment_vouchers  ADD COLUMN routing_type TEXT NOT NULL DEFAULT 'procurement',
                              ADD COLUMN proof_of_payment JSONB,
                              ADD COLUMN paid_at TIMESTAMPTZ,
                              ADD COLUMN paid_by UUID REFERENCES users(id);

CREATE INDEX idx_purchase_orders_routing_type_org  ON purchase_orders  (organization_id, routing_type);
CREATE INDEX idx_payment_vouchers_routing_type_org ON payment_vouchers (organization_id, routing_type);
```

`routing_type` propagates from requisition → PO → PV at creation time. `proof_of_payment` JSONB shape:

```json
{
  "fileId": "uuid",
  "fileName": "bank-slip.pdf",
  "fileUrl": "...",
  "mimeType": "application/pdf",
  "uploadedBy": "uuid",
  "uploadedAt": "2026-05-11T14:00:00Z"
}
```

### Workflow conditions

`WorkflowConditions.RoutingType` enum gains `direct_payment`. No new flags. The presence of `direct_payment` implies:

- `AutoApprove: true` (workflow auto-approves the requisition itself)
- `AutoGeneratePO: true`
- `AutoApprovePO: true`
- Always auto-create a draft PV
- Always require POP to close
- `ProcurementFlow` on the resulting PV is `payment_first` (so no GRN gate)

Backend validation rejects saving a workflow with `RoutingType=direct_payment` that has stages > 0.

### PV status enum

Add terminal value `paid`. Transitions:

```text
draft → submitted → approved → paid
        │             │
        └→ rejected   └→ rejected
```

`approved → paid` requires `proof_of_payment IS NOT NULL` and is invoked via the new `MarkPaymentVoucherPaid` handler.

### Migrations

Single Goose migration `NNNN_direct_payment.sql` (up + down):

1. Add columns with default `routing_type='procurement'`.
2. One-time backfill: `UPDATE requisitions SET routing_type = w.conditions->>'routingType' FROM workflows w WHERE requisitions.workflow_id = w.id AND w.conditions->>'routingType' IS NOT NULL;` — repeat for `purchase_orders` (via source requisition) and `payment_vouchers` (via linked PO). Rows still on `procurement` after backfill stay correct since `procurement` is the default and matches the most common case.
3. Create `payees` table + indexes (including `pg_trgm` extension if not already enabled).
4. Add `proof_of_payment`, `paid_at`, `paid_by` to `payment_vouchers`.

Down migration drops the new columns + `payees` table. No destructive effect on historical procurement/accounting rows.

## Backend Flow

### Submission (`workflow_execution_service.go`)

`SubmitRequisitionWithRouting` extends:

```text
1. Resolve workflow, read Conditions.RoutingType
2. Write requisition.routing_type = Conditions.RoutingType (denormalized)
3. If routing_type == 'direct_payment':
     - validate payee_id OR payee_snapshot present
     - validate workflow has 0 stages
4. Existing auto-approve / auto-PO path runs for accounting AND direct_payment
5. If routing_type == 'direct_payment':
     - call autoCreateDraftPV(req, po)
6. Return SubmitRoutingResult{ AutoCreatedPOID, AutoCreatedPVID, RoutingType }
```

### `autoCreateDraftPV(req, po)` (new)

```go
pv := &models.PaymentVoucher{
    Status:          "draft",
    CreatedBy:       systemUserID,
    AssignedRole:    "finance",
    LinkedPO:        po.Number,
    RoutingType:     "direct_payment",
    VendorName:      req.PayeeSnapshot.Name,
    Amount:          req.TotalAmount,
    ProcurementFlow: "payment_first",
    Metadata:        models.JSON{"autoCreated": true, "sourceReqID": req.ID, "payeeSnapshot": req.PayeeSnapshot},
}
```

Failure path: if PV creation fails after PO succeeded, the PO is **not rolled back** (audit trail intact). The failure is logged + an admin alert raised; finance recovers via the manual recovery endpoint (below).

### Mark-as-Paid handler

`POST /api/payment-vouchers/:id/mark-paid` (multipart, `popFile` + `paidDate` + optional `notes`):

```text
- Auth: finance or admin role
- Validate: pv.status == 'approved'
- Validate: popFile present, size ≤ 10MB, mime in {application/pdf, image/jpeg, image/png}
- Upload file via existing attachment service → fileId, fileUrl
- Set pv.proof_of_payment = {...}, pv.paid_at = paidDate, pv.paid_by = userID, pv.status = 'paid'
- Cascade: requisition.status = 'completed' (only for direct_payment chain)
- Return updated PV
```

### Recovery endpoint

`POST /api/payment-vouchers/recover-from-po/:poId` (admin only) — manually creates a draft PV linked to an existing auto-PO when `autoCreateDraftPV` failed. Idempotent: returns existing PV if one already links to the PO.

### Payee CRUD

```text
GET    /api/payees?type=vendor|employee|other&q=...   list + search (trigram)
POST   /api/payees                                     create (used by "New" radio)
GET    /api/payees/:id                                 single
PUT    /api/payees/:id                                 update (does not retroactively affect snapshots)
DELETE /api/payees/:id                                 soft-delete only; snapshots still valid
```

Vendor / Employee dropdowns hit existing endpoints; "Other" hits `/api/payees?type=other`; "New" creates via POST then proceeds.

### Document scope filter

`DocumentScope` gets one new derived field:

```go
HideDirectPayment bool  // true when role==procurement AND no finance/admin permissions
```

`ApplyToQuery` appends `AND routing_type != 'direct_payment'` when `HideDirectPayment` is true. Applied at the **list and single-get** handlers for: `requisition.go`, `purchase_order.go`, `payment_voucher.go`. Procurement users hitting a direct-URL of a direct-payment doc receive a 404 (no info leak about existence).

### Workflow admin UI (backend side)

The workflow editor exposes routing type as a 3-option picker: `procurement | accounting | direct_payment`. Selecting `direct_payment` disables the stages section and forces `AutoApprove=AutoGeneratePO=AutoApprovePO=true` in the saved conditions. The other auto-flags are hidden for this routing type.

## Frontend Flow

### Requisition create form

New top-level radio at the top of the form: **Request type**

- **Goods / Services** (default) — existing form
- **Direct Payment** — wages, allowances, individual payouts

When `Direct Payment` is selected:

- Hide the vendor section
- Show the **Payee block** with a 4-option radio (`New | Vendor | Employee | Other`):
  - **New** — free-text {name, email, phone, bank name, bank account, tax id} + a sub-radio for payee type {employee | vendor | other}. On submit, POSTs to `/api/payees` and stores the returned ID.
  - **Vendor** — Combobox sourced from existing vendor lookup.
  - **Employee** — Combobox sourced from existing users (role-filtered or all users in org — keep all for now).
  - **Other** — Combobox sourced from `/api/payees?type=other`.
- Reuse the **same line-item table** used for goods/services (description, qty, unit price, total). No specialized payment-line component; consistency over specialization.

The hidden field `paymentType` (frontend-side name; maps to backend `routing_type`) is written based on the radio selection. The backend resolves which `direct_payment` workflow matches at submission time.

Validation: Direct-payment requisitions must have at least one line item with non-zero amount; payee block must resolve to either a `payeeId` or a complete `payeeSnapshot` (the "New" path always produces a `payeeId` after the POST).

### Requisition detail page

- "Direct Payment" badge in the header.
- Payee card replaces the vendor card. Shows: name, payee type, masked bank account (last 4 digits), tax id.
- Linked documents section with status pills, e.g. `REQ-2026-001 ▸ PO-2026-014 (approved) ▸ PV-2026-022 (draft → submitted → approved → paid)`.

### Purchase Order list / detail

- Direct-payment POs are visible to finance + admin only (server-side filter).
- Row gets a "Direct Payment" badge (use a distinct color from the accounting badge — recommend purple).
- Filter chip group above the table: `[All] [Procurement] [Accounting] [Direct Payment]`.
- Procurement users see neither the badge column nor the filter chip — rows are simply absent.

### Payment Voucher list / detail

- Direct-payment PV row gets the badge + a smaller pill: `Auto-created from REQ-XXXX`.
- New `Mark as Paid` action appears only when `pv.status === 'approved'` and user is finance/admin.
  - Modal contents:
    - File upload (POP) — required, accept `.pdf,.jpg,.jpeg,.png`, max 10MB
    - Paid date — required, default today
    - Notes — optional textarea
    - Submit button disabled until file + date present
  - On success: status pill → `paid`, modal closes, list refetches via React Query invalidation.
- New status filter option `paid` joins the existing filter set.

### Sidebar nav rename

`frontend/src/components/layout/sidebar/nav-main.tsx:95`:

- Group title `Procurement` → `Source to Pay`.
- Group icon: keep cart icon (no swap — visual continuity for existing users).
- Child items unchanged: Requisitions, Purchase Orders, Goods Received Notes, Payment Vouchers.

### Dashboard widgets

- Finance dashboard: new **Awaiting Payment** widget — lists PVs where `status='approved' AND proof_of_payment IS NULL`. Click row → opens PV detail with mark-as-paid CTA prominent.
- Requester dashboard: a pipeline visual for the requester's direct-payment requisitions: `Req → PO → PV → Paid`, with the current step highlighted.

### Type updates

```ts
type RoutingType = 'procurement' | 'accounting' | 'direct_payment'; // denormalized doc-level

interface Requisition {
  routingType: RoutingType;
  payeeId?: string;
  payeeSnapshot?: Payee;
  // ...
}

interface Payee {
  id: string;
  organizationId: string;
  payeeType: 'vendor' | 'employee' | 'other';
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

type PaymentVoucherStatus = 'draft' | 'submitted' | 'approved' | 'rejected' | 'paid';

interface PaymentVoucher {
  routingType: RoutingType;
  proofOfPayment?: ProofOfPayment;
  paidAt?: string;
  paidBy?: string;
  // ...
}

interface ProofOfPayment {
  fileId: string;
  fileName: string;
  fileUrl: string;
  mimeType: string;
  uploadedBy: string;
  uploadedAt: string;
}
```

## Edge Cases

| Case | Behavior |
| --- | --- |
| Auto-PO creation fails after req approval | Requisition status rolls back to `pending`; no PV attempted. Existing behavior. |
| Auto-PV creation fails after PO success | PO stays approved (audit intact). Failure logged + admin alert. Manual recovery via `POST /api/payment-vouchers/recover-from-po/:poId`. |
| Payee deleted from `payees` table mid-flow | `payee_snapshot` JSON on the requisition is authoritative. Downstream PO/PV continue to render correctly. |
| POP file too large or wrong mime type | Client-side check (≤10MB, PDF/JPG/PNG) + server-side enforcement. Reject with explicit error. |
| Manual PV not from direct-payment flow | `Mark as Paid` action does not appear (it is keyed to `routing_type === 'direct_payment'`). Existing manual PV completion path unchanged. |
| Procurement user hits direct URL `/payment-vouchers/{id}` for a direct-payment PV | 404 from handler (scope filter excludes the row at single-get). |
| Workflow misconfig: `routingType=direct_payment` saved with stages > 0 | Backend validation rejects save with `"direct_payment workflows must have 0 approval stages"`. |
| Requester picks "Direct Payment" but no matching workflow exists in org | Submission fails with `"No direct payment workflow configured for this organization. Contact admin."` |
| New payee created during requisition then submission fails | Payee row remains in `payees` table — acceptable, no orphan cleanup needed; can be reused next time. |

## Testing Strategy

### Backend (Go, table-driven)

- `payees_test.go` — CRUD, search by trigram, dedup behavior on name+type+org.
- `workflow_execution_service_test.go`:
  - direct-payment requisition → auto-PO → auto-PV chain creates with correct ownership and routing_type values.
  - `accounting`-route requisition is unchanged (no PV auto-created).
  - PV auto-creation failure leaves PO intact; failure flag set.
- `document_scope_test.go` — for each entity (requisition, PO, PV), procurement-role user excludes `routing_type='direct_payment'` rows; finance + admin see all.
- `payment_voucher_test.go`:
  - `MarkPaymentVoucherPaid` rejects when status != approved.
  - Rejects when no POP file uploaded.
  - Cascades requisition status to completed for direct-payment chain.
- Integration test: full HTTP happy path requester → submit → poll PV → finance mark-as-paid → req status=completed.

### Frontend (Vitest unit + Playwright E2E)

- Requisition form:
  - Request type toggle shows/hides payee block.
  - Payee radio: New flow creates payee via API + selects returned ID.
  - Vendor/Employee/Other comboboxes resolve via correct endpoints.
- PV detail mark-as-paid modal:
  - POP required, paid date required.
  - Submit disabled until both present.
  - Success: status pill flips to paid, modal closes.
- Nav: "Source to Pay" label renders in expanded + collapsed sidebar.
- Visibility: procurement-role mock — direct-payment rows absent from req/PO/PV lists.

### Migration test

- Up/down apply cleanly on a dev DB with seed data.
- Existing rows backfilled to the correct routing_type from their workflow; behavior unchanged.

## Performance

- New indexed columns on `routing_type` for hot list filters.
- Trigram GIN index on `payees.name` for fast typeahead.
- POP file uploads go through existing attachment service — no new storage infra.

## Rollout

1. Ship migration in one release; columns are nullable / defaulted so no downtime risk.
2. Ship backend handlers + scope filter changes (feature stays inert until a `direct_payment` workflow exists in any org).
3. Ship frontend changes.
4. Admin configures the first `direct_payment` workflow per org via the workflow editor.
5. Rename sidebar in same release as the frontend changes.

## Open Questions

None at design time. All implementation choices not specified above (e.g., specific PV approval workflow used for direct-payment PVs, exact column types where ambiguous) defer to existing patterns in the codebase and will be settled during plan-writing.
