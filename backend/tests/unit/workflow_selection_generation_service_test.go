package unit

import (
	"context"
	"testing"

	"github.com/liyali/liyali-gateway/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAssignWorkflowToDocumentWithID_ValidatesWorkflowID(t *testing.T) {
	executionService := &services.WorkflowExecutionService{}

	_, err := executionService.AssignWorkflowToDocumentWithID(
		context.Background(),
		"org-1",
		"doc-1",
		"requisition",
		"not-a-uuid",
		"user-1",
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid workflow ID format")
}

func TestDocumentGenerationService_RequestValidation(t *testing.T) {
	generationService := services.NewDocumentGenerationService(nil, &services.DocumentAutomationService{})

	t.Run("requires source ID", func(t *testing.T) {
		_, err := generationService.GenerateFromSource(context.Background(), "org-1", "", "REQUISITION", "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "source ID is required")
	})

	t.Run("rejects unsupported docType", func(t *testing.T) {
		_, err := generationService.GenerateFromSource(context.Background(), "org-1", "doc-1", "INVOICE", "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported docType")
	})

	t.Run("rejects target type mismatch", func(t *testing.T) {
		_, err := generationService.GenerateFromSource(context.Background(), "org-1", "doc-1", "REQUISITION", "GRN")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid targetDocType")
	})
}
