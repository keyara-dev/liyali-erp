# End-to-End Testing Plan - Liyali Gateway

**Date**: 2025-12-26
**Status**: Ready to Execute
**Scope**: Complete MVP feature testing
**Estimated Duration**: 2-3 hours

---

## 🎯 Testing Objectives

1. **Verify all core workflows work end-to-end**
2. **Validate multi-tenancy and data isolation**
3. **Test RBAC and permission enforcement**
4. **Confirm all CRUD operations persist to database**
5. **Verify API integration with frontend**
6. **Test approval workflows with multiple stages**
7. **Validate user experience flows**
8. **Check error handling and edge cases**

---

## 🏗️ Test Environment Setup

### Prerequisites
- Docker installed (or PostgreSQL + Go + Node.js locally)
- Git cloned to `d:\dev\next-apps\liyali-gateway`
- Network connectivity for API calls

### Option 1: Docker Setup (Recommended)

```bash
cd d:\dev\next-apps\liyali-gateway
docker-compose up -d

# Wait for services to be ready (60 seconds)
sleep 60

# Check services are running
docker-compose ps
```

### Option 2: Local Setup

```bash
# Terminal 1: Start PostgreSQL
# (Ensure PostgreSQL is running on localhost:5432)

# Terminal 2: Start Backend
cd backend
go run main.go

# Terminal 3: Start Frontend
cd frontend
npm run dev
```

---

## 📋 Test Cases

### Phase 1: Authentication & Authorization (30 min)

#### TC-1.1: User Registration
**Steps**:
1. Open http://localhost:3000
2. Click "Sign Up"
3. Enter email: `testuser1@example.com`
4. Enter password: `TestPass123!`
5. Click "Create Account"

**Expected**:
- ✅ Account created successfully
- ✅ Redirected to login page
- ✅ Can now login with credentials
- ✅ Personal organization auto-created in backend

**Verification**:
```bash
curl -X GET http://localhost:8080/api/v1/organizations \
  -H "Authorization: Bearer <token>"
```

---

#### TC-1.2: User Login
**Steps**:
1. Go to http://localhost:3000/login
2. Enter email: `testuser1@example.com`
3. Enter password: `TestPass123!`
4. Click "Login"

**Expected**:
- ✅ Login successful
- ✅ JWT token returned
- ✅ Redirected to dashboard
- ✅ User context shows correct organization

**Verification**:
- Token stored in localStorage
- Dashboard loads with user's organizations

---

#### TC-1.3: Role-Based Access Control
**Steps**:
1. Login as Admin user
2. Go to /admin/users/roles
3. View list of 5 system roles
4. Try to delete "Admin" role

**Expected**:
- ✅ See system roles (Admin, Approver, Requester, Finance, Viewer)
- ✅ Cannot delete system roles (disabled or error)
- ✅ Can create custom roles
- ✅ Can assign permissions to roles

---

#### TC-1.4: Permission Enforcement
**Steps**:
1. Login as "Viewer" role
2. Try to create a new requisition
3. Check if blocked at UI level
4. Try to access /admin/workflows directly

**Expected**:
- ✅ Create button disabled or unavailable
- ✅ Cannot create requisition (backend rejects)
- ✅ Admin pages show permission guard
- ✅ Error message: "Insufficient permissions"

---

### Phase 2: Multi-Tenancy (30 min)

#### TC-2.1: Personal Organization Auto-Creation
**Steps**:
1. Register new user: `tenant1@example.com`
2. After registration, check organizations

**Expected**:
- ✅ Personal org created automatically
- ✅ Org name: `<email> Personal Org`
- ✅ User is owner of personal org
- ✅ Can switch to personal org

---

#### TC-2.2: Multiple Organizations
**Steps**:
1. Login as `testuser1@example.com`
2. Go to organization switcher
3. Create new organization: "Company A"
4. Switch to "Company A"
5. Create requisition in "Company A"
6. Switch back to personal org
7. Verify requisition not visible in personal org

**Expected**:
- ✅ Can create multiple orgs
- ✅ Organization switcher works
- ✅ Data isolated per org
- ✅ Requisition only visible in "Company A"

---

#### TC-2.3: Organization Member Management
**Steps**:
1. In "Company A" org
2. Go to Settings → Members
3. Click "Invite Member"
4. Add `testuser2@example.com` as "Approver"
5. Logout and login as `testuser2@example.com`
6. Verify can see "Company A" in org switcher

**Expected**:
- ✅ Can invite members
- ✅ Members have correct role
- ✅ New member can access org
- ✅ Org appears in member's list

---

### Phase 3: Core Workflow - Requisitions (45 min)

#### TC-3.1: Create Requisition
**Steps**:
1. Login as Requester role
2. Go to Requisitions
3. Click "New Requisition"
4. Fill in:
   - Department: "IT"
   - Budget Limit: "5000"
   - Items:
     - Item 1: "Laptops" x2 @ $1000 each
     - Item 2: "Monitors" x2 @ $500 each
5. Click "Submit"

**Expected**:
- ✅ Requisition created
- ✅ Status: "Draft"
- ✅ Can edit draft requisitions
- ✅ Data persists in database

**Verification**:
```bash
curl -X GET http://localhost:8080/api/v1/requisitions \
  -H "Authorization: Bearer <token>" | jq
```

---

#### TC-3.2: Submit for Approval
**Steps**:
1. In the created requisition
2. Click "Submit for Approval"
3. Confirm submission

**Expected**:
- ✅ Status changes to "Pending Approval"
- ✅ Moved to Approver's queue
- ✅ Requester cannot edit anymore
- ✅ Approval history shows "Submitted"

---

#### TC-3.3: Multi-Stage Approval
**Steps**:
1. Login as first Approver
2. Go to Approvals → Requisitions
3. See the submitted requisition
4. Click "Approve"
5. Logout and login as second Approver
6. See the requisition (now at stage 2)
7. Click "Approve"

**Expected**:
- ✅ First approver can see and approve
- ✅ Moves to next stage
- ✅ Second approver can see and approve
- ✅ After final approval: Status = "Approved"

---

#### TC-3.4: Rejection Workflow
**Steps**:
1. Create new requisition
2. Submit for approval
3. Login as Approver
4. Click "Reject"
5. Enter reason: "Budget exceeded"
6. Submit rejection

**Expected**:
- ✅ Status changes to "Rejected"
- ✅ Rejection reason visible
- ✅ Requester can edit and resubmit
- ✅ Approval history shows rejection with reason

---

#### TC-3.5: Reassignment
**Steps**:
1. Create and submit requisition
2. Login as Approver
3. Click "Reassign"
4. Select different approver
5. Logout and login as new approver
6. Verify requisition appears in their queue

**Expected**:
- ✅ Can reassign to different person
- ✅ New approver sees it in their queue
- ✅ Original approver no longer sees it
- ✅ History shows reassignment

---

### Phase 4: Budgets Workflow (30 min)

#### TC-4.1: Create & Approve Budget
**Steps**:
1. Login as Finance role
2. Go to Budgets
3. Click "New Budget"
4. Enter:
   - Name: "Q1 2025 Budget"
   - Amount: "100,000"
   - Department: "Operations"
5. Submit for approval
6. Login as Budget Approver
7. Approve the budget

**Expected**:
- ✅ Budget created with status "Draft"
- ✅ Can submit for approval
- ✅ Approver can approve
- ✅ Status = "Approved"
- ✅ Requisitions can reference this budget

---

#### TC-4.2: Budget Constraints
**Steps**:
1. Create budget with limit 10,000
2. Create requisition requesting 8,000
3. Approve requisition
4. Create another requisition requesting 3,000
5. Try to approve

**Expected**:
- ✅ First requisition approved (8,000 <= 10,000)
- ✅ Second requisition blocked (8,000 + 3,000 > 10,000)
- ✅ Error message: "Budget exceeded"
- ✅ Cannot proceed with approval

---

### Phase 5: Purchase Orders (30 min)

#### TC-5.1: Create PO from Requisition
**Steps**:
1. Approve a requisition
2. Click "Create Purchase Order"
3. System pre-fills:
   - Vendor: (select from dropdown)
   - Items: (from requisition)
   - Total: (auto-calculated)
4. Click "Submit"

**Expected**:
- ✅ PO created from requisition
- ✅ Items pre-populated
- ✅ Linked to original requisition
- ✅ Status: "Draft"

---

#### TC-5.2: PO Approval Workflow
**Steps**:
1. Submit PO for approval
2. Approver reviews and approves
3. Check PO status

**Expected**:
- ✅ Goes through approval stages
- ✅ Status changes to "Approved"
- ✅ Can now be used for GRN

---

### Phase 6: GRN (Goods Received Notes) (30 min)

#### TC-6.1: Create GRN from PO
**Steps**:
1. Approve a PO
2. Go to GRN
3. Click "Create GRN"
4. Select PO
5. Confirm receipt of items
6. Submit

**Expected**:
- ✅ GRN created
- ✅ Items from PO shown
- ✅ Can confirm receipt
- ✅ Status: "Received"

---

#### TC-6.2: GRN Rejection
**Steps**:
1. Create GRN from PO
2. Click "Reject Items"
3. Mark specific items as rejected
4. Submit rejection

**Expected**:
- ✅ Can reject specific items
- ✅ Status: "Partially Rejected"
- ✅ Rejected items tracked
- ✅ Can create replacement GRN

---

### Phase 7: Data Integrity & Isolation (30 min)

#### TC-7.1: Cross-Organization Isolation
**Steps**:
1. User A creates requisition in Org A
2. User B (in Org B) tries to access Org A's requisition via:
   - UI (should not see it)
   - API call with Org A's ID

**Expected**:
- ✅ User B cannot see Org A's data
- ✅ API returns 403 Forbidden or empty list
- ✅ Database queries scoped to org
- ✅ No information leakage

---

#### TC-7.2: Data Persistence
**Steps**:
1. Create requisition in Org A
2. Close browser/app
3. Reopen and login
4. Navigate to requisition

**Expected**:
- ✅ Requisition still exists
- ✅ All data intact
- ✅ No loss of information
- ✅ Database persists correctly

---

### Phase 8: Reporting & Analytics (20 min)

#### TC-8.1: Approval Reports
**Steps**:
1. Go to Admin → Reports → Approval Reports
2. View approval metrics

**Expected**:
- ✅ Shows total requisitions approved
- ✅ Shows pending approvals
- ✅ Shows rejected count
- ✅ Shows average approval time

---

#### TC-8.2: System Statistics
**Steps**:
1. Go to Admin → Reports → System Statistics
2. View system metrics

**Expected**:
- ✅ Total documents count
- ✅ By status breakdown
- ✅ Approval success rate
- ✅ User activity metrics

---

#### TC-8.3: Activity Logs
**Steps**:
1. Go to Admin → Logs → Activity Logs
2. Filter by action: "approved"
3. Search for specific user

**Expected**:
- ✅ Shows all activities
- ✅ Filters work correctly
- ✅ Search finds activities
- ✅ Timestamps accurate

---

### Phase 9: Error Handling & Edge Cases (20 min)

#### TC-9.1: Invalid Input
**Steps**:
1. Try to create requisition with:
   - Empty department
   - Negative amount
   - Future date

**Expected**:
- ✅ Form validation shows errors
- ✅ Cannot submit invalid data
- ✅ Error messages clear
- ✅ No partial records created

---

#### TC-9.2: Concurrent Operations
**Steps**:
1. Open same requisition in 2 browsers
2. Edit in both
3. Submit from browser 1
4. Try to submit from browser 2

**Expected**:
- ✅ Browser 2 detects conflict
- ✅ Shows "Document was modified"
- ✅ Reload before saving
- ✅ No data loss or corruption

---

#### TC-9.3: Permission Boundary Testing
**Steps**:
1. User with "Viewer" role
2. Try all admin endpoints with API calls
3. Check each returns proper error

**Expected**:
- ✅ All protected endpoints return 403
- ✅ Clear error messages
- ✅ No sensitive data in errors
- ✅ Logged for security audit

---

### Phase 10: UI/UX Experience (20 min)

#### TC-10.1: Navigation & Layout
**Steps**:
1. Login and navigate through all main pages
2. Test breadcrumbs
3. Test sidebar navigation
4. Test mobile responsiveness

**Expected**:
- ✅ All pages load quickly
- ✅ Navigation works
- ✅ Breadcrumbs accurate
- ✅ Mobile layout responsive

---

#### TC-10.2: Forms & Validation
**Steps**:
1. Fill out all forms in the app
2. Test required field validation
3. Test format validation (email, date, etc.)
4. Test long text handling

**Expected**:
- ✅ Forms validate correctly
- ✅ Error messages helpful
- ✅ No truncation of data
- ✅ Forms recover state on error

---

#### TC-10.3: Loading & Error States
**Steps**:
1. Create action that takes time
2. Watch loading indicator
3. Simulate network error
4. Check error message

**Expected**:
- ✅ Loading spinners appear
- ✅ Errors shown clearly
- ✅ Retry options available
- ✅ UI remains usable

---

## 🔍 Test Execution Log Template

```
Test Case: [TC-X.X]
Date: [Date]
Tester: [Name]
Browser/Device: [Chrome/Firefox/Mobile]

Status: ✅ PASS / ❌ FAIL

Expected: [What should happen]
Actual: [What actually happened]

Issues Found:
- [If any issues]

Notes:
- [Any observations]

Time: [Duration]
```

---

## 📊 Test Results Summary

Use this table to track results:

| Phase | Test Case | Status | Duration | Notes |
|-------|-----------|--------|----------|-------|
| 1 | TC-1.1 | ⏳ Pending | - | - |
| 1 | TC-1.2 | ⏳ Pending | - | - |
| 1 | TC-1.3 | ⏳ Pending | - | - |
| 2 | TC-2.1 | ⏳ Pending | - | - |
| 3 | TC-3.1 | ⏳ Pending | - | - |
| ... | ... | ... | ... | ... |

---

## 🛠️ Tools & Commands

### API Testing with cURL

```bash
# Login and get token
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password"
  }' | jq .token

# Export token
TOKEN="<token_from_above>"

# Create requisition
curl -X POST http://localhost:8080/api/v1/requisitions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "department": "IT",
    "budget_limit": 5000,
    "items": [
      {
        "name": "Laptops",
        "quantity": 2,
        "unit_price": 1000
      }
    ]
  }'

# Get all requisitions
curl -X GET http://localhost:8080/api/v1/requisitions \
  -H "Authorization: Bearer $TOKEN" | jq

# Approve requisition
curl -X PUT http://localhost:8080/api/v1/requisitions/{id}/approve \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "stage": 1,
    "notes": "Approved"
  }'
```

### Database Queries

```sql
-- Check organizations
SELECT id, name, created_by FROM organizations;

-- Check users
SELECT id, email, organization_id FROM users;

-- Check requisitions
SELECT id, status, organization_id FROM requisitions;

-- Check approval flow
SELECT * FROM approval_flows WHERE requisition_id = '<id>';

-- Check permissions
SELECT role_id, permission_name FROM permission_assignments;
```

### Logs & Debugging

```bash
# View backend logs
docker-compose logs backend -f

# View database logs
docker-compose logs db -f

# View frontend console
# Open browser DevTools (F12) → Console tab
```

---

## 🚨 Critical Test Cases (Must Pass)

For MVP launch, these must all pass:

1. ✅ User can register and login
2. ✅ Personal org auto-created
3. ✅ Can create requisition
4. ✅ Can submit for approval
5. ✅ Approver can approve/reject
6. ✅ Status updates correctly
7. ✅ Data persists to DB
8. ✅ Org isolation works
9. ✅ Permissions enforced
10. ✅ No 500 errors in happy path

---

## 📝 Defect Logging

If you find issues, document them:

```
Defect #: [Number]
Title: [Short description]
Severity: Critical / High / Medium / Low
Component: Backend / Frontend / API / Database

Steps to Reproduce:
1. [Step 1]
2. [Step 2]
3. [Step 3]

Expected: [What should happen]
Actual: [What actually happened]

Environment: [OS, Browser, etc.]
Attachments: [Screenshots/logs if applicable]
```

---

## ✅ Sign-Off Criteria

E2E testing is complete when:
- [ ] All 50+ test cases executed
- [ ] At least 45/50 passed (90%+)
- [ ] All critical test cases passed
- [ ] No critical defects remaining
- [ ] No 500 errors in logs
- [ ] Database integrity verified
- [ ] Multi-tenancy verified
- [ ] RBAC verified
- [ ] All workflows tested
- [ ] Documentation updated with findings

---

## 📚 Related Documents

- **TESTING-GUIDE.md** - General testing procedures
- **API-DOCUMENTATION.md** - API endpoint details
- **IMPLEMENTATION-CHECKLIST.md** - Feature status
- **MVP-READINESS-SUMMARY.md** - MVP scope

---

**Ready to Test**: ✅ YES

**Estimated Duration**: 2-3 hours

**Next Step**: Execute test cases and document results

---

**Created By**: Claude Code
**Date**: 2025-12-26
