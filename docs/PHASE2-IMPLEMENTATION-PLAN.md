# Phase 2 Implementation Plan: Auto-Create Personal Organization on Signup

## 🎯 Objective
Implement Scenario C: When a user registers, automatically create a personal organization and set it as their current organization. User redirects immediately to dashboard with zero intermediate screens.

## 📊 Project Scope
- **Duration:** 8-10 hours (estimated)
- **Complexity:** Medium
- **Risk Level:** Low (changes isolated to auth/org flows)
- **Status:** Planning → In Progress → Testing → Complete

---

## 📋 Implementation Breakdown

### Phase 2A: Backend Core Implementation (3-4 hours)

#### Task 2A.1: Fix Password Storage ⚠️ CRITICAL
**Time:** 15 minutes
**Importance:** High - Security issue

**File:** `backend/handlers/auth.go` (Register function)

**Current Issue:**
```go
// Note: In a full implementation, you'd want to store the hashedPassword
// For now, we're storing the plain password for demo purposes
```

**Action:**
- [ ] Change to store `hashedPassword` instead of plain text password
- [ ] Verify bcrypt hashing is applied before storage
- [ ] Test password verification works

**Code Change:**
```go
// Before:
Password: req.Password, // ❌ Plain text

// After:
Password: hashedPassword, // ✅ Hashed
```

---

#### Task 2A.2: Update AuthResponse Type
**Time:** 15 minutes
**Importance:** Medium

**File:** `backend/types/auth.go`

**Changes:**
- [ ] Add `OrganizationResponse` struct with fields: ID, Name, Slug, Description, Active, Tier, CreatedAt
- [ ] Add `Organization *OrganizationResponse` field to `AuthResponse`

**New Struct:**
```go
type OrganizationResponse struct {
    ID          string `json:"id"`
    Name        string `json:"name"`
    Slug        string `json:"slug"`
    Description string `json:"description,omitempty"`
    Active      bool   `json:"active"`
    Tier        string `json:"tier"`
    CreatedAt   string `json:"createdAt"`
}
```

---

#### Task 2A.3: Implement Auto-Organization Creation in Register
**Time:** 60 minutes
**Importance:** Critical

**File:** `backend/handlers/auth.go` (Register function)

**Changes Needed:**

1. **Import OrganizationService**
   ```go
   import "github.com/liyali/liyali-gateway/services"
   ```

2. **After creating user, create organization**
   ```go
   // Create personal organization
   orgService := services.NewOrganizationService(config.DB)
   org, err := orgService.CreateOrganization(
       context.Background(),
       &models.Organization{
           Name: newUser.Name,
           Slug: strings.ToLower(strings.ReplaceAll(newUser.Email, "@", "-")),
           Description: "Personal Organization",
           Type: "personal",
       },
       newUser.ID,
   )
   if err != nil {
       // Log error but continue - org can be created on next login
       log.Printf("Failed to create personal org: %v", err)
   }
   ```

3. **Add user as admin member**
   ```go
   if org != nil {
       member := &models.OrganizationMember{
           ID:             utils.GenerateID(),
           OrganizationID: org.ID,
           UserID:         newUser.ID,
           Role:           "admin",
           Active:         true,
           JoinedAt:       time.Now(),
       }
       config.DB.Create(&member)
   }
   ```

4. **Set as current organization**
   ```go
   if org != nil {
       config.DB.Model(&newUser).Update("current_organization_id", org.ID)
   }
   ```

5. **Generate token with org context**
   ```go
   token, _ := utils.GenerateToken(
       newUser.ID,
       newUser.Email,
       newUser.Name,
       newUser.Role,
       &org.ID, // ← Include org ID
   )
   ```

6. **Update response**
   ```go
   return c.Status(fiber.StatusCreated).JSON(types.AuthResponse{
       Success: true,
       Message: "Registration successful",
       Token: token,
       User: &userResponse,
       Organization: &types.OrganizationResponse{
           ID: org.ID,
           Name: org.Name,
           Slug: org.Slug,
           // ... other fields
       },
   })
   ```

**Checklist:**
- [ ] Code compiles
- [ ] All imports present
- [ ] Error handling in place
- [ ] Response includes org
- [ ] Password is hashed

---

#### Task 2A.4: Backend Testing
**Time:** 1 hour
**Importance:** High

**Test Cases:**
- [ ] Register with valid data → 201 + org created
- [ ] New user is org admin → check members table
- [ ] Current org set → check users.current_organization_id
- [ ] JWT has org ID → decode token and verify
- [ ] Email exists → 409 response
- [ ] Weak password → 400 response
- [ ] Invalid role → 400 response
- [ ] Org creation failure → still create user (graceful)

**Testing Method:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "newuser@example.com",
    "password": "Password123",
    "name": "New User",
    "role": "requester"
  }'
```

**Verify Response Contains:**
- ✅ `organization` field with ID, name, slug
- ✅ `token` field with valid JWT
- ✅ `user` field with user details
- ✅ `success: true`

---

### Phase 2B: Frontend Implementation (2.5-3 hours)

#### Task 2B.1: Update Auth Types
**Time:** 15 minutes
**Importance:** Medium

**Files:**
- `frontend/src/types/auth.ts` (add new types)
- `frontend/src/types/` (create organization.ts if needed)

**Changes:**
- [ ] Add `Organization` interface
- [ ] Add `RegistrationResponse` interface
- [ ] Update imports in auth actions

**New Types:**
```typescript
export interface Organization {
  id: string;
  name: string;
  slug: string;
  description?: string;
  active: boolean;
  tier: string;
  createdAt: string;
  updatedAt?: string;
}

export interface RegistrationResponse {
  user: User;
  organization: Organization;
}
```

---

#### Task 2B.2: Implement createNewAccount Action
**Time:** 45 minutes
**Importance:** Critical

**File:** `frontend/src/app/_actions/auth.ts`

**Replace Stub Implementation:**

```typescript
export async function createNewAccount(data: {
  email: string;
  name: string;
  password: string;
  role?: string;
}): Promise<APIResponse<any>> {
  const url = `/api/v1/auth/register`;

  try {
    const response = await axios.post(url, {
      email: data.email,
      name: data.name,
      password: data.password,
      role: data.role || "requester", // Default role
    });

    const responseData = response?.data;

    if (!responseData.success || !responseData.token) {
      return unauthorizedResponse(
        responseData.message || "Registration failed"
      );
    }

    // Create session with token AND org context
    await createAuthSession({
      access_token: responseData.token,
      role: responseData.user.role,
      user_id: responseData.user.id,
      organization_id: responseData.organization.id, // ← NEW
    });

    return successResponse(
      {
        user: responseData.user,
        organization: responseData.organization, // ← NEW
      },
      responseData.message
    );
  } catch (error: any) {
    return handleError(error, "POST", url);
  }
}
```

**Checklist:**
- [ ] Calls correct endpoint: `/api/v1/auth/register`
- [ ] Sends: email, name, password, role
- [ ] Extracts: token, user, organization
- [ ] Stores org_id in session
- [ ] Uses helper functions (successResponse, handleError)
- [ ] Handles errors gracefully

---

#### Task 2B.3: Update Signup Component
**Time:** 60 minutes
**Importance:** High

**File:** `frontend/src/app/(auth)/_components/signup.tsx` (or existing signup location)

**Changes:**

1. **Simplify form fields** (if multi-step exists)
   - Collect: email, name, password, password confirmation
   - Optional: role (default to "requester")
   - Remove: WhatsApp, shop name, description (these can come later)

2. **Update form submission**
   ```typescript
   const handleSubmit = async (e: React.FormEvent) => {
     e.preventDefault();
     setError("");
     setLoading(true);

     try {
       const result = await createNewAccount({
         email,
         name,
         password,
         role: "requester", // or from form if desired
       });

       if (!result.success) {
         setError(result.message || "Registration failed");
         return;
       }

       // Redirect to dashboard - org context ready
       router.push("/home");
     } catch (err: any) {
       setError(err.message || "An error occurred");
     } finally {
       setLoading(false);
     }
   };
   ```

3. **Update error messages**
   ```typescript
   {error && (
     <div className="p-3 bg-red-100 border border-red-300 rounded">
       {error}
     </div>
   )}
   ```

4. **Add password validation**
   - Min 8 characters
   - At least 1 uppercase
   - At least 1 lowercase
   - At least 1 digit

**Checklist:**
- [ ] Form collects all required fields
- [ ] Password validation works
- [ ] Submit button disabled while loading
- [ ] Error messages displayed
- [ ] Success → redirect to /home (not org selection)
- [ ] Loading spinner shown during submission

---

#### Task 2B.4: Frontend Testing
**Time:** 15 minutes
**Importance:** Medium

**Test Cases:**
- [ ] Fill form with valid data
- [ ] Click submit
- [ ] Verify loading state
- [ ] Verify API call made
- [ ] Verify redirect to /home
- [ ] Check session has org_id
- [ ] Try registering with existing email → error shown
- [ ] Try weak password → error shown
- [ ] Try with missing fields → error shown

**Manual Testing:**
1. Open http://localhost:3001/signup
2. Fill form with new user data
3. Submit
4. Verify redirected to dashboard
5. Check browser DevTools → Application → Cookies
6. Verify AUTH_SESSION cookie exists
7. Can access organization switcher

---

### Phase 2C: Integration & Testing (2-3 hours)

#### Task 2C.1: Update Organization Context
**Time:** 30 minutes
**Importance:** Medium

**File:** `frontend/src/contexts/organization-context.tsx`

**Changes:**
- [ ] Handle initial load with org from signup
- [ ] Set current org from session or first available org
- [ ] Load organizations list from backend
- [ ] Handle new user with single personal org

**Key Update:**
```typescript
useEffect(() => {
  const loadOrganizations = async () => {
    const orgs = await fetchUserOrganizations();

    // New user has single personal org
    if (orgs.length === 1) {
      setCurrentOrganization(orgs[0]);
    } else if (orgs.length > 1) {
      // User has multiple orgs - use session current org
      const currentOrgId = await getCurrentOrganizationId();
      const current = orgs.find(o => o.id === currentOrgId);
      setCurrentOrganization(current || orgs[0]);
    }
  };

  loadOrganizations();
}, []);
```

---

#### Task 2C.2: Comprehensive Testing
**Time:** 1.5 hours
**Importance:** Critical

**Backend Unit Tests:**
- [ ] Test password hashing
- [ ] Test org creation
- [ ] Test member addition
- [ ] Test token generation with org
- [ ] Test all error cases

**Frontend Component Tests:**
- [ ] Signup form renders
- [ ] Form validation works
- [ ] Submit calls API correctly
- [ ] Redirect works
- [ ] Error handling works

**Integration Tests:**
- [ ] Complete signup flow end-to-end
- [ ] Session contains org_id
- [ ] Can access dashboard immediately
- [ ] Can switch orgs if added to another
- [ ] Org context available in all routes

**Manual E2E Test Checklist:**
- [ ] Register with: john+test@example.com / SecurePassword123 / John Test
- [ ] Verify redirected to /home
- [ ] Check DevTools cookies for AUTH_SESSION
- [ ] Verify can create requisition (requester permission)
- [ ] Verify cannot approve (not approver)
- [ ] Logout and login with same credentials
- [ ] Verify org context persists
- [ ] Check database: user has current_organization_id set
- [ ] Check database: organization created with user's name
- [ ] Check database: user is admin member of org

---

#### Task 2C.3: Documentation Update
**Time:** 30 minutes
**Importance:** Medium

**Files to Update:**
- [ ] IMPLEMENTATION-SUMMARY.md → Add Phase 2 completion notes
- [ ] RBAC-AND-ORGANIZATION-ARCHITECTURE.md → Verify Scenario C documented
- [ ] README or QUICK-START → Add registration flow

**Document:**
- Registration endpoint behavior
- Auto-organization creation details
- Test user creation process
- Troubleshooting guide for org creation

---

## 🎯 Success Criteria

### Backend Success
- ✅ Register endpoint returns 201
- ✅ Response includes `organization` field
- ✅ User has admin role in created org
- ✅ User.current_organization_id is set
- ✅ JWT token includes org ID
- ✅ Password stored as hash
- ✅ All validation working
- ✅ Error cases handled gracefully

### Frontend Success
- ✅ Signup form submits correctly
- ✅ API call made with correct payload
- ✅ Session created with token + org_id
- ✅ User redirected to /home immediately
- ✅ Dashboard loads without org selection step
- ✅ Organization context available
- ✅ Error messages displayed properly
- ✅ Loading state shown during submission

### Integration Success
- ✅ End-to-end signup works
- ✅ New user can create requisitions immediately
- ✅ Permissions enforced based on role
- ✅ Can logout and login with new credentials
- ✅ Org context persists across sessions
- ✅ Organization switcher works

### Quality Success
- ✅ All tests passing
- ✅ No console errors
- ✅ No database errors
- ✅ Code follows patterns established
- ✅ Response helpers used consistently
- ✅ Error handling comprehensive

---

## 📅 Timeline

| Phase | Task | Est. Time | Status |
|-------|------|-----------|--------|
| 2A | Password Storage Fix | 15 min | ⏳ TODO |
| 2A | AuthResponse Type Update | 15 min | ⏳ TODO |
| 2A | Register Handler Implementation | 60 min | ⏳ TODO |
| 2A | Backend Testing | 60 min | ⏳ TODO |
| 2B | Auth Types Update | 15 min | ⏳ TODO |
| 2B | createNewAccount Action | 45 min | ⏳ TODO |
| 2B | Signup Component Update | 60 min | ⏳ TODO |
| 2B | Frontend Testing | 15 min | ⏳ TODO |
| 2C | Organization Context Update | 30 min | ⏳ TODO |
| 2C | Comprehensive Testing | 90 min | ⏳ TODO |
| 2C | Documentation Update | 30 min | ⏳ TODO |
| **TOTAL** | | **8.5 hours** | |

---

## 🚀 Implementation Order

**Recommended sequence:**

1. **Start Backend:** Fix password storage + update types (30 min)
2. **Implement Registration:** Core org creation logic (60 min)
3. **Test Backend:** Verify with Postman/curl (1 hour)
4. **Frontend Types:** Add new interfaces (15 min)
5. **Implement Auth Action:** createNewAccount function (45 min)
6. **Update Signup:** UI component and form (60 min)
7. **Test Frontend:** Manual signup flow (15 min)
8. **Integration:** Org context + full E2E testing (2 hours)
9. **Documentation:** Update guides (30 min)

---

## 🔒 Security Checklist

Before deployment:
- [ ] Password always hashed (never stored plain)
- [ ] JWT validation working
- [ ] Organization isolation verified
- [ ] Admin role properly assigned
- [ ] No SQL injection vectors
- [ ] CORS headers correct
- [ ] Input validation comprehensive
- [ ] Error messages don't leak sensitive info
- [ ] Rate limiting on register endpoint (optional)

---

## 📞 Troubleshooting Guide

### "Organization not created"
- [ ] Check OrganizationService is properly injected
- [ ] Verify database connection works
- [ ] Check error logs for DB issues
- [ ] Verify foreign key constraints

### "User not member of org"
- [ ] Verify member.role is set to "admin"
- [ ] Check organization_members table has record
- [ ] Verify JoinedAt timestamp is set

### "JWT doesn't include org"
- [ ] Verify org.ID is passed to GenerateToken
- [ ] Check token claims after decoding
- [ ] Verify org creation succeeded

### "Frontend redirect fails"
- [ ] Check /home route exists
- [ ] Verify router.push syntax correct
- [ ] Check console for JavaScript errors
- [ ] Verify session has org_id

### "Can't login after signup"
- [ ] Verify password is hashed, not plain
- [ ] Check bcrypt.CompareHashAndPassword is used
- [ ] Verify user exists in database
- [ ] Check for typos in email/password check

---

## 📝 Notes

- Scenario C (auto-personal org) is recommended for MVP
- No database migrations needed (tables already exist)
- Code follows established patterns in auth.ts
- Zero breaking changes to existing users
- Can add invitation system (Scenario B) later

---

## Related Documentation

See also:
- [ORGANIZATION-ONBOARDING-STRATEGY.md](./ORGANIZATION-ONBOARDING-STRATEGY.md) - Detailed scenario descriptions
- [RBAC-AND-ORGANIZATION-ARCHITECTURE.md](./RBAC-AND-ORGANIZATION-ARCHITECTURE.md) - Architecture overview
- [IMPLEMENTATION-SUMMARY.md](./IMPLEMENTATION-SUMMARY.md) - Current implementation status

---

**Status:** Ready for Implementation
**Assigned To:** [Developer Name]
**Start Date:** [Today]
**Target Completion:** [8.5 hours from start]
