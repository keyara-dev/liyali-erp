package unit

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/datatypes"
)

// Mock Document Repository for testing
type MockDocumentRepository struct {
	mock.Mock
}

func (m *MockDocumentRepository) Create(ctx context.Context, document *models.Document) (*models.Document, error) {
	args := m.Called(ctx, document)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Document), args.Error(1)
}

func (m *MockDocumentRepository) GetByID(ctx context.Context, id uuid.UUID, organizationID string) (*models.Document, error) {
	args := m.Called(ctx, id, organizationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Document), args.Error(1)
}

func (m *MockDocumentRepository) GetByNumber(ctx context.Context, documentNumber, organizationID string) (*models.Document, error) {
	args := m.Called(ctx, documentNumber, organizationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Document), args.Error(1)
}

func (m *MockDocumentRepository) Update(ctx context.Context, document *models.Document) (*models.Document, error) {
	args := m.Called(ctx, document)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Document), args.Error(1)
}

func (m *MockDocumentRepository) Delete(ctx context.Context, id uuid.UUID, organizationID string) error {
	args := m.Called(ctx, id, organizationID)
	return args.Error(0)
}

func (m *MockDocumentRepository) List(ctx context.Context, organizationID string, filter *models.DocumentFilter, limit, offset int) ([]*models.Document, error) {
	args := m.Called(ctx, organizationID, filter, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Document), args.Error(1)
}

func (m *MockDocumentRepository) Count(ctx context.Context, organizationID string, filter *models.DocumentFilter) (int64, error) {
	args := m.Called(ctx, organizationID, filter)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockDocumentRepository) ListByUser(ctx context.Context, organizationID, userID string, limit, offset int) ([]*models.Document, error) {
	args := m.Called(ctx, organizationID, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Document), args.Error(1)
}

func (m *MockDocumentRepository) CountByUser(ctx context.Context, organizationID, userID string) (int64, error) {
	args := m.Called(ctx, organizationID, userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockDocumentRepository) Search(ctx context.Context, organizationID, query string, filter *models.DocumentFilter, limit, offset int) ([]*models.DocumentSearchResult, int64, error) {
	args := m.Called(ctx, organizationID, query, filter, limit, offset)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*models.DocumentSearchResult), args.Get(1).(int64), args.Error(2)
}

func (m *MockDocumentRepository) Submit(ctx context.Context, id uuid.UUID, organizationID string) error {
	args := m.Called(ctx, id, organizationID)
	return args.Error(0)
}

func (m *MockDocumentRepository) GetStats(ctx context.Context, organizationID string) (*models.DocumentStats, error) {
	args := m.Called(ctx, organizationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.DocumentStats), args.Error(1)
}

// Sync methods (simplified for testing)
func (m *MockDocumentRepository) SyncFromRequisition(ctx context.Context, req *models.Requisition) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockDocumentRepository) SyncFromBudget(ctx context.Context, budget *models.Budget) error {
	args := m.Called(ctx, budget)
	return args.Error(0)
}

func (m *MockDocumentRepository) SyncFromPurchaseOrder(ctx context.Context, po *models.PurchaseOrder) error {
	args := m.Called(ctx, po)
	return args.Error(0)
}

func (m *MockDocumentRepository) SyncFromPaymentVoucher(ctx context.Context, pv *models.PaymentVoucher) error {
	args := m.Called(ctx, pv)
	return args.Error(0)
}

func (m *MockDocumentRepository) SyncFromGRN(ctx context.Context, grn *models.GoodsReceivedNote) error {
	args := m.Called(ctx, grn)
	return args.Error(0)
}

func TestDocumentService_CreateDocument(t *testing.T) {
	tests := []struct {
		name          string
		request       services.CreateDocumentRequest
		setupMocks    func(*MockDocumentRepository)
		expectedError string
		expectSuccess bool
	}{
		{
			name: "Valid requisition document creation",
			request: services.CreateDocumentRequest{
				DocumentType: "REQUISITION",
				Title:        "Test Requisition",
				Description:  "Test Description",
				Amount:       1000.0,
				Currency:     "USD",
				Department:   "IT",
				Data: map[string]interface{}{
					"items": []map[string]interface{}{
						{
							"description": "Laptop",
							"quantity":    1,
							"unitPrice":   1000.0,
						},
					},
				},
				Metadata: map[string]interface{}{
					"priority": "high",
				},
			},
			setupMocks: func(repo *MockDocumentRepository) {
				expectedDoc := &models.Document{
					ID:             uuid.New(),
					DocumentType:   "REQUISITION",
					Title:          "Test Requisition",
					Status:         "draft",
					OrganizationID: "org-123",
					CreatedBy:      "user-123",
				}
				repo.On("Create", mock.Anything, mock.MatchedBy(func(doc *models.Document) bool {
					return doc.DocumentType == "REQUISITION" && doc.Title == "Test Requisition"
				})).Return(expectedDoc, nil)
			},
			expectSuccess: true,
		},
		{
			name: "Invalid document type",
			request: services.CreateDocumentRequest{
				DocumentType: "INVALID_TYPE",
				Title:        "Test Document",
				Data:         map[string]interface{}{"test": "data"},
			},
			setupMocks:    func(repo *MockDocumentRepository) {},
			expectedError: "invalid document type",
			expectSuccess: false,
		},
		{
			name: "Repository error",
			request: services.CreateDocumentRequest{
				DocumentType: "REQUISITION",
				Title:        "Test Requisition",
				Data:         map[string]interface{}{"test": "data"},
			},
			setupMocks: func(repo *MockDocumentRepository) {
				repo.On("Create", mock.Anything, mock.Anything).Return(nil, assert.AnError)
			},
			expectedError: "failed to create document",
			expectSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			docRepo := &MockDocumentRepository{}
			auditService := &services.AuditService{}
			tt.setupMocks(docRepo)

			// Create service
			docService := services.NewDocumentService(docRepo, auditService)

			// Execute
			result, err := docService.CreateDocument(
				context.Background(),
				"org-123",
				"user-123",
				tt.request,
			)

			// Verify
			if tt.expectSuccess {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.request.DocumentType, result.DocumentType)
				assert.Equal(t, tt.request.Title, result.Title)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			}

			docRepo.AssertExpectations(t)
		})
	}
}

func TestDocumentService_UpdateDocument(t *testing.T) {
	docID := uuid.New()
	
	tests := []struct {
		name          string
		documentID    uuid.UUID
		request       services.UpdateDocumentRequest
		setupMocks    func(*MockDocumentRepository)
		expectedError string
		expectSuccess bool
	}{
		{
			name:       "Valid document update",
			documentID: docID,
			request: services.UpdateDocumentRequest{
				Title:       "Updated Title",
				Description: "Updated Description",
				Amount:      2000.0,
				Data: map[string]interface{}{
					"updated": true,
				},
			},
			setupMocks: func(repo *MockDocumentRepository) {
				existingDoc := &models.Document{
					ID:           docID,
					DocumentType: "REQUISITION",
					Title:        "Original Title",
					Status:       "draft", // Editable status
					CreatedBy:    "user-123",
				}
				
				updatedDoc := &models.Document{
					ID:           docID,
					DocumentType: "REQUISITION",
					Title:        "Updated Title",
					Status:       "draft",
					CreatedBy:    "user-123",
				}

				repo.On("GetByID", mock.Anything, docID, "org-123").Return(existingDoc, nil)
				repo.On("Update", mock.Anything, mock.MatchedBy(func(doc *models.Document) bool {
					return doc.Title == "Updated Title"
				})).Return(updatedDoc, nil)
			},
			expectSuccess: true,
		},
		{
			name:       "Document not found",
			documentID: docID,
			request: services.UpdateDocumentRequest{
				Title: "Updated Title",
			},
			setupMocks: func(repo *MockDocumentRepository) {
				repo.On("GetByID", mock.Anything, docID, "org-123").Return(nil, assert.AnError)
			},
			expectedError: "document not found",
			expectSuccess: false,
		},
		{
			name:       "Document not editable",
			documentID: docID,
			request: services.UpdateDocumentRequest{
				Title: "Updated Title",
			},
			setupMocks: func(repo *MockDocumentRepository) {
				existingDoc := &models.Document{
					ID:           docID,
					DocumentType: "REQUISITION",
					Title:        "Original Title",
					Status:       "approved", // Not editable
					CreatedBy:    "user-123",
				}

				repo.On("GetByID", mock.Anything, docID, "org-123").Return(existingDoc, nil)
			},
			expectedError: "cannot be edited",
			expectSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			docRepo := &MockDocumentRepository{}
			auditService := &services.AuditService{}
			tt.setupMocks(docRepo)

			// Create service
			docService := services.NewDocumentService(docRepo, auditService)

			// Execute
			result, err := docService.UpdateDocument(
				context.Background(),
				tt.documentID,
				"org-123",
				"user-123",
				tt.request,
			)

			// Verify
			if tt.expectSuccess {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			}

			docRepo.AssertExpectations(t)
		})
	}
}

func TestDocumentService_DeleteDocument(t *testing.T) {
	docID := uuid.New()

	tests := []struct {
		name          string
		setupMocks    func(*MockDocumentRepository)
		expectedError string
		expectSuccess bool
	}{
		{
			name: "Valid document deletion",
			setupMocks: func(repo *MockDocumentRepository) {
				existingDoc := &models.Document{
					ID:           docID,
					DocumentType: "REQUISITION",
					Title:        "Test Document",
					Status:       "draft", // Deletable status
					CreatedBy:    "user-123",
				}

				repo.On("GetByID", mock.Anything, docID, "org-123").Return(existingDoc, nil)
				repo.On("Delete", mock.Anything, docID, "org-123").Return(nil)
			},
			expectSuccess: true,
		},
		{
			name: "Document not deletable",
			setupMocks: func(repo *MockDocumentRepository) {
				existingDoc := &models.Document{
					ID:           docID,
					DocumentType: "REQUISITION",
					Title:        "Test Document",
					Status:       "approved", // Not deletable
					CreatedBy:    "user-123",
				}

				repo.On("GetByID", mock.Anything, docID, "org-123").Return(existingDoc, nil)
			},
			expectedError: "cannot be deleted",
			expectSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			docRepo := &MockDocumentRepository{}
			auditService := &services.AuditService{}
			tt.setupMocks(docRepo)

			// Create service
			docService := services.NewDocumentService(docRepo, auditService)

			// Execute
			err := docService.DeleteDocument(
				context.Background(),
				docID,
				"org-123",
				"user-123",
			)

			// Verify
			if tt.expectSuccess {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			}

			docRepo.AssertExpectations(t)
		})
	}
}

func TestDocumentService_ListDocuments(t *testing.T) {
	tests := []struct {
		name          string
		filter        *models.DocumentFilter
		setupMocks    func(*MockDocumentRepository)
		expectedCount int
		expectSuccess bool
	}{
		{
			name: "List all documents",
			filter: &models.DocumentFilter{
				DocumentType: "",
				Status:       "",
			},
			setupMocks: func(repo *MockDocumentRepository) {
				documents := []*models.Document{
					{
						ID:           uuid.New(),
						DocumentType: "REQUISITION",
						Title:        "Req 1",
						Status:       "draft",
					},
					{
						ID:           uuid.New(),
						DocumentType: "BUDGET",
						Title:        "Budget 1",
						Status:       "approved",
					},
				}

				repo.On("List", mock.Anything, "org-123", mock.Anything, 20, 0).Return(documents, nil)
				repo.On("Count", mock.Anything, "org-123", mock.Anything).Return(int64(2), nil)
			},
			expectedCount: 2,
			expectSuccess: true,
		},
		{
			name: "Filter by document type",
			filter: &models.DocumentFilter{
				DocumentType: "REQUISITION",
				Status:       "",
			},
			setupMocks: func(repo *MockDocumentRepository) {
				documents := []*models.Document{
					{
						ID:           uuid.New(),
						DocumentType: "REQUISITION",
						Title:        "Req 1",
						Status:       "draft",
					},
				}

				repo.On("List", mock.Anything, "org-123", mock.MatchedBy(func(filter *models.DocumentFilter) bool {
					return filter.DocumentType == "REQUISITION"
				}), 20, 0).Return(documents, nil)
				repo.On("Count", mock.Anything, "org-123", mock.Anything).Return(int64(1), nil)
			},
			expectedCount: 1,
			expectSuccess: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			docRepo := &MockDocumentRepository{}
			auditService := &services.AuditService{}
			tt.setupMocks(docRepo)

			// Create service
			docService := services.NewDocumentService(docRepo, auditService)

			// Execute
			documents, total, err := docService.ListDocuments(
				context.Background(),
				"org-123",
				tt.filter,
				20,
				0,
			)

			// Verify
			if tt.expectSuccess {
				assert.NoError(t, err)
				assert.Len(t, documents, tt.expectedCount)
				assert.Equal(t, int64(tt.expectedCount), total)
			} else {
				assert.Error(t, err)
			}

			docRepo.AssertExpectations(t)
		})
	}
}

func TestDocumentService_SubmitDocument(t *testing.T) {
	docID := uuid.New()

	tests := []struct {
		name          string
		setupMocks    func(*MockDocumentRepository)
		expectedError string
		expectSuccess bool
	}{
		{
			name: "Valid document submission",
			setupMocks: func(repo *MockDocumentRepository) {
				draftDoc := &models.Document{
					ID:           docID,
					DocumentType: "REQUISITION",
					Title:        "Test Document",
					Status:       "draft", // Can be submitted
					CreatedBy:    "user-123",
				}

				submittedDoc := &models.Document{
					ID:           docID,
					DocumentType: "REQUISITION",
					Title:        "Test Document",
					Status:       "submitted",
					CreatedBy:    "user-123",
				}

				repo.On("GetByID", mock.Anything, docID, "org-123").Return(draftDoc, nil).Once()
				repo.On("Submit", mock.Anything, docID, "org-123").Return(nil)
				repo.On("GetByID", mock.Anything, docID, "org-123").Return(submittedDoc, nil).Once()
			},
			expectSuccess: true,
		},
		{
			name: "Document cannot be submitted",
			setupMocks: func(repo *MockDocumentRepository) {
				approvedDoc := &models.Document{
					ID:           docID,
					DocumentType: "REQUISITION",
					Title:        "Test Document",
					Status:       "approved", // Cannot be submitted again
					CreatedBy:    "user-123",
				}

				repo.On("GetByID", mock.Anything, docID, "org-123").Return(approvedDoc, nil)
			},
			expectedError: "cannot be submitted",
			expectSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			docRepo := &MockDocumentRepository{}
			auditService := &services.AuditService{}
			tt.setupMocks(docRepo)

			// Create service
			docService := services.NewDocumentService(docRepo, auditService)

			// Execute
			result, err := docService.SubmitDocument(
				context.Background(),
				docID,
				"org-123",
				"user-123",
			)

			// Verify
			if tt.expectSuccess {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, "submitted", result.Status)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			}

			docRepo.AssertExpectations(t)
		})
	}
}

func TestDocumentService_SyncFromSpecificModel(t *testing.T) {
	tests := []struct {
		name          string
		modelType     string
		model         interface{}
		setupMocks    func(*MockDocumentRepository)
		expectedError string
		expectSuccess bool
	}{
		{
			name:      "Sync from requisition",
			modelType: "REQUISITION",
			model: &models.Requisition{
				ID:    "req-123",
				Title: "Test Requisition",
			},
			setupMocks: func(repo *MockDocumentRepository) {
				repo.On("SyncFromRequisition", mock.Anything, mock.AnythingOfType("*models.Requisition")).Return(nil)
			},
			expectSuccess: true,
		},
		{
			name:      "Sync from budget",
			modelType: "BUDGET",
			model: &models.Budget{
				ID:   "budget-123",
				Name: "Test Budget",
			},
			setupMocks: func(repo *MockDocumentRepository) {
				repo.On("SyncFromBudget", mock.Anything, mock.AnythingOfType("*models.Budget")).Return(nil)
			},
			expectSuccess: true,
		},
		{
			name:          "Unsupported model type",
			modelType:     "UNSUPPORTED",
			model:         &struct{}{},
			setupMocks:    func(repo *MockDocumentRepository) {},
			expectedError: "unsupported model type",
			expectSuccess: false,
		},
		{
			name:      "Wrong model type for sync",
			modelType: "REQUISITION",
			model:     &models.Budget{}, // Wrong type
			setupMocks: func(repo *MockDocumentRepository) {
				// No mock setup as it should fail before reaching repo
			},
			expectedError: "unsupported model type",
			expectSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			docRepo := &MockDocumentRepository{}
			auditService := &services.AuditService{}
			tt.setupMocks(docRepo)

			// Create service
			docService := services.NewDocumentService(docRepo, auditService)

			// Execute
			err := docService.SyncFromSpecificModel(
				context.Background(),
				tt.modelType,
				tt.model,
			)

			// Verify
			if tt.expectSuccess {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			}

			docRepo.AssertExpectations(t)
		})
	}
}

func TestDocumentService_DataIntegrity(t *testing.T) {
	t.Run("JSON data marshaling and unmarshaling", func(t *testing.T) {
		docRepo := &MockDocumentRepository{}
		auditService := &services.AuditService{}

		// Test data with complex nested structures
		testData := map[string]interface{}{
			"items": []map[string]interface{}{
				{
					"description": "Test Item",
					"quantity":    10,
					"unitPrice":   100.50,
					"metadata": map[string]interface{}{
						"category": "electronics",
						"tags":     []string{"laptop", "business"},
					},
				},
			},
			"approvals": []map[string]interface{}{
				{
					"stage":      1,
					"approver":   "manager@example.com",
					"timestamp":  "2023-01-01T00:00:00Z",
					"comments":   "Approved for budget allocation",
				},
			},
		}

		expectedDoc := &models.Document{
			ID:             uuid.New(),
			DocumentType:   "REQUISITION",
			Title:          "Complex Data Test",
			Status:         "draft",
			OrganizationID: "org-123",
			CreatedBy:      "user-123",
		}

		// Mock repository to capture the document with marshaled data
		docRepo.On("Create", mock.Anything, mock.MatchedBy(func(doc *models.Document) bool {
			// Verify that data was properly marshaled
			var unmarshaledData map[string]interface{}
			err := json.Unmarshal(doc.Data, &unmarshaledData)
			if err != nil {
				return false
			}

			// Check if the data structure is preserved
			items, ok := unmarshaledData["items"].([]interface{})
			if !ok || len(items) != 1 {
				return false
			}

			firstItem, ok := items[0].(map[string]interface{})
			if !ok {
				return false
			}

			return firstItem["description"] == "Test Item" && firstItem["quantity"] == float64(10)
		})).Return(expectedDoc, nil)

		docService := services.NewDocumentService(docRepo, auditService)

		request := services.CreateDocumentRequest{
			DocumentType: "REQUISITION",
			Title:        "Complex Data Test",
			Data:         testData,
		}

		result, err := docService.CreateDocument(
			context.Background(),
			"org-123",
			"user-123",
			request,
		)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		docRepo.AssertExpectations(t)
	})

	t.Run("Invalid JSON data handling", func(t *testing.T) {
		docRepo := &MockDocumentRepository{}
		auditService := &services.AuditService{}
		docService := services.NewDocumentService(docRepo, auditService)

		// Test with data that cannot be marshaled to JSON
		invalidData := map[string]interface{}{
			"invalid": make(chan int), // Channels cannot be marshaled to JSON
		}

		request := services.CreateDocumentRequest{
			DocumentType: "REQUISITION",
			Title:        "Invalid Data Test",
			Data:         invalidData,
		}

		result, err := docService.CreateDocument(
			context.Background(),
			"org-123",
			"user-123",
			request,
		)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to marshal document data")
		assert.Nil(t, result)
	})
}