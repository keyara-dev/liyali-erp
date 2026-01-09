package bootstrap_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/liyali/liyali-gateway/bootstrap"
	"github.com/liyali/liyali-gateway/bootstrap/circuit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TestBootstrapIntegration tests the complete bootstrap process
func TestBootstrapIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup test database connection
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	// Create bootstrap configuration for testing
	config := &bootstrap.BootstrapConfig{
		Environment:        "test",
		SkipSeeding:       false,
		SeedRetryAttempts: 2,
		SeedRetryDelay:    time.Millisecond * 100,
		CircuitBreakerConfig: circuit.Config{
			MaxFailures: 3,
			Timeout:     time.Second * 5,
			Interval:    time.Second * 10,
		},
		ValidationTimeout: time.Second * 10,
		MigrationTimeout:  time.Second * 30,
	}

	// Create bootstrapper
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	bootstrapper := bootstrap.NewBootstrapper(db, config, logger)

	// Run bootstrap
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	result := bootstrapper.Bootstrap(ctx)

	// Verify results
	assert.True(t, result.Success, "Bootstrap should succeed")
	assert.Equal(t, bootstrap.PhaseComplete, result.Phase)
	assert.NoError(t, result.Error)
	assert.Greater(t, result.Duration, time.Duration(0))
	assert.NotEmpty(t, result.Metrics)
}

// TestBootstrapWithMissingTables tests bootstrap behavior when tables don't exist
func TestBootstrapWithMissingTables(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup test database connection (without migrations)
	db := setupTestDBWithoutMigrations(t)
	defer cleanupTestDB(t, db)

	config := bootstrap.DefaultBootstrapConfig()
	config.Environment = "test"

	bootstrapper := bootstrap.NewBootstrapper(db, config, nil)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	result := bootstrapper.Bootstrap(ctx)

	// Should fail at migration verification phase
	assert.False(t, result.Success)
	assert.Equal(t, bootstrap.PhaseMigrate, result.Phase)
	assert.Error(t, result.Error)
	assert.Contains(t, result.Error.Error(), "missing required tables")
}

// TestBootstrapIdempotency tests that bootstrap can be run multiple times safely
func TestBootstrapIdempotency(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	config := bootstrap.DefaultBootstrapConfig()
	config.Environment = "test"

	bootstrapper := bootstrap.NewBootstrapper(db, config, nil)
	ctx := context.Background()

	// Run bootstrap first time
	result1 := bootstrapper.Bootstrap(ctx)
	require.True(t, result1.Success)

	// Run bootstrap second time - should still succeed
	result2 := bootstrapper.Bootstrap(ctx)
	assert.True(t, result2.Success)
	assert.Equal(t, bootstrap.PhaseComplete, result2.Phase)
}

// TestBootstrapHealthCheck tests the health check functionality
func TestBootstrapHealthCheck(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	bootstrapper := bootstrap.NewBootstrapper(db, nil, nil)

	// Bootstrap first
	ctx := context.Background()
	result := bootstrapper.Bootstrap(ctx)
	require.True(t, result.Success)

	// Test health check
	err := bootstrapper.HealthCheck(ctx)
	assert.NoError(t, err)
}

// TestBootstrapMetrics tests metrics collection
func TestBootstrapMetrics(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	bootstrapper := bootstrap.NewBootstrapper(db, nil, nil)

	// Get initial metrics
	metrics := bootstrapper.GetMetrics()
	assert.NotEmpty(t, metrics)
	assert.Contains(t, metrics, "circuit_breaker_state")
	assert.Contains(t, metrics, "db_connections_open")
}

// TestBootstrapCancellation tests context cancellation
func TestBootstrapCancellation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	bootstrapper := bootstrap.NewBootstrapper(db, nil, nil)

	// Create a context that will be cancelled quickly
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*10)
	defer cancel()

	result := bootstrapper.Bootstrap(ctx)

	// Should fail due to context cancellation
	assert.False(t, result.Success)
	assert.Error(t, result.Error)
}

// setupTestDB creates a test database connection with migrations
func setupTestDB(t *testing.T) *gorm.DB {
	// Use test database configuration
	dsn := getTestDSN(t)
	
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	// Run migrations (simplified for testing)
	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id VARCHAR(255) PRIMARY KEY,
			email VARCHAR(255) UNIQUE NOT NULL,
			name VARCHAR(255) NOT NULL,
			password VARCHAR(255) NOT NULL,
			role VARCHAR(50) NOT NULL DEFAULT 'requester',
			active BOOLEAN DEFAULT true,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		
		CREATE TABLE IF NOT EXISTS organizations (
			id VARCHAR(255) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			slug VARCHAR(255) UNIQUE NOT NULL,
			active BOOLEAN DEFAULT true,
			created_by VARCHAR(255),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		
		CREATE TABLE IF NOT EXISTS vendors (
			id VARCHAR(255) PRIMARY KEY,
			vendor_code VARCHAR(100) UNIQUE NOT NULL,
			name VARCHAR(255) NOT NULL,
			active BOOLEAN DEFAULT true,
			created_by VARCHAR(255),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		
		CREATE TABLE IF NOT EXISTS categories (
			id VARCHAR(255) PRIMARY KEY,
			organization_id VARCHAR(255) NOT NULL,
			name VARCHAR(255) NOT NULL,
			active BOOLEAN DEFAULT true,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		
		CREATE TABLE IF NOT EXISTS organization_members (
			id VARCHAR(255) PRIMARY KEY,
			organization_id VARCHAR(255) NOT NULL,
			user_id VARCHAR(255) NOT NULL,
			role VARCHAR(50) NOT NULL,
			active BOOLEAN DEFAULT true,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(organization_id, user_id)
		);
		
		-- Add other required tables for testing
		CREATE TABLE IF NOT EXISTS requisitions (id VARCHAR(255) PRIMARY KEY, organization_id VARCHAR(255));
		CREATE TABLE IF NOT EXISTS budgets (id VARCHAR(255) PRIMARY KEY, organization_id VARCHAR(255));
		CREATE TABLE IF NOT EXISTS purchase_orders (id VARCHAR(255) PRIMARY KEY, organization_id VARCHAR(255));
		CREATE TABLE IF NOT EXISTS payment_vouchers (id VARCHAR(255) PRIMARY KEY, organization_id VARCHAR(255));
		CREATE TABLE IF NOT EXISTS goods_received_notes (id VARCHAR(255) PRIMARY KEY, organization_id VARCHAR(255));
		CREATE TABLE IF NOT EXISTS approval_tasks (id VARCHAR(255) PRIMARY KEY, organization_id VARCHAR(255));
		CREATE TABLE IF NOT EXISTS notifications (id VARCHAR(255) PRIMARY KEY, organization_id VARCHAR(255));
		CREATE TABLE IF NOT EXISTS audit_logs (id VARCHAR(255) PRIMARY KEY, document_id VARCHAR(255));
		
		-- Create some indexes
		CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
		CREATE INDEX IF NOT EXISTS idx_organizations_active ON organizations(active);
		CREATE INDEX IF NOT EXISTS idx_requisitions_organization_id ON requisitions(organization_id);
		CREATE INDEX IF NOT EXISTS idx_vendors_active ON vendors(active);
	`).Error
	require.NoError(t, err)

	return db
}

// setupTestDBWithoutMigrations creates a test database connection without running migrations
func setupTestDBWithoutMigrations(t *testing.T) *gorm.DB {
	dsn := getTestDSN(t)
	
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	return db
}

// cleanupTestDB cleans up test database
func cleanupTestDB(t *testing.T, db *gorm.DB) {
	// Drop test tables
	tables := []string{
		"organization_members", "categories", "vendors", "organizations", "users",
		"requisitions", "budgets", "purchase_orders", "payment_vouchers",
		"goods_received_notes", "approval_tasks", "notifications", "audit_logs",
	}

	for _, table := range tables {
		db.Exec("DROP TABLE IF EXISTS " + table + " CASCADE")
	}
}

// getTestDSN returns the test database DSN
func getTestDSN(t *testing.T) string {
	host := os.Getenv("TEST_DB_HOST")
	if host == "" {
		host = "localhost"
	}

	port := os.Getenv("TEST_DB_PORT")
	if port == "" {
		port = "5432"
	}

	user := os.Getenv("TEST_DB_USER")
	if user == "" {
		user = "postgres"
	}

	password := os.Getenv("TEST_DB_PASSWORD")
	if password == "" {
		password = "postgres"
	}

	dbname := os.Getenv("TEST_DB_NAME")
	if dbname == "" {
		dbname = "liyali_gateway_test"
	}

	sslmode := os.Getenv("TEST_DB_SSL_MODE")
	if sslmode == "" {
		sslmode = "disable"
	}

	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)
}

// BenchmarkBootstrap benchmarks the bootstrap process
func BenchmarkBootstrap(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping benchmark in short mode")
	}

	db := setupTestDB(&testing.T{})
	defer cleanupTestDB(&testing.T{}, db)

	config := bootstrap.DefaultBootstrapConfig()
	config.Environment = "test"

	bootstrapper := bootstrap.NewBootstrapper(db, config, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx := context.Background()
		result := bootstrapper.Bootstrap(ctx)
		if !result.Success {
			b.Fatalf("Bootstrap failed: %v", result.Error)
		}
	}
}