# Workflow Design Decisions & Clarifications

## Overview

Based on your excellent questions, I'll provide detailed clarifications about the workflow system design decisions and implementation approach. The current system has both **implemented features** and **planned architecture** that need to be clearly distinguished.

## 🎯 **Your Questions Answered**

### 1. **Who defines the workflow processes for each resource? (e.g., Requisition)**

#### Current Implementation:
- **System-Defined**: Currently, workflows are hardcoded in the business logic
- **Fixed Process**: Each document type has a predefined approval flow
- **No Configuration**: Users cannot modify the workflow steps

#### Planned Architecture:
```typescript
// Who can define workflows (in order of priority):
interface WorkflowDefinitionRoles {
  systemAdmin: boolean;      // Can create system-wide templates
  organizationAdmin: boolean; // Can create org-specific workflows
  departmentManager: boolean; // Can customize dept workflows
  workflowDesigner: boolean;  // Dedicated workflow design role
}

// Workflow creation permissions
const workflowCreationMatrix = {
  "system-admin": ["all-entity-types", "all-organizations"],
  "org-admin": ["requisition", "budget", "purchase-order", "payment-voucher"],
  "dept-manager": ["requisition", "budget"], // Limited scope
  "workflow-designer": ["assigned-entity-types"] // Role-based access
};
```

#### **Recommendation**: 
- **Phase 1**: Organization Admins define workflows
- **Phase 2**: Add department-level customization
- **Phase 3**: Role-based workflow design permissions

### 2. **How many workflows can exist for one resource?**

#### Current Implementation:
- **One Workflow**: Each document type has exactly one hardcoded workflow
- **No Variants**: No support for different approval paths

#### Planned Architecture:
```typescript
interface WorkflowVariants {
  // Multiple workflows per entity type
  entityType: "requisition";
  workflows: [
    {
      id: "req-standard",
      name: "Standard Requisition Approval",
      conditions: { amount: { max: 10000 } },
      isDefault: true
    },
    {
      id: "req-high-value", 
      name: "High-Value Requisition Approval",
      conditions: { amount: { min: 10000 } },
      isDefault: false
    },
    {
      id: "req-emergency",
      name: "Emergency Requisition",
      conditions: { priority: "urgent" },
      isDefault: false
    }
  ]
}

// Workflow selection logic
interface WorkflowSelectionCriteria {
  amount?: { min?: number; max?: number };
  department?: string[];
  priority?: "low" | "medium" | "high" | "urgent";
  category?: string[];
  vendorType?: string[];
  customFields?: Record<string, any>;
}
```

#### **Recommendation**:
- **Unlimited workflows per resource type**
- **Condition-based selection** (amount, department, priority)
- **Default workflow** as fallback
- **Workflow versioning** for changes over time

### 3. **How many approvals are required per stage?**

#### Current Implementation:
```go
// Simple single-approver model
type ApprovalRecord struct {
    ApproverID   string    `json:"approverId"`
    ApproverName string    `json:"approverName"`
    Status       string    `json:"status"` // approved/rejected
    Comments     string    `json:"comments"`
    Signature    string    `json:"signature"`
    ApprovedAt   time.Time `json:"approvedAt"`
}

// Current: One approval per stage
order.ApprovalStage++  // Move to next stage after single approval
```

#### Planned Architecture:
```typescript
interface WorkflowStage {
  stageNumber: number;
  stageName: string;
  approvalRequirements: {
    // Multiple approval strategies
    strategy: "single" | "majority" | "unanimous" | "quorum";
    
    // For single strategy
    requiredApprovers?: 1;
    
    // For majority strategy  
    minimumApprovals?: number;
    totalApprovers?: number;
    
    // For quorum strategy
    quorumPercentage?: number; // e.g., 60% must approve
    
    // For unanimous strategy
    allMustApprove?: boolean;
  };
  
  // Approver assignment
  approvers: {
    type: "role" | "specific-users" | "dynamic";
    roles?: UserRole[];
    userIds?: string[];
    dynamicRule?: string; // e.g., "department-manager"
  };
  
  // Stage behavior
  allowParallelApprovals: boolean;
  timeoutHours?: number;
  escalationRule?: EscalationRule;
}

// Example configurations:
const stageExamples = {
  singleApprover: {
    strategy: "single",
    requiredApprovers: 1,
    approvers: { type: "role", roles: ["DEPARTMENT_MANAGER"] }
  },
  
  majorityApproval: {
    strategy: "majority", 
    minimumApprovals: 2,
    totalApprovers: 3,
    approvers: { type: "role", roles: ["FINANCE_MANAGER", "PROCUREMENT_MANAGER", "DEPARTMENT_MANAGER"] }
  },
  
  unanimousApproval: {
    strategy: "unanimous",
    allMustApprove: true,
    approvers: { type: "specific-users", userIds: ["cfo-id", "ceo-id"] }
  }
};
```

#### **Recommendation**:
- **Flexible approval requirements per stage**
- **Support multiple approval strategies**
- **Configurable per workflow and stage**

### 4. **What/Who defines the next stage of the workflow if a required user performs an action?**

#### Current Implementation:
```go
// Linear progression - hardcoded
if order.ApprovalStage >= requiredStages {
    order.Status = "approved"  // Final approval
} else {
    order.Status = "pending"   // Move to next stage
    order.ApprovalStage++      // Increment stage number
}
```

#### Planned Architecture:
```typescript
interface StageTransitionLogic {
  // Who defines transitions
  definedBy: "workflow-designer" | "system-rules" | "dynamic-conditions";
  
  // Transition rules
  transitions: {
    onApprove: {
      nextStage?: number;
      conditions?: TransitionCondition[];
      actions?: WorkflowAction[];
    };
    onReject: {
      targetStage: number; // Which stage to return to
      allowResubmission: boolean;
      requiredChanges?: string[];
    };
    onTimeout: {
      escalationStage?: number;
      notificationRules: NotificationRule[];
    };
  };
}

// Example transition configurations:
const transitionExamples = {
  // Simple linear progression
  linearTransition: {
    onApprove: { nextStage: "current + 1" },
    onReject: { targetStage: 1, allowResubmission: true }
  },
  
  // Conditional branching
  conditionalTransition: {
    onApprove: {
      conditions: [
        { if: "amount > 50000", then: { nextStage: 3 } }, // Skip to finance
        { if: "amount <= 50000", then: { nextStage: 2 } }  // Normal flow
      ]
    }
  },
  
  // Parallel approval paths
  parallelTransition: {
    onApprove: {
      nextStage: "parallel",
      parallelStages: [3, 4], // Both must complete
      convergenceStage: 5
    }
  }
};

// Dynamic stage resolution
interface StageResolver {
  resolveNextStage(
    currentStage: number,
    action: "approve" | "reject",
    document: any,
    workflow: CustomWorkflow
  ): {
    nextStage: number | null;
    isComplete: boolean;
    requiredActions: string[];
  };
}
```

#### **Recommendation**:
- **Workflow Designer** defines transition rules during workflow creation
- **System** executes transitions based on predefined rules
- **Support conditional branching** based on document properties
- **Enable parallel approval paths** where needed

### 5. **Is the Workflow task assigned to a specific user or does it only assign to a role for actioning?**

#### Current Implementation:
```go
// Role-based assignment (implicit)
approverID := c.Locals("user_id").(string)  // Any authenticated user can approve
// No explicit task assignment mechanism
```

#### Planned Architecture:
```typescript
interface TaskAssignmentStrategy {
  // Assignment types
  assignmentType: "specific-user" | "role-based" | "pool-based" | "dynamic";
  
  // Specific user assignment
  specificUser?: {
    userId: string;
    userName: string;
    fallbackUsers?: string[]; // If primary user unavailable
  };
  
  // Role-based assignment  
  roleBased?: {
    roles: UserRole[];
    assignmentMethod: "round-robin" | "load-balanced" | "manual-claim" | "auto-assign";
    departmentRestriction?: string[];
  };
  
  // Pool-based assignment
  poolBased?: {
    poolId: string;
    poolName: string;
    members: string[];
    claimTimeout: number; // Hours before reassignment
  };
  
  // Dynamic assignment
  dynamic?: {
    rule: string; // e.g., "document.department.manager"
    fallbackRule?: string;
  };
}

// Task assignment examples:
const assignmentExamples = {
  // Direct user assignment
  specificAssignment: {
    assignmentType: "specific-user",
    specificUser: {
      userId: "john-doe-id",
      userName: "John Doe",
      fallbackUsers: ["jane-smith-id"] // If John is unavailable
    }
  },
  
  // Role-based with round-robin
  roleBasedRoundRobin: {
    assignmentType: "role-based", 
    roleBased: {
      roles: ["DEPARTMENT_MANAGER"],
      assignmentMethod: "round-robin",
      departmentRestriction: ["IT", "Finance"]
    }
  },
  
  // Approval pool
  poolBasedAssignment: {
    assignmentType: "pool-based",
    poolBased: {
      poolId: "finance-approvers",
      poolName: "Finance Approval Pool",
      members: ["cfo-id", "finance-mgr-1", "finance-mgr-2"],
      claimTimeout: 24 // 24 hours to claim
    }
  },
  
  // Dynamic assignment based on document properties
  dynamicAssignment: {
    assignmentType: "dynamic",
    dynamic: {
      rule: "document.department.manager", // Assign to dept manager
      fallbackRule: "role:DEPARTMENT_MANAGER" // Fallback to any dept manager
    }
  }
};

// Task management
interface WorkflowTask {
  id: string;
  workflowAssignmentId: string;
  stageNumber: number;
  stageName: string;
  
  // Assignment details
  assignedTo?: string;      // Specific user ID
  assignedRole?: UserRole;  // Role assignment
  assignedPool?: string;    // Pool assignment
  
  // Task lifecycle
  status: "pending" | "claimed" | "in-progress" | "completed" | "expired";
  assignedAt: Date;
  claimedAt?: Date;
  claimedBy?: string;
  completedAt?: Date;
  
  // Task properties
  priority: "low" | "medium" | "high" | "urgent";
  dueDate?: Date;
  estimatedDuration?: number; // minutes
  
  // Reassignment capability
  canReassign: boolean;
  reassignmentHistory: ReassignmentRecord[];
}
```

#### **Current vs. Planned Comparison**:

| Aspect | Current Implementation | Planned Architecture |
|--------|----------------------|---------------------|
| **Assignment** | Implicit role-based | Explicit task assignment |
| **Task Tracking** | None | Full task lifecycle |
| **Reassignment** | Manual/Ad-hoc | Systematic with rules |
| **Load Balancing** | None | Multiple strategies |
| **Notifications** | Basic | Rich task notifications |
| **Deadlines** | None | Configurable SLAs |

#### **Recommendation**:
- **Support all assignment types** for maximum flexibility
- **Default to role-based** with manual claim for simplicity
- **Add specific user assignment** for critical approvals
- **Implement approval pools** for high-volume scenarios
- **Enable dynamic assignment** for complex organizational structures

## 🏗️ **Recommended Implementation Approach**

### Phase 1: Enhanced Current System (Weeks 1-4)
```typescript
// Extend current approval system
interface EnhancedApprovalStage {
  stageNumber: number;
  requiredRole: UserRole;
  requiredApprovals: number; // Default: 1
  assignmentType: "role" | "specific-user";
  specificUsers?: string[];
  timeoutHours?: number;
}

// Simple workflow definition
interface SimpleWorkflow {
  entityType: "requisition" | "purchase-order" | "grn" | "payment-voucher";
  name: string;
  stages: EnhancedApprovalStage[];
  conditions?: {
    minAmount?: number;
    maxAmount?: number;
    departments?: string[];
  };
}
```

### Phase 2: Advanced Workflow Engine (Weeks 5-12)
```typescript
// Full workflow system
interface AdvancedWorkflow {
  // Multiple workflows per entity
  variants: WorkflowVariant[];
  
  // Complex approval requirements
  stages: ComplexWorkflowStage[];
  
  // Advanced assignment strategies
  taskAssignment: TaskAssignmentStrategy;
  
  // Conditional transitions
  transitions: StageTransitionLogic;
}
```

### Phase 3: Enterprise Features (Months 4-6)
```typescript
// Enterprise workflow capabilities
interface EnterpriseWorkflow {
  // Parallel approval paths
  parallelStages: ParallelStageConfig[];
  
  // Dynamic approver resolution
  dynamicApprovers: DynamicApproverRule[];
  
  // Integration with external systems
  externalIntegrations: ExternalSystemConfig[];
  
  // Advanced analytics and reporting
  analytics: WorkflowAnalyticsConfig;
}
```

## 🎯 **Immediate Action Items**

### 1. **Clarify Business Requirements** (This Week)
- [ ] Interview stakeholders about approval processes
- [ ] Document current manual approval workflows
- [ ] Identify approval complexity requirements
- [ ] Define user roles and responsibilities

### 2. **Design Workflow Configuration** (Next Week)
- [ ] Create workflow definition schema
- [ ] Design workflow management UI
- [ ] Plan migration from current system
- [ ] Define default workflows for each entity type

### 3. **Implement Basic Multi-Workflow Support** (Weeks 3-4)
- [ ] Add workflow selection logic
- [ ] Create workflow assignment system
- [ ] Implement task assignment
- [ ] Add workflow configuration API

## 📋 **Decision Matrix**

| Question | Current State | Recommended Approach | Priority |
|----------|---------------|---------------------|----------|
| **Who defines workflows?** | System/Developers | Organization Admins → Workflow Designers | High |
| **Multiple workflows per resource?** | No (1 hardcoded) | Yes (unlimited with conditions) | High |
| **Approvals per stage?** | 1 (hardcoded) | Configurable (1 to N) | Medium |
| **Next stage definition?** | Linear progression | Workflow Designer with conditions | High |
| **Task assignment?** | Implicit role | Explicit with multiple strategies | High |

## 🚀 **Quick Start Implementation**

Based on your questions, I recommend starting with this minimal viable workflow system:

```typescript
// Minimal workflow configuration
interface MinimalWorkflow {
  id: string;
  name: string;
  entityType: "requisition" | "purchase-order" | "grn" | "payment-voucher";
  
  // Simple stage definition
  stages: {
    stageNumber: number;
    stageName: string;
    requiredRole: UserRole;
    requiredApprovals: number; // Start with 1, expand later
  }[];
  
  // Basic conditions
  conditions?: {
    amountRange?: { min?: number; max?: number };
    departments?: string[];
  };
  
  // Simple assignment
  assignmentType: "role" | "specific-user";
  
  isDefault: boolean;
  isActive: boolean;
}
```

This provides immediate value while building toward the full enterprise workflow system.

Would you like me to proceed with implementing this minimal workflow system, or would you prefer to discuss any of these design decisions in more detail?