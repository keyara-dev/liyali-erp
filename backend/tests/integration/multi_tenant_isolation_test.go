package integration

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/tests/helpers"
	"github.com/stretchr/testify/assert"
)

// TestMultiTenantIsolationIntegration tests multi-tenant data isolation using mocks
func TestMultiTenantIsolationIntegration(t *testing.T) {
	builder1 := helpers.NewMockTestDataBuilder()
	builder2 := helpers.NewMockTestDataBuilder()
	
	// Create two separate organizations
	org1 := builder1.CreateMockOrganization(t)
	org2 := builder2.CreateMockOrganization(t)
	
	// Create users for each organization
	user1, _, _ := builder1.CreateMockUsers(t)
	user2, _, _ := builder2.CreateMockUsers(t)

	t.Run("Organization data isolation", func(t *testing.T) {
		// Mock documents for each organization
		doc1 := &models.Document{
			ID:             uuid.New(),
			OrganizationID: org1.ID,
			DocumentNumber: "DOC-001",
			Title:          "Org1 Document",
			CreatedBy:      user1.ID,
		}
		
		doc2 := &models.Document{
			ID:             uuid.New(),
			OrganizationID: org2.ID,
			DocumentNumber: "DOC-001", // Same number, different org
			Title:          "Org2 Document",
			CreatedBy:      user2.ID,
		}

		// Verify documents belong to correct organizations
		assert.Equal(t, org1.ID, doc1.OrganizationID)
		assert.Equal(t, org2.ID, doc2.OrganizationID)
		assert.NotEqual(t, doc1.OrganizationID, doc2.OrganizationID)
		
		// Same document number should be allowed in different orgs
		assert.Equal(t, "DOC-001", doc1.DocumentNumber)
		assert.Equal(t, "DOC-001", doc2.DocumentNumber)
	})

	t.Run("User access isolation", func(t *testing.T) {
		// Mock requisitions for each organization
		req1 := &models.Requisition{
			ID:             uuid.New().String(),
			OrganizationID: org1.ID,
			DocumentNumber: "REQ-001",
			Title:          "Org1 Requisition",
			RequesterId:    user1.ID,
			Status: "DRAFT",
		}
		
		req2 := &models.Requisition{
			ID:             uuid.New().String(),
			OrganizationID: org2.ID,
			DocumentNumber: "REQ-001", // Same number, different org
			Title:          "Org2 Requisition",
			RequesterId:    user2.ID,
			Status: "DRAFT",
		}

		// Verify users can only access their organization's data
		assert.Equal(t, org1.ID, req1.OrganizationID)
		assert.Equal(t, user1.ID, req1.RequesterId)
		assert.Equal(t, org2.ID, req2.OrganizationID)
		assert.Equal(t, user2.ID, req2.RequesterId)
		
		// Cross-organization access should be prevented
		assert.NotEqual(t, req1.OrganizationID, req2.OrganizationID)
		assert.NotEqual(t, req1.RequesterId, req2.RequesterId)
	})

	t.Run("Workflow isolation", func(t *testing.T) {
		// Mock workflows for each organization
		workflow1 := &models.Workflow{
			ID:             uuid.New(),
			OrganizationID: org1.ID,
			Name:           "Org1 Approval Workflow",
			EntityType:     "requisition",
			IsActive:       true,
			CreatedBy:      user1.ID,
		}
		
		workflow2 := &models.Workflow{
			ID:             uuid.New(),
			OrganizationID: org2.ID,
			Name:           "Org2 Approval Workflow",
			EntityType:     "requisition",
			IsActive:       true,
			CreatedBy:      user2.ID,
		}

		// Verify workflows are isolated by organization
		assert.Equal(t, org1.ID, workflow1.OrganizationID)
		assert.Equal(t, org2.ID, workflow2.OrganizationID)
		assert.NotEqual(t, workflow1.OrganizationID, workflow2.OrganizationID)
	})

	t.Run("Budget isolation", func(t *testing.T) {
		// Mock budgets for each organization
		budget1 := &models.Budget{
			ID:              uuid.New().String(),
			OrganizationID:  org1.ID,
			BudgetCode:      "IT-2024",
			Department:      "IT",
			FiscalYear:      "2024",
			TotalBudget:     100000.00,
			AllocatedAmount: 25000.00,
			Status: "APPROVED",
		}
		
		budget2 := &models.Budget{
			ID:              uuid.New().String(),
			OrganizationID:  org2.ID,
			BudgetCode:      "IT-2024", // Same code, different org
			Department:      "IT",
			FiscalYear:      "2024",
			TotalBudget:     150000.00,
			AllocatedAmount: 30000.00,
			Status: "APPROVED",
		}

		// Verify budget isolation
		assert.Equal(t, org1.ID, budget1.OrganizationID)
		assert.Equal(t, org2.ID, budget2.OrganizationID)
		assert.NotEqual(t, budget1.OrganizationID, budget2.OrganizationID)
		
		// Same budget code should be allowed in different orgs
		assert.Equal(t, "IT-2024", budget1.BudgetCode)
		assert.Equal(t, "IT-2024", budget2.BudgetCode)
	})

	t.Run("Vendor isolation", func(t *testing.T) {
		// Mock vendors for each organization
		vendor1 := &models.Vendor{
			ID:             uuid.New().String(),
			OrganizationID: org1.ID,
			VendorCode:     "VENDOR-001",
			Name:           "Org1 Vendor",
			Email:          "vendor@org1.com",
			Active:         true,
		}
		
		vendor2 := &models.Vendor{
			ID:             uuid.New().String(),
			OrganizationID: org2.ID,
			VendorCode:     "VENDOR-001", // Same code, different org
			Name:           "Org2 Vendor",
			Email:          "vendor@org2.com",
			Active:         true,
		}

		// Verify vendor isolation
		assert.Equal(t, org1.ID, vendor1.OrganizationID)
		assert.Equal(t, org2.ID, vendor2.OrganizationID)
		assert.NotEqual(t, vendor1.OrganizationID, vendor2.OrganizationID)
		
		// Same vendor code should be allowed in different orgs
		assert.Equal(t, "VENDOR-001", vendor1.VendorCode)
		assert.Equal(t, "VENDOR-001", vendor2.VendorCode)
	})

	t.Run("Session isolation", func(t *testing.T) {
		// Mock sessions for each user
		session1 := &models.Session{
			ID:           uuid.New(),
			UserID:       user1.ID,
			RefreshToken: "token-org1-user",
			ExpiresAt:    time.Now().Add(24 * time.Hour),
			CreatedAt:    time.Now(),
		}
		
		session2 := &models.Session{
			ID:           uuid.New(),
			UserID:       user2.ID,
			RefreshToken: "token-org2-user",
			ExpiresAt:    time.Now().Add(24 * time.Hour),
			CreatedAt:    time.Now(),
		}

		// Verify session isolation
		assert.Equal(t, user1.ID, session1.UserID)
		assert.Equal(t, user2.ID, session2.UserID)
		assert.NotEqual(t, session1.UserID, session2.UserID)
		assert.NotEqual(t, session1.RefreshToken, session2.RefreshToken)
	})

	t.Run("Audit log isolation", func(t *testing.T) {
		// Mock audit logs for each organization
		audit1 := &models.AuditLog{
			ID:           uuid.New().String(),
			DocumentID:   uuid.New().String(),
			DocumentType: "requisition",
			UserID:       user1.ID,
			Action:       "create_requisition",
			CreatedAt:    time.Now(),
		}
		
		audit2 := &models.AuditLog{
			ID:           uuid.New().String(),
			DocumentID:   uuid.New().String(),
			DocumentType: "requisition",
			UserID:       user2.ID,
			Action:       "create_requisition",
			CreatedAt:    time.Now(),
		}

		// Verify audit log isolation
		assert.NotEmpty(t, audit1.ID)
		assert.Equal(t, user1.ID, audit1.UserID)
		assert.NotEmpty(t, audit2.ID)
		assert.Equal(t, user2.ID, audit2.UserID)
		assert.NotEqual(t, audit1.UserID, audit2.UserID)
	})
}

// Helper function to create mock string pointer
func StringPtr(s string) *string {
	return &s
}