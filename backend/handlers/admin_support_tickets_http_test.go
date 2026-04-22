package handlers

import (
	"net/http"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/models"
	"github.com/stretchr/testify/assert"
)

func newAdminSupportTicketsApp() *fiber.App {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Use(withTenantCtx(testOrgID, testUserID, "admin"))
	app.Get("/admin/support/tickets", AdminListSupportTickets)
	app.Get("/admin/support/tickets/stats", AdminGetSupportTicketStats)
	app.Post("/admin/support/tickets", AdminCreateSupportTicket)
	app.Get("/admin/support/tickets/:id", AdminGetSupportTicket)
	app.Put("/admin/support/tickets/:id", AdminUpdateSupportTicket)
	return app
}

func TestAdminCreateSupportTicket_Success(t *testing.T) {
	db := setupRemainingAdminDB(t)
	defer teardownRemainingAdminDB(t, db)

	app := newAdminSupportTicketsApp()
	resp := testRequest(app, http.MethodPost, "/admin/support/tickets", map[string]interface{}{
		"subject":     "Cannot log in",
		"description": "Customer cannot access the portal",
		"priority":    "high",
		"category":    "access",
	})

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	body := decodeResponse(resp)
	assert.NotNil(t, body)
	assert.Equal(t, true, body["success"])
	data, ok := body["data"].(map[string]interface{})
	assert.True(t, ok)
	assert.Contains(t, data["ticket_number"].(string), "TKT-")
}

func TestAdminListSupportTickets_WithData(t *testing.T) {
	db := setupRemainingAdminDB(t)
	defer teardownRemainingAdminDB(t, db)

	now := time.Now()
	db.Exec(`INSERT INTO support_tickets (
		id, ticket_number, subject, description, status, priority, source, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		uuid.New().String(), "TKT-ABC12345", "Payment failed", "Card declined", "open", "medium", "manual", now, now)

	app := newAdminSupportTicketsApp()
	resp := testRequest(app, http.MethodGet, "/admin/support/tickets", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body := decodeResponse(resp)
	assert.NotNil(t, body)
	data := body["data"].(map[string]interface{})
	assert.EqualValues(t, 1, data["total"])
}

func TestAdminUpdateSupportTicket_Status(t *testing.T) {
	db := setupRemainingAdminDB(t)
	defer teardownRemainingAdminDB(t, db)

	ticketID := uuid.New().String()
	now := time.Now()
	db.Exec(`INSERT INTO support_tickets (
		id, ticket_number, subject, description, status, priority, source, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		ticketID, "TKT-XYZ98765", "Need help", "Manual ticket", "open", "medium", "manual", now, now)

	app := newAdminSupportTicketsApp()
	resp := testRequest(app, http.MethodPut, "/admin/support/tickets/"+ticketID, map[string]interface{}{
		"status":   "resolved",
		"priority": "high",
	})

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var ticket models.SupportTicket
	if err := db.First(&ticket, "id = ?", ticketID).Error; err != nil {
		t.Fatalf("failed to reload ticket: %v", err)
	}
	assert.Equal(t, "resolved", ticket.Status)
	assert.Equal(t, "high", ticket.Priority)
	assert.NotNil(t, ticket.ResolvedAt)
}

func TestAdminGetSupportTicketStats(t *testing.T) {
	db := setupRemainingAdminDB(t)
	defer teardownRemainingAdminDB(t, db)

	now := time.Now()
	db.Exec(`INSERT INTO support_tickets (
		id, ticket_number, subject, description, status, priority, source, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		uuid.New().String(), "TKT-STATS1", "Open issue", "Open issue", "open", "medium", "manual", now, now)
	db.Exec(`INSERT INTO support_tickets (
		id, ticket_number, subject, description, status, priority, source, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		uuid.New().String(), "TKT-STATS2", "Resolved issue", "Resolved issue", "resolved", "medium", "user_app", now, now)

	app := newAdminSupportTicketsApp()
	resp := testRequest(app, http.MethodGet, "/admin/support/tickets/stats", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body := decodeResponse(resp)
	assert.NotNil(t, body)
	data := body["data"].(map[string]interface{})
	assert.EqualValues(t, 2, data["total_tickets"])
	assert.EqualValues(t, 1, data["open_tickets"])
	assert.EqualValues(t, 1, data["resolved_tickets"])
}
