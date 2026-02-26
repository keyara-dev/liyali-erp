package main

import (
	"context"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/handlers"
	"github.com/liyali/liyali-gateway/logging"
	"github.com/liyali/liyali-gateway/middleware"
	"github.com/liyali/liyali-gateway/repository"
	"github.com/liyali/liyali-gateway/routes"
	"github.com/liyali/liyali-gateway/services"
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

	// JWT_SECRET handling - temporarily more lenient for debugging
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		if isProduction {
			println("WARNING: JWT_SECRET not found, using temporary fallback")
			println("Available environment variables:")
			for _, env := range os.Environ() {
				if strings.Contains(strings.ToUpper(env), "JWT") || strings.Contains(strings.ToUpper(env), "SECRET") {
					println(env)
				}
			}
			// Use a temporary secret for now
			os.Setenv("JWT_SECRET", "temp-production-secret-change-me")
		} else {
			// Development default
			os.Setenv("JWT_SECRET", "dev-only-secret-do-not-use-in-production")
		}
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
	logger := &logging.Logger{}

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
	reportsRepo := repository.NewReportsRepository(config.PgxDB)

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

	// Bootstrap global system roles (super_admin, admin, approver, requester, finance, viewer)
	roleManagementService := services.NewRoleManagementService(config.DB)
	if err := roleManagementService.EnsureGlobalSystemRoles(); err != nil {
		logging.WithError(err).Error("failed_to_ensure_global_system_roles")
	}

	workflowService := services.NewWorkflowService(workflowRepo, auditService, config.DB)

	// Initialize notification service (placeholder for now)
	notificationService := &services.NotificationService{}

	// Initialize automation service
	automationService := services.NewDocumentAutomationService(config.DB, auditService, notificationService)
	documentGenerationService := services.NewDocumentGenerationService(config.DB, automationService)

	// Initialize workflow execution service with automation
	workflowExecutionService := services.NewWorkflowExecutionService(config.DB, workflowService, auditService, automationService)
	documentService := services.NewDocumentService(documentRepo, auditService)

	// Initialize subscription service
	subscriptionService := services.NewSubscriptionService(config.PgxDB, logger)

	// Initialize reports service
	reportsService := services.NewReportsService(reportsRepo)

	// Initialize handler registry
	handlerRegistry := handlers.NewHandlerRegistry(
		authService,
		rbacService,
		workflowService,
		workflowExecutionService,
		documentService,
		documentGenerationService,
		subscriptionService,
		reportsService,
		logger,
	)

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
		logging.WithError(err).Fatal("⚠️server_forced_shutdown")
	}

	logging.Info("✅server_stopped_gracefully!!")
}

// customErrorHandler handles errors globally with structured logging
func customErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	message := "Internal Server Error!"

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
		message = e.Message
	}

	// Log error with structured logging
	logger := logging.FromContext(c)
	logger.WithError(err).WithFields(map[string]interface{}{
		"status_code":   code,
		"error_message": message,
		"method":        c.Method(),
		"path":          c.Path(),
	}).Error("global_error_handler!!")

	return c.Status(code).JSON(fiber.Map{
		"error":      message,
		"request_id": logger.GetRequestID(),
	})
}
