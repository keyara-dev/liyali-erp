# Liyali Gateway - API Documentation

## Server Actions Overview

Server Actions are secure server-side functions called directly from client components. All requests are authenticated and validated on the server.

---

## Budget API

### Location
`src/app/_actions/budgets.ts`

### Functions

#### 1. `createBudget(data: CreateBudgetRequest): Promise<APIResponse<Budget>>`

Creates a new budget in DRAFT status.

**Parameters:**
```typescript
interface CreateBudgetRequest {
  name: string
  description?: string
  department: string
  departmentId: string
  fiscalYear: string
  totalAmount: number
  currency: string // "ZMW" or "USD"
  items: Omit<BudgetItem, "id" | "createdAt" | "updatedAt">[]
  createdBy: string // User ID
}
```

**Response:**
```typescript
{
  success: boolean
  message: string
  data?: Budget
  status: number
  statusText: string
}
```

**Example:**
```typescript
const result = await createBudget({
  name: "IT Department Annual Budget",
  description: "2024 allocation",
  department: "Information Technology",
  departmentId: "dept-it",
  fiscalYear: "2024",
  totalAmount: 100000,
  currency: "ZMW",
  items: [],
  createdBy: "user-123"
})
```

#### 2. `getBudgets(): Promise<APIResponse<Budget[]>>`

Retrieves all budgets (cached with React cache()).

**Response:** Array of Budget objects

#### 3. `getBudgetById(budgetId: string): Promise<APIResponse<Budget>>`

Retrieves a specific budget by ID.

**Parameters:**
- `budgetId: string` - Budget ID

**Response:** Single Budget object

#### 4. `submitBudgetForApproval(data: SubmitBudgetRequest): Promise<APIResponse<Budget>>`

Submits a budget for approval workflow.

**Parameters:**
```typescript
interface SubmitBudgetRequest {
  budgetId: string
  submittedBy: string // User ID
  comments?: string
}
```

**Response:** Updated Budget with IN_APPROVAL status

#### 5. `approveBudget(data: ApproveBudgetRequest): Promise<APIResponse<Budget>>`

Approves a budget at current stage.

**Parameters:**
```typescript
interface ApproveBudgetRequest {
  budgetId: string
  approvingUserId: string
  approvingUserRole: string
  comments?: string
  stageNumber?: number
}
```

**Response:** Updated Budget with approval record added

#### 6. `rejectBudget(data: RejectBudgetRequest): Promise<APIResponse<Budget>>`

Rejects a budget and returns to DRAFT.

**Parameters:**
```typescript
interface RejectBudgetRequest {
  budgetId: string
  rejectingUserId: string
  rejectionReason: string
  comments?: string
}
```

**Response:** Updated Budget with rejection record

---

## Tasks API

### Location
`src/app/_actions/tasks.ts`

### Functions

#### 1. `getTasksForUser(userId: string, status?: TaskStatus): Promise<APIResponse<Task[]>>`

Retrieves tasks assigned to a user with optional filtering.

**Parameters:**
- `userId: string` - User ID
- `status?: TaskStatus` - Filter by status (PENDING, IN_PROGRESS, COMPLETED)

**Response:** Array of Task objects

**Example:**
```typescript
// Get all pending tasks
const result = await getTasksForUser("user-123", "PENDING")
```

#### 2. `getTaskStats(userId: string): Promise<APIResponse<TaskStats>>`

Calculates task statistics for a user.

**Response:**
```typescript
interface TaskStats {
  totalTasks: number
  pendingTasks: number
  inProgressTasks: number
  completedTasks: number
  overdueTasks: number
  urgentTasks: number
  tasksByType: {
    BUDGET_APPROVAL: number
    REQUISITION_APPROVAL: number
    PURCHASE_ORDER_APPROVAL: number
    PAYMENT_VOUCHER_APPROVAL: number
    GOODS_RECEIVED_NOTE_CONFIRMATION: number
  }
}
```

#### 3. `getTaskById(taskId: string): Promise<APIResponse<Task>>`

Retrieves a specific task by ID.

**Parameters:**
- `taskId: string` - Task ID

**Response:** Single Task object

#### 4. `startTask(taskId: string, userId: string): Promise<APIResponse<Task>>`

Marks a task as IN_PROGRESS.

**Parameters:**
- `taskId: string` - Task ID
- `userId: string` - Current user ID

**Response:** Updated Task with IN_PROGRESS status

#### 5. `completeTask(taskId: string, userId: string): Promise<APIResponse<Task>>`

Marks a task as COMPLETED.

**Parameters:**
- `taskId: string` - Task ID
- `userId: string` - Current user ID

**Response:** Updated Task with COMPLETED status

---

## Workflow (Approval) API

### Location
`src/app/_actions/workflow.ts`

### Functions

#### 1. `approveDocument(documentId: string, comments?: string, signature?: string): Promise<APIResponse>`

Approves a workflow document.

**Parameters:**
- `documentId: string` - Document ID
- `comments?: string` - Optional approval comments
- `signature?: string` - Digital signature (base64 PNG)

**Response:** Success status and message

#### 2. `rejectDocument(documentId: string, remarks: string): Promise<APIResponse>`

Rejects a workflow document.

**Parameters:**
- `documentId: string` - Document ID
- `remarks: string` - Required rejection remarks

**Response:** Success status and message

---

## Settings API

### Location
`src/app/_actions/settings.ts`

### Functions

#### 1. `getUserProfile(): Promise<APIResponse<User>>`

Retrieves current user profile.

**Response:**
```typescript
interface User {
  id: string
  name: string
  email: string
  role: UserType
  department?: string
  avatar?: string
}
```

#### 2. `updateUserProfile(data: ProfileUpdateData): Promise<APIResponse<User>>`

Updates user profile information.

**Parameters:**
```typescript
interface ProfileUpdateData {
  name?: string
  email?: string
  department?: string
  avatar?: string
}
```

**Response:** Updated User object

#### 3. `changePassword(current: string, newPassword: string, confirm: string): Promise<APIResponse>`

Changes user password with validation.

**Parameters:**
- `current: string` - Current password
- `newPassword: string` - New password (min 8 chars)
- `confirm: string` - Password confirmation

**Validations:**
- Passwords must match
- Min 8 characters
- Cannot be same as current password

**Response:** Success status with change timestamp

#### 4. `updateGeneralSettings(settings: GeneralSettings): Promise<APIResponse>`

Updates user preferences.

**Parameters:**
```typescript
interface GeneralSettings {
  language?: string // "en", "es", "fr", "pt"
  theme?: "light" | "dark" | "system"
  timezone?: string
  emailNotifications?: boolean
  pushNotifications?: boolean
  activityNotifications?: boolean
}
```

**Response:** Updated settings object

#### 5. `getUserSessions(): Promise<APIResponse<Session[]>>`

Retrieves all active login sessions.

**Response:**
```typescript
interface Session {
  id: string
  device: string
  location: string
  ipAddress: string
  lastActive: string
  createdAt: string
  isCurrent: boolean
}
```

#### 6. `revokeSession(sessionId: string): Promise<APIResponse>`

Revokes a specific session (logout from device).

**Parameters:**
- `sessionId: string` - Session ID to revoke

**Response:** Success status

---

## Authentication API

### Location
`src/app/_actions/auth.ts`

### Functions

#### 1. `getCurrentUser(): Promise<APIResponse<User>>`

Retrieves current authenticated user.

**Response:** Current User object or null if not authenticated

**Example:**
```typescript
const result = await getCurrentUser()
if (result.success && result.data) {
  console.log(`Welcome ${result.data.name}`)
}
```

#### 2. `signOutAction(): Promise<APIResponse>`

Signs out current user and clears session.

**Response:** Success status

#### 3. `verifyAdminRole(): Promise<APIResponse>`

Verifies if current user has admin role.

**Response:** Admin status and user info

---

## Error Handling

All API responses follow a standard format:

```typescript
interface APIResponse<T = any> {
  success: boolean
  message: string
  data?: T
  status: number
  statusText: string
}
```

### Common Status Codes

| Code | Status | Meaning |
|------|--------|---------|
| 200 | OK | Request successful |
| 400 | BAD_REQUEST | Invalid input or validation error |
| 401 | UNAUTHORIZED | User not authenticated |
| 403 | FORBIDDEN | User lacks permission |
| 404 | NOT_FOUND | Resource not found |
| 500 | ERROR | Server error |

### Error Example

```typescript
const result = await createBudget(invalidData)
if (!result.success) {
  console.error(result.message)
  // Handle error
}
```

---

## Usage Patterns

### Pattern 1: Simple Operation

```typescript
'use client'
import { getBudgetById } from '@/app/_actions/budgets'

export function BudgetDetail({ budgetId }) {
  const [budget, setBudget] = useState(null)

  useEffect(() => {
    async function load() {
      const result = await getBudgetById(budgetId)
      if (result.success) {
        setBudget(result.data)
      }
    }
    load()
  }, [budgetId])
}
```

### Pattern 2: Form Submission

```typescript
const handleSubmit = async (data) => {
  try {
    const result = await createBudget(data)
    if (result.success) {
      toast.success('Budget created')
      router.push(`/budgets/${result.data.id}`)
    } else {
      toast.error(result.message)
    }
  } catch (error) {
    toast.error('Failed to create budget')
  }
}
```

### Pattern 3: Data Loading with Error Handling

```typescript
const { data, loading, error } = useAsync(
  async () => {
    const result = await getTasksForUser(userId)
    if (!result.success) throw new Error(result.message)
    return result.data
  },
  [userId]
)
```

---

## Mock Data

All APIs currently use mock data stored in server actions. For production:

1. Replace mock data with database queries
2. Add authentication middleware
3. Implement data validation
4. Add logging and monitoring
5. Set up error tracking (Sentry, etc.)

---

**Last Updated**: 2025-11-30
**Version**: 1.0.0
