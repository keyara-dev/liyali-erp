package routes

import (
	"github.com/gofiber/fiber/v3"
	"github.com/liyali/liyali-gateway/config"
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

	// Auth routes (protected, no tenant required)
	protected.Get("/auth/profile", handlers.GetProfile)

	// Organization routes (authentication required, no tenant middleware)
	orgs := protected.Group("/organizations")
	orgs.Get("/", handlers.GetUserOrganizations)
	orgs.Post("/", handlers.CreateOrganization)
	orgs.Put("/:id", handlers.UpdateOrganization)
	orgs.Post("/:id/switch", handlers.SwitchOrganization)

	// Tenant-scoped routes (authentication + tenant context required)
	tenant := apiV1.Group("", middleware.AuthMiddleware(), middleware.TenantMiddleware())

	// Organization management (within tenant context)
	orgMgmt := tenant.Group("/organization")
	orgMgmt.Get("/members", handlers.GetOrganizationMembers)
	orgMgmt.Post("/members", handlers.AddOrganizationMember)
	orgMgmt.Delete("/members/:userId", handlers.RemoveOrganizationMember)
	orgMgmt.Get("/settings", handlers.GetOrganizationSettings)
	orgMgmt.Put("/settings", handlers.UpdateOrganizationSettings)

	// Organization role management (Phase 3.5)
	orgRoles := tenant.Group("/organization/roles")
	orgRoles.Get("/",
		middleware.RequirePermission(config.DB, "organization", "manage_workflows"),
		handlers.GetOrganizationRoles)
	orgRoles.Post("/",
		middleware.RequirePermission(config.DB, "organization", "manage_workflows"),
		handlers.CreateOrganizationRole)
	orgRoles.Put("/:roleId",
		middleware.RequirePermission(config.DB, "organization", "manage_workflows"),
		handlers.UpdateOrganizationRole)
	orgRoles.Delete("/:roleId",
		middleware.RequirePermission(config.DB, "organization", "manage_workflows"),
		handlers.DeleteOrganizationRole)
	orgRoles.Get("/:roleId/permissions",
		middleware.RequirePermission(config.DB, "organization", "manage_workflows"),
		handlers.GetRolePermissions)
	orgRoles.Post("/:roleId/permissions/:permissionId",
		middleware.RequirePermission(config.DB, "organization", "manage_workflows"),
		handlers.AssignPermissionToRole)
	orgRoles.Delete("/:roleId/permissions/:permissionId",
		middleware.RequirePermission(config.DB, "organization", "manage_workflows"),
		handlers.RemovePermissionFromRole)

	// Organization permissions (Phase 3.5)
	permissions := tenant.Group("/organization/permissions")
	permissions.Get("/",
		middleware.RequirePermission(config.DB, "organization", "manage_workflows"),
		handlers.GetOrganizationPermissions)

	// User routes (now tenant-scoped)
	users := tenant.Group("/users")
	users.Get("/", handlers.GetUsers)
	users.Get("/:id", handlers.GetUser)
	users.Put("/:id", handlers.UpdateUser)

	// Requisition routes (tenant-scoped)
	requisitions := tenant.Group("/requisitions")
	requisitions.Get("/", middleware.RequirePermission(config.DB, "requisition", "view"), handlers.GetRequisitions)
	requisitions.Post("/", middleware.RequirePermission(config.DB, "requisition", "create"), handlers.CreateRequisition)
	requisitions.Get("/:id", middleware.RequirePermission(config.DB, "requisition", "view"), handlers.GetRequisition)
	requisitions.Put("/:id", middleware.RequirePermission(config.DB, "requisition", "edit"), handlers.UpdateRequisition)
	requisitions.Delete("/:id", middleware.RequirePermission(config.DB, "requisition", "delete"), handlers.DeleteRequisition)
	requisitions.Post("/:id/approve", middleware.RequirePermission(config.DB, "requisition", "approve"), handlers.ApproveRequisition)
	requisitions.Post("/:id/reject", middleware.RequirePermission(config.DB, "requisition", "reject"), handlers.RejectRequisition)
	requisitions.Post("/:id/reassign", middleware.RequirePermission(config.DB, "requisition", "approve"), handlers.ReassignRequisition)

	// Budget routes (tenant-scoped)
	budgets := tenant.Group("/budgets")
	budgets.Get("/", middleware.RequirePermission(config.DB, "budget", "view"), handlers.GetBudgets)
	budgets.Post("/", middleware.RequirePermission(config.DB, "budget", "create"), handlers.CreateBudget)
	budgets.Get("/:id", middleware.RequirePermission(config.DB, "budget", "view"), handlers.GetBudget)
	budgets.Put("/:id", middleware.RequirePermission(config.DB, "budget", "edit"), handlers.UpdateBudget)
	budgets.Delete("/:id", middleware.RequirePermission(config.DB, "budget", "delete"), handlers.DeleteBudget)
	budgets.Post("/:id/approve", middleware.RequirePermission(config.DB, "budget", "approve"), handlers.ApproveBudget)
	budgets.Post("/:id/reject", middleware.RequirePermission(config.DB, "budget", "reject"), handlers.RejectBudget)

	// Purchase Order routes (tenant-scoped)
	pos := tenant.Group("/purchase-orders")
	pos.Get("/", middleware.RequirePermission(config.DB, "purchase_order", "view"), handlers.GetPurchaseOrders)
	pos.Post("/", middleware.RequirePermission(config.DB, "purchase_order", "create"), handlers.CreatePurchaseOrder)
	pos.Get("/:id", middleware.RequirePermission(config.DB, "purchase_order", "view"), handlers.GetPurchaseOrder)
	pos.Put("/:id", middleware.RequirePermission(config.DB, "purchase_order", "edit"), handlers.UpdatePurchaseOrder)
	pos.Delete("/:id", middleware.RequirePermission(config.DB, "purchase_order", "delete"), handlers.DeletePurchaseOrder)
	pos.Post("/:id/approve", middleware.RequirePermission(config.DB, "purchase_order", "approve"), handlers.ApprovePurchaseOrder)
	pos.Post("/:id/reject", middleware.RequirePermission(config.DB, "purchase_order", "reject"), handlers.RejectPurchaseOrder)

	// Payment Voucher routes (tenant-scoped)
	pvs := tenant.Group("/payment-vouchers")
	pvs.Get("/", middleware.RequirePermission(config.DB, "payment_voucher", "view"), handlers.GetPaymentVouchers)
	pvs.Post("/", middleware.RequirePermission(config.DB, "payment_voucher", "create"), handlers.CreatePaymentVoucher)
	pvs.Get("/:id", middleware.RequirePermission(config.DB, "payment_voucher", "view"), handlers.GetPaymentVoucher)
	pvs.Put("/:id", middleware.RequirePermission(config.DB, "payment_voucher", "edit"), handlers.UpdatePaymentVoucher)
	pvs.Delete("/:id", middleware.RequirePermission(config.DB, "payment_voucher", "delete"), handlers.DeletePaymentVoucher)
	pvs.Post("/:id/approve", middleware.RequirePermission(config.DB, "payment_voucher", "approve"), handlers.ApprovePaymentVoucher)
	pvs.Post("/:id/reject", middleware.RequirePermission(config.DB, "payment_voucher", "reject"), handlers.RejectPaymentVoucher)

	// GRN routes (tenant-scoped)
	grns := tenant.Group("/grns")
	grns.Get("/", middleware.RequirePermission(config.DB, "grn", "view"), handlers.GetGRNs)
	grns.Post("/", middleware.RequirePermission(config.DB, "grn", "create"), handlers.CreateGRN)
	grns.Get("/:id", middleware.RequirePermission(config.DB, "grn", "view"), handlers.GetGRN)
	grns.Put("/:id", middleware.RequirePermission(config.DB, "grn", "edit"), handlers.UpdateGRN)
	grns.Delete("/:id", middleware.RequirePermission(config.DB, "grn", "delete"), handlers.DeleteGRN)
	grns.Post("/:id/approve", middleware.RequirePermission(config.DB, "grn", "approve"), handlers.ApproveGRN)
	grns.Post("/:id/reject", middleware.RequirePermission(config.DB, "grn", "reject"), handlers.RejectGRN)

	// Category routes (tenant-scoped)
	categories := tenant.Group("/categories")
	categories.Get("/", middleware.RequirePermission(config.DB, "category", "view"), handlers.GetCategories)
	categories.Post("/", middleware.RequirePermission(config.DB, "category", "create"), handlers.CreateCategory)
	categories.Get("/:id", middleware.RequirePermission(config.DB, "category", "view"), handlers.GetCategory)
	categories.Put("/:id", middleware.RequirePermission(config.DB, "category", "edit"), handlers.UpdateCategory)
	categories.Delete("/:id", middleware.RequirePermission(config.DB, "category", "delete"), handlers.DeleteCategory)
	categories.Get("/:id/budget-codes", middleware.RequirePermission(config.DB, "category", "view"), handlers.GetCategoryBudgetCodes)
	categories.Post("/:id/budget-codes", middleware.RequirePermission(config.DB, "category", "edit"), handlers.AddBudgetCodeToCategory)
	categories.Delete("/:id/budget-codes/:budgetCode", middleware.RequirePermission(config.DB, "category", "edit"), handlers.RemoveBudgetCodeFromCategory)

	// Vendor routes (tenant-scoped)
	vendors := tenant.Group("/vendors")
	vendors.Get("/", middleware.RequirePermission(config.DB, "vendor", "view"), handlers.GetVendors)
	vendors.Post("/", middleware.RequirePermission(config.DB, "vendor", "create"), handlers.CreateVendor)
	vendors.Get("/:id", middleware.RequirePermission(config.DB, "vendor", "view"), handlers.GetVendor)
	vendors.Put("/:id", middleware.RequirePermission(config.DB, "vendor", "edit"), handlers.UpdateVendor)

	// Approval Tasks routes (tenant-scoped)
	approvals := tenant.Group("/approvals")
	approvals.Get("/", handlers.GetApprovalTasks)
	approvals.Get("/:id", handlers.GetApprovalTask)
	approvals.Get("/pending/:userId", handlers.GetPendingApprovals)

	// Bulk operations (tenant-scoped)
	bulk := tenant.Group("/bulk")
	bulk.Post("/approve", handlers.BulkApprove)
	bulk.Post("/reject", handlers.BulkReject)
	bulk.Post("/reassign", handlers.BulkReassign)

	// Analytics routes (tenant-scoped)
	analytics := tenant.Group("/analytics")
	analytics.Get("/dashboard", middleware.RequirePermission(config.DB, "analytics", "view"), handlers.GetDashboard)
	analytics.Get("/requisitions/metrics", middleware.RequirePermission(config.DB, "analytics", "view"), handlers.GetRequisitionMetrics)
	analytics.Get("/approvals/metrics", middleware.RequirePermission(config.DB, "analytics", "view"), handlers.GetApprovalMetrics)

	// Notifications (tenant-scoped)
	notifications := tenant.Group("/notifications")
	notifications.Get("/", handlers.GetNotifications)
	notifications.Get("/:id", handlers.GetNotification)
	notifications.Put("/:id/read", handlers.MarkNotificationAsRead)

	// Audit Logs (tenant-scoped)
	audit := tenant.Group("/audit-logs")
	audit.Get("/", middleware.RequirePermission(config.DB, "audit_log", "view"), handlers.GetAuditLogs)
	audit.Get("/document/:documentId", middleware.RequirePermission(config.DB, "audit_log", "view"), handlers.GetDocumentAuditLogs)
}
