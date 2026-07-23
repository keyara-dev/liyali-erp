package services

import (
	"context"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ─────────────────────────────────────────────────────────────────────────────
// Shared test helpers for DocumentAutomationService
// ─────────────────────────────────────────────────────────────────────────────

func setupDocAutomationTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err, "open in-memory SQLite")

	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)

	require.NoError(t,
		db.AutoMigrate(
			&models.PurchaseOrder{},
			&models.PaymentVoucher{},
			&models.GoodsReceivedNote{},
		),
		"auto-migrate PO+PV+GRN",
	)

	return db
}

func seedPO(t *testing.T, db *gorm.DB, orgID string) models.PurchaseOrder {
	t.Helper()
	po := models.PurchaseOrder{
		ID:             uuid.New().String(),
		DocumentNumber: "PO-" + uuid.New().String()[:8],
		OrganizationID: orgID,
		Status:         "APPROVED",
		TotalAmount:    1000.0,
		Currency:       "ZMW",
		DeliveryDate:   time.Now().AddDate(0, 1, 0),
		ApprovalStage:  1,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		Items:          datatypes.NewJSONType([]types.POItem{}),
		ApprovalHistory: datatypes.NewJSONType([]types.ApprovalRecord{}),
	}
	require.NoError(t, db.Create(&po).Error, "seed PO")
	return po
}

func seedGRN(t *testing.T, db *gorm.DB, orgID string, poDocNumber string) models.GoodsReceivedNote {
	t.Helper()
	grn := models.GoodsReceivedNote{
		ID:               uuid.New().String(),
		DocumentNumber:   "GRN-" + uuid.New().String()[:8],
		PODocumentNumber: poDocNumber,
		OrganizationID:   orgID,
		Status:           "APPROVED",
		ReceivedDate:     time.Now(),
		ReceivedBy:       "test-user",
		ApprovalStage:    1,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
		Items:            datatypes.NewJSONType([]types.GRNItem{}),
		ApprovalHistory:  datatypes.NewJSONType([]types.ApprovalRecord{}),
	}
	require.NoError(t, db.Create(&grn).Error, "seed GRN")
	return grn
}

func seedPV(t *testing.T, db *gorm.DB, orgID, poDocNumber, status string) models.PaymentVoucher {
	t.Helper()
	pv := models.PaymentVoucher{
		ID:             uuid.New().String(),
		DocumentNumber: "PV-" + uuid.New().String()[:8],
		OrganizationID: orgID,
		LinkedPO:       poDocNumber,
		Status:         status,
		Amount:         500.0,
		Currency:       "ZMW",
		ApprovalStage:  0,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		ApprovalHistory: datatypes.NewJSONType([]types.ApprovalRecord{}),
	}
	require.NoError(t, db.Create(&pv).Error, "seed PV")
	return pv
}

// ─────────────────────────────────────────────────────────────────────────────
// Tests for CreatePaymentVoucherFromGRN filter alignment
// ─────────────────────────────────────────────────────────────────────────────

// TestCreatePaymentVoucherFromGRN_RejectedPVAllowsAutoCreate verifies that a
// REJECTED PV no longer blocks auto-PV creation from GRN (filter fix: align from
// != 'CANCELLED' to NOT IN ('CANCELLED','REJECTED')).
func TestCreatePaymentVoucherFromGRN_RejectedPVAllowsAutoCreate(t *testing.T) {
	db := setupDocAutomationTestDB(t)
	ctx := context.Background()
	orgID := "org-test-1"

	// Seed PO and GRN
	po := seedPO(t, db, orgID)
	grn := seedGRN(t, db, orgID, po.DocumentNumber)

	// Seed a REJECTED PV for this PO
	// Before fix: this REJECTED PV will block auto-create (filter: != 'CANCELLED' matches REJECTED)
	// After fix: this REJECTED PV will NOT block auto-create (filter: NOT IN ('CANCELLED','REJECTED') excludes REJECTED)
	seedPV(t, db, orgID, po.DocumentNumber, "REJECTED")

	// Create DocumentAutomationService and attempt auto-create
	svc := NewDocumentAutomationService(db, nil, nil)
	config := AutomationConfig{
		AutoCreatePVFromGRN: true,
	}

	result, err := svc.CreatePaymentVoucherFromGRN(ctx, &grn, config)

	// After fix: should succeed (REJECTED PV is now allowed to be overridden)
	require.NoError(t, err)
	assert.True(t, result.Success, "CreatePaymentVoucherFromGRN should succeed when only a REJECTED PV exists")
	assert.Equal(t, "payment_voucher", result.DocumentType)
	assert.NotEmpty(t, result.DocumentID)
}

// TestCreatePaymentVoucherFromGRN_CancelledPVAllowsAutoCreate verifies that a
// CANCELLED PV does NOT block auto-PV creation (unchanged behavior).
func TestCreatePaymentVoucherFromGRN_CancelledPVAllowsAutoCreate(t *testing.T) {
	db := setupDocAutomationTestDB(t)
	ctx := context.Background()
	orgID := "org-test-2"

	// Seed PO and GRN
	po := seedPO(t, db, orgID)
	grn := seedGRN(t, db, orgID, po.DocumentNumber)

	// Seed a CANCELLED PV for this PO
	// Both before and after fix: CANCELLED PV should NOT block auto-create
	seedPV(t, db, orgID, po.DocumentNumber, "CANCELLED")

	// Create DocumentAutomationService and attempt auto-create
	svc := NewDocumentAutomationService(db, nil, nil)
	config := AutomationConfig{
		AutoCreatePVFromGRN: true,
	}

	result, err := svc.CreatePaymentVoucherFromGRN(ctx, &grn, config)

	// Should succeed: CANCELLED PV does not block creation
	require.NoError(t, err)
	assert.True(t, result.Success, "CreatePaymentVoucherFromGRN should succeed when only a CANCELLED PV exists")
	assert.Equal(t, "payment_voucher", result.DocumentType)
	assert.NotEmpty(t, result.DocumentID)
}

// TestCreatePaymentVoucherFromGRN_DraftPVBlocksAutoCreate verifies that a
// DRAFT PV still blocks auto-PV creation (unchanged behavior).
func TestCreatePaymentVoucherFromGRN_DraftPVBlocksAutoCreate(t *testing.T) {
	db := setupDocAutomationTestDB(t)
	ctx := context.Background()
	orgID := "org-test-3"

	// Seed PO and GRN
	po := seedPO(t, db, orgID)
	grn := seedGRN(t, db, orgID, po.DocumentNumber)

	// Seed a DRAFT PV for this PO
	// Both before and after fix: DRAFT PV should block auto-create
	seedPV(t, db, orgID, po.DocumentNumber, "DRAFT")

	// Create DocumentAutomationService and attempt auto-create
	svc := NewDocumentAutomationService(db, nil, nil)
	config := AutomationConfig{
		AutoCreatePVFromGRN: true,
	}

	result, err := svc.CreatePaymentVoucherFromGRN(ctx, &grn, config)

	// Should fail: DRAFT PV blocks creation
	require.NoError(t, err)
	assert.False(t, result.Success, "CreatePaymentVoucherFromGRN should fail when a DRAFT PV exists")
	assert.Contains(t, result.Error.Error(), "already exists")
}

// TestCreatePaymentVoucherFromGRN_NoExistingPVAllowsAutoCreate verifies that
// auto-create succeeds when no live PV exists for the PO.
func TestCreatePaymentVoucherFromGRN_NoExistingPVAllowsAutoCreate(t *testing.T) {
	db := setupDocAutomationTestDB(t)
	ctx := context.Background()
	orgID := "org-test-4"

	// Seed PO and GRN without any existing PV
	po := seedPO(t, db, orgID)
	grn := seedGRN(t, db, orgID, po.DocumentNumber)

	// Create DocumentAutomationService and attempt auto-create
	svc := NewDocumentAutomationService(db, nil, nil)
	config := AutomationConfig{
		AutoCreatePVFromGRN: true,
	}

	result, err := svc.CreatePaymentVoucherFromGRN(ctx, &grn, config)

	// Should succeed: no existing PV blocks creation
	require.NoError(t, err)
	assert.True(t, result.Success, "CreatePaymentVoucherFromGRN should succeed when no live PV exists")
	assert.Equal(t, "payment_voucher", result.DocumentType)
	assert.NotEmpty(t, result.DocumentID)
}

// ─────────────────────────────────────────────────────────────────────────────
// Regression tests for createDraftPVFromPO (WorkflowExecutionService) — the
// payment-first auto-create path keeps its own one-live-PV-per-PO skip guard
// (automation_engine.go:84). These lock the skip behavior in place after the
// balance-gate rework (multiple PVs are now allowed via the HTTP gate, but
// AUTO-creation still fires only when no live PV exists yet).
// ─────────────────────────────────────────────────────────────────────────────

// TestCreateDraftPVFromPO_SkipsWhenLivePVExists verifies createDraftPVFromPO
// returns ("", nil) — a skip, not an error — when a live PV already exists for
// the PO, so automation never produces a second PV behind the user's back.
func TestCreateDraftPVFromPO_SkipsWhenLivePVExists(t *testing.T) {
	db := setupDocAutomationTestDB(t)
	orgID := "org-draftpv-skip"

	po := seedPO(t, db, orgID)
	seedPV(t, db, orgID, po.DocumentNumber, "DRAFT") // live PV

	svc := &WorkflowExecutionService{db: db}
	pvID, err := svc.createDraftPVFromPO(&po)

	require.NoError(t, err)
	assert.Equal(t, "", pvID, "createDraftPVFromPO must skip (return empty id) when a live PV exists")

	// No new PV was created — the seeded one is the only row.
	var count int64
	require.NoError(t, db.Model(&models.PaymentVoucher{}).
		Where("linked_po = ? AND organization_id = ?", po.DocumentNumber, orgID).Count(&count).Error)
	assert.Equal(t, int64(1), count, "no additional PV should be created")
}

// TestCreateDraftPVFromPO_CreatesWhenOnlyRejectedPVExists verifies the skip
// guard excludes REJECTED (and CANCELLED) PVs, so a rejected auto-PV can be
// re-created rather than blocking the payment-first flow forever.
func TestCreateDraftPVFromPO_CreatesWhenOnlyRejectedPVExists(t *testing.T) {
	db := setupDocAutomationTestDB(t)
	orgID := "org-draftpv-rejected"

	po := seedPO(t, db, orgID)
	seedPV(t, db, orgID, po.DocumentNumber, "REJECTED") // terminal, must not block

	svc := &WorkflowExecutionService{db: db}
	pvID, err := svc.createDraftPVFromPO(&po)

	require.NoError(t, err)
	assert.NotEqual(t, "", pvID, "createDraftPVFromPO should create when only a REJECTED PV exists")

	var count int64
	require.NoError(t, db.Model(&models.PaymentVoucher{}).
		Where("linked_po = ? AND organization_id = ?", po.DocumentNumber, orgID).Count(&count).Error)
	assert.Equal(t, int64(2), count, "a new DRAFT PV should be created alongside the REJECTED one")
}
