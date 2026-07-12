package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/config"
	db "github.com/liyali/liyali-gateway/database/sqlc"
	"github.com/liyali/liyali-gateway/middleware"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/services"
	"github.com/liyali/liyali-gateway/types"
	"github.com/liyali/liyali-gateway/utils"
	"gorm.io/datatypes"
)

// GetGRNs retrieves all goods received notes with pagination and filtering
func GetGRNs(c *fiber.Ctx) error {
	// Get organization context from tenant middleware
	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Organization context required",
			"error":   err.Error(),
		})
	}

	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	status := c.Query("status")
	poDocumentNumber := c.Query("poDocumentNumber")

	scope := utils.GetDocumentScope(config.DB, tenant.UserID, tenant.UserRole, tenant.OrganizationID)

	ctx := c.Context()
	offset := int32((page - 1) * limit)
	orgRoleIDs := scope.OrgRoleIDs
	if orgRoleIDs == nil {
		orgRoleIDs = []string{}
	}

	var total int64
	var ids []string

	// In production sqlc.Queries is wired against pgx. The SQLite-backed
	// test harness leaves it nil — fall back to a gorm equivalent that
	// covers the same scope semantics.
	if config.Queries == nil {
		total, ids, err = listGRNIDsGorm(tenant.OrganizationID, status, poDocumentNumber, scope, limit, int(offset))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Failed to fetch GRNs",
				"error":   err.Error(),
			})
		}
	} else if scope.CanViewAll || scope.IsProcurement {
		total, err = config.Queries.CountGRNsAll(ctx, db.CountGRNsAllParams{
			OrganizationID:    tenant.OrganizationID,
			Column2:           status,
			Column3:           poDocumentNumber,
			HideDirectPayment: scope.HideDirectPayment,
		})
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Failed to count GRNs",
				"error":   err.Error(),
			})
		}
		ids, err = config.Queries.ListGRNIDsAll(ctx, db.ListGRNIDsAllParams{
			OrganizationID:    tenant.OrganizationID,
			Column2:           status,
			Column3:           poDocumentNumber,
			HideDirectPayment: scope.HideDirectPayment,
			Limit:             int32(limit),
			Offset:            offset,
		})
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Failed to fetch GRNs",
				"error":   err.Error(),
			})
		}
	} else {
		total, err = config.Queries.CountGRNsLimited(ctx, db.CountGRNsLimitedParams{
			OrganizationID: tenant.OrganizationID,
			Column2:        status,
			Column3:        poDocumentNumber,
			CreatedBy:      &scope.UserID,
			Lower:          scope.UserRole,
			Column6:        orgRoleIDs,
		})
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Failed to count GRNs",
				"error":   err.Error(),
			})
		}
		ids, err = config.Queries.ListGRNIDsLimited(ctx, db.ListGRNIDsLimitedParams{
			OrganizationID: tenant.OrganizationID,
			Column2:        status,
			Column3:        poDocumentNumber,
			CreatedBy:      &scope.UserID,
			Lower:          scope.UserRole,
			Column6:        orgRoleIDs,
			Limit:          int32(limit),
			Offset:         offset,
		})
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Failed to fetch GRNs",
				"error":   err.Error(),
			})
		}
	}

	var grns []models.GoodsReceivedNote
	if len(ids) > 0 {
		if err := config.DB.
			Where("id IN ?", ids).
			Order("created_at DESC").
			Find(&grns).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Failed to fetch GRNs",
				"error":   err.Error(),
			})
		}
	}

	responses := make([]types.GRNResponse, 0, len(grns))
	for _, grn := range grns {
		responses = append(responses, modelToGRNResponse(grn))
	}
	enrichGRNUsers(responses, tenant.OrganizationID)

	return utils.SendPaginatedSuccess(c, responses, "GRNs retrieved successfully", page, limit, total)
}

// enrichGRNUsers batch-resolves the user-reference IDs on each GRN response into
// {id,name,email,role} objects (single query) so clients render a name + role
// instead of a bare UUID. Safe to call with a one-element slice for detail
// responses.
func enrichGRNUsers(responses []types.GRNResponse, orgID string) {
	if len(responses) == 0 {
		return
	}
	ids := make([]string, 0, len(responses)*5)
	for _, r := range responses {
		ids = append(ids, r.ReceivedBy, r.CreatedBy, r.OwnerID, r.ApprovedBy, r.CertifiedByID)
	}
	users := utils.ResolveUserRefs(config.DB, orgID, ids)
	for i := range responses {
		r := &responses[i]
		if u, ok := users[r.ReceivedBy]; ok {
			r.Receiver = &u
		}
		creatorID := r.CreatedBy
		if creatorID == "" {
			creatorID = r.OwnerID
		}
		if u, ok := users[creatorID]; ok {
			r.Creator = &u
		}
		if u, ok := users[r.ApprovedBy]; ok {
			r.Approver = &u
		}
		if u, ok := users[r.CertifiedByID]; ok {
			r.Certifier = &u
		}
	}
}

// listGRNIDsGorm is the sqlc-free implementation of the GRN listing query.
// Used as the fallback when config.Queries is nil (i.e. SQLite-backed unit
// tests). Mirrors the filters from CountGRNsAll/ListGRNIDsAll +
// CountGRNsLimited/ListGRNIDsLimited so behaviour stays in sync.
func listGRNIDsGorm(orgID, status, poDocumentNumber string, scope utils.DocumentScope, limit, offset int) (int64, []string, error) {
	q := config.DB.Table("goods_received_notes").Where("organization_id = ?", orgID)

	if status != "" {
		q = q.Where("UPPER(status) = UPPER(?)", status)
	}
	if poDocumentNumber != "" {
		q = q.Where("po_document_number = ?", poDocumentNumber)
	}
	if scope.HideDirectPayment {
		// Direct-payment GRNs are flagged in metadata; SQLite stores the JSON
		// column as TEXT so we treat a NULL/empty value as "not direct".
		q = q.Where("COALESCE(metadata ->> 'directPayment', '') <> 'true'")
	}

	if !(scope.CanViewAll || scope.IsProcurement) {
		// Limited scope: creator OR receiver.
		q = q.Where("(created_by = ? OR received_by = ?)", scope.UserID, scope.UserID)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return 0, nil, err
	}

	var ids []string
	if err := q.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Pluck("id", &ids).Error; err != nil {
		return 0, nil, err
	}
	return total, ids, nil
}

// CreateGRN creates a new goods received note
func CreateGRN(c *fiber.Ctx) error {
	// Get organization context from tenant middleware
	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Organization context required",
			"error":   err.Error(),
		})
	}

	var req types.CreateGRNRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	if req.PODocumentNumber == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "PO number is required",
		})
	}
	// Validate PO document number format (should start with "PO-" and be at least 10 characters)
	if len(req.PODocumentNumber) < 10 || req.PODocumentNumber[:3] != "PO-" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid PO document number format",
		})
	}
	if len(req.Items) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "At least one item is required",
		})
	}
	// Validate items have positive quantities
	for _, item := range req.Items {
		if item.QuantityOrdered <= 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "All items must have positive quantities",
			})
		}
	}
	if req.ReceivedBy == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "ReceivedBy is required",
		})
	}

	// Verify PO exists and belongs to organization
	var po models.PurchaseOrder
	if err := config.DB.Where("document_number = ? AND organization_id = ?", req.PODocumentNumber, tenant.OrganizationID).First(&po).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Purchase order not found",
		})
	}

	// Resolve effective procurement flow: PO override → org default → "goods_first"
	orgDefaultFlow := ""
	if strings.TrimSpace(po.ProcurementFlow) == "" {
		orgSvc := services.NewOrganizationService(config.DB)
		if orgSettings, _ := orgSvc.GetOrganizationSettings(tenant.OrganizationID); orgSettings != nil {
			orgDefaultFlow = orgSettings.ProcurementFlow
		}
	}
	effectiveFlow := utils.ResolveProcurementFlow(po.ProcurementFlow, orgDefaultFlow)

	// The PO must be APPROVED before goods can be received against it — in BOTH
	// flows. Previously payment_first was exempt, which let a GRN be created
	// against a non-approved PO from which it could never be submitted (SubmitGRN
	// requires an APPROVED PO) nor close the PO via the cascades.
	if strings.ToUpper(po.Status) != "APPROVED" {
		return utils.SendBadRequestError(c, fmt.Sprintf(
			"Cannot create GRN: linked PO %s is in %s status and must be APPROVED first.",
			po.DocumentNumber, po.Status))
	}

	// One-to-one: reject if any non-cancelled GRN already exists for this PO/PV
	if effectiveFlow == "payment_first" && req.LinkedPV != "" {
		var existingGRN models.GoodsReceivedNote
		if err := config.DB.
			Where("linked_pv = ? AND organization_id = ? AND UPPER(status) NOT IN ('CANCELLED','REJECTED')",
				req.LinkedPV, tenant.OrganizationID).
			First(&existingGRN).Error; err == nil {
			return utils.SendConflictError(c, fmt.Sprintf(
				"Goods received note %s already exists for payment voucher %s (status: %s).",
				existingGRN.DocumentNumber, req.LinkedPV, existingGRN.Status))
		}
	} else {
		var existingGRN models.GoodsReceivedNote
		if err := config.DB.
			Where("po_document_number = ? AND organization_id = ? AND UPPER(status) NOT IN ('CANCELLED','REJECTED')",
				req.PODocumentNumber, tenant.OrganizationID).
			First(&existingGRN).Error; err == nil {
			return utils.SendConflictError(c, fmt.Sprintf(
				"Goods received note %s already exists for purchase order %s (status: %s).",
				existingGRN.DocumentNumber, req.PODocumentNumber, existingGRN.Status))
		}
	}

	// Item-level validation: each GRN item must match a PO line by description and
	// must not exceed the ordered quantity on that line.
	// Also snapshot itemCode from the matching PO line so the printed GRN
	// matches the PDF sample's "Item Code" column.
	poItemCodeByDesc := make(map[string]string)
	{
		poItems := po.Items.Data()
		poByDesc := make(map[string]int, len(poItems))
		for _, it := range poItems {
			key := strings.TrimSpace(strings.ToLower(it.Description))
			poByDesc[key] += it.Quantity
			if _, exists := poItemCodeByDesc[key]; !exists && it.ItemCode != "" {
				poItemCodeByDesc[key] = it.ItemCode
			}
		}

		for i, ln := range req.Items {
			key := strings.TrimSpace(strings.ToLower(ln.Description))
			ordered, ok := poByDesc[key]
			if !ok {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"success": false,
					"message": fmt.Sprintf("GRN item %q does not match any line on PO %s", ln.Description, po.DocumentNumber),
				})
			}
			if req.Items[i].ItemCode == "" {
				req.Items[i].ItemCode = poItemCodeByDesc[key]
			}
			if ln.QuantityReceived <= 0 {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"success": false,
					"message": fmt.Sprintf("GRN item %q must have quantityReceived > 0", ln.Description),
				})
			}
			if ln.QuantityReceived > ordered {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"success": false,
					"message": fmt.Sprintf("GRN item %q: quantityReceived %d exceeds ordered %d on PO %s",
						ln.Description, ln.QuantityReceived, ordered, po.DocumentNumber),
				})
			}
		}

		// Cross-GRN aggregate guard: total received across all non-cancelled GRNs
		// for this PO must not exceed the PO ordered quantity for any item.
		var existingGRNs []models.GoodsReceivedNote
		config.DB.Where("po_document_number = ? AND organization_id = ? AND UPPER(status) != ?",
			req.PODocumentNumber, tenant.OrganizationID, "CANCELLED").
			Find(&existingGRNs)
		receivedByDesc := make(map[string]int)
		for _, g := range existingGRNs {
			for _, it := range g.Items.Data() {
				receivedByDesc[strings.TrimSpace(strings.ToLower(it.Description))] += it.QuantityReceived
			}
		}
		for _, ln := range req.Items {
			key := strings.TrimSpace(strings.ToLower(ln.Description))
			total := receivedByDesc[key] + ln.QuantityReceived
			if total > poByDesc[key] {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"success": false,
					"message": fmt.Sprintf("GRN item %q: total received across GRNs would be %d, exceeds PO %s ordered %d",
						ln.Description, total, po.DocumentNumber, poByDesc[key]),
				})
			}
		}
	}

	// Payment-first enforcement: require an approved PV before goods can be received
	var linkedPVDoc *models.PaymentVoucher
	if effectiveFlow == "payment_first" {
		if req.LinkedPV == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "A linked payment voucher document number is required for payment-first procurement flow",
			})
		}
		var pv models.PaymentVoucher
		if err := config.DB.Where("document_number = ? AND organization_id = ?", req.LinkedPV, tenant.OrganizationID).First(&pv).Error; err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Linked payment voucher not found",
			})
		}
		if strings.ToUpper(pv.Status) != "APPROVED" && strings.ToUpper(pv.Status) != "PAID" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Linked payment voucher must be approved or paid before receiving goods (payment-first flow)",
			})
		}
		if pv.LinkedPO != po.DocumentNumber {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Linked payment voucher does not belong to the selected purchase order",
			})
		}
		linkedPVDoc = &pv
	}

	// Generate GRN number
	documentNumber := utils.GenerateDocumentNumber("GRN")

	linkedPVDocNum := ""
	if linkedPVDoc != nil {
		linkedPVDocNum = linkedPVDoc.DocumentNumber
	}

	// Build initial action history — chain origin
	var grnInitialHistory []types.ActionHistoryEntry
	if linkedPVDoc != nil {
		grnInitialHistory = append(grnInitialHistory, types.ActionHistoryEntry{
			ID:          uuid.New().String(),
			Action:      "CREATED_FROM_PV",
			PerformedBy: tenant.UserID,
			Timestamp:   time.Now(),
			Metadata: map[string]interface{}{
				"linkedDocNumber": linkedPVDoc.DocumentNumber,
				"linkedDocType":   "payment_voucher",
				"flow":            "payment_first",
			},
		})
	} else {
		grnInitialHistory = append(grnInitialHistory, types.ActionHistoryEntry{
			ID:          uuid.New().String(),
			Action:      "CREATED_FROM_PO",
			PerformedBy: tenant.UserID,
			Timestamp:   time.Now(),
			Metadata: map[string]interface{}{
				"linkedDocNumber": po.DocumentNumber,
				"linkedDocType":   "purchase_order",
				"flow":            "goods_first",
			},
		})
	}

	// Snapshot vendor name + address from the PO at creation time so the printed
	// GRN remains stable if the vendor record is later edited / deleted.
	vendorNameSnapshot := po.VendorName
	vendorAddressSnapshot := ""
	if po.VendorID != nil && *po.VendorID != "" {
		var vendor models.Vendor
		if err := config.DB.Where("id = ?", *po.VendorID).First(&vendor).Error; err == nil {
			if vendorNameSnapshot == "" {
				vendorNameSnapshot = vendor.Name
			}
			// Prefer the postal/physical address; fall back to city/country.
			vendorAddressSnapshot = vendor.PhysicalAddress
			if vendorAddressSnapshot == "" {
				parts := []string{}
				if vendor.City != "" {
					parts = append(parts, vendor.City)
				}
				if vendor.Country != "" {
					parts = append(parts, vendor.Country)
				}
				vendorAddressSnapshot = strings.Join(parts, ", ")
			}
		}
	}

	grn := models.GoodsReceivedNote{
		ID:                uuid.New().String(),
		OrganizationID:    tenant.OrganizationID,
		DocumentNumber:    documentNumber,
		PODocumentNumber:  req.PODocumentNumber,
		Status:            models.StatusDraft,
		ReceivedDate:      time.Now(),
		ReceivedBy:        req.ReceivedBy,
		ApprovalStage:     0,
		LinkedPV:          linkedPVDocNum,
		WarehouseLocation: req.WarehouseLocation,
		Notes:             req.Notes,
		ConsignmentNote:   req.ConsignmentNote,
		VendorName:        vendorNameSnapshot,
		VendorAddress:     vendorAddressSnapshot,
		SignoffStatus:     "PENDING_RECEIVER",
		CreatedBy:         tenant.UserID,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	grn.Items = datatypes.NewJSONType(req.Items)

	emptyQuality := []types.QualityIssue{}
	grn.QualityIssues = datatypes.NewJSONType(emptyQuality)

	emptyHistory := []types.ApprovalRecord{}
	grn.ApprovalHistory = datatypes.NewJSONType(emptyHistory)
	var grnCreateUser models.User
	config.DB.Where("id = ?", tenant.UserID).First(&grnCreateUser)
	grnCreateNow := time.Now()
	grnInitialHistory = append(grnInitialHistory, types.ActionHistoryEntry{
		ID:              uuid.New().String(),
		Action:          "CREATE",
		ActionType:      "CREATE",
		PerformedBy:     tenant.UserID,
		PerformedByName: grnCreateUser.Name,
		PerformedByRole: grnCreateUser.Role,
		Timestamp:       grnCreateNow,
		PerformedAt:     grnCreateNow,
		Comments:        "GRN created",
		NewStatus:       models.StatusDraft,
	})
	grn.ActionHistory = datatypes.NewJSONType(grnInitialHistory)

	if err := config.DB.Create(&grn).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to create GRN",
			"error":   err.Error(),
		})
	}

	// Record GRN_CREATED on the parent document for chain traceability
	grnCreatedEntry := types.ActionHistoryEntry{
		ID:          uuid.New().String(),
		Action:      "GRN_CREATED",
		PerformedBy: tenant.UserID,
		Timestamp:   time.Now(),
		Metadata: map[string]interface{}{
			"linkedDocNumber": grn.DocumentNumber,
			"linkedDocType":   "grn",
			"flow":            effectiveFlow,
		},
	}
	if linkedPVDoc != nil {
		pvHistory := linkedPVDoc.ActionHistory.Data()
		pvHistory = append(pvHistory, grnCreatedEntry)
		linkedPVDoc.ActionHistory = datatypes.NewJSONType(pvHistory)
		config.DB.Save(linkedPVDoc)
	} else {
		poHistory := po.ActionHistory.Data()
		poHistory = append(poHistory, grnCreatedEntry)
		po.ActionHistory = datatypes.NewJSONType(poHistory)
		config.DB.Save(&po)
	}

	go utils.SyncDocumentAs(config.DB, "GRN", grn.ID, tenant.UserID)
	go services.LogDocumentEvent(config.DB, services.DocumentEvent{
		OrganizationID: tenant.OrganizationID,
		DocumentID:     grn.ID,
		DocumentType:   "grn",
		UserID:         tenant.UserID,
		ActorName:      grnCreateUser.Name,
		ActorRole:      tenant.UserRole,
		Action:         "created",
		Details:        map[string]interface{}{"documentNumber": grn.DocumentNumber},
	})

	return c.Status(fiber.StatusCreated).JSON(types.DetailResponse{
		Success: true,
		Data:    modelToGRNResponse(grn),
	})
}

// GetGRN retrieves a single GRN by ID
func GetGRN(c *fiber.Ctx) error {
	// Set cache control headers to ensure fresh data for PDF generation
	c.Set("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Set("Pragma", "no-cache")
	c.Set("Expires", "0")

	// Get organization context from tenant middleware
	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Organization context required",
			"error":   err.Error(),
		})
	}

	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "GRN ID is required",
		})
	}

	// Org + role/ownership scope. GRNs have a second owner column (received_by)
	// so both creator and receiver can access the detail view.
	scope := utils.GetDocumentScope(config.DB, tenant.UserID, tenant.UserRole, tenant.OrganizationID)
	query := config.DB.Where("id = ? AND organization_id = ?", id, tenant.OrganizationID)
	query = scope.ApplyToQuery(query, "created_by", "grn", "received_by")

	var grn models.GoodsReceivedNote
	if err := query.First(&grn).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "GRN not found",
		})
	}

	response := modelToGRNResponse(grn)
	if liveHistory := utils.GetDocumentApprovalHistory(config.DB, grn.ID, "grn"); len(liveHistory) > 0 {
		response.ApprovalHistory = liveHistory
	}
	single := []types.GRNResponse{response}
	enrichGRNUsers(single, tenant.OrganizationID)
	response = single[0]
	return c.JSON(types.DetailResponse{
		Success: true,
		Data:    response,
	})
}

// UpdateGRN updates an existing GRN
func UpdateGRN(c *fiber.Ctx) error {
	// Get organization context from tenant middleware
	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Organization context required",
			"error":   err.Error(),
		})
	}

	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "GRN ID is required",
		})
	}

	var req types.UpdateGRNRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	// Scope to org + owner/involvement (created_by or received_by) for edit access.
	scope := utils.GetDocumentScope(config.DB, tenant.UserID, tenant.UserRole, tenant.OrganizationID)
	loadQuery := config.DB.Where("id = ? AND organization_id = ?", id, tenant.OrganizationID)
	loadQuery = scope.ApplyToQuery(loadQuery, "created_by", "grn", "received_by")
	var grn models.GoodsReceivedNote
	if err := loadQuery.First(&grn).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "GRN not found",
		})
	}

	// Metadata-only updates (e.g. supporting-document attachments) are allowed
	// on any status — mirrors the PO carve-out in purchase_order.go
	// UpdatePurchaseOrder. isMetadataOnly requires every other field on the
	// request to be absent/zero; a single non-metadata field falls through to
	// the status guard below.
	isMetadataOnly := len(req.Metadata) > 0 &&
		len(req.Items) == 0 &&
		req.ReceivedBy == "" &&
		len(req.QualityIssues) == 0 &&
		req.WarehouseLocation == nil &&
		req.Notes == nil &&
		req.ConsignmentNote == nil

	if strings.ToUpper(grn.Status) != "DRAFT" && strings.ToUpper(grn.Status) != "PENDING" && !isMetadataOnly {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": fmt.Sprintf("Cannot update GRN in %s status", grn.Status),
		})
	}

	// Once the receiver has signed the GRN the captured signature is bound
	// to the line items as they stood at that moment; mutating items after
	// the fact would invalidate that signature. Item edits are therefore only
	// permitted while signoff_status = PENDING_RECEIVER.
	if len(req.Items) > 0 && grn.SignoffStatus != "PENDING_RECEIVER" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "Line items cannot be edited after the receiver has signed the GRN",
		})
	}

	if len(req.Items) > 0 {
		grn.Items = datatypes.NewJSONType(req.Items)
	}
	if req.ReceivedBy != "" {
		grn.ReceivedBy = req.ReceivedBy
	}
	if len(req.QualityIssues) > 0 {
		grn.QualityIssues = datatypes.NewJSONType(req.QualityIssues)
	}
	if req.WarehouseLocation != nil {
		grn.WarehouseLocation = *req.WarehouseLocation
	}
	if req.Notes != nil {
		grn.Notes = *req.Notes
	}
	if req.ConsignmentNote != nil {
		grn.ConsignmentNote = *req.ConsignmentNote
	}
	if len(req.Metadata) > 0 {
		// Deep-merge incoming metadata with existing — never wipe keys other
		// parts of the system manage independently. Mirrors the PO carve-out
		// in purchase_order.go UpdatePurchaseOrder.
		existingMeta := map[string]interface{}{}
		if len(grn.Metadata) > 0 {
			_ = json.Unmarshal(grn.Metadata, &existingMeta)
		}
		for k, v := range req.Metadata {
			existingMeta[k] = v
		}
		if metaBytes, err := json.Marshal(existingMeta); err == nil {
			grn.Metadata = datatypes.JSON(metaBytes)
		}
	}

	var grnUpdateUser models.User
	config.DB.Where("id = ?", tenant.UserID).First(&grnUpdateUser)
	grnUpdateNow := time.Now()
	grnUpdateHistory := grn.ActionHistory.Data()
	grnUpdateHistory = append(grnUpdateHistory, types.ActionHistoryEntry{
		ID:              uuid.New().String(),
		Action:          "UPDATE",
		ActionType:      "UPDATE",
		PerformedBy:     tenant.UserID,
		PerformedByName: grnUpdateUser.Name,
		PerformedByRole: grnUpdateUser.Role,
		Timestamp:       grnUpdateNow,
		PerformedAt:     grnUpdateNow,
		Comments:        "GRN updated",
		NewStatus:       grn.Status,
	})
	grn.ActionHistory = datatypes.NewJSONType(grnUpdateHistory)
	grn.UpdatedAt = grnUpdateNow

	if err := config.DB.Save(&grn).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to update GRN",
			"error":   err.Error(),
		})
	}

	go utils.SyncDocumentAs(config.DB, "GRN", grn.ID, tenant.UserID)
	go services.LogDocumentEvent(config.DB, services.DocumentEvent{
		OrganizationID: tenant.OrganizationID,
		DocumentID:     grn.ID,
		DocumentType:   "grn",
		UserID:         tenant.UserID,
		ActorName:      grnUpdateUser.Name,
		ActorRole:      tenant.UserRole,
		Action:         "updated",
		Details:        map[string]interface{}{"documentNumber": grn.DocumentNumber},
	})

	return c.JSON(types.DetailResponse{
		Success: true,
		Data:    modelToGRNResponse(grn),
	})
}

// DeleteGRN deletes a GRN
func DeleteGRN(c *fiber.Ctx) error {
	// Get organization context from tenant middleware
	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Organization context required",
			"error":   err.Error(),
		})
	}

	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "GRN ID is required",
		})
	}

	// Scope to org + owner/involvement (created_by or received_by) for delete access.
	scope := utils.GetDocumentScope(config.DB, tenant.UserID, tenant.UserRole, tenant.OrganizationID)
	loadQuery := config.DB.Where("id = ? AND organization_id = ?", id, tenant.OrganizationID)
	loadQuery = scope.ApplyToQuery(loadQuery, "created_by", "grn", "received_by")
	var grn models.GoodsReceivedNote
	if err := loadQuery.First(&grn).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "GRN not found",
		})
	}

	if strings.ToUpper(grn.Status) != "DRAFT" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "Only draft GRNs can be deleted",
		})
	}

	if err := config.DB.Delete(&grn).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to delete GRN",
			"error":   err.Error(),
		})
	}

	return c.JSON(types.MessageResponse{
		Success: true,
		Message: "GRN deleted successfully",
	})
}

// Helper function to convert model to response
func modelToGRNResponse(grn models.GoodsReceivedNote) types.GRNResponse {
	var items []types.GRNItem
	items = grn.Items.Data()

	var qualityIssues []types.QualityIssue
	qualityIssues = grn.QualityIssues.Data()

	var approvalHistory []types.ApprovalRecord
	approvalHistory = grn.ApprovalHistory.Data()

	var actionHistory []types.ActionHistoryEntry
	actionHistory = grn.ActionHistory.Data()
	
	// Unmarshal metadata
	var metadata map[string]interface{}
	if len(grn.Metadata) > 0 {
		_ = json.Unmarshal(grn.Metadata, &metadata)
	}
	
	// Unmarshal autoCreatedPV
	var autoCreatedPV interface{}
	if len(grn.AutoCreatedPV) > 0 {
		_ = json.Unmarshal(grn.AutoCreatedPV, &autoCreatedPV)
	}

	return types.GRNResponse{
		ID:                grn.ID,
		OrganizationID:    grn.OrganizationID,
		DocumentNumber:    grn.DocumentNumber,
		PODocumentNumber:  grn.PODocumentNumber,
		Status:            grn.Status,
		ReceivedDate:      grn.ReceivedDate,
		ReceivedBy:        grn.ReceivedBy,
		Items:             items,
		QualityIssues:     qualityIssues,
		ApprovalStage:     grn.ApprovalStage,
		ApprovalHistory:   approvalHistory,
		ActionHistory:     actionHistory,
		LinkedPV:          grn.LinkedPV,
		BudgetCode:        grn.BudgetCode,
		CostCenter:        grn.CostCenter,
		ProjectCode:       grn.ProjectCode,
		CreatedBy:         grn.CreatedBy,
		OwnerID:           grn.OwnerID,
		WarehouseLocation: grn.WarehouseLocation,
		Notes:             grn.Notes,
		CurrentStage:      grn.ApprovalStage,
		StageName:         grn.StageName,
		ApprovedBy:        grn.ApprovedBy,
		AutomationUsed:    grn.AutomationUsed,
		AutoCreatedPV:     autoCreatedPV,
		Metadata:          metadata,

		ConsignmentNote:      grn.ConsignmentNote,
		VendorName:           grn.VendorName,
		VendorAddress:        grn.VendorAddress,
		ReceivedByName:       grn.ReceivedByName,
		ReceivedBySignature:  grn.ReceivedBySignature,
		ReceivedAt:           grn.ReceivedAt,
		CertifiedByID:        grn.CertifiedByID,
		CertifiedByName:      grn.CertifiedByName,
		CertifiedBySignature: grn.CertifiedBySignature,
		CertifiedAt:          grn.CertifiedAt,
		SignoffStatus:        grn.SignoffStatus,
		StampImageURL:        grn.StampImageURL,

		CreatedAt:         grn.CreatedAt,
		UpdatedAt:         grn.UpdatedAt,
	}
}

// SubmitGRN submits a GRN for approval using the workflow system
func SubmitGRN(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "GRN ID is required",
		})
	}

	// Get organization ID and user ID from context
	organizationID := c.Locals("organizationID").(string)
	userID := c.Locals("userID").(string)

	var submitReq types.SubmitDocumentRequest
	if err := c.BodyParser(&submitReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
		})
	}
	if submitReq.WorkflowID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "workflowId is required",
		})
	}

	// Get existing GRN
	var grn models.GoodsReceivedNote
	if err := config.DB.Where("id = ? AND organization_id = ?", id, organizationID).First(&grn).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "GRN not found",
		})
	}

	// Check if GRN is in draft status
	if strings.ToUpper(grn.Status) != "DRAFT" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": fmt.Sprintf("Cannot submit GRN in %s status", grn.Status),
		})
	}

	// Workflow is optional and can only be triggered after both the receiver
	// and the certifying officer have signed the GRN.
	if grn.SignoffStatus != "READY" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "GRN must be signed by both the receiver and a certifying officer before it can be submitted to a workflow",
		})
	}

	// Gate: linked PO must still be APPROVED and linked PV (if any) APPROVED/PAID.
	if msg := revalidateGRNLinks(&grn, organizationID, "submit"); msg != "" {
		return utils.SendBadRequestError(c, msg)
	}

	// Get workflow execution service from context
	workflowExecutionService := c.Locals("workflowExecutionService").(*services.WorkflowExecutionService)

	// Atomic submit: status change + workflow assignment in a single transaction.
	tx := config.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	assignment, err := workflowExecutionService.AssignWorkflowToDocumentWithIDTx(
		c.Context(), tx, organizationID, grn.ID, "grn", submitReq.WorkflowID, userID,
	)
	if err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to assign workflow to GRN",
			"error":   err.Error(),
		})
	}

	grn.Status = models.StatusPending
	grn.UpdatedAt = time.Now()

	// Use the open transaction's connection so we don't deadlock against
	// the in-flight tx when the connection pool is restricted to a single
	// conn (e.g. SQLite-backed unit test harness).
	var user models.User
	_ = tx.Where("id = ?", userID).First(&user).Error
	actionHistory := grn.ActionHistory.Data()
	actionHistory = append(actionHistory, types.ActionHistoryEntry{
		ID:              uuid.New().String(),
		Action:          "SUBMIT",
		PerformedBy:     userID,
		PerformedByName: user.Name,
		PerformedByRole: user.Role,
		Timestamp:       time.Now(),
		Comments:        "GRN submitted for approval",
		ActionType:      "SUBMIT",
		PreviousStatus:  models.StatusDraft,
		NewStatus:       models.StatusPending,
	})
	grn.ActionHistory = datatypes.NewJSONType(actionHistory)

	if err := tx.Save(&grn).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to update GRN status",
			"error":   err.Error(),
		})
	}

	if err := tx.Commit().Error; err != nil {
		return utils.SendInternalError(c, "Failed to submit GRN", err)
	}

	// Preload purchase order and vendor
	config.DB.Preload("PurchaseOrder").Preload("Vendor").First(&grn)

	go utils.SyncDocumentAs(config.DB, "GRN", grn.ID, userID)
	go services.LogDocumentEvent(config.DB, services.DocumentEvent{
		OrganizationID: organizationID,
		DocumentID:     grn.ID,
		DocumentType:   "grn",
		UserID:         userID,
		ActorName:      user.Name,
		ActorRole:      user.Role,
		Action:         "submitted",
		Details:        map[string]interface{}{"documentNumber": grn.DocumentNumber},
	})

	return c.JSON(types.DetailResponse{
		Success: true,
		Data: fiber.Map{
			"grn": modelToGRNResponse(grn),
			"workflow": fiber.Map{
				"assignmentId": assignment.ID,
				"workflowId":   assignment.WorkflowID,
				"currentStage": assignment.CurrentStage,
				"status":       assignment.Status,
			},
		},
	})
}

// privilegedGRNCertifierRoles is the canonical set of roles that may certify
// a GRN as an "issuing officer". Matches the canonical SystemRole values used
// across the rest of the app (see frontend/src/types/core.ts).
var privilegedGRNCertifierRoles = map[string]bool{
	"admin":       true,
	"super_admin": true,
	"manager":     true,
	"finance":     true,
	"approver":    true,
}

// SignReceiveGRN captures the receiver's name + digital signature, moving the
// sign-off state from PENDING_RECEIVER -> PENDING_CERTIFIER.
func SignReceiveGRN(c *fiber.Ctx) error {
	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return utils.SendUnauthorizedError(c, "Organization context required")
	}

	id := c.Params("id")
	if id == "" {
		return utils.SendBadRequestError(c, "GRN ID is required")
	}

	var req types.SignReceiveGRNRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendBadRequestError(c, "Invalid request body")
	}
	if strings.TrimSpace(req.ReceivedByName) == "" || strings.TrimSpace(req.Signature) == "" {
		return utils.SendBadRequestError(c, "receivedByName and signature are required")
	}

	var grn models.GoodsReceivedNote
	if err := config.DB.
		Where("id = ? AND organization_id = ?", id, tenant.OrganizationID).
		First(&grn).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "GRN not found",
		})
	}

	if strings.ToUpper(grn.Status) != "DRAFT" {
		return utils.SendBadRequestError(c, fmt.Sprintf("Cannot sign GRN in %s status", grn.Status))
	}
	if grn.SignoffStatus != "PENDING_RECEIVER" {
		return utils.SendBadRequestError(c, fmt.Sprintf("Receiver sign-off not allowed in state %s", grn.SignoffStatus))
	}

	now := time.Now()
	grn.ReceivedByName = req.ReceivedByName
	grn.ReceivedBySignature = req.Signature
	grn.ReceivedAt = &now
	if grn.ReceivedBy == "" {
		grn.ReceivedBy = tenant.UserID
	}
	grn.SignoffStatus = "PENDING_CERTIFIER"
	grn.UpdatedAt = now

	var actor models.User
	_ = config.DB.Where("id = ?", tenant.UserID).First(&actor).Error
	history := grn.ActionHistory.Data()
	history = append(history, types.ActionHistoryEntry{
		ID:              uuid.New().String(),
		Action:          "RECEIVED_SIGNOFF",
		ActionType:      "SIGNOFF",
		PerformedBy:     tenant.UserID,
		PerformedByName: actor.Name,
		PerformedByRole: actor.Role,
		Timestamp:       now,
		PerformedAt:     now,
		Comments:        fmt.Sprintf("GRN received and signed by %s", req.ReceivedByName),
		NewStatus:       grn.Status,
	})
	grn.ActionHistory = datatypes.NewJSONType(history)

	if err := config.DB.Save(&grn).Error; err != nil {
		return utils.SendInternalError(c, "Failed to record receiver sign-off", err)
	}

	go services.LogDocumentEvent(config.DB, services.DocumentEvent{
		OrganizationID: tenant.OrganizationID,
		DocumentID:     grn.ID,
		DocumentType:   "grn",
		UserID:         tenant.UserID,
		ActorName:      actor.Name,
		ActorRole:      tenant.UserRole,
		Action:         "received_signoff",
		Details:        map[string]interface{}{"documentNumber": grn.DocumentNumber},
	})

	return c.JSON(types.DetailResponse{Success: true, Data: modelToGRNResponse(grn)})
}

// CertifyGRN captures the issuing officer's certification. Requires a
// privileged role and that the receiver has already signed.
// Moves PENDING_CERTIFIER -> READY.
func CertifyGRN(c *fiber.Ctx) error {
	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return utils.SendUnauthorizedError(c, "Organization context required")
	}

	id := c.Params("id")
	if id == "" {
		return utils.SendBadRequestError(c, "GRN ID is required")
	}

	if !privilegedGRNCertifierRoles[strings.ToLower(tenant.UserRole)] {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "Only admin, manager, finance or approver roles may certify a GRN",
		})
	}

	var req types.CertifyGRNRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendBadRequestError(c, "Invalid request body")
	}
	if strings.TrimSpace(req.Signature) == "" {
		return utils.SendBadRequestError(c, "signature is required")
	}

	var grn models.GoodsReceivedNote
	if err := config.DB.
		Where("id = ? AND organization_id = ?", id, tenant.OrganizationID).
		First(&grn).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "GRN not found",
		})
	}

	if strings.ToUpper(grn.Status) != "DRAFT" {
		return utils.SendBadRequestError(c, fmt.Sprintf("Cannot certify GRN in %s status", grn.Status))
	}
	if grn.SignoffStatus != "PENDING_CERTIFIER" {
		return utils.SendBadRequestError(c, fmt.Sprintf("Certifier sign-off not allowed in state %s", grn.SignoffStatus))
	}
	// Separation-of-duties: the certifier cannot be the same user as the creator
	// or the receiver, mirroring the two-signature requirement on the printed form.
	if grn.CreatedBy == tenant.UserID || grn.ReceivedBy == tenant.UserID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "The certifying officer must be different from the GRN creator and receiver",
		})
	}

	var certifier models.User
	if err := config.DB.Where("id = ?", tenant.UserID).First(&certifier).Error; err != nil {
		return utils.SendInternalError(c, "Failed to load certifier profile", err)
	}

	now := time.Now()
	grn.CertifiedByID = tenant.UserID
	grn.CertifiedByName = certifier.Name
	grn.CertifiedBySignature = req.Signature
	grn.CertifiedAt = &now
	grn.SignoffStatus = "READY"
	grn.UpdatedAt = now
	if strings.TrimSpace(req.StampImageURL) != "" {
		grn.StampImageURL = req.StampImageURL
	}

	history := grn.ActionHistory.Data()
	comments := req.Comments
	if comments == "" {
		comments = fmt.Sprintf("GRN certified by %s", certifier.Name)
	}
	history = append(history, types.ActionHistoryEntry{
		ID:              uuid.New().String(),
		Action:          "CERTIFIED",
		ActionType:      "SIGNOFF",
		PerformedBy:     tenant.UserID,
		PerformedByName: certifier.Name,
		PerformedByRole: certifier.Role,
		Timestamp:       now,
		PerformedAt:     now,
		Comments:        comments,
		NewStatus:       grn.Status,
	})
	grn.ActionHistory = datatypes.NewJSONType(history)

	if err := config.DB.Save(&grn).Error; err != nil {
		return utils.SendInternalError(c, "Failed to record certifier sign-off", err)
	}

	// Optional auto-submit: when the org has enabled AutoSubmitGRNToWorkflow,
	// hand the GRN straight to the default workflow on certification.
	// (Manual submit endpoint is still available.)
	orgSvc := services.NewOrganizationService(config.DB)
	orgSettings, _ := orgSvc.GetOrganizationSettings(tenant.OrganizationID)
	if orgSettings != nil && orgSettings.AutoSubmitGRNToWorkflow {
		if wfSvc, ok := c.Locals("workflowExecutionService").(*services.WorkflowExecutionService); ok && wfSvc != nil {
			if wfRegSvc := services.NewWorkflowService(nil, nil, config.DB); wfRegSvc != nil {
				if defaultWF, _ := wfRegSvc.GetDefaultWorkflowForEntity(tenant.OrganizationID, "grn"); defaultWF != nil {
					autoTx := config.DB.Begin()
					autoSubmitOK := false
					if autoTx.Error != nil {
						// Couldn't open the transaction — skip auto-submit; the
						// GRN stays DRAFT and can be submitted manually.
					} else if _, err := wfSvc.AssignWorkflowToDocumentWithIDTx(
						c.Context(), autoTx, tenant.OrganizationID, grn.ID, "grn",
						defaultWF.ID.String(), tenant.UserID,
					); err != nil {
						autoTx.Rollback()
					} else if updErr := autoTx.Model(&models.GoodsReceivedNote{}).
						Where("id = ?", grn.ID).
						Update("status", models.StatusPending).Error; updErr != nil {
						// Status flip failed: roll back so we don't leave a workflow
						// assigned to a GRN that's still DRAFT, and don't report success.
						autoTx.Rollback()
					} else if commitErr := autoTx.Commit().Error; commitErr == nil {
						autoSubmitOK = true
					}

					// Post-commit: append a system-actor audit entry so the
					// trail reflects that no human signed off this submission.
					if autoSubmitOK {
						var autoGRN models.GoodsReceivedNote
						if err := config.DB.Where("id = ?", grn.ID).First(&autoGRN).Error; err == nil {
							autoNow := time.Now()
							autoHistory := autoGRN.ActionHistory.Data()
							autoHistory = append(autoHistory, types.ActionHistoryEntry{
								ID:              uuid.New().String(),
								Action:          "AUTO_SUBMIT",
								ActionType:      "SUBMIT",
								PerformedBy:     "system",
								PerformedByName: "System (auto-submit)",
								PerformedByRole: "system",
								Timestamp:       autoNow,
								PerformedAt:     autoNow,
								PreviousStatus:  models.StatusDraft,
								NewStatus:       models.StatusPending,
								Comments:        "Auto-submitted via AutoSubmitGRNToWorkflow org setting",
								Metadata: map[string]interface{}{
									"triggeredBy":    tenant.UserID,
									"orgSettingFlag": "AutoSubmitGRNToWorkflow",
								},
							})
							autoGRN.ActionHistory = datatypes.NewJSONType(autoHistory)
							_ = config.DB.Model(&autoGRN).
								Where("id = ?", autoGRN.ID).
								Update("action_history", autoGRN.ActionHistory).Error
						}
					}
				}
			}
		}
	}

	go services.LogDocumentEvent(config.DB, services.DocumentEvent{
		OrganizationID: tenant.OrganizationID,
		DocumentID:     grn.ID,
		DocumentType:   "grn",
		UserID:         tenant.UserID,
		ActorName:      certifier.Name,
		ActorRole:      tenant.UserRole,
		Action:         "certified",
		Details:        map[string]interface{}{"documentNumber": grn.DocumentNumber},
	})

	return c.JSON(types.DetailResponse{Success: true, Data: modelToGRNResponse(grn)})
}

// revalidateGRNLinks ensures, just before a completion path cascades, that the
// GRN's linked PO is still APPROVED and the linked PV (if any) is still APPROVED
// or PAID. Returns "" when valid, otherwise a caller-facing message. `verb`
// ("submit" / "complete") is interpolated so each path reads naturally. Shared
// by SubmitGRN and MarkGRNComplete so both enforce identical preconditions.
func revalidateGRNLinks(grn *models.GoodsReceivedNote, orgID, verb string) string {
	if grn.PODocumentNumber != "" {
		var linkedPO models.PurchaseOrder
		if err := config.DB.
			Where("document_number = ? AND organization_id = ?", grn.PODocumentNumber, orgID).
			First(&linkedPO).Error; err != nil {
			return "Linked purchase order not found"
		}
		if strings.ToUpper(linkedPO.Status) != "APPROVED" {
			return fmt.Sprintf("Cannot %s GRN: linked PO %s is in %s status and must be APPROVED.",
				verb, grn.PODocumentNumber, linkedPO.Status)
		}
	}
	if grn.LinkedPV != "" {
		var linkedPV models.PaymentVoucher
		if err := config.DB.
			Where("document_number = ? AND organization_id = ?", grn.LinkedPV, orgID).
			First(&linkedPV).Error; err != nil {
			return "Linked payment voucher not found"
		}
		pvStatus := strings.ToUpper(linkedPV.Status)
		if pvStatus != "APPROVED" && pvStatus != "PAID" {
			return fmt.Sprintf("Cannot %s GRN: linked PV %s is in %s status and must be APPROVED or PAID.",
				verb, grn.LinkedPV, linkedPV.Status)
		}
	}
	return ""
}

// MarkGRNComplete closes the GRN without workflow approval. The two
// signatures (receiver + certifier) stand in for an approval chain.
// Requires SignoffStatus = READY and Status = DRAFT.
func MarkGRNComplete(c *fiber.Ctx) error {
	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return utils.SendUnauthorizedError(c, "Organization context required")
	}

	id := c.Params("id")
	if id == "" {
		return utils.SendBadRequestError(c, "GRN ID is required")
	}

	var req types.CompleteGRNRequest
	_ = c.BodyParser(&req)

	var grn models.GoodsReceivedNote
	if err := config.DB.
		Where("id = ? AND organization_id = ?", id, tenant.OrganizationID).
		First(&grn).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "GRN not found",
		})
	}

	if grn.SignoffStatus != "READY" {
		return utils.SendBadRequestError(c, "GRN must be signed by both the receiver and a certifying officer before it can be marked complete")
	}
	if strings.ToUpper(grn.Status) != "DRAFT" {
		return utils.SendBadRequestError(c, fmt.Sprintf("Cannot complete GRN in %s status", grn.Status))
	}

	// Re-validate links before cascading — a GRN signed READY while its PO was
	// APPROVED must not force-complete if the PO/PV later changed state.
	if msg := revalidateGRNLinks(&grn, tenant.OrganizationID, "complete"); msg != "" {
		return utils.SendBadRequestError(c, msg)
	}

	now := time.Now()
	grn.Status = models.StatusCompleted
	grn.SignoffStatus = "COMPLETED"
	grn.UpdatedAt = now

	tx := config.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var actor models.User
	_ = tx.Where("id = ?", tenant.UserID).First(&actor).Error
	history := grn.ActionHistory.Data()
	comments := req.Comments
	if comments == "" {
		comments = "GRN marked complete (workflow skipped)"
	}
	history = append(history, types.ActionHistoryEntry{
		ID:              uuid.New().String(),
		Action:          "COMPLETE_NO_WORKFLOW",
		ActionType:      "COMPLETE",
		PerformedBy:     tenant.UserID,
		PerformedByName: actor.Name,
		PerformedByRole: actor.Role,
		Timestamp:       now,
		PerformedAt:     now,
		Comments:        comments,
		PreviousStatus:  models.StatusDraft,
		NewStatus:       models.StatusCompleted,
	})
	grn.ActionHistory = datatypes.NewJSONType(history)

	if err := tx.Save(&grn).Error; err != nil {
		tx.Rollback()
		return utils.SendInternalError(c, "Failed to complete GRN", err)
	}

	// Cascade the same way the workflow terminal-approve path does — keeps
	// PO.delivery_status, per-item received quantities, and PO-terminal
	// auto-completion in sync regardless of which path closed the GRN.
	if wfSvc, ok := c.Locals("workflowExecutionService").(*services.WorkflowExecutionService); ok && wfSvc != nil {
		if err := wfSvc.CascadeGRNApprovalToPO(tx, grn.ID); err != nil {
			tx.Rollback()
			return utils.SendInternalError(c, "Failed to cascade GRN completion to PO", err)
		}
		// Mirror the workflow path: fire AutoCreatePVFromPO if enabled.
		if err := wfSvc.AutoCreatePVFromCompletedGRN(tx, grn.ID); err != nil {
			fmt.Printf("Warning: AutoCreatePVFromCompletedGRN failed: %v\n", err)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return utils.SendInternalError(c, "Failed to commit GRN completion", err)
	}

	// Post-commit: honor PVAutomationLevel on the PV that AutoCreatePVFromCompletedGRN
	// created in-tx (submit / auto-approve), matching the workflow GRN-completion path.
	if wfSvc, ok := c.Locals("workflowExecutionService").(*services.WorkflowExecutionService); ok && wfSvc != nil {
		if err := wfSvc.ApplyPVAutomationForCompletedGRN(context.Background(), grn.ID); err != nil {
			fmt.Printf("Warning: ApplyPVAutomationForCompletedGRN failed: %v\n", err)
		}
	}

	go utils.SyncDocumentAs(config.DB, "GRN", grn.ID, tenant.UserID)
	go services.LogDocumentEvent(config.DB, services.DocumentEvent{
		OrganizationID: tenant.OrganizationID,
		DocumentID:     grn.ID,
		DocumentType:   "grn",
		UserID:         tenant.UserID,
		ActorName:      actor.Name,
		ActorRole:      tenant.UserRole,
		Action:         "completed_no_workflow",
		Details:        map[string]interface{}{"documentNumber": grn.DocumentNumber},
	})

	return c.JSON(types.DetailResponse{Success: true, Data: modelToGRNResponse(grn)})
}
