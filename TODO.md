# Liyali Gateway - TODO & Future Enhancements

> Last updated: 2026-02-24
> Status: Post Phase 1-8 audit. All P0-P3 gaps from the original plan are closed.
> This document captures every remaining stub, placeholder, and enhancement opportunity.

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

| # | Item | File | Line | Priority | Notes |
|---|------|------|------|----------|-------|
| [ ] | Implement API alert rules CRUD | `backend/handlers/admin_api_monitoring_handler.go` | 177 | Medium | `CreateAPIAlertRule` returns 501. Needs alert rules model + storage. |
| [ ] | Implement API alerts retrieval | `backend/handlers/admin_api_monitoring_handler.go` | 82 | Medium | `GetAPIAlerts` returns empty array. Depends on alert rules. |
| [ ] | Implement endpoint testing | `backend/handlers/admin_api_monitoring_handler.go` | 122 | Low | `TestAPIEndpoint` returns 501. Would need HTTP client to ping endpoints. |
| [ ] | Implement endpoint config updates | `backend/handlers/admin_api_monitoring_handler.go` | 127 | Low | `UpdateAPIEndpointConfig` returns 501. |

### 1.2 Database Management (Intentional 501s)

| # | Item | File | Line | Priority | Notes |
|---|------|------|------|----------|-------|
| [SKIP] | Database backup via web UI | `backend/handlers/admin_database_handler.go` | 510 | N/A | Safety decision: use `pg_dump` CLI instead. |
| [SKIP] | Database restore via web UI | `backend/handlers/admin_database_handler.go` | 515 | N/A | Safety decision: use `pg_restore` CLI instead. |
| [SKIP] | Run migrations via web UI | `backend/handlers/admin_database_handler.go` | 575 | N/A | Safety decision: use deployment pipeline. |
| [SKIP] | Rollback migrations via web UI | `backend/handlers/admin_database_handler.go` | 580 | N/A | Safety decision: use manual SQL scripts. |
| [SKIP] | Export database via web UI | `backend/handlers/admin_database_handler.go` | 618 | N/A | Safety decision: use `pg_dump` CLI instead. |

### 1.3 System Health (Intentional 501s)

| # | Item | File | Line | Priority | Notes |
|---|------|------|------|----------|-------|
| [SKIP] | Update system config via web UI | `backend/handlers/admin_system_health_handler.go` | 316 | N/A | Safety: config should be managed via environment variables/deployment. |
| [SKIP] | Restart services via web UI | `backend/handlers/admin_system_health_handler.go` | 327 | N/A | Safety: use deployment tools (Fly.io, Docker, etc.). |

### 1.4 Admin Users

| # | Item | File | Line | Priority | Notes |
|---|------|------|------|----------|-------|
| [~] | Full TOTP 2FA infrastructure | `backend/handlers/admin_console_user_handler.go` | 542 | High | Currently records preference only. Needs TOTP secret generation, QR code, verification, backup codes. |
| [~] | Admin user sessions (real) | `backend/handlers/admin_console_user_handler.go` | 602 | Medium | MVP stub querying audit logs as proxy. Needs real session tracking table. |

---

## 2. Admin Console - Frontend

### 2.1 Settings Page

| # | Item | File | Line | Priority | Notes |
|---|------|------|------|----------|-------|
| [ ] | Bulk update settings | `admin-console/src/app/_actions/settings.ts` | 250 | Medium | Returns "not yet implemented". Needs backend endpoint `POST /admin/settings/bulk-update`. |
| [ ] | Validate configuration | `admin-console/src/app/_actions/settings.ts` | 309 | Low | Returns validation error stub. Could check required settings exist + correct types. |
| [ ] | Import configuration | `admin-console/src/app/_actions/settings.ts` | 410 | Low | Returns "not yet implemented". Needs backend `POST /admin/settings/import`. |
| [ ] | Reset to defaults | `admin-console/src/app/_actions/settings.ts` | 418 | Low | Returns "not yet implemented". Needs backend endpoint. |
| [ ] | Restore configuration | `admin-console/src/app/_actions/settings.ts` | 426 | Low | Returns "not yet implemented". Needs snapshot/backup system for settings. |
| [ ] | Export settings (real download) | `admin-console/src/app/admin/settings/page.tsx` | 157 | Low | Shows "coming soon" toast. Needs to call backend and trigger download. |

### 2.2 Feature Flags Page

| # | Item | File | Line | Priority | Notes |
|---|------|------|------|----------|-------|
| [ ] | Bulk update feature flags | `admin-console/src/app/_actions/feature-flags.ts` | 299 | Medium | Console.warn only. Needs backend `POST /admin/feature-flags/bulk-update`. |
| [ ] | Feature flag templates | `admin-console/src/app/_actions/feature-flags.ts` | 372 | Low | Returns empty array. Needs `feature_flag_templates` table + CRUD handlers. |
| [ ] | Feature flag evaluation log | `admin-console/src/app/_actions/feature-flags.ts` | 335 | Low | Returns empty array. `getFeatureFlagEvaluations` needs evaluation tracking. |
| [ ] | Feature flag audit history | `admin-console/src/app/_actions/feature-flags.ts` | 384 | Low | Returns empty array. Could query `admin_audit_logs` filtered by flag actions. |
| [ ] | Import feature flags | `admin-console/src/app/_actions/feature-flags.ts` | 425 | Low | Returns "not yet implemented". Needs backend `POST /admin/feature-flags/import`. |
| [ ] | Export feature flags (real download) | `admin-console/src/app/admin/feature-flags/page.tsx` | 206 | Low | Shows "coming soon" toast. |

---

## 3. Backend Services & Infrastructure

### 3.1 Core Services

| # | Item | File | Line | Priority | Notes |
|---|------|------|------|----------|-------|
| [ ] | Audit service implementation | `backend/services/audit_service.go` | 11-37 | High | Multiple TODOs. `LogAction` and `LogEvent` are empty stubs. Needed for proper audit trail. |
| [ ] | Subscription upgrade with payment | `backend/services/subscription_service.go` | 246 | High | TODO: "Implement actual upgrade logic with payment processing." Currently just changes tier. |
| [ ] | Department module assignment | `backend/services/department_service.go` | 302-330 | Medium | `GetDepartmentModules`, `AssignModuleToDepartment`, `RemoveModuleFromDepartment` are placeholders. Needs `department_modules` junction table. |
| [ ] | Auth service `GetUserProfile` | `backend/handlers/auth_handler.go` | 470 | Medium | TODO comment. Currently queries DB directly. Should use service layer. |

### 3.2 Infrastructure

| # | Item | File | Line | Priority | Notes |
|---|------|------|------|----------|-------|
| [ ] | Admin middleware org membership check | `backend/middleware/admin.go` | 117 | Medium | TODO: "Implement proper organization membership check." Currently placeholder. |
| [~] | Account lockout repository | `backend/repository/account_lockout_repository.go` | - | Medium | "Temporary implementations that return not implemented errors." Waiting for sqlc. |
| [~] | Password reset repository | `backend/repository/password_reset_repository.go` | - | Medium | Same as above - temporary until sqlc generation works. |
| [ ] | Performance logger memory usage | `backend/logging/middleware/performance_logger.go` | 213 | Low | `getCurrentMemoryUsage()` returns hardcoded 0. Should use `runtime.ReadMemStats`. |
| [ ] | Enhanced auth model placeholder | `backend/models/enhanced_auth.go` | 601 | Low | "returning true as a placeholder" - needs proper implementation. |
| [ ] | Notification service initialization | `backend/main.go` | 116 | Low | "placeholder for now" - may need proper lifecycle management. |

### 3.3 Tests

| # | Item | File | Priority | Notes |
|---|------|------|----------|-------|
| [ ] | Workflow state machine tests | `backend/tests/unit/workflow_state_machine_test.go` | Medium | Needs `WorkflowState` constants defined. |
| [ ] | Notification service tests | `backend/tests/unit/notification_service_test.go` | Medium | Needs `NotificationEvent` model. |
| [ ] | Budget validation tests | `backend/tests/unit/budget_validation_test.go` | Low | Needs `BudgetConstraint` model. |
| [ ] | Document linking tests | `backend/tests/unit/document_linking_test.go` | Low | Needs `DocumentLink` model. |

---

## 4. Main Frontend (Tenant App)

### 4.1 Dashboard & Analytics

| # | Item | File | Line | Priority | Notes |
|---|------|------|------|----------|-------|
| [ ] | Average processing time metric | `frontend/src/app/_actions/dashboard.ts` | 38 | Medium | TODO: "Add to backend analytics." |
| [ ] | Budget utilization metric | `frontend/src/app/_actions/dashboard.ts` | 39 | Medium | TODO: "Add to backend analytics." |
| [ ] | Recent activity feed | `frontend/src/app/_actions/dashboard.ts` | 40 | Medium | TODO: "Add to backend analytics." |
| [ ] | Purchase order metrics | `frontend/src/app/_actions/dashboard.ts` | 48 | Medium | TODO: "Add to backend analytics." |
| [ ] | Payment voucher metrics | `frontend/src/app/_actions/dashboard.ts` | 49 | Medium | TODO: "Add to backend analytics." |
| [ ] | Budget metrics | `frontend/src/app/_actions/dashboard.ts` | 50 | Medium | TODO: "Add to backend analytics." |
| [ ] | Average approval time | `frontend/src/app/_actions/dashboard.ts` | 60 | Medium | TODO: "Add to backend analytics." |
| [ ] | Document type breakdown | `frontend/src/app/_actions/dashboard.ts` | 64-66 | Medium | TODO: "Add to backend analytics." |

### 4.2 Bulk Operations

| # | Item | File | Line | Priority | Notes |
|---|------|------|------|----------|-------|
| [ ] | Replace mock bulk operations with real DB calls | `frontend/src/app/_actions/bulk-operations.ts` | 31-288 | High | Multiple TODOs: "In production, replace with actual database operations." All bulk approve/reject/reassign use simulated operations. |
| [ ] | Audit trail logging for bulk ops | `frontend/src/app/_actions/bulk-operations.ts` | 59, 120 | Medium | TODO: "Log to audit trail." |

### 4.3 Offline / PWA Features

| # | Item | File | Line | Priority | Notes |
|---|------|------|------|----------|-------|
| [ ] | Purchase order offline operations | `frontend/src/hooks/use-offline-queue-processor.ts` | 290 | Low | Throws "not implemented yet." |
| [ ] | Payment voucher offline operations | `frontend/src/hooks/use-offline-queue-processor.ts` | 320 | Low | Throws "not implemented yet." |
| [ ] | Budget offline operations | `frontend/src/hooks/use-offline-queue-processor.ts` | 368 | Low | Throws "not implemented yet." |
| [ ] | Vendor offline operations | `frontend/src/hooks/use-offline-queue-processor.ts` | 378 | Low | Throws "not implemented yet." |
| [ ] | Offline queue backend endpoint | `frontend/src/lib/storage/hooks.ts` | 27 | Low | TODO: "Replace with real backend API endpoint." |

### 4.4 Document & File Generation

| # | Item | File | Line | Priority | Notes |
|---|------|------|------|----------|-------|
| [ ] | Purchase order PDF generation | `frontend/src/lib/pdf-generators/purchase-order-pdf.tsx` | 19 | Medium | Placeholder. Needs `@react-pdf/renderer`. |
| [ ] | Payment voucher PDF generation | `frontend/src/lib/pdf-generators/payment-voucher-pdf.tsx` | 19 | Medium | Placeholder. Needs `@react-pdf/renderer`. |
| [ ] | PDF batch export (JSZip) | `frontend/src/lib/pdf/pdf-batch-export.ts` | 40+ | Low | "JSZip library not available" - needs `jszip` package. |
| [ ] | Image upload integration | `frontend/src/components/ui/image-upload.tsx` | 31 | Medium | Placeholder toast. Needs `@imagekit/next` or alternative upload provider. |

### 4.5 Payments & Subscriptions

| # | Item | File | Line | Priority | Notes |
|---|------|------|------|----------|-------|
| [ ] | Payment method integration | `frontend/src/components/subscription/upgrade-modal.tsx` | 128 | High | TODO: "Add payment method integration." Stripe/payment provider needed. |
| [ ] | Contact sales flow | `frontend/src/components/subscription/upgrade-modal.tsx` | 114 | Medium | TODO: "Implement contact sales flow." |
| [ ] | Organization tier upgrade API | `frontend/src/hooks/use-organization-tier.ts` | 27 | Medium | TODO: "Implement actual API call to upgrade organization." |

### 4.6 User Management

| # | Item | File | Line | Priority | Notes |
|---|------|------|------|----------|-------|
| [ ] | Update org member role endpoint | `frontend/src/app/_actions/user-actions.ts` | 230 | Medium | TODO: "Backend needs to implement PUT /api/v1/organization/members/:id endpoint." |
| [ ] | User profile update via backend | `frontend/src/app/_actions/settings.ts` | 67 | Medium | TODO: "Call backend API to update user profile." |

### 4.7 Miscellaneous

| # | Item | File | Line | Priority | Notes |
|---|------|------|------|----------|-------|
| [ ] | Notifications pagination | `frontend/src/app/(private)/(main)/notifications/_components/notifications-client.tsx` | 75 | Low | TODO: "Backend should return PaginatedResponse with pagination info." |
| [ ] | Workflow document attachments | `frontend/src/components/workflows/workflow-stage-form.tsx` | 70 | Low | "Attach relevant documents (not yet implemented)." Needs file upload. |
| [ ] | Permissions hook fallback | `frontend/src/hooks/use-permissions.ts` | 152 | Low | Uses "hardcoded permissions for built-in roles" as fallback. |
| [ ] | Cache manager concept | `frontend/src/lib/cache-manager.ts` | 284 | Low | "This is a placeholder for the concept." |

---

## 5. Backend Data Issues

| # | Item | File | Line | Priority | Notes |
|---|------|------|------|----------|-------|
| [ ] | Vendor placeholder ID | `backend/services/document_automation_service.go` | 81-90 | Medium | Uses hardcoded "vendor-placeholder-001" when vendor not found. |
| [ ] | Organization contact/billing stubs | `backend/handlers/admin_organization_handler.go` | 232 | Low | "Add contact_info and billing_info stubs (frontend expects these nested objects)." |
| [ ] | Organization service logging | `backend/services/organization_service.go` | 66 | Low | TODO: "Add proper logging here." |

---

## 6. Suggested Implementation Roadmap

### Sprint 1: High-Priority Backend (estimated: ~2-3 days of work)
1. **Audit service** - Implement `LogAction` and `LogEvent` so all actions are tracked
2. **Bulk operations** - Replace simulated bulk approve/reject with real DB operations
3. **2FA infrastructure** - Add TOTP secret generation, QR enrollment, verification endpoint
4. **Payment integration** - Stripe checkout for subscription upgrades

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
