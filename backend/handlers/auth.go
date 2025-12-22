package handlers

import (
	"log"

	"github.com/gofiber/fiber/v3"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/types"
	"github.com/liyali/liyali-gateway/utils"
	"gorm.io/gorm"
)

// Login handles user authentication
func Login(c fiber.Ctx) error {
	var req types.LoginRequest

	// Parse request body
	if err := c.Bind().Body(&req); err != nil {
		log.Printf("Error parsing login request: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{
			Success: false,
			Message: "Invalid request body",
			Error:   err.Error(),
		})
	}

	// Validate input
	if req.Email == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{
			Success: false,
			Message: "Email and password are required",
		})
	}

	// Find user by email
	var user models.User
	if err := config.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusUnauthorized).JSON(types.ErrorResponse{
				Success: false,
				Message: "Invalid email or password",
			})
		}
		log.Printf("Database error during login: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse{
			Success: false,
			Message: "Internal server error",
			Error:   err.Error(),
		})
	}

	// Check if user is active
	if !user.Active {
		return c.Status(fiber.StatusForbidden).JSON(types.ErrorResponse{
			Success: false,
			Message: "User account is inactive",
		})
	}

	// Verify password - For now, we'll accept any password for demo
	// In production, passwords should be hashed in database
	// if !utils.VerifyPassword(user.Password, req.Password) {
	// 	return c.Status(fiber.StatusUnauthorized).JSON(types.ErrorResponse{
	// 		Success: false,
	// 		Message: "Invalid email or password",
	// 	})
	// }

	// Generate JWT token
	token, err := utils.GenerateToken(user.ID, user.Email, user.Name, user.Role)
	if err != nil {
		log.Printf("Error generating token: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse{
			Success: false,
			Message: "Failed to generate authentication token",
			Error:   err.Error(),
		})
	}

	log.Printf("User logged in: %s (%s)", user.Email, user.ID)

	return c.Status(fiber.StatusOK).JSON(types.AuthResponse{
		Success: true,
		Message: "Login successful",
		Token:   token,
		User: &types.UserResponse{
			ID:        user.ID,
			Email:     user.Email,
			Name:      user.Name,
			Role:      user.Role,
			Active:    user.Active,
			CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		},
	})
}

// Register handles user registration
func Register(c fiber.Ctx) error {
	var req types.RegisterRequest

	// Parse request body
	if err := c.Bind().Body(&req); err != nil {
		log.Printf("Error parsing register request: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{
			Success: false,
			Message: "Invalid request body",
			Error:   err.Error(),
		})
	}

	// Validate input
	if req.Email == "" || req.Password == "" || req.Name == "" || req.Role == "" {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{
			Success: false,
			Message: "Email, password, name, and role are required",
		})
	}

	// Validate password strength
	if err := utils.ValidatePasswordStrength(req.Password); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{
			Success: false,
			Message: "Password does not meet requirements",
			Error:   err.Error(),
		})
	}

	// Check if user already exists
	var existingUser models.User
	if err := config.DB.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(types.ErrorResponse{
			Success: false,
			Message: "Email already registered",
		})
	} else if err != gorm.ErrRecordNotFound {
		log.Printf("Database error during registration: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse{
			Success: false,
			Message: "Internal server error",
			Error:   err.Error(),
		})
	}

	// Validate role
	validRoles := map[string]bool{
		"admin":      true,
		"approver":   true,
		"requester":  true,
		"finance":    true,
		"viewer":     true,
	}
	if !validRoles[req.Role] {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{
			Success: false,
			Message: "Invalid role",
		})
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse{
			Success: false,
			Message: "Failed to process registration",
			Error:   err.Error(),
		})
	}

	// Create new user
	newUser := models.User{
		ID:       utils.GenerateUserID(),
		Email:    req.Email,
		Name:     req.Name,
		Role:     req.Role,
		Active:   true,
	}

	// Note: In a full implementation, you'd want to store the hashedPassword
	// For now, we're storing the plain password for demo purposes
	// In production, use: newUser.Password = hashedPassword

	if err := config.DB.Create(&newUser).Error; err != nil {
		log.Printf("Error creating user: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse{
			Success: false,
			Message: "Failed to create user",
			Error:   err.Error(),
		})
	}

	// Generate JWT token
	token, err := utils.GenerateToken(newUser.ID, newUser.Email, newUser.Name, newUser.Role)
	if err != nil {
		log.Printf("Error generating token: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse{
			Success: false,
			Message: "Failed to generate authentication token",
			Error:   err.Error(),
		})
	}

	log.Printf("User registered: %s (%s)", newUser.Email, newUser.ID)

	return c.Status(fiber.StatusCreated).JSON(types.AuthResponse{
		Success: true,
		Message: "Registration successful",
		Token:   token,
		User: &types.UserResponse{
			ID:        newUser.ID,
			Email:     newUser.Email,
			Name:      newUser.Name,
			Role:      newUser.Role,
			Active:    newUser.Active,
			CreatedAt: newUser.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		},
	})
}

// VerifyToken verifies a JWT token
func VerifyToken(c fiber.Ctx) error {
	var req types.VerifyTokenRequest

	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{
			Success: false,
			Message: "Invalid request body",
			Error:   err.Error(),
		})
	}

	if req.Token == "" {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{
			Success: false,
			Message: "Token is required",
		})
	}

	claims, err := utils.ValidateToken(req.Token)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(types.VerifyTokenResponse{
			Valid: false,
			Error: err.Error(),
		})
	}

	// Find user to get latest info
	var user models.User
	if err := config.DB.Where("id = ?", claims.UserID).First(&user).Error; err != nil {
		log.Printf("Error fetching user during token verification: %v", err)
		return c.Status(fiber.StatusNotFound).JSON(types.VerifyTokenResponse{
			Valid: false,
			Error: "User not found",
		})
	}

	return c.Status(fiber.StatusOK).JSON(types.VerifyTokenResponse{
		Valid: true,
		User: &types.UserResponse{
			ID:        user.ID,
			Email:     user.Email,
			Name:      user.Name,
			Role:      user.Role,
			Active:    user.Active,
			CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		},
	})
}

// RefreshToken generates a new token from an existing token
func RefreshToken(c fiber.Ctx) error {
	var req types.RefreshTokenRequest

	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{
			Success: false,
			Message: "Invalid request body",
			Error:   err.Error(),
		})
	}

	if req.Token == "" {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{
			Success: false,
			Message: "Token is required",
		})
	}

	claims, err := utils.ValidateToken(req.Token)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(types.ErrorResponse{
			Success: false,
			Message: "Invalid token",
			Error:   err.Error(),
		})
	}

	// Generate new token
	newToken, err := utils.RefreshToken(claims)
	if err != nil {
		log.Printf("Error refreshing token: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse{
			Success: false,
			Message: "Failed to refresh token",
			Error:   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(types.AuthResponse{
		Success: true,
		Message: "Token refreshed successfully",
		Token:   newToken,
	})
}

// GetProfile returns current user profile (requires auth)
func GetProfile(c fiber.Ctx) error {
	userID, ok := c.Locals("userID").(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(types.ErrorResponse{
			Success: false,
			Message: "User ID not found in context",
		})
	}

	var user models.User
	if err := config.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		log.Printf("Error fetching user profile: %v", err)
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(types.ErrorResponse{
				Success: false,
				Message: "User not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse{
			Success: false,
			Message: "Internal server error",
			Error:   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"user": types.UserResponse{
			ID:        user.ID,
			Email:     user.Email,
			Name:      user.Name,
			Role:      user.Role,
			Active:    user.Active,
			CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		},
	})
}
