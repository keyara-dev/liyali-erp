package handlers

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/repository"
	"github.com/liyali/liyali-gateway/services"
	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// Mock repository
// ---------------------------------------------------------------------------

// mockDocumentRepo implements repository.DocumentRepositoryInterface for tests.
// Each field is a hook the test can set to control behaviour.
type mockDocumentRepo struct {
	createFn                      func(ctx context.Context, doc *models.Document) (*models.Document, error)
	getByIDFn                     func(ctx context.Context, id uuid.UUID, orgID string) (*models.Document, error)
	getByNumberFn                 func(ctx context.Context, number, orgID string) (*models.Document, error)
	getByNumberOnlyFn             func(ctx context.Context, number string) (*models.Document, error)
	updateFn                      func(ctx context.Context, doc *models.Document) (*models.Document, error)
	deleteFn                      func(ctx context.Context, id uuid.UUID, orgID string) error
	listFn                        func(ctx context.Context, orgID string, filter *models.DocumentFilter, limit, offset int) ([]*models.Document, error)
	listByUserFn                  func(ctx context.Context, orgID, userID string, limit, offset int) ([]*models.Document, error)
	listByTypeFn                  func(ctx context.Context, orgID, docType string, limit, offset int) ([]*models.Document, error)
	listByStatusFn                func(ctx context.Context, orgID, status string, limit, offset int) ([]*models.Document, error)
	listByDepartmentFn            func(ctx context.Context, orgID, dept string, limit, offset int) ([]*models.Document, error)
	searchFn                      func(ctx context.Context, orgID, query string, filter *models.DocumentFilter, limit, offset int) ([]*models.DocumentSearchResult, error)
	countFn                       func(ctx context.Context, orgID string, filter *models.DocumentFilter) (int64, error)
	countSearchFn                 func(ctx context.Context, orgID, query string, filter *models.DocumentFilter) (int64, error)
	countByTypeFn                 func(ctx context.Context, orgID, docType string) (int64, error)
	countByStatusFn               func(ctx context.Context, orgID, status string) (int64, error)
	countByUserFn                 func(ctx context.Context, orgID, userID string) (int64, error)
	updateStatusFn                func(ctx context.Context, id uuid.UUID, orgID, status string) error
	submitFn                      func(ctx context.Context, id uuid.UUID, orgID string) error
	getStatsFn                    func(ctx context.Context, orgID string) (*models.DocumentStats, error)
	getRequisitionByNumberFn      func(ctx context.Context, number string) (*models.Requisition, error)
	getPurchaseOrderByNumberFn    func(ctx context.Context, number string) (*models.PurchaseOrder, error)
	getPaymentVoucherByNumberFn   func(ctx context.Context, number string) (*models.PaymentVoucher, error)
	getGRNByNumberFn              func(ctx context.Context, number string) (*models.GoodsReceivedNote, error)
	getOrganizationBrandingFn     func(ctx context.Context, orgID string) (*models.Organization, error)
}

var _ repository.DocumentRepositoryInterface = (*mockDocumentRepo)(nil)

func (m *mockDocumentRepo) Create(ctx context.Context, doc *models.Document) (*models.Document, error) {
	if m.createFn != nil {
		return m.createFn(ctx, doc)
	}
	doc.ID = uuid.New()
	doc.DocumentNumber = "DOC-TEST-001"
	return doc, nil
}

func (m *mockDocumentRepo) GetByID(ctx context.Context, id uuid.UUID, orgID string) (*models.Document, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id, orgID)
	}
	return nil, errors.New("not found")
}

func (m *mockDocumentRepo) GetByNumber(ctx context.Context, number, orgID string) (*models.Document, error) {
	if m.getByNumberFn != nil {
		return m.getByNumberFn(ctx, number, orgID)
	}
	return nil, errors.New("not found")
}

func (m *mockDocumentRepo) GetByNumberOnly(ctx context.Context, number string) (*models.Document, error) {
	if m.getByNumberOnlyFn != nil {
		return m.getByNumberOnlyFn(ctx, number)
	}
	return nil, errors.New("not found")
}

func (m *mockDocumentRepo) Update(ctx context.Context, doc *models.Document) (*models.Document, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, doc)
	}
	return doc, nil
}

func (m *mockDocumentRepo) Delete(ctx context.Context, id uuid.UUID, orgID string) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id, orgID)
	}
	return nil
}

func (m *mockDocumentRepo) List(ctx context.Context, orgID string, filter *models.DocumentFilter, limit, offset int) ([]*models.Document, error) {
	if m.listFn != nil {
		return m.listFn(ctx, orgID, filter, limit, offset)
	}
	return []*models.Document{}, nil
}

func (m *mockDocumentRepo) ListByUser(ctx context.Context, orgID, userID string, limit, offset int) ([]*models.Document, error) {
	if m.listByUserFn != nil {
		return m.listByUserFn(ctx, orgID, userID, limit, offset)
	}
	return []*models.Document{}, nil
}

func (m *mockDocumentRepo) ListByType(ctx context.Context, orgID, docType string, limit, offset int) ([]*models.Document, error) {
	if m.listByTypeFn != nil {
		return m.listByTypeFn(ctx, orgID, docType, limit, offset)
	}
	return []*models.Document{}, nil
}

func (m *mockDocumentRepo) ListByStatus(ctx context.Context, orgID, status string, limit, offset int) ([]*models.Document, error) {
	if m.listByStatusFn != nil {
		return m.listByStatusFn(ctx, orgID, status, limit, offset)
	}
	return []*models.Document{}, nil
}

func (m *mockDocumentRepo) ListByDepartment(ctx context.Context, orgID, dept string, limit, offset int) ([]*models.Document, error) {
	if m.listByDepartmentFn != nil {
		return m.listByDepartmentFn(ctx, orgID, dept, limit, offset)
	}
	return []*models.Document{}, nil
}

func (m *mockDocumentRepo) Search(ctx context.Context, orgID, query string, filter *models.DocumentFilter, limit, offset int) ([]*models.DocumentSearchResult, error) {
	if m.searchFn != nil {
		return m.searchFn(ctx, orgID, query, filter, limit, offset)
	}
	return []*models.DocumentSearchResult{}, nil
}

func (m *mockDocumentRepo) Count(ctx context.Context, orgID string, filter *models.DocumentFilter) (int64, error) {
	if m.countFn != nil {
		return m.countFn(ctx, orgID, filter)
	}
	return 0, nil
}

func (m *mockDocumentRepo) CountSearch(ctx context.Context, orgID, query string, filter *models.DocumentFilter) (int64, error) {
	if m.countSearchFn != nil {
		return m.countSearchFn(ctx, orgID, query, filter)
	}
	return 0, nil
}

func (m *mockDocumentRepo) CountByType(ctx context.Context, orgID, docType string) (int64, error) {
	if m.countByTypeFn != nil {
		return m.countByTypeFn(ctx, orgID, docType)
	}
	return 0, nil
}

func (m *mockDocumentRepo) CountByStatus(ctx context.Context, orgID, status string) (int64, error) {
	if m.countByStatusFn != nil {
		return m.countByStatusFn(ctx, orgID, status)
	}
	return 0, nil
}

func (m *mockDocumentRepo) CountByUser(ctx context.Context, orgID, userID string) (int64, error) {
	if m.countByUserFn != nil {
		return m.countByUserFn(ctx, orgID, userID)
	}
	return 0, nil
}

func (m *mockDocumentRepo) UpdateStatus(ctx context.Context, id uuid.UUID, orgID, status string) error {
	if m.updateStatusFn != nil {
		return m.updateStatusFn(ctx, id, orgID, status)
	}
	return nil
}

func (m *mockDocumentRepo) Submit(ctx context.Context, id uuid.UUID, orgID string) error {
	if m.submitFn != nil {
		return m.submitFn(ctx, id, orgID)
	}
	return nil
}

func (m *mockDocumentRepo) GetStats(ctx context.Context, orgID string) (*models.DocumentStats, error) {
	if m.getStatsFn != nil {
		return m.getStatsFn(ctx, orgID)
	}
	return &models.DocumentStats{
		TotalDocuments:    0,
		DocumentsByType:   map[string]int64{},
		DocumentsByStatus: map[string]int64{},
		DocumentsByDept:   map[string]int64{},
	}, nil
}

func (m *mockDocumentRepo) GetRequisitionByNumberPublic(ctx context.Context, number string) (*models.Requisition, error) {
	if m.getRequisitionByNumberFn != nil {
		return m.getRequisitionByNumberFn(ctx, number)
	}
	return nil, errors.New("not found")
}

func (m *mockDocumentRepo) GetPurchaseOrderByNumberPublic(ctx context.Context, number string) (*models.PurchaseOrder, error) {
	if m.getPurchaseOrderByNumberFn != nil {
		return m.getPurchaseOrderByNumberFn(ctx, number)
	}
	return nil, errors.New("not found")
}

func (m *mockDocumentRepo) GetPaymentVoucherByNumberPublic(ctx context.Context, number string) (*models.PaymentVoucher, error) {
	if m.getPaymentVoucherByNumberFn != nil {
		return m.getPaymentVoucherByNumberFn(ctx, number)
	}
	return nil, errors.New("not found")
}

func (m *mockDocumentRepo) GetGRNByNumberPublic(ctx context.Context, number string) (*models.GoodsReceivedNote, error) {
	if m.getGRNByNumberFn != nil {
		return m.getGRNByNumberFn(ctx, number)
	}
	return nil, errors.New("not found")
}

func (m *mockDocumentRepo) GetOrganizationBranding(ctx context.Context, orgID string) (*models.Organization, error) {
	if m.getOrganizationBrandingFn != nil {
		return m.getOrganizationBrandingFn(ctx, orgID)
	}
	return nil, errors.New("not found")
}

// ---------------------------------------------------------------------------
// App factory helpers
// ---------------------------------------------------------------------------

// newDocumentApp builds a Fiber app wired to DocumentHandler.
// All routes require the tenant middleware.
func newDocumentApp(repo repository.DocumentRepositoryInterface) *fiber.App {
	auditSvc := services.NewAuditService()
	docSvc := services.NewDocumentService(repo, auditSvc, nil)
	h := NewDocumentHandler(docSvc)

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"error":   err.Error(),
			})
		},
	})

	auth := withTenantCtx(testOrgID, testUserID, testUserRole)

	app.Get("/documents", auth, h.GetDocuments)
	app.Get("/documents/my", auth, h.GetMyDocuments)
	app.Get("/documents/search", auth, h.SearchDocuments)
	app.Get("/documents/stats", auth, h.GetDocumentStats)
	app.Get("/documents/number/:number", auth, h.GetDocumentByNumber)
	app.Get("/documents/:id", auth, h.GetDocumentByID)
	app.Post("/documents", auth, h.CreateDocument)
	app.Put("/documents/:id", auth, h.UpdateDocument)
	app.Post("/documents/:id/submit", auth, h.SubmitDocument)
	app.Delete("/documents/:id", auth, h.DeleteDocument)
	app.Get("/public/verify/:documentNumber", h.VerifyDocumentPublic)

	return app
}

// newDocumentAppNoAuth builds a Fiber app WITHOUT tenant middleware — used to
// verify that handlers return a non-200 response when locals are absent.
// recover.New() converts panics (from type assertions on nil locals) to 500s.
func newDocumentAppNoAuth(repo repository.DocumentRepositoryInterface) *fiber.App {
	auditSvc := services.NewAuditService()
	docSvc := services.NewDocumentService(repo, auditSvc, nil)
	h := NewDocumentHandler(docSvc)

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"error":   err.Error(),
			})
		},
	})
	app.Use(recover.New())

	app.Get("/documents", h.GetDocuments)
	app.Get("/documents/my", h.GetMyDocuments)
	app.Get("/documents/search", h.SearchDocuments)
	app.Get("/documents/stats", h.GetDocumentStats)
	app.Get("/documents/number/:number", h.GetDocumentByNumber)
	app.Get("/documents/:id", h.GetDocumentByID)
	app.Post("/documents", h.CreateDocument)
	app.Put("/documents/:id", h.UpdateDocument)
	app.Post("/documents/:id/submit", h.SubmitDocument)
	app.Delete("/documents/:id", h.DeleteDocument)

	return app
}

// sampleDocument returns a minimal valid Document for seeding mocks.
func sampleDocument() *models.Document {
	id := uuid.New()
	now := time.Now()
	status := "DRAFT"
	currency := "USD"
	return &models.Document{
		ID:             id,
		OrganizationID: testOrgID,
		DocumentType:   "REQUISITION",
		DocumentNumber: "DOC-TEST-001",
		Title:          "Test Document",
		Status:         status,
		Currency:       &currency,
		CreatedBy:      testUserID,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

// validCreateDocumentPayload returns a map with all required fields for CreateDocument.
func validCreateDocumentPayload() map[string]interface{} {
	return map[string]interface{}{
		"documentType": "REQUISITION",
		"title":        "My Test Requisition",
		"data": map[string]interface{}{
			"requestedBy": "John Doe",
		},
	}
}

// ---------------------------------------------------------------------------
// GET /documents — GetDocuments
// ---------------------------------------------------------------------------

func TestGetDocuments_NoAuth(t *testing.T) {
	repo := &mockDocumentRepo{}
	app := newDocumentAppNoAuth(repo)

	resp := testRequest(app, http.MethodGet, "/documents", nil)
	// Without tenant locals the type assertion panics → recovered as 500.
	assert.NotEqual(t, http.StatusOK, resp.StatusCode)
}

func TestGetDocuments_Empty(t *testing.T) {
	repo := &mockDocumentRepo{}
	app := newDocumentApp(repo)

	resp := testRequest(app, http.MethodGet, "/documents", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeResponse(resp)
	assert.Equal(t, true, body["success"])
	data, ok := body["data"].([]interface{})
	assert.True(t, ok, "data should be a JSON array")
	assert.Len(t, data, 0)
}

func TestGetDocuments_WithResults(t *testing.T) {
	doc := sampleDocument()
	repo := &mockDocumentRepo{
		listFn: func(_ context.Context, orgID string, _ *models.DocumentFilter, _, _ int) ([]*models.Document, error) {
			assert.Equal(t, testOrgID, orgID)
			return []*models.Document{doc}, nil
		},
		countFn: func(_ context.Context, _ string, _ *models.DocumentFilter) (int64, error) {
			return 1, nil
		},
	}

	app := newDocumentApp(repo)

	resp := testRequest(app, http.MethodGet, "/documents", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeResponse(resp)
	assert.Equal(t, true, body["success"])
	data, ok := body["data"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, data, 1)
	first := data[0].(map[string]interface{})
	assert.Equal(t, "Test Document", first["title"])
}

func TestGetDocuments_RepoError(t *testing.T) {
	repo := &mockDocumentRepo{
		listFn: func(_ context.Context, _ string, _ *models.DocumentFilter, _, _ int) ([]*models.Document, error) {
			return nil, errors.New("db error")
		},
	}
	app := newDocumentApp(repo)

	resp := testRequest(app, http.MethodGet, "/documents", nil)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

func TestGetDocuments_PaginationParams(t *testing.T) {
	repo := &mockDocumentRepo{}
	app := newDocumentApp(repo)

	resp := testRequest(app, http.MethodGet, "/documents?page=2&limit=5", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeResponse(resp)
	assert.Equal(t, true, body["success"])
	// Pagination meta should be present.
	assert.NotNil(t, body["pagination"])
}

func TestGetDocuments_FilterByType(t *testing.T) {
	called := false
	repo := &mockDocumentRepo{
		listFn: func(_ context.Context, _ string, filter *models.DocumentFilter, _, _ int) ([]*models.Document, error) {
			called = true
			assert.Contains(t, filter.DocumentTypes, "BUDGET")
			return []*models.Document{}, nil
		},
	}
	app := newDocumentApp(repo)

	resp := testRequest(app, http.MethodGet, "/documents?documentTypes=BUDGET", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.True(t, called)
}

func TestGetDocuments_FilterByStatus(t *testing.T) {
	called := false
	repo := &mockDocumentRepo{
		listFn: func(_ context.Context, _ string, filter *models.DocumentFilter, _, _ int) ([]*models.Document, error) {
			called = true
			assert.Contains(t, filter.Statuses, "APPROVED")
			return []*models.Document{}, nil
		},
	}
	app := newDocumentApp(repo)

	resp := testRequest(app, http.MethodGet, "/documents?statuses=APPROVED", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.True(t, called)
}

// ---------------------------------------------------------------------------
// GET /documents/my — GetMyDocuments
// ---------------------------------------------------------------------------

func TestGetMyDocuments_NoAuth(t *testing.T) {
	repo := &mockDocumentRepo{}
	app := newDocumentAppNoAuth(repo)

	resp := testRequest(app, http.MethodGet, "/documents/my", nil)
	assert.NotEqual(t, http.StatusOK, resp.StatusCode)
}

func TestGetMyDocuments_Empty(t *testing.T) {
	repo := &mockDocumentRepo{}
	app := newDocumentApp(repo)

	resp := testRequest(app, http.MethodGet, "/documents/my", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeResponse(resp)
	assert.Equal(t, true, body["success"])
	data, ok := body["data"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, data, 0)
}

func TestGetMyDocuments_WithResults(t *testing.T) {
	doc := sampleDocument()
	repo := &mockDocumentRepo{
		listByUserFn: func(_ context.Context, orgID, userID string, _, _ int) ([]*models.Document, error) {
			assert.Equal(t, testOrgID, orgID)
			assert.Equal(t, testUserID, userID)
			return []*models.Document{doc}, nil
		},
		countByUserFn: func(_ context.Context, _, _ string) (int64, error) {
			return 1, nil
		},
	}
	app := newDocumentApp(repo)

	resp := testRequest(app, http.MethodGet, "/documents/my", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeResponse(resp)
	assert.Equal(t, true, body["success"])
	data := body["data"].([]interface{})
	assert.Len(t, data, 1)
}

func TestGetMyDocuments_RepoError(t *testing.T) {
	repo := &mockDocumentRepo{
		listByUserFn: func(_ context.Context, _, _ string, _, _ int) ([]*models.Document, error) {
			return nil, errors.New("db error")
		},
	}
	app := newDocumentApp(repo)

	resp := testRequest(app, http.MethodGet, "/documents/my", nil)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// GET /documents/:id — GetDocumentByID
// ---------------------------------------------------------------------------

func TestGetDocumentByID_NoAuth(t *testing.T) {
	repo := &mockDocumentRepo{}
	app := newDocumentAppNoAuth(repo)

	resp := testRequest(app, http.MethodGet, "/documents/"+uuid.New().String(), nil)
	assert.NotEqual(t, http.StatusOK, resp.StatusCode)
}

func TestGetDocumentByID_InvalidUUID(t *testing.T) {
	repo := &mockDocumentRepo{}
	app := newDocumentApp(repo)

	resp := testRequest(app, http.MethodGet, "/documents/not-a-uuid", nil)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	body := decodeResponse(resp)
	assert.Equal(t, false, body["success"])
}

func TestGetDocumentByID_NotFound(t *testing.T) {
	repo := &mockDocumentRepo{
		getByIDFn: func(_ context.Context, _ uuid.UUID, _ string) (*models.Document, error) {
			return nil, errors.New("not found")
		},
	}
	app := newDocumentApp(repo)

	resp := testRequest(app, http.MethodGet, "/documents/"+uuid.New().String(), nil)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	body := decodeResponse(resp)
	assert.Equal(t, false, body["success"])
}

func TestGetDocumentByID_Success(t *testing.T) {
	doc := sampleDocument()
	repo := &mockDocumentRepo{
		getByIDFn: func(_ context.Context, id uuid.UUID, orgID string) (*models.Document, error) {
			assert.Equal(t, doc.ID, id)
			assert.Equal(t, testOrgID, orgID)
			return doc, nil
		},
	}
	app := newDocumentApp(repo)

	resp := testRequest(app, http.MethodGet, "/documents/"+doc.ID.String(), nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeResponse(resp)
	assert.Equal(t, true, body["success"])
	data := body["data"].(map[string]interface{})
	assert.Equal(t, doc.ID.String(), data["id"])
	assert.Equal(t, "Test Document", data["title"])
}

// ---------------------------------------------------------------------------
// GET /documents/number/:number — GetDocumentByNumber
// ---------------------------------------------------------------------------

func TestGetDocumentByNumber_NoAuth(t *testing.T) {
	repo := &mockDocumentRepo{}
	app := newDocumentAppNoAuth(repo)

	resp := testRequest(app, http.MethodGet, "/documents/number/DOC-001", nil)
	assert.NotEqual(t, http.StatusOK, resp.StatusCode)
}

func TestGetDocumentByNumber_NotFound(t *testing.T) {
	repo := &mockDocumentRepo{
		getByNumberFn: func(_ context.Context, _ string, _ string) (*models.Document, error) {
			return nil, errors.New("not found")
		},
	}
	app := newDocumentApp(repo)

	resp := testRequest(app, http.MethodGet, "/documents/number/NONEXISTENT-001", nil)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestGetDocumentByNumber_Success(t *testing.T) {
	doc := sampleDocument()
	repo := &mockDocumentRepo{
		getByNumberFn: func(_ context.Context, number, orgID string) (*models.Document, error) {
			assert.Equal(t, "DOC-TEST-001", number)
			assert.Equal(t, testOrgID, orgID)
			return doc, nil
		},
	}
	app := newDocumentApp(repo)

	resp := testRequest(app, http.MethodGet, "/documents/number/DOC-TEST-001", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeResponse(resp)
	assert.Equal(t, true, body["success"])
	data := body["data"].(map[string]interface{})
	assert.Equal(t, "DOC-TEST-001", data["documentNumber"])
}

// ---------------------------------------------------------------------------
// POST /documents — CreateDocument
// ---------------------------------------------------------------------------

func TestCreateDocument_NoAuth(t *testing.T) {
	repo := &mockDocumentRepo{}
	app := newDocumentAppNoAuth(repo)

	resp := testRequest(app, http.MethodPost, "/documents", validCreateDocumentPayload())
	assert.NotEqual(t, http.StatusCreated, resp.StatusCode)
}

func TestCreateDocument_MissingDocumentType(t *testing.T) {
	repo := &mockDocumentRepo{}
	app := newDocumentApp(repo)

	payload := validCreateDocumentPayload()
	delete(payload, "documentType")

	resp := testRequest(app, http.MethodPost, "/documents", payload)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	body := decodeResponse(resp)
	assert.Equal(t, false, body["success"])
}

func TestCreateDocument_MissingTitle(t *testing.T) {
	repo := &mockDocumentRepo{}
	app := newDocumentApp(repo)

	payload := validCreateDocumentPayload()
	delete(payload, "title")

	resp := testRequest(app, http.MethodPost, "/documents", payload)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	body := decodeResponse(resp)
	assert.Equal(t, false, body["success"])
}

func TestCreateDocument_MissingData(t *testing.T) {
	repo := &mockDocumentRepo{}
	app := newDocumentApp(repo)

	payload := validCreateDocumentPayload()
	delete(payload, "data")

	resp := testRequest(app, http.MethodPost, "/documents", payload)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	body := decodeResponse(resp)
	assert.Equal(t, false, body["success"])
}

func TestCreateDocument_InvalidDocumentType(t *testing.T) {
	// The service rejects unknown document types with an error
	// which the handler converts to 500 (internal error path).
	repo := &mockDocumentRepo{}
	app := newDocumentApp(repo)

	payload := validCreateDocumentPayload()
	payload["documentType"] = "UNKNOWN_TYPE"

	resp := testRequest(app, http.MethodPost, "/documents", payload)
	// service returns error → handler returns 500
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

func TestCreateDocument_Success(t *testing.T) {
	doc := sampleDocument()
	repo := &mockDocumentRepo{
		createFn: func(_ context.Context, d *models.Document) (*models.Document, error) {
			d.ID = doc.ID
			d.DocumentNumber = doc.DocumentNumber
			return d, nil
		},
	}
	app := newDocumentApp(repo)

	resp := testRequest(app, http.MethodPost, "/documents", validCreateDocumentPayload())
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	body := decodeResponse(resp)
	assert.Equal(t, true, body["success"])
	data := body["data"].(map[string]interface{})
	assert.Equal(t, "My Test Requisition", data["title"])
	assert.Equal(t, "REQUISITION", data["documentType"])
	assert.NotEmpty(t, data["id"])
}

func TestCreateDocument_RepoError(t *testing.T) {
	repo := &mockDocumentRepo{
		createFn: func(_ context.Context, _ *models.Document) (*models.Document, error) {
			return nil, errors.New("insert failed")
		},
	}
	app := newDocumentApp(repo)

	resp := testRequest(app, http.MethodPost, "/documents", validCreateDocumentPayload())
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

func TestCreateDocument_WithOptionalFields(t *testing.T) {
	doc := sampleDocument()
	repo := &mockDocumentRepo{
		createFn: func(_ context.Context, d *models.Document) (*models.Document, error) {
			d.ID = doc.ID
			d.DocumentNumber = doc.DocumentNumber
			return d, nil
		},
	}
	app := newDocumentApp(repo)

	payload := validCreateDocumentPayload()
	payload["description"] = "A detailed description"
	payload["amount"] = 5000.00
	payload["currency"] = "ZMW"
	payload["department"] = "Finance"

	resp := testRequest(app, http.MethodPost, "/documents", payload)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	body := decodeResponse(resp)
	assert.Equal(t, true, body["success"])
}

// ---------------------------------------------------------------------------
// PUT /documents/:id — UpdateDocument
// ---------------------------------------------------------------------------

func TestUpdateDocument_NoAuth(t *testing.T) {
	repo := &mockDocumentRepo{}
	app := newDocumentAppNoAuth(repo)

	resp := testRequest(app, http.MethodPut, "/documents/"+uuid.New().String(), map[string]interface{}{"title": "New Title"})
	assert.NotEqual(t, http.StatusOK, resp.StatusCode)
}

func TestUpdateDocument_InvalidUUID(t *testing.T) {
	repo := &mockDocumentRepo{}
	app := newDocumentApp(repo)

	resp := testRequest(app, http.MethodPut, "/documents/not-a-uuid", map[string]interface{}{"title": "New Title"})
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestUpdateDocument_NotFound(t *testing.T) {
	repo := &mockDocumentRepo{
		getByIDFn: func(_ context.Context, _ uuid.UUID, _ string) (*models.Document, error) {
			return nil, errors.New("not found")
		},
	}
	app := newDocumentApp(repo)

	resp := testRequest(app, http.MethodPut, "/documents/"+uuid.New().String(), map[string]interface{}{"title": "New Title"})
	// Service returns error → handler sends 500 (UpdateDocument uses SendInternalError)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

func TestUpdateDocument_NotEditable(t *testing.T) {
	doc := sampleDocument()
	doc.Status = "SUBMITTED" // not editable

	repo := &mockDocumentRepo{
		getByIDFn: func(_ context.Context, _ uuid.UUID, _ string) (*models.Document, error) {
			return doc, nil
		},
	}
	app := newDocumentApp(repo)

	resp := testRequest(app, http.MethodPut, "/documents/"+doc.ID.String(), map[string]interface{}{"title": "New Title"})
	// Service rejects non-editable document → 500
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

func TestUpdateDocument_Success(t *testing.T) {
	doc := sampleDocument()
	updated := *doc
	updated.Title = "Updated Title"

	repo := &mockDocumentRepo{
		getByIDFn: func(_ context.Context, _ uuid.UUID, _ string) (*models.Document, error) {
			return doc, nil
		},
		updateFn: func(_ context.Context, d *models.Document) (*models.Document, error) {
			return d, nil
		},
	}
	app := newDocumentApp(repo)

	resp := testRequest(app, http.MethodPut, "/documents/"+doc.ID.String(), map[string]interface{}{
		"title": "Updated Title",
	})
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeResponse(resp)
	assert.Equal(t, true, body["success"])
	data := body["data"].(map[string]interface{})
	assert.Equal(t, "Updated Title", data["title"])
}

// ---------------------------------------------------------------------------
// POST /documents/:id/submit — SubmitDocument
// ---------------------------------------------------------------------------

func TestSubmitDocument_NoAuth(t *testing.T) {
	repo := &mockDocumentRepo{}
	app := newDocumentAppNoAuth(repo)

	resp := testRequest(app, http.MethodPost, "/documents/"+uuid.New().String()+"/submit", nil)
	assert.NotEqual(t, http.StatusOK, resp.StatusCode)
}

func TestSubmitDocument_InvalidUUID(t *testing.T) {
	repo := &mockDocumentRepo{}
	app := newDocumentApp(repo)

	resp := testRequest(app, http.MethodPost, "/documents/not-a-uuid/submit", nil)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestSubmitDocument_NotFound(t *testing.T) {
	repo := &mockDocumentRepo{
		getByIDFn: func(_ context.Context, _ uuid.UUID, _ string) (*models.Document, error) {
			return nil, errors.New("not found")
		},
	}
	app := newDocumentApp(repo)

	resp := testRequest(app, http.MethodPost, "/documents/"+uuid.New().String()+"/submit", nil)
	// SubmitDocument calls GetByID then checks CanBeSubmitted; not found → 500
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

func TestSubmitDocument_AlreadySubmitted(t *testing.T) {
	doc := sampleDocument()
	doc.Status = "SUBMITTED" // CanBeSubmitted returns false

	repo := &mockDocumentRepo{
		getByIDFn: func(_ context.Context, _ uuid.UUID, _ string) (*models.Document, error) {
			return doc, nil
		},
	}
	app := newDocumentApp(repo)

	resp := testRequest(app, http.MethodPost, "/documents/"+doc.ID.String()+"/submit", nil)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

func TestSubmitDocument_Success(t *testing.T) {
	doc := sampleDocument() // status is DRAFT → CanBeSubmitted = true
	submittedDoc := *doc
	submittedDoc.Status = "SUBMITTED"

	callCount := 0
	repo := &mockDocumentRepo{
		getByIDFn: func(_ context.Context, _ uuid.UUID, _ string) (*models.Document, error) {
			callCount++
			if callCount == 1 {
				return doc, nil // first call: check CanBeSubmitted
			}
			return &submittedDoc, nil // second call: return updated doc
		},
		submitFn: func(_ context.Context, _ uuid.UUID, _ string) error {
			return nil
		},
	}
	app := newDocumentApp(repo)

	resp := testRequest(app, http.MethodPost, "/documents/"+doc.ID.String()+"/submit", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeResponse(resp)
	assert.Equal(t, true, body["success"])
}

// ---------------------------------------------------------------------------
// DELETE /documents/:id — DeleteDocument
// ---------------------------------------------------------------------------

func TestDeleteDocument_NoAuth(t *testing.T) {
	repo := &mockDocumentRepo{}
	app := newDocumentAppNoAuth(repo)

	resp := testRequest(app, http.MethodDelete, "/documents/"+uuid.New().String(), nil)
	assert.NotEqual(t, http.StatusOK, resp.StatusCode)
}

func TestDeleteDocument_InvalidUUID(t *testing.T) {
	repo := &mockDocumentRepo{}
	app := newDocumentApp(repo)

	resp := testRequest(app, http.MethodDelete, "/documents/not-a-uuid", nil)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestDeleteDocument_NotFound(t *testing.T) {
	repo := &mockDocumentRepo{
		getByIDFn: func(_ context.Context, _ uuid.UUID, _ string) (*models.Document, error) {
			return nil, errors.New("not found")
		},
	}
	app := newDocumentApp(repo)

	resp := testRequest(app, http.MethodDelete, "/documents/"+uuid.New().String(), nil)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

func TestDeleteDocument_NotDeletable(t *testing.T) {
	doc := sampleDocument()
	doc.Status = "APPROVED" // IsEditable returns false → cannot delete

	repo := &mockDocumentRepo{
		getByIDFn: func(_ context.Context, _ uuid.UUID, _ string) (*models.Document, error) {
			return doc, nil
		},
	}
	app := newDocumentApp(repo)

	resp := testRequest(app, http.MethodDelete, "/documents/"+doc.ID.String(), nil)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

func TestDeleteDocument_Success(t *testing.T) {
	doc := sampleDocument() // DRAFT status → deletable

	repo := &mockDocumentRepo{
		getByIDFn: func(_ context.Context, _ uuid.UUID, _ string) (*models.Document, error) {
			return doc, nil
		},
		deleteFn: func(_ context.Context, _ uuid.UUID, _ string) error {
			return nil
		},
	}
	app := newDocumentApp(repo)

	resp := testRequest(app, http.MethodDelete, "/documents/"+doc.ID.String(), nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeResponse(resp)
	assert.Equal(t, true, body["success"])
}

// ---------------------------------------------------------------------------
// GET /documents/search — SearchDocuments
// ---------------------------------------------------------------------------

func TestSearchDocuments_NoAuth(t *testing.T) {
	repo := &mockDocumentRepo{}
	app := newDocumentAppNoAuth(repo)

	resp := testRequest(app, http.MethodGet, "/documents/search?q=test", nil)
	assert.NotEqual(t, http.StatusOK, resp.StatusCode)
}

func TestSearchDocuments_Empty(t *testing.T) {
	repo := &mockDocumentRepo{}
	app := newDocumentApp(repo)

	resp := testRequest(app, http.MethodGet, "/documents/search", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeResponse(resp)
	assert.Equal(t, true, body["success"])
	data, ok := body["data"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, data, 0)
}

func TestSearchDocuments_WithQuery(t *testing.T) {
	doc := sampleDocument()
	repo := &mockDocumentRepo{
		searchFn: func(_ context.Context, orgID, query string, _ *models.DocumentFilter, _, _ int) ([]*models.DocumentSearchResult, error) {
			assert.Equal(t, testOrgID, orgID)
			assert.Equal(t, "test", query)
			return []*models.DocumentSearchResult{
				{Document: *doc, Relevance: 1.0},
			}, nil
		},
		countSearchFn: func(_ context.Context, _, _ string, _ *models.DocumentFilter) (int64, error) {
			return 1, nil
		},
	}
	app := newDocumentApp(repo)

	resp := testRequest(app, http.MethodGet, "/documents/search?q=test", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeResponse(resp)
	assert.Equal(t, true, body["success"])
	data := body["data"].([]interface{})
	assert.Len(t, data, 1)
}

func TestSearchDocuments_FilterByDocumentType(t *testing.T) {
	called := false
	repo := &mockDocumentRepo{
		searchFn: func(_ context.Context, _ string, _ string, filter *models.DocumentFilter, _, _ int) ([]*models.DocumentSearchResult, error) {
			called = true
			assert.Contains(t, filter.DocumentTypes, "PURCHASE_ORDER")
			return []*models.DocumentSearchResult{}, nil
		},
	}
	app := newDocumentApp(repo)

	resp := testRequest(app, http.MethodGet, "/documents/search?documentType=PURCHASE_ORDER", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.True(t, called)
}

func TestSearchDocuments_FilterByStatus(t *testing.T) {
	called := false
	repo := &mockDocumentRepo{
		searchFn: func(_ context.Context, _ string, _ string, filter *models.DocumentFilter, _, _ int) ([]*models.DocumentSearchResult, error) {
			called = true
			assert.Contains(t, filter.Statuses, "APPROVED")
			return []*models.DocumentSearchResult{}, nil
		},
	}
	app := newDocumentApp(repo)

	resp := testRequest(app, http.MethodGet, "/documents/search?status=APPROVED", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.True(t, called)
}

func TestSearchDocuments_FilterByDocumentNumber(t *testing.T) {
	called := false
	repo := &mockDocumentRepo{
		searchFn: func(_ context.Context, _ string, _ string, filter *models.DocumentFilter, _, _ int) ([]*models.DocumentSearchResult, error) {
			called = true
			assert.Equal(t, "DOC-001", filter.DocumentNumber)
			return []*models.DocumentSearchResult{}, nil
		},
	}
	app := newDocumentApp(repo)

	resp := testRequest(app, http.MethodGet, "/documents/search?documentNumber=DOC-001", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.True(t, called)
}

func TestSearchDocuments_FilterByDateRange(t *testing.T) {
	called := false
	repo := &mockDocumentRepo{
		searchFn: func(_ context.Context, _ string, _ string, filter *models.DocumentFilter, _, _ int) ([]*models.DocumentSearchResult, error) {
			called = true
			assert.NotNil(t, filter.DateFrom)
			assert.NotNil(t, filter.DateTo)
			return []*models.DocumentSearchResult{}, nil
		},
	}
	app := newDocumentApp(repo)

	resp := testRequest(app, http.MethodGet, "/documents/search?startDate=2026-01-01&endDate=2026-12-31", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.True(t, called)
}

func TestSearchDocuments_PageSizeParam(t *testing.T) {
	repo := &mockDocumentRepo{}
	app := newDocumentApp(repo)

	// Test 'pageSize' alias for 'limit'
	resp := testRequest(app, http.MethodGet, "/documents/search?pageSize=5&page=1", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeResponse(resp)
	assert.Equal(t, true, body["success"])
}

func TestSearchDocuments_RepoError(t *testing.T) {
	repo := &mockDocumentRepo{
		searchFn: func(_ context.Context, _ string, _ string, _ *models.DocumentFilter, _, _ int) ([]*models.DocumentSearchResult, error) {
			return nil, errors.New("search failed")
		},
	}
	app := newDocumentApp(repo)

	resp := testRequest(app, http.MethodGet, "/documents/search?q=test", nil)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// GET /documents/stats — GetDocumentStats
// ---------------------------------------------------------------------------

func TestGetDocumentStats_NoAuth(t *testing.T) {
	repo := &mockDocumentRepo{}
	app := newDocumentAppNoAuth(repo)

	resp := testRequest(app, http.MethodGet, "/documents/stats", nil)
	assert.NotEqual(t, http.StatusOK, resp.StatusCode)
}

func TestGetDocumentStats_Success(t *testing.T) {
	stats := &models.DocumentStats{
		TotalDocuments:    10,
		DocumentsByType:   map[string]int64{"REQUISITION": 5, "BUDGET": 5},
		DocumentsByStatus: map[string]int64{"DRAFT": 3, "APPROVED": 7},
		DocumentsByDept:   map[string]int64{"Finance": 10},
		RecentDocuments:   3,
		PendingApprovals:  2,
		TotalValue:        100000.0,
		AverageValue:      10000.0,
	}

	repo := &mockDocumentRepo{
		getStatsFn: func(_ context.Context, orgID string) (*models.DocumentStats, error) {
			assert.Equal(t, testOrgID, orgID)
			return stats, nil
		},
	}
	app := newDocumentApp(repo)

	resp := testRequest(app, http.MethodGet, "/documents/stats", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeResponse(resp)
	assert.Equal(t, true, body["success"])
	data := body["data"].(map[string]interface{})
	assert.Equal(t, float64(10), data["totalDocuments"])
}

func TestGetDocumentStats_RepoError(t *testing.T) {
	repo := &mockDocumentRepo{
		getStatsFn: func(_ context.Context, _ string) (*models.DocumentStats, error) {
			return nil, errors.New("stats query failed")
		},
	}
	app := newDocumentApp(repo)

	resp := testRequest(app, http.MethodGet, "/documents/stats", nil)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// GET /public/verify/:documentNumber — VerifyDocumentPublic
// ---------------------------------------------------------------------------

func TestVerifyDocumentPublic_MissingNumber(t *testing.T) {
	// Fiber routes won't match an empty param, so the router simply returns 404
	// for an unmatched path; we test an actually-registered path variant.
	repo := &mockDocumentRepo{}
	app := newDocumentApp(repo)

	// No auth required — public endpoint.
	// A plausible but unknown prefix causes fallback to GetByNumberOnly.
	resp := testRequest(app, http.MethodGet, "/public/verify/UNKNOWN-DOC-001", nil)
	// All repo calls return errors → 404
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestVerifyDocumentPublic_NotFound(t *testing.T) {
	repo := &mockDocumentRepo{
		getByNumberOnlyFn: func(_ context.Context, _ string) (*models.Document, error) {
			return nil, errors.New("not found")
		},
	}
	app := newDocumentApp(repo)

	resp := testRequest(app, http.MethodGet, "/public/verify/UNKNOWN-999", nil)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	body := decodeResponse(resp)
	assert.Equal(t, false, body["success"])
}

func TestVerifyDocumentPublic_SuccessGenericDoc(t *testing.T) {
	doc := sampleDocument()
	doc.DocumentNumber = "DOC-GENERIC-001"

	repo := &mockDocumentRepo{
		getByNumberOnlyFn: func(_ context.Context, number string) (*models.Document, error) {
			assert.Equal(t, "DOC-GENERIC-001", number)
			return doc, nil
		},
	}
	app := newDocumentApp(repo)

	resp := testRequest(app, http.MethodGet, "/public/verify/DOC-GENERIC-001", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeResponse(resp)
	assert.Equal(t, true, body["success"])
	data := body["data"].(map[string]interface{})
	assert.Equal(t, true, data["verified"])
	assert.Equal(t, "DOC-GENERIC-001", data["documentNumber"])
}

func TestVerifyDocumentPublic_RequisitionPrefix(t *testing.T) {
	// REQ- prefix → routes to GetRequisitionByNumberPublic
	req := &models.Requisition{
		ID:             uuid.New().String(),
		OrganizationID: testOrgID,
		DocumentNumber: "REQ-20260101-abcd1234",
		Title:          "Requisition for Supplies",
		Status:         "APPROVED",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	repo := &mockDocumentRepo{
		getRequisitionByNumberFn: func(_ context.Context, number string) (*models.Requisition, error) {
			assert.Equal(t, "REQ-20260101-abcd1234", number)
			return req, nil
		},
	}
	app := newDocumentApp(repo)

	resp := testRequest(app, http.MethodGet, "/public/verify/REQ-20260101-abcd1234", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeResponse(resp)
	assert.Equal(t, true, body["success"])
	data := body["data"].(map[string]interface{})
	assert.Equal(t, true, data["verified"])
	assert.Equal(t, "REQUISITION", data["documentType"])
}

func TestVerifyDocumentPublic_POPrefix(t *testing.T) {
	// PO- prefix → routes to GetPurchaseOrderByNumberPublic
	po := &models.PurchaseOrder{
		ID:             uuid.New().String(),
		OrganizationID: testOrgID,
		DocumentNumber: "PO-20260101-ef56",
		Status:         "APPROVED",
		TotalAmount:    5000.0,
		Currency:       "USD",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	repo := &mockDocumentRepo{
		getPurchaseOrderByNumberFn: func(_ context.Context, number string) (*models.PurchaseOrder, error) {
			return po, nil
		},
	}
	app := newDocumentApp(repo)

	resp := testRequest(app, http.MethodGet, "/public/verify/PO-20260101-ef56", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeResponse(resp)
	data := body["data"].(map[string]interface{})
	assert.Equal(t, "PURCHASE_ORDER", data["documentType"])
}

func TestVerifyDocumentPublic_PVPrefix(t *testing.T) {
	// PV- prefix → routes to GetPaymentVoucherByNumberPublic
	pv := &models.PaymentVoucher{
		ID:             uuid.New().String(),
		OrganizationID: testOrgID,
		DocumentNumber: "PV-20260101-gh78",
		Status:         "APPROVED",
		Amount:         2500.0,
		Currency:       "USD",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	repo := &mockDocumentRepo{
		getPaymentVoucherByNumberFn: func(_ context.Context, number string) (*models.PaymentVoucher, error) {
			return pv, nil
		},
	}
	app := newDocumentApp(repo)

	resp := testRequest(app, http.MethodGet, "/public/verify/PV-20260101-gh78", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeResponse(resp)
	data := body["data"].(map[string]interface{})
	assert.Equal(t, "PAYMENT_VOUCHER", data["documentType"])
}

func TestVerifyDocumentPublic_GRNPrefix(t *testing.T) {
	// GRN- prefix → routes to GetGRNByNumberPublic
	grn := &models.GoodsReceivedNote{
		ID:             uuid.New().String(),
		OrganizationID: testOrgID,
		DocumentNumber: "GRN-20260101-ij90",
		Status:         "CONFIRMED",
		ReceivedBy:     "warehouse-user",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	repo := &mockDocumentRepo{
		getGRNByNumberFn: func(_ context.Context, number string) (*models.GoodsReceivedNote, error) {
			return grn, nil
		},
	}
	app := newDocumentApp(repo)

	resp := testRequest(app, http.MethodGet, "/public/verify/GRN-20260101-ij90", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeResponse(resp)
	data := body["data"].(map[string]interface{})
	assert.Equal(t, "GRN", data["documentType"])
}

func TestVerifyDocumentPublic_NoAuthRequired(t *testing.T) {
	// Confirm the public endpoint does NOT require tenant locals.
	doc := sampleDocument()
	doc.DocumentNumber = "DOC-PUBLIC-001"

	repo := &mockDocumentRepo{
		getByNumberOnlyFn: func(_ context.Context, _ string) (*models.Document, error) {
			return doc, nil
		},
	}

	// Use the no-auth app — the public endpoint is still registered there.
	auditSvc := services.NewAuditService()
	docSvc := services.NewDocumentService(repo, auditSvc, nil)
	h := NewDocumentHandler(docSvc)

	app := fiber.New()
	app.Get("/public/verify/:documentNumber", h.VerifyDocumentPublic)

	resp := testRequest(app, http.MethodGet, "/public/verify/DOC-PUBLIC-001", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestVerifyDocumentPublic_CacheControlHeaders(t *testing.T) {
	doc := sampleDocument()
	doc.DocumentNumber = "DOC-CACHE-001"

	repo := &mockDocumentRepo{
		getByNumberOnlyFn: func(_ context.Context, _ string) (*models.Document, error) {
			return doc, nil
		},
	}
	app := newDocumentApp(repo)

	resp := testRequest(app, http.MethodGet, "/public/verify/DOC-CACHE-001", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "no-cache, no-store, must-revalidate", resp.Header.Get("Cache-Control"))
	assert.Equal(t, "no-cache", resp.Header.Get("Pragma"))
	assert.Equal(t, "0", resp.Header.Get("Expires"))
}
