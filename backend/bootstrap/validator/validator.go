package validator

import (
	"context"
	"fmt"
	"log"
	"strings"

	"gorm.io/gorm"
)

// DatabaseValidator handles database schema validation
type DatabaseValidator struct {
	db     *gorm.DB
	logger *log.Logger
}

// New creates a new database validator
func New(db *gorm.DB, logger *log.Logger) *DatabaseValidator {
	return &DatabaseValidator{
		db:     db,
		logger: logger,
	}
}

// RequiredTables defines the core tables that must exist
var RequiredTables = []string{
	"users",
	"organizations",
	"organization_settings",
	"organization_members",
	"vendors",
	"categories",
	"requisitions",
	"budgets",
	"purchase_orders",
	"payment_vouchers",
	"goods_received_notes",
	// "approval_tasks", // DEPRECATED: Legacy approval system, kept in DB for backward compatibility
	"notifications",
	"audit_logs",
}

// RequiredIndexes defines critical indexes that must exist
var RequiredIndexes = []string{
	"idx_users_email",
	"idx_organizations_active",
	"idx_requisitions_organization_id",
	"idx_vendors_active",
}

// ValidateSchemaReadiness checks if the database is ready for operations
func (v *DatabaseValidator) ValidateSchemaReadiness(ctx context.Context) error {
	v.logger.Println("🔍 Validating database schema readiness...")

	// Check if database exists and is accessible
	var dbName string
	err := v.db.WithContext(ctx).Raw("SELECT current_database()").Scan(&dbName).Error
	if err != nil {
		return fmt.Errorf("failed to query current database: %w", err)
	}

	v.logger.Printf("📊 Connected to database: %s", dbName)

	// Check PostgreSQL version
	var version string
	err = v.db.WithContext(ctx).Raw("SELECT version()").Scan(&version).Error
	if err != nil {
		return fmt.Errorf("failed to query PostgreSQL version: %w", err)
	}

	v.logger.Printf("📊 PostgreSQL version: %s", strings.Split(version, " ")[1])

	return nil
}

// VerifyMigrations ensures all required tables exist
func (v *DatabaseValidator) VerifyMigrations(ctx context.Context) error {
	v.logger.Println("🔍 Verifying database migrations...")

	missingTables := []string{}

	for _, tableName := range RequiredTables {
		exists, err := v.tableExists(ctx, tableName)
		if err != nil {
			return fmt.Errorf("failed to check table %s: %w", tableName, err)
		}

		if !exists {
			missingTables = append(missingTables, tableName)
		}
	}

	if len(missingTables) > 0 {
		return fmt.Errorf("missing required tables: %v. Please run migrations first", missingTables)
	}

	v.logger.Printf("✅ All %d required tables exist", len(RequiredTables))
	return nil
}

// VerifySchemaIntegrity performs comprehensive schema validation
func (v *DatabaseValidator) VerifySchemaIntegrity(ctx context.Context) error {
	v.logger.Println("🔍 Verifying schema integrity...")

	// Verify table structures
	if err := v.verifyTableStructures(ctx); err != nil {
		return fmt.Errorf("table structure validation failed: %w", err)
	}

	// Verify foreign key constraints
	if err := v.verifyForeignKeys(ctx); err != nil {
		return fmt.Errorf("foreign key validation failed: %w", err)
	}

	// Verify critical indexes
	if err := v.verifyIndexes(ctx); err != nil {
		return fmt.Errorf("index validation failed: %w", err)
	}

	// Verify triggers
	if err := v.verifyTriggers(ctx); err != nil {
		return fmt.Errorf("trigger validation failed: %w", err)
	}

	v.logger.Println("✅ Schema integrity verification completed")
	return nil
}

// QuickSchemaCheck performs a lightweight schema validation for health checks
func (v *DatabaseValidator) QuickSchemaCheck(ctx context.Context) error {
	// Just check a few critical tables
	criticalTables := []string{"users", "organizations", "vendors"}
	
	for _, tableName := range criticalTables {
		exists, err := v.tableExists(ctx, tableName)
		if err != nil {
			return fmt.Errorf("failed to check critical table %s: %w", tableName, err)
		}
		if !exists {
			return fmt.Errorf("critical table %s does not exist", tableName)
		}
	}

	return nil
}

// tableExists checks if a table exists in the database
func (v *DatabaseValidator) tableExists(ctx context.Context, tableName string) (bool, error) {
	var count int64
	err := v.db.WithContext(ctx).Raw(`
		SELECT COUNT(*) 
		FROM information_schema.tables 
		WHERE table_schema = 'public' 
		AND table_name = ?
	`, tableName).Scan(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// verifyTableStructures checks that tables have expected columns
func (v *DatabaseValidator) verifyTableStructures(ctx context.Context) error {
	// Define critical columns for key tables
	criticalColumns := map[string][]string{
		"users": {"id", "email", "name", "password", "role", "active"},
		"organizations": {"id", "name", "slug", "active"},
		"vendors": {"id", "vendor_code", "name", "active"},
		"requisitions": {"id", "organization_id", "document_number", "requester_id", "status"},
	}

	for tableName, columns := range criticalColumns {
		for _, columnName := range columns {
			exists, err := v.columnExists(ctx, tableName, columnName)
			if err != nil {
				return fmt.Errorf("failed to check column %s.%s: %w", tableName, columnName, err)
			}
			if !exists {
				return fmt.Errorf("critical column %s.%s does not exist", tableName, columnName)
			}
		}
	}

	return nil
}

// columnExists checks if a column exists in a table
func (v *DatabaseValidator) columnExists(ctx context.Context, tableName, columnName string) (bool, error) {
	var count int64
	err := v.db.WithContext(ctx).Raw(`
		SELECT COUNT(*) 
		FROM information_schema.columns 
		WHERE table_schema = 'public' 
		AND table_name = ? 
		AND column_name = ?
	`, tableName, columnName).Scan(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// verifyForeignKeys checks critical foreign key constraints
func (v *DatabaseValidator) verifyForeignKeys(ctx context.Context) error {
	// Check some critical foreign keys
	criticalFKs := []struct {
		table      string
		constraint string
	}{
		{"organizations", "fk_organizations_creator"},
		{"requisitions", "fk_requisitions_organization"},
		{"requisitions", "fk_requisitions_requester"},
		{"vendors", "fk_vendors_created_by"},
	}

	for _, fk := range criticalFKs {
		exists, err := v.constraintExists(ctx, fk.table, fk.constraint)
		if err != nil {
			return fmt.Errorf("failed to check constraint %s.%s: %w", fk.table, fk.constraint, err)
		}
		if !exists {
			v.logger.Printf("⚠️  Foreign key constraint %s.%s does not exist", fk.table, fk.constraint)
			// Don't fail for missing FKs, just warn
		}
	}

	return nil
}

// constraintExists checks if a constraint exists
func (v *DatabaseValidator) constraintExists(ctx context.Context, tableName, constraintName string) (bool, error) {
	var count int64
	err := v.db.WithContext(ctx).Raw(`
		SELECT COUNT(*) 
		FROM information_schema.table_constraints 
		WHERE table_schema = 'public' 
		AND table_name = ? 
		AND constraint_name = ?
	`, tableName, constraintName).Scan(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// verifyIndexes checks that critical indexes exist
func (v *DatabaseValidator) verifyIndexes(ctx context.Context) error {
	for _, indexName := range RequiredIndexes {
		exists, err := v.indexExists(ctx, indexName)
		if err != nil {
			return fmt.Errorf("failed to check index %s: %w", indexName, err)
		}
		if !exists {
			v.logger.Printf("⚠️  Index %s does not exist", indexName)
			// Don't fail for missing indexes, just warn
		}
	}

	return nil
}

// indexExists checks if an index exists
func (v *DatabaseValidator) indexExists(ctx context.Context, indexName string) (bool, error) {
	var count int64
	err := v.db.WithContext(ctx).Raw(`
		SELECT COUNT(*) 
		FROM pg_indexes 
		WHERE schemaname = 'public' 
		AND indexname = ?
	`, indexName).Scan(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// verifyTriggers checks that critical triggers exist
func (v *DatabaseValidator) verifyTriggers(ctx context.Context) error {
	// Check for the update timestamp trigger function
	var count int64
	err := v.db.WithContext(ctx).Raw(`
		SELECT COUNT(*) 
		FROM information_schema.routines 
		WHERE routine_schema = 'public' 
		AND routine_name = 'update_updated_at_column'
		AND routine_type = 'FUNCTION'
	`).Scan(&count).Error

	if err != nil {
		return fmt.Errorf("failed to check trigger function: %w", err)
	}

	if count == 0 {
		v.logger.Printf("⚠️  Trigger function update_updated_at_column does not exist")
	}

	return nil
}

// GetTableStats returns statistics about database tables
func (v *DatabaseValidator) GetTableStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Get table count
	var tableCount int64
	err := v.db.WithContext(ctx).Raw(`
		SELECT COUNT(*) 
		FROM information_schema.tables 
		WHERE table_schema = 'public'
	`).Scan(&tableCount).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get table count: %w", err)
	}

	stats["table_count"] = tableCount

	// Get row counts for key tables
	rowCounts := make(map[string]int64)
	for _, tableName := range RequiredTables {
		var count int64
		err := v.db.WithContext(ctx).Raw(fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)).Scan(&count).Error
		if err != nil {
			v.logger.Printf("⚠️  Failed to get row count for %s: %v", tableName, err)
			continue
		}
		rowCounts[tableName] = count
	}

	stats["row_counts"] = rowCounts
	return stats, nil
}