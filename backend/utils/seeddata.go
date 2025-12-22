package utils

import (
	"log"

	"github.com/liyali/liyali-gateway/models"
	"gorm.io/gorm"
)

// SeedTestUsers creates test users for development
func SeedTestUsers(db *gorm.DB) error {
	testUsers := []models.User{
		{
			ID:     "user-admin-001",
			Email:  "admin@liyali.com",
			Name:   "Admin User",
			Role:   "admin",
			Active: true,
		},
		{
			ID:     "user-approver-001",
			Email:  "approver@liyali.com",
			Name:   "John Approver",
			Role:   "approver",
			Active: true,
		},
		{
			ID:     "user-requester-001",
			Email:  "requester@liyali.com",
			Name:   "Jane Requester",
			Role:   "requester",
			Active: true,
		},
		{
			ID:     "user-finance-001",
			Email:  "finance@liyali.com",
			Name:   "Finance Officer",
			Role:   "finance",
			Active: true,
		},
		{
			ID:     "user-viewer-001",
			Email:  "viewer@liyali.com",
			Name:   "Viewer User",
			Role:   "viewer",
			Active: true,
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

// SeedTestVendors creates test vendors for development
func SeedTestVendors(db *gorm.DB) error {
	testVendors := []models.Vendor{
		{
			ID:         "vendor-001",
			VendorCode: "VND-001",
			Name:       "ABC Supplies Ltd",
			Email:      "contact@abcsupplies.com",
			Phone:      "+1-555-0101",
			Country:    "United States",
			City:       "New York",
			BankAccount: "1234567890",
			TaxID:      "12-3456789",
			Active:     true,
		},
		{
			ID:         "vendor-002",
			VendorCode: "VND-002",
			Name:       "Global Tech Solutions",
			Email:      "sales@globaltech.com",
			Phone:      "+1-555-0102",
			Country:    "United States",
			City:       "San Francisco",
			BankAccount: "0987654321",
			TaxID:      "98-7654321",
			Active:     true,
		},
		{
			ID:         "vendor-003",
			VendorCode: "VND-003",
			Name:       "Premium Services Inc",
			Email:      "info@premiumservices.com",
			Phone:      "+1-555-0103",
			Country:    "Canada",
			City:       "Toronto",
			BankAccount: "5555666677",
			TaxID:      "55-5555555",
			Active:     true,
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

// SeedDatabase seeds all test data
func SeedDatabase(db *gorm.DB) error {
	log.Println("🌱 Seeding database with test data...")

	if err := SeedTestUsers(db); err != nil {
		log.Printf("Error seeding users: %v", err)
		return err
	}

	if err := SeedTestVendors(db); err != nil {
		log.Printf("Error seeding vendors: %v", err)
		return err
	}

	log.Println("✓ Database seeding completed")
	return nil
}
