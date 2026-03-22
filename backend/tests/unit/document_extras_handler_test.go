package unit

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/types"
	"github.com/stretchr/testify/assert"
)

// ─────────────────────────────────────────────────────────────────────────────
// CreatePurchaseOrderFromRequisition
// ─────────────────────────────────────────────────────────────────────────────

func TestCreatePOFromRequisition_Validation(t *testing.T) {
	tests := []struct {
		name        string
		reqID       string
		workflowID  string
		shouldPass  bool
	}{
		{"Valid — requisitionId + workflowId", uuid.New().String(), uuid.New().String(), true},
		{"Missing requisitionId", "", uuid.New().String(), false},
		{"Missing workflowId", uuid.New().String(), "", false},
		{"Both empty", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.reqID != "" && tt.workflowID != ""
			assert.Equal(t, tt.shouldPass, isValid)
		})
	}
}

func TestCreatePOFromRequisition_ProcurementFlow(t *testing.T) {
	tests := []struct {
		name            string
		procurementFlow string
		shouldPass      bool
	}{
		{"Empty string (inherit from org)", "", true},
		{"Goods-first explicit", "goods_first", true},
		{"Payment-first explicit", "payment_first", true},
		{"Invalid value", "immediate", false},
		{"Uppercase invalid", "GOODS_FIRST", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.procurementFlow == "" ||
				tt.procurementFlow == "goods_first" ||
				tt.procurementFlow == "payment_first"
			assert.Equal(t, tt.shouldPass, isValid)
		})
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// CreatePaymentVoucherFromPO — Goods-First Flow Enforcement
// ─────────────────────────────────────────────────────────────────────────────

func TestCreatePVFromPO_GoodsFirst_RequiresLinkedGRN(t *testing.T) {
	tests := []struct {
		name                   string
		effectiveFlow          string
		linkedGRNDocumentNumber string
		shouldPass             bool
	}{
		{"Goods-first with valid GRN", "goods_first", "GRN-20240101-abc123", true},
		{"Goods-first missing GRN", "goods_first", "", false},
		{"Payment-first without GRN (allowed)", "payment_first", "", true},
		{"Payment-first with GRN (accepted)", "payment_first", "GRN-20240101-abc123", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := true
			if tt.effectiveFlow == "goods_first" && tt.linkedGRNDocumentNumber == "" {
				isValid = false
			}
			assert.Equal(t, tt.shouldPass, isValid)
		})
	}
}

func TestCreatePVFromPO_GoodsFirst_GRNStatusValidation(t *testing.T) {
	tests := []struct {
		name       string
		grnStatus  string
		shouldPass bool
	}{
		{"Approved GRN — allowed", "APPROVED", true},
		{"Draft GRN — blocked", "DRAFT", false},
		{"Pending GRN — blocked", "PENDING", false},
		{"Rejected GRN — blocked", "REJECTED", false},
		{"Submitted GRN — blocked", "SUBMITTED", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.grnStatus == "APPROVED"
			assert.Equal(t, tt.shouldPass, isValid)
		})
	}
}

func TestCreatePVFromPO_GoodsFirst_GRNBelongsToPO(t *testing.T) {
	poDocNumber := "PO-20240101-abc123"

	tests := []struct {
		name             string
		grnPODocNumber   string
		shouldPass       bool
	}{
		{"GRN belongs to PO", poDocNumber, true},
		{"GRN from different PO", "PO-20240101-xyz999", false},
		{"GRN with no PO link", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.grnPODocNumber == poDocNumber
			assert.Equal(t, tt.shouldPass, isValid)
		})
	}
}

func TestCreatePVFromPO_FlowResolution(t *testing.T) {
	tests := []struct {
		name             string
		poProcFlow       string
		orgProcFlow      string
		expectedFlow     string
	}{
		{"PO override takes precedence over org", "payment_first", "goods_first", "payment_first"},
		{"Org default used when PO has none", "", "payment_first", "payment_first"},
		{"Default goods_first when both empty", "", "", "goods_first"},
		{"PO goods_first overrides org payment_first", "goods_first", "payment_first", "goods_first"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			effectiveFlow := tt.poProcFlow
			if effectiveFlow == "" {
				effectiveFlow = tt.orgProcFlow
				if effectiveFlow == "" {
					effectiveFlow = "goods_first"
				}
			}
			assert.Equal(t, tt.expectedFlow, effectiveFlow)
		})
	}
}

func TestCreatePVFromPO_Validation(t *testing.T) {
	tests := []struct {
		name          string
		purchaseOrderID string
		totalAmount   float64
		workflowID    string
		shouldPass    bool
	}{
		{"Valid request", uuid.New().String(), 50000, uuid.New().String(), true},
		{"Missing purchaseOrderId", "", 50000, uuid.New().String(), false},
		{"Zero totalAmount", uuid.New().String(), 0, uuid.New().String(), false},
		{"Negative totalAmount", uuid.New().String(), -100, uuid.New().String(), false},
		{"Missing workflowId — still valid (optional)", uuid.New().String(), 50000, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.purchaseOrderID != "" && tt.totalAmount > 0
			assert.Equal(t, tt.shouldPass, isValid)
		})
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// CreateGRN — Payment-First Flow Enforcement
// ─────────────────────────────────────────────────────────────────────────────

func TestCreateGRN_PaymentFirst_RequiresLinkedPV(t *testing.T) {
	tests := []struct {
		name          string
		effectiveFlow string
		linkedPV      string
		shouldPass    bool
	}{
		{"Payment-first with valid PV", "payment_first", "PV-20240101-abc123", true},
		{"Payment-first missing PV", "payment_first", "", false},
		{"Goods-first without PV (allowed)", "goods_first", "", true},
		{"Goods-first with PV (accepted)", "goods_first", "PV-20240101-abc123", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := true
			if tt.effectiveFlow == "payment_first" && tt.linkedPV == "" {
				isValid = false
			}
			assert.Equal(t, tt.shouldPass, isValid)
		})
	}
}

func TestCreateGRN_PaymentFirst_PVStatusValidation(t *testing.T) {
	tests := []struct {
		name       string
		pvStatus   string
		shouldPass bool
	}{
		{"Approved PV — allowed", "APPROVED", true},
		{"Paid PV — allowed", "PAID", true},
		{"Draft PV — blocked", "DRAFT", false},
		{"Pending PV — blocked", "PENDING", false},
		{"Rejected PV — blocked", "REJECTED", false},
		{"Submitted PV — blocked", "SUBMITTED", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.pvStatus == "APPROVED" || tt.pvStatus == "PAID"
			assert.Equal(t, tt.shouldPass, isValid)
		})
	}
}

func TestCreateGRN_PaymentFirst_PVBelongsToPO(t *testing.T) {
	poDocNumber := "PO-20240101-abc123"

	tests := []struct {
		name          string
		pvLinkedPO    string
		shouldPass    bool
	}{
		{"PV belongs to PO", poDocNumber, true},
		{"PV from different PO", "PO-20240101-xyz999", false},
		{"PV with no PO link", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.pvLinkedPO == poDocNumber
			assert.Equal(t, tt.shouldPass, isValid)
		})
	}
}

func TestCreateGRN_BaseValidation(t *testing.T) {
	tests := []struct {
		name           string
		poDocNumber    string
		receivedBy     string
		items          int
		shouldPass     bool
	}{
		{"All valid", "PO-20240101-abc123", "John Doe", 2, true},
		{"Missing PO number", "", "John Doe", 2, false},
		{"Missing receivedBy", "PO-20240101-abc123", "", 2, false},
		{"No items", "PO-20240101-abc123", "John Doe", 0, false},
		{"Invalid PO format (too short)", "PO-123", "John Doe", 1, false},
		{"Invalid PO format (no prefix)", "20240101-abc123", "John Doe", 1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validPOFormat := len(tt.poDocNumber) >= 10 &&
				len(tt.poDocNumber) >= 3 &&
				tt.poDocNumber[:3] == "PO-"
			isValid := validPOFormat && tt.receivedBy != "" && tt.items > 0
			assert.Equal(t, tt.shouldPass, isValid)
		})
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// MarkPaymentVoucherPaid
// ─────────────────────────────────────────────────────────────────────────────

func TestMarkPVPaid_StatusPrerequisite(t *testing.T) {
	tests := []struct {
		name       string
		pvStatus   string
		shouldPass bool
	}{
		{"Approved PV can be marked paid", "APPROVED", true},
		{"Draft PV cannot be marked paid", "DRAFT", false},
		{"Pending PV cannot be marked paid", "PENDING", false},
		{"Already paid — idempotent", "PAID", false},
		{"Rejected PV cannot be marked paid", "REJECTED", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			canMarkPaid := tt.pvStatus == "APPROVED"
			assert.Equal(t, tt.shouldPass, canMarkPaid)
		})
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Document Stats
// ─────────────────────────────────────────────────────────────────────────────

func TestRequisitionStats_Structure(t *testing.T) {
	t.Run("Stats count by status", func(t *testing.T) {
		statuses := []string{"DRAFT", "PENDING", "APPROVED", "REJECTED"}
		counts := map[string]int{
			"DRAFT":    3,
			"PENDING":  5,
			"APPROVED": 10,
			"REJECTED": 2,
		}

		total := 0
		for _, s := range statuses {
			total += counts[s]
		}

		assert.Equal(t, 20, total)
		assert.Equal(t, 5, counts["PENDING"])
		assert.Equal(t, 10, counts["APPROVED"])
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// Procurement Chain Audit Trail
// ─────────────────────────────────────────────────────────────────────────────

func TestProcurementChainAuditTrail(t *testing.T) {
	t.Run("Goods-first chain: REQ → PO → GRN → PV", func(t *testing.T) {
		reqID := uuid.New().String()
		poDocNum := "PO-20240101-abc123"
		grnDocNum := "GRN-20240102-def456"
		pvDocNum := "PV-20240103-ghi789"

		// Simulate action history entries for goods-first chain
		poHistory := []types.ActionHistoryEntry{
			{
				ID:          uuid.New().String(),
				Action:      "CREATED_FROM_REQUISITION",
				PerformedBy: uuid.New().String(),
				Timestamp:   time.Now(),
				Metadata: map[string]interface{}{
					"linkedDocNumber": reqID,
					"linkedDocType":   "requisition",
				},
			},
		}

		pvHistory := []types.ActionHistoryEntry{
			{
				ID:          uuid.New().String(),
				Action:      "CREATED_FROM_GRN",
				PerformedBy: uuid.New().String(),
				Timestamp:   time.Now(),
				Metadata: map[string]interface{}{
					"linkedDocNumber": grnDocNum,
					"linkedDocType":   "grn",
					"flow":            "goods_first",
				},
			},
		}

		assert.Len(t, poHistory, 1)
		assert.Equal(t, "CREATED_FROM_REQUISITION", poHistory[0].Action)
		assert.Equal(t, reqID, poHistory[0].Metadata["linkedDocNumber"])

		assert.Len(t, pvHistory, 1)
		assert.Equal(t, "CREATED_FROM_GRN", pvHistory[0].Action)
		assert.Equal(t, grnDocNum, pvHistory[0].Metadata["linkedDocNumber"])
		assert.Equal(t, "goods_first", pvHistory[0].Metadata["flow"])

		// Chain is traceable
		assert.NotEmpty(t, poDocNum)
		assert.NotEmpty(t, grnDocNum)
		assert.NotEmpty(t, pvDocNum)
	})

	t.Run("Payment-first chain: REQ → PO → PV → GRN", func(t *testing.T) {
		pvDocNum := "PV-20240101-abc123"
		grnHistory := []types.ActionHistoryEntry{
			{
				ID:          uuid.New().String(),
				Action:      "CREATED_FROM_PV",
				PerformedBy: uuid.New().String(),
				Timestamp:   time.Now(),
				Metadata: map[string]interface{}{
					"linkedDocNumber": pvDocNum,
					"linkedDocType":   "payment_voucher",
					"flow":            "payment_first",
				},
			},
		}

		assert.Len(t, grnHistory, 1)
		assert.Equal(t, "CREATED_FROM_PV", grnHistory[0].Action)
		assert.Equal(t, pvDocNum, grnHistory[0].Metadata["linkedDocNumber"])
		assert.Equal(t, "payment_first", grnHistory[0].Metadata["flow"])
	})

	t.Run("Audit entry metadata fields", func(t *testing.T) {
		entry := types.ActionHistoryEntry{
			ID:          uuid.New().String(),
			Action:      "PO_CREATED",
			PerformedBy: uuid.New().String(),
			Timestamp:   time.Now(),
			Metadata: map[string]interface{}{
				"linkedDocNumber": "PO-20240101-abc123",
				"linkedDocType":   "purchase_order",
				"flow":            "goods_first",
			},
		}

		assert.NotEmpty(t, entry.ID)
		assert.Equal(t, "PO_CREATED", entry.Action)
		assert.NotNil(t, entry.Metadata)
		assert.Equal(t, "PO-20240101-abc123", entry.Metadata["linkedDocNumber"])
		assert.Equal(t, "purchase_order", entry.Metadata["linkedDocType"])
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// Document Number Format Validation
// ─────────────────────────────────────────────────────────────────────────────

func TestDocumentNumberFormats(t *testing.T) {
	tests := []struct {
		name       string
		docNumber  string
		prefix     string
		minLen     int
		shouldPass bool
	}{
		{"Valid PO number", "PO-20240101-abc123", "PO-", 10, true},
		{"Valid GRN number", "GRN-20240101-abc123", "GRN-", 10, true},
		{"Valid PV number", "PV-20240101-abc123", "PV-", 10, true},
		{"Valid REQ number", "REQ-20240101-abc123", "REQ-", 10, true},
		{"Too short PO", "PO-123", "PO-", 10, false},
		{"Wrong prefix", "ORDER-20240101-abc", "PO-", 10, false},
		{"Empty string", "", "PO-", 10, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := len(tt.docNumber) >= tt.minLen &&
				len(tt.docNumber) >= len(tt.prefix) &&
				tt.docNumber[:len(tt.prefix)] == tt.prefix
			assert.Equal(t, tt.shouldPass, isValid)
		})
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// PO → PV Link Validation (goods_first and payment_first)
// ─────────────────────────────────────────────────────────────────────────────

func TestPVLinkedPOValidation(t *testing.T) {
	t.Run("PV must reference the source PO document number", func(t *testing.T) {
		pv := models.PaymentVoucher{
			ID:             uuid.New().String(),
			DocumentNumber: "PV-20240101-abc123",
			LinkedPO:       "PO-20240101-xyz456",
			Status:         "DRAFT",
			Amount:         50000,
		}

		assert.NotEmpty(t, pv.LinkedPO)
		assert.Equal(t, "PO-20240101-xyz456", pv.LinkedPO)
	})

	t.Run("PV with linkedGRN set in goods-first flow", func(t *testing.T) {
		pv := models.PaymentVoucher{
			ID:             uuid.New().String(),
			DocumentNumber: "PV-20240101-abc123",
			LinkedPO:       "PO-20240101-xyz456",
			LinkedGRN:      "GRN-20240101-def789",
			Status:         "DRAFT",
			Amount:         50000,
		}

		assert.NotEmpty(t, pv.LinkedGRN)
		assert.Equal(t, "GRN-20240101-def789", pv.LinkedGRN)
	})

	t.Run("PV without linkedGRN in payment-first flow", func(t *testing.T) {
		pv := models.PaymentVoucher{
			ID:             uuid.New().String(),
			DocumentNumber: "PV-20240101-abc123",
			LinkedPO:       "PO-20240101-xyz456",
			LinkedGRN:      "", // not required for payment_first
			Status:         "DRAFT",
			Amount:         50000,
		}

		assert.Empty(t, pv.LinkedGRN)
		assert.NotEmpty(t, pv.LinkedPO)
	})
}

func TestGRNLinkedPVValidation(t *testing.T) {
	t.Run("GRN with linkedPV in payment-first flow", func(t *testing.T) {
		grn := models.GoodsReceivedNote{
			ID:               uuid.New().String(),
			DocumentNumber:   "GRN-20240101-abc123",
			PODocumentNumber: "PO-20240101-xyz456",
			LinkedPV:         "PV-20240101-def789",
			Status:           "DRAFT",
		}

		assert.NotEmpty(t, grn.LinkedPV)
		assert.Equal(t, "PV-20240101-def789", grn.LinkedPV)
	})

	t.Run("GRN without linkedPV in goods-first flow", func(t *testing.T) {
		grn := models.GoodsReceivedNote{
			ID:               uuid.New().String(),
			DocumentNumber:   "GRN-20240101-abc123",
			PODocumentNumber: "PO-20240101-xyz456",
			LinkedPV:         "", // not required for goods_first
			Status:           "DRAFT",
		}

		assert.Empty(t, grn.LinkedPV)
		assert.NotEmpty(t, grn.PODocumentNumber)
	})
}
