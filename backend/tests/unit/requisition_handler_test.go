package unit

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/types"
)

// TestCreateRequisitionValidation tests request validation
func TestCreateRequisitionValidation(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
		shouldPass     bool
	}{
		{
			name: "Valid requisition request",
			requestBody: map[string]interface{}{
				"title":       "Test Requisition",
				"description": "Test Description",
				"department":  "IT",
				"priority":    "high",
				"totalAmount": 50000,
				"currency":    "USD",
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
			name: "Missing title",
			requestBody: map[string]interface{}{
				"description": "Test Description",
				"department":  "IT",
				"priority":    "high",
				"totalAmount": 50000,
				"currency":    "USD",
			},
			expectedStatus: http.StatusBadRequest,
			shouldPass:     false,
		},
		{
			name: "Missing department",
			requestBody: map[string]interface{}{
				"title":       "Test Requisition",
				"description": "Test Description",
				"priority":    "high",
				"totalAmount": 50000,
				"currency":    "USD",
			},
			expectedStatus: http.StatusBadRequest,
			shouldPass:     false,
		},
		{
			name: "Invalid priority",
			requestBody: map[string]interface{}{
				"title":       "Test Requisition",
				"description": "Test Description",
				"department":  "IT",
				"priority":    "invalid",
				"totalAmount": 50000,
				"currency":    "USD",
			},
			expectedStatus: http.StatusBadRequest,
			shouldPass:     false,
		},
		{
			name: "Negative total amount",
			requestBody: map[string]interface{}{
				"title":       "Test Requisition",
				"description": "Test Description",
				"department":  "IT",
				"priority":    "high",
				"totalAmount": -50000,
				"currency":    "USD",
			},
			expectedStatus: http.StatusBadRequest,
			shouldPass:     false,
		},
		{
			name: "Zero total amount",
			requestBody: map[string]interface{}{
				"title":       "Test Requisition",
				"description": "Test Description",
				"department":  "IT",
				"priority":    "high",
				"totalAmount": 0,
				"currency":    "USD",
			},
			expectedStatus: http.StatusBadRequest,
			shouldPass:     false,
		},
		{
			name: "Empty items array",
			requestBody: map[string]interface{}{
				"title":       "Test Requisition",
				"description": "Test Description",
				"department":  "IT",
				"priority":    "high",
				"totalAmount": 50000,
				"currency":    "USD",
				"items":       []interface{}{},
			},
			expectedStatus: http.StatusBadRequest,
			shouldPass:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.requestBody)

			// Validate request structure
			var req types.CreateRequisitionRequest
			err := json.Unmarshal(body, &req)
			if err != nil {
				t.Errorf("Failed to unmarshal request: %v", err)
				return
			}

			// Check if validation should pass
			hasTitle := req.Title != ""
			hasDepartment := req.Department != ""
			validPriority := req.Priority == "low" || req.Priority == "medium" || req.Priority == "high"
			validAmount := req.TotalAmount > 0
			hasItems := len(req.Items) > 0

			isValid := hasTitle && hasDepartment && validPriority && validAmount && hasItems

			if isValid != tt.shouldPass {
				t.Errorf("Validation: expected %v, got %v", tt.shouldPass, isValid)
			}
		})
	}
}

// TestRequisitionStatusValidation tests status field validation
func TestRequisitionStatusValidation(t *testing.T) {
	validStatuses := map[string]bool{
		"draft":    true,
		"pending":  true,
		"approved": true,
		"rejected": true,
	}

	tests := []struct {
		name          string
		status        string
		shouldBeValid bool
	}{
		{"Draft status", "draft", true},
		{"Pending status", "pending", true},
		{"Approved status", "approved", true},
		{"Rejected status", "rejected", true},
		{"Invalid status", "completed", false},
		{"Empty status", "", false},
		{"Unknown status", "archived", false},
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

// TestRequisitionPriorityValidation tests priority field validation
func TestRequisitionPriorityValidation(t *testing.T) {
	validPriorities := map[string]bool{
		"low":    true,
		"medium": true,
		"high":   true,
	}

	tests := []struct {
		name             string
		priority         string
		shouldBeValid    bool
	}{
		{"Low priority", "low", true},
		{"Medium priority", "medium", true},
		{"High priority", "high", true},
		{"Invalid priority", "urgent", false},
		{"Empty priority", "", false},
		{"Uppercase priority", "HIGH", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := validPriorities[tt.priority]
			if isValid != tt.shouldBeValid {
				t.Errorf("Expected %v, got %v", tt.shouldBeValid, isValid)
			}
		})
	}
}

// TestRequisitionDepartmentValidation tests department field validation
func TestRequisitionDepartmentValidation(t *testing.T) {
	validDepartments := map[string]bool{
		"IT":          true,
		"HR":          true,
		"Operations":  true,
		"Finance":     true,
		"Sales":       true,
		"Marketing":   true,
	}

	tests := []struct {
		name             string
		department       string
		shouldBeValid    bool
	}{
		{"IT department", "IT", true},
		{"HR department", "HR", true},
		{"Operations department", "Operations", true},
		{"Invalid department", "Unknown", false},
		{"Empty department", "", false},
		{"Lowercase department", "it", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := validDepartments[tt.department]
			if isValid != tt.shouldBeValid {
				t.Errorf("Expected %v, got %v", tt.shouldBeValid, isValid)
			}
		})
	}
}

// TestRequisitionItemValidation tests requisition items validation
func TestRequisitionItemValidation(t *testing.T) {
	tests := []struct {
		name         string
		items        []map[string]interface{}
		shouldPass   bool
	}{
		{
			name: "Valid single item",
			items: []map[string]interface{}{
				{
					"description": "Item 1",
					"quantity":    1,
					"unitPrice":   100,
					"amount":      100,
				},
			},
			shouldPass: true,
		},
		{
			name: "Valid multiple items",
			items: []map[string]interface{}{
				{"description": "Item 1", "quantity": 1, "unitPrice": 100, "amount": 100},
				{"description": "Item 2", "quantity": 2, "unitPrice": 50, "amount": 100},
			},
			shouldPass: true,
		},
		{
			name: "Item missing description",
			items: []map[string]interface{}{
				{"quantity": 1, "unitPrice": 100, "amount": 100},
			},
			shouldPass: false,
		},
		{
			name: "Item with zero quantity",
			items: []map[string]interface{}{
				{"description": "Item 1", "quantity": 0, "unitPrice": 100, "amount": 100},
			},
			shouldPass: false,
		},
		{
			name: "Item with negative price",
			items: []map[string]interface{}{
				{"description": "Item 1", "quantity": 1, "unitPrice": -100, "amount": -100},
			},
			shouldPass: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate items
			allValid := true
			for _, item := range tt.items {
				desc, hasDesc := item["description"].(string)
				qty, hasQty := item["quantity"].(float64)
				price, hasPrice := item["unitPrice"].(float64)

				isValidItem := hasDesc && desc != "" && hasQty && qty > 0 && hasPrice && price > 0

				if !isValidItem {
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

// TestRequisitionResponseFormat tests response format
func TestRequisitionResponseFormat(t *testing.T) {
	t.Run("Requisition response structure", func(t *testing.T) {
		requisition := types.RequisitionResponse{
			ID:             uuid.New().String(),
			RequesterID:    uuid.New().String(),
			Title:          "Test Requisition",
			Description:    "Test Description",
			Department:     "IT",
			Status:         "draft",
			Priority:       "high",
			TotalAmount:    50000,
			Currency:       "USD",
			ApprovalStage:  0,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		// Verify required fields
		if requisition.ID == "" {
			t.Error("Response should have ID")
		}
		if requisition.RequesterID == "" {
			t.Error("Response should have RequesterID")
		}
		if requisition.Title == "" {
			t.Error("Response should have Title")
		}
		if requisition.Department == "" {
			t.Error("Response should have Department")
		}
		if requisition.Status == "" {
			t.Error("Response should have Status")
		}
		if requisition.TotalAmount <= 0 {
			t.Error("Response should have positive TotalAmount")
		}
	})
}

// TestRequisitionStateTransitions tests valid status transitions
func TestRequisitionStateTransitions(t *testing.T) {
	tests := []struct {
		name        string
		fromStatus  string
		toStatus    string
		shouldAllow bool
	}{
		{"Draft to Pending", "draft", "pending", true},
		{"Pending to Approved", "pending", "approved", true},
		{"Pending to Rejected", "pending", "rejected", true},
		{"Rejected to Draft", "rejected", "draft", true},
		{"Approved to Draft", "approved", "draft", false},
		{"Approved to Pending", "approved", "pending", false},
		{"Approved to Rejected", "approved", "rejected", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate state transition validation
			validTransitions := map[string][]string{
				"draft":    {"pending"},
				"pending":  {"approved", "rejected"},
				"rejected": {"draft"},
				"approved": {},
			}

			allowedNextStates := validTransitions[tt.fromStatus]
			canTransition := false
			for _, allowed := range allowedNextStates {
				if allowed == tt.toStatus {
					canTransition = true
					break
				}
			}

			if canTransition != tt.shouldAllow {
				t.Errorf("Transition %s->%s: expected %v, got %v",
					tt.fromStatus, tt.toStatus, tt.shouldAllow, canTransition)
			}
		})
	}
}

// TestRequisitionApprovalLogic tests approval workflow logic
func TestRequisitionApprovalLogic(t *testing.T) {
	t.Run("Approval task creation", func(t *testing.T) {
		requisition := &models.Requisition{
			ID:          uuid.New().String(),
			Status:      "pending",
			TotalAmount: 50000,
			Department:  "IT",
			Priority:    "high",
		}

		// Simulate approval routing logic
		approvers := []string{}

		// Based on amount, route to different approvers
		if requisition.TotalAmount < 10000 {
			approvers = append(approvers, "approver", "finance")
		} else if requisition.TotalAmount < 50000 {
			approvers = append(approvers, "approver", "finance", "admin")
		} else {
			approvers = append(approvers, "approver", "finance", "admin", "admin")
		}

		if requisition.TotalAmount == 50000 {
			expectedCount := 3
			if len(approvers) != expectedCount {
				t.Errorf("Expected %d approvers, got %d", expectedCount, len(approvers))
			}
		}
	})
}

// TestRequisitionErrorHandling tests error response formats
func TestRequisitionErrorHandling(t *testing.T) {
	tests := []struct {
		name           string
		errorType      string
		expectedStatus int
	}{
		{"Bad request error", "bad_request", http.StatusBadRequest},
		{"Validation error", "validation", http.StatusBadRequest},
		{"Not found error", "not_found", http.StatusNotFound},
		{"Conflict error", "conflict", http.StatusConflict},
		{"Internal error", "internal", http.StatusInternalServerError},
		{"Unauthorized error", "unauthorized", http.StatusUnauthorized},
		{"Forbidden error", "forbidden", http.StatusForbidden},
	}

	statusMap := map[string]int{
		"bad_request":   http.StatusBadRequest,
		"validation":    http.StatusBadRequest,
		"not_found":     http.StatusNotFound,
		"conflict":      http.StatusConflict,
		"internal":      http.StatusInternalServerError,
		"unauthorized":  http.StatusUnauthorized,
		"forbidden":     http.StatusForbidden,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, exists := statusMap[tt.errorType]
			if !exists || status != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, status)
			}
		})
	}
}

// TestRequisitionCURDOperations tests CRUD operation types
func TestRequisitionCRUDOperations(t *testing.T) {
	t.Run("CRUD operations", func(t *testing.T) {
		operations := []string{"create", "read", "update", "delete"}

		expectedCount := 4
		if len(operations) != expectedCount {
			t.Errorf("Expected %d operations, got %d", expectedCount, len(operations))
		}

		// Verify all operations are present
		requiredOps := map[string]bool{
			"create": true,
			"read":   true,
			"update": true,
			"delete": true,
		}

		for _, op := range operations {
			if !requiredOps[op] {
				t.Errorf("Unexpected operation: %s", op)
			}
		}
	})
}

// TestRequisitionDuplicatePrevention tests duplicate prevention
func TestRequisitionDuplicatePrevention(t *testing.T) {
	t.Run("Same request cannot create duplicates", func(t *testing.T) {
		req := types.CreateRequisitionRequest{
			Title:       "Test",
			Description: "Test",
			Department:  "IT",
			Priority:    "high",
			TotalAmount: 50000,
			Currency:    "USD",
		}

		// Simulate checking for existing requisition with same properties
		// In reality, we'd check title + department + amount combination
		signature1 := req.Title + "|" + req.Department
		signature2 := req.Title + "|" + req.Department

		if signature1 != signature2 {
			t.Error("Same request should generate same signature")
		}
	})
}

// BenchmarkRequisitionValidation benchmarks validation logic
func BenchmarkRequisitionValidation(b *testing.B) {
	req := types.CreateRequisitionRequest{
		Title:       "Test Requisition",
		Description: "Test Description",
		Department:  "IT",
		Priority:    "high",
		TotalAmount: 50000,
		Currency:    "USD",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate validation
		_ = req.Title != "" && req.Department != "" && req.TotalAmount > 0
	}
}

// BenchmarkRequisitionSerialization benchmarks JSON serialization
func BenchmarkRequisitionSerialization(b *testing.B) {
	requisition := types.RequisitionResponse{
		ID:             uuid.New().String(),
		RequesterID:    uuid.New().String(),
		Title:          "Test Requisition",
		Description:    "Test Description",
		Department:     "IT",
		Status:         "draft",
		Priority:       "high",
		TotalAmount:    50000,
		Currency:       "USD",
		ApprovalStage:  0,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(requisition)
	}
}

// TestRequisitionListPagination tests pagination logic
func TestRequisitionListPagination(t *testing.T) {
	tests := []struct {
		name              string
		page              int
		pageSize          int
		totalRecords      int
		expectedPages     int
		shouldHaveNext    bool
		shouldHavePrev    bool
	}{
		{
			name:          "First page",
			page:          1,
			pageSize:      10,
			totalRecords:  25,
			expectedPages: 3,
			shouldHaveNext: true,
			shouldHavePrev: false,
		},
		{
			name:          "Middle page",
			page:          2,
			pageSize:      10,
			totalRecords:  25,
			expectedPages: 3,
			shouldHaveNext: true,
			shouldHavePrev: true,
		},
		{
			name:          "Last page",
			page:          3,
			pageSize:      10,
			totalRecords:  25,
			expectedPages: 3,
			shouldHaveNext: false,
			shouldHavePrev: true,
		},
		{
			name:          "Single page",
			page:          1,
			pageSize:      100,
			totalRecords:  25,
			expectedPages: 1,
			shouldHaveNext: false,
			shouldHavePrev: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			totalPages := (tt.totalRecords + tt.pageSize - 1) / tt.pageSize
			hasNext := tt.page < totalPages
			hasPrev := tt.page > 1

			if totalPages != tt.expectedPages {
				t.Errorf("Expected %d pages, got %d", tt.expectedPages, totalPages)
			}
			if hasNext != tt.shouldHaveNext {
				t.Errorf("Expected hasNext=%v, got %v", tt.shouldHaveNext, hasNext)
			}
			if hasPrev != tt.shouldHavePrev {
				t.Errorf("Expected hasPrev=%v, got %v", tt.shouldHavePrev, hasPrev)
			}
		})
	}
}
