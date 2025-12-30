package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/liyali/liyali-gateway/bootstrap"
	"github.com/liyali/liyali-gateway/bootstrap/circuit"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// Load environment variables
	if err := godotenv.Load("../../.env"); err != nil {
		log.Println("Note: .env file not found, using environment variables")
	}

	// Set default values if not provided
	setDefaultEnvVars()

	// Create database connection
	db, err := connectToDatabase()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Create custom bootstrap configuration
	config := createBootstrapConfig()

	// Create custom logger
	logger := log.New(os.Stdout, "[BOOTSTRAP] ", log.LstdFlags|log.Lshortfile)

	// Create bootstrapper
	bootstrapper := bootstrap.NewBootstrapper(db, config, logger)

	// Run bootstrap with timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
	defer cancel()

	logger.Println("Starting database bootstrap process...")
	result := bootstrapper.Bootstrap(ctx)

	// Handle results
	if result.Success {
		logger.Printf("✅ Bootstrap completed successfully!")
		logger.Printf("   Duration: %v", result.Duration)
		logger.Printf("   Phase: %s", result.Phase)
		
		// Print metrics
		printMetrics(bootstrapper.GetMetrics())
		
		// Demonstrate health check
		demonstrateHealthCheck(bootstrapper)
		
	} else {
		logger.Printf("❌ Bootstrap failed!")
		logger.Printf("   Failed at phase: %s", result.Phase)
		logger.Printf("   Duration: %v", result.Duration)
		logger.Printf("   Error: %v", result.Error)
		
		// Print failure metrics
		if result.Metrics != nil {
			logger.Printf("   Failure metrics: %+v", result.Metrics)
		}
		
		os.Exit(1)
	}
}

func setDefaultEnvVars() {
	envDefaults := map[string]string{
		"DB_HOST":     "localhost",
		"DB_PORT":     "5432",
		"DB_USER":     "postgres",
		"DB_PASSWORD": "postgres",
		"DB_NAME":     "liyali_gateway",
		"DB_SSL_MODE": "disable",
		"APP_ENV":     "development",
	}

	for key, defaultValue := range envDefaults {
		if os.Getenv(key) == "" {
			os.Setenv(key, defaultValue)
		}
	}
}

func connectToDatabase() (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_SSL_MODE"),
	)

	// Configure GORM logger based on environment
	var gormLogger logger.Interface
	if os.Getenv("APP_ENV") == "production" {
		gormLogger = logger.Default.LogMode(logger.Error)
	} else {
		gormLogger = logger.Default.LogMode(logger.Info)
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Set connection pool parameters
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}

func createBootstrapConfig() *bootstrap.BootstrapConfig {
	config := bootstrap.DefaultBootstrapConfig()
	
	// Customize based on environment
	config.Environment = os.Getenv("APP_ENV")
	
	switch config.Environment {
	case "production":
		// Production settings - more conservative
		config.SkipSeeding = os.Getenv("ENABLE_SEEDING") != "true"
		config.SeedRetryAttempts = 5
		config.SeedRetryDelay = time.Second * 5
		config.CircuitBreakerConfig = circuit.Config{
			MaxFailures: 3,
			Timeout:     time.Minute,
			Interval:    time.Minute * 2,
		}
		config.ValidationTimeout = time.Minute
		config.MigrationTimeout = time.Minute * 10
		
	case "staging":
		// Staging settings - balanced
		config.SkipSeeding = false
		config.SeedRetryAttempts = 3
		config.SeedRetryDelay = time.Second * 2
		config.ValidationTimeout = time.Second * 45
		config.MigrationTimeout = time.Minute * 5
		
	default: // development, test
		// Development settings - fast and permissive
		config.SkipSeeding = false
		config.SeedRetryAttempts = 2
		config.SeedRetryDelay = time.Second
		config.ValidationTimeout = time.Second * 30
		config.MigrationTimeout = time.Minute * 2
	}
	
	return config
}

func printMetrics(metrics map[string]interface{}) {
	log.Println("📊 Bootstrap Metrics:")
	log.Println("=" + string(make([]byte, 40)))
	
	for key, value := range metrics {
		log.Printf("  %-30s: %v", key, value)
	}
	
	log.Println("=" + string(make([]byte, 40)))
}

func demonstrateHealthCheck(bootstrapper *bootstrap.Bootstrapper) {
	log.Println("🏥 Running health check...")
	
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	
	if err := bootstrapper.HealthCheck(ctx); err != nil {
		log.Printf("❌ Health check failed: %v", err)
	} else {
		log.Println("✅ Health check passed")
	}
}

// Example of how to integrate bootstrap into a web server
func integrateWithWebServer() {
	// This would be in your main application
	
	// 1. Initialize database with bootstrap
	db, err := connectToDatabase()
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	
	config := createBootstrapConfig()
	bootstrapper := bootstrap.NewBootstrapper(db, config, nil)
	
	// 2. Run bootstrap during application startup
	ctx := context.Background()
	result := bootstrapper.Bootstrap(ctx)
	if !result.Success {
		log.Fatalf("Bootstrap failed: %v", result.Error)
	}
	
	// 3. Use health check in readiness probe
	healthCheckHandler := func() error {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		return bootstrapper.HealthCheck(ctx)
	}
	
	// 4. Expose metrics endpoint
	metricsHandler := func() map[string]interface{} {
		return bootstrapper.GetMetrics()
	}
	
	// Use these handlers in your HTTP server
	_ = healthCheckHandler
	_ = metricsHandler
	
	log.Println("Application ready to serve requests")
}