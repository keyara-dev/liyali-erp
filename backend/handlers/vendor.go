package handlers

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/types"
)

// GetVendors retrieves all vendors with pagination and filtering
func GetVendors(c fiber.Ctx) error {
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

	query := db
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

	return c.JSON(types.ListResponse{
		Success: true,
		Data:    responses,
		Total:   total,
		Page:    page,
		Limit:   limit,
	})
}

// CreateVendor creates a new vendor
func CreateVendor(c fiber.Ctx) error {
	var req types.CreateVendorRequest

	if err := c.BindJSON(&req); err != nil {
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
	if req.BankAccount == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Bank account is required",
		})
	}
	if req.TaxID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Tax ID is required",
		})
	}

	// Check if vendor with same email already exists
	var existingVendor models.Vendor
	if err := config.DB.Where("email = ?", req.Email).First(&existingVendor).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"success": false,
			"message": "Vendor with this email already exists",
		})
	}

	// Generate vendor code
	vendorCode := fmt.Sprintf("VND-%d-%s", time.Now().Unix(), uuid.New().String()[:6])

	vendor := models.Vendor{
		ID:          uuid.New().String(),
		VendorCode:  vendorCode,
		Name:        req.Name,
		Email:       req.Email,
		Phone:       req.Phone,
		Country:     req.Country,
		City:        req.City,
		BankAccount: req.BankAccount,
		TaxID:       req.TaxID,
		Active:      true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
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
func GetVendor(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Vendor ID is required",
		})
	}

	var vendor models.Vendor
	if err := config.DB.Where("id = ?", id).First(&vendor).Error; err != nil {
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
func UpdateVendor(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Vendor ID is required",
		})
	}

	var req types.UpdateVendorRequest
	if err := c.BindJSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	var vendor models.Vendor
	if err := config.DB.Where("id = ?", id).First(&vendor).Error; err != nil {
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
		// Check if email is already used by another vendor
		var existingVendor models.Vendor
		if err := config.DB.Where("email = ? AND id != ?", req.Email, id).First(&existingVendor).Error; err == nil {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"success": false,
				"message": "Vendor with this email already exists",
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
func DeleteVendor(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Vendor ID is required",
		})
	}

	var vendor models.Vendor
	if err := config.DB.Where("id = ?", id).First(&vendor).Error; err != nil {
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
		ID:          vendor.ID,
		VendorCode:  vendor.VendorCode,
		Name:        vendor.Name,
		Email:       vendor.Email,
		Phone:       vendor.Phone,
		Country:     vendor.Country,
		City:        vendor.City,
		BankAccount: vendor.BankAccount,
		TaxID:       vendor.TaxID,
		Active:      vendor.Active,
		CreatedAt:   vendor.CreatedAt,
		UpdatedAt:   vendor.UpdatedAt,
	}
}
