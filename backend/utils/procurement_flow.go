package utils

import "strings"

// ResolveProcurementFlow returns the normalized effective procurement flow with
// precedence: PO override → org default → "goods_first".
//
// Both inputs are normalized with ToLower+TrimSpace so callers that read the
// raw stored values (which may carry stray case/whitespace) classify the flow
// consistently. Centralizing this avoids the drift where some call sites
// normalized and others compared the raw string.
func ResolveProcurementFlow(poOverride, orgDefault string) string {
	if v := strings.ToLower(strings.TrimSpace(poOverride)); v != "" {
		return v
	}
	if v := strings.ToLower(strings.TrimSpace(orgDefault)); v != "" {
		return v
	}
	return "goods_first"
}
