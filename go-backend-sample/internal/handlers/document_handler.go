package handlers

import (
	"strconv"

	"github.com/cozyCodr/liyali-gateway/internal/db"
	"github.com/cozyCodr/liyali-gateway/internal/repository"
	"github.com/cozyCodr/liyali-gateway/internal/utils"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type DocumentHandler struct {
	docRepo repository.DocumentRepositoryInterface
}

func NewDocumentHandler(docRepo repository.DocumentRepositoryInterface) *DocumentHandler {
	return &DocumentHandler{
		docRepo: docRepo,
	}
}

// Request/Response Types
type CreateDocumentRequest struct {
	DocumentType   string  `json:"documentType" validate:"required"`
	Title          string  `json:"title" validate:"required"`
	Description    string  `json:"description"`
	Amount         float64 `json:"amount"`
	Currency       string  `json:"currency"`
	Department     string  `json:"department"`
	WorkflowID     *string `json:"workflowId"`
	Data           []byte  `json:"data" validate:"required"` // JSONB - type-specific fields
	Metadata       []byte  `json:"metadata"`                 // JSONB
}

type UpdateDocumentRequest struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	Department  string  `json:"department"`
	Data        []byte  `json:"data"`
	Metadata    []byte  `json:"metadata"`
}

// GetDocuments retrieves all documents with optional filtering
// GET /api/documents
func (h *DocumentHandler) GetDocuments(c fiber.Ctx) error {
	// Get query parameters
	documentType := c.Query("documentType", "")
	status := c.Query("status", "")
	department := c.Query("department", "")
	limitStr := c.Query("limit", "20")
	offsetStr := c.Query("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	var documents []db.Document
	var err error

	// Filter based on query parameters
	if documentType != "" && status != "" {
		documents, err = h.docRepo.ListDocumentsByTypeAndStatus(c.Context(), documentType, status, int32(limit), int32(offset))
	} else if documentType != "" {
		documents, err = h.docRepo.ListDocumentsByType(c.Context(), documentType, int32(limit), int32(offset))
	} else if status != "" {
		documents, err = h.docRepo.ListDocumentsByStatus(c.Context(), status, int32(limit), int32(offset))
	} else if department != "" {
		documents, err = h.docRepo.ListDocumentsByDepartment(c.Context(), department, int32(limit), int32(offset))
	} else {
		documents, err = h.docRepo.ListDocuments(c.Context(), int32(limit), int32(offset))
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve documents",
		})
	}

	// Get total count
	var total int64
	if documentType != "" {
		total, _ = h.docRepo.CountDocumentsByType(c.Context(), documentType)
	} else if status != "" {
		total, _ = h.docRepo.CountDocumentsByStatus(c.Context(), status)
	} else {
		total, _ = h.docRepo.CountDocuments(c.Context())
	}

	return c.JSON(fiber.Map{
		"documents": documents,
		"total":     total,
		"limit":     limit,
		"offset":    offset,
	})
}

// GetMyDocuments retrieves documents created by the authenticated user
// GET /api/documents/my
func (h *DocumentHandler) GetMyDocuments(c fiber.Ctx) error {
	// Get user ID from context
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	// Get query parameters
	limitStr := c.Query("limit", "20")
	offsetStr := c.Query("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	// Get documents
	documents, err := h.docRepo.ListDocumentsByCreator(c.Context(), userID, int32(limit), int32(offset))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve documents",
		})
	}

	// Get total count
	total, _ := h.docRepo.CountDocumentsByCreator(c.Context(), userID)

	return c.JSON(fiber.Map{
		"documents": documents,
		"total":     total,
		"limit":     limit,
		"offset":    offset,
	})
}

// GetDocumentByID retrieves a single document by ID
// GET /api/documents/:id
func (h *DocumentHandler) GetDocumentByID(c fiber.Ctx) error {
	// Get document ID from params
	documentIDStr := c.Params("id")
	documentID, err := uuid.Parse(documentIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid document ID",
		})
	}

	// Get document
	document, err := h.docRepo.GetDocumentByID(c.Context(), documentID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Document not found",
		})
	}

	return c.JSON(fiber.Map{
		"document": document,
	})
}

// GetDocumentByNumber retrieves a document by its document number
// GET /api/documents/number/:number
func (h *DocumentHandler) GetDocumentByNumber(c fiber.Ctx) error {
	// Get document number from params
	documentNumber := c.Params("number")
	if documentNumber == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Document number is required",
		})
	}

	// Get document
	document, err := h.docRepo.GetDocumentByNumber(c.Context(), documentNumber)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Document not found",
		})
	}

	return c.JSON(fiber.Map{
		"document": document,
	})
}

// CreateDocument creates a new document
// POST /api/documents
func (h *DocumentHandler) CreateDocument(c fiber.Ctx) error {
	// Get user ID from context
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	// Parse request body
	var req CreateDocumentRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate request
	if req.DocumentType == "" || req.Title == "" || req.Data == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Document type, title, and data are required",
		})
	}

	// Validate document type
	validTypes := map[string]bool{
		"REQUISITION":     true,
		"BUDGET":          true,
		"PURCHASE_ORDER":  true,
		"PAYMENT_VOUCHER": true,
		"GRN":             true,
	}
	if !validTypes[req.DocumentType] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid document type",
		})
	}

	// Generate document number (simple format: TYPE-UUID-short)
	docID := uuid.New()
	documentNumber := req.DocumentType + "-" + docID.String()[:8]

	// Parse workflow ID if provided
	var workflowID pgtype.UUID
	if req.WorkflowID != nil && *req.WorkflowID != "" {
		wfID, err := uuid.Parse(*req.WorkflowID)
		if err == nil {
			workflowID = utils.UUIDToPgtype(wfID)
		}
	}

	// Set default currency if not provided
	currency := req.Currency
	if currency == "" {
		currency = "USD"
	}

	// Create document
	document, err := h.docRepo.CreateDocument(c.Context(), db.CreateDocumentParams{
		DocumentType:   req.DocumentType,
		DocumentNumber: documentNumber,
		Title:          req.Title,
		Description:    pgtype.Text{String: req.Description, Valid: req.Description != ""},
		Amount:         pgtype.Numeric{Valid: req.Amount > 0}, // TODO: Convert float64 to pgtype.Numeric properly
		Currency:       pgtype.Text{String: currency, Valid: true},
		Status:         "DRAFT",
		CreatedBy:      utils.UUIDToPgtype(userID),
		Department:     pgtype.Text{String: req.Department, Valid: req.Department != ""},
		WorkflowID:     workflowID,
		Data:           req.Data,
		Metadata:       req.Metadata,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create document",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":  "Document created successfully",
		"document": document,
	})
}

// UpdateDocument updates an existing document
// PUT /api/documents/:id
func (h *DocumentHandler) UpdateDocument(c fiber.Ctx) error {
	// Get user ID from context (for authorization check)
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	// Get document ID from params
	documentIDStr := c.Params("id")
	documentID, err := uuid.Parse(documentIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid document ID",
		})
	}

	// Parse request body
	var req UpdateDocumentRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Get existing document to preserve fields and check ownership
	existing, err := h.docRepo.GetDocumentByID(c.Context(), documentID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Document not found",
		})
	}

	// Check if user is the creator (only allow creator to update in DRAFT status)
	if utils.PgtypeToUUID(existing.CreatedBy) != userID {
		// TODO: Add role-based permission check here
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "You are not authorized to update this document",
		})
	}

	// Only allow updates if document is in DRAFT status
	if existing.Status != "DRAFT" {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "Cannot update document that is not in DRAFT status",
		})
	}

	// Prepare update params (preserve existing values if not provided)
	title := req.Title
	if title == "" {
		title = existing.Title
	}

	description := req.Description
	if description == "" {
		description = utils.PgtypeToString(existing.Description)
	}

	currency := req.Currency
	if currency == "" {
		currency = utils.PgtypeToString(existing.Currency)
	}

	data := req.Data
	if data == nil {
		data = existing.Data
	}

	metadata := req.Metadata
	if metadata == nil {
		metadata = existing.Metadata
	}

	// Update document
	document, err := h.docRepo.UpdateDocument(c.Context(), db.UpdateDocumentParams{
		ID:          utils.UUIDToPgtype(documentID),
		Title:       title,
		Description: pgtype.Text{String: description, Valid: description != ""},
		Amount:      pgtype.Numeric{Valid: req.Amount > 0}, // TODO: Convert properly
		Currency:    pgtype.Text{String: currency, Valid: true},
		Data:        data,
		Metadata:    metadata,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update document",
		})
	}

	return c.JSON(fiber.Map{
		"message":  "Document updated successfully",
		"document": document,
	})
}

// SubmitDocument submits a document for approval
// POST /api/documents/:id/submit
func (h *DocumentHandler) SubmitDocument(c fiber.Ctx) error {
	// Get user ID from context
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	// Get document ID from params
	documentIDStr := c.Params("id")
	documentID, err := uuid.Parse(documentIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid document ID",
		})
	}

	// Get existing document to check ownership and status
	existing, err := h.docRepo.GetDocumentByID(c.Context(), documentID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Document not found",
		})
	}

	// Check if user is the creator
	if utils.PgtypeToUUID(existing.CreatedBy) != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "You are not authorized to submit this document",
		})
	}

	// Only allow submission if document is in DRAFT status
	if existing.Status != "DRAFT" {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "Document is not in DRAFT status",
		})
	}

	// Submit document
	document, err := h.docRepo.SubmitDocument(c.Context(), documentID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to submit document",
		})
	}

	// TODO: Create approval tasks based on workflow

	return c.JSON(fiber.Map{
		"message":  "Document submitted successfully",
		"document": document,
	})
}

// DeleteDocument deletes a document
// DELETE /api/documents/:id
func (h *DocumentHandler) DeleteDocument(c fiber.Ctx) error {
	// Get user ID from context
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	// Get document ID from params
	documentIDStr := c.Params("id")
	documentID, err := uuid.Parse(documentIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid document ID",
		})
	}

	// Get existing document to check ownership
	existing, err := h.docRepo.GetDocumentByID(c.Context(), documentID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Document not found",
		})
	}

	// Check if user is the creator or has admin role
	if utils.PgtypeToUUID(existing.CreatedBy) != userID {
		// TODO: Check if user has ADMIN role
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "You are not authorized to delete this document",
		})
	}

	// Only allow deletion if document is in DRAFT or REJECTED status
	if existing.Status != "DRAFT" && existing.Status != "REJECTED" {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "Cannot delete document that is in approval process",
		})
	}

	// Delete document
	err = h.docRepo.DeleteDocument(c.Context(), documentID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete document",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Document deleted successfully",
	})
}
