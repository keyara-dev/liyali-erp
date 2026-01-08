package main

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/liyali/liyali-gateway/bootstrap"
	"github.com/liyali/liyali-gateway/bootstrap/seeder"
	"github.com/liyali/liyali-gateway/config"
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

	// Force seeding regardless of environment
	bootstrapConfig := bootstrap.DefaultBootstrapConfig()
	bootstrapConfig.Environment = "development" // Force development mode
	bootstrapConfig.SkipSeeding = false         // Force seeding

	ctx := context.Background()
	
	// Run only the seeding phase
	log.Println("🌱 Running database seeding...")
	
	// Create seeder directly and run it
	seeder := seeder.New(config.DB, log.Default())
	err = seeder.SeedAll(ctx)
	
	if err != nil {
		log.Fatalf("Seeding failed: %v", err)
	}

	log.Println("✅ Database seeding completed successfully!")
	
	// Show stats
	stats, err := seeder.GetSeedingStats(ctx)
	if err != nil {
		log.Printf("Warning: Could not get seeding stats: %v", err)
	} else {
		log.Println("📊 Seeding Statistics:")
		for entity, count := range stats {
			log.Printf("  %s: %d", entity, count)
		}
	}
}