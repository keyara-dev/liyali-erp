# Phase 12: Database Integration & Real Authentication Implementation Plan

**Status**: PLANNING

**Estimated Duration**: 20-30 hours (4-6 weeks full-time)

**Target Database**: PostgreSQL

**Authentication Method**: OAuth 2.0 + Session Management

---

## Executive Summary

Phase 12 will transition the Liyali Gateway from a simulated, localStorage-based system to a production-ready application with real database persistence and enterprise authentication. This phase maintains 100% of the functionality delivered in Phases 1-11 while replacing mock data with real data sources.

### Phase 12 Objectives
1. ✅ Replace localStorage mock data with PostgreSQL database
2. ✅ Implement real OAuth 2.0 authentication
3. ✅ Add email notification system
4. ✅ Implement audit logging to database
5. ✅ Enforce real permission checks
6. ✅ Add payment processing integration hooks
7. ✅ Maintain 100% type safety
8. ✅ Zero breaking changes to UI/Components
9. ✅ Full backward compatibility with Phase 1-11

---

## Part A: Database Architecture

### A.1 Database Schema Overview

**Technology Stack**
- Database: PostgreSQL 14+
- ORM: Prisma
- Connection Pooling: pgBouncer or Prisma's built-in pooling
- Backups: AWS RDS automated backups

### A.2 Database Tables

#### Users Table
```sql
CREATE TABLE users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email VARCHAR(255) UNIQUE NOT NULL,
  name VARCHAR(255) NOT NULL,
  phone VARCHAR(20),
  role ENUM('REQUESTER', 'DEPT_MANAGER', 'FINANCE_OFFICER', 'DIRECTOR', 'CFO', 'COMPLIANCE_OFFICER', 'ADMIN') NOT NULL,
  department_id UUID,
  is_active BOOLEAN DEFAULT true,
  last_login_at TIMESTAMP,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (department_id) REFERENCES departments(id)
);
```

#### Sessions Table
```sql
CREATE TABLE sessions (
  id VARCHAR(255) PRIMARY KEY,
  user_id UUID NOT NULL,
  expires_at TIMESTAMP NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  ip_address VARCHAR(45),
  user_agent TEXT,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
```

#### Approval Tasks Table
```sql
CREATE TABLE approval_tasks (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  entity_id UUID NOT NULL,
  entity_type ENUM('REQUISITION', 'BUDGET', 'PURCHASE_ORDER', 'PAYMENT_VOUCHER', 'GRN') NOT NULL,
  entity_number VARCHAR(50) NOT NULL,
  status ENUM('pending', 'approved', 'rejected') DEFAULT 'pending',
  stage_name VARCHAR(255) NOT NULL,
  stage_index INTEGER NOT NULL,
  importance ENUM('LOW', 'MEDIUM', 'HIGH') DEFAULT 'MEDIUM',
  approver_id UUID NOT NULL,
  approver_name VARCHAR(255) NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  due_date TIMESTAMP,
  workflow_id UUID,
  workflow_name VARCHAR(255),
  FOREIGN KEY (approver_id) REFERENCES users(id),
  FOREIGN KEY (workflow_id) REFERENCES workflows(id)
);
```

#### Approval History Table
```sql
CREATE TABLE approval_history (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  task_id UUID NOT NULL,
  action ENUM('approved', 'rejected', 'reassigned') NOT NULL,
  approver_id UUID NOT NULL,
  remarks TEXT,
  signature_data TEXT, -- base64 encoded signature
  approved_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (task_id) REFERENCES approval_tasks(id) ON DELETE CASCADE,
  FOREIGN KEY (approver_id) REFERENCES users(id)
);
```

#### Documents Table (Requisitions, Budgets, POs, PVs)
```sql
CREATE TABLE documents (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  document_type ENUM('REQUISITION', 'BUDGET', 'PURCHASE_ORDER', 'PAYMENT_VOUCHER', 'GRN') NOT NULL,
  document_number VARCHAR(50) UNIQUE NOT NULL,
  status ENUM('DRAFT', 'SUBMITTED', 'IN_APPROVAL', 'APPROVED', 'REJECTED') DEFAULT 'DRAFT',
  created_by_id UUID NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  total_amount DECIMAL(15, 2),
  currency VARCHAR(3) DEFAULT 'ZMW',
  metadata JSONB, -- Flexible storage for document-specific fields
  FOREIGN KEY (created_by_id) REFERENCES users(id)
);
```

#### Audit Log Table
```sql
CREATE TABLE audit_logs (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL,
  action VARCHAR(255) NOT NULL,
  resource_type VARCHAR(50) NOT NULL,
  resource_id UUID NOT NULL,
  old_values JSONB,
  new_values JSONB,
  timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  ip_address VARCHAR(45),
  user_agent TEXT,
  FOREIGN KEY (user_id) REFERENCES users(id)
);
```

#### Notifications Table
```sql
CREATE TABLE notifications (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL,
  title VARCHAR(255) NOT NULL,
  message TEXT NOT NULL,
  type ENUM('INFO', 'WARNING', 'ERROR', 'SUCCESS') DEFAULT 'INFO',
  read_at TIMESTAMP,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
```

#### Workflows Table
```sql
CREATE TABLE workflows (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name VARCHAR(255) NOT NULL,
  entity_type VARCHAR(50) NOT NULL,
  stages JSONB NOT NULL, -- Array of stage configurations
  is_active BOOLEAN DEFAULT true,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### A.3 Prisma Schema

**File**: `prisma/schema.prisma`

```prisma
datasource db {
  provider = "postgresql"
  url      = env("DATABASE_URL")
}

generator client {
  provider = "prisma-client-js"
}

// Models matching the tables above
model User {
  id        String   @id @default(cuid())
  email     String   @unique
  name      String
  phone     String?
  role      UserRole
  department_id String?
  is_active Boolean  @default(true)
  lastLoginAt DateTime?
  createdAt DateTime @default(now())
  updatedAt DateTime @updatedAt

  sessions Session[]
  approvalTasks ApprovalTask[]
  approvalHistory ApprovalHistory[]
  auditLogs AuditLog[]
  notifications Notification[]
  documents Document[]

  @@index([role])
  @@index([email])
}

model Session {
  id String @id
  userId String
  expiresAt DateTime
  createdAt DateTime @default(now())
  ipAddress String?
  userAgent String?

  user User @relation(fields: [userId], references: [id], onDelete: Cascade)

  @@index([userId])
}

// ... Continue for all other models
```

### A.4 Database Initialization

**Step 1**: Create PostgreSQL database
```bash
createdb liyali_gateway
```

**Step 2**: Run migrations
```bash
npx prisma migrate dev --name init
```

**Step 3**: Seed initial data
```bash
npx prisma db seed
```

**Seed Script** (`prisma/seed.ts`):
- 11 pre-configured users (same as current mock data)
- 3 departments
- 5 sample workflows
- 20+ sample approval tasks
- Audit logs for initial setup

---

## Part B: Authentication Implementation

### B.1 OAuth 2.0 Setup

**Supported Providers** (Priority Order)
1. Entra ID (Azure AD) - Enterprise priority
2. Google Workspace - Fallback
3. GitHub - Development
4. Custom SAML 2.0 - Enterprise SSO

**Configuration Files**
- `.env.local` - Local secrets (never commit)
- `auth.config.ts` - Auth configuration
- `middleware.ts` - Route protection

### B.2 Session Management

**Session Architecture**
```
OAuth Provider
  ↓
NextAuth.js/Auth0 callback
  ↓
Create session in database
  ↓
Set secure HTTP-only cookie
  ↓
Validate on every request via middleware
```

**Session Properties**
```typescript
interface Session {
  user: {
    id: string
    email: string
    name: string
    role: UserRole
    department: string
    avatar?: string
  }
  expires: string
  sessionId: string
  createdAt: string
  ipAddress: string
}
```

**Idle Timeout**: 1 hour (configurable)
**Absolute Timeout**: 8 hours
**Session Refresh**: 30 minutes before expiry

### B.3 Implementation Steps

#### Step 1: Setup NextAuth.js (or Auth0)

**Install Dependencies**
```bash
npm install next-auth @next-auth/prisma-adapter
npm install @auth/prisma-adapter
```

**Create auth configuration** (`lib/auth.config.ts`)
```typescript
import { type NextAuthOptions } from "next-auth"
import AzureADProvider from "next-auth/providers/azure-ad"
import GoogleProvider from "next-auth/providers/google"
import { PrismaAdapter } from "@next-auth/prisma-adapter"
import { prisma } from "@/lib/prisma"

export const authOptions: NextAuthOptions = {
  adapter: PrismaAdapter(prisma),
  providers: [
    AzureADProvider({
      clientId: process.env.AZURE_AD_CLIENT_ID,
      clientSecret: process.env.AZURE_AD_CLIENT_SECRET,
      tenantId: process.env.AZURE_AD_TENANT_ID,
    }),
    GoogleProvider({
      clientId: process.env.GOOGLE_CLIENT_ID,
      clientSecret: process.env.GOOGLE_CLIENT_SECRET,
    }),
  ],
  callbacks: {
    async jwt({ token, user, account }) {
      if (user) {
        token.id = user.id
        token.role = user.role
      }
      return token
    },
    async session({ session, user }) {
      session.user.id = user.id
      session.user.role = user.role
      session.user.department = user.departmentId
      return session
    },
  },
  pages: {
    signIn: "/login",
    signOut: "/logout",
    error: "/auth/error",
  },
  session: {
    strategy: "jwt",
    maxAge: 8 * 60 * 60, // 8 hours
    updateAge: 30 * 60, // 30 minutes
  },
}
```

#### Step 2: Create Middleware for Route Protection

**File**: `middleware.ts`
```typescript
import { withAuth } from "next-auth/middleware"

export default withAuth(
  function middleware(req) {
    // Route protection logic
    if (req.nextUrl.pathname.startsWith("/admin")) {
      const isAdmin = req.nextauth?.token?.role === "ADMIN"
      if (!isAdmin) {
        return new Response("Unauthorized", { status: 403 })
      }
    }
  },
  {
    callbacks: {
      authorized: ({ token }) => !!token,
    },
  }
)

export const config = {
  matcher: [
    /*
     * Match all request paths except for the ones starting with:
     * - _next/static
     * - public files
     * - auth pages
     */
    "/((?!_next/static|public|auth|login|api/auth).*)",
  ],
}
```

#### Step 3: Replace Current Auth Implementation

**Remove**: `src/lib/auth.ts` (current simulated auth)

**Replace with**: NextAuth.js integration
- Update `src/app/(private)/layout.tsx` - Use SessionProvider
- Update `src/app/(private)/page.tsx` - Use useSession hook
- Update all `getCurrentUser()` calls - Use getServerSession()

### B.4 Session Validation Middleware

**Continuous Session Validation**
```typescript
// middleware.ts
async function validateSession(token: JWT) {
  const session = await prisma.session.findUnique({
    where: { id: token.sessionId },
  })

  if (!session || session.expiresAt < new Date()) {
    // Session expired or invalid
    return null
  }

  return session
}
```

**Idle Timeout Implementation**
```typescript
// hooks/useIdleTimeout.ts
export function useIdleTimeout() {
  useEffect(() => {
    let timeoutId: NodeJS.Timeout

    const resetTimer = () => {
      clearTimeout(timeoutId)
      timeoutId = setTimeout(() => {
        // Show warning at 55 minutes
        showIdleWarning()
      }, 55 * 60 * 1000)
    }

    window.addEventListener('mousemove', resetTimer)
    window.addEventListener('keypress', resetTimer)

    return () => {
      clearTimeout(timeoutId)
      window.removeEventListener('mousemove', resetTimer)
      window.removeEventListener('keypress', resetTimer)
    }
  }, [])
}
```

---

## Part C: Data Migration Strategy

### C.1 Migration Approach: Parallel Running

**Phase 1**: Parallel systems (2 weeks)
- Keep localStorage running
- Add database reads/writes
- Validate data consistency
- Zero downtime

**Phase 2**: Cutover (1 week)
- Switch to database-only
- Archive localStorage data
- Monitor for issues

### C.2 Mock Data to Database

**Current Mock Data Source**
```typescript
// Current: src/lib/approval-store.ts
export const approvalStore = {
  getTasks: () => [...],
  // ... mock data
}
```

**Migration Path**
1. Extract all mock data from store
2. Transform to database format
3. Create seed script
4. Run migration
5. Validate records match

**Seed Script** (`scripts/seed-database.ts`)
```typescript
async function seedDatabase() {
  // 1. Create users
  const users = await prisma.user.createMany({
    data: mockUsers.map(u => ({
      email: u.email,
      name: u.name,
      role: u.role,
      // ... other fields
    }))
  })

  // 2. Create approval tasks
  const tasks = await prisma.approvalTask.createMany({
    data: mockTasks.map(t => ({
      entityId: t.entityId,
      entityType: t.entityType,
      // ... other fields
    }))
  })

  // 3. Create approval history
  // ... continue for all entities

  console.log(`Seeded ${users.length} users and ${tasks.length} tasks`)
}
```

### C.3 Verification Checklist

- [ ] User count matches mock data
- [ ] Approval task count matches
- [ ] Workflow configurations match
- [ ] Approval history records match
- [ ] All foreign keys valid
- [ ] No data loss
- [ ] Performance acceptable

---

## Part D: Server Actions Migration

### D.1 Current vs. New Pattern

**Current Pattern** (Using Mock Store)
```typescript
// src/app/_actions/approval-actions.ts
export async function approveTask(request: ApproveTaskRequest) {
  const approvalStore = useApprovalStore() // Mock store
  approvalStore.approveTask(request.taskId, ...)
  return { success: true }
}
```

**New Pattern** (Using Database)
```typescript
// src/app/_actions/approval-actions.ts
export async function approveTask(request: ApproveTaskRequest) {
  const session = await getServerSession(authOptions)

  // Database query
  const task = await prisma.approvalTask.findUnique({
    where: { id: request.taskId },
  })

  // Verify permissions
  if (task.approverId !== session.user.id) {
    throw new Error('Unauthorized')
  }

  // Update in database
  await prisma.approvalTask.update({
    where: { id: request.taskId },
    data: {
      status: 'approved',
      approvalHistory: {
        create: {
          action: 'approved',
          approverId: session.user.id,
          remarks: request.remarks,
        }
      }
    }
  })

  // Invalidate cache
  revalidatePath('/workflows/tasks')

  return { success: true }
}
```

### D.2 Migration List

**Files to Update** (18 server action files)

1. `src/app/_actions/approval-actions.ts` (8 functions)
   - approveTask() → database query
   - rejectTask() → database query
   - reassignTask() → database query
   - getApprovalTasks() → database query
   - getApprovalTaskDetail() → database query
   - getApprovalStats() → database aggregation
   - getApprovalHistory() → database query
   - validateSignature() → database verify

2. `src/app/_actions/bulk-operations.ts` (6 functions)
   - bulkApproveTasks() → batch update
   - bulkRejectTasks() → batch update
   - bulkReassignTasks() → batch update
   - getAnalyticsMetrics() → database aggregation
   - getWorkflowTrends() → database time-series
   - getBottleneckAnalysis() → database analysis

3. `src/app/_actions/workflows.ts` (55+ functions)
   - Document CRUD operations
   - Workflow creation/update
   - Status transitions
   - Document retrieval
   - Search operations

4. `src/app/_actions/notifications.ts` (6+ functions)
   - Create notification → database
   - Mark as read → database
   - Get notifications → database query
   - Delete notification → database

5. Others:
   - `user-management.ts` (10+ functions)
   - `rbac.ts` (13+ functions)
   - `dashboard.ts` (5+ functions)

### D.3 Error Handling Pattern

**Add Standard Error Handling**
```typescript
export async function approveTask(request: ApproveTaskRequest) {
  try {
    const session = await getServerSession(authOptions)
    if (!session?.user) {
      return { success: false, error: 'Unauthorized' }
    }

    const task = await prisma.approvalTask.findUnique({
      where: { id: request.taskId },
    })

    if (!task) {
      return { success: false, error: 'Task not found' }
    }

    // Verify permissions
    if (task.approverId !== session.user.id) {
      // Log unauthorized attempt
      await logAudit({
        userId: session.user.id,
        action: 'UNAUTHORIZED_APPROVAL_ATTEMPT',
        resourceId: request.taskId,
      })
      return { success: false, error: 'Unauthorized' }
    }

    // Update task
    const updated = await prisma.approvalTask.update(...)

    return { success: true, data: updated }
  } catch (error) {
    console.error('[APPROVE_TASK_ERROR]', error)
    return { success: false, error: 'Internal server error' }
  }
}
```

---

## Part E: Email Notification System

### E.1 Email Service Setup

**Recommended Service**: SendGrid or AWS SES

**Configuration**
```env
# .env.local
SENDGRID_API_KEY=sg_...
SENDGRID_FROM_EMAIL=noreply@liyaligw.com
SENDGRID_FROM_NAME="Liyali Gateway"
```

**Prisma Email Log Table**
```prisma
model EmailLog {
  id        String   @id @default(cuid())
  to        String
  subject   String
  body      String
  status    String   @default("pending") // pending, sent, failed
  sentAt    DateTime?
  createdAt DateTime @default(now())
}
```

### E.2 Email Templates

**Template 1: Task Assigned**
```html
<h2>New Approval Task</h2>
<p>Hi {{ approverName }},</p>
<p>A new approval task has been assigned to you:</p>
<ul>
  <li>Document: {{ documentNumber }}</li>
  <li>Amount: {{ amount }}</li>
  <li>Stage: {{ stageName }}</li>
  <li>Due: {{ dueDate }}</li>
</ul>
<a href="{{ approvalLink }}">Review Task</a>
```

**Template 2: Approval Completed**
```html
<p>Hi {{ requesterName }},</p>
<p>Your {{ documentType }} request {{ documentNumber }} has been {{ action }}.</p>
<p>Status: {{ status }}</p>
<a href="{{ documentLink }}">View Document</a>
```

**Template 3: Rejection Notice**
```html
<p>Hi {{ requesterName }},</p>
<p>Your {{ documentType }} {{ documentNumber }} has been rejected.</p>
<p>Reason: {{ rejectionReason }}</p>
<a href="{{ documentLink }}">View Details</a>
```

### E.3 Email Service Implementation

**File**: `lib/email.ts`
```typescript
import sgMail from '@sendgrid/mail'

sgMail.setApiKey(process.env.SENDGRID_API_KEY!)

export async function sendApprovalEmail(
  to: string,
  type: 'assigned' | 'approved' | 'rejected',
  data: EmailData
) {
  const template = getTemplate(type, data)

  try {
    await sgMail.send({
      to,
      from: process.env.SENDGRID_FROM_EMAIL!,
      subject: template.subject,
      html: template.html,
    })

    // Log email
    await prisma.emailLog.create({
      data: {
        to,
        subject: template.subject,
        body: template.html,
        status: 'sent',
        sentAt: new Date(),
      },
    })

    return { success: true }
  } catch (error) {
    console.error('[EMAIL_ERROR]', error)

    // Log failed email
    await prisma.emailLog.create({
      data: {
        to,
        subject: template.subject,
        body: template.html,
        status: 'failed',
      },
    })

    return { success: false, error: String(error) }
  }
}
```

### E.4 Trigger Points for Emails

- [ ] Task assigned → Email approver
- [ ] Task approved → Email requester + next approver
- [ ] Task rejected → Email requester
- [ ] Task reassigned → Email new approver
- [ ] Bulk approve → Email all affected
- [ ] Daily digest → Email summary to admins
- [ ] SLA alert → Email when approaching due date

---

## Part F: Audit Logging Implementation

### F.1 Audit Log Schema

```typescript
interface AuditLog {
  id: string
  userId: string
  userName: string
  action: string // 'APPROVE', 'REJECT', 'CREATE', 'UPDATE', 'DELETE'
  resourceType: string // 'APPROVAL_TASK', 'DOCUMENT', 'USER'
  resourceId: string
  oldValues: Record<string, any>
  newValues: Record<string, any>
  ipAddress: string
  userAgent: string
  timestamp: Date
  status: 'SUCCESS' | 'FAILURE'
  errorMessage?: string
}
```

### F.2 Audit Logging Function

**File**: `lib/audit.ts`
```typescript
export async function logAudit(
  userId: string,
  action: string,
  resourceType: string,
  resourceId: string,
  changes?: {
    oldValues: Record<string, any>
    newValues: Record<string, any>
  },
  context?: {
    ipAddress: string
    userAgent: string
  }
) {
  try {
    await prisma.auditLog.create({
      data: {
        userId,
        action,
        resourceType,
        resourceId,
        oldValues: changes?.oldValues || {},
        newValues: changes?.newValues || {},
        ipAddress: context?.ipAddress || 'unknown',
        userAgent: context?.userAgent || 'unknown',
      },
    })
  } catch (error) {
    console.error('[AUDIT_LOG_ERROR]', error)
    // Don't throw - audit failure shouldn't block operations
  }
}
```

### F.3 Integration in Server Actions

**Example**: Approve Task with Audit
```typescript
export async function approveTask(request: ApproveTaskRequest) {
  const session = await getServerSession()

  const task = await prisma.approvalTask.findUnique({
    where: { id: request.taskId },
  })

  const oldValues = { status: task.status, stageName: task.stageName }

  const updated = await prisma.approvalTask.update({
    where: { id: request.taskId },
    data: { status: 'approved', stageIndex: task.stageIndex + 1 },
  })

  // Log audit
  await logAudit(
    session.user.id,
    'APPROVE_TASK',
    'APPROVAL_TASK',
    request.taskId,
    {
      oldValues,
      newValues: { status: updated.status, stageName: updated.stageName },
    },
    {
      ipAddress: request.headers.get('x-forwarded-for') || '',
      userAgent: request.headers.get('user-agent') || '',
    }
  )

  return { success: true }
}
```

---

## Part G: Permission Enforcement

### G.1 Permission Model

**Current**: Mock-based (all operations allowed)

**Target**: Role-based Access Control (RBAC)

**Permission Matrix**
```
ROLE                  | APPROVE | REJECT | REASSIGN | CREATE | VIEW | BULK_OPS
REQUESTER            | ✗       | ✗      | ✗        | ✓      | ✓    | ✗
DEPT_MANAGER         | ✓       | ✓      | ✓        | ✓      | ✓    | ✓
FINANCE_OFFICER      | ✓       | ✓      | ✓        | ✗      | ✓    | ✓
DIRECTOR             | ✓       | ✓      | ✓        | ✗      | ✓    | ✓
CFO                  | ✓       | ✓      | ✓        | ✗      | ✓    | ✓
COMPLIANCE_OFFICER   | ✗       | ✗      | ✗        | ✗      | ✓    | ✗
ADMIN                | ✓       | ✓      | ✓        | ✓      | ✓    | ✓
```

### G.2 Permission Check Function

**File**: `lib/permissions.ts`
```typescript
export async function checkPermission(
  userId: string,
  action: string,
  resourceType: string
): Promise<boolean> {
  const user = await prisma.user.findUnique({
    where: { id: userId },
  })

  if (!user) return false

  const permissions: Record<string, Set<string>> = {
    'APPROVE': new Set(['DEPT_MANAGER', 'FINANCE_OFFICER', 'DIRECTOR', 'CFO', 'ADMIN']),
    'REJECT': new Set(['DEPT_MANAGER', 'FINANCE_OFFICER', 'DIRECTOR', 'CFO', 'ADMIN']),
    'REASSIGN': new Set(['DEPT_MANAGER', 'FINANCE_OFFICER', 'DIRECTOR', 'CFO', 'ADMIN']),
    'CREATE_DOCUMENT': new Set(['REQUESTER', 'DEPT_MANAGER', 'ADMIN']),
    'VIEW_DOCUMENT': new Set(['REQUESTER', 'DEPT_MANAGER', 'FINANCE_OFFICER', 'DIRECTOR', 'CFO', 'COMPLIANCE_OFFICER', 'ADMIN']),
    'BULK_OPERATIONS': new Set(['DEPT_MANAGER', 'FINANCE_OFFICER', 'DIRECTOR', 'CFO', 'ADMIN']),
    'VIEW_ANALYTICS': new Set(['DEPT_MANAGER', 'DIRECTOR', 'CFO', 'COMPLIANCE_OFFICER', 'ADMIN']),
    'MANAGE_USERS': new Set(['ADMIN']),
  }

  return permissions[action]?.has(user.role) ?? false
}
```

### G.3 Authorization Middleware

**Apply in All Server Actions**
```typescript
export async function approveTask(request: ApproveTaskRequest) {
  const session = await getServerSession()

  // Check permission
  const canApprove = await checkPermission(
    session.user.id,
    'APPROVE',
    'APPROVAL_TASK'
  )

  if (!canApprove) {
    return { success: false, error: 'Permission denied' }
  }

  // ... rest of function
}
```

---

## Part H: React Query Cache Management

### H.1 Cache Key Refactoring

**Current**: Hardcoded strings

**Target**: Centralized query keys

**File**: `lib/query-keys.ts`
```typescript
export const queryKeys = {
  approval: {
    all: ['approval'] as const,
    tasks: () => [...queryKeys.approval.all, 'tasks'] as const,
    tasksByUser: (userId: string) => [...queryKeys.approval.tasks(), userId] as const,
    taskDetail: (taskId: string) => [...queryKeys.approval.all, 'detail', taskId] as const,
    stats: () => [...queryKeys.approval.all, 'stats'] as const,
    history: (taskId: string) => [...queryKeys.approval.all, 'history', taskId] as const,
  },
  document: {
    all: ['documents'] as const,
    list: (type: string) => [...queryKeys.document.all, 'list', type] as const,
    detail: (id: string) => [...queryKeys.document.all, 'detail', id] as const,
  },
  analytics: {
    all: ['analytics'] as const,
    metrics: () => [...queryKeys.analytics.all, 'metrics'] as const,
    trends: () => [...queryKeys.analytics.all, 'trends'] as const,
  },
} as const
```

### H.2 Cache Invalidation After Database Updates

**Pattern**:
```typescript
export async function approveTask(request: ApproveTaskRequest) {
  // ... database update ...

  // Invalidate related caches
  const queryClient = useQueryClient()
  await queryClient.invalidateQueries({
    queryKey: queryKeys.approval.tasks(),
  })
  await queryClient.invalidateQueries({
    queryKey: queryKeys.approval.stats(),
  })
  await queryClient.invalidateQueries({
    queryKey: queryKeys.analytics.metrics(),
  })

  return { success: true }
}
```

---

## Part I: Testing & Validation Strategy

### I.1 Unit Test Coverage

**Target**: 90%+ coverage for server actions

**Test Files to Create**:
- `__tests__/actions/approval.test.ts`
- `__tests__/actions/bulk-operations.test.ts`
- `__tests__/actions/documents.test.ts`
- `__tests__/lib/permissions.test.ts`
- `__tests__/lib/audit.test.ts`
- `__tests__/lib/email.test.ts`

### I.2 Integration Tests

**Test Scenarios**:
1. User login → Session created
2. Task assigned → Email sent + notification created
3. Approve task → History logged + next stage assigned
4. Bulk approve → All tasks updated + emails sent
5. Permission check → Unauthorized attempt blocked

### I.3 End-to-End Tests

**Using Playwright/Cypress**:
1. Login flow
2. Approve workflow (end-to-end)
3. Reject workflow
4. Bulk operations
5. Dashboard access

### I.4 Data Validation

**Pre-cutover Checklist**:
- [ ] All mock data migrated
- [ ] No data loss
- [ ] Performance acceptable
- [ ] Queries optimized
- [ ] Indexes present
- [ ] Backups working
- [ ] Restoration tested

---

## Part J: Rollout Strategy

### J.1 Deployment Phases

**Phase 1: Staging Environment** (Week 1)
- Deploy database schema
- Deploy code changes
- Run full test suite
- Performance validation
- Load testing

**Phase 2: Beta Users** (Week 2)
- Deploy to production with feature flag
- 10% of users → database backend
- Monitor for issues
- Collect feedback

**Phase 3: Full Rollout** (Week 3)
- Feature flag → 100%
- Parallel system running
- Monitor performance
- Archive localStorage

**Phase 4: Cleanup** (Week 4)
- Remove mock data code
- Remove localStorage logic
- Clean up feature flags
- Document final system

### J.2 Rollback Strategy

**If issues occur**:
```
Issue detected
  ↓
Feature flag → disable database
  ↓
Route traffic back to localStorage
  ↓
Investigate in staging
  ↓
Fix and redeploy
  ↓
Test thoroughly
  ↓
Re-enable with monitoring
```

### J.3 Monitoring & Alerts

**Metrics to Monitor**:
- [ ] Page load times
- [ ] API response times
- [ ] Database query times
- [ ] Error rates
- [ ] Email delivery rate
- [ ] Audit log creation success
- [ ] Session creation/validation

**Alert Thresholds**:
- Response time > 2 seconds
- Error rate > 1%
- Email failure rate > 5%
- Database connection pool > 80%

---

## Part K: File Changes Summary

### K.1 New Files to Create

```
lib/
  ├── prisma.ts (Prisma client singleton)
  ├── auth.config.ts (NextAuth configuration)
  ├── permissions.ts (RBAC functions)
  ├── audit.ts (Audit logging)
  ├── email.ts (Email service)
  ├── database/ (Database utilities)
  │   ├── migrations.ts
  │   ├── seeds.ts
  │   └── schema.ts

prisma/
  ├── schema.prisma (Database schema)
  ├── seed.ts (Seeding script)
  └── migrations/ (Migration files)

middleware.ts (Route protection)

scripts/
  ├── seed-database.ts
  ├── migrate-data.ts
  └── validate-migration.ts

__tests__/
  ├── actions/
  │   ├── approval.test.ts
  │   ├── bulk-operations.test.ts
  │   └── documents.test.ts
  └── lib/
      ├── permissions.test.ts
      ├── audit.test.ts
      └── email.test.ts
```

### K.2 Files to Modify

```
src/app/_actions/
  ├── approval-actions.ts (Replace mock with DB)
  ├── bulk-operations.ts (Replace mock with DB)
  ├── workflows.ts (Replace mock with DB)
  ├── notifications.ts (Replace mock with DB)
  ├── user-management.ts (Replace mock with DB)
  └── rbac.ts (Replace mock with DB)

src/lib/
  └── auth.ts (Remove simulated auth)

src/components/
  └── layout/
      └── session-provider.tsx (Add SessionProvider)

src/app/(private)/
  └── layout.tsx (Update with SessionProvider)

.env.example (Add new variables)
package.json (Add dependencies)
```

---

## Part L: Environment Variables

**New Variables to Add**:
```env
# Database
DATABASE_URL=postgresql://user:password@localhost:5432/liyali_gateway
DATABASE_POOL_SIZE=20

# Authentication
NEXTAUTH_URL=http://localhost:3000
NEXTAUTH_SECRET=your-secret-here

# OAuth Providers
AZURE_AD_CLIENT_ID=
AZURE_AD_CLIENT_SECRET=
AZURE_AD_TENANT_ID=

GOOGLE_CLIENT_ID=
GOOGLE_CLIENT_SECRET=

# Email Service
SENDGRID_API_KEY=
SENDGRID_FROM_EMAIL=noreply@liyaligw.com

# Session
SESSION_IDLE_TIMEOUT=3600 # 1 hour
SESSION_ABSOLUTE_TIMEOUT=28800 # 8 hours

# Feature Flags
ENABLE_DATABASE_BACKEND=false # Gradual rollout
USE_MOCK_DATA=false # Parallel running
```

---

## Part M: Implementation Checklist

### M.1 Pre-Implementation
- [ ] Review Phase 12 plan with team
- [ ] Set up PostgreSQL database
- [ ] Configure OAuth provider (Entra ID)
- [ ] Set up SendGrid account
- [ ] Create GitHub branch for Phase 12
- [ ] Set up staging environment

### M.2 Database Setup (Days 1-2)
- [ ] Create PostgreSQL database
- [ ] Write Prisma schema
- [ ] Run migrations
- [ ] Create seed script
- [ ] Seed initial data
- [ ] Validate data integrity

### M.3 Authentication (Days 3-4)
- [ ] Install NextAuth.js
- [ ] Configure OAuth provider
- [ ] Create auth config
- [ ] Implement middleware
- [ ] Test login flow
- [ ] Test session management

### M.4 Server Actions Migration (Days 5-7)
- [ ] Migrate approval-actions.ts
- [ ] Migrate bulk-operations.ts
- [ ] Migrate workflows.ts
- [ ] Migrate notifications.ts
- [ ] Migrate other action files
- [ ] Run unit tests

### M.5 Email System (Days 8-9)
- [ ] Set up SendGrid
- [ ] Create email templates
- [ ] Implement email service
- [ ] Add email triggers
- [ ] Test email delivery

### M.6 Audit Logging (Days 10-11)
- [ ] Create audit log schema
- [ ] Implement audit function
- [ ] Add logging to all actions
- [ ] Test audit trail

### M.7 Permissions (Days 12-13)
- [ ] Create permission matrix
- [ ] Implement permission check
- [ ] Add authorization to all endpoints
- [ ] Test access control

### M.8 Testing & Validation (Days 14-16)
- [ ] Unit tests
- [ ] Integration tests
- [ ] E2E tests
- [ ] Load tests
- [ ] Data migration validation

### M.9 Staging Deployment (Days 17-18)
- [ ] Deploy to staging
- [ ] Full test suite
- [ ] Performance validation
- [ ] Security audit

### M.10 Production Rollout (Days 19-21)
- [ ] Enable feature flag (10%)
- [ ] Monitor metrics
- [ ] Scale to 50%
- [ ] Scale to 100%
- [ ] Remove feature flag

### M.11 Cleanup & Monitoring (Days 22-25)
- [ ] Remove mock data code
- [ ] Archive localStorage
- [ ] Final documentation
- [ ] Team training
- [ ] Handover to ops team

---

## Part N: Known Challenges & Solutions

### N.1 Challenge: Data Consistency During Migration

**Problem**: Mock data might not match database expectations

**Solution**:
- Create validation script
- Compare mock vs. database
- Run parallel systems during transition
- Have rollback plan

### N.2 Challenge: Email Delivery Reliability

**Problem**: Emails might get marked as spam

**Solution**:
- Configure SPF/DKIM records
- Use SendGrid's authenticated sending
- Monitor delivery rates
- Implement retry logic

### N.3 Challenge: Session Management Edge Cases

**Problem**: Sessions might get invalidated unexpectedly

**Solution**:
- Test idle timeout thoroughly
- Implement grace period for refresh
- Show warning before expiry
- Log session events
- Monitor session metrics

### N.4 Challenge: Permission Enforcement Inconsistency

**Problem**: Permissions might be bypassed

**Solution**:
- Implement at every level (client + server)
- Add audit logging for all permission checks
- Implement in middleware
- Test with unauthorized users
- Regular security audit

---

## Appendix A: PostgreSQL Setup Commands

```bash
# Create database
createdb liyali_gateway

# Create user
createuser liyali_user --password

# Grant privileges
psql -d liyali_gateway -c "GRANT ALL PRIVILEGES ON DATABASE liyali_gateway TO liyali_user"

# Install extensions
psql -d liyali_gateway -c "CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\""
psql -d liyali_gateway -c "CREATE EXTENSION IF NOT EXISTS \"pgcrypto\""

# Set timezone
psql -d liyali_gateway -c "SET timezone = 'UTC'"

# Backup
pg_dump liyali_gateway > backup.sql

# Restore
psql liyali_gateway < backup.sql
```

---

## Appendix B: Prisma Commands

```bash
# Initialize Prisma
npx prisma init

# Create migration
npx prisma migrate dev --name init

# Reset database (development only)
npx prisma migrate reset

# Seed database
npx prisma db seed

# Generate Prisma client
npx prisma generate

# View database
npx prisma studio
```

---

## Appendix C: Next Steps After Phase 12

### Post-Phase 12 Enhancements
1. **Payment Processing** - Integrate payment gateway
2. **Advanced Analytics** - Machine learning for bottleneck prediction
3. **Mobile App** - React Native mobile application
4. **API Gateway** - Public REST/GraphQL API for integrations
5. **Compliance** - SOC 2, ISO 27001 certifications
6. **Internationalization** - Multi-language support
7. **Advanced Reporting** - Business intelligence dashboards

### Maintenance & Operations
- Database backups & recovery procedures
- Performance monitoring & optimization
- Security patching & updates
- User support & documentation
- Capacity planning

---

## Summary

Phase 12 represents the transition from a feature-complete simulation to a production-ready system. By following this plan:

✅ All Phase 1-11 functionality remains intact
✅ 100% type safety maintained
✅ Zero breaking changes to UI
✅ Enterprise authentication implemented
✅ Real data persistence
✅ Audit trail for compliance
✅ Email notifications
✅ Permission enforcement
✅ Ready for enterprise deployment

**Estimated Timeline**: 4-6 weeks (20-30 hours)
**Success Criteria**: All items in M.11 checklist complete
**Go-Live Readiness**: Enterprise-grade system with 99.9% uptime SLA capability

---

**Document Version**: 1.0
**Last Updated**: 2024-12-01
**Next Review**: After Phase 11 completion, before Phase 12 start
