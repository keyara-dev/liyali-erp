package helpers

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/models"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TestDatabase represents a test database instance
type TestDatabase struct {
	DB *gorm.DB
}

// SetupTestDB creates an in-memory SQLite database for testing
func SetupTestDB(t *testing.T) *TestDatabase {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Reduce noise in tests
	})
	require.NoError(t, err)

	// Auto-migrate all models
	err = db.AutoMigrate(
		&models.Organization{},
		&models.User{},
		&models.Workflow{},
		&models.WorkflowAssignment{},
		&models.WorkflowTask{},
		&models.StageApprovalRecord{},
		&models.Requisition{},
		&models.Budget{},
		&models.PurchaseOrder{},
		&models.PaymentVoucher{},
		&models.GoodsReceivedNote{},
		&models.Category{},
		&models.Vendor{},
	)
	require.NoError(t, err)

	return &TestDatabase{DB: db}
}

// Cleanup closes the test database connection
func (td *TestDatabase) Cleanup() {
	sqlDB, _ := td.DB.DB()
	if sqlDB != nil {
		sqlDB.Close()
	}
}

// TestDataBuilder helps create test data
type TestDataBuilder struct {
	db             *gorm.DB
	organizationID string
	userID         string
	managerID      string
	financeID      string
}

// NewTestDataBuilder creates a new test data builder
func NewTestDataBuilder(db *gorm.DB) *TestDataBuilder {
	return &TestDataBuilder{
		db:             db,
		organizationID: "test-org-" + uuid.New().String()[:8],
		userID:         "test-user-" + uuid.New().String()[:8],
		managerID:      "test-manager-" + uuid.New().String()[:8],
		financeID:      "test-finance-" + uuid.New().String()[:8],
	}
}

// CreateOrganization creates a test organization
func (tdb *TestDataBuilder) CreateOrganization(t *testing.T) *models.Organization {
	org := &models.Organization{
		ID:   tdb.organizationID,
		Name: "Test Organization",
		Slug: "test-org",
	}
	require.NoError(t, tdb.db.Create(org).Error)
	return org
}

// CreateUsers creates test users with different roles
func (tdb *TestDataBuilder) CreateUsers(t *testing.T) (*models.User, *models.User, *models.User) {
	// Regular user
	user := &models.User{
		ID:                    tdb.userID,
		Email:                 "user@example.com",
		Name:                  "Test User",
		Role:                  "requester",
		CurrentOrganizationID: &tdb.organizationID,
		Active:                true,
	}
	require.NoError(t, tdb.db.Create(user).Error)

	// Manager
	manager := &models.User{
		ID:                    tdb.managerID,
		Email:                 "manager@example.com",
		Name:                  "Test Manager",
		Role:                  "manager",
		CurrentOrganizationID: &tdb.organizationID,
		Active:                true,
	}
	require.NoError(t, tdb.db.Create(manager).Error)

	// Finance user
	finance := &models.User{
		ID:                    tdb.financeID,
		Email:                 "finance@example.com",
		Name:                  "Test Finance",
		Role:                  "finance",
		CurrentOrganizationID: &tdb.organizationID,
		Active:                true,
	}
	require.NoError(t, tdb.db.Create(finance).Error)

	return user, manager, finance
}

// CreateWorkflow creates a test workflow
func (tdb *TestDataBuilder) CreateWorkflow(t *testing.T, entityType string, stages []models.WorkflowStage) *models.Workflow {
	workflow := &models.Workflow{
		ID:             uuid.New(),
		OrganizationID: tdb.organizationID,
		Name:           fmt.Sprintf("Test %s Workflow", entityType),
		EntityType:     entityType,
		IsActive:       true,
		IsDefault:      true,
		CreatedBy:      tdb.userID,
	}

	require.NoError(t, workflow.SetStages(stages))
	require.NoError(t, tdb.db.Create(workflow).Error)
	return workflow
}

// CreateSingleStageWorkflow creates a simple single-stage workflow
func (tdb *TestDataBuilder) CreateSingleStageWorkflow(t *testing.T, entityType string) *models.Workflow {
	stages := []models.WorkflowStage{
		{
			StageNumber:           1,
			StageName:             "Manager Approval",
			RequiredRole:          "manager",
			RequiredApprovals:     1,
			RequiredApprovalCount: 1,
			ApprovalType:          "any",
			CanReject:             true,
		},
	}
	return tdb.CreateWorkflow(t, entityType, stages)
}

// CreateMultiStageWorkflow creates a multi-stage workflow
func (tdb *TestDataBuilder) CreateMultiStageWorkflow(t *testing.T, entityType string) *models.Workflow {
	stages := []models.WorkflowStage{
		{
			StageNumber:           1,
			StageName:             "Manager Approval",
			RequiredRole:          "manager",
			RequiredApprovals:     1,
			RequiredApprovalCount: 1,
			ApprovalType:          "any",
			CanReject:             true,
		},
		{
			StageNumber:           2,
			StageName:             "Finance Approval",
			RequiredRole:          "finance",
			RequiredApprovals:     1,
			RequiredApprovalCount: 1,
			ApprovalType:          "any",
			CanReject:             true,
		},
	}
	return tdb.CreateWorkflow(t, entityType, stages)
}

// CreateWorkflowAssignment creates a workflow assignment
func (tdb *TestDataBuilder) CreateWorkflowAssignment(t *testing.T, workflow *models.Workflow, entityID string) *models.WorkflowAssignment {
	assignment := &models.WorkflowAssignment{
		ID:              uuid.New().String(),
		OrganizationID:  tdb.organizationID,
		EntityID:        entityID,
		EntityType:      workflow.EntityType,
		WorkflowID:      workflow.ID,
		WorkflowVersion: workflow.Version,
		CurrentStage:    1,
		Status: "IN_PROGRESS",
		AssignedBy:      tdb.userID,
		AssignedAt:      time.Now(),
	}
	require.NoError(t, tdb.db.Create(assignment).Error)
	return assignment
}

// CreateWorkflowTask creates a workflow task
func (tdb *TestDataBuilder) CreateWorkflowTask(t *testing.T, assignment *models.WorkflowAssignment, stageNumber int, requiredRole string) *models.WorkflowTask {
	task := &models.WorkflowTask{
		ID:                   uuid.New().String(),
		OrganizationID:       assignment.OrganizationID,
		WorkflowAssignmentID: assignment.ID,
		EntityID:             assignment.EntityID,
		EntityType:           assignment.EntityType,
		StageNumber:          stageNumber,
		StageName:            fmt.Sprintf("Stage %d", stageNumber),
		AssignmentType:       "role",
		AssignedRole:         &requiredRole,
		Status: "PENDING",
		Priority:             "medium",
		Version:              1,
		CreatedAt:            time.Now(),
	}
	require.NoError(t, tdb.db.Create(task).Error)
	return task
}

// CreateRequisition creates a test requisition
func (tdb *TestDataBuilder) CreateRequisition(t *testing.T) *models.Requisition {
	requisition := &models.Requisition{
		ID:             uuid.New().String(),
		OrganizationID: tdb.organizationID,
		DocumentNumber: fmt.Sprintf("REQ-%d", time.Now().Unix()),
		Title:          "Test Requisition",
		Description:    "Test requisition for workflow testing",
		Status: "DRAFT",
		RequesterId:    tdb.userID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	require.NoError(t, tdb.db.Create(requisition).Error)
	return requisition
}

// GetOrganizationID returns the test organization ID
func (tdb *TestDataBuilder) GetOrganizationID() string {
	return tdb.organizationID
}

// GetUserID returns the test user ID
func (tdb *TestDataBuilder) GetUserID() string {
	return tdb.userID
}

// GetManagerID returns the test manager ID
func (tdb *TestDataBuilder) GetManagerID() string {
	return tdb.managerID
}

// GetFinanceID returns the test finance user ID
func (tdb *TestDataBuilder) GetFinanceID() string {
	return tdb.financeID
}

// WorkflowTestScenario represents a complete workflow test scenario
type WorkflowTestScenario struct {
	Organization *models.Organization
	Users        struct {
		Requester *models.User
		Manager   *models.User
		Finance   *models.User
	}
	Workflow   *models.Workflow
	Assignment *models.WorkflowAssignment
	Task       *models.WorkflowTask
	Document   *models.Requisition
}

// CreateCompleteWorkflowScenario creates a complete workflow test scenario
func CreateCompleteWorkflowScenario(t *testing.T, db *gorm.DB, entityType string) *WorkflowTestScenario {
	builder := NewTestDataBuilder(db)
	
	scenario := &WorkflowTestScenario{}
	
	// Create organization
	scenario.Organization = builder.CreateOrganization(t)
	
	// Create users
	scenario.Users.Requester, scenario.Users.Manager, scenario.Users.Finance = builder.CreateUsers(t)
	
	// Create workflow
	scenario.Workflow = builder.CreateSingleStageWorkflow(t, entityType)
	
	// Create document (requisition for now)
	scenario.Document = builder.CreateRequisition(t)
	
	// Create workflow assignment
	scenario.Assignment = builder.CreateWorkflowAssignment(t, scenario.Workflow, scenario.Document.ID)
	
	// Create workflow task
	scenario.Task = builder.CreateWorkflowTask(t, scenario.Assignment, 1, "manager")
	
	return scenario
}

// AssertTaskStatus asserts the status of a workflow task
func AssertTaskStatus(t *testing.T, db *gorm.DB, taskID string, expectedStatus string) {
	var task models.WorkflowTask
	err := db.Where("id = ?", taskID).First(&task).Error
	require.NoError(t, err)
	require.Equal(t, expectedStatus, task.Status)
}

// AssertWorkflowStatus asserts the status of a workflow assignment
func AssertWorkflowStatus(t *testing.T, db *gorm.DB, assignmentID string, expectedStatus string) {
	var assignment models.WorkflowAssignment
	err := db.Where("id = ?", assignmentID).First(&assignment).Error
	require.NoError(t, err)
	require.Equal(t, expectedStatus, assignment.Status)
}

// AssertApprovalRecordExists asserts that an approval record exists
func AssertApprovalRecordExists(t *testing.T, db *gorm.DB, taskID string, action string, approverID string) {
	var record models.StageApprovalRecord
	err := db.Where("workflow_task_id = ? AND action = ? AND approver_id = ?", taskID, action, approverID).First(&record).Error
	require.NoError(t, err)
}

// WaitForCondition waits for a condition to be true with timeout
func WaitForCondition(t *testing.T, condition func() bool, timeout time.Duration, message string) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			t.Fatalf("Timeout waiting for condition: %s", message)
		case <-ticker.C:
			if condition() {
				return
			}
		}
	}
}

// StringPtr returns a pointer to a string
func StringPtr(s string) *string {
	return &s
}

// IntPtr returns a pointer to an int
func IntPtr(i int) *int {
	return &i
}

// TimePtr returns a pointer to a time
func TimePtr(t time.Time) *time.Time {
	return &t
}