package handlers

// payment_voucher_goods_first_received_value_test.go — Fix 3: the goods-first
// over-invoicing guard in validateProcurementPVGate must compute "received
// value" cumulatively across every live GRN linked to the PO, not just the
// single GRN passed as linkedGRN. The old single-GRN computation meant a
// second (or later) delivery's PV was always blocked, because that GRN's own
// received value had usually already been fully "committed" by an earlier
// PV against a *different* GRN on the same PO.
//
// REJECTED GRNs are excluded from the cumulative received value (alongside
// CANCELLED) — a rejected delivery was never accepted, so it must not count
// as goods received. TestValidateProcurementPVGate_GoodsFirst_RejectedGRNNotCounted
// locks in that decision.

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/types"
	"gorm.io/datatypes"
)

// seedGoodsFirstPO creates a goods_first APPROVED PO with a single line item
// ("Widget A") at the given quantity/unit price.
func seedGoodsFirstPO(t *testing.T, docNum string, qty int, unitPrice, poTotal float64) models.PurchaseOrder {
	t.Helper()
	po := models.PurchaseOrder{
		ID:              uuid.New().String(),
		OrganizationID:  testOrgID,
		DocumentNumber:  docNum,
		Status:          "APPROVED",
		ProcurementFlow: "goods_first",
		TotalAmount:     poTotal,
		Currency:        "ZMW",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	po.Items = datatypes.NewJSONType([]types.POItem{{
		Description: "Widget A", Quantity: qty, UnitPrice: unitPrice, Amount: float64(qty) * unitPrice,
	}})
	if err := config.DB.Create(&po).Error; err != nil {
		t.Fatalf("seed goods-first PO: %v", err)
	}
	return po
}

// seedGRNForPO creates a GoodsReceivedNote linked to po with a single
// "Widget A" line, defaulting to APPROVED status.
func seedGRNForPO(t *testing.T, docNum string, po models.PurchaseOrder, qtyOrdered, qtyReceived int, status string) models.GoodsReceivedNote {
	t.Helper()
	grn := models.GoodsReceivedNote{
		ID:               uuid.New().String(),
		OrganizationID:   testOrgID,
		DocumentNumber:   docNum,
		PODocumentNumber: po.DocumentNumber,
		Status:           status,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
	grn.Items = datatypes.NewJSONType([]types.GRNItem{{
		Description: "Widget A", QuantityOrdered: qtyOrdered, QuantityReceived: qtyReceived, Condition: "good",
	}})
	if err := config.DB.Create(&grn).Error; err != nil {
		t.Fatalf("seed GRN: %v", err)
	}
	return grn
}

func TestValidateProcurementPVGate_GoodsFirst_FirstDeliveryPV_OK(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	po := seedGoodsFirstPO(t, "PO-GF-1", 20, 50, 1000)
	grn1 := seedGRNForPO(t, "GRN-GF-1", po, 10, 10, "APPROVED") // received value 500

	msg, code := validateProcurementPVGate(db, testOrgID, po.DocumentNumber, grn1.DocumentNumber, 500)
	if code != 0 {
		t.Fatalf("expected ok (0) for PV against first delivery, got %d (%s)", code, msg)
	}
}

// The key regression test: a second delivery's PV must not be blocked just
// because the second GRN's own received value was already "committed" by the
// first PV. Before the fix, receivedValue was computed only from the single
// linkedGRN (500), so remaining = 500 - committed(500, PO-wide) = 0, and the
// second PV was always rejected.
func TestValidateProcurementPVGate_GoodsFirst_SecondDeliveryPV_OK(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	po := seedGoodsFirstPO(t, "PO-GF-2", 20, 50, 1000)
	grn1 := seedGRNForPO(t, "GRN-GF-2A", po, 10, 10, "APPROVED") // 500 received
	seedLivePV(t, "PV-GF-2A", po.DocumentNumber, "APPROVED", 500)
	grn2 := seedGRNForPO(t, "GRN-GF-2B", po, 10, 10, "APPROVED") // cumulative 1000 received

	msg, code := validateProcurementPVGate(db, testOrgID, po.DocumentNumber, grn2.DocumentNumber, 500)
	if code != 0 {
		t.Fatalf("expected ok (0) for PV against second delivery, got %d (%s)", code, msg)
	}
	_ = grn1
}

func TestValidateProcurementPVGate_GoodsFirst_OverInvoicing_Blocked(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	po := seedGoodsFirstPO(t, "PO-GF-3", 20, 50, 1000)
	grn1 := seedGRNForPO(t, "GRN-GF-3", po, 10, 10, "APPROVED") // 500 received

	msg, code := validateProcurementPVGate(db, testOrgID, po.DocumentNumber, grn1.DocumentNumber, 600)
	if code != http.StatusBadRequest {
		t.Fatalf("expected 400 for PV over received value, got %d (%s)", code, msg)
	}
	if !strings.Contains(msg, "remaining received value") {
		t.Fatalf("expected message to mention remaining received value, got: %s", msg)
	}
}

func TestValidateProcurementPVGate_GoodsFirst_RejectedGRNNotCounted(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	po := seedGoodsFirstPO(t, "PO-GF-4", 20, 50, 1000)
	grn1 := seedGRNForPO(t, "GRN-GF-4A", po, 10, 10, "APPROVED")  // 500 received, valid
	seedGRNForPO(t, "GRN-GF-4B", po, 10, 10, "REJECTED")          // another 500, but REJECTED

	// If the REJECTED GRN wrongly counted, cumulative received value would be
	// 1000 and this 600 PV would pass. It must still be blocked because only
	// the 500 from the APPROVED GRN counts.
	msg, code := validateProcurementPVGate(db, testOrgID, po.DocumentNumber, grn1.DocumentNumber, 600)
	if code != http.StatusBadRequest {
		t.Fatalf("expected 400: REJECTED GRN must not count toward received value, got %d (%s)", code, msg)
	}
}
