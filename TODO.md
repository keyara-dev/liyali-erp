# Liyali Gateway - TODO & Future Enhancements

> Last updated: 2026-03-26
> Status: Post full MVP readiness audit (2026-03-26). Added Section 0 (MVP blockers). Corrected stale items: bulk operations confirmed implemented, migration numbering confirmed 000-009 only. Roadmap updated.
> This document captures every remaining stub, placeholder, and enhancement opportunity.

---

## 0. MVP Readiness — Audit 2026-03-26

> Full system audit completed 2026-03-26. Items below must be resolved before production deploy.
> Confirmed working (no action needed): CORS, graceful shutdown, role-based doc visibility, workflow claim expiry, rejection with stage return, session revocation, rate limiting, bulk operations (real API calls), migrations 000-009 (all have .down.sql). Frontend .env is gitignored — not committed.

### 0.1 Blockers — Must fix before any production deploy

| ID  | Item                                    | File                                               | Notes                                                                                                       |
| --- | --------------------------------------- | -------------------------------------------------- | ----------------------------------------------------------------------------------------------------------- |
| B1  | JWT_SECRET falls back to known string   | `backend/main.go:56-64`                            | Sets `"temp-production-secret-change-me"` when env var missing. Anyone who reads the code can forge tokens. Change to `log.Fatal`. |
| B2  | Email service is a stub — never sends   | `backend/services/email_service.go`                | `SendInvitationEmail` logs and returns nil. Password resets and invitations never delivered. Wire SMTP/Resend/SendGrid. |
| B3  | Audit service LogAction/LogEvent empty  | `backend/services/audit_service.go:11-37`          | Both methods are no-ops. No audit trail is written anywhere. Implement or at minimum log to structured output. |

### 0.2 High Priority — Fix before or immediately after launch

| ID  | Item                                      | File                                                                 | Notes                                                                                                       |
| --- | ----------------------------------------- | -------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------- |
| H1  | Fail-open session check                   | `backend/middleware/middleware.go:278-280`                           | If DB is unreachable during session validation, ALL tokens pass (including revoked). Reject on DB timeout.   |
| H2  | No startup env var validation             | `backend/main.go`                                                    | Server starts with empty DB_PASSWORD, empty FRONTEND_URL etc. — silently uses bad defaults. Add `log.Fatal` guard before `InitDatabase()`. |
| H3  | Debug log.Printf in workflow hot paths    | `backend/services/workflow_execution_service.go:651,1013`            | `[DEBUG]` log statements in approval and rejection paths log sensitive user/role data in production. Remove or gate behind log level. |
| H4  | Async notification errors swallowed       | `backend/services/workflow_execution_service.go` (multiple)          | Notification goroutines print errors with `fmt.Printf` only — no tracking, no retry. At minimum log to structured logger. |
| H5  | Update org member role — endpoint missing | `frontend/src/app/_actions/user-actions.ts:230`                      | Frontend calls `PUT /api/v1/organization/members/:id` which does not exist in backend routes. Implement endpoint. |
| H6  | User profile update skips backend         | `frontend/src/app/_actions/settings.ts:67`                           | Settings save action has TODO "Call backend API to update user profile." Profile changes lost on refresh.     |

### 0.3 Medium Priority — Fix within first week post-launch

| ID  | Item                                      | File                                                                     | Notes                                                                                           |
| --- | ----------------------------------------- | ------------------------------------------------------------------------ | ----------------------------------------------------------------------------------------------- |
| M1  | Admin org membership check placeholder   | `backend/middleware/admin.go:117`                                        | Any admin role can access any org. TODO comment in place. Implement or explicitly document the intentional behavior. |
| M2  | Rate limiter ignores proxy headers        | `backend/middleware/middleware.go`                                        | Keys off raw IP only. Behind a load balancer all requests share one IP. Add `X-Forwarded-For` / `X-Real-IP` parsing. |
| M3  | `session.user as any` casts — 24+ places | Multiple frontend components                                             | Session type from `@auth/core` has no `role` field. Extend `next-auth.d.ts`, then remove all `as any` casts. |
| M4  | PO and PV PDF generators are placeholders | `frontend/src/lib/pdf-generators/purchase-order-pdf.tsx` `:19`          | Both files return placeholder. Install `@react-pdf/renderer`, implement generators. |
| M5  | Subscription upgrade — no payment         | `backend/services/subscription_service.go:246`                           | UpgradeSubscription changes tier in DB only. No payment processor wired. Gate behind payment or remove upgrade UI. |
| M6  | TOTP 2FA records preference only          | `backend/handlers/admin_console_user_handler.go:542`                     | Enable2FA stores a flag but no secret generation, QR enrollment, or verification endpoint. Implement fully or remove 2FA UI. |
| M7  | Vendor placeholder ID hardcoded           | `backend/services/document_automation_service.go:81-90`                  | Uses `"vendor-placeholder-001"` when vendor not found. Should fail or prompt user. |

### 0.4 Low Priority / Post-MVP

| ID  | Item                                     | File                                                               |
| --- | ---------------------------------------- | ------------------------------------------------------------------ |
| L1  | Offline PWA queue — PO/PV/Budget/Vendor  | `frontend/src/hooks/use-offline-queue-processor.ts:290,320,368,378` |
| L2  | Image upload is a placeholder toast      | `frontend/src/components/ui/image-upload.tsx:31`                   |
| L3  | PDF batch export needs `jszip`           | `frontend/src/lib/pdf/pdf-batch-export.ts:40+`                     |
| L4  | Department module assignment stubs       | `backend/services/department_service.go:302-330`                   |
| L5  | Performance logger memory usage = 0      | `backend/logging/middleware/performance_logger.go:213`             |
| L6  | Admin console settings bulk/import/export | `admin-console/src/app/_actions/settings.ts:250-426`              |
| L7  | Admin console feature flag templates     | `admin-console/src/app/_actions/feature-flags.ts:372`              |
| L8  | Google OAuth button is "Coming soon"     | `frontend/src/app/(auth)/login/_components/login-form.tsx:135`     |
| L9  | Notifications need paginated response    | `frontend/src/app/(private)/(main)/notifications/_components/notifications-client.tsx:75` |
| L10 | Workflow document attachments            | `frontend/src/components/workflows/workflow-stage-form.tsx:70`     |
| L11 | Cache manager is a placeholder concept   | `frontend/src/lib/cache-manager.ts:284`                            |
| L12 | Admin user sessions — MVP stub           | `backend/handlers/admin_console_user_handler.go:602`               |
| L13 | API alert rules CRUD — 501 stubs         | `backend/handlers/admin_api_monitoring_handler.go:82,177`          |

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
| [x] | Replace mock bulk operations with real DB calls | `frontend/src/app/_actions/bulk-operations.ts` | 31-288  | High     | **Confirmed implemented (2026-03-26 audit).** Frontend calls real `/api/v1/approvals/bulk/{approve,reject,reassign}` endpoints. Backend routes exist with `RequireFeature("bulk_operations")` middleware. |
| [ ] | Audit trail logging for bulk ops                | `frontend/src/app/_actions/bulk-operations.ts` | 59, 120 | Medium   | TODO: "Log to audit trail." Blocked on audit service implementation (see B3 in Section 0).                                           |

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

> **2026-03-26 audit note**: Migrations directory contains only 000–009. Previous references to 027/031 duplicate/missing files were stale — those migrations no longer exist. All 9 migration files (000–009) have matching `.down.sql` files.

| #      | Item                                | File                                       | Priority | Notes                                                                                                              |
| ------ | ----------------------------------- | ------------------------------------------ | -------- | ------------------------------------------------------------------------------------------------------------------ |
| [DONE] | Fix duplicate migration 027         | `backend/database/migrations/`             | N/A      | **Not present (2026-03-26).** Migrations restructured to 000–009 only. No duplicate exists.                        |
| [DONE] | Commit migration 031                | `backend/database/migrations/`             | N/A      | **Not present (2026-03-26).** File no longer exists in the migrations directory.                                   |
| [DONE] | Add missing DOWN migrations         | `backend/database/migrations/`             | N/A      | **All 9 migrations (000–009) have .down.sql counterparts (2026-03-26 audit).**                                     |
| [ ]    | Verify IsSuperAdmin column nullable | `backend/models/models.go` + migration 017 | Medium   | Model uses `*bool` (nullable pointer). Confirm DB column matches — if NOT NULL default false, change to `bool`.    |

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

## 9. Implementation Roadmap

> Updated 2026-03-26 post-audit. Sprint 0 is new — must complete before any production deploy.

### Sprint 0: MVP Blockers (do this first — ~1 day)

1. **B1** — `backend/main.go`: Replace JWT_SECRET fallback with `log.Fatal` in production
2. **B2** — `backend/services/email_service.go`: Wire real email provider (SMTP/Resend/SendGrid)
3. **B3** — `backend/services/audit_service.go`: Implement `LogAction` / `LogEvent`
4. **H1** — `backend/middleware/middleware.go:278`: Reject (not pass) when session DB check fails
5. **H2** — `backend/main.go`: Add `log.Fatal` guard for required env vars before server starts
6. **H3** — `backend/services/workflow_execution_service.go`: Remove `[DEBUG]` log.Printf statements
7. **H4** — workflow_execution_service.go: Route async notification errors to structured logger
8. **H5** — Implement `PUT /api/v1/organization/members/:id` in backend
9. **H6** — `frontend/src/app/_actions/settings.ts`: Wire profile save to backend API

### Sprint 1: High-Priority Backend (~2 days)

1. **M1** — `backend/middleware/admin.go:117`: Implement org membership check (or document skip)
2. **M5** — `backend/services/subscription_service.go`: Gate upgrade behind payment or disable UI
3. **M6** — Admin TOTP 2FA: Implement secret generation, QR enrollment, verify endpoint — or remove UI
4. **M7** — `backend/services/document_automation_service.go`: Replace vendor placeholder ID with real lookup
5. **M2** — `backend/middleware/middleware.go`: Parse `X-Forwarded-For` for rate limiter

### Sprint 2: Frontend Polish (~2 days)

1. **M3** — Extend NextAuth session type in `next-auth.d.ts`; remove all `session.user as any` casts (24+ occurrences)
2. **M4** — Install `@react-pdf/renderer`; implement PO and PV PDF generators
3. **Admin console** — Settings bulk/import/export + feature flag templates (see Section 2.1, 2.2)
4. **L8** — Remove "Coming soon" Google OAuth button or implement it

### Sprint 3: Completeness (~3 days)

1. **L1** — Offline PWA queue processors for PO, PV, Budget, Vendor
2. **L2** — Image upload: integrate ImageKit or S3-compatible provider
3. **L3** — PDF batch export: install `jszip`, implement batch export
4. **L4** — Department modules: create junction table + assignment handlers
5. **L12** — Admin user sessions: real session tracking table

### Sprint 4: Infrastructure & Testing (~2 days)

1. Repository layer — complete sqlc migration for account lockout + password reset
2. Unit tests — fix workflow state machine, notification service, budget validation tests
3. Performance logger — implement real memory usage via `runtime.ReadMemStats`
4. `IsSuperAdmin` column nullable — verify model vs DB schema (Section 5)

### Deferred / Won't Do

- Database backup/restore/migration via web UI (security decision)
- System config update/service restart via web UI (security decision)
- Cache manager implementation (`frontend/src/lib/cache-manager.ts`) — premature optimization

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
