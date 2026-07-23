package services

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
// Shared test helpers
// ─────────────────────────────────────────────────────────────────────────────

func setupPaymentSummaryTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err, "open in-memory SQLite")

	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)

	require.NoError(t, db.AutoMigrate(&models.PurchaseOrder{}, &models.PaymentVoucher{}), "auto-migrate PO+PV")

	return db
}

func seedSummaryPV(t *testing.T, db *gorm.DB, orgID, poDocNumber, status string, amount float64) models.PaymentVoucher {
	t.Helper()
	pv := models.PaymentVoucher{
		ID:             uuid.New().String(),
		OrganizationID: orgID,
		DocumentNumber: "PV-" + uuid.New().String(),
		LinkedPO:       poDocNumber,
		Status:         status,
		Amount:         amount,
		Currency:       "ZMW",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	require.NoError(t, db.Create(&pv).Error, "seed PV")
	return pv
}

// ─────────────────────────────────────────────────────────────────────────────
// ComputePOPaymentSummary
// ─────────────────────────────────────────────────────────────────────────────

func TestComputePOPaymentSummary_MixedStatuses(t *testing.T) {
	db := setupPaymentSummaryTestDB(t)
	orgID := "org-1"
	poDocNumber := "PO-1"

	seedSummaryPV(t, db, orgID, poDocNumber, "PAID", 300)
	seedSummaryPV(t, db, orgID, poDocNumber, "DRAFT", 200)
	seedSummaryPV(t, db, orgID, poDocNumber, "REJECTED", 500)
	seedSummaryPV(t, db, orgID, poDocNumber, "CANCELLED", 400)
	seedSummaryPV(t, db, orgID, poDocNumber, "COMPLETED", 100)

	summary, err := ComputePOPaymentSummary(db, orgID, poDocNumber)
	require.NoError(t, err)

	assert.Equal(t, 600.0, summary.Committed, "Committed excludes CANCELLED+REJECTED: PAID 300 + DRAFT 200 + COMPLETED 100")
	assert.Equal(t, 400.0, summary.Paid, "Paid counts PAID+COMPLETED only: 300 + 100")
	assert.Equal(t, int64(3), summary.LivePVs, "LivePVs counts non-CANCELLED/REJECTED rows: PAID, DRAFT, COMPLETED")
}

func TestComputePOPaymentSummary_EmptyPO(t *testing.T) {
	db := setupPaymentSummaryTestDB(t)

	summary, err := ComputePOPaymentSummary(db, "org-1", "PO-DOES-NOT-EXIST")
	require.NoError(t, err)

	assert.Equal(t, 0.0, summary.Committed)
	assert.Equal(t, 0.0, summary.Paid)
	assert.Equal(t, int64(0), summary.LivePVs)
}

func TestComputePOPaymentSummary_ScopedByOrgAndPO(t *testing.T) {
	db := setupPaymentSummaryTestDB(t)

	seedSummaryPV(t, db, "org-1", "PO-1", "PAID", 100)
	seedSummaryPV(t, db, "org-2", "PO-1", "PAID", 999) // same PO doc number, different org
	seedSummaryPV(t, db, "org-1", "PO-2", "PAID", 999) // same org, different PO

	summary, err := ComputePOPaymentSummary(db, "org-1", "PO-1")
	require.NoError(t, err)

	assert.Equal(t, 100.0, summary.Committed)
	assert.Equal(t, 100.0, summary.Paid)
	assert.Equal(t, int64(1), summary.LivePVs)
}

func TestComputePOPaymentSummary_AllTerminalFailures(t *testing.T) {
	db := setupPaymentSummaryTestDB(t)
	orgID := "org-1"
	poDocNumber := "PO-1"

	seedSummaryPV(t, db, orgID, poDocNumber, "REJECTED", 500)
	seedSummaryPV(t, db, orgID, poDocNumber, "CANCELLED", 400)

	summary, err := ComputePOPaymentSummary(db, orgID, poDocNumber)
	require.NoError(t, err)

	assert.Equal(t, 0.0, summary.Committed)
	assert.Equal(t, 0.0, summary.Paid)
	assert.Equal(t, int64(0), summary.LivePVs)
}

// ─────────────────────────────────────────────────────────────────────────────
// DerivePaymentStatus
// ─────────────────────────────────────────────────────────────────────────────

func TestDerivePaymentStatus(t *testing.T) {
	cases := []struct {
		name     string
		poTotal  float64
		paid     float64
		expected string
	}{
		{"zero paid -> unpaid", 1000, 0, "unpaid"},
		{"paid within epsilon of zero -> unpaid", 1000, 0.01, "unpaid"},
		{"partial payment -> partially_paid", 1000, 500, "partially_paid"},
		{"just below fully-paid epsilon boundary -> partially_paid", 1000, 999.98, "partially_paid"},
		{"exactly at fully-paid epsilon boundary -> fully_paid", 1000, 999.99, "fully_paid"},
		{"exact full payment -> fully_paid", 1000, 1000, "fully_paid"},
		{"overpayment -> fully_paid", 1000, 1000.5, "fully_paid"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := DerivePaymentStatus(tc.poTotal, tc.paid)
			assert.Equal(t, tc.expected, got)
		})
	}
}
