# Project Roadmap

**Last Updated**: 2026-05-08
**Status**: 🟡 **MVP COMPLETE — pre-production hardening required**
**Source of truth for outstanding work**: [`/TODO.md`](../TODO.md)

---

## Stack (verified 2026-05-08)

- **Frontend**: Next.js 16 + React 19 + TypeScript, Tailwind v4, ShadCN UI, TanStack Query, Radix UI
- **Backend**: Go 1.22, Fiber, GORM, PostgreSQL
- **Deployment**: Railway (Docker, port 8080, `DATABASE_URL`)
- **Auth**: NextAuth (frontend) + JWT (backend), RBAC

## Canonical Roles

`SystemRole = "admin" | "approver" | "finance" | "requester"` — see [`frontend/src/types/core.ts`](../frontend/src/types/core.ts). Default new-user role: `requester`.

---

## Shipped

### Core

- JWT auth + RBAC, multi-tenant org isolation
- 5 workflow types (Requisition, PO, PV, GRN, Budget)
- 60+ REST endpoints, document numbering, QR verification
- Bulk approve/reject/reassign w/ feature-flag gating
- Migrations 000–009 (all w/ `.down.sql`)

### Workflow Engine

- Claim expiry (30 min): two-layer auto-reset — UPDATE on `GetApprovalTasks` + `StartClaimExpiryWorker` goroutine (60s tick)
- Rejection w/ stage return: `rejectionType` (`reject` | `return_for_revision`) + `returnToStage`
- Session revocation, rate limiting (raw IP — proxy-aware pending)

### Role-Based Access

- Document scope filter ([`backend/utils/document_scope.go`](../backend/utils/document_scope.go)) applied to Req/PO/PV/GRN/Budget
- Procurement filter via `workflow_assignments JOIN workflows`
- Frontend action gating: creator-or-role checks on PO/PV/GRN tables
- Role-based dashboard variants (admin/approver/procurement/requester) — [`frontend/src/lib/dashboard-role.ts`](../frontend/src/lib/dashboard-role.ts)
- Nav filtering via `requiredRoles`/`requiredPermissions` per item

### Dashboard Analytics (Phases 1–4, 2026-03-08)

- Unified `/api/v1/reports/*` endpoints
- All doc types visible: Req, PO, PV, GRN, Budget
- Real avg approval time, recent activity feed (50 actions), status breakdown, approval/rejection rates
- Budget utilization (`SUM(allocated)/SUM(total)`)
- Role-based filtering infrastructure
- Processing time (creation → completion) separate from approval time

### Frontend Polish

- Document hooks refactor — 5 doc types, ~310 LOC removed, zero TS errors
- Sidebar collapsed (icon) mode, double-padding fix
- WorkflowActionButtons modal extraction across all variants
- ApprovalActionModal mobile fix (canvas signature pad)
- Static OG/Twitter PNG (replaced ImageResponse generators)

### Tests

- Unit, integration, component coverage (gaps remain — see [`/TODO.md`](../TODO.md) §3.3)

---

## Remaining Work

> Full breakdown in [`/TODO.md`](../TODO.md). Sprints below mirror §9 of that doc.

### Sprint 0 — MVP Blockers (~1 day) — ship-stoppers

| ID  | Item                                                                  | File                                                       |
| --- | --------------------------------------------------------------------- | ---------------------------------------------------------- |
| B1  | JWT_SECRET hard-fail (no `temp-production-secret-change-me` fallback) | `backend/main.go:56-64`                                    |
| B2  | Wire real email provider (SMTP/Resend/SendGrid)                       | `backend/services/email_service.go`                        |
| B3  | Implement `audit_service.LogAction` / `LogEvent`                      | `backend/services/audit_service.go:11-37`                  |
| H1  | Session check: reject on DB timeout (currently fail-open)             | `backend/middleware/middleware.go:278`                     |
| H2  | Startup env var `log.Fatal` guard                                     | `backend/main.go`                                          |
| H3  | Remove `[DEBUG]` log.Printf in workflow hot paths                     | `backend/services/workflow_execution_service.go:651,1013`  |
| H4  | Route async notification errors to structured logger                  | `workflow_execution_service.go`                            |
| H5  | Implement `PUT /api/v1/organization/members/:id`                      | backend                                                    |
| H6  | Wire profile save to backend                                          | `frontend/src/app/_actions/settings.ts:67`                 |

### Sprint 1 — High-Priority Backend (~2 days)

- M1 Admin org membership check ([`backend/middleware/admin.go:117`](../backend/middleware/admin.go#L117))
- M2 Rate limiter `X-Forwarded-For` / `X-Real-IP` parsing
- M5 Subscription upgrade — gate behind payment or remove UI
- M6 TOTP 2FA: secret gen, QR enroll, verify, backup codes
- M7 Replace `vendor-placeholder-001` w/ real lookup

### Sprint 2 — Frontend Polish (~2 days)

- M3 Extend NextAuth session type w/ `role`; remove 24+ `session.user as any` casts
- M4 Install `@react-pdf/renderer`; PO + PV PDF generators
- Admin console settings: bulk update / import / export / validate / restore
- Admin console feature flags: bulk / templates / evaluation log / audit history / import / export
- L8 Google OAuth — implement or remove "Coming soon" button

### Sprint 3 — Completeness (~3 days)

- L1 PWA offline queue: PO, PV, Budget, Vendor processors
- L2 Image upload provider (ImageKit / S3-compatible)
- L3 PDF batch export — install `jszip`
- L4 Department module assignment + junction table
- L12 Admin user sessions — real session tracking table
- L9 Notifications paginated response
- L10 Workflow document attachments

### Sprint 4 — Infra & Tests (~2 days)

- sqlc migration: account_lockout + password_reset repositories
- Unit tests: workflow_state_machine, notification_service, budget_validation, document_linking
- Performance logger: real memory via `runtime.ReadMemStats`
- Verify `IsSuperAdmin` model (`*bool`) vs DB column nullability

### Deferred / Won't Do

- DB backup/restore/migrations via web UI — use `pg_dump` / `pg_restore` / pipeline
- System config update / service restart via web UI — use deployment tools
- Cache manager implementation — premature optimization
- Admin console feature flag templates table (low value pre-scale)

---

## Architecture

```text
┌──────────────┐    ┌─────────────┐    ┌──────────────┐
│  Frontend    │    │   Backend   │    │   Database   │
│  Next.js 16  │◄──►│  Go Fiber   │◄──►│  PostgreSQL  │
│  React 19    │    │  GORM       │    │  Multi-tenant│
└──────────────┘    └─────────────┘    └──────────────┘
       │                   │                    │
   NextAuth +          JWT + RBAC +         Migrations
   role-aware UI       doc_scope filter     000–009
```

## Estimated Time to Production-Hardened

~10 working days of focused work across Sprints 0–4. Sprint 0 must ship before any production traffic.

## Next Review

After Sprint 0 completion.
