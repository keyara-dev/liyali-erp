package config

import (
	"fmt"
	"log"
	"os"

	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/utils"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDatabase initializes the PostgreSQL database connection
func InitDatabase() {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_SSL_MODE"),
	)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("✓ Database connected successfully")

	// Auto migrate models
	MigrateModels()

	// Seed test data if in development
	if os.Getenv("APP_ENV") != "production" {
		if err := utils.SeedDatabase(DB); err != nil {
			log.Printf("Warning: Failed to seed database: %v", err)
		}
	}
}

// MigrateModels creates/updates all database tables
func MigrateModels() {
	tables := []interface{}{
		&models.User{},

		// Organization tables (must come before business tables)
		&models.Organization{},
		&models.OrganizationSettings{},
		&models.OrganizationMember{},
		&models.OrganizationDepartment{},

		// Business tables (now with organization_id)
		&models.Category{},
		&models.CategoryBudgetCode{},
		&models.Requisition{},
		&models.Budget{},
		&models.PurchaseOrder{},
		&models.PaymentVoucher{},
		&models.GoodsReceivedNote{},
		&models.Vendor{},
		&models.ApprovalTask{},
		&models.AuditLog{},
		&models.Notification{},
	}

	for _, table := range tables {
		if err := DB.AutoMigrate(table); err != nil {
			log.Fatalf("Migration failed for %v: %v", table, err)
		}
	}

	// Create unique constraint on organization_members (organization_id, user_id)
	if !DB.Migrator().HasConstraint(&models.OrganizationMember{}, "uk_org_user") {
		DB.Migrator().CreateConstraint(&models.OrganizationMember{}, "uk_org_user")
	}

	// Create index on organization_id for better query performance
	indexes := []struct {
		model     interface{}
		indexName string
		columns   string
	}{
		{&models.Requisition{}, "idx_req_org_id", "organization_id"},
		{&models.Budget{}, "idx_budget_org_id", "organization_id"},
		{&models.PurchaseOrder{}, "idx_po_org_id", "organization_id"},
		{&models.PaymentVoucher{}, "idx_pv_org_id", "organization_id"},
		{&models.GoodsReceivedNote{}, "idx_grn_org_id", "organization_id"},
		{&models.Category{}, "idx_cat_org_id", "organization_id"},
		{&models.Vendor{}, "idx_vendor_org_id", "organization_id"},
		{&models.ApprovalTask{}, "idx_approval_org_id", "organization_id"},
		{&models.Notification{}, "idx_notif_org_id", "organization_id"},
	}

	for _, idx := range indexes {
		if !DB.Migrator().HasIndex(idx.model, idx.indexName) {
			DB.Migrator().CreateIndex(idx.model, idx.indexName)
		}
	}

	log.Println("✓ Database migrations completed")
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return DB
}
