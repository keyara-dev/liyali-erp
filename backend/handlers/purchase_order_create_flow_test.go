package handlers

import (
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/liyali/liyali-gateway/models"
)

func createPOApp() *fiber.App {
	app := fiber.New()
	auth := withTenantCtx(testOrgID, testUserID, testUserRole)
	app.Post("/purchase-orders", auth, CreatePurchaseOrder)
	return app
}

// Regression: the per-PO procurement flow chosen in the creation wizard must be
// persisted on the PO. It used to be dropped (the handler never copied
// req.ProcurementFlow onto the model), so a "goods_first" PO was saved with an
// empty flow and, on approval, inherited a payment_first org default — routing
// the PO to a PV instead of a GRN.
func TestCreatePurchaseOrder_PersistsProcurementFlow(t *testing.T) {
	cases := []struct {
		name string
		flow string
	}{
		{"goods_first override is stored", "goods_first"},
		{"payment_first override is stored", "payment_first"},
		{"empty inherits org default", ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db := setupTestDB(t)
			defer teardownTestDB(t, db)

			body := map[string]interface{}{
				"vendorName":      "Acme Co",
				"totalAmount":     1000,
				"currency":        "ZMW",
				"procurementFlow": tc.flow,
				"items": []map[string]interface{}{
					{"description": "Widget A", "quantity": 10, "unitPrice": 100, "amount": 1000},
				},
			}

			resp := testRequest(createPOApp(), http.MethodPost, "/purchase-orders", body)
			if resp.StatusCode != http.StatusCreated {
				t.Fatalf("expected 201, got %d", resp.StatusCode)
			}

			var po models.PurchaseOrder
			if err := db.Where("organization_id = ?", testOrgID).First(&po).Error; err != nil {
				t.Fatalf("created PO not found: %v", err)
			}
			if po.ProcurementFlow != tc.flow {
				t.Errorf("ProcurementFlow not persisted: want %q, got %q", tc.flow, po.ProcurementFlow)
			}
		})
	}
}
