package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/utils"
)

// GetProvinces returns all Zambian provinces ordered by name.
// GET /api/v1/provinces
func GetProvinces(c *fiber.Ctx) error {
	var provinces []models.Province
	if err := config.DB.Order("name ASC").Find(&provinces).Error; err != nil {
		return utils.SendInternalError(c, "Failed to retrieve provinces", err)
	}
	return utils.SendSimpleSuccess(c, provinces, "Provinces retrieved successfully")
}

// GetTowns returns towns/districts, optionally filtered by province_id.
// GET /api/v1/towns?province_id=<uuid>
func GetTowns(c *fiber.Ctx) error {
	query := config.DB.Model(&models.Town{})
	if pid := c.Query("province_id"); pid != "" {
		query = query.Where("province_id = ?", pid)
	}
	var towns []models.Town
	if err := query.Order("name ASC").Find(&towns).Error; err != nil {
		return utils.SendInternalError(c, "Failed to retrieve towns", err)
	}
	return utils.SendSimpleSuccess(c, towns, "Towns retrieved successfully")
}
