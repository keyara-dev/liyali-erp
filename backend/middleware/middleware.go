package middleware

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
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
