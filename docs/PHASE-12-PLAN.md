# Phase 12: Database Integration Plan

**Status**: Ready to Implement
**Duration**: 20-30 hours
**Complexity**: Medium

## Executive Summary

Phase 12 transitions from localStorage to PostgreSQL, adds real authentication via OAuth 2.0, implements email notifications, and enforces role-based permissions.

## Phase 12A: Database Setup

### Tasks
1. Create PostgreSQL database
2. Design and define schema
3. Create Prisma schema
4. Run migrations
5. Seed initial data

### Database Schema (8 tables)

```sql
users              -- User accounts
sessions           -- Session tokens
approval_tasks     -- Tasks to approve
approval_history   -- Approval records
documents          -- Base documents
audit_logs         -- Action tracking
notifications      -- System notifications
workflows          -- Workflow definitions
```

### Implementation
- Use Prisma ORM for type safety
- Add indexes for performance
- Set up foreign keys
- Configure constraints

## Phase 12B: Authentication

### OAuth 2.0 Setup

#### Entra ID (Microsoft 365)
1. Register app in Azure AD
2. Get client ID and secret
3. Configure redirect URI
4. Add to NextAuth.js

#### Google
1. Create OAuth app in Google Cloud
2. Get credentials
3. Configure consent screen
4. Add to NextAuth.js

#### GitHub (Optional)
1. Register OAuth app
2. Get credentials
3. Add to NextAuth.js

### Session Management
- 1-hour idle timeout
- 8-hour absolute timeout
- JWT-based sessions
- Secure session storage

### Implementation
- Install NextAuth.js v5
- Configure providers
- Implement session callbacks
- Protect routes with middleware

## Phase 12C: Data Migration

### Migration Strategy
1. Export all localStorage data
2. Transform to database format
3. Load into PostgreSQL
4. Verify all records
5. Test with real database

### Verification Checklist
- [ ] All tasks imported
- [ ] All approvals imported
- [ ] All notifications imported
- [ ] No data loss
- [ ] Data integrity verified
- [ ] Counts match

## Phase 12D: Server Actions Migration

### Changes Required
- Replace `approvalStore.*` with `db.*` (Prisma)
- Update error handling
- Add transaction support
- Update cache keys

### Affected Files (18+)
```
src/app/_actions/
├── approval-actions.ts       (6 functions)
├── bulk-operations.ts        (6 functions)
├── workflows.ts              (8 functions)
└── notifications.ts          (4 functions)
```

### Migration Pattern
```typescript
// Before (Phase 11)
const task = approvalStore.getTaskDetail(id)

// After (Phase 12)
const task = await db.approvalTask.findUnique({
  where: { id },
  include: { workflow: true }
})
```

## Phase 12E: Email Notifications

### SendGrid Setup
1. Create SendGrid account
2. Get API key
3. Verify sender email
4. Create email templates

### Email Templates (3 types)
1. **Task Assigned**: New approval waiting
2. **Task Approved**: Task was approved
3. **Task Rejected**: Task was rejected with reason

### Implementation
- Create email service module
- Implement 3 email functions
- Add error handling
- Test with real emails

## Phase 12F: Audit Logging

### What to Log
- Every approval action
- Every rejection
- Every reassignment
- System configuration changes
- Permission changes

### Logging Function
```typescript
await logAudit({
  userId: user.id,
  action: 'approve_task',
  entityId: taskId,
  timestamp: new Date(),
  changes: { status: 'approved' }
})
```

### Implementation
- Create audit_logs table
- Create logAudit() function
- Add logging to server actions
- Implement audit reporting

## Phase 12G: RBAC Implementation

### 7 Roles
1. **Department Manager** - Can approve at stage 1
2. **Finance Officer** - Can approve at stage 2
3. **Director** - Can approve at stage 3
4. **CFO** - Can approve at stage 3
5. **Compliance Officer** - Can view all, no approve
6. **Admin** - Full access
7. **User** - Can submit, view own

### Permission Matrix
```
                Submit  Approve  Reject  Reassign  View-All  Delete
Manager           ✓       ✓        ✓       ✓         ✗        ✗
Finance           ✓       ✓        ✓       ✓         ✗        ✗
Director          ✓       ✓        ✓       ✓         ✓        ✗
CFO               ✓       ✓        ✓       ✓         ✓        ✗
Compliance        ✓       ✗        ✗       ✗         ✓        ✗
Admin             ✓       ✓        ✓       ✓         ✓        ✓
User              ✓       ✗        ✗       ✗         ✗        ✗
```

### Implementation
- Create permissions table
- Create checkPermission() function
- Add to all server actions
- Add to route protection

## Phase 12H: Testing & Validation

### Unit Tests
- Server action inputs/outputs
- Permission checks
- Email formatting
- Audit logging

### Integration Tests
- Full approval workflow
- Database transactions
- Email delivery
- Permission enforcement

### E2E Tests
- User login flow
- Approval workflow
- Bulk operations
- Analytics

### Performance Tests
- Database query performance
- API response time
- Page load time
- Concurrent user load

### Security Tests
- Permission enforcement
- SQL injection protection
- Session security
- Token validation

## Phase 12I: Deployment

### Staging Environment
1. Deploy database
2. Configure authentication
3. Deploy application
4. Run full test suite
5. Get approval

### 4-Phase Rollout
1. **Phase 1 (10%)**: 10% of users
2. **Phase 2 (25%)**: 25% of users
3. **Phase 3 (50%)**: 50% of users
4. **Phase 4 (100%)**: All users

### Monitoring
- Error rates
- API response times
- User feedback
- System health

### Rollback Plan
- Feature flags to disable new code
- Keep old system available
- Database backup and restore plan
- Communication plan for issues

## Implementation Checklist

### Database Setup
- [ ] PostgreSQL database created
- [ ] Prisma schema written
- [ ] Migrations created
- [ ] Connection verified

### Authentication
- [ ] NextAuth.js installed
- [ ] OAuth providers configured
- [ ] Session management working
- [ ] Routes protected

### Data Migration
- [ ] Data exported from localStorage
- [ ] Migration script created
- [ ] Data loaded to PostgreSQL
- [ ] Verification complete

### Server Actions
- [ ] All 18+ actions migrated
- [ ] Testing complete
- [ ] Error handling updated
- [ ] Logging added

### Email Notifications
- [ ] SendGrid configured
- [ ] Templates created
- [ ] Email function working
- [ ] Testing complete

### Audit Logging
- [ ] Table created
- [ ] Logging function implemented
- [ ] Integrated in all actions
- [ ] Querying working

### RBAC
- [ ] Permission table created
- [ ] checkPermission() function
- [ ] Integrated in server actions
- [ ] Route protection added

### Testing
- [ ] Unit tests passing
- [ ] Integration tests passing
- [ ] E2E tests passing
- [ ] Performance acceptable
- [ ] Security verified

### Deployment
- [ ] Staging environment ready
- [ ] All tests passing
- [ ] Documentation updated
- [ ] Team trained
- [ ] Go/no-go decision

## Success Criteria

- ✅ PostgreSQL database operational
- ✅ OAuth 2.0 authentication working
- ✅ All data migrated successfully
- ✅ Email notifications sending
- ✅ Audit logging recording
- ✅ RBAC enforced
- ✅ All tests passing
- ✅ 0 new build errors
- ✅ Performance acceptable (<1s page load)

## Known Challenges

1. **Data Migration**: Ensure no data loss
2. **Auth Setup**: OAuth provider configuration
3. **Email Delivery**: Handle bounces/spam
4. **Performance**: Database slower than localStorage
5. **Testing**: Need production-like data volume

## Next Steps

1. Review this plan with team
2. Estimate detailed effort
3. Plan implementation timeline
4. Set up development environment
5. Begin Phase 12A (Database Setup)

---

**Ready to start Phase 12?**
Start with Phase 12A: Database Setup
