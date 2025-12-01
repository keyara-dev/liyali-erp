# Notification System for Workflow Tasks - Design Document

## Overview

Users must receive **real-time notifications** when new tasks are assigned, with quick-action capabilities and a complete notification history.

---

## Part 1: Requirements

### Notification Types
```
1. TASK_ASSIGNED
   - New workflow task created for user
   - Show: "New approval task: Requisition #REQ-001 needs your approval"
   - Action: "Review" → Opens entity detail + approval modal

2. TASK_REASSIGNED
   - Task reassigned to this user
   - Show: "Task reassigned to you: Requisition #REQ-001 (reassigned by Admin)"
   - Action: "Review" → Opens entity detail

3. TASK_APPROVED
   - One of your created tasks was approved
   - Show: "Your Requisition #REQ-001 was approved by John Manager"
   - Action: "View" → Opens entity detail (read-only)

4. TASK_REJECTED
   - One of your created tasks was rejected
   - Show: "Your Requisition #REQ-001 was rejected by John Manager: Budget exceeded"
   - Action: "Revise & Resubmit" → Opens entity for editing

5. WORKFLOW_COMPLETE
   - Workflow you initiated completed
   - Show: "Your Requisition #REQ-001 was fully approved!"
   - Action: "View" → Opens completed entity
```

### Display Locations

**1. Top App Bar Notification Component**
   - Bell icon with unread count badge
   - Dropdown showing latest 5 notifications
   - Each with:
     - Icon (task type)
     - Title & message
     - Time ago
     - Quick action button
     - Mark as read on hover
   - "View All" link → Full page

**2. Quick Actions (In Notification Dropdown)**
   ```
   For TASK_ASSIGNED / TASK_REASSIGNED:
     Button: "Review Now"
     → Opens approval modal with signature/remarks

   For TASK_APPROVED / WORKFLOW_COMPLETE:
     Button: "View"
     → Opens entity detail read-only

   For TASK_REJECTED:
     Button: "Revise & Resubmit"
     → Opens entity for editing
   ```

**3. Confirmation Modal (When Clicking Quick Action)**
   - Show task details
   - Pre-fill action (approve/reject)
   - Optional remarks field
   - Signature capture
   - "Confirm" button submits approval

**4. Notifications Page**
   - `/workflows/notifications` (new page)
   - All notifications (read + unread)
   - Filters: Type, Date, Read Status
   - Pagination (20 per page)
   - Bulk actions: Mark all read, Delete old
   - Each notification shows:
     - Icon & type badge
     - Full message
     - Timestamp
     - Read/unread indicator
     - Action button
     - Delete button

---

## Part 2: Data Model

### Notification Type

```typescript
interface Notification {
  // Identity
  id: string                              // UUID
  userId: string                          // Who receives it

  // Content
  type: NotificationType
  title: string                           // "New approval task"
  message: string                         // Full message
  icon?: string                           // Icon name

  // Context
  entityId?: string                       // req-001, budget-002, etc.
  entityType?: WorkflowEntityType
  entityNumber?: string                   // REQ-2024-001
  relatedUserId?: string                  // Who caused the notification

  // State
  isRead: boolean                         // User viewed it
  readAt?: Date
  actionTaken?: boolean                   // Did user act on it?
  actionTakenAt?: Date

  // Quick Action
  quickAction: QuickAction
  quickActionData?: Record<string, any>   // Context for action

  // Metadata
  createdAt: Date
  expiresAt?: Date                        // Auto-delete old notifications
  importance: 'LOW' | 'MEDIUM' | 'HIGH'   // Affects display priority
}

type NotificationType =
  | 'TASK_ASSIGNED'
  | 'TASK_REASSIGNED'
  | 'TASK_APPROVED'
  | 'TASK_REJECTED'
  | 'WORKFLOW_COMPLETE'
  | 'APPROVAL_OVERDUE'                   // SLA breach
  | 'COMMENT_ADDED'                      // Someone commented

interface QuickAction {
  type: 'REVIEW_AND_APPROVE'              // Opens approval modal
       | 'VIEW_ONLY'                      // Opens entity read-only
       | 'REVISE_AND_RESUBMIT'            // Opens for editing
       | 'NONE'                           // No action button
  label: string                           // Button text: "Review Now", "Revise"
  route?: string                          // Navigation path
  params?: Record<string, string>         // URL params
}

interface NotificationPreferences {
  userId: string
  emailNotifications: boolean             // Send email?
  pushNotifications: boolean              // Browser push?
  inAppNotifications: boolean             // Show in app?
  notifyOn: {
    taskAssigned: boolean
    taskReassigned: boolean
    taskApproved: boolean
    taskRejected: boolean
    workflowComplete: boolean
    approvalOverdue: boolean
    commentsAdded: boolean
  }
}
```

---

## Part 3: System Architecture

```
┌─────────────────────────────────────────────────────┐
│  WORKFLOW EVENTS                                    │
│  (Approval, Rejection, Reassignment, etc.)          │
└────────────────────┬────────────────────────────────┘
                     │
                     ↓
        ┌────────────────────────┐
        │ Notification Generator │
        │ (Event handlers)        │
        └────────────┬───────────┘
                     │
         ┌───────────┼───────────┐
         │           │           │
         ↓           ↓           ↓
    Create        Send         Update
    Notification  Email        UI
                  (Future)
         │           │           │
         └───────────┼───────────┘
                     │
                     ↓
        ┌────────────────────────┐
        │  Notification Store    │
        │  (Persistence)         │
        └────────────┬───────────┘
                     │
         ┌───────────┴───────────┐
         │                       │
         ↓                       ↓
    App Bar Icon          Notifications Page
    (Real-time)           (Historical)
    with Badge            with Filters
```

---

## Part 4: Implementation Components

### New Types File: `src/types/notifications.ts`

```typescript
// All Notification types above
```

### New Persistence: `src/lib/notification-persistence.ts`

```typescript
// Save, query, mark read
export async function createNotification(notification: Notification)
export async function getNotification(id: string)
export async function getUserNotifications(userId: string, limit: number)
export async function getUserUnreadCount(userId: string)
export async function markAsRead(notificationId: string)
export async function markAllAsRead(userId: string)
export async function deleteNotification(id: string)
export async function deleteOldNotifications(olderThanDays: number)
```

### New Actions: `src/app/_actions/notifications.ts`

```typescript
// Server actions for notification operations
export async function getNotifications(userId: string, page: number)
export async function markNotificationRead(id: string)
export async function deleteNotification(id: string)
export async function getUserNotificationPreferences(userId: string)
export async function updateNotificationPreferences(userId: string, prefs: Partial<NotificationPreferences>)
```

### New Hooks: `src/hooks/use-notifications.ts`

```typescript
// React Query hooks for notifications
export function useUserNotifications(userId: string)
export function useUnreadNotificationCount(userId: string)
export function useMarkNotificationRead()
export function useNotificationPreferences(userId: string)
```

### New Components

**1. App Bar Notification Bell**
`src/components/notifications/notification-bell.tsx`
```typescript
- Show bell icon
- Unread count badge
- On click: Show dropdown with 5 latest notifications
- Each notification with:
  - Icon + type
  - Title & message
  - Time ago
  - Quick action button
  - Mark read on hover
- "View All" link at bottom
- Real-time updates via polling or WebSocket
```

**2. Notification Dropdown Component**
`src/components/notifications/notification-dropdown.tsx`
```typescript
- List of latest notifications
- Scrollable
- Mark as read indicator
- Quick action buttons
- Show loading state
- Show empty state
```

**3. Single Notification Item**
`src/components/notifications/notification-item.tsx`
```typescript
- Icon based on type
- Title & message
- Timestamp (relative: "2 mins ago")
- Unread badge
- Quick action button with onClick handler
- Hover effect to show delete/read buttons
```

**4. Notification Modal for Quick Actions**
`src/components/notifications/notification-action-modal.tsx`
```typescript
When user clicks "Review Now":
  - Show entity summary (Requisition #REQ-001)
  - Current stage info
  - Required signature/remarks
  - Modal with:
    - Entity preview
    - Signature canvas (required)
    - Remarks field (optional)
    - "Approve" / "Reject" buttons
    - "Cancel" button
  - On submit:
    - Call approval action
    - Close modal
    - Show success toast
    - Mark notification as actionTaken
    - Mark as read
    - Update UI
```

**5. Notifications Page**
`src/app/(private)/workflows/notifications/page.tsx`
```typescript
- Header: "Notifications" with total count
- Filters section:
  - Type dropdown (All, Task Assigned, Approved, etc.)
  - Date range picker
  - Read status toggle (All, Unread, Read)
  - Search box
- Notification list:
  - Paginated (20 per page)
  - Sort by date (newest first)
  - Each item shows:
    - Icon + type badge
    - Full message
    - Sender name (if applicable)
    - Timestamp
    - Read/unread indicator (visual)
    - Action button
    - Delete button (trash icon)
- Bulk actions:
  - "Mark all as read" button (top)
  - Checkbox to select multiple
  - "Delete selected" button
- Empty state: "No notifications"
```

**6. Notification Client Component (Real-time)**
`src/components/notifications/notification-provider.tsx`
```typescript
- React context provider
- Hooks into useUserNotifications
- Polls every 30 seconds (or WebSocket when available)
- Updates unread count
- Triggers toast on new notification
- Passes data to bell component
```

---

## Part 5: Event Triggers

### When Notification is Created

**Trigger 1: Task Assigned**
```typescript
When: progressToNextStage() moves to next stage
What: Create TASK_ASSIGNED notification
Who: assignee of next stage
Data: {
  type: 'TASK_ASSIGNED',
  title: 'New approval task',
  message: `Requisition #${req.requisitionNumber} needs your approval`,
  entityId: req.id,
  entityType: 'REQUISITION',
  quickAction: {
    type: 'REVIEW_AND_APPROVE',
    label: 'Review Now',
    params: { entityId: req.id }
  }
}
```

**Trigger 2: Task Reassigned**
```typescript
When: reassignStage() happens
What: Create TASK_REASSIGNED notification
Who: new assignee
Data: {
  type: 'TASK_REASSIGNED',
  title: 'Task reassigned to you',
  message: `You were assigned: Requisition #${req.requisitionNumber}`,
  relatedUserId: reassignedBy,
  quickAction: { type: 'REVIEW_AND_APPROVE', ... }
}
```

**Trigger 3: Task Approved**
```typescript
When: approveDocument() completes
What: Create TASK_APPROVED notification
Who: entity creator
Data: {
  type: 'TASK_APPROVED',
  title: 'Task approved',
  message: `Your Requisition #${req.requisitionNumber} was approved by ${approver.name}`,
  relatedUserId: approverId,
  quickAction: { type: 'VIEW_ONLY', ... }
}
```

**Trigger 4: Task Rejected**
```typescript
When: rejectDocument() completes
What: Create TASK_REJECTED notification
Who: entity creator
Data: {
  type: 'TASK_REJECTED',
  title: 'Task rejected',
  message: `Your Requisition was rejected: ${rejectionRemarks}`,
  relatedUserId: rejectorId,
  quickAction: { type: 'REVISE_AND_RESUBMIT', ... }
}
```

**Trigger 5: Workflow Complete**
```typescript
When: progressToNextStage() hits FINAL
What: Create WORKFLOW_COMPLETE notification
Who: entity creator
Data: {
  type: 'WORKFLOW_COMPLETE',
  title: 'Approval complete',
  message: `Your Requisition #${req.requisitionNumber} was fully approved!`,
  quickAction: { type: 'VIEW_ONLY', ... }
}
```

---

## Part 6: Integration Points

### In approval.ts (When Approving)
```typescript
async function approveDocument(...) {
  // Existing approval logic
  const result = await progressToNextStage(...)

  // NEW: Create notification for next approver
  if (result.nextApprover) {
    await createNotification({
      userId: result.nextApprover.userId,
      type: 'TASK_ASSIGNED',
      message: `New approval needed: ${entity.number}`,
      // ...
    })
  }

  // NEW: Notify entity creator if complete
  if (result.isComplete) {
    await createNotification({
      userId: entity.createdBy,
      type: 'WORKFLOW_COMPLETE',
      message: `Your ${entity.type} was approved!`,
      // ...
    })
  }
}
```

### In reassignment handler
```typescript
async function reassignStage(...) {
  const result = await reassignStage(...)

  // Create notification for new approver
  await createNotification({
    userId: request.newApproverId,
    type: 'TASK_REASSIGNED',
    message: `Task reassigned to you: ${entity.number}`,
    relatedUserId: request.reassignedBy,
    // ...
  })
}
```

### In layout/provider
```typescript
// Add notification provider to wrap app
// Starts polling for notifications
<NotificationProvider>
  {children}
</NotificationProvider>
```

---

## Part 7: App Bar Integration

**Location**: `src/components/layout/header/`

**Current**: User menu component
**Add**: Notification bell to the left of user menu

```
┌─────────────────────────────────────────┐
│  [Logo]  [Search]       [🔔 5] [👤]     │
│                         Notifications  User
│                         Bell icon       Menu
│                         with badge
│
│                    On Click:
│                    ┌──────────────┐
│                    │ Notification │
│                    │ Dropdown     │
│                    │              │
│                    │ • Task 1     │
│                    │ • Task 2     │
│                    │ • Task 3     │
│                    │              │
│                    │ [View All]   │
│                    └──────────────┘
└─────────────────────────────────────────┘
```

---

## Part 8: Implementation Phases

### Phase 5A: Notification Types & Persistence
- [ ] Create `src/types/notifications.ts`
- [ ] Create `src/lib/notification-persistence.ts`
- [ ] Add notification columns to data model
- [ ] Seed sample notifications

### Phase 5B: Server Actions
- [ ] Create `src/app/_actions/notifications.ts`
- [ ] Create notification creation trigger helpers
- [ ] Integrate into existing approval flows

### Phase 5C: React Query Hooks
- [ ] Create `src/hooks/use-notifications.ts`
- [ ] Real-time polling setup
- [ ] Unread count tracking

### Phase 5D: UI Components
- [ ] Notification bell component
- [ ] Dropdown component
- [ ] Notification item component
- [ ] Quick action modal
- [ ] Notifications page

### Phase 5E: Integration
- [ ] Add notification provider to layout
- [ ] Integrate with approval actions
- [ ] Integrate with reassignment actions
- [ ] Add bell to app bar

### Phase 5F: Features
- [ ] Real-time notification updates
- [ ] Mark as read functionality
- [ ] Notification preferences
- [ ] Email notifications (future)
- [ ] Browser push notifications (future)

---

## Part 9: Quick Action Flow

```
User sees notification in bell icon
  ↓
Clicks "Review Now" button
  ↓
Modal opens showing:
  - Entity details (Requisition #REQ-001)
  - Current stage & approver
  - Required fields
  ↓
User fills form:
  - Signature (required)
  - Remarks (optional)
  ↓
Clicks "Approve" button
  ↓
Form validates:
  - Signature present? ✓
  - Remarks if required? ✓
  ↓
Call approveDocument() server action
  ↓
System:
  - Records approval
  - Progresses workflow
  - Creates notifications for next stage
  - Updates entity status
  ↓
Modal closes with success toast
Notification marked as:
  - Read: true
  - actionTaken: true
  ↓
Bell icon count decreases
Entity detail view updates
  ↓
Done ✅
```

---

## Part 10: Summary

**What Will Be Built:**
- ✅ Notification data model
- ✅ Persistence layer
- ✅ Real-time notification generation
- ✅ Unread count tracking
- ✅ Top app bar notification bell
- ✅ Quick action modal with approval
- ✅ Full notifications history page
- ✅ Mark as read / delete
- ✅ Notification preferences

**User Experience:**
- Gets notification immediately when task assigned
- Can approve right from notification (2-click approval)
- Can see notification history anytime
- Knows what was approved/rejected via notifications

---

**Status**: Ready for Phase 5A implementation
**Priority**: HIGH - Critical for user workflow
**Estimated Effort**: 2 days (types, persistence, basic UI)
