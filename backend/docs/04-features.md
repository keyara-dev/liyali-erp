# Features

## Approval Workflows

Workflows define multi-stage approval chains. Each stage has:
- An assignee (user or role)
- Optional conditions (amount threshold, department, etc.)
- An action type (approve / reject / return-for-revision)

**Claim system:** Tasks are claimed for 30 minutes. Expired claims auto-release via a background goroutine running every 60s (`StartClaimExpiryWorker`).

**Rejection types:**
- `reject` — terminates the workflow, document status → `rejected`
- `return_for_revision` — resets to a target stage, document status → `revision`

Key files: `services/workflow_execution_service.go`, `handlers/workflow_handler.go`

## Document System

All procurement documents (requisitions, POs, PVs, GRNs) share:
- A `status` field driven by workflow state
- A `metadata` JSONB field (`createdBy`, `updatedBy`, etc.)
- `deleted_at` for soft delete
- QR-code-based public verification at `GET /verify/:documentNumber`
- PDF generation with org branding

Visibility is scoped per role — see `utils/document_scope.go`.

## Organization Invitations

Admins can invite existing platform users to their org:
1. `POST /organization/invitations` — creates a 72h token, sends in-app notification
2. Invitee sees it at `GET /invitations/pending`
3. `POST /invitations/:token/accept` — adds them as org member
4. `POST /invitations/:token/decline`

Stale invitations auto-expire via hourly background goroutine in `main.go`.

Key files: `services/invitation_service.go`, `handlers/invitation_handler.go`

## Subscriptions & Tiers

Three tiers: **Starter** (free), **Pro** ($99/mo), **Enterprise** ($499/mo).

Limits enforced via `organization_has_feature(org_id, feature_key)` PostgreSQL function (defined in `002_subscription_system.up.sql`). Usage checked before create operations.

Organizations start on a 72h trial. After expiry, a grace period can be extended by super admins.

Key files: `services/subscription_service.go`, `handlers/subscription_tier_handler.go`

## Notifications

In-app notifications created on:
- Workflow task assigned / approved / rejected
- Organization invitation received
- Budget threshold reached

Stored in `notifications` table. Fetched via `GET /notifications` (paginated, unread count).

Email notifications: stubbed (`EMAIL_ENABLED=false`). Set `EMAIL_ENABLED=true` and configure SMTP in `services/email_service.go` to enable.

## Audit Logs

Every state-changing operation writes to `audit_logs`:
```
action, entity_type, entity_id, organization_id, user_id, changes (JSONB), created_at
```

Immutable — no update/delete on this table. Queried via admin analytics endpoints.
