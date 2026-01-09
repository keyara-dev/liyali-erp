package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/database/seeders"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	// Initialize database connection
	config.InitDatabase()

	if len(os.Args) > 1 && os.Args[1] == "cleanup" {
		// Cleanup existing test data
		if err := seeders.CleanupApprovalTestData(config.DB); err != nil {
			log.Fatalf("Failed to cleanup test data: %v", err)
		}
		log.Println("✅ Test data cleanup completed!")
		return
	}

	// Seed approval test data
	if err := seeders.SeedApprovalTestData(config.DB); err != nil {
		log.Fatalf("Failed to seed approval test data: %v", err)
	}

	log.Println("🎉 Approval test data seeding completed successfully!")
	log.Println("")
	log.Println("📋 Test Scenarios Created:")
	log.Println("1. REQ-001: Office Furniture Purchase (Pending Manager Approval)")
	log.Println("2. REQ-002: Software Licenses Renewal (Pending Finance Approval)")
	log.Println("3. REQ-003: Emergency IT Equipment (Ready for Submission)")
	log.Println("4. REQ-004: Marketing Campaign Materials (Fully Approved)")
	log.Println("5. REQ-005: Training and Development (Rejected by Finance)")
	log.Println("")
	log.Println("🧪 You can now test:")
	log.Println("- Approval workflows")
	log.Println("- Rejection workflows")
	log.Println("- Auto-generation services")
	log.Println("- Action history tracking")
	log.Println("- Approval chain visualization")
}