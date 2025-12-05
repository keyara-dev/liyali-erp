# Backend Implementation Guide - Node.js with Prisma ORM

**Status**: Ready for Phase 13 Implementation
**Database**: PostgreSQL 12+
**Framework**: Express.js or Fastify
**ORM**: Prisma 5+
**Frontend Status**: ✅ Production Ready (Phases 1-12+ Complete)

## Current Frontend State (As of Dec 5, 2025)

The Liyali Gateway frontend is production-ready with the following capabilities:
- ✅ 5 Workflow Types (Requisition, Budget, PO, Payment Voucher, GRN)
- ✅ Multi-stage Approvals (2-3 stages with digital signatures)
- ✅ PDF Exports (Government-compliant with QR codes, preview, batch export, watermarks)
- ✅ Real-time Analytics Dashboard
- ✅ Bulk Operations (Approve/Reject/Reassign multiple items)
- ✅ 18+ Server Actions (Ready for migration to Prisma)
- ✅ React Query Integration (Cache management ready)
- ✅ Full TypeScript Type Safety

## What This Guide Covers

This guide provides a complete backend implementation in Node.js/Express that:
1. **Replaces localStorage** with PostgreSQL persistence
2. **Implements OAuth 2.0** authentication (via NextAuth.js on frontend)
3. **Provides REST/GraphQL APIs** for all workflow operations
4. **Handles Email Notifications** via SendGrid
5. **Implements Audit Logging** for compliance
6. **Enforces RBAC** with 7 user roles

## Integration Points

Your backend will serve these frontend endpoints:
- **GET /api/requisitions** - Fetch requisitions
- **POST /api/requisitions** - Create requisition
- **PUT /api/requisitions/:id** - Update requisition
- **POST /api/requisitions/:id/approve** - Approve workflow
- **POST /api/requisitions/:id/reject** - Reject workflow
- Similar endpoints for: budgets, purchase-orders, payment-vouchers, grn

See the frontend `src/app/_actions/` folder for all server action definitions.

## Table of Contents

1. [Project Setup](#project-setup)
2. [Prisma Schema](#prisma-schema)
3. [Database Setup](#database-setup)
4. [Express.js Setup](#expressjs-setup)
5. [API Routes & Controllers](#api-routes--controllers)
6. [Service Layer](#service-layer)
7. [Middleware](#middleware)
8. [Error Handling](#error-handling)
9. [Database Optimization](#database-optimization)
10. [NoSQL Considerations](#nosql-considerations)

---

## Project Setup

### Initialize Node.js Project

```bash
mkdir liyali-api
cd liyali-api
npm init -y

# Install dependencies
npm install express cors dotenv prisma @prisma/client
npm install -D typescript ts-node @types/node @types/express nodemon

# Initialize TypeScript
npx tsc --init

# Initialize Prisma
npx prisma init
```

### Package.json Scripts

```json
{
  "name": "liyali-api",
  "version": "1.0.0",
  "scripts": {
    "dev": "nodemon --exec ts-node src/index.ts",
    "build": "tsc",
    "start": "node dist/index.js",
    "prisma:migrate": "prisma migrate dev",
    "prisma:generate": "prisma generate",
    "prisma:seed": "ts-node prisma/seed.ts"
  },
  "dependencies": {
    "express": "^4.18.2",
    "cors": "^2.8.5",
    "dotenv": "^16.3.1",
    "@prisma/client": "^5.7.1",
    "jsonwebtoken": "^9.1.2",
    "bcrypt": "^5.1.1"
  },
  "devDependencies": {
    "typescript": "^5.3.3",
    "ts-node": "^10.9.2",
    "@types/node": "^20.10.6",
    "@types/express": "^4.17.21",
    "nodemon": "^3.0.2"
  }
}
```

### Environment Variables (.env)

```
DATABASE_URL="postgresql://user:password@localhost:5432/liyali_db"
JWT_SECRET="your-secret-key-change-in-production"
PORT=3001
NODE_ENV=development
SENDGRID_API_KEY="your-sendgrid-api-key"
ALLOWED_ORIGINS="http://localhost:3000,http://localhost:3001"
```

---

## Prisma Schema

### Complete Schema File (prisma/schema.prisma)

```prisma
// This is your Prisma schema file,
// learn more about it in the docs: https://pris.ly/d/prisma-schema

generator client {
  provider = "prisma-client-js"
}

datasource db {
  provider = "postgresql"
  url      = env("DATABASE_URL")
}

// User model with roles
enum UserRole {
  DEPARTMENT_MANAGER
  FINANCE_OFFICER
  DIRECTOR
  CFO
  COMPLIANCE_OFFICER
  ADMIN
  USER
}

model User {
  id           String    @id @default(uuid())
  email        String    @unique
  name         String
  role         UserRole
  department   String?
  isActive     Boolean   @default(true)
  lastLogin    DateTime?
  createdAt    DateTime  @default(now())
  updatedAt    DateTime  @updatedAt

  // Relations
  approvalTasks    ApprovalTask[]
  approvalHistory  ApprovalHistory[]
  auditLogs        AuditLog[]
  sessions         Session[]

  @@index([email])
  @@index([role])
  @@map("users")
}

// Session model for auth tokens
model Session {
  id        String   @id @default(uuid())
  userId    String
  token     String   @unique
  expiresAt DateTime
  createdAt DateTime @default(now())

  // Relations
  user      User     @relation(fields: [userId], references: [id], onDelete: Cascade)

  @@index([userId])
  @@index([expiresAt])
  @@map("sessions")
}

// Approval task - the core of the workflow
enum EntityType {
  REQUISITION
  BUDGET
  PO
  PV
  GRN
}

enum TaskStatus {
  pending
  approved
  rejected
}

model ApprovalTask {
  id              String      @id @default(uuid())
  entityId        String
  entityType      EntityType
  entityNumber    String
  status          TaskStatus  @default(pending)
  stageName       String
  stageIndex      Int
  importance      String      // LOW, MEDIUM, HIGH
  approverUserId  String
  createdAt       DateTime    @default(now())
  dueDate         DateTime
  workflowId      String
  workflowName    String
  document        Json?       // Flexible document storage

  // Relations
  approverUser    User                @relation(fields: [approverUserId], references: [id])
  history         ApprovalHistory[]

  @@index([status])
  @@index([entityId])
  @@index([approverUserId])
  @@index([entityType])
  @@index([createdAt])
  @@unique([entityId, stageIndex]) // Composite unique constraint
  @@map("approval_tasks")
}

// History of approvals/rejections
enum ActionType {
  approved
  rejected
  reassigned
  submitted
}

model ApprovalHistory {
  id              String      @id @default(uuid())
  taskId          String
  action          ActionType
  approverUserId  String
  timestamp       DateTime    @default(now())
  signature       String?     // Base64 encoded image
  remarks         String?     @db.Text
  previousApprover String?    // For reassignments

  // Relations
  task            ApprovalTask @relation(fields: [taskId], references: [id], onDelete: Cascade)
  approverUser    User         @relation(fields: [approverUserId], references: [id])

  @@index([taskId])
  @@index([action])
  @@index([timestamp])
  @@map("approval_history")
}

// Base document model
model Document {
  id          String      @id @default(uuid())
  type        EntityType
  number      String      @unique
  status      String
  creatorId   String
  data        Json
  createdAt   DateTime    @default(now())
  updatedAt   DateTime    @updatedAt

  @@index([type])
  @@index([number])
  @@index([createdAt])
  @@map("documents")
}

// Workflow definition
model Workflow {
  id          String      @id @default(uuid())
  name        String
  description String?     @db.Text
  entityType  EntityType
  status      String      // published, draft
  stages      Json        // Array of workflow stages
  createdAt   DateTime    @default(now())
  updatedAt   DateTime    @updatedAt
  createdBy   String

  @@index([entityType])
  @@index([status])
  @@map("workflows")
}

// Audit logging for compliance
model AuditLog {
  id          String      @id @default(uuid())
  userId      String
  action      String
  entityId    String?
  entityType  String?
  oldValue    Json?
  newValue    Json?
  timestamp   DateTime    @default(now())
  ipAddress   String?
  userAgent   String?     @db.Text

  // Relations
  user        User        @relation(fields: [userId], references: [id])

  @@index([userId])
  @@index([action])
  @@index([timestamp])
  @@map("audit_logs")
}

// Notifications
enum NotificationType {
  task_assigned
  approved
  rejected
}

model Notification {
  id        String             @id @default(uuid())
  userId    String
  type      NotificationType
  title     String
  message   String             @db.Text
  taskId    String?
  isRead    Boolean            @default(false)
  readAt    DateTime?
  createdAt DateTime           @default(now())

  // Relations
  user      User               @relation(fields: [userId], references: [id], onDelete: Cascade)

  @@index([userId])
  @@index([isRead])
  @@index([createdAt])
  @@map("notifications")
}
```

### Prisma Migrations

```bash
# Create migration after schema update
npx prisma migrate dev --name init

# Apply migrations in production
npx prisma migrate deploy

# Check migration status
npx prisma migrate status

# Create migration without applying
npx prisma migrate diff --from-empty --to-schema-file ./prisma/schema.prisma
```

---

## Database Setup

### PostgreSQL Connection Configuration

```typescript
// src/lib/prisma.ts

import { PrismaClient } from '@prisma/client'

const globalForPrisma = global as unknown as { prisma: PrismaClient }

export const prisma =
  globalForPrisma.prisma ||
  new PrismaClient({
    log: process.env.NODE_ENV === 'development'
      ? ['query', 'error', 'warn']
      : ['error'],
  })

if (process.env.NODE_ENV !== 'production') globalForPrisma.prisma = prisma
```

### Connection Pool Configuration

```typescript
// .env additions for connection pooling
DATABASE_URL="postgresql://user:password@localhost:5432/liyali_db?schema=public&sslmode=prefer&statement_cache_size=0&max_pool_size=20"
```

### Seed Data (prisma/seed.ts)

```typescript
import { PrismaClient, UserRole, EntityType, TaskStatus } from '@prisma/client'

const prisma = new PrismaClient()

async function main() {
  // Create test users
  const manager = await prisma.user.upsert({
    where: { email: 'manager@test.com' },
    update: {},
    create: {
      email: 'manager@test.com',
      name: 'John Manager',
      role: UserRole.DEPARTMENT_MANAGER,
      department: 'Sales',
      isActive: true,
    },
  })

  const finance = await prisma.user.upsert({
    where: { email: 'finance@test.com' },
    update: {},
    create: {
      email: 'finance@test.com',
      name: 'Jane Finance',
      role: UserRole.FINANCE_OFFICER,
      department: 'Finance',
      isActive: true,
    },
  })

  const cfo = await prisma.user.upsert({
    where: { email: 'cfo@test.com' },
    update: {},
    create: {
      email: 'cfo@test.com',
      name: 'Bob CFO',
      role: UserRole.CFO,
      department: 'Executive',
      isActive: true,
    },
  })

  // Create sample approval tasks
  await prisma.approvalTask.createMany({
    data: [
      {
        id: 'task-1',
        entityId: 'req-001',
        entityType: EntityType.REQUISITION,
        entityNumber: 'REQ-2024-001',
        status: TaskStatus.pending,
        stageName: 'Manager Review',
        stageIndex: 0,
        importance: 'HIGH',
        approverUserId: manager.id,
        dueDate: new Date(Date.now() + 5 * 24 * 60 * 60 * 1000),
        workflowId: 'wf-req-001',
        workflowName: '3-Stage Requisition',
        document: {
          description: 'Office supplies',
          amount: 2500,
          departmentId: 'dept-001',
        },
      },
      {
        id: 'task-2',
        entityId: 'po-001',
        entityType: EntityType.PO,
        entityNumber: 'PO-2024-001',
        status: TaskStatus.pending,
        stageName: 'Finance Review',
        stageIndex: 1,
        importance: 'MEDIUM',
        approverUserId: finance.id,
        dueDate: new Date(Date.now() + 3 * 24 * 60 * 60 * 1000),
        workflowId: 'wf-po-001',
        workflowName: '3-Stage Purchase Order',
        document: {
          vendorId: 'vendor-001',
          vendorName: 'Supplier Inc',
          amount: 5000,
        },
      },
    ],
  })

  // Create workflows
  await prisma.workflow.upsert({
    where: { id: 'wf-req-001' },
    update: {},
    create: {
      id: 'wf-req-001',
      name: '3-Stage Requisition',
      description: 'Approval workflow for purchase requisitions',
      entityType: EntityType.REQUISITION,
      status: 'published',
      stages: JSON.parse(JSON.stringify([
        {
          order: 0,
          name: 'Manager Review',
          approverRoles: ['DEPARTMENT_MANAGER'],
          allowReassign: true,
        },
        {
          order: 1,
          name: 'Finance Officer Review',
          approverRoles: ['FINANCE_OFFICER'],
          allowReassign: true,
        },
        {
          order: 2,
          name: 'CFO Approval',
          approverRoles: ['CFO'],
          allowReassign: false,
        },
      ])),
      createdBy: 'system',
    },
  })

  console.log('Seed data created successfully')
}

main()
  .catch((e) => {
    console.error(e)
    process.exit(1)
  })
  .finally(async () => {
    await prisma.$disconnect()
  })
```

---

## Express.js Setup

### Main Application File

```typescript
// src/index.ts

import express, { Express } from 'express'
import cors from 'cors'
import dotenv from 'dotenv'

import { prisma } from './lib/prisma'
import { errorHandler } from './middleware/errorHandler'
import { requestLogger } from './middleware/requestLogger'
import authRoutes from './routes/auth'
import approvalRoutes from './routes/approvals'
import workflowRoutes from './routes/workflows'
import analyticsRoutes from './routes/analytics'

dotenv.config()

const app: Express = express()
const port = process.env.PORT || 3001

// Middleware
app.use(express.json())
app.use(express.urlencoded({ extended: true }))
app.use(cors({
  origin: process.env.ALLOWED_ORIGINS?.split(','),
  credentials: true,
}))
app.use(requestLogger)

// Health check
app.get('/health', (req, res) => {
  res.json({ status: 'ok', timestamp: new Date() })
})

// Routes
app.use('/api/auth', authRoutes)
app.use('/api/approvals', approvalRoutes)
app.use('/api/workflows', workflowRoutes)
app.use('/api/analytics', analyticsRoutes)

// Error handling
app.use(errorHandler)

// Start server
app.listen(port, () => {
  console.log(`Server running on port ${port}`)
})

// Graceful shutdown
process.on('SIGINT', async () => {
  console.log('Shutting down gracefully...')
  await prisma.$disconnect()
  process.exit(0)
})
```

---

## API Routes & Controllers

### Approval Routes

```typescript
// src/routes/approvals.ts

import { Router } from 'express'
import { authMiddleware } from '../middleware/auth'
import * as approvalController from '../controllers/approvalController'

const router = Router()

// All routes require authentication
router.use(authMiddleware)

// Get all approval tasks
router.get('/tasks', approvalController.getApprovalTasks)

// Get single task detail
router.get('/tasks/:id', approvalController.getApprovalTaskDetail)

// Approve a task
router.post('/tasks/:id/approve', approvalController.approveTask)

// Reject a task
router.post('/tasks/:id/reject', approvalController.rejectTask)

// Reassign a task
router.post('/tasks/:id/reassign', approvalController.reassignTask)

// Bulk operations
router.post('/bulk/approve', approvalController.bulkApprove)
router.post('/bulk/reject', approvalController.bulkReject)
router.post('/bulk/reassign', approvalController.bulkReassign)

export default router
```

### Approval Controller

```typescript
// src/controllers/approvalController.ts

import { Request, Response, NextFunction } from 'express'
import { prisma } from '../lib/prisma'
import { TaskStatus, ActionType } from '@prisma/client'
import { v4 as uuid } from 'uuid'
import { logAudit } from '../services/auditService'
import { sendTaskNotification } from '../services/emailService'

// Types
interface ApproveTaskRequest {
  assignmentId: string
  stageNumber: number
  signature: string
  comments?: string
}

interface RejectTaskRequest {
  signature: string
  remarks: string
}

interface ReassignTaskRequest {
  newApproverId: string
  newApproverName: string
  reason?: string
}

// Get all approval tasks
export async function getApprovalTasks(
  req: Request,
  res: Response,
  next: NextFunction
) {
  try {
    const user = (req as any).user
    const { status, page = '1', limit = '20' } = req.query

    const pageNum = Math.max(1, parseInt(page as string))
    const limitNum = Math.min(100, Math.max(1, parseInt(limit as string)))
    const skip = (pageNum - 1) * limitNum

    // Build filter
    const where: any = {
      approverUserId: user.id,
    }

    if (status) {
      where.status = status as TaskStatus
    }

    // Get total count
    const total = await prisma.approvalTask.count({ where })

    // Get paginated tasks
    const tasks = await prisma.approvalTask.findMany({
      where,
      include: {
        approverUser: {
          select: {
            id: true,
            name: true,
            email: true,
            role: true,
          },
        },
      },
      orderBy: { createdAt: 'desc' },
      skip,
      take: limitNum,
    })

    res.json({
      success: true,
      data: {
        tasks,
        total,
        page: pageNum,
        limit: limitNum,
        pageCount: Math.ceil(total / limitNum),
      },
    })
  } catch (error) {
    next(error)
  }
}

// Get single task detail
export async function getApprovalTaskDetail(
  req: Request,
  res: Response,
  next: NextFunction
) {
  try {
    const { id } = req.params
    const user = (req as any).user

    const task = await prisma.approvalTask.findUnique({
      where: { id },
      include: {
        approverUser: {
          select: {
            id: true,
            name: true,
            email: true,
            role: true,
          },
        },
        history: {
          orderBy: { timestamp: 'desc' },
          include: {
            approverUser: {
              select: {
                id: true,
                name: true,
                email: true,
              },
            },
          },
        },
      },
    })

    if (!task) {
      return res.status(404).json({
        success: false,
        error: 'Task not found',
      })
    }

    // Verify user is the approver
    if (task.approverUserId !== user.id && user.role !== 'ADMIN') {
      return res.status(403).json({
        success: false,
        error: 'Insufficient permissions',
      })
    }

    res.json({
      success: true,
      data: {
        task,
        workflow: {
          id: task.workflowId,
          name: task.workflowName,
          // Fetch full workflow from database if needed
        },
      },
    })
  } catch (error) {
    next(error)
  }
}

// Approve a task
export async function approveTask(
  req: Request,
  res: Response,
  next: NextFunction
) {
  try {
    const { id } = req.params
    const user = (req as any).user
    const { assignmentId, stageNumber, signature, comments }: ApproveTaskRequest =
      req.body

    if (!assignmentId || !signature) {
      return res.status(400).json({
        success: false,
        error: 'Missing required fields',
      })
    }

    // Get current task
    const task = await prisma.approvalTask.findUnique({
      where: { id },
    })

    if (!task) {
      return res.status(404).json({
        success: false,
        error: 'Task not found',
      })
    }

    // Verify user is the approver
    if (task.approverUserId !== user.id) {
      return res.status(403).json({
        success: false,
        error: 'User is not the assigned approver',
      })
    }

    // Use transaction for consistency
    const result = await prisma.$transaction(async (tx) => {
      // Create approval history record
      await tx.approvalHistory.create({
        data: {
          id: uuid(),
          taskId: id,
          action: ActionType.approved,
          approverUserId: user.id,
          signature,
          remarks: comments,
          timestamp: new Date(),
        },
      })

      // Update task status
      const updatedTask = await tx.approvalTask.update({
        where: { id },
        data: { status: TaskStatus.approved },
      })

      // Log audit
      await logAudit(tx, {
        userId: user.id,
        action: 'approve_task',
        entityId: id,
        entityType: task.entityType.toString(),
        newValue: { status: TaskStatus.approved },
        ipAddress: req.ip,
        userAgent: req.get('user-agent'),
      })

      return updatedTask
    })

    // Send notification to next approver (if applicable)
    // await sendTaskNotification(...)

    res.json({
      success: true,
      data: {
        taskId: id,
        action: 'approved',
        newStatus: TaskStatus.approved,
        timestamp: new Date(),
      },
    })
  } catch (error) {
    next(error)
  }
}

// Reject a task
export async function rejectTask(
  req: Request,
  res: Response,
  next: NextFunction
) {
  try {
    const { id } = req.params
    const user = (req as any).user
    const { signature, remarks }: RejectTaskRequest = req.body

    if (!signature || !remarks) {
      return res.status(400).json({
        success: false,
        error: 'Signature and remarks are required',
      })
    }

    const task = await prisma.approvalTask.findUnique({
      where: { id },
    })

    if (!task) {
      return res.status(404).json({
        success: false,
        error: 'Task not found',
      })
    }

    if (task.approverUserId !== user.id) {
      return res.status(403).json({
        success: false,
        error: 'User is not the assigned approver',
      })
    }

    // Use transaction
    const result = await prisma.$transaction(async (tx) => {
      await tx.approvalHistory.create({
        data: {
          id: uuid(),
          taskId: id,
          action: ActionType.rejected,
          approverUserId: user.id,
          signature,
          remarks,
          timestamp: new Date(),
        },
      })

      const updatedTask = await tx.approvalTask.update({
        where: { id },
        data: {
          status: TaskStatus.rejected,
          stageIndex: 0, // Reset to first stage
        },
      })

      await logAudit(tx, {
        userId: user.id,
        action: 'reject_task',
        entityId: id,
        entityType: task.entityType.toString(),
        newValue: { status: TaskStatus.rejected, reason: remarks },
        ipAddress: req.ip,
        userAgent: req.get('user-agent'),
      })

      return updatedTask
    })

    res.json({
      success: true,
      data: {
        taskId: id,
        action: 'rejected',
        newStatus: TaskStatus.rejected,
        reason: remarks,
        timestamp: new Date(),
      },
    })
  } catch (error) {
    next(error)
  }
}

// Reassign a task
export async function reassignTask(
  req: Request,
  res: Response,
  next: NextFunction
) {
  try {
    const { id } = req.params
    const user = (req as any).user
    const {
      newApproverId,
      newApproverName,
      reason,
    }: ReassignTaskRequest = req.body

    if (!newApproverId) {
      return res.status(400).json({
        success: false,
        error: 'New approver ID is required',
      })
    }

    const task = await prisma.approvalTask.findUnique({
      where: { id },
    })

    if (!task) {
      return res.status(404).json({
        success: false,
        error: 'Task not found',
      })
    }

    if (task.approverUserId !== user.id) {
      return res.status(403).json({
        success: false,
        error: 'User is not the assigned approver',
      })
    }

    const previousApprover = task.approverUserId

    const result = await prisma.$transaction(async (tx) => {
      await tx.approvalHistory.create({
        data: {
          id: uuid(),
          taskId: id,
          action: ActionType.reassigned,
          approverUserId: newApproverId,
          remarks: reason,
          previousApprover,
          timestamp: new Date(),
        },
      })

      const updatedTask = await tx.approvalTask.update({
        where: { id },
        data: { approverUserId: newApproverId },
      })

      await logAudit(tx, {
        userId: user.id,
        action: 'reassign_task',
        entityId: id,
        entityType: task.entityType.toString(),
        newValue: { approverUserId: newApproverId, reason },
        ipAddress: req.ip,
        userAgent: req.get('user-agent'),
      })

      return updatedTask
    })

    res.json({
      success: true,
      data: {
        taskId: id,
        action: 'reassigned',
        newApprover: newApproverName,
        timestamp: new Date(),
      },
    })
  } catch (error) {
    next(error)
  }
}

// Bulk approve
export async function bulkApprove(
  req: Request,
  res: Response,
  next: NextFunction
) {
  try {
    const user = (req as any).user
    const { taskIds, remarks } = req.body

    if (!Array.isArray(taskIds) || taskIds.length === 0) {
      return res.status(400).json({
        success: false,
        error: 'taskIds array is required',
      })
    }

    const result = await prisma.$transaction(async (tx) => {
      let approved = 0

      for (const taskId of taskIds) {
        const task = await tx.approvalTask.findUnique({
          where: { id: taskId },
        })

        if (!task || task.approverUserId !== user.id) {
          continue
        }

        await tx.approvalHistory.create({
          data: {
            id: uuid(),
            taskId,
            action: ActionType.approved,
            approverUserId: user.id,
            remarks,
            timestamp: new Date(),
          },
        })

        await tx.approvalTask.update({
          where: { id: taskId },
          data: { status: TaskStatus.approved },
        })

        approved++
      }

      return approved
    })

    res.json({
      success: true,
      data: {
        approved: result,
        failed: taskIds.length - result,
        message: `Successfully approved ${result} tasks`,
        timestamp: new Date(),
      },
    })
  } catch (error) {
    next(error)
  }
}

// Bulk reject
export async function bulkReject(
  req: Request,
  res: Response,
  next: NextFunction
) {
  try {
    const user = (req as any).user
    const { taskIds, remarks } = req.body

    if (!remarks) {
      return res.status(400).json({
        success: false,
        error: 'Remarks are required for rejection',
      })
    }

    const result = await prisma.$transaction(async (tx) => {
      let rejected = 0

      for (const taskId of taskIds) {
        const task = await tx.approvalTask.findUnique({
          where: { id: taskId },
        })

        if (!task || task.approverUserId !== user.id) {
          continue
        }

        await tx.approvalHistory.create({
          data: {
            id: uuid(),
            taskId,
            action: ActionType.rejected,
            approverUserId: user.id,
            remarks,
            timestamp: new Date(),
          },
        })

        await tx.approvalTask.update({
          where: { id: taskId },
          data: {
            status: TaskStatus.rejected,
            stageIndex: 0,
          },
        })

        rejected++
      }

      return rejected
    })

    res.json({
      success: true,
      data: {
        rejected: result,
        failed: taskIds.length - result,
        message: `Successfully rejected ${result} tasks`,
        timestamp: new Date(),
      },
    })
  } catch (error) {
    next(error)
  }
}

// Bulk reassign
export async function bulkReassign(
  req: Request,
  res: Response,
  next: NextFunction
) {
  try {
    const user = (req as any).user
    const { taskIds, newApproverId, newApproverName, reason } = req.body

    if (!newApproverId) {
      return res.status(400).json({
        success: false,
        error: 'New approver ID is required',
      })
    }

    const result = await prisma.$transaction(async (tx) => {
      let reassigned = 0

      for (const taskId of taskIds) {
        const task = await tx.approvalTask.findUnique({
          where: { id: taskId },
        })

        if (!task || task.approverUserId !== user.id) {
          continue
        }

        await tx.approvalHistory.create({
          data: {
            id: uuid(),
            taskId,
            action: ActionType.reassigned,
            approverUserId: newApproverId,
            remarks: reason,
            previousApprover: user.id,
            timestamp: new Date(),
          },
        })

        await tx.approvalTask.update({
          where: { id: taskId },
          data: { approverUserId: newApproverId },
        })

        reassigned++
      }

      return reassigned
    })

    res.json({
      success: true,
      data: {
        reassigned: result,
        failed: taskIds.length - result,
        newApprover: newApproverName,
        timestamp: new Date(),
      },
    })
  } catch (error) {
    next(error)
  }
}
```

---

## Service Layer

### Audit Service

```typescript
// src/services/auditService.ts

import { PrismaClient, Prisma } from '@prisma/client'
import { v4 as uuid } from 'uuid'

interface AuditLogData {
  userId: string
  action: string
  entityId?: string
  entityType?: string
  oldValue?: any
  newValue?: any
  ipAddress?: string
  userAgent?: string
}

export async function logAudit(
  tx: Prisma.TransactionClient,
  data: AuditLogData
) {
  return tx.auditLog.create({
    data: {
      id: uuid(),
      userId: data.userId,
      action: data.action,
      entityId: data.entityId,
      entityType: data.entityType,
      oldValue: data.oldValue,
      newValue: data.newValue,
      ipAddress: data.ipAddress,
      userAgent: data.userAgent,
      timestamp: new Date(),
    },
  })
}

export async function getAuditLogs(
  userId?: string,
  action?: string,
  limit: number = 100
) {
  const prisma = new PrismaClient()

  const where: Prisma.AuditLogWhereInput = {}
  if (userId) where.userId = userId
  if (action) where.action = action

  return prisma.auditLog.findMany({
    where,
    orderBy: { timestamp: 'desc' },
    take: limit,
  })
}
```

### Email Service

```typescript
// src/services/emailService.ts

import sgMail from '@sendgrid/mail'
import { User } from '@prisma/client'

sgMail.setApiKey(process.env.SENDGRID_API_KEY || '')

interface EmailOptions {
  to: string
  subject: string
  html: string
}

export async function sendEmail(options: EmailOptions) {
  try {
    await sgMail.send({
      to: options.to,
      from: process.env.SENDGRID_FROM_EMAIL || 'noreply@liyali.com',
      subject: options.subject,
      html: options.html,
    })
  } catch (error) {
    console.error('Email send error:', error)
  }
}

export async function sendTaskNotification(
  user: User,
  taskNumber: string,
  action: 'assigned' | 'approved' | 'rejected'
) {
  const templates = {
    assigned: {
      subject: `New Task Assigned: ${taskNumber}`,
      html: `<p>A new approval task has been assigned to you: <strong>${taskNumber}</strong></p>
             <p>Please log in to review and take action.</p>`,
    },
    approved: {
      subject: `Task Approved: ${taskNumber}`,
      html: `<p>The task <strong>${taskNumber}</strong> has been approved and moved to the next stage.</p>`,
    },
    rejected: {
      subject: `Task Rejected: ${taskNumber}`,
      html: `<p>The task <strong>${taskNumber}</strong> has been rejected and returned to the requester.</p>`,
    },
  }

  await sendEmail({
    to: user.email,
    ...templates[action],
  })
}
```

---

## Middleware

### Authentication Middleware

```typescript
// src/middleware/auth.ts

import { Request, Response, NextFunction } from 'express'
import jwt from 'jsonwebtoken'
import { prisma } from '../lib/prisma'

declare global {
  namespace Express {
    interface Request {
      user?: any
    }
  }
}

export async function authMiddleware(
  req: Request,
  res: Response,
  next: NextFunction
) {
  try {
    const authHeader = req.get('Authorization')

    if (!authHeader) {
      return res.status(401).json({
        success: false,
        error: 'Missing authorization header',
      })
    }

    const parts = authHeader.split(' ')
    if (parts.length !== 2 || parts[0] !== 'Bearer') {
      return res.status(401).json({
        success: false,
        error: 'Invalid authorization format',
      })
    }

    const token = parts[1]

    try {
      const decoded = jwt.verify(token, process.env.JWT_SECRET || '') as {
        userId: string
      }

      const user = await prisma.user.findUnique({
        where: { id: decoded.userId },
      })

      if (!user || !user.isActive) {
        return res.status(401).json({
          success: false,
          error: 'User not found or inactive',
        })
      }

      req.user = user
      next()
    } catch (err) {
      return res.status(401).json({
        success: false,
        error: 'Invalid token',
      })
    }
  } catch (error) {
    next(error)
  }
}

// Role-based access control
export function roleMiddleware(...roles: string[]) {
  return (req: Request, res: Response, next: NextFunction) => {
    if (!req.user || !roles.includes(req.user.role)) {
      return res.status(403).json({
        success: false,
        error: 'Insufficient permissions',
      })
    }
    next()
  }
}
```

### Error Handler

```typescript
// src/middleware/errorHandler.ts

import { Request, Response, NextFunction } from 'express'

export function errorHandler(
  err: any,
  req: Request,
  res: Response,
  next: NextFunction
) {
  console.error('Error:', err)

  if (err.name === 'ValidationError') {
    return res.status(400).json({
      success: false,
      error: 'Validation failed',
      details: err.details,
    })
  }

  res.status(err.status || 500).json({
    success: false,
    error: err.message || 'Internal server error',
  })
}

export function requestLogger(
  req: Request,
  res: Response,
  next: NextFunction
) {
  const start = Date.now()
  res.on('finish', () => {
    const duration = Date.now() - start
    console.log(
      `${req.method} ${req.path} - ${res.statusCode} - ${duration}ms`
    )
  })
  next()
}
```

---

## Database Optimization

### Query Optimization Tips

```typescript
// 1. Use select to fetch only needed columns
const tasks = await prisma.approvalTask.findMany({
  select: {
    id: true,
    entityNumber: true,
    status: true,
    createdAt: true,
  },
})

// 2. Use include for relations
const tasks = await prisma.approvalTask.findMany({
  include: {
    approverUser: {
      select: {
        id: true,
        name: true,
        email: true,
      },
    },
    history: {
      take: 5, // Limit related records
    },
  },
})

// 3. Use batch operations
const updates = await prisma.approvalTask.updateMany({
  where: { status: 'pending' },
  data: { status: 'approved' },
})

// 4. Pagination always
const tasks = await prisma.approvalTask.findMany({
  skip: (page - 1) * limit,
  take: limit,
})
```

### Database Indexes

```sql
-- Indexes for common queries (auto-generated by Prisma)
-- For manual optimization:

CREATE INDEX idx_approval_tasks_approver_status
ON approval_tasks(approver_user_id, status);

CREATE INDEX idx_approval_history_task_action
ON approval_history(task_id, action);

CREATE INDEX idx_audit_logs_user_timestamp
ON audit_logs(user_id, "timestamp");

CREATE INDEX idx_notifications_user_read
ON notifications(user_id, is_read, created_at);
```

---

## NoSQL Considerations

While PostgreSQL is ideal for this workflow system, consider MongoDB for specific use cases:

```typescript
// Example: Using MongoDB for audit logs archival
import { MongoClient } from 'mongodb'

const mongoClient = new MongoClient(process.env.MONGODB_URI || '')

// Archive old audit logs to MongoDB for long-term retention
export async function archiveAuditLogs(daysOld: number = 90) {
  const cutoffDate = new Date(Date.now() - daysOld * 24 * 60 * 60 * 1000)

  // Get old logs from PostgreSQL
  const oldLogs = await prisma.auditLog.findMany({
    where: {
      timestamp: { lt: cutoffDate },
    },
  })

  // Insert to MongoDB
  const db = mongoClient.db('liyali_archive')
  const collection = db.collection('audit_logs')

  if (oldLogs.length > 0) {
    await collection.insertMany(oldLogs)

    // Delete from PostgreSQL
    await prisma.auditLog.deleteMany({
      where: {
        timestamp: { lt: cutoffDate },
      },
    })
  }
}
```

### MongoDB Connection Pooling

```typescript
// src/lib/mongodb.ts

import { MongoClient, ServerApiVersion } from 'mongodb'

const uri = process.env.MONGODB_URI || ''

const client = new MongoClient(uri, {
  serverApi: {
    version: ServerApiVersion.v1,
    strict: true,
    deprecationErrors: true,
  },
  maxPoolSize: 10,
  minPoolSize: 5,
})

export async function connectMongoDB() {
  try {
    await client.connect()
    console.log('MongoDB connected successfully')
  } catch (error) {
    console.error('MongoDB connection error:', error)
    throw error
  }
}

export default client
```

---

## Key Differences from Go Implementation

| Aspect | Go (Fiber) | Node.js (Express) |
|--------|-----------|------------------|
| Type Safety | Struct tags, compile-time | TypeScript, runtime |
| ORM | GORM | Prisma |
| Middleware | Fiber handlers | Express middleware |
| Transactions | Manual `tx` parameter | Prisma `$transaction` |
| Error Handling | Explicit error returns | Next(error) with handler |
| Async | Goroutines | Async/await |
| Deployment | Single binary | Node runtime |

---

## Performance Recommendations

1. **Connection Pooling**: Prisma handles automatically with configurable limits
2. **Query Caching**: Use Redis for frequently accessed analytics data
3. **Batch Processing**: Group multiple operations in transactions
4. **Pagination**: Always limit result sets
5. **Indexes**: Prisma creates these automatically based on schema
6. **Monitoring**: Use APM tools to track slow queries

---

**Status**: Ready for Phase 12 Implementation
**Next**: Commit consolidated documentation with both backend guides
