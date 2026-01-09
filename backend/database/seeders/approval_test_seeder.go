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

// SeedApprovalTestData creates sample requisitions with approval tasks for testing
func SeedApprovalTestData(db *gorm.DB) error {
	log.Println("🌱 Seeding approval test data...")

	// Get the first organization and users
	var org models.Organization
	var users []models.User
	
	// Try to find existing test users first to determine the correct organization
	var existingUsers []models.User
	db.Where("id IN ?", []string{"user-requester-001", "user-manager-001", "user-finance-001"}).Find(&existingUsers)
	
	if len(existingUsers) >= 3 {
		log.Printf("Found %d existing test users", len(existingUsers))
		users = existingUsers
		// Use the organization from the first user
		if existingUsers[0].CurrentOrganizationID != nil {
			if err := db.Where("id = ?", *existingUsers[0].CurrentOrganizationID).First(&org).Error; err == nil {
				log.Printf("Using organization from users: %s (%s)", org.Name, org.ID)
			} else {
				// Fallback to first organization
				if err := db.First(&org).Error; err != nil {
					return fmt.Errorf("no organization found: %w", err)
				}
			}
		} else {
			// Fallback to first organization
			if err := db.First(&org).Error; err != nil {
				return fmt.Errorf("no organization found: %w", err)
			}
		}
	} else {
		// Fallback to original logic
		if err := db.First(&org).Error; err != nil {
			return fmt.Errorf("no organization found: %w", err)
		}

		if err := db.Where("current_organization_id = ?", org.ID).Limit(5).Find(&users).Error; err != nil {
			// Try to find users without organization filter if none found
			if err := db.Limit(5).Find(&users).Error; err != nil {
				return fmt.Errorf("no users found: %w", err)
			}
		}
	}

	// If we don't have enough users, try to find existing test users or create them
	if len(users) < 3 {
		log.Println("Looking for existing test users...")
		
		// Try to find existing test users
		var existingUsers []models.User
		db.Where("id IN ?", []string{"user-requester-001", "user-manager-001", "user-finance-001"}).Find(&existingUsers)
		
		if len(existingUsers) >= 3 {
			log.Printf("Found %d existing test users", len(existingUsers))
			users = existingUsers
			// Update organization to match the users' organization
			if existingUsers[0].CurrentOrganizationID != nil {
				if err := db.Where("id = ?", *existingUsers[0].CurrentOrganizationID).First(&org).Error; err == nil {
					log.Printf("Using organization: %s (%s)", org.Name, org.ID)
				}
			}
		} else {
			log.Println("Creating test users for approval testing...")
			
			testUsers := []models.User{
				{
					ID:                    "user-requester-002",
					Email:                 "requester2@test.com",
					Name:                  "John Requester",
					Password:              "$2a$10$dummy.hash.for.testing", // bcrypt hash for "password"
					Role:                  "requester",
					Active:                true,
					CurrentOrganizationID: &org.ID,
					CreatedAt:             time.Now(),
					UpdatedAt:             time.Now(),
				},
				{
					ID:                    "user-manager-002",
					Email:                 "manager2@test.com",
					Name:                  "Jane Manager",
					Password:              "$2a$10$dummy.hash.for.testing",
					Role:                  "manager",
					Active:                true,
					CurrentOrganizationID: &org.ID,
					CreatedAt:             time.Now(),
					UpdatedAt:             time.Now(),
				},
				{
					ID:                    "user-finance-002",
					Email:                 "finance2@test.com",
					Name:                  "Bob Finance",
					Password:              "$2a$10$dummy.hash.for.testing",
					Role:                  "finance_manager",
					Active:                true,
					CurrentOrganizationID: &org.ID,
					CreatedAt:             time.Now(),
					UpdatedAt:             time.Now(),
				},
			}

			for _, user := range testUsers {
				if err := db.Create(&user).Error; err != nil {
					log.Printf("Warning: Could not create test user %s: %v", user.Email, err)
				} else {
					log.Printf("✅ Created test user: %s (%s)", user.Name, user.Role)
					users = append(users, user)
				}
			}
		}
	}

	if len(users) < 3 {
		return fmt.Errorf("need at least 3 users for approval testing")
	}

	requester := users[0]
	manager := users[1]
	financeManager := users[2]

	// Update user roles for testing
	manager.Role = "manager"
	financeManager.Role = "finance_manager"
	db.Save(&manager)
	db.Save(&financeManager)

	// Get or create a category
	var category models.Category
	db.FirstOrCreate(&category, models.Category{
		Name:           "Office Supplies",
		OrganizationID: org.ID,
		Active:         true,
	})

	// Create test requisitions with different approval scenarios
	testRequisitions := []struct {
		title       string
		description string
		amount      float64
		priority    string
		status      string
		scenario    string
	}{
		{
			title:       "Office Furniture Purchase",
			description: "New desks and chairs for the marketing department",
			amount:      15000.00,
			priority:    "medium",
			status:      "pending",
			scenario:    "pending_manager_approval",
		},
		{
			title:       "Software Licenses Renewal",
			description: "Annual renewal of Microsoft Office and Adobe Creative Suite licenses",
			amount:      8500.00,
			priority:    "high",
			status:      "pending",
			scenario:    "pending_finance_approval",
		},
		{
			title:       "Emergency IT Equipment",
			description: "Replacement laptops for damaged equipment after office incident",
			amount:      25000.00,
			priority:    "urgent",
			status:      "draft",
			scenario:    "ready_for_submission",
		},
		{
			title:       "Marketing Campaign Materials",
			description: "Promotional materials and banners for Q1 marketing campaign",
			amount:      5000.00,
			priority:    "medium",
			status:      "approved",
			scenario:    "fully_approved",
		},
		{
			title:       "Training and Development",
			description: "Professional development courses for team members",
			amount:      12000.00,
			priority:    "low",
			status:      "rejected",
			scenario:    "rejected_by_finance",
		},
	}

	for i, reqData := range testRequisitions {
		// Create requisition items
		items := []types.RequisitionItem{
			{
				ID:          &[]string{uuid.New().String()}[0],
				Description: fmt.Sprintf("Item 1 for %s", reqData.title),
				Quantity:    2,
				UnitPrice:   reqData.amount * 0.6,
				Amount:      reqData.amount * 0.6,
				Unit:        &[]string{"pcs"}[0],
				Category:    &[]string{"Equipment"}[0],
			},
			{
				ID:          &[]string{uuid.New().String()}[0],
				Description: fmt.Sprintf("Item 2 for %s", reqData.title),
				Quantity:    1,
				UnitPrice:   reqData.amount * 0.4,
				Amount:      reqData.amount * 0.4,
				Unit:        &[]string{"set"}[0],
				Category:    &[]string{"Services"}[0],
			},
		}

		// Create metadata
		metadata := map[string]interface{}{
			"requestedFor":      fmt.Sprintf("Department %d", i+1),
			"otherCategoryText": "",
			"testScenario":      reqData.scenario,
			"budgetCode":        fmt.Sprintf("BUD-%03d", i+1),
			"costCenter":        fmt.Sprintf("CC-%03d", i+1),
			"projectCode":       fmt.Sprintf("PRJ-%03d", i+1),
		}
		metadataBytes, _ := json.Marshal(metadata)

		// Create action history based on scenario
		actionHistory := []types.ActionHistoryEntry{
			{
				ID:              uuid.New().String(),
				Action:          "CREATE",
				PerformedBy:     requester.ID,
				PerformedByName: requester.Name,
				PerformedByRole: requester.Role,
				Timestamp:       time.Now().AddDate(0, 0, -7), // 7 days ago
				Comments:        "Requisition created",
				ActionType:      "CREATE",
				NewStatus:       "draft",
			},
		}

		// Create approval history based on scenario
		approvalHistory := []types.ApprovalRecord{}

		// Add more history based on scenario
		switch reqData.scenario {
		case "pending_finance_approval":
			// Manager already approved
			actionHistory = append(actionHistory, types.ActionHistoryEntry{
				ID:              uuid.New().String(),
				Action:          "APPROVE",
				PerformedBy:     manager.ID,
				PerformedByName: manager.Name,
				PerformedByRole: manager.Role,
				Timestamp:       time.Now().AddDate(0, 0, -3), // 3 days ago
				Comments:        "Approved by department manager",
				ActionType:      "APPROVE",
				PreviousStatus:  "pending",
				NewStatus:       "pending",
			})
			approvalHistory = append(approvalHistory, types.ApprovalRecord{
				ApproverID:   manager.ID,
				ApproverName: manager.Name,
				Status:       "approved",
				Comments:     "Approved by department manager",
				Signature:    "digital_signature_" + manager.ID,
				ApprovedAt:   time.Now().AddDate(0, 0, -3),
			})

		case "fully_approved":
			// Both manager and finance approved
			actionHistory = append(actionHistory,
				types.ActionHistoryEntry{
					ID:              uuid.New().String(),
					Action:          "APPROVE",
					PerformedBy:     manager.ID,
					PerformedByName: manager.Name,
					PerformedByRole: manager.Role,
					Timestamp:       time.Now().AddDate(0, 0, -5),
					Comments:        "Approved by department manager",
					ActionType:      "APPROVE",
					PreviousStatus:  "pending",
					NewStatus:       "pending",
				},
				types.ActionHistoryEntry{
					ID:              uuid.New().String(),
					Action:          "APPROVE",
					PerformedBy:     financeManager.ID,
					PerformedByName: financeManager.Name,
					PerformedByRole: financeManager.Role,
					Timestamp:       time.Now().AddDate(0, 0, -2),
					Comments:        "Approved by finance manager",
					ActionType:      "APPROVE",
					PreviousStatus:  "pending",
					NewStatus:       "approved",
				})
			approvalHistory = append(approvalHistory,
				types.ApprovalRecord{
					ApproverID:   manager.ID,
					ApproverName: manager.Name,
					Status:       "approved",
					Comments:     "Approved by department manager",
					Signature:    "digital_signature_" + manager.ID,
					ApprovedAt:   time.Now().AddDate(0, 0, -5),
				},
				types.ApprovalRecord{
					ApproverID:   financeManager.ID,
					ApproverName: financeManager.Name,
					Status:       "approved",
					Comments:     "Approved by finance manager",
					Signature:    "digital_signature_" + financeManager.ID,
					ApprovedAt:   time.Now().AddDate(0, 0, -2),
				})

		case "rejected_by_finance":
			// Manager approved, finance rejected
			actionHistory = append(actionHistory,
				types.ActionHistoryEntry{
					ID:              uuid.New().String(),
					Action:          "APPROVE",
					PerformedBy:     manager.ID,
					PerformedByName: manager.Name,
					PerformedByRole: manager.Role,
					Timestamp:       time.Now().AddDate(0, 0, -4),
					Comments:        "Approved by department manager",
					ActionType:      "APPROVE",
					PreviousStatus:  "pending",
					NewStatus:       "pending",
				},
				types.ActionHistoryEntry{
					ID:              uuid.New().String(),
					Action:          "REJECT",
					PerformedBy:     financeManager.ID,
					PerformedByName: financeManager.Name,
					PerformedByRole: financeManager.Role,
					Timestamp:       time.Now().AddDate(0, 0, -1),
					Comments:        "Budget constraints - please reduce amount",
					Remarks:         "Budget constraints - please reduce amount",
					ActionType:      "REJECT",
					PreviousStatus:  "pending",
					NewStatus:       "rejected",
				})
			approvalHistory = append(approvalHistory,
				types.ApprovalRecord{
					ApproverID:   manager.ID,
					ApproverName: manager.Name,
					Status:       "approved",
					Comments:     "Approved by department manager",
					Signature:    "digital_signature_" + manager.ID,
					ApprovedAt:   time.Now().AddDate(0, 0, -4),
				},
				types.ApprovalRecord{
					ApproverID:   financeManager.ID,
					ApproverName: financeManager.Name,
					Status:       "rejected",
					Comments:     "Budget constraints - please reduce amount",
					Signature:    "digital_signature_" + financeManager.ID,
					ApprovedAt:   time.Now().AddDate(0, 0, -1),
				})
		}

		// Create requisition
		requisition := models.Requisition{
			ID:                uuid.New().String(),
			OrganizationID:    org.ID,
			REQNumber:         fmt.Sprintf("REQ-%03d", i+1),
			RequesterId:       requester.ID,
			RequesterName:     requester.Name,
			Title:             reqData.title,
			Description:       reqData.description,
			Department:        "Test Department",
			DepartmentId:      uuid.New().String(),
			Status:            reqData.status,
			Priority:          reqData.priority,
			TotalAmount:       reqData.amount,
			Currency:          "ZMW",
			CategoryID:        &category.ID,
			IsEstimate:        false,
			ApprovalStage:     len(approvalHistory),
			RequiredByDate:    time.Now().AddDate(0, 0, 30), // 30 days from now
			Items:             datatypes.NewJSONType(items),
			ApprovalHistory:   datatypes.NewJSONType(approvalHistory),
			Metadata:          datatypes.JSON(metadataBytes),
			CreatedAt:         time.Now().AddDate(0, 0, -7),
			UpdatedAt:         time.Now(),
		}

		if err := db.Create(&requisition).Error; err != nil {
			log.Printf("Error creating requisition %s: %v", reqData.title, err)
			continue
		}

		// Create approval tasks based on scenario
		switch reqData.scenario {
		case "pending_manager_approval":
			// Create pending task for manager
			task := models.ApprovalTask{
				ID:             uuid.New().String(),
				DocumentID:     requisition.ID,
				DocumentType:   "REQUISITION",
				OrganizationID: org.ID,
				Stage:          1,
				AssignedTo:     manager.ID,
				Status:         "pending",
				DueAt:          &[]time.Time{time.Now().AddDate(0, 0, 3)}[0], // 3 days from now
				CreatedAt:      time.Now().AddDate(0, 0, -7),
				UpdatedAt:      time.Now(),
			}
			db.Create(&task)

		case "pending_finance_approval":
			// Create completed task for manager
			managerTask := models.ApprovalTask{
				ID:             uuid.New().String(),
				DocumentID:     requisition.ID,
				DocumentType:   "REQUISITION",
				OrganizationID: org.ID,
				Stage:          1,
				AssignedTo:     manager.ID,
				Status:         "approved",
				ApprovedBy:     &manager.ID,
				ApprovedAt:     &[]time.Time{time.Now().AddDate(0, 0, -3)}[0],
				Signature:      &[]string{"digital_signature_" + manager.ID}[0],
				Comments:       &[]string{"Approved by department manager"}[0],
				CreatedAt:      time.Now().AddDate(0, 0, -7),
				UpdatedAt:      time.Now().AddDate(0, 0, -3),
			}
			db.Create(&managerTask)

			// Create pending task for finance
			financeTask := models.ApprovalTask{
				ID:             uuid.New().String(),
				DocumentID:     requisition.ID,
				DocumentType:   "REQUISITION",
				OrganizationID: org.ID,
				Stage:          2,
				AssignedTo:     financeManager.ID,
				Status:         "pending",
				DueAt:          &[]time.Time{time.Now().AddDate(0, 0, 2)}[0], // 2 days from now
				CreatedAt:      time.Now().AddDate(0, 0, -3),
				UpdatedAt:      time.Now(),
			}
			db.Create(&financeTask)
		}

		log.Printf("✅ Created test requisition: %s (%s)", reqData.title, reqData.scenario)
	}

	log.Println("🎉 Approval test data seeded successfully!")
	return nil
}

// CleanupApprovalTestData removes test data (useful for re-seeding)
func CleanupApprovalTestData(db *gorm.DB) error {
	log.Println("🧹 Cleaning up approval test data...")

	// Delete test requisitions (those with REQ-001 to REQ-005 pattern)
	if err := db.Where("req_number LIKE 'REQ-%'").Delete(&models.Requisition{}).Error; err != nil {
		log.Printf("Error cleaning up requisitions: %v", err)
	}

	// Delete test approval tasks
	if err := db.Where("document_type = 'REQUISITION'").Delete(&models.ApprovalTask{}).Error; err != nil {
		log.Printf("Error cleaning up approval tasks: %v", err)
	}

	log.Println("✅ Cleanup completed!")
	return nil
}