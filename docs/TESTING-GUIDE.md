# Testing Guide

## Manual Testing Checklist

### Workflow Testing

#### Requisition Workflow
- [ ] Navigate to requisition list
- [ ] Click on a requisition
- [ ] View details
- [ ] Go to approval page
- [ ] Draw signature
- [ ] Click Approve
- [ ] See success notification
- [ ] Status updated in list

#### Budget Workflow
- [ ] Navigate to budget list
- [ ] Click on a budget
- [ ] View details
- [ ] Go to approval page
- [ ] Draw signature
- [ ] Click Reject with reason
- [ ] See success notification

#### Purchase Order Workflow
- [ ] Navigate to PO
- [ ] View vendor information
- [ ] View cost breakdown
- [ ] See stage progress
- [ ] Complete approval flow

#### Payment Voucher Workflow
- [ ] Navigate to PV
- [ ] View invoice details
- [ ] View payment method
- [ ] View GL codes
- [ ] Complete approval flow

#### GRN Workflow
- [ ] Navigate to GRN
- [ ] View item matching
- [ ] See variances
- [ ] View damage tracking
- [ ] Complete confirmation
- [ ] Note 2-stage workflow

### Bulk Operations Testing

- [ ] Select multiple items with checkboxes
- [ ] See selection counter update
- [ ] Click Approve All
- [ ] See dialog with count
- [ ] Add remarks
- [ ] Click Approve
- [ ] See success notification
- [ ] Items marked as approved

- [ ] Select items
- [ ] Click Reject All
- [ ] Try to submit without reason
- [ ] See validation error
- [ ] Add reason
- [ ] Submit successfully

- [ ] Select items
- [ ] Click Reassign All
- [ ] Select approver from dropdown
- [ ] Add reason
- [ ] Submit successfully

### Analytics Testing

- [ ] Go to Admin Reports
- [ ] Click Analytics tab
- [ ] See 5 metric cards
- [ ] Scroll to Trends section
- [ ] See 7-day data
- [ ] Scroll to Distribution
- [ ] See all document types
- [ ] Scroll to Performance
- [ ] See stage metrics
- [ ] Scroll to Bottleneck
- [ ] See recommendations
- [ ] Click Refresh button
- [ ] See loading state
- [ ] Metrics update
- [ ] Click Export button
- [ ] CSV file downloads

### Data Persistence Testing

- [ ] Complete an approval
- [ ] Press F5 (refresh page)
- [ ] See task still approved
- [ ] Open DevTools (F12)
- [ ] Go to Application → Local Storage
- [ ] Look for `approval_tasks_v1`
- [ ] Expand and see JSON
- [ ] Close and reopen browser
- [ ] Data still persists

### Error Handling Testing

- [ ] Try to approve without drawing signature
- [ ] See error message
- [ ] Try to reject without reason
- [ ] See validation error
- [ ] Try to reassign without selecting approver
- [ ] See error message
- [ ] Complete action successfully

### UI/UX Testing

- [ ] Check responsive design on mobile
- [ ] Check responsive design on tablet
- [ ] Check responsive design on desktop
- [ ] All buttons clickable
- [ ] All forms usable
- [ ] Colors display correctly
- [ ] Icons display correctly
- [ ] Loading states visible
- [ ] Animations smooth

### Search & Filter Testing

- [ ] Search for task by number
- [ ] Filter by status (Pending)
- [ ] Filter by status (Approved)
- [ ] Filter by status (Rejected)
- [ ] Filter by priority (High)
- [ ] Filter by priority (Medium)
- [ ] Filter by priority (Low)
- [ ] Sort by date
- [ ] Combinations work

## Performance Testing

### Page Load Times
- [ ] Home page: <2 seconds
- [ ] Tasks page: <2 seconds
- [ ] Detail pages: <3 seconds
- [ ] Analytics: <3 seconds
- [ ] Admin Reports: <2 seconds

### Operation Times
- [ ] Single approval: <1 second
- [ ] Bulk approve (10 items): <2 seconds
- [ ] Signature drawing: Smooth
- [ ] Form submission: Quick response

### Data Operations
- [ ] localStorage read: Instant
- [ ] localStorage write: Instant
- [ ] Cache invalidation: Fast
- [ ] Query execution: <100ms (after Phase 12)

## Browser Testing

### Chrome
- [ ] All features work
- [ ] No console errors
- [ ] localStorage visible

### Firefox
- [ ] All features work
- [ ] No console errors
- [ ] localStorage visible

### Safari
- [ ] All features work
- [ ] No console errors
- [ ] localStorage visible

### Edge
- [ ] All features work
- [ ] No console errors
- [ ] localStorage visible

## Device Testing

### Desktop (1920x1080)
- [ ] All UI visible
- [ ] No horizontal scroll
- [ ] Forms usable

### Tablet (768x1024)
- [ ] UI responsive
- [ ] Forms usable
- [ ] Navigation accessible

### Mobile (375x667)
- [ ] UI responsive
- [ ] Touch-friendly
- [ ] Readable text

## Accessibility Testing

- [ ] Tab navigation works
- [ ] Keyboard-only navigation possible
- [ ] Color contrast sufficient
- [ ] Form labels present
- [ ] Error messages clear
- [ ] Screen reader compatible (basic)

## Data Integrity Testing

- [ ] No data duplication
- [ ] No data loss on refresh
- [ ] No data loss on browser restart
- [ ] Signatures preserved
- [ ] Comments preserved
- [ ] Timestamps accurate
- [ ] Counts accurate

## Build Testing

```bash
# Check TypeScript
npm run type-check

# Build for production
npm run build

# Should complete with 0 new errors
```

## Testing Scenarios

### Happy Path: Complete Approval
1. Start: 3 pending tasks
2. Select all 3 tasks
3. Click Approve All
4. Add remarks
5. Submit
6. End: 3 approved tasks
7. Analytics updated

### Error Path: Missing Reason
1. Select tasks
2. Click Reject All
3. Try submit without reason
4. See validation error
5. Add reason
6. Submit successfully

### Data Persistence Path
1. Approve a task
2. Refresh page
3. Task still approved
4. Close browser
5. Reopen browser
6. Task still approved

## Test Results Template

```
Date: ____
Tester: ____

✅ Approval Workflows: PASS/FAIL
✅ Bulk Operations: PASS/FAIL
✅ Analytics: PASS/FAIL
✅ Data Persistence: PASS/FAIL
✅ Error Handling: PASS/FAIL
✅ UI/UX: PASS/FAIL
✅ Performance: PASS/FAIL

Notes:
____

Issues Found:
____
```

## Known Issues

- None currently

## Sign Off

When all tests pass, system is ready for:
- [ ] Stakeholder demo
- [ ] User acceptance testing
- [ ] Phase 12 development
