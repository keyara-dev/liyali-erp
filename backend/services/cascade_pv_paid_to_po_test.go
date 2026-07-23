package services

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// ─────────────────────────────────────────────────────────────────────────────
// CascadePVPaidToPO — full-coverage completion gate (Task B6)
// ─────────────────────────────────────────────────────────────────────────────

// seedFullyDeliveredPO creates an APPROVED PurchaseOrder, already FULLY_DELIVERED,
// with the given total — the state CascadePVPaidToPO expects before it will even
// consider flipping the PO to COMPLETED.
func seedFullyDeliveredPO(t *testing.T, db *gorm.DB, orgID string, total float64) models.PurchaseOrder {
	t.Helper()
	po := models.PurchaseOrder{
		ID:             uuid.New().String(),
		OrganizationID: orgID,
		DocumentNumber: "PO-" + uuid.New().String()[:8],
		Status:         models.StatusApproved,
		Currency:       "ZMW",
		TotalAmount:    total,
		DeliveryStatus: models.DeliveryStatusFullyDelivered,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	po.Items = datatypes.NewJSONType([]types.POItem{})
	po.ApprovalHistory = datatypes.NewJSONType([]types.ApprovalRecord{})
	require.NoError(t, db.Create(&po).Error)
	return po
}

func TestCascadePVPaidToPO_DepositOnly_PODoesNotComplete(t *testing.T) {
	db := setupExecutionTestDB(t)
	const orgID = "org-cascade-cov-1"
	seedOrg(t, db, orgID)

	po := seedFullyDeliveredPO(t, db, orgID, 1000)
	pv := seedSummaryPV(t, db, orgID, po.DocumentNumber, "PAID", 500) // 50% deposit only

	svc := newExecutionService(t, db)
	require.NoError(t, svc.CascadePVPaidToPO(db, pv.ID))

	var updated models.PurchaseOrder
	require.NoError(t, db.First(&updated, "id = ?", po.ID).Error)
	assert.Equal(t, models.StatusApproved, updated.Status,
		"PO must stay open when the only paid PV is a deposit that doesn't cover the full total")
}

func TestCascadePVPaidToPO_FullCoverage_CompletesPO(t *testing.T) {
	db := setupExecutionTestDB(t)
	const orgID = "org-cascade-cov-2"
	seedOrg(t, db, orgID)

	po := seedFullyDeliveredPO(t, db, orgID, 1000)
	seedSummaryPV(t, db, orgID, po.DocumentNumber, "PAID", 500)
	pv2 := seedSummaryPV(t, db, orgID, po.DocumentNumber, "PAID", 500) // now covers the full 1000

	svc := newExecutionService(t, db)
	require.NoError(t, svc.CascadePVPaidToPO(db, pv2.ID))

	var updated models.PurchaseOrder
	require.NoError(t, db.First(&updated, "id = ?", po.ID).Error)
	assert.Equal(t, models.StatusCompleted, updated.Status,
		"PO must complete once its linked PVs fully cover the total")
}

func TestCascadePVPaidToPO_RejectedPVDoesNotBlockCompletion(t *testing.T) {
	db := setupExecutionTestDB(t)
	const orgID = "org-cascade-cov-3"
	seedOrg(t, db, orgID)

	po := seedFullyDeliveredPO(t, db, orgID, 1000)
	seedSummaryPV(t, db, orgID, po.DocumentNumber, "PAID", 500)
	pv2 := seedSummaryPV(t, db, orgID, po.DocumentNumber, "PAID", 500)
	seedSummaryPV(t, db, orgID, po.DocumentNumber, "REJECTED", 9999) // must not block completion

	svc := newExecutionService(t, db)
	require.NoError(t, svc.CascadePVPaidToPO(db, pv2.ID))

	var updated models.PurchaseOrder
	require.NoError(t, db.First(&updated, "id = ?", po.ID).Error)
	assert.Equal(t, models.StatusCompleted, updated.Status,
		"a REJECTED PV must not block PO completion once live PVs fully cover the total")
}
