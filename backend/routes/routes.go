package routes

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"github.com/liyali/liyali-gateway/handlers"
	"github.com/liyali/liyali-gateway/middleware"
	"github.com/liyali/liyali-gateway/services"
)

// SetupRoutes configures all API routes
func SetupRoutes(app *fiber.App, handlerRegistry *handlers.HandlerRegistry, rbacService *services.RBACService, db *gorm.DB) {
	// Health check (no versioning)
	app.Get("/health", handlers.HealthCheck)

	// API v1 - Version 1 routes
	apiV1 := app.Group("/api/v1")

	// Public routes (no authentication required)
	public := apiV1.Group("")
	public.Post("/auth/login", handlerRegistry.Auth.Login)
	public.Post("/auth/register", handlerRegistry.Auth.Register)
	public.Post("/auth/verify", handlerRegistry.Auth.VerifyToken)
	public.Post("/auth/refresh", handlerRegistry.Auth.RefreshToken)
	public.Post("/auth/password-reset/request", handlerRegistry.Auth.RequestPasswordReset)
	public.Post("/auth/password-reset/confirm", handlerRegistry.Auth.ResetPassword)

	// Protected routes (authentication required)
	protected := apiV1.Group("", middleware.AuthMiddleware())

	// Auth routes (protected, no tenant required)
	protected.Get("/auth/profile", handlerRegistry.Auth.GetProfile)
	protected.Post("/auth/logout", handlerRegistry.Auth.Logout)
	protected.Post("/auth/logout-all", handlerRegistry.Auth.LogoutAll)
	protected.Post("/auth/change-password", handlerRegistry.Auth.ChangePassword)

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

	// Organization role management (Phase 3.5) - ENABLED
	orgRoles := tenant.Group("/organization/roles")
	orgRoles.Get("/",
		middleware.RequirePermission(rbacService, "organization", "manage_workflows"),
		handlers.GetOrganizationRoles)
	orgRoles.Post("/",
		middleware.RequirePermission(rbacService, "organization", "manage_workflows"),
		handlers.CreateOrganizationRole)
	orgRoles.Put("/:roleId",
		middleware.RequirePermission(rbacService, "organization", "manage_workflows"),
		handlers.UpdateOrganizationRole)
	orgRoles.Delete("/:roleId",
		middleware.RequirePermission(rbacService, "organization", "manage_workflows"),
		handlers.DeleteOrganizationRole)
	orgRoles.Get("/:roleId/permissions",
		middleware.RequirePermission(rbacService, "organization", "manage_workflows"),
		handlers.GetRolePermissions)
	orgRoles.Post("/:roleId/permissions/:permissionId",
		middleware.RequirePermission(rbacService, "organization", "manage_workflows"),
		handlers.AssignPermissionToRole)
	orgRoles.Delete("/:roleId/permissions/:permissionId",
		middleware.RequirePermission(rbacService, "organization", "manage_workflows"),
		handlers.RemovePermissionFromRole)

	// Organization permissions (Phase 3.5) - ENABLED
	permissions := tenant.Group("/organization/permissions")
	permissions.Get("/",
		middleware.RequirePermission(rbacService, "organization", "manage_workflows"),
		handlers.GetOrganizationPermissions)

	// Requisition routes (tenant-scoped)
	requisitions := tenant.Group("/requisitions")
	requisitions.Get("/", middleware.RequirePermission(rbacService, "requisition", "view"), handlers.GetRequisitions)
	requisitions.Post("/", middleware.RequirePermission(rbacService, "requisition", "create"), handlers.CreateRequisition)
	requisitions.Get("/:id", middleware.RequirePermission(rbacService, "requisition", "view"), handlers.GetRequisition)
	requisitions.Put("/:id", middleware.RequirePermission(rbacService, "requisition", "edit"), handlers.UpdateRequisition)
	requisitions.Delete("/:id", middleware.RequirePermission(rbacService, "requisition", "delete"), handlers.DeleteRequisition)
	requisitions.Post("/:id/approve", middleware.RequirePermission(rbacService, "requisition", "approve"), handlers.ApproveRequisition)
	requisitions.Post("/:id/reject", middleware.RequirePermission(rbacService, "requisition", "reject"), handlers.RejectRequisition)
	requisitions.Post("/:id/reassign", middleware.RequirePermission(rbacService, "requisition", "approve"), handlers.ReassignRequisition)

	// Budget routes (tenant-scoped)
	budgets := tenant.Group("/budgets")
	budgets.Get("/", middleware.RequirePermission(rbacService, "budget", "view"), handlers.GetBudgets)
	budgets.Post("/", middleware.RequirePermission(rbacService, "budget", "create"), handlers.CreateBudget)
	budgets.Get("/:id", middleware.RequirePermission(rbacService, "budget", "view"), handlers.GetBudget)
	budgets.Put("/:id", middleware.RequirePermission(rbacService, "budget", "edit"), handlers.UpdateBudget)
	budgets.Delete("/:id", middleware.RequirePermission(rbacService, "budget", "delete"), handlers.DeleteBudget)
	budgets.Post("/:id/approve", middleware.RequirePermission(rbacService, "budget", "approve"), handlers.ApproveBudget)
	budgets.Post("/:id/reject", middleware.RequirePermission(rbacService, "budget", "reject"), handlers.RejectBudget)

	// Purchase Order routes (tenant-scoped)
	pos := tenant.Group("/purchase-orders")
	pos.Get("/", middleware.RequirePermission(rbacService, "purchase_order", "view"), handlers.GetPurchaseOrders)
	pos.Post("/", middleware.RequirePermission(rbacService, "purchase_order", "create"), handlers.CreatePurchaseOrder)
	pos.Get("/:id", middleware.RequirePermission(rbacService, "purchase_order", "view"), handlers.GetPurchaseOrder)
	pos.Put("/:id", middleware.RequirePermission(rbacService, "purchase_order", "edit"), handlers.UpdatePurchaseOrder)
	pos.Delete("/:id", middleware.RequirePermission(rbacService, "purchase_order", "delete"), handlers.DeletePurchaseOrder)
	pos.Post("/:id/approve", middleware.RequirePermission(rbacService, "purchase_order", "approve"), handlers.ApprovePurchaseOrder)
	pos.Post("/:id/reject", middleware.RequirePermission(rbacService, "purchase_order", "reject"), handlers.RejectPurchaseOrder)

	// Payment Voucher routes (tenant-scoped)
	pvs := tenant.Group("/payment-vouchers")
	pvs.Get("/", middleware.RequirePermission(rbacService, "payment_voucher", "view"), handlers.GetPaymentVouchers)
	pvs.Post("/", middleware.RequirePermission(rbacService, "payment_voucher", "create"), handlers.CreatePaymentVoucher)
	pvs.Get("/:id", middleware.RequirePermission(rbacService, "payment_voucher", "view"), handlers.GetPaymentVoucher)
	pvs.Put("/:id", middleware.RequirePermission(rbacService, "payment_voucher", "edit"), handlers.UpdatePaymentVoucher)
	pvs.Delete("/:id", middleware.RequirePermission(rbacService, "payment_voucher", "delete"), handlers.DeletePaymentVoucher)
	pvs.Post("/:id/approve", middleware.RequirePermission(rbacService, "payment_voucher", "approve"), handlers.ApprovePaymentVoucher)
	pvs.Post("/:id/reject", middleware.RequirePermission(rbacService, "payment_voucher", "reject"), handlers.RejectPaymentVoucher)

	// GRN routes (tenant-scoped)
	grns := tenant.Group("/grns")
	grns.Get("/", middleware.RequirePermission(rbacService, "grn", "view"), handlers.GetGRNs)
	grns.Post("/", middleware.RequirePermission(rbacService, "grn", "create"), handlers.CreateGRN)
	grns.Get("/:id", middleware.RequirePermission(rbacService, "grn", "view"), handlers.GetGRN)
	grns.Put("/:id", middleware.RequirePermission(rbacService, "grn", "edit"), handlers.UpdateGRN)
	grns.Delete("/:id", middleware.RequirePermission(rbacService, "grn", "delete"), handlers.DeleteGRN)
	grns.Post("/:id/approve", middleware.RequirePermission(rbacService, "grn", "approve"), handlers.ApproveGRN)
	grns.Post("/:id/reject", middleware.RequirePermission(rbacService, "grn", "reject"), handlers.RejectGRN)

	// Category routes (tenant-scoped)
	categories := tenant.Group("/categories")
	categories.Get("/", middleware.RequirePermission(rbacService, "category", "view"), handlers.GetCategories)
	categories.Post("/", middleware.RequirePermission(rbacService, "category", "create"), handlers.CreateCategory)
	categories.Get("/:id", middleware.RequirePermission(rbacService, "category", "view"), handlers.GetCategory)
	categories.Put("/:id", middleware.RequirePermission(rbacService, "category", "edit"), handlers.UpdateCategory)
	categories.Delete("/:id", middleware.RequirePermission(rbacService, "category", "delete"), handlers.DeleteCategory)
	categories.Get("/:id/budget-codes", middleware.RequirePermission(rbacService, "category", "view"), handlers.GetCategoryBudgetCodes)
	categories.Post("/:id/budget-codes", middleware.RequirePermission(rbacService, "category", "edit"), handlers.AddBudgetCodeToCategory)
	categories.Delete("/:id/budget-codes/:budgetCode", middleware.RequirePermission(rbacService, "category", "edit"), handlers.RemoveBudgetCodeFromCategory)

	// Vendor routes (tenant-scoped)
	vendors := tenant.Group("/vendors")
	vendors.Get("/", middleware.RequirePermission(rbacService, "vendor", "view"), handlers.GetVendors)
	vendors.Post("/", middleware.RequirePermission(rbacService, "vendor", "create"), handlers.CreateVendor)
	vendors.Get("/:id", middleware.RequirePermission(rbacService, "vendor", "view"), handlers.GetVendor)
	vendors.Put("/:id", middleware.RequirePermission(rbacService, "vendor", "edit"), handlers.UpdateVendor)

	// Approval Tasks routes (tenant-scoped) - Updated to use new handler
	approvals := tenant.Group("/approvals")
	approvals.Get("/", handlerRegistry.Approval.GetApprovalTasks)
	approvals.Get("/:id", handlerRegistry.Approval.GetApprovalTask)
	approvals.Post("/:id/approve", middleware.RequirePermission(rbacService, "approval", "approve"), handlerRegistry.Approval.ApproveTask)
	approvals.Post("/:id/reject", middleware.RequirePermission(rbacService, "approval", "reject"), handlerRegistry.Approval.RejectTask)
	approvals.Post("/:id/reassign", middleware.RequirePermission(rbacService, "approval", "reassign"), handlerRegistry.Approval.ReassignTask)
	approvals.Get("/tasks/overdue", middleware.RequirePermission(rbacService, "approval", "view"), handlerRegistry.Approval.GetOverdueTasks)

	// Bulk approval operations (tenant-scoped) - ENABLED
	bulk := approvals.Group("/bulk")
	bulk.Post("/approve", middleware.RequirePermission(rbacService, "approval", "approve"), handlerRegistry.Approval.BulkApprove)
	bulk.Post("/reject", middleware.RequirePermission(rbacService, "approval", "reject"), handlerRegistry.Approval.BulkReject)
	bulk.Post("/reassign", middleware.RequirePermission(rbacService, "approval", "reassign"), handlerRegistry.Approval.BulkReassign)

	// Approval history routes (tenant-scoped) - Updated to use new handler
	documents := tenant.Group("/documents")
	documents.Get("/:documentId/approval-history", handlerRegistry.Approval.GetApprovalHistory)

	// Generic Document System routes (tenant-scoped) - NEW
	genericDocs := tenant.Group("/documents")
	genericDocs.Get("/", middleware.RequirePermission(rbacService, "document", "view"), handlerRegistry.Document.GetDocuments)
	genericDocs.Get("/my", middleware.RequirePermission(rbacService, "document", "view"), handlerRegistry.Document.GetMyDocuments)
	genericDocs.Get("/search", middleware.RequirePermission(rbacService, "document", "view"), handlerRegistry.Document.SearchDocuments)
	genericDocs.Get("/stats", middleware.RequirePermission(rbacService, "document", "view"), handlerRegistry.Document.GetDocumentStats)
	genericDocs.Get("/:id", middleware.RequirePermission(rbacService, "document", "view"), handlerRegistry.Document.GetDocumentByID)
	genericDocs.Get("/number/:number", middleware.RequirePermission(rbacService, "document", "view"), handlerRegistry.Document.GetDocumentByNumber)
	genericDocs.Post("/", middleware.RequirePermission(rbacService, "document", "create"), handlerRegistry.Document.CreateDocument)
	genericDocs.Put("/:id", middleware.RequirePermission(rbacService, "document", "edit"), handlerRegistry.Document.UpdateDocument)
	genericDocs.Post("/:id/submit", middleware.RequirePermission(rbacService, "document", "submit"), handlerRegistry.Document.SubmitDocument)
	genericDocs.Delete("/:id", middleware.RequirePermission(rbacService, "document", "delete"), handlerRegistry.Document.DeleteDocument)

	// Workflow routes (tenant-scoped) - ENABLED
	workflows := tenant.Group("/workflows")
	workflows.Get("/", middleware.RequirePermission(rbacService, "workflow", "view"), handlerRegistry.Workflow.GetWorkflows)
	workflows.Get("/:id", middleware.RequirePermission(rbacService, "workflow", "view"), handlerRegistry.Workflow.GetWorkflowByID)
	workflows.Get("/default/:documentType", middleware.RequirePermission(rbacService, "workflow", "view"), handlerRegistry.Workflow.GetDefaultWorkflow)
	workflows.Post("/", middleware.RequirePermission(rbacService, "workflow", "create"), handlerRegistry.Workflow.CreateWorkflow)
	workflows.Put("/:id", middleware.RequirePermission(rbacService, "workflow", "edit"), handlerRegistry.Workflow.UpdateWorkflow)
	workflows.Post("/:id/activate", middleware.RequirePermission(rbacService, "workflow", "manage"), handlerRegistry.Workflow.ActivateWorkflow)
	workflows.Post("/:id/deactivate", middleware.RequirePermission(rbacService, "workflow", "manage"), handlerRegistry.Workflow.DeactivateWorkflow)
	workflows.Delete("/:id", middleware.RequirePermission(rbacService, "workflow", "delete"), handlerRegistry.Workflow.DeleteWorkflow)

	// MVP Workflow routes (tenant-scoped) - NEW MVP SYSTEM
	workflowsMVP := tenant.Group("/workflows-mvp")
	workflowsMVPHandler := handlers.NewWorkflowMVPHandler(db) // We'll need to pass db here
	workflowsMVP.Get("/", middleware.RequirePermission(rbacService, "workflow", "view"), workflowsMVPHandler.GetWorkflows)
	workflowsMVP.Get("/:id", middleware.RequirePermission(rbacService, "workflow", "view"), workflowsMVPHandler.GetWorkflowByID)
	workflowsMVP.Get("/default/:entityType", middleware.RequirePermission(rbacService, "workflow", "view"), workflowsMVPHandler.GetDefaultWorkflow)
	workflowsMVP.Post("/", middleware.RequirePermission(rbacService, "workflow", "create"), workflowsMVPHandler.CreateWorkflow)
	workflowsMVP.Put("/:id", middleware.RequirePermission(rbacService, "workflow", "edit"), workflowsMVPHandler.UpdateWorkflow)
	workflowsMVP.Delete("/:id", middleware.RequirePermission(rbacService, "workflow", "delete"), workflowsMVPHandler.DeleteWorkflow)
	workflowsMVP.Post("/:id/duplicate", middleware.RequirePermission(rbacService, "workflow", "create"), workflowsMVPHandler.DuplicateWorkflow)
	workflowsMVP.Post("/:id/set-default", middleware.RequirePermission(rbacService, "workflow", "manage"), workflowsMVPHandler.SetDefaultWorkflow)
	workflowsMVP.Post("/resolve", middleware.RequirePermission(rbacService, "workflow", "view"), workflowsMVPHandler.ResolveWorkflow)
	workflowsMVP.Get("/:id/usage", middleware.RequirePermission(rbacService, "workflow", "view"), workflowsMVPHandler.GetWorkflowUsage)
	workflowsMVP.Post("/validate", middleware.RequirePermission(rbacService, "workflow", "create"), workflowsMVPHandler.ValidateWorkflow)

	// Analytics routes (tenant-scoped) - ENABLED
	analytics := tenant.Group("/analytics")
	analytics.Get("/dashboard", middleware.RequirePermission(rbacService, "analytics", "view"), handlers.GetDashboard)
	analytics.Get("/requisitions/metrics", middleware.RequirePermission(rbacService, "analytics", "view"), handlers.GetRequisitionMetrics)
	analytics.Get("/approvals/metrics", middleware.RequirePermission(rbacService, "analytics", "view"), handlers.GetApprovalMetrics)

	// Notifications (tenant-scoped) - ENABLED
	notifications := tenant.Group("/notifications")
	notifications.Get("/", handlers.GetNotifications)
	notifications.Get("/:id", handlers.GetNotification)
	notifications.Put("/:id/read", handlers.MarkNotificationAsRead)
	notifications.Put("/read-all", handlers.MarkAllNotificationsAsRead)
	notifications.Get("/stats", handlers.GetNotificationStats)
	notifications.Delete("/:id", handlers.DeleteNotification)

	// Audit Logs (tenant-scoped) - ENABLED
	audit := tenant.Group("/audit-logs")
	audit.Get("/", middleware.RequirePermission(rbacService, "audit_log", "view"), handlers.GetAuditLogs)
	audit.Get("/document/:documentId", middleware.RequirePermission(rbacService, "audit_log", "view"), handlers.GetDocumentAuditLogs)
}
