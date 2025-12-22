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

	log.Println("✓ Database migrations completed")
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return DB
}
