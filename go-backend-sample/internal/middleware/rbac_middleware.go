package middleware

import (
	"github.com/cozyCodr/liyali-gateway/internal/rbac"
	"github.com/gofiber/fiber/v3"
)

// RequirePermission creates middleware that requires specific permission
func RequirePermission(permission rbac.Permission) fiber.Handler {
	return func(c fiber.Ctx) error {
		// Get user role from context (set by auth middleware)
		userRole, ok := GetUserRole(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "user role not found in context",
			})
		}

		// Check if role has permission
		if !rbac.HasPermission(rbac.Role(userRole), permission) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "insufficient permissions",
			})
		}

		return c.Next()
	}
}

// RequireAnyPermission creates middleware that requires any of the specified permissions
func RequireAnyPermission(permissions ...rbac.Permission) fiber.Handler {
	return func(c fiber.Ctx) error {
		// Get user role from context
		userRole, ok := GetUserRole(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "user role not found in context",
			})
		}

		// Check if role has any of the permissions
		if !rbac.HasAnyPermission(rbac.Role(userRole), permissions...) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "insufficient permissions",
			})
		}

		return c.Next()
	}
}

// RequireAllPermissions creates middleware that requires all of the specified permissions
func RequireAllPermissions(permissions ...rbac.Permission) fiber.Handler {
	return func(c fiber.Ctx) error {
		// Get user role from context
		userRole, ok := GetUserRole(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "user role not found in context",
			})
		}

		// Check if role has all of the permissions
		if !rbac.HasAllPermissions(rbac.Role(userRole), permissions...) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "insufficient permissions",
			})
		}

		return c.Next()
	}
}

// RequireRole creates middleware that requires a specific role
func RequireRole(role rbac.Role) fiber.Handler {
	return func(c fiber.Ctx) error {
		// Get user role from context
		userRole, ok := GetUserRole(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "user role not found in context",
			})
		}

		// Check if user has the required role
		if rbac.Role(userRole) != role {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "insufficient privileges",
			})
		}

		return c.Next()
	}
}

// RequireAnyRole creates middleware that requires any of the specified roles
func RequireAnyRole(roles ...rbac.Role) fiber.Handler {
	return func(c fiber.Ctx) error {
		// Get user role from context
		userRole, ok := GetUserRole(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "user role not found in context",
			})
		}

		// Check if user has any of the required roles
		for _, role := range roles {
			if rbac.Role(userRole) == role {
				return c.Next()
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "insufficient privileges",
		})
	}
}
