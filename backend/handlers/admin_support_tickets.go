package handlers

import (
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/utils"
	"gorm.io/gorm"
)

var (
	validSupportTicketStatuses = map[string]struct{}{
		"open":                {},
		"in_progress":         {},
		"waiting_on_customer": {},
		"resolved":            {},
		"closed":              {},
	}
	validSupportTicketPriorities = map[string]struct{}{
		"low":    {},
		"medium": {},
		"high":   {},
		"urgent": {},
	}
)

type supportTicketCreateRequest struct {
	OrganizationID    *string `json:"organization_id"`
	UserID            *string `json:"user_id"`
	AssignedToAdminID *string `json:"assigned_to_admin_id"`
	Subject           string  `json:"subject"`
	Description       string  `json:"description"`
	Category          string  `json:"category"`
	Priority          string  `json:"priority"`
	ExternalReference string  `json:"external_reference"`
	InternalNotes     string  `json:"internal_notes"`
}

type supportTicketUpdateRequest struct {
	OrganizationID    *string `json:"organization_id"`
	UserID            *string `json:"user_id"`
	AssignedToAdminID *string `json:"assigned_to_admin_id"`
	Subject           *string `json:"subject"`
	Description       *string `json:"description"`
	Category          *string `json:"category"`
	Priority          *string `json:"priority"`
	Status            *string `json:"status"`
	ExternalReference *string `json:"external_reference"`
	InternalNotes     *string `json:"internal_notes"`
	ResolutionSummary *string `json:"resolution_summary"`
}

func applySupportTicketFilters(query *gorm.DB, c *fiber.Ctx) *gorm.DB {
	if orgID := strings.TrimSpace(c.Query("organization_id")); orgID != "" {
		query = query.Where("organization_id = ?", orgID)
	}
	if userID := strings.TrimSpace(c.Query("user_id")); userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	if status := strings.TrimSpace(c.Query("status")); status != "" && status != "all" {
		query = query.Where("status = ?", strings.ToLower(status))
	}
	if priority := strings.TrimSpace(c.Query("priority")); priority != "" && priority != "all" {
		query = query.Where("priority = ?", strings.ToLower(priority))
	}
	if source := strings.TrimSpace(c.Query("source")); source != "" && source != "all" {
		query = query.Where("source = ?", strings.ToLower(source))
	}
	if search := strings.TrimSpace(c.Query("search")); search != "" {
		pattern := "%" + search + "%"
		query = query.Where(
			"ticket_number ILIKE ? OR subject ILIKE ? OR description ILIKE ? OR COALESCE(external_reference, '') ILIKE ?",
			pattern, pattern, pattern, pattern,
		)
	}

	return query
}

func isValidSupportTicketStatus(status string) bool {
	_, ok := validSupportTicketStatuses[strings.ToLower(strings.TrimSpace(status))]
	return ok
}

func isValidSupportTicketPriority(priority string) bool {
	_, ok := validSupportTicketPriorities[strings.ToLower(strings.TrimSpace(priority))]
	return ok
}

func validateUserExists(db *gorm.DB, userID string) error {
	if userID == "" {
		return nil
	}

	var count int64
	if err := db.Table("users").
		Where("id = ? AND deleted_at IS NULL", userID).
		Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

func validateOrganizationExists(db *gorm.DB, orgID string) error {
	if orgID == "" {
		return nil
	}

	var count int64
	if err := db.Table("organizations").
		Where("id = ? AND deleted_at IS NULL", orgID).
		Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("organization not found")
	}

	return nil
}

// AdminListSupportTickets returns support tickets for the admin console.
func AdminListSupportTickets(c *fiber.Ctx) error {
	db := config.DB

	baseQuery := applySupportTicketFilters(db.Model(&models.SupportTicket{}), c)

	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return utils.SendInternalError(c, "Failed to fetch support tickets", err)
	}

	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)
	page, limit = utils.NormalizePaginationParams(page, limit)
	offset := (page - 1) * limit

	var tickets []models.SupportTicket
	if err := applySupportTicketFilters(db.Model(&models.SupportTicket{}), c).
		Preload("Organization").
		Preload("User").
		Preload("CreatedByAdmin").
		Preload("AssignedToAdmin").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&tickets).Error; err != nil {
		return utils.SendInternalError(c, "Failed to fetch support tickets", err)
	}

	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Support tickets retrieved successfully",
		"data": fiber.Map{
			"tickets":     tickets,
			"total":       total,
			"page":        page,
			"limit":       limit,
			"totalPages":  totalPages,
			"hasNext":     page < totalPages,
			"hasPrevious": page > 1,
		},
	})
}

// AdminGetSupportTicketStats returns summary statistics for support tickets.
func AdminGetSupportTicketStats(c *fiber.Ctx) error {
	db := config.DB

	countWhere := func(queryFn func(*gorm.DB) *gorm.DB) int64 {
		var count int64
		query := queryFn(db.Model(&models.SupportTicket{}))
		_ = query.Count(&count).Error
		return count
	}

	total := countWhere(func(query *gorm.DB) *gorm.DB { return query })
	open := countWhere(func(query *gorm.DB) *gorm.DB { return query.Where("status = ?", "open") })
	inProgress := countWhere(func(query *gorm.DB) *gorm.DB { return query.Where("status = ?", "in_progress") })
	waiting := countWhere(func(query *gorm.DB) *gorm.DB { return query.Where("status = ?", "waiting_on_customer") })
	resolved := countWhere(func(query *gorm.DB) *gorm.DB { return query.Where("status = ?", "resolved") })
	closed := countWhere(func(query *gorm.DB) *gorm.DB { return query.Where("status = ?", "closed") })
	manual := countWhere(func(query *gorm.DB) *gorm.DB { return query.Where("source = ?", "manual") })
	userApp := countWhere(func(query *gorm.DB) *gorm.DB { return query.Where("source = ?", "user_app") })
	email := countWhere(func(query *gorm.DB) *gorm.DB { return query.Where("source = ?", "email") })
	overdue := countWhere(func(query *gorm.DB) *gorm.DB {
		return query.Where(
			"status IN ? AND created_at < ?",
			[]string{"open", "in_progress", "waiting_on_customer"},
			time.Now().Add(-72*time.Hour),
		)
	})

	return utils.SendSimpleSuccess(c, fiber.Map{
		"total_tickets":       total,
		"open_tickets":        open,
		"in_progress_tickets": inProgress,
		"waiting_tickets":     waiting,
		"resolved_tickets":    resolved,
		"closed_tickets":      closed,
		"manual_tickets":      manual,
		"user_app_tickets":    userApp,
		"email_tickets":       email,
		"overdue_tickets":     overdue,
	}, "Support ticket statistics retrieved successfully")
}

// AdminGetSupportTicket returns a single ticket with related context.
func AdminGetSupportTicket(c *fiber.Ctx) error {
	db := config.DB
	ticketID := c.Params("id")

	var ticket models.SupportTicket
	if err := db.
		Preload("Organization").
		Preload("User").
		Preload("CreatedByAdmin").
		Preload("AssignedToAdmin").
		First(&ticket, "id = ?", ticketID).Error; err != nil {
		return utils.SendNotFound(c, "Support ticket not found")
	}

	return utils.SendSimpleSuccess(c, ticket, "Support ticket retrieved successfully")
}

// AdminCreateSupportTicket creates a manual support ticket from the admin console.
func AdminCreateSupportTicket(c *fiber.Ctx) error {
	db := config.DB
	callerID, _ := c.Locals("userID").(string)

	var req supportTicketCreateRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendBadRequest(c, "Invalid request body")
	}

	subject := strings.TrimSpace(req.Subject)
	if subject == "" {
		return utils.SendBadRequest(c, "subject is required")
	}
	description := strings.TrimSpace(req.Description)
	if description == "" {
		return utils.SendBadRequest(c, "description is required")
	}

	category := strings.ToLower(strings.TrimSpace(req.Category))
	if category == "" {
		category = "general"
	}

	priority := strings.ToLower(strings.TrimSpace(req.Priority))
	if priority == "" {
		priority = "medium"
	}
	if !isValidSupportTicketPriority(priority) {
		return utils.SendBadRequest(c, "priority must be one of low, medium, high, or urgent")
	}

	if req.OrganizationID != nil {
		orgID := strings.TrimSpace(*req.OrganizationID)
		if err := validateOrganizationExists(db, orgID); err != nil {
			return utils.SendNotFound(c, "Organization not found")
		}
	}

	if req.UserID != nil {
		userID := strings.TrimSpace(*req.UserID)
		if err := validateUserExists(db, userID); err != nil {
			return utils.SendNotFound(c, "User not found")
		}
	}

	if req.AssignedToAdminID != nil {
		adminID := strings.TrimSpace(*req.AssignedToAdminID)
		if err := validateUserExists(db, adminID); err != nil {
			return utils.SendNotFound(c, "Assigned admin user not found")
		}
	}

	now := time.Now()
	ticket := models.SupportTicket{
		ID:                utils.GenerateID(),
		TicketNumber:      fmt.Sprintf("TKT-%s", strings.ToUpper(utils.GenerateID()[:8])),
		Source:            "manual",
		Category:          category,
		Priority:          priority,
		Status:            "open",
		Subject:           subject,
		Description:       description,
		ExternalReference: strings.TrimSpace(req.ExternalReference),
		InternalNotes:     strings.TrimSpace(req.InternalNotes),
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	if req.OrganizationID != nil && strings.TrimSpace(*req.OrganizationID) != "" {
		orgID := strings.TrimSpace(*req.OrganizationID)
		ticket.OrganizationID = &orgID
	}
	if req.UserID != nil && strings.TrimSpace(*req.UserID) != "" {
		userID := strings.TrimSpace(*req.UserID)
		ticket.UserID = &userID
	}
	if callerID != "" {
		ticket.CreatedByAdminID = &callerID
	}
	if req.AssignedToAdminID != nil && strings.TrimSpace(*req.AssignedToAdminID) != "" {
		adminID := strings.TrimSpace(*req.AssignedToAdminID)
		ticket.AssignedToAdminID = &adminID
	}

	if err := db.Create(&ticket).Error; err != nil {
		return utils.SendInternalError(c, "Failed to create support ticket", err)
	}

	var created models.SupportTicket
	if err := db.
		Preload("Organization").
		Preload("User").
		Preload("CreatedByAdmin").
		Preload("AssignedToAdmin").
		First(&created, "id = ?", ticket.ID).Error; err != nil {
		return utils.SendInternalError(c, "Support ticket created but failed to reload", err)
	}

	return utils.SendCreatedSuccess(c, created, "Support ticket created successfully")
}

// AdminUpdateSupportTicket updates an existing support ticket.
func AdminUpdateSupportTicket(c *fiber.Ctx) error {
	db := config.DB
	ticketID := c.Params("id")

	var req supportTicketUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendBadRequest(c, "Invalid request body")
	}

	var ticket models.SupportTicket
	if err := db.First(&ticket, "id = ?", ticketID).Error; err != nil {
		return utils.SendNotFound(c, "Support ticket not found")
	}

	updates := map[string]interface{}{
		"updated_at": time.Now(),
	}

	if req.OrganizationID != nil {
		orgID := strings.TrimSpace(*req.OrganizationID)
		if orgID != "" {
			if err := validateOrganizationExists(db, orgID); err != nil {
				return utils.SendNotFound(c, "Organization not found")
			}
			updates["organization_id"] = orgID
		} else {
			updates["organization_id"] = nil
		}
	}
	if req.UserID != nil {
		userID := strings.TrimSpace(*req.UserID)
		if userID != "" {
			if err := validateUserExists(db, userID); err != nil {
				return utils.SendNotFound(c, "User not found")
			}
			updates["user_id"] = userID
		} else {
			updates["user_id"] = nil
		}
	}
	if req.AssignedToAdminID != nil {
		adminID := strings.TrimSpace(*req.AssignedToAdminID)
		if adminID != "" {
			if err := validateUserExists(db, adminID); err != nil {
				return utils.SendNotFound(c, "Assigned admin user not found")
			}
			updates["assigned_to_admin_id"] = adminID
		} else {
			updates["assigned_to_admin_id"] = nil
		}
	}
	if req.Subject != nil {
		subject := strings.TrimSpace(*req.Subject)
		if subject == "" {
			return utils.SendBadRequest(c, "subject cannot be empty")
		}
		updates["subject"] = subject
	}
	if req.Description != nil {
		description := strings.TrimSpace(*req.Description)
		if description == "" {
			return utils.SendBadRequest(c, "description cannot be empty")
		}
		updates["description"] = description
	}
	if req.Category != nil {
		category := strings.ToLower(strings.TrimSpace(*req.Category))
		if category == "" {
			category = "general"
		}
		updates["category"] = category
	}
	if req.Priority != nil {
		priority := strings.ToLower(strings.TrimSpace(*req.Priority))
		if !isValidSupportTicketPriority(priority) {
			return utils.SendBadRequest(c, "priority must be one of low, medium, high, or urgent")
		}
		updates["priority"] = priority
	}
	if req.Status != nil {
		status := strings.ToLower(strings.TrimSpace(*req.Status))
		if !isValidSupportTicketStatus(status) {
			return utils.SendBadRequest(c, "status must be one of open, in_progress, waiting_on_customer, resolved, or closed")
		}
		updates["status"] = status
		if status == "resolved" || status == "closed" {
			now := time.Now()
			if status == "resolved" {
				updates["resolved_at"] = now
			}
			if status == "closed" {
				updates["closed_at"] = now
			}
		} else {
			updates["resolved_at"] = nil
			updates["closed_at"] = nil
		}
	}
	if req.ExternalReference != nil {
		updates["external_reference"] = strings.TrimSpace(*req.ExternalReference)
	}
	if req.InternalNotes != nil {
		updates["internal_notes"] = strings.TrimSpace(*req.InternalNotes)
	}
	if req.ResolutionSummary != nil {
		updates["resolution_summary"] = strings.TrimSpace(*req.ResolutionSummary)
	}

	if err := db.Model(&ticket).Where("id = ?", ticketID).Updates(updates).Error; err != nil {
		return utils.SendInternalError(c, "Failed to update support ticket", err)
	}

	var updated models.SupportTicket
	if err := db.
		Preload("Organization").
		Preload("User").
		Preload("CreatedByAdmin").
		Preload("AssignedToAdmin").
		First(&updated, "id = ?", ticketID).Error; err != nil {
		return utils.SendInternalError(c, "Support ticket updated but failed to reload", err)
	}

	return utils.SendSimpleSuccess(c, updated, "Support ticket updated successfully")
}
