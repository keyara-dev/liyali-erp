# User Creation Flow - Complete Audit

## Overview

Complete audit of the user creation flow from frontend dialog → frontend action → backend API → database.

---

## Critical Issues Found

### 1. ❌ **MAJOR: Type Mismatch Between Dialog and Action**

**Problem:** The dialog component and the action function expect different field structures.

**Dialog sends:**

```typescript
{
  name: fullName,              // ✅ Computed from first_name + last_name
  email: formData.email,
  phone: formData.phone || "",
  password: formData.password,
  first_name: formData.first_name,
  last_name: formData.last_name,
  department_id: formData.department_id || "",
  role: formData.role,
  position: formData.position,
  manNumber: formData.manNumber,
  nrcNumber: formData.nrcNumber,
  contact: formData.contact,
}
```

**Action expects (CreateUserRequest):**

```typescript
{
  username: string;            // ❌ Dialog doesn't send this
  email: string;
  phone?: string;
  password: string;
  first_name: string;
  last_name: string;
  department_id: string;
  role: UserType;
  position?: string;
  manNumber?: string;
  nrcNumber?: string;
  contact?: string;
}
```

**Action actually sends to backend:**

```typescript
{
  name: `${data.first_name} ${data.last_name}`,  // ✅ Computed
  first_name: data.first_name,
  last_name: data.last_name,
  email: data.email,
  password: data.password,
  role: data.role || "requester",
  department_id: data.department_id,
  // ❌ Missing: phone, position, manNumber, nrcNumber, contact
}
```

**Backend expects:**

```go
{
  Email       string `json:"email"`
  Password    string `json:"password"`
  Name        string `json:"name"`
  FirstName   string `json:"first_name"`
  LastName    string `json:"last_name"`
  Role        string `json:"role"`
  DepartmentID string `json:"department_id"`
  Position    string `json:"position"`
  ManNumber   string `json:"manNumber"`
  NrcNumber   string `json:"nrcNumber"`
  Contact     string `json:"contact"`
}
```

**Impact:**

- ❌ `username` field is required in CreateUserRequest but never sent
- ❌ `phone`, `position`, `manNumber`, `nrcNumber`, `contact` are collected in dialog but NOT sent to backend
- ❌ Users fill in these fields thinking they'll be saved, but they're silently dropped
- ❌ TypeScript error: "Object literal may only specify known properties, and 'name' does not exist in type 'CreateUserRequest'"

---

### 2. ❌ **MAJOR: Profile Fields Not Being Sent**

**Problem:** Dialog collects profile fields but action doesn't send them to backend.

**Fields collected but NOT sent:**

- `phone` - User enters it, but it's dropped
- `position` - User enters it, but it's dropped
- `manNumber` - User enters it, but it's dropped
- `nrcNumber` - User enters it, but it's dropped
- `contact` - User enters it, but it's dropped

**Current action code:**

```typescript
const response = await authenticatedApiClient({
  url: url,
  data: {
    name: `${data.first_name} ${data.last_name}`,
    first_name: data.first_name,
    last_name: data.last_name,
    email: data.email,
    password: data.password,
    role: data.role || "requester",
    department_id: data.department_id,
    // ❌ Missing all profile fields!
  },
  method: "POST",
});
```

---

### 3. ❌ **MAJOR: Username Field Required But Never Used**

**Problem:** CreateUserRequest requires `username` field, but:

- Dialog doesn't have username input (correctly removed)
- Dialog doesn't send username
- Backend doesn't accept or use username
- Action expects username but it's never provided

**Impact:** TypeScript compilation error when trying to call createNewUser

---

## Medium Priority Issues

### 4. ⚠️ **Type Definition Inconsistency**

**Problem:** Two different CreateUserRequest types exist:

**Type 1:** `frontend/src/app/_actions/user-actions.ts`

```typescript
export interface CreateUserRequest {
  username: string; // ❌ Required but not used
  email: string;
  phone?: string;
  password: string;
  first_name: string;
  last_name: string;
  department_id: string;
  role: UserType;
  position?: string;
  manNumber?: string;
  nrcNumber?: string;
  contact?: string;
}
```

**Type 2:** `frontend/src/types/user-management.ts`

```typescript
export interface CreateUserRequest {
  email: string;
  name?: string; // Different structure
  role: string;
  department?: string; // Different field name
  departmentId?: string; // Different field name
  password?: string;
  active?: boolean;
  position?: string;
  manNumber?: string;
  nrcNumber?: string;
  contact?: string;
}
```

**Impact:** Confusion about which type to use, inconsistent field names

---

### 5. ⚠️ **Backend Accepts Fields That Frontend Doesn't Send**

**Backend accepts but frontend doesn't send:**

- `phone` - Backend accepts it, frontend collects it, but action doesn't send it
- `position` - Backend accepts it, frontend collects it, but action doesn't send it
- `manNumber` - Backend accepts it, frontend collects it, but action doesn't send it
- `nrcNumber` - Backend accepts it, frontend collects it, but action doesn't send it
- `contact` - Backend accepts it, frontend collects it, but action doesn't send it

---

## Low Priority Issues

### 6. 📝 **Inconsistent Field Naming**

**Problem:** Different naming conventions across layers:

- Frontend dialog: `department_id`
- Type 1: `department_id`
- Type 2: `departmentId` and `department`
- Backend: `department_id`

**Recommendation:** Standardize on snake_case for API communication

---

### 7. 📝 **Missing Validation**

**Problem:** No validation for:

- Email format (relies on HTML5 validation only)
- Password strength (frontend doesn't validate, only backend does)
- Phone number format
- NRC number format
- Man number format

---

## Backend Validation (✅ Working Correctly)

The backend handler has good validation:

```go
// ✅ Validates email is required
if req.Email == "" {
  return utils.SendBadRequestError(c, "Email is required")
}

// ✅ Validates password length
if req.Password == "" || len(req.Password) < 8 {
  return utils.SendBadRequestError(c, "Password is required and must be at least 8 characters")
}

// ✅ Validates name
if req.Name == "" && req.FirstName == "" {
  return utils.SendBadRequestError(c, "Name or first name is required")
}

// ✅ Sets default role
if req.Role == "" {
  req.Role = "requester"
}

// ✅ Validates password strength
if err := utils.ValidatePasswordStrength(req.Password); err != nil {
  return utils.SendBadRequestError(c, "Password validation failed: "+err.Error())
}

// ✅ Hashes password
hashedPassword, err := utils.HashPassword(req.Password)

// ✅ Handles role UUID → role name conversion
if _, err := uuid.Parse(req.Role); err == nil {
  // Look up role name from UUID
}

// ✅ Creates user with all fields
user := &models.User{
  ID:                    uuid.New().String(),
  Email:                 req.Email,
  Name:                  req.Name,
  Password:              hashedPassword,
  Role:                  roleName,
  Active:                true,
  CurrentOrganizationID: &tenant.OrganizationID,
  Position:              req.Position,
  ManNumber:             req.ManNumber,
  NrcNumber:             req.NrcNumber,
  Contact:               req.Contact,
}

// ✅ Adds user to organization with department
if err := orgService.AddMemberWithDepartment(tenant.OrganizationID, user.ID, roleName, departmentPtr); err != nil {
  // Handle error
}
```

**Backend is ready to accept all fields - frontend just needs to send them!**

---

## Required Fixes

### Fix 1: Update CreateUserRequest Type

**File:** `frontend/src/app/_actions/user-actions.ts`

```typescript
export interface CreateUserRequest {
  // Remove username - not used
  email: string;
  phone?: string;
  password: string;
  first_name: string;
  last_name: string;
  name?: string; // Add for computed full name
  department_id?: string;
  role: UserType;
  // Profile fields
  position?: string;
  manNumber?: string;
  nrcNumber?: string;
  contact?: string;
}
```

### Fix 2: Update createNewUser Action

**File:** `frontend/src/app/_actions/user-actions.ts`

```typescript
export async function createNewUser(
  data: CreateUserRequest,
): Promise<APIResponse> {
  const url = `/api/v1/organization/users`;

  try {
    // Compute full name if not provided
    const fullName = data.name || `${data.first_name} ${data.last_name}`.trim();

    const response = await authenticatedApiClient({
      url: url,
      data: {
        name: fullName,
        first_name: data.first_name,
        last_name: data.last_name,
        email: data.email,
        password: data.password,
        role: data.role || "requester",
        department_id: data.department_id,
        // ADD PROFILE FIELDS:
        phone: data.phone,
        position: data.position,
        manNumber: data.manNumber,
        nrcNumber: data.nrcNumber,
        contact: data.contact,
      },
      method: "POST",
    });

    if (!response.data?.success) {
      return handleError(
        new Error(response.data?.message || "Failed to create user"),
        "POST",
        url,
      );
    }

    console.log("User created successfully:", response.data);
    revalidatePath("/admin/users");

    return successResponse(response.data?.data, "User created successfully");
  } catch (error: Error | any) {
    return handleError(error, "POST", url);
  }
}
```

### Fix 3: Remove Duplicate Type Definition

**File:** `frontend/src/types/user-management.ts`

```typescript
// REMOVE this duplicate type definition
// Use the one from user-actions.ts instead
```

### Fix 4: Update Dialog Component

**File:** `frontend/src/app/(private)/admin/_components/create-user-dialog.tsx`

The dialog is already correct - it computes name and sends all fields.
Just need to ensure the type matches.

---

## Testing Checklist

After implementing fixes:

### Frontend Tests

- [ ] Create user with all fields filled
- [ ] Verify no TypeScript errors
- [ ] Check console for API request payload
- [ ] Verify all fields are in the request body

### Backend Tests

- [ ] Create user and check database
- [ ] Verify `position` is saved
- [ ] Verify `phone` is saved
- [ ] Verify `manNumber` is saved
- [ ] Verify `nrcNumber` is saved
- [ ] Verify `contact` is saved
- [ ] Verify user is added to organization
- [ ] Verify user is assigned to department
- [ ] Verify role is set correctly

### Integration Tests

- [ ] Create user from dialog
- [ ] Refresh user list
- [ ] View user details
- [ ] Verify all profile fields are displayed
- [ ] Edit user and verify fields persist

---

## Data Flow Diagram

```
┌─────────────────────────────────────────────────────────────┐
│ 1. User fills form in CreateUserDialog                      │
│    - first_name, last_name, email, role, department         │
│    - position, phone, manNumber, nrcNumber, contact         │
│    - password (auto-generated)                              │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│ 2. Dialog computes name and calls createUserMutation        │
│    name = `${first_name} ${last_name}`                      │
│    ❌ Currently: Sends all fields                           │
│    ❌ Problem: Type mismatch with CreateUserRequest         │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│ 3. createNewUser action (user-actions.ts)                   │
│    ❌ Currently: Only sends basic fields                    │
│    ❌ Problem: Drops phone, position, manNumber, etc.       │
│    ✅ Should: Send ALL fields to backend                    │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│ 4. Backend API (CreateOrganizationUser)                     │
│    ✅ Validates all fields                                  │
│    ✅ Hashes password                                       │
│    ✅ Resolves role UUID → role name                        │
│    ✅ Creates user with ALL fields                          │
│    ✅ Adds user to organization                             │
│    ✅ Assigns user to department                            │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│ 5. Database                                                  │
│    ✅ User record created with all fields                   │
│    ✅ Organization membership created                       │
│    ✅ Department assignment created                         │
└─────────────────────────────────────────────────────────────┘
```

---

## Summary

### What's Working ✅

- Backend API is fully functional and accepts all fields
- Backend validation is comprehensive
- Backend creates user, adds to org, assigns department
- Dialog collects all necessary fields
- Dialog computes full name correctly
- Password generation works
- Role and department selection works

### What's Broken ❌

1. **Type mismatch** - CreateUserRequest expects `username` but dialog doesn't send it
2. **Profile fields dropped** - Dialog collects them but action doesn't send them
3. **Duplicate type definitions** - Two different CreateUserRequest types
4. **Missing fields in API call** - Action only sends basic fields, not profile fields

### Impact

- Users can fill in profile fields but they're never saved
- TypeScript compilation error when calling createNewUser
- Confusing UX - fields appear to work but data is lost

### Effort to Fix

- Update CreateUserRequest type: 5 minutes
- Update createNewUser action: 10 minutes
- Remove duplicate type: 2 minutes
- Testing: 20 minutes

**Total: ~40 minutes**

---

## Recommendation

**Priority: HIGH - Fix immediately**

The user creation flow is partially broken. Users can create accounts, but profile information (phone, position, manNumber, nrcNumber, contact) is silently dropped. This creates a poor user experience and data loss.

The backend is ready and working correctly. We just need to:

1. Fix the frontend type definition
2. Update the action to send all fields
3. Remove duplicate types

All fixes are in the frontend - no backend changes needed.
