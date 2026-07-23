package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/handlers"
	"github.com/liyali/liyali-gateway/logging"
	"github.com/liyali/liyali-gateway/middleware"
	"github.com/liyali/liyali-gateway/repository"
	"github.com/liyali/liyali-gateway/routes"
	"github.com/liyali/liyali-gateway/services"
	"github.com/liyali/liyali-gateway/utils"
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
	// Railway (and other PaaS) inject PORT; fall back to APP_PORT, then 8080
	if os.Getenv("APP_PORT") == "" {
		if port := os.Getenv("PORT"); port != "" {
			os.Setenv("APP_PORT", port)
		} else {
			os.Setenv("APP_PORT", "8080")
		}
	}
	if os.Getenv("FRONTEND_URL") == "" {
		os.Setenv("FRONTEND_URL", "http://localhost:3000")
	}

	// Environment-specific configuration
	appEnv := os.Getenv("APP_ENV")
	isProduction := appEnv == "production" || appEnv == "prod"

	// JWT_SECRET handling: fail fast in production, dev-only fallback otherwise.
	// Never substitute a hardcoded secret in production — a known fallback lets
	// anyone forge tokens for any tenant.
	if os.Getenv("JWT_SECRET") == "" {
		if isProduction {
			log.Fatal("FATAL: JWT_SECRET environment variable must be set in production")
		}
		// Development default — never reachable in production.
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
	logger := &logging.Logger{}

	// Validate critical environment variables before proceeding
	appEnv := os.Getenv("APP_ENV")
	isProduction := appEnv == "production" || appEnv == "prod"
	
	// Database connection validation
	databaseURL := os.Getenv("DATABASE_URL")
	dbPassword := os.Getenv("DB_PASSWORD")
	
	if databaseURL == "" && dbPassword == "" && isProduction {
		log.Fatal("FATAL: Either DATABASE_URL or DB_PASSWORD must be set in production")
	}
	
	if databaseURL == "" {
		// Using individual DB_* vars - ensure DB_PASSWORD is set in production
		if dbPassword == "" && isProduction {
			log.Fatal("FATAL: DB_PASSWORD environment variable must be set in production")
		}
		// Warn if DB_NAME is not set (will use default which might not be intended)
		if os.Getenv("DB_NAME") == "" {
			log.Println("WARNING: DB_NAME not set, database connection may fail")
		}
	}
	
	// Frontend URL validation (required for CORS and email links)
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" || frontendURL == "http://localhost:3000" {
		if isProduction {
			log.Fatal("FATAL: FRONTEND_URL must be set to your production frontend URL (currently set to dev default)")
		}
		log.Println("WARNING: FRONTEND_URL not configured, using development default")
	}
	
	// JWT secret validation (already done in init(), but double-check)
	if os.Getenv("JWT_SECRET") == "" {
		log.Fatal("FATAL: JWT_SECRET must be set (should have been caught in init)")
	}
	
	// Warn about optional but recommended settings
	if isProduction {
		if os.Getenv("SMTP_HOST") == "" && os.Getenv("EMAIL_ENABLED") == "true" {
			log.Println("WARNING: EMAIL_ENABLED=true but SMTP_HOST not configured - emails will not be sent")
		}
		if os.Getenv("REDIS_HOST") == "" {
			log.Println("INFO: REDIS_HOST not set, caching will be limited")
		}
	}
	
	log.Printf("Environment: %s", appEnv)
	log.Println("Starting Liyali Gateway...")

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
	auditService := services.NewAuditServiceWithDB(config.DB)

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

	// Start background worker to auto-expire stale task claims
	claimExpiryCtx, cancelClaimExpiry := context.WithCancel(context.Background())
	defer cancelClaimExpiry()
	go workflowExecutionService.StartClaimExpiryWorker(claimExpiryCtx)

	// Start background worker to expire stale organization invitations (hourly)
	invExpiryCtx, cancelInvExpiry := context.WithCancel(context.Background())
	defer cancelInvExpiry()
	go func() {
		defer utils.RecoverPanic("invitation-expiry-worker")
		invSvc := services.NewInvitationService(config.DB)
		ticker := time.NewTicker(time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := invSvc.ExpireStaleInvitations(); err != nil {
					log.Printf("[InvitationExpiry] error: %v", err)
				}
			case <-invExpiryCtx.Done():
				return
			}
		}
	}()

	// Initialize activity logging service
	activityRepo := repository.NewActivityRepository(config.DB)
	activityService := services.NewActivityService(activityRepo)

	// Start retention cleanup worker (runs daily, cleans up old activity logs)
	retentionCtx, cancelRetention := context.WithCancel(context.Background())
	defer cancelRetention()
	go activityService.StartRetentionCleanupWorker(retentionCtx)

	// Initialize session service (enriches session data with device/browser info)
	sessionService := services.NewSessionService(sessionRepo)

	documentService := services.NewDocumentService(documentRepo, auditService, config.DB)

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

	// Wire activity and session services into AuthHandler
	handlerRegistry.Auth.SetActivityService(activityService)
	handlerRegistry.Auth.SetSessionService(sessionService)

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
	routes.SetupRoutes(app, handlerRegistry, rbacService, config.DB, activityService)

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

	// Client response: never leak internal 5xx detail in production. 4xx
	// messages are intentional/safe and kept. The full error, method, path and
	// request_id are in the server log above — debug via the request_id, which
	// is returned to the client so it can be quoted in a support request.
	clientMessage := message
	if appEnv := os.Getenv("APP_ENV"); (appEnv == "production" || appEnv == "prod") && code >= fiber.StatusInternalServerError {
		clientMessage = "Internal Server Error"
	}

	return c.Status(code).JSON(fiber.Map{
		"error":      clientMessage,
		"request_id": logger.GetRequestID(),
	})
}
