package middleware

import (
	"strings"

	"github.com/cozyCodr/liyali-gateway/internal/services"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type AuthMiddleware struct {
	authService *services.AuthService
}

func NewAuthMiddleware(authService *services.AuthService) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
	}
}

// Authenticate validates JWT token and adds user info to context
func (m *AuthMiddleware) Authenticate(c fiber.Ctx) error {
	// Get Authorization header
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "missing authorization header",
		})
	}

	// Check if it's a Bearer token
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid authorization header format",
		})
	}

	tokenString := parts[1]

	// Validate token
	claims, err := m.authService.ValidateAccessToken(tokenString)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid or expired token",
		})
	}

	// Set user information in context
	c.Locals("userID", claims.UserID)
	c.Locals("userEmail", claims.Email)
	c.Locals("userRole", claims.Role)

	return c.Next()
}

// GetUserID retrieves user ID from context
func GetUserID(c fiber.Ctx) (uuid.UUID, bool) {
	userID, ok := c.Locals("userID").(uuid.UUID)
	return userID, ok
}

// GetUserEmail retrieves user email from context
func GetUserEmail(c fiber.Ctx) (string, bool) {
	email, ok := c.Locals("userEmail").(string)
	return email, ok
}

// GetUserRole retrieves user role from context
func GetUserRole(c fiber.Ctx) (string, bool) {
	role, ok := c.Locals("userRole").(string)
	return role, ok
}
