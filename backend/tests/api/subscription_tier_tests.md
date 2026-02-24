# Subscription Tier System API Tests

Manual API tests for the 3-tier subscription system (Starter/Pro/Custom).

---

## Prerequisites

1. Backend server running on `http://localhost:8081`
2. Database migrations applied (including migration 014)
3. Super admin user exists (admin@liyali.com / password)
4. Test organization exists (org-demo-001)

---

## Setup: Get Authentication Token

```bash
# Login as super admin
curl -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@liyali.com",
    "password": "password"
  }'

# Save the token from response
export TOKEN="your_jwt_token_here"
```

---

## Test 1: Successful Tier Upgrade (Starter → Pro)

### Request

```bash
curl -X POST http://localhost:8081/api/v1/admin/organizations/org-demo-001/change-tier \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "newTier": "pro",
    "reason": "Customer upgrade request - moving from trial to paid plan"
  }'
```

### Expected Response (200 OK)

```json
{
  "success": true,
  "message": "Tier changed successfully",
  "data": {
    "organization_id": "org-demo-001",
    "old_tier": "starter",
    "new_tier": "pro"
  }
}
```

### Verification

```bash
# Check organization tier
curl -X GET http://localhost:8081/api/v1/admin/organizations/org-demo-001 \
  -H "Authorization: Bearer $TOKEN"

# Check subscription events
psql $DATABASE_URL -c "SELECT * FROM subscription_events WHERE organization_id = 'org-demo-001' ORDER BY created_at DESC LIMIT 1;"

# Check audit logs
psql $DATABASE_URL -c "SELECT * FROM admin_audit_logs WHERE organization_id = 'org-demo-001' AND action = 'tier_change' ORDER BY created_at DESC LIMIT 1;"
```

---

## Test 2: Tier Downgrade (Pro → Starter)

### Request

```bash
curl -X POST http://localhost:8081/api/v1/admin/organizations/org-demo-001/change-tier \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "newTier": "starter",
    "reason": "Customer requested downgrade due to budget constraints"
  }'
```

### Expected Response (200 OK)

```json
{
  "success": true,
  "message": "Tier changed successfully",
  "data": {
    "organization_id": "org-demo-001",
    "old_tier": "pro",
    "new_tier": "starter"
  }
}
```

---

## Test 3: Upgrade to Custom (Unlimited)

### Request

```bash
curl -X POST http://localhost:8081/api/v1/admin/organizations/org-demo-001/change-tier \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "newTier": "custom",
    "reason": "Large organization requiring unlimited resources and dedicated support"
  }'
```

### Expected Response (200 OK)

Event type should be "subscription_upgraded"

---

## Test 4: Invalid Tier

### Request

```bash
curl -X POST http://localhost:8081/api/v1/admin/organizations/org-demo-001/change-tier \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "newTier": "invalid_tier",
    "reason": "Testing invalid tier validation"
  }'
```

### Expected Response (400 Bad Request)

```json
{
  "success": false,
  "message": "Invalid subscription tier",
  "error": "tier must be one of: starter, pro, custom"
}
```

---

## Test 5: Missing Reason

### Request

```bash
curl -X POST http://localhost:8081/api/v1/admin/organizations/org-demo-001/change-tier \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "newTier": "pro"
  }'
```

### Expected Response (400 Bad Request)

```json
{
  "success": false,
  "message": "Validation failed",
  "error": "reason is required"
}
```

---

## Test 6: Reason Too Short

### Request

```bash
curl -X POST http://localhost:8081/api/v1/admin/organizations/org-demo-001/change-tier \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "newTier": "pro",
    "reason": "Short"
  }'
```

### Expected Response (400 Bad Request)

```json
{
  "success": false,
  "message": "Validation failed",
  "error": "reason must be at least 10 characters"
}
```

---

## Test 7: Same Tier (No Change)

### Request

```bash
# First, check current tier
curl -X GET http://localhost:8081/api/v1/admin/organizations/org-demo-001 \
  -H "Authorization: Bearer $TOKEN" | jq '.data.subscription_tier'

# Try to change to same tier
curl -X POST http://localhost:8081/api/v1/admin/organizations/org-demo-001/change-tier \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "newTier": "starter",
    "reason": "Testing same tier change"
  }'
```

### Expected Response (400 Bad Request)

```json
{
  "success": false,
  "message": "Organization is already on this tier",
  "error": "Current tier is already starter"
}
```

---

## Test 8: Organization Not Found

### Request

```bash
curl -X POST http://localhost:8081/api/v1/admin/organizations/org-nonexistent/change-tier \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "newTier": "pro",
    "reason": "Testing non-existent organization"
  }'
```

### Expected Response (404 Not Found)

```json
{
  "success": false,
  "message": "Organization not found",
  "error": "No organization found with ID: org-nonexistent"
}
```

---

## Test 9: Unauthorized (No Token)

### Request

```bash
curl -X POST http://localhost:8081/api/v1/admin/organizations/org-demo-001/change-tier \
  -H "Content-Type: application/json" \
  -d '{
    "newTier": "pro",
    "reason": "Testing unauthorized access"
  }'
```

### Expected Response (401 Unauthorized)

```json
{
  "success": false,
  "message": "Unauthorized",
  "error": "Missing or invalid authentication token"
}
```

---

## Test 10: Forbidden (Non-Super Admin)

### Request

```bash
# Login as regular user
curl -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "requester@liyali.com",
    "password": "password"
  }'

# Save regular user token
export REGULAR_TOKEN="regular_user_token_here"

# Try to change tier
curl -X POST http://localhost:8081/api/v1/admin/organizations/org-demo-001/change-tier \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $REGULAR_TOKEN" \
  -d '{
    "newTier": "pro",
    "reason": "Testing non-super admin access"
  }'
```

### Expected Response (403 Forbidden)

```json
{
  "success": false,
  "message": "Insufficient permissions",
  "error": "Only super admins can change subscription tiers"
}
```

---

## Test 12: Malformed JSON

### Request

```bash
curl -X POST http://localhost:8081/api/v1/admin/organizations/org-demo-001/subscription-tier \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "tier": "professional",
    "reason": "Testing malformed JSON"
  '
```

### Expected Response (400 Bad Request)

```json
{
  "success": false,
  "message": "Invalid request body",
  "error": "unexpected end of JSON input"
}
```

---

## Test 13: Empty Request Body

### Request

```bash
curl -X POST http://localhost:8081/api/v1/admin/organizations/org-demo-001/subscription-tier \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{}'
```

### Expected Response (400 Bad Request)

```json
{
  "success": false,
  "message": "Validation failed",
  "error": "tier is required"
}
```

---

## Test 11: All Tier Combinations

### Requests

```bash
# Starter → Pro
curl -X POST http://localhost:8081/api/v1/admin/organizations/org-demo-001/change-tier \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"newTier": "pro", "reason": "Testing starter to pro"}'

# Pro → Custom
curl -X POST http://localhost:8081/api/v1/admin/organizations/org-demo-001/change-tier \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"newTier": "custom", "reason": "Testing pro to custom"}'

# Custom → Starter (big downgrade)
curl -X POST http://localhost:8081/api/v1/admin/organizations/org-demo-001/change-tier \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"newTier": "starter", "reason": "Testing custom to starter downgrade"}'
```

### Expected

All should return 200 OK with appropriate event types (upgraded/downgraded)

---

## Test 12: Concurrent Requests

### Request

```bash
# Run multiple requests simultaneously
for i in {1..5}; do
  curl -X POST http://localhost:8081/api/v1/admin/organizations/org-demo-001/change-tier \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN" \
    -d "{\"newTier\": \"pro\", \"reason\": \"Concurrent test $i\"}" &
done
wait
```

### Expected

- First request should succeed (200 OK)
- Subsequent requests should fail with "already on this tier" (400 Bad Request)
- Database should have exactly one successful tier change
- Transaction isolation should prevent race conditions

---

## Verification Queries

### Check Organization Tier

```sql
SELECT id, name, subscription_tier, subscription_status, updated_at
FROM organizations
WHERE id = 'org-demo-001';
```

### Check Subscription Events

```sql
SELECT id, event_type, from_tier, to_tier, created_by, created_at, metadata
FROM subscription_events
WHERE organization_id = 'org-demo-001'
ORDER BY created_at DESC
LIMIT 10;
```

### Check Audit Logs

```sql
SELECT id, action, old_value, new_value, reason, admin_user_id, created_at, details
FROM admin_audit_logs
WHERE organization_id = 'org-demo-001'
AND action = 'tier_change'
ORDER BY created_at DESC
LIMIT 10;
```

### Check Event Type Distribution

```sql
SELECT event_type, COUNT(*) as count
FROM subscription_events
WHERE organization_id = 'org-demo-001'
GROUP BY event_type;
```

---

## Postman Collection

Import this JSON into Postman:

```json
{
  "info": {
    "name": "Subscription Tier Change API",
    "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
  },
  "item": [
    {
      "name": "Login as Super Admin",
      "request": {
        "method": "POST",
        "header": [{ "key": "Content-Type", "value": "application/json" }],
        "body": {
          "mode": "raw",
          "raw": "{\"email\":\"admin@liyali.com\",\"password\":\"password\"}"
        },
        "url": { "raw": "{{baseUrl}}/api/v1/auth/login" }
      }
    },
    {
      "name": "Change Tier - Success",
      "request": {
        "method": "POST",
        "header": [
          { "key": "Content-Type", "value": "application/json" },
          { "key": "Authorization", "value": "Bearer {{token}}" }
        ],
        "body": {
          "mode": "raw",
          "raw": "{\"tier\":\"professional\",\"reason\":\"Customer upgrade request\"}"
        },
        "url": {
          "raw": "{{baseUrl}}/api/v1/admin/organizations/org-demo-001/subscription-tier"
        }
      }
    },
    {
      "name": "Change Tier - Invalid Tier",
      "request": {
        "method": "POST",
        "header": [
          { "key": "Content-Type", "value": "application/json" },
          { "key": "Authorization", "value": "Bearer {{token}}" }
        ],
        "body": {
          "mode": "raw",
          "raw": "{\"tier\":\"invalid\",\"reason\":\"Testing invalid tier\"}"
        },
        "url": {
          "raw": "{{baseUrl}}/api/v1/admin/organizations/org-demo-001/subscription-tier"
        }
      }
    }
  ],
  "variable": [
    { "key": "baseUrl", "value": "http://localhost:8081" },
    { "key": "token", "value": "" }
  ]
}
```

---

## Test Summary Checklist

- [ ] Test 1: Successful upgrade (Starter → Pro)
- [ ] Test 2: Successful downgrade (Pro → Starter)
- [ ] Test 3: Upgrade to Custom (unlimited)
- [ ] Test 4: Invalid tier validation
- [ ] Test 5: Missing reason validation
- [ ] Test 6: Short reason validation
- [ ] Test 7: Same tier rejection
- [ ] Test 8: Organization not found
- [ ] Test 9: Unauthorized (no token)
- [ ] Test 10: Forbidden (non-super admin)
- [ ] Test 11: All tier combinations
- [ ] Test 12: Concurrent requests handling

---

**Total Tests:** 12  
**Expected Pass Rate:** 100%  
**Estimated Test Time:** 10-15 minutes
