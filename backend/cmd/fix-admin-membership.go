//go:build ignore

package main

import (
	"log"
	"os"
	"time"

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

	// Create admin membership in org-demo-001
	log.Println("🔧 Creating admin membership in org-demo-001...")
	
	now := time.Now()
	adminMember := models.OrganizationMember{
		ID:             "member-admin-demo",
		OrganizationID: "org-demo-001",
		UserID:         "user-admin-001",
		Role:           "admin",
		Active:         true,
		JoinedAt:       &now,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	// Check if already exists
	var existing models.OrganizationMember
	err = config.DB.Where("organization_id = ? AND user_id = ?", "org-demo-001", "user-admin-001").First(&existing).Error
	
	if err != nil {
		// Doesn't exist, create it
		err = config.DB.Create(&adminMember).Error
		if err != nil {
			log.Fatalf("Failed to create admin membership: %v", err)
		}
		log.Println("✅ Created admin membership in org-demo-001")
	} else {
		// Already exists, update it
		err = config.DB.Model(&existing).Updates(map[string]interface{}{
			"role":       "admin",
			"active":     true,
			"updated_at": now,
		}).Error
		if err != nil {
			log.Fatalf("Failed to update admin membership: %v", err)
		}
		log.Println("✅ Updated admin membership in org-demo-001")
	}

	log.Println("🎉 Admin can now access org-demo-001!")
}