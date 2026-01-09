package main

import (
	"flag"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/database/seeders"
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

	// Parse command line flags
	var (
		multiTenant = flag.Bool("multi-tenant", false, "Seed multi-tenant data with proper workspace separation")
		cleanup     = flag.Bool("cleanup", false, "Clean up existing multi-tenant test data")
		help        = flag.Bool("help", false, "Show help message")
	)
	flag.Parse()

	if *help {
		showHelp()
		return
	}

	// Initialize database connection
	config.InitDatabase()

	db := config.DB
	if db == nil {
		log.Fatal("Database connection is nil")
	}

	log.Println("🚀 Starting database seeding operations...")

	// Handle cleanup operation
	if *cleanup {
		log.Println("🧹 Cleaning up multi-tenant test data...")
		if err := seeders.CleanupMultiTenantData(db); err != nil {
			log.Fatalf("Failed to cleanup multi-tenant data: %v", err)
		}
		log.Println("✅ Cleanup completed successfully!")
		return
	}

	// Handle multi-tenant seeding
	if *multiTenant {
		log.Println("🌱 Seeding multi-tenant data with proper workspace separation...")
		if err := seeders.SeedMultiTenantData(db); err != nil {
			log.Fatalf("Failed to seed multi-tenant data: %v", err)
		}
		log.Println("✅ Multi-tenant seeding completed successfully!")
		return
	}

	// Default behavior - show help
	showHelp()
}

func showHelp() {
	log.Println("Database Seeding Tool")
	log.Println("====================")
	log.Println("")
	log.Println("Usage:")
	log.Println("  go run cmd/seed/main.go [options]")
	log.Println("")
	log.Println("Options:")
	log.Println("  --multi-tenant    Seed multi-tenant data with proper workspace separation")
	log.Println("  --cleanup         Clean up existing multi-tenant test data")
	log.Println("  --help            Show this help message")
	log.Println("")
	log.Println("Examples:")
	log.Println("  go run cmd/seed/main.go --multi-tenant")
	log.Println("  go run cmd/seed/main.go --cleanup")
	log.Println("")
	log.Println("Description:")
	log.Println("  This tool creates properly separated test data for multiple organizations")
	log.Println("  to ensure that each workspace has its own isolated dataset for testing")
	log.Println("  the multi-tenant functionality of the application.")
	log.Println("")
	log.Println("  The multi-tenant seeder creates:")
	log.Println("  - Separate users for each organization")
	log.Println("  - Organization-specific categories")
	log.Println("  - Isolated requisitions and budgets")
	log.Println("  - Proper organization memberships")
	log.Println("")
}