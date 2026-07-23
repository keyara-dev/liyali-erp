package handlers

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/middleware"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/types"
	"github.com/liyali/liyali-gateway/utils"
)

// validPayeeTypes is the set of allowed payeeType values.
var validPayeeTypes = map[string]bool{
	"vendor":   true,
	"employee": true,
	"other":    true,
}

// GetPayees returns a paginated, org-scoped list of payees.
// Supports ?type= (payee type filter) and ?q= (name search).
func GetPayees(c *fiber.Ctx) error {
	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Organization context required",
		})
	}

	db := config.DB

	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	payeeType := c.Query("type")
	q := c.Query("q")

	// SECURITY: Always scope by organization.
	query := db.Where("organization_id = ?", tenant.OrganizationID)

	if payeeType != "" {
		query = query.Where("payee_type = ?", payeeType)
	}
	if q != "" {
		query = query.Where("LOWER(name) LIKE LOWER(?)", "%"+q+"%")
	}

	var total int64
	if err := query.Model(&models.Payee{}).Count(&total).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to count payees",
			"error":   err.Error(),
		})
	}

	var payees []models.Payee
	offset := (page - 1) * limit
	if err := query.
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&payees).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to fetch payees",
			"error":   err.Error(),
		})
	}

	responses := make([]types.PayeeResponse, 0, len(payees))
	for _, p := range payees {
		responses = append(responses, modelToPayeeResponse(p))
	}

	return utils.SendPaginatedSuccess(c, responses, "Payees retrieved successfully", page, limit, total)
}

// CreatePayee creates a new payee record scoped to the current organization.
func CreatePayee(c *fiber.Ctx) error {
	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Organization context required",
		})
	}

	var req types.CreatePayeeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	// Validate payeeType.
	if !validPayeeTypes[strings.ToLower(req.PayeeType)] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "payeeType must be one of: vendor, employee, other",
		})
	}

	// Validate name.
	if strings.TrimSpace(req.Name) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "name is required",
		})
	}

	createdBy := tenant.UserID
	payee := models.Payee{
		ID:             uuid.New().String(),
		OrganizationID: tenant.OrganizationID,
		PayeeType:      strings.ToLower(req.PayeeType),
		Name:           req.Name,
		Email:          req.Email,
		Phone:          req.Phone,
		BankName:       req.BankName,
		BankAccount:    req.BankAccount,
		TaxID:          req.TaxID,
		CreatedBy:      &createdBy,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := config.DB.Create(&payee).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to create payee",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(types.DetailResponse{
		Success: true,
		Data:    modelToPayeeResponse(payee),
	})
}

// GetPayee retrieves a single payee by ID, scoped to the current organization.
func GetPayee(c *fiber.Ctx) error {
	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Organization context required",
		})
	}

	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Payee ID is required",
		})
	}

	var payee models.Payee
	// SECURITY: filter by organization_id.
	if err := config.DB.Where("id = ? AND organization_id = ?", id, tenant.OrganizationID).First(&payee).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Payee not found",
		})
	}

	return c.JSON(types.DetailResponse{
		Success: true,
		Data:    modelToPayeeResponse(payee),
	})
}

// UpdatePayee partially updates an existing payee.
func UpdatePayee(c *fiber.Ctx) error {
	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Organization context required",
		})
	}

	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Payee ID is required",
		})
	}

	var req types.UpdatePayeeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	var payee models.Payee
	// SECURITY: filter by organization_id.
	if err := config.DB.Where("id = ? AND organization_id = ?", id, tenant.OrganizationID).First(&payee).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Payee not found",
		})
	}

	if req.PayeeType != nil {
		if !validPayeeTypes[strings.ToLower(*req.PayeeType)] {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "payeeType must be one of: vendor, employee, other",
			})
		}
		payee.PayeeType = strings.ToLower(*req.PayeeType)
	}
	if req.Name != "" {
		payee.Name = req.Name
	}
	if req.Email != "" {
		payee.Email = req.Email
	}
	if req.Phone != "" {
		payee.Phone = req.Phone
	}
	if req.BankName != "" {
		payee.BankName = req.BankName
	}
	if req.BankAccount != "" {
		payee.BankAccount = req.BankAccount
	}
	if req.TaxID != "" {
		payee.TaxID = req.TaxID
	}

	payee.UpdatedAt = time.Now()

	if err := config.DB.Save(&payee).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to update payee",
			"error":   err.Error(),
		})
	}

	return c.JSON(types.DetailResponse{
		Success: true,
		Data:    modelToPayeeResponse(payee),
	})
}

// DeletePayee soft-deletes a payee via GORM's DeletedAt mechanism, returns 204.
func DeletePayee(c *fiber.Ctx) error {
	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Organization context required",
		})
	}

	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Payee ID is required",
		})
	}

	var payee models.Payee
	// SECURITY: filter by organization_id.
	if err := config.DB.Where("id = ? AND organization_id = ?", id, tenant.OrganizationID).First(&payee).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Payee not found",
		})
	}

	// Soft-delete via GORM (sets deleted_at timestamp).
	if err := config.DB.Delete(&payee).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to delete payee",
			"error":   err.Error(),
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// modelToPayeeResponse converts a Payee model to its response DTO.
func modelToPayeeResponse(p models.Payee) types.PayeeResponse {
	return types.PayeeResponse{
		ID:             p.ID,
		OrganizationID: p.OrganizationID,
		PayeeType:      p.PayeeType,
		Name:           p.Name,
		Email:          p.Email,
		Phone:          p.Phone,
		BankName:       p.BankName,
		BankAccount:    p.BankAccount,
		TaxID:          p.TaxID,
		SourceVendorID: p.SourceVendorID,
		SourceUserID:   p.SourceUserID,
		CreatedBy:      p.CreatedBy,
		CreatedAt:      p.CreatedAt,
		UpdatedAt:      p.UpdatedAt,
	}
}
