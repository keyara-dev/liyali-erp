package utils

import "testing"

func TestResolveProcurementFlow(t *testing.T) {
	cases := []struct {
		name       string
		poOverride string
		orgDefault string
		want       string
	}{
		{"po override wins, normalized", "  Payment_First ", "goods_first", "payment_first"},
		{"falls back to org default, normalized", "", " Goods_First ", "goods_first"},
		{"org default payment_first", "", "payment_first", "payment_first"},
		{"final default when both empty", "", "", "goods_first"},
		{"whitespace-only override falls through", "   ", "payment_first", "payment_first"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := ResolveProcurementFlow(tc.poOverride, tc.orgDefault); got != tc.want {
				t.Fatalf("ResolveProcurementFlow(%q,%q)=%q, want %q", tc.poOverride, tc.orgDefault, got, tc.want)
			}
		})
	}
}
