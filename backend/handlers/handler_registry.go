package handlers

import (
	"github.com/liyali/liyali-gateway/services"
)

// HandlerRegistry holds all application handlers
type HandlerRegistry struct {
	Auth     *AuthHandler
	Approval *ApprovalHandler
	Workflow *WorkflowHandler
	Document *DocumentHandler
	// Add other handlers here as we migrate them
}

// NewHandlerRegistry creates a new handler registry with all handlers
func NewHandlerRegistry(
	authService *services.AuthService,
	rbacService *services.RBACService,
	workflowService *services.WorkflowService,
	documentService *services.DocumentService,
) *HandlerRegistry {
	return &HandlerRegistry{
		Auth:     NewAuthHandler(authService, rbacService),
		Approval: NewApprovalHandler(),
		Workflow: NewWorkflowHandler(workflowService),
		Document: NewDocumentHandler(documentService),
	}
}