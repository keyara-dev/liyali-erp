package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/handlers"
	"github.com/liyali/liyali-gateway/middleware"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/repository"
	"github.com/liyali/liyali-gateway/services"
	"github.com/liyali/liyali-gateway/types"
	"github.com/stretchr/testify/assert"
	"gorm.io/datatypes"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestCompleteWorkflowIntegration tests the complete workflow integration
// including document status updates and automation triggers
func TestCompleteWorkflowIntegration(t *testing.T) {
	// Setup test database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// Auto-migrate all models
	err = db.AutoMigrate(
		&models.User{},
		&models.Organization{},
		&models.Vendor{},
		&models.Category{},
		&models.Requisition{},
		&models.PurchaseOrder{},
		&models.PaymentVoucher{},
		&models.GoodsReceivedNote{},
		&models.Budget{},
		&models.Workflow{},
		&models.WorkflowAssignment{},
		&models.WorkflowTask{},
		&models.ApprovalTask{},
	)
	assert.NoError(t, err)

	// Set global DB for handlers
	config.DB = db

	// Create test data
	orgID := uuid.New().String()
	userID := uuid.New().String()
	vendorID := uuid.New().String()

	// Create organization
	org := models.Organization{
		ID:   orgID,
		Name: "Test Organization",
	}
	db.Create(&org)

	// Create user
	user := models.User{
		ID:                     userID,
		Name:                   "Test User",
		Email:                  "test@example.com",
		Role:                   "manager",
		CurrentOrganizationID:  &orgID,
		Active:                 true,
	}
	db.Create(&user)

	// Create vendor
	vendor := models.Vendor{
		ID:             vendorID,
		Name:           "Test Vendor",
		Email:          "vendor@example.com",
		Active:         true,
	}
	db.Create(&vendor)

	// Create default workflow for requisitions
	workflowID := uuid.New().String()
	workflow := models.Workflow{
		ID:             workflowID,
		OrganizationID: orgID,
		Name:           "Default Requisition Workflow",
		DocumentType:   "REQUISITION",
		IsDefault:      true,
		IsActive:       true,
		Version:        1,
		Stages: datatypes.JSON(`[
			{
				"stageNumber": 1,
				"stageName": "Manager Approval",
				"requiredRole": "manager",
				"timeoutHours": 24
			},
			{
				"stageNumber": 2,
				"stageName": "Finance Approval",
				"requiredRole": "finance",
				"timeoutHours": 48
			}
		]`),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	db.Create(&workflow)

	// Initialize services
	auditService := services.NewAuditService()
	notificationService := services.NewNotificationService(db)
	automationService := services.NewDocumentAutomationService(db, auditService, notificationService)
	
	// For workflow service, we need to create a simple repository implementation
	workflowRepo := &SimpleWorkflowRepo{db: db}
	workflowService := services.NewWorkflowService(workflowRepo, auditService, db)
	workflowExecutionService := services.NewWorkflowExecutionService(db, workflowService, auditService, automationService)

	// Initialize handlers
	handlerRegistry := handlers.NewHandlerRegistry(
		nil, // authService not needed for this test
		nil, // rbacService not needed for this test
		workflowService,
		workflowExecutionService,
		nil, // documentService not needed for this test
		automationService,
	)

	// Setup Fiber app
	app := fiber.New()

	// Add middleware to inject services
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("organizationID", orgID)
		c.Locals("userID", userID)
		c.Locals("workflowExecutionService", workflowExecutionService)
		return c.Next()
	})

	// Setup routes
	app.Post("/api/v1/requisitions", handlers.CreateRequisition)
	app.Post("/api/v1/requisitions/:id/submit", handlers.SubmitRequisition)
	app.Get("/api/v1/requisitions/:id", handlers.GetRequisition)
	app.Get("/api/v1/approvals", handlerRegistry.Approval.GetApprovalTasks)
	app.Post("/api/v1/approvals/:id/approve", handlerRegistry.Approval.ApproveTask)
	app.Get("/api/v1/documents/:documentId/approval-status", handlerRegistry.Approval.GetApprovalWorkflowStatus)

	t.Run("Complete Workflow Integration Test", func(t *testing.T) {
		// Step 1: Create a requisition
		reqData := types.CreateRequisitionRequest{
			Title:             "Test Requisition for Workflow",
			Description:       "Testing complete workflow integration",
			Department:        "IT",
			Priority:          "medium",
			Items: []types.RequisitionItem{
				{
					Description: "Test Item",
					Quantity:    1,
					Unit:        &[]string{"pcs"}[0],
					UnitPrice:   1000.00,
					Amount:      1000.00,
				},
			},
			TotalAmount:       1000.00,
			Currency:          "USD",
			RequiredByDate:    time.Now().AddDate(0, 0, 30),
			PreferredVendorID: &vendorID,
		}

		reqBody, _ := json.Marshal(reqData)
		req := httptest.NewRequest("POST", "/api/v1/requisitions", bytes.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var createResp types.DetailResponse
		json.NewDecoder(resp.Body).Decode(&createResp)
		assert.True(t, createResp.Success)

		requisitionData := createResp.Data.(map[string]interface{})
		requisitionID := requisitionData["id"].(string)

		// Verify requisition is in draft status
		assert.Equal(t, "draft", requisitionData["status"])

		// Step 2: Submit requisition for approval (should create workflow)
		submitData := map[string]interface{}{
			"comments": "Submitting for workflow approval",
		}

		submitBody, _ := json.Marshal(submitData)
		submitReq := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/requisitions/%s/submit", requisitionID), bytes.NewReader(submitBody))
		submitReq.Header.Set("Content-Type", "application/json")

		submitResp, err := app.Test(submitReq)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, submitResp.StatusCode)

		var submitResponse types.DetailResponse
		json.NewDecoder(submitResp.Body).Decode(&submitResponse)
		assert.True(t, submitResponse.Success)

		// Verify requisition status changed to pending
		var updatedRequisition models.Requisition
		db.Where("id = ?", requisitionID).First(&updatedRequisition)
		assert.Equal(t, "pending", updatedRequisition.Status)

		// Verify workflow assignment was created
		var workflowAssignment models.WorkflowAssignment
		err = db.Where("entity_id = ? AND entity_type = ?", requisitionID, "REQUISITION").First(&workflowAssignment).Error
		assert.NoError(t, err)
		assert.Equal(t, "in_progress", workflowAssignment.Status)
		assert.Equal(t, 1, workflowAssignment.CurrentStage)

		// Verify workflow task was created
		var workflowTask models.WorkflowTask
		err = db.Where("workflow_assignment_id = ? AND status = ?", workflowAssignment.ID, "pending").First(&workflowTask).Error
		assert.NoError(t, err)
		assert.Equal(t, 1, workflowTask.StageNumber)
		assert.Equal(t, "Manager Approval", workflowTask.StageName)
		assert.Equal(t, "manager", *workflowTask.AssignedRole)

		// Step 3: Get approval tasks (should show the pending task)
		tasksReq := httptest.NewRequest("GET", "/api/v1/approvals?assigned_to_me=true", nil)
		tasksResp, err := app.Test(tasksReq)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, tasksResp.StatusCode)

		var tasksResponse types.DetailResponse
		json.NewDecoder(tasksResp.Body).Decode(&tasksResponse)
		assert.True(t, tasksResponse.Success)

		tasks := tasksResponse.Data.([]interface{})
		assert.Len(t, tasks, 1)

		taskData := tasks[0].(map[string]interface{})
		taskID := taskData["id"].(string)

		// Step 4: Approve the first stage (Manager Approval)
		approveData := types.ApproveTaskRequest{
			Signature: "manager_signature_123",
			Comments:  "Approved by manager",
		}

		approveBody, _ := json.Marshal(approveData)
		approveReq := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/approvals/%s/approve", taskID), bytes.NewReader(approveBody))
		approveReq.Header.Set("Content-Type", "application/json")

		approveResp, err := app.Test(approveReq)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, approveResp.StatusCode)

		// Verify workflow progressed to next stage
		db.Where("id = ?", workflowAssignment.ID).First(&workflowAssignment)
		assert.Equal(t, 2, workflowAssignment.CurrentStage)
		assert.Equal(t, "in_progress", workflowAssignment.Status)

		// Verify new task was created for finance approval
		var financeTask models.WorkflowTask
		err = db.Where("workflow_assignment_id = ? AND status = ? AND stage_number = ?", workflowAssignment.ID, "pending", 2).First(&financeTask).Error
		assert.NoError(t, err)
		assert.Equal(t, "Finance Approval", financeTask.StageName)
		assert.Equal(t, "finance", *financeTask.AssignedRole)

		// Step 5: Simulate finance user and approve final stage
		// Update user role to finance for final approval
		db.Model(&user).Update("role", "finance")

		// Approve the finance stage
		financeApproveReq := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/approvals/%s/approve", financeTask.ID), bytes.NewReader(approveBody))
		financeApproveReq.Header.Set("Content-Type", "application/json")

		financeApproveResp, err := app.Test(financeApproveReq)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, financeApproveResp.StatusCode)

		// Step 6: Verify workflow completion and document status update
		db.Where("id = ?", workflowAssignment.ID).First(&workflowAssignment)
		assert.Equal(t, "completed", workflowAssignment.Status)
		assert.NotNil(t, workflowAssignment.CompletedAt)

		// Verify requisition status was updated to "approved"
		db.Where("id = ?", requisitionID).First(&updatedRequisition)
		assert.Equal(t, "approved", updatedRequisition.Status)

		// Verify action history was added
		actionHistory := updatedRequisition.ActionHistory.Data()
		assert.Greater(t, len(actionHistory), 0)

		// Find the workflow completion entry
		var workflowCompletedEntry *types.ActionHistoryEntry
		for _, entry := range actionHistory {
			if entry.ActionType == "WORKFLOW_COMPLETED" {
				workflowCompletedEntry = &entry
				break
			}
		}
		assert.NotNil(t, workflowCompletedEntry)
		assert.Equal(t, "WORKFLOW_COMPLETED", workflowCompletedEntry.ActionType)
		assert.Equal(t, userID, workflowCompletedEntry.PerformedBy)

		// Step 7: Check if automation was triggered (PO creation)
		// Note: This depends on the automation service configuration
		// For this test, we'll just verify the automation fields are set correctly
		if updatedRequisition.AutomationUsed {
			assert.NotNil(t, updatedRequisition.AutoCreatedPO)
			
			// Verify a purchase order was created
			var createdPO models.PurchaseOrder
			err = db.Where("source_requisition_id = ?", requisitionID).First(&createdPO).Error
			if err == nil {
				assert.Equal(t, "draft", createdPO.Status)
				assert.Equal(t, updatedRequisition.TotalAmount, createdPO.TotalAmount)
			}
		}

		// Step 8: Verify approval status endpoint
		statusReq := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/documents/%s/approval-status", requisitionID), nil)
		statusResp, err := app.Test(statusReq)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, statusResp.StatusCode)

		var statusResponse types.DetailResponse
		json.NewDecoder(statusResp.Body).Decode(&statusResponse)
		assert.True(t, statusResponse.Success)

		statusData := statusResponse.Data.(map[string]interface{})
		assert.Equal(t, float64(2), statusData["currentStage"])
		assert.Equal(t, float64(2), statusData["totalStages"])
		assert.Equal(t, "completed", statusData["status"])
		assert.Equal(t, false, statusData["canApprove"]) // No more approvals needed
		assert.Equal(t, false, statusData["canReject"])  // No more rejections possible
	})
}

// TestWorkflowRejectionIntegration tests workflow rejection and document status update
func TestWorkflowRejectionIntegration(t *testing.T) {
	// Setup test database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// Auto-migrate all models
	err = db.AutoMigrate(
		&models.User{},
		&models.Organization{},
		&models.Vendor{},
		&models.Requisition{},
		&models.Workflow{},
		&models.WorkflowAssignment{},
		&models.WorkflowTask{},
	)
	assert.NoError(t, err)

	config.DB = db

	// Create test data
	orgID := uuid.New().String()
	userID := uuid.New().String()

	// Create organization and user
	org := models.Organization{ID: orgID, Name: "Test Organization"}
	db.Create(&org)

	user := models.User{
		ID:                    userID,
		Name:                  "Test User",
		Email:                 "test@example.com",
		Role:                  "manager",
		CurrentOrganizationID: &orgID,
		Active:                true,
	}
	db.Create(&user)

	// Create workflow
	workflowID := uuid.New().String()
	workflow := models.Workflow{
		ID:             workflowID,
		OrganizationID: orgID,
		Name:           "Test Workflow",
		DocumentType:   "REQUISITION",
		IsDefault:      true,
		IsActive:       true,
		Version:        1,
		Stages: datatypes.JSON(`[
			{
				"stageNumber": 1,
				"stageName": "Manager Approval",
				"requiredRole": "manager",
				"timeoutHours": 24
			}
		]`),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	db.Create(&workflow)

	// Initialize services
	auditService := services.NewAuditService()
	automationService := services.NewDocumentAutomationService(db, auditService, nil)
	workflowRepo := &SimpleWorkflowRepo{db: db}
	workflowService := services.NewWorkflowService(workflowRepo, auditService, db)
	workflowExecutionService := services.NewWorkflowExecutionService(db, workflowService, auditService, automationService)

	// Create requisition with workflow assignment
	requisitionID := uuid.New().String()
	requisition := models.Requisition{
		ID:             requisitionID,
		OrganizationID: orgID,
		DocumentNumber: "REQ-TEST-001",
		Title:          "Test Requisition",
		Status:         "pending",
		RequesterID:    userID,
		TotalAmount:    1000.00,
		Currency:       "USD",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	db.Create(&requisition)

	// Create workflow assignment and task
	assignment := models.WorkflowAssignment{
		ID:              uuid.New().String(),
		OrganizationID:  orgID,
		EntityID:        requisitionID,
		EntityType:      "REQUISITION",
		WorkflowID:      workflow.ID,
		WorkflowVersion: 1,
		CurrentStage:    1,
		Status:          "in_progress",
		AssignedAt:      time.Now(),
		AssignedBy:      userID,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	db.Create(&assignment)

	taskID := uuid.New().String()
	task := models.WorkflowTask{
		ID:                   taskID,
		OrganizationID:       orgID,
		WorkflowAssignmentID: assignment.ID,
		EntityID:             requisitionID,
		EntityType:           "REQUISITION",
		StageNumber:          1,
		StageName:            "Manager Approval",
		AssignmentType:       "role",
		AssignedRole:         &[]string{"manager"}[0],
		Status:               "pending",
		Priority:             "medium",
		CreatedAt:            time.Now(),
	}
	db.Create(&task)

	t.Run("Workflow Rejection Updates Document Status", func(t *testing.T) {
		// Reject the workflow task
		err := workflowExecutionService.RejectWorkflowTask(
			context.Background(),
			taskID,
			userID,
			"rejection_signature_123",
			"Insufficient justification provided",
		)
		assert.NoError(t, err)

		// Verify workflow assignment is marked as rejected
		var updatedAssignment models.WorkflowAssignment
		db.Where("id = ?", assignment.ID).First(&updatedAssignment)
		assert.Equal(t, "rejected", updatedAssignment.Status)
		assert.NotNil(t, updatedAssignment.CompletedAt)

		// Verify requisition status was updated to "rejected"
		var updatedRequisition models.Requisition
		db.Where("id = ?", requisitionID).First(&updatedRequisition)
		assert.Equal(t, "rejected", updatedRequisition.Status)

		// Verify action history was added
		actionHistory := updatedRequisition.ActionHistory.Data()
		assert.Greater(t, len(actionHistory), 0)

		// Find the workflow rejection entry
		var workflowRejectedEntry *types.ActionHistoryEntry
		for _, entry := range actionHistory {
			if entry.ActionType == "WORKFLOW_REJECTED" {
				workflowRejectedEntry = &entry
				break
			}
		}
		assert.NotNil(t, workflowRejectedEntry)
		assert.Equal(t, "WORKFLOW_REJECTED", workflowRejectedEntry.ActionType)
		assert.Equal(t, userID, workflowRejectedEntry.PerformedBy)
		assert.Equal(t, "Insufficient justification provided", workflowRejectedEntry.Comments)
	})
}