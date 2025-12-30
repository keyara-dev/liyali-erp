package handlers

import (
	"github.com/gofiber/fiber/v2"
)

// Health check endpoint
func HealthCheck(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status": "ok",
		"message": "Liyali Gateway Backend API is running",
	})
}