package main

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/liyali/liyali-gateway/bootstrap/seeder"
	"github.com/liyali/liyali-gateway/config"
)

func init() {
	// Load environment variables
	err := godotenv.Load(".env")
	if err != nil && os.Getenv("APP_ENV") == "" {
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
}

func main() {
	// Initialize database
	config.InitDatabase()

	// Create logger
	logger := log.New(os.Stdout, "[SEEDER] ", log.LstdFlags)

	// Create seeder
	dbSeeder := seeder.New(config.DB, logger)

	// Run seeding
	ctx := context.Background()
	err := dbSeeder.SeedAll(ctx)
	if err != nil {
		logger.Fatalf("Seeding failed: %v", err)
	}

	logger.Println("✅ Database seeding completed successfully!")
}