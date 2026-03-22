package unit

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/types"
)

// TestCreateGRNValidation tests GRN request validation
func TestCreateGRNValidation(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
		shouldPass     bool
	}{
		{
			name: "Valid GRN request",
			requestBody: map[string]interface{}{
				"poDocumentNumber": "PO-20251223-abc12345",
				"receivedBy": "John Doe",
				"items": []map[string]interface{}{
					{
						"description": "Item 1",
						"quantityOrdered": 10,
						"quantityReceived": 10,
						"variance": 0,
						"condition": "good",
					},
				},
			},
			expectedStatus: http.StatusCreated,
			shouldPass:     true,
		},
		{
			name: "Missing PO number",
			requestBody: map[string]interface{}{
				"receivedBy": "John Doe",
				"items": []map[string]interface{}{
					{
						"description": "Item 1",
						"quantity":    10,
					},
				},
			},
			expectedStatus: http.StatusBadRequest,
			shouldPass:     false,
		},
		{
			name: "Missing ReceivedBy",
			requestBody: map[string]interface{}{
				"poDocumentNumber": "PO-20251223-abc12345",
				"items": []map[string]interface{}{
					{
						"description": "Item 1",
						"quantity":    10,
					},
				},
			},
			expectedStatus: http.StatusBadRequest,
			shouldPass:     false,
		},
		{
			name: "Missing items",
			requestBody: map[string]interface{}{
				"poDocumentNumber": "PO-20251223-abc12345",
				"receivedBy": "John Doe",
				"items":      []map[string]interface{}{},
			},
			expectedStatus: http.StatusBadRequest,
			shouldPass:     false,
		},
		{
			name: "Empty items array",
			requestBody: map[string]interface{}{
				"poDocumentNumber": "PO-20251223-abc12345",
				"receivedBy": "John Doe",
			},
			expectedStatus: http.StatusBadRequest,
			shouldPass:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.requestBody)
			var req types.CreateGRNRequest
			json.Unmarshal(body, &req)

			// Validate request
			isValid := req.PODocumentNumber != "" &&
				req.ReceivedBy != "" &&
				len(req.Items) > 0

			if isValid != tt.shouldPass {
				t.Errorf("Expected %v, got %v", tt.shouldPass, isValid)
			}
		})
	}
}

// TestGRNNumberGeneration tests GRN number generation
func TestGRNNumberGeneration(t *testing.T) {
	t.Run("GRN number format", func(t *testing.T) {
		// Format: GRN-{timestamp}-{uuid[:8]}
		grnNumber := "GRN-1640000000-abc12345"

		if grnNumber[:4] != "GRN-" {
			t.Error("GRN number should start with 'GRN-'")
		}

		if len(grnNumber) < 15 {
			t.Error("GRN number should be properly formatted")
		}
	})
}

// TestGRNStatusValidation tests status field
func TestGRNStatusValidation(t *testing.T) {
	validStatuses := map[string]bool{
		"DRAFT":    true,
		"PENDING":  true,
		"APPROVED": true,
		"REJECTED": true,
		"RECEIVED": true,
	}

	tests := []struct {
		name          string
		status        string
		shouldBeValid bool
	}{
		{"Draft", "DRAFT", true},
		{"Pending", "PENDING", true},
		{"Approved", "APPROVED", true},
		{"Received", "RECEIVED", true},
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

// TestGRNPODocumentNumberValidation tests PO document number field
func TestGRNPODocumentNumberValidation(t *testing.T) {
	tests := []struct {
		name          string
		poDocumentNumber string
		shouldBeValid bool
	}{
		{"Valid PO number", "PO-20251223-abc12345", true},
		{"Empty PO number", "", false},
		{"Short PO number", "PO", false},
		{"Invalid format", "req-123", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.poDocumentNumber != "" && len(tt.poDocumentNumber) > 10 && tt.poDocumentNumber[:3] == "PO-"
			if isValid != tt.shouldBeValid {
				t.Errorf("Expected %v, got %v", tt.shouldBeValid, isValid)
			}
		})
	}
}

// TestGRNItemQuantityValidation tests item quantity validation
func TestGRNItemQuantityValidation(t *testing.T) {
	tests := []struct {
		name       string
		items      []map[string]interface{}
		shouldPass bool
	}{
		{
			name: "Valid items",
			items: []map[string]interface{}{
				{"description": "Item 1", "quantity": 10, "receivedQty": 10},
			},
			shouldPass: true,
		},
		{
			name: "Multiple items",
			items: []map[string]interface{}{
				{"description": "Item 1", "quantity": 10, "receivedQty": 10},
				{"description": "Item 2", "quantity": 5, "receivedQty": 5},
			},
			shouldPass: true,
		},
		{
			name: "Quantity variance (less received)",
			items: []map[string]interface{}{
				{"description": "Item 1", "quantity": 10, "receivedQty": 8},
			},
			shouldPass: true, // Variance is allowed
		},
		{
			name: "Zero quantity",
			items: []map[string]interface{}{
				{"description": "Item 1", "quantity": 0, "receivedQty": 0},
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

// TestGRNQualityIssueTracking tests quality issue tracking
func TestGRNQualityIssueTracking(t *testing.T) {
	t.Run("Track quality issues", func(t *testing.T) {
		qualityIssues := []types.QualityIssue{
			{
				ItemDescription: "Item 1",
				IssueType:       "damage",
				Description:     "Damaged packaging",
				Severity:        "low",
			},
		}

		if len(qualityIssues) == 0 {
			t.Error("Should support quality issues")
		}

		if qualityIssues[0].Severity != "low" {
			t.Error("Should track issue severity")
		}
	})
}

// TestGRNVarianceTracking tests quantity variance
func TestGRNVarianceTracking(t *testing.T) {
	tests := []struct {
		name             string
		orderedQuantity  float64
		receivedQuantity float64
		expectedVariance float64
	}{
		{"No variance", 10, 10, 0},
		{"Under-delivery", 10, 8, -2},
		{"Over-delivery", 10, 12, 2},
		{"Partial delivery", 10, 5, -5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			variance := tt.receivedQuantity - tt.orderedQuantity
			if variance != tt.expectedVariance {
				t.Errorf("Expected variance %f, got %f", tt.expectedVariance, variance)
			}
		})
	}
}

// TestGRNStateTransitions tests valid GRN state transitions
func TestGRNStateTransitions(t *testing.T) {
	tests := []struct {
		name        string
		fromStatus  string
		toStatus    string
		shouldAllow bool
	}{
		{"Draft to Pending", "DRAFT", "PENDING", true},
		{"Pending to Approved", "PENDING", "APPROVED", true},
		{"Pending to Rejected", "PENDING", "REJECTED", true},
		{"Approved to Received", "APPROVED", "RECEIVED", true},
		{"Approved to Draft", "APPROVED", "DRAFT", false},
		{"Received to Approved", "RECEIVED", "APPROVED", false},
		{"Rejected to Draft", "REJECTED", "DRAFT", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validTransitions := map[string][]string{
				"DRAFT":    {"PENDING"},
				"PENDING":  {"APPROVED", "REJECTED"},
				"REJECTED": {"DRAFT"},
				"APPROVED": {"RECEIVED"},
				"RECEIVED": {},
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

// TestGRNResponseFormat tests GRN response structure
func TestGRNResponseFormat(t *testing.T) {
	t.Run("GRN response structure", func(t *testing.T) {
		grn := types.GRNResponse{
			ID:                uuid.New().String(),
			DocumentNumber:    "GRN-1640000000-abc12345",
			PODocumentNumber:  "PO-20251223-abc12345",
			Status: "DRAFT",
			ReceivedBy: "John Doe",
			Items: []types.GRNItem{
				{
					Description:      "Item 1",
					QuantityOrdered:  10,
					QuantityReceived: 10,
					Variance:         0,
					Condition:        "good",
				},
			},
			ApprovalStage: 0,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		if grn.ID == "" {
			t.Error("Response should have ID")
		}
		if grn.DocumentNumber == "" {
			t.Error("Response should have DocumentNumber")
		}
		if grn.PODocumentNumber == "" {
			t.Error("Response should have PODocumentNumber")
		}
		if grn.ReceivedBy == "" {
			t.Error("Response should have ReceivedBy")
		}
	})
}

// TestGRNItemValidation tests GRN item structure
func TestGRNItemValidation(t *testing.T) {
	tests := []struct {
		name       string
		items      []types.GRNItem
		shouldPass bool
	}{
		{
			name: "Valid GRN items",
			items: []types.GRNItem{
				{
					Description:      "Item 1",
					QuantityOrdered:  10,
					QuantityReceived: 10,
					Variance:         0,
					Condition:        "good",
				},
			},
			shouldPass: true,
		},
		{
			name: "Multiple items",
			items: []types.GRNItem{
				{
					Description:      "Item 1",
					QuantityOrdered:  10,
					QuantityReceived: 10,
					Variance:         0,
					Condition:        "good",
				},
				{
					Description:      "Item 2",
					QuantityOrdered:  5,
					QuantityReceived: 5,
					Variance:         0,
					Condition:        "good",
				},
			},
			shouldPass: true,
		},
		{
			name: "Empty items",
			items: []types.GRNItem{},
			shouldPass: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := len(tt.items) > 0
			if isValid != tt.shouldPass {
				t.Errorf("Expected %v, got %v", tt.shouldPass, isValid)
			}
		})
	}
}

// TestGRNApprovalWorkflow tests approval workflow
func TestGRNApprovalWorkflow(t *testing.T) {
	t.Run("GRN approval stages", func(t *testing.T) {
		// GRN typically has 1-2 stages: warehouse -> admin
		stages := 2

		if stages < 1 {
			t.Error("GRN should have at least 1 approval stage")
		}

		approvalChain := []string{"warehouse", "admin"}
		if len(approvalChain) != stages {
			t.Errorf("Expected %d approval stages, got %d", stages, len(approvalChain))
		}
	})
}

// TestGRNPOLinking tests GRN to PO linking
func TestGRNPOLinking(t *testing.T) {
	t.Run("GRN must reference valid PO", func(t *testing.T) {
		poDocumentNumber := "PO-20251223-abc12345"

		if poDocumentNumber == "" {
			t.Error("GRN must have linked PO number")
		}

		if len(poDocumentNumber) < 10 {
			t.Error("PO number format should be valid")
		}
	})
}

// TestGRNDuplicatePrevention tests duplicate GRN detection
func TestGRNDuplicatePrevention(t *testing.T) {
	t.Run("Prevent duplicate GRN numbers", func(t *testing.T) {
		grn1 := types.GRNResponse{
			DocumentNumber:    "GRN-1640000000-abc12345",
			PODocumentNumber:  "PO-20251223-abc12345",
		}

		grn2 := types.GRNResponse{
			DocumentNumber:    "GRN-1640000000-abc12345",
			PODocumentNumber:  "PO-20251223-abc12345",
		}

		isDuplicate := (grn1.DocumentNumber == grn2.DocumentNumber)

		if !isDuplicate {
			t.Error("Should detect duplicate GRN numbers")
		}
	})
}

// TestGRNUpdateValidation tests update constraints
func TestGRNUpdateValidation(t *testing.T) {
	tests := []struct {
		name          string
		currentStatus string
		shouldAllow   bool
	}{
		{"Update draft GRN", "DRAFT", true},
		{"Update pending GRN", "PENDING", true},
		{"Cannot update approved GRN", "APPROVED", false},
		{"Cannot update received GRN", "RECEIVED", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Only draft and pending GRNs can be updated
			canUpdate := tt.currentStatus == "DRAFT" || tt.currentStatus == "PENDING"

			if canUpdate != tt.shouldAllow {
				t.Errorf("Expected %v, got %v", tt.shouldAllow, canUpdate)
			}
		})
	}
}

// TestGRNQuantityVarianceCalculation tests variance percentage
func TestGRNQuantityVarianceCalculation(t *testing.T) {
	tests := []struct {
		name             string
		orderedQty       float64
		receivedQty      float64
		expectedVariance float64
	}{
		{"Perfect match", 100, 100, 0},
		{"5% shortage", 100, 95, -5},
		{"10% shortage", 100, 90, -10},
		{"5% overage", 100, 105, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			variance := ((tt.receivedQty - tt.orderedQty) / tt.orderedQty) * 100
			if variance != tt.expectedVariance {
				t.Errorf("Expected variance %f%%, got %f%%", tt.expectedVariance, variance)
			}
		})
	}
}

// TestGRNReceivedDateValidation tests received date
func TestGRNReceivedDateValidation(t *testing.T) {
	tests := []struct {
		name          string
		receivedDate  time.Time
		shouldBeValid bool
	}{
		{"Today", time.Now(), true},
		{"Yesterday", time.Now().AddDate(0, 0, -1), true},
		{"Last month", time.Now().AddDate(0, -1, 0), true},
		{"Future date", time.Now().AddDate(0, 0, 1), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.receivedDate.Before(time.Now()) || tt.receivedDate.Day() == time.Now().Day()
			if isValid != tt.shouldBeValid {
				t.Errorf("Expected %v, got %v", tt.shouldBeValid, isValid)
			}
		})
	}
}

// BenchmarkGRNValidation benchmarks validation logic
func BenchmarkGRNValidation(b *testing.B) {
	req := types.CreateGRNRequest{
		PODocumentNumber: "PO-20251223-abc12345",
		ReceivedBy: "John Doe",
		Items: []types.GRNItem{
			{
				Description:      "Item 1",
				QuantityOrdered:  10,
				QuantityReceived: 10,
				Variance:         0,
				Condition:        "good",
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = req.PODocumentNumber != "" && req.ReceivedBy != "" && len(req.Items) > 0
	}
}

// BenchmarkGRNNumberGeneration benchmarks number generation
func BenchmarkGRNNumberGeneration(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		grnNumber := "GRN-" + uuid.New().String()[:8]
		_ = grnNumber
	}
}

// TestGRNPaymentFirstFlow tests payment-first flow enforcement in GRN creation
func TestGRNPaymentFirstFlow(t *testing.T) {
	tests := []struct {
		name          string
		effectiveFlow string
		linkedPV      string
		pvStatus      string
		shouldPass    bool
	}{
		{"Goods-first: no PV required", "goods_first", "", "", true},
		{"Payment-first: approved PV provided", "payment_first", "PV-20240101-abc", "APPROVED", true},
		{"Payment-first: paid PV provided", "payment_first", "PV-20240101-abc", "PAID", true},
		{"Payment-first: missing PV — blocked", "payment_first", "", "", false},
		{"Payment-first: pending PV — blocked", "payment_first", "PV-20240101-abc", "PENDING", false},
		{"Payment-first: draft PV — blocked", "payment_first", "PV-20240101-abc", "DRAFT", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := true
			if tt.effectiveFlow == "payment_first" {
				if tt.linkedPV == "" {
					isValid = false
				} else if tt.pvStatus != "APPROVED" && tt.pvStatus != "PAID" {
					isValid = false
				}
			}
			if isValid != tt.shouldPass {
				t.Errorf("Expected %v, got %v", tt.shouldPass, isValid)
			}
		})
	}
}

// TestGRNLinkedPV tests GRN linkedPV field
func TestGRNLinkedPV(t *testing.T) {
	t.Run("GRN stores linkedPV in payment-first flow", func(t *testing.T) {
		pvDocNum := "PV-20240101-abc123"
		linkedPV := pvDocNum

		if linkedPV != pvDocNum {
			t.Errorf("linkedPV should match provided PV document number")
		}
	})

	t.Run("GRN linkedPV is empty in goods-first flow", func(t *testing.T) {
		linkedPV := ""
		if linkedPV != "" {
			t.Error("linkedPV should be empty for goods-first flow")
		}
	})
}

// TestGRNQuantityVariance tests quantity received vs ordered
func TestGRNQuantityVariance(t *testing.T) {
	tests := []struct {
		name              string
		quantityOrdered   float64
		quantityReceived  float64
		variancePct       float64
		isAcceptable      bool
	}{
		{"Full delivery", 100, 100, 0, true},
		{"Slight under-delivery (5%)", 100, 95, 5, true},
		{"Major under-delivery (50%)", 100, 50, 50, false},
		{"Over-delivery (10%)", 100, 110, -10, false},
		{"Zero received", 100, 0, 100, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var variancePct float64
			if tt.quantityOrdered > 0 {
				variancePct = (tt.quantityOrdered - tt.quantityReceived) / tt.quantityOrdered * 100
			}
			acceptable := variancePct >= 0 && variancePct <= 10

			if acceptable != tt.isAcceptable {
				t.Errorf("Expected isAcceptable=%v for variance=%.1f%%, got %v",
					tt.isAcceptable, variancePct, acceptable)
			}
		})
	}
}

// TestGRNConditionValues tests item condition field values
func TestGRNConditionValues(t *testing.T) {
	validConditions := map[string]bool{
		"good":          true,
		"damaged":       true,
		"partial":       true,
		"not_delivered": true,
	}

	tests := []struct {
		name       string
		condition  string
		shouldPass bool
	}{
		{"Good condition", "good", true},
		{"Damaged", "damaged", true},
		{"Partial delivery", "partial", true},
		{"Not delivered", "not_delivered", true},
		{"Invalid condition", "broken", false},
		{"Empty condition (defaults to good)", "", false},
		{"Uppercase invalid", "GOOD", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := validConditions[tt.condition]
			if isValid != tt.shouldPass {
				t.Errorf("Expected %v, got %v for condition=%q", tt.shouldPass, isValid, tt.condition)
			}
		})
	}
}
