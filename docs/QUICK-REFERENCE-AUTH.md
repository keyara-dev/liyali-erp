# Quick Reference: Authentication & RBAC

## 🎯 TL;DR - What Happened

### Before
```
❌ DEMO_USERS hardcoded
❌ No password validation
❌ Single role globally
❌ No organization context
```

### After
```
✅ Real backend API integration
✅ bcrypt password verification
✅ Role per organization
✅ Multi-tenancy support
```

---

## 🚀 Get Started in 3 Steps

### Step 1: Seed Database
```bash
cd backend
go run cmd/seed/main.go
```

### Step 2: Start Backend
```bash
cd backend
go run cmd/main.go  # Runs on :8080
```

### Step 3: Login
```
Frontend: http://localhost:3001/login
Email: requester@liyali.com
Password: password123
```

---

## 📚 Documentation Map

| Need | Document |
|------|----------|
| **Where to start** | [AUTHENTICATION-INTEGRATION-INDEX.md](./AUTHENTICATION-INTEGRATION-INDEX.md) |
| **What was done** | [IMPLEMENTATION-SUMMARY.md](./IMPLEMENTATION-SUMMARY.md) |
| **How RBAC works** | [RBAC-AND-ORGANIZATION-ARCHITECTURE.md](./RBAC-AND-ORGANIZATION-ARCHITECTURE.md) |
| **Session complete** | [00-SESSION-COMPLETION-AUTH-INTEGRATION.md](./00-SESSION-COMPLETION-AUTH-INTEGRATION.md) |

---

## 🔑 Key Files Changed

### Backend
| File | What Changed |
|------|-------------|
| `backend/handlers/auth.go` | ✅ Password verification enabled |
| `backend/cmd/seed/main.go` | ✅ Test users created |
| `backend/utils/jwt.go` | ✅ Org context in token |

### Frontend
| File | What Changed |
|------|-------------|
| `frontend/.env` | ✅ Backend API URLs |
| `frontend/src/app/_actions/auth.ts` | ✅ Uses real API |
| `frontend/src/lib/auth.ts` | ✅ Removed DEMO_USERS |
| `frontend/src/types/auth.ts` | ✅ Added organization_id |
| `frontend/src/app/_actions/api-config.ts` | ✅ Added X-Org header |

---

## 🧪 Test Users

```
requester@liyali.com → password123 → Can create requisitions
manager@liyali.com → password123 → Can approve
finance@liyali.com → password123 → Can manage budgets
director@liyali.com → password123 → Can approve
cfo@liyali.com → password123 → Can manage finances
compliance@liyali.com → password123 → Read-only
admin@liyali.com → password123 → All permissions
```

---

## 🔐 How It Works Now

### Login Flow
```
User enters email/password
    ↓
POST /api/v1/auth/login → Backend
    ↓
Backend checks: password hash matches ✓
    ↓
Generate JWT with: { userID, email, role, orgID }
    ↓
Frontend: Store token in session cookie
    ↓
Frontend: Add Bearer token to all API requests
    ↓
Backend: Validate token + org membership
    ↓
Approve request, return org-scoped data
```

### Authorization
```
Request arrives
    ↓
AuthMiddleware: Validate JWT token
    ↓
TenantMiddleware: Verify org membership
    ↓
Handler: Check permissions/role
    ↓
Data filtered by organization_id
    ↓
Response: Only user can access
```

---

## 📊 RBAC Model

### User in Multiple Organizations
```
User: john@example.com
  Global role: "requester"

  Organization A: role = "admin"
  Organization B: role = "requester"
  Organization C: role = "viewer"

In Org A: Can do admin things
In Org B: Can create requisitions
In Org C: Can only view
```

### Permission Mapping (Coming in Phase 2)
```
Role: requester
Permissions: [create_requisition, view_requisition, create_draft]

Role: approver
Permissions: [view_requisition, approve_requisition, reject_requisition]

Role: finance
Permissions: [create_budget, view_budget, manage_vendors, ...]

Role: admin
Permissions: [*] (everything)
```

---

## 🧩 Response Pattern

### All responses use helpers:
```typescript
// Success
return successResponse(data, "message");

// Unauthorized
return unauthorizedResponse("message");

// Bad request
return badRequestResponse("message");

// Error
return handleError(error, "GET", "/api/endpoint");
```

---

## 🔒 Security Checklist

- ✅ Passwords hashed (bcrypt)
- ✅ JWT validated
- ✅ Org membership verified
- ✅ httpOnly cookies
- ✅ Data filtered by org_id
- ✅ Role checks enforced
- ✅ Token refresh works

---

## ❓ Common Issues

| Problem | Solution |
|---------|----------|
| **"Invalid credentials"** | Run seed script, check user exists |
| **"No org found"** | Verify X-Organization-ID header sent |
| **CORS error** | Check BASE_URL matches backend |
| **Can't create requisition** | Check user role, may need approver role |

---

## 📞 Need Help?

1. **Can't login?** → Check IMPLEMENTATION-SUMMARY.md troubleshooting
2. **How does RBAC work?** → See RBAC-AND-ORGANIZATION-ARCHITECTURE.md
3. **What files changed?** → See SESSION-COMPLETION doc
4. **How do I test?** → See IMPLEMENTATION-SUMMARY.md verification checklist

---

## 🚀 Next: Phase 2

Implementing permission-based access control:
- Create permissions.go service
- Add RequirePermission middleware
- Update handlers for permission checks
- Add frontend permission utilities

See [RBAC-AND-ORGANIZATION-ARCHITECTURE.md](./RBAC-AND-ORGANIZATION-ARCHITECTURE.md#8-implementation-roadmap) Section 8 for details.

---

## 📈 File Structure Reference

```
Frontend API Calls:
src/app/_actions/auth.ts
├─ loginAction(email, password)
├─ getCurrentUserAction()
├─ getRefreshToken()
├─ changePassword()
├─ logoutAction()
└─ logUserOut()

All use: authenticatedApiClient() with Bearer token

Session Management:
src/lib/auth.ts
├─ createAuthSession()
├─ verifySession()
├─ updateAuthSession()
├─ deleteSession()
├─ getCurrentUser()
├─ hasRole()
└─ isAdmin()
```

---

## ✨ What's New

### Backend
- Real password validation (bcrypt)
- JWT with org context
- Test user seed script

### Frontend
- Real API integration (no mock data)
- Session with organization_id
- X-Organization-ID header support
- Helper response functions

### Documentation
- Complete RBAC architecture (2000+ lines)
- Implementation roadmap (4 phases)
- Test user credentials
- Quick reference guides

---

**Status: ✅ Ready to use**

*For detailed information, see [AUTHENTICATION-INTEGRATION-INDEX.md](./AUTHENTICATION-INTEGRATION-INDEX.md)*
