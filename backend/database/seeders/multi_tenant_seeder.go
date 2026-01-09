package seeders

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/types"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// SeedMultiTenantData creates properly separated data for multiple organizations
func SeedMultiTenantData(db *gorm.DB) error {
	log.Println("🌱 Seeding multi-tenant data with proper workspace separation...")

	// First, ensure we have organizations
	if err := ensureOrganizations(db); err != nil {
		return fmt.Errorf("failed to ensure organizations: %w", err)
	}

	// Create users for each organization
	if err := seedOrganizationUsers(db); err != nil {
		return fmt.Errorf("failed to seed organization users: %w", err)
	}

	// Create organization memberships
	if err := seedOrganizationMemberships(db); err != nil {
		return fmt.Errorf("failed to seed organization memberships: %w", err)
	}

	// Create organization-specific categories
	if err := seedOrganizationCategories(db); err != nil {
		return fmt.Errorf("failed to seed organization categories: %w", err)
	}

	// Create organization-specific workflows
	if err := seedOrganizationWorkflows(db); err != nil {
		return fmt.Errorf("failed to seed organization workflows: %w", err)
	}

	// Create organization-specific sample documents
	if err := seedOrganizationDocuments(db); err != nil {
		return fmt.Errorf("failed to seed organization documents: %w", err)
	}

	log.Println("🎉 Multi-tenant data seeded successfully with proper workspace separation!")
	return nil
}

// ensureOrganizations creates the two test organizations if they don't exist
func ensureOrganizations(db *gorm.DB) error {
	organizations := []models.Organization{
		{
			ID:          "org-demo-001",
			Name:        "Demo Organization",
			Slug:        "demo-org",
			Description: "Demo organization for testing and development",
			Active:      true,
			Tier:        "pro",
			CreatedBy:   "user-admin-001",
		},
		{
			ID:          "org-acme-001",
			Name:        "ACME Corporation",
			Slug:        "acme-corp",
			Description: "ACME Corporation for procurement testing",
			Active:      true,
			Tier:        "enterprise",
			CreatedBy:   "user-admin-001",
		},
	}

	for _, org := range organizations {
		var existing models.Organization
		if err := db.Where("id = ?", org.ID).First(&existing).Error; err != nil {
			// Organization doesn't exist, create it
			if err := db.Create(&org).Error; err != nil {
				return fmt.Errorf("failed to create organization %s: %w", org.Name, err)
			}
			log.Printf("✅ Created organization: %s (%s)", org.Name, org.ID)
		} else {
			log.Printf("📋 Organization already exists: %s (%s)", org.Name, org.ID)
		}
	}

	return nil
}

// seedOrganizationUsers creates users for each organization
func seedOrganizationUsers(db *gorm.DB) error {
	// First, ensure the super admin user exists and is updated
	superAdmin := models.User{
		ID:                    "user-admin-001",
		Email:                 "admin@liyali.com",
		Name:                  "System Administrator",
		Password:              "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi", // password
		Role:                  "admin",
		Active:                true,
		IsSuperAdmin:          true,
		CurrentOrganizationID: stringPtr("org-demo-001"), // Default to demo org
	}

	var existingSuperAdmin models.User
	if err := db.Where("id = ? OR email = ?", superAdmin.ID, superAdmin.Email).First(&existingSuperAdmin).Error; err != nil {
		// Super admin doesn't exist, create it
		if err := db.Create(&superAdmin).Error; err != nil {
			return fmt.Errorf("failed to create super admin %s: %w", superAdmin.Email, err)
		}
		log.Printf("✅ Created super admin: %s (%s)", superAdmin.Name, superAdmin.Email)
	} else {
		// Update existing super admin
		if err := db.Model(&existingSuperAdmin).Updates(map[string]interface{}{
			"name":                      superAdmin.Name,
			"role":                      superAdmin.Role,
			"active":                    superAdmin.Active,
			"is_super_admin":           superAdmin.IsSuperAdmin,
			"current_organization_id":   superAdmin.CurrentOrganizationID,
			"updated_at":               time.Now(),
		}).Error; err != nil {
			return fmt.Errorf("failed to update super admin %s: %w", superAdmin.Email, err)
		}
		log.Printf("📋 Updated super admin: %s (%s)", superAdmin.Name, superAdmin.Email)
	}

	// Demo Organization Users
	demoUsers := []models.User{
		{
			ID:                    "user-demo-admin-001",
			Email:                 "admin@demo-org.com",
			Name:                  "Demo Admin",
			Password:              "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi", // password
			Role:                  "admin",
			Active:                true,
			CurrentOrganizationID: stringPtr("org-demo-001"),
		},
		{
			ID:                    "user-demo-manager-001",
			Email:                 "manager@demo-org.com",
			Name:                  "Demo Manager",
			Password:              "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi",
			Role:                  "manager",
			Active:                true,
			CurrentOrganizationID: stringPtr("org-demo-001"),
		},
		{
			ID:                    "user-demo-requester-001",
			Email:                 "requester@demo-org.com",
			Name:                  "Demo Requester",
			Password:              "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi",
			Role:                  "requester",
			Active:                true,
			CurrentOrganizationID: stringPtr("org-demo-001"),
		},
		{
			ID:                    "user-demo-finance-001",
			Email:                 "finance@demo-org.com",
			Name:                  "Demo Finance Officer",
			Password:              "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi",
			Role:                  "finance_manager",
			Active:                true,
			CurrentOrganizationID: stringPtr("org-demo-001"),
		},
	}

	// ACME Corporation Users
	acmeUsers := []models.User{
		{
			ID:                    "user-acme-admin-001",
			Email:                 "admin@acme-corp.com",
			Name:                  "ACME Admin",
			Password:              "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi",
			Role:                  "admin",
			Active:                true,
			CurrentOrganizationID: stringPtr("org-acme-001"),
		},
		{
			ID:                    "user-acme-manager-001",
			Email:                 "manager@acme-corp.com",
			Name:                  "ACME Manager",
			Password:              "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi",
			Role:                  "manager",
			Active:                true,
			CurrentOrganizationID: stringPtr("org-acme-001"),
		},
		{
			ID:                    "user-acme-requester-001",
			Email:                 "requester@acme-corp.com",
			Name:                  "ACME Requester",
			Password:              "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi",
			Role:                  "requester",
			Active:                true,
			CurrentOrganizationID: stringPtr("org-acme-001"),
		},
		{
			ID:                    "user-acme-finance-001",
			Email:                 "finance@acme-corp.com",
			Name:                  "ACME Finance Officer",
			Password:              "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi",
			Role:                  "finance_manager",
			Active:                true,
			CurrentOrganizationID: stringPtr("org-acme-001"),
		},
	}

	allUsers := append(demoUsers, acmeUsers...)

	for _, user := range allUsers {
		var existing models.User
		if err := db.Where("id = ? OR email = ?", user.ID, user.Email).First(&existing).Error; err != nil {
			// User doesn't exist, create it
			if err := db.Create(&user).Error; err != nil {
				return fmt.Errorf("failed to create user %s: %w", user.Email, err)
			}
			log.Printf("✅ Created user: %s (%s)", user.Name, user.Email)
		} else {
			// Update existing user
			if err := db.Model(&existing).Updates(map[string]interface{}{
				"name":                      user.Name,
				"role":                      user.Role,
				"active":                    user.Active,
				"current_organization_id":   user.CurrentOrganizationID,
				"updated_at":               time.Now(),
			}).Error; err != nil {
				return fmt.Errorf("failed to update user %s: %w", user.Email, err)
			}
			log.Printf("📋 Updated user: %s (%s)", user.Name, user.Email)
		}
	}

	return nil
}

// seedOrganizationMemberships creates organization memberships
func seedOrganizationMemberships(db *gorm.DB) error {
	now := time.Now()
	
	memberships := []models.OrganizationMember{
		// Super Admin memberships - admin@liyali.com should be in both organizations
		{
			ID:             "member-super-admin-demo",
			OrganizationID: "org-demo-001",
			UserID:         "user-admin-001",
			Role:           "admin",
			Active:         true,
			JoinedAt:       &now,
		},
		{
			ID:             "member-super-admin-acme",
			OrganizationID: "org-acme-001",
			UserID:         "user-admin-001",
			Role:           "admin",
			Active:         true,
			JoinedAt:       &now,
		},
		// Demo Organization Members
		{
			ID:             "member-demo-admin-001",
			OrganizationID: "org-demo-001",
			UserID:         "user-demo-admin-001",
			Role:           "admin",
			Active:         true,
			JoinedAt:       &now,
		},
		{
			ID:             "member-demo-manager-001",
			OrganizationID: "org-demo-001",
			UserID:         "user-demo-manager-001",
			Role:           "manager",
			Active:         true,
			JoinedAt:       &now,
		},
		{
			ID:             "member-demo-requester-001",
			OrganizationID: "org-demo-001",
			UserID:         "user-demo-requester-001",
			Role:           "requester",
			Active:         true,
			JoinedAt:       &now,
		},
		{
			ID:             "member-demo-finance-001",
			OrganizationID: "org-demo-001",
			UserID:         "user-demo-finance-001",
			Role:           "finance_manager",
			Active:         true,
			JoinedAt:       &now,
		},
		// ACME Corporation Members
		{
			ID:             "member-acme-admin-001",
			OrganizationID: "org-acme-001",
			UserID:         "user-acme-admin-001",
			Role:           "admin",
			Active:         true,
			JoinedAt:       &now,
		},
		{
			ID:             "member-acme-manager-001",
			OrganizationID: "org-acme-001",
			UserID:         "user-acme-manager-001",
			Role:           "manager",
			Active:         true,
			JoinedAt:       &now,
		},
		{
			ID:             "member-acme-requester-001",
			OrganizationID: "org-acme-001",
			UserID:         "user-acme-requester-001",
			Role:           "requester",
			Active:         true,
			JoinedAt:       &now,
		},
		{
			ID:             "member-acme-finance-001",
			OrganizationID: "org-acme-001",
			UserID:         "user-acme-finance-001",
			Role:           "finance_manager",
			Active:         true,
			JoinedAt:       &now,
		},
	}

	for _, member := range memberships {
		var existing models.OrganizationMember
		if err := db.Where("organization_id = ? AND user_id = ?", member.OrganizationID, member.UserID).First(&existing).Error; err != nil {
			// Membership doesn't exist, create it
			if err := db.Create(&member).Error; err != nil {
				return fmt.Errorf("failed to create membership %s-%s: %w", member.OrganizationID, member.UserID, err)
			}
			log.Printf("✅ Created membership: %s -> %s (%s)", member.UserID, member.OrganizationID, member.Role)
		} else {
			// Update existing membership
			if err := db.Model(&existing).Updates(map[string]interface{}{
				"role":       member.Role,
				"active":     member.Active,
				"updated_at": time.Now(),
			}).Error; err != nil {
				return fmt.Errorf("failed to update membership %s-%s: %w", member.OrganizationID, member.UserID, err)
			}
			log.Printf("📋 Updated membership: %s -> %s (%s)", member.UserID, member.OrganizationID, member.Role)
		}
	}

	return nil
}

// seedOrganizationCategories creates categories for each organization
func seedOrganizationCategories(db *gorm.DB) error {
	// Demo Organization Categories
	demoCategories := []models.Category{
		{
			ID:             "cat-demo-001",
			OrganizationID: "org-demo-001",
			Name:           "Office Supplies",
			Description:    "General office supplies and stationery for Demo Org",
			Active:         true,
		},
		{
			ID:             "cat-demo-002",
			OrganizationID: "org-demo-001",
			Name:           "IT Equipment",
			Description:    "Computers, software, and IT hardware for Demo Org",
			Active:         true,
		},
		{
			ID:             "cat-demo-003",
			OrganizationID: "org-demo-001",
			Name:           "Marketing Materials",
			Description:    "Marketing and promotional materials for Demo Org",
			Active:         true,
		},
	}

	// ACME Corporation Categories
	acmeCategories := []models.Category{
		{
			ID:             "cat-acme-001",
			OrganizationID: "org-acme-001",
			Name:           "Manufacturing Equipment",
			Description:    "Industrial equipment and machinery for ACME Corp",
			Active:         true,
		},
		{
			ID:             "cat-acme-002",
			OrganizationID: "org-acme-001",
			Name:           "Raw Materials",
			Description:    "Raw materials and components for ACME Corp production",
			Active:         true,
		},
		{
			ID:             "cat-acme-003",
			OrganizationID: "org-acme-001",
			Name:           "Safety Equipment",
			Description:    "Safety equipment and protective gear for ACME Corp",
			Active:         true,
		},
	}

	allCategories := append(demoCategories, acmeCategories...)

	for _, category := range allCategories {
		var existing models.Category
		if err := db.Where("id = ?", category.ID).First(&existing).Error; err != nil {
			// Category doesn't exist, create it
			if err := db.Create(&category).Error; err != nil {
				return fmt.Errorf("failed to create category %s: %w", category.Name, err)
			}
			log.Printf("✅ Created category: %s for %s", category.Name, category.OrganizationID)
		} else {
			log.Printf("📋 Category already exists: %s for %s", category.Name, category.OrganizationID)
		}
	}

	return nil
}

// seedOrganizationWorkflows creates workflows for each organization
func seedOrganizationWorkflows(db *gorm.DB) error {
	// This would create organization-specific workflows
	// For now, we'll skip this as it requires the workflow model structure
	log.Println("📋 Skipping workflow seeding (requires workflow model updates)")
	return nil
}

// seedOrganizationDocuments creates sample documents for each organization
func seedOrganizationDocuments(db *gorm.DB) error {
	// Seed requisitions for Demo Organization
	if err := seedDemoRequisitions(db); err != nil {
		return fmt.Errorf("failed to seed demo requisitions: %w", err)
	}

	// Seed requisitions for ACME Corporation
	if err := seedAcmeRequisitions(db); err != nil {
		return fmt.Errorf("failed to seed ACME requisitions: %w", err)
	}

	// Seed budgets for both organizations
	if err := seedOrganizationBudgets(db); err != nil {
		return fmt.Errorf("failed to seed organization budgets: %w", err)
	}

	return nil
}

// seedDemoRequisitions creates requisitions for Demo Organization
func seedDemoRequisitions(db *gorm.DB) error {
	orgID := "org-demo-001"
	requesterID := "user-demo-requester-001"
	managerID := "user-demo-manager-001"
	financeID := "user-demo-finance-001"

	// Get category for Demo Org
	var category models.Category
	if err := db.Where("organization_id = ? AND name = ?", orgID, "Office Supplies").First(&category).Error; err != nil {
		return fmt.Errorf("failed to find Demo Org category: %w", err)
	}

	requisitions := []struct {
		title       string
		description string
		amount      float64
		priority    string
		status      string
		department  string
	}{
		{
			title:       "Demo Office Furniture Purchase",
			description: "New desks and chairs for the Demo marketing department",
			amount:      12000.00,
			priority:    "medium",
			status:      "pending",
			department:  "Marketing",
		},
		{
			title:       "Demo Software Licenses",
			description: "Annual renewal of software licenses for Demo organization",
			amount:      6500.00,
			priority:    "high",
			status:      "approved",
			department:  "IT",
		},
		{
			title:       "Demo Training Materials",
			description: "Training materials and resources for Demo staff development",
			amount:      3500.00,
			priority:    "low",
			status:      "draft",
			department:  "HR",
		},
	}

	for i, reqData := range requisitions {
		reqNumber := fmt.Sprintf("DEMO-REQ-%03d", i+1)
		
		// Create requisition items
		items := []types.RequisitionItem{
			{
				ID:          &[]string{uuid.New().String()}[0],
				Description: fmt.Sprintf("Demo Item 1 for %s", reqData.title),
				Quantity:    2,
				UnitPrice:   reqData.amount * 0.6,
				Amount:      reqData.amount * 0.6,
				Unit:        &[]string{"pcs"}[0],
				Category:    &[]string{"Equipment"}[0],
			},
			{
				ID:          &[]string{uuid.New().String()}[0],
				Description: fmt.Sprintf("Demo Item 2 for %s", reqData.title),
				Quantity:    1,
				UnitPrice:   reqData.amount * 0.4,
				Amount:      reqData.amount * 0.4,
				Unit:        &[]string{"set"}[0],
				Category:    &[]string{"Services"}[0],
			},
		}

		// Create metadata
		metadata := map[string]interface{}{
			"requestedFor":      reqData.department,
			"budgetCode":        fmt.Sprintf("DEMO-BUD-%03d", i+1),
			"costCenter":        fmt.Sprintf("DEMO-CC-%03d", i+1),
			"projectCode":       fmt.Sprintf("DEMO-PRJ-%03d", i+1),
		}
		metadataBytes, _ := json.Marshal(metadata)

		// Create action history
		actionHistory := []types.ActionHistoryEntry{
			{
				ID:              uuid.New().String(),
				Action:          "CREATE",
				PerformedBy:     requesterID,
				PerformedByName: "Demo Requester",
				PerformedByRole: "requester",
				Timestamp:       time.Now().AddDate(0, 0, -5),
				Comments:        "Demo requisition created",
				ActionType:      "CREATE",
				NewStatus:       "draft",
			},
		}

		// Create approval history based on status
		approvalHistory := []types.ApprovalRecord{}
		if reqData.status == "approved" {
			approvalHistory = append(approvalHistory,
				types.ApprovalRecord{
					ApproverID:   managerID,
					ApproverName: "Demo Manager",
					Status:       "approved",
					Comments:     "Approved by Demo manager",
					Signature:    "digital_signature_" + managerID,
					ApprovedAt:   time.Now().AddDate(0, 0, -3),
				},
				types.ApprovalRecord{
					ApproverID:   financeID,
					ApproverName: "Demo Finance Officer",
					Status:       "approved",
					Comments:     "Approved by Demo finance",
					Signature:    "digital_signature_" + financeID,
					ApprovedAt:   time.Now().AddDate(0, 0, -1),
				})
		}

		requisition := models.Requisition{
			ID:                uuid.New().String(),
			OrganizationID:    orgID,
			REQNumber:         reqNumber,
			RequesterId:       requesterID,
			RequesterName:     "Demo Requester",
			Title:             reqData.title,
			Description:       reqData.description,
			Department:        reqData.department,
			DepartmentId:      uuid.New().String(),
			Status:            reqData.status,
			Priority:          reqData.priority,
			TotalAmount:       reqData.amount,
			Currency:          "ZMW",
			CategoryID:        &category.ID,
			IsEstimate:        false,
			ApprovalStage:     len(approvalHistory),
			RequiredByDate:    time.Now().AddDate(0, 0, 30),
			Items:             datatypes.NewJSONType(items),
			ApprovalHistory:   datatypes.NewJSONType(approvalHistory),
			ActionHistory:     datatypes.NewJSONType(actionHistory),
			Metadata:          datatypes.JSON(metadataBytes),
			CreatedAt:         time.Now().AddDate(0, 0, -5),
			UpdatedAt:         time.Now(),
		}

		var existing models.Requisition
		if err := db.Where("req_number = ? AND organization_id = ?", reqNumber, orgID).First(&existing).Error; err != nil {
			// Requisition doesn't exist, create it
			if err := db.Create(&requisition).Error; err != nil {
				return fmt.Errorf("failed to create Demo requisition %s: %w", reqData.title, err)
			}
			log.Printf("✅ Created Demo requisition: %s (%s)", reqData.title, reqNumber)
		} else {
			log.Printf("📋 Demo requisition already exists: %s (%s)", reqData.title, reqNumber)
		}
	}

	return nil
}

// seedAcmeRequisitions creates requisitions for ACME Corporation
func seedAcmeRequisitions(db *gorm.DB) error {
	orgID := "org-acme-001"
	requesterID := "user-acme-requester-001"
	managerID := "user-acme-manager-001"
	financeID := "user-acme-finance-001"

	// Get category for ACME Corp
	var category models.Category
	if err := db.Where("organization_id = ? AND name = ?", orgID, "Manufacturing Equipment").First(&category).Error; err != nil {
		return fmt.Errorf("failed to find ACME Corp category: %w", err)
	}

	requisitions := []struct {
		title       string
		description string
		amount      float64
		priority    string
		status      string
		department  string
	}{
		{
			title:       "ACME Production Line Upgrade",
			description: "Upgrade manufacturing equipment for increased production capacity",
			amount:      45000.00,
			priority:    "high",
			status:      "pending",
			department:  "Production",
		},
		{
			title:       "ACME Safety Equipment Renewal",
			description: "Annual renewal of safety equipment and protective gear",
			amount:      18000.00,
			priority:    "urgent",
			status:      "approved",
			department:  "Safety",
		},
		{
			title:       "ACME Raw Materials Stock",
			description: "Quarterly stock replenishment of raw materials",
			amount:      32000.00,
			priority:    "medium",
			status:      "rejected",
			department:  "Procurement",
		},
		{
			title:       "ACME Quality Control Equipment",
			description: "New quality control and testing equipment",
			amount:      25000.00,
			priority:    "medium",
			status:      "draft",
			department:  "Quality",
		},
	}

	for i, reqData := range requisitions {
		reqNumber := fmt.Sprintf("ACME-REQ-%03d", i+1)
		
		// Create requisition items
		items := []types.RequisitionItem{
			{
				ID:          &[]string{uuid.New().String()}[0],
				Description: fmt.Sprintf("ACME Item 1 for %s", reqData.title),
				Quantity:    3,
				UnitPrice:   reqData.amount * 0.7,
				Amount:      reqData.amount * 0.7,
				Unit:        &[]string{"units"}[0],
				Category:    &[]string{"Equipment"}[0],
			},
			{
				ID:          &[]string{uuid.New().String()}[0],
				Description: fmt.Sprintf("ACME Item 2 for %s", reqData.title),
				Quantity:    1,
				UnitPrice:   reqData.amount * 0.3,
				Amount:      reqData.amount * 0.3,
				Unit:        &[]string{"lot"}[0],
				Category:    &[]string{"Materials"}[0],
			},
		}

		// Create metadata
		metadata := map[string]interface{}{
			"requestedFor":      reqData.department,
			"budgetCode":        fmt.Sprintf("ACME-BUD-%03d", i+1),
			"costCenter":        fmt.Sprintf("ACME-CC-%03d", i+1),
			"projectCode":       fmt.Sprintf("ACME-PRJ-%03d", i+1),
		}
		metadataBytes, _ := json.Marshal(metadata)

		// Create action history
		actionHistory := []types.ActionHistoryEntry{
			{
				ID:              uuid.New().String(),
				Action:          "CREATE",
				PerformedBy:     requesterID,
				PerformedByName: "ACME Requester",
				PerformedByRole: "requester",
				Timestamp:       time.Now().AddDate(0, 0, -7),
				Comments:        "ACME requisition created",
				ActionType:      "CREATE",
				NewStatus:       "draft",
			},
		}

		// Create approval history based on status
		approvalHistory := []types.ApprovalRecord{}
		if reqData.status == "approved" {
			approvalHistory = append(approvalHistory,
				types.ApprovalRecord{
					ApproverID:   managerID,
					ApproverName: "ACME Manager",
					Status:       "approved",
					Comments:     "Approved by ACME manager",
					Signature:    "digital_signature_" + managerID,
					ApprovedAt:   time.Now().AddDate(0, 0, -4),
				},
				types.ApprovalRecord{
					ApproverID:   financeID,
					ApproverName: "ACME Finance Officer",
					Status:       "approved",
					Comments:     "Approved by ACME finance",
					Signature:    "digital_signature_" + financeID,
					ApprovedAt:   time.Now().AddDate(0, 0, -2),
				})
		} else if reqData.status == "rejected" {
			approvalHistory = append(approvalHistory,
				types.ApprovalRecord{
					ApproverID:   managerID,
					ApproverName: "ACME Manager",
					Status:       "approved",
					Comments:     "Approved by ACME manager",
					Signature:    "digital_signature_" + managerID,
					ApprovedAt:   time.Now().AddDate(0, 0, -5),
				},
				types.ApprovalRecord{
					ApproverID:   financeID,
					ApproverName: "ACME Finance Officer",
					Status:       "rejected",
					Comments:     "Budget exceeded - please reduce amount",
					Signature:    "digital_signature_" + financeID,
					ApprovedAt:   time.Now().AddDate(0, 0, -2),
				})
		}

		requisition := models.Requisition{
			ID:                uuid.New().String(),
			OrganizationID:    orgID,
			REQNumber:         reqNumber,
			RequesterId:       requesterID,
			RequesterName:     "ACME Requester",
			Title:             reqData.title,
			Description:       reqData.description,
			Department:        reqData.department,
			DepartmentId:      uuid.New().String(),
			Status:            reqData.status,
			Priority:          reqData.priority,
			TotalAmount:       reqData.amount,
			Currency:          "USD",
			CategoryID:        &category.ID,
			IsEstimate:        false,
			ApprovalStage:     len(approvalHistory),
			RequiredByDate:    time.Now().AddDate(0, 0, 45),
			Items:             datatypes.NewJSONType(items),
			ApprovalHistory:   datatypes.NewJSONType(approvalHistory),
			ActionHistory:     datatypes.NewJSONType(actionHistory),
			Metadata:          datatypes.JSON(metadataBytes),
			CreatedAt:         time.Now().AddDate(0, 0, -7),
			UpdatedAt:         time.Now(),
		}

		var existing models.Requisition
		if err := db.Where("req_number = ? AND organization_id = ?", reqNumber, orgID).First(&existing).Error; err != nil {
			// Requisition doesn't exist, create it
			if err := db.Create(&requisition).Error; err != nil {
				return fmt.Errorf("failed to create ACME requisition %s: %w", reqData.title, err)
			}
			log.Printf("✅ Created ACME requisition: %s (%s)", reqData.title, reqNumber)
		} else {
			log.Printf("📋 ACME requisition already exists: %s (%s)", reqData.title, reqNumber)
		}
	}

	return nil
}

// seedOrganizationBudgets creates budgets for each organization
func seedOrganizationBudgets(db *gorm.DB) error {
	// Demo Organization Budgets
	demoBudgets := []struct {
		budgetCode      string
		department      string
		totalBudget     float64
		allocatedAmount float64
		status          string
	}{
		{
			budgetCode:      "DEMO-BUD-001",
			department:      "Marketing",
			totalBudget:     50000.00,
			allocatedAmount: 12000.00,
			status:          "approved",
		},
		{
			budgetCode:      "DEMO-BUD-002",
			department:      "IT",
			totalBudget:     75000.00,
			allocatedAmount: 25000.00,
			status:          "approved",
		},
		{
			budgetCode:      "DEMO-BUD-003",
			department:      "HR",
			totalBudget:     30000.00,
			allocatedAmount: 8000.00,
			status:          "pending",
		},
	}

	// ACME Corporation Budgets
	acmeBudgets := []struct {
		budgetCode      string
		department      string
		totalBudget     float64
		allocatedAmount float64
		status          string
	}{
		{
			budgetCode:      "ACME-BUD-001",
			department:      "Production",
			totalBudget:     200000.00,
			allocatedAmount: 75000.00,
			status:          "approved",
		},
		{
			budgetCode:      "ACME-BUD-002",
			department:      "Safety",
			totalBudget:     80000.00,
			allocatedAmount: 35000.00,
			status:          "approved",
		},
		{
			budgetCode:      "ACME-BUD-003",
			department:      "Quality",
			totalBudget:     60000.00,
			allocatedAmount: 15000.00,
			status:          "draft",
		},
	}

	// Create Demo budgets
	for _, budgetData := range demoBudgets {
		budget := models.Budget{
			ID:              uuid.New().String(),
			OrganizationID:  "org-demo-001",
			OwnerID:         "user-demo-admin-001",
			BudgetCode:      budgetData.budgetCode,
			Department:      budgetData.department,
			Status:          budgetData.status,
			FiscalYear:      "2024",
			TotalBudget:     budgetData.totalBudget,
			AllocatedAmount: budgetData.allocatedAmount,
			RemainingAmount: budgetData.totalBudget - budgetData.allocatedAmount,
			ApprovalStage:   0,
			CreatedAt:       time.Now().AddDate(0, 0, -10),
			UpdatedAt:       time.Now(),
		}

		budget.ApprovalHistory = datatypes.NewJSONType([]types.ApprovalRecord{})

		var existing models.Budget
		if err := db.Where("budget_code = ? AND organization_id = ?", budgetData.budgetCode, "org-demo-001").First(&existing).Error; err != nil {
			// Budget doesn't exist, create it
			if err := db.Create(&budget).Error; err != nil {
				return fmt.Errorf("failed to create Demo budget %s: %w", budgetData.budgetCode, err)
			}
			log.Printf("✅ Created Demo budget: %s (%s)", budgetData.department, budgetData.budgetCode)
		} else {
			log.Printf("📋 Demo budget already exists: %s (%s)", budgetData.department, budgetData.budgetCode)
		}
	}

	// Create ACME budgets
	for _, budgetData := range acmeBudgets {
		budget := models.Budget{
			ID:              uuid.New().String(),
			OrganizationID:  "org-acme-001",
			OwnerID:         "user-acme-admin-001",
			BudgetCode:      budgetData.budgetCode,
			Department:      budgetData.department,
			Status:          budgetData.status,
			FiscalYear:      "2024",
			TotalBudget:     budgetData.totalBudget,
			AllocatedAmount: budgetData.allocatedAmount,
			RemainingAmount: budgetData.totalBudget - budgetData.allocatedAmount,
			ApprovalStage:   0,
			CreatedAt:       time.Now().AddDate(0, 0, -10),
			UpdatedAt:       time.Now(),
		}

		budget.ApprovalHistory = datatypes.NewJSONType([]types.ApprovalRecord{})

		var existing models.Budget
		if err := db.Where("budget_code = ? AND organization_id = ?", budgetData.budgetCode, "org-acme-001").First(&existing).Error; err != nil {
			// Budget doesn't exist, create it
			if err := db.Create(&budget).Error; err != nil {
				return fmt.Errorf("failed to create ACME budget %s: %w", budgetData.budgetCode, err)
			}
			log.Printf("✅ Created ACME budget: %s (%s)", budgetData.department, budgetData.budgetCode)
		} else {
			log.Printf("📋 ACME budget already exists: %s (%s)", budgetData.department, budgetData.budgetCode)
		}
	}

	return nil
}

// CleanupMultiTenantData removes all multi-tenant test data
func CleanupMultiTenantData(db *gorm.DB) error {
	log.Println("🧹 Cleaning up multi-tenant test data...")

	// Delete requisitions for both organizations
	if err := db.Where("organization_id IN ?", []string{"org-demo-001", "org-acme-001"}).Delete(&models.Requisition{}).Error; err != nil {
		log.Printf("Error cleaning up requisitions: %v", err)
	}

	// Delete budgets for both organizations
	if err := db.Where("organization_id IN ?", []string{"org-demo-001", "org-acme-001"}).Delete(&models.Budget{}).Error; err != nil {
		log.Printf("Error cleaning up budgets: %v", err)
	}

	// Delete categories for both organizations
	if err := db.Where("organization_id IN ?", []string{"org-demo-001", "org-acme-001"}).Delete(&models.Category{}).Error; err != nil {
		log.Printf("Error cleaning up categories: %v", err)
	}

	// Delete organization memberships (including super admin memberships)
	if err := db.Where("organization_id IN ?", []string{"org-demo-001", "org-acme-001"}).Delete(&models.OrganizationMember{}).Error; err != nil {
		log.Printf("Error cleaning up organization memberships: %v", err)
	}

	// Delete organization-specific users (but preserve super admin)
	userIDs := []string{
		"user-demo-admin-001", "user-demo-manager-001", "user-demo-requester-001", "user-demo-finance-001",
		"user-acme-admin-001", "user-acme-manager-001", "user-acme-requester-001", "user-acme-finance-001",
	}
	if err := db.Where("id IN ?", userIDs).Delete(&models.User{}).Error; err != nil {
		log.Printf("Error cleaning up users: %v", err)
	}

	// Reset super admin's current organization to null
	if err := db.Model(&models.User{}).Where("id = ?", "user-admin-001").Update("current_organization_id", nil).Error; err != nil {
		log.Printf("Error resetting super admin organization: %v", err)
	}

	log.Println("✅ Multi-tenant cleanup completed!")
	log.Println("📋 Note: Super admin (admin@liyali.com) preserved but organization memberships removed")
	return nil
}

// stringPtr returns a pointer to a string
func stringPtr(s string) *string {
	return &s
}