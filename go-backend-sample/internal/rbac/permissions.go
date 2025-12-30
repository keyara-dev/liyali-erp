package rbac

// Permission represents a single permission
type Permission string

// Define all permissions in the system
const (
	// Document permissions
	PermissionCreateDocument   Permission = "document:create"
	PermissionViewDocument     Permission = "document:view"
	PermissionEditDocument     Permission = "document:edit"
	PermissionDeleteDocument   Permission = "document:delete"
	PermissionViewAllDocuments Permission = "document:view_all"

	// Approval permissions
	PermissionApproveStage1   Permission = "approval:stage1"
	PermissionApproveStage2   Permission = "approval:stage2"
	PermissionApproveStage3   Permission = "approval:stage3"
	PermissionViewApprovals   Permission = "approval:view"
	PermissionRejectDocument  Permission = "approval:reject"
	PermissionRequestRevision Permission = "approval:request_revision"

	// User management permissions
	PermissionCreateUser       Permission = "user:create"
	PermissionViewUser         Permission = "user:view"
	PermissionEditUser         Permission = "user:edit"
	PermissionDeleteUser       Permission = "user:delete"
	PermissionAssignRole       Permission = "user:assign_role"
	PermissionViewAllUsers     Permission = "user:view_all"
	PermissionActivateUser     Permission = "user:activate"
	PermissionDeactivateUser   Permission = "user:deactivate"

	// Workflow permissions
	PermissionCreateWorkflow Permission = "workflow:create"
	PermissionEditWorkflow   Permission = "workflow:edit"
	PermissionDeleteWorkflow Permission = "workflow:delete"
	PermissionViewWorkflow   Permission = "workflow:view"

	// Audit permissions
	PermissionViewAuditLogs Permission = "audit:view"
	PermissionExportAuditLogs Permission = "audit:export"

	// Settings permissions
	PermissionManageSettings Permission = "settings:manage"
	PermissionViewSettings   Permission = "settings:view"

	// Report permissions
	PermissionGenerateReports Permission = "reports:generate"
	PermissionViewReports     Permission = "reports:view"
	PermissionExportReports   Permission = "reports:export"

	// Department permissions
	PermissionManageDepartment Permission = "department:manage"
	PermissionViewDepartment   Permission = "department:view"

	// Notification permissions
	PermissionSendNotification Permission = "notification:send"
	PermissionViewNotifications Permission = "notification:view"
)

// Role represents a user role
type Role string

// Define all roles in the system
const (
	RoleAdmin             Role = "ADMIN"
	RoleCFO               Role = "CFO"
	RoleDirector          Role = "DIRECTOR"
	RoleFinanceOfficer    Role = "FINANCE_OFFICER"
	RoleDepartmentManager Role = "DEPARTMENT_MANAGER"
	RoleComplianceOfficer Role = "COMPLIANCE_OFFICER"
	RoleRequester         Role = "REQUESTER"
)

// rolePermissions maps roles to their permissions
var rolePermissions = map[Role][]Permission{
	RoleAdmin: {
		// All permissions
		PermissionCreateDocument,
		PermissionViewDocument,
		PermissionEditDocument,
		PermissionDeleteDocument,
		PermissionViewAllDocuments,
		PermissionApproveStage1,
		PermissionApproveStage2,
		PermissionApproveStage3,
		PermissionViewApprovals,
		PermissionRejectDocument,
		PermissionRequestRevision,
		PermissionCreateUser,
		PermissionViewUser,
		PermissionEditUser,
		PermissionDeleteUser,
		PermissionAssignRole,
		PermissionViewAllUsers,
		PermissionActivateUser,
		PermissionDeactivateUser,
		PermissionCreateWorkflow,
		PermissionEditWorkflow,
		PermissionDeleteWorkflow,
		PermissionViewWorkflow,
		PermissionViewAuditLogs,
		PermissionExportAuditLogs,
		PermissionManageSettings,
		PermissionViewSettings,
		PermissionGenerateReports,
		PermissionViewReports,
		PermissionExportReports,
		PermissionManageDepartment,
		PermissionViewDepartment,
		PermissionSendNotification,
		PermissionViewNotifications,
	},
	RoleCFO: {
		PermissionViewDocument,
		PermissionViewAllDocuments,
		PermissionApproveStage3,
		PermissionViewApprovals,
		PermissionRejectDocument,
		PermissionRequestRevision,
		PermissionViewUser,
		PermissionViewAllUsers,
		PermissionViewWorkflow,
		PermissionViewAuditLogs,
		PermissionViewSettings,
		PermissionGenerateReports,
		PermissionViewReports,
		PermissionExportReports,
		PermissionViewDepartment,
		PermissionViewNotifications,
	},
	RoleDirector: {
		PermissionViewDocument,
		PermissionViewAllDocuments,
		PermissionApproveStage3,
		PermissionViewApprovals,
		PermissionRejectDocument,
		PermissionRequestRevision,
		PermissionViewUser,
		PermissionViewAllUsers,
		PermissionViewWorkflow,
		PermissionViewAuditLogs,
		PermissionViewSettings,
		PermissionGenerateReports,
		PermissionViewReports,
		PermissionExportReports,
		PermissionViewDepartment,
		PermissionViewNotifications,
	},
	RoleFinanceOfficer: {
		PermissionCreateDocument,
		PermissionViewDocument,
		PermissionEditDocument,
		PermissionViewAllDocuments,
		PermissionApproveStage2,
		PermissionViewApprovals,
		PermissionRejectDocument,
		PermissionRequestRevision,
		PermissionViewUser,
		PermissionViewWorkflow,
		PermissionViewSettings,
		PermissionViewReports,
		PermissionViewDepartment,
		PermissionViewNotifications,
	},
	RoleDepartmentManager: {
		PermissionCreateDocument,
		PermissionViewDocument,
		PermissionEditDocument,
		PermissionApproveStage1,
		PermissionViewApprovals,
		PermissionRejectDocument,
		PermissionRequestRevision,
		PermissionViewUser,
		PermissionViewWorkflow,
		PermissionViewSettings,
		PermissionViewReports,
		PermissionManageDepartment,
		PermissionViewDepartment,
		PermissionViewNotifications,
	},
	RoleComplianceOfficer: {
		PermissionViewDocument,
		PermissionViewAllDocuments,
		PermissionViewApprovals,
		PermissionViewUser,
		PermissionViewAllUsers,
		PermissionViewWorkflow,
		PermissionViewAuditLogs,
		PermissionExportAuditLogs,
		PermissionViewSettings,
		PermissionViewReports,
		PermissionExportReports,
		PermissionViewDepartment,
		PermissionViewNotifications,
	},
	RoleRequester: {
		PermissionCreateDocument,
		PermissionViewDocument,
		PermissionViewApprovals,
		PermissionViewSettings,
		PermissionViewNotifications,
	},
}

// HasPermission checks if a role has a specific permission
func HasPermission(role Role, permission Permission) bool {
	permissions, exists := rolePermissions[role]
	if !exists {
		return false
	}

	for _, p := range permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// HasAnyPermission checks if a role has any of the specified permissions
func HasAnyPermission(role Role, permissions ...Permission) bool {
	for _, permission := range permissions {
		if HasPermission(role, permission) {
			return true
		}
	}
	return false
}

// HasAllPermissions checks if a role has all of the specified permissions
func HasAllPermissions(role Role, permissions ...Permission) bool {
	for _, permission := range permissions {
		if !HasPermission(role, permission) {
			return false
		}
	}
	return true
}

// GetRolePermissions returns all permissions for a role
func GetRolePermissions(role Role) []Permission {
	permissions, exists := rolePermissions[role]
	if !exists {
		return []Permission{}
	}
	return permissions
}

// IsValidRole checks if a role is valid
func IsValidRole(role string) bool {
	_, exists := rolePermissions[Role(role)]
	return exists
}
