# Code Structure

## Directory Organization

```
src/app/
├── (auth)/              # Auth pages
├── (private)/           # Protected pages
│   ├── admin/           # Admin reports
│   └── workflows/       # Workflow pages
├── _actions/            # Server actions
└── layout.tsx

src/components/
├── workflows/           # Workflow components
├── notifications/       # Notification UI
├── layout/              # Layout components
└── ui/                  # shadcn components

src/hooks/
├── use-approval-flow.ts
├── use-approval-mutations.ts
├── use-approval-task-queries.ts
└── use-workflows.ts

src/lib/
├── approval-store.ts    # Mock database
├── constants.ts
└── auth.ts

src/types/
├── index.ts
├── tasks.ts
└── notifications.ts
```

## Server Actions

### Approval Actions (src/app/_actions/approval-actions.ts)
- getApprovalTasks() - Fetch tasks
- getApprovalTaskDetail() - Single task
- approveTask() - Approve with signature
- rejectTask() - Reject with reason
- reassignTask() - Reassign to approver

### Bulk Operations (src/app/_actions/bulk-operations.ts)
- bulkApproveTasks() - Approve multiple
- bulkRejectTasks() - Reject multiple
- bulkReassignTasks() - Reassign multiple
- getAnalyticsMetrics() - Dashboard data

## React Hooks

### Query Hooks
```typescript
useGetApprovalTasks()       // Fetch tasks
useGetApprovalTaskDetail()  // Fetch single
useGetApprovalStats()       // Statistics
useWorkflows()              // List workflows
```

### Mutation Hooks
```typescript
useApproveTaskMutation()    // Approve action
useRejectTaskMutation()     // Reject action
useReassignTaskMutation()   // Reassign action
useCreateWorkflow()         // Create workflow
```

## Key Components

**approval-action-panel.tsx**
- Approval form with signature
- Approve/Reject/Reassign buttons
- Loading states

**bulk-operations-toolbar.tsx**
- Multi-select actions
- Approve/Reject/Reassign dialogs
- Validation

**analytics-dashboard.tsx**
- Metrics cards
- Trends and distribution
- Bottleneck analysis

## Mock Database (approval-store.ts)

```typescript
// Core operations
loadFromStorage()    // Load from localStorage
saveToStorage()      // Save to localStorage
getAllTasks()        // Get all tasks
approveTask()        // Record approval
rejectTask()         // Record rejection
```

## Type Definitions (src/types/)

```typescript
ApprovalTask          // Single task
ApprovalTaskDetail    // Task with full data
Workflow              // Workflow definition
WorkflowStage         // Approval stage
Notification          // Notification object
```

## File Naming

```
Components:    PascalCase        ApprovalActionPanel.tsx
Hooks:         camelCase use*    use-approval-flow.ts
Stores:        camelCase         approval-store.ts
Types:         PascalCase        ApprovalTask
Constants:     UPPER_SNAKE_CASE  MAX_ITEMS
```

## Code Imports

```typescript
// External
import { useState } from 'react'

// Next.js
import { useRouter } from 'next/navigation'

// Local
import { Card } from '@/components/ui/card'
import { useApproveTaskMutation } from '@/hooks'
import { ApprovalTask } from '@/types'
```

## Workflow Pattern

Each workflow follows:
```
workflow-type/[id]/
├── page.tsx                     # Detail page
├── _components/detail-client.tsx
└── approval/
    ├── page.tsx
    └── _components/approval-client.tsx
```

All workflow types use the same components and patterns.

## Key Technologies

- Next.js 13+ (App Router)
- React (UI)
- TypeScript (Types)
- React Query (Data)
- Tailwind CSS (Styling)
- shadcn/ui (Components)
