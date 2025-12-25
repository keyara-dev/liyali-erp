# Organization Onboarding Strategy: User Registration & RBAC Integration

## Overview

This document addresses the question: **"When a user creates an account, how do they get into an organization?"**

It provides three strategic approaches with implementation guides, pros/cons, and recommendations.

---

## Current State Analysis

### What Works ✅
- Backend authentication system is real (password verification, JWT tokens)
- RBAC model supports multiple organizations per user
- Multi-tenancy is properly implemented
- Test users can be created and authenticated

### What's Missing ❌
- User registration flow not integrated
- No automatic organization assignment
- No invitation system
- No organization creation UI

---

## Three Organization Onboarding Scenarios

### Scenario A: Organization Join Code
**"I have a company code from my organization"**

#### Flow
```
1. User signs up (email, password, name)
2. User enters organization code during signup
3. Backend validates code, adds user to organization
4. User automatically logged in
5. Redirected to organization dashboard
```

#### Pros
- ✅ User joins specific organization immediately
- ✅ Admin controls who joins via codes
- ✅ No intermediate setup screens
- ✅ Secure (code prevents spam)

#### Cons
- ❌ Requires admin to generate codes
- ❌ User must know their code
- ❌ What if user doesn't have a code?

#### Implementation Effort
- **Backend:** 2-3 hours (code generation, validation)
- **Frontend:** 1-2 hours (signup form enhancement)
- **Total:** 3-5 hours

#### Code Example

```go
// Backend: handlers/auth.go - Enhanced Register
func Register(c *fiber.Ctx) error {
    var req struct {
        Email       string `json:"email"`
        Password    string `json:"password"`
        Name        string `json:"name"`
        OrgCode     string `json:"org_code"` // ← NEW
    }
    c.BodyParser(&req)

    // Create user
    newUser := models.User{
        ID:     utils.GenerateUserID(),
        Email:  req.Email,
        Name:   req.Name,
        Role:   "requester", // Default role
        Active: true,
    }
    config.DB.Create(&newUser)

    var orgID string

    // If org code provided, validate and add user
    if req.OrgCode != "" {
        var orgCode models.OrganizationCode
        result := config.DB.Where("code = ? AND expires_at > ?",
            req.OrgCode, time.Now()).
            First(&orgCode)

        if result.Error != nil {
            return c.Status(400).JSON(fiber.Map{
                "error": "Invalid or expired organization code",
            })
        }

        orgID = orgCode.OrganizationID

        // Add user to organization
        member := models.OrganizationMember{
            ID:             utils.GenerateID(),
            OrganizationID: orgID,
            UserID:         newUser.ID,
            Role:           "requester",
            Active:         true,
            JoinedAt:       time.Now(),
        }
        config.DB.Create(&member)

        // Update code usage
        config.DB.Model(&orgCode).Update("used_count", orgCode.UsedCount+1)
    } else {
        // No code: Create personal organization
        return c.Status(400).JSON(fiber.Map{
            "error": "Organization code is required for registration",
        })
    }

    newUser.CurrentOrganizationID = &orgID
    config.DB.Save(&newUser)

    token, _ := utils.GenerateToken(
        newUser.ID, newUser.Email, newUser.Name, newUser.Role, &orgID)

    return c.JSON(types.SuccessResponse{
        Success: true,
        Message: "User registered successfully",
        Token:   token,
        User:    newUser,
    })
}
```

---

### Scenario B: Email Invitation Link
**"Admin invited me - I click a link to sign up"**

#### Flow
```
1. Organization admin clicks "Invite Member"
2. Admin enters invitee email and selects role
3. System generates unique invite token (JWT)
4. Email sent with signup link: /signup?token=xxx
5. New user clicks link
6. Signup form pre-filled with organization + role
7. User completes signup
8. System auto-adds user to organization with selected role
9. User logged in and ready to go
```

#### Pros
- ✅ Seamless user onboarding
- ✅ Pre-filled organization + role
- ✅ Admin controls membership
- ✅ User only needs email invitation
- ✅ Can track who was invited

#### Cons
- ❌ Requires email service integration
- ❌ Most complex implementation
- ❌ Invite tokens can expire
- ❌ Need invite UI in admin section

#### Implementation Effort
- **Backend:** 4-5 hours (token system, invite endpoints)
- **Frontend:** 2-3 hours (signup acceptance flow)
- **Email:** 1-2 hours (email service integration)
- **Total:** 7-10 hours

#### Code Example (Backend)

```go
// Backend: models/organization.go - Add invite model
type OrganizationInvite struct {
    ID             string     `gorm:"primaryKey"`
    OrganizationID string
    Organization   *Organization
    Email          string
    Role           string
    Token          string
    UsedAt         *time.Time
    ExpiresAt      time.Time
    CreatedBy      string
    CreatedAt      time.Time
}

// Backend: handlers/organization.go - Create invite
func InviteMember(c *fiber.Ctx) error {
    tenant := middleware.GetTenantContext(c)
    userID := c.Locals("userID").(string)

    // Only admins can invite
    if tenant.UserRole != "admin" {
        return c.Status(403).JSON(fiber.Map{
            "error": "Only admins can invite members",
        })
    }

    var req struct {
        Email string `json:"email"`
        Role  string `json:"role"`
    }
    c.BodyParser(&req)

    // Create invite token (valid for 7 days)
    inviteToken, _ := utils.GenerateInviteToken(
        tenant.OrganizationID, req.Email, req.Role)

    invite := models.OrganizationInvite{
        ID:             utils.GenerateID(),
        OrganizationID: tenant.OrganizationID,
        Email:          req.Email,
        Role:           req.Role,
        Token:          inviteToken,
        ExpiresAt:      time.Now().AddDate(0, 0, 7),
        CreatedBy:      userID,
        CreatedAt:      time.Now(),
    }
    config.DB.Create(&invite)

    // Send email (implementation depends on email service)
    signupLink := fmt.Sprintf(
        "http://localhost:3001/signup?inviteToken=%s", inviteToken)
    // sendInviteEmail(req.Email, signupLink)

    return c.JSON(invite)
}

// Backend: handlers/auth.go - Register with invite
func Register(c *fiber.Ctx) error {
    var req struct {
        Email       string `json:"email"`
        Password    string `json:"password"`
        Name        string `json:"name"`
        InviteToken string `json:"invite_token"`
    }
    c.BodyParser(&req)

    // Create user
    newUser := models.User{
        ID:     utils.GenerateUserID(),
        Email:  req.Email,
        Name:   req.Name,
        Role:   "requester",
        Active: true,
    }
    config.DB.Create(&newUser)

    var orgID string

    // If invite token provided, validate and add to org
    if req.InviteToken != "" {
        var invite models.OrganizationInvite
        result := config.DB.Where(
            "token = ? AND email = ? AND expires_at > ? AND used_at IS NULL",
            req.InviteToken, req.Email, time.Now()).
            First(&invite)

        if result.Error != nil {
            return c.Status(400).JSON(fiber.Map{
                "error": "Invalid, expired, or already used invitation",
            })
        }

        orgID = invite.OrganizationID

        // Add user to organization with invited role
        member := models.OrganizationMember{
            ID:             utils.GenerateID(),
            OrganizationID: orgID,
            UserID:         newUser.ID,
            Role:           invite.Role,
            Active:         true,
            JoinedAt:       time.Now(),
        }
        config.DB.Create(&member)

        // Mark invite as used
        now := time.Now()
        config.DB.Model(&invite).Update("used_at", now)
    } else {
        return c.Status(400).JSON(fiber.Map{
            "error": "Invitation token required",
        })
    }

    newUser.CurrentOrganizationID = &orgID
    config.DB.Save(&newUser)

    token, _ := utils.GenerateToken(
        newUser.ID, newUser.Email, newUser.Name, newUser.Role, &orgID)

    return c.JSON(types.SuccessResponse{
        Success: true,
        Message: "User registered successfully",
        Token:   token,
        User:    newUser,
    })
}
```

---

### Scenario C: Auto-Create Personal Organization (RECOMMENDED)
**"I'm starting my own thing - I get my own personal org"**

#### Flow
```
1. User signs up (email, password, name)
2. AUTOMATIC: Create "Personal" org
3. AUTOMATIC: Add user as admin of org
4. User automatically logged in
5. Redirected to dashboard
6. Later: Can invite others to their org
7. Later: Can be invited to other orgs
```

#### Pros
- ✅ **Simplest implementation** (easiest for MVP)
- ✅ **Zero friction** - no intermediate screens
- ✅ **Immediate access** - can use app right away
- ✅ **No email required** - works offline
- ✅ **Self-contained** - doesn't need admin setup
- ✅ **Flexible** - can later join other orgs
- ✅ **No code required** - user doesn't need anything

#### Cons
- ❌ Each user gets own org (may increase org count)
- ❌ Single-user orgs if no collaboration
- ❌ Can be cleaned up later (archive empty orgs)

#### Implementation Effort
- **Backend:** 1-2 hours (org auto-creation in register)
- **Frontend:** 0 hours (no changes needed)
- **Total:** 1-2 hours ⭐ **FASTEST**

#### Code Example (Backend)

```go
// Backend: handlers/auth.go - Register with auto-org
func Register(c *fiber.Ctx) error {
    var req struct {
        Email    string `json:"email"`
        Password string `json:"password"`
        Name     string `json:"name"`
    }
    c.BodyParser(&req)

    // Hash password
    hashedPassword, err := utils.HashPassword(req.Password)
    if err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid password"})
    }

    // Create user
    newUser := models.User{
        ID:       utils.GenerateUserID(),
        Email:    req.Email,
        Name:     req.Name,
        Password: hashedPassword,
        Role:     "requester",
        Active:   true,
    }
    config.DB.Create(&newUser)

    // AUTO-CREATE personal organization
    personalOrg := models.Organization{
        ID:   utils.GenerateOrgID(),
        Name: fmt.Sprintf("%s's Organization", req.Name),
        Slug: strings.ToLower(strings.ReplaceAll(req.Email, "@", "-")),
        Type: "personal", // Mark as personal org
    }
    config.DB.Create(&personalOrg)

    // Add user as admin of personal org
    member := models.OrganizationMember{
        ID:             utils.GenerateID(),
        OrganizationID: personalOrg.ID,
        UserID:         newUser.ID,
        Role:           "admin", // User is admin of their org
        Active:         true,
        JoinedAt:       time.Now(),
    }
    config.DB.Create(&member)

    // Set as current organization
    newUser.CurrentOrganizationID = &personalOrg.ID
    config.DB.Save(&newUser)

    // Generate token WITH org context
    token, _ := utils.GenerateToken(
        newUser.ID,
        newUser.Email,
        newUser.Name,
        newUser.Role,
        &personalOrg.ID,
    )

    return c.JSON(types.SuccessResponse{
        Success: true,
        Message: "User registered and organization created",
        Token:   token,
        User:    newUser,
    })
}
```

---

## Recommendation: Scenario C (Auto-Create Personal Organization)

### Why Scenario C?

**MVP Best Practice:**
- Lowest implementation cost (1-2 hours)
- Zero user friction
- No intermediate setup screens
- Immediate app access
- Can upgrade to other scenarios later

**User Journey:**
```
Sign up → Personal org created → Dashboard
(no intermediate steps)
```

**Future Enhancement Path:**
```
Phase 1: MVP with Scenario C (auto-personal org)
Phase 2: Add Scenario B (email invitations)
Phase 3: Add Scenario A (org codes)
Phase 4: Mix and match based on org preferences
```

---

## Hybrid Approach (Recommended Long-Term)

### Best of All Worlds

```
User Registration Page:
├─ Simple signup form (email, password, name)
├─ Auto-creates personal organization
└─ User redirected to dashboard

Organization Settings (after login):
├─ "Invite Members" → Send email invites (Scenario B)
├─ "Join Organization" → Enter code (Scenario A)
└─ "My Organization" → See personal org details

Admin Controls:
├─ Generate shareable codes (Scenario A)
├─ Send bulk invitations (Scenario B)
└─ Auto-created personal orgs (Scenario C)
```

---

## Permission Gaps & Solutions

### Current Gap
Using roles directly instead of permissions:
```go
if tenant.UserRole != "requester" {
    // User cannot create requisition
}
```

### Solution (Phase 2)
Permission-based access:
```go
if !permissions.HasPermission(userRole, "create_requisition") {
    // User cannot create requisition
}
```

### How Permissions Work with Onboarding

```
New User Signup
    ↓
Role assigned: "requester" (default)
    ↓
Permission mapping:
├─ Requester can: create_requisition, view_requisition
├─ Approver can: approve_requisition, view_requisition
├─ Finance can: create_budget, manage_vendors
├─ Admin can: * (all permissions)
    ↓
User lands in dashboard
    ↓
UI checks permissions:
├─ Show "Create Requisition" button? ✓ (requester)
├─ Show "Approve" button? ✗ (not approver)
├─ Show member management? ✗ (not admin)
    ↓
User can only access allowed features
```

---

## Implementation Roadmap

### Phase 1: ✅ COMPLETE
- [x] Backend authentication working
- [x] Password verification enabled
- [x] RBAC architecture designed

### Phase 2: TODO (Next) - Scenario C (Auto-Personal Org)
**Time:** 1-2 hours
- [ ] Implement auto-org creation in register endpoint
- [ ] Test registration creates org and adds user as admin
- [ ] Implement permission-based access control
- [ ] Update frontend signup form
- [ ] Verify user can use app immediately after signup

### Phase 3: TODO - Scenario B (Email Invitations)
**Time:** 7-10 hours
- [ ] Create invitation model and endpoints
- [ ] Implement token generation and validation
- [ ] Update register to accept invite tokens
- [ ] Integrate email service (SendGrid, SES, etc.)
- [ ] Build admin invitation UI
- [ ] Test full invitation flow

### Phase 4: TODO - Scenario A (Organization Codes)
**Time:** 3-5 hours
- [ ] Create organization code model
- [ ] Implement code generation/expiration
- [ ] Update register to accept codes
- [ ] Build code management admin UI
- [ ] Test code redemption

---

## Decision Tree

### Which scenario should we use?

```
                   New User Signup
                         |
         ________________|________________
         |               |               |
      [MVP?]          [Admin        [Immediate
                   Controlled?]      Access?]
         |               |               |
         ↓               ↓               ↓
      [YES]          [YES]          [YES]
         |               |               |
      Scenario C     Scenario B     Scenario C
      (Personal)    (Invitations)   (Personal)
```

**For MVP:** Use Scenario C
**For Enterprise:** Use Scenario B + Admin integration
**For Hybrid:** Use all three, let org admins choose

---

## Security Considerations

### Scenario C (Personal Org)
- ✅ **Most Secure** - User fully controls their org
- ✅ No code sharing risks
- ✅ No invitation token interception
- ✅ Automatic isolation

### Scenario B (Email Invitations)
- ⚠️ Token expiration required (7 days)
- ⚠️ Token validation on every request
- ⚠️ Email verification recommended
- ✅ Admin-controlled access

### Scenario A (Organization Codes)
- ⚠️ Code sharing possible
- ⚠️ Code rotation recommended
- ✅ Usage tracking helps
- ✅ Simple to implement

---

## Testing Strategy

### For Scenario C (Personal Org)

```typescript
// Test: User can signup and get immediate access
describe("User Registration - Scenario C", () => {
  it("should create personal org on signup", async () => {
    const result = await registerAction({
      email: "test@example.com",
      password: "Password123!",
      name: "Test User"
    });

    expect(result.success).toBe(true);
    expect(result.token).toBeDefined();
    // Should have org context in token
    expect(decodedToken.currentOrgId).toBeDefined();
  });

  it("should add user as admin of personal org", async () => {
    // After signup, user should have admin role in their org
    const perms = await getPermissions();
    expect(perms).toContain("manage_members");
  });

  it("should redirect to dashboard immediately", async () => {
    // No intermediate org selection screens
    const response = await signup();
    expect(response.redirectUrl).toBe("/home");
  });
});
```

---

## Conclusion

### For MVP Phase 2 Implementation:
**Use Scenario C (Auto-Create Personal Organization)**

**Reason:**
- Lowest implementation effort (1-2 hours)
- Zero user friction (no intermediate screens)
- Aligns with "permission-based" access control
- Can be enhanced later with other scenarios

**Next Steps:**
1. Implement Scenario C in Phase 2
2. Add permission-based access control
3. Test full signup → dashboard flow
4. Later: Add email invitations (Scenario B)
5. Later: Add organization codes (Scenario A)

---

*For more details on RBAC architecture, see [RBAC-AND-ORGANIZATION-ARCHITECTURE.md](./RBAC-AND-ORGANIZATION-ARCHITECTURE.md)*

*For implementation details, see [IMPLEMENTATION-SUMMARY.md](./IMPLEMENTATION-SUMMARY.md)*
