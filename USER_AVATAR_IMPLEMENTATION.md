# User Avatar Upload Implementation

## Summary

Implemented user profile picture upload functionality using ImageKit CDN, similar to the organization logo feature.

## Components Created

### 1. UserAvatarUpload Component

**File**: `frontend/src/components/ui/user-avatar-upload.tsx`

Full-featured avatar upload component with:

- Drag-and-drop support
- Click to browse
- Real-time progress tracking
- File validation (type, size)
- Preview before save
- Remove avatar option
- Multiple size variants (sm, md, lg)
- Circular avatar display

### 2. UserAvatar Component

**File**: `frontend/src/components/ui/user-avatar.tsx`

Display component for consistent user avatar rendering:

- Automatic ImageKit optimization
- Multiple size variants (xs, sm, md, lg, xl)
- Fallback to initials
- Circular design
- Optimized image loading

## Integration Points

### Account Settings Page

**File**: `frontend/src/app/(private)/settings/_components/account-settings.tsx`

Updated to include:

- Avatar upload section at the top
- Form state management for avatar URL
- useEffect to sync with user data changes
- Avatar included in profile update

## Features

### Upload Functionality

- ✅ Drag-and-drop support
- ✅ Click to browse
- ✅ Real-time progress (0-100%)
- ✅ File validation (JPG, PNG, GIF, WebP)
- ✅ Size validation (max 10MB)
- ✅ Preview before save
- ✅ Remove avatar option

### Display Functionality

- ✅ Automatic image optimization via ImageKit
- ✅ CDN delivery
- ✅ WebP format when supported
- ✅ Fallback to user initials
- ✅ Consistent circular styling
- ✅ Multiple size variants

### Storage

- ✅ Images stored in ImageKit under `/avatars` folder
- ✅ URL saved to user profile
- ✅ Automatic optimization and transformation

## How It Works

### Upload Flow

1. User clicks "Upload Avatar" or drags image
2. Client validates file (type, size)
3. Client requests authentication from `/api/imagekit-auth`
4. Server generates secure token and signature
5. Client uploads to ImageKit `/avatars` folder
6. ImageKit returns image URL
7. URL stored in form state
8. User clicks "Save Changes"
9. Avatar URL saved to user profile

### Display Flow

1. Component receives user data with avatar URL
2. `UserAvatar` applies ImageKit transformations
3. Optimized image loaded from CDN
4. Falls back to initials if no avatar

## Usage Examples

### In Account Settings (Already Implemented)

```tsx
<UserAvatarUpload
  currentAvatarUrl={formData.avatar}
  userName={formData.name || "User"}
  onAvatarChange={handleAvatarChange}
  disabled={isLoading}
  size="lg"
/>
```

### Display User Avatar Anywhere

```tsx
import { UserAvatar } from "@/components/ui/user-avatar";

<UserAvatar name={user.name} avatarUrl={user.avatar} size="md" />;
```

### In User Menu/Dropdown

```tsx
<UserAvatar name={user.name} avatarUrl={user.avatar} size="sm" />
```

## Where to Use UserAvatar Component

Replace existing avatar implementations in:

1. **Sidebar User Menu** (`frontend/src/components/layout/sidebar/nav-user.tsx`)
   - Currently uses hardcoded fallback URL
   - Replace with `UserAvatar` component

2. **Header User Menu** (`frontend/src/components/layout/header/user-menu.tsx`)
   - Currently uses hardcoded fallback URL
   - Replace with `UserAvatar` component

3. **Workflow Reassignment Modal** (`frontend/src/components/workflows/reassignment-modal.tsx`)
   - Currently uses basic Avatar
   - Replace with `UserAvatar` for consistency

## Backend Support

### Current Status

The `updateUserProfile` function in `frontend/src/app/_actions/settings.ts` already accepts an `avatar` parameter.

**Note**: Currently a mock implementation. For production:

1. Backend User model needs `avatar` field (or `avatar_url`)
2. Backend handler needs to accept and save avatar URL
3. Database migration to add avatar column

### Backend Changes Needed

#### 1. Update User Model

```go
// backend/models/models.go
type User struct {
    // ... existing fields
    Avatar    string `json:"avatar,omitempty"`
    // ... rest of fields
}
```

#### 2. Create Update Profile Handler

```go
// backend/handlers/user.go
func UpdateUserProfile(c *fiber.Ctx) error {
    var req struct {
        Name       string  `json:"name"`
        Email      string  `json:"email"`
        Department string  `json:"department"`
        Avatar     *string `json:"avatar"`
    }

    // Parse, validate, and update user
    // Return updated user data
}
```

#### 3. Database Migration

```sql
ALTER TABLE users ADD COLUMN avatar VARCHAR(500);
```

## Testing Checklist

### Upload Functionality

- [ ] Drag-and-drop works
- [ ] Click to browse works
- [ ] Progress shows 0-100%
- [ ] Preview updates after upload
- [ ] Success toast appears
- [ ] File validation works (reject .txt)
- [ ] Size validation works (reject >10MB)

### Display Functionality

- [ ] Avatar appears in account settings
- [ ] Avatar appears in sidebar
- [ ] Avatar appears in header menu
- [ ] Fallback to initials works
- [ ] Images are optimized (check Network tab)
- [ ] Multiple sizes work correctly

### Update Functionality

- [ ] Can change avatar
- [ ] Can remove avatar
- [ ] Save button enables on change
- [ ] Changes persist after save
- [ ] Avatar updates everywhere

## Image Specifications

### Upload Constraints

- **Formats**: JPG, PNG, GIF, WebP
- **Max Size**: 10MB
- **Folder**: `/avatars` in ImageKit

### Display Sizes

- **xs**: 24px (6x6)
- **sm**: 32px (8x8)
- **md**: 40px (10x10) - Default
- **lg**: 48px (12x12)
- **xl**: 64px (16x16)

### Optimization

- Automatic resizing based on display size
- Quality: 80%
- Format: auto (WebP when supported)
- Crop: maintain aspect ratio
- Circular crop for avatars

## Security

- ✅ Server-side authentication
- ✅ Private key never exposed
- ✅ Token-based uploads
- ✅ Signature verification
- ✅ 1-hour token expiration
- ✅ File type validation
- ✅ File size validation

## Next Steps

### Immediate

1. Test avatar upload in account settings
2. Verify ImageKit credentials are set
3. Test with different image formats

### Short Term

1. Replace hardcoded avatars in sidebar
2. Replace hardcoded avatars in header
3. Update workflow components to use UserAvatar

### Backend Integration

1. Add avatar field to User model
2. Create/update user profile endpoint
3. Add database migration
4. Test end-to-end flow

## Files Created

1. `frontend/src/components/ui/user-avatar-upload.tsx` - Upload component
2. `frontend/src/components/ui/user-avatar.tsx` - Display component
3. `USER_AVATAR_IMPLEMENTATION.md` - This documentation

## Files Modified

1. `frontend/src/app/(private)/settings/_components/account-settings.tsx`
   - Added avatar upload section
   - Added avatar state management
   - Added useEffect for data sync
   - Updated submit handler

## Related Documentation

- Organization Logo: `ORGANIZATION_LOGO_IMPLEMENTATION.md`
- ImageKit Setup: `frontend/docs/IMAGEKIT_SETUP.md`
- ImageKit Testing: `frontend/docs/IMAGEKIT_TESTING_GUIDE.md`

## Status

✅ **Frontend Complete**

- Upload component created
- Display component created
- Account settings integrated
- Form state management added

⏳ **Backend Pending**

- User model needs avatar field
- Update profile endpoint needed
- Database migration needed

🔄 **Integration Pending**

- Replace hardcoded avatars in sidebar
- Replace hardcoded avatars in header
- Update workflow components

---

**Implementation Date**: 2024
**Status**: Frontend Ready, Backend Integration Needed
**Breaking Changes**: None
