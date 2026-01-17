package integration

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/services"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupMultiTenantTestDB creates an in-memory database for multi-tenant testing
func setupMultiTenantTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Auto-migrate all models
	db.AutoMigrate(
		&models.Organization{},
		&models.OrganizationMember{},
		&models.OrganizationSettings{},
		&models.User{},
		&models.Workflow{},
		&models.WorkflowDefault{},
		&models.WorkflowAssignment{},
		&models.Document{},
		&models.Requisition{},
		&models.Budget{},
		&models.PurchaseOrder{},
		&models.PaymentVoucher{},
		&models.GoodsReceivedNote{},
		&models.Vendor{},
		&models.Session{},
		&models.LoginAttempt{},
		&models.AccountLockout{},
		&models.PasswordReset{},
		&models.AuditLog{},
		&models.Notification{},
	)

	return db
}

// createTestOrganizations creates two separate organizations with users for testing
func createTestOrganizations(db *gorm.DB) (org1, org2 *models.Organization, user1, user2 *models.User) {
	// Create users
	user1 = &models.User{
		ID:     "user-org1-admin",
		Email:  "admin1@org1.com",
		Name:   "Org1 Admin",
		Active: true,
	}
	user2 = &models.User{
		ID:     "user-org2-admin",
		Email:  "admin2@org2.com",
		Name:   "Org2 Admin",
		Active: true,
	}
	db.Create(user1)
	db.Create(user2)

	// Create organizations
	org1 = &models.Organization{
		ID:          "org-1",
		Name:        "Organization One",
		Slug:        "org-one",
		Description: "First test organization",
		Active:      true,
		Tier:        "starter",
		CreatedBy:   user1.ID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	org2 = &models.Organization{
		ID:          "org-2",
		Name:        "Organization Two",
		Slug:        "org-two",
		Description: "Second test organization",
		Active:      true,
		Tier:        "starter",
		CreatedBy:   user2.ID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	db.Create(org1)
	db.Create(org2)

	// Create organization memberships
	member1 := &models.OrganizationMember{
		ID:             "member-1",
		OrganizationID: org1.ID,
		UserID:         user1.ID,
		Role:           "admin",
		Active:         true,
		JoinedAt:       &[]time.Time{time.Now()}[0],
	}
	member2 := &models.OrganizationMember{
		ID:             "member-2",
		OrganizationID: org2.ID,
		UserID:         user2.ID,
		Role:           "admin",
		Active:         true,
		JoinedAt:       &[]time.Time{time.Now()}[0],
	}
	db.Create(member1)
	db.Create(member2)

	// Set current organizations
	user1.CurrentOrganizationID = &org1.ID
	user2.CurrentOrganizationID = &org2.ID
	db.Save(user1)
	db.Save(user2)

	return org1, org2, user1, user2
}

func TestMultiTenantWorkflowIsolation(t *testing.T) {
	db := setupMultiTenantTestDB()
	org1, org2, user1, user2 := createTestOrganizations(db)

	workflowService := services.NewWorkflowService(nil, nil, db) // Simplified for testing

	t.Run("Workflows are isolated between organizations", func(t *testing.T) {
		// Create workflow in org1
		workflow1Request := services.CreateWorkflowRequest{
			Name:        "Org1 Workflow",
			Description: "Workflow for organization 1",
			EntityType:  "REQUISITION",
			Stages: []models.WorkflowStage{
				{
					StageNumber:  1,
					StageName:    "Manager Approval",
					RequiredRole: "manager",
				},
			},
		}

		workflow1, err := workflowService.CreateWorkflow(context.Background(), org1.ID, user1.ID, workflow1Request)
		assert.NoError(t, err)
		assert.NotNil(t, workflow1)

		// Create workflow in org2
		workflow2Request := services.CreateWorkflowRequest{
			Name:        "Org2 Workflow",
			Description: "Workflow for organization 2",
			EntityType:  "REQUISITION",
			Stages: []models.WorkflowStage{
				{
					StageNumber:  1,
					StageName:    "Director Approval",
					RequiredRole: "director",
				},
			},
		}

		workflow2, err := workflowService.CreateWorkflow(context.Background(), org2.ID, user2.ID, workflow2Request)
		assert.NoError(t, err)
		assert.NotNil(t, workflow2)

		// Verify org1 can only see its workflow
		org1Workflows, err := workflowService.GetWorkflows(context.Background(), org1.ID, services.WorkflowListFilter{})
		assert.NoError(t, err)
		assert.Len(t, org1Workflows, 1)
		assert.Equal(t, "Org1 Workflow", org1Workflows[0].Name)

		// Verify org2 can only see its workflow
		org2Workflows, err := workflowService.GetWorkflows(context.Background(), org2.ID, services.WorkflowListFilter{})
		assert.NoError(t, err)
		assert.Len(t, org2Workflows, 1)
		assert.Equal(t, "Org2 Workflow", org2Workflows[0].Name)

		// Verify cross-organization access is prevented
		_, err = workflowService.GetWorkflow(context.Background(), workflow1.ID, org2.ID)
		assert.Error(t, err)

		_, err = workflowService.GetWorkflow(context.Background(), workflow2.ID, org1.ID)
		assert.Error(t, err)
	})

	t.Run("Default workflows are organization-specific", func(t *testing.T) {
		// Get default workflow for org1
		defaultWorkflow1, err := workflowService.GetDefaultWorkflow(context.Background(), org1.ID, "REQUISITION")
		assert.NoError(t, err)
		assert.Equal(t, org1.ID, defaultWorkflow1.OrganizationID)

		// Get default workflow for org2
		defaultWorkflow2, err := workflowService.GetDefaultWorkflow(context.Background(), org2.ID, "REQUISITION")
		assert.NoError(t, err)
		assert.Equal(t, org2.ID, defaultWorkflow2.OrganizationID)

		// Verify they are different workflows
		assert.NotEqual(t, defaultWorkflow1.ID, defaultWorkflow2.ID)
	})
}

func TestMultiTenantDocumentIsolation(t *testing.T) {
	db := setupMultiTenantTestDB()
	org1, org2, user1, user2 := createTestOrganizations(db)

	// Create test documents in each organization
	doc1 := &models.Document{
		ID:             uuid.New(),
		OrganizationID: org1.ID,
		DocumentType:   "REQUISITION",
		Title:          "Org1 Document",
		Status:         "draft",
		CreatedBy:      user1.ID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	doc2 := &models.Document{
		ID:             uuid.New(),
		OrganizationID: org2.ID,
		DocumentType:   "REQUISITION",
		Title:          "Org2 Document",
		Status:         "draft",
		CreatedBy:      user2.ID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	db.Create(doc1)
	db.Create(doc2)

	t.Run("Documents are isolated between organizations", func(t *testing.T) {
		// Query documents for org1
		var org1Docs []models.Document
		err := db.Where("organization_id = ?", org1.ID).Find(&org1Docs).Error
		assert.NoError(t, err)
		assert.Len(t, org1Docs, 1)
		assert.Equal(t, "Org1 Document", org1Docs[0].Title)

		// Query documents for org2
		var org2Docs []models.Document
		err = db.Where("organization_id = ?", org2.ID).Find(&org2Docs).Error
		assert.NoError(t, err)
		assert.Len(t, org2Docs, 1)
		assert.Equal(t, "Org2 Document", org2Docs[0].Title)

		// Verify cross-organization queries return empty
		var crossOrgDocs []models.Document
		err = db.Where("organization_id = ? AND id = ?", org1.ID, doc2.ID).Find(&crossOrgDocs).Error
		assert.NoError(t, err)
		assert.Len(t, crossOrgDocs, 0)

		err = db.Where("organization_id = ? AND id = ?", org2.ID, doc1.ID).Find(&crossOrgDocs).Error
		assert.NoError(t, err)
		assert.Len(t, crossOrgDocs, 0)
	})
}

func TestMultiTenantUserIsolation(t *testing.T) {
	db := setupMultiTenantTestDB()
	org1, org2, user1, user2 := createTestOrganizations(db)

	// Create additional users for each organization
	org1User := &models.User{
		ID:                      "org1-user",
		Email:                   "user@org1.com",
		Name:                    "Org1 User",
		Active:                  true,
		CurrentOrganizationID:   &org1.ID,
	}
	org2User := &models.User{
		ID:                      "org2-user",
		Email:                   "user@org2.com",
		Name:                    "Org2 User",
		Active:                  true,
		CurrentOrganizationID:   &org2.ID,
	}
	db.Create(org1User)
	db.Create(org2User)

	// Add users to their respective organizations
	org1Member := &models.OrganizationMember{
		ID:             "org1-member",
		OrganizationID: org1.ID,
		UserID:         org1User.ID,
		Role:           "requester",
		Active:         true,
		JoinedAt:       &[]time.Time{time.Now()}[0],
	}
	org2Member := &models.OrganizationMember{
		ID:             "org2-member",
		OrganizationID: org2.ID,
		UserID:         org2User.ID,
		Role:           "requester",
		Active:         true,
		JoinedAt:       &[]time.Time{time.Now()}[0],
	}
	db.Create(org1Member)
	db.Create(org2Member)

	t.Run("Organization members are isolated", func(t *testing.T) {
		// Get org1 members
		var org1Members []models.OrganizationMember
		err := db.Where("organization_id = ? AND active = ?", org1.ID, true).Find(&org1Members).Error
		assert.NoError(t, err)
		assert.Len(t, org1Members, 2) // Admin + user

		// Get org2 members
		var org2Members []models.OrganizationMember
		err = db.Where("organization_id = ? AND active = ?", org2.ID, true).Find(&org2Members).Error
		assert.NoError(t, err)
		assert.Len(t, org2Members, 2) // Admin + user

		// Verify no cross-organization membership
		var crossMembers []models.OrganizationMember
		err = db.Where("organization_id = ? AND user_id = ?", org1.ID, org2User.ID).Find(&crossMembers).Error
		assert.NoError(t, err)
		assert.Len(t, crossMembers, 0)

		err = db.Where("organization_id = ? AND user_id = ?", org2.ID, org1User.ID).Find(&crossMembers).Error
		assert.NoError(t, err)
		assert.Len(t, crossMembers, 0)
	})
}

func TestMultiTenantAuditLogIsolation(t *testing.T) {
	db := setupMultiTenantTestDB()
	org1, org2, user1, user2 := createTestOrganizations(db)

	// Create audit logs for each organization
	auditLog1 := &models.AuditLog{
		ID:             uuid.New().String(),
		OrganizationID: org1.ID,
		UserID:         user1.ID,
		Action:         "document_created",
		ResourceType:   "document",
		ResourceID:     "doc-1",
		Details:        "Created test document",
		IPAddress:      "192.168.1.1",
		UserAgent:      "test-agent",
		CreatedAt:      time.Now(),
	}

	auditLog2 := &models.AuditLog{
		ID:             uuid.New().String(),
		OrganizationID: org2.ID,
		UserID:         user2.ID,
		Action:         "workflow_created",
		ResourceType:   "workflow",
		ResourceID:     "workflow-1",
		Details:        "Created test workflow",
		IPAddress:      "192.168.1.2",
		UserAgent:      "test-agent",
		CreatedAt:      time.Now(),
	}

	db.Create(auditLog1)
	db.Create(auditLog2)

	t.Run("Audit logs are isolated between organizations", func(t *testing.T) {
		// Query audit logs for org1
		var org1Logs []models.AuditLog
		err := db.Where("organization_id = ?", org1.ID).Find(&org1Logs).Error
		assert.NoError(t, err)
		assert.Len(t, org1Logs, 1)
		assert.Equal(t, "document_created", org1Logs[0].Action)

		// Query audit logs for org2
		var org2Logs []models.AuditLog
		err = db.Where("organization_id = ?", org2.ID).Find(&org2Logs).Error
		assert.NoError(t, err)
		assert.Len(t, org2Logs, 1)
		assert.Equal(t, "workflow_created", org2Logs[0].Action)

		// Verify no cross-organization access
		var crossLogs []models.AuditLog
		err = db.Where("organization_id = ? AND id = ?", org1.ID, auditLog2.ID).Find(&crossLogs).Error
		assert.NoError(t, err)
		assert.Len(t, crossLogs, 0)
	})
}

func TestMultiTenantSessionIsolation(t *testing.T) {
	db := setupMultiTenantTestDB()
	org1, org2, user1, user2 := createTestOrganizations(db)

	// Create sessions for users from different organizations
	session1 := &models.Session{
		ID:           uuid.New(),
		UserID:       user1.ID,
		RefreshToken: "refresh-token-org1-user",
		IPAddress:    "192.168.1.1",
		UserAgent:    "test-agent-1",
		ExpiresAt:    time.Now().Add(24 * time.Hour),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	session2 := &models.Session{
		ID:           uuid.New(),
		UserID:       user2.ID,
		RefreshToken: "refresh-token-org2-user",
		IPAddress:    "192.168.1.2",
		UserAgent:    "test-agent-2",
		ExpiresAt:    time.Now().Add(24 * time.Hour),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	db.Create(session1)
	db.Create(session2)

	t.Run("Sessions are properly isolated by user", func(t *testing.T) {
		// Query sessions for user1
		var user1Sessions []models.Session
		err := db.Where("user_id = ?", user1.ID).Find(&user1Sessions).Error
		assert.NoError(t, err)
		assert.Len(t, user1Sessions, 1)
		assert.Equal(t, "refresh-token-org1-user", user1Sessions[0].RefreshToken)

		// Query sessions for user2
		var user2Sessions []models.Session
		err = db.Where("user_id = ?", user2.ID).Find(&user2Sessions).Error
		assert.NoError(t, err)
		assert.Len(t, user2Sessions, 1)
		assert.Equal(t, "refresh-token-org2-user", user2Sessions[0].RefreshToken)

		// Verify no cross-user session access
		var crossSessions []models.Session
		err = db.Where("user_id = ? AND refresh_token = ?", user1.ID, "refresh-token-org2-user").Find(&crossSessions).Error
		assert.NoError(t, err)
		assert.Len(t, crossSessions, 0)
	})
}

func TestMultiTenantDataLeakagePrevention(t *testing.T) {
	db := setupMultiTenantTestDB()
	org1, org2, user1, user2 := createTestOrganizations(db)

	// Create sensitive data in each organization
	sensitiveDoc1 := &models.Document{
		ID:             uuid.New(),
		OrganizationID: org1.ID,
		DocumentType:   "BUDGET",
		Title:          "Confidential Budget Org1",
		Status:         "approved",
		CreatedBy:      user1.ID,
		Amount:         &[]float64{1000000.0}[0], // Sensitive financial data
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	sensitiveDoc2 := &models.Document{
		ID:             uuid.New(),
		OrganizationID: org2.ID,
		DocumentType:   "BUDGET",
		Title:          "Confidential Budget Org2",
		Status:         "approved",
		CreatedBy:      user2.ID,
		Amount:         &[]float64{2000000.0}[0], // Different sensitive financial data
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	db.Create(sensitiveDoc1)
	db.Create(sensitiveDoc2)

	t.Run("Prevent data leakage through SQL injection-like queries", func(t *testing.T) {
		// Simulate malicious query attempts
		maliciousOrgID := org1.ID + "' OR '1'='1"

		var leakedDocs []models.Document
		err := db.Where("organization_id = ?", maliciousOrgID).Find(&leakedDocs).Error
		assert.NoError(t, err)
		assert.Len(t, leakedDocs, 0) // Should not return any documents

		// Test with UNION injection attempt
		maliciousOrgID2 := org1.ID + "' UNION SELECT * FROM documents WHERE organization_id = '" + org2.ID + "' --"
		err = db.Where("organization_id = ?", maliciousOrgID2).Find(&leakedDocs).Error
		assert.NoError(t, err)
		assert.Len(t, leakedDocs, 0) // Should not return any documents
	})

	t.Run("Verify aggregate queries respect tenant boundaries", func(t *testing.T) {
		// Count documents for org1
		var org1Count int64
		err := db.Model(&models.Document{}).Where("organization_id = ?", org1.ID).Count(&org1Count).Error
		assert.NoError(t, err)
		assert.Equal(t, int64(1), org1Count)

		// Count documents for org2
		var org2Count int64
		err = db.Model(&models.Document{}).Where("organization_id = ?", org2.ID).Count(&org2Count).Error
		assert.NoError(t, err)
		assert.Equal(t, int64(1), org2Count)

		// Sum amounts for org1 (should only include org1 data)
		var org1Sum float64
		err = db.Model(&models.Document{}).Where("organization_id = ?", org1.ID).Select("COALESCE(SUM(amount), 0)").Scan(&org1Sum).Error
		assert.NoError(t, err)
		assert.Equal(t, 1000000.0, org1Sum)

		// Sum amounts for org2 (should only include org2 data)
		var org2Sum float64
		err = db.Model(&models.Document{}).Where("organization_id = ?", org2.ID).Select("COALESCE(SUM(amount), 0)").Scan(&org2Sum).Error
		assert.NoError(t, err)
		assert.Equal(t, 2000000.0, org2Sum)
	})

	t.Run("Test concurrent access isolation", func(t *testing.T) {
		done := make(chan bool, 2)
		results := make(chan []models.Document, 2)

		// Simulate concurrent access from different organizations
		go func() {
			var docs []models.Document
			db.Where("organization_id = ?", org1.ID).Find(&docs)
			results <- docs
			done <- true
		}()

		go func() {
			var docs []models.Document
			db.Where("organization_id = ?", org2.ID).Find(&docs)
			results <- docs
			done <- true
		}()

		// Wait for both queries to complete
		<-done
		<-done

		// Verify results
		result1 := <-results
		result2 := <-results

		// Each should only see their own organization's data
		assert.Len(t, result1, 1)
		assert.Len(t, result2, 1)

		// Verify the documents belong to the correct organizations
		if result1[0].OrganizationID == org1.ID {
			assert.Equal(t, org1.ID, result1[0].OrganizationID)
			assert.Equal(t, org2.ID, result2[0].OrganizationID)
		} else {
			assert.Equal(t, org2.ID, result1[0].OrganizationID)
			assert.Equal(t, org1.ID, result2[0].OrganizationID)
		}
	})
}

func TestMultiTenantPerformanceIsolation(t *testing.T) {
	db := setupMultiTenantTestDB()
	org1, org2, user1, user2 := createTestOrganizations(db)

	// Create a large number of documents for org1 to simulate heavy load
	for i := 0; i < 1000; i++ {
		doc := &models.Document{
			ID:             uuid.New(),
			OrganizationID: org1.ID,
			DocumentType:   "REQUISITION",
			Title:          fmt.Sprintf("Bulk Document %d", i),
			Status:         "draft",
			CreatedBy:      user1.ID,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}
		db.Create(doc)
	}

	// Create a few documents for org2
	for i := 0; i < 5; i++ {
		doc := &models.Document{
			ID:             uuid.New(),
			OrganizationID: org2.ID,
			DocumentType:   "REQUISITION",
			Title:          fmt.Sprintf("Small Document %d", i),
			Status:         "draft",
			CreatedBy:      user2.ID,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}
		db.Create(doc)
	}

	t.Run("Heavy load in one tenant doesn't affect other tenant queries", func(t *testing.T) {
		// Measure query time for org2 (should be fast despite org1's large dataset)
		start := time.Now()
		var org2Docs []models.Document
		err := db.Where("organization_id = ?", org2.ID).Find(&org2Docs).Error
		queryTime := time.Since(start)

		assert.NoError(t, err)
		assert.Len(t, org2Docs, 5)
		
		// Query should complete quickly (less than 100ms for this small dataset)
		assert.Less(t, queryTime, 100*time.Millisecond, "Query took too long, possible performance isolation issue")

		// Verify org2 only sees its own documents
		for _, doc := range org2Docs {
			assert.Equal(t, org2.ID, doc.OrganizationID)
			assert.Contains(t, doc.Title, "Small Document")
		}
	})
}