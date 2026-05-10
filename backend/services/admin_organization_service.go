package services

import (
	"database/sql"
	"fmt"
	"time"
)

// AdminOrganizationService handles administrative operations on organizations,
// including subscription tier management.
type AdminOrganizationService struct {
	db *sql.DB
}

// SubscriptionTierChangeResult contains the result of a tier change operation.
type SubscriptionTierChangeResult struct {
	OrganizationID string    `json:"organizationId"`
	OldTier        string    `json:"oldTier"`
	NewTier        string    `json:"newTier"`
	ChangedBy      string    `json:"changedBy"`
	Reason         string    `json:"reason"`
	ChangedAt      time.Time `json:"changedAt"`
}

// validTiers is the set of accepted subscription tier values.
var validTiers = map[string]bool{
	"basic":        true,
	"professional": true,
	"enterprise":   true,
	"unlimited":    true,
}

// ChangeSubscriptionTier changes an organization's subscription tier and records
// an audit trail.  It validates the tier names, rejects no-op changes, and
// wraps everything in a single DB transaction.
func (s *AdminOrganizationService) ChangeSubscriptionTier(
	organizationID, oldTier, newTier, reason, adminUserID, ipAddress string,
) (*SubscriptionTierChangeResult, error) {
	if !validTiers[newTier] {
		return nil, fmt.Errorf("invalid tier: %q is not a recognised subscription tier", newTier)
	}
	if oldTier == newTier {
		return nil, fmt.Errorf("same tier: organization is already on %q", newTier)
	}
	if reason == "" {
		return nil, fmt.Errorf("reason is required for tier changes")
	}

	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// Update organization tier.
	res, execErr := tx.Exec("UPDATE organizations SET subscription_tier = ? WHERE id = ?", newTier, organizationID)
	if execErr != nil {
		err = fmt.Errorf("update organization tier: %w", execErr)
		return nil, err
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		err = fmt.Errorf("no rows affected: organization %s not found or concurrent modification", organizationID)
		return nil, err
	}

	// Determine event type.
	eventType := "subscription_upgraded"
	for _, t := range []string{"basic", "professional", "enterprise"} {
		if t == oldTier {
			break
		}
		if t == newTier {
			eventType = "subscription_downgraded"
			break
		}
	}

	now := time.Now()

	// Insert subscription event.
	_, err = tx.Exec(
		"INSERT INTO subscription_events (id, organization_id, event_type, old_tier, new_tier, created_at, changed_by) VALUES (?,?,?,?,?,?,?)",
		fmt.Sprintf("evt-%d", now.UnixNano()), organizationID, eventType, oldTier, newTier, now, adminUserID,
	)
	if err != nil {
		return nil, fmt.Errorf("insert subscription event: %w", err)
	}

	// Insert admin audit log.
	_, err = tx.Exec(
		"INSERT INTO admin_audit_logs (id, organization_id, action, old_value, new_value, created_at, reason, admin_user_id) VALUES (?,?,?,?,?,?,?,?)",
		fmt.Sprintf("audit-%d", now.UnixNano()), organizationID, "tier_change", oldTier, newTier, now, reason, adminUserID,
	)
	if err != nil {
		return nil, fmt.Errorf("insert audit log: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	return &SubscriptionTierChangeResult{
		OrganizationID: organizationID,
		OldTier:        oldTier,
		NewTier:        newTier,
		ChangedBy:      adminUserID,
		Reason:         reason,
		ChangedAt:      now,
	}, nil
}
