package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/handlers"
	"github.com/liyali/liyali-gateway/middleware"
	"github.com/liyali/liyali-gateway/routes"
	"github.com/liyali/liyali-gateway/services"
	"github.com/liyali/liyali-gateway/repository"
)

func init() {
	// Load environment variables
	err := godotenv.Load(".env")
	if err != nil && os.Getenv("APP_ENV") == "" {
		log.Println("Note: .env file not found, using environment variables")
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
	if os.Getenv("DB_SSL_MODE") == "" {
		os.Setenv("DB_SSL_MODE", "disable")
	}
	if os.Getenv("APP_PORT") == "" {
		os.Setenv("APP_PORT", "8080")
	}
	if os.Getenv("FRONTEND_URL") == "" {
		os.Setenv("FRONTEND_URL", "http://localhost:3000")
	}
	if os.Getenv("JWT_SECRET") == "" {
		os.Setenv("JWT_SECRET", "your-super-secret-jwt-key-change-in-production")
	}
}

func main() {
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
	documentService := services.NewDocumentService(documentRepo, auditService)

	// Initialize handler registry
	handlerRegistry := handlers.NewHandlerRegistry(authService, rbacService, workflowService, documentService)

	// Create Fiber app with global error handler
	app := fiber.New(fiber.Config{
		AppName:      "Liyali Gateway Backend API",
		ErrorHandler: customErrorHandler,
	})

	// Middleware
	app.Use(middleware.ErrorHandlingMiddleware())
	app.Use(middleware.LoggerMiddleware())
	app.Use(middleware.CORSMiddleware())

	// Setup routes with handler registry
	routes.SetupRoutes(app, handlerRegistry, rbacService, config.DB)

	// Start server with graceful shutdown
	go func() {
		port := os.Getenv("APP_PORT")
		log.Printf("🚀 Starting Enhanced Liyali Gateway Backend on port %s", port)
		log.Printf("📊 Features: Enhanced Auth, Session Management, Custom RBAC, Workflow Engine")
		log.Printf("🔐 Security: Account Lockout, Password Reset, Audit Logging")
		log.Printf("🏗️  Architecture: Clean Architecture (Repository → Service → Handler)")
		log.Printf("💾 Database: GORM + pgx with sqlc for type-safe queries")
		log.Printf("🔄 Workflows: Dynamic workflow management with bulk operations")
		log.Printf("❤️  Health check: http://localhost:%s/health", port)

		if err := app.Listen(":" + port); err != nil {
			log.Fatalf("Error starting server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down server...")
	if err := app.ShutdownWithContext(context.Background()); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("✅ Server stopped gracefully")
}

// customErrorHandler handles errors globally
func customErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	message := "Internal Server Error"

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
		message = e.Message
	}

	return c.Status(code).JSON(fiber.Map{
		"error": message,
	})
}
