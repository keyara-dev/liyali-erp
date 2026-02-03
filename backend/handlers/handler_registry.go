package handlers

import (
	"github.com/liyali/liyali-gateway/logging"
	"github.com/liyali/liyali-gateway/services"
)

// HandlerRegistry holds all application handlers
type HandlerRegistry struct {
	Auth                    *AuthHandler
	Approval                *ApprovalHandler
	Workflow                *WorkflowHandler
	Document                *DocumentHandler
	Notification            *NotificationHandler
	Subscription            *SubscriptionHandler
	WorkflowExecutionService *services.WorkflowExecutionService
	// Add other handlers here as we migrate them
}

// NewHandlerRegistry creates a new handler registry with all handlers
func NewHandlerRegistry(
	authService *services.AuthService,
	rbacService *services.RBACService,
	workflowService *services.WorkflowService,
	workflowExecutionService *services.WorkflowExecutionService,
	documentService *services.DocumentService,
	automationService *services.DocumentAutomationService,
	subscriptionService *services.SubscriptionService,
	logger *logging.Logger,
) *HandlerRegistry {
	return &HandlerRegistry{
		Auth:                     NewAuthHandler(authService, rbacService),
		Approval:                 NewApprovalHandler(),
		Workflow:                 NewWorkflowHandler(workflowService),
		Document:                 NewDocumentHandler(documentService),
		Notification:             NewNotificationHandler(),
		Subscription:             NewSubscriptionHandler(subscriptionService, logger),
		WorkflowExecutionService: workflowExecutionService,
	}
}