# Custom Workflow Management System - Design Plan

## Executive Summary

This document outlines a comprehensive plan for implementing a **user-defined workflow management system** that allows organizations to design and manage custom approval workflows without code changes. Users will be able to:

1. **Create custom workflows** - Define workflow stages, naming, and sequencing
2. **Assign approvers** - Specify which users/roles must approve at each stage
3. **Define state transitions** - Specify next states on approval vs. rejection
4. **Attach workflows to entities** - Apply workflows to requisitions, budgets, POs, etc.
5. **Entity lifecycle** - Create in DRAFT, submit, follow defined workflow, reach final approval

---

## Part 1: Architecture Analysis & Design Decisions

### 1.1 Current System Overview

**Existing Strengths:**

- ✅ Configuration-driven approval system (not hardcoded)
- ✅ Flexible approval stages with role-based assignment
- ✅ Complete audit trail and approval history
- ✅ Signature capture and validation
- ✅ Multiple entity types (Requisition, Budget, PO, GRN)
- ✅ Task-based assignment system
- ✅ RBAC with 7 predefined roles

**Current Limitations:**

- ❌ Fixed workflows per document type (hardcoded in approval-config.ts)
- ❌ No way for users to create custom workflows
- ❌ No workflow assignment to entities (implicit via document type)
- ❌ No workflow versioning or lifecycle management
- ❌ In-memory storage (not persisted)
- ❌ No workflow template system
- ❌ No workflow builder/designer UI
- ❌ Requisitions hardcoded with 3-stage default workflow
- ❌ No conditional approval routing
- ❌ No parallel/concurrent approvals support

### 1.2 Design Approach: Layered Architecture

We will implement **three new layers** above the existing system:

```
┌─────────────────────────────────────────────┐
│  CUSTOM WORKFLOWS LAYER (NEW)               │
│  - Workflow Designer                         │
│  - Workflow CRUD                             │
│  - Workflow Templates                        │
└──────────────┬──────────────────────────────┘
               │
┌──────────────▼──────────────────────────────┐
│  WORKFLOW ASSIGNMENT LAYER (NEW)            │
│  - Attach workflows to entities             │
│  - Workflow mapping configuration            │
│  - Dynamic workflow resolution              │
└──────────────┬──────────────────────────────┘
               │
┌──────────────▼──────────────────────────────┐
│  APPROVAL ORCHESTRATION LAYER (ENHANCED)    │
│  - Use workflow definition for routing      │
│  - Map custom stages to approval flow       │
│  - Fallback to defaults if no workflow      │
└──────────────┬──────────────────────────────┘
               │
┌──────────────▼──────────────────────────────┐
│  EXISTING APPROVAL ENGINE (UNCHANGED)       │
│  - Config-driven state machine              │
│  - Stage progression logic                  │
│  - Signature/audit trail                    │
└─────────────────────────────────────────────┘
```

### 1.3 Key Design Decisions

#### Decision 1: Workflow Definition Model

**Option A: Stage-Based (Chosen)**

- User defines named stages
- Each stage has: name, description, required role/user, next stage on approve/reject
- Simple, intuitive, matches current mental model
- Easier to visualize and design

**Option B: State-Machine Based**

- Define states and transitions as graph
- More complex but more flexible
- Harder to understand for non-technical users

**Decision: STAGE-BASED** - Simpler for UI/UX, sufficient for requirements

#### Decision 2: Workflow Scope & Binding

**Option A: Global Workflows (Many entities can use)**

- Create workflow once, reuse across multiple requisitions/budgets
- Reduces duplication
- Requires workflow versioning

**Option B: Entity-Specific Workflows**

- Each entity has its own workflow definition
- More flexibility, no versioning needed
- More storage, harder to manage

**Decision: GLOBAL WORKFLOWS** - More enterprise-like, enable template reuse

#### Decision 3: Custom Approver Assignment

**Option A: Role-Based Only**

- "DEPARTMENT_MANAGER approves this stage"
- Automatic assignment to any user with that role
- Current system already does this

**Option B: Role-Based + Specific User Assignment**

- "Stage 1: Any DEPARTMENT_MANAGER can approve OR specifically John Smith"
- More flexible
- More complex data model

**Decision: ROLE-BASED PRIMARY, OPTIONAL USER OVERRIDE** - Best balance

#### Decision 4: State Transition Definition

**Option A: Implicit (same for all stages)**

- Approval always → next stage
- Rejection always → REJECTED (back to DRAFT for resubmit)
- Simple, current behavior

**Option B: Explicit per Stage**

- Each stage defines its own next states
- "On approve → stage 2, On reject → revert to requester"
- More powerful

**Decision: EXPLICIT PER STAGE** - Match user requirements for "next state on approve/reject"

#### Decision 5: Workflow Attachment Mechanism

**Option A: Entity-Level**

- When creating requisition/budget, select which workflow
- Entity has workflowId field
- Clean, explicit

**Option B: Entity Type Level**

- Configure: "All Requisitions use Workflow X"
- Implicit, less flexible
- Current system does this

**Option C: Hybrid** (Chosen)

- Default workflow per entity type (configured globally)
- Override per entity (when creating/editing)
- Best flexibility

**Decision: HYBRID** - Flexible but with sensible defaults

#### Decision 6: Persistence Layer

**Current:** In-memory Maps (lose on restart)
**Option A:** Keep in-memory (for MVP)

- Fast iteration
- Demo-friendly
- Not production-ready

**Option B:** Add basic file-based persistence

- Use JSON file storage
- Persists across restarts
- Better than nothing

**Option C:** Full database migration

- Requires DB schema changes
- Beyond scope of this design

**Decision: FILE-BASED PERSISTENCE** - Bridge between current and production

---

## Part 2: Data Model Design

### 2.1 New Types to Add

#### CustomWorkflow Type

```typescript
interface CustomWorkflow {
  // Identity & Metadata
  id: string; // UUID
  name: string; // e.g., "2-Stage Approval"
  description: string;
  version: number; // 1, 2, 3... for versioning

  // Scope & Applicability
  applicableEntityTypes: WorkflowEntityType[]; // REQUISITION, BUDGET, PURCHASE_ORDER
  organizationId?: string; // Multi-tenancy support (future)
  isTemplate: boolean; // Can be used as template?
  isActive: boolean; // Can be used for new entities?

  // Workflow Definition
  stages: WorkflowStage[]; // Array of approval stages
  totalStages: number;
  allowConcurrentApprovals: boolean; // Future: parallel approvals

  // Metadata
  createdBy: string; // User ID
  createdAt: Date;
  updatedBy: string;
  updatedAt: Date;

  // Audit
  usageCount: number; // How many entities use this?
  lastUsedAt?: Date;
  deprecatedAt?: Date; // Mark old workflows
}

interface WorkflowStage {
  id: string; // UUID or stageNumber
  stageNumber: number; // 1, 2, 3...
  stageName: string; // e.g., "Department Manager Review"
  description: string;

  // Approval Assignment
  approverType: "ROLE" | "USER" | "ROLE_OR_USER";
  requiredRole?: UserRole; // If ROLE or ROLE_OR_USER
  optionalSpecificUser?: string; // UserId if role override

  // Requirements
  requiresSignature: boolean; // Default: true
  requiresComments: boolean; // Can be optional, required, or optional
  commentsType: "OPTIONAL" | "REQUIRED" | "DISABLED";

  // Validation
  requiredValidations?: string[]; // e.g., budgetAvailable, complianceCheck
  slaHours?: number; // SLA deadline
  escalationRole?: UserRole; // Who to escalate to if SLA breached

  // Actions on Approval
  onApprovalActions?: {
    sendNotification: boolean;
    generateQRCode?: boolean;
    generatePaymentReference?: boolean;
    createAuditLog: boolean;
  };

  // State Transitions
  onApprove: {
    nextStage: number | "FINAL"; // Stage number or FINAL for end
    setEntityStatus: DocumentStatus; // Optional: override entity status
    notifyUsers: boolean;
  };

  onReject: {
    nextStage: number | "REJECTED"; // Stage to revert to, or mark rejected
    setEntityStatus: DocumentStatus; // What status when rejected
    notifyRequester: boolean;
    requiresRejectionReason: boolean;
  };

  onReverse?: {
    previousStage: number;
    resetApprovals: boolean; // Reset subsequent stages?
  };

  // UI Configuration
  displayOrder: number;
  isOptional: boolean; // Can be skipped?
  canBeReassigned: boolean;
  canBeRejected: boolean;
  canBeReversed: boolean;
}

type WorkflowEntityType =
  | "REQUISITION"
  | "BUDGET"
  | "PURCHASE_ORDER"
  | "GOODS_RECEIVED_NOTE"
  | "PAYMENT_VOUCHER"
  | "CUSTOM";

enum DocumentStatus {
  DRAFT = "DRAFT",
  SUBMITTED = "SUBMITTED",
  IN_APPROVAL = "IN_APPROVAL",
  APPROVED = "APPROVED",
  REJECTED = "REJECTED",
  REVERSED = "REVERSED",
}
```

#### WorkflowAssignment Type

```typescript
interface WorkflowAssignment {
  // Identity
  id: string; // UUID

  // Entity Reference
  entityId: string; // requisitionId, budgetId, etc.
  entityType: WorkflowEntityType;

  // Workflow Reference
  workflowId: string; // CustomWorkflow ID
  workflowVersion: number; // Version of workflow used

  // Current State
  currentStageNumber: number; // Which stage we're on
  stageStartedAt: Date; // When did this stage start?

  // History
  stageHistory: StageExecution[]; // Execution record for each stage

  // Metadata
  assignedAt: Date;
  assignedBy: string; // Who submitted it?
}

interface StageExecution {
  stageNumber: number;
  stageName: string;

  // Assignment
  assignedTo: string | string[]; // User ID or role
  assignedRole: UserRole;

  // Execution
  status: "PENDING" | "APPROVED" | "REJECTED" | "REVERSED" | "SKIPPED";
  startedAt: Date;
  completedAt?: Date;
  completedBy?: string;

  // Audit
  comments?: string;
  remarks?: string; // For rejections
  signature?: string; // Base64 PNG
  validationsPerformed?: Record<string, boolean>;

  // Reversal
  reversedAt?: Date;
  reversedBy?: string;
  reversalReason?: string;
}
```

#### WorkflowDefault Type

```typescript
interface WorkflowDefault {
  // Configuration
  id: string;
  entityType: WorkflowEntityType;
  defaultWorkflowId: string; // Which workflow to use by default
  workflowVersion: number;

  // Control
  canEntityOverride: boolean; // Can individual entities pick different workflow?

  // Metadata
  effectiveDate: Date;
  deprecatedDate?: Date;
  createdBy: string;
  createdAt: Date;
}
```

### 2.2 Enhanced Existing Types

#### Requisition (Enhanced)

```typescript
interface Requisition {
  // ... existing fields ...

  // NEW: Workflow Association
  workflowId?: string; // Custom workflow if specified
  workflowAssignmentId?: string; // Reference to WorkflowAssignment

  // NEW: Approval Tracking (aligned with custom workflow)
  currentWorkflowStage?: number; // Current stage in custom workflow
  workflowStatus?: "NOT_STARTED" | "IN_PROGRESS" | "COMPLETED";
}
```

#### Budget (Enhanced)

```typescript
interface Budget {
  // ... existing fields ...

  // NEW: Workflow Association
  workflowId?: string;
  workflowAssignmentId?: string;

  // NEW: Approval Tracking
  currentWorkflowStage?: number;
  workflowStatus?: "NOT_STARTED" | "IN_PROGRESS" | "COMPLETED";
}
```

---

## Part 3: System Interactions & Flows

### 3.1 Workflow Creation Flow

```
User opens "Workflows" page
    ↓
Click "Create New Workflow"
    ↓
Workflow Designer Form:
  - Name & Description
  - Applicable Entity Types (multi-select)
  - Add Stages:
    * Stage Name
    * Approver Type (Role/User/Both)
    * Requirements (signature, comments)
    * On Approve: next stage
    * On Reject: next stage
    * Validations (if any)
    ↓
Click "Create"
    ↓
Save CustomWorkflow
  - Generate UUID
  - Version = 1
  - isActive = true
  - Store in database/file
    ↓
Success message
Redirect to workflow list
```

### 3.2 Entity Creation & Workflow Assignment Flow

```
User clicks "Create Requisition"
    ↓
Requisition Form:
  - Title, Items, etc.
  - NEW: Workflow Selection Dropdown
    * Shows applicable workflows
    * Default selected if exists
    * Can change if allowed
    ↓
Click "Save as Draft"
    ↓
Create Requisition
  - Set status = DRAFT
  - Set workflowId = selected workflow
    ↓
Create WorkflowAssignment
  - Bind requisition to workflow
  - currentStageNumber = 0 (not started)
    ↓
Store both entities
    ↓
Redirect to detail view
```

### 3.3 Submission & Approval Flow

```
User views Requisition (DRAFT status)
    ↓
Clicks "Submit for Approval"
    ↓
System:
  1. Validate requisition (items, amounts, etc.)
  2. Check if workflow is assigned
  3. If no workflow: use default for REQUISITION type
  4. Get Workflow (CustomWorkflow)
  5. Get Stage 1 from Workflow
  6. Determine approver:
     - If ROLE: find user with that role
     - If USER: use specific user
     - If ROLE_OR_USER: prefer specific, fallback to role
  7. Create Task for approver
  8. Set WorkflowAssignment.currentStageNumber = 1
  9. Set Requisition.status = SUBMITTED
     ↓
Approver gets Task notification
    ↓
Approver clicks task → Opens requisition
    ↓
Views Stage 1 details
  - Signature requirement: YES/NO
  - Comments requirement: OPTIONAL/REQUIRED/DISABLED
  - Validations to check
    ↓
Clicks "Approve"
    ↓
System:
  1. Record approval in StageExecution
  2. Get next stage from workflow.stages[1].onApprove.nextStage
  3. If nextStage == 'FINAL':
     - Set Requisition.status = APPROVED
     - Set WorkflowAssignment.complete
     - Notify requester
  4. Else:
     - Get next stage config
     - Find approver for next stage
     - Create Task for next approver
     - Set WorkflowAssignment.currentStageNumber = 2
     - Set Requisition.status = IN_APPROVAL
     ↓
Process continues until FINAL stage
```

### 3.4 Rejection Flow

```
Approver reviews Requisition
    ↓
Clicks "Reject"
    ↓
Get Stage Config
  - onReject.nextStage = 'REJECTED'
  - requiresRejectionReason = true
    ↓
System:
  1. Validate rejection reason provided
  2. Record rejection in StageExecution
  3. Set Requisition.status = REJECTED
  4. Set WorkflowAssignment.currentStageNumber = back to "needs resubmit"
  5. Notify requester of rejection
     ↓
Requester notified
    ↓
Can update requisition and resubmit
    ↓
Process restarts from Stage 1
```

### 3.5 Reversal Flow

```
Current Approver or Manager views approved/rejected stage
    ↓
If stage.canBeReversed == true:
  Clicks "Reverse"
    ↓
System:
  1. Check if user has permission to reverse
  2. Get onReverse config
  3. Record reversal in StageExecution
  4. Revert to previous stage
  5. If resetApprovals: clear subsequent stages
  6. Set status back to IN_APPROVAL
  7. Create new Task for that stage
    ↓
Approval continues from that point
```

---

## Part 4: Implementation Architecture

### 4.1 New Files to Create

```
src/types/
  ├── custom-workflow.ts          # CustomWorkflow, WorkflowStage types
  ├── workflow-assignment.ts      # WorkflowAssignment types

src/lib/
  ├── workflow-persistence.ts     # File-based storage for workflows
  ├── workflow-resolution.ts      # Logic to resolve workflow for entity
  ├── workflow-validation.ts      # Validate workflow definitions
  ├── workflow-defaults.ts        # Default workflow configuration

src/app/_actions/
  ├── workflows.ts                # CRUD for custom workflows
  ├── workflow-assignments.ts     # Manage workflow-entity binding
  ├── workflow-builder-actions.ts # Form processing for designer

src/hooks/
  ├── use-workflow-queries.ts     # Query hooks for workflows
  ├── use-workflow-assignments.ts # Query hooks for assignments

src/app/(private)/(main)/
  ├── custom/
  │   ├── page.tsx                # Workflow list page
  │   ├── create/
  │   │   └── page.tsx            # Create workflow page
  │   ├── [id]/
  │   │   └── page.tsx            # Edit workflow page
  │   └── _components/
  │       ├── workflow-list.tsx
  │       ├── workflow-designer.tsx
  │       └── stage-editor.tsx

docs/
  ├── CUSTOM_WORKFLOWS_GUIDE.md   # User guide for workflow creation
  ├── WORKFLOW_API.md             # Developer documentation
```

### 4.2 Modified Files

```
src/types/
  ├── requisition.ts              # Add workflowId, workflowAssignmentId fields
  ├── budget.ts                   # Add workflow fields
  ├── workflow.ts                 # Add WorkflowEntityType enum

src/app/_actions/
  ├── requisitions.ts             # Modify submit to use workflow system
  ├── approval.ts                 # Enhance to check CustomWorkflow

src/app/(private)/(main)/
  ├── requisitions/[id]/page.tsx  # Show workflow stage info
  ├── dashboard/page.tsx          # Add workflow stats

src/lib/
  ├── approval-config.ts          # Add fallback to CustomWorkflow
```

### 4.3 Database Schema (Future)

For when moving to real database:

```sql
CREATE TABLE custom_workflows (
  id UUID PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  description TEXT,
  version INT,
  applicable_entity_types TEXT[], -- ARRAY of types
  is_template BOOLEAN,
  is_active BOOLEAN,
  created_by UUID REFERENCES users(id),
  created_at TIMESTAMP,
  updated_by UUID REFERENCES users(id),
  updated_at TIMESTAMP,
  usage_count INT DEFAULT 0,
  last_used_at TIMESTAMP,
  deprecated_at TIMESTAMP
);

CREATE TABLE workflow_stages (
  id UUID PRIMARY KEY,
  workflow_id UUID REFERENCES custom_workflows(id),
  stage_number INT,
  stage_name VARCHAR(255),
  description TEXT,
  approver_type ENUM('ROLE', 'USER', 'ROLE_OR_USER'),
  required_role VARCHAR(100),
  specific_user_id UUID REFERENCES users(id),
  requires_signature BOOLEAN,
  comments_type ENUM('OPTIONAL', 'REQUIRED', 'DISABLED'),
  on_approve_next_stage INT,
  on_approve_set_status VARCHAR(100),
  on_reject_next_stage INT,
  on_reject_set_status VARCHAR(100),
  created_at TIMESTAMP
);

CREATE TABLE workflow_assignments (
  id UUID PRIMARY KEY,
  entity_id VARCHAR(255),
  entity_type VARCHAR(100),
  workflow_id UUID REFERENCES custom_workflows(id),
  workflow_version INT,
  current_stage_number INT,
  stage_started_at TIMESTAMP,
  assigned_at TIMESTAMP,
  assigned_by UUID REFERENCES users(id)
);

CREATE TABLE stage_executions (
  id UUID PRIMARY KEY,
  assignment_id UUID REFERENCES workflow_assignments(id),
  stage_number INT,
  assigned_to UUID REFERENCES users(id),
  status ENUM('PENDING', 'APPROVED', 'REJECTED', 'REVERSED'),
  started_at TIMESTAMP,
  completed_at TIMESTAMP,
  completed_by UUID REFERENCES users(id),
  comments TEXT,
  remarks TEXT,
  signature TEXT
);
```

---

## Part 5: Loopholes & Considerations

### 5.1 Security Loopholes to Close

#### 1. Unauthorized Workflow Creation

**Risk:** Any user creates malicious workflows
**Solution:**

- Add permission check: only ADMIN or COMPLIANCE_OFFICER can create workflows
- Audit trail for all workflow changes
- Approval gate for publishing workflows as templates

#### 2. Role-Based Assignment Circumvention

**Risk:** Attacker creates workflow with non-existent roles
**Solution:**

- Validate roles against DEMO_USERS at creation time
- Only allow roles from UserRole enum
- Don't allow custom role creation (for MVP)

#### 3. Workflow Versioning & Tampering

**Risk:** Changing workflow retroactively affects ongoing approvals
**Solution:**

- Immutable workflow versions (version number increments)
- WorkflowAssignment stores workflowVersion
- Cannot modify active versions, only create new ones
- Mark old versions as deprecated, not deleted

#### 4. Task Assignment to Wrong User

**Risk:** Workflow routes approval to wrong person
**Solution:**

- When assigning task, validate user exists and has required role
- Log all task assignments with audit trail
- Can manually reassign if needed

#### 5. State Transition Validation

**Risk:** Invalid state transitions (e.g., DRAFT → APPROVED without IN_APPROVAL)
**Solution:**

- Validate transitions against workflow definition
- Enforce strict state machine in approval logic
- Cannot skip stages (unless explicitly marked optional)

### 5.2 Business Logic Loopholes

#### 1. Infinite Loops in Stage Transitions

**Risk:** Workflow stage A → A (approval goes nowhere)
**Solution:**

- At workflow creation, validate stage order is linear
- onApprove.nextStage must be > current stage number OR 'FINAL'
- onReject.nextStage must be < current stage number OR 'REJECTED'

#### 2. Missing Final Stage

**Risk:** Workflow has 5 stages, last one doesn't have onApprove → FINAL
**Solution:**

- Validate last stage always has onApprove.nextStage = 'FINAL'
- Warn user if missing

#### 3. Orphaned Workflows

**Risk:** Workflow becomes inactive/deprecated, but entities still referencing it
**Solution:**

- Don't allow deprecating workflow with active usage
- Show usage count before allowing changes
- Provide migration tool to move entities to new workflow

#### 4. Missing Approver

**Risk:** Workflow requires FINANCE_OFFICER role, but no user has that role
**Solution:**

- At submission time, verify required approver exists
- If missing, show error "No approver found for stage X with role Y"
- Prevent submission until resolved

#### 5. Entity Status Misalignment

**Risk:** Requisition.status = APPROVED but WorkflowAssignment shows stage 2 pending
**Solution:**

- Keep Requisition.status as source of truth
- Derive workflow stage from approval chain
- Validate consistency on every read

### 5.3 Data Integrity Loopholes

#### 1. Workflow State Inconsistency

**Risk:** WorkflowAssignment and Requisition have different stage info
**Solution:**

- Use WorkflowAssignment as system of record for stages
- Keep Requisition.status for display/filtering only
- Enforce this in all approval logic

#### 2. Missing Audit Trail

**Risk:** Workflow changes don't leave audit trail
**Solution:**

- All CustomWorkflow changes logged
- All WorkflowAssignment updates recorded
- StageExecution immutable (no updates, only creates)

#### 3. Concurrent Stage Execution

**Risk:** Multiple people approve same stage simultaneously
**Solution:**

- Set stage status to IN_PROGRESS when first approval attempt starts
- Lock stage from other approvers
- Show "This stage is being processed by John Smith"
- Handle timeout (if approver leaves without finishing)

#### 4. Missing Cascading Deletes

**Risk:** Delete workflow, orphan assignments
**Solution:**

- Cannot delete active workflows (check usageCount)
- Deprecate instead of delete
- Mark as archived after 6 months of non-use

#### 5. Version Mismatch Issues

**Risk:** Entity references workflowVersion: 1, but current is version: 3
**Solution:**

- Always fetch CustomWorkflow(id, version) tuple
- Never update workflow, only create new version
- WorkflowAssignment.workflowVersion immutable
- Can compare with current version for notifications

### 5.4 Operational Loopholes

#### 1. No Workflow Rollback

**Risk:** Accidentally created bad workflow, need to undo
**Solution:**

- Keep history of all versions
- Can "revert" to previous version (creates new version marked "reverted from v3")
- Don't truly delete

#### 2. Task Management

**Risk:** Workflow assigns task, but task never completed
**Solution:**

- Task linked to WorkflowAssignment.id
- SLA tracking (slaHours from stage config)
- Escalation logic (escalationRole)
- Reminders at 24h, 12h, 1h before SLA breach

#### 3. Performance with Many Stages

**Risk:** 50-stage workflow is slow
**Solution:**

- Limit to max 20 stages (configurable)
- Validate at creation time
- Cache workflow definition (rarely changes)

#### 4. No Conditional Logic

**Risk:** "If amount > $50,000 add extra approval stage"
**Solution:**

- Out of scope for MVP
- Store placeholder for future: stage.condition?: string
- Plan for graphical condition builder later

---

## Part 6: Implementation Checklist

### Phase 1: Data Model & Types (Week 1)

- [ ] Create `src/types/custom-workflow.ts` with all types
- [ ] Create `src/types/workflow-assignment.ts`
- [ ] Update `src/types/requisition.ts` to add workflowId fields
- [ ] Update `src/types/budget.ts` to add workflow fields
- [ ] Create enum for WorkflowEntityType in workflow.ts
- [ ] Add types to `src/types/index.ts` exports

### Phase 2: Persistence Layer (Week 1)

- [ ] Create `src/lib/workflow-persistence.ts` with file storage
- [ ] Implement `saveWorkflow()` function
- [ ] Implement `getWorkflow(id, version?)` function
- [ ] Implement `listWorkflows()` function
- [ ] Implement `updateWorkflowStatus()` (deprecate/archive)
- [ ] Create `src/lib/workflow-validation.ts` for validation logic
- [ ] Implement stage order validation
- [ ] Implement role validation against UserRole enum
- [ ] Create sample workflows as JSON fixtures

### Phase 3: Core Logic (Week 2)

- [ ] Create `src/lib/workflow-resolution.ts`
  - [ ] `resolveWorkflowForEntity()` - returns CustomWorkflow for entity
  - [ ] `getDefaultWorkflow()` - gets default for entity type
  - [ ] `getApproverForStage()` - determines who approves next
- [ ] Create `src/lib/workflow-defaults.ts`
  - [ ] Store default workflow per entity type
  - [ ] Allow admin to update defaults
- [ ] Enhance `src/lib/approval-config.ts`
  - [ ] Add `getWorkflowConfig(customWorkflow)` to convert to ApprovalStageConfig
  - [ ] Add fallback logic to use CustomWorkflow if available

### Phase 4: Server Actions (Week 2)

- [ ] Create `src/app/_actions/workflows.ts`
  - [ ] `createWorkflow(data)` - validates and saves
  - [ ] `updateWorkflow(id, data)` - creates new version
  - [ ] `listWorkflows(filters)` - list with pagination
  - [ ] `getWorkflow(id)` - single workflow
  - [ ] `deprecateWorkflow(id)` - mark as inactive
- [ ] Create `src/app/_actions/workflow-assignments.ts`
  - [ ] `assignWorkflowToEntity()` - link entity to workflow
  - [ ] `getAssignmentForEntity()` - get current assignment
- [ ] Modify `src/app/_actions/requisitions.ts`
  - [ ] Update `submitRequisitionForApproval()` to use workflow system
  - [ ] Pass workflow to approval engine
- [ ] Modify `src/app/_actions/approval.ts`
  - [ ] Update approval logic to check CustomWorkflow
  - [ ] Use stage definitions for routing

### Phase 5: React Query Hooks (Week 2)

- [ ] Create `src/hooks/use-workflow-queries.ts`
  - [ ] `useWorkflows()` - list all
  - [ ] `useWorkflow(id)` - single with cache
  - [ ] `useWorkflowStats()` - usage stats
  - [ ] `useCreateWorkflow()` - mutation
  - [ ] `useUpdateWorkflow()` - mutation
  - [ ] `useDeprecateWorkflow()` - mutation
- [ ] Create `src/hooks/use-workflow-assignments.ts`
  - [ ] `useWorkflowAssignment(entityId)` - get assignment
  - [ ] `useAssignWorkflow()` - mutation

### Phase 6: UI - Workflow List (Week 3)

- [ ] Create `src/app/(private)/(main)/custom/page.tsx`
- [ ] Create `src/app/(private)/(main)/custom/_components/workflow-list.tsx`
  - [ ] Display workflows in table
  - [ ] Show name, entity types, version, usage count, created by
  - [ ] Filter by entity type, active/deprecated
  - [ ] Sort by created date, usage count
  - [ ] Actions: View, Edit, Deprecate, Clone, Delete
  - [ ] Link to create new

### Phase 7: UI - Workflow Designer (Week 3)

- [ ] Create `src/app/(private)/(main)/custom/create/page.tsx`
- [ ] Create `src/app/(private)/(main)/custom/_components/workflow-designer.tsx`
  - [ ] Form for workflow name, description
  - [ ] Multi-select for applicable entity types
  - [ ] Stages panel (add/edit/remove stages)
  - [ ] Drag-to-reorder stages (nice to have)
- [ ] Create `src/app/(private)/(main)/custom/_components/stage-editor.tsx`
  - [ ] Modal/panel for editing single stage
  - [ ] Stage name, description
  - [ ] Approver type dropdown
  - [ ] Role selection if role-based
  - [ ] User selection if user-specific
  - [ ] Signature requirement toggle
  - [ ] Comments requirement dropdown
  - [ ] On Approve: next stage, status, notify
  - [ ] On Reject: revert to, status, require reason
  - [ ] Validations multi-select (if any)
  - [ ] SLA hours input
  - [ ] Escalation role dropdown
  - [ ] Save/Cancel buttons
- [ ] Create `src/app/(private)/(main)/custom/[id]/page.tsx`
  - [ ] View/edit existing workflow
  - [ ] Show version history
  - [ ] Can only edit if not deprecated and no active usage

### Phase 8: UI - Workflow Selection (Week 3-4)

- [ ] Modify `/workflows/requisitions/create/page.tsx`
  - [ ] Add workflow selector dropdown
  - [ ] Show applicable workflows for REQUISITION type
  - [ ] Default selected if configured
  - [ ] Optional if canOverride=true
- [ ] Similar changes for Budget, PO creation pages

### Phase 9: UI - Approval Stage Display (Week 4)

- [ ] Modify `/workflows/requisitions/[id]/page.tsx`
  - [ ] Show current workflow stage
  - [ ] Show all stages with status (pending/approved/rejected)
  - [ ] Highlight current stage
  - [ ] Show approver info for each stage
  - [ ] Show stage requirements (signature, comments)
  - [ ] Show on-approve/on-reject transitions
- [ ] Update approval action panel
  - [ ] Show stage-specific requirements
  - [ ] Signature: required/optional based on workflow
  - [ ] Comments: required/optional based on workflow
  - [ ] Rejection reason: required if stage specifies

### Phase 10: UI - Admin Dashboard (Week 4)

- [ ] Create workflow management section
  - [ ] Set default workflows per entity type
  - [ ] Enable/disable custom workflows
  - [ ] View workflow usage statistics
  - [ ] Manage deprecated workflows
  - [ ] View workflow change history

### Phase 11: Integration & Testing (Week 4-5)

- [ ] End-to-end test: Create workflow → Create entity → Submit → Approve
- [ ] Test state transitions (approval → next stage)
- [ ] Test rejections (back to DRAFT)
- [ ] Test reversals
- [ ] Test with missing approver (error handling)
- [ ] Test workflow versioning
- [ ] Test role validation
- [ ] Performance test with large workflows
- [ ] Security test: unauthorized workflow creation
- [ ] Data consistency test: Requisition status ↔ WorkflowAssignment

### Phase 12: Documentation (Week 5)

- [ ] Create `docs/CUSTOM_WORKFLOWS_GUIDE.md` for users
  - [ ] Step-by-step workflow creation guide
  - [ ] Screenshots of workflow designer
  - [ ] Best practices and examples
  - [ ] Troubleshooting section
- [ ] Create `docs/WORKFLOW_API.md` for developers
  - [ ] API endpoint documentation
  - [ ] Type definitions
  - [ ] Code examples
  - [ ] Migration guide from old system

---

## Part 7: Backward Compatibility & Migration

### 7.1 Gradual Migration Strategy

**Phase A (MVP):** Workflows optional, defaults hardcoded

- New workflows can be created and used
- Existing entities default to hardcoded workflows
- No breaking changes

**Phase B (Opt-in):** Admin can set defaults

- Admin can set CustomWorkflow as default for entity type
- All new entities of that type use CustomWorkflow
- Old entities unaffected
- Grandfathered workflows still work

**Phase C (Migration):** Batch migration tool

- Provide tool to migrate old entities to new workflow
- Show preview of what will change
- Rollback capability
- Run after hours

**Phase D (Sunset):** Hardcoded workflows removed

- After 3 months, all entities using CustomWorkflows
- Remove hardcoded configs
- All workflows in CustomWorkflow system

### 7.2 Data Migration Strategy

```
// Existing requisition with implicit 3-stage workflow
Requisition {
  id: "req-001",
  status: "IN_APPROVAL",
  currentApprovalStage: 2,
  approvalChain: [...]
}

// NEW: Explicitly linked to workflow
Requisition {
  id: "req-001",
  status: "IN_APPROVAL",
  currentApprovalStage: 2,
  approvalChain: [...],
  workflowId: "wf-default-requisition",      // NEW
  workflowAssignmentId: "assign-001"         // NEW
}

// NEW: WorkflowAssignment created
WorkflowAssignment {
  id: "assign-001",
  entityId: "req-001",
  entityType: "REQUISITION",
  workflowId: "wf-default-requisition",
  workflowVersion: 1,
  currentStageNumber: 2,
  stageHistory: [
    { stageNumber: 1, status: "APPROVED", ... },
    { stageNumber: 2, status: "PENDING", ... }
  ]
}
```

---

## Part 8: Future Enhancements (Out of Scope for MVP)

1. **Conditional Approval Routing**
   - If amount > $50K, require CFO approval
   - If department = Finance, skip DEPARTMENT_MANAGER
   - Rules engine for complex logic

2. **Parallel/Concurrent Approvals**
   - Stage 2: All of DEPARTMENT_MANAGER, FINANCE_OFFICER must approve
   - Can approve in any order
   - Move to stage 3 when all done

3. **Workflow Templates & Marketplace**
   - Pre-built templates: "Standard 4-Stage", "CFO Only", "2-Approver Fast Track"
   - Clone template to customize
   - Share templates between organizations

4. **Advanced Approver Assignment**
   - Approver groups: "Finance Team" = {John, Sarah, Mike}
   - Any one from group can approve
   - Manager delegation: "When John is on leave, delegate to Sarah"

5. **SLA & Escalation Management**
   - Auto-escalate if approver takes > 48 hours
   - Notification escalation (email → SMS → call)
   - Reporting on SLA compliance

6. **Workflow Analytics**
   - Average approval time per stage
   - Bottleneck detection
   - Approval rate (how many approved vs rejected)
   - Heat maps showing slow stages

7. **Graphical Workflow Designer**
   - Drag-and-drop stage canvas
   - Visual connections between stages
   - Preview of approval flow
   - Real-time validation

8. **Workflow Simulation**
   - Simulate approval process before going live
   - Show which users would be notified
   - Estimate total approval time

9. **Workflow Versioning & A/B Testing**
   - Run two versions in parallel
   - Compare metrics
   - Gradually switch to winning version

10. **Integration with External Systems**
    - Webhook on workflow completion
    - Sync to ERP (SAP, Oracle)
    - Integration with email/Teams/Slack

---

## Part 9: Success Criteria

### MVP Definition

✅ **Minimum Viable Product**

- Users can create custom workflows (UI + backend)
- Workflows have 2-10 stages with defined transitions
- Workflows can be attached to requisitions/budgets
- Entities follow workflow approval chain
- Approvals/rejections work as defined
- No breaking changes to existing system

### Success Metrics

1. **Functional**: 95%+ workflow entities complete approval process successfully
2. **Performance**: Workflow resolution < 100ms
3. **Usability**: New user can create workflow in < 10 minutes
4. **Reliability**: Zero data consistency issues between entity status and workflow stage
5. **Security**: All workflow operations audit-logged
6. **Adoption**: 50%+ of requisitions use custom workflows within 1 month

### Testing Coverage

- Unit tests: 80% coverage for workflow-resolution.ts, workflow-validation.ts
- Integration tests: Complete create → submit → approve flow
- E2E tests: User journey from workflow creation to approval
- Security tests: Authorization, validation, SQL injection (when DB added)
- Load tests: 1000 workflows, 10,000 assignments

---

## Part 10: Risk Assessment

| Risk                                               | Likelihood | Impact | Mitigation                                                 |
| -------------------------------------------------- | ---------- | ------ | ---------------------------------------------------------- |
| Data inconsistency between entity & workflow stage | Medium     | High   | Enforce WorkflowAssignment as SSOT, validate on every read |
| Missing approver at runtime                        | Medium     | High   | Validate approver exists at submission time                |
| Infinite loop in stage transitions                 | Low        | High   | Validate stage order at creation time                      |
| Performance degradation with many workflows        | Low        | Medium | Cache workflow config, limit to 20 stages                  |
| Unauthorized workflow creation                     | Medium     | Medium | Permission checks (ADMIN only) + audit trail               |
| Workflow tampering                                 | Low        | High   | Immutable versions, deprecate don't delete                 |
| Task assignment failures                           | Medium     | Medium | Implement retry logic, dead letter queue                   |
| Backward compatibility issues                      | Low        | High   | Gradual migration strategy, keep defaults                  |

---

## Conclusion

This design provides a **flexible, secure, and scalable** custom workflow management system that:

1. ✅ Maintains backward compatibility
2. ✅ Empowers users to design their own workflows
3. ✅ Closes security loopholes with validation and audit trails
4. ✅ Handles edge cases and data consistency
5. ✅ Provides clear migration path to production database
6. ✅ Enables future enhancements (parallel approvals, conditions, etc.)

The implementation is structured in 12 phases over 5 weeks, with each phase delivering testable, working functionality.

---

**Document Version:** 1.0
**Created:** 2024
**Last Updated:** 2024
**Status:** Ready for Implementation
