package seeder

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/liyali/liyali-gateway/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
	requiredTables := []string{"users", "organizations", "vendors", "categories"}
	
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
			ID:           "user-admin-001",
			Email:        "admin@liyali.com",
			Name:         "System Administrator",
			Password:     "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi", // password
			Role:         "admin",
			Active:       true,
			IsSuperAdmin: true,
		},
		{
			ID:       "user-approver-001",
			Email:    "approver@liyali.com",
			Name:     "John Approver",
			Password: "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi",
			Role:     "approver",
			Active:   true,
		},
		{
			ID:       "user-requester-001",
			Email:    "requester@liyali.com",
			Name:     "Jane Requester",
			Password: "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi",
			Role:     "requester",
			Active:   true,
		},
		{
			ID:       "user-finance-001",
			Email:    "finance@liyali.com",
			Name:     "Finance Officer",
			Password: "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi",
			Role:     "finance",
			Active:   true,
		},
	}

	for _, user := range users {
		// Use UPSERT (ON CONFLICT DO UPDATE)
		err := tx.WithContext(ctx).Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "email"}},
			DoUpdates: clause.AssignmentColumns([]string{"name", "role", "active", "updated_at"}),
		}).Create(&user).Error

		if err != nil {
			result.Error = fmt.Errorf("failed to upsert user %s: %w", user.Email, err)
			return result, result.Error
		}

		// Check if it was created or updated
		if tx.RowsAffected > 0 {
			var existingUser models.User
			if err := tx.WithContext(ctx).Where("email = ?", user.Email).First(&existingUser).Error; err == nil {
				if existingUser.CreatedAt.After(startTime.Add(-time.Second)) {
					result.Created++
				} else {
					result.Updated++
				}
			}
		} else {
			result.Skipped++
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
		err := tx.WithContext(ctx).Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "slug"}},
			DoUpdates: clause.AssignmentColumns([]string{"name", "description", "active", "tier", "updated_at"}),
		}).Create(&org).Error

		if err != nil {
			result.Error = fmt.Errorf("failed to upsert organization %s: %w", org.Slug, err)
			return result, result.Error
		}

		if tx.RowsAffected > 0 {
			result.Created++
		} else {
			result.Skipped++
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
		err := tx.WithContext(ctx).Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "organization_id"}, {Name: "user_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"role", "active", "updated_at"}),
		}).Create(&member).Error

		if err != nil {
			result.Error = fmt.Errorf("failed to upsert member %s-%s: %w", member.OrganizationID, member.UserID, err)
			return result, result.Error
		}

		if tx.RowsAffected > 0 {
			result.Created++
		} else {
			result.Skipped++
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

	for _, vendor := range vendors {
		err := tx.WithContext(ctx).Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "vendor_code"}},
			DoUpdates: clause.AssignmentColumns([]string{"name", "email", "phone", "country", "city", "active", "updated_at"}),
		}).Create(&vendor).Error

		if err != nil {
			result.Error = fmt.Errorf("failed to upsert vendor %s: %w", vendor.VendorCode, err)
			return result, result.Error
		}

		if tx.RowsAffected > 0 {
			result.Created++
		} else {
			result.Skipped++
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
		err := tx.WithContext(ctx).Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "organization_id"}, {Name: "name"}},
			DoUpdates: clause.AssignmentColumns([]string{"description", "active", "updated_at"}),
		}).Create(&category).Error

		if err != nil {
			result.Error = fmt.Errorf("failed to upsert category %s: %w", category.Name, err)
			return result, result.Error
		}

		if tx.RowsAffected > 0 {
			result.Created++
		} else {
			result.Skipped++
		}
	}

	result.Duration = time.Since(startTime)
	return result, nil
}

// seedSampleData creates sample business documents for development
func (s *DatabaseSeeder) seedSampleData(ctx context.Context, tx *gorm.DB) (*SeedResult, error) {
	startTime := time.Now()
	result := &SeedResult{Entity: "sample_data"}

	// Only seed sample data in development environment
	// This would be controlled by environment variables
	// For now, we'll skip this to keep seeding minimal
	
	result.Duration = time.Since(startTime)
	result.Skipped = 1 // Indicate we skipped sample data
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