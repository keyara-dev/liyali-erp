package handlers

// extras_reports_http_test.go — supplementary tests for:
//   - document_extras_handler.go: CreatePurchaseOrderFromRequisition
//     (zero amount, vendor not found, success, default currency),
//     CreatePaymentVoucherFromPO (missing PO id, zero amount, PO not found,
//     goods_first variants, payment_first success, GRN wrong PO),
//     Stats handlers with seeded data,
//     ConfirmGRN status persistence + draft rejection
//   - reports.go: GetSystemStatistics, GetApprovalMetrics,
//     GetUserActivityMetrics, GetAnalyticsDashboard, GetDashboardReports
//     (all roles, via a stub that avoids pgxpool)

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	fiberrecover "github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/services"
	"github.com/liyali/liyali-gateway/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
)

// emptyPOItems / emptyGRNItems / emptyActionHistory / emptyApprovalHistory
// are typed helpers to avoid verbose inline type assertions.
func emptyPOItems() datatypes.JSONType[[]types.POItem] {
	return datatypes.NewJSONType([]types.POItem{})
}
func emptyGRNItems() datatypes.JSONType[[]types.GRNItem] {
	return datatypes.NewJSONType([]types.GRNItem{})
}
func emptyReqItems() datatypes.JSONType[[]types.RequisitionItem] {
	return datatypes.NewJSONType([]types.RequisitionItem{})
}
func emptyActionHistory() datatypes.JSONType[[]types.ActionHistoryEntry] {
	return datatypes.NewJSONType([]types.ActionHistoryEntry{})
}
func emptyApprovalHistory() datatypes.JSONType[[]types.ApprovalRecord] {
	return datatypes.NewJSONType([]types.ApprovalRecord{})
}

// ─────────────────────────────────────────────────────────────────────────────
// DB helpers
// ─────────────────────────────────────────────────────────────────────────────

// setupOrgSettingsForExtras creates the organization_settings table needed by
// CreatePaymentVoucherFromPO when it looks up the org procurement flow.
func setupOrgSettingsForExtras(t *testing.T) {
	t.Helper()
	sql := `CREATE TABLE IF NOT EXISTS organization_settings (
		id TEXT PRIMARY KEY,
		organization_id TEXT UNIQUE,
		require_digital_signatures BOOLEAN DEFAULT true,
		default_approval_chain TEXT,
		currency TEXT DEFAULT 'USD',
		fiscal_year_start INTEGER DEFAULT 1,
		enable_budget_validation BOOLEAN DEFAULT true,
		budget_variance_threshold REAL DEFAULT 5.0,
		procurement_flow TEXT DEFAULT 'payment_first',
		created_at DATETIME,
		updated_at DATETIME
	)`
	if err := config.DB.Exec(sql).Error; err != nil {
		t.Fatalf("setupOrgSettingsForExtras: %v", err)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// App factories
// ─────────────────────────────────────────────────────────────────────────────

func newExtrasTestApp() *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		},
	})
	app.Use(fiberrecover.New())

	grp := app.Group("/api/v1", withTenantCtx(testOrgID, testUserID, testUserRole))
	grp.Post("/purchase-orders/from-requisition", CreatePurchaseOrderFromRequisition)
	grp.Post("/payment-vouchers/from-po", CreatePaymentVoucherFromPO)
	grp.Get("/requisitions/stats", GetRequisitionStats)
	grp.Get("/purchase-orders/stats", GetPurchaseOrderStats)
	grp.Get("/payment-vouchers/stats", GetPaymentVoucherStats)
	// ConfirmGRN removed; workflow auto-cascades APPROVED → COMPLETED.
	return app
}

// ─────────────────────────────────────────────────────────────────────────────
// Stub ReportsService — lets us test the handler layer without pgxpool.
//
// ReportsService uses a *repository.ReportsRepository backed by pgxpool.Pool,
// so we cannot use the real service against SQLite. Instead we wire a thin
// Fiber app (newReportsAppWithStub) that mirrors each handler's auth/role guard
// logic and delegates to stub functions for the service calls.
// newReportsAppViaHandler uses the REAL ReportsHandler with a nil repo to
// exercise only the early-return 403 paths cheaply (no service called there).
// ─────────────────────────────────────────────────────────────────────────────

type reportsHandlerStub struct {
	getSystemStatsFn     func(ctx context.Context, orgID, start, end string) (*models.SystemStatistics, error)
	getApprovalMetricsFn func(ctx context.Context, orgID, start, end string) (*models.ApprovalMetrics, error)
	getUserActivityFn    func(ctx context.Context, orgID string) (*models.UserActivityMetrics, error)
	getAnalyticsFn       func(ctx context.Context, orgID, start, end string) (*models.AnalyticsDashboard, error)
}

type extrasStubTenant struct {
	OrganizationID string
	UserID         string
	UserRole       string
}

func readExtrasStubTenant(c *fiber.Ctx) (*extrasStubTenant, error) {
	if c.Locals("tenant") == nil {
		return nil, fiber.ErrUnauthorized
	}
	orgID, _ := c.Locals("organizationID").(string)
	userID, _ := c.Locals("userID").(string)
	role, _ := c.Locals("userRole").(string)
	return &extrasStubTenant{OrganizationID: orgID, UserID: userID, UserRole: role}, nil
}

func defaultEmptyStats() *models.SystemStatistics {
	return &models.SystemStatistics{
		DocumentTypeBreakdown: models.DocumentTypeBreakdown{},
		StatusBreakdown:       models.StatusBreakdown{},
	}
}

func defaultEmptyApprovalMetrics() *models.ApprovalMetrics {
	return &models.ApprovalMetrics{RecentApprovals: []models.ApprovalActivity{}}
}

func defaultStub() *reportsHandlerStub {
	return &reportsHandlerStub{
		getSystemStatsFn: func(_ context.Context, _, _, _ string) (*models.SystemStatistics, error) {
			return defaultEmptyStats(), nil
		},
		getApprovalMetricsFn: func(_ context.Context, _, _, _ string) (*models.ApprovalMetrics, error) {
			return defaultEmptyApprovalMetrics(), nil
		},
		getUserActivityFn: func(_ context.Context, _ string) (*models.UserActivityMetrics, error) {
			return &models.UserActivityMetrics{Users: []models.UserActivity{}}, nil
		},
		getAnalyticsFn: func(_ context.Context, _, _, _ string) (*models.AnalyticsDashboard, error) {
			return &models.AnalyticsDashboard{
				ApprovalTrends:       []models.ApprovalTrend{},
				DocumentDistribution: []models.DocumentDistribution{},
				StageMetrics:         []models.StageMetric{},
			}, nil
		},
	}
}

func newReportsAppWithStub(role string, stub *reportsHandlerStub) *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		},
	})
	app.Use(fiberrecover.New())

	withRole := withTenantCtx(testOrgID, testUserID, role)

	app.Get("/reports/system-stats", withRole, func(c *fiber.Ctx) error {
		tc, err := readExtrasStubTenant(c)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"success": false})
		}
		if tc.UserRole != "admin" && tc.UserRole != "superadmin" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "message": "Admin access required"})
		}
		stats, err := stub.getSystemStatsFn(c.Context(), tc.OrganizationID, c.Query("start_date"), c.Query("end_date"))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false})
		}
		return c.JSON(stats)
	})

	app.Get("/reports/approval-metrics", withRole, func(c *fiber.Ctx) error {
		tc, err := readExtrasStubTenant(c)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"success": false})
		}
		if tc.UserRole != "admin" && tc.UserRole != "superadmin" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "message": "Admin access required"})
		}
		m, err := stub.getApprovalMetricsFn(c.Context(), tc.OrganizationID, c.Query("start_date"), c.Query("end_date"))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false})
		}
		return c.JSON(m)
	})

	app.Get("/reports/user-activity", withRole, func(c *fiber.Ctx) error {
		tc, err := readExtrasStubTenant(c)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"success": false})
		}
		if tc.UserRole != "admin" && tc.UserRole != "superadmin" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "message": "Admin access required"})
		}
		m, err := stub.getUserActivityFn(c.Context(), tc.OrganizationID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false})
		}
		return c.JSON(m)
	})

	app.Get("/reports/analytics", withRole, func(c *fiber.Ctx) error {
		tc, err := readExtrasStubTenant(c)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"success": false})
		}
		if tc.UserRole != "admin" && tc.UserRole != "superadmin" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "message": "Admin access required"})
		}
		a, err := stub.getAnalyticsFn(c.Context(), tc.OrganizationID, c.Query("start_date"), c.Query("end_date"))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false})
		}
		return c.JSON(a)
	})

	app.Get("/reports/dashboard", withRole, func(c *fiber.Ctx) error {
		tc, err := readExtrasStubTenant(c)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"success": false})
		}
		stats, err := stub.getSystemStatsFn(c.Context(), tc.OrganizationID, c.Query("start_date"), c.Query("end_date"))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false})
		}
		recentApprovals, _ := stub.getApprovalMetricsFn(c.Context(), tc.OrganizationID, c.Query("start_date"), c.Query("end_date"))
		dashboard := fiber.Map{
			"organizationId":        tc.OrganizationID,
			"userRole":              tc.UserRole,
			"totalDocuments":        stats.TotalDocuments,
			"approvedDocuments":     stats.ApprovedDocuments,
			"rejectedDocuments":     stats.RejectedDocuments,
			"draftDocuments":        stats.DraftDocuments,
			"submittedDocuments":    stats.SubmittedDocuments,
			"pendingApproval":       stats.PendingApproval,
			"averageApprovalTime":   stats.AverageApprovalTime,
			"averageProcessingTime": stats.AverageProcessingTime,
			"approvalRate":          stats.ApprovalRate,
			"rejectionRate":         stats.RejectionRate,
			"budgetUtilization":     stats.BudgetUtilization,
		}
		if recentApprovals != nil {
			dashboard["recentActivity"] = recentApprovals.RecentApprovals
		} else {
			dashboard["recentActivity"] = []interface{}{}
		}
		return c.JSON(fiber.Map{"success": true, "message": "Dashboard reports retrieved successfully", "data": dashboard})
	})

	return app
}

// newReportsAppViaHandler wires the REAL ReportsHandler (nil repo) so we can
// exercise the 403 role-guard paths without touching the service.
func newReportsAppViaHandler(role string) *fiber.App {
	svc := services.NewReportsService(nil)
	h := NewReportsHandler(svc)

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		},
	})
	app.Use(fiberrecover.New())

	withRole := withTenantCtx(testOrgID, testUserID, role)
	app.Get("/reports/system-stats", withRole, h.GetSystemStatistics)
	app.Get("/reports/approval-metrics", withRole, h.GetApprovalMetrics)
	app.Get("/reports/user-activity", withRole, h.GetUserActivityMetrics)
	app.Get("/reports/analytics", withRole, h.GetAnalyticsDashboard)
	app.Get("/reports/dashboard", withRole, h.GetDashboardReports)
	return app
}

// ─────────────────────────────────────────────────────────────────────────────
// TestCreatePurchaseOrderFromRequisition — new paths only
// (NoAuth/MissingReqID/MissingItems already in purchase_orders_http_test.go)
// ─────────────────────────────────────────────────────────────────────────────

func TestCreatePOFromRequisition_TotalAmountZero(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newExtrasTestApp()
	resp := testRequest(app, http.MethodPost, "/api/v1/purchase-orders/from-requisition", map[string]interface{}{
		"requisitionId": "req-001",
		"items":         []map[string]interface{}{{"description": "item"}},
		"totalAmount":   0,
	})
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	body := decodeResponse(resp)
	require.NotNil(t, body)
	assert.False(t, body["success"].(bool))
}

func TestCreatePOFromRequisition_VendorNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newExtrasTestApp()
	resp := testRequest(app, http.MethodPost, "/api/v1/purchase-orders/from-requisition", map[string]interface{}{
		"requisitionId": "req-001",
		"vendorId":      "nonexistent-vendor",
		"items":         []map[string]interface{}{{"description": "item", "quantity": 1, "unitPrice": 50}},
		"totalAmount":   50.0,
		"currency":      "ZMW",
	})
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	body := decodeResponse(resp)
	require.NotNil(t, body)
	assert.False(t, body["success"].(bool))
}

func TestCreatePOFromRequisition_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	req := models.Requisition{
		ID:             "req-chain-001",
		OrganizationID: testOrgID,
		DocumentNumber: "REQ-CHAIN-001",
		Status:         "APPROVED",
		TotalAmount:    200.0,
		Currency:       "ZMW",
	}
	req.ActionHistory = emptyActionHistory()
	require.NoError(t, db.Create(&req).Error)

	app := newExtrasTestApp()
	resp := testRequest(app, http.MethodPost, "/api/v1/purchase-orders/from-requisition", map[string]interface{}{
		"requisitionId":             "req-chain-001",
		"requisitionDocumentNumber": "REQ-CHAIN-001",
		"title":                     "Test PO from REQ",
		"items": []map[string]interface{}{
			{"description": "Chairs", "quantity": 2, "unitPrice": 100.0},
		},
		"totalAmount": 200.0,
		"currency":    "ZMW",
	})
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	body := decodeResponse(resp)
	require.NotNil(t, body)
	assert.True(t, body["success"].(bool))
}

func TestCreatePOFromRequisition_DefaultCurrency(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newExtrasTestApp()
	// currency omitted — handler defaults to "ZMW"
	resp := testRequest(app, http.MethodPost, "/api/v1/purchase-orders/from-requisition", map[string]interface{}{
		"requisitionId": "req-no-currency",
		"title":         "No Currency PO",
		"items": []map[string]interface{}{
			{"description": "Widget", "quantity": 1, "unitPrice": 75.0},
		},
		"totalAmount": 75.0,
	})
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	body := decodeResponse(resp)
	require.NotNil(t, body)
	assert.True(t, body["success"].(bool))
}

// ─────────────────────────────────────────────────────────────────────────────
// TestCreatePaymentVoucherFromPO — new paths only
// (NoAuth already in payment_vouchers_http_test.go)
// ─────────────────────────────────────────────────────────────────────────────

func TestCreatePVFromPO_MissingPurchaseOrderID(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newExtrasTestApp()
	resp := testRequest(app, http.MethodPost, "/api/v1/payment-vouchers/from-po", map[string]interface{}{
		"totalAmount": 500.0,
	})
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	body := decodeResponse(resp)
	require.NotNil(t, body)
	assert.False(t, body["success"].(bool))
}

func TestCreatePVFromPO_TotalAmountZero(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newExtrasTestApp()
	resp := testRequest(app, http.MethodPost, "/api/v1/payment-vouchers/from-po", map[string]interface{}{
		"purchaseOrderId": "po-001",
		"totalAmount":     0,
	})
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestCreatePVFromPO_PONotFound(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupOrgSettingsForExtras(t)

	app := newExtrasTestApp()
	resp := testRequest(app, http.MethodPost, "/api/v1/payment-vouchers/from-po", map[string]interface{}{
		"purchaseOrderId": "nonexistent-po-999",
		"totalAmount":     500.0,
	})
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	body := decodeResponse(resp)
	require.NotNil(t, body)
	assert.False(t, body["success"].(bool))
}

func TestCreatePVFromPO_GoodsFirst_MissingGRN(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupOrgSettingsForExtras(t)

	po := models.PurchaseOrder{
		ID: "po-gf-001", OrganizationID: testOrgID, DocumentNumber: "PO-GF-001",
		Status: "APPROVED", TotalAmount: 500.0, Currency: "ZMW", ProcurementFlow: "goods_first",
	}
	po.Items = emptyPOItems()
	po.ActionHistory = emptyActionHistory()
	require.NoError(t, db.Create(&po).Error)

	app := newExtrasTestApp()
	resp := testRequest(app, http.MethodPost, "/api/v1/payment-vouchers/from-po", map[string]interface{}{
		"purchaseOrderId":             "po-gf-001",
		"purchaseOrderDocumentNumber": "PO-GF-001",
		"totalAmount":                 500.0,
		// linkedGRNDocumentNumber intentionally omitted
	})
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	body := decodeResponse(resp)
	require.NotNil(t, body)
	assert.False(t, body["success"].(bool))
}

func TestCreatePVFromPO_GoodsFirst_GRNNotApproved(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupOrgSettingsForExtras(t)

	po := models.PurchaseOrder{
		ID: "po-gf-003", OrganizationID: testOrgID, DocumentNumber: "PO-GF-003",
		Status: "APPROVED", TotalAmount: 100.0, Currency: "ZMW", ProcurementFlow: "goods_first",
	}
	po.Items = emptyPOItems()
	po.ActionHistory = emptyActionHistory()
	require.NoError(t, db.Create(&po).Error)

	grn := models.GoodsReceivedNote{
		ID: "grn-pending-test", OrganizationID: testOrgID, DocumentNumber: "GRN-PENDING-TEST",
		PODocumentNumber: "PO-GF-003", Status: "PENDING",
		ReceivedDate: time.Now(), ReceivedBy: testUserID,
	}
	grn.Items = emptyGRNItems()
	grn.ActionHistory = emptyActionHistory()
	require.NoError(t, db.Create(&grn).Error)

	app := newExtrasTestApp()
	resp := testRequest(app, http.MethodPost, "/api/v1/payment-vouchers/from-po", map[string]interface{}{
		"purchaseOrderId":             "po-gf-003",
		"purchaseOrderDocumentNumber": "PO-GF-003",
		"totalAmount":                 100.0,
		"linkedGRNDocumentNumber":     "GRN-PENDING-TEST",
	})
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestCreatePVFromPO_GoodsFirst_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupOrgSettingsForExtras(t)

	po := models.PurchaseOrder{
		ID: "po-gf-002", OrganizationID: testOrgID, DocumentNumber: "PO-GF-002",
		Status: "APPROVED", TotalAmount: 300.0, Currency: "ZMW", ProcurementFlow: "goods_first",
	}
	po.Items = emptyPOItems()
	po.ActionHistory = emptyActionHistory()
	require.NoError(t, db.Create(&po).Error)

	grn := models.GoodsReceivedNote{
		ID: "grn-approved-test", OrganizationID: testOrgID, DocumentNumber: "GRN-APPROVED-TEST",
		PODocumentNumber: "PO-GF-002", Status: "APPROVED",
		ReceivedDate: time.Now(), ReceivedBy: testUserID,
	}
	grn.Items = emptyGRNItems()
	grn.ActionHistory = emptyActionHistory()
	require.NoError(t, db.Create(&grn).Error)

	app := newExtrasTestApp()
	resp := testRequest(app, http.MethodPost, "/api/v1/payment-vouchers/from-po", map[string]interface{}{
		"purchaseOrderId":             "po-gf-002",
		"purchaseOrderDocumentNumber": "PO-GF-002",
		"title":                       "PV from GRN",
		"totalAmount":                 300.0,
		"currency":                    "ZMW",
		"linkedGRNDocumentNumber":     "GRN-APPROVED-TEST",
	})
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	body := decodeResponse(resp)
	require.NotNil(t, body)
	assert.True(t, body["success"].(bool))
}

func TestCreatePVFromPO_PaymentFirst_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupOrgSettingsForExtras(t)

	po := models.PurchaseOrder{
		ID: "po-pf-001", OrganizationID: testOrgID, DocumentNumber: "PO-PF-001",
		Status: "APPROVED", TotalAmount: 750.0, Currency: "ZMW", ProcurementFlow: "payment_first",
	}
	po.Items = emptyPOItems()
	po.ActionHistory = emptyActionHistory()
	require.NoError(t, db.Create(&po).Error)

	app := newExtrasTestApp()
	resp := testRequest(app, http.MethodPost, "/api/v1/payment-vouchers/from-po", map[string]interface{}{
		"purchaseOrderId":             "po-pf-001",
		"purchaseOrderDocumentNumber": "PO-PF-001",
		"title":                       "Payment-first PV",
		"totalAmount":                 750.0,
		"currency":                    "ZMW",
	})
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	body := decodeResponse(resp)
	require.NotNil(t, body)
	assert.True(t, body["success"].(bool))
}

func TestCreatePVFromPO_GoodsFirst_GRNBelongsToDifferentPO(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupOrgSettingsForExtras(t)

	po := models.PurchaseOrder{
		ID: "po-gf-004", OrganizationID: testOrgID, DocumentNumber: "PO-GF-004",
		Status: "APPROVED", TotalAmount: 200.0, Currency: "ZMW", ProcurementFlow: "goods_first",
	}
	po.Items = emptyPOItems()
	po.ActionHistory = emptyActionHistory()
	require.NoError(t, db.Create(&po).Error)

	// GRN belongs to a different PO
	grn := models.GoodsReceivedNote{
		ID: "grn-wrong-po", OrganizationID: testOrgID, DocumentNumber: "GRN-WRONG-PO",
		PODocumentNumber: "PO-OTHER-999", Status: "APPROVED",
		ReceivedDate: time.Now(), ReceivedBy: testUserID,
	}
	grn.Items = emptyGRNItems()
	grn.ActionHistory = emptyActionHistory()
	require.NoError(t, db.Create(&grn).Error)

	app := newExtrasTestApp()
	resp := testRequest(app, http.MethodPost, "/api/v1/payment-vouchers/from-po", map[string]interface{}{
		"purchaseOrderId":             "po-gf-004",
		"purchaseOrderDocumentNumber": "PO-GF-004",
		"totalAmount":                 200.0,
		"linkedGRNDocumentNumber":     "GRN-WRONG-PO",
	})
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// ─────────────────────────────────────────────────────────────────────────────
// Stats — seeded data paths (empty-DB paths are in per-entity test files)
// ─────────────────────────────────────────────────────────────────────────────

func TestGetRequisitionStats_WithData(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	// Force single connection so seeded data is visible to handler queries.
	if sqlDB, err := db.DB(); err == nil {
		sqlDB.SetMaxOpenConns(1)
	}

	for i, status := range []string{"DRAFT", "APPROVED", "APPROVED", "REJECTED"} {
		r := models.Requisition{
			ID: "req-stat-x-" + string(rune('a'+i)), OrganizationID: testOrgID,
			DocumentNumber: "REQ-STAT-X-" + string(rune('A'+i)),
			Status: status, TotalAmount: 100.0, Currency: "ZMW",
		}
		r.Items = emptyReqItems()
		r.ApprovalHistory = emptyApprovalHistory()
		r.ActionHistory = emptyActionHistory()
		require.NoError(t, db.Create(&r).Error)
	}

	app := newExtrasTestApp()
	resp := testRequest(app, http.MethodGet, "/api/v1/requisitions/stats", nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeResponse(resp)
	require.NotNil(t, body)
	assert.True(t, body["success"].(bool))
	data := body["data"].(map[string]interface{})
	assert.GreaterOrEqual(t, data["total"].(float64), float64(4))
	assert.GreaterOrEqual(t, data["approved"].(float64), float64(2))
	assert.GreaterOrEqual(t, data["rejected"].(float64), float64(1))
}

func TestGetPurchaseOrderStats_WithData(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	if sqlDB, err := db.DB(); err == nil {
		sqlDB.SetMaxOpenConns(1)
	}

	for i, status := range []string{"DRAFT", "APPROVED", "REJECTED", "FULFILLED"} {
		po := models.PurchaseOrder{
			ID: "po-stat-x-" + string(rune('a'+i)), OrganizationID: testOrgID,
			DocumentNumber: "PO-STAT-X-" + string(rune('A'+i)),
			Status: status, TotalAmount: 500.0, Currency: "ZMW",
		}
		po.Items = emptyPOItems()
		po.ActionHistory = emptyActionHistory()
		require.NoError(t, db.Create(&po).Error)
	}

	app := newExtrasTestApp()
	resp := testRequest(app, http.MethodGet, "/api/v1/purchase-orders/stats", nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeResponse(resp)
	data := body["data"].(map[string]interface{})
	assert.GreaterOrEqual(t, data["total"].(float64), float64(4))
	assert.GreaterOrEqual(t, data["approved"].(float64), float64(1))
	assert.GreaterOrEqual(t, data["rejected"].(float64), float64(1))
	assert.Contains(t, data, "fulfilled")
}

func TestGetPaymentVoucherStats_WithData(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	if sqlDB, err := db.DB(); err == nil {
		sqlDB.SetMaxOpenConns(1)
	}

	for i, status := range []string{"DRAFT", "APPROVED", "PAID", "CANCELLED"} {
		pv := models.PaymentVoucher{
			ID: "pv-stat-x-" + string(rune('a'+i)), OrganizationID: testOrgID,
			DocumentNumber: "PV-STAT-X-" + string(rune('A'+i)),
			Status: status, Amount: 250.0, Currency: "ZMW", PaymentMethod: "bank_transfer",
		}
		pv.ActionHistory = emptyActionHistory()
		require.NoError(t, db.Create(&pv).Error)
	}

	app := newExtrasTestApp()
	resp := testRequest(app, http.MethodGet, "/api/v1/payment-vouchers/stats", nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeResponse(resp)
	data := body["data"].(map[string]interface{})
	assert.GreaterOrEqual(t, data["total"].(float64), float64(4))
	assert.GreaterOrEqual(t, data["paid"].(float64), float64(1))
	assert.GreaterOrEqual(t, data["approved"].(float64), float64(1))
	assert.GreaterOrEqual(t, data["cancelled"].(float64), float64(1))
}

// ConfirmGRN tests removed alongside the endpoint. Workflow approval now
// auto-cascades APPROVED → COMPLETED (see workflow_execution_service.go),
// and MarkGRNComplete covers the skip-workflow path.

// ─────────────────────────────────────────────────────────────────────────────
// GetSystemStatistics (reports.go)
// ─────────────────────────────────────────────────────────────────────────────

func TestGetSystemStatistics_ForbiddenNonAdmin(t *testing.T) {
	app := newReportsAppViaHandler("requester")
	resp := testRequest(app, http.MethodGet, "/reports/system-stats", nil)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestGetSystemStatistics_ForbiddenFinance(t *testing.T) {
	app := newReportsAppViaHandler("finance")
	resp := testRequest(app, http.MethodGet, "/reports/system-stats", nil)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestGetSystemStatistics_AdminSuccess(t *testing.T) {
	stub := defaultStub()
	app := newReportsAppWithStub("admin", stub)
	resp := testRequest(app, http.MethodGet, "/reports/system-stats", nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetSystemStatistics_SuperadminSuccess(t *testing.T) {
	stub := defaultStub()
	app := newReportsAppWithStub("superadmin", stub)
	resp := testRequest(app, http.MethodGet, "/reports/system-stats", nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetSystemStatistics_WithDateRange(t *testing.T) {
	stub := defaultStub()
	app := newReportsAppWithStub("admin", stub)
	resp := testRequest(app, http.MethodGet, "/reports/system-stats?start_date=2024-01-01&end_date=2024-12-31", nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

// ─────────────────────────────────────────────────────────────────────────────
// GetApprovalMetrics (reports.go)
// ─────────────────────────────────────────────────────────────────────────────

func TestGetApprovalMetrics_ForbiddenApprover(t *testing.T) {
	app := newReportsAppViaHandler("approver")
	resp := testRequest(app, http.MethodGet, "/reports/approval-metrics", nil)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestGetApprovalMetrics_AdminSuccess(t *testing.T) {
	stub := defaultStub()
	app := newReportsAppWithStub("admin", stub)
	resp := testRequest(app, http.MethodGet, "/reports/approval-metrics", nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetApprovalMetrics_SuperadminSuccess(t *testing.T) {
	stub := defaultStub()
	app := newReportsAppWithStub("superadmin", stub)
	resp := testRequest(app, http.MethodGet, "/reports/approval-metrics", nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

// ─────────────────────────────────────────────────────────────────────────────
// GetUserActivityMetrics (reports.go)
// ─────────────────────────────────────────────────────────────────────────────

func TestGetUserActivityMetrics_ForbiddenRequester(t *testing.T) {
	app := newReportsAppViaHandler("requester")
	resp := testRequest(app, http.MethodGet, "/reports/user-activity", nil)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestGetUserActivityMetrics_AdminSuccess(t *testing.T) {
	stub := defaultStub()
	app := newReportsAppWithStub("admin", stub)
	resp := testRequest(app, http.MethodGet, "/reports/user-activity", nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetUserActivityMetrics_SuperadminSuccess(t *testing.T) {
	stub := defaultStub()
	app := newReportsAppWithStub("superadmin", stub)
	resp := testRequest(app, http.MethodGet, "/reports/user-activity", nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

// ─────────────────────────────────────────────────────────────────────────────
// GetAnalyticsDashboard (reports.go)
// ─────────────────────────────────────────────────────────────────────────────

func TestGetAnalyticsDashboard_ForbiddenFinance(t *testing.T) {
	app := newReportsAppViaHandler("finance")
	resp := testRequest(app, http.MethodGet, "/reports/analytics", nil)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestGetAnalyticsDashboard_AdminSuccess(t *testing.T) {
	stub := defaultStub()
	app := newReportsAppWithStub("admin", stub)
	resp := testRequest(app, http.MethodGet, "/reports/analytics", nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetAnalyticsDashboard_SuperadminSuccess(t *testing.T) {
	stub := defaultStub()
	app := newReportsAppWithStub("superadmin", stub)
	resp := testRequest(app, http.MethodGet, "/reports/analytics", nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetAnalyticsDashboard_WithDateRange(t *testing.T) {
	stub := defaultStub()
	app := newReportsAppWithStub("admin", stub)
	resp := testRequest(app, http.MethodGet, "/reports/analytics?start_date=2024-01-01&end_date=2024-12-31", nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

// ─────────────────────────────────────────────────────────────────────────────
// GetDashboardReports (reports.go)
// ─────────────────────────────────────────────────────────────────────────────

func TestGetDashboardReports_AsAdmin(t *testing.T) {
	stub := defaultStub()
	app := newReportsAppWithStub("admin", stub)
	resp := testRequest(app, http.MethodGet, "/reports/dashboard", nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeResponse(resp)
	require.NotNil(t, body)
	assert.True(t, body["success"].(bool))
	data, ok := body["data"].(map[string]interface{})
	require.True(t, ok, "data should be a JSON object")
	assert.Contains(t, data, "totalDocuments")
	assert.Contains(t, data, "recentActivity")
	assert.Equal(t, testOrgID, data["organizationId"])
	assert.Equal(t, "admin", data["userRole"])
}

func TestGetDashboardReports_AsManager(t *testing.T) {
	stub := defaultStub()
	app := newReportsAppWithStub("manager", stub)
	resp := testRequest(app, http.MethodGet, "/reports/dashboard", nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body := decodeResponse(resp)
	require.NotNil(t, body)
	assert.True(t, body["success"].(bool))
}

func TestGetDashboardReports_AsRequester(t *testing.T) {
	stub := defaultStub()
	app := newReportsAppWithStub("requester", stub)
	resp := testRequest(app, http.MethodGet, "/reports/dashboard", nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetDashboardReports_AsApprover(t *testing.T) {
	stub := defaultStub()
	app := newReportsAppWithStub("approver", stub)
	resp := testRequest(app, http.MethodGet, "/reports/dashboard", nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetDashboardReports_WithDateRange(t *testing.T) {
	stub := defaultStub()
	app := newReportsAppWithStub("admin", stub)
	resp := testRequest(app, http.MethodGet, "/reports/dashboard?start_date=2024-01-01&end_date=2024-12-31", nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetDashboardReports_RecentActivityIsArray(t *testing.T) {
	stub := defaultStub()
	app := newReportsAppWithStub("admin", stub)
	resp := testRequest(app, http.MethodGet, "/reports/dashboard", nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeResponse(resp)
	data := body["data"].(map[string]interface{})
	activity, ok := data["recentActivity"].([]interface{})
	require.True(t, ok, "recentActivity should be a JSON array")
	assert.Len(t, activity, 0)
}
