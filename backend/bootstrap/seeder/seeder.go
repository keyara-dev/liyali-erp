package seeder

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/models"
	"gorm.io/gorm"
)

// DatabaseSeeder handles idempotent database seeding
type DatabaseSeeder struct {
	db     *gorm.DB
	logger *log.Logger
}

// New creates a new database seeder
func New(db *gorm.DB, logger *log.Logger) *DatabaseSeeder {
	return &DatabaseSeeder{
		db:     db,
		logger: logger,
	}
}

// SeedResult contains the results of a seeding operation
type SeedResult struct {
	Entity   string
	Created  int
	Updated  int
	Skipped  int
	Duration time.Duration
	Error    error
}

// SeedAll performs complete database seeding with transactions
func (s *DatabaseSeeder) SeedAll(ctx context.Context) error {
	s.logger.Println("🌱 Starting comprehensive database seeding...")
	startTime := time.Now()

	// Check if tables exist before seeding
	if err := s.validateTablesExist(ctx); err != nil {
		return fmt.Errorf("table validation failed: %w", err)
	}

	// Seed in dependency order with transactions
	seedOperations := []struct {
		name string
		fn   func(context.Context, *gorm.DB) (*SeedResult, error)
	}{
		{"users", s.seedUsers},
		{"organizations", s.seedOrganizations},
		{"organization_members", s.seedOrganizationMembers},
		{"vendors", s.seedVendors},
		{"categories", s.seedCategories},
		{"workflows", s.seedWorkflows},
		{"sample_data", s.seedSampleData},
	}

	totalResults := make([]*SeedResult, 0, len(seedOperations))

	for _, op := range seedOperations {
		result, err := s.executeSeeding(ctx, op.name, op.fn)
		if err != nil {
			return fmt.Errorf("seeding %s failed: %w", op.name, err)
		}
		totalResults = append(totalResults, result)
	}

	// Log summary
	duration := time.Since(startTime)
	s.logSeedingSummary(totalResults, duration)

	return nil
}

// executeSeeding runs a seeding operation within a transaction
func (s *DatabaseSeeder) executeSeeding(ctx context.Context, name string, fn func(context.Context, *gorm.DB) (*SeedResult, error)) (*SeedResult, error) {
	s.logger.Printf("🌱 Seeding %s...", name)
	
	var result *SeedResult
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var txErr error
		result, txErr = fn(ctx, tx)
		return txErr
	})

	if err != nil {
		return nil, err
	}

	s.logger.Printf("✅ Seeded %s: %d created, %d updated, %d skipped (took %v)",
		name, result.Created, result.Updated, result.Skipped, result.Duration)

	return result, nil
}

// validateTablesExist ensures all required tables exist before seeding
func (s *DatabaseSeeder) validateTablesExist(ctx context.Context) error {
	requiredTables := []string{"users", "organizations", "vendors", "categories", "workflows"}
	
	for _, tableName := range requiredTables {
		var count int64
		err := s.db.WithContext(ctx).Raw(`
			SELECT COUNT(*) 
			FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = ?
		`, tableName).Scan(&count).Error

		if err != nil {
			return fmt.Errorf("failed to check table %s: %w", tableName, err)
		}

		if count == 0 {
			return fmt.Errorf("required table %s does not exist", tableName)
		}
	}

	return nil
}

// seedUsers creates initial system users
func (s *DatabaseSeeder) seedUsers(ctx context.Context, tx *gorm.DB) (*SeedResult, error) {
	startTime := time.Now()
	result := &SeedResult{Entity: "users"}

	users := []models.User{
		{
			ID:                    "user-admin-001",
			Email:                 "admin@liyali.com",
			Name:                  "System Administrator",
			Password:              "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi", // password
			Role:                  "admin",
			Active:                true,
			IsSuperAdmin:          true,
			CurrentOrganizationID: stringPtr("org-demo-001"),
		},
		{
			ID:                    "user-approver-001",
			Email:                 "approver@liyali.com",
			Name:                  "John Approver",
			Password:              "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi",
			Role:                  "approver",
			Active:                true,
			CurrentOrganizationID: stringPtr("org-demo-001"),
		},
		{
			ID:                    "user-requester-001",
			Email:                 "requester@liyali.com",
			Name:                  "Jane Requester",
			Password:              "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi",
			Role:                  "requester",
			Active:                true,
			CurrentOrganizationID: stringPtr("org-demo-001"),
		},
		{
			ID:                    "user-finance-001",
			Email:                 "finance@liyali.com",
			Name:                  "Finance Officer",
			Password:              "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi",
			Role:                  "finance",
			Active:                true,
			CurrentOrganizationID: stringPtr("org-demo-001"),
		},
	}

	for _, user := range users {
		// Check if user exists by ID or email
		var existingUser models.User
		err := tx.WithContext(ctx).Where("id = ? OR email = ?", user.ID, user.Email).First(&existingUser).Error
		
		if err == nil {
			// User exists, update it
			err = tx.WithContext(ctx).Model(&existingUser).Updates(map[string]interface{}{
				"name":                      user.Name,
				"password":                  user.Password, // Add password update
				"role":                      user.Role,
				"active":                    user.Active,
				"current_organization_id":   user.CurrentOrganizationID,
				"is_super_admin":           user.IsSuperAdmin,
				"updated_at":               time.Now(),
			}).Error
			
			if err != nil {
				result.Error = fmt.Errorf("failed to update user %s: %w", user.Email, err)
				return result, result.Error
			}
			result.Updated++
		} else {
			// User doesn't exist, create it
			err = tx.WithContext(ctx).Create(&user).Error
			if err != nil {
				result.Error = fmt.Errorf("failed to create user %s: %w", user.Email, err)
				return result, result.Error
			}
			result.Created++
		}
	}

	result.Duration = time.Since(startTime)
	return result, nil
}

// seedOrganizations creates initial organizations
func (s *DatabaseSeeder) seedOrganizations(ctx context.Context, tx *gorm.DB) (*SeedResult, error) {
	startTime := time.Now()
	result := &SeedResult{Entity: "organizations"}

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
		// Check if organization exists by ID or slug
		var existingOrg models.Organization
		err := tx.WithContext(ctx).Where("id = ? OR slug = ?", org.ID, org.Slug).First(&existingOrg).Error
		
		if err == nil {
			// Organization exists, update it
			err = tx.WithContext(ctx).Model(&existingOrg).Updates(map[string]interface{}{
				"name":        org.Name,
				"description": org.Description,
				"active":      org.Active,
				"tier":        org.Tier,
				"updated_at":  time.Now(),
			}).Error
			
			if err != nil {
				result.Error = fmt.Errorf("failed to update organization %s: %w", org.Slug, err)
				return result, result.Error
			}
			result.Updated++
		} else {
			// Organization doesn't exist, create it
			err = tx.WithContext(ctx).Create(&org).Error
			if err != nil {
				result.Error = fmt.Errorf("failed to create organization %s: %w", org.Slug, err)
				return result, result.Error
			}
			result.Created++
		}
	}

	result.Duration = time.Since(startTime)
	return result, nil
}

// seedOrganizationMembers creates organization memberships
func (s *DatabaseSeeder) seedOrganizationMembers(ctx context.Context, tx *gorm.DB) (*SeedResult, error) {
	startTime := time.Now()
	result := &SeedResult{Entity: "organization_members"}

	members := []models.OrganizationMember{
		{
			ID:             "member-001",
			OrganizationID: "org-demo-001",
			UserID:         "user-admin-001",
			Role:           "admin",
			Active:         true,
			JoinedAt:       &startTime,
		},
		{
			ID:             "member-002",
			OrganizationID: "org-demo-001",
			UserID:         "user-approver-001",
			Role:           "approver",
			Active:         true,
			JoinedAt:       &startTime,
		},
		{
			ID:             "member-003",
			OrganizationID: "org-demo-001",
			UserID:         "user-requester-001",
			Role:           "requester",
			Active:         true,
			JoinedAt:       &startTime,
		},
	}

	for _, member := range members {
		// Check if member exists by ID or org+user combination
		var existingMember models.OrganizationMember
		err := tx.WithContext(ctx).Where("id = ? OR (organization_id = ? AND user_id = ?)", 
			member.ID, member.OrganizationID, member.UserID).First(&existingMember).Error
		
		if err == nil {
			// Member exists, update it
			err = tx.WithContext(ctx).Model(&existingMember).Updates(map[string]interface{}{
				"role":       member.Role,
				"active":     member.Active,
				"updated_at": time.Now(),
			}).Error
			
			if err != nil {
				result.Error = fmt.Errorf("failed to update member %s-%s: %w", member.OrganizationID, member.UserID, err)
				return result, result.Error
			}
			result.Updated++
		} else {
			// Member doesn't exist, create it
			err = tx.WithContext(ctx).Create(&member).Error
			if err != nil {
				result.Error = fmt.Errorf("failed to create member %s-%s: %w", member.OrganizationID, member.UserID, err)
				return result, result.Error
			}
			result.Created++
		}
	}

	result.Duration = time.Since(startTime)
	return result, nil
}

// seedVendors creates global vendors
func (s *DatabaseSeeder) seedVendors(ctx context.Context, tx *gorm.DB) (*SeedResult, error) {
	startTime := time.Now()
	result := &SeedResult{Entity: "vendors"}

	vendors := []models.Vendor{
		{
			ID:             "vendor-001",
			OrganizationID: "org-demo-001",
			VendorCode:     "VND-001",
			Name:           "ABC Supplies Ltd",
			Email:          "contact@abcsupplies.com",
			Phone:          "+1-555-0101",
			Country:        "United States",
			City:           "New York",
			BankAccount:    "1234567890",
			TaxID:          "12-3456789",
			Active:         true,
			CreatedBy:      "user-admin-001",
		},
		{
			ID:             "vendor-002",
			OrganizationID: "org-demo-001",
			VendorCode:     "VND-002",
			Name:           "Global Tech Solutions",
			Email:          "sales@globaltech.com",
			Phone:          "+1-555-0102",
			Country:        "United States",
			City:           "San Francisco",
			BankAccount:    "0987654321",
			TaxID:          "98-7654321",
			Active:         true,
			CreatedBy:      "user-admin-001",
		},
		{
			ID:             "vendor-003",
			OrganizationID: "org-demo-001",
			VendorCode:     "VND-003",
			Name:           "Premium Services Inc",
			Email:          "info@premiumservices.com",
			Phone:          "+1-555-0103",
			Country:        "Canada",
			City:           "Toronto",
			BankAccount:    "5555666677",
			TaxID:          "55-5555555",
			Active:         true,
			CreatedBy:      "user-admin-001",
		},
	}

	for _, vendor := range vendors {
		// Check if vendor exists by ID or vendor_code
		var existingVendor models.Vendor
		err := tx.WithContext(ctx).Where("id = ? OR vendor_code = ?", vendor.ID, vendor.VendorCode).First(&existingVendor).Error
		
		if err == nil {
			// Vendor exists, update it
			err = tx.WithContext(ctx).Model(&existingVendor).Updates(map[string]interface{}{
				"name":        vendor.Name,
				"email":       vendor.Email,
				"phone":       vendor.Phone,
				"country":     vendor.Country,
				"city":        vendor.City,
				"active":      vendor.Active,
				"updated_at":  time.Now(),
			}).Error
			
			if err != nil {
				result.Error = fmt.Errorf("failed to update vendor %s: %w", vendor.VendorCode, err)
				return result, result.Error
			}
			result.Updated++
		} else {
			// Vendor doesn't exist, create it
			err = tx.WithContext(ctx).Create(&vendor).Error
			if err != nil {
				result.Error = fmt.Errorf("failed to create vendor %s: %w", vendor.VendorCode, err)
				return result, result.Error
			}
			result.Created++
		}
	}

	result.Duration = time.Since(startTime)
	return result, nil
}

// seedCategories creates organization-specific categories
func (s *DatabaseSeeder) seedCategories(ctx context.Context, tx *gorm.DB) (*SeedResult, error) {
	startTime := time.Now()
	result := &SeedResult{Entity: "categories"}

	categories := []models.Category{
		{
			ID:             "cat-001",
			OrganizationID: "org-demo-001",
			Name:           "Office Supplies",
			Description:    "General office supplies and stationery",
			Active:         true,
		},
		{
			ID:             "cat-002",
			OrganizationID: "org-demo-001",
			Name:           "IT Equipment",
			Description:    "Computers, software, and IT hardware",
			Active:         true,
		},
		{
			ID:             "cat-003",
			OrganizationID: "org-demo-001",
			Name:           "Facilities",
			Description:    "Facility maintenance and utilities",
			Active:         true,
		},
	}

	for _, category := range categories {
		// Check if category exists by ID or org+name combination
		var existingCategory models.Category
		err := tx.WithContext(ctx).Where("id = ? OR (organization_id = ? AND name = ?)", 
			category.ID, category.OrganizationID, category.Name).First(&existingCategory).Error
		
		if err == nil {
			// Category exists, update it
			err = tx.WithContext(ctx).Model(&existingCategory).Updates(map[string]interface{}{
				"description": category.Description,
				"active":      category.Active,
				"updated_at":  time.Now(),
			}).Error
			
			if err != nil {
				result.Error = fmt.Errorf("failed to update category %s: %w", category.Name, err)
				return result, result.Error
			}
			result.Updated++
		} else {
			// Category doesn't exist, create it
			err = tx.WithContext(ctx).Create(&category).Error
			if err != nil {
				result.Error = fmt.Errorf("failed to create category %s: %w", category.Name, err)
				return result, result.Error
			}
			result.Created++
		}
	}

	result.Duration = time.Since(startTime)
	return result, nil
}

// seedWorkflows creates default workflows for each entity type
func (s *DatabaseSeeder) seedWorkflows(ctx context.Context, tx *gorm.DB) (*SeedResult, error) {
	startTime := time.Now()
	result := &SeedResult{Entity: "workflows"}

	// Define default workflow stages for different entity types
	requisitionStages := []models.WorkflowStage{
		{
			StageNumber:       1,
			StageName:         "Manager Approval",
			Description:       "Department manager review and approval",
			RequiredRole:      "manager",
			RequiredApprovals: 1,
			TimeoutHours:      intPtr(48),
			CanReject:         true,
			CanReassign:       true,
		},
		{
			StageNumber:       2,
			StageName:         "Finance Approval",
			Description:       "Finance team review for budget compliance",
			RequiredRole:      "finance",
			RequiredApprovals: 1,
			TimeoutHours:      intPtr(24),
			CanReject:         true,
			CanReassign:       false,
		},
	}

	purchaseOrderStages := []models.WorkflowStage{
		{
			StageNumber:       1,
			StageName:         "Procurement Review",
			Description:       "Procurement team review and vendor selection",
			RequiredRole:      "procurement",
			RequiredApprovals: 1,
			TimeoutHours:      intPtr(24),
			CanReject:         true,
			CanReassign:       true,
		},
		{
			StageNumber:       2,
			StageName:         "Finance Approval",
			Description:       "Final finance approval before PO issuance",
			RequiredRole:      "finance",
			RequiredApprovals: 1,
			TimeoutHours:      intPtr(12),
			CanReject:         true,
			CanReassign:       false,
		},
	}

	grnStages := []models.WorkflowStage{
		{
			StageNumber:       1,
			StageName:         "Warehouse Verification",
			Description:       "Warehouse team verifies received goods",
			RequiredRole:      "warehouse",
			RequiredApprovals: 1,
			TimeoutHours:      intPtr(8),
			CanReject:         true,
			CanReassign:       true,
		},
	}

	paymentVoucherStages := []models.WorkflowStage{
		{
			StageNumber:       1,
			StageName:         "Finance Review",
			Description:       "Finance team reviews payment request",
			RequiredRole:      "finance",
			RequiredApprovals: 1,
			TimeoutHours:      intPtr(24),
			CanReject:         true,
			CanReassign:       true,
		},
		{
			StageNumber:       2,
			StageName:         "Admin Approval",
			Description:       "Final admin approval for payment processing",
			RequiredRole:      "admin",
			RequiredApprovals: 1,
			TimeoutHours:      intPtr(12),
			CanReject:         true,
			CanReassign:       false,
		},
	}

	workflows := []struct {
		workflow models.Workflow
		stages   []models.WorkflowStage
	}{
		{
			workflow: models.Workflow{
				ID:             uuid.New(), // Generate a new UUID
				OrganizationID: "org-demo-001",
				Name:           "Standard Requisition Workflow",
				Description:    "Default workflow for requisition approvals",
				DocumentType:   "requisition",
				EntityType:     "requisition",
				Version:        1,
				IsActive:       true,
				IsDefault:      true,
				CreatedBy:      "user-admin-001",
			},
			stages: requisitionStages,
		},
		{
			workflow: models.Workflow{
				ID:             uuid.New(), // Generate a new UUID
				OrganizationID: "org-demo-001",
				Name:           "Standard Purchase Order Workflow",
				Description:    "Default workflow for purchase order approvals",
				DocumentType:   "purchase_order",
				EntityType:     "purchase_order",
				Version:        1,
				IsActive:       true,
				IsDefault:      true,
				CreatedBy:      "user-admin-001",
			},
			stages: purchaseOrderStages,
		},
		{
			workflow: models.Workflow{
				ID:             uuid.New(), // Generate a new UUID
				OrganizationID: "org-demo-001",
				Name:           "Standard GRN Workflow",
				Description:    "Default workflow for goods receipt note processing",
				DocumentType:   "grn",
				EntityType:     "grn",
				Version:        1,
				IsActive:       true,
				IsDefault:      true,
				CreatedBy:      "user-admin-001",
			},
			stages: grnStages,
		},
		{
			workflow: models.Workflow{
				ID:             uuid.New(), // Generate a new UUID
				OrganizationID: "org-demo-001",
				Name:           "Standard Payment Voucher Workflow",
				Description:    "Default workflow for payment voucher approvals",
				DocumentType:   "payment_voucher",
				EntityType:     "payment_voucher",
				Version:        1,
				IsActive:       true,
				IsDefault:      true,
				CreatedBy:      "user-admin-001",
			},
			stages: paymentVoucherStages,
		},
	}

	for _, wf := range workflows {
		// Set stages using the model method
		if err := wf.workflow.SetStages(wf.stages); err != nil {
			result.Error = fmt.Errorf("failed to set stages for workflow %s: %w", wf.workflow.Name, err)
			return result, result.Error
		}

		// Check if workflow exists by ID or org+entity_type+is_default combination
		var existingWorkflow models.Workflow
		err := tx.WithContext(ctx).Where("id = ? OR (organization_id = ? AND entity_type = ? AND is_default = ?)", 
			wf.workflow.ID, wf.workflow.OrganizationID, wf.workflow.EntityType, wf.workflow.IsDefault).First(&existingWorkflow).Error
		
		if err == nil {
			// Workflow exists, update it
			err = tx.WithContext(ctx).Model(&existingWorkflow).Updates(map[string]interface{}{
				"name":        wf.workflow.Name,
				"description": wf.workflow.Description,
				"stages":      wf.workflow.Stages,
				"is_active":   wf.workflow.IsActive,
				"updated_at":  time.Now(),
			}).Error
			
			if err != nil {
				result.Error = fmt.Errorf("failed to update workflow %s: %w", wf.workflow.Name, err)
				return result, result.Error
			}
			result.Updated++
		} else {
			// Workflow doesn't exist, create it
			err = tx.WithContext(ctx).Create(&wf.workflow).Error
			if err != nil {
				result.Error = fmt.Errorf("failed to create workflow %s: %w", wf.workflow.Name, err)
				return result, result.Error
			}
			result.Created++
		}
	}

	result.Duration = time.Since(startTime)
	return result, nil
}

// seedSampleData creates sample business documents for development
func (s *DatabaseSeeder) seedSampleData(ctx context.Context, tx *gorm.DB) (*SeedResult, error) {
	startTime := time.Now()
	result := &SeedResult{Entity: "sample_data"}

	// Note: Sample data is now provided by the consolidated SQL seed migration
	// The migration 002_consolidated_seed_data.up.sql contains comprehensive seed data
	
	s.logger.Println("📋 Sample data is provided by SQL migration 002_consolidated_seed_data.up.sql")
	
	result.Duration = time.Since(startTime)
	result.Created = 1 // Indicate we have sample data available
	return result, nil
}

// logSeedingSummary logs a summary of all seeding operations
func (s *DatabaseSeeder) logSeedingSummary(results []*SeedResult, totalDuration time.Duration) {
	s.logger.Println("📊 Database Seeding Summary:")
	s.logger.Println(strings.Repeat("=", 50))

	totalCreated := 0
	totalUpdated := 0
	totalSkipped := 0

	for _, result := range results {
		s.logger.Printf("  %-20s: %3d created, %3d updated, %3d skipped (%v)",
			result.Entity, result.Created, result.Updated, result.Skipped, result.Duration)
		
		totalCreated += result.Created
		totalUpdated += result.Updated
		totalSkipped += result.Skipped
	}

	s.logger.Println(strings.Repeat("=", 50))
	s.logger.Printf("  %-20s: %3d created, %3d updated, %3d skipped (%v)",
		"TOTAL", totalCreated, totalUpdated, totalSkipped, totalDuration)
	s.logger.Printf("✅ Database seeding completed successfully in %v", totalDuration)
}

// IsSeeded checks if the database has been seeded
func (s *DatabaseSeeder) IsSeeded(ctx context.Context) (bool, error) {
	// Check if we have any users
	var userCount int64
	err := s.db.WithContext(ctx).Model(&models.User{}).Count(&userCount).Error
	if err != nil {
		return false, fmt.Errorf("failed to count users: %w", err)
	}

	// Check if we have any organizations
	var orgCount int64
	err = s.db.WithContext(ctx).Model(&models.Organization{}).Count(&orgCount).Error
	if err != nil {
		return false, fmt.Errorf("failed to count organizations: %w", err)
	}

	// Consider seeded if we have at least one user and one organization
	return userCount > 0 && orgCount > 0, nil
}

// GetSeedingStats returns statistics about seeded data
func (s *DatabaseSeeder) GetSeedingStats(ctx context.Context) (map[string]int64, error) {
	stats := make(map[string]int64)

	entities := map[string]interface{}{
		"users":         &models.User{},
		"organizations": &models.Organization{},
		"vendors":       &models.Vendor{},
		"categories":    &models.Category{},
		"workflows":     &models.Workflow{},
	}

	for name, model := range entities {
		var count int64
		err := s.db.WithContext(ctx).Model(model).Count(&count).Error
		if err != nil {
			return nil, fmt.Errorf("failed to count %s: %w", name, err)
		}
		stats[name] = count
	}

	return stats, nil
}

// stringPtr returns a pointer to a string
func stringPtr(s string) *string {
	return &s
}

// intPtr returns a pointer to an int
func intPtr(i int) *int {
	return &i
}