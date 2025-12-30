package handlers

import (
	"log"

	"github.com/liyali/liyali-gateway/services"
	"github.com/liyali/liyali-gateway/types"
	"github.com/liyali/liyali-gateway/utils"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	authService *services.AuthService
	rbacService *services.RBACService
	validate    *validator.Validate
}

func NewAuthHandler(authService *services.AuthService, rbacService *services.RBACService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		rbacService: rbacService,
		validate:    validator.New(),
	}
}

// GetAuthService returns the auth service instance
func (h *AuthHandler) GetAuthService() *services.AuthService {
	return h.authService
}

// Login handles user authentication with enhanced security
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req types.LoginRequest

	// Parse request body
	if err := c.BodyParser(&req); err != nil {
		log.Printf("Error parsing login request: %v", err)
		return utils.SendBadRequestError(c, "Failed to parse login request")
	}

	// Validate request
	if err := h.validate.Struct(req); err != nil {
		return utils.SendValidationError(c, err.Error())
	}

	// Get client info
	ipAddress := c.IP()
	userAgent := c.Get("User-Agent")

	// Attempt login
	result, err := h.authService.Login(c.Context(), req.Email, req.Password, ipAddress, userAgent)
	if err != nil {
		log.Printf("Login failed for email %s: %v", req.Email, err)
		
		// Return generic error for security
		return utils.SendUnauthorizedError(c, "Invalid email or password")
	}

	return utils.SendSimpleSuccess(c, result, "Login successful")
}

// RefreshToken handles token refresh with enhanced security
func (h *AuthHandler) RefreshToken(c *fiber.Ctx) error {
	var req types.RefreshTokenRequest

	if err := c.BodyParser(&req); err != nil {
		return utils.SendBadRequestError(c, "Failed to parse refresh token request")
	}

	// Validate request
	if err := h.validate.Struct(req); err != nil {
		return utils.SendValidationError(c, err.Error())
	}

	// Refresh token
	result, err := h.authService.RefreshToken(c.Context(), req.RefreshToken)
	if err != nil {
		log.Printf("Token refresh failed: %v", err)
		return utils.SendUnauthorizedError(c, "Invalid or expired refresh token")
	}

	return utils.SendSimpleSuccess(c, result, "Token refreshed successfully")
}

// Logout handles user logout with session cleanup
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	var req types.RefreshTokenRequest

	if err := c.BodyParser(&req); err != nil {
		return utils.SendBadRequestError(c, "Failed to parse logout request")
	}

	// Validate request
	if err := h.validate.Struct(req); err != nil {
		return utils.SendValidationError(c, err.Error())
	}

	// Logout
	if err := h.authService.Logout(c.Context(), req.RefreshToken); err != nil {
		log.Printf("Logout failed: %v", err)
		return utils.SendInternalError(c, "Failed to invalidate session", err)
	}

	return utils.SendSimpleSuccess(c, nil, "Logged out successfully")
}

// LogoutAll handles logout from all devices
func (h *AuthHandler) LogoutAll(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(string)
	if !ok {
		return utils.SendUnauthorizedError(c, "User not authenticated")
	}

	// Logout from all devices
	if err := h.authService.LogoutAll(c.Context(), userID); err != nil {
		log.Printf("Logout all failed for user %s: %v", userID, err)
		return utils.SendInternalError(c, "Failed to invalidate all sessions", err)
	}

	return utils.SendSimpleSuccess(c, nil, "Logged out from all devices successfully")
}

// RequestPasswordReset handles password reset requests
func (h *AuthHandler) RequestPasswordReset(c *fiber.Ctx) error {
	var req types.PasswordResetRequest

	if err := c.BodyParser(&req); err != nil {
		return utils.SendBadRequestError(c, "Failed to parse password reset request")
	}

	// Validate request
	if err := h.validate.Struct(req); err != nil {
		return utils.SendValidationError(c, err.Error())
	}

	// Create password reset token
	token, err := h.authService.CreatePasswordReset(c.Context(), req.Email)
	if err != nil {
		log.Printf("Password reset request failed for email %s: %v", req.Email, err)
		// Don't reveal if user exists or not for security
	}

	// Always return success for security (don't reveal if email exists)
	return utils.SendSimpleSuccess(c, map[string]interface{}{
		"token": token, // TODO: Remove in production, send via email instead
	}, "If the email exists, a password reset link has been sent")
}

// ResetPassword handles password reset with token
func (h *AuthHandler) ResetPassword(c *fiber.Ctx) error {
	var req types.ResetPasswordRequest

	if err := c.BodyParser(&req); err != nil {
		return utils.SendBadRequestError(c, "Failed to parse reset password request")
	}

	// Validate request
	if err := h.validate.Struct(req); err != nil {
		return utils.SendValidationError(c, err.Error())
	}

	// Reset password
	if err := h.authService.ResetPassword(c.Context(), req.Token, req.NewPassword); err != nil {
		log.Printf("Password reset failed: %v", err)
		return utils.SendBadRequestError(c, "Invalid or expired reset token")
	}

	return utils.SendSimpleSuccess(c, nil, "Password reset successfully")
}

// ChangePassword handles password change (requires current password)
func (h *AuthHandler) ChangePassword(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(string)
	if !ok {
		return utils.SendUnauthorizedError(c, "User not authenticated")
	}

	var req types.ChangePasswordRequest

	if err := c.BodyParser(&req); err != nil {
		return utils.SendBadRequestError(c, "Failed to parse change password request")
	}

	// Validate request
	if err := h.validate.Struct(req); err != nil {
		return utils.SendValidationError(c, err.Error())
	}

	// Change password
	if err := h.authService.ChangePassword(c.Context(), userID, req.CurrentPassword, req.NewPassword); err != nil {
		log.Printf("Password change failed for user %s: %v", userID, err)
		return utils.SendBadRequestError(c, "Current password is incorrect")
	}

	return utils.SendSimpleSuccess(c, nil, "Password changed successfully")
}

// Register handles user registration
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req types.RegisterRequest

	if err := c.BodyParser(&req); err != nil {
		return utils.SendBadRequestError(c, "Failed to parse registration request")
	}

	// Validate request
	if err := h.validate.Struct(req); err != nil {
		return utils.SendValidationError(c, err.Error())
	}

	// Get client info
	// ipAddress := c.IP()
	// userAgent := c.Get("User-Agent")

	// Register user
	response, err := h.authService.Register(c.Context(), req.Email, req.Password, req.Name, req.Role)
	if err != nil {
		log.Printf("Registration failed for email %s: %v", req.Email, err)
		
		// Handle specific errors
		switch err {
		case services.ErrEmailAlreadyExists:
			return utils.SendConflictError(c, "Email already exists")
		default:
			return utils.SendInternalError(c, "Registration failed", err)
		}
	}

	// Return success response in the format expected by frontend
	return utils.SendCreatedSuccess(c, map[string]interface{}{
		"token":        response.AccessToken, // Frontend expects "token" field
		"user":         response.User,
		"organization": response.Organization,
	}, "User registered successfully")
}

// VerifyToken verifies a JWT token
func (h *AuthHandler) VerifyToken(c *fiber.Ctx) error {
	var req types.VerifyTokenRequest

	if err := c.BodyParser(&req); err != nil {
		return utils.SendBadRequestError(c, "Failed to parse token verification request")
	}

	// Validate request
	if err := h.validate.Struct(req); err != nil {
		return utils.SendValidationError(c, err.Error())
	}

	// Verify token using the auth service
	claims, err := h.authService.ValidateAccessToken(req.Token)
	if err != nil {
		return utils.SendUnauthorizedError(c, "Invalid or expired token")
	}

	return utils.SendSimpleSuccess(c, map[string]interface{}{
		"user_id":         claims.UserID,
		"email":           claims.Email,
		"role":            claims.Role,
		"organization_id": claims.OrganizationID,
		"expires_at":      claims.ExpiresAt,
	}, "Token is valid")
}

// GetProfile returns current user profile (requires auth)
func (h *AuthHandler) GetProfile(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(string)
	if !ok {
		return utils.SendUnauthorizedError(c, "User not authenticated")
	}

	// For now, return a simple response with the user ID
	// TODO: Implement GetUserProfile method in the auth service
	return utils.SendSimpleSuccess(c, map[string]interface{}{
		"id": userID,
	}, "Profile retrieved successfully")
}