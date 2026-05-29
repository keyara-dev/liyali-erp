package handlers

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/models"
	"github.com/stretchr/testify/assert"
)

// ─────────────────────────────────────────────────────────────────────────────
// Sign-off test helpers
// ─────────────────────────────────────────────────────────────────────────────

const (
	certifierUserID = "cert-user-001"
	certifierRole   = "manager"
)

// newGRNSignoffApp returns a Fiber app with the receiver, certifier, and
// completion endpoints mounted. The auth middleware applied here is the
// caller-supplied one so individual tests can swap the acting user / role.
func newGRNSignoffApp(t *testing.T, auth fiber.Handler) *fiber.App {
	t.Helper()
	app := fiber.New()
	app.Post("/grns/:id/sign-receive", auth, SignReceiveGRN)
	app.Post("/grns/:id/certify", auth, CertifyGRN)
	app.Post("/grns/:id/complete", auth, MarkGRNComplete)
	app.Post("/grns/:id/submit", auth, SubmitGRN)
	return app
}

// makeDraftGRNInState writes a DRAFT GRN whose SignoffStatus the test controls.
func makeDraftGRNInState(t *testing.T, docNum, signoff string) models.GoodsReceivedNote {
	t.Helper()
	grn := makeGRN(t, docNum, "PO-SIGNOFF-001", "DRAFT")
	grn.SignoffStatus = signoff
	grn.ReceivedBy = "creator-other"
	grn.CreatedBy = "creator-other"
	if err := config.DB.Save(&grn).Error; err != nil {
		t.Fatalf("update signoff state: %v", err)
	}
	return grn
}

func seedCertifier(t *testing.T) {
	t.Helper()
	u := models.User{
		ID:        certifierUserID,
		Name:      "Cert User",
		Email:     "cert@example.com",
		Role:      certifierRole,
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := config.DB.Create(&u).Error; err != nil {
		t.Fatalf("seedCertifier: %v", err)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// SignReceiveGRN
// ─────────────────────────────────────────────────────────────────────────────

func TestSignReceiveGRN_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	seedTestUser(t)

	grn := makeDraftGRNInState(t, "GRN-SIGNOFF-RECV-OK", "PENDING_RECEIVER")
	app := newGRNSignoffApp(t, withTenantCtx(testOrgID, testUserID, testUserRole))

	resp := testRequest(app, http.MethodPost, "/grns/"+grn.ID+"/sign-receive", map[string]interface{}{
		"receivedByName": "Jane Receiver",
		"signature":      "data:image/png;base64,SIG",
	})
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var updated models.GoodsReceivedNote
	config.DB.Where("id = ?", grn.ID).First(&updated)
	assert.Equal(t, "PENDING_CERTIFIER", updated.SignoffStatus)
	assert.Equal(t, "Jane Receiver", updated.ReceivedByName)
	assert.NotEmpty(t, updated.ReceivedBySignature)
	assert.NotNil(t, updated.ReceivedAt)
}

func TestSignReceiveGRN_MissingFields(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	seedTestUser(t)

	grn := makeDraftGRNInState(t, "GRN-SIGNOFF-RECV-MISS", "PENDING_RECEIVER")
	app := newGRNSignoffApp(t, withTenantCtx(testOrgID, testUserID, testUserRole))

	resp := testRequest(app, http.MethodPost, "/grns/"+grn.ID+"/sign-receive", map[string]interface{}{
		"receivedByName": "",
		"signature":      "",
	})
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestSignReceiveGRN_WrongSignoffState(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	seedTestUser(t)

	// Already past the receiver step → must reject.
	grn := makeDraftGRNInState(t, "GRN-SIGNOFF-RECV-BAD", "PENDING_CERTIFIER")
	app := newGRNSignoffApp(t, withTenantCtx(testOrgID, testUserID, testUserRole))

	resp := testRequest(app, http.MethodPost, "/grns/"+grn.ID+"/sign-receive", map[string]interface{}{
		"receivedByName": "Jane",
		"signature":      "sig",
	})
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestSignReceiveGRN_NonDraftStatus(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	seedTestUser(t)

	grn := makeGRN(t, "GRN-SIGNOFF-RECV-NONDRAFT", "PO-001", "PENDING")
	grn.SignoffStatus = "PENDING_RECEIVER"
	config.DB.Save(&grn)

	app := newGRNSignoffApp(t, withTenantCtx(testOrgID, testUserID, testUserRole))
	resp := testRequest(app, http.MethodPost, "/grns/"+grn.ID+"/sign-receive", map[string]interface{}{
		"receivedByName": "Jane",
		"signature":      "sig",
	})
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// ─────────────────────────────────────────────────────────────────────────────
// CertifyGRN
// ─────────────────────────────────────────────────────────────────────────────

func TestCertifyGRN_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	seedTestUser(t)
	seedCertifier(t)

	grn := makeDraftGRNInState(t, "GRN-CERT-OK", "PENDING_CERTIFIER")

	app := newGRNSignoffApp(t, withTenantCtx(testOrgID, certifierUserID, certifierRole))
	resp := testRequest(app, http.MethodPost, "/grns/"+grn.ID+"/certify", map[string]interface{}{
		"signature":     "data:image/png;base64,STAMP",
		"comments":      "All items in good condition",
		"stampImageUrl": "data:image/png;base64,STAMPIMG",
	})
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var updated models.GoodsReceivedNote
	config.DB.Where("id = ?", grn.ID).First(&updated)
	assert.Equal(t, "READY", updated.SignoffStatus)
	assert.Equal(t, certifierUserID, updated.CertifiedByID)
	assert.Equal(t, "Cert User", updated.CertifiedByName)
	assert.NotEmpty(t, updated.CertifiedBySignature)
	assert.NotEmpty(t, updated.StampImageURL)
	assert.NotNil(t, updated.CertifiedAt)
}

func TestCertifyGRN_NonPrivilegedRoleForbidden(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	seedTestUser(t)

	grn := makeDraftGRNInState(t, "GRN-CERT-LOWROLE", "PENDING_CERTIFIER")

	// Use the "requester" role — not in the privileged certifier set.
	app := newGRNSignoffApp(t, withTenantCtx(testOrgID, "req-user", "requester"))
	resp := testRequest(app, http.MethodPost, "/grns/"+grn.ID+"/certify", map[string]interface{}{
		"signature": "sig",
	})
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestCertifyGRN_SeparationOfDuties(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	seedTestUser(t)
	seedCertifier(t)

	// GRN whose creator/receiver is the same person attempting to certify.
	grn := makeGRN(t, "GRN-CERT-SOD", "PO-001", "DRAFT")
	grn.SignoffStatus = "PENDING_CERTIFIER"
	grn.CreatedBy = certifierUserID
	grn.ReceivedBy = certifierUserID
	config.DB.Save(&grn)

	app := newGRNSignoffApp(t, withTenantCtx(testOrgID, certifierUserID, certifierRole))
	resp := testRequest(app, http.MethodPost, "/grns/"+grn.ID+"/certify", map[string]interface{}{
		"signature": "sig",
	})
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestCertifyGRN_WrongState(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	seedTestUser(t)
	seedCertifier(t)

	// Receiver hasn't signed yet.
	grn := makeDraftGRNInState(t, "GRN-CERT-WRONG", "PENDING_RECEIVER")

	app := newGRNSignoffApp(t, withTenantCtx(testOrgID, certifierUserID, certifierRole))
	resp := testRequest(app, http.MethodPost, "/grns/"+grn.ID+"/certify", map[string]interface{}{
		"signature": "sig",
	})
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestCertifyGRN_MissingSignature(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	seedTestUser(t)
	seedCertifier(t)

	grn := makeDraftGRNInState(t, "GRN-CERT-NOSIG", "PENDING_CERTIFIER")

	app := newGRNSignoffApp(t, withTenantCtx(testOrgID, certifierUserID, certifierRole))
	resp := testRequest(app, http.MethodPost, "/grns/"+grn.ID+"/certify", map[string]interface{}{
		"signature": "   ",
	})
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// ─────────────────────────────────────────────────────────────────────────────
// MarkGRNComplete
// ─────────────────────────────────────────────────────────────────────────────

func TestMarkGRNComplete_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	seedTestUser(t)

	grn := makeDraftGRNInState(t, "GRN-COMPLETE-OK", "READY")

	app := newGRNSignoffApp(t, withTenantCtx(testOrgID, testUserID, testUserRole))
	resp := testRequest(app, http.MethodPost, "/grns/"+grn.ID+"/complete", map[string]interface{}{
		"comments": "Direct completion",
	})
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var updated models.GoodsReceivedNote
	config.DB.Where("id = ?", grn.ID).First(&updated)
	assert.Equal(t, strings.ToUpper(models.StatusCompleted), strings.ToUpper(updated.Status))
	assert.Equal(t, "COMPLETED", updated.SignoffStatus)
}

func TestMarkGRNComplete_NotReady(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	seedTestUser(t)

	grn := makeDraftGRNInState(t, "GRN-COMPLETE-NR", "PENDING_CERTIFIER")

	app := newGRNSignoffApp(t, withTenantCtx(testOrgID, testUserID, testUserRole))
	resp := testRequest(app, http.MethodPost, "/grns/"+grn.ID+"/complete", nil)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestMarkGRNComplete_NonDraft(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	seedTestUser(t)

	grn := makeGRN(t, "GRN-COMPLETE-ND", "PO-001", "PENDING")
	grn.SignoffStatus = "READY"
	config.DB.Save(&grn)

	app := newGRNSignoffApp(t, withTenantCtx(testOrgID, testUserID, testUserRole))
	resp := testRequest(app, http.MethodPost, "/grns/"+grn.ID+"/complete", nil)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// ─────────────────────────────────────────────────────────────────────────────
// SubmitGRN — sign-off gate
// ─────────────────────────────────────────────────────────────────────────────

func TestSubmitGRN_SignoffNotReady_Rejected(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	seedTestUser(t)

	// Linked PO must exist or the linked-PO check fires first.
	makeApprovedPO(t, "PO-SIGNOFF-001")
	grn := makeDraftGRNInState(t, "GRN-SUB-NOTREADY", "PENDING_CERTIFIER")

	app := newGRNSignoffApp(t, withTenantCtx(testOrgID, testUserID, testUserRole))
	resp := testRequest(app, http.MethodPost, "/grns/"+grn.ID+"/submit", map[string]interface{}{
		"workflowId": uuid.New().String(),
	})
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	body := decodeResponse(resp)
	assert.Contains(t, body["message"], "signed by both")
}
