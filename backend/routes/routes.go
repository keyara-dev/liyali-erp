package routes

import (
	"github.com/gofiber/fiber/v3"
	"github.com/liyali/liyali-gateway/handlers"
	"github.com/liyali/liyali-gateway/middleware"
)

// SetupRoutes configures all API routes
func SetupRoutes(app *fiber.App) {
	// Health check (no versioning)
	app.Get("/health", handlers.HealthCheck)

	// API v1 - Version 1 routes
	apiV1 := app.Group("/api/v1")

	// Public routes (no authentication required)
	public := apiV1.Group("")
	public.Post("/auth/login", handlers.Login)
	public.Post("/auth/register", handlers.Register)
	public.Post("/auth/verify", handlers.VerifyToken)
	public.Post("/auth/refresh", handlers.RefreshToken)

	// Protected routes (authentication required)
	protected := apiV1.Group("", middleware.AuthMiddleware())

	// Auth routes (protected)
	protected.Get("/auth/profile", handlers.GetProfile)

	// User routes
	users := protected.Group("/users")
	users.Get("/", handlers.GetUsers)
	users.Get("/:id", handlers.GetUser)
	users.Put("/:id", handlers.UpdateUser)

	// Requisition routes
	requisitions := protected.Group("/requisitions")
	requisitions.Get("/", handlers.GetRequisitions)
	requisitions.Post("/", handlers.CreateRequisition)
	requisitions.Get("/:id", handlers.GetRequisition)
	requisitions.Put("/:id", handlers.UpdateRequisition)
	requisitions.Delete("/:id", handlers.DeleteRequisition)
	requisitions.Post("/:id/approve", handlers.ApproveRequisition)
	requisitions.Post("/:id/reject", handlers.RejectRequisition)
	requisitions.Post("/:id/reassign", handlers.ReassignRequisition)

	// Budget routes
	budgets := protected.Group("/budgets")
	budgets.Get("/", handlers.GetBudgets)
	budgets.Post("/", handlers.CreateBudget)
	budgets.Get("/:id", handlers.GetBudget)
	budgets.Put("/:id", handlers.UpdateBudget)
	budgets.Delete("/:id", handlers.DeleteBudget)
	budgets.Post("/:id/approve", handlers.ApproveBudget)
	budgets.Post("/:id/reject", handlers.RejectBudget)

	// Purchase Order routes
	pos := protected.Group("/purchase-orders")
	pos.Get("/", handlers.GetPurchaseOrders)
	pos.Post("/", handlers.CreatePurchaseOrder)
	pos.Get("/:id", handlers.GetPurchaseOrder)
	pos.Put("/:id", handlers.UpdatePurchaseOrder)
	pos.Delete("/:id", handlers.DeletePurchaseOrder)
	pos.Post("/:id/approve", handlers.ApprovePurchaseOrder)
	pos.Post("/:id/reject", handlers.RejectPurchaseOrder)

	// Payment Voucher routes
	pvs := protected.Group("/payment-vouchers")
	pvs.Get("/", handlers.GetPaymentVouchers)
	pvs.Post("/", handlers.CreatePaymentVoucher)
	pvs.Get("/:id", handlers.GetPaymentVoucher)
	pvs.Put("/:id", handlers.UpdatePaymentVoucher)
	pvs.Delete("/:id", handlers.DeletePaymentVoucher)
	pvs.Post("/:id/approve", handlers.ApprovePaymentVoucher)
	pvs.Post("/:id/reject", handlers.RejectPaymentVoucher)

	// GRN routes
	grns := protected.Group("/grns")
	grns.Get("/", handlers.GetGRNs)
	grns.Post("/", handlers.CreateGRN)
	grns.Get("/:id", handlers.GetGRN)
	grns.Put("/:id", handlers.UpdateGRN)
	grns.Delete("/:id", handlers.DeleteGRN)
	grns.Post("/:id/approve", handlers.ApproveGRN)
	grns.Post("/:id/reject", handlers.RejectGRN)

	// Vendor routes
	vendors := protected.Group("/vendors")
	vendors.Get("/", handlers.GetVendors)
	vendors.Post("/", handlers.CreateVendor)
	vendors.Get("/:id", handlers.GetVendor)
	vendors.Put("/:id", handlers.UpdateVendor)

	// Approval Tasks routes
	approvals := protected.Group("/approvals")
	approvals.Get("/", handlers.GetApprovalTasks)
	approvals.Get("/:id", handlers.GetApprovalTask)
	approvals.Get("/pending/:userId", handlers.GetPendingApprovals)

	// Bulk operations
	bulk := protected.Group("/bulk")
	bulk.Post("/approve", handlers.BulkApprove)
	bulk.Post("/reject", handlers.BulkReject)
	bulk.Post("/reassign", handlers.BulkReassign)

	// Analytics routes
	analytics := protected.Group("/analytics")
	analytics.Get("/dashboard", handlers.GetDashboard)
	analytics.Get("/requisitions/metrics", handlers.GetRequisitionMetrics)
	analytics.Get("/approvals/metrics", handlers.GetApprovalMetrics)

	// Notifications
	notifications := protected.Group("/notifications")
	notifications.Get("/", handlers.GetNotifications)
	notifications.Get("/:id", handlers.GetNotification)
	notifications.Put("/:id/read", handlers.MarkNotificationAsRead)

	// Audit Logs
	audit := protected.Group("/audit-logs")
	audit.Get("/", handlers.GetAuditLogs)
	audit.Get("/document/:documentId", handlers.GetDocumentAuditLogs)
}
