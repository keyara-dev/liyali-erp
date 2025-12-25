package middleware

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/models"
)

// TenantContext holds the current organization context
type TenantContext struct {
	OrganizationID string
	UserID         string
	UserRole       string
	Department     string
}

// TenantMiddleware extracts and validates organization context
// Must be used after AuthMiddleware
func TenantMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		// 1. Get user ID from auth middleware (must come after AuthMiddleware)
		userID, ok := c.Locals("userID").(string)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "User context required",
			})
		}

		// 2. Get organization ID from header
		orgID := c.Get("X-Organization-ID")

		// 3. If no header, get user's current organization
		if orgID == "" {
			var user models.User
			if err := config.DB.Where("id = ?", userID).First(&user).Error; err != nil {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error": "User not found",
				})
			}

			if user.CurrentOrganizationID == nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "No organization context. Please select an organization.",
				})
			}

			orgID = *user.CurrentOrganizationID
		}

		// 4. Verify user is member of this organization
		var membership models.OrganizationMember
		if err := config.DB.Where(
			"organization_id = ? AND user_id = ? AND active = ?",
			orgID, userID, true,
		).First(&membership).Error; err != nil {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "You are not a member of this organization",
			})
		}

		// 5. Create tenant context
		tenantCtx := &TenantContext{
			OrganizationID: orgID,
			UserID:         userID,
			UserRole:       membership.Role,
			Department:     membership.Department,
		}

		// 6. Store in context
		c.Locals("tenant", tenantCtx)
		c.Locals("organizationId", orgID) // For easy access

		return c.Next()
	}
}

// GetTenantContext retrieves tenant context from Fiber context
func GetTenantContext(c fiber.Ctx) (*TenantContext, error) {
	tenant, ok := c.Locals("tenant").(*TenantContext)
	if !ok {
		return nil, errors.New("tenant context not found")
	}
	return tenant, nil
}
