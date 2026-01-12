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

// TestCustomRoleWorkflowIntegration tests complete workflow integration with custom roles
func TestCustomRoleWorkflowIntegration(t *testing.T) {
	// Setup test database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// Auto-migrate all models
	err = db.AutoMigrate(
		&models.User{},
		&models.Organization{},
		&models.OrganizationRole{},
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

	// Create test organization with custom roles
	orgID, userIDs, err := createCustomRoleTestData(db)
	assert.NoError(t, err)

	// Setup Fiber app with handlers
	app := setupFiberAppWithCustomRoles(db)

	t.Run("Complete Custom Role Workflow: Requisition -> Custom Approval Chain", func(t *testing.T) {
		// Step 1: Create requisition
		requisitionData := types.CreateRequisitionRequest{
			Title:       "Custom Role Test Requisition",
			Description: "Testing custom role approval workflow",
			Department:  "Procurement",
			Priority:    "high",
			Items: []types.RequisitionItem{
				{
					Description: "Custom workflow test item",
					Quantity:    5,
					UnitPrice:   1000.00,
					Amount:      5000.00,
				},
			},
			TotalAmount: 5000.00,
			Currency:    "USD",
		}

		reqBody, _ := json.Marshal(requisitionData)
		req := httptest.NewRequest("POST", "/api/v1/requisitions", bytes.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-token")
		req.Header.Set("X-User-ID", userIDs["requester"])
		req.Header.Set("X-Organization-ID", orgID)

		resp, err := app.Test(req, -1)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var createResp struct {
			Data types.RequisitionResponse `json:"data"`
		}
		json.NewDecoder(resp.Body).Decode(&createResp)
		requisitionID := createResp.Data.ID

		// Step 2: Submit requisition for approval (triggers custom role workflow)
		submitData := types.SubmitRequisitionRequest{
			RequisitionID:   requisitionID,
			Comments:        "Submitting for custom role approval",
			SubmittedBy:     userIDs["requester"],
			SubmittedByName: "Test Requester",
			SubmittedByRole: "employee",
		}

		submitBody, _ := json.Marshal(submitData)
		submitReq := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/requisitions/%s/submit", requisitionID), bytes.NewReader(submitBody))
		submitReq.Header.Set("Content-Type", "application/json")
		submitReq.Header.Set("Authorization", "Bearer test-token")
		submitReq.Header.Set("X-User-ID", userIDs["requester"])
		submitReq.Header.Set("X-Organization-ID", orgID)

		submitResp, err := app.Test(submitReq, -1)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, submitResp.StatusCode)

		// Step 3: First approval by procurement_specialist (custom role)
		// Get workflow tasks for procurement specialist
		tasksReq := httptest.NewRequest("GET", "/api/v1/tasks/my-tasks", nil)
		tasksReq.Header.Set("Authorization", "Bearer test-token")
		tasksReq.Header.Set("X-User-ID", userIDs["procurement_specialist"])
		tasksReq.Header.Set("X-Organization-ID", orgID)

		tasksResp, err := app.Test(tasksReq, -1)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, tasksResp.StatusCode)

		var tasksResponse struct {
			Data []models.WorkflowTask `json:"data"`
		}
		json.NewDecoder(tasksResp.Body).Decode(&tasksResponse)
		assert.Greater(t, len(tasksResponse.Data), 0, "Procurement specialist should have workflow tasks")

		// Find the task for our requisition
		var procurementTask *models.WorkflowTask
		for _, task := range tasksResponse.Data {
			if task.EntityID == requisitionID && task.EntityType == "requisition" {
				procurementTask = &task
				break
			}
		}
		assert.NotNil(t, procurementTask, "Should find workflow task for requisition")
		assert.Equal(t, "procurement_specialist", *procurementTask.AssignedRole)

		// Approve with procurement specialist
		approvalData := types.ApprovalRequest{
			TaskID:        procurementTask.ID,
			Action:        "approve",
			Comments:      "Approved by procurement specialist - custom role",
			Signature:     "procurement-signature",
			ApproverID:    userIDs["procurement_specialist"],
			ApproverName:  "Procurement Specialist",
			ApproverRole:  "procurement_specialist",
		}

		approvalBody, _ := json.Marshal(approvalData)
		approvalReq := httptest.NewRequest("POST", "/api/v1/workflows/tasks/approve", bytes.NewReader(approvalBody))
		approvalReq.Header.Set("Content-Type", "application/json")
		approvalReq.Header.Set("Authorization", "Bearer test-token")
		approvalReq.Header.Set("X-User-ID", userIDs["procurement_specialist"])
		approvalReq.Header.Set("X-Organization-ID", orgID)

		approvalResp, err := app.Test(approvalReq, -1)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, approvalResp.StatusCode)

		// Step 4: Second approval by department_head_procurement (custom role)
		// Get workflow tasks for department head
		deptTasksReq := httptest.NewRequest("GET", "/api/v1/tasks/my-tasks", nil)
		deptTasksReq.Header.Set("Authorization", "Bearer test-token")
		deptTasksReq.Header.Set("X-User-ID", userIDs["department_head_procurement"])
		deptTasksReq.Header.Set("X-Organization-ID", orgID)

		deptTasksResp, err := app.Test(deptTasksReq, -1)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, deptTasksResp.StatusCode)

		var deptTasksResponse struct {
			Data []models.WorkflowTask `json:"data"`
		}
		json.NewDecoder(deptTasksResp.Body).Decode(&deptTasksResponse)

		// Find the task for our requisition (should be at stage 2 now)
		var deptTask *models.WorkflowTask
		for _, task := range deptTasksResponse.Data {
			if task.EntityID == requisitionID && task.EntityType == "requisition" && task.StageNumber == 2 {
				deptTask = &task
				break
			}
		}
		assert.NotNil(t, deptTask, "Should find stage 2 workflow task for requisition")
		assert.Equal(t, "department_head_procurement", *deptTask.AssignedRole)

		// Approve with department head
		deptApprovalData := types.ApprovalRequest{
			TaskID:        deptTask.ID,
			Action:        "approve",
			Comments:      "Final approval by department head - custom role",
			Signature:     "dept-head-signature",
			ApproverID:    userIDs["department_head_procurement"],
			ApproverName:  "Department Head Procurement",
			ApproverRole:  "department_head_procurement",
		}

		deptApprovalBody, _ := json.Marshal(deptApprovalData)
		deptApprovalReq := httptest.NewRequest("POST", "/api/v1/workflows/tasks/approve", bytes.NewReader(deptApprovalBody))
		deptApprovalReq.Header.Set("Content-Type", "application/json")
		deptApprovalReq.Header.Set("Authorization", "Bearer test-token")
		deptApprovalReq.Header.Set("X-User-ID", userIDs["department_head_procurement"])
		deptApprovalReq.Header.Set("X-Organization-ID", orgID)

		deptApprovalResp, err := app.Test(deptApprovalReq, -1)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, deptApprovalResp.StatusCode)

		// Step 5: Verify requisition is fully approved
		finalReq := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/requisitions/%s", requisitionID), nil)
		finalReq.Header.Set("Authorization", "Bearer test-token")
		finalReq.Header.Set("X-User-ID", userIDs["requester"])
		finalReq.Header.Set("X-Organization-ID", orgID)

		finalResp, err := app.Test(finalReq, -1)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, finalResp.StatusCode)

		var finalRequisition struct {
			Data types.RequisitionResponse `json:"data"`
		}
		json.NewDecoder(finalResp.Body).Decode(&finalRequisition)
		assert.Equal(t, "approved", finalRequisition.Data.Status)

		// Verify approval history contains custom roles
		assert.Greater(t, len(finalRequisition.Data.ApprovalHistory), 1)
		
		// Check that custom roles are recorded in approval history
		foundProcurementApproval := false
		foundDeptHeadApproval := false
		
		for _, approval := range finalRequisition.Data.ApprovalHistory {
			if approval.ApproverRole == "procurement_specialist" && approval.Status == "approved" {
				foundProcurementApproval = true
			}
			if approval.ApproverRole == "department_head_procurement" && approval.Status == "approved" {
				foundDeptHeadApproval = true
			}
		}
		
		assert.True(t, foundProcurementApproval, "Should find procurement specialist approval in history")
		assert.True(t, foundDeptHeadApproval, "Should find department head approval in history")
	})
}

// TestCustomRoleRejectionWorkflow tests rejection scenarios with custom roles
func TestCustomRoleRejectionWorkflow(t *testing.T) {
	// Setup test database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// Auto-migrate all models
	err = db.AutoMigrate(
		&models.User{},
		&models.Organization{},
		&models.OrganizationRole{},
		&models.Requisition{},
		&models.Workflow{},
		&models.WorkflowAssignment{},
		&models.WorkflowTask{},
	)
	assert.NoError(t, err)

	config.DB = db

	// Create test data
	orgID, userIDs, err := createCustomRoleTestData(db)
	assert.NoError(t, err)

	app := setupFiberAppWithCustomRoles(db)

	t.Run("Custom role rejects workflow and document returns to draft", func(t *testing.T) {
		// Create and submit requisition (similar to approval test)
		requisitionData := types.CreateRequisitionRequest{
			Title:       "Rejection Test Requisition",
			Description: "Testing custom role rejection workflow",
			Department:  "Procurement",
			Priority:    "medium",
			Items: []types.RequisitionItem{
				{
					Description: "Rejection test item",
					Quantity:    1,
					UnitPrice:   2000.00,
					Amount:      2000.00,
				},
			},
			TotalAmount: 2000.00,
			Currency:    "USD",
		}

		// Create requisition
		reqBody, _ := json.Marshal(requisitionData)
		req := httptest.NewRequest("POST", "/api/v1/requisitions", bytes.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-token")
		req.Header.Set("X-User-ID", userIDs["requester"])
		req.Header.Set("X-Organization-ID", orgID)

		resp, err := app.Test(req, -1)
		assert.NoError(t, err)

		var createResp struct {
			Data types.RequisitionResponse `json:"data"`
		}
		json.NewDecoder(resp.Body).Decode(&createResp)
		requisitionID := createResp.Data.ID

		// Submit for approval
		submitData := types.SubmitRequisitionRequest{
			RequisitionID:   requisitionID,
			Comments:        "Submitting for rejection test",
			SubmittedBy:     userIDs["requester"],
			SubmittedByName: "Test Requester",
			SubmittedByRole: "employee",
		}

		submitBody, _ := json.Marshal(submitData)
		submitReq := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/requisitions/%s/submit", requisitionID), bytes.NewReader(submitBody))
		submitReq.Header.Set("Content-Type", "application/json")
		submitReq.Header.Set("Authorization", "Bearer test-token")
		submitReq.Header.Set("X-User-ID", userIDs["requester"])
		submitReq.Header.Set("X-Organization-ID", orgID)

		app.Test(submitReq, -1)

		// Get workflow task for procurement specialist
		tasksReq := httptest.NewRequest("GET", "/api/v1/tasks/my-tasks", nil)
		tasksReq.Header.Set("Authorization", "Bearer test-token")
		tasksReq.Header.Set("X-User-ID", userIDs["procurement_specialist"])
		tasksReq.Header.Set("X-Organization-ID", orgID)

		tasksResp, err := app.Test(tasksReq, -1)
		assert.NoError(t, err)

		var tasksResponse struct {
			Data []models.WorkflowTask `json:"data"`
		}
		json.NewDecoder(tasksResp.Body).Decode(&tasksResponse)

		// Find and reject the task
		var procurementTask *models.WorkflowTask
		for _, task := range tasksResponse.Data {
			if task.EntityID == requisitionID && task.EntityType == "requisition" {
				procurementTask = &task
				break
			}
		}
		assert.NotNil(t, procurementTask)

		// Reject with custom role
		rejectionData := types.ApprovalRequest{
			TaskID:        procurementTask.ID,
			Action:        "reject",
			Comments:      "Rejected by procurement specialist - insufficient justification",
			Signature:     "procurement-rejection-signature",
			ApproverID:    userIDs["procurement_specialist"],
			ApproverName:  "Procurement Specialist",
			ApproverRole:  "procurement_specialist",
		}

		rejectionBody, _ := json.Marshal(rejectionData)
		rejectionReq := httptest.NewRequest("POST", "/api/v1/workflows/tasks/reject", bytes.NewReader(rejectionBody))
		rejectionReq.Header.Set("Content-Type", "application/json")
		rejectionReq.Header.Set("Authorization", "Bearer test-token")
		rejectionReq.Header.Set("X-User-ID", userIDs["procurement_specialist"])
		rejectionReq.Header.Set("X-Organization-ID", orgID)

		rejectionResp, err := app.Test(rejectionReq, -1)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rejectionResp.StatusCode)

		// Verify requisition is rejected
		finalReq := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/requisitions/%s", requisitionID), nil)
		finalReq.Header.Set("Authorization", "Bearer test-token")
		finalReq.Header.Set("X-User-ID", userIDs["requester"])
		finalReq.Header.Set("X-Organization-ID", orgID)

		finalResp, err := app.Test(finalReq, -1)
		assert.NoError(t, err)

		var finalRequisition struct {
			Data types.RequisitionResponse `json:"data"`
		}
		json.NewDecoder(finalResp.Body).Decode(&finalRequisition)
		assert.Equal(t, "rejected", finalRequisition.Data.Status)

		// Verify rejection is recorded with custom role
		assert.Greater(t, len(finalRequisition.Data.ApprovalHistory), 0)
		
		rejectionFound := false
		for _, approval := range finalRequisition.Data.ApprovalHistory {
			if approval.ApproverRole == "procurement_specialist" && approval.Status == "rejected" {
				rejectionFound = true
				assert.Contains(t, approval.Comments, "insufficient justification")
			}
		}
		assert.True(t, rejectionFound, "Should find procurement specialist rejection in history")
	})
}

// createCustomRoleTestData creates test organization with custom roles and users
func createCustomRoleTestData(db *gorm.DB) (string, map[string]string, error) {
	orgID := uuid.New().String()
	
	// Create organization
	org := &models.Organization{
		ID:   orgID,
		Name: "Custom Role Test Organization",
		Tier: "enterprise",
	}
	if err := db.Create(org).Error; err != nil {
		return "", nil, err
	}

	// Create custom roles
	customRoles := []models.OrganizationRole{
		{
			ID:             uuid.New().String(),
			OrganizationID: orgID,
			Name:           "procurement_specialist",
			DisplayName:    "Procurement Specialist",
			Description:    "Specialist in procurement processes",
			IsSystemRole:   false,
			IsActive:       true,
			Permissions:    datatypes.JSON(`["approve_low_value_requisitions", "create_purchase_orders"]`),
		},
		{
			ID:             uuid.New().String(),
			OrganizationID: orgID,
			Name:           "department_head_procurement",
			DisplayName:    "Procurement Department Head",
			Description:    "Head of Procurement Department",
			IsSystemRole:   false,
			IsActive:       true,
			Permissions:    datatypes.JSON(`["approve_all_requisitions", "manage_procurement_team", "approve_high_value_requisitions"]`),
		},
		{
			ID:             uuid.New().String(),
			OrganizationID: orgID,
			Name:           "finance_controller",
			DisplayName:    "Finance Controller",
			Description:    "Controller in finance department",
			IsSystemRole:   false,
			IsActive:       true,
			Permissions:    datatypes.JSON(`["approve_budgets", "review_financial_documents", "approve_payments"]`),
		},
	}

	for _, role := range customRoles {
		if err := db.Create(&role).Error; err != nil {
			return "", nil, err
		}
	}

	// Create users with custom roles
	userIDs := make(map[string]string)
	
	users := []struct {
		key  string
		role string
		name string
	}{
		{"requester", "employee", "Test Requester"},
		{"procurement_specialist", "procurement_specialist", "Procurement Specialist"},
		{"department_head_procurement", "department_head_procurement", "Procurement Department Head"},
		{"finance_controller", "finance_controller", "Finance Controller"},
	}

	for _, u := range users {
		userID := uuid.New().String()
		user := &models.User{
			ID:             userID,
			OrganizationID: orgID,
			Email:          fmt.Sprintf("%s@test.com", u.key),
			Name:           u.name,
			Role:           u.role,
			IsActive:       true,
		}
		if err := db.Create(user).Error; err != nil {
			return "", nil, err
		}
		userIDs[u.key] = userID
	}

	// Create workflow with custom roles
	workflowID := uuid.New().String()
	stages := []models.WorkflowStage{
		{
			StageNumber:   1,
			StageName:     "Procurement Specialist Review",
			RequiredRole:  "procurement_specialist",
			TimeoutHours:  func(i int) *int { return &i }(24),
		},
		{
			StageNumber:   2,
			StageName:     "Department Head Approval",
			RequiredRole:  "department_head_procurement",
			TimeoutHours:  func(i int) *int { return &i }(48),
		},
	}
	
	stagesJSON, _ := json.Marshal(stages)
	
	workflow := &models.Workflow{
		ID:             workflowID,
		OrganizationID: orgID,
		Name:           "Custom Role Procurement Workflow",
		DocumentType:   "requisition",
		IsDefault:      true,
		IsActive:       true,
		Stages:         datatypes.JSON(stagesJSON),
	}
	
	if err := db.Create(workflow).Error; err != nil {
		return "", nil, err
	}

	return orgID, userIDs, nil
}

// setupFiberAppWithCustomRoles sets up Fiber app with all necessary handlers for custom role testing
func setupFiberAppWithCustomRoles(db *gorm.DB) *fiber.App {
	app := fiber.New()

	// Create services
	auditService := services.NewAuditService(db)
	notificationService := services.NewNotificationService(db)
	automationService := services.NewDocumentAutomationService(db, auditService, notificationService)
	
	// Create repositories
	requisitionRepo := repository.NewRequisitionRepository(nil, db)
	workflowRepo := &SimpleWorkflowRepo{db: db}
	
	// Create services
	requisitionService := services.NewRequisitionService(requisitionRepo, auditService)
	workflowExecutionService := services.NewWorkflowExecutionService(db, workflowRepo, auditService, automationService)
	rbacService := services.NewRBACService(db)
	
	// Create handlers
	requisitionHandler := handlers.NewRequisitionHandler(requisitionService)
	workflowHandler := handlers.NewWorkflowHandler(workflowExecutionService)
	taskHandler := handlers.NewTaskHandler(db, workflowExecutionService)
	
	// Add middleware
	app.Use(func(c *fiber.Ctx) error {
		// Mock authentication middleware
		c.Locals("userID", c.Get("X-User-ID"))
		c.Locals("organizationID", c.Get("X-Organization-ID"))
		return c.Next()
	})

	// Setup routes
	api := app.Group("/api/v1")
	
	// Requisition routes
	api.Post("/requisitions", requisitionHandler.CreateRequisition)
	api.Get("/requisitions/:id", requisitionHandler.GetRequisitionByID)
	api.Post("/requisitions/:id/submit", requisitionHandler.SubmitRequisitionForApproval)
	
	// Workflow routes
	api.Post("/workflows/tasks/approve", workflowHandler.ApproveTask)
	api.Post("/workflows/tasks/reject", workflowHandler.RejectTask)
	
	// Task routes
	api.Get("/tasks/my-tasks", taskHandler.GetMyTasks)

	return app
}

// SimpleWorkflowRepo implements basic workflow repository for testing
type SimpleWorkflowRepo struct {
	db *gorm.DB
}

func (r *SimpleWorkflowRepo) GetByID(ctx context.Context, id string) (*models.Workflow, error) {
	var workflow models.Workflow
	err := r.db.Where("id = ?", id).First(&workflow).Error
	return &workflow, err
}

func (r *SimpleWorkflowRepo) GetDefaultByDocumentType(ctx context.Context, organizationID, documentType string) (*models.Workflow, error) {
	var workflow models.Workflow
	err := r.db.Where("organization_id = ? AND document_type = ? AND is_default = ? AND is_active = ?", 
		organizationID, documentType, true, true).First(&workflow).Error
	return &workflow, err
}

func (r *SimpleWorkflowRepo) Create(ctx context.Context, workflow *models.Workflow) error {
	return r.db.Create(workflow).Error
}

func (r *SimpleWorkflowRepo) Update(ctx context.Context, workflow *models.Workflow) error {
	return r.db.Save(workflow).Error
}

func (r *SimpleWorkflowRepo) Delete(ctx context.Context, id string) error {
	return r.db.Delete(&models.Workflow{}, "id = ?", id).Error
}

func (r *SimpleWorkflowRepo) List(ctx context.Context, organizationID string, filters map[string]interface{}) ([]*models.Workflow, error) {
	var workflows []*models.Workflow
	query := r.db.Where("organization_id = ?", organizationID)
	
	for key, value := range filters {
		query = query.Where(key+" = ?", value)
	}
	
	err := query.Find(&workflows).Error
	return workflows, err
}