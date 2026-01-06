# How the Enhanced Workflow System Works

## 🎯 **Overview**

The enhanced workflow system provides a flexible, database-driven approach to document approvals that combines **Document Automation** with **Custom Workflow Management**. Here's how it works from setup to execution.

## 📋 **Table of Contents**

1. [System Components](#system-components)
2. [Workflow Setup Process](#workflow-setup-process)
3. [Document Approval Flow](#document-approval-flow)
4. [Automation Integration](#automation-integration)
5. [User Experience](#user-experience)
6. [Administrative Management](#administrative-management)

---

## 🏗️ **System Components**

### **Core Models**
```
Workflow → WorkflowAssignment → WorkflowTask
    ↓              ↓               ↓
 Definition    Execution       Individual Tasks
```

- **Workflow**: Template defining approval stages and conditions
- **WorkflowAssignment**: Tracks execution for a specific document
- **WorkflowTask**: Individual approval tasks assigned to users/roles
- **WorkflowDefault**: Default workflow mappings per entity type

---

## ⚙️ **Workflow Setup Process**

### **Step 1: Admin Creates Workflows**

An organization admin uses the frontend UI to create workflows:

```typescript
// Example: Creating a Requisition Workflow
const requisitionWorkflow = {
  name: "Standard Requisition Approval",
  description: "Standard approval process for requisitions under $10,000",
  entityType: "requisition",
  conditions: {
    amountRange: { min: 0, max: 10000 },
    departments: ["IT", "Marketing", "Operations"]
  },
  stages: [
    {
      stageNumber: 1,
      stageName: "Department Manager Review",
      requiredRole: "DEPARTMENT_MANAGER",
      requiredApprovals: 1,
      timeoutHours: 24,
      canReject: true,
      canReassign: true
    },
    {
      stageNumber: 2,
      stageName: "Finance Review",
      requiredRole: "FINANCE_OFFICER",
      requiredApprovals: 1,
      timeoutHours: 48,
      canReject: true,
      canReassign: false
    }
  ],
  isDefault: true
}
```

### **Step 2: System Stores Workflow**

The backend processes and stores the workflow:

```go
// Backend creates workflow record
workflow := &models.Workflow{
    ID:             "wf-req-standard-001",
    OrganizationID: "org-123",
    Name:           "Standard Requisition Approval",
    EntityType:     "requisition",
    Version:        1,
    IsActive:       true,
    IsDefault:      true,
    Stages:         [JSON array of stages],
    Conditions:     [JSON conditions],
    CreatedBy:      "admin-user-id",
}
```

### **Step 3: Default Workflow Assignment**

If marked as default, the system creates a default mapping:

```go
// Default workflow record
defaultRecord := &models.WorkflowDefault{
    OrganizationID:         "org-123",
    EntityType:             "requisition",
    DefaultWorkflowID:      "wf-req-standard-001",
    DefaultWorkflowVersion: 1,
    SetBy:                  "admin-user-id",
}
```

---

## 📄 **Document Approval Flow**

### **Phase 1: Document Creation & Workflow Assignment**

When a user creates a document (e.g., requisition):

```go
// 1. User creates requisition
requisition := CreateRequisition(...)

// 2. System resolves appropriate workflow
workflow, err := workflowService.ResolveWorkflowForEntity(
    organizationID, 
    "requisition", 
    requisition,
)

// 3. System creates workflow assignment
assignment := &models.WorkflowAssignment{
    ID:              "wa-req-001",
    OrganizationID:  "org-123",
    EntityID:        requisition.ID,
    EntityType:      "requisition",
    WorkflowID:      workflow.ID,
    CurrentStage:    0, // Not started yet
    Status:          "in_progress",
    AssignedBy:      userID,
}
```

### **Phase 2: Workflow Initiation**

When the document is submitted for approval:

```go
// 1. Move to first stage
assignment.CurrentStage = 1

// 2. Create approval task for first stage
task := &models.WorkflowTask{
    ID:                   "wt-req-001-stage1",
    OrganizationID:       "org-123",
    WorkflowAssignmentID: assignment.ID,
    EntityID:             requisition.ID,
    EntityType:           "requisition",
    StageNumber:          1,
    StageName:            "Department Manager Review",
    AssignmentType:       "role",
    AssignedRole:         "DEPARTMENT_MANAGER",
    Status:               "pending",
    Priority:             "medium",
    DueDate:              time.Now().Add(24 * time.Hour),
}

// 3. Notify eligible approvers
notifyUsersWithRole("DEPARTMENT_MANAGER", task)
```

### **Phase 3: Approval Process**

When an approver takes action:

```go
// 1. User with DEPARTMENT_MANAGER role approves
func ApproveTask(taskID, approverID, comments, signature string) {
    // Validate approver has required role
    if !hasRequiredRole(approverID, task.AssignedRole) {
        return error("Insufficient permissions")
    }
    
    // Record stage execution
    execution := StageExecution{
        StageNumber:  1,
        StageName:    "Department Manager Review",
        ApproverID:   approverID,
        ApproverName: "John Smith",
        ApproverRole: "DEPARTMENT_MANAGER",
        Action:       "approved",
        Comments:     comments,
        Signature:    signature,
        ExecutedAt:   time.Now(),
    }
    
    // Add to assignment history
    assignment.AddStageExecution(execution)
    
    // Move to next stage
    assignment.CurrentStage = 2
    
    // Create next task
    createTaskForStage(assignment, 2)
}
```

### **Phase 4: Stage Progression**

The system automatically progresses through stages:

```go
// Stage 2: Finance Review
task2 := &models.WorkflowTask{
    StageNumber:    2,
    StageName:      "Finance Review",
    AssignedRole:   "FINANCE_OFFICER",
    Status:         "pending",
    DueDate:       time.Now().Add(48 * time.Hour),
}

// When finance officer approves...
// If this was the final stage:
if assignment.CurrentStage >= len(workflow.Stages) {
    assignment.Status = "completed"
    assignment.CompletedAt = &now
    
    // Trigger document automation if configured
    triggerDocumentAutomation(requisition)
}
```

---

## 🤖 **Automation Integration**

The workflow system integrates with the existing document automation:

### **Automation Trigger Points**

```go
// After workflow completion
func (s *WorkflowService) CompleteWorkflow(assignment *WorkflowAssignment) {
    // Mark workflow as completed
    assignment.Status = "completed"
    
    // Trigger automation based on entity type
    switch assignment.EntityType {
    case "requisition":
        // Auto-create Purchase Order
        automationService.CreatePurchaseOrderFromRequisition(assignment.EntityID)
        
    case "purchase_order":
        // Auto-create GRN
        automationService.CreateGRNFromPurchaseOrder(assignment.EntityID)
        
    case "grn":
        // Auto-create Payment Voucher
        automationService.CreatePaymentVoucherFromGRN(assignment.EntityID)
    }
}
```

### **Automation Chain Example**

```
Requisition Created
       ↓
   Workflow Assigned (Standard Requisition Approval)
       ↓
   Stage 1: Department Manager → APPROVED
       ↓
   Stage 2: Finance Officer → APPROVED
       ↓
   Workflow Complete → AUTO-CREATE Purchase Order
       ↓
   PO Workflow Assigned (Standard PO Approval)
       ↓
   Stage 1: Procurement Officer → APPROVED
       ↓
   PO Workflow Complete → AUTO-CREATE GRN
       ↓
   GRN Workflow Assigned (Standard GRN Approval)
       ↓
   Stage 1: Warehouse Manager → APPROVED
       ↓
   GRN Workflow Complete → AUTO-CREATE Payment Voucher
```

---

## 👥 **User Experience**

### **For Document Creators**

1. **Create Document**: User creates requisition through normal UI
2. **Submit for Approval**: Document automatically enters workflow
3. **Track Progress**: User can see current stage and approval history
4. **Receive Notifications**: Updates on approvals, rejections, completions

### **For Approvers**

1. **Receive Notifications**: Email/in-app notifications for pending tasks
2. **View Approval Queue**: Dashboard showing all pending approvals
3. **Review Documents**: Access to document details and history
4. **Take Action**: Approve, reject, or reassign with comments/signatures
5. **Track Workload**: See approval statistics and overdue items

### **Example Approval Dashboard**

```typescript
// Approver sees pending tasks
const pendingTasks = [
  {
    id: "wt-req-001-stage1",
    entityType: "requisition",
    entityId: "req-001",
    stageName: "Department Manager Review",
    priority: "medium",
    dueDate: "2025-01-07T10:00:00Z",
    document: {
      title: "Office Supplies Request",
      amount: 2500,
      department: "IT"
    }
  }
]
```

---

## 🔧 **Administrative Management**

### **Workflow Management**

Admins can:

1. **Create Multiple Workflows**: Different processes for different scenarios
2. **Set Conditions**: Automatic workflow selection based on document properties
3. **Manage Defaults**: Set default workflows per entity type
4. **Version Control**: Update workflows without breaking existing assignments
5. **Usage Tracking**: See which workflows are being used
6. **Duplicate Workflows**: Copy existing workflows for customization

### **Monitoring & Analytics**

```go
// Admin dashboard data
type WorkflowAnalytics struct {
    TotalWorkflows      int
    ActiveAssignments   int
    CompletedToday      int
    AverageApprovalTime time.Duration
    OverdueTasks        int
    TopBottlenecks      []string
}
```

### **Workflow Conditions Examples**

```typescript
// Different workflows based on conditions
const workflows = [
  {
    name: "Small Purchase Workflow",
    conditions: { amountRange: { max: 1000 } },
    stages: [{ stageName: "Manager Approval" }] // Single stage
  },
  {
    name: "Large Purchase Workflow", 
    conditions: { amountRange: { min: 1000, max: 10000 } },
    stages: [
      { stageName: "Manager Approval" },
      { stageName: "Finance Approval" }
    ] // Two stages
  },
  {
    name: "Capital Expenditure Workflow",
    conditions: { amountRange: { min: 10000 } },
    stages: [
      { stageName: "Manager Approval" },
      { stageName: "Finance Approval" },
      { stageName: "CFO Approval" }
    ] // Three stages
  }
]
```

---

## 🔄 **Complete Flow Example**

### **Scenario**: IT Department requests office supplies worth $2,500

```
1. USER ACTION: John (IT Staff) creates requisition
   └─ Amount: $2,500, Department: IT, Items: Laptops & Monitors

2. SYSTEM: Resolves workflow
   └─ Matches "Standard Requisition Approval" (amount < $10,000, dept = IT)

3. SYSTEM: Creates workflow assignment
   └─ Assignment ID: wa-req-12345
   └─ Current Stage: 0 (not started)

4. USER ACTION: John submits requisition for approval
   └─ Assignment moves to Stage 1
   └─ Task created for DEPARTMENT_MANAGER role
   └─ Notification sent to IT managers

5. APPROVER ACTION: Sarah (IT Manager) receives notification
   └─ Reviews requisition details
   └─ Approves with signature and comments
   └─ Stage 1 execution recorded

6. SYSTEM: Progresses to Stage 2
   └─ Task created for FINANCE_OFFICER role
   └─ Notification sent to finance team

7. APPROVER ACTION: Mike (Finance Officer) approves
   └─ Final stage completed
   └─ Workflow marked as "completed"

8. AUTOMATION: System auto-creates Purchase Order
   └─ PO-12345 created from approved requisition
   └─ PO enters its own workflow process

9. NOTIFICATIONS: All parties notified
   └─ John: "Your requisition has been approved"
   └─ Sarah & Mike: "Approval process completed"
   └─ Procurement: "New PO ready for processing"
```

---

## 🎯 **Key Benefits**

### **Flexibility**
- Multiple workflows per document type
- Condition-based automatic selection
- Easy workflow modification without code changes

### **Transparency**
- Complete audit trail of all approvals
- Real-time status tracking
- Clear approval queues and notifications

### **Efficiency**
- Automated workflow assignment
- Parallel processing where appropriate
- Integration with document automation

### **Control**
- Role-based approval validation
- Configurable timeouts and escalations
- Usage tracking and analytics

### **Scalability**
- Database-driven configuration
- Version control for workflows
- Organization-specific customization

---

This enhanced workflow system provides a robust foundation for managing complex approval processes while maintaining the simplicity and automation that users expect. The combination of flexible workflow configuration and intelligent document automation creates a powerful business process management solution.