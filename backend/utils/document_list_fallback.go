package utils

import (
	"gorm.io/gorm"
)

// DocumentListFilters bundles the common WHERE-clause inputs the list
// handlers (REQ/PO/PV/GRN) accept.
type DocumentListFilters struct {
	OrganizationID    string
	Status            string
	// Optional secondary filter — e.g. po_document_number on GRN or vendor_id on PV.
	RefField          string
	RefValue          string
	// When true, exclude rows where metadata->>'directPayment' = 'true'.
	HideDirectPayment bool
	// When true, restrict to procurement-routed documents (workflow conditions).
	ProcurementOnly   bool
}

// ListDocumentIDsFallback is the gorm-based fallback for the GRN/PO/PV/REQ
// list handlers. Used when the sqlc Queries handle is not initialised — e.g.
// inside the SQLite-backed unit test harness.
//
// The fallback intentionally ignores ProcurementOnly because that path joins
// against workflow_assignments → workflows JSON conditions which is awkward
// to mirror in sqlite; tests that need it should run against Postgres.
func ListDocumentIDsFallback(db *gorm.DB, table string, f DocumentListFilters, scope DocumentScope, limit, offset int) (int64, []string, error) {
	q := db.Table(table).Where("organization_id = ?", f.OrganizationID)

	if f.Status != "" {
		q = q.Where("UPPER(status) = UPPER(?)", f.Status)
	}
	if f.RefField != "" && f.RefValue != "" {
		q = q.Where(f.RefField+" = ?", f.RefValue)
	}
	if f.HideDirectPayment {
		q = q.Where("COALESCE(metadata ->> 'directPayment', '') <> 'true'")
	}
	if !(scope.CanViewAll || scope.IsProcurement) {
		// Limited scope: creator only. GRN also widens via received_by but
		// that is handled inline in grn.go's bespoke helper.
		q = q.Where("created_by = ?", scope.UserID)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return 0, nil, err
	}

	var ids []string
	if err := q.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Pluck("id", &ids).Error; err != nil {
		return 0, nil, err
	}
	return total, ids, nil
}
