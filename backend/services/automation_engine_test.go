package services

import "testing"

func TestResolveAutomationAction(t *testing.T) {
	cases := []struct {
		name      string
		level     string
		amount    float64
		maxAmount float64
		want      string
	}{
		{"manual leaves draft", "manual", 100, 1000, "draft"},
		{"empty leaves draft", "", 100, 1000, "draft"},
		{"unknown leaves draft", "weird", 100, 1000, "draft"},
		{"auto_submit submits", "auto_submit", 100, 1000, "submit"},
		{"auto_approve under cap approves", "auto_approve", 500, 1000, "approve"},
		{"auto_approve at cap approves", "auto_approve", 1000, 1000, "approve"},
		{"auto_approve over cap falls back to submit", "auto_approve", 5000, 1000, "submit"},
		{"auto_approve with zero cap never approves", "auto_approve", 100, 0, "submit"},
		{"level is case/space insensitive", "  Auto_Submit ", 100, 1000, "submit"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := resolveAutomationAction(tc.level, tc.amount, tc.maxAmount); got != tc.want {
				t.Fatalf("resolveAutomationAction(%q,%v,%v)=%q, want %q", tc.level, tc.amount, tc.maxAmount, got, tc.want)
			}
		})
	}
}
