package unit

import (
	"testing"
	"time"

	"github.com/liyali/liyali-gateway/utils"
)

// TestSimpleUnit is a basic test to verify unit test setup
func TestSimpleUnit(t *testing.T) {
	t.Log("Testing unit test setup")
	
	// Test a simple utility function
	password := "testpassword123"
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}
	
	if hashedPassword == "" {
		t.Error("Hashed password should not be empty")
	}
	
	// Test password verification
	if !utils.VerifyPassword(hashedPassword, password) {
		t.Error("Password verification should succeed")
	}
	
	if utils.VerifyPassword(hashedPassword, "wrongpassword") {
		t.Error("Password verification should fail for wrong password")
	}
	
	t.Log("Unit test setup working correctly")
}

// TestTimeOperations tests basic time operations
func TestTimeOperations(t *testing.T) {
	now := time.Now()
	future := now.Add(1 * time.Hour)
	
	if !future.After(now) {
		t.Error("Future time should be after current time")
	}
	
	duration := future.Sub(now)
	if duration != 1*time.Hour {
		t.Errorf("Expected duration of 1 hour, got %v", duration)
	}
}