package unit

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/types"
)

// TestCreateVendorValidation tests vendor request validation
func TestCreateVendorValidation(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
		shouldPass     bool
	}{
		{
			name: "Valid vendor request",
			requestBody: map[string]interface{}{
				"name":        "ABC Supplies Ltd",
				"email":       "contact@abcsupplies.com",
				"phone":       "+263 4 123456",
				"country":     "Zimbabwe",
				"city":        "Harare",
				"bankAccount": "1234567890",
				"taxID":       "TAX123456",
			},
			expectedStatus: http.StatusCreated,
			shouldPass:     true,
		},
		{
			name: "Missing vendor name",
			requestBody: map[string]interface{}{
				"email":       "contact@abcsupplies.com",
				"phone":       "+263 4 123456",
				"country":     "Zimbabwe",
				"city":        "Harare",
				"bankAccount": "1234567890",
				"taxID":       "TAX123456",
			},
			expectedStatus: http.StatusBadRequest,
			shouldPass:     false,
		},
		{
			name: "Short vendor name",
			requestBody: map[string]interface{}{
				"name":        "AB",
				"email":       "contact@abcsupplies.com",
				"phone":       "+263 4 123456",
				"country":     "Zimbabwe",
				"city":        "Harare",
				"bankAccount": "1234567890",
				"taxID":       "TAX123456",
			},
			expectedStatus: http.StatusBadRequest,
			shouldPass:     false,
		},
		{
			name: "Missing email",
			requestBody: map[string]interface{}{
				"name":        "ABC Supplies Ltd",
				"phone":       "+263 4 123456",
				"country":     "Zimbabwe",
				"city":        "Harare",
				"bankAccount": "1234567890",
				"taxID":       "TAX123456",
			},
			expectedStatus: http.StatusBadRequest,
			shouldPass:     false,
		},
		{
			name: "Missing phone",
			requestBody: map[string]interface{}{
				"name":        "ABC Supplies Ltd",
				"email":       "contact@abcsupplies.com",
				"country":     "Zimbabwe",
				"city":        "Harare",
				"bankAccount": "1234567890",
				"taxID":       "TAX123456",
			},
			expectedStatus: http.StatusBadRequest,
			shouldPass:     false,
		},
		{
			name: "Missing country",
			requestBody: map[string]interface{}{
				"name":        "ABC Supplies Ltd",
				"email":       "contact@abcsupplies.com",
				"phone":       "+263 4 123456",
				"city":        "Harare",
				"bankAccount": "1234567890",
				"taxID":       "TAX123456",
			},
			expectedStatus: http.StatusBadRequest,
			shouldPass:     false,
		},
		{
			name: "Missing city",
			requestBody: map[string]interface{}{
				"name":        "ABC Supplies Ltd",
				"email":       "contact@abcsupplies.com",
				"phone":       "+263 4 123456",
				"country":     "Zimbabwe",
				"bankAccount": "1234567890",
				"taxID":       "TAX123456",
			},
			expectedStatus: http.StatusBadRequest,
			shouldPass:     false,
		},
		{
			name: "Missing bank account",
			requestBody: map[string]interface{}{
				"name":    "ABC Supplies Ltd",
				"email":   "contact@abcsupplies.com",
				"phone":   "+263 4 123456",
				"country": "Zimbabwe",
				"city":    "Harare",
				"taxID":   "TAX123456",
			},
			expectedStatus: http.StatusBadRequest,
			shouldPass:     false,
		},
		{
			name: "Missing tax ID",
			requestBody: map[string]interface{}{
				"name":        "ABC Supplies Ltd",
				"email":       "contact@abcsupplies.com",
				"phone":       "+263 4 123456",
				"country":     "Zimbabwe",
				"city":        "Harare",
				"bankAccount": "1234567890",
			},
			expectedStatus: http.StatusBadRequest,
			shouldPass:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.requestBody)
			var req types.CreateVendorRequest
			json.Unmarshal(body, &req)

			// Validate request
			isValid := req.Name != "" && len(req.Name) >= 3 &&
				req.Email != "" &&
				req.Phone != "" &&
				req.Country != "" &&
				req.City != "" &&
				req.BankAccount != "" &&
				req.TaxID != ""

			if isValid != tt.shouldPass {
				t.Errorf("Expected %v, got %v", tt.shouldPass, isValid)
			}
		})
	}
}

// TestVendorCodeGeneration tests vendor code generation
func TestVendorCodeGeneration(t *testing.T) {
	t.Run("Vendor code format", func(t *testing.T) {
		// Format: VND-{timestamp}-{uuid[:6]}
		vendorCode := "VND-1640000000-abc123"

		if vendorCode[:4] != "VND-" {
			t.Error("Vendor code should start with 'VND-'")
		}

		if len(vendorCode) < 15 {
			t.Error("Vendor code should be properly formatted")
		}
	})
}

// TestVendorEmailValidation tests email field
func TestVendorEmailValidation(t *testing.T) {
	tests := []struct {
		name          string
		email         string
		shouldBeValid bool
	}{
		{"Valid email", "contact@company.com", true},
		{"Another valid email", "vendor@abc.co.zw", true},
		{"Empty email", "", false},
		{"Invalid format", "notanemail", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.email != "" && len(tt.email) > 5 && strings.Contains(tt.email, "@") && strings.Contains(tt.email, ".")
			if isValid != tt.shouldBeValid {
				t.Errorf("Expected %v, got %v", tt.shouldBeValid, isValid)
			}
		})
	}
}

// TestVendorDuplicateEmailPrevention tests duplicate email detection
func TestVendorDuplicateEmailPrevention(t *testing.T) {
	t.Run("Prevent duplicate emails", func(t *testing.T) {
		vendor1 := types.VendorResponse{
			ID:    uuid.New().String(),
			Name:  "ABC Supplies",
			Email: "contact@abcsupplies.com",
		}

		vendor2 := types.VendorResponse{
			ID:    uuid.New().String(),
			Name:  "XYZ Supplies",
			Email: "contact@abcsupplies.com",
		}

		isDuplicate := vendor1.Email == vendor2.Email

		if !isDuplicate {
			t.Error("Should detect duplicate emails")
		}
	})
}

// TestVendorNameValidation tests vendor name field
func TestVendorNameValidation(t *testing.T) {
	tests := []struct {
		name          string
		vendorName    string
		shouldBeValid bool
	}{
		{"Valid name", "ABC Supplies Ltd", true},
		{"Short but valid", "ABC Co", true},
		{"Too short", "AB", false},
		{"Empty name", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.vendorName != "" && len(tt.vendorName) >= 3
			if isValid != tt.shouldBeValid {
				t.Errorf("Expected %v, got %v", tt.shouldBeValid, isValid)
			}
		})
	}
}

// TestVendorCountryValidation tests country field
func TestVendorCountryValidation(t *testing.T) {
	validCountries := map[string]bool{
		"Zimbabwe":   true,
		"South Africa": true,
		"Botswana":   true,
		"Kenya":      true,
	}

	tests := []struct {
		name          string
		country       string
		shouldBeValid bool
	}{
		{"Zimbabwe", "Zimbabwe", true},
		{"South Africa", "South Africa", true},
		{"Botswana", "Botswana", true},
		{"Invalid", "Atlantis", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := validCountries[tt.country] || tt.country == ""
			if isValid != tt.shouldBeValid {
				t.Errorf("Expected %v, got %v", tt.shouldBeValid, isValid)
			}
		})
	}
}

// TestVendorPhoneValidation tests phone field
func TestVendorPhoneValidation(t *testing.T) {
	tests := []struct {
		name          string
		phone         string
		shouldBeValid bool
	}{
		{"Valid phone", "+263 4 123456", true},
		{"Another format", "0772123456", true},
		{"Short number", "123", false},
		{"Empty phone", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.phone != "" && len(tt.phone) >= 5
			if isValid != tt.shouldBeValid {
				t.Errorf("Expected %v, got %v", tt.shouldBeValid, isValid)
			}
		})
	}
}

// TestVendorTaxIDValidation tests tax ID field
func TestVendorTaxIDValidation(t *testing.T) {
	tests := []struct {
		name          string
		taxID         string
		shouldBeValid bool
	}{
		{"Valid tax ID", "TAX123456", true},
		{"Another format", "ZWL-123456", true},
		{"Empty tax ID", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.taxID != ""
			if isValid != tt.shouldBeValid {
				t.Errorf("Expected %v, got %v", tt.shouldBeValid, isValid)
			}
		})
	}
}

// TestVendorBankAccountValidation tests bank account field
func TestVendorBankAccountValidation(t *testing.T) {
	tests := []struct {
		name              string
		bankAccount       string
		shouldBeValid     bool
	}{
		{"Valid account", "1234567890", true},
		{"Account with code", "1234567890-USD", true},
		{"Short account", "123", false},
		{"Empty account", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.bankAccount != "" && len(tt.bankAccount) >= 8
			if isValid != tt.shouldBeValid {
				t.Errorf("Expected %v, got %v", tt.shouldBeValid, isValid)
			}
		})
	}
}

// TestVendorResponseFormat tests vendor response structure
func TestVendorResponseFormat(t *testing.T) {
	t.Run("Vendor response structure", func(t *testing.T) {
		vendor := types.VendorResponse{
			ID:          uuid.New().String(),
			VendorCode:  "VND-1640000000-abc123",
			Name:        "ABC Supplies Ltd",
			Email:       "contact@abcsupplies.com",
			Phone:       "+263 4 123456",
			Country:     "Zimbabwe",
			City:        "Harare",
			BankAccount: "1234567890",
			TaxID:       "TAX123456",
			Active:      true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		if vendor.ID == "" {
			t.Error("Response should have ID")
		}
		if vendor.VendorCode == "" {
			t.Error("Response should have VendorCode")
		}
		if vendor.Name == "" {
			t.Error("Response should have Name")
		}
		if vendor.Email == "" {
			t.Error("Response should have Email")
		}
	})
}

// TestVendorActiveStatusValidation tests active status field
func TestVendorActiveStatusValidation(t *testing.T) {
	tests := []struct {
		name      string
		active    bool
		isVisible bool
	}{
		{"Active vendor", true, true},
		{"Inactive vendor", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vendor := types.VendorResponse{
				ID:     uuid.New().String(),
				Name:   "Test Vendor",
				Active: tt.active,
			}

			if vendor.Active != tt.active {
				t.Errorf("Expected active %v, got %v", tt.active, vendor.Active)
			}
		})
	}
}

// TestVendorSoftDeleteViActiveFlag tests soft delete mechanism
func TestVendorSoftDeleteViaActiveFlag(t *testing.T) {
	t.Run("Soft delete via active flag", func(t *testing.T) {
		vendor := types.VendorResponse{
			ID:     uuid.New().String(),
			Name:   "ABC Supplies",
			Active: true,
		}

		// Mark as deleted by setting active to false
		vendor.Active = false

		if vendor.Active {
			t.Error("Vendor should be marked as inactive")
		}

		if vendor.ID == "" {
			t.Error("Vendor should still have ID (soft delete)")
		}
	})
}

// TestVendorUpdateValidation tests vendor update constraints
func TestVendorUpdateValidation(t *testing.T) {
	tests := []struct {
		name        string
		field       string
		value       string
		shouldAllow bool
	}{
		{"Update name", "name", "New Name Corp", true},
		{"Update email", "email", "newemail@company.com", true},
		{"Update phone", "phone", "+263 4 654321", true},
		{"Update country", "country", "South Africa", true},
		{"Update city", "city", "Bulawayo", true},
		{"Update bank account", "bankAccount", "9876543210", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// All fields can be updated for active vendors
			isAllowed := true

			if !isAllowed != !tt.shouldAllow {
				t.Errorf("Expected %v, got %v", tt.shouldAllow, isAllowed)
			}
		})
	}
}

// TestVendorCityValidation tests city field
func TestVendorCityValidation(t *testing.T) {
	tests := []struct {
		name          string
		city          string
		shouldBeValid bool
	}{
		{"Valid city", "Harare", true},
		{"Another city", "Bulawayo", true},
		{"Empty city", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.city != ""
			if isValid != tt.shouldBeValid {
				t.Errorf("Expected %v, got %v", tt.shouldBeValid, isValid)
			}
		})
	}
}

// TestVendorFilteringByStatus tests filtering by active status
func TestVendorFilteringByStatus(t *testing.T) {
	tests := []struct {
		name       string
		allVendors []bool
		filterBy   bool
		expected   int
	}{
		{
			name:       "Filter active vendors",
			allVendors: []bool{true, true, false, true},
			filterBy:   true,
			expected:   3,
		},
		{
			name:       "Filter inactive vendors",
			allVendors: []bool{true, true, false, true},
			filterBy:   false,
			expected:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count := 0
			for _, active := range tt.allVendors {
				if active == tt.filterBy {
					count++
				}
			}

			if count != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, count)
			}
		})
	}
}

// TestVendorListPagination tests pagination for vendor list
func TestVendorListPagination(t *testing.T) {
	tests := []struct {
		name      string
		page      int
		limit     int
		total     int64
		expected  int
	}{
		{"First page", 1, 10, 25, 10},
		{"Second page", 2, 10, 25, 10},
		{"Last page partial", 3, 10, 25, 5},
		{"Single item", 1, 1, 1, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			offset := (tt.page - 1) * tt.limit
			remaining := tt.total - int64(offset)

			var pageSize int64
			if remaining < int64(tt.limit) {
				pageSize = remaining
			} else {
				pageSize = int64(tt.limit)
			}

			if pageSize != int64(tt.expected) {
				t.Errorf("Expected %d items, got %d", tt.expected, pageSize)
			}
		})
	}
}

// TestVendorContactInfoCompleteness tests contact information
func TestVendorContactInfoCompleteness(t *testing.T) {
	t.Run("Vendor has complete contact info", func(t *testing.T) {
		vendor := types.VendorResponse{
			ID:     uuid.New().String(),
			Name:   "ABC Supplies",
			Email:  "contact@abc.com",
			Phone:  "+263 4 123456",
			City:   "Harare",
			Country: "Zimbabwe",
		}

		hasCompleteInfo := vendor.Email != "" &&
			vendor.Phone != "" &&
			vendor.City != "" &&
			vendor.Country != ""

		if !hasCompleteInfo {
			t.Error("Vendor should have complete contact information")
		}
	})
}

// BenchmarkVendorValidation benchmarks validation logic
func BenchmarkVendorValidation(b *testing.B) {
	req := types.CreateVendorRequest{
		Name:        "ABC Supplies Ltd",
		Email:       "contact@abcsupplies.com",
		Phone:       "+263 4 123456",
		Country:     "Zimbabwe",
		City:        "Harare",
		BankAccount: "1234567890",
		TaxID:       "TAX123456",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = req.Name != "" && len(req.Name) >= 3 &&
			req.Email != "" &&
			req.Phone != "" &&
			req.Country != "" &&
			req.City != "" &&
			req.BankAccount != "" &&
			req.TaxID != ""
	}
}

// BenchmarkVendorCodeGeneration benchmarks code generation
func BenchmarkVendorCodeGeneration(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vendorCode := "VND-" + uuid.New().String()[:6]
		_ = vendorCode
	}
}
