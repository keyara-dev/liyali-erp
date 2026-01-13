package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
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

	// Build connection string
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_SSL_MODE"),
	)

	// Connect to database
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Get migrations directory and mode
	migrationsDir := "database/migrations"
	resetMode := false // Default: normal migration mode (skip cleanup)
	
	// Check command line arguments
	for i, arg := range os.Args[1:] {
		if arg == "--reset" {
			resetMode = true
		} else if arg == "--help" || arg == "-h" {
			fmt.Println("Database Migration Tool")
			fmt.Println("")
			fmt.Println("Usage:")
			fmt.Println("  go run database/migrate_all.go [options] [migrations_directory]")
			fmt.Println("")
			fmt.Println("Options:")
			fmt.Println("  --reset    Include cleanup migrations (000_*) for full database reset")
			fmt.Println("  --help     Show this help message")
			fmt.Println("")
			fmt.Println("Examples:")
			fmt.Println("  go run database/migrate_all.go                    # Run normal migrations (skip cleanup)")
			fmt.Println("  go run database/migrate_all.go --reset            # Run all migrations including cleanup")
			fmt.Println("  make db-migrate                                   # Run normal migrations via Makefile")
			fmt.Println("  make db-reset                                     # Run reset migrations via Makefile")
			return
		} else if i == 0 && !strings.HasPrefix(arg, "--") {
			// First non-flag argument is migrations directory
			migrationsDir = arg
		}
	}

	// Read all .up.sql files from migrations directory
	files, err := ioutil.ReadDir(migrationsDir)
	if err != nil {
		log.Fatalf("Failed to read migrations directory: %v", err)
	}

	// Filter and sort migration files
	var migrationFiles []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".up.sql") {
			filename := file.Name()
			
			// In normal mode, skip cleanup migrations (000_*)
			// In reset mode, include all migrations
			if !resetMode && strings.HasPrefix(filename, "000_") {
				fmt.Printf("⏭️  Skipping cleanup migration: %s (use --reset to include)\n", filename)
				continue
			}
			
			migrationFiles = append(migrationFiles, filename)
		}
	}
	sort.Strings(migrationFiles)

	if len(migrationFiles) == 0 {
		if resetMode {
			fmt.Println("No migration files found in", migrationsDir)
		} else {
			fmt.Println("No non-cleanup migration files found in", migrationsDir)
		}
		return
	}

	if resetMode {
		fmt.Printf("🔄 RESET MODE: Found %d migration files to run (including cleanup)\n", len(migrationFiles))
	} else {
		fmt.Printf("Found %d migration files to run\n", len(migrationFiles))
	}

	// Create migrations tracking table (but handle reset mode differently)
	if !resetMode {
		createMigrationsTable(db)
	}

	// Run each migration
	migrationsTableCreated := !resetMode
	for _, filename := range migrationFiles {
		// In reset mode, create migrations table after schema is created
		if resetMode && !migrationsTableCreated && strings.Contains(filename, "schema") {
			// Wait until after schema migration to create tracking table
		} else if resetMode && !migrationsTableCreated && strings.Contains(filename, "seed") {
			// Create tracking table before seed data
			createMigrationsTable(db)
			migrationsTableCreated = true
		}

		// Check if migration has been applied (only if tracking table exists)
		if migrationsTableCreated && hasBeenApplied(db, filename) {
			fmt.Printf("⏭️  Skipping %s (already applied)\n", filename)
			continue
		}

		filePath := filepath.Join(migrationsDir, filename)
		
		// Read migration file
		content, err := ioutil.ReadFile(filePath)
		if err != nil {
			log.Fatalf("Failed to read migration file %s: %v", filename, err)
		}

		// Execute migration
		fmt.Printf("🔄 Running migration: %s\n", filename)
		
		if _, err := db.Exec(string(content)); err != nil {
			log.Fatalf("Migration %s failed: %v", filename, err)
		}

		// Create tracking table after schema migration in reset mode
		if resetMode && !migrationsTableCreated && strings.Contains(filename, "schema") {
			createMigrationsTable(db)
			migrationsTableCreated = true
		}

		// Mark as applied (only if tracking table exists)
		if migrationsTableCreated {
			markAsApplied(db, filename)
		}
		fmt.Printf("✅ Migration %s completed successfully!\n", filename)
	}

	fmt.Println("\n🎉 All migrations completed successfully!")
}

func createMigrationsTable(db *sql.DB) {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			id SERIAL PRIMARY KEY,
			filename VARCHAR(255) UNIQUE NOT NULL,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`
	
	if _, err := db.Exec(query); err != nil {
		log.Fatalf("Failed to create migrations table: %v", err)
	}
}

func hasBeenApplied(db *sql.DB, filename string) bool {
	var count int
	query := "SELECT COUNT(*) FROM schema_migrations WHERE filename = $1"
	err := db.QueryRow(query, filename).Scan(&count)
	if err != nil {
		log.Printf("Warning: Failed to check migration status for %s: %v", filename, err)
		return false
	}
	return count > 0
}

func markAsApplied(db *sql.DB, filename string) {
	query := "INSERT INTO schema_migrations (filename) VALUES ($1) ON CONFLICT (filename) DO NOTHING"
	if _, err := db.Exec(query, filename); err != nil {
		log.Printf("Warning: Failed to mark migration %s as applied: %v", filename, err)
	}
}