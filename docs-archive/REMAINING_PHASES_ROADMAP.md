# Remaining Phases Roadmap (Phase 6-12)

Complete implementation roadmap for finishing the custom workflow management system.

---

## Current Status

✅ **Phases 1-5 Complete**: 6,725+ lines of code and documentation
- Phase 1-4: Foundation (types, persistence, validation, orchestration)
- Phase 5: Server Actions & Notifications (2,225 lines)

🚀 **Ready for Phase 6**: All server infrastructure in place

---

## Phase 6: React Query Hooks Enhancement (1-2 days)

### Objectives
- Additional workflow query hooks
- Batch operations support
- Pagination utilities
- Cache management helpers

### Files to Create

**src/hooks/use-workflows.ts** (~400 lines)
```typescript
// Query hooks
useWorkflows()           // List with filters
useWorkflow()           // Single workflow
useWorkflowAssignment()  // Get assignment
useWorkflowDefault()     // Get default for entity type
usePendingApprovals()    // User's pending queue

// Mutation hooks
useCreateWorkflow()
useUpdateWorkflow()
useDeprecateWorkflow()
useAssignWorkflow()
useApproveStage()
useRejectStage()
useReassignStage()

// Advanced hooks
useWorkflowPolling()     // Auto-refetch
useInvalidateWorkflows()
```

**src/hooks/use-approval-flow.ts** (~350 lines)
```typescript
// Combined hooks for approval workflows
useApprovalModal()       // Handle modal state + submission
useReassignmentFlow()    // Permission check + reassign
useQuickApprove()        // 2-click approval
useApprovalHistory()     // Stage execution history
```

### Tasks

- [ ] Create `use-workflows.ts` with 8+ query hooks
- [ ] Create `use-approval-flow.ts` with combined hooks
- [ ] Add batch operation helpers
- [ ] Add pagination utilities
- [ ] Write unit tests for hooks
- [ ] Document hook usage
- [ ] Update index.ts exports

### Testing
- Mock React Query setup
- Test cache updates
- Test error scenarios
- Test polling behavior

---

## Phase 7: UI Components - Notifications (2-3 days)

### Objectives
- Notification display components
- Quick action interface
- Real-time updates
- Notification preferences UI

### Files to Create

**src/components/notifications/notification-bell.tsx** (~200 lines)
```typescript
// Bell icon with badge
// Shows unread count
// Dropdown trigger
// Real-time updates
```

**src/components/notifications/notification-dropdown.tsx** (~250 lines)
```typescript
// Dropdown showing recent notifications
// Scrollable list
// Quick action buttons
// "View All" link
// Empty state
```

**src/components/notifications/notification-item.tsx** (~180 lines)
```typescript
// Single notification display
// Icon + type
// Message
// Time ago
// Unread badge
// Quick action button
// Hover actions (delete, mark read)
```

**src/components/notifications/notification-action-modal.tsx** (~300 lines)
```typescript
// Modal for quick actions
// Shows entity summary
// Signature canvas
// Remarks field
// Approve/Reject buttons
// Form validation
```

**src/app/(private)/workflows/notifications/page.tsx** (~400 lines)
```typescript
// Full notifications page
// List of all notifications
// Filters: Type, Date, Status
// Pagination
// Bulk actions
// Search
// Settings link
```

**src/components/notifications/notification-preferences.tsx** (~250 lines)
```typescript
// User notification settings
// Email toggle
// Push toggle
// Notification type toggles
// Quiet hours
// Save button
```

### Tasks

- [ ] Create notification-bell component
- [ ] Create notification-dropdown component
- [ ] Create notification-item component
- [ ] Create notification-action-modal component
- [ ] Create notifications page
- [ ] Create preferences component
- [ ] Add to app bar/header
- [ ] Test real-time updates
- [ ] Add animations/transitions
- [ ] Mobile responsiveness

### Integration Points

- Add bell to [src/components/layout/header.tsx](src/components/layout/header.tsx)
- Add notifications page route
- Add preferences page route
- Connect to `useNotificationBell()` hook
- Connect to `useNotificationPolling()` hook

---

## Phase 8: UI Components - Workflows (2-3 days)

### Objectives
- Workflow display and selection
- Approval flow visualization
- Reassignment interface
- Stage information display

### Files to Create

**src/components/workflows/workflow-selector.tsx** (~250 lines)
```typescript
// Dropdown to select workflow for entity
// Shows default and custom workflows
// Preview workflow stages
// Validation feedback
```

**src/components/workflows/approval-flow-display.tsx** (~300 lines)
```typescript
// Visual representation of workflow stages
// Current stage highlight
// Approver information
// Stage status badges
// Timeline view
```

**src/components/workflows/stage-execution-display.tsx** (~200 lines)
```typescript
// Details of current stage
// Approver name and role
// Stage name and description
// Requirements (signature, comments)
// Action buttons
```

**src/components/workflows/reassignment-modal.tsx** (~280 lines)
```typescript
// Modal for reassigning approval task
// Current approver display
// User selector for new approver
// Reason field
// Permission check feedback
// Reassign button with loading
```

**src/components/workflows/approval-history.tsx** (~250 lines)
```typescript
// Timeline of all approvals/rejections
// Approver info
// Timestamp
// Comments/remarks
// Signature verification badge
// Reassignment info if applicable
```

**src/components/workflows/approval-action-panel.tsx** (~300 lines)
```typescript
// Action buttons for current approver
// Approve button → Opens approval modal
// Reject button → Opens rejection form
// Reassign button → Opens reassignment modal
// Loading states
// Permission checks
```

### Tasks

- [ ] Create workflow-selector component
- [ ] Create approval-flow-display component
- [ ] Create stage-execution-display component
- [ ] Create reassignment-modal component
- [ ] Create approval-history component
- [ ] Create/update approval-action-panel component
- [ ] Add to requisition/budget detail pages
- [ ] Add to approval view pages
- [ ] Connect to workflow hooks
- [ ] Add form validation

### Integration Points

- Add to [src/app/(private)/workflows/requisitions/[id]/page.tsx](src/app/(private)/workflows/requisitions/[id]/page.tsx)
- Add to budget detail page
- Add to dashboard approval view
- Connect to `useApprovalFlow()` hook
- Connect to `useReassignmentFlow()` hook

---

## Phase 9: Integration & Testing (2 days)

### Objectives
- Connect all components and hooks
- E2E workflow scenarios
- Error handling and fallbacks
- Performance optimization

### Integration Tasks

**Requisition Flow:**
- [ ] Select workflow when creating requisition
- [ ] Show approval flow on requisition detail
- [ ] Enable approve/reject actions
- [ ] Enable reassignment from detail view
- [ ] Show notifications on approval events
- [ ] Redirect on workflow complete

**Budget Flow:**
- [ ] Repeat requisition flow for budgets
- [ ] Support same workflows or different workflows

**Dashboard Integration:**
- [ ] Show pending approvals dashboard
- [ ] Quick stats (pending count, avg time, etc.)
- [ ] Reassignment interface
- [ ] Notification bell in top nav

**Admin Dashboard:**
- [ ] Workflow management page
- [ ] Workflow usage statistics
- [ ] Pending approvals overview
- [ ] Reassignment history
- [ ] System health metrics

### Testing Tasks

- [ ] Create workflow + assign to requisition
- [ ] Approve through all stages
- [ ] Reject and return to draft
- [ ] Reassign approval task
- [ ] Verify notification creation
- [ ] Test quick action approval
- [ ] Test reassignment audit trail
- [ ] Test permission denials
- [ ] Test error scenarios
- [ ] Test with multiple approvers

### Performance Optimization

- [ ] Pagination in notification lists
- [ ] Pagination in workflow lists
- [ ] Lazy loading components
- [ ] Image optimization
- [ ] Bundle size analysis
- [ ] Query cache optimization

---

## Phase 10: Admin Dashboard (1-2 days)

### Objectives
- Workflow management interface
- System monitoring
- Analytics and reporting
- User management for workflows

### Files to Create

**src/app/(private)/admin/workflows/page.tsx** (~400 lines)
```typescript
// Workflow management dashboard
// Create new workflow button
// List of workflows with:
// - Name, description
// - Applicable entity types
// - Stage count
// - Usage count
// - Active status
// - Version number
// - Actions: View, Edit, Deprecate, View Versions
```

**src/app/(private)/admin/workflows/[id]/page.tsx** (~350 lines)
```typescript
// Workflow detail page
// Workflow definition view
// Stage visualization
// Edit workflow (creates new version)
// Usage statistics
// Version history
// Associated assignments
```

**src/app/(private)/admin/workflows/designer/page.tsx** (~500 lines)
```typescript
// Visual workflow designer
// Drag-and-drop stage creation
// Stage configuration:
// - Name, description
// - Approver type and assignment
// - Transitions (onApprove, onReject)
// - Requirements
// - Permissions
// Stage preview
// Validation feedback
// Save workflow
```

**src/app/(private)/admin/approvals/page.tsx** (~300 lines)
```typescript
// All pending approvals dashboard
// Filter by: User, Entity type, Stage
// Sort by: Created, Due date
// Bulk actions: Reassign multiple
// Each approval shows:
// - Entity info
// - Current stage
// - Current approver
// - Time elapsed
// - Reassign button
```

**src/app/(private)/admin/analytics/page.tsx** (~300 lines)
```typescript
// Workflow analytics dashboard
// Charts:
// - Approval times by workflow
// - Success/rejection rates
// - Approver workload distribution
// - Stage bottlenecks
// Tables:
// - Top slow workflows
// - Most reassigned approvals
// - Recent completions
```

### Tasks

- [ ] Create workflow management page
- [ ] Create workflow detail view
- [ ] Create workflow designer (visual builder)
- [ ] Create admin approvals dashboard
- [ ] Create analytics dashboard
- [ ] Add workflow creation form
- [ ] Add workflow editing form
- [ ] Add export functionality
- [ ] Add filtering/searching
- [ ] Add user management

### Authorization

- [ ] Only ADMIN can access admin pages
- [ ] Only ADMIN can create/edit workflows
- [ ] Only ADMIN can view all approvals
- [ ] Track all admin actions

---

## Phase 11: Documentation & User Guides (1 day)

### Objectives
- Complete system documentation
- User guides
- Admin guides
- Developer guides

### Documents to Create

**User Guides:**
- [ ] How to submit for approval
- [ ] How to approve/reject
- [ ] How to reassign
- [ ] How to check notification history
- [ ] How to manage preferences

**Admin Guides:**
- [ ] How to create workflows
- [ ] How to configure stages
- [ ] How to set defaults
- [ ] How to monitor approvals
- [ ] How to analyze bottlenecks

**Developer Guides:**
- [ ] API reference
- [ ] Hook usage
- [ ] Component integration
- [ ] Custom workflow examples
- [ ] Testing guide

**System Documentation:**
- [ ] Architecture diagram
- [ ] Data flow diagrams
- [ ] Database schema
- [ ] API endpoints
- [ ] Error codes

---

## Phase 12: Testing, Polish & Deployment (1-2 days)

### Objectives
- Comprehensive testing
- Bug fixes
- Performance optimization
- Deployment preparation

### Testing Tasks

**Unit Tests:**
- [ ] Workflow validation rules
- [ ] Stage progression logic
- [ ] Permission checks
- [ ] Notification creation

**Integration Tests:**
- [ ] Complete approval flows
- [ ] Reassignment workflows
- [ ] Notification triggers
- [ ] Hook behavior

**E2E Tests:**
- [ ] Create requisition → Approve → Complete
- [ ] Create workflow → Use in entity → Verify
- [ ] Reassign approval task
- [ ] Reject and return to draft

**Performance Tests:**
- [ ] Load time under 2s
- [ ] Database queries < 100ms
- [ ] Notification polling not wasteful
- [ ] Bundle size < 500KB added

### Polish Tasks

- [ ] UI consistency across pages
- [ ] Error messages user-friendly
- [ ] Loading states clear
- [ ] Empty states designed
- [ ] Mobile responsive
- [ ] Accessibility checks
- [ ] Browser compatibility

### Deployment Preparation

- [ ] Environment variables documented
- [ ] Database migration scripts
- [ ] Migration from in-memory to PostgreSQL
- [ ] Seed production data
- [ ] Backup procedures
- [ ] Monitoring setup
- [ ] Logging setup
- [ ] Performance monitoring

---

## Database Migration (During Phase 12)

### Current: In-Memory Maps
```typescript
const notificationStore = new Map<string, Notification>();
const preferencesStore = new Map<string, NotificationPreferences>();
```

### Target: PostgreSQL Tables

**notifications table:**
```sql
CREATE TABLE notifications (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES users(id),
  type VARCHAR(50) NOT NULL,
  title VARCHAR(255) NOT NULL,
  message TEXT,
  entity_id VARCHAR(255),
  entity_type VARCHAR(50),
  entity_number VARCHAR(255),
  related_user_id UUID,
  is_read BOOLEAN DEFAULT FALSE,
  read_at TIMESTAMP,
  action_taken BOOLEAN DEFAULT FALSE,
  action_taken_at TIMESTAMP,
  quick_action JSONB,
  importance VARCHAR(20) DEFAULT 'MEDIUM',
  created_at TIMESTAMP DEFAULT NOW(),
  expires_at TIMESTAMP,
  INDEX(user_id, created_at DESC),
  INDEX(user_id, is_read)
);

CREATE TABLE notification_preferences (
  user_id UUID PRIMARY KEY REFERENCES users(id),
  email_notifications BOOLEAN DEFAULT FALSE,
  push_notifications BOOLEAN DEFAULT TRUE,
  in_app_notifications BOOLEAN DEFAULT TRUE,
  notify_on JSONB,
  quiet_hours JSONB,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);
```

**workflows table:**
```sql
CREATE TABLE custom_workflows (
  id UUID PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  description TEXT,
  version INTEGER NOT NULL DEFAULT 1,
  applicable_entity_types VARCHAR(255)[],
  is_template BOOLEAN DEFAULT TRUE,
  is_active BOOLEAN DEFAULT TRUE,
  stages JSONB NOT NULL,
  total_stages INTEGER NOT NULL,
  usage_count INTEGER DEFAULT 0,
  created_by UUID NOT NULL REFERENCES users(id),
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP,
  UNIQUE(id, version),
  INDEX(is_active, applicable_entity_types)
);
```

### Migration Steps

1. [ ] Create PostgreSQL tables
2. [ ] Migrate in-memory data to tables
3. [ ] Update persistence layer to use database
4. [ ] Add indexes for performance
5. [ ] Test all operations
6. [ ] Backup data
7. [ ] Deploy with backward compatibility
8. [ ] Monitor for issues

---

## Timeline Estimate

| Phase | Scope | Est. Time | Status |
|-------|-------|-----------|--------|
| 1-4 | Types, Persistence, Validation, Orchestration | Complete | ✅ Done |
| 5 | Server Actions & Notifications | Complete | ✅ Done |
| 6 | React Query Hooks | 1-2 days | 🚀 Next |
| 7 | Notification UI Components | 2-3 days | ⏳ Pending |
| 8 | Workflow UI Components | 2-3 days | ⏳ Pending |
| 9 | Integration & Testing | 2 days | ⏳ Pending |
| 10 | Admin Dashboard | 1-2 days | ⏳ Pending |
| 11 | Documentation | 1 day | ⏳ Pending |
| 12 | Testing, Polish, Deployment | 1-2 days | ⏳ Pending |
| **Total** | **Complete System** | **~13-17 days** | **In Progress** |

---

## Success Criteria

### Functional
- ✅ Users create custom workflows graphically
- ✅ Workflows are global and reusable
- ✅ Approvals follow workflow config
- ✅ Reassignments tracked with audit trail
- ✅ Notifications trigger on task assignment
- ✅ Quick actions work from notifications
- ✅ Complete workflow history visible

### Performance
- ✅ Workflow resolution < 100ms
- ✅ Notification polling < 500ms response
- ✅ Page loads < 2s
- ✅ Database queries < 100ms

### User Experience
- ✅ New workflow created in < 10 minutes
- ✅ Approval via notification in < 2 clicks
- ✅ Reassignment from dashboard UI
- ✅ Complete notification history available
- ✅ Mobile responsive
- ✅ Accessible to screen readers

### Reliability
- ✅ Zero data inconsistency
- ✅ All operations audit-logged
- ✅ No lost notifications
- ✅ Immutable audit trail
- ✅ Error handling for all paths
- ✅ Graceful degradation

### Security
- ✅ Only ADMIN creates workflows
- ✅ Role validation enforced
- ✅ Permission checks on actions
- ✅ Signature capture on approvals
- ✅ User isolation (can't see others' data)
- ✅ SQL injection protection
- ✅ XSS protection

---

## Notes for Implementation

### Architecture Patterns Used

1. **Server Actions Pattern** - All mutations use server actions
2. **React Query** - All data fetching with hooks
3. **Immutable Data** - No in-place modifications
4. **Event Triggers** - Actions trigger notifications
5. **Audit Trail** - Complete history of all changes
6. **Permission Model** - Role-based access control

### Common Gotchas to Avoid

1. Don't modify workflows in place (create new versions)
2. Don't skip validation before saving
3. Don't create notifications without triggered event
4. Don't allow reassignment of completed stages
5. Don't forget to update audit trail
6. Don't hardcode next stages (use workflow config)
7. Don't cache user permissions

### Code Quality Standards

- All functions documented with JSDoc
- TypeScript strict mode enabled
- 100% type coverage (no `any`)
- Error messages user-friendly
- Logging at key points
- Unit tests for logic
- E2E tests for flows

---

## Repository Structure After All Phases

```
src/
├── types/
│   ├── notifications.ts          (Phase 5)
│   ├── custom-workflow.ts        (Phase 1-4)
│   └── index.ts                  (updated)
│
├── lib/
│   ├── notification-persistence.ts    (Phase 5)
│   ├── workflow-persistence.ts        (Phase 1-4)
│   ├── workflow-validation.ts         (Phase 1-4)
│   └── workflow-resolution.ts         (Phase 1-4)
│
├── app/_actions/
│   ├── notifications.ts          (Phase 5)
│   └── workflows.ts              (Phase 5)
│
├── hooks/
│   ├── use-notifications.ts      (Phase 5)
│   ├── use-workflows.ts          (Phase 6)
│   └── use-approval-flow.ts      (Phase 6)
│
└── components/
    ├── notifications/            (Phase 7)
    │   ├── notification-bell.tsx
    │   ├── notification-dropdown.tsx
    │   ├── notification-item.tsx
    │   ├── notification-action-modal.tsx
    │   └── notification-preferences.tsx
    │
    ├── workflows/                (Phase 8)
    │   ├── workflow-selector.tsx
    │   ├── approval-flow-display.tsx
    │   ├── stage-execution-display.tsx
    │   ├── reassignment-modal.tsx
    │   └── approval-history.tsx
    │
    └── layout/
        └── header.tsx            (updated Phase 7)

app/(private)/
├── workflows/
│   ├── notifications/            (Phase 7)
│   │   └── page.tsx
│   ├── requisitions/             (updated Phase 9)
│   ├── budgets/                  (updated Phase 9)
│   └── dashboard/                (updated Phase 9)
│
├── admin/
│   ├── workflows/                (Phase 10)
│   │   ├── page.tsx
│   │   ├── [id]/
│   │   │   └── page.tsx
│   │   └── designer/
│   │       └── page.tsx
│   ├── approvals/                (Phase 10)
│   │   └── page.tsx
│   └── analytics/                (Phase 10)
        └── page.tsx
```

---

## Next Immediate Steps

1. ✅ Complete Phase 5 (DONE - 2,225 lines)
2. 🚀 Start Phase 6: Create `use-workflows.ts` and `use-approval-flow.ts`
3. ⏳ Phase 7: Build notification UI components
4. ⏳ Phase 8: Build workflow UI components
5. ⏳ Phase 9: Integration and testing
6. ⏳ Phases 10-12: Admin and deployment

---

**Current Status**: Phase 5 Complete ✅ | Phase 6-12 Roadmap Ready 🚀

All server infrastructure is in place. Ready to begin UI implementation in Phase 6.
