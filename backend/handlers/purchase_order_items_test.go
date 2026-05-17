package handlers

import (
	"testing"

	"github.com/liyali/liyali-gateway/types"
)

// makeItem is a convenience constructor for test POItems.
func makeItem(id, description string, quantity int, unitPrice, totalPrice float64) types.POItem {
	return types.POItem{
		ID:          id,
		Description: description,
		Quantity:    quantity,
		UnitPrice:   unitPrice,
		TotalPrice:  totalPrice,
		Amount:      totalPrice,
	}
}

// TestBuildItemChanges_Identical verifies that identical old/new arrays produce an empty result.
func TestBuildItemChanges_Identical(t *testing.T) {
	items := []types.POItem{
		makeItem("item-1", "Widget A", 2, 10.00, 20.00),
		makeItem("item-2", "Widget B", 5, 3.50, 17.50),
	}

	changes := buildItemChanges(items, items)

	if len(changes) != 0 {
		t.Errorf("expected 0 changes for identical items, got %d: %v", len(changes), changes)
	}
}

// TestBuildItemChanges_OneUnitPriceChange verifies that a single changed unitPrice
// produces exactly one entry with the correct field, old, and new values.
func TestBuildItemChanges_OneUnitPriceChange(t *testing.T) {
	oldItems := []types.POItem{
		makeItem("item-1", "Widget A", 2, 10.00, 20.00),
	}
	newItems := []types.POItem{
		makeItem("item-1", "Widget A", 2, 12.50, 25.00),
	}

	changes := buildItemChanges(oldItems, newItems)

	// Expect two entries: unitPrice changed and totalPrice changed
	unitPriceFound := false
	for _, ch := range changes {
		if ch["field"] == "unitPrice" {
			unitPriceFound = true
			if ch["itemId"] != "item-1" {
				t.Errorf("expected itemId 'item-1', got %v", ch["itemId"])
			}
			if ch["old"] != 10.00 {
				t.Errorf("expected old unitPrice 10.00, got %v", ch["old"])
			}
			if ch["new"] != 12.50 {
				t.Errorf("expected new unitPrice 12.50, got %v", ch["new"])
			}
		}
	}
	if !unitPriceFound {
		t.Errorf("expected a 'unitPrice' change entry, got: %v", changes)
	}
}

// TestBuildItemChanges_MultipleChangedItems verifies that changes across multiple items
// each produce their own entries.
func TestBuildItemChanges_MultipleChangedItems(t *testing.T) {
	oldItems := []types.POItem{
		makeItem("item-1", "Widget A", 2, 10.00, 20.00),
		makeItem("item-2", "Widget B", 5, 3.50, 17.50),
		makeItem("item-3", "Widget C", 1, 100.00, 100.00),
	}
	newItems := []types.POItem{
		makeItem("item-1", "Widget A", 3, 10.00, 30.00), // quantity changed
		makeItem("item-2", "Widget B Updated", 5, 4.00, 20.00), // description + unitPrice changed
		makeItem("item-3", "Widget C", 1, 100.00, 100.00), // unchanged
	}

	changes := buildItemChanges(oldItems, newItems)

	// Collect changed fields by itemId
	fieldsByItem := map[string][]string{}
	for _, ch := range changes {
		id := ch["itemId"].(string)
		field := ch["field"].(string)
		fieldsByItem[id] = append(fieldsByItem[id], field)
	}

	// item-1: quantity and totalPrice changed
	if !containsField(fieldsByItem["item-1"], "quantity") {
		t.Errorf("expected 'quantity' change for item-1, got fields: %v", fieldsByItem["item-1"])
	}

	// item-2: description and unitPrice changed
	if !containsField(fieldsByItem["item-2"], "description") {
		t.Errorf("expected 'description' change for item-2, got fields: %v", fieldsByItem["item-2"])
	}
	if !containsField(fieldsByItem["item-2"], "unitPrice") {
		t.Errorf("expected 'unitPrice' change for item-2, got fields: %v", fieldsByItem["item-2"])
	}

	// item-3: no changes
	if len(fieldsByItem["item-3"]) != 0 {
		t.Errorf("expected no changes for item-3, got fields: %v", fieldsByItem["item-3"])
	}
}

// TestBuildItemChanges_EmptySlices verifies that two empty slices produce an empty result.
func TestBuildItemChanges_EmptySlices(t *testing.T) {
	changes := buildItemChanges([]types.POItem{}, []types.POItem{})
	if len(changes) != 0 {
		t.Errorf("expected 0 changes for empty slices, got %d", len(changes))
	}
}

// TestBuildItemChanges_FallbackItemID verifies that items without an ID use the index as itemId.
func TestBuildItemChanges_FallbackItemID(t *testing.T) {
	oldItems := []types.POItem{
		{Description: "No ID Item", Quantity: 1, UnitPrice: 5.00, Amount: 5.00},
	}
	newItems := []types.POItem{
		{Description: "No ID Item", Quantity: 1, UnitPrice: 9.99, Amount: 9.99},
	}

	changes := buildItemChanges(oldItems, newItems)

	if len(changes) == 0 {
		t.Fatal("expected at least one change entry")
	}
	for _, ch := range changes {
		if ch["itemId"] != "0" {
			t.Errorf("expected fallback itemId '0', got %v", ch["itemId"])
		}
	}
}

// containsField is a helper to check if a string slice contains a value.
func containsField(fields []string, target string) bool {
	for _, f := range fields {
		if f == target {
			return true
		}
	}
	return false
}
