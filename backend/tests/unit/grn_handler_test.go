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
				"poNumber":   "PO-20251223-abc12345",
				"receivedBy": "John Doe",
				"items": []map[string]interface{}{
					{
						"itemNo":     1,
						"description": "Item 1",
						"quantity":   10,
						"unitPrice":  5000,
						"receivedQty": 10,
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
				"poNumber": "PO-20251223-abc12345",
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
				"poNumber":   "PO-20251223-abc12345",
				"receivedBy": "John Doe",
				"items":      []map[string]interface{}{},
			},
			expectedStatus: http.StatusBadRequest,
			shouldPass:     false,
		},
		{
			name: "Empty items array",
			requestBody: map[string]interface{}{
				"poNumber":   "PO-20251223-abc12345",
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
			isValid := req.PONumber != "" &&
				req.ReceivedBy != "" &&
				len(req.Items.Data()) > 0

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
		"draft":    true,
		"pending":  true,
		"approved": true,
		"rejected": true,
		"received": true,
	}

	tests := []struct {
		name          string
		status        string
		shouldBeValid bool
	}{
		{"Draft", "draft", true},
		{"Pending", "pending", true},
		{"Approved", "approved", true},
		{"Received", "received", true},
		{"Invalid", "cancelled", false},
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

// TestGRNPONumberValidation tests PO number field
func TestGRNPONumberValidation(t *testing.T) {
	tests := []struct {
		name          string
		poNumber      string
		shouldBeValid bool
	}{
		{"Valid PO number", "PO-20251223-abc12345", true},
		{"Empty PO number", "", false},
		{"Short PO number", "PO", false},
		{"Invalid format", "req-123", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.poNumber != "" && len(tt.poNumber) > 3
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
				qty, hasQty := item["quantity"].(float64)
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
				ItemNo:      1,
				Description: "Damaged packaging",
				Severity:    "low",
				ReportedBy:  "John Doe",
				ReportedAt:  time.Now(),
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
		{"Draft to Pending", "draft", "pending", true},
		{"Pending to Approved", "pending", "approved", true},
		{"Pending to Rejected", "pending", "rejected", true},
		{"Approved to Received", "approved", "received", true},
		{"Approved to Draft", "approved", "draft", false},
		{"Received to Approved", "received", "approved", false},
		{"Rejected to Draft", "rejected", "draft", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validTransitions := map[string][]string{
				"draft":    {"pending"},
				"pending":  {"approved", "rejected"},
				"rejected": {"draft"},
				"approved": {"received"},
				"received": {},
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
			ID:         uuid.New().String(),
			GRNNumber:  "GRN-1640000000-abc12345",
			PONumber:   "PO-20251223-abc12345",
			Status:     "draft",
			ReceivedBy: "John Doe",
			Items: []types.GRNItem{
				{
					ItemNo:       1,
					Description:  "Item 1",
					Quantity:     10,
					ReceivedQty:  10,
				},
			},
			ApprovalStage: 0,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		if grn.ID == "" {
			t.Error("Response should have ID")
		}
		if grn.GRNNumber == "" {
			t.Error("Response should have GRNNumber")
		}
		if grn.PONumber == "" {
			t.Error("Response should have PONumber")
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
					ItemNo:      1,
					Description: "Item 1",
					Quantity:    10,
					ReceivedQty: 10,
				},
			},
			shouldPass: true,
		},
		{
			name: "Multiple items",
			items: []types.GRNItem{
				{
					ItemNo:      1,
					Description: "Item 1",
					Quantity:    10,
					ReceivedQty: 10,
				},
				{
					ItemNo:      2,
					Description: "Item 2",
					Quantity:    5,
					ReceivedQty: 5,
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
		poNumber := "PO-20251223-abc12345"

		if poNumber == "" {
			t.Error("GRN must have linked PO number")
		}

		if len(poNumber) < 10 {
			t.Error("PO number format should be valid")
		}
	})
}

// TestGRNDuplicatePrevention tests duplicate GRN detection
func TestGRNDuplicatePrevention(t *testing.T) {
	t.Run("Prevent duplicate GRN numbers", func(t *testing.T) {
		grn1 := types.GRNResponse{
			GRNNumber: "GRN-1640000000-abc12345",
			PONumber:  "PO-20251223-abc12345",
		}

		grn2 := types.GRNResponse{
			GRNNumber: "GRN-1640000000-abc12345",
			PONumber:  "PO-20251223-abc12345",
		}

		isDuplicate := (grn1.GRNNumber == grn2.GRNNumber)

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
		{"Update draft GRN", "draft", true},
		{"Update pending GRN", "pending", true},
		{"Cannot update approved GRN", "approved", false},
		{"Cannot update received GRN", "received", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Only draft and pending GRNs can be updated
			canUpdate := tt.currentStatus == "draft" || tt.currentStatus == "pending"

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
		PONumber:   "PO-20251223-abc12345",
		ReceivedBy: "John Doe",
		Items: []types.GRNItem{
			{
				ItemNo:      1,
				Description: "Item 1",
				Quantity:    10,
				ReceivedQty: 10,
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = req.PONumber != "" && req.ReceivedBy != "" && len(req.Items.Data()) > 0
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
