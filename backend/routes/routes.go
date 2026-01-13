package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/liyali/liyali-gateway/handlers"
	"github.com/liyali/liyali-gateway/middleware"
	"github.com/liyali/liyali-gateway/services"
	"gorm.io/gorm"
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
	orgs.Delete("/:id", handlers.DeleteOrganization)
	orgs.Post("/:id/switch", handlers.SwitchOrganization)

	// Tenant-scoped routes (authentication + tenant context required)
	tenant := apiV1.Group("", middleware.AuthMiddleware(), middleware.TenantMiddleware())

	// Organization management (within tenant context)
	orgMgmt := tenant.Group("/organization")
	orgMgmt.Get("/members",
		middleware.RequirePermission(rbacService, "organization", "view"),
		handlers.GetOrganizationMembers)
	orgMgmt.Post("/members",
		middleware.RequirePermission(rbacService, "organization", "manage"),
		handlers.AddOrganizationMember)
	orgMgmt.Delete("/members/:userId",
		middleware.RequirePermission(rbacService, "organization", "manage"),
		handlers.RemoveOrganizationMember)
	orgMgmt.Get("/settings",
		middleware.RequirePermission(rbacService, "organization", "view"),
		handlers.GetOrganizationSettings)
	orgMgmt.Put("/settings",
		middleware.RequirePermission(rbacService, "organization", "manage"),
		handlers.UpdateOrganizationSettings)

	// Organization departments (Phase 3.5) - NEW
	orgDepts := tenant.Group("/organization/departments")
	orgDepts.Get("/",
		middleware.RequirePermission(rbacService, "organization", "view"),
		handlers.GetOrganizationDepartments)
	orgDepts.Get("/:id",
		middleware.RequirePermission(rbacService, "organization", "view"),
		handlers.GetOrganizationDepartment)
	orgDepts.Post("/",
		middleware.RequirePermission(rbacService, "organization", "manage"),
		handlers.CreateOrganizationDepartment)
	orgDepts.Put("/:id",
		middleware.RequirePermission(rbacService, "organization", "manage"),
		handlers.UpdateOrganizationDepartment)
	orgDepts.Delete("/:id",
		middleware.RequirePermission(rbacService, "organization", "manage"),
		handlers.DeleteOrganizationDepartment)
	orgDepts.Post("/:id/restore",
		middleware.RequirePermission(rbacService, "organization", "manage"),
		handlers.RestoreOrganizationDepartment)
	orgDepts.Get("/:id/modules",
		middleware.RequirePermission(rbacService, "organization", "view"),
		handlers.GetDepartmentModules)
	orgDepts.Post("/:id/modules",
		middleware.RequirePermission(rbacService, "organization", "manage"),
		handlers.AssignModuleToDepartment)
	orgDepts.Delete("/:departmentId/modules/:moduleId",
		middleware.RequirePermission(rbacService, "organization", "manage"),
		handlers.RemoveModuleFromDepartment)
	orgDepts.Get("/:departmentId/users",
		middleware.RequirePermission(rbacService, "organization", "view"),
		handlers.GetDepartmentUsers)

	// User-Department Management (Phase 3.5) - NEW
	userDepts := tenant.Group("/users")
	userDepts.Post("/:userId/department/:departmentId",
		middleware.RequirePermission(rbacService, "organization", "manage"),
		handlers.AssignUserToDepartment)
	userDepts.Get("/:userId/department",
		middleware.RequirePermission(rbacService, "organization", "view"),
		handlers.GetUserDepartment)
	userDepts.Delete("/:userId/department",
		middleware.RequirePermission(rbacService, "organization", "manage"),
		handlers.RemoveUserFromDepartment)

	// Organization role management (Phase 3.5) - ENABLED
	orgRoles := tenant.Group("/organization/roles")
	orgRoles.Get("/",
		middleware.RequirePermission(rbacService, "organization", "manage"),
		handlers.GetOrganizationRoles)
	orgRoles.Post("/",
		middleware.RequirePermission(rbacService, "organization", "manage"),
		handlers.CreateOrganizationRole)
	orgRoles.Put("/:roleId",
		middleware.RequirePermission(rbacService, "organization", "manage"),
		handlers.UpdateOrganizationRole)
	orgRoles.Delete("/:roleId",
		middleware.RequirePermission(rbacService, "organization", "manage"),
		handlers.DeleteOrganizationRole)
	orgRoles.Get("/:roleId/permissions",
		middleware.RequirePermission(rbacService, "organization", "manage"),
		handlers.GetRolePermissions)
	orgRoles.Post("/:roleId/permissions/:permissionId",
		middleware.RequirePermission(rbacService, "organization", "manage"),
		handlers.AssignPermissionToRole)
	orgRoles.Delete("/:roleId/permissions/:permissionId",
		middleware.RequirePermission(rbacService, "organization", "manage"),
		handlers.RemovePermissionFromRole)

	// Organization permissions (Phase 3.5) - ENABLED
	permissions := tenant.Group("/organization/permissions")
	permissions.Get("/",
		middleware.RequirePermission(rbacService, "organization", "manage"),
		handlers.GetOrganizationPermissions)

	// User permission management (admin only) - ENABLED
	userPerms := tenant.Group("/users")
	userPerms.Get("/:userId/permissions",
		middleware.RequirePermission(rbacService, "user", "view"),
		handlers.GetUserPermissions)
	userPerms.Post("/:userId/permissions/:resource/:action",
		middleware.RequirePermission(rbacService, "organization", "manage"),
		handlers.GrantUserPermission)
	userPerms.Delete("/:userId/permissions/:resource/:action",
		middleware.RequirePermission(rbacService, "organization", "manage"),
		handlers.RevokeUserPermission)

	// System permissions list (admin only) - ENABLED
	systemPerms := tenant.Group("/permissions")
	systemPerms.Get("/",
		middleware.RequirePermission(rbacService, "organization", "view"),
		handlers.ListAllPermissions)

	// Requisition routes (tenant-scoped)
	requisitions := tenant.Group("/requisitions", middleware.InjectWorkflowExecutionService(handlerRegistry.WorkflowExecutionService))
	requisitions.Get("/", middleware.RequirePermission(rbacService, "requisition", "view"), handlers.GetRequisitions)
	requisitions.Post("/", middleware.RequirePermission(rbacService, "requisition", "create"), handlers.CreateRequisition)
	requisitions.Get("/:id", middleware.RequirePermission(rbacService, "requisition", "view"), handlers.GetRequisition)
	requisitions.Put("/:id", middleware.RequirePermission(rbacService, "requisition", "edit"), handlers.UpdateRequisition)
	requisitions.Delete("/:id", middleware.RequirePermission(rbacService, "requisition", "delete"), handlers.DeleteRequisition)
	requisitions.Post("/:id/submit", middleware.RequirePermission(rbacService, "requisition", "edit"), handlers.SubmitRequisition)
	requisitions.Post("/:id/reassign", middleware.RequirePermission(rbacService, "requisition", "approve"), handlers.ReassignRequisition)

	// Budget routes (tenant-scoped)
	budgets := tenant.Group("/budgets", middleware.InjectWorkflowExecutionService(handlerRegistry.WorkflowExecutionService))
	budgets.Get("/", middleware.RequirePermission(rbacService, "budget", "view"), handlers.GetBudgets)
	budgets.Post("/", middleware.RequirePermission(rbacService, "budget", "create"), handlers.CreateBudget)
	budgets.Get("/:id", middleware.RequirePermission(rbacService, "budget", "view"), handlers.GetBudget)
	budgets.Put("/:id", middleware.RequirePermission(rbacService, "budget", "edit"), handlers.UpdateBudget)
	budgets.Delete("/:id", middleware.RequirePermission(rbacService, "budget", "delete"), handlers.DeleteBudget)
	budgets.Post("/:id/submit", middleware.RequirePermission(rbacService, "budget", "edit"), handlers.SubmitBudget)

	// Purchase Order routes (tenant-scoped)
	pos := tenant.Group("/purchase-orders", middleware.InjectWorkflowExecutionService(handlerRegistry.WorkflowExecutionService))
	pos.Get("/", middleware.RequirePermission(rbacService, "purchase_order", "view"), handlers.GetPurchaseOrders)
	pos.Post("/", middleware.RequirePermission(rbacService, "purchase_order", "create"), handlers.CreatePurchaseOrder)
	pos.Get("/:id", middleware.RequirePermission(rbacService, "purchase_order", "view"), handlers.GetPurchaseOrder)
	pos.Put("/:id", middleware.RequirePermission(rbacService, "purchase_order", "edit"), handlers.UpdatePurchaseOrder)
	pos.Delete("/:id", middleware.RequirePermission(rbacService, "purchase_order", "delete"), handlers.DeletePurchaseOrder)
	pos.Post("/:id/submit", middleware.RequirePermission(rbacService, "purchase_order", "edit"), handlers.SubmitPurchaseOrder)

	// Payment Voucher routes (tenant-scoped)
	pvs := tenant.Group("/payment-vouchers", middleware.InjectWorkflowExecutionService(handlerRegistry.WorkflowExecutionService))
	pvs.Get("/", middleware.RequirePermission(rbacService, "payment_voucher", "view"), handlers.GetPaymentVouchers)
	pvs.Post("/", middleware.RequirePermission(rbacService, "payment_voucher", "create"), handlers.CreatePaymentVoucher)
	pvs.Get("/:id", middleware.RequirePermission(rbacService, "payment_voucher", "view"), handlers.GetPaymentVoucher)
	pvs.Put("/:id", middleware.RequirePermission(rbacService, "payment_voucher", "edit"), handlers.UpdatePaymentVoucher)
	pvs.Delete("/:id", middleware.RequirePermission(rbacService, "payment_voucher", "delete"), handlers.DeletePaymentVoucher)
	pvs.Post("/:id/submit", middleware.RequirePermission(rbacService, "payment_voucher", "edit"), handlers.SubmitPaymentVoucher)

	// GRN routes (tenant-scoped)
	grns := tenant.Group("/grns", middleware.InjectWorkflowExecutionService(handlerRegistry.WorkflowExecutionService))
	grns.Get("/", middleware.RequirePermission(rbacService, "grn", "view"), handlers.GetGRNs)
	grns.Post("/", middleware.RequirePermission(rbacService, "grn", "create"), handlers.CreateGRN)
	grns.Get("/:id", middleware.RequirePermission(rbacService, "grn", "view"), handlers.GetGRN)
	grns.Put("/:id", middleware.RequirePermission(rbacService, "grn", "edit"), handlers.UpdateGRN)
	grns.Delete("/:id", middleware.RequirePermission(rbacService, "grn", "delete"), handlers.DeleteGRN)
	grns.Post("/:id/submit", middleware.RequirePermission(rbacService, "grn", "edit"), handlers.SubmitGRN)

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
	approvals := tenant.Group("/approvals", middleware.InjectWorkflowExecutionService(handlerRegistry.WorkflowExecutionService))
	approvals.Get("/", handlerRegistry.Approval.GetApprovalTasks)

	// Specific routes must come before parameterized routes
	approvals.Get("/available-approvers", handlerRegistry.Approval.GetAvailableApprovers)
	approvals.Get("/tasks/overdue", middleware.RequirePermission(rbacService, "approval", "view"), handlerRegistry.Approval.GetOverdueTasks)

	// Task claiming routes (NEW)
	approvals.Post("/tasks/:id/claim", middleware.RequireWorkflowPermission("approve"), handlerRegistry.Approval.ClaimTask)
	approvals.Post("/tasks/:id/unclaim", middleware.RequireWorkflowPermission("approve"), handlerRegistry.Approval.UnclaimTask)

	// Parameterized routes come after specific routes
	approvals.Get("/:id", handlerRegistry.Approval.GetApprovalTask)
	approvals.Post("/:id/approve", middleware.RequireWorkflowPermission("approve"), handlerRegistry.Approval.ApproveTask)
	approvals.Post("/:id/reject", middleware.RequireWorkflowPermission("reject"), handlerRegistry.Approval.RejectTask)
	approvals.Post("/:id/reassign", middleware.RequirePermission(rbacService, "approval", "reassign"), handlerRegistry.Approval.ReassignTask)

	// Bulk approval operations (tenant-scoped) - ENABLED
	bulk := approvals.Group("/bulk")
	bulk.Post("/approve", middleware.RequireWorkflowPermission("approve"), handlerRegistry.Approval.BulkApprove)
	bulk.Post("/reject", middleware.RequireWorkflowPermission("reject"), handlerRegistry.Approval.BulkReject)
	bulk.Post("/reassign", middleware.RequirePermission(rbacService, "approval", "reassign"), handlerRegistry.Approval.BulkReassign)

	// Approval history routes (tenant-scoped) - Updated to use new handler
	documents := tenant.Group("/documents", middleware.InjectWorkflowExecutionService(handlerRegistry.WorkflowExecutionService))
	documents.Get("/:documentId/approval-history", handlerRegistry.Approval.GetApprovalHistory)
	documents.Get("/:documentId/approval-status", handlerRegistry.Approval.GetApprovalWorkflowStatus)

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

	// Workflow routes (tenant-scoped) - ENHANCED
	workflows := tenant.Group("/workflows")
	workflows.Get("/", middleware.RequirePermission(rbacService, "workflow", "view"), handlerRegistry.Workflow.GetWorkflows)
	workflows.Get("/:id", middleware.RequirePermission(rbacService, "workflow", "view"), handlerRegistry.Workflow.GetWorkflowByID)
	workflows.Get("/default/:documentType", middleware.RequirePermission(rbacService, "workflow", "view"), handlerRegistry.Workflow.GetDefaultWorkflow)
	workflows.Post("/", middleware.RequirePermission(rbacService, "workflow", "create"), handlerRegistry.Workflow.CreateWorkflow)
	workflows.Put("/:id", middleware.RequirePermission(rbacService, "workflow", "edit"), handlerRegistry.Workflow.UpdateWorkflow)
	workflows.Post("/:id/activate", middleware.RequirePermission(rbacService, "workflow", "manage"), handlerRegistry.Workflow.ActivateWorkflow)
	workflows.Post("/:id/deactivate", middleware.RequirePermission(rbacService, "workflow", "manage"), handlerRegistry.Workflow.DeactivateWorkflow)
	workflows.Delete("/:id", middleware.RequirePermission(rbacService, "workflow", "delete"), handlerRegistry.Workflow.DeleteWorkflow)

	// New frontend-compatible workflow endpoints
	workflows.Post("/:id/duplicate", middleware.RequirePermission(rbacService, "workflow", "create"), handlerRegistry.Workflow.DuplicateWorkflow)
	workflows.Post("/:id/set-default", middleware.RequirePermission(rbacService, "workflow", "manage"), handlerRegistry.Workflow.SetDefaultWorkflow)
	workflows.Post("/resolve", middleware.RequirePermission(rbacService, "workflow", "view"), handlerRegistry.Workflow.ResolveWorkflow)
	workflows.Get("/:id/usage", middleware.RequirePermission(rbacService, "workflow", "view"), handlerRegistry.Workflow.GetWorkflowUsage)
	workflows.Post("/validate", middleware.RequirePermission(rbacService, "workflow", "create"), handlerRegistry.Workflow.ValidateWorkflow)

	// Analytics routes (tenant-scoped) - ENABLED
	analytics := tenant.Group("/analytics")
	analytics.Get("/dashboard", middleware.RequirePermission(rbacService, "analytics", "view"), handlers.GetDashboard)
	analytics.Get("/requisitions/metrics", middleware.RequirePermission(rbacService, "analytics", "view"), handlers.GetRequisitionMetrics)
	analytics.Get("/approvals/metrics", middleware.RequirePermission(rbacService, "analytics", "view"), handlers.GetApprovalMetrics)

	// Notifications (tenant-scoped) - ENABLED
	notifications := tenant.Group("/notifications")
	notifications.Get("/", handlerRegistry.Notification.GetNotifications)
	notifications.Get("/recent", handlerRegistry.Notification.GetRecentNotifications)
	notifications.Get("/stats", handlerRegistry.Notification.GetNotificationStats)
	notifications.Post("/mark-as-read", handlerRegistry.Notification.MarkAsRead)
	notifications.Post("/mark-all-as-read", handlerRegistry.Notification.MarkAllAsRead)
	notifications.Delete("/:id", handlerRegistry.Notification.DeleteNotification)

	// Audit Logs (tenant-scoped) - ENABLED
	audit := tenant.Group("/audit-logs")
	audit.Get("/", middleware.RequirePermission(rbacService, "audit_log", "view"), handlers.GetAuditLogs)
	audit.Get("/document/:documentId", middleware.RequirePermission(rbacService, "audit_log", "view"), handlers.GetDocumentAuditLogs)

	// Note: Development tools and test workflow tasks are now created via seed data migrations
}
