package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/utils"
)

// RequireOrgParamMembership authorizes the caller against the organization named
// in the ":id" URL parameter (as opposed to TenantMiddleware, which uses the
// X-Organization-ID header). Use it on routes like /organizations/:id/... that
// take the target org from the path so the caller cannot act on an arbitrary
// org by guessing its ID.
//
//   - requireManage=false → caller must be an active member of the org.
//   - requireManage=true  → caller must be an active "admin" member of the org.
//
// super_admin bypasses both checks (platform-wide access). Must run after
// AuthMiddleware so userID/userRole locals are populated.
func RequireOrgParamMembership(requireManage bool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, _ := c.Locals("userID").(string)
		if userID == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(utils.ErrorResponse("Authentication required"))
		}

		// Platform super-admins have org-wide access (e.g. to manage any tenant's
		// subscription from the admin console).
		if role, _ := c.Locals("userRole").(string); role == "super_admin" {
			return c.Next()
		}

		orgID := c.Params("id")
		if orgID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse("Organization ID is required"))
		}

		var membership models.OrganizationMember
		if err := config.DB.Where(
			"organization_id = ? AND user_id = ? AND active = ?",
			orgID, userID, true,
		).First(&membership).Error; err != nil {
			return c.Status(fiber.StatusForbidden).JSON(utils.ErrorResponse("You are not a member of this organization"))
		}

		if requireManage && membership.Role != "admin" {
			return c.Status(fiber.StatusForbidden).JSON(utils.ErrorResponse("You do not have permission to manage this organization"))
		}

		return c.Next()
	}
}
