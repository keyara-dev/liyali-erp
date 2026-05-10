package services

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/repository"
	"github.com/liyali/liyali-gateway/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ─────────────────────────────────────────────────────────────────────────────
// Shared test helpers (mirrors handlers/testutils_test.go patterns)
// ─────────────────────────────────────────────────────────────────────────────

func setupExecutionTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err, "open in-memory SQLite")

	// Pin to single connection so all goroutines share the same in-memory DB.
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)

	// Models with plain GORM tags are safe for AutoMigrate.
	err = db.AutoMigrate(
		&models.Organization{},
		&models.User{},
		&models.Vendor{},
		&models.Payee{},
		&models.Requisition{},
		&models.PurchaseOrder{},
		&models.PaymentVoucher{},
	)
	require.NoError(t, err, "auto-migrate models")

	// Workflow* models use type:uuid / type:jsonb / PostgreSQL defaults — use raw DDL.
	for _, ddl := range []string{
		`CREATE TABLE IF NOT EXISTS workflows (
			id TEXT PRIMARY KEY,
			organization_id TEXT NOT NULL DEFAULT '',
			name TEXT NOT NULL DEFAULT '',
			description TEXT,
			document_type TEXT NOT NULL DEFAULT '',
			entity_type TEXT NOT NULL DEFAULT '',
			version INTEGER DEFAULT 1,
			is_active NUMERIC DEFAULT 1,
			is_default NUMERIC DEFAULT 0,
			conditions JSON,
			stages JSON NOT NULL DEFAULT '[]',
			created_by TEXT NOT NULL DEFAULT '',
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS workflow_assignments (
			id TEXT PRIMARY KEY,
			organization_id TEXT NOT NULL DEFAULT '',
			entity_id TEXT NOT NULL DEFAULT '',
			entity_type TEXT NOT NULL DEFAULT '',
			workflow_id TEXT NOT NULL DEFAULT '',
			workflow_version INTEGER NOT NULL DEFAULT 1,
			current_stage INTEGER DEFAULT 0,
			status TEXT DEFAULT 'IN_PROGRESS',
			stage_history JSON,
			assigned_at DATETIME,
			assigned_by TEXT NOT NULL DEFAULT '',
			completed_at DATETIME,
			created_at DATETIME,
			updated_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS workflow_tasks (
			id TEXT PRIMARY KEY,
			organization_id TEXT NOT NULL DEFAULT '',
			workflow_assignment_id TEXT NOT NULL DEFAULT '',
			entity_id TEXT NOT NULL DEFAULT '',
			entity_type TEXT NOT NULL DEFAULT '',
			stage_number INTEGER NOT NULL DEFAULT 0,
			stage_name TEXT NOT NULL DEFAULT '',
			assignment_type TEXT DEFAULT 'role',
			assigned_role TEXT,
			assigned_user_id TEXT,
			status TEXT DEFAULT 'PENDING',
			priority TEXT DEFAULT 'MEDIUM',
			created_at DATETIME,
			claimed_at DATETIME,
			claimed_by TEXT,
			completed_at DATETIME,
			due_date DATETIME,
			version INTEGER DEFAULT 1,
			updated_by TEXT,
			claim_expiry DATETIME,
			updated_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS audit_logs (
			id TEXT PRIMARY KEY,
			organization_id TEXT,
			user_id TEXT,
			action TEXT,
			entity_type TEXT,
			entity_id TEXT,
			details TEXT,
			ip_address TEXT,
			user_agent TEXT,
			created_at DATETIME
		)`,
	} {
		require.NoError(t, db.Exec(ddl).Error, "create DDL table")
	}

	return db
}

// seedOrg creates a minimal Organization row for FK constraints.
func seedOrg(t *testing.T, db *gorm.DB, orgID string) {
	t.Helper()
	org := models.Organization{
		ID:        orgID,
		Name:      "Test Org",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	require.NoError(t, db.Create(&org).Error)
}

// seedUser creates a minimal User row.
func seedUser(t *testing.T, db *gorm.DB, userID string) {
	t.Helper()
	u := models.User{
		ID:        userID,
		Email:     userID + "@test.local",
		Name:      "Test User",
		Role:      "requester",
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	require.NoError(t, db.Create(&u).Error)
}

// seedWorkflow inserts a workflow row directly via SQL (bypasses uuid/jsonb type issues).
func seedWorkflow(t *testing.T, db *gorm.DB, orgID, routingType string, hasStages bool) string {
	t.Helper()

	wfID := uuid.New().String()

	stages := "[]"
	if hasStages {
		stagesData, _ := json.Marshal([]models.WorkflowStage{{
			StageNumber: 1, StageName: "Review", RequiredRole: "approver",
		}})
		stages = string(stagesData)
	}

	conditions := map[string]interface{}{
		"routingType":    routingType,
		"autoApprove":    true,
		"autoGeneratePO": true,
		"autoApprovePO":  true,
	}
	if routingType == models.RoutingTypeDirectPayment {
		conditions["autoApprovalMaxAmount"] = 999999.0
	}
	condJSON, _ := json.Marshal(conditions)

	now := time.Now().Format(time.RFC3339)
	err := db.Exec(`INSERT INTO workflows
		(id, organization_id, name, description, document_type, entity_type, version,
		 is_active, is_default, conditions, stages, created_by, created_at, updated_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		wfID, orgID, "Test Workflow", "", "requisition", "requisition",
		1, 1, 0, string(condJSON), stages, "system", now, now,
	).Error
	require.NoError(t, err, "seed workflow")
	return wfID
}

// seedRequisition inserts a Requisition with optional payee fields.
func seedRequisition(t *testing.T, db *gorm.DB, orgID, userID string, payeeSnap []byte) models.Requisition {
	t.Helper()

	req := models.Requisition{
		ID:             uuid.New().String(),
		OrganizationID: orgID,
		DocumentNumber: "REQ-" + uuid.New().String()[:8],
		RequesterId:    userID,
		RequesterName:  "Test User",
		Title:          "Direct Payment Requisition",
		Status:         models.StatusDraft,
		TotalAmount:    1500.00,
		Currency:       "ZMW",
		RoutingType:    models.RoutingTypeProcurement, // will be updated by submit
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	req.Items = datatypes.NewJSONType([]types.RequisitionItem{})
	req.ApprovalHistory = datatypes.NewJSONType([]types.ApprovalRecord{})

	if len(payeeSnap) > 0 {
		req.PayeeSnapshot = datatypes.JSON(payeeSnap)
	}

	require.NoError(t, db.Create(&req).Error)
	return req
}

// newExecutionService wires up a WorkflowExecutionService backed by the test DB.
func newExecutionService(t *testing.T, db *gorm.DB) *WorkflowExecutionService {
	t.Helper()
	repo := repository.NewWorkflowRepository(nil, db)
	auditSvc := NewAuditService()
	wfSvc := NewWorkflowService(repo, auditSvc, db)
	autoSvc := NewDocumentAutomationService(db, auditSvc, nil)
	return NewWorkflowExecutionService(db, wfSvc, auditSvc, autoSvc)
}

// ─────────────────────────────────────────────────────────────────────────────
// Tests
// ─────────────────────────────────────────────────────────────────────────────

func TestSubmitRequisitionWithRouting_DirectPayment_CreatesAutoPOAndDraftPV(t *testing.T) {
	db := setupExecutionTestDB(t)
	const orgID = "org-dp-001"
	const userID = "user-dp-001"
	seedOrg(t, db, orgID)
	seedUser(t, db, userID)

	// Seed placeholder vendor (required by automation service).
	db.Exec(`INSERT OR IGNORE INTO vendors (id, name, created_at, updated_at) VALUES ('vendor-placeholder-001','TBD Vendor',datetime('now'),datetime('now'))`)

	snap, _ := json.Marshal(map[string]interface{}{
		"name":      "John Doe",
		"payeeType": "employee",
	})
	req := seedRequisition(t, db, orgID, userID, snap)
	wfID := seedWorkflow(t, db, orgID, models.RoutingTypeDirectPayment, false)

	svc := newExecutionService(t, db)
	res, err := svc.SubmitRequisitionWithRouting(context.Background(), orgID, req.ID, wfID, userID, &req)

	require.NoError(t, err)
	assert.NotEmpty(t, res.AutoCreatedPOID, "auto-created PO ID must be set")
	assert.NotEmpty(t, res.AutoCreatedPVID, "auto-created PV ID must be set for direct_payment")
	assert.Equal(t, models.RoutingTypeDirectPayment, res.RoutingType)

	// Verify PV row in DB.
	var pv models.PaymentVoucher
	require.NoError(t, db.First(&pv, "id = ?", res.AutoCreatedPVID).Error)
	assert.Equal(t, models.StatusDraft, pv.Status)
	assert.Equal(t, models.RoutingTypeDirectPayment, pv.RoutingType)
	assert.Equal(t, "payment_first", pv.ProcurementFlow)
	assert.Equal(t, "John Doe", pv.VendorName)
	assert.Equal(t, req.TotalAmount, pv.Amount)

	// Verify routing_type denormalized onto requisition.
	var updatedReq models.Requisition
	require.NoError(t, db.First(&updatedReq, "id = ?", req.ID).Error)
	assert.Equal(t, models.RoutingTypeDirectPayment, updatedReq.RoutingType)
}

func TestSubmitRequisitionWithRouting_Accounting_DoesNotCreatePV(t *testing.T) {
	db := setupExecutionTestDB(t)
	const orgID = "org-acc-001"
	const userID = "user-acc-001"
	seedOrg(t, db, orgID)
	seedUser(t, db, userID)

	db.Exec(`INSERT OR IGNORE INTO vendors (id, name, created_at, updated_at) VALUES ('vendor-placeholder-001','TBD Vendor',datetime('now'),datetime('now'))`)

	req := seedRequisition(t, db, orgID, userID, nil)
	wfID := seedWorkflow(t, db, orgID, models.RoutingTypeAccounting, false)

	svc := newExecutionService(t, db)
	res, err := svc.SubmitRequisitionWithRouting(context.Background(), orgID, req.ID, wfID, userID, &req)

	require.NoError(t, err)
	assert.NotEmpty(t, res.AutoCreatedPOID, "accounting auto-approval should still create PO")
	assert.Empty(t, res.AutoCreatedPVID, "accounting path must NOT create a PV")
	assert.Equal(t, models.RoutingTypeAccounting, res.RoutingType)
}

func TestSubmitRequisitionWithRouting_DirectPayment_MissingPayee(t *testing.T) {
	db := setupExecutionTestDB(t)
	const orgID = "org-nopayee-001"
	const userID = "user-nopayee-001"
	seedOrg(t, db, orgID)
	seedUser(t, db, userID)

	// Requisition has no payee_id and no payee_snapshot.
	req := seedRequisition(t, db, orgID, userID, nil)
	wfID := seedWorkflow(t, db, orgID, models.RoutingTypeDirectPayment, false)

	svc := newExecutionService(t, db)
	_, err := svc.SubmitRequisitionWithRouting(context.Background(), orgID, req.ID, wfID, userID, &req)

	require.Error(t, err)
	assert.True(t, strings.Contains(strings.ToLower(err.Error()), "payee"),
		"error must mention 'payee', got: %s", err.Error())
}
