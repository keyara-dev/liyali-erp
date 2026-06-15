package utils

import (
	"github.com/liyali/liyali-gateway/types"
	"gorm.io/gorm"
)

// ResolveUserRefs batch-loads {id,name,email,role} for the given user IDs in a
// single query and returns a map keyed by user ID. Blank/duplicate IDs are
// dropped; IDs with no matching user are simply absent from the map.
//
// When orgID is non-empty the role is resolved to the user's role IN that
// organization (via organization_members), falling back to the global
// users.role when the user has no active membership row. Pass orgID="" to use
// the global role only.
//
// Use it during list/detail response enrichment to avoid N+1 user lookups —
// collect every user-reference ID across the page, call this once, then assign
// the resulting UserRefs onto the response objects.
func ResolveUserRefs(db *gorm.DB, orgID string, ids []string) map[string]types.UserRef {
	out := map[string]types.UserRef{}

	// Dedupe and drop blanks.
	seen := make(map[string]struct{}, len(ids))
	uniq := make([]string, 0, len(ids))
	for _, id := range ids {
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		uniq = append(uniq, id)
	}
	if len(uniq) == 0 {
		return out
	}

	type row struct {
		ID      string
		Name    string
		Email   string
		Role    string
		OrgRole *string
	}

	// Table() bypasses the soft-delete scope on purpose: a user who was later
	// deactivated/deleted should still resolve to their name on historical docs.
	q := db.Table("users u").Where("u.id IN ?", uniq)
	if orgID != "" {
		q = q.Select("u.id, u.name, u.email, u.role, om.role AS org_role").
			Joins("LEFT JOIN organization_members om ON om.user_id = u.id AND om.organization_id = ? AND om.active = ?", orgID, true)
	} else {
		q = q.Select("u.id, u.name, u.email, u.role")
	}

	var rows []row
	if err := q.Scan(&rows).Error; err != nil {
		return out
	}

	for _, r := range rows {
		role := r.Role
		if r.OrgRole != nil && *r.OrgRole != "" {
			role = *r.OrgRole
		}
		out[r.ID] = types.UserRef{ID: r.ID, Name: r.Name, Email: r.Email, Role: role}
	}
	return out
}
