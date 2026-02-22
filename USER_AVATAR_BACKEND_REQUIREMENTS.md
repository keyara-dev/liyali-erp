# User Avatar Backend Implementation Requirements

## Current Issue

User profile pictures (avatars) are not persisting after page reload because the backend does not support storing avatar URLs.

## Root Cause

1. The `User` model in `backend/models/models.go` does not have an `avatar` field
2. There is no backend API endpoint to update user profile information
3. The frontend `updateUserProfile` action is currently a mock implementation

## Required Backend Changes

### 1. Update User Model

**File:** `backend/models/models.go`

Add the `Avatar` field to the User struct:

```go
type User struct {
	ID        string     `gorm:"primaryKey" json:"id"`
	Email     string     `gorm:"uniqueIndex" json:"email"`
	Name      string     `json:"name"`
	Password  string     `json:"-"` // Hidden from JSON responses
	Role      string     `json:"role"` // admin, approver, requester, finance, viewer
	Active    bool       `json:"active"`
	LastLogin *time.Time `json:"lastLogin,omitempty"`
	Avatar    string     `json:"avatar,omitempty"` // NEW: User profile picture URL
	Department string    `json:"department,omitempty"` // Consider adding if not exists

	// Multi-tenancy fields
	CurrentOrganizationID *string        `json:"currentOrganizationId,omitempty"`
	CurrentOrganization   *Organization `gorm:"foreignKey:CurrentOrganizationID" json:"currentOrganization,omitempty"`
	IsSuperAdmin          bool           `gorm:"default:false" json:"isSuperAdmin"`
	Preferences           datatypes.JSON `gorm:"type:jsonb" json:"preferences,omitempty"`

	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `json:"deletedAt" gorm:"index"` // Soft delete
}
```

### 2. Create Database Migration

**File:** `backend/database/migrations/XXX_add_user_avatar.up.sql`

```sql
-- Add avatar column to users table
ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar VARCHAR(500);

-- Add department column if it doesn't exist
ALTER TABLE users ADD COLUMN IF NOT EXISTS department VARCHAR(100);

-- Add index for faster lookups
CREATE INDEX IF NOT EXISTS idx_users_avatar ON users(avatar) WHERE avatar IS NOT NULL;
```

**File:** `backend/database/migrations/XXX_add_user_avatar.down.sql`

```sql
-- Remove avatar column
ALTER TABLE users DROP COLUMN IF EXISTS avatar;

-- Remove department column (only if added in this migration)
-- ALTER TABLE users DROP COLUMN IF EXISTS department;

-- Remove index
DROP INDEX IF EXISTS idx_users_avatar;
```

### 3. Create User Update Handler

**File:** `backend/handlers/user_handler.go` (new file)

```go
package handlers

import (
	"github.com/gofiber/fiber/v2"
	"liyali-gateway/backend/config"
	"liyali-gateway/backend/logging"
	"liyali-gateway/backend/middleware"
	"liyali-gateway/backend/models"
	"liyali-gateway/backend/utils"
)

// UpdateUserProfile updates the current user's profile
// PUT /api/v1/users/profile
func UpdateUserProfile(c *fiber.Ctx) error {
	logger := logging.FromContext(c)
	logger.Info("update_user_profile_attempt")

	// Get authenticated user ID
	userID, ok := c.Locals("userID").(string)
	if !ok {
		return utils.SendError(c, fiber.StatusUnauthorized, "User not authenticated")
	}

	// Parse request body
	type UpdateProfileRequest struct {
		Name       *string `json:"name"`
		Email      *string `json:"email"`
		Department *string `json:"department"`
		Avatar     *string `json:"avatar"`
	}

	var req UpdateProfileRequest
	if err := c.BodyParser(&req); err != nil {
		logger.Error("invalid_request_body", map[string]interface{}{
			"error": err.Error(),
		})
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	// Get tenant context
	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, err.Error())
	}

	db := config.DB.Scoped().Where("tenant_id = ?", tenant.ID)

	// Find user
	var user models.User
	if err := db.First(&user, "id = ?", userID).Error; err != nil {
		logger.Error("user_not_found", map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		})
		return utils.SendError(c, fiber.StatusNotFound, "User not found")
	}

	// Update fields if provided
	updates := make(map[string]interface{})

	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Email != nil {
		// Validate email uniqueness
		var existingUser models.User
		if err := db.Where("email = ? AND id != ?", *req.Email, userID).First(&existingUser).Error; err == nil {
			return utils.SendError(c, fiber.StatusConflict, "Email already in use")
		}
		updates["email"] = *req.Email
	}
	if req.Department != nil {
		updates["department"] = *req.Department
	}
	if req.Avatar != nil {
		updates["avatar"] = *req.Avatar
	}

	// Perform update
	if err := db.Model(&user).Updates(updates).Error; err != nil {
		logger.Error("profile_update_failed", map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		})
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to update profile")
	}

	// Reload user to get updated data
	if err := db.First(&user, "id = ?", userID).Error; err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to retrieve updated profile")
	}

	logger.Info("profile_updated_successfully", map[string]interface{}{
		"user_id": userID,
	})

	return utils.SendSimpleSuccess(c, user, "Profile updated successfully")
}

// GetUserProfile returns the current user's profile
// GET /api/v1/users/profile
func GetUserProfile(c *fiber.Ctx) error {
	logger := logging.FromContext(c)
	logger.Info("get_user_profile_attempt")

	// Get authenticated user ID
	userID, ok := c.Locals("userID").(string)
	if !ok {
		return utils.SendError(c, fiber.StatusUnauthorized, "User not authenticated")
	}

	// Get tenant context
	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, err.Error())
	}

	db := config.DB.Scoped().Where("tenant_id = ?", tenant.ID)

	// Find user with current organization
	var user models.User
	if err := db.Preload("CurrentOrganization").First(&user, "id = ?", userID).Error; err != nil {
		logger.Error("user_not_found", map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		})
		return utils.SendError(c, fiber.StatusNotFound, "User not found")
	}

	logger.Info("profile_retrieved_successfully", map[string]interface{}{
		"user_id": userID,
	})

	return utils.SendSimpleSuccess(c, user, "Profile retrieved successfully")
}
```

### 4. Register Routes

**File:** `backend/main.go` or route registration file

Add these routes to the API:

```go
// User profile routes (authenticated)
api.Get("/users/profile", middleware.AuthRequired(), handlers.GetUserProfile)
api.Put("/users/profile", middleware.AuthRequired(), handlers.UpdateUserProfile)
```

### 5. Update Frontend Action

**File:** `frontend/src/app/_actions/settings.ts`

Replace the mock implementation with actual API call:

```typescript
export async function updateUserProfile(profileData: {
  name?: string;
  email?: string;
  department?: string;
  avatar?: string;
}): Promise<APIResponse> {
  try {
    const response = await fetch(
      `${process.env.NEXT_PUBLIC_API_URL}/api/v1/users/profile`,
      {
        method: "PUT",
        headers: {
          "Content-Type": "application/json",
        },
        credentials: "include",
        body: JSON.stringify(profileData),
      },
    );

    const data = await response.json();

    if (!response.ok) {
      return {
        success: false,
        message: data.message || "Failed to update profile",
        data: null,
        status: response.status,
        statusText: response.statusText,
      };
    }

    return {
      success: true,
      message: data.message || "Profile updated successfully",
      data: data.data,
      status: response.status,
      statusText: "OK",
    };
  } catch (error: any) {
    return {
      success: false,
      message: error.message || "Failed to update profile",
      data: null,
      status: 500,
      statusText: "ERROR",
    };
  }
}
```

## Testing Checklist

After implementing the backend changes:

1. ✅ Run database migration to add avatar column
2. ✅ Test GET /api/v1/users/profile endpoint
3. ✅ Test PUT /api/v1/users/profile endpoint with avatar URL
4. ✅ Verify avatar persists after page reload
5. ✅ Test avatar display in:
   - Account settings page
   - User menu dropdown
   - Sidebar user section
   - Any other places showing user avatar
6. ✅ Test avatar removal (setting to empty string)
7. ✅ Test with different image formats (JPG, PNG, WebP)
8. ✅ Verify ImageKit URLs are stored correctly

## Priority

**HIGH** - This is blocking the user avatar feature from working properly.

## Estimated Effort

- Backend changes: 2-3 hours
- Testing: 1 hour
- Total: 3-4 hours

## Related Files

- Backend: `backend/models/models.go`, `backend/handlers/user_handler.go`
- Frontend: `frontend/src/app/_actions/settings.ts`
- Components: `frontend/src/components/ui/user-avatar-upload.tsx`
