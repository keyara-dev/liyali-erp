# Trial Reset API Documentation

## Overview

The Trial Reset API allows administrators to reset the trial period for organizations, giving them a fresh trial with new start and end dates.

## Endpoint

```
POST /api/v1/organizations/{id}/trial/reset
```

## Authentication & Authorization

- **Authentication**: Required (JWT token)
- **Authorization**: Admin role required (`admin`, `super_admin`, or `compliance_officer`)
- **Middleware**: `AdminMiddleware()` applied

## Request Parameters

### Path Parameters

| Parameter | Type   | Required | Description     |
| --------- | ------ | -------- | --------------- |
| `id`      | string | Yes      | Organization ID |

### Request Body

```json
{
  "trialDays": 30,
  "reason": "Customer requested trial extension due to implementation delays"
}
```

| Field       | Type   | Required | Validation  | Description                    |
| ----------- | ------ | -------- | ----------- | ------------------------------ |
| `trialDays` | number | Yes      | 1-90        | Number of trial days to grant  |
| `reason`    | string | Yes      | 5-200 chars | Reason for resetting the trial |

## Response

### Success Response (200 OK)

```json
{
  "success": true,
  "data": {
    "organizationId": "org-123",
    "subscriptionStatus": "trial",
    "trialStartDate": "2026-02-03T15:30:00Z",
    "trialEndDate": "2026-03-05T15:30:00Z",
    "gracePeriodEndsAt": null,
    "planSlug": "STARTER_PLAN",
    "planName": "Starter Plan",
    "daysRemaining": 30,
    "isExpired": false,
    "isActive": true,
    "inGracePeriod": false
  },
  "message": "Trial reset successfully"
}
```

### Error Responses

#### 400 Bad Request

```json
{
  "success": false,
  "message": "Organization ID is required"
}
```

#### 401 Unauthorized

```json
{
  "success": false,
  "message": "Authentication required"
}
```

#### 403 Forbidden

```json
{
  "success": false,
  "message": "Admin privileges required"
}
```

#### 500 Internal Server Error

```json
{
  "success": false,
  "message": "Failed to reset trial"
}
```

## What the API Does

1. **Validates Input**: Checks organization ID, trial days (1-90), and reason (5-200 chars)
2. **Calculates New Dates**: Sets trial start to current time, end to start + trial days
3. **Updates Organization**:
   - Sets `trial_start_date` to current timestamp
   - Sets `trial_end_date` to calculated end date
   - Sets `subscription_status` to `'trial'`
   - Clears `grace_period_ends_at` (resets any previous extensions)
4. **Creates Audit Log**: Records the trial reset action with metadata
5. **Returns Updated Status**: Provides the new trial status information

## Database Changes

The API updates the `organizations` table:

```sql
UPDATE organizations
SET trial_start_date = NOW(),
    trial_end_date = NOW() + INTERVAL '30 days',
    subscription_status = 'trial',
    grace_period_ends_at = NULL,
    updated_at = CURRENT_TIMESTAMP
WHERE id = 'org-123';
```

And creates an audit log entry:

```sql
INSERT INTO subscription_audit_logs (
    organization_id, action, metadata, performed_by
) VALUES (
    'org-123',
    'trial_reset',
    '{"trial_days": 30, "reason": "...", "action_type": "trial_reset", ...}',
    'admin-user-id'
);
```

## Usage Examples

### cURL Example

```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "trialDays": 30,
    "reason": "Customer needs additional time for evaluation"
  }' \
  "https://api.liyali.com/api/v1/organizations/org-123/trial/reset"
```

### JavaScript Example

```javascript
const response = await fetch("/api/v1/organizations/org-123/trial/reset", {
  method: "POST",
  headers: {
    "Content-Type": "application/json",
    Authorization: `Bearer ${token}`,
  },
  body: JSON.stringify({
    trialDays: 30,
    reason: "Customer needs additional time for evaluation",
  }),
});

const result = await response.json();
console.log(result);
```

## Related Endpoints

- `GET /api/v1/organizations/{id}/trial-status` - Get current trial status
- `POST /api/v1/organizations/{id}/trial/extend` - Extend existing trial (adds to grace period)
- `GET /api/v1/organizations/{id}/subscription` - Get full subscription details

## Differences from Trial Extension

| Feature          | Trial Reset                  | Trial Extension                 |
| ---------------- | ---------------------------- | ------------------------------- |
| **Purpose**      | Start fresh trial            | Add time to existing trial      |
| **Trial Dates**  | Sets new start/end dates     | Extends grace period            |
| **Status**       | Always sets to 'trial'       | Maintains current status        |
| **Use Case**     | Restart expired trials       | Give more time to active trials |
| **Grace Period** | Clears existing grace period | Extends grace period            |

## Security Considerations

- Only admin users can reset trials
- All actions are logged for audit purposes
- Reason is required to ensure accountability
- Trial days are limited to 90 to prevent abuse

## Testing

Use the provided test script:

```bash
./backend/scripts/test_trial_reset.sh org-123 30 "Testing trial reset"
```

## Monitoring

Monitor the following for trial reset activities:

- Subscription audit logs (`subscription_audit_logs` table)
- Application logs (search for "trial reset")
- Organization status changes
- Trial conversion metrics after resets
