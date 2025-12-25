package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/models"
	"gorm.io/gorm"
)

func main() {
	// Parse command line flags
	migrate := flag.Bool("migrate", false, "Run migration to multi-tenant")
	rollback := flag.Bool("rollback", false, "Rollback migration (requires backup)")
	verify := flag.Bool("verify", false, "Verify migration completed successfully")

	flag.Parse()

	// Initialize database
	config.InitDatabase()

	if *migrate {
		runMigration()
	} else if *verify {
		verifyMigration()
	} else if *rollback {
		log.Println("Rollback requires manual database restore from backup")
		log.Println("Restore from backup and then remove organization-related columns")
	} else {
		log.Println("Usage:")
		log.Println("  -migrate : Run migration to multi-tenant")
		log.Println("  -verify  : Verify migration completed")
		log.Println("  -rollback: Instructions for rollback")
	}
}

func runMigration() {
	log.Println("\n=== STARTING MULTI-TENANCY MIGRATION ===\n")

	// Step 1: Create default "Legacy" organization
	log.Println("Step 1: Creating default organization...")

	var firstUser models.User
	if err := config.DB.Where("active = ?", true).First(&firstUser).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Fatalf("No active users found. Please create at least one user before migration.")
		}
		log.Fatalf("Failed to find user: %v", err)
	}

	defaultOrg := &models.Organization{
		ID:          uuid.New().String(),
		Name:        "Legacy System",
		Slug:        "legacy-system",
		Description: "Default organization for existing data",
		Active:      true,
		Tier:        "pro",
		CreatedBy:   firstUser.ID,
	}

	if err := config.DB.Create(defaultOrg).Error; err != nil {
		log.Fatalf("Failed to create default organization: %v", err)
	}

	log.Printf("✓ Created organization: %s (ID: %s)\n", defaultOrg.Name, defaultOrg.ID)

	// Step 2: Create default settings
	log.Println("\nStep 2: Creating organization settings...")

	settings := &models.OrganizationSettings{
		ID:             uuid.New().String(),
		OrganizationID: defaultOrg.ID,
		Currency:       "USD",
		FiscalYearStart: 1,
	}

	if err := config.DB.Create(settings).Error; err != nil {
		log.Printf("Warning: Failed to create settings: %v\n", err)
	} else {
		log.Println("✓ Created organization settings")
	}

	// Step 3: Add all active users as members
	log.Println("\nStep 3: Adding existing users to organization...")

	var users []models.User
	if err := config.DB.Where("active = ?", true).Find(&users).Error; err != nil {
		log.Fatalf("Failed to fetch users: %v", err)
	}

	now := time.Now()
	for _, user := range users {
		member := &models.OrganizationMember{
			ID:             uuid.New().String(),
			OrganizationID: defaultOrg.ID,
			UserID:         user.ID,
			Role:           user.Role,
			Active:         true,
			JoinedAt:       &now,
		}

		if err := config.DB.Create(member).Error; err != nil {
			log.Printf("Warning: Failed to add user %s: %v\n", user.Email, err)
			continue
		}

		// Set as current organization
		if err := config.DB.Model(&user).Update("current_organization_id", defaultOrg.ID).Error; err != nil {
			log.Printf("Warning: Failed to set current org for user %s: %v\n", user.Email, err)
		}
	}

	log.Printf("✓ Added %d users to organization\n", len(users))

	// Step 4: Migrate existing documents
	log.Println("\nStep 4: Migrating existing documents...")

	migrations := []struct {
		name   string
		model  interface{}
		count  *int64
	}{
		{"requisitions", &models.Requisition{}, &int64(0)},
		{"budgets", &models.Budget{}, &int64(0)},
		{"purchase_orders", &models.PurchaseOrder{}, &int64(0)},
		{"payment_vouchers", &models.PaymentVoucher{}, &int64(0)},
		{"goods_received_notes", &models.GoodsReceivedNote{}, &int64(0)},
		{"categories", &models.Category{}, &int64(0)},
		{"vendors", &models.Vendor{}, &int64(0)},
		{"approval_tasks", &models.ApprovalTask{}, &int64(0)},
		{"notifications", &models.Notification{}, &int64(0)},
	}

	for _, m := range migrations {
		result := config.DB.Model(m.model).
			Where("organization_id IS NULL OR organization_id = ''").
			Update("organization_id", defaultOrg.ID)

		if result.Error != nil {
			log.Printf("Error migrating %s: %v\n", m.name, result.Error)
			continue
		}

		log.Printf("✓ Migrated %d %s\n", result.RowsAffected, m.name)
	}

	// Step 5: Verify migration
	log.Println("\nStep 5: Verifying migration...")

	if err := verifyMigration(); err != nil {
		log.Fatalf("Migration verification failed: %v", err)
	}

	log.Println("\n=== MIGRATION COMPLETED SUCCESSFULLY ===\n")
	log.Println("All existing data has been migrated to the 'Legacy System' organization.")
	log.Println("Users can now switch between organizations using the workspace switcher.")
}

func verifyMigration() error {
	log.Println("\nVerifying migration...")

	checks := []struct {
		name  string
		check func() error
	}{
		{
			"Organizations exist",
			func() error {
				var count int64
				config.DB.Model(&models.Organization{}).Count(&count)
				if count == 0 {
					return fmt.Errorf("no organizations found")
				}
				log.Printf("  ✓ Found %d organizations\n", count)
				return nil
			},
		},
		{
			"Organization members exist",
			func() error {
				var count int64
				config.DB.Model(&models.OrganizationMember{}).Count(&count)
				if count == 0 {
					return fmt.Errorf("no organization members found")
				}
				log.Printf("  ✓ Found %d organization members\n", count)
				return nil
			},
		},
		{
			"No orphaned requisitions",
			func() error {
				var count int64
				config.DB.Model(&models.Requisition{}).
					Where("organization_id IS NULL OR organization_id = ''").
					Count(&count)
				if count > 0 {
					return fmt.Errorf("found %d orphaned requisitions", count)
				}
				log.Printf("  ✓ No orphaned requisitions\n")
				return nil
			},
		},
		{
			"No orphaned budgets",
			func() error {
				var count int64
				config.DB.Model(&models.Budget{}).
					Where("organization_id IS NULL OR organization_id = ''").
					Count(&count)
				if count > 0 {
					return fmt.Errorf("found %d orphaned budgets", count)
				}
				log.Printf("  ✓ No orphaned budgets\n")
				return nil
			},
		},
		{
			"All users have current organization",
			func() error {
				var count int64
				config.DB.Model(&models.User{}).
					Where("active = ? AND current_organization_id IS NULL", true).
					Count(&count)
				if count > 0 {
					return fmt.Errorf("found %d users without current organization", count)
				}
				log.Printf("  ✓ All active users have current organization\n")
				return nil
			},
		},
	}

	for _, check := range checks {
		if err := check.check(); err != nil {
			return fmt.Errorf("%s: %v", check.name, err)
		}
	}

	log.Println("\n✓ All verification checks passed!\n")
	return nil
}
