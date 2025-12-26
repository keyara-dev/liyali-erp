# E2E Test Execution Guide - Liyali Gateway

**Date**: 2025-12-26
**Status**: Ready for Execution
**Duration**: 2-3 hours
**Team**: QA/Testing Team

---

## 🚀 Quick Start (5 minutes)

### Option A: Docker Compose (Easiest)

```bash
cd d:\dev\next-apps\liyali-gateway

# Start services
docker-compose up -d

# Wait 60 seconds for services to be ready
sleep 60

# Verify services running
docker-compose ps

# Access:
# - Frontend: http://localhost:3000
# - Backend API: http://localhost:8080
# - API Docs: http://localhost:8080/docs
```

### Option B: Local Setup

```bash
# Terminal 1: Backend
cd backend
go run main.go

# Terminal 2: Frontend
cd frontend
npm run dev

# Terminal 3: Database (if not running)
# Make sure PostgreSQL is running on localhost:5432

# Access:
# - Frontend: http://localhost:3000
# - Backend: http://localhost:8080
```

---

## 📋 Test Execution Steps

### Step 1: Pre-Test Verification (10 minutes)

#### 1.1 Verify Services Are Running

```bash
# Check backend health
curl http://localhost:8080/health

# Expected response:
# {"status":"OK"} or similar

# Check frontend
curl http://localhost:3000

# Should return HTML (frontend loaded)
```

#### 1.2 Check Database

```bash
# Check if backend can connect to database
curl http://localhost:8080/api/v1/health

# Should return success response indicating DB is connected
```

#### 1.3 Browser Setup

- [ ] Open Chrome/Firefox in Private/Incognito mode
- [ ] Go to http://localhost:3000
- [ ] Open DevTools (F12)
- [ ] Go to Console tab
- [ ] Keep it open to catch any errors

---

### Step 2: Authentication Tests (15 minutes)

#### TC-1.1: User Registration

**Scenario**: New user signs up

**Steps**:
1. On login page, click "Sign Up"
2. Fill in form:
   - Email: `testuser1-<timestamp>@example.com`
   - Password: `TestPass123!`
   - Confirm Password: `TestPass123!`
3. Click "Create Account"
4. Should redirect to login page

**Verification**:
- [ ] Account created successfully (no errors in console)
- [ ] Can login with new credentials
- [ ] Personal organization created

**Documentation**:
```
Date: _______
Tester: _______
Status: PASS / FAIL
Issues: _______________________________
Time: _____ minutes
```

---

#### TC-1.2: User Login

**Scenario**: User logs in

**Steps**:
1. Enter email from TC-1.1
2. Enter password: `TestPass123!`
3. Click "Login"

**Verification**:
- [ ] Login successful
- [ ] Redirected to dashboard
- [ ] User profile shows correct name/email
- [ ] Organization switcher shows personal org
- [ ] No errors in console

**Documentation**:
```
Status: PASS / FAIL
Issues: _______________________________
Time: _____ minutes
```

---

#### TC-1.3: Role-Based Access Control

**Scenario**: Different roles have different permissions

**Steps**:
1. Login as admin (use default admin account if available)
2. Navigate to `/admin/users/roles`
3. View available roles
4. Try to delete "Admin" role (should be prevented)
5. Logout

**Verification**:
- [ ] Can see 5 system roles: Admin, Approver, Requester, Finance, Viewer
- [ ] System roles cannot be deleted
- [ ] Can create custom roles (if feature available)
- [ ] Permissions shown correctly

**Documentation**:
```
Status: PASS / FAIL
Issues: _______________________________
Time: _____ minutes
```

---

### Step 3: Multi-Tenancy Tests (20 minutes)

#### TC-2.1: Personal Organization Auto-Creation

**Scenario**: New user has personal org created automatically

**Prerequisite**: Completed TC-1.1 (registered user)

**Steps**:
1. Login with account from TC-1.1
2. Click organization switcher (top of page)

**Verification**:
- [ ] Personal organization is listed
- [ ] Name follows pattern: `<email> Personal Org`
- [ ] User is owner
- [ ] Can select/switch to personal org

**Documentation**:
```
Status: PASS / FAIL
Issues: _______________________________
Time: _____ minutes
```

---

#### TC-2.2: Create Multiple Organizations

**Scenario**: User can create multiple organizations

**Steps**:
1. In organization switcher, click "Create Organization"
2. Enter name: `Test Company A`
3. Click "Create"
4. Should see new org in switcher

**Verification**:
- [ ] Organization created successfully
- [ ] Appears in switcher dropdown
- [ ] Can switch between orgs
- [ ] UI shows current org name

**Documentation**:
```
Status: PASS / FAIL
Issues: _______________________________
Time: _____ minutes
```

---

#### TC-2.3: Data Isolation Between Orgs

**Scenario**: Requisitions in one org are not visible in another

**Steps**:
1. Switch to "Test Company A" org
2. Create a requisition (see TC-3.1)
3. Switch to personal org
4. Check requisitions list

**Verification**:
- [ ] Requisition created in "Test Company A" is NOT visible in personal org
- [ ] Each org has separate data
- [ ] Cross-org access is blocked

**Documentation**:
```
Status: PASS / FAIL
Issues: _______________________________
Time: _____ minutes
```

---

### Step 4: Requisition Workflow Tests (45 minutes)

#### TC-3.1: Create Requisition

**Scenario**: Requester creates new requisition

**Steps**:
1. Login as user with Requester role
2. Navigate to Requisitions module
3. Click "New Requisition"
4. Fill in form:
   - Department: `Information Technology`
   - Budget Limit: `5000`
   - Add Items:
     - Item 1: Laptops, Qty: 2, Unit Price: 1000
     - Item 2: Monitors, Qty: 2, Unit Price: 500
5. Click "Save as Draft"

**Verification**:
- [ ] Requisition created with status "Draft"
- [ ] Can edit draft requisition
- [ ] Total amount calculated: 3000 (2*1000 + 2*500)
- [ ] Data saved to database (persists after logout/login)

**API Verification**:
```bash
curl -X GET http://localhost:8080/api/v1/requisitions \
  -H "Authorization: Bearer <token>" | jq '.data[] | {id, status, total}'
```

**Documentation**:
```
Status: PASS / FAIL
Requisition ID: _______
Issues: _______________________________
Time: _____ minutes
```

---

#### TC-3.2: Submit for Approval

**Scenario**: Requester submits draft for approval

**Prerequisites**: Completed TC-3.1

**Steps**:
1. Open the draft requisition from TC-3.1
2. Click "Submit for Approval"
3. Confirm submission

**Verification**:
- [ ] Status changed to "Pending Approval"
- [ ] Cannot edit anymore (form disabled)
- [ ] Moved to approver's queue
- [ ] Approval history shows "Submitted"

**Documentation**:
```
Status: PASS / FAIL
Issues: _______________________________
Time: _____ minutes
```

---

#### TC-3.3: Approver Reviews and Approves

**Scenario**: Approver reviews and approves requisition

**Prerequisites**: Completed TC-3.2

**Steps**:
1. Logout and login as user with Approver role
2. Navigate to Approvals section
3. Find requisition from TC-3.2
4. Click "Review"
5. Check details
6. Click "Approve"
7. Add note: "Approved - Budget looks good"
8. Confirm

**Verification**:
- [ ] Requisition appears in approver's queue
- [ ] Can view all details and items
- [ ] Can add approval notes
- [ ] Status changes to "Approved"
- [ ] Approval history updated

**Documentation**:
```
Status: PASS / FAIL
Issues: _______________________________
Time: _____ minutes
```

---

#### TC-3.4: Rejection Workflow

**Scenario**: Approver rejects requisition

**Steps**:
1. Create new requisition (repeat TC-3.1-3.2 with different data)
2. Login as approver
3. Find requisition in approval queue
4. Click "Reject"
5. Enter reason: "Budget exceeded department limit"
6. Confirm rejection

**Verification**:
- [ ] Status changes to "Rejected"
- [ ] Rejection reason visible in history
- [ ] Original requester can now edit and resubmit
- [ ] Rejection notification visible

**Documentation**:
```
Status: PASS / FAIL
Issues: _______________________________
Time: _____ minutes
```

---

#### TC-3.5: Reassignment

**Scenario**: Approver reassigns to different approver

**Steps**:
1. Create and submit new requisition
2. Login as first approver
3. Find requisition in queue
4. Click "Reassign"
5. Select different approver from dropdown
6. Click "Reassign"
7. Logout and login as second approver
8. Check approval queue

**Verification**:
- [ ] First approver can reassign
- [ ] Second approver sees requisition
- [ ] First approver no longer sees it
- [ ] Reassignment logged in history

**Documentation**:
```
Status: PASS / FAIL
Issues: _______________________________
Time: _____ minutes
```

---

### Step 5: Budget Workflow Tests (20 minutes)

#### TC-4.1: Create and Approve Budget

**Scenario**: Finance user creates budget and approver approves it

**Steps**:
1. Login as Finance role
2. Navigate to Budgets
3. Click "New Budget"
4. Fill in:
   - Name: `Q1 2025 Operations Budget`
   - Department: `Operations`
   - Amount: `100,000`
5. Submit for approval
6. Login as Budget Approver
7. Find budget in approval queue
8. Click "Approve"

**Verification**:
- [ ] Budget created with status "Draft"
- [ ] Can submit for approval
- [ ] Approver can see in queue
- [ ] Status changes to "Approved"

**Documentation**:
```
Status: PASS / FAIL
Budget ID: _______
Issues: _______________________________
Time: _____ minutes
```

---

### Step 6: Purchase Order Tests (20 minutes)

#### TC-5.1: Create PO from Approved Requisition

**Scenario**: Create purchase order from approved requisition

**Prerequisites**: Have an approved requisition (TC-3.3)

**Steps**:
1. Navigate to Requisitions
2. Open an approved requisition
3. Click "Create Purchase Order"
4. System shows pre-filled items
5. Select vendor from dropdown
6. Click "Submit for Approval"

**Verification**:
- [ ] PO created from requisition data
- [ ] Items pre-populated correctly
- [ ] Total amount matches requisition
- [ ] Status: "Draft" initially
- [ ] Can submit for approval

**Documentation**:
```
Status: PASS / FAIL
PO ID: _______
Issues: _______________________________
Time: _____ minutes
```

---

### Step 7: GRN Tests (20 minutes)

#### TC-6.1: Create and Confirm GRN

**Scenario**: Warehouse staff receives goods and creates GRN

**Prerequisites**: Have approved PO from TC-5.1

**Steps**:
1. Navigate to GRN section
2. Click "Create GRN"
3. Select PO from dropdown
4. Confirm receipt of items:
   - [ ] Item 1: Laptops - Qty received: 2
   - [ ] Item 2: Monitors - Qty received: 2
5. Click "Submit"

**Verification**:
- [ ] GRN created with status "Received"
- [ ] Items from PO shown
- [ ] Can confirm quantities
- [ ] Data persists

**Documentation**:
```
Status: PASS / FAIL
GRN ID: _______
Issues: _______________________________
Time: _____ minutes
```

---

### Step 8: Data Integrity & Isolation (30 minutes)

#### TC-7.1: Cross-Organization Access Prevention

**Scenario**: User cannot access data from other organizations

**Prerequisites**: Two users in different orgs

**Steps**:
1. Create requisition in "Test Company A"
2. Note the requisition ID
3. Logout and login as different user
4. Try to access that requisition via URL

**Verification**:
- [ ] Cannot see requisition in UI
- [ ] API call returns 403 Forbidden or empty
- [ ] No error leak of information
- [ ] Only org's own data visible

**API Test**:
```bash
# Get token for user in different org
TOKEN2="..."

# Try to access organization 1's data
curl -X GET http://localhost:8080/api/v1/requisitions/requisition1 \
  -H "Authorization: Bearer $TOKEN2"

# Should return 403 Forbidden or "not found"
```

**Documentation**:
```
Status: PASS / FAIL
Issues: _______________________________
Time: _____ minutes
```

---

#### TC-7.2: Data Persistence

**Scenario**: Data persists across sessions

**Prerequisites**: Created various documents (requisitions, budgets, etc.)

**Steps**:
1. Create requisition with specific details
2. Logout completely
3. Close browser
4. Open browser
5. Login again
6. Navigate to requisitions
7. Find the requisition

**Verification**:
- [ ] Requisition still exists
- [ ] All data intact (items, amounts, etc.)
- [ ] Status unchanged
- [ ] No data loss

**Documentation**:
```
Status: PASS / FAIL
Issues: _______________________________
Time: _____ minutes
```

---

### Step 9: Error Handling Tests (20 minutes)

#### TC-8.1: Input Validation

**Scenario**: Invalid inputs are rejected with clear errors

**Steps**:
1. Try to create requisition with empty department
2. Try to submit with negative budget amount
3. Try to add item with zero quantity
4. Try to submit form with missing required fields

**Verification**:
- [ ] Form shows validation errors
- [ ] Cannot submit invalid data
- [ ] Error messages are clear
- [ ] No partial records created
- [ ] No errors in console

**Documentation**:
```
Status: PASS / FAIL
Issues: _______________________________
Time: _____ minutes
```

---

#### TC-8.2: Permission Enforcement

**Scenario**: Users without permissions get proper error messages

**Steps**:
1. Login as "Viewer" role
2. Try to create new requisition (button should be disabled)
3. Try API call directly:

```bash
TOKEN="viewer_token"
curl -X POST http://localhost:8080/api/v1/requisitions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "department": "IT",
    "budget_limit": 5000
  }'
```

**Verification**:
- [ ] Create button disabled in UI
- [ ] API returns 403 Forbidden
- [ ] Error message: "Insufficient permissions"
- [ ] No data created

**Documentation**:
```
Status: PASS / FAIL
Issues: _______________________________
Time: _____ minutes
```

---

### Step 10: Reporting & Analytics (15 minutes)

#### TC-9.1: Approval Reports

**Scenario**: Admin can view approval statistics

**Steps**:
1. Login as Admin
2. Navigate to Admin → Reports → Approval Reports
3. Check metrics shown:
   - Total requisitions processed
   - Pending approvals
   - Rejected count
   - Average approval time

**Verification**:
- [ ] Reports load without error
- [ ] Numbers are accurate (match created documents)
- [ ] Charts render correctly
- [ ] Filters work (if available)

**Documentation**:
```
Status: PASS / FAIL
Issues: _______________________________
Time: _____ minutes
```

---

#### TC-9.2: System Statistics

**Scenario**: Admin can view system-wide statistics

**Steps**:
1. Navigate to Admin → Reports → System Statistics
2. Review metrics

**Verification**:
- [ ] Statistics load correctly
- [ ] Breakdown by status is accurate
- [ ] Document count correct
- [ ] No errors in console

**Documentation**:
```
Status: PASS / FAIL
Issues: _______________________________
Time: _____ minutes
```

---

#### TC-9.3: Activity Logs

**Scenario**: Admin can view activity audit trail

**Steps**:
1. Navigate to Admin → Activity Logs
2. Filter by action: "submitted" or "approved"
3. Search for specific user

**Verification**:
- [ ] Logs display all activities
- [ ] Timestamps accurate
- [ ] Filters work
- [ ] Search finds correct entries

**Documentation**:
```
Status: PASS / FAIL
Issues: _______________________________
Time: _____ minutes
```

---

## 📊 Test Results Summary

### Overall Metrics

| Metric | Target | Actual | Status |
|--------|--------|--------|--------|
| Total Test Cases | 25+ | _____ | ⏳ |
| Passed | 23+ | _____ | ⏳ |
| Failed | <2 | _____ | ⏳ |
| Pass Rate | 90%+ | _____ | ⏳ |
| Critical Pass | 100% | _____ | ⏳ |
| Total Duration | 3 hours | _____ | ⏳ |

### Phase Results

| Phase | Test Cases | Passed | Failed | Duration |
|-------|-----------|--------|--------|----------|
| 1 (Auth) | 3 | ___ | ___ | ___ min |
| 2 (Multi-Tenant) | 3 | ___ | ___ | ___ min |
| 3 (Requisitions) | 5 | ___ | ___ | ___ min |
| 4 (Budgets) | 1 | ___ | ___ | ___ min |
| 5 (PO) | 1 | ___ | ___ | ___ min |
| 6 (GRN) | 1 | ___ | ___ | ___ min |
| 7 (Data Integrity) | 2 | ___ | ___ | ___ min |
| 8 (Error Handling) | 2 | ___ | ___ | ___ min |
| 9 (Reporting) | 3 | ___ | ___ | ___ min |
| **TOTAL** | **21** | ___ | ___ | ___ min |

---

## 🐛 Defect Log

### Critical Defects (Blocks Release)

| ID | Title | Status | Notes |
|----|-------|--------|-------|
| D-001 | [Issue] | ⏳ Open | |

### High Priority Defects

| ID | Title | Status | Notes |
|----|-------|--------|-------|
| D-002 | [Issue] | ⏳ Open | |

### Medium Priority Defects

| ID | Title | Status | Notes |
|----|-------|--------|-------|
| D-003 | [Issue] | ⏳ Open | |

---

## ✅ Sign-Off

**Test Execution Completed By**: ___________________________

**Date**: ___________________________

**Pass Criteria Met**: [ ] Yes [ ] No

**Ready for Production**: [ ] Yes [ ] No

**Comments**:
```
_________________________________________________________________
_________________________________________________________________
_________________________________________________________________
```

**Approved By**: ___________________________

**Date**: ___________________________

---

## 🔗 Related Resources

- **E2E-TEST-PLAN.md** - Detailed test cases
- **TESTING-GUIDE.md** - General testing procedures
- **API.http** - API endpoint examples
- **docker-compose.yml** - Services configuration

---

**Start Testing**: ✅ READY

**Good luck with testing!**

---

**Created By**: Claude Code
**Date**: 2025-12-26
