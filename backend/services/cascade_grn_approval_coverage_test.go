package services

import (
	"context"
	"testing"

	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ─────────────────────────────────────────────────────────────────────────────
// cascadeGRNApprovalToPO — full-coverage completion gate (delivery-side mirror
// of Task B6 / CascadePVPaidToPO)
// ─────────────────────────────────────────────────────────────────────────────
//
// seedApprovedPOWithItems always sets TotalAmount = 999.0 regardless of the
// item list (see its definition), so PV amounts below are chosen to land
// below/at that fixed total.

func TestGRNApprovalCascadesToPO_DepositOnly_FullDeliveryLeavesFulfilledNotCompleted(t *testing.T) {
	db := setupExecutionTestDB(t)
	const orgID = "org-cascade-grn-cov-1"
	seedOrg(t, db, orgID)

	items := []types.POItem{{Description: "Widget", Quantity: 10, UnitPrice: 5.0, Amount: 50.0}}
	po := seedApprovedPOWithItems(t, db, orgID, items)

	// Only a deposit has been paid — nowhere near the PO's 999 total.
	seedSummaryPV(t, db, orgID, po.DocumentNumber, "PAID", 500)

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
	require.NoError(t, svc.ApproveWorkflowTaskWithVersion(context.Background(), taskID, approverID, "LGTM", "", 1))

	var updatedPO models.PurchaseOrder
	require.NoError(t, db.First(&updatedPO, "id = ?", po.ID).Error)
	assert.Equal(t, models.DeliveryStatusFullyDelivered, updatedPO.DeliveryStatus)
	assert.Equal(t, models.StatusFulfilled, updatedPO.Status,
		"a fully-delivered PO with only a deposit paid must go FULFILLED, not COMPLETED")
}

func TestGRNApprovalCascadesToPO_FullCoverage_FullDeliveryCompletes(t *testing.T) {
	db := setupExecutionTestDB(t)
	const orgID = "org-cascade-grn-cov-2"
	seedOrg(t, db, orgID)

	items := []types.POItem{{Description: "Widget", Quantity: 10, UnitPrice: 5.0, Amount: 50.0}}
	po := seedApprovedPOWithItems(t, db, orgID, items)

	// Two PVs together cover the full 999 total.
	seedSummaryPV(t, db, orgID, po.DocumentNumber, "PAID", 500)
	seedSummaryPV(t, db, orgID, po.DocumentNumber, "PAID", 499)

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
	require.NoError(t, svc.ApproveWorkflowTaskWithVersion(context.Background(), taskID, approverID, "LGTM", "", 1))

	var updatedPO models.PurchaseOrder
	require.NoError(t, db.First(&updatedPO, "id = ?", po.ID).Error)
	assert.Equal(t, models.DeliveryStatusFullyDelivered, updatedPO.DeliveryStatus)
	assert.Equal(t, models.StatusCompleted, updatedPO.Status,
		"a fully-delivered PO whose live PVs fully cover the total must go COMPLETED")
}

func TestGRNApprovalCascadesToPO_RejectedPVDoesNotBlockCompletion(t *testing.T) {
	db := setupExecutionTestDB(t)
	const orgID = "org-cascade-grn-cov-3"
	seedOrg(t, db, orgID)

	items := []types.POItem{{Description: "Widget", Quantity: 10, UnitPrice: 5.0, Amount: 50.0}}
	po := seedApprovedPOWithItems(t, db, orgID, items)

	seedSummaryPV(t, db, orgID, po.DocumentNumber, "PAID", 500)
	seedSummaryPV(t, db, orgID, po.DocumentNumber, "PAID", 499)
	seedSummaryPV(t, db, orgID, po.DocumentNumber, "REJECTED", 9999)

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
	require.NoError(t, svc.ApproveWorkflowTaskWithVersion(context.Background(), taskID, approverID, "LGTM", "", 1))

	var updatedPO models.PurchaseOrder
	require.NoError(t, db.First(&updatedPO, "id = ?", po.ID).Error)
	assert.Equal(t, models.StatusCompleted, updatedPO.Status,
		"a REJECTED PV must neither count toward coverage nor block completion once live PVs cover the total")
}

// ─────────────────────────────────────────────────────────────────────────────
// Fix 2 end-to-end invariant: deposit -> full delivery (FULFILLED) -> balance
// PV -> both paid -> COMPLETED. Exercises both cascades together: the
// delivery-side cascade (Fix 1, this file) must park the PO at FULFILLED
// rather than deadlocking it at APPROVED-only semantics, and the payment-side
// cascade (CascadePVPaidToPO, Task B6) must accept FULFILLED as a source
// state once the balance PV is also paid.
// ─────────────────────────────────────────────────────────────────────────────

func TestPartialPaymentLifecycle_DepositThenFullDeliveryThenBalancePV_Completes(t *testing.T) {
	db := setupExecutionTestDB(t)
	const orgID = "org-cascade-lifecycle-1"
	seedOrg(t, db, orgID)

	items := []types.POItem{{Description: "Widget", Quantity: 10, UnitPrice: 5.0, Amount: 50.0}}
	po := seedApprovedPOWithItems(t, db, orgID, items) // TotalAmount fixed at 999

	deposit := seedSummaryPV(t, db, orgID, po.DocumentNumber, "PAID", 500)

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
	require.NoError(t, svc.ApproveWorkflowTaskWithVersion(context.Background(), taskID, approverID, "full delivery", "", 1))

	var afterDelivery models.PurchaseOrder
	require.NoError(t, db.First(&afterDelivery, "id = ?", po.ID).Error)
	require.Equal(t, models.StatusFulfilled, afterDelivery.Status,
		"precondition: full delivery with only a deposit paid must park the PO at FULFILLED")
	_ = deposit

	// Balance PV created and paid while the PO sits at FULFILLED — this is
	// exactly what Fix 2's gates (validateProcurementPVGate / SubmitPaymentVoucher)
	// now allow instead of rejecting with "must be APPROVED".
	balance := seedSummaryPV(t, db, orgID, po.DocumentNumber, "PAID", 499)
	require.NoError(t, svc.CascadePVPaidToPO(db, balance.ID))

	var final models.PurchaseOrder
	require.NoError(t, db.First(&final, "id = ?", po.ID).Error)
	assert.Equal(t, models.StatusCompleted, final.Status,
		"once the balance PV is paid on top of the deposit, the FULFILLED PO must complete")
}
