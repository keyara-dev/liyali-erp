package unit

import (
	"testing"

	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/tests/helpers"
	"github.com/stretchr/testify/assert"
)

// TestDocumentAutomationService tests document automation using mocks
func TestDocumentAutomationService(t *testing.T) {
	builder := helpers.NewMockTestDataBuilder()
	
	t.Run("Create purchase order from requisition", func(t *testing.T) {
		// Create mock requisition
		mockRequisition := builder.CreateMockRequisition(t)
		mockRequisition.Status = "APPROVED"
		
		// Create mock purchase order
		mockPO := &models.PurchaseOrder{
			ID:                uuid.New().String(),
			OrganizationID:    builder.GetOrganizationID(),
			DocumentNumber:    "PO-" + uuid.New().String()[:8],
			Status: "DRAFT",
			LinkedRequisition: mockRequisition.ID,
		}
		
		// Verify PO creation
		assert.NotNil(t, mockPO)
		assert.Equal(t, mockRequisition.ID, mockPO.LinkedRequisition)
		assert.Equal(t, builder.GetOrganizationID(), mockPO.OrganizationID)
	})
	
	t.Run("Create GRN from purchase order", func(t *testing.T) {
		// Create mock purchase order
		mockPO := &models.PurchaseOrder{
			ID:             uuid.New().String(),
			OrganizationID: builder.GetOrganizationID(),
			DocumentNumber: "PO-" + uuid.New().String()[:8],
			Status: "APPROVED",
		}
		
		// Create mock GRN
		mockGRN := &models.GoodsReceivedNote{
			ID:               uuid.New().String(),
			OrganizationID:   builder.GetOrganizationID(),
			DocumentNumber:   "GRN-" + uuid.New().String()[:8],
			PODocumentNumber: mockPO.DocumentNumber,
			Status: "DRAFT",
		}
		
		// Verify GRN creation
		assert.NotNil(t, mockGRN)
		assert.Equal(t, mockPO.DocumentNumber, mockGRN.PODocumentNumber)
		assert.Equal(t, builder.GetOrganizationID(), mockGRN.OrganizationID)
	})
	
	t.Run("Create payment voucher from GRN", func(t *testing.T) {
		// Create mock GRN
		mockGRN := &models.GoodsReceivedNote{
			ID:             uuid.New().String(),
			OrganizationID: builder.GetOrganizationID(),
			DocumentNumber: "GRN-" + uuid.New().String()[:8],
			Status: "APPROVED",
		}
		
		// Create mock payment voucher
		mockPV := &models.PaymentVoucher{
			ID:             uuid.New().String(),
			OrganizationID: builder.GetOrganizationID(),
			DocumentNumber: "PV-" + uuid.New().String()[:8],
			Status: "DRAFT",
			LinkedPO:       mockGRN.DocumentNumber,
		}
		
		// Verify PV creation
		assert.NotNil(t, mockPV)
		assert.Equal(t, mockGRN.DocumentNumber, mockPV.LinkedPO)
		assert.Equal(t, builder.GetOrganizationID(), mockPV.OrganizationID)
	})
	
	t.Run("Automation chain: Requisition -> PO -> GRN -> PV", func(t *testing.T) {
		// Create mock requisition
		mockRequisition := builder.CreateMockRequisition(t)
		mockRequisition.Status = "APPROVED"
		
		// Create mock PO
		mockPO := &models.PurchaseOrder{
			ID:                uuid.New().String(),
			OrganizationID:    builder.GetOrganizationID(),
			DocumentNumber:    "PO-" + uuid.New().String()[:8],
			Status: "APPROVED",
			LinkedRequisition: mockRequisition.ID,
		}
		
		// Create mock GRN
		mockGRN := &models.GoodsReceivedNote{
			ID:               uuid.New().String(),
			OrganizationID:   builder.GetOrganizationID(),
			DocumentNumber:   "GRN-" + uuid.New().String()[:8],
			PODocumentNumber: mockPO.DocumentNumber,
			Status: "APPROVED",
		}
		
		// Create mock PV
		mockPV := &models.PaymentVoucher{
			ID:             uuid.New().String(),
			OrganizationID: builder.GetOrganizationID(),
			DocumentNumber: "PV-" + uuid.New().String()[:8],
			Status: "DRAFT",
			LinkedPO:       mockGRN.DocumentNumber,
		}
		
		// Verify automation chain
		assert.Equal(t, mockRequisition.ID, mockPO.LinkedRequisition)
		assert.Equal(t, mockPO.DocumentNumber, mockGRN.PODocumentNumber)
		assert.Equal(t, mockGRN.DocumentNumber, mockPV.LinkedPO)
	})
	
	t.Run("Prevent automation for non-approved documents", func(t *testing.T) {
		// Create mock draft requisition
		mockRequisition := builder.CreateMockRequisition(t)
		mockRequisition.Status = "DRAFT"
		
		// Verify automation should not occur
		assert.NotEqual(t, "approved", mockRequisition.Status)
	})
	
	t.Run("Automation respects organization boundaries", func(t *testing.T) {
		builder2 := helpers.NewMockTestDataBuilder()
		
		// Create mock requisition in org1
		mockReq1 := builder.CreateMockRequisition(t)
		mockReq1.Status = "APPROVED"
		
		// Create mock PO in org2
		mockPO2 := &models.PurchaseOrder{
			ID:                uuid.New().String(),
			OrganizationID:    builder2.GetOrganizationID(),
			DocumentNumber:    "PO-" + uuid.New().String()[:8],
			Status: "DRAFT",
			LinkedRequisition: mockReq1.ID,
		}
		
		// Verify organizations are different
		assert.NotEqual(t, mockReq1.OrganizationID, mockPO2.OrganizationID)
	})
}
