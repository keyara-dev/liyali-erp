# Quick Start: Testing Approval System (2-Minute Setup)

## What You Need to Do

### Nothing! 🎉

Just open your browser and navigate to:

```
http://localhost:3000/workflows/approvals
```

All the mock data, server actions, and database simulation are already built in.

---

## 30-Second Demo Flow

### Step 1: See Dashboard (5 seconds)

You'll see 3 pending approval tasks:

- **REQ-2024-001** - HIGH priority (25K requisition)
- **BUD-2024-Q1-001** - MEDIUM priority (500K budget)
- **REQ-2024-002** - LOW priority (5K software)

### Step 2: Review a Task (10 seconds)

Click "Review" on REQ-2024-001

- See 10 items with prices
- See 3-stage workflow on the right
- See current stage: "Manager Approval"

### Step 3: Approve It (15 seconds)

1. Click green "Approve" button
2. Draw on signature canvas (just scribble)
3. Add remarks: "Looks good to me"
4. Click "Approve"

### Result: Task progresses to next stage ✅

---

## Full 5-Minute Test

### Test Approval

```
1. Dashboard → Click Review on REQ-2024-001
2. See requisition details (10 items)
3. Click Approve → Draw signature → Click Approve
4. Task moves to "Director Approval" stage
5. Refresh page → Changes persist
```

### Test Rejection

```
1. Dashboard → Click Review on BUD-2024-Q1-001
2. See budget allocations (K500K total)
3. Click Reject → Enter reason: "Needs justification"
4. Draw signature → Click Reject
5. Task shows as "Rejected"
6. Refresh → Rejection persists
```

### Test Reassignment

```
1. Dashboard → Click Review on REQ-2024-002
2. Click Reassign → Select "Jane Smith"
3. Enter reason: "Manager unavailable"
4. Click Reassign
5. Task now assigned to Jane Smith
6. Refresh → Reassignment persists
```

**Total time: ~5 minutes for all 3 workflows**

---

## How It Works (Behind the Scenes)

### No Backend Needed

- All data stored in **browser localStorage**
- Server actions run locally with mock database
- Survives page refreshes and browser restarts

### Data Structure

```
When you approve → approvalStore updates → localStorage saved → dashboard refreshes
```

### What Gets Saved

```javascript
localStorage["approval_tasks_v1"];
// Contains all 3 tasks with their approval history

localStorage["approval_history_v1"];
// Contains timeline of all approval actions
```

---

## Verify It's Working

### Check Browser Storage

1. Press `F12` to open DevTools
2. Go to **Application** tab
3. Click **LocalStorage** in sidebar
4. Click your domain (localhost:3000)
5. Look for `approval_tasks_v1` key
6. Click it to see all stored tasks

### View Approval History

1. After approving a task
2. Open DevTools Console
3. Type:
   ```javascript
   JSON.parse(localStorage.getItem("approval_tasks_v1"))["task-req-001"]
     .approvalHistory;
   ```
4. You'll see your approval with timestamp and remarks

---

## What You Can't Do (Yet)

❌ These are commented as TODOs for production:

- Real database (we use mock store instead)
- Send email notifications (we have template)
- Permission checks (we allow all)
- Audit logging (we simulate this)

All the commented code shows exactly where to add these features.

---

## Troubleshooting

### "I don't see any tasks"

→ Check if localStorage got cleared. Go to DevTools and refresh the page.

### "Tasks disappeared after refresh"

→ You might have disabled localStorage. Check DevTools under Storage.

### "Can't draw signature"

→ Try a different browser or check console for errors. Works in Chrome, Firefox, Safari, Edge.

### "Want to reset everything"

→ Open DevTools Console and type:

```javascript
localStorage.clear();
```

Then refresh the page.

---

## Sample Test Scenarios

### Scenario 1: Simple Approval (1 minute)

```
REQ-2024-001 → Approve → Next Stage → Done
```

### Scenario 2: Rejection with Reason (2 minutes)

```
BUD-2024-Q1-001 → Reject with reason → Sign → Task rejected
```

### Scenario 3: Reassign (2 minutes)

```
REQ-2024-002 → Reassign to Jane Smith → Reason: on leave → Done
```

### Scenario 4: Full Chain (5 minutes)

```
REQ-2024-001 → Approve (Stage 1) → Approve (Stage 2) → Approve (Stage 3) → Complete
```

---

## Key Features You'll Test

✅ **Approval with Signature** - Digital signature capture on canvas
✅ **Rejection with Reason** - Reason required before rejection
✅ **Reassignment** - Assign to different user with reason
✅ **Multi-Stage Workflow** - Tasks progress through stages
✅ **Status Tracking** - See current stage and progress
✅ **Data Persistence** - Changes survive page refresh
✅ **Approval History** - View all past actions
✅ **Filtering** - Filter by status, priority, search by ID
✅ **Statistics** - Dashboard shows pending, high priority, overdue counts

---

## Files You're Interacting With

### Backend (Mock)

- `src/lib/approval-store.ts` - Database simulation
- `src/app/_actions/approval-actions.ts` - Server actions

### Frontend

- `src/app/(private)/(main)/approvals/page.tsx` - Dashboard
- `src/app/(private)/(main)/requisitions/[id]/approval/page.tsx` - Approval page
- `src/app/(private)/(main)/budgets/[id]/approval/page.tsx` - Approval page
- `src/components/workflows/approval-action-panel.tsx` - Action buttons

### Hooks

- `src/hooks/use-approval-mutations.ts` - Update actions
- `src/hooks/use-approval-task-queries.ts` - Data fetching

---

## Next Steps

### After Testing:

1. Provide feedback on UI/UX
2. Identify bugs or improvements
3. When ready, replace mock with real database (all TODOs marked)
4. Add notifications and audit logging

### Real Integration:

All commented lines show exactly where to:

```typescript
// TODO: Replace approvalStore with real database
const task = approvalStore.getTaskDetail(taskId);

// TODO: Becomes:
const task = await db.approvalTasks.findUnique({ ... });
```

---

## Browser Console Commands

### View All Tasks

```javascript
JSON.parse(localStorage.getItem("approval_tasks_v1"));
```

### View Task Details

```javascript
JSON.parse(localStorage.getItem("approval_tasks_v1"))["task-req-001"];
```

### Count Approvals on a Task

```javascript
const tasks = JSON.parse(localStorage.getItem("approval_tasks_v1"));
tasks["task-req-001"].approvalHistory.length;
```

### Clear All Data

```javascript
localStorage.clear();
// Then refresh: Ctrl+R
```

---

## That's It!

You now have a fully functional approval system with:

- ✅ Complete workflows
- ✅ Data persistence
- ✅ Mock backend
- ✅ Real approval operations
- ✅ History tracking

**Start here**: http://localhost:3000/workflows/approvals

**Questions?** Check:

- APPROVAL_TESTING_GUIDE.md - Detailed testing scenarios
- PHASE_10_COMPLETION.md - Technical documentation
- PHASE_10_SUMMARY.md - Feature overview

Enjoy testing! 🚀
