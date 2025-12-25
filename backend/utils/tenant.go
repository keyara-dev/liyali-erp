package utils

import (
	"gorm.io/gorm"
	"github.com/liyali/liyali-gateway/middleware"
)

// WithTenant adds automatic organization_id filtering to GORM queries
// This ensures all queries are scoped to the current organization
func WithTenant(db *gorm.DB, tenant *middleware.TenantContext) *gorm.DB {
	if tenant == nil {
		// If no tenant context, return db as-is (for system operations)
		// This should rarely happen in production
		return db
	}
	return db.Where("organization_id = ?", tenant.OrganizationID)
}

// WithTenantID adds organization_id filtering using org ID directly
// Useful when you have the org ID but not the full tenant context
func WithTenantID(db *gorm.DB, orgID string) *gorm.DB {
	if orgID == "" {
		return db
	}
	return db.Where("organization_id = ?", orgID)
}

// WithoutTenant returns the query without tenant filtering
// Use this only for admin operations that need to see all organizations
// Should be used sparingly and with proper authorization checks
func WithoutTenant(db *gorm.DB) *gorm.DB {
	return db
}
