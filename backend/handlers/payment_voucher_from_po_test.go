package handlers

import (
	"net/http"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/types"
	"gorm.io/datatypes"
)

func fromPOApp() *fiber.App {
	app := fiber.New()
	auth := withTenantCtx(testOrgID, testUserID, testUserRole)
	app.Post("/payment-vouchers/from-po", auth, CreatePaymentVoucherFromPO)
	return app
}

func TestCreatePVFromPO_RejectsNonApprovedPO(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	po := seedPO(t, "PO-FP-1", "DRAFT", "payment_first", 1000)

	body := map[string]interface{}{
		"purchaseOrderId":             po.ID,
		"purchaseOrderDocumentNumber": po.DocumentNumber,
		"totalAmount":                 500,
	}
	resp := testRequest(fromPOApp(), http.MethodPost, "/payment-vouchers/from-po", body)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for non-approved PO, got %d", resp.StatusCode)
	}
}

// B2: the old "one live PV per PO" 409 is gone. A second PV against the same
// PO now succeeds as long as it fits the remaining balance (here: PO total
// 1000, an existing live PV of 500 leaves 500 remaining, and the new request
// is exactly 500). Requesting more than the remaining balance still fails —
// see TestCreatePVFromPO_RejectsAmountOverRemainingBalance.
func TestCreatePVFromPO_AllowsSecondPVWithinRemainingBalance(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	po := seedPO(t, "PO-FP-2", "APPROVED", "payment_first", 1000)
	dup := models.PaymentVoucher{
		ID: uuid.New().String(), OrganizationID: testOrgID, DocumentNumber: "PV-FP-2",
		LinkedPO: po.DocumentNumber, Status: "APPROVED", Amount: 500,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	if err := db.Create(&dup).Error; err != nil {
		t.Fatalf("seed dup PV: %v", err)
	}

	body := map[string]interface{}{
		"purchaseOrderId":             po.ID,
		"purchaseOrderDocumentNumber": po.DocumentNumber,
		"totalAmount":                 500,
	}
	resp := testRequest(fromPOApp(), http.MethodPost, "/payment-vouchers/from-po", body)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 for second PV within remaining balance, got %d", resp.StatusCode)
	}
}

func TestCreatePVFromPO_RejectsAmountOverRemainingBalance(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	po := seedPO(t, "PO-FP-2B", "APPROVED", "payment_first", 1000)
	live := models.PaymentVoucher{
		ID: uuid.New().String(), OrganizationID: testOrgID, DocumentNumber: "PV-FP-2B",
		LinkedPO: po.DocumentNumber, Status: "APPROVED", Amount: 600,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	if err := db.Create(&live).Error; err != nil {
		t.Fatalf("seed live PV: %v", err)
	}

	// Remaining balance is 400; requesting 500 must be rejected.
	body := map[string]interface{}{
		"purchaseOrderId":             po.ID,
		"purchaseOrderDocumentNumber": po.DocumentNumber,
		"totalAmount":                 500,
	}
	resp := testRequest(fromPOApp(), http.MethodPost, "/payment-vouchers/from-po", body)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for amount over remaining balance, got %d", resp.StatusCode)
	}
}

func TestCreatePVFromPO_RejectsAmountOverPOTotal(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	po := seedPO(t, "PO-FP-3", "APPROVED", "payment_first", 1000)

	body := map[string]interface{}{
		"purchaseOrderId":             po.ID,
		"purchaseOrderDocumentNumber": po.DocumentNumber,
		"totalAmount":                 2000,
	}
	resp := testRequest(fromPOApp(), http.MethodPost, "/payment-vouchers/from-po", body)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for over-total amount, got %d", resp.StatusCode)
	}
}

// The key regression: goods-first GRN in COMPLETED (its normal terminal state)
// must be accepted by from-po, not rejected with "must be approved".
func TestCreatePVFromPO_AcceptsCompletedGRN(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	po := models.PurchaseOrder{
		ID: uuid.New().String(), OrganizationID: testOrgID, DocumentNumber: "PO-FP-4",
		Status: "APPROVED", ProcurementFlow: "", TotalAmount: 1000, Currency: "ZMW",
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	po.Items = datatypes.NewJSONType([]types.POItem{{Description: "Widget A", Quantity: 10, UnitPrice: 100, Amount: 1000}})
	if err := db.Create(&po).Error; err != nil {
		t.Fatalf("seed PO: %v", err)
	}

	grn := models.GoodsReceivedNote{
		ID: uuid.New().String(), OrganizationID: testOrgID, DocumentNumber: "GRN-FP-4",
		PODocumentNumber: po.DocumentNumber, Status: "COMPLETED",
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	grn.Items = datatypes.NewJSONType([]types.GRNItem{{Description: "Widget A", QuantityOrdered: 10, QuantityReceived: 10, Condition: "good"}})
	if err := db.Create(&grn).Error; err != nil {
		t.Fatalf("seed GRN: %v", err)
	}

	body := map[string]interface{}{
		"purchaseOrderId":             po.ID,
		"purchaseOrderDocumentNumber": po.DocumentNumber,
		"linkedGRNDocumentNumber":     grn.DocumentNumber,
		"totalAmount":                 500,
	}
	resp := testRequest(fromPOApp(), http.MethodPost, "/payment-vouchers/from-po", body)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 for COMPLETED GRN, got %d", resp.StatusCode)
	}
}

// Fix 2: a fully-delivered, partially-paid PO parks at FULFILLED, not
// APPROVED. The balance PV must still be creatable via from-po, otherwise a
// FULFILLED PO can never receive its remaining payment.
func TestCreatePVFromPO_AcceptsFulfilledPOForBalancePV(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	po := seedPO(t, "PO-FP-5", "FULFILLED", "payment_first", 1000)
	deposit := models.PaymentVoucher{
		ID: uuid.New().String(), OrganizationID: testOrgID, DocumentNumber: "PV-FP-5",
		LinkedPO: po.DocumentNumber, Status: "PAID", Amount: 500,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	if err := db.Create(&deposit).Error; err != nil {
		t.Fatalf("seed deposit PV: %v", err)
	}

	body := map[string]interface{}{
		"purchaseOrderId":             po.ID,
		"purchaseOrderDocumentNumber": po.DocumentNumber,
		"totalAmount":                 500,
	}
	resp := testRequest(fromPOApp(), http.MethodPost, "/payment-vouchers/from-po", body)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 for balance PV against FULFILLED PO, got %d", resp.StatusCode)
	}
}
