package integration

import (
	"context"
	"testing"

	"github.com/cozyCodr/liyali-gateway/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

// setupTestDB initializes a test database connection
func setupTestDB(t *testing.T) *pgxpool.Pool {
	// Load config
	cfg, err := config.Load()
	require.NoError(t, err)

	// Connect to test database
	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	require.NoError(t, err)

	// Test database connection
	err = pool.Ping(context.Background())
	require.NoError(t, err)

	return pool
}
