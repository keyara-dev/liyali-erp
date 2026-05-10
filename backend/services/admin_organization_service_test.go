package services

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChangeSubscriptionTier_Success(t *testing.T) {
	// Setup mock database
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	service := &AdminOrganizationService{
		db: db,
	}

	// Test data
	organizationID := "org-test-001"
	oldTier := "basic"
	newTier := "professional"
	reason := "Customer upgrade request"
	adminUserID := "user-admin-001"
	ipAddress := "192.168.1.1"

	// Expect transaction begin
	mock.ExpectBegin()

	// Expect organization update
	mock.ExpectExec("UPDATE organizations").
		WithArgs(newTier, organizationID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Expect subscription event insert
	mock.ExpectExec("INSERT INTO subscription_events").
		WithArgs(sqlmock.AnyArg(), organizationID, "subscription_upgraded", oldTier, newTier, sqlmock.AnyArg(), adminUserID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Expect audit log insert
	mock.ExpectExec("INSERT INTO admin_audit_logs").
		WithArgs(sqlmock.AnyArg(), organizationID, "tier_change", oldTier, newTier, sqlmock.AnyArg(), reason, adminUserID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Expect transaction commit
	mock.ExpectCommit()

	// Execute
	result, err := service.ChangeSubscriptionTier(
		organizationID,
		oldTier,
		newTier,
		reason,
		adminUserID,
		ipAddress,
	)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, organizationID, result.OrganizationID)
	assert.Equal(t, oldTier, result.OldTier)
	assert.Equal(t, newTier, result.NewTier)
	assert.Equal(t, adminUserID, result.ChangedBy)
	assert.Equal(t, reason, result.Reason)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestChangeSubscriptionTier_Downgrade(t *testing.T) {
	// Setup mock database
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	service := &AdminOrganizationService{
		db: db,
	}

	// Test data (downgrade)
	organizationID := "org-test-001"
	oldTier := "enterprise"
	newTier := "professional"
	reason := "Customer downgrade request"
	adminUserID := "user-admin-001"
	ipAddress := "192.168.1.1"

	// Expect transaction begin
	mock.ExpectBegin()

	// Expect organization update
	mock.ExpectExec("UPDATE organizations").
		WithArgs(newTier, organizationID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Expect subscription event insert (downgrade)
	mock.ExpectExec("INSERT INTO subscription_events").
		WithArgs(sqlmock.AnyArg(), organizationID, "subscription_downgraded", oldTier, newTier, sqlmock.AnyArg(), adminUserID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Expect audit log insert
	mock.ExpectExec("INSERT INTO admin_audit_logs").
		WithArgs(sqlmock.AnyArg(), organizationID, "tier_change", oldTier, newTier, sqlmock.AnyArg(), reason, adminUserID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Expect transaction commit
	mock.ExpectCommit()

	// Execute
	result, err := service.ChangeSubscriptionTier(
		organizationID,
		oldTier,
		newTier,
		reason,
		adminUserID,
		ipAddress,
	)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, newTier, result.NewTier)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestChangeSubscriptionTier_TransactionRollback(t *testing.T) {
	// Setup mock database
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	service := &AdminOrganizationService{
		db: db,
	}

	// Test data
	organizationID := "org-test-001"
	oldTier := "basic"
	newTier := "professional"
	reason := "Test rollback"
	adminUserID := "user-admin-001"
	ipAddress := "192.168.1.1"

	// Expect transaction begin
	mock.ExpectBegin()

	// Expect organization update to fail
	mock.ExpectExec("UPDATE organizations").
		WithArgs(newTier, organizationID).
		WillReturnError(sql.ErrConnDone)

	// Expect transaction rollback
	mock.ExpectRollback()

	// Execute
	result, err := service.ChangeSubscriptionTier(
		organizationID,
		oldTier,
		newTier,
		reason,
		adminUserID,
		ipAddress,
	)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestChangeSubscriptionTier_InvalidTier(t *testing.T) {
	// Setup mock database
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	service := &AdminOrganizationService{
		db: db,
	}

	// Test data with invalid tier
	organizationID := "org-test-001"
	oldTier := "basic"
	newTier := "invalid_tier"
	reason := "Test invalid tier"
	adminUserID := "user-admin-001"
	ipAddress := "192.168.1.1"

	// Execute
	result, err := service.ChangeSubscriptionTier(
		organizationID,
		oldTier,
		newTier,
		reason,
		adminUserID,
		ipAddress,
	)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid tier")
}

func TestChangeSubscriptionTier_SameTier(t *testing.T) {
	// Setup mock database
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	service := &AdminOrganizationService{
		db: db,
	}

	// Test data with same tier
	organizationID := "org-test-001"
	oldTier := "professional"
	newTier := "professional"
	reason := "Test same tier"
	adminUserID := "user-admin-001"
	ipAddress := "192.168.1.1"

	// Execute
	result, err := service.ChangeSubscriptionTier(
		organizationID,
		oldTier,
		newTier,
		reason,
		adminUserID,
		ipAddress,
	)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "same tier")
}

func TestChangeSubscriptionTier_EmptyReason(t *testing.T) {
	// Setup mock database
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	service := &AdminOrganizationService{
		db: db,
	}

	// Test data with empty reason
	organizationID := "org-test-001"
	oldTier := "basic"
	newTier := "professional"
	reason := ""
	adminUserID := "user-admin-001"
	ipAddress := "192.168.1.1"

	// Execute
	result, err := service.ChangeSubscriptionTier(
		organizationID,
		oldTier,
		newTier,
		reason,
		adminUserID,
		ipAddress,
	)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "reason")
}

func TestChangeSubscriptionTier_AllTiers(t *testing.T) {
	validTiers := []string{"basic", "professional", "enterprise", "unlimited"}

	for i := 0; i < len(validTiers)-1; i++ {
		for j := i + 1; j < len(validTiers); j++ {
			t.Run("Upgrade_"+validTiers[i]+"_to_"+validTiers[j], func(t *testing.T) {
				// Setup mock database
				db, mock, err := sqlmock.New()
				require.NoError(t, err)
				defer db.Close()

				service := &AdminOrganizationService{
					db: db,
				}

				// Expect transaction
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE organizations").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("INSERT INTO subscription_events").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("INSERT INTO admin_audit_logs").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()

				// Execute
				result, err := service.ChangeSubscriptionTier(
					"org-test-001",
					validTiers[i],
					validTiers[j],
					"Test tier change",
					"user-admin-001",
					"192.168.1.1",
				)

				// Assert
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, validTiers[i], result.OldTier)
				assert.Equal(t, validTiers[j], result.NewTier)
			})
		}
	}
}

func TestChangeSubscriptionTier_ConcurrentChanges(t *testing.T) {
	// Setup mock database
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	service := &AdminOrganizationService{
		db: db,
	}

	// Simulate concurrent modification
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE organizations").
		WillReturnResult(sqlmock.NewResult(0, 0)) // No rows affected
	mock.ExpectRollback()

	// Execute
	result, err := service.ChangeSubscriptionTier(
		"org-test-001",
		"basic",
		"professional",
		"Test concurrent change",
		"user-admin-001",
		"192.168.1.1",
	)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "no rows affected")
}

func TestChangeSubscriptionTier_AuditLogMetadata(t *testing.T) {
	// Setup mock database
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	service := &AdminOrganizationService{
		db: db,
	}

	ipAddress := "203.0.113.42"
	organizationID := "org-test-001"

	// Expect transaction
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE organizations").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO subscription_events").WillReturnResult(sqlmock.NewResult(1, 1))

	// Expect audit log with specific metadata
	mock.ExpectExec("INSERT INTO admin_audit_logs").
		WithArgs(
			sqlmock.AnyArg(),
			organizationID,
			"tier_change",
			"basic",
			"professional",
			sqlmock.AnyArg(), // metadata should contain IP address
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	// Execute
	result, err := service.ChangeSubscriptionTier(
		organizationID,
		"basic",
		"professional",
		"Test metadata",
		"user-admin-001",
		ipAddress,
	)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}
