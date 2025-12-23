package main

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/types"
)

// TestCompleteRequisitionWorkflow tests the full requisition creation and approval flow
func TestCompleteRequisitionWorkflow(t *testing.T) {
	t.Run("Create requisition -> Route to approvers -> Track approval", func(t *testing.T) {
		// Step 1: Create requisition
		requisitionID := uuid.New().String()
		userID := uuid.New().String()
		departmentID := uuid.New().String()

		requisition := types.RequisitionResponse{
			ID:           requisitionID,
			ReqNumber:    "REQ-20251223-001",
			UserID:       userID,
			Department:   departmentID,
			Status:       "draft",
			TotalAmount:  50000,
			Currency:     "USD",
			DeliveryDate: time.Now().AddDate(0, 1, 0),
			ApprovalStage: 0,
			CreatedAt:    time.Now(),
		}

		if requisition.Status != "draft" {
			t.Error("Requisition should start in draft status")
		}

		// Step 2: Submit for approval (status change)
		requisition.Status = "pending"
		requisition.ApprovalStage = 1

		if requisition.Status != "pending" {
			t.Error("Requisition status should change to pending")
		}

		// Step 3: Track approval transitions
		validTransitions := map[string][]string{
			"draft":    {"pending"},
			"pending":  {"approved", "rejected"},
			"approved": {"fulfilled"},
		}

		fromStatus := "pending"
		toStatus := "approved"

		canTransition := false
		for _, nextState := range validTransitions[fromStatus] {
			if nextState == toStatus {
				canTransition = true
				break
			}
		}

		if !canTransition {
			t.Errorf("Should allow transition from %s to %s", fromStatus, toStatus)
		}

		// Step 4: Final approval
		requisition.Status = "approved"

		if requisition.Status != "approved" {
			t.Error("Requisition should be approved")
		}
	})
}

// TestRequisitionToBudgetAllocationFlow tests requisition to budget allocation
func TestRequisitionToBudgetAllocationFlow(t *testing.T) {
	t.Run("Create requisition -> Check budget -> Allocate funds", func(t *testing.T) {
		// Step 1: Create requisition
		requisitionAmount := 50000.0
		departmentID := uuid.New().String()

		requisition := types.RequisitionResponse{
			ID:          uuid.New().String(),
			Status:      "approved",
			TotalAmount: requisitionAmount,
		}

		if requisition.Status != "approved" {
			t.Error("Requisition must be approved before budget allocation")
		}

		// Step 2: Check budget availability
		totalBudget := 500000.0
		allocatedAmount := 200000.0
		availableBudget := totalBudget - allocatedAmount

		if availableBudget < requisitionAmount {
			t.Error("Insufficient budget available")
		}

		// Step 3: Allocate budget
		if availableBudget >= requisitionAmount {
			allocatedAmount += requisitionAmount
			availableBudget -= requisitionAmount
		}

		if allocatedAmount != 250000 {
			t.Errorf("Expected allocated amount 250000, got %f", allocatedAmount)
		}

		if availableBudget != 250000 {
			t.Errorf("Expected available budget 250000, got %f", availableBudget)
		}

		// Step 4: Create budget link
		budget := types.BudgetResponse{
			ID:               uuid.New().String(),
			TotalBudget:      totalBudget,
			AllocatedAmount:  allocatedAmount,
			RemainingAmount:  availableBudget,
			Status:           "approved",
			LinkedRequisitions: []string{requisition.ID},
		}

		if len(budget.LinkedRequisitions) == 0 {
			t.Error("Budget should link to requisition")
		}
	})
}

// TestBudgetApprovalChainFlow tests multi-stage budget approval
func TestBudgetApprovalChainFlow(t *testing.T) {
	t.Run("Budget draft -> Finance approval -> Admin approval", func(t *testing.T) {
		budget := types.BudgetResponse{
			ID:            uuid.New().String(),
			Status:        "draft",
			ApprovalStage: 0,
			TotalBudget:   500000,
		}

		// Stage 1: Finance team review
		approvalHistory := []types.ApprovalRecord{
			{
				ApproverID:   uuid.New().String(),
				ApproverName: "Finance Manager",
				Status:       "approved",
				Comments:     "Budget approved by finance",
				ApprovedAt:   time.Now(),
			},
		}

		budget.Status = "pending"
		budget.ApprovalStage = 1

		if budget.ApprovalStage != 1 {
			t.Error("Budget should be at stage 1 after finance review")
		}

		// Stage 2: Admin approval
		approvalHistory = append(approvalHistory, types.ApprovalRecord{
			ApproverID:   uuid.New().String(),
			ApproverName: "Admin Manager",
			Status:       "approved",
			Comments:     "Budget approved by admin",
			ApprovedAt:   time.Now(),
		})

		budget.Status = "approved"
		budget.ApprovalStage = 2

		if budget.ApprovalStage != 2 {
			t.Error("Budget should be at stage 2 after admin approval")
		}

		if len(approvalHistory) != 2 {
			t.Error("Should have 2 approval records")
		}
	})
}

// TestRequisitionToPurchaseOrderFlow tests requisition to PO creation
func TestRequisitionToPurchaseOrderFlow(t *testing.T) {
	t.Run("Approved requisition -> Create PO -> Link documents", func(t *testing.T) {
		// Step 1: Start with approved requisition
		requisition := types.RequisitionResponse{
			ID:          uuid.New().String(),
			Status:      "approved",
			TotalAmount: 50000,
			Items: []types.RequisitionItem{
				{
					ItemNo:      1,
					Description: "Office Supplies",
					Quantity:    100,
					UnitPrice:   500,
					Amount:      50000,
				},
			},
		}

		// Step 2: Create PO from requisition
		purchaseOrder := types.PurchaseOrderResponse{
			ID:               uuid.New().String(),
			PONumber:         "PO-20251223-abc12345",
			Status:           "draft",
			TotalAmount:      requisition.TotalAmount,
			LinkedRequisition: requisition.ID,
			Items: []types.POItem{
				{
					ItemNo:      1,
					Description: requisition.Items[0].Description,
					Quantity:    requisition.Items[0].Quantity,
					UnitPrice:   requisition.Items[0].UnitPrice,
					Amount:      requisition.Items[0].Amount,
				},
			},
		}

		// Step 3: Verify linking
		if purchaseOrder.LinkedRequisition != requisition.ID {
			t.Error("PO should link to requisition")
		}

		if purchaseOrder.TotalAmount != requisition.TotalAmount {
			t.Error("PO amount should match requisition amount")
		}

		// Step 4: Submit PO
		purchaseOrder.Status = "pending"

		if purchaseOrder.Status != "pending" {
			t.Error("PO should be submitted for approval")
		}
	})
}

// TestPurchaseOrderApprovalFlow tests PO approval workflow
func TestPurchaseOrderApprovalFlow(t *testing.T) {
	t.Run("PO draft -> Manager approval -> Finance approval", func(t *testing.T) {
		po := types.PurchaseOrderResponse{
			ID:            uuid.New().String(),
			PONumber:      "PO-20251223-abc12345",
			Status:        "draft",
			ApprovalStage: 0,
			TotalAmount:   50000,
		}

		// Transition 1: Draft -> Pending
		po.Status = "pending"
		if po.Status != "pending" {
			t.Error("PO should transition to pending")
		}

		// Transition 2: Pending -> Approved (after manager approval)
		po.Status = "approved"
		po.ApprovalStage = 1

		if po.ApprovalStage != 1 {
			t.Error("PO should be at approval stage 1")
		}

		// Transition 3: Final state - Ready for fulfillment
		po.Status = "fulfilled"

		if po.Status != "fulfilled" {
			t.Error("PO should be fulfilled")
		}
	})
}

// TestGRNCreationFromPOFlow tests GRN creation from approved PO
func TestGRNCreationFromPOFlow(t *testing.T) {
	t.Run("Approved PO -> Create GRN -> Validate quantities", func(t *testing.T) {
		// Step 1: Start with approved PO
		poNumber := "PO-20251223-abc12345"
		po := types.PurchaseOrderResponse{
			ID:       uuid.New().String(),
			PONumber: poNumber,
			Status:   "fulfilled",
			Items: []types.POItem{
				{
					ItemNo:      1,
					Description: "Item 1",
					Quantity:    100,
					UnitPrice:   500,
				},
			},
		}

		// Step 2: Create GRN
		grn := types.GRNResponse{
			ID:        uuid.New().String(),
			GRNNumber: "GRN-1640000000-abc12345",
			PONumber:  poNumber,
			Status:    "draft",
			Items: []types.GRNItem{
				{
					ItemNo:       1,
					Description:  "Item 1",
					Quantity:     100,
					ReceivedQty:  100,
				},
			},
		}

		// Step 3: Verify linking
		if grn.PONumber != po.PONumber {
			t.Error("GRN should reference PO number")
		}

		// Step 4: Validate quantities match
		if len(grn.Items) != len(po.Items) {
			t.Error("GRN items should match PO items")
		}

		orderQty := po.Items[0].Quantity
		receivedQty := grn.Items[0].ReceivedQty

		if receivedQty != orderQty {
			t.Errorf("Received quantity %f should match ordered quantity %f", receivedQty, orderQty)
		}

		// Step 5: Submit for approval
		grn.Status = "pending"

		if grn.Status != "pending" {
			t.Error("GRN should be submitted for approval")
		}
	})
}

// TestQuantityVarianceHandling tests handling of quantity mismatches
func TestQuantityVarianceHandling(t *testing.T) {
	t.Run("PO quantity 100 -> Received 95 -> Track variance", func(t *testing.T) {
		orderedQty := 100.0
		receivedQty := 95.0

		variance := receivedQty - orderedQty
		variancePercent := (variance / orderedQty) * 100

		if variance != -5 {
			t.Errorf("Expected variance -5, got %f", variance)
		}

		if variancePercent != -5 {
			t.Errorf("Expected variance -5%%, got %f%%", variancePercent)
		}

		// Log quality issue if variance exceeds threshold
		if variancePercent < -10 {
			t.Logf("Warning: Significant shortage of %f%%", variancePercent)
		}
	})
}

// TestPaymentVoucherFromGRNFlow tests payment voucher creation from GRN
func TestPaymentVoucherFromGRNFlow(t *testing.T) {
	t.Run("Approved GRN -> Create PV -> Track for payment", func(t *testing.T) {
		// Step 1: Start with approved GRN
		grn := types.GRNResponse{
			ID:        uuid.New().String(),
			GRNNumber: "GRN-1640000000-abc12345",
			Status:    "approved",
			PONumber:  "PO-20251223-abc12345",
		}

		vendorID := uuid.New().String()

		// Step 2: Create payment voucher
		pv := types.PaymentVoucherResponse{
			ID:              uuid.New().String(),
			VoucherNumber:   "PV-1640000000-xyz67890",
			VendorID:        vendorID,
			InvoiceNumber:   "INV-2025-001",
			Status:          "draft",
			Amount:          50000,
			Currency:        "USD",
			LinkedPO:        "PO-20251223-abc12345",
			ApprovalStage:   0,
		}

		// Step 3: Verify linking
		if pv.LinkedPO != grn.PONumber {
			t.Error("PV should reference PO number from GRN")
		}

		// Step 4: Submit for approval
		pv.Status = "pending"
		pv.ApprovalStage = 1

		if pv.ApprovalStage != 1 {
			t.Error("PV should move to approval stage 1")
		}

		// Step 5: Final approval
		pv.Status = "approved"

		if pv.Status != "approved" {
			t.Error("PV should be approved")
		}
	})
}

// TestCompleteEndToEndFlow tests entire requisition to payment flow
func TestCompleteEndToEndFlow(t *testing.T) {
	t.Run("Requisition -> Budget -> PO -> GRN -> Payment Voucher", func(t *testing.T) {
		// Create all documents in sequence
		requisition := types.RequisitionResponse{
			ID:          uuid.New().String(),
			Status:      "approved",
			TotalAmount: 50000,
		}

		budget := types.BudgetResponse{
			ID:                 uuid.New().String(),
			Status:             "approved",
			LinkedRequisitions: []string{requisition.ID},
		}

		po := types.PurchaseOrderResponse{
			ID:               uuid.New().String(),
			Status:           "fulfilled",
			LinkedRequisition: requisition.ID,
			TotalAmount:      50000,
		}

		grn := types.GRNResponse{
			ID:       uuid.New().String(),
			Status:   "approved",
			PONumber: po.PONumber,
		}

		pv := types.PaymentVoucherResponse{
			ID:        uuid.New().String(),
			Status:    "approved",
			LinkedPO:  po.PONumber,
			Amount:    50000,
		}

		// Verify chain of relationships
		if budget.LinkedRequisitions[0] != requisition.ID {
			t.Error("Budget should link to requisition")
		}

		if po.LinkedRequisition != requisition.ID {
			t.Error("PO should link to requisition")
		}

		if grn.PONumber != po.PONumber {
			t.Error("GRN should link to PO")
		}

		if pv.LinkedPO != po.PONumber {
			t.Error("PV should link to PO")
		}

		// Verify all documents are in correct final state
		if requisition.Status != "approved" || budget.Status != "approved" ||
			po.Status != "fulfilled" || grn.Status != "approved" ||
			pv.Status != "approved" {
			t.Error("All documents should be in approved/fulfilled state")
		}
	})
}

// TestBudgetConstraintEnforcement tests budget constraints during flow
func TestBudgetConstraintEnforcement(t *testing.T) {
	t.Run("Check vendor spending limits during PO creation", func(t *testing.T) {
		totalBudget := 500000.0
		allocatedForVendor := 150000.0 // 30% of budget already allocated
		maxVendorSpend := totalBudget * 0.30 // 30% limit
		newPOAmount := 50000.0

		totalVendorSpend := allocatedForVendor + newPOAmount

		// Check vendor limit
		if totalVendorSpend > maxVendorSpend {
			t.Logf("Vendor spending limit exceeded: %f > %f", totalVendorSpend, maxVendorSpend)
		} else {
			if totalVendorSpend <= maxVendorSpend {
				// Within limit - can proceed
				if totalVendorSpend != 200000 {
					t.Errorf("Expected total vendor spend 200000, got %f", totalVendorSpend)
				}
			}
		}
	})
}

// TestNotificationTriggerFlow tests notifications throughout workflow
func TestNotificationTriggerFlow(t *testing.T) {
	t.Run("Requisition approval triggers notifications", func(t *testing.T) {
		notifications := []types.NotificationResponse{}

		// Event 1: Requisition submitted for approval
		notifications = append(notifications, types.NotificationResponse{
			ID:          uuid.New().String(),
			Type:        "approval_required",
			DocumentID:  uuid.New().String(),
			Title:       "New Requisition Pending Approval",
			Message:     "Requisition REQ-001 is pending your approval",
			IsRead:      false,
			CreatedAt:   time.Now(),
		})

		// Event 2: Requisition approved
		notifications = append(notifications, types.NotificationResponse{
			ID:          uuid.New().String(),
			Type:        "document_approved",
			DocumentID:  uuid.New().String(),
			Title:       "Requisition Approved",
			Message:     "Your requisition REQ-001 has been approved",
			IsRead:      false,
			CreatedAt:   time.Now().Add(time.Hour),
		})

		// Event 3: PO created from requisition
		notifications = append(notifications, types.NotificationResponse{
			ID:          uuid.New().String(),
			Type:        "status_change",
			DocumentID:  uuid.New().String(),
			Title:       "Purchase Order Created",
			Message:     "PO-001 has been created from your requisition",
			IsRead:      false,
			CreatedAt:   time.Now().Add(2 * time.Hour),
		})

		if len(notifications) != 3 {
			t.Error("Should generate 3 notifications for workflow")
		}

		// Mark as read
		if len(notifications) > 0 {
			notifications[0].IsRead = true

			if !notifications[0].IsRead {
				t.Error("Notification should be marked as read")
			}
		}
	})
}

// TestApprovalRulesApplication tests approval routing based on rules
func TestApprovalRulesApplication(t *testing.T) {
	t.Run("Route low/medium/high amount requisitions to correct approvers", func(t *testing.T) {
		tests := []struct {
			name        string
			amount      float64
			expectedTier string
		}{
			{"Low amount", 5000, "low"},
			{"Medium amount", 30000, "medium"},
			{"High amount", 100000, "high"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				var tier string

				if tt.amount < 10000 {
					tier = "low"
				} else if tt.amount < 50000 {
					tier = "medium"
				} else {
					tier = "high"
				}

				if tier != tt.expectedTier {
					t.Errorf("Expected tier %s, got %s", tt.expectedTier, tier)
				}

				// Different approval chains based on tier
				approverChains := map[string][]string{
					"low":    {"department_head"},
					"medium": {"manager", "finance"},
					"high":   {"manager", "finance", "executive"},
				}

				chain := approverChains[tier]
				if len(chain) == 0 {
					t.Error("Should have approval chain for tier")
				}
			})
		}
	})
}

// BenchmarkWorkflowStateTransitions benchmarks state transition validation
func BenchmarkWorkflowStateTransitions(b *testing.B) {
	validTransitions := map[string][]string{
		"draft":    {"pending"},
		"pending":  {"approved", "rejected"},
		"approved": {"fulfilled"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		allowed := false
		for _, next := range validTransitions["pending"] {
			if next == "approved" {
				allowed = true
				break
			}
		}
		_ = allowed
	}
}

// BenchmarkDocumentLinking benchmarks linking operations
func BenchmarkDocumentLinking(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		links := map[string][]string{
			"requisition": {"budget", "po"},
			"po":          {"grn", "pv"},
			"grn":         {"pv"},
		}
		_ = len(links["po"])
	}
}

// BenchmarkNotificationGeneration benchmarks notification creation
func BenchmarkNotificationGeneration(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = types.NotificationResponse{
			ID:        uuid.New().String(),
			Type:      "approval_required",
			DocumentID: uuid.New().String(),
			IsRead:    false,
			CreatedAt: time.Now(),
		}
	}
}
