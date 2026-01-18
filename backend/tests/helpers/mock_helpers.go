package helpers

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/models"
	"github.com/stretchr/testify/mock"
)

// MockDB represents a mock database for testing without CGO
type MockDB struct {
	mock.Mock
}

// MockRepository represents a mock repository interface
type MockRepository struct {
	mock.Mock
}

// MockTestDatabase represents a mock test database instance
type MockTestDatabase struct {
	MockDB *MockDB
}

// SetupMockTestDB creates a mock database for testing without CGO dependencies
func SetupMockTestDB(t *testing.T) *MockTestDatabase {
	mockDB := &MockDB{}
	return &MockTestDatabase{MockDB: mockDB}
}

// Cleanup is a no-op for mock database
func (mtd *MockTestDatabase) Cleanup() {
	// No cleanup needed for mock
}

// MockDataBuilder helps create mock test data
type MockDataBuilder struct {
	organizationID string
	userID         string
	managerID      string
	financeID      string
}

// NewMockTestDataBuilder creates a new mock test data builder
func NewMockTestDataBuilder() *MockDataBuilder {
	return &MockDataBuilder{
		organizationID: "test-org-" + uuid.New().String()[:8],
		userID:         "test-user-" + uuid.New().String()[:8],
		managerID:      "test-manager-" + uuid.New().String()[:8],
		financeID:      "test-finance-" + uuid.New().String()[:8],
	}
}

// CreateMockOrganization creates a mock organization
func (mdb *MockDataBuilder) CreateMockOrganization(t *testing.T) *models.Organization {
	return &models.Organization{
		ID:   mdb.organizationID,
		Name: "Test Organization",
		Slug: "test-org",
	}
}

// CreateMockUsers creates mock users with different roles
func (mdb *MockDataBuilder) CreateMockUsers(t *testing.T) (*models.User, *models.User, *models.User) {
	// Regular user
	user := &models.User{
		ID:                    mdb.userID,
		Email:                 "user@example.com",
		Name:                  "Test User",
		Role:                  "requester",
		CurrentOrganizationID: &mdb.organizationID,
		Active:                true,
	}

	// Manager
	manager := &models.User{
		ID:                    mdb.managerID,
		Email:                 "manager@example.com",
		Name:                  "Test Manager",
		Role:                  "manager",
		CurrentOrganizationID: &mdb.organizationID,
		Active:                true,
	}

	// Finance user
	finance := &models.User{
		ID:                    mdb.financeID,
		Email:                 "finance@example.com",
		Name:                  "Test Finance",
		Role:                  "finance",
		CurrentOrganizationID: &mdb.organizationID,
		Active:                true,
	}

	return user, manager, finance
}

// CreateMockWorkflow creates a mock workflow
func (mdb *MockDataBuilder) CreateMockWorkflow(t *testing.T, entityType string, stages []models.WorkflowStage) *models.Workflow {
	workflow := &models.Workflow{
		ID:             uuid.New(),
		OrganizationID: mdb.organizationID,
		Name:           "Test " + entityType + " Workflow",
		EntityType:     entityType,
		IsActive:       true,
		IsDefault:      true,
		CreatedBy:      mdb.userID,
	}

	workflow.SetStages(stages)
	return workflow
}

// CreateMockSingleStageWorkflow creates a mock single-stage workflow
func (mdb *MockDataBuilder) CreateMockSingleStageWorkflow(t *testing.T, entityType string) *models.Workflow {
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
	return mdb.CreateMockWorkflow(t, entityType, stages)
}

// CreateMockRequisition creates a mock requisition
func (mdb *MockDataBuilder) CreateMockRequisition(t *testing.T) *models.Requisition {
	return &models.Requisition{
		ID:             uuid.New().String(),
		OrganizationID: mdb.organizationID,
		DocumentNumber: "REQ-" + uuid.New().String()[:8],
		Title:          "Test Requisition",
		Description:    "Test requisition for workflow testing",
		Status:         "draft",
		RequesterId:    mdb.userID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}

// GetOrganizationID returns the mock organization ID
func (mdb *MockDataBuilder) GetOrganizationID() string {
	return mdb.organizationID
}

// GetUserID returns the mock user ID
func (mdb *MockDataBuilder) GetUserID() string {
	return mdb.userID
}

// GetManagerID returns the mock manager ID
func (mdb *MockDataBuilder) GetManagerID() string {
	return mdb.managerID
}

// GetFinanceID returns the mock finance user ID
func (mdb *MockDataBuilder) GetFinanceID() string {
	return mdb.financeID
}

// MockWorkflowTestScenario represents a complete mock workflow test scenario
type MockWorkflowTestScenario struct {
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

// CreateMockCompleteWorkflowScenario creates a complete mock workflow test scenario
func CreateMockCompleteWorkflowScenario(t *testing.T, entityType string) *MockWorkflowTestScenario {
	builder := NewMockTestDataBuilder()
	
	scenario := &MockWorkflowTestScenario{}
	
	// Create mock organization
	scenario.Organization = builder.CreateMockOrganization(t)
	
	// Create mock users
	scenario.Users.Requester, scenario.Users.Manager, scenario.Users.Finance = builder.CreateMockUsers(t)
	
	// Create mock workflow
	scenario.Workflow = builder.CreateMockSingleStageWorkflow(t, entityType)
	
	// Create mock document (requisition for now)
	scenario.Document = builder.CreateMockRequisition(t)
	
	// Create mock workflow assignment
	scenario.Assignment = &models.WorkflowAssignment{
		ID:              uuid.New().String(),
		OrganizationID:  builder.GetOrganizationID(),
		EntityID:        scenario.Document.ID,
		EntityType:      scenario.Workflow.EntityType,
		WorkflowID:      scenario.Workflow.ID,
		WorkflowVersion: scenario.Workflow.Version,
		CurrentStage:    1,
		Status:          "in_progress",
		AssignedBy:      builder.GetUserID(),
		AssignedAt:      time.Now(),
	}
	
	// Create mock workflow task
	scenario.Task = &models.WorkflowTask{
		ID:                   uuid.New().String(),
		OrganizationID:       scenario.Assignment.OrganizationID,
		WorkflowAssignmentID: scenario.Assignment.ID,
		EntityID:             scenario.Assignment.EntityID,
		EntityType:           scenario.Assignment.EntityType,
		StageNumber:          1,
		StageName:            "Manager Approval",
		AssignmentType:       "role",
		AssignedRole:         StringPtr("manager"),
		Status:               "pending",
		Priority:             "medium",
		Version:              1,
		CreatedAt:            time.Now(),
	}
	
	return scenario
}

// CheckDatabaseAvailable always returns false to force mock usage
func CheckDatabaseAvailable(t *testing.T) bool {
	return false
}

// SkipIfNoDatabaseAvailable does nothing - we always use mocks
func SkipIfNoDatabaseAvailable(t *testing.T) {
	// Always use mocks, never skip
}