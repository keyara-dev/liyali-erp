package handlers

import (
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/logging"
	"github.com/liyali/liyali-gateway/middleware"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/services"
	"github.com/liyali/liyali-gateway/types"
	"github.com/liyali/liyali-gateway/utils"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// ============================================================================
// PURCHASE ORDER — FROM REQUISITION
// POST /api/v1/purchase-orders/from-requisition
// ============================================================================

// CreatePurchaseOrderFromRequisition creates a PO pre-populated from an approved requisition.
func CreatePurchaseOrderFromRequisition(c *fiber.Ctx) error {
	logger := logging.FromContext(c)
	logger.Info("create_po_from_requisition_request")

	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return utils.SendUnauthorizedError(c, "Organization context required")
	}

	var req struct {
		RequisitionID             string        `json:"requisitionId"`
		RequisitionDocumentNumber string        `json:"requisitionDocumentNumber"`
		Title                     string        `json:"title"`
		Description               string        `json:"description"`
		VendorID                  string        `json:"vendorId"`
		VendorName                string        `json:"vendorName"`
		Department                string        `json:"department"`
		DepartmentID              string        `json:"departmentId"`
		RequiredByDate            *time.Time    `json:"requiredByDate"`
		Priority                  string        `json:"priority"`
		Items                     []types.POItem `json:"items"`
		TotalAmount               float64       `json:"totalAmount"`
		Currency                  string        `json:"currency"`
		BudgetCode                string        `json:"budgetCode"`
		CostCenter                string        `json:"costCenter"`
		ProjectCode               string        `json:"projectCode"`
		WorkflowID                string        `json:"workflowId"`
		// "" = inherit from org, "goods_first" or "payment_first" to override per-PO
		ProcurementFlow           string        `json:"procurementFlow"`
	}

	if err := c.BodyParser(&req); err != nil {
		return utils.SendBadRequestError(c, "Invalid request body")
	}
	if req.RequisitionID == "" {
		return utils.SendBadRequestError(c, "requisitionId is required")
	}
	if len(req.Items) == 0 {
		return utils.SendBadRequestError(c, "At least one item is required")
	}
	if req.TotalAmount <= 0 {
		return utils.SendBadRequestError(c, "totalAmount must be greater than 0")
	}
	if req.Currency == "" {
		req.Currency = "ZMW"
	}

	// One-to-one: reject if any non-cancelled PO already exists for this REQ
	var existingPO models.PurchaseOrder
	if err := config.DB.
		Where("source_requisition_id = ? AND organization_id = ? AND UPPER(status) != 'CANCELLED'",
			req.RequisitionID, tenant.OrganizationID).
		First(&existingPO).Error; err == nil {
		return utils.SendConflictError(c, fmt.Sprintf(
			"Purchase order %s already exists for this requisition (status: %s).",
			existingPO.DocumentNumber, existingPO.Status))
	}

	// Load requisition (with preferred vendor) to compare for audit trail
	var requisition models.Requisition
	config.DB.Preload("PreferredVendor").Where("id = ? AND organization_id = ?", req.RequisitionID, tenant.OrganizationID).First(&requisition)

	// Verify vendor belongs to this org if provided
	var vendorIDPtr *string
	if req.VendorID != "" {
		var vendor models.Vendor
		if err := config.DB.Where("id = ? AND organization_id = ?", req.VendorID, tenant.OrganizationID).First(&vendor).Error; err != nil {
			return utils.SendBadRequestError(c, "Vendor not found")
		}
		vendorIDPtr = &req.VendorID
	}

	documentNumber := utils.GenerateDocumentNumber("PO")
	orderID := uuid.New().String()

	order := models.PurchaseOrder{
		ID:                orderID,
		OrganizationID:    tenant.OrganizationID,
		DocumentNumber:    documentNumber,
		VendorID:          vendorIDPtr,
		Status: "DRAFT",
		TotalAmount:       req.TotalAmount,
		Currency:          req.Currency,
		ApprovalStage:     0,
		LinkedRequisition: req.RequisitionDocumentNumber,
		Title:             req.Title,
		Description:       req.Description,
		Department:        req.Department,
		DepartmentID:      req.DepartmentID,
		Priority:          req.Priority,
		BudgetCode:        req.BudgetCode,
		CostCenter:        req.CostCenter,
		ProjectCode:       req.ProjectCode,
		ProcurementFlow:   req.ProcurementFlow,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if req.RequiredByDate != nil {
		order.RequiredByDate = req.RequiredByDate
	}
	if req.RequisitionID != "" {
		order.SourceRequisitionId = &req.RequisitionID
	}

	order.Items = datatypes.NewJSONType(req.Items)
	order.ApprovalHistory = datatypes.NewJSONType([]types.ApprovalRecord{})

	// Build initial action history — log vendor change + chain link
	var initialHistory []types.ActionHistoryEntry

	// Always record the chain origin
	if requisition.DocumentNumber != "" {
		initialHistory = append(initialHistory, types.ActionHistoryEntry{
			ID:          uuid.New().String(),
			Action:      "CREATED_FROM_REQUISITION",
			PerformedBy: tenant.UserID,
			Timestamp:   time.Now(),
			Metadata: map[string]interface{}{
				"linkedDocNumber": requisition.DocumentNumber,
				"linkedDocType":   "requisition",
			},
		})
	}

	reqPreferredVendorID := ""
	if requisition.PreferredVendorID != nil {
		reqPreferredVendorID = *requisition.PreferredVendorID
	}
	if req.VendorID != reqPreferredVendorID && reqPreferredVendorID != "" {
		oldVendorName := ""
		if requisition.PreferredVendor != nil {
			oldVendorName = requisition.PreferredVendor.Name
		}
		initialHistory = append(initialHistory, types.ActionHistoryEntry{
			ID:          uuid.New().String(),
			Action:      "VENDOR_CHANGED",
			PerformedBy: tenant.UserID,
			Timestamp:   time.Now(),
			ChangedFields: map[string]interface{}{
				"vendor": map[string]interface{}{
					"from": oldVendorName,
					"to":   req.VendorName,
				},
			},
		})
	}
	order.ActionHistory = datatypes.NewJSONType(initialHistory)

	if err := config.DB.Create(&order).Error; err != nil {
		logging.LogError(c, err, "create_po_from_requisition_failed", nil)
		return utils.SendInternalError(c, "Failed to create purchase order", err)
	}

	// Record PO_CREATED on the source requisition for full chain traceability
	if req.RequisitionID != "" {
		reqHistory := requisition.ActionHistory.Data()
		reqHistory = append(reqHistory, types.ActionHistoryEntry{
			ID:          uuid.New().String(),
			Action:      "PO_CREATED",
			PerformedBy: tenant.UserID,
			Timestamp:   time.Now(),
			Metadata: map[string]interface{}{
				"linkedDocNumber": order.DocumentNumber,
				"linkedDocType":   "purchase_order",
			},
		})
		requisition.ActionHistory = datatypes.NewJSONType(reqHistory)
		config.DB.Save(&requisition)
	}

	config.DB.Preload("Vendor").First(&order)
	go utils.SyncDocument(config.DB, "PURCHASE_ORDER", order.ID)

	logger.Info("po_from_requisition_created")
	return utils.SendCreatedSuccess(c, modelToPurchaseOrderResponse(order), "Purchase order created from requisition successfully")
}

// ============================================================================
// PAYMENT VOUCHER — FROM PURCHASE ORDER
// POST /api/v1/payment-vouchers/from-po
// ============================================================================

// CreatePaymentVoucherFromPO creates a PV pre-populated from an approved PO.
func CreatePaymentVoucherFromPO(c *fiber.Ctx) error {
	logger := logging.FromContext(c)
	logger.Info("create_pv_from_po_request")

	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return utils.SendUnauthorizedError(c, "Organization context required")
	}

	var req struct {
		PurchaseOrderID             string              `json:"purchaseOrderId"`
		PurchaseOrderDocumentNumber string              `json:"purchaseOrderDocumentNumber"`
		Title                       string              `json:"title"`
		Description                 string              `json:"description"`
		VendorID                    string              `json:"vendorId"`
		VendorName                  string              `json:"vendorName"`
		Department                  string              `json:"department"`
		DepartmentID                string              `json:"departmentId"`
		Items                       []types.PaymentItem `json:"items"`
		TotalAmount                 float64             `json:"totalAmount"`
		Currency                    string              `json:"currency"`
		BudgetCode                  string              `json:"budgetCode"`
		CostCenter                  string              `json:"costCenter"`
		ProjectCode                 string              `json:"projectCode"`
		SourceRequisitionID         string              `json:"sourceRequisitionId"`
		WorkflowID                  string              `json:"workflowId"`
		// Goods-first flow: required GRN document number (e.g. "GRN-20240101-001")
		LinkedGRNDocumentNumber     string              `json:"linkedGRNDocumentNumber"`
	}

	if err := c.BodyParser(&req); err != nil {
		return utils.SendBadRequestError(c, "Invalid request body")
	}
	if req.PurchaseOrderID == "" {
		return utils.SendBadRequestError(c, "purchaseOrderId is required")
	}
	if req.TotalAmount <= 0 {
		return utils.SendBadRequestError(c, "totalAmount must be greater than 0")
	}
	if req.Currency == "" {
		req.Currency = "ZMW"
	}

	// Load the PO with Vendor preload (needed for audit trail comparison)
	var po models.PurchaseOrder
	if err := config.DB.Preload("Vendor").Where("id = ? AND organization_id = ?", req.PurchaseOrderID, tenant.OrganizationID).First(&po).Error; err != nil {
		return utils.SendBadRequestError(c, "Purchase order not found")
	}

	// Resolve effective procurement flow: PO override → org default → "goods_first"
	effectiveFlow := po.ProcurementFlow
	if effectiveFlow == "" {
		orgSvc := services.NewOrganizationService(config.DB)
		orgSettings, _ := orgSvc.GetOrganizationSettings(tenant.OrganizationID)
		if orgSettings != nil && orgSettings.ProcurementFlow != "" {
			effectiveFlow = orgSettings.ProcurementFlow
		} else {
			effectiveFlow = "goods_first"
		}
	}

	// Goods-first enforcement: require an approved GRN before creating PV
	var linkedGRN *models.GoodsReceivedNote
	if effectiveFlow == "goods_first" {
		if req.LinkedGRNDocumentNumber == "" {
			return utils.SendBadRequestError(c, "A linked GRN document number is required for goods-first procurement flow")
		}
		var grn models.GoodsReceivedNote
		if err := config.DB.Where("document_number = ? AND organization_id = ?", req.LinkedGRNDocumentNumber, tenant.OrganizationID).First(&grn).Error; err != nil {
			return utils.SendBadRequestError(c, "Linked GRN not found")
		}
		if strings.ToUpper(grn.Status) != "APPROVED" {
			return utils.SendBadRequestError(c, "Linked GRN must be approved before creating a payment voucher (goods-first flow)")
		}
		if grn.PODocumentNumber != po.DocumentNumber {
			return utils.SendBadRequestError(c, "Linked GRN does not belong to the selected purchase order")
		}
		linkedGRN = &grn
	}

	// Verify vendor belongs to this org if provided
	var vendorIDPtr *string
	if req.VendorID != "" {
		var vendor models.Vendor
		if err := config.DB.Where("id = ? AND organization_id = ?", req.VendorID, tenant.OrganizationID).First(&vendor).Error; err != nil {
			return utils.SendBadRequestError(c, "Vendor not found")
		}
		vendorIDPtr = &req.VendorID
	}

	documentNumber := utils.GenerateDocumentNumber("PV")
	invoiceRef := "INV-" + po.DocumentNumber

	linkedGRNDocNum := ""
	if linkedGRN != nil {
		linkedGRNDocNum = linkedGRN.DocumentNumber
	}

	voucher := models.PaymentVoucher{
		ID:             uuid.New().String(),
		OrganizationID: tenant.OrganizationID,
		DocumentNumber: documentNumber,
		VendorID:       vendorIDPtr,
		InvoiceNumber:  invoiceRef,
		Status: "DRAFT",
		Amount:         req.TotalAmount,
		Currency:       req.Currency,
		PaymentMethod:  "bank_transfer",
		Description:    req.Description,
		ApprovalStage:  0,
		LinkedPO:       req.PurchaseOrderDocumentNumber,
		LinkedGRN:      linkedGRNDocNum,
		Title:          req.Title,
		Department:     req.Department,
		DepartmentID:   req.DepartmentID,
		BudgetCode:     req.BudgetCode,
		CostCenter:     req.CostCenter,
		ProjectCode:    req.ProjectCode,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if len(req.Items) > 0 {
		voucher.Items = datatypes.NewJSONType(req.Items)
	}
	voucher.ApprovalHistory = datatypes.NewJSONType([]types.ApprovalRecord{})

	// Build initial action history — chain origin + vendor change if applicable
	var pvInitialHistory []types.ActionHistoryEntry

	// Record which document this PV was created from
	if linkedGRN != nil {
		pvInitialHistory = append(pvInitialHistory, types.ActionHistoryEntry{
			ID:          uuid.New().String(),
			Action:      "CREATED_FROM_GRN",
			PerformedBy: tenant.UserID,
			Timestamp:   time.Now(),
			Metadata: map[string]interface{}{
				"linkedDocNumber": linkedGRN.DocumentNumber,
				"linkedDocType":   "grn",
				"flow":            "goods_first",
			},
		})
	} else {
		pvInitialHistory = append(pvInitialHistory, types.ActionHistoryEntry{
			ID:          uuid.New().String(),
			Action:      "CREATED_FROM_PO",
			PerformedBy: tenant.UserID,
			Timestamp:   time.Now(),
			Metadata: map[string]interface{}{
				"linkedDocNumber": po.DocumentNumber,
				"linkedDocType":   "purchase_order",
				"flow":            "payment_first",
			},
		})
	}

	poVendorID := ""
	if po.VendorID != nil {
		poVendorID = *po.VendorID
	}
	if req.VendorID != poVendorID && poVendorID != "" {
		oldVendorName := ""
		if po.Vendor != nil {
			oldVendorName = po.Vendor.Name
		}
		pvInitialHistory = append(pvInitialHistory, types.ActionHistoryEntry{
			ID:          uuid.New().String(),
			Action:      "VENDOR_CHANGED",
			PerformedBy: tenant.UserID,
			Timestamp:   time.Now(),
			ChangedFields: map[string]interface{}{
				"vendor": map[string]interface{}{
					"from": oldVendorName,
					"to":   req.VendorName,
				},
			},
		})
	}
	voucher.ActionHistory = datatypes.NewJSONType(pvInitialHistory)

	if err := config.DB.Create(&voucher).Error; err != nil {
		logging.LogError(c, err, "create_pv_from_po_failed", nil)
		return utils.SendInternalError(c, "Failed to create payment voucher", err)
	}

	// Record PV_CREATED on the parent document (GRN for goods_first, PO for payment_first)
	pvCreatedEntry := types.ActionHistoryEntry{
		ID:          uuid.New().String(),
		Action:      "PV_CREATED",
		PerformedBy: tenant.UserID,
		Timestamp:   time.Now(),
		Metadata: map[string]interface{}{
			"linkedDocNumber": voucher.DocumentNumber,
			"linkedDocType":   "payment_voucher",
			"flow":            effectiveFlow,
		},
	}
	if linkedGRN != nil {
		grnHistory := linkedGRN.ActionHistory.Data()
		grnHistory = append(grnHistory, pvCreatedEntry)
		linkedGRN.ActionHistory = datatypes.NewJSONType(grnHistory)
		config.DB.Save(linkedGRN)
	} else {
		poHistory := po.ActionHistory.Data()
		poHistory = append(poHistory, pvCreatedEntry)
		po.ActionHistory = datatypes.NewJSONType(poHistory)
		config.DB.Save(&po)
	}

	config.DB.Preload("Vendor").First(&voucher)
	go utils.SyncDocument(config.DB, "PAYMENT_VOUCHER", voucher.ID)

	logger.Info("pv_from_po_created")
	return utils.SendCreatedSuccess(c, modelToPaymentVoucherResponse(voucher), "Payment voucher created from purchase order successfully")
}

// ============================================================================
// PAYMENT VOUCHER — MARK PAID
// POST /api/v1/payment-vouchers/:id/mark-paid
// ============================================================================

// MarkPaymentVoucherPaid marks an approved PV as paid and records payment details.
func MarkPaymentVoucherPaid(c *fiber.Ctx) error {
	logger := logging.FromContext(c)
	logger.Info("mark_pv_paid_request")

	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return utils.SendUnauthorizedError(c, "Organization context required")
	}

	id := c.Params("id")
	if id == "" {
		return utils.SendBadRequestError(c, "Payment voucher ID is required")
	}

	var req struct {
		PaidAmount      float64    `json:"paidAmount"`
		PaidDate        *time.Time `json:"paidDate"`
		ReferenceNumber string     `json:"referenceNumber"`
		Comments        string     `json:"comments"`
	}

	if err := c.BodyParser(&req); err != nil {
		return utils.SendBadRequestError(c, "Invalid request body")
	}
	if req.PaidAmount <= 0 {
		return utils.SendBadRequestError(c, "paidAmount must be greater than 0")
	}

	var voucher models.PaymentVoucher
	if err := config.DB.Where("id = ? AND organization_id = ?", id, tenant.OrganizationID).First(&voucher).Error; err != nil {
		return utils.SendNotFoundError(c, "Payment voucher not found")
	}

	if strings.ToUpper(voucher.Status) != "APPROVED" {
		return utils.SendBadRequestError(c, "Only approved payment vouchers can be marked as paid")
	}

	now := time.Now()
	paidDate := &now
	if req.PaidDate != nil {
		paidDate = req.PaidDate
	}

	voucher.Status = "PAID"
	voucher.PaidAmount = &req.PaidAmount
	voucher.PaidDate = paidDate
	voucher.UpdatedAt = now

	userID := c.Locals("userID").(string)
	var user models.User
	config.DB.Where("id = ?", userID).First(&user)

	actionHistory := voucher.ActionHistory.Data()
	actionHistory = append(actionHistory, types.ActionHistoryEntry{
		ID:              uuid.New().String(),
		Action:          "MARK_PAID",
		PerformedBy:     userID,
		PerformedByName: user.Name,
		PerformedByRole: user.Role,
		Timestamp:       now,
		Comments:        req.Comments,
		ActionType:      "MARK_PAID",
		PreviousStatus:  "APPROVED",
		NewStatus:       "PAID",
	})
	voucher.ActionHistory = datatypes.NewJSONType(actionHistory)

	if err := config.DB.Save(&voucher).Error; err != nil {
		logging.LogError(c, err, "mark_pv_paid_failed", nil)
		return utils.SendInternalError(c, "Failed to mark payment voucher as paid", err)
	}

	go utils.SyncDocument(config.DB, "PAYMENT_VOUCHER", voucher.ID)

	logger.Info("pv_marked_paid")
	return utils.SendSimpleSuccess(c, modelToPaymentVoucherResponse(voucher), "Payment voucher marked as paid successfully")
}

// ============================================================================
// STATS ENDPOINTS
// ============================================================================

// GetRequisitionStats returns count summaries for requisitions in the org.
// GET /api/v1/requisitions/stats
func GetRequisitionStats(c *fiber.Ctx) error {
	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return utils.SendUnauthorizedError(c, "Organization context required")
	}

	db := config.DB
	scope := utils.GetDocumentScope(db, tenant.UserID, tenant.UserRole, tenant.OrganizationID)
	base := db.Model(&models.Requisition{}).Where("organization_id = ?", tenant.OrganizationID)
	base = scope.ApplyToQuery(base, "requester_id", "requisition", "")

	stats := fiber.Map{}
	for _, status := range []string{"draft", "pending", "approved", "rejected", "completed", "cancelled"} {
		var count int64
		base.Session(&gorm.Session{}).Where("UPPER(status) = ?", strings.ToUpper(status)).Count(&count)
		stats[status] = count
	}

	var total int64
	base.Session(&gorm.Session{}).Count(&total)
	stats["total"] = total

	return utils.SendSimpleSuccess(c, stats, "Requisition statistics retrieved successfully")
}

// GetPurchaseOrderStats returns count summaries for POs in the org.
// GET /api/v1/purchase-orders/stats
func GetPurchaseOrderStats(c *fiber.Ctx) error {
	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return utils.SendUnauthorizedError(c, "Organization context required")
	}

	db := config.DB
	scope := utils.GetDocumentScope(db, tenant.UserID, tenant.UserRole, tenant.OrganizationID)
	base := db.Model(&models.PurchaseOrder{}).Where("organization_id = ?", tenant.OrganizationID)
	base = scope.ApplyToQuery(base, "created_by", "purchase_order", "")

	stats := fiber.Map{}
	for _, status := range []string{"draft", "pending", "approved", "rejected", "fulfilled", "completed", "cancelled"} {
		var count int64
		base.Session(&gorm.Session{}).Where("UPPER(status) = ?", strings.ToUpper(status)).Count(&count)
		stats[status] = count
	}

	var total int64
	base.Session(&gorm.Session{}).Count(&total)
	stats["total"] = total

	return utils.SendSimpleSuccess(c, stats, "Purchase order statistics retrieved successfully")
}

// GetPaymentVoucherStats returns count summaries for PVs in the org.
// GET /api/v1/payment-vouchers/stats
func GetPaymentVoucherStats(c *fiber.Ctx) error {
	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return utils.SendUnauthorizedError(c, "Organization context required")
	}

	db := config.DB
	scope := utils.GetDocumentScope(db, tenant.UserID, tenant.UserRole, tenant.OrganizationID)
	base := db.Model(&models.PaymentVoucher{}).Where("organization_id = ?", tenant.OrganizationID)
	base = scope.ApplyToQuery(base, "vendor_id", "payment_voucher", "")

	stats := fiber.Map{}
	for _, status := range []string{"draft", "pending", "approved", "rejected", "paid", "completed", "cancelled"} {
		var count int64
		base.Session(&gorm.Session{}).Where("UPPER(status) = ?", strings.ToUpper(status)).Count(&count)
		stats[status] = count
	}

	var total int64
	base.Session(&gorm.Session{}).Count(&total)
	stats["total"] = total

	return utils.SendSimpleSuccess(c, stats, "Payment voucher statistics retrieved successfully")
}

// ============================================================================
// DEPARTMENT HEADS LIST
// GET /api/v1/users/department-heads/list
// ============================================================================

// GetDepartmentHeadsList returns organisation members with roles that can act as approvers/HODs.
func GetDepartmentHeadsList(c *fiber.Ctx) error {
	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return utils.SendUnauthorizedError(c, "Organization context required")
	}

	departmentID := c.Query("department_id")
	roleID := c.Query("role_id")
	isActiveStr := c.Query("is_active")
	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("page_size", 50)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}

	// Roles eligible to be department heads / approvers
	eligibleRoles := []string{"admin", "approver", "finance"}

	query := config.DB.Table("users u").
		Select("u.id, u.name, u.email, u.role, u.position, om.department_id").
		Joins("JOIN organization_members om ON om.user_id = u.id").
		Where("om.organization_id = ? AND om.active = true AND u.role IN ?", tenant.OrganizationID, eligibleRoles)

	if departmentID != "" {
		query = query.Where("om.department_id = ?", departmentID)
	}
	if roleID != "" {
		query = query.Where("u.role = ?", roleID)
	}
	if isActiveStr == "true" {
		query = query.Where("u.active = true")
	} else if isActiveStr == "false" {
		query = query.Where("u.active = false")
	}

	var total int64
	query.Count(&total)

	type HODUser struct {
		ID           string `json:"id"`
		Name         string `json:"name"`
		Email        string `json:"email"`
		Role         string `json:"role"`
		Position     string `json:"position"`
		DepartmentID string `json:"departmentId"`
	}

	var users []HODUser
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Scan(&users).Error; err != nil {
		return utils.SendInternalError(c, "Failed to retrieve department heads", err)
	}

	totalPages := (total + int64(pageSize) - 1) / int64(pageSize)
	return utils.SendSuccess(c, fiber.StatusOK, users, "Department heads retrieved successfully", &types.PaginationMeta{
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    int64(page) < totalPages,
		HasPrev:    page > 1,
	})
}

// ============================================================================
// SIGNATURE VALIDATION
// POST /api/v1/approvals/validate-signature
// ============================================================================

// ValidateSignature checks that a submitted digital signature is non-empty and well-formed.
func ValidateSignature(c *fiber.Ctx) error {
	var req struct {
		Signature string `json:"signature"`
		UserID    string `json:"userId"`
	}

	if err := c.BodyParser(&req); err != nil {
		return utils.SendBadRequestError(c, "Invalid request body")
	}
	if req.Signature == "" {
		return utils.SendBadRequestError(c, "Signature is required")
	}

	// Accept both plain base64 and data-URI format (data:image/png;base64,...)
	raw := req.Signature
	if idx := strings.Index(raw, "base64,"); idx != -1 {
		raw = raw[idx+7:]
	}

	_, decodeErr := base64.StdEncoding.DecodeString(raw)
	if decodeErr != nil {
		// Try URL-safe variant
		_, decodeErr = base64.URLEncoding.DecodeString(raw)
	}

	if decodeErr != nil {
		return utils.SendSimpleSuccess(c, fiber.Map{
			"valid":   false,
			"message": "Signature is not valid base64 encoded data",
		}, "Signature validation completed")
	}

	return utils.SendSimpleSuccess(c, fiber.Map{
		"valid":   true,
		"message": "Signature is valid",
	}, "Signature validation completed")
}

// ============================================================================
// APPROVER WORKLOAD
// GET /api/v1/approvals/approver-workload/:approverId
// ============================================================================

// GetApproverWorkload returns pending task count and basic stats for a given approver.
func GetApproverWorkload(c *fiber.Ctx) error {
	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return utils.SendUnauthorizedError(c, "Organization context required")
	}

	approverID := c.Params("approverId")
	if approverID == "" {
		return utils.SendBadRequestError(c, "Approver ID is required")
	}

	db := config.DB

	// Count pending tasks assigned to this approver in this org
	var pendingCount int64
	db.Table("workflow_tasks wt").
		Joins("JOIN workflow_assignments wa ON wa.id = wt.workflow_assignment_id").
		Where("wt.assigned_to = ? AND UPPER(wt.status) = 'PENDING' AND wa.organization_id = ?", approverID, tenant.OrganizationID).
		Count(&pendingCount)

	// Count tasks completed this month
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	var completedThisMonth int64
	db.Table("workflow_tasks wt").
		Joins("JOIN workflow_assignments wa ON wa.id = wt.workflow_assignment_id").
		Where("wt.assigned_to = ? AND UPPER(wt.status) IN ? AND wt.updated_at >= ? AND wa.organization_id = ?",
			approverID, []string{"APPROVED", "REJECTED"}, startOfMonth, tenant.OrganizationID).
		Count(&completedThisMonth)

	// Count overdue tasks (past due_date and still pending)
	var overdueTasks int64
	db.Table("workflow_tasks wt").
		Joins("JOIN workflow_assignments wa ON wa.id = wt.workflow_assignment_id").
		Where("wt.assigned_to = ? AND UPPER(wt.status) = 'PENDING' AND wt.due_date < ? AND wa.organization_id = ?",
			approverID, now, tenant.OrganizationID).
		Count(&overdueTasks)

	return utils.SendSimpleSuccess(c, fiber.Map{
		"pendingCount":        pendingCount,
		"averageResponseTime": 0, // would require time-series aggregation
		"completedThisMonth":  completedThisMonth,
		"overdueTasks":        overdueTasks,
	}, "Approver workload retrieved successfully")
}

// ============================================================================
// GRN CONFIRM
// POST /api/v1/grns/:id/confirm
// ============================================================================

// ConfirmGRN marks an approved GRN as confirmed/completed.
func ConfirmGRN(c *fiber.Ctx) error {
	logger := logging.FromContext(c)
	logger.Info("confirm_grn_request")

	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return utils.SendUnauthorizedError(c, "Organization context required")
	}

	id := c.Params("id")
	if id == "" {
		return utils.SendBadRequestError(c, "GRN ID is required")
	}

	var req struct {
		Comments string `json:"comments"`
	}
	c.BodyParser(&req) // optional body

	var grn models.GoodsReceivedNote
	if err := config.DB.Where("id = ? AND organization_id = ?", id, tenant.OrganizationID).First(&grn).Error; err != nil {
		return utils.SendNotFoundError(c, "GRN not found")
	}

	if strings.ToUpper(grn.Status) != "APPROVED" {
		return utils.SendBadRequestError(c, "Only approved GRNs can be confirmed")
	}

	userID := c.Locals("userID").(string)
	var user models.User
	config.DB.Where("id = ?", userID).First(&user)

	now := time.Now()
	grn.Status = "COMPLETED"
	grn.UpdatedAt = now

	actionHistory := grn.ActionHistory.Data()
	actionHistory = append(actionHistory, types.ActionHistoryEntry{
		ID:              uuid.New().String(),
		Action:          "CONFIRM",
		PerformedBy:     userID,
		PerformedByName: user.Name,
		PerformedByRole: user.Role,
		Timestamp:       now,
		Comments:        req.Comments,
		ActionType:      "CONFIRM",
		PreviousStatus:  "APPROVED",
		NewStatus:       "COMPLETED",
	})
	grn.ActionHistory = datatypes.NewJSONType(actionHistory)

	if err := config.DB.Save(&grn).Error; err != nil {
		logging.LogError(c, err, "confirm_grn_failed", nil)
		return utils.SendInternalError(c, "Failed to confirm GRN", err)
	}

	go utils.SyncDocument(config.DB, "GRN", grn.ID)

	logger.Info("grn_confirmed")
	return utils.SendSimpleSuccess(c, modelToGRNResponse(grn), "GRN confirmed successfully")
}
