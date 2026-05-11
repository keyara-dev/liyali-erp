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
		&models.GoodsReceivedNote{},
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
			kind TEXT NOT NULL DEFAULT 'approval',
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
		`CREATE TABLE IF NOT EXISTS stage_approval_records (
			id TEXT PRIMARY KEY,
			organization_id TEXT NOT NULL DEFAULT '',
			workflow_task_id TEXT NOT NULL DEFAULT '',
			stage_number INTEGER NOT NULL DEFAULT 0,
			approver_id TEXT NOT NULL DEFAULT '',
			approver_name TEXT NOT NULL DEFAULT '',
			approver_role TEXT NOT NULL DEFAULT '',
			man_number TEXT NOT NULL DEFAULT '',
			position TEXT NOT NULL DEFAULT '',
			action TEXT NOT NULL DEFAULT '',
			comments TEXT NOT NULL DEFAULT '',
			signature TEXT NOT NULL DEFAULT '',
			approved_at DATETIME,
			ip_address TEXT NOT NULL DEFAULT '',
			user_agent TEXT NOT NULL DEFAULT '',
			created_at DATETIME,
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
	assert.Equal(t, "John Doe", pv.VendorName)
	assert.Equal(t, req.TotalAmount, pv.Amount)

	// procurement_flow lives on the PO, not the PV (PV has no such column).
	var po models.PurchaseOrder
	require.NoError(t, db.First(&po, "id = ?", res.AutoCreatedPOID).Error)
	assert.Equal(t, "payment_first", po.ProcurementFlow)

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

func TestSubmitRequisitionWithRouting_DirectPayment_RejectsStagesGT0(t *testing.T) {
	db := setupExecutionTestDB(t)
	const orgID = "org-stages-001"
	const userID = "user-stages-001"
	seedOrg(t, db, orgID)
	seedUser(t, db, userID)

	snap, _ := json.Marshal(map[string]interface{}{
		"name":      "Alice",
		"payeeType": "employee",
	})
	req := seedRequisition(t, db, orgID, userID, snap)
	// Seed a direct_payment workflow that (incorrectly) has stages > 0.
	wfID := seedWorkflow(t, db, orgID, models.RoutingTypeDirectPayment, true)

	svc := newExecutionService(t, db)
	_, err := svc.SubmitRequisitionWithRouting(context.Background(), orgID, req.ID, wfID, userID, &req)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "must have 0 approval stages",
		"error must mention stage count, got: %s", err.Error())
}

func TestSubmitRequisitionWithRouting_DirectPayment_PVFailureDoesNotRollbackPO(t *testing.T) {
	db := setupExecutionTestDB(t)
	const orgID = "org-pvfail-001"
	const userID = "user-pvfail-001"
	seedOrg(t, db, orgID)
	seedUser(t, db, userID)

	db.Exec(`INSERT OR IGNORE INTO vendors (id, name, created_at, updated_at) VALUES ('vendor-placeholder-001','TBD Vendor',datetime('now'),datetime('now'))`)

	snap, _ := json.Marshal(map[string]interface{}{
		"name":      "Bob",
		"payeeType": "supplier",
	})
	req := seedRequisition(t, db, orgID, userID, snap)
	wfID := seedWorkflow(t, db, orgID, models.RoutingTypeDirectPayment, false)

	svc := newExecutionService(t, db)

	// First submission: creates PO + PV successfully.
	res1, err := svc.SubmitRequisitionWithRouting(context.Background(), orgID, req.ID, wfID, userID, &req)
	require.NoError(t, err)
	require.NotEmpty(t, res1.AutoCreatedPOID, "first run must create PO")
	require.NotEmpty(t, res1.AutoCreatedPVID, "first run must create PV")

	// Verify PO exists.
	var po models.PurchaseOrder
	require.NoError(t, db.First(&po, "id = ?", res1.AutoCreatedPOID).Error)

	// Pre-seed a second PV with the same linked_po to cause a duplicate when
	// autoCreateDraftPV is called again, simulating a PV creation failure.
	// We do this by submitting a second (duplicate) requisition whose PV will
	// collide with the one already created — instead, we directly verify
	// the invariant: if we cannot insert the PV, the PO must survive.
	//
	// The simplest reliable proof: the PO row from the first run is still in DB
	// after the service returned without error, confirming the "PO preserved"
	// design even under the happy path (structure-level test).
	assert.Equal(t, models.RoutingTypeDirectPayment, po.RoutingType)
	assert.Equal(t, "payment_first", po.ProcurementFlow)

	// Simulate PV failure by inserting a PV with a conflicting document number
	// equal to what the NEXT call would generate, then call the service again
	// with a fresh requisition — the duplicate linked_po unique constraint
	// (if it existed) or any other DB error in PV create must not roll back the PO.
	// Since SQLite has no unique index on linked_po for PVs in these tests,
	// we instead verify the code path: result has AutoCreatedPOID, empty PVError
	// is silent in logs. The critical assertion is that res1.AutoCreatedPOID != ""
	// and the PO persists even when PV fails.
	//
	// Full proof: second requisition with autoCreateDraftPV pre-blocked.
	req2 := seedRequisition(t, db, orgID, userID, snap)
	// Drop the payment_vouchers table to force a PV insert error, but
	// keep purchase_orders so PO creation succeeds.
	db.Exec(`DROP TABLE payment_vouchers`)

	res2, err2 := svc.SubmitRequisitionWithRouting(context.Background(), orgID, req2.ID, wfID, userID, &req2)

	// PV failure must NOT surface as an error to the caller.
	require.NoError(t, err2, "PV failure must not propagate as an error")
	assert.NotEmpty(t, res2.AutoCreatedPOID, "PO must be created even when PV fails")
	assert.Empty(t, res2.AutoCreatedPVID, "PV ID must be empty when PV creation fails")

	// PO must exist in DB (not rolled back).
	var po2 models.PurchaseOrder
	require.NoError(t, db.First(&po2, "id = ?", res2.AutoCreatedPOID).Error,
		"PO must survive a PV creation failure")
}

// ─────────────────────────────────────────────────────────────────────────────
// Helpers for GRN cascade tests
// ─────────────────────────────────────────────────────────────────────────────

// seedApprovedPOWithItems creates an APPROVED PurchaseOrder with the given items.
func seedApprovedPOWithItems(t *testing.T, db *gorm.DB, orgID string, items []types.POItem) models.PurchaseOrder {
	t.Helper()
	docNum := "PO-" + uuid.New().String()[:8]
	po := models.PurchaseOrder{
		ID:             uuid.New().String(),
		OrganizationID: orgID,
		DocumentNumber: docNum,
		Status:         models.StatusApproved,
		Currency:       "ZMW",
		TotalAmount:    999.0,
		DeliveryStatus: models.DeliveryStatusNotDelivered,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	po.Items = datatypes.NewJSONType(items)
	po.ApprovalHistory = datatypes.NewJSONType([]types.ApprovalRecord{})
	require.NoError(t, db.Create(&po).Error)
	return po
}

// seedPendingGRN creates a PENDING GoodsReceivedNote linked to the given PO.
func seedPendingGRN(t *testing.T, db *gorm.DB, orgID string, po models.PurchaseOrder, grnItems []types.GRNItem) models.GoodsReceivedNote {
	t.Helper()
	grn := models.GoodsReceivedNote{
		ID:               uuid.New().String(),
		OrganizationID:   orgID,
		DocumentNumber:   "GRN-" + uuid.New().String()[:8],
		PODocumentNumber: po.DocumentNumber,
		Status:           models.StatusPending,
		ReceivedDate:     time.Now(),
		ReceivedBy:       "warehouse-user",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
	grn.Items = datatypes.NewJSONType(grnItems)
	grn.QualityIssues = datatypes.NewJSONType([]types.QualityIssue{})
	grn.ApprovalHistory = datatypes.NewJSONType([]types.ApprovalRecord{})
	require.NoError(t, db.Create(&grn).Error)
	return grn
}

// seedGRNWorkflowAndClaim creates a 1-stage workflow assignment + claimed task for a GRN,
// and returns (assignmentID, taskID, approverUserID).
func seedGRNWorkflowAndClaim(t *testing.T, db *gorm.DB, orgID, grnID string) (string, string, string) {
	t.Helper()
	approverID := "approver-" + uuid.New().String()[:8]
	// Must have role="approver" to pass the permission check in ApproveWorkflowTaskWithVersion.
	require.NoError(t, db.Exec(
		`INSERT INTO users (id, email, name, role, active, created_at, updated_at) VALUES (?,?,?,?,?,datetime('now'),datetime('now'))`,
		approverID, approverID+"@test.local", "Test Approver", "approver", true,
	).Error)

	now := time.Now()
	nowStr := now.Format(time.RFC3339)

	wfID := uuid.New().String()
	stages, _ := json.Marshal([]models.WorkflowStage{{
		StageNumber: 1, StageName: "Review", RequiredRole: "approver",
	}})
	require.NoError(t, db.Exec(`INSERT INTO workflows
		(id, organization_id, name, description, document_type, entity_type, version,
		 is_active, is_default, conditions, stages, created_by, created_at, updated_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		wfID, orgID, "GRN Workflow", "", "grn", "grn",
		1, 1, 0, `{}`, string(stages), "system", nowStr, nowStr,
	).Error)

	assignID := uuid.New().String()
	require.NoError(t, db.Exec(`INSERT INTO workflow_assignments
		(id, organization_id, entity_id, entity_type, workflow_id, workflow_version,
		 current_stage, status, assigned_at, assigned_by, created_at, updated_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`,
		assignID, orgID, grnID, "grn", wfID, 1,
		1, "IN_PROGRESS", nowStr, approverID, nowStr, nowStr,
	).Error)

	taskID := uuid.New().String()
	expiry := now.Add(30 * time.Minute).Format(time.RFC3339)
	require.NoError(t, db.Exec(`INSERT INTO workflow_tasks
		(id, organization_id, workflow_assignment_id, entity_id, entity_type,
		 stage_number, stage_name, kind, assignment_type, assigned_role,
		 status, priority, version, claimed_at, claimed_by, claim_expiry, created_at, updated_at)
		VALUES (?,?,?,?,?, ?,?,?,?,?, ?,?,?,?,?,?,?,?)`,
		taskID, orgID, assignID, grnID, "grn",
		1, "Review", "approval", "role", "approver",
		"CLAIMED", "MEDIUM", 1, nowStr, approverID, expiry, nowStr, nowStr,
	).Error)

	return assignID, taskID, approverID
}

// ─────────────────────────────────────────────────────────────────────────────
// GRN Approval Cascade Tests
// ─────────────────────────────────────────────────────────────────────────────

func TestGRNApprovalCascadesToPO_SingleFullGRN_FullyDelivered(t *testing.T) {
	db := setupExecutionTestDB(t)
	const orgID = "org-cascade-full-001"
	seedOrg(t, db, orgID)

	items := []types.POItem{{Description: "Widget", Quantity: 10, UnitPrice: 5.0, Amount: 50.0}}
	po := seedApprovedPOWithItems(t, db, orgID, items)

	grnItems := []types.GRNItem{{
		Description:      "Widget",
		QuantityOrdered:  10,
		QuantityReceived: 10,
		Variance:         0,
		Condition:        "good",
	}}
	grn := seedPendingGRN(t, db, orgID, po, grnItems)
	_, taskID, approverID := seedGRNWorkflowAndClaim(t, db, orgID, grn.ID)

	svc := newExecutionService(t, db)
	err := svc.ApproveWorkflowTaskWithVersion(context.Background(), taskID, approverID, "LGTM", "", 1)
	require.NoError(t, err)

	var updatedPO models.PurchaseOrder
	require.NoError(t, db.First(&updatedPO, "id = ?", po.ID).Error)
	assert.Equal(t, models.DeliveryStatusFullyDelivered, updatedPO.DeliveryStatus,
		"PO delivery_status must be FULLY_DELIVERED after full GRN approval")

	poItems := updatedPO.Items.Data()
	require.Len(t, poItems, 1)
	assert.Equal(t, 10, poItems[0].ReceivedQuantity,
		"POItem.ReceivedQuantity must equal GRN QuantityReceived")
}

func TestGRNApprovalCascadesToPO_PartialGRN_PartiallyDelivered(t *testing.T) {
	db := setupExecutionTestDB(t)
	const orgID = "org-cascade-partial-001"
	seedOrg(t, db, orgID)

	items := []types.POItem{{Description: "Gadget", Quantity: 10, UnitPrice: 8.0, Amount: 80.0}}
	po := seedApprovedPOWithItems(t, db, orgID, items)

	grnItems := []types.GRNItem{{
		Description:      "Gadget",
		QuantityOrdered:  10,
		QuantityReceived: 4,
		Variance:         6,
		Condition:        "good",
	}}
	grn := seedPendingGRN(t, db, orgID, po, grnItems)
	_, taskID, approverID := seedGRNWorkflowAndClaim(t, db, orgID, grn.ID)

	svc := newExecutionService(t, db)
	err := svc.ApproveWorkflowTaskWithVersion(context.Background(), taskID, approverID, "partial receipt", "", 1)
	require.NoError(t, err)

	var updatedPO models.PurchaseOrder
	require.NoError(t, db.First(&updatedPO, "id = ?", po.ID).Error)
	assert.Equal(t, models.DeliveryStatusPartiallyDelivered, updatedPO.DeliveryStatus,
		"PO delivery_status must be PARTIALLY_DELIVERED when GRN covers only part of the order")

	poItems := updatedPO.Items.Data()
	require.Len(t, poItems, 1)
	assert.Equal(t, 4, poItems[0].ReceivedQuantity,
		"POItem.ReceivedQuantity must equal partial GRN QuantityReceived")
}
