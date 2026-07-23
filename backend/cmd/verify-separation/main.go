package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/models"
	"gorm.io/gorm"
)

func main() {
	// Load environment variables
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Note: .env file not found, using environment variables")
	}

	// Set default values if not provided
	if os.Getenv("DB_PORT") == "" {
		os.Setenv("DB_PORT", "5432")
	}
	if os.Getenv("DB_HOST") == "" {
		os.Setenv("DB_HOST", "localhost")
	}
	if os.Getenv("DB_USER") == "" {
		os.Setenv("DB_USER", "postgres")
	}
	if os.Getenv("DB_SSL_MODE") == "" {
		os.Setenv("DB_SSL_MODE", "disable")
	}

	// Initialize database connection
	config.InitDatabase()

	db := config.DB
	if db == nil {
		log.Fatal("Database connection is nil")
	}

	log.Println("🔍 Verifying multi-tenant data separation...")
	log.Println("=" + fmt.Sprintf("%50s", "="))

	// Verify organizations exist
	var orgs []models.Organization
	if err := db.Find(&orgs).Error; err != nil {
		log.Fatalf("Failed to fetch organizations: %v", err)
	}

	log.Printf("📊 Found %d organizations:", len(orgs))
	for _, org := range orgs {
		log.Printf("  - %s (%s) - %s", org.Name, org.ID, org.Tier)
	}
	log.Println()

	// Verify super admin memberships
	verifySuperAdminMemberships(db)

	// Verify data separation for each organization
	for _, org := range orgs {
		if org.ID == "org-demo-001" || org.ID == "org-acme-001" {
			verifyOrganizationData(db, org)
		}
	}

	log.Println("✅ Multi-tenant data separation verification completed!")
}

func verifySuperAdminMemberships(db *gorm.DB) {
	log.Printf("👑 Verifying Super Admin (admin@liyali.com) memberships:")
	log.Println("-" + fmt.Sprintf("%48s", "-"))

	// Check if super admin exists
	var superAdmin models.User
	if err := db.Where("email = ?", "admin@liyali.com").First(&superAdmin).Error; err != nil {
		log.Printf("  ❌ Super admin not found: %v", err)
		return
	}

	log.Printf("  ✅ Super admin found: %s (%s)", superAdmin.Name, superAdmin.Email)
	log.Printf("  📋 Is Super Admin: %t", superAdmin.IsSuperAdmin)
	if superAdmin.CurrentOrganizationID != nil {
		log.Printf("  🏢 Current Organization: %s", *superAdmin.CurrentOrganizationID)
	} else {
		log.Printf("  🏢 Current Organization: None")
	}

	// Check memberships in both organizations
	var memberships []models.OrganizationMember
	if err := db.Where("user_id = ? AND active = ?", superAdmin.ID, true).Find(&memberships).Error; err != nil {
		log.Printf("  ❌ Failed to fetch memberships: %v", err)
		return
	}

	log.Printf("  🤝 Active Memberships: %d", len(memberships))
	for _, membership := range memberships {
		var org models.Organization
		if err := db.Where("id = ?", membership.OrganizationID).First(&org).Error; err == nil {
			log.Printf("    - %s (%s) as %s", org.Name, org.ID, membership.Role)
		} else {
			log.Printf("    - %s as %s", membership.OrganizationID, membership.Role)
		}
	}

	// Verify can access both organizations
	expectedOrgs := []string{"org-demo-001", "org-acme-001"}
	for _, expectedOrgID := range expectedOrgs {
		var membership models.OrganizationMember
		if err := db.Where("user_id = ? AND organization_id = ? AND active = ?", 
			superAdmin.ID, expectedOrgID, true).First(&membership).Error; err != nil {
			log.Printf("  ❌ Missing membership in %s", expectedOrgID)
		} else {
			log.Printf("  ✅ Has membership in %s", expectedOrgID)
		}
	}

	log.Println()
}

func verifyOrganizationData(db *gorm.DB, org models.Organization) {
	log.Printf("🏢 Verifying data for: %s (%s)", org.Name, org.ID)
	log.Println("-" + fmt.Sprintf("%48s", "-"))

	// Count users in this organization
	var userCount int64
	db.Model(&models.User{}).Where("current_organization_id = ?", org.ID).Count(&userCount)
	log.Printf("  👥 Users: %d", userCount)

	// Count organization members
	var memberCount int64
	db.Model(&models.OrganizationMember{}).Where("organization_id = ? AND active = ?", org.ID, true).Count(&memberCount)
	log.Printf("  🤝 Active Members: %d", memberCount)

	// Count categories
	var categoryCount int64
	db.Model(&models.Category{}).Where("organization_id = ? AND active = ?", org.ID, true).Count(&categoryCount)
	log.Printf("  📂 Categories: %d", categoryCount)

	// List categories
	var categories []models.Category
	db.Where("organization_id = ? AND active = ?", org.ID, true).Find(&categories)
	for _, cat := range categories {
		log.Printf("    - %s", cat.Name)
	}

	// Count requisitions
	var reqCount int64
	db.Model(&models.Requisition{}).Where("organization_id = ?", org.ID).Count(&reqCount)
	log.Printf("  📄 Requisitions: %d", reqCount)

	// List requisitions with status
	var requisitions []models.Requisition
	db.Where("organization_id = ?", org.ID).Find(&requisitions)
	statusCounts := make(map[string]int)
	for _, req := range requisitions {
		statusCounts[req.Status]++
		log.Printf("    - %s (%s) - %s - $%.2f", req.DocumentNumber, req.Status, req.Title, req.TotalAmount)
	}

	// Show status breakdown
	log.Printf("  📊 Requisition Status Breakdown:")
	for status, count := range statusCounts {
		log.Printf("    - %s: %d", status, count)
	}

	// Count budgets
	var budgetCount int64
	db.Model(&models.Budget{}).Where("organization_id = ?", org.ID).Count(&budgetCount)
	log.Printf("  💰 Budgets: %d", budgetCount)

	// List budgets
	var budgets []models.Budget
	db.Where("organization_id = ?", org.ID).Find(&budgets)
	for _, budget := range budgets {
		utilization := (budget.AllocatedAmount / budget.TotalBudget) * 100
		log.Printf("    - %s (%s) - $%.2f/$%.2f (%.1f%% utilized)", 
			budget.BudgetCode, budget.Department, budget.AllocatedAmount, budget.TotalBudget, utilization)
	}

	log.Println()
}