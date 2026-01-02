package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/handlers"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/services"
	"github.com/liyali/liyali-gateway/types"
	"gorm.io/datatypes"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupIntegrationTestDB creates a test database for integration tests
func setupIntegrationTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Auto-migrate all models
	db.AutoMigrate(
		&models.Organization{},
		&models.User{},
		&models.Vendor{},
		&models.Category{},
		&models.Requisition{},
		&models.PurchaseOrder{},
		&models.GoodsReceivedNote{},
		&models.PaymentVoucher{},
	)

	return db
}

// createIntegrationTestData creates comprehensive test data
func createIntegrationTestData(db *gorm.DB) (string, string, string, string) {
	// Create organization
	org := models.Organization{
		ID:   uuid.New().String(),
		Name: "Integration Test Org",
	}
	db.Create(&org)

	// Create user
	user := models.User{
		ID:             uuid.New().String(),
		Email:          "integrationtest@example.com",
		Name:           "Integration Test User",
		OrganizationID: org.ID,
		Role:           "admin",
	}
	db.Create(&user)

	// Create vendor
	vendor := models.Vendor{
		ID:             uuid.New().String(),
		Name:           "Integration Test Vendor",
		Email:          "vendor@integration.com",
		OrganizationID: org.ID,
	}
	db.Create(&vendor)

	// Create category
	category := models.Category{
		ID:             uuid.New().String(),
		Name:           "Integration Test Category",
		OrganizationID: org.ID,
	}
	db.Create(&category)

	return org.ID, user.ID, vendor.ID, category.ID
}

// setupFiberApp creates a Fiber app with test routes
func setupFiberApp(db *gorm.DB) *fiber.App {
	app := fiber.New()

	// Set up config
	config.DB = db

	// Add middleware to set user context
	app.Use(func(c *fiber.Ctx) error {
		// Mock authentication - set a test user ID
		c.Locals("user_id", "test-user-id")
		return c.Next()
	})

	// Add routes
	app.Post("/api/v1/requisitions", handlers.CreateRequisition)
	app.Post("/api/v1/requisitions/:id/approve", handlers.ApproveRequisition)
	app.Post("/api/v1/purchase-orders/:id/approve", handlers.ApprovePurchaseOrder)
	app.Post("/api/v1/grns/:id/approve", handlers.ApproveGRN)

	return app
}

// TestCompleteAutomationWorkflow tests the full automation from requisition to payment voucher
func TestCompleteAutomationWorkflow(t *testing.T) {
	db := setupIntegrationTestDB()
	app := setupFiberApp(db)

	orgID, userID, vendorID, categoryID := createIntegrationTestData(db)

	// Update the test user ID to match our created user
	db.Model(&models.User{}).Where("id = ?", userID).Update("id", "test-user-id")

	t.Run("Complete workflow: Requisition -> PO -> GRN -> PV", func(t *testing.T) {
		// Step 1: Create and approve requisition
		requisitionData := types.CreateRequisitionRequest{
			Title:             "Integration Test Requisition",
			Description:       "Testing complete automation workflow",
			Department:        "IT",
			Priority:          "medium",
			TotalAmount:       100000,
			Currency:          "USD",
			CategoryID:        &categoryID,
			PreferredVendorID: &vendorID,
			Items: []types.RequisitionItem{
				{
					ItemNo:      1,
					Description: "Test Item 1",
					Quantity:    10,
					UnitPrice:   5000,
					Amount:      50000,
					Category:    "IT Equipment",
				},
				{
					ItemNo:      2,
					Description: "Test Item 2",
					Quantity:    20,
					UnitPrice:   2500,
					Amount:      50000,
					Category:    "IT Equipment",
				},
			},
		}

		// Create requisition
		reqBody, _ := json.Marshal(requisitionData)
		req := httptest.NewRequest("POST", "/api/v1/requisitions", bytes.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Failed to create requisition: %v", err)
		}

		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("Expected status 201, got %d", resp.StatusCode)
		}

		// Parse response to get requisition ID
		var createResp types.DetailResponse
		json.NewDecoder(resp.Body).Decode(&createResp)
		
		requisitionResp, ok := createResp.Data.(map[string]interface{})
		if !ok {
			t.Fatal("Failed to parse requisition response")
		}

		requisitionID := requisitionResp["id"].(string)

		// Step 2: Approve requisition (should auto-create PO)
		approveData := types.ApproveDocumentRequest{
			Comments:  "Approved for automation test",
			Signature: "test-signature-123",
		}

		approveBody, _ := json.Marshal(approveData)
		approveReq := httptest.NewRequest("POST", 
			fmt.Sprintf("/api/v1/requisitions/%s/approve", requisitionID), 
			bytes.NewReader(approveBody))
		approveReq.Header.Set("Content-Type", "application/json")

		approveResp, err := app.Test(approveReq)
		if err != nil {
			t.Fatalf("Failed to approve requisition: %v", err)
		}

		if approveResp.StatusCode != http.StatusOK {
			t.Fatalf("Expected status 200, got %d", approveResp.StatusCode)
		}

		// Parse approval response
		var approvalResp types.DetailResponse
		json.NewDecoder(approveResp.Body).Decode(&approvalResp)

		// Check if PO was auto-created
		approvalData, ok := approvalResp.Data.(map[string]interface{})
		if !ok {
			t.Fatal("Failed to parse approval response")
		}

		var poID string
		if automationUsed, exists := approvalData["automationUsed"]; exists && automationUsed.(bool) {
			if autoCreatedPO, exists := approvalData["autoCreatedPO"]; exists {
				poData := autoCreatedPO.(map[string]interface{})
				poID = poData["id"].(string)
				
				// Verify PO was created correctly
				if poData["status"] != "draft" {
					t.Errorf("Expected PO status 'draft', got %s", poData["status"])
				}
				
				if poData["totalAmount"].(float64) != 100000 {
					t.Errorf("Expected PO amount 100000, got %f", poData["totalAmount"])
				}
			} else {
				t.Error("Expected auto-created PO in response")
			}
		} else {
			t.Error("Expected automation to be used for PO creation")
		}

		// Step 3: Approve PO (should auto-create GRN)
		if poID != "" {
			poApproveReq := httptest.NewRequest("POST", 
				fmt.Sprintf("/api/v1/purchase-orders/%s/approve", poID), 
				bytes.NewReader(approveBody))
			poApproveReq.Header.Set("Content-Type", "application/json")

			poApproveResp, err := app.Test(poApproveReq)
			if err != nil {
				t.Fatalf("Failed to approve PO: %v", err)
			}

			if poApproveResp.StatusCode != http.StatusOK {
				t.Fatalf("Expected status 200 for PO approval, got %d", poApproveResp.StatusCode)
			}

			// Parse PO approval response
			var poApprovalResp types.DetailResponse
			json.NewDecoder(poApproveResp.Body).Decode(&poApprovalResp)

			poApprovalData, ok := poApprovalResp.Data.(map[string]interface{})
			if !ok {
				t.Fatal("Failed to parse PO approval response")
			}

			var grnID string
			if automationUsed, exists := poApprovalData["automationUsed"]; exists && automationUsed.(bool) {
				if autoCreatedGRN, exists := poApprovalData["autoCreatedGRN"]; exists {
					grnData := autoCreatedGRN.(map[string]interface{})
					grnID = grnData["id"].(string)
					
					// Verify GRN was created correctly
					if grnData["status"] != "draft" {
						t.Errorf("Expected GRN status 'draft', got %s", grnData["status"])
					}
				} else {
					t.Error("Expected auto-created GRN in response")
				}
			} else {
				t.Error("Expected automation to be used for GRN creation")
			}

			// Step 4: Approve GRN (should auto-create PV)
			if grnID != "" {
				grnApproveReq := httptest.NewRequest("POST", 
					fmt.Sprintf("/api/v1/grns/%s/approve", grnID), 
					bytes.NewReader(approveBody))
				grnApproveReq.Header.Set("Content-Type", "application/json")

				grnApproveResp, err := app.Test(grnApproveReq)
				if err != nil {
					t.Fatalf("Failed to approve GRN: %v", err)
				}

				if grnApproveResp.StatusCode != http.StatusOK {
					t.Fatalf("Expected status 200 for GRN approval, got %d", grnApproveResp.StatusCode)
				}

				// Parse GRN approval response
				var grnApprovalResp types.DetailResponse
				json.NewDecoder(grnApproveResp.Body).Decode(&grnApprovalResp)

				grnApprovalData, ok := grnApprovalResp.Data.(map[string]interface{})
				if !ok {
					t.Fatal("Failed to parse GRN approval response")
				}

				if automationUsed, exists := grnApprovalData["automationUsed"]; exists && automationUsed.(bool) {
					if autoCreatedPV, exists := grnApprovalData["autoCreatedPV"]; exists {
						pvData := autoCreatedPV.(map[string]interface{})
						
						// Verify PV was created correctly
						if pvData["status"] != "draft" {
							t.Errorf("Expected PV status 'draft', got %s", pvData["status"])
						}
						
						if pvData["amount"].(float64) != 100000 {
							t.Errorf("Expected PV amount 100000, got %f", pvData["amount"])
						}
					} else {
						t.Error("Expected auto-created PV in response")
					}
				} else {
					t.Error("Expected automation to be used for PV creation")
				}
			}
		}

		// Step 5: Verify all documents exist in database
		var requisitionCount, poCount, grnCount, pvCount int64
		
		db.Model(&models.Requisition{}).Where("organization_id = ?", orgID).Count(&requisitionCount)
		db.Model(&models.PurchaseOrder{}).Where("organization_id = ?", orgID).Count(&poCount)
		db.Model(&models.GoodsReceivedNote{}).Where("organization_id = ?", orgID).Count(&grnCount)
		db.Model(&models.PaymentVoucher{}).Where("organization_id = ?", orgID).Count(&pvCount)

		if requisitionCount != 1 {
			t.Errorf("Expected 1 requisition, found %d", requisitionCount)
		}
		if poCount != 1 {
			t.Errorf("Expected 1 PO, found %d", poCount)
		}
		if grnCount != 1 {
			t.Errorf("Expected 1 GRN, found %d", grnCount)
		}
		if pvCount != 1 {
			t.Errorf("Expected 1 PV, found %d", pvCount)
		}

		// Step 6: Verify document linking
		var po models.PurchaseOrder
		var grn models.GoodsReceivedNote
		var pv models.PaymentVoucher

		db.Where("organization_id = ?", orgID).First(&po)
		db.Where("organization_id = ?", orgID).First(&grn)
		db.Where("organization_id = ?", orgID).First(&pv)

		if po.LinkedRequisition != requisitionID {
			t.Errorf("PO should link to requisition %s, got %s", requisitionID, po.LinkedRequisition)
		}

		if grn.PONumber != po.PONumber {
			t.Errorf("GRN should link to PO %s, got %s", po.PONumber, grn.PONumber)
		}

		if pv.LinkedPO != po.PONumber {
			t.Errorf("PV should link to PO %s, got %s", po.PONumber, pv.LinkedPO)
		}
	})
}

// TestAutomationWithMissingPrerequisites tests automation failure scenarios
func TestAutomationWithMissingPrerequisites(t *testing.T) {
	db := setupIntegrationTestDB()
	app := setupFiberApp(db)

	orgID, userID, _, categoryID := createIntegrationTestData(db)

	// Update the test user ID
	db.Model(&models.User{}).Where("id = ?", userID).Update("id", "test-user-id")

	t.Run("Requisition without preferred vendor should not auto-create PO", func(t *testing.T) {
		// Create requisition without preferred vendor
		requisitionData := types.CreateRequisitionRequest{
			Title:       "Test Requisition No Vendor",
			Description: "Testing without vendor",
			Department:  "IT",
			Priority:    "medium",
			TotalAmount: 50000,
			Currency:    "USD",
			CategoryID:  &categoryID,
			// PreferredVendorID: nil, // No vendor
			Items: []types.RequisitionItem{
				{
					ItemNo:      1,
					Description: "Test Item",
					Quantity:    10,
					UnitPrice:   5000,
					Amount:      50000,
				},
			},
		}

		// Create requisition
		reqBody, _ := json.Marshal(requisitionData)
		req := httptest.NewRequest("POST", "/api/v1/requisitions", bytes.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Failed to create requisition: %v", err)
		}

		// Parse response to get requisition ID
		var createResp types.DetailResponse
		json.NewDecoder(resp.Body).Decode(&createResp)
		
		requisitionResp := createResp.Data.(map[string]interface{})
		requisitionID := requisitionResp["id"].(string)

		// Approve requisition
		approveData := types.ApproveDocumentRequest{
			Comments:  "Approved without vendor",
			Signature: "test-signature-456",
		}

		approveBody, _ := json.Marshal(approveData)
		approveReq := httptest.NewRequest("POST", 
			fmt.Sprintf("/api/v1/requisitions/%s/approve", requisitionID), 
			bytes.NewReader(approveBody))
		approveReq.Header.Set("Content-Type", "application/json")

		approveResp, err := app.Test(approveReq)
		if err != nil {
			t.Fatalf("Failed to approve requisition: %v", err)
		}

		// Parse approval response
		var approvalResp types.DetailResponse
		json.NewDecoder(approveResp.Body).Decode(&approvalResp)

		// Should not have automation used
		if approvalData, ok := approvalResp.Data.(map[string]interface{}); ok {
			if automationUsed, exists := approvalData["automationUsed"]; exists && automationUsed.(bool) {
				t.Error("Expected automation NOT to be used when no vendor is specified")
			}
		}

		// Verify no PO was created
		var poCount int64
		db.Model(&models.PurchaseOrder{}).Where("organization_id = ?", orgID).Count(&poCount)
		if poCount != 0 {
			t.Errorf("Expected 0 POs, found %d", poCount)
		}
	})
}

// TestAutomationPerformance tests the performance of automation operations
func TestAutomationPerformance(t *testing.T) {
	db := setupIntegrationTestDB()
	orgID, userID, vendorID, _ := createIntegrationTestData(db)

	// Initialize automation service
	auditService := &services.AuditService{}
	notificationService := &services.NotificationService{}
	automationService := services.NewDocumentAutomationService(
		db, auditService, notificationService,
	)

	config := services.AutomationConfig{
		AutoCreatePOFromRequisition: true,
		AutoCreateGRNFromPO:         true,
		AutoCreatePVFromGRN:         true,
	}

	t.Run("Performance test: Create 100 documents through automation", func(t *testing.T) {
		startTime := time.Now()

		for i := 0; i < 100; i++ {
			// Create requisition
			requisition := models.Requisition{
				ID:                uuid.New().String(),
				REQNumber:         fmt.Sprintf("PERF-REQ-%d", i),
				RequesterID:       userID,
				Title:             fmt.Sprintf("Performance Test Requisition %d", i),
				Description:       "Performance testing",
				Status:            "approved",
				TotalAmount:       10000,
				Currency:          "USD",
				PreferredVendorID: &vendorID,
				OrganizationID:    orgID,
				CreatedAt:         time.Now(),
				UpdatedAt:         time.Now(),
			}

			items := []models.RequisitionItem{
				{
					ItemNo:      1,
					Description: fmt.Sprintf("Perf Item %d", i),
					Quantity:    1,
					UnitPrice:   10000,
					Amount:      10000,
				},
			}
			requisition.Items = datatypes.NewJSONType(items)
			requisition.ApprovalHistory = datatypes.NewJSONType([]models.ApprovalRecord{})

			db.Create(&requisition)

			// Auto-create PO
			result, err := automationService.CreatePurchaseOrderFromRequisition(
				nil, &requisition, config,
			)

			if err != nil || !result.Success {
				t.Fatalf("Failed to create PO %d: %v", i, err)
			}
		}

		duration := time.Since(startTime)
		t.Logf("Created 100 requisitions and POs in %v", duration)

		// Verify all documents were created
		var reqCount, poCount int64
		db.Model(&models.Requisition{}).Where("organization_id = ?", orgID).Count(&reqCount)
		db.Model(&models.PurchaseOrder{}).Where("organization_id = ?", orgID).Count(&poCount)

		if reqCount != 100 {
			t.Errorf("Expected 100 requisitions, got %d", reqCount)
		}
		if poCount != 100 {
			t.Errorf("Expected 100 POs, got %d", poCount)
		}

		// Performance benchmark: should complete in reasonable time
		if duration > 30*time.Second {
			t.Errorf("Performance test took too long: %v", duration)
		}
	})
}

// BenchmarkCompleteAutomationWorkflow benchmarks the full automation workflow
func BenchmarkCompleteAutomationWorkflow(b *testing.B) {
	db := setupIntegrationTestDB()
	orgID, userID, vendorID, _ := createIntegrationTestData(db)

	auditService := &services.AuditService{}
	notificationService := &services.NotificationService{}
	automationService := services.NewDocumentAutomationService(
		db, auditService, notificationService,
	)

	config := services.AutomationConfig{
		AutoCreatePOFromRequisition: true,
		AutoCreateGRNFromPO:         true,
		AutoCreatePVFromGRN:         true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Create requisition
		requisition := models.Requisition{
			ID:                uuid.New().String(),
			RequesterID:       userID,
			Status:            "approved",
			TotalAmount:       50000,
			PreferredVendorID: &vendorID,
			OrganizationID:    orgID,
		}
		requisition.Items = datatypes.NewJSONType([]models.RequisitionItem{
			{ItemNo: 1, Description: "Bench Item", Quantity: 1, UnitPrice: 50000, Amount: 50000},
		})
		requisition.ApprovalHistory = datatypes.NewJSONType([]models.ApprovalRecord{})

		// Create PO
		poResult, _ := automationService.CreatePurchaseOrderFromRequisition(nil, &requisition, config)
		if !poResult.Success {
			b.Fatalf("Failed to create PO: %v", poResult.Error)
		}

		po := poResult.CreatedDocument.(models.PurchaseOrder)
		po.Status = "approved"

		// Create GRN
		grnResult, _ := automationService.CreateGRNFromPurchaseOrder(nil, &po, config)
		if !grnResult.Success {
			b.Fatalf("Failed to create GRN: %v", grnResult.Error)
		}

		grn := grnResult.CreatedDocument.(models.GoodsReceivedNote)
		grn.Status = "approved"

		// Create PV
		pvResult, _ := automationService.CreatePaymentVoucherFromGRN(nil, &grn, config)
		if !pvResult.Success {
			b.Fatalf("Failed to create PV: %v", pvResult.Error)
		}
	}
}