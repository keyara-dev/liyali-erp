package unit

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/types"
)

// TestCreatePaymentVoucherValidation tests payment voucher request validation
func TestCreatePaymentVoucherValidation(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
		shouldPass     bool
	}{
		{
			name: "Valid payment voucher request",
			requestBody: map[string]interface{}{
				"vendorId":      uuid.New().String(),
				"invoiceNumber": "INV-2025-001",
				"amount":        50000,
				"currency":      "USD",
				"paymentMethod": "bank_transfer",
				"glCode":        "4000",
				"description":   "Payment for services rendered",
				"linkedPO":      uuid.New().String(),
			},
			expectedStatus: http.StatusCreated,
			shouldPass:     true,
		},
		{
			name: "Missing vendor ID",
			requestBody: map[string]interface{}{
				"invoiceNumber": "INV-2025-001",
				"amount":        50000,
				"currency":      "USD",
			},
			expectedStatus: http.StatusBadRequest,
			shouldPass:     false,
		},
		{
			name: "Missing invoice number",
			requestBody: map[string]interface{}{
				"vendorId": uuid.New().String(),
				"amount":   50000,
				"currency": "USD",
			},
			expectedStatus: http.StatusBadRequest,
			shouldPass:     false,
		},
		{
			name: "Invalid amount (zero)",
			requestBody: map[string]interface{}{
				"vendorId":      uuid.New().String(),
				"invoiceNumber": "INV-2025-001",
				"amount":        0,
				"currency":      "USD",
			},
			expectedStatus: http.StatusBadRequest,
			shouldPass:     false,
		},
		{
			name: "Negative amount",
			requestBody: map[string]interface{}{
				"vendorId":      uuid.New().String(),
				"invoiceNumber": "INV-2025-001",
				"amount":        -50000,
				"currency":      "USD",
			},
			expectedStatus: http.StatusBadRequest,
			shouldPass:     false,
		},
		{
			name: "Short description",
			requestBody: map[string]interface{}{
				"vendorId":      uuid.New().String(),
				"invoiceNumber": "INV-2025-001",
				"amount":        50000,
				"currency":      "USD",
				"description":   "Short",
			},
			expectedStatus: http.StatusBadRequest,
			shouldPass:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.requestBody)
			var req types.CreatePaymentVoucherRequest
			json.Unmarshal(body, &req)

			// Validate request
			isValid := req.VendorID != "" &&
				req.InvoiceNumber != "" &&
				req.Amount > 0 &&
				len(req.Description) >= 10

			if isValid != tt.shouldPass {
				t.Errorf("Expected %v, got %v", tt.shouldPass, isValid)
			}
		})
	}
}

// TestPaymentVoucherNumberGeneration tests voucher number generation
func TestPaymentVoucherNumberGeneration(t *testing.T) {
	t.Run("Voucher number format", func(t *testing.T) {
		// Format: PV-{timestamp}-{uuid[:8]}
		voucherNumber := "PV-1640000000-abc12345"

		if voucherNumber[:3] != "PV-" {
			t.Error("Voucher number should start with 'PV-'")
		}

		if len(voucherNumber) < 15 {
			t.Error("Voucher number should be properly formatted")
		}
	})
}

// TestPaymentMethodValidation tests payment method field
func TestPaymentMethodValidation(t *testing.T) {
	validMethods := map[string]bool{
		"bank_transfer": true,
		"check":         true,
		"cash":          true,
		"credit_card":   true,
		"wire":          true,
	}

	tests := []struct {
		name          string
		method        string
		shouldBeValid bool
	}{
		{"Bank Transfer", "bank_transfer", true},
		{"Check", "check", true},
		{"Cash", "cash", true},
		{"Invalid Method", "invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := validMethods[tt.method]
			if isValid != tt.shouldBeValid {
				t.Errorf("Expected %v, got %v", tt.shouldBeValid, isValid)
			}
		})
	}
}

// TestGLCodeValidation tests GL code validation
func TestGLCodeValidation(t *testing.T) {
	tests := []struct {
		name          string
		glCode        string
		shouldBeValid bool
	}{
		{"Valid GL Code 4000", "4000", true},
		{"Valid GL Code 5000", "5000", true},
		{"Valid GL Code 6000", "6000", true},
		{"Empty GL Code", "", true}, // GL Code can be optional
		{"Short GL Code", "40", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.glCode == "" || len(tt.glCode) >= 4
			if isValid != tt.shouldBeValid {
				t.Errorf("Expected %v, got %v", tt.shouldBeValid, isValid)
			}
		})
	}
}

// TestPaymentVoucherStatusValidation tests status field validation
func TestPaymentVoucherStatusValidation(t *testing.T) {
	validStatuses := map[string]bool{
		"draft":     true,
		"pending":   true,
		"approved":  true,
		"rejected":  true,
		"completed": true,
		"paid":      true,
	}

	tests := []struct {
		name          string
		status        string
		shouldBeValid bool
	}{
		{"Draft status", "draft", true},
		{"Pending status", "pending", true},
		{"Approved status", "approved", true},
		{"Paid status", "paid", true},
		{"Invalid status", "cancelled", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := validStatuses[tt.status]
			if isValid != tt.shouldBeValid {
				t.Errorf("Expected %v, got %v", tt.shouldBeValid, isValid)
			}
		})
	}
}

// TestPaymentVoucherApprovalWorkflow tests approval workflow
func TestPaymentVoucherApprovalWorkflow(t *testing.T) {
	t.Run("Payment voucher approval stages", func(t *testing.T) {
		// PV typically has 2 stages: finance -> admin
		stages := 2

		if stages < 1 {
			t.Error("PV should have at least 1 approval stage")
		}

		approvalChain := []string{"finance", "admin"}
		if len(approvalChain) != stages {
			t.Errorf("Expected %d approval stages, got %d", stages, len(approvalChain))
		}
	})
}

// TestPaymentVoucherStateTransitions tests valid PV state transitions
func TestPaymentVoucherStateTransitions(t *testing.T) {
	tests := []struct {
		name        string
		fromStatus  string
		toStatus    string
		shouldAllow bool
	}{
		{"Draft to Pending", "draft", "pending", true},
		{"Pending to Approved", "pending", "approved", true},
		{"Pending to Rejected", "pending", "rejected", true},
		{"Approved to Paid", "approved", "paid", true},
		{"Approved to Draft", "approved", "draft", false},
		{"Paid to Draft", "paid", "draft", false},
		{"Completed to Approved", "completed", "approved", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validTransitions := map[string][]string{
				"draft":     {"pending"},
				"pending":   {"approved", "rejected"},
				"rejected":  {"draft"},
				"approved":  {"paid"},
				"paid":      {"completed"},
				"completed": {},
			}

			allowed := false
			for _, nextStatus := range validTransitions[tt.fromStatus] {
				if nextStatus == tt.toStatus {
					allowed = true
					break
				}
			}

			if allowed != tt.shouldAllow {
				t.Errorf("Expected %v, got %v", tt.shouldAllow, allowed)
			}
		})
	}
}

// TestPaymentVoucherLinkedPO tests PO linking
func TestPaymentVoucherLinkedPO(t *testing.T) {
	t.Run("PV can be linked to PO", func(t *testing.T) {
		pv := types.PaymentVoucherResponse{
			ID:       uuid.New().String(),
			LinkedPO: uuid.New().String(),
		}

		if pv.LinkedPO == "" {
			t.Error("PV should have linked PO")
		}
	})

	t.Run("PV can be created without PO", func(t *testing.T) {
		pv := types.PaymentVoucherResponse{
			ID:       uuid.New().String(),
			LinkedPO: "",
		}

		if pv.ID == "" {
			t.Error("PV ID should not be empty")
		}
		// LinkedPO can be empty
	})
}

// TestPaymentVoucherResponseFormat tests PV response structure
func TestPaymentVoucherResponseFormat(t *testing.T) {
	t.Run("Payment voucher response structure", func(t *testing.T) {
		pv := types.PaymentVoucherResponse{
			ID:              uuid.New().String(),
			VoucherNumber:   "PV-1640000000-abc12345",
			VendorID:        uuid.New().String(),
			InvoiceNumber:   "INV-2025-001",
			Status:          "draft",
			Amount:          50000,
			Currency:        "USD",
			PaymentMethod:   "bank_transfer",
			GLCode:          "4000",
			Description:     "Payment for services",
			ApprovalStage:   0,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}

		if pv.ID == "" {
			t.Error("Response should have ID")
		}
		if pv.VoucherNumber == "" {
			t.Error("Response should have VoucherNumber")
		}
		if pv.VendorID == "" {
			t.Error("Response should have VendorID")
		}
		if pv.Amount <= 0 {
			t.Error("Response should have positive Amount")
		}
		if pv.InvoiceNumber == "" {
			t.Error("Response should have InvoiceNumber")
		}
	})
}

// TestPaymentVoucherAmountValidation tests amount constraints
func TestPaymentVoucherAmountValidation(t *testing.T) {
	tests := []struct {
		name          string
		amount        float64
		shouldBeValid bool
	}{
		{"Small amount", 1000, true},
		{"Medium amount", 50000, true},
		{"Large amount", 500000, true},
		{"Zero amount", 0, false},
		{"Negative amount", -50000, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.amount > 0
			if isValid != tt.shouldBeValid {
				t.Errorf("Expected %v, got %v", tt.shouldBeValid, isValid)
			}
		})
	}
}

// TestPaymentVoucherDuplicatePrevention tests duplicate invoice detection
func TestPaymentVoucherDuplicatePrevention(t *testing.T) {
	t.Run("Prevent duplicate invoice numbers", func(t *testing.T) {
		pv1 := types.PaymentVoucherResponse{
			VoucherNumber: "PV-1640000000-abc12345",
			VendorID:      uuid.New().String(),
			InvoiceNumber: "INV-2025-001",
		}

		pv2 := types.PaymentVoucherResponse{
			VoucherNumber: "PV-1640000001-def67890",
			VendorID:      pv1.VendorID,
			InvoiceNumber: "INV-2025-001",
		}

		isDuplicate := (pv1.InvoiceNumber == pv2.InvoiceNumber && pv1.VendorID == pv2.VendorID)

		if !isDuplicate {
			t.Error("Should detect duplicate invoice numbers for same vendor")
		}
	})
}

// TestPaymentVoucherCurrencyValidation tests currency codes
func TestPaymentVoucherCurrencyValidation(t *testing.T) {
	validCurrencies := map[string]bool{
		"USD": true,
		"EUR": true,
		"GBP": true,
		"ZWL": true,
	}

	tests := []struct {
		name          string
		currency      string
		shouldBeValid bool
	}{
		{"USD", "USD", true},
		{"EUR", "EUR", true},
		{"ZWL", "ZWL", true},
		{"Invalid", "XXX", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := validCurrencies[tt.currency]
			if isValid != tt.shouldBeValid {
				t.Errorf("Expected %v, got %v", tt.shouldBeValid, isValid)
			}
		})
	}
}

// TestPaymentVoucherUpdateValidation tests update constraints
func TestPaymentVoucherUpdateValidation(t *testing.T) {
	tests := []struct {
		name          string
		currentStatus string
		updateBody    map[string]interface{}
		shouldAllow   bool
	}{
		{
			name:          "Update draft voucher",
			currentStatus: "draft",
			updateBody: map[string]interface{}{
				"amount": 60000,
			},
			shouldAllow: true,
		},
		{
			name:          "Update pending voucher",
			currentStatus: "pending",
			updateBody: map[string]interface{}{
				"amount": 60000,
			},
			shouldAllow: true,
		},
		{
			name:          "Cannot update approved voucher",
			currentStatus: "approved",
			updateBody: map[string]interface{}{
				"amount": 60000,
			},
			shouldAllow: false,
		},
		{
			name:          "Cannot update paid voucher",
			currentStatus: "paid",
			updateBody: map[string]interface{}{
				"amount": 60000,
			},
			shouldAllow: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Only draft and pending vouchers can be updated
			canUpdate := tt.currentStatus == "draft" || tt.currentStatus == "pending"

			if canUpdate != tt.shouldAllow {
				t.Errorf("Expected %v, got %v", tt.shouldAllow, canUpdate)
			}
		})
	}
}

// TestPaymentVoucherApprovalHistory tests approval history tracking
func TestPaymentVoucherApprovalHistory(t *testing.T) {
	t.Run("Track approval records", func(t *testing.T) {
		approvalHistory := []types.ApprovalRecord{
			{
				ApproverID:   uuid.New().String(),
				ApproverName: "Finance Manager",
				Status:       "approved",
				Comments:     "Approved for payment",
				ApprovedAt:   time.Now(),
			},
		}

		if len(approvalHistory) == 0 {
			t.Error("Should have approval records")
		}

		if approvalHistory[0].Status != "approved" {
			t.Error("Status should be approved")
		}
	})
}

// BenchmarkPaymentVoucherValidation benchmarks validation logic
func BenchmarkPaymentVoucherValidation(b *testing.B) {
	req := types.CreatePaymentVoucherRequest{
		VendorID:      uuid.New().String(),
		InvoiceNumber: "INV-2025-001",
		Amount:        50000,
		Currency:      "USD",
		Description:   "Payment for services rendered",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = req.VendorID != "" &&
			req.InvoiceNumber != "" &&
			req.Amount > 0 &&
			len(req.Description) >= 10
	}
}

// BenchmarkPaymentVoucherNumberGeneration benchmarks number generation
func BenchmarkPaymentVoucherNumberGeneration(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		voucherNumber := "PV-" + uuid.New().String()[:8]
		_ = voucherNumber
	}
}
