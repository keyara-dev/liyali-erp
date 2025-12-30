package integration

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/models"
	"gorm.io/gorm"
)

// TestDatabase holds test database connection and utilities
type TestDatabase struct {
	DB *gorm.DB
}

// SetupTestDatabase initializes a test database connection
func SetupTestDatabase(t *testing.T) *TestDatabase {
	// Check if database environment variables are set
	if os.Getenv("DB_HOST") == "" {
		t.Skip("Database environment variables not set - skipping integration test")
		return nil
	}
	
	// Initialize the database using the existing config
	config.InitDatabase()
	
	if config.DB == nil {
		t.Skip("Database not initialized - skipping integration test")
		return nil
	}

	return &TestDatabase{
		DB: config.DB,
	}
}

// CleanupTestDatabase cleans up test database resources
func (td *TestDatabase) CleanupTestDatabase(t *testing.T) {
	// Database cleanup is handled by the config package
}

// CreateTestUser creates a test user for integration tests
func (td *TestDatabase) CreateTestUser(t *testing.T, email, name, role string) *models.User {
	user := &models.User{
		ID:       uuid.New().String(),
		Email:    email,
		Name:     name,
		Role:     role,
		Password: "$2a$10$test.hash.for.testing.purposes.only", // Test password hash
		Active:   true,
	}

	if err := td.DB.Create(user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	return user
}

// CreateTestOrganization creates a test organization
func (td *TestDatabase) CreateTestOrganization(t *testing.T, name, slug string) *models.Organization {
	org := &models.Organization{
		ID:          uuid.New().String(),
		Name:        name,
		Slug:        slug,
		Description: fmt.Sprintf("Test organization: %s", name),
		Active:      true,
		Tier:        "standard",
	}

	if err := td.DB.Create(org).Error; err != nil {
		t.Fatalf("Failed to create test organization: %v", err)
	}

	return org
}

// CleanupTestData removes all test data from the database
func (td *TestDatabase) CleanupTestData(t *testing.T) {
	// Clean up in reverse dependency order
	tables := []string{
		"audit_logs",
		"notifications", 
		"user_organization_roles",
		"organization_roles",
		"account_lockouts",
		"login_attempts",
		"password_resets",
		"sessions",
		"email_verifications",
		"organization_members",
		"organizations",
		"users",
	}

	for _, table := range tables {
		if err := td.DB.Exec(fmt.Sprintf("DELETE FROM %s", table)).Error; err != nil {
			t.Logf("Warning: Failed to clean up table %s: %v", table, err)
		}
	}
}

// BeginTransaction starts a database transaction for test isolation
func (td *TestDatabase) BeginTransaction(t *testing.T) *gorm.DB {
	tx := td.DB.Begin()
	if tx.Error != nil {
		t.Fatalf("Failed to begin transaction: %v", tx.Error)
	}
	return tx
}

// RollbackTransaction rolls back a database transaction
func (td *TestDatabase) RollbackTransaction(t *testing.T, tx *gorm.DB) {
	if err := tx.Rollback().Error; err != nil {
		t.Logf("Warning: Failed to rollback transaction: %v", err)
	}
}

// WaitForCondition waits for a condition to be met with timeout
func WaitForCondition(t *testing.T, condition func() bool, timeout time.Duration, message string) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf("Timeout waiting for condition: %s", message)
}

// CreateTestContext creates a context with timeout for tests
func CreateTestContext(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}

// Helper functions

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Test data generators

// GenerateTestEmail generates a unique test email
func GenerateTestEmail(prefix string) string {
	return fmt.Sprintf("%s+%s@test.example.com", prefix, uuid.New().String()[:8])
}

// GenerateTestSlug generates a unique test slug
func GenerateTestSlug(prefix string) string {
	return fmt.Sprintf("%s-%s", prefix, uuid.New().String()[:8])
}

// LogTestStep logs a test step for better debugging
func LogTestStep(t *testing.T, step string) {
	t.Logf("=== Test Step: %s ===", step)
}

// LogTestInfo logs test information
func LogTestInfo(t *testing.T, format string, args ...interface{}) {
	t.Logf("INFO: "+format, args...)
}

// LogTestError logs test error information
func LogTestError(t *testing.T, format string, args ...interface{}) {
	log.Printf("ERROR: "+format, args...)
}