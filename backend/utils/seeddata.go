package utils

import (
	"log"

	"github.com/liyali/liyali-gateway/models"
	"gorm.io/gorm"
)

// SeedTestUsers creates test users for development
func SeedTestUsers(db *gorm.DB) error {
	// Default password hash for "password" - bcrypt with cost 10
	defaultPasswordHash := "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi"
	
	testUsers := []models.User{
		{
			ID:       "user-admin-001",
			Email:    "admin@liyali.com",
			Name:     "Admin User",
			Password: defaultPasswordHash,
			Role:     "admin",
			Active:   true,
		},
		{
			ID:       "user-approver-001",
			Email:    "approver@liyali.com",
			Name:     "John Approver",
			Password: defaultPasswordHash,
			Role:     "approver",
			Active:   true,
		},
		{
			ID:       "user-requester-001",
			Email:    "requester@liyali.com",
			Name:     "Jane Requester",
			Password: defaultPasswordHash,
			Role:     "requester",
			Active:   true,
		},
		{
			ID:       "user-finance-001",
			Email:    "finance@liyali.com",
			Name:     "Finance Officer",
			Password: defaultPasswordHash,
			Role:     "finance",
			Active:   true,
		},
		{
			ID:       "user-viewer-001",
			Email:    "viewer@liyali.com",
			Name:     "Viewer User",
			Password: defaultPasswordHash,
			Role:     "viewer",
			Active:   true,
		},
	}

	for _, user := range testUsers {
		// Check if user already exists
		var existing models.User
		if err := db.Where("email = ?", user.Email).First(&existing).Error; err == nil {
			log.Printf("User already exists: %s", user.Email)
			continue
		}

		if err := db.Create(&user).Error; err != nil {
			log.Printf("Error creating seed user %s: %v", user.Email, err)
			return err
		}
		log.Printf("Created seed user: %s (%s)", user.Email, user.Role)
	}

	return nil
}

// SeedTestOrganizations creates test organizations for development
func SeedTestOrganizations(db *gorm.DB) error {
	testOrganizations := []models.Organization{
		{
			ID:          "org-demo-001",
			Name:        "Demo Organization",
			Slug:        "demo-org",
			Description: "Demo organization for testing",
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

	for _, org := range testOrganizations {
		// Check if organization already exists
		var existing models.Organization
		if err := db.Where("slug = ?", org.Slug).First(&existing).Error; err == nil {
			log.Printf("Organization already exists: %s", org.Slug)
			continue
		}

		if err := db.Create(&org).Error; err != nil {
			log.Printf("Error creating seed organization %s: %v", org.Name, err)
			return err
		}
		log.Printf("Created seed organization: %s (%s)", org.Name, org.Slug)
	}

	return nil
}

// SeedTestVendors creates test vendors for development
func SeedTestVendors(db *gorm.DB) error {
	testVendors := []models.Vendor{
		{
			ID:          "vendor-001",
			VendorCode:  "VND-001",
			Name:        "ABC Supplies Ltd",
			Email:       "contact@abcsupplies.com",
			Phone:       "+1-555-0101",
			Country:     "United States",
			City:        "New York",
			BankAccount: "1234567890",
			TaxID:       "12-3456789",
			Active:      true,
			CreatedBy:   "user-admin-001",
		},
		{
			ID:          "vendor-002",
			VendorCode:  "VND-002",
			Name:        "Global Tech Solutions",
			Email:       "sales@globaltech.com",
			Phone:       "+1-555-0102",
			Country:     "United States",
			City:        "San Francisco",
			BankAccount: "0987654321",
			TaxID:       "98-7654321",
			Active:      true,
			CreatedBy:   "user-admin-001",
		},
		{
			ID:          "vendor-003",
			VendorCode:  "VND-003",
			Name:        "Premium Services Inc",
			Email:       "info@premiumservices.com",
			Phone:       "+1-555-0103",
			Country:     "Canada",
			City:        "Toronto",
			BankAccount: "5555666677",
			TaxID:       "55-5555555",
			Active:      true,
			CreatedBy:   "user-admin-001",
		},
	}

	for _, vendor := range testVendors {
		// Check if vendor already exists
		var existing models.Vendor
		if err := db.Where("vendor_code = ?", vendor.VendorCode).First(&existing).Error; err == nil {
			log.Printf("Vendor already exists: %s", vendor.VendorCode)
			continue
		}

		if err := db.Create(&vendor).Error; err != nil {
			log.Printf("Error creating seed vendor %s: %v", vendor.VendorCode, err)
			return err
		}
		log.Printf("Created seed vendor: %s (%s)", vendor.Name, vendor.VendorCode)
	}

	return nil
}

// SeedTestCategories creates test categories for development
func SeedTestCategories(db *gorm.DB) error {
	// Use the demo organization for categories
	defaultOrgID := "org-demo-001"

	testCategories := []models.Category{
		{
			ID:             "cat-001",
			OrganizationID: defaultOrgID,
			Name:           "Office Supplies",
			Description:    "General office supplies and stationery",
			Active:         true,
		},
		{
			ID:             "cat-002",
			OrganizationID: defaultOrgID,
			Name:           "IT Equipment",
			Description:    "Computers, software, and IT hardware",
			Active:         true,
		},
		{
			ID:             "cat-003",
			OrganizationID: defaultOrgID,
			Name:           "Facilities",
			Description:    "Facility maintenance and utilities",
			Active:         true,
		},
	}

	for _, category := range testCategories {
		// Check if category already exists
		var existing models.Category
		if err := db.Where("name = ? AND organization_id = ?", category.Name, category.OrganizationID).First(&existing).Error; err == nil {
			log.Printf("Category already exists: %s", category.Name)
			continue
		}

		if err := db.Create(&category).Error; err != nil {
			log.Printf("Error creating seed category %s: %v", category.Name, err)
			return err
		}
		log.Printf("Created seed category: %s", category.Name)
	}

	return nil
}

// SeedDatabase seeds all test data
func SeedDatabase(db *gorm.DB) error {
	log.Println("🌱 Seeding database with test data...")

	// Seed in dependency order
	if err := SeedTestUsers(db); err != nil {
		log.Printf("Error seeding users: %v", err)
		return err
	}

	if err := SeedTestOrganizations(db); err != nil {
		log.Printf("Error seeding organizations: %v", err)
		return err
	}

	if err := SeedTestVendors(db); err != nil {
		log.Printf("Error seeding vendors: %v", err)
		return err
	}

	if err := SeedTestCategories(db); err != nil {
		log.Printf("Error seeding categories: %v", err)
		return err
	}

	if err := SeedTestWorkflowTasks(db); err != nil {
		log.Printf("Error seeding workflow tasks: %v", err)
		return err
	}

	log.Println("✓ Database seeding completed")
	return nil
}

// SeedTestWorkflowTasks creates test workflow tasks for development
func SeedTestWorkflowTasks(db *gorm.DB) error {
	// Check if there's a submitted requisition that needs a workflow task
	var submittedReq models.Requisition
	if err := db.Where("status = ? AND document_number = ?", "submitted", "REQ-260111-003").First(&submittedReq).Error; err != nil {
		log.Printf("No submitted requisition found for workflow task creation")
		return nil // Not an error, just no data to work with
	}

	log.Printf("Found submitted requisition: %s - %s", submittedReq.DocumentNumber, submittedReq.Title)

	// Create workflow assignment
	workflowAssignment := models.WorkflowAssignment{
		ID:              "wa-req-260111-003",
		OrganizationID:  "org-demo-001",
		EntityID:        submittedReq.ID,
		EntityType:      "requisition",
		WorkflowVersion: 1,
		CurrentStage:    1,
		Status:          "in_progress",
		AssignedBy:      "user-admin-001",
	}

	// Check if workflow assignment already exists
	var existingWA models.WorkflowAssignment
	if err := db.Where("id = ?", workflowAssignment.ID).First(&existingWA).Error; err != nil {
		if err := db.Create(&workflowAssignment).Error; err != nil {
			log.Printf("Error creating workflow assignment: %v", err)
			return err
		}
		log.Printf("Created workflow assignment for %s", submittedReq.DocumentNumber)
	}

	// Create workflow task
	workflowTask := models.WorkflowTask{
		ID:                   "wt-req-260111-003-stage1",
		OrganizationID:       "org-demo-001",
		WorkflowAssignmentID: "wa-req-260111-003",
		EntityID:             submittedReq.ID,
		EntityType:           "requisition",
		StageNumber:          1,
		StageName:            "Manager Approval",
		AssignmentType:       "role",
		AssignedRole:         stringPtr("approver"),
		Status:               "pending",
		Priority:             "medium",
		Version:              1,
	}

	// Check if workflow task already exists
	var existingWT models.WorkflowTask
	if err := db.Where("id = ?", workflowTask.ID).First(&existingWT).Error; err != nil {
		if err := db.Create(&workflowTask).Error; err != nil {
			log.Printf("Error creating workflow task: %v", err)
			return err
		}
		log.Printf("Created workflow task for %s", submittedReq.DocumentNumber)
	}

	return nil
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}
