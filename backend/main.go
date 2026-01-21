package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/handlers"
	"github.com/liyali/liyali-gateway/logging"
	"github.com/liyali/liyali-gateway/middleware"
	"github.com/liyali/liyali-gateway/routes"
	"github.com/liyali/liyali-gateway/services"
	"github.com/liyali/liyali-gateway/repository"
)

func init() {
	// Load environment variables
	err := godotenv.Load(".env")
	if err != nil && os.Getenv("APP_ENV") == "" {
		// Use basic logging before structured logging is initialized
		println("Note: .env file not found, using environment variables")
	}

	// Set default values if not provided
	if os.Getenv("DB_PORT") == "" {
		os.Setenv("DB_PORT", "5432")
	}
	if os.Getenv("DB_HOST") == "" {
		os.Setenv("DB_HOST", "localhost")
	}
	if os.Getenv("DB_USER") == "" {
		os.Setenv("DB_USER", "postgres")
	}
	if os.Getenv("APP_PORT") == "" {
		os.Setenv("APP_PORT", "8080")
	}
	if os.Getenv("FRONTEND_URL") == "" {
		os.Setenv("FRONTEND_URL", "http://localhost:3000")
	}

	// Environment-specific configuration
	appEnv := os.Getenv("APP_ENV")
	isProduction := appEnv == "production" || appEnv == "prod"

	// JWT_SECRET is required in production
	if os.Getenv("JWT_SECRET") == "" {
		if isProduction {
			println("FATAL: JWT_SECRET environment variable is required in production mode")
			os.Exit(1)
		}
		// Only use default in development
		os.Setenv("JWT_SECRET", "dev-only-secret-do-not-use-in-production")
	}

	// SSL mode defaults: require in production, disable in development
	if os.Getenv("DB_SSL_MODE") == "" {
		if isProduction {
			os.Setenv("DB_SSL_MODE", "require")
		} else {
			os.Setenv("DB_SSL_MODE", "disable")
		}
	}
}

func main() {
	// Initialize structured logging system
	loggingConfig := logging.SetupLogging()
	
	// Initialize database (both GORM and pgx)
	config.InitDatabase()

	// Initialize repositories
	userRepo := repository.NewUserRepository(config.PgxDB, config.DB)
	sessionRepo := repository.NewSessionRepository(config.PgxDB)
	passwordResetRepo := repository.NewPasswordResetRepository(config.PgxDB)
	loginAttemptRepo := repository.NewLoginAttemptRepository(config.PgxDB)
	lockoutRepo := repository.NewAccountLockoutRepository(config.PgxDB)
	roleRepo := repository.NewOrganizationRoleRepository(config.PgxDB)
	workflowRepo := repository.NewWorkflowRepository(config.PgxDB, config.DB)
	documentRepo := repository.NewDocumentRepository(config.PgxDB, config.DB)
	
	// Initialize audit service
	auditService := &services.AuditService{}
	
	// Initialize enhanced services
	authService := services.NewAuthService(
		userRepo,
		sessionRepo,
		passwordResetRepo,
		loginAttemptRepo,
		lockoutRepo,
		auditService,
		os.Getenv("JWT_SECRET"),
		config.DB, // Add GORM database connection
	)
	
	rbacService := services.NewRBACService(roleRepo, auditService, config.DB)
	workflowService := services.NewWorkflowService(workflowRepo, auditService, config.DB)
	
	// Initialize notification service (placeholder for now)
	notificationService := &services.NotificationService{}
	
	// Initialize automation service
	automationService := services.NewDocumentAutomationService(config.DB, auditService, notificationService)
	
	// Initialize workflow execution service with automation
	workflowExecutionService := services.NewWorkflowExecutionService(config.DB, workflowService, auditService, automationService)
	documentService := services.NewDocumentService(documentRepo, auditService)

	// Initialize handler registry
	handlerRegistry := handlers.NewHandlerRegistry(authService, rbacService, workflowService, workflowExecutionService, documentService, automationService)

	// Create Fiber app with global error handler
	app := fiber.New(fiber.Config{
		AppName:      "Liyali Gateway Backend API",
		ErrorHandler: customErrorHandler,
	})

	// Setup structured logging middleware (replaces old LoggerMiddleware)
	logging.SetupFiberMiddleware(app, loggingConfig)
	
	// Other middleware
	app.Use(middleware.ErrorHandlingMiddleware())
	app.Use(middleware.CORSMiddleware())

	// Setup routes with handler registry
	routes.SetupRoutes(app, handlerRegistry, rbacService, config.DB)

	// Start server with graceful shutdown
	go func() {
		port := os.Getenv("APP_PORT")
		
		// Log startup information using structured logging
		logging.LogStartupInfo(port)

		if err := app.Listen(":" + port); err != nil {
			logging.WithError(err).Fatal("failed_to_start_server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	// Log shutdown information
	logging.LogShutdownInfo()
	
	if err := app.ShutdownWithContext(context.Background()); err != nil {
		logging.WithError(err).Fatal("server_forced_shutdown")
	}

	logging.Info("server_stopped_gracefully")
}

// customErrorHandler handles errors globally with structured logging
func customErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	message := "Internal Server Error"

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
		message = e.Message
	}

	// Log error with structured logging
	logger := logging.FromContext(c)
	logger.WithError(err).WithFields(map[string]interface{}{
		"status_code": code,
		"error_message": message,
		"method": c.Method(),
		"path": c.Path(),
	}).Error("global_error_handler")

	return c.Status(code).JSON(fiber.Map{
		"error": message,
		"request_id": logger.GetRequestID(),
	})
}
