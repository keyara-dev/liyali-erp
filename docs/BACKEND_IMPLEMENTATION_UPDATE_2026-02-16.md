# Backend Implementation Update (2026-02-16)

This document summarizes backend changes delivered for:

1. Organization/workspace details management (CRUD enhancements)
2. Workflow selection before document submission
3. Removal of automatic post-approval generation and introduction of explicit generation endpoints

## Scope

Backend only. No frontend implementation changes are included in this update.

## 1) Organization/Workspace Management Changes

### What changed

- Added support for `logoUrl` in organization create/update APIs.
- Added read endpoint to fetch organization by ID.
- Kept soft-delete behavior non-destructive (data remains in DB; organization becomes inactive).

### API behavior

- `POST /api/v1/organizations`
  - Accepts: `name`, `description`, `logoUrl`
- `PUT /api/v1/organizations/:id`
  - Accepts: `name`, `description`, `logoUrl`
- `GET /api/v1/organizations/:id`
  - Returns organization details if caller can manage the organization.
- `DELETE /api/v1/organizations/:id`
  - Soft delete via `active=false` and member deactivation (no hard delete).

### Files

- `backend/handlers/organization.go`
- `backend/services/organization_service.go`
- `backend/routes/routes.go`
- `backend/services/auth_service.go` (constructor callsite updated)

## 2) Workflow Selection Before Submit

### What changed

- Submit flows now require a `workflowId` in request body for document-specific submit endpoints.
- Backend validates selected workflow:
  - valid UUID format
  - exists within same organization
  - active
  - matching entity type (e.g., requisition workflow for requisition submit)

### Submit request contract

For these endpoints:

- `POST /api/v1/requisitions/:id/submit`
- `POST /api/v1/budgets/:id/submit`
- `POST /api/v1/purchase-orders/:id/submit`
- `POST /api/v1/payment-vouchers/:id/submit`
- `POST /api/v1/grns/:id/submit`

Request body now includes:

```json
{
  "workflowId": "<uuid>"
}
```

### Files

- `backend/services/workflow_execution_service.go`
- `backend/handlers/requisition.go`
- `backend/handlers/budget.go`
- `backend/handlers/purchase_order.go`
- `backend/handlers/payment_voucher.go`
- `backend/handlers/grn.go`
- `backend/types/documents.go`

## 3) Manual Generation Only (No Auto Generation)

### What changed

- Disabled automatic post-approval generation by default.
- Added explicit generation endpoint to generate downstream documents from approved source docs.

### New endpoint

- `POST /api/v1/documents/generate`

Request body:

```json
{
  "id": "<sourceDocId>",
  "docType": "REQUISITION|PURCHASE_ORDER|GRN",
  "targetDocType": "PURCHASE_ORDER|GRN|PAYMENT_VOUCHER"
}
```

Notes:

- `targetDocType` is optional; if provided, it must match valid mapping:
  - `REQUISITION -> PURCHASE_ORDER`
  - `PURCHASE_ORDER -> GRN`
  - `GRN -> PAYMENT_VOUCHER`
- Source document must be `approved`.
- Generation is organization-scoped.
- Duplicate generation for the same source is blocked.

### Files

- `backend/services/document_automation_service.go` (default automation config disabled)
- `backend/services/document_generation_service.go` (new)
- `backend/handlers/document_generation_handler.go` (new)
- `backend/handlers/handler_registry.go`
- `backend/main.go`
- `backend/routes/routes.go`

## Tests

### Added

- `backend/tests/unit/workflow_selection_generation_service_test.go`

### Executed

- `go test ./tests/unit -count=1` (pass)
- `go test ./handlers ./services ./routes -count=1` (pass)

Note: `go test ./...` currently fails in this repository for pre-existing reasons unrelated to this implementation (utility folders with multiple `main` declarations and bootstrap DB credential requirements).

## Branch

- `codex/backend-org-workflow-generation`
