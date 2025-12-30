package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/cozyCodr/liyali-gateway/internal/config"
	"github.com/cozyCodr/liyali-gateway/internal/db"
	"github.com/cozyCodr/liyali-gateway/internal/handlers"
	"github.com/cozyCodr/liyali-gateway/internal/middleware"
	"github.com/cozyCodr/liyali-gateway/internal/repository"
	"github.com/cozyCodr/liyali-gateway/internal/services"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to database
	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	// Test database connection
	if err := pool.Ping(context.Background()); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("✅ Connected to database")

	// Initialize sqlc queries
	queries := db.New(pool)

	// Initialize repositories
	userRepo := repository.NewUserRepository(queries)
	sessionRepo := repository.NewSessionRepository(queries)
	passwordResetRepo := repository.NewPasswordResetRepository(queries)
	workflowRepo := repository.NewWorkflowRepository(queries)
	documentRepo := repository.NewDocumentRepository(queries)
	approvalTaskRepo := repository.NewApprovalTaskRepository(queries)
	approvalHistoryRepo := repository.NewApprovalHistoryRepository(queries)
	auditLogRepo := repository.NewAuditLogRepository(queries)
	notificationRepo := repository.NewNotificationRepository(queries)

	// Initialize services
	authService := services.NewAuthService(userRepo, sessionRepo, passwordResetRepo, cfg.JWTSecret)
	approvalService := services.NewApprovalService(approvalTaskRepo, approvalHistoryRepo, documentRepo, auditLogRepo, notificationRepo)
	analyticsService := services.NewAnalyticsService(*documentRepo, *approvalTaskRepo, *approvalHistoryRepo, *workflowRepo)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(authService)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	approvalHandler := handlers.NewApprovalHandler(approvalService)
	workflowHandler := handlers.NewWorkflowHandler(workflowRepo)
	documentHandler := handlers.NewDocumentHandler(documentRepo)
	analyticsHandler := handlers.NewAnalyticsHandler(analyticsService)
	notificationHandler := handlers.NewNotificationHandler(*notificationRepo)
	auditLogHandler := handlers.NewAuditLogHandler(*auditLogRepo)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName:      "Liyali Gateway API",
		ErrorHandler: customErrorHandler,
	})

	// Global middleware
	app.Use(recover.New())
	app.Use(logger.New())

	// Parse allowed origins
	allowedOrigins := strings.Split(cfg.AllowedOrigins, ",")
	for i, origin := range allowedOrigins {
		allowedOrigins[i] = strings.TrimSpace(origin)
	}

	app.Use(cors.New(cors.Config{
		AllowOrigins: allowedOrigins,
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
	}))

	// Health check
	app.Get("/health", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "healthy",
			"service": "liyali-gateway-api",
		})
	})

	// API routes
	api := app.Group("/api")

	// Auth routes (public)
	auth := api.Group("/auth")
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)
	auth.Post("/refresh", authHandler.RefreshToken)
	auth.Post("/logout", authHandler.Logout)
	auth.Post("/password-reset/request", authHandler.RequestPasswordReset)
	auth.Post("/password-reset/confirm", authHandler.ResetPassword)

	// Protected auth routes (requires authentication)
	auth.Post("/change-password", authMiddleware.Authenticate, authHandler.ChangePassword)
	auth.Get("/me", authMiddleware.Authenticate, authHandler.GetCurrentUser)

	// Approval routes (protected)
	approvals := api.Group("/approvals", authMiddleware.Authenticate)
	approvals.Get("/tasks", approvalHandler.GetTasks)
	approvals.Get("/tasks/overdue", approvalHandler.GetOverdueTasks)
	approvals.Get("/tasks/:id", approvalHandler.GetTaskByID)
	approvals.Post("/tasks/:id/approve", approvalHandler.ApproveTask)
	approvals.Post("/tasks/:id/reject", approvalHandler.RejectTask)
	approvals.Post("/tasks/:id/reassign", approvalHandler.ReassignTask)
	approvals.Post("/tasks/:id/comment", approvalHandler.AddComment)

	// Bulk approval operations
	approvalsBulk := approvals.Group("/bulk")
	approvalsBulk.Post("/approve", approvalHandler.BulkApprove)
	approvalsBulk.Post("/reject", approvalHandler.BulkReject)
	approvalsBulk.Post("/reassign", approvalHandler.BulkReassign)

	// Workflow routes (protected)
	workflows := api.Group("/workflows", authMiddleware.Authenticate)
	workflows.Get("/", workflowHandler.GetWorkflows)
	workflows.Get("/:id", workflowHandler.GetWorkflowByID)
	workflows.Get("/default/:documentType", workflowHandler.GetDefaultWorkflow)
	workflows.Post("/", workflowHandler.CreateWorkflow)
	workflows.Put("/:id", workflowHandler.UpdateWorkflow)
	workflows.Post("/:id/activate", workflowHandler.ActivateWorkflow)
	workflows.Post("/:id/deactivate", workflowHandler.DeactivateWorkflow)
	workflows.Delete("/:id", workflowHandler.DeleteWorkflow)

	// Document routes (protected)
	documents := api.Group("/documents", authMiddleware.Authenticate)
	documents.Get("/", documentHandler.GetDocuments)
	documents.Get("/my", documentHandler.GetMyDocuments)
	documents.Get("/:id", documentHandler.GetDocumentByID)
	documents.Get("/number/:number", documentHandler.GetDocumentByNumber)
	documents.Post("/", documentHandler.CreateDocument)
	documents.Put("/:id", documentHandler.UpdateDocument)
	documents.Post("/:id/submit", documentHandler.SubmitDocument)
	documents.Delete("/:id", documentHandler.DeleteDocument)

	// Analytics routes (protected)
	analytics := api.Group("/analytics", authMiddleware.Authenticate)
	analytics.Get("/metrics", analyticsHandler.GetDashboardMetrics)
	analytics.Get("/trends", analyticsHandler.GetTrendData)
	analytics.Get("/bottlenecks", analyticsHandler.GetBottlenecks)

	// Notification routes (protected)
	notifications := api.Group("/notifications", authMiddleware.Authenticate)
	notifications.Get("/", notificationHandler.GetNotifications)
	notifications.Get("/unread", notificationHandler.GetUnreadNotifications)
	notifications.Get("/unread/count", notificationHandler.GetUnreadCount)
	notifications.Get("/:id", notificationHandler.GetNotificationByID)
	notifications.Post("/:id/read", notificationHandler.MarkAsRead)
	notifications.Post("/read-all", notificationHandler.MarkAllAsRead)
	notifications.Delete("/:id", notificationHandler.DeleteNotification)

	// Audit Log routes (protected - admin/manager only)
	auditLogs := api.Group("/audit-logs", authMiddleware.Authenticate)
	auditLogs.Get("/", auditLogHandler.GetAuditLogs)
	auditLogs.Get("/my", auditLogHandler.GetMyAuditLogs)
	auditLogs.Get("/:id", auditLogHandler.GetAuditLogByID)
	auditLogs.Get("/resource/:resource_type/:resource_id", auditLogHandler.GetAuditLogsByResource)

	// Start server with graceful shutdown
	go func() {
		addr := fmt.Sprintf(":%s", cfg.Port)
		log.Printf("🚀 Server starting on http://localhost%s", addr)
		log.Printf("📝 Environment: %s", cfg.Environment)
		log.Printf("🔐 Auth endpoints: http://localhost%s/api/auth/*", addr)
		log.Printf("❤️  Health check: http://localhost%s/health", addr)

		if err := app.Listen(addr); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down server...")
	if err := app.Shutdown(); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("✅ Server stopped gracefully")
}

// customErrorHandler handles errors globally
func customErrorHandler(c fiber.Ctx, err error) error {
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
