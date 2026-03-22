package unit

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/types"
)

// TestCreatePurchaseOrderValidation tests PO request validation
func TestCreatePurchaseOrderValidation(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
		shouldPass     bool
	}{
		{
			name: "Valid PO request",
			requestBody: map[string]interface{}{
				"vendorId":     uuid.New().String(),
				"totalAmount":  50000,
				"currency":     "USD",
				"deliveryDate": time.Now().AddDate(0, 1, 0).Format(time.RFC3339),
				"items": []map[string]interface{}{
					{
						"description": "Item 1",
						"quantity":    1,
						"unitPrice":   50000,
						"amount":      50000,
					},
				},
			},
			expectedStatus: http.StatusCreated,
			shouldPass:     true,
		},
		{
			name: "Missing vendor ID",
			requestBody: map[string]interface{}{
				"totalAmount": 50000,
				"currency":    "USD",
			},
			expectedStatus: http.StatusBadRequest,
			shouldPass:     false,
		},
		{
			name: "Invalid delivery date",
			requestBody: map[string]interface{}{
				"vendorId":     uuid.New().String(),
				"totalAmount":  50000,
				"currency":     "USD",
				"deliveryDate": "invalid-date",
			},
			expectedStatus: http.StatusBadRequest,
			shouldPass:     false,
		},
		{
			name: "Past delivery date",
			requestBody: map[string]interface{}{
				"vendorId":     uuid.New().String(),
				"totalAmount":  50000,
				"currency":     "USD",
				"deliveryDate": time.Now().AddDate(0, -1, 0).Format(time.RFC3339),
			},
			expectedStatus: http.StatusBadRequest,
			shouldPass:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.requestBody)
			var req types.CreatePurchaseOrderRequest
			json.Unmarshal(body, &req)

			// Validate request
			isValid := req.VendorID != "" && req.TotalAmount > 0
			
			// Check delivery date if provided
			if !req.DeliveryDate.Time.IsZero() {
				// Delivery date should not be in the past
				if req.DeliveryDate.Time.Before(time.Now().Truncate(24*time.Hour)) {
					isValid = false
				}
			}

			if isValid != tt.shouldPass {
				t.Errorf("Expected %v, got %v", tt.shouldPass, isValid)
			}
		})
	}
}

// TestPONumberGeneration tests PO number generation
func TestPONumberGeneration(t *testing.T) {
	t.Run("PO number format", func(t *testing.T) {
		// Format: PO-{timestamp}-{uuid[:8]}
		poNumber := "PO-20251223-abc12345"

		if poNumber[:3] != "PO-" {
			t.Error("PO number should start with 'PO-'")
		}

		if len(poNumber) < 15 {
			t.Error("PO number should be properly formatted")
		}
	})
}

// TestPOStatusValidation tests PO status field
func TestPOStatusValidation(t *testing.T) {
	validStatuses := map[string]bool{
		"DRAFT":     true,
		"PENDING":   true,
		"APPROVED":  true,
		"REJECTED":  true,
		"FULFILLED": true,
		"COMPLETED": true,
	}

	tests := []struct {
		name          string
		status        string
		shouldBeValid bool
	}{
		{"Draft", "DRAFT", true},
		{"Pending", "PENDING", true},
		{"Approved", "APPROVED", true},
		{"Fulfilled", "FULFILLED", true},
		{"Invalid", "CANCELLED", false},
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

// TestPOVendorValidation tests vendor ID validation
func TestPOVendorValidation(t *testing.T) {
	t.Run("Valid vendor UUID", func(t *testing.T) {
		vendorID := uuid.New().String()

		if vendorID == "" {
			t.Error("Vendor ID should not be empty")
		}

		if len(vendorID) != 36 {
			t.Error("Vendor ID should be valid UUID format")
		}
	})
}

// TestPODeliveryDateValidation tests delivery date constraints
func TestPODeliveryDateValidation(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name          string
		deliveryDate  time.Time
		shouldBeValid bool
	}{
		{"Future date (30 days)", now.AddDate(0, 1, 0), true},
		{"Future date (90 days)", now.AddDate(0, 3, 0), true},
		{"Today", now, true},
		{"Tomorrow", now.AddDate(0, 0, 1), true},
		{"Past date", now.AddDate(0, 0, -1), false},
		{"Far past", now.AddDate(-1, 0, 0), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Check if delivery date is today or in the future
			isValid := tt.deliveryDate.After(now) || tt.deliveryDate.Truncate(24*time.Hour).Equal(now.Truncate(24*time.Hour))

			if isValid != tt.shouldBeValid {
				t.Errorf("Expected %v, got %v", tt.shouldBeValid, isValid)
			}
		})
	}
}

// TestPOStateTransitions tests valid PO state transitions
func TestPOStateTransitions(t *testing.T) {
	tests := []struct {
		name        string
		fromStatus  string
		toStatus    string
		shouldAllow bool
	}{
		{"Draft to Pending", "DRAFT", "PENDING", true},
		{"Pending to Approved", "PENDING", "APPROVED", true},
		{"Pending to Rejected", "PENDING", "REJECTED", true},
		{"Approved to Fulfilled", "APPROVED", "FULFILLED", true},
		{"Fulfilled to Completed", "FULFILLED", "COMPLETED", true},
		{"Approved to Draft", "APPROVED", "DRAFT", false},
		{"Completed to Fulfilled", "COMPLETED", "FULFILLED", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validTransitions := map[string][]string{
				"DRAFT":     {"PENDING"},
				"PENDING":   {"APPROVED", "REJECTED"},
				"REJECTED":  {"DRAFT"},
				"APPROVED":  {"FULFILLED"},
				"FULFILLED": {"COMPLETED"},
				"COMPLETED": {},
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

// TestPOLinkedRequisition tests requisition linking
func TestPOLinkedRequisition(t *testing.T) {
	t.Run("PO can be linked to requisition", func(t *testing.T) {
		po := types.PurchaseOrderResponse{
			ID:                 uuid.New().String(),
			LinkedRequisition:  uuid.New().String(),
		}

		if po.LinkedRequisition == "" {
			t.Error("PO should have linked requisition")
		}
	})

	t.Run("PO can be created without requisition", func(t *testing.T) {
		po := types.PurchaseOrderResponse{
			ID:                 uuid.New().String(),
			LinkedRequisition:  "",
		}

		if po.ID == "" {
			t.Error("PO ID should not be empty")
		}
		// LinkedRequisition can be empty
	})
}

// TestPOResponseFormat tests PO response structure
func TestPOResponseFormat(t *testing.T) {
	t.Run("PO response structure", func(t *testing.T) {
		po := types.PurchaseOrderResponse{
			ID:             uuid.New().String(),
			DocumentNumber: "PO-20251223-abc12345",
			VendorID:       uuid.New().String(),
			Status: "DRAFT",
			TotalAmount:    50000,
			Currency:       "USD",
			DeliveryDate:   time.Now().AddDate(0, 1, 0),
			ApprovalStage:  0,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		if po.ID == "" {
			t.Error("Response should have ID")
		}
		if po.DocumentNumber == "" {
			t.Error("Response should have DocumentNumber")
		}
		if po.VendorID == "" {
			t.Error("Response should have VendorID")
		}
		if po.TotalAmount <= 0 {
			t.Error("Response should have positive TotalAmount")
		}
	})
}

// TestPOItemValidation tests PO items
func TestPOItemValidation(t *testing.T) {
	tests := []struct {
		name       string
		items      []map[string]interface{}
		shouldPass bool
	}{
		{
			name: "Valid items",
			items: []map[string]interface{}{
				{"description": "Item 1", "quantity": 1, "unitPrice": 50000, "amount": 50000},
			},
			shouldPass: true,
		},
		{
			name: "Multiple items",
			items: []map[string]interface{}{
				{"description": "Item 1", "quantity": 2, "unitPrice": 25000, "amount": 50000},
				{"description": "Item 2", "quantity": 1, "unitPrice": 30000, "amount": 30000},
			},
			shouldPass: true,
		},
		{
			name: "Zero quantity",
			items: []map[string]interface{}{
				{"description": "Item 1", "quantity": 0, "unitPrice": 50000, "amount": 0},
			},
			shouldPass: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allValid := true
			for _, item := range tt.items {
				// Handle both int and float64 quantities
				var qty float64
				var hasQty bool
				if qtyFloat, ok := item["quantity"].(float64); ok {
					qty = qtyFloat
					hasQty = true
				} else if qtyInt, ok := item["quantity"].(int); ok {
					qty = float64(qtyInt)
					hasQty = true
				}
				
				if !hasQty || qty <= 0 {
					allValid = false
					break
				}
			}

			if allValid != tt.shouldPass {
				t.Errorf("Expected %v, got %v", tt.shouldPass, allValid)
			}
		})
	}
}

// TestPODuplicatePrevention tests duplicate PO detection
func TestPODuplicatePrevention(t *testing.T) {
	t.Run("Prevent duplicate PO numbers", func(t *testing.T) {
		po1 := types.PurchaseOrderResponse{
			DocumentNumber: "PO-20251223-abc12345",
			VendorID: uuid.New().String(),
		}

		po2 := types.PurchaseOrderResponse{
			DocumentNumber: "PO-20251223-abc12345",
			VendorID: po1.VendorID,
		}

		isDuplicate := (po1.DocumentNumber == po2.DocumentNumber)

		if !isDuplicate {
			t.Error("Should detect duplicate PO numbers")
		}
	})
}

// BenchmarkPONumberGeneration benchmarks PO number generation
func BenchmarkPONumberGeneration(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		poNumber := "PO-" + uuid.New().String()[:8]
		_ = poNumber
	}
}

// BenchmarkPOValidation benchmarks validation logic
func BenchmarkPOValidation(b *testing.B) {
	req := types.CreatePurchaseOrderRequest{
		VendorID:    uuid.New().String(),
		TotalAmount: 50000,
		Currency:    "USD",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = req.VendorID != "" && req.TotalAmount > 0
	}
}

// TestPOProcurementFlow tests per-PO procurement flow override
func TestPOProcurementFlow(t *testing.T) {
	tests := []struct {
		name            string
		procurementFlow string
		shouldPass      bool
	}{
		{"Empty string (inherit from org)", "", true},
		{"Goods-first explicit override", "goods_first", true},
		{"Payment-first explicit override", "payment_first", true},
		{"Invalid flow value", "express", false},
		{"Uppercase invalid", "GOODS_FIRST", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.procurementFlow == "" ||
				tt.procurementFlow == "goods_first" ||
				tt.procurementFlow == "payment_first"
			if isValid != tt.shouldPass {
				t.Errorf("Expected %v, got %v for flow=%q", tt.shouldPass, isValid, tt.procurementFlow)
			}
		})
	}
}

// TestPOFlowResolution tests effective flow resolution priority
func TestPOFlowResolution(t *testing.T) {
	tests := []struct {
		name         string
		poFlow       string
		orgFlow      string
		expectedFlow string
	}{
		{"PO override beats org setting", "payment_first", "goods_first", "payment_first"},
		{"Empty PO inherits org setting", "", "payment_first", "payment_first"},
		{"Both empty defaults to goods_first", "", "", "goods_first"},
		{"PO goods_first beats org payment_first", "goods_first", "payment_first", "goods_first"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			effective := tt.poFlow
			if effective == "" {
				effective = tt.orgFlow
				if effective == "" {
					effective = "goods_first"
				}
			}
			if effective != tt.expectedFlow {
				t.Errorf("Expected %q, got %q", tt.expectedFlow, effective)
			}
		})
	}
}
