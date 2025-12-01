# System Architecture

## Overview

```
Next.js Application (Frontend)
    ├── React Components
    │   ├── Workflow UI components
    │   ├── Layout components
    │   └── Base UI components (shadcn/ui)
    │
    ├── Server Actions (Next.js)
    │   ├── Approval actions
    │   ├── Bulk operations
    │   └── Workflow queries
    │
    └── React Query Hooks
        ├── useGetApprovalTasks
        ├── useApproveTaskMutation
        └── ... 10+ hooks

Data Layer (Phase 11: localStorage | Phase 12: PostgreSQL)
    ├── Current: Browser Storage (localStorage)
    ├── Future: PostgreSQL Database
    └── ORM: Prisma (Phase 12)

Authentication (Phase 11: Mock | Phase 12: OAuth 2.0)
    ├── Current: Mock user context
    └── Future: NextAuth.js with OAuth
```

## Data Flow

### Approval Action Flow
```
User clicks "Approve"
    ↓
approval-action-panel.tsx calls useApproveTaskMutation
    ↓
Mutation calls approveTask (server action)
    ↓
Server action validates input
    ↓
Writes to approval-store.ts (localStorage)
    ↓
Server action returns success
    ↓
React Query invalidates caches
    ↓
UI refetches data
    ↓
Toast notification shows
    ↓
User sees updated status
```

## Technology Stack

| Layer | Technology | Purpose |
|-------|-----------|---------|
| **Frontend** | Next.js 13+ | Framework |
| **UI Library** | React | Component framework |
| **Styling** | Tailwind CSS | Utility CSS |
| **UI Components** | shadcn/ui | Pre-built components |
| **State** | React Query | Data fetching/caching |
| **Language** | TypeScript | Type safety |
| **Forms** | React Hook Form | Form management |
| **Storage** | localStorage (Ph11) | Current persistence |
| **Storage** | PostgreSQL (Ph12) | Future database |
| **ORM** | Prisma (Ph12) | Database queries |
| **Auth** | NextAuth.js (Ph12) | Authentication |

## Component Architecture

### Workflow Pages Structure
```
workflows/tasks/
├── page.tsx (Server Component)
│   └── TasksClient (Client Component)
│       ├── TasksTable
│       └── ApprovalsList

workflows/purchase-orders/[id]/
├── page.tsx (Server Component)
│   └── PODetailClient
│       ├── PODetailsSection
│       └── POItemsTable

workflows/purchase-orders/[id]/approval/
├── page.tsx (Server Component)
│   └── POApprovalClient
│       └── ApprovalActionPanel
```

## Data Persistence Layers

### Phase 11 (Current)
```
React Component
    ↓ calls mutation
Server Action
    ↓ reads/writes
approval-store.ts
    ↓ localStorage JSON
Browser LocalStorage
    ↓ survives refresh
```

### Phase 12 (Planned)
```
React Component
    ↓ calls mutation
Server Action
    ↓ Prisma queries
PostgreSQL Database
    ↓ real persistence
```

## Server Action Pattern

All server actions follow this pattern:

```typescript
export async function actionName(request: RequestType) {
  try {
    // Validate input
    if (!request.field) {
      return { success: false, error: 'message' }
    }

    // Process (Phase 11: use store, Phase 12: database)
    const result = store.operation(request)

    // Log
    console.log('[ACTION]', result)

    // Return result
    return { success: true, data: result }
  } catch (error) {
    console.error('[ERROR]', error)
    return { success: false, error: 'message' }
  }
}
```

## React Query Integration

### Hook Pattern
```typescript
export function useCustomQuery() {
  return useQuery({
    queryKey: ['namespace', 'key'],
    queryFn: async () => await serverAction(),
    refetchInterval: 30000, // Auto-refresh
  })
}
```

### Mutation Pattern
```typescript
export function useCustomMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: serverAction,
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ['namespace']
      })
      toast.success('Success message')
    },
  })
}
```

## File Organization

```
src/
├── app/
│   ├── (auth)/              # Auth pages
│   ├── (private)/           # Protected pages
│   │   ├── admin/           # Admin pages
│   │   └── workflows/       # Workflow pages
│   ├── _actions/            # Server actions
│   └── layout.tsx
│
├── components/
│   ├── workflows/           # Workflow components
│   ├── layout/              # Layout components
│   ├── notifications/       # Notification components
│   └── ui/                  # Base UI (shadcn)
│
├── hooks/
│   ├── use-approval-*       # Approval hooks
│   ├── use-notifications.ts
│   └── use-workflows.ts
│
├── lib/
│   ├── approval-store.ts    # Mock database
│   ├── constants.ts         # App constants
│   └── auth.ts              # Auth utilities
│
├── types/
│   ├── index.ts             # Main exports
│   ├── tasks.ts             # Approval types
│   └── notifications.ts
│
└── styles/
    └── globals.css          # Global styles
```

## 5 Workflow Types

Each workflow type follows the same pattern:

```
workflow-type/[id]/
├── page.tsx                 # Detail page
├── _components/
│   ├── *-detail-client.tsx  # Display details
│   └── *-items-table.tsx    # Line items (if applicable)
└── approval/
    ├── page.tsx             # Approval page
    └── _components/
        └── *-approval-client.tsx  # Approval form
```

## Authentication Approach

### Current (Phase 11)
- Mock user in context
- No real authentication
- All users can see all items

### Phase 12
```
OAuth 2.0 Providers
├── Entra ID (Microsoft 365)
├── Google
├── GitHub
└── SAML

Session Management
├── JWT tokens
├── 1-hour idle timeout
└── 8-hour absolute timeout

Role-Based Access (RBAC)
├── Department Manager
├── Finance Officer
├── Director/CFO
├── Compliance Officer
└── Admin
```

## Caching Strategy

### React Query Cache Keys
```
['approvals', 'tasks']          # Task list
['approvals', 'detail', id]     # Task detail
['approvals', 'stats']          # Statistics
['analytics', 'metrics']        # Dashboard metrics
['analytics', 'trends']         # Trends data
```

### Cache Invalidation
When mutations complete, invalidate these keys to refresh UI

## Error Handling

All operations include:
- Input validation
- Try/catch blocks
- User-friendly error messages
- Console logging for debugging
- Toast notifications for feedback

## Performance Considerations

- Pages load in <3 seconds
- React Query caches reduce API calls
- localStorage is instant
- PostgreSQL queries optimized (Phase 12)
- No N+1 queries

## Security (Phase 12)

- OAuth 2.0 for authentication
- Role-based access control
- API-level permission checks
- Audit logging for all actions
- Encrypted passwords
- Session management
