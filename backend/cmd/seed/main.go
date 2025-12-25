package main

import (
	"fmt"
	"log"
	"os"

	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/utils"
	"gorm.io/gorm"
)

// User structure for seeding
type SeedUser struct {
	Email    string
	Password string
	Name     string
	Role     string
}

func main() {
	// Initialize database connection
	config.InitConfig()
	config.ConnectDatabase()

	db := config.DB

	// List of test users to seed (matching DEMO_USERS structure)
	users := []SeedUser{
		{
			Email:    "requester@liyali.com",
			Password: "password123",
			Name:     "John Requester",
			Role:     "requester",
		},
		{
			Email:    "manager@liyali.com",
			Password: "password123",
			Name:     "Sarah Manager",
			Role:     "approver",
		},
		{
			Email:    "finance@liyali.com",
			Password: "password123",
			Name:     "James Finance",
			Role:     "finance",
		},
		{
			Email:    "director@liyali.com",
			Password: "password123",
			Name:     "Paul Director",
			Role:     "approver",
		},
		{
			Email:    "cfo@liyali.com",
			Password: "password123",
			Name:     "Michelle CFO",
			Role:     "finance",
		},
		{
			Email:    "compliance@liyali.com",
			Password: "password123",
			Name:     "David Compliance",
			Role:     "viewer",
		},
		{
			Email:    "admin@liyali.com",
			Password: "password123",
			Name:     "Admin User",
			Role:     "admin",
		},
	}

	log.Println("Starting database seeding...")

	for _, seedUser := range users {
		// Check if user already exists
		var existingUser models.User
		if err := db.Where("email = ?", seedUser.Email).First(&existingUser).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				// User doesn't exist, create it
				hashedPassword, err := utils.HashPassword(seedUser.Password)
				if err != nil {
					log.Printf("Error hashing password for %s: %v", seedUser.Email, err)
					continue
				}

				user := models.User{
					Email:    seedUser.Email,
					Name:     seedUser.Name,
					Password: hashedPassword,
					Role:     seedUser.Role,
					Active:   true,
				}

				if err := db.Create(&user).Error; err != nil {
					log.Printf("Error creating user %s: %v", seedUser.Email, err)
					continue
				}

				log.Printf("✓ Created user: %s (%s) with role: %s", seedUser.Email, seedUser.Name, seedUser.Role)
			} else {
				log.Printf("Error checking user %s: %v", seedUser.Email, err)
				continue
			}
		} else {
			log.Printf("✓ User already exists: %s", seedUser.Email)
		}
	}

	log.Println("Database seeding completed successfully!")
	os.Exit(0)
}
