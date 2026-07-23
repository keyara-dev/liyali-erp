package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/logging"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/utils"
)

// AdminMiddleware ensures the user has admin privileges
func AdminMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		logger := logging.FromContext(c)

		userRole, ok := c.Locals("userRole").(string)
		if !ok {
			logger.Error("User role not found in context")
			return c.Status(fiber.StatusUnauthorized).JSON(utils.ErrorResponse("Authentication required"))
		}

		adminRoles := []string{"admin", "super_admin"}
		isAdmin := false
		for _, role := range adminRoles {
			if userRole == role {
				isAdmin = true
				break
			}
		}

		if !isAdmin {
			logger.WithFields(map[string]interface{}{
				"user_role":      userRole,
				"required_roles": adminRoles,
			}).Warn("User attempted to access admin endpoint without proper role")
			return c.Status(fiber.StatusForbidden).JSON(utils.ErrorResponse("Admin privileges required"))
		}

		return c.Next()
	}
}

// SuperAdminMiddleware ensures the user has super admin privileges
func SuperAdminMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		logger := logging.FromContext(c)

		userRole, ok := c.Locals("userRole").(string)
		if !ok {
			logger.Error("User role not found in context")
			return c.Status(fiber.StatusUnauthorized).JSON(utils.ErrorResponse("Authentication required"))
		}

		if userRole != "super_admin" {
			logger.WithFields(map[string]interface{}{
				"user_role":     userRole,
				"required_role": "super_admin",
			}).Warn("User attempted to access super admin endpoint without proper role")
			return c.Status(fiber.StatusForbidden).JSON(utils.ErrorResponse("Super admin privileges required"))
		}

		return c.Next()
	}
}

// OrganizationAdminMiddleware ensures the user is an admin of the specific organization.
// Super admins bypass the check (platform-wide access). All other users must be an
// active admin/owner member of that specific org in the DB — this prevents the previous
// cross-tenant bypass where any global "admin" could access any arbitrary org.
func OrganizationAdminMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		logger := logging.FromContext(c)

		organizationID := c.Params("id")
		if organizationID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse("Organization ID is required"))
		}

		userID, ok := c.Locals("userID").(string)
		if !ok {
			logger.Error("User ID not found in context")
			return c.Status(fiber.StatusUnauthorized).JSON(utils.ErrorResponse("Authentication required"))
		}

		userRole, ok := c.Locals("userRole").(string)
		if !ok {
			logger.Error("User role not found in context")
			return c.Status(fiber.StatusUnauthorized).JSON(utils.ErrorResponse("Authentication required"))
		}

		// Super admins have platform-wide access
		if userRole == "super_admin" {
			return c.Next()
		}

		// All other users: validate they are an active admin/owner of this specific org
		var membership models.OrganizationMember
		if err := config.DB.Where(
			"organization_id = ? AND user_id = ? AND active = ? AND role IN ('admin','owner')",
			organizationID, userID, true,
		).First(&membership).Error; err != nil {
			logger.WithFields(map[string]interface{}{
				"user_id":         userID,
				"user_role":       userRole,
				"organization_id": organizationID,
			}).Warn("User is not an admin member of the requested organization")
			return c.Status(fiber.StatusForbidden).JSON(utils.ErrorResponse("Organization admin privileges required"))
		}

		logger.WithFields(map[string]interface{}{
			"user_id":         userID,
			"org_role":        membership.Role,
			"organization_id": organizationID,
		}).Info("Organization admin access granted")

		return c.Next()
	}
}
