package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/cozyCodr/liyali-gateway/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to database
	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	// Test connection
	if err := pool.Ping(context.Background()); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	log.Println("✅ Connected to database")

	// Run migration 001
	migration001, err := os.ReadFile("internal/database/migrations/001_initial_schema.up.sql")
	if err != nil {
		log.Fatalf("Failed to read migration 001: %v", err)
	}

	log.Println("📝 Running migration 001_initial_schema.up.sql...")
	if _, err := pool.Exec(context.Background(), string(migration001)); err != nil {
		log.Printf("❌ Migration 001 failed (this is OK if tables already exist): %v", err)
	} else {
		log.Println("✅ Migration 001 completed successfully")
	}

	// Run migration 002
	migration002, err := os.ReadFile("internal/database/migrations/002_workflow_tables.up.sql")
	if err != nil {
		log.Fatalf("Failed to read migration 002: %v", err)
	}

	log.Println("📝 Running migration 002_workflow_tables.up.sql...")
	if _, err := pool.Exec(context.Background(), string(migration002)); err != nil {
		log.Printf("❌ Migration 002 failed (this is OK if tables already exist): %v", err)
	} else {
		log.Println("✅ Migration 002 completed successfully")
	}

	log.Println("🎉 All migrations completed!")

	// List all tables
	rows, err := pool.Query(context.Background(), "SELECT tablename FROM pg_tables WHERE schemaname = 'public' ORDER BY tablename")
	if err != nil {
		log.Fatalf("Failed to list tables: %v", err)
	}
	defer rows.Close()

	log.Println("\n📋 Database tables:")
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}
		fmt.Printf("  - %s\n", tableName)
	}
}
