# Workflow Concurrency Issues - Implementation Plan

## Executive Summary

This document outlines a comprehensive plan to fix the critical concurrency issues in the Liyali Gateway workflow system. The plan is structured in phases to minimize disruption while addressing the most critical issues first.

## Implementation Strategy

### Phase 1: Critical Fixes (Week 1-2) - URGENT

**Goal**: Prevent race conditions and conflicting actions

### Phase 2: Enhanced Features (Week 3-4) - HIGH PRIORITY

**Goal**: Add multiple approval support and task claiming

### Phase 3: System Improvements (Week 5-6) - MEDIUM PRIORITY

**Goal**: Optimize performance and user experience

### Phase 4: Advanced Features (Week 7-8) - FUTURE ENHANCEMENT

**Goal**: Add advanced workflow capabilities

---

## Phase 1: Critical Fixes (URGENT)

### 1.1 Add Optimistic Locking to WorkflowTask

**Problem**: Race conditions when multiple users act on same task
**Solution**: Add version control for optimistic locking

**Database Migration:**

```sql
-- Add version column to workflow_tasks table
ALTER TABLE workflow_tasks ADD COLUMN version INTEGER DEFAULT 1 NOT NULL;
ALTER TABLE workflow_tasks ADD COLUMN updated_by VARCHAR(255);
```

**Model Changes:**

```go
// backend/models/enhanced_auth.go
type WorkflowTask struct {
    ID                   string             `gorm:"primaryKey" json:"id"`
    OrganizationID       string             `gorm:"index;not null" json:"organizationId"`
    // ... existing fields ...

    // NEW: Optimistic locking fields
    Version              int                `gorm:"default:1;not null" json:"version"`
    UpdatedBy            *string            `json:"updatedBy,omitempty"`

    // Enhanced claiming fields
    Status               string             `gorm:"default:'pending'" json:"status"`
    ClaimedBy            *string            `json:"claimedBy,omitempty"`
    ClaimedAt            *time.Time         `json:"claimedAt,omitempty"`
    ClaimExpiry          *time.Time         `json:"claimExpiry,omitempty"` // NEW: Claim timeout
}
```

**Service Changes:**

```go
// backend/services/workflow_execution_service.go

// NEW: ClaimWorkflowTask method
func (s *WorkflowExecutionService) ClaimWorkflowTask(ctx context.Context, taskID, userID string) error {
    tx := s.db.Begin()
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
        }
    }()

    var task models.WorkflowTask

    // Atomic claim operation with optimistic locking
    result := tx.Model(&task).
        Where("id = ? AND status = ? AND (claimed_by IS NULL OR claim_expiry < ?)",
              taskID, "pending", time.Now()).
        Updates(map[string]interface{}{
            "claimed_by":    userID,
            "claimed_at":    time.Now(),
            "claim_expiry":  time.Now().Add(30 * time.Minute), // 30-minute claim
            "version":       gorm.Expr("version + 1"),
        })

    if result.Error != nil {
        tx.Rollback()
        return fmt.Errorf("failed to claim task: %w", result.Error)
    }

    if result.RowsAffected == 0 {
        tx.Rollback()
        return fmt.Errorf("task is not available for claiming (already claimed or completed)")
    }

    return tx.Commit().Error
}

// MODIFIED: ApproveWorkflowTask with optimistic locking
func (s *WorkflowExecutionService) ApproveWorkflowTask(ctx context.Context, taskID, userID, signature, comments string, expectedVersion int) error {
    tx := s.db.Begin()
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
        }
    }()

    var task models.WorkflowTask
    if err := tx.Where("id = ?", taskID).First(&task).Error; err != nil {
        tx.Rollback()
        return fmt.Errorf("task not found: %w", err)
    }

    // Check optimistic locking
    if task.Version != expectedVersion {
        tx.Rollback()
        return fmt.Errorf("task was modified by another user (expected version %d, current version %d). Please refresh and try again", expectedVersion, task.Version)
    }

    // Check task is claimed by this user
    if task.ClaimedBy == nil || *task.ClaimedBy != userID {
        tx.Rollback()
        return fmt.Errorf("task must be claimed by you before approval")
    }

    // Check claim hasn't expired
    if task.ClaimExpiry != nil && time.Now().After(*task.ClaimExpiry) {
        tx.Rollback()
        return fmt.Errorf("task claim has expired, please reclaim the task")
    }

    // Validate user role (existing logic)
    if task.AssignedRole != nil {
        var user models.User
        if err := tx.Where("id = ?", userID).First(&user).Error; err != nil {
            tx.Rollback()
            return fmt.Errorf("user not found: %w", err)
        }

        if user.Role != *task.AssignedRole {
            tx.Rollback()
            return fmt.Errorf("insufficient permissions: user role '%s' does not match required role '%s'", user.Role, *task.AssignedRole)
        }
    }

    // Update task with version increment
    result := tx.Model(&task).
        Where("id = ? AND version = ?", taskID, expectedVersion).
        Updates(map[string]interface{}{
            "status":      "completed",
            "completed_at": time.Now(),
            "updated_by":   userID,
            "version":      expectedVersion + 1,
        })

    if result.Error != nil {
        tx.Rollback()
        return fmt.Errorf("failed to update task: %w", result.Error)
    }

    if result.RowsAffected == 0 {
        tx.Rollback()
        return fmt.Errorf("task was modified by another user, please refresh and try again")
    }

    // Continue with existing workflow progression logic...
    // ... rest of the method remains the same

    return tx.Commit().Error
}
```

### 1.2 Enhanced API Endpoints

**New Endpoints:**

```go
// backend/handlers/approval_handler.go

// ClaimTask claims a workflow task for exclusive access
// POST /api/v1/approvals/tasks/:id/claim
func (h *ApprovalHandler) ClaimTask(c *fiber.Ctx) error {
    taskID := c.Params("id")
    userID := c.Locals("userID").(string)

    workflowExecutionService := c.Locals("workflowExecutionService").(*services.WorkflowExecutionService)

    err := workflowExecutionService.ClaimWorkflowTask(c.Context(), taskID, userID)
    if err != nil {
        return c.Status(fiber.StatusConflict).JSON(types.ErrorResponse{
            Error:   "Claim failed",
            Message: err.Error(),
        })
    }

    return c.Status(fiber.StatusOK).JSON(types.SuccessResponse{
        Message: "Task claimed successfully",
        Data:    map[string]interface{}{"taskId": taskID, "claimedBy": userID},
    })
}

// UnclaimTask releases a claimed task
// POST /api/v1/approvals/tasks/:id/unclaim
func (h *ApprovalHandler) UnclaimTask(c *fiber.Ctx) error {
    taskID := c.Params("id")
    userID := c.Locals("userID").(string)

    workflowExecutionService := c.Locals("workflowExecutionService").(*services.WorkflowExecutionService)

    err := workflowExecutionService.UnclaimWorkflowTask(c.Context(), taskID, userID)
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{
            Error:   "Unclaim failed",
            Message: err.Error(),
        })
    }

    return c.Status(fiber.StatusOK).JSON(types.SuccessResponse{
        Message: "Task unclaimed successfully",
        Data:    map[string]interface{}{"taskId": taskID},
    })
}

// MODIFIED: ApproveTask with version control
func (h *ApprovalHandler) ApproveTask(c *fiber.Ctx) error {
    taskID := c.Params("id")
    userID := c.Locals("userID").(string)

    var req struct {
        Signature       string `json:"signature" validate:"required"`
        Comment         string `json:"comment"`
        ExpectedVersion int    `json:"expectedVersion" validate:"required"`
    }

    if err := c.BodyParser(&req); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{
            Error:   "Invalid request body",
            Message: "Failed to parse approval request",
        })
    }

    if err := h.validate.Struct(req); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{
            Error:   "Validation failed",
            Message: err.Error(),
        })
    }

    workflowExecutionService := c.Locals("workflowExecutionService").(*services.WorkflowExecutionService)

    err := workflowExecutionService.ApproveWorkflowTask(
        c.Context(),
        taskID,
        userID,
        req.Signature,
        req.Comment,
        req.ExpectedVersion,
    )

    if err != nil {
        if strings.Contains(err.Error(), "version") {
            return c.Status(fiber.StatusConflict).JSON(types.ErrorResponse{
                Error:   "Concurrent modification",
                Message: err.Error(),
            })
        }
        return c.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse{
            Error:   "Approval failed",
            Message: err.Error(),
        })
    }

    return c.Status(fiber.StatusOK).JSON(types.SuccessResponse{
        Message: "Task approved successfully",
        Data:    map[string]interface{}{"taskId": taskID},
    })
}
```

### 1.3 Frontend Integration

**Enhanced Hook:**

```typescript
// frontend/src/hooks/use-approval-workflow.ts

export const useClaimTask = (taskId: string, onSuccess?: () => void) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async () => {
      const response = await fetch(`/api/v1/approvals/tasks/${taskId}/claim`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
      });
      if (!response.ok) throw new Error(await response.text());
      return response.json();
    },
    onSuccess: () => {
      toast.success("Task claimed successfully");
      queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.APPROVALS.ALL] });
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.APPROVALS.BY_ID, taskId],
      });
      onSuccess?.();
    },
    onError: (error: Error) => {
      toast.error(error.message || "Failed to claim task");
    },
  });
};

// Enhanced approval with version control
export const useApproveTask = (taskId: string, onSuccess?: () => void) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (
      data: ApproveTaskRequest & { expectedVersion: number }
    ) => {
      const response = await approveApprovalTask(taskId, data);
      if (!response.success) throw new Error(response.message);
      return response;
    },
    onSuccess: (response) => {
      toast.success("Task approved successfully");
      // Invalidate queries...
      onSuccess?.();
    },
    onError: (error: Error) => {
      if (error.message.includes("version")) {
        toast.error(
          "Task was modified by another user. Please refresh and try again."
        );
      } else {
        toast.error(error.message || "Failed to approve task");
      }
    },
  });
};
```

---

## Phase 2: Enhanced Features (HIGH PRIORITY)

### 2.1 Multiple Approval Support

**Enhanced WorkflowStage Model:**

```go
// backend/models/enhanced_auth.go
type WorkflowStage struct {
    StageNumber           int                    `json:"stageNumber"`
    StageName             string                 `json:"stageName"`
    RequiredRole          string                 `json:"requiredRole"`
    IsRequired            bool                   `json:"isRequired"`

    // NEW: Multiple approval support
    RequiredApprovalCount int                    `json:"requiredApprovalCount"` // Default: 1
    ApprovalType          string                 `json:"approvalType"`          // "any", "all", "majority", "quorum"
    QuorumCount           *int                   `json:"quorumCount,omitempty"` // For quorum-based approval

    // NEW: Advanced features
    AllowSelfApproval     bool                   `json:"allowSelfApproval"`     // Can creator approve their own document
    RequireUnanimous      bool                   `json:"requireUnanimous"`      // All qualified users must approve
    TimeoutHours          *int                   `json:"timeoutHours,omitempty"`
    EscalationUserID      *string                `json:"escalationUserId,omitempty"`
}
```

**Stage Approval Tracking:**

```go
// NEW: Track individual approvals per stage
type StageApprovalRecord struct {
    ID               string    `gorm:"primaryKey" json:"id"`
    WorkflowTaskID   string    `gorm:"index;not null" json:"workflowTaskId"`
    StageNumber      int       `gorm:"not null" json:"stageNumber"`
    ApproverID       string    `gorm:"not null" json:"approverId"`
    ApproverName     string    `json:"approverName"`
    ApproverRole     string    `json:"approverRole"`
    Action           string    `json:"action"` // "approved", "rejected"
    Comments         string    `json:"comments"`
    Signature        string    `json:"signature"`
    ApprovedAt       time.Time `json:"approvedAt"`
    IPAddress        string    `json:"ipAddress"`
    UserAgent        string    `json:"userAgent"`
}
```

**Enhanced Approval Logic:**

```go
// backend/services/workflow_execution_service.go

func (s *WorkflowExecutionService) ApproveWorkflowTask(ctx context.Context, taskID, userID, signature, comments string, expectedVersion int) error {
    // ... existing validation logic ...

    // Get workflow stage configuration
    stages, err := assignment.Workflow.GetStages()
    if err != nil {
        return fmt.Errorf("failed to get workflow stages: %w", err)
    }

    currentStage := stages[task.StageNumber-1]

    // Record this approval
    approvalRecord := &models.StageApprovalRecord{
        ID:             uuid.New().String(),
        WorkflowTaskID: taskID,
        StageNumber:    task.StageNumber,
        ApproverID:     userID,
        ApproverName:   user.Name,
        ApproverRole:   user.Role,
        Action:         "approved",
        Comments:       comments,
        Signature:      signature,
        ApprovedAt:     time.Now(),
    }

    if err := tx.Create(approvalRecord).Error; err != nil {
        tx.Rollback()
        return fmt.Errorf("failed to record approval: %w", err)
    }

    // Check if stage completion criteria are met
    stageComplete, err := s.checkStageCompletionCriteria(tx, taskID, currentStage)
    if err != nil {
        tx.Rollback()
        return fmt.Errorf("failed to check stage completion: %w", err)
    }

    if stageComplete {
        // Mark task as completed and progress workflow
        // ... existing workflow progression logic
    } else {
        // Update task status to indicate partial approval
        task.Status = "partially_approved"
        if err := tx.Save(&task).Error; err != nil {
            tx.Rollback()
            return fmt.Errorf("failed to update task status: %w", err)
        }
    }

    return tx.Commit().Error
}

func (s *WorkflowExecutionService) checkStageCompletionCriteria(tx *gorm.DB, taskID string, stage models.WorkflowStage) (bool, error) {
    // Get all approvals for this stage
    var approvals []models.StageApprovalRecord
    if err := tx.Where("workflow_task_id = ? AND stage_number = ? AND action = ?",
                      taskID, stage.StageNumber, "approved").Find(&approvals).Error; err != nil {
        return false, err
    }

    approvalCount := len(approvals)

    switch stage.ApprovalType {
    case "any":
        return approvalCount >= 1, nil
    case "all":
        // Get total number of qualified users
        var totalQualified int64
        if err := tx.Model(&models.User{}).
            Where("role = ? AND active = ?", stage.RequiredRole, true).
            Count(&totalQualified).Error; err != nil {
            return false, err
        }
        return approvalCount >= int(totalQualified), nil
    case "majority":
        // Get total number of qualified users
        var totalQualified int64
        if err := tx.Model(&models.User{}).
            Where("role = ? AND active = ?", stage.RequiredRole, true).
            Count(&totalQualified).Error; err != nil {
            return false, err
        }
        required := int(totalQualified)/2 + 1
        return approvalCount >= required, nil
    case "quorum":
        if stage.QuorumCount == nil {
            return false, fmt.Errorf("quorum count not specified for quorum-based approval")
        }
        return approvalCount >= *stage.QuorumCount, nil
    default:
        // Default: require specified count
        return approvalCount >= stage.RequiredApprovalCount, nil
    }
}
```

### 2.2 Enhanced Task Assignment Strategies

**Round-Robin Assignment:**

```go
// NEW: Task assignment strategies
type TaskAssignmentStrategy string

const (
    AssignmentStrategyRole       TaskAssignmentStrategy = "role"
    AssignmentStrategyRoundRobin TaskAssignmentStrategy = "round_robin"
    AssignmentStrategySpecific   TaskAssignmentStrategy = "specific_user"
    AssignmentStrategyUserGroup  TaskAssignmentStrategy = "user_group"
)

// Enhanced WorkflowStage with assignment strategy
type WorkflowStage struct {
    // ... existing fields ...
    AssignmentStrategy TaskAssignmentStrategy `json:"assignmentStrategy"`
    AssignedUserIDs    []string              `json:"assignedUserIds,omitempty"`
    AssignedGroupID    *string               `json:"assignedGroupId,omitempty"`
}

// Round-robin assignment implementation
func (s *WorkflowExecutionService) assignTaskRoundRobin(ctx context.Context, organizationID, role string) (string, error) {
    // Get all qualified users
    var users []models.User
    if err := s.db.Where("current_organization_id = ? AND role = ? AND active = ?",
                         organizationID, role, true).Find(&users).Error; err != nil {
        return "", err
    }

    if len(users) == 0 {
        return "", fmt.Errorf("no qualified users found for role: %s", role)
    }

    // Get last assigned user for this role
    var lastAssignment models.TaskAssignmentHistory
    err := s.db.Where("organization_id = ? AND role = ?", organizationID, role).
        Order("assigned_at DESC").First(&lastAssignment).Error

    var nextUserIndex int
    if err == gorm.ErrRecordNotFound {
        nextUserIndex = 0
    } else {
        // Find current user index and get next
        for i, user := range users {
            if user.ID == lastAssignment.AssignedUserID {
                nextUserIndex = (i + 1) % len(users)
                break
            }
        }
    }

    selectedUser := users[nextUserIndex]

    // Record assignment for round-robin tracking
    assignmentHistory := &models.TaskAssignmentHistory{
        ID:             uuid.New().String(),
        OrganizationID: organizationID,
        Role:           role,
        AssignedUserID: selectedUser.ID,
        AssignedAt:     time.Now(),
    }
    s.db.Create(assignmentHistory)

    return selectedUser.ID, nil
}
```

---

## Phase 3: System Improvements (MEDIUM PRIORITY)

### 3.1 Enhanced Error Handling and User Experience

**Improved Error Messages:**

```go
// backend/types/workflow_errors.go
type WorkflowError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Details string `json:"details,omitempty"`
    Action  string `json:"suggestedAction,omitempty"`
}

var (
    ErrTaskAlreadyClaimed = WorkflowError{
        Code:    "TASK_ALREADY_CLAIMED",
        Message: "This task has been claimed by another user",
        Action:  "Please refresh the page to see current task status",
    }

    ErrConcurrentModification = WorkflowError{
        Code:    "CONCURRENT_MODIFICATION",
        Message: "Task was modified by another user",
        Action:  "Please refresh and try again with the latest version",
    }

    ErrInsufficientApprovals = WorkflowError{
        Code:    "INSUFFICIENT_APPROVALS",
        Message: "More approvals required to complete this stage",
        Details: "This stage requires %d approvals, currently has %d",
    }
)
```

### 3.2 Real-time Updates

**WebSocket Integration:**

```go
// backend/services/websocket_service.go
type WorkflowWebSocketService struct {
    hub *websocket.Hub
}

func (s *WorkflowWebSocketService) NotifyTaskUpdate(organizationID, taskID string, update TaskUpdateEvent) {
    // Send real-time updates to all connected users in organization
    s.hub.BroadcastToOrganization(organizationID, map[string]interface{}{
        "type":   "task_update",
        "taskId": taskID,
        "update": update,
    })
}

type TaskUpdateEvent struct {
    Type      string    `json:"type"` // "claimed", "approved", "rejected", "unclaimed"
    TaskID    string    `json:"taskId"`
    UserID    string    `json:"userId"`
    UserName  string    `json:"userName"`
    Timestamp time.Time `json:"timestamp"`
}
```

### 3.3 Performance Optimizations

**Database Indexing:**

```sql
-- Add indexes for better query performance
CREATE INDEX idx_workflow_tasks_org_status ON workflow_tasks(organization_id, status);
CREATE INDEX idx_workflow_tasks_claimed_by ON workflow_tasks(claimed_by) WHERE claimed_by IS NOT NULL;
CREATE INDEX idx_workflow_tasks_assigned_role ON workflow_tasks(assigned_role) WHERE assigned_role IS NOT NULL;
CREATE INDEX idx_stage_approval_records_task_stage ON stage_approval_records(workflow_task_id, stage_number);
```

**Caching Strategy:**

```go
// backend/services/workflow_cache_service.go
type WorkflowCacheService struct {
    redis *redis.Client
}

func (s *WorkflowCacheService) CacheWorkflowDefinition(workflowID string, workflow *models.Workflow) error {
    data, err := json.Marshal(workflow)
    if err != nil {
        return err
    }

    return s.redis.Set(context.Background(),
        fmt.Sprintf("workflow:%s", workflowID),
        data,
        30*time.Minute).Err()
}
```

---

## Phase 4: Advanced Features (FUTURE ENHANCEMENT)

### 4.1 Workflow Analytics and Reporting

**Approval Metrics:**

```go
type WorkflowMetrics struct {
    AverageApprovalTime   time.Duration `json:"averageApprovalTime"`
    TasksClaimedButNotCompleted int     `json:"tasksClaimedButNotCompleted"`
    ConcurrentAttempts    int           `json:"concurrentAttempts"`
    StageBottlenecks      []string      `json:"stageBottlenecks"`
}
```

### 4.2 Advanced Workflow Rules

**Conditional Workflows:**

```go
type WorkflowCondition struct {
    Field    string      `json:"field"`
    Operator string      `json:"operator"` // "equals", "greater_than", "contains"
    Value    interface{} `json:"value"`
}

type ConditionalStage struct {
    Condition WorkflowCondition `json:"condition"`
    Stage     WorkflowStage     `json:"stage"`
}
```

---

## Implementation Timeline

### Week 1-2: Critical Fixes

- [ ] Add optimistic locking to WorkflowTask model
- [ ] Implement task claiming mechanism
- [ ] Add version control to approval/rejection APIs
- [ ] Update frontend hooks for version control
- [ ] Add comprehensive error handling

### Week 3-4: Enhanced Features

- [ ] Implement multiple approval support
- [ ] Add stage approval tracking
- [ ] Create round-robin assignment strategy
- [ ] Build enhanced approval logic
- [ ] Add user group assignment support

### Week 5-6: System Improvements

- [ ] Add real-time WebSocket updates
- [ ] Implement performance optimizations
- [ ] Add comprehensive logging and monitoring
- [ ] Create workflow analytics dashboard
- [ ] Add automated claim expiry cleanup

### Week 7-8: Advanced Features

- [ ] Implement conditional workflows
- [ ] Add workflow templates
- [ ] Create approval delegation system
- [ ] Add workflow simulation and testing tools
- [ ] Implement advanced reporting features

---

## Testing Strategy

### Unit Tests

- [ ] Test optimistic locking scenarios
- [ ] Test task claiming race conditions
- [ ] Test multiple approval logic
- [ ] Test assignment strategies

### Integration Tests

- [ ] Test concurrent user scenarios
- [ ] Test workflow progression with multiple approvals
- [ ] Test error handling and recovery
- [ ] Test WebSocket real-time updates

### Load Tests

- [ ] Test system under concurrent load
- [ ] Test database performance with new indexes
- [ ] Test cache performance
- [ ] Test WebSocket scalability

---

## Risk Mitigation

### Database Migration Risks

- **Risk**: Downtime during migration
- **Mitigation**: Use online schema changes, deploy during low-traffic periods

### Backward Compatibility

- **Risk**: Breaking existing workflows
- **Mitigation**: Maintain API versioning, gradual rollout with feature flags

### Performance Impact

- **Risk**: New features may slow down system
- **Mitigation**: Comprehensive performance testing, database optimization

### User Adoption

- **Risk**: Users may resist new claiming workflow
- **Mitigation**: Gradual rollout, comprehensive training, clear benefits communication

---

## Success Metrics

### Technical Metrics

- Zero race condition incidents
- < 100ms average task claim response time
- 99.9% workflow completion success rate
- < 1% concurrent modification errors

### Business Metrics

- Reduced approval processing time
- Improved user satisfaction scores
- Decreased support tickets related to workflow issues
- Increased workflow adoption across organization

---

## Conclusion

This implementation plan addresses the critical concurrency issues in the workflow system while adding enhanced features for better user experience and business requirements. The phased approach ensures minimal disruption while delivering immediate value through critical fixes.

The plan prioritizes:

1. **Safety**: Preventing race conditions and data corruption
2. **Usability**: Clear task ownership and better error messages
3. **Flexibility**: Support for complex approval requirements
4. **Performance**: Optimized queries and caching
5. **Scalability**: Real-time updates and advanced features

Regular progress reviews and stakeholder feedback will ensure successful implementation and adoption of these improvements.
