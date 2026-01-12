package unit

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/services"
	"github.com/liyali/liyali-gateway/types"
	"gorm.io/datatypes"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Auto-migrate the schema
	db.AutoMigrate(
		&models.Organization{},
		&models.User{},
		&models.Vendor{},
		&models.Requisition{},
		&models.PurchaseOrder{},
		&models.GoodsReceivedNote{},
		&models.PaymentVoucher{},
	)

	return db
}

// createTestData creates test data for automation tests
func createTestData(db *gorm.DB) (string, string, string) {
	// Create organization
	org := models.Organization{
		ID:   uuid.New().String(),
		Name: "Test Organization",
	}
	db.Create(&org)

	// Create user
	user := models.User{
		ID:    uuid.New().String(),
		Email: "test@example.com",
		Name:  "Test User",
	}
	db.Create(&user)

	// Create vendor
	vendor := models.Vendor{
		ID:    uuid.New().String(),
		Name:  "Test Vendor",
		Email: "vendor@example.com",
	}
	db.Create(&vendor)

	return org.ID, user.ID, vendor.ID
}

// TestCreatePurchaseOrderFromRequisition tests automatic PO creation
func TestCreatePurchaseOrderFromRequisition(t *testing.T) {
	db := setupTestDB()
	auditService := &services.AuditService{}
	notificationService := &services.NotificationService{}
	
	automationService := services.NewDocumentAutomationService(
		db, auditService, notificationService,
	)

	_, userID, vendorID := createTestData(db)
	ctx := context.Background()

	t.Run("Successfully creates PO from approved requisition", func(t *testing.T) {
		// Create approved requisition with preferred vendor
		requisition := models.Requisition{
			ID:                uuid.New().String(),
			REQNumber:         "REQ-TEST-001",
			RequesterId:       userID,
			Title:             "Test Requisition",
			Description:       "Test Description",
			Status:            "approved",
			TotalAmount:       50000,
			Currency:          "USD",
			PreferredVendorID: &vendorID,
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		}

		// Add test items
		items := []types.RequisitionItem{
			{
				Description: "Test Item 1",
				Quantity:    10,
				UnitPrice:   1000,
				Amount:      10000,
			},
			{
				Description: "Test Item 2",
				Quantity:    20,
				UnitPrice:   2000,
				Amount:      40000,
			},
		}
		requisition.Items = datatypes.NewJSONType(items)
		requisition.ApprovalHistory = datatypes.NewJSONType([]types.ApprovalRecord{})

		db.Create(&requisition)

		// Test automation
		config := services.AutomationConfig{
			AutoCreatePOFromRequisition: true,
		}

		result, err := automationService.CreatePurchaseOrderFromRequisition(
			ctx, &requisition, config,
		)

		// Assertions
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if !result.Success {
			t.Fatalf("Expected success, got failure: %v", result.Error)
		}

		if result.DocumentType != "purchase_order" {
			t.Errorf("Expected document type 'purchase_order', got %s", result.DocumentType)
		}

		// Verify PO was created in database
		var createdPO models.PurchaseOrder
		if err := db.Where("linked_requisition = ?", requisition.ID).First(&createdPO).Error; err != nil {
			t.Fatalf("PO not found in database: %v", err)
		}

		// Verify PO data
		if createdPO.VendorID != vendorID {
			t.Errorf("Expected vendor ID %s, got %s", vendorID, createdPO.VendorID)
		}

		if createdPO.TotalAmount != requisition.TotalAmount {
			t.Errorf("Expected amount %f, got %f", requisition.TotalAmount, createdPO.TotalAmount)
		}

		if createdPO.Status != "draft" {
			t.Errorf("Expected status 'draft', got %s", createdPO.Status)
		}

		// Verify items were copied correctly
		var poItems []types.POItem
		poItems = createdPO.Items.Data()

		if len(poItems) != len(items) {
			t.Errorf("Expected %d items, got %d", len(items), len(poItems))
		}

		for i, poItem := range poItems {
			if poItem.Description != items[i].Description {
				t.Errorf("Item %d: expected description %s, got %s", 
					i, items[i].Description, poItem.Description)
			}
			if poItem.Quantity != items[i].Quantity {
				t.Errorf("Item %d: expected quantity %d, got %d", 
					i, items[i].Quantity, poItem.Quantity)
			}
		}
	})

	t.Run("Fails when automation is disabled", func(t *testing.T) {
		requisition := models.Requisition{
			ID:                uuid.New().String(),
			Status:            "approved",
			PreferredVendorID: &vendorID,
		}

		config := services.AutomationConfig{
			AutoCreatePOFromRequisition: false, // Disabled
		}

		result, err := automationService.CreatePurchaseOrderFromRequisition(
			ctx, &requisition, config,
		)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if result.Success {
			t.Error("Expected failure when automation is disabled")
		}

		if result.Error == nil {
			t.Error("Expected error message when automation is disabled")
		}
	})

	t.Run("Fails when requisition is not approved", func(t *testing.T) {
		requisition := models.Requisition{
			ID:                uuid.New().String(),
			Status:            "draft", // Not approved
			PreferredVendorID: &vendorID,
		}

		config := services.AutomationConfig{
			AutoCreatePOFromRequisition: true,
		}

		result, err := automationService.CreatePurchaseOrderFromRequisition(
			ctx, &requisition, config,
		)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if result.Success {
			t.Error("Expected failure when requisition is not approved")
		}
	})

	t.Run("Fails when no preferred vendor", func(t *testing.T) {
		requisition := models.Requisition{
			ID:                uuid.New().String(),
			Status:            "approved",
			PreferredVendorID: nil, // No vendor
		}

		config := services.AutomationConfig{
			AutoCreatePOFromRequisition: true,
		}

		result, err := automationService.CreatePurchaseOrderFromRequisition(
			ctx, &requisition, config,
		)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if result.Success {
			t.Error("Expected failure when no preferred vendor")
		}
	})

	t.Run("Fails when vendor does not exist", func(t *testing.T) {
		nonExistentVendorID := uuid.New().String()
		requisition := models.Requisition{
			ID:                uuid.New().String(),
			Status:            "approved",
			PreferredVendorID: &nonExistentVendorID,
		}

		config := services.AutomationConfig{
			AutoCreatePOFromRequisition: true,
		}

		result, err := automationService.CreatePurchaseOrderFromRequisition(
			ctx, &requisition, config,
		)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if result.Success {
			t.Error("Expected failure when vendor does not exist")
		}
	})
}

// TestCreateGRNFromPurchaseOrder tests automatic GRN creation
func TestCreateGRNFromPurchaseOrder(t *testing.T) {
	db := setupTestDB()
	auditService := &services.AuditService{}
	notificationService := &services.NotificationService{}
	
	automationService := services.NewDocumentAutomationService(
		db, auditService, notificationService,
	)

	_, _, vendorID := createTestData(db)
	ctx := context.Background()

	t.Run("Successfully creates GRN from approved PO", func(t *testing.T) {
		// Create approved purchase order
		purchaseOrder := models.PurchaseOrder{
			ID:             uuid.New().String(),
			PONumber:       "PO-TEST-001",
			VendorID:       vendorID,
			Status:         "approved",
			TotalAmount:    75000,
			Currency:       "USD",
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		// Add test items
		items := []types.POItem{
			{
				Description: "PO Item 1",
				Quantity:    15,
				UnitPrice:   2000,
				Amount:      30000,
			},
			{
				Description: "PO Item 2",
				Quantity:    25,
				UnitPrice:   1800,
				Amount:      45000,
			},
		}
		purchaseOrder.Items = datatypes.NewJSONType(items)
		purchaseOrder.ApprovalHistory = datatypes.NewJSONType([]types.ApprovalRecord{})

		db.Create(&purchaseOrder)

		// Test automation
		config := services.AutomationConfig{
			AutoCreateGRNFromPO: true,
		}

		result, err := automationService.CreateGRNFromPurchaseOrder(
			ctx, &purchaseOrder, config,
		)

		// Assertions
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if !result.Success {
			t.Fatalf("Expected success, got failure: %v", result.Error)
		}

		if result.DocumentType != "grn" {
			t.Errorf("Expected document type 'grn', got %s", result.DocumentType)
		}

		// Verify GRN was created in database
		var createdGRN models.GoodsReceivedNote
		if err := db.Where("po_number = ?", purchaseOrder.PONumber).First(&createdGRN).Error; err != nil {
			t.Fatalf("GRN not found in database: %v", err)
		}

		// Verify GRN data
		if createdGRN.PONumber != purchaseOrder.PONumber {
			t.Errorf("Expected PO number %s, got %s", purchaseOrder.PONumber, createdGRN.PONumber)
		}

		if createdGRN.Status != "draft" {
			t.Errorf("Expected status 'draft', got %s", createdGRN.Status)
		}

		// Verify items were copied correctly
		var grnItems []types.GRNItem
		grnItems = createdGRN.Items.Data()

		if len(grnItems) != len(items) {
			t.Errorf("Expected %d items, got %d", len(items), len(grnItems))
		}

		for i, grnItem := range grnItems {
			if grnItem.Description != items[i].Description {
				t.Errorf("Item %d: expected description %s, got %s", 
					i, items[i].Description, grnItem.Description)
			}
			if grnItem.QuantityOrdered != items[i].Quantity {
				t.Errorf("Item %d: expected quantity %d, got %d", 
					i, items[i].Quantity, grnItem.QuantityOrdered)
			}
			if grnItem.QuantityReceived != 0 {
				t.Errorf("Item %d: expected received quantity 0, got %d", 
					i, grnItem.QuantityReceived)
			}
		}
	})

	t.Run("Fails when automation is disabled", func(t *testing.T) {
		purchaseOrder := models.PurchaseOrder{
			ID:             uuid.New().String(),
			Status:         "approved",
		}

		config := services.AutomationConfig{
			AutoCreateGRNFromPO: false, // Disabled
		}

		result, err := automationService.CreateGRNFromPurchaseOrder(
			ctx, &purchaseOrder, config,
		)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if result.Success {
			t.Error("Expected failure when automation is disabled")
		}
	})

	t.Run("Fails when PO is not approved", func(t *testing.T) {
		purchaseOrder := models.PurchaseOrder{
			ID:             uuid.New().String(),
			Status:         "draft", // Not approved
		}

		config := services.AutomationConfig{
			AutoCreateGRNFromPO: true,
		}

		result, err := automationService.CreateGRNFromPurchaseOrder(
			ctx, &purchaseOrder, config,
		)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if result.Success {
			t.Error("Expected failure when PO is not approved")
		}
	})
}

// TestCreatePaymentVoucherFromGRN tests automatic PV creation
func TestCreatePaymentVoucherFromGRN(t *testing.T) {
	db := setupTestDB()
	auditService := &services.AuditService{}
	notificationService := &services.NotificationService{}
	
	automationService := services.NewDocumentAutomationService(
		db, auditService, notificationService,
	)

	_, _, vendorID := createTestData(db)
	ctx := context.Background()

	t.Run("Successfully creates PV from approved GRN", func(t *testing.T) {
		// Create linked PO first
		purchaseOrder := models.PurchaseOrder{
			ID:             uuid.New().String(),
			PONumber:       "PO-TEST-002",
			VendorID:       vendorID,
			Status:         "approved",
			TotalAmount:    100000,
			Currency:       "USD",
		}
		db.Create(&purchaseOrder)

		// Create approved GRN
		grn := models.GoodsReceivedNote{
			ID:             uuid.New().String(),
			GRNNumber:      "GRN-TEST-001",
			PONumber:       purchaseOrder.PONumber,
			Status:         "approved",
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}
		grn.ApprovalHistory = datatypes.NewJSONType([]types.ApprovalRecord{})
		db.Create(&grn)

		// Test automation
		config := services.AutomationConfig{
			AutoCreatePVFromGRN: true,
		}

		result, err := automationService.CreatePaymentVoucherFromGRN(
			ctx, &grn, config,
		)

		// Assertions
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if !result.Success {
			t.Fatalf("Expected success, got failure: %v", result.Error)
		}

		if result.DocumentType != "payment_voucher" {
			t.Errorf("Expected document type 'payment_voucher', got %s", result.DocumentType)
		}

		// Verify PV was created in database
		var createdPV models.PaymentVoucher
		if err := db.Where("linked_po = ?", purchaseOrder.PONumber).First(&createdPV).Error; err != nil {
			t.Fatalf("PV not found in database: %v", err)
		}

		// Verify PV data
		if createdPV.VendorID != vendorID {
			t.Errorf("Expected vendor ID %s, got %s", vendorID, createdPV.VendorID)
		}

		if createdPV.Amount != purchaseOrder.TotalAmount {
			t.Errorf("Expected amount %f, got %f", purchaseOrder.TotalAmount, createdPV.Amount)
		}

		if createdPV.Status != "draft" {
			t.Errorf("Expected status 'draft', got %s", createdPV.Status)
		}

		if createdPV.LinkedPO != purchaseOrder.PONumber {
			t.Errorf("Expected linked PO %s, got %s", purchaseOrder.PONumber, createdPV.LinkedPO)
		}
	})

	t.Run("Fails when automation is disabled", func(t *testing.T) {
		grn := models.GoodsReceivedNote{
			ID:             uuid.New().String(),
			Status:         "approved",
		}

		config := services.AutomationConfig{
			AutoCreatePVFromGRN: false, // Disabled
		}

		result, err := automationService.CreatePaymentVoucherFromGRN(
			ctx, &grn, config,
		)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if result.Success {
			t.Error("Expected failure when automation is disabled")
		}
	})

	t.Run("Fails when GRN is not approved", func(t *testing.T) {
		grn := models.GoodsReceivedNote{
			ID:             uuid.New().String(),
			Status:         "draft", // Not approved
		}

		config := services.AutomationConfig{
			AutoCreatePVFromGRN: true,
		}

		result, err := automationService.CreatePaymentVoucherFromGRN(
			ctx, &grn, config,
		)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if result.Success {
			t.Error("Expected failure when GRN is not approved")
		}
	})

	t.Run("Fails when linked PO not found", func(t *testing.T) {
		grn := models.GoodsReceivedNote{
			ID:             uuid.New().String(),
			PONumber:       "NONEXISTENT-PO",
			Status:         "approved",
		}

		config := services.AutomationConfig{
			AutoCreatePVFromGRN: true,
		}

		result, err := automationService.CreatePaymentVoucherFromGRN(
			ctx, &grn, config,
		)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if result.Success {
			t.Error("Expected failure when linked PO not found")
		}
	})
}

// TestValidateAutomationPrerequisites tests prerequisite validation
func TestValidateAutomationPrerequisites(t *testing.T) {
	db := setupTestDB()
	auditService := &services.AuditService{}
	notificationService := &services.NotificationService{}
	
	automationService := services.NewDocumentAutomationService(
		db, auditService, notificationService,
	)

	_, _, vendorID := createTestData(db)

	t.Run("Valid requisition passes validation", func(t *testing.T) {
		requisition := &models.Requisition{
			ID:                uuid.New().String(),
			Status:            "approved",
			PreferredVendorID: &vendorID,
		}

		err := automationService.ValidateAutomationPrerequisites("requisition", requisition)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("Invalid requisition status fails validation", func(t *testing.T) {
		requisition := &models.Requisition{
			ID:                uuid.New().String(),
			Status:            "draft", // Not approved
			PreferredVendorID: &vendorID,
		}

		err := automationService.ValidateAutomationPrerequisites("requisition", requisition)
		if err == nil {
			t.Error("Expected error for non-approved requisition")
		}
	})

	t.Run("Requisition without vendor fails validation", func(t *testing.T) {
		requisition := &models.Requisition{
			ID:                uuid.New().String(),
			Status:            "approved",
			PreferredVendorID: nil, // No vendor
		}

		err := automationService.ValidateAutomationPrerequisites("requisition", requisition)
		if err == nil {
			t.Error("Expected error for requisition without vendor")
		}
	})

	t.Run("Valid PO passes validation", func(t *testing.T) {
		po := &models.PurchaseOrder{
			ID:             uuid.New().String(),
			Status:         "approved",
		}

		err := automationService.ValidateAutomationPrerequisites("purchase_order", po)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("Invalid PO status fails validation", func(t *testing.T) {
		po := &models.PurchaseOrder{
			ID:             uuid.New().String(),
			Status:         "draft", // Not approved
		}

		err := automationService.ValidateAutomationPrerequisites("purchase_order", po)
		if err == nil {
			t.Error("Expected error for non-approved PO")
		}
	})

	t.Run("Unsupported document type fails validation", func(t *testing.T) {
		err := automationService.ValidateAutomationPrerequisites("invalid_type", nil)
		if err == nil {
			t.Error("Expected error for unsupported document type")
		}
	})
}

// TestGetDefaultAutomationConfig tests default configuration
func TestGetDefaultAutomationConfig(t *testing.T) {
	db := setupTestDB()
	auditService := &services.AuditService{}
	notificationService := &services.NotificationService{}
	
	automationService := services.NewDocumentAutomationService(
		db, auditService, notificationService,
	)

	config := automationService.GetDefaultAutomationConfig()

	if !config.AutoCreatePOFromRequisition {
		t.Error("Expected AutoCreatePOFromRequisition to be true by default")
	}

	if !config.AutoCreateGRNFromPO {
		t.Error("Expected AutoCreateGRNFromPO to be true by default")
	}

	if !config.AutoCreatePVFromGRN {
		t.Error("Expected AutoCreatePVFromGRN to be true by default")
	}

	if !config.RequireApprovalForAuto {
		t.Error("Expected RequireApprovalForAuto to be true by default")
	}
}

// BenchmarkPOCreationFromRequisition benchmarks PO creation performance
func BenchmarkPOCreationFromRequisition(b *testing.B) {
	db := setupTestDB()
	auditService := &services.AuditService{}
	notificationService := &services.NotificationService{}
	
	automationService := services.NewDocumentAutomationService(
		db, auditService, notificationService,
	)

	_, userID, vendorID := createTestData(db)
	ctx := context.Background()

	// Create test requisition
	requisition := models.Requisition{
		ID:                uuid.New().String(),
		RequesterId:       userID,
		Status:            "approved",
		TotalAmount:       50000,
		PreferredVendorID: &vendorID,
	}
	requisition.Items = datatypes.NewJSONType([]types.RequisitionItem{
	})
	requisition.ApprovalHistory = datatypes.NewJSONType([]types.ApprovalRecord{})

	config := services.AutomationConfig{AutoCreatePOFromRequisition: true}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Create unique requisition for each iteration
		testReq := requisition
		testReq.ID = uuid.New().String()
		testReq.REQNumber = fmt.Sprintf("REQ-BENCH-%d", i)
		db.Create(&testReq)

		_, err := automationService.CreatePurchaseOrderFromRequisition(ctx, &testReq, config)
		if err != nil {
			b.Fatalf("Benchmark failed: %v", err)
		}
	}
}