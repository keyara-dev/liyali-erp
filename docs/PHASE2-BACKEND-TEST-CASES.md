# Phase 2: Backend Test Cases for Auto-Organization Registration

## Test Environment Setup

### Prerequisites
- Backend running on `http://localhost:8080`
- Database initialized and running
- POST tool ready (curl, Postman, or Thunder Client)

### Database State Before Testing
- Clean slate (no test users from new emails)
- Organization tables ready
- All migrations applied

---

## Test Case 1: Valid Registration - Happy Path ✅

### Objective
Verify that a new user can successfully register, organization is auto-created, and response includes all required data.

### Request
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "testuser1@example.com",
    "password": "SecurePass123",
    "name": "Test User One",
    "role": "requester"
  }'
```

### Expected Response (201 Created)
```json
{
  "success": true,
  "message": "Registration successful",
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": "user_abc123...",
    "email": "testuser1@example.com",
    "name": "Test User One",
    "role": "requester",
    "active": true,
    "createdAt": "2025-12-25T10:30:00Z"
  },
  "organization": {
    "id": "org_def456...",
    "name": "Test User One",
    "slug": "testuser1-example.com",
    "description": "Personal Organization",
    "active": true,
    "tier": "free",
    "createdAt": "2025-12-25T10:30:00Z"
  }
}
```

### Verification Checklist
- [ ] Status code: **201 Created**
- [ ] `success` field: **true**
- [ ] `token` field: **Not empty, valid JWT format**
- [ ] `user` field: **Contains id, email, name, role, active, createdAt**
- [ ] `organization` field: **Present with id, name, slug, description, active, tier, createdAt**
- [ ] User email matches request
- [ ] Organization name matches user name
- [ ] Organization slug derived from email (lowercase, hyphenated)

### Post-Test Database Checks
```sql
-- Verify user was created
SELECT id, email, name, role, active FROM users WHERE email = 'testuser1@example.com';

-- Verify organization was created
SELECT id, name, slug, description, active, tier FROM organizations WHERE name = 'Test User One';

-- Verify user is member of organization (as admin)
SELECT user_id, organization_id, role FROM organization_members
  WHERE user_id = '<user_id_from_response>';

-- Verify organization settings were created
SELECT organization_id, feature_flags FROM organization_settings
  WHERE organization_id = '<org_id_from_response>';

-- Verify current_organization_id is set on user
SELECT current_organization_id FROM users WHERE id = '<user_id_from_response>';
```

---

## Test Case 2: Verify JWT Contains Organization Context ✅

### Objective
Verify that the JWT token includes the organization ID in the claims.

### Process
1. From Test Case 1, extract the `token` value
2. Decode the JWT at [jwt.io](https://jwt.io) or programmatically

### Decoded JWT Should Contain
```json
{
  "sub": "user_abc123...",
  "email": "testuser1@example.com",
  "name": "Test User One",
  "role": "requester",
  "currentOrgId": "org_def456...",
  "iat": 1703502600,
  "exp": 1703589000
}
```

### Verification Checklist
- [ ] **sub** claim contains user ID
- [ ] **email** claim matches registration email
- [ ] **name** claim matches registration name
- [ ] **role** claim matches registration role
- [ ] **currentOrgId** claim: **Present and matches organization ID from response**
- [ ] **exp** claim indicates 24-hour expiration from iat

---

## Test Case 3: Duplicate Email Registration - Conflict ❌

### Objective
Verify that registering with an email that already exists returns a 409 Conflict error.

### Request
```bash
# First registration
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "duplicate@example.com",
    "password": "SecurePass123",
    "name": "First User",
    "role": "requester"
  }'

# Second registration with same email
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "duplicate@example.com",
    "password": "DifferentPass456",
    "name": "Second User",
    "role": "approver"
  }'
```

### Expected Response (409 Conflict)
```json
{
  "success": false,
  "message": "Email already registered"
}
```

### Verification Checklist
- [ ] Status code: **409 Conflict**
- [ ] `success` field: **false**
- [ ] `message` field: **"Email already registered"**
- [ ] No `token` field in response
- [ ] No `user` field in response
- [ ] No new user created in database
- [ ] No additional organization created

---

## Test Case 4: Weak Password Validation ❌

### Objective
Verify that password strength validation rejects weak passwords.

### Test Cases
```bash
# Password too short (less than 8 characters)
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "weakpass1@example.com",
    "password": "Pass1",
    "name": "Test User",
    "role": "requester"
  }'

# No uppercase letter
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "weakpass2@example.com",
    "password": "password123",
    "name": "Test User",
    "role": "requester"
  }'

# No lowercase letter
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "weakpass3@example.com",
    "password": "PASSWORD123",
    "name": "Test User",
    "role": "requester"
  }'

# No digit
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "weakpass4@example.com",
    "password": "PasswordOnly",
    "name": "Test User",
    "role": "requester"
  }'
```

### Expected Response (400 Bad Request)
```json
{
  "success": false,
  "message": "Password does not meet requirements",
  "error": "password must contain: at least 8 characters, 1 uppercase, 1 lowercase, 1 digit"
}
```

### Verification Checklist
- [ ] Status code: **400 Bad Request**
- [ ] `success` field: **false**
- [ ] `message` field: **"Password does not meet requirements"**
- [ ] No user created for any of these requests
- [ ] No organizations created

---

## Test Case 5: Missing Required Fields ❌

### Objective
Verify that incomplete requests are rejected with proper validation.

### Test Cases
```bash
# Missing email
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "password": "SecurePass123",
    "name": "Test User",
    "role": "requester"
  }'

# Missing password
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "missing@example.com",
    "name": "Test User",
    "role": "requester"
  }'

# Missing name
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "missing@example.com",
    "password": "SecurePass123",
    "role": "requester"
  }'

# Missing role
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "missing@example.com",
    "password": "SecurePass123",
    "name": "Test User"
  }'
```

### Expected Response (400 Bad Request)
```json
{
  "success": false,
  "message": "Email, password, name, and role are required"
}
```

### Verification Checklist
- [ ] Status code: **400 Bad Request** for all missing field cases
- [ ] `success` field: **false**
- [ ] `message` indicates which fields are required
- [ ] No users/organizations created for any of these requests

---

## Test Case 6: Invalid Role ❌

### Objective
Verify that invalid roles are rejected.

### Request
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "invalidrole@example.com",
    "password": "SecurePass123",
    "name": "Test User",
    "role": "superadmin"
  }'
```

### Expected Response (400 Bad Request)
```json
{
  "success": false,
  "message": "Invalid role"
}
```

### Valid Roles
- `admin`
- `approver`
- `requester` ← Default
- `finance`
- `viewer`

### Verification Checklist
- [ ] Status code: **400 Bad Request**
- [ ] `success` field: **false**
- [ ] `message` field: **"Invalid role"**
- [ ] No user created
- [ ] No organization created

---

## Test Case 7: Password Correctly Hashed (Security) ✅

### Objective
Verify that passwords are stored as bcrypt hashes, not plaintext.

### Process
1. Register a new user (use Test Case 1)
2. Query database directly: `SELECT password FROM users WHERE email = 'testuser1@example.com'`

### Verification Checklist
- [ ] Password in database **starts with `$2b$` or `$2a$`** (bcrypt signature)
- [ ] Password in database **is NOT** the plaintext password
- [ ] Password length is **~60 characters** (bcrypt hash length)
- [ ] Can login with plaintext password (bcrypt.CompareHashAndPassword works)

### Negative Test
- [ ] Login with wrong password fails
- [ ] Login with password as registered plaintext fails

---

## Test Case 8: Organization Creation Failure Handling (Graceful Degradation) ✅

### Objective
Verify that if organization creation fails, the user is still created and returns partial response.

### Simulate Failure
To test this scenario, temporarily modify the code to force org creation to fail:

```go
// In auth.go Register function, temporarily add:
org, err := orgService.CreateOrganization(newUser.Name, "Personal Organization", newUser.ID)
if err != nil {
  log.Printf("Warning: Failed to create personal organization for user %s: %v", newUser.Email, err)
  // org remains nil
}
```

### Request
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "orgerror@example.com",
    "password": "SecurePass123",
    "name": "Test User",
    "role": "requester"
  }'
```

### Expected Response (201 Created - Partial)
```json
{
  "success": true,
  "message": "Registration successful",
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": "user_xyz789...",
    "email": "orgerror@example.com",
    "name": "Test User",
    "role": "requester",
    "active": true,
    "createdAt": "2025-12-25T10:30:00Z"
  },
  "organization": null
}
```

### Verification Checklist
- [ ] Status code: **201 Created** (still succeeds)
- [ ] `success` field: **true**
- [ ] `token` field: **Present and valid**
- [ ] `user` field: **Complete with all data**
- [ ] `organization` field: **null or omitted**
- [ ] User exists in database with correct data
- [ ] JWT includes null currentOrgId (or omitted)
- [ ] Log shows warning about org creation failure

### Post-Implementation Note
After org creation is fixed, this user could create their org later through a separate endpoint.

---

## Test Case 9: Valid Roles - All Should Succeed ✅

### Objective
Verify that all valid roles can register successfully.

### Valid Roles to Test
- `admin`
- `approver`
- `requester`
- `finance`
- `viewer`

### Request Template
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user_ROLE@example.com",
    "password": "SecurePass123",
    "name": "User Role",
    "role": "ROLE"
  }'
```

### Verification Checklist (for each role)
- [ ] Status code: **201 Created**
- [ ] `success` field: **true**
- [ ] User created with correct role
- [ ] Organization created successfully
- [ ] JWT includes org ID
- [ ] Database shows correct role

---

## Test Case 10: Multiple Registrations - Independent Organizations ✅

### Objective
Verify that multiple users get separate organizations.

### Request
```bash
# User 1
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user1@example.com",
    "password": "SecurePass123",
    "name": "User One",
    "role": "requester"
  }'

# User 2
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user2@example.com",
    "password": "SecurePass456",
    "name": "User Two",
    "role": "approver"
  }'
```

### Verification Checklist
- [ ] Both users created successfully
- [ ] Both receive 201 responses
- [ ] Two different organizations created
- [ ] User 1's `current_organization_id` points to Org 1
- [ ] User 2's `current_organization_id` points to Org 2
- [ ] Users cannot access each other's organizations
- [ ] Each user is admin of only their own org

### Database Verification
```sql
-- Verify two users exist
SELECT id, email, current_organization_id FROM users
  WHERE email IN ('user1@example.com', 'user2@example.com');

-- Verify two organizations exist
SELECT id, name, created_by FROM organizations
  WHERE name IN ('User One', 'User Two');

-- Verify membership isolation
SELECT user_id, organization_id, role FROM organization_members
  WHERE organization_id IN
    (SELECT id FROM organizations WHERE name IN ('User One', 'User Two'));
```

---

## Test Case 11: Organization Settings Auto-Created ✅

### Objective
Verify that organization_settings table is populated when org is created.

### Database Query
```sql
SELECT
  os.organization_id,
  o.name as org_name,
  os.feature_flags,
  os.created_at
FROM organization_settings os
  JOIN organizations o ON o.id = os.organization_id
WHERE o.name = 'Test User One';
```

### Expected Result
- [ ] **organization_id**: Matches organization ID
- [ ] **feature_flags**: Contains settings (JSON or serialized)
- [ ] **created_at**: Timestamp of creation
- [ ] Row exists for each organization

---

## Test Case 12: Login After Registration ✅

### Objective
Verify that user can immediately login after registration with registered credentials.

### Process
1. Register new user (Test Case 1)
2. Extract email and password
3. Call login endpoint

### Request
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "testuser1@example.com",
    "password": "SecurePass123"
  }'
```

### Expected Response (200 OK)
```json
{
  "success": true,
  "message": "Login successful",
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": "user_abc123...",
    "email": "testuser1@example.com",
    "name": "Test User One",
    "role": "requester",
    "active": true,
    "lastLogin": "2025-12-25T10:45:00Z",
    "createdAt": "2025-12-25T10:30:00Z"
  }
}
```

### Verification Checklist
- [ ] Status code: **200 OK**
- [ ] `success` field: **true**
- [ ] `token` field: **Valid JWT, includes org context**
- [ ] `user` field: **Complete**
- [ ] `lastLogin` updated to current timestamp
- [ ] JWT organization ID matches the one created during registration

---

## Testing Execution Order

**Recommended sequence:**

1. ✅ Test Case 1 (Happy Path) - Verify basic registration works
2. ✅ Test Case 2 (JWT Verification) - Verify org in token
3. ✅ Test Case 7 (Password Security) - Verify hashing
4. ✅ Test Case 11 (Organization Settings) - Verify DB state
5. ✅ Test Case 12 (Login After Registration) - Verify full flow
6. ❌ Test Case 3 (Duplicate Email) - Verify error handling
7. ❌ Test Case 4 (Weak Password) - Verify validation
8. ❌ Test Case 5 (Missing Fields) - Verify validation
9. ❌ Test Case 6 (Invalid Role) - Verify validation
10. ✅ Test Case 8 (Org Failure) - Verify graceful degradation
11. ✅ Test Case 9 (All Roles) - Verify all roles work
12. ✅ Test Case 10 (Multiple Users) - Verify isolation

---

## Test Results Summary

### Backend Testing Status
- [ ] All happy path tests passing
- [ ] All validation tests passing
- [ ] All security tests passing
- [ ] All database state verified
- [ ] All error cases handled correctly

### Issue Log
```
# Use this section during testing to log any issues found

Issue #1:
- Test Case:
- Error:
- Expected:
- Actual:
- Fix Applied:
```

---

## Automation Script (Optional)

For future automation, use this shell script:

```bash
#!/bin/bash
# File: test-registration.sh

BASE_URL="http://localhost:8080"
ENDPOINT="/api/v1/auth/register"

test_case() {
  local name=$1
  local email=$2
  local password=$3
  local name_field=$4
  local role=$5

  echo "Testing: $name"

  response=$(curl -s -X POST "$BASE_URL$ENDPOINT" \
    -H "Content-Type: application/json" \
    -d "{
      \"email\": \"$email\",
      \"password\": \"$password\",
      \"name\": \"$name_field\",
      \"role\": \"$role\"
    }")

  echo "Response: $response"
  echo "---"
}

# Run all test cases
test_case "Test 1: Valid Registration" "test1@example.com" "SecurePass123" "Test User 1" "requester"
test_case "Test 2: Valid Registration" "test2@example.com" "SecurePass456" "Test User 2" "approver"
test_case "Test 3: Duplicate Email" "test1@example.com" "Different456" "Duplicate User" "requester"
test_case "Test 4: Weak Password" "weak@example.com" "Pass1" "Weak Pass" "requester"
```

---

**Status**: Ready for backend testing phase
**Assigned To**: Development team
**Date**: 2025-12-25

