# Phase 3+: Workflow Stages with Role-Based Access

## 🎯 Problem Statement

Organizations need to create custom approval workflows where:
1. **Multiple stages** in a workflow (e.g., Department Manager → Finance → Director)
2. **Each stage requires specific role(s)** (e.g., only "Finance Approver" can approve at Finance stage)
3. **Different organizations have different workflows** (Org A: 2 stages, Org B: 4 stages)
4. **Admins can define workflows** without coding
5. **Workflows apply to different document types** (Requisitions, Purchase Orders, Budgets, etc.)

---

## 🏗️ Current Workflow System

### Current State
```
Requisition
├─ approvalStage: int (1, 2, 3, ...)
├─ approvalHistory: JSON
│  └─ [{stage: 1, approver: user_id, status: approved, date: ...}]
└─ status: "draft" | "pending" | "approved" | "rejected"
```

### Problem with Current System
- No workflow definition
- Approval stages are hardcoded (magic numbers)
- No role requirements per stage
- No way to configure who can approve at each stage
- Not flexible for different organization needs

---

## ✅ Proposed Solution

### New Data Model

#### WorkflowTemplate (Org-Scoped)
```sql
CREATE TABLE workflow_templates (
    id UUID PRIMARY KEY,
    organization_id UUID NOT NULL REFERENCES organizations(id),

    -- Workflow definition
    name VARCHAR(255) NOT NULL,                      -- "Requisition Approval"
    description TEXT,
    document_type VARCHAR(100) NOT NULL,             -- "requisition", "purchase_order", "budget"

    -- Configuration
    is_active BOOLEAN DEFAULT true,
    requires_all_stages BOOLEAN DEFAULT true,        -- Must complete ALL stages?
    allow_skip_stages BOOLEAN DEFAULT false,         -- Can admin skip stages?
    notify_on_rejection BOOLEAN DEFAULT true,        -- Notify requester on reject?

    -- Metadata
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP,

    UNIQUE(organization_id, document_type),
    INDEX(organization_id),
    INDEX(document_type)
);
```

#### WorkflowStage (Part of WorkflowTemplate)
```sql
CREATE TABLE workflow_stages (
    id UUID PRIMARY KEY,
    workflow_id UUID NOT NULL REFERENCES workflow_templates(id),

    -- Stage definition
    sequence_number INT NOT NULL,                    -- 1, 2, 3, ...
    name VARCHAR(255) NOT NULL,                      -- "Department Approval"
    description TEXT,

    -- Role requirements
    required_role_ids UUID[] NOT NULL,               -- Array of OrganizationRole IDs
    -- OR for backward compat:
    required_roles VARCHAR[] NOT NULL,               -- ["approver", "finance"]

    -- Configuration
    allow_multiple_approvers BOOLEAN DEFAULT false,  -- All roles must approve?
    allow_skip BOOLEAN DEFAULT false,                -- Can this stage be skipped?
    auto_approve BOOLEAN DEFAULT false,              -- Auto-approve without human?

    -- Optional: Conditions for stage
    min_amount DECIMAL(12,2),                        -- Only if amount > min_amount
    max_amount DECIMAL(12,2),                        -- Only if amount < max_amount
    department_ids UUID[],                           -- Only for specific departments

    -- Optional: Notifications
    notify_before_days INT,                          -- Notify reviewer N days before due
    due_days INT,                                    -- Stage must complete in N days

    -- Metadata
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    UNIQUE(workflow_id, sequence_number),
    INDEX(workflow_id)
);
```

#### WorkflowApproval (Runtime Instance)
```sql
CREATE TABLE workflow_approvals (
    id UUID PRIMARY KEY,

    -- Reference to document and workflow
    document_type VARCHAR(100) NOT NULL,             -- "requisition", "purchase_order"
    document_id UUID NOT NULL,                       -- ID of the document
    workflow_id UUID NOT NULL REFERENCES workflow_templates(id),
    stage_id UUID NOT NULL REFERENCES workflow_stages(id),

    -- Approval status
    status VARCHAR(50) NOT NULL,                     -- "pending", "approved", "rejected", "skipped"
    current_stage_number INT NOT NULL,

    -- Approval details
    assigned_roles VARCHAR[] NOT NULL,               -- Roles required for this stage
    assigned_to_users UUID[],                        -- (Optional) Specific users if assigned

    -- Completion
    approved_by UUID REFERENCES users(id),
    approved_at TIMESTAMP,
    rejection_reason TEXT,
    rejected_at TIMESTAMP,

    -- Metadata
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    UNIQUE(document_id, workflow_id, stage_id),
    INDEX(document_id),
    INDEX(status),
    INDEX(workflow_id)
);
```

---

## 📊 Example: Multi-Stage Requisition Workflow

### Organization "Tech Corp" Creates Workflow

**Workflow: "Requisition Approval"**

```
Stage 1: Department Manager Review
├─ Required Role: "Manager" or "Senior Manager"
├─ Allow Multiple: false (one person sufficient)
└─ Conditions: None

Stage 2: Finance Approval
├─ Required Role: "Finance Approver" or "Finance Manager"
├─ Allow Multiple: true (all must approve)
├─ Min Amount: $0
└─ Max Amount: $50,000

Stage 3: Director Approval
├─ Required Role: "Director" or "CFO"
├─ Allow Multiple: false
└─ Min Amount: $50,000  (only if amount > $50k)

Stage 4: CEO Sign-off
├─ Required Role: "CEO" or "COO"
├─ Allow Multiple: false
└─ Min Amount: $100,000  (only if amount > $100k)
```

### Requisition Flows Through Stages

```
User Creates Requisition ($25,000)
    ↓
Stage 1: Manager Reviews
    └─ Checks: Is user a "Manager"? ✓
    └─ Manager approves
    ↓
Stage 2: Finance Reviews
    └─ Checks: Is user "Finance Approver"? ✓
    └─ Finance Approver approves
    ↓
Stage 3: Skip (amount < $50,000)
    └─ Automatically skip to next
    ↓
Stage 4: Skip (amount < $100,000)
    └─ Automatically skip (doesn't exist for this amount)
    ↓
Status: APPROVED ✓
```

---

## 🔄 Data Flow

### 1. Admin Defines Workflow (Setup)

```
Admin goes to:
  /organizations/{orgId}/settings/workflows/requisition

Creates workflow:
  Name: "Requisition Approval"
  Document Type: "requisition"

  Adds Stage 1:
    Name: "Manager Review"
    Sequence: 1
    Required Roles: ["Manager", "Senior Manager"]
    Allow Multiple: false
    Conditions: none

  Adds Stage 2:
    Name: "Finance Review"
    Sequence: 2
    Required Roles: ["Finance Approver"]
    Allow Multiple: true
    Min Amount: 0
    Max Amount: 50000

  Adds Stage 3:
    Name: "Director Review"
    Sequence: 3
    Required Roles: ["Director"]
    Allow Multiple: false
    Min Amount: 50000

  Saves workflow
```

### 2. User Creates Document

```
User creates Requisition:
  - Amount: $25,000
  - Department: "Marketing"

System:
  1. Look up workflow for "requisition" in org
  2. Evaluate all stages (conditions)
  3. Find applicable stages:
     - Stage 1: ✓ (no conditions)
     - Stage 2: ✓ (amount <= 50,000)
     - Stage 3: ✗ (amount < 50,000)
  4. Create WorkflowApproval for Stage 1
  5. Set status to "pending_approval"
  6. Notify Manager to approve
```

### 3. User Approves at Current Stage

```
Manager gets task:
  "Review Requisition #12345 for Marketing"

Manager:
  1. Reviews requisition
  2. Checks: Do I have "Manager" role? ✓
  3. Clicks "Approve"

System:
  1. Verify user has required role
  2. Create approval record
  3. Mark Stage 1 as "approved"
  4. Get next stage (Stage 2)
  5. Create WorkflowApproval for Stage 2
  6. Set status to "pending_approval"
  7. Notify Finance to approve
```

### 4. User Rejects at Current Stage

```
Finance Approver rejects:
  "Amount exceeds budget for this quarter"

System:
  1. Verify user has required role
  2. Create approval record with "rejected" status
  3. Mark Stage 2 as "rejected"
  4. Set document status to "rejected"
  5. Notify requester of rejection
  6. Allow requester to edit and resubmit
  7. When resubmitted, start from Stage 1 again
```

---

## 🔐 Permission Model

### Who Can Do What?

```
User CAN approve at Stage 1 IF:
  1. Document status = "pending_approval"
  2. Current stage = Stage 1
  3. User's org role ∈ Stage 1 required roles
  4. (Optional) Stage condition met (amount, department, etc.)

Example:
  Document: Requisition #123, Amount: $25,000
  Current Stage: Stage 2 (Finance)
  User: Finance Approver (has required role)

  User CAN approve ✓
  User CANNOT skip to Stage 3 (only admin can)
  User CANNOT approve at Stage 1 (already done)
```

---

## 💻 Backend Implementation

### API Endpoints

#### Workflow Management

```
GET    /api/v1/organizations/{orgId}/workflows
       └─ List all workflows in org

POST   /api/v1/organizations/{orgId}/workflows
       ├─ name: string
       ├─ description: string
       ├─ documentType: string
       ├─ requiresAllStages: boolean
       └─ stages: [...]

GET    /api/v1/organizations/{orgId}/workflows/{workflowId}
       └─ Get workflow with all stages

PATCH  /api/v1/organizations/{orgId}/workflows/{workflowId}
       └─ Update workflow

DELETE /api/v1/organizations/{orgId}/workflows/{workflowId}
       └─ Delete workflow

---

POST   /api/v1/organizations/{orgId}/workflows/{workflowId}/stages
       ├─ name: string
       ├─ sequence: int
       ├─ requiredRoles: [roleIds]
       ├─ allowMultiple: boolean
       ├─ conditions: {minAmount?, maxAmount?, departments?}
       └─ notifications: {dueInDays?, notifyBeforeDays?}

PATCH  /api/v1/organizations/{orgId}/workflows/{workflowId}/stages/{stageId}
       └─ Update stage

DELETE /api/v1/organizations/{orgId}/workflows/{workflowId}/stages/{stageId}
       └─ Delete stage
```

#### Document Approval

```
GET    /api/v1/documents/{documentId}/workflow
       └─ Get current approval status

POST   /api/v1/documents/{documentId}/approve
       ├─ stageId: UUID
       └─ notes: string

POST   /api/v1/documents/{documentId}/reject
       ├─ stageId: UUID
       └─ reason: string

POST   /api/v1/documents/{documentId}/skip-stage
       ├─ stageId: UUID
       └─ reason: string (admin only)
```

### Key Business Logic

```go
// backend/services/workflow_service.go

func (ws *WorkflowService) GetApplicableStages(
    document *models.Requisition,
    workflow *models.WorkflowTemplate,
) []models.WorkflowStage {
    applicableStages := []models.WorkflowStage{}

    for _, stage := range workflow.Stages {
        // Check amount conditions
        if stage.MinAmount != nil && document.TotalAmount < *stage.MinAmount {
            continue
        }
        if stage.MaxAmount != nil && document.TotalAmount > *stage.MaxAmount {
            continue
        }

        // Check department conditions
        if len(stage.DepartmentIds) > 0 {
            if !contains(stage.DepartmentIds, document.Department) {
                continue
            }
        }

        applicableStages = append(applicableStages, stage)
    }

    return applicableStages
}

func (ws *WorkflowService) CanUserApproveStage(
    user *models.User,
    stage *models.WorkflowStage,
    orgId string,
) bool {
    // Get user's role in organization
    member, err := ws.db.GetOrganizationMember(user.ID, orgId)
    if err != nil {
        return false
    }

    // Get user's role permissions
    userRole, err := ws.db.GetOrganizationRole(member.OrganizationRoleId)
    if err != nil {
        return false
    }

    // Check if user's role is in required roles for stage
    for _, requiredRoleId := range stage.RequiredRoleIds {
        if userRole.ID == requiredRoleId {
            return true
        }
    }

    return false
}

func (ws *WorkflowService) ApproveStage(
    user *models.User,
    document *models.Requisition,
    stage *models.WorkflowStage,
) error {
    // 1. Verify user can approve
    if !ws.CanUserApproveStage(user, stage, document.OrganizationID) {
        return fmt.Errorf("user does not have permission to approve")
    }

    // 2. Create approval record
    approval := &models.WorkflowApproval{
        DocumentType: "requisition",
        DocumentId: document.ID,
        WorkflowId: stage.WorkflowId,
        StageId: stage.ID,
        Status: "approved",
        ApprovedBy: user.ID,
        ApprovedAt: time.Now(),
    }

    if err := ws.db.Create(approval).Error; err != nil {
        return err
    }

    // 3. Get next stage
    nextStage, err := ws.db.GetNextStage(stage.WorkflowId, stage.SequenceNumber)
    if err != nil && err != gorm.ErrRecordNotFound {
        return err
    }

    if nextStage != nil {
        // 4. Create approval record for next stage
        nextApproval := &models.WorkflowApproval{
            DocumentType: "requisition",
            DocumentId: document.ID,
            WorkflowId: stage.WorkflowId,
            StageId: nextStage.ID,
            Status: "pending",
        }
        ws.db.Create(nextApproval)
    } else {
        // 5. No more stages - mark document as approved
        ws.db.Model(document).Update("status", "approved")
    }

    return nil
}
```

---

## 🎨 Frontend Implementation

### UI Workflow Builder

```typescript
// /organizations/{orgId}/settings/workflows

<WorkflowBuilder>
  <WorkflowName />
  <DocumentTypeSelector />

  <StagesContainer>
    {stages.map(stage => (
      <StageCard key={stage.id}>
        <h4>{stage.name}</h4>

        <RoleSelector
          label="Required Roles"
          selectedRoles={stage.requiredRoles}
          onUpdate={updateStageRoles}
        />

        <CheckboxGroup>
          <Checkbox
            label="Allow Multiple Approvers"
            checked={stage.allowMultiple}
          />
          <Checkbox
            label="Allow Skip This Stage"
            checked={stage.allowSkip}
          />
        </CheckboxGroup>

        <ConditionsBuilder>
          <NumberInput
            label="Min Amount"
            value={stage.minAmount}
          />
          <NumberInput
            label="Max Amount"
            value={stage.maxAmount}
          />
          <DepartmentSelector
            label="Only for Departments"
            selected={stage.departments}
          />
        </ConditionsBuilder>

        <Button onClick={() => deleteStage(stage.id)}>
          Delete Stage
        </Button>
      </StageCard>
    ))}

    <Button onClick={addStage}>+ Add Stage</Button>
  </StagesContainer>

  <Button onClick={saveWorkflow}>Save Workflow</Button>
</WorkflowBuilder>
```

### UI Approval Task

```typescript
// /approvals or /documents/{id}/approve

<ApprovalCard>
  <DocumentDetails document={document} />

  <CurrentStage>
    <h3>{currentStage.name}</h3>
    <p>Step {currentStage.sequence} of {totalStages}</p>

    <RoleIndicator
      requiredRoles={currentStage.requiredRoles}
      userRole={userRole}
    />
  </CurrentStage>

  <DocumentPreview document={document} />

  <ApprovalForm>
    <TextArea
      label="Approval Notes"
      placeholder="Add notes for record"
    />

    <ButtonGroup>
      <Button
        onClick={approve}
        variant="success"
      >
        Approve
      </Button>

      <Button
        onClick={reject}
        variant="danger"
      >
        Reject
      </Button>
    </ButtonGroup>
  </ApprovalForm>

  <ApprovalHistory>
    {approvals.map(approval => (
      <ApprovalRecord key={approval.id}>
        <span>{approval.stage.name}</span>
        <span>{approval.approvedBy.name}</span>
        <span>{approval.status}</span>
        <span>{format(approval.approvedAt)}</span>
      </ApprovalRecord>
    ))}
  </ApprovalHistory>
</ApprovalCard>
```

---

## 📋 Workflow Examples by Organization Type

### Small Startup
```
Workflow: Quick Approval
├─ Stage 1: Owner Approval (required role: owner)
└─ Done
```

### Growing Company
```
Workflow: Standard
├─ Stage 1: Department Manager (required: manager)
├─ Stage 2: Finance Review (required: finance_approver)
└─ Stage 3: CEO Approval (required: ceo, min_amount: $10,000)
```

### Enterprise
```
Workflow: Comprehensive
├─ Stage 1: Department Head (required: dept_head)
├─ Stage 2: Budget Committee (required: [budget_officer, finance_lead], all must approve)
├─ Stage 3: Vendor Review (required: procurement, min_amount: $5,000)
├─ Stage 4: Legal Review (required: legal, only if vendor_type: "new")
├─ Stage 5: CFO Approval (required: cfo, min_amount: $50,000)
└─ Stage 6: Board Approval (required: board_member, min_amount: $500,000)
```

---

## 🔄 Integration with Phase 3

### Phase 3 Still Needed
- ✅ Permission service (what users CAN do)
- ✅ Organization roles
- ✅ Organization permissions

### Phase 3.5+ Additions
- ✅ WorkflowTemplate and WorkflowStage
- ✅ Workflow builder UI
- ✅ Approval task assignment
- ✅ Workflow engine

### Where They Connect
```
Phase 3: Permissions System
├─ Defines what user.role can do globally
└─ "admin can do X, approver can do Y"

Phase 3.5+: Workflow System
├─ Uses permissions system
├─ "Stage 1 requires approver role"
├─ At runtime: "Does user have approver role?"
└─ If yes: "Can approve stage"
```

---

## 🎯 Key Features

### For Organization Admins
✅ Define workflows without coding
✅ Create multi-stage approval processes
✅ Set role requirements per stage
✅ Add conditional logic (amount, department)
✅ View all workflows and stats
✅ Modify workflows anytime

### For Document Creators
✅ Documents route through workflow automatically
✅ See current approval status
✅ Get notifications on approval/rejection
✅ Resubmit if rejected
✅ Track approval history

### For Approvers
✅ See all documents awaiting their approval
✅ View document details
✅ Approve/reject with notes
✅ See approval chain
✅ Bulk actions (optional)

---

## 📅 Implementation Timeline

### Phase 3 (Now): 4-6 hours
- Hardcoded permission mapping
- No workflow stages yet

### Phase 3.5: 12-16 hours
- WorkflowTemplate model
- WorkflowStage model
- Basic API endpoints
- Workflow builder UI
- Document routing logic

### Phase 4+: 20+ hours
- Advanced conditions
- Role inheritance
- Custom notifications
- Workflow analytics
- Approval dashboard

---

## ✨ Summary

**With this system, organization admins can:**

1. ✅ Create custom roles (Phase 3.5+)
2. ✅ Assign permissions to roles (Phase 3.5+)
3. ✅ Create approval workflows (Phase 3.5+)
4. ✅ Add multiple stages (Phase 3.5+)
5. ✅ Assign roles to each stage (Phase 3.5+)
6. ✅ Set stage conditions (amount, department) (Phase 3.5+)
7. ✅ Modify workflows anytime (Phase 3.5+)

**Without touching any code!**

---

**Next Steps**:
1. Complete Phase 3 (permissions)
2. Plan Phase 3.5 (custom roles)
3. Design Phase 3.5+ (workflow system)
4. Implement Phase 4 (advanced features)

