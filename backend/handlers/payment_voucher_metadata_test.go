package handlers

// payment_voucher_metadata_test.go — Task B4: PV creation records paymentType
// ("full" | "partial", inferred from amount vs PO total when not sent) and a
// free-text narration into PaymentVoucher.Metadata, surfaced on the response.
// Drives the full from-PO HTTP handler; seedPO / fromPOApp / testRequest are
// shared helpers from the payment-voucher test suite.

import (
	"encoding/json"
	"net/http"
	"testing"
)

// pvResponseMetadata posts a from-PO create and returns the response's
// data.metadata object (nil if absent).
func pvResponseMetadata(t *testing.T, resp *http.Response) map[string]interface{} {
	t.Helper()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	data, _ := result["data"].(map[string]interface{})
	if data == nil {
		t.Fatal("response has no data field")
	}
	meta, _ := data["metadata"].(map[string]interface{})
	return meta
}

// TestCreatePVFromPO_InfersPartialAndStoresNarration: a PV for less than the
// PO total, with no explicit paymentType, is inferred "partial" and its
// narration + PO-total snapshot are persisted to metadata.
func TestCreatePVFromPO_InfersPartialAndStoresNarration(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	po := seedPO(t, "PO-META-1", "APPROVED", "payment_first", 1000)
	resp := testRequest(fromPOApp(), http.MethodPost, "/payment-vouchers/from-po", map[string]interface{}{
		"purchaseOrderId":             po.ID,
		"purchaseOrderDocumentNumber": po.DocumentNumber,
		"totalAmount":                 400.0,
		"narration":                   "Initial deposit — 40%",
	})

	meta := pvResponseMetadata(t, resp)
	if meta == nil {
		t.Fatal("expected metadata on response, got none")
	}
	if pt, _ := meta["paymentType"].(string); pt != "partial" {
		t.Fatalf("paymentType: want partial, got %q", pt)
	}
	if n, _ := meta["narration"].(string); n != "Initial deposit — 40%" {
		t.Fatalf("narration not persisted, got %q", n)
	}
	if total, _ := meta["poTotalAtCreation"].(float64); total != 1000 {
		t.Fatalf("poTotalAtCreation: want 1000, got %v", total)
	}
	if committed, ok := meta["committedBefore"].(float64); !ok || committed != 0 {
		t.Fatalf("committedBefore: want 0, got %v", meta["committedBefore"])
	}
}

// TestCreatePVFromPO_InfersFullWhenAmountEqualsTotal: a PV for the whole PO
// total, no explicit paymentType, is inferred "full".
func TestCreatePVFromPO_InfersFullWhenAmountEqualsTotal(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	po := seedPO(t, "PO-META-2", "APPROVED", "payment_first", 1000)
	resp := testRequest(fromPOApp(), http.MethodPost, "/payment-vouchers/from-po", map[string]interface{}{
		"purchaseOrderId":             po.ID,
		"purchaseOrderDocumentNumber": po.DocumentNumber,
		"totalAmount":                 1000.0,
	})

	meta := pvResponseMetadata(t, resp)
	if meta == nil {
		t.Fatal("expected metadata on response, got none")
	}
	if pt, _ := meta["paymentType"].(string); pt != "full" {
		t.Fatalf("paymentType: want full, got %q", pt)
	}
	if _, present := meta["narration"]; present {
		t.Fatalf("narration should be omitted when empty, got %v", meta["narration"])
	}
}

// TestCreatePVFromPO_ExplicitPaymentTypeWins: an explicit paymentType is
// honored even when the amount would infer the opposite.
func TestCreatePVFromPO_ExplicitPaymentTypeWins(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	po := seedPO(t, "PO-META-3", "APPROVED", "payment_first", 1000)
	resp := testRequest(fromPOApp(), http.MethodPost, "/payment-vouchers/from-po", map[string]interface{}{
		"purchaseOrderId":             po.ID,
		"purchaseOrderDocumentNumber": po.DocumentNumber,
		"totalAmount":                 1000.0, // would infer "full"
		"paymentType":                 "partial",
	})

	meta := pvResponseMetadata(t, resp)
	if meta == nil {
		t.Fatal("expected metadata on response, got none")
	}
	if pt, _ := meta["paymentType"].(string); pt != "partial" {
		t.Fatalf("explicit paymentType not honored: want partial, got %q", pt)
	}
}
