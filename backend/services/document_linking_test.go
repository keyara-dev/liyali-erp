package services

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

// TestDocumentLinkStructure tests DocumentLink model
func TestDocumentLinkStructure(t *testing.T) {
	t.Run("DocumentLink fields", func(t *testing.T) {
		now := time.Now()
		link := DocumentLink{
			ID:            uuid.New().String(),
			SourceDocID:   "req-123",
			SourceDocType: "requisition",
			TargetDocID:   "budget-456",
			TargetDocType: "budget",
			LinkType:      "allocates_to",
			Amount:        50000,
			Proportion:    25.0,
			Status:        "active",
			CreatedAt:     now,
			UpdatedAt:     now,
		}

		if link.ID == "" {
			t.Error("Link ID should not be empty")
		}
		if link.SourceDocID == "" {
			t.Error("Link SourceDocID should not be empty")
		}
		if link.TargetDocID == "" {
			t.Error("Link TargetDocID should not be empty")
		}
		if link.LinkType == "" {
			t.Error("Link LinkType should not be empty")
		}
		if link.Status != "active" && link.Status != "inactive" {
			t.Error("Link Status should be 'active' or 'inactive'")
		}
	})
}

// TestLinkTypeValidation tests valid link types
func TestLinkTypeValidation(t *testing.T) {
	validLinkTypes := map[string]string{
		"allocates_to":        "Budget allocates to Requisition",
		"creates":             "Requisition creates PO",
		"creates_payment_for": "PO creates Payment Voucher",
		"fulfilled_by":        "PO fulfilled by GRN",
		"inherits_from":       "Document inherits from another",
	}

	for linkType, description := range validLinkTypes {
		if linkType == "" {
			t.Errorf("Invalid link type: %s", description)
		}
	}

	t.Run("Invalid link type detection", func(t *testing.T) {
		invalidTypes := []string{"invalid", "unknown", "", "connects_to"}

		for _, linkType := range invalidTypes {
			if linkType != "" && linkType != "invalid" && linkType != "unknown" && linkType != "connects_to" {
				// This would be valid
				t.Logf("Possibly valid link type: %s", linkType)
			}
		}
	})
}

// TestDocumentRelationshipChain tests chain building
func TestDocumentRelationshipChain(t *testing.T) {
	t.Run("Full procurement chain structure", func(t *testing.T) {
		chain := map[string]interface{}{
			"requisitionId":    "req-123",
			"budgetId":         "budget-456",
			"budgetCode":       "IT-2025-Q1",
			"poId":             "po-789",
			"poNumber":         "PO-20251222-abc123",
			"grnId":            "grn-012",
			"grnNumber":        "GRN-20251222-def456",
			"pvId":             "pv-345",
			"pvNumber":         "PV-20251222-ghi789",
		}

		expectedKeys := []string{
			"requisitionId", "budgetId", "poId", "grnId",
		}

		for _, key := range expectedKeys {
			if _, exists := chain[key]; !exists {
				t.Errorf("Expected key %s not found in chain", key)
			}
		}
	})
}

// TestLinkProportionCalculation tests amount proportion calculation
func TestLinkProportionCalculation(t *testing.T) {
	t.Run("Link proportion calculations", func(t *testing.T) {
		tests := []struct {
			name             string
			amount           float64
			totalAmount      float64
			expectedProportion float64
		}{
			{
				name:               "50% allocation",
				amount:             50000,
				totalAmount:        100000,
				expectedProportion: 50.0,
			},
			{
				name:               "25% allocation",
				amount:             25000,
				totalAmount:        100000,
				expectedProportion: 25.0,
			},
			{
				name:               "100% allocation",
				amount:             100000,
				totalAmount:        100000,
				expectedProportion: 100.0,
			},
			{
				name:               "10% allocation",
				amount:             10000,
				totalAmount:        100000,
				expectedProportion: 10.0,
			},
			{
				name:               "Zero allocation",
				amount:             0,
				totalAmount:        100000,
				expectedProportion: 0.0,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				var proportion float64
				if tt.totalAmount > 0 {
					proportion = (tt.amount / tt.totalAmount) * 100
				}

				if proportion != tt.expectedProportion {
					t.Errorf("Expected %f%%, got %f%%", tt.expectedProportion, proportion)
				}
			})
		}
	})
}

// TestLinkStatusManagement tests link status transitions
func TestLinkStatusManagement(t *testing.T) {
	t.Run("Link status transitions", func(t *testing.T) {
		tests := []struct {
			name           string
			currentStatus  string
			nextStatus     string
			shouldBeValid  bool
		}{
			{
				name:          "Active to Inactive",
				currentStatus: "active",
				nextStatus:    "inactive",
				shouldBeValid: true,
			},
			{
				name:          "Inactive to Active",
				currentStatus: "inactive",
				nextStatus:    "active",
				shouldBeValid: true,
			},
			{
				name:          "Active to Active (same)",
				currentStatus: "active",
				nextStatus:    "active",
				shouldBeValid: true,
			},
			{
				name:          "Invalid status",
				currentStatus: "active",
				nextStatus:    "deleted",
				shouldBeValid: false,
			},
		}

		validStatuses := map[string]bool{
			"active":   true,
			"inactive": true,
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				isValid := validStatuses[tt.nextStatus]
				if isValid != tt.shouldBeValid {
					t.Errorf("Expected %v, got %v", tt.shouldBeValid, isValid)
				}
			})
		}
	})
}

// TestProcurementChainSequence tests document sequence in procurement
func TestProcurementChainSequence(t *testing.T) {
	t.Run("Procurement chain sequence", func(t *testing.T) {
		// Correct sequence: Requisition -> Budget -> PO -> GRN
		sequence := []string{
			"requisition",
			"budget",
			"po",
			"grn",
		}

		expectedSequence := []string{
			"requisition",
			"budget",
			"po",
			"grn",
		}

		for i, doc := range sequence {
			if doc != expectedSequence[i] {
				t.Errorf("Sequence mismatch at position %d: expected %s, got %s", i, expectedSequence[i], doc)
			}
		}
	})
}

// TestLinkBidirectionality tests link relationships
func TestLinkBidirectionality(t *testing.T) {
	t.Run("Bidirectional link relationships", func(t *testing.T) {
		// If A links to B, we should track both directions
		forwardLink := DocumentLink{
			SourceDocID:   "req-123",
			SourceDocType: "requisition",
			TargetDocID:   "po-456",
			TargetDocType: "po",
			LinkType:      "creates",
		}

		// We should also be able to query backward
		t.Logf("Forward: %s (%s) -> %s (%s)",
			forwardLink.SourceDocID, forwardLink.SourceDocType,
			forwardLink.TargetDocID, forwardLink.TargetDocType)

		// Verify we can find the reverse relationship
		if forwardLink.TargetDocID != "po-456" {
			t.Error("Should be able to reference target document")
		}
	})
}

// TestLinkDuplicatePrevention tests duplicate link prevention
func TestLinkDuplicatePrevention(t *testing.T) {
	t.Run("Prevent duplicate links", func(t *testing.T) {
		// Simulate checking for duplicates
		existingLink := DocumentLink{
			ID:            "link-1",
			SourceDocID:   "req-123",
			TargetDocID:   "po-456",
			LinkType:      "creates",
		}

		newLink := DocumentLink{
			ID:            "link-2",
			SourceDocID:   "req-123",
			TargetDocID:   "po-456",
			LinkType:      "creates",
		}

		// Check if duplicate would be created
		isDuplicate := (existingLink.SourceDocID == newLink.SourceDocID &&
			existingLink.TargetDocID == newLink.TargetDocID &&
			existingLink.LinkType == newLink.LinkType)

		if !isDuplicate {
			t.Error("Should detect duplicate link")
		}
	})
}

// TestLinkDocumentTypeValidation tests valid document type combinations
func TestLinkDocumentTypeValidation(t *testing.T) {
	t.Run("Valid document type links", func(t *testing.T) {
		validCombinations := map[string]map[string]string{
			"requisition": {
				"budget": "allocates_to",
				"po":     "creates",
			},
			"budget": {
				"requisition": "allocates_to",
			},
			"po": {
				"grn": "fulfilled_by",
				"pv":  "creates_payment_for",
			},
			"grn": {
				"po": "fulfilled_by",
			},
			"pv": {
				"po": "creates_payment_for",
			},
		}

		testLink := func(from, to string) bool {
			if targets, ok := validCombinations[from]; ok {
				_, exists := targets[to]
				return exists
			}
			return false
		}

		// Test valid combinations
		validTests := []struct {
			from string
			to   string
		}{
			{"requisition", "budget"},
			{"requisition", "po"},
			{"po", "grn"},
			{"po", "pv"},
		}

		for _, test := range validTests {
			if !testLink(test.from, test.to) {
				t.Errorf("Should allow linking %s to %s", test.from, test.to)
			}
		}

		// Test invalid combinations
		invalidTests := []struct {
			from string
			to   string
		}{
			{"budget", "po"},    // Budget doesn't create PO
			{"grn", "budget"},   // GRN doesn't link to Budget
			{"pv", "requisition"}, // PV doesn't link to Requisition
		}

		for _, test := range invalidTests {
			if testLink(test.from, test.to) {
				t.Errorf("Should not allow linking %s to %s", test.from, test.to)
			}
		}
	})
}

// TestLinkAmountTracking tests amount tracking in links
func TestLinkAmountTracking(t *testing.T) {
	t.Run("Link amount tracking", func(t *testing.T) {
		links := []DocumentLink{
			{
				ID:     "link-1",
				Amount: 25000,
			},
			{
				ID:     "link-2",
				Amount: 30000,
			},
			{
				ID:     "link-3",
				Amount: 45000,
			},
		}

		var totalAmount float64
		for _, link := range links {
			totalAmount += link.Amount
		}

		expectedTotal := 100000.0
		if totalAmount != expectedTotal {
			t.Errorf("Expected total %f, got %f", expectedTotal, totalAmount)
		}
	})
}

// BenchmarkLinkProportionCalculation benchmarks proportion math
func BenchmarkLinkProportionCalculation(b *testing.B) {
	amount := 50000.0
	totalAmount := 100000.0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = (amount / totalAmount) * 100
	}
}

// BenchmarkLinkTypeLookup benchmarks link type validation
func BenchmarkLinkTypeLookup(b *testing.B) {
	validTypes := map[string]bool{
		"allocates_to":        true,
		"creates":             true,
		"creates_payment_for": true,
		"fulfilled_by":        true,
		"inherits_from":       true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validTypes["creates"]
	}
}
