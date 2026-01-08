package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/models"
)

func main() {
	// Load environment variables
	err := godotenv.Load(".env")
	if err != nil {
		log.Printf("Warning: .env file not found: %v", err)
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

	// Initialize database
	config.InitDatabase()

	// Check organization members
	log.Println("🔍 Checking organization members...")
	var members []models.OrganizationMember
	err = config.DB.Find(&members).Error
	if err != nil {
		log.Fatalf("Failed to query organization members: %v", err)
	}

	log.Printf("Found %d organization members:", len(members))
	for _, member := range members {
		log.Printf("  ID: %s, OrgID: %s, UserID: %s, Role: %s, Active: %t", 
			member.ID, member.OrganizationID, member.UserID, member.Role, member.Active)
	}

	// Check workflows
	log.Println("\n🔍 Checking workflows...")
	var workflows []models.Workflow
	err = config.DB.Find(&workflows).Error
	if err != nil {
		log.Fatalf("Failed to query workflows: %v", err)
	}

	log.Printf("Found %d workflows:", len(workflows))
	for _, workflow := range workflows {
		log.Printf("  ID: %s, Name: %s, EntityType: %s, IsDefault: %t, IsActive: %t", 
			workflow.ID, workflow.Name, workflow.EntityType, workflow.IsDefault, workflow.IsActive)
	}
}