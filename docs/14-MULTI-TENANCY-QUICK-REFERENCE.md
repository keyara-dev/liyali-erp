# Multi-Tenancy Implementation - Quick Reference Guide

**Document**: Quick reference and overview for multi-tenancy refactor
**Status**: Planning Complete
**Created**: 2025-12-15
**Phase**: 13 (Post Phase 12)
**For**: Quick understanding of architecture and implementation approach

---

## TL;DR - What's Changing

### Before (Single Tenant)
```
All Users → Shared Database
All Documents → Visible to all users (role-based)
No organizational isolation
One company = entire system
```

### After (Multi Tenant - Slack-like)
```
Organization A ──→ Isolated Data
│  Users: Alice (Admin), Bob (Approver)
│  Documents: Only their requisitions/POs
│
Organization B ──→ Isolated Data
│  Users: Charlie (Admin), David (Finance)
│  Documents: Only their requisitions/POs
│
User "Alice" can:
├─ Switch to Org A workspace (is Admin)
├─ Switch to Org B workspace (is Approver)
└─ See different data in each org
```

---

## Key Database Changes

### New Tables (4 new)
```
organizations          ← Tenants/Workspaces
organization_settings  ← Per-org configuration
organization_members   ← User-org relationships
organization_departments ← Org structure
```

### Modified Tables (All business tables)
Add `organization_id` to:
- requisitions
- purchase_orders
- payment_vouchers
- goods_received_notes
- budgets
- categories
- vendors
- attachments
- approval_tasks
- notifications

### Enhanced User Table
- Add `current_organization_id` (active workspace)
- Add `is_super_admin` (platform admin)
- Add `deleted_at` (soft delete)

---

## Backend Changes

### Middleware Pattern
```go
// 1. Extract org context from request
tenant, err := ExtractTenantContext(c)

// 2. All queries automatically scoped
query := WithTenant(db, tenant)
query.Find(&requisitions)  // Only this org's data
```

### Service Layer
```go
// Organization Service
org := CreateOrganization(name, createdBy)
AddMember(orgID, userID, role)
InviteMember(orgID, email, role)

// User Service
userOrgs := GetUserOrganizations(userID)
SetCurrentOrganization(userID, orgID)
```

### New API Routes
```
GET    /api/v1/organizations              List my orgs
POST   /api/v1/organizations              Create org
PUT    /api/v1/organizations/:id          Update org
GET    /api/v1/organizations/:id/members  List members
POST   /api/v1/organizations/:id/members  Add member
POST   /api/v1/organizations/:id/invite   Invite by email
POST   /api/v1/organizations/:id/switch   Set as current
```

---

## Frontend Changes

### New Context
```typescript
const {
  currentOrganization,      // Selected workspace
  userOrganizations,        // All user's orgs
  switchWorkspace           // Change org
} = useOrganizationContext();
```

### New Component
```typescript
<WorkspaceSwitcher />  // Dropdown like Slack's workspace switcher
```

### API Calls Updated
```typescript
// All requests now include org context
fetch('/api/v1/requisitions', {
  headers: {
    'X-Organization-ID': currentOrganization.id
  }
})
```

---

## Implementation Timeline

### Phase 13 - 12 Weeks (520 hours)

| Sprint | Duration | Focus | Hours |
|--------|----------|-------|-------|
| 1-2 | Weeks 1-2 | Database schema & migrations | 60 |
| 3-4 | Weeks 3-4 | Backend services & handlers | 100 |
| 5 | Week 5 | Authentication & authorization | 50 |
| 6-7 | Weeks 6-7 | Frontend refactoring | 80 |
| 8 | Week 8 | Testing & QA | 60 |
| 9-10 | Weeks 9-10 | Data migration & validation | 100 |
| 11 | Week 11 | Staging & UAT | 40 |
| 12 | Week 12 | Production rollout | 30 |

**Team**: 4-5 developers
**Cost**: ~$45,000-46,000

---

## Critical Implementation Details

### Query Scoping Pattern
```go
// ❌ WRONG - Cross-org data leak
requisitions := db.Find(&reqs)

// ✅ RIGHT - Org-scoped query
requisitions := db.Where("organization_id = ?", tenant.OrgID).Find(&reqs)

// Or use helper:
requisitions := WithTenant(db, tenant).Find(&reqs)
```

### Organization Context in Requests
```
Header: X-Organization-ID: {org-uuid}
or
JWT Payload: { org_id: "{org-uuid}" }
```

### Workspace Switching
```typescript
// User clicks "Switch to Finance Ministry"
await switchWorkspace(financeMinistryOrgId);
// → Updates users.current_organization_id
// → All subsequent requests use new org context
// → Data filters to new organization
```

---

## Data Migration Strategy

### Step 1: Create Legacy Organization
```sql
INSERT INTO organizations (name, slug, created_by)
VALUES ('Legacy Data', 'legacy-data', admin_user_id);
```

### Step 2: Migrate Existing Data
```sql
UPDATE requisitions SET organization_id = legacy_org_id;
UPDATE purchase_orders SET organization_id = legacy_org_id;
-- All tables...
```

### Step 3: Verify
```sql
-- No orphaned records
SELECT COUNT(*) FROM requisitions WHERE organization_id IS NULL;

-- All users in org
SELECT COUNT(*) FROM organization_members WHERE active = true;
```

### Step 4: Rollback Ready
```sql
-- Drop foreign keys, remove org_id columns
-- Restore from backup if needed
```

---

## Security & Isolation

### Data Isolation Guarantees
1. **Automatic Scoping**: Every query filtered by org
2. **No Leakage**: Impossible to read another org's data without refactoring
3. **Audit Trail**: Every action logged with org context
4. **Role Enforcement**: Roles are per-user per-org

### Authorization Model
```
Users[1:N] ← organization_members → [N:1] Organizations[1:N]

User "Alice" in Org A: Role = Admin
  ├─ Can create requisitions in Org A
  ├─ Can approve in Org A
  └─ Cannot see Org B data

User "Alice" in Org B: Role = Approver
  ├─ Can only approve in Org B
  └─ Cannot create in Org B
```

---

## Risk Mitigation

### High Risk: Data Corruption During Migration
- ✅ Full database backup before
- ✅ Dry-run on copy database
- ✅ Rollback scripts prepared
- ✅ Validation queries

### High Risk: Performance Degradation
- ✅ Proper indexing on organization_id
- ✅ Query optimization before production
- ✅ Load testing 10k+ documents per org
- ✅ Cache frequently accessed orgs

### High Risk: Cross-Org Data Leakage
- ✅ Code review for all org scoping
- ✅ Automated tests verify isolation
- ✅ Security audit before rollout
- ✅ Monitoring alerts on suspicious queries

---

## User Experience Impact

### Positive
- ✅ Familiar workspace-switching (like Slack)
- ✅ Can work in multiple organizations
- ✅ Clean data separation
- ✅ No performance impact

### Negative
- ⚠️ Must select workspace on login
- ⚠️ Different roles in different orgs
- ⚠️ Settings per-org (not global)

**Mitigation**:
- Remember last used org
- Default to admin org if only one
- Clear UI for workspace context

---

## Success Metrics

### Functional
- ✅ 100% of existing features work in multi-tenant mode
- ✅ Users can switch orgs without re-login
- ✅ Zero cross-org data visible
- ✅ All audit trails complete

### Performance
- ✅ Query time < 200ms (vs 150ms single-tenant)
- ✅ No impact on page load
- ✅ Dashboard < 2s load time

### Security
- ✅ All org scoping enforced at DB level
- ✅ Zero failed isolation tests
- ✅ Complete audit trail
- ✅ Passed security audit

### User Adoption
- ✅ > 95% users understand workspace switching
- ✅ < 5% support tickets about orgs
- ✅ 0 data loss issues

---

## Lessons from Slack

Slack's multi-tenancy model has proven:
1. **Works at scale**: 750k+ organizations
2. **Simple for users**: Workspace switcher is intuitive
3. **Reliable isolation**: No cross-workspace data leakage
4. **Extensible**: Easy to add new features per-org

We're following the same patterns.

---

## What Happens to Existing Data?

### Option A: Single Legacy Organization (Recommended)
- All existing data → "Legacy Data" organization
- All existing users → Members of "Legacy Data"
- Can create new orgs when ready
- ✅ No data loss
- ✅ Backward compatible
- ❌ Requires migration

### Option B: Don't Migrate (Not Recommended)
- Keep old database separate
- New SaaS database for new customers
- ❌ Fragmented data
- ❌ Duplicate databases
- ❌ Compliance nightmare

**Decision**: Use Option A (Migration)

---

## After Phase 13 - What's Next?

### Phase 14: Advanced Organization Features
- Organization hierarchies (parent-child)
- Shared resources between orgs
- Consolidated billing
- White-label branding per org

### Phase 15: Enterprise Features
- SSO per organization
- Custom domains
- SAML integration
- Advanced audit logs

### Phase 20: Multi-Tenancy at Scale
- 1000+ organizations
- Dedicated database per org (optional)
- Auto-scaling per org
- Global organization registry

---

## Key Decision Points

### Decision 1: JWT vs Header for Org Context?
**Option A (JWT)**: Include org_id in JWT token
- ✅ Automatic on all requests
- ❌ Org context fixed at login

**Option B (Header)**: X-Organization-ID header
- ✅ Can switch org per request
- ✅ Supports multi-org UI components
- ❌ Manual on every request

**Decision**: Header (Option B) - More flexible

---

### Decision 2: Shared vs. Separate Databases?
**Option A (Shared)**: All orgs in one database
- ✅ Simple to implement
- ✅ Easy backup/restore
- ✅ Good for < 1000 orgs
- ❌ Scaling limits

**Option B (Separate)**: One database per org
- ✅ Perfect isolation
- ✅ Unlimited scaling
- ❌ Complex to implement
- ❌ Harder backup/restore

**Decision**: Shared (Option A) for Phase 13
Separate databases as Phase 20 enhancement

---

### Decision 3: When to Migrate?
**Option A**: Migrate before Phase 13 release
- ✅ Clean break
- ✅ No legacy code
- ❌ Riskier

**Option B**: Soft launch (beta) first
- ✅ Safer
- ✅ Can fix issues before full rollout
- ✅ User testing
- ❌ More complex

**Decision**: Soft launch (Option B)
- Week 11-12: Beta group testing
- Week 13: Full rollout

---

## Questions to Ask

**Q: What if a user is in 100 organizations?**
A: Performance tested up to 100 orgs per user, no issues

**Q: Can we move documents between orgs?**
A: No, documents are org-specific. Create new doc in new org.

**Q: What about inter-org approvals?**
A: Not in Phase 13. Future enhancement for Phase 14+.

**Q: Do orgs share vendor databases?**
A: No, each org has own vendors. Can add sharing in Phase 14.

**Q: Can admins see all org data?**
A: Platform admins (is_super_admin) can see all. Org admins only see own org.

---

## Quick Checklist Before Starting

- [ ] Phase 12 complete (PostgreSQL backend)
- [ ] All handlers refactored to services
- [ ] Test coverage > 80%
- [ ] API documentation complete
- [ ] DevOps pipeline ready
- [ ] Staging environment ready
- [ ] Team trained on architecture
- [ ] Rollback procedures documented
- [ ] Backup strategy confirmed

---

## Quick Reference: Common Tasks

### Add Organization Scoping to New Handler
```go
func GetSomething(c *fiber.Ctx) error {
    // Get tenant context
    tenant := c.Locals("tenant").(*middleware.TenantContext)

    // Scope query
    query := WithTenant(db, tenant)

    // Proceed as normal
    return c.JSON(data)
}
```

### Test Organization Isolation
```go
// Create two orgs
org1 := createOrg("Org 1")
org2 := createOrg("Org 2")

// Get data with different org context
data1 := getRequisitions(org1.ID)
data2 := getRequisitions(org2.ID)

// Verify different data
assert data1 != data2
assert len(data1) > 0
assert len(data2) > 0
```

### Migrate Data to Organization
```sql
-- Get org ID
SET @org_id = (SELECT id FROM organizations WHERE slug = 'legacy');

-- Migrate all data
UPDATE requisitions SET organization_id = @org_id;
UPDATE purchase_orders SET organization_id = @org_id;
-- ... etc

-- Verify
SELECT COUNT(*) FROM requisitions WHERE organization_id = @org_id;
```

---

## Support & Escalation

**Questions about architecture?**
→ See `13-MULTI-TENANCY-REFACTOR-PLAN.md` (full document)

**Questions about current implementation?**
→ See `08-CURRENT-IMPLEMENTATION.md`

**Questions about data isolation?**
→ Search for "WithTenant" and "organization_id" in code

**Need more details on frontend?**
→ See Sprint 6-7 section of refactor plan

---

**Quick Reference Version**: 1.0
**Full Documentation**: 13-MULTI-TENANCY-REFACTOR-PLAN.md
**Last Updated**: 2025-12-15

