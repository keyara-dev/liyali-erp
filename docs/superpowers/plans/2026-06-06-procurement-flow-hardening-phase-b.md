# Procurement Flow Hardening — Phase B Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Per-org, per-document procurement automation — auto-create the next document in the chain and drive it to a configurable level (manual / auto_submit / auto_approve, capped), configured from a tenant-admin settings UI.

**Architecture:** Extend `OrganizationSettings` with automation fields (new migration). A single `applyAutomationLevel` helper resolves manual/submit/approve with an amount cap. The PO terminal-approve hook fans out to mode-aware auto-create (goods_first → GRN, payment_first → PV); the GRN terminal-approve hook already auto-creates the PV. The settings UI is built with the frontend-design skill.

**Tech Stack:** Go (Fiber + GORM + PostgreSQL migrations), Next.js (TypeScript). Reference spec: `docs/superpowers/specs/2026-06-05-procurement-flow-hardening-design.md` (Phase B section). Run backend from `d:\dev\next-apps\liyali-gateway\backend`, frontend from `…\frontend`.

**Decisions (locked):** goods_first = PO→GRN→PV, payment_first = PO→PV→GRN. Levels: `manual` | `auto_submit` | `auto_approve`. `auto_approve` only when `amount ≤ AutoApproveMaxAmount`, else falls back to `auto_submit`. Built on the Phase-A guards (PO=APPROVED entry gate, `validateProcurementPVGate`, `resolveDeliveryFromGRNs`).

---

## File Structure

- `backend/database/migrations/022_org_automation_settings.up.sql` / `.down.sql` (new) — add 5 columns.
- `backend/models/organization.go` — add 5 fields to `OrganizationSettings`.
- `backend/handlers/organization.go` — `UpdateOrganizationSettings` request struct + model build.
- `backend/services/organization_service.go` — update column map.
- `backend/services/automation_engine.go` (new) — `AutomationLevel` constants + `applyAutomationLevel` + `autoCreateFromApprovedPO`.
- `backend/services/workflow_execution_service.go` — PO terminal-approve hook; switch goods_first auto-PV flag.
- `frontend/src/types/organization.ts` (or existing settings type) — add fields.
- `frontend/src/app/_actions/organizations.ts` — include fields in update payload.
- `frontend/src/app/(private)/(main)/settings/_components/workspace-settings.tsx` — automation section (frontend-design).

---

## PART 1 — BACKEND

## Task 1: Migration — add automation columns

**Files:** Create `backend/database/migrations/022_org_automation_settings.up.sql` and `…_down.sql`. (Confirm next number with `ls backend/database/migrations` — use the integer after the current highest.)

- [ ] **Step 1: Write the up migration**

```sql
-- 022_org_automation_settings.up.sql
ALTER TABLE organization_settings
  ADD COLUMN IF NOT EXISTS auto_create_grn_from_po BOOLEAN NOT NULL DEFAULT FALSE,
  ADD COLUMN IF NOT EXISTS auto_create_pv_from_grn BOOLEAN NOT NULL DEFAULT FALSE,
  ADD COLUMN IF NOT EXISTS grn_automation_level VARCHAR(20) NOT NULL DEFAULT 'manual',
  ADD COLUMN IF NOT EXISTS pv_automation_level  VARCHAR(20) NOT NULL DEFAULT 'manual',
  ADD COLUMN IF NOT EXISTS auto_approve_max_amount DOUBLE PRECISION NOT NULL DEFAULT 0;
```

- [ ] **Step 2: Write the down migration**

```sql
-- 022_org_automation_settings.down.sql
ALTER TABLE organization_settings
  DROP COLUMN IF EXISTS auto_create_grn_from_po,
  DROP COLUMN IF EXISTS auto_create_pv_from_grn,
  DROP COLUMN IF EXISTS grn_automation_level,
  DROP COLUMN IF EXISTS pv_automation_level,
  DROP COLUMN IF EXISTS auto_approve_max_amount;
```

- [ ] **Step 3: Verify migration parity** — confirm column names match the gorm tags added in Task 2.
- [ ] **Step 4: Commit** — `chore(db): migration 022 — org automation settings columns`

---

## Task 2: Model fields

**Files:** Modify `backend/models/organization.go` (`OrganizationSettings`, after the existing automation flags ~line 64).

- [ ] **Step 1: Add fields**

```go
	// Auto-create the next document in the chain when a PO/GRN is approved.
	AutoCreateGRNFromPO bool `gorm:"column:auto_create_grn_from_po;default:false" json:"autoCreateGRNFromPO"` // goods_first: PO approved → GRN
	AutoCreatePVFromGRN bool `gorm:"column:auto_create_pv_from_grn;default:false" json:"autoCreatePVFromGRN"` // goods_first: GRN completed → PV
	// Automation level applied to an auto-created document: "manual" | "auto_submit" | "auto_approve".
	GRNAutomationLevel string `gorm:"column:grn_automation_level;default:'manual'" json:"grnAutomationLevel"`
	PVAutomationLevel  string `gorm:"column:pv_automation_level;default:'manual'" json:"pvAutomationLevel"`
	// auto_approve only applies at/below this amount; above it falls back to auto_submit. 0 = never auto-approve.
	AutoApproveMaxAmount float64 `gorm:"column:auto_approve_max_amount;default:0" json:"autoApproveMaxAmount"`
```

(Note: `AutoCreatePVFromPO`, `AutoSubmitGRNToWorkflow`, `AutoSubmitPVToWorkflow` already exist and remain — `AutoCreatePVFromPO` drives payment_first PO→PV.)

- [ ] **Step 2:** `go build ./...` — clean.
- [ ] **Step 3: Commit** — `feat(models): OrganizationSettings automation fields`

---

## Task 3: Wire settings through update handler + service

**Files:** `backend/handlers/organization.go` (`UpdateOrganizationSettings` 308-334), `backend/services/organization_service.go` (`UpdateOrganizationSettings` map 318-327). Test: `backend/handlers/*organization*_test.go` (or add one).

- [ ] **Step 1: Write a failing test** — PUT `/organization/settings` with the new fields persists them.

```go
func TestUpdateOrganizationSettings_PersistsAutomationFields(t *testing.T) {
	// Arrange: seed an org_settings row (AutoMigrate &models.OrganizationSettings{}).
	// Act: call UpdateOrganizationSettings with body {autoCreateGRNFromPO:true,
	//      grnAutomationLevel:"auto_submit", autoApproveMaxAmount: 5000}.
	// Assert: reload row → AutoCreateGRNFromPO==true, GRNAutomationLevel=="auto_submit",
	//         AutoApproveMaxAmount==5000.
}
```

- [ ] **Step 2: Run, verify FAIL** (fields ignored).
- [ ] **Step 3: Implement** — add the 5 fields to the handler request struct (organization.go:308-317), the `orgSettings` model build (325-334), and the service `Updates` map (organization_service.go:318-327):

```go
// request struct additions
AutoCreateGRNFromPO  bool    `json:"autoCreateGRNFromPO"`
AutoCreatePVFromGRN  bool    `json:"autoCreatePVFromGRN"`
AutoCreatePVFromPO   bool    `json:"autoCreatePVFromPO"`
GRNAutomationLevel   string  `json:"grnAutomationLevel"`
PVAutomationLevel    string  `json:"pvAutomationLevel"`
AutoApproveMaxAmount float64 `json:"autoApproveMaxAmount"`
// model build additions: AutoCreateGRNFromPO: settings.AutoCreateGRNFromPO, … etc.
// service map additions:
"auto_create_grn_from_po": settings.AutoCreateGRNFromPO,
"auto_create_pv_from_grn": settings.AutoCreatePVFromGRN,
"auto_create_pv_from_po":  settings.AutoCreatePVFromPO,
"grn_automation_level":    settings.GRNAutomationLevel,
"pv_automation_level":     settings.PVAutomationLevel,
"auto_approve_max_amount": settings.AutoApproveMaxAmount,
```

Validate the level strings: reject anything other than `manual`/`auto_submit`/`auto_approve` with 400.

- [ ] **Step 4: Run, verify PASS.**
- [ ] **Step 5:** `go build ./...`
- [ ] **Step 6: Commit** — `feat(org): persist automation settings via update endpoint`

---

## Task 4: `applyAutomationLevel` engine

**Files:** Create `backend/services/automation_engine.go`. Test: `backend/services/automation_engine_test.go`.

The engine decides what to do with a freshly-created DRAFT document. `auto_submit`/`auto_approve` reuse the existing submit/approve machinery; this task implements the **decision** + the submit path, and unit-tests the decision.

- [ ] **Step 1: Write failing test for the decision function**

```go
func TestResolveAutomationAction(t *testing.T) {
	// manual → "draft"
	require.Equal(t, "draft", resolveAutomationAction("manual", 100, 1000))
	// auto_submit → "submit"
	require.Equal(t, "submit", resolveAutomationAction("auto_submit", 100, 1000))
	// auto_approve under cap → "approve"
	require.Equal(t, "approve", resolveAutomationAction("auto_approve", 500, 1000))
	// auto_approve over cap → falls back to "submit"
	require.Equal(t, "submit", resolveAutomationAction("auto_approve", 5000, 1000))
	// auto_approve with zero cap → never approves → "submit"
	require.Equal(t, "submit", resolveAutomationAction("auto_approve", 100, 0))
	// unknown/empty → "draft"
	require.Equal(t, "draft", resolveAutomationAction("", 100, 1000))
}
```

- [ ] **Step 2: Run, verify FAIL.**
- [ ] **Step 3: Implement the decision** in `automation_engine.go`:

```go
package services

const (
	AutomationManual     = "manual"
	AutomationAutoSubmit = "auto_submit"
	AutomationAutoApprove = "auto_approve"
)

// resolveAutomationAction returns "draft" | "submit" | "approve" for a level +
// amount + cap. auto_approve only yields "approve" when amount ≤ cap (cap > 0),
// otherwise it falls back to "submit". Unknown levels are treated as manual.
func resolveAutomationAction(level string, amount, cap float64) string {
	switch level {
	case AutomationAutoSubmit:
		return "submit"
	case AutomationAutoApprove:
		if cap > 0 && amount <= cap+0.01 {
			return "approve"
		}
		return "submit"
	default:
		return "draft"
	}
}
```

- [ ] **Step 4: Run, verify PASS.**
- [ ] **Step 5: Add the side-effecting `applyAutomationLevel`** (no new unit test — exercised by integration in Task 5/6). It loads the doc, and for "submit"/"approve" reuses the existing workflow submit path. Signature:

```go
// applyAutomationLevel submits and/or approves an auto-created DRAFT document
// per the resolved action. entityType is "grn"|"payment_voucher". A nil/zero
// workflow is a no-op (stays DRAFT) with a logged warning — never fatal.
func (s *WorkflowExecutionService) applyAutomationLevel(tx *gorm.DB, entityType, entityID, orgID, systemUserID string, action string) error {
	if action == "draft" {
		return nil
	}
	// 1. Resolve the org's default workflow for entityType (reuse the same lookup
	//    SubmitGRN / SubmitPaymentVoucher use to pick a workflow).
	// 2. Assign the workflow + set status PENDING (mirror grn.go:1220-1265 AutoSubmit).
	// 3. If action == "approve": drive the assignment to terminal approval as the
	//    system user (reuse ApproveWorkflowTaskWithVersion per stage, or the
	//    auto-approval path used by autoApproveAndGeneratePO at
	//    workflow_execution_service.go:408). Bounded by amount cap already.
	return nil // replace with the wiring above
}
```

Reference implementations to mirror: AutoSubmit (grn.go:1215-1265, payment_voucher.go:430-449) for submit; `autoApproveAndGeneratePO` (workflow_execution_service.go:408+) for system auto-approval.

- [ ] **Step 6:** `go build ./...`
- [ ] **Step 7: Commit** — `feat(automation): applyAutomationLevel engine (manual/submit/approve + cap)`

---

## Task 5: PO terminal-approve → auto-create next doc

**Files:** `backend/services/workflow_execution_service.go` — add a `purchase_order` branch beside the GRN branch at 1189-1209. New helper `autoCreateFromApprovedPO` in `automation_engine.go`. Test: integration in `workflow_execution_service_test.go`.

- [ ] **Step 1: Write failing integration test** — approve a goods_first PO with `AutoCreateGRNFromPO=true` → a DRAFT GRN now exists for that PO; approve a payment_first PO with `AutoCreatePVFromPO=true` → a DRAFT PV exists.
- [ ] **Step 2: Run, verify FAIL.**
- [ ] **Step 3: Implement** the PO branch (after the GRN branch, before the `else` at 1210):

```go
if strings.EqualFold(assignment.EntityType, "purchase_order") {
	if err := s.autoCreateFromApprovedPO(tx, assignment.EntityID, assignment.OrganizationID); err != nil {
		fmt.Printf("Warning: autoCreateFromApprovedPO failed: %v\n", err)
	}
}
```

And `autoCreateFromApprovedPO(tx, poID, orgID)` in automation_engine.go:
- Load PO + org settings; resolve flow via `utils.ResolveProcurementFlow`.
- **goods_first** + `AutoCreateGRNFromPO`: build a DRAFT GRN from the PO lines (mirror the GRN create item-copy), then `applyAutomationLevel(tx, "grn", grnID, orgID, systemUser, resolveAutomationAction(settings.GRNAutomationLevel, poTotal, settings.AutoApproveMaxAmount))`.
- **payment_first** + `AutoCreatePVFromPO`: build a DRAFT PV from the PO (reuse the same item math as `AutoCreatePVFromCompletedGRN`), then `applyAutomationLevel(tx, "payment_voucher", pvID, …, settings.PVAutomationLevel …)`.
- Respect `validateProcurementPVGate` semantics (dedup) before creating a PV.

- [ ] **Step 4: Run, verify PASS.**
- [ ] **Step 5:** `go build ./... && go test ./services/`
- [ ] **Step 6: Commit** — `feat(automation): auto-create GRN/PV from approved PO per org settings`

---

## Task 6: goods_first GRN→PV auto-create honors level + correct flag

**Files:** `backend/services/workflow_execution_service.go` (`AutoCreatePVFromCompletedGRN` ~2358). Test: `workflow_execution_service_test.go`.

`AutoCreatePVFromCompletedGRN` currently gates on `settings.AutoCreatePVFromPO`. Switch it to the dedicated `settings.AutoCreatePVFromGRN`, and after creating the DRAFT PV, apply `PVAutomationLevel`.

- [ ] **Step 1: Failing test** — goods_first GRN completes with `AutoCreatePVFromGRN=true` + `PVAutomationLevel="auto_submit"` → the created PV is PENDING (submitted), not DRAFT.
- [ ] **Step 2: Run, verify FAIL.**
- [ ] **Step 3: Implement** — change the flag check (`if !settings.AutoCreatePVFromGRN { return nil }`), and after `tx.Create(&pv)` call `applyAutomationLevel(tx, "payment_voucher", pv.ID, …, resolveAutomationAction(settings.PVAutomationLevel, pvTotal, settings.AutoApproveMaxAmount))`.
- [ ] **Step 4: Run, verify PASS.**
- [ ] **Step 5:** `go build ./... && go test ./services/`
- [ ] **Step 6: Commit** — `feat(automation): goods_first GRN→PV uses AutoCreatePVFromGRN + level`

---

## Task 7: Retire the duplicate automation config

**Files:** `backend/services/document_automation_service.go`, `backend/services/document_generation_service.go`. Test: existing suites.

The `AutomationConfig` struct + `triggerAutomation` paths duplicate the OrganizationSettings flags (and `document_generation_service.go` hard-codes `AutoCreateGRNFromPO/AutoCreatePVFromGRN = true`). Point any live caller at `OrganizationSettings`, or remove the dead `triggerAutomation` branches so there is one automation source of truth.

- [ ] **Step 1:** Grep callers of `triggerAutomation` / `AutomationConfig`; confirm which (if any) run in production paths.
- [ ] **Step 2:** Remove/redirect the duplicate; keep behavior identical for live callers.
- [ ] **Step 3:** `go build ./... && go test ./...` — all green.
- [ ] **Step 4: Commit** — `refactor(automation): single source of truth in OrganizationSettings`

---

## PART 2 — FRONTEND (settings UI via frontend-design)

## Task 8: Frontend settings type + action payload

**Files:** `frontend/src/app/_actions/organizations.ts` (the settings update action) + the settings type. Test: `tsc`.

- [ ] **Step 1:** Add the 6 fields to the settings TypeScript type and include them in the `updateOrganizationSettings` action payload (so they round-trip).
- [ ] **Step 2:** `npx tsc --noEmit` — clean.
- [ ] **Step 3: Commit** — `feat(settings): wire automation fields in org settings action`

---

## Task 9: Build the automation settings section (frontend-design)

**Files:** `frontend/src/app/(private)/(main)/settings/_components/workspace-settings.tsx`.

- [ ] **Step 1: Invoke the `frontend-design` skill** to design + build a "Procurement Automation" section, matching the existing settings styling. Controls:
  - Procurement flow (already present) — keep.
  - Auto-create toggles: `AutoCreateGRNFromPO` (goods_first), `AutoCreatePVFromPO` (payment_first), `AutoCreatePVFromGRN` (goods_first).
  - Per-document automation level selects: GRN, PV → manual / auto_submit / auto_approve.
  - `AutoApproveMaxAmount` numeric input (shown/enabled only when a level is auto_approve), with helper text "auto-approve only at or below this amount".
  - Clear labels explaining the mode→chain implication (goods_first = PO→GRN→PV, payment_first = PO→PV→GRN).
- [ ] **Step 2:** Load current values from settings, bind to the controls, save via the Task 8 action.
- [ ] **Step 3:** `npx tsc --noEmit` + visually verify (use the `run` skill if needed).
- [ ] **Step 4: Commit** — `feat(settings): procurement automation UI (frontend-design)`

---

## Task 10: End-to-end verification

- [ ] `go build ./... && go test ./...` — all green.
- [ ] `npx tsc --noEmit` (frontend) — clean.
- [ ] Manual smoke: set goods_first + AutoCreateGRNFromPO + GRNAutomationLevel=auto_submit + AutoApproveMaxAmount; approve a PO; confirm a submitted GRN appears. Repeat payment_first → PV.
- [ ] Confirm auto_approve respects the cap (above-cap doc only gets submitted, not approved).

---

## Self-review notes
- Migration column names ↔ gorm tags ↔ service map keys must match exactly (Task 1/2/3).
- `resolveAutomationAction` signature is reused verbatim in Tasks 5 & 6.
- The hardest task is Task 4 Step 5 (`applyAutomationLevel` submit/approve wiring) — reuse the cited existing AutoSubmit + autoApproveAndGeneratePO code; do not hand-roll workflow assignment.
