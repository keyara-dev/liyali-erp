package services

import (
	"github.com/liyali/liyali-gateway/models"
	"gorm.io/gorm"
)

// POPaymentSummary aggregates the payment vouchers linked to a single
// purchase order so callers (balance gate, derived payment status, cascade
// completion) can reason about partial payments without re-deriving the SQL
// themselves.
type POPaymentSummary struct {
	Committed float64 `gorm:"column:committed"` // Σ amount, status NOT IN (CANCELLED, REJECTED)
	Paid      float64 `gorm:"column:paid"`      // Σ amount, status IN (PAID, COMPLETED)
	LivePVs   int64   `gorm:"column:live_pvs"`  // count of rows, status NOT IN (CANCELLED, REJECTED)
}

// ComputePOPaymentSummary aggregates all payment vouchers linked to
// poDocNumber (within orgID) into a POPaymentSummary. It uses a single
// portable GORM Select (COALESCE + CASE WHEN) so the same query runs
// unmodified against Postgres (production) and SQLite (test harness) — no
// sqlc, since sqlc's generated Queries are nil under the SQLite test
// harness.
//
// Committed excludes CANCELLED/REJECTED PVs (those are terminal failures and
// never consume PO budget). Paid only counts PAID/COMPLETED PVs. LivePVs
// counts every PV that is not CANCELLED/REJECTED (mirrors the "live PV"
// definition already used by validateProcurementPVGate).
func ComputePOPaymentSummary(tx *gorm.DB, orgID, poDocNumber string) (POPaymentSummary, error) {
	var summary POPaymentSummary

	err := tx.Model(&models.PaymentVoucher{}).
		Where("linked_po = ? AND organization_id = ?", poDocNumber, orgID).
		Select(
			"COALESCE(SUM(CASE WHEN UPPER(status) NOT IN ('CANCELLED','REJECTED') THEN amount ELSE 0 END), 0) AS committed, " +
				"COALESCE(SUM(CASE WHEN UPPER(status) IN ('PAID','COMPLETED') THEN amount ELSE 0 END), 0) AS paid, " +
				"COALESCE(SUM(CASE WHEN UPPER(status) NOT IN ('CANCELLED','REJECTED') THEN 1 ELSE 0 END), 0) AS live_pvs",
		).
		Scan(&summary).Error

	return summary, err
}

// paymentStatusEpsilon absorbs floating-point rounding noise when comparing
// accumulated payment totals against a PO total (e.g. currency amounts that
// don't divide evenly across multiple PVs).
const paymentStatusEpsilon = 0.01

// DerivePaymentStatus classifies how much of poTotal has been paid so far,
// returning "unpaid", "partially_paid", or "fully_paid". paid is treated as
// effectively zero within paymentStatusEpsilon of 0, and effectively
// complete within paymentStatusEpsilon of poTotal.
func DerivePaymentStatus(poTotal, paid float64) string {
	switch {
	case paid <= paymentStatusEpsilon:
		return "unpaid"
	case paid >= poTotal-paymentStatusEpsilon:
		return "fully_paid"
	default:
		return "partially_paid"
	}
}
