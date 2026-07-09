package handlers

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/models"
	"gorm.io/gorm"
)

// ─────────────────────────────────────────────────────────────────────────
// Task A2: PO creation warn-only compliance flag + vendor metadata snapshot
//
// Both PO-creation paths (CreatePurchaseOrder and
// CreatePurchaseOrderFromRequisition) must:
//   - never reject creation because a vendor is missing ZRA TPIN / PACRA
//     registration data (warn-only)
//   - snapshot the vendor's compliance fields onto the PO's metadata for
//     audit purposes
//   - surface live complianceWarnings (computed from the current vendor
//     record, not the snapshot) on the response
// ─────────────────────────────────────────────────────────────────────────

// createComplianceTestVendor inserts a vendor for testOrgID with the given
// ZRA TPIN / PACRA registration values. An empty string simulates a vendor
// that hasn't supplied that compliance field yet.
func createComplianceTestVendor(t *testing.T, db *gorm.DB, zraTpin, pacraRegNumber string) models.Vendor {
	t.Helper()
	v := models.Vendor{
		ID:             uuid.New().String(),
		OrganizationID: testOrgID,
		VendorCode:     "VND-CMP-" + uuid.New().String()[:8],
		Name:           "Compliance Test Vendor",
		Email:          uuid.New().String() + "@example.com",
		Country:        "Zambia",
		City:           "Lusaka",
		TaxID:          zraTpin,
		ZraTpin:        zraTpin,
		PacraRegNumber: pacraRegNumber,
		Active:         true,
		CreatedBy:      testUserID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	if err := db.Create(&v).Error; err != nil {
		t.Fatalf("failed to create test vendor: %v", err)
	}
	return v
}

// asStringSlice converts a decoded-JSON []interface{} (or nil/absent value)
// into []string for easier assertions.
func asStringSlice(v interface{}) []string {
	arr, ok := v.([]interface{})
	if !ok {
		return nil
	}
	out := make([]string, 0, len(arr))
	for _, e := range arr {
		if s, ok := e.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

// ─────────────────────────────────────────────────────────────────────────
// CreatePurchaseOrder
// ─────────────────────────────────────────────────────────────────────────

func TestCreatePurchaseOrder_VendorMissingPacra_WarnsButCreates(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	vendor := createComplianceTestVendor(t, db, "1000000000", "") // missing PACRA

	app := newPurchaseOrderApp(t)
	body := map[string]interface{}{
		"vendorId":    vendor.ID,
		"vendorName":  vendor.Name,
		"totalAmount": 1000,
		"currency":    "ZMW",
		"items": []map[string]interface{}{
			{"description": "Widget A", "quantity": 10, "unitPrice": 100, "amount": 1000},
		},
	}

	resp := testRequest(app, http.MethodPost, "/purchase-orders", body)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 (warn-only, never blocks creation), got %d", resp.StatusCode)
	}

	respBody := decodeResponse(resp)
	data, ok := respBody["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected data object in response, got %#v", respBody)
	}

	warnings := asStringSlice(data["complianceWarnings"])
	if len(warnings) == 0 {
		t.Fatalf("expected non-empty complianceWarnings for vendor missing PACRA, got %#v", data["complianceWarnings"])
	}

	metadata, ok := data["metadata"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected metadata object in response, got %#v", data["metadata"])
	}
	snapshot, ok := metadata["vendorCompliance"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected vendorCompliance snapshot in metadata, got %#v", metadata)
	}
	if snapshot["zraTpin"] != "1000000000" {
		t.Errorf("snapshot zraTpin = %v, want 1000000000", snapshot["zraTpin"])
	}
	if snapshot["pacraRegNumber"] != "" {
		t.Errorf("snapshot pacraRegNumber = %v, want empty", snapshot["pacraRegNumber"])
	}
	if s, _ := snapshot["snapshotAt"].(string); s == "" {
		t.Errorf("expected snapshotAt to be set, got %#v", snapshot["snapshotAt"])
	}
	if metaWarnings := asStringSlice(metadata["complianceWarnings"]); len(metaWarnings) == 0 {
		t.Errorf("expected complianceWarnings persisted in metadata snapshot, got %#v", metadata["complianceWarnings"])
	}

	// Verify the metadata snapshot is actually persisted in the DB, not just
	// echoed back in the HTTP response.
	orderID, _ := data["id"].(string)
	var order models.PurchaseOrder
	if err := db.Where("id = ?", orderID).First(&order).Error; err != nil {
		t.Fatalf("failed to reload created PO: %v", err)
	}
	var persistedMeta map[string]interface{}
	if err := json.Unmarshal(order.Metadata, &persistedMeta); err != nil {
		t.Fatalf("failed to unmarshal persisted metadata: %v", err)
	}
	if _, ok := persistedMeta["vendorCompliance"]; !ok {
		t.Errorf("expected vendorCompliance snapshot persisted in DB metadata, got %#v", persistedMeta)
	}
}

func TestCreatePurchaseOrder_VendorFullyCompliant_NoWarnings(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	vendor := createComplianceTestVendor(t, db, "1000000001", "PACRA-REG-000111") // fully compliant

	app := newPurchaseOrderApp(t)
	body := map[string]interface{}{
		"vendorId":    vendor.ID,
		"vendorName":  vendor.Name,
		"totalAmount": 1000,
		"currency":    "ZMW",
		"items": []map[string]interface{}{
			{"description": "Widget A", "quantity": 10, "unitPrice": 100, "amount": 1000},
		},
	}

	resp := testRequest(app, http.MethodPost, "/purchase-orders", body)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}

	respBody := decodeResponse(resp)
	data, ok := respBody["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected data object in response, got %#v", respBody)
	}

	if warnings := asStringSlice(data["complianceWarnings"]); len(warnings) != 0 {
		t.Errorf("expected no complianceWarnings for fully compliant vendor, got %v", warnings)
	}

	metadata, ok := data["metadata"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected metadata object in response, got %#v", data["metadata"])
	}
	if _, present := metadata["complianceWarnings"]; present {
		t.Errorf("expected no complianceWarnings key in metadata for compliant vendor, got %#v", metadata["complianceWarnings"])
	}
	snapshot, ok := metadata["vendorCompliance"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected vendorCompliance snapshot in metadata, got %#v", metadata)
	}
	if snapshot["pacraRegNumber"] != "PACRA-REG-000111" {
		t.Errorf("snapshot pacraRegNumber = %v, want PACRA-REG-000111", snapshot["pacraRegNumber"])
	}
}

// ─────────────────────────────────────────────────────────────────────────
// CreatePurchaseOrderFromRequisition — same warn-only behavior
// ─────────────────────────────────────────────────────────────────────────

func TestCreatePurchaseOrderFromRequisition_VendorMissingTpin_WarnsButCreates(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	vendor := createComplianceTestVendor(t, db, "", "PACRA-REG-000222") // missing ZRA TPIN

	app := newPurchaseOrderApp(t)
	body := map[string]interface{}{
		"requisitionId": uuid.New().String(),
		"vendorId":      vendor.ID,
		"vendorName":    vendor.Name,
		"totalAmount":   500,
		"currency":      "ZMW",
		"items": []map[string]interface{}{
			{"description": "Widget B", "quantity": 5, "unitPrice": 100, "amount": 500},
		},
	}

	resp := testRequest(app, http.MethodPost, "/purchase-orders/from-requisition", body)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 (warn-only, never blocks creation), got %d", resp.StatusCode)
	}

	respBody := decodeResponse(resp)
	data, ok := respBody["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected data object in response, got %#v", respBody)
	}

	warnings := asStringSlice(data["complianceWarnings"])
	if len(warnings) == 0 {
		t.Fatalf("expected non-empty complianceWarnings for vendor missing ZRA TPIN, got %#v", data["complianceWarnings"])
	}

	metadata, ok := data["metadata"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected metadata object in response, got %#v", data["metadata"])
	}
	snapshot, ok := metadata["vendorCompliance"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected vendorCompliance snapshot in metadata, got %#v", metadata)
	}
	if snapshot["pacraRegNumber"] != "PACRA-REG-000222" {
		t.Errorf("snapshot pacraRegNumber = %v, want PACRA-REG-000222", snapshot["pacraRegNumber"])
	}
	if snapshot["zraTpin"] != "" {
		t.Errorf("snapshot zraTpin = %v, want empty", snapshot["zraTpin"])
	}

	// Verify persisted in DB too (from-requisition path loads the vendor
	// fresh — see document_extras_handler.go — so this also guards against a
	// regression where only the vendor ID, not the record, is at hand).
	orderID, _ := data["id"].(string)
	var order models.PurchaseOrder
	if err := db.Where("id = ?", orderID).First(&order).Error; err != nil {
		t.Fatalf("failed to reload created PO: %v", err)
	}
	var persistedMeta map[string]interface{}
	if err := json.Unmarshal(order.Metadata, &persistedMeta); err != nil {
		t.Fatalf("failed to unmarshal persisted metadata: %v", err)
	}
	if _, ok := persistedMeta["vendorCompliance"]; !ok {
		t.Errorf("expected vendorCompliance snapshot persisted in DB metadata, got %#v", persistedMeta)
	}
}

func TestCreatePurchaseOrderFromRequisition_VendorFullyCompliant_NoWarnings(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	vendor := createComplianceTestVendor(t, db, "1000000002", "PACRA-REG-000333")

	app := newPurchaseOrderApp(t)
	body := map[string]interface{}{
		"requisitionId": uuid.New().String(),
		"vendorId":      vendor.ID,
		"vendorName":    vendor.Name,
		"totalAmount":   500,
		"currency":      "ZMW",
		"items": []map[string]interface{}{
			{"description": "Widget B", "quantity": 5, "unitPrice": 100, "amount": 500},
		},
	}

	resp := testRequest(app, http.MethodPost, "/purchase-orders/from-requisition", body)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}

	respBody := decodeResponse(resp)
	data, ok := respBody["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected data object in response, got %#v", respBody)
	}

	if warnings := asStringSlice(data["complianceWarnings"]); len(warnings) != 0 {
		t.Errorf("expected no complianceWarnings for fully compliant vendor, got %v", warnings)
	}
}

// ─────────────────────────────────────────────────────────────────────────
// No vendor at all — metadata snapshot must not be added, and no warnings.
// ─────────────────────────────────────────────────────────────────────────

func TestCreatePurchaseOrder_NoVendor_NoComplianceSnapshot(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newPurchaseOrderApp(t)
	body := map[string]interface{}{
		"vendorName":  "Walk-in Supplier",
		"totalAmount": 250,
		"currency":    "ZMW",
		"items": []map[string]interface{}{
			{"description": "Misc", "quantity": 1, "unitPrice": 250, "amount": 250},
		},
	}

	resp := testRequest(app, http.MethodPost, "/purchase-orders", body)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}

	respBody := decodeResponse(resp)
	data, ok := respBody["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected data object in response, got %#v", respBody)
	}

	if warnings := asStringSlice(data["complianceWarnings"]); len(warnings) != 0 {
		t.Errorf("expected no complianceWarnings when no vendor is linked, got %v", warnings)
	}
	if metadata, ok := data["metadata"].(map[string]interface{}); ok {
		if _, present := metadata["vendorCompliance"]; present {
			t.Errorf("expected no vendorCompliance snapshot when no vendor is linked, got %#v", metadata["vendorCompliance"])
		}
	}
}
