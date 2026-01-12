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
	"github.com/liyali/liyali-gateway/handlers"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/services"
	"github.com/liyali/liyali-gateway/tests/helpers"
	"github.com/liyali/liyali-gateway/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type WorkflowAPITestSuite struct {
	app                      *fiber.App
	testDB                   *helpers.TestDatabase
	workflowExecutionService *services.WorkflowExecutionService
	scenario                 *helpers.WorkflowTestScenario
}

func setupWorkflowAPITest(t *testing.T) *WorkflowAPITestSuite {
	// Initialize test database
	testDB := helpers.SetupTestDB(t)

	// Create test services
	workflowService := services.NewWorkflowService(nil, nil, testDB.DB) // Simplified for testing
	auditService := services.NewAuditService()
	automationService := services.NewDocumentAutomationService(testDB.DB, auditService, nil)
	workflowExecutionService := services.NewWorkflowExecutionService(testDB.DB, workflowService, auditService, automationService)

	// Create handler registry
	handlerRegistry := &handlers.HandlerRegistry{
		WorkflowExecutionService: workflowExecutionService,
		Approval:                 handlers.NewApprovalHandler(),
	}

	// Create Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// Setup routes (simplified for testing)
	setupTestRoutes(app, handlerRegistry, workflowExecutionService)

	// Create complete test scenario
	scenario := helpers.CreateCompleteWorkflowScenario(t, testDB.DB, "requisition")

	return &WorkflowAPITestSuite{
		app:                      app,
		testDB:                   testDB,
		workflowExecutionService: workflowExecutionService,
		scenario:                 scenario,
	}
}

func setupTestRoutes(app *fiber.App, handlerRegistry *handlers.HandlerRegistry, workflowService *services.WorkflowExecutionService) {
	// Add test middleware to inject services
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("workflowExecutionService", workflowService)
		c.Locals("organizationID", "test-org")
		c.Locals("userID", "test-user")
		return c.Next()
	})

	api := app.Group("/api/v1")

	// Approval routes
	approvals := api.Group("/approvals")
	approvals.Post("/tasks/:id/claim", handlerRegistry.Approval.ClaimTask)
	approvals.Post("/tasks/:id/unclaim", handlerRegistry.Approval.UnclaimTask)
	approvals.Post("/:id/approve", handlerRegistry.Approval.ApproveTask)
	approvals.Post("/:id/reject", handlerRegistry.Approval.RejectTask)
	approvals.Get("/available-approvers", handlerRegistry.Approval.GetAvailableApprovers)

	// Document routes
	documents := api.Group("/documents")
	documents.Get("/:documentId/approval-status", handlerRegistry.Approval.GetApprovalWorkflowStatus)
	documents.Get("/:documentId/approval-history", handlerRegistry.Approval.GetApprovalHistory)
}

func (suite *WorkflowAPITestSuite) cleanup() {
	suite.testDB.Cleanup()
}

func TestWorkflowAPI_ClaimTask(t *testing.T) {
	suite := setupWorkflowAPITest(t)
	defer suite.cleanup()

	tests := []struct {
		name           string
		taskID         string
		userID         string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "successful claim",
			taskID:         suite.scenario.Task.ID,
			userID:         suite.scenario.Users.Manager.ID,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "claim non-existent task",
			taskID:         "non-existent",
			userID:         suite.scenario.Users.Manager.ID,
			expectedStatus: http.StatusConflict,
			expectedError:  "not available for claiming",
		},
		{
			name:           "claim already claimed task",
			taskID:         suite.scenario.Task.ID,
			userID:         suite.scenario.Users.Requester.ID,
			expectedStatus: http.StatusConflict,
			expectedError:  "not available for claiming",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set user context
			req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/approvals/tasks/%s/claim", tt.taskID), nil)
			req.Header.Set("Content-Type", "application/json")

			// Override user context for this test
			suite.app.Use(func(c *fiber.Ctx) error {
				c.Locals("userID", tt.userID)
				c.Locals("organizationID", suite.scenario.Organization.ID)
				return c.Next()
			})

			resp, err := suite.app.Test(req, -1)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.expectedError != "" {
				var response map[string]interface{}
				err = json.NewDecoder(resp.Body).Decode(&response)
				require.NoError(t, err)
				assert.Contains(t, response["message"].(string), tt.expectedError)
			}

			if tt.expectedStatus == http.StatusOK {
				// Verify task is claimed in database
				helpers.AssertTaskStatus(t, suite.testDB.DB, tt.taskID, "claimed")
				
				var task models.WorkflowTask
				err = suite.testDB.DB.Where("id = ?", tt.taskID).First(&task).Error
				require.NoError(t, err)
				assert.Equal(t, tt.userID, *task.ClaimedBy)
				assert.NotNil(t, task.ClaimExpiry)
			}
		})
	}
}

func TestWorkflowAPI_UnclaimTask(t *testing.T) {
	suite := setupWorkflowAPITest(t)
	defer suite.cleanup()

	// First claim the task
	err := suite.workflowExecutionService.ClaimWorkflowTask(context.Background(), suite.scenario.Task.ID, suite.scenario.Users.Manager.ID)
	require.NoError(t, err)

	tests := []struct {
		name           string
		taskID         string
		userID         string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "successful unclaim",
			taskID:         suite.scenario.Task.ID,
			userID:         suite.scenario.Users.Manager.ID,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "unclaim non-existent task",
			taskID:         "non-existent",
			userID:         suite.scenario.Users.Manager.ID,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/approvals/tasks/%s/unclaim", tt.taskID), nil)
			req.Header.Set("Content-Type", "application/json")

			// Override user context
			suite.app.Use(func(c *fiber.Ctx) error {
				c.Locals("userID", tt.userID)
				c.Locals("organizationID", suite.scenario.Organization.ID)
				return c.Next()
			})

			resp, err := suite.app.Test(req, -1)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.expectedStatus == http.StatusOK {
				// Verify task is unclaimed in database
				helpers.AssertTaskStatus(t, suite.testDB.DB, tt.taskID, "pending")
				
				var task models.WorkflowTask
				err = suite.testDB.DB.Where("id = ?", tt.taskID).First(&task).Error
				require.NoError(t, err)
				assert.Nil(t, task.ClaimedBy)
				assert.Nil(t, task.ClaimExpiry)
			}
		})
	}
}

func TestWorkflowAPI_ApproveTask(t *testing.T) {
	suite := setupWorkflowAPITest(t)
	defer suite.cleanup()

	// First claim the task
	err := suite.workflowExecutionService.ClaimWorkflowTask(context.Background(), suite.scenario.Task.ID, suite.scenario.Users.Manager.ID)
	require.NoError(t, err)

	approveRequest := map[string]interface{}{
		"signature":       "test-signature",
		"comment":         "Approved for testing",
		"expectedVersion": 2, // Version should be 2 after claiming
	}

	requestBody, _ := json.Marshal(approveRequest)
	req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/approvals/%s/approve", suite.scenario.Task.ID), bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")

	// Set user context
	suite.app.Use(func(c *fiber.Ctx) error {
		c.Locals("userID", suite.scenario.Users.Manager.ID)
		c.Locals("organizationID", suite.scenario.Organization.ID)
		return c.Next()
	})

	resp, err := suite.app.Test(req, -1)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Verify task is completed in database
	helpers.AssertTaskStatus(t, suite.testDB.DB, suite.scenario.Task.ID, "completed")
	
	var task models.WorkflowTask
	err = suite.testDB.DB.Where("id = ?", suite.scenario.Task.ID).First(&task).Error
	require.NoError(t, err)
	assert.NotNil(t, task.CompletedAt)

	// Verify approval record was created
	helpers.AssertApprovalRecordExists(t, suite.testDB.DB, suite.scenario.Task.ID, "approved", suite.scenario.Users.Manager.ID)
}

func TestWorkflowAPI_RejectTask(t *testing.T) {
	suite := setupWorkflowAPITest(t)
	defer suite.cleanup()

	// First claim the task
	err := suite.workflowExecutionService.ClaimWorkflowTask(context.Background(), suite.scenario.Task.ID, suite.scenario.Users.Manager.ID)
	require.NoError(t, err)

	rejectRequest := map[string]interface{}{
		"signature":       "test-signature",
		"reason":          "Rejected for testing",
		"expectedVersion": 2, // Version should be 2 after claiming
	}

	requestBody, _ := json.Marshal(rejectRequest)
	req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/approvals/%s/reject", suite.scenario.Task.ID), bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")

	// Set user context
	suite.app.Use(func(c *fiber.Ctx) error {
		c.Locals("userID", suite.scenario.Users.Manager.ID)
		c.Locals("organizationID", suite.scenario.Organization.ID)
		return c.Next()
	})

	resp, err := suite.app.Test(req, -1)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Verify task is completed in database
	helpers.AssertTaskStatus(t, suite.testDB.DB, suite.scenario.Task.ID, "completed")
	
	var task models.WorkflowTask
	err = suite.testDB.DB.Where("id = ?", suite.scenario.Task.ID).First(&task).Error
	require.NoError(t, err)
	assert.NotNil(t, task.CompletedAt)

	// Verify rejection record was created
	helpers.AssertApprovalRecordExists(t, suite.testDB.DB, suite.scenario.Task.ID, "rejected", suite.scenario.Users.Manager.ID)

	// Verify workflow assignment is rejected
	helpers.AssertWorkflowStatus(t, suite.testDB.DB, suite.scenario.Assignment.ID, "rejected")
}

func TestWorkflowAPI_GetWorkflowStatus(t *testing.T) {
	suite := setupWorkflowAPITest(t)
	defer suite.cleanup()

	req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/documents/%s/approval-status", suite.scenario.Document.ID), nil)

	// Set user context
	suite.app.Use(func(c *fiber.Ctx) error {
		c.Locals("userID", suite.scenario.Users.Manager.ID)
		c.Locals("organizationID", suite.scenario.Organization.ID)
		return c.Next()
	})

	resp, err := suite.app.Test(req, -1)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	data := response["data"].(map[string]interface{})
	assert.Equal(t, float64(1), data["currentStage"])
	assert.Equal(t, float64(1), data["totalStages"])
	assert.Equal(t, "in_progress", data["status"])
}

func TestWorkflowAPI_GetAvailableApprovers(t *testing.T) {
	suite := setupWorkflowAPITest(t)
	defer suite.cleanup()

	req := httptest.NewRequest("GET", "/api/v1/approvals/available-approvers?documentType=requisition", nil)

	// Set user context
	suite.app.Use(func(c *fiber.Ctx) error {
		c.Locals("userID", suite.scenario.Users.Manager.ID)
		c.Locals("organizationID", suite.scenario.Organization.ID)
		return c.Next()
	})

	resp, err := suite.app.Test(req, -1)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	// Should return available approvers
	assert.True(t, response["success"].(bool))
	data := response["data"].([]interface{})
	assert.GreaterOrEqual(t, len(data), 1) // Should have at least the test manager
}

func TestWorkflowAPI_ConcurrentClaims(t *testing.T) {
	suite := setupWorkflowAPITest(t)
	defer suite.cleanup()

	// Create another manager
	anotherManager := &models.User{
		ID:                    "another-manager-" + uuid.New().String()[:8],
		Email:                 "another@example.com",
		Name:                  "Another Manager",
		Role:                  "manager",
		CurrentOrganizationID: &suite.scenario.Organization.ID,
		Active:                true,
	}
	require.NoError(t, suite.testDB.DB.Create(anotherManager).Error)

	// Test concurrent claims
	results := make(chan error, 2)

	// First claim
	go func() {
		err := suite.workflowExecutionService.ClaimWorkflowTask(context.Background(), suite.scenario.Task.ID, suite.scenario.Users.Manager.ID)
		results <- err
	}()

	// Second claim (should fail)
	go func() {
		time.Sleep(10 * time.Millisecond) // Small delay to ensure first claim starts first
		err := suite.workflowExecutionService.ClaimWorkflowTask(context.Background(), suite.scenario.Task.ID, anotherManager.ID)
		results <- err
	}()

	// Collect results
	var errors []error
	for i := 0; i < 2; i++ {
		errors = append(errors, <-results)
	}

	// One should succeed, one should fail
	successCount := 0
	failureCount := 0
	for _, err := range errors {
		if err == nil {
			successCount++
		} else {
			failureCount++
			assert.Contains(t, err.Error(), "not available for claiming")
		}
	}

	assert.Equal(t, 1, successCount, "Exactly one claim should succeed")
	assert.Equal(t, 1, failureCount, "Exactly one claim should fail")

	// Verify only one task is claimed
	helpers.AssertTaskStatus(t, suite.testDB.DB, suite.scenario.Task.ID, "claimed")
	
	var task models.WorkflowTask
	err := suite.testDB.DB.Where("id = ?", suite.scenario.Task.ID).First(&task).Error
	require.NoError(t, err)
	assert.NotNil(t, task.ClaimedBy)
}

func TestWorkflowAPI_OptimisticLocking(t *testing.T) {
	suite := setupWorkflowAPITest(t)
	defer suite.cleanup()

	// Claim the task
	err := suite.workflowExecutionService.ClaimWorkflowTask(context.Background(), suite.scenario.Task.ID, suite.scenario.Users.Manager.ID)
	require.NoError(t, err)

	// Get current version
	var task models.WorkflowTask
	err = suite.testDB.DB.Where("id = ?", suite.scenario.Task.ID).First(&task).Error
	require.NoError(t, err)
	currentVersion := task.Version

	// Try to approve with wrong version
	approveRequest := map[string]interface{}{
		"signature":       "test-signature",
		"comment":         "Approved for testing",
		"expectedVersion": currentVersion - 1, // Wrong version
	}

	requestBody, _ := json.Marshal(approveRequest)
	req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/approvals/%s/approve", suite.scenario.Task.ID), bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")

	// Set user context
	suite.app.Use(func(c *fiber.Ctx) error {
		c.Locals("userID", suite.scenario.Users.Manager.ID)
		c.Locals("organizationID", suite.scenario.Organization.ID)
		return c.Next()
	})

	resp, err := suite.app.Test(req, -1)
	require.NoError(t, err)

	assert.Equal(t, http.StatusConflict, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)
	assert.Contains(t, response["message"].(string), "modified by another user")
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}