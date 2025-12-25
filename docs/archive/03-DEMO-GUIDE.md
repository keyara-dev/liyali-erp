# Demo Guide

## Demo Overview

Demonstrate the complete workflow approval system with real-time analytics.

**Duration**: 5-90 minutes (choose your session)

## Quick Demo (5 minutes)

1. Open http://localhost:3000/workflows/tasks
2. Click Approvals tab
3. Click any approval card
4. Draw signature in canvas
5. Click Approve
6. See success notification

## Core Demo (15 minutes)

1. Navigate to /workflows/tasks
2. Show Tasks tab
3. Click Approvals tab (show 3 cards)
4. Click PO-2024-001 to view details
5. Explain vendor info and cost breakdown
6. Go to approval page
7. Draw signature and approve
8. Show status updated
9. Go to Admin Reports
10. Click Analytics tab
11. Show metrics and trends

## Complete Demo (60-90 minutes)

### Session 1: Route Consolidation (5 min)
- Show Tasks tab (task management)
- Show Approvals tab (approval cards)
- Demonstrate deep linking (?tab=approvals)

### Session 2: Mock Database (10 min)
- Show 3 pre-loaded tasks
- Refresh page - data persists
- Open DevTools → Application → Local Storage
- Show `approval_tasks_v1` key with JSON data

### Session 3: Purchase Order Workflow (12 min)
- Navigate to PO detail page
- Show vendor information (company, contact, email, phone)
- Show cost breakdown (items, subtotal, tax, total)
- Show stage progress indicator
- Navigate to approval page
- Draw signature
- Add remarks
- Click Approve

### Session 4: Payment Voucher (10 min)
- Navigate to PV page
- Show invoice info
- Show payment method
- Show GL codes
- Complete approval

### Session 5: GRN Workflow (12 min)
- Navigate to GRN page
- Explain 2-stage workflow (different from 3-stage)
- Show item matching (PO vs Received)
- Show variance calculations
- Show damage tracking
- Complete GRN confirmation

### Session 6: Bulk Operations (15 min)
- Go back to Approvals tab
- Select multiple items with checkboxes
- Show selection counter
- Click Approve All
- Show dialog with count
- Enter remarks
- Complete approval
- Show success notification
- Demonstrate Reject All with required reason
- Demonstrate Reassign All with approver dropdown

### Session 7: Analytics Dashboard (12 min)
- Go to Admin Reports
- Click Analytics tab
- Show 5 metric cards (pending, approved, rejected, avg time, SLA)
- Scroll to Approval Trends (7-day history)
- Show Document Distribution
- Show Stage Performance Metrics
- Show Bottleneck Analysis with recommendations
- Show Performance Summary

### Session 8: End-to-End Journey (8 min)
- Complete one full approval workflow
- Watch analytics update in real-time
- Verify data persistence with page refresh

## Talking Points

**About the System**
- Automates approval workflows
- Supports 5 different document types
- Flexible (2-3 stages per type)
- Real-time visibility

**About the Demo**
- All workflows fully functional
- Data persists (survives refresh)
- Built with modern tech (Next.js, React, TypeScript)
- Production-ready code quality

**About Phase 12**
- Will add real database (PostgreSQL)
- Will add real authentication (OAuth 2.0)
- Will add email notifications
- Will add audit logging
- Complete roadmap documented

## Demo Ready Checklist

Before demo:
- [ ] npm run dev is running
- [ ] Open http://localhost:3000/workflows/tasks
- [ ] See Approvals tab
- [ ] See 3 approval cards
- [ ] No console errors (F12)
- [ ] DevTools show localStorage data
- [ ] Admin Reports page loads
- [ ] Analytics tab works

## Troubleshooting

**No data showing?**
- Hard refresh: Ctrl+Shift+R
- Check localStorage in DevTools
- Restart dev server

**Signature not drawing?**
- Use mouse on desktop
- Use touch on mobile
- Check browser console for canvas errors

**Analytics showing old data?**
- Click Refresh button
- Metrics auto-update after approvals

**Questions about Phase 12?**
- Reference PHASE-12-PLAN.md
- Explain: Database, Auth, Notifications, Audit, RBAC
- Show: 20-30 hour implementation plan
