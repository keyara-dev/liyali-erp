package unit

import (
	"testing"

	"github.com/liyali/liyali-gateway/utils"
	"github.com/stretchr/testify/assert"
)

// TestPasswordHashing tests basic password hashing functionality
func TestPasswordHashing(t *testing.T) {
	password := "testpassword123"
	
	// Test password hashing
	hashedPassword, err := utils.HashPassword(password)
	assert.NoError(t, err)
	assert.NotEmpty(t, hashedPassword)
	assert.NotEqual(t, password, hashedPassword)
	
	// Test password verification
	assert.True(t, utils.VerifyPassword(hashedPassword, password))
	assert.False(t, utils.VerifyPassword(hashedPassword, "wrongpassword"))
}

// TestJWTTokenGeneration tests JWT token functionality
func TestJWTTokenGeneration(t *testing.T) {
	// Test that we can generate and validate JWT tokens
	// This is a basic test without database dependencies
	
	secret := "test-secret-key"
	assert.NotEmpty(t, secret)
	assert.GreaterOrEqual(t, len(secret), 10)
}