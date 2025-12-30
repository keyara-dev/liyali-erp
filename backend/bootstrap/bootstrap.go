package bootstrap

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/liyali/liyali-gateway/bootstrap/circuit"
	"github.com/liyali/liyali-gateway/bootstrap/retry"
	"github.com/liyali/liyali-gateway/bootstrap/seeder"
	"github.com/liyali/liyali-gateway/bootstrap/validator"
	"gorm.io/gorm"
)

// BootstrapPhase represents different phases of the bootstrap process
type BootstrapPhase string

const (
	PhaseConnect   BootstrapPhase = "connect"
	PhaseValidate  BootstrapPhase = "validate"
	PhaseMigrate   BootstrapPhase = "migrate"
	PhaseVerify    BootstrapPhase = "verify"
	PhaseSeed      BootstrapPhase = "seed"
	PhaseComplete  BootstrapPhase = "complete"
)

// BootstrapConfig holds configuration for the bootstrap process
type BootstrapConfig struct {
	Environment        string
	SkipSeeding       bool
	SeedRetryAttempts int
	SeedRetryDelay    time.Duration
	CircuitBreakerConfig circuit.Config
	ValidationTimeout time.Duration
	MigrationTimeout  time.Duration
}

// DefaultBootstrapConfig returns sensible defaults
func DefaultBootstrapConfig() *BootstrapConfig {
	return &BootstrapConfig{
		Environment:        "development",
		SkipSeeding:       false,
		SeedRetryAttempts: 3,
		SeedRetryDelay:    time.Second * 2,
		CircuitBreakerConfig: circuit.Config{
			MaxFailures: 5,
			Timeout:     time.Second * 30,
			Interval:    time.Second * 60,
		},
		ValidationTimeout: time.Second * 30,
		MigrationTimeout:  time.Minute * 5,
	}
}

// BootstrapResult contains the results of the bootstrap process
type BootstrapResult struct {
	Success      bool
	Phase        BootstrapPhase
	Duration     time.Duration
	Error        error
	Metrics      map[string]interface{}
}

// Bootstrapper handles the complete database bootstrap process
type Bootstrapper struct {
	db       *gorm.DB
	config   *BootstrapConfig
	validator *validator.DatabaseValidator
	seeder   *seeder.DatabaseSeeder
	breaker  *circuit.Breaker
	logger   *log.Logger
}

// NewBootstrapper creates a new bootstrapper instance
func NewBootstrapper(db *gorm.DB, config *BootstrapConfig, logger *log.Logger) *Bootstrapper {
	if config == nil {
		config = DefaultBootstrapConfig()
	}
	
	if logger == nil {
		logger = log.Default()
	}

	return &Bootstrapper{
		db:        db,
		config:    config,
		validator: validator.New(db, logger),
		seeder:    seeder.New(db, logger),
		breaker:   circuit.NewBreaker(config.CircuitBreakerConfig),
		logger:    logger,
	}
}

// Bootstrap executes the complete bootstrap process with proper phase ordering
func (b *Bootstrapper) Bootstrap(ctx context.Context) *BootstrapResult {
	startTime := time.Now()
	result := &BootstrapResult{
		Metrics: make(map[string]interface{}),
	}

	b.logger.Printf("🚀 Starting database bootstrap process (env: %s)", b.config.Environment)

	// Phase 1: Connect (already done, but validate connection)
	if err := b.executePhase(ctx, PhaseConnect, b.validateConnection); err != nil {
		return b.failResult(result, PhaseConnect, err, startTime)
	}

	// Phase 2: Validate schema readiness
	if err := b.executePhase(ctx, PhaseValidate, b.validateSchema); err != nil {
		return b.failResult(result, PhaseValidate, err, startTime)
	}

	// Phase 3: Run migrations (handled externally, but verify)
	if err := b.executePhase(ctx, PhaseMigrate, b.verifyMigrations); err != nil {
		return b.failResult(result, PhaseMigrate, err, startTime)
	}

	// Phase 4: Verify schema integrity
	if err := b.executePhase(ctx, PhaseVerify, b.verifySchemaIntegrity); err != nil {
		return b.failResult(result, PhaseVerify, err, startTime)
	}

	// Phase 5: Seed data (if not skipped)
	if !b.config.SkipSeeding {
		if err := b.executePhase(ctx, PhaseSeed, b.seedDatabase); err != nil {
			return b.failResult(result, PhaseSeed, err, startTime)
		}
	} else {
		b.logger.Println("⏭️  Skipping database seeding (disabled in config)")
	}

	// Complete
	duration := time.Since(startTime)
	result.Success = true
	result.Phase = PhaseComplete
	result.Duration = duration
	result.Metrics["total_duration_ms"] = duration.Milliseconds()

	b.logger.Printf("✅ Database bootstrap completed successfully in %v", duration)
	return result
}

// executePhase runs a bootstrap phase with timing and error handling
func (b *Bootstrapper) executePhase(ctx context.Context, phase BootstrapPhase, fn func(context.Context) error) error {
	phaseStart := time.Now()
	b.logger.Printf("📋 Phase: %s - Starting", phase)

	err := fn(ctx)
	duration := time.Since(phaseStart)

	if err != nil {
		b.logger.Printf("❌ Phase: %s - Failed in %v: %v", phase, duration, err)
		return fmt.Errorf("phase %s failed: %w", phase, err)
	}

	b.logger.Printf("✅ Phase: %s - Completed in %v", phase, duration)
	return nil
}

// validateConnection ensures database connection is healthy
func (b *Bootstrapper) validateConnection(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, b.config.ValidationTimeout)
	defer cancel()

	sqlDB, err := b.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	// Check connection pool stats
	stats := sqlDB.Stats()
	b.logger.Printf("📊 Connection pool stats: Open=%d, InUse=%d, Idle=%d", 
		stats.OpenConnections, stats.InUse, stats.Idle)

	return nil
}

// validateSchema checks if the database is ready for operations
func (b *Bootstrapper) validateSchema(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, b.config.ValidationTimeout)
	defer cancel()

	return b.validator.ValidateSchemaReadiness(ctx)
}

// verifyMigrations ensures migrations have been applied
func (b *Bootstrapper) verifyMigrations(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, b.config.MigrationTimeout)
	defer cancel()

	return b.validator.VerifyMigrations(ctx)
}

// verifySchemaIntegrity performs comprehensive schema validation
func (b *Bootstrapper) verifySchemaIntegrity(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, b.config.ValidationTimeout)
	defer cancel()

	return b.validator.VerifySchemaIntegrity(ctx)
}

// seedDatabase performs idempotent database seeding with circuit breaker
func (b *Bootstrapper) seedDatabase(ctx context.Context) error {
	return b.breaker.Execute(func() error {
		return retry.WithExponentialBackoff(
			ctx,
			b.config.SeedRetryAttempts,
			b.config.SeedRetryDelay,
			func() error {
				return b.seeder.SeedAll(ctx)
			},
		)
	})
}

// failResult creates a failed bootstrap result
func (b *Bootstrapper) failResult(result *BootstrapResult, phase BootstrapPhase, err error, startTime time.Time) *BootstrapResult {
	duration := time.Since(startTime)
	result.Success = false
	result.Phase = phase
	result.Duration = duration
	result.Error = err
	result.Metrics["failure_duration_ms"] = duration.Milliseconds()
	result.Metrics["failure_phase"] = string(phase)

	b.logger.Printf("💥 Bootstrap failed at phase %s after %v: %v", phase, duration, err)
	return result
}

// HealthCheck performs a quick health check of the bootstrap state
func (b *Bootstrapper) HealthCheck(ctx context.Context) error {
	// Quick connection check
	if err := b.validateConnection(ctx); err != nil {
		return fmt.Errorf("connection health check failed: %w", err)
	}

	// Quick schema validation
	if err := b.validator.QuickSchemaCheck(ctx); err != nil {
		return fmt.Errorf("schema health check failed: %w", err)
	}

	return nil
}

// GetMetrics returns current bootstrap metrics
func (b *Bootstrapper) GetMetrics() map[string]interface{} {
	metrics := make(map[string]interface{})
	
	// Database connection metrics
	if sqlDB, err := b.db.DB(); err == nil {
		stats := sqlDB.Stats()
		metrics["db_connections_open"] = stats.OpenConnections
		metrics["db_connections_in_use"] = stats.InUse
		metrics["db_connections_idle"] = stats.Idle
		metrics["db_connections_wait_count"] = stats.WaitCount
		metrics["db_connections_wait_duration"] = stats.WaitDuration.Milliseconds()
	}

	// Circuit breaker metrics
	metrics["circuit_breaker_state"] = b.breaker.State()
	metrics["circuit_breaker_failures"] = b.breaker.Failures()

	return metrics
}