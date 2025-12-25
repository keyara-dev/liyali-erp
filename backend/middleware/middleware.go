package middleware

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/liyali/liyali-gateway/services"
	"gorm.io/gorm"
)

// CORS middleware
func CORSMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		c.Set("Access-Control-Allow-Origin", os.Getenv("FRONTEND_URL"))
		c.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		c.Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Set("Access-Control-Allow-Credentials", "true")

		if c.Method() == "OPTIONS" {
			return c.SendStatus(200)
		}

		return c.Next()
	}
}

// AuthMiddleware validates JWT token
func AuthMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authorization header required",
			})
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid authorization header format",
			})
		}

		tokenString := parts[1]

		// Parse and validate token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token",
			})
		}

		// Store user ID in context for later use
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			c.Locals("userID", claims["sub"])
			c.Locals("userRole", claims["role"])
		}

		return c.Next()
	}
}

// LoggerMiddleware logs request details
func LoggerMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()

		// Process request
		err := c.Next()

		// Log request details
		duration := time.Since(start).String()
		method := c.Method()
		path := c.Path()
		status := c.Response().StatusCode()

		log.Printf("[%s] %s %s - %d (%s)", method, path, status, duration)

		return err
	}
}

// RoleBasedAccess checks if user has required role
func RoleBasedAccess(requiredRoles ...string) fiber.Handler {
	return func(c fiber.Ctx) error {
		userRole, ok := c.Locals("userRole").(string)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "User role not found in context",
			})
		}

		// Check if user role is in required roles
		for _, role := range requiredRoles {
			if userRole == role {
				return c.Next()
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Insufficient permissions",
		})
	}
}

// ErrorHandlingMiddleware handles panics and errors
func ErrorHandlingMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recovered: %v", err)
				c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Internal server error",
				})
			}
		}()

		return c.Next()
	}
}

// RequirePermission checks if user has specific permission(s)
// Pass permissions as (resource, action) pairs
// Example: RequirePermission(db, "requisition", "approve")
func RequirePermission(db *gorm.DB, requiredPermissions ...string) fiber.Handler {
	return func(c fiber.Ctx) error {
		// Get user info from context (set by AuthMiddleware)
		userID, ok := c.Locals("userID").(string)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "User ID not found in context",
			})
		}

		userRole, ok := c.Locals("userRole").(string)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "User role not found in context",
			})
		}

		// Get organization ID from context (set by TenantMiddleware)
		organizationID, ok := c.Locals("organizationID").(string)
		if !ok {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Organization ID not found in context",
			})
		}

		// Create permission service
		permissionService := services.NewPermissionService(db)

		// Check if we have pairs of (resource, action)
		if len(requiredPermissions)%2 != 0 {
			log.Printf("RequirePermission called with odd number of arguments")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Internal server error",
			})
		}

		// Check each required permission
		hasAllPermissions := true
		for i := 0; i < len(requiredPermissions); i += 2 {
			resource := requiredPermissions[i]
			action := requiredPermissions[i+1]

			if !permissionService.HasPermission(userID, organizationID, userRole, resource, action) {
				hasAllPermissions = false
				break
			}
		}

		if !hasAllPermissions {
			log.Printf("User %s with role %s denied access (missing permission)", userID, userRole)
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Insufficient permissions for this action",
			})
		}

		return c.Next()
	}
}

// RequirePermissionOr checks if user has ANY of the required permissions
// Pass permissions as (resource, action) pairs
// Example: RequirePermissionOr(db, "requisition", "approve", "budget", "approve")
func RequirePermissionOr(db *gorm.DB, requiredPermissions ...string) fiber.Handler {
	return func(c fiber.Ctx) error {
		// Get user info from context (set by AuthMiddleware)
		userID, ok := c.Locals("userID").(string)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "User ID not found in context",
			})
		}

		userRole, ok := c.Locals("userRole").(string)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "User role not found in context",
			})
		}

		// Get organization ID from context (set by TenantMiddleware)
		organizationID, ok := c.Locals("organizationID").(string)
		if !ok {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Organization ID not found in context",
			})
		}

		// Create permission service
		permissionService := services.NewPermissionService(db)

		// Check if we have pairs of (resource, action)
		if len(requiredPermissions)%2 != 0 {
			log.Printf("RequirePermissionOr called with odd number of arguments")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Internal server error",
			})
		}

		// Check if user has ANY of the required permissions
		hasAnyPermission := false
		for i := 0; i < len(requiredPermissions); i += 2 {
			resource := requiredPermissions[i]
			action := requiredPermissions[i+1]

			if permissionService.HasPermission(userID, organizationID, userRole, resource, action) {
				hasAnyPermission = true
				break
			}
		}

		if !hasAnyPermission {
			log.Printf("User %s with role %s denied access (missing all permissions)", userID, userRole)
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Insufficient permissions for this action",
			})
		}

		return c.Next()
	}
}
