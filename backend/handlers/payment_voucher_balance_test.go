package handlers

// payment_voucher_balance_test.go — Task B2: the single-PV-per-PO gate became
// a remaining-balance guard, so a PO can now carry multiple *live* partial
// PVs as long as their sum never exceeds the PO total (or, in goods-first,
// the received value). These tests exercise both the gate function directly
// and the full HTTP create-from-PO path. seedPO / fromPOApp are shared
// helpers defined in payment_voucher_gate_test.go / payment_voucher_from_po_test.go.

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/models"
)

// seedLivePV inserts a PaymentVoucher linked to poDocNum with the given
// status/amount directly via the test DB (bypassing the gate/handler), so
// tests can set up a "prior committed PV" fixture.
func seedLivePV(t *testing.T, docNum, poDocNum, status string, amount float64) models.PaymentVoucher {
	t.Helper()
	pv := models.PaymentVoucher{
		ID:             uuid.New().String(),
		OrganizationID: testOrgID,
		DocumentNumber: docNum,
		LinkedPO:       poDocNum,
		Status:         status,
		Amount:         amount,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	if err := config.DB.Create(&pv).Error; err != nil {
		t.Fatalf("seed PV %s: %v", docNum, err)
	}
	return pv
}

func TestValidateProcurementPVGate_RemainingBalance(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	t.Run("PO 1000, live 600, new 400 -> ok", func(t *testing.T) {
		po := seedPO(t, "PO-BAL-1", "APPROVED", "payment_first", 1000)
		seedLivePV(t, "PV-BAL-1A", po.DocumentNumber, "APPROVED", 600)

		msg, code := validateProcurementPVGate(db, testOrgID, po.DocumentNumber, "", 400)
		if code != 0 {
			t.Fatalf("expected ok (0), got %d (%s)", code, msg)
		}
	})

	t.Run("PO 1000, live 600, new 500 -> 400 remaining balance", func(t *testing.T) {
		po := seedPO(t, "PO-BAL-2", "APPROVED", "payment_first", 1000)
		seedLivePV(t, "PV-BAL-2A", po.DocumentNumber, "APPROVED", 600)

		msg, code := validateProcurementPVGate(db, testOrgID, po.DocumentNumber, "", 500)
		if code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d (%s)", code, msg)
		}
		if !strings.Contains(msg, "remaining balance") {
			t.Fatalf("expected message to mention remaining balance, got: %s", msg)
		}
	})

	t.Run("REJECTED PV 1000 does not consume budget -> retry with new PV 1000 ok", func(t *testing.T) {
		po := seedPO(t, "PO-BAL-3", "APPROVED", "payment_first", 1000)
		seedLivePV(t, "PV-BAL-3A", po.DocumentNumber, "REJECTED", 1000)

		msg, code := validateProcurementPVGate(db, testOrgID, po.DocumentNumber, "", 1000)
		if code != 0 {
			t.Fatalf("expected ok (0), got %d (%s)", code, msg)
		}
	})

	t.Run("CANCELLED PV 1000 does not consume budget -> retry with new PV 1000 ok", func(t *testing.T) {
		po := seedPO(t, "PO-BAL-4", "APPROVED", "payment_first", 1000)
		seedLivePV(t, "PV-BAL-4A", po.DocumentNumber, "CANCELLED", 1000)

		msg, code := validateProcurementPVGate(db, testOrgID, po.DocumentNumber, "", 1000)
		if code != 0 {
			t.Fatalf("expected ok (0), got %d (%s)", code, msg)
		}
	})
}

// TestCreatePVFromPO_MultiplePartialPVs drives the full HTTP handler through
// two successive partial PVs that together exactly exhaust the PO's balance
// (60% then 40%), then asserts a third PV of any positive amount is rejected.
func TestCreatePVFromPO_MultiplePartialPVs(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	po := seedPO(t, "PO-BAL-HTTP-1", "APPROVED", "payment_first", 1000)
	app := fromPOApp()

	// First partial PV: 60% of the PO total.
	resp1 := testRequest(app, http.MethodPost, "/payment-vouchers/from-po", map[string]interface{}{
		"purchaseOrderId":             po.ID,
		"purchaseOrderDocumentNumber": po.DocumentNumber,
		"totalAmount":                 600.0,
	})
	if resp1.StatusCode != http.StatusCreated {
		t.Fatalf("first partial PV (60%%): expected 201, got %d", resp1.StatusCode)
	}

	// Second partial PV: the remaining 40%.
	resp2 := testRequest(app, http.MethodPost, "/payment-vouchers/from-po", map[string]interface{}{
		"purchaseOrderId":             po.ID,
		"purchaseOrderDocumentNumber": po.DocumentNumber,
		"totalAmount":                 400.0,
	})
	if resp2.StatusCode != http.StatusCreated {
		t.Fatalf("second partial PV (40%%): expected 201, got %d", resp2.StatusCode)
	}

	// Third PV of any positive amount: the PO's balance is now fully
	// committed (600 + 400 = 1000), so this must be rejected.
	resp3 := testRequest(app, http.MethodPost, "/payment-vouchers/from-po", map[string]interface{}{
		"purchaseOrderId":             po.ID,
		"purchaseOrderDocumentNumber": po.DocumentNumber,
		"totalAmount":                 50.0,
	})
	if resp3.StatusCode != http.StatusBadRequest {
		t.Fatalf("third PV over exhausted balance: expected 400, got %d", resp3.StatusCode)
	}
}
