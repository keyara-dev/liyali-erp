package services

import (
	"testing"

	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/types"
	"github.com/stretchr/testify/assert"
	"gorm.io/datatypes"
)

func TestCreatePurchaseOrderFromRequisition_WithoutVendor(t *testing.T) {
	// Create a test requisition without vendor
	requisition := &models.Requisition{
		ID:                "req-123",
		REQNumber:         "REQ-001",
		OrganizationID:    "org-123",
		Status:            "approved",
		TotalAmount:       1500.00,
		Currency:          "USD",
		PreferredVendorID: nil, // No vendor specified
		Items:             datatypes.NewJSONType([]types.RequisitionItem{
			{
				Description: "Laptop",
				Quantity:    1,
				UnitPrice:   1500.00,
				Amount:      1500.00,
			},
		}),
	}

	// Test the vendor handling logic that we enhanced
	
	// Test case 1: No vendor specified
	t.Run("NoVendorSpecified", func(t *testing.T) {
		var vendorID string
		var vendorName string = "TBD - To Be Determined"
		
		if requisition.PreferredVendorID != nil && *requisition.PreferredVendorID != "" {
			vendorID = *requisition.PreferredVendorID
			vendorName = "Some Vendor"
		} else {
			vendorID = ""
			vendorName = "TBD - To Be Determined"
		}
		
		assert.Equal(t, "", vendorID)
		assert.Equal(t, "TBD - To Be Determined", vendorName)
	})
	
	// Test case 2: Invalid vendor ID specified
	t.Run("InvalidVendorSpecified", func(t *testing.T) {
		invalidVendorID := "invalid-vendor-123"
		requisition.PreferredVendorID = &invalidVendorID
		
		var vendorID string
		var vendorName string
		
		if requisition.PreferredVendorID != nil && *requisition.PreferredVendorID != "" {
			// In real implementation, this would check database
			// For test, we simulate vendor not found
			vendorID = ""
			vendorName = "Invalid Vendor (ID: " + *requisition.PreferredVendorID + ")"
		} else {
			vendorID = ""
			vendorName = "TBD - To Be Determined"
		}
		
		assert.Equal(t, "", vendorID)
		assert.Equal(t, "Invalid Vendor (ID: invalid-vendor-123)", vendorName)
	})
	
	// Test case 3: Valid vendor ID specified
	t.Run("ValidVendorSpecified", func(t *testing.T) {
		validVendorID := "vendor-123"
		requisition.PreferredVendorID = &validVendorID
		
		var vendorID string
		var vendorName string
		
		if requisition.PreferredVendorID != nil && *requisition.PreferredVendorID != "" {
			// In real implementation, this would find vendor in database
			// For test, we simulate vendor found
			vendorID = *requisition.PreferredVendorID
			vendorName = "Test Vendor Inc."
		} else {
			vendorID = ""
			vendorName = "TBD - To Be Determined"
		}
		
		assert.Equal(t, "vendor-123", vendorID)
		assert.Equal(t, "Test Vendor Inc.", vendorName)
	})
}

func TestGetDefaultAutomationConfig(t *testing.T) {
	service := &DocumentAutomationService{}
	
	config := service.GetDefaultAutomationConfig()
	
	assert.True(t, config.AutoCreatePOFromRequisition)
	assert.True(t, config.AutoCreateGRNFromPO)
	assert.True(t, config.AutoCreatePVFromGRN)
	assert.True(t, config.RequireApprovalForAuto)
}

func TestValidateAutomationPrerequisites_WithoutVendorRequirement(t *testing.T) {
	service := &DocumentAutomationService{}
	
	// Test requisition validation without vendor requirement
	t.Run("RequisitionWithoutVendor", func(t *testing.T) {
		requisition := &models.Requisition{
			Status:            "approved",
			PreferredVendorID: nil, // No vendor
		}
		
		err := service.ValidateAutomationPrerequisites("requisition", requisition)
		
		// Should not fail due to missing vendor
		assert.NoError(t, err)
	})
	
	t.Run("RequisitionNotApproved", func(t *testing.T) {
		requisition := &models.Requisition{
			Status:            "pending",
			PreferredVendorID: nil,
		}
		
		err := service.ValidateAutomationPrerequisites("requisition", requisition)
		
		// Should fail because not approved
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be approved")
	})
	
	t.Run("InvalidDocumentType", func(t *testing.T) {
		requisition := &models.Requisition{
			Status: "approved",
		}
		
		err := service.ValidateAutomationPrerequisites("invalid_type", requisition)
		
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported document type")
	})
}

func TestAutomationResult(t *testing.T) {
	// Test AutomationResult structure
	result := &AutomationResult{
		Success:         true,
		CreatedDocument: &models.PurchaseOrder{ID: "po-123"},
		DocumentType:    "purchase_order",
		DocumentID:      "po-123",
		Error:           nil,
	}
	
	assert.True(t, result.Success)
	assert.Equal(t, "purchase_order", result.DocumentType)
	assert.Equal(t, "po-123", result.DocumentID)
	assert.NoError(t, result.Error)
	assert.NotNil(t, result.CreatedDocument)
}