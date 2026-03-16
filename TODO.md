# Liyali Gateway - TODO & Future Enhancements

> Last updated: 2026-03-16
> Status: Post system-wide audit (2026-03-16). New sections added: Database/Migrations, Security Hardening, Frontend Type Safety.
> This document captures every remaining stub, placeholder, and enhancement opportunity.

---

## Recently Completed ✅

### Dashboard Analytics Phases 1-4 (2026-03-08)

**Phase 1: Expose Existing Data**

- ✅ Conducted comprehensive audit of dashboard analytics
- ✅ Created unified reports endpoints for all users
- ✅ Exposed all document types on dashboard (Req, PO, PV, GRN, Budget)
- ✅ Implemented average approval time with real data
- ✅ Added recent activity feed (last 50 actions)
- ✅ Complete document type breakdown

**Phase 2: Budget Utilization**

- ✅ Added budget utilization calculation to backend
- ✅ Formula: `SUM(allocated_amount) / SUM(total_budget) * 100`
- ✅ Excludes rejected and cancelled budgets
- ✅ Handles zero budget gracefully
- ✅ Integrated into dashboard response

**Phase 3: Role-Based Views**

- ✅ Implemented role-based filtering in GetDashboardReports handler
- ✅ Admin/Superadmin: Full organization visibility
- ✅ Manager: Full organization visibility (department filtering ready for future)
- ✅ User: Full organization visibility (personal filtering ready for future)
- ✅ Added userRole to dashboard response
- ✅ Infrastructure ready for granular filtering when needed

**Phase 4: Processing Time**

- ✅ Added processing time calculation (creation → completion)
- ✅ Separate from approval time (workflow only)
- ✅ Integrated into dashboard response
- ✅ Handles all document types

**Results**: Organization-scoped, multi-tenant safe, role-aware, zero TypeScript errors, backend compiles successfully

**Documentation**: `DASHBOARD_STATUS.md`, `PHASE1_IMPLEMENTATION_SUMMARY.md`, `PHASE2_IMPLEMENTATION_SUMMARY.md`, `PHASE3_IMPLEMENTATION_SUMMARY.md`, `PHASE4_IMPLEMENTATION_SUMMARY.md`

### User Profile & Requisition Enhancements (2026-03-08)

- ✅ Added user profile fields (Position, Man Number, NRC Number, Contact)
- ✅ Added source of funds field to requisitions
- ✅ Ensured fresh data in PDF previews and QR code verification
- ✅ Eliminated mock data from all document components

### Document Hooks Refactor (2026-03-08)

- ✅ Created reusable hook system for all document detail pages
- ✅ Refactored 5 document types (Requisitions, POs, PVs, GRNs, Budgets)
- ✅ Eliminated ~310 lines of duplicate code
- ✅ Achieved zero TypeScript errors across all components
- ✅ Implemented permissions-based UI controls
- ✅ Created comprehensive documentation

See `IMPLEMENTATION_SUMMARY.md` for detailed information.

---

## How to Use This Document

- **[ ]** = Not started
- **[~]** = Partially done / workaround in place
- **[x]** = Complete
- **[SKIP]** = Intentionally deferred (with reason)

Items are grouped by area and sorted by priority within each group.

---

## 1. Admin Console - Backend Handlers

### 1.1 API Monitoring (Partial Stubs)

| #   | Item                              | File                                               | Line | Priority | Notes                                                                    |
| --- | --------------------------------- | -------------------------------------------------- | ---- | -------- | ------------------------------------------------------------------------ |
| [ ] | Implement API alert rules CRUD    | `backend/handlers/admin_api_monitoring_handler.go` | 177  | Medium   | `CreateAPIAlertRule` returns 501. Needs alert rules model + storage.     |
| [ ] | Implement API alerts retrieval    | `backend/handlers/admin_api_monitoring_handler.go` | 82   | Medium   | `GetAPIAlerts` returns empty array. Depends on alert rules.              |
| [ ] | Implement endpoint testing        | `backend/handlers/admin_api_monitoring_handler.go` | 122  | Low      | `TestAPIEndpoint` returns 501. Would need HTTP client to ping endpoints. |
| [ ] | Implement endpoint config updates | `backend/handlers/admin_api_monitoring_handler.go` | 127  | Low      | `UpdateAPIEndpointConfig` returns 501.                                   |

### 1.2 Database Management (Intentional 501s)

| #      | Item                           | File                                         | Line | Priority | Notes                                          |
| ------ | ------------------------------ | -------------------------------------------- | ---- | -------- | ---------------------------------------------- |
| [SKIP] | Database backup via web UI     | `backend/handlers/admin_database_handler.go` | 510  | N/A      | Safety decision: use `pg_dump` CLI instead.    |
| [SKIP] | Database restore via web UI    | `backend/handlers/admin_database_handler.go` | 515  | N/A      | Safety decision: use `pg_restore` CLI instead. |
| [SKIP] | Run migrations via web UI      | `backend/handlers/admin_database_handler.go` | 575  | N/A      | Safety decision: use deployment pipeline.      |
| [SKIP] | Rollback migrations via web UI | `backend/handlers/admin_database_handler.go` | 580  | N/A      | Safety decision: use manual SQL scripts.       |
| [SKIP] | Export database via web UI     | `backend/handlers/admin_database_handler.go` | 618  | N/A      | Safety decision: use `pg_dump` CLI instead.    |

### 1.3 System Health (Intentional 501s)

| #      | Item                            | File                                              | Line | Priority | Notes                                                                  |
| ------ | ------------------------------- | ------------------------------------------------- | ---- | -------- | ---------------------------------------------------------------------- |
| [SKIP] | Update system config via web UI | `backend/handlers/admin_system_health_handler.go` | 316  | N/A      | Safety: config should be managed via environment variables/deployment. |
| [SKIP] | Restart services via web UI     | `backend/handlers/admin_system_health_handler.go` | 327  | N/A      | Safety: use deployment tools (Fly.io, Docker, etc.).                   |

### 1.4 Admin Users

| #   | Item                         | File                                             | Line | Priority | Notes                                                                                                 |
| --- | ---------------------------- | ------------------------------------------------ | ---- | -------- | ----------------------------------------------------------------------------------------------------- |
| [~] | Full TOTP 2FA infrastructure | `backend/handlers/admin_console_user_handler.go` | 542  | High     | Currently records preference only. Needs TOTP secret generation, QR code, verification, backup codes. |
| [~] | Admin user sessions (real)   | `backend/handlers/admin_console_user_handler.go` | 602  | Medium   | MVP stub querying audit logs as proxy. Needs real session tracking table.                             |

---

## 2. Admin Console - Frontend

### 2.1 Settings Page

| #   | Item                            | File                                            | Line | Priority | Notes                                                                                     |
| --- | ------------------------------- | ----------------------------------------------- | ---- | -------- | ----------------------------------------------------------------------------------------- |
| [ ] | Bulk update settings            | `admin-console/src/app/_actions/settings.ts`    | 250  | Medium   | Returns "not yet implemented". Needs backend endpoint `POST /admin/settings/bulk-update`. |
| [ ] | Validate configuration          | `admin-console/src/app/_actions/settings.ts`    | 309  | Low      | Returns validation error stub. Could check required settings exist + correct types.       |
| [ ] | Import configuration            | `admin-console/src/app/_actions/settings.ts`    | 410  | Low      | Returns "not yet implemented". Needs backend `POST /admin/settings/import`.               |
| [ ] | Reset to defaults               | `admin-console/src/app/_actions/settings.ts`    | 418  | Low      | Returns "not yet implemented". Needs backend endpoint.                                    |
| [ ] | Restore configuration           | `admin-console/src/app/_actions/settings.ts`    | 426  | Low      | Returns "not yet implemented". Needs snapshot/backup system for settings.                 |
| [ ] | Export settings (real download) | `admin-console/src/app/admin/settings/page.tsx` | 157  | Low      | Shows "coming soon" toast. Needs to call backend and trigger download.                    |

### 2.2 Feature Flags Page

| #   | Item                                 | File                                                 | Line | Priority | Notes                                                                            |
| --- | ------------------------------------ | ---------------------------------------------------- | ---- | -------- | -------------------------------------------------------------------------------- |
| [ ] | Bulk update feature flags            | `admin-console/src/app/_actions/feature-flags.ts`    | 299  | Medium   | Console.warn only. Needs backend `POST /admin/feature-flags/bulk-update`.        |
| [ ] | Feature flag templates               | `admin-console/src/app/_actions/feature-flags.ts`    | 372  | Low      | Returns empty array. Needs `feature_flag_templates` table + CRUD handlers.       |
| [ ] | Feature flag evaluation log          | `admin-console/src/app/_actions/feature-flags.ts`    | 335  | Low      | Returns empty array. `getFeatureFlagEvaluations` needs evaluation tracking.      |
| [ ] | Feature flag audit history           | `admin-console/src/app/_actions/feature-flags.ts`    | 384  | Low      | Returns empty array. Could query `admin_audit_logs` filtered by flag actions.    |
| [ ] | Import feature flags                 | `admin-console/src/app/_actions/feature-flags.ts`    | 425  | Low      | Returns "not yet implemented". Needs backend `POST /admin/feature-flags/import`. |
| [ ] | Export feature flags (real download) | `admin-console/src/app/admin/feature-flags/page.tsx` | 206  | Low      | Shows "coming soon" toast.                                                       |

---

## 3. Backend Services & Infrastructure

### 3.1 Core Services

| #   | Item                              | File                                       | Line    | Priority | Notes                                                                                                                                         |
| --- | --------------------------------- | ------------------------------------------ | ------- | -------- | --------------------------------------------------------------------------------------------------------------------------------------------- |
| [ ] | Audit service implementation      | `backend/services/audit_service.go`        | 11-37   | High     | Multiple TODOs. `LogAction` and `LogEvent` are empty stubs. Needed for proper audit trail.                                                    |
| [ ] | Subscription upgrade with payment | `backend/services/subscription_service.go` | 246     | High     | TODO: "Implement actual upgrade logic with payment processing." Currently just changes tier.                                                  |
| [ ] | Department module assignment      | `backend/services/department_service.go`   | 302-330 | Medium   | `GetDepartmentModules`, `AssignModuleToDepartment`, `RemoveModuleFromDepartment` are placeholders. Needs `department_modules` junction table. |
| [ ] | Auth service `GetUserProfile`     | `backend/handlers/auth_handler.go`         | 470     | Medium   | TODO comment. Currently queries DB directly. Should use service layer.                                                                        |

### 3.2 Infrastructure

| #   | Item                                  | File                                               | Line | Priority | Notes                                                                             |
| --- | ------------------------------------- | -------------------------------------------------- | ---- | -------- | --------------------------------------------------------------------------------- |
| [ ] | Admin middleware org membership check | `backend/middleware/admin.go`                      | 117  | Medium   | TODO: "Implement proper organization membership check." Currently placeholder.    |
| [~] | Account lockout repository            | `backend/repository/account_lockout_repository.go` | -    | Medium   | "Temporary implementations that return not implemented errors." Waiting for sqlc. |
| [~] | Password reset repository             | `backend/repository/password_reset_repository.go`  | -    | Medium   | Same as above - temporary until sqlc generation works.                            |
| [ ] | Performance logger memory usage       | `backend/logging/middleware/performance_logger.go` | 213  | Low      | `getCurrentMemoryUsage()` returns hardcoded 0. Should use `runtime.ReadMemStats`. |
| [ ] | Enhanced auth model placeholder       | `backend/models/enhanced_auth.go`                  | 601  | Low      | "returning true as a placeholder" - needs proper implementation.                  |
| [ ] | Notification service initialization   | `backend/main.go`                                  | 116  | Low      | "placeholder for now" - may need proper lifecycle management.                     |

### 3.3 Tests

| #   | Item                         | File                                                | Priority | Notes                                    |
| --- | ---------------------------- | --------------------------------------------------- | -------- | ---------------------------------------- |
| [ ] | Workflow state machine tests | `backend/tests/unit/workflow_state_machine_test.go` | Medium   | Needs `WorkflowState` constants defined. |
| [ ] | Notification service tests   | `backend/tests/unit/notification_service_test.go`   | Medium   | Needs `NotificationEvent` model.         |
| [ ] | Budget validation tests      | `backend/tests/unit/budget_validation_test.go`      | Low      | Needs `BudgetConstraint` model.          |
| [ ] | Document linking tests       | `backend/tests/unit/document_linking_test.go`       | Low      | Needs `DocumentLink` model.              |

---

## 4. Main Frontend (Tenant App)

### 4.1 Dashboard & Analytics

**Audit Status**: ✅ ALL PHASES COMPLETE (2026-03-08)

**Summary**: All 4 phases complete - dashboard shows all document types with real metrics, budget utilization, processing time, and role-based filtering infrastructure.

**See**: `DASHBOARD_STATUS.md` for detailed status | `DASHBOARD_ANALYTICS_AUDIT.md` for full audit

#### ✅ Phase 1: Expose Existing Data (COMPLETE)

| #   | Item                        | Status      | Notes                                    |
| --- | --------------------------- | ----------- | ---------------------------------------- |
| [x] | Unified reports endpoints   | ✅ Complete | Created `/api/v1/reports/*` endpoints    |
| [x] | All document types visible  | ✅ Complete | Dashboard shows Req, PO, PV, GRN, Budget |
| [x] | Average approval time       | ✅ Complete | Real data from AvgApprovalDays           |
| [x] | Recent activity feed        | ✅ Complete | Last 50 approval actions                 |
| [x] | Document type breakdown     | ✅ Complete | Complete breakdown with counts           |
| [x] | Status breakdown (all docs) | ✅ Complete | All document statuses tracked            |
| [x] | Approval/rejection rates    | ✅ Complete | Calculated across all documents          |
| [x] | Backend compilation         | ✅ Complete | No errors                                |
| [x] | Frontend TypeScript         | ✅ Complete | No errors                                |

**New Endpoints Created:**

- ✅ `GET /api/v1/reports/dashboard` - Comprehensive dashboard (all users)
- ✅ `GET /api/v1/reports/system-stats` - System statistics (all users)
- ✅ `GET /api/v1/reports/approval-metrics` - Approval metrics (all users)
- ✅ `GET /api/v1/reports/user-activity` - User activity (admin/manager)
- ✅ `GET /api/v1/reports/analytics` - Advanced analytics (admin/manager)

**Files Modified:**

- `backend/routes/routes.go` - Added unified reports routes
- `backend/handlers/reports.go` - Added GetDashboardReports handler
- `frontend/src/app/_actions/dashboard.ts` - Updated to use new endpoint

#### ✅ Phase 2: Budget Utilization (COMPLETE)

| #   | Item                          | Priority | Status      | Notes                                  |
| --- | ----------------------------- | -------- | ----------- | -------------------------------------- |
| [x] | Backend budget util query     | Medium   | ✅ Complete | Calculate SUM(allocated)/SUM(total)    |
| [x] | Update SystemStatistics model | Medium   | ✅ Complete | Added budgetUtilization field          |
| [x] | Frontend dashboard display    | Medium   | ✅ Complete | Now shows real budget utilization data |

**Completed**: 2026-03-08 | **Effort**: 1 hour | **See**: `PHASE2_IMPLEMENTATION_SUMMARY.md`

#### ✅ Phase 3: Role-Based Views (COMPLETE)

| #   | Item                   | Priority | Status      | Effort | Notes                                    |
| --- | ---------------------- | -------- | ----------- | ------ | ---------------------------------------- |
| [x] | Backend role filtering | Medium   | ✅ Complete | 1h     | Role-based switch in GetDashboardReports |
| [x] | Admin view             | Medium   | ✅ Complete | -      | Full organization visibility             |
| [x] | Manager view           | Medium   | ✅ Complete | -      | Full organization visibility (ready)     |
| [x] | User view              | Medium   | ✅ Complete | -      | Full organization visibility (ready)     |

**Completed**: 2026-03-08 | **Effort**: 1 hour | **See**: `PHASE3_IMPLEMENTATION_SUMMARY.md`

**Note**: All roles currently see full organization data (system overview). Infrastructure is in place for granular department/personal filtering when business requirements are defined.

#### ✅ Phase 4: Processing Time (COMPLETE)

| #   | Item                          | Priority | Status      | Notes                             |
| --- | ----------------------------- | -------- | ----------- | --------------------------------- |
| [x] | Backend processing time calc  | Low      | ✅ Complete | Track creation → completion time  |
| [x] | Update SystemStatistics model | Low      | ✅ Complete | Added averageProcessingTime field |
| [x] | Frontend display              | Low      | ✅ Complete | Shows processing time metric      |

**Completed**: 2026-03-08 | **Effort**: 1 hour | **See**: `PHASE4_IMPLEMENTATION_SUMMARY.md`

**Note**: Processing time (creation → completion) is now separate from approval time (workflow only)

---

**Legacy Endpoints** (Still available for backward compatibility):

- ✅ `GET /api/v1/analytics/dashboard` - Requisition metrics only (deprecated)
- ✅ `GET /api/v1/analytics/requisitions/metrics` - Detailed requisition analytics
- ✅ `GET /api/v1/analytics/approvals/metrics` - Approval metrics

### 4.2 Bulk Operations

| #   | Item                                            | File                                           | Line    | Priority | Notes                                                                                                                                |
| --- | ----------------------------------------------- | ---------------------------------------------- | ------- | -------- | ------------------------------------------------------------------------------------------------------------------------------------ |
| [ ] | Replace mock bulk operations with real DB calls | `frontend/src/app/_actions/bulk-operations.ts` | 31-288  | High     | Multiple TODOs: "In production, replace with actual database operations." All bulk approve/reject/reassign use simulated operations. |
| [ ] | Audit trail logging for bulk ops                | `frontend/src/app/_actions/bulk-operations.ts` | 59, 120 | Medium   | TODO: "Log to audit trail."                                                                                                          |

### 4.3 Offline / PWA Features

| #   | Item                               | File                                                | Line | Priority | Notes                                           |
| --- | ---------------------------------- | --------------------------------------------------- | ---- | -------- | ----------------------------------------------- |
| [ ] | Purchase order offline operations  | `frontend/src/hooks/use-offline-queue-processor.ts` | 290  | Low      | Throws "not implemented yet."                   |
| [ ] | Payment voucher offline operations | `frontend/src/hooks/use-offline-queue-processor.ts` | 320  | Low      | Throws "not implemented yet."                   |
| [ ] | Budget offline operations          | `frontend/src/hooks/use-offline-queue-processor.ts` | 368  | Low      | Throws "not implemented yet."                   |
| [ ] | Vendor offline operations          | `frontend/src/hooks/use-offline-queue-processor.ts` | 378  | Low      | Throws "not implemented yet."                   |
| [ ] | Offline queue backend endpoint     | `frontend/src/lib/storage/hooks.ts`                 | 27   | Low      | TODO: "Replace with real backend API endpoint." |

### 4.4 Document & File Generation

| #   | Item                           | File                                                      | Line | Priority | Notes                                                                     |
| --- | ------------------------------ | --------------------------------------------------------- | ---- | -------- | ------------------------------------------------------------------------- |
| [ ] | Purchase order PDF generation  | `frontend/src/lib/pdf-generators/purchase-order-pdf.tsx`  | 19   | Medium   | Placeholder. Needs `@react-pdf/renderer`.                                 |
| [ ] | Payment voucher PDF generation | `frontend/src/lib/pdf-generators/payment-voucher-pdf.tsx` | 19   | Medium   | Placeholder. Needs `@react-pdf/renderer`.                                 |
| [ ] | PDF batch export (JSZip)       | `frontend/src/lib/pdf/pdf-batch-export.ts`                | 40+  | Low      | "JSZip library not available" - needs `jszip` package.                    |
| [ ] | Image upload integration       | `frontend/src/components/ui/image-upload.tsx`             | 31   | Medium   | Placeholder toast. Needs `@imagekit/next` or alternative upload provider. |

### 4.5 Payments & Subscriptions

| #   | Item                          | File                                                     | Line | Priority | Notes                                                                   |
| --- | ----------------------------- | -------------------------------------------------------- | ---- | -------- | ----------------------------------------------------------------------- |
| [ ] | Payment method integration    | `frontend/src/components/subscription/upgrade-modal.tsx` | 128  | High     | TODO: "Add payment method integration." Stripe/payment provider needed. |
| [ ] | Contact sales flow            | `frontend/src/components/subscription/upgrade-modal.tsx` | 114  | Medium   | TODO: "Implement contact sales flow."                                   |
| [ ] | Organization tier upgrade API | `frontend/src/hooks/use-organization-tier.ts`            | 27   | Medium   | TODO: "Implement actual API call to upgrade organization."              |

### 4.6 User Management

| #   | Item                            | File                                        | Line | Priority | Notes                                                                             |
| --- | ------------------------------- | ------------------------------------------- | ---- | -------- | --------------------------------------------------------------------------------- |
| [ ] | Update org member role endpoint | `frontend/src/app/_actions/user-actions.ts` | 230  | Medium   | TODO: "Backend needs to implement PUT /api/v1/organization/members/:id endpoint." |
| [ ] | User profile update via backend | `frontend/src/app/_actions/settings.ts`     | 67   | Medium   | TODO: "Call backend API to update user profile."                                  |

### 4.7 Document Detail Pages ✅

| #   | Item                           | Status   | Notes                                                                       |
| --- | ------------------------------ | -------- | --------------------------------------------------------------------------- |
| [x] | Reusable document hooks system | Complete | All 5 document types refactored. See `DOCUMENT_HOOKS_REFACTOR_COMPLETE.md`. |
| [x] | Eliminate duplicate PDF logic  | Complete | ~310 lines of duplicate code removed.                                       |
| [x] | Permissions-based UI controls  | Complete | All components use hook-based permissions.                                  |
| [x] | Type-safe document operations  | Complete | Zero TypeScript errors across all components.                               |

### 4.8 Miscellaneous

| #   | Item                          | File                                                                                   | Line | Priority | Notes                                                                 |
| --- | ----------------------------- | -------------------------------------------------------------------------------------- | ---- | -------- | --------------------------------------------------------------------- |
| [ ] | Notifications pagination      | `frontend/src/app/(private)/(main)/notifications/_components/notifications-client.tsx` | 75   | Low      | TODO: "Backend should return PaginatedResponse with pagination info." |
| [ ] | Workflow document attachments | `frontend/src/components/workflows/workflow-stage-form.tsx`                            | 70   | Low      | "Attach relevant documents (not yet implemented)." Needs file upload. |
| [ ] | Permissions hook fallback     | `frontend/src/hooks/use-permissions.ts`                                                | 152  | Low      | Uses "hardcoded permissions for built-in roles" as fallback.          |
| [ ] | Cache manager concept         | `frontend/src/lib/cache-manager.ts`                                                    | 284  | Low      | "This is a placeholder for the concept."                              |

---

## 5. Database & Migrations

| #   | Item                                     | File                                                                        | Priority | Notes                                                                                                                              |
| --- | ---------------------------------------- | --------------------------------------------------------------------------- | -------- | ---------------------------------------------------------------------------------------------------------------------------------- |
| [ ] | Fix duplicate migration 027              | `backend/database/migrations/`                                              | High     | Two files share number 027: `027_fix_pro_tier_clears_trial.up.sql` and `027_organization_branches.up.sql`. Rename second to 032+. |
| [ ] | Commit migration 031                     | `backend/database/migrations/031_seed_global_system_roles.up.sql`          | High     | File exists locally but is untracked (`??` in git status). Commit or remove.                                                      |
| [ ] | Write DOWN migration for 031             | `backend/database/migrations/031_seed_global_system_roles.down.sql`        | Medium   | No rollback script exists for the global system roles seed.                                                                       |
| [ ] | Add missing DOWN migrations              | `backend/database/migrations/`                                              | Low      | Only 13 of 31 UP migrations have a corresponding DOWN file. Add rollbacks for: user profile (021), activity logs (023), MFA/LDAP (024), impersonation logs (025), org branches (030), etc. |
| [ ] | Verify IsSuperAdmin column nullable      | `backend/models/models.go` + migration 017                                  | Medium   | Model uses `*bool` (nullable pointer). Confirm DB column matches — if NOT NULL default false, change to `bool` in model.           |

---

## 6. Security Hardening

| #   | Item                                         | File                                               | Priority | Notes                                                                                                               |
| --- | -------------------------------------------- | -------------------------------------------------- | -------- | ------------------------------------------------------------------------------------------------------------------- |
| [ ] | Remove JWT_SECRET insecure fallback          | `backend/main.go`                                  | High     | Falls back to `"temp-production-secret-change-me"` when `JWT_SECRET` env var is missing. Should hard-fail instead. |
| [ ] | Add startup env var validation               | `backend/bootstrap/` or `backend/main.go`          | High     | No validation that required env vars (DB_PASSWORD, JWT_SECRET, etc.) are set before the server starts.             |
| [ ] | Fix rate limiter for proxied/load-balanced   | `backend/middleware/middleware.go`                  | Medium   | Rate limiting keys off raw IP only. Behind a load balancer, all requests share one IP. Add `X-Forwarded-For` / `X-Real-IP` header parsing. |
| [ ] | Remove debug log.Printf from approval flow   | `backend/handlers/approval_handler.go` (line ~649) | Low      | Several `log.Printf` debug statements left in production code path.                                                 |
| [ ] | Verify document scope filters are exhaustive | `backend/handlers/`                                | Medium   | `GetDocumentScope` + `ApplyToQuery` applied to Req, PO, PV, GRN, Budget — re-confirm no handler bypasses the scope filter on list endpoints. |

---

## 7. Frontend Type Safety

| #   | Item                                         | File                                               | Priority | Notes                                                                                                                        |
| --- | -------------------------------------------- | -------------------------------------------------- | -------- | ---------------------------------------------------------------------------------------------------------------------------- |
| [ ] | Add `role` to NextAuth session user type     | `frontend/src/types/` or `next-auth.d.ts`          | Medium   | Session type from `@auth/core` does not include `role`. Every component works around this with `session.user as any`.       |
| [ ] | Eliminate `session.user as any` casts        | Multiple components (24+ occurrences)              | Medium   | Once the session type is extended, replace all unsafe casts with typed access.                                               |
| [ ] | Reduce broad `any` typings in hooks          | `frontend/src/hooks/`                              | Low      | ~734 implicit `any` occurrences across hooks and components. Tighten incrementally, starting with the most-used hooks.      |

---

## 8. Backend Data Issues

| #   | Item                               | File                                              | Line  | Priority | Notes                                                                              |
| --- | ---------------------------------- | ------------------------------------------------- | ----- | -------- | ---------------------------------------------------------------------------------- |
| [ ] | Vendor placeholder ID              | `backend/services/document_automation_service.go` | 81-90 | Medium   | Uses hardcoded "vendor-placeholder-001" when vendor not found.                     |
| [ ] | Organization contact/billing stubs | `backend/handlers/admin_organization_handler.go`  | 232   | Low      | "Add contact_info and billing_info stubs (frontend expects these nested objects)." |
| [ ] | Organization service logging       | `backend/services/organization_service.go`        | 66    | Low      | TODO: "Add proper logging here."                                                   |

---

## 6. Suggested Implementation Roadmap

### Sprint 1: High-Priority Backend (estimated: ~2-3 days of work)

1. **Audit service** - Implement `LogAction` and `LogEvent` so all actions are tracked
2. **Bulk operations** - Replace simulated bulk approve/reject with real DB operations
3. **2FA infrastructure** - Add TOTP secret generation, QR enrollment, verification endpoint
4. **Payment integration** - PayBoss checkout for subscription upgrades

### Sprint 2: Admin Console Polish (estimated: ~2 days)

1. **Feature flag bulk operations** - Backend endpoint + wire frontend
2. **Settings bulk/import/export** - Backend endpoints + wire frontend
3. **Feature flag audit trail** - Query `admin_audit_logs` for flag-related actions
4. **API alert rules** - Create model + CRUD + threshold evaluation

### Sprint 3: Tenant App Completeness (estimated: ~3 days)

1. **Dashboard analytics** - Add missing metrics to backend analytics endpoints
2. **PDF generation** - Install `@react-pdf/renderer`, implement PO and PV generators
3. **Image upload** - Integrate ImageKit or S3-compatible upload
4. **Offline operations** - Implement remaining offline queue processors

### Sprint 4: Infrastructure & Testing (estimated: ~2 days)

1. **Repository layer** - Complete sqlc migration for account lockout + password reset
2. **Unit tests** - Uncomment and fix workflow, notification, budget, document tests
3. **Department modules** - Create junction table + implement assignment handlers
4. **Session tracking** - Real session table for admin users

### Deferred / Won't Do

- Database backup/restore/migration via web UI (security decision)
- System config update/service restart via web UI (security decision)
- Cache manager implementation (premature optimization)

---

## 7. File Reference Index

Quick lookup of files mentioned above:

```
backend/handlers/
  admin_api_monitoring_handler.go    -- API monitoring stubs
  admin_console_user_handler.go      -- 2FA, sessions, impersonation
  admin_database_handler.go          -- DB management (501s intentional)
  admin_organization_handler.go      -- Org contact/billing stubs
  admin_system_health_handler.go     -- Config/restart (501s intentional)
  auth_handler.go                    -- GetUserProfile TODO

backend/services/
  audit_service.go                   -- Empty audit logging
  department_service.go              -- Module assignment stubs
  document_automation_service.go     -- Vendor placeholder
  subscription_service.go            -- Payment processing TODO

backend/middleware/
  admin.go                           -- Org membership check TODO

admin-console/src/app/_actions/
  settings.ts                        -- Bulk/import/export/validate/restore
  feature-flags.ts                   -- Bulk/templates/audit/import

admin-console/src/app/admin/
  settings/page.tsx                  -- "Coming soon" toasts
  feature-flags/page.tsx             -- "Coming soon" toasts

frontend/src/app/_actions/
  dashboard.ts                       -- Missing analytics metrics
  bulk-operations.ts                 -- Simulated operations

frontend/src/hooks/
  use-offline-queue-processor.ts     -- Unimplemented offline ops

frontend/src/lib/pdf-generators/
  purchase-order-pdf.tsx             -- Placeholder
  payment-voucher-pdf.tsx            -- Placeholder

frontend/src/components/
  subscription/upgrade-modal.tsx     -- Payment integration TODO
  ui/image-upload.tsx                -- Upload placeholder
```
