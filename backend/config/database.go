package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/liyali/liyali-gateway/bootstrap"
	db "github.com/liyali/liyali-gateway/database/sqlc"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	DB      *gorm.DB
	PgxDB   *pgxpool.Pool
	Queries *db.Queries
)

// InitDatabase initializes both GORM and pgx database connections with proper bootstrap
func InitDatabase() {
	// Prefer DATABASE_URL if set; otherwise build DSN from individual DB_* vars.
	// This keeps the config simple for PaaS deployments (Railway, Fly.io, etc.)
	// that inject a single DATABASE_URL secret.
	databaseURL := os.Getenv("DATABASE_URL")

	var dsn, pgxDSN string
	if databaseURL != "" {
		dsn = databaseURL
		pgxDSN = databaseURL
	} else {
		dsn = fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			os.Getenv("DB_HOST"),
			os.Getenv("DB_PORT"),
			os.Getenv("DB_USER"),
			os.Getenv("DB_PASSWORD"),
			os.Getenv("DB_NAME"),
			os.Getenv("DB_SSL_MODE"),
		)
		pgxDSN = fmt.Sprintf(
			"postgres://%s:%s@%s:%s/%s?sslmode=%s",
			os.Getenv("DB_USER"),
			os.Getenv("DB_PASSWORD"),
			os.Getenv("DB_HOST"),
			os.Getenv("DB_PORT"),
			os.Getenv("DB_NAME"),
			os.Getenv("DB_SSL_MODE"),
		)
	}

	// Initialize GORM connection (for existing functionality)
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		log.Fatalf("Failed to connect to database with GORM: %v", err)
	}

	// Initialize pgx connection pool (for new enhanced features)
	PgxDB, err = pgxpool.New(context.Background(), pgxDSN)
	if err != nil {
		log.Fatalf("Failed to connect to database with pgx: %v", err)
	}

	// Test pgx connection
	if err := PgxDB.Ping(context.Background()); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Initialize sqlc typed query interface
	Queries = db.New(PgxDB)

	log.Println("✓ Database connected successfully (GORM + pgx + sqlc)")

	// Initialize bootstrap system
	bootstrapConfig := bootstrap.DefaultBootstrapConfig()
	bootstrapConfig.Environment = os.Getenv("APP_ENV")
	if bootstrapConfig.Environment == "" {
		bootstrapConfig.Environment = "development"
	}

	// For production with slow networks, skip validation if SKIP_DB_VALIDATION is set
	skipValidation := os.Getenv("SKIP_DB_VALIDATION") == "true"
	
	// Increase timeouts for production/remote databases
	if bootstrapConfig.Environment == "production" || bootstrapConfig.Environment == "staging" {
		bootstrapConfig.ValidationTimeout = time.Minute * 5  // 5 minutes for very slow networks
		bootstrapConfig.MigrationTimeout = time.Minute * 10  // 10 minutes for migrations
		bootstrapConfig.SkipSeeding = os.Getenv("ENABLE_SEEDING") != "true"
	}

	// Skip seeding in production unless explicitly enabled
	if bootstrapConfig.Environment == "production" {
		bootstrapConfig.SkipSeeding = os.Getenv("ENABLE_SEEDING") != "true"
	}

	// Skip bootstrap entirely if validation is disabled (for slow networks)
	if skipValidation {
		log.Println("⚠️  Skipping database validation (SKIP_DB_VALIDATION=true)")
		log.Println("✓ Database ready (validation skipped)")
		return
	}

	// Create bootstrapper and run bootstrap process
	bootstrapper := bootstrap.NewBootstrapper(DB, bootstrapConfig, log.Default())
	
	ctx := context.Background()
	result := bootstrapper.Bootstrap(ctx)
	
	if !result.Success {
		log.Printf("⚠️  Database bootstrap failed at phase %s: %v", result.Phase, result.Error)
		log.Println("⚠️  Continuing anyway - database may not be fully validated")
		// Don't fatal - allow app to start even if validation fails
		return
	}

	log.Printf("✓ Database bootstrap completed successfully in %v", result.Duration)
}

// MigrateModels is now deprecated - use SQL migrations instead
func MigrateModels() {
	log.Println("⚠️  MigrateModels is deprecated - using SQL migrations via bootstrap system")
	log.Println("✓ Run migrations manually using: go run database/run_migration.go database/migrations/001_create_complete_schema.up.sql")
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return DB
}
