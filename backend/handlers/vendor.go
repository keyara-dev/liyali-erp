package handlers

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/logging"
	"github.com/liyali/liyali-gateway/middleware"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/types"
	"github.com/liyali/liyali-gateway/utils"
)

// GetVendors retrieves all vendors with pagination and filtering
func GetVendors(c *fiber.Ctx) error {
	logger := logging.FromContext(c)
	logger.Info("get_vendors_request")

	// Get organization context from tenant middleware
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

	active := c.Query("active")
	country := c.Query("country")

	// Add query parameters to context
	logging.AddFieldsToRequest(c, map[string]interface{}{
		"operation":      "get_vendors",
		"page":           page,
		"limit":          limit,
		"active":         active,
		"country":        country,
		"organizationID": tenant.OrganizationID,
	})

	// SECURITY: Always filter by organization ID first
	query := db.Where("organization_id = ?", tenant.OrganizationID)

	if active == "true" {
		query = query.Where("active = ?", true)
	} else if active == "false" {
		query = query.Where("active = ?", false)
	}
	if country != "" {
		query = query.Where("country = ?", country)
	}

	var total int64
	if err := query.Model(&models.Vendor{}).Count(&total).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to count vendors",
			"error":   err.Error(),
		})
	}

	var vendors []models.Vendor
	offset := (page - 1) * limit
	if err := query.
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&vendors).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to fetch vendors",
			"error":   err.Error(),
		})
	}

	responses := make([]types.VendorResponse, 0, len(vendors))
	for _, vendor := range vendors {
		responses = append(responses, modelToVendorResponse(vendor))
	}

	return utils.SendPaginatedSuccess(c, responses, "Vendors retrieved successfully", page, limit, total)
}

// CreateVendor creates a new vendor
func CreateVendor(c *fiber.Ctx) error {
	// Get organization context from tenant middleware
	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Organization context required",
		})
	}

	var req types.CreateVendorRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	if req.Name == "" || len(req.Name) < 3 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Vendor name is required and must be at least 3 characters",
		})
	}
	if req.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Email is required",
		})
	}
	// Basic email validation
	if len(req.Email) < 5 || !strings.Contains(req.Email, "@") || !strings.Contains(req.Email, ".") {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid email format",
		})
	}
	if req.Phone == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Phone is required",
		})
	}
	if req.Country == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Country is required",
		})
	}
	if req.City == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "City is required",
		})
	}
	if req.TaxID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Tax ID is required",
		})
	}

	// SECURITY: Check if vendor with same email already exists in THIS organization
	var existingVendor models.Vendor
	if err := config.DB.Where("email = ? AND organization_id = ?", req.Email, tenant.OrganizationID).First(&existingVendor).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"success": false,
			"message": "Vendor with this email already exists in your organization",
		})
	}

	// Generate vendor code
	vendorCode := utils.GenerateVendorCode()

	vendor := models.Vendor{
		ID:              uuid.New().String(),
		OrganizationID:  tenant.OrganizationID, // SECURITY: Set organization ID
		VendorCode:      vendorCode,
		Name:            req.Name,
		Email:           req.Email,
		Phone:           req.Phone,
		Country:         req.Country,
		City:            req.City,
		BankAccount:     req.BankAccount,
		TaxID:           req.TaxID,
		Active:          true,
		CreatedBy:       tenant.UserID,
		BankName:        req.BankName,
		AccountName:     req.AccountName,
		AccountNumber:   req.AccountNumber,
		BranchCode:      req.BranchCode,
		SwiftCode:       req.SwiftCode,
		ContactPerson:   req.ContactPerson,
		PhysicalAddress: req.PhysicalAddress,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := config.DB.Create(&vendor).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to create vendor",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(types.DetailResponse{
		Success: true,
		Data:    modelToVendorResponse(vendor),
	})
}

// GetVendor retrieves a single vendor by ID
func GetVendor(c *fiber.Ctx) error {
	// Get organization context from tenant middleware
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
			"message": "Vendor ID is required",
		})
	}

	var vendor models.Vendor
	// SECURITY: Filter by organization ID to prevent cross-organization access
	if err := config.DB.Where("id = ? AND organization_id = ?", id, tenant.OrganizationID).First(&vendor).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Vendor not found",
		})
	}

	return c.JSON(types.DetailResponse{
		Success: true,
		Data:    modelToVendorResponse(vendor),
	})
}

// UpdateVendor updates an existing vendor
func UpdateVendor(c *fiber.Ctx) error {
	// Get organization context from tenant middleware
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
			"message": "Vendor ID is required",
		})
	}

	var req types.UpdateVendorRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	var vendor models.Vendor
	// SECURITY: Filter by organization ID to prevent cross-organization access
	if err := config.DB.Where("id = ? AND organization_id = ?", id, tenant.OrganizationID).First(&vendor).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Vendor not found",
		})
	}

	if req.Name != "" {
		if len(req.Name) < 3 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Vendor name must be at least 3 characters",
			})
		}
		vendor.Name = req.Name
	}
	if req.Email != "" {
		// SECURITY: Check if email is already used by another vendor in THIS organization
		var existingVendor models.Vendor
		if err := config.DB.Where("email = ? AND id != ? AND organization_id = ?", req.Email, id, tenant.OrganizationID).First(&existingVendor).Error; err == nil {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"success": false,
				"message": "Vendor with this email already exists in your organization",
			})
		}
		vendor.Email = req.Email
	}
	if req.Phone != "" {
		vendor.Phone = req.Phone
	}
	if req.Country != "" {
		vendor.Country = req.Country
	}
	if req.City != "" {
		vendor.City = req.City
	}
	if req.BankAccount != "" {
		vendor.BankAccount = req.BankAccount
	}
	if req.TaxID != "" {
		vendor.TaxID = req.TaxID
	}
	if req.BankName != "" {
		vendor.BankName = req.BankName
	}
	if req.AccountName != "" {
		vendor.AccountName = req.AccountName
	}
	if req.AccountNumber != "" {
		vendor.AccountNumber = req.AccountNumber
	}
	if req.BranchCode != "" {
		vendor.BranchCode = req.BranchCode
	}
	if req.SwiftCode != "" {
		vendor.SwiftCode = req.SwiftCode
	}
	if req.ContactPerson != "" {
		vendor.ContactPerson = req.ContactPerson
	}
	if req.PhysicalAddress != "" {
		vendor.PhysicalAddress = req.PhysicalAddress
	}

	// Update active status (if explicitly provided in request)
	if c.Query("active") != "" {
		vendor.Active = req.Active
	}

	vendor.UpdatedAt = time.Now()

	if err := config.DB.Save(&vendor).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to update vendor",
			"error":   err.Error(),
		})
	}

	return c.JSON(types.DetailResponse{
		Success: true,
		Data:    modelToVendorResponse(vendor),
	})
}

// DeleteVendor deactivates a vendor (soft delete via active flag)
func DeleteVendor(c *fiber.Ctx) error {
	// Get organization context from tenant middleware
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
			"message": "Vendor ID is required",
		})
	}

	var vendor models.Vendor
	// SECURITY: Filter by organization ID to prevent cross-organization access
	if err := config.DB.Where("id = ? AND organization_id = ?", id, tenant.OrganizationID).First(&vendor).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Vendor not found",
		})
	}

	// Soft delete by marking as inactive
	vendor.Active = false
	vendor.UpdatedAt = time.Now()

	if err := config.DB.Save(&vendor).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to delete vendor",
			"error":   err.Error(),
		})
	}

	return c.JSON(types.MessageResponse{
		Success: true,
		Message: "Vendor deactivated successfully",
	})
}

// Helper function to convert model to response
func modelToVendorResponse(vendor models.Vendor) types.VendorResponse {
	return types.VendorResponse{
		ID:              vendor.ID,
		VendorCode:      vendor.VendorCode,
		Name:            vendor.Name,
		Email:           vendor.Email,
		Phone:           vendor.Phone,
		Country:         vendor.Country,
		City:            vendor.City,
		BankAccount:     vendor.BankAccount,
		TaxID:           vendor.TaxID,
		Active:          vendor.Active,
		BankName:        vendor.BankName,
		AccountName:     vendor.AccountName,
		AccountNumber:   vendor.AccountNumber,
		BranchCode:      vendor.BranchCode,
		SwiftCode:       vendor.SwiftCode,
		ContactPerson:   vendor.ContactPerson,
		PhysicalAddress: vendor.PhysicalAddress,
		CreatedAt:       vendor.CreatedAt,
		UpdatedAt:       vendor.UpdatedAt,
	}
}
