package middleware

import (
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/utils"
)

// TenantMiddleware extracts and validates organization context.
// Must be used after AuthMiddleware.
// super_admin bypasses org membership — they have platform-wide access.
func TenantMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 1. Get user ID from auth middleware
		userIDRaw := c.Locals("userID")

		if userIDRaw == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "User context required - userID is nil",
			})
		}

		userID, ok := userIDRaw.(string)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": fmt.Sprintf("User context required - userID is not a string, got %T", userIDRaw),
			})
		}

		if userID == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "User context required - userID is empty string",
			})
		}

		// 2. Get organization ID from header
		orgID := c.Get("X-Organization-ID")

		// 3. super_admin bypasses org membership validation — platform-wide access
		userRole, _ := c.Locals("userRole").(string)
		if userRole == "super_admin" {
			// Prefer org from header; fall back to user's current org if set
			if orgID == "" {
				var user models.User
				if err := config.DB.Where("id = ?", userID).First(&user).Error; err == nil &&
					user.CurrentOrganizationID != nil {
					orgID = *user.CurrentOrganizationID
				}
			}
			tenantCtx := &utils.TenantContext{
				OrganizationID: orgID,
				UserID:         userID,
				UserRole:       "super_admin",
				Department:     "",
			}
			c.Locals("tenant", tenantCtx)
			c.Locals("organizationID", orgID)
			return c.Next()
		}

		// 4. Non-super_admin: resolve org from header or user's current org
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

		// 5. Verify user is an active member of this organization
		var membership models.OrganizationMember
		if err := config.DB.Where(
			"organization_id = ? AND user_id = ? AND active = ?",
			orgID, userID, true,
		).First(&membership).Error; err != nil {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": fmt.Sprintf("You are not a member of this organization: %v", err),
			})
		}

		// 5b. Block access to suspended organizations (org-level enforcement).
		var org models.Organization
		if err := config.DB.Select("id", "active").Where("id = ?", orgID).First(&org).Error; err != nil {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Organization not found",
			})
		}
		if !org.Active {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "This organization is suspended. Please contact support.",
			})
		}

		// 6. Store tenant context
		tenantCtx := &utils.TenantContext{
			OrganizationID: orgID,
			UserID:         userID,
			UserRole:       membership.Role,
			Department:     membership.Department,
		}
		c.Locals("tenant", tenantCtx)
		c.Locals("organizationID", orgID)

		return c.Next()
	}
}

// GetTenantContext retrieves tenant context from Fiber context
func GetTenantContext(c *fiber.Ctx) (*utils.TenantContext, error) {
	tenant, ok := c.Locals("tenant").(*utils.TenantContext)
	if !ok {
		return nil, errors.New("tenant context not found")
	}
	return tenant, nil
}
