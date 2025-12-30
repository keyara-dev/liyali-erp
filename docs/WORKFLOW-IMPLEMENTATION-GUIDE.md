# Workflow Backend Integration - Implementation Guide

**Date**: 2025-12-26
**Status**: 🚀 READY FOR IMPLEMENTATION
**Purpose**: Step-by-step developer guide for implementing workflow backend integration

---

## Table of Contents

1. [Quick Start](#quick-start)
2. [Backend Implementation](#backend-implementation)
3. [Frontend Implementation](#frontend-implementation)
4. [Testing](#testing)
5. [Troubleshooting](#troubleshooting)
6. [Code Templates](#code-templates)

---

## Quick Start

### For Backend Developers

1. **Create approval task handlers** in `backend/handlers/approval.go`
2. **Add routes** to `backend/routes/routes.go`
3. **Test with Postman** before frontend integration
4. **Commit and notify frontend team**

### For Frontend Developers

1. **Wait for backend APIs** to be created
2. **Create server actions** from template in `frontend/src/app/_actions/approval-workflow.ts`
3. **Build React hooks** from template in `frontend/src/hooks/use-approval-workflow.ts`
4. **Update components** to use hooks instead of localStorage
5. **Test end-to-end**

### Parallel Work

- Backend: Create APIs (Days 1-2)
- Frontend: Prepare components for API integration (Days 1-2)
- Integration: Connect when APIs ready (Days 3-4)

---

## Backend Implementation

### Step 1: Create Approval Handler File

**File**: `backend/handlers/approval.go`

```go
package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/types"
	"github.com/liyali/liyali-gateway/utils"
	"gorm.io/gorm"
)

// GetApprovalTasks retrieves approval tasks with pagination and filtering
func GetApprovalTasks(c fiber.Ctx) error {
	db := config.DB
	organizationID := c.Locals("organization_id").(string)
	userID := c.Locals("user_id").(string)

	// Extract query parameters
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	status := c.Query("status")
	documentType := c.Query("document_type")
	assignedToMe := c.QueryBool("assigned_to_me", false)

	// Build query
	query := db.Where("organization_id = ?", organizationID)

	if assignedToMe {
		query = query.Where("approver_id = ?", userID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if documentType != "" {
		query = query.Where("document_type = ?", documentType)
	}

	// Get total count
	var total int64
	if err := query.Model(&models.ApprovalTask{}).Count(&total).Error; err != nil {
		return utils.SendInternalError(c, "Failed to count tasks", err)
	}

	// Fetch paginated results
	var tasks []models.ApprovalTask
	offset := (page - 1) * limit
	if err := query.
		Offset(offset).
		Limit(limit).
		Preload("Approver").
		Order("created_at DESC").
		Find(&tasks).Error; err != nil {
		return utils.SendInternalError(c, "Failed to fetch tasks", err)
	}

	// Convert to response format
	responses := make([]types.ApprovalTaskResponse, 0, len(tasks))
	for _, task := range tasks {
		responses = append(responses, modelToApprovalTaskResponse(task))
	}

	pagination := utils.CalculatePagination(page, limit, total)
	return utils.SendSuccess(c, fiber.StatusOK, responses, "Approval tasks retrieved", pagination)
}

// GetApprovalTask retrieves a single approval task with full details
func GetApprovalTask(c fiber.Ctx) error {
	taskID := c.Params("id")
	if taskID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Task ID is required",
		})
	}

	organizationID := c.Locals("organization_id").(string)

	var task models.ApprovalTask
	if err := config.DB.
		Preload("Approver").
		Where("id = ? AND organization_id = ?", taskID, organizationID).
		First(&task).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Task not found",
			})
		}
		return utils.SendInternalError(c, "Failed to fetch task", err)
	}

	// Fetch related document based on document_type
	var documentDetail interface{}
	switch task.DocumentType {
	case "requisition":
		var req models.Requisition
		if err := config.DB.
			Preload("Requester").
			Where("id = ?", task.DocumentID).
			First(&req).Error; err == nil {
			documentDetail = modelToRequisitionResponse(req)
		}
	case "purchase_order":
		var po models.PurchaseOrder
		if err := config.DB.
			Where("id = ?", task.DocumentID).
			First(&po).Error; err == nil {
			documentDetail = modelToPurchaseOrderResponse(po)
		}
	// Add other document types as needed
	}

	return utils.SendSuccess(c, fiber.StatusOK, types.ApprovalTaskDetailResponse{
		Task:     modelToApprovalTaskResponse(task),
		Document: documentDetail,
	}, "Approval task retrieved")
}

// ApproveTask marks a task as approved and moves to next stage
func ApproveTask(c fiber.Ctx) error {
	taskID := c.Params("id")
	userID := c.Locals("user_id").(string)
	organizationID := c.Locals("organization_id").(string)

	var req types.ApproveTaskRequest
	if err := c.BindJSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	// Validate signature
	if req.Signature == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Digital signature is required",
		})
	}

	// Fetch approval task
	var task models.ApprovalTask
	if err := config.DB.
		Where("id = ? AND organization_id = ?", taskID, organizationID).
		First(&task).Error; err != nil {
		return utils.SendInternalError(c, "Task not found", err)
	}

	// Verify user is the assigned approver
	if task.ApproverID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "You are not the assigned approver for this task",
		})
	}

	// Update task status
	task.Status = "approved"
	task.Comments = req.Comments
	task.Signature = req.Signature
	task.UpdatedAt = time.Now()

	if err := config.DB.Save(&task).Error; err != nil {
		return utils.SendInternalError(c, "Failed to approve task", err)
	}

	// Update document approval history
	if err := updateDocumentApprovalHistory(task.DocumentID, task.DocumentType, task, "approved"); err != nil {
		return utils.SendInternalError(c, "Failed to update document", err)
	}

	// Create audit log
	createAuditLog(organizationID, task.DocumentID, task.DocumentType, userID, "approve", fiber.Map{
		"stage":    task.Stage,
		"comments": req.Comments,
	})

	// Send notification to next approver if not final stage
	// (Implementation depends on approval config)

	return utils.SendSuccess(c, fiber.StatusOK, task, "Task approved successfully")
}

// RejectTask marks a task as rejected and returns document to requester
func RejectTask(c fiber.Ctx) error {
	taskID := c.Params("id")
	userID := c.Locals("user_id").(string)
	organizationID := c.Locals("organization_id").(string)

	var req types.RejectTaskRequest
	if err := c.BindJSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
		})
	}

	if req.Signature == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Digital signature is required",
		})
	}

	// Fetch approval task
	var task models.ApprovalTask
	if err := config.DB.
		Where("id = ? AND organization_id = ?", taskID, organizationID).
		First(&task).Error; err != nil {
		return utils.SendInternalError(c, "Task not found", err)
	}

	if task.ApproverID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "You are not the assigned approver for this task",
		})
	}

	// Update task
	task.Status = "rejected"
	task.Comments = req.Comments
	task.Signature = req.Signature
	task.UpdatedAt = time.Now()

	if err := config.DB.Save(&task).Error; err != nil {
		return utils.SendInternalError(c, "Failed to reject task", err)
	}

	// Update document - set back to DRAFT
	if err := updateDocumentStatus(task.DocumentID, task.DocumentType, "draft"); err != nil {
		return utils.SendInternalError(c, "Failed to update document", err)
	}

	if err := updateDocumentApprovalHistory(task.DocumentID, task.DocumentType, task, "rejected"); err != nil {
		return utils.SendInternalError(c, "Failed to update document", err)
	}

	// Create audit log
	createAuditLog(organizationID, task.DocumentID, task.DocumentType, userID, "reject", fiber.Map{
		"stage":   task.Stage,
		"remarks": req.Remarks,
	})

	return utils.SendSuccess(c, fiber.StatusOK, task, "Task rejected successfully")
}

// ReassignTask reassigns task to different approver
func ReassignTask(c fiber.Ctx) error {
	taskID := c.Params("id")
	organizationID := c.Locals("organization_id").(string)

	var req types.ReassignTaskRequest
	if err := c.BindJSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
		})
	}

	// Fetch task
	var task models.ApprovalTask
	if err := config.DB.
		Where("id = ? AND organization_id = ?", taskID, organizationID).
		First(&task).Error; err != nil {
		return utils.SendInternalError(c, "Task not found", err)
	}

	// Store old approver for audit
	oldApproverID := task.ApproverID

	// Update task
	task.ApproverID = req.NewApproverId
	task.UpdatedAt = time.Now()

	if err := config.DB.Save(&task).Error; err != nil {
		return utils.SendInternalError(c, "Failed to reassign task", err)
	}

	// Create audit log
	createAuditLog(organizationID, task.DocumentID, task.DocumentType,
		c.Locals("user_id").(string), "reassign", fiber.Map{
		"from":   oldApproverID,
		"to":     req.NewApproverId,
		"reason": req.Reason,
	})

	return utils.SendSuccess(c, fiber.StatusOK, task, "Task reassigned successfully")
}

// GetApprovalHistory retrieves approval history for a document
func GetApprovalHistory(c fiber.Ctx) error {
	documentID := c.Params("documentId")
	organizationID := c.Locals("organization_id").(string)

	if documentID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Document ID is required",
		})
	}

	// Fetch document to get approval history
	var document models.Requisition
	if err := config.DB.
		Where("id = ? AND organization_id = ?", documentID, organizationID).
		First(&document).Error; err != nil {
		return utils.SendInternalError(c, "Document not found", err)
	}

	// Parse approval history JSON
	var history []types.ApprovalRecord
	if err := json.Unmarshal(document.ApprovalHistory, &history); err != nil {
		history = []types.ApprovalRecord{}
	}

	return utils.SendSuccess(c, fiber.StatusOK, history, "Approval history retrieved")
}

// Helper functions

func updateDocumentApprovalHistory(docID string, docType string, task models.ApprovalTask, status string) error {
	// Implementation depends on document type
	// Get document, unmarshal ApprovalHistory JSON, add new record, marshal back
	return nil
}

func updateDocumentStatus(docID string, docType string, newStatus string) error {
	// Update document status based on document type
	switch docType {
	case "requisition":
		return config.DB.Model(&models.Requisition{}).
			Where("id = ?", docID).
			Update("status", newStatus).Error
	case "purchase_order":
		return config.DB.Model(&models.PurchaseOrder{}).
			Where("id = ?", docID).
			Update("status", newStatus).Error
	// Add other types
	}
	return nil
}

func createAuditLog(orgID, docID, docType, userID, action string, changes interface{}) error {
	changesJSON, _ := json.Marshal(changes)
	log := models.AuditLog{
		ID:           uuid.New().String(),
		DocumentID:   docID,
		DocumentType: docType,
		UserID:       userID,
		Action:       action,
		Changes:      changesJSON,
		CreatedAt:    time.Now(),
	}
	return config.DB.Create(&log).Error
}

func modelToApprovalTaskResponse(task models.ApprovalTask) types.ApprovalTaskResponse {
	return types.ApprovalTaskResponse{
		ID:             task.ID,
		OrganizationID: task.OrganizationID,
		DocumentID:     task.DocumentID,
		DocumentType:   task.DocumentType,
		ApproverID:     task.ApproverID,
		Status:         task.Status,
		Stage:          task.Stage,
		Comments:       task.Comments,
		CreatedAt:      task.CreatedAt,
		UpdatedAt:      task.UpdatedAt,
	}
}
```

### Step 2: Add Response Types

**File**: `backend/types/approval.go`

```go
package types

import "time"

type ApprovalTaskResponse struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organizationId"`
	DocumentID     string    `json:"documentId"`
	DocumentType   string    `json:"documentType"`
	ApproverID     string    `json:"approverId"`
	Status         string    `json:"status"`
	Stage          int       `json:"stage"`
	Comments       string    `json:"comments"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type ApprovalTaskDetailResponse struct {
	Task     ApprovalTaskResponse `json:"task"`
	Document interface{}          `json:"document"`
}

type ApproveTaskRequest struct {
	Comments   string `json:"comments"`
	Signature  string `json:"signature"`
	StageNumber int    `json:"stageNumber"`
}

type RejectTaskRequest struct {
	Remarks   string `json:"remarks"`
	Comments  string `json:"comments"`
	Signature string `json:"signature"`
	ReturnTo  string `json:"returnTo,omitempty"`
}

type ReassignTaskRequest struct {
	NewApproverId string `json:"newApproverId"`
	Reason        string `json:"reason"`
}

type ApprovalRecord struct {
	ApproverID   string    `json:"approverId"`
	ApproverName string    `json:"approverName"`
	Status       string    `json:"status"`
	Comments     string    `json:"comments"`
	Signature    string    `json:"signature"`
	ApprovedAt   time.Time `json:"approvedAt"`
}
```

### Step 3: Add Routes

**File**: `backend/routes/routes.go` (Add to approval routes section)

```go
// Approval task routes
approvalGroup := app.Group("/api/v1/approvals")
approvalGroup.Use(middleware.AuthMiddleware())
approvalGroup.Use(middleware.TenantMiddleware())

approvalGroup.Get("", handlers.GetApprovalTasks)                    // GET all tasks
approvalGroup.Get("/:id", handlers.GetApprovalTask)                  // GET single task
approvalGroup.Post("/:id/approve", handlers.ApproveTask)            // Approve task
approvalGroup.Post("/:id/reject", handlers.RejectTask)              // Reject task
approvalGroup.Post("/:id/reassign", handlers.ReassignTask)          // Reassign task

// Approval history route
documentsGroup := app.Group("/api/v1/documents")
documentsGroup.Use(middleware.AuthMiddleware())
documentsGroup.Use(middleware.TenantMiddleware())

documentsGroup.Get("/:documentId/approval-history", handlers.GetApprovalHistory)
```

### Step 4: Test Endpoints

Use Postman or curl to test:

```bash
# Get approval tasks
curl -X GET "http://localhost:3000/api/v1/approvals?status=PENDING" \
  -H "Authorization: Bearer YOUR_TOKEN"

# Get single task
curl -X GET "http://localhost:3000/api/v1/approvals/TASK_ID" \
  -H "Authorization: Bearer YOUR_TOKEN"

# Approve task
curl -X POST "http://localhost:3000/api/v1/approvals/TASK_ID/approve" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "comments": "Approved",
    "signature": "data:image/png;base64,...",
    "stageNumber": 1
  }'
```

---

## Frontend Implementation

### Step 1: Create Query Keys Constant

**File**: `frontend/src/lib/constants.ts` (Update)

```typescript
export const QUERY_KEYS = {
  // ... existing keys ...
  APPROVALS: {
    ALL: ['approvals'],
    BY_ID: ['approvals_by_id'],
    PENDING_COUNT: ['approvals_pending_count'],
    HISTORY: ['approval_history'],
  },
  DOCUMENTS: {
    BY_ID: ['documents_by_id'],
  },
};
```

### Step 2: Create Types

**File**: `frontend/src/types/workflow.ts` (Add)

```typescript
// Already documented in plan, add these types:

export interface ApprovalTask {
  id: string;
  organizationId: string;
  documentId: string;
  documentType: 'requisition' | 'purchase_order' | 'payment_voucher' | 'grn' | 'budget';
  documentNumber: string;
  approverId: string;
  approverName: string;
  approverRole: string;
  status: 'PENDING' | 'APPROVED' | 'REJECTED';
  stage: number;
  priority: string;
  createdAt: Date;
  dueAt: Date;
  overdue: boolean;
}

export interface ApproveTaskRequest {
  taskId: string;
  comments: string;
  signature: string;
  stageNumber: number;
}

export interface RejectTaskRequest {
  taskId: string;
  remarks: string;
  comments: string;
  signature: string;
  returnTo?: string;
}

export interface ReassignTaskRequest {
  taskId: string;
  newApproverId: string;
  reason: string;
}
```

### Step 3: Create Server Actions

From template in WORKFLOW-BACKEND-INTEGRATION-PLAN.md

### Step 4: Create Hooks

From template in WORKFLOW-BACKEND-INTEGRATION-PLAN.md

### Step 5: Update Components

Remove `approval-store.ts` usage and replace with hooks:

**Before**:
```typescript
const store = new ApprovalStore();
const task = store.getTaskDetail(taskId);
```

**After**:
```typescript
const { data: task } = useApprovalTaskDetail(taskId);
```

---

## Testing

### Backend Testing Checklist

- [ ] Create ApprovalTask in database
- [ ] GET /api/v1/approvals returns tasks
- [ ] GET /api/v1/approvals?assigned_to_me=true returns only user's tasks
- [ ] GET /api/v1/approvals/:id returns full details
- [ ] POST /api/v1/approvals/:id/approve updates task status
- [ ] POST /api/v1/approvals/:id/reject updates task and document
- [ ] POST /api/v1/approvals/:id/reassign changes approver
- [ ] Non-approver cannot approve
- [ ] AuditLog created on approval
- [ ] Notification sent to next approver

### Frontend Testing Checklist

- [ ] useApprovalTasks hook fetches and displays tasks
- [ ] useApprovalTaskDetail hook loads specific task
- [ ] useApproveTask calls API and updates cache
- [ ] useRejectTask calls API and updates cache
- [ ] Toast notifications show on success
- [ ] Error messages display on failure
- [ ] Removed all mock data dependencies
- [ ] localStorage not used for approval data
- [ ] RBAC permissions checked on frontend

### E2E Testing

1. Create requisition (Draft)
2. Submit for approval (IN_REVIEW)
3. Approver 1 approves (moves to stage 2)
4. Approver 2 approves (moves to stage 3)
5. Approver 3 approves (APPROVED)
6. Verify PO created automatically
7. Check audit log has all approvals
8. Verify notifications sent

---

## Troubleshooting

### Issue: "You are not the assigned approver"

**Cause**: Task's approver_id doesn't match current user_id

**Solution**: Check that user is actually assigned. View ApprovalTask in database.

### Issue: Signature not saved

**Cause**: Signature string too long or format invalid

**Solution**: Validate signature is base64 encoded image data before sending

### Issue: Document status not updating

**Cause**: updateDocumentStatus function not implemented for document type

**Solution**: Add implementation for document type in switch statement

### Issue: Hooks not fetching data

**Cause**: Query key doesn't match backend response structure

**Solution**: Check API response structure matches type definitions

---

## Code Templates

### Template: Complete Approval Flow Component

```typescript
'use client';

import { useState } from 'react';
import { useApprovalTaskDetail, useApproveTask, useRejectTask } from '@/hooks/use-approval-workflow';
import { SignatureCanvas } from './signature-canvas';
import { toast } from 'sonner';

export function ApprovalFlowComponent({ taskId }: { taskId: string }) {
  const [signature, setSignature] = useState<string>('');
  const [comments, setComments] = useState('');

  const { data: task, isLoading } = useApprovalTaskDetail(taskId);
  const approveMutation = useApproveTask(taskId);
  const rejectMutation = useRejectTask(taskId);

  const handleApprove = async () => {
    if (!signature) {
      toast.error('Signature is required');
      return;
    }

    try {
      await approveMutation.mutateAsync({
        comments,
        signature,
        stageNumber: task!.stage,
      });
      setSignature('');
      setComments('');
    } catch (error) {
      // Error handled by mutation
    }
  };

  const handleReject = async () => {
    if (!signature) {
      toast.error('Signature is required');
      return;
    }

    try {
      await rejectMutation.mutateAsync({
        remarks: 'Rejected by approver',
        comments,
        signature,
        returnTo: 'REQUESTER',
      });
      setSignature('');
      setComments('');
    } catch (error) {
      // Error handled by mutation
    }
  };

  if (isLoading) return <div>Loading task...</div>;
  if (!task) return <div>Task not found</div>;

  return (
    <div className="approval-panel">
      <h2>{task.document?.title}</h2>
      <p>Stage {task.stage} of {task.totalStages}</p>

      <textarea
        value={comments}
        onChange={(e) => setComments(e.target.value)}
        placeholder="Comments (optional)"
      />

      <SignatureCanvas
        onChange={setSignature}
        disabled={approveMutation.isPending || rejectMutation.isPending}
      />

      <button
        onClick={handleApprove}
        disabled={approveMutation.isPending || !signature}
      >
        Approve
      </button>

      <button
        onClick={handleReject}
        disabled={rejectMutation.isPending || !signature}
      >
        Reject
      </button>
    </div>
  );
}
```

---

## References

- [Workflow Backend Integration Plan](WORKFLOW-BACKEND-INTEGRATION-PLAN.md)
- [Approval Config](frontend/src/lib/approval-config.ts)
- [Approval Types](frontend/src/types/workflow.ts)
- [Approval Models](backend/models/models.go)

---

**Last Updated**: 2025-12-26
**Status**: 🚀 READY FOR DEVELOPMENT
**Questions?**: Refer to WORKFLOW-BACKEND-INTEGRATION-PLAN.md

