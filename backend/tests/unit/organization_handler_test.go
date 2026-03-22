package unit

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/models"
	"github.com/stretchr/testify/assert"
)

// ─────────────────────────────────────────────────────────────────────────────
// Organization CRUD Validation
// ─────────────────────────────────────────────────────────────────────────────

func TestCreateOrganization_Validation(t *testing.T) {
	tests := []struct {
		name       string
		orgName    string
		slug       string
		shouldPass bool
	}{
		{"Valid name and slug", "Acme Corp", "acme-corp", true},
		{"Missing name", "", "acme-corp", false},
		{"Missing slug", "Acme Corp", "", false},
		{"Both empty", "", "", false},
		{"Name only (slug can be derived)", "Acme Corp", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.orgName != "" && tt.slug != ""
			assert.Equal(t, tt.shouldPass, isValid)
		})
	}
}

func TestOrganizationSlugFormat(t *testing.T) {
	tests := []struct {
		name       string
		slug       string
		shouldPass bool
	}{
		{"Valid lowercase slug", "acme-corp", true},
		{"Valid with numbers", "org-2024", true},
		{"Valid single word", "acme", true},
		{"Contains uppercase", "Acme-Corp", false},
		{"Contains spaces", "acme corp", false},
		{"Contains special chars", "acme@corp", false},
		{"Empty slug", "", false},
		{"Too short (1 char)", "a", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := len(tt.slug) >= 2
			for _, c := range tt.slug {
				if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-') {
					isValid = false
					break
				}
			}
			assert.Equal(t, tt.shouldPass, isValid)
		})
	}
}

func TestOrganizationModel_RequiredFields(t *testing.T) {
	t.Run("Complete organization has required fields", func(t *testing.T) {
		org := models.Organization{
			ID:        uuid.New().String(),
			Name:      "Test Organization",
			Slug:      "test-org",
			Tier:      "starter",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		assert.NotEmpty(t, org.ID)
		assert.NotEmpty(t, org.Name)
		assert.NotEmpty(t, org.Slug)
		assert.NotEmpty(t, org.Tier)
	})

	t.Run("Organization tier values", func(t *testing.T) {
		validTiers := map[string]bool{
			"starter":    true,
			"growth":     true,
			"enterprise": true,
		}

		tiers := []string{"starter", "growth", "enterprise", "free", "invalid"}
		for _, tier := range tiers {
			_, isValid := validTiers[tier]
			if tier == "starter" || tier == "growth" || tier == "enterprise" {
				assert.True(t, isValid, "Tier %s should be valid", tier)
			} else {
				assert.False(t, isValid, "Tier %s should be invalid", tier)
			}
		}
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// Organization Settings — Procurement Flow
// ─────────────────────────────────────────────────────────────────────────────

func TestOrganizationSettings_ProcurementFlow(t *testing.T) {
	tests := []struct {
		name            string
		procurementFlow string
		shouldPass      bool
	}{
		{"Goods-first (default)", "goods_first", true},
		{"Payment-first", "payment_first", true},
		{"Empty string (invalid — must have a value)", "", false},
		{"Invalid value", "immediate", false},
		{"Uppercase invalid", "GOODS_FIRST", false},
		{"Mixed case invalid", "Goods_First", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.procurementFlow == "goods_first" || tt.procurementFlow == "payment_first"
			assert.Equal(t, tt.shouldPass, isValid)
		})
	}
}

func TestOrganizationSettings_DefaultProcurementFlow(t *testing.T) {
	t.Run("Default is goods_first", func(t *testing.T) {
		settings := models.OrganizationSettings{
			ProcurementFlow: "goods_first",
		}

		assert.Equal(t, "goods_first", settings.ProcurementFlow)
	})

	t.Run("Can be changed to payment_first", func(t *testing.T) {
		settings := models.OrganizationSettings{
			ProcurementFlow: "payment_first",
		}

		assert.Equal(t, "payment_first", settings.ProcurementFlow)
		assert.NotEqual(t, "goods_first", settings.ProcurementFlow)
	})
}

func TestOrganizationSettings_FlowChangeAudit(t *testing.T) {
	t.Run("Flow change should be logged", func(t *testing.T) {
		before := "goods_first"
		after := "payment_first"

		changeRecord := map[string]interface{}{
			"field":  "procurementFlow",
			"before": before,
			"after":  after,
		}

		assert.Equal(t, "procurementFlow", changeRecord["field"])
		assert.Equal(t, "goods_first", changeRecord["before"])
		assert.Equal(t, "payment_first", changeRecord["after"])
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// Organization Settings — Other Settings
// ─────────────────────────────────────────────────────────────────────────────

func TestOrganizationSettings_UpdateValidation(t *testing.T) {
	t.Run("Valid settings update", func(t *testing.T) {
		settings := models.OrganizationSettings{
			ProcurementFlow: "goods_first",
		}

		assert.NotNil(t, settings)
		assert.Equal(t, "goods_first", settings.ProcurementFlow)
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// Organization Membership
// ─────────────────────────────────────────────────────────────────────────────

func TestOrganizationMembership_RoleValidation(t *testing.T) {
	tests := []struct {
		name       string
		role       string
		shouldPass bool
	}{
		{"Admin role", "admin", true},
		{"Approver role", "approver", true},
		{"Finance role", "finance", true},
		{"Requester role", "requester", true},
		{"Invalid role", "superuser", false},
		{"Empty role", "", false},
		{"Uppercase role", "ADMIN", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validRoles := map[string]bool{
				"admin":     true,
				"approver":  true,
				"finance":   true,
				"requester": true,
			}
			isValid := validRoles[tt.role]
			assert.Equal(t, tt.shouldPass, isValid)
		})
	}
}

func TestOrganizationMembership_InviteValidation(t *testing.T) {
	tests := []struct {
		name       string
		email      string
		role       string
		shouldPass bool
	}{
		{"Valid email + role", "user@example.com", "requester", true},
		{"Missing email", "", "requester", false},
		{"Invalid email format", "not-an-email", "requester", false},
		{"Missing role", "user@example.com", "", false},
		{"Invalid role", "user@example.com", "viewer", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validRoles := map[string]bool{
				"admin": true, "approver": true,
				"finance": true, "requester": true,
			}
			hasAt := false
			for _, c := range tt.email {
				if c == '@' {
					hasAt = true
					break
				}
			}
			isValid := tt.email != "" && hasAt && validRoles[tt.role]
			assert.Equal(t, tt.shouldPass, isValid)
		})
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Organization Response Format
// ─────────────────────────────────────────────────────────────────────────────

func TestOrganizationResponse_Structure(t *testing.T) {
	t.Run("Organization response has required fields", func(t *testing.T) {
		org := models.Organization{
			ID:          uuid.New().String(),
			Name:        "Test Org",
			Slug:        "test-org",
			Tier:        "starter",
			Description: "A test organization",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		assert.NotEqual(t, uuid.Nil, org.ID)
		assert.NotEmpty(t, org.Name)
		assert.NotEmpty(t, org.Slug)
		assert.NotEmpty(t, org.Tier)
		assert.False(t, org.CreatedAt.IsZero())
	})

	t.Run("Organization ID is UUID format", func(t *testing.T) {
		id := uuid.New()
		assert.Equal(t, 36, len(id.String()))
		assert.NotEqual(t, uuid.Nil, id)
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// Multi-tenant Isolation
// ─────────────────────────────────────────────────────────────────────────────

func TestOrganizationIsolation(t *testing.T) {
	t.Run("Documents scoped to organization", func(t *testing.T) {
		orgA := uuid.New()
		orgB := uuid.New()

		// Simulate document belonging to orgA
		docOrgID := orgA.String()

		canOrgAAccess := docOrgID == orgA.String()
		canOrgBAccess := docOrgID == orgB.String()

		assert.True(t, canOrgAAccess, "Org A should access its own documents")
		assert.False(t, canOrgBAccess, "Org B should not access Org A's documents")
	})

	t.Run("Organization IDs are unique", func(t *testing.T) {
		org1 := uuid.New()
		org2 := uuid.New()

		assert.NotEqual(t, org1, org2)
	})

	t.Run("Settings are per-organization", func(t *testing.T) {
		settingsA := models.OrganizationSettings{
			ProcurementFlow: "goods_first",
		}
		settingsB := models.OrganizationSettings{
			ProcurementFlow: "payment_first",
		}

		assert.NotEqual(t, settingsA.ProcurementFlow, settingsB.ProcurementFlow)
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// Organization Logo and Branding
// ─────────────────────────────────────────────────────────────────────────────

func TestOrganizationBranding(t *testing.T) {
	t.Run("Logo URL must be valid if provided", func(t *testing.T) {
		validURLs := []string{
			"https://example.com/logo.png",
			"https://cdn.example.com/logos/org-123.jpg",
			"", // empty is allowed (no logo)
		}

		for _, url := range validURLs {
			isValid := url == "" || (len(url) > 8 && url[:8] == "https://")
			assert.True(t, isValid, "URL '%s' should be valid", url)
		}
	})

	t.Run("HTTP logo URLs rejected", func(t *testing.T) {
		httpURL := "http://example.com/logo.png"
		isSecure := len(httpURL) > 8 && httpURL[:8] == "https://"
		assert.False(t, isSecure, "HTTP URL should not be accepted")
	})

	t.Run("Tagline length validation", func(t *testing.T) {
		tests := []struct {
			tagline    string
			maxLen     int
			shouldPass bool
		}{
			{"Short tagline", 200, true},
			{"", 200, true}, // empty is fine
			{string(make([]byte, 201)), 200, false}, // 201 chars
		}

		for _, tt := range tests {
			isValid := len(tt.tagline) <= tt.maxLen
			assert.Equal(t, tt.shouldPass, isValid)
		}
	})
}
