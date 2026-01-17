package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	_ "github.com/lib/pq"
)

func main() {
	log.Println("🚀 Starting simple database migration...")

	// Get database connection string
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_SSL_MODE"),
	)

	// Connect to database
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("❌ Failed to ping database: %v", err)
	}
	log.Println("✅ Database connection successful")

	// Check if tables already exist
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name IN ('users', 'organizations', 'vendors')").Scan(&count)
	if err != nil {
		log.Fatalf("❌ Failed to check existing tables: %v", err)
	}

	if count >= 3 {
		log.Println("✅ Database tables already exist, skipping migration")
		return
	}

	log.Println("📋 Running database migration...")

	// Read migration file
	migrationFile := "/app/database/migrations/001_init_system.up.sql"
	content, err := ioutil.ReadFile(migrationFile)
	if err != nil {
		log.Fatalf("❌ Failed to read migration file: %v", err)
	}

	// Split into individual statements (simple approach)
	statements := strings.Split(string(content), ";")
	
	successCount := 0
	for i, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" || strings.HasPrefix(stmt, "--") {
			continue
		}

		_, err := db.Exec(stmt)
		if err != nil {
			// Log error but continue with next statement
			log.Printf("⚠️  Statement %d failed (continuing): %v", i+1, err)
			continue
		}
		successCount++
	}

	log.Printf("✅ Migration completed: %d statements executed successfully", successCount)

	// Verify migration
	err = db.QueryRow("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name IN ('users', 'organizations', 'vendors', 'requisitions', 'budgets')").Scan(&count)
	if err != nil {
		log.Fatalf("❌ Failed to verify migration: %v", err)
	}

	if count >= 5 {
		log.Println("✅ Migration verification successful")
	} else {
		log.Printf("⚠️  Migration verification: only %d/5 expected tables found", count)
	}

	log.Println("🎉 Migration process completed!")
}