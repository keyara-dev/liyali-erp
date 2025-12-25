# Phase 2: End-to-End Testing Guide

## Overview

This guide covers complete end-to-end testing of the Scenario C implementation: Auto-Create Personal Organization on Signup. Tests verify the entire flow from user registration through dashboard access with organization context.

---

## Environment Setup

### Prerequisites
- Backend running: `http://localhost:8080`
- Frontend running: `http://localhost:3001`
- Database initialized and accessible
- Browser: Chrome/Firefox with DevTools (F12)
- Postman/curl for API testing (optional)
- Two separate browser windows or incognito tabs (for multi-user testing)

### Pre-Test Checklist
- [ ] Backend server is running
- [ ] Frontend dev server is running
- [ ] Database connection verified
- [ ] No existing test users (or clean test database)
- [ ] Browser cache cleared or using incognito mode
- [ ] DevTools console is open and ready to monitor

---

## Test Scenario 1: Complete Registration → Dashboard Flow ✅

### Objective
Verify the entire user journey from signup to dashboard access with organization context.

### Steps

#### 1.1: Navigate to Signup Page
```
1. Open browser: http://localhost:3001/signup
2. Verify page loads without errors
3. Check DevTools Console for any errors
```

**Expected Result:**
- [ ] Page loads in < 2 seconds
- [ ] Form displays with all fields
- [ ] No 404 or network errors
- [ ] Console shows no JavaScript errors

#### 1.2: Enter Registration Data
```
Email: e2e-test-1@example.com
Name: E2E Test User One
Password: TestPassword123
Confirm Password: TestPassword123
```

**Expected Result:**
- [ ] Form accepts all inputs
- [ ] Password visibility toggle works
- [ ] No validation errors for valid data
- [ ] Submit button is enabled

#### 1.3: Submit Registration Form
```
1. Click "Sign Up" button
2. Observe loading state (2-5 seconds typical)
3. Watch Network tab in DevTools
```

**Expected Result:**
- [ ] Button shows "Creating account..." text
- [ ] Form fields become disabled
- [ ] Network shows POST to `/api/v1/auth/register`
- [ ] Request includes all required fields:
  ```json
  {
    "email": "e2e-test-1@example.com",
    "name": "E2E Test User One",
    "password": "TestPassword123",
    "role": "requester"
  }
  ```
- [ ] Response status: 201 Created
- [ ] Response includes:
  - `success: true`
  - `token: <JWT>`
  - `user: {...}`
  - `organization: {...}`

#### 1.4: Session Creation Verification
```
1. After registration but before redirect, check DevTools
2. Go to: DevTools → Application → Cookies
3. Look for: AUTH_SESSION
4. Check localStorage for current-organization-id
```

**Expected Result:**
- [ ] `AUTH_SESSION` cookie created
- [ ] Cookie has `httpOnly` flag
- [ ] Cookie contains encrypted session
- [ ] localStorage has `current-organization-id` key
- [ ] localStorage value matches `organization.id` from response

#### 1.5: Automatic Redirect to Dashboard
```
1. Wait for redirect (should be automatic)
2. URL changes from /signup to /home
3. Dashboard displays
4. Verify user is logged in
```

**Expected Result:**
- [ ] Redirect happens automatically (no manual action needed)
- [ ] No intermediate screens
- [ ] URL: `http://localhost:3001/home`
- [ ] Dashboard loads successfully
- [ ] User name/email displays in header or sidebar
- [ ] No "Select Organization" modal appears
- [ ] No "Create Organization" form appears

#### 1.6: Verify Organization Context on Dashboard
```
1. Look for organization switcher component
2. Check organization name displayed
3. Verify it matches the user's name from registration
```

**Expected Result:**
- [ ] Organization switcher visible
- [ ] Shows: "E2E Test User One" (user's name)
- [ ] Organization is marked as active
- [ ] Can see organization details (click to expand)

#### 1.7: Verify User Can Perform Role-Based Actions
```
1. User role: "requester"
2. Try to create a requisition
3. Verify permission is granted
```

**Expected Result:**
- [ ] User can create requisitions (has "requester" role)
- [ ] Cannot access admin functions
- [ ] Cannot approve requisitions (not "approver")
- [ ] Appropriate UI elements shown/hidden based on role

#### 1.8: Session Persistence - Page Refresh
```
1. While on dashboard, press F5 to refresh
2. Wait for page to reload
3. Verify still logged in
```

**Expected Result:**
- [ ] Page refreshes without redirect to login
- [ ] User data still visible
- [ ] Organization context maintained
- [ ] Can still perform actions
- [ ] Session cookie still exists

#### 1.9: Logout Verification
```
1. Click logout button
2. Observe redirect to login page
3. Check session cleanup
```

**Expected Result:**
- [ ] Redirected to `/login` page
- [ ] `AUTH_SESSION` cookie deleted
- [ ] Cannot access `/home` without logging in
- [ ] Attempting to access `/home` redirects to `/login`

---

## Test Scenario 2: Multiple Users - Organization Isolation ✅

### Objective
Verify that multiple users each get their own organization and cannot access each other's data.

### Setup
- User 1: `e2e-test-user1@example.com` (Email: User One)
- User 2: `e2e-test-user2@example.com` (Email: User Two)

### Steps

#### 2.1: Register User 1
```
1. In Browser Window 1, register User 1 with:
   Email: e2e-test-user1@example.com
   Name: User One
   Password: TestPassword123
2. Verify redirect to dashboard
3. Note the organization ID from DevTools localStorage
```

**Expected Result:**
- [ ] User 1 registered successfully
- [ ] Organization created: "User One"
- [ ] Logged in and on dashboard
- [ ] localStorage shows org ID

#### 2.2: Register User 2
```
1. In Browser Window 2 (or incognito), register User 2 with:
   Email: e2e-test-user2@example.com
   Name: User Two
   Password: TestPassword123
2. Verify redirect to dashboard
3. Note the organization ID
```

**Expected Result:**
- [ ] User 2 registered successfully
- [ ] Different organization created: "User Two"
- [ ] Org IDs for User 1 and User 2 are different
- [ ] Logged in and on dashboard

#### 2.3: Verify Data Isolation
```
1. User 1 creates a requisition in their org
2. Switch to User 2 session
3. Check that User 2 cannot see User 1's requisition
```

**Expected Result:**
- [ ] User 1's requisition visible only in User 1's org
- [ ] User 2 sees empty/different requisitions
- [ ] API requests include correct X-Organization-ID header
- [ ] No data leakage between organizations

#### 2.4: Verify Independent Organization Settings
```
1. Each user's org can have independent settings
2. User 1 changes a setting in their org
3. User 2's org settings unaffected
```

**Expected Result:**
- [ ] Each organization independent
- [ ] Changes in one org don't affect others
- [ ] Proper data isolation verified

---

## Test Scenario 3: Password Security ✅

### Objective
Verify that passwords are properly hashed and validated.

### Steps

#### 3.1: Register with Valid Password
```
Email: security-test@example.com
Name: Security Test User
Password: ValidPass123
```

**Expected Result:**
- [ ] Registration succeeds
- [ ] User can login immediately

#### 3.2: Verify Password Not Stored Plaintext
```
1. Use database client or backend API
2. Query user password field: SELECT password FROM users WHERE email = 'security-test@example.com'
```

**Expected Result:**
- [ ] Password starts with `$2b$` (bcrypt signature)
- [ ] Password length ~60 characters (bcrypt hash)
- [ ] NOT the plaintext password
- [ ] Plaintext password cannot be recovered from hash

#### 3.3: Verify Login with Correct Password Works
```
1. Logout
2. Try to login with: security-test@example.com / ValidPass123
```

**Expected Result:**
- [ ] Login succeeds
- [ ] JWT token generated
- [ ] Session created
- [ ] Redirected to dashboard

#### 3.4: Verify Login with Wrong Password Fails
```
1. Try to login with: security-test@example.com / WrongPassword123
```

**Expected Result:**
- [ ] Login fails
- [ ] Error message: "Invalid email or password"
- [ ] No session created
- [ ] Remain on login page

#### 3.5: Verify Login with Plaintext Password Fails
```
1. Try various wrong passwords
2. Verify none allow access
```

**Expected Result:**
- [ ] All wrong passwords rejected
- [ ] Consistent error message
- [ ] Account remains secure

---

## Test Scenario 4: Error Handling ✅

### Objective
Verify proper error handling throughout the flow.

### Steps

#### 4.1: Duplicate Email Registration
```
1. Register with: dup-test@example.com
2. Verify success and logout
3. Try to register again with same email
4. Enter different password and name
```

**Expected Result:**
- [ ] First registration succeeds
- [ ] Second attempt fails with 409 Conflict
- [ ] Error message: "Email already registered"
- [ ] Form remains populated for retry
- [ ] User can modify form and fix (try different email)

#### 4.2: Network Error Handling
```
1. Open DevTools → Network
2. Set network throttling to "Offline"
3. Try to register
4. Observe error handling
```

**Expected Result:**
- [ ] Appropriate error message displayed
- [ ] Button becomes enabled for retry
- [ ] Form data preserved
- [ ] No broken UI state

#### 4.3: Backend Validation Error
```
1. Send invalid role (if possible through code)
2. Backend returns 400
3. Frontend handles error
```

**Expected Result:**
- [ ] Error message displays
- [ ] User can retry
- [ ] Form functional for new attempt

#### 4.4: Server Error (500)
```
1. Simulate server error (or trigger with test data)
2. Observe error handling
```

**Expected Result:**
- [ ] Graceful error message
- [ ] User can retry
- [ ] No crash or broken state

---

## Test Scenario 5: Browser Compatibility ✅

### Objective
Verify the signup and dashboard flow works across browsers.

### Browsers to Test
- [ ] Chrome (latest)
- [ ] Firefox (latest)
- [ ] Safari (if available)
- [ ] Edge (if available)

### Steps (for each browser)
```
1. Clear cache
2. Navigate to signup
3. Register new user: testuser-{browser}@example.com
4. Verify redirect to dashboard
5. Verify organization context working
6. Logout
```

**Expected Result:**
- [ ] All browsers: signup form displays correctly
- [ ] All browsers: form submission works
- [ ] All browsers: redirect to dashboard succeeds
- [ ] All browsers: organization switcher works
- [ ] All browsers: session management works
- [ ] All browsers: logout clears session

---

## Test Scenario 6: Responsive Design - Mobile/Tablet ✅

### Objective
Verify the flow works on mobile and tablet devices.

### Steps

#### 6.1: Desktop (1920x1080)
```
1. Register user
2. Verify form layout
3. Verify button clickable
4. Verify no horizontal scroll
```

**Expected Result:**
- [ ] Form displays properly
- [ ] All elements accessible
- [ ] Responsive design working

#### 6.2: Tablet (768x1024)
```
1. In DevTools, set to Tablet view (iPad)
2. Repeat signup flow
```

**Expected Result:**
- [ ] Form stacks appropriately
- [ ] Inputs properly sized
- [ ] Touch-friendly (buttons clickable)
- [ ] No layout issues

#### 6.3: Mobile (375x667)
```
1. In DevTools, set to Mobile view (iPhone SE)
2. Repeat signup flow
3. Verify password toggle buttons accessible
```

**Expected Result:**
- [ ] Form readable at mobile size
- [ ] Font sizes >= 16px (mobile friendly)
- [ ] Buttons large enough to tap (touch friendly)
- [ ] No horizontal scroll
- [ ] Form submission works on mobile

---

## Test Scenario 7: Performance ✅

### Objective
Verify the signup and dashboard load within acceptable timeframes.

### Steps

#### 7.1: Signup Page Load Time
```
1. Open DevTools → Network
2. Load: http://localhost:3001/signup
3. Check page load time
```

**Expected Result:**
- [ ] Signup page loads in < 2 seconds
- [ ] No slow requests blocking
- [ ] All resources load successfully

#### 7.2: Registration Request Time
```
1. With Network throttling: "Slow 3G"
2. Register a user
3. Monitor response time
```

**Expected Result:**
- [ ] Registration request < 5 seconds (with Slow 3G)
- [ ] Even with network latency, reasonable wait time
- [ ] User gets loading feedback

#### 7.3: Dashboard Load After Signup
```
1. Register user
2. Measure time from redirect to dashboard fully loaded
3. Verify organization switcher loads
```

**Expected Result:**
- [ ] Dashboard loads in < 3 seconds
- [ ] Organization context available immediately
- [ ] No blank screens or loading states

#### 7.4: Lighthouse Performance Score
```
1. Open DevTools → Lighthouse
2. Run audit on: /signup
3. Run audit on: /home (after login)
4. Note scores
```

**Expected Result:**
- [ ] Performance score >= 80 (mobile)
- [ ] No critical audits failing
- [ ] Load time reasonable

---

## Test Scenario 8: Security Checks ✅

### Objective
Verify security measures are in place.

### Steps

#### 8.1: HTTPS in Production Check
```
Note: On localhost, HTTP is acceptable
Production: Verify HTTPS is enforced
```

**Expected Result:**
- [ ] Production uses HTTPS only
- [ ] No mixed content warnings

#### 8.2: Session Cookie Security
```
1. Register and login
2. Check cookie settings
```

**Expected Result:**
- [ ] `AUTH_SESSION` cookie has `httpOnly` flag
- [ ] Cookie has `Secure` flag (production)
- [ ] Cookie has appropriate `SameSite` setting
- [ ] Cannot access from JavaScript (httpOnly)

#### 8.3: Password Field Security
```
1. On signup form, enter password: TestPassword123
2. Check DevTools → Elements
3. Verify password field type is "password" not "text"
```

**Expected Result:**
- [ ] Password field: `type="password"`
- [ ] Password masked when entered
- [ ] Autocomplete set to "new-password"

#### 8.4: CSRF Protection (if implemented)
```
1. Check Network tab for CSRF tokens
2. Verify registration request includes CSRF protection
```

**Expected Result:**
- [ ] CSRF tokens implemented (if backend has it)
- [ ] POST requests include token
- [ ] Token validated server-side

#### 8.5: Input Validation
```
1. Try XSS in email: <script>alert('xss')</script>@example.com
2. Try XSS in name: <img src=x onerror="alert('xss')">
3. Try special characters in password
```

**Expected Result:**
- [ ] No JavaScript execution
- [ ] Data properly escaped/sanitized
- [ ] No HTML injection
- [ ] Registration either succeeds (sanitized) or fails (validation)

---

## Test Scenario 9: API Contract Compliance ✅

### Objective
Verify the frontend and backend communicate correctly.

### Steps

#### 9.1: Request Format Verification
```
Using Postman/curl, make manual registration request:

POST /api/v1/auth/register
Content-Type: application/json

{
  "email": "api-test@example.com",
  "password": "TestPassword123",
  "name": "API Test User",
  "role": "requester"
}
```

**Expected Result:**
- [ ] Request accepted
- [ ] Response status: 201 Created
- [ ] Response format matches expectations

#### 9.2: Response Format Verification
```
Response should include:
{
  "success": true,
  "message": "Registration successful",
  "token": "<JWT>",
  "user": {
    "id": "...",
    "email": "api-test@example.com",
    "name": "API Test User",
    "role": "requester",
    "active": true,
    "createdAt": "2025-12-25T..."
  },
  "organization": {
    "id": "...",
    "name": "API Test User",
    "slug": "api-test-example.com",
    "description": "Personal Organization",
    "active": true,
    "tier": "free",
    "createdAt": "2025-12-25T..."
  }
}
```

**Expected Result:**
- [ ] Response includes all required fields
- [ ] No missing or extra fields
- [ ] Data types correct (string, boolean, etc.)
- [ ] Organization data present

#### 9.3: JWT Token Validation
```
1. Copy token from response
2. Decode at jwt.io
3. Verify claims
```

**Expected Result:**
- [ ] Token valid and decodable
- [ ] Claims include:
  - `sub`: user ID
  - `email`: user email
  - `role`: user role
  - `currentOrgId`: organization ID
- [ ] Expiration: 24 hours from issued
- [ ] Signature validates

---

## Test Scenario 10: Database State Verification ✅

### Objective
Verify that database correctly stores all data created during signup.

### Steps

#### 10.1: Verify User Record
```sql
SELECT
  id, email, name, role, active, current_organization_id, created_at
FROM users
WHERE email = 'e2e-test-1@example.com';
```

**Expected Result:**
- [ ] User record exists
- [ ] Fields match registration data:
  - email: e2e-test-1@example.com
  - name: E2E Test User One
  - role: requester
  - active: true
  - current_organization_id: NOT NULL (points to created org)

#### 10.2: Verify Organization Record
```sql
SELECT
  id, name, slug, description, active, tier, created_by, created_at
FROM organizations
WHERE created_by = '<user_id_from_above>';
```

**Expected Result:**
- [ ] Organization record exists
- [ ] Fields match:
  - name: E2E Test User One
  - slug: e2e-test-user-one (or derived from email/name)
  - description: Personal Organization
  - active: true
  - tier: free (or default tier)
  - created_by: matches user ID

#### 10.3: Verify Organization Member Record
```sql
SELECT
  id, user_id, organization_id, role, active, joined_at
FROM organization_members
WHERE user_id = '<user_id>' AND organization_id = '<org_id>';
```

**Expected Result:**
- [ ] Member record exists
- [ ] role: admin (user is admin of their org)
- [ ] active: true
- [ ] joined_at: timestamp of creation

#### 10.4: Verify Organization Settings
```sql
SELECT
  organization_id, feature_flags, created_at
FROM organization_settings
WHERE organization_id = '<org_id>';
```

**Expected Result:**
- [ ] Settings record exists
- [ ] feature_flags: contains any defaults
- [ ] created_at: matches org creation time

#### 10.5: Verify Password Hash
```sql
SELECT password FROM users WHERE id = '<user_id>';
```

**Expected Result:**
- [ ] Password starts with `$2b$` (bcrypt)
- [ ] Length ~60 characters
- [ ] NOT plaintext password

---

## Test Scenario 11: Logout & Session Cleanup ✅

### Objective
Verify that logout properly cleans up session.

### Steps

#### 11.1: Logout from Dashboard
```
1. While logged in on dashboard
2. Click logout button
3. Observe redirect
```

**Expected Result:**
- [ ] Redirected to login page
- [ ] No error messages
- [ ] Smooth transition

#### 11.2: Verify Session Cleanup
```
1. Check DevTools → Application → Cookies
2. Verify AUTH_SESSION cookie deleted
3. Check localStorage
```

**Expected Result:**
- [ ] `AUTH_SESSION` cookie no longer exists
- [ ] localStorage cleared (or appropriate keys removed)
- [ ] No sensitive data remains

#### 11.3: Verify Access Denied to Protected Routes
```
1. After logout, try to access: http://localhost:3001/home
2. Observe redirect
```

**Expected Result:**
- [ ] Automatically redirected to login
- [ ] Cannot access protected pages
- [ ] Session required for access

---

## Test Scenario 12: Cross-Tab Session Sync (Optional) ✅

### Objective
Verify that logout in one tab affects other tabs (if implemented).

### Steps

#### 12.1: Open Multiple Tabs
```
1. Open Tab 1: Register and login
2. Open Tab 2: Also access the app (same browser)
3. Both tabs show logged in
```

**Expected Result:**
- [ ] Both tabs show user is logged in
- [ ] Same session across tabs (same cookie)

#### 12.2: Logout in One Tab
```
1. In Tab 1, click logout
2. Switch to Tab 2
3. Try to access protected page
```

**Expected Result:**
- [ ] Tab 2 redirects to login (or shows logout prompt)
- [ ] Tab 2 detects logout in Tab 1
- [ ] Proper session sync across tabs

---

## Quick Test Checklist (Fast Path)

For quick validation, run this minimal test:

```
[ ] 1. Navigate to /signup
[ ] 2. Register with: quicktest@example.com / QuickTest123 / Quick Test
[ ] 3. Verify form validation works (try weak password first)
[ ] 4. Submit registration
[ ] 5. Verify automatic redirect to /home
[ ] 6. Verify logged in (user name visible)
[ ] 7. Verify organization switcher shows "Quick Test" org
[ ] 8. Verify can click organization (doesn't error)
[ ] 9. Refresh page - still logged in? [ ]
[ ] 10. Logout and verify redirect to login
```

**Time**: ~5-10 minutes

---

## Full Test Execution (Comprehensive Path)

Run all test scenarios:

| Scenario | Time | Status |
|----------|------|--------|
| Test 1: Complete Flow | 15 min | [ ] PASS |
| Test 2: Multi-User Isolation | 15 min | [ ] PASS |
| Test 3: Password Security | 10 min | [ ] PASS |
| Test 4: Error Handling | 10 min | [ ] PASS |
| Test 5: Browser Compatibility | 20 min | [ ] PASS |
| Test 6: Responsive Design | 15 min | [ ] PASS |
| Test 7: Performance | 15 min | [ ] PASS |
| Test 8: Security Checks | 15 min | [ ] PASS |
| Test 9: API Contract | 10 min | [ ] PASS |
| Test 10: Database State | 10 min | [ ] PASS |
| Test 11: Session Cleanup | 10 min | [ ] PASS |
| Test 12: Cross-Tab Sync | 10 min | [ ] PASS |
| **TOTAL** | **2.5 hours** | |

---

## Test Results Template

### Test Execution Report

**Date:** 2025-12-25
**Tester:** [Name]
**Environment:** Local Dev (Backend 8080, Frontend 3001)
**Database:** [PostgreSQL/MySQL]
**Browsers Tested:** [Chrome, Firefox, etc.]

### Results Summary

| Scenario | Status | Notes |
|----------|--------|-------|
| Complete Flow | PASS / FAIL | |
| Multi-User Isolation | PASS / FAIL | |
| Password Security | PASS / FAIL | |
| Error Handling | PASS / FAIL | |
| Browser Compatibility | PASS / FAIL | |
| Responsive Design | PASS / FAIL | |
| Performance | PASS / FAIL | |
| Security | PASS / FAIL | |
| API Contract | PASS / FAIL | |
| Database State | PASS / FAIL | |
| Session Cleanup | PASS / FAIL | |
| Cross-Tab Sync | PASS / FAIL | |

### Issues Found

```
Issue #1: [Description]
  Severity: CRITICAL / HIGH / MEDIUM / LOW
  Component: [Frontend/Backend/Database]
  Steps to Reproduce: [Steps]
  Expected: [What should happen]
  Actual: [What actually happened]
  Resolution: [How it was fixed or action item]
  Status: OPEN / CLOSED

Issue #2: ...
```

### Sign-Off

- [ ] All tests passing
- [ ] No critical issues
- [ ] Ready for deployment to staging
- [ ] Approved by: [Name]

---

## Troubleshooting During Testing

### "Invalid credentials" on login after signup
- **Cause**: Password not hashing/verifying correctly
- **Fix**: Verify backend password hashing is enabled in auth.go (lines 65-70)
- **Check**: `utils.VerifyPassword()` function called correctly

### Organization not showing in switcher
- **Cause**: Organization not returned in signup response
- **Fix**: Verify `authResponse.Organization` is set (auth.go lines 249-259)
- **Check**: Backend returns 201 with organization field

### Redirect to /home not happening
- **Cause**: Frontend not redirecting after successful registration
- **Fix**: Check signup.tsx router.push('/home') is called
- **Check**: No errors in console blocking redirect

### Session not persisting after refresh
- **Cause**: Session cookie not being set or cleared
- **Fix**: Verify `createAuthSession()` called with organization_id
- **Check**: Session type includes organization_id field

### CORS errors on API calls
- **Cause**: Frontend and backend CORS mismatch
- **Fix**: Verify backend CORS configured for http://localhost:3001
- **Check**: Response headers include Access-Control-Allow-Origin

---

**Status**: Ready for end-to-end testing
**Created**: 2025-12-25
**Next Phase**: Phase 2C.3 - Documentation Update

