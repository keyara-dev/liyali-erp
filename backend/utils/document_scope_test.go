package utils

import (
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ─────────────────────────────────────────────────────────────────────────────
// Shared test helper
// ─────────────────────────────────────────────────────────────────────────────

func setupScopeTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err, "open in-memory SQLite")

	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)

	// AutoMigrate models that have plain GORM tags.
	err = db.AutoMigrate(
		&models.Organization{},
		&models.User{},
		&models.Requisition{},
		&models.PurchaseOrder{},
		&models.PaymentVoucher{},
		&models.GoodsReceivedNote{},
	)
	require.NoError(t, err, "auto-migrate base models")

	// UserOrganizationRole and OrganizationRole use type:uuid / type:jsonb which
	// break SQLite AutoMigrate — create via raw DDL instead.
	for _, ddl := range []string{
		`CREATE TABLE IF NOT EXISTS organization_roles (
			id TEXT PRIMARY KEY,
			organization_id TEXT,
			name TEXT NOT NULL DEFAULT '',
			description TEXT DEFAULT '',
			is_system_role NUMERIC DEFAULT 0,
			permissions JSON DEFAULT '[]',
			active NUMERIC DEFAULT 1,
			created_by TEXT,
			created_at DATETIME,
			updated_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS user_organization_roles (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL DEFAULT '',
			organization_id TEXT NOT NULL DEFAULT '',
			role_id TEXT NOT NULL DEFAULT '',
			assigned_by TEXT,
			assigned_at DATETIME,
			active NUMERIC DEFAULT 1
		)`,
	} {
		require.NoError(t, db.Exec(ddl).Error, "create DDL table")
	}

	return db
}

func seedScopeOrg(t *testing.T, db *gorm.DB, orgID string) {
	t.Helper()
	org := models.Organization{
		ID:        orgID,
		Name:      "Test Org",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	require.NoError(t, db.Create(&org).Error)
}

func seedScopeUser(t *testing.T, db *gorm.DB, userID, role string) {
	t.Helper()
	u := models.User{
		ID:        userID,
		Email:     userID + "@test.local",
		Name:      "Test " + role,
		Role:      role,
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	require.NoError(t, db.Create(&u).Error)
}

// ─────────────────────────────────────────────────────────────────────────────
// GetDocumentScope — HideDirectPayment derivation
// ─────────────────────────────────────────────────────────────────────────────

func TestGetDocumentScope_ProcurementUserHidesDirectPayment(t *testing.T) {
	db := setupScopeTestDB(t)
	orgID := uuid.New().String()
	userID := uuid.New().String()
	seedScopeOrg(t, db, orgID)
	seedScopeUser(t, db, userID, "procurement")

	scope := GetDocumentScope(db, userID, "procurement", orgID)

	assert.True(t, scope.HideDirectPayment, "procurement role must have HideDirectPayment=true")
	assert.False(t, scope.CanViewAll, "procurement role must not have CanViewAll")
}

func TestGetDocumentScope_FinanceUserDoesNotHide(t *testing.T) {
	db := setupScopeTestDB(t)
	orgID := uuid.New().String()
	userID := uuid.New().String()
	seedScopeOrg(t, db, orgID)
	seedScopeUser(t, db, userID, "finance")

	scope := GetDocumentScope(db, userID, "finance", orgID)

	assert.False(t, scope.HideDirectPayment, "finance role must not have HideDirectPayment")
	assert.True(t, scope.CanViewAll, "finance role must have CanViewAll")
}

func TestGetDocumentScope_AdminDoesNotHide(t *testing.T) {
	db := setupScopeTestDB(t)
	orgID := uuid.New().String()
	userID := uuid.New().String()
	seedScopeOrg(t, db, orgID)
	seedScopeUser(t, db, userID, "admin")

	scope := GetDocumentScope(db, userID, "admin", orgID)

	assert.False(t, scope.HideDirectPayment, "admin role must not have HideDirectPayment")
	assert.True(t, scope.CanViewAll, "admin role must have CanViewAll")
}

// ─────────────────────────────────────────────────────────────────────────────
// ApplyToQuery — HideDirectPayment filtering
// ─────────────────────────────────────────────────────────────────────────────

func makeHidingScope(userID, orgID, role string) DocumentScope {
	return DocumentScope{
		CanViewAll:        false,
		IsProcurement:     true,
		HideDirectPayment: true,
		UserID:            userID,
		OrgID:             orgID,
		UserRole:          role,
	}
}

func TestApplyToQuery_HidesDirectPaymentRequisitions(t *testing.T) {
	db := setupScopeTestDB(t)
	orgID := uuid.New().String()
	userID := uuid.New().String()
	seedScopeOrg(t, db, orgID)
	seedScopeUser(t, db, userID, "procurement")

	// Seed two requisitions — one direct_payment, one procurement.
	directReq := models.Requisition{
		ID: uuid.New().String(), OrganizationID: orgID,
		DocumentNumber: "REQ-DIRECT-001",
		RequesterId:    userID, Status: "draft", Priority: "medium",
		RoutingType: "direct_payment",
		CreatedAt:   time.Now(), UpdatedAt: time.Now(),
	}
	procReq := models.Requisition{
		ID: uuid.New().String(), OrganizationID: orgID,
		DocumentNumber: "REQ-PROC-001",
		RequesterId:    userID, Status: "draft", Priority: "medium",
		RoutingType: "procurement",
		CreatedAt:   time.Now(), UpdatedAt: time.Now(),
	}
	require.NoError(t, db.Create(&directReq).Error)
	require.NoError(t, db.Create(&procReq).Error)

	scope := makeHidingScope(userID, orgID, "procurement")

	var results []models.Requisition
	q := db.Where("organization_id = ?", orgID)
	q = scope.ApplyToQuery(q, "requester_id", "requisition", "")
	require.NoError(t, q.Find(&results).Error)

	assert.Len(t, results, 1, "only the procurement requisition should be visible")
	assert.Equal(t, "REQ-PROC-001", results[0].DocumentNumber)
}

func TestApplyToQuery_HidesDirectPaymentPurchaseOrders(t *testing.T) {
	db := setupScopeTestDB(t)
	orgID := uuid.New().String()
	userID := uuid.New().String()
	seedScopeOrg(t, db, orgID)
	seedScopeUser(t, db, userID, "procurement")

	directPO := models.PurchaseOrder{
		ID: uuid.New().String(), OrganizationID: orgID,
		DocumentNumber: "PO-DIRECT-001",
		Status:         "draft",
		RoutingType:    "direct_payment",
		CreatedAt:      time.Now(), UpdatedAt: time.Now(),
	}
	procPO := models.PurchaseOrder{
		ID: uuid.New().String(), OrganizationID: orgID,
		DocumentNumber: "PO-PROC-001",
		Status:         "draft",
		RoutingType:    "procurement",
		CreatedAt:      time.Now(), UpdatedAt: time.Now(),
	}
	require.NoError(t, db.Create(&directPO).Error)
	require.NoError(t, db.Create(&procPO).Error)

	scope := makeHidingScope(userID, orgID, "procurement")

	var results []models.PurchaseOrder
	q := db.Where("organization_id = ?", orgID)
	q = scope.ApplyToQuery(q, "created_by", "purchase_order", "")
	require.NoError(t, q.Find(&results).Error)

	assert.Len(t, results, 1, "only the procurement PO should be visible")
	assert.Equal(t, "PO-PROC-001", results[0].DocumentNumber)
}

func TestApplyToQuery_HidesDirectPaymentPaymentVouchers(t *testing.T) {
	db := setupScopeTestDB(t)
	orgID := uuid.New().String()
	userID := uuid.New().String()
	seedScopeOrg(t, db, orgID)
	seedScopeUser(t, db, userID, "procurement")

	directPV := models.PaymentVoucher{
		ID: uuid.New().String(), OrganizationID: orgID,
		DocumentNumber: "PV-DIRECT-001",
		Status:         "draft",
		RoutingType:    "direct_payment",
		CreatedAt:      time.Now(), UpdatedAt: time.Now(),
	}
	procPV := models.PaymentVoucher{
		ID: uuid.New().String(), OrganizationID: orgID,
		DocumentNumber: "PV-PROC-001",
		Status:         "draft",
		RoutingType:    "procurement",
		CreatedAt:      time.Now(), UpdatedAt: time.Now(),
	}
	require.NoError(t, db.Create(&directPV).Error)
	require.NoError(t, db.Create(&procPV).Error)

	scope := makeHidingScope(userID, orgID, "procurement")

	var results []models.PaymentVoucher
	q := db.Where("organization_id = ?", orgID)
	q = scope.ApplyToQuery(q, "created_by", "payment_voucher", "")
	require.NoError(t, q.Find(&results).Error)

	assert.Len(t, results, 1, "only the procurement PV should be visible")
	assert.Equal(t, "PV-PROC-001", results[0].DocumentNumber)
}

func TestApplyToQuery_HidesDirectPaymentGRNsViaLinkedPO(t *testing.T) {
	db := setupScopeTestDB(t)
	orgID := uuid.New().String()
	userID := uuid.New().String()
	seedScopeOrg(t, db, orgID)
	seedScopeUser(t, db, userID, "procurement")

	// Seed two POs with different routing types.
	directPO := models.PurchaseOrder{
		ID: uuid.New().String(), OrganizationID: orgID,
		DocumentNumber: "PO-DIRECT-GRN-001",
		Status:         "approved",
		RoutingType:    "direct_payment",
		CreatedAt:      time.Now(), UpdatedAt: time.Now(),
	}
	procPO := models.PurchaseOrder{
		ID: uuid.New().String(), OrganizationID: orgID,
		DocumentNumber: "PO-PROC-GRN-001",
		Status:         "approved",
		RoutingType:    "procurement",
		CreatedAt:      time.Now(), UpdatedAt: time.Now(),
	}
	require.NoError(t, db.Create(&directPO).Error)
	require.NoError(t, db.Create(&procPO).Error)

	// GRNs linked to each PO.
	directGRN := models.GoodsReceivedNote{
		ID: uuid.New().String(), OrganizationID: orgID,
		DocumentNumber:   "GRN-DIRECT-001",
		PODocumentNumber: directPO.DocumentNumber,
		ReceivedBy:       userID, Status: "draft",
		ReceivedDate: time.Now(), CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	procGRN := models.GoodsReceivedNote{
		ID: uuid.New().String(), OrganizationID: orgID,
		DocumentNumber:   "GRN-PROC-001",
		PODocumentNumber: procPO.DocumentNumber,
		ReceivedBy:       userID, Status: "draft",
		ReceivedDate: time.Now(), CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	require.NoError(t, db.Create(&directGRN).Error)
	require.NoError(t, db.Create(&procGRN).Error)

	scope := makeHidingScope(userID, orgID, "procurement")

	var results []models.GoodsReceivedNote
	q := db.Where("organization_id = ?", orgID)
	q = scope.ApplyToQuery(q, "created_by", "grn", "received_by")
	require.NoError(t, q.Find(&results).Error)

	assert.Len(t, results, 1, "only the GRN linked to a procurement PO should be visible")
	assert.Equal(t, "GRN-PROC-001", results[0].DocumentNumber)
}
