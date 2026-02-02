package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/liyali/liyali-gateway/logging"
	"github.com/liyali/liyali-gateway/services"
	"github.com/liyali/liyali-gateway/utils"
)

// SubscriptionHandler handles subscription-related HTTP requests
type SubscriptionHandler struct {
	subscriptionService *services.SubscriptionService
}

// NewSubscriptionHandler creates a new subscription handler
func NewSubscriptionHandler(subscriptionService *services.SubscriptionService, logger *logging.Logger) *SubscriptionHandler {
	return &SubscriptionHandler{
		subscriptionService: subscriptionService,
	}
}

// GetSubscriptionPlans returns all available subscription plans
func (h *SubscriptionHandler) GetSubscriptionPlans(c *fiber.Ctx) error {
	logger := logging.FromContext(c)
	
	logger.Info("Getting subscription plans")

	plans, err := h.subscriptionService.GetAllSubscriptionPlans()
	if err != nil {
		logger.Error("Failed to get subscription plans")
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse("Failed to retrieve subscription plans"))
	}

	response := map[string]interface{}{
		"plans": plans,
	}

	logger.Info("Subscription plans retrieved successfully")

	return c.JSON(utils.SuccessResponse(response, "Subscription plans retrieved successfully", nil))
}

// GetOrganizationTrialStatus returns trial status for an organization
func (h *SubscriptionHandler) GetOrganizationTrialStatus(c *fiber.Ctx) error {
	logger := logging.FromContext(c)
	
	organizationID := c.Params("id")
	if organizationID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse("Organization ID is required"))
	}

	logger.Info("Getting organization trial status")

	trialStatus, err := h.subscriptionService.GetOrganizationTrialStatus(organizationID)
	if err != nil {
		logger.Error("Failed to get organization trial status")
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse("Failed to retrieve trial status"))
	}

	logger.Info("Organization trial status retrieved successfully")

	return c.JSON(utils.SuccessResponse(trialStatus, "Trial status retrieved successfully", nil))
}

// CheckFeatureAccess checks if an organization has access to a specific feature
func (h *SubscriptionHandler) CheckFeatureAccess(c *fiber.Ctx) error {
	logger := logging.FromContext(c)
	
	organizationID := c.Params("id")
	featureName := c.Query("feature")

	if organizationID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse("Organization ID is required"))
	}

	if featureName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse("Feature name is required"))
	}

	logger.Info("Checking feature access")

	result, err := h.subscriptionService.CheckFeatureAccess(organizationID, featureName)
	if err != nil {
		logger.Error("Failed to check feature access")
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse("Failed to check feature access"))
	}

	logger.Info("Feature access checked successfully")

	return c.JSON(utils.SuccessResponse(result, "Feature access checked successfully", nil))
}

// UpgradeOrganization handles organization upgrade requests
func (h *SubscriptionHandler) UpgradeOrganization(c *fiber.Ctx) error {
	logger := logging.FromContext(c)
	
	organizationID := c.Params("id")
	if organizationID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse("Organization ID is required"))
	}

	var request map[string]interface{}
	if err := c.BodyParser(&request); err != nil {
		logger.Error("Failed to parse upgrade request")
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse("Invalid request body"))
	}

	logger.Info("Processing organization upgrade")

	response, err := h.subscriptionService.UpgradeOrganization(organizationID, request)
	if err != nil {
		logger.Error("Failed to process upgrade")
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse("Failed to process upgrade request"))
	}

	logger.Info("Organization upgrade processed successfully")

	return c.JSON(utils.SuccessResponse(response, "Upgrade request processed successfully", nil))
}

// GetOrganizationSubscription returns comprehensive subscription details
func (h *SubscriptionHandler) GetOrganizationSubscription(c *fiber.Ctx) error {
	logger := logging.FromContext(c)
	
	organizationID := c.Params("id")
	if organizationID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse("Organization ID is required"))
	}

	logger.Info("Getting organization subscription details")

	details, err := h.subscriptionService.GetOrganizationSubscriptionDetails(organizationID)
	if err != nil {
		logger.Error("Failed to get organization subscription details")
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse("Failed to retrieve subscription details"))
	}

	logger.Info("Organization subscription details retrieved successfully")

	return c.JSON(utils.SuccessResponse(details, "Subscription details retrieved successfully", nil))
}

// ExtendOrganizationTrial extends trial period (admin only)
func (h *SubscriptionHandler) ExtendOrganizationTrial(c *fiber.Ctx) error {
	logger := logging.FromContext(c)
	
	organizationID := c.Params("id")
	if organizationID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse("Organization ID is required"))
	}

	var request struct {
		DaysToAdd int    `json:"daysToAdd" validate:"required,min=1,max=30"`
		Reason    string `json:"reason" validate:"required,min=5,max=200"`
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse("Invalid request body"))
	}

	// Get user info from context (set by auth middleware)
	userID := c.Locals("user_id").(string)
	
	logger.Info("Extending organization trial")

	err := h.subscriptionService.ExtendOrganizationTrial(organizationID, request.DaysToAdd, request.Reason, userID)
	if err != nil {
		logger.Error("Failed to extend organization trial")
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse("Failed to extend trial"))
	}

	// Get updated trial status
	trialStatus, err := h.subscriptionService.GetOrganizationTrialStatus(organizationID)
	if err != nil {
		logger.Warn("Failed to get updated trial status after extension")
		// Don't fail the request, just return success without updated status
		return c.JSON(utils.SuccessResponse(map[string]interface{}{
			"message": "Trial extended successfully",
		}, "Trial extended successfully", nil))
	}

	logger.Info("Organization trial extended successfully")

	return c.JSON(utils.SuccessResponse(trialStatus, "Trial extended successfully", nil))
}