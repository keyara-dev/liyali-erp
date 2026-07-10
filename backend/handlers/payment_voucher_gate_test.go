package handlers

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/models"
)

func seedPO(t *testing.T, docNum, status, flow string, total float64) models.PurchaseOrder {
	t.Helper()
	po := models.PurchaseOrder{
		ID:              uuid.New().String(),
		OrganizationID:  testOrgID,
		DocumentNumber:  docNum,
		Status:          status,
		ProcurementFlow: flow,
		TotalAmount:     total,
		Currency:        "ZMW",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	if err := config.DB.Create(&po).Error; err != nil {
		t.Fatalf("seed PO: %v", err)
	}
	return po
}

func TestValidateProcurementPVGate(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	t.Run("PO not approved -> 400", func(t *testing.T) {
		po := seedPO(t, "PO-GATE-1", "DRAFT", "payment_first", 1000)
		msg, code := validateProcurementPVGate(db, testOrgID, po.DocumentNumber, "", 500)
		if code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d (%s)", code, msg)
		}
	})

	// B2: the old "one live PV per PO" 409 is gone — a second PV is allowed as
	// long as it fits within the PO's remaining balance (PO total minus what's
	// already committed by live PVs). Exceeding-remaining-balance behavior is
	// covered in payment_voucher_balance_test.go.
	t.Run("second live PV within remaining balance -> ok", func(t *testing.T) {
		po := seedPO(t, "PO-GATE-2", "APPROVED", "payment_first", 1000)
		pv := models.PaymentVoucher{
			ID:             uuid.New().String(),
			OrganizationID: testOrgID,
			DocumentNumber: "PV-GATE-2",
			LinkedPO:       po.DocumentNumber,
			Status:         "APPROVED",
			Amount:         500,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}
		if err := db.Create(&pv).Error; err != nil {
			t.Fatalf("seed PV: %v", err)
		}
		// Committed so far = 500; remaining balance = 500. Requesting exactly the
		// remaining balance must succeed now that multiple PVs per PO are allowed.
		msg, code := validateProcurementPVGate(db, testOrgID, po.DocumentNumber, "", 500)
		if code != 0 {
			t.Fatalf("expected ok (0), got %d (%s)", code, msg)
		}
	})

	t.Run("amount exceeds PO total -> 400", func(t *testing.T) {
		po := seedPO(t, "PO-GATE-3", "APPROVED", "payment_first", 1000)
		msg, code := validateProcurementPVGate(db, testOrgID, po.DocumentNumber, "", 2000)
		if code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d (%s)", code, msg)
		}
	})

	t.Run("valid payment_first -> ok", func(t *testing.T) {
		po := seedPO(t, "PO-GATE-4", "APPROVED", "payment_first", 1000)
		msg, code := validateProcurementPVGate(db, testOrgID, po.DocumentNumber, "", 750)
		if code != 0 {
			t.Fatalf("expected ok (0), got %d (%s)", code, msg)
		}
	})

	t.Run("no linked PO -> ok (manual non-PO PV)", func(t *testing.T) {
		msg, code := validateProcurementPVGate(db, testOrgID, "", "", 999)
		if code != 0 {
			t.Fatalf("expected ok (0), got %d (%s)", code, msg)
		}
	})
}
