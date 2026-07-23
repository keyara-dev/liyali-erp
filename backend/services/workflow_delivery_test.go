package services

import (
	"testing"

	"github.com/liyali/liyali-gateway/types"
)

func TestResolveDeliveryFromGRNs_MatchesByItemCodeDespiteDescriptionDrift(t *testing.T) {
	poItems := []types.POItem{
		{Description: "Widget A", ItemCode: "SKU-1", Quantity: 10},
		{Description: "Widget B", ItemCode: "SKU-2", Quantity: 5},
	}
	// GRN descriptions drifted from the PO, but ItemCode still matches.
	grnItems := []types.GRNItem{
		{Description: "Widget A (relabelled)", ItemCode: "SKU-1", QuantityReceived: 10},
		{Description: "Widget B (relabelled)", ItemCode: "SKU-2", QuantityReceived: 5},
	}

	updated, allFull, anyReceived := resolveDeliveryFromGRNs(poItems, [][]types.GRNItem{grnItems})
	if !anyReceived {
		t.Fatalf("expected anyReceived=true")
	}
	if !allFull {
		t.Fatalf("expected allFull=true via ItemCode match, got false")
	}
	if updated[0].ReceivedQuantity != 10 || updated[1].ReceivedQuantity != 5 {
		t.Fatalf("received quantities not set correctly: %+v", updated)
	}
}

func TestResolveDeliveryFromGRNs_FallsBackToDescription(t *testing.T) {
	poItems := []types.POItem{{Description: "Widget A", Quantity: 10}} // no ItemCode
	grnItems := []types.GRNItem{{Description: "Widget A", QuantityReceived: 10}}

	_, allFull, anyReceived := resolveDeliveryFromGRNs(poItems, [][]types.GRNItem{grnItems})
	if !allFull || !anyReceived {
		t.Fatalf("expected fully delivered via description fallback")
	}
}

func TestResolveDeliveryFromGRNs_Partial(t *testing.T) {
	poItems := []types.POItem{{Description: "Widget A", ItemCode: "SKU-1", Quantity: 10}}
	grnItems := []types.GRNItem{{Description: "Widget A", ItemCode: "SKU-1", QuantityReceived: 4}}

	_, allFull, anyReceived := resolveDeliveryFromGRNs(poItems, [][]types.GRNItem{grnItems})
	if allFull {
		t.Fatalf("expected partial (not allFull)")
	}
	if !anyReceived {
		t.Fatalf("expected anyReceived=true")
	}
}

func TestResolveDeliveryFromGRNs_AggregatesAcrossGRNs(t *testing.T) {
	poItems := []types.POItem{{Description: "Widget A", ItemCode: "SKU-1", Quantity: 10}}
	grn1 := []types.GRNItem{{ItemCode: "SKU-1", Description: "Widget A", QuantityReceived: 6}}
	grn2 := []types.GRNItem{{ItemCode: "SKU-1", Description: "Widget A", QuantityReceived: 4}}

	updated, allFull, _ := resolveDeliveryFromGRNs(poItems, [][]types.GRNItem{grn1, grn2})
	if !allFull || updated[0].ReceivedQuantity != 10 {
		t.Fatalf("expected aggregated 10 across GRNs, got %d", updated[0].ReceivedQuantity)
	}
}
