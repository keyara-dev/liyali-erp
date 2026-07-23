package services

import "testing"

// Custom org roles are assigned permissions via the UI in "resource:action"
// (colon) form, but HasPermission looks them up as "resource.action" (dot).
// canonicalPermission bridges that gap so granted custom-role permissions
// actually take effect at the route guards.
func TestCanonicalPermission(t *testing.T) {
	cases := map[string]string{
		"payment_voucher:pay":     "payment_voucher.pay",
		"payment_voucher.approve": "payment_voucher.approve", // already canonical
		"purchase_order:edit":     "purchase_order.edit",
		"grn:approve":             "grn.approve",
		"":                        "",
	}
	for in, want := range cases {
		if got := canonicalPermission(in); got != want {
			t.Errorf("canonicalPermission(%q)=%q, want %q", in, got, want)
		}
	}
}
