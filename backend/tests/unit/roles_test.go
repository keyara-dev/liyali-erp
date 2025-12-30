package unit

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/gofiber/fiber/v2"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/services"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestApp creates a test Fiber app with test database
func setupTestApp(t *testing.T) (*fiber.App, *gorm.DB) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Auto-migrate models
	err = db.AutoMigrate(
		&models.OrganizationRole{},
		&models.OrganizationPermission{},
		&models.PermissionAssignment{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	// Store DB in config for handlers
	config.DB = db

	app := fiber.New()

	// Add test routes for role handlers
	api := app.Group("/api/v1")
	roles := api.Group("/organization/roles")

	roles.Get("", GetOrganizationRoles)
	roles.Post("", CreateOrganizationRole)
	roles.Put("/:roleId", UpdateOrganizationRole)
	roles.Delete("/:roleId", DeleteOrganizationRole)
	roles.Get("/:roleId/permissions", GetRolePermissions)
	roles.Post("/:roleId/permissions/:permissionId", AssignPermissionToRole)
	roles.Delete("/:roleId/permissions/:permissionId", RemovePermissionFromRole)
	api.Get("/organization/permissions", GetOrganizationPermissions)

	return app, db
}

// TestGetOrganizationRoles tests retrieving roles
func TestGetOrganizationRoles(t *testing.T) {
	app, db := setupTestApp(t)

	orgID := uuid.New().String()

	// Create test roles
	svc := services.NewRoleManagementService(db)
	svc.CreateOrganizationRole(orgID, "Manager", "Manages team")
	svc.CreateOrganizationRole(orgID, "Coordinator", "Coordinates tasks")

	req := httptest.NewRequest("GET", "/api/v1/organization/roles", nil)
	req.Header.Set("X-Organization-ID", orgID)

	resp, err := app.Test(req)

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

// TestCreateOrganizationRole tests creating a new role
func TestCreateOrganizationRole(t *testing.T) {
	app, _ := setupTestApp(t)

	orgID := uuid.New().String()
	body := map[string]string{
		"name":        "Manager",
		"description": "Manages team operations",
	}

	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/v1/organization/roles", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Organization-ID", orgID)

	resp, err := app.Test(req)

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", resp.StatusCode)
	}
}

// TestCreateOrganizationRole_InvalidRequest tests creating role with invalid data
func TestCreateOrganizationRole_InvalidRequest(t *testing.T) {
	app, _ := setupTestApp(t)

	orgID := uuid.New().String()
	body := map[string]string{
		"name": "", // Missing name
	}

	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/v1/organization/roles", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Organization-ID", orgID)

	resp, err := app.Test(req)

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}
}

// TestUpdateOrganizationRole tests updating a role
func TestUpdateOrganizationRole(t *testing.T) {
	app, db := setupTestApp(t)

	orgID := uuid.New().String()
	svc := services.NewRoleManagementService(db)
	role, _ := svc.CreateOrganizationRole(orgID, "Manager", "Old description")

	body := map[string]string{
		"name":        "Senior Manager",
		"description": "New description",
	}

	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest("PUT", "/api/v1/organization/roles/"+role.ID, bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Organization-ID", orgID)

	resp, err := app.Test(req)

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

// TestDeleteOrganizationRole tests deleting a role
func TestDeleteOrganizationRole(t *testing.T) {
	app, db := setupTestApp(t)

	orgID := uuid.New().String()
	svc := services.NewRoleManagementService(db)
	role, _ := svc.CreateOrganizationRole(orgID, "Manager", "Description")

	req := httptest.NewRequest("DELETE", "/api/v1/organization/roles/"+role.ID, nil)
	req.Header.Set("X-Organization-ID", orgID)

	resp, err := app.Test(req)

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

// TestDeleteOrganizationRole_DefaultRoleProtection tests that default roles cannot be deleted
func TestDeleteOrganizationRole_DefaultRoleProtection(t *testing.T) {
	app, db := setupTestApp(t)

	orgID := uuid.New().String()

	// Create a default role directly in database
	defaultRole := models.OrganizationRole{
		ID:             uuid.New().String(),
		OrganizationID: orgID,
		Name:           "admin",
		Description:    "System admin role",
		IsDefault:      true,
		IsActive:       true,
	}
	db.Create(&defaultRole)

	req := httptest.NewRequest("DELETE", "/api/v1/organization/roles/"+defaultRole.ID, nil)
	req.Header.Set("X-Organization-ID", orgID)

	resp, err := app.Test(req)

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	// Should return forbidden or error status
	if resp.StatusCode == http.StatusOK {
		t.Errorf("Expected non-200 status for default role deletion, got %d", resp.StatusCode)
	}
}

// TestGetOrganizationPermissions tests retrieving available permissions
func TestGetOrganizationPermissions(t *testing.T) {
	app, db := setupTestApp(t)

	orgID := uuid.New().String()
	svc := services.NewRoleManagementService(db)

	// Create test permissions
	svc.CreateOrganizationPermission(orgID, "requisition", "approve", "Approve requisitions")
	svc.CreateOrganizationPermission(orgID, "requisition", "create", "Create requisitions")
	svc.CreateOrganizationPermission(orgID, "budget", "view", "View budgets")

	req := httptest.NewRequest("GET", "/api/v1/organization/permissions", nil)
	req.Header.Set("X-Organization-ID", orgID)

	resp, err := app.Test(req)

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

// TestGetRolePermissions tests retrieving permissions for a specific role
func TestGetRolePermissions(t *testing.T) {
	app, db := setupTestApp(t)

	orgID := uuid.New().String()
	svc := services.NewRoleManagementService(db)

	// Create role and permissions
	role, _ := svc.CreateOrganizationRole(orgID, "Manager", "Description")
	perm1, _ := svc.CreateOrganizationPermission(orgID, "requisition", "approve", "Approve")
	perm2, _ := svc.CreateOrganizationPermission(orgID, "budget", "view", "View budgets")

	// Assign permissions
	svc.AssignPermissionToRole(role.ID, perm1.ID)
	svc.AssignPermissionToRole(role.ID, perm2.ID)

	req := httptest.NewRequest("GET", "/api/v1/organization/roles/"+role.ID+"/permissions", nil)
	req.Header.Set("X-Organization-ID", orgID)

	resp, err := app.Test(req)

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

// TestAssignPermissionToRole tests assigning a permission to a role
func TestAssignPermissionToRole(t *testing.T) {
	app, db := setupTestApp(t)

	orgID := uuid.New().String()
	svc := services.NewRoleManagementService(db)

	role, _ := svc.CreateOrganizationRole(orgID, "Manager", "Description")
	perm, _ := svc.CreateOrganizationPermission(orgID, "requisition", "approve", "Approve")

	req := httptest.NewRequest("POST", "/api/v1/organization/roles/"+role.ID+"/permissions/"+perm.ID, nil)
	req.Header.Set("X-Organization-ID", orgID)

	resp, err := app.Test(req)

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", resp.StatusCode)
	}
}

// TestRemovePermissionFromRole tests removing a permission from a role
func TestRemovePermissionFromRole(t *testing.T) {
	app, db := setupTestApp(t)

	orgID := uuid.New().String()
	svc := services.NewRoleManagementService(db)

	role, _ := svc.CreateOrganizationRole(orgID, "Manager", "Description")
	perm, _ := svc.CreateOrganizationPermission(orgID, "requisition", "approve", "Approve")

	svc.AssignPermissionToRole(role.ID, perm.ID)

	req := httptest.NewRequest("DELETE", "/api/v1/organization/roles/"+role.ID+"/permissions/"+perm.ID, nil)
	req.Header.Set("X-Organization-ID", orgID)

	resp, err := app.Test(req)

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}
