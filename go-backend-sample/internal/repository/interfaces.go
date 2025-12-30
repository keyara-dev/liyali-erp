package repository

import (
	"context"
	"time"

	"github.com/cozyCodr/liyali-gateway/internal/db"
	"github.com/google/uuid"
)

// UserRepositoryInterface defines the interface for user repository operations
type UserRepositoryInterface interface {
	CreateUser(ctx context.Context, params db.CreateUserParams) (*db.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*db.User, error)
	GetUserByEmail(ctx context.Context, email string) (*db.User, error)
	UpdateUser(ctx context.Context, params db.UpdateUserParams) (*db.User, error)
	UpdateUserPassword(ctx context.Context, id uuid.UUID, passwordHash string) error
	UpdateUserLastLogin(ctx context.Context, id uuid.UUID) error
	IncrementFailedLoginAttempts(ctx context.Context, id uuid.UUID) error
	ResetFailedLoginAttempts(ctx context.Context, id uuid.UUID) error
	LockUserAccount(ctx context.Context, id uuid.UUID, lockedUntil time.Time) error
	DeactivateUser(ctx context.Context, id uuid.UUID) error
	ActivateUser(ctx context.Context, id uuid.UUID) error
	DeleteUser(ctx context.Context, id uuid.UUID) error
	ListUsers(ctx context.Context, limit, offset int32) ([]db.User, error)
	ListUsersByRole(ctx context.Context, role string) ([]db.User, error)
	ListUsersByDepartment(ctx context.Context, department string) ([]db.User, error)
	CountUsers(ctx context.Context) (int64, error)
	CountActiveUsers(ctx context.Context) (int64, error)
}

// SessionRepositoryInterface defines the interface for session repository operations
type SessionRepositoryInterface interface {
	CreateSession(ctx context.Context, userID uuid.UUID, refreshToken, ipAddress, userAgent string, expiresAt time.Time) (*db.Session, error)
	GetSessionByRefreshToken(ctx context.Context, refreshToken string) (*db.Session, error)
	GetSessionsByUserID(ctx context.Context, userID uuid.UUID) ([]db.Session, error)
	DeleteSession(ctx context.Context, id uuid.UUID) error
	DeleteSessionByRefreshToken(ctx context.Context, refreshToken string) error
	DeleteSessionsByUserID(ctx context.Context, userID uuid.UUID) error
	DeleteExpiredSessions(ctx context.Context) error
	CountActiveSessions(ctx context.Context) (int64, error)
	CountUserActiveSessions(ctx context.Context, userID uuid.UUID) (int64, error)
}

// PasswordResetRepositoryInterface defines the interface for password reset repository operations
type PasswordResetRepositoryInterface interface {
	CreatePasswordReset(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time) (*db.PasswordReset, error)
	GetPasswordResetByToken(ctx context.Context, token string) (*db.PasswordReset, error)
	MarkPasswordResetAsUsed(ctx context.Context, id uuid.UUID) error
	DeletePasswordResetsByUserID(ctx context.Context, userID uuid.UUID) error
	DeleteExpiredPasswordResets(ctx context.Context) error
	DeleteUsedPasswordResets(ctx context.Context) error
}

// WorkflowRepositoryInterface defines the interface for workflow repository operations
type WorkflowRepositoryInterface interface {
	CreateWorkflow(ctx context.Context, params db.CreateWorkflowParams) (*db.Workflow, error)
	GetWorkflowByID(ctx context.Context, id uuid.UUID) (*db.Workflow, error)
	ListWorkflows(ctx context.Context, limit, offset int32) ([]db.Workflow, error)
	ListActiveWorkflows(ctx context.Context, limit, offset int32) ([]db.Workflow, error)
	ListWorkflowsByDocumentType(ctx context.Context, documentType string, limit, offset int32) ([]db.Workflow, error)
	ListActiveWorkflowsByDocumentType(ctx context.Context, documentType string, limit, offset int32) ([]db.Workflow, error)
	GetDefaultWorkflowByDocumentType(ctx context.Context, documentType string) (*db.Workflow, error)
	UpdateWorkflow(ctx context.Context, params db.UpdateWorkflowParams) (*db.Workflow, error)
	ActivateWorkflow(ctx context.Context, id uuid.UUID) (*db.Workflow, error)
	DeactivateWorkflow(ctx context.Context, id uuid.UUID) (*db.Workflow, error)
	DeleteWorkflow(ctx context.Context, id uuid.UUID) error
	CountWorkflows(ctx context.Context) (int64, error)
	CountActiveWorkflows(ctx context.Context) (int64, error)
	CountWorkflowsByDocumentType(ctx context.Context, documentType string) (int64, error)
}

// DocumentRepositoryInterface defines the interface for document repository operations
type DocumentRepositoryInterface interface {
	CreateDocument(ctx context.Context, params db.CreateDocumentParams) (*db.Document, error)
	GetDocumentByID(ctx context.Context, id uuid.UUID) (*db.Document, error)
	GetDocumentByNumber(ctx context.Context, documentNumber string) (*db.Document, error)
	ListDocuments(ctx context.Context, limit, offset int32) ([]db.Document, error)
	ListDocumentsByType(ctx context.Context, documentType string, limit, offset int32) ([]db.Document, error)
	ListDocumentsByStatus(ctx context.Context, status string, limit, offset int32) ([]db.Document, error)
	ListDocumentsByCreator(ctx context.Context, creatorID uuid.UUID, limit, offset int32) ([]db.Document, error)
	ListDocumentsByDepartment(ctx context.Context, department string, limit, offset int32) ([]db.Document, error)
	ListDocumentsByTypeAndStatus(ctx context.Context, documentType, status string, limit, offset int32) ([]db.Document, error)
	ListDocumentsByWorkflow(ctx context.Context, workflowID uuid.UUID, limit, offset int32) ([]db.Document, error)
	UpdateDocument(ctx context.Context, params db.UpdateDocumentParams) (*db.Document, error)
	UpdateDocumentStatus(ctx context.Context, params db.UpdateDocumentStatusParams) (*db.Document, error)
	SubmitDocument(ctx context.Context, id uuid.UUID) (*db.Document, error)
	ApproveDocument(ctx context.Context, id uuid.UUID) (*db.Document, error)
	RejectDocument(ctx context.Context, id uuid.UUID) (*db.Document, error)
	DeleteDocument(ctx context.Context, id uuid.UUID) error
	CountDocuments(ctx context.Context) (int64, error)
	CountDocumentsByType(ctx context.Context, documentType string) (int64, error)
	CountDocumentsByStatus(ctx context.Context, status string) (int64, error)
	CountDocumentsByCreator(ctx context.Context, creatorID uuid.UUID) (int64, error)
}

// ApprovalTaskRepositoryInterface defines the interface for approval task repository operations
type ApprovalTaskRepositoryInterface interface {
	CreateApprovalTask(ctx context.Context, params db.CreateApprovalTaskParams) (*db.ApprovalTask, error)
	GetApprovalTaskByID(ctx context.Context, id uuid.UUID) (*db.ApprovalTask, error)
	ListApprovalTasksByAssignee(ctx context.Context, assigneeID uuid.UUID, limit, offset int32) ([]db.ApprovalTask, error)
	ListApprovalTasksByStatus(ctx context.Context, status string, limit, offset int32) ([]db.ApprovalTask, error)
	ListApprovalTasksByAssigneeAndStatus(ctx context.Context, assigneeID uuid.UUID, status string, limit, offset int32) ([]db.ApprovalTask, error)
	ListApprovalTasksByDocument(ctx context.Context, documentID uuid.UUID) ([]db.ApprovalTask, error)
	ListPendingApprovalTasks(ctx context.Context, limit, offset int32) ([]db.ApprovalTask, error)
	ListOverdueApprovalTasks(ctx context.Context, limit, offset int32) ([]db.ApprovalTask, error)
	UpdateApprovalTaskStatus(ctx context.Context, params db.UpdateApprovalTaskStatusParams) (*db.ApprovalTask, error)
	UpdateApprovalTaskStage(ctx context.Context, params db.UpdateApprovalTaskStageParams) (*db.ApprovalTask, error)
	ReassignApprovalTask(ctx context.Context, params db.ReassignApprovalTaskParams) (*db.ApprovalTask, error)
	UpdateApprovalTaskNotes(ctx context.Context, params db.UpdateApprovalTaskNotesParams) (*db.ApprovalTask, error)
	DeleteApprovalTask(ctx context.Context, id uuid.UUID) error
	CountApprovalTasksByAssignee(ctx context.Context, assigneeID uuid.UUID) (int64, error)
	CountApprovalTasksByStatus(ctx context.Context, status string) (int64, error)
	CountPendingApprovalTasksByAssignee(ctx context.Context, assigneeID uuid.UUID) (int64, error)
}

// ApprovalHistoryRepositoryInterface defines the interface for approval history repository operations
type ApprovalHistoryRepositoryInterface interface {
	CreateApprovalHistoryEntry(ctx context.Context, params db.CreateApprovalHistoryEntryParams) (*db.ApprovalHistory, error)
	GetApprovalHistoryByID(ctx context.Context, id uuid.UUID) (*db.ApprovalHistory, error)
	ListApprovalHistoryByTask(ctx context.Context, taskID uuid.UUID) ([]db.ApprovalHistory, error)
	ListApprovalHistoryByUser(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]db.ApprovalHistory, error)
	ListApprovalHistoryByAction(ctx context.Context, action string, limit, offset int32) ([]db.ApprovalHistory, error)
	GetLatestApprovalHistoryByTask(ctx context.Context, taskID uuid.UUID) (*db.ApprovalHistory, error)
	DeleteApprovalHistory(ctx context.Context, id uuid.UUID) error
	CountApprovalHistoryByTask(ctx context.Context, taskID uuid.UUID) (int64, error)
	CountApprovalHistoryByUser(ctx context.Context, userID uuid.UUID) (int64, error)
}

// AuditLogRepositoryInterface defines the interface for audit log repository operations
type AuditLogRepositoryInterface interface {
	CreateAuditLog(ctx context.Context, params db.CreateAuditLogParams) (*db.AuditLog, error)
	GetAuditLogByID(ctx context.Context, id uuid.UUID) (*db.AuditLog, error)
	ListAuditLogs(ctx context.Context, limit, offset int32) ([]db.AuditLog, error)
	ListAuditLogsByUser(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]db.AuditLog, error)
	ListAuditLogsByResource(ctx context.Context, resourceType string, resourceID uuid.UUID, limit, offset int32) ([]db.AuditLog, error)
	ListAuditLogsByResourceType(ctx context.Context, resourceType string, limit, offset int32) ([]db.AuditLog, error)
	ListAuditLogsByAction(ctx context.Context, action string, limit, offset int32) ([]db.AuditLog, error)
	DeleteOldAuditLogs(ctx context.Context, before time.Time) error
	CountAuditLogs(ctx context.Context) (int64, error)
	CountAuditLogsByUser(ctx context.Context, userID uuid.UUID) (int64, error)
	CountAuditLogsByResource(ctx context.Context, resourceType string, resourceID uuid.UUID) (int64, error)
}

// NotificationRepositoryInterface defines the interface for notification repository operations
type NotificationRepositoryInterface interface {
	CreateNotification(ctx context.Context, params db.CreateNotificationParams) (*db.Notification, error)
	GetNotificationByID(ctx context.Context, id uuid.UUID) (*db.Notification, error)
	ListNotificationsByUser(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]db.Notification, error)
	ListUnreadNotificationsByUser(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]db.Notification, error)
	ListNotificationsByType(ctx context.Context, notificationType string, limit, offset int32) ([]db.Notification, error)
	ListNotificationsByRelatedID(ctx context.Context, relatedID uuid.UUID) ([]db.Notification, error)
	MarkNotificationAsRead(ctx context.Context, id uuid.UUID) (*db.Notification, error)
	MarkAllNotificationsAsRead(ctx context.Context, userID uuid.UUID) error
	DeleteNotification(ctx context.Context, id uuid.UUID) error
	DeleteOldNotifications(ctx context.Context, before time.Time) error
	CountUnreadNotificationsByUser(ctx context.Context, userID uuid.UUID) (int64, error)
	CountNotificationsByUser(ctx context.Context, userID uuid.UUID) (int64, error)
	CountNotificationsByType(ctx context.Context, notificationType string) (int64, error)
}
