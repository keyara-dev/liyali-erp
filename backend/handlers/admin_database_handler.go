package handlers

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/utils"
)

// connectionID is the fixed ID for the primary database connection
const connectionID = "primary-postgresql"

// validateConnectionID checks that the connection ID matches the primary connection
func validateConnectionID(c *fiber.Ctx) error {
	id := c.Params("id")
	if id != connectionID {
		return utils.SendNotFound(c, "Database connection not found")
	}
	return nil
}

// getConnectionInfo returns real connection metadata from environment
func getConnectionInfo() map[string]interface{} {
	now := time.Now().Format(time.RFC3339)

	info := map[string]interface{}{
		"id":                 connectionID,
		"name":               "Primary PostgreSQL",
		"type":               "postgresql",
		"host":               os.Getenv("DB_HOST"),
		"port":               os.Getenv("DB_PORT"),
		"database":           os.Getenv("DB_NAME"),
		"username":           os.Getenv("DB_USER"),
		"ssl_mode":           os.Getenv("DB_SSL_MODE"),
		"is_primary":         true,
		"is_replica":         false,
		"status":             "unknown",
		"active_connections": 0,
		"idle_connections":   0,
		"max_connections":    0,
		"last_health_check":  now,
		"created_at":         now,
		"updated_at":         now,
	}

	sqlDB, err := config.DB.DB()
	if err == nil {
		stats := sqlDB.Stats()
		info["active_connections"] = stats.InUse
		info["idle_connections"] = stats.Idle
		info["open_connections"] = stats.OpenConnections
		info["max_connections"] = stats.MaxOpenConnections
		info["wait_count"] = stats.WaitCount
		info["wait_duration_ms"] = stats.WaitDuration.Milliseconds()

		if pingErr := sqlDB.Ping(); pingErr == nil {
			info["status"] = "connected"
		} else {
			info["status"] = "error"
		}
	} else {
		info["status"] = "error"
	}

	return info
}

// GetDatabaseConnections returns the list of database connections with real pool stats
func GetDatabaseConnections(c *fiber.Ctx) error {
	connections := []map[string]interface{}{getConnectionInfo()}
	return utils.SendSimpleSuccess(c, connections, "Database connections retrieved successfully")
}

// GetDatabaseConnection returns a single database connection by ID
func GetDatabaseConnection(c *fiber.Ctx) error {
	if err := validateConnectionID(c); err != nil {
		return err
	}
	return utils.SendSimpleSuccess(c, getConnectionInfo(), "Database connection retrieved successfully")
}

// TestDatabaseConnection tests the database connection with a real ping
func TestDatabaseConnection(c *fiber.Ctx) error {
	if err := validateConnectionID(c); err != nil {
		return err
	}

	result := map[string]interface{}{
		"connection_id": connectionID,
		"status":        "failed",
		"latency_ms":    0,
		"message":       "Connection test failed",
		"tested_at":     time.Now().Format(time.RFC3339),
	}

	sqlDB, err := config.DB.DB()
	if err != nil {
		log.Printf("Error getting database connection for test: %v", err)
		result["message"] = "Failed to get database connection"
		return utils.SendSimpleSuccess(c, result, "Connection test completed")
	}

	start := time.Now()
	if err := sqlDB.Ping(); err != nil {
		log.Printf("Error pinging database: %v", err)
		result["message"] = "Ping failed: " + err.Error()
		return utils.SendSimpleSuccess(c, result, "Connection test completed")
	}
	latency := time.Since(start).Milliseconds()

	result["status"] = "success"
	result["latency_ms"] = latency
	result["message"] = "Connection is healthy"

	return utils.SendSimpleSuccess(c, result, "Connection test completed successfully")
}

// GetDatabaseStats returns real database statistics via PostgreSQL introspection
func GetDatabaseStats(c *fiber.Ctx) error {
	now := time.Now().Format(time.RFC3339)

	stats := map[string]interface{}{
		"total_connections":   1,
		"active_connections":  0,
		"idle_connections":    0,
		"total_databases":     1,
		"total_tables":        0,
		"total_size_bytes":    int64(0),
		"total_size_pretty":   "0 bytes",
		"uptime_seconds":      0,
		"replication_enabled": false,
		"collected_at":        now,
	}

	// Real connection pool stats
	sqlDB, err := config.DB.DB()
	if err == nil {
		dbStats := sqlDB.Stats()
		stats["active_connections"] = dbStats.InUse
		stats["idle_connections"] = dbStats.Idle
		stats["total_connections"] = dbStats.OpenConnections
		stats["max_open_connections"] = dbStats.MaxOpenConnections
		stats["wait_count"] = dbStats.WaitCount
		stats["wait_duration_ms"] = dbStats.WaitDuration.Milliseconds()
	}

	// Real database size
	var dbSize int64
	var dbSizePretty string
	if err := config.DB.Raw("SELECT pg_database_size(current_database())").Scan(&dbSize).Error; err == nil {
		stats["total_size_bytes"] = dbSize
	}
	if err := config.DB.Raw("SELECT pg_size_pretty(pg_database_size(current_database()))").Scan(&dbSizePretty).Error; err == nil {
		stats["total_size_pretty"] = dbSizePretty
	}

	// Real table count
	var tableCount int64
	config.DB.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_type = 'BASE TABLE'").Scan(&tableCount)
	stats["total_tables"] = tableCount

	// Real index count
	var indexCount int64
	config.DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE schemaname = 'public'").Scan(&indexCount)
	stats["total_indexes"] = indexCount

	// Database uptime
	var uptimeSeconds float64
	config.DB.Raw("SELECT EXTRACT(EPOCH FROM (now() - pg_postmaster_start_time()))").Scan(&uptimeSeconds)
	stats["uptime_seconds"] = int64(uptimeSeconds)

	// Database version
	var version string
	config.DB.Raw("SELECT version()").Scan(&version)
	stats["version"] = version

	return utils.SendSimpleSuccess(c, stats, "Database stats retrieved successfully")
}

// GetDatabaseMetrics returns real database metrics
func GetDatabaseMetrics(c *fiber.Ctx) error {
	now := time.Now().Format(time.RFC3339)

	metrics := map[string]interface{}{
		"collected_at": now,
	}

	// Connection pool stats
	sqlDB, err := config.DB.DB()
	if err == nil {
		stats := sqlDB.Stats()
		metrics["connections"] = map[string]interface{}{
			"active":     stats.InUse,
			"idle":       stats.Idle,
			"total":      stats.OpenConnections,
			"max":        stats.MaxOpenConnections,
			"wait_count": stats.WaitCount,
		}
	}

	// Storage metrics from pg_stat_database
	var dbSize int64
	config.DB.Raw("SELECT pg_database_size(current_database())").Scan(&dbSize)
	var tableCount int64
	config.DB.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_type = 'BASE TABLE'").Scan(&tableCount)
	var indexCount int64
	config.DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE schemaname = 'public'").Scan(&indexCount)

	metrics["storage"] = map[string]interface{}{
		"database_size_bytes": dbSize,
		"tables_count":        tableCount,
		"indexes_count":       indexCount,
	}

	// Transaction stats from pg_stat_database
	type dbStatRow struct {
		XactCommit   int64 `gorm:"column:xact_commit"`
		XactRollback int64 `gorm:"column:xact_rollback"`
		BlksRead     int64 `gorm:"column:blks_read"`
		BlksHit      int64 `gorm:"column:blks_hit"`
		TupReturned  int64 `gorm:"column:tup_returned"`
		TupFetched   int64 `gorm:"column:tup_fetched"`
		TupInserted  int64 `gorm:"column:tup_inserted"`
		TupUpdated   int64 `gorm:"column:tup_updated"`
		TupDeleted   int64 `gorm:"column:tup_deleted"`
		Deadlocks    int64 `gorm:"column:deadlocks"`
	}
	var dbStat dbStatRow
	config.DB.Raw(`SELECT xact_commit, xact_rollback, blks_read, blks_hit,
		tup_returned, tup_fetched, tup_inserted, tup_updated, tup_deleted, deadlocks
		FROM pg_stat_database WHERE datname = current_database()`).Scan(&dbStat)

	metrics["queries"] = map[string]interface{}{
		"total_commits":   dbStat.XactCommit,
		"total_rollbacks": dbStat.XactRollback,
		"tuples_returned": dbStat.TupReturned,
		"tuples_fetched":  dbStat.TupFetched,
		"tuples_inserted": dbStat.TupInserted,
		"tuples_updated":  dbStat.TupUpdated,
		"tuples_deleted":  dbStat.TupDeleted,
		"deadlocks":       dbStat.Deadlocks,
	}

	// Cache hit ratio
	totalBlocks := dbStat.BlksRead + dbStat.BlksHit
	hitRatio := 0.0
	if totalBlocks > 0 {
		hitRatio = float64(dbStat.BlksHit) / float64(totalBlocks) * 100
	}
	metrics["cache"] = map[string]interface{}{
		"hit_ratio":   fmt.Sprintf("%.2f%%", hitRatio),
		"blocks_read": dbStat.BlksRead,
		"blocks_hit":  dbStat.BlksHit,
	}

	metrics["replication"] = map[string]interface{}{
		"enabled":  false,
		"lag_ms":   0,
		"replicas": 0,
	}

	return utils.SendSimpleSuccess(c, metrics, "Database metrics retrieved successfully")
}

// GetDatabaseTables returns real table information via information_schema and pg_class
func GetDatabaseTables(c *fiber.Ctx) error {
	if err := validateConnectionID(c); err != nil {
		return err
	}

	type tableInfo struct {
		TableName  string `gorm:"column:table_name"`
		RowEstimate int64  `gorm:"column:row_estimate"`
		TotalSize  string `gorm:"column:total_size"`
		DataSize   string `gorm:"column:data_size"`
		IndexSize  string `gorm:"column:index_size"`
		TotalBytes int64  `gorm:"column:total_bytes"`
	}

	var tables []tableInfo
	err := config.DB.Raw(`
		SELECT
			t.table_name,
			COALESCE(c.reltuples::bigint, 0) AS row_estimate,
			pg_size_pretty(pg_total_relation_size(quote_ident(t.table_name))) AS total_size,
			pg_size_pretty(pg_relation_size(quote_ident(t.table_name))) AS data_size,
			pg_size_pretty(pg_indexes_size(quote_ident(t.table_name))) AS index_size,
			pg_total_relation_size(quote_ident(t.table_name)) AS total_bytes
		FROM information_schema.tables t
		LEFT JOIN pg_class c ON c.relname = t.table_name AND c.relkind = 'r'
		WHERE t.table_schema = 'public' AND t.table_type = 'BASE TABLE'
		ORDER BY pg_total_relation_size(quote_ident(t.table_name)) DESC
	`).Scan(&tables).Error

	if err != nil {
		log.Printf("Error fetching database tables: %v", err)
		return utils.SendInternalError(c, "Failed to fetch database tables", err)
	}

	result := make([]map[string]interface{}, 0, len(tables))
	for _, t := range tables {
		// Count indexes for this table
		var indexCount int64
		config.DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE schemaname = 'public' AND tablename = ?", t.TableName).Scan(&indexCount)

		result = append(result, map[string]interface{}{
			"name":         t.TableName,
			"schema":       "public",
			"row_estimate": t.RowEstimate,
			"total_size":   t.TotalSize,
			"data_size":    t.DataSize,
			"index_size":   t.IndexSize,
			"total_bytes":  t.TotalBytes,
			"index_count":  indexCount,
		})
	}

	return utils.SendSimpleSuccess(c, result, "Database tables retrieved successfully")
}

// GetRunningQueries returns currently running queries from pg_stat_activity
func GetRunningQueries(c *fiber.Ctx) error {
	type queryInfo struct {
		PID         int     `gorm:"column:pid"`
		Username    string  `gorm:"column:usename"`
		Database    string  `gorm:"column:datname"`
		State       string  `gorm:"column:state"`
		Query       string  `gorm:"column:query"`
		QueryStart  *string `gorm:"column:query_start"`
		WaitEvent   *string `gorm:"column:wait_event_type"`
		DurationSec float64 `gorm:"column:duration_sec"`
		BackendType string  `gorm:"column:backend_type"`
	}

	var queries []queryInfo
	err := config.DB.Raw(`
		SELECT
			pid,
			usename,
			datname,
			COALESCE(state, 'unknown') AS state,
			COALESCE(query, '') AS query,
			query_start::text AS query_start,
			wait_event_type,
			COALESCE(EXTRACT(EPOCH FROM (now() - query_start)), 0) AS duration_sec,
			COALESCE(backend_type, 'unknown') AS backend_type
		FROM pg_stat_activity
		WHERE datname = current_database()
			AND pid != pg_backend_pid()
			AND state IS NOT NULL
		ORDER BY query_start ASC NULLS LAST
	`).Scan(&queries).Error

	if err != nil {
		log.Printf("Error fetching running queries: %v", err)
		return utils.SendInternalError(c, "Failed to fetch running queries", err)
	}

	result := make([]map[string]interface{}, 0, len(queries))
	for _, q := range queries {
		// Truncate long queries for display
		queryText := q.Query
		if len(queryText) > 500 {
			queryText = queryText[:500] + "..."
		}

		entry := map[string]interface{}{
			"pid":          q.PID,
			"username":     q.Username,
			"database":     q.Database,
			"state":        q.State,
			"query":        queryText,
			"duration_sec": fmt.Sprintf("%.2f", q.DurationSec),
			"backend_type": q.BackendType,
		}
		if q.QueryStart != nil {
			entry["query_start"] = *q.QueryStart
		}
		if q.WaitEvent != nil {
			entry["wait_event_type"] = *q.WaitEvent
		}
		result = append(result, entry)
	}

	return utils.SendSimpleSuccess(c, result, "Running queries retrieved successfully")
}

// ExecuteDatabaseQuery executes a read-only SQL query with a timeout
func ExecuteDatabaseQuery(c *fiber.Ctx) error {
	if err := validateConnectionID(c); err != nil {
		return err
	}

	type queryRequest struct {
		Query   string `json:"query"`
		Timeout int    `json:"timeout"` // seconds, default 5
	}

	var req queryRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendBadRequest(c, "Invalid request body")
	}

	if strings.TrimSpace(req.Query) == "" {
		return utils.SendBadRequest(c, "Query cannot be empty")
	}

	// Safety: only allow SELECT statements
	normalized := strings.TrimSpace(strings.ToUpper(req.Query))
	if !strings.HasPrefix(normalized, "SELECT") &&
		!strings.HasPrefix(normalized, "EXPLAIN") &&
		!strings.HasPrefix(normalized, "SHOW") {
		return utils.SendBadRequest(c, "Only SELECT, EXPLAIN, and SHOW statements are allowed")
	}

	// Reject dangerous patterns
	dangerousPatterns := []string{
		"DROP ", "DELETE ", "UPDATE ", "INSERT ", "ALTER ", "CREATE ",
		"TRUNCATE ", "GRANT ", "REVOKE ", "COPY ", "EXECUTE ",
		"INTO OUTFILE", "INTO DUMPFILE", "LOAD_FILE",
	}
	for _, pattern := range dangerousPatterns {
		if strings.Contains(normalized, pattern) {
			return utils.SendBadRequest(c, "Query contains disallowed operations")
		}
	}

	timeout := req.Timeout
	if timeout <= 0 || timeout > 30 {
		timeout = 5
	}

	start := time.Now()

	// Set statement timeout and execute
	var rows []map[string]interface{}
	tx := config.DB.Raw(fmt.Sprintf("SET LOCAL statement_timeout = '%ds'", timeout))
	if tx.Error != nil {
		log.Printf("Error setting statement timeout: %v", tx.Error)
	}

	err := config.DB.Raw(req.Query).Scan(&rows).Error
	duration := time.Since(start).Milliseconds()

	if err != nil {
		return utils.SendSimpleSuccess(c, map[string]interface{}{
			"success":     false,
			"error":       err.Error(),
			"duration_ms": duration,
			"row_count":   0,
		}, "Query execution failed")
	}

	// Limit result rows to prevent overwhelming the UI
	truncated := false
	if len(rows) > 1000 {
		rows = rows[:1000]
		truncated = true
	}

	return utils.SendSimpleSuccess(c, map[string]interface{}{
		"success":     true,
		"rows":        rows,
		"row_count":   len(rows),
		"truncated":   truncated,
		"duration_ms": duration,
	}, "Query executed successfully")
}

// CancelDatabaseQuery cancels a running query by PID
func CancelDatabaseQuery(c *fiber.Ctx) error {
	pidStr := c.Params("id")

	var cancelled bool
	err := config.DB.Raw("SELECT pg_cancel_backend(?::int)", pidStr).Scan(&cancelled).Error
	if err != nil {
		log.Printf("Error cancelling query (pid=%s): %v", pidStr, err)
		return utils.SendInternalError(c, "Failed to cancel query", err)
	}

	if cancelled {
		return utils.SendSimpleSuccess(c, map[string]interface{}{
			"pid":       pidStr,
			"cancelled": true,
		}, "Query cancelled successfully")
	}

	return utils.SendSimpleSuccess(c, map[string]interface{}{
		"pid":       pidStr,
		"cancelled": false,
		"message":   "Query could not be cancelled (may have already finished)",
	}, "Query cancellation attempted")
}

// GetDatabaseBackups returns information about backups (informational)
func GetDatabaseBackups(c *fiber.Ctx) error {
	return utils.SendSimpleSuccess(c, map[string]interface{}{
		"backups": []interface{}{},
		"message": "Backups are managed via pg_dump CLI or your hosting provider's backup system. This interface provides monitoring only.",
	}, "Database backups info retrieved")
}

// CreateDatabaseBackup provides instructions for creating a backup
func CreateDatabaseBackup(c *fiber.Ctx) error {
	return utils.SendNotImplementedError(c, "Database backups should be created using pg_dump or your hosting provider's backup tools. Running backups through the web UI is not supported for safety reasons.")
}

// RestoreDatabaseBackup provides instructions for restoring a backup
func RestoreDatabaseBackup(c *fiber.Ctx) error {
	return utils.SendNotImplementedError(c, "Database restores should be performed using pg_restore or your hosting provider's restore tools. Running restores through the web UI is not supported for safety reasons.")
}

// GetDatabaseMigrations returns GORM migration history if available
func GetDatabaseMigrations(c *fiber.Ctx) error {
	if err := validateConnectionID(c); err != nil {
		return err
	}

	// Check if GORM's migration tracking table exists
	var tableExists bool
	config.DB.Raw("SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'schema_migrations')").Scan(&tableExists)

	if !tableExists {
		// Try GORM's auto-migration tables - check for any migration-related tables
		type migrationTable struct {
			TableName string `gorm:"column:table_name"`
		}
		var migTables []migrationTable
		config.DB.Raw(`SELECT table_name FROM information_schema.tables
			WHERE table_schema = 'public' AND table_name LIKE '%migration%'
			ORDER BY table_name`).Scan(&migTables)

		if len(migTables) == 0 {
			return utils.SendSimpleSuccess(c, map[string]interface{}{
				"migrations": []interface{}{},
				"message":    "No migration tracking table found. Migrations are managed by GORM AutoMigrate.",
			}, "Database migrations retrieved")
		}
	}

	// List all tables with creation timestamps as a proxy for migration history
	type tableRecord struct {
		TableName string `gorm:"column:table_name"`
	}
	var tables []tableRecord
	config.DB.Raw(`SELECT table_name FROM information_schema.tables
		WHERE table_schema = 'public' AND table_type = 'BASE TABLE'
		ORDER BY table_name`).Scan(&tables)

	migrations := make([]map[string]interface{}, 0, len(tables))
	for i, t := range tables {
		migrations = append(migrations, map[string]interface{}{
			"id":        i + 1,
			"name":      fmt.Sprintf("create_%s", t.TableName),
			"table":     t.TableName,
			"status":    "applied",
			"engine":    "gorm_automigrate",
		})
	}

	return utils.SendSimpleSuccess(c, map[string]interface{}{
		"migrations": migrations,
		"total":      len(migrations),
		"message":    "Migrations are managed by GORM AutoMigrate. Showing current table state.",
	}, "Database migrations retrieved successfully")
}

// RunDatabaseMigration - not supported via web UI
func RunDatabaseMigration(c *fiber.Ctx) error {
	return utils.SendNotImplementedError(c, "Running migrations through the web UI is not supported for safety reasons. Use the application's migration CLI or GORM AutoMigrate during deployment.")
}

// RollbackDatabaseMigration - not supported via web UI
func RollbackDatabaseMigration(c *fiber.Ctx) error {
	return utils.SendNotImplementedError(c, "Rolling back migrations through the web UI is not supported for safety reasons. Use the application's migration CLI or manual SQL scripts.")
}

// OptimizeDatabaseTable runs ANALYZE on the specified table
func OptimizeDatabaseTable(c *fiber.Ctx) error {
	if err := validateConnectionID(c); err != nil {
		return err
	}

	tableName := c.Params("tableName")

	// Validate table exists in public schema
	var exists bool
	config.DB.Raw("SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = ?)", tableName).Scan(&exists)

	if !exists {
		return utils.SendNotFound(c, fmt.Sprintf("Table '%s' not found in public schema", tableName))
	}

	start := time.Now()
	err := config.DB.Exec(fmt.Sprintf("ANALYZE %s", sanitizeIdentifier(tableName))).Error
	duration := time.Since(start).Milliseconds()

	if err != nil {
		log.Printf("Error analyzing table %s: %v", tableName, err)
		return utils.SendInternalError(c, "Failed to optimize table", err)
	}

	return utils.SendSimpleSuccess(c, map[string]interface{}{
		"table":       tableName,
		"operation":   "ANALYZE",
		"duration_ms": duration,
		"message":     fmt.Sprintf("Table '%s' has been analyzed. PostgreSQL will now have updated statistics for query planning.", tableName),
	}, "Table optimized successfully")
}

// ExportDatabase provides instructions for database export
func ExportDatabase(c *fiber.Ctx) error {
	return utils.SendNotImplementedError(c, "Database export should be performed using pg_dump or your hosting provider's export tools. Running exports through the web UI is not supported for safety and performance reasons.")
}

// GetDatabaseSchemas returns real schema information via information_schema
func GetDatabaseSchemas(c *fiber.Ctx) error {
	if err := validateConnectionID(c); err != nil {
		return err
	}

	type schemaInfo struct {
		SchemaName string `gorm:"column:schema_name"`
		SchemaOwner string `gorm:"column:schema_owner"`
	}

	var schemas []schemaInfo
	err := config.DB.Raw(`
		SELECT schema_name, COALESCE(schema_owner, 'unknown') as schema_owner
		FROM information_schema.schemata
		WHERE schema_name NOT IN ('pg_toast', 'pg_catalog', 'information_schema')
		ORDER BY schema_name
	`).Scan(&schemas).Error

	if err != nil {
		log.Printf("Error fetching schemas: %v", err)
		return utils.SendInternalError(c, "Failed to fetch database schemas", err)
	}

	result := make([]map[string]interface{}, 0, len(schemas))
	for _, s := range schemas {
		// Count tables in this schema
		var tableCount int64
		config.DB.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = ? AND table_type = 'BASE TABLE'", s.SchemaName).Scan(&tableCount)

		// Get schema size
		var sizeBytes int64
		config.DB.Raw(`
			SELECT COALESCE(SUM(pg_total_relation_size(quote_ident(t.table_name))), 0)
			FROM information_schema.tables t
			WHERE t.table_schema = ? AND t.table_type = 'BASE TABLE'
		`, s.SchemaName).Scan(&sizeBytes)

		var sizePretty string
		config.DB.Raw("SELECT pg_size_pretty(?::bigint)", sizeBytes).Scan(&sizePretty)

		result = append(result, map[string]interface{}{
			"name":         s.SchemaName,
			"owner":        s.SchemaOwner,
			"tables_count": tableCount,
			"size_bytes":   sizeBytes,
			"size_pretty":  sizePretty,
		})
	}

	return utils.SendSimpleSuccess(c, result, "Database schemas retrieved successfully")
}

// GetDatabasePerformance returns real performance data from PostgreSQL stats
func GetDatabasePerformance(c *fiber.Ctx) error {
	if err := validateConnectionID(c); err != nil {
		return err
	}

	now := time.Now().Format(time.RFC3339)

	performance := map[string]interface{}{
		"connection_id": connectionID,
		"collected_at":  now,
	}

	// Connection pool stats
	sqlDB, err := config.DB.DB()
	if err == nil {
		stats := sqlDB.Stats()
		performance["connections"] = map[string]interface{}{
			"active": stats.InUse,
			"idle":   stats.Idle,
			"total":  stats.OpenConnections,
			"max":    stats.MaxOpenConnections,
		}
	}

	// Database-level stats from pg_stat_database
	type dbPerfRow struct {
		XactCommit   int64 `gorm:"column:xact_commit"`
		XactRollback int64 `gorm:"column:xact_rollback"`
		BlksRead     int64 `gorm:"column:blks_read"`
		BlksHit      int64 `gorm:"column:blks_hit"`
		Deadlocks    int64 `gorm:"column:deadlocks"`
		Conflicts    int64 `gorm:"column:conflicts"`
	}
	var dbPerf dbPerfRow
	config.DB.Raw(`SELECT xact_commit, xact_rollback, blks_read, blks_hit, deadlocks, conflicts
		FROM pg_stat_database WHERE datname = current_database()`).Scan(&dbPerf)

	performance["transactions"] = map[string]interface{}{
		"total_commits":   dbPerf.XactCommit,
		"total_rollbacks": dbPerf.XactRollback,
		"deadlocks":       dbPerf.Deadlocks,
		"conflicts":       dbPerf.Conflicts,
	}

	// Cache hit ratio
	totalBlocks := dbPerf.BlksRead + dbPerf.BlksHit
	hitRatio := 0.0
	if totalBlocks > 0 {
		hitRatio = float64(dbPerf.BlksHit) / float64(totalBlocks) * 100
	}
	performance["cache"] = map[string]interface{}{
		"hit_ratio_pct": fmt.Sprintf("%.2f", hitRatio),
		"blocks_read":   dbPerf.BlksRead,
		"blocks_hit":    dbPerf.BlksHit,
	}

	// Lock information
	type lockInfo struct {
		Total   int64 `gorm:"column:total"`
		Waiting int64 `gorm:"column:waiting"`
	}
	var locks lockInfo
	config.DB.Raw(`SELECT
		COUNT(*) as total,
		COUNT(*) FILTER (WHERE NOT granted) as waiting
		FROM pg_locks WHERE database = (SELECT oid FROM pg_database WHERE datname = current_database())`).Scan(&locks)

	performance["locks"] = map[string]interface{}{
		"total":   locks.Total,
		"waiting": locks.Waiting,
	}

	// Active query count
	var activeQueries int64
	config.DB.Raw("SELECT COUNT(*) FROM pg_stat_activity WHERE datname = current_database() AND state = 'active' AND pid != pg_backend_pid()").Scan(&activeQueries)
	performance["active_queries"] = activeQueries

	// Table with most sequential scans (potential optimization target)
	type seqScanInfo struct {
		TableName string `gorm:"column:relname"`
		SeqScan   int64  `gorm:"column:seq_scan"`
		IdxScan   int64  `gorm:"column:idx_scan"`
	}
	var topSeqScans []seqScanInfo
	config.DB.Raw(`SELECT relname, COALESCE(seq_scan, 0) as seq_scan, COALESCE(idx_scan, 0) as idx_scan
		FROM pg_stat_user_tables
		WHERE seq_scan > 0
		ORDER BY seq_scan DESC LIMIT 5`).Scan(&topSeqScans)

	seqScanResult := make([]map[string]interface{}, 0, len(topSeqScans))
	for _, s := range topSeqScans {
		seqScanResult = append(seqScanResult, map[string]interface{}{
			"table":     s.TableName,
			"seq_scan":  s.SeqScan,
			"idx_scan":  s.IdxScan,
		})
	}
	performance["top_sequential_scans"] = seqScanResult

	return utils.SendSimpleSuccess(c, performance, "Database performance data retrieved successfully")
}

// sanitizeIdentifier ensures a SQL identifier is safe (alphanumeric + underscores only)
func sanitizeIdentifier(name string) string {
	var safe strings.Builder
	for _, ch := range name {
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_' {
			safe.WriteRune(ch)
		}
	}
	return safe.String()
}
