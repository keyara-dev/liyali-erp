//go:build ignore

package main

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/liyali/liyali-gateway/bootstrap/seeder"
	"github.com/liyali/liyali-gateway/config"
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
	if config.DB == nil {
		log.Fatalf("Failed to initialize database connection")
	}

	// Create seeder
	logger := log.New(os.Stdout, "[SEEDER] ", log.LstdFlags)
	dbSeeder := seeder.New(config.DB, logger)

	// Run seeding
	ctx := context.Background()
	if err := dbSeeder.SeedAll(ctx); err != nil {
		log.Fatalf("Seeding failed: %v", err)
	}

	log.Println("✅ Database seeding completed successfully!")
}