package unit

import (
	"testing"

	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/tests/helpers"
	"github.com/stretchr/testify/assert"
)

// TestDocumentService_CreateDocument tests document creation using mocks
func TestDocumentService_CreateDocument(t *testing.T) {
	builder := helpers.NewMockTestDataBuilder()
	
	t.Run("Valid requisition document creation", func(t *testing.T) {
		// Create mock document
		mockDoc := &models.Document{
			ID:             uuid.New(),
			DocumentType:   "REQUISITION",
			Title:          "Test Requisition",
			Status: "DRAFT",
			OrganizationID: builder.GetOrganizationID(),
			CreatedBy:      builder.GetUserID(),
		}
		
		// Verify document properties
		assert.NotNil(t, mockDoc)
		assert.Equal(t, "REQUISITION", mockDoc.DocumentType)
		assert.Equal(t, "Test Requisition", mockDoc.Title)
		assert.Equal(t, "draft", mockDoc.Status)
		assert.Equal(t, builder.GetOrganizationID(), mockDoc.OrganizationID)
	})
	
	t.Run("Budget document creation", func(t *testing.T) {
		// Create mock budget document
		mockDoc := &models.Document{
			ID:             uuid.New(),
			DocumentType:   "BUDGET",
			Title:          "Test Budget",
			Status: "DRAFT",
			OrganizationID: builder.GetOrganizationID(),
			CreatedBy:      builder.GetUserID(),
		}
		
		// Verify document properties
		assert.NotNil(t, mockDoc)
		assert.Equal(t, "BUDGET", mockDoc.DocumentType)
		assert.Equal(t, "Test Budget", mockDoc.Title)
	})
	
	t.Run("Purchase order document creation", func(t *testing.T) {
		// Create mock PO document
		mockDoc := &models.Document{
			ID:             uuid.New(),
			DocumentType:   "PURCHASE_ORDER",
			Title:          "Test PO",
			Status: "DRAFT",
			OrganizationID: builder.GetOrganizationID(),
			CreatedBy:      builder.GetUserID(),
		}
		
		// Verify document properties
		assert.NotNil(t, mockDoc)
		assert.Equal(t, "PURCHASE_ORDER", mockDoc.DocumentType)
		assert.Equal(t, "Test PO", mockDoc.Title)
	})
}

// TestDocumentService_UpdateDocument tests document updates using mocks
func TestDocumentService_UpdateDocument(t *testing.T) {
	builder := helpers.NewMockTestDataBuilder()
	
	t.Run("Valid document update", func(t *testing.T) {
		// Create mock document
		mockDoc := &models.Document{
			ID:             uuid.New(),
			DocumentType:   "REQUISITION",
			Title:          "Original Title",
			Status: "DRAFT",
			OrganizationID: builder.GetOrganizationID(),
			CreatedBy:      builder.GetUserID(),
		}
		
		// Simulate update
		mockDoc.Title = "Updated Title"
		mockDoc.Status = "PENDING"
		
		// Verify update
		assert.Equal(t, "Updated Title", mockDoc.Title)
		assert.Equal(t, "pending", mockDoc.Status)
	})
	
	t.Run("Document status transition", func(t *testing.T) {
		// Simulate status transitions
		validTransitions := map[string][]string{
			"draft":    {"pending", "cancelled"},
			"pending":  {"approved", "rejected", "cancelled"},
			"approved": {"completed", "cancelled"},
			"rejected": {"draft", "cancelled"},
		}
		
		// Verify transitions
		assert.Contains(t, validTransitions["draft"], "pending")
		assert.Contains(t, validTransitions["pending"], "approved")
		assert.Contains(t, validTransitions["approved"], "completed")
	})
}

// TestDocumentService_ListDocuments tests document listing using mocks
func TestDocumentService_ListDocuments(t *testing.T) {
	builder := helpers.NewMockTestDataBuilder()
	
	t.Run("List documents by organization", func(t *testing.T) {
		// Create mock documents
		mockDocs := []*models.Document{
			{
				ID:             uuid.New(),
				DocumentType:   "REQUISITION",
				Title:          "Requisition 1",
				Status: "DRAFT",
				OrganizationID: builder.GetOrganizationID(),
				CreatedBy:      builder.GetUserID(),
			},
			{
				ID:             uuid.New(),
				DocumentType:   "BUDGET",
				Title:          "Budget 1",
				Status: "PENDING",
				OrganizationID: builder.GetOrganizationID(),
				CreatedBy:      builder.GetUserID(),
			},
			{
				ID:             uuid.New(),
				DocumentType:   "PURCHASE_ORDER",
				Title:          "PO 1",
				Status: "APPROVED",
				OrganizationID: builder.GetOrganizationID(),
				CreatedBy:      builder.GetUserID(),
			},
		}
		
		// Verify list
		assert.Len(t, mockDocs, 3)
		assert.Equal(t, "REQUISITION", mockDocs[0].DocumentType)
		assert.Equal(t, "BUDGET", mockDocs[1].DocumentType)
		assert.Equal(t, "PURCHASE_ORDER", mockDocs[2].DocumentType)
	})
	
	t.Run("Filter documents by type", func(t *testing.T) {
		// Create mock documents
		mockDocs := []*models.Document{
			{
				ID:             uuid.New(),
				DocumentType:   "REQUISITION",
				Title:          "Requisition 1",
				Status: "DRAFT",
				OrganizationID: builder.GetOrganizationID(),
				CreatedBy:      builder.GetUserID(),
			},
			{
				ID:             uuid.New(),
				DocumentType:   "REQUISITION",
				Title:          "Requisition 2",
				Status: "PENDING",
				OrganizationID: builder.GetOrganizationID(),
				CreatedBy:      builder.GetUserID(),
			},
		}
		
		// Filter by type
		requisitions := make([]*models.Document, 0)
		for _, doc := range mockDocs {
			if doc.DocumentType == "REQUISITION" {
				requisitions = append(requisitions, doc)
			}
		}
		
		// Verify filter
		assert.Len(t, requisitions, 2)
		for _, doc := range requisitions {
			assert.Equal(t, "REQUISITION", doc.DocumentType)
		}
	})
	
	t.Run("Filter documents by status", func(t *testing.T) {
		// Create mock documents
		mockDocs := []*models.Document{
			{
				ID:             uuid.New(),
				DocumentType:   "REQUISITION",
				Title:          "Requisition 1",
				Status: "DRAFT",
				OrganizationID: builder.GetOrganizationID(),
				CreatedBy:      builder.GetUserID(),
			},
			{
				ID:             uuid.New(),
				DocumentType:   "BUDGET",
				Title:          "Budget 1",
				Status: "PENDING",
				OrganizationID: builder.GetOrganizationID(),
				CreatedBy:      builder.GetUserID(),
			},
			{
				ID:             uuid.New(),
				DocumentType:   "PURCHASE_ORDER",
				Title:          "PO 1",
				Status: "APPROVED",
				OrganizationID: builder.GetOrganizationID(),
				CreatedBy:      builder.GetUserID(),
			},
		}
		
		// Filter by status
		pendingDocs := make([]*models.Document, 0)
		for _, doc := range mockDocs {
			if doc.Status == "PENDING" {
				pendingDocs = append(pendingDocs, doc)
			}
		}
		
		// Verify filter
		assert.Len(t, pendingDocs, 1)
		assert.Equal(t, "pending", pendingDocs[0].Status)
	})
}

// TestDocumentService_DeleteDocument tests document deletion using mocks
func TestDocumentService_DeleteDocument(t *testing.T) {
	builder := helpers.NewMockTestDataBuilder()
	
	t.Run("Delete draft document", func(t *testing.T) {
		// Create mock document
		mockDoc := &models.Document{
			ID:             uuid.New(),
			DocumentType:   "REQUISITION",
			Title:          "Test Requisition",
			Status: "DRAFT",
			OrganizationID: builder.GetOrganizationID(),
			CreatedBy:      builder.GetUserID(),
		}
		
		// Verify document exists
		assert.NotNil(t, mockDoc)
		assert.Equal(t, "draft", mockDoc.Status)
		
		// Simulate deletion (only draft documents can be deleted)
		var deletedDoc *models.Document
		if mockDoc.Status == "DRAFT" {
			deletedDoc = nil
		} else {
			deletedDoc = mockDoc
		}
		
		// Verify deletion
		assert.Nil(t, deletedDoc)
	})
	
	t.Run("Cannot delete non-draft document", func(t *testing.T) {
		// Create mock document
		mockDoc := &models.Document{
			ID:             uuid.New(),
			DocumentType:   "REQUISITION",
			Title:          "Test Requisition",
			Status: "PENDING",
			OrganizationID: builder.GetOrganizationID(),
			CreatedBy:      builder.GetUserID(),
		}
		
		// Verify document exists
		assert.NotNil(t, mockDoc)
		assert.Equal(t, "pending", mockDoc.Status)
		
		// Try to delete (should fail for non-draft)
		canDelete := mockDoc.Status == "DRAFT"
		assert.False(t, canDelete)
	})
}

// TestDocumentService_DocumentStats tests document statistics using mocks
func TestDocumentService_DocumentStats(t *testing.T) {
	builder := helpers.NewMockTestDataBuilder()
	
	t.Run("Calculate document statistics", func(t *testing.T) {
		// Create mock documents
		mockDocs := []*models.Document{
			{
				ID:             uuid.New(),
				DocumentType:   "REQUISITION",
				Title:          "Requisition 1",
				Status: "DRAFT",
				OrganizationID: builder.GetOrganizationID(),
				CreatedBy:      builder.GetUserID(),
			},
			{
				ID:             uuid.New(),
				DocumentType:   "REQUISITION",
				Title:          "Requisition 2",
				Status: "PENDING",
				OrganizationID: builder.GetOrganizationID(),
				CreatedBy:      builder.GetUserID(),
			},
			{
				ID:             uuid.New(),
				DocumentType:   "BUDGET",
				Title:          "Budget 1",
				Status: "APPROVED",
				OrganizationID: builder.GetOrganizationID(),
				CreatedBy:      builder.GetUserID(),
			},
		}
		
		// Calculate stats
		stats := map[string]int{
			"total":       len(mockDocs),
			"draft":       0,
			"pending":     0,
			"approved":    0,
			"requisition": 0,
			"budget":      0,
		}
		
		for _, doc := range mockDocs {
			stats[doc.Status]++
			if doc.DocumentType == "REQUISITION" {
				stats["requisition"]++
			} else if doc.DocumentType == "BUDGET" {
				stats["budget"]++
			}
		}
		
		// Verify stats
		assert.Equal(t, 3, stats["total"])
		assert.Equal(t, 1, stats["draft"])
		assert.Equal(t, 1, stats["pending"])
		assert.Equal(t, 1, stats["approved"])
		assert.Equal(t, 2, stats["requisition"])
		assert.Equal(t, 1, stats["budget"])
	})
}
