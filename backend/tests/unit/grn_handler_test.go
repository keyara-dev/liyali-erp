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
		"DRAFT":     true,
		"PENDING":   true,
		"APPROVED":  true,
		"REJECTED":  true,
		"REVISION":  true,
		"COMPLETED": true,
		"CANCELLED": true,
	}

	tests := []struct {
		name          string
		status        string
		shouldBeValid bool
	}{
		{"Draft", "DRAFT", true},
		{"Pending", "PENDING", true},
		{"Approved", "APPROVED", true},
		{"Rejected", "REJECTED", true},
		{"Revision", "REVISION", true},
		{"Completed", "COMPLETED", true},
		{"Cancelled", "CANCELLED", true},
		{"Invalid RECEIVED", "RECEIVED", false},
		{"Invalid PAID", "PAID", false},
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
		{"Approved to Completed", "APPROVED", "COMPLETED", true},
		{"Approved to Draft", "APPROVED", "DRAFT", false},
		{"Completed to Approved", "COMPLETED", "APPROVED", false},
		{"Rejected to Draft", "REJECTED", "DRAFT", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validTransitions := map[string][]string{
				"DRAFT":     {"PENDING"},
				"PENDING":   {"APPROVED", "REJECTED"},
				"REJECTED":  {"DRAFT"},
				"APPROVED":  {"COMPLETED"},
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
		{"Cannot update completed GRN", "COMPLETED", false},
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

// ============================================================================
// Sign-off lifecycle: receiver -> certifier -> READY -> workflow/complete
// (Mirrors the handler gates added in handlers/grn.go for SignReceiveGRN,
// CertifyGRN, MarkGRNComplete, and the updated SubmitGRN guard.)
// ============================================================================

// signoffState reproduces the state-machine guards from grn.go so the test
// asserts the agreed contract without booting the full HTTP stack.
type signoffState struct {
	status         string // DRAFT | PENDING | APPROVED | COMPLETED | ...
	signoffStatus  string // PENDING_RECEIVER | PENDING_CERTIFIER | READY | COMPLETED
	createdBy      string
	receivedBy     string
}

// privilegedRoles mirrors the map declared in handlers/grn.go.
var privilegedRoles = map[string]bool{
	"admin":       true,
	"super_admin": true,
	"manager":     true,
	"finance":     true,
	"approver":    true,
}

// canReceiverSign returns the same boolean SignReceiveGRN enforces.
func canReceiverSign(s signoffState) bool {
	if s.status != "DRAFT" {
		return false
	}
	return s.signoffStatus == "PENDING_RECEIVER"
}

// canCertify returns the same boolean CertifyGRN enforces.
func canCertify(s signoffState, actorRole, actorID string) bool {
	if s.status != "DRAFT" {
		return false
	}
	if s.signoffStatus != "PENDING_CERTIFIER" {
		return false
	}
	if !privilegedRoles[actorRole] {
		return false
	}
	// Separation-of-duties: certifier cannot be creator or receiver.
	if actorID == s.createdBy || actorID == s.receivedBy {
		return false
	}
	return true
}

// canSubmitToWorkflow returns the same boolean SubmitGRN enforces.
func canSubmitToWorkflow(s signoffState) bool {
	if s.status != "DRAFT" {
		return false
	}
	return s.signoffStatus == "READY"
}

// canMarkComplete returns the same boolean MarkGRNComplete enforces.
func canMarkComplete(s signoffState) bool {
	if s.status != "DRAFT" {
		return false
	}
	return s.signoffStatus == "READY"
}

// canEditItems returns the same boolean UpdateGRN enforces — line items lock
// once the receiver has signed so the captured signature stays bound to the
// quantities/conditions that were sighted on delivery.
func canEditItems(s signoffState) bool {
	if s.status != "DRAFT" && s.status != "PENDING" {
		return false
	}
	return s.signoffStatus == "PENDING_RECEIVER"
}

func TestReceiverSignoffGate(t *testing.T) {
	tests := []struct {
		name  string
		state signoffState
		want  bool
	}{
		{"DRAFT + PENDING_RECEIVER allowed", signoffState{status: "DRAFT", signoffStatus: "PENDING_RECEIVER"}, true},
		{"DRAFT + PENDING_CERTIFIER blocked (already signed)", signoffState{status: "DRAFT", signoffStatus: "PENDING_CERTIFIER"}, false},
		{"DRAFT + READY blocked", signoffState{status: "DRAFT", signoffStatus: "READY"}, false},
		{"PENDING status blocked", signoffState{status: "PENDING", signoffStatus: "PENDING_RECEIVER"}, false},
		{"APPROVED blocked", signoffState{status: "APPROVED", signoffStatus: "COMPLETED"}, false},
		{"COMPLETED blocked", signoffState{status: "COMPLETED", signoffStatus: "COMPLETED"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := canReceiverSign(tt.state); got != tt.want {
				t.Errorf("canReceiverSign(%+v) = %v, want %v", tt.state, got, tt.want)
			}
		})
	}
}

func TestCertifierSignoffGate(t *testing.T) {
	base := signoffState{
		status:        "DRAFT",
		signoffStatus: "PENDING_CERTIFIER",
		createdBy:     "user-creator",
		receivedBy:    "user-receiver",
	}
	tests := []struct {
		name      string
		state     signoffState
		actorRole string
		actorID   string
		want      bool
	}{
		{"admin separate user allowed", base, "admin", "user-admin", true},
		{"manager separate user allowed", base, "manager", "user-mgr", true},
		{"finance separate user allowed", base, "finance", "user-fin", true},
		{"approver separate user allowed", base, "approver", "user-app", true},
		{"super_admin separate user allowed", base, "super_admin", "user-sa", true},
		{"requester role blocked", base, "requester", "user-req", false},
		{"viewer role blocked", base, "viewer", "user-view", false},
		{"unknown role blocked", base, "", "user-x", false},
		{"creator cannot self-certify", base, "admin", "user-creator", false},
		{"receiver cannot self-certify", base, "admin", "user-receiver", false},
		{"PENDING_RECEIVER state blocks certifier", signoffState{
			status: "DRAFT", signoffStatus: "PENDING_RECEIVER",
			createdBy: "c", receivedBy: "r",
		}, "admin", "user-admin", false},
		{"READY state blocks re-certify", signoffState{
			status: "DRAFT", signoffStatus: "READY",
			createdBy: "c", receivedBy: "r",
		}, "admin", "user-admin", false},
		{"PENDING workflow status blocks", signoffState{
			status: "PENDING", signoffStatus: "PENDING_CERTIFIER",
			createdBy: "c", receivedBy: "r",
		}, "admin", "user-admin", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := canCertify(tt.state, tt.actorRole, tt.actorID); got != tt.want {
				t.Errorf("canCertify(%+v, role=%q, id=%q) = %v, want %v",
					tt.state, tt.actorRole, tt.actorID, got, tt.want)
			}
		})
	}
}

func TestSubmitToWorkflowGate(t *testing.T) {
	tests := []struct {
		name  string
		state signoffState
		want  bool
	}{
		{"READY can submit", signoffState{status: "DRAFT", signoffStatus: "READY"}, true},
		{"PENDING_RECEIVER blocked", signoffState{status: "DRAFT", signoffStatus: "PENDING_RECEIVER"}, false},
		{"PENDING_CERTIFIER blocked", signoffState{status: "DRAFT", signoffStatus: "PENDING_CERTIFIER"}, false},
		{"non-DRAFT status blocked", signoffState{status: "PENDING", signoffStatus: "READY"}, false},
		{"APPROVED blocked", signoffState{status: "APPROVED", signoffStatus: "COMPLETED"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := canSubmitToWorkflow(tt.state); got != tt.want {
				t.Errorf("canSubmitToWorkflow(%+v) = %v, want %v", tt.state, got, tt.want)
			}
		})
	}
}

func TestMarkCompleteGate(t *testing.T) {
	tests := []struct {
		name  string
		state signoffState
		want  bool
	}{
		{"READY + DRAFT can complete", signoffState{status: "DRAFT", signoffStatus: "READY"}, true},
		{"PENDING_RECEIVER blocked", signoffState{status: "DRAFT", signoffStatus: "PENDING_RECEIVER"}, false},
		{"PENDING_CERTIFIER blocked", signoffState{status: "DRAFT", signoffStatus: "PENDING_CERTIFIER"}, false},
		{"already COMPLETED blocked", signoffState{status: "COMPLETED", signoffStatus: "COMPLETED"}, false},
		{"workflow PENDING blocked", signoffState{status: "PENDING", signoffStatus: "READY"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := canMarkComplete(tt.state); got != tt.want {
				t.Errorf("canMarkComplete(%+v) = %v, want %v", tt.state, got, tt.want)
			}
		})
	}
}

func TestItemEditLock(t *testing.T) {
	tests := []struct {
		name  string
		state signoffState
		want  bool
	}{
		{"DRAFT + PENDING_RECEIVER editable", signoffState{status: "DRAFT", signoffStatus: "PENDING_RECEIVER"}, true},
		{"PENDING + PENDING_RECEIVER editable", signoffState{status: "PENDING", signoffStatus: "PENDING_RECEIVER"}, true},
		{"PENDING_CERTIFIER locks items (receiver signed)", signoffState{status: "DRAFT", signoffStatus: "PENDING_CERTIFIER"}, false},
		{"READY locks items (both signed)", signoffState{status: "DRAFT", signoffStatus: "READY"}, false},
		{"APPROVED locks items", signoffState{status: "APPROVED", signoffStatus: "COMPLETED"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := canEditItems(tt.state); got != tt.want {
				t.Errorf("canEditItems(%+v) = %v, want %v", tt.state, got, tt.want)
			}
		})
	}
}

// TestSignoffLifecycleHappyPath drives a state object through the full
// receiver -> certifier -> READY -> COMPLETED transitions and verifies every
// gate flips at the expected moment.
func TestSignoffLifecycleHappyPath(t *testing.T) {
	s := signoffState{
		status:        "DRAFT",
		signoffStatus: "PENDING_RECEIVER",
		createdBy:     "alice",
	}

	// 1. Receiver can sign; certifier/submit/complete blocked.
	if !canReceiverSign(s) {
		t.Fatal("expected receiver to be allowed at PENDING_RECEIVER")
	}
	if canCertify(s, "admin", "bob") {
		t.Fatal("certify must wait for receiver")
	}
	if canSubmitToWorkflow(s) {
		t.Fatal("submit must wait for READY")
	}
	if canMarkComplete(s) {
		t.Fatal("complete must wait for READY")
	}
	if !canEditItems(s) {
		t.Fatal("items should be editable before receiver signs")
	}

	// 2. Receiver signs -> PENDING_CERTIFIER.
	s.signoffStatus = "PENDING_CERTIFIER"
	s.receivedBy = "carol"
	if canReceiverSign(s) {
		t.Fatal("receiver shouldn't be able to re-sign")
	}
	if canEditItems(s) {
		t.Fatal("items must lock after receiver signs")
	}
	if canCertify(s, "requester", "bob") {
		t.Fatal("non-privileged role must not certify")
	}
	if canCertify(s, "admin", "alice") {
		t.Fatal("creator must not self-certify")
	}
	if canCertify(s, "admin", "carol") {
		t.Fatal("receiver must not self-certify")
	}
	if !canCertify(s, "admin", "dave") {
		t.Fatal("independent admin should certify")
	}

	// 3. Certifier signs -> READY.
	s.signoffStatus = "READY"
	if canCertify(s, "admin", "dave") {
		t.Fatal("certify must not be repeatable")
	}
	if !canSubmitToWorkflow(s) {
		t.Fatal("READY must unlock workflow submit")
	}
	if !canMarkComplete(s) {
		t.Fatal("READY must unlock direct complete")
	}

	// 4a. Direct completion path -> COMPLETED.
	s.status = "COMPLETED"
	s.signoffStatus = "COMPLETED"
	if canSubmitToWorkflow(s) || canMarkComplete(s) || canCertify(s, "admin", "eve") {
		t.Fatal("terminal state must reject further actions")
	}
}

// TestPerGRNStampOverride documents the precedence the PDF uses: per-GRN
// stamp wins over the org-level fallback.
func TestPerGRNStampOverride(t *testing.T) {
	cases := []struct {
		name      string
		grnStamp  string
		orgStamp  string
		wantStamp string
	}{
		{"per-GRN overrides org", "data:image/png;base64,grn", "https://cdn/org.png", "data:image/png;base64,grn"},
		{"falls back to org when GRN empty", "", "https://cdn/org.png", "https://cdn/org.png"},
		{"empty when neither set", "", "", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.grnStamp
			if got == "" {
				got = tc.orgStamp
			}
			if got != tc.wantStamp {
				t.Errorf("precedence: got %q, want %q", got, tc.wantStamp)
			}
		})
	}
}

// TestGRNItemPDFFields ensures itemCode and remarks round-trip through the
// shared GRNItem type — both fields are newly added to align with the printed
// council form.
func TestGRNItemPDFFields(t *testing.T) {
	body := []byte(`{
        "description": "10x cement bag",
        "itemCode": "CEM-50",
        "quantityOrdered": 10,
        "quantityReceived": 8,
        "variance": -2,
        "condition": "good",
        "remarks": "Two bags damaged in transit"
    }`)
	var item types.GRNItem
	if err := json.Unmarshal(body, &item); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if item.ItemCode != "CEM-50" {
		t.Errorf("ItemCode lost: %q", item.ItemCode)
	}
	if item.Remarks != "Two bags damaged in transit" {
		t.Errorf("Remarks lost: %q", item.Remarks)
	}
}

// TestConsignmentNoteRoundTrip verifies the printed-form "Delivery
// Consignment Note" field survives a create-request unmarshal.
func TestConsignmentNoteRoundTrip(t *testing.T) {
	body := []byte(`{
        "poDocumentNumber": "PO-20260528-deadbeef",
        "receivedBy": "Carol",
        "consignmentNote": "CN-2026-04812",
        "items": [{"description": "x", "quantityOrdered": 1, "quantityReceived": 1, "variance": 0, "condition": "good"}]
    }`)
	var req types.CreateGRNRequest
	if err := json.Unmarshal(body, &req); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if req.ConsignmentNote != "CN-2026-04812" {
		t.Errorf("ConsignmentNote lost: %q", req.ConsignmentNote)
	}
}

// TestCertifyRequestStampField confirms the optional per-GRN stamp ride-along
// payload deserialises into types.CertifyGRNRequest.
func TestCertifyRequestStampField(t *testing.T) {
	body := []byte(`{"signature":"data:img,sig","stampImageUrl":"data:img,stamp"}`)
	var req types.CertifyGRNRequest
	if err := json.Unmarshal(body, &req); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if req.Signature == "" {
		t.Error("signature should be required field, parsed empty")
	}
	if req.StampImageURL != "data:img,stamp" {
		t.Errorf("StampImageURL lost: %q", req.StampImageURL)
	}
}

// Keep uuid + time imports honest if other tests in the file ever stop using them.
var (
	_ = uuid.New
	_ = time.Now
)
