package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

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

	// Check command line arguments
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run database/simple_migration.go <migration_file.sql>")
	}

	migrationFile := os.Args[1]
	
	// Check if file exists
	if _, err := os.Stat(migrationFile); os.IsNotExist(err) {
		log.Fatalf("Migration file does not exist: %s", migrationFile)
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

	// Read migration file
	content, err := ioutil.ReadFile(migrationFile)
	if err != nil {
		log.Fatalf("Failed to read migration file: %v", err)
	}

	// Execute migration
	fmt.Printf("Running migration: %s\n", filepath.Base(migrationFile))
	
	if _, err := db.Exec(string(content)); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	fmt.Println("✅ Migration completed successfully!")
}